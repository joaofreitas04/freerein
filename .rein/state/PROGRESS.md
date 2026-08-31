# Progress

Managed by the working agent, structured for the next session's cold
start. Update before ending every session. State is more than a task
list: record what you now believe and have not proven, or the next
session re-derives it — or worse, trusts it.

## Current state

Clean baseline as of 2026-08-31. `bash scripts/verify` GREEN,
`rein doctor` on this repo reports 0 findings over 5 managed files,
working tree committed through `877cd1c`. No TODO/FIXME markers in
the tree. The repo is self-hosted: CLAUDE.md is rendered by rein —
edit `.rein/overrides/`, never CLAUDE.md or scripts/verify directly,
then `rein apply --yes`.

What is shipped is described in `AGENTS.md` § Current state; do not
duplicate that inventory here. This file carries only what does not
survive in the code: what is next, what is unproven, what is blocked.

**Next, in `docs/lifecycle.md` §4 order** — observe ships before
evolve, and evolve's mechanics do not land until its substrates
exist:

1. `rein journal` — an engine-side read path. The engine writes
   `.rein/journal.jsonl` but offers no way to read it; rein-diagnose
   currently instructs the agent to read the raw JSONL by hand. That
   is an unbounded append-only file entering context with no
   `OUTPUT_OFFLOADED` and no engine-computed recurrence or
   trajectory, though `spec/journal.md` rule 4 names both as the
   substrate readers want. Closest thing to a live contract gap.
2. Gate-can-fail as a standing check. `GATE_STUB` catches a *stub*
   verify; nothing checks that a *real* gate can still fail.
   Breakage injection lives only in rein-setup step 7 and is
   deliberately unmechanized (running project code from doctor would
   either execute arbitrary code or prove the wrong thing) — so the
   open question is what a truthful engine-side version even is.
3. Fragment citation telemetry — not started.
4. Cost surfaces beyond the instruction file — not started; plan
   prices the instruction composition only.

Also unstarted and named in `AGENTS.md`: the minimal profile as a
shipped control mode; release automation.

## In flight

Nothing.

## Noted, not yet actioned

Found in the 2026-08-31 audit, recorded here rather than bundled into
an unrelated change:

- **Coverage is concentrated and partly unmeasured.** `engine` 75.2%,
  `publish` 73.2%; `registry`, `resolve`, `component`, `adapter`,
  `envelope`, `lockfile` have no direct tests and report 0.0% because
  plain `-cover` does not count cross-package exercise. A
  `-coverpkg=./...` run would say what is actually untouched.
- **`AGENTS.md` § Current state is accreting as chronology.** It
  reads as a milestone log rather than a current-state description,
  and it is the first thing a cold agent reads. Compressing it is a
  deliberate edit someone should make on purpose, not a drive-by.

## Assumptions / open questions

- The 80% NEAR_CONTEXT_BUDGET threshold is a chosen default, not a
  measured one; revisit when real repos hit it.
- `addresses` vocabulary is free-form; if overlap warnings prove noisy
  or vacuous, a registered vocabulary (like probes) is the likely fix.
- inspect's detection tables (manifests, test/lint/CI configs) are a
  curated list, not exhaustive; grow them from real setup runs, not
  speculation.
- Whether `rein doctor` clean belongs in the gate is unsettled. This
  repo is self-hosted, so a hand-edited CLAUDE.md is a real failure
  mode that only doctor catches — but `scripts/verify` does not run
  it, and CI deliberately runs nothing the gate does not. Putting
  doctor in CI alone would make CI stricter than the gate: verify
  green locally, red on push. If it should block, it belongs in
  `.rein/overrides/scripts/verify`, and that is a deliberate change to
  the definition of done, not a drive-by.

## Blocked / needs a human

- **Registry hosting.** No default registry URL and no hosted
  `index.json` anywhere, so `rein add` is unusable without an
  explicit `--registry <url-or-path>`. `NO_REGISTRY` is the correct
  well-formed error, but the product surface stays closed until a
  hosting decision is made (GitHub Pages vs. releases).
- **Codex skill frontmatter dialect.** Still a declared degradation
  in `content/adapters/codex.yaml` (`user_invoked_flag: ""`,
  "invocation-governance flags unverified for this host"). Needs
  verified facts about the host before implementing — guessing here
  would ship a silent lie in a contract that exists to prevent them.
