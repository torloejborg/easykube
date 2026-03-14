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

type CreateOpts struct {
	Secrets string
}

func createActualCmd(opts CreateOpts) error {

	ezk := ez.Kube
	gr := ez.NewGraph[Task]()

	var clusterCreateReport = ""

	pullImageFunc := func(image string) error {

		if img, err := ezk.HasImage(image); err != nil {
			return err
		} else if !img {

			err := ezk.PullImage(image, nil)
			if err != nil {
				return err
			}

		}

		return nil
	}

	pullRegistryImageTask := NewTaskWithSkip(gr, "pull registry image", func() error {
		return pullImageFunc(constants.REGISTRY_IMAGE)
	}, func() bool {
		has, _ := ezk.HasImage(constants.REGISTRY_IMAGE)
		return has
	})

	pullKindImageTask := NewTaskWithSkip(gr, "pull kind image", func() error {
		return pullImageFunc(constants.KIND_IMAGE)
	}, func() bool {
		has, _ := ezk.HasImage(constants.KIND_IMAGE)
		return has
	})

	inspectPortsFreeTask := NewTaskWithSkip(gr, "check free ports", func() error {

		addons, err := ezk.GetAddons()
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
	}, func() bool { return ezk.IsClusterRunning() })

	connectRegistryToKindTask := NewTaskWithSkip(gr, "connecting registry to kind network", func() error {
		if e := ezk.NetworkConnect(constants.REGISTRY_CONTAINER, constants.KIND_NETWORK_NAME); e != nil {
			return e
		}
		return nil
	}, func() bool {
		connected, _ := ezk.IsNetworkConnectedToContainer(constants.REGISTRY_CONTAINER, constants.KIND_NETWORK_NAME)
		return connected
	})

	createClusterTask := NewTaskWithSkip(gr, "create easykube-kind cluster", func() error {

		addons, err := ezk.GetAddons()
		if err != nil {
			return err
		}

		clusterCreateReport, err = ezk.CreateKindCluster(addons)
		if err != nil {
			return err
		}

		return nil
	}, func() bool { return ezk.IsClusterRunning() })

	createSecretsTask := NewTaskWithSkip(gr, "importing secrets from property file", func() error {
		configmap, err := ez.ReadPropertyFile(opts.Secrets)
		if err != nil {
			return errors.New(fmt.Sprintf("Error reading property file %s, %v", opts.Secrets, err.Error()))
		}
		err = ezk.CreateSecret("default", constants.EASYKUBE_SECRET_NAME, configmap)
		if err != nil {
			return err
		}

		return nil

	}, func() bool {
		return len(opts.Secrets) == 0 && ezk.IsClusterRunning()
	})

	startRegistryTask := NewTaskWithSkip(gr, "start registry", func() error {
		return ez.Kube.StartContainerRegistry()
	}, func() bool {
		running, _ := ezk.IsContainerRunning(constants.REGISTRY_CONTAINER)
		return running
	})

	createRegistryTask := NewTaskWithSkip(gr, "create local container registry", func() error {
		err := ez.Kube.CreateContainerRegistry()
		if err != nil {
			return err
		}
		return nil
	}, func() bool { // if already running, return
		search, _ := ezk.FindContainer(constants.REGISTRY_CONTAINER)
		return search.Found
	})

	patchCoreDNSTask := NewTaskWithSkip(gr, "patch coreDNS", func() error {
		ezk.PatchCoreDNS()
		return nil
	}, func() bool { return ezk.IsClusterRunning() })

	ensureAddonConfigMapTask := NewTaskWithSkip(gr, "ensure addon config map", func() error {
		if err := ezk.CreateConfigmap(constants.ADDON_CM, constants.DEFAULT_NS); err != nil {
			return err
		}
		return nil
	}, func() bool {
		_, err := ezk.ReadConfigmap(constants.ADDON_CM, constants.DEFAULT_NS)
		return err == nil
	})

	ensureLocalClusterContextTask := NewTask(gr, "ensure local cluster context", func() error {
		err := ezk.ReloadClientSet()
		if err != nil {
			return err
		}
		return nil
	})

	ensurePersistenceDirectoriesTask := NewTaskWithSkip(gr, "ensure persistence directories", func() error {
		pdErr := ez.Kube.EnsurePersistenceDirectory()
		if pdErr != nil {
			return pdErr
		}
		return nil
	}, func() bool {
		return ezk.IsClusterRunning()
	})

	ensureContainerRuntimeTask := NewTaskWithSkip(gr, "ensure container runtime", func() error {
		return errors.New("container runtime not available check docker/podman started")
	}, func() bool {
		return ezk.IsContainerRuntimeAvailable()
	})

	gr.AppendNode(ensureContainerRuntimeTask)
	gr.AppendNode(inspectPortsFreeTask)
	gr.AppendNode(pullKindImageTask)
	gr.AppendNode(pullRegistryImageTask)
	gr.AppendNode(createRegistryTask)
	gr.AppendNode(startRegistryTask)
	gr.AppendNode(ensurePersistenceDirectoriesTask)
	gr.AppendNode(createClusterTask)
	gr.AppendNode(connectRegistryToKindTask)
	gr.AppendNode(ensureLocalClusterContextTask)
	gr.AppendNode(patchCoreDNSTask)
	gr.AppendNode(createSecretsTask)
	gr.AppendNode(ensureAddonConfigMapTask)

	res := gr.Nodes

	ExecuteTasks(res)

	if clusterCreateReport != "" {
		fmt.Println(clusterCreateReport)
	}

	return nil
}
