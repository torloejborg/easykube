package jsutils

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/core"

	"github.com/dop251/goja"
)

type JsUtils struct {
	ek                 *core.Ek
	vm                 *goja.Runtime
	CobraCommandHelper core.ICobraCommandHelper
	AddonRoot          string
	isNoop             bool
}

type AddonContext struct {
	addon  core.IAddon
	vm     *goja.Runtime
	IsNoop bool
	ek     *core.Ek
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

func NewJsUtils(ek *core.Ek, source core.IAddon, isNoop bool) core.IJsUtils {
	vm := goja.New()

	ac := &AddonContext{
		addon:  source,
		vm:     vm,
		IsNoop: isNoop,
		ek:     ek,
	}

	export := func(name string, action interface{}) {
		err := vm.Set(name, action)
		if err != nil {
			panic(err)
		}
	}

	ConfigureEasykubeScript(ek, ac)

	export("console", NewCons(ek).Console(isNoop))

	return &JsUtils{
		vm:        vm,
		AddonRoot: source.GetRootDir(),
		ek:        ek,
	}
}

func (jsu *JsUtils) ExecAddonScript(a core.IAddon) error {
	script := a.ReadScriptFile(jsu.ek.Fs)

	jsu.ek.Printer.FmtGreen("🔧 Processing %s", a.GetName())

	// Wrap the JavaScript execution in a deferred function
	err := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("caught JavaScript panic: %v", r)
			}
		}()

		_, err = jsu.vm.RunString(jsu.GetPseudoJsIncludes() + script)

		cfg := jsu.vm.Get("configuration")
		fmt.Println(cfg)

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
	jsScriptDir := filepath.Join(jsu.AddonRoot, constants.JsLib)
	data := make([]string, 0)
	exists, _ := afero.DirExists(jsu.ek.Fs, jsScriptDir)

	if exists {
		walkFunc := func(path string, info fs.FileInfo, err error) error {
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".js") {
				dat, err := afero.ReadFile(jsu.ek.Fs, filepath.Join(jsScriptDir, info.Name()))
				if err != nil {
					panic(err)
				}
				data = append(data, string(dat))
				clear(dat)
			}
			return nil
		}

		err := afero.Walk(jsu.ek.Fs, jsScriptDir, walkFunc)

		if err != nil {
			panic(err)
		}

		return strings.Join(data, "")
	} else {
		return ""
	}
}
