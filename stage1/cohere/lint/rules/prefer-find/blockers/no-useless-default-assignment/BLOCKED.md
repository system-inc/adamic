# @typescript-eslint/no-useless-default-assignment

Status: blocked on the shared area API/helper, not an implemented descriptor.

Exact upstream call: `checker.Checker_getTypeOfSymbol(ctx.TypeChecker, parameterSymbol)` at `cohere/internal/lint/rules/typescript/no_useless_default_assignment.go:396`.

Missing: contextual signature parameter identities/types beyond the first parameter; call-parameters only exports the first parameter of each signature.

Reproducer: [reproducer.txt](reproducer.txt). The source requires the same ordinary dependencies/import files as the corresponding upstream case.

The shared checker model, CHECKER_REPORT.md, descriptor contract and current area Inspect dispatch were reviewed. Old private bridge questions are not an allowed substitute. No shared files were edited.

Upstream cases matched for this rule on this branch: 0. Mutant: not run because no rule is registered. Stop here until the shared question/helper lands.
