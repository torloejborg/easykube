package jsutils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
)

func (e *Easykube) GitCheckout(noop bool) func(goja.FunctionCall) goja.Value {
	if noop {
		return NoopFunc()
	}
	return e.gitCheckout()
}

func (e *Easykube) gitCheckout() func(goja.FunctionCall) goja.Value {

	return func(call goja.FunctionCall) goja.Value {
		e.checkArgs(call, GitCheckout)
		ezk := ez.Kube

		currentDir, _ := os.Getwd()
		defer func() {
			if !ezk.IsDryRun() {
				err := os.Chdir(currentDir)
				if err != nil {
					panic(err)
				}
			}
			if ezk.IsVerbose() {
				ezk.FmtVerbose("cd %s", currentDir)
			}
		}()

		repo := call.Argument(0).String()
		branch := call.Argument(1).String()
		addonDir := filepath.Dir(e.AddonCtx.addon.GetAddonFile())
		destination := filepath.Join(addonDir, call.Argument(2).String())

		if ez.FileOrDirExists(destination) {
			ezk.FmtYellow("%s already exists, skipping checkout", destination)
			return call.This
		}

		if !ezk.IsDryRun() {
			err := ezk.MkdirAll(destination, 0777)
			if err != nil {
				panic(err)
			}
			err = os.Chdir(destination)
			if err != nil {
				panic(err)
			}
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

		gitCmd([]string{"clone", repo, destination})
		gitCmd([]string{"checkout", branch})
		gitCmd([]string{"pull"})

		return call.This
	}

}
