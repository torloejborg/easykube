package ez

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
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
	"github.com/torloejborg/easykube/pkg/core"
)

// Custom error types for container runtime operations.
var (
	ErrContainerNotFound    = errors.New("container not found")
	ErrContainerNotRunning  = errors.New("container not running")
	ErrImageNotFound        = errors.New("image not found")
	ErrImagePushFailed      = errors.New("failed to push image")
	ErrImagePullFailed      = errors.New("failed to pull image")
	ErrContainerOperation   = errors.New("container operation failed")
	ErrDockerClientCreation = errors.New("failed to create Docker client")
	ErrUnsupportedRuntime   = errors.New("unsupported container runtime")
	ErrBinaryNotFound       = errors.New("required binary not found")
)

type ContainerRuntimeImpl struct {
	Docker      *client.Client
	ctx         context.Context
	Fs          afero.Fs
	RuntimeType string
	ek          *core.Ek
}

func NewContainerRuntimeImpl(ek *core.Ek, runtime string) (core.IContainerRuntime, error) {
	clientsOpts := make([]client.Opt, 0)
	clientsOpts = append(clientsOpts, client.WithAPIVersionNegotiation())

	switch runtime {
	case "docker":
		if !ek.Utils.HasBinary("docker") {
			return nil, errors.Join(ErrBinaryNotFound, errors.New("docker binary not found, is it installed"))
		}
		clientsOpts = append(clientsOpts, client.FromEnv)
	case "podman":
		if !ek.Utils.HasBinary("podman") {
			return nil, errors.Join(ErrBinaryNotFound, errors.New("podman binary not found, is it installed"))
		}

		// Get the socket location for podman.
		sout, stderr, err := ek.ExternalTools.RunCommand("podman", []string{"info", "--format", "{{.Host.RemoteSocket.Path}}"}...)
		if err != nil {
			return nil, errors.Join(ErrDockerClientCreation, fmt.Errorf("failed to determine podman runtime: %w (stderr: %s)", err, stderr))
		}

		// Some podman versions report unix://, others not.
		if !strings.Contains(string(sout), "unix") {
			sout = "unix://" + sout
		}

		socket := strings.TrimSpace(sout)
		clientsOpts = append(clientsOpts, client.WithHost(socket))
	default:
		return nil, errors.Join(ErrUnsupportedRuntime, fmt.Errorf("unknown container runtime: %s", runtime))
	}

	docker, err := client.NewClientWithOpts(clientsOpts...)
	if err != nil {
		return nil, errors.Join(ErrDockerClientCreation, fmt.Errorf("failed to create Docker client for %s: %w", runtime, err))
	}

	return &ContainerRuntimeImpl{
		Docker:      docker,
		ctx:         context.Background(),
		RuntimeType: runtime,
		ek:          ek,
	}, nil
}

func (cri *ContainerRuntimeImpl) IsClusterRunning() bool {
	running, err := cri.IsContainerRunning(constants.KindContainer)
	if err != nil {
		cri.ek.Printer.FmtRed("failed to check if cluster is running: %v", err)
		return false
	}
	return running
}

func (cri *ContainerRuntimeImpl) IsNetworkConnectedToContainer(containerID, networkID string) (bool, error) {
	jsonData, err := cri.Docker.ContainerInspect(cri.ctx, containerID)
	if err != nil {
		return false, fmt.Errorf("failed to inspect container %s: %w", containerID, err)
	}
	networkData := jsonData.NetworkSettings.Networks[networkID]
	return networkData != nil, nil
}

func (cri *ContainerRuntimeImpl) IsContainerRunning(containerID string) (bool, error) {
	result, err := cri.Docker.ContainerInspect(cri.ctx, containerID)
	if err != nil {
		if client.IsErrNotFound(err) {
			return false, errors.Join(ErrContainerNotFound, fmt.Errorf("container %s not found: %w", containerID, err))
		}
		return false, fmt.Errorf("failed to inspect container %s: %w", containerID, err)
	}
	return result.State.Running, nil
}

