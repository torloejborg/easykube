package ek

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	image2 "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/torloj/easykube/pkg/constants"
)

type DockerImpl struct {
	Docker *client.Client
	ctx    context.Context
	Common ContainerRuntimeCommon
}

func NewDockerImpl() IContainerRuntime {

	docker, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Println("No Docker context found. Is docker running??")
		os.Exit(-1)
	}

	return &DockerImpl{
		Docker: docker,
		ctx:    context.Background(),
		Common: ContainerRuntimeCommon{},
	}

}
func (cr *DockerImpl) IsClusterRunning() bool {
	return cr.IsContainerRunning(constants.KIND_CONTAINER)
}

func (cr *DockerImpl) IsNetworkConnectedToContainer(containerID string, networkID string) bool {
	jsonData, err := cr.Docker.ContainerInspect(cr.ctx, containerID)
	if err != nil {
		return false
	}
	networkData := jsonData.NetworkSettings.Networks[networkID]
	return networkData != nil
}

func (cr *DockerImpl) IsContainerRunning(containerID string) bool {
	result, err := cr.Docker.ContainerInspect(cr.ctx, containerID)

	if err != nil {
		return false
	}

	return result.State.Running
}

func (cr *DockerImpl) HasImage(image string) bool {

	f := filters.NewArgs()
	f.Add("reference", image)

	opts := image2.ListOptions{
		All:     true,
		Filters: f,
	}

	res, err := cr.Docker.ImageList(cr.ctx, opts)
	if err != nil {
		panic(err)
	}
	for _, it := range res {
		tags := it.RepoTags
		for idx := range tags {
			if tags[idx] == image {
				return true
			}
		}
	}

	return false
}

func (cr *DockerImpl) Push(image string) {

	opts := image2.PushOptions{
		All:           false,
		RegistryAuth:  "anything",
		PrivilegeFunc: nil,
		Platform:      nil,
	}

	reader, err := cr.Docker.ImagePush(cr.ctx, image, opts)
	if err != nil {
		panic(err)
	}
	defer reader.Close()
	_, err = io.ReadAll(reader)

}

func (cr *DockerImpl) Pull(image string, privateRegistryCredentials *string) {

	opts := image2.PullOptions{
		All: false,
	}

	if privateRegistryCredentials != nil {
		opts.RegistryAuth = *privateRegistryCredentials
	}

	reader, err := cr.Docker.ImagePull(cr.ctx, image, opts)
	if err != nil {
		panic(err)
	}
	defer reader.Close()

	// Wait for the pull to complete by reading the output stream
	decoder := json.NewDecoder(reader)
	for {
		var msg map[string]interface{}
		if err := decoder.Decode(&msg); err == io.EOF {
			break // Done pulling
		} else if err != nil {
			log.Fatalf("error decoding pull response: %v", err)
		}
	}

}

func (cr *DockerImpl) FindContainer(name string) ContainerSearch {

	f := filters.NewArgs()
	f.Add("name", name)
	opts := container.ListOptions{
		All:     true,
		Filters: f,
	}

	resp, err := cr.Docker.ContainerList(cr.ctx, opts)
	if err != nil {
		panic(err)
	}

	if len(resp) == 1 {

		state := resp[0].State == "running"

		return ContainerSearch{
			Found:       true,
			IsRunning:   state,
			ContainerID: resp[0].ID,
		}
	} else {
		return ContainerSearch{
			Found:       false,
			IsRunning:   false,
			ContainerID: "",
		}
	}
}

func (cr *DockerImpl) HasImageInKindRegistry(name string) bool {
	return cr.Common.ImageExistsInKindRegistry(name)
}

func (cr *DockerImpl) StartContainer(id string) {
	err := cr.Docker.ContainerStart(cr.ctx, id, container.StartOptions{})
	if err != nil {
		panic(err)
	}
}

func (cr *DockerImpl) StopContainer(id string) {
	err := cr.Docker.ContainerStop(cr.ctx, id, container.StopOptions{})
	if err != nil {
		panic(err)
	}
}

func (cr *DockerImpl) RemoveContainer(id string) {
	err := cr.Docker.ContainerRemove(cr.ctx, id, container.RemoveOptions{})
	if err != nil {
		panic(err)
	}
}

func (cr *DockerImpl) Exec(containerId string, cmd []string) {
	exec := container.ExecOptions{
		Cmd:          cmd,
		AttachStderr: true,
		AttachStdout: true,
	}

	x, err := cr.Docker.ContainerExecCreate(cr.ctx, containerId, exec)
	if err != nil {
		panic(err)
	}

	err = cr.Docker.ContainerExecStart(cr.ctx, x.ID, container.ExecStartOptions{
		Detach: true,
	})

	if err != nil {
		panic(err)
	}

	for i := 1; i < 20; i++ {
		resp, err := cr.Docker.ContainerInspect(cr.ctx, containerId)
		if err != nil {
			panic(err)
		}
		time.Sleep(500 * time.Millisecond)
		if len(resp.ExecIDs) == 0 {
			break
		}
	}
}

