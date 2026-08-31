// Package resolve implements spec/resolution.md: ordered layers,
// first match per path, whole-file, no deep merge — and the render
// step that walks the resolved set through a host adapter.
package resolve

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/joaofreitas04/freerein/engine/internal/adapter"
	"github.com/joaofreitas04/freerein/engine/internal/component"
)

// FragmentDir is the host-neutral instruction-fragment prefix; the
// renderer concatenates these (sorted) into the adapter's
// instruction file.
const FragmentDir = "AGENTS.md.d/"

// SkillsDir is the host-neutral skills prefix; the renderer maps it
// onto the adapter's skills directory.
const SkillsDir = "skills/"

type Provider struct {
	Layer     string `json:"layer"`
	Component string `json:"component"`
}

type Entry struct {
	Path     string     `json:"path"`
	Winner   Provider   `json:"winner"`
	Shadowed []Provider `json:"shadowed"`
	Content  []byte     `json:"-"`
	Seed     bool       `json:"seed,omitempty"` // install-if-absent, agent-owned after
}

// Layer is one rung of the stack, highest priority first.
type Layer struct {
	ID         string
	Components []*component.Loaded
}

type Set struct {
	Layers  []string
	Entries map[string]*Entry
}

// Resolve applies layers in priority order. First match per path
// wins; later (lower-priority) providers of the same path are
// recorded as shadowed, so removal can restore them.
func Resolve(layers []Layer) (*Set, error) {
	s := &Set{Entries: map[string]*Entry{}}
	for _, l := range layers {
		s.Layers = append(s.Layers, l.ID)
		for _, c := range l.Components {
			// spec rule: a preset may only shadow, never add.
			for _, p := range c.Manifest.Provides {
				if c.Manifest.Kind == "preset" {
					if _, exists := s.Entries[p]; !exists && !providedBelow(layers, l, p) {
						return nil, fmt.Errorf("preset %s provides %q which no other layer provides — a preset cannot add capability", c.Manifest.Name, p)
					}
				}
				prov := Provider{Layer: l.ID, Component: c.Manifest.Ref()}
				if e, exists := s.Entries[p]; exists {
					e.Shadowed = append(e.Shadowed, prov)
				} else {
					seed := false
					for _, sd := range c.Manifest.Seeds {
						if sd == p {
							seed = true
						}
					}
					s.Entries[p] = &Entry{Path: p, Winner: prov, Content: c.Files[p], Seed: seed}
				}
			}
		}
	}
	return s, nil
}

func providedBelow(layers []Layer, above Layer, p string) bool {
	seen := false
	for _, l := range layers {
		if l.ID == above.ID {
			seen = true
			continue
		}
		if !seen {
			continue
		}
		for _, c := range l.Components {
			for _, q := range c.Manifest.Provides {
				if q == p {
					return true
				}
			}
		}
	}
	return false
}

// Rendered is a final file destined for the target repo.
type Rendered struct {
	Path    string   `json:"path"`
	Hash    string   `json:"hash"`
	Refs    []string `json:"refs"`  // contributing components
	Layer   string   `json:"layer"` // winning layer ("render" for composed files)
	Seed    bool     `json:"seed,omitempty"`
	Content []byte   `json:"-"`
	Mode    uint32   `json:"-"`
}

// FragmentID is a fragment's stable citation identity
// (spec/citation.md): component name, version-free so counts survive
// upgrades, then the fragment filename without extension —
// "instructions-base:00-base". The render marker and the cite
// validator both derive it here, so they cannot disagree.
func FragmentID(componentRef, path string) string {
	name := componentRef
	if i := strings.IndexByte(name, '@'); i >= 0 {
		name = name[:i]
	}
	base := strings.TrimPrefix(path, FragmentDir)
	base = strings.TrimSuffix(base, ".md")
	return name + ":" + base
}

// Render walks the resolved set through the adapter. Instruction
// fragments are concatenated (sorted by path) into the adapter's
// instruction file, each preceded by its citation marker
// (spec/citation.md); every other path passes through unchanged.
func Render(s *Set, a *adapter.Adapter) (map[string]*Rendered, error) {
	out := map[string]*Rendered{}
	var fragPaths []string
	for p := range s.Entries {
		if strings.HasPrefix(p, FragmentDir) {
			fragPaths = append(fragPaths, p)
		}
	}
	sort.Strings(fragPaths)
	if len(fragPaths) > 0 {
		var b strings.Builder
		var refs []string
		b.WriteString("<!-- rendered by rein; do not hand-edit — edit .rein/overrides/ and run `rein apply` -->\n\n")
		for _, p := range fragPaths {
			e := s.Entries[p]
			b.WriteString("<!-- rein:fragment " + FragmentID(e.Winner.Component, p) + " -->\n")
			b.Write(e.Content)
			if !strings.HasSuffix(string(e.Content), "\n") {
				b.WriteString("\n")
			}
			b.WriteString("\n")
			refs = append(refs, e.Winner.Component)
		}
		content := []byte(strings.TrimSuffix(b.String(), "\n"))
		if a.InstructionFile.MaxBytes > 0 && len(content) > a.InstructionFile.MaxBytes {
			return nil, fmt.Errorf("rendered instruction file is %d bytes, over the adapter's %d limit", len(content), a.InstructionFile.MaxBytes)
		}
		out[a.InstructionFile.Path] = &Rendered{
			Path: a.InstructionFile.Path, Hash: Hash(content),
			Refs: refs, Layer: "render", Content: content, Mode: 0o644,
		}
	}
	for p, e := range s.Entries {
		if strings.HasPrefix(p, FragmentDir) {
			continue
		}
		final := p
		if strings.HasPrefix(p, SkillsDir) {
			final = a.Skills.Dir + "/" + strings.TrimPrefix(p, SkillsDir)
		}
		mode := uint32(0o644)
		if strings.HasPrefix(p, "scripts/") {
			mode = 0o755
		}
		out[final] = &Rendered{
			Path: final, Hash: Hash(e.Content), Refs: []string{e.Winner.Component},
			Layer: e.Winner.Layer, Seed: e.Seed, Content: e.Content, Mode: mode,
		}
	}
	return out, nil
}

func Hash(b []byte) string {
	h := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(h[:])
}
