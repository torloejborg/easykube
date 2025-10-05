package cmd

import (
	"errors"
	"fmt"
	"slices"

	"github.com/torloejborg/easykube/pkg/ez"
)

type RemoveOpts struct {
	AddonsToRemove []string
}

func removeActual(opts RemoveOpts) error {
	ezk := ez.Kube
	// switch to the easykube context
	ezk.EnsureLocalContext()

	allAddons, aerr := ez.Kube.GetAddons()
	if aerr != nil {
		ezk.FmtRed("could not get addons %s", aerr.Error())
	}

	installedAddons, e := ezk.GetInstalledAddons()
	if e != nil {
		eMsg := fmt.Sprintf("Cannot get installed addons: %s (was the configmap deleted by accident?)", e.Error())
		return errors.New(eMsg)
	}

	if len(opts.AddonsToRemove) == 0 {
		ez.Kube.FmtRed("Please specify one or more addons to remove, usage below\n")
	}

	for i := range opts.AddonsToRemove {
		// is args[i] installed
		if slices.Contains(installedAddons, opts.AddonsToRemove[i]) {
			remove(allAddons[opts.AddonsToRemove[i]])
		} else {
			ez.Kube.FmtYellow("%s is not installed", opts.AddonsToRemove[i])
		}
	}

	return nil
}
