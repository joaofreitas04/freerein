# Progress

Managed by the working agent, structured for the next session's cold
start. Update before ending every session.

## Current state

2026-08-31 (third pass): native `measure` family in `rein inspect`,
implemented to spec/inspection.md v0.2 — authored by the concurrent
session (freerein-knowledge-17) and treated as the governing text;
code comments reconciled to its rule numbering: count-or-
classify per file (7 states), classified list, method-in-report,
formatVersion + engine stamp, generated/duplicate/shebang detection,
generic dot-dir pruning, cap-notes on high-touch and classified.
rein-setup 0.6 reads the measure and treats large generated/duplicate
shares as debt-ledger material. Verify green, self-hosted, live-run
on this repo (Go 25 files / ~4.5k lines; go.sum + harness.lock
classified generated). The read-error ruling is RESOLVED: spec and
engine gained an `error` state (eighth), split from `unknown`, with
a chmod-guarded test. bash scripts/verify GREEN at
hand-off; tree released back to the concurrent session for its
integration sweep.

2026-08-31 (second pass): setup-deepening milestone landed on top of
the lifecycle groundwork. New: spec/inspection.md + `rein inspect`
(mechanical discovery, never executes project code, works pre-init);
probe vocabulary now five entries sharing detection with inspect;
doctor: GATE_STUB (standing is-the-gate-real check) + DEBT_ROW_INCOMPLETE
/ DEBT_EXPIRED ledger audits; `rein note` + journal spec v0.2 (note
kind — procedures record rulings through the engine); dump/inspect
emit OUTPUT_OFFLOADED. Content: rein-setup 0.5 (inspect-first, chunked
rulings, verify + project-fragment skeletons, journaled rulings),
procedure-decide 0.1 (rein-decide: immutable records, refuse-before-
generate). Earlier same day: journal substrate, plan pricing,
addresses/overlap, seed-disposal fix, envelope v0.2, fitness +
lifecycle docs. All verify-green, self-host re-applied (rein-decide
skill installed), doctor clean. Changes uncommitted for review.

## In flight

Nothing.

## Assumptions / open questions

- The 80% NEAR_CONTEXT_BUDGET threshold is a chosen default, not a
  measured one; revisit when real repos hit it.
- `addresses` vocabulary is free-form; if overlap warnings prove noisy
  or vacuous, a registered vocabulary (like probes) is the likely fix.
- inspect's detection tables (manifests, test/lint/CI configs) are a
  curated list, not exhaustive; grow them from real setup runs, not
  speculation.

## Blocked / needs a human

- Registry hosting: publish the official index somewhere real
  (GitHub Pages / releases) and set a default registry URL.
- Codex skill frontmatter dialect: still a declared degradation —
  needs verified facts about the host before implementing.
- Review the uncommitted 2026-08-31 changes (two milestones: lifecycle
  groundwork + setup deepening; spec bumps are contract changes).
