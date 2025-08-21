package cmd

import (
	"github.com/spf13/cobra"
	"github.com/torloj/easykube/ekctx"
)

// destroyCmd represents the destroy command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "shows the version of easykube CLI",
	Long:  `shows the version of easykube CLI`,
	Run: func(cmd *cobra.Command, args []string) {
		ekCtx := ekctx.GetAppContext(cmd)
		out := ekCtx.Printer
		out.FmtGreen("1.1.0")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
