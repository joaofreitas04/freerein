// Package lockfile implements spec/lockfile.md: the machine-written
// resolved truth every later run diffs against.
package lockfile

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const Name = "harness.lock"

type LayerRef struct {
	ID     string `json:"id"`
	Source string `json:"source"`
}

type FileEntry struct {
	Layer     string   `json:"layer"`
	Component string   `json:"component"`
	Hash      string   `json:"hash"`
	Shadowed  []string `json:"shadowed"`
	Refs      []string `json:"refs"`
}

type Lock struct {
	Version    int    `json:"version"`
	ResolvedAt string `json:"resolved_at"`
	Engine     string `json:"engine"`
	Adapter    struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"adapter"`
	Layers []LayerRef            `json:"layers"`
	Files  map[string]*FileEntry `json:"files"`
}

// Read returns (nil, nil) when no lockfile exists yet.
func Read(repo string) (*Lock, error) {
	b, err := os.ReadFile(filepath.Join(repo, Name))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var l Lock
	if err := json.Unmarshal(b, &l); err != nil {
		return nil, err
	}
	return &l, nil
}

func (l *Lock) Write(repo string) error {
	b, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(repo, Name), append(b, '\n'), 0o644)
}
