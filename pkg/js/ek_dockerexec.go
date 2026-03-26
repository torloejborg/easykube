package jsutils

import (
	"strings"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
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
		ezk := ez.Kube

		container := call.Argument(0).String()
		command := ctx.extractStringSliceFromArgument(call.Argument(1))

		if ezk.IsVerbose() {
			ezk.FmtVerbose("docker exec %s %s", container, strings.Join(command, " "))
		}

		if ezk.IsDryRun() {
			ezk.FmtDryRun("skipping dockerExec")
			return call.This
		}

		ezk.Exec(container, command)

		return call.This
	}

}
