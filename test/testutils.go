package test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

func CopyTestAddonToMemFs(addonDir, testAddonName, dest string, destFs afero.Fs) {

	osfs := afero.NewOsFs()

	err := copyDirToMemFS(osfs, destFs, filepath.Join(addonDir, testAddonName), filepath.Join(dest, testAddonName))
	if err != nil {
		panic(err)
	}
	err = copyDirToMemFS(osfs, destFs, filepath.Join(addonDir, "__jslib"), filepath.Join(dest, "__jslib"))
	if err != nil {
		panic(err)
	}
}

func copyDirToMemFS(osFs, memFs afero.Fs, srcDir, dstDir string) error {
	return afero.Walk(osFs, srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Calculate the relative path
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// Create the destination path in the memory filesystem
		dstPath := filepath.Join(dstDir, relPath)

		// Ensure the destination directory exists
		if err := memFs.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}

		// Copy the file
		return copyFileToMemFS(osFs, memFs, path, dstPath)
	})
}

func copyFileToMemFS(osFs afero.Fs, memFs afero.Fs, srcPath, dstPath string) error {
	// Open the source file from the OS filesystem
	srcFile, err := osFs.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create the destination file in the memory filesystem
	dstFile, err := memFs.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy the contents
	_, err = io.Copy(dstFile, srcFile)
	return err
}

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
