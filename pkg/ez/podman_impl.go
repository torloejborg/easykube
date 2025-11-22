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

	"github.com/containers/podman/v6/pkg/api/handlers"
	"github.com/containers/podman/v6/pkg/bindings"
	"github.com/containers/podman/v6/pkg/bindings/containers"
	"github.com/containers/podman/v6/pkg/bindings/images"
	"github.com/containers/podman/v6/pkg/bindings/network"
	"github.com/containers/podman/v6/pkg/specgen"
	"github.com/spf13/afero"
	"github.com/torloejborg/easykube/pkg/constants"
	"go.podman.io/common/libnetwork/types"
	"k8s.io/utils/ptr"
)

type PodmanImpl struct {
	conn context.Context
	Fs   afero.Fs
}

func NewPodmanImpl() IContainerRuntime {

	conn, err := bindings.NewConnection(context.Background(), "unix:///run/user/1000/podman/podman.sock")
	if err != nil {
		fmt.Println("No Podman context found. Is podman running??")
		os.Exit(-1)
	}

	return &PodmanImpl{
		conn: conn,
		Fs:   Kube.Fs,
	}

}
func (cr *PodmanImpl) IsClusterRunning() bool {
	running, _ := cr.IsContainerRunning(constants.KIND_CONTAINER)
	return running
}

func (cr *PodmanImpl) IsNetworkConnectedToContainer(containerID string, networkID string) (bool, error) {

	//cl, err := containers.List(cr.conn, nil)
	//if err != nil {
	//	Kube.FmtRed(err.Error())
	//}
	//
	//for _, c := range cl {
	//	for _, n := range c.Networks {
	//		c.Networks[n]
	//	}
	//}

	return false, nil
}

func (cr *PodmanImpl) IsContainerRunning(containerID string) (bool, error) {
	// Get list of all containers
	opts := &containers.ListOptions{
		All: ptr.To(true),
	}

	cl, err := containers.List(cr.conn, opts)
	if err != nil {
		log.Fatalf("Failed to list containers: %v", err)
	}

	// Check if container exists and is running
	foundRunning := false
	for _, container := range cl {
		if container.ID == containerID || container.Names[0] == containerID {
			if container.State == "running" {
				foundRunning = true
				break
			}
		}
	}

	if foundRunning {
		fmt.Printf("Container %s is running\n", containerID)
		return true, nil
	} else {
		fmt.Printf("Container %s is not running or does not exist\n", containerID)
		return false, nil
	}
}

