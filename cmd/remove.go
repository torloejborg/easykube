package cmd

import (
	"path/filepath"

	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"

	"github.com/spf13/cobra"
)

func remove(addon *ez.Addon) {
	// enter the addon directory
	addonDir := filepath.Dir(addon.File)
	yamlFile := ez.Kube.KustomizeBuild(addonDir)
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

		opts := RemoveOpts{AddonsToRemove: args}

		return removeActual(opts)
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
