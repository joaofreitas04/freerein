package engine_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joaofreitas04/freerein/content"
	"github.com/joaofreitas04/freerein/engine/internal/engine"
	"github.com/joaofreitas04/freerein/engine/internal/envelope"
)

func newRepo(t *testing.T, adapter string) (*engine.Engine, string) {
	t.Helper()
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	g := &engine.Engine{Repo: repo, Content: content.FS}
	e := envelope.New("init")
	g.Init(e, adapter)
	if !e.OK {
		t.Fatalf("init failed: %+v", e.Diagnostics)
	}
	return g, repo
}

func diagCodes(e *envelope.Envelope) map[string]bool {
	m := map[string]bool{}
	for _, d := range e.Diagnostics {
		m[d.Code] = true
	}
	return m
}

func apply(t *testing.T, g *engine.Engine) *envelope.Envelope {
	t.Helper()
	e := envelope.New("apply")
	g.Apply(e, true)
	if !e.OK {
		t.Fatalf("apply failed: %+v", e.Diagnostics)
	}
	return e
}

// The codex adapter renders AGENTS.md, maps skills to .agents/skills,
// and its degradations are declared, not silent.
func TestCodexAdapter(t *testing.T) {
	g, repo := newRepo(t, "codex")
	e := apply(t, g)
	codes := diagCodes(e)
	if !codes["HOST_NO_COMPACTION_HOOK"] || !codes["HOST_DEGRADATION"] {
		t.Fatalf("codex degradations must be declared, got %+v", e.Diagnostics)
	}
	if _, err := os.Stat(filepath.Join(repo, "AGENTS.md")); err != nil {
		t.Fatal("codex must render the instruction file as AGENTS.md")
	}
	if _, err := os.Stat(filepath.Join(repo, "CLAUDE.md")); err == nil {
		t.Fatal("codex must not render CLAUDE.md")
	}
	if _, err := os.Stat(filepath.Join(repo, ".agents/skills/rein-setup/SKILL.md")); err != nil {
		t.Fatal("skills/ must map to the codex skills dir")
	}
}

// The claude-code adapter maps skills/ to .claude/skills.
func TestSkillMapping(t *testing.T) {
	g, repo := newRepo(t, "claude-code")
	apply(t, g)
	if _, err := os.Stat(filepath.Join(repo, ".claude/skills/rein-setup/SKILL.md")); err != nil {
		t.Fatal("skills/ must map to .claude/skills on claude-code")
	}
}

