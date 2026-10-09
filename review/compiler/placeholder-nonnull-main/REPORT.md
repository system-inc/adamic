Rebuilt step 09 literal placeholder initialization on main, including typed-use checks and honest null/undefined observations.
Five requested commits replayed as db195ccd2, 0131f51f3, 6416a497e, 6458bff12 and 28af90505; main adaptation is 895034708.
Build, focused lowerer/oracle/CLI tests, a-check and vet pass; counts regeneration passed with 40 changed rows.
Seven flow IR mutants, null-tag and spread mutants, readiness mutants and eight independent compiler mutants are caught.
Native tsc over the measured corpus and the full gate were not run; Weak/accessor/escaping reset boundaries remain unsupported.

Base: origin/main e8b02f8ca. No main or area branch was merged into or pushed. No c2 dependency was carried: main plus the five requested commits builds without c2 additions.

The source report and compressed source logs are historical evidence. Current logs and results are in this review directory. No mutant source here is compilable Go; sources use .go.txt.

Cherry-pick resolutions:

| Source | Rebuilt | Conflicts and resolution |
| --- | --- | --- |
| 71131229 | db195ccd2 | None. Move its read-first results to review/. |
| 19193932 | 0131f51f3 | javascript.go and emit_statements.go: retain namespace readiness alongside placeholder resets. refusals.go: retain enumeration index admission and main checkedAssertionSource alongside literal placeholder admission. counts.md: retain main rows plus new placeholder rows. |
| b2528aea | 6416a497e | ir.go: retain checked-view metadata plus placeholder provenance/check metadata. javascript.go: retain physical representation tracking plus placeholder readiness. view_fields.go: keep checked-view adapter plus placeholder read helper. view_fields_test.go: keep parallel tests, update placeholder expectations and remove obsolete eager-assertion expectations. counts.md: combine main and placeholder rows; regenerate. |
| c8f858b9 | 6458bff12 | None. Standalone refusal-pass result bookkeeping preserved. |
| 74f36d3d | 28af90505 | None. Catalog patches preserved. |

Main adaptations: native placeholder view reads recognize explicit null/undefined representation tags on reset slots. JavaScript placeholder allocation/definition keeps the union storage tag, matching native fieldInitialRepresentation, so a completing scalar write is accepted. Other physical representation tags and checked-view dispatch remain in place. New placeholder case tables are split into top-level tests; subprocess tests are parallel.

Setup first ran while cherry-pick conflicts were unresolved and failed with Go syntax errors at conflict markers in javascript.go, emit_statements.go and refusals.go. Retrying after resolution succeeded: go ready 0.059s, node ready 0.063s, markdown ready 0.163s, submodules ready 0.227s, clang ready 0.323s, go build ready 43.157s, test binaries deferred 43.666s, cache warm 43.696s, done 43.835s. nproc 5; cpu.max 400000 100000. GOPROXY=https://proxy.golang.org|direct; environment /workspace/adamic-tools/env.sh.

Commands:

- go build ./...: exit 0, including setup build.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle ./cmd/adamic -run 'Placeholder|Readiness|NonNull|Uninitialized|OptionalWideningSpreadOverwrite|TestNativeAgreesWithNode/internal/oracle/testdata/placeholder_nonnull' -count=1 -v: first run exposed physical nullish tag mismatch. Corrected focused rerun passes (focus.log).
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^(TestRequiredViewFieldPrimitive|TestRequiredViewFieldOperandOnce|TestNarrowedFieldUsesSharedReadiness|TestViewFieldInheritedStaticReadiness|TestDefaultTaggedSourceViews|TestPlaceholder.*)$|TestNativeAgreesWithNode/internal/oracle/testdata/placeholder_nonnull' -count=1 -v: final both-backend oracles pass, 22.500s; sanitizers and completed-program leaks. Checked stops pin exit 70 separately from allowed Node nullish observations.
- go test ./internal/lower ./internal/oracle ./cmd/adamic -run Placeholder -count=1 -v: split tests pass, 0.717s lowerer, 2.458s oracle, 0.627s CLI before CLI case split.
- python3 review/compiler/placeholder-nonnull-main/a-check.py: integration Gate.aCheck implementation, 20 changed .a files, exit 0, 15.2s.
- python3 review/compiler/placeholder-nonnull-main/mutants.py: all eight independent mutants caught, exit 0. No compiler or clang failure is credited except the standalone refusal-pass bookkeeping regression's intended Go panic.
- go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts: exit 0, 99.182s aggregate across the recorded fixture census.
- go vet ./cmd/adamic ./internal/flow ./internal/fresh ./internal/ir ./internal/javascript ./internal/lower ./internal/native ./internal/oracle: exit 0.
- gofmt on changed Go files and git diff --check: clean.

