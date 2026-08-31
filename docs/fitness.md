# FreeRein — fitness and non-goals

Status: draft, 2026-08-31. Companion to `docs/design.md`. Every
component FreeRein ships must declare a "Do NOT use for" clause; this
document is the same clause for the product. An installer whose pitch
is "any repository" is making a claim we decline to make — the honest
answer to "should I use this here?" is sometimes no, and a tool that
cannot say so cannot be trusted when it says yes.

## 1. Where FreeRein pays for itself

- A repository that agents will work in **repeatedly, over months** —
  the lifecycle loop only compounds across sessions. One session has
  nothing to observe, diagnose, or evolve.
- Teams that want **shared, versioned, reviewable** harness decisions
  instead of per-developer configuration drift.
- Brownfield repos with real debt — *provided* the goal is to measure
  and ledger the mess and install the paydown loop, not to be rescued
  from it on day one.
- Repos with at least some ambient affordances: a test suite that
  runs, a reproducible setup, module boundaries a check could hold.
  The harness amplifies what the repo affords; it cannot conjure
  affordances from nothing.

## 2. Where FreeRein is the wrong tool

- **Firefighting.** A production incident, a hotfix under time
  pressure. Ceremony added at the moment of least patience is
  ceremony that gets bypassed, and a bypassed harness teaches
  everyone the harness is optional.
- **Throwaway work.** Spikes, one-off scripts, prototypes you intend
  to delete. The loop's value is accrual; nothing accrues to a repo
  with no future. Write the script, delete the script.
- **Repos where nobody can state the rules.** The setup interview
  derives proposals from evidence and asks the human to rule. Where
  no one can rule — no owner, no conventions, no definition of done
  anyone will defend — the harness records the ambiguity, it does not
  resolve it. Fix ownership first.
- **As a substitute for a vendor harness.** FreeRein authors the
  artifacts hosts read; it is not an agent runtime, not a sandbox,
  and not a scheduler. A repo gets the isolation its host provides,
  not an isolation FreeRein adds.

## 3. Non-goals, stated so they stay visible

1. **We do not detect whether the human is steering or surrendering.**
   Every artifact we install — gates, state files, procedures — works
   identically for a human who reviews and a human who rubber-stamps.
   No sensor we could ship distinguishes them, so we say it here
   instead of pretending: the harness relocates your judgment, it
   cannot replace it, and it cannot tell you if you stop supplying
   it.
2. **We do not install fluency.** Operating agents well is a skill
   that takes real practice on real work. FreeRein shortens the setup
   from days to an hour; it does not shorten the months.
3. **We do not solve functional correctness.** Deterministic gates
   catch what they test; nothing we ship proves the software does
   what the human meant, and a green gate over a wrong spec is a
   wrong program, verified. This is the standing open problem of the
   whole field; we ship the state of the art and label it as such.
4. **We do not promise the content lasts.** The engine's guarantees
   are durable; the shipped content encodes assumptions about what
   models can't yet do, and those assumptions expire. That is why
   every component carries a rent class and `doctor` nags about
   compensations — the product is built to shrink its own content
   over time, and a release where the core got smaller is a good
   release.

## 4. Claims discipline

Rules for every number FreeRein publishes about itself, because a
lifecycle manager that measures other people's harnesses must survive
the same audit:

- **Effect sizes are measured or absent.** No "X% better" about any
  component without a paired measurement behind it; a ranking a
  reader would act on deserves evidence or silence.
- **Claims carry their class.** Measured on our repos, reported by a
  user, or inferred — labelled, so a reader knows which audit applies.
- **Counts carry their date.** Adapter tiers, component counts, host
  capabilities: all true-on-a-date facts that rot without looking
  changed. Published with the date, re-probed on release.
- **A retired capability stays visible.** An adapter demoted or a
  component deprecated keeps its row, with the cause — an absence
  that explains itself, instead of a silent deletion someone else
  re-discovers.
