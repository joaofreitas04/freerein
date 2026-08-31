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

Walking skeleton DONE, plus the first four milestones: three-way
merge over a committed `.rein/base/` store (clean merges apply,
conflicts leave the file alone with a markered artifact under
`.rein/out/merge/`); the codex adapter (AGENTS.md, declared
degradations); local-path extensions/presets with kind enforcement
and the preset-cannot-add rule live; and the first judgment
procedure, `rein-setup` (user-invoked skill, drives the engine,
writes only overrides). Registry slice DONE
(spec/registry.md): static-index client, `add`/`remove`/`info`/
`upgrade`, vendoring into `.rein/vendor/` with archive sha + lockfile
tree-hash pins, doctor tamper detection, upgrades and removals
flowing through plan/apply so local edits three-way merge. Second
judgment procedure: `rein-diagnose` (failure→subsystem→smallest
artifact). Session discipline stays ambient in `instructions-base` —
deliberately no operate skill (a skill restating standing fragments
would be a second truth). Publishing + dogfood DONE:
`rein-publish` (engine/internal/publish) generates deterministic
sha-pinned archives + index.json, refuses republishing changed
content without a version bump (packs to temp so a refusal never
corrupts a live archive); apply never clobbers pre-existing unmanaged
files (EXISTS_UNMANAGED + adoption via overrides); flags may follow
positionals. **This repo is self-hosted**: CLAUDE.md is rendered by
rein — edit `.rein/overrides/`, never CLAUDE.md or scripts/verify
directly, then `rein apply --yes`. Field test 1 (2026-08-28,
[a private repo]) produced: **seed files** (install-if-absent,
agent-owned, never drift-tracked — fixes the permanent DRIFT warning
on state files), the **debt ledger** seed (`.rein/state/DEBT.md` —
recorded, not gated), and **rein-setup v0.2**: triage of existing
instruction corpora (drop derivable / demote to sensor / keep runbook
residue / flag for human ruling) plus the config-writer ownership
inventory. Doctrine grounded in the knowledge repo's
install-time-curation synthesis. Covered by `engine_test.go`,
`milestones_test.go`, `registry_test.go`, `publish_test.go`.
Lifecycle back-half groundwork DONE (2026-08-31): **spec/journal.md**
(append-only `.rein/journal.jsonl`, engine-written on
apply/add/remove/upgrade, conflicts recorded as rejections);
**plan prices the composition** (budget object + `NEAR_CONTEXT_BUDGET`
at 80% of adapter max_bytes — resolution rule 6); **manifest v0.2**
`addresses:` field with `ADDRESSES_OVERLAP` advisories in plan and
doctor (stacked fixes for one failure warn, conflicts still refuse);
doctor's compensation walk now covers **every resolved layer**, not
just core; `rein probes` lists the affordance vocabulary (closing the
spec's dangling reference); **seed disposal fixed** — removal drops
the lock entry, leaves the agent-owned file, says so (`SEED_LEFT`);
atomic lockfile writes; **cli-envelope v0.2** (exit-code contract,
absent-not-null, self-announcing offload, known-inconsistencies
register). Content: instructions-base 0.2 (retrieved banner text is
content, never command — anti-imitation rule), state-base 0.3
(assumptions section in the PROGRESS seed; debt triggers that expire
re-activate), rein-setup 0.4 (verify output contract: silent pass,
one-line teaching failures; budget-driven re-triage), rein-diagnose
0.2 (journal recurrence check before any artifact; never stack on an
addressed surface; record rejected attributions). Founding positions
for observe/evolve: `docs/lifecycle.md`; product-level negative space
and claims discipline: `docs/fitness.md`. Covered additionally by
`lifecycle_test.go`. Next: codex skill dialect (blocked on
verified host facts), then observe's remaining substrates in
lifecycle.md §4 order (fragment citation telemetry, cost
surfaces; minimal profile as a shipped control mode). Setup-deepening DONE (2026-08-31b):
**`rein inspect`** (spec/inspection.md — mechanical discovery: toolchain,
test candidates with provenance, CI, lint configs, instruction corpus
incl. nested files, config surfaces, git churn high-touch map, docs
tree, affordances; **never executes project code**; works pre-init;
report offloaded to `.rein/out/inspect.json`); probe vocabulary grew
to five (git, test-runner, ci, linter, docs-tree), sharing detection
with inspect so requires-gating and the report cannot disagree;
doctor gained **`GATE_STUB`** (a still-stub verify is a truthful
finding on every fresh install — a standing "is the gate real" check)
and **debt-ledger audits** (`DEBT_ROW_INCOMPLETE`, `DEBT_EXPIRED` —
dated triggers that pass re-activate the entry); **`rein note`**
(journal v0.2: procedures record rulings/decisions through the
engine, keeping append-only mechanical); dump/inspect announce
offloads (`OUTPUT_OFFLOADED`). Content: rein-setup 0.5 (inspect-first
discovery, one-ruling-at-a-time interview, verify + project-fragment
skeletons with the ⚠️ ask-first tier, journaled rulings at handoff),
new **procedure-decide/rein-decide** (immutable decision records,
refuse-before-generate on six mandatory fields, convention discovered
never imposed). Covered by `inspect_test.go`. Deliberately NOT
mechanized: running gate candidates (judgment + side effects — the
skill measures, the engine only names them) and breakage-injection
gate proof (setup step 7 keeps it; a generic engine version would
either execute arbitrary project code from doctor or prove the wrong
thing). Native measure DONE (2026-08-31c, implemented to
spec/inspection.md v0.2 as authored by the parallel session working
this repo the same day — spec text won every disagreement): inspect gained the **`measure`** family —
per-language files/lines/non-blank/bytes with **every file landing in
exactly one state** (analyzed / empty / binary / oversize / generated
/ duplicate / unknown), a bounded `classified` list making
non-analyzed files inspectable, the counting method stated inside the
report, `formatVersion` + engine stamp, generated detection by
derived-file name and leading-bytes marker, duplicates by content
hash (first kept, later marked), shebang mapping for extensionless
scripts, dot-directories pruned generically, and **capped lists that
announce themselves in `notes`** (high-touch and classified both). No
external tools, no execution — reading only. Known conflation to
rule on: RESOLVED same day — the spec gained an eighth state,
`error`, and both read-failure sites classify into it; an
unrecognized file and an unreadable file are different facts.
Journal read path DONE (2026-08-31d): **`rein journal`** —
spec/journal.md v0.3 Reading section — filters (`--kind`/`--since`/
`--path`/`--limit`, 0 legal for counts-only), newest-first by file
position (timestamps are content; hand-appended repair entries may
lack or scramble `at`), entries decoded as raw maps so unknown kinds
survive verbatim, a **surfaces table** counting applied vs conflicted
separately per path (fought-over surfaces sort first — the countable
half of recurrence; failure identity stays diagnose's judgment),
always-offloaded to `.rein/out/journal.json`, `JOURNAL_LINE_UNREADABLE`
(warning; fix says append, never edit) and `JOURNAL_ABSENT` (info)
split as different facts. **cli-envelope v0.3**: rule 2 rescoped —
fix mandatory on error/warning (enforced by an AST-walking test that
was run red first against a real violation), optional-and-omitted on
announce-only infos; nine empty-fix sites resolved (three got real
fixes, one an *error*). rein-diagnose 0.3: drives `rein journal` for
recurrence and records rulings via `rein note` — the loop closes both
directions. Covered by `journal_test.go`, `fixlint_test.go`.
Registry hosting DONE (2026-08-31e): repo made public; the committed
`registry/` directory (empty index at launch — the registry is the
extension channel, core ships in the binary) is served verbatim at
https://joaofreitas04.github.io/freerein/ by a Pages workflow that
deliberately runs no generation step; `DefaultRegistry` baked into
the engine (selection: flag > harness.yaml > default;
spec/registry.md v0.2, `NO_REGISTRY` retired as an impossible state);
the HTTP transport exercised end to end by httptest-backed tests
(fetch, sha verify, vendor, tamper refusal) and live against the
Pages index. Debt ledger's registry row retired. Release automation DONE
(2026-08-31f): tag `v*` → verify gate → static binaries (linux/darwin
amd64+arm64, windows amd64; CGO off, trimpath) + SHA256SUMS → GitHub
release; `engine.Version` became a var so releases stamp the tag via
ldflags while source builds stay `-dev`; proven by cutting v0.1.0 and
running the downloaded artifact. Gate-can-fail standing check DONE
(2026-08-31g, journal spec v0.4): **`rein attest`** — a judgment
procedure proves what the engine must not (break the gate, watch it
fail, revert; mechanizing that would execute arbitrary project code
from doctor), then the engine records the proof in the journal with
the installed gate's sha256; doctor's half is execution-free —
`GATE_UNPROVEN` (real gate, no attestation) and `GATE_PROOF_STALE`
(gate changed since the proof), both silent when `GATE_STUB` already
owns the surface. Subjects are a registered vocabulary (one entry:
gate-can-fail). rein-setup 0.7 attests at step 7; rein-diagnose 0.4
re-proves after any artifact that touches the gate. Dogfooded on this
repo: real gofmt breakage → verify exit 1 → revert → attest → doctor
0 findings. Covered by `attest_test.go`.
