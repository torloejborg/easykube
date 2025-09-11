package jsutils

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/torloejborg/easykube/pkg/constants"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
)

type JsUtils struct {
	vm                 *goja.Runtime
	CobraCommandHelper *ez.CobraCommandHelperImpl
	AddonRoot          string
}

type IJsUtils interface {
	ExecAddonScript(a *ez.Addon)
}

type AddonContext struct {
	addon              *ez.Addon
	vm                 *goja.Runtime
	CobraCommandHelper *ez.CobraCommandHelperImpl
}

func (ac *AddonContext) ExportFunction(name string, action interface{}) {
	err := ac.vm.Set(name, action)
	if err != nil {
		panic(err)
	}
}

func (ac *AddonContext) NewObject() *goja.Object {
	return ac.vm.NewObject()
}

func NewJsUtils(commandHelper *ez.CobraCommandHelperImpl, source *ez.Addon) IJsUtils {
	vm := goja.New()

	ac := &AddonContext{
		addon:              source,
		vm:                 vm,
		CobraCommandHelper: commandHelper,
	}

	export := func(name string, action interface{}) {
		err := vm.Set(name, action)
		if err != nil {
			panic(err)
		}
	}

	ConfigureEasykubeScript(commandHelper, ac)

	export("console", NewCons(commandHelper).Console())
	export("_utils", NewUtils(commandHelper))

	return &JsUtils{
		vm:                 vm,
		AddonRoot:          source.RootDir,
		CobraCommandHelper: commandHelper,
	}
}

func (jsu *JsUtils) ExecAddonScript(a *ez.Addon) {
	script := a.ReadScriptFile(ez.Kube.Fs)
	ezk := ez.Kube

	// Before we execute the addon javascript, set the working directory, such that all file operations for
	// the addon will be relative to the addon directory, when we are done, go back where we came from.
	ez.PushDir(filepath.Dir(a.File))
	defer ez.PopDir()

	ezk.FmtGreen("🔧 Processing %s", a.Name)

	// Wrap the JavaScript execution in a deferred function
	err := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("caught JavaScript panic: %v", r)
			}
		}()

		_, err = jsu.vm.RunString(jsu.GetPseudoJsIncludes() + script)
		return
	}()

	if err != nil {
		ezk.FmtRed("Error in %s addon!\n", a.Name)
		ezk.FmtRed("Cause: %s\n", err)
		os.Exit(-1)
	}
}

// GetPseudoJsIncludes Concatenates all the library javascript functions found in the _jslib directory into
// one string and returns it for loading in the JS vm.
func (jsu *JsUtils) GetPseudoJsIncludes() string {
	jsScriptDir := filepath.Join(jsu.AddonRoot, constants.JS_LIB)
	data := make([]string, 0)
	if directoryExists(jsScriptDir) {
		err := filepath.Walk(jsScriptDir, func(path string, info fs.FileInfo, err error) error {
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".js") {
				dat, err := os.ReadFile(filepath.Join(jsScriptDir, info.Name()))
				if err != nil {
					panic(err)
				}
				data = append(data, string(dat))
				clear(dat)
			}
			return nil
		})
		if err != nil {
			panic(err)
		}

		return strings.Join(data, "")
	} else {
		return ""
	}
}

func directoryExists(dirName string) bool {
	info, err := os.Stat(dirName)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}
