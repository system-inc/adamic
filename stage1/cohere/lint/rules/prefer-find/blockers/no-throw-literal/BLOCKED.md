# no-throw-literal

Status: blocked on the shared area API/helper, not an implemented descriptor.

Exact upstream call: `ctx.TypeChecker.GetSymbolAtLocation(identifier)` at `cohere/internal/lint/rules/core/no_undef_init.go:163`.

Missing: all identifier declarations and each source IsDeclarationFile bit required by identifierIsShadowed; declarations only accepts ClassDeclaration/InterfaceDeclaration, symbol-origin only exports one file.

Reproducer: [reproducer.txt](reproducer.txt). The source requires the same ordinary dependencies/import files as the corresponding upstream case.

The shared checker model, CHECKER_REPORT.md, descriptor contract and current area Inspect dispatch were reviewed. Old private bridge questions are not an allowed substitute. No shared files were edited.

Upstream cases matched for this rule on this branch: 0. Mutant: not run because no rule is registered. Stop here until the shared question/helper lands.
