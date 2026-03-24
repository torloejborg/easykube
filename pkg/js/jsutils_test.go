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

	ez.Kube.UseFilesystemLayer(afero.NewMemMapFs())
	ez.Kube.UseOsDetails(osd)

	config := ez.NewEasykubeConfig()
	ez.Kube.UseEasykubeConfig(config)
	_ = ez.Kube.MakeConfig()

	ez.Kube.UseAddonReader(ez.CreateAddonReaderImpl(config))
	ez.Kube.UseClusterUtils(ez.CreateClusterUtilsImpl())

	test.CopyTestAddonToMemFs("../../test_addons", "diamond", "/home/some-user/addons", ez.Kube.Fs)
}

func CreateSyntheticAddon(script string, controller *gomock.Controller) *mock_ez.MockIAddon {

	mock := mock_ez.NewMockIAddon(controller)
	mock.EXPECT().GetName().Return("synthetic").AnyTimes()
	mock.EXPECT().ReadScriptFile(gomock.Any()).Return(script).AnyTimes()
	mock.EXPECT().GetRootDir().Return("/home/some-user/addons").AnyTimes()

	return mock
}
