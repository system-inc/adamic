# no-invalid-regexp

Status: blocked on the shared area API/helper, not an implemented descriptor.

Exact upstream call: `esregexp.Compile(pattern, compileFlags)` at `cohere/internal/lint/rules/core/no_invalid_regexp.go:263`.

Missing: shared ECMAScript pattern validation with Go error descriptions; dynamic RegExp construction is still refused at internal/lower/regexp.go:39.

Reproducer: [reproducer.txt](reproducer.txt). The source requires the same ordinary dependencies/import files as the corresponding upstream case.

The shared checker model, CHECKER_REPORT.md, descriptor contract and current area Inspect dispatch were reviewed. Old private bridge questions are not an allowed substitute. No shared files were edited.

Upstream cases matched for this rule on this branch: 0. Mutant: not run because no rule is registered. Stop here until the shared question/helper lands.
