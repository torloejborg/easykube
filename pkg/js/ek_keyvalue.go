package jsutils

import (
	"strings"

	"github.com/dop251/goja"
	"github.com/torloj/easykube/pkg/constants"
)

func (ctx *Easykube) KeyValue() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		ctx.checkArgs(call, KEY_VALUE)
		key := call.Argument(0).String()
		kvmap := parseKeyValuePairs(ctx.EKContext.GetStringFlag(constants.FLAG_KEYVALUE))

		if kvmap[key] == "" {
			return goja.Undefined()
		} else {
			return ctx.AddonCtx.vm.ToValue(kvmap[key])
		}

	}

}

func parseKeyValuePairs(args string) map[string]string {
	result := make(map[string]string)
	if len(args) == 0 {
		return result
	}

	// the pairs
	pairs := strings.Split(args, ",")
	for x := range pairs {
		items := strings.Split(pairs[x], "=")
		result[items[0]] = strings.TrimSpace(items[1])
	}

	return result
}
