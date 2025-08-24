package ek

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/torloj/easykube/pkg/resources"
)

var dirstack = &Stack[string]{}

func PushDir(dir string) {
	//fmt.Println("pushdir " + dir)
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
	dirstack.Push(dir)
}

func PopDir() {
	dir, result := dirstack.Pop()
	if result {
		//fmt.Println("popped dir " + dir)
		err := os.Chdir(dir)
		if err != nil {
			panic(err)
		}
	}
}

func FileOrDirExists(path string) bool {

	path = filepath.Clean(path)

	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func Check(err error) {
	log.Fatal(err)
}

// Copies an embedded resource from src into the user configuration directory
// dest is a relative path to ~/.config/easykube
func CopyResource(src, dest string) {

	configDir, err := os.UserConfigDir()
	if nil != err {
		panic(err)
	}

	configDir = filepath.Join(configDir, "easykube")

	os.MkdirAll(configDir, 0755)

	destinationPath := filepath.Join(configDir, dest)
	stat, _ := os.Stat(destinationPath)

	if stat == nil {

		f, _ := os.Create(destinationPath)
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

func SaveFile(data string, dest string) {

	file, err := os.Create(dest)
	if err != nil {
		log.Fatal(err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	_, err = file.WriteString(data)
	if err != nil {
		log.Fatal(err)
	}
}

// Returns a sorted list of string keys in a map
func Keys[K any](in map[string]K) []string {
	result := make([]string, 0)

	for k, _ := range in {
		result = append(result, k)
	}

	slices.Sort(result)
	return result
}

func ReadPropertyFile(path string) (map[string]string, error) {
	props, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
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
