# FreeRein — founding design document

Status: draft, 2026-08-28. This document consolidates the design
brainstorm that founded the project. Decisions here are starting
positions; each hard contract lives in `spec/` and is versioned there.

## 1. Thesis

An engineering team's job is shifting from writing code to designing
environments, specifying intent, and building feedback loops so coding
agents do reliable work. The harness — everything around the model —
is a first-order term in agent performance, and roughly half of real
agent development is harness work. Yet building a harness is an
ongoing engineering practice, not a one-time configuration: its
components encode assumptions about model inability that go stale with
every model release, and installed copies drift from upstream.

**FreeRein is therefore a harness lifecycle manager, not an
installer.** The defensible product is the loop: init → operate →
observe → diagnose → evolve → upgrade.

Positioning: FreeRein is the repo-side slice of "everything that isn't
the model." It never replaces a vendor harness (Claude Code, Codex); it
authors the artifacts every host already reads — AGENTS.md, skills,
hooks, settings, state files — and manages them over time.

## 2. Division of labor

Three parts, one dependency direction:

1. **Engine (`rein`)** — Go, single static binary. Computational,
   deterministic, offline. Owns composition: resolve, plan, apply,
   dump, doctor, upgrade, verify, registry. Agent-facing CLI contract
   (`spec/cli-envelope.md`): the primary caller is the host's agent,
   not a human.
2. **Judgment procedures** — skills the engine installs. Inferential.
   Own discovery, interview, design, audit, failure→artifact
   diagnosis. They drive the engine; they never write harness files
   directly.
3. **The installed harness** — the output: guides (feedforward) and
   sensors (feedback) in the target repo.

Rule of thumb: reason over ambiguity, compute with code. Anything with
guarantees attached (idempotency, refcounting, offline determinism) is
engine. Anything requiring judgment is a procedure.

## 3. Composition model

- Layered resolution: **project overrides → presets → extensions →
  core**. First match per path, whole-file replacement, no deep merge
  (the effective value must be readable in one place).
- Extension kinds split **by intent**: extensions add capability;
  presets only re-render existing artifacts and structurally cannot
  grow the command surface; bundles are pinned, versioned stacks;
  policy slots are team-filled principles graded like built-ins.
- **Four guarantees** (hard requirements, tested):
  1. `plan`/`info` shows exactly what `apply` would change, before it does.
  2. Installs are idempotent and confined to the target repo.
  3. `remove` is refcounted: it never breaks what another bundle needs,
     and restores the next-highest-priority version of a shadowed file.
  4. Every consume-side command works offline.
- **The composition prints its own fixed point**: `rein dump` emits
  the resolved harness as data; `harness.lock` records, per installed
  file: source layer, source version + sha, content hash at install,
  refcount. Plan/apply/doctor/upgrade all diff against it.

## 4. The component contract (`spec/component-manifest.md`)

Every unit of content declares machine-enforced metadata: `kind`,
`subsystem` (instructions | tools | environment | state | feedback),
`rung` (instruction | conditional | permission | hook | isolation),
`rent` (compensation | amplifier, compensations carrying an expiry
trigger), `provides`, `requires` (repo affordances). This buys:

- `doctor` lists compensations older than the current model generation
  as removal candidates — simplification as a command, not a discipline.
- `plan` prices each addition (context cost, rung).
- `requires` makes brownfield honesty mechanical: an extension that
  needs affordances the repo lacks fails at plan time, with the reason.

## 5. Host adapters (`spec/host-adapter.md`)

An adapter is **data, not code**: where the instruction file lives,
where skills go, which hook events exist, whether a post-compaction
re-injection point exists. v1 targets: **Claude Code and Codex** —
two hosts done honestly over fourteen done nominally. Degradations are
declared at apply time (e.g. "this host has no post-compaction hook;
the bootstrap will not survive compaction"), never silent.

## 6. The core (opinionated content)

Covers the five subsystems — missing any one is an incomplete
harness — with **verification first** (lowest investment, highest
return). Every rule the core ships declares its enforcement rung;
"where should this rule live?" is answered per rule, by us, so teams
don't have to. A **minimal profile** ships as a permanent control
condition, because harness components stack sub-additively and any
bundle's value must be checkable against a baseline.

## 7. Trust and registry

Sources are `install-allowed` or `discovery-only`; installs are
sha-pinned in the lockfile; version pinning is not review and is not
presented as such. The registry is a static JSON index (offline-cached)
with catalog-level checks no current registry runs — starting with
referential integrity of description cross-references. Everything
FreeRein installs into a repo is code someone runs by opening that
repo: the installed surface must be enumerable by `rein dump`.

Destructive or side-effectful operations (`apply`, `upgrade`,
`remove`) require explicit confirmation; judgment procedures are
user-invoked, never model-invoked. Humans steer at exactly these
checkpoints, and autonomy ramps via graduated trust rather than
being promised up front.

## 8. Honest constraints (kept in the pitch)

- The harness is most needed where it is hardest to build (brownfield).
  FreeRein ships a fitness posture, not universality: `requires`
  gating + start-with-sensors-the-repo-affords.
- "Humans only steer" has a relocating ceiling (sessions → tracker →
  review bandwidth). Team fluency takes months and cannot be installed.
- The behaviour harness (functional correctness) is unsolved by the
  whole field; we ship the state of the art and say so.

## 9. Repo layout

```
engine/            Go: cmd/rein + internal/{resolve,plan,render,lock,doctor,registry}
content/
  core/            the five-subsystem opinionated core (embedded at build)
  procedures/      judgment skills (init-interview, diagnose, audit)
  adapters/        claude-code.yaml, codex.yaml
  presets/ …       official presets/extensions/bundles
registry/          static index generator
spec/              the public contracts (versioned like an API)
docs/              this document and its successors
```

## 10. Walking skeleton (first milestone)

`rein init → plan → apply → dump → doctor` against one toy repo:
core only, no extensions, Claude Code adapter only. Smallest slice
exercising resolver, lockfile, and renderer end-to-end; proves the
plan/apply model before any content investment.

## 11. Naming

Product **FreeRein**, binary **`rein`**. npm `freerein`, crate
`freerein` free at check time (2026-08-28); `freerein.dev` /
`freerein.sh` unregistered; GitHub org `freerein` dormant-taken — use
`freerein-dev` or negotiate. Idioms are product vocabulary: *free
rein* (autonomy), *rein in* (graduated trust).