func writeComponent(t *testing.T, dir, manifest string, files map[string]string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "component.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	for rel, content := range files {
		full := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

const extManifest = `name: toy-ext
kind: extension
version: 0.1.0
subsystem: instructions
rung: instruction
rent:
  class: amplifier
provides:
  - AGENTS.md.d/50-toy.md
description: a toy extension
`

func declareExtension(t *testing.T, repo, src string) {
	t.Helper()
	cfg := "adapter: claude-code\npresets: []\nextensions: [\"" + src + "\"]\n"
	if err := os.WriteFile(filepath.Join(repo, "harness.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
}

// A local-path extension joins the layer stack and its fragment lands
// in the rendered instruction file; a preset that adds a new path is
// refused.
func TestLocalExtensionAndPresetRule(t *testing.T) {
	g, repo := newRepo(t, "claude-code")
	writeComponent(t, filepath.Join(repo, "vendor/toy-ext"), extManifest,
		map[string]string{"AGENTS.md.d/50-toy.md": "## Toy rule\n\nAlways be toy.\n"})
	declareExtension(t, repo, "./vendor/toy-ext")
	apply(t, g)
	b, _ := os.ReadFile(filepath.Join(repo, "CLAUDE.md"))
	if !strings.Contains(string(b), "Always be toy.") {
		t.Fatal("extension fragment must render into the instruction file")
	}
	// preset adding a new path must be refused at resolve time
	presetManifest := strings.NewReplacer("toy-ext", "toy-preset", "kind: extension", "kind: preset",
		"50-toy.md", "99-new.md").Replace(extManifest)
	writeComponent(t, filepath.Join(repo, "vendor/toy-preset"), presetManifest,
		map[string]string{"AGENTS.md.d/99-new.md": "new capability\n"})
	cfg := "adapter: claude-code\npresets: [\"./vendor/toy-preset\"]\nextensions: [\"./vendor/toy-ext\"]\n"
	if err := os.WriteFile(filepath.Join(repo, "harness.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	e := envelope.New("plan")
	g.Plan(e)
	if e.OK || !diagCodes(e)["RESOLVE_FAILED"] {
		t.Fatalf("preset adding a path must fail resolution, got %+v", e.Diagnostics)
	}
	// a kind mismatch is its own error
	cfg = "adapter: claude-code\npresets: [\"./vendor/toy-ext\"]\nextensions: []\n"
	_ = os.WriteFile(filepath.Join(repo, "harness.yaml"), []byte(cfg), 0o644)
	e = envelope.New("plan")
	g.Plan(e)
	if e.OK || !diagCodes(e)["KIND_MISMATCH"] {
		t.Fatalf("extension declared as preset must fail, got %+v", e.Diagnostics)
	}
}

// Local edit + upstream change on the same file: clean edits merge
// (both survive), overlapping edits conflict and leave the file alone.
func TestThreeWayMerge(t *testing.T) {
	g, repo := newRepo(t, "claude-code")
	writeComponent(t, filepath.Join(repo, "vendor/toy-ext"), extManifest,
		map[string]string{"AGENTS.md.d/50-toy.md": "## Toy rule\n\nAlways be toy.\n"})
	declareExtension(t, repo, "./vendor/toy-ext")
	apply(t, g)

	// local edit at the TOP of CLAUDE.md — disjoint from the
	// extension's fragment, which sorts last
	cm := filepath.Join(repo, "CLAUDE.md")
	orig, _ := os.ReadFile(cm)
	_ = os.WriteFile(cm, append([]byte("## Local addendum\n\nkeep me\n\n"), orig...), 0o644)

	// upstream change at the END
	_ = os.WriteFile(filepath.Join(repo, "vendor/toy-ext/AGENTS.md.d/50-toy.md"),
		[]byte("## Toy rule v2\n\nAlways be very toy.\n"), 0o644)

	e := apply(t, g)
	if !diagCodes(e)["MERGED"] {
		t.Fatalf("disjoint edits must merge cleanly, got %+v", e.Diagnostics)
	}
	after, _ := os.ReadFile(cm)
	s := string(after)
	if !strings.Contains(s, "keep me") || !strings.Contains(s, "very toy") {
		t.Fatalf("merge must keep both sides, got:\n%s", s)
	}
	// drift must still be visible (the local addendum remains)
	ed := envelope.New("doctor")
	g.Doctor(ed)
	if !diagCodes(ed)["DRIFT"] {
		t.Fatalf("merged file with local edits must still read as drift, got %+v", ed.Diagnostics)
	}

	// overlapping edit: local rewrite of the same fragment region
	before, _ := os.ReadFile(cm)
	_ = os.WriteFile(cm, []byte(strings.Replace(string(before), "Always be very toy.", "Always be local.", 1)), 0o644)
	_ = os.WriteFile(filepath.Join(repo, "vendor/toy-ext/AGENTS.md.d/50-toy.md"),
		[]byte("## Toy rule v3\n\nAlways be upstream.\n"), 0o644)
	e = envelope.New("apply")
	g.Apply(e, true)
	if !diagCodes(e)["MERGE_CONFLICT"] {
		t.Fatalf("overlapping edits must conflict, got %+v", e.Diagnostics)
	}
	final, _ := os.ReadFile(cm)
	if !strings.Contains(string(final), "Always be local.") {
		t.Fatal("conflicting file must be left alone")
	}
	artifact := filepath.Join(repo, ".rein/out/merge/CLAUDE.md")
	b, err := os.ReadFile(artifact)
	if err != nil || !strings.Contains(string(b), "<<<<<<<") {
		t.Fatal("conflict artifact with markers must be written under .rein/out/merge/")
	}
}
