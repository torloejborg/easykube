package cmd

import (
	"testing"

	"github.com/spf13/afero"
	mock_ez "github.com/torloejborg/easykube/mock"
	"github.com/torloejborg/easykube/pkg/ez"
	"go.uber.org/mock/gomock"
)

func initAddTest(ctrl *gomock.Controller) {

	fakeAddons := make(map[string]*ez.Addon)
	fakeAddons["foo"] = &ez.Addon{
		Name:      "foo addon",
		ShortName: "foo",
		Config:    ez.AddonConfig{},
		File:      "",
		RootDir:   "",
	}

	addonReader := mock_ez.NewMockIAddonReader(ctrl)
	addonReader.EXPECT().GetAddons().Return(fakeAddons, nil)

	tools := mock_ez.NewMockIExternalTools(ctrl)

	tools.EXPECT().EnsureLocalContext().Do(func() {})

	mk8s := mock_ez.NewMockIK8SUtils(ctrl)
	mk8s.EXPECT().GetInstalledAddons().Return([]string{""}, nil)

	containerRuntime := mock_ez.NewMockIContainerRuntime(ctrl)
	containerRuntime.EXPECT().IsClusterRunning().Return(true).AnyTimes()

	clusterUtils := mock_ez.NewMockIClusterUtils(ctrl)

	ez.Kube.Fs = afero.NewMemMapFs()
	ez.Kube.IK8SUtils = mk8s
	ez.Kube.IClusterUtils = clusterUtils
	ez.Kube.IContainerRuntime = containerRuntime
	ez.Kube.IAddonReader = addonReader
	ez.Kube.IExternalTools = tools

}

func TestAdd(t *testing.T) {
	ctrl := gomock.NewController(t)
	initAddTest(ctrl)
	commandHelper := mock_ez.NewMockICobraCommandHelper(ctrl)

	addOpts := AddOptions{
		Args:          []string{"foo"},
		ForceInstall:  false,
		TargetCluster: "",
		NoDepends:     false,
		DryRun:        false,
	}

	err := addActual(addOpts, commandHelper)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

}
