package cmd

import (
	"github.com/torloejborg/easykube/pkg/constants"

	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/ez"
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

		err := ez.InitializeEasykube(
			ez.WithKubernetes(false),
			ez.WithContainerRuntime(false))
		if err != nil {
			return err
		}

		commandHelper := ez.CommandHelper(cmd)
		opts := SkaffoldOpts{
			AddonName:     commandHelper.GetStringFlag(constants.ArgSkaffoldName),
			AddonLocation: commandHelper.GetStringFlag(constants.ArgSkaffoldLocation),
		}

		return skaffoldActual(opts)
	},
}

func init() {
	rootCmd.AddCommand(skaffoldCmd)
	skaffoldCmd.Flags().String(constants.ArgSkaffoldName, "", "Name of new addon")
	skaffoldCmd.Flags().String(constants.ArgSkaffoldLocation, "", "Destination within the addons repository")

	_ = skaffoldCmd.MarkFlagRequired(constants.ArgSkaffoldName)
	_ = skaffoldCmd.MarkFlagRequired(constants.ArgSkaffoldLocation)

}
