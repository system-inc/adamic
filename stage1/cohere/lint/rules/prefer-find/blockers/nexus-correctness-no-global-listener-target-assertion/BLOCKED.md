# nexus/correctness-no-global-listener-target-assertion

Status: blocked on the shared area API/helper, not an implemented descriptor.

Exact upstream call: `ctx.TypeChecker.GetNonNullableType(known)` at `cohere/internal/lint/rules/nexus/correctness_no_global_listener_target_assertion.go:257`.

Missing: GetNonNullableType and default-library/symbol/declaration identity questions; the area bridge has no nonnullable-shape or declaration details question.

Reproducer: [reproducer.txt](reproducer.txt). The source requires the same ordinary dependencies/import files as the corresponding upstream case.

The shared checker model, CHECKER_REPORT.md, descriptor contract and current area Inspect dispatch were reviewed. Old private bridge questions are not an allowed substitute. No shared files were edited.

Upstream cases matched for this rule on this branch: 0. Mutant: not run because no rule is registered. Stop here until the shared question/helper lands.
