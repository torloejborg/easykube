package cmd

import (
	"io"
	"os"
	"slices"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	jsutils "github.com/torloejborg/easykube/pkg/js"
	"github.com/torloejborg/easykube/pkg/resources"
)

// destroyCmd represents the destroy command
var jsDocCmd = &cobra.Command{
	Use:   "schema",
	Short: "generate a typescript schema",
	Long:  `generate a typescript schema which documents scriptable easykube functions`,
	RunE: func(cmd *cobra.Command, args []string) error {

		f, err := os.Create("easykube.d.ts")
		if err != nil {
			return err
		}

		docTemplate, _ := resources.AppResources.ReadFile("data/schema.template")
		sb := new(strings.Builder)
		t := new(template.Template)

		_ = template.Must(t.Parse(string(docTemplate))).Execute(sb,
			slices.Concat(
				jsutils.KeyValueDoc,
				jsutils.CreateSecretDoc,
				jsutils.CopyToDoc,
				jsutils.WaitForCRDDoc,
				jsutils.SkopeoPreloadDoc,
				jsutils.AndThenApplyDoc,
				jsutils.AddonDirDoc,
				jsutils.EkDocKustomize,
				jsutils.EkDocEnv,
				jsutils.PreloadImagesDoc))

		io.WriteString(f, sb.String())

		err = f.Close()
		if err != nil {
			return err
		}

		return nil

	},
}

func init() {
	jsDocCmd.Flags().StringP("name", "n", "easykube.d.ts", "specify output filename")
	rootCmd.AddCommand(jsDocCmd)
}
