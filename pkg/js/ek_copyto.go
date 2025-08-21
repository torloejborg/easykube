package jsutils

import (
	"github.com/dop251/goja"
	"github.com/torloj/easykube/pkg/ek"
	"log"
	"path/filepath"
)

func (ctx *Easykube) CopyTo() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		ctx.checkArgs(call, COPY_TO)

		k8sutils := ek.NewK8SUtils(ctx.EKContext)
		deployment := call.Argument(0).String()
		namespace := call.Argument(1).String()
		containerLike := call.Argument(2).String()
		sourceFile := call.Argument(3).String()
		destinationFile := call.Argument(4).String()

		// the addon.ek.js file - we will resolve the manifest relative to that
		addon := ctx.AddonCtx.addon.File

		fullPath := filepath.Dir(addon.Name())

		podName, containerName, err := k8sutils.FindContainer(deployment, namespace, containerLike)
		if err != nil {
			log.Fatalf("LocatePod failed: %v", err)
		}

		err = k8sutils.CopyFileToPod(namespace, podName, containerName, filepath.Join(fullPath, sourceFile), destinationFile)
		if err != nil {
			log.Fatalf("%s failed: %v", COPY_TO, err)
		}

		return goja.Undefined()
	}

}
