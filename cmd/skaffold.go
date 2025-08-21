package cmd

import (
	"github.com/torloj/easykube/ekctx"
	"github.com/torloj/easykube/pkg/constants"
	"os"

	"github.com/spf13/cobra"
	"github.com/torloj/easykube/pkg/ek"
)

// createAddonCmd represents the createAddon command
var skaffoldCmd = &cobra.Command{
	Use:   "skaffold --name [] --location []",
	Short: "creates a new addon using a basic template",
	Long: `creates a new addon with a default deployment, service, ingress and configmap
  
  if installed without modification, will appear at http://<addonName>.localtest.me 
  and display "Hello <addonName>" in your browser.

  Useful for starting a new addon.
`,
	Run: func(cmd *cobra.Command, args []string) {
		ekCtx := ekctx.GetAppContext(cmd)
		out := ekCtx.Printer

		addonName := ekCtx.GetStringFlag(constants.ARG_SKAFFOLD_NAME)
		addonDest := ekCtx.GetStringFlag(constants.ARG_SKAFFOLD_LOCATION)

		conf := ek.NewEasykubeConfig(ekCtx)
		ekc, err := conf.LoadConfig()
		if nil != err {
			out.FmtGreen("cannot proceed without easykube configuration")
			os.Exit(-1)
		}

		skaf := ek.NewSkaffold(ekc.AddonDir)
		skaf.CreateNewAddon(addonName, addonDest)

	},
}

func init() {
	rootCmd.AddCommand(skaffoldCmd)
	skaffoldCmd.Flags().String(constants.ARG_SKAFFOLD_NAME, "", "Name of new addon")
	skaffoldCmd.Flags().String(constants.ARG_SKAFFOLD_LOCATION, "", "Destination within the addons repository")
}
