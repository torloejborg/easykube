package jsutils

import (
	"github.com/dop251/goja"
)

func (ctx *Easykube) GetSecret(noop bool) func(goja.FunctionCall) goja.Value {
	if noop {
		return NoopFunc()
	}
	return ctx.getSecret()
}

func (ctx *Easykube) getSecret() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		ctx.checkArgs(call, GetSecret)

		namespace := call.Argument(0).String()
		name := call.Argument(1).String()

		res, err := ctx.ek.Kubernetes.GetSecret(name, namespace)

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
