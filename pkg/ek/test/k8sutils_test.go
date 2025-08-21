package test

import (
	"log"
	"testing"
	"time"

	"github.com/torloj/easykube/pkg/ek"
)

func TestWaitForDelployment(t *testing.T) {

	utils := ek.NewK8SUtils(GetEKContext())
	err := utils.WaitForDeploymentReadyWatch("ingress-nginx-controller", "ingress-nginx")

	if err != nil {
		log.Panic(err)
	}

}

func TestWaitforCustomResource(t *testing.T) {
	utils := ek.NewK8SUtils(GetEKContext())

	err := utils.WaitForCRD("cert-manager.io", "v1", "ClusterIssuer", 100*time.Second)
	if err != nil {
		panic(err)
	}

}
