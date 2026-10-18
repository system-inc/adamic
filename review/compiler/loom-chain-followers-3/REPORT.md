Built: recut chain followers on current main, preserving lowering and adding conservative freshness compatibility.
Commits: source d4c561e317fd66aaf029fdd1ca12084c12e3af9e; base d79302e50c022c997c00f7c68ddb7475f9ee65fc; delivery SHA in final response.
Commands and outputs: exact bounded commands, exit codes and wall seconds in checks.json; final positive checks passed.
Mutants: all 14 recorded source mutants, seven recorded output mutants, and one added freshness mutant caught.
Not covered: whole packages, full Loom gate, WASI; the explicit no-whole-packages instruction was followed with focused consumer checks.

No squash conflicts. Git automatically merged counts.md, which was then replaced with main's table before exactly one Linux -update-counts run. Main 1f07ce2d removed the digest column again and moved digests into measurement-product keys. This newer infrastructure is retained. Only the fifteen source fixtures were measured; the complete file diff has exactly fifteen added rows and no other changed line. counts.patch and counts-attribution.json list every row.

The additional production fix is internal/fresh/library_language.go: ir.PhantomMember now uses the existing conservative opaque-call model, retaining every operand effect, escaping operands, clobbering parameters and treating the result as outside. Assumption: make no new freshness claim, even though a successful phantom read produces undefined. TestFreshCorpusRemainder exposed the missing node in four phantom fixtures; removing the case again is the added mutant, caught there. No consumer test was changed. Two nested source tests were also made parallel in generic_function_value_test.go and phantom_brands_test.go.

Runtime +/- lines are IDENTICAL to the source: both diffs under internal/native/runtime are empty. runtime-comparison.json contains the arrays. No runtime line differs.

Guards satisfied:

- TestEveryMapAProgramHoldsIsSorted passed. The new node holds no map.
- TestRuntimeSetsAreEachUsed passed. No runtime set was added.
- TestNativeAgreementRowsNameRegisteredPrograms passed. No new native-agreement call was added.
- TestEveryTestIsParallelOrSaysWhy passed. All top-level source tests comply; two nested tests were also repaired.
- TestCallTargetReaders passed with -timeout 90s, again after the last test edits.
- No new package consumes the compiler, so compiler-dependencies.json requires no new entry. No product or test adds a build-directory or machine-path literal.

Focused verification includes source Node, JavaScript backend, sanitized native and leak checks for the fifteen fixtures, seven output mutants, changed stage3 status rows, flow and freshness remainders, native and JavaScript consumers, guards and vet. Vet covers internal/flow, ir, javascript, lower, native, oracle and fresh. All test commands use -count=1 -timeout 90s -parallel 4, GOMAXPROCS=4 and ADAMIC_GATE_UNCACHED=1. Output goes directly to files. Each test's seconds are in test-times.tsv and test-times.json, including parents and intended failures. No test reached 60 seconds; none was deleted or split. The cold counts command includes compilation separately from test time.

Setup: the initial attempt was interrupted with exit 143 during an oversized TypeScript promisor history fetch. Fetching the exact pinned d92d9bfee114c80be2c375d72edae966176e3a4f with --depth=1 succeeded. The next attempt exited 1 because GIT_NO_LAZY_FETCH=1 also disabled a blob needed for checkout. Normal blob fetching then completed setup, exit 0. The printed /workspace/adamic-tools/env.sh was sourced. nproc=5, cpu.max=400000 100000 (four CPUs). clang 20.1.8, Go 1.27.1, Node v24.19.0. Final cumulative timings: Node 0.019s; Go 0.021s; Markdown 0.058s; clang 0.129s; submodules 33.572s; shared cache 38.397s; Go build 324.920s; tests deferred 325.035s; cache warm 325.036s; done 325.089s. All attempts and recovery are logged. An early guard invocation failed before tests because the submodule go.mod was absent; its actual later run passed.

No cohere code was copied. Generic inference refers to the pinned shim. Historical evidence remains byte for byte; git diff --check reports inherited whitespace in historical patch/log files and a fixture EOF, while Go source formatting is clean. No PR or Loom submission was made. The new branch is pushed without force.

Commands and observed outcomes:

- counts-regeneration: exit 0, 127.205s.
- counts-table: exit 0, 3.259s.
- lower-focus: exit 0, 12.96s.
- oracle-focus: exit 0, 8.933s.
- ir-focus: exit 0, 15.993s.
- fixtures: exit 0, 8.078s.
- stage3: exit 0, 42.48s.
- flow-consumers: exit 0, 19.927s.
- fresh-consumers: exit 1, 7.311s.
- fresh-consumers: exit 0, 13.452s.
- native-consumers: exit 0, 8.813s.
- javascript-consumers: exit 0, 1.824s.
- guards-ir: exit 0, 8.41s.
- guards-runtime: exit 0, 3.701s.
- guards-registration: exit 0, 8.575s.
- guards-parallel: exit 0, 1.117s.
- readers: exit 0, 17.117s.
- vet: exit 0, 38.028s.
- mutant-phantom-fresh: exit 1, 3.789s.
- mutant-non-null-check: exit 1, 13.429s.
- mutant-non-null-narrowing: exit 1, 13.694s.
- mutant-nested-empty: exit 1, 9.418s.
- mutant-string-concatenation: exit 1, 13.593s.
- mutant-phantom-primitive: exit 1, 9.674s.
- mutant-proven-predicates: exit 1, 10.878s.
- mutant-generic-empty: exit 1, 13.461s.
- mutant-debugger: exit 1, 13.686s.
- mutant-generic-values: exit 1, 10.336s.
- mutant-generic-namespace-interaction: exit 1, 10.01s.
- mutant-phantom-overload-interaction: exit 1, 9.627s.
- mutant-proven-cleanup: exit 1, 9.467s.
- mutant-arrow-constructor-frame: exit 1, 9.21s.
- mutant-arrow-receiver-capture: exit 1, 9.209s.
- final-parallel-rows: exit 0, 9.373s.
- final-fixtures: exit 0, 7.173s.
- final-readers: exit 0, 8.236s.
- final-parallel-guard: exit 0, 1.019s.
- final-vet: exit 0, 1.733s.

