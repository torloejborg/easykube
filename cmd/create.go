package cmd

import (
	"fmt"
	"github.com/torloj/easykube/ekctx"
	"os"

	"github.com/spf13/cobra"
	"github.com/torloj/easykube/pkg/constants"
	"github.com/torloj/easykube/pkg/ek"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "creates the easykube cluster",
	Long:  `bootstraps a kind cluster with an opinionated configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		appContext := ekctx.GetAppContext(cmd)
		out := appContext.Printer

		cru := ek.NewContainerRuntime(appContext)
		if cru.IsContainerRunning(constants.KIND_CONTAINER) {
			out.FmtYellow("Cluster was already created, exiting.")
			os.Exit(0)
		}

		out.FmtGreen("Bootstrapping easykube single node cluster")
		// Ensure configation exists
		ek.NewEasykubeConfig(appContext).MakeConfig()

		ct := ek.NewContainerRuntime(appContext)
		cu := ek.NewClusterUtils(appContext)

		if !ct.HasImage(constants.REGISTRY_IMAGE) {
			out.FmtYellow("Pulling docker registry image")
			ct.Pull(constants.REGISTRY_IMAGE, nil)
		}

		if !ct.HasImage(constants.KIND_IMAGE) {
			out.FmtYellow("Pulling kind image")
			ct.Pull(constants.KIND_IMAGE, nil)
		}

		cu.EnsurePersistenceDirectory()

		addons := ek.NewAddonReader(appContext).GetAddons()
		ct.CreateContainerRegistry()
		report := cu.CreateKindCluster(addons)

		ct.NetworkConnect(constants.REGISTRY_CONTAINER, constants.KIND_NETWORK_NAME)

		k8sutils := ek.NewK8SUtils(appContext)
		k8sutils.PatchCoreDNS()

		err := k8sutils.CreateConfigmap(constants.ADDON_CM, "default")
		if err != nil {
			panic(err)
		}

		out.FmtGreen(report)

		// switch to the easykube context
		ek.NewExternalTools(appContext).EnsureLocalContext()

		// ensure secret
		createSecret := appContext.GetStringFlag("secret")
		if len(createSecret) != 0 {

			out.FmtGreen("importing property %s file as secret %s containing:", createSecret, "easykube-secrets")
			fmt.Println()
			configmap, err := ek.ReadPropertyFile(createSecret)

			for key := range configmap {
				out.FmtGreen("⚿ %s", key)
			}

			if err != nil {
				out.FmtRed("Did not locate %s\n", createSecret)
				os.Exit(-1)
			}

			k8sutils.CreateSecret("default", "easykube-secrets", configmap)
		} else {
			out.FmtYellow("Warning, cluster created without importing secrets, this might affect your ability to pull images from private registries.")
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringP(constants.ARG_SECRETS, "s", "", "Property file to load as 'easykube-secrets', useful for imagepull secrets and other custom configuration")
}
