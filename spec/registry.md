# Registry — contract v0.3

A registry is a **static JSON index** over HTTPS (or a filesystem
path — used for local/private registries and tests). No server-side
logic: the index maps `name` → `version` → a sha256-pinned archive.

```json
{
  "components": {
    "spec-flow": {
      "0.3.0": {
        "url": "spec-flow-0.3.0.tar.gz",
        "sha256": "…",
        "kind": "extension",
        "description": "…",
        "addresses": ["…"]
      }
    }
  }
}
```

Relative `url`s resolve against the index location. Archives are
tar.gz of a component directory (`component.yaml` + payload); entries
that escape the root or are not regular files/dirs (symlinks
included) are refused at unpack.

## The trust model

1. **Vendoring, not caching.** `rein add name@version` fetches,
   verifies the archive sha256 against the index pin, and unpacks
   into `.rein/vendor/name@version/` — **committed to the repo**.
   Consequences: consume-side commands stay offline (guarantee 4),
   and what was installed is reviewable in the diff, which is the
   only review that actually happens.
1b. **The index carries each release's `addresses`** (v0.3, copied
   verbatim from the manifest at publish time): a consumer joining
   detected gaps to available extensions reads coverage from the
   index alone, without fetching an archive. Addresses are coverage
   claims, never lift claims — "X addresses this gap" is checkable;
   "X will help here" is not.
2. **Pin at two levels.** The index pins the archive (sha256); the
   lockfile pins the unpacked tree (deterministic tree hash on the
   layer ref). `doctor` recomputes the tree hash: a mismatch is
   `VENDOR_TAMPERED` — content changed after install, whether by an
   attacker, a bad merge, or a hand edit that belongs in a
   local-path source instead.
3. **A sha mismatch at fetch refuses and cleans up.** The registry
   and the archive disagreeing is never resolved in the archive's
   favor.
4. **Version pinning is not review** and is never presented as such.
   `info` exists so a human can inspect a component's manifest —
   kind, subsystem, rung, rent, provides, requires — before anything
   is written; registry fetches for `info` go to a temp dir and are
   discarded.

## Command surface

| Command | Network | Writes |
|---|---|---|
| `info <name[@ver] \| ./path>` | index (+ archive to temp) | nothing |
| `add <name[@ver]>` | index + archive | `.rein/vendor/`, `harness.yaml` |
| `remove <name>` | none | `harness.yaml`, drops unreferenced vendor dirs |
| `upgrade [--yes]` | index (+ archives with `--yes`) | two-phase: reports, then re-vendors + updates declarations |
| plan/apply/dump/doctor | **none** | (as specified elsewhere) |

`add`/`remove`/`upgrade` never touch managed files: installation and
removal always go through `plan`/`apply`, where the two-phase confirm
and three-way merge live. An upgrade over a locally edited file is
therefore an ordinary merge, not a special case.

Registry selection: `--registry` flag > `registry:` in
`harness.yaml` > the built-in default,
`https://joaofreitas04.github.io/freerein/index.json`. Selection can
therefore never come up empty, and v0.1's `NO_REGISTRY` diagnostic is
retired with this version — an unknown component against the default
registry fails as what it is (`NOT_IN_REGISTRY`), not as a missing
setting. Only the commands in the table above ever touch the network;
a default URL changes nothing about plan/apply/doctor staying
offline.

## The official registry

The default URL serves the `registry/` directory committed to the
FreeRein repo, verbatim, via GitHub Pages: `index.json` plus the
sha-pinned archives beside it. It is written only by `rein-publish`
(deterministic archives; republishing changed content without a
version bump is refused), so publishing is a reviewed commit like any
other change — there is deliberately no generation step at deploy
time. The index launched empty: core content ships inside the engine
binary, and the registry is the extension/preset channel.
