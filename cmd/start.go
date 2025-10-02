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
	Long:  "", RunE: func(cmd *cobra.Command, args []string) error {
		ezk := ez.Kube
		type StartStatus struct {
			Name    string
			Message string
			OK      bool
		}

		x := func(container string) (*StartStatus, error) {
			f, err := ez.Kube.FindContainer(container)
			if err != nil {
				return nil, err
			}

			if !f.Found {
				return &StartStatus{
					Name:    container,
					Message: container + " container does not exist",
					OK:      false,
				}, nil
			} else if f.IsRunning {
				return &StartStatus{
					Name:    container,
					Message: container + " running",
					OK:      true,
				}, nil
			} else if !f.IsRunning {
				ezk.StartContainer(container)
				return &StartStatus{
					Name:    container,
					Message: container + " started",
					OK:      true,
				}, nil
			}
			return &StartStatus{}, nil
		}

		cluster, err := x(constants.KIND_CONTAINER)
		if err != nil {
			return err
		}
		registry, err := x(constants.REGISTRY_CONTAINER)
		if err != nil {
			return err
		}

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
			ezk.FmtGreen("Hint:\n")
			createCmd.Help()
		}

		return nil

	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
