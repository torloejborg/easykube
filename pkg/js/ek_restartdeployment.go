package jsutils

import (
	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
)

func (ctx *Easykube) RestartDeployment(noop bool) func(goja.FunctionCall) goja.Value {
	if noop {
		return NoopFunc()
	}

	return ctx.restartDeployment()

}

func (ctx *Easykube) restartDeployment() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {

		deployment := call.Argument(0).String()
		namespace := call.Argument(1).String()

		err := ez.Kube.RestartDeployment(deployment, namespace)
		if err != nil {
			panic(err)
		}

		return call.This
	}
}
