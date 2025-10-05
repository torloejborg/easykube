package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/spf13/afero"
	mock_ez "github.com/torloejborg/easykube/mock"
	"github.com/torloejborg/easykube/pkg/ez"
	"github.com/torloejborg/easykube/test"
	"go.uber.org/mock/gomock"
)

func initAddTest(ctrl *gomock.Controller) {
	osd := mock_ez.NewMockOsDetails(ctrl)

	osd.EXPECT().GetUserConfigDir().Return("/home/some-user/.config", nil).AnyTimes()
	osd.EXPECT().GetUserHomeDir().Return("/home/some-user", nil).AnyTimes()

	mk8s := mock_ez.NewMockIK8SUtils(ctrl)
	//mk8s.EXPECT().GetInstalledAddons().Return([]string{""}, nil)

	containerRuntime := mock_ez.NewMockIContainerRuntime(ctrl)
	//containerRuntime.EXPECT().IsClusterRunning().Return(true).AnyTimes()

	clusterUtils := mock_ez.NewMockIClusterUtils(ctrl)

	config := ez.NewEasykubeConfig(osd)
	ez.Kube.OsDetails = osd
	ez.Kube.IEasykubeConfig = config
	ez.Kube.Fs = afero.NewMemMapFs()
	ez.Kube.IK8SUtils = mk8s
	ez.Kube.IClusterUtils = clusterUtils
	ez.Kube.IContainerRuntime = containerRuntime
	ez.Kube.IAddonReader = ez.NewAddonReader(config)
	ez.Kube.IExternalTools = ez.NewExternalTools()
	
	ez.Kube.MakeConfig()

}

func TestAdd(t *testing.T) {
	ctrl := gomock.NewController(t)
	initAddTest(ctrl)
	//commandHelper := mock_ez.NewMockICobraCommandHelper(ctrl)
	//
	//addOpts := AddOptions{
	//	Args:          []string{"a", "b"},
	//	ForceInstall:  false,
	//	TargetCluster: "",
	//	NoDepends:     false,
	//	DryRun:        true,
	//}

	// load some addons
	_ = test.CopyTestAddonToMemFs("../test_addons/diamond", "./addons", ez.Kube.Fs)

	//err := addActual(addOpts, commandHelper)
	//if err != nil {
	//	t.Errorf("Unexpected error: %s", err)
	//}

	afero.Walk(ez.Kube.Fs, "/", func(path string, info os.FileInfo, err error) error {
		fmt.Println(path)
		return nil
	})

}
