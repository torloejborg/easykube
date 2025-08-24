package jsutils

import (
	"os"

	"github.com/dop251/goja"
)

func (ctx *Easykube) Env() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		ctx.checkArgs(call, KEY_VALUE)
		key := call.Argument(0).String()

		val := os.Getenv(key)

		if val == "" {
			return goja.Undefined()
		} else {
			return ctx.AddonCtx.vm.ToValue(val)
		}

	}
}
