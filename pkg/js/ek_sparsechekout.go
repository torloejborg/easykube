package jsutils

import (
	"os"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
)

func (e *Easykube) GitSparseCheckout() func(goja.FunctionCall) goja.Value {

	return func(call goja.FunctionCall) goja.Value {
		e.checkArgs(call, SPARSE_CHECKOUT)

		currentDir, _ := os.Getwd()
		defer os.Chdir(currentDir)

		repo := call.Argument(0).String()
		branch := call.Argument(1).String()
		source := call.Argument(2)

		gitSparseDirectoryList := e.extractStringSliceFromArgument(source)

		destination := call.Argument(3).String()
		if ez.FileOrDirExists(destination) {
			ez.Kube.FmtYellow("Repository %s already checked out at %s", repo, destination)
			return call.This
		}

		err := ez.Kube.MkdirAll(destination, 0777)
		if err != nil {
			panic(err)
		}

		os.Chdir(destination)

		handleCmd := func(stdout, stderr string, err error) {
			if err != nil {
				panic(err)
			}

			if stderr != "" {
				ez.Kube.FmtGreen(stderr)
			}

			if stdout != "" {
				ez.Kube.FmtGreen(stdout)
			}
		}

		handleCmd(ez.Kube.RunCommand("git", "init"))
		handleCmd(ez.Kube.RunCommand("git", "config", "core.sparsecheckout", "true"))
		handleCmd(ez.Kube.RunCommand("git", "remote", "add", "-f", "origin", repo))
		handleCmd(ez.Kube.RunCommand("git", "pull", "origin", branch))

		gitArgs := []string{"sparse-checkout", "set"}
		allArgs := append(gitArgs, gitSparseDirectoryList...)

		handleCmd(ez.Kube.RunCommand("git", allArgs...))

		return call.This
	}

}
