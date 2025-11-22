package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/torloejborg/easykube/pkg/ez"
)

func runConfigActual(cmd *cobra.Command, args []string) error {

	ezk := ez.Kube
	err := ezk.MakeConfig()
	if err != nil {
		return err
	}
	ezk.EditConfig()

	return nil
}

func patchConfigWithPrivateRegistryTemplate(cfg *ez.EasykubeConfigData) {
	configStanza := `  
  # Declare private registries. Whenever easykube pulls an image, and the registry name contains a substring
  # defined by repositoryMatch, the credentials are resolved by looking up the values in easykube-secrets.
  #private-registries:
  #  - repositoryMatch: partial-registry-name.io
  #    userKey: userCredentialsKey
  #    passwordKey: userCredentialPasswordKey
`

	if cfg.PrivateRegistries == nil {

		fmt.Println(cfg.ConfigurationFile)

		// patch config
		f, err := os.OpenFile(cfg.ConfigurationFile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)

		if err != nil {
			panic(err)
		}
		f.WriteString(configStanza)
	}
	fmt.Printf(configStanza)
}
