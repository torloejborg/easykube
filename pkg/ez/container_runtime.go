package ez

type ContainerSearch struct {
	ContainerID string
	Found       bool
	IsRunning   bool
}

type ContainerConnection struct {
	Host string
	Type string
}

type ImageSearch struct {
	SHA256 string
	Found  bool
}

type PrivateRegistryCredentials struct {
	Username string
	Password string
}

type IContainerRuntime interface {
	DiscoverContainerRuntimeConnection(runtime string) (ContainerConnection, error)
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
	Commit(containerID string)
}

func NewContainerRuntime() (IContainerRuntime, error) {

	cfg, err := Kube.LoadConfig()
	if err != nil {
		panic(err)
	}

	return NewContainerRuntimeImpl(cfg.ContainerRuntime)
}
