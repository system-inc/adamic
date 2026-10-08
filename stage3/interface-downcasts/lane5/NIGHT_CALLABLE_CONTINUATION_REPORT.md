Built: five groups certify 22 additional callable pairs, 132 candidate reads and 144 fixtures.
Commits: five separately pushed group commits listed below; this report commit records final validation.
Checks: original declarations/read spans and alias checks, Node, native release/sanitized, JavaScript, leaks, restored focused tests and measured lane counts pass; whole-table counts fails outside lane.
Mutants: 34 runtime runs catch 150 pair-level assertions; one original-alias mutation caught; all sources restored.
Uncovered: 156/2818 pairs and 2699/11063 candidate reads certified; 2662 pairs and 8364 reads remain; Set and arrow remain deferred.

Reporting date October 12. Continued from e7297f1115a4035d454329914c3435a63d821622 on codex/views-callables. Original TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. These are member-contract fixtures with complete original declarations and member read expressions. Adjacent AST carriers, producer/helper bodies and surrounding application data are reduced. Candidate reads are inventory counts; full original compiler execution and exact reached read counts remain unmeasured.

| Group | Commit | Pairs / reads | Fixtures | Runtime runs / assertions |
|---|---|---:|---:|---:|
| array-bundle-export-import-spread-types | 7d406f6c | 7 / 42 | 45 | 7 / 49 |
| initializer-import-catch | d5d8a5eb | 3 / 18 | 21 | 7 / 21 |
| reflect-try-update-system | 3729c616 | 7 / 42 | 46 | 7 / 47 |
| info-duration | ae16a660 | 2 / 12 | 11 | 6 / 12 |
| class-function-create-update | 4810c561 | 3 / 18 | 21 | 7 / 21 |

Certified ranks: 274,275,278,286,287,288,289,290,293,295,296,297,298,299,300,301,303,304,307,315,316,325. Original GetCanonicalFileName remains a callable property alias at Info's original helper argument read, verified separately in core.ts. Factory families retain method signatures and original receiver/read contexts. Optional arrays, undefined results, scalar/object/undefined names, optional initializer/receiver arguments and required function-expression bodies are exercised. Descendant parameter/result reads remain lazy. No production compiler/runtime change or calling-convention guard removal remains.

Each runtime mutation and affected family is recorded in the five group reports and raw logs. Native and JavaScript arity, result and parameter omissions are caught at the viewed member read. Omitted payload-read registration is caught by descendant string-read assertions with sanitizer evidence. No compilation failure is counted as a runtime mutant result. Changing Info's fixture alias parameter from string to number fails the original alias verifier; restored evidence is recorded in INFO_DURATION_REPORT.md. The first initializer/import/catch reduced bodies needed supported expressions and corrected Node expectations before certification. The first EmitHelperFactory source lookup needed its actual factory/emitHelpers.ts declaration file; the final lookup verifies its complete original declaration. These corrections do not count as mutant results.

Set remains deferred and needs collection-to-view conversion plus recorded intrinsic callable signatures, including add's self result. Arrow rank 120 remains deferred and needs the original ConciseBody union/read contract. The original IdentifierNameMap.toKey is a static class method, DebugType is an intersection alias, Version is nominal, and createNewExpression rank 294 is a destructuring method read; each still needs its original receiver/calling context preserved and proved. Higher-order AnyFunction/hook, generic/assertion/predicate, collection/Date intrinsic, original broad union, optional-host binding/condition and remaining tagged factory families remain pending in the group reports. Untested families have no newly observed refusal claimed. No integrator Union exception decision was received. Narrower literal-subtype producer admission and recursive original AST contracts remain outside these reduced carrier certificates.

Commands and outputs for each group are in ARRAY_BUNDLE_EXPORT_IMPORT_SPREAD_TYPES_REPORT.md, INITIALIZER_IMPORT_CATCH_REPORT.md, REFLECT_TRY_UPDATE_SYSTEM_REPORT.md, INFO_DURATION_REPORT.md and CLASS_FUNCTION_CREATE_UPDATE_REPORT.md. Each group ran the original declaration/read verifier, uncached focused family oracle, executable mutant runner, restored uncached oracle, required TestCountsAreRecorded updater and lane TestCheckedViewCallableCounts updater, all to log files. The Info group also ran original alias verification and its mutation. All five required whole-table updaters fail only outside-lane fixtures; all lane updaters pass. No whole-package tests or full gate.

Final commands:

```sh
source /workspace/adamic-tools/env.sh
go vet ./internal/oracle > /tmp/lane5-night-final-vet.log 2>&1
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original "$(cat /tmp/lane5-night-final-ranks.txt)" > /tmp/lane5-night-final-fixtures.log 2>&1
node stage3/interface-downcasts/lane5/verify-tagged-callable-carriers.cjs /workspace/scratch/lane5-original 122,141,144,145,159,160,176,177,192,194,212,217,241,249,278 > /tmp/lane5-night-final-carriers.log 2>&1
gofmt -l internal/oracle/checked_views_callable_later_ranked_test.go internal/oracle/checked_views_callable_counts_test.go
git diff --check
```

Verified 770 fixtures retain complete original declarations and reads.
Verified 101 fixtures retain requested original aliases, discriminators and enum values.

Oracle vet, formatting and whitespace checks pass without output. The previous counts rows are unchanged and 144 rows are appended. Source evidence covers 139 members and 1629 candidate reads including the two excluded optional-host probes. Static totals stay 4/308 pairs and 34/1503 reads certified, 304/1469 remaining. Setup reused: done 417.425s, nproc 5, quota 4 CPUs; complete timing lines remain in earlier reports. Only codex/views-callables pushed; no PR opened.

Full group SHAs:
7d406f6c1f0f2249b81f5fe4f1fdb6124ef9be43
d5d8a5eb25014ad7d3b4ee9c98aad4e50dc9ce51
3729c616e5add52c1c1addab253717fe3754a562
ae16a6605f0f0f2f16e8d2fc14b75ca7f912346e
4810c561d96b3b3c2fb2bf83657be4f9e9a3cbf5
