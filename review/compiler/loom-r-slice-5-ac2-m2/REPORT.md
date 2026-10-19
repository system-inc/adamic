Re-cut slice 5 onto current main while retaining its link closure and stack-test shards.
Delivery SHA is in the final handoff; base a22f4251067a5f53a00d07fc64b54a2ac72d340b.
Full touched-package suites and fixtures completed; initial product-visibility failures are resolved by explicit preparation and focused reruns.
Stale-reader, missing-product, borrowed-record guard and wrong outcome-pin mutants are all caught by their intended checkers.
Not covered: full repository/Loom gate or ahra; built-in optional skips remain explicit.

Source: 2939a32de407d33259e6f175a7cf1b93f7ae1e98. Applied only its own parent-to-commit change after fetching origin, using the recorded merge-tree and cherry-pick --no-commit. Slice 4 was already pushed before the new fetch established this base.

Conflicts and resolutions:

- internal/native/regexp_test.go: one source-map hunk. Preserve main host_runtime.c and add source construction.c and library_object.c, retaining every other runtime source. The reduced regex runtime needs both main host support and the construction/object helper closure.
- internal/oracle/stack_test.go: one large old-loop versus new-shards hunk. Preserve main entire file with all eight top-level stack cases and its Linux argv/V8 budget adjustments. Apply source named panic expectation, changing only `RangeError: Maximum call stack size exceeded\n` to `RangeError: Maximum call stack size exceeded in sum\n`. Source named-stack emitter and oracle normalization require that exact name. No main case deleted or combined.

Runtime clearance: runtime-clearance.json compares each signed C/header patch line with the supplied commit own patch. Extra lines: none. Missing lines: none. The seven previously cleared json_stringify.c lines remain exact. Automatic merges retain newer main infrastructure in adamic.h, heap.c, object.c and stack.c. Runtime source/final patches are retained for independent comparison.

Counts: restored base table and used the Linux clang 20.1.8 updater. An initial anchored obsolete test selector measured no rows and failed without writing; corrected to main documented unanchored TestCountsAreRecorded shard selection. Final changed rows follow below.

Toolchain: cloud/setup.sh --wasi-sdk completed in 29.903 s; go build 29.670 s, build cache warm 29.878 s. nproc=5, cgroup 400000/100000 (four CPUs), GOMAXPROCS=4, Go 1.27.1, Node 24.19.0, house clang 20.1.8. Full setup log retained. GCC uses the real /usr/bin/gcc with LLVM public sanitizer headers supplied explicitly, as required by this box; no compiler output is faked. Root cache is read fallback and new products go to scratch because root disk is constrained. CLI child processes use the explicit scratch cache to see prepared products.

Assumptions: retain source explicit checked-placeholder outcome for the moved review non_null witness, including its intended-difference pin; source program bytes are unchanged. Keep optional test skips and report them. No cohere code copied. New compiler packages: none. All 59 changed .a files are byte-identical to supplied source and pass pinned Gate.aCheck: 43 checked, 16 refused, no header edits.

Counts updater passed in 146.037 package seconds. Every changed row (A/F/R/L/P/G):

