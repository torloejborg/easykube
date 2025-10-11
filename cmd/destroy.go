package cmd

import (
	"github.com/spf13/cobra"
)

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "kills the current easykube cluster",
	Long:  `stops and removes the easykube container, leaves the registry running`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return destroyActual()
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
