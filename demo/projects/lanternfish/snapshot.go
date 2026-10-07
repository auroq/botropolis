package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

func (s *Store) Save(path string) error {
	data, err := json.MarshalIndent(s.snapshot(), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func (s *Store) Load(path string) error {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var ls []Lantern
	if err := json.Unmarshal(data, &ls); err != nil {
		return fmt.Errorf("snapshot %s: %w", path, err)
	}
	s.restore(ls)
	return nil
}
