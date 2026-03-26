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

var skipRestartRegistryTask bool = true

func createActualCmd(opts BootOpts, currentConfig *ez.EasykubeConfigData) error {

	tasks := ez.NewTaskContainer()

	tasks.AddTask(ensureContainerRuntimeTask())
	tasks.AddTask(inspectPortsFreeTask())
	tasks.AddTask(pullKindImageTask())
	tasks.AddTask(pullRegistryImageTask())
	tasks.AddTask(configureZotRegistry(currentConfig))
	tasks.AddTask(createRegistryTask())
	tasks.AddTask(restartRegistryTask())
	tasks.AddTask(startRegistryTask())
	tasks.AddTask(createClusterTask())
	tasks.AddTask(ensurePersistenceDirectoriesTask())
	tasks.AddTask(connectRegistryToKindTask())
	tasks.AddTask(ensureLocalClusterContextTask())
	tasks.AddTask(patchCoreDNSTask())
	tasks.AddTask(ensureAddonConfigMapTask())

	ez.ExecuteTasks(tasks)

	if clusterCreateReport != "" {
		fmt.Println(clusterCreateReport)
	}

	return nil
}

func ensureContainerRuntimeTask() ez.Task {
	return ez.NewTaskWithSkip("ensure container runtime", func() error {
		return errors.New("container runtime not available check docker/podman started")
	}, func() bool {
		return ez.Kube.IsContainerRuntimeAvailable()
	})
}

func inspectPortsFreeTask() ez.Task {
	return ez.NewTaskWithSkip("check free ports", func() error {

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

func pullKindImageTask() ez.Task {
	return ez.NewTaskWithSkip(fmt.Sprintf("pull kind image: %s", constants.KindImage), func() error {
		return pullImageFunc(constants.KindImage)
	}, func() bool {
		has, _ := ez.Kube.HasImage(constants.KindImage)
		return has
	})
}

func pullRegistryImageTask() ez.Task {

	return ez.NewTaskWithSkip(fmt.Sprintf("pull registry image: %s", constants.RegistryImage), func() error {
		return pullImageFunc(constants.RegistryImage)
	}, func() bool {
		has, _ := ez.Kube.HasImage(constants.RegistryImage)
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

func connectRegistryToKindTask() ez.Task {
	return ez.NewTaskWithSkip("connect registry to kind network", func() error {
		if e := ez.Kube.NetworkConnect(constants.RegistryContainer, constants.KindNetworkName); e != nil {
			return e
		}
		return nil
	}, func() bool {
		connected, _ := ez.Kube.IsNetworkConnectedToContainer(constants.RegistryContainer, constants.KindNetworkName)
		return connected
	})
}

var clusterCreateReport = ""

func createClusterTask() ez.Task {
	return ez.NewTaskWithSkip("create easykube-kind cluster", func() error {
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

func startRegistryTask() ez.Task {
	return ez.NewTaskWithSkip("start zot-registry", func() error {
		return ez.Kube.StartContainerRegistry()
	}, func() bool {
		running, _ := ez.Kube.IsContainerRunning(constants.RegistryContainer)
		return running
	})
}

func restartRegistryTask() ez.Task {
	return ez.NewTaskWithSkip("restart zot-registry", func() error {
		err := ez.Kube.StopContainer(constants.RegistryContainer)
		if err != nil {
			return err
		}
		return ez.Kube.StartContainer(constants.RegistryContainer)
	}, func() bool {
		return skipRestartRegistryTask
	})
}

func createRegistryTask() ez.Task {
	return ez.NewTaskWithSkip("create local container registry", func() error {
		err := ez.Kube.CreateContainerRegistry()
		if err != nil {
			return err
		}
		return nil
	}, func() bool { // if already running, return
		search, _ := ez.Kube.FindContainer(constants.RegistryContainer)
		return search.Found
	})
}

func patchCoreDNSTask() ez.Task {
	return ez.NewTaskWithSkip("patch coreDNS", func() error {
		ez.Kube.PatchCoreDNS()
		ez.Kube.IK8SUtils.RestartDeployment("coredns", "kube-system")
		return nil
	}, func() bool {

		if ez.Kube.IsClusterRunning() {
			cm, _ := ez.Kube.ReadConfigmap("coredns", "kube-system")
			if len(cm) <= 1 {
				return false
			}
		}

		return ez.Kube.IsClusterRunning()
	})
}

func ensureAddonConfigMapTask() ez.Task {

	return ez.NewTaskWithSkip("ensure addon config map", func() error {
		if err := ez.Kube.CreateConfigmap(constants.AddonCm, constants.DefaultNs); err != nil {
			return err
		}
		return nil
	}, func() bool {
		_, err := ez.Kube.ReadConfigmap(constants.AddonCm, constants.DefaultNs)
		return err == nil
	})
}

func ensureLocalClusterContextTask() ez.Task {
	return ez.NewTask("ensure local cluster context", func() error {
		err := ez.Kube.ReloadClientSet()
		if err != nil {
			return err
		}
		return nil
	})

}

func ensurePersistenceDirectoriesTask() ez.Task {
	return ez.NewTaskWithSkip("ensure persistence directories", func() error {
		pdErr := ez.Kube.EnsurePersistenceDirectory()
		if pdErr != nil {
			return pdErr
		}
		return nil
	}, func() bool {
		return ez.Kube.IsClusterRunning()
	})

}

func configureZotRegistry(config *ez.EasykubeConfigData) ez.Task {

	return ez.NewTaskWithSkip("re-configure zot registry", func() error {

		err := ez.Kube.GenerateZotRegistryConfig(config)
		if err != nil {
			panic(err)
		}

		err = ez.Kube.GenerateZotRegistryCredentials(config)
		if err != nil {
			panic(err)
		}

		return nil
	}, func() bool {

		sync, err := ez.Kube.IsZotConfigInSync(config)
		skipRestartRegistryTask = sync
		if err != nil {
			panic(err)
		}

		return sync
	})
}
