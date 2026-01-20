package ez

import (
	"bytes"
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
	"github.com/torloejborg/easykube/pkg/resources"
	"github.com/torloejborg/easykube/pkg/textutils"
)

type PrivateRegistry struct {
	RepositoryMatch string
	UserKey         string
	PasswordKey     string
}

type EasykubeConfigData struct {
	AddonDir          string `mapstructure:"addon-root"`
	PersistenceDir    string `mapstructure:"persistence-dir"`
	ConfigurationDir  string `mapstructure:"config-dir"`
	ContainerRuntime  string `mapstructure:"container-runtime"`
	ConfigurationFile string
	PrivateRegistries []PrivateRegistry `mapstructure:"private-registries"`
}

type IEasykubeConfig interface {
	LoadConfig() (*EasykubeConfigData, error)
	MakeConfig() error
	EditConfig()
	LaunchEditor(config, editor string)
	PatchConfig() error
	PathToConfigFile() string
}

type EasykubeConfig struct {
	ConfigDirName string
	UserConfigDir string
}

func NewEasykubeConfig(os OsDetails) IEasykubeConfig {
	configDir, _ := os.GetUserConfigDir()
	return &EasykubeConfig{
		UserConfigDir: configDir,
		ConfigDirName: "easykube",
	}
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
	configDir, _ := Kube.GetUserConfigDir()
	return filepath.Join(configDir, ec.ConfigDirName, "config.yaml")
}

func (ec *EasykubeConfig) LoadConfig() (*EasykubeConfigData, error) {
	// if no config file, create a default
	if err := ec.MakeConfig(); err != nil {
		return nil, err
	}

	config.ClearAll()
	// config.ParseEnv: will parse env var in string value. eg: shell: ${SHELL}
	config.WithOptions(config.ParseEnv)

	// add driver for support yaml content
	config.AddDriver(yaml.Driver)

	data, err := ReadFileToBytes(ec.PathToConfigFile())
	if err != nil {
		return nil, err
	}

	err = config.LoadSources("yaml", data)
	if err != nil {
		return nil, err
	}

	easykube := EasykubeConfigData{}
	err = config.BindStruct("easykube", &easykube)

	if err != nil {
		return nil, err
	}

	return &easykube, nil
}

func (ec *EasykubeConfig) EditConfig() {
	editor := os.Getenv("VISUAL")
	if len(editor) == 0 {
		fmt.Println("VISUAL environment variable not set")
		os.Exit(-1)
	} else {
		_, err := os.UserConfigDir()

		if err != nil {
			fmt.Println("Could not determine $HOMEDIR")
			os.Exit(-1)
		}

		if err := ec.MakeConfig(); err == nil {
			ec.LaunchEditor(ec.PathToConfigFile(), editor)
		} else {
			Kube.FmtRed("Failed to launch editor: %s", err.Error())
		}
	}
}

func (ec *EasykubeConfig) MakeConfig() error {
	userConfigDir, err := Kube.GetUserConfigDir()
	if nil != err {
		return err
	}

	userHomeDir, err := Kube.GetUserHomeDir()
	if nil != err {
		return err
	}

	pathToConfigFile := ec.PathToConfigFile()
	_, err = Kube.Fs.Stat(pathToConfigFile)

	if os.IsNotExist(err) {
		merr := Kube.Fs.MkdirAll(filepath.Join(userConfigDir, "easykube"), os.ModePerm)
		if merr != nil {
			return merr
		}

		model := EasykubeConfigData{
			AddonDir:         filepath.Join(userHomeDir, "addons"),
			ConfigurationDir: filepath.Join(ec.UserConfigDir, ec.ConfigDirName),
			PersistenceDir:   filepath.Join(ec.UserConfigDir, ec.ConfigDirName, "persistence"),
		}

		file, err := Kube.Fs.Create(pathToConfigFile)
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

		err = templ.Execute(buf, model)
		if err != nil {
			return err
		}

		_, err = file.WriteString(buf.String())
		if nil != err {
			return err
		}

		err = CopyResource("cert/localtest.me.crt", "localtest.me.crt")
		if nil != err {
			return err
		}

		err = CopyResource("cert/localtest.me.ca.crt", "localtest.me.ca.crt")
		if nil != err {
			return err
		}

		err = CopyResource("cert/localtest.me.key", "localtest.me.key")
		if nil != err {
			return err
		}

		err = CopyResource("registry-config.yaml", "registry-config.yaml")
		if nil != err {
			return err
		}

		// Make podman aware of a selfsigned certificate
		certData, err := resources.AppResources.ReadFile("data/cert/localtest.me.ca.crt")
		if nil != err {
			return err
		}
		certDestDir := filepath.Join(userHomeDir, ".config", "containers", "certs.d", constants.LOCAL_REGISTRY)
		Kube.Fs.MkdirAll(certDestDir, os.ModePerm)
		cert := filepath.Join(certDestDir, "ca.crt")
		SaveFileByte(certData, cert)
	}

	return nil
}

func (ec *EasykubeConfig) PatchConfig() error {
	cfg, _ := ec.LoadConfig()
	ec.patchConfigWithPrivateRegistryTemplate(cfg)

	return nil
}

func (ec *EasykubeConfig) patchConfigWithPrivateRegistryTemplate(cfg *EasykubeConfigData) {

	configStanza := textutils.TrimMargin(`
	  |  #Declare private registries. Whenever easykube pulls an image, and the registry name contains a substring
	  |  #defined by repositoryMatch, the credentials are resolved by looking up the values in easykube-secrets.
	  |  #private-registries:
	  |   #- repositoryMatch: partial-registry-name.io
	  |   #  userKey: userCredentialsKey
	  |   #  passwordKey: userCredentialPasswordKey`, "|")

	cfgFile := ec.PathToConfigFile()

	data, _ := ReadFileToBytes(cfgFile)
	cfgText := string(data)

	if cfg.PrivateRegistries == nil && !strings.Contains(cfgText, "repositoryMatch") &&
		!strings.Contains(cfgText, "userKey") &&
		!strings.Contains(cfgText, "passwordKey") {

		// patch config
		f, err := Kube.Fs.OpenFile(cfgFile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			panic(err)
		}

		defer func(f afero.File) {
			_ = f.Close()
		}(f)

		_, err = f.WriteString(configStanza)
		if err != nil {
			return
		}

		Kube.FmtGreen("Your configuration was patched with the following stanza")
		fmt.Println()
		Kube.FmtGreen(configStanza)
		fmt.Println()
		Kube.FmtGreen("Edit your configuration to enable private registries")
		fmt.Println()
	}
}
