package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/joaofreitas04/freerein/engine/internal/envelope"
	"github.com/joaofreitas04/freerein/engine/internal/lockfile"
)

// rein report (spec/field-report.md, producer side): assemble the
// component-scoped shape from the lock, the journal, and the citation
// stats. The judgment fields — subsystem attribution, failure class,
// reproduction, requested disposition — arrive as flags, refused when
// missing (the propose precedent); everything else is mechanically
// scoped to the component, which is what makes the privacy rule hold
// by construction rather than by scrubbing. The report lands under
// .rein/out/report/ and nothing leaves: `off` is the only submission
// level implemented; propose/auto arrive with the transport.

// FieldReportFormatVersion matches spec/field-report.md's contract
// version; consumers may branch on it, additions bump the minor.
const FieldReportFormatVersion = "0.1.0"

// reportJournalCap bounds evidence.journal; the cut announces itself
// in notes (cli-envelope rule 5).
const reportJournalCap = 20

type ReportFields struct {
	Subsystem    string // the diagnose attribution, one of the five
	FailureClass string // short generalized statement of the defect
	Reproduction string // minimal generalized trigger, stated as affordances
	Disposition  string // fix | default-change | table-entry | docs
}

var reportSubsystems = map[string]bool{
	"instructions": true, "tools": true, "environment": true,
	"state": true, "feedback": true,
}

var reportDispositions = map[string]bool{
	"fix": true, "default-change": true, "table-entry": true, "docs": true,
}

type ReportShadow struct {
	Path        string `json:"path"`
	DisplacedBy string `json:"displaced_by"`
	Since       string `json:"since,omitempty"`        // first apply that landed the path
	LastApplied string `json:"last_applied,omitempty"` // newest apply that touched it
}

type ReportJournalEntry struct {
	At      string `json:"at,omitempty"`
	Kind    string `json:"kind"`
	Text    string `json:"text,omitempty"`
	Outcome string `json:"outcome,omitempty"`
}

type ReportCitation struct {
	Count int    `json:"count"`
	Last  string `json:"last"`
}

type FieldReport struct {
	FormatVersion string `json:"formatVersion"`
	Engine        string `json:"engine"`
	Component     struct {
		ID      string `json:"id"`
		Version string `json:"version"`
	} `json:"component"`
	Subsystem            string `json:"subsystem"`
	FailureClass         string `json:"failure_class"`
	Reproduction         string `json:"reproduction"`
	DispositionRequested string `json:"disposition_requested"`
	Evidence             struct {
		Journal   []ReportJournalEntry      `json:"journal"`
		Citations map[string]ReportCitation `json:"citations"`
		Shadows   []ReportShadow            `json:"shadows"`
	} `json:"evidence"`
	Fingerprint string   `json:"fingerprint"`
	Notes       []string `json:"notes,omitempty"`
}

// componentVersion finds name's installed version in the lock: file
// entries carry refs as name@version, comma-joined on rendered files,
// and shadows keep the displaced provider's ref.
func componentVersion(lock *lockfile.Lock, name string) string {
	scan := func(ref string) string {
		ref = strings.TrimSpace(ref)
		if strings.HasPrefix(ref, name+"@") {
			return strings.TrimPrefix(ref, name+"@")
		}
		return ""
	}
	for _, entry := range lock.Files {
		for _, ref := range strings.Split(entry.Component, ",") {
			if v := scan(ref); v != "" {
				return v
			}
		}
		for _, ref := range entry.Shadowed {
			if v := scan(ref); v != "" {
				return v
			}
		}
	}
	return ""
}

// reportShadows lists standing overrides displacing the component's
// files, with since-dates read from the journal's apply entries.
func (g *Engine) reportShadows(lock *lockfile.Lock, name string) []ReportShadow {
	firstApplied, lastApplied := g.applyDates()
	var out []ReportShadow
	var paths []string
	byPath := map[string]ReportShadow{}
	for path, entry := range lock.Files {
		for _, ref := range entry.Shadowed {
			if strings.HasPrefix(strings.TrimSpace(ref), name+"@") {
				byPath[path] = ReportShadow{
					Path:        path,
					DisplacedBy: strings.TrimSpace(entry.Component),
					Since:       firstApplied[path],
					LastApplied: lastApplied[path],
				}
				paths = append(paths, path)
			}
		}
	}
	sort.Strings(paths)
	for _, p := range paths {
		out = append(out, byPath[p])
	}
	return out
}

// applyDates scans the journal's apply entries oldest-first and
// returns, per path, the first and newest `at` that landed it. Facts
// only; what a standing shadow means is the reader's judgment.
func (g *Engine) applyDates() (first, last map[string]string) {
	first, last = map[string]string{}, map[string]string{}
	raw, err := os.ReadFile(filepath.Join(g.Repo, JournalName))
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var entry struct {
			At      string   `json:"at"`
			Kind    string   `json:"kind"`
			Applied []string `json:"applied"`
		}
		if json.Unmarshal([]byte(line), &entry) != nil || entry.Kind != "apply" {
			continue
		}
		for _, p := range entry.Applied {
			if _, ok := first[p]; !ok {
				first[p] = entry.At
			}
			last[p] = entry.At
		}
	}
	return
}

