package cmd

import (
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/ekctx"
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ek"
	jsutils "github.com/torloejborg/easykube/pkg/js"
)

// addCmd represents the install command
var addCmd = &cobra.Command{
	Use:   "add [addon...]",
	Short: "applies one or more addons located in the addon repository",
	Long:  `by default addons also applies their dependencies`,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		ekCtx := ekctx.GetAppContext(cmd)
		out := ekCtx.Printer
		reader := ek.NewAddonReader(ekCtx)
		addons := reader.GetAddons()
		tools := ek.NewExternalTools(ekCtx)

		forceInstall := ekCtx.GetBoolFlag(constants.FLAG_FORCE)
		//noDepends := ekCtx.GetBoolFlag(constants.FLAG_NODEPENDS)
		targetCluster := ekCtx.GetStringFlag(constants.FLAG_CLUSTER)
		installed := ek.NewK8SUtils(ekCtx).GetInstalledAddons()

		// switch to the easykube context - this is purely to avoid trouble
		// user might have switched to another context to do work, and forgot to change
		// context back to easykube. --context argument overrides this
		ek.NewExternalTools(ekCtx).EnsureLocalContext()

		wanted, missing := pickAddons(args, addons)

		if len(missing) > 0 {
			out.FmtRed("%d unknown addon(s) specified; %v", len(missing), strings.Join(missing, ", "))
			os.Exit(-1)
		}

		if len(targetCluster) > 0 {
			tools.SwitchContext(targetCluster)
			defer tools.SwitchContext(constants.CLUSTER_CONTEXT)
		}

		if ekCtx.GetBoolFlag(constants.FLAG_NODEPENDS) {
			jsutils.NewJsUtils(ekCtx, wanted[0]).ExecAddonScript(wanted[0])
		} else {
			toInstall, err := ek.ResolveDependencies(wanted, addons)
			if err != nil {
				out.FmtRed("dependency resolution failed: %v", err)
			}

			for idx := range toInstall {

				current := toInstall[idx]
				if slices.Contains(installed, current.ShortName) && !forceInstall {
					out.FmtYellow("%s already present in cluster", current.ShortName)
					continue
				}

				jsutils.NewJsUtils(ekCtx, toInstall[idx]).ExecAddonScript(toInstall[idx])
			}
		}

	},
}

func pickAddons(name []string, addons map[string]*ek.Addon) ([]*ek.Addon, []string) {
	result := make([]*ek.Addon, 0)
	missing := make([]string, 0)

	for ni := range name {
		n := name[ni]
		found := false

		for i := range addons {
			if addons[i].ShortName == n || addons[i].Name == n {
				result = append(result, addons[i])
				found = true
				break
			}
		}

		if !found {
			missing = append(missing, n)
		}
	}

	return result, missing
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().BoolP(constants.FLAG_NODEPENDS, "n", false, "Do not apply dependent addons")
	addCmd.Flags().BoolP(constants.FLAG_FORCE, "f", false, "If already applied, force")
	addCmd.Flags().BoolP(constants.FLAG_PULL, "p", false, "Download newer local images")
	addCmd.Flags().String(constants.FLAG_CLUSTER, "", "Specify a different kube-context for installation")
	addCmd.Flags().String(constants.FLAG_KEYVALUE, "", "pass key/value pairs into script context")
}
