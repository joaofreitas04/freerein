// Package engine orchestrates the rein commands over the embedded
// content, the target repo, and the lockfile.
package engine

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/joaofreitas04/freerein/engine/internal/adapter"
	"github.com/joaofreitas04/freerein/engine/internal/component"
	"github.com/joaofreitas04/freerein/engine/internal/envelope"
	"github.com/joaofreitas04/freerein/engine/internal/lockfile"
	"github.com/joaofreitas04/freerein/engine/internal/resolve"
)

// Version is the engine version. A var, not a const: release builds
// stamp the tag over it with -ldflags "-X"; source builds stay -dev
// so a binary always says which kind it is.
var Version = "0.1.0-dev"

const (
	ConfigName   = "harness.yaml"
	OverridesDir = ".rein/overrides"
	OutDir       = ".rein/out"
)

type Config struct {
	Adapter    string   `yaml:"adapter"`
	Registry   string   `yaml:"registry,omitempty"`
	Presets    []string `yaml:"presets"`
	Extensions []string `yaml:"extensions"`
}

type Engine struct {
	Repo    string // target repo root
	Content fs.FS  // embedded content: core/, adapters/
}

func (g *Engine) engineID() string { return "rein " + Version }

// ---------- config ----------

func (g *Engine) readConfig() (*Config, error) {
	b, err := os.ReadFile(filepath.Join(g.Repo, ConfigName))
	if err != nil {
		return nil, err
	}
	var c Config
	dec := yaml.NewDecoder(bytes.NewReader(b))
	dec.KnownFields(true)
	if err := dec.Decode(&c); err != nil {
		return nil, fmt.Errorf("%s: %w", ConfigName, err)
	}
	if c.Adapter == "" {
		return nil, fmt.Errorf("%s: adapter is required", ConfigName)
	}
	return &c, nil
}

// ---------- init ----------

func (g *Engine) Init(e *envelope.Envelope, adapterName string) {
	cfgPath := filepath.Join(g.Repo, ConfigName)
	if _, err := os.Stat(cfgPath); err == nil {
		e.Fail("ALREADY_INITIALIZED", ConfigName+" already exists in this repo",
			"run `rein plan` to see pending changes, or edit "+ConfigName+" directly")
		return
	}
	if _, err := adapter.Load(g.Content, adapterName); err != nil {
		e.Fail("UNKNOWN_ADAPTER", err.Error(), "run `rein adapters` to list available adapters")
		return
	}
	cfg := fmt.Sprintf("# FreeRein harness declaration — see spec/resolution.md\nadapter: %s\npresets: []\nextensions: []\n", adapterName)
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o644); err != nil {
		e.Fail("WRITE_FAILED", err.Error(), "check permissions on the target repo")
		return
	}
	if err := os.MkdirAll(filepath.Join(g.Repo, ".rein"), 0o755); err == nil {
		_ = os.WriteFile(filepath.Join(g.Repo, ".rein", ".gitignore"), []byte("out/\n"), 0o644)
	}
	e.Result = map[string]any{"created": []string{ConfigName, ".rein/.gitignore"}, "adapter": adapterName}
	e.Diag(envelope.Info, "INITIALIZED", "harness declared; nothing installed yet",
		"run `rein plan` to see what `rein apply` would install")
}

// ---------- resolution pipeline ----------

type resolution struct {
	cfg        *Config
	adapter    *adapter.Adapter
	set        *resolve.Set
	rendered   map[string]*resolve.Rendered
	layers     []lockfile.LayerRef
	components []*component.Loaded // every loaded component, layer order
}

