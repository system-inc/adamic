Re-cut slice 4 onto newest fetched main, retaining both property-lowering paths.
Delivery SHA is in the final handoff; base 5565a7cafb0ee8d538f8a5ece2304d5aa4254145.
Linux counts regeneration passed; full lowering, oracle and guards are recorded below.
A targeted presence-check omission mutant proves main's refusal checks still reject the regression.
Not run: full repository/Loom gate, ahra, or a repeat of every historical slice-4 suite and standalone mutant script.

Branch: compiler/loom-r-slice-4-ac-m3. Source: d449a6ae094d288a5283bcf7495eb71435e9030d, applied only as its own parent-to-commit change using the requested merge-tree and cherry-pick --no-commit. Newest fetched origin/main was 5565a7cafb0ee8d538f8a5ece2304d5aa4254145.

One conflict, internal/lower/object.go at the start of property(): main 638f4576 added eepPropertyPresence, protecting union/optional tuple length, getter/method name collisions and static reads entered by earlier initializers. Source adds generatorResultRead for generator result value lowering. Keep main's presence check first, then the generator handler, then all existing property lowering. This preserves main's refusals before any successful early return. No main line is reverted. Conservative assumption: safety checks apply to every property access, including a specialized handler's early-return path.

counts.md was restored from the new base and regenerated once on Linux using the house clang 20.1.8, unanchored TestCountsAreRecorded with -update-counts. It passed in 77.258 package seconds / 83.688 command seconds. Exactly 60 keys changed: 55 additions and five modifications. All 60 keys and resulting values match source d449a6ae's own counts diff; no unrelated keys and no additional value differences.

Runtime C/header changes relative to newest main: none. Extra runtime lines beyond source: none.

Toolchain: GOPROXY=https://proxy.golang.org|direct; cloud/setup.sh --wasi-sdk passed, go build ready 28.991 s, build cache warm 29.209 s, done 29.238 s. Linux Go 1.27.1, Node 24.19.0, clang 20.1.8; nproc=5, cgroup 400000/100000, GOMAXPROCS=4. A direct gofmt invocation before sourcing env.sh was unavailable; it was rerun successfully after sourcing the printed env file. No source compile failed. Old completed-unit export products were reclaimed; current products are preserved between the explicit scratch output and root read-fallback cache.

Tests use direct JSONL log output, hard command limits and four-CPU quota. Counts, full internal/lower and full internal/oracle are followed by each IR call-target export/reader group, final TestCallTargetReaders and TestEveryMapAProgramHoldsIsSorted, cmd/adamic-gate parallel/product guards, compiler-dependency guards, vet on all changed Go packages, and gofmt on all changed Go files. Full lower includes TestNativeAgreementRowsNameRegisteredPrograms and main EEPPresence cases; full oracle includes TestRuntimeSetsAreEachUsed, Node/native/JS comparisons, sanitizers and leak checks. Exact commands, exits and timings are retained in commands.json. Final per-test results and leaf durations are in validation.json.

Mutant: a Go overlay removes only the eepPropertyPresence invocation from the actual resolved object.go, leaving valid lowering code and all generator support intact. TestEEPPresence0 through TestEEPPresence6 must reject it. No compilation-warning kill is credited. Existing generator and in-suite oracle checks are exercised by the complete requested suites; previous broad slice-4 evidence remains under review/compiler/loom-r-slice-4-ac-m2/.

Every changed counts row, before -> after A/F/R/L/P/G:

