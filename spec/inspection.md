# Inspection — contract v0.2 (draft)

`rein inspect` is the discovery half of setup made mechanical: a
read-only sweep of the target repo that reports what is *there*, so
the judgment procedure spends its budget on what the facts *mean*.
Full report to `.rein/out/inspect.json`; the envelope carries a
summary and the path.

## The one hard rule

**Inspection never executes project code.** It reads files, and it may
read local git history; it runs no build, no test, no script. A test
*candidate* in the report is a command inspection derived from a
manifest — running it, measuring it, and judging it belongs to the
setup procedure, which does so deliberately and reports the outcome
to the human. Detection is computation; execution is judgment.

## Families

| Family | Reports |
|---|---|
| `toolchain` | languages, manifests, package managers, monorepo markers |
| `measure` | native size-and-composition measurement: per-language file/line counts by a method the report itself states, and a per-file classification so nothing is silently skipped — every file lands in exactly one state (`analyzed`, `empty`, `binary`, `oversize`, `generated`, `duplicate`, `unknown`, `error`) — an unrecognized file and an unreadable file are different facts, never conflated |
| `tests` | test configurations found, and manifest-derived run candidates (command + the file it came from) |
| `ci` | CI configuration files — checks some system already enforces are gate candidates with provenance |
| `lint_format` | linter and formatter configurations |
| `instruction_corpus` | every agent instruction file (root and nested, bounded walk), with byte sizes — the input to setup's triage |
| `config_surfaces` | agent-config files more than one tool commonly writes — the input to the ownership interview |
| `high_touch` | the most-changed paths from local git history — an enumerated map of where erosion happens (omitted, with a note, when history is unavailable) |
| `docs_tree` | whether a docs territory exists for the instruction file to be a map of |
| `affordances` | the boolean projection of all of the above, aligned with the probe vocabulary |

## Rules

1. **Offline, read-only, deterministic** for a given tree + history.
   Walks are pruned at dot-directories and dependency/build
   directories; the corpus walk is additionally depth-bounded.
   Reading a file is not executing it — the one hard rule holds
   through every family.
2. **The probe vocabulary is inspection's boolean shadow.** Every
   `rein probes` entry is answerable from this report; `requires:`
   gating and `inspect` cannot disagree because they share detection.
3. **Absence is a finding.** No test configuration, no CI, no docs
   tree — each is reported plainly; the affordance map is what makes
   the brownfield conversation honest before anything is installed.
4. `inspect` works before `rein init` — discovery precedes
   declaration.
5. **Measurement discipline.** The report carries a semver'd
   `formatVersion` and the engine that wrote it. Counting is opinion,
   so the method is stated *inside* the report next to the numbers
   (physical lines as newline count, non-blank lines, no
   code/comment split — that classification is a lexer's job this
   surface deliberately does not take on). Generated files are
   recognized two ways — by name pattern and by a marker in the
   file's leading bytes — and duplicates by equal size then equal
   content hash, later path marked, first kept. Any capped list in
   the report says so in `notes`: a cut announces itself.
