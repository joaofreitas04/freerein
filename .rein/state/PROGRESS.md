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
   journal data. The first evolve proposal on THIS repo waits for
   that data (day-one zeros are the misreading, not evidence).
2. **SHIPPED 2026-08-31.** All field-demanded fixes from field
   test 2 landed, each with its fixture red before the fix: the
   detection-table entries, candidate derivation, plan-side
   EXISTS_UNMANAGED, flag-anywhere dispatch, and rein-setup 0.8's
   six patches. The field case corpus lives at
   `engine/internal/engine/field_test.go` and only grows; the
   consumed findings' full text is in git history.
3. **The paired eval** (`rein eval`) — designed 2026-08-31, awaiting
   ruling (`spec/eval.md` v0.1): engine keeps score, never runs a
   task; tasks from the repo's own work with mechanical done-checks;
   deterministic counterbalanced assignment (the operator asks,
   never picks); conditions read from the installed lock, not
   asserted; no blinding, stated plainly; aggregates carry
   misreadings; standard-vs-minimal as the default arms (§2.6).
   Implementation waits on the ruling.
4. **Field-report channel — ruled ACCEPTED 2026-08-31, and mostly
   SHIPPED the same day** (`spec/field-report.md` v0.2): `rein
   report` (producer-side assembly: lock facts, journal since-dates
   on shadows, component-prefixed citations, judgment fields as
   mandatory flags; written to `.rein/out/report/`, never
   transmitted — `off` is the only submission level) and
   rein-diagnose 0.5 (disposition read from the lock at attribution;
   shipped-component closes with a report, never a silent fork), and
   the `SHADOW_STANDING` doctor advisory — ruled threshold-free
   2026-08-31 (spec amended to v0.3): every standing shadow listed
   with its journal-read age, both readings attached, the human
   judges; a number can be earned once real ages exist. On THIS repo
   the advisory correctly fires once, for the adopted verify gate —
   expected, not a regression. Remaining: only the propose/auto
   submission transport, which waits for a real report to demand it.
5. **Setup deepening** — next up, and it needs scoping before code
   (see Blocked): inspect breadth (hooks, skills directories, state
   files, prior installs — [real 15]'s skills-dir half is consumed;
   the rest has no named paths yet, and detection tables grow from
   real runs, not speculation), symlink resolution, a
   wired-vs-declared cross-check (shape undecided: doctor advisory
   on foreign files in managed dirs vs inspect-side
   declared-vs-present diff); allow/block asked as proposals with
   affordance-backed feasibility verdicts instead of open
   questions; detected gaps matched to extension `addresses:` —
   which is also the honest answer to what the registry publishes
   first (a deliberate product decision, per Assumptions).

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
- Whether `rein doctor` clean belongs in the gate is unsettled. This
  repo is self-hosted, so a hand-edited CLAUDE.md is a real failure
  mode that only doctor catches — but `scripts/verify` does not run
  it, and CI deliberately runs nothing the gate does not. Putting
  doctor in CI alone would make CI stricter than the gate: verify
  green locally, red on push. If it should block, it belongs in
  `.rein/overrides/scripts/verify`, and that is a deliberate change to
  the definition of done, not a drive-by.

## Blocked / needs a human

- **Ruling: the paired eval design** (`spec/eval.md` v0.1, drafted
  2026-08-31). Accept, amend, or redirect; implementation waits.
- **Scoping: § Next item 5 (setup deepening).** Which inspect
  surfaces earn table rows without fresh field evidence (hooks?
  state files? prior-install markers — which?); the
  wired-vs-declared cross-check's shape (doctor advisory on foreign
  files in managed dirs vs inspect-side declared-vs-present diff);
  and what the registry publishes first.
- **Codex skill frontmatter dialect.** Still a declared degradation
  in `content/adapters/codex.yaml` (`user_invoked_flag: ""`,
  "invocation-governance flags unverified for this host"). Needs
  verified facts about the host before implementing — guessing here
  would ship a silent lie in a contract that exists to prevent them.
