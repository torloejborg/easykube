package jsutils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"
)

func (ctx *Easykube) KeyValue() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		ctx.checkArgs(call, KEY_VALUE)
		inputStr := ctx.CobraCommandHelder.GetStringFlag(constants.FLAG_KEYVALUE)
		if inputStr == "" {
			return goja.Undefined()
		}

		key := call.Argument(0).String()
		kvmap, err := parseKVPairs(inputStr)

		if err != nil {
			ez.Kube.FmtYellow("Failed to parse key-value pairs: %s ,input was %s", err.Error(), inputStr)
			return goja.Undefined()
		}

		if kvmap[key] == "" {
			return goja.Undefined()
		} else {
			return ctx.AddonCtx.vm.ToValue(kvmap[key])
		}
	}
}

func parseKVPairs(input string) (map[string]string, error) {
	kvMap := make(map[string]string)

	// Remove extra spaces around commas
	input = strings.TrimSpace(input)
	input = regexp.MustCompile(`\s*,\s*`).ReplaceAllString(input, ",")

	// Split by comma to get individual key-value pairs
	pairs := strings.Split(input, ",")

	for _, pair := range pairs {
		// Split each pair by the first '=' found (ignoring spaces)
		keyValue := strings.FieldsFunc(pair, func(r rune) bool {
			return r == '=' && !strings.ContainsRune(strings.TrimSpace(string([]rune(pair)[0:strings.IndexRune(pair, r)])), ' ')
		})

		if len(keyValue) != 2 {
			return nil, fmt.Errorf("invalid key-value pair format: %s", pair)
		}

		key := strings.TrimSpace(keyValue[0])
		value := strings.TrimSpace(strings.Join(keyValue[1:], "="))

		kvMap[key] = value
	}

	return kvMap, nil
}
