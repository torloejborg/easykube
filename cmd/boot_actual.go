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

func createActualCmd(opts BootOpts, currentConfig *ez.EasykubeConfigData) error {

	tasks := ez.NewTaskContainer()

	tasks.AddTask(inspectConfigurationPresent(currentConfig))
	tasks.AddTask(ensureContainerRuntimeTask())
	tasks.AddTask(inspectPortsFreeTask())
	tasks.AddTask(pullKindImageTask())
	tasks.AddTask(pullRegistryImageTask())
	tasks.AddTask(configureZotRegistry(currentConfig))
	tasks.AddTask(createRegistryTask())
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
	return ez.NewTaskWithSkip("pull kind image", func() error {
		return pullImageFunc(constants.KindImage)
	}, func() bool {
		has, _ := ez.Kube.HasImage(constants.KindImage)
		return has
	})
}

func pullRegistryImageTask() ez.Task {

	return ez.NewTaskWithSkip("pull registry image", func() error {
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
	return ez.NewTaskWithSkip("start registry", func() error {
		return ez.Kube.StartContainerRegistry()
	}, func() bool {
		running, _ := ez.Kube.IsContainerRunning(constants.RegistryContainer)
		return running
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
		return nil
	}, func() bool { return ez.Kube.IsClusterRunning() })
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

func configureZotRegistry(cfg *ez.EasykubeConfigData) ez.Task {

	return ez.NewTaskWithSkip("configure zot", func() error {
		err := ez.Kube.GenerateZotRegistryConfig(cfg)
		if err != nil {
			panic(err)
		}

		err = ez.Kube.GenerateZotRegistryCredentials(cfg)
		if err != nil {
			panic(err)
		}

		return nil
	}, func() bool {
		update, err := ez.Kube.ShouldRegenerateZotConfig(cfg)
		if err != nil {
			panic(err)
		}
		return update
	})
}

func inspectConfigurationPresent(curr *ez.EasykubeConfigData) ez.Task {
	return ez.NewTaskWithSkip("inspect configuration", func() error {

		_, err := ez.Kube.LoadConfig()
		if err != nil {
			return errors.New("configuration not found, please run `easykube config | config --use-defaults`")
		}

		return nil
	}, func() bool {

		return curr != nil
	})
}
