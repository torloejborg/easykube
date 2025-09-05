package jsutils

import (
	"time"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ek"
)

func (ctx *Easykube) WaitForCRD() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		ctx.checkArgs(call, WAIT_FOR_CRD)

		group := call.Argument(0).String()
		version := call.Argument(1).String()
		kind := call.Argument(2).String()
		timeout := call.Argument(3).ToInteger()

		k8sutils := ek.NewK8SUtils(ctx.EKContext)
		err := k8sutils.WaitForCRD(group, version, kind, time.Duration(timeout)*time.Second)
		if err != nil {
			panic(err)
		}

		return call.This
	}
}
