// Package demo stages a disposable Claude home from a recorded corpus,
// so the city can be filmed without any real session behind it.
package demo

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"
)

// Retime shifts every timestamp in a transcript by one delta, chosen so
// the latest lands at end. A corpus is recorded once and filmed for
// months; the power plant, the trains and the parked catalogue only see
// the last day or week, so a recording has to be moved to now.
func Retime(lines [][]byte, end time.Time) ([][]byte, error) {
	latest, ok := Latest(lines)
	if !ok {
		return nil, errors.New("no line carries a timestamp")
	}
	return Shift(lines, end.Sub(latest))
}

// Latest is the newest timestamp anywhere in the lines.
func Latest(lines [][]byte) (time.Time, bool) {
	var latest time.Time
	for _, line := range lines {
		v, err := decode(line)
		if err != nil {
			continue
		}
		walk(v, func(at time.Time) time.Time {
			if at.After(latest) {
				latest = at
			}
			return at
		})
	}
	return latest, !latest.IsZero()
}

// Shift moves every timestamp in the lines by delta. A line with none is
// returned byte for byte.
func Shift(lines [][]byte, delta time.Duration) ([][]byte, error) {
	out := make([][]byte, len(lines))
	for i, line := range lines {
		v, err := decode(line)
		if err != nil {
			return nil, err
		}
		changed := false
		walk(v, func(at time.Time) time.Time {
			changed = true
			return at.Add(delta)
		})
		if !changed {
			out[i] = line
			continue
		}
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(v); err != nil {
			return nil, err
		}
		out[i] = bytes.TrimRight(buf.Bytes(), "\n")
	}
	return out, nil
}

func decode(line []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(line))
	dec.UseNumber()
	var v any
	err := dec.Decode(&v)
	return v, err
}

func walk(v any, at func(time.Time) time.Time) {
	switch node := v.(type) {
	case map[string]any:
		for k, child := range node {
			if s, ok := child.(string); ok && k == "timestamp" {
				if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
					node[k] = at(t).UTC().Format("2006-01-02T15:04:05.000Z07:00")
				}
				continue
			}
			walk(child, at)
		}
	case []any:
		for _, child := range node {
			walk(child, at)
		}
	}
}
