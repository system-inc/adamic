# Collapse package: counter helpers and byte-input stop

Package claim: c386192b78ec6fef304ebf5da5f89000f638a2ec, on lint-helpers/rules-tailwind-collapse from origin/area/stage1-lint. The post-push fetch found no competing package claim. This is a partial delivery, not a complete collapse package or a consuming-rule parity claim.

## Delivered helpers

- `rules/tailwind/collapse.nextBuildCount`: `next_build_count.a`, built fresh. One-line comment records the user-authorized number representation: process-local builds remain far below 2^53.
- `rules/tailwind/collapse.BuildsSoFar`: `builds_so_far.a`, built fresh; the read shares `counter_state.a` with the increment.

The two functions execute synchronously within a stage 1 process. They do not provide cross-worker shared memory or a concurrent-worker locking API. Go's mutex-protected functions remain unchanged in the comparison oracle; instrumentation only records the returned count under the existing lock. Captured actions replay every observed return, including interleaved reads, against a fresh Go process and each stage 1 backend.

## Counter output audit and agreement

An audit of all Go production references found `nextBuildCount` assigns `LoadedDesignSystem.BuildCount`. No production consumer reads that field or `BuildsSoFar`; tests use them to check cache/build counts. They do not enter lint findings, messages, ranges, fixes or suggestions.

The initial capture recorded 297 actual calls; the selected helper run captured and compared 295. The helpers-package run captured and compared **293** actual upstream calls on source Node, emitted JavaScript and ASan/UBSan native. Counts vary with the scheduling of Go's concurrent cache tests; no count is treated as a fixed coverage total. The original Tailwind test package passes before the capture is accepted. Every replayed decimal line is byte-identical to Go's `int` output (64-bit on this platform).

Both helpers have one compiling semantic mutant, caught by output comparison on all three backends: increment by two; read the counter plus one. See `evidence/helpers-whole.jsonl.gz` and `evidence/selected.txt.gz`. No compilation error, crash or omitted-input skip receives mutant credit.

## Known bigint gap, informational only

The prior signed-64 support probe remains in `testdata/atomic-counter.a.txt` and its compiler output in `testdata/atomic-counter.compiler.txt`:

`atomic-counter.a:4:10: stage 0 can't lower a function returning bigint yet`

The user's number-counter clarification resolves this helper's representation. This gap is retained for @system_adamic as information, not as a package blocker.

## Remaining input boundary

Stopped helper: `rules/tailwind/collapse.*stylesheetCollector.loadFile`, `cohere/internal/lint/rules/tailwind/collapse/design_system.go:597` (`os.ReadFile`), passed unchanged into `ParseCSS` at line 602. Go preserves arbitrary file bytes. The available stage 1 primitive, `readTextFile`, replaces invalid UTF-8 with U+FFFD, so it cannot provide the same input to the loader.

`TestStylesheetByteInputBoundary` calls the actual Go loader through an oracle-only export and reads its installed theme value. It compares ASCII and Unicode controls on all three backends, then explicitly records three invalid-byte mismatches. The first mismatch is `invalid80`: Go retains byte 128 while all three stage 1 backends re-encode U+FFFD as bytes 239, 191, 189. The subsequent `invalid81` and `invalidFF` cases confirm the same input loss. All programs exit cleanly. These negative observations are not agreement cases, skipped tests, delivered loader code or mutant catches.

The byte-preserving-input API feasibility probe is `testdata/raw-file-input.a.txt`; `testdata/raw-file-input.compiler.txt` records TS2305: module `adamic` has no exported member `readFileBytes`. The source Node/emitted/native input boundary probe and narrow private-Go loader adapter are retained from `origin/codex/lint-helpers-from-codex/lint-wave1-09`; no old pin guard or old success log is used as current evidence. No private filesystem/decoder substitute is delivered.

This boundary prevents claiming the complete package and a complete consuming-rule proof. The rule proof also needs the sibling `rules/tailwind` package, which is not landed on this base and is reserved by `lint-helpers/rules-tailwind`; for example `enforce_consistent_variant_order.go:92` calls `DesignSystemForProgramAt`. No private copy of that shared package is built. No consuming rule was ported and no new rule is credited as unblocked. The six candidate consumers remain: better-tailwindcss/enforce-canonical-classes, enforce-consistent-class-order, enforce-consistent-variant-order, enforce-shorthand-classes, no-conflicting-classes and no-unknown-classes. The remaining retained helper slices are not represented as landed or certified.

## Gates

The root helpers package includes this package through the separately owned `collapse_package_test.go`. The whole helpers package passes 9 tests, 0 failures, 0 skips, in 66.633 seconds. It includes the fresh capture, three-backend counter comparison, both mutants and explicit byte-boundary observations.

The whole lint package passes **116 tests, 0 failures, 1 skip**, in **1890.285 seconds**. The sole skip is `TestCheckerBridgeRefusalPending`, an unconditional pending checker-bridge test, not a missing-input skip. `nproc=5`; load at gate start was `0.26 0.21 0.58`, and at lint completion `1.49 2.00 2.03`. See `evidence/lint-whole-summary.json` and `evidence/lint-whole.jsonl.gz`. Both gates use all requested inputs: clean TypeScript at 050880ce59e30b356b686bd3144efe24f875ebc8, WASI sysroot, ADAMIC_LINT_BENCH=1, and one fresh directory for both profile variables. Inputs and resource measurements are retained under evidence. No generated registry files or other workers' modules are committed.

The user's standing finished-unit push rule arrived during this run. Because the package remains blocked, this counter implementation and its current evidence are committed locally only; no partial implementation push is made. The earlier published claim and blocker commits predate that standing rule.