- `internal/oracle/testdata/4ddd17f_weak_single_narrowed.a`: absent -> 4/2/4/6/4/0
- `internal/oracle/testdata/backend_stop_catches.a`: absent -> 2/2/6/4/1/0
- `internal/oracle/testdata/d96d304_try_stack.a`: absent -> 0/0/0/0/0/0
- `internal/oracle/testdata/non_null_initialized.ts`: 0/0/0/0/0/0 -> 55/55/50/109/18/0
- `internal/oracle/testdata/non_null_literal_assignment.ts`: 0/0/0/0/0/0 -> 1/1/2/2/1/0
- `internal/oracle/testdata/non_null_static_initialized.ts`: 1/0/0/1/1/0 -> 9/9/12/22/6/0
- `internal/oracle/testdata/non_null_static_uninitialized.ts`: 1/0/0/1/1/0 -> 1/0/3/2/1/0
- `internal/oracle/testdata/non_null_uninitialized_capture.ts`: 0/0/0/0/0/0 -> 2/0/2/0/2/0
- `internal/oracle/testdata/non_null_uninitialized_const.ts`: 0/0/0/0/0/0 -> 0/0/2/1/0/0
- `internal/oracle/testdata/non_null_uninitialized_exception.ts`: 0/0/0/0/0/0 -> 1/1/3/2/1/0
- `internal/oracle/testdata/non_null_uninitialized_field.ts`: 1/0/0/0/1/0 -> 1/0/3/2/1/0
- `internal/oracle/testdata/non_null_uninitialized_interface.ts`: 1/0/1/0/1/0 -> 1/0/2/1/1/0
- `internal/oracle/testdata/non_null_uninitialized_iteration.ts`: 1/0/0/1/1/0 -> 6/1/10/8/6/0
- `internal/oracle/testdata/non_null_uninitialized_loop.ts`: 0/0/0/0/0/0 -> 0/0/1/0/0/0
- `internal/oracle/testdata/non_null_uninitialized_map_entry.ts`: 0/0/0/0/0/0 -> 0/0/1/1/0/0
- `internal/oracle/testdata/non_null_uninitialized_optional.ts`: 1/0/0/0/1/0 -> 3/3/2/6/3/0
- `internal/oracle/testdata/non_null_uninitialized_spread.ts`: 1/0/0/0/1/0 -> 2/0/5/5/2/0
- `internal/oracle/testdata/notyet_element_access/nodearray.a`: absent -> 0/0/0/0/0/0
- `internal/oracle/testdata/placeholder_nonnull_alias_reset.a`: absent -> 2/1/6/7/2/0
- `internal/oracle/testdata/placeholder_nonnull_assignment_result.a`: absent -> 0/0/2/3/0/0
- `internal/oracle/testdata/placeholder_nonnull_before_use.a`: absent -> 0/0/1/1/0/0
- `internal/oracle/testdata/placeholder_nonnull_field.a`: absent -> 9/9/15/23/5/0
- `internal/oracle/testdata/placeholder_nonnull_local.a`: absent -> 12/12/11/24/4/0
- `internal/oracle/testdata/placeholder_nonnull_null_before_use.a`: absent -> 0/0/2/1/0/0
- `internal/oracle/testdata/placeholder_nonnull_null_coalesce.a`: absent -> 8/8/15/27/4/0
- `internal/oracle/testdata/placeholder_nonnull_null_copy.a`: absent -> 26/26/102/121/7/0
- `internal/oracle/testdata/placeholder_nonnull_null_equality.a`: absent -> 4/4/10/16/2/0
- `internal/oracle/testdata/placeholder_nonnull_null_json.a`: absent -> 5/5/0/6/3/0
- `internal/oracle/testdata/placeholder_nonnull_null_loose.a`: absent -> 10/10/23/35/4/0
- `internal/oracle/testdata/placeholder_nonnull_null_optional.a`: absent -> 6/6/16/25/3/0
- `internal/oracle/testdata/placeholder_nonnull_null_saved_leak.a`: absent -> 1/0/4/4/1/0
- `internal/oracle/testdata/placeholder_nonnull_null_truthiness.a`: absent -> 1/1/7/10/1/0
- `internal/oracle/testdata/placeholder_nonnull_null_typeof.a`: absent -> 3/3/9/11/3/0
- `internal/oracle/testdata/placeholder_nonnull_observe.a`: absent -> 27/27/48/79/9/0
- `internal/oracle/testdata/placeholder_nonnull_return_assignment.a`: absent -> 0/0/2/3/0/0
- `internal/oracle/testdata/placeholder_nonnull_saved_leak.a`: absent -> 1/0/3/4/1/0
- `internal/oracle/testdata/placeholder_nonnull_scanner.a`: absent -> 4/4/9/8/3/0
- `stage3/interface-downcasts/default-boxed-write.ts`: 0/0/0/0/0/0 -> 5/5/5/12/4/0
- `stage3/interface-downcasts/default-read-before-set.ts`: 0/0/0/0/0/0 -> 1/0/4/2/1/0
- `stage3/interface-downcasts/default-staged.ts`: 0/0/0/0/0/0 -> 2/2/6/7/2/0
- `stage3/interface-downcasts/readiness-identifier-uninitialized.ts`: 0/0/0/0/0/0 -> 1/0/4/3/1/0
- `stage3/interface-downcasts/readiness-identifier.ts`: 0/0/0/0/0/0 -> 2/2/6/8/2/0
- `stage3/interface-downcasts/readiness-number-uninitialized.ts`: 0/0/0/0/0/0 -> 1/0/5/3/1/0
- `stage3/interface-downcasts/readiness-number.ts`: 0/0/0/0/0/0 -> 5/5/5/10/5/0
- `stage3/parser-ahead/factories/already-complete.a`: absent -> 1/1/4/4/1/0
- `stage3/parser-ahead/factories/complete.a`: absent -> 1/1/5/4/1/0
- `stage3/parser-ahead/factories/escaped.a`: absent -> 1/1/3/4/1/0
- `stage3/parser-ahead/node-array/fields.a`: absent -> 12/12/10/23/12/0
- `stage3/parser-ahead/rulings/factory-unset.a`: absent -> 3/3/13/20/3/0
- `stage3/parser-ahead/rulings/factory-use.a`: absent -> 1/0/2/3/1/0
- `stage3/parser-ahead/rulings/node-array-json.a`: absent -> 4/4/5/11/2/0
- `stage3/parser-ahead/rulings/node-array-keys.a`: absent -> 15/15/12/16/9/0
- `stage3/parser-ahead/rulings/node-array-length.a`: absent -> 4/4/4/9/2/0
- `stage3/parser-ahead/rulings/node-array-unset.a`: absent -> 4/4/18/29/4/0
- `stage3/parser-ahead/rulings/null-placeholder-pending.a`: absent -> 2/2/4/8/2/0
- `stage3/parser-ahead/rulings/own-key-order.a`: absent -> 28/28/17/37/10/0
- `stage3/parser-next/speculation/arrow-tryparse.a`: absent -> 3/3/8/11/2/0
- `stage3/parser-next/speculation/misfit.a`: absent -> 1/0/2/1/1/0
- `stage3/parser-next/speculation/named-lookahead.a`: absent -> 5/5/1/7/4/0
- `stage3/parser-next/speculation/named-tryparse.a`: absent -> 3/3/10/16/3/0
- `stage3/parser-next/speculation/truthiness-rewind.a`: absent -> 34/34/46/85/6/0
- `stage3/parser-next/stack/benchmark.a`: absent -> 7/7/0/17200008/4/0
- `stage3/parser-next/stack/legal.a`: absent -> 7/7/0/1728/4/0

