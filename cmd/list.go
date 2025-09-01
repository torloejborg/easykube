package cmd

import (
	"os"
	"slices"
	"sort"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/torloj/easykube/ekctx"
	"github.com/torloj/easykube/pkg/ek"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "lists available modules in the addon repository",
	Long:  "installed addons are marked with tickmark",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := ekctx.GetAppContext(cmd)
		cru := ek.NewContainerRuntime(ctx)
		modules := ek.NewAddonReader(ctx).GetAddons()
		installed := make([]string, 0)
		if cru.IsClusterRunning() {
			installed = append(installed, ek.NewK8SUtils(ctx).GetInstalledAddons()...)
		} else {
			ctx.Printer.FmtYellow("Kind cluster not running, will not show installed addons\n")
		}

		// Extract and sort the keys
		var keys []string
		for k := range modules {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		if ctx.GetBoolFlag("plain") {
			for _, pm := range keys {
				ctx.Printer.FmtGreen(pm)
			}
		} else {

			table := tablewriter.NewWriter(os.Stdout)
			table.Header([]string{"Addon", "Description"})

			for _, m := range keys {

				isInstalled := modules[m].ShortName

				if slices.Contains(installed, modules[m].ShortName) {
					isInstalled = modules[m].ShortName + " ✓"
				}

				row := []string{
					isInstalled,
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
}
