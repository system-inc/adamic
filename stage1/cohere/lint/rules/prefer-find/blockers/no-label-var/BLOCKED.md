# no-label-var

Status: blocked on the shared area API/helper, not an implemented descriptor.

Exact upstream call: `ctx.TypeChecker.GetSymbolsInScope(node, ast.SymbolFlagsValue)` at `cohere/internal/lint/rules/core/no_label_var.go:77`.

Missing: GetSymbolsInScope values across the complete scope chain; scope-locals is a different question, and symbols-in-scope is absent from the area bridge.

Reproducer: [reproducer.txt](reproducer.txt). The source requires the same ordinary dependencies/import files as the corresponding upstream case.

The shared checker model, CHECKER_REPORT.md, descriptor contract and current area Inspect dispatch were reviewed. Old private bridge questions are not an allowed substitute. No shared files were edited.

Upstream cases matched for this rule on this branch: 0. Mutant: not run because no rule is registered. Stop here until the shared question/helper lands.
