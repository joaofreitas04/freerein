---
name: rein-diagnose
description: Diagnose and repair this repository's harness after an observed agent failure — attribute the failure to a subsystem, make the smallest artifact change that addresses it, apply through rein. Use when agents repeatedly fail the same way, ignore a rule, or the verify gate misfires. Do NOT use for debugging application code.
disable-model-invocation: true
---

# rein-diagnose — the harness repair loop

A harness failure is never "the model isn't good enough". Your job is
to name the missing capability and ship the smallest artifact that
supplies it. You drive `rein`; your write surfaces are `harness.yaml`
and `.rein/overrides/`.

## 1. Collect the evidence

- Ask the human for the failure: what was asked, what the agent did,
  what should have happened. Get the transcript excerpt or output if
  they have it.
- Run `rein doctor` and `bash scripts/verify`; note every finding.
- Run `rein dump` and read `.rein/out/dump.json` — know what the
  harness currently says before judging what it failed to say.
- Run `rein journal --path <suspected surface>` and
  `rein journal --kind note` for this failure's history, and read the
  progress file. The surfaces table counts applied and conflicted
  separately — a surface that keeps conflicting is being fought over.
  A single occurrence can be accidental — say so and stop at a note
  unless the human insists. A recurrence after an earlier fix means
  the earlier attribution was wrong; diagnose the attribution, do not
  stack a second artifact on the first.

## 2. Attribute to a subsystem

Name exactly one primary bucket. Never write "the model failed" —
name the layer that let it:

| Failure smells like | Subsystem |
|---|---|
| wrong/ignored conventions, unclear task framing | instructions |
| couldn't discover or misused a capability | tools |
| broken setup, wrong versions, irreproducible runs | environment |
| cold starts, lost progress, repeated work across sessions | state |
| claimed done without proof, gate green on broken work | feedback |

## 3. Choose the smallest artifact

Prefer, in order (cheapest binding that actually holds):

1. A **verify check** in `.rein/overrides/scripts/verify` — if the
   failure is detectable deterministically, detect it there; error
   messages must teach the fix.
2. A **short fragment** in `.rein/overrides/AGENTS.md.d/` — only for
   rules a check cannot express; one rule, stated once, with its
   trigger.
3. An **extension** (`rein add` / local path) when the fix is a real
   capability, not a rule.

Never solve a reliability problem by dumping more prose into the
instruction file — every added rule costs compliance on the others.
And never ship a second artifact for a failure surface an installed
component already addresses (`rein plan` warns `ADDRESSES_OVERLAP`):
strengthen or replace the existing one — stacked fixes for one
failure add cost faster than effect.

## 4. Apply and predict

- `rein plan`, show the human, `rein apply --yes` on approval.
- State the falsifiable prediction to the human: "next time X is
  attempted, Y should happen instead of Z." If the failure recurs
  anyway, the attribution was wrong — return to step 2, do not
  stack another artifact on top.

## 5. Close

- `rein doctor` clean, verify green.
- Record the ruling where the next diagnosis will look:
  `rein note "diagnose: <failure> -> <subsystem>; shipped <artifact>;
  predicted <Y instead of Z>; rejected: <candidates, why>"`. The
  rejected candidates are what stop the next diagnosis from
  re-walking this one — and step 1's recurrence check reads the
  journal, so a ruling recorded only in the progress file is
  invisible to it.
- Summarize in `.rein/state/PROGRESS.md` (current state only): the
  artifact now installed and the prediction now standing.
