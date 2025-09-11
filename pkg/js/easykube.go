package jsutils

import (
	"fmt"
	"os"
	"reflect"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
)

type Easykube struct {
	EKContext *ez.CobraCommandHelperImpl
	AddonCtx  *AddonContext
}

func ConfigureEasykubeScript(ctx *ez.CobraCommandHelperImpl, addon *AddonContext) {
	check := func(e error) {
		if e != nil {
			panic(e)
		}
	}

	e := &Easykube{EKContext: ctx, AddonCtx: addon}

	easykubeObj := addon.NewObject()
	check(easykubeObj.Set(KUSTOMIZE, e.Kustomize()))
	check(easykubeObj.Set(KUSTOMIZE_WITH_OVERLAY, e.KustomizeWithOverlay()))
	check(easykubeObj.Set(WAIT_FOR_DEPLOYMENT, e.WaitForDeployment()))
	check(easykubeObj.Set(AND_THEN_APPLY, e.AndThenApply()))
	check(easykubeObj.Set(EXEC_IN_CONTAINER, e.ExecInContainer()))
	check(easykubeObj.Set(PRELOAD, e.PreloadImages()))
	check(easykubeObj.Set(WAIT_FOR_CRD, e.WaitForCRD()))
	check(easykubeObj.Set(COPY_TO, e.CopyTo()))
	check(easykubeObj.Set(CREATE_SECRET, e.CreateSecret()))
	check(easykubeObj.Set(GET_SECRET, e.GetSecret()))
	check(easykubeObj.Set(SPARSE_CHECKOUT, e.GitSparseCheckout()))
	check(easykubeObj.Set(HELM_TEMPLATE, e.HelmTemplate()))
	check(easykubeObj.Set(PROCESS_SECRETS, e.ProcessExternalSecrets()))
	check(easykubeObj.Set(KEY_VALUE, e.KeyValue()))
	check(easykubeObj.Set(KEY_ENV, e.Env()))

	addon.ExportFunction("_ek", easykubeObj)

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
		ez.Kube.FmtRed("check addon %s, Call to %s expected %d arguments, %d is missing",
			e.AddonCtx.addon.Name,
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
