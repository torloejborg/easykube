package cmd

import (
	"github.com/torloejborg/easykube/ekctx"
	"github.com/torloejborg/easykube/pkg"
	"github.com/torloejborg/easykube/pkg/constants"

	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stops the cluster node and registry container",
	Long:  "", Run: func(cmd *cobra.Command, args []string) {

		ctx := ekctx.GetAppContext(cmd)
		cru := pkg.CreateContainerRuntime()

		if cru.IsContainerRunning(constants.KIND_CONTAINER) {
			cru.StopContainer(constants.KIND_CONTAINER)
			ctx.Printer.FmtGreen("stopping %s", constants.KIND_CONTAINER)
		}

		if cru.IsContainerRunning(constants.REGISTRY_CONTAINER) {
			cru.StopContainer(constants.REGISTRY_CONTAINER)
			ctx.Printer.FmtGreen("stopping %s", constants.REGISTRY_CONTAINER)
		}
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
