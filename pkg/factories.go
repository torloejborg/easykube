package pkg

import (
	"github.com/spf13/afero"
	"github.com/torloejborg/easykube/ekctx"
	"github.com/torloejborg/easykube/pkg/ek"
)

var FILESYSTEM = afero.NewOsFs()

// EkCmdContext Initialized to empty struct, context will be set later when Cobra is initializing (root.go in cmd package)
var EkCmdContext = &ekctx.EKContext{}

func CreateK8sUtils() ek.IK8SUtils {
	return ek.NewK8SUtils(EkCmdContext, FILESYSTEM)
}

func CreateEasykubeConfig() ek.IEasykubeConfig {
	return ek.NewEasykubeConfig(EkCmdContext, FILESYSTEM)
}

func CreateAddonReader() ek.IAddonReader {
	return ek.NewAddonReader(EkCmdContext, FILESYSTEM)
}

func CreateExternalTools() ek.IExternalTools {
	return ek.NewExternalTools(EkCmdContext)
}

func CreateContainerRuntime() ek.IContainerRuntime {
	return ek.NewContainerRuntime(EkCmdContext, FILESYSTEM)
}

func CreateClusterUtils() ek.IClusterUtils {
	return ek.NewClusterUtils(EkCmdContext, FILESYSTEM)
}
