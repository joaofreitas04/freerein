package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/joaofreitas04/freerein/engine/internal/envelope"
	"github.com/joaofreitas04/freerein/engine/internal/resolve"
)

// CitationsPath is the citation store (spec/citation.md): counters,
// not history, which is why it is a rewritten JSON file and not the
// journal. Committed, engine-written, deterministic key order.
const CitationsPath = ".rein/stats/citations.json"

// CitationFormatVersion matches spec/citation.md's contract version.
const CitationFormatVersion = "0.1.0"

type citation struct {
	Count int    `json:"count"`
	Last  string `json:"last"`
}

type citationStore struct {
	FormatVersion string              `json:"formatVersion"`
	Citations     map[string]citation `json:"citations"`
}

// The §1.5 misreading, attached wherever counts are shown: a number
// without its characteristic lie invites the wrong conclusion.
const citationMisreading = "zero citations can mean a dead rule or a fully internalized one — " +
	"decay candidates are ranked for human ablation, never for automatic removal"

// fragmentIDs lists the current composition's citable fragments,
// sorted. Derivation is shared with the render marker
// (resolve.FragmentID) so the vocabulary an agent sees and the one
// cite accepts cannot disagree.
func (r *resolution) fragmentIDs() []string {
	var ids []string
	for p, entry := range r.set.Entries {
		if strings.HasPrefix(p, resolve.FragmentDir) {
			ids = append(ids, resolve.FragmentID(entry.Winner.Component, p))
		}
	}
	sort.Strings(ids)
	return ids
}

func (g *Engine) readCitations() (*citationStore, error) {
	store := &citationStore{FormatVersion: CitationFormatVersion, Citations: map[string]citation{}}
	b, err := os.ReadFile(filepath.Join(g.Repo, CitationsPath))
	if os.IsNotExist(err) {
		return store, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, store); err != nil {
		return nil, err
	}
	if store.Citations == nil {
		store.Citations = map[string]citation{}
	}
	return store, nil
}

// Cite records that a fragment shaped an action (id non-empty), or
// reads the counts back joined against the current composition
// (id empty). The counts are advisory and rank decay candidates —
// deciding what a zero means is judgment, and the engine says so in
// its own output.
func (g *Engine) Cite(e *envelope.Envelope, id string) {
	r := g.resolveAll(e)
	if r == nil {
		return
	}
	ids := r.fragmentIDs()
	if id == "" {
		g.citeReport(e, r, ids)
		return
	}
	valid := false
	for _, known := range ids {
		if known == id {
			valid = true
			break
		}
	}
	if !valid {
		e.Fail("INVALID_ARGUMENT", "unknown fragment id "+id,
			"the current composition's fragments are: "+strings.Join(ids, ", ")+
				" (the `rein:fragment` markers in the instruction file carry them)")
		return
	}
	store, err := g.readCitations()
	if err != nil {
		e.Fail("CITATIONS_INVALID", CitationsPath+": "+err.Error(),
			"the store is engine-written — restore it from git, or delete it to start counting fresh")
		return
	}
	c := store.Citations[id]
	c.Count++
	c.Last = time.Now().UTC().Format(time.RFC3339)
	store.Citations[id] = c
	b, _ := json.MarshalIndent(store, "", "  ")
	path := filepath.Join(g.Repo, CitationsPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err == nil {
		err = os.WriteFile(path, append(b, '\n'), 0o644)
	}
	if err != nil {
		e.Fail("WRITE_FAILED", err.Error(), "check permissions on "+filepath.Dir(CitationsPath))
		return
	}
	e.Result = map[string]any{"id": id, "count": c.Count, "recorded": CitationsPath}
}

// citeReport joins counts against the composition. Fragments in the
// store but no longer composed keep their rows in the file (counts
// survive a remove/re-add) and stay out of the report.
func (g *Engine) citeReport(e *envelope.Envelope, r *resolution, ids []string) {
	store, err := g.readCitations()
	if err != nil {
		e.Fail("CITATIONS_INVALID", CitationsPath+": "+err.Error(),
			"the store is engine-written — restore it from git, or delete it to start counting fresh")
		return
	}
	type row struct {
		ID    string `json:"id"`
		Count int    `json:"count"`
		Last  string `json:"last,omitempty"`
	}
	rows := make([]row, 0, len(ids))
	var decay []string
	for _, id := range ids {
		c := store.Citations[id]
		rows = append(rows, row{ID: id, Count: c.Count, Last: c.Last})
		if c.Count == 0 {
			decay = append(decay, id)
		}
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].Count < rows[j].Count })
	result := map[string]any{
		"formatVersion": CitationFormatVersion,
		"fragments":     rows,
		"misreading":    citationMisreading,
	}
	if len(decay) > 0 {
		result["decay_candidates"] = decay
	}
	e.Result = result
}
