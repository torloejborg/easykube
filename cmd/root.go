package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/ez"
)

func WithAppContext(ctx context.Context, appCtx *ez.CobraCommandHelperImpl) context.Context {
	return context.WithValue(ctx, ez.AppCtxKey, appCtx)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "easykube",
	Short: "applies yaml to a local cluster, batteries included.",
	Long: `
bootstrap a single node kubernetes cluster, install development platforms via addon-repositories

hint: start with 'easykube config'`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		ctx := ez.CobraCommandHelperImpl{
			Command: cmd,
		}

		cmd.SetContext(WithAppContext(cmd.Context(), &ctx))

		ez.Kube.UseCmdContext(&ez.CobraCommandHelperImpl{Command: cmd})
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolP("dry-run", "d", false, "dry-run")
}
