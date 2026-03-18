package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"
)

func destroyActual() error {

	tasks := NewTaskContainer()

	tasks.AddTask(stopAndDeleteCluster())
	tasks.AddTask(stopAndDeleteRegistry())
	tasks.AddTask(purgeData())

	ExecuteTasks(tasks)

	return nil
}

func running(name string) bool {
	result, _ := ez.Kube.IsContainerRunning(name)
	return result
}

func stopAndDeleteCluster() Task {
	return NewTaskWithSkip(fmt.Sprintf("stop and delete %s", constants.KindContainer), func() error {
		return stopAndDeleteContainer(constants.KindContainer)
	}, func() bool {
		return !running(constants.KindContainer)
	})
}

func stopAndDeleteRegistry() Task {
	return NewTaskWithSkip(fmt.Sprintf("stop and delete %s", constants.RegistryContainer), func() error {
		return stopAndDeleteContainer(constants.RegistryContainer)
	}, func() bool {
		return !running(constants.RegistryContainer)
	})
}

func purgeData() Task {
	return NewTaskWithSkip("purge data", func() error {

		user, _ := ez.Kube.OsDetails.GetUserConfigDir()
		configDir := filepath.Join(user, constants.ConfigDirName)

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
