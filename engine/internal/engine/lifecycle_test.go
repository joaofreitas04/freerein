package engine_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joaofreitas04/freerein/engine/internal/engine"
	"github.com/joaofreitas04/freerein/engine/internal/envelope"
)

// `rein probes` lists the affordance vocabulary (spec rule 4).
func TestProbesCommand(t *testing.T) {
	g, _ := newRepo(t, "claude-code")
	e := envelope.New("probes")
	g.Probes(e)
	if !e.OK {
		t.Fatalf("probes failed: %+v", e.Diagnostics)
	}
	b, _ := json.Marshal(e.Result)
	if !strings.Contains(string(b), `"git"`) {
		t.Fatalf("probes must list the git probe, got %s", b)
	}
}

func addressedManifest(name, frag, address string) string {
	return `name: ` + name + `
kind: extension
version: 0.1.0
subsystem: feedback
rung: instruction
rent:
  class: amplifier
addresses:
  - ` + address + `
provides:
  - AGENTS.md.d/` + frag + `
description: a test extension
`
}

// plan prices the composition (spec/resolution.md rule 6) and warns
// when two components address the same failure surface (manifest
// rule 7).
func TestPlanBudgetAndOverlap(t *testing.T) {
	g, repo := newRepo(t, "claude-code")
	writeComponent(t, filepath.Join(repo, "vendor/ext-a"),
		addressedManifest("ext-a", "60-a.md", "premature-completion"),
		map[string]string{"AGENTS.md.d/60-a.md": "## A\n\nrule a\n"})
	writeComponent(t, filepath.Join(repo, "vendor/ext-b"),
		addressedManifest("ext-b", "61-b.md", "premature-completion"),
		map[string]string{"AGENTS.md.d/61-b.md": "## B\n\nrule b\n"})
	cfg := "adapter: claude-code\npresets: []\nextensions: [\"./vendor/ext-a\", \"./vendor/ext-b\"]\n"
	if err := os.WriteFile(filepath.Join(repo, "harness.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	e := envelope.New("plan")
	g.Plan(e)
	if !e.OK {
		t.Fatalf("plan failed: %+v", e.Diagnostics)
	}
	p, ok := e.Result.(*engine.Plan)
	if !ok {
		t.Fatalf("plan result is %T, want *engine.Plan", e.Result)
	}
	if p.Budget == nil || p.Budget.Bytes == 0 || p.Budget.MaxBytes != 32768 || p.Budget.Fragments < 5 {
		t.Fatalf("plan must price the instruction file, got %+v", p.Budget)
	}
	if !diagCodes(e)["ADDRESSES_OVERLAP"] {
		t.Fatalf("two components sharing an address must warn, got %+v", e.Diagnostics)
	}
	// doctor reports the same composition advisory after install
	apply(t, g)
	e = envelope.New("doctor")
	g.Doctor(e)
	if !diagCodes(e)["ADDRESSES_OVERLAP"] {
		t.Fatalf("doctor must repeat the overlap advisory, got %+v", e.Diagnostics)
	}
	// a near-limit instruction file warns before the render refuses
	big := "## Big\n\n" + strings.Repeat("x", 27000) + "\n"
	writeComponent(t, filepath.Join(repo, "vendor/ext-c"),
		addressedManifest("ext-c", "62-c.md", "cold-start"),
		map[string]string{"AGENTS.md.d/62-c.md": big})
	cfg = "adapter: claude-code\npresets: []\nextensions: [\"./vendor/ext-a\", \"./vendor/ext-b\", \"./vendor/ext-c\"]\n"
	if err := os.WriteFile(filepath.Join(repo, "harness.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	e = envelope.New("plan")
	g.Plan(e)
	if !diagCodes(e)["NEAR_CONTEXT_BUDGET"] {
		t.Fatalf("a near-limit instruction file must warn, got %+v", e.Diagnostics)
	}
}

// The journal is append-only history: every completed state change
// appends, a no-op apply does not (spec/journal.md).
func TestJournalAppends(t *testing.T) {
	g, repo := newRepo(t, "claude-code")
	journalPath := filepath.Join(repo, ".rein/journal.jsonl")
	kinds := func() []string {
		b, err := os.ReadFile(journalPath)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			t.Fatal(err)
		}
		var out []string
		for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
			var entry struct {
				Kind   string `json:"kind"`
				At     string `json:"at"`
				Engine string `json:"engine"`
			}
			if err := json.Unmarshal([]byte(line), &entry); err != nil {
				t.Fatalf("journal line is not JSON: %q", line)
			}
			if entry.Kind == "" || entry.At == "" || entry.Engine == "" {
				t.Fatalf("journal entry missing required fields: %q", line)
			}
			out = append(out, entry.Kind)
		}
		return out
	}
	apply(t, g)
	if got := kinds(); len(got) != 1 || got[0] != "apply" {
		t.Fatalf("first apply must journal once, got %v", got)
	}
	// a no-op apply records nothing — no state changed
	e := envelope.New("apply")
	g.Apply(e, true)
	if got := kinds(); len(got) != 1 {
		t.Fatalf("no-op apply must not journal, got %v", got)
	}
	// declaring and applying an extension journals the apply
	writeComponent(t, filepath.Join(repo, "vendor/toy-ext"), extManifest,
		map[string]string{"AGENTS.md.d/50-toy.md": "## Toy\n\ntoy rule\n"})
	declareExtension(t, repo, "./vendor/toy-ext")
	apply(t, g)
	if got := kinds(); len(got) != 2 || got[1] != "apply" {
		t.Fatalf("changed apply must journal, got %v", got)
	}
	// undeclaring journals the remove, and the removal apply journals
	e = envelope.New("remove")
	g.Remove(e, "./vendor/toy-ext")
	if !e.OK {
		t.Fatalf("remove failed: %+v", e.Diagnostics)
	}
	apply(t, g)
	if got := kinds(); len(got) != 4 || got[2] != "remove" || got[3] != "apply" {
		t.Fatalf("remove + apply must journal, got %v", got)
	}
}

// Un-installing a component never deletes a seeded file: the agent
// owns it (spec/component-manifest.md, Seed disposal).
func TestSeedLeftOnRemove(t *testing.T) {
	g, repo := newRepo(t, "claude-code")
	manifest := `name: seeder
kind: extension
version: 0.1.0
subsystem: state
rung: instruction
rent:
  class: amplifier
provides:
  - .rein/state/NOTES.md
seeds:
  - .rein/state/NOTES.md
description: a seeding extension
`
	writeComponent(t, filepath.Join(repo, "vendor/seeder"), manifest,
		map[string]string{".rein/state/NOTES.md": "# Notes\n(template)\n"})
	declareExtension(t, repo, "./vendor/seeder")
	apply(t, g)
	seeded := filepath.Join(repo, ".rein/state/NOTES.md")
	if err := os.WriteFile(seeded, []byte("# Notes\nAGENT STATE\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// undeclare and apply the removal
	cfg := "adapter: claude-code\npresets: []\nextensions: []\n"
	if err := os.WriteFile(filepath.Join(repo, "harness.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	e := apply(t, g)
	if !diagCodes(e)["SEED_LEFT"] {
		t.Fatalf("removing a seeding component must report SEED_LEFT, got %+v", e.Diagnostics)
	}
	b, err := os.ReadFile(seeded)
	if err != nil || !strings.Contains(string(b), "AGENT STATE") {
		t.Fatal("apply deleted an agent-owned seed — spec violated")
	}
	lockBytes, _ := os.ReadFile(filepath.Join(repo, "harness.lock"))
	if strings.Contains(string(lockBytes), "NOTES.md") {
		t.Fatal("the seed's lock entry must be dropped on removal")
	}
}
