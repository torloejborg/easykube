package ez

import (
	"testing"

	"github.com/spf13/afero"
)

func initAddonReaderTest() {

	y := &OsDetailsStub{CreateOsDetailsImpl()}
	x := &EasykubeConfigStub{CreateEasykubeConfigImpl(y)}

	Kube.UseOsDetails(y)
	Kube.UseFilesystemLayer(afero.NewMemMapFs())
	Kube.UseEasykubeConfig(x)
	Kube.UseAddonReader(CreateAddonReaderImpl(x))
}

func TestDiscoverAddons(t *testing.T) {

	initAddonReaderTest()
	Kube.MakeConfig()

	CopyTestAddonToMemFs("circular", "./addons")

	all := Kube.GetAddons()
	for _, addon := range all {
		t.Log(addon.Name)
	}

}
