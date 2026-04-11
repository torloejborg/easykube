package jsutils

import (
	"fmt"
	"os"
	"reflect"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/core"
)

type Easykube struct {
	AddonCtx *AddonContext
	ek       *core.Ek
}

func ConfigureEasykubeScript(ek *core.Ek, addon *AddonContext) {
	check := func(e error) {
		if e != nil {
			panic(e)
		}
	}

	// tag::export[]
	e := &Easykube{ek: ek, AddonCtx: addon}
	isNoop := addon.IsNoop
	easykubeObj := addon.NewObject()
	check(easykubeObj.Set(Kustomize, e.Kustomize(isNoop)))
	check(easykubeObj.Set(WaitForDeployment, e.WaitForDeployment(isNoop)))
	check(easykubeObj.Set(AndThenApply, e.AndThenApply(isNoop)))
	check(easykubeObj.Set(ExecInContainer, e.ExecInContainer(isNoop)))
	check(easykubeObj.Set(Preload, e.PreloadImages(isNoop)))
	check(easykubeObj.Set(WaitForCrd, e.WaitForCRD(isNoop)))
	check(easykubeObj.Set(CopyTo, e.CopyTo(isNoop)))
	check(easykubeObj.Set(CreateSecret, e.CreateSecret(isNoop)))
	check(easykubeObj.Set(GetSecret, e.GetSecret(isNoop)))
	check(easykubeObj.Set(GitSparseCheckout, e.GitSparseCheckout(isNoop)))
	check(easykubeObj.Set(GitCheckout, e.GitCheckout(isNoop)))
	check(easykubeObj.Set(HelmTemplate, e.HelmTemplate(isNoop)))
	check(easykubeObj.Set(ProcessSecrets, e.ProcessExternalSecrets(isNoop)))
	check(easykubeObj.Set(KeyValue, e.KeyValue(isNoop)))
	check(easykubeObj.Set(KeyEnv, e.Env(isNoop)))
	check(easykubeObj.Set(Http, e.Http(isNoop)))
	check(easykubeObj.Set(Exec, e.Exec(isNoop)))
	check(easykubeObj.Set(DockerExec, e.DockerExec(isNoop)))
	check(easykubeObj.Set(AddonDir, e.AddonDir(isNoop)))
	check(easykubeObj.Set(Config, e.Config(isNoop)))
	check(easykubeObj.Set(RestartDeployment, e.Config(isNoop)))
	check(easykubeObj.Set(SkopeoPreload, e.SkopeoPreload(isNoop)))

	addon.ExportFunction("_ek", easykubeObj)

	utilsObj := addon.NewObject()
	check(utilsObj.Set("UUID", e.NewUUID(isNoop)))
	addon.ExportFunction("_utils", utilsObj)
	// end::export[]
}

func (e *Easykube) checkArgs(f goja.FunctionCall, jsName string) {
	argLen := len(f.Arguments)
	var undef = 0
	for v := range f.Arguments {
		if f.Arguments[v] == goja.Undefined() {
			undef = undef + 1
		}
	}

	if undef != 0 {
		e.ek.Printer.FmtRed("check addon %s, Call to %s expected %d arguments, %d is missing",
			e.AddonCtx.addon.GetName(),
			jsName,
			argLen,
			undef)

		os.Exit(-1)
	}
}

func (e *Easykube) extractStringSliceFromArgument(arg goja.Value) []string {

	// Ensure it's an array
	if arg.ExportType().Kind() != reflect.Slice {
		panic(e.AddonCtx.vm.ToValue("Expected an array"))
	}

	// Convert to []interface{} first
	ifaceArray := arg.Export().([]interface{})
	strings := make([]string, len(ifaceArray))

	for i, v := range ifaceArray {
		if s, ok := v.(string); ok {
			strings[i] = s
		} else {
			panic(e.AddonCtx.vm.ToValue(fmt.Sprintf("Element at index %d is not a string", i)))
		}
	}

	return strings
}

func NoopFunc() func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		return goja.Undefined()
	}
}
