package jsutils_test

import (
	"testing"

	"github.com/spf13/afero"
	mock_ez "github.com/torloejborg/easykube/mock"
	"github.com/torloejborg/easykube/pkg/ez"
	"github.com/torloejborg/easykube/test"
	"go.uber.org/mock/gomock"
)

func SetUpJsTestEnvironment(t *testing.T, controller *gomock.Controller) {

	osd := test.CreateOsDetailsMock(t)
	osd.EXPECT().GetUserConfigDir().Return("/home/some-user/.config", nil).AnyTimes()
	osd.EXPECT().GetUserHomeDir().Return("/home/some-user", nil).AnyTimes()
	config := ez.NewEasykubeConfig(osd)

	ez.Kube.UseOsDetails(osd)
	ez.Kube.UseFilesystemLayer(afero.NewMemMapFs())
	ez.Kube.UseEasykubeConfig(config)
	ez.Kube.UseAddonReader(ez.CreateAddonReaderImpl(config))
	ez.Kube.UseClusterUtils(ez.CreateClusterUtilsImpl())
	ez.Kube.UseContainerRuntime(ez.CreateContainerRuntimeImpl())

	_ = ez.Kube.MakeConfig()
	test.CopyTestAddonToMemFs("../../test_addons", "diamond", "/home/some-user/addons", ez.Kube.Fs)
}

func CreateSyntheticAddon(script string, controller *gomock.Controller) *mock_ez.MockIAddon {

	mock := mock_ez.NewMockIAddon(controller)
	mock.EXPECT().GetName().Return("synthetic").AnyTimes()
	mock.EXPECT().ReadScriptFile(gomock.Any()).Return(script).AnyTimes()
	mock.EXPECT().GetRootDir().Return("/home/some-user/addons").AnyTimes()

	return mock
}
