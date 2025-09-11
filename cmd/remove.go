package cmd

import (
	"os"
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
	ez.Kube.DeleteYaml(yamlFile)
	ez.Kube.FmtGreen("removed %s", addon.ShortName)
	ez.Kube.DeleteKeyFromConfigmap(constants.ADDON_CM, constants.DEFAULT_NS, addon.ShortName)

	err := ez.Kube.Remove(constants.KUSTOMIZE_TARGET_OUTPUT)
	if err != nil {
		ez.Kube.FmtYellow("%s could not be deleted", constants.KUSTOMIZE_TARGET_OUTPUT)
	}
}

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove [addon...]",
	Short: "removes a previously installed addon",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		// switch to the easykube context
		ez.Kube.EnsureLocalContext()

		allAddons, aerr := ez.Kube.GetAddons()
		if aerr != nil {
			ez.Kube.FmtRed("could not get addons %s", aerr.Error())
		}
		installedAddons, e := ez.Kube.GetInstalledAddons()
		if e != nil {
			ez.Kube.FmtRed("Cannot get installed addons: %v (was the configmap deleted by accident?)", e)
			os.Exit(1)
		}

		if len(args) == 0 {
			ez.Kube.FmtRed("Please specify one or more addons to remove, usage below\n")
			err := cmd.Help()
			if err != nil {
				// ignore
			}
			os.Exit(-1)
		}

		for i := range args {
			// is args[i] installed
			if slices.Contains(installedAddons, args[i]) {
				remove(allAddons[args[i]])
			} else {
				ez.Kube.FmtYellow("%s is not installed", args[i])
			}
		}
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
