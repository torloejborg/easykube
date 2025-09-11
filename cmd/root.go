package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"
)

func withAppContext(ctx context.Context, appCtx *CobraCommandHelperImpl) context.Context {
	return context.WithValue(ctx, AppCtxKey, appCtx)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "easykube",
	Short: "applies yaml to a local cluster, batteries included.",
	Long: `
bootstrap a single node kubernetes cluster, install development platforms via addon-repositories

hint: start with 'easykube config'
`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
		}
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		ctx := CobraCommandHelperImpl{
			Command: cmd,
		}

		cmd.SetContext(withAppContext(cmd.Context(), &ctx))

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

}
