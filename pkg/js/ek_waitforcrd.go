package jsutils

import (
	"time"

	"github.com/dop251/goja"
)

func (ctx *Easykube) WaitForCRD(noop bool) func(goja.FunctionCall) goja.Value {
	if noop {
		return NoopFunc()
	}
	return ctx.waitForCRD()
}

func (ctx *Easykube) waitForCRD() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if ctx.ek.CommandContext.IsDryRun() {
			ctx.ek.Printer.FmtDryRun("skipping waitForCRD")
		}
		ctx.checkArgs(call, WaitForCrd)

		group := call.Argument(0).String()
		version := call.Argument(1).String()
		kind := call.Argument(2).String()
		timeout := call.Argument(3).ToInteger()

		err := ctx.ek.Kubernetes.WaitForCRD(group, version, kind, time.Duration(timeout)*time.Second)
		if err != nil {
			panic(err)
		}

		return call.This
	}
}
