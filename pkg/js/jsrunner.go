package jsutils

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"github.com/torloejborg/easykube/pkg/constants"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
)

type JsUtils struct {
	vm                 *goja.Runtime
	CobraCommandHelper ez.ICobraCommandHelper
	AddonRoot          string
}

type IJsUtils interface {
	ExecAddonScript(a ez.IAddon) error
}

type AddonContext struct {
	addon               ez.IAddon
	vm                  *goja.Runtime
	ICobraCommandHelper ez.ICobraCommandHelper
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

func NewJsUtils(commandHelper ez.ICobraCommandHelper, source ez.IAddon) IJsUtils {
	vm := goja.New()

	ac := &AddonContext{
		addon:               source,
		vm:                  vm,
		ICobraCommandHelper: commandHelper,
	}

	export := func(name string, action interface{}) {
		err := vm.Set(name, action)
		if err != nil {
			panic(err)
		}
	}

	ConfigureEasykubeScript(commandHelper, ac)

	export("console", NewCons(commandHelper).Console())

	return &JsUtils{
		vm:                 vm,
		AddonRoot:          source.GetRootDir(),
		CobraCommandHelper: commandHelper,
	}
}

func (jsu *JsUtils) ExecAddonScript(a ez.IAddon) error {
	script := a.ReadScriptFile(ez.Kube.Fs)
	ezk := ez.Kube

	ezk.FmtGreen("🔧 Processing %s", a.GetName())

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
		e := errors.New("in addon " + a.GetName())
		return errors.Join(e, err)
	}

	return nil
}

// GetPseudoJsIncludes Concatenates all the library javascript functions found in the _jslib directory into
// one string and returns it for loading in the JS vm.
func (jsu *JsUtils) GetPseudoJsIncludes() string {
	jsScriptDir := filepath.Join(jsu.AddonRoot, constants.JS_LIB)
	data := make([]string, 0)
	exists, _ := afero.DirExists(ez.Kube.Fs, jsScriptDir)

	if exists {
		walkFunc := func(path string, info fs.FileInfo, err error) error {
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".js") {
				dat, err := afero.ReadFile(ez.Kube.Fs, filepath.Join(jsScriptDir, info.Name()))
				if err != nil {
					panic(err)
				}
				data = append(data, string(dat))
				clear(dat)
			}
			return nil
		}

		err := afero.Walk(ez.Kube.Fs, jsScriptDir, walkFunc)

		if err != nil {
			panic(err)
		}

		return strings.Join(data, "")
	} else {
		return ""
	}
}
