package ez

import (
	"context"
)

type PodmanImpl struct {
	PodmanContext context.Context
	Common        ContainerRuntimeCommon
}

func NewPodmanImpl() IContainerRuntime {
	/*
		conn, err := bindings.NewConnection(context.Background(), "unix://run/user/1000/podman/podman.sock")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	*/
	return &PodmanImpl{
		PodmanContext: nil,
		Common:        ContainerRuntimeCommon{},
	}
}

func (cr *PodmanImpl) IsClusterRunning() bool {
	return false
}

func (cr *PodmanImpl) IsNetworkConnectedToContainer(containerID string, networkID string) bool {
	return false
}

func (cr *PodmanImpl) IsContainerRunning(containerID string) bool {
	return false
}

func (cr *PodmanImpl) HasImage(image string) bool {

	/*
		searchReport, _ := images.List(cr.PodmanContext, nil)

		for _, x := range searchReport {
			for i := range x.Names {
				if strings.Contains(x.Names[i], image) {
					return true
				}
			}
		}*/

	return false
}

func (cr *PodmanImpl) HasImageInKindRegistry(image string) bool {
	return cr.Common.ImageExistsInKindRegistry(image)
}

func (cr *PodmanImpl) PushImage(image string) {

	/*
		if !cr.HasImageInKindRegistry(image) {

			opts := images.PushOptions{
				SkipTLSVerify: ptr.To(true),
			}
			err := images.PushImage(cr.PodmanContext, image, image, &opts)
			if nil != err {
				log.Fatalln(err)
			}
		}*/
}

func (cr *PodmanImpl) PullImage(image string, privateRegistryCredentials *string) {

	/*
		opts := images.PullOptions{}

		_, err := images.PullImage(cr.PodmanContext, image, &opts)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}*/
}

func (cr *PodmanImpl) FindContainer(containerName string) ContainerSearch {
	/*
		opts := containers.ListOptions{
			All: ptr.To(true),
		}

		containerList, err := containers.List(cr.PodmanContext, &opts)

		if err != nil {
			log.Panic(err)
		}

		for _, x := range containerList {

			for i := range x.Names {
				if strings.Contains(x.Names[i], containerName) {
					return ContainerSearch{
						IsRunning:   strings.Contains(x.State, "running"),
						Found:       true,
						ContainerID: x.ID,
					}
				}
			}
		}
	*/
	return ContainerSearch{
		Found:     false,
		IsRunning: false,
	}
}

func (cr *PodmanImpl) StartContainer(id string) {
	/*
		opts := containers.StartOptions{}

		err := containers.Start(cr.PodmanContext, id, &opts)
		if nil != err {
			panic(err)
		}*/
}

func (cr *PodmanImpl) StopContainer(id string) {
	/*
		opts := containers.StopOptions{}

		err := containers.Stop(cr.PodmanContext, id, &opts)
		if nil != err {
			panic(err)
		}*/
}

func (cr *PodmanImpl) RemoveContainer(id string) {
	/*
		opts := containers.RemoveOptions{}

		_, err := containers.Remove(cr.PodmanContext, id, &opts)
		if nil != err {
			panic(err)
		}
	*/
}

func (cr *PodmanImpl) Exec(containerId string, cmd []string) {
}

func (cr *PodmanImpl) ContainerWriteFile(containerId string, dst string, filename string, data []byte) {
}

func (cr *PodmanImpl) NetworkConnect(containerId string, networkId string) {
}

func (cr *PodmanImpl) CloseContainerRuntime() {
}

func (cr *PodmanImpl) IsContainerRuntimeAvailable() bool {
	return false
}

func (cr *PodmanImpl) CreateContainerRegistry() {
	/*
		// If we alreday have the image don't pull it

		if !cr.HasImage(constants.REGISTRY_IMAGE) {
			cr.PullImage(constants.REGISTRY_IMAGE)
		}

		// Does the container exists, if not create it
		search := cr.FindContainer("kind-registry")
		if !search.Found && !search.IsRunning {
			CopyResource("registry-config.yaml", "registry-config.yaml")
			CopyResource("cert/server.crt", "localtest.me.crt")
			CopyResource("cert/server.key", "localtest.me.key")

			s := specgen.NewSpecGenerator(constants.REGISTRY_IMAGE, false)
			s.Name = "kind-registry"
			configDir, _ := os.UserConfigDir()

			s.Mounts = []specs.Mount{
				{
					Source:      filepath.Join(configDir, "easykube", "registry-config.yaml"),
					Destination: "/etc/docker/registry/config.yml",
					Type:        "bind",
				},
				{
					Source:      filepath.Join(configDir, "easykube", "localtest.me.crt"),
					Destination: "/etc/ssl/localtest.me.crt",
					Type:        "bind",
				},
				{
					Source:      filepath.Join(configDir, "easykube", "localtest.me.key"),
					Destination: "/etc/ssl/localtest.me.key",
					Type:        "bind",
				},
			}

			s.PortMappings = []types.PortMapping{
				{
					ContainerPort: constants.LOCAL_REGISTRY_PORT,
					HostPort:      constants.LOCAL_REGISTRY_PORT,
				},
			}

			create, err := containers.CreateWithSpec(cr.PodmanContext, s, nil)

			if nil != err {
				log.Fatalln(err)
			}

			search.ContainerID = create.ID
			search.Found = true
			search.IsRunning = false
		}

		if !search.IsRunning {
			cr.StartContainer(search.ContainerID)
		}*/

}

func (cr *PodmanImpl) Commit(containerID string) {

}

func (cr *PodmanImpl) TagImage(source, target string) {

}
