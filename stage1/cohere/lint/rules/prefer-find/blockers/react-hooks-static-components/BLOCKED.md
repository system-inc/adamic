# react-hooks/static-components

Status: blocked on the shared area API/helper, not an implemented descriptor.

Exact upstream call: `high_level_intermediate_representation.ForFunction(ctx, functionNode)` at `cohere/internal/lint/rules/react/static_components.go:120`.

Missing: shared native HIR/capture/phi taint and compilation-unit analysis; RuleContext has no shared HIR provider.

Reproducer: [reproducer.txt](reproducer.txt). The source requires the same ordinary dependencies/import files as the corresponding upstream case.

The shared checker model, CHECKER_REPORT.md, descriptor contract and current area Inspect dispatch were reviewed. Old private bridge questions are not an allowed substitute. No shared files were edited.

Upstream cases matched for this rule on this branch: 0. Mutant: not run because no rule is registered. Stop here until the shared question/helper lands.
