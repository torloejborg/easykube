package textutils

import (
	"strings"
)

// TrimIndent removes the smallest common leading indentation
// from every line and trims leading/trailing empty lines.
func TrimIndent(s string) string {
	lines := strings.Split(s, "\n")

	// Remove leading/trailing empty lines
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	// Find minimal indentation
	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " \t"))
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	// Remove common indentation
	for i, line := range lines {
		if len(line) >= minIndent {
			lines[i] = line[minIndent:]
		}
	}

	return strings.Join(lines, "\n")
}

// TrimMargin removes the margin prefix from every line,
// after trimming leading whitespace. Default is "|" if empty.
func TrimMargin(s string, marginPrefix string) string {
	if marginPrefix == "" {
		marginPrefix = "|"
	}

	lines := strings.Split(s, "\n")

	// Remove leading/trailing empty lines
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	// Strip margin prefix
	for i, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trimmed, marginPrefix) {
			lines[i] = trimmed[len(marginPrefix):]
		} else {
			lines[i] = trimmed
		}
	}

	return strings.Join(lines, "\n")
}
