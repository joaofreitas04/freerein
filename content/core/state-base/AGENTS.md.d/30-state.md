## Session state

Start every session by reading `.rein/state/PROGRESS.md`. End every
session by updating it: what changed, what is in flight, what is
blocked and why. Keep it current-state-only — history lives in git.

Known-broken things live in `.rein/state/DEBT.md` — when you hit one,
check the ledger before diagnosing from scratch; when you fix one,
retire its entry in the same change.
