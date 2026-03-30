package jsutils

import (
	"fmt"
	"strings"

	"github.com/dop251/goja"
)

func (ctx *Easykube) ExecInContainer(noop bool) func(goja.FunctionCall) goja.Value {
	if noop {

		er := &ExecResult{runtime: ctx.AddonCtx.vm}
		obj := ctx.AddonCtx.NewObject()
		er.self = obj

		// bind methods
		_ = obj.Set("onSuccess", NoopFunc)
		_ = obj.Set("onFail", NoopFunc)

		return NoopFunc()
	}

	return ctx.execInContainer()
}

func (ctx *Easykube) execInContainer() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		ctx.checkArgs(call, ExecInContainer)

		deployment := call.Argument(0).String()
		namespace := call.Argument(1).String()
		command := call.Argument(2).String()
		args := ctx.extractStringSliceFromArgument(call.Argument(3))
		infostr := fmt.Sprintf("docker exec (in %s) %s %s ", deployment, command, strings.Join(args, " "))

		er := &ExecResult{runtime: ctx.AddonCtx.vm}
		obj := ctx.AddonCtx.NewObject()
		er.self = obj

		// bind methods
		_ = obj.Set("onSuccess", er.OnSuccess)
		_ = obj.Set("onFail", er.OnFail)

		if ctx.ek.CommandContext.IsDryRun() {
			ctx.ek.Printer.FmtDryRun(infostr)
			er.success = true
			return obj
		}

		pods, _ := ctx.ek.Kubernetes.ListPods(namespace)
		for i := range pods {
			if strings.Contains(pods[i], deployment) {

				if ctx.ek.CommandContext.IsVerbose() {
					ctx.ek.Printer.FmtVerbose(infostr)
				}

				stdout, stderr, err := ctx.ek.Kubernetes.ExecInPod(namespace, pods[i], command, args)

				if err != nil {
					er.output = stdout + stderr + err.Error()
					er.success = false

				} else {
					er.output = stdout + stderr
					er.success = true
				}

				return obj

			}
		}

		return goja.Undefined()

	}
}
