package cmd

import (
	"os"

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

		if ez.Kube.GetBoolFlag("use-defaults") {
			ez.Kube.MakeConfig()
			os.Exit(0)
		}

		if ez.Kube.GetBoolFlag("edit") {
			ez.Kube.EditConfig()
			os.Exit(0)
		}

		return runConfigActualInteractive(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.Flags().BoolP("edit", "e", false, "edit config file with editor")
	configCmd.Flags().BoolP("use-defaults", "u", false, "create a configuration with default values")
}
