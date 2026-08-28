package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joaofreitas04/freerein/engine/internal/component"
)

// loadSource loads one preset/extension from a harness.yaml entry.
// Local paths (./…, ../…, /…) load a component directory; registry
// refs (name@version) are not implemented yet.
func (g *Engine) loadSource(src string) (*component.Loaded, error) {
	if !strings.HasPrefix(src, "./") && !strings.HasPrefix(src, "../") && !strings.HasPrefix(src, "/") {
		return nil, fmt.Errorf("registry sources are not implemented")
	}
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
	return component.LoadDir(os.DirFS(filepath.Clean(parent)), base)
}
