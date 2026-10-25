Re-cut slice 5 preserves placeholder lowering, parser construction, NodeArray, speculation, named stack stops and the witness ledger.
Delivery is one commit on main d79302e50c022c997c00f7c68ddb7475f9ee65fc; source 2f7f122fa83ea7014b62c34e055e883351d873d4; the handoff supplies the delivery SHA.
Focused tests, 55 differential fixtures, eight ledger witnesses, structural guards, call-target readers and changed-package vet passed; counts regenerated once in 72.394 seconds.
All 27 independent overlay mutants were caught, along with the embedded runtime, flow, rewind, payload and retained-stop mutants in the focused tests.
Not covered: whole-package suites, all transitive consumer suites, the full gate, admission census, Program-region lifetime, optional guard-cost measurements, WASI or non-Linux systems.

The squash started from the fetched main tip. The one conflict was modify/delete on internal/lower/non_null_impossible_test.go. Main c1718f67 deleted the former refusal-only test after an audit replay. This slice changes placeholder result/storage classification, so the conservative choice is to keep its adapted test rather than remove coverage for the new semantics. No production infrastructure from that audit is changed. This is a deliberate test resurrection, identical to the source file. All its cases pass. No other test was deleted or weakened.

Main's newer count machinery is retained. Despite the task's digest-column description, this fetched main stores digests in count-measurement products and keeps numeric rows in counts.md. The file was restored exactly from origin/main before one successful -update-counts invocation. Parser countCase registrations now return (runs, measure), and their assertions invoke measure. Those are the only source-test API adaptations. The recorder changes exactly 41 added rows and 21 existing rows, with no removal or unrelated movement. Every measured row equals the source slice's row, including its before/after attribution; counts-attribution.json lists all 62 rows and reasons. No Mac table was used.

Main's native clang jobserver changes in tsgo.go merge automatically and remain. The named stack guard change is applied around them. Every internal/native/runtime added/deleted line is identical to the source slice's net patch: no differences in any of the 13 runtime files. runtime-comparison.json retains both sequences and the source merge base.

Guards satisfied:

- TestEveryMapAProgramHoldsIsSorted passes. The slice's PlaceholderSources, PlaceholderViews and FactoryFields retain SortedMap; main's existing SortedMaps remain. A plain PlaceholderSources map mutant fails this guard.
- TestRuntimeSetsAreEachUsed passes in 1.02 seconds. No runtime feature set was added and all existing witnesses still turn on exactly their registered set.
- TestNativeAgreementRowsNameRegisteredPrograms passes in 0.01 seconds. Main's literal program registry is retained; this slice adds no unregistered native-agreement call.
- TestEveryTestIsParallelOrSaysWhy passes in 1.18 seconds. Source tests already call t.Parallel first or carry the required reason comment; no exemption was added.
- Products keep main's displayed paths, prefix maps, placeless keys and runtime products. All three source relocation checks pass; independent physical-path mutants fail each check.
- No Go package was added, so compiler-dependencies.json needs no declaration.
- No leaf reached 60 seconds on the 4-CPU quota. The largest observed elapsed time is 14.65 seconds; test-seconds.json records each leaf, including every new test. The source already supplies split top-level tests, so no further split was needed.
- TestCallTargetReaders passes in 9.80 seconds, including cold GoExports setup, with -timeout 90s. Changed-package vet exits 0 with no diagnostics.

Validation used the printed env.sh, GOPROXY=https://proxy.golang.org|direct, GOMAXPROCS=4 and -p=4, preserving -buildvcs=false -trimpath. Test output went directly to files. The focused runner executed 38 bounded commands, requiring every selected top-level test to emit a terminal event. All 55 fixture leaves in eight commands passed uncached source Node, backend Node, native ASan/UBSan, release native and successful-program leak checks. Eight TestSlice5Witness tests passed in 2.799 seconds. Existing stack regressions passed in 3.992 seconds. Flow's new-fixture remainder passed in 8.35 seconds, freshness's in 1.01 seconds; two changed non-null flow corpus units and a JavaScript selector control also pass. The initial freshness subtest selector omitted fixture children, so it is not credited; the corrected remainder covers them. All final substantive commands and selectors are in the runners and JSON logs. Whole packages and the full gate were deliberately not selected.

