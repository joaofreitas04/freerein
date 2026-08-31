# FreeRein official registry

This directory is the published registry: `index.json` plus
sha256-pinned archives, served verbatim at
<https://joaofreitas04.github.io/freerein/> by the `registry` Pages
workflow.

It is machine-managed — written by `rein-publish` (which reads the
existing index, refuses republishing changed content without a
version bump, and packs deterministic archives), reviewed in the
diff like everything else, and never hand-edited. To publish:

    go run ./engine/cmd/rein-publish --out registry ./path/to/component

then commit the result. The index launched empty on purpose: core
content ships inside the `rein` binary, and the registry is the
extension/preset channel — what to publish first is a product
decision, not an infrastructure one.
