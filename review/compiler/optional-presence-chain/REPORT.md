Built step 1 of optional-presence-chain toward roadmap step 17 on after-chain 2391c655, retaining the chain presence layout.
Commits: b215cedb, 4d505a5c, 5f277ebb, 84f194fc; final corrections and evidence follow these on the delivery branch.
Commands: compiler build and vet pass; seven lane fixtures, three stage3 entries, counts updates, frontend checks, and focused controls pass, with logs here.
Mutants: 15 lane/catalog/admission/registration mutants caught, plus readiness and initializer regression mutants; each catcher is listed below.
Not covered: step 2 checked-view optional writes, full fuzz sweep, full packages/gate, WASI, Darwin, and performance; runtime clearance remains pending.

This is the first rebuild unit permitted by the brief. The four source commits were applied in order on origin/compiler/after-chain. The second commit is empty because the chain already carries a newer catalog patch for placeholder storage; its existing patch context and provenance were kept. All other commits retain their own changes. No other worker branch was merged, and the four excluded compiler files were not edited.

The chain already supplies SIZE_MAX absence, write_order, presence queries, publication, enumeration sorting and filtering, static write publication, tuple and namespace behavior, and shared region sizing. Those implementations remain. Optional reservation, deletion, and dynamic copies now use that state. No zero-rank second presence representation was added. Readiness and representation byte tails remain in the chain order. Reserved rank storage is aligned in the shared size helper. Construction and NodeArray interior metadata initialize the added descriptor fields. Dynamic shape caches bypass pointer identity because a freed descriptor address can recur.

Copies carry order, absence, readiness, and physical tags through a shared field-copy path. Accessor results publish ready physical representation state. Numeric unknown reads of uninitialized declared fields are admitted using the runtime's readiness-before-storage check; getter, method, nullable and other uninitialized representation guards remain. The unknown fixture's assertions were preserved, and native and backend JavaScript agree with source Node.

The reduced seed-1 fixture prints 0. Alias deletion produces empty Object.keys before reinsertion. Reverse alias deletion, Object.entries, Object.values, spread, hasOwn, and reinsertion order agree with Node in release native, sanitized native, and backend JavaScript. The checked-copy fixture retains its intentional checked stop and is not claimed to match raw Node after that stop. The checked-view pending test still observes the precise current NotYet boundary. Step 2 requires work on V1's checked-view optional representation path and is not claimed complete here.

Eight counts rows were added, and zero existing rows moved. See counts-diff.md for every new value. Only own fixtures and conflict-related rows were selected for regeneration. The existing harness also invokes its registered additional counters; those measured rows were merged into the preserved union table, and unselected rows were retained. The two source-branch fixture counts that differ each add one retain and release; retained chain ownership is an inference, not an isolated measurement.

Runtime clearance material for @system_adamic_runtime is runtime-diff.md, which lists each net hunk, and runtime.patch, which records its exact code. class_static.c required no net change because the chain already publishes writes. This report does not grant runtime clearance.

Setup used GOPROXY='https://proxy.golang.org|direct'. First setup failed because its dependency list was produced before branch checkout while load.go compiled after checkout: projectOptionsForRoots, projectSourceRoots, sourceProgramOptions, uniqueSourceRoots, hasHostConsole, OptionSite, ProjectOptionReport, alreadyStricter, auditProjectOptions and nearestProject were undefined. The stable-checkout rerun passed. Its timing lines were Node 0.027s, Go 0.030s, submodules 0.077s, markdown 0.081s, clang 0.196s, build 57.045s, deferred tests 57.216s, cache 57.218s, done 57.252s. nproc is 5, with cgroup quota 4 CPUs. The environment is /workspace/adamic-tools/env.sh. Node types were missing when the counts harness ran its extra adapters; timeout 120 npm ci --ignore-scripts --prefix stage3/api installed the pinned dependencies and the counts rerun passed.

All test output went directly to log files. Every test invocation uses -count=1 and -timeout 90s, with a timeout 180 outer process limit. No complete package test or full gate was run. Final semantic runs use ADAMIC_GATE_UNCACHED=1.

