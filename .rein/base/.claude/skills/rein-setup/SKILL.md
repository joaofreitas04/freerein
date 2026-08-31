---
name: rein-setup
description: Configure this repository's FreeRein harness — discover the stack, triage any existing instruction files, interview the human, wire the verify gate, ledger the debt, and apply. Use when the user asks to set up, configure, or finish setting up the harness. Do NOT use for routine coding sessions.
disable-model-invocation: true
---

# rein-setup — the configuration interview

You are configuring this repository's harness by driving the `rein`
engine. You never write managed files directly: your write surfaces
are `harness.yaml`, `.rein/overrides/`, and the seed files under
`.rein/state/`; everything else goes through `rein apply`.

Every `rein` command returns one JSON envelope. Read `diagnostics[].fix`
on any failure — it names your next action.

## 1. Ground yourself

- Run `rein inspect` — discovery works before init, and it writes the
  fact report the rest of this procedure reads.
- Run `rein plan`. If `NOT_INITIALIZED`, run `rein init` (pick the
  adapter matching this host), then `rein plan` again.
- Run `rein dump` and read `.rein/out/dump.json` — know what you are
  composing before you compose it.

## 2. Discover before asking

Read `.rein/out/inspect.json`. The facts arrive detected — toolchain,
the size-and-language measure (every file counted or classified with
its reason: generated, duplicate, binary, oversize, unknown), test
candidates with provenance, CI configs, the instruction corpus,
shared config surfaces, high-touch files, docs tree, affordances —
and your job is what they mean. Never ask the human something the
report answers, and never re-derive by hand what it already holds.
On top of the report:

- **The gate proposal.** The report lists candidate checks and where
  each came from. **Run each candidate yourself** — the engine never
  executes project code; measuring is your job — and record outcome
  and wall-clock time. Produce a proposal table: check | command |
  evidence | runtime | recommendation (gate / slow-tier / ledger).
  A check that fails on the untouched tree is debt, not a gate
  candidate. You derive this table; the human only rules on it.
- **Evidence-check every proposed rule against the tree, here, before
  anything applies.** Count what the tree does today: a "violation"
  it commits pervasively is the convention, and the honest proposal
  is a Flag with file:line proof — not a check born red. A rule
  amended after apply means this step was skipped.
- **The instruction corpus** the report lists gets triaged in step 3
  — never ignored, never adopted wholesale.
- **Ownership conflicts**: the report's config surfaces are files
  more than one tool commonly writes. Find who writes each (other
  harness tools, MCP installers, scripts appending to CLAUDE.md);
  every writer found is a ruling for step 4 — two writers on one
  file silently destroy each other's work.
- **Debt**: every candidate that failed on the untouched tree, plus
  never-run tests, pre-existing lint errors, version mismatches,
  gates CI doesn't enforce. Collect with evidence (counts, paths);
  it feeds the ledger, not the gate.
- **Affordances**: if the report shows no test runner, no CI, or no
  docs tree, say so plainly in step 4 — which sensors are even
  buildable here is a fact the human sees before anything installs.
  The measure sharpens the same conversation: a large generated or
  duplicate share is entropy worth a debt-ledger row, not something
  to silently lint around.

## 3. Triage the existing instruction corpus

If instruction files exist, sort every block into exactly one bucket:

| Bucket | Goes where | Test |
|---|---|---|
| **Drop** | nowhere | Derivable from the tree: overviews, file maps, command lists a manifest already declares, conventions the code itself shows. Copying these costs context and buys nothing. |
| **Demote** | a check in `.rein/overrides/scripts/verify` | A rule a check can express (layering, import bans, done criteria). A check binds; prose only suggests. Its error message must teach the fix. |
| **Keep** | `.rein/overrides/AGENTS.md.d/10-project.md` | The non-derivable residue: team protocols, locale/language rules, tribal thresholds, exact commands with done-conditions. Runbook register — executable, not descriptive. |
| **Flag** | the step-4 review | Stale claims, rules contradicting the tree, anything you'd drop but a human wrote deliberately. Never silently discard a human's sentence — propose, with file:line evidence. |

The kept fragment should be **short**. If it exceeds ~120 lines,
you kept derivable material — re-triage.

## 4. Interview the human

One topic at a time, each approved separately — triage, then the
gate, then ownership, then forbidden surfaces. A single batched
approval converts review into a formality; four small rulings are
four real ones. **Every item is a ruling on a proposal you present
with evidence and a recommended default — never an open question.** If you
are about to ask something open-ended, you skipped discovery; go back
and derive the proposal first. What stays with the human is policy —
what blocks "done", what risk is acceptable — not facts.

- The triage table: Drop/Demote/Flag decisions with file:line
  evidence; the human rules on every Flag.
- The gate proposal table from step 2, with your recommended
  composition marked. The human approves, adjusts, or re-tiers —
  they should never have to name a command themselves.
- Each config-writer conflict: who owns the file from now on — rein
  (the other tool must stop writing it) or the other tool (rein
  leaves that file unmanaged)?
- Forbidden surfaces: propose candidates you found (deploy scripts,
  migration dirs, secrets, generated code), then ask what to add —
  the one item where an open follow-up is legitimate, because risk
  tolerance is theirs.

## 5. Configure

