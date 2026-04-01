// Package core
//
// defines all interfaces and central functionality for easykube
package core

import (
	"time"

	"github.com/spf13/afero"
)

// IAddon something that can be added, listed and removed from easykube
type IAddon interface {
	// ReadScriptFile Returns the <addon>.ek.js javascript file as string
	ReadScriptFile(fs afero.Fs) string

	// GetName Name of the addon, this will be <name>.ek.js
	GetName() string

	// GetShortName Name of the addon, this will be <name>
	GetShortName() string

	// GetConfig The addon's configuration - it's name, what other addons it depends on,
	// filesystem mounts and exposed ports
	GetConfig() AddonConfig

	// GetAddonFile Get the path to the addon *.ek.js file
	GetAddonFile() string

	// GetRootDir Get the directory of the addon
	GetRootDir() string

	// GetDependencies Get a list of named dependencies to other addons
	GetDependencies() []string
}

// IAddonReader Handles reading and discovery of addons
type IAddonReader interface {
	// GetAddons Returns a map keyed by addon name and the addon struct
	GetAddons() (map[string]IAddon, error)

	// ExtractConfiguration Addons always declare a "configuration" stanza, this function evaluates the script file
	// in the JS runtime and extracts the configuration to a struct
	ExtractConfiguration(unconfigured IAddon) (*AddonConfig, error)

	// CheckAddonCompatibility Determine if this addon is compatible with the JS script library functions
	CheckAddonCompatibility() (string, error)
}

// IClusterUtils Manage kind cluster creation
type IClusterUtils interface {

	// CreateKindCluster creates a kind cluster, based on the addons provided, a kind cluster configuration is generated,
	// and the cluster booted with that custom configuration
	CreateKindCluster(modules map[string]IAddon) (string, error)

	// RenderToYAML Creates the actual kind cluster configuration text
	RenderToYAML(addonList []IAddon, config *EasykubeConfigData) string

	// ConfigurationReport Returns a report of how addons contributed to the cluster configuration
	ConfigurationReport(addonList []IAddon) string

	// EnsurePersistenceDirectories EnsurePersistenceDirectory Based on discovered addons, create directories used in kinds
	// extraMounts config stanza
	EnsurePersistenceDirectories() error
}

// ICobraCommandHelper Handles common flags specified on the command line and shortcuts to get a flag of a given type
type ICobraCommandHelper interface {
	// GetBoolFlag bool value for a given flag
	// see constants/flags.go
	GetBoolFlag(name string) bool
	GetStringFlag(name string) string
	GetIntFlag(name string) int
	IsVerbose() bool
	IsDryRun() bool
}

type IEasykubeConfig interface {
	LoadConfig() (*EasykubeConfigData, error)
	MakeConfig() error
	EditConfig() error
	LaunchEditor(config, editor string)
	PathToConfigFile() string
	IsZotConfigInSync(configData *EasykubeConfigData) (bool, error)
	GenerateZotRegistryConfig(*EasykubeConfigData) error
	GenerateZotRegistryCredentials(*EasykubeConfigData) error
	WriteConfig(*EasykubeConfigData) error
	CopyConfigResources() error
	HasConfiguration() bool
}

type IContainerRuntime interface {
	IsContainerRunning(containerID string) (bool, error)
	PushImage(src, image string) error
	PullImage(image string, credentials *PrivateRegistryCredentials) error
	HasImage(image string) (bool, error)
	TagImage(source, target string) error
	FindContainer(name string) (*ContainerSearch, error)
	StartContainer(id string) error
	StopContainer(id string) error
	RemoveContainer(id string) error
	ContainerWriteFile(containerId string, dst string, filename string, data []byte) error
	NetworkConnect(containerId string, networkId string) error
	IsNetworkConnectedToContainer(containerID string, networkID string) (bool, error)
	IsClusterRunning() bool
	HasImageInKindRegistry(name string) (bool, error)
	Exec(containerId string, cmd []string) error
	CloseContainerRuntime()
	IsContainerRuntimeAvailable() bool
	CreateContainerRegistry() error
	StartContainerRegistry() error
	Commit(containerID string)
}

type IJsUtils interface {
	ExecAddonScript(a IAddon) error
	ExtractConfigurationObject(a IAddon) (*AddonConfig, error)
}

