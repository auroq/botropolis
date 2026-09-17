package hooks

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	BackupSuffix   = ".botropolis-bak"
	TimeoutSeconds = 5
	commandName    = "botropolis-hook"
)

var Events = []string{
	"SessionStart", "SessionEnd", "UserPromptSubmit",
	"PreToolUse", "PostToolUse", "Notification", "Stop",
	"SubagentStart", "SubagentStop",
}

func Install(settingsPath, command string) (bool, error) {
	return edit(settingsPath, func(settings map[string]any) bool {
		hooksByEvent := hooksMap(settings)
		changed := false
		for _, event := range Events {
			entries, _ := hooksByEvent[event].([]any)
			if hasOurs(entries, command) {
				continue
			}
			hooksByEvent[event] = append(withoutOurs(entries), group(command))
			changed = true
		}
		settings["hooks"] = hooksByEvent
		return changed
	})
}

func Remove(settingsPath string) (bool, error) {
	return edit(settingsPath, func(settings map[string]any) bool {
		hooksByEvent, ok := settings["hooks"].(map[string]any)
		if !ok {
			return false
		}
		changed := false
		for event, raw := range hooksByEvent {
			entries, _ := raw.([]any)
			kept := withoutOurs(entries)
			if len(kept) == len(entries) {
				continue
			}
			changed = true
			if len(kept) == 0 {
				delete(hooksByEvent, event)
			} else {
				hooksByEvent[event] = kept
			}
		}
		if len(hooksByEvent) == 0 {
			delete(settings, "hooks")
		}
		return changed
	})
}

func Render(command string) string {
	block := map[string]any{}
	for _, event := range Events {
		block[event] = []any{group(command)}
	}
	data, _ := json.MarshalIndent(block, "", "  ")
	return string(data)
}

func edit(settingsPath string, apply func(map[string]any) bool) (bool, error) {
	original, err := os.ReadFile(settingsPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, err
	}
	settings := map[string]any{}
	if len(bytes.TrimSpace(original)) > 0 {
		if err := json.Unmarshal(original, &settings); err != nil {
			return false, err
		}
	}
	if !apply(settings) {
		return false, nil
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return false, err
	}
	if original != nil {
		if err := os.WriteFile(settingsPath+BackupSuffix, original, 0o600); err != nil {
			return false, err
		}
	}
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o700); err != nil {
		return false, err
	}
	return true, writeAtomic(settingsPath, append(data, '\n'))
}

func writeAtomic(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func hooksMap(settings map[string]any) map[string]any {
	if hooksByEvent, ok := settings["hooks"].(map[string]any); ok {
		return hooksByEvent
	}
	return map[string]any{}
}

func group(command string) map[string]any {
	return map[string]any{
		"hooks": []any{map[string]any{
			"type":    "command",
			"command": command,
			"timeout": TimeoutSeconds,
		}},
	}
}

func isOurs(entry any) bool {
	return ourCommand(entry) != ""
}

func hasOurs(entries []any, command string) bool {
	for _, entry := range entries {
		if isOurs(entry) && ourCommand(entry) == command {
			return true
		}
	}
	return false
}

func ourCommand(entry any) string {
	g, _ := entry.(map[string]any)
	inner, _ := g["hooks"].([]any)
	for _, h := range inner {
		hook, _ := h.(map[string]any)
		command, _ := hook["command"].(string)
		if strings.Contains(command, commandName) {
			return command
		}
	}
	return ""
}

func withoutOurs(entries []any) []any {
	var kept []any
	for _, entry := range entries {
		if !isOurs(entry) {
			kept = append(kept, entry)
		}
	}
	return kept
}
