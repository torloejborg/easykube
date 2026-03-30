package cmd

import (
	"context"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/core"
)

func WithAppContext(ctx context.Context, appCtx *core.CobraCommandHelperImpl) context.Context {
	return context.WithValue(ctx, core.AppCtxKey, appCtx)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "easykube",
	Short: "applies yaml to a local cluster, batteries included.",
	Long: `
bootstrap a single node kubernetes cluster, install development platforms via addon-repositories

hint: start with 'easykube config'`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		ctx := core.CobraCommandHelperImpl{
			Command: cmd,
		}

		cmd.SetContext(WithAppContext(cmd.Context(), &ctx))

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

	userConfDir, _ := os.UserConfigDir()
	configDir := filepath.Join(userConfDir, "easykube")

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolP("dry-run", "d", false, "dry-run")
	rootCmd.PersistentFlags().String("config-dir", configDir, "override the config dir")
}
