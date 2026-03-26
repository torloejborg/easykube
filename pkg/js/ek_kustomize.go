package jsutils

import (
	"path/filepath"
	"time"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"
)

func (ctx *Easykube) Kustomize(noop bool) func(goja.FunctionCall) goja.Value {
	if noop {
		return NoopFunc()
	}

	return ctx.kustomize()
}

func (ctx *Easykube) kustomize() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		ezk := ez.Kube
		yamlFile := ez.Kube.KustomizeBuild(filepath.Dir(ctx.AddonCtx.addon.GetAddonFile()))

		ezk.ApplyYaml(yamlFile)

		if ezk.IsDryRun() {
			return call.This
		} else {
			ezk.UpdateConfigMap(constants.AddonCm,
				constants.DefaultNs,
				ctx.AddonCtx.addon.GetShortName(),
				[]byte(time.Now().String()))
			ezk.FmtGreen("kustomize applied for %s", ctx.AddonCtx.addon.GetShortName())
		}

		return call.This
	}
}
