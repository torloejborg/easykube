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

		if search.IsRunning {

			if _, err := ezk.FmtSpinner(func() (any, error) {
				return nil, ezk.StopContainer(search.ContainerID)
			}, "Stopping %s", constants.KIND_CONTAINER); err != nil {
				return err
			}
		}

		if _, err := ezk.FmtSpinner(func() (any, error) {
			return nil, ezk.RemoveContainer(search.ContainerID)
		}, "Removing %s", constants.KIND_CONTAINER); err != nil {
			return err
		}
	}

	return nil
}
