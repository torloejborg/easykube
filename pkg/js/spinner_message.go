package jsutils

import (
	"fmt"

	"github.com/torloejborg/easykube/pkg/textutils"
)

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
	return textutils.TruncateToTerminalWidth(fmt.Sprintf(format, a...))
}
