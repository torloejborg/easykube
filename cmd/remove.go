package cmd

import (
	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/core"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove [addon...]",
	Short: "removes a previously installed addon",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {

		ek, err := CreateEasykube(core.CommandHelper(cmd),
			WithKubernetes(true),
			WithContainerRuntime(true),
			WithAddonReader(true),
			WithClusterUtils(true),
			WithRequiresConfigurationCreated(true),
		)
		if err != nil {
			return err
		}

		opts := RemoveOpts{AddonsToRemove: args}

		return removeActual(opts, ek)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {

		ek, _ := CreateEasykube(core.CommandHelper(cmd),
			WithKubernetes(true),
			WithContainerRuntime(true),
			WithAddonReader(true),
			WithClusterUtils(true),
			WithRequiresConfigurationCreated(true),
		)

		clusterAddons, e := ek.Kubernetes.GetInstalledAddons()
		if e == nil {
			return clusterAddons, cobra.ShellCompDirectiveNoFileComp
		} else {
			panic(e)
		}

		return nil, cobra.ShellCompDirectiveNoFileComp
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
