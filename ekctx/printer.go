package ekctx

import (
	"fmt"
	"github.com/gookit/color"
)

type Printer struct {
}

func (p *Printer) FmtGreen(out string, args ...any) {
	colorize(color.Green, ""+out, args...)
}

func (p *Printer) FmtRed(out string, args ...any) {
	colorize(color.Red, ""+out, args...)
}

func (p *Printer) FmtYellow(out string, args ...any) {
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
