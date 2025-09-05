/*
Copyright © 2025 Tor Løjborg <torverner@proton.me>
*/
package cmd

import (
	"github.com/torloejborg/easykube/ekctx"

	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/ek"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "edits your easykube configuration file",
	Long:  "editor is chosen via VISUAL or EDITOR environment variable", Run: func(cmd *cobra.Command, args []string) {
		cfg := ek.NewEasykubeConfig(ekctx.GetAppContext(cmd))
		cfg.MakeConfig()
		cfg.EditConfig()
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
