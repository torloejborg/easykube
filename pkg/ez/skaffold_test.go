package ez

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

func initCreateAddonsTest() {
	Kube = &Toolbox{}

	y := &OsDetailsStub{CreateOsDetailsImpl()}
	x := &EasykubeConfigStub{CreateEasykubeConfigImpl(y)}

	Kube.UseOsDetails(y)
	Kube.UseFilesystemLayer(afero.NewMemMapFs())
	Kube.UseEasykubeConfig(x)

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
	initCreateAddonsTest()

	Kube.MakeConfig()
	config, _ := Kube.LoadConfig()

	cut := NewSkaffold(config.AddonDir)

	for _, tt := range addonsToCreate {
		t.Run(tt.addonName, func(t *testing.T) {
			cut.CreateNewAddon(tt.addonName, tt.location)
		})

		for _, yy := range expectedAddonFiles {
			t.Run(yy.file, func(t *testing.T) {
				ok := FileOrDirExists(filepath.Join(config.AddonDir, tt.location, tt.addonName, yy.file))
				t.Log(filepath.Join(config.AddonDir, tt.location, tt.addonName, yy.file))
				if !ok {
					t.Errorf("nope")
				}
			})
		}
	}
}