Source mutants and their catchers:

- arrow-constructor-frame: TestArrowFieldCaptureFrames (0.03s).
- arrow-receiver-capture: TestArrowSuperCaptureFrames (0.04s), TestArrowStoredCaptureFrames (0.04s).
- debugger: TestDebuggerNativeEmitsNothing/.ts (0.04s), TestDebuggerNativeEmitsNothing/.a (0.04s), TestDebuggerNativeEmitsNothing (0s).
- generic-empty: TestNativeAgreesWithNode_059/internal/oracle/testdata/generic_empty_array.a (0.03s), TestNativeAgreesWithNode_059 (0s).
- generic-namespace-interaction: TestGenericFunctionValueNamespaceDirectCallsKeepIdentity (0.04s).
- generic-values: TestGenericFunctionValuesHaveSeparateInstances (0.04s).
- nested-empty: TestNestedEmptyArrayElementKinds/strings (0.05s), TestNestedEmptyArrayElementKinds/objects (0.06s), TestNestedEmptyArrayElementKinds (0s).
- non-null-check: TestCheckedNonNullTypeScript/undefined (0.03s), TestCheckedNonNullTypeScript (0s).
- non-null-narrowing: TestCheckedNonNullTypeScript/narrowed (0.03s), TestCheckedNonNullTypeScript (0s).
- phantom-fresh: TestFreshCorpusRemainder/../oracle/testdata/host_phantom_overload_primitive.a (0.02s), TestFreshCorpusRemainder/../oracle/testdata/phantom_brands.a (0.05s), TestFreshCorpusRemainder/../oracle/testdata/phantom_catch.a (0.08s), TestFreshCorpusRemainder/../oracle/testdata/phantom_undefined.a (0.03s), TestFreshCorpusRemainder (0.67s).
- phantom-overload-interaction: TestPhantomOverloadOnlyBrandsKeepMainProof (0.03s).
- phantom-primitive: TestPhantomCastsAreErased (0.03s).
- proven-cleanup: TestCensusGenericRefusalInsideClosure (0.03s).
- proven-predicates: TestCensusComputedRelation (0.03s).
- string-concatenation: TestNativeAgreesWithNode_081/internal/oracle/testdata/scanner_expressions/primitive_concatenation.a (0.03s), TestNativeAgreesWithNode_081 (0s).

Recorded output mutants: empty-array-falsiness; fresh-array-sharing; generic-wrong-result; numeric-concatenation-spelling; boolean-concatenation-spelling; undefined-concatenation-spelling; null-concatenation-spelling. Each is caught only by stdout disagreement against Node in both backends; its test rejects build, sanitizer, exit or leak failure as a catcher. Source mutants fail their named assertion or lowering check, rather than invalid C or a Go build failure.

Counts additions:

| internal/oracle/testdata/arrow_this_field.a | 7 | 7 | 9 | 11 | 6 | 0 |
| internal/oracle/testdata/arrow_this_stored.a | 7 | 7 | 9 | 11 | 7 | 0 |
| internal/oracle/testdata/arrow_this_super.a | 5 | 5 | 8 | 8 | 5 | 0 |
| internal/oracle/testdata/generic_empty_array.a | 32 | 32 | 24 | 60 | 17 | 0 |
| internal/oracle/testdata/generic_function_value.a | 15 | 15 | 27 | 45 | 7 | 0 |
| internal/oracle/testdata/generic_function_value_arms.a | 18 | 18 | 72 | 99 | 7 | 0 |
| internal/oracle/testdata/host_phantom_optional_array.a | 5 | 5 | 2 | 7 | 2 | 0 |
| internal/oracle/testdata/host_phantom_overload_primitive.a | 25 | 25 | 13 | 43 | 7 | 0 |
| internal/oracle/testdata/native-generic-empty-array.a | 3 | 3 | 1 | 5 | 3 | 0 |
| internal/oracle/testdata/phantom_brands.a | 15 | 15 | 29 | 42 | 6 | 0 |
| internal/oracle/testdata/phantom_catch.a | 8 | 8 | 10 | 18 | 2 | 0 |
| internal/oracle/testdata/phantom_undefined.a | 1 | 0 | 5 | 4 | 1 | 0 |
| internal/oracle/testdata/scanner_expressions/number_spellings.a | 85 | 85 | 2 | 88 | 3 | 0 |
| internal/oracle/testdata/scanner_expressions/primitive_concatenation.a | 33 | 33 | 2 | 38 | 3 | 0 |
| internal/oracle/testdata/scanner_expressions/string-number-concatenation.a | 2 | 2 | 3 | 6 | 2 | 0 |
