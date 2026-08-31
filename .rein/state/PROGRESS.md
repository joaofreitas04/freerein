# Progress

Managed by the working agent, structured for the next session's cold
start. Update before ending every session. State is more than a task
list: record what you now believe and have not proven, or the next
session re-derives it — or worse, trusts it.

## Current state

Clean baseline as of 2026-08-31. `bash scripts/verify` GREEN,
`rein doctor` on this repo reports 0 findings over 5 managed files,
working tree committed through `52a98f3`. No TODO/FIXME markers in
the tree. The repo is self-hosted: CLAUDE.md is rendered by rein —
edit `.rein/overrides/`, never CLAUDE.md or scripts/verify directly,
then `rein apply --yes`.

What is shipped is described in `AGENTS.md` § Current state; do not
duplicate that inventory here. This file carries only what does not
survive in the code: what is next, what is unproven, what is blocked.

**Next, in `docs/lifecycle.md` §4 order:**

1. **Use the loop on real work.** The lifecycle exists end to end —
   init → operate → observe → diagnose → evolve → upgrade — and
   every further mechanism should be demanded by a real run, not
   speculated. Field test 2 is concluded (below); what remains is
   sessions, on this repo and the field repo, accruing citation and
   journal data. The first evolve proposal on THIS repo waits for
   that data (day-one zeros are the misreading, not evidence).
2. **Field-demanded fixes from field test 2** — cheap, evidence all
   in this file: detection tables (nx.json as a monorepo marker
   [real 10]; run candidates derived from angular.json targets and
   package.json scripts [bench 1, real 10]; CONTEXT.md-class names
   in instruction_corpus [bench 4]; shared skills dirs in
   config_surfaces [real 15]); leading-flag dispatch fixed or the
   error taught [real 9]; rein-setup patches (attest wording
   [bench 6, real 12], the adoption move documented [real 11],
   EXISTS_UNMANAGED aligned between plan and apply [bench 5],
   proposed rules evidence-checked against the tree before apply
   [owner 11], verify skeleton: environment section first
   [real 13], ratchet-with-baseline [owner 9], evidence comments
   and cheap-first ordering [owner 10]).
3. **The paired eval** (`rein eval`): with/without differential over
   a task set. The minimal profile is the shipped control arm
   waiting for its experiment, and evolve verdicts need a
   measurement stronger than citations-plus-journal can supply.
4. **Field-report channel — designed 2026-08-31, awaiting ruling**
   (`spec/field-report.md` v0.2): local-vs-shipped attribution read
   from the lock, privacy-by-construction reports (component-scoped
   facts only), case-before-fix intake with a grow-only fixture
   corpus, publish as the never-automated signing boundary, upgrade
   as the trust reset. Submission via gh under `report.submit`
   (off / propose / auto) with no-unattended-egress as the
   invariant: auto is structured-fields-only, per-destination,
   reset-on-upgrade, never headless, owner-file only, journaled.
   The procedure changes it requires (diagnose disposition step,
   SHADOW_STANDING advisory, `rein report`) land only after the
   ruling.
5. **Setup deepening**: inspect breadth (hooks, skills directories,
   state files, prior installs — [real 15] is the evidence),
   symlink resolution, a wired-vs-declared cross-check; allow/block
   asked as proposals with affordance-backed feasibility verdicts
   instead of open questions; detected gaps matched to extension
   `addresses:` — which is also the honest answer to what the
   registry publishes first.

Deliberately not scheduled, each waiting on its trigger: the trust
ledger, per-model profiles, the visible-index budget check, and the
write quarantine — whose trigger is named: before any accepted
proposal auto-lands agent-authored content, the staging gate must
exist (until then, publish-signing covers the boundary).

Evolve mechanics shipped 2026-08-31 (journal v0.5, as designed and
ruled): `rein propose` (six mandatory fields, refuse-before-generate,
measurement named before the change, one open proposal at a time as
a hard refusal — id hashes salted with the proposal sequence after
the test caught same-second identical re-proposals resurrecting old
verdicts) / `rein verdict` (evidence-only, immutable, pre-named
measurement copied in) / doctor `PROPOSAL_OPEN` / procedure-evolve
0.1 (rein-evolve: nominate one candidate from observe evidence,
bias remove > strengthen > add). The engine enforces the shape of
acceptance; that the proposer never accepts is procedure
discipline — stated in the spec, not faked.

