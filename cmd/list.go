package cmd

import (
	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/core"
)

// listCmd represents the list command
var listCmd = &cobra.Command{

	Use:   "list",
	Short: "lists available modules in the addon repository",
	Long:  "installed addons has a tick-mark",
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

		opts := ListOpts{
			PlainListing:  ek.CommandContext.GetBoolFlag("plain"),
			ShowInstalled: ek.CommandContext.GetBoolFlag("installed"),
		}

		return listActual(opts, ek)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolP("plain", "p", false, "plain listing, do not render table")
	listCmd.Flags().BoolP("installed", "i", false, "only list installed addons")
}
