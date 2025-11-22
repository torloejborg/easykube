package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"
)

type RemoveOpts struct {
	AddonsToRemove []string
}

func removeActual(opts RemoveOpts) error {
	ezk := ez.Kube
	// switch to the easykube context
	ezk.EnsureLocalContext()

	allAddons, err := ez.Kube.GetAddons()
	if err != nil {
		eMsg := "could not read addons"
		return errors.Join(errors.New(eMsg), err)
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

func remove(addon ez.IAddon) {
	addonDir := filepath.Dir(addon.GetAddonFile())
	ezk := ez.Kube

	outYaml := filepath.Join(addonDir, constants.KUSTOMIZE_TARGET_OUTPUT)

	ezk.DeleteYaml(outYaml)
	if ezk.IsDryRun() {
		ezk.FmtDryRun("rm %s", outYaml)
	} else {
		ezk.DeleteKeyFromConfigmap(constants.ADDON_CM, constants.DEFAULT_NS, addon.GetShortName())

		if ezk.IsVerbose() {
			ezk.FmtVerbose("rm %s", outYaml)
		}
		err := ezk.Remove(outYaml)
		if err != nil {
			ezk.FmtYellow("%s could not be deleted", constants.KUSTOMIZE_TARGET_OUTPUT)
		}
	}

}
