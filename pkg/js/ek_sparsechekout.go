package jsutils

import (
	"fmt"
	"os"
	"strings"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
)

func (e *Easykube) GitSparseCheckout() func(goja.FunctionCall) goja.Value {

	return func(call goja.FunctionCall) goja.Value {
		e.checkArgs(call, SPARSE_CHECKOUT)
		ezk := ez.Kube

		currentDir, _ := os.Getwd()
		defer os.Chdir(currentDir)

		repo := call.Argument(0).String()
		branch := call.Argument(1).String()
		source := call.Argument(2)

		gitSparseDirectoryList := e.extractStringSliceFromArgument(source)

		destination := call.Argument(3).String()
		if ez.FileOrDirExists(destination) {
			ezk.FmtYellow("Repository %s already checked out at %s", repo, destination)
			return call.This
		}

		err := ezk.MkdirAll(destination, 0777)
		if err != nil {
			panic(err)
		}

		os.Chdir(destination)

		gitCmd := func(args []string) {

			git := "git"
			cmdStr := fmt.Sprintf("%s %s", git, strings.Join(args, " "))

			if ezk.IsVerbose() {
				ezk.FmtVerbose(cmdStr)
			}

			_, stderr, err := ezk.RunCommand("git", args...)

			if err != nil {
				ezk.FmtRed(stderr, err.Error())
				os.Exit(1)
			}
		}

		gitCmd([]string{"init"})
		gitCmd([]string{"config", "core.sparsecheckout", "true"})
		gitCmd([]string{"remote", "add", "-f", "origin", repo})
		gitCmd([]string{"pull", "origin", branch})

		gitArgs := []string{"sparse-checkout", "set"}
		allArgs := append(gitArgs, gitSparseDirectoryList...)

		gitCmd(allArgs)

		return call.This
	}

}
