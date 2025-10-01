package jsutils

import (
	"strings"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
)

func (ctx *Easykube) DockerExec() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		ctx.checkArgs(call, COPY_TO)
		ezk := ez.Kube

		container := call.Argument(0).String()
		//command := call.Argument(1).String()
		command, _ := exportStringArray(call.Argument(1).Export())

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
