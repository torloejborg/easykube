package jsutils

import (
	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
)

func (ctx *Easykube) Config() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {

		cfg, err := ez.Kube.LoadConfig()
		if err != nil {
			panic(err)
		}

		return ctx.AddonCtx.vm.ToValue(cfg)
	}
}
