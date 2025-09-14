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
	RunE: func(cmd *cobra.Command, args []string) error {
		ezk := ez.Kube
		search, err := ezk.FindContainer(constants.KIND_CONTAINER)
		if err != nil {
			return err
		}

		if search.Found {
			ezk.FmtYellow("Stopping %s", constants.KIND_CONTAINER)
			if search.IsRunning {
				ezk.StopContainer(search.ContainerID)
			}
			ezk.RemoveContainer(search.ContainerID)
			ezk.FmtYellow("Removing %s", constants.KIND_CONTAINER)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
	destroyCmd.Flags().BoolP("purge", "p", false, "Also remove any configuration and persisted data")
}
