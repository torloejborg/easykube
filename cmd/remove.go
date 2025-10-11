package cmd

import (
	"github.com/torloejborg/easykube/pkg/ez"

	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove [addon...]",
	Short: "removes a previously installed addon",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {

		opts := RemoveOpts{AddonsToRemove: args}

		return removeActual(opts)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		k8sutils := ez.CreateK8sUtilsImpl()
		clusterAddons, e := k8sutils.GetInstalledAddons()
		if e == nil {
			return clusterAddons, cobra.ShellCompDirectiveNoFileComp
		}

		return nil, cobra.ShellCompDirectiveNoFileComp
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
