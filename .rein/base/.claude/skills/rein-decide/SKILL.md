---
name: rein-decide
description: Record a settled decision as an immutable, numbered decision record — collect every mandatory field first, follow the repo's existing numbering convention, never edit an old record. Use when the user says a decision was made and should be written down. Do NOT use for open questions still being weighed, or for diagnosing harness failures.
disable-model-invocation: true
---

# rein-decide — record a decision

A decision record answers *why is it like this* for a reader arriving
months later. An editable record converges on describing the present,
which the code already does — so records are **immutable: superseded,
never edited**.

## 1. Collect before writing — refuse to produce with a hole

Every field below is mandatory. If one is missing, ask for it; never
fill a gap with a guess — a record with a fabricated rationale is
worse than no record.

| Field | Form |
|---|---|
| Title | a noun phrase naming the decision, not a question |
| Date | today, ISO |
| Status | `Accepted` (or `Proposed` if explicitly not yet settled) |
| Context | the forces: what made this decision necessary, in the state of the world at the time |
| Decision | what was decided, active voice |
| Consequences | what becomes easier, what becomes harder, what is now owed |

If the human cannot state the context or the consequences, the
decision is not settled — stop and say so; it belongs in
`.rein/state/PROGRESS.md` under assumptions, not here.

## 2. Discover the convention, never impose one

- Look for an existing record directory: `docs/adr/`,
  `docs/decisions/`, `.adr/`, `adr/`. If one exists, match its
  numbering and filename style exactly and take the next number.
- If none exists, propose `docs/decisions/0001-<slug>.md` and let the
  human choose the location before writing anything.

## 3. Write and close

- Write the record with the six fields as sections. Decision records
  are project documents, not managed harness files — write the file
  directly; `rein` does not own it.
- Superseding an older record: set the old one's status to
  `Superseded by NNNN` with a link — that status line is the **only**
  edit an existing record may ever receive.
- Run `rein note "decision NNNN: <title>"` so the harness journal
  carries the pointer.
