package cmd

import (
	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/core"
)

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "kills the current easykube cluster",
	Long:  `stops and removes the easykube container, leaves the registry running`,
	RunE: func(cmd *cobra.Command, args []string) error {

		ek, err := CreateEasykube(core.CommandHelper(cmd),
			WithKubernetes(false),
			WithContainerRuntime(true),
			WithAddonReader(false),
			WithClusterUtils(false),
			WithRequiresConfigurationCreated(true),
		)
		if err != nil {
			return err
		}

		return destroyActual(ek)
	},
}

func init() {
	destroyCmd.Flags().BoolP("purge", "p", false, "deletes configuration, local registry, and persisted data")
	rootCmd.AddCommand(destroyCmd)
}
