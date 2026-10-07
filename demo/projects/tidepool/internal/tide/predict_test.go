package tide

import (
	"math"
	"testing"
	"time"
)

func semidiurnal() *Station {
	return &Station{
		ID:           "test",
		Location:     time.UTC,
		Datum:        2.0,
		Constituents: []Constituent{{Name: "M2", Amplitude: 1.0, Phase: 0, Speed: Speeds["M2"]}},
	}
}

func TestHeight(t *testing.T) {
	st := semidiurnal()
	period := time.Duration(360 / Speeds["M2"] * float64(time.Hour))

	cases := []struct {
		name string
		at   time.Time
		want float64
	}{
		{"at the epoch it should be datum plus amplitude", Epoch, 3.0},
		{"half a period later it should be datum minus amplitude", Epoch.Add(period / 2), 1.0},
		{"a quarter period later it should be the datum", Epoch.Add(period / 4), 2.0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := st.Height(c.at); math.Abs(got-c.want) > 1e-3 {
				t.Errorf("got %.4f, want %.4f", got, c.want)
			}
		})
	}

	t.Run("when a station has no constituents it should be the datum", func(t *testing.T) {
		if got := (&Station{Datum: 1.25}).Height(Epoch); got != 1.25 {
			t.Errorf("got %v", got)
		}
	})
}

func TestNextExtremes(t *testing.T) {
	st := semidiurnal()
	period := time.Duration(360 / Speeds["M2"] * float64(time.Hour))

	t.Run("when starting just after a high", func(t *testing.T) {
		got := st.NextExtremes(Epoch.Add(10*time.Minute), 2)

		t.Run("it should find two extremes", func(t *testing.T) {
			if len(got) != 2 {
				t.Fatalf("got %d", len(got))
			}
		})

		t.Run("it should find the low first", func(t *testing.T) {
			if got[0].Kind != Low {
				t.Errorf("got %s", got[0].Kind)
			}
		})

		t.Run("it should place the low within a minute of half a period", func(t *testing.T) {
			if d := got[0].Time.Sub(Epoch.Add(period / 2)).Abs(); d > time.Minute {
				t.Errorf("off by %s", d)
			}
		})

		t.Run("it should then find the next high", func(t *testing.T) {
			if got[1].Kind != High {
				t.Errorf("got %s", got[1].Kind)
			}
		})
	})
}

func TestTable(t *testing.T) {
	st := semidiurnal()
	start := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	rows := st.Table(start, time.Hour, 24)

	t.Run("it should produce one row per step", func(t *testing.T) {
		if len(rows) != 24 {
			t.Errorf("got %d", len(rows))
		}
	})

	t.Run("it should start at the start time", func(t *testing.T) {
		if !rows[0].Time.Equal(start) {
			t.Errorf("got %s", rows[0].Time)
		}
	})

	t.Run("it should end one step before the next day", func(t *testing.T) {
		if want := start.Add(23 * time.Hour); !rows[23].Time.Equal(want) {
			t.Errorf("got %s", rows[23].Time)
		}
	})
}