- `internal/fresh/testdata/regexp_tree.ts`: 91/91/81/85/35/0 -> 91/91/85/89/35/0
- `internal/oracle/testdata/class_wrong_output_refused/iterators_hidden_return.a`: absent -> 28/28/16/20/8/0
- `internal/oracle/testdata/class_wrong_output_refused/iterators_override_this.a`: absent -> 57/57/31/49/9/0
- `internal/oracle/testdata/generators/basic.a`: absent -> 30/30/85/101/14/0
- `internal/oracle/testdata/generators/cancellation.a`: absent -> 23/23/65/74/14/0
- `internal/oracle/testdata/generators/captures.a`: absent -> 31/31/95/95/24/0
- `internal/oracle/testdata/generators/checker.a`: absent -> 20/20/104/93/16/0
- `internal/oracle/testdata/generators/close.a`: absent -> 27/27/57/59/13/0
- `internal/oracle/testdata/generators/completions.a`: absent -> 59/59/166/175/38/0
- `internal/oracle/testdata/generators/default_error.a`: absent -> 13/13/31/30/12/0
- `internal/oracle/testdata/generators/defaults.a`: absent -> 29/29/104/107/21/0
- `internal/oracle/testdata/generators/delegate.a`: absent -> 89/89/316/332/51/0
- `internal/oracle/testdata/generators/delegate_throw.a`: absent -> 28/28/158/151/22/0
- `internal/oracle/testdata/generators/first_reference.a`: absent -> 17/17/73/72/12/0
- `internal/oracle/testdata/generators/map_iterator.a`: absent -> 24/24/95/103/15/0
- `internal/oracle/testdata/generators/owned.a`: absent -> 32/32/132/130/21/0
- `internal/oracle/testdata/generators/reentrant.a`: absent -> 14/14/53/54/11/0
- `internal/oracle/testdata/generators/throw.a`: absent -> 16/16/75/74/12/0
- `internal/oracle/testdata/generic_body_indexed.a`: absent -> 2/2/0/2/2/0
- `internal/oracle/testdata/generic_body_read.a`: absent -> 10/10/0/10/3/0
- `internal/oracle/testdata/miscompile_2a/controls.a`: 33/32/34/56/10/1 -> 39/39/37/58/13/0
- `internal/oracle/testdata/notyet_element_access/nodearray.a`: absent -> 0/0/0/0/0/0
- `internal/oracle/testdata/notyet_element_access/sorted.a`: absent -> 0/0/0/0/0/0
- `internal/oracle/testdata/notyet_element_access/template.a`: absent -> 0/0/0/0/0/0
- `internal/oracle/testdata/regexp.a`: 437/437/382/417/58/0 -> 437/437/383/418/58/0
- `internal/oracle/testdata/user_iterators.a`: 663/663/455/904/64/0 -> 1050/1050/677/994/97/0
- `internal/oracle/testdata/user_iterators_rest_tdz.a`: 5/0/5/3/5/0 -> 11/3/7/3/8/0
- `stage3/fixtures/generics/01_identity.a`: absent -> 3/3/3/7/3/0
- `stage3/fixtures/generics/02_optional_return.a`: absent -> 5/5/3/9/3/0
- `stage3/fixtures/generics/03_callback_return.a`: absent -> 8/8/3/11/3/0
- `stage3/fixtures/generics/05_nested_class.a`: absent -> 8/8/5/16/7/0
- `stage3/fixtures/generics/09_recursive_optional.a`: absent -> 3/3/1/4/3/0
- `stage3/fixtures/generics/10_identifier_multimap.a`: absent -> 11/11/27/36/9/0
- `stage3/fixtures/generics/14_optional_literal_union.a`: absent -> 3/3/0/3/3/0
- `stage3/fixtures/generics/15_array_callback.a`: absent -> 18/18/7/27/4/0
- `stage3/fixtures/iteration-dispatch/body_throw.a`: absent -> 22/22/17/24/12/0
- `stage3/fixtures/iteration-dispatch/break.a`: absent -> 20/20/11/18/10/0
- `stage3/fixtures/iteration-dispatch/cached_next.a`: absent -> 29/29/17/34/12/0
- `stage3/fixtures/iteration-dispatch/class.a`: absent -> 22/22/11/22/7/0
- `stage3/fixtures/iteration-dispatch/close_throw_body.a`: absent -> 22/22/20/26/12/0
- `stage3/fixtures/iteration-dispatch/close_throw_break.a`: absent -> 20/20/16/21/10/0
- `stage3/fixtures/iteration-dispatch/completion.a`: absent -> 20/20/10/20/11/0
- `stage3/fixtures/iteration-dispatch/done_throw.a`: absent -> 13/13/14/16/10/0
- `stage3/fixtures/iteration-dispatch/labeled_continue.a`: absent -> 44/44/22/38/12/0
- `stage3/fixtures/iteration-dispatch/method_getters.a`: absent -> 26/26/16/24/15/0
- `stage3/fixtures/iteration-dispatch/next_throw.a`: absent -> 12/12/13/16/7/0
- `stage3/fixtures/iteration-dispatch/object.a`: absent -> 26/26/14/28/11/0
- `stage3/fixtures/iteration-dispatch/return.a`: absent -> 21/21/11/18/13/0
- `stage3/fixtures/iteration-dispatch/value_throw.a`: absent -> 13/13/14/16/10/0
- `stage3/fixtures/iteration/array_view_stress.a`: absent -> 34/34/50/72/16/0
- `stage3/fixtures/iteration/array_view_weak.a`: absent -> 5/5/8/13/4/0
- `stage3/fixtures/iteration/collections.a`: absent -> 35/35/24/47/11/0
- `stage3/fixtures/iteration/delegated_generator.a`: absent -> 20/20/70/74/15/0
- `stage3/fixtures/iteration/generator.a`: absent -> 14/14/33/35/13/0
- `stage3/fixtures/iteration/object_iteration.a`: absent -> 5/5/1/6/3/0
- `stage3/fixtures/iteration/strings.a`: absent -> 20/20/14/27/9/0
- `stage3/fixtures/iteration/test262_array_views.a`: absent -> 11/11/25/38/11/0
- `stage3/fixtures/iteration/user_forwarding.a`: absent -> 22/22/7/17/14/0
- `stage3/project-references-source/fixture/chain/app/main-print.a`: absent -> 2/2/0/2/2/0
- `stage3/project-references-source/fixture/two/app/main.a`: absent -> 2/2/0/2/2/0

