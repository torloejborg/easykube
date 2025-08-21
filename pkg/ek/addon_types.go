package ek

import "os"

type PortConfig struct {
	NodePort int    `json:"nodePort"`
	HostPort int    `json:"hostPort"`
	Protocol string `json:"protocol"`
}

type MountConfig struct {
	PersistenceDir string `json:"persistenceDir"`
	HostPath       string `json:"hostPath"`
	ContainerPath  string `json:"containerPath"`
}

type AddonConfig struct {
	Description string        `json:"description"`
	DependsOn   []string      `json:"dependsOn"`
	ExtraPorts  []PortConfig  `json:"extraPorts"`
	ExtraMounts []MountConfig `json:"extraMounts"`
}

type Addon struct {
	// Addon name including .ek.js
	Name string
	// Addon name excluding .ek.js
	ShortName string
	// Configuration for the addon
	Config AddonConfig
	// Addon javascript file
	File *os.File
	// Root of the addon directory
	RootDir string
}

func (a *Addon) ReadScriptFile() string {
	val, err := os.ReadFile(a.File.Name())
	if err != nil {
		panic(err)
	}
	return string(val)
}
