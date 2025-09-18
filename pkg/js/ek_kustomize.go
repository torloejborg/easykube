package jsutils

import (
	"time"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"
)

func (ctx *Easykube) Kustomize() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		ezk := ez.Kube

		yamlFile := ez.Kube.KustomizeBuild(".")
		ezk.ApplyYaml(yamlFile)

		ezk.UpdateConfigMap(constants.ADDON_CM,
			constants.DEFAULT_NS,
			ctx.AddonCtx.addon.ShortName,
			[]byte(time.Now().String()))

		ezk.FmtGreen("kustomize applied for %s", ctx.AddonCtx.addon.ShortName)

		return call.This
	}
}

func (ctx *Easykube) KustomizeWithOverlay() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		ezk := ez.Kube
		ezk.FmtYellow("kustomize with overlay")
		overlay := call.Argument(0).String()

		ezk.KustomizeBuild(overlay)
		//tools.ApplyYaml(yamlFile)

		ezk.UpdateConfigMap(constants.ADDON_CM,
			constants.DEFAULT_NS,
			ctx.AddonCtx.addon.ShortName,
			[]byte(time.Now().String()))

		ezk.FmtGreen("kustomize applied for %s using overlay %s", ctx.AddonCtx.addon.ShortName, overlay)

		return call.This
	}
}