func (g *Engine) resolveAll(e *envelope.Envelope) *resolution {
	cfg, err := g.readConfig()
	if errors.Is(err, os.ErrNotExist) {
		e.Fail("NOT_INITIALIZED", ConfigName+" not found in "+g.Repo, "run `rein init` first")
		return nil
	}
	if err != nil {
		e.Fail("CONFIG_INVALID", err.Error(), "fix "+ConfigName+" and re-run")
		return nil
	}
	ad, err := adapter.Load(g.Content, cfg.Adapter)
	if err != nil {
		e.Fail("UNKNOWN_ADAPTER", err.Error(), "fix the adapter field in "+ConfigName)
		return nil
	}
	core, err := component.LoadAll(g.Content, "core")
	if err != nil {
		e.Fail("CORE_INVALID", err.Error(), "this is an engine bug — report it")
		return nil
	}
	// layer order per spec/resolution.md: overrides > presets > extensions > core
	layers := []resolve.Layer{}
	sources := map[string]lockfile.LayerRef{
		"overrides": {ID: "overrides", Source: "local"},
		"core":      {ID: "core", Source: "embedded@" + Version},
	}
	if ov := g.loadOverrides(e); ov != nil {
		layers = append(layers, resolve.Layer{ID: "overrides", Components: []*component.Loaded{ov}})
	}
	for _, kind := range []struct {
		kind    string
		entries []string
	}{{"preset", cfg.Presets}, {"extension", cfg.Extensions}} {
		for _, src := range kind.entries {
			si, err := g.loadSource(src)
			if err != nil {
				e.Fail("SOURCE_INVALID", fmt.Sprintf("%s %q: %v", kind.kind, src, err),
					"use a local path (./…) to a component directory, or `rein add name@version` for registry components")
				return nil
			}
			c := si.comp
			if c.Manifest.Kind != kind.kind {
				e.Fail("KIND_MISMATCH",
					fmt.Sprintf("%q is declared under %ss but its manifest says kind: %s", src, kind.kind, c.Manifest.Kind),
					"move it to the matching list in "+ConfigName+" or fix its component.yaml")
				return nil
			}
			id := kind.kind + ":" + c.Manifest.Name
			layers = append(layers, resolve.Layer{ID: id, Components: []*component.Loaded{c}})
			sources[id] = lockfile.LayerRef{ID: id, Source: si.ref, Sha: si.sha}
		}
	}
	layers = append(layers, resolve.Layer{ID: "core", Components: core})
	// probes (spec/component-manifest.md rule 4) and declared conflicts,
	// across every loaded component
	all := map[string]*component.Loaded{}
	var comps []*component.Loaded
	for _, l := range layers {
		for _, c := range l.Components {
			all[c.Manifest.Name] = c
			comps = append(comps, c)
			for _, req := range c.Manifest.Requires {
				if ok, detail := g.probe(req); !ok {
					e.Fail("REQUIREMENT_UNMET",
						fmt.Sprintf("component %s requires %q: %s", c.Manifest.Name, req, detail),
						"satisfy the requirement in the target repo, or remove the component")
					return nil
				}
			}
		}
	}
	for name, c := range all {
		for _, rival := range c.Manifest.Conflicts {
			if _, loaded := all[rival]; loaded {
				e.Fail("CONFLICT", fmt.Sprintf("%s declares a conflict with %s and both are loaded", name, rival),
					"remove one of the two from "+ConfigName)
				return nil
			}
		}
	}
	set, err := resolve.Resolve(layers)
	if err != nil {
		e.Fail("RESOLVE_FAILED", err.Error(), "fix the offending component and re-run")
		return nil
	}
	rendered, err := resolve.Render(set, ad)
	if err != nil {
		e.Fail("RENDER_FAILED", err.Error(), "reduce instruction fragments or raise the adapter limit")
		return nil
	}
	// declared degradations: spec/host-adapter.md rule 1
	for _, d := range ad.Degradations {
		e.Diag(envelope.Info, "HOST_DEGRADATION", ad.Name+": "+d, "")
	}
	if !ad.Hooks.PostCompactionReinjection {
		e.Diag(envelope.Warning, "HOST_NO_COMPACTION_HOOK",
			ad.Name+": no post-compaction re-injection point; the instruction file is the only bootstrap",
			"keep load-bearing rules in the instruction file, not in session state")
	}
	var lrefs []lockfile.LayerRef
	for _, l := range layers {
		lrefs = append(lrefs, sources[l.ID])
	}
	return &resolution{cfg: cfg, adapter: ad, set: set, rendered: rendered, layers: lrefs, components: comps}
}

