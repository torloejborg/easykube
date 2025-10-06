package jsutils

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
)

type ExecResult struct {
	runtime *goja.Runtime
	self    goja.Value // the JS "this" object for chaining
	success bool
	output  string
}

func (er *ExecResult) OnSuccess(call goja.FunctionCall) goja.Value {
	if er.success && len(call.Arguments) == 1 {
		if fn, ok := goja.AssertFunction(call.Arguments[0]); ok {
			_, _ = fn(nil, er.runtime.ToValue(er.output))
		}
	}
	return er.self
}

func (er *ExecResult) OnFail(call goja.FunctionCall) goja.Value {
	if !er.success && len(call.Arguments) == 1 {
		if fn, ok := goja.AssertFunction(call.Arguments[0]); ok {
			_, _ = fn(nil, er.runtime.ToValue(er.output))
		}
	}
	return er.self
}

func (ctx *Easykube) Exec() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		er := &ExecResult{runtime: ctx.AddonCtx.vm}
		obj := ctx.AddonCtx.NewObject()
		er.self = obj

		// bind methods
		_ = obj.Set("onSuccess", er.OnSuccess)
		_ = obj.Set("onFail", er.OnFail)

		osCommand := call.Argument(0).String()
		args, _ := exportStringArray(call.Argument(1).Export())

		if ez.Kube.IsDryRun() {
			ez.Kube.FmtDryRun("%s %s", osCommand, strings.Join(args, " "))
			return obj
		} else {

			_, notfoundErr := exec.LookPath(osCommand)

			if notfoundErr != nil {
				er.output = notfoundErr.Error()
				er.success = false
				return obj

			} else {

				var outBuf, errBuf bytes.Buffer
				cmd := exec.Command(osCommand, args...)
				cmd.Stdout = &outBuf
				cmd.Stderr = &errBuf

				_ = cmd.Run()

				er.output = outBuf.String() + errBuf.String()
				er.success = cmd.ProcessState.Success()

				return obj
			}
		}
	}
}
