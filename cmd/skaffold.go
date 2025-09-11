package cmd

import (
	"os"

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
	Run: func(cmd *cobra.Command, args []string) {
		commandHelper := CommandHelper(cmd)

		addonName := commandHelper.GetStringFlag(constants.ARG_SKAFFOLD_NAME)
		addonDest := commandHelper.GetStringFlag(constants.ARG_SKAFFOLD_LOCATION)

		ekc, err := ez.Kube.LoadConfig()
		if nil != err {
			ez.Kube.FmtGreen("cannot proceed without easykube configuration")
			os.Exit(-1)
		}

		skaf := ez.NewSkaffold(ekc.AddonDir)
		skaf.CreateNewAddon(addonName, addonDest)

	},
}

func init() {
	rootCmd.AddCommand(skaffoldCmd)
	skaffoldCmd.Flags().String(constants.ARG_SKAFFOLD_NAME, "", "Name of new addon")
	skaffoldCmd.Flags().String(constants.ARG_SKAFFOLD_LOCATION, "", "Destination within the addons repository")
}
