package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"strings"

	"github.com/ergochat/readline"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/ez"
)

type Registry struct {
	URL      string
	Username string
	Password string
}

func prompt(promptMessage, defaultValue string, validate func(string) error) string {

	color.Set(color.White)
	fmt.Printf(" %s\n ", promptMessage)
	color.Reset()

	rl, err := readline.New("> ")
	if err != nil {
		panic(err)
	}

	for {
		line, err := rl.ReadLineWithDefault(defaultValue)
		if err != nil {
			if err.Error() == "Interrupt" {
				os.Exit(0)
			}
			panic(err)
		}

		if err := validate(line); err != nil {
			fmt.Println(err.Error())
			fmt.Println()
			continue
		}

		fmt.Println()
		return line
	}
}

func runConfigActualInteractive(cmd *cobra.Command, args []string) error {

	userConfigDir, err := ez.Kube.GetEasykubeConfigDir()
	if err != nil {
		return err
	}
	loadedCfg, _ := ez.Kube.LoadConfig()

	if loadedCfg == nil {
		loadedCfg = &ez.EasykubeConfigData{
			AddonDir:          "",
			PersistenceDir:    filepath.Join(userConfigDir, "persistence"),
			ConfigurationDir:  userConfigDir,
			ContainerRuntime:  "docker",
			ConfigurationFile: ez.Kube.PathToConfigFile(),
			MirrorRegistries: []ez.MirrorRegistry{
				{RegistryUrl: "https://registry-1.docker.io"},
				{RegistryUrl: "https://quay.io"},
				{RegistryUrl: "https://ghcr.io"},
				{RegistryUrl: "https://registry.k8s.io"},
			},
		}
	} else {

	}

	nopValidator := func(s string) error { return nil }
	yesNoValidator := func(s string) error {
		if s != "y" && s != "n" {
			return errors.New("invalid choice. Please enter 'y' or 'n'")
		}
		return nil
	}

	// Prompt for addon repository path
	addonDir := prompt("Enter the path to the addon repository:", loadedCfg.AddonDir, func(s string) error {

		fi, err := ez.Kube.Fs.Stat(s)
		if err != nil {
			return err
		}

		if !fi.IsDir() {
			return errors.New("addon dir %s is not a directory")
		}

		// detect addons

		return nil
	})

	// prompt for configuration dir
	configurationDir := prompt("Enter the path to the easykube config dir:", loadedCfg.ConfigurationDir, nopValidator)

	// prompt for configuration dir
	persistenceDir := prompt("Enter the path to the ek easykube persistence directory:", loadedCfg.PersistenceDir, nopValidator)

	// Prompt for container runtime
	containerRuntime := prompt("Which container runtime do you wish to use (docker/podman)", loadedCfg.ContainerRuntime, func(s string) error {
		if s != "docker" && s != "podman" {
			return errors.New("invalid choice. Please enter 'docker' or 'podman'")

		}
		return nil
	})

	// Prompt for private registries
	configureRegistries := strings.ToLower(
		prompt("Do you wish to configure any mirror registries? (y/n):", "y", yesNoValidator)) == "y"

	var registries []ez.MirrorRegistry
	if configureRegistries {
		for {
			// Prompt for registry URL
			registryURL := prompt("Enter URL of the mirror registry:", "", nopValidator)
			registryUsername := prompt(fmt.Sprintf("Username for %s (leave blank for no credentials)", registryURL), "", nopValidator)
			registryPassword := prompt(fmt.Sprintf("Password/token for %s (leave blank for no credentials)", registryURL), "", nopValidator)

			registries = append(registries, ez.MirrorRegistry{
				RegistryUrl: registryURL,
				UserKey:     registryUsername,
				PasswordKey: registryPassword,
			})

			// Ask if user wants to configure another registry
			if strings.ToLower(prompt("Do you want to configure another registry? (y/n):", "", yesNoValidator)) != "y" {
				break
			}
		}
	}

	for _, registry := range registries {
		loadedCfg.MirrorRegistries = append(loadedCfg.MirrorRegistries, registry)
	}

	cfg := &ez.EasykubeConfigData{
		AddonDir:          addonDir,
		PersistenceDir:    persistenceDir,
		ConfigurationDir:  configurationDir,
		ContainerRuntime:  containerRuntime,
		ConfigurationFile: ez.Kube.PathToConfigFile(),
		MirrorRegistries:  loadedCfg.MirrorRegistries,
	}

	err = ez.Kube.WriteConfig(cfg)
	if err != nil {
		return err
	}

	return nil
}
