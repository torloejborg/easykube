package cmd

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/ekctx"
	"github.com/torloejborg/easykube/pkg"
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ek"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "creates the easykube cluster",
	Long:  `bootstraps a kind cluster with an opinionated configuration`,
	Run: func(cmd *cobra.Command, args []string) {

		// create a 'toolbox' of needed utilities for cluster creation
		tb := struct {
			cluster   ek.IClusterUtils
			container ek.IContainerRuntime
			addon     ek.IAddonReader
			k8s       ek.IK8SUtils
			cfg       ek.IEasykubeConfig
			tools     ek.IExternalTools
		}{
			pkg.CreateClusterUtils(),
			pkg.CreateContainerRuntime(),
			pkg.CreateAddonReader(),
			pkg.CreateK8sUtils(),
			pkg.CreateEasykubeConfig(),
			pkg.CreateExternalTools(),
		}

		appContext := ekctx.GetAppContext(cmd)
		out := appContext.Printer

		if tb.container.IsContainerRunning(constants.KIND_CONTAINER) {
			out.FmtYellow("Cluster was already created, exiting.")
			os.Exit(0)
		}

		out.FmtGreen("Bootstrapping easykube single node cluster")
		// Ensure configation exists
		tb.cfg.MakeConfig()

		if !tb.container.HasImage(constants.REGISTRY_IMAGE) {
			out.FmtYellow("Pulling docker registry image")
			tb.container.Pull(constants.REGISTRY_IMAGE, nil)
		}

		if !tb.container.HasImage(constants.KIND_IMAGE) {
			out.FmtYellow("Pulling kind image")
			tb.container.Pull(constants.KIND_IMAGE, nil)
		}

		tb.cluster.EnsurePersistenceDirectory()
		tb.container.CreateContainerRegistry()

		u := ek.Utils{pkg.FILESYSTEM}
		occupiedPorts, _ := ensureClusterPortsFree(tb.addon.GetAddons())
		if nil != occupiedPorts {
			out.FmtGreen("Can not create easykube cluster")
			fmt.Println()
			for k, v := range occupiedPorts {
				out.FmtGreen("* %s wants to bind to: 127.0.0.1:[%s]", k.Name, strings.Join(u.IntSliceToStrings(v), ","))
			}
			fmt.Println()
			out.FmtRed("Please halt your local services, or remove the ExtraPorts configuration from the addons listed above ")
			os.Exit(-1)
		}

		report := tb.cluster.CreateKindCluster(tb.addon.GetAddons())
		tb.container.NetworkConnect(constants.REGISTRY_CONTAINER, constants.KIND_NETWORK_NAME)
		tb.k8s.PatchCoreDNS()

		err := tb.k8s.CreateConfigmap(constants.ADDON_CM, "default")
		if err != nil {
			panic(err)
		}

		out.FmtGreen(report)

		// switch to the easykube context
		tb.tools.EnsureLocalContext()

		// ensure secret
		createSecret := appContext.GetStringFlag("secret")
		if len(createSecret) != 0 {

			out.FmtGreen("importing property %s file as secret %s containing:", createSecret, "easykube-secrets")
			fmt.Println()
			configmap, err := u.ReadPropertyFile(createSecret)

			for key := range configmap {
				out.FmtGreen("⚿ %s", key)
			}

			if err != nil {
				out.FmtRed("Did not locate %s\n", createSecret)
				os.Exit(-1)
			}

			tb.k8s.CreateSecret("default", "easykube-secrets", configmap)
		} else {
			out.FmtYellow("Warning, cluster created without importing secrets, this might affect your ability to pull images from private registries.")
		}
	},
}

func ensureClusterPortsFree(addons map[string]*ek.Addon) (map[*ek.Addon][]int, error) {

	IsPortAvailable := func(host string, port int) bool {
		addr := fmt.Sprintf("%s:%d", host, port)
		l, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
		if err != nil {
			return true
		}
		_ = l.Close()
		return false
	}

	failed := make(map[*ek.Addon][]int)

	for _, a := range addons {
		for _, p := range a.Config.ExtraPorts {
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

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringP(constants.ARG_SECRETS, "s", "", "Property file to load as 'easykube-secrets', useful for image pull secrets and other custom configuration")
}
