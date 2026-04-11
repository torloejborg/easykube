package jsutils

import (
	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/pkg/core"
)

type ConsImpl struct {
	ek *core.Ek
}

type ICons interface {
	Console(noop bool) map[string]func(goja.FunctionCall) goja.Value
	console() map[string]func(goja.FunctionCall) goja.Value
}

func NewCons(ek *core.Ek) ICons {
	return &ConsImpl{
		ek: ek,
	}
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
			},
			"error": func(call goja.FunctionCall) goja.Value {
				return goja.Undefined()
			},
		}

		return m
	}

	return cons.console()
}

func (cons *ConsImpl) console() map[string]func(goja.FunctionCall) goja.Value {
	printer := cons.ek.Printer

	return map[string]func(goja.FunctionCall) goja.Value{
		"log": func(call goja.FunctionCall) goja.Value {
			for _, arg := range call.Arguments {
				printer.FmtGreen(arg.String())
			}
			return goja.Undefined()
		},
		"info": func(call goja.FunctionCall) goja.Value {
			for _, arg := range call.Arguments {
				printer.FmtGreen(arg.String())
			}
			return goja.Undefined()
		},
		"warn": func(call goja.FunctionCall) goja.Value {
			for _, arg := range call.Arguments {
				printer.FmtYellow(arg.String())
			}
			return goja.Undefined()
		}, "error": func(call goja.FunctionCall) goja.Value {
			for _, arg := range call.Arguments {
				printer.FmtRed(arg.String())
			}
			return goja.Undefined()
		},
	}

}
