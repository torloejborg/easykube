package jsutils

import (
	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ek"
)

func (ctx *Easykube) CreateSecret() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		ctx.checkArgs(call, CREATE_SECRET)
		namespace := call.Argument(0).String()
		name := call.Argument(1).String()
		secret := make(map[string]string)
		ctx.AddonCtx.vm.ExportTo(call.Argument(2), &secret)

		k8 := ek.NewK8SUtils(ctx.EKContext)
		k8.CreateSecret(namespace, name, secret)

		return call.This
	}
}
