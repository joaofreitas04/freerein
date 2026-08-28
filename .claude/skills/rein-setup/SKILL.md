---
name: rein-setup
description: Configure this repository's FreeRein harness — discover the stack, interview the human, wire the verify gate, and apply. Use when the user asks to set up, configure, or finish setting up the harness. Do NOT use for routine coding sessions.
disable-model-invocation: true
---

# rein-setup — the configuration interview

You are configuring this repository's harness by driving the `rein`
engine. You never write managed files directly: your only write
surfaces are `harness.yaml` and `.rein/overrides/`; everything else
goes through `rein apply`.

Every `rein` command returns one JSON envelope. Read `diagnostics[].fix`
on any failure — it names your next action.

## 1. Ground yourself

- Run `rein plan`. If it reports `NOT_INITIALIZED`, run `rein init`
  (pick the adapter matching this host), then `rein plan` again.
- Run `rein dump` and read `.rein/out/dump.json` to see the resolved
  composition you are configuring.

## 2. Discover before asking

Inspect the repo yourself first — never ask the human something the
tree already answers:

- Stack and toolchain: manifests (package.json, go.mod, pyproject,
  Cargo.toml, …), lockfiles, CI config.
- Candidate verify commands: existing test/lint/build invocations in
  CI files, Makefiles, package scripts.
- Conventions worth encoding: existing docs, formatting configs.

## 3. Interview the human

Ask only what discovery could not settle, one round, concrete options:

- Which commands define "done" here (build? tests? lint? typecheck?),
  and are any too slow for every change?
- Anything agents must never touch (paths, services, credentials)?
- Any project conventions that are NOT derivable from the code and
  belong in the instruction file?

## 4. Configure through overrides

- Write `.rein/overrides/scripts/verify`: a POSIX sh script running
  the agreed commands, `set -eu`, failing loudly. This is the
  definition of done — it must be able to fail.
- If the interview produced non-derivable conventions, write them to
  `.rein/overrides/AGENTS.md.d/10-project.md` (a short fragment; the
  renderer folds it into the instruction file between the base and
  verification fragments).

## 5. Apply with the human steering

- Run `rein plan`; show the human the change-set and every warning
  (host degradations included).
- On their approval, run `rein apply --yes`.

## 6. Prove the gate

- Run `bash scripts/verify` and show the outcome. Green is required.
- Then prove it can fail: introduce a trivial breakage (or ask the
  human for a known-failing case), run verify, confirm it exits
  non-zero, revert. A gate that cannot fail is not a gate — do not
  skip this step.

## 7. Hand off

- Run `rein doctor`; resolve or report every finding.
- Summarize for the human: what is installed, what verify runs, what
  was deliberately left out, and that future edits to managed files
  belong in `.rein/overrides/`.