Independent compiler mutants:

| Mutation | File | Catcher |
| --- | --- | --- |
| Mark every typed use proven | internal/lower/placeholder.go | TestPlaceholderUseCheckSavedLeak: unset reaches take instead of exit 70. |
| Keep alias field readiness after reset | internal/lower/readiness.go | TestPlaceholderUseCheckAliasReset: unset reaches typed assignment. |
| Test effectful nullish receiver twice | internal/lower/expression.go | null_loose Node oracle: call count changes. |
| Treat JSON temporary placeholder as ordinary assertion | internal/lower/library_json_stringify.go | null_json Node oracle: stops instead of serializing null/omitting undefined. |
| Admit Weak placeholder | internal/lower/expression.go | TestPlaceholderWeakSlotStaysNotYet: expected NotYet disappears. The first attempt changed only object.go and was masked by the independent declaration guard in expression.go; that attempt is not counted as caught. |
| Omit standalone refusal result state | internal/lower/placeholder.go | TestOptionalWideningSpreadOverwrite: intended nil-result Go panic. |

The seven independent IR flow-check removals change each pinned exit 70 to source Node's successful output, in both backends and under successful-program leak checks. Collapsing null to undefined changes the equality fixture's Node output. Clearing honest nullish storage markers makes the spread fixture stop where Node finishes. Readiness zero/proof/capture/exception mutants and ordinary assertion/Weak diagnostics remain covered by focused tests.

Counts: 19 new rows and 21 changed existing rows. A deliberate stop's counts describe its stopping point, not a completed leak check. Each row is listed with its reason below (columns allocations, frees, retains, releases, peak, region frees).

