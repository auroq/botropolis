package format

import "strings"

// Wrap breaks text into lines no longer than width characters, splitting on
// spaces and keeping separators such as "|" with the word before them.
func Wrap(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}
	var lines []string
	var line strings.Builder
	for _, word := range strings.Fields(text) {
		switch {
		case line.Len() == 0:
			line.WriteString(word)
		case line.Len()+1+len(word) <= width:
			line.WriteString(" ")
			line.WriteString(word)
		default:
			lines = append(lines, line.String())
			line.Reset()
			line.WriteString(word)
		}
	}
	if line.Len() > 0 || len(lines) == 0 {
		lines = append(lines, line.String())
	}
	return lines
}

// Clip shortens s to at most width characters, ending with an ellipsis when cut.
func Clip(s string, width int) string {
	if width <= 0 || len(s) <= width {
		return s
	}
	if width == 1 {
		return "…"
	}
	return s[:width-1] + "…"
}
