package engine

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Cost surfaces beyond the instruction file (spec/resolution.md rule
// 6, docs/lifecycle.md §1.4): a component's cost is what the agent
// pays on every turn it is loaded, and a report that shows benefit
// without cost is half a report. Tiers by when the price is paid —
// always (every turn), per session (read once by convention), or
// conditionally (on invocation). Reporting only: no thresholds, no
// warnings — inventing a cutoff without evidence would be exactly
// the unearned number the fitness doc forbids.

type SkillCost struct {
	Skill            string `json:"skill"`
	DescriptionBytes int    `json:"description_bytes"` // listed to the agent every session
	BodyBytes        int    `json:"body_bytes"`        // paid only on invocation
}

type Costs struct {
	Always      map[string]int `json:"always"`      // path -> bytes, paid every turn
	PerSession  map[string]int `json:"per_session"` // path -> bytes on disk today (state grows after install)
	Conditional []SkillCost    `json:"conditional,omitempty"`
	NotPriced   []string       `json:"not_priced"`
	Misreading  string         `json:"misreading"`
}

const costsMisreading = "bytes price the load, not the worth — a cheap rule that earns nothing " +
	"is overpriced and a costly one that prevents failures is a bargain; cut by necessity, never by size"

// computeCosts prices the whole rendered composition. Seeds are
// priced at their on-disk size when installed — the agent-owned file
// a session actually reads — falling back to the seed template before
// first install.
func (g *Engine) computeCosts(r *resolution) *Costs {
	c := &Costs{
		Always:     map[string]int{},
		PerSession: map[string]int{},
		NotPriced: []string{
			"scripts: executed, never loaded — output enters context only when run",
			"vendor tree (.rein/vendor): never enters context",
			"hooks: no adapter exposes a hook surface yet",
		},
		Misreading: costsMisreading,
	}
	if rf, ok := r.rendered[r.adapter.InstructionFile.Path]; ok {
		c.Always[rf.Path] = len(rf.Content)
	}
	skillsPrefix := r.adapter.Skills.Dir + "/"
	for path, rf := range r.rendered {
		switch {
		case rf.Seed:
			size := len(rf.Content)
			if fi, err := os.Stat(filepath.Join(g.Repo, path)); err == nil {
				size = int(fi.Size())
			}
			c.PerSession[path] = size
		case strings.HasPrefix(path, skillsPrefix):
			name := strings.SplitN(strings.TrimPrefix(path, skillsPrefix), "/", 2)[0]
			desc, body := splitSkill(rf.Content)
			c.Conditional = append(c.Conditional, SkillCost{
				Skill: name, DescriptionBytes: desc, BodyBytes: body,
			})
		}
	}
	sort.Slice(c.Conditional, func(i, j int) bool { return c.Conditional[i].Skill < c.Conditional[j].Skill })
	return c
}

// splitSkill prices a skill's two loads: the frontmatter description
// (listed to the agent every session) and the body (loaded on
// invocation). A file without parseable frontmatter is all body.
func splitSkill(content []byte) (descriptionBytes, bodyBytes int) {
	const fence = "---\n"
	s := string(content)
	if !strings.HasPrefix(s, fence) {
		return 0, len(content)
	}
	end := strings.Index(s[len(fence):], fence)
	if end < 0 {
		return 0, len(content)
	}
	front := s[len(fence) : len(fence)+end]
	body := s[len(fence)+end+len(fence):]
	var meta struct {
		Description string `yaml:"description"`
	}
	if yaml.Unmarshal([]byte(front), &meta) != nil {
		return 0, len(content)
	}
	return len(bytes.TrimSpace([]byte(meta.Description))), len(body)
}
