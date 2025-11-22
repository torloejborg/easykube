package cmd

import (
	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/ez"
)

func runConfigActual(cmd *cobra.Command, args []string) error {

	ezk := ez.Kube
	err := ezk.MakeConfig()
	if err != nil {
		return err
	}
	ezk.EditConfig()

	return nil
}
