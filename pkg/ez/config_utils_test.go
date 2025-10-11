package ez_test

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/torloejborg/easykube/pkg/ez"
	"github.com/torloejborg/easykube/test"
)

func initConfigTests(t *testing.T) {

	osd := test.CreateOsDetailsMock(t)
	osd.EXPECT().GetUserConfigDir().Return("/home/some-user/.config", nil).AnyTimes()
	osd.EXPECT().GetUserHomeDir().Return("/home/some-user", nil).AnyTimes()
	config := ez.NewEasykubeConfig(osd)

	ez.Kube.UseOsDetails(osd)
	ez.Kube.UseFilesystemLayer(afero.NewMemMapFs())
	ez.Kube.UseEasykubeConfig(config)
	ez.Kube.UseAddonReader(ez.CreateAddonReaderImpl(config))
	ez.Kube.UseClusterUtils(ez.CreateClusterUtilsImpl())

	_ = ez.Kube.MakeConfig()

}

func TestMakeDefaultConfig(t *testing.T) {
	initConfigTests(t)
	cfgdir, _ := ez.Kube.GetUserConfigDir()
	homeDir, _ := ez.Kube.GetUserHomeDir()

	exists := ez.FileOrDirExists(filepath.Join(cfgdir, "easykube", "config.yaml"))
	if !exists {
		t.Errorf("expected easykube config file to exist")
	}

	data, err := ez.Kube.LoadConfig()
	if err != nil {
		panic(err)
	}

	if data.AddonDir != filepath.Join(homeDir, "addons") {
		t.Errorf("expected addons dir to be ./addons")
	}

}

var filesExist = []struct {
	file   string
	exists bool
}{
	{"config.yaml", true},
	{"localtest.me.crt", true},
	{"localtest.me.key", true},
	{"registry-config.yaml", true},
	{"persistence", false},
	{"easykube-cluster.yaml", false},
}

func TestVerifyConfigurationFilesCopiedToConfigDir(t *testing.T) {
	initConfigTests(t)
	cfgdir, _ := ez.Kube.GetUserConfigDir()

	for _, tt := range filesExist {
		t.Run(tt.file, func(t *testing.T) {
			found := ez.FileOrDirExists(filepath.Join(cfgdir, "easykube", tt.file))
			if found != tt.exists {
				t.Errorf("expected file %v file to exist in %v, was %v", tt.file, cfgdir, found)
			}
		})
	}
}
