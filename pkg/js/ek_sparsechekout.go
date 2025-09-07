package jsutils

import (
	"os"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ek"
)

func (e *Easykube) GitSparseCheckout() func(goja.FunctionCall) goja.Value {

	return func(call goja.FunctionCall) goja.Value {
		e.checkArgs(call, SPARSE_CHECKOUT)

		currentDir, _ := os.Getwd()
		defer os.Chdir(currentDir)

		out := e.EKContext.Printer
		repo := call.Argument(0).String()
		branch := call.Argument(1).String()
		source := call.Argument(2)

		gitSparseDirectoryList := e.extractStringSliceFromArgument(source)

		destination := call.Argument(3).String()
		u := ek.Utils{Fs: e.EKContext.Fs}
		if u.FileOrDirExists(destination) {
			out.FmtYellow("Repository %s already checked out at %s", repo, destination)
			return call.This
		}

		err := e.EKContext.Fs.MkdirAll(destination, 0777)
		if err != nil {
			panic(err)
		}

		os.Chdir(destination)

		handleCmd := func(stdout, stderr string, err error) {
			if err != nil {
				panic(err)
			}

			if stderr != "" {
				out.FmtGreen(stderr)
			}

			if stdout != "" {
				out.FmtGreen(stdout)
			}
		}

		tool := ek.NewExternalTools(e.EKContext)

		handleCmd(tool.RunCommand("git", "init"))
		handleCmd(tool.RunCommand("git", "config", "core.sparsecheckout", "true"))
		handleCmd(tool.RunCommand("git", "remote", "add", "-f", "origin", repo))
		handleCmd(tool.RunCommand("git", "pull", "origin", branch))

		gitArgs := []string{"sparse-checkout", "set"}
		allArgs := append(gitArgs, gitSparseDirectoryList...)

		handleCmd(tool.RunCommand("git", allArgs...))

		return call.This
	}

}
