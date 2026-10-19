Re-cut slice 4 onto newest fetched main, retaining the package-split reader checks and adding oracle coverage.
Commit: one commit on compiler/loom-r-slice-4-ac-m2; exact SHA is in the handoff.
Commands/results: full touched packages, focused readers, guards, vet, gofmt and 84/84 a-check cases; detailed actual exits and recovery runs below.
Mutants: generator ownership, union layout, oracle reader registration, two a-check headers and fixture classification are caught by their intended assertions.
Not covered: Loom/ahra, the full repository gate, optional benchmark/ledger inputs or every historical standalone mutation script.

Base: 561aeb4b851ca0820f39440a09dc5fb3fd27ba40. Source own commit: fbf32fa6b4ac9feb3ea0c21d90c14acbfcf054eb. Applied with cherry-pick --no-commit, corroborated by merge-tree --write-tree --merge-base=<source>~1 origin/main <source>. No main-to-source diff was used. No cohere implementation was copied; the pinned submodule remains referenced.

Conflicts, both in internal/ir/call_targets_guard_test.go:

1. Package filter: main's 67733635 split checks by selected package, while the source adds oracle. Keep main's base != selected filter and extend the allowed package set with oracle. No main filter is removed.
2. Export factory: main builds one selected package plus leakcheck; the source requested one combined native/lower/fresh/flow/oracle export group. Retain main's per-package factory to preserve test grain. Add TestCallTargetExportInputsOracle, TestProduct_CallTargetReaderArchivesOracle and TestCallTargetReadersOracle, matching main's existing preparer/product/reader pattern. Oracle's nine precise exemptions remain scoped by file, function and field. The ordinary control passes and removing its new iteration-dispatch exemption fails with unapproved call-target read.

Main's newer callback adaptation in object.go and its other clean automatic merges remain intact. Source-owned changes are applied on top. No main compiler behavior is deliberately reverted.

Additional integration work found by the full stage3/fixtures suite:

- Main now requires a literal top-level test for every fixture directory. The source introduces generics, iteration and iteration-dispatch without those registrations or status manifests. Added three direct parallel wrappers using main's unchanged testFixtureDirectory helper, plus manifests measured from the 50 existing source programs on Node and this compiler. No program body changed. Both upstream reductions and authored controls are labeled truthfully in provenance.
- The existing objects/25_map_generator.a advances from a generator Refused stop to an explicit iterable-view NotYet stop. Only that entry's stage0 classification and measured diagnostic changed; its source, provenance and Node golden remain unchanged.
- The first manifest pass exposed the CLI's adamic: prefix and absolute paths; normalized those to the existing helper's repository-relative diagnostic format. Reasons, source locations and Node observations were preserved. The complete fixture suite then passed, 922 terminal checks, 79.737 package seconds. New top-level test times: iteration 0.18s; iteration-dispatch 0.22s; generics 0.22s. Nested fixture/native leaves also stay under 60s.

Assumptions: keep main's per-package export infrastructure and add an equivalent oracle group; the three new fixture directories belong to the same unchanged stage3 fixture runner, not an exception in its census. Preserve unsupported Iterable views as NotYet. Full touched-package validation supersedes the earlier focused-only instruction; consumers outside those packages use focused checks or explicit compile-only coverage.

