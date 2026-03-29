package core

import (
	"github.com/spf13/afero"
)

type Ek struct {
	AddonReader      IAddonReader
	ClusterUtils     IClusterUtils
	CommandContext   ICobraCommandHelper
	Config           IEasykubeConfig
	ContainerRuntime IContainerRuntime
	ExternalTools    IExternalTools
	Fs               afero.Fs
	Kubernetes       IK8SUtils
	OsDetails        IOsDetails
	Printer          IPrinter
	Utils            IUtils
	Status           IStatusBuilder
}