func (cr *DockerImpl) WriteFile(containerId string, dst string, filename string, data []byte) {
	opts := container.CopyToContainerOptions{
		AllowOverwriteDirWithFile: true,
	}

	dat, err := memtar(data, filename)
	if err != nil {
		panic(err)
	}

	err = cr.Docker.CopyToContainer(cr.ctx, containerId, dst, bytes.NewReader(dat), opts)
	if err != nil {
		panic(err)
	}
}

func (cr *DockerImpl) NetworkConnect(containerId string, networkId string) {
	cr.Docker.NetworkConnect(cr.ctx, networkId, containerId, nil)
}

func (cr *DockerImpl) CloseContainerRuntime() {
}

func (cr *DockerImpl) IsContainerRuntimeAvailable() bool {
	_, err := cr.Docker.Info(cr.ctx)
	return err == nil
}

func (cr *DockerImpl) CreateContainerRegistry() {

	registry := constants.REGISTRY_IMAGE
	containerName := constants.REGISTRY_CONTAINER

	// make sure that the registry-config file exists
	configDir, _ := os.UserConfigDir()
	CopyResource("registry-config.yaml", "registry-config.yaml")
	CopyResource("cert/server.crt", "localtest.me.crt")
	CopyResource("cert/server.key", "localtest.me.key")

	imageSearch := cr.HasImage(registry)
	if !imageSearch {
		cr.Pull(registry, nil)
	}

	containerSearch := cr.FindContainer(containerName)
	if !containerSearch.Found {

		//registryPort := fmt.Sprintf("%d", constants.LOCAL_REGISTRY_PORT)

		containerConfig := &container.Config{
			ExposedPorts: nat.PortSet{nat.Port("5000"): struct{}{}},
			Image:        registry,
		}

		binds := make([]string, 3)
		binds[0] = filepath.Join(configDir, "easykube", "registry-config.yaml") + ":/etc/docker/registry/config.yml"
		binds[1] = filepath.Join(configDir, "easykube", "localtest.me.crt") + ":/etc/ssl/localtest.me.crt"
		binds[2] = filepath.Join(configDir, "easykube", "localtest.me.key") + ":/etc/ssl/localtest.me.key"

		hostConfig := &container.HostConfig{
			LogConfig:    container.LogConfig{},
			NetworkMode:  "bridge",
			PortBindings: map[nat.Port][]nat.PortBinding{nat.Port("5000"): {{HostIP: "127.0.0.1", HostPort: "5001"}}},
			RestartPolicy: container.RestartPolicy{
				Name:              "always",
				MaximumRetryCount: 0,
			},
			Binds: binds,
		}

		resp, err := cr.Docker.ContainerCreate(cr.ctx, containerConfig, hostConfig, nil, nil, constants.REGISTRY_CONTAINER)
		if err != nil {
			panic(err)
		}

		err = cr.Docker.ContainerStart(cr.ctx, resp.ID, container.StartOptions{})
		if err != nil {
			panic(err)
		}
	}

	if containerSearch.Found && !containerSearch.IsRunning {
		cr.StartContainer(containerSearch.ContainerID)
	}
}

func (cr *DockerImpl) Commit(containerID string) {

	opts := container.CommitOptions{
		Reference: "",
	}

	resp, err := cr.Docker.ContainerCommit(cr.ctx, containerID, opts)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.ID)
}

func (cr *DockerImpl) Tag(source string, target string) {

	err := cr.Docker.ImageTag(cr.ctx, source, target)
	if err != nil {
		panic(err)
	}
}

func memtar(data []byte, filename string) ([]byte, error) {
	var buf bytes.Buffer

	// creating tar writer from new buffer
	tw := tar.NewWriter(&buf)
	defer tw.Close()

	// manually create tar header
	hdr := &tar.Header{
		Name:     filename,
		Size:     int64(len(data)),
		Mode:     509,
		ModTime:  time.Now(),
		Typeflag: tar.TypeReg, // regular file
	}

	err := tw.WriteHeader(hdr)
	if err != nil {
		return nil, err
	}

	num, err := tw.Write(data)
	if err != nil {
		return nil, err
	}

	// check if all data written
	if num == 0 || num != len(data) {
		return nil, errors.New("tar wrote zero or wrong num of bytes")
	}

	return buf.Bytes(), nil
}
