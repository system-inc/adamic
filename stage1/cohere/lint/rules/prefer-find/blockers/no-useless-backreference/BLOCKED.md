# no-useless-backreference

Status: blocked on the shared area API/helper, not an implemented descriptor.

Exact upstream call: `reference.ConstantStringIn(ctx, arguments[0])` at `cohere/internal/lint/rules/core/no_useless_backreference.go:118`.

Missing: shared reference constant evaluator/tracker and capture/backreference analysis; no private regex scanner may replace them.

Reproducer: [reproducer.txt](reproducer.txt). The source requires the same ordinary dependencies/import files as the corresponding upstream case.

The shared checker model, CHECKER_REPORT.md, descriptor contract and current area Inspect dispatch were reviewed. Old private bridge questions are not an allowed substitute. No shared files were edited.

Upstream cases matched for this rule on this branch: 0. Mutant: not run because no rule is registered. Stop here until the shared question/helper lands.
