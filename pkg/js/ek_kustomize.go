package jsutils

import (
	"time"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg"
	"github.com/torloejborg/easykube/pkg/constants"
)

func (ctx *Easykube) Kustomize() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		out := ctx.EKContext.Printer
		k8sutils := pkg.CreateK8sUtils()
		tools := pkg.CreateExternalTools()

		yamlFile := tools.KustomizeBuild(".")
		tools.ApplyYaml(yamlFile)

		k8sutils.UpdateConfigMap(constants.ADDON_CM,
			constants.DEFAULT_NS,
			ctx.AddonCtx.addon.ShortName,
			[]byte(time.Now().String()))

		out.FmtGreen("kustomize applied for %s", ctx.AddonCtx.addon.ShortName)

		return call.This
	}
}

func (ctx *Easykube) KustomizeWithOverlay() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		out := ctx.EKContext.Printer
		k8sutils := pkg.CreateK8sUtils()
		tools := pkg.CreateExternalTools()

		out.FmtYellow("kustomize with overlay")
		overlay := call.Argument(0).String()

		tools.KustomizeBuild(overlay)
		//tools.ApplyYaml(yamlFile)

		k8sutils.UpdateConfigMap(constants.ADDON_CM,
			constants.DEFAULT_NS,
			ctx.AddonCtx.addon.ShortName,
			[]byte(time.Now().String()))

		out.FmtGreen("kustomize applied for %s using overlay %s", ctx.AddonCtx.addon.ShortName, overlay)

		return call.This
	}
}