// compositionAdvisories prices the composition's shape (spec rule 7):
// two loaded components sharing an `addresses` entry is a warning —
// stacked fixes for one failure compound cost faster than effect.
func compositionAdvisories(e *envelope.Envelope, comps []*component.Loaded) {
	byAddr := map[string][]string{}
	for _, c := range comps {
		for _, a := range c.Manifest.Addresses {
			byAddr[a] = append(byAddr[a], c.Manifest.Name)
		}
	}
	var overlapping []string
	for a, names := range byAddr {
		if len(names) > 1 {
			overlapping = append(overlapping, a)
		}
	}
	sort.Strings(overlapping)
	for _, a := range overlapping {
		e.Diag(envelope.Warning, "ADDRESSES_OVERLAP",
			fmt.Sprintf("%q is addressed by %s — stacked fixes for one failure usually add cost faster than effect", a, strings.Join(byAddr[a], " and ")),
			"strengthen one of them instead of keeping both; `rein remove` the weaker, or narrow its addresses")
	}
}

// loadOverrides builds the pseudo-component for .rein/overrides/**.
func (g *Engine) loadOverrides(e *envelope.Envelope) *component.Loaded {
	root := filepath.Join(g.Repo, OverridesDir)
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return nil
	}
	lc := &component.Loaded{
		Manifest: component.Manifest{
			Name: "local-overrides", Kind: "core", Version: "local",
			Subsystem: "instructions", Rung: "instruction",
			Rent: component.Rent{Class: "amplifier"},
		},
		Files: map[string][]byte{},
	}
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		rel = filepath.ToSlash(rel)
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		lc.Files[rel] = b
		lc.Manifest.Provides = append(lc.Manifest.Provides, rel)
		return nil
	})
	if len(lc.Files) == 0 {
		return nil
	}
	sort.Strings(lc.Manifest.Provides)
	return lc
}

// probe answers one affordance question (spec/component-manifest.md
// rule 4). Detection is shared with `rein inspect` (spec/inspection.md
// rule 2), so requires-gating and the report cannot disagree.
func (g *Engine) probe(name string) (bool, string) {
	switch name {
	case "git":
		if _, err := os.Stat(filepath.Join(g.Repo, ".git")); err == nil {
			return true, ""
		}
		return false, "no .git directory — the target is not a git repository"
	case "test-runner":
		configs, candidates := g.detectTests()
		if len(configs) > 0 || len(candidates) > 0 {
			return true, ""
		}
		return false, "no test configuration or manifest-derived test command found"
	case "ci":
		if len(g.detectCI()) > 0 {
			return true, ""
		}
		return false, "no CI configuration found"
	case "linter":
		if len(g.detectFiles(lintFormatConfigs)) > 0 {
			return true, ""
		}
		return false, "no linter or formatter configuration found"
	case "docs-tree":
		if g.detectDocsTree() != "" {
			return true, ""
		}
		return false, "no docs/ tree with markdown pages"
	}
	return false, "unknown probe"
}

// ---------- plan ----------

type PlanItem struct {
	Path      string `json:"path"`
	Component string `json:"component,omitempty"`
	Detail    string `json:"detail,omitempty"`
}

type Plan struct {
	Adds      []PlanItem `json:"adds"`
	Changes   []PlanItem `json:"changes"`
	Drift     []PlanItem `json:"drift"`
	Removes   []PlanItem `json:"removes"`
	Unmanaged []PlanItem `json:"unmanaged"`
	Budget    *Budget    `json:"budget,omitempty"`
	Costs     *Costs     `json:"costs,omitempty"`
}

// Budget prices the rendered instruction file against the adapter's
// limit (spec/resolution.md rule 6): the agent pays this on every
// turn, so growth must be a decision, never an accident.
type Budget struct {
	InstructionFile string `json:"instruction_file"`
	Bytes           int    `json:"bytes"`
	MaxBytes        int    `json:"max_bytes"`
	Fragments       int    `json:"fragments"`
}

func (p *Plan) empty() bool {
	return len(p.Adds)+len(p.Changes)+len(p.Drift)+len(p.Removes)+len(p.Unmanaged) == 0
}

