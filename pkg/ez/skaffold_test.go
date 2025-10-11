package ez_test

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/torloejborg/easykube/pkg/ez"
	"github.com/torloejborg/easykube/test"
)

func initCreateAddonsTest(t *testing.T) {
	osd := test.CreateOsDetailsMock(t)
	osd.EXPECT().GetUserConfigDir().Return("/home/some-user/.config", nil).AnyTimes()
	osd.EXPECT().GetUserHomeDir().Return("/home/some-user", nil).AnyTimes()
	config := ez.NewEasykubeConfig(osd)

	ez.Kube.UseOsDetails(osd)
	ez.Kube.UseFilesystemLayer(afero.NewMemMapFs())
	ez.Kube.UseEasykubeConfig(config)

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
	initCreateAddonsTest(t)

	config, _ := ez.Kube.LoadConfig()

	cut := ez.NewSkaffold(config.AddonDir)

	for _, tt := range addonsToCreate {
		t.Run(tt.addonName, func(t *testing.T) {
			cut.CreateNewAddon(tt.addonName, tt.location)
		})

		for _, yy := range expectedAddonFiles {
			t.Run(yy.file, func(t *testing.T) {
				ok := ez.FileOrDirExists(filepath.Join(config.AddonDir, tt.location, tt.addonName, yy.file))
				t.Log(filepath.Join(config.AddonDir, tt.location, tt.addonName, yy.file))
				if !ok {
					t.Errorf("nope")
				}
			})
		}
	}
}
