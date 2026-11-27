package jsutils

import (
	"github.com/dop251/goja"
)

var CreateSecretDoc = []DocumentedFunction{{
	Name: CreateSecret, Description: "Creates a secret in a given namespace",
	Returns: "this",
	Params: []DocumentedParam{
		{Name: "namespace", Type: "string", Description: "target namespace for secret"},
		{Name: "name", Type: "string", Description: "name of secret"},
		{Name: "data", Type: "Map<string, string>", Description: "the secret as key/value pairs"},
	},
}}

func (ctx *Easykube) CreateSecret(noop bool) func(goja.FunctionCall) goja.Value {
	if noop {
		return NoopFunc()
	}
	return ctx.createSecret()
}

func (ctx *Easykube) createSecret() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {

		if ctx.ek.CommandContext.IsDryRun() {
			ctx.ek.Printer.FmtDryRun("skipping createSecret")
			return call.This
		}

		ctx.checkArgs(call, CreateSecret)
		namespace := call.Argument(0).String()
		name := call.Argument(1).String()
		secret := make(map[string]string)
		err := ctx.AddonCtx.vm.ExportTo(call.Argument(2), &secret)
		if err != nil {
			panic(err)
		}

		err = ctx.ek.Kubernetes.CreateSecret(namespace, name, secret)
		if err != nil {
			panic(err)
		}

		return call.This
	}
}
