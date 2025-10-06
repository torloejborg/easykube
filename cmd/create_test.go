package cmd

import (
	"testing"

	"github.com/spf13/afero"
	mock_ez "github.com/torloejborg/easykube/mock"
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"
	"github.com/torloejborg/easykube/pkg/textutils"
	"github.com/torloejborg/easykube/test"
	"go.uber.org/mock/gomock"
)

func setupMockForCreate(ctrl *gomock.Controller) {

	// set up user details
	osd := mock_ez.NewMockOsDetails(ctrl)
	osd.EXPECT().GetUserConfigDir().Return("/home/some-user/.config", nil).AnyTimes()
	osd.EXPECT().GetUserHomeDir().Return("/home/some-user", nil).AnyTimes()

	// configures interaction with kubernetes
	mk8s := mock_ez.NewMockIK8SUtils(ctrl)
	mk8s.EXPECT().UpdateConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mk8s.EXPECT().CreateConfigmap(constants.ADDON_CM, constants.DEFAULT_NS).Return(nil).AnyTimes()
	mk8s.EXPECT().PatchCoreDNS()
	mk8s.EXPECT().ReloadClientSet()
	mk8s.EXPECT().CreateSecret("default", "easykube-secrets", map[string]string{
		"key":        "somekey",
		"value":      "somevalue",
		"anotherkey": "anothervalue",
	})

	// mocks interactions with docker
	containerRuntime := mock_ez.NewMockIContainerRuntime(ctrl)
	containerRuntime.EXPECT().IsClusterRunning().Return(false).AnyTimes()
	containerRuntime.EXPECT().IsContainerRunning(gomock.Any()).Return(false)
	containerRuntime.EXPECT().HasImage(constants.REGISTRY_IMAGE).Return(false)
	containerRuntime.EXPECT().HasImage(constants.KIND_IMAGE).Return(false)
	containerRuntime.EXPECT().PullImage(constants.REGISTRY_IMAGE, gomock.Any())
	containerRuntime.EXPECT().PullImage(constants.KIND_IMAGE, gomock.Any())
	containerRuntime.EXPECT().CreateContainerRegistry()
	containerRuntime.EXPECT().NetworkConnect(constants.REGISTRY_CONTAINER, constants.KIND_NETWORK_NAME)

	// not creating
	clusterUtils := mock_ez.NewMockIClusterUtils(ctrl)
	clusterUtils.EXPECT().EnsurePersistenceDirectory()
	clusterUtils.EXPECT().CreateKindCluster(gomock.Any())

	commandHelper := mock_ez.NewMockICobraCommandHelper(ctrl)
	config := ez.NewEasykubeConfig(osd)

	ez.Kube.ICobraCommandHelper = commandHelper
	ez.Kube.OsDetails = osd
	ez.Kube.IEasykubeConfig = config
	ez.Kube.Fs = afero.NewMemMapFs()
	ez.Kube.IK8SUtils = mk8s
	ez.Kube.IClusterUtils = clusterUtils
	ez.Kube.IContainerRuntime = containerRuntime
	ez.Kube.IAddonReader = ez.NewAddonReader(config)
	ez.Kube.IExternalTools = ez.NewExternalTools()
	_ = ez.Kube.MakeConfig()
}

func TestCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	setupMockForCreate(ctrl)

	props := textutils.TrimMargin(
		`
		|key=somekey
		|value=somevalue
		|
		|anotherkey=anothervalue
		`, "|")

	_ = afero.WriteFile(ez.Kube.Fs, "/home/some-user/prop.properties", []byte(props), 0644)

	// load some addons
	test.CopyTestAddonToMemFs("../test_addons", "exec", "/home/some-user/addons", ez.Kube.Fs)
	_ = test.PrintFiles(ez.Kube.Fs, "/")

	opts := CreateOpts{
		Secrets: "/home/some-user/prop.properties",
	}

	err := createActualCmd(opts, ez.Kube.ICobraCommandHelper)

	if err != nil {
		t.Error(err)
	}
}
