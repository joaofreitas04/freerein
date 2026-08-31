package engine_test

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

func seedFile(t *testing.T, repo, rel, content string) {
	t.Helper()
	full := filepath.Join(repo, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// inspect surveys the tree without executing anything, works before
// init, and offloads the full report (spec/inspection.md).
func TestInspectFamilies(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	seedFile(t, repo, "go.mod", "module toy\n")
	seedFile(t, repo, "package.json", `{"scripts":{"test":"jest"}}`)
	seedFile(t, repo, "pytest.ini", "[pytest]\n")
	seedFile(t, repo, "requirements.txt", "pytest\n")
	seedFile(t, repo, ".github/workflows/ci.yml", "on: push\n")
	seedFile(t, repo, ".eslintrc.json", "{}\n")
	seedFile(t, repo, "CLAUDE.md", "# rules\n")
	seedFile(t, repo, "sub/AGENTS.md", "# nested\n")
	seedFile(t, repo, "node_modules/dep/AGENTS.md", "# must be skipped\n")
	seedFile(t, repo, "docs/design.md", "# doc\n")
	seedFile(t, repo, ".mcp.json", "{}\n")

	g := &engine.Engine{Repo: repo, Content: content.FS}
	e := envelope.New("inspect")
	g.Inspect(e)
	if !e.OK {
		t.Fatalf("inspect failed: %+v", e.Diagnostics)
	}
	if !diagCodes(e)["OUTPUT_OFFLOADED"] {
		t.Fatalf("inspect must announce its offload, got %+v", e.Diagnostics)
	}
	b, err := os.ReadFile(filepath.Join(repo, ".rein/out/inspect.json"))
	if err != nil {
		t.Fatalf("inspect must write the report: %v", err)
	}
	var r engine.InspectReport
	if err := json.Unmarshal(b, &r); err != nil {
		t.Fatalf("report is not valid JSON: %v", err)
	}
	langs := strings.Join(r.Toolchain.Languages, ",")
	for _, want := range []string{"go", "node", "python"} {
		if !strings.Contains(langs, want) {
			t.Fatalf("toolchain must detect %s, got %v", want, r.Toolchain.Languages)
		}
	}
	var cands []string
	for _, c := range r.Tests.Candidates {
		cands = append(cands, c.Command)
	}
	joined := strings.Join(cands, ";")
	for _, want := range []string{"go test ./...", "npm test", "pytest"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("candidates must include %q, got %v", want, cands)
		}
	}
	if len(r.CI) != 1 || r.CI[0] != ".github/workflows/ci.yml" {
		t.Fatalf("ci detection wrong: %v", r.CI)
	}
	if len(r.LintFormat) == 0 {
		t.Fatal("lint config must be detected")
	}
	corpus := ""
	for _, c := range r.Instruction {
		corpus += c.Path + ";"
	}
	if !strings.Contains(corpus, "CLAUDE.md") || !strings.Contains(corpus, "sub/AGENTS.md") {
		t.Fatalf("instruction corpus must include root and nested files, got %v", r.Instruction)
	}
	if strings.Contains(corpus, "node_modules") {
		t.Fatalf("the walk must skip dependency dirs, got %v", r.Instruction)
	}
	if len(r.ConfigSurfaces) == 0 || r.DocsTree != "docs/" {
		t.Fatalf("config surfaces / docs tree wrong: %v %q", r.ConfigSurfaces, r.DocsTree)
	}
	for _, aff := range []string{"test-runner", "ci", "linter", "docs-tree"} {
		if !r.Affordances[aff] {
			t.Fatalf("affordance %s must be true, got %v", aff, r.Affordances)
		}
	}
	if len(r.Notes) == 0 {
		t.Fatal("an unreadable git history must be noted, not silent")
	}
}

// requires-gating and inspect share detection: a component requiring
// test-runner fails plan in a bare repo and passes once tests exist.
func TestProbeGating(t *testing.T) {
	g, repo := newRepo(t, "claude-code")
	manifest := `name: needs-tests
kind: extension
version: 0.1.0
subsystem: feedback
rung: instruction
rent:
  class: amplifier
provides:
  - AGENTS.md.d/70-tests.md
requires:
  - test-runner
description: an extension gated on a test runner
`
	writeComponent(t, filepath.Join(repo, "vendor/needs-tests"), manifest,
		map[string]string{"AGENTS.md.d/70-tests.md": "## Tests\n\nrun them\n"})
	declareExtension(t, repo, "./vendor/needs-tests")
	e := envelope.New("plan")
	g.Plan(e)
	if e.OK || !diagCodes(e)["REQUIREMENT_UNMET"] {
		t.Fatalf("plan must refuse an unmet requirement, got ok=%v %+v", e.OK, e.Diagnostics)
	}
	seedFile(t, repo, "pytest.ini", "[pytest]\n")
	seedFile(t, repo, "requirements.txt", "pytest\n")
	e = envelope.New("plan")
	g.Plan(e)
	if !e.OK {
		t.Fatalf("plan must pass once the affordance exists: %+v", e.Diagnostics)
	}
}

// doctor audits the debt ledger: rows need evidence + trigger, and a
// dated trigger that has passed re-activates the entry.
func TestDebtDoctorChecks(t *testing.T) {
	g, repo := newRepo(t, "claude-code")
	apply(t, g)
	debt := `# Debt ledger

| Debt | Evidence | Paydown trigger |
|---|---|---|
| flaky auth test | | someday |
| old lint errors | 14 errors in pkg/x | re-check 2020-01-01 |
| healthy row | 2 skipped tests in y | when y ships v2 |
`
	seedFile(t, repo, ".rein/state/DEBT.md", debt)
	e := envelope.New("doctor")
	g.Doctor(e)
	codes := diagCodes(e)
	if !codes["DEBT_ROW_INCOMPLETE"] || !codes["DEBT_EXPIRED"] {
		t.Fatalf("doctor must flag incomplete and expired debt, got %+v", e.Diagnostics)
	}
	incomplete, expired := 0, 0
	for _, d := range e.Diagnostics {
		switch d.Code {
		case "DEBT_ROW_INCOMPLETE":
			incomplete++
		case "DEBT_EXPIRED":
			expired++
		}
	}
	if incomplete != 1 || expired != 1 {
		t.Fatalf("healthy rows must not be flagged: %d incomplete, %d expired", incomplete, expired)
	}
}

// rein note appends a journal entry; empty text is refused.
func TestNoteJournals(t *testing.T) {
	g, repo := newRepo(t, "claude-code")
	e := envelope.New("note")
	g.Note(e, "setup: demoted 3 rules; rejected e2e suite as gate (14m runtime)")
	if !e.OK {
		t.Fatalf("note failed: %+v", e.Diagnostics)
	}
	b, err := os.ReadFile(filepath.Join(repo, ".rein/journal.jsonl"))
	if err != nil || !strings.Contains(string(b), `"kind":"note"`) || !strings.Contains(string(b), "demoted 3 rules") {
		t.Fatalf("note must land in the journal: %v %s", err, b)
	}
	e = envelope.New("note")
	g.Note(e, "   ")
	if e.OK || !diagCodes(e)["MISSING_ARGUMENT"] {
		t.Fatalf("empty note must be refused, got ok=%v %+v", e.OK, e.Diagnostics)
	}
}

// composition counts what it can and classifies what it cannot —
// never a silent skip (spec/inspection.md rules 2-4).
func TestInspectComposition(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	goSrc := "package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(1) }\n"
	seedFile(t, repo, "main.go", goSrc)
	seedFile(t, repo, "copy.go", goSrc) // exact duplicate
	seedFile(t, repo, "pic.png", "\x89PNG\x00\x00binary")
	seedFile(t, repo, "empty.txt", "")
	seedFile(t, repo, "app.min.js", "var a=1;\n")                                         // generated by name
	seedFile(t, repo, "gen.go", "// Code generated by toygen. DO NOT EDIT.\npackage x\n") // by marker
	seedFile(t, repo, "notes.xyz", "a\nb\nc\n")
	seedFile(t, repo, "tool", "#!/bin/sh\necho hi\n")
	seedFile(t, repo, "node_modules/dep/big.js", "must not be visited\n")

	g := &engine.Engine{Repo: repo, Content: content.FS}
	e := envelope.New("inspect")
	g.Inspect(e)
	if !e.OK {
		t.Fatalf("inspect failed: %+v", e.Diagnostics)
	}
	b, _ := os.ReadFile(filepath.Join(repo, ".rein/out/inspect.json"))
	var r engine.InspectReport
	if err := json.Unmarshal(b, &r); err != nil {
		t.Fatal(err)
	}
	if r.FormatVersion != engine.InspectFormatVersion || r.Engine == "" {
		t.Fatalf("report must carry its format version and the engine that wrote it, got %q / %q", r.FormatVersion, r.Engine)
	}
	m := r.Measure
	want := map[string]int{"analyzed": 2, "binary": 1, "duplicate": 1, "empty": 1, "generated": 2, "oversize": 0, "unknown": 1, "error": 0}
	for k, v := range want {
		if m.States[k] != v {
			t.Fatalf("state %s: want %d got %d (%v)", k, v, m.States[k], m.States)
		}
	}
	if !strings.Contains(m.Method, "newline") {
		t.Fatal("the counting method must travel with the numbers")
	}
	if len(m.Languages) != 2 || m.Languages[0].Language != "Go" ||
		m.Languages[0].Files != 1 || m.Languages[0].Lines != 5 || m.Languages[0].NonBlank != 3 {
		t.Fatalf("Go must top the languages (1 file, 5 lines, 3 non-blank), got %+v", m.Languages)
	}
	if m.Languages[1].Language != "Shell" || m.Languages[1].Lines != 2 {
		t.Fatalf("an extensionless script must map by shebang, got %+v", m.Languages)
	}
	if m.TotalFiles != 8 || m.TotalLines != 7 {
		t.Fatalf("totals wrong (node_modules must be skipped, only analyzed files count lines): files=%d lines=%d", m.TotalFiles, m.TotalLines)
	}
	sawUnknown := false
	for _, c := range m.Classified {
		if c.Path == "notes.xyz" && c.State == "unknown" {
			sawUnknown = true
		}
	}
	if !sawUnknown {
		t.Fatalf("the classified list must make non-analyzed files inspectable, got %+v", m.Classified)
	}
}

// an unrecognized file and an unreadable file are different facts —
// the latter is state "error", never conflated into "unknown".
func TestMeasureErrorState(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission bits do not bind root")
	}
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	seedFile(t, repo, "sealed.go", "package sealed\n")
	if err := os.Chmod(filepath.Join(repo, "sealed.go"), 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(repo, "sealed.go"), 0o644) })
	g := &engine.Engine{Repo: repo, Content: content.FS}
	e := envelope.New("inspect")
	g.Inspect(e)
	if !e.OK {
		t.Fatalf("inspect failed: %+v", e.Diagnostics)
	}
	b, _ := os.ReadFile(filepath.Join(repo, ".rein/out/inspect.json"))
	var r engine.InspectReport
	if err := json.Unmarshal(b, &r); err != nil {
		t.Fatal(err)
	}
	if r.Measure.States["error"] != 1 || r.Measure.States["unknown"] != 0 {
		t.Fatalf("unreadable file must be state error, got %v", r.Measure.States)
	}
}
