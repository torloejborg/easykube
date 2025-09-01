package test

import (
	"testing"
	"time"

	"github.com/torloj/easykube/pkg/ek"
)

func TestWaitforCustomResource(t *testing.T) {
	utils := ek.NewK8SUtils(GetEKContext())

	err := utils.WaitForCRD("cert-manager.io", "v1", "ClusterIssuer", 100*time.Second)
	if err != nil {
		panic(err)
	}

}