// plan = resolved set x lockfile x working tree (spec/resolution.md rule 5).
func (g *Engine) computePlan(r *resolution) (*Plan, error) {
	lock, err := lockfile.Read(g.Repo)
	if err != nil {
		return nil, err
	}
	p := &Plan{Adds: []PlanItem{}, Changes: []PlanItem{}, Drift: []PlanItem{}, Removes: []PlanItem{}, Unmanaged: []PlanItem{}}
	var paths []string
	for path := range r.rendered {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		rf := r.rendered[path]
		var entry *lockfile.FileEntry
		if lock != nil {
			entry = lock.Files[path]
		}
		if rf.Seed {
			// install-if-absent: the agent owns it after install
			if _, err := g.treeHash(path); err != nil {
				p.Adds = append(p.Adds, PlanItem{Path: path, Component: strings.Join(rf.Refs, ", "), Detail: "seed — installed once, agent-owned after"})
			}
			continue
		}
		if entry == nil {
			if th, err := g.treeHash(path); err == nil && th != rf.Hash {
				p.Unmanaged = append(p.Unmanaged, PlanItem{Path: path, Component: strings.Join(rf.Refs, ", "),
					Detail: "already exists with different content and is not managed by rein — apply will NOT overwrite it"})
				continue
			}
			p.Adds = append(p.Adds, PlanItem{Path: path, Component: strings.Join(rf.Refs, ", ")})
			continue
		}
		treeHash, treeErr := g.treeHash(path)
		drifted := treeErr == nil && treeHash != entry.Hash
		missing := errors.Is(treeErr, os.ErrNotExist)
		switch {
		case drifted && rf.Hash != entry.Hash:
			p.Drift = append(p.Drift, PlanItem{Path: path, Component: entry.Component,
				Detail: "locally edited AND upstream changed — apply will attempt a three-way merge"})
		case drifted:
			p.Drift = append(p.Drift, PlanItem{Path: path, Component: entry.Component,
				Detail: "locally edited since install — respected, apply will not touch it"})
		case missing:
			p.Adds = append(p.Adds, PlanItem{Path: path, Component: strings.Join(rf.Refs, ", "), Detail: "in lockfile but missing from tree — will be restored"})
		case rf.Hash != entry.Hash:
			p.Changes = append(p.Changes, PlanItem{Path: path, Component: strings.Join(rf.Refs, ", ")})
		}
	}
	if lock != nil {
		var lockPaths []string
		for path := range lock.Files {
			lockPaths = append(lockPaths, path)
		}
		sort.Strings(lockPaths)
		for _, path := range lockPaths {
			if _, ok := r.rendered[path]; !ok {
				detail := ""
				if lock.Files[path].Seed {
					detail = "seed — agent-owned; apply drops the lock entry and leaves the file in place"
				}
				p.Removes = append(p.Removes, PlanItem{Path: path, Component: lock.Files[path].Component, Detail: detail})
			}
		}
	}
	return p, nil
}

func (g *Engine) treeHash(path string) (string, error) {
	b, err := os.ReadFile(filepath.Join(g.Repo, path))
	if err != nil {
		return "", err
	}
	return resolve.Hash(b), nil
}

func (g *Engine) Plan(e *envelope.Envelope) {
	r := g.resolveAll(e)
	if r == nil {
		return
	}
	p, err := g.computePlan(r)
	if err != nil {
		e.Fail("LOCKFILE_INVALID", err.Error(), "inspect or delete "+lockfile.Name+" and re-run")
		return
	}
	// spec/resolution.md rule 6: plan prices the composition.
	if rf, ok := r.rendered[r.adapter.InstructionFile.Path]; ok {
		p.Budget = &Budget{
			InstructionFile: r.adapter.InstructionFile.Path,
			Bytes:           len(rf.Content),
			MaxBytes:        r.adapter.InstructionFile.MaxBytes,
			Fragments:       len(rf.Refs),
		}
		if p.Budget.MaxBytes > 0 && p.Budget.Bytes*5 >= p.Budget.MaxBytes*4 {
			e.Diag(envelope.Warning, "NEAR_CONTEXT_BUDGET",
				fmt.Sprintf("%s renders to %d bytes — over 80%% of the adapter's %d-byte limit",
					p.Budget.InstructionFile, p.Budget.Bytes, p.Budget.MaxBytes),
				"instruction content is paid on every turn: drop fragments an agent could derive from the tree, and move rules that only apply sometimes behind a condition (a skill, a hook) instead of stating them always")
		}
	}
	p.Costs = g.computeCosts(r)
	compositionAdvisories(e, r.components)
	e.Result = p
	if p.empty() {
		e.Diag(envelope.Info, "UP_TO_DATE", "installed harness matches the resolved composition", "")
	}
}

