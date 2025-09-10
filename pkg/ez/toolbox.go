package ez

import (
	"os"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/ekctx"
)

var FILESYSTEM = afero.NewOsFs()

// EkCmdContext Initialized to empty struct, context will be set later when Cobra is initializing (root.go in cmd package)
var EkCmdContext = &ekctx.EKContext{}

type OsDetails interface {
	GetUserConfigDir() (string, error)
	GetUserHomeDir() (string, error)
}

type OsDetailsImpl struct{}

func (d *OsDetailsImpl) GetUserConfigDir() (string, error) {
	r, err := os.UserConfigDir()

	if err != nil {
		return "", err
	}
	return r, nil
}

func (d *OsDetailsImpl) GetUserHomeDir() (string, error) {
	r, err := os.UserHomeDir()

	if err != nil {
		return "", err
	}
	return r, nil
}

type Toolbox struct {
	IK8SUtils
	IEasykubeConfig
	IAddonReader
	IExternalTools
	IContainerRuntime
	IClusterUtils
	ekctx.EKContext
	afero.Fs
	OsDetails
}

var Kube = &Toolbox{}

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

func (t *Toolbox) UseOsDetails(ctx OsDetails) {
	t.OsDetails = ctx
}

func CreateK8sUtilsImpl() IK8SUtils {
	return NewK8SUtils(EkCmdContext, FILESYSTEM)
}

func CreateEasykubeConfigImpl(osd OsDetails) IEasykubeConfig {
	return NewEasykubeConfig(FILESYSTEM, osd)
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
	return NewClusterUtils(EkCmdContext)
}

func CreateOsDetails() OsDetails {
	return &OsDetailsImpl{}
}

func InitializeKubeSingleton(cmd *cobra.Command, ctx ekctx.EKContext) {
	Kube = &Toolbox{}
	// My go-fu is not strong yet, so this feels a bit funky, I want to have some factories
	// produce the utility instances, and they need the EKContext upfront - since the variable is
	// initialized here where the application is bootstrapped, perhaps that's ok. - Send PR's :)
	EkCmdContext = &ctx

	osd := CreateOsDetails()

	Kube.UseOsDetails(osd)
	Kube.UseCmdContext(ctx)
	Kube.UseFilesystemLayer(FILESYSTEM)
	Kube.UseK8sUtils(CreateK8sUtilsImpl())
	Kube.UseEasykubeConfig(CreateEasykubeConfigImpl(osd))
	Kube.UseAddonReader(CreateAddonReaderImpl())
	Kube.UseExternalTools(CreateExternalToolsImpl())
	Kube.UseContainerRuntime(CreateContainerRuntimeImpl())
	Kube.UseClusterUtils(CreateClusterUtilsImpl())
}
