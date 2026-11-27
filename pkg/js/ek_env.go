package jsutils

import (
	"os"

	"github.com/dop251/goja"
)

var EkDocEnv = []DocumentedFunction{
	{
		Name:        KeyEnv,
		Description: "Fetch the value of an environment variable",
		Returns:     "string",
		Params: []DocumentedParam{
			{Name: "key", Description: "name of environment variable", Type: "string"},
		},
	},
}

// tag::envfunc[]
func (ctx *Easykube) Env(noop bool) func(goja.FunctionCall) goja.Value {
	if noop {
		return NoopFunc()
	}
	return ctx.env()
}

func (ctx *Easykube) env() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		ctx.checkArgs(call, KeyValue)
		key := call.Argument(0).String()
		val := os.Getenv(key)

		if val == "" {
			return goja.Null()
		} else {
			return ctx.AddonCtx.vm.ToValue(val)
		}
	}
}

// end::envfunc[]
