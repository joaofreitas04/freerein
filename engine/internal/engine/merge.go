package engine

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
)

// BaseDir mirrors the installed (upstream) content of every managed
// file — the merge base for reconciling local edits with upgrades.
// Committed to the target repo so merges work across machines.
const BaseDir = ".rein/base"

func (g *Engine) basePath(path string) string {
	return filepath.Join(g.Repo, BaseDir, path)
}

func (g *Engine) writeBase(path string, content []byte) error {
	full := g.basePath(path)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	return os.WriteFile(full, content, 0o644)
}

func (g *Engine) readBase(path string) ([]byte, bool) {
	b, err := os.ReadFile(g.basePath(path))
	return b, err == nil
}

// mergeFile three-way merges via `git merge-file -p`. Returns the
// merged content and whether it carries conflict markers.
func mergeFile(ours, base, theirs []byte) (merged []byte, conflict bool, err error) {
	dir, err := os.MkdirTemp("", "rein-merge-")
	if err != nil {
		return nil, false, err
	}
	defer os.RemoveAll(dir)
	files := map[string][]byte{"ours": ours, "base": base, "theirs": theirs}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), content, 0o644); err != nil {
			return nil, false, err
		}
	}
	cmd := exec.Command("git", "merge-file", "-p",
		"-L", "local", "-L", "installed", "-L", "upstream",
		filepath.Join(dir, "ours"), filepath.Join(dir, "base"), filepath.Join(dir, "theirs"))
	out, err := cmd.Output()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) && exit.ExitCode() > 0 {
			return out, true, nil // >0 = number of conflicts
		}
		return nil, false, err
	}
	return out, false, nil
}
