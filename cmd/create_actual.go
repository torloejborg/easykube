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

func createActualCmd(opts CreateOpts, cmdHelper ez.ICobraCommandHelper) error {
	ezk := ez.Kube

	if ezk.IsContainerRunning(constants.KIND_CONTAINER) {
		return errors.New("cluster already running")
	}

	ezk.FmtGreen("Bootstrapping easykube single node cluster")
	// Ensure configation exists
	err := ezk.MakeConfig()
	if err != nil {
		return err
	}

	if !ezk.HasImage(constants.REGISTRY_IMAGE) {
		ezk.FmtYellow("Pulling docker registry image")
		ezk.PullImage(constants.REGISTRY_IMAGE, nil)
	}

	if !ezk.HasImage(constants.KIND_IMAGE) {
		ezk.FmtYellow("Pulling kind image")
		ezk.PullImage(constants.KIND_IMAGE, nil)
	}

	pdErr := ez.Kube.EnsurePersistenceDirectory()
	if pdErr != nil {
		return pdErr
	}

	regerr := ez.Kube.CreateContainerRegistry()
	if regerr != nil {
		return regerr
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

	report := ezk.CreateKindCluster(addons)

	// The cluster is created, and so a new context will exist, tell the k8sutils to
	// create a new ClientSet, so we can bootstrap
	cerr := ezk.ReloadClientSet()
	if cerr != nil {
		return cerr
	}

	ezk.NetworkConnect(constants.REGISTRY_CONTAINER, constants.KIND_NETWORK_NAME)
	ezk.PatchCoreDNS()

	err = ezk.CreateConfigmap(constants.ADDON_CM, constants.DEFAULT_NS)
	if err != nil {
		return err
	}

	ezk.FmtGreen(report)

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

		ezk.CreateSecret("default", constants.EASYKUBE_SECRET_NAME, configmap)
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
