package jsutils_test

import (
	"testing"

	mock "github.com/torloejborg/easykube/mock"
	jsutils "github.com/torloejborg/easykube/pkg/js"
	"github.com/torloejborg/easykube/test"
	"go.uber.org/mock/gomock"
)

func TestPreloadImages(t *testing.T) {
	ctl := gomock.NewController(t)
	ek := SetUpJsTestEnvironment(t, ctl)

	mockCommand := mock.NewMockICobraCommandHelper(ctl)
	mockCommand.EXPECT().IsDryRun().Return(false).AnyTimes()
	mockCommand.EXPECT().GetBoolFlag(gomock.AnyOf("pull")).Return(true).AnyTimes()

	ek.CommandContext = mockCommand

	sec := make(map[string][]byte)
	sec["artifactoryUsername"] = []byte("user")
	sec["artifactoryPassword"] = []byte("ohsosecret")

	k8s := mock.NewMockIK8SUtils(ctl)
	k8s.EXPECT().GetSecret(gomock.Any(), gomock.Any()).Return(sec, nil).AnyTimes()
	ek.Kubernetes = k8s

	dockerMock := mock.NewMockIContainerRuntime(ctl)
	dockerMock.EXPECT().HasImageInKindRegistry(gomock.Any()).Return(false, nil).AnyTimes()
	dockerMock.EXPECT().PullImage(gomock.Any(), gomock.Any()).AnyTimes()
	dockerMock.EXPECT().TagImage(gomock.Any(), gomock.Any()).AnyTimes()
	dockerMock.EXPECT().PushImage(gomock.Any(), gomock.Any()).AnyTimes()

	ek.ContainerRuntime = dockerMock

	test.CopyTestAddonToMemFs("../../test_addons", "diamond", "/home/some-user/addons", ek.Fs)

	script := `
		const images = new Map([
    		["foo:1.0.0", "localhost:5001/foo:1.0.0"],
    		["service.ccta.dk/bar:2.0.0", "localhost:5001/bar:2.0.0"],
		]);

		easykube.preload(images);
	`

	mockAddon := CreateSyntheticAddon(script, ctl)
	jsu := jsutils.NewJsUtils(ek, mockAddon, false)
	err := jsu.ExecAddonScript(mockAddon)

	if err != nil {
		t.Fatal(err)
	}

}
