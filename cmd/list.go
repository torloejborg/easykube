package cmd

import (
	"os"
	"slices"
	"sort"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/ekctx"
	"github.com/torloejborg/easykube/pkg/ez"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "lists available modules in the addon repository",
	Long:  "installed addons are marked with tickmark",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := ekctx.GetAppContext(cmd)
		modules := ez.Kube.GetAddons()
		installed := make([]string, 0)
		if ez.Kube.IsClusterRunning() {
			i, err := ez.Kube.GetInstalledAddons()
			if err != nil {
				ez.Kube.FmtRed("Cannot get installed addons: %v (was the configmap deleted by accident?)", err)
				os.Exit(1)
			}
			installed = append(installed, i...)
		} else {
			ez.Kube.FmtYellow("Kind cluster not running, will not show installed addons\n")
		}

		// Extract and sort the keys
		var keys []string
		for k := range modules {

			// Only listing installed addons
			if ctx.GetBoolFlag("installed") {
				if slices.Contains(installed, k) {
					keys = append(keys, k)
				}
			} else {
				keys = append(keys, k)
			}
		}

		sort.Strings(keys)

		if ctx.GetBoolFlag("plain") {
			for _, pm := range keys {
				ez.Kube.FmtGreen(pm)
			}
		} else {

			table := tablewriter.NewWriter(os.Stdout)
			table.Header([]string{"Addon", "Description"})

			for _, m := range keys {

				addonStr := modules[m].ShortName
				if slices.Contains(installed, modules[m].ShortName) {
					addonStr = modules[m].ShortName + " ✓"
				}

				row := []string{
					addonStr,
					modules[m].Config.Description,
				}

				table.Append(row)

			}

			table.Render() // Send output to stdout
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolP("plain", "p", false, "plain listing, do not render table")
	listCmd.Flags().BoolP("installed", "i", false, "only list installed addons")
}
