package ez_test

import (
	"testing"

	"github.com/torloejborg/easykube/pkg/ez"
)

func TestContainerRegistry(t *testing.T) {

	ez.Kube = &ez.EasykubeSingleton{}
	ez.InitializeKubeSingleton()

	err := ez.Kube.CreateContainerRegistry()
	if err != nil {
		t.Errorf("%v", err)
	}

	ez.Kube.NetworkConnect("easykube-registry", "kind")

	ez.Kube.PullImage("registry.k8s.io/ingress-nginx/controller:v1.12.1", nil)
	ez.Kube.TagImage("registry.k8s.io/ingress-nginx/controller:v1.12.1", "localhost:5001/registry.k8s.io/ingress-nginx/controller:v1.12.1")
	ez.Kube.PushImage("registry.k8s.io/ingress-nginx/controller:v1.12.1", "localhost:5001/registry.k8s.io/ingress-nginx/controller:v1.12.1")

	ez.Kube.PullImage("docker.io/library/nginx:latest", nil)
	ez.Kube.TagImage("docker.io/library/nginx:latest", "localhost:5001/nginx/nginx:latest")
	ez.Kube.PushImage("docker.io/library/nginx:latest", "localhost:5001/nginx/nginx:latest")

}
