package cmd

import (
	"errors"
	"fmt"

	"strings"

	"github.com/chzyer/readline"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/ez"
)

type Registry struct {
	URL      string
	Username string
	Password string
}

func prompt(p, deflt string, validate func(string) error) string {

	color.Set(color.White)
	fmt.Printf(" %s\n ", p)
	color.Reset()

	rl, err := readline.New("> " + deflt)
	if err != nil {
		panic(err)
	}

	result, _ := rl.Readline()

	err = validate(string(result))

	if err == nil {
		fmt.Println()
		return result
	}

	fmt.Println(err.Error())
	fmt.Println()

	return prompt(p, "", validate)

}

func runConfigActualInteractive(cmd *cobra.Command, args []string) error {
	noValidate := func(s string) error { return nil }

	yesNoValidator := func(s string) error {
		if s != "y" && s != "n" {
			return errors.New("Invalid choice. Please enter 'y' or 'n'")
		}
		return nil
	}

	// Prompt for addon repository path
	addonRepoPath := prompt("Please enter the path to the addon repository:", "", func(s string) error {

		fi, err := ez.Kube.Fs.Stat(s)
		if err != nil {
			return err
		}

		if !fi.IsDir() {
			return errors.New("addon dir %s is not a directory")
		}

		return nil
	})

	// Prompt for container runtime
	containerRuntime := prompt("Which container runtime do you wish to use (docker/podman)", "docker", func(s string) error {
		if s != "docker" && s != "podman" {
			return errors.New("Invalid choice. Please enter 'docker' or 'podman'")

		}
		return nil
	})

	// Prompt for private registries
	configureRegistries := strings.ToLower(
		prompt("Do you wish to configure any private registries? (y/n):", "", yesNoValidator)) == "y"

	var registries []Registry
	if configureRegistries {
		for {
			// Prompt for registry URL
			registryURL := prompt("Please enter the URL of the registry:", "", noValidate)
			registryUsername := prompt("Please enter the username for the registry: ", "", noValidate)
			registryPassword := prompt("Please enter the password for the registry: ", "", noValidate)

			registries = append(registries, Registry{
				URL:      registryURL,
				Username: registryUsername,
				Password: registryPassword,
			})

			// Ask if user wants to configure another registry

			if strings.ToLower(prompt("Do you want to configure another registry? (y/n):", "", yesNoValidator)) != "y" {
				break
			}
		}
	}

	// Print collected data
	fmt.Println("\nCollected Data:")
	fmt.Printf("Addon Repository Path: %s\n", addonRepoPath)
	fmt.Printf("Container Runtime: %s\n", containerRuntime)
	fmt.Printf("Configure Private Registries: %t\n", configureRegistries)
	if configureRegistries {
		for i, reg := range registries {
			fmt.Printf("Registry %d:\n", i+1)
			fmt.Printf("  URL: %s\n", reg.URL)
			fmt.Printf("  Username: %s\n", reg.Username)
			fmt.Printf("  Password: %s\n", strings.Repeat("*", len(reg.Password)))
		}
	}

	return nil
}
