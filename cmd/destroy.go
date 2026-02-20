package cmd

import (
	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/ez"
)

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "kills the current easykube cluster",
	Long:  `stops and removes the easykube container, leaves the registry running`,
	RunE: func(cmd *cobra.Command, args []string) error {

		err := ez.InitializeEasykube()
		if err != nil {
			return err
		}

		return destroyActual()
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
