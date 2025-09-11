package cmd

import (
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"

	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "starts the cluster node and registry container",
	Long:  "", Run: func(cmd *cobra.Command, args []string) {
		ezk := ez.Kube
		type StartStatus struct {
			Name    string
			Message string
			OK      bool
		}

		x := func(container string) StartStatus {
			f := ez.Kube.FindContainer(container)
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
				ezk.StartContainer(container)
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
			ezk.FmtGreen(cluster.Message)
		} else {
			ezk.FmtRed(cluster.Message)
		}

		if registry.OK {
			ezk.FmtGreen(registry.Message)
		} else {
			ezk.FmtRed(registry.Message)
		}

		if !registry.OK && !cluster.OK {
			ezk.FmtYellow("Hint:\n")
			createCmd.Help()
		}

	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
