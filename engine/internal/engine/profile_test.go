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

// The minimal profile is the control condition (lifecycle §2.6): the
// floor composition, not a starter tier.
func TestMinimalProfileIsTheFloor(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	g := &engine.Engine{Repo: repo, Content: content.FS}
	e := envelope.New("init")
	g.Init(e, "claude-code", "minimal")
	if !e.OK || !diagCodes(e)["PROFILE_CONTROL"] {
		t.Fatalf("minimal init must state the control doctrine, got %+v", e.Diagnostics)
	}
	e = envelope.New("apply")
	g.Apply(e, true)
	if !e.OK {
		t.Fatalf("apply failed: %+v", e.Diagnostics)
	}

	b, _ := os.ReadFile(filepath.Join(repo, "CLAUDE.md"))
	if !strings.Contains(string(b), "Done means verified") {
		t.Fatal("the floor keeps verification")
	}
	for _, richer := range []string{"PROGRESS", "rein cite", "Stay in scope"} {
		if strings.Contains(string(b), richer) {
			t.Fatalf("the floor must not carry richer content (%q):\n%s", richer, b)
		}
	}
	if _, err := os.Stat(filepath.Join(repo, ".claude/skills")); err == nil {
		t.Fatal("the floor installs no skills")
	}
	if _, err := os.Stat(filepath.Join(repo, ".rein/state/PROGRESS.md")); err == nil {
		t.Fatal("the floor seeds no state files")
	}
	if _, err := os.Stat(filepath.Join(repo, "scripts/verify")); err != nil {
		t.Fatal("the floor keeps the gate")
	}
}

// Switching standard -> minimal removes richer content through the
// ordinary plan/apply path; agent-owned seeds are left behind
// (SEED_LEFT), never deleted.
func TestProfileSwitchLeavesSeeds(t *testing.T) {
	g, repo := appliedRepo(t)
	seedPath := filepath.Join(repo, ".rein/state/PROGRESS.md")
	if err := os.WriteFile(seedPath, []byte("# agent-owned history\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfgPath := filepath.Join(repo, "harness.yaml")
	cfg, _ := os.ReadFile(cfgPath)
	updated := strings.Replace(string(cfg), "adapter: claude-code\n", "adapter: claude-code\nprofile: minimal\n", 1)
	if err := os.WriteFile(cfgPath, []byte(updated), 0o644); err != nil {
		t.Fatal(err)
	}

	e := envelope.New("apply")
	g.Apply(e, true)
	if !e.OK {
		t.Fatalf("apply after switch failed: %+v", e.Diagnostics)
	}
	if _, err := os.Stat(filepath.Join(repo, ".claude/skills/rein-setup/SKILL.md")); err == nil {
		t.Fatal("switching to minimal must remove the skills")
	}
	b, err := os.ReadFile(seedPath)
	if err != nil || !strings.Contains(string(b), "agent-owned history") {
		t.Fatal("the switch must leave agent-owned seeds untouched")
	}
	cm, _ := os.ReadFile(filepath.Join(repo, "CLAUDE.md"))
	if strings.Contains(string(cm), "Stay in scope") {
		t.Fatal("the render must shrink to the floor")
	}

	d := envelope.New("doctor")
	g.Doctor(d)
	if !d.OK {
		t.Fatalf("doctor after switch must be clean, got %+v", d.Diagnostics)
	}
}

func TestUnknownProfileRefused(t *testing.T) {
	g, repo := newRepo(t, "claude-code")
	cfgPath := filepath.Join(repo, "harness.yaml")
	cfg, _ := os.ReadFile(cfgPath)
	updated := strings.Replace(string(cfg), "adapter: claude-code\n", "adapter: claude-code\nprofile: turbo\n", 1)
	os.WriteFile(cfgPath, []byte(updated), 0o644)
	e := envelope.New("plan")
	g.Plan(e)
	if e.OK || !diagCodes(e)["CONFIG_INVALID"] {
		t.Fatalf("unknown profile must refuse with the list, got %+v", e.Diagnostics)
	}
	if !strings.Contains(e.Diagnostics[0].Fix, "minimal, standard") {
		t.Fatalf("refusal must teach the profiles, got %q", e.Diagnostics[0].Fix)
	}
	e2 := envelope.New("init")
	g2 := &engine.Engine{Repo: t.TempDir(), Content: content.FS}
	g2.Init(e2, "claude-code", "turbo")
	if e2.OK || !diagCodes(e2)["INVALID_ARGUMENT"] {
		t.Fatalf("init must refuse an unknown profile, got %+v", e2.Diagnostics)
	}
}
