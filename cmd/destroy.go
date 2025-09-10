package cmd

import (
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"

	"github.com/spf13/cobra"
)

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "kills the current easykube cluster",
	Long:  `stops and removes the easykube container, leaves the registry running`,
	Run: func(cmd *cobra.Command, args []string) {

		search := ez.Kube.FindContainer(constants.KIND_CONTAINER)

		if search.Found {
			ez.Kube.FmtYellow("Stopping %s", constants.KIND_CONTAINER)
			if search.IsRunning {
				ez.Kube.StopContainer(search.ContainerID)
			}
			ez.Kube.RemoveContainer(search.ContainerID)
			ez.Kube.FmtYellow("Removing %s", constants.KIND_CONTAINER)
		}
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
	destroyCmd.Flags().BoolP("purge", "p", false, "Also remove any configuration and persisted data")
}
