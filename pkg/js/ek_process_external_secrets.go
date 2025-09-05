package jsutils

import (
	"bytes"
	"fmt"
	"os"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/ekctx"
	"github.com/torloejborg/easykube/pkg/ek"
	"gopkg.in/yaml.v3"
)

func extractExternalSecrets(filePath string, ctx *ekctx.EKContext) ([]ek.ExternalSecret, error) {
	// Read the YAML file
	yamlFile, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading YAML file: %v", err)
	}
	// Create a bytes.Reader from the byte slice
	yamlReader := bytes.NewReader(yamlFile)

	// Parse the YAML content document by document
	var externalSecrets []ek.ExternalSecret
	decoder := yaml.NewDecoder(yamlReader)
	for {
		var item map[string]interface{}
		err := decoder.Decode(&item)
		if err != nil {
			break // Exit loop on error (e.g., EOF)
		}

		// Check if the document is an ExternalSecret
		if item != nil && item["kind"] == "ExternalSecret" {
			var es ek.ExternalSecret
			itemBytes, err := yaml.Marshal(item)
			if err != nil {
				ctx.Printer.FmtRed("error marshaling item: %v", err)
				continue
			}
			err = yaml.Unmarshal(itemBytes, &es)
			if err != nil {
				ctx.Printer.FmtRed("error unmarshaling item into ExternalSecret: %v", err)
				continue
			}
			externalSecrets = append(externalSecrets, es)
		}
	}
	return externalSecrets, nil
}

func (ctx *Easykube) ProcessExternalSecrets() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		k8sutils := ek.NewK8SUtils(ctx.EKContext)
		out := ctx.EKContext.Printer
		ctx.checkArgs(call, PROCESS_SECRETS)

		var arg = call.Argument(0)
		var namespace = call.Argument(1).String()
		manifest := call.Argument(2).String() // defaults to ".out.yaml"

		secretSource := make(map[string]map[string]string)

		err := ctx.AddonCtx.vm.ExportTo(arg, &secretSource)
		if err != nil {
			panic(err)
		}

		out.FmtGreen("Processing secrets and applying to %s", namespace)

		filePath := manifest
		externalSecrets, err := extractExternalSecrets(filePath, ctx.EKContext)

		for i := range externalSecrets {
			secret := k8sutils.TransformExternalSecret(externalSecrets[i], secretSource, namespace)
			k8sutils.CreateSecret(namespace, externalSecrets[i].Metadata.Name, secret.Data)
		}

		if err != nil {
			out.FmtRed("Error extracting ExternalSecrets: %v", err)
		}

		for _, es := range externalSecrets {
			out.FmtGreen("Found ExternalSecret: %s", es.Metadata.Name)
		}

		return goja.Undefined()
	}
}
