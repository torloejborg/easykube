package jsutils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
)

func (ctx *Easykube) HelmTemplate() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		ezk := ez.Kube
		addonDir := filepath.Dir(ctx.AddonCtx.addon.GetAddonFile())

		chart := call.Argument(0).String()
		if !strings.HasPrefix(chart, "/") {
			chart = filepath.Join(addonDir, call.Argument(0).String())
		}

		values := filepath.Join(addonDir, call.Argument(1).String())
		destination := filepath.Join(addonDir, call.Argument(2).String())
		namespace := call.Argument(3).String()

		if !ez.FileOrDirExists(chart) {
			ez.Kube.FmtRed("specified chart %s does not exist", chart)
			os.Exit(-1)
		}

		if !ez.FileOrDirExists(values) {
			ezk.FmtRed("the value file %s does not exist", values)
			os.Exit(-1)
		}

		if namespace == "" {
			namespace = "default"
		}

		cmd := "helm"
		args := []string{"template", chart,
			"--values", values,
			"--namespace", namespace}

		cmdStr := fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
		if ezk.IsVerbose() {
			ezk.FmtVerbose(cmdStr)
		}

		if ezk.IsDryRun() {
			ezk.FmtDryRun(cmdStr)
		} else {
			stdout, stderr, err := ezk.RunCommand(cmd, args...)

			if err != nil {
				ezk.FmtRed("helm failed %s", stderr)
				os.Exit(-1)
			}

			ez.SaveFile(stdout, destination)
		}
		return call.This
	}
}
