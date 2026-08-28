# FreeRein (this repository)

The harness lifecycle manager. Product name **FreeRein**, binary
**`rein`**. Read `docs/design.md` first; the public contracts live in
`spec/` and are versioned like an API — changing one is a breaking
change with a version bump, never a silent edit.

## Ground rules

1. **The research stays out.** The knowledge bundle lives in the
   sibling private repo `freerein-knowledge` and is never cited,
   copied, or linked from this repo — this repo's history must stay
   publishable from commit zero. Design documents state positions in
   their own words, without referencing the research or naming rival
   products studied there.
2. **`spec/` is the source of truth.** Engine code, content, and
   adapters conform to the contracts; when implementation and spec
   disagree, the disagreement is the work item — fix one deliberately.
3. **Judgment vs computation.** Anything with a guarantee attached
   (idempotency, refcounting, determinism, offline) is engine code.
   Anything requiring judgment is a procedure (skill). A skill that
   writes harness files directly instead of driving `rein` is a bug.
4. **Agent-facing output.** Every engine command emits one JSON
   envelope per `spec/cli-envelope.md`; every diagnostic carries a
   `fix`. Failures stay well-formed.
5. **Declared degradations.** Host differences are surfaced at apply
   time per `spec/host-adapter.md`, never absorbed silently.

## Layout

| Need | Location |
|---|---|
| Founding design | `docs/design.md` |
| Public contracts | `spec/` |
| Engine (Go) | `engine/` |
| Core content, procedures, adapters | `content/` |
| Registry index generator | `registry/` |

## Current state

Design phase. Next milestone: the walking skeleton
(`docs/design.md` §10) — `rein init → plan → apply → dump → doctor`
on a toy repo, core only, Claude Code adapter only.
