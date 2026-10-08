# no-irregular-whitespace: complete upstream parity after dependency merges

Merged `origin/lint-fix/jsx-inventory-discovery` at `e7c196a9755cbba48809859033158bc9b06f3c7b` and `origin/codex/parser-recovery-land` at `dbe2fff02270fe9356ee560a699e706a8b444105`, in that order, with merge commits and no rebase. The tested implementation commit is `030d88f1`; subsequent changes contain only this rule's report and evidence. No shared code was edited locally.

All **247/247** captured unique upstream source/file/rule/options combinations now match the unmodified Go oracle on source Node, emitted JavaScript and ASan/UBSan native: **69,379 identical bytes**, with no excluded cases. See `evidence/merged-upstream-parity.log`. The three formerly refused cases now match: U+3000 adjacent to é or 中, and U+180E in `var any`. The port retains the exact 24-character set, run grouping, individual line-separator findings, initial-BOM exemption, option defaults and skipped-region boundaries. No fixes or suggestions exist in this rule. No unsupported Adamic construct was needed.

All four owned witnesses and inherited witness/all-rule rows pass across all runtimes (**343,588 identical bytes**). `split_irregular_runs` compiles, runs and is caught on all three runtimes: mutated range 3–4 versus Go's grouped range 3–51. See `evidence/merged-selected.log` and `evidence/merged-mutant.log`. All 77 registered mutants pass their checks.

The registry generator passed, followed by selected `TestRulesAgree`, `TestMutants` and `TestOwnedWitnesses`, then the complete package once. TypeScript was clean at pin `050880ce59e30b356b686bd3144efe24f875ebc8`; WASI SDK 27 was installed; benchmark input was enabled; both profile variables named the same fresh `/tmp/no-irregular-whitespace-merged-whole-profile.40TtkX` directory.

## Complete package

See `evidence/merged-whole.log` and `evidence/merged-whole-summary.json`: **119 pass, 2 fail, 1 skip**, including direct subtests; top-level **33 pass, 2 fail, 1 skip**. Wall time **1,912.0149 seconds**, nproc **5** (CPU quota 4). Load before: **2.63 / 4.14 / 3.39**; after: **2.65 / 2.08 / 2.45**.

Both JSX integration tests, JSX discovery, upstream comparison, node-table linking and sharding now pass. The two remaining failures are outside this rule:

1. **TestCompilerAndStage1Agree**, `stage1/cohere/lint/lint_test.go:467`: the expanded 912-file corpus includes `stage1/typescript/parser/testdata/lint_cases/decorated_async_promise_executor.ts:1`, whose entire input is `new Promise(@dec async () => {})`. Go's oracle rejects it with Expression expected and comma expected because this test supplies no recovery marker. The guard is `stage1/cohere/lint/testdata/oracle.go:186` (overlay path `cohere/adamic_lint_oracle.go:186`). A smallest nonempty reproducer is the single byte `@`, with a plain manifest row, also rejected by that guard. Exact and minimal inputs: `evidence/invalid-corpus-fixture.ts.txt`, `evidence/invalid-corpus-min.ts.txt`; verified output: `evidence/invalid-corpus-min.log`. No inherited-corpus parity certification is claimed past this refusal.
2. **TestProfileSnapshotsAgree**, `stage1/cohere/lint/profile_test.go:225`: first mismatch is case 5006, `stage1/typescript/parser/testdata/lint_cases/wave13_top_level_await_new.ts:1`, input `await new Promise();` followed by `export {};`. The shortest valid reduced form preserving this await/new mismatch is **`await new a`** (11 bytes, no newline). Go reports no `no-new` finding; source Node, emitted JavaScript and sanitized native all report `noNewStatement` at bytes 6–11. Exact/minimal inputs and all-runtime output: `evidence/await-new-fixture.ts.txt`, `evidence/await-new-min.ts.txt`, `evidence/await-new-min.log`. This contains no irregular whitespace and is a shared parser/lint dependency, left unchanged.

The skip is **TestCheckerBridgeRefusalPending**, `stage1/cohere/lint/checker_pending_test.go:49`, awaiting `codex/tsgo-errors-as-values` and the shared `TSGoError` API. Every optional input was set; no failure or skip was suppressed.

Earlier evidence without the `merged-` prefix records the pre-merge 244/247 result and 108/6/1 package run. It is retained as history; current results are above. The upstream header's per-character wording is stale: Go groups runs, as pinned by `TestNoIrregularWhitespaceReportsTheCharacterItself`; this port matches Go.
