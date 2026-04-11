package cmd

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/core"
	jsutils "github.com/torloejborg/easykube/pkg/js"
)

type AddOptions struct {
	Args          []string
	ForceInstall  bool
	TargetCluster string
	NoDepends     bool
	DryRun        bool
}

func addActual(opts AddOptions, ek *core.Ek) error {

	if !opts.DryRun && !ek.ContainerRuntime.IsClusterRunning() {
		return errors.New("please create or start the cluster before installing addons")
	}
	allAddons, err := ek.AddonReader.GetAddons()

	ek.ExternalTools.EnsureLocalContext()

	wanted, missing := pickAddons(opts.Args, allAddons)
	if len(missing) > 0 {
		return fmt.Errorf("%d unknown addon(s) specified; %v", len(missing), strings.Join(missing, ", "))
	}

	if opts.TargetCluster != "" {
		ek.ExternalTools.SwitchContext(opts.TargetCluster)
		defer ek.ExternalTools.SwitchContext(constants.ClusterContext)
	}

	if opts.NoDepends {
		return jsutils.NewJsUtils(ek, wanted[0], false).ExecAddonScript(wanted[0])
	}

	toInstall, err := core.ResolveDependencies(wanted, allAddons)
	if err != nil {
		return err
	}

	installed, err := ek.Kubernetes.GetInstalledAddons()

	for _, addon := range toInstall {
		if slices.Contains(installed, addon.GetShortName()) && !opts.ForceInstall {
			ek.Printer.FmtGreen("%s already present in cluster", addon.GetShortName())
			continue
		}

		if err := jsutils.NewJsUtils(ek, addon, false).ExecAddonScript(addon); err != nil {
			return err
		}
	}

	return nil
}

func pickAddons(name []string, addons map[string]core.IAddon) ([]core.IAddon, []string) {
	result := make([]core.IAddon, 0)
	missing := make([]string, 0)

	for ni := range name {
		n := name[ni]
		found := false

		for i := range addons {
			if addons[i].GetShortName() == n || addons[i].GetName() == n {
				result = append(result, addons[i])
				found = true
				break
			}
		}

		if !found {
			missing = append(missing, n)
		}
	}

	return result, missing
}
