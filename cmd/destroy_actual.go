package cmd

import (
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"
)

func destroyActual() error {

	ezk := ez.Kube
	if search, err := ezk.FindContainer(constants.KIND_CONTAINER); err != nil {
		return err
	} else if search.Found {
		ezk.FmtGreen("Stopping %s", constants.KIND_CONTAINER)
		if search.IsRunning {
			if err := ezk.StopContainer(search.ContainerID); err != nil {
				return err
			}
		}
		if err := ezk.RemoveContainer(search.ContainerID); err != nil {
			return err
		}
		ezk.FmtGreen("Removing %s", constants.KIND_CONTAINER)
	}

	return nil
}