Tests and commands:

All test output is written directly to logs, never piped. Package runs use `go test ./<package>/... -count=1 -parallel=4 -timeout 1200s -json` under a hard outer limit and continuous monitoring, GOMAXPROCS=4. Full suites invoked: cmd/adamic, internal/flow, internal/fresh, internal/ir, internal/javascript, internal/lower, internal/native, internal/oracle, stage3/fixtures.

- flow: exit 0, 144.960 command seconds; fresh: exit 0, 17.290; ir: exit 0, 32.640; javascript: exit 0, 0.580; lower: exit 0, 44.050.
- oracle: complete suite exit 0, 457.930 command seconds / 456.368 package seconds, including Node, ASan/UBSan/leak, exact runtime sets, counts and source in-suite mutants.
- stage3/fixtures: complete suite exit 0, 54.470 command seconds.
- cmd/adamic: initial complete run failed four behavior tests because CLI children could not see runtime products in the separate output cache. An earlier anchored preparation selector named no current shard. Prepare every TestProduct_CommandRuntime shard with the explicit scratch cache shared by children; complete rerun exit 0, 31.090 package seconds. No source/test contract weakened.
- native: complete invocation 283.420 command seconds / 281.753 package seconds, with exactly two failed tests: TestSplitTSGoAgreesUnit00 and Unit01 ran before the separately owned split baseline was prepared. TestProduct_SplitTSGoBaseline completed in 19.140 s in that full invocation. Explicit product preparation and both behavior reruns pass. All other full-suite tests pass or take their built-in skips. Initial log retained; no over-budget leaf was rerun.
- Count updater: obsolete anchored selector failed without measuring or writing rows; official unanchored `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -parallel=4 -timeout 600s -json -args -update-counts` completed successfully in 146.037 package seconds, once. The resulting 63 keys and all their values exactly match source commit's own changed rows, with no unrelated row.
- Pinned Gate.aCheck over all 59 changed .a programs: exit 0, 43 checked / 16 refused. No program body or header changed versus supplied source.
- Final IR TestCallTargetReaders prefix: exit 0 with -timeout 90s. Full ir also passes TestEveryMapAProgramHoldsIsSorted; full lower passes TestNativeAgreementRowsNameRegisteredPrograms; full oracle passes TestRuntimeSetsAreEachUsed (1.950 s).
- cmd/adamic-gate parallel/product guards: exit 0. No new compiler-consuming Go package or dependency declaration is needed.
- go vet on every changed Go package recursively: exit 0. gofmt -l every changed Go file: exit 0 with empty output.

