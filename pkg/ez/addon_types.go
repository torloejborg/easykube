package ez

import (
	"encoding/json"

	"github.com/spf13/afero"
)

type PortConfig struct {
	NodePort int    `json:"nodePort"`
	HostPort int    `json:"hostPort"`
	Protocol string `json:"protocol"`
}

func (pc *PortConfig) UnmarshalJSON(data []byte) error {
	// Using a helper struct to avoid recursion during unmarshaling
	type Alias PortConfig
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(pc),
	}

	// Set default values
	aux.NodePort = 0     // Assuming 0 is acceptable as a default
	aux.HostPort = 0     // Assuming 0 is acceptable as a default
	aux.Protocol = "TCP" // Default protocol set to "TCP"

	// Unmarshal JSON data into the aux struct
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Ensure Protocol is set if not present in JSON
	if aux.Protocol == "" {
		aux.Protocol = "TCP"
	}

	return nil
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
	File string
	// Root of the addon directory
	RootDir string
}

func (a *Addon) ReadScriptFile(fs afero.Fs) string {
	val, err := afero.ReadFile(fs, a.File)
	if err != nil {
		panic(err)
	}
	return string(val)
}
