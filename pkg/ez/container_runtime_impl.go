package ez

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/tls"
	base64 "encoding/base64"
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
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/spf13/afero"
	"github.com/torloejborg/easykube/pkg/constants"
)

type ContainerRuntimeImpl struct {
	Docker      *client.Client
	ctx         context.Context
	Fs          afero.Fs
	RuntimeType string
}

func NewContainerRuntimeImpl(runtime string) (IContainerRuntime, error) {

	clientsOpts := make([]client.Opt, 0)
	clientsOpts = append(clientsOpts, client.WithAPIVersionNegotiation())

	switch runtime {
	case "docker":
		if !HasBinary("docker") {
			return nil, errors.New("docker binary not found, is it installed")
		}
		clientsOpts = append(clientsOpts, client.FromEnv)
		break
	case "podman":

		if !HasBinary("podman") {
			return nil, errors.New("podman binary not found, is it installed")
		}

		// get the socket location
		sout, _, err := Kube.RunCommand("podman", []string{"info", "--format", "{{.Host.RemoteSocket.Path}}"}...)
		if err != nil {
			return nil, errors.Join(errors.New("Failed to determine podman runtime"), err)
		}
		socket := strings.TrimSpace(sout)
		clientsOpts = append(clientsOpts, client.WithHost(socket))
		break
	default:
		panic("unknown container runtime")
	}

	docker, err := client.NewClientWithOpts(clientsOpts...)
	if err != nil {
		fmt.Println("No container context/runtime found. Is docker running??")
		os.Exit(-1)
	}

	return &ContainerRuntimeImpl{
		Docker:      docker,
		ctx:         context.Background(),
		RuntimeType: runtime,
	}, nil

}
func (cri *ContainerRuntimeImpl) IsClusterRunning() bool {

	running, err := cri.IsContainerRunning(constants.KIND_CONTAINER)
	if err != nil {
		return false
	} else {
		return running
	}
}

func (cri *ContainerRuntimeImpl) DiscoverContainerRuntimeConnection(runtimeType string) (ContainerConnection, error) {

	res := ContainerConnection{
		Host: "nil",
		Type: "docker",
	}

	return res, nil
}

func (cri *ContainerRuntimeImpl) IsNetworkConnectedToContainer(containerID string, networkID string) (bool, error) {
	jsonData, err := cri.Docker.ContainerInspect(cri.ctx, containerID)
	if err != nil {
		return false, err
	}
	networkData := jsonData.NetworkSettings.Networks[networkID]
	return networkData != nil, nil
}

func (cri *ContainerRuntimeImpl) IsContainerRunning(containerID string) (bool, error) {
	result, err := cri.Docker.ContainerInspect(cri.ctx, containerID)

	if err != nil {
		return false, err
	}

	return result.State.Running, nil
}

