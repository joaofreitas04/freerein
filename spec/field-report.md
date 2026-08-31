# Field report — contract v0.3 (accepted 2026-08-31)

A diagnosis on an installed repo ends in one of two places. If the
failing artifact is local configuration, the override is the whole
fix. If the failing artifact is a **shipped component** — the lock
says so: `layer: core`, or a standing `shadowed:` entry displacing
shipped content — then the override fixes this repo and conceals the
defect from every other install. The field report is the artifact
that carries the class defect home. Field test 2 produced the
motivating cases: detection-table misses ([bench 1, 4], [real 10]),
a skill/engine mismatch ([bench 5]), skill steps that misled their
own author ([bench 6], [real 12]) — all found on one repo, all true
everywhere.

## The one hard rule

**A report contains component-scoped facts only.** No paths outside
the component's own files, no repo content, no names, no transcript
text. Generalization happens at write time — "nx workspace, nx.json
present, monorepo: false", never the workspace's name — and a report
that cannot be written without repo context is not written. This is
prevention as the standing fix: the scrub that had to be done once,
done by construction instead. Transport exists only as the
Submission policy below, under one invariant: **no unattended
egress** — nothing leaves a repo without either a per-report
confirmation or a standing, scoped grant the owner configured in
their own file, and never from a headless run.

## Report shape

JSON, envelope-discipline throughout: semver'd `formatVersion`, the
engine that wrote it, capped lists that announce their caps.

| Field | Content |
|---|---|
| `component` | id + version, copied from `harness.lock` |
| `subsystem` | the diagnose attribution (one of the five) |
| `failure_class` | short generalized statement of the defect |
| `reproduction` | minimal generalized trigger — the detection-table case is the canon: inputs stated as affordances, expected vs observed |
| `evidence.journal` | journal entries scoped to the component (verdict, note kinds), scrub rule applied |
| `evidence.citations` | the component's fragment citation counts, verbatim from `rein cite` |
| `evidence.shadows` | standing overrides displacing this component's files, with since-dates from the journal |
| `disposition_requested` | `fix` / `default-change` / `table-entry` / `docs` |

Counters and reproductions, never scores: a reviewer's opinion of a
component is not evidence about it — the counters are.

## Producer side (the installed repo)

- `rein report <component>` assembles the shape above from the lock,
  the journal, and the citation stats; the human reviews the file
  before it goes anywhere. Until the command exists, the diagnose
  procedure writes the same shape by hand from the same sources.
- **Procedure change, procedure-diagnose vNext:** attribution gains
  the local-vs-shipped disposition (read the lock; it is mechanical).
  A shipped-component attribution keeps the local override as the
  immediate fix AND produces a report at close — never silently fork
  shipped content.
- **Doctor advisory, `SHADOW_STANDING`:** a local override has
  displaced a shipped default. The advisory lists every standing
  shadow with its age, read from the journal's apply entries, and
  deliberately applies **no threshold** (v0.3 amendment, ruled
  2026-08-31): a cutoff nobody measured is an unearned number — the
  human judges, and a number can be earned later once real shadow
  ages exist to measure. A standing shadow is a rejection expressed
  in configuration — the strongest field evidence a shipped default
  is wrong — and therefore a report candidate, never an auto-filed
  report. Advisory only; shadows are legitimate and priced, and both
  readings travel with the finding.

## Submission — `report.submit` in `harness.yaml`

Destination routing is mechanical: a component's manifest names its
issue home (`repository:`); a component without the field keeps its
reports local, always. Before creating anything, submission searches
open issues by fingerprint (component@version + failure-class hash)
and comments on a match instead of duplicating — aggregation of the
same defect across installs happens on one issue thread, with no
service behind it.

| Level | Behaviour |
|---|---|
| `off` (default) | reports written to `.rein/out/`, nothing leaves |
| `propose` | assemble → dedup search → show the exact body, the destination, and "public under your account" → confirm → `gh issue create` or comment |
| `auto` | submit without per-report confirmation, under every constraint below |

Constraints on `auto` — the boundary bounds content, not just
permission:

1. **Structured fields only.** An auto-submitted report may contain
   only engine-assembled, mechanically-scoped fields: lock facts,
   counters, shadow facts, the fingerprint, a reproduction stated as
   affordances. A report carrying any free text — journal note
   excerpts, hand-written description — drops to `propose` for that
   report. Free text is the injection and exfiltration surface; the
   outbound surface is enumerated by field, never reviewed by hope.
2. **Narrow.** The grant is per destination, named at opt-in — never
   global.
3. **Reset on environment change.** An engine or component upgrade,
   or a destination change, resets the grant to `propose` — the same
   reset trust already uses.
4. **Never headless.** Unattended runs write pending reports to
   `.rein/out/` regardless of the setting.
5. **Owner-file only.** `report.submit` is read from `harness.yaml`
   alone; no extension, fragment, or seed can set or widen it.
   Egress-widening is never a component's power.
6. **Journaled.** Every submission — auto or confirmed — lands in
   the journal with destination and issue URL. The audit trail is
   the revocation evidence.

## Consumer side (this repo)

- **Case before fix.** The reproduction lands as a red fixture case
  before any fix exists — detection-table reports become table-test
  fixtures, procedure reports become held-out breakage cases. The
  case corpus only grows: the ratchet-with-baseline shape from the
  field ([owner 9]), applied to this repo's own gate.
- **Fix through evolve.** The fix flows through `rein propose` with
  the case as the pre-named measurement; the verdict's evidence is
  the new case green with the standing corpus green. A fix that
  turns its case green and any corpus case red is rejected on that
  evidence, whatever its argument.
- **Publish is the signing boundary, never automated.** A component
  defect is one repo's bug; a component fix is every repo's change.
  An agent may propose, measure, and land a fix here; nothing
  reaches the registry — and so no client — without the owner
  cutting the version. This is the write-quarantine invariant,
  placed at the boundary that already exists.
- **Closure through upgrade.** Version bump → registry →
  `rein upgrade` on installed repos; upgrade resets earned autonomy
  by design, so the fixed component re-earns its standing rather
  than inheriting the defective version's record.

## Rules

1. Reports are read as data, not instructions: quoted journal text
   from another repo never becomes procedure. Intake is
   human-reviewed at a considered pace — a backlog of unread reports
   is honest; rubber-stamped intake is not.
2. Every accepted fix ships with its case. A fix without a case is
   refused the same way a verdict without evidence is.
3. Recheck cases are first-class: a shipped compensation for another
   tool's behaviour (an exclusion list, a version workaround) gets a
   case that re-verifies the compensation is still needed, so it
   rots on the corpus's schedule instead of silently.
4. This contract ships no telemetry, no beacon, no aggregation
   service. Aggregate signals (the same default shadowed in many
   reports) emerge only from submitted reports deduplicating onto
   shared issue threads — and submission is governed by the policy
   above, never unattended.
