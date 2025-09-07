package test

import (
	"log"

	"github.com/spf13/afero"
	"github.com/torloejborg/easykube/ekctx"
	"github.com/torloejborg/easykube/pkg/ez"
)

var FILESYSTEM = afero.NewMemMapFs()

// EkCmdContext Initialized to empty struct, context will be set later when Cobra is initializing (root.go in cmd package)
var EkCmdContext = &ekctx.EKContext{
	Logger:  log.Default(),
	Printer: &ekctx.Printer{},
	Command: nil,
	Fs:      FILESYSTEM,
}

func CreateFakeK8sUtil() ez.IK8SUtils {
	return ez.NewK8SUtils(EkCmdContext, FILESYSTEM)
}

func CreateFakeEasykubeConfig() ez.IEasykubeConfig {
	return ez.NewEasykubeConfig(EkCmdContext, FILESYSTEM)
}

func CreateFakeAddonReader() ez.IAddonReader {
	return ez.NewAddonReader(EkCmdContext, FILESYSTEM)
}

func CreateFakeExternalTools() ez.IExternalTools {
	return ez.NewExternalTools(EkCmdContext)
}

func CreateFakeContainerRuntime() ez.IContainerRuntime {
	return ez.NewContainerRuntime(EkCmdContext, FILESYSTEM)
}

func CreateFakeClusterUtils() ez.IClusterUtils {
	return ez.NewClusterUtils(EkCmdContext, FILESYSTEM)
}
