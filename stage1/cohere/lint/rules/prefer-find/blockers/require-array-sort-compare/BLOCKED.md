# @typescript-eslint/require-array-sort-compare

Status: blocked on the shared area API/helper, not an implemented descriptor.

Exact upstream call: `type_checking.GetTypeName(ctx.TypeChecker, argument)` at `cohere/internal/lint/rules/typescript/require_array_sort_compare.go:169`.

Missing: shared GetTypeName normalization, including annotation/declaration-based type-parameter string constraints; checker name is TypeToString, not this helper.

Reproducer: [reproducer.txt](reproducer.txt). The source requires the same ordinary dependencies/import files as the corresponding upstream case.

The shared checker model, CHECKER_REPORT.md, descriptor contract and current area Inspect dispatch were reviewed. Old private bridge questions are not an allowed substitute. No shared files were edited.

Upstream cases matched for this rule on this branch: 0. Mutant: not run because no rule is registered. Stop here until the shared question/helper lands.
