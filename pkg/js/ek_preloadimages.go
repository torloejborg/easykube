package jsutils

import (
	"strings"
	"sync"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/ez"
)

func (ctx *Easykube) PreloadImages() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {

		ezk := ez.Kube
		if ezk.IsDryRun() {
			ezk.FmtDryRun("skipping preload")
			return call.This
		}

		mustPull := ctx.CobraCommandHelder.GetBoolFlag(constants.FLAG_PULL)
		ctx.checkArgs(call, PRELOAD)
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

				registryCredentials := getPrivateRegistryCredentials(source, config.PrivateRegistries)
				if hasImage, err := ezk.HasImageInKindRegistry(dest); err != nil {
					panic(err)
				} else if !hasImage || mustPull {
					if registryCredentials != nil {

						ezk.FmtGreen("🖼  pull from private registry %s using credentials (%s,%s)", source,
							registryCredentials.Username,
							"[redacted]")

						if err := ezk.PullImage(source, registryCredentials); err != nil {
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

			if i > 0 {
				wg.Wait()
			}
		}
		return goja.Undefined()
	}
}

func getPrivateRegistryCredentials(registry string, config []ez.PrivateRegistry) *ez.PrivateRegistryCredentials {

	secret, err := ez.Kube.GetSecret("easykube-secrets", "default")

	if err != nil {
		return nil
	}

	for i := range config {

		if strings.Contains(registry, config[i].RepositoryMatch) {

			if secret[config[i].UserKey] == nil || secret[config[i].PasswordKey] == nil {
				ez.Kube.FmtYellow("Did not find credential keys for registry-partial %s", config[i].RepositoryMatch)
				return nil
			}
			return &ez.PrivateRegistryCredentials{
				Username: string(secret[config[i].UserKey]),
				Password: string(secret[config[i].PasswordKey]),
			}
		}
	}

	return nil
}
