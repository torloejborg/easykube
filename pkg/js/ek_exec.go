package jsutils

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
)

func (ctx *Easykube) Exec() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {

		ezk := ez.Kube
		if ezk.IsDryRun() {
			ezk.FmtDryRun("skipping exec")
			return goja.Undefined()
		}

		osCommand := call.Argument(0).String()
		args, _ := exportStringArray(call.Argument(1).Export())

		if ezk.IsVerbose() {
			ezk.FmtVerbose(fmt.Sprintf("%s %s", osCommand, strings.Join(args, " ")))
		}

		_, err := exec.LookPath(osCommand)
		if err != nil {
			ezk.FmtRed("⚠ %s could not execute (not found)", osCommand)
			return goja.Undefined()
		}

		stdout, stderr, errx := ezk.RunCommand(osCommand, args...)

		if errx != nil {
			panic(errx)
		}

		fmt.Println(stdout)
		fmt.Println(stderr)

		obj := ctx.AddonCtx.vm.NewObject()
		obj.Set("stdout", stdout)
		obj.Set("stderr", stderr)
		return obj
	}
}
