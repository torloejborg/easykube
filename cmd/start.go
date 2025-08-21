package cmd

import (
	"github.com/torloj/easykube/ekctx"
	"github.com/torloj/easykube/pkg/constants"

	"github.com/spf13/cobra"
	"github.com/torloj/easykube/pkg/ek"
)

// configCmd represents the config command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "starts the cluster node and registry container",
	Long:  "", Run: func(cmd *cobra.Command, args []string) {
		ctx := ekctx.GetAppContext(cmd)
		cru := ek.NewContainerRuntime(ctx)

		type StartStatus struct {
			Name    string
			Message string
			OK      bool
		}

		x := func(container string) StartStatus {
			f := cru.FindContainer(container)
			if !f.Found {
				return StartStatus{
					Name:    container,
					Message: container + " container does not exist",
					OK:      false,
				}
			} else if f.IsRunning {
				return StartStatus{
					Name:    container,
					Message: container + " running",
					OK:      true,
				}
			} else if !f.IsRunning {
				cru.StartContainer(container)
				return StartStatus{
					Name:    container,
					Message: container + " started",
					OK:      true,
				}
			}
			return StartStatus{}
		}

		cluster := x(constants.KIND_CONTAINER)
		registry := x(constants.REGISTRY_CONTAINER)

		if cluster.OK {
			ctx.Printer.FmtGreen(cluster.Message)
		} else {
			ctx.Printer.FmtRed(cluster.Message)
		}

		if registry.OK {
			ctx.Printer.FmtGreen(registry.Message)
		} else {
			ctx.Printer.FmtRed(registry.Message)
		}

		if !registry.OK && !cluster.OK {
			ctx.Printer.FmtYellow("Hint:\n")
			createCmd.Help()
		}

	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
