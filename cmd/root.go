package cmd

import (
	"context"
	"log"
	"os"

	"github.com/torloejborg/easykube/ekctx"

	"github.com/spf13/cobra"
)

func withAppContext(ctx context.Context, appCtx *ekctx.EKContext) context.Context {
	return context.WithValue(ctx, ekctx.AppCtxKey, appCtx)
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
		cmd.SetContext(withAppContext(cmd.Context(), &ekctx.EKContext{
			Command: cmd,
			Logger:  log.New(os.Stdout, "", log.LstdFlags),
			Printer: &ekctx.Printer{}}))

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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	//	 rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.myapp.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	//rootCmd.Flags().BoolP("ci", "c", false, "use Easykube in CI mode for remote clusters")
}
