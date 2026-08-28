# harness.lock — contract v0.1 (draft)

The resolved truth of the installed harness. Machine-written only;
JSON; committed to the target repo. Every later `rein` run diffs
against it — it is what makes the four guarantees mechanical.

```json
{
  "version": 1,
  "resolved_at": "2026-08-28T12:00:00Z",
  "engine": "rein 0.1.0",
  "adapter": { "name": "claude-code", "version": "0.1.0" },
  "layers": [
    { "id": "overrides", "source": "local" },
    { "id": "preset:strict-review", "source": "github.com/x/y@1.2.0", "sha": "…" },
    { "id": "core", "source": "embedded@0.1.0" }
  ],
  "files": {
    "AGENTS.md": {
      "layer": "core",
      "component": "instructions-base@0.1.0",
      "hash": "sha256:…",
      "shadowed": [],
      "refs": ["core"]
    }
  }
}
```

## Semantics

- `hash` is the rendered content at install. `doctor` compares it to
  the working tree: a mismatch is **local drift — detected and
  respected**, never silently overwritten. Upgrades on drifted files
  go through a three-way merge (base = lockfile's recorded content
  source, ours = working tree, theirs = new upstream).
- `shadowed` records the next-highest-priority provider of the same
  path, so `remove` restores it — guarantee 3.
- `refs` is the refcount: which components/bundles need this file.
  `remove` deletes a file only when `refs` drains to empty.
- The lockfile plus `rein dump` together are the enumeration of
  everything FreeRein put in the repo — the trust boundary as a list
  the machine maintains.
