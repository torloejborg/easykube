package ez

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

type EasykubeConfigStub struct {
	IEasykubeConfig
}
type OsDetailsStub struct {
	OsDetails
}

func (o *OsDetailsStub) GetUserConfigDir() (string, error) {
	return "/home/some-user/.config", nil
}

func (o *OsDetailsStub) GetUserHomeDir() (string, error) {
	return "/home/some-user", nil
}

func init() {
	Kube = &Toolbox{}

	y := &OsDetailsStub{CreateOsDetails()}
	x := &EasykubeConfigStub{CreateEasykubeConfigImpl(y)}

	Kube.UseOsDetails(y)
	Kube.UseFilesystemLayer(afero.NewMemMapFs())
	Kube.UseEasykubeConfig(x)

}

func TestMakeDefaultConfig(t *testing.T) {
	cfgdir, _ := Kube.GetUserConfigDir()

	Kube.MakeConfig()

	exists := FileOrDirExists(filepath.Join(cfgdir, "easykube", "config.yaml"))
	if !exists {
		t.Errorf("expected easykube config file to exist")
	}

	data, err := Kube.LoadConfig()
	if err != nil {
		panic(err)
	}

	if data.AddonDir != "./addons" {
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

	cfgdir, _ := Kube.GetUserConfigDir()
	Kube.MakeConfig()

	for _, tt := range filesExist {
		t.Run(tt.file, func(t *testing.T) {
			found := FileOrDirExists(filepath.Join(cfgdir, "easykube", tt.file))
			if found != tt.exists {
				t.Errorf("expected file %v file to exist in %v, was %v", tt.file, cfgdir, found)
			}

		})
	}

}
