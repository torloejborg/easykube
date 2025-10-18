package jsutils

import (
	"testing"

	"github.com/dop251/goja"
	"github.com/spf13/afero"
	mock_ez "github.com/torloejborg/easykube/mock"
	"github.com/torloejborg/easykube/pkg/ez"
	"github.com/torloejborg/easykube/test"
	"go.uber.org/mock/gomock"
)

func setup(t *testing.T) {

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

func TestPreloadImages(t *testing.T) {
	setup(t)
	ctl := gomock.NewController(t)
	test.CopyTestAddonToMemFs("../../test_addons", "diamond", "/home/some-user/addons", ez.Kube.Fs)
	script := "preloadImages()"
	am := mock_ez.NewMockIAddon(ctl).EXPECT().ReadScriptFile(gomock.Any()).Return(script).AnyTimes()

	jsu := JsUtils{
		vm:                 goja.New(),
		CobraCommandHelper: nil,
		AddonRoot:          "/home/some-user/addons",
	}

	jsu.ExecAddonScript(&am)

}
