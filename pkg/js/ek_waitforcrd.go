package jsutils

import (
	"time"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
)

func (ctx *Easykube) WaitForCRD() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		ezk := ez.Kube
		if ezk.IsDryRun() {
			ezk.FmtDryRun("skipping waitForCRD")
		}
		ctx.checkArgs(call, WAIT_FOR_CRD)

		group := call.Argument(0).String()
		version := call.Argument(1).String()
		kind := call.Argument(2).String()
		timeout := call.Argument(3).ToInteger()

		err := ezk.WaitForCRD(group, version, kind, time.Duration(timeout)*time.Second)
		if err != nil {
			panic(err)
		}

		return call.This
	}
}
