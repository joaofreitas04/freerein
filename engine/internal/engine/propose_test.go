package engine_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joaofreitas04/freerein/engine/internal/engine"
	"github.com/joaofreitas04/freerein/engine/internal/envelope"
)

func fullProposal() engine.ProposalFields {
	return engine.ProposalFields{
		Surface: "AGENTS.md.d/30-state.md", Change: "drop the assumptions section",
		Prediction:  "next session cold-starts as fast without it",
		Measurement: "paired-vs-minimal on the next three sessions",
		Baseline:    "profile: standard as composed today", NominatedBy: "cite-decay",
	}
}

func propose(t *testing.T, g *engine.Engine, f engine.ProposalFields) *envelope.Envelope {
	t.Helper()
	e := envelope.New("propose")
	g.Propose(e, f)
	return e
}

// Refuse-before-generate: a proposal with a hole is refused listing
// every missing field, and nominators are a registered vocabulary.
func TestProposeRefusesHoles(t *testing.T) {
	g, _ := newRepo(t, "claude-code")
	e := propose(t, g, engine.ProposalFields{Surface: "x"})
	if e.OK || !diagCodes(e)["MISSING_ARGUMENT"] {
		t.Fatalf("want refusal, got %+v", e.Diagnostics)
	}
	for _, f := range []string{"--change", "--prediction", "--measurement", "--baseline", "--nominated-by"} {
		if !strings.Contains(e.Diagnostics[0].Message, f) {
			t.Fatalf("refusal must list %s, got %q", f, e.Diagnostics[0].Message)
		}
	}
	bad := fullProposal()
	bad.NominatedBy = "vibes"
	e = propose(t, g, bad)
	if e.OK || !diagCodes(e)["INVALID_ARGUMENT"] {
		t.Fatalf("unknown nominator must refuse, got %+v", e.Diagnostics)
	}
}

// lifecycle §2.3 as a hard invariant: one open proposal at a time,
// visible in doctor until its verdict lands; §2.2: rejected verdicts
// are kept; §2.1 shape: the verdict carries the pre-named measurement.
func TestProposalLifecycle(t *testing.T) {
	g, repo := appliedRepo(t)

	e := propose(t, g, fullProposal())
	if !e.OK {
		t.Fatalf("propose failed: %+v", e.Diagnostics)
	}
	id := e.Result.(map[string]any)["id"].(string)
	if !strings.HasPrefix(id, "p-") {
		t.Fatalf("want p-<hash> id, got %q", id)
	}

	// second proposal refused, naming the open one
	e = propose(t, g, fullProposal())
	if e.OK || !diagCodes(e)["PROPOSAL_OPEN"] || !strings.Contains(e.Diagnostics[0].Message, id) {
		t.Fatalf("one change at a time, got %+v", e.Diagnostics)
	}

	// doctor keeps it visible
	d := envelope.New("doctor")
	g.Doctor(d)
	if !diagCodes(d)["PROPOSAL_OPEN"] {
		t.Fatalf("open proposal must be visible in doctor, got %+v", d.Diagnostics)
	}

	// verdict validation
	v := envelope.New("verdict")
	g.Verdict(v, id, "accepted", "")
	if v.OK || !diagCodes(v)["MISSING_ARGUMENT"] {
		t.Fatal("empty evidence must refuse")
	}
	v = envelope.New("verdict")
	g.Verdict(v, id, "maybe", "gate green")
	if v.OK || !diagCodes(v)["INVALID_ARGUMENT"] {
		t.Fatal("murky outcome must refuse")
	}
	v = envelope.New("verdict")
	g.Verdict(v, "p-00000000", "rejected", "gate red")
	if v.OK || !diagCodes(v)["NOT_FOUND"] {
		t.Fatal("unknown proposal must refuse")
	}

	// reject on evidence; the verdict carries the pre-named measurement
	v = envelope.New("verdict")
	g.Verdict(v, id, "rejected", "cold-start time unchanged; the section was being read")
	if !v.OK {
		t.Fatalf("verdict failed: %+v", v.Diagnostics)
	}
	b, _ := os.ReadFile(filepath.Join(repo, engine.JournalName))
	if !strings.Contains(string(b), `"measurement":"paired-vs-minimal on the next three sessions"`) {
		t.Fatalf("verdict must carry the pre-named measurement, got %s", b)
	}

	// verdicts are immutable
	v = envelope.New("verdict")
	g.Verdict(v, id, "accepted", "second thoughts")
	if v.OK || !diagCodes(v)["ALREADY_DECIDED"] {
		t.Fatal("a second verdict must refuse")
	}

	// the slot reopens after the verdict, and the rejection stays read-back
	e = propose(t, g, fullProposal())
	if !e.OK {
		t.Fatalf("verdicted slot must reopen, got %+v", e.Diagnostics)
	}
	je := envelope.New("journal")
	g.Journal(je, engine.JournalOpts{Kind: "verdict", Limit: 20})
	r := je.Result.(map[string]any)
	if r["total"].(int) != 1 {
		t.Fatalf("rejected verdicts are kept and readable, got %v", r["total"])
	}
	d = envelope.New("doctor")
	g.Doctor(d)
	if !diagCodes(d)["PROPOSAL_OPEN"] {
		t.Fatal("the new open proposal must be visible")
	}
}
