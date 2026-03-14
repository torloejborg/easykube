package cmd

import (
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"
)

/*
1 .stop control plane
2. delete registry
3. delete config
4. delete persistence
*/

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

	if ezk.GetBoolFlag("purge") {

		if search, err := ezk.FindContainer(constants.REGISTRY_CONTAINER); err != nil {
			return err
		} else if search.Found {
			if search.IsRunning {
				if _, err := ezk.FmtSpinner(func() (any, error) {
					return nil, ezk.StopContainer(search.ContainerID)
				}, "Stopping %s", constants.REGISTRY_CONTAINER); err != nil {
					return err
				}
			}
			if _, err := ezk.FmtSpinner(func() (any, error) {
				return nil, ezk.RemoveContainer(search.ContainerID)
			}, "Removing %s", constants.KIND_CONTAINER); err != nil {
				return err
			}
		}
		cfg, e := ezk.LoadConfig()
		if e != nil {
			return e
		}

		e = ezk.Fs.RemoveAll(cfg.ConfigurationDir)
		if e != nil {
			ezk.FmtYellow("could not remove configuration dir: %s", cfg.ConfigurationDir)
		}
	}

	return nil
}
