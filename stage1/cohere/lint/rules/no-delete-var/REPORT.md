Built: no-delete-var in .a with directory registration and exact Go findings; no repairs.
Commits: claim ab49eb4d was pushed before code; implementation 0b6011d5; evidence commit in git history.
Commands and outputs: Go differential corpus on Node, emitted JavaScript and sanitized native; see the unit evidence.
Mutant: mutant.json changes the rule judgment while preserving compilation and clean execution; all three comparisons must reject it.
Not covered: unmodified shared harness, pinned CLI self-lint of .a, full repository gate or arbitrary malformed inputs.

This directory owns the rule, message, descriptor, independent Go adapter, raw witness and semantic mutant. None of these three rules offers a fix or suggestion. No shared source was edited.

See [the unit report](../no-constructor-return/REPORT.md) for complete coverage, commands, throughput and integration limits.
