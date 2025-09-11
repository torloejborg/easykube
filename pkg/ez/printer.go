package ez

import (
	"fmt"

	"github.com/gookit/color"
)

type IPrinter interface {
	FmtGreen(out string, args ...any)
	FmtRed(out string, args ...any)
	FmtYellow(out string, args ...any)
}

func NewPrinter() IPrinter {
	return &PrinterImpl{}
}

type PrinterImpl struct {
}

func (p *PrinterImpl) FmtGreen(out string, args ...any) {
	colorize(color.Green, ""+out, args...)
}

func (p *PrinterImpl) FmtRed(out string, args ...any) {
	colorize(color.Red, ""+out, args...)
}

func (p *PrinterImpl) FmtYellow(out string, args ...any) {
	colorize(color.Yellow, "⚠ "+out, args...)
}

func colorize(col color.Color, out string, args ...any) {
	_, err := color.Set(col)
	if err != nil {
		panic(err)
	}

	defer func() {
		fmt.Println(fmt.Sprintf(out, args...))
		_, err := color.Reset()
		if err != nil {
			panic(err)
		}
	}()
}
