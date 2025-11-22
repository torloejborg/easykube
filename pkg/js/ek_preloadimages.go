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
			ezk.FmtYellow("🖼 will pull fresh images")
		}

		for source, dest := range result {
			i++
			wg.Add(1)
			go func() {
				if !ezk.HasImageInKindRegistry(dest) || mustPull {

					registryCredentials := getPrivateRegistryCredentials(source, config.PrivateRegistries)

					if registryCredentials.Base64EncodedDockerSecret != "" {

						ezk.FmtGreen("🖼  pull from private registry %s using secret keys (%s,%s)", source,
							registryCredentials.UserKey,
							registryCredentials.PasswordKey)

						ezk.PullImage(source, ptr.To(registryCredentials.Base64EncodedDockerSecret))

					} else {
						ezk.FmtGreen("🖼  pull %s", source)
						ezk.PullImage(source, nil)
					}

					ezk.FmtGreen("🖼  tag %s to %s", source, dest)
					ezk.TagImage(source, dest)

					ezk.PushImage(dest)
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

type privateRegistryCredentials struct {
	UserKey                   string
	PasswordKey               string
	Base64EncodedDockerSecret string
}

func getPrivateRegistryCredentials(registry string, config []ez.PrivateRegistry) privateRegistryCredentials {

	secret, err := ez.Kube.GetSecret("easykube-secrets", "default")

	if err != nil {
		panic(err)
	}

	if strings.Contains(registry, "ccta.dk") {
		jsonBytes, _ := json.Marshal(map[string]string{
			"username": string(secret["artifactoryUsername"]),
			"password": string(secret["artifactoryPassword"]),
		})

		return privateRegistryCredentials{
			UserKey:                   "artifactoryUsername",
			PasswordKey:               "artifactoryPassword",
			Base64EncodedDockerSecret: base64.StdEncoding.EncodeToString(jsonBytes),
		}
	}

	for i := range config {

		if strings.Contains(registry, config[i].RepositoryMatch) {

			if secret[config[i].UserKey] == nil || secret[config[i].PasswordKey] == nil {
				ez.Kube.FmtYellow("Did not find credential keys for registry-partial %s", config[i].RepositoryMatch)
				return privateRegistryCredentials{"", "", ""}
			}

			jsonBytes, _ := json.Marshal(map[string]string{
				"username": string(secret[config[i].UserKey]),
				"password": string(secret[config[i].PasswordKey]),
			})

			return privateRegistryCredentials{
				UserKey:                   config[i].UserKey,
				PasswordKey:               config[i].PasswordKey,
				Base64EncodedDockerSecret: base64.StdEncoding.EncodeToString(jsonBytes)}
		}
	}

	return privateRegistryCredentials{"", "", ""}
}
