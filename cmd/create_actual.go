package cmd

import (
	"errors"
	"fmt"
	"net"
	"os"
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

	s, err := ezk.FindContainer(constants.KIND_CONTAINER)
	if err != nil {
		return err
	}

	if s.Found && !s.IsRunning {
		return fmt.Errorf("cluster container %s exists but is not running", constants.KIND_CONTAINER)
	} else if s.Found && s.IsRunning {
		return fmt.Errorf("cluster container %s is already running", constants.KIND_CONTAINER)
	}

	ezk.FmtGreen("Bootstrapping easykube single node cluster")
	// Ensure configuration exists
	if err := ezk.MakeConfig(); err != nil {
		return err
	}

	if img, err := ezk.HasImage(constants.REGISTRY_IMAGE); err != nil {
		return err
	} else if !img {
		if _, err := ezk.FmtSpinner(func() (any, error) {
			return nil, ezk.PullImage(constants.REGISTRY_IMAGE, nil)
		}, "Pulling registry image %s", constants.REGISTRY_IMAGE); err != nil {
			return err
		}
	}

	if img, err := ezk.HasImage(constants.KIND_IMAGE); err != nil {
		return err
	} else if !img {
		if _, err := ezk.FmtSpinner(func() (any, error) {
			return nil, ezk.PullImage(constants.KIND_IMAGE, nil)
		}, "Pulling kind image %s", constants.KIND_IMAGE); err != nil {
			return err
		}
	}

	pdErr := ez.Kube.EnsurePersistenceDirectory()
	if pdErr != nil {
		return pdErr
	}

	if _, err := ez.Kube.FmtSpinner(func() (any, error) {
		return nil, ez.Kube.CreateContainerRegistry()
	}, "Ensure container registry"); err != nil {
		return err
	}

	addons, aerr := ez.Kube.GetAddons()
	if aerr != nil {
		return aerr
	}

	occupiedPorts, _ := ensureClusterPortsFree(addons)
	if nil != occupiedPorts {
		ezk.FmtYellow("Can not create easykube cluster")
		fmt.Println()
		for k, v := range occupiedPorts {
			ez.Kube.FmtGreen("* %s wants to bind to: 127.0.0.1:[%s]", k.GetName(), strings.Join(ez.IntSliceToStrings(v), ","))
		}
		fmt.Println()
		ezk.FmtRed("Please halt your local services, or remove the ExtraPorts configuration from the addons listed above ")
		os.Exit(-1)
	}

	report, _ := ezk.FmtSpinner(func() (any, error) {
		r := ezk.CreateKindCluster(addons)
		return r, nil
	}, "Creating kind-easykube control plane")

	// The cluster is created, and so a new context will exist, tell the k8sUtils to
	// create a new ClientSet, so we can continue bootstrapping
	cerr := ezk.ReloadClientSet()
	if cerr != nil {
		return cerr
	}

	if connected, _ := ezk.IsNetworkConnectedToContainer(constants.REGISTRY_CONTAINER, constants.KIND_NETWORK_NAME); !connected {
		if e := ezk.NetworkConnect(constants.REGISTRY_CONTAINER, constants.KIND_NETWORK_NAME); e != nil {
			return e
		}
	}

	_, _ = ezk.FmtSpinner(func() (any, error) {
		ezk.PatchCoreDNS()
		return nil, nil
	}, "Patching coredns")

	if err := ezk.CreateConfigmap(constants.ADDON_CM, constants.DEFAULT_NS); err != nil {
		return err
	}

	ezk.FmtGreen(report.(string))

	// switch to the easykube context
	ezk.EnsureLocalContext()

	// ensure secret
	if len(opts.Secrets) != 0 {

		ezk.FmtGreen("importing property %s file as secret %s containing:", opts.Secrets, constants.EASYKUBE_SECRET_NAME)
		fmt.Println()
		configmap, err := ez.ReadPropertyFile(opts.Secrets)

		for key := range configmap {
			ezk.FmtGreen("⚿ %s", key)
		}

		if err != nil {
			return errors.New(fmt.Sprintf("Error reading property file %s, %v", opts.Secrets, err.Error()))
		}

		_ = ezk.CreateSecret("default", constants.EASYKUBE_SECRET_NAME, configmap)
	} else {
		ezk.FmtYellow("Warning, cluster created without importing secrets, this might affect your ability to pull images from private registries.")
	}

	return nil
}

func ensureClusterPortsFree(addons map[string]ez.IAddon) (map[ez.IAddon][]int, error) {

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
		return failed, errors.New("some ports are not available")
	} else {
		return nil, nil
	}
}
