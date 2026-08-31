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

// spec/field-report.md producer side: rein report assembles the
// component-scoped shape from the lock, the journal, and the citation
// stats; the judgment fields arrive as flags (refuse-before-generate);
// the report lands under .rein/out/ and nothing leaves — `off` is the
// only implemented submission level.

func TestReportAssemblesShape(t *testing.T) {
	g, repo := newRepo(t, "claude-code")
	apply(t, g)
	// a standing local override displacing the component's file
	seedFile(t, repo, ".rein/overrides/scripts/verify", "#!/bin/sh\necho custom gate\n")
	apply(t, g)
	e := envelope.New("note")
	g.Note(e, "diagnose: verification-gate check misfired on a scratch tree")
	if !e.OK {
		t.Fatalf("note failed: %+v", e.Diagnostics)
	}
	for _, id := range []string{"verification-gate:20-verification", "instructions-base:00-base"} {
		e = envelope.New("cite")
		g.Cite(e, id)
		if !e.OK {
			t.Fatalf("cite %s failed: %+v", id, e.Diagnostics)
		}
	}

	e = envelope.New("report")
	g.Report(e, "verification-gate", engine.ReportFields{
		Subsystem:    "feedback",
		FailureClass: "gate check misfires when the target file is outside its scope",
		Reproduction: "repo with a compile check bounded to an import graph; breakage in an unimported file; expected red, observed green",
		Disposition:  "fix",
	})
	if !e.OK {
		t.Fatalf("report failed: %+v", e.Diagnostics)
	}
	if !diagCodes(e)["OUTPUT_OFFLOADED"] {
		t.Fatalf("report must announce its offload, got %+v", e.Diagnostics)
	}
	b, err := os.ReadFile(filepath.Join(repo, ".rein/out/report/verification-gate.json"))
	if err != nil {
		t.Fatalf("report file must be written: %v", err)
	}
	var r engine.FieldReport
	if err := json.Unmarshal(b, &r); err != nil {
		t.Fatalf("report is not valid JSON: %v", err)
	}
	if r.FormatVersion != engine.FieldReportFormatVersion || r.Engine == "" {
		t.Fatalf("report must carry formatVersion and engine, got %q %q", r.FormatVersion, r.Engine)
	}
	if r.Component.ID != "verification-gate" || r.Component.Version == "" {
		t.Fatalf("component id+version must come from the lock, got %+v", r.Component)
	}
	if r.Subsystem != "feedback" || r.DispositionRequested != "fix" ||
		r.FailureClass == "" || r.Reproduction == "" {
		t.Fatalf("judgment fields must be copied in, got %+v", r)
	}
	if r.Fingerprint == "" {
		t.Fatal("the fingerprint is part of the enumerated outbound surface")
	}
	// shadows: the standing override on the component's file, with a
	// since-date read from the journal
	foundShadow := false
	for _, s := range r.Evidence.Shadows {
		if s.Path == "scripts/verify" {
			foundShadow = true
			if s.Since == "" {
				t.Fatalf("a shadow carries its since-date from the journal, got %+v", s)
			}
		}
	}
	if !foundShadow {
		t.Fatalf("the standing override must appear in evidence.shadows, got %+v", r.Evidence.Shadows)
	}
	// citations: this component's counters verbatim — and only this
	// component's (privacy by construction is scoping, not scrubbing)
	if _, ok := r.Evidence.Citations["verification-gate:20-verification"]; !ok {
		t.Fatalf("the component's citation counters must be included, got %+v", r.Evidence.Citations)
	}
	for id := range r.Evidence.Citations {
		if !strings.HasPrefix(id, "verification-gate:") {
			t.Fatalf("another component's counters leaked into the report: %s", id)
		}
	}
	// journal evidence: judgment kinds mentioning the component
	foundNote := false
	for _, j := range r.Evidence.Journal {
		if j.Kind == "note" && strings.Contains(j.Text, "misfired") {
			foundNote = true
		}
	}
	if !foundNote {
		t.Fatalf("the component-scoped note must appear in evidence.journal, got %+v", r.Evidence.Journal)
	}
}

func TestReportRefusals(t *testing.T) {
	g, _ := newRepo(t, "claude-code")
	apply(t, g)

	e := envelope.New("report")
	g.Report(e, "verification-gate", engine.ReportFields{})
	if e.OK || !diagCodes(e)["MISSING_ARGUMENT"] {
		t.Fatalf("empty judgment fields must be refused, got ok=%v %+v", e.OK, e.Diagnostics)
	}

	full := engine.ReportFields{Subsystem: "feedback", FailureClass: "x", Reproduction: "y", Disposition: "fix"}

	bad := full
	bad.Subsystem = "vibes"
	e = envelope.New("report")
	g.Report(e, "verification-gate", bad)
	if e.OK || !diagCodes(e)["INVALID_ARGUMENT"] {
		t.Fatalf("unknown subsystem must be refused, got ok=%v %+v", e.OK, e.Diagnostics)
	}

	bad = full
	bad.Disposition = "yell"
	e = envelope.New("report")
	g.Report(e, "verification-gate", bad)
	if e.OK || !diagCodes(e)["INVALID_ARGUMENT"] {
		t.Fatalf("unknown disposition must be refused, got ok=%v %+v", e.OK, e.Diagnostics)
	}

	e = envelope.New("report")
	g.Report(e, "no-such-component", full)
	if e.OK || !diagCodes(e)["NOT_FOUND"] {
		t.Fatalf("a component absent from the lock must be refused, got ok=%v %+v", e.OK, e.Diagnostics)
	}
}

func TestReportNeedsLock(t *testing.T) {
	g, _ := newRepo(t, "claude-code") // init only, never applied — no lock
	e := envelope.New("report")
	g.Report(e, "verification-gate", engine.ReportFields{Subsystem: "feedback", FailureClass: "x", Reproduction: "y", Disposition: "fix"})
	if e.OK || !diagCodes(e)["NOT_INITIALIZED"] {
		t.Fatalf("report without an installed harness must refuse, got ok=%v %+v", e.OK, e.Diagnostics)
	}
}
