package cmd

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/core"
)

// bootCmd represents the create command
var bootCmd = &cobra.Command{
	Use:   "boot",
	Short: "boots the easykube cluster",
	Long:  `bootstraps a kind cluster with an opinionated configuration`,
	RunE: func(cmd *cobra.Command, args []string) error {

		cmdhelper := core.CommandHelper(cmd)

		ek, err := CreateEasykube(cmdhelper,
			WithKubernetes(true),
			WithContainerRuntime(true),
			WithAddonReader(true),
			WithClusterUtils(true),
			WithRequiresConfigurationCreated(true),
		)
		if err != nil {
			return err
		}

		currentConfig, err := ek.Config.LoadConfig()
		if err != nil {
			return errors.New("no configuration detected, create a configuration before booting")
		}

		return createActualCmd(ek, currentConfig)
	},
}

func init() {
	rootCmd.AddCommand(bootCmd)
}
