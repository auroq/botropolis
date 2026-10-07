package main

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

type Status string

const (
	Available  Status = "available"
	CheckedOut Status = "checked-out"
)

type Lantern struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Colour       string     `json:"colour"`
	Status       Status     `json:"status"`
	Borrower     string     `json:"borrower,omitempty"`
	CheckedOutAt *time.Time `json:"checked_out_at,omitempty"`
}

var (
	ErrNotFound     = errors.New("lantern not found")
	ErrExists       = errors.New("lantern already exists")
	ErrCheckedOut   = errors.New("lantern is already checked out")
	ErrNotOut       = errors.New("lantern is not checked out")
	ErrMissingField = errors.New("missing field")
)

type Store struct {
	mu       sync.Mutex
	lanterns map[string]*Lantern
	now      func() time.Time
}

func NewStore(now func() time.Time) *Store {
	return &Store{lanterns: map[string]*Lantern{}, now: now}
}

func (s *Store) List() []*Lantern {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*Lantern, 0, len(s.lanterns))
	for _, l := range s.lanterns {
		out = append(out, l)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (s *Store) Get(id string) (Lantern, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	l, ok := s.lanterns[id]
	if !ok {
		return Lantern{}, ErrNotFound
	}
	return *l, nil
}

func (s *Store) Add(l Lantern) (Lantern, error) {
	if l.ID == "" || l.Name == "" {
		return Lantern{}, fmt.Errorf("%w: id and name are required", ErrMissingField)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.lanterns[l.ID]; ok {
		return Lantern{}, ErrExists
	}
	l.Status = Available
	l.Borrower = ""
	l.CheckedOutAt = nil
	s.lanterns[l.ID] = &l
	return l, nil
}

func (s *Store) CheckOut(id, borrower string) (Lantern, error) {
	if borrower == "" {
		return Lantern{}, fmt.Errorf("%w: borrower is required", ErrMissingField)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	l, ok := s.lanterns[id]
	if !ok {
		return Lantern{}, ErrNotFound
	}
	if l.Status == CheckedOut {
		return Lantern{}, ErrCheckedOut
	}
	at := s.now()
	l.Status = CheckedOut
	l.Borrower = borrower
	l.CheckedOutAt = &at
	return *l, nil
}

func (s *Store) Return(id string) (Lantern, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	l, ok := s.lanterns[id]
	if !ok {
		return Lantern{}, ErrNotFound
	}
	if l.Status != CheckedOut {
		return Lantern{}, ErrNotOut
	}
	l.Status = Available
	l.Borrower = ""
	l.CheckedOutAt = nil
	return *l, nil
}

func (s *Store) snapshot() []Lantern {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Lantern, 0, len(s.lanterns))
	for _, l := range s.lanterns {
		out = append(out, *l)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (s *Store) restore(ls []Lantern) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lanterns = make(map[string]*Lantern, len(ls))
	for i := range ls {
		l := ls[i]
		s.lanterns[l.ID] = &l
	}
}
