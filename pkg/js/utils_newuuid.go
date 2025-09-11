package jsutils

import (
	"github.com/dop251/goja"
	"github.com/google/uuid"
)

func (ctx *Easykube) NewUUID() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {

		return ctx.AddonCtx.vm.ToValue(uuid.New().String())
	}
}
