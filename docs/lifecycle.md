# FreeRein — the back half of the loop: observe and evolve

Status: draft, 2026-08-31. Companion to `docs/design.md` §1. The
product is the loop — init → operate → observe → diagnose → evolve →
upgrade — and the front half is built. This document fixes the design
positions for the back half before its machinery lands, because these
are the stages where a lifecycle manager either becomes one or stays
an installer with a roadmap.

## 1. Observe: evidence before opinion

An installed harness emits almost no evidence about itself. A rule
can be absent from every prompt, a state file written every session
and read by no one, a gate green because it silently stopped
executing — and all three look exactly like health. The observe stage
exists to make those states distinguishable, with mechanisms that
stay inside the engine's constraints: offline, deterministic, local.

**Positions:**

1. **The journal is the substrate** (`spec/journal.md`, shipped).
   Append-only, engine-written, committed. Diagnosis needs
   recurrence, evolution needs history, and both need a record that
   survives the revert that undoes a change. A wrong entry is
   corrected by a later entry; nothing rewrites the past.
2. **Consumption evidence beats content review.** The cheapest
   telling signal about any installed artifact is the asymmetry
   between how often it is written and how often it is read. A
   progress file updated every session and never consulted, a skill
   never invoked, a fragment never cited — each is indistinguishable
   from working until something counts. Planned mechanism: rendered
   fragments carry stable IDs; a lightweight convention asks agents
   to cite the ID of a rule that shaped an action; counts accumulate
   in a separate stats file so the instruction file itself stays
   byte-stable. The counts are advisory and they rank *decay
   candidates*, not importance — a rule so internalized nobody cites
   it is the success case, which is why removal stays a human
   decision informed by a proper ablation, never an automatic sweep.
3. **A gate must be proven able to fail.** A gate that cannot fail
   and a gate that silently stopped running produce identical output.
   The setup procedure already breaks the tree once to prove the gate
   red; the observe stage makes that a standing check rather than a
   birth certificate.
4. **Cost is part of the picture.** `plan` now prices the rendered
   instruction file against the host's budget; the same discipline
   extends to everything installed: a component's cost is what the
   agent pays on every turn it is loaded, and a report that shows
   benefit without cost is half a report.
5. **Every metric ships with its misreading.** Numbers lie in
   characteristic directions, and a ratio needs bounds on both sides
   — a citation count falling can mean decay or mastery; a gate that
   never fires can mean quality or blindness. Wherever the engine
   emits a measurement, the adjacent text names the wrong conclusion
   a reader would jump to. A dashboard column that says how a number
   lies is cheaper than the postmortem that discovers it.
6. **A maturing harness slows down.** The journal makes one trajectory
   readable for free: a harness that needs a new rule every week is
   still compensating for something; one whose growth is flattening
   is settling. Direction, not a score.

## 2. Evolve: changes earn their keep

Evolution is where a harness manager can do the most damage: stacked
plausible improvements routinely make a harness worse, because
components interact, overlap, and go stale — and whoever proposes a
change is structurally the worst judge of what it breaks, since the
evidence that motivated it says nothing about what it endangers.

**Positions:**

1. **The proposer never accepts.** Whether the proposer is a human, a
   procedure, or a future automated loop: acceptance rests on a
   measurement the proposer does not control — the gate re-run, the
   paired check, the held-out case — not on the proposer's argument.
   An argument for a change is evidence it might help, never evidence
   it won't hurt.
2. **Rejections are kept.** A change that was tried and not taken is
   knowledge — the exact knowledge that stops the same change being
   proposed again next month. The journal records what happened;
   procedures record what was considered and declined, with reasons.
   A memory that keeps only successes re-litigates its failures
   forever.
3. **One change at a time, with a falsifiable prediction.** The
   diagnose procedure already ends with "next time X, Y instead of
   Z". Evolution generalizes it: every harness change states what it
   should fix, and a recurrence after the change means the
   attribution was wrong — return to diagnosis, do not stack.
4. **Overlap is the default failure of composition.** Two components
   addressing the same failure surface usually cost more together
   than the better one alone; the manifest now declares `addresses`
   and the engine warns on overlap. The bias this encodes: before
   adding, strengthen; before strengthening, remove.
5. **Upgrades re-check assumptions, not just versions.** Every
   compensation in the composition — whatever layer it came from — is
   listed by `doctor` with its re-check trigger. A model release is a
   standing invitation to remove components, and the removal is
   always a measured candidate list, never automatic: staleness
   boundaries are specific to each harness, and someone else's
   release notes cannot answer whether *this* repo still needs *this*
   crutch.
6. **The baseline is a permanent fixture.** Any claim that a
   composition helps is a comparison, and the comparison needs a
   floor: the minimal profile is not a starter tier, it is the
   control condition every richer profile must beat on the repo's own
   work. Whole-composition comparisons are valid; per-component
   attribution requires actually varying the component, which is why
   consumption evidence (observe §2) only nominates candidates and
   measurement decides.

## 3. The protected surface

An evolving harness needs parts that evolution cannot touch, chosen
precisely because they are what a shortcut-seeking optimizer — human
or agent — would weaken first: the verify gate's authority, the
lockfile, the base store, the journal's append-only rule. These are
frozen by design, and the freezing is both the safety story and a
stated limit: FreeRein evolves the harness *content* and refuses to
let the harness evolve its own *evaluator*. Where an agent authors
new harness components, they enter as proposals through the ordinary
plan/apply path with a human at the confirm step — never directly
into the composition they will then be judged by.

## 4. Sequencing

Observe ships before evolve, and within observe the order is:
journal (shipped) → gate-can-fail as a standing check → fragment
citation telemetry → cost surfaces beyond the instruction file.
Evolve's mechanics (paired measurement, held-out acceptance) build on
those substrates and do not land until they exist — an evolution loop
without observation is a random walk with confidence.
