package jsutils_test

import (
	"testing"

	"github.com/spf13/afero"
	mock_ez "github.com/torloejborg/easykube/mock"
	"github.com/torloejborg/easykube/pkg/ez"
	jsutils "github.com/torloejborg/easykube/pkg/js"
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
	ez.Kube.UseContainerRuntime(ez.CreateContainerRuntimeImpl())

	_ = ez.Kube.MakeConfig()

}

func TestPreloadImages(t *testing.T) {
	setup(t)
	ctl := gomock.NewController(t)

	mockCommand := mock_ez.NewMockICobraCommandHelper(ctl)
	mockCommand.EXPECT().IsDryRun().Return(false).AnyTimes()
	mockCommand.EXPECT().GetBoolFlag(gomock.AnyOf("pull")).Return(true).AnyTimes()
	ez.Kube.ICobraCommandHelper = mockCommand

	sec := make(map[string][]byte)
	sec["artifactoryUsername"] = []byte("user")
	sec["artifactoryPassword"] = []byte("ohsosecret")

	k8s := mock_ez.NewMockIK8SUtils(ctl)
	k8s.EXPECT().GetSecret(gomock.Any(), gomock.Any()).Return(sec, nil).AnyTimes()
	ez.Kube.UseK8sUtils(k8s)

	dockerMock := mock_ez.NewMockIContainerRuntime(ctl)
	dockerMock.EXPECT().HasImageInKindRegistry(gomock.Any()).Return(false).AnyTimes()
	dockerMock.EXPECT().PullImage(gomock.Any(), gomock.Any()).AnyTimes()
	dockerMock.EXPECT().TagImage(gomock.Any(), gomock.Any()).AnyTimes()
	dockerMock.EXPECT().PushImage(gomock.Any()).AnyTimes()

	ez.Kube.UseContainerRuntime(dockerMock)

	test.CopyTestAddonToMemFs("../../test_addons", "diamond", "/home/some-user/addons", ez.Kube.Fs)

	script := `
		const images = new Map([
    		["foo:1.0.0", "localhost:5001/foo:1.0.0"],
    		["service.ccta.dk/bar:2.0.0", "localhost:5001/bar:2.0.0"],
		]);

		easykube.preload(images);
	`

	mock := mock_ez.NewMockIAddon(ctl)
	mock.EXPECT().GetName().Return("synthetic").AnyTimes()
	mock.EXPECT().ReadScriptFile(gomock.Any()).Return(script).AnyTimes()
	mock.EXPECT().GetRootDir().Return("/home/some-user/addons").AnyTimes()

	jsu := jsutils.NewJsUtils(mockCommand, mock)
	err := jsu.ExecAddonScript(mock)

	if err != nil {
		t.Fatal(err)
	}

}
