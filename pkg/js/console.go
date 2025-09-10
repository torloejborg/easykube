package jsutils

import (
	"github.com/dop251/goja"
	"github.com/torloejborg/easykube/ekctx"
	"github.com/torloejborg/easykube/pkg/ez"
)

type ConsImpl struct {
	out *ekctx.Printer
}

type ICons interface {
	Console() map[string]func(goja.FunctionCall) goja.Value
}

func NewCons(ctx *ekctx.EKContext) ICons {
	return &ConsImpl{out: &ez.Kube.Printer}
}

func (cons *ConsImpl) Console() map[string]func(goja.FunctionCall) goja.Value {

	return map[string]func(goja.FunctionCall) goja.Value{
		"log": func(call goja.FunctionCall) goja.Value {
			for _, arg := range call.Arguments {
				cons.out.FmtGreen(arg.String())
			}
			return goja.Undefined()
		},
		"info": func(call goja.FunctionCall) goja.Value {
			for _, arg := range call.Arguments {
				cons.out.FmtGreen(arg.String())
			}
			return goja.Undefined()
		},
		"warn": func(call goja.FunctionCall) goja.Value {
			for _, arg := range call.Arguments {
				cons.out.FmtYellow(arg.String())
			}
			return goja.Undefined()
		}, "error": func(call goja.FunctionCall) goja.Value {
			for _, arg := range call.Arguments {
				cons.out.FmtRed(arg.String())
			}
			return goja.Undefined()
		},
	}

}
