package cmd

import (
	"fmt"

	"github.com/torloejborg/easykube/pkg/ez"

	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "inspects you environment to see if prerequisites are met",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {

		err := ez.InitializeEasykube()
		if err != nil {
			return err
		}

		status := ez.NewStatusBuilder()

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