func (i *PodmanImpl) HasImageInKindRegistry(image string) (bool, error) {
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

func (cr *PodmanImpl) HasImage(image string) (bool, error) {

	getopt := new(images.GetOptions)

	img, err := images.GetImage(cr.conn, image, getopt)
	if nil != err {
		return false, err
	}

	if img == nil {
		return false, nil
	} else {
		return true, nil
	}
}

func (cr *PodmanImpl) PushImage(src, dest string) error {
	pushOpts := images.PushOptions{
		All:           ptr.To(true),
		SkipTLSVerify: ptr.To(true),
	}

	if err := images.Push(cr.conn, src, dest, &pushOpts); err != nil {
		return err
	} else {
		return nil
	}
}

func (cr *PodmanImpl) PullImage(image string, credentials *PrivateRegistryCredentials) error {

	var opts *images.PullOptions = nil

	if credentials != nil {
		opts = &images.PullOptions{
			Password: ptr.To(credentials.Password),
			Username: ptr.To(credentials.Username),
		}
	}

	if _, err := images.Pull(cr.conn, image, opts); err != nil {
		return err
	} else {
		return nil
	}
}

func (cr *PodmanImpl) FindContainer(name string) (*ContainerSearch, error) {

	list, err := containers.List(cr.conn, &containers.ListOptions{
		All: ptr.To(false),
	})

	if err != nil {
		return &ContainerSearch{
			Found:       false,
			IsRunning:   false,
			ContainerID: "",
		}, err
	}

	for _, container := range list {
		if container.Names[0] == name {
			return &ContainerSearch{
				Found:       true,
				IsRunning:   container.State == "running",
				ContainerID: container.ID,
			}, nil
		}
	}

	return &ContainerSearch{
		Found:       false,
		IsRunning:   false,
		ContainerID: "",
	}, nil

}

func (cr *PodmanImpl) StartContainer(id string) error {
	err := containers.Start(cr.conn, id, &containers.StartOptions{
		DetachKeys: nil,
		Recursive:  nil,
	})

	if err != nil {
		return err
	}

	return nil
}

func (cr *PodmanImpl) StopContainer(id string) error {
	err := containers.Stop(cr.conn, id, &containers.StopOptions{})

	if err != nil {
		return err
	}

	return nil
}

func (cr *PodmanImpl) RemoveContainer(id string) error {
	_, err := containers.Remove(cr.conn, id, &containers.RemoveOptions{
		Force:   ptr.To(true),
		Volumes: ptr.To(true),
	})

	if err != nil {
		return err
	}

	return nil
}

func (cr *PodmanImpl) Exec(containerId string, cmd []string) error {

	execOpts := new(handlers.ExecCreateConfig)
	execOpts.Cmd = cmd
	execOpts.Tty = false
	execOpts.AttachStderr = true
	execOpts.AttachStdout = true

	session, err := containers.ExecCreate(cr.conn, containerId, execOpts)
	if err != nil {
		Kube.FmtRed(err.Error())
		os.Exit(1)
	}

	err = containers.ExecStart(cr.conn, session, nil)
	if err != nil {
		Kube.FmtRed(err.Error())
		os.Exit(1)
	}

	return nil

}

func (cr *PodmanImpl) ContainerWriteFile(containerId string, dst string, filename string, data []byte) error {

	dat, err := memtar(data, filename)
	if err != nil {
		panic(err)
	}

	cp, cperr := containers.CopyFromArchive(cr.conn, containerId, dst, bytes.NewReader(dat))

	if cperr != nil {
		Kube.FmtRed(cperr.Error())
	}

	err = cp()

	if err != nil {
		Kube.FmtRed(err.Error())
		os.Exit(1)
	}

	return nil

}

func (cr *PodmanImpl) NetworkConnect(containerId string, networkId string) error {

	err := network.Connect(cr.conn, constants.KIND_NETWORK_NAME, containerId, nil)
	if err != nil {

		if strings.Contains(err.Error(), "already connected to network") {
			return nil
		} else {
			return err
		}
	}

	return nil
}

func (cr *PodmanImpl) CloseContainerRuntime() {
}

func (cr *PodmanImpl) IsContainerRuntimeAvailable() bool {
	if cr.conn != nil {
		return true
	} else {
		return false
	}
}

func (cr *PodmanImpl) CreateContainerRegistry() error {

	registry := constants.REGISTRY_IMAGE
	containerName := constants.REGISTRY_CONTAINER

	// make sure that the registry-config file exists
	configDir, _ := os.UserConfigDir()
	err := CopyResource("registry-config.yaml", "registry-config.yaml")
	if err != nil {
		return err
	}

	err = CopyResource("cert/server.crt", "localtest.me.crt")
	if err != nil {
		return err
	}

	err = CopyResource("cert/server.key", "localtest.me.key")
	if err != nil {
		return err
	}

	imageSearch, err := cr.HasImage(registry)
	if err != nil {
		return err
	}

	if !imageSearch {
		pErr := cr.PullImage(registry, nil)
		if pErr != nil {
			return pErr
		}
	}

	containerSearch, err := cr.FindContainer(containerName)
	if err != nil {
		return err
	}

	if !containerSearch.Found {

		spec := specgen.NewSpecGenerator(registry, false) // false = no systemd integration
		spec.Name = containerName
		spec.Privileged = ptr.To(true)
		spec.NetNS.NSMode = specgen.Bridge

		spec.PortMappings = append(spec.PortMappings, types.PortMapping{
			ContainerPort: 5000,
			HostPort:      5001,
		})

		resp, err := containers.CreateWithSpec(cr.conn, spec, nil)
		if err != nil {
			return err
		}

		if err := containers.Start(cr.conn, resp.ID, nil); err != nil {
			return err
		}

		file, _ := ReadFileToBytes(filepath.Join(configDir, "easykube", "registry-config.yaml"))

		if err := cr.ContainerWriteFile(constants.REGISTRY_CONTAINER, "/etc/docker/registry", "config.yml", file); err != nil {
			return err
		}

		file, _ = ReadFileToBytes(filepath.Join(configDir, "easykube", "localtest.me.crt"))
		if err := cr.ContainerWriteFile(constants.REGISTRY_CONTAINER, "/etc/ssl", "localtest.me.crt", file); err != nil {
			return err
		}

		file, _ = ReadFileToBytes(filepath.Join(configDir, "easykube", "localtest.me.key"))
		if err := cr.ContainerWriteFile(constants.REGISTRY_CONTAINER, "/etc/ssl", "localtest.me.key", file); err != nil {
			return err
		}

	}

	return nil
}

func (cr *PodmanImpl) Commit(containerID string) {
	// nop
}

func (cr *PodmanImpl) TagImage(source string, target string) error {
	target = strings.TrimPrefix(target, constants.LOCAL_REGISTRY+"/")
	parts := strings.Split(target, ":")
	imageName := parts[0]
	imageTag := "latest"
	if len(parts) > 1 {
		imageTag = parts[1]
	}

	repo := fmt.Sprintf("%s/%s", constants.LOCAL_REGISTRY, imageName)

	if err := images.Tag(cr.conn, source, imageTag, repo, nil); err != nil {
		return err
	} else {
		return nil
	}

}

func (p *PodmanImpl) memtar(data []byte, filename string) ([]byte, error) {
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

	if err := tw.WriteHeader(hdr); err != nil {
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
