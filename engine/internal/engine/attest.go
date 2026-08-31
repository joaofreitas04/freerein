package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/joaofreitas04/freerein/engine/internal/envelope"
)

// GatePath is the installed verification gate. Attest and doctor
// share it (and the hash function below) so the proof and the audit
// cannot disagree about what was proven.
const GatePath = "scripts/verify"

// attestSubjects is the attest vocabulary. One entry today; grows
// like the probe vocabulary — registered, never free-form, so doctor
// can branch on subjects mechanically.
var attestSubjects = map[string]bool{"gate-can-fail": true}

// gateHash is the identity of the gate as it runs: sha256 of the
// installed file's bytes.
func (g *Engine) gateHash() (string, error) {
	b, err := os.ReadFile(filepath.Join(g.Repo, GatePath))
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// Attest records that a judgment procedure proved something the
// engine cannot prove itself. For gate-can-fail: the skill introduced
// a breakage, watched the gate exit non-zero, reverted — judgment and
// side effects, deliberately unmechanized (an engine version would
// execute arbitrary project code from doctor, or prove the wrong
// thing). The engine's share is mechanical: record the fact in the
// journal with the gate's current hash, so doctor can later say
// whether the proof still describes the gate that exists.
func (g *Engine) Attest(e *envelope.Envelope, subject string) {
	if !attestSubjects[subject] {
		known := make([]string, 0, len(attestSubjects))
		for s := range attestSubjects {
			known = append(known, s)
		}
		e.Fail("INVALID_ARGUMENT", "unknown attest subject "+subject,
			"the attest vocabulary is: "+strings.Join(known, ", "))
		return
	}
	hash, err := g.gateHash()
	if err != nil {
		e.Fail("GATE_MISSING", GatePath+": "+err.Error()+" — nothing to attest",
			"install the gate first (`rein apply --yes`), prove it can fail, then re-run `rein attest "+subject+"`")
		return
	}
	// like Note: the write IS the command, so failure fails
	if err := g.appendJournal("attest", map[string]any{
		"subject": subject, "gate": GatePath, "gate_hash": hash,
	}); err != nil {
		e.Fail("JOURNAL_WRITE_FAILED", JournalName+": "+err.Error(),
			"check permissions on .rein/ and re-run")
		return
	}
	e.Result = map[string]any{"recorded": JournalName, "subject": subject, "gate_hash": hash}
}

// lastAttestedGateHash returns the gate_hash of the most recent
// gate-can-fail attestation, or "" if none exists. Malformed journal
// lines are skipped here — `rein journal` owns reporting those.
func (g *Engine) lastAttestedGateHash() string {
	raw, err := os.ReadFile(filepath.Join(g.Repo, JournalName))
	if err != nil {
		return ""
	}
	last := ""
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var entry struct {
			Kind     string `json:"kind"`
			Subject  string `json:"subject"`
			GateHash string `json:"gate_hash"`
		}
		if json.Unmarshal([]byte(line), &entry) != nil {
			continue
		}
		if entry.Kind == "attest" && entry.Subject == "gate-can-fail" {
			last = entry.GateHash
		}
	}
	return last
}

// gateProofChecks is doctor's half of the gate-can-fail standing
// check (docs/lifecycle.md §4): never executing the gate, it audits
// whether the last proof still describes the installed gate. When the
// gate is the shipped stub, GATE_STUB already owns the surface and
// this stays silent — never stack a second finding on an addressed
// surface.
func (g *Engine) gateProofChecks(e *envelope.Envelope, gateIsStub bool) {
	if gateIsStub {
		return
	}
	current, err := g.gateHash()
	if err != nil {
		return // no installed gate: plan/apply territory, not a proof question
	}
	attested := g.lastAttestedGateHash()
	if attested == "" {
		e.Diag(envelope.Warning, "GATE_UNPROVEN",
			GatePath+" has never been proven able to fail — a gate that cannot fail is indistinguishable from a stub",
			"introduce a trivial breakage, confirm `bash "+GatePath+"` exits non-zero, revert, then run `rein attest gate-can-fail` (rein-setup step 7)")
		return
	}
	if attested != current {
		e.Diag(envelope.Warning, "GATE_PROOF_STALE",
			GatePath+" changed since it was last proven able to fail ("+short(attested)+" attested, "+short(current)+" installed)",
			"re-prove the changed gate: break something it covers, confirm non-zero exit, revert, then re-run `rein attest gate-can-fail`")
	}
}

func short(hash string) string {
	if len(hash) > 19 { // "sha256:" + 12
		return hash[:19]
	}
	return hash
}
