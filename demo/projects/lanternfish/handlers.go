package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type Server struct {
	store    *Store
	snapshot string
}

func NewServer(store *Store, snapshot string) *Server {
	return &Server{store: store, snapshot: snapshot}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /lanterns", s.list)
	mux.HandleFunc("POST /lanterns", s.add)
	mux.HandleFunc("GET /lanterns/{id}", s.get)
	mux.HandleFunc("POST /lanterns/{id}/checkout", s.checkOut)
	mux.HandleFunc("POST /lanterns/{id}/return", s.giveBack)
	return mux
}

func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.store.List())
}

func (s *Server) get(w http.ResponseWriter, r *http.Request) {
	l, err := s.store.Get(r.PathValue("id"))
	s.respond(w, http.StatusOK, l, err, false)
}

func (s *Server) add(w http.ResponseWriter, r *http.Request) {
	var l Lantern
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		writeError(w, http.StatusBadRequest, "body must be a lantern as JSON")
		return
	}
	added, err := s.store.Add(l)
	s.respond(w, http.StatusCreated, added, err, true)
}

func (s *Server) checkOut(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Borrower string `json:"borrower"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, `body must be {"borrower": "..."}`)
		return
	}
	l, err := s.store.CheckOut(r.PathValue("id"), body.Borrower)
	s.respond(w, http.StatusOK, l, err, true)
}

func (s *Server) giveBack(w http.ResponseWriter, r *http.Request) {
	l, err := s.store.Return(r.PathValue("id"))
	s.respond(w, http.StatusOK, l, err, true)
}

func (s *Server) respond(w http.ResponseWriter, status int, l Lantern, err error, changed bool) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrExists), errors.Is(err, ErrCheckedOut), errors.Is(err, ErrNotOut):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, ErrMissingField):
		writeError(w, http.StatusBadRequest, err.Error())
	case err != nil:
		writeError(w, http.StatusInternalServerError, err.Error())
	default:
		if changed && s.snapshot != "" {
			if err := s.store.Save(s.snapshot); err != nil {
				log.Printf("saving snapshot: %v", err)
			}
		}
		writeJSON(w, status, l)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writing response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
