package cmd

import (
	"errors"

	"github.com/spf13/afero"
	"github.com/torloejborg/easykube/pkg/core"
	"github.com/torloejborg/easykube/pkg/ez"
	"github.com/torloejborg/easykube/pkg/textutils"
)

type EasyKubeFactory struct {
	initializeContainerRuntime          bool
	initializeKubernetesClient          bool
	initializeAddonReader               bool
	initializeClusterUtils              bool
	initializeWithMustHaveConfiguration bool
}

type EkOpt func(easykube *EasyKubeFactory) error

func WithContainerRuntime(initialize bool) EkOpt {
	return func(e *EasyKubeFactory) error {
		e.initializeContainerRuntime = initialize
		return nil
	}
}

func WithKubernetes(initialize bool) EkOpt {
	return func(e *EasyKubeFactory) error {
		e.initializeKubernetesClient = initialize
		return nil
	}
}
func WithAddonReader(initialize bool) EkOpt {
	return func(e *EasyKubeFactory) error {
		e.initializeAddonReader = initialize
		return nil
	}
}

func WithClusterUtils(initialize bool) EkOpt {
	return func(e *EasyKubeFactory) error {
		e.initializeClusterUtils = initialize
		return nil
	}
}

func WithRequiresConfigurationCreated(configIsCreated bool) EkOpt {
	return func(e *EasyKubeFactory) error {
		e.initializeWithMustHaveConfiguration = configIsCreated
		return nil
	}
}

func CreateEasykube(cmd *core.CobraCommandHelperImpl, opts ...EkOpt) (ek *core.Ek, err error) {

	ekOpts := EasyKubeFactory{
		initializeContainerRuntime:          true,
		initializeKubernetesClient:          true,
		initializeAddonReader:               true,
		initializeClusterUtils:              true,
		initializeWithMustHaveConfiguration: true,
	}

	for _, opt := range opts {
		if err := opt(&ekOpts); err != nil {
			return nil, err
		}
	}

	ek = &core.Ek{
		Fs:             afero.NewOsFs(),
		CommandContext: cmd,
		Printer:        textutils.NewPrinter(),
	}

	ek.Config = ez.NewEasykubeConfig(ek)
	ek.Utils = ez.NewUtils(ek)
	ek.OsDetails = ez.OsDetailsImpl{Ek: ek}
	ek.ExternalTools = ez.NewExternalTools(ek)
	ek.Status = ez.NewStatusBuilder(ek)

	if ekOpts.initializeWithMustHaveConfiguration {
		_, err := ek.Config.LoadConfig()
		if err != nil {
			return nil, errors.New("failed to load configuration")
		}
	}

	if ekOpts.initializeKubernetesClient {
		ek.Kubernetes = ez.NewK8SUtils(ek)
	}

	if ekOpts.initializeAddonReader {
		ek.AddonReader = ez.NewAddonReader(ek)
	}

	if ekOpts.initializeContainerRuntime {
		cri, err := ez.NewContainerRuntime(ek)
		if err != nil {
			return nil, err
		}
		ek.ContainerRuntime = cri
	}

	if ekOpts.initializeClusterUtils {
		ek.ClusterUtils = ez.NewClusterUtils(ek)
	}

	return ek, nil

}
