package jsutils

import (
	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
)

func (ctx *Easykube) AndThenApply() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {

		toApply := call.Argument(0).String()

		if !ez.FileOrDirExists(toApply) {
			ez.Kube.FmtRed("could not locate %s to apply", toApply)
		}
		ctx.checkArgs(call, AND_THEN_APPLY)

		ext := ez.CreateExternalToolsImpl()
		ext.ApplyYaml(toApply)

		return call.This
	}
}
