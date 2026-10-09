# Verified failures from the fresh gate

## TestJsxLintReleaseAndThroughput

`stage1/cohere/lint/jsx_integration_test.go:85`: Node case 588 line 8719 reports `range 0 9 unusedExpression`, Go reports `range 0 15 unusedExpression` for captured `Duplicate.tsx`, source `<A a /><B a />;`.

Fixture introduced by `react-jsx-no-duplicate-props@e98e9024e`; affected rule `@typescript-eslint/no-unused-expressions` introduced by `lint-rules/wave1-03-ready@2920f4dbf`. This is a recovery/tree interaction, not a duplicate-props finding discrepancy. A smaller reproducer `<A a/><B a/>` still differs: Go one `range 0 12`, Node `range 0 8` and `range 9 12`. Fresh oracle and Node logs retained under repro.

## Named skip

`TestCheckerBridgeRefusalPending`, `checker_pending_test.go:51`: awaits codex/tsgo-errors-as-values: tsgoInspect must return TSGoError from the C error buffer. All required gate inputs were supplied. The test was not altered.

## TestOwnedWitnesses

`stage1/cohere/lint/registration_test.go:76`: Node case 277 line 4747 returns malformed fixed source where Go first says `rejected unicode-bom 0 3  the rewritten file does not parse (TS1128: Declaration or statement expected.)`, then rejects two one-var edits. Fixture `rules/no-irregular-whitespace/testdata/runs.ts.txt` (captured temporary name `no-irregular-whitespace-3.ts`); incoming `lint-landing/no-irregular-whitespace-wave2-02@19795e9e`, interacting with `one-var@b5bca97f8` and the base's unicode-bom rule.

Smaller fresh reproducer: `let a;\u00a0let b` (actual NBSP). Go rejects one-var edits 8..8 and 8..11 in that order with TS1128 and preserves the source. Node returns `let a;\u00a0, b`. Reproduces both with just one-var and with all rules. The defect is the shared fixed-source reparse/rejection in lint.ts, outside this batch's permitted source edits. Fresh logs retained in repro/fixer-Go.log and fixer-Node.log.

## TestJsxLintTrees

`stage1/cohere/lint/jsx_integration_test.go:140`: Node parser `--whole` exits 70: `adamic: panic: source has parse diagnostics; use --recovery to inspect the recovered tree`. The gate failure does not name the first failing captured fixture and its test temporary manifest was removed after failure.

Fresh smallest tested reproducer `<a/><a/>` (8 bytes): Go parser oracle `--whole --jsx-recovery` exits 0 with a recovered BinaryExpression/CommaToken tree; Node parser `--whole` exits 70 with the above diagnostic. The merged `react-jsx-no-duplicate-props@e98e9024e` contributes adjacent JSX (`Duplicate.tsx`, `<A a /><B a />;`) which also exposes this strict/recovery-mode mismatch. This is a shared parser/harness limitation; attribution to that fixture as the *first* failed row is not claimed. Commands, exit codes and complete outputs are in repro/parser-repro.json.

## TestRulesAgree typed count guard

`stage1/cohere/lint/lint_test.go:513`: `42 typed cases come from programs with other fixture files, so they aren't held to upstream's count; the replay lints one file, so decide each one before accepting it`. Capture: 415 under captured options, 4 deliberate existing strict-alone controls, 373 held to upstream counts, 42 not. New rule `@typescript-eslint/prefer-reduce-type-parameter`, incoming `codex/lint-land-wave-23@25bccfc59`. Every upstream case uses RunTypedFiles and includes `/repository/source/class.ts`, not a deliberate Run control. No strict-alone exemption was added.

Shortest captured subject `/repository/source/Reducing.ts`: `[1, 2, 3][null]((sum, num) => sum + num, 0);`; the additional fixture is the Reducable class declared in upstream prefer_reduce_type_parameter_test.go:23. All 42 raw captured program/count records are in prefer-reduce-typed-records.json. This is a multi-file replay/count harness gap and remains red.

TestRulesAgree subsequently compared all syntax/generated rows: Go, Node, emitted JavaScript and native identical at 15,996,279 bytes. Its terminal failure is the above count guard, not a backend discrepancy.
