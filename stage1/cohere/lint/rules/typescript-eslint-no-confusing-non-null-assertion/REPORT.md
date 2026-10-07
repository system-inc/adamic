# no-confusing-non-null-assertion

Checks precisely `=`, `==`, `===`, `in` and `instanceof`. The full ordered suggestion arrays preserve remove-versus-wrap behavior, two independent insertion edits for wrapping, exact operator interpolation and the diagnostic's full binary-expression range.

Supported upstream fixtures plus the independent witness: **29 files, 12 findings, 4975 identical bytes** across Go cohere, Node source, emitted JavaScript and ASan/UBSan native. The original Go assertions all pass before captured cases are consumed.

The compiler corpus is pinned to `050880ce59e30b356b686bd3144efe24f875ebc8`: 77 compiler sources. All 152 current stage1 `.ts`/`.a` sources are covered, using the initial complete corpus plus the four-file delta where applicable. Full messages, UTF-8 diagnostic ranges, every fix edit, every ordered suggestion and every suggested/fixed source are compared byte for byte.

The mutant **wrap close parenthesis changed** changes `end, end, ')'` to `end, end, ']'` in `rule.a`. It compiles and exits 0 with empty stderr on all three Adamic runtimes; only comparison to Go catches it. The raw witness and descriptor are directory-owned.

Findings/s, best of five interleaved count-only runs: native **413.55**, Node **583.67**, Go **2355.71**. The natural corpus has zero findings, so these include 500 copies of the witness. Count=500; files=230.

Shared integration is blocked. `registry/registry.go:95` and `:187` require `rule.ts`; `Finding` cannot represent this rule's independent repair ranges/arrays, and the shared fixer applies diagnostic ranges. The registered factory explicitly refuses this unsupported contract. The complete listener and repairs run through the owned runner, with no shared Adamic files modified. The shared-contract refusal is tested on all three runtimes and a compiling mutant proves it can fail.

Reproduction, logs, full results and environment limits are in `../typescript-eslint-prefer-as-const/COMPLETE_REPORT.md`.
