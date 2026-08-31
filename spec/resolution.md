# Resolution — contract v0.4 (draft)

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
   file — fragments are separate paths, so the rule holds. Each
   fragment renders behind its citation marker,
   `<!-- rein:fragment <component:filename> -->`, the id contract of
   spec/citation.md; the marker bytes are part of the render and so
   part of rule 6's budget. Paths under `skills/` are likewise
   host-neutral and map onto the adapter's skills directory at render
   time.)
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
6. **`plan` prices the composition.** Instruction content is paid for
   out of a budget the host enforces and the agent pays on every
   turn, so the plan reports it before apply spends it: the plan
   result carries a `budget` object — rendered instruction-file
   bytes, the adapter's `max_bytes`, and the fragment count — and a
   `NEAR_CONTEXT_BUDGET` warning fires at 80% of the adapter limit,
   while the render itself refuses past 100%. What a fragment costs
   is not its length but its necessity; the number exists so that
   growth is a decision someone made, never an accident nobody saw.
   The same pricing covers the whole composition (lifecycle §1.4):
   the plan result's `costs` object tiers every installed artifact by
   when its price is paid — `always` (the instruction file, every
   turn), `per_session` (state seeds at their on-disk size, since the
   agent-owned file a session reads is the real cost, not the
   template), `conditional` (each skill's two loads: the description
   listed every session and the body paid on invocation) — names what
   it deliberately does not price (scripts are executed, never
   loaded; the vendor tree never enters context; no adapter exposes
   hooks yet), and carries its misreading: bytes price the load, not
   the worth. Reporting only — no thresholds beyond the instruction
   file's, because a cutoff nobody measured is an unearned number.
