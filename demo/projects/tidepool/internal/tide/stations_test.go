package tide

import (
	"errors"
	"strings"
	"testing"
)

const stationsCSV = `id,name,timezone,datum_m
north,North Jetty,UTC,1.50
`

func TestReadStations(t *testing.T) {
	t.Run("when the file is well formed", func(t *testing.T) {
		stations, err := ReadStations(strings.NewReader(stationsCSV), "stations.csv")
		if err != nil {
			t.Fatal(err)
		}

		t.Run("it should key stations by id", func(t *testing.T) {
			if stations["north"].Name != "North Jetty" {
				t.Errorf("got %q", stations["north"].Name)
			}
		})

		t.Run("it should read the datum", func(t *testing.T) {
			if stations["north"].Datum != 1.5 {
				t.Errorf("got %v", stations["north"].Datum)
			}
		})
	})

	t.Run("when a timezone is not recognised", func(t *testing.T) {
		_, err := ReadStations(strings.NewReader("id,name,timezone,datum_m\nx,X,Mars/Olympus,1\n"), "stations.csv")

		t.Run("it should name the file and line", func(t *testing.T) {
			if err == nil || !strings.Contains(err.Error(), "stations.csv:2") {
				t.Errorf("got %v", err)
			}
		})
	})
}

func TestReadConstituents(t *testing.T) {
	load := func(t *testing.T, body string) (Stations, error) {
		t.Helper()
		stations, err := ReadStations(strings.NewReader(stationsCSV), "stations.csv")
		if err != nil {
			t.Fatal(err)
		}
		return stations, ReadConstituents(strings.NewReader(body), "constituents.csv", stations)
	}

	t.Run("when the file is well formed", func(t *testing.T) {
		stations, err := load(t, "station,constituent,amplitude_m,phase_deg\nnorth,M2,1.2,90\nnorth,K1,0.3,10\n")
		if err != nil {
			t.Fatal(err)
		}

		t.Run("it should attach every constituent to its station", func(t *testing.T) {
			if got := len(stations["north"].Constituents); got != 2 {
				t.Errorf("got %d constituents", got)
			}
		})

		t.Run("it should look up the angular speed", func(t *testing.T) {
			if got := stations["north"].Constituents[0].Speed; got != Speeds["M2"] {
				t.Errorf("got %v", got)
			}
		})
	})

	t.Run("when a row names a station that does not exist", func(t *testing.T) {
		_, err := load(t, "station,constituent,amplitude_m,phase_deg\nsouth,M2,1.2,90\n")

		t.Run("it should report an unknown station", func(t *testing.T) {
			if !errors.Is(err, ErrUnknownStation) {
				t.Errorf("got %v", err)
			}
		})
	})

	t.Run("when an amplitude is not a number", func(t *testing.T) {
		_, err := load(t, "station,constituent,amplitude_m,phase_deg\nnorth,M2,1.2,90\nnorth,S2,big,90\n")

		t.Run("it should name the file and line", func(t *testing.T) {
			if err == nil || !strings.Contains(err.Error(), "constituents.csv:3") {
				t.Errorf("got %v", err)
			}
		})
	})
}

func TestStationsGet(t *testing.T) {
	t.Run("when the station is missing", func(t *testing.T) {
		_, err := Stations{}.Get("nowhere")

		t.Run("it should wrap ErrUnknownStation", func(t *testing.T) {
			if !errors.Is(err, ErrUnknownStation) {
				t.Errorf("got %v", err)
			}
		})
	})
}
