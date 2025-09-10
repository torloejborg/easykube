package ez

import (
	"os"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/ekctx"
)

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
	ekctx.Printer
	afero.Fs
	OsDetails
}

var Kube = &Toolbox{
	Printer: ekctx.Printer{},
}

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
	return NewK8SUtils(EkCmdContext)
}

func CreateEasykubeConfigImpl(osd OsDetails) IEasykubeConfig {
	return NewEasykubeConfig(osd)
}

func CreateAddonReaderImpl(config IEasykubeConfig) IAddonReader {
	return NewAddonReader(config)
}

func CreateExternalToolsImpl() IExternalTools {
	return NewExternalTools(EkCmdContext)
}

func CreateContainerRuntimeImpl() IContainerRuntime {
	return NewContainerRuntime()
}

func CreateClusterUtilsImpl() IClusterUtils {
	return NewClusterUtils(EkCmdContext)
}

func CreateOsDetailsImpl() OsDetails {
	return &OsDetailsImpl{}
}

func InitializeKubeSingleton(cmd *cobra.Command, ctx ekctx.EKContext) {
	Kube = &Toolbox{}
	// My go-fu is not strong yet, so this feels a bit funky, I want to have some factories
	// produce the utility instances, and they need the EKContext upfront - since the variable is
	// initialized here where the application is bootstrapped, perhaps that's ok. - Send PR's :)
	EkCmdContext = &ctx

	osd := CreateOsDetailsImpl()
	config := CreateEasykubeConfigImpl(osd)

	Kube.UseFilesystemLayer(afero.NewOsFs())
	Kube.UseOsDetails(osd)
	Kube.UseCmdContext(ctx)
	Kube.UseK8sUtils(CreateK8sUtilsImpl())
	Kube.UseEasykubeConfig(config)
	Kube.UseAddonReader(CreateAddonReaderImpl(config))
	Kube.UseExternalTools(CreateExternalToolsImpl())
	Kube.UseContainerRuntime(CreateContainerRuntimeImpl())
	Kube.UseClusterUtils(CreateClusterUtilsImpl())
}
