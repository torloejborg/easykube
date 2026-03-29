package core

import "github.com/spf13/afero"

type Ek struct {
	Config           IEasykubeConfig
	AddonReader      IAddonReader
	ContainerRuntime IContainerRuntime
	CommandContext   ICobraCommandHelper
	ExternalTools    IExternalTools
	OsDetails        IOsDetails
	Printer          IPrinter
	Kubernetes       IK8SUtils
	Utils            IUtils
	Fs               afero.Fs
}