Runtime clearance: there are no internal/native/runtime/*.c or *.h changes, exactly as in the source own patch. Additional runtime line list: empty (runtime-audit.json).

Counts: restored the entire main table before regeneration, never merged the generated rows. One completed Linux clang 20.1.8 regeneration: go test ./internal/oracle -run TestCountsAreRecorded -count=1 -json -timeout 600s -parallel 4 -args -update-counts, exit 0; 91.985 package seconds, 97.982 command seconds. No leaf exceeded 60s. 60 rows changed: all 59 source-own changed rows have exactly the source's regenerated values, plus main's newer miscompile_2a/controls.a. That new fixture contains PlainIterator. Exact-main versus re-cut C in controls-emission.patch shows two protocol binders, each allocating a closure and two cells, accounting for six added allocations. The recorded row moves 33/32/34/56/10/1 to 39/39/37/58/13/0. Its full Node/native/oracle coverage passes. This is an interaction with new main input, not unrelated table churn. The new status manifests do not add programs to oracle's count fixture inventory, so no second regeneration was needed.

Every changed row (A/F/retains/releases/peak/in-regions):

| Fixture | Main | Re-cut |
|---|---|---|
| internal/fresh/testdata/regexp_tree.ts | 91/91/81/85/35/0 | 91/91/85/89/35/0 |
| internal/oracle/testdata/class_wrong_output_refused/iterators_hidden_return.a | absent | 28/28/16/20/8/0 |
| internal/oracle/testdata/class_wrong_output_refused/iterators_override_this.a | absent | 57/57/31/49/9/0 |
| internal/oracle/testdata/generators/basic.a | absent | 30/30/85/101/14/0 |
| internal/oracle/testdata/generators/cancellation.a | absent | 23/23/65/74/14/0 |
| internal/oracle/testdata/generators/captures.a | absent | 31/31/95/95/24/0 |
| internal/oracle/testdata/generators/checker.a | absent | 20/20/104/93/16/0 |
| internal/oracle/testdata/generators/close.a | absent | 27/27/57/59/13/0 |
| internal/oracle/testdata/generators/completions.a | absent | 59/59/166/175/38/0 |
| internal/oracle/testdata/generators/default_error.a | absent | 13/13/31/30/12/0 |
| internal/oracle/testdata/generators/defaults.a | absent | 29/29/104/107/21/0 |
| internal/oracle/testdata/generators/delegate.a | absent | 89/89/316/332/51/0 |
| internal/oracle/testdata/generators/delegate_throw.a | absent | 28/28/158/151/22/0 |
| internal/oracle/testdata/generators/first_reference.a | absent | 17/17/73/72/12/0 |
| internal/oracle/testdata/generators/map_iterator.a | absent | 24/24/95/103/15/0 |
| internal/oracle/testdata/generators/owned.a | absent | 32/32/132/130/21/0 |
| internal/oracle/testdata/generators/reentrant.a | absent | 14/14/53/54/11/0 |
| internal/oracle/testdata/generators/throw.a | absent | 16/16/75/74/12/0 |
| internal/oracle/testdata/generic_body_indexed.a | absent | 2/2/0/2/2/0 |
| internal/oracle/testdata/generic_body_read.a | absent | 10/10/0/10/3/0 |
| internal/oracle/testdata/miscompile_2a/controls.a | 33/32/34/56/10/1 | 39/39/37/58/13/0 |
| internal/oracle/testdata/notyet_element_access/nodearray.a | absent | 0/0/0/0/0/0 |
| internal/oracle/testdata/notyet_element_access/sorted.a | absent | 0/0/0/0/0/0 |
| internal/oracle/testdata/notyet_element_access/template.a | absent | 0/0/0/0/0/0 |
| internal/oracle/testdata/regexp.a | 437/437/382/417/58/0 | 437/437/383/418/58/0 |
| internal/oracle/testdata/user_iterators.a | 663/663/455/904/64/0 | 1050/1050/677/994/97/0 |
| internal/oracle/testdata/user_iterators_rest_tdz.a | 5/0/5/3/5/0 | 11/3/7/3/8/0 |
| stage3/fixtures/generics/01_identity.a | absent | 3/3/3/7/3/0 |
| stage3/fixtures/generics/02_optional_return.a | absent | 5/5/3/9/3/0 |
| stage3/fixtures/generics/03_callback_return.a | absent | 8/8/3/11/3/0 |
| stage3/fixtures/generics/05_nested_class.a | absent | 8/8/5/16/7/0 |
| stage3/fixtures/generics/09_recursive_optional.a | absent | 3/3/1/4/3/0 |
| stage3/fixtures/generics/10_identifier_multimap.a | absent | 11/11/27/36/9/0 |
| stage3/fixtures/generics/14_optional_literal_union.a | absent | 3/3/0/3/3/0 |
| stage3/fixtures/generics/15_array_callback.a | absent | 18/18/7/27/4/0 |
| stage3/fixtures/iteration-dispatch/body_throw.a | absent | 22/22/17/24/12/0 |
| stage3/fixtures/iteration-dispatch/break.a | absent | 20/20/11/18/10/0 |
| stage3/fixtures/iteration-dispatch/cached_next.a | absent | 29/29/17/34/12/0 |
| stage3/fixtures/iteration-dispatch/class.a | absent | 22/22/11/22/7/0 |
| stage3/fixtures/iteration-dispatch/close_throw_body.a | absent | 22/22/20/26/12/0 |
| stage3/fixtures/iteration-dispatch/close_throw_break.a | absent | 20/20/16/21/10/0 |
| stage3/fixtures/iteration-dispatch/completion.a | absent | 20/20/10/20/11/0 |
| stage3/fixtures/iteration-dispatch/done_throw.a | absent | 13/13/14/16/10/0 |
| stage3/fixtures/iteration-dispatch/labeled_continue.a | absent | 44/44/22/38/12/0 |
| stage3/fixtures/iteration-dispatch/method_getters.a | absent | 26/26/16/24/15/0 |
| stage3/fixtures/iteration-dispatch/next_throw.a | absent | 12/12/13/16/7/0 |
| stage3/fixtures/iteration-dispatch/object.a | absent | 26/26/14/28/11/0 |
| stage3/fixtures/iteration-dispatch/return.a | absent | 21/21/11/18/13/0 |
| stage3/fixtures/iteration-dispatch/value_throw.a | absent | 13/13/14/16/10/0 |
| stage3/fixtures/iteration/array_view_stress.a | absent | 34/34/50/72/16/0 |
| stage3/fixtures/iteration/array_view_weak.a | absent | 5/5/8/13/4/0 |
| stage3/fixtures/iteration/collections.a | absent | 35/35/24/47/11/0 |
| stage3/fixtures/iteration/delegated_generator.a | absent | 20/20/70/74/15/0 |
| stage3/fixtures/iteration/generator.a | absent | 14/14/33/35/13/0 |
| stage3/fixtures/iteration/object_iteration.a | absent | 5/5/1/6/3/0 |
| stage3/fixtures/iteration/strings.a | absent | 20/20/14/27/9/0 |
| stage3/fixtures/iteration/test262_array_views.a | absent | 11/11/25/38/11/0 |
| stage3/fixtures/iteration/user_forwarding.a | absent | 22/22/7/17/14/0 |
| stage3/project-references-source/fixture/chain/app/main-print.a | absent | 2/2/0/2/2/0 |
| stage3/project-references-source/fixture/two/app/main.a | absent | 2/2/0/2/2/0 |

Full package commands all use -count=1 -json -parallel 4 -p 1 with GOMAXPROCS=4, direct log files and hard limits. results.json preserves exact commands, exits, durations and failures:

- counts: exit 0, 97.982s; 1198 passed and 0 skipped.
- ir-prepare: exit 1, 39.659s; 3 passed and 0 skipped.
- internal_lower: exit 1, 40.601s; 1456 passed and 3 skipped.
- internal_ir: exit 1, 2.934s; 24 passed and 0 skipped.
- ir-prepare: exit 0, 4.746s; 2 passed and 0 skipped.
- internal_lower: exit 0, 44.521s; 1500 passed and 3 skipped.
- internal_ir: exit 0, 2.984s; 30 passed and 0 skipped.
- internal_flow: exit 0, 69.941s; 922 passed and 0 skipped.
- internal_fresh: exit 0, 17.583s; 955 passed and 0 skipped.
- internal_javascript: exit 0, 0.569s; 7 passed and 0 skipped.
- internal_load: exit 0, 10.517s; 86 passed and 0 skipped.
- stage3_project-loader_probe: exit 0, 0.164s; 0 passed and 0 skipped.
- stage3_fixtures: exit 1, 74.934s; 729 passed and 1 skipped.
- internal_native: exit 1, 305.142s; 800 passed and 92 skipped.
- internal_oracle: exit 0, 550.667s; 5224 passed and 291 skipped.
- guards: exit 0, 1.924s; 2 passed and 0 skipped.

Original nonzero commands are retained rather than relabeled green. Their recoveries:

- Root filled while export groups were copied. The source counts regeneration had already completed successfully. Flow and oracle export preparers and several native-consuming lower tests failed only while writes reported no space left on device. Repeated the affected preparers and whole lower package after recovery; they passed. No completed test exceeded 60 seconds or was rerun after exceeding its leaf budget.
- Reversibly relocated obsolete generated export products into a named scratch area, then removed those obsolete scratch copies to reclaim space. Active current-tree products, source and evidence were preserved. Used main's supported ADAMIC_BUILD_OUTPUT_DIR with the original cache as read fallback. A runtimeproducts export census initially filled scratch as well; its preparer and three affected tests passed after obsolete scratch data was removed. This changes no recipe or identity rule.
- GCC portability lacked sanitizer/allocator_interface.h. Ambient C_INCLUDE_PATH is intentionally stripped by the native builder, so that first retry also failed. A scratch gcc wrapper invokes the real /usr/bin/gcc with only an explicit -isystem directory exposing the house public sanitizer headers. Actual GNU compilation, execution, ASan, UBSan and leak checks pass. No diagnostic or compiler result is faked.
- The unconditionally tested wasm32-wasi attributed-binding case required WASI_SYSROOT. cloud/setup.sh --wasi-sdk installed SDK 27; the real WASI retry passed. Native clang remains house 20.1.8.
- The initial stage3 directory/old-generator-classification failures and the first manifest path-format failures were fixed as described above. Full stage3/fixtures/... then exited 0. Full native runtimeproducts and core failed cases all have later passing results.

Full oracle: exit 0, 550.667 command seconds, 5,224 passed checks and 291 built-in skips. Full native original command: 800 passes, 92 skips, with only the documented SDK/header/disk cases failing and passing targeted recovery. The merged per-test evidence in validation-summary.json retains each log, all original command results, latest terminal actions and any unresolved failures; no ordinary failure remains after recovery.

Guards: the full IR suite includes every map-held-by-Program sorted guard and the new oracle call-target shard. Full lower includes literal nativeAgreementPrograms registration. Full oracle includes exact runtime-feature-set witness checks. cmd/adamic-gate parallel and duplicate-product declaration guards passed. Compiler-dependency census passed; stage3/project-loader/probe remains declared in cloud/fast-gate/compiler-dependencies.json. Reader guards and vet were repeated after adding the fixture tests.

Mutation proof:

- generator-owner: omit noteLocal for the complete generator parameter; TestGeneratorDestructuredParameterCycle fails with got <nil>, intended cycle-refusal assertion.
- union-layout: disable main's retained narrower-literal union-layout refusal; TestFX5NumberLiteralRefused fails with got <nil>.
- oracle-reader: remove only the iteration-dispatch exemption; TestCallTargetReadersOracle fails with its exact unapproved call-target read.
- removed-header and wrong-reason: body-identical scratch copies still receive the real compiler's Refused diagnostic, and unchanged pinned Gate.aCheck exits 1 for the wrong metadata. No source mutation races package tests.
- fixture classification: a scratch copy of generics/04_function_value.a with its manifest incorrectly claiming Compiles fails main's unchanged gap assertion while source Node still agrees. The original absent-directory registration state was also observed to fail main's unchanged directory census.
- Existing in-suite lowering, oracle, ownership, sanitizer and backend mutants ran as part of the full packages. Fresh overlay sources are .go.txt, never compilable package files. The mutation log classifier was corrected to decode Go JSON's escaped angle brackets; the already-caught generator mutation was reclassified from its existing actual diagnostic, not rerun.

A-check: README-pinned unchanged Gate.aCheck, 84/84 accepted; 63 checked, five NotYet, 16 refused. All changed .a paths, including historical review inputs, were selected by git diff HEAD without directory exclusions. No .a body was changed by the re-cut resolution.

Toolchain: GOPROXY=https://proxy.golang.org|direct, ADAMIC_GOCACHE_OFF=1, bash cloud/setup.sh, source /workspace/adamic-tools/env.sh. Go 1.27.1, Node 24.19.0, Linux clang 20.1.8. nproc=5, cpu.max=400000 100000 (four CPUs). Setup succeeded in 38.900s: Node 0.025s, Go 0.026s, submodules 0.064s, markdown dependencies 0.075s, clang 0.167s, shared cache off 0.168s, Go build 38.682s, test binaries deferred 38.872s, cache warm 38.874s. Optional SDK setup succeeded in 22.129s, with exact timing and link probe in wasi-setup.log. The 30-minute first-green push target was exceeded by cache recovery, SDK setup and complete consumer validation; no red compiler state was published to save that target.

Final audit: all 55 consumer jobs exited 0, with no leaf reaching 60s. Two new ledger readers and packages without selected checks are explicitly compile-only. consumer-selection.json and consumer-results.json list exact selectors and outcomes; broader Python lane wrappers were not repeated after the recorded source-report limitations. Latest combined positive evidence has 11025 test identities, zero unresolved failures, zero completed events at or over 60s, maximum 33.05s. Vet exit 0 after all additions. gofmt checked all 73 changed Go files with no output. All 84 changed .a files are byte-identical to the supplied source tip, not only body-equivalent.

Completed JSON-lines logs are archived losslessly as .jsonl.gz. evidence-archives.json records original paths, hashes and byte lengths; compact validation-summary.json preserves per-test outcomes. Original failed commands and successful recoveries remain distinct.
