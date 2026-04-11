package jsutils

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/core"
	"gopkg.in/yaml.v3"
)

func (ctx *Easykube) extractExternalSecrets(filePath string) ([]core.ExternalSecret, error) {

	// Read the YAML file
	yamlFile, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading YAML file: %v", err)
	}
	// Create a bytes.Reader from the byte slice
	yamlReader := bytes.NewReader(yamlFile)

	// Parse the YAML content document by document
	var externalSecrets []core.ExternalSecret
	decoder := yaml.NewDecoder(yamlReader)
	for {
		var item map[string]interface{}
		decodeErr := decoder.Decode(&item)
		if decodeErr != nil {
			break // Exit loop on error (e.g., EOF)
		}

		// Check if the document is an ExternalSecret
		if item != nil && item["kind"] == "ExternalSecret" {
			var es core.ExternalSecret
			itemBytes, err := yaml.Marshal(item)
			if err != nil {
				ctx.ek.Printer.FmtRed("error marshaling item: %v", err)
				continue
			}
			err = yaml.Unmarshal(itemBytes, &es)
			if err != nil {
				ctx.ek.Printer.FmtRed("error unmarshaling item into ExternalSecret: %v", err)
				continue
			}
			externalSecrets = append(externalSecrets, es)
		}
	}
	return externalSecrets, nil
}

func (ctx *Easykube) ProcessExternalSecrets(noop bool) func(goja.FunctionCall) goja.Value {
	if noop {
		return NoopFunc()
	}
	return ctx.processExternalSecrets()
}

func (ctx *Easykube) processExternalSecrets() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {

		if ctx.ek.CommandContext.IsDryRun() {
			ctx.ek.Printer.FmtDryRun("skipping extractExternalSecrets")
			return call.This
		}
		addonDir := filepath.Dir(ctx.AddonCtx.addon.GetAddonFile())

		ctx.checkArgs(call, ProcessSecrets)
		var arg = call.Argument(0)
		var namespace = call.Argument(1).String()
		manifest := call.Argument(2).String() // defaults to ".out.yaml"

		secretSource := make(map[string]map[string]string)

		err := ctx.AddonCtx.vm.ExportTo(arg, &secretSource)
		if err != nil {
			panic(err)
		}

		ctx.ek.Printer.FmtGreen("Processing secrets and applying to %s", namespace)

		pathToYaml := filepath.Join(addonDir, manifest)
		externalSecrets, err := ctx.extractExternalSecrets(pathToYaml)
		datasource := ctx.AddonCtx.addon.GetName()

		for i := range externalSecrets {
			secret := ctx.ek.Kubernetes.TransformExternalSecret(externalSecrets[i], secretSource, datasource, namespace)
			if err := ctx.ek.Kubernetes.CreateSecret(namespace, externalSecrets[i].Metadata.Name, secret.Data); err != nil {
				ctx.ek.Printer.FmtRed("error creating secret: %v", err)
				panic(err)
			}
		}

		if err != nil {
			ctx.ek.Printer.FmtRed("Error extracting ExternalSecrets: %v", err)
		}

		for _, es := range externalSecrets {
			ctx.ek.Printer.FmtGreen("Found ExternalSecret: %s", es.Metadata.Name)
		}

		return call.This
	}
}
