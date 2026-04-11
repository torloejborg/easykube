package cmd

import (
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/core"
)

func startActual(ek *core.Ek) error {

	type StartStatus struct {
		Name    string
		Message string
		OK      bool
	}

	findAndStart := func(container string) (*StartStatus, error) {
		f, err := ek.ContainerRuntime.FindContainer(container)
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
			if err := ek.ContainerRuntime.StartContainer(container); err != nil {
				return nil, err
			}
			return &StartStatus{
				Name:    container,
				Message: container + " started",
				OK:      true,
			}, nil
		}
		return &StartStatus{}, nil
	}

	cluster, err := findAndStart(constants.KindContainer)
	if err != nil {
		return err
	}
	registry, err := findAndStart(constants.RegistryContainer)
	if err != nil {
		return err
	}

	if cluster.OK {
		ek.Printer.FmtGreen(cluster.Message)
	} else {
		ek.Printer.FmtRed(cluster.Message)
	}

	if registry.OK {
		ek.Printer.FmtGreen(registry.Message)
	} else {
		ek.Printer.FmtRed(registry.Message)
	}

	if !registry.OK && !cluster.OK {
		ek.Printer.FmtGreen("Hint:\n")
		_ = bootCmd.Help()
	}

	return nil

}
