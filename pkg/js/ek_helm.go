package jsutils

import (
	"github.com/dop251/goja"
	"github.com/torloj/easykube/pkg/ek"
	"os"
)

func (ctx *Easykube) HelmTemplate() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {

		out := ctx.EKContext.Printer

		chart := call.Argument(0).String()
		values := call.Argument(1).String()
		destination := call.Argument(2).String()

		if !ek.FileOrDirExists(chart) {
			out.FmtRed("specified chart %s does not exist", chart)
			os.Exit(-1)
		}

		if !ek.FileOrDirExists(values) {
			out.FmtRed("the value file %s does not exist", values)
			os.Exit(-1)
		}

		tools := ek.NewExternalTools(ctx.EKContext)
		stdout, stderr, err := tools.RunCommand("helm", "template", chart, "--values", values)

		if err != nil {
			out.FmtRed("helm failed %s", stderr)
			os.Exit(-1)
		}

		ek.SaveFile(stdout, destination)

		return call.This
	}
}
