package core

import (
	"github.com/spf13/afero"
)

func (a Addon) ReadScriptFile(fs afero.Fs) string {
	val, err := afero.ReadFile(fs, a.File)
	if err != nil {
		panic(err)
	}
	return string(val)
}

func (a Addon) GetName() string {
	return a.Name
}

func (a Addon) GetShortName() string {
	return a.ShortName
}

func (a Addon) GetConfig() AddonConfig {
	return a.Config
}

func (a Addon) GetAddonFile() string {
	return a.File
}

func (a Addon) GetRootDir() string {
	return a.RootDir
}

func (a Addon) GetDependencies() []string {
	return a.Dependencies
}
