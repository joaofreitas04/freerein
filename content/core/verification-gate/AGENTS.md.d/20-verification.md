## Verification

`bash scripts/verify` green is the definition of done for every
change — run it before reporting completion and paste its outcome.
If it fails, the failure is the task. Never edit the gate to pass it;
project checks are added via `.rein/overrides/scripts/verify`.
