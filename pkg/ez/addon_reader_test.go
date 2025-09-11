package ez

import (
	"strings"
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

var expectedAddonsForDiscoverTest = []struct {
	name string
}{
	{"a"},
	{"b"},
	{"c"},
	{"d"},
}

func TestDiscoverAddons(t *testing.T) {

	initAddonReaderTest()
	Kube.MakeConfig()

	CopyTestAddonToMemFs("diamond", "./addons")

	all, err := Kube.GetAddons()
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range expectedAddonsForDiscoverTest {
		t.Run(tt.name, func(t *testing.T) {
			if all[tt.name] == nil {
				t.Errorf("expected addon %v to be in the list", tt.name)
			}
		})
	}
}

func TestABrokenAddon(t *testing.T) {

	initAddonReaderTest()
	Kube.MakeConfig()

	CopyTestAddonToMemFs("broken", "./addons")

	_, err := Kube.GetAddons()
	if err != nil {

		if !strings.Contains(err.Error(), "invalid character 'x' looking for beginning of object key string") {
			t.Error("expected different errormessage from JS runtime")
		}
	} else {
		t.Error("expected error, the broken addon should not parse it's configuration")
	}
}
