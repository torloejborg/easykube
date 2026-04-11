package cmd

import (
	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/textutils"
	"github.com/torloejborg/easykube/pkg/vars"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "shows the version of easykube CLI",
	Long:  `shows the version of easykube CLI`,
	RunE: func(cmd *cobra.Command, args []string) error {

		printer := textutils.PrinterImpl{}
		printer.FmtGreen(vars.Version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
