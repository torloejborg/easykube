package jsutils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/constants"
)

func (e *Easykube) GitSparseCheckout(noop bool) func(goja.FunctionCall) goja.Value {
	if noop {
		return NoopFunc()
	}
	return e.gitCheckout()
}

func (e *Easykube) gitSparseCheckout() func(goja.FunctionCall) goja.Value {

	return func(call goja.FunctionCall) goja.Value {
		e.checkArgs(call, GitSparseCheckout)

		currentDir, _ := os.Getwd()
		defer func() {
			if !e.ek.CommandContext.IsDryRun() {
				err := os.Chdir(currentDir)
				if err != nil {
					panic(err)
				}
			}
			if e.ek.CommandContext.IsVerbose() {
				e.ek.Printer.FmtVerbose("cd %s", currentDir)
			}
		}()

		repo := call.Argument(0).String()
		branch := call.Argument(1).String()
		source := call.Argument(2)

		gitSparseDirectoryList := e.extractStringSliceFromArgument(source)
		addonDir := filepath.Dir(e.AddonCtx.addon.GetAddonFile())
		destination := filepath.Join(addonDir, call.Argument(3).String())

		if e.ek.Utils.FileOrDirExists(destination) {
			e.ek.Printer.FmtYellow("%s already exists, skipping sparseCheckout", destination)
			return call.This
		}

		if !e.ek.CommandContext.IsDryRun() {
			err := e.ek.Fs.MkdirAll(destination, 0777)
			if err != nil {
				panic(err)
			}
			err = os.Chdir(destination)
			if err != nil {
				panic(err)
			}
		} else {
			e.ek.Printer.FmtDryRun("mkdir -p %s", destination)
			e.ek.Printer.FmtDryRun("cd %s", destination)
		}

		if e.ek.CommandContext.IsVerbose() {
			e.ek.Printer.FmtVerbose("cd %s", destination)
		}

		gitCmd := func(args []string) {

			cmdStr := fmt.Sprintf("%s %s", constants.GitBinary, strings.Join(args, " "))

			if e.ek.CommandContext.IsVerbose() {
				e.ek.Printer.FmtVerbose(cmdStr)
			}
			if e.ek.CommandContext.IsDryRun() {
				e.ek.Printer.FmtDryRun(cmdStr)
			} else {

				_, stderr, err := e.ek.ExternalTools.RunCommand(constants.GitBinary, args...)

				if err != nil {
					e.ek.Printer.FmtRed(stderr, err.Error())
					os.Exit(1)
				}
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
