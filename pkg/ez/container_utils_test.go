package ez

import (
	"fmt"
	"testing"
)

func TestPullImage(*testing.T) {
	pm := NewPodmanImpl()

	fmt.Println("Pulling nginx")
	pm.PullImage("docker.io/nginx", nil)
}

func TestContainerExists(*testing.T) {
	pm := NewPodmanImpl()
	exists := pm.HasImageInKindRegistry("registry:2")

	fmt.Println(exists)
}

func TestPushImagePodman(*testing.T) {
	pm := NewPodmanImpl()
	fmt.Println("Pushing image to local registry")

	pm.PushImage("busybox")
}

func TestPushImageDocker(*testing.T) {
}

func TestCreateRegistry(*testing.T) {
}

func TestFindContainer(t *testing.T) {
}

func TestHasImage(t *testing.T) {
}

func TestPushLocal(t *testing.T) {
}

func TestHasKindImage(*testing.T) {
}
