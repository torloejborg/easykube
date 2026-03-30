package jsutils

import (
	"github.com/dop251/goja"
	"github.com/google/uuid"
)

func (ctx *Easykube) NewUUID(noop bool) func(goja.FunctionCall) goja.Value {
	if noop {
		return NoopFunc()
	}

	return ctx.newUUID()
}

func (ctx *Easykube) newUUID() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {

		return ctx.AddonCtx.vm.ToValue(uuid.New().String())
	}
}
