package engine

import (
	"os"
	"sort"

	"github.com/joaofreitas04/freerein/engine/internal/component"
	"github.com/joaofreitas04/freerein/engine/internal/envelope"
)

// rein gaps: the join between what the repo lacks and what the
// registry offers. Each absent affordance (shared probe detection —
// gating, inspect, and gaps cannot disagree) is matched against the
// index's addresses fields (registry v0.3): a match is a coverage
// fact naming the extension; no match makes the gap a creation
// candidate for a human to commission. The join answers the empty
// registry honestly — publish first what detected gaps demand — and
// it deliberately stops at coverage: lift is what measurement
// decides, never a listing.

type Gap struct {
	Gap               string   `json:"gap"`
	Evidence          string   `json:"evidence"`
	AddressedBy       []string `json:"addressed_by,omitempty"`
	CreationCandidate bool     `json:"creation_candidate"`
}

var gapMisreadings = []string{
	"addressed-by is a coverage claim, never a lift claim — that an extension exists for a gap says nothing about whether it helps this repo; a correctly implemented component can still be a net loss",
	"an unaddressed gap is a finding for a human to commission, never an authoring licence — aggregated creation candidates are the registry's demand signal",
}

// Gaps works before init: discovery precedes declaration, and a
// repo without a harness is exactly where the gap map matters.
func (g *Engine) Gaps(e *envelope.Envelope, registryOverride string) {
	cfg, err := g.readConfig()
	if os.IsNotExist(err) {
		cfg = &Config{}
	} else if err != nil {
		e.Fail("CONFIG_INVALID", err.Error(), "fix "+ConfigName+" and re-run")
		return
	}
	idx := g.openRegistry(e, cfg, registryOverride)
	if idx == nil {
		return
	}
	rows := []Gap{}
	for _, p := range component.ProbeVocabulary {
		ok, reason := g.probe(p.Name)
		if ok {
			continue
		}
		row := Gap{Gap: p.Name, Evidence: reason}
		for name := range idx.Components {
			v, rel, lerr := idx.Latest(name)
			if lerr != nil {
				continue
			}
			for _, a := range rel.Addresses {
				if a == p.Name {
					row.AddressedBy = append(row.AddressedBy, name+"@"+v)
				}
			}
		}
		sort.Strings(row.AddressedBy)
		row.CreationCandidate = len(row.AddressedBy) == 0
		rows = append(rows, row)
	}
	e.Result = map[string]any{
		"registry":    registrySource(cfg, registryOverride),
		"gaps":        rows,
		"misreadings": gapMisreadings,
	}
}
