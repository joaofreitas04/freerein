---
name: rein-evolve
description: Improve or shrink this repository's harness deliberately — nominate one candidate from the observe evidence, propose it through rein with a measurement named in advance, make the change, measure, and record the verdict on evidence. Use for removing stale components, strengthening weak rules, or testing whether a composition earns its keep. Do NOT use to repair an observed agent failure (rein-diagnose) or to record an unrelated decision (rein-decide).
disable-model-invocation: true
---

# rein-evolve — changes earn their keep

Stacked plausible improvements routinely make a harness worse. You
drive `rein`; your write surfaces are `harness.yaml` and
`.rein/overrides/`. One change at a time — the engine enforces it.

## 1. Nominate exactly one candidate

Candidates come from evidence, not taste:

- `rein cite` — zero-cited fragments are decay candidates (zero can
  also mean internalized: nominate, never assume).
- `rein doctor` — `COMPENSATION_RECHECK` after a model release;
  `ADDRESSES_OVERLAP` where two components fight one failure.
- `rein journal --kind verdict` — check first: a change rejected
  before needs new evidence to re-propose, not a fresh attempt at
  the same argument.

Bias, in order: before adding, strengthen; before strengthening,
remove.

## 2. Propose before touching anything

`rein propose --surface <artifact> --change "…" --prediction
"next time X, Y instead of Z" --measurement "…" --baseline "…"
--nominated-by <source>`

Name the measurement NOW, before the change exists: the gate re-run
for behavior claims, a paired run against `profile: minimal` for
composition claims, a held-out breakage case for gate claims. A
measurement chosen after the change is you accepting your own
proposal.

## 3. Change, measure, verdict

- Make the change through overrides / `rein`, apply, verify green.
- Run the pre-named measurement. Ambiguous result → the human rules;
  your argument is evidence it might help, never evidence it won't
  hurt.
- `rein verdict --proposal <id> --outcome accepted|rejected
  --evidence "<the measurement's result>"`.
- **Rejected**: revert the change; the verdict stays — it is what
  stops this proposal returning next month.
- **Accepted**: state the prediction in PROGRESS. A recurrence after
  acceptance means the attribution was wrong — that is a diagnosis,
  not a second evolve.