- `.rein/overrides/scripts/verify` — the agreed commands plus every
  Demote rule as a real check; POSIX sh, `set -eu`. **Output is a
  budget the agent pays**: a passing check prints one line at most; a
  failing check prints `ERROR` and the reason on one greppable line,
  and the error text teaches the fix — a check that only says no is
  half an instrument. Thousands of lines of passing output cost the
  agent the very context it needs to act on a failure. Start from
  this shape:

  ```sh
  #!/bin/sh
  # The definition of done for this repository.
  set -eu
  cd "$(dirname "$0")/.."

  # Environment first: activate what the checks assume, as teaching
  # errors — a real gate starts here, not at the first check.
  command -v <tool> >/dev/null || { echo "ERROR env: <tool> missing — <activation or install move>" >&2; exit 1; }

  echo "verify: <name>"
  <command> >/dev/null || { echo "ERROR <name>: <what failed, and the fix>" >&2; exit 1; }

  echo "verify: green"
  ```

  Three shapes real gates need:
  - **Environment before checks.** Toolchain activation and resource
    limits are part of the gate (a required runtime major, a heap
    size a big typecheck needs) — absorbed here with teaching
    errors, never left for every session to rediscover.
  - **Cheap first, and say why.** Order checks so a fast failure
    precedes a slow one — the agent pays the slowest green check on
    every red loop. Any check or setting that is not self-evident
    carries its evidence as a comment ("proven on this tree: OOMs
    default heap at 47s without this"), so a future session can tell
    load-bearing from cargo cult.
  - **The ratchet, for rules the tree already breaks.** A rule with
    existing violations cannot gate as-is (red from birth trains
    agents to ignore it) and dropping it loses the rule. Commit
    today's violations to a baseline file and have the check diff
    against it: old debt stays ledgered, new violations refuse.
    Give the baseline a DEBT.md row so shrinking it has a trigger.
- `.rein/overrides/AGENTS.md.d/10-project.md` — the Keep residue
  only. `rein plan` prices the rendered file; if it warns
  `NEAR_CONTEXT_BUDGET`, you kept derivable material — re-triage.
  Shape it as a runbook, not a README:

  ```markdown
  ## Boundaries
  | ✅ always | ⚠️ ask first | 🚫 never |
  |---|---|---|
  | <safe action> | <discretionary action> | <forbidden surface> |

  ## Commands
  <exact command> — <its done-condition>

  ## High-touch files
  <path> — <why it attracts unrelated changes> (from inspect's churn map)

  ## Change budget
  Keep one change under <N> lines; past that, land the smallest coherent stage first.
  ```

  Skip any section with nothing real in it — an empty scaffold reads
  as a decision nobody made. The ⚠️ middle tier is the load-bearing
  column: a two-valued rule forces every discretionary case into a
  silent yes or a blocking no.
- `.rein/state/DEBT.md` — one row per debt found in discovery:
  the debt, the evidence, the paydown trigger. Debt is ledgered,
  not gated: a gate red from birth trains agents to ignore it. Give
  each trigger an event or a date the project cannot miss — `rein
  doctor` re-raises any entry whose date has passed, so an exception
  cannot quietly become policy.

## 6. Apply with the human steering

- `rein plan`; show the change-set and every warning (host
  degradations, EXISTS_UNMANAGED) to the human.
- **EXISTS_UNMANAGED is the adoption decision, not an error.** Apply
  will leave the pre-existing file alone; resolving the collision is
  a ruling, executed like this: to keep the human's version, move it
  to `.rein/overrides/<path>` — it becomes the winning layer. For an
  instruction file you triaged in step 3, the residue already lives
  in overrides, so delete the old file and let the render land — git
  is the recovery if anyone regrets it. Never leave the collision
  standing: an unmanaged twin shadows every future upgrade.
- On approval, `rein apply --yes`.

## 7. Prove the gate

- `bash scripts/verify` — green is required.
- Prove it can fail: introduce a trivial breakage, watch the gate go
  red, revert. If you demoted rules in step 3, break one of those
  too — a demoted rule whose check can't fail wasn't demoted, it was
  dropped. Do not skip this step.
- **Read the failure before trusting it.** A green gate on your
  breakage means the injection missed the gate's scope, not that the
  gate works — a compile check bounded to an import graph never sees
  an unimported file. Confirm the broken file is inside what the
  check actually reads before concluding anything.
- Record the proof: `rein attest gate-can-fail` — **only after
  watching every breakage fail, and never in the same command chain
  as the proof.** A chained attest records the proof before its
  evidence exists, and on a gate that never changes again a false
  attest stands forever. The engine journals the attest with the
  gate's hash; from then on `rein doctor` audits that the proof
  still describes the installed gate (`GATE_UNPROVEN` /
  `GATE_PROOF_STALE`) without ever executing it. An unattested proof
  expires with your session.

## 8. Hand off

- `rein doctor`; resolve or report every finding.
- Record the rulings: `rein note "setup: <drop/demote/keep counts>;
  <rejected gate candidates, with why>; <ownership rulings>"`. The
  journal entry is what stops a future diagnosis from re-litigating
  today's decisions — rejected candidates especially.
- Summarize: what is installed, what verify runs, what was dropped
  or demoted in triage and why, what the debt ledger holds, which
  ownership rulings were made, and that future edits to managed
  files belong in `.rein/overrides/`.
