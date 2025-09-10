package jsutils

import (
	"time"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"
)

func (ctx *Easykube) Kustomize() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {

		yamlFile := ez.Kube.KustomizeBuild(".")
		ez.Kube.ApplyYaml(yamlFile)

		ez.Kube.UpdateConfigMap(constants.ADDON_CM,
			constants.DEFAULT_NS,
			ctx.AddonCtx.addon.ShortName,
			[]byte(time.Now().String()))

		ez.Kube.FmtGreen("kustomize applied for %s", ctx.AddonCtx.addon.ShortName)

		return call.This
	}
}

func (ctx *Easykube) KustomizeWithOverlay() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {

		ez.Kube.FmtYellow("kustomize with overlay")
		overlay := call.Argument(0).String()

		ez.Kube.KustomizeBuild(overlay)
		//tools.ApplyYaml(yamlFile)

		ez.Kube.UpdateConfigMap(constants.ADDON_CM,
			constants.DEFAULT_NS,
			ctx.AddonCtx.addon.ShortName,
			[]byte(time.Now().String()))

		ez.Kube.FmtGreen("kustomize applied for %s using overlay %s", ctx.AddonCtx.addon.ShortName, overlay)

		return call.This
	}
}
