package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/joaofreitas04/freerein/engine/internal/envelope"
)

// Evolve mechanics (spec/journal.md, docs/lifecycle.md §2): the
// engine records proposals and verdicts and enforces what is
// mechanically enforceable — complete fields, one open proposal at a
// time, a verdict that carries evidence and the pre-named
// measurement. Who runs the measurement is procedure discipline; the
// engine does not pretend to verify it.

// ProposalFields are the six mandatory fields of a proposal
// (refuse-before-generate, the rein-decide precedent). The
// measurement is named BEFORE the change is made: an acceptance
// measured by something chosen afterward is the proposer accepting.
type ProposalFields struct {
	Surface     string // the artifact or component being changed
	Change      string // one line: what changes
	Prediction  string // falsifiable: next time X, Y instead of Z
	Measurement string // what will decide: gate re-run / paired-vs-minimal / held-out case
	Baseline    string // the condition compared against
	NominatedBy string // cite-decay | compensation-recheck | overlap | recurrence | human
}

var nominators = map[string]bool{
	"cite-decay": true, "compensation-recheck": true, "overlap": true,
	"recurrence": true, "human": true,
}

var outcomes = map[string]bool{"accepted": true, "rejected": true}

type openProposal struct {
	ID          string
	Surface     string
	Measurement string
}

// journalProposals scans the journal for proposals and the ids their
// verdicts closed. Malformed lines are skipped — `rein journal` owns
// reporting those.
func (g *Engine) journalProposals() (proposals []openProposal, closed map[string]bool) {
	closed = map[string]bool{}
	raw, err := os.ReadFile(filepath.Join(g.Repo, JournalName))
	if err != nil {
		return nil, closed
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var entry struct {
			Kind        string `json:"kind"`
			ID          string `json:"id"`
			Proposal    string `json:"proposal"`
			Surface     string `json:"surface"`
			Measurement string `json:"measurement"`
		}
		if json.Unmarshal([]byte(line), &entry) != nil {
			continue
		}
		switch entry.Kind {
		case "proposal":
			proposals = append(proposals, openProposal{ID: entry.ID, Surface: entry.Surface, Measurement: entry.Measurement})
		case "verdict":
			closed[entry.Proposal] = true
		}
	}
	return proposals, closed
}

func (g *Engine) openProposals() []openProposal {
	proposals, closed := g.journalProposals()
	var open []openProposal
	for _, p := range proposals {
		if !closed[p.ID] {
			open = append(open, p)
		}
	}
	return open
}

