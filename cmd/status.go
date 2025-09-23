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
	Run: func(cmd *cobra.Command, args []string) {
		status := ez.NewStatusBuilder()

		_ = status.DoBinaryCheck()
		fmt.Println()
		_ = status.DoContainerCheck()
		fmt.Println()
		_ = status.DoAddonRepositoryCheck()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
