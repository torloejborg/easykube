package cmd

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"sort"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/ez"
)

// listCmd represents the list command
var listCmd = &cobra.Command{

	Use:   "list",
	Short: "lists available modules in the addon repository",
	Long:  "installed addons has a tick-mark",
	RunE: func(cmd *cobra.Command, args []string) error {
		ezk := ez.Kube
		commandHelper := ez.CommandHelper(cmd)

		modules, aerr := ez.Kube.GetAddons()
		if aerr != nil {
			return errors.New(fmt.Sprintf("list failed: %s", aerr.Error()))
		}
		installed := make([]string, 0)
		if ezk.IsClusterRunning() {
			i, err := ez.Kube.GetInstalledAddons()
			if err != nil {
				errMsg := fmt.Sprintf("list failed, cannot get installed addons: %s (was the configmap deleted by accident?)", err.Error())
				return errors.New(errMsg)
			}
			installed = append(installed, i...)
		} else {
			ezk.FmtYellow("Kind cluster not running, will not show installed addons\n")
		}

		// Extract and sort the keys
		var keys []string
		for k := range modules {

			// Only listing installed addons
			if commandHelper.GetBoolFlag("installed") {
				if slices.Contains(installed, k) {
					keys = append(keys, k)
				}
			} else {
				keys = append(keys, k)
			}
		}

		sort.Strings(keys)

		if commandHelper.GetBoolFlag("plain") {
			for _, pm := range keys {
				ezk.FmtGreen(pm)
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

				_ = table.Append(row)

			}

			_ = table.Render() // Send output to stdout
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolP("plain", "p", false, "plain listing, do not render table")
	listCmd.Flags().BoolP("installed", "i", false, "only list installed addons")
}
