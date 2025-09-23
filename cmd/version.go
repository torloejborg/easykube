package cmd

import (
	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "shows the version of easykube CLI",
	Long:  `shows the version of easykube CLI`,
	Run: func(cmd *cobra.Command, args []string) {

		ez.Kube.FmtGreen(constants.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
