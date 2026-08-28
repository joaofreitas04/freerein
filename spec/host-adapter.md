# Host adapter — contract v0.1 (draft)

An adapter is a declarative mapping, not a plugin. The renderer walks
the resolved component set through it to produce host-specific
artifacts. Community adapters are YAML files programming against this
schema.

```yaml
name: claude-code
version: 0.1.0
instruction_file:
  path: CLAUDE.md            # or AGENTS.md where the host reads it
  imports: true              # supports @path imports
  max_bytes: 32768
skills:
  dir: .claude/skills
  frontmatter: claude        # dialect for invocation-governance flags
  user_invoked_flag: "disable-model-invocation"
hooks:
  events: [PreToolUse, PostToolUse, SessionStart, PreCompact, Stop]
  post_compaction_reinjection: true   # load-bearing: see degradations
settings:
  path: .claude/settings.json
  permissions: true
state_dir: .rein/state
degradations: []             # filled by the engine at apply time for
                             # capabilities the resolved set needs and
                             # this adapter lacks
```

## Rules

1. **Degradations are declared, never silent.** If the resolved set
   needs a capability the adapter lacks (e.g. a post-compaction hook
   for the bootstrap), `apply` emits a warning diagnostic naming the
   consequence and the fallback used. A methodology dying quietly on
   one host is the failure this contract exists to prevent.
2. Adapters carry **no logic** — no templating, no conditionals. If a
   host needs behavior, that behavior lives in the engine, keyed on a
   declared capability field, so all adapters benefit.
3. v1 ships `claude-code` and `codex`. Adapter support is **tiered
   and published** (supported / community / unmaintained) — a
   maintenance ranking users can see, not a promise.
