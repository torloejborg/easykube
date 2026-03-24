package cmd

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"
)

// bootCmd represents the create command
var bootCmd = &cobra.Command{
	Use:   "boot",
	Short: "boots the easykube cluster",
	Long:  `bootstraps a kind cluster with an opinionated configuration`,
	RunE: func(cmd *cobra.Command, args []string) error {

		_ = ez.InitializeEasykube(
			ez.WithKubernetes(false),
			ez.WithContainerRuntime(false),
			ez.WithAddonReader(false),
			ez.WithClusterUtils(false))

		currentConfig, err := ez.Kube.LoadConfig()
		if err != nil {
			return errors.New("no configuration detected, create a configuration before booting")
		}

		err = ez.InitializeEasykube()
		if err != nil {
			return err
		}

		opts := BootOpts{
			Secrets: ez.CommandHelper(cmd).GetStringFlag(constants.ArgSecrets),
		}

		return createActualCmd(opts, currentConfig)
	},
}

func init() {
	bootCmd.Flags().StringP(constants.ArgSecrets, "s", "", "Property file to load as 'easykube-secrets', useful for image pull secrets and other custom configuration")
	rootCmd.AddCommand(bootCmd)
}
