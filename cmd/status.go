package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"

	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "inspects you environment to see if prerequisites are met",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		ezk := ez.Kube
		cfg, _ := ezk.LoadConfig()

		if !ezk.IsContainerRuntimeAvailable() {
			ezk.FmtRed("Container runtime not available, is docker running??")
			os.Exit(-1)
		}

		hasBinary := func(name string) {
			_, err := exec.LookPath(name)
			if err != nil {
				ezk.FmtRed("⚠ " + name)
			} else {
				ezk.FmtGreen("✓ " + name)
			}
		}
		running := func(containerID string) {
			if ezk.IsContainerRunning(containerID) {
				ezk.FmtGreen("✓ %s container", containerID)
			} else {
				ezk.FmtRed("⚠ %s container not running", containerID)
			}
		}

		ezk.FmtGreen("Binaries")
		hasBinary("docker")
		hasBinary("helm")
		hasBinary("kustomize")

		fmt.Println()
		ezk.FmtGreen("Container configuration")
		running(constants.REGISTRY_CONTAINER)
		running(constants.KIND_CONTAINER)

		if ez.Kube.IsNetworkConnectedToContainer(constants.REGISTRY_CONTAINER, "kind") {
			ezk.FmtGreen("✓ %s connected to kind network", constants.REGISTRY_CONTAINER)
		} else {
			ezk.FmtRed("⚠ %s not connected to kind network", constants.REGISTRY_CONTAINER)
		}

		fmt.Println()
		ezk.FmtGreen("Repository configuration")

		addons, aerr := ez.Kube.GetAddons()
		if aerr != nil {
			ezk.FmtRed(aerr.Error())
		}

		na := len(addons)
		if _, err := os.Stat(cfg.AddonDir); err == nil {
			if na == 0 {
				ezk.FmtYellow("⚠ %d Addons discovered, check if '%s' is an addon repository", na, cfg.AddonDir)
			} else {
				ezk.FmtGreen("✓ %d Addons discovered at '%s'", na, cfg.AddonDir)
			}

		} else {
			ezk.FmtRed("⚠ Addon directory '%s' does not exist, check your config", cfg.AddonDir)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
