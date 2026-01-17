package cmd

import (
	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/ez"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "starts the cluster node and registry container",
	Long:  "", RunE: func(cmd *cobra.Command, args []string) error {

		err := ez.InitializeWithOpts(
			ez.WithKubernetes(false))
		if err != nil {
			return err
		}

		return startActual()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
