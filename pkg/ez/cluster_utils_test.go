package ez_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/torloejborg/easykube/pkg/core"
	"github.com/torloejborg/easykube/pkg/ez"
	"github.com/torloejborg/easykube/test"
)

func initClusterUtilsTest(t *testing.T) *core.Ek {

	ek := &core.Ek{
		OsDetails: test.CreateOsDetailsMock(t),
		Fs:        afero.NewMemMapFs(),
	}

	ek.Utils = ez.NewUtils(ek)
	ek.Config = ez.NewEasykubeConfig(ek)
	_ = ek.Config.MakeConfig()

	ek.AddonReader = ez.NewAddonReader(ek)
	ek.ClusterUtils = ez.NewClusterUtils(ek)

	return ek
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
	ek := initClusterUtilsTest(t)

	test.CopyTestAddonToMemFs("../../test_addons", "diamond", "/home/some-user/addons", ek.Fs)
	err := ek.ClusterUtils.EnsurePersistenceDirectory()

	if err != nil {
		t.Errorf("Failed to create directories %v", err)
	}

	for _, tt := range expectedPersistenceDirectories {
		t.Run(tt.dir, func(t *testing.T) {
			cfg, _ := ek.OsDetails.GetEasykubeConfigDir()
			persistenceDir := filepath.Join(cfg, "persistence", tt.dir)
			exists := ek.Utils.FileOrDirExists(persistenceDir)
			if !exists {
				t.Errorf("expected %v to exist", persistenceDir)
			}
		})
	}
}

func TestRenderKindConfigurationFromSetOfAddons(t *testing.T) {
	ek := initClusterUtilsTest(t)
	test.CopyTestAddonToMemFs("../../test_addons", "diamond", "/home/some-user/addons", ek.Fs)
	addons, err := ek.AddonReader.GetAddons()
	if err != nil {
		t.Errorf("Failed to get addons %v", err)
	}
	addonList := make([]core.IAddon, 0)
	// unmap addons
	for _, addon := range addons {
		addonList = append(addonList, addon)
	}

	cfg, _ := ek.Config.LoadConfig()
	result := ek.ClusterUtils.RenderToYAML(addonList, cfg)

	fmt.Println(result)
}

func TestWrongPortConfigShouldFail(t *testing.T) {

	ek := initClusterUtilsTest(t)
	test.CopyTestAddonToMemFs("../../test_addons", "portconfig", "/home/some-user/addons", ek.Fs)

	_, err := ek.AddonReader.GetAddons()
	if err != nil {

		expectedContains := "requires both hostPort and nodePort to be set"

		if !strings.Contains(err.Error(), expectedContains) {
			t.Errorf("Got error %v, expected '%v' in errormessage", err, expectedContains)
		}
	}
}
