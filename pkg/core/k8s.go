package core

import "time"

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
