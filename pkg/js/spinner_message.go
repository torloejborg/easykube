package jsutils

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// spinnerMessageReserve is the number of terminal columns reserved for the
// spinner character, its separating space and a column of slack, so that
// the spinner glyph plus its message always fit on a single terminal line.
const spinnerMessageReserve = 4

// spinnerMessage formats a spinner message and truncates it so that it
// always fits on a single terminal line.
//
// ysmrr assumes that every spinner occupies exactly one terminal line. If a
// message wraps onto a second line, the rendered block becomes taller than
// the number of spinners, and ysmrr's cursor-up escape sequence then
// under-counts. The render loop drifts, and once the top of the spinner
// block has scrolled off-screen the cursor clamps to the top of the
// terminal. From that point every frame is painted top-down over the
// existing console output.
func spinnerMessage(format string, a ...any) string {
	return truncateToTerminalWidth(fmt.Sprintf(format, a...))
}

// truncateToTerminalWidth shortens msg so that it fits on a single terminal
// line, falling back to a width of 80 columns when the terminal size cannot
// be determined (for example when output is piped or in tests).
func truncateToTerminalWidth(msg string) string {
	width := 0
	if w, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width = w
	}
	if width <= 0 {
		width = 80
	}

	max := width - spinnerMessageReserve
	if max < 10 {
		max = 10
	}

	runes := []rune(msg)
	if len(runes) <= max {
		return msg
	}

	if max <= 1 {
		return string(runes[:1])
	}

	return string(runes[:max-1]) + "…"
}
