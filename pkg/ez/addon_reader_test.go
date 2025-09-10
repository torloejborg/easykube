package ez

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/torloejborg/easykube/ekctx"
)

func init() {
	Kube = &Toolbox{Printer: ekctx.Printer{}}

	y := &OsDetailsStub{CreateOsDetailsImpl()}
	x := &EasykubeConfigStub{CreateEasykubeConfigImpl(y)}

	Kube.UseOsDetails(y)
	Kube.UseFilesystemLayer(afero.NewMemMapFs())
	Kube.UseEasykubeConfig(x)
	Kube.UseAddonReader(CreateAddonReaderImpl(x))

}

func TestDiscoverAddons(t *testing.T) {
	Kube.MakeConfig()
	osFs := afero.NewOsFs()

	err := copyDirToMemFS(osFs, Kube.Fs, "../../test_addons/diamond", "/addons")
	if err != nil {
		t.Errorf("%v", err)
	}

	Kube.GetAddons()
	//for _, addon := range addons {
	//	t.Log(addon.Name)
	//}

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
