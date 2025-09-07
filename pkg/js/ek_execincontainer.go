package jsutils

import (
	"fmt"
	"strings"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
)

func (ctx *Easykube) ExecInContainer() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		ctx.checkArgs(call, EXEC_IN_CONTAINER)

		out := ctx.EKContext.Printer

		deployment := call.Argument(0).String()
		namespace := call.Argument(1).String()
		command := call.Argument(2).String()
		args, _ := exportStringArray(call.Argument(3).Export())

		pods, _ := ez.Kube.ListPods(namespace)
		for i := range pods {
			if strings.Contains(pods[i], deployment) {
				stdout, stderr, err := ez.Kube.ExecInPod(namespace, pods[i], command, args)
				if err != nil {
					out.FmtRed(stderr)
					out.FmtRed(err.Error())
					return ctx.AddonCtx.vm.ToValue(stderr)
				}

				return ctx.AddonCtx.vm.ToValue(stdout)
			}
		}

		return goja.Undefined()

	}
}

func exportStringArray(val interface{}) ([]string, error) {
	items, ok := val.([]interface{})
	if !ok {
		return nil, fmt.Errorf("not an array")
	}

	result := make([]string, 0, len(items))
	for _, item := range items {
		str, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("element is not a string: %v", item)
		}
		result = append(result, str)
	}
	return result, nil
}
