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
	// skopeoExec runs concurrently with the spinner manager's render loop.
	// Anything written directly to stdout while ysmrr is rendering breaks its
	// cursor bookkeeping (save/restore cursor and cursor-up escape codes),
	// which makes the cursor jump to stale positions and output get painted
	// over existing console content. Therefore:
	//
	//   - error details are reported through the spinner itself, and
	//   - verbose command logging is collected and printed only after the
	//     spinner manager has stopped.
	skopeoExec := func(src, dest string,
		wg *sync.WaitGroup,
		sem chan struct{},
		sm ysmrr.SpinnerManager,
		verboseCmds *[]string,
		verboseMu *sync.Mutex) {

		defer wg.Done()
		var args = []string{"copy", "docker://" + src, "docker://" + dest}

		cmdStr := fmt.Sprintf("%s %s", constants.SkopeoBinary, strings.Join(args, " "))

		if ctx.ek.CommandContext.IsDryRun() {
			ctx.ek.Printer.FmtDryRun(cmdStr)
		} else {
			spinner := sm.AddSpinner(spinnerMessage("%s", dest))

			_, stderr, err := ctx.ek.ExternalTools.RunCommand(constants.SkopeoBinary, args...)
			if err != nil {
				spinner.ErrorWithMessage(spinnerMessage("skopeo copy failed: %s", stderr))
				panic(err)
			}

			if ctx.ek.CommandContext.IsVerbose() {
				verboseMu.Lock()
				*verboseCmds = append(*verboseCmds, cmdStr)
				verboseMu.Unlock()
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

		verboseCmds := make([]string, 0)
		verboseMu := sync.Mutex{}

		sm := ysmrr.NewSpinnerManager()
		// in dry-run mode every worker prints to stdout, so the spinner
		// manager must not render at all.
		if !ctx.ek.CommandContext.IsDryRun() {
			sm.Start()
		}

		for src, dest := range input {

			exists, err := ctx.ek.ContainerRuntime.HasImageInKindRegistry(dest)

			if err != nil {
				panic(err)
			}

			if !exists || mustPull {
				wg.Add(1)
				sem <- struct{}{}
				wg.Go(func() { skopeoExec(src, dest, &wg, sem, sm, &verboseCmds, &verboseMu) })
			}
		}

		wg.Wait()
		sm.Stop()
		close(sem)

		// deferred verbose logging: safe to write to stdout again now that
		// the spinner manager is no longer rendering.
		for _, cmd := range verboseCmds {
			ctx.ek.Printer.FmtVerbose(cmd)
		}

		return call.This
	}
}
