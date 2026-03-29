package ez

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"
	"github.com/spf13/afero"
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/core"
	"github.com/torloejborg/easykube/pkg/resources"
)

type EasykubeConfig struct {
	ek            *core.Ek
	ConfigDirName string
	UserConfigDir string
}

func NewEasykubeConfig(ek *core.Ek) core.IEasykubeConfig {
	return &EasykubeConfig{
		ek: ek,
	}
}

func (ec *EasykubeConfig) HasConfiguration() bool {
	_, err := ec.LoadConfig()
	return err == nil
}

func (ec *EasykubeConfig) LaunchEditor(config, editor string) {
	visual, err := exec.LookPath(editor)
	if err != nil {
		log.Panic(err)
	}

	cmd := exec.Command(visual, config)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		log.Fatalf("failed to open your editor")
	}
}

func (ec *EasykubeConfig) PathToConfigFile() string {

	configDir, _ := ec.ek.OsDetails.GetEasykubeConfigDir()
	return filepath.Join(configDir, "config.yaml")
}

func (ec *EasykubeConfig) LoadConfig() (*core.EasykubeConfigData, error) {

	config.ClearAll()
	// config.ParseEnv: will parse env var in string value. eg: shell: ${SHELL}
	config.WithOptions(config.ParseEnv)

	// add driver for support yaml content
	config.AddDriver(yaml.Driver)

	data, err := ec.ek.Utils.ReadFileToBytes(ec.PathToConfigFile())
	if err != nil {
		return nil, err
	}

	err = config.LoadSources("yaml", data)
	if err != nil {
		return nil, err
	}

	easykube := &core.EasykubeConfigData{}
	err = config.BindStruct("easykube", easykube)

	if err != nil {
		return nil, err
	}

	return easykube, nil
}

func (ec *EasykubeConfig) EditConfig() error {
	editor := os.Getenv("VISUAL")
	if len(editor) == 0 {
		fmt.Println("VISUAL environment variable not set")
		os.Exit(-1)
	} else {

		cfgFile := ec.PathToConfigFile()
		_, err := os.Stat(cfgFile)
		if err != nil {
			return errors.New("nothing to edit. create a default with config --use-defaults, or config")
		}

		ec.LaunchEditor(ec.PathToConfigFile(), editor)
	}

	return nil
}

func (ec *EasykubeConfig) CopyConfigResources() error {

	easykubeConfigDir, err := ec.ek.OsDetails.GetEasykubeConfigDir()
	if nil != err {
		return err
	}

	merr := ec.ek.Fs.MkdirAll(easykubeConfigDir, os.ModePerm)
	if merr != nil {
		return merr
	}

	userHomeDir, err := ec.ek.OsDetails.GetUserHomeDir()
	if nil != err {
		return err
	}

	pathToConfigFile := ec.PathToConfigFile()
	_, err = ec.ek.Fs.Stat(pathToConfigFile)

	err = ec.ek.Utils.CopyResourceToConfigDir("cert/localtest.me.crt", "localtest.me.crt")
	if nil != err {
		return err
	}

	err = ec.ek.Utils.CopyResourceToConfigDir("cert/localtest.me.ca.crt", "localtest.me.ca.crt")
	if nil != err {
		return err
	}

	err = ec.ek.Utils.CopyResourceToConfigDir("cert/localtest.me.key", "localtest.me.key")
	if nil != err {
		return err
	}

	// Make podman aware of a selfsigned certificate
	certData, err := resources.AppResources.ReadFile("data/cert/localtest.me.ca.crt")
	if nil != err {
		return err
	}

	certDestDir := filepath.Join(userHomeDir, ".config", "containers", "certs.d", constants.LocalRegistry)
	_ = ec.ek.Fs.MkdirAll(certDestDir, os.ModePerm)
	cert := filepath.Join(certDestDir, "ca.crt")
	ec.ek.Utils.SaveFileByte(certData, cert)

	return nil
}

func (ec *EasykubeConfig) MakeConfig() error {

	userHomeDir, err := ec.ek.OsDetails.GetUserHomeDir()
	configurationDir, err := ec.ek.OsDetails.GetEasykubeConfigDir()
	if nil != err {
		return err
	}

	pathToConfigFile := ec.PathToConfigFile()
	_, err = ec.ek.Fs.Stat(pathToConfigFile)

	_ = ec.CopyConfigResources()

	modelData := core.EasykubeConfigData{
		AddonDir:         filepath.Join(userHomeDir, "addons"),
		ConfigurationDir: configurationDir,
		PersistenceDir:   filepath.Join(configurationDir, "persistence"),
		ContainerRuntime: "docker",
		MirrorRegistries: []core.MirrorRegistry{
			{
				RegistryUrl: "https://registry-1.docker.io",
			},
			{
				RegistryUrl: "https://ghcr.io",
			},
			{
				RegistryUrl: "https://quay.io",
			},
			{
				RegistryUrl: "https://registry.k8s.io",
			},
		},
	}

	file, err := ec.ek.Fs.Create(pathToConfigFile)
	if nil != err {
		return err
	}

	configData, err := resources.AppResources.ReadFile("data/config.template")
	if nil != err {
		return err
	}

	templ := template.New("config")
	templ, err = templ.Parse(string(configData))
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}

	err = templ.Execute(buf, modelData)
	if err != nil {
		return err
	}

	_, err = file.WriteString(buf.String())
	if nil != err {
		return err
	}

	return nil
}