Final results: 7,054 test identities, zero unresolved failures and zero test leaves >=60 s. Maximum completed leaf: 22.06 s. Full lower passed in 50.954 command seconds; full oracle passed in 471.611 command seconds / 470.023 package seconds. All five call-target export/reader groups and the final combined IR guard run pass. Parallel/product guards, compiler dependency guards, vet and gofmt pass; gofmt output is empty. The presence omission mutant fails all seven TestEEPPresence0-6 checks. All 84 changed .a files are byte-identical to d449a6ae and pass pinned Gate.aCheck: 63 checked, five NotYet, 16 refused. No runtime C/header line changed. All command outputs are retained as lossless compressed logs with SHA-256 manifest. No PR, Loom submission, main push or force push.

- `counts`: exit 0, 83.688 s; `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -parallel=4 -timeout 600s -json -args -update-counts`.

- `lower`: exit 0, 50.954 s; `go test ./internal/lower/... -count=1 -parallel=4 -timeout 300s -json`.

- `oracle`: exit 0, 471.611 s; `go test ./internal/oracle/... -count=1 -parallel=4 -timeout 1200s -json`.

- `ir-native`: exit 0, 19.385 s; `go test ./internal/ir -run ^(TestCallTargetExportInputsNative|TestCallTargetReadersNative|TestProduct_CallTargetReaderArchivesNative)$ -count=1 -parallel=4 -timeout 90s -json`.

- `ir-lower`: exit 0, 3.544 s; `go test ./internal/ir -run ^(TestCallTargetExportInputsLower|TestCallTargetReadersLower|TestProduct_CallTargetReaderArchivesLower)$ -count=1 -parallel=4 -timeout 90s -json`.

- `ir-fresh`: exit 0, 6.659 s; `go test ./internal/ir -run ^(TestCallTargetExportInputsFresh|TestCallTargetReadersFresh|TestProduct_CallTargetReaderArchivesFresh)$ -count=1 -parallel=4 -timeout 90s -json`.

- `ir-flow`: exit 0, 15.597 s; `go test ./internal/ir -run ^(TestCallTargetExportInputsFlow|TestCallTargetReadersFlow|TestProduct_CallTargetReaderArchivesFlow)$ -count=1 -parallel=4 -timeout 90s -json`.

