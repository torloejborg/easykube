package cmd

import (
	"github.com/torloejborg/easykube/ekctx"
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
		ekCtx := ekctx.GetAppContext(cmd)
		out := ekCtx.Printer

		search := ez.Kube.FindContainer(constants.KIND_CONTAINER)

		if search.Found {
			out.FmtYellow("Stopping %s", constants.KIND_CONTAINER)
			if search.IsRunning {
				ez.Kube.StopContainer(search.ContainerID)
			}
			ez.Kube.RemoveContainer(search.ContainerID)
			out.FmtYellow("Removing %s", constants.KIND_CONTAINER)
		}
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
	destroyCmd.Flags().BoolP("purge", "p", false, "Also remove any configuration and persisted data")
}
