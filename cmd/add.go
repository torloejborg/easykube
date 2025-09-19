package cmd

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"
	jsutils "github.com/torloejborg/easykube/pkg/js"
)

// addCmd represents the install command
var addCmd = &cobra.Command{
	Use:   "add [addon...]",
	Short: "applies one or more addons located in the addon repository",
	Long:  `by default addons also applies their dependencies`,
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ezk := ez.Kube

		cmdHelper := ez.CommandHelper(cmd)

		addons, err := ezk.GetAddons()
		if err != nil {
			return err
		}

		forceInstall := cmdHelper.GetBoolFlag(constants.FLAG_FORCE)
		targetCluster := cmdHelper.GetStringFlag(constants.FLAG_CLUSTER)

		if !ezk.IsDryRun() {
			// ignore this check when in dry-run mode
			if !ezk.IsClusterRunning() {
				return errors.New("please create or start the cluster before installing addons")
			}
		}

		var installed []string

		if ezk.IsDryRun() {
			installed = make([]string, 0)
		} else {
			installed, err = ezk.GetInstalledAddons()
			if err != nil {
				return err
			}
		}

		// switch to the easykube context - this is purely to avoid trouble
		// user might have switched to another context to do work and forgot to change
		// context back to easykube. --context argument overrides this
		ezk.EnsureLocalContext()

		wanted, missing := pickAddons(args, addons)

		if len(missing) > 0 {
			return fmt.Errorf("%d unknown addon(s) specified; %v", len(missing), strings.Join(missing, ", "))
		}

		if len(targetCluster) > 0 {
			ezk.SwitchContext(targetCluster)
			defer ezk.SwitchContext(constants.CLUSTER_CONTEXT)
		}

		if cmdHelper.GetBoolFlag(constants.FLAG_NODEPENDS) {
			jsutils.NewJsUtils(cmdHelper, wanted[0]).ExecAddonScript(wanted[0])
		} else {
			toInstall, err := ez.ResolveDependencies(wanted, addons)
			if err != nil {
				return err
			}

			for idx := range toInstall {

				current := toInstall[idx]
				if slices.Contains(installed, current.ShortName) && !forceInstall {
					ezk.FmtYellow("%s already present in cluster", current.ShortName)
					continue
				}

				jserr := jsutils.NewJsUtils(cmdHelper, toInstall[idx]).ExecAddonScript(toInstall[idx])
				if jserr != nil {
					return jserr
				}
			}
		}
		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {

		addons := make([]string, 0)
		a, e := ez.Kube.GetAddons()
		if e != nil {
			// ignore for now
		}
		for _, i := range a {
			addons = append(addons, i.ShortName)
		}
		return addons, cobra.ShellCompDirectiveNoFileComp
	},
}

func pickAddons(name []string, addons map[string]*ez.Addon) ([]*ez.Addon, []string) {
	result := make([]*ez.Addon, 0)
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
