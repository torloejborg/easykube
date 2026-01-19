package cmd

import (
	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/ez"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "edits your easykube configuration file",
	Long:  "editor is chosen via VISUAL or EDITOR environment variable",
	RunE: func(cmd *cobra.Command, args []string) error {

		err := ez.InitializeEasykube(
			ez.WithKubernetes(false),
			ez.WithContainerRuntime(false))
		if err != nil {
			return err
		}

		return runConfigActual(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
