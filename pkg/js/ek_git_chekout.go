package jsutils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dop251/goja"
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
		addonDir := filepath.Dir(e.AddonCtx.addon.GetAddonFile())
		destination := filepath.Join(addonDir, call.Argument(2).String())

		if e.ek.Utils.FileOrDirExists(destination) {
			e.ek.Printer.FmtYellow("%s already exists, skipping checkout", destination)
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

			git := "git"
			cmdStr := fmt.Sprintf("%s %s", git, strings.Join(args, " "))

			if e.ek.CommandContext.IsVerbose() {
				e.ek.Printer.FmtVerbose(cmdStr)
			}
			if e.ek.CommandContext.IsDryRun() {
				e.ek.Printer.FmtDryRun(cmdStr)
			} else {

				_, stderr, err := e.ek.ExternalTools.RunCommand("git", args...)

				if err != nil {
					e.ek.Printer.FmtRed(stderr, err.Error())
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
