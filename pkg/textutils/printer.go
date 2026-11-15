package textutils

import (
	"fmt"
	"os"

	"github.com/gookit/color"
	"github.com/torloejborg/easykube/pkg/core"
	"golang.org/x/term"
)

func NewPrinter() core.IPrinter {
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
		fmt.Println(fmt.Sprintf(TruncateToTerminalWidth(out), args...))
		_, err := color.Reset()
		if err != nil {
			panic(err)
		}
	}()
}

// spinnerMessageReserve is the number of terminal columns reserved for the
// spinner character, its separating space and a column of slack, so that
// the spinner glyph plus its message always fit on a single terminal line.
const SpinnerMessageReserve = 4

// truncateToTerminalWidth shortens msg so that it fits on a single terminal
// line, falling back to a width of 80 columns when the terminal size cannot
// be determined (for example when output is piped or in tests).
func TruncateToTerminalWidth(msg string) string {
	width := 0
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width = w
	}
	if width <= 0 {
		width = 80
	}

	maxWidth := max(width-SpinnerMessageReserve, 10)

	runes := []rune(msg)
	if len(runes) <= maxWidth {
		return msg
	}

	if maxWidth <= 1 {
		return string(runes[:1])
	}

	return string(runes[:maxWidth-1]) + "…"
}
