package jsutils

import (
	"github.com/dop251/goja"
)

func (ctx *Easykube) WaitForDeployment(noop bool) func(goja.FunctionCall) goja.Value {
	if noop {
		return NoopFunc()
	}
	return ctx.waitForDeployment()
}

func (ctx *Easykube) waitForDeployment() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		ctx.checkArgs(call, WaitForDeployment)

		if ctx.ek.CommandContext.IsDryRun() {
			ctx.ek.Printer.FmtDryRun("skipping waitForDeployment")
			return call.This
		}

		deployment := call.Arguments[0].ToString().String()
		namespace := call.Arguments[1].ToString().String()

		err := ctx.ek.Kubernetes.WaitForDeploymentReadyWatch(deployment, namespace)
		if err != nil {
			ctx.ek.Printer.FmtRed("% did not come online before time out", deployment)
			return goja.Undefined()
		}

		return call.This
	}
}
