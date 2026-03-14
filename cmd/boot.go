package cmd

import (
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

		err := ez.InitializeEasykube()
		if err != nil {
			return err
		}

		opts := BootOpts{
			Secrets: ez.CommandHelper(cmd).GetStringFlag(constants.ARG_SECRETS),
		}

		return createActualCmd(opts)
	},
}

func init() {
	bootCmd.Flags().StringP(constants.ARG_SECRETS, "s", "", "Property file to load as 'easykube-secrets', useful for image pull secrets and other custom configuration")
	rootCmd.AddCommand(bootCmd)
}
