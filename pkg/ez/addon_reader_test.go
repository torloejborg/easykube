package ez_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/torloejborg/easykube/pkg/ez"
	"github.com/torloejborg/easykube/pkg/vars"
	"github.com/torloejborg/easykube/test"
)

func initAddonReaderTest(t *testing.T) {
	osd := test.CreateOsDetailsMock(t)

	config := ez.NewEasykubeConfig(osd)
	ez.Kube.UseOsDetails(osd)
	ez.Kube.UseFilesystemLayer(afero.NewMemMapFs())
	ez.Kube.UseEasykubeConfig(config)
	ez.Kube.UseAddonReader(ez.CreateAddonReaderImpl(config))
	_ = ez.Kube.MakeConfig()
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

	initAddonReaderTest(t)

	test.CopyTestAddonToMemFs("../../test_addons", "diamond", "/home/some-user/addons", ez.Kube.Fs)

	all, err := ez.Kube.GetAddons()
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

	initAddonReaderTest(t)
	test.CopyTestAddonToMemFs("../../test_addons", "broken", "/home/some-user/addons", ez.Kube.Fs)

	_, err := ez.Kube.GetAddons()
	if err != nil {

		if !strings.Contains(err.Error(), "invalid character 'x' looking for beginning of object key string") {
			t.Error("expected different errormessage from JS runtime")
		}
	} else {
		t.Error("expected error, the broken addon should not parse it's configuration")
	}
}

func TestVersionCompatibilityReader(t *testing.T) {

	initAddonReaderTest(t)

	test.CopyTestAddonToMemFs("../../test_addons", "diamond", "/home/some-user/addons", ez.Kube.Fs)

	vars.Version = "1.1.9"

	version, err := ez.Kube.CheckAddonCompatibility()
	if err != nil {
		fmt.Println(err.Error())
		if !strings.Contains(err.Error(), "addon repository want easykube ~1.1.4 but easykube is 1.4.4") {
			t.Fail()
		}
	}

	fmt.Println(version)
}
