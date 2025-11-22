package textutils

import (
	"fmt"

	"github.com/gookit/color"
	"github.com/slok/gospinner"
)

type IPrinter interface {
	FmtGreen(out string, args ...any)
	FmtRed(out string, args ...any)
	FmtYellow(out string, args ...any)
	FmtVerbose(out string, args ...any)
	FmtDryRun(out string, args ...any)
	FmtSpinner(action func() (any, error), out string, args ...any) (any, error)
}

func NewPrinter() IPrinter {
	return &PrinterImpl{}
}

type PrinterImpl struct {
}

func (p *PrinterImpl) FmtSpinner(action func() (any, error), out string, args ...any) (any, error) {
	s, _ := gospinner.NewSpinner(gospinner.Dots)

	var message = ""
	if args != nil {
		message = fmt.Sprintf(out, args...)
	} else {
		message = out
	}

	_ = s.Start(message)
	result, err := action()
	if err != nil {
		_ = s.Fail()
		return result, err
	} else {
		_ = s.Succeed()
	}
	return result, nil
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

func (p *PrinterImpl) FmtVerbose(out string, args ...any) {
	colorize(color.LightBlue, "(ver) "+out, args...)
}

func (p *PrinterImpl) FmtDryRun(out string, args ...any) {
	colorize(color.LightMagenta, "(dry) "+out, args...)
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
