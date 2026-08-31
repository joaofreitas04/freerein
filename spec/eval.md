# Paired eval — contract v0.1 (accepted 2026-08-31)

The measurement instrument the back half of the loop is missing:
evolve verdicts and profile claims currently rest on citations plus
journal recurrence, which nominate and narrate but cannot decide. The
paired eval is a task-level differential between two named harness
conditions on this repo's own work — the comparison lifecycle §2.6
says every claim of benefit already is, made explicit and recorded.

## The one hard line

**The engine never runs a task.** Running work under a condition is
the operator's job — an agent session or a human doing real work.
The engine's share is everything mechanical around it: hold the task
set, assign the next run so the operator cannot cherry-pick, read
the active condition from the installed harness rather than trusting
a flag, record outcomes append-only, and compute the paired
comparison with its misreadings attached. Judgment executes;
computation keeps score.

## Objects

| Object | Shape |
|---|---|
| task | id, one-line work statement, `done_check` (a command whose exit code decides), source (where the task came from) |
| condition | a whole installed composition, named by profile or by a deliberate single-component variation; read from `harness.yaml` + lock at record time, never asserted |
| run | task id, condition, done_check exit, wall-clock, operator note (optional), engine + timestamp |
| pair | the same task completed under both conditions; the unit every aggregate counts in |

Stores: `.rein/eval/tasks.jsonl` (curated by the human, committed)
and `.rein/eval/runs.jsonl` (append-only, engine-written, committed —
journal discipline: a bad run is voided by a later entry, never
rewritten).

## Rules

1. **Tasks are the repo's own work.** The set is curated from real
   work items — debt rows, § Next items, recurring session shapes —
   never synthetic puzzles alone. A task without a runnable
   `done_check` is not a task yet; outcomes are what the check says,
   not what the operator felt.
2. **The operator asks, never picks.** `rein eval next` assigns
   task + condition deterministically from the run history,
   counterbalanced (which condition goes first alternates per task).
   Choosing your own arm is the eval's version of the proposer
   accepting.
3. **Conditions are installed, not claimed.** Recording a run reads
   the active profile/composition from the lock; switching arms flows
   through the existing plan/apply + journal path. A run recorded
   under a condition the lock does not show is refused.
4. **No blinding, said plainly.** An agent reads its own instruction
   file; the condition is visible to the operator by construction.
   The compensation is mechanical outcomes (rule 1) and fixed
   assignment (rule 2), and the limit is stated in every report
   rather than papered over.
5. **Aggregates carry their misreadings** (lifecycle §1.5). Every
   `rein eval` read reports N pairs beside every rate; names that
   small N decides nothing alone; that ties are information (the
   harness may not bind on that task class); and that runs by
   different operators or models are different experiments — the
   report groups by operator label instead of pooling silently.
6. **The floor is the control arm.** The default comparison is
   `standard` vs `minimal` (lifecycle §2.6). Per-component claims
   require a condition that varies exactly that component; consumption
   evidence only nominates which component earns the experiment.
7. **Verdicts may name the eval.** An evolve proposal can pre-name
   "pair delta over tasks T under conditions A/B" as its measurement;
   the verdict then copies the eval read in as evidence. The eval
   never auto-closes a proposal — it is an instrument, not a judge.

## CLI sketch

`rein eval next` (assign) · `rein eval record --task <id> --exit <n>`
(append a run under the installed condition) · `rein eval` (read:
pairs, deltas, misreadings; offloaded like every report). All three
offline, deterministic, envelope-disciplined.

## Deliberately out

No model API calls, no sandbox orchestration, no scoring rubrics, no
automatic condition switching. Each would move execution or judgment
into the engine; the contract keeps the engine on the bookkeeping
side of the line.