// reportJournal collects judgment-kind entries (note, proposal,
// verdict) that mention the component by name — the mechanical half
// of "scoped to the component"; whether they are relevant stays the
// reader's call. Newest first, capped.
func (g *Engine) reportJournal(name string) (entries []ReportJournalEntry, total int) {
	raw, err := os.ReadFile(filepath.Join(g.Repo, JournalName))
	if err != nil {
		return nil, 0
	}
	lines := strings.Split(string(raw), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if strings.TrimSpace(line) == "" || !strings.Contains(line, name) {
			continue
		}
		var entry struct {
			At      string `json:"at"`
			Kind    string `json:"kind"`
			Text    string `json:"text"`
			Change  string `json:"change"`
			Outcome string `json:"outcome"`
		}
		if json.Unmarshal([]byte(line), &entry) != nil {
			continue
		}
		switch entry.Kind {
		case "note", "proposal", "verdict":
			total++
			if len(entries) < reportJournalCap {
				text := entry.Text
				if text == "" {
					text = entry.Change
				}
				entries = append(entries, ReportJournalEntry{
					At: entry.At, Kind: entry.Kind, Text: text, Outcome: entry.Outcome,
				})
			}
		}
	}
	return entries, total
}

// Report assembles a field report for one installed component. It
// writes; it never transmits.
func (g *Engine) Report(e *envelope.Envelope, name string, f ReportFields) {
	var missing []string
	for _, req := range []struct{ flag, v string }{
		{"--subsystem", f.Subsystem}, {"--failure-class", f.FailureClass},
		{"--reproduction", f.Reproduction}, {"--disposition", f.Disposition},
	} {
		if strings.TrimSpace(req.v) == "" {
			missing = append(missing, req.flag)
		}
	}
	if len(missing) > 0 {
		e.Fail("MISSING_ARGUMENT", "a report is refused with a hole: "+strings.Join(missing, ", ")+" missing",
			"every judgment field is mandatory — subsystem (the diagnose attribution), "+
				"failure-class (generalized, component-scoped), reproduction (inputs as affordances, "+
				"expected vs observed), disposition (fix | default-change | table-entry | docs)")
		return
	}
	if !reportSubsystems[f.Subsystem] {
		known := make([]string, 0, len(reportSubsystems))
		for s := range reportSubsystems {
			known = append(known, s)
		}
		sort.Strings(known)
		e.Fail("INVALID_ARGUMENT", "unknown subsystem "+f.Subsystem,
			"--subsystem is one of: "+strings.Join(known, ", "))
		return
	}
	if !reportDispositions[f.Disposition] {
		e.Fail("INVALID_ARGUMENT", "unknown disposition "+f.Disposition,
			"--disposition is one of: fix, default-change, table-entry, docs")
		return
	}
	lock, err := lockfile.Read(g.Repo)
	if err != nil {
		e.Fail("LOCKFILE_INVALID", err.Error(), "inspect or delete "+lockfile.Name+" and re-run")
		return
	}
	if lock == nil {
		e.Fail("NOT_INITIALIZED", "no "+lockfile.Name+" — a report attributes to an installed component",
			"run `rein init` and `rein apply` first; a defect in an uninstalled component has no lock facts to carry")
		return
	}
	version := componentVersion(lock, name)
	if version == "" {
		e.Fail("NOT_FOUND", "no component "+name+" in "+lockfile.Name,
			"`rein dump` lists the installed composition; the component id is the version-free name")
		return
	}

	r := &FieldReport{FormatVersion: FieldReportFormatVersion, Engine: g.engineID()}
	r.Component.ID = name
	r.Component.Version = version
	r.Subsystem = f.Subsystem
	r.FailureClass = f.FailureClass
	r.Reproduction = f.Reproduction
	r.DispositionRequested = f.Disposition

	var journalTotal int
	r.Evidence.Journal, journalTotal = g.reportJournal(name)
	if journalTotal > reportJournalCap {
		r.Notes = append(r.Notes, fmt.Sprintf("journal evidence capped at %d of %d matching entries", reportJournalCap, journalTotal))
	}
	r.Evidence.Citations = map[string]ReportCitation{}
	if store, err := g.readCitations(); err == nil {
		for id, c := range store.Citations {
			if strings.HasPrefix(id, name+":") {
				r.Evidence.Citations[id] = ReportCitation{Count: c.Count, Last: c.Last}
			}
		}
	}
	r.Evidence.Shadows = g.reportShadows(lock, name)

	sum := sha256.Sum256([]byte(name + "@" + version + "|" + f.FailureClass))
	r.Fingerprint = hex.EncodeToString(sum[:])[:16]

	rel := filepath.Join(OutDir, "report", name+".json")
	full := filepath.Join(g.Repo, rel)
	b, _ := json.MarshalIndent(r, "", "  ")
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		e.Fail("WRITE_FAILED", err.Error(), "check permissions on "+OutDir)
		return
	}
	if err := os.WriteFile(full, append(b, '\n'), 0o644); err != nil {
		e.Fail("WRITE_FAILED", err.Error(), "check permissions on "+OutDir)
		return
	}
	e.Diag(envelope.Info, "OUTPUT_OFFLOADED", "report written to "+rel,
		"the human reviews the file before it goes anywhere — submission is governed by report.submit, and only `off` exists today")
	e.Result = map[string]any{
		"detail":      rel,
		"component":   name + "@" + version,
		"subsystem":   f.Subsystem,
		"disposition": f.Disposition,
		"fingerprint": r.Fingerprint,
		"evidence": map[string]int{
			"journal":   len(r.Evidence.Journal),
			"citations": len(r.Evidence.Citations),
			"shadows":   len(r.Evidence.Shadows),
		},
		"submit": "off — nothing leaves the repo without the submission policy (spec/field-report.md)",
	}
}
