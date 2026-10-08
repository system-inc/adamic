# no-extra-non-null-assertion

Sees through parenthesized wrappers and deletes only the redundant bang. Optional-chain detection checks the actual `?.` token on the immediate parent link, rather than the propagated optional-chain flag.

Supported upstream fixtures plus the independent witness: **20 files, 12 findings, 5447 identical bytes** across Go cohere, Node source, emitted JavaScript and ASan/UBSan native. The original Go assertions all pass before captured cases are consumed.

The compiler corpus is pinned to `050880ce59e30b356b686bd3144efe24f875ebc8`: 77 compiler sources. All 152 current stage1 `.ts`/`.a` sources are covered, using the initial complete corpus plus the four-file delta where applicable. Full messages, UTF-8 diagnostic ranges, every fix edit, every ordered suggestion and every suggested/fixed source are compared byte for byte.

The mutant **extra bang deleted from wrong position** changes `node.end - 1` to `node.end - 2` in `rule.a`. It compiles and exits 0 with empty stderr on all three Adamic runtimes; only comparison to Go catches it. The raw witness and descriptor are directory-owned.

Findings/s, best of five interleaved count-only runs: native **840.64**, Node **1124.90**, Go **4345.51**. The natural corpus has zero findings, so these include 500 copies of the witness. Count=1000; files=230.

Shared integration is blocked. `registry/registry.go:95` and `:187` require `rule.ts`; `Finding` cannot represent this rule's independent repair ranges/arrays, and the shared fixer applies diagnostic ranges. The registered factory explicitly refuses this unsupported contract. The complete listener and repairs run through the owned runner, with no shared Adamic files modified. The shared-contract refusal is tested on all three runtimes and a compiling mutant proves it can fail.

Reproduction, logs, full results and environment limits are in `../typescript-eslint-prefer-as-const/COMPLETE_REPORT.md`.
