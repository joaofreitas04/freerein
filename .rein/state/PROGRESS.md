# Progress

Managed by the working agent, structured for the next session's cold
start. Update before ending every session. State is more than a task
list: record what you now believe and have not proven, or the next
session re-derives it — or worse, trusts it.

## Current state

Clean baseline as of 2026-08-31 (evening session). `bash
scripts/verify` GREEN; `rein doctor` reports exactly one finding,
the expected `SHADOW_STANDING` advisory on the adopted verify gate
(info — local policy, no action). All work committed and pushed.
No TODO/FIXME markers in the tree. The repo is self-hosted: CLAUDE.md is rendered by rein —
edit `.rein/overrides/`, never CLAUDE.md or scripts/verify directly,
then `rein apply --yes`.

What is shipped is described in `AGENTS.md` § Current state; do not
duplicate that inventory here. This file carries only what does not
survive in the code: what is next, what is unproven, what is blocked.

**Next, in `docs/lifecycle.md` §4 order:**

1. **Use the loop on real work.** The lifecycle exists end to end —
   init → operate → observe → diagnose → evolve → upgrade — and
   every further mechanism should be demanded by a real run, not
   speculated. Field test 2 is concluded (below); what remains is
   sessions, on this repo and the field repo, accruing citation and
   journal data — and now eval runs: the instrument is live, the
   task set awaits the owner's curation (Blocked). The first evolve
   proposal on THIS repo waits for that data (day-one zeros are the
   misreading, not evidence).
2. **Doctor in the gate, on THIS repo.** The standing open question
   has its answer on file: doctor belongs in
   `.rein/overrides/scripts/verify` as repo policy — self-hosted
   means a hand-edited CLAUDE.md is a real failure mode only doctor
   catches — never as a product default, and CI keeps running
   nothing the gate does not. Landing it is a change to the
   definition of done: prove the changed gate red, then re-attest.
   The check must key on doctor's ok plus error/warning findings,
   not the findings count — the SHADOW_STANDING info advisory is
   the expected baseline here.
3. **First registry publication.** The demand signal now exists —
   `rein gaps` aggregates creation candidates — and the index
   carries `addresses` (registry v0.3). What ships first stays the
   owner's product decision; a spec-flow procedure remains the
   named candidate.
4. **Submission transport** (`report.submit` propose/auto via gh,
   fingerprint dedup onto shared issue threads) — built when a real
   report demands it; the invariant and constraints are already
   written (field-report v0.3).
5. **Detection rulings waiting on more repos** (evidence below):
   plain-angular multi-project as a monorepo marker; scaffold-vs-
   copy duplicate grouping in the measure; the host-file redirect
   stub question.

Shipped 2026-08-31, this list's previous occupants: every field-
demanded fix from field test 2 (fixture-red-first, corpus at
`field_test.go`); the paired eval as ruled (spec/eval.md v0.1);
the field-report channel as ruled (spec v0.3 — `rein report`,
diagnose 0.5 disposition, threshold-free SHADOW_STANDING); and
setup deepening as ruled by the owner (inspect v0.3: per-host
surfaces, prior_install, wiring, symlink
resolution; registry v0.3 addresses; `rein gaps`; rein-setup 0.9:
incumbent dispositions, the gap map, permissions as rung-placed
proposals with feasibility verdicts).

Deliberately not scheduled, each waiting on its trigger: the trust
ledger, per-model profiles, the visible-index budget check, and the
write quarantine — whose trigger is named: before any accepted
proposal auto-lands agent-authored content, the staging gate must
exist (until then, publish-signing covers the boundary).

Evolve mechanics shipped 2026-08-31 (journal v0.5, as designed and
ruled): `rein propose` (six mandatory fields, refuse-before-generate,
measurement named before the change, one open proposal at a time as
a hard refusal — id hashes salted with the proposal sequence after
the test caught same-second identical re-proposals resurrecting old
verdicts) / `rein verdict` (evidence-only, immutable, pre-named
measurement copied in) / doctor `PROPOSAL_OPEN` / procedure-evolve
0.1 (rein-evolve: nominate one candidate from observe evidence,
bias remove > strengthen > add). The engine enforces the shape of
acceptance; that the proposer never accepts is procedure
discipline — stated in the spec, not faked.

Minimal profile shipped 2026-08-31 (resolution v0.5): `profile:` in
harness.yaml selects core membership — standard (full) or minimal
(instructions-minimal 0.1 + verification-gate: the floor, 809 bytes
always-tier on a fresh install vs 2290 standard on this repo, zero
per-session and conditional costs). The control condition of
lifecycle §2.6, stated as such by init (`PROFILE_CONTROL`); switching
flows through plan/apply, leaves agent-owned seeds (SEED_LEFT), and
lands in the journal so a paired measurement can read when the
condition changed. Overrides/extensions apply under every profile:
the comparison varies the shipped core, holds project config
constant.

