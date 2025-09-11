package cmd

import (
	"github.com/torloejborg/easykube/pkg/ez"

	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "edits your easykube configuration file",
	Long:  "editor is chosen via VISUAL or EDITOR environment variable", Run: func(cmd *cobra.Command, args []string) {
		ezk := ez.Kube
		ezk.MakeConfig()
		ezk.EditConfig()
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
