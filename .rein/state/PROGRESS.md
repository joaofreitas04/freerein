# Progress

Managed by the working agent, structured for the next session's cold
start. Update before ending every session. State is more than a task
list: record what you now believe and have not proven, or the next
session re-derives it — or worse, trusts it.

## Current state

Clean baseline as of 2026-08-31. `bash scripts/verify` GREEN,
`rein doctor` on this repo reports 0 findings over 5 managed files,
working tree committed through `52a98f3`. No TODO/FIXME markers in
the tree. The repo is self-hosted: CLAUDE.md is rendered by rein —
edit `.rein/overrides/`, never CLAUDE.md or scripts/verify directly,
then `rein apply --yes`.

What is shipped is described in `AGENTS.md` § Current state; do not
duplicate that inventory here. This file carries only what does not
survive in the code: what is next, what is unproven, what is blocked.

**Next, in `docs/lifecycle.md` §4 order** — observe ships before
evolve, and evolve's mechanics do not land until its substrates
exist:

1. **Use the loop on real work.** The full lifecycle now exists —
   init → operate → observe (journal, gate proof, citations, costs)
   → diagnose → evolve (propose/verdict) → upgrade — and every
   further mechanism should be demanded by a real run, not
   speculated. Field test 2 on a brownfield repo is the natural next
   move: it exercises rein-setup 0.7 end to end and starts real
   citation data accruing, which the first real evolve proposal
   needs. The first proposal on THIS repo waits for that data
   (day-one zeros are the misreading, not evidence).

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

Also unstarted and named in `AGENTS.md`: the minimal profile as a
shipped control mode; release automation.

## In flight

Nothing.

## Noted, not yet actioned

Found in the 2026-08-31 audit, recorded here rather than bundled into
an unrelated change:

- **Coverage, measured (2026-08-31, `-coverpkg=./...`): 74.0%
  overall.** The formerly "0.0%" packages (registry, resolve,
  component, adapter, lockfile) are in fact exercised through the
  engine tests. True cold spots, worth tests when touched next:
  `Dump` has no test at all; `envelope.Emit`/`emitHuman` are
  untested, so the exit-code contract (cli-envelope rule 6) rests on
  reading the code; `Adapters` 0%; both `main` functions 0% (the
  flags-after-positionals loop is real logic); `detectHighTouch` 18%
  (needs a fixture repo with actual git history).

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

- **Codex skill frontmatter dialect.** Still a declared degradation
  in `content/adapters/codex.yaml` (`user_invoked_flag: ""`,
  "invocation-governance flags unverified for this host"). Needs
  verified facts about the host before implementing — guessing here
  would ship a silent lie in a contract that exists to prevent them.
