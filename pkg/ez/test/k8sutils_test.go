package test

import (
	"testing"
	"time"
)

func TestWaitForCustomResource(t *testing.T) {
	utils := CreateFakeK8sUtil()

	err := utils.WaitForCRD("cert-manager.io", "v1", "ClusterIssuer", 100*time.Second)
	if err != nil {
		panic(err)
	}

}
