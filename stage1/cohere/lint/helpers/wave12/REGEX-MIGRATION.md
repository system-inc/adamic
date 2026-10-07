Built: replaced the earlier hand-written Tailwind normalizer with the shared table's four JS RegExp literals.
Commits: table provenance b39979305d880d7d05083704b73d405cd3694d68; nested-label delivery 4b9465715c6c392fff27902ecc1580206f024b65; this regex update follows on the same own branch.
Checks: actual Go, source Node, emitted JavaScript and ASan/UBSan native match 9,364 observations / 309,422 bytes; focused suite PASS 9.358s and vet clean.
Mutants: namespace-suffix row 8, newline-crossing row 168 and nested-key-join row 12; every mutant compiles and exits cleanly before actual-Go comparison catches it on all three backends.
Uncovered: no arbitrary regex-option translation, invalid UTF-8 or full rule integration claim; the existing counter and shared rule-harness blockers remain.

The user's effective-now regex instruction supersedes the prior manual equivalent. normalizeValueFunctionArgument now declares the exact four table literals, with gu flags, once at module scope. It applies string replacement in the original Go order and preserves the final namespace guard. No hand-written regex matcher remains in this helper. The rows for tailwind/collapse/utility_nodes.go:131, :138, :140 and :142 are retained in testdata/normalization_regex_rows.json, including actual Go patterns and JS translations. This is a copy of the table's translations into the owned helper, not a change to the shared translation table.

The actual Go replacement contract is tested, including capture-group substitution, ASCII whitespace rather than JS Unicode whitespace, lazy groups that cannot cross LF, escaped stars, repeated suffixes, Unicode and malformed controls. Every expected result still comes from the actual Go helper through an export overlay. The corpus retains the original six consumer fixture sources and real helper/engine arguments; see README.md for the exact consumer capture boundaries. All successful outputs preserve every UTF-16 unit and order. No finding is constructed here, so no byte-to-UTF-16 position conversion occurs inside the helper.

The namespace mutant is unchanged. The newline mutant now changes only the first lazy capture from [^\n]*? to [\s\S]*?, and the nested-key mutant removes the star insertion in the replacement string. Each compiles successfully on native and emitted JavaScript, runs successfully with no stderr, then differs only under Go output comparison. No compiler rejection is counted as a caught semantic mutant.

The complete owned package passed 105.859s immediately before this isolated helper rewrite; all other six helpers are unchanged. The focused normalization suite was rerun after the rewrite, followed by full owned-package vet. The normalizer is imported only by its own observation entry main.a, so repeating unrelated helper oracles is not needed for this isolated edit. This is scoped verification, not a new complete package run or repository gate claim.

Commands after sourcing /workspace/adamic-tools/env.sh:

```sh
go test ./stage1/cohere/lint/helpers/wave12 -run '^TestNormalization' -count=1 -v > /tmp/wave12-regex-normalization.log 2>&1
go vet ./stage1/cohere/lint/helpers/wave12 > /tmp/wave12-final-helper-vet.log 2>&1
```

Final fetch still reports origin/main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06. Both own branches contain that main, and both are pushed to their own branch names. The rule branch remains 7f99254378ee9fd8b50f4cfd9aa18cebe5291b2d with its supported-domain oracle green and shared registration/profile/extraction blockers named in PARKED.md. The harness ab70f38d4 is not on main yet. No shared compiler, harness or readiness file is edited, and no main or area push occurs.

The two new CFG helpers add 60,686 successful observations / 246,314 identical bytes and six new compiling mutants to the previously delivered five-helper set. Four consumers each lose two more helper dependencies, zero final blockers. Together the final scoped checks cover all seven delivered helper contracts and 20 existing/new negative controls. The other 11 controls are bigint hex-digit-value/negative-zero/invalid-fallback, breakable missing-switch/nil-loop, block alias-slot, segment drop-final-empty/pop-unmatched-closer/escape-disabled/separator-guard and the number-counter approximation; all are retained in labels-full.log. Semantic controls are caught only by external Go comparison; the separator guard is an explicit unsupported-domain refusal comparison.

The exact nextBuildCount support probe still refuses at gaps/atomic-counter.a:4:10 with stage 0 can't lower a function returning bigint yet. No production counter or concurrent/native exact-counter claim is made. Other helpers remain unclaimed; no exhaustion claim and no ninth helper reservation are made at this blocker boundary. Full CFG findings/fixes, incomplete JSX/Tailwind candidates, default shared registration/profile gate and the full repository gate remain outside the passing claim. Setup ready times were 0s for Go/clang/Node/submodules, cache warm 62s, total 62s; nproc 5.
