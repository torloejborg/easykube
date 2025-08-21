package jsutils

import (
	"github.com/dop251/goja"
	"github.com/torloj/easykube/pkg/constants"
	"github.com/torloj/easykube/pkg/ek"
	"time"
)

func (ctx *Easykube) Kustomize() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		out := ctx.EKContext.Printer
		k8sutils := ek.NewK8SUtils(ctx.EKContext)
		tools := ek.NewExternalTools(ctx.EKContext)

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
