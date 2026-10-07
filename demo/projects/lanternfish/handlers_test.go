package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandlers(t *testing.T) {
	type call struct {
		method, path, body string
	}
	do := func(t *testing.T, srv http.Handler, c call) *httptest.ResponseRecorder {
		t.Helper()
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, httptest.NewRequest(c.method, c.path, strings.NewReader(c.body)))
		return rec
	}
	setup := func(t *testing.T) (http.Handler, string) {
		t.Helper()
		path := filepath.Join(t.TempDir(), "snapshot.json")
		return NewServer(newTestStore(t, "amber-01"), path).Routes(), path
	}

	statuses := []struct {
		name string
		call call
		want int
	}{
		{"when listing it should succeed", call{"GET", "/lanterns", ""}, http.StatusOK},
		{"when getting a known lantern it should succeed", call{"GET", "/lanterns/amber-01", ""}, http.StatusOK},
		{"when getting an unknown lantern it should be not found", call{"GET", "/lanterns/ghost", ""}, http.StatusNotFound},
		{"when adding a lantern it should be created", call{"POST", "/lanterns", `{"id":"jade-01","name":"Jade Moth"}`}, http.StatusCreated},
		{"when adding a duplicate it should conflict", call{"POST", "/lanterns", `{"id":"amber-01","name":"Again"}`}, http.StatusConflict},
		{"when adding garbage it should be a bad request", call{"POST", "/lanterns", `{`}, http.StatusBadRequest},
		{"when checking out with a borrower it should succeed", call{"POST", "/lanterns/amber-01/checkout", `{"borrower":"Wren"}`}, http.StatusOK},
		{"when checking out without a borrower it should be a bad request", call{"POST", "/lanterns/amber-01/checkout", `{}`}, http.StatusBadRequest},
		{"when returning a lantern that is in it should conflict", call{"POST", "/lanterns/amber-01/return", ""}, http.StatusConflict},
	}
	for _, c := range statuses {
		t.Run(c.name, func(t *testing.T) {
			srv, _ := setup(t)
			if rec := do(t, srv, c.call); rec.Code != c.want {
				t.Errorf("got %d, want %d: %s", rec.Code, c.want, rec.Body)
			}
		})
	}

	t.Run("when a lantern is checked out and returned", func(t *testing.T) {
		srv, path := setup(t)
		do(t, srv, call{"POST", "/lanterns/amber-01/checkout", `{"borrower":"Wren"}`})
		rec := do(t, srv, call{"POST", "/lanterns/amber-01/return", ""})

		t.Run("it should report it available", func(t *testing.T) {
			var l Lantern
			if err := json.Unmarshal(rec.Body.Bytes(), &l); err != nil || l.Status != Available {
				t.Errorf("got %s (%v)", rec.Body, err)
			}
		})

		t.Run("it should have written a snapshot", func(t *testing.T) {
			s := NewStore(nil)
			if err := s.Load(path); err != nil || len(s.List()) != 1 {
				t.Errorf("got %d lanterns (%v)", len(s.List()), err)
			}
		})
	})
}
