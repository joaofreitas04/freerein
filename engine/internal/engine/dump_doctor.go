package engine

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"

	"github.com/joaofreitas04/freerein/engine/internal/component"
	"github.com/joaofreitas04/freerein/engine/internal/envelope"
	"github.com/joaofreitas04/freerein/engine/internal/lockfile"
)

// Dump prints the resolved composition — the fixed point
// (spec/resolution.md rule 3). Full detail goes to a file; the
// envelope carries the summary (spec/cli-envelope.md rule 5).
func (g *Engine) Dump(e *envelope.Envelope) {
	r := g.resolveAll(e)
	if r == nil {
		return
	}
	type dumpFile struct {
		Path     string   `json:"path"`
		Layer    string   `json:"layer"`
		Refs     []string `json:"refs"`
		Hash     string   `json:"hash"`
		Shadowed []string `json:"shadowed,omitempty"`
		Content  string   `json:"content"`
	}
	var files []dumpFile
	for path, rf := range r.rendered {
		var shadowed []string
		if entry, ok := r.set.Entries[path]; ok {
			for _, s := range entry.Shadowed {
				shadowed = append(shadowed, s.Component)
			}
		}
		files = append(files, dumpFile{
			Path: path, Layer: rf.Layer, Refs: rf.Refs, Hash: rf.Hash,
			Shadowed: shadowed, Content: string(rf.Content),
		})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	full := map[string]any{
		"adapter": r.adapter.Name, "layers": r.layers, "files": files,
	}
	outPath := filepath.Join(OutDir, "dump.json")
	b, _ := json.MarshalIndent(full, "", "  ")
	if err := os.MkdirAll(filepath.Join(g.Repo, OutDir), 0o755); err == nil {
		err = os.WriteFile(filepath.Join(g.Repo, outPath), append(b, '\n'), 0o644)
		if err != nil {
			e.Fail("WRITE_FAILED", err.Error(), "check permissions on "+OutDir)
			return
		}
	}
	summary := []map[string]any{}
	for _, f := range files {
		summary = append(summary, map[string]any{"path": f.Path, "layer": f.Layer, "refs": f.Refs})
	}
	e.Result = map[string]any{"detail": outPath, "adapter": r.adapter.Name, "files": summary}
}

// Doctor reads the lockfile and reports: drift, missing files,
// composition mismatch, and stale compensations.
func (g *Engine) Doctor(e *envelope.Envelope) {
	lock, err := lockfile.Read(g.Repo)
	if err != nil {
		e.Fail("LOCKFILE_INVALID", err.Error(), "inspect or delete "+lockfile.Name+" and re-run `rein apply`")
		return
	}
	if lock == nil {
		e.Fail("NOT_APPLIED", "no "+lockfile.Name+" — the harness has never been applied", "run `rein plan`, then `rein apply`")
		return
	}
	checks := 0
	// 1. tree vs lockfile: drift and missing files
	var paths []string
	for p := range lock.Files {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		checks++
		th, err := g.treeHash(p)
		switch {
		case errors.Is(err, os.ErrNotExist):
			e.Diag(envelope.Error, "FILE_MISSING", p+" is in the lockfile but gone from the tree",
				"run `rein apply --yes` to restore it, or `rein plan` to review first")
		case err != nil:
			e.Diag(envelope.Error, "FILE_UNREADABLE", p+": "+err.Error(), "check permissions")
		case th != lock.Files[p].Hash:
			e.Diag(envelope.Warning, "DRIFT", p+" differs from the installed content",
				"intentional edits belong in "+OverridesDir+" so they survive upgrades; run `rein plan` to review")
		}
	}
	// 2. resolved composition vs lockfile
	r := g.resolveAll(e)
	if r == nil {
		return
	}
	p, err := g.computePlan(r)
	if err == nil && !p.empty() {
		e.Diag(envelope.Warning, "COMPOSITION_BEHIND",
			"the resolved composition differs from what is installed",
			"run `rein plan` to review, then `rein apply --yes`")
	}
	// 3. stale compensations (rent doctrine, spec/component-manifest.md)
	core, err := component.LoadAll(g.Content, "core")
	if err == nil {
		for _, c := range core {
			if c.Manifest.Rent.Class == "compensation" {
				e.Diag(envelope.Info, "COMPENSATION_RECHECK",
					c.Manifest.Name+" is a compensation — re-check trigger: "+c.Manifest.Rent.Expires,
					"if the current model no longer needs it, remove the component and measure")
			}
		}
	}
	e.Result = map[string]any{"files_checked": checks, "findings": len(e.Diagnostics)}
}

// Adapters lists the embedded adapters.
func (g *Engine) Adapters(e *envelope.Envelope) {
	names, err := readDirFS(g.Content, "adapters")
	if err != nil {
		e.Fail("CONTENT_INVALID", err.Error(), "this is an engine bug — report it")
		return
	}
	sort.Strings(names)
	e.Result = map[string]any{"adapters": names}
}
