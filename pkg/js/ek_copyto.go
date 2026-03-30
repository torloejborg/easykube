package jsutils

import (
	"path/filepath"

	"github.com/dop251/goja"
)

func (ctx *Easykube) CopyTo(noop bool) func(goja.FunctionCall) goja.Value {
	if noop {
		return NoopFunc()
	}
	return ctx.andThenApply()
}

func (ctx *Easykube) copyTo() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		ctx.checkArgs(call, CopyTo)

		if ctx.ek.CommandContext.IsDryRun() {
			ctx.ek.Printer.FmtDryRun("skipping copyTo")
			return call.This
		}

		deployment := call.Argument(0).String()
		namespace := call.Argument(1).String()
		containerLike := call.Argument(2).String()
		sourceFile := call.Argument(3).String()
		destinationFile := call.Argument(4).String()

		// the addon.ek.js file - we will resolve the manifest relative to that
		addon := ctx.AddonCtx.addon.GetAddonFile()

		fullPath := filepath.Dir(addon)

		podName, containerName, err := ctx.ek.Kubernetes.FindContainerInPod(deployment, namespace, containerLike)
		if err != nil {
			ctx.ek.Printer.FmtRed("LocatePod failed: %v", err)
		}

		err = ctx.ek.Kubernetes.CopyFileToPod(namespace, podName, containerName, filepath.Join(fullPath, sourceFile), destinationFile)
		if err != nil {
			ctx.ek.Printer.FmtRed("%s failed: %v", CopyTo, err)
		}

		return call.This
	}

}
