package jsutils

import (
	"path/filepath"

	"github.com/dop251/goja"
)

func (ctx *Easykube) AndThenApply(noop bool) func(goja.FunctionCall) goja.Value {
	if noop {
		return NoopFunc()
	}

	return ctx.andThenApply()
}

func (ctx *Easykube) andThenApply() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {

		addonDir := filepath.Dir(ctx.AddonCtx.addon.GetAddonFile())
		toApply := filepath.Join(addonDir, call.Argument(0).String())

		if !ctx.ek.Utils.FileOrDirExists(toApply) {
			ctx.ek.Printer.FmtRed("could not locate %s to apply", toApply)
		}
		ctx.checkArgs(call, AndThenApply)

		ctx.ek.ExternalTools.ApplyYaml(toApply)

		return call.This
	}
}