func (cri *ContainerRuntimeImpl) HasImageInKindRegistry(image string) (bool, error) {
	image = strings.ReplaceAll(image, constants.LocalRegistry+"/", "")
	parts := strings.Split(image, ":")
	if len(parts) < 2 {
		return false, fmt.Errorf("invalid image format: %s", image)
	}
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

	resp, err := httpClient.Get(fmt.Sprintf("https://%s/v2/%s/tags/list", constants.LocalRegistry, imageWithoutTag))
	if err != nil {
		return false, fmt.Errorf("failed to fetch tags for image %s: %w", image, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode == 404 {
		return false, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %w", err)
	}

	var dat TagList
	if err := json.Unmarshal(body, &dat); err != nil {
		return false, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if strings.Contains(dat.Name, imageWithoutTag) {
		for _, tag := range dat.Tags {
			if tag == imageTag {
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
		return false, fmt.Errorf("failed to list images: %w", err)
	}

	for _, it := range res {
		for _, tag := range it.RepoTags {
			if tag == image {
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
		return errors.Join(ErrImagePushFailed, fmt.Errorf("failed to push image %s to %s: %w", src, dest, err))
	}
	defer reader.Close()

	if _, err := io.ReadAll(reader); err != nil {
		return errors.Join(ErrImagePushFailed, fmt.Errorf("failed to read push output: %w", err))
	}
	return nil
}

func (cri *ContainerRuntimeImpl) PullImage(image string, credentials *core.PrivateRegistryCredentials) error {
	opts := image2.PullOptions{
		All:      false,
		Platform: "linux/amd64",
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
				return fmt.Errorf("failed to encode auth config: %w", err)
			}
			opts.RegistryAuth = encoded
		}

		if cri.RuntimeType == "docker" {
			jsonBytes, err := json.Marshal(map[string]string{
				"username": credentials.Username,
				"password": credentials.Password,
			})
			if err != nil {
				return fmt.Errorf("failed to marshal credentials: %w", err)
			}
			opts.RegistryAuth = base64.StdEncoding.EncodeToString(jsonBytes)
		}
	}

	reader, err := cri.Docker.ImagePull(cri.ctx, image, opts)
	if err != nil {
		return errors.Join(ErrImagePullFailed, fmt.Errorf("failed to pull image %s: %w", image, err))
	}
	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("failed to close pull reader: %v", err)
		}
	}()

	// Wait for the pull to complete by reading the output stream.
	decoder := json.NewDecoder(reader)
	for {
		var msg map[string]interface{}
		if err := decoder.Decode(&msg); err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("error decoding pull response: %w", err)
		}
	}
	return nil
}

func (cri *ContainerRuntimeImpl) FindContainer(name string) (*core.ContainerSearch, error) {
	f := filters.NewArgs()
	f.Add("name", name)
	opts := container.ListOptions{
		All:     true,
		Filters: f,
	}

	resp, err := cri.Docker.ContainerList(cri.ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	if len(resp) == 1 {
		return &core.ContainerSearch{
			Found:       true,
			IsRunning:   resp[0].State == "running",
			ContainerID: resp[0].ID,
		}, nil
	}

	return &core.ContainerSearch{
		Found:       false,
		IsRunning:   false,
		ContainerID: "",
	}, nil
}

func (cri *ContainerRuntimeImpl) StartContainer(id string) error {
	if err := cri.Docker.ContainerStart(cri.ctx, id, container.StartOptions{}); err != nil {
		return errors.Join(ErrContainerOperation, fmt.Errorf("failed to start container %s: %w", id, err))
	}
	return nil
}

func (cri *ContainerRuntimeImpl) StopContainer(id string) error {
	if err := cri.Docker.ContainerStop(cri.ctx, id, container.StopOptions{}); err != nil {
		return errors.Join(ErrContainerOperation, fmt.Errorf("failed to stop container %s: %w", id, err))
	}
	return nil
}

func (cri *ContainerRuntimeImpl) RemoveContainer(id string) error {
	if err := cri.Docker.ContainerRemove(cri.ctx, id, container.RemoveOptions{}); err != nil {
		return errors.Join(ErrContainerOperation, fmt.Errorf("failed to remove container %s: %w", id, err))
	}
	return nil
}

func (cri *ContainerRuntimeImpl) Exec(containerId string, cmd []string) error {
	exec := container.ExecOptions{
		Cmd:          cmd,
		AttachStderr: true,
		AttachStdout: true,
	}

	x, err := cri.Docker.ContainerExecCreate(cri.ctx, containerId, exec)
	if err != nil {
		return errors.Join(ErrContainerOperation, fmt.Errorf("failed to create exec in container %s: %w", containerId, err))
	}

	if err := cri.Docker.ContainerExecStart(cri.ctx, x.ID, container.ExecStartOptions{
		Detach: false,
	}); err != nil {
		return errors.Join(ErrContainerOperation, fmt.Errorf("failed to start exec in container %s: %w", containerId, err))
	}

	// Wait for exec to complete by polling container inspect.
	for i := 1; i < 20; i++ {
		resp, err := cri.Docker.ContainerInspect(cri.ctx, containerId)
		if err != nil {
			return fmt.Errorf("failed to inspect container %s: %w", containerId, err)
		}
		time.Sleep(500 * time.Millisecond)
		if len(resp.ExecIDs) == 0 {
			break
		}
	}
	return nil
}

func (cri *ContainerRuntimeImpl) ContainerWriteFile(containerId, dst, filename string, data []byte) error {
	opts := container.CopyToContainerOptions{
		AllowOverwriteDirWithFile: true,
	}

	dataTar, err := memtar(data, filename)
	if err != nil {
		return fmt.Errorf("failed to create tar for file %s: %w", filename, err)
	}

	if err := cri.Docker.CopyToContainer(cri.ctx, containerId, dst, bytes.NewReader(dataTar), opts); err != nil {
		return errors.Join(ErrContainerOperation, fmt.Errorf("failed to write file %s in container %s: %w", filename, containerId, err))
	}
	return nil
}

func (cri *ContainerRuntimeImpl) NetworkConnect(containerId, networkId string) error {
	if err := cri.Docker.NetworkConnect(cri.ctx, networkId, containerId, nil); err != nil {
		return errors.Join(ErrContainerOperation, fmt.Errorf("failed to connect container %s to network %s: %w", containerId, networkId, err))
	}
	return nil
}

func (cri *ContainerRuntimeImpl) CloseContainerRuntime() {
	// No error handling needed for cleanup.
}

func (cri *ContainerRuntimeImpl) IsContainerRuntimeAvailable() bool {
	_, err := cri.Docker.Info(cri.ctx)
	return err == nil
}

func (cri *ContainerRuntimeImpl) StartContainerRegistry() error {
	containerSearch, err := cri.FindContainer(constants.RegistryContainer)
	if err != nil {
		return fmt.Errorf("failed to find registry container: %w", err)
	}

	if containerSearch.Found && !containerSearch.IsRunning {
		if err := cri.StartContainer(containerSearch.ContainerID); err != nil {
			return fmt.Errorf("failed to start registry container: %w", err)
		}
	}
	return nil
}

func (cri *ContainerRuntimeImpl) CreateContainerRegistry() error {
	registryImg := constants.RegistryImage
	containerName := constants.RegistryContainer

	// Ensure the registry-config file exists.
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get user config dir: %w", err)
	}

	if err := cri.ek.Utils.CopyResourceToConfigDir(constants.ZotConfig, constants.ZotConfig); err != nil {
		return fmt.Errorf("failed to copy zot config: %w", err)
	}

	if err := cri.ek.Utils.CopyResourceToConfigDir("cert/server.crt", "localtest.me.crt"); err != nil {
		return fmt.Errorf("failed to copy server certificate: %w", err)
	}

	if err := cri.ek.Utils.CopyResourceToConfigDir("cert/server.key", "localtest.me.key"); err != nil {
		return fmt.Errorf("failed to copy server key: %w", err)
	}

	containerSearch, err := cri.FindContainer(containerName)
	if err != nil {
		return fmt.Errorf("failed to find container %s: %w", containerName, err)
	}

	if !containerSearch.Found {
		containerConfig := &container.Config{
			ExposedPorts: nat.PortSet{nat.Port("5000"): struct{}{}},
			Image:        registryImg,
		}

		configDir, err = cri.ek.OsDetails.GetEasykubeConfigDir()
		if err != nil {
			return fmt.Errorf("failed to get easykube config dir: %w", err)
		}

		binds := []string{
			filepath.Join(configDir, "localtest.me.crt") + ":/etc/ssl/localtest.me.crt:z",
			filepath.Join(configDir, "localtest.me.key") + ":/etc/ssl/localtest.me.key:z",
			filepath.Join(configDir, "persistence", "zot") + ":/var/lib/zot:z",
			filepath.Join(configDir, constants.ZotCredentials) + ":/var/lib/zot/credentials.json:z",
			filepath.Join(configDir, constants.ZotConfig) + ":/etc/zot/config.json:z",
		}

		networkingConfig := &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				"kind": {
					Aliases: []string{"registry.localtest.me"},
				},
			},
		}

		hostConfig := &container.HostConfig{
			LogConfig:    container.LogConfig{},
			NetworkMode:  "kind",
			PortBindings: map[nat.Port][]nat.PortBinding{nat.Port(constants.LocalRegistryPort): {{HostIP: "127.0.0.1", HostPort: constants.LocalRegistryPort}}},
			RestartPolicy: container.RestartPolicy{
				Name:              "always",
				MaximumRetryCount: 0,
			},
			Binds: binds,
		}

		_, err = cri.Docker.ContainerCreate(cri.ctx, containerConfig, hostConfig, networkingConfig, nil, constants.RegistryContainer)
		if err != nil {
			return fmt.Errorf("failed to create registry container: %w", err)
		}
	}
	return nil
}

