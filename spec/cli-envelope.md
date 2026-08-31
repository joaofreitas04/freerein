# CLI envelope — contract v0.3 (draft)

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
   diagnostic, always. A command that could not complete leaves
   `result` null — never a partial shape an agent might mistake for
   output. A command that completed but *found* problems (`doctor`)
   keeps its result: `ok` reflects the findings, `result` reflects
   completion, and the two are independent by design.
2. **Every error and warning carries a `fix`** — the next command or
   edit that resolves it, phrased to be executed by the model reading
   it. An error message that only names the problem is a contract
   violation, and the rule is enforced mechanically (a source-level
   check fails the build on an empty fix). An **info** diagnostic
   carries a `fix` when a next step genuinely exists
   (`OUTPUT_OFFLOADED` names the file to read) and omits the key
   otherwise (rule 7): forcing filler like "no action needed" onto
   announcements such as `UP_TO_DATE` would train agents to skip the
   field on the diagnostics where it matters.
3. **`code` values are stable identifiers** (SCREAMING_SNAKE),
   versioned with this spec; agents may branch on them.
4. **Side-effectful commands are two-phase.** `apply`, `upgrade`,
   `remove` without `--yes` return `confirm_required` describing the
   pending change-set; the judgment procedure shows it to the human
   and re-invokes with `--yes`. The engine never prompts
   interactively — prompting is the host agent's job. Headless
   callers therefore never hang: a pending confirmation is a
   well-formed success (`ok: true`, exit 0) whose `confirm_required`
   is non-null, and nothing was applied. An agent must treat a
   non-null `confirm_required` as "not done", never as a result.
5. **Output is bounded.** Large results (full dumps, diffs) are
   written to files under `.rein/out/` and referenced by path in the
   envelope, so the agent chooses what enters its context. **A cut
   announces itself in the channel the caller already reads**: any
   command that offloads or truncates emits an `info` diagnostic
   (`OUTPUT_OFFLOADED`) naming the path and, when applicable, the
   limit that forced the cut. Silent truncation anywhere in the
   envelope is a contract violation.
6. **Exit codes are contractual.** `0` = `ok: true`, including
   envelopes that carry warnings or a pending `confirm_required`;
   `1` = `ok: false`; `2` = the invocation itself failed before a
   command was identified (unknown flags, no arguments) — usage goes
   to stderr and no envelope is emitted, the one legitimate
   envelope-free exit. Agents may branch on the exit code before
   parsing anything.
7. **Absent, not null.** Optional keys are omitted when they carry
   nothing. The two exceptions are structural and fixed: `result` is
   explicitly null on failure (rule 1) and `confirm_required` is
   explicitly null when no confirmation is pending (rule 4), so the
   two fields an agent must branch on are always present.

## Known inconsistencies

A contract with no defect register is either new or unexamined.
Discrepancies between this document and the engine are work items,
recorded here until fixed — never silently absorbed.

1. (none currently recorded — last checked against `engine/` at
   contract v0.3; the v0.2-era finding that nine call sites shipped
   empty `fix` strings — one of them an error — was resolved by this
   version's rule 2 rescope plus real fixes on the three sites that
   had a next step)
