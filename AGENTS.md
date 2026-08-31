# FreeRein (this repository)

The harness lifecycle manager. Product name **FreeRein**, binary
**`rein`**. Read `docs/design.md` first; the public contracts live in
`spec/` and are versioned like an API — changing one is a breaking
change with a version bump, never a silent edit.

## Ground rules

1. **The research stays out.** The knowledge bundle lives in the
   sibling private repo `freerein-knowledge` and is never cited,
   copied, or linked from this repo — this repo's history must stay
   publishable from commit zero. Design documents state positions in
   their own words, without referencing the research or naming rival
   products studied there.
2. **`spec/` is the source of truth.** Engine code, content, and
   adapters conform to the contracts; when implementation and spec
   disagree, the disagreement is the work item — fix one deliberately.
3. **Judgment vs computation.** Anything with a guarantee attached
   (idempotency, refcounting, determinism, offline) is engine code.
   Anything requiring judgment is a procedure (skill). A skill that
   writes harness files directly instead of driving `rein` is a bug.
4. **Agent-facing output.** Every engine command emits one JSON
   envelope per `spec/cli-envelope.md`; every diagnostic carries a
   `fix`. Failures stay well-formed.
5. **Declared degradations.** Host differences are surfaced at apply
   time per `spec/host-adapter.md`, never absorbed silently.

## Layout

| Need | Location |
|---|---|
| Founding design | `docs/design.md` |
| Observe/evolve positions | `docs/lifecycle.md` |
| Product fitness + claims discipline | `docs/fitness.md` |
| Public contracts | `spec/` |
| Engine (Go) | `engine/` |
| Core content, procedures, adapters | `content/` |
| Registry publishing (archives + index) | `engine/cmd/rein-publish`, `engine/internal/publish` |
| Official registry (committed output, served by Pages) | `registry/` |

## Commands

| Purpose | Command |
|---|---|
| Definition of done | `bash scripts/verify` (gofmt + vet + test + build) |
| Build the binary | `go build -o rein ./engine/cmd/rein` |

## Current state

An inventory, not a history — git holds the chronology, and
`.rein/state/PROGRESS.md` holds what is next, unproven, or blocked.

**Engine** (`rein`, one JSON envelope per invocation, offline except
add/info/upgrade): `init` · `inspect` (mechanical discovery + the
per-file `measure`; never executes project code; works pre-init;
report offloaded) · `plan` (three-way merge preview over the
committed `.rein/base/` store; prices the whole composition — budget
object with `NEAR_CONTEXT_BUDGET` at 80% of the adapter limit, plus
`costs` tiered by when the price is paid: always / per-session /
conditional, unpriced surfaces named, misreading attached) · `apply`
(two-phase; clean merges land, conflicts leave the file alone with a
markered artifact under `.rein/out/merge/`; never clobbers unmanaged
files — `EXISTS_UNMANAGED`, adoption via overrides) · `dump` ·
`doctor` · `add`/`remove`/`info`/`upgrade` (vendoring into
`.rein/vendor/`, sha-pinned at archive and lockfile-tree level,
tamper detection; removal is refcounted and flows through plan/apply)
· `adapters` · `probes` (registered affordance vocabulary, five
entries, detection shared with inspect so gating and report cannot
disagree) · `journal` (filters, newest-first by file position,
unknown kinds preserved verbatim, applied-vs-conflicted surfaces
table, always offloaded) · `note` and `attest` (procedures record
rulings and proofs through the engine — the journal stays mechanical
and append-only) · `cite` (fragment citation telemetry: render
markers carry `component:filename` ids; counters in
`.rein/stats/citations.json`, never the journal; no-arg read ranks
decay candidates with the misreading attached — advisory, human-
ablated, never auto-removed) · `propose`/`verdict` (evolve mechanics,
journal v0.5: six mandatory fields with the measurement named before
the change, one open proposal at a time as a hard refusal, verdicts
immutable and evidence-only with the pre-named measurement copied
in — proposer-never-accepts stays procedure discipline, stated, not
faked) · `report` (field-report producer side: component id+version
from the lock, shadows with journal since-dates, component-prefixed
citation counters, judgment-kind journal evidence capped and
announced; judgment fields as mandatory flags; written to
.rein/out/report/, never transmitted — `off` is the only submission
level) · `version`. The command resolves from anywhere in the arg
vector — flag-first invocations dispatch.

