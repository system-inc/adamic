Built: d04ed5d3's net compiler changes squashed onto current main, with three arrow capture frame checks.
Commits: source d04ed5d3064bbe5981224d52473a63a62bdb5707; base main c0fd8945b872e57a341ccee5973e3f4d1b75c34b; final commit SHA is in the delivery message.
Checks: focused lowering and oracle tests, fifteen Node/sanitizer fixtures, seven stage3 status rows, TestCallTargetReaders, changed-package go vet, counts.
Mutants: fourteen source overlays fail their named tests; seven IR output mutations are caught by Node comparisons in both backends.
Not covered: full packages/full gate, WASI, all existing count measurements, and the three dependency-blocked tasks already excluded by the source.

Branch: compiler/loom-chain-followers. No PR or Loom submission. Main stayed at the base above during the final fetch, so no rebase was needed.

The source merge base is 5e33a17b186a8a2218d27b69b21e2de5acc5b750. The requested git merge --squash d04ed5d3 had exactly one conflict: internal/oracle/counts.md, two appended groups at the same insertion point. Preserved main's three getters_census rows and source's seven generic/phantom rows. One successful regeneration placed the rows in registration order, preserving every old row's values.

Runtime files changed against main: none under internal/native/runtime. Compiler implementation files changed:

- internal/flow/build.go
- internal/ir/phantom_brands.go
- internal/javascript/javascript.go
- internal/lower/cast_proof.go
- internal/lower/census_small.go
- internal/lower/class_inheritance.go
- internal/lower/empty_arrays.go
- internal/lower/exceptions.go
- internal/lower/expression.go
- internal/lower/generic.go
- internal/lower/generic_function_value.go
- internal/lower/invariance.go
- internal/lower/library_string.go
- internal/lower/object.go
- internal/lower/phantom_brands.go
- internal/lower/proven_relations.go
- internal/lower/refusals.go
- internal/lower/statements.go
- internal/native/emit_expressions.go
- internal/native/phantom_brands.go

Counts: exactly fifteen added rows; no old row changed or disappeared. Allocations, frees, retains, releases, peak and regions are listed below. These are the fifteen new oracle fixtures, and each was measured once in the successful regeneration. counts-attribution.json independently compares every old row to main.

| Fixture | Allocations | Frees | Retains | Releases | Peak | Regions |
|---|---:|---:|---:|---:|---:|---:|
| internal/oracle/testdata/arrow_this_super.a | 5 | 5 | 8 | 8 | 5 | 0 |
| internal/oracle/testdata/arrow_this_field.a | 7 | 7 | 9 | 11 | 6 | 0 |
| internal/oracle/testdata/arrow_this_stored.a | 7 | 7 | 9 | 11 | 7 | 0 |
| internal/oracle/testdata/native-generic-empty-array.a | 3 | 3 | 1 | 5 | 3 | 0 |
| internal/oracle/testdata/generic_empty_array.a | 32 | 32 | 24 | 60 | 17 | 0 |
| internal/oracle/testdata/generic_function_value_arms.a | 18 | 18 | 72 | 99 | 7 | 0 |
| internal/oracle/testdata/generic_function_value.a | 15 | 15 | 27 | 45 | 7 | 0 |
| internal/oracle/testdata/phantom_brands.a | 15 | 15 | 29 | 42 | 6 | 0 |
| internal/oracle/testdata/phantom_undefined.a | 1 | 0 | 5 | 4 | 1 | 0 |
| internal/oracle/testdata/phantom_catch.a | 8 | 8 | 10 | 18 | 2 | 0 |
| internal/oracle/testdata/host_phantom_overload_primitive.a | 25 | 25 | 13 | 43 | 7 | 0 |
| internal/oracle/testdata/host_phantom_optional_array.a | 5 | 5 | 2 | 7 | 2 | 0 |
| internal/oracle/testdata/scanner_expressions/string-number-concatenation.a | 2 | 2 | 3 | 6 | 2 | 0 |
| internal/oracle/testdata/scanner_expressions/number_spellings.a | 85 | 85 | 2 | 88 | 3 | 0 |
| internal/oracle/testdata/scanner_expressions/primitive_concatenation.a | 33 | 33 | 2 | 38 | 3 | 0 |

