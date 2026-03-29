package jsutils

import (
	"github.com/dop251/goja"
)

func (ctx *Easykube) Config(noop bool) func(goja.FunctionCall) goja.Value {
	if noop {
		return NoopFunc()
	}

	return ctx.config()
}

func (ctx *Easykube) config() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {

		cfg, err := ctx.ek.Config.LoadConfig()
		if err != nil {
			panic(err)
		}

		return ctx.AddonCtx.vm.ToValue(cfg)
	}
}