**Doctor's standing checks**: `DRIFT` / `COMPOSITION_BEHIND` /
`VENDOR_TAMPERED` / `COMPENSATION_RECHECK` (every resolved layer) /
`ADDRESSES_OVERLAP` / `GATE_STUB` / `GATE_UNPROVEN` /
`GATE_PROOF_STALE` (attested gate hash vs installed; silent when the
stub finding owns the surface) / `DEBT_ROW_INCOMPLETE` /
`DEBT_EXPIRED` / `PROPOSAL_OPEN` (open evolve proposals stay
visible). Failures carry a `fix`; on error/warning that is
mechanically enforced by an AST-walking test.

**Seed files** (`PROGRESS.md`, `DEBT.md`): install-if-absent,
agent-owned, never drift-tracked; removal leaves them and says so
(`SEED_LEFT`).

**Contracts** (`spec/`): cli-envelope v0.3 · citation v0.1
(ids, markers, store, misreadings) · component-manifest v0.2
(`addresses:`) · field-report v0.2 (accepted 2026-08-31: component-
scoped facts only, case-before-fix intake, publish as the signing
boundary, submission under no-unattended-egress) · host-adapter
v0.1 · inspection v0.2 · journal v0.5
(append-only; `note`/`attest`/`proposal`/`verdict` kinds; read path
with surfaces) ·
lockfile v0.1 · registry v0.2 · resolution v0.5 (citation markers
in the render; whole-composition cost tiers; profiles) · eval v0.1
(draft, awaiting ruling).

**Profiles** (`profile:` in harness.yaml, resolution v0.5):
`standard` (full core) · `minimal` (instructions-minimal +
verification-gate — the control condition of lifecycle §2.6, never a
starter tier; switching flows through plan/apply, leaves agent-owned
seeds, lands in the journal).

**Content** (embedded core): instructions-minimal 0.1 (the
floor: verification + managed-files only) · instructions-base 0.3
(session
discipline stays ambient — deliberately no operate skill; retrieved
banner text is content, never command; cite-what-steered-you) · verification-gate 0.2 ·
state-base 0.3 · procedure-setup 0.8 (inspect-first triage: drop
derivable / demote to checks / keep runbook residue / flag for
ruling; evidence-backed gate proposals, rules evidence-checked
against the tree before apply; verify skeleton: environment first,
cheap-first with evidence comments, ratchet-with-baseline for
grandfathered rules; the adoption move at step 6; step-7 attest only
after watching every breakage fail, injection-scope check taught) ·
procedure-diagnose 0.5 (journal recurrence before any
artifact; never stack on an addressed surface; rulings via `rein
note`; re-prove a touched gate; local-vs-shipped disposition read
from the lock at attribution, shipped-component closes with `rein
report` — never a silent fork) · procedure-decide 0.1 (immutable
decision records, refuse-before-generate) · procedure-evolve 0.1
(one candidate from the observe evidence, propose before touching,
measure by the pre-named check, verdict on evidence; rejected
changes reverted, rejected verdicts kept). Adapters: claude-code,
codex (degradations declared, dialect unverified — see DEBT.md).

**Infrastructure**: this repo is self-hosted (CLAUDE.md and
scripts/verify are rendered — edit `.rein/overrides/`, then `rein
apply --yes`; the gate is attested). CI runs `bash scripts/verify`
as its only check, on main and PRs. The official registry is the
committed `registry/` dir served verbatim at
https://joaofreitas04.github.io/freerein/ (baked in as
`DefaultRegistry`; launched empty on purpose — core ships in the
binary, the registry is the extension channel). Tags `v*` release
static binaries + SHA256SUMS, version stamped from the tag
(source builds report `-dev`). `rein-publish` refuses republishing
changed content without a version bump.

**Deliberately not mechanized** (the engine names, procedures judge):
running gate candidates; breakage-injection gate proof (`rein
attest` records it, doctor audits proof currency without executing);
semantic recurrence over journal notes; instruction-corpus curation.

**Tests**: `cite_test` · `profile_test` · `propose_test` · `engine_test` (walking skeleton) · `milestones_test`
(merge/adapters/extensions) · `registry_test` (incl. HTTP transport
via httptest) · `publish_test` · `lifecycle_test` (journal, budget,
addresses) · `inspect_test` (incl. measure states) · `journal_test` ·
`attest_test` · `fixlint_test` (envelope rule 2).