Stack resolution proof: main all eight stack cases pass individually in 0.35-0.46 s; Node comparisons retained where main runs them, native plain and sanitized comparisons retained everywhere. stack-seconds.json records all times. The entire json_stringify.c is byte-identical to the supplied commit, recorded by SHA-256 in json-runtime-identity.json.

Recorded mutants and catchers:

- Stale reader: change one current literal path back to its missing old non_null path via Go overlay. TestFreshCorpusUnitsCoverEveryProgram rejects corpus coverage; exit 1.
- Missing checker baseline: change only product name via Go overlay. TestSplitTSGoAgreesUnit00 rejects the missing preparation immediately with must be built ahead, exit 1.
- Borrowed-record guard: remove only shape != NULL protection in the production C compiled by current record harness via Go test overlay. TestJSONEncoderRecordView_000 catches UBSan member access through null adamic_shape; exit 1. Valid compilation, no warning/compile-kill credit.
- Wrong intended-outcome pin: change exit 70 to 71 temporarily, run the non_null review witness, restore original bytes in finally. Native, JavaScript and release each independently reject exit disagreement; exit 1. Actual source Node behavior remains explicitly distinct from the ruled checked-use pin.
- Source in-suite construction mutants (synthesized key, declaration order, array extras JSON) and parser stack guard mutant all pass their catchers in the complete native/oracle suites. Historical standalone Python mutant scripts are not all repeated.

Evidence: commands.json and additional-commands.json retain exact commands/exits; validation.json records every current test identity, per-log failures and individual seconds, including the original failures and resolved retries. No unresolved failure or leaf >=60 s is accepted before push. Consumer package selections and per-package results are retained separately, including explicit compile-only selections for packages without a previous focused reader selection.

Cache housekeeping: cache-retired.json lists four obsolete previous-main export products retired through a named repository scratch directory; cache-relocated.json lists three current export products preserved in the root read-fallback cache. No current product or source was discarded. No PR, ahra, full repository gate, Loom submission, main push, area-branch push, force push, or branch deletion. Optional built-in skips and broad Python-wrapper/benchmark lanes remain outside this focused unit.

Final validation: 55/55 reader packages exit 0; 10,859 completed test identities, zero unresolved failures, zero leaves >=60 s. Maximum observed completed leaf is 32.91 s. Selections include the complete internal/load reader suite and explicitly named compile-only packages. Every individual test time is retained in validation.json and consumer-results.json. Branch: compiler/loom-r-slice-5-ac2-m2. Source own runtime signed lines match exactly; no additional runtime clearance lines.
