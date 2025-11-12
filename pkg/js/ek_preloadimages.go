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

		ezk := ez.Kube
		if ezk.IsDryRun() {
			ezk.FmtDryRun("skipping preload")
			return goja.Undefined()
		}

		mustPull := ctx.CobraCommandHelder.GetBoolFlag(constants.FLAG_PULL)
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
			ezk.FmtGreen("🖼 will pull fresh images")
		}

		for source, dest := range result {
			i++
			wg.Add(1)
			go func() {

				if hasImage, err := ezk.HasImageInKindRegistry(dest); err != nil {
					panic(err)
				} else if !hasImage || mustPull {
					if strings.Contains(source, "ccta.dk") {
						s, err := ezk.GetSecret("easykube-secrets", "default")
						if err != nil {
							panic(err)
						}

						jsonBytes, _ := json.Marshal(map[string]string{
							"username": string(s["artifactoryUsername"]),
							"password": string(s["artifactoryPassword"]),
						})

						ezk.FmtGreen("🖼  pull from private registry %s", source)
						if err := ezk.PullImage(source, ptr.To(base64.StdEncoding.EncodeToString(jsonBytes))); err != nil {
							panic(err)
						}

					} else {
						ezk.FmtGreen("🖼  pull %s", source)
						if err := ezk.PullImage(source, nil); err != nil {
							panic(err)
						}
					}

					ezk.FmtGreen("🖼  tag %s to %s", source, dest)
					if err := ezk.TagImage(source, dest); err != nil {
						panic(err)
					}

					if err := ezk.PushImage(source, dest); err != nil {
						panic(err)
					}
					ezk.FmtGreen("🖼  pushed %s", dest)
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
