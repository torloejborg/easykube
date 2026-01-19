package cmd

import (
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"

	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stops the cluster node and registry container",
	Long:  "", RunE: func(cmd *cobra.Command, args []string) error {

		err := ez.InitializeEasykube(
			ez.WithKubernetes(false))
		if err != nil {
			return err
		}

		ezk := ez.Kube
		if running, err := ezk.IsContainerRunning(constants.KIND_CONTAINER); err != nil {
			return err
		} else if running {
			ezk.FmtGreen("stopping %s", constants.KIND_CONTAINER)
			if e := ezk.StopContainer(constants.KIND_CONTAINER); e != nil {
				return e
			}
		}

		if running, err := ezk.IsContainerRunning(constants.REGISTRY_CONTAINER); err != nil {
			return err
		} else if running {
			ezk.FmtGreen("stopping %s", constants.REGISTRY_CONTAINER)
			if e := ezk.StopContainer(constants.REGISTRY_CONTAINER); e != nil {
				return e
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