func (cri *ContainerRuntimeImpl) Commit(containerID string) (string, error) {
	opts := container.CommitOptions{
		Reference: "",
	}

	resp, err := cri.Docker.ContainerCommit(cri.ctx, containerID, opts)
	if err != nil {
		return "", fmt.Errorf("failed to commit container %s: %w", containerID, err)
	}
	return resp.ID, nil
}

func (cri *ContainerRuntimeImpl) TagImage(source, target string) error {
	if err := cri.Docker.ImageTag(cri.ctx, source, target); err != nil {
		return fmt.Errorf("failed to tag image %s as %s: %w", source, target, err)
	}
	return nil
}

// memtar creates a tar archive in memory for a single file.
func memtar(data []byte, filename string) ([]byte, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	defer tw.Close()

	hdr := &tar.Header{
		Name:     filename,
		Size:     int64(len(data)),
		Mode:     509,
		ModTime:  time.Now(),
		Typeflag: tar.TypeReg,
	}

	if err := tw.WriteHeader(hdr); err != nil {
		return nil, fmt.Errorf("failed to write tar header: %w", err)
	}

	num, err := tw.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed to write tar data: %w", err)
	}

	if num == 0 || num != len(data) {
		return nil, errors.New("tar wrote zero or wrong number of bytes")
	}

	return buf.Bytes(), nil
}
