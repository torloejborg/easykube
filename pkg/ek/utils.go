package ek

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/afero"
	"github.com/torloejborg/easykube/pkg/resources"
)

var dirstack = &Stack[string]{}

type Utils struct {
	Fs afero.Fs
}

func (u Utils) PushDir(dir string) {
	//fmt.Println("pushdir " + dir)
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
	dirstack.Push(dir)
}

func (u Utils) PopDir() {
	dir, result := dirstack.Pop()
	if result {
		//fmt.Println("popped dir " + dir)
		err := os.Chdir(dir)
		if err != nil {
			panic(err)
		}
	}
}

func (u Utils) FileOrDirExists(path string) bool {

	path = filepath.Clean(path)

	_, err := u.Fs.Stat(path)
	return err == nil || os.IsExist(err)
}

func (u Utils) Check(err error) {
	log.Fatal(err)
}

// Copies an embedded resource from src into the user configuration directory
// dest is a relative path to ~/.config/easykube
func (u Utils) CopyResource(src, dest string) {

	configDir, err := os.UserConfigDir()
	if nil != err {
		panic(err)
	}

	configDir = filepath.Join(configDir, "easykube")

	u.Fs.MkdirAll(configDir, 0755)

	destinationPath := filepath.Join(configDir, dest)
	stat, _ := u.Fs.Stat(destinationPath)

	if stat == nil {

		f, _ := u.Fs.Create(destinationPath)
		defer f.Close()

		sourceData, err := resources.AppResources.ReadFile("data/" + src)
		if nil != err {
			panic(err)
		}

		_, err = f.WriteString(string(sourceData))
		if nil != err {
			panic(err)
		}
	}
}

func (u Utils) SaveFile(data string, dest string) {

	file, err := u.Fs.Create(dest)
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

func (u Utils) ReadPropertyFile(path string) (map[string]string, error) {
	props, err := u.Fs.OpenFile(path, os.O_RDONLY, os.ModePerm)
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

func (u Utils) IntSliceToStrings(input []int) []string {
	result := make([]string, len(input))
	for i, n := range input {
		result[i] = strconv.Itoa(n)
	}
	return result
}
