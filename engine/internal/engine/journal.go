package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

// JournalFormatVersion stamps the read result and matches
// spec/journal.md's contract version (the inspect convention):
// consumers may branch on it, and additions bump the minor.
const JournalFormatVersion = "0.3.0"

// surfaceCap bounds the surfaces table inline; the cap announces
// itself in notes and the detail file carries the full table.
const surfaceCap = 20

// JournalOpts are the read filters for Journal.
type JournalOpts struct {
	Kind  string // exact match, including kinds the engine does not recognize
	Since string // RFC3339, or YYYY-MM-DD meaning UTC midnight
	Path  string // entries whose applied[]/conflicts[] contain this path
	Limit int    // inline entry cap; 0 = counts and surfaces only
}

// journalSurface is the mechanical per-path aggregation over apply
// entries. Applied and conflicted are counted separately on purpose:
// "touched three times" and "conflicted three times" are different
// facts, and the second is what marks a surface as fought over.
// Deciding whether two failures are the same failure is judgment and
// stays in the diagnose procedure — the engine only counts.
type journalSurface struct {
	Path       string `json:"path"`
	Applied    int    `json:"applied"`
	Conflicted int    `json:"conflicted"`
	First      string `json:"first,omitempty"`
	Last       string `json:"last,omitempty"`
}

// Journal reads the harness history back out (spec/journal.md,
// Reading). Entries are decoded as raw maps so kinds the engine does
// not recognize survive verbatim — rule 3 held by construction.
// Ordering is file position reversed (newest first): position is
// mechanical, timestamps are content — hand-appended entries are a
// sanctioned repair and their `at` may be absent or out of order.
func (g *Engine) Journal(e *envelope.Envelope, opts JournalOpts) {
	if opts.Limit < 0 {
		e.Fail("INVALID_ARGUMENT", "--limit must be >= 0",
			"pass a non-negative entry cap; 0 returns counts and surfaces only")
		return
	}
	var since time.Time
	if opts.Since != "" {
		t, err := time.Parse(time.RFC3339, opts.Since)
		if err != nil {
			t, err = time.Parse("2006-01-02", opts.Since)
		}
		if err != nil {
			e.Fail("INVALID_ARGUMENT", "--since "+opts.Since+": not RFC3339 or YYYY-MM-DD",
				"use e.g. 2026-08-31T12:00:00Z or 2026-08-31 (UTC midnight)")
			return
		}
		since = t.UTC()
	}

	raw, err := os.ReadFile(filepath.Join(g.Repo, JournalName))
	if os.IsNotExist(err) {
		e.Diag(envelope.Info, "JOURNAL_ABSENT", JournalName+": no journal yet",
			"the journal begins with the first completed state change (`rein apply`)")
		e.Result = map[string]any{
			"formatVersion": JournalFormatVersion, "engine": g.engineID(),
			"total": 0, "returned": 0, "entries": []any{},
		}
		return
	}
	if err != nil {
		e.Fail("READ_FAILED", JournalName+": "+err.Error(),
			"check permissions on .rein/ and re-run")
		return
	}

	var (
		filtered   []map[string]any
		unreadable int
		firstBad   int
		noAt       int
	)
	for i, line := range strings.Split(string(raw), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var entry map[string]any
		if json.Unmarshal([]byte(line), &entry) != nil {
			if unreadable == 0 {
				firstBad = i + 1
			}
			unreadable++
			continue
		}
		if opts.Kind != "" {
			if k, _ := entry["kind"].(string); k != opts.Kind {
				continue
			}
		}
		if !since.IsZero() {
			at, ok := entryTime(entry)
			if !ok {
				noAt++
				continue
			}
			if at.Before(since) {
				continue
			}
		}
		if opts.Path != "" && !entryTouches(entry, opts.Path) {
			continue
		}
		filtered = append(filtered, entry)
	}
	if unreadable > 0 {
		e.Diag(envelope.Warning, "JOURNAL_LINE_UNREADABLE",
			fmt.Sprintf("%s: %d line(s) unreadable (first at line %d); valid entries are still included",
				JournalName, unreadable, firstBad),
			"append a corrected entry; never edit or delete the bad line — the journal is append-only")
	}

	surfaces, surfaceNote := journalSurfaces(filtered, opts.Path)

	// newest first: reverse file order
	for i, j := 0, len(filtered)-1; i < j; i, j = i+1, j-1 {
		filtered[i], filtered[j] = filtered[j], filtered[i]
	}
	total := len(filtered)
	inline := filtered
	if opts.Limit < total {
		inline = filtered[:opts.Limit]
	}

	var notes []string
	if noAt > 0 {
		notes = append(notes, fmt.Sprintf("%d entries lack a parsable 'at' and were excluded by --since", noAt))
	}
	if surfaceNote != "" {
		notes = append(notes, surfaceNote)
	}

	outPath := filepath.ToSlash(filepath.Join(OutDir, "journal.json"))
	full := map[string]any{
		"formatVersion": JournalFormatVersion, "engine": g.engineID(),
		"total": total, "entries": filtered,
	}
	if len(surfaces) > 0 {
		full["surfaces"] = surfaces
	}
	b, _ := json.MarshalIndent(full, "", "  ")
	if err := os.MkdirAll(filepath.Join(g.Repo, OutDir), 0o755); err == nil {
		err = os.WriteFile(filepath.Join(g.Repo, outPath), append(b, '\n'), 0o644)
	}
	if err != nil {
		e.Fail("WRITE_FAILED", err.Error(), "check permissions on "+OutDir)
		return
	}
	offload := "full journal read written to " + outPath
	if len(inline) < total {
		offload += fmt.Sprintf("; inline entries capped at %d of %d by --limit", len(inline), total)
	}
	e.Diag(envelope.Info, "OUTPUT_OFFLOADED", offload,
		"read "+outPath+" for the entries omitted from this envelope")

	result := map[string]any{
		"formatVersion": JournalFormatVersion, "engine": g.engineID(),
		"total": total, "returned": len(inline),
		"detail": outPath, "entries": inline,
	}
	capped := surfaces
	if len(capped) > surfaceCap {
		capped = capped[:surfaceCap]
	}
	if len(capped) > 0 {
		result["surfaces"] = capped
	}
	if len(notes) > 0 {
		result["notes"] = notes
	}
	e.Result = result
}

