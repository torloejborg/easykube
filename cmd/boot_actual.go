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

	tasks := NewTaskContainer()

	tasks.AddTask(inspectConfigurationPresent(currentConfig))
	tasks.AddTask(ensureContainerRuntimeTask())
	tasks.AddTask(inspectPortsFreeTask())
	tasks.AddTask(pullKindImageTask())
	tasks.AddTask(pullRegistryImageTask())
	tasks.AddTask(zotRegistryConfigurationTask())
	tasks.AddTask(zotRegistryCredentialConfigurationTask())
	tasks.AddTask(createRegistryTask())
	tasks.AddTask(startRegistryTask())
	tasks.AddTask(createClusterTask())
	tasks.AddTask(ensurePersistenceDirectoriesTask())
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
		return pullImageFunc(constants.KindImage)
	}, func() bool {
		has, _ := ez.Kube.HasImage(constants.KindImage)
		return has
	})
}

func pullRegistryImageTask() Task {

	return NewTaskWithSkip("pull registry image", func() error {
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

func connectRegistryToKindTask() Task {
	return NewTaskWithSkip("connect registry to kind network", func() error {
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
		running, _ := ez.Kube.IsContainerRunning(constants.RegistryContainer)
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
		search, _ := ez.Kube.FindContainer(constants.RegistryContainer)
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
		if err := ez.Kube.CreateConfigmap(constants.AddonCm, constants.DefaultNs); err != nil {
			return err
		}
		return nil
	}, func() bool {
		_, err := ez.Kube.ReadConfigmap(constants.AddonCm, constants.DefaultNs)
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

func zotRegistryConfigurationTask() Task {

	return NewTaskWithSkip("update zot mirror registries", func() error {

		cfg, err := ez.Kube.LoadConfig()
		if err != nil {
			return err
		}

		err = ez.Kube.GenerateZotRegistryConfig(cfg)
		if err != nil {
			return err
		}

		return nil
	}, func() bool {
		// todo: find a good skip condition
		return false
	})
}

func zotRegistryCredentialConfigurationTask() Task {
	return NewTaskWithSkip("update zot mirror credentials", func() error {

		cfg, err := ez.Kube.LoadConfig()
		if err != nil {
			return err
		}

		err = ez.Kube.GenerateZotRegistryCredentials(cfg)
		if err != nil {
			return err
		}

		return nil
	}, func() bool {
		// todo: find a good skip condition
		return false
	})
}

func inspectConfigurationPresent(curr *ez.EasykubeConfigData) Task {
	return NewTaskWithSkip("inspect configuration", func() error {

		_, err := ez.Kube.LoadConfig()
		if err != nil {
			return errors.New("configuration not found, please run `easykube config | config --use-defaults`")
		}

		return nil
	}, func() bool {

		return curr != nil
	})
}
