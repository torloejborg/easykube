package jsutils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
)

func (e *Easykube) GitSparseCheckout() func(goja.FunctionCall) goja.Value {

	return func(call goja.FunctionCall) goja.Value {
		e.checkArgs(call, SPARSE_CHECKOUT)
		ezk := ez.Kube

		currentDir, _ := os.Getwd()
		defer func() {
			if !ezk.IsDryRun() {
				os.Chdir(currentDir)
			}
			if ezk.IsVerbose() {
				ezk.FmtVerbose("cd %s", currentDir)
			}
		}()

		repo := call.Argument(0).String()
		branch := call.Argument(1).String()
		source := call.Argument(2)

		gitSparseDirectoryList := e.extractStringSliceFromArgument(source)
		addonDir := filepath.Dir(e.AddonCtx.addon.GetAddonFile())
		destination := filepath.Join(addonDir, call.Argument(3).String())

		if ez.FileOrDirExists(destination) {
			ezk.FmtYellow("%s already exists, skipping sparseCheckout", destination)
			return call.This
		}

		if !ezk.IsDryRun() {
			err := ezk.MkdirAll(destination, 0777)
			if err != nil {
				panic(err)
			}
			os.Chdir(destination)
		} else {
			ezk.FmtDryRun("mkdir -p %s", destination)
			ezk.FmtDryRun("cd %s", destination)
		}

		if ezk.IsVerbose() {
			ezk.FmtVerbose("cd %s", destination)
		}

		gitCmd := func(args []string) {

			git := "git"
			cmdStr := fmt.Sprintf("%s %s", git, strings.Join(args, " "))

			if ezk.IsVerbose() {
				ezk.FmtVerbose(cmdStr)
			}
			if ezk.IsDryRun() {
				ezk.FmtDryRun(cmdStr)
			} else {

				_, stderr, err := ezk.RunCommand("git", args...)

				if err != nil {
					ezk.FmtRed(stderr, err.Error())
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