// Propose opens an evolve proposal (lifecycle §2.3: one change at a
// time, with a falsifiable prediction). Refuses on any missing field,
// listing all of them; refuses while another proposal is open, no
// force flag — the doctrine is one at a time, not one at a time
// unless inconvenient.
func (g *Engine) Propose(e *envelope.Envelope, f ProposalFields) {
	var missing []string
	for _, req := range []struct{ name, v string }{
		{"--surface", f.Surface}, {"--change", f.Change},
		{"--prediction", f.Prediction}, {"--measurement", f.Measurement},
		{"--baseline", f.Baseline}, {"--nominated-by", f.NominatedBy},
	} {
		if strings.TrimSpace(req.v) == "" {
			missing = append(missing, req.name)
		}
	}
	if len(missing) > 0 {
		e.Fail("MISSING_ARGUMENT", "a proposal is refused with a hole: "+strings.Join(missing, ", ")+" missing",
			"every field is mandatory — surface, change, prediction (next time X, Y instead of Z), "+
				"measurement (named before the change: gate re-run / paired-vs-minimal / held-out case), "+
				"baseline, nominated-by")
		return
	}
	if !nominators[f.NominatedBy] {
		known := make([]string, 0, len(nominators))
		for n := range nominators {
			known = append(known, n)
		}
		sort.Strings(known)
		e.Fail("INVALID_ARGUMENT", "unknown nominator "+f.NominatedBy,
			"--nominated-by is one of: "+strings.Join(known, ", ")+" — evolve candidates come from the observe substrates or a human")
		return
	}
	if open := g.openProposals(); len(open) > 0 {
		e.Fail("PROPOSAL_OPEN", "proposal "+open[0].ID+" ("+open[0].Surface+") has no verdict yet — one change at a time",
			"run its pre-named measurement ("+open[0].Measurement+") and record `rein verdict --proposal "+open[0].ID+
				" --outcome accepted|rejected --evidence \"…\"` before proposing again")
		return
	}
	at := time.Now().UTC().Format(time.RFC3339)
	// the sequence number salts the hash: re-proposing identical
	// fields (legitimate after a rejection, with new evidence) must
	// mint a new id, not resurrect the old verdict
	prior, _ := g.journalProposals()
	sum := sha256.Sum256([]byte(f.Surface + "|" + f.Change + "|" + f.Prediction + "|" +
		f.Measurement + "|" + f.Baseline + "|" + f.NominatedBy + "|" + at +
		"|" + strconv.Itoa(len(prior))))
	id := "p-" + hex.EncodeToString(sum[:])[:8]
	if err := g.appendJournal("proposal", map[string]any{
		"at": at, "id": id, "surface": f.Surface, "change": f.Change,
		"prediction": f.Prediction, "measurement": f.Measurement,
		"baseline": f.Baseline, "nominated_by": f.NominatedBy,
	}); err != nil {
		e.Fail("JOURNAL_WRITE_FAILED", JournalName+": "+err.Error(), "check permissions on .rein/ and re-run")
		return
	}
	e.Result = map[string]any{"id": id, "recorded": JournalName,
		"next": "make the change, run the measurement (" + f.Measurement + "), then `rein verdict --proposal " + id + "`"}
}

// Verdict closes a proposal on a measurement's result (lifecycle
// §2.1: evidence, never the argument). The engine copies the
// proposal's pre-named measurement into the verdict entry, so what
// decided and what was promised to decide cannot disagree. Rejected
// verdicts are kept forever — that is the journal's job.
func (g *Engine) Verdict(e *envelope.Envelope, proposalID, outcome, evidence string) {
	if proposalID == "" || outcome == "" || strings.TrimSpace(evidence) == "" {
		e.Fail("MISSING_ARGUMENT", "a verdict needs --proposal, --outcome, and --evidence",
			"the evidence is the measurement's result — what the gate/paired check/held-out case showed, never the argument for the change")
		return
	}
	if !outcomes[outcome] {
		e.Fail("INVALID_ARGUMENT", "unknown outcome "+outcome,
			"--outcome is accepted or rejected; anything murkier goes back to measurement or to the human")
		return
	}
	proposals, closed := g.journalProposals()
	var target *openProposal
	for i := range proposals {
		if proposals[i].ID == proposalID {
			target = &proposals[i]
		}
	}
	if target == nil {
		e.Fail("NOT_FOUND", "no proposal "+proposalID+" in the journal",
			"`rein journal --kind proposal` lists them")
		return
	}
	if closed[proposalID] {
		e.Fail("ALREADY_DECIDED", proposalID+" already has a verdict — verdicts are immutable",
			"a wrong verdict is corrected by a new proposal, never by a second verdict (the journal is append-only)")
		return
	}
	if err := g.appendJournal("verdict", map[string]any{
		"proposal": proposalID, "outcome": outcome, "evidence": evidence,
		"measurement": target.Measurement,
	}); err != nil {
		e.Fail("JOURNAL_WRITE_FAILED", JournalName+": "+err.Error(), "check permissions on .rein/ and re-run")
		return
	}
	result := map[string]any{"proposal": proposalID, "outcome": outcome, "recorded": JournalName}
	if outcome == "rejected" {
		result["next"] = "revert the change; the verdict stays — it is what stops this proposal returning next month"
	}
	e.Result = result
}
