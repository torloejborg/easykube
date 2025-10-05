package cmd

import (
	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/ez"
)

// listCmd represents the list command
var listCmd = &cobra.Command{

	Use:   "list",
	Short: "lists available modules in the addon repository",
	Long:  "installed addons has a tick-mark",
	RunE: func(cmd *cobra.Command, args []string) error {

		cmdhelper := ez.CommandHelper(cmd)

		opts := ListOpts{
			PlainListing:  cmdhelper.GetBoolFlag("plain"),
			ShowInstalled: cmdhelper.GetBoolFlag("installed"),
		}

		return listActual(opts)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolP("plain", "p", false, "plain listing, do not render table")
	listCmd.Flags().BoolP("installed", "i", false, "only list installed addons")
}
