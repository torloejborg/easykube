package jsutils

import (
	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
)

func (ctx *Easykube) WaitForDeployment() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {

		ctx.checkArgs(call, WAIT_FOR_DEPLOYMENT)

		deployment := call.Arguments[0].ToString().String()
		namespace := call.Arguments[1].ToString().String()

		err := ez.Kube.WaitForDeploymentReadyWatch(deployment, namespace)
		if err != nil {
			ez.Kube.FmtRed("% did not come online before timed out", deployment)
			return goja.Undefined()
		}

		return call.This
	}
}
