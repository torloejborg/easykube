package jsutils

import (
	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
)

func (ctx *Easykube) GetSecret() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		ctx.checkArgs(call, GET_SECRET)

		namespace := call.Argument(0).String()
		name := call.Argument(1).String()

		k8 := ez.CreateK8sUtilsImpl()
		res, err := k8.GetSecret(name, namespace)

		m := make(map[string]string)

		for k, v := range res {
			m[k] = string(v)
		}

		if err != nil {
			return goja.Undefined()
		} else {
			return ctx.AddonCtx.vm.ToValue(m)
		}
	}
}