Rules 11 and 12: no measured test was 60 seconds or more, so none needed splitting or deletion. Every new/touched test remains present. The new IR checks are top-level Test functions, each with t.Parallel. Their seconds: TestArrowFieldCaptureFrames 0.070, TestArrowSuperCaptureFrames 0.080, TestArrowStoredCaptureFrames 0.080. TestCallTargetReaders passed in 23.500 test seconds (40.965 command seconds including compilation). Every test outcome and seconds, including nested leaves and selected family wrappers, is in test-times.tsv and test-times.json. Tests ran with GOMAXPROCS=4 and -parallel 4 on cpu.max=400000 100000, nproc=5. Products/tests introduce no build-directory or machine-path literals. No new compiler-consuming package exists, so cloud/fast-gate/compiler-dependencies.json needs no change.

Commands and observed outputs: checks.json contains complete argv, exits, and command wall seconds. All test stdout/stderr went directly to log files; no test output was piped. Tests use -count=1 -timeout 90s, subprocess hard limits 240 seconds, and overall monitored limits 900 seconds. The source's lowering tests and oracle mutants passed. Fifteen fixture leaves passed using TestNativeAgreesWithNode_[0-9]+ with the specific new paths, including Node truth, JS backend, ASan/UBSan and leak checks. Seven status rows passed through TestFixturesAssertions/Cycles/Namespaces, selected by directory and file, retaining original Node observations. go vet passed on ./internal/flow ./internal/ir ./internal/javascript ./internal/lower ./internal/native ./internal/oracle. Changed Go files are gofmt clean.

The initial fixture/count selector combined slash-containing paths inside one alternation, which Go splits into components. That fixture run matched no leaves, and the counts run exited 1 explicitly refusing to write a table. The first stage3 selector also matched no rows. Corrected selectors ran all fifteen fixtures and seven status rows; logs retain the initial empty runs and the actual results separately. Counts were successfully written only once, using the fifteen changed fixture leaves, followed by TestCountsAreRecordedTable. No test crossed 60 seconds or was rerun after crossing it.

Mutants and what caught each:

- non-null-check: TestCheckedNonNullTypeScript/undefined, TestCheckedNonNullTypeScript. non_null_checked_test.go:43: want proven 0 checked 1, got ir.NonNullCheckCounts{Proven:0, Checked:0, Sites:[]ir.NonNullCheck(nil)}
- non-null-narrowing: TestCheckedNonNullTypeScript/narrowed, TestCheckedNonNullTypeScript. non_null_checked_test.go:40: want proven 1 checked 0, got ir.NonNullCheckCounts{Proven:0, Checked:1, Sites:[]ir.NonNullCheck{ir.NonNullCheck{Where:"/workspace/adamic/internal/oracle/testdata/non_null_checked_narrowed.ts:2:43", Expression:"value!", Proven:false}}}
- nested-empty: TestNestedEmptyArrayElementKinds/strings, TestNestedEmptyArrayElementKinds/objects, TestNestedEmptyArrayElementKinds. empty_literal_test.go:35: empty element kind 1, want 3 | empty_literal_test.go:35: empty element kind 1, want 4
- string-concatenation: TestNativeAgreesWithNode_081/internal/oracle/testdata/scanner_expressions/primitive_concatenation.a, TestNativeAgreesWithNode_081. oracle_test.go:766: Lower: /workspace/adamic/internal/oracle/testdata/scanner_expressions/primitive_concatenation.a:3:13: stage 0 can't lower a BinaryExpression with a string and a value yet
- phantom-primitive: TestPhantomCastsAreErased. phantom_brands_test.go:64: /tmp/adamic-gate/TestPhantomCastsAreErased622513172/001/main.a:2:48: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)
- proven-predicates: TestCensusComputedRelation. Named test failure recorded in its log.
- generic-empty: TestNativeAgreesWithNode_059/internal/oracle/testdata/generic_empty_array.a, TestNativeAgreesWithNode_059. oracle_test.go:766: Lower: /workspace/adamic/internal/oracle/testdata/generic_empty_array.a:33:22: Adamic 0.1 refuses a value of type never seen as T, a type parameter whose constraint { text: string; } can be written, so it can write what never can't hold; take it as never, or constrain T to something readonly, which can't write (adamic/invariant-mutable)
- debugger: TestDebuggerNativeEmitsNothing/.ts, TestDebuggerNativeEmitsNothing/.a, TestDebuggerNativeEmitsNothing. debugger_test.go:45: /tmp/adamic-gate/TestDebuggerNativeEmitsNothing.ts2547499085/001/main.ts:4:5: Adamic 0.1 refuses debugger; remove it | debugger_test.go:45: /tmp/adamic-gate/TestDebuggerNativeEmitsNothing.a684804024/001/main.a:4:5: Adamic 0.1 refuses debugger; remove it
- generic-values: TestGenericFunctionValuesHaveSeparateInstances. generic_function_value_test.go:19: /tmp/adamic-gate/TestGenericFunctionValuesHaveSeparateInstances3286443157/001/main.a:4:47: stage 0 can't lower a generic function as a value yet
- generic-namespace-interaction: TestGenericFunctionValueNamespaceDirectCallsKeepIdentity. generic_function_value_test.go:151: /tmp/adamic-gate/TestGenericFunctionValueNamespaceDirectCallsKeepIdentity2623227545/001/main.a:6:56: stage 0 can't lower function identity comparison in a program with specialized generic function values yet
- phantom-overload-interaction: TestPhantomOverloadOnlyBrandsKeepMainProof. phantom_brands_test.go:156: /tmp/adamic-gate/TestPhantomOverloadOnlyBrandsKeepMainProof1180697418/001/main.a:1:13: Adamic 0.1 refuses a primitive brand member pathBrand whose type is not void; make pathBrand void (or optional and typed undefined) so the brand is phantom
- proven-cleanup: TestCensusGenericRefusalInsideClosure. Named test failure recorded in its log.
- arrow-constructor-frame: TestArrowFieldCaptureFrames. arrow_capture_frames_test.go:10: receiver frames: got 0 empty and 2 receiver arrows, want 1 and 1
- arrow-receiver-capture: TestArrowSuperCaptureFrames, TestArrowStoredCaptureFrames. arrow_capture_frames_test.go:15: receiver frames: got 1 empty and 0 receiver arrows, want 0 and 1 | arrow_capture_frames_test.go:20: receiver frames: got 1 empty and 0 receiver arrows, want 0 and 1

