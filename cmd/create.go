package cmd

import (
	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "creates the easykube cluster",
	Long:  `bootstraps a kind cluster with an opinionated configuration`,
	RunE: func(cmd *cobra.Command, args []string) error {

		cmdhelper := ez.CommandHelper(cmd)
		opts := CreateOpts{
			Secrets: cmdhelper.GetStringFlag(constants.ARG_SECRETS),
		}

		return createActualCmd(opts, cmdhelper)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringP(constants.ARG_SECRETS, "s", "", "Property file to load as 'easykube-secrets', useful for image pull secrets and other custom configuration")
}
