package cmd

import (
	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/ekctx"
)

var Version = "latest" // set by linker flag

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "shows the version of easykube CLI",
	Long:  `shows the version of easykube CLI`,
	Run: func(cmd *cobra.Command, args []string) {
		ekCtx := ekctx.GetAppContext(cmd)
		out := ekCtx.Printer
		out.FmtGreen(Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
