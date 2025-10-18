package jsutils

import (
	"path/filepath"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
)

func (ctx *Easykube) CopyTo() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		ctx.checkArgs(call, COPY_TO)
		ezk := ez.Kube

		if ezk.IsDryRun() {
			ezk.FmtDryRun("skipping copyTo")
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

		podName, containerName, err := ezk.FindContainerInPod(deployment, namespace, containerLike)
		if err != nil {
			ezk.FmtRed("LocatePod failed: %v", err)
		}

		err = ezk.CopyFileToPod(namespace, podName, containerName, filepath.Join(fullPath, sourceFile), destinationFile)
		if err != nil {
			ezk.FmtRed("%s failed: %v", COPY_TO, err)
		}

		return call.This
	}

}
