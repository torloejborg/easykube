package cmd

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"
)

type BootOpts struct {
	Secrets string
}

func createActualCmd(opts BootOpts) error {

	tasks := NewTaskContainer()

	tasks.AddTask(ensureContainerRuntimeTask())
	tasks.AddTask(inspectPortsFreeTask())
	tasks.AddTask(pullKindImageTask())
	tasks.AddTask(pullRegistryImageTask())
	tasks.AddTask(createRegistryTask())
	tasks.AddTask(startRegistryTask())
	tasks.AddTask(ensurePersistenceDirectoriesTask())
	tasks.AddTask(createClusterTask())
	tasks.AddTask(connectRegistryToKindTask())
	tasks.AddTask(ensureLocalClusterContextTask())
	tasks.AddTask(patchCoreDNSTask())
	tasks.AddTask(ensureAddonConfigMapTask())

	ExecuteTasks(tasks)

	if clusterCreateReport != "" {
		fmt.Println(clusterCreateReport)
	}

	return nil
}

func ensureContainerRuntimeTask() Task {
	return NewTaskWithSkip("ensure container runtime", func() error {
		return errors.New("container runtime not available check docker/podman started")
	}, func() bool {
		return ez.Kube.IsContainerRuntimeAvailable()
	})
}

func inspectPortsFreeTask() Task {
	return NewTaskWithSkip("check free ports", func() error {

		addons, err := ez.Kube.GetAddons()
		if err != nil {
			return err
		}

		IsPortAvailable := func(host string, port int) bool {
			addr := fmt.Sprintf("%s:%d", host, port)
			l, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
			if err != nil {
				return true
			}
			_ = l.Close()
			return false
		}

		failed := make(map[ez.IAddon][]int)

		for _, a := range addons {
			for _, p := range a.GetConfig().ExtraPorts {
				if !IsPortAvailable("127.0.0.1", p.HostPort) {
					failed[a] = append(failed[a], p.HostPort)
				}
			}
		}

		if len(failed) != 0 {

			errorList := make([]string, 0)

			for k, v := range failed {
				errorList = append(errorList, fmt.Sprintf("%s->%d", k.GetName(), v))
			}

			return errors.New("some ports are not available: " + strings.Join(errorList, ","))
		}
		return nil
	}, func() bool { return ez.Kube.IsClusterRunning() })

}

func pullKindImageTask() Task {
	return NewTaskWithSkip("pull kind image", func() error {
		return pullImageFunc(constants.KIND_IMAGE)
	}, func() bool {
		has, _ := ez.Kube.HasImage(constants.KIND_IMAGE)
		return has
	})
}

func pullRegistryImageTask() Task {

	return NewTaskWithSkip("pull registry image", func() error {
		return pullImageFunc(constants.REGISTRY_IMAGE)
	}, func() bool {
		has, _ := ez.Kube.HasImage(constants.REGISTRY_IMAGE)
		return has
	})
}

func pullImageFunc(image string) error {

	if img, err := ez.Kube.HasImage(image); err != nil {
		return err
	} else if !img {

		err := ez.Kube.PullImage(image, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func connectRegistryToKindTask() Task {
	return NewTaskWithSkip("connecting registry to kind network", func() error {
		if e := ez.Kube.NetworkConnect(constants.REGISTRY_CONTAINER, constants.KIND_NETWORK_NAME); e != nil {
			return e
		}
		return nil
	}, func() bool {
		connected, _ := ez.Kube.IsNetworkConnectedToContainer(constants.REGISTRY_CONTAINER, constants.KIND_NETWORK_NAME)
		return connected
	})
}

var clusterCreateReport = ""

func createClusterTask() Task {
	return NewTaskWithSkip("create easykube-kind cluster", func() error {

		addons, err := ez.Kube.GetAddons()
		if err != nil {
			return err
		}

		clusterCreateReport, err = ez.Kube.CreateKindCluster(addons)
		if err != nil {
			return err
		}

		return nil
	}, func() bool { return ez.Kube.IsClusterRunning() })
}

func startRegistryTask() Task {
	return NewTaskWithSkip("start registry", func() error {
		return ez.Kube.StartContainerRegistry()
	}, func() bool {
		running, _ := ez.Kube.IsContainerRunning(constants.REGISTRY_CONTAINER)
		return running
	})
}

func createRegistryTask() Task {
	return NewTaskWithSkip("create local container registry", func() error {
		err := ez.Kube.CreateContainerRegistry()
		if err != nil {
			return err
		}
		return nil
	}, func() bool { // if already running, return
		search, _ := ez.Kube.FindContainer(constants.REGISTRY_CONTAINER)
		return search.Found
	})
}

func patchCoreDNSTask() Task {
	return NewTaskWithSkip("patch coreDNS", func() error {
		ez.Kube.PatchCoreDNS()
		return nil
	}, func() bool { return ez.Kube.IsClusterRunning() })
}

func ensureAddonConfigMapTask() Task {

	return NewTaskWithSkip("ensure addon config map", func() error {
		if err := ez.Kube.CreateConfigmap(constants.ADDON_CM, constants.DEFAULT_NS); err != nil {
			return err
		}
		return nil
	}, func() bool {
		_, err := ez.Kube.ReadConfigmap(constants.ADDON_CM, constants.DEFAULT_NS)
		return err == nil
	})
}

func ensureLocalClusterContextTask() Task {
	return NewTask("ensure local cluster context", func() error {
		err := ez.Kube.ReloadClientSet()
		if err != nil {
			return err
		}
		return nil
	})

}

func ensurePersistenceDirectoriesTask() Task {
	return NewTaskWithSkip("ensure persistence directories", func() error {
		pdErr := ez.Kube.EnsurePersistenceDirectory()
		if pdErr != nil {
			return pdErr
		}
		return nil
	}, func() bool {
		return ez.Kube.IsClusterRunning()
	})

}
