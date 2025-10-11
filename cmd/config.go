package cmd

import (
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "edits your easykube configuration file",
	Long:  "editor is chosen via VISUAL or EDITOR environment variable",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigActual(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
