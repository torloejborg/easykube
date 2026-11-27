package jsutils

import (
	"time"

	"github.com/dop251/goja"
)

var WaitForCRDDoc = []DocumentedFunction{
	{Name: WaitForCrd, Description: "Waits until a custom resource definition appears (Operators exemplifies this issue, it can" +
		"\ntake some time for an operator to apply its own resources)", Returns: "this", Params: []DocumentedParam{
		{Name: "group", Description: "group name of crd", Type: "string", Optional: false},
		{Name: "version", Description: "version of the crd", Type: "string", Optional: false},
		{Name: "kind", Description: "type of crd", Type: "string", Optional: false},
		{Name: "timeout", Description: "seconds to wait before timeout", Type: "number", Optional: false},
	}},
}

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
