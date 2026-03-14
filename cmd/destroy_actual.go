package cmd

import (
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"
)

func destroyActual() error {
	ezk := ez.Kube
	tasks := &ez.Graph[Task]{}

	tasks.AppendNode(NewTaskWithSkip(tasks, "stop registry container", func() error {
		return nil
	}, func() bool {
		return false
	}))

	tasks.AppendNode(NewTaskWithSkip(tasks, "delete registry container", func() error {
		return nil
	}, func() bool {
		return false
	}))

	tasks.AppendNode(NewTaskWithSkip(tasks, "stop easykube cluster", func() error {
		return nil
	}, func() bool {
		return false
	}))

	tasks.AppendNode(NewTaskWithSkip(tasks, "delete easykube cluster", func() error {
		return nil
	}, func() bool {
		return false
	}))

	tasks.AppendNode(NewTaskWithSkip(tasks, "purge data", func() error {
		return nil
	}, func() bool {
		return false
	}))

	ExecuteTasks(tasks.Nodes)

	err := stopAndDeleteContainer(constants.REGISTRY_CONTAINER)
	if nil != err {
		return err
	}

	err = stopAndDeleteContainer(constants.KIND_CONTAINER)
	if nil != err {
		return err
	}

	if ezk.GetBoolFlag("purge") {
		cfg, err := ezk.LoadConfig()
		if err != nil {
			return err
		}

		err = ezk.Fs.RemoveAll(cfg.ConfigurationDir)
		if err != nil {
			ezk.FmtYellow("could not remove configuration dir: %s", cfg.ConfigurationDir)
		}
	}

	return nil
}

func stopAndDeleteContainer(name string) error {
	ezk := ez.Kube
	if search, err := ezk.FindContainer(name); err != nil {
		return err
	} else if search.Found {
		if search.IsRunning {
			if _, err := ezk.FmtSpinner(func() (any, error) {
				return nil, ezk.StopContainer(search.ContainerID)
			}, "Stopping %s", name); err != nil {
				return err
			}
		}
		if _, err := ezk.FmtSpinner(func() (any, error) {
			return nil, ezk.RemoveContainer(search.ContainerID)
		}, "Removing %s", name); err != nil {
			return err
		}
	}

	return nil
}