// entryTime parses an entry's `at` as RFC3339, reporting whether it
// could. Timestamps are content, not position — they may be absent or
// malformed on hand-appended entries and the reader must say so
// rather than guess.
func entryTime(entry map[string]any) (time.Time, bool) {
	s, _ := entry["at"].(string)
	if s == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, s)
	return t, err == nil
}

// entryTouches reports whether path appears in the entry's applied[]
// or conflicts[] lists. Only apply entries carry paths; add records a
// component, upgrade records "from -> to" refs.
func entryTouches(entry map[string]any, path string) bool {
	for _, key := range []string{"applied", "conflicts"} {
		list, _ := entry[key].([]any)
		for _, v := range list {
			if s, _ := v.(string); s == path {
				return true
			}
		}
	}
	return false
}

// journalSurfaces aggregates apply entries per path, in filtered file
// order. Returned sorted conflicted desc, applied desc, path asc; the
// note is non-empty when the table was capped.
func journalSurfaces(entries []map[string]any, onlyPath string) ([]journalSurface, string) {
	byPath := map[string]*journalSurface{}
	touch := func(entry map[string]any, key string) {
		list, _ := entry[key].([]any)
		for _, v := range list {
			p, _ := v.(string)
			if p == "" || (onlyPath != "" && p != onlyPath) {
				continue
			}
			s := byPath[p]
			if s == nil {
				s = &journalSurface{Path: p}
				byPath[p] = s
			}
			if key == "applied" {
				s.Applied++
			} else {
				s.Conflicted++
			}
			if at, ok := entryTime(entry); ok {
				ts := at.UTC().Format(time.RFC3339)
				if s.First == "" || ts < s.First {
					s.First = ts
				}
				if ts > s.Last {
					s.Last = ts
				}
			}
		}
	}
	for _, entry := range entries {
		if k, _ := entry["kind"].(string); k != "apply" {
			continue
		}
		touch(entry, "applied")
		touch(entry, "conflicts")
	}
	out := make([]journalSurface, 0, len(byPath))
	for _, s := range byPath {
		out = append(out, *s)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Conflicted != out[j].Conflicted {
			return out[i].Conflicted > out[j].Conflicted
		}
		if out[i].Applied != out[j].Applied {
			return out[i].Applied > out[j].Applied
		}
		return out[i].Path < out[j].Path
	})
	note := ""
	if len(out) > surfaceCap {
		note = fmt.Sprintf("surfaces capped at %d of %d paths; full table in the detail file", surfaceCap, len(out))
	}
	return out, note
}
