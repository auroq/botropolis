package control

import (
	"path/filepath"
	"strings"
)

var knownTerminals = []string{"foot", "kitty", "alacritty", "wezterm", "gnome-terminal", "konsole", "terminator", "xterm"}

// KnownTerminals is the list tried when nothing names a terminal.
func KnownTerminals() []string {
	return append([]string(nil), knownTerminals...)
}

var terminalPrefix = map[string][]string{
	"foot":           {},
	"kitty":          {},
	"alacritty":      {"-e"},
	"wezterm":        {"start", "--"},
	"gnome-terminal": {"--"},
	"konsole":        {"-e"},
	"xterm":          {"-e"},
	// terminator's -e takes one string; -x takes the rest of argv.
	"terminator": {"-x"},
}

func TerminalCommand(getenv func(string) string, onPath func(string) bool, command []string) ([]string, error) {
	if custom := strings.Fields(getenv("BOTROPOLIS_TERMINAL")); len(custom) > 0 {
		return append(custom, command...), nil
	}
	if terminal := getenv("TERMINAL"); terminal != "" {
		return withConvention(terminal, command), nil
	}
	for _, terminal := range knownTerminals {
		if onPath(terminal) {
			return withConvention(terminal, command), nil
		}
	}
	return nil, errNoTerminal
}

func withConvention(terminal string, command []string) []string {
	prefix, known := terminalPrefix[filepath.Base(terminal)]
	if !known {
		prefix = []string{"-e"}
	}
	argv := append([]string{terminal}, prefix...)
	return append(argv, command...)
}