| Command or selection | Result and evidence |
| --- | --- |
| timeout 180 go build ./cmd/adamic | exit 0, build-final.log |
| timeout 180 go vet ./internal/lower ./internal/native ./internal/oracle | exit 0, vet.log |
| go test ./internal/oracle -run 'TestOptionalField\|TestLiteralOptionalOracleCatchesMutant\|TestNativeAgreesWithNode/internal/oracle/testdata/optional_field' -count=1 -v -timeout 90s | final oracle-final.log, seven semantic fixtures and built-in mutants |
| go test ./internal/oracle -run '^TestOptionalFieldCopyState$' -count=1 -v -timeout 90s | 3.131s, copy-state.log |
| go test ./internal/lower ./internal/native ./internal/oracle -run 'TestUnknownReflectionRefusals\|TestUnknownAbsentSpreadPresence\|TestObject\|TestOptionalDeletion\|TestOptionalAssign\|TestUniformFieldsMatchNode\|TestRuntimeFieldLayoutsAreIncluded\|TestRegexProgramsKeepCheckedFieldReads\|TestOptionalWriteReservedSlotMatchesNode\|TestReadinessMutants\|TestUninitializedIsNotNullishMutant\|TestLazyInitializerIsNotEagerMutant' -count=1 -v -timeout 90s | lower 1.935s, native 1.815s, oracle 2.458s; controls.log; the old lazy-initializer name matches no current test |
| go test ./internal/oracle -run 'TestNonliteralInitializerCannotSkipCheck\|TestImportCycle' -count=1 -v -timeout 90s | 2.663s, initializer-cycles.log; current renamed initializer control and both import-cycle controls |
| go test ./stage3/fixtures -run '(TestFixturesObjects\|TestFixturesTaste)/(objects\|taste)/(08_loop_state\|12_decorator_descriptor\|17_binder_flow)' -count=1 -v -timeout 90s | 3.644s, stage3.log |
| go test ./internal/oracle -run 'TestNativeAgreesWithNode/stage3/fixtures/taste/17_binder_flow' -count=1 -v -timeout 90s | 0.547s, binder-control.log |
| go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/e4eec87_u01_undefined_field_widened' -count=1 -v -timeout 90s | 0.429s, catalog-control.log |
| go test ./internal/oracle -run 'TestCountsAreRecorded/fixtures/internal/oracle/testdata/(optional_field\|optional_after_call\|optional_indexing\|library_object\|literal_optional\|e4eec87_u0[12]\|write_order\|construction\|taste_optional_join)' -count=1 -timeout 90s -args -update-counts | 29.882s, counts.log |
| go test ./internal/oracle -run 'TestCountsAreRecorded/fixtures/stage3/fixtures/taste/(17_binder_flow\|24_proportional_conditions)' -count=1 -timeout 90s -args -update-counts | 16.901s, counts-taste.log |
| timeout 480 python3 review/compiler/optional-presence-chain/run-mutants.py | exit 0, mutants.log and per-mutant logs; production files unchanged by overlays |
| changed .a frontend checks using /tmp/optional-presence-chain-adamic c | fast-gate aCheck success/NotYet rules, acheck.log |

The source commits' mutants were rerun as follows. Native generated-code and state mutants run both release and sanitizer modes where their original tests do. None of the successful catches below is a compiler warning or build failure.

| Mutant | Catcher |
| --- | --- |
| Drop omitted optional slot | TestOptionalFieldWriteCatchesDroppedSlot: missing-field stop, exit 70, differing from Node |
| Initially present omitted slots | TestOptionalFieldPresenceCatchesMutants: stdout differs from Node |
| Layout rank instead of insertion rank | TestOptionalFieldPresenceCatchesMutants: stdout differs from Node |
| Suppress deletion | TestOptionalFieldPresenceCatchesMutants: stdout differs from Node |
| Drop spread reservation | TestOptionalFieldConstructionCatchesDroppedReservation: missing-field stop, exit 70 |
| Restore static key list | TestOptionalFieldAliasCatchesStaticEnumeration: clean exit but [first] instead of [] after deletion, release and sanitized |
| Drop copied presence | TestOptionalFieldCopyState: state assertions, release and sanitized |
| Drop copied readiness | TestOptionalFieldCopyState: state assertions, release and sanitized |
| Drop copied representation | TestOptionalFieldCopyState: physical tags become zero, release and sanitized |
| Overlap rank and readiness storage | TestOptionalFieldCopyState: state assertions, release and sanitized |
| Required lookup in place of optional lookup | TestLiteralOptionalOracleCatchesMutant: missing-field stop after a green unreserved-layout control |
| Remove deletion origin proof | TestOptionalDeletionRequiresPlainStorage: wrongly admitted class alias, plain-origin.log |
| Replace data-origin proof with names-only proof | TestObjectUnprovenShapesStayNotYet: wrongly admitted getter alias, hasown-origin.log |
| Restore stale binder NotYet registration | TestNativeAgreesWithNode/stage3/fixtures/taste/17_binder_flow: want stage 0 to refuse, binder-registration.log |
| Literal undefined catalog patch | TestNativeAgreesWithNode/internal/oracle/testdata/e4eec87_u01_undefined_field_widened: stdout differs, catalog.log |
| miss-exception-path, replace-unset-check-with-zero, replace-missing-assertion-with-empty, weak-generic-message, initialize-to-zero, miss-captured-read, erase-without-proof | TestReadinessMutants: pinned diagnostics or stdout, controls.log |
| Uninitialized represented as nullish | TestUninitializedIsNotNullishMutant: pinned observation, controls.log |
| Drop nonliteral initialization check | TestNonliteralInitializerCannotSkipCheck: pinned output, initializer-cycles.log |

New or changed lane top-level test durations from the green run: dropped-slot 0.48s, presence 1.23s, spread-reservation 0.66s, static-enumeration 0.60s, copy-state 3.12s, literal-optional mutant below 1s, reserved-slot native control 1.36s, hasOwn interface 0.23s, hidden-key admission 0.48s, deletion admission 0.40s, absent-spread admission 0.19s. All their test leaves are below 60s. The pending view boundary was checked then skipped in 0.20s; it is not semantic coverage.

Automatic approval review rejected an attempted weakening of the unknown-read fixture and pending-view annotation. That attempt was not executed. The safe resolution retained all unknown assertions and implemented the numeric readiness/tag path, which passed Node parity. The first-unit checked-view skip remains as inherited from the source commits.
