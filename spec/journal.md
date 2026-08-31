# Journal — contract v0.5 (draft)

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

## Reading

`rein journal [--kind K] [--since D] [--path P] [--limit N]` reads the
history back out, because the alternative — an agent slurping an
unbounded append-only file into its context — is the read-side twin of
a skill writing harness files by hand. The result carries a
`formatVersion` matching this contract's version and the engine that
produced it; consumers may branch on it, and additions bump the minor.

5. **Ordering is file position, reversed** (newest first). Position is
   mechanical; timestamps are content — `JOURNAL_WRITE_FAILED` invites
   hand-appended repair entries, whose `at` may be absent, malformed,
   or out of order. `--since` therefore filters on `at` as content
   (RFC3339, or `YYYY-MM-DD` meaning UTC midnight), and entries
   without a parsable `at` are excluded from a `--since` query with
   the exclusion announced in `notes` — a silent cut anywhere is a
   contract violation (cli-envelope rule 5).
6. **Entries survive verbatim.** The reader decodes lines as raw
   objects, so kinds and fields the engine does not recognize pass
   through untouched — rule 3 held by construction. A malformed line
   is `JOURNAL_LINE_UNREADABLE` (warning, naming the count and first
   bad line) and never a refusal: valid entries are still returned,
   and the fix directs correction by *appending*, never by editing —
   append-only is a protected surface. An absent journal is
   `JOURNAL_ABSENT` (info): a fact, not a failure.
7. **The engine counts; it never judges.** The result's `surfaces`
   table aggregates `apply` entries per path — `applied` and
   `conflicted` counted separately (a surface that keeps conflicting
   is being fought over; one that is merely touched often is not),
   with first/last timestamps, sorted most-conflicted first. Only
   `apply` entries carry paths (`add` records a component, `upgrade`
   records refs). Whether two failures are *the same failure* is
   judgment and belongs to the diagnose procedure; the engine hands
   it the countable half.
8. **Reads offload like everything else.** The full filtered set is
   written to `.rein/out/journal.json` and announced
   (`OUTPUT_OFFLOADED`, naming the limit when `--limit` cut the
   inline list); `--limit 0` is a legal ask for counts and surfaces
   only. Reading never mutates the journal.

## Attestation: proofs the engine cannot perform

9. **`attest` records a judgment procedure's proof.** Some facts are
   provable only with judgment and side effects — the canonical one
   is *the gate can fail*: introduce a breakage, watch
   `scripts/verify` exit non-zero, revert. Mechanizing that would
   mean doctor executing arbitrary project code, or proving the wrong
   thing; so the procedure proves, and `rein attest gate-can-fail`
   records the fact with the installed gate's current
   `sha256` (`{"kind":"attest","subject":"gate-can-fail",
   "gate":"scripts/verify","gate_hash":"sha256:…"}`). Subjects are a
   registered vocabulary like probes, never free text. Doctor's half
   is then mechanical and execution-free: no attestation on a real
   (non-stub) gate is `GATE_UNPROVEN`; an attestation whose hash no
   longer matches the installed gate is `GATE_PROOF_STALE` — the
   proof described a gate that no longer exists. A stub gate raises
   neither: `GATE_STUB` already owns that surface, and stacked
   findings for one failure are the thing doctor exists to warn
   about. This is the gate-can-fail standing check of
   docs/lifecycle.md §4, split on the judgment/computation line.

## Evolution: proposals and verdicts

10. **A harness change is proposed before it is made.** The
    `proposal` kind carries six mandatory fields — `surface`,
    `change`, `prediction` (falsifiable: next time X, Y instead of
    Z), `measurement` (what will decide, named BEFORE the change
    exists: gate re-run, paired-vs-minimal, held-out case),
    `baseline`, `nominated_by` (a registered vocabulary: cite-decay,
    compensation-recheck, overlap, recurrence, human — evolve's
    inputs are observe's outputs) — and `rein propose` refuses on
    any hole, listing all of them. Ids are short content hashes
    salted with the proposal sequence, so re-proposing an identical
    change after a rejection mints a new id rather than resurrecting
    the old verdict.
11. **One open proposal at a time** (lifecycle §2.3). `rein propose`
    refuses while any proposal lacks a verdict, naming it; there is
    no force flag. `doctor` keeps open proposals visible
    (`PROPOSAL_OPEN`, info — open is normal mid-work, invisible is
    the failure mode).
12. **A verdict is evidence, never the argument** (lifecycle §2.1).
    The `verdict` kind carries the proposal id, `outcome`
    (accepted | rejected — anything murkier goes back to measurement
    or to the human), and non-empty `evidence`: the measurement's
    result. The engine copies the proposal's pre-named `measurement`
    into the verdict, so what decided and what was promised to
    decide cannot disagree. Verdicts are immutable (`ALREADY_DECIDED`
    — a wrong verdict is corrected by a new proposal), and rejected
    verdicts are kept forever: they are what stops the same change
    being proposed again next month. **Stated limit:** the engine
    enforces the shape of acceptance, not the identity of the
    accepter — that the proposer never accepts is procedure
    discipline (`rein-evolve`), named here rather than faked.

