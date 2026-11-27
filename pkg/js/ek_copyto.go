package jsutils

import (
	"path/filepath"

	"github.com/dop251/goja"
)

var CopyToDoc = []DocumentedFunction{{
	Name: CopyTo, Description: "Copies a local file to a container in a pod. This is done by best effort. " +
		"A pod can have many containers. If two (or more) containers exist which matches the partial container name, " +
		"the first becomes the target. " +
		"Limitation; The destination MUST be aimed at a deployment, you cannot target a bare container. " +
		"This is why you must specify a deployment name. Also tar must be available in the container.", Returns: "this",
	Params: []DocumentedParam{
		{Name: "deployment", Description: "Name of deployment", Type: "string"},
		{Name: "namespace", Description: "Where the target deployment lives", Type: "string"},
		{Name: "container", Description: "A partial name of the target container", Type: "string"},
		{Name: "source", Description: "Source file, relative to the *.ek.js module definition", Type: "string"},
		{Name: "destination", Description: "Destination file in container, destination must be writeable", Type: "string"},
	},
}}

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
