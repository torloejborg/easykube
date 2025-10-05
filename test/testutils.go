package test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

func PrintFiles(fs afero.Fs, dir string) error {

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			rel, _ := filepath.Rel(dir, path)
			fmt.Println(rel)
		} else {

		}
		return nil
	}

	return afero.Walk(fs, dir, walkFn)
}
