package jsutils

import (
	"path/filepath"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
)

func (ctx *Easykube) AndThenApply(noop bool) func(goja.FunctionCall) goja.Value {
	if noop {
		return NoopFunc()
	}

	return ctx.andThenApply()
}

func (ctx *Easykube) andThenApply() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		ezk := ez.Kube
		addonDir := filepath.Dir(ctx.AddonCtx.addon.GetAddonFile())
		toApply := filepath.Join(addonDir, call.Argument(0).String())

		if !ez.FileOrDirExists(toApply) {
			ezk.FmtRed("could not locate %s to apply", toApply)
		}
		ctx.checkArgs(call, AndThenApply)

		ezk.ApplyYaml(toApply)

		return call.This
	}
}
