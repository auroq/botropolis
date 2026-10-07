package main

import (
	"errors"
	"path/filepath"
	"testing"
	"time"
)

var noon = time.Date(2026, 9, 12, 21, 0, 0, 0, time.UTC)

func newTestStore(t *testing.T, ids ...string) *Store {
	t.Helper()
	s := NewStore(func() time.Time { return noon })
	for _, id := range ids {
		if _, err := s.Add(Lantern{ID: id, Name: "Lantern " + id}); err != nil {
			t.Fatal(err)
		}
	}
	return s
}

func TestStoreAdd(t *testing.T) {
	cases := []struct {
		name    string
		lantern Lantern
		want    error
	}{
		{"when the id is missing it should refuse", Lantern{Name: "x"}, ErrMissingField},
		{"when the name is missing it should refuse", Lantern{ID: "x"}, ErrMissingField},
		{"when the id is taken it should refuse", Lantern{ID: "a", Name: "again"}, ErrExists},
		{"when the lantern is new it should accept", Lantern{ID: "b", Name: "Bee"}, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := newTestStore(t, "a").Add(c.lantern)
			if !errors.Is(err, c.want) {
				t.Errorf("got %v, want %v", err, c.want)
			}
		})
	}

	t.Run("when the lantern arrives marked as checked out it should be stored as available", func(t *testing.T) {
		l, _ := newTestStore(t).Add(Lantern{ID: "c", Name: "Sea", Status: CheckedOut, Borrower: "someone"})
		if l.Status != Available {
			t.Errorf("got %s", l.Status)
		}
	})
}

func TestStoreCheckOut(t *testing.T) {
	t.Run("when the lantern is available", func(t *testing.T) {
		s := newTestStore(t, "a")
		l, err := s.CheckOut("a", "Wren")
		if err != nil {
			t.Fatal(err)
		}

		t.Run("it should record the borrower", func(t *testing.T) {
			if l.Borrower != "Wren" {
				t.Errorf("got %q", l.Borrower)
			}
		})

		t.Run("it should record when", func(t *testing.T) {
			if l.CheckedOutAt == nil || !l.CheckedOutAt.Equal(noon) {
				t.Errorf("got %v", l.CheckedOutAt)
			}
		})

		t.Run("and it is checked out again it should refuse", func(t *testing.T) {
			if _, err := s.CheckOut("a", "Finch"); !errors.Is(err, ErrCheckedOut) {
				t.Errorf("got %v", err)
			}
		})
	})

	t.Run("when there is no borrower it should refuse", func(t *testing.T) {
		if _, err := newTestStore(t, "a").CheckOut("a", ""); !errors.Is(err, ErrMissingField) {
			t.Errorf("got %v", err)
		}
	})

	t.Run("when the lantern does not exist it should say so", func(t *testing.T) {
		if _, err := newTestStore(t).CheckOut("nope", "Wren"); !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v", err)
		}
	})
}

func TestStoreReturn(t *testing.T) {
	t.Run("when the lantern is checked out", func(t *testing.T) {
		s := newTestStore(t, "a")
		if _, err := s.CheckOut("a", "Wren"); err != nil {
			t.Fatal(err)
		}
		l, err := s.Return("a")
		if err != nil {
			t.Fatal(err)
		}

		t.Run("it should be available again", func(t *testing.T) {
			if l.Status != Available {
				t.Errorf("got %s", l.Status)
			}
		})

		t.Run("it should forget the borrower", func(t *testing.T) {
			if l.Borrower != "" {
				t.Errorf("got %q", l.Borrower)
			}
		})
	})

	t.Run("when the lantern is not checked out it should refuse", func(t *testing.T) {
		if _, err := newTestStore(t, "a").Return("a"); !errors.Is(err, ErrNotOut) {
			t.Errorf("got %v", err)
		}
	})
}

func TestSnapshot(t *testing.T) {
	t.Run("when a store is saved and loaded into another", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "snapshot.json")
		s := newTestStore(t, "a", "b")
		if _, err := s.CheckOut("b", "Wren"); err != nil {
			t.Fatal(err)
		}
		if err := s.Save(path); err != nil {
			t.Fatal(err)
		}
		loaded := NewStore(time.Now)
		if err := loaded.Load(path); err != nil {
			t.Fatal(err)
		}

		t.Run("it should keep every lantern", func(t *testing.T) {
			if n := len(loaded.List()); n != 2 {
				t.Errorf("got %d", n)
			}
		})

		t.Run("it should keep who has what", func(t *testing.T) {
			if l, _ := loaded.Get("b"); l.Borrower != "Wren" {
				t.Errorf("got %q", l.Borrower)
			}
		})
	})

	t.Run("when there is no snapshot yet it should start empty", func(t *testing.T) {
		s := NewStore(time.Now)
		if err := s.Load(filepath.Join(t.TempDir(), "missing.json")); err != nil {
			t.Errorf("got %v", err)
		}
	})
}
