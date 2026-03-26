package jsutils

import (
	"strings"
	"sync"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"
)

func (ctx *Easykube) PreloadImages(noop bool) func(goja.FunctionCall) goja.Value {
	if noop {
		return NoopFunc()
	}

	return ctx.preloadImages()
}

func (ctx *Easykube) preloadImages() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {

		ezk := ez.Kube

		mustPull := ctx.CobraCommandHelder.GetBoolFlag(constants.FlagPull)
		ctx.checkArgs(call, Preload)
		config, _ := ez.Kube.LoadConfig()

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

				registryCredentials := getPrivateRegistryCredentials(source, config.MirrorRegistries)
				if hasImage, err := ezk.HasImageInKindRegistry(dest); err != nil {
					panic(err)
				} else if !hasImage || mustPull {
					if registryCredentials != nil {
						ezk.FmtGreen("🖼 pull from private registry %s using secret keys (%s,%s)", source,
							registryCredentials.Username,
							"[redacted]")
						if err := ezk.PullImage(source, registryCredentials); err != nil {
							panic(err)
						}

					} else {
						ezk.FmtGreen("🖼 pull %s", source)
						if err := ezk.PullImage(source, nil); err != nil {
							panic(err)
						}
					}

					ezk.FmtGreen("🖼 tag %s to %s", source, dest)
					if err := ezk.TagImage(source, dest); err != nil {
						panic(err)
					}

					if err := ezk.PushImage(source, dest); err != nil {
						panic(err)
					}
					ezk.FmtGreen("🖼 pushed %s", dest)
				}
				defer wg.Done()
			}()

			if i > 0 {
				wg.Wait()
			}
		}
		return goja.Undefined()
	}
}

func getPrivateRegistryCredentials(registry string, config []ez.MirrorRegistry) *ez.PrivateRegistryCredentials {

	for i := range config {

		x := strings.ReplaceAll(config[i].RegistryUrl, "https://", "")
		x = strings.ReplaceAll(x, "http://", "")

		if strings.Contains(registry, x) {
			return &ez.PrivateRegistryCredentials{
				Username: config[i].UserKey,
				Password: config[i].PasswordKey,
			}
		}
	}

	return nil
}
