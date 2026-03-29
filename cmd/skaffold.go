package cmd

import (
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/core"

	"github.com/spf13/cobra"
)

// skaffoldCmd represents the skaffold command
var skaffoldCmd = &cobra.Command{
	Use:   "skaffold --name [] --location []",
	Short: "creates a new addon using a basic template",
	Long: `creates a new addon with a default deployment, service, ingress and configmap
  
  if installed without modification, will appear at http://<addonName>.localtest.me 
  and display "Hello <addonName>" in your browser.

  Useful for starting a new addon.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ek, err := CreateEasykube(core.CommandHelper(cmd),
			WithKubernetes(false),
			WithContainerRuntime(false),
			WithAddonReader(false),
			WithClusterUtils(false),
			WithRequiresConfigurationCreated(true),
		)
		if err != nil {
			return err
		}

		opts := SkaffoldOpts{
			AddonName:     ek.CommandContext.GetStringFlag(constants.ArgSkaffoldName),
			AddonLocation: ek.CommandContext.GetStringFlag(constants.ArgSkaffoldLocation),
		}

		return skaffoldActual(opts, ek)
	},
}

func init() {
	rootCmd.AddCommand(skaffoldCmd)
	skaffoldCmd.Flags().String(constants.ArgSkaffoldName, "", "Name of new addon")
	skaffoldCmd.Flags().String(constants.ArgSkaffoldLocation, "", "Destination within the addons repository")

	_ = skaffoldCmd.MarkFlagRequired(constants.ArgSkaffoldName)
	_ = skaffoldCmd.MarkFlagRequired(constants.ArgSkaffoldLocation)

}
