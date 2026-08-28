<!-- rendered by rein; do not hand-edit — edit .rein/overrides/ and run `rein apply` -->

# Working in this repository

This repo carries a FreeRein-managed harness. The rules below are the
contract for any coding agent working here.

- **Done means verified.** Run `scripts/verify` before claiming any
  task complete; a failing check is the work, never something to
  weaken.
- **State survives you.** Read `.rein/state/PROGRESS.md` before
  starting; update it before ending a session (what changed, what is
  in flight, what is blocked).
- **Stay in scope.** One task at a time; unrelated fixes get noted in
  the progress file, not bundled into the change.
- Harness files (`CLAUDE.md`, `scripts/verify`, `.rein/`) are managed
  by `rein`; edit them through `.rein/overrides/` so changes survive
  upgrades.

## This repository

The full contract for working on FreeRein lives in the repo root:

@AGENTS.md

## Verification

`bash scripts/verify` green is the definition of done for every
change — run it before reporting completion and paste its outcome.
If it fails, the failure is the task. Never edit the gate to pass it;
project checks are added via `.rein/overrides/scripts/verify`.

## Session state

Start every session by reading `.rein/state/PROGRESS.md`. End every
session by updating it: what changed, what is in flight, what is
blocked and why. Keep it current-state-only — history lives in git.

Known-broken things live in `.rein/state/DEBT.md` — when you hit one,
check the ledger before diagnosing from scratch; when you fix one,
retire its entry in the same change.
