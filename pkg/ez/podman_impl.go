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
	"strings"
	"time"

	"github.com/containers/podman/v6/pkg/bindings"
	"github.com/containers/podman/v6/pkg/bindings/containers"
	"github.com/spf13/afero"
	"github.com/torloejborg/easykube/pkg/constants"
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

	return false
}

func (cr *PodmanImpl) IsContainerRunning(containerID string) bool {
	// Get list of all containers
	opts := &containers.ListOptions{
		All: ptr.To(true),
	}

	containers, err := containers.List(cr.conn, opts)
	if err != nil {
		log.Fatalf("Failed to list containers: %v", err)
	}

	// Container ID or name we want to check

	// Check if container exists and is running
	foundRunning := false
	for _, container := range containers {
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

	return false
}

func (cr *PodmanImpl) PushImage(image string) {

}

func (cr *PodmanImpl) PullImage(image string, privateRegistryCredentials *string) {

}

func (cr *PodmanImpl) FindContainer(name string) (*ContainerSearch, error) {
	return &ContainerSearch{
		Found:       false,
		IsRunning:   false,
		ContainerID: "",
	}, nil

}

func (cr *PodmanImpl) StartContainer(id string) {

}

func (cr *PodmanImpl) StopContainer(id string) {

}

func (cr *PodmanImpl) RemoveContainer(id string) {

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
	if cr.conn != nil {
		return true
	} else {
		return false
	}
}

func (cr *PodmanImpl) CreateContainerRegistry() error {
	return nil
}

func (cr *PodmanImpl) Commit(containerID string) {

}

func (cr *PodmanImpl) TagImage(source string, target string) {

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