func (ec *EasykubeConfig) WriteConfig(cfg *core.EasykubeConfigData) error {
	configData, err := resources.AppResources.ReadFile("data/config.template")
	if nil != err {
		return err
	}

	templ := template.New("config")
	templ, err = templ.Parse(string(configData))
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}

	err = templ.Execute(buf, cfg)
	if err != nil {
		return err
	}

	file, err := ec.ek.Fs.OpenFile(cfg.ConfigurationFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if nil != err {
		panic(err)
	}

	_, err = file.WriteString(buf.String())
	if nil != err {
		panic(err)
	}

	return nil
}

func searchInFile(source afero.File, searchFor []string) (int, error) {
	matches := 0

	scanner := bufio.NewScanner(source)
	for scanner.Scan() {
		line := scanner.Text()
		for _, search := range searchFor {
			if strings.Contains(line, search) {
				matches = matches + 1
			}
		}
	}

	return matches, nil
}

// return false if config and zot-config has drifted
func (ec *EasykubeConfig) IsZotConfigInSync(configData *core.EasykubeConfigData) (bool, error) {

	configDir, err := ec.ek.OsDetails.GetEasykubeConfigDir()
	if nil != err {
		return false, err
	}

	// if the zot-config is not generated yet, return false
	_, err = ec.ek.Fs.Stat(filepath.Join(configDir, constants.ZotConfig))
	if err != nil {
		return false, nil
	}

	// registries & credentials
	registries := make([]string, 0)
	credentials := make([]string, 0)

	for _, reg := range configData.MirrorRegistries {
		registries = append(registries, reg.RegistryUrl)

		if reg.UserKey != "" {
			credentials = append(credentials, reg.RegistryUrl)
		}

	}

	// compare with zot credentials
	zotConfig, err := ec.ek.Fs.Open(filepath.Join(configDir, constants.ZotConfig))
	defer func(zotConfig afero.File) {
		_ = zotConfig.Close()
	}(zotConfig)

	zotCredentials, err := ec.ek.Fs.Open(filepath.Join(configDir, constants.ZotCredentials))
	defer func(zotCredentials afero.File) {
		_ = zotCredentials.Close()
	}(zotCredentials)

	// config
	configMatches, err := searchInFile(zotConfig, []string{"urls"})
	credentialMatches, err := searchInFile(zotCredentials, credentials)

	inSync := len(registries) == configMatches && len(credentials) == credentialMatches

	return inSync, err

}

func (ec *EasykubeConfig) GenerateZotRegistryConfig(cfg *core.EasykubeConfigData) error {

	zotRegistry := `
	      {
	        "urls": [
   	          "%s"
	         ],
	         "onDemand": true,
	         "content": [{"prefix": "**"}]
	      }`

	if len(cfg.MirrorRegistries) > 0 {
		// make registries
		registries := make([]string, 0, len(cfg.MirrorRegistries))

		for _, reg := range cfg.MirrorRegistries {
			registries = append(registries, fmt.Sprintf(zotRegistry, reg.RegistryUrl))
		}

		configTemplate, err := resources.AppResources.ReadFile("data/zot-config.json.template")
		if err != nil {
			return err
		}
		templ := template.Must(template.New("zot-config").Funcs(template.FuncMap{
			"join": strings.Join, // Add strings.Join as a helper
		}).Parse(string(configTemplate)))

		buf := &bytes.Buffer{}

		err = templ.Execute(buf, registries)
		if err != nil {
			return err
		}

		configDir, _ := ec.ek.OsDetails.GetEasykubeConfigDir()
		zotconfig := filepath.Join(configDir, constants.ZotConfig)

		file, err := ec.ek.Fs.OpenFile(zotconfig, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		defer func(file afero.File) {
			_ = file.Close()
		}(file)

		if nil != err {
			return err
		}

		_, err = file.WriteString(buf.String())
		if nil != err {
			return err
		}

	}

	return nil
}

func (ec *EasykubeConfig) GenerateZotRegistryCredentials(cfg *core.EasykubeConfigData) error {

	if len(cfg.MirrorRegistries) > 0 {

		mirrorsWithCredentials := make([]core.MirrorRegistry, 0)

		for _, reg := range cfg.MirrorRegistries {
			if reg.PasswordKey != "" {
				mirrorsWithCredentials = append(mirrorsWithCredentials, reg)
			}
		}

		configTemplate, err := resources.AppResources.ReadFile("data/zot-credentials.json.template")
		if err != nil {
			return err
		}
		templ := template.Must(template.New("zot-credentials").Funcs(template.FuncMap{
			"sub": func(a, b int) int { return a - b }, // Add the sub function
		}).Parse(string(configTemplate)))

		buf := &bytes.Buffer{}
		err = templ.Execute(buf, mirrorsWithCredentials)
		if err != nil {
			return err
		}

		configDir, _ := ec.ek.OsDetails.GetEasykubeConfigDir()
		zotCredentials := filepath.Join(configDir, constants.ZotCredentials)

		file, err := ec.ek.Fs.OpenFile(zotCredentials, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if nil != err {
			panic(err)
		}

		_, err = file.WriteString(buf.String())

		if nil != err {
			return err
		}
	}

	return nil
}
