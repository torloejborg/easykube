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
	Long:  "", RunE: func(cmd *cobra.Command, args []string) error {
		ezk := ez.Kube
		if ezk.IsContainerRunning(constants.KIND_CONTAINER) {
			ezk.StopContainer(constants.KIND_CONTAINER)
			ezk.FmtGreen("stopping %s", constants.KIND_CONTAINER)
		}

		if ezk.IsContainerRunning(constants.REGISTRY_CONTAINER) {
			ezk.StopContainer(constants.REGISTRY_CONTAINER)
			ezk.FmtGreen("stopping %s", constants.REGISTRY_CONTAINER)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
