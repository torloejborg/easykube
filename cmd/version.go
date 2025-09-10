package cmd

import (
	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/ez"
)

var Version = "latest" // set by linker flag

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "shows the version of easykube CLI",
	Long:  `shows the version of easykube CLI`,
	Run: func(cmd *cobra.Command, args []string) {
		out := ez.Kube.Printer
		out.FmtGreen(Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
