package jsutils_test

import (
	"testing"

	"github.com/spf13/afero"
	mock "github.com/torloejborg/easykube/mock"
	"github.com/torloejborg/easykube/pkg/core"
	"github.com/torloejborg/easykube/pkg/textutils"

	"github.com/torloejborg/easykube/pkg/ez"
	"github.com/torloejborg/easykube/test"
	"go.uber.org/mock/gomock"
)

func SetUpJsTestEnvironment(t *testing.T, controller *gomock.Controller) *core.Ek {

	osd := test.CreateOsDetailsMock(t)
	ek := &core.Ek{
		Fs:        afero.NewMemMapFs(),
		OsDetails: osd,
		Printer:   textutils.NewPrinter(),
	}
	ek.Utils = ez.NewUtils(ek)
	ek.Config = ez.NewEasykubeConfig(ek)
	_ = ek.Config.MakeConfig()

	ek.AddonReader = ez.NewAddonReader(ek)
	ek.ClusterUtils = ez.NewClusterUtils(ek)

	test.CopyTestAddonToMemFs("../../test_addons", "diamond", "/home/some-user/addons", ek.Fs)

	return ek
}

func CreateSyntheticAddon(script string, controller *gomock.Controller) *mock.MockIAddon {

	mockAddon := mock.NewMockIAddon(controller)
	mockAddon.EXPECT().GetName().Return("synthetic").AnyTimes()
	mockAddon.EXPECT().ReadScriptFile(gomock.Any()).Return(script).AnyTimes()
	mockAddon.EXPECT().GetRootDir().Return("/home/some-user/addons").AnyTimes()

	return mockAddon
}
