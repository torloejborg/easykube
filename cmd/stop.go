package cmd

import (
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"

	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stops the cluster node and registry container",
	Long:  "", Run: func(cmd *cobra.Command, args []string) {

		if ez.Kube.IsContainerRunning(constants.KIND_CONTAINER) {
			ez.Kube.StopContainer(constants.KIND_CONTAINER)
			ez.Kube.FmtGreen("stopping %s", constants.KIND_CONTAINER)
		}

		if ez.Kube.IsContainerRunning(constants.REGISTRY_CONTAINER) {
			ez.Kube.StopContainer(constants.REGISTRY_CONTAINER)
			ez.Kube.FmtGreen("stopping %s", constants.REGISTRY_CONTAINER)
		}
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
