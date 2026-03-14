package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"
)

func destroyActual() error {
	ezk := ez.Kube
	tasks := ez.NewGraph[Task]()

	user, _ := ezk.OsDetails.GetUserConfigDir()
	configDir := filepath.Join(user, constants.CONFIG_DIR_NAME)

	running := func(name string) bool {
		result, _ := ezk.IsContainerRunning(name)
		return result
	}

	stopAndDeleteCluster := NewTaskWithSkip(tasks, fmt.Sprintf("stop and delete %s", constants.KIND_CONTAINER), func() error {
		return stopAndDeleteContainer(constants.KIND_CONTAINER)
	}, func() bool {
		return !running(constants.KIND_CONTAINER)
	})

	stopAndDeleteRegistry := NewTaskWithSkip(tasks, fmt.Sprintf("stop and delete %s", constants.REGISTRY_CONTAINER), func() error {
		return stopAndDeleteContainer(constants.REGISTRY_CONTAINER)
	}, func() bool {
		return !running(constants.REGISTRY_CONTAINER)
	})

	purgeData := NewTaskWithSkip(tasks, "purge data", func() error {

		s, _ := ezk.Fs.Stat(configDir)
		if s.IsDir() {
			return ezk.Fs.RemoveAll(configDir)
		}

		return nil

	}, func() bool {
		return !ezk.GetBoolFlag("purge")
	})

	tasks.AppendNode(stopAndDeleteCluster)
	tasks.AppendNode(stopAndDeleteRegistry)
	tasks.AppendNode(purgeData)

	nodes := tasks.Nodes

	ExecuteTasks(nodes)

	return nil
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
