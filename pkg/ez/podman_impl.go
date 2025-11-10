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
	return cr.IsContainerRunning(constants.KIND_CONTAINER)
}

func (cr *PodmanImpl) IsNetworkConnectedToContainer(containerID string, networkID string) bool {

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

	return false
}

func (cr *PodmanImpl) IsContainerRunning(containerID string) bool {
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
		return true
	} else {
		fmt.Printf("Container %s is not running or does not exist\n", containerID)
		return false
	}
}

func (i *PodmanImpl) HasImageInKindRegistry(image string) bool {
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

func (cr *PodmanImpl) HasImage(image string) bool {

	getopt := new(images.GetOptions)

	img, err := images.GetImage(cr.conn, image, getopt)
	if nil != err {
		Kube.FmtRed(err.Error())
		return false
	}

	if img == nil {
		return false
	} else {
		return true
	}
}

// TODO: Return error, pass destination
func (cr *PodmanImpl) PushImage(src, dest string) {
	pushOpts := images.PushOptions{
		All:           ptr.To(true),
		SkipTLSVerify: ptr.To(true),
	}

	err := images.Push(cr.conn, src, dest, &pushOpts)
	if err != nil {
		Kube.FmtRed(err.Error())
		os.Exit(1)
	}

}

// TODO: Return error
func (cr *PodmanImpl) PullImage(image string, privateRegistryCredentials *string) {

	_, err := images.Pull(cr.conn, image, nil)
	if err != nil {
		Kube.FmtRed(err.Error())
		os.Exit(1)
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

// TODO: Return error
func (cr *PodmanImpl) StartContainer(id string) {
	err := containers.Start(cr.conn, id, &containers.StartOptions{
		DetachKeys: nil,
		Recursive:  nil,
	})

	if err != nil {
		Kube.FmtRed(err.Error())
		os.Exit(1)
	}
}

// TODO: Return error
func (cr *PodmanImpl) StopContainer(id string) {
	err := containers.Stop(cr.conn, id, &containers.StopOptions{})

	if err != nil {
		Kube.FmtRed(err.Error())
		os.Exit(1)
	}
}

// TODO: Return error
func (cr *PodmanImpl) RemoveContainer(id string) {
	_, err := containers.Remove(cr.conn, id, &containers.RemoveOptions{
		Force:   ptr.To(true),
		Volumes: ptr.To(true),
	})

	if err != nil {
		Kube.FmtRed(err.Error())
		os.Exit(1)
	}

}

// TODO: Return error
func (cr *PodmanImpl) Exec(containerId string, cmd []string) {

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

}

func (cr *PodmanImpl) ContainerWriteFile(containerId string, dst string, filename string, data []byte) {

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

}

func (cr *PodmanImpl) NetworkConnect(containerId string, networkId string) {

	err := network.Connect(cr.conn, constants.KIND_NETWORK_NAME, containerId, nil)
	if err != nil {

		if strings.Contains(err.Error(), "already connected to network") {
			return
		} else {
			Kube.FmtRed(err.Error())
			os.Exit(1)
		}
	}
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

	imageSearch := cr.HasImage(registry)
	if !imageSearch {
		cr.PullImage(registry, nil)
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

		err = containers.Start(cr.conn, resp.ID, nil)

		if err != nil {
			return err
		}

		file, _ := ReadFileToBytes(filepath.Join(configDir, "easykube", "registry-config.yaml"))
		cr.ContainerWriteFile(constants.REGISTRY_CONTAINER, "/etc/docker/registry", "config.yml", file)

		file, _ = ReadFileToBytes(filepath.Join(configDir, "easykube", "localtest.me.crt"))
		cr.ContainerWriteFile(constants.REGISTRY_CONTAINER, "/etc/ssl", "localtest.me.crt", file)

		file, _ = ReadFileToBytes(filepath.Join(configDir, "easykube", "localtest.me.key"))
		cr.ContainerWriteFile(constants.REGISTRY_CONTAINER, "/etc/ssl", "localtest.me.key", file)

	}

	return nil
}

func copyFileToPodmanVolume(volumeName, srcPath, dstFilename string) error {

	// TODO: implement

	return nil
}

func (cr *PodmanImpl) Commit(containerID string) {

}

func (cr *PodmanImpl) TagImage(source string, target string) {
	images.Tag(cr.conn, source, target, constants.LOCAL_REGISTRY, nil)
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
