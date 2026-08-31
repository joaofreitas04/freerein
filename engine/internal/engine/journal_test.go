package engine_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joaofreitas04/freerein/content"
	"github.com/joaofreitas04/freerein/engine/internal/engine"
	"github.com/joaofreitas04/freerein/engine/internal/envelope"
)

func journalRepo(t *testing.T, lines ...string) (*engine.Engine, string) {
	t.Helper()
	repo := t.TempDir()
	if len(lines) > 0 {
		seedFile(t, repo, engine.JournalName, strings.Join(lines, "\n")+"\n")
	}
	return &engine.Engine{Repo: repo, Content: content.FS}, repo
}

func readJournal(t *testing.T, g *engine.Engine, opts engine.JournalOpts) (*envelope.Envelope, map[string]any) {
	t.Helper()
	e := envelope.New("journal")
	g.Journal(e, opts)
	r, _ := e.Result.(map[string]any)
	return e, r
}

func diag(e *envelope.Envelope, code string) *envelope.Diagnostic {
	for i := range e.Diagnostics {
		if e.Diagnostics[i].Code == code {
			return &e.Diagnostics[i]
		}
	}
	return nil
}

// spec/journal.md Reading: newest first by file position, filters on
// kind, since (at is content), and path.
func TestJournalFiltersAndOrder(t *testing.T) {
	g, _ := journalRepo(t,
		`{"at":"2026-08-01T00:00:00Z","kind":"apply","applied":["CLAUDE.md"],"conflicts":[]}`,
		`{"at":"2026-08-02T00:00:00Z","kind":"note","text":"a ruling"}`,
		`{"at":"2026-08-03T00:00:00Z","kind":"apply","applied":["scripts/verify"],"conflicts":["CLAUDE.md"]}`,
	)
	e, r := readJournal(t, g, engine.JournalOpts{Limit: 20})
	if !e.OK {
		t.Fatalf("journal failed: %+v", e.Diagnostics)
	}
	entries := r["entries"].([]map[string]any)
	if len(entries) != 3 || entries[0]["at"] != "2026-08-03T00:00:00Z" {
		t.Fatalf("want 3 entries newest first, got %+v", entries)
	}
	if r["formatVersion"] != engine.JournalFormatVersion {
		t.Fatalf("result must carry the contract version, got %v", r["formatVersion"])
	}

	_, r = readJournal(t, g, engine.JournalOpts{Kind: "note", Limit: 20})
	if r["total"].(int) != 1 {
		t.Fatalf("--kind note: want 1, got %v", r["total"])
	}
	_, r = readJournal(t, g, engine.JournalOpts{Since: "2026-08-02", Limit: 20})
	if r["total"].(int) != 2 {
		t.Fatalf("--since date: want 2, got %v", r["total"])
	}
	_, r = readJournal(t, g, engine.JournalOpts{Path: "CLAUDE.md", Limit: 20})
	if r["total"].(int) != 2 { // applied once, conflicted once
		t.Fatalf("--path must match applied and conflicts, got %v", r["total"])
	}
}

// Applied and conflicted are different facts, counted separately; the
// most-conflicted surface sorts first.
func TestJournalSurfaces(t *testing.T) {
	g, _ := journalRepo(t,
		`{"at":"2026-08-01T00:00:00Z","kind":"apply","applied":["a.md","b.md"],"conflicts":[]}`,
		`{"at":"2026-08-02T00:00:00Z","kind":"apply","applied":["a.md"],"conflicts":["b.md"]}`,
		`{"at":"2026-08-03T00:00:00Z","kind":"upgrade","upgraded":["x@1 -> x@2"]}`,
	)
	_, r := readJournal(t, g, engine.JournalOpts{Limit: 20})
	b, _ := json.Marshal(r["surfaces"])
	var surfaces []struct {
		Path       string `json:"path"`
		Applied    int    `json:"applied"`
		Conflicted int    `json:"conflicted"`
		First      string `json:"first"`
		Last       string `json:"last"`
	}
	if err := json.Unmarshal(b, &surfaces); err != nil {
		t.Fatal(err)
	}
	if len(surfaces) != 2 {
		t.Fatalf("upgrade refs are not paths; want 2 surfaces, got %+v", surfaces)
	}
	// b.md conflicted once, so it outranks a.md (applied twice)
	if surfaces[0].Path != "b.md" || surfaces[0].Applied != 1 || surfaces[0].Conflicted != 1 {
		t.Fatalf("most-conflicted sorts first, got %+v", surfaces[0])
	}
	if surfaces[1].Path != "a.md" || surfaces[1].Applied != 2 || surfaces[1].Conflicted != 0 {
		t.Fatalf("applied and conflicted are separate counts, got %+v", surfaces[1])
	}
	if surfaces[1].First != "2026-08-01T00:00:00Z" || surfaces[1].Last != "2026-08-02T00:00:00Z" {
		t.Fatalf("first/last span the touches, got %+v", surfaces[1])
	}
}

