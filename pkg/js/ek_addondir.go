package jsutils

import (
	"path/filepath"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
)

func (ctx *Easykube) AddonDir() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {

		base := filepath.Dir(ctx.AddonCtx.addon.GetAddonFile())

		if len(call.Arguments) == 1 {
			search := call.Arguments[0].String()
			addons, _ := ez.Kube.GetAddons()

			addon := addons[search]
			if addon == nil {
				ez.Kube.FmtYellow("No addons found for %s", search)
				return ctx.AddonCtx.vm.ToValue("")
			} else {
				other := filepath.Dir(addon.GetAddonFile())
				return ctx.AddonCtx.vm.ToValue(other)
			}
		}

		return ctx.AddonCtx.vm.ToValue(base)
	}
}
