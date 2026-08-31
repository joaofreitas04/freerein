package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joaofreitas04/freerein/engine/internal/component"
	"github.com/joaofreitas04/freerein/engine/internal/envelope"
	"github.com/joaofreitas04/freerein/engine/internal/registry"
)

// writeConfig rewrites harness.yaml canonically (comments are not
// preserved — the file is machine-managed; human knobs live in the
// values, not in prose).
func (g *Engine) writeConfig(cfg *Config) error {
	var b strings.Builder
	b.WriteString("# FreeRein harness declaration — see spec/resolution.md\n")
	fmt.Fprintf(&b, "adapter: %s\n", cfg.Adapter)
	if cfg.Registry != "" {
		fmt.Fprintf(&b, "registry: %q\n", cfg.Registry)
	}
	writeList := func(name string, xs []string) {
		if len(xs) == 0 {
			fmt.Fprintf(&b, "%s: []\n", name)
			return
		}
		fmt.Fprintf(&b, "%s:\n", name)
		for _, x := range xs {
			fmt.Fprintf(&b, "  - %q\n", x)
		}
	}
	writeList("presets", cfg.Presets)
	writeList("extensions", cfg.Extensions)
	return os.WriteFile(filepath.Join(g.Repo, ConfigName), []byte(b.String()), 0o644)
}

func (g *Engine) openRegistry(e *envelope.Envelope, cfg *Config, override string) *registry.Index {
	src := override
	if src == "" {
		src = cfg.Registry
	}
	if src == "" {
		e.Fail("NO_REGISTRY", "no registry configured",
			"set `registry:` in "+ConfigName+" or pass --registry <url-or-path>")
		return nil
	}
	idx, err := registry.LoadIndex(src)
	if err != nil {
		e.Fail("REGISTRY_UNREACHABLE", err.Error(), "check the registry URL/path and your network")
		return nil
	}
	return idx
}

func splitRef(ref string) (name, version string) {
	if i := strings.Index(ref, "@"); i >= 0 {
		return ref[:i], ref[i+1:]
	}
	return ref, ""
}

// vendorFetch fetches and unpacks name@version into the vendor tree
// and validates that it loads as a component of some kind.
func (g *Engine) vendorFetch(e *envelope.Envelope, idx *registry.Index, name, version string) (*component.Loaded, string) {
	rel, err := idx.Lookup(name, version)
	if err != nil {
		e.Fail("NOT_IN_REGISTRY", err.Error(), "check the name and version with `rein info`")
		return nil, ""
	}
	ref := name + "@" + version
	dir := filepath.Join(g.Repo, VendorDir, ref)
	if err := os.RemoveAll(dir); err != nil {
		e.Fail("WRITE_FAILED", err.Error(), "check permissions on "+VendorDir)
		return nil, ""
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		e.Fail("WRITE_FAILED", err.Error(), "check permissions on "+VendorDir)
		return nil, ""
	}
	if err := idx.Fetch(rel, dir); err != nil {
		_ = os.RemoveAll(dir)
		e.Fail("FETCH_FAILED", ref+": "+err.Error(),
			"a sha mismatch means the registry and archive disagree — do not install; report it to the registry")
		return nil, ""
	}
	parent, base := filepath.Split(filepath.Clean(dir))
	c, err := component.LoadDir(os.DirFS(filepath.Clean(parent)), base)
	if err != nil {
		_ = os.RemoveAll(dir)
		e.Fail("COMPONENT_INVALID", ref+": "+err.Error(), "the archive is not a valid component; report it to its author")
		return nil, ""
	}
	if c.Manifest.Name != name || c.Manifest.Version != version {
		_ = os.RemoveAll(dir)
		e.Fail("COMPONENT_MISMATCH",
			fmt.Sprintf("archive says %s, registry says %s", c.Manifest.Ref(), ref),
			"the registry entry is mislabeled; report it")
		return nil, ""
	}
	return c, ref
}

