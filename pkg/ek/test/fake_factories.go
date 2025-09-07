package test

import (
	"log"

	"github.com/spf13/afero"
	"github.com/torloejborg/easykube/ekctx"
	"github.com/torloejborg/easykube/pkg/ek"
)

var FILESYSTEM = afero.NewMemMapFs()

// EkCmdContext Initialized to empty struct, context will be set later when Cobra is initializing (root.go in cmd package)
var EkCmdContext = &ekctx.EKContext{
	Logger:  log.Default(),
	Printer: &ekctx.Printer{},
	Command: nil,
	Fs:      FILESYSTEM,
}

func CreateFakeK8sUtil() ek.IK8SUtils {
	return ek.NewK8SUtils(EkCmdContext, FILESYSTEM)
}

func CreateFakeEasykubeConfig() ek.IEasykubeConfig {
	return ek.NewEasykubeConfig(EkCmdContext, FILESYSTEM)
}

func CreateFakeAddonReader() ek.IAddonReader {
	return ek.NewAddonReader(EkCmdContext, FILESYSTEM)
}

func CreateFakeExternalTools() ek.IExternalTools {
	return ek.NewExternalTools(EkCmdContext)
}

func CreateFakeContainerRuntime() ek.IContainerRuntime {
	return ek.NewContainerRuntime(EkCmdContext, FILESYSTEM)
}

func CreateFakeClusterUtils() ek.IClusterUtils {
	return ek.NewClusterUtils(EkCmdContext, FILESYSTEM)
}
