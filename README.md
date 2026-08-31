# FreeRein

**Give your agents free rein. Keep the reins.**

FreeRein is a harness lifecycle manager: it installs, composes, and —
above all — *maintains* the agent harness of any repository, so AI
coding agents can work autonomously while humans steer.

It is not an installer. Generating an AGENTS.md is a commodity; the
product is the loop around it:

```
init → operate → observe → diagnose → evolve → upgrade
```

## Shape

- **`rein`** — a single static binary (Go). Agent-facing by contract:
  one JSON envelope per invocation, diagnostics that carry the fix,
  offline and deterministic. It owns composition: resolve → plan →
  apply over an ordered layer stack, recorded in a lockfile — and
  history: every completed change appends to a committed, append-only
  journal, because a lifecycle manager without a history has nothing
  to diagnose against.
- **Judgment procedures** — skills the engine installs into the host
  (Claude Code, Codex). They own what a CLI cannot: what the facts
  mean, the interview, design, failure diagnosis. Detection itself is
  the engine's (`rein inspect` surveys a repo without executing any of
  its code); the procedures drive `rein` as hands.
- **The installed harness** — the repo's actual guides and sensors:
  AGENTS.md, verification gates, hooks, state files, minted commands.

## Composition model

Opinionated core covering the five harness subsystems (instructions,
tools, environment, state, feedback — verification first), extended by:

| Kind | Adds | Cannot |
|---|---|---|
| **Extension** | new capability (commands, templates, gates) | — |
| **Preset** | how existing artifacts render | add capability |
| **Bundle** | a pinned, versioned stack of both | — |
| **Policy slots** | team-filled principles, graded like built-ins | — |

Resolution: project overrides → presets → extensions → core. First
match per path, whole-file, no deep merge. `rein dump` prints the
resolved composition; `rein plan` shows exactly what `apply` would do;
removal is refcounted; everything works offline.

Every component declares its coordinates: subsystem, enforcement rung
(instruction / conditional / permission / hook / isolation), rent
class — **compensation** (patches a model deficiency; re-examined per
model release) or **amplifier** (durable) — and the failure surfaces
it **addresses**. `rein doctor` uses them: drift vs lockfile, stale
compensations across every layer, overlapping fixes for one failure,
host degradations. `rein plan` prices what it proposes: the rendered
instruction file is a budget the agent pays on every turn, so growth
is a decision, never an accident.

## Status

Design phase. Start at `docs/design.md`, then `docs/lifecycle.md`
(the observe/evolve positions) and `docs/fitness.md` (where FreeRein
does **not** apply, and the discipline behind every number we
publish); the public contracts live in `spec/`.
