package cmd

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"
	jsutils "github.com/torloejborg/easykube/pkg/js"
)

type AddOptions struct {
	Args          []string
	ForceInstall  bool
	TargetCluster string
	NoDepends     bool
	DryRun        bool
}

func addActual(opts AddOptions, cmdHelper ez.ICobraCommandHelper) error {
	ezk := ez.Kube

	if !opts.DryRun && !ezk.IsClusterRunning() {
		return errors.New("please create or start the cluster before installing addons")
	}
	allAddons, err := ez.Kube.GetAddons()

	ezk.EnsureLocalContext()

	wanted, missing := pickAddons(opts.Args, allAddons)
	if len(missing) > 0 {
		return fmt.Errorf("%d unknown addon(s) specified; %v", len(missing), strings.Join(missing, ", "))
	}

	if opts.TargetCluster != "" {
		ezk.SwitchContext(opts.TargetCluster)
		defer ezk.SwitchContext(constants.CLUSTER_CONTEXT)
	}

	if opts.NoDepends {
		return jsutils.NewJsUtils(cmdHelper, wanted[0]).ExecAddonScript(wanted[0])
	}

	toInstall, err := ez.ResolveDependencies(wanted, allAddons)
	if err != nil {
		return err
	}

	installed, err := ez.Kube.GetInstalledAddons()

	for _, addon := range toInstall {
		if slices.Contains(installed, addon.GetShortName()) && !opts.ForceInstall {
			ezk.FmtGreen("%s already present in cluster", addon.GetShortName())
			continue
		}

		if err := jsutils.NewJsUtils(cmdHelper, addon).ExecAddonScript(addon); err != nil {
			return err
		}
	}

	return nil
}

func pickAddons(name []string, addons map[string]ez.IAddon) ([]ez.IAddon, []string) {
	result := make([]ez.IAddon, 0)
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
