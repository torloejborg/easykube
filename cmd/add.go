package cmd

import (
	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"
)

func NewAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [addon...]",
		Short: "applies one or more addons located in the addon repository",
		Long:  `by default addons also applies their dependencies`,
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			cmdhelper := ez.CommandHelper(cmd)

			addOpts := AddOptions{
				Args:          args,
				ForceInstall:  cmdhelper.GetBoolFlag(constants.FLAG_FORCE),
				TargetCluster: cmdhelper.GetStringFlag(constants.FLAG_CLUSTER),
				NoDepends:     cmdhelper.GetBoolFlag(constants.FLAG_NODEPENDS),
				DryRun:        cmdhelper.IsDryRun(),
			}

			return addActual(addOpts, cmdhelper)
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

	return cmd
}

func init() {
	addCmd := NewAddCmd()

	addCmd.Flags().BoolP(constants.FLAG_NODEPENDS, "n", false, "Do not apply dependent addons")
	addCmd.Flags().BoolP(constants.FLAG_FORCE, "f", false, "If already applied, force")
	addCmd.Flags().BoolP(constants.FLAG_PULL, "p", false, "Download newer local images")
	addCmd.Flags().String(constants.FLAG_CLUSTER, "", "Specify a different kube-context for installation")
	addCmd.Flags().String(constants.FLAG_KEYVALUE, "", "pass key/value pairs into script context")

	rootCmd.AddCommand(addCmd)

}
