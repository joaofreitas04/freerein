package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joaofreitas04/freerein/engine/internal/envelope"
)

// JournalName is the append-only harness history (spec/journal.md):
// one JSON object per line, engine-written on every completed state
// change, never rewritten. The lockfile is the harness's state; this
// is how it got there.
const JournalName = ".rein/journal.jsonl"

// appendJournal writes one raw entry, returning any error.
func (g *Engine) appendJournal(kind string, fields map[string]any) error {
	entry := map[string]any{
		"at":     time.Now().UTC().Format(time.RFC3339),
		"engine": g.engineID(),
		"kind":   kind,
	}
	for k, v := range fields {
		entry[k] = v
	}
	line, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	path := filepath.Join(g.Repo, JournalName)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	_, werr := f.Write(append(line, '\n'))
	if cerr := f.Close(); werr == nil {
		werr = cerr
	}
	return werr
}

// journal appends one entry after a completed state change. The change
// has already happened, so a write failure is loud but never fatal —
// the command must not fail because its record could not be kept.
func (g *Engine) journal(e *envelope.Envelope, kind string, fields map[string]any) {
	if err := g.appendJournal(kind, fields); err != nil {
		e.Diag(envelope.Warning, "JOURNAL_WRITE_FAILED",
			JournalName+": "+err.Error()+" — the change itself succeeded but was not recorded",
			"check permissions on .rein/ and re-run the command, or append the entry by hand")
	}
}

// Note records free text a judgment procedure wants kept (spec/journal.md
// rule 3): setup rulings, rejected candidates, recorded decisions. Here
// the write IS the command, so failure fails.
func (g *Engine) Note(e *envelope.Envelope, text string) {
	if strings.TrimSpace(text) == "" {
		e.Fail("MISSING_ARGUMENT", "note needs text to record",
			"e.g. `rein note \"setup: demoted 3 rules to verify checks; rejected slow e2e suite as a gate\"`")
		return
	}
	if err := g.appendJournal("note", map[string]any{"text": text}); err != nil {
		e.Fail("JOURNAL_WRITE_FAILED", JournalName+": "+err.Error(),
			"check permissions on .rein/ and re-run")
		return
	}
	e.Result = map[string]any{"recorded": JournalName}
}
