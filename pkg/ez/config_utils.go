package ez

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/resources"
)

type PrivateRegistry struct {
	RepositoryURL string `mapstructure:"repository-url"`
	UserKey       string `mapstructure:"username"`
	PasswordKey   string `mapstructure:"password"`
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
	PathToConfigFile() string
	SyncWithZot() error
	WriteConfig(*EasykubeConfigData) error
	CopyConfigResources() error
}

type EasykubeConfig struct {
	ConfigDirName string
	UserConfigDir string
}

func NewEasykubeConfig(os OsDetails) IEasykubeConfig {
	return &EasykubeConfig{}
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

	configDir, _ := Kube.GetEasykubeConfigDir()
	return filepath.Join(configDir, "config.yaml")
}

func (ec *EasykubeConfig) LoadConfig() (*EasykubeConfigData, error) {

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
		ec.LaunchEditor(ec.PathToConfigFile(), editor)
	}
}

func (ec *EasykubeConfig) CopyConfigResources() error {

	easykubeConfigDir, err := Kube.GetEasykubeConfigDir()
	if nil != err {
		return err
	}

	merr := Kube.Fs.MkdirAll(easykubeConfigDir, os.ModePerm)
	if merr != nil {
		return merr
	}

	userHomeDir, err := Kube.GetUserHomeDir()
	if nil != err {
		return err
	}

	pathToConfigFile := ec.PathToConfigFile()
	_, err = Kube.Fs.Stat(pathToConfigFile)

	err = CopyResourceToConfigDir("cert/localtest.me.crt", "localtest.me.crt")
	if nil != err {
		return err
	}

	err = CopyResourceToConfigDir("cert/localtest.me.ca.crt", "localtest.me.ca.crt")
	if nil != err {
		return err
	}

	err = CopyResourceToConfigDir("cert/localtest.me.key", "localtest.me.key")
	if nil != err {
		return err
	}

	err = CopyResourceToConfigDir("zot-config.json", "zot-config.json")
	if nil != err {
		return err
	}

	err = CopyResourceToConfigDir("zot-credentials.json", "zot-credentials.json")
	if nil != err {
		return err
	}

	// Make podman aware of a selfsigned certificate
	certData, err := resources.AppResources.ReadFile("data/cert/localtest.me.ca.crt")
	if nil != err {
		return err
	}

	certDestDir := filepath.Join(userHomeDir, ".config", "containers", "certs.d", constants.LocalRegistry)
	_ = Kube.Fs.MkdirAll(certDestDir, os.ModePerm)
	cert := filepath.Join(certDestDir, "ca.crt")
	SaveFileByte(certData, cert)

	return nil
}

func (ec *EasykubeConfig) MakeConfig() error {

	userHomeDir, err := Kube.GetUserHomeDir()
	configurationDir, err := Kube.GetEasykubeConfigDir()
	if nil != err {
		return err
	}

	pathToConfigFile := ec.PathToConfigFile()
	_, err = Kube.Fs.Stat(pathToConfigFile)

	_ = ec.CopyConfigResources()

	model := EasykubeConfigData{
		AddonDir:         filepath.Join(userHomeDir, "addons"),
		ConfigurationDir: configurationDir,
		PersistenceDir:   filepath.Join(configurationDir, "persistence"),
		ContainerRuntime: "docker",
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

	return nil
}

func (ec *EasykubeConfig) WriteConfig(cfg *EasykubeConfigData) error {
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

	fmt.Println(cfg)

	err = templ.Execute(buf, cfg)
	if err != nil {
		return err
	}

	fmt.Println(buf.String())

	file, err := Kube.Fs.OpenFile(cfg.ConfigurationFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if nil != err {
		panic(err)
	}

	_, err = file.WriteString(buf.String())
	if nil != err {
		panic(err)
	}

	return nil
}

func (ec *EasykubeConfig) SyncWithZot() error {
	return nil
}
