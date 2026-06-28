package jsutils

import (
	"fmt"
	"strings"
	"sync"

	"github.com/chelnak/ysmrr"
	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/constants"
)

func (ctx *Easykube) SkopeoPreload(noop bool) func(goja.FunctionCall) goja.Value {
	if noop {
		return NoopFunc()
	}
	return ctx.skopeoPreload()
}

func (ctx *Easykube) skopeoPreload() func(goja.FunctionCall) goja.Value {
	skopeoExec := func(src, dest string, wg *sync.WaitGroup, sem chan struct{}, sm ysmrr.SpinnerManager) {
		defer wg.Done()
		var args = []string{"copy", "docker://" + src, "docker://" + dest}

		cmdStr := fmt.Sprintf("%s %s", constants.SkopeoBinary, strings.Join(args, " "))
		if ctx.ek.CommandContext.IsVerbose() {
			ctx.ek.Printer.FmtVerbose(cmdStr)
		}

		if ctx.ek.CommandContext.IsDryRun() {
			ctx.ek.Printer.FmtDryRun(cmdStr)
		} else {

			spinner := sm.AddSpinner(dest)

			_, stderr, err := ctx.ek.ExternalTools.RunCommand(constants.SkopeoBinary, args...)
			if err != nil {
				ctx.ek.Printer.FmtRed("%s failed %s", constants.SkopeoBinary, stderr)
				spinner.Error()
				panic(err)
			}

			spinner.Complete()

		}

		<-sem
	}

	return func(call goja.FunctionCall) goja.Value {

		mustPull := ctx.ek.CommandContext.GetBoolFlag(constants.FlagPull)
		input := make(map[string]string)
		arg := call.Argument(0)
		err := ctx.AddonCtx.vm.ExportTo(arg, &input)

		if err != nil {
			panic(err)
		}
		wg := sync.WaitGroup{}
		sem := make(chan struct{}, 3)
		sm := ysmrr.NewSpinnerManager()
		sm.Start()
		defer sm.Stop()

		for src, dest := range input {

			exists, err := ctx.ek.ContainerRuntime.HasImageInKindRegistry(dest)

			if err != nil {
				panic(err)
			}

			if !exists || mustPull {
				wg.Add(1)
				sem <- struct{}{}
				wg.Go(func() { skopeoExec(src, dest, &wg, sem, sm) })
			}
		}

		wg.Wait()

		close(sem)
		return call.This
	}
}
