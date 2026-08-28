// Package registry implements the client side of spec/registry.md: a
// static JSON index mapping name@version to a sha256-pinned archive.
// Fetch happens only in author-side commands (add, info, upgrade);
// consume-side commands read the committed vendor tree offline.
package registry

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Release struct {
	URL         string `json:"url"`
	Sha256      string `json:"sha256"` // of the archive
	Kind        string `json:"kind"`
	Description string `json:"description,omitempty"`
}

type Index struct {
	Components map[string]map[string]Release `json:"components"`
	base       string                        // for resolving relative URLs
}

// LoadIndex reads a registry index from an https:// URL or a
// filesystem path.
func LoadIndex(src string) (*Index, error) {
	var raw []byte
	var err error
	if strings.HasPrefix(src, "https://") || strings.HasPrefix(src, "http://") {
		raw, err = httpGet(src)
	} else {
		raw, err = os.ReadFile(src)
	}
	if err != nil {
		return nil, fmt.Errorf("registry %s: %w", src, err)
	}
	var idx Index
	if err := json.Unmarshal(raw, &idx); err != nil {
		return nil, fmt.Errorf("registry %s: %w", src, err)
	}
	idx.base = baseOf(src)
	return &idx, nil
}

func baseOf(src string) string {
	if strings.Contains(src, "://") {
		i := strings.LastIndex(src, "/")
		return src[:i+1]
	}
	return filepath.Dir(src) + string(filepath.Separator)
}

func (i *Index) Lookup(name, version string) (*Release, error) {
	versions, ok := i.Components[name]
	if !ok {
		return nil, fmt.Errorf("component %q not in the registry", name)
	}
	r, ok := versions[version]
	if !ok {
		return nil, fmt.Errorf("component %s has no version %s (available: %s)",
			name, version, strings.Join(sortedVersions(versions), ", "))
	}
	return &r, nil
}

// Latest returns the highest version by semver-ish ordering.
func (i *Index) Latest(name string) (string, *Release, error) {
	versions, ok := i.Components[name]
	if !ok {
		return "", nil, fmt.Errorf("component %q not in the registry", name)
	}
	vs := sortedVersions(versions)
	v := vs[len(vs)-1]
	r := versions[v]
	return v, &r, nil
}

func sortedVersions(m map[string]Release) []string {
	var vs []string
	for v := range m {
		vs = append(vs, v)
	}
	sort.Slice(vs, func(a, b int) bool { return CompareVersions(vs[a], vs[b]) < 0 })
	return vs
}

// CompareVersions orders dotted numeric versions; non-numeric
// segments fall back to string order.
func CompareVersions(a, b string) int {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(as) || i < len(bs); i++ {
		var x, y string
		if i < len(as) {
			x = as[i]
		}
		if i < len(bs) {
			y = bs[i]
		}
		xi, xe := strconv.Atoi(x)
		yi, ye := strconv.Atoi(y)
		switch {
		case xe == nil && ye == nil && xi != yi:
			if xi < yi {
				return -1
			}
			return 1
		case x != y:
			if x < y {
				return -1
			}
			return 1
		}
	}
	return 0
}

// Fetch downloads (or reads) the release archive, verifies its
// sha256 against the index pin, and unpacks it into destDir.
func (i *Index) Fetch(r *Release, destDir string) error {
	src := r.URL
	if !strings.Contains(src, "://") && !filepath.IsAbs(src) {
		src = i.base + src
	}
	var raw []byte
	var err error
	if strings.HasPrefix(src, "https://") || strings.HasPrefix(src, "http://") {
		raw, err = httpGet(src)
	} else {
		raw, err = os.ReadFile(src)
	}
	if err != nil {
		return err
	}
	sum := sha256.Sum256(raw)
	if got := hex.EncodeToString(sum[:]); got != r.Sha256 {
		return fmt.Errorf("sha256 mismatch: index pins %s, archive is %s — refusing to unpack", r.Sha256, got)
	}
	return untar(raw, destDir)
}

func httpGet(url string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 32<<20))
}

func untar(raw []byte, destDir string) error {
	gz, err := gzip.NewReader(strings.NewReader(string(raw)))
	if err != nil {
		return err
	}
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		clean := path.Clean(hdr.Name)
		if strings.HasPrefix(clean, "..") || path.IsAbs(clean) {
			return fmt.Errorf("archive escapes its root: %s", hdr.Name)
		}
		full := filepath.Join(destDir, filepath.FromSlash(clean))
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(full, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
				return err
			}
			b, err := io.ReadAll(io.LimitReader(tr, 8<<20))
			if err != nil {
				return err
			}
			if err := os.WriteFile(full, b, 0o644); err != nil {
				return err
			}
		default:
			return fmt.Errorf("archive contains unsupported entry type for %s (symlinks are refused)", hdr.Name)
		}
	}
}

// TreeHash is a deterministic content hash of a directory: sha256
// over sorted (path, content) pairs. Doctor compares it against the
// lockfile pin to detect vendor tampering.
func TreeHash(root string) (string, error) {
	h := sha256.New()
	var files []string
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(files)
	for _, f := range files {
		b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(f)))
		if err != nil {
			return "", err
		}
		fmt.Fprintf(h, "%s\x00%d\x00", f, len(b))
		h.Write(b)
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}
