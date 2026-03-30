package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/core"
)

type RemoveOpts struct {
	AddonsToRemove []string
}

func removeActual(opts RemoveOpts, ek *core.Ek) error {

	// switch to the easykube context
	ek.ExternalTools.EnsureLocalContext()

	allAddons, err := ek.AddonReader.GetAddons()
	if err != nil {
		eMsg := "could not read addons"
		return errors.Join(errors.New(eMsg), err)
	}

	installedAddons, e := ek.Kubernetes.GetInstalledAddons()
	if e != nil {
		eMsg := fmt.Sprintf("Cannot get installed addons: %s (was the configmap deleted by accident?)", e.Error())
		return errors.New(eMsg)
	}

	if len(opts.AddonsToRemove) == 0 {
		ek.Printer.FmtRed("Please specify one or more addons to remove, usage below\n")
	}

	for i := range opts.AddonsToRemove {
		// is args[i] installed
		if slices.Contains(installedAddons, opts.AddonsToRemove[i]) {
			remove(allAddons[opts.AddonsToRemove[i]], ek)
		} else {
			ek.Printer.FmtYellow("%s is not installed", opts.AddonsToRemove[i])
		}
	}

	return nil
}

func remove(addon core.IAddon, ek *core.Ek) {
	addonDir := filepath.Dir(addon.GetAddonFile())

	outYaml := filepath.Join(addonDir, constants.KustomizeTargetOutput)

	ek.ExternalTools.DeleteYaml(outYaml)
	if ek.CommandContext.IsDryRun() {
		ek.Printer.FmtDryRun("rm %s", outYaml)
	} else {
		ek.Kubernetes.DeleteKeyFromConfigmap(constants.AddonCm, constants.DefaultNs, addon.GetShortName())

		if ek.CommandContext.IsVerbose() {
			ek.Printer.FmtVerbose("rm %s", outYaml)
		}
		err := ek.Fs.Remove(outYaml)
		if err != nil {
			ek.Printer.FmtYellow("%s could not be deleted", constants.KustomizeTargetOutput)
		}
	}

}
