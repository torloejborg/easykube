package ez

import "github.com/torloejborg/easykube/pkg/core"

func NewContainerRuntime(ek *core.Ek) (core.IContainerRuntime, error) {

	cfg, err := ek.Config.LoadConfig()
	if err != nil {
		panic(err)
	}

	return NewContainerRuntimeImpl(ek, cfg.ContainerRuntime)
}
