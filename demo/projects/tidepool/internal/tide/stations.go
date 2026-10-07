package tide

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"
)

type Station struct {
	ID           string
	Name         string
	Location     *time.Location
	Datum        float64
	Constituents []Constituent
}

type Stations map[string]*Station

var ErrUnknownStation = errors.New("unknown station")

func (s Stations) Get(id string) (*Station, error) {
	st, ok := s[id]
	if !ok {
		return nil, fmt.Errorf("%w: %q (known: %s)", ErrUnknownStation, id, strings.Join(s.IDs(), ", "))
	}
	return st, nil
}

func (s Stations) IDs() []string {
	ids := make([]string, 0, len(s))
	for id := range s {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func LoadDir(dir string) (Stations, error) {
	sf, err := os.Open(filepath.Join(dir, "stations.csv"))
	if err != nil {
		return nil, err
	}
	defer sf.Close()
	stations, err := ReadStations(sf, "stations.csv")
	if err != nil {
		return nil, err
	}

	cf, err := os.Open(filepath.Join(dir, "constituents.csv"))
	if err != nil {
		return nil, err
	}
	defer cf.Close()
	if err := ReadConstituents(cf, "constituents.csv", stations); err != nil {
		return nil, err
	}
	return stations, nil
}

func ReadStations(r io.Reader, name string) (Stations, error) {
	stations := Stations{}
	err := eachRow(r, name, 4, func(line int, rec []string) error {
		loc, err := time.LoadLocation(rec[2])
		if err != nil {
			return fmt.Errorf("%s:%d: timezone %q: %w", name, line, rec[2], err)
		}
		datum, err := strconv.ParseFloat(rec[3], 64)
		if err != nil {
			return fmt.Errorf("%s:%d: datum %q is not a number", name, line, rec[3])
		}
		stations[rec[0]] = &Station{ID: rec[0], Name: rec[1], Location: loc, Datum: datum}
		return nil
	})
	return stations, err
}

func ReadConstituents(r io.Reader, name string, stations Stations) error {
	return eachRow(r, name, 4, func(line int, rec []string) error {
		st, ok := stations[rec[0]]
		if !ok {
			return fmt.Errorf("%s:%d: %w %q", name, line, ErrUnknownStation, rec[0])
		}
		speed, ok := Speeds[rec[1]]
		if !ok {
			return nil
		}
		amp, err := strconv.ParseFloat(rec[2], 64)
		if err != nil {
			return fmt.Errorf("%s:%d: amplitude %q is not a number", name, line, rec[2])
		}
		phase, err := strconv.ParseFloat(rec[3], 64)
		if err != nil {
			return fmt.Errorf("%s:%d: phase %q is not a number", name, line, rec[3])
		}
		st.Constituents = append(st.Constituents, Constituent{Name: rec[1], Amplitude: amp, Phase: phase, Speed: speed})
		return nil
	})
}

func eachRow(r io.Reader, name string, fields int, fn func(line int, rec []string) error) error {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = fields
	cr.TrimLeadingSpace = true
	if _, err := cr.Read(); err != nil {
		return fmt.Errorf("%s: reading header: %w", name, err)
	}
	for {
		rec, err := cr.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		line, _ := cr.FieldPos(0)
		if err := fn(line, rec); err != nil {
			return err
		}
	}
}
