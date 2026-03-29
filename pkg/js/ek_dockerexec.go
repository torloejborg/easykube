package jsutils

import (
	"strings"

	"github.com/dop251/goja"
)

func (ctx *Easykube) DockerExec(noop bool) func(goja.FunctionCall) goja.Value {
	if noop {
		return NoopFunc()
	}
	return ctx.dockerExec()
}

func (ctx *Easykube) dockerExec() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		ctx.checkArgs(call, CopyTo)

		container := call.Argument(0).String()
		command := ctx.extractStringSliceFromArgument(call.Argument(1))

		if ctx.ek.CommandContext.IsVerbose() {
			ctx.ek.Printer.FmtVerbose("docker exec %s %s", container, strings.Join(command, " "))
		}

		if ctx.ek.CommandContext.IsDryRun() {
			ctx.ek.Printer.FmtDryRun("skipping dockerExec")
			return call.This
		}

		ctx.ek.ContainerRuntime.Exec(container, command)

		return call.This
	}

}
