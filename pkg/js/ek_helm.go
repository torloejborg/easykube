package jsutils

import (
	"fmt"
	"os"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
)

func (ctx *Easykube) HelmTemplate() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		ezk := ez.Kube
		chart := call.Argument(0).String()
		values := call.Argument(1).String()
		destination := call.Argument(2).String()
		namespace := call.Argument(3).String()

		if !ez.FileOrDirExists(chart) {
			ez.Kube.FmtRed("specified chart %s does not exist", chart)
			os.Exit(-1)
		}

		if !ez.FileOrDirExists(values) {
			ezk.FmtRed("the value file %s does not exist", values)
			os.Exit(-1)
		}

		if namespace == "" {
			namespace = "default"
		}

		cmd := "helm"
		args := []string{"template", chart,
			"--values", values,
			"--namespace", namespace}

		cmdStr := fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
		if ezk.IsVerbose() {
			ezk.FmtVerbose(cmdStr)
		}

		stdout, stderr, err := ezk.RunCommand(cmd, args...)

		if err != nil {
			ezk.FmtRed("helm failed %s", stderr)
			os.Exit(-1)
		}

		ez.SaveFile(stdout, destination)

		return call.This
	}
}