func (cri *ContainerRuntimeImpl) HasImageInKindRegistry(image string) (bool, error) {
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

	resp, err := httpClient.Get(fmt.Sprintf("https://%s/v2/%s/tags/list", constants.LOCAL_REGISTRY, imageWithoutTag))
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

func (cri *ContainerRuntimeImpl) HasImage(image string) (bool, error) {

	f := filters.NewArgs()
	f.Add("reference", image)

	opts := image2.ListOptions{
		All:     true,
		Filters: f,
	}

	res, err := cri.Docker.ImageList(cri.ctx, opts)
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

func (cri *ContainerRuntimeImpl) PushImage(src, dest string) error {

	auth := base64.StdEncoding.EncodeToString([]byte(`{}`))

	opts := image2.PushOptions{
		All:           false,
		RegistryAuth:  auth,
		PrivilegeFunc: nil,
		Platform:      nil,
	}

	reader, err := cri.Docker.ImagePush(cri.ctx, dest, opts)
	if err != nil {
		panic(err)
	}
	defer reader.Close()
	if _, err = io.ReadAll(reader); err != nil {
		return err
	}

	return nil

}

func (cri *ContainerRuntimeImpl) PullImage(image string, credentials *PrivateRegistryCredentials) error {

	opts := image2.PullOptions{
		All: false,
	}

	if credentials != nil {

		if cri.RuntimeType == "podman" {

			extractRegistry := func(image string) string {
				parts := strings.Split(image, "/")
				if len(parts) >= 2 && strings.Contains(parts[0], ".") {
					return parts[0]
				}
				return ""
			}

			auth := base64.StdEncoding.EncodeToString([]byte(credentials.Username + ":" + credentials.Password))
			authConfig := registry.AuthConfig{
				Auth:          auth,
				ServerAddress: extractRegistry(image),
			}

			encoded, err := registry.EncodeAuthConfig(authConfig)
			if err != nil {
				return err
			}

			opts.RegistryAuth = encoded

		}

		if cri.RuntimeType == "docker" {
			jsonBytes, _ := json.Marshal(map[string]string{
				"username": credentials.Username,
				"password": credentials.Password,
			})

			opts.RegistryAuth = base64.StdEncoding.EncodeToString(jsonBytes)
		}

	}

	reader, err := cri.Docker.ImagePull(cri.ctx, image, opts)
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

func (cri *ContainerRuntimeImpl) FindContainer(name string) (*ContainerSearch, error) {

	f := filters.NewArgs()
	f.Add("name", name)
	opts := container.ListOptions{
		All:     true,
		Filters: f,
	}

	resp, err := cri.Docker.ContainerList(cri.ctx, opts)
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

func (cri *ContainerRuntimeImpl) StartContainer(id string) error {
	if err := cri.Docker.ContainerStart(cri.ctx, id, container.StartOptions{}); err != nil {
		return err
	} else {
		return nil
	}
}

func (cri *ContainerRuntimeImpl) StopContainer(id string) error {
	if err := cri.Docker.ContainerStop(cri.ctx, id, container.StopOptions{}); err != nil {
		return err
	} else {
		return nil
	}
}

func (cri *ContainerRuntimeImpl) RemoveContainer(id string) error {
	if err := cri.Docker.ContainerRemove(cri.ctx, id, container.RemoveOptions{}); err != nil {
		return err
	} else {
		return nil
	}
}

func (cri *ContainerRuntimeImpl) Exec(containerId string, cmd []string) error {
	exec := container.ExecOptions{
		Cmd:          cmd,
		AttachStderr: true,
		AttachStdout: true,
	}

	x, err := cri.Docker.ContainerExecCreate(cri.ctx, containerId, exec)
	if err != nil {
		return err
	}

	if err := cri.Docker.ContainerExecStart(cri.ctx, x.ID, container.ExecStartOptions{
		Detach: false,
	}); err != nil {
		return err
	}

	for i := 1; i < 20; i++ {
		resp, err := cri.Docker.ContainerInspect(cri.ctx, containerId)
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

func (cri *ContainerRuntimeImpl) ContainerWriteFile(containerId string, dst string, filename string, data []byte) error {
	opts := container.CopyToContainerOptions{
		AllowOverwriteDirWithFile: true,
	}

	data, err := memtar(data, filename)
	if err != nil {
		return err
	}

	if err := cri.Docker.CopyToContainer(cri.ctx, containerId, dst, bytes.NewReader(data), opts); err != nil {
		return errors.Join(errors.New("failed to write file in docker container"), err)
	}

	return nil
}

func (cri *ContainerRuntimeImpl) NetworkConnect(containerId string, networkId string) error {
	if err := cri.Docker.NetworkConnect(cri.ctx, networkId, containerId, nil); err != nil {
		return err
	} else {
		return nil
	}
}

func (cri *ContainerRuntimeImpl) CloseContainerRuntime() {
}

func (cri *ContainerRuntimeImpl) IsContainerRuntimeAvailable() bool {
	_, err := cri.Docker.Info(cri.ctx)
	return err == nil
}

func (cri *ContainerRuntimeImpl) CreateContainerRegistry() error {

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

	imageSearch, err := cri.HasImage(registry)
	if !imageSearch {
		if err := cri.PullImage(registry, nil); err != nil {
			return err
		}
	}

	containerSearch, err := cri.FindContainer(containerName)
	if err != nil {
		return err
	}

	if !containerSearch.Found {

		containerConfig := &container.Config{
			ExposedPorts: nat.PortSet{nat.Port("5000"): struct{}{}},
			Image:        registry,
		}

		binds := make([]string, 3)
		binds[0] = filepath.Join(configDir, "easykube", "registry-config.yaml") + ":/etc/docker/registry/config.yml:z"
		binds[1] = filepath.Join(configDir, "easykube", "localtest.me.crt") + ":/etc/ssl/localtest.me.crt:z"
		binds[2] = filepath.Join(configDir, "easykube", "localtest.me.key") + ":/etc/ssl/localtest.me.key:z"

		networkingConfig := &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				"kind": {
					Aliases: []string{
						"registry.localtest.me",
					},
				},
			},
		}

		hostConfig := &container.HostConfig{
			LogConfig:    container.LogConfig{},
			NetworkMode:  "kind",
			PortBindings: map[nat.Port][]nat.PortBinding{nat.Port("5000"): {{HostIP: "127.0.0.1", HostPort: "5001"}}},
			RestartPolicy: container.RestartPolicy{
				Name:              "always",
				MaximumRetryCount: 0,
			},
			Binds: binds,
		}

		resp, err := cri.Docker.ContainerCreate(cri.ctx, containerConfig, hostConfig, networkingConfig, nil, constants.REGISTRY_CONTAINER)
		if err != nil {
			panic(err)
		}

		err = cri.Docker.ContainerStart(cri.ctx, resp.ID, container.StartOptions{})
		if err != nil {
			panic(err)
		}
	}

	if containerSearch.Found && !containerSearch.IsRunning {
		cri.StartContainer(containerSearch.ContainerID)
	}

	return nil
}

func (cri *ContainerRuntimeImpl) Commit(containerID string) {

	opts := container.CommitOptions{
		Reference: "",
	}

	resp, err := cri.Docker.ContainerCommit(cri.ctx, containerID, opts)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.ID)
}

func (cri *ContainerRuntimeImpl) TagImage(source string, target string) error {
	if err := cri.Docker.ImageTag(cri.ctx, source, target); err != nil {
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
