package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/joaofreitas04/freerein/engine/internal/component"
	"github.com/joaofreitas04/freerein/engine/internal/registry"
)

// VendorDir holds fetched registry components, committed to the
// target repo so consume-side commands stay offline and installs are
// reviewable in the diff.
const VendorDir = ".rein/vendor"

var registryRef = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*@[0-9][0-9A-Za-z.+-]*$`)

type sourceInfo struct {
	comp *component.Loaded
	ref  string // lockfile layer source, e.g. "path:./x" or "registry:name@1.0.0"
	sha  string // tree hash for vendored components, "" for local paths
}

// loadSource loads one preset/extension from a harness.yaml entry:
// a local path (./…, ../…, /…) or a registry ref (name@version)
// resolved through the committed vendor tree.
func (g *Engine) loadSource(src string) (*sourceInfo, error) {
	switch {
	case strings.HasPrefix(src, "./") || strings.HasPrefix(src, "../") || strings.HasPrefix(src, "/"):
		abs := src
		if !filepath.IsAbs(abs) {
			abs = filepath.Join(g.Repo, src)
		}
		info, err := os.Stat(abs)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("%s is not a directory", src)
		}
		parent, base := filepath.Split(filepath.Clean(abs))
		c, err := component.LoadDir(os.DirFS(filepath.Clean(parent)), base)
		if err != nil {
			return nil, err
		}
		return &sourceInfo{comp: c, ref: "path:" + src}, nil
	case registryRef.MatchString(src):
		dir := filepath.Join(g.Repo, VendorDir, src)
		if _, err := os.Stat(dir); err != nil {
			return nil, fmt.Errorf("%s is not vendored under %s — run `rein add %s` to fetch it", src, VendorDir, src)
		}
		parent, base := filepath.Split(filepath.Clean(dir))
		c, err := component.LoadDir(os.DirFS(filepath.Clean(parent)), base)
		if err != nil {
			return nil, err
		}
		sha, err := registry.TreeHash(dir)
		if err != nil {
			return nil, err
		}
		return &sourceInfo{comp: c, ref: "registry:" + src, sha: sha}, nil
	default:
		return nil, fmt.Errorf("unrecognized source (use ./path or name@version)")
	}
}
