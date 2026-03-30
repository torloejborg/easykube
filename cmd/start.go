package cmd

import (
	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/core"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "starts the cluster node and registry container",
	Long:  "", RunE: func(cmd *cobra.Command, args []string) error {

		ek, err := CreateEasykube(core.CommandHelper(cmd),
			WithKubernetes(false),
			WithContainerRuntime(true),
			WithAddonReader(false),
			WithClusterUtils(false),
			WithRequiresConfigurationCreated(false),
		)
		if err != nil {
			return err
		}

		return startActual(ek)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
