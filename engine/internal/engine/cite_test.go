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

// spec/citation.md rule 2: the render carries the ids, derived from
// the same function the validator uses.
func TestRenderCarriesFragmentMarkers(t *testing.T) {
	g, repo := appliedRepo(t)
	b, err := os.ReadFile(filepath.Join(repo, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{
		"<!-- rein:fragment instructions-base:00-base -->",
		"<!-- rein:fragment state-base:30-state -->",
	} {
		if !strings.Contains(string(b), marker) {
			t.Fatalf("render must carry %q, got:\n%s", marker, b)
		}
	}
	_ = g
}

func TestCiteRecordsAndReads(t *testing.T) {
	g, repo := appliedRepo(t)

	// unknown id: refused with the vocabulary
	e := envelope.New("cite")
	g.Cite(e, "instructions-base:99-invented")
	if e.OK || !diagCodes(e)["INVALID_ARGUMENT"] {
		t.Fatalf("unknown id must refuse, got %+v", e.Diagnostics)
	}
	if !strings.Contains(e.Diagnostics[0].Fix, "instructions-base:00-base") {
		t.Fatalf("refusal must teach the vocabulary, got %q", e.Diagnostics[0].Fix)
	}

	// record twice
	for i := 0; i < 2; i++ {
		e = envelope.New("cite")
		g.Cite(e, "instructions-base:00-base")
		if !e.OK {
			t.Fatalf("cite failed: %+v", e.Diagnostics)
		}
	}
	res := envelope.New("cite")
	g.Cite(res, "instructions-base:00-base")
	if res.Result.(map[string]any)["count"].(int) != 3 {
		t.Fatalf("count must accumulate, got %v", res.Result)
	}

	// the store is deterministic, versioned, and not the journal
	b, err := os.ReadFile(filepath.Join(repo, engine.CitationsPath))
	if err != nil {
		t.Fatal(err)
	}
	var store struct {
		FormatVersion string `json:"formatVersion"`
	}
	if json.Unmarshal(b, &store) != nil || store.FormatVersion != engine.CitationFormatVersion {
		t.Fatalf("store must carry the contract version, got %s", b)
	}
	if _, err := os.Stat(filepath.Join(repo, engine.JournalName)); err == nil {
		j, _ := os.ReadFile(filepath.Join(repo, engine.JournalName))
		if strings.Contains(string(j), "cite") {
			t.Fatal("citations are counters, never journal entries")
		}
	}

	// read-back: counts joined against the composition, zero-count
	// fragments ranked as decay candidates, misreading attached
	e = envelope.New("cite")
	g.Cite(e, "")
	if !e.OK {
		t.Fatalf("read-back failed: %+v", e.Diagnostics)
	}
	r := e.Result.(map[string]any)
	decay := r["decay_candidates"].([]string)
	for _, d := range decay {
		if d == "instructions-base:00-base" {
			t.Fatal("a cited fragment must not be a decay candidate")
		}
	}
	if len(decay) == 0 {
		t.Fatal("uncited fragments must rank as decay candidates")
	}
	if !strings.Contains(r["misreading"].(string), "internalized") {
		t.Fatal("the count must ship with its misreading (lifecycle §1.5)")
	}
}

// The instruction file stays byte-stable as counts change
// (spec/citation.md rule 3).
func TestCitingChangesNoRenderedBytes(t *testing.T) {
	g, repo := appliedRepo(t)
	before, _ := os.ReadFile(filepath.Join(repo, "CLAUDE.md"))
	e := envelope.New("cite")
	g.Cite(e, "instructions-base:00-base")
	if !e.OK {
		t.Fatalf("cite failed: %+v", e.Diagnostics)
	}
	after, _ := os.ReadFile(filepath.Join(repo, "CLAUDE.md"))
	if string(before) != string(after) {
		t.Fatal("citing must never touch the instruction file")
	}
}

// spec/resolution.md rule 6: plan prices the whole composition, in
// tiers by when the price is paid, reporting only — and the number
// ships with its misreading.
func TestPlanPricesTheWholeComposition(t *testing.T) {
	g, repo := appliedRepo(t)
	e := envelope.New("plan")
	g.Plan(e)
	if !e.OK {
		t.Fatalf("plan failed: %+v", e.Diagnostics)
	}
	b, _ := json.Marshal(e.Result)
	var p struct {
		Costs struct {
			Always      map[string]int `json:"always"`
			PerSession  map[string]int `json:"per_session"`
			Conditional []struct {
				Skill            string `json:"skill"`
				DescriptionBytes int    `json:"description_bytes"`
				BodyBytes        int    `json:"body_bytes"`
			} `json:"conditional"`
			NotPriced  []string `json:"not_priced"`
			Misreading string   `json:"misreading"`
		} `json:"costs"`
	}
	if err := json.Unmarshal(b, &p); err != nil {
		t.Fatal(err)
	}
	c := p.Costs
	if c.Always["CLAUDE.md"] == 0 {
		t.Fatalf("the instruction file is the always tier, got %+v", c.Always)
	}
	if c.PerSession[".rein/state/PROGRESS.md"] == 0 || c.PerSession[".rein/state/DEBT.md"] == 0 {
		t.Fatalf("seeds are the per-session tier, got %+v", c.PerSession)
	}
	if len(c.Conditional) == 0 {
		t.Fatal("skills are the conditional tier")
	}
	for _, s := range c.Conditional {
		if s.DescriptionBytes == 0 || s.BodyBytes == 0 {
			t.Fatalf("a skill has two loads (listing + invocation), got %+v", s)
		}
	}
	if len(c.NotPriced) == 0 || c.Misreading == "" {
		t.Fatal("unpriced surfaces and the misreading must be named (lifecycle §1.5)")
	}

	// seeds are priced as the agent-owned file a session actually
	// reads: growth after install must show up
	before := c.PerSession[".rein/state/PROGRESS.md"]
	f, _ := os.OpenFile(filepath.Join(repo, ".rein/state/PROGRESS.md"), os.O_APPEND|os.O_WRONLY, 0o644)
	f.WriteString(strings.Repeat("session history accretes\n", 40))
	f.Close()
	e = envelope.New("plan")
	g.Plan(e)
	b, _ = json.Marshal(e.Result)
	json.Unmarshal(b, &p)
	if p.Costs.PerSession[".rein/state/PROGRESS.md"] <= before {
		t.Fatalf("a grown seed must price at its on-disk size, got %d <= %d",
			p.Costs.PerSession[".rein/state/PROGRESS.md"], before)
	}
}
