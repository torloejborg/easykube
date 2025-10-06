package jsutils

import (
	"fmt"
	"strings"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
)

func (ctx *Easykube) ExecInContainer() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		ezk := ez.Kube

		ctx.checkArgs(call, EXEC_IN_CONTAINER)

		deployment := call.Argument(0).String()
		namespace := call.Argument(1).String()
		command := call.Argument(2).String()
		args := ctx.extractStringSliceFromArgument(call.Argument(3))
		infostr := fmt.Sprintf("docker exec (in %s) %s %s ", deployment, command, strings.Join(args, " "))
		if ezk.IsDryRun() {
			ezk.FmtDryRun(infostr)
			return goja.Undefined()
		}

		pods, _ := ezk.ListPods(namespace)
		for i := range pods {
			if strings.Contains(pods[i], deployment) {

				if ezk.IsVerbose() {
					ezk.FmtVerbose(infostr)
				}
				stdout, stderr, err := ez.Kube.ExecInPod(namespace, pods[i], command, args)

				if err != nil {
					ezk.FmtRed(stderr)
					ezk.FmtRed(err.Error())
					return ctx.AddonCtx.vm.ToValue(stderr)
				}

				return ctx.AddonCtx.vm.ToValue(stdout)
			}
		}

		return goja.Undefined()

	}
}
