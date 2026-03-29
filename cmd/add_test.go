package cmd

import (
	"os"
	"testing"

	"github.com/spf13/afero"
	mock "github.com/torloejborg/easykube/mock"
	"github.com/torloejborg/easykube/pkg/core"
	"github.com/torloejborg/easykube/pkg/ez"
	"github.com/torloejborg/easykube/pkg/textutils"
	"github.com/torloejborg/easykube/test"
	"go.uber.org/mock/gomock"
)

func setupForDryRun(ctrl *gomock.Controller, t *testing.T) *core.Ek {
	_ = os.Setenv("KUBECONFIG", "mock-kubeconfig")

	osd := mock.NewMockIOsDetails(ctrl)
	osd.EXPECT().GetEasykubeConfigDir().Return("/home/some-user/easykube/.config", nil).AnyTimes()
	osd.EXPECT().GetUserHomeDir().Return("/home/some-user", nil).AnyTimes()

	mk8s := mock.NewMockIK8SUtils(ctrl)
	mk8s.EXPECT().GetInstalledAddons().Return([]string{""}, nil).AnyTimes()

	containerRuntime := mock.NewMockIContainerRuntime(ctrl)
	containerRuntime.EXPECT().IsClusterRunning().Return(true).AnyTimes()

	clusterUtils := mock.NewMockIClusterUtils(ctrl)

	commandHelper := mock.NewMockICobraCommandHelper(ctrl)
	commandHelper.EXPECT().IsDryRun().Return(true).AnyTimes()
	commandHelper.EXPECT().IsVerbose().Return(true).AnyTimes()

	ek := &core.Ek{
		OsDetails:        osd,
		ContainerRuntime: containerRuntime,
		ClusterUtils:     clusterUtils,
		CommandContext:   commandHelper,
		Kubernetes:       mk8s,
		Fs:               afero.NewMemMapFs(),
		Printer:          textutils.NewPrinter(),
	}
	ek.Utils = ez.NewUtils(ek)
	ek.Config = ez.NewEasykubeConfig(ek)
	ek.ExternalTools = ez.NewExternalTools(ek)
	_ = ek.Config.MakeConfig()

	ek.AddonReader = ez.NewAddonReader(ek)

	return ek
}

func TestAdd(t *testing.T) {
	ctrl := gomock.NewController(t)
	ek := setupForDryRun(ctrl, t)

	addOpts := AddOptions{
		Args:          []string{"a", "b"},
		ForceInstall:  false,
		TargetCluster: "",
		NoDepends:     false,
		DryRun:        true,
	}

	// load some addons
	test.CopyTestAddonToMemFs("../test_addons", "exec", "/home/some-user/addons", ek.Fs)
	_ = test.PrintFiles(ek.Fs, "/")

	err := addActual(addOpts, ek)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

}
