package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"

	"github.com/spf13/cobra"
)

func remove(addon *ez.Addon) {
	// enter the addon directory
	ez.PushDir(filepath.Dir(addon.File))
	defer ez.PopDir()

	yamlFile := ez.Kube.KustomizeBuild(".")
	ezk := ez.Kube
	ezk.DeleteYaml(yamlFile)
	ezk.FmtGreen("removed %s", addon.ShortName)
	ezk.DeleteKeyFromConfigmap(constants.ADDON_CM, constants.DEFAULT_NS, addon.ShortName)

	err := ezk.Remove(constants.KUSTOMIZE_TARGET_OUTPUT)
	if err != nil {
		ezk.FmtYellow("%s could not be deleted", constants.KUSTOMIZE_TARGET_OUTPUT)
	}
}

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove [addon...]",
	Short: "removes a previously installed addon",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		if len(args) == 0 {
			ez.Kube.FmtRed("Please specify one or more addons to remove, usage below\n")
			err := cmd.Help()
			if err != nil {
				return err
			}
		}

		for i := range args {
			// is args[i] installed
			if slices.Contains(installedAddons, args[i]) {
				remove(allAddons[args[i]])
			} else {
				ez.Kube.FmtYellow("%s is not installed", args[i])
			}
		}

		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		k8sutils := ez.CreateK8sUtilsImpl()
		clusterAddons, e := k8sutils.GetInstalledAddons()
		if e == nil {
			return clusterAddons, cobra.ShellCompDirectiveNoFileComp
		}

		return nil, cobra.ShellCompDirectiveNoFileComp
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
