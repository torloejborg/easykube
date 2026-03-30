package jsutils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dop251/goja"
)

func (ctx *Easykube) HelmTemplate(noop bool) func(goja.FunctionCall) goja.Value {
	if noop {
		return NoopFunc()
	}
	return ctx.helmTemplate()
}

func (ctx *Easykube) helmTemplate() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {

		addonDir := filepath.Dir(ctx.AddonCtx.addon.GetAddonFile())
		chart := filepath.Clean(call.Argument(0).String())

		if !filepath.IsAbs(chart) {
			chart = filepath.Join(addonDir, call.Argument(0).String())
		}

		values := filepath.Join(addonDir, call.Argument(1).String())
		destination := filepath.Join(addonDir, call.Argument(2).String())
		namespace := call.Argument(3).String()
		releasename := call.Argument(4).String()

		if !ctx.ek.Utils.FileOrDirExists(chart) {
			ctx.ek.Printer.FmtRed("specified chart %s does not exist", chart)
			os.Exit(-1)
		}

		if !ctx.ek.Utils.FileOrDirExists(values) {
			ctx.ek.Printer.FmtRed("the value file %s does not exist", values)
			os.Exit(-1)
		}

		if namespace == "" {
			namespace = "default"
		}

		cmd := "helm"
		var args = []string{}
		if len(releasename) > 0 {
			args = append(args, "template", chart,
				"--name-template", releasename,
				"--values", values,
				"--namespace", namespace)

		} else {
			args = append(args, "template", chart,
				"--values", values,
				"--namespace", namespace)

		}

		cmdStr := fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
		if ctx.ek.CommandContext.IsVerbose() {
			ctx.ek.Printer.FmtVerbose(cmdStr)
		}

		if ctx.ek.CommandContext.IsDryRun() {
			ctx.ek.Printer.FmtDryRun(cmdStr)
		} else {
			stdout, stderr, err := ctx.ek.ExternalTools.RunCommand(cmd, args...)

			if err != nil {
				ctx.ek.Printer.FmtRed("helm failed %s", stderr)
				os.Exit(-1)
			}

			ctx.ek.Utils.SaveFile(stdout, destination)
		}
		return call.This
	}
}
