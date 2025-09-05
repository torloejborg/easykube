package test

import (
	"fmt"
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ek"
	"testing"
)

func TestPullImage(*testing.T) {
	pm := ek.NewPodmanImpl()

	fmt.Println("Pulling nginx")
	pm.Pull("docker.io/nginx", nil)
}

func TestContainerExists(*testing.T) {
	pm := ek.NewPodmanImpl()
	exists := pm.HasImageInKindRegistry("registry:2")

	fmt.Println(exists)
}

func TestPushImagePodman(*testing.T) {
	pm := ek.NewPodmanImpl()
	fmt.Println("Pushing image to local registry")

	pm.Push("busybox")
}

func TestPushImageDocker(*testing.T) {
	pm := ek.NewDockerImpl()
	pm.Pull("busybox:1", nil)
	pm.Tag("busybox:1", "localhost:5000/busybox:1")
	pm.Push("localhost:5000/busybox:1")
	fmt.Println("Pushing image to local registry")

}

func TestCreateRegistry(*testing.T) {
	pm := ek.NewContainerRuntime(GetEKContext())
	pm.CreateContainerRegistry()
}

func TestFindContainer(t *testing.T) {

	pm := ek.NewContainerRuntime(GetEKContext())
	cs := pm.FindContainer("kind-registry")

	fmt.Println(cs.Found)
	fmt.Println(cs.IsRunning)
	fmt.Println(cs.ContainerID)

}

func TestHasImage(t *testing.T) {
	pm := ek.NewContainerRuntime(GetEKContext())
	fmt.Printf("registry:2 exists ? %t\n", pm.HasImage("docker.io/library/registry:2"))
	fmt.Printf("foo/bar exists ? %t\n", pm.HasImage("foo/bar"))
}

func TestPushLocal(t *testing.T) {
	pm := ek.NewContainerRuntime(GetEKContext())
	pm.Push("registry:2")
}

func TestHasKindImage(*testing.T) {
	pm := ek.NewContainerRuntime(GetEKContext())
	fmt.Println(pm.HasImageInKindRegistry(constants.KIND_IMAGE))

}
