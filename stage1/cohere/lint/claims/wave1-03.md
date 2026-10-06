# Lint wave 1 slot 03

Branch: codex/lint-wave1-03. Main base: d090af5.

Assigned positions 7, 8 and 9 in the helper-ready ordering published in
stage1/cohere/lint/HELPERS.md on origin/codex/lint-helpers (5d13f5b).
The helper REPORT.md links to this list rather than containing it itself.

1. @typescript-eslint/no-unused-expressions
2. @typescript-eslint/unified-signatures
3. class-methods-use-this

No implementations of these rules were found in stage1/cohere/lint on main or
any fetched origin branch. Occurrences in historical frequency logs are not ports.
No rule is skipped as already ported. This claim precedes all implementation.

## Foundation blocker

Registration (48ecd93) merges into current main. Helpers (5d13f5b) then conflicts
in README.md, lint.ts, lint_test.go, main.ts, settings.ts and testdata/oracle.go
under stage1/cohere/lint. Its history also introduces the earlier monolithic
rule engine and shared tests. The attempted helper merge was aborted to keep
registration intact. Neither shared engine implementation was silently discarded.

The requested docs/parallel-work.md is absent from main and both foundation
branches. The available registration contract is docs/lint-registration.md.

These rules are reserved by this claim but not implemented or certified.
The foundation integration must be resolved before this unit can honestly
provide the requested registered ports and byte-for-byte evidence.

A second independently reproduced blocker: registry.Discover opens rule.ts,
registry.Render emits imports to rule.ts, and mutant validation accepts only
.ts modules. A temporary rule.a-only copy of the existing no-debugger rule
fails generation with exit 1 and a missing rule.ts error. The copied-port
harness also discovers only .ts files. Supporting .a rule modules requires
shared registration and harness changes outside these three rule directories.
See wave1-03-evidence/extension.log and merge.log for observed failures.
