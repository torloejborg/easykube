package cmd

import (
	"fmt"
	"time"

	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/core"
)

func destroyActual(ek *core.Ek) error {

	tasks := core.NewTaskContainer()

	tasks.AddTask(stopAndDeleteCluster(ek))
	tasks.AddTask(stopAndDeleteRegistry(ek))
	tasks.AddTask(purgeData(ek))

	core.ExecuteTasks(tasks)

	return nil
}

func running(name string, ek *core.Ek) bool {
	result, _ := ek.ContainerRuntime.IsContainerRunning(name)
	return result
}

func stopAndDeleteCluster(ek *core.Ek) core.Task {
	return core.NewTaskWithSkip(fmt.Sprintf("stop and delete %s", constants.KindContainer), func() error {
		return stopAndDeleteContainer(constants.KindContainer, ek)
	}, func() bool {
		return !running(constants.KindContainer, ek)
	})
}

func stopAndDeleteRegistry(ek *core.Ek) core.Task {
	return core.NewTaskWithSkip(fmt.Sprintf("stop and delete %s", constants.RegistryContainer), func() error {
		return stopAndDeleteContainer(constants.RegistryContainer, ek)
	}, func() bool {
		return !running(constants.RegistryContainer, ek)
	})
}

func purgeData(ek *core.Ek) core.Task {
	return core.NewTaskWithSkip("purge data", func() error {

		configDir, err := ek.OsDetails.GetEasykubeConfigDir()
		if err != nil {
			return err
		}

		s, _ := ek.Fs.Stat(configDir)
		if s.IsDir() {
			return ek.Fs.RemoveAll(configDir)
		}

		return nil

	}, func() bool {
		return !ek.CommandContext.GetBoolFlag("purge")
	})

}

func stopAndDeleteContainer(name string, ek *core.Ek) error {

	find := func(name string) (*core.ContainerSearch, error) {
		if search, err := ek.ContainerRuntime.FindContainer(name); err != nil {
			return &core.ContainerSearch{
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

			stopErr := ek.ContainerRuntime.StopContainer(s.ContainerID)
			if stopErr != nil {
				return stopErr
			}
		} else {
			err := ek.ContainerRuntime.RemoveContainer(s.ContainerID)
			if err != nil {
				return err
			}
			return nil
		}

		time.Sleep(5 * time.Second)
	}

	return nil
}
