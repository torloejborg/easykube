package jsutils

import (
	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg"
)

func (ctx *Easykube) WaitForDeployment() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		out := ctx.EKContext.Printer
		ctx.checkArgs(call, WAIT_FOR_DEPLOYMENT)

		deployment := call.Arguments[0].ToString().String()
		namespace := call.Arguments[1].ToString().String()

		err := pkg.CreateK8sUtils().WaitForDeploymentReadyWatch(deployment, namespace)
		if err != nil {
			out.FmtRed("% did not come online before timed out", deployment)
			return goja.Undefined()
		}

		return call.This
	}
}