Cost surfaces shipped 2026-08-31 (resolution v0.4): plan's `costs`
object tiers every installed artifact by when its price is paid —
always (instruction file), per-session (seeds at on-disk size: this
repo's PROGRESS.md prices at ~4.9KB, more than twice the instruction
file, which is exactly the kind of fact the tier exists to surface),
conditional (each skill's description + body loads) — with unpriced
surfaces named and the misreading attached. Reporting only, no new
thresholds: a cutoff nobody measured is an unearned number.

Fragment citation telemetry shipped 2026-08-31 as designed and
ruled: spec/citation.md v0.1, resolution v0.3 markers, `rein cite`,
instructions-base 0.3. Live on this repo with three honest citations;
`local-overrides:10-project` is the sole day-one decay candidate —
which is itself the misreading in action (day-one zeros mean nothing;
the counts need sessions of real use before ablation is informed).

Gate-can-fail (previously first here) shipped 2026-08-31: the open
question resolved on the judgment/computation line — the skill still
performs the proof (unmechanized on purpose), `rein attest
gate-can-fail` journals it with the gate's hash, and doctor audits
proof currency without executing anything (GATE_UNPROVEN /
GATE_PROOF_STALE, silent when GATE_STUB owns the surface). This
repo's gate is attested; the proof was a real injected gofmt
breakage, not a ceremony.

`rein journal` (previously first in this list) shipped 2026-08-31:
read path with filters and a surfaces table, spec/journal.md v0.3,
cli-envelope v0.3 (fix rescoped to error/warning, mechanically
enforced), rein-diagnose 0.3 driving the journal both directions.
Deliberately not included: semantic recurrence over note text (a
procedure's judgment) and component-level aggregation (mechanical
and legitimate, but no consumer asks for it yet — add when one
does).

## In flight

Nothing.

## Field test 2 — CONCLUDED (2026-08-31, owner-run)

The owner ran the full setup on the field-test-1 repo (a private
React/Nx monorepo, ~370-line pre-existing instruction file, a
competing agent kit installed, real debt). Read-only analysis of the
artifacts; everything below is generalized per the privacy rule.

**The product held.** Converged end-state: doctor 0 findings, plan
empty, 9 managed files. Triage: ~210 of ~370 lines dropped (content
duplicating a routed docs/ tree, aliases derivable from tsconfig,
references to a deleted app), 5 rules demoted into the gate, ~97
kept — render at 23% of budget. The previously unmanaged instruction
file was fully adopted into the render. The competing kit's files
were staged for deletion — an ownership ruling executed, not just
recorded. The debt ledger finally held a real brownfield: 9 rows,
all evidence-backed (red lint/format repo-wide, empty test targets,
orphan specs never run, a typecheck that OOMs default heap, no CI,
grandfathered structure). Journal shows the designed loops firing in
the field: gate applied → attested → gate changed → re-attested
(GATE_PROOF_STALE recovery, second field catch); a wrong proposed
check overruled by the human on tree evidence (13/15 libs proved the
"violation" was the convention) with the amendment journaled; a
negative ruling recorded ("human declined all forbidden-surface
candidates — do not re-add").

**Still open from field test 2** (the consumed findings — bench 1,
4, 5, 6; real 9–13, 15; owner 9–11 — shipped 2026-08-31 with their
fixture cases in `field_test.go`; full text in git history):

- **[bench 2, remainder]**: `monorepo: false` on a plain
  angular.json multi-project layout *without* nx.json. The nx.json
  marker shipped; multi-project-angular-as-monorepo would need
  content parsing and a deliberate position on what "monorepo"
  means there. Detection-table material for the next real run that
  hits it.
- **[bench 3, 7]**: duplicate share dominated by per-app scaffold
  configs (tsconfig ×12, test-setup ×12) — whether the measure
  should distinguish scaffold-identical config from copy-pasted
  source, and/or add a group-by-basename note, is a ruling for
  after more repos.
- **[owner 12]**: the pre-existing AGENTS.md became a hand-authored
  8-line pointer at the rendered CLAUDE.md — unmanaged and
  undocumented; whether the adapter should own a redirect stub for
  the other host's file is a real design question.

Day-one citations are all zero on the field repo, as expected; real
sessions accrue them. The staged kit deletion awaits the owner's
commit.

## Noted, not yet actioned

- A pre-rewrite backup mirror sits at
  ~/Work/.freerein-backup-prescrub.git (local-only by design; it
  contains the scrubbed names). Owner deletes it once confident in
  the rewritten history.

Found in the 2026-08-31 audit, recorded here rather than bundled into
an unrelated change:

- **Coverage cold spots: mostly paid down (2026-08-31 second pass).**
  Emit/emitHuman now tested behind an `EmitTo(io.Writer)` seam (exit
  codes, always-present branch fields, fix omission on the wire);
  Dump and Adapters tested; the flags-after-positionals loop
  extracted as `splitArgs` and tested (`engine/cmd/rein/main_test.go`).
  Remaining, deliberately: `detectHighTouch` (needs a fixture repo
  with real git history — build it when that code is next touched)
  and the two `main` wiring functions themselves.

## Assumptions / open questions

- The 80% NEAR_CONTEXT_BUDGET threshold is a chosen default, not a
  measured one; revisit when real repos hit it.
- `addresses` vocabulary is free-form; if overlap warnings prove noisy
  or vacuous, a registered vocabulary (like probes) is the likely fix.
- inspect's detection tables (manifests, test/lint/CI configs) are a
  curated list, not exhaustive; grow them from real setup runs, not
  speculation.
- The official registry is live but empty; what to publish first is a
  deliberate product decision (a spec-flow procedure is the obvious
  candidate), not something to bundle into infrastructure work. Until
  a first component is published, the archive-fetch path over live
  HTTPS rests on the httptest coverage, not a production run.
- Doctor-in-gate: settled, not yet landed — the answer and the
  landing ceremony are § Next item 2.

## Blocked / needs a human

- **Curation: the eval task set.** `rein eval` is live but
  `.rein/eval/tasks.jsonl` is empty — tasks are the repo's own work
  and the owner curates them (spec/eval.md rule 1); the engine
  refuses to invent any.
- **First registry publication** (§ Next item 3): the demand-signal
  mechanism exists; the choice of what ships first is the owner's.
- **Codex skill frontmatter dialect.** Still a declared degradation
  in `content/adapters/codex.yaml` (`user_invoked_flag: ""`,
  "invocation-governance flags unverified for this host"). Needs
  verified facts about the host before implementing — guessing here
  would ship a silent lie in a contract that exists to prevent them.
