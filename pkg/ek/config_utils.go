package ek

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/torloj/easykube/ekctx"

	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"
	"github.com/torloj/easykube/pkg/resources"
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
	MakeConfig()
	EditConfig()
	LaunchEditor(config, editor string)
}

type EasykubeConfig struct {
	ConfigDirName string
	UserConfigDir string
}

func NewEasykubeConfig(ctx *ekctx.EKContext) IEasykubeConfig {
	configDir, _ := os.UserConfigDir()
	return &EasykubeConfig{UserConfigDir: configDir, ConfigDirName: "easykube"}
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
	configDir := ec.UserConfigDir

	err := config.LoadFiles(filepath.Join(configDir, ec.ConfigDirName, "config.yaml"))
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

func (ec *EasykubeConfig) MakeConfig() {
	userConfigDir, err := os.UserConfigDir()
	if nil != err {
		log.Fatal(err)
	}

	pathToConfigFile := filepath.Join(userConfigDir, ec.ConfigDirName, "config.yaml")
	_, err = os.Stat(pathToConfigFile)

	if os.IsNotExist(err) {
		os.MkdirAll(filepath.Join(userConfigDir, "easykube"), os.ModePerm)

		model := EasykubeConfigData{
			AddonDir:         "./addons",
			ConfigurationDir: filepath.Join(ec.UserConfigDir, ec.ConfigDirName),
			PersistenceDir:   filepath.Join(ec.UserConfigDir, ec.ConfigDirName, "persistence"),
		}

		file, err := os.Create(pathToConfigFile)
		if nil != err {
			log.Panic(err)
		}

		configData, err := resources.AppResources.ReadFile("data/config.template")
		if nil != err {
			log.Panic(err)
		}

		templ := template.New("config")
		templ, err = templ.Parse(string(configData))
		if err != nil {
			panic(err)
		}

		buf := &bytes.Buffer{}

		err = templ.Execute(buf, model)
		if err != nil {
			panic(err)
		}

		_, err = file.WriteString(buf.String())
		if nil != err {
			log.Panic(err)
		}

		CopyResource("cert/localtest.me.crt", "localtest.me.crt")
		CopyResource("cert/localtest.me.key", "localtest.me.key")
		CopyResource("registry-config.yaml", "registry-config.yaml")

	}
}
