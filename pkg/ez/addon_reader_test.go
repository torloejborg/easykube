package ez_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/afero"
	mock "github.com/torloejborg/easykube/mock"
	"github.com/torloejborg/easykube/pkg/core"
	"github.com/torloejborg/easykube/pkg/ez"
	"github.com/torloejborg/easykube/pkg/vars"
	"github.com/torloejborg/easykube/test"
	"go.uber.org/mock/gomock"
)

func initAddonReaderTest(t *testing.T) *core.Ek {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	command := mock.NewMockICobraCommandHelper(mockCtrl)

	ek := &core.Ek{
		OsDetails:      test.CreateOsDetailsMock(t),
		CommandContext: command,
		Fs:             afero.NewMemMapFs(),
	}
	ek.Utils = ez.NewUtils(ek)
	ek.Config = ez.NewEasykubeConfig(ek)
	ek.Config.MakeConfig()
	ek.AddonReader = ez.NewAddonReader(ek)

	return ek
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

	cut := initAddonReaderTest(t)

	test.CopyTestAddonToMemFs("../../test_addons", "diamond", "/home/some-user/addons", cut.Fs)

	all, err := cut.AddonReader.GetAddons()
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

	cut := initAddonReaderTest(t)
	test.CopyTestAddonToMemFs("../../test_addons", "broken", "/home/some-user/addons", cut.Fs)

	_, err := cut.AddonReader.GetAddons()
	if err != nil {

		if !strings.Contains(err.Error(), "invalid character 'x' looking for beginning of object key string") {
			t.Error("expected different errormessage from JS runtime")
		}
	} else {
		t.Error("expected error, the broken addon should not parse it's configuration")
	}
}

func TestVersionCompatibilityReader(t *testing.T) {

	cut := initAddonReaderTest(t)

	test.CopyTestAddonToMemFs("../../test_addons", "diamond", "/home/some-user/addons", cut.Fs)

	vars.Version = "2.3.0"

	version, err := cut.AddonReader.CheckAddonCompatibility()
	if err != nil {
		fmt.Println(err.Error())
		if !strings.Contains(err.Error(), "addon repository want easykube ~3.0 but easykube is 2.3.0") {
			t.Fail()
		}
	}

	fmt.Println(version)
}
