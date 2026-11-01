# 95g evidence

The current implementation is the scoped third-arm fix. `scoped-final/` contains actual built-compiler witnesses, full formatted diagnostics, stacks, source/idempotence/census proofs, mutant failures, apply log and full oracle report/logs. `comparison-summary.json` checks the actual build against the trace-only scoped candidate and preserves all successful query results.

The earlier broad empty-declaration guard is superseded. `empty-guard-hook.cjs`, `original-with-guard.json`, `super-with-guard.json`, `proposed-reader-guard.diff.txt` and `historical-status-before-scoped.json` are historical rejected evidence, not the shipped patch. The proposed reader edit is not applied.

Run scripts with Node 24 and stock TypeScript 6.0.3 in NODE_PATH. `witness.cjs BUILT_TYPESCRIPT_JS --expect-fixed` asserts the two fixes. `super-witness.cjs BUILT_TYPESCRIPT_JS`, `decl-free-index.cjs BUILT_TYPESCRIPT_JS`, `successful-index.cjs BUILT_TYPESCRIPT_JS`, `service-matrix.cjs BUILT_TYPESCRIPT_JS` record successful controls and services. `controls/fixtures.cjs BUILT_TREE` records the 14-program matrix. `harness-caller.cjs BUILT_TYPESCRIPT_JS SOURCE_TREE/src/harness/fourslashImpl.ts` extracts and executes the actual FourSlash assertion method, with range/error-service stubs; this is a method-level probe, not a full FourSlash run.

`caller-audit-source.json` inventories 123 parsed call sites (one testRunner-only). The reader audit in the adaptation README gives path preconditions and the two moved exceptions. `scoped-hook.cjs` instruments only an in-memory baseline bundle for stack tracing; final results come from the real apply/oracle build.