| Fixture | Before | After | Reason |
| --- | --- | --- | --- |
| internal/oracle/testdata/placeholder_nonnull_scanner.a | new | 4, 4, 9, 8, 3, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_local.a | new | 12, 12, 11, 24, 4, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_field.a | new | 9, 9, 15, 23, 5, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_observe.a | new | 27, 27, 48, 79, 9, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_null_equality.a | new | 4, 4, 10, 16, 2, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_null_loose.a | new | 10, 10, 23, 35, 4, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_null_truthiness.a | new | 1, 1, 7, 10, 1, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_null_typeof.a | new | 3, 3, 9, 11, 3, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_null_coalesce.a | new | 8, 8, 15, 27, 4, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_null_optional.a | new | 6, 6, 16, 25, 3, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_null_copy.a | new | 26, 26, 102, 121, 7, 0 | New fixture; copy, save/restore and spread preserve nullish payloads and release their storage. |
| internal/oracle/testdata/placeholder_nonnull_null_json.a | new | 5, 5, 0, 6, 3, 0 | New fixture; exact null/undefined JSON temporary fields use existing schema arms and release their objects and output strings. |
| internal/oracle/testdata/placeholder_nonnull_before_use.a | new | 0, 0, 1, 1, 0, 0 | New negative fixture; counts stop at the pinned typed-use check. |
| internal/oracle/testdata/placeholder_nonnull_saved_leak.a | new | 1, 0, 3, 4, 1, 0 | New negative fixture; counts stop at the pinned typed-use check. |
| internal/oracle/testdata/placeholder_nonnull_null_before_use.a | new | 0, 0, 2, 1, 0, 0 | New negative fixture; counts stop at the pinned typed-use check. |
| internal/oracle/testdata/placeholder_nonnull_null_saved_leak.a | new | 1, 0, 4, 4, 1, 0 | New negative fixture; counts stop at the pinned typed-use check. |
| internal/oracle/testdata/placeholder_nonnull_assignment_result.a | new | 0, 0, 2, 3, 0, 0 | New negative fixture; counts stop at the pinned typed-use check. |
| internal/oracle/testdata/placeholder_nonnull_return_assignment.a | new | 0, 0, 2, 3, 0, 0 | New negative fixture; counts stop at the pinned typed-use check. |
| internal/oracle/testdata/placeholder_nonnull_alias_reset.a | new | 2, 1, 6, 7, 2, 0 | New negative fixture; counts stop at the pinned typed-use check. |
| internal/oracle/testdata/non_null_initialized.ts | 0, 0, 0, 0, 0, 0 | 55, 55, 50, 109, 18, 0 | Completing writes now execute; scalar placeholder slots box their values and captured saves own their cells. |
| internal/oracle/testdata/non_null_uninitialized_field.ts | 1, 0, 0, 0, 1, 0 | 1, 0, 3, 2, 1, 0 | Allocate the object, transfer its nullish field and stop at the typed property use. |
| internal/oracle/testdata/non_null_uninitialized_capture.ts | 0, 0, 0, 0, 0, 0 | 2, 0, 2, 0, 2, 0 | Create the captured cell and closure, then stop when the callback uses the unset value as T. |
| internal/oracle/testdata/non_null_uninitialized_exception.ts | 0, 0, 0, 0, 0, 0 | 1, 1, 3, 2, 1, 0 | Execute catch handling and its string output before the later typed-use stop. |
| internal/oracle/testdata/non_null_uninitialized_loop.ts | 0, 0, 0, 0, 0, 0 | 0, 0, 1, 0, 0, 0 | Carry unset into the loop exit and stop at the typed use, after its reference transfer. |
| internal/oracle/testdata/non_null_uninitialized_const.ts | 0, 0, 0, 0, 0, 0 | 0, 0, 2, 1, 0, 0 | Save the nullish reference then stop at the property receiver. |
| internal/oracle/testdata/non_null_static_initialized.ts | 1, 0, 0, 1, 1, 0 | 9, 9, 12, 22, 6, 0 | Complete the static writes; scalar boxes and strings are released on the successful path. |
| internal/oracle/testdata/non_null_static_uninitialized.ts | 1, 0, 0, 1, 1, 0 | 1, 0, 3, 2, 1, 0 | Allocate static storage and transfer its nullish value before the property-use check. |
| internal/oracle/testdata/non_null_uninitialized_optional.ts | 1, 0, 0, 0, 1, 0 | 3, 3, 2, 6, 3, 0 | Observe unset through optional handling and complete rather than stopping at initialization. |
| internal/oracle/testdata/non_null_uninitialized_spread.ts | 1, 0, 0, 0, 1, 0 | 2, 0, 5, 5, 2, 0 | Copy the actual nullish field, allocate the copy and stop only at its later typed use. |
| internal/oracle/testdata/non_null_literal_assignment.ts | 0, 0, 0, 0, 0, 0 | 1, 1, 2, 2, 1, 0 | Execute the reset, later completing write and observation; own and release the resulting value. |
| internal/oracle/testdata/non_null_uninitialized_iteration.ts | 1, 0, 0, 1, 1, 0 | 6, 1, 10, 8, 6, 0 | Allocate captured iteration cells and closures before the callback typed-use stop. |
| internal/oracle/testdata/non_null_uninitialized_interface.ts | 1, 0, 1, 0, 1, 0 | 1, 0, 2, 1, 1, 0 | Read the nullish field through the checked structural view before stopping at use as T. |
| internal/oracle/testdata/non_null_uninitialized_map_entry.ts | 0, 0, 0, 0, 0, 0 | 0, 0, 1, 1, 0, 0 | Transfer the nullish reference before the map argument boundary stops. |
| stage3/interface-downcasts/default-staged.ts | 0, 0, 0, 0, 0, 0 | 2, 2, 6, 7, 2, 0 | Complete staged placeholder field writes and the structural-view reads. |
| stage3/interface-downcasts/default-boxed-write.ts | 0, 0, 0, 0, 0, 0 | 5, 5, 5, 12, 4, 0 | Complete scalar field writes using union boxes and release their values. |
| stage3/interface-downcasts/default-read-before-set.ts | 0, 0, 0, 0, 0, 0 | 1, 0, 4, 2, 1, 0 | Allocate the placeholder object and check its typed structural field read. |
| stage3/interface-downcasts/readiness-identifier.ts | 0, 0, 0, 0, 0, 0 | 2, 2, 6, 8, 2, 0 | Complete reference field assignment and checked structural-view transfer. |
| stage3/interface-downcasts/readiness-identifier-uninitialized.ts | 0, 0, 0, 0, 0, 0 | 1, 0, 4, 3, 1, 0 | Transfer the unset reference through a structural view and stop at use as T. |
| stage3/interface-downcasts/readiness-number.ts | 0, 0, 0, 0, 0, 0 | 5, 5, 5, 10, 5, 0 | Box completed numeric placeholder fields and release the boxes after structural reads. |
| stage3/interface-downcasts/readiness-number-uninitialized.ts | 0, 0, 0, 0, 0, 0 | 1, 0, 5, 3, 1, 0 | Preserve the numeric field nullish payload and stop at its typed structural read. |

New top-level test seconds (reported Go execution, every leaf under 60s; setup is the shared successful toolchain setup above):

