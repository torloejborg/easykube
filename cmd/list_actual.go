package cmd

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"sort"

	"github.com/olekukonko/tablewriter"
	"github.com/torloejborg/easykube/pkg/ez"
)

type ListOpts struct {
	PlainListing  bool
	ShowInstalled bool
}

func listActual(opts ListOpts) error {
	ezk := ez.Kube

	modules, aerr := ez.Kube.GetAddons()
	if aerr != nil {
		return errors.New(fmt.Sprintf("list failed: %s", aerr.Error()))
	}
	installed := make([]string, 0)
	if ezk.IsClusterRunning() {
		i, err := ez.Kube.GetInstalledAddons()
		if err != nil {
			errMsg := fmt.Sprintf("list failed, cannot get installed addons: %s (was the configmap deleted by accident?)", err.Error())
			return errors.New(errMsg)
		}
		installed = append(installed, i...)
	} else {
		ezk.FmtYellow("Kind cluster not running, will not show installed addons\n")
	}

	// Extract and sort the keys
	var keys []string
	for k := range modules {

		// Only listing installed addons
		if opts.ShowInstalled {
			if slices.Contains(installed, k) {
				keys = append(keys, k)
			}
		} else {
			keys = append(keys, k)
		}
	}

	sort.Strings(keys)

	if opts.PlainListing {
		for _, pm := range keys {
			ezk.FmtGreen(pm)
		}
	} else {

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"Addon", "Description"})

		for _, m := range keys {

			addonStr := modules[m].ShortName
			if slices.Contains(installed, modules[m].ShortName) {
				addonStr = modules[m].ShortName + " ✓"
			}

			row := []string{
				addonStr,
				modules[m].Config.Description,
			}

			_ = table.Append(row)

		}

		_ = table.Render() // Send output to stdout
	}
	
	return nil
}
