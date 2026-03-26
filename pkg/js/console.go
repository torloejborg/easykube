package jsutils

import (
	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/ez"
)

type ConsImpl struct {
}

type ICons interface {
	Console(noop bool) map[string]func(goja.FunctionCall) goja.Value
	console() map[string]func(goja.FunctionCall) goja.Value
}

func NewCons(ctx ez.ICobraCommandHelper) ICons {
	return &ConsImpl{}
}

func (cons *ConsImpl) Console(noop bool) map[string]func(goja.FunctionCall) goja.Value {
	if noop {

		m := map[string]func(goja.FunctionCall) goja.Value{
			"log": func(call goja.FunctionCall) goja.Value {
				return goja.Undefined()
			},
			"info": func(call goja.FunctionCall) goja.Value {
				return goja.Undefined()
			},
			"warn": func(call goja.FunctionCall) goja.Value {
				return goja.Undefined()
			}, "error": func(call goja.FunctionCall) goja.Value {
				return goja.Undefined()
			},
		}

		return m
	}

	return cons.console()
}

func (cons *ConsImpl) console() map[string]func(goja.FunctionCall) goja.Value {
	ezk := ez.Kube.IPrinter

	return map[string]func(goja.FunctionCall) goja.Value{
		"log": func(call goja.FunctionCall) goja.Value {
			for _, arg := range call.Arguments {
				ezk.FmtGreen(arg.String())
			}
			return goja.Undefined()
		},
		"info": func(call goja.FunctionCall) goja.Value {
			for _, arg := range call.Arguments {
				ezk.FmtGreen(arg.String())
			}
			return goja.Undefined()
		},
		"warn": func(call goja.FunctionCall) goja.Value {
			for _, arg := range call.Arguments {
				ezk.FmtYellow(arg.String())
			}
			return goja.Undefined()
		}, "error": func(call goja.FunctionCall) goja.Value {
			for _, arg := range call.Arguments {
				ezk.FmtRed(arg.String())
			}
			return goja.Undefined()
		},
	}

}
