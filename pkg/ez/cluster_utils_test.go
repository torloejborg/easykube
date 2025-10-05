package ez

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func initClusterUtilsTest() {

	y := &OsDetailsStub{CreateOsDetailsImpl()}
	x := &EasykubeConfigStub{CreateEasykubeConfigImpl(y)}

	Kube.UseOsDetails(y)
	Kube.UseFilesystemLayer(afero.NewMemMapFs())
	Kube.UseEasykubeConfig(x)
	Kube.UseAddonReader(CreateAddonReaderImpl(x))
	Kube.UseClusterUtils(CreateClusterUtilsImpl())
}

var expectedPersistenceDirectories = []struct {
	dir string
}{
	{"addon-a"},
	{"addon-b"},
	{"addon-c"},
	{"addon-d"},
}

func TestCreatePersistenceDirectories(t *testing.T) {
	initClusterUtilsTest()
	Kube.MakeConfig()
	CopyTestAddonToMemFs("diamond", "./addons")

	err := Kube.EnsurePersistenceDirectory()

	if err != nil {
		t.Errorf("Failed to create directories %v", err)
	}

	for _, tt := range expectedPersistenceDirectories {
		t.Run(tt.dir, func(t *testing.T) {
			cfg, _ := Kube.GetUserConfigDir()
			persistenceDir := filepath.Join(cfg, "easykube", "persistence", tt.dir)
			exists := FileOrDirExists(persistenceDir)
			if !exists {
				t.Errorf("expected %v to exist", persistenceDir)
			}
		})
	}
}

func TestRenderKindConfigurationFromSetOfAddons(t *testing.T) {
	initClusterUtilsTest()
	_ = Kube.MakeConfig()
	CopyTestAddonToMemFs("diamond", "./addons")
	addons, err := Kube.GetAddons()
	if err != nil {
		t.Errorf("Failed to get addons %v", err)
	}
	addonList := make([]*Addon, 0)
	// unmap addons
	for _, addon := range addons {
		addonList = append(addonList, addon)
	}

	result := Kube.RenderToYAML(addonList)

	fmt.Println(result)
}

func TestWrongPortConfigShouldFail(t *testing.T) {

	initClusterUtilsTest()
	Kube.MakeConfig()
	CopyTestAddonToMemFs("portconfig", "./addons")

	_, err := Kube.GetAddons()
	if err != nil {

		expectedContains := "requires both hostPort and nodePort to be set"

		if !strings.Contains(err.Error(), expectedContains) {
			t.Errorf("Got error %v, expected '%v' in errormessage", err, expectedContains)
		}
	}
}