func (g *Engine) Add(e *envelope.Envelope, target, registryOverride string) {
	cfg, err := g.readConfig()
	if err != nil {
		e.Fail("NOT_INITIALIZED", err.Error(), "run `rein init` first")
		return
	}
	idx := g.openRegistry(e, cfg, registryOverride)
	if idx == nil {
		return
	}
	name, version := splitRef(target)
	if version == "" {
		v, _, err := idx.Latest(name)
		if err != nil {
			e.Fail("NOT_IN_REGISTRY", err.Error(), "check the name; `rein info <name>` shows registry metadata")
			return
		}
		version = v
	}
	c, ref := g.vendorFetch(e, idx, name, version)
	if c == nil {
		return
	}
	list := &cfg.Extensions
	if c.Manifest.Kind == "preset" {
		list = &cfg.Presets
	} else if c.Manifest.Kind != "extension" {
		e.Fail("KIND_INVALID", ref+" has kind "+c.Manifest.Kind+", which cannot be added to a harness",
			"only extensions and presets are addable")
		return
	}
	for i, existing := range *list {
		en, _ := splitRef(strings.TrimPrefix(existing, "registry:"))
		if en == name {
			(*list)[i] = ref // version change in place
			goto write
		}
	}
	*list = append(*list, ref)
write:
	if err := g.writeConfig(cfg); err != nil {
		e.Fail("WRITE_FAILED", err.Error(), "check permissions")
		return
	}
	e.Result = map[string]any{
		"added": ref, "kind": c.Manifest.Kind, "vendored": VendorDir + "/" + ref,
		"subsystem": c.Manifest.Subsystem, "rung": c.Manifest.Rung, "rent": c.Manifest.Rent.Class,
	}
	e.Diag(envelope.Info, "ADDED", ref+" vendored and declared; nothing installed yet",
		"run `rein plan` to review, then `rein apply --yes`")
	g.journal(e, "add", map[string]any{"component": ref, "kind": c.Manifest.Kind})
}

func (g *Engine) Remove(e *envelope.Envelope, name string) {
	cfg, err := g.readConfig()
	if err != nil {
		e.Fail("NOT_INITIALIZED", err.Error(), "run `rein init` first")
		return
	}
	removed := ""
	for _, list := range []*[]string{&cfg.Presets, &cfg.Extensions} {
		var kept []string
		for _, entry := range *list {
			en, _ := splitRef(entry)
			if entry == name || en == name {
				removed = entry
				continue
			}
			kept = append(kept, entry)
		}
		*list = kept
	}
	if removed == "" {
		e.Fail("NOT_DECLARED", name+" is not declared in "+ConfigName, "run `rein dump` to see what is composed")
		return
	}
	if err := g.writeConfig(cfg); err != nil {
		e.Fail("WRITE_FAILED", err.Error(), "check permissions")
		return
	}
	// vendor dir is only removed when nothing references it anymore
	if registryRef.MatchString(removed) && !stillReferenced(cfg, removed) {
		_ = os.RemoveAll(filepath.Join(g.Repo, VendorDir, removed))
	}
	e.Result = map[string]any{"removed": removed}
	e.Diag(envelope.Info, "REMOVED", removed+" undeclared; its installed files are still in the tree",
		"run `rein plan` to review the removals, then `rein apply --yes`")
	g.journal(e, "remove", map[string]any{"component": removed})
}

func stillReferenced(cfg *Config, ref string) bool {
	for _, list := range [][]string{cfg.Presets, cfg.Extensions} {
		for _, entry := range list {
			if entry == ref {
				return true
			}
		}
	}
	return false
}

// Info shows what a component would add, before anything is written —
// the pre-install half of the inspectability guarantee.
func (g *Engine) Info(e *envelope.Envelope, target, registryOverride string) {
	var c *component.Loaded
	var origin string
	switch {
	case strings.HasPrefix(target, "./") || strings.HasPrefix(target, "../") || strings.HasPrefix(target, "/"):
		si, err := g.loadSource(target)
		if err != nil {
			e.Fail("SOURCE_INVALID", err.Error(), "point at a directory containing component.yaml")
			return
		}
		c, origin = si.comp, "local path"
	default:
		name, version := splitRef(target)
		if dir := filepath.Join(g.Repo, VendorDir, target); version != "" && exists(dir) {
			parent, base := filepath.Split(filepath.Clean(dir))
			lc, err := component.LoadDir(os.DirFS(filepath.Clean(parent)), base)
			if err != nil {
				e.Fail("COMPONENT_INVALID", err.Error(), "re-fetch with `rein add "+target+"`")
				return
			}
			c, origin = lc, "vendored"
		} else {
			cfg, err := g.readConfig()
			if err != nil {
				e.Fail("NOT_INITIALIZED", err.Error(), "run `rein init` first, or pass a local path")
				return
			}
			idx := g.openRegistry(e, cfg, registryOverride)
			if idx == nil {
				return
			}
			if version == "" {
				v, _, err := idx.Latest(name)
				if err != nil {
					e.Fail("NOT_IN_REGISTRY", err.Error(), "check the name")
					return
				}
				version = v
			}
			rel, err := idx.Lookup(name, version)
			if err != nil {
				e.Fail("NOT_IN_REGISTRY", err.Error(), "check the name and version")
				return
			}
			tmp, err := os.MkdirTemp("", "rein-info-")
			if err != nil {
				e.Fail("WRITE_FAILED", err.Error(),
					"check free space and permissions on the system temp dir ($TMPDIR)")
				return
			}
			defer os.RemoveAll(tmp)
			if err := idx.Fetch(rel, tmp); err != nil {
				e.Fail("FETCH_FAILED", err.Error(), "a sha mismatch means registry and archive disagree — do not install")
				return
			}
			lc, err := component.LoadDir(os.DirFS(filepath.Dir(tmp)), filepath.Base(tmp))
			if err != nil {
				e.Fail("COMPONENT_INVALID", err.Error(), "the archive is not a valid component")
				return
			}
			c, origin = lc, "registry (fetched for inspection, discarded)"
		}
	}
	m := c.Manifest
	e.Result = map[string]any{
		"name": m.Name, "version": m.Version, "kind": m.Kind, "origin": origin,
		"subsystem": m.Subsystem, "rung": m.Rung,
		"rent": m.Rent, "provides": m.Provides, "requires": m.Requires,
		"conflicts": m.Conflicts, "description": strings.TrimSpace(m.Description),
	}
}

