package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/ez"
	"github.com/torloejborg/easykube/pkg/vars"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "shows the version of easykube CLI",
	Long:  `shows the version of easykube CLI`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := ez.InitializeWithOpts(
			ez.WithKubernetes(false),
			ez.WithContainerRuntime(false))
		if err != nil {
			ez.Kube.FmtRed(err.Error())
			os.Exit(1)
		}
		ez.Kube.FmtGreen(vars.Version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
