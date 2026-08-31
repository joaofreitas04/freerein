package engine_test

// The field case corpus (spec/field-report.md, consumer side): every
// detection defect found on a real repo lands here as a red fixture
// before its fix, and the corpus only grows — a fix that turns its
// own case green while turning a standing case red is rejected on
// that evidence. Evidence tags ([bench N] / [real N] / [owner N])
// index the field-test findings recorded in .rein/state/PROGRESS.md
// when each case landed; the fixtures themselves are generalized
// (affordances, never repo content).

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/joaofreitas04/freerein/content"
	"github.com/joaofreitas04/freerein/engine/internal/engine"
	"github.com/joaofreitas04/freerein/engine/internal/envelope"
)

func inspectReport(t *testing.T, repo string) engine.InspectReport {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	g := &engine.Engine{Repo: repo, Content: content.FS}
	e := envelope.New("inspect")
	g.Inspect(e)
	if !e.OK {
		t.Fatalf("inspect failed: %+v", e.Diagnostics)
	}
	b, err := os.ReadFile(filepath.Join(repo, ".rein/out/inspect.json"))
	if err != nil {
		t.Fatalf("inspect must write the report: %v", err)
	}
	var r engine.InspectReport
	if err := json.Unmarshal(b, &r); err != nil {
		t.Fatalf("report is not valid JSON: %v", err)
	}
	return r
}

// [real 10]: a 127-project nx workspace reported monorepo: false —
// nx.json is the canonical marker and was missing from the table.
func TestFieldNxMarksMonorepo(t *testing.T) {
	repo := t.TempDir()
	seedFile(t, repo, "nx.json", `{"targetDefaults":{}}`)
	seedFile(t, repo, "package.json", `{"name":"ws"}`)
	r := inspectReport(t, repo)
	if !r.Toolchain.Monorepo {
		t.Fatal("nx.json must mark a monorepo [real 10]")
	}
}
