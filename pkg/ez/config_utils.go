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
	"github.com/torloejborg/easykube/pkg/resources"
)

type EasykubeConfigData struct {
	AddonDir          string `mapstructure:"addon-root"`
	PersistenceDir    string `mapstructure:"persistence-dir"`
	ConfigurationDir  string `mapstructure:"config-dir"`
	ContainerRuntime  string `mapstructure:"container-runtime"`
	ConfigurationFile string
}

type IEasykubeConfig interface {
	LoadConfig() (*EasykubeConfigData, error)
	MakeConfig() error
	EditConfig()
	LaunchEditor(config, editor string)
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

func (ec *EasykubeConfig) LoadConfig() (*EasykubeConfigData, error) {
	// if no config file, create a default
	ec.MakeConfig()

	config.ClearAll()
	// config.ParseEnv: will parse env var in string value. eg: shell: ${SHELL}
	config.WithOptions(config.ParseEnv)

	// add driver for support yaml content
	config.AddDriver(yaml.Driver)
	configDir, _ := Kube.GetUserConfigDir()

	data, err := ReadFileToBytes(filepath.Join(configDir, ec.ConfigDirName, "config.yaml"))
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
		configDir, err := os.UserConfigDir()
		pathToConfig := filepath.Join(configDir, ec.ConfigDirName, "config.yaml")

		if err != nil {
			fmt.Println("Could not determine $HOMEDIR")
			os.Exit(-1)
		}

		ec.MakeConfig()
		ec.LaunchEditor(pathToConfig, editor)

	}
}

func (ec *EasykubeConfig) MakeConfig() error {
	userConfigDir, err := Kube.GetUserConfigDir()
	if nil != err {
		return err
	}

	pathToConfigFile := filepath.Join(userConfigDir, ec.ConfigDirName, "config.yaml")
	_, err = Kube.Fs.Stat(pathToConfigFile)

	if os.IsNotExist(err) {
		Kube.Fs.MkdirAll(filepath.Join(userConfigDir, "easykube"), os.ModePerm)

		model := EasykubeConfigData{
			AddonDir:         "./addons",
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

		err = CopyResource("cert/localtest.me.key", "localtest.me.key")
		if nil != err {
			return err
		}

		err = CopyResource("registry-config.yaml", "registry-config.yaml")
		if nil != err {
			return err
		}

	}

	return nil
}