The seven in-test IR output mutants are empty-array falsiness, fresh-number-array sharing, wrong generic result, numeric concatenation spelling, boolean spelling, undefined spelling, and null spelling. TestEmptyArrayRuntimeMutants, TestGenericFunctionValueWrongResultMutant, TestScannerConcatenationMutant, and TestScannerPrimitiveConcatenationMutants require clean execution and stdout-only disagreement with Node in both native and JS backends. Their passing test logs are the evidence that each mutant was caught. The two new arrow source mutants are caught by IR frame assertions before any native emission: constructor frame isolation removed yields two receiver arrows instead of one empty and one receiver arrow; receiver capture suppressed yields an empty arrow instead of a receiver arrow. Both super and stored checks fail for the latter. Old source reports' arrow mutants failed at C compilation, which is why three independent lowering checks were added.

Conservative scope assumption: d04ed5d3 is the authoritative requested net change. Its two commits already keep #rh16bxg and #4my3dqg, including phantom-brand representation, while excluding #agwccbw, #p6a00tt and #k881crd for unavailable dependencies. This recut does not import those excluded changes or reinterpret the source. Original evidence remains under review/compiler/chain-followers; new evidence is under review/compiler/loom-chain-followers. There was no code copied from cohere; generic inference references its pinned checker through shim/linkname.

Setup: exported GOPROXY=https://proxy.golang.org|direct, ran timeout 600 bash cloud/setup.sh with log output, and sourced /workspace/adamic-tools/env.sh. Setup succeeded, including its real heap-buffer-overflow sanitizer probe. Timing lines, cumulative seconds: Node ready 0.087; Go ready 0.092; clang ready 0.566; Markdown dependency install step 0.674, dependencies ready 0.919; submodules ready 153.714; shared cache ready 158.860; Go build ready 416.525; test binaries deferred 416.610; build cache warm 416.612; done 416.639. nproc=5; CPU quota=4; Go 1.27.1, clang 20.1.8, Node v24.19.0. Full timing/build flags are in setup.log.

Limitations: the unchanged main counts were preserved rather than remeasured, as this worker runs focused checks and Loom owns the full gate. The source evidence has historical whitespace in patch/log witnesses; git diff --check reports that and a pre-existing source fixture's extra EOF blank line. Those evidence bytes were preserved. The three skipped source tasks remain dependency-blocked as described in the source REPORT.md; this report makes no claim that they are implemented.
