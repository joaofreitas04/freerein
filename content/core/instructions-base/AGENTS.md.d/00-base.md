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
