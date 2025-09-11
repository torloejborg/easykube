package ez

import (
	"os"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/ekctx"
)

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

type EasykubeSingleton struct {
	IK8SUtils
	IEasykubeConfig
	IAddonReader
	IExternalTools
	IContainerRuntime
	IClusterUtils
	OsDetails
	afero.Fs
	*cobra.Command
	CobraCommandHelperImpl
	ekctx.IPrinter
}

var Kube = &EasykubeSingleton{
	IPrinter: ekctx.NewPrinter(),
}

func (t *EasykubeSingleton) UseK8sUtils(newUtils IK8SUtils) *EasykubeSingleton {
	t.IK8SUtils = newUtils
	return t
}

func (t *EasykubeSingleton) UseEasykubeConfig(c IEasykubeConfig) *EasykubeSingleton {
	t.IEasykubeConfig = c
	return t
}

func (t *EasykubeSingleton) UseAddonReader(r IAddonReader) *EasykubeSingleton {
	t.IAddonReader = r
	return t
}

func (t *EasykubeSingleton) UseExternalTools(e IExternalTools) *EasykubeSingleton {
	t.IExternalTools = e
	return t
}

func (t *EasykubeSingleton) UseContainerRuntime(r IContainerRuntime) *EasykubeSingleton {
	t.IContainerRuntime = r
	return t
}

func (t *EasykubeSingleton) UseClusterUtils(u IClusterUtils) *EasykubeSingleton {
	t.IClusterUtils = u
	return t
}

func (t *EasykubeSingleton) UseFilesystemLayer(f afero.Fs) *EasykubeSingleton {
	t.Fs = f
	return t
}

func (t *EasykubeSingleton) UseCmdContext(ctx CobraCommandHelperImpl) {
	t.CobraCommandHelperImpl = ctx
}

func (t *EasykubeSingleton) UseOsDetails(ctx OsDetails) {
	t.OsDetails = ctx
}

func (t *EasykubeSingleton) UsePrinter(printer ekctx.IPrinter) {
	t.IPrinter = printer
}

//func (t *EasykubeSingleton) UseCobraCommand(cmd *cobra.Command) {
//	t.Command = cmd
//}

func CreateK8sUtilsImpl() IK8SUtils {
	return NewK8SUtils()
}

func CreateEasykubeConfigImpl(osd OsDetails) IEasykubeConfig {
	return NewEasykubeConfig(osd)
}

func CreateAddonReaderImpl(config IEasykubeConfig) IAddonReader {
	return NewAddonReader(config)
}

func CreateExternalToolsImpl() IExternalTools {
	return NewExternalTools()
}

func CreateContainerRuntimeImpl() IContainerRuntime {
	return NewContainerRuntime()
}

func CreateClusterUtilsImpl() IClusterUtils {
	return NewClusterUtils()
}

func CreateOsDetailsImpl() OsDetails {
	return &OsDetailsImpl{}
}

func InitializeKubeSingleton() {

	// I'm damaged by Java, We could inject anything anywhere, now this is my attempt at destructuring
	// the application into smaller parts and assembling it with an initialization function. This allows
	// parts or aspects of the application to be configured differently for tests.

	osd := CreateOsDetailsImpl()
	config := CreateEasykubeConfigImpl(osd)

	Kube.UsePrinter(ekctx.NewPrinter())
	Kube.UseFilesystemLayer(afero.NewOsFs())
	Kube.UseOsDetails(osd)
	Kube.UseK8sUtils(CreateK8sUtilsImpl())
	Kube.UseEasykubeConfig(config)
	Kube.UseAddonReader(CreateAddonReaderImpl(config))
	Kube.UseExternalTools(CreateExternalToolsImpl())
	Kube.UseContainerRuntime(CreateContainerRuntimeImpl())
	Kube.UseClusterUtils(CreateClusterUtilsImpl())
}