type IK8SUtils interface {
	// ReloadClientSet Reloads the kubernetes client
	//
	// After creating a kind cluster a new kubeconfig is generated, this function make
	// sure we operate on the current one
	ReloadClientSet() error
	// CreateSecret
	//
	// Creates a kubernetes secret, kubernetes likes to base64 encode secrets, but a map of plain
	// strings can be passed, some fuzzy check is done to detect if encoding should happen or not. This is
	// not 100% perfect, and can fail in some instances - A good candidate for rewriting.
	CreateSecret(namespace, secretName string, data map[string]string) error
	//GetSecret Fetches a secret from kubernetes
	//
	// Returns
	GetSecret(name, namespace string) (map[string][]byte, error)
	// CreateConfigmap
	//
	// Creates an empty configmap
	CreateConfigmap(name, namespace string) error

	// DeleteKeyFromConfigmap
	//
	// Removes a key from a configmap and saves it
	DeleteKeyFromConfigmap(name, namespace, key string)
	// ReadConfigmap
	//
	// Fetch a configmap from kubernetes
	ReadConfigmap(name string, namespace string) (map[string]string, error)
	// UpdateConfigMap
	//
	// Updates a key in a configmap
	UpdateConfigMap(name, namespace, key string, data []byte)
	// HasKubeConfig
	//
	// true if K8sUtils are ready to go - the clientset and rest client are initialized.
	HasKubeConfig() bool

	// FindContainerInPod
	//
	// Attempts to locate a container in a given deployment, if there are multiple containers matching
	// containerPartialName, the first is matched. Returns a triplet of (pod.Name, container.Name, error)
	FindContainerInPod(deploymentName, namespace, containerPartialName string) (string, string, error)
	// ExecInPod
	//
	// Runs a 'kubectl exec' using the API, returns stdout and stderr as strings
	ExecInPod(namespace, pod, command string, args []string) (string, string, error)

	// CopyFileToPod
	//
	// Copies a local file into a remote pod
	CopyFileToPod(namespace, pod, container, localPath, remotePath string) error
	// ListPods
	//
	// Gets all pod names in a namespace
	ListPods(namespace string) ([]string, error)
	// GetInstalledAddons
	//
	// Queries the constants.AddonCm configmap for installed addons
	GetInstalledAddons() ([]string, error)

	// PatchCoreDNS
	//
	// Teaches CoreDNS to understand the localtest.me loopback domain
	PatchCoreDNS()

	// WaitForDeploymentReadyWatch pauses execution until some deployment enters ready state
	WaitForDeploymentReadyWatch(name, namespace string) error

	// WaitForCRD pauses execution for a duration of time until some custom resources have appeared in a cluster
	//
	// group, version, kind are the coordinates for the CRD, timeout specifies wait time
	WaitForCRD(group, version, kind string, timeout time.Duration) error
	// TransformExternalSecret creates a kubernetes secret from an external secret definition (ExternalSecretOperator)
	//
	// secret is the source which contains keys and references, mockData specifies the replacements for the keys,
	// addonName is the source of the mockData, namespace specifies where to create the secret
	TransformExternalSecret(secret ExternalSecret, mockData map[string]map[string]string, addonName, namespace string) KubernetesSecret

	WaitForKindClusterReady(kubeconfig string, timeout time.Duration) error

	RestartDeployment(deploymentName, namespace string) error
}

type IExternalTools interface {
	KustomizeBuild(dir string) string
	ApplyYaml(yamlFile string)
	DeleteYaml(yamlFile string)
	EnsureLocalContext()
	// SwitchContext Change kube context to name
	SwitchContext(name string)
	// RunCommand Runs an OS command
	RunCommand(name string, args ...string) (stdout string, stderr string, err error)
}

type IOsDetails interface {
	GetEasykubeConfigDir() (string, error)
	GetUserHomeDir() (string, error)
}

type IPrinter interface {
	FmtGreen(out string, args ...any)
	FmtRed(out string, args ...any)
	FmtYellow(out string, args ...any)
	FmtVerbose(out string, args ...any)
	FmtDryRun(out string, args ...any)
}

type ISkaffold interface {
	CreateNewAddon(name, dest string)
}

type IStatusBuilder interface {
	DoContainerCheck() error
	DoBinaryCheck() error
	DoAddonRepositoryCheck() error
	GetDockerVersion() string
	GetHelmVersion() string
	GetKubectlVersion() string
	GetKustomizeVersion() string
	GetPodmanVersion() string
	GetVersionStr(in, wants string, inErr error) string
}

type IUtils interface {
	HasBinary(bin string) bool
	FileOrDirExists(path string) bool
	CopyResourceToConfigDir(src, dest string) error
	SaveFile(data string, dest string)
	SaveFileByte(data []byte, dest string)
	ReadPropertyFile(path string) (map[string]string, error)
	IntSliceToStrings(input []int) []string
	ReadFileToBytes(filename string) ([]byte, error)
}

type EzNode interface {
	GetName() string
	GetDependencies() []string
}

type IOrderedTask interface {
	GetName() string
	GetDependencies() []string
}
