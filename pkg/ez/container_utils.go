package ez

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/torloejborg/easykube/pkg/constants"
)

type ContainerSearch struct {
	ContainerID string
	Found       bool
	IsRunning   bool
}

type ImageSearch struct {
	SHA256 string
	Found  bool
}

type ContainerRuntimeCommon struct {
}

func (i *ContainerRuntimeCommon) ImageExistsInKindRegistry(image string) bool {
	image = strings.ReplaceAll(image, constants.LOCAL_REGISTRY+"/", "")
	parts := strings.Split(image, ":")
	imageWithoutTag := parts[0]
	imageTag := parts[1]

	type TagList struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(fmt.Sprintf("http://%s/v2/%s/tags/list", constants.LOCAL_REGISTRY, imageWithoutTag))
	if nil != err {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if nil != err {
		log.Fatalln(err)
	}

	var dat TagList

	if jsonerr := json.Unmarshal(body, &dat); jsonerr != nil {
		panic(jsonerr)
	}

	if strings.Contains(dat.Name, imageWithoutTag) {
		for i := range dat.Tags {
			if dat.Tags[i] == imageTag {
				return true
			}
		}
	}

	return false
}

type ContainerImageManager interface {
	IsContainerRunning(containerID string) bool
	PushImage(image string)
	PullImage(image string, privateRegistryCredentials *string)
	HasImage(image string) bool
	TagImage(source, target string)
}

type ContainerContainerManager interface {
	FindContainer(name string) ContainerSearch
	StartContainer(id string)
	StopContainer(id string)
	RemoveContainer(id string)
	ContainerWriteFile(containerId string, dst string, filename string, data []byte)
}

type ContainerNetworkManager interface {
	NetworkConnect(containerId string, networkId string)
	IsNetworkConnectedToContainer(containerID string, networkID string) bool
}

type IContainerRuntime interface {
	ContainerImageManager
	ContainerContainerManager
	ContainerNetworkManager

	IsClusterRunning() bool

	HasImageInKindRegistry(name string) bool

	Exec(containerId string, cmd []string)

	CloseContainerRuntime()
	IsContainerRuntimeAvailable() bool
	CreateContainerRegistry()
	Commit(containerID string)
}

func NewContainerRuntime() IContainerRuntime {

	cfg, err := Kube.LoadConfig()
	if err != nil {
		panic(err)
	}
	if cfg.ContainerRuntime == "podman" {
		return NewPodmanImpl()
	} else {
		return NewDockerImpl()
	}
}
