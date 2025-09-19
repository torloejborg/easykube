package jsutils

import (
	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
)

func (ctx *Easykube) WaitForDeployment() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		ctx.checkArgs(call, WAIT_FOR_DEPLOYMENT)
		ezk := ez.Kube

		if ezk.IsDryRun() {
			ezk.FmtDryRun("skipping waitForDeployment")
			return call.This
		}

		deployment := call.Arguments[0].ToString().String()
		namespace := call.Arguments[1].ToString().String()

		err := ezk.WaitForDeploymentReadyWatch(deployment, namespace)
		if err != nil {
			ezk.FmtRed("% did not come online before time out", deployment)
			return goja.Undefined()
		}

		return call.This
	}
}
