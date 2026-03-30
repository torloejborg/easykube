package cmd

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/core"
)

type BootOpts struct {
	Secrets string
}

var skipRestartRegistryTask bool = true

func createActualCmd(ek *core.Ek, currentConfig *core.EasykubeConfigData) error {

	tasks := core.NewTaskContainer()

	tasks.AddTask(ensureContainerRuntimeTask(ek))
	tasks.AddTask(inspectPortsFreeTask(ek))
	tasks.AddTask(pullKindImageTask(ek))
	tasks.AddTask(pullRegistryImageTask(ek))
	tasks.AddTask(ensurePersistenceDirectoriesTask(ek))
	tasks.AddTask(createClusterTask(ek))
	tasks.AddTask(createRegistryTask(ek))
	tasks.AddTask(configureZotRegistry(currentConfig, ek))
	tasks.AddTask(connectRegistryToKindTask(ek))
	tasks.AddTask(startRegistryTask(ek))
	tasks.AddTask(restartRegistryTask(ek))
	tasks.AddTask(ensureLocalClusterContextTask(ek))
	tasks.AddTask(patchCoreDNSTask(ek))
	tasks.AddTask(ensureAddonConfigMapTask(ek))

	core.ExecuteTasks(tasks)

	if clusterCreateReport != "" {
		fmt.Println(clusterCreateReport)
	}

	return nil
}

func ensureContainerRuntimeTask(ek *core.Ek) core.Task {
	return core.NewTaskWithSkip("check container runtime", func() error {
		return errors.New("container runtime not available check docker/podman started")
	}, func() bool {
		return ek.ContainerRuntime.IsContainerRuntimeAvailable()
	})
}

func inspectPortsFreeTask(ek *core.Ek) core.Task {
	return core.NewTaskWithSkip("check free ports", func() error {

		addons, err := ek.AddonReader.GetAddons()
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

		failed := make(map[core.IAddon][]int)

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
	}, func() bool { return ek.ContainerRuntime.IsClusterRunning() })

}

func pullKindImageTask(ek *core.Ek) core.Task {
	return core.NewTaskWithSkip(fmt.Sprintf("pull kind image: %s", constants.KindImage), func() error {
		return pullImageFunc(constants.KindImage, ek)
	}, func() bool {
		has, _ := ek.ContainerRuntime.HasImage(constants.KindImage)
		return has
	})
}

func pullRegistryImageTask(ek *core.Ek) core.Task {

	return core.NewTaskWithSkip(fmt.Sprintf("pull registry image: %s", constants.RegistryImage), func() error {
		return pullImageFunc(constants.RegistryImage, ek)
	}, func() bool {
		has, _ := ek.ContainerRuntime.HasImage(constants.RegistryImage)
		return has
	})
}

func pullImageFunc(image string, ek *core.Ek) error {

	if img, err := ek.ContainerRuntime.HasImage(image); err != nil {
		return err
	} else if !img {

		err := ek.ContainerRuntime.PullImage(image, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func connectRegistryToKindTask(ek *core.Ek) core.Task {
	return core.NewTaskWithSkip("connect registry to kind network", func() error {
		if e := ek.ContainerRuntime.NetworkConnect(constants.RegistryContainer, constants.KindNetworkName); e != nil {
			return e
		}
		return nil
	}, func() bool {
		connected, _ := ek.ContainerRuntime.IsNetworkConnectedToContainer(constants.RegistryContainer, constants.KindNetworkName)
		return connected
	})
}

var clusterCreateReport = ""

func createClusterTask(ek *core.Ek) core.Task {
	return core.NewTaskWithSkip("create easykube-kind cluster", func() error {
		addons, err := ek.AddonReader.GetAddons()
		if err != nil {
			return err
		}

		clusterCreateReport, err = ek.ClusterUtils.CreateKindCluster(addons)
		if err != nil {
			return err
		}

		return nil
	}, func() bool { return ek.ContainerRuntime.IsClusterRunning() })
}

func startRegistryTask(ek *core.Ek) core.Task {
	return core.NewTaskWithSkip("start zot-registry", func() error {
		return ek.ContainerRuntime.StartContainerRegistry()
	}, func() bool {
		running, _ := ek.ContainerRuntime.IsContainerRunning(constants.RegistryContainer)
		return running
	})
}

func restartRegistryTask(ek *core.Ek) core.Task {
	return core.NewTaskWithSkip("restart zot-registry", func() error {
		err := ek.ContainerRuntime.StopContainer(constants.RegistryContainer)
		if err != nil {
			return err
		}
		return ek.ContainerRuntime.StartContainer(constants.RegistryContainer)
	}, func() bool {
		return skipRestartRegistryTask
	})
}

func createRegistryTask(ek *core.Ek) core.Task {
	return core.NewTaskWithSkip("create local container registry", func() error {
		err := ek.ContainerRuntime.CreateContainerRegistry()
		if err != nil {
			return err
		}
		return nil
	}, func() bool { // if already running, return
		search, _ := ek.ContainerRuntime.FindContainer(constants.RegistryContainer)
		return search.Found
	})
}

func patchCoreDNSTask(ek *core.Ek) core.Task {
	return core.NewTaskWithSkip("patch coreDNS", func() error {
		ek.Kubernetes.PatchCoreDNS()
		_ = ek.Kubernetes.RestartDeployment("coredns", "kube-system")
		return nil
	}, func() bool {

		if ek.ContainerRuntime.IsClusterRunning() {
			cm, _ := ek.Kubernetes.ReadConfigmap("coredns", "kube-system")
			if len(cm) <= 1 {
				return false
			}
		}

		return ek.ContainerRuntime.IsClusterRunning()
	})
}

func ensureAddonConfigMapTask(ek *core.Ek) core.Task {

	return core.NewTaskWithSkip("ensure addon config map", func() error {
		if err := ek.Kubernetes.CreateConfigmap(constants.AddonCm, constants.DefaultNs); err != nil {
			return err
		}
		return nil
	}, func() bool {
		_, err := ek.Kubernetes.ReadConfigmap(constants.AddonCm, constants.DefaultNs)
		return err == nil
	})
}

func ensureLocalClusterContextTask(ek *core.Ek) core.Task {
	return core.NewTask("ensure local cluster context", func() error {
		err := ek.Kubernetes.ReloadClientSet()
		if err != nil {
			return err
		}
		return nil
	})

}

func ensurePersistenceDirectoriesTask(ek *core.Ek) core.Task {
	return core.NewTaskWithSkip("ensure persistence directories", func() error {
		pdErr := ek.ClusterUtils.EnsurePersistenceDirectory()
		if pdErr != nil {
			return pdErr
		}
		return nil
	}, func() bool {
		return ek.ContainerRuntime.IsClusterRunning()
	})

}

func configureZotRegistry(config *core.EasykubeConfigData, ek *core.Ek) core.Task {

	return core.NewTaskWithSkip("re-configure zot registry", func() error {

		err := ek.Config.GenerateZotRegistryConfig(config)
		if err != nil {
			panic(err)
		}

		err = ek.Config.GenerateZotRegistryCredentials(config)
		if err != nil {
			panic(err)
		}

		return nil
	}, func() bool {

		sync, err := ek.Config.IsZotConfigInSync(config)
		skipRestartRegistryTask = sync
		if err != nil {
			panic(err)
		}

		return sync
	})
}
