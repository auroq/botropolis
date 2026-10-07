package tide

import (
	"math"
	"time"
)

var Epoch = time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)

type Kind string

const (
	High Kind = "high"
	Low  Kind = "low"
)

type Extreme struct {
	Kind   Kind
	Time   time.Time
	Height float64
}

type Reading struct {
	Time   time.Time
	Height float64
}

func (s *Station) Height(t time.Time) float64 {
	hours := t.Sub(Epoch).Hours()
	h := s.Datum
	for _, c := range s.Constituents {
		h += c.Amplitude * math.Cos((c.Speed*hours-c.Phase)*math.Pi/180)
	}
	return h
}

const (
	searchStep    = 6 * time.Minute
	searchHorizon = 26 * time.Hour
)

func (s *Station) NextExtremes(from time.Time, n int) []Extreme {
	var out []Extreme
	prev := s.Height(from)
	rising := s.Height(from.Add(searchStep)) > prev
	for t := from.Add(searchStep); len(out) < n && t.Sub(from) <= searchHorizon; t = t.Add(searchStep) {
		h := s.Height(t)
		if rising && h < prev {
			out = append(out, s.refine(t.Add(-2*searchStep), t, High))
			rising = false
		} else if !rising && h > prev {
			out = append(out, s.refine(t.Add(-2*searchStep), t, Low))
			rising = true
		}
		prev = h
	}
	return out
}

func (s *Station) refine(lo, hi time.Time, kind Kind) Extreme {
	best := Extreme{Kind: kind, Time: lo, Height: s.Height(lo)}
	for t := lo; !t.After(hi); t = t.Add(time.Minute) {
		h := s.Height(t)
		if (kind == High && h > best.Height) || (kind == Low && h < best.Height) {
			best = Extreme{Kind: kind, Time: t, Height: h}
		}
	}
	return best
}

func (s *Station) Table(start time.Time, step time.Duration, count int) []Reading {
	rows := make([]Reading, 0, count)
	for i := range count {
		t := start.Add(time.Duration(i) * step)
		rows = append(rows, Reading{Time: t, Height: s.Height(t)})
	}
	return rows
}
