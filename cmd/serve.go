package cmd

import (
	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/ez"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "starts the embedded webserver",
	Long:  `embedded webserver which provides API access to easykube`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ez.Kube.FmtGreen("todo: implement api server")
		return nil
	},
}

func init() {
	serveCmd.Flags().Int("port", 8080, "port to use for the embedded webserver")
	rootCmd.AddCommand(serveCmd)
}