// Probes lists the affordance vocabulary `requires` entries may use
// (spec/component-manifest.md rule 4).
func (g *Engine) Probes(e *envelope.Envelope) {
	e.Result = map[string]any{"probes": component.ProbeVocabulary}
}

// ---------- apply ----------

func (g *Engine) Apply(e *envelope.Envelope, yes bool) {
	r := g.resolveAll(e)
	if r == nil {
		return
	}
	p, err := g.computePlan(r)
	if err != nil {
		e.Fail("LOCKFILE_INVALID", err.Error(), "inspect or delete "+lockfile.Name+" and re-run")
		return
	}
	if p.empty() {
		e.Diag(envelope.Info, "UP_TO_DATE", "nothing to apply", "")
		e.Result = p
		return
	}
	if !yes {
		// two-phase confirm: spec/cli-envelope.md rule 4
		e.ConfirmRequired = p
		e.Diag(envelope.Info, "CONFIRM_REQUIRED",
			"apply is side-effectful and was invoked without --yes",
			"show this change-set to the human; re-run `rein apply --yes` once approved")
		return
	}
	prevLock, _ := lockfile.Read(g.Repo)
	applied := []string{}
	conflicts := []string{}
	for _, item := range p.Unmanaged {
		e.Diag(envelope.Warning, "EXISTS_UNMANAGED",
			item.Path+" already exists and is not managed by rein; apply left it alone",
			"to keep your version: move it to "+OverridesDir+"/"+item.Path+" and re-run apply (it becomes the winning layer); to take rein's version: delete the file and re-run apply")
		continue
	}
	merged := map[string]bool{} // drifted paths reconciled this run
	for path, rf := range r.rendered {
		if inList(p.Unmanaged, path) {
			continue // reported above; never written, never based
		}
		item := findItem(p, path)
		if item == nil {
			// unchanged — backfill the base store for locks that
			// predate it, so future upgrades can merge
			if !rf.Seed {
				if _, ok := g.readBase(path); !ok {
					_ = g.writeBase(path, rf.Content)
				}
			}
			continue
		}
		if inDrift(p, path) {
			var prevHash string
			if prevLock != nil {
				if prev, ok := prevLock.Files[path]; ok {
					prevHash = prev.Hash
				}
			}
			if rf.Hash == prevHash {
				// local edit only — respected, untouched
				e.Diag(envelope.Warning, "DRIFT_SKIPPED",
					path+" was locally edited since install; apply left it alone",
					"intentional edits belong in "+OverridesDir+" so they survive upgrades")
				continue
			}
			// local edit AND upstream change: three-way merge
			base, ok := g.readBase(path)
			if !ok {
				e.Diag(envelope.Warning, "MERGE_NO_BASE",
					path+" was locally edited and upstream changed, but no merge base is recorded",
					"reconcile manually against .rein/out/dump.json, or move your edit to "+OverridesDir+" and re-run apply")
				continue
			}
			ours, err := os.ReadFile(filepath.Join(g.Repo, path))
			if err != nil {
				e.Fail("FILE_UNREADABLE", path+": "+err.Error(), "check permissions")
				return
			}
			out, conflict, err := mergeFile(ours, base, rf.Content)
			if err != nil {
				e.Fail("MERGE_FAILED", path+": "+err.Error(), "ensure `git` is on PATH; reconcile manually otherwise")
				return
			}
			if conflict {
				artifact := filepath.Join(OutDir, "merge", path)
				full := filepath.Join(g.Repo, artifact)
				_ = os.MkdirAll(filepath.Dir(full), 0o755)
				_ = os.WriteFile(full, out, 0o644)
				conflicts = append(conflicts, path)
				e.Diag(envelope.Warning, "MERGE_CONFLICT",
					path+": local edit and upstream change overlap; the file was left alone",
					"resolve the conflict markers in "+artifact+", copy the result over "+path+", then re-run `rein apply --yes`")
				continue
			}
			if err := os.WriteFile(filepath.Join(g.Repo, path), out, os.FileMode(rf.Mode)); err != nil {
				e.Fail("WRITE_FAILED", err.Error(), "check permissions")
				return
			}
			_ = g.writeBase(path, rf.Content)
			merged[path] = true
			applied = append(applied, path+" (merged)")
			e.Diag(envelope.Info, "MERGED", path+": upstream change merged with your local edit", "")
			continue
		}
		full := filepath.Join(g.Repo, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			e.Fail("WRITE_FAILED", err.Error(), "check permissions")
			return
		}
		if err := os.WriteFile(full, rf.Content, os.FileMode(rf.Mode)); err != nil {
			e.Fail("WRITE_FAILED", err.Error(), "check permissions")
			return
		}
		if !rf.Seed {
			if err := g.writeBase(path, rf.Content); err != nil {
				e.Fail("WRITE_FAILED", err.Error(), "check permissions on "+BaseDir)
				return
			}
		}
		applied = append(applied, path)
	}
	for _, item := range p.Removes {
		// Seed disposal (spec/component-manifest.md): a seeded file is
		// the agent's after install — un-installing its component must
		// never delete state an agent wrote. Drop the lock entry,
		// leave the file, and say so.
		if prevLock != nil {
			if prev, ok := prevLock.Files[item.Path]; ok && prev.Seed {
				e.Diag(envelope.Info, "SEED_LEFT",
					item.Path+" was seeded and is agent-owned; the file stays, only its lock entry was dropped",
					"delete the file by hand if the state it holds is no longer wanted")
				continue
			}
		}
		// refcount rule: every ref must be gone from the resolution.
		_ = os.Remove(filepath.Join(g.Repo, item.Path))
		_ = os.Remove(g.basePath(item.Path))
		applied = append(applied, item.Path+" (removed)")
	}
	lock := &lockfile.Lock{
		Version:    1,
		ResolvedAt: time.Now().UTC().Format(time.RFC3339),
		Engine:     g.engineID(),
		Layers:     r.layers,
		Files:      map[string]*lockfile.FileEntry{},
	}
	lock.Adapter.Name = r.adapter.Name
	lock.Adapter.Version = r.adapter.Version
	for path, rf := range r.rendered {
		if inList(p.Unmanaged, path) {
			continue // not ours; adopting it is the human's move
		}
		var shadowed []string
		if entry, ok := r.set.Entries[path]; ok {
			for _, s := range entry.Shadowed {
				shadowed = append(shadowed, s.Component)
			}
		}
		hash := rf.Hash
		// A drifted path that was not merged keeps its
		// installed-content hash: the local edit must stay visible as
		// drift, or the next apply would read it as an ordinary
		// change and clobber it. A merged path banks the new
		// upstream hash — the local edit stays visible because the
		// merged tree content still differs from it.
		if inDrift(p, path) && !merged[path] && prevLock != nil {
			if prev, ok := prevLock.Files[path]; ok {
				hash = prev.Hash
			}
		}
		lock.Files[path] = &lockfile.FileEntry{
			Layer: rf.Layer, Component: strings.Join(rf.Refs, ", "),
			Hash: hash, Seed: rf.Seed, Shadowed: shadowed, Refs: rf.Refs,
		}
	}
	if err := lock.Write(g.Repo); err != nil {
		e.Fail("LOCKFILE_WRITE_FAILED", err.Error(), "check permissions")
		return
	}
	sort.Strings(applied)
	e.Result = map[string]any{"applied": applied, "lockfile": lockfile.Name}
	// spec/journal.md rule 1: every completed state change is history —
	// including the conflicts this apply left behind (rule 2).
	var layerIDs []string
	for _, l := range r.layers {
		layerIDs = append(layerIDs, l.ID)
	}
	g.journal(e, "apply", map[string]any{"applied": applied, "conflicts": conflicts, "layers": layerIDs})
}

func findItem(p *Plan, path string) *PlanItem {
	for _, list := range [][]PlanItem{p.Adds, p.Changes, p.Drift} {
		for i := range list {
			if list[i].Path == path {
				return &list[i]
			}
		}
	}
	return nil
}

func inList(items []PlanItem, path string) bool {
	for _, it := range items {
		if it.Path == path {
			return true
		}
	}
	return false
}

func inDrift(p *Plan, path string) bool {
	for _, d := range p.Drift {
		if d.Path == path {
			return true
		}
	}
	return false
}
