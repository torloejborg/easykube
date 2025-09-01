package ek

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/torloj/easykube/ekctx"
	"github.com/torloj/easykube/pkg/constants"
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

type IContainerRuntime interface {
	IsClusterRunning() bool
	IsNetworkConnectedToContainer(containerID string, networkID string) bool
	IsContainerRunning(containerID string) bool
	Push(image string)
	Pull(image string, privateRegistryCredentials *string)
	HasImage(image string) bool
	FindContainer(name string) ContainerSearch
	HasImageInKindRegistry(name string) bool
	StartContainer(id string)
	StopContainer(id string)
	RemoveContainer(id string)
	Exec(containerId string, cmd []string)
	WriteFile(containerId string, dst string, filename string, data []byte)
	NetworkConnect(containerId string, networkId string)
	CloseContainerRuntime()
	IsContainerRuntimeAvailable() bool
	CreateContainerRegistry()
	Commit(containerID string)
	Tag(source, target string)
}

func NewContainerRuntime(ctx *ekctx.EKContext) IContainerRuntime {

	cfg, err := NewEasykubeConfig(ctx).LoadConfig()
	if err != nil {
		panic(err)
	}
	if cfg.ContainerRuntime == "podman" {
		return NewPodmanImpl()
	} else {
		return NewDockerImpl()
	}
}
