package jsutils

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"sync"

	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"
	"k8s.io/utils/ptr"

	"github.com/dop251/goja"
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

		var i = 0
		var wg sync.WaitGroup

		if mustPull {
			out.FmtYellow("🖼 will pull fresh images")
		}

		for source, dest := range result {
			i++
			wg.Add(1)
			go func() {
				if !ez.Kube.HasImageInKindRegistry(dest) || mustPull {

					if strings.Contains(source, "ccta.dk") {
						s, err := ez.CreateK8sUtilsImpl().GetSecret("easykube-secrets", "default")
						if err != nil {
							panic(err)
						}

						jsonBytes, _ := json.Marshal(map[string]string{
							"username": string(s["artifactoryUsername"]),
							"password": string(s["artifactoryPassword"]),
						})

						out.FmtGreen("🖼  pull from private registry %s", source)
						ez.Kube.PullImage(source, ptr.To(base64.StdEncoding.EncodeToString(jsonBytes)))

					} else {
						out.FmtGreen("🖼  pull %s", source)
						ez.Kube.PullImage(source, nil)
					}

					out.FmtGreen("🖼  tag %s to %s", source, dest)
					ez.Kube.TagImage(source, dest)

					ez.Kube.PushImage(dest)
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
