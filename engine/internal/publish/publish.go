// Package publish implements the author-side registry generator
// (spec/registry.md): deterministic sha256-pinned archives plus the
// static index.json. See cmd/rein-publish for the CLI.
package publish

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/joaofreitas04/freerein/engine/internal/component"
)

type release struct {
	URL         string `json:"url"`
	Sha256      string `json:"sha256"`
	Kind        string `json:"kind"`
	Description string `json:"description,omitempty"`
}

type index struct {
	Components map[string]map[string]release `json:"components"`
}

// Run publishes each component directory into the registry at out.
func Run(out string, dirs []string) error {
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	idx := index{Components: map[string]map[string]release{}}
	if b, err := os.ReadFile(filepath.Join(out, "index.json")); err == nil {
		if err := json.Unmarshal(b, &idx); err != nil {
			return fmt.Errorf("existing index.json is invalid: %w", err)
		}
	}
	for _, dir := range dirs {
		parent, base := filepath.Split(filepath.Clean(dir))
		if parent == "" {
			parent = "."
		}
		c, err := component.LoadDir(os.DirFS(filepath.Clean(parent)), base)
		if err != nil {
			return err // publishing enforces the same manifest contract installing does
		}
		m := c.Manifest
		if m.Kind != "extension" && m.Kind != "preset" {
			return fmt.Errorf("%s: kind %q is not publishable (extensions and presets only)", m.Name, m.Kind)
		}
		archive := fmt.Sprintf("%s-%s.tar.gz", m.Name, m.Version)
		// pack to a temp path first: a refused publish must not
		// overwrite the archive a live index still pins
		tmp := filepath.Join(out, "."+archive+".tmp")
		sha, err := pack(dir, tmp)
		if err != nil {
			_ = os.Remove(tmp)
			return err
		}
		if idx.Components[m.Name] == nil {
			idx.Components[m.Name] = map[string]release{}
		}
		if prev, exists := idx.Components[m.Name][m.Version]; exists && prev.Sha256 != sha {
			_ = os.Remove(tmp)
			return fmt.Errorf("%s@%s is already published with different content — bump the version instead of republishing", m.Name, m.Version)
		}
		if err := os.Rename(tmp, filepath.Join(out, archive)); err != nil {
			return err
		}
		idx.Components[m.Name][m.Version] = release{
			URL: archive, Sha256: sha, Kind: m.Kind,
			Description: strings.TrimSpace(m.Description),
		}
		fmt.Printf("published %s@%s (%s) sha256:%s\n", m.Name, m.Version, m.Kind, sha[:12])
	}
	b, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(out, "index.json"), append(b, '\n'), 0o644)
}

// pack writes a deterministic tar.gz of dir and returns its sha256.
func pack(dir, dest string) (string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("%s: symlinks are not publishable", p)
		}
		if !d.IsDir() {
			rel, _ := filepath.Rel(dir, p)
			files = append(files, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(files)
	f, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	gz, _ := gzip.NewWriterLevel(io2{f, h}, gzip.BestCompression)
	tw := tar.NewWriter(gz)
	for _, rel := range files {
		b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
		if err != nil {
			return "", err
		}
		if err := tw.WriteHeader(&tar.Header{
			Name: rel, Mode: 0o644, Size: int64(len(b)), Typeflag: tar.TypeReg,
			// fixed metadata keeps archives byte-identical across runs
		}); err != nil {
			return "", err
		}
		if _, err := tw.Write(b); err != nil {
			return "", err
		}
	}
	if err := tw.Close(); err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// io2 tees archive bytes into the hash while writing the file.
type io2 struct {
	f interface{ Write([]byte) (int, error) }
	h interface{ Write([]byte) (int, error) }
}

func (w io2) Write(p []byte) (int, error) {
	_, _ = w.h.Write(p)
	return w.f.Write(p)
}
