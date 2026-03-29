package jsutils

import (
	"strings"
	"sync"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/core"
)

func (ctx *Easykube) PreloadImages(noop bool) func(goja.FunctionCall) goja.Value {
	if noop {
		return NoopFunc()
	}

	return ctx.preloadImages()
}

func (ctx *Easykube) preloadImages() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {

		mustPull := ctx.ek.CommandContext.GetBoolFlag(constants.FlagPull)
		ctx.checkArgs(call, Preload)
		config, _ := ctx.ek.Config.LoadConfig()

		var arg = call.Argument(0)
		result := make(map[string]string)

		err := ctx.AddonCtx.vm.ExportTo(arg, &result)
		if err != nil {
			panic(err)
		}

		var i = 0
		var wg sync.WaitGroup

		if mustPull {
			ctx.ek.Printer.FmtGreen("🖼 will pull fresh images")
		}

		for source, dest := range result {
			i++
			wg.Add(1)
			go func() {

				registryCredentials := getPrivateRegistryCredentials(source, config.MirrorRegistries)
				if hasImage, err := ctx.ek.ContainerRuntime.HasImageInKindRegistry(dest); err != nil {
					panic(err)
				} else if !hasImage || mustPull {
					if registryCredentials != nil {
						ctx.ek.Printer.FmtGreen("🖼 pull from private registry %s using secret keys (%s,%s)", source,
							registryCredentials.Username,
							"[redacted]")
						if err := ctx.ek.ContainerRuntime.PullImage(source, registryCredentials); err != nil {
							panic(err)
						}

					} else {
						ctx.ek.Printer.FmtGreen("🖼 pull %s", source)
						if err := ctx.ek.ContainerRuntime.PullImage(source, nil); err != nil {
							panic(err)
						}
					}

					ctx.ek.Printer.FmtGreen("🖼 tag %s to %s", source, dest)
					if err := ctx.ek.ContainerRuntime.TagImage(source, dest); err != nil {
						panic(err)
					}

					if err := ctx.ek.ContainerRuntime.PushImage(source, dest); err != nil {
						panic(err)
					}
					ctx.ek.Printer.FmtGreen("🖼 pushed %s", dest)
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

func getPrivateRegistryCredentials(registry string, config []core.MirrorRegistry) *core.PrivateRegistryCredentials {

	for i := range config {

		x := strings.ReplaceAll(config[i].RegistryUrl, "https://", "")
		x = strings.ReplaceAll(x, "http://", "")

		if strings.Contains(registry, x) {
			return &core.PrivateRegistryCredentials{
				Username: config[i].UserKey,
				Password: config[i].PasswordKey,
			}
		}
	}

	return nil
}
