package cmd

import (
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"
)

func destroyActual() error {

	ezk := ez.Kube
	search, err := ezk.FindContainer(constants.KIND_CONTAINER)
	if err != nil {
		return err
	}

	if search.Found {
		ezk.FmtYellow("Stopping %s", constants.KIND_CONTAINER)
		if search.IsRunning {
			ezk.StopContainer(search.ContainerID)
		}
		ezk.RemoveContainer(search.ContainerID)
		ezk.FmtYellow("Removing %s", constants.KIND_CONTAINER)
	}

	return nil
}
