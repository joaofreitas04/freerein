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

**Next, in `docs/lifecycle.md` §4 order** — observe ships before
evolve, and evolve's mechanics do not land until its substrates
exist:

1. **Use the loop on real work.** The full lifecycle now exists —
   init → operate → observe (journal, gate proof, citations, costs)
   → diagnose → evolve (propose/verdict) → upgrade — and every
   further mechanism should be demanded by a real run, not
   speculated. Field test 2 on a brownfield repo is the natural next
   move: it exercises rein-setup 0.7 end to end and starts real
   citation data accruing, which the first real evolve proposal
   needs. The first proposal on THIS repo waits for that data
   (day-one zeros are the misreading, not evidence).

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

Also unstarted and named in `AGENTS.md`: the minimal profile as a
shipped control mode; release automation.

## In flight

Field test 2: with the owner (2026-08-31). The target is a private
multi-app Angular/Nx workspace of the owner's — its name and paths
stay out of this public repo (the owner knows which; the bench path
below locates it locally). The owner runs the real install and
interview
themselves — their explicit ruling, and the setup interview is
designed around rulings only they can make. rein 0.2.0 (released
binary, checksum-verified) is on their PATH (~/.local/bin/rein).

A parallel session, unaware of that ruling, completed a full
autonomous pass on a scratch clone (bench at ~/.cache/ft2/<target>,
1.6GB, kept as evidence; the real repo verified untouched). Its
interview rulings are VOID — defaults for demonstration, decided by
nobody — but its engine-side results are objective and are merged
into the findings below, marked [bench]. The bench's DEBT.md holds
target-side facts for the owner's real install (format drift: 325
files fail nx format:check; system pnpm crashes under volta node 20;
fresh clones need the codegen script before anything compiles).

UPDATE, same day: the owner then explicitly directed this session to
run the full setup autonomously on a SECOND target — a larger private
workspace of theirs (real repo, in place; names stay out of this
public repo). That live instruction supersedes the earlier
interview-is-theirs ruling for that target only; every triage and
ownership ruling made there is still a DEFAULT, marked pending the
owner in that repo's PROGRESS, journal note, and local branch
`field-test-2/rein-setup` (committed locally, NOT pushed, per the
owner's explicit no-push rule; their original branch restored).
Findings from that pass are merged below, marked [real]. One
coordination flag FOR THE OWNER: that repo's journal shows an apply
at 14:04:12Z this session did not issue — consistent with the
owner's own install (rein on their PATH) running concurrently;
if so, this session's 14:13 apply landed on top of it. The journal
made the collision legible; the owner should reconcile.

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

- **GitHub support purge (privacy, 2026-08-31).** The full history
  rewrite removed a private repo name from 11 pre-tag commits and
  today's leak from the tip; force-pushed, tags re-pointed, releases
  intact. But force-pushed-away objects stay fetchable by raw SHA on
  GitHub until garbage collection — completing the removal needs a
  support request (https://support.github.com/contact, "remove
  cached/unreachable objects after a sensitive-data history
  rewrite"), which only the account owner can file. A local backup
  mirror with the pre-rewrite history sits at
  ~/Work/.freerein-backup-prescrub.git (deliberately local-only;
  delete once confident).

- **Codex skill frontmatter dialect.** Still a declared degradation
  in `content/adapters/codex.yaml` (`user_invoked_flag: ""`,
  "invocation-governance flags unverified for this host"). Needs
  verified facts about the host before implementing — guessing here
  would ship a silent lie in a contract that exists to prevent them.
