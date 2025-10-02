package ez

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/textutils"
)

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
	textutils.IPrinter
}

var Kube = &EasykubeSingleton{
	IPrinter: textutils.NewPrinter(),
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

func (t *EasykubeSingleton) UsePrinter(printer textutils.IPrinter) {
	t.IPrinter = printer
}

func CreateK8sUtilsImpl() IK8SUtils {
	return NewK8SUtils()
}

func CreateEasykubeConfigImpl(osd OsDetails) IEasykubeConfig {
	return NewEasykubeConfig(osd)
}

func CreateAddonReaderImpl(config IEasykubeConfig) IAddonReader {
	return NewAddonReader(config)
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

	// I'm damaged by Java, there we could inject anything anywhere, now this is my attempt at destructuring
	// the application into smaller parts and assembling it with an initialization function. This allows
	// parts or aspects of the application to be configured differently for tests.

	osd := CreateOsDetailsImpl()
	config := CreateEasykubeConfigImpl(osd)

	Kube.UseFilesystemLayer(afero.NewOsFs())
	Kube.UsePrinter(textutils.NewPrinter())
	Kube.UseOsDetails(osd)
	Kube.UseK8sUtils(CreateK8sUtilsImpl())
	Kube.UseEasykubeConfig(config)
	Kube.UseAddonReader(CreateAddonReaderImpl(config))
	Kube.UseExternalTools(NewExternalTools())
	Kube.UseContainerRuntime(CreateContainerRuntimeImpl())
	Kube.UseClusterUtils(CreateClusterUtilsImpl())
}
