package ez

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/torloejborg/easykube/pkg/core"
	"github.com/torloejborg/easykube/pkg/resources"
)

type UtilsImpl struct {
	ek *core.Ek
}

func NewUtils(ek *core.Ek) core.IUtils {
	return UtilsImpl{ek: ek}
}

func (u UtilsImpl) HasBinary(binary string) bool {
	_, err := exec.LookPath(binary)
	if err != nil {
		return false
	}
	return true
}

func (u UtilsImpl) FileOrDirExists(path string) bool {

	path = filepath.Clean(path)

	_, err := u.ek.Fs.Stat(path)
	return err == nil || os.IsExist(err)
}

// Copies an embedded resource from src into the user configuration directory
// dest is a relative path to ~/.config/easykube
func (u UtilsImpl) CopyResourceToConfigDir(src, dest string) error {

	configDir, err := u.ek.OsDetails.GetEasykubeConfigDir()
	if nil != err {
		return err
	}

	u.ek.Fs.MkdirAll(configDir, 0755)

	destinationPath := filepath.Join(configDir, dest)
	base := filepath.Dir(destinationPath)
	u.ek.Fs.MkdirAll(base, 0755)

	stat, _ := u.ek.Fs.Stat(destinationPath)

	if stat == nil {

		f, _ := u.ek.Fs.Create(destinationPath)
		defer f.Close()

		sourceData, err := resources.AppResources.ReadFile("data/" + src)
		if nil != err {
			return err
		}

		_, err = f.WriteString(string(sourceData))
		if nil != err {
			return err
		}
	}

	return nil
}

func (u UtilsImpl) SaveFile(data string, dest string) {

	file, err := u.ek.Fs.Create(dest)
	if err != nil {
		log.Fatal(err)
	}

	_, err = file.WriteString(data)
	if err != nil {
		log.Fatal(err)
	}
	e := file.Close()
	if e != nil {
		panic(e)
	}
}

func (u UtilsImpl) SaveFileByte(data []byte, dest string) {

	file, err := u.ek.Fs.Create(dest)
	if err != nil {
		log.Fatal(err)
	}

	_, err = file.Write(data)
	if err != nil {
		log.Fatal(err)
	}
	e := file.Close()
	if e != nil {
		panic(e)
	}
}

func (u UtilsImpl) ReadPropertyFile(path string) (map[string]string, error) {
	props, err := u.ek.Fs.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(props)
	configmap := make(map[string]string)

	for scanner.Scan() {
		txt := scanner.Text()
		if len(txt) > 0 && !strings.Contains(txt, `#`) && strings.Contains(txt, `=`) {
			parts := strings.SplitN(txt, "=", 2)
			configmap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	return configmap, nil

}

func (u UtilsImpl) IntSliceToStrings(input []int) []string {
	result := make([]string, len(input))
	for i, n := range input {
		result[i] = strconv.Itoa(n)
	}
	return result
}

func (u UtilsImpl) ReadFileToBytes(filename string) ([]byte, error) {
	file, err := u.ek.Fs.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close() // Ensure the file is closed after reading

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return data, nil
}
