# Component manifest — contract v0.2 (draft)

Every unit of content — core module, extension, preset — carries a
`component.yaml` (or frontmatter block for single-file components) the
engine validates and enforces. Unknown keys are rejected, not ignored.

```yaml
name: spec-flow                  # kebab-case, unique within its source
kind: extension                  # core | extension | preset | bundle
                                 # (core is reserved for engine-embedded components)
version: 0.3.0                   # semver
subsystem: feedback              # instructions | tools | environment | state | feedback
rung: hook                       # instruction | conditional | permission | hook | isolation
rent:
  class: amplifier               # compensation | amplifier
  expires: ""                    # REQUIRED iff class: compensation —
                                 # the re-check trigger, e.g. "per model release"
addresses: []                    # failure surfaces this component exists to
                                 # close, from a small shared vocabulary
                                 # (e.g. premature-completion, scope-sprawl,
                                 # cold-start, unverified-claims). Two loaded
                                 # components sharing an entry is a warning,
                                 # not an error: stacked fixes for one failure
                                 # compound cost faster than effect.
provides:                        # paths this component renders, repo-relative
  - AGENTS.md.d/spec-flow.md
  - .claude/skills/spec/SKILL.md
seeds: []                        # subset of provides: installed if absent,
                                 # agent-owned after — never drift-tracked,
                                 # never merged, restored if deleted
                                 # (for agent-written state files)
requires:                        # repo affordances; checked at plan time
  - test-runner                  # from the engine's probe vocabulary
conflicts: []                    # component names this cannot compose with
description: >                   # ≤ 1024 chars; used by search/info
  [What it does] + [Use when …] + [Do NOT use for …]
```

## Enforced rules

1. `kind: preset` components may only list `provides` paths that
   shadow paths some other layer already provides — a preset
   structurally cannot add capability.
2. `kind: bundle` components list only pinned dependencies
   (`name@version`), no `provides` of their own.
3. `rent.class: compensation` without `expires` is a validation error.
   The rent doctrine applies to **every resolved layer**, not just the
   core: `doctor` lists each loaded compensation with its re-check
   trigger, wherever it came from — a vendored extension's stale
   assumption is as expensive as an embedded one.
4. `requires` entries must come from the engine's probe vocabulary
   (`rein probes` lists it); unknown requirements are errors, so the
   vocabulary and the ecosystem grow in lockstep.
5. `description` must contain a "Do NOT use for" clause naming its
   nearest neighbours; the registry checks those names resolve.
6. Every `seeds` entry must also appear in `provides`.
7. `addresses` entries are advisory metadata with one mechanical
   consequence: when two loaded components share an entry, `plan` and
   `doctor` emit an `ADDRESSES_OVERLAP` warning naming both. Harness
   components stack sub-additively, and the usual mechanism is two
   artifacts pushing at the same failure; the overlap is where to
   look before adding a third. `conflicts` stays the hard form —
   overlap warns, conflict refuses.

## Seed disposal

Seeds invert the ownership of every other provided path, and their
lifecycle follows: after install the file is the agent's, so `remove`
and `upgrade` leave a seeded file in place — un-installing a component
must never delete state an agent wrote. `remove` reports each
left-behind seed by path so the disposition is visible, not silent.
A seed exists because of the agent that writes it; a cleanup pass that
cannot see that will delete it as dead weight, which is why seed
templates should say what they are in their own opening lines.
