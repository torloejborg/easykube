package jsutils

import (
	"path/filepath"
	"time"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/constants"
)

func (ctx *Easykube) Kustomize(noop bool) func(goja.FunctionCall) goja.Value {
	if noop {
		return NoopFunc()
	}

	return ctx.kustomize()
}

func (ctx *Easykube) kustomize() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		yamlFile := ctx.ek.ExternalTools.KustomizeBuild(filepath.Dir(ctx.AddonCtx.addon.GetAddonFile()))

		ctx.ek.ExternalTools.ApplyYaml(yamlFile)

		if ctx.ek.CommandContext.IsDryRun() {
			return call.This
		} else {
			ctx.ek.Kubernetes.UpdateConfigMap(constants.AddonCm,
				constants.DefaultNs,
				ctx.AddonCtx.addon.GetShortName(),
				[]byte(time.Now().String()))

			ctx.ek.Printer.FmtGreen("kustomize applied for %s", ctx.AddonCtx.addon.GetShortName())
		}

		return call.This
	}
}
