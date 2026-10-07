Built: no-octal candidate in .a with exact Go rule-phase records; decimal suggestions are complete.
Commits: claim 6f28de0b was pushed before code; implementation 1dab0e83; evidence commit in git history.
Commands and outputs: bounded Go parity on Node, emitted JavaScript and sanitized native; raw logs and replay saved.
Mutant: mutant.json must compile and exit cleanly before all three comparisons reject its changed behavior.
Not covered: JSX/recovery parser gaps, registered decimal suggestion rendering, ordinary harness, pinned .a self-lint or full gate.

This directory owns its implementation, message, registration, unmodified Go adapter, witness and semantic mutant. No shared parser, compiler, registration or harness source was edited. Complete record types come from the previously owned typescript-prefer-as-const/records.a on this branch; these candidates require that module during integration.

See [the unit report](../no-nonoctal-decimal-escape/REPORT.md) for exact coverage, blockers, commands and throughput.
