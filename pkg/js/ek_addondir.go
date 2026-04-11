package jsutils

import (
	"path/filepath"

	"github.com/dop251/goja"
)

func (ctx *Easykube) AddonDir(noop bool) func(goja.FunctionCall) goja.Value {
	if noop {
		return NoopFunc()
	}
	return ctx.addonDir()
}

func (ctx *Easykube) addonDir() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {

		base := filepath.Dir(ctx.AddonCtx.addon.GetAddonFile())

		if len(call.Arguments) == 1 {
			search := call.Arguments[0].String()
			addons, _ := ctx.ek.AddonReader.GetAddons()

			addon := addons[search]
			if addon == nil {
				ctx.ek.Printer.FmtYellow("No addons found for %s", search)
				return ctx.AddonCtx.vm.ToValue("")
			} else {
				other := filepath.Dir(addon.GetAddonFile())
				return ctx.AddonCtx.vm.ToValue(other)
			}
		}

		return ctx.AddonCtx.vm.ToValue(base)
	}
}
