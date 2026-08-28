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

## Commands

| Purpose | Command |
|---|---|
| Definition of done | `bash scripts/verify` (gofmt + vet + test + build) |
| Build the binary | `go build -o rein ./engine/cmd/rein` |

## Current state

Walking skeleton DONE, plus the first four milestones: three-way
merge over a committed `.rein/base/` store (clean merges apply,
conflicts leave the file alone with a markered artifact under
`.rein/out/merge/`); the codex adapter (AGENTS.md, declared
degradations); local-path extensions/presets with kind enforcement
and the preset-cannot-add rule live; and the first judgment
procedure, `rein-setup` (user-invoked skill, drives the engine,
writes only overrides). Registry slice DONE
(spec/registry.md): static-index client, `add`/`remove`/`info`/
`upgrade`, vendoring into `.rein/vendor/` with archive sha + lockfile
tree-hash pins, doctor tamper detection, upgrades and removals
flowing through plan/apply so local edits three-way merge. Second
judgment procedure: `rein-diagnose` (failure→subsystem→smallest
artifact). Session discipline stays ambient in `instructions-base` —
deliberately no operate skill (a skill restating standing fragments
would be a second truth). Publishing + dogfood DONE:
`rein-publish` (engine/internal/publish) generates deterministic
sha-pinned archives + index.json, refuses republishing changed
content without a version bump (packs to temp so a refusal never
corrupts a live archive); apply never clobbers pre-existing unmanaged
files (EXISTS_UNMANAGED + adoption via overrides); flags may follow
positionals. **This repo is self-hosted**: CLAUDE.md is rendered by
rein — edit `.rein/overrides/`, never CLAUDE.md or scripts/verify
directly, then `rein apply --yes`. Field test 1 (2026-08-28,
[a private repo]) produced: **seed files** (install-if-absent,
agent-owned, never drift-tracked — fixes the permanent DRIFT warning
on state files), the **debt ledger** seed (`.rein/state/DEBT.md` —
recorded, not gated), and **rein-setup v0.2**: triage of existing
instruction corpora (drop derivable / demote to sensor / keep runbook
residue / flag for human ruling) plus the config-writer ownership
inventory. Doctrine grounded in the knowledge repo's
install-time-curation synthesis. Covered by `engine_test.go`,
`milestones_test.go`, `registry_test.go`, `publish_test.go`. Next:
host the official registry index + default registry URL, codex skill
dialect (blocked on verified host facts), release automation.
