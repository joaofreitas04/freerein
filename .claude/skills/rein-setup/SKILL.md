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

- Run `rein plan`. If `NOT_INITIALIZED`, run `rein init` (pick the
  adapter matching this host), then `rein plan` again.
- Run `rein dump` and read `.rein/out/dump.json` — know what you are
  composing before you compose it.

## 2. Discover before asking

Inspect the tree yourself — never ask the human something it answers:

- Stack, toolchain, candidate verify commands (manifests, lockfiles,
  CI config, Makefiles, package scripts).
- **Existing instruction files**: CLAUDE.md, AGENTS.md, .cursorrules,
  and anything similar. These get triaged in step 3, never ignored
  and never adopted wholesale.
- **Other config writers**: search for scripts or tools that WRITE
  agent config (installers appending to CLAUDE.md, MCP config
  generators, other harness tools). Each one found is an ownership
  conflict the human must rule on in step 4 — two writers on one
  file silently destroy each other's work.
- **Debt**: broken or never-run tests, pre-existing lint errors,
  version mismatches, gates CI doesn't enforce. Collect with
  evidence (counts, paths); it feeds the ledger, not the gate.

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

One round, concrete options, covering only what discovery and triage
could not settle:

- The triage table itself: show Drop/Demote/Flag decisions with
  evidence; the human rules on every Flag.
- Each config-writer conflict: who owns the file from now on —
  rein (the other tool must stop writing it) or the other tool
  (rein leaves that file unmanaged)?
- Which commands define "done", and which are too slow for every
  change?
- Anything agents must never touch.

## 5. Configure

- `.rein/overrides/scripts/verify` — the agreed commands plus every
  Demote rule as a real check; POSIX sh, `set -eu`, fails loudly.
- `.rein/overrides/AGENTS.md.d/10-project.md` — the Keep residue
  only.
- `.rein/state/DEBT.md` — one row per debt found in discovery:
  the debt, the evidence, the paydown trigger. Debt is ledgered,
  not gated: a gate red from birth trains agents to ignore it.

## 6. Apply with the human steering

- `rein plan`; show the change-set and every warning (host
  degradations, EXISTS_UNMANAGED) to the human.
- On approval, `rein apply --yes`.

## 7. Prove the gate

- `bash scripts/verify` — green is required.
- Prove it can fail: introduce a trivial breakage, confirm non-zero
  exit, revert. If you demoted rules in step 3, break one of those
  too — a demoted rule whose check can't fail wasn't demoted, it was
  dropped. Do not skip this step.

## 8. Hand off

- `rein doctor`; resolve or report every finding.
- Summarize: what is installed, what verify runs, what was dropped
  or demoted in triage and why, what the debt ledger holds, which
  ownership rulings were made, and that future edits to managed
  files belong in `.rein/overrides/`.
