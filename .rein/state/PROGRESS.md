# Progress

Managed by the working agent, structured for the next session's cold
start. Update before ending every session.

## Current state

2026-08-28: engine at 0.1.0-dev, self-hosted — this repo's harness is
rein-managed as of this commit (CLAUDE.md rendered, verify adopted via
.rein/overrides/, skills installed). Done and tested: walking skeleton,
three-way merge over .rein/base/, codex adapter, local + registry
sources (sha-pinned vendor, tamper detection), rein-publish generator,
unmanaged-clobber guard, setup + diagnose procedures.

## In flight

Nothing.

## Blocked / needs a human

- Registry hosting: publish the official index somewhere real
  (GitHub Pages / releases) and set a default registry URL.
- Codex skill frontmatter dialect: still a declared degradation —
  needs verified facts about the host before implementing.
