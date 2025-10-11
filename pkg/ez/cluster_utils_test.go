package ez_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/torloejborg/easykube/pkg/ez"
	"github.com/torloejborg/easykube/test"
)

func initClusterUtilsTest(t *testing.T) {

	osd := test.CreateOsDetailsMock(t)
	config := ez.NewEasykubeConfig(osd)

	ez.Kube.UseOsDetails(osd)
	ez.Kube.UseFilesystemLayer(afero.NewMemMapFs())
	ez.Kube.UseEasykubeConfig(config)
	ez.Kube.UseAddonReader(ez.CreateAddonReaderImpl(config))
	ez.Kube.UseClusterUtils(ez.CreateClusterUtilsImpl())

	_ = ez.Kube.MakeConfig()
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
	initClusterUtilsTest(t)

	test.CopyTestAddonToMemFs("../../test_addons", "diamond", "/home/some-user/addons", ez.Kube.Fs)
	err := ez.Kube.EnsurePersistenceDirectory()

	if err != nil {
		t.Errorf("Failed to create directories %v", err)
	}

	for _, tt := range expectedPersistenceDirectories {
		t.Run(tt.dir, func(t *testing.T) {
			cfg, _ := ez.Kube.GetUserConfigDir()
			persistenceDir := filepath.Join(cfg, "easykube", "persistence", tt.dir)
			exists := ez.FileOrDirExists(persistenceDir)
			if !exists {
				t.Errorf("expected %v to exist", persistenceDir)
			}
		})
	}
}

func TestRenderKindConfigurationFromSetOfAddons(t *testing.T) {
	initClusterUtilsTest(t)
	test.CopyTestAddonToMemFs("../../test_addons", "diamond", "/home/some-user/addons", ez.Kube.Fs)
	addons, err := ez.Kube.GetAddons()
	if err != nil {
		t.Errorf("Failed to get addons %v", err)
	}
	addonList := make([]*ez.Addon, 0)
	// unmap addons
	for _, addon := range addons {
		addonList = append(addonList, addon)
	}

	result := ez.Kube.RenderToYAML(addonList)

	fmt.Println(result)
}

func TestWrongPortConfigShouldFail(t *testing.T) {

	initClusterUtilsTest(t)
	test.CopyTestAddonToMemFs("../../test_addons", "portconfig", "/home/some-user/addons", ez.Kube.Fs)

	_, err := ez.Kube.GetAddons()
	if err != nil {

		expectedContains := "requires both hostPort and nodePort to be set"

		if !strings.Contains(err.Error(), expectedContains) {
			t.Errorf("Got error %v, expected '%v' in errormessage", err, expectedContains)
		}
	}
}
