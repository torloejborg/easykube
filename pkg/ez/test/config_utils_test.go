package test

import (
	"fmt"
	"testing"
)

func TestLoadConfig(t *testing.T) {

	cut := CreateFakeEasykubeConfig()
	cfg, err := cut.LoadConfig()

	if err != nil {
		t.Error(err)
	}

	fmt.Println(cfg.AddonDir)
	fmt.Println(cfg.PersistenceDir)

}
