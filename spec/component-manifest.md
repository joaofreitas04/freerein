# Component manifest — contract v0.1 (draft)

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
4. `requires` entries must come from the engine's probe vocabulary
   (`rein probes` lists it); unknown requirements are errors, so the
   vocabulary and the ecosystem grow in lockstep.
5. `description` must contain a "Do NOT use for" clause naming its
   nearest neighbours; the registry checks those names resolve.
6. Every `seeds` entry must also appear in `provides`.
