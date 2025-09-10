package jsutils

import (
	"os"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
)

func (ctx *Easykube) HelmTemplate() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {

		out := ctx.EKContext.Printer

		chart := call.Argument(0).String()
		values := call.Argument(1).String()
		destination := call.Argument(2).String()
		namespace := call.Argument(3).String()

		if !ez.FileOrDirExists(chart) {
			out.FmtRed("specified chart %s does not exist", chart)
			os.Exit(-1)
		}

		if !ez.FileOrDirExists(values) {
			out.FmtRed("the value file %s does not exist", values)
			os.Exit(-1)
		}

		if namespace == "" {
			namespace = "default"
		}

		stdout, stderr, err := ez.Kube.RunCommand("helm", "template", chart,
			"--values", values,
			"--namespace", namespace)

		if err != nil {
			out.FmtRed("helm failed %s", stderr)
			os.Exit(-1)
		}

		ez.SaveFile(stdout, destination)

		return call.This
	}
}