func TestJournalLimitZero(t *testing.T) {
	g, _ := journalRepo(t,
		`{"at":"2026-08-01T00:00:00Z","kind":"apply","applied":["a.md"],"conflicts":[]}`,
	)
	_, r := readJournal(t, g, engine.JournalOpts{Limit: 0})
	if r["total"].(int) != 1 || r["returned"].(int) != 0 {
		t.Fatalf("limit 0 = counts only, got total=%v returned=%v", r["total"], r["returned"])
	}
	if _, ok := r["surfaces"]; !ok {
		t.Fatal("surfaces must still be computed at --limit 0")
	}
}

// A malformed line is a warning, never a refusal: valid entries are
// still returned, and the fix says append, never edit — the journal
// is append-only.
func TestJournalUnreadableLine(t *testing.T) {
	g, _ := journalRepo(t,
		`{"at":"2026-08-01T00:00:00Z","kind":"apply","applied":["a.md"],"conflicts":[]}`,
		`{"at":"2026-08-02T00:00:00Z","kind":"appl`,
	)
	e, r := readJournal(t, g, engine.JournalOpts{Limit: 20})
	if !e.OK {
		t.Fatalf("a bad line must not fail the read: %+v", e.Diagnostics)
	}
	d := diag(e, "JOURNAL_LINE_UNREADABLE")
	if d == nil || d.Severity != envelope.Warning {
		t.Fatalf("want JOURNAL_LINE_UNREADABLE warning, got %+v", e.Diagnostics)
	}
	if !strings.Contains(d.Message, "line 2") {
		t.Fatalf("must name the first bad line, got %q", d.Message)
	}
	if !strings.Contains(d.Fix, "append") || !strings.Contains(d.Fix, "never edit") {
		t.Fatalf("fix must protect append-only, got %q", d.Fix)
	}
	if r["total"].(int) != 1 {
		t.Fatalf("valid entries must survive, got %v", r["total"])
	}
}

// Rule 3: kinds the engine does not recognize survive verbatim,
// custom fields included.
func TestJournalUnknownKindPreserved(t *testing.T) {
	g, _ := journalRepo(t,
		`{"at":"2026-08-01T00:00:00Z","kind":"acceptance-verdict","verdict":"held","trial":7}`,
	)
	_, r := readJournal(t, g, engine.JournalOpts{Kind: "acceptance-verdict", Limit: 20})
	entries := r["entries"].([]map[string]any)
	if len(entries) != 1 || entries[0]["verdict"] != "held" || entries[0]["trial"] != float64(7) {
		t.Fatalf("unknown kind must survive verbatim, got %+v", entries)
	}
}

// An absent journal is a fact, not a failure.
func TestJournalAbsent(t *testing.T) {
	g, _ := journalRepo(t)
	e, r := readJournal(t, g, engine.JournalOpts{Limit: 20})
	if !e.OK || diag(e, "JOURNAL_ABSENT") == nil {
		t.Fatalf("want ok + JOURNAL_ABSENT, got ok=%v %+v", e.OK, e.Diagnostics)
	}
	if r["total"].(int) != 0 {
		t.Fatalf("want empty result, got %+v", r)
	}
}

// Rule 5: the cut announces itself with the limit, and the detail
// file holds what the envelope omits.
func TestJournalOffloadCut(t *testing.T) {
	g, repo := journalRepo(t,
		`{"at":"2026-08-01T00:00:00Z","kind":"note","text":"one"}`,
		`{"at":"2026-08-02T00:00:00Z","kind":"note","text":"two"}`,
	)
	e, r := readJournal(t, g, engine.JournalOpts{Limit: 1})
	d := diag(e, "OUTPUT_OFFLOADED")
	if d == nil || !strings.Contains(d.Message, "capped at 1 of 2") {
		t.Fatalf("cut must announce the limit, got %+v", e.Diagnostics)
	}
	if d.Fix == "" {
		t.Fatal("OUTPUT_OFFLOADED carries a fix naming the detail path")
	}
	b, err := os.ReadFile(filepath.Join(repo, r["detail"].(string)))
	if err != nil {
		t.Fatal(err)
	}
	var full struct {
		Entries []map[string]any `json:"entries"`
	}
	if err := json.Unmarshal(b, &full); err != nil {
		t.Fatal(err)
	}
	if len(full.Entries) != 2 {
		t.Fatalf("detail file must hold the full set, got %d", len(full.Entries))
	}
}

// Timestamps are content: entries without a parsable `at` are
// excluded by --since, and the exclusion announces itself in notes —
// a silent cut anywhere is a contract violation.
func TestJournalSinceExcludesUnparseableAt(t *testing.T) {
	g, _ := journalRepo(t,
		`{"kind":"note","text":"hand-appended, no at"}`,
		`{"at":"2026-08-02T00:00:00Z","kind":"note","text":"dated"}`,
	)
	_, r := readJournal(t, g, engine.JournalOpts{Since: "2026-08-01", Limit: 20})
	if r["total"].(int) != 1 {
		t.Fatalf("want 1, got %v", r["total"])
	}
	notes, _ := r["notes"].([]string)
	if len(notes) == 0 || !strings.Contains(notes[0], "excluded by --since") {
		t.Fatalf("exclusion must announce itself, got %v", r["notes"])
	}
}
