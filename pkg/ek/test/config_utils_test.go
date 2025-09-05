package test

import (
	"fmt"
	"github.com/torloejborg/easykube/pkg/ek"
	"testing"
)

func TestLoadConfig(t *testing.T) {

	cut := ek.NewEasykubeConfig(GetEKContext())
	cfg, err := cut.LoadConfig()

	if err != nil {
		t.Error(err)
	}

	fmt.Println(cfg.AddonDir)
	fmt.Println(cfg.PersistenceDir)

}
