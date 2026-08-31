# Citation — contract v0.1 (draft)

Fragment citation telemetry: the mechanics for `docs/lifecycle.md`
§1.2, whose positions this contract inherits wholesale — consumption
evidence beats content review, counts are advisory, zero ranks a
*decay candidate* and never triggers removal, and every number ships
with its misreading.

## Rules

1. **A fragment's id is `component:filename`** — the providing
   component's name (version-free, so counts survive upgrades) and
   the fragment file's basename without extension:
   `instructions-base:00-base`. The render marker and the cite
   validator derive it from one function, so the vocabulary an agent
   sees and the one the engine accepts cannot disagree. Identity
   survives content edits by design: a reworded rule keeps its
   history, and that smear is accepted.
2. **The render carries the ids.** Each fragment in the rendered
   instruction file is preceded by `<!-- rein:fragment <id> -->`
   (spec/resolution.md). The marker bytes are part of the render and
   therefore part of plan's budget: the telemetry pays rent in the
   same ledger as everything else.
3. **`rein cite <id>` records; the store is not the journal.**
   Counters go to `.rein/stats/citations.json` — engine-written,
   committed, deterministic key order, `formatVersion`-stamped. A
   counter is not history: it is rewritten in place, which is exactly
   why it lives outside the append-only journal. An id not in the
   current composition is refused with the valid vocabulary; rows for
   fragments no longer composed are kept (counts survive a
   remove/re-add) and omitted from reports. The instruction file
   itself stays byte-stable as counts change.
4. **`rein cite` (no argument) reads back**: every composed fragment
   with count and last-cited stamp, lowest first; zero-count
   fragments listed as `decay_candidates`; and the misreading in the
   result itself — zero can mean a dead rule *or* a fully
   internalized one, so candidates are ranked for human ablation,
   never swept. An engine that emits a number without its
   characteristic lie invites the wrong conclusion (lifecycle §1.5).
5. **The convention is instruction-rung and says so.** One standing
   rule in `instructions-base` asks agents to cite a rule that
   changed what they did — once, and never to be polite. Citation
   compliance is itself imperfect evidence: hosts offer no verified
   hook to automate it, so under-counting is expected and named
   here rather than corrected for silently.

Deliberately out: automatic removal (evolve-stage, human-ruled, needs
ablation); per-rule ids (only if real data demands them); hook-based
auto-citation (blocked on verified host facts).
