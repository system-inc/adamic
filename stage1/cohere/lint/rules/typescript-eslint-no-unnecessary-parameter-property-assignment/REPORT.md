# no-unnecessary-parameter-property-assignment

Reproduces modifier detection, structural shadowing, assignment and function descent boundaries, arrow traversal, logical versus arithmetic assignments and class-initializer suppression. Suggestions remove the reported assignment and Go's blind extra trailing byte. A non-ASCII trailing character is explicitly refused because a UTF-16 edit cannot represent half a UTF-8 character.

Supported upstream fixtures plus the independent witness: **92 files, 63 findings, 56815 identical bytes** across Go cohere, Node source, emitted JavaScript and ASan/UBSan native. The original Go assertions all pass before captured cases are consumed.

The compiler corpus is pinned to `050880ce59e30b356b686bd3144efe24f875ebc8`: 77 compiler sources. All 152 current stage1 `.ts`/`.a` sources are covered, using the initial complete corpus plus the four-file delta where applicable. Full messages, UTF-8 diagnostic ranges, every fix edit, every ordered suggestion and every suggested/fixed source are compared byte for byte.

The mutant **trailing semicolon retained** changes `node.end + 1` to `node.end` in `rule.a`. It compiles and exits 0 with empty stderr on all three Adamic runtimes; only comparison to Go catches it. The raw witness and descriptor are directory-owned.

Findings/s, best of five interleaved count-only runs: native **427.22**, Node **591.76**, Go **2159.53**. The natural corpus has zero findings, so these include 500 copies of the witness. Count=500; files=230.

One upstream fixture is excluded from Adamic comparison: `testdata/unsupported-destructured-parameter.ts.txt`, because the shared parser panics at the `{` in `constructor(public { a } ...)`. Its original Go assertion passes. `testdata/unsupported-byte-boundary.ts.txt` proves the deliberate split-UTF-8 refusal on all three runtimes; a clean compiling mutant disabling that guard is caught only by Go byte comparison.

Shared integration is blocked. `registry/registry.go:95` and `:187` require `rule.ts`; `Finding` cannot represent this rule's independent repair ranges/arrays, and the shared fixer applies diagnostic ranges. The registered factory explicitly refuses this unsupported contract. The complete listener and repairs run through the owned runner, with no shared Adamic files modified. The shared-contract refusal is tested on all three runtimes and a compiling mutant proves it can fail.

Reproduction, logs, full results and environment limits are in `../typescript-eslint-prefer-as-const/COMPLETE_REPORT.md`.
