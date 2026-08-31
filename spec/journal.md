# Journal — contract v0.2 (draft)

`.rein/journal.jsonl` is the append-only history of the installed
harness: one JSON object per line, engine-written, committed to the
target repo. The lockfile records the harness's current *state*; the
journal records how it got there — and a lifecycle manager without a
history has nothing to diagnose against, nothing to detect recurrence
with, and no way to tell a maturing harness from a churning one.

```json
{"at":"2026-08-31T12:00:00Z","engine":"rein 0.1.0","kind":"apply","applied":["AGENTS.md","scripts/verify"],"conflicts":[],"layers":["overrides","core"]}
{"at":"2026-08-31T12:05:00Z","engine":"rein 0.1.0","kind":"add","component":"spec-flow@0.3.0","sha":"…"}
```

## Rules

1. **Append-only, forever.** The engine appends on every completed
   state change — `apply` (with the paths written and any merge
   conflicts left behind), `add`, `remove`, `upgrade` — and never
   rewrites or deletes a line. A wrong entry is corrected by a later
   entry. History that can be rewritten is not history.
2. **Rejections are entries too.** An apply that left conflict
   artifacts records them; a change that was proposed and not taken
   is exactly the knowledge that stops it being proposed again. The
   journal keeps what the working tree reverts.
3. **`kind` is open.** Readers must preserve entries whose `kind`
   they do not recognize; future engine versions add kinds (decision
   records, acceptance verdicts) without a format break. One such
   kind exists now: `note` — free text a judgment procedure asks the
   engine to record via `rein note` (setup rulings, rejected gate
   candidates, recorded decisions). The engine still does the
   writing, so the append-only guarantee stays mechanical.
4. **The journal is a record, not a measurement.** It says what the
   engine did on this machine, not whether the harness works, and it
   never leaves the repo. Efficacy claims need a paired measurement;
   the journal is the substrate such measurements and the diagnose
   procedure read — most cheaply as recurrence ("this failure was
   attributed and patched twice before") and as trajectory (a harness
   whose rule-count growth is slowing is settling; one that gains
   artifacts every week is still compensating for something).
