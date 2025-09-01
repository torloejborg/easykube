package jsutils

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"sync"

	"github.com/torloj/easykube/pkg/constants"
	"k8s.io/utils/ptr"

	"github.com/dop251/goja"
	"github.com/torloj/easykube/pkg/ek"
)

func (ctx *Easykube) PreloadImages() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		out := ctx.EKContext.Printer
		mustPull := ctx.EKContext.GetBoolFlag(constants.FLAG_PULL)
		ctx.checkArgs(call, PRELOAD)

		var arg = call.Argument(0)
		result := make(map[string]string)

		err := ctx.AddonCtx.vm.ExportTo(arg, &result)
		if err != nil {
			panic(err)
		}
		cru := ek.NewContainerRuntime(ctx.EKContext)

		var i = 0
		var wg sync.WaitGroup

		if mustPull {
			out.FmtYellow("🖼 will pull fresh images")
		}

		for source, dest := range result {
			i++
			wg.Add(1)
			go func() {
				if !cru.HasImageInKindRegistry(dest) || mustPull {

					if strings.Contains(source, "ccta.dk") {
						s, err := ek.NewK8SUtils(ctx.EKContext).GetSecret("easykube-secrets", "default")
						if err != nil {
							panic(err)
						}

						jsonBytes, _ := json.Marshal(map[string]string{
							"username": string(s["artifactoryUsername"]),
							"password": string(s["artifactoryPassword"]),
						})

						out.FmtGreen("🖼  pull from private registry %s", source)
						cru.Pull(source, ptr.To(base64.StdEncoding.EncodeToString(jsonBytes)))

					} else {
						out.FmtGreen("🖼  pull %s", source)
						cru.Pull(source, nil)
					}

					out.FmtGreen("🖼  tag %s to %s", source, dest)
					cru.Tag(source, dest)

					cru.Push(dest)
					out.FmtGreen("🖼  push %s", dest)
				}
				defer wg.Done()
			}()
		}

		if i > 0 {
			wg.Wait()
		}

		return goja.Undefined()
	}
}