Setup succeeds: Go ready 0.053s, Node ready 0.059s, clang ready 0.490s, markdown dependencies ready 0.974s, submodules ready 234.881s, shared cache ready 240.722s, Go build ready 512.296s, build cache warm 512.618s, done 512.722s. nproc=5 and cpu.max=400000 100000, a four-CPU quota. Go 1.27.1, clang 20.1.8, Node 24.19.0. The environment file is /workspace/adamic-tools/env.sh. A premature guard attempt reached the unrelated system Go, another preceded module checkout, and a cold shared-cache guard command hit its outer 180s limit before test events. These are not test failures or mutant kills. The completed local-cache run passes. npm ci installed the pinned stage3/api declarations. Setup itself completed normally; later tests unset GOCACHEPROG to use the local cache.

Every independent mutant and its selected catcher follows. All exit nonzero through the intended assertion or observed behavior, not a compiler/build failure. omit-refusal-state is specifically a nil-result initialization guard caught by a Go panic, not a backend behavioral proof.

| Mutant | Selected catcher |
|---|---|
| plain-placeholder-map | `^TestEveryMapAProgramHoldsIsSorted$` |
| placeholder-physical-path | `^TestParserPlaceholderProductsUseDisplayedPaths$` |
| factory-physical-path | `^TestParserFactoryProductsUseDisplayedPaths$` |
| speculation-physical-path | `^TestParserSpeculationProductsUseDisplayedPaths$` |
| branch-union | `^TestParserFactory(Completion.* / ReadBeforeCompletion)$` |
| omit-escape-missing | `^TestParserFactory(Completion.* / ReadBeforeCompletion)$` |
| erase-earlier-read | `^TestParserFactory(Completion.* / ReadBeforeCompletion)$` |
| trust-assertion | `^TestParserFactory(Completion.* / ReadBeforeCompletion)$` |
| slot-pos | `^TestParserNodeArrayLayout$` |
| slot-end | `^TestParserNodeArrayLayout$` |
| slot-hasTrailingComma | `^TestParserNodeArrayLayout$` |
| slot-transformFlags | `^TestParserNodeArrayLayout$` |
| drop-unset-use | `^TestParserConstructionUnsetUse$` |
| drop-speculation-check | `^TestParserSpeculationMisfit$` |
| disable-tiny-stack | `^TestParserStackTinyLimit$` |
| drop-flow-check | `^TestPlaceholderUseCheckSavedLeak$` |
| retain-alias-proof | `^TestPlaceholderUseCheckAliasReset$` |
| nullish-receiver-twice | `^TestNativeAgreesWithNode_[0-9]+$/internal/oracle/testdata/placeholder_nonnull_null_loose[.]a$` |
| json-ordinary-assertion | `^TestNativeAgreesWithNode_[0-9]+$/internal/oracle/testdata/placeholder_nonnull_null_json[.]a$` |
| admit-weak-slot | `^TestPlaceholderWeakSlotStaysNotYet$` |
| omit-refusal-state | `^TestOptionalWideningSpreadOverwrite$` |
| drop-native-nullish-tags | `^TestNativeAgreesWithNode_[0-9]+$/internal/oracle/testdata/placeholder_nonnull_null_copy[.]a$` |
| drop-javascript-union-tag | `^TestDefaultTaggedSourceViewsDefaultBoxedWrite$` |
| unlisted-output | TestUnlistedBackendSourceJavaScriptAgreement: stdout differs |
| drop-terminal-guard | TestTerminalStackStopJavaScript: exit codes differ |
| drop-range-error-kind | backend_stop_catches.a oracle fixture: exit codes differ |
| drop-overflow-message | TestNonOverflowRangeErrorCatch: non-overflow RangeError incorrectly terminal |

Embedded mutant evidence is retained in focused-results.json and its logs: three runtime publication/key/JSON mutants, factory payload replacement, scanner rewind removal, stack guard removal, seven placeholder flow-check removals, null-tag and spread-readiness changes, readiness wrong-zero/erasure/capture/exception/missing-value/message changes, and all eight witness repairs compared against retained stopping artifacts. Their test assertions require the mutant to execute and disagree in the intended way.

Conservative semantic assumption: Outcome 24 remains the bounded heap-backed parser construction, NodeArray, speculation and stack witnesses delivered by the source. Program-region lifetime remains pending. Existing ruled backend divergences and exact declared-type negative witnesses retain their separate contracts. No cohere code was copied, no PR was opened and no Loom submission was made.
