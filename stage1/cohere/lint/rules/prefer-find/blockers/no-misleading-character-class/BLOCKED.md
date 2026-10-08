# no-misleading-character-class

Status: blocked on the shared area API/helper, not an implemented descriptor.

Exact upstream call: `reference.NewTracker(ctx.SourceFile, ctx.TypeChecker, nil)` at `cohere/internal/lint/rules/core/no_misleading_character_class.go:187`.

Missing: shared reference tracker and constant/RegExp analysis; no native shared RuleContext helper exports those operations.

Reproducer: [reproducer.txt](reproducer.txt). The source requires the same ordinary dependencies/import files as the corresponding upstream case.

The shared checker model, CHECKER_REPORT.md, descriptor contract and current area Inspect dispatch were reviewed. Old private bridge questions are not an allowed substitute. No shared files were edited.

Upstream cases matched for this rule on this branch: 0. Mutant: not run because no rule is registered. Stop here until the shared question/helper lands.
