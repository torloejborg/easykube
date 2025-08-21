package cmd

import (
	"github.com/torloj/easykube/ekctx"
	"github.com/torloj/easykube/pkg/constants"

	"github.com/spf13/cobra"
	"github.com/torloj/easykube/pkg/ek"
)

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "kills the current easykube cluster",
	Long:  `stops and removes the easykube container, leaves the registry running`,
	Run: func(cmd *cobra.Command, args []string) {
		ekCtx := ekctx.GetAppContext(cmd)
		out := ekCtx.Printer

		cru := ek.NewContainerRuntime(ekCtx)
		search := cru.FindContainer(constants.KIND_CONTAINER)

		if search.Found {
			out.FmtYellow("Stopping %s", constants.KIND_CONTAINER)
			if search.IsRunning {
				cru.StopContainer(search.ContainerID)
			}
			cru.RemoveContainer(search.ContainerID)
			out.FmtYellow("Removing %s", constants.KIND_CONTAINER)
		}
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
	destroyCmd.Flags().BoolP("purge", "p", false, "Also remove any configuration and persisted data")
}
