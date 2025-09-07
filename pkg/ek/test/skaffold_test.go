package test

import (
	"testing"

	"github.com/torloejborg/easykube/pkg/ek"
)

func TestCreateAddon(t *testing.T) {

	conf := CreateFakeEasykubeConfig()
	cfg, _ := conf.LoadConfig()
	skaf := ek.NewSkaffold(cfg.AddonDir, FILESYSTEM)
	skaf.CreateNewAddon("foo", "middleware")

}
