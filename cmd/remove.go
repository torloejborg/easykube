package cmd

import (
	"os"
	"path/filepath"
	"slices"

	"github.com/torloj/easykube/ekctx"
	"github.com/torloj/easykube/pkg/constants"
	"github.com/torloj/easykube/pkg/ek"

	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "removes a previously installed addon",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		ekCtx := ekctx.GetAppContext(cmd)
		out := ekCtx.Printer

		// switch to the easykube context
		ek.NewExternalTools(ekCtx).EnsureLocalContext()

		k8sutils := ek.NewK8SUtils(ekCtx)
		clusterAddons := k8sutils.GetInstalledAddons()

		if len(args) == 0 {
			out.FmtRed("Please specify an addon to remove, usage below\n")
			err := cmd.Help()
			if err != nil {
				// ignore
			}
			os.Exit(-1)
		}
		target := args[0]
		addons := ek.NewAddonReader(ekCtx).GetAddons()

		if slices.Contains(clusterAddons, target) {

			// enter the addon directory
			ek.PushDir(filepath.Dir(addons[target].File.Name()))
			defer ek.PopDir()

			tools := ek.NewExternalTools(ekCtx)
			yamlFile := tools.KustomizeBuild(".")
			tools.DeleteYaml(yamlFile)
			out.FmtGreen("removed %s", target)
			k8sutils.DeleteKeyFromConfigmap(constants.ADDON_CM, constants.DEFAULT_NS, target)

			err := os.Remove(constants.KUSTOMIZE_TARGET_OUTPUT)
			if err != nil {
				out.FmtYellow("%s could not be deleted", constants.KUSTOMIZE_TARGET_OUTPUT)
			}
		} else {
			out.FmtYellow("%s not applied", target)
		}
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
