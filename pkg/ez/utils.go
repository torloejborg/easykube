package ez

import (
	"bufio"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/torloejborg/easykube/pkg/resources"
)

func FileOrDirExists(path string) bool {

	path = filepath.Clean(path)

	_, err := Kube.Fs.Stat(path)
	return err == nil || os.IsExist(err)
}

// Copies an embedded resource from src into the user configuration directory
// dest is a relative path to ~/.config/easykube
func CopyResource(src, dest string) error {

	configDir, err := Kube.GetUserConfigDir()
	if nil != err {
		return err
	}

	configDir = filepath.Join(configDir, "easykube")

	Kube.Fs.MkdirAll(configDir, 0755)

	destinationPath := filepath.Join(configDir, dest)
	stat, _ := Kube.Fs.Stat(destinationPath)

	if stat == nil {

		f, _ := Kube.Fs.Create(destinationPath)
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

func SaveFile(data string, dest string) {

	file, err := Kube.Fs.Create(dest)
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

func ReadPropertyFile(path string) (map[string]string, error) {
	props, err := Kube.Fs.OpenFile(path, os.O_RDONLY, os.ModePerm)
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

func IntSliceToStrings(input []int) []string {
	result := make([]string, len(input))
	for i, n := range input {
		result[i] = strconv.Itoa(n)
	}
	return result
}

func ReadFileToBytes(filename string) ([]byte, error) {
	file, err := Kube.Fs.Open(filename)
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
