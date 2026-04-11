package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/core"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "edits your easykube configuration file",
	Long:  "editor is chosen via VISUAL or EDITOR environment variable",
	RunE: func(cmd *cobra.Command, args []string) error {

		ek, err := CreateEasykube(core.CommandHelper(cmd),
			WithKubernetes(false),
			WithContainerRuntime(false),
			WithAddonReader(false),
			WithClusterUtils(false),
			WithRequiresConfigurationCreated(false),
		)
		if err != nil {
			return err
		}

		if ek.CommandContext.GetBoolFlag(constants.FlagUseDefaults) {
			err := ek.Config.MakeConfig()
			if err != nil {
				panic(err)
			}
			os.Exit(0)
		}

		if ek.CommandContext.GetBoolFlag(constants.FlagEdit) {
			err := ek.Config.EditConfig()
			if err != nil {
				ek.Printer.FmtGreen(err.Error())
			}
			os.Exit(0)
		}

		return runConfigActualInteractive(ek)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.Flags().BoolP("edit", "e", false, "edit config file with editor")
	configCmd.Flags().BoolP("use-defaults", "u", false, "create a configuration with default values")
}
