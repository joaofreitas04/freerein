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

// TestWalkingSkeleton exercises the whole init → plan → apply → dump →
// doctor cycle against a temp repo, including the guarantees:
// idempotency, drift respected, overrides shadowing core.
func TestWalkingSkeleton(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	g := &engine.Engine{Repo: repo, Content: content.FS}

	run := func(name string, f func(*envelope.Envelope)) *envelope.Envelope {
		t.Helper()
		e := envelope.New(name)
		f(e)
		return e
	}
	codes := func(e *envelope.Envelope) []string {
		var out []string
		for _, d := range e.Diagnostics {
			out = append(out, d.Code)
		}
		return out
	}
	has := func(e *envelope.Envelope, code string) bool {
		for _, c := range codes(e) {
			if c == code {
				return true
			}
		}
		return false
	}

	// plan before init fails well-formed
	if e := run("plan", g.Plan); e.OK || !has(e, "NOT_INITIALIZED") {
		t.Fatalf("plan before init: want NOT_INITIALIZED, got ok=%v %v", e.OK, codes(e))
	}
	// init
	if e := run("init", func(e *envelope.Envelope) { g.Init(e, "claude-code") }); !e.OK {
		t.Fatalf("init failed: %v", codes(e))
	}
	// double init refused
	if e := run("init", func(e *envelope.Envelope) { g.Init(e, "claude-code") }); e.OK || !has(e, "ALREADY_INITIALIZED") {
		t.Fatalf("double init: want ALREADY_INITIALIZED, got ok=%v %v", e.OK, codes(e))
	}
	// apply without yes: confirm required, nothing written
	if e := run("apply", func(e *envelope.Envelope) { g.Apply(e, false) }); e.ConfirmRequired == nil {
		t.Fatal("apply without --yes must set confirm_required")
	}
	if _, err := os.Stat(filepath.Join(repo, "CLAUDE.md")); err == nil {
		t.Fatal("apply without --yes must not write files")
	}
	// apply --yes installs
	if e := run("apply", func(e *envelope.Envelope) { g.Apply(e, true) }); !e.OK {
		t.Fatalf("apply failed: %v", codes(e))
	}
	for _, f := range []string{"CLAUDE.md", "scripts/verify", ".rein/state/PROGRESS.md", "harness.lock"} {
		if _, err := os.Stat(filepath.Join(repo, f)); err != nil {
			t.Fatalf("apply must install %s: %v", f, err)
		}
	}
	if fi, _ := os.Stat(filepath.Join(repo, "scripts/verify")); fi.Mode()&0o111 == 0 {
		t.Fatal("scripts/verify must be executable")
	}
	// idempotency: second apply is a no-op
	if e := run("apply", func(e *envelope.Envelope) { g.Apply(e, true) }); !has(e, "UP_TO_DATE") {
		t.Fatalf("second apply: want UP_TO_DATE, got %v", codes(e))
	}
	// doctor clean
	if e := run("doctor", g.Doctor); !e.OK || len(e.Diagnostics) != 0 {
		t.Fatalf("doctor on clean install: want no findings, got %v", codes(e))
	}
	// drift: local edit is detected and never clobbered
	cm := filepath.Join(repo, "CLAUDE.md")
	orig, _ := os.ReadFile(cm)
	if err := os.WriteFile(cm, append(orig, []byte("\nLOCAL EDIT\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	if e := run("doctor", g.Doctor); !has(e, "DRIFT") {
		t.Fatalf("doctor after edit: want DRIFT, got %v", codes(e))
	}
	if e := run("apply", func(e *envelope.Envelope) { g.Apply(e, true) }); !has(e, "DRIFT_SKIPPED") {
		t.Fatalf("apply over drift: want DRIFT_SKIPPED, got %v", codes(e))
	}
	after, _ := os.ReadFile(cm)
	if !strings.Contains(string(after), "LOCAL EDIT") {
		t.Fatal("apply clobbered a locally edited file — guarantee violated")
	}
	// drift stays visible across applies (the lock keeps the installed hash)
	if e := run("doctor", g.Doctor); !has(e, "DRIFT") {
		t.Fatalf("drift must persist after apply, got %v", codes(e))
	}
	// overrides shadow core, shadowed provider recorded
	ov := filepath.Join(repo, ".rein/overrides/scripts")
	if err := os.MkdirAll(ov, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ov, "verify"), []byte("#!/bin/sh\necho ok\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if e := run("apply", func(e *envelope.Envelope) { g.Apply(e, true) }); !e.OK {
		t.Fatalf("apply with override failed: %v", codes(e))
	}
	installed, _ := os.ReadFile(filepath.Join(repo, "scripts/verify"))
	if !strings.Contains(string(installed), "echo ok") {
		t.Fatal("override did not win over core")
	}
}
