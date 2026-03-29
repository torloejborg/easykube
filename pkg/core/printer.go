package core

type IPrinter interface {
	FmtGreen(out string, args ...any)
	FmtRed(out string, args ...any)
	FmtYellow(out string, args ...any)
	FmtVerbose(out string, args ...any)
	FmtDryRun(out string, args ...any)
}
