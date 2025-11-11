package ez

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	image2 "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/spf13/afero"
	"github.com/torloejborg/easykube/pkg/constants"
)

type DockerImpl struct {
	Docker *client.Client
	ctx    context.Context
	Fs     afero.Fs
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
	}

}
func (cr *DockerImpl) IsClusterRunning() bool {

	running, err := cr.IsContainerRunning(constants.KIND_CONTAINER)
	if err != nil {
		return false
	} else {
		return running
	}
}

func (cr *DockerImpl) IsNetworkConnectedToContainer(containerID string, networkID string) (bool, error) {
	jsonData, err := cr.Docker.ContainerInspect(cr.ctx, containerID)
	if err != nil {
		return false, err
	}
	networkData := jsonData.NetworkSettings.Networks[networkID]
	return networkData != nil, nil
}

func (cr *DockerImpl) IsContainerRunning(containerID string) (bool, error) {
	result, err := cr.Docker.ContainerInspect(cr.ctx, containerID)

	if err != nil {
		return false, err
	}

	return result.State.Running, nil
}

func (i *DockerImpl) HasImageInKindRegistry(image string) (bool, error) {
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
	httpClient := &http.Client{Transport: tr}

	resp, err := httpClient.Get(fmt.Sprintf("http://%s/v2/%s/tags/list", constants.LOCAL_REGISTRY, imageWithoutTag))
	if nil != err {
		return false, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if nil != err {
		return false, err
	}

	var dat TagList

	if jsonerr := json.Unmarshal(body, &dat); jsonerr != nil {
		panic(jsonerr)
	}

	if strings.Contains(dat.Name, imageWithoutTag) {
		for i := range dat.Tags {
			if dat.Tags[i] == imageTag {
				return true, nil
			}
		}
	}

	return false, nil
}

func (cr *DockerImpl) HasImage(image string) (bool, error) {

	f := filters.NewArgs()
	f.Add("reference", image)

	opts := image2.ListOptions{
		All:     true,
		Filters: f,
	}

	res, err := cr.Docker.ImageList(cr.ctx, opts)
	if err != nil {
		return false, err
	}
	for _, it := range res {
		tags := it.RepoTags
		for idx := range tags {
			if tags[idx] == image {
				return true, nil
			}
		}
	}

	return false, nil
}

func (cr *DockerImpl) PushImage(src, dest string) error {

	opts := image2.PushOptions{
		All:           false,
		RegistryAuth:  "anything",
		PrivilegeFunc: nil,
		Platform:      nil,
	}

	reader, err := cr.Docker.ImagePush(cr.ctx, dest, opts)
	if err != nil {
		panic(err)
	}
	defer reader.Close()
	if _, err = io.ReadAll(reader); err != nil {
		return err
	}

	return nil

}

func (cr *DockerImpl) PullImage(image string, privateRegistryCredentials *string) error {

	opts := image2.PullOptions{
		All: false,
	}

	if privateRegistryCredentials != nil {
		opts.RegistryAuth = *privateRegistryCredentials
	}

	reader, err := cr.Docker.ImagePull(cr.ctx, image, opts)
	if err != nil {
		return err
	}
	defer func(reader io.ReadCloser) {
		_ = reader.Close()
	}(reader)

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

	return nil
}

func (cr *DockerImpl) FindContainer(name string) (*ContainerSearch, error) {

	f := filters.NewArgs()
	f.Add("name", name)
	opts := container.ListOptions{
		All:     true,
		Filters: f,
	}

	resp, err := cr.Docker.ContainerList(cr.ctx, opts)
	if err != nil {
		return nil, err
	}

	if len(resp) == 1 {

		state := resp[0].State == "running"

		return &ContainerSearch{
			Found:       true,
			IsRunning:   state,
			ContainerID: resp[0].ID,
		}, nil
	} else {
		return &ContainerSearch{
			Found:       false,
			IsRunning:   false,
			ContainerID: "",
		}, nil
	}
}

func (cr *DockerImpl) StartContainer(id string) error {
	if err := cr.Docker.ContainerStart(cr.ctx, id, container.StartOptions{}); err != nil {
		return err
	} else {
		return nil
	}
}

func (cr *DockerImpl) StopContainer(id string) error {
	if err := cr.Docker.ContainerStop(cr.ctx, id, container.StopOptions{}); err != nil {
		return err
	} else {
		return nil
	}
}

func (cr *DockerImpl) RemoveContainer(id string) error {
	if err := cr.Docker.ContainerRemove(cr.ctx, id, container.RemoveOptions{}); err != nil {
		return err
	} else {
		return nil
	}
}

func (cr *DockerImpl) Exec(containerId string, cmd []string) error {
	exec := container.ExecOptions{
		Cmd:          cmd,
		AttachStderr: true,
		AttachStdout: true,
	}

	x, err := cr.Docker.ContainerExecCreate(cr.ctx, containerId, exec)
	if err != nil {
		return err
	}

	if err := cr.Docker.ContainerExecStart(cr.ctx, x.ID, container.ExecStartOptions{
		Detach: false,
	}); err != nil {
		return err
	}

	for i := 1; i < 20; i++ {
		resp, err := cr.Docker.ContainerInspect(cr.ctx, containerId)
		if err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)
		if len(resp.ExecIDs) == 0 {
			break
		}
	}

	return nil
}

func (cr *DockerImpl) ContainerWriteFile(containerId string, dst string, filename string, data []byte) error {
	opts := container.CopyToContainerOptions{
		AllowOverwriteDirWithFile: true,
	}

	data, err := memtar(data, filename)
	if err != nil {
		return err
	}

	if err := cr.Docker.CopyToContainer(cr.ctx, containerId, dst, bytes.NewReader(data), opts); err != nil {
		return errors.Join(errors.New("failed to write file in docker container"), err)
	}

	return nil
}

func (cr *DockerImpl) NetworkConnect(containerId string, networkId string) error {
	if err := cr.Docker.NetworkConnect(cr.ctx, networkId, containerId, nil); err != nil {
		return err
	} else {
		return nil
	}
}

func (cr *DockerImpl) CloseContainerRuntime() {
}

func (cr *DockerImpl) IsContainerRuntimeAvailable() bool {
	_, err := cr.Docker.Info(cr.ctx)
	return err == nil
}

func (cr *DockerImpl) CreateContainerRegistry() error {

	registry := constants.REGISTRY_IMAGE
	containerName := constants.REGISTRY_CONTAINER

	// make sure that the registry-config file exists
	configDir, _ := os.UserConfigDir()
	if err := CopyResource("registry-config.yaml", "registry-config.yaml"); err != nil {
		return err
	}

	if err := CopyResource("cert/server.crt", "localtest.me.crt"); err != nil {
		return err
	}

	if err := CopyResource("cert/server.key", "localtest.me.key"); err != nil {
		return err
	}

	imageSearch, err := cr.HasImage(registry)
	if !imageSearch {
		if err := cr.PullImage(registry, nil); err != nil {
			return err
		}
	}

	containerSearch, err := cr.FindContainer(containerName)
	if err != nil {
		return err
	}

	if !containerSearch.Found {

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

	return nil
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

func (cr *DockerImpl) TagImage(source string, target string) error {
	if err := cr.Docker.ImageTag(cr.ctx, source, target); err != nil {
		return err
	} else {
		return nil
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