func exists(p string) bool { _, err := os.Stat(p); return err == nil }

// Upgrade checks every registry-declared component for newer versions.
// Two-phase: without --yes it reports; with --yes it re-vendors and
// updates the declaration, leaving installation to plan/apply (where
// three-way merge handles locally edited files).
func (g *Engine) Upgrade(e *envelope.Envelope, yes bool, registryOverride string) {
	cfg, err := g.readConfig()
	if err != nil {
		e.Fail("NOT_INITIALIZED", err.Error(), "run `rein init` first")
		return
	}
	var refs []string
	for _, list := range [][]string{cfg.Presets, cfg.Extensions} {
		for _, entry := range list {
			if registryRef.MatchString(entry) {
				refs = append(refs, entry)
			}
		}
	}
	if len(refs) == 0 {
		e.Diag(envelope.Info, "NOTHING_TO_UPGRADE", "no registry components are declared", "")
		e.Result = map[string]any{"upgrades": []string{}}
		return
	}
	idx := g.openRegistry(e, cfg, registryOverride)
	if idx == nil {
		return
	}
	type up struct{ From, To string }
	var ups []up
	for _, ref := range refs {
		name, version := splitRef(ref)
		latest, _, err := idx.Latest(name)
		if err != nil {
			e.Diag(envelope.Warning, "GONE_FROM_REGISTRY", ref+": "+err.Error(),
				"the component may have been unpublished; consider `rein remove "+name+"`")
			continue
		}
		if registry.CompareVersions(latest, version) > 0 {
			ups = append(ups, up{From: ref, To: name + "@" + latest})
		}
	}
	if len(ups) == 0 {
		e.Diag(envelope.Info, "UP_TO_DATE", "every registry component is at its latest version", "")
		e.Result = map[string]any{"upgrades": []string{}}
		return
	}
	if !yes {
		e.ConfirmRequired = ups
		e.Diag(envelope.Info, "CONFIRM_REQUIRED", "upgrades are available",
			"show this list to the human; re-run `rein upgrade --yes` once approved")
		return
	}
	applied := []string{}
	for _, u := range ups {
		name, version := splitRef(u.To)
		c, _ := g.vendorFetch(e, idx, name, version)
		if c == nil {
			return
		}
		for _, list := range []*[]string{&cfg.Presets, &cfg.Extensions} {
			for i, entry := range *list {
				if entry == u.From {
					(*list)[i] = u.To
				}
			}
		}
		_ = os.RemoveAll(filepath.Join(g.Repo, VendorDir, u.From))
		applied = append(applied, u.From+" -> "+u.To)
	}
	if err := g.writeConfig(cfg); err != nil {
		e.Fail("WRITE_FAILED", err.Error(), "check permissions")
		return
	}
	e.Result = map[string]any{"upgraded": applied}
	e.Diag(envelope.Info, "VENDORED", "new versions vendored and declared; nothing installed yet",
		"run `rein plan` to review (locally edited files will three-way merge), then `rein apply --yes`")
	g.journal(e, "upgrade", map[string]any{"upgraded": applied})
}