Minimal profile shipped 2026-08-31 (resolution v0.5): `profile:` in
harness.yaml selects core membership — standard (full) or minimal
(instructions-minimal 0.1 + verification-gate: the floor, 809 bytes
always-tier on a fresh install vs 2290 standard on this repo, zero
per-session and conditional costs). The control condition of
lifecycle §2.6, stated as such by init (`PROFILE_CONTROL`); switching
flows through plan/apply, leaves agent-owned seeds (SEED_LEFT), and
lands in the journal so a paired measurement can read when the
condition changed. Overrides/extensions apply under every profile:
the comparison varies the shipped core, holds project config
constant.

Cost surfaces shipped 2026-08-31 (resolution v0.4): plan's `costs`
object tiers every installed artifact by when its price is paid —
always (instruction file), per-session (seeds at on-disk size: this
repo's PROGRESS.md prices at ~4.9KB, more than twice the instruction
file, which is exactly the kind of fact the tier exists to surface),
conditional (each skill's description + body loads) — with unpriced
surfaces named and the misreading attached. Reporting only, no new
thresholds: a cutoff nobody measured is an unearned number.

Fragment citation telemetry shipped 2026-08-31 as designed and
ruled: spec/citation.md v0.1, resolution v0.3 markers, `rein cite`,
instructions-base 0.3. Live on this repo with three honest citations;
`local-overrides:10-project` is the sole day-one decay candidate —
which is itself the misreading in action (day-one zeros mean nothing;
the counts need sessions of real use before ablation is informed).

Gate-can-fail (previously first here) shipped 2026-08-31: the open
question resolved on the judgment/computation line — the skill still
performs the proof (unmechanized on purpose), `rein attest
gate-can-fail` journals it with the gate's hash, and doctor audits
proof currency without executing anything (GATE_UNPROVEN /
GATE_PROOF_STALE, silent when GATE_STUB owns the surface). This
repo's gate is attested; the proof was a real injected gofmt
breakage, not a ceremony.

`rein journal` (previously first in this list) shipped 2026-08-31:
read path with filters and a surfaces table, spec/journal.md v0.3,
cli-envelope v0.3 (fix rescoped to error/warning, mechanically
enforced), rein-diagnose 0.3 driving the journal both directions.
Deliberately not included: semantic recurrence over note text (a
procedure's judgment) and component-level aggregation (mechanical
and legitimate, but no consumer asks for it yet — add when one
does).

## In flight

Nothing.

## Field test 2 — CONCLUDED (2026-08-31, owner-run)

The owner ran the full setup on the field-test-1 repo (a private
React/Nx monorepo, ~370-line pre-existing instruction file, a
competing agent kit installed, real debt). Read-only analysis of the
artifacts; everything below is generalized per the privacy rule.

**The product held.** Converged end-state: doctor 0 findings, plan
empty, 9 managed files. Triage: ~210 of ~370 lines dropped (content
duplicating a routed docs/ tree, aliases derivable from tsconfig,
references to a deleted app), 5 rules demoted into the gate, ~97
kept — render at 23% of budget. The previously unmanaged instruction
file was fully adopted into the render. The competing kit's files
were staged for deletion — an ownership ruling executed, not just
recorded. The debt ledger finally held a real brownfield: 9 rows,
all evidence-backed (red lint/format repo-wide, empty test targets,
orphan specs never run, a typecheck that OOMs default heap, no CI,
grandfathered structure). Journal shows the designed loops firing in
the field: gate applied → attested → gate changed → re-attested
(GATE_PROOF_STALE recovery, second field catch); a wrong proposed
check overruled by the human on tree evidence (13/15 libs proved the
"violation" was the convention) with the amendment journaled; a
negative ruling recorded ("human declined all forbidden-surface
candidates — do not re-add").

**Patterns that emerged in the field, candidates to fold back into
rein-setup (9-12 continue the numbering above):**

9.  **The ratchet-with-baseline.** The setup grandfathered 75
    existing placement violations into a committed baseline file the
    gate diffs against: old debt ledgered, new violations refused.
    The skill's verify skeleton doesn't teach this shape; it should —
    it is the only honest way to gate a rule a brownfield already
    breaks 75 times.
10. **Evidence comments inside the gate.** The authored verify
    carries its own justification ("proven on this tree: OOM at 47s
    without this") and orders checks cheap-first before a 54s
    typecheck. Both are skill-worthy guidance.
11. **Amendment-after-apply ordering.** The wrong check shipped in
    the first apply and was corrected in a second after the human
    saw it. The interview's one-topic-at-a-time review happened, but
    evidence-checking a proposed rule against the tree (does the
    tree already violate it pervasively?) belongs in step 2, before
    anything applies.
12. **Host-file redirect for the non-adapter file.** The pre-existing
    AGENTS.md became a hand-authored 8-line pointer at the rendered
    CLAUDE.md. Reasonable — but it is unmanaged and undocumented;
    whether the adapter should own a redirect stub for the other
    host's file is a real design question.

Day-one citations are all zero there, as expected; real sessions
accrue them. The staged kit deletion awaits the owner's commit.

## Field test 2 — preliminary findings (engine-side, from the dry run)

Facts about `rein inspect` 0.2.0 against the real target tree, before
any interview; each is detection-table material ("grow them from real
setup runs"):

1. **`tests.candidates: null` while `jest.config.ts` is detected and
   the test-runner affordance is true.** The report names configs but
   derives no runnable command, so setup step 2's "run each
   candidate" gets zero engine help and the human ends up naming
   commands — which the skill explicitly says should never happen.
   Angular/Nx-style runners (commands live in angular.json project
   targets / package.json scripts, not a config file that IS the
   command) are outside the current tables.
2. **`monorepo: false` on a multi-app workspace** (apps/* + libs/*,
   pnpm). Detection presumably keys on workspace manifests
   (pnpm-workspace/lerna/nx.json) and misses plain angular.json
   multi-project layouts.
3. **Duplicate share is a quarter of the tree** (806 of 3153 files),
   dominated by per-app scaffold config: test-setup.ts ×12,
   tsconfig.json ×12, tsconfig.spec.json ×11, .eslintrc.json ×7.
   Honest signal (real config duplication), but scaffold-identical
   config and copy-pasted source land in one bucket; whether the
   measure should distinguish them is a ruling for after more repos.
4. **`instruction_corpus` found only CLAUDE.md** (1272 bytes) and
   missed root-level CONTEXT.md and product-named notes files — instruction-ish
   files a triage should at least see. The filename table is
   deliberately curated; CONTEXT.md is common enough to be a
   candidate.

Positive signals from the same run: inspect on 3164 files took 74ms
pre-init with a well-formed envelope and offload; classified/high-
touch caps announced themselves in notes; high-touch top hit
(an `apps/<app>/version.json`, ×75 commits) is a genuinely useful churn
fact.

From the completed bench pass, additional and confirming [bench]:

5. **Skill/engine mismatch on EXISTS_UNMANAGED**: rein-setup step 6
   says show the warning at plan time; the engine emits it at apply
   only (plan carries `unmanaged[]` silently in the result). Align
   one, deliberately. The guarantee itself held byte-for-byte: apply
   left the pre-existing CLAUDE.md alone and said so.
6. **Step 7 needs wording against same-chain attests.** The bench
   pass attested in the same command chain as an unproven breakage —
   the proposed compile check silently passed a broken unimported
   file (tsconfig `files:[main.ts]` checks the import graph, not the
   tree; fixed on the editor config). GATE_PROOF_STALE caught the
   premature attest the moment the gate changed — the standing
   check's first field catch, on its own author — but a false attest
   on a never-changing gate would stand. Candidate skill text:
   "attest only after watching every breakage fail, never in the
   same command chain as the proof".
7. **Measure aggregate readability**: the 806 duplicates are
   dominated by per-app scaffold configs (tsconfig ×12, test-setup
   ×12); the capped classified list hides that story — candidate:
   group-by-basename note in the measure.
8. **Held under load** [bench]: envelope discipline everywhere, 74ms
   inspect on 3k files, plan priced the real corpus (3.6KB render,
   11% of budget; costs tiers populated), triage buckets fit an
   already-curated 34-line corpus (drop 1 / demote 1 / keep / flag
   root notes), and the breakage proof caught a real gate flaw
   before it shipped.

From the real-repo pass on the second target, additional [real]:

9. **Leading global flags break dispatch.** `rein --dir X inspect` →
   `UNKNOWN_COMMAND: no such command: --dir`, and the fix ("run
   `rein` with no arguments for usage") never names the actual
   mistake. Agents write flag-first constantly. Either main learns to
   skip leading flags, or the fix says "flags come after the
   command".
10. **Confirms #1 at scale, worse**: `tests.candidates: null` on a
    127-project nx workspace where `nx.json` EXISTS and vitest is
    the actual runner (root jest.config.ts is stale bait) — and
    `monorepo: false` despite nx.json, sharpening #2: the canonical
    marker is missing from the table.
11. **The adoption move is undocumented.** Plan held the pre-existing
    19KB instruction file as unmanaged (guarantee held, confirms #5),
    but nothing — engine fix or skill step 6 — says HOW to adopt
    (remove the old file after triage; git is the recovery). Done
    here from product knowledge the skill should carry.
12. **Cousin of #6, injection side**: the first breakage proof landed
    in a file outside the gate's compile scope and the gate stayed
    green on a "broken" tree. Step 7 should say: a green gate on your
    breakage means the injection missed, not that the gate works —
    confirm the broken file is in scope before concluding anything.
13. **The verify skeleton needs an environment section.** Before any
    check could run, the gate had to absorb node>=22 activation (the
    shell carried v20; pnpm crashes on it — confirms the bench
    DEBT row) and an 8G heap (the workspace typecheck OOMs node's
    default at 51s, green at 90s with it). The skeleton starts at
    checks; real gates start at environment, with teaching errors.
14. **A triage rule died on evidence, correctly**: the corpus rule
    "no features/ dirs inside feature libs" is violated by 9+ libs
    AND contradicted by the corpus's own reference example — so the
    Demote candidate was downgraded to Flag with file:line proof
    instead of becoming a gate check red on the untouched tree. The
    buckets held under a 374-line corpus (→ 90-line runbook + a
    foreign tool's auto-managed block preserved verbatim as its own
    override fragment, pending the ownership ruling).
15. **Skills-dir cohabitation is real**: rein's skills now share
    `.claude/skills/` with another tool's (including `~origin_*`
    duplicate dirs from some earlier merge) — a shared write surface
    inspect's config_surfaces does not list.
16. **Held on the real repo** [real]: journal/attest/doctor end to
    end foreign-side (GATE_UNPROVEN → proof → attest → 0 findings);
    six evidence-backed debt rows seeded; plan priced the triaged
    render at 7,987 of 32,768 bytes; `rein note` carried the full
    ruling record; and the journal's first practical win — see the
    coordination flag above.

## Noted, not yet actioned

- A pre-rewrite backup mirror sits at
  ~/Work/.freerein-backup-prescrub.git (local-only by design; it
  contains the scrubbed names). Owner deletes it once confident in
  the rewritten history.

Found in the 2026-08-31 audit, recorded here rather than bundled into
an unrelated change:

- **Coverage cold spots: mostly paid down (2026-08-31 second pass).**
  Emit/emitHuman now tested behind an `EmitTo(io.Writer)` seam (exit
  codes, always-present branch fields, fix omission on the wire);
  Dump and Adapters tested; the flags-after-positionals loop
  extracted as `splitArgs` and tested (`engine/cmd/rein/main_test.go`).
  Remaining, deliberately: `detectHighTouch` (needs a fixture repo
  with real git history — build it when that code is next touched)
  and the two `main` wiring functions themselves.

## Assumptions / open questions

- The 80% NEAR_CONTEXT_BUDGET threshold is a chosen default, not a
  measured one; revisit when real repos hit it.
- `addresses` vocabulary is free-form; if overlap warnings prove noisy
  or vacuous, a registered vocabulary (like probes) is the likely fix.
- inspect's detection tables (manifests, test/lint/CI configs) are a
  curated list, not exhaustive; grow them from real setup runs, not
  speculation.
- The official registry is live but empty; what to publish first is a
  deliberate product decision (a spec-flow procedure is the obvious
  candidate), not something to bundle into infrastructure work. Until
  a first component is published, the archive-fetch path over live
  HTTPS rests on the httptest coverage, not a production run.
- Whether `rein doctor` clean belongs in the gate is unsettled. This
  repo is self-hosted, so a hand-edited CLAUDE.md is a real failure
  mode that only doctor catches — but `scripts/verify` does not run
  it, and CI deliberately runs nothing the gate does not. Putting
  doctor in CI alone would make CI stricter than the gate: verify
  green locally, red on push. If it should block, it belongs in
  `.rein/overrides/scripts/verify`, and that is a deliberate change to
  the definition of done, not a drive-by.

## Blocked / needs a human


- **Codex skill frontmatter dialect.** Still a declared degradation
  in `content/adapters/codex.yaml` (`user_invoked_flag: ""`,
  "invocation-governance flags unverified for this host"). Needs
  verified facts about the host before implementing — guessing here
  would ship a silent lie in a contract that exists to prevent them.
