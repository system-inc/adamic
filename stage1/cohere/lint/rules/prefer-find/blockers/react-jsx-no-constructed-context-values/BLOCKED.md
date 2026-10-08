# react/jsx-no-constructed-context-values

Status: blocked on the shared area API/helper, not an implemented descriptor.

Exact upstream call: `ctx.TypeChecker.GetSymbolAtLocation(identifier)` at `cohere/internal/lint/rules/react/jsx_no_constructed_context_values.go:320`.

Missing: identifier/local symbol declaration and initializer details plus shared memo/reference/capture analysis; no private helper copy is introduced.

Reproducer: [reproducer.txt](reproducer.txt). The source requires the same ordinary dependencies/import files as the corresponding upstream case.

The shared checker model, CHECKER_REPORT.md, descriptor contract and current area Inspect dispatch were reviewed. Old private bridge questions are not an allowed substitute. No shared files were edited.

Upstream cases matched for this rule on this branch: 0. Mutant: not run because no rule is registered. Stop here until the shared question/helper lands.
