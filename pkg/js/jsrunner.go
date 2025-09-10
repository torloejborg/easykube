package jsutils

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/torloejborg/easykube/ekctx"
	"github.com/torloejborg/easykube/pkg/constants"

	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
)

type JsUtils struct {
	vm        *goja.Runtime
	EKContext *ekctx.EKContext
	AddonRoot string
}

type IJsUtils interface {
	ExecAddonScript(a *ez.Addon)
}

type AddonContext struct {
	addon     *ez.Addon
	vm        *goja.Runtime
	EKContext *ekctx.EKContext
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

func NewJsUtils(ctx *ekctx.EKContext, source *ez.Addon) IJsUtils {

	vm := goja.New()

	ac := &AddonContext{
		addon:     source,
		vm:        vm,
		EKContext: ctx,
	}

	export := func(name string, action interface{}) {
		err := vm.Set(name, action)
		if err != nil {
			panic(err)
		}
	}

	ConfigureEasykubeScript(ctx, ac)

	export("console", NewCons(ctx).Console())
	export("_utils", NewUtils(ctx))

	return &JsUtils{
		vm:        vm,
		AddonRoot: source.RootDir,
		EKContext: ctx,
	}
}

func (jsu *JsUtils) ExecAddonScript(a *ez.Addon) {
	script := a.ReadScriptFile(ez.Kube.Fs)

	// Before we execute the addon javascript, set the working directory, such that all fileoperations for
	// the addon will be relative to the addon directory, when we are done, go back where we came from.

	ez.PushDir(filepath.Dir(a.File))
	defer ez.PopDir()

	ez.Kube.FmtGreen("🔧 Processing %s", a.Name)

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
		fmt.Printf("Error in %s addon!\n", a.Name)
		fmt.Printf("Cause: %s\n", err)
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
