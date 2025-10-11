package cmd

import (
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "starts the cluster node and registry container",
	Long:  "", RunE: func(cmd *cobra.Command, args []string) error {
		return startActual()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