- `ir-oracle`: exit 0, 3.934 s; `go test ./internal/ir -run ^(TestCallTargetExportInputsOracle|TestCallTargetReadersOracle|TestProduct_CallTargetReaderArchivesOracle)$ -count=1 -parallel=4 -timeout 90s -json`.

- `ir-guards`: exit 0, 2.692 s; `go test ./internal/ir -run TestCallTargetReaders|TestEveryMapAProgramHoldsIsSorted -count=1 -parallel=4 -timeout 90s -json`.

- `gate-guards`: exit 0, 1.24 s; `go test ./cmd/adamic-gate -run TestEveryTestIsParallelOrSaysWhy|Test.*Product -count=1 -timeout 90s -json`.

- `dependency-guards`: exit 0, 0.667 s; `go test ./internal/compilercensus -run TestDependencies -count=1 -timeout 90s -json`.

- `presence-mutant`: exit 1, 7.449 s; `go test ./internal/lower -overlay /workspace/adamic/review/compiler/loom-r-slice-4-ac-m3/omit-presence-overlay.json -run ^TestEEPPresence[0-6]$ -count=1 -timeout 90s -json`.

- `vet`: exit 0, 1.334 s; `go vet ./internal/flow/... ./internal/fresh/... ./internal/ir/... ./internal/javascript/... ./internal/load/... ./internal/lower/... ./internal/native/... ./internal/oracle/... ./stage3/fixtures/... ./stage3/project-loader/probe/...`.

- `gofmt`: exit 0, 0.033 s; `gofmt -l internal/flow/build.go internal/flow/generators_test.go internal/flow/infer.go internal/fresh/fresh.go internal/fresh/generators_test.go internal/fresh/iteration_test.go internal/ir/call_targets_guard_test.go internal/ir/generator.go internal/ir/ir.go internal/ir/iteration.go internal/javascript/javascript.go internal/load/load.go internal/load/mixed_file_build_test.go internal/load/project_console.go internal/load/project_loader.go internal/load/project_loader_test.go internal/load/project_options.go internal/load/project_references.go internal/load/project_references_test.go internal/load/source_fs.go internal/lower/borrow.go internal/lower/class.go internal/lower/class_wrong_output_test.go internal/lower/cycles.go internal/lower/exceptions.go internal/lower/expression.go internal/lower/functions.go internal/lower/generator_body.go internal/lower/generator_cycles.go internal/lower/generator_delegate.go internal/lower/generator_expressions.go internal/lower/generator_machine.go internal/lower/generator_origin.go internal/lower/generator_protocol.go internal/lower/generator_types.go internal/lower/generators_test.go internal/lower/generic.go internal/lower/generic_body_relations.go internal/lower/generic_body_relations_census_test.go internal/lower/generic_body_relations_test.go internal/lower/invariance.go internal/lower/iteration.go internal/lower/iteration_array_view.go internal/lower/iteration_consume.go internal/lower/iteration_intrinsic.go internal/lower/iteration_test.go internal/lower/labels.go internal/lower/library_function_expressions.go internal/lower/lower_test.go internal/lower/object.go internal/lower/refusals.go internal/lower/statements.go internal/native/emit_expressions.go internal/native/iteration.go internal/native/reuse.go internal/oracle/census_small_test.go internal/oracle/class_wrong_output_test.go internal/oracle/element_access_boundaries_test.go internal/oracle/enums_stage3_test.go internal/oracle/generator_borrow_mutants_test.go internal/oracle/generator_cycle_test.go internal/oracle/generators_test.go internal/oracle/generic_body_relations_test.go internal/oracle/iteration_dispatch_mutant_test.go internal/oracle/iteration_dispatch_test.go internal/oracle/lowering_chain_mixed_test.go internal/oracle/project_references_test.go internal/oracle/step16_generics_test.go internal/oracle/step20_intrinsic_test.go internal/oracle/step20_iteration_test.go internal/oracle/step20_variance_test.go stage3/fixtures/slice4_units_test.go stage3/project-loader/probe/main.go`.