| Test | Seconds |
| --- | --- |
| TestPlaceholderExplainChecksOutput | 0.00 |
| TestPlaceholderExplainChecksOutputBuild | 0.23 |
| TestPlaceholderExplainChecksOutputC | 0.08 |
| TestPlaceholderExplainChecksOutputJavaScript | 0.08 |
| TestPlaceholderFlowMutantAliasReset | 0.90 |
| TestPlaceholderFlowMutantAssignmentResult | 1.90 |
| TestPlaceholderFlowMutantBeforeUse | 1.35 |
| TestPlaceholderFlowMutantNullBeforeUse | 2.20 |
| TestPlaceholderFlowMutantNullSavedLeak | 2.06 |
| TestPlaceholderFlowMutantReturnAssignment | 1.91 |
| TestPlaceholderFlowMutantSavedLeak | 1.10 |
| TestPlaceholderNonliteralAssertionStillChecks | 0.83 |
| TestPlaceholderNullTagMutant | 1.39 |
| TestPlaceholderSpreadReadinessMutant | 2.69 |
| TestPlaceholderTypedBoundaryAliasReset | 0.12 |
| TestPlaceholderTypedBoundaryArgument | 0.10 |
| TestPlaceholderTypedBoundaryArrayElement | 0.15 |
| TestPlaceholderTypedBoundaryAssignmentResult | 0.24 |
| TestPlaceholderTypedBoundaryCall | 0.13 |
| TestPlaceholderTypedBoundaryDominatingWrite | 0.13 |
| TestPlaceholderTypedBoundaryElementWrite | 0.17 |
| TestPlaceholderTypedBoundaryOptionalArrayElement | 0.12 |
| TestPlaceholderTypedBoundaryOptionalParameterLaterTUse | 0.14 |
| TestPlaceholderTypedBoundaryOptionalReceiver | 0.11 |
| TestPlaceholderTypedBoundaryOrdinaryField | 0.21 |
| TestPlaceholderTypedBoundaryPlaceholderFieldCopy | 0.33 |
| TestPlaceholderTypedBoundaryPresence | 0.20 |
| TestPlaceholderTypedBoundaryProperty | 0.14 |
| TestPlaceholderTypedBoundaryReturn | 0.17 |
| TestPlaceholderTypedBoundaryReturnAssignment | 0.05 |
| TestPlaceholderTypedBoundarySavedAfterWrite | 0.22 |
| TestPlaceholderTypedBoundarySavedCopy | 0.15 |
| TestPlaceholderTypedBoundaryTypedLocal | 0.12 |
| TestPlaceholderUseCheckAliasReset | 0.80 |
| TestPlaceholderUseCheckAssignmentResult | 1.66 |
| TestPlaceholderUseCheckBeforeUse | 0.95 |
| TestPlaceholderUseCheckNullBeforeUse | 1.15 |
| TestPlaceholderUseCheckNullSavedLeak | 1.09 |
| TestPlaceholderUseCheckReturnAssignment | 1.86 |
| TestPlaceholderUseCheckSavedLeak | 1.22 |
| TestPlaceholderWeakSlotStaysNotYet | 0.23 |

The first parallel catalog checks and final CLI build exhausted temporary disk during concurrent Go builds, reporting no space left on device. Those failed controls are not mutant catches. They are retried sequentially after temporary worktrees are cleaned up. Final CLI case functions pass: C 0.08s, JavaScript 0.08s, Build 0.23s.

Catalog entry 09 retry at 895034708: check.sh HEAD --entry 09 -jobs 1 exits 0, applies-and-fails-as-recorded; clean control passes and numeric zero changes stdout versus Node. Entry 12 at f39014171: check.sh HEAD --entry 12 -jobs 1 exits 0, applies-and-fails-as-recorded; clean control passes and the removed suppression-directive refusal fails the required cases. Both patch application checks pass unchanged. Entry 09 is also rerun on f39014171.

Integration lane checks after commit f39014171: lane checks 3.2 s: gofmt and tools on 39 Go files, t.Parallel on 3 test packages; vet 3 packages. The first lane command omitted the toolchain environment and could not find gofmt; sourcing /workspace/adamic-tools/env.sh and rerunning the prescribed command passes.

The two main compatibility mutants are independently caught: dropping native nullish metadata handling fails the null_copy source Node oracle; dropping JavaScript placeholder union initialization tags fails TestDefaultTaggedSourceViews/default-boxed-write (exit 70 instead of Node output 42 and true). Both are behavior failures, not compiler or sanitizer failures.

Final entry 09 on f39014171 passes, applies-and-fails-as-recorded; clean control 0 and expected mutant failure. Entry 12 on the same code tip passes likewise. Their disposable worktree results and logs are retained in catalog-09-final/ and catalog-12-final/. Changes after this tip are evidence and mutant runners only.
