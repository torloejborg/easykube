package core

import (
	"sync"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type Ek struct {
	AddonReader      IAddonReader
	ClusterUtils     IClusterUtils
	CommandContext   ICobraCommandHelper
	Config           IEasykubeConfig
	ContainerRuntime IContainerRuntime
	ExternalTools    IExternalTools
	Fs               afero.Fs
	Kubernetes       IK8SUtils
	OsDetails        IOsDetails
	Printer          IPrinter
	Utils            IUtils
	Status           IStatusBuilder
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
	// Addon JavaScript file, the absolute path
	File string
	// Root of the addon directory
	RootDir string
	// The addon's immediate dependencies
	Dependencies []string
}

type CobraCommandHelperImpl struct {
	Command *cobra.Command
}

type Task struct {
	Name          string
	Description   string
	Dependencies  []string
	SkipCondition func() bool
	Execute       func() error
	graph         *Graph[Task]
}

type TaskContainer struct {
	graph *Graph[Task]
	mu    sync.Mutex // Mutex to synchronize spinner access
}

type Graph[T EzNode] struct {
	Nodes    []T
	adj      map[string][]T // Key: node ID (string), Value: list of dependent nodes
	inDegree map[string]int // Key: node ID (string), Value: in-degree count
	idToNode map[string]T   // Key: node ID (string), Value: node
}

type Stack[T any] []T

type MirrorRegistry struct {
	RegistryUrl string `mapstructure:"registry-url"`
	UserKey     string `mapstructure:"username"`
	PasswordKey string `mapstructure:"password"`
}

type EasykubeConfigData struct {
	AddonDir          string `mapstructure:"addon-root"`
	PersistenceDir    string `mapstructure:"persistence-dir"`
	ConfigurationDir  string `mapstructure:"config-dir"`
	ContainerRuntime  string `mapstructure:"container-runtime"`
	ConfigurationFile string
	MirrorRegistries  []MirrorRegistry `mapstructure:"mirror-registries"`
}

// KubernetesSecret represents the structure of a Kubernetes Secret resource.
type KubernetesSecret struct {
	ApiVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Metadata   map[string]string `yaml:"metadata"`
	Data       map[string]string `yaml:"data"`
	Type       string            `yaml:"type"`
}

// ExternalSecret Defines a struct to capture the structure of an ExternalSecret from the ExternalSecretOperator
type ExternalSecret struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Spec struct {
		// Define the structure of the spec according to your needs
		RefreshInterval string `yaml:"refreshInterval"`
		SecretStoreRef  struct {
			Name string `yaml:"name"`
			Kind string `yaml:"kind"`
		} `yaml:"secretStoreRef"`
		Data []struct {
			SecretKey string `yaml:"secretKey"`
			RemoteRef struct {
				Key      string `yaml:"key"`
				Property string `yaml:"property"`
			} `yaml:"remoteRef"`
		} `yaml:"data"`
	} `yaml:"spec"`
}

type ContainerSearch struct {
	ContainerID string
	Found       bool
	IsRunning   bool
}

type PrivateRegistryCredentials struct {
	Username string
	Password string
}

type PortConfig struct {
	NodePort int    `json:"nodePort"`
	HostPort int    `json:"hostPort"`
	Protocol string `json:"protocol"`
}
