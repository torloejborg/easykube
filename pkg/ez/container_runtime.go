package ez

type ContainerSearch struct {
	ContainerID string
	Found       bool
	IsRunning   bool
}

type ImageSearch struct {
	SHA256 string
	Found  bool
}

type IContainerRuntime interface {
	IsContainerRunning(containerID string) bool
	PushImage(image string)
	PullImage(image string, privateRegistryCredentials *string)
	HasImage(image string) bool
	TagImage(source, target string)
	FindContainer(name string) (*ContainerSearch, error)
	StartContainer(id string)
	StopContainer(id string)
	RemoveContainer(id string)
	ContainerWriteFile(containerId string, dst string, filename string, data []byte)
	NetworkConnect(containerId string, networkId string)
	IsNetworkConnectedToContainer(containerID string, networkID string) bool
	IsClusterRunning() bool
	HasImageInKindRegistry(name string) bool
	Exec(containerId string, cmd []string)
	CloseContainerRuntime()
	IsContainerRuntimeAvailable() bool
	CreateContainerRegistry() error
	Commit(containerID string)
}

func NewContainerRuntime() IContainerRuntime {

	_, err := Kube.LoadConfig()
	if err != nil {
		panic(err)
	}

	return NewPodmanImpl()

}
