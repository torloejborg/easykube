package ez

import (
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/torloejborg/easykube/pkg/vars"
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

func TestBrokenAddon(t *testing.T) {

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

func TestVersionCompatibilityReader(t *testing.T) {

	initAddonReaderTest()
	Kube.MakeConfig()

	CopyTestAddonToMemFs("diamond", "./addons")

	vars.Version = "1.4.4"

	version, err := Kube.CheckAddonCompatibility()
	if err != nil {
		if !strings.Contains(err.Error(), "addon repository want easykube ~1.1.4 but easykube is 1.4.4") {
			t.Fail()
		}
	}

	fmt.Println(version)
}
