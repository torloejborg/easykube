package jsutils

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
	"gopkg.in/yaml.v3"
)

func extractExternalSecrets(filePath string) ([]ez.ExternalSecret, error) {
	ezk := ez.Kube

	// Read the YAML file
	yamlFile, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading YAML file: %v", err)
	}
	// Create a bytes.Reader from the byte slice
	yamlReader := bytes.NewReader(yamlFile)

	// Parse the YAML content document by document
	var externalSecrets []ez.ExternalSecret
	decoder := yaml.NewDecoder(yamlReader)
	for {
		var item map[string]interface{}
		decodeErr := decoder.Decode(&item)
		if decodeErr != nil {
			break // Exit loop on error (e.g., EOF)
		}

		// Check if the document is an ExternalSecret
		if item != nil && item["kind"] == "ExternalSecret" {
			var es ez.ExternalSecret
			itemBytes, err := yaml.Marshal(item)
			if err != nil {
				ezk.FmtRed("error marshaling item: %v", err)
				continue
			}
			err = yaml.Unmarshal(itemBytes, &es)
			if err != nil {
				ezk.FmtRed("error unmarshaling item into ExternalSecret: %v", err)
				continue
			}
			externalSecrets = append(externalSecrets, es)
		}
	}
	return externalSecrets, nil
}

func (ctx *Easykube) ProcessExternalSecrets() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {

		ezk := ez.Kube
		if ezk.IsDryRun() {
			ezk.FmtDryRun("skipping extractExternalSecrets")
			return call.This
		}
		addonDir := filepath.Dir(ctx.AddonCtx.addon.File)

		ctx.checkArgs(call, PROCESS_SECRETS)
		var arg = call.Argument(0)
		var namespace = call.Argument(1).String()
		manifest := call.Argument(2).String() // defaults to ".out.yaml"

		secretSource := make(map[string]map[string]string)

		err := ctx.AddonCtx.vm.ExportTo(arg, &secretSource)
		if err != nil {
			panic(err)
		}

		ezk.FmtGreen("Processing secrets and applying to %s", namespace)

		pathToYaml := filepath.Join(addonDir, manifest)
		externalSecrets, err := extractExternalSecrets(pathToYaml)

		for i := range externalSecrets {
			secret := ezk.TransformExternalSecret(externalSecrets[i], secretSource, namespace)
			ezk.CreateSecret(namespace, externalSecrets[i].Metadata.Name, secret.Data)
		}

		if err != nil {
			ezk.FmtRed("Error extracting ExternalSecrets: %v", err)
		}

		for _, es := range externalSecrets {
			ezk.FmtGreen("Found ExternalSecret: %s", es.Metadata.Name)
		}

		return call.This
	}
}
