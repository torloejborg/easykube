package cmd

import (
	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/core"
)

func NewAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [addon...]",
		Short: "applies one or more addons located in the addon repository",
		Long:  `by default addons also applies their dependencies`,
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdhelper := core.CommandHelper(cmd)

			ek, err := CreateEasykube(cmdhelper,
				WithAddonReader(true),
				WithKubernetes(true),
				WithContainerRuntime(true),
				WithRequiresConfigurationCreated(true),
			)

			if err != nil {
				return err
			}

			addOpts := AddOptions{
				Args:          args,
				ForceInstall:  cmdhelper.GetBoolFlag(constants.FlagForce),
				TargetCluster: cmdhelper.GetStringFlag(constants.FlagCluster),
				NoDepends:     cmdhelper.GetBoolFlag(constants.FlagNoDepends),
				DryRun:        cmdhelper.IsDryRun(),
			}

			return addActual(addOpts, ek)
		},
		//ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		//
		//	_ = ez.InitializeEasykube()
		//	addons := make([]string, 0)
		//	a, e := ez.Kube.GetAddons()
		//	if e != nil {
		//		// ignore for now
		//		panic(e)
		//	}
		//	for _, i := range a {
		//		addons = append(addons, i.GetShortName())
		//	}
		//	return addons, cobra.ShellCompDirectiveNoFileComp
		//},
	}

	return cmd
}

func init() {
	addCmd := NewAddCmd()

	addCmd.Flags().BoolP(constants.FlagNoDepends, "n", false, "Do not apply dependent addons")
	addCmd.Flags().BoolP(constants.FlagForce, "f", false, "If already applied, force")
	addCmd.Flags().BoolP(constants.FlagPull, "p", false, "Download newer local images")
	addCmd.Flags().String(constants.FlagCluster, "", "Specify a different kube-context for installation")
	addCmd.Flags().String(constants.FlagKeyValue, "", "pass key/value pairs into script context")

	rootCmd.AddCommand(addCmd)

}
