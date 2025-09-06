package cmd

import (
	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/ekctx"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "starts the embedded webserver",
	Long:  `embedded webserver which provides API access to easykube`,
	Run: func(cmd *cobra.Command, args []string) {
		ekCtx := ekctx.GetAppContext(cmd)
		out := ekCtx.Printer
		out.FmtGreen("todo: implement api server")
	},
}

func init() {
	serveCmd.Flags().Int("port", 8080, "port to use for the embedded webserver")
	rootCmd.AddCommand(serveCmd)
}
