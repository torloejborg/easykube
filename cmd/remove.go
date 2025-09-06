package cmd

import (
	"os"
	"path/filepath"
	"slices"

	"github.com/torloejborg/easykube/ekctx"
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ek"

	"github.com/spf13/cobra"
)

func remove(addon *ek.Addon, ctx *ekctx.EKContext, k8s ek.IK8SUtils) {
	// enter the addon directory
	ek.PushDir(filepath.Dir(addon.File.Name()))
	defer ek.PopDir()

	tools := ek.NewExternalTools(ctx)
	yamlFile := tools.KustomizeBuild(".")
	tools.DeleteYaml(yamlFile)
	ctx.Printer.FmtGreen("removed %s", addon.ShortName)
	k8s.DeleteKeyFromConfigmap(constants.ADDON_CM, constants.DEFAULT_NS, addon.ShortName)

	err := os.Remove(constants.KUSTOMIZE_TARGET_OUTPUT)
	if err != nil {
		ctx.Printer.FmtYellow("%s could not be deleted", constants.KUSTOMIZE_TARGET_OUTPUT)
	}
}

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove [addon...]",
	Short: "removes a previously installed addon",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		ekCtx := ekctx.GetAppContext(cmd)
		out := ekCtx.Printer

		// switch to the easykube context
		ek.NewExternalTools(ekCtx).EnsureLocalContext()

		k8s := ek.NewK8SUtils(ekCtx)
		allAddons := ek.NewAddonReader(ekCtx).GetAddons()
		installedAddons, e := k8s.GetInstalledAddons()
		if e != nil {
			out.FmtRed("Cannot get installed addons: %v (was the configmap deleted by accident?)", e)
			os.Exit(1)
		}

		if len(args) == 0 {
			out.FmtRed("Please specify one or more addons to remove, usage below\n")
			err := cmd.Help()
			if err != nil {
				// ignore
			}
			os.Exit(-1)
		}

		for i := range args {
			// is args[i] installed
			if slices.Contains(installedAddons, args[i]) {
				remove(allAddons[args[i]], ekCtx, k8s)
			} else {
				out.FmtYellow("%s is not installed", args[i])
			}
		}
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		k8sutils := ek.NewK8SUtils(ekctx.GetAppContext(cmd))
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
