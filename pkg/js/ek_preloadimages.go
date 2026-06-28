package jsutils

import (
	"strings"
	"sync"

	"github.com/chelnak/ysmrr"
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

	pullAndTag := func(source, destination string,
		config *core.EasykubeConfigData,
		group *sync.WaitGroup,
		sm ysmrr.SpinnerManager,
		sem chan struct{}) error {

		defer group.Done()
		registryCredentials := getPrivateRegistryCredentials(source, config.MirrorRegistries)
		mustPull := ctx.ek.CommandContext.GetBoolFlag(constants.FlagPull)

		if hasImage, err := ctx.ek.ContainerRuntime.HasImageInKindRegistry(destination); err != nil {
			panic(err)
		} else if !hasImage || mustPull {
			spinner := sm.AddSpinner("")
			if registryCredentials != nil {
				spinner.UpdateMessagef("pull %s from private registry using credentials (%s,%s)", source,
					registryCredentials.Username,
					"[redacted]")

				if err := ctx.ek.ContainerRuntime.PullImage(source, registryCredentials); err != nil {
					spinner.ErrorWithMessagef("pull %s from private registry failed: %s", source, err)
				}

			} else {
				spinner.UpdateMessagef("pull %s", source)
				if err := ctx.ek.ContainerRuntime.PullImage(source, nil); err != nil {
					spinner.ErrorWithMessagef("pull %s failed: %s", source, err)
					<-sem
					return err
				}
			}

			if err := ctx.ek.ContainerRuntime.TagImage(source, destination); err != nil {
				spinner.ErrorWithMessagef("tag %s to %s failed: %s", source, destination, err)
				<-sem
				return err
			}

			if err := ctx.ek.ContainerRuntime.PushImage(source, destination); err != nil {
				spinner.ErrorWithMessagef("push %s to %s failed: %s", source, destination, err)
				<-sem
				return err
			}
			spinner.CompleteWithMessagef("pushed %s", destination)
		}

		<-sem
		return nil
	}

	return func(call goja.FunctionCall) goja.Value {

		ctx.checkArgs(call, Preload)
		config, _ := ctx.ek.Config.LoadConfig()

		var arg = call.Argument(0)
		result := make(map[string]string)

		err := ctx.AddonCtx.vm.ExportTo(arg, &result)
		if err != nil {
			panic(err)
		}

		var wg sync.WaitGroup
		sem := make(chan struct{}, 3)
		sm := ysmrr.NewSpinnerManager()
		defer sm.Stop()
		sm.Start()

		for source, dest := range result {
			wg.Add(1)
			sem <- struct{}{}
			wg.Go(func() {
				_ = pullAndTag(source, dest, config, &wg, sm, sem)
			})
		}
		wg.Wait()
		return goja.Undefined()
	}
}

func getPrivateRegistryCredentials(registry string, config []core.MirrorRegistry) *core.PrivateRegistryCredentials {

	for i := range config {

		if config[i].UserKey != "" {
			x := strings.ReplaceAll(config[i].RegistryUrl, "https://", "")
			x = strings.ReplaceAll(x, "http://", "")

			if strings.Contains(registry, x) {
				return &core.PrivateRegistryCredentials{
					Username: config[i].UserKey,
					Password: config[i].PasswordKey,
				}
			}
		}
	}

	return nil
}
