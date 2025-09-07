package jsutils

import (
	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg"
	"github.com/torloejborg/easykube/pkg/ek"
)

func (ctx *Easykube) AndThenApply() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {

		out := ctx.EKContext.Printer
		toApply := call.Argument(0).String()
		u := ek.Utils{Fs: ctx.EKContext.Fs}

		if !u.FileOrDirExists(toApply) {
			out.FmtRed("could not locate %s to apply", toApply)
		}
		ctx.checkArgs(call, AND_THEN_APPLY)

		ext := pkg.CreateExternalTools()
		ext.ApplyYaml(toApply)

		return call.This
	}
}
