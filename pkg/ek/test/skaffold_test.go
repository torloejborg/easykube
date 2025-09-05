package test

import (
	"github.com/torloejborg/easykube/pkg/ek"
	"testing"
)

func TestCreateAddon(t *testing.T) {

	conf := ek.NewEasykubeConfig(GetEKContext())
	cfg, _ := conf.LoadConfig()
	skaf := ek.NewSkaffold(cfg.AddonDir)
	skaf.CreateNewAddon("foo", "middleware")
}
