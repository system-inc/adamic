# no-irregular-whitespace: port awaiting shared parser recovery

The listener and typed oracle adapter are implemented. The upstream Go rule is unchanged. The exact 24-character set, run grouping, individually reported line separators, leading-BOM exemption, five option defaults, token-only literal spans, template quasis, JSX text spans, comments and hashbangs follow Go. There are no fixes or suggestions in this rule.

247 unique source/file/rule/options combinations were captured by running all `TestNoIrregularWhitespace*` tests through the normal capture overlay. 244 match Go on source Node, emitted JavaScript and ASan/UBSan native: 67,478 identical output bytes. Three cases fail in the shared parser before a rule context exists. This is not full parity certification. All 247 remain in `/tmp/no-irregular-whitespace-capture/manifest.txt`; `/tmp/no-irregular-whitespace-capture/supported-manifest.txt` explicitly excludes only the three documented refusals for the partial comparison.

The four witnesses cover all 24 codepoints, runs and line separators, the initial BOM, multibyte offsets, default literal behavior, all five configured skip regions, a literal comment opener and reportable template interpolation. `regions.options.json` enables the skips while keeping strings reportable. Every witness reports with Go. The selected `TestOwnedWitnesses` passes across all runtimes, including all-rule rows and the inherited witnesses (338,425 identical syntax-output bytes). `evidence/selected.log` records the final witness and mutant passes. `evidence/selected-upstream.log` records the initial upstream comparison and its parser refusal; the final whole-package run retains the same refusal.

`mutant.json` splits a run into individual-character findings. It compiles and runs, and differs from Go on source Node, emitted JavaScript and native; see `evidence/mutant.log`.

## Blocking dependency

No unsupported Adamic language construct was needed: the complete rule graph lowers successfully and runs on native. The blocker is the shared parser's recovery contract, outside this rule's directory. Its semicolon check in `stage1/typescript/parser/statements.ts:79` refuses these Go cases with `parser slice expected semicolon at 8`:

- `TestNoIrregularWhitespaceFindsAnIrregularCharacterAbuttingAMultiByteOne`, `cohere/internal/lint/rules/core/no_irregular_whitespace_test.go:565`: `const \u00e9\u3000x = 1;`.
- The same test at `cohere/internal/lint/rules/core/no_irregular_whitespace_test.go:570`: `const \u4e2d\u3000x = 1;`.
- `TestNoIrregularWhitespaceCoversCharactersUpstreamNeverFails`, `cohere/internal/lint/rules/core/no_irregular_whitespace_test.go:431`: `var any \u180e = 1;`.

Go's recovering parser reports an irregular-whitespace finding for each. All three port runtimes exit 70 before the listener runs. `evidence/blockers.log` records each refusal; `evidence/blockers-go.log` records the unchanged oracle's answers. No source rewriting, alternate parser, private shared-helper copy or recovery exclusion was added to the registered rule or package harness. The package's ordinary upstream comparison still fails on these cases.

Literal skip regions stay in UTF-16 units; comment regions retain the shared shelf's UTF-8 byte contract. Only a candidate finding checked against comments needs conversion. A lone initial BOM does not trigger the tree walk. These equivalent representations avoid quadratic source-prefix rescans on the large inherited corpus.

The shared JSX harness also has a fixed expected case-count table at `stage1/cohere/lint/jsx_integration_test.go:64`. Its expected map lacks this rule's 23 JSX cases, causing `TestJsxLintReleaseAndThroughput` and `TestJsxLintTrees` to fail. The table was not edited under the directory-only scope.

## Complete-package validation

The final selected witness and mutant tests passed, then the final complete lint package ran once with every requested input set. See `evidence/selected.log`, `evidence/whole.log` and `evidence/whole-summary.json`.

The complete run has **108 pass, 6 fail, 1 skip**, counting top-level tests and direct subtests; its top-level counts are **28 pass, 6 fail, 1 skip**. Wall time was **2,254.57 seconds**, on **nproc 5** (cgroup CPU quota 4). Load averages were **0.67 / 1.07 / 1.60** before and **2.60 / 2.65 / 2.71** afterward. TypeScript was clean at `050880ce59e30b356b686bd3144efe24f875ebc8`. WASI SDK 27 was installed. `ADAMIC_LINT_BENCH=1`; both `ADAMIC_LINT_PROFILE_DIR` and `ADAMIC_LINT_PROFILE_SNAPSHOTS` named the same fresh `/tmp/no-irregular-whitespace-profile` directory.

`TestCompilerAndStage1Agree` passed across 874 compiler/stage1 files on Go, Node, emitted JavaScript and sanitized native. `TestThroughput`, all registered mutants, owned witnesses, profile artifact generation and profile compilation passed.

The six failures are shared integration dependencies:

- Parser recovery: `TestRulesAgree`, `TestNodeTableIsLinkOnly`, `TestShardsAgree`, and `TestProfileSnapshotsAgree` reach the three exact upstream cases listed above.
- Fixed JSX count table: `TestJsxLintReleaseAndThroughput` and `TestJsxLintTrees` reject the 23 newly discovered JSX cases.

The one skip is `TestCheckerBridgeRefusalPending`: the landed shared checker declarations do not yet contain `TSGoError` (`stage1/cohere/lint/checker_pending_test.go:49`). Setting every optional input cannot enable this explicitly pending API test. No skip or failure was suppressed.

The upstream header's per-character granularity prose is stale. Implemented Go behavior groups runs, pinned by `TestNoIrregularWhitespaceReportsTheCharacterItself`'s “one finding per run, as ESLint reports it” case; this port follows that behavior.
