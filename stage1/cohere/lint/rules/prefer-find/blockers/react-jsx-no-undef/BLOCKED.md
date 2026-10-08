# react/jsx-no-undef

Status: blocked on the shared area API/helper, not an implemented descriptor.

Exact upstream call: `ctx.TypeChecker.GetSymbolAtLocation(reference)` at `cohere/internal/lint/rules/react/jsx_no_undef.go:101`.

Missing: complete identifier declarations and same-source identity for jsxNoUndefDeclaredInFile; symbol-origin does not provide that set.

Reproducer: [reproducer.txt](reproducer.txt). The source requires the same ordinary dependencies/import files as the corresponding upstream case.

The shared checker model, CHECKER_REPORT.md, descriptor contract and current area Inspect dispatch were reviewed. Old private bridge questions are not an allowed substitute. No shared files were edited.

Upstream cases matched for this rule on this branch: 0. Mutant: not run because no rule is registered. Stop here until the shared question/helper lands.
