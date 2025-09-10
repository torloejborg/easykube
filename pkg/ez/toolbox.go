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
	afero.Fs
}

var Kube *Toolbox = &Toolbox{}

func (t *Toolbox) UseK8sUtils(newUtils IK8SUtils) *Toolbox {
	t.IK8SUtils = newUtils
	return t
}

func (t *Toolbox) UseEasykubeConfig(c IEasykubeConfig) *Toolbox {
	t.IEasykubeConfig = c
	return t
}

func (t *Toolbox) UseAddonReader(r IAddonReader) *Toolbox {
	t.IAddonReader = r
	return t
}

func (t *Toolbox) UseExternalTools(e IExternalTools) *Toolbox {
	t.IExternalTools = e
	return t
}

func (t *Toolbox) UseContainerRuntime(r IContainerRuntime) *Toolbox {
	t.IContainerRuntime = r
	return t
}

func (t *Toolbox) UseClusterUtils(u IClusterUtils) *Toolbox {
	t.IClusterUtils = u
	return t
}

func (t *Toolbox) UseFilesystemLayer(f afero.Fs) *Toolbox {
	t.Fs = f
	return t
}

func (t *Toolbox) UseCmdContext(ctx ekctx.EKContext) {
	t.EKContext = ctx
}

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
