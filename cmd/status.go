package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/core"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "inspects you environment to see if prerequisites are met",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {

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
		status := ek.Status

		_ = status.DoBinaryCheck()
		fmt.Println()
		_ = status.DoContainerCheck()
		fmt.Println()
		_ = status.DoAddonRepositoryCheck()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
