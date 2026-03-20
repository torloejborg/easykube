package cmd

import (
	"fmt"
	"time"

	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"
)

func destroyActual() error {

	tasks := ez.NewTaskContainer()

	tasks.AddTask(stopAndDeleteCluster())
	tasks.AddTask(stopAndDeleteRegistry())
	tasks.AddTask(purgeData())

	ez.ExecuteTasks(tasks)

	return nil
}

func running(name string) bool {
	result, _ := ez.Kube.IsContainerRunning(name)
	return result
}

func stopAndDeleteCluster() ez.Task {
	return ez.NewTaskWithSkip(fmt.Sprintf("stop and delete %s", constants.KindContainer), func() error {
		return stopAndDeleteContainer(constants.KindContainer)
	}, func() bool {
		return !running(constants.KindContainer)
	})
}

func stopAndDeleteRegistry() ez.Task {
	return ez.NewTaskWithSkip(fmt.Sprintf("stop and delete %s", constants.RegistryContainer), func() error {
		return stopAndDeleteContainer(constants.RegistryContainer)
	}, func() bool {
		return !running(constants.RegistryContainer)
	})
}

func purgeData() ez.Task {
	return ez.NewTaskWithSkip("purge data", func() error {

		configDir, err := ez.Kube.GetEasykubeConfigDir()
		if err != nil {
			return err
		}

		s, _ := ez.Kube.Fs.Stat(configDir)
		if s.IsDir() {
			return ez.Kube.Fs.RemoveAll(configDir)
		}

		return nil

	}, func() bool {
		return !ez.Kube.GetBoolFlag("purge")
	})

}

func stopAndDeleteContainer(name string) error {
	ezk := ez.Kube

	find := func(name string) (*ez.ContainerSearch, error) {
		if search, err := ezk.FindContainer(name); err != nil {
			return &ez.ContainerSearch{
				ContainerID: "",
				Found:       false,
				IsRunning:   false,
			}, err
		} else if search.Found {
			return search, nil
		}

		return nil, fmt.Errorf("container %s not found", name)
	}

	for i := 0; i < 10; i++ {

		s, e := find(name)
		if e != nil {
			return e
		}

		if s.IsRunning {

			stopErr := ezk.StopContainer(s.ContainerID)
			if stopErr != nil {
				return stopErr
			}
		} else {
			err := ezk.RemoveContainer(s.ContainerID)
			if err != nil {
				return err
			}
			return nil
		}

		time.Sleep(5 * time.Second)
	}

	return nil
}
