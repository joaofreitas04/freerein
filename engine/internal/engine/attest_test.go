package engine_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joaofreitas04/freerein/engine/internal/engine"
	"github.com/joaofreitas04/freerein/engine/internal/envelope"
)

// appliedRepo is newRepo + apply --yes: an installed harness whose
// gate is still the core stub until the test overrides it.
func appliedRepo(t *testing.T) (*engine.Engine, string) {
	t.Helper()
	g, repo := newRepo(t, "claude-code")
	e := envelope.New("apply")
	g.Apply(e, true)
	if !e.OK {
		t.Fatalf("apply failed: %+v", e.Diagnostics)
	}
	return g, repo
}

func overrideGate(t *testing.T, g *engine.Engine, repo, body string) {
	t.Helper()
	seedFile(t, repo, ".rein/overrides/scripts/verify", body)
	e := envelope.New("apply")
	g.Apply(e, true)
	if !e.OK {
		t.Fatalf("apply with override failed: %+v", e.Diagnostics)
	}
}

func TestAttestVocabularyAndMissingGate(t *testing.T) {
	g, _ := newRepo(t, "claude-code")
	e := envelope.New("attest")
	g.Attest(e, "vibes-are-good")
	if e.OK || !diagCodes(e)["INVALID_ARGUMENT"] {
		t.Fatalf("unknown subject must refuse with the vocabulary, got %+v", e.Diagnostics)
	}
	e = envelope.New("attest")
	g.Attest(e, "gate-can-fail") // init only: no scripts/verify installed
	if e.OK || !diagCodes(e)["GATE_MISSING"] {
		t.Fatalf("attesting an absent gate must refuse, got %+v", e.Diagnostics)
	}
}

// The standing check across its whole life: a real gate starts
// unproven, an attest makes doctor clean, a gate edit makes the proof
// stale, re-attesting clears it — and doctor never executes anything.
func TestGateProofLifecycle(t *testing.T) {
	g, repo := appliedRepo(t)
	overrideGate(t, g, repo, "#!/bin/sh\nexit 0\n")

	e := envelope.New("doctor")
	g.Doctor(e)
	if !diagCodes(e)["GATE_UNPROVEN"] || diagCodes(e)["GATE_STUB"] {
		t.Fatalf("real unattested gate: want GATE_UNPROVEN without GATE_STUB, got %+v", e.Diagnostics)
	}

	e = envelope.New("attest")
	g.Attest(e, "gate-can-fail")
	if !e.OK {
		t.Fatalf("attest failed: %+v", e.Diagnostics)
	}
	res := e.Result.(map[string]any)
	if !strings.HasPrefix(res["gate_hash"].(string), "sha256:") {
		t.Fatalf("attest must record the gate hash, got %v", res)
	}
	b, _ := os.ReadFile(filepath.Join(repo, engine.JournalName))
	if !strings.Contains(string(b), `"kind":"attest"`) || !strings.Contains(string(b), `"subject":"gate-can-fail"`) {
		t.Fatalf("attest must journal through the engine, got %s", b)
	}

	e = envelope.New("doctor")
	g.Doctor(e)
	if diagCodes(e)["GATE_UNPROVEN"] || diagCodes(e)["GATE_PROOF_STALE"] {
		t.Fatalf("attested gate must be clean, got %+v", e.Diagnostics)
	}

	// the gate changes: the proof no longer describes what runs
	overrideGate(t, g, repo, "#!/bin/sh\necho stricter\nexit 0\n")
	e = envelope.New("doctor")
	g.Doctor(e)
	if !diagCodes(e)["GATE_PROOF_STALE"] {
		t.Fatalf("edited gate must go stale, got %+v", e.Diagnostics)
	}

	e = envelope.New("attest")
	g.Attest(e, "gate-can-fail")
	if !e.OK {
		t.Fatal("re-attest failed")
	}
	e = envelope.New("doctor")
	g.Doctor(e)
	if diagCodes(e)["GATE_PROOF_STALE"] || diagCodes(e)["GATE_UNPROVEN"] {
		t.Fatalf("re-attested gate must be clean, got %+v", e.Diagnostics)
	}
}

// A stub cannot be proven: GATE_STUB owns that surface alone.
func TestGateProofSilentOnStub(t *testing.T) {
	g, _ := appliedRepo(t)
	e := envelope.New("doctor")
	g.Doctor(e)
	if !diagCodes(e)["GATE_STUB"] {
		t.Fatalf("clean install: want GATE_STUB, got %+v", e.Diagnostics)
	}
	if diagCodes(e)["GATE_UNPROVEN"] || diagCodes(e)["GATE_PROOF_STALE"] {
		t.Fatalf("stub gate must not stack proof findings, got %+v", e.Diagnostics)
	}
}
