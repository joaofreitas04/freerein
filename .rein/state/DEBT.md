# Debt ledger

Known-broken or never-run things, recorded instead of gated — a gate
that is red from birth trains agents to ignore it. Each entry names
the debt, the evidence, and the paydown trigger. Retire entries by
fixing them, never by deleting the line without the fix.

| Debt | Evidence | Paydown trigger |
|---|---|---|
| Codex adapter never validated against a real codex host | `content/adapters/codex.yaml` ships `user_invoked_flag: ""` plus the degradation "skills: invocation-governance flags unverified for this host"; no codex install has ever applied it | First access to a codex host: verify the skills directory, the frontmatter dialect and whether a user-invoke flag exists, then either implement the dialect or restate the degradation as a measured fact rather than an assumed one. |
