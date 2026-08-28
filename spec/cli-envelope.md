# CLI envelope — contract v0.1 (draft)

The primary caller of `rein` is a coding agent. Every invocation —
success or failure — emits exactly one JSON document on stdout.
`--human` renders the same data as text for people; the JSON form is
the contract.

```json
{
  "ok": true,
  "command": "plan",
  "result": { "adds": [], "changes": [], "removes": [] },
  "diagnostics": [
    {
      "severity": "warning",
      "code": "HOST_NO_COMPACTION_HOOK",
      "message": "claude-code adapter: no post-compaction re-injection point on this host version",
      "fix": "pin the bootstrap in AGENTS.md instead: run `rein apply --bootstrap=instruction-file`"
    }
  ],
  "confirm_required": null
}
```

## Rules

1. **Failures stay well-formed.** A crash without an envelope is a bug
   with the same severity as data corruption. `ok: false` + a
   diagnostic, always; exit code mirrors `ok`.
2. **Every diagnostic carries a `fix`** — the next command or edit
   that resolves it, phrased to be executed by the model reading it.
   An error message that only names the problem is a contract
   violation.
3. **`code` values are stable identifiers** (SCREAMING_SNAKE),
   versioned with this spec; agents may branch on them.
4. **Side-effectful commands are two-phase.** `apply`, `upgrade`,
   `remove` without `--yes` return `confirm_required` describing the
   pending change-set; the judgment procedure shows it to the human
   and re-invokes with `--yes`. The engine never prompts
   interactively — prompting is the host agent's job.
5. **Output is bounded.** Large results (full dumps, diffs) are
   written to files under `.rein/out/` and referenced by path in the
   envelope, so the agent chooses what enters its context.
