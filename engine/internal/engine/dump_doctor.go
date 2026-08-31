package engine

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/joaofreitas04/freerein/engine/internal/envelope"
	"github.com/joaofreitas04/freerein/engine/internal/lockfile"
	"github.com/joaofreitas04/freerein/engine/internal/registry"
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
	e.Diag(envelope.Info, "OUTPUT_OFFLOADED", "full dump written to "+outPath,
		"read "+outPath+" for the per-file detail omitted from this envelope")
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
		if lock.Files[p].Seed {
			continue // agent-owned; content is theirs to change
		}
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
	// 2. vendored components vs their sha pins (tamper detection)
	for _, layer := range lock.Layers {
		if layer.Sha == "" {
			continue
		}
		checks++
		ref := strings.TrimPrefix(layer.Source, "registry:")
		dir := filepath.Join(g.Repo, VendorDir, ref)
		if _, err := os.Stat(dir); err != nil {
			e.Diag(envelope.Error, "VENDOR_MISSING", ref+" is pinned in the lockfile but gone from "+VendorDir,
				"run `rein add "+ref+"` to re-fetch it")
			continue
		}
		sha, err := registry.TreeHash(dir)
		if err != nil {
			e.Diag(envelope.Error, "VENDOR_UNREADABLE", ref+": "+err.Error(), "check permissions")
			continue
		}
		if sha != layer.Sha {
			e.Diag(envelope.Error, "VENDOR_TAMPERED",
				ref+" no longer matches its pinned hash — its content changed after install",
				"if the edit was yours, move it to a local-path source; otherwise re-fetch with `rein add "+ref+"` and review the diff")
		}
	}
	// 3. resolved composition vs lockfile
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
	// 4. stale compensations (rent doctrine, spec/component-manifest.md
	// rule 3) — over every resolved layer, not just the core: a
	// vendored extension's stale assumption is as expensive as an
	// embedded one.
	for _, c := range r.components {
		if c.Manifest.Rent.Class == "compensation" {
			e.Diag(envelope.Info, "COMPENSATION_RECHECK",
				c.Manifest.Ref()+" is a compensation — re-check trigger: "+c.Manifest.Rent.Expires,
				"if the current model no longer needs it, remove the component and measure")
		}
	}
	// 5. composition shape (spec/component-manifest.md rule 7)
	compositionAdvisories(e, r.components)
	// 6. the gate is real — a stub that always fails is not a gate,
	// and a repo carrying one has verification in name only.
	if rf, ok := r.rendered["scripts/verify"]; ok && rf.Layer == "core" {
		e.Diag(envelope.Warning, "GATE_STUB",
			"scripts/verify is still the shipped stub — it fails until configured, so this repo has no working gate",
			"author the project's checks in .rein/overrides/scripts/verify (rein-setup step 5 has the skeleton), then run `rein apply --yes`")
	}
	// 7. the debt ledger's own discipline (state-base seed schema)
	g.debtChecks(e, r.adapter.StateDir)
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

// debtChecks audits the debt ledger's rows: a debt without evidence
// and a trigger is unactionable, and a dated trigger that has passed
// re-activates the entry — recorded debt with no expiry is how an
// exception quietly becomes policy.
var debtDate = regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}\b`)

func (g *Engine) debtChecks(e *envelope.Envelope, stateDir string) {
	if stateDir == "" {
		stateDir = ".rein/state"
	}
	rel := filepath.ToSlash(filepath.Join(stateDir, "DEBT.md"))
	b, err := os.ReadFile(filepath.Join(g.Repo, stateDir, "DEBT.md"))
	if err != nil {
		return // seed absent: plan handles restoration
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	for _, line := range strings.Split(string(b), "\n") {
		cells := tableRow(line)
		if cells == nil || cells[0] == "Debt" {
			continue
		}
		label := cells[0]
		if len(label) > 40 {
			label = label[:40] + "…"
		}
		if cells[1] == "" || cells[2] == "" {
			e.Diag(envelope.Warning, "DEBT_ROW_INCOMPLETE",
				rel+": "+label+" — a debt without evidence and a paydown trigger is unactionable",
				"add the evidence (counts, paths) and a trigger — an event or date the project cannot miss")
			continue
		}
		if m := debtDate.FindString(cells[2]); m != "" {
			if d, perr := time.Parse("2006-01-02", m); perr == nil && d.Before(today) {
				e.Diag(envelope.Warning, "DEBT_EXPIRED",
					rel+": "+label+" — its trigger date ("+m+") has passed, so the entry is active again",
					"pay the debt down now, or re-rule it with a new trigger; an expired entry left in place is an exception becoming policy")
			}
		}
	}
}

// tableRow parses one markdown table row into trimmed cells, returning
// nil for non-rows and separator rows.
func tableRow(line string) []string {
	t := strings.TrimSpace(line)
	if !strings.HasPrefix(t, "|") {
		return nil
	}
	parts := strings.Split(t, "|")
	if len(parts) < 4 { // leading empty + 3 cells minimum
		return nil
	}
	cells := make([]string, 0, len(parts)-2)
	for _, p := range parts[1 : len(parts)-1] {
		cells = append(cells, strings.TrimSpace(p))
	}
	sep := true
	for _, c := range cells {
		if strings.Trim(c, "-: ") != "" {
			sep = false
			break
		}
	}
	if sep || len(cells) < 3 {
		return nil
	}
	return cells[:3]
}
