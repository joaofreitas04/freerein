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
	"strings"
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

func candidateCommands(r engine.InspectReport) []string {
	var out []string
	for _, c := range r.Tests.Candidates {
		out = append(out, c.Command)
	}
	return out
}

// [real 10]: candidates were null on an nx workspace whose runner is
// mediated by nx itself — a root jest.config.ts was stale bait, and
// project entries in angular.json can be path strings with targets
// living elsewhere. nx.json alone is the manifest fact; run-many is
// its canonical whole-workspace test invocation, exactly as go.mod
// derives `go test ./...` without promising tests exist.
func TestFieldNxWorkspaceRunCandidate(t *testing.T) {
	repo := t.TempDir()
	seedFile(t, repo, "nx.json", `{"targetDefaults":{}}`)
	seedFile(t, repo, "angular.json", `{"version":2,"projects":{"app1":"apps/app1","app2":"apps/app2"}}`)
	seedFile(t, repo, "jest.config.ts", "export default {};\n")
	seedFile(t, repo, "package.json", `{"name":"ws","scripts":{"build":"nx build"}}`)
	r := inspectReport(t, repo)
	cmds := strings.Join(candidateCommands(r), ";")
	if !strings.Contains(cmds, "npx nx run-many -t test") {
		t.Fatalf("an nx workspace must derive its canonical test candidate [real 10], got %v", r.Tests.Candidates)
	}
	if strings.Contains(cmds, "npx ng test") {
		t.Fatalf("nx workspaces must not add per-project ng noise, got %v", r.Tests.Candidates)
	}
}

// [bench 1]: a workspace where commands live in angular.json project
// targets derived zero candidates, so the human ended up naming
// commands — which the setup procedure says should never happen.
// Plain angular has no run-all; per-project candidates are the
// mechanical truth, capped with the cut announced (spec rule 5).
func TestFieldAngularPerProjectCandidates(t *testing.T) {
	repo := t.TempDir()
	seedFile(t, repo, "angular.json",
		`{"projects":{"app-a":{"architect":{"test":{"builder":"x"},"build":{"builder":"x"}}},"lib-b":{"targets":{"test":{"executor":"y"}}},"lib-c":{"architect":{"build":{"builder":"x"}}}}}`)
	seedFile(t, repo, "package.json", `{"name":"ws"}`)
	r := inspectReport(t, repo)
	cmds := strings.Join(candidateCommands(r), ";")
	for _, want := range []string{"npx ng test app-a", "npx ng test lib-b"} {
		if !strings.Contains(cmds, want) {
			t.Fatalf("angular.json test targets must derive candidates [bench 1], want %q in %v", want, r.Tests.Candidates)
		}
	}
	if strings.Contains(cmds, "lib-c") {
		t.Fatalf("a project without a test target is not a candidate, got %v", r.Tests.Candidates)
	}
	if strings.Contains(cmds, "run-many") {
		t.Fatalf("run-many is nx's command, not plain angular's, got %v", r.Tests.Candidates)
	}

	// seven test targets: the list is capped and the cut announces
	// itself in notes rather than silently truncating
	big := t.TempDir()
	projects := `{"projects":{`
	for i, name := range []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7"} {
		if i > 0 {
			projects += ","
		}
		projects += `"` + name + `":{"architect":{"test":{"builder":"x"}}}`
	}
	projects += `}}`
	seedFile(t, big, "angular.json", projects)
	r = inspectReport(t, big)
	ngCount := 0
	for _, c := range candidateCommands(r) {
		if strings.HasPrefix(c, "npx ng test ") {
			ngCount++
		}
	}
	if ngCount != 5 {
		t.Fatalf("per-project candidates must cap at 5, got %d: %v", ngCount, r.Tests.Candidates)
	}
	capped := false
	for _, n := range r.Notes {
		if strings.Contains(n, "capped") && strings.Contains(n, "candidate") {
			capped = true
		}
	}
	if !capped {
		t.Fatalf("a capped candidate list must announce itself in notes, got %v", r.Notes)
	}
}

// [bench 4]: the corpus walk found only CLAUDE.md and missed a
// root-level CONTEXT.md — an instruction-ish file a triage should at
// least see. The table stays curated; CONTEXT.md is common enough to
// earn its row. (Product-named notes files cannot generalize into a
// static table and stay a procedure's judgment.)
func TestFieldContextMdInCorpus(t *testing.T) {
	repo := t.TempDir()
	seedFile(t, repo, "CLAUDE.md", "# rules\n")
	seedFile(t, repo, "CONTEXT.md", "# how this repo works\n")
	r := inspectReport(t, repo)
	corpus := ""
	for _, c := range r.Instruction {
		corpus += c.Path + ";"
	}
	if !strings.Contains(corpus, "CONTEXT.md") {
		t.Fatalf("root CONTEXT.md must land in the instruction corpus [bench 4], got %v", r.Instruction)
	}
}

// [bench 1]: scripts named test:* are manifest-derived run commands
// the same way scripts.test is; only exact "test" was in the table.
func TestFieldScriptVariantCandidates(t *testing.T) {
	repo := t.TempDir()
	seedFile(t, repo, "package.json",
		`{"name":"ws","scripts":{"test:unit":"vitest run","test:e2e":"playwright test","lint":"eslint .","testing-util":"node x"}}`)
	r := inspectReport(t, repo)
	cmds := strings.Join(candidateCommands(r), ";")
	for _, want := range []string{"npm run test:unit", "npm run test:e2e"} {
		if !strings.Contains(cmds, want) {
			t.Fatalf("test:* scripts must derive candidates [bench 1], want %q in %v", want, r.Tests.Candidates)
		}
	}
	if strings.Contains(cmds, "lint") || strings.Contains(cmds, "testing-util") {
		t.Fatalf("only test / test:* scripts are test candidates, got %v", r.Tests.Candidates)
	}
}
