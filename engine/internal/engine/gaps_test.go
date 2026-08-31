package engine_test

import (
	"strings"
	"testing"

	"github.com/joaofreitas04/freerein/content"
	"github.com/joaofreitas04/freerein/engine/internal/engine"
	"github.com/joaofreitas04/freerein/engine/internal/envelope"
)

// rein gaps joins absent affordances to the registry's addresses:
// coverage claims, never lift claims. An addressed gap names the
// extension; an unaddressed gap is a creation candidate for a human
// to commission — aggregated, those candidates are the registry's
// demand signal. The misreading travels in the result.
func TestGapsJoinsAffordancesToAddresses(t *testing.T) {
	repo := t.TempDir()
	seedFile(t, repo, ".git/HEAD", "ref: refs/heads/main\n")
	seedFile(t, repo, "go.mod", "module toy\n")

	// a hand-written index: gaps reads coverage from the index alone
	reg := t.TempDir()
	seedFile(t, reg, "index.json",
		`{"components":{"ci-ext":{"1.0.0":{"url":"ci-ext-1.0.0.tar.gz","sha256":"deadbeef","kind":"extension","description":"wires ci","addresses":["ci"]}}}}`)

	g := &engine.Engine{Repo: repo, Content: content.FS}
	e := envelope.New("gaps")
	g.Gaps(e, reg+"/index.json")
	if !e.OK {
		t.Fatalf("gaps failed: %+v", e.Diagnostics)
	}
	res := e.Result.(map[string]any)
	rows := res["gaps"].([]engine.Gap)
	byName := map[string]engine.Gap{}
	for _, r := range rows {
		byName[r.Gap] = r
	}
	if _, ok := byName["git"]; ok {
		t.Fatalf("a present affordance is not a gap, got %+v", rows)
	}
	if _, ok := byName["test-runner"]; ok {
		t.Fatalf("go.mod derives a test candidate — not a gap, got %+v", rows)
	}
	ci, ok := byName["ci"]
	if !ok || len(ci.AddressedBy) != 1 || ci.AddressedBy[0] != "ci-ext@1.0.0" {
		t.Fatalf("the ci gap is addressed by the indexed extension, got %+v", rows)
	}
	if ci.CreationCandidate {
		t.Fatalf("an addressed gap is not a creation candidate, got %+v", ci)
	}
	if ci.Evidence == "" {
		t.Fatalf("each gap carries the probe's own evidence, got %+v", ci)
	}
	for _, name := range []string{"linter", "docs-tree"} {
		row, ok := byName[name]
		if !ok || !row.CreationCandidate || len(row.AddressedBy) != 0 {
			t.Fatalf("%s has no addressing extension and must be a creation candidate, got %+v", name, rows)
		}
	}
	mis, ok := res["misreadings"].([]string)
	if !ok || len(mis) == 0 {
		t.Fatal("coverage-not-lift must travel in the result")
	}
	joined := strings.Join(mis, ";")
	if !strings.Contains(joined, "lift") || !strings.Contains(joined, "commission") {
		t.Fatalf("the misreadings name coverage-vs-lift and the no-auto-author rule, got %v", mis)
	}
}
