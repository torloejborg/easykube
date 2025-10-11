package jsutils

import (
	"path/filepath"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
)

func (ctx *Easykube) AndThenApply() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		ezk := ez.Kube
		addonDir := filepath.Dir(ctx.AddonCtx.addon.File)
		toApply := filepath.Join(addonDir, call.Argument(0).String())

		if !ez.FileOrDirExists(toApply) {
			ezk.FmtRed("could not locate %s to apply", toApply)
		}
		ctx.checkArgs(call, AND_THEN_APPLY)

		ezk.ApplyYaml(toApply)

		return call.This
	}
}
