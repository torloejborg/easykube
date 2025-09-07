package test

import (
	"testing"

	"github.com/torloejborg/easykube/pkg/ez"
)

func TestCreateAddon(t *testing.T) {

	conf := CreateFakeEasykubeConfig()
	cfg, _ := conf.LoadConfig()
	skaf := ez.NewSkaffold(cfg.AddonDir, FILESYSTEM)
	skaf.CreateNewAddon("foo", "middleware")

}
