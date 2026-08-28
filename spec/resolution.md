# Resolution — contract v0.1 (draft)

How `harness.yaml` (intent) becomes a resolved file set.

## Layer order (highest priority first)

1. **Project overrides** — `.rein/overrides/**`
2. **Presets** — in the order declared in `harness.yaml`
3. **Extensions** — in the order declared in `harness.yaml`
4. **Core** — embedded in the engine binary

Preset and extension entries are source strings: local paths (`./…`,
`../…`, `/…`) naming a component directory — the component-development
loop — or registry refs (`name@version`), resolved through the
committed vendor tree at `.rein/vendor/` and sha-pinned in the
lockfile (see spec/registry.md). A source's manifest `kind` must match
the list it is declared under.

## Rules

1. **First match per path wins. Whole-file. No deep merge, ever.**
   The effective content of any path is readable in exactly one
   place. (Instruction-file composition happens by rendering an
   `AGENTS.md.d/` fragment directory into the adapter's instruction
   file — fragments are separate paths, so the rule holds. Paths
   under `skills/` are likewise host-neutral and map onto the
   adapter's skills directory at render time.)
2. A later layer cannot prevent an earlier (higher-priority) layer
   from overriding it: distribution does not confer authority.
3. `rein dump` prints the resolved set — every path, its winning
   component, and what it shadowed. A composition mechanism must be
   able to print its own fixed point; "what is actually installed?"
   is a query, not an inference.
4. Resolution is pure: same `harness.yaml` + same sources → same
   resolved set, offline, no environment reads. Probes (`requires`
   checking) read the repo but never the network.
5. `plan` = resolved set ⋄ lockfile ⋄ working tree, three-way:
   adds (in resolution, not in lock), upgrades (both, source version
   differs), drift (lock and tree hashes differ — respected, merged,
   never clobbered), removals (in lock, no longer in resolution —
   refcount-gated).
