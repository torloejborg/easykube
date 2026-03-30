package ez_test

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/torloejborg/easykube/pkg/core"
	"github.com/torloejborg/easykube/pkg/ez"
	"github.com/torloejborg/easykube/test"
)

func initCreateAddonsTest(t *testing.T) *core.Ek {
	osd := test.CreateOsDetailsMock(t)
	osd.EXPECT().GetEasykubeConfigDir().Return("/home/some-user/.config", nil).AnyTimes()
	osd.EXPECT().GetUserHomeDir().Return("/home/some-user", nil).AnyTimes()

	ek := &core.Ek{
		OsDetails: osd,
		Fs:        afero.NewMemMapFs(),
	}

	ek.Utils = ez.NewUtils(ek)
	ek.Config = ez.NewEasykubeConfig(ek)
	_ = ek.Config.MakeConfig()

	return ek
}

var addonsToCreate = []struct {
	addonName string
	location  string
}{
	{"foo", "utils"},
	{"super-db", "middleware/databases"},
	{"lab-result-tester", "research/statistics/number-crunching"},
}

var expectedAddonFiles = []struct {
	file string
}{
	{"manifests/configmap.yaml"},
	{"manifests/deployment.yaml"},
	{"manifests/ingress.yaml"},
	{"manifests/service.yaml"},
	{"kustomization.yaml"},
}

func TestCreateAddon(t *testing.T) {
	ek := initCreateAddonsTest(t)

	config, err := ek.Config.LoadConfig()
	if err != nil {
		t.Fatalf("error loading config: %v", err)
	}

	cut := ez.NewSkaffold(ek, config.AddonDir)

	for _, tt := range addonsToCreate {
		t.Run(tt.addonName, func(t *testing.T) {
			cut.CreateNewAddon(tt.addonName, tt.location)
		})

		for _, yy := range expectedAddonFiles {
			t.Run(yy.file, func(t *testing.T) {
				ok := ek.Utils.FileOrDirExists(filepath.Join(config.AddonDir, tt.location, tt.addonName, yy.file))
				t.Log(filepath.Join(config.AddonDir, tt.location, tt.addonName, yy.file))
				if !ok {
					t.Errorf("nope")
				}
			})
		}
	}
}
