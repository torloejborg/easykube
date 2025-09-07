package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/torloejborg/easykube/pkg"
	"github.com/torloejborg/easykube/pkg/constants"

	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/ekctx"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "inspects you environment to see if prerequisites are met",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		ekCtx := ekctx.GetAppContext(cmd)
		out := ekCtx.Printer
		cru := pkg.CreateContainerRuntime()

		conf := pkg.CreateEasykubeConfig()
		cfg, _ := conf.LoadConfig()
		ar := pkg.CreateAddonReader()
		if !cru.IsContainerRuntimeAvailable() {
			out.FmtRed("Container runtime not available, is docker running??")
			os.Exit(-1)
		}

		hasBinary := func(name string) {
			_, err := exec.LookPath(name)
			if err != nil {
				out.FmtRed("⚠ " + name)
			} else {
				out.FmtGreen("✓ " + name)
			}
		}
		running := func(containerID string) {
			if cru.IsContainerRunning(containerID) {
				out.FmtGreen("✓ %s container", containerID)
			} else {
				out.FmtRed("⚠ %s container not running", containerID)
			}
		}

		out.FmtGreen("Binaries")
		hasBinary("docker")
		hasBinary("helm")
		hasBinary("kustomize")

		fmt.Println()
		out.FmtGreen("Container configuration")
		running(constants.REGISTRY_CONTAINER)
		running(constants.KIND_CONTAINER)

		if cru.IsNetworkConnectedToContainer(constants.REGISTRY_CONTAINER, "kind") {
			out.FmtGreen("✓ %s connected to kind network", constants.REGISTRY_CONTAINER)
		} else {
			out.FmtRed("⚠ %s not connected to kind network", constants.REGISTRY_CONTAINER)
		}

		fmt.Println()
		out.FmtGreen("Repository configuration")
		na := len(ar.GetAddons())
		if _, err := os.Stat(cfg.AddonDir); err == nil {
			if na == 0 {
				out.FmtYellow("⚠ %d Addons discovered, check if '%s' is an addon repository", na, cfg.AddonDir)
			} else {
				out.FmtGreen("✓ %d Addons discovered at '%s'", na, cfg.AddonDir)
			}

		} else {
			out.FmtRed("⚠ Addon directory '%s' does not exist, check your config", cfg.AddonDir)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
