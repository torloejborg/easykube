package cmd

import (
	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/core"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stops the cluster node and registry container",
	Long:  "", RunE: func(cmd *cobra.Command, args []string) error {

		ek, err := CreateEasykube(core.CommandHelper(cmd),
			WithKubernetes(false),
			WithContainerRuntime(true),
			WithAddonReader(false),
			WithClusterUtils(false),
			WithRequiresConfigurationCreated(false),
		)
		if err != nil {
			return err
		}

		if running, err := ek.ContainerRuntime.IsContainerRunning(constants.KindContainer); err != nil {
			return err
		} else if running {
			ek.Printer.FmtGreen("stopping %s", constants.KindContainer)
			if e := ek.ContainerRuntime.StopContainer(constants.KindContainer); e != nil {
				return e
			}
		}

		if running, err := ek.ContainerRuntime.IsContainerRunning(constants.RegistryContainer); err != nil {
			return err
		} else if running {
			ek.Printer.FmtGreen("stopping %s", constants.RegistryContainer)
			if e := ek.ContainerRuntime.StopContainer(constants.RegistryContainer); e != nil {
				return e
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
