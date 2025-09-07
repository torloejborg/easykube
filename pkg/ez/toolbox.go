package ez

import (
	"github.com/spf13/afero"
	"github.com/torloejborg/easykube/ekctx"
)

var FILESYSTEM = afero.NewOsFs()

// EkCmdContext Initialized to empty struct, context will be set later when Cobra is initializing (root.go in cmd package)
var EkCmdContext = &ekctx.EKContext{}

type Toolbox struct {
	IK8SUtils
	IEasykubeConfig
	IAddonReader
	IExternalTools
	IContainerRuntime
	IClusterUtils
	ekctx.EKContext
}

func (t *Toolbox) InitK8s(newUtils IK8SUtils) {
	t.IK8SUtils = newUtils
}

func (t *Toolbox) InitCmdContext(ctx ekctx.EKContext) {
	t.EKContext = ctx
}

var Kube = Toolbox{}

func CreateK8sUtilsImpl() IK8SUtils {
	return NewK8SUtils(EkCmdContext, FILESYSTEM)
}

func CreateEasykubeConfigImpl() IEasykubeConfig {
	return NewEasykubeConfig(FILESYSTEM)
}

func CreateAddonReaderImpl() IAddonReader {
	return NewAddonReader(EkCmdContext, FILESYSTEM)
}

func CreateExternalToolsImpl() IExternalTools {
	return NewExternalTools(EkCmdContext)
}

func CreateContainerRuntimeImpl() IContainerRuntime {
	return NewContainerRuntime(FILESYSTEM)
}

func CreateClusterUtilsImpl() IClusterUtils {
	return NewClusterUtils(EkCmdContext, FILESYSTEM)
}
