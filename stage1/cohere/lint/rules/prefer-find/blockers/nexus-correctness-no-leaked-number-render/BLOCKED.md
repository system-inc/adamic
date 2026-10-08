# nexus/correctness-no-leaked-number-render

Status: blocked on the shared area API/helper, not an implemented descriptor.

Exact upstream call: `part.AsLiteralType().Value().(fmt.Stringer)` at `cohere/internal/lint/rules/nexus/correctness_no_leaked_number_render.go:220`.

Missing: literal value and held-value metadata, including enum members; TypeToString cannot substitute for the actual literal value.

Reproducer: [reproducer.txt](reproducer.txt). The source requires the same ordinary dependencies/import files as the corresponding upstream case.

The shared checker model, CHECKER_REPORT.md, descriptor contract and current area Inspect dispatch were reviewed. Old private bridge questions are not an allowed substitute. No shared files were edited.

Upstream cases matched for this rule on this branch: 0. Mutant: not run because no rule is registered. Stop here until the shared question/helper lands.
