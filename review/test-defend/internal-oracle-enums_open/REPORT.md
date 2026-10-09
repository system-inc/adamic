Fresh-write location contract defended within the bounded matrix.
Sparse-array and import-cycle rows not defended after three production attempts each.
All five added tests included; production sources restored; no test changed.

[
  {
    "test": "TestStage3EnumSparseArrayBoundary",
    "package": "internal/oracle",
    "prior_verdict": "subsumed",
    "subsumed_by": "TestNumericEnumNeverPinned",
    "defense": "not defended",
    "unique_mutant": null,
    "attempts": [
      {
        "mutant": "D2",
        "file_line": "internal/native/runtime/array.c:407",
        "change": "is outside an array of length -> is beyond an array of length",
        "rows_failed": [
          "TestNativeAgreesWithNode",
          "TestStage3EnumSparseArrayBoundary"
        ]
      },
      {
        "mutant": "D3",
        "file_line": "internal/javascript/javascript.go:91",
        "change": "is outside an array of length -> is beyond an array of length",
        "rows_failed": [
          "TestNativeAgreesWithNode",
          "TestStage3EnumSparseArrayBoundary",
          "TestTypedArrayWriteStopIsPinned"
        ]
      },
      {
        "mutant": "D4",
        "file_line": "internal/native/runtime/array.c:407, internal/native/runtime/typed_array.c:101, internal/javascript/javascript.go:91",
        "change": "is outside an array of length -> is beyond an array of length; is outside an array of length -> is beyond an array of length; is outside an array of length -> is beyond an array of length",
        "rows_failed": [
          "TestStage3EnumSparseArrayBoundary",
          "TestTypedArrayWriteStopIsPinned"
        ]
      }
    ],
    "evidence": "source /workspace/adamic-tools/env.sh; python3 review/test-defend/internal-oracle-enums_open/run.py; D2-matrix.log: enums_stage3_test.go:82: native: stderr differs; {stdout:[] stderr:[97 100 97 109 105 99 58 32 112 97 110 105 99 58 32 105 110 100 101 120 32 49 32 105 115 32 98 101 121 111 110 100 32 97 110 32 97 114 114 97 121 32 111 102 32 108 101 110 103 116 104 32 48 10] exitCode:70}",
    "bounded": true
  },
  {
    "test": "TestFreshWriteProbesStayRefused",
    "package": "internal/oracle",
    "prior_verdict": "subsumed",
    "subsumed_by": "TestStage3EnumBoundaries",
    "defense": "defended",
    "unique_mutant": "D1 internal/lower/fresh.go:69",
    "attempts": [
      {
        "mutant": "D1",
        "file_line": "internal/lower/fresh.go:69",
        "change": "return \"the write at \" + -> return \"the assignment at \" +",
        "rows_failed": [
          "TestFreshWriteProbesStayRefused"
        ]
      }
    ],
    "evidence": "source /workspace/adamic-tools/env.sh; python3 review/test-defend/internal-oracle-enums_open/run.py; D1-matrix.log: fresh_test.go:53: refused, but not naming the marked write (\"the write at /workspace/adamic/internal/oracle/testdata/fresh_refused/another_view.a:15:\"):",
    "bounded": true
  },
  {
    "test": "TestImportCycleRuntimeCalls",
    "package": "internal/oracle",
    "prior_verdict": "subsumed",
    "subsumed_by": "TestNumericEnumNeverPinned",
    "defense": "not defended",
    "unique_mutant": null,
    "attempts": [
      {
        "mutant": "D5",
        "file_line": "internal/lower/expression.go:784",
        "change": "ast.KindMinusToken:            ir.Subtract, -> ast.KindMinusToken:            ir.Remainder,",
        "rows_failed": [
          "TestImportCycleRuntimeCalls",
          "TestNestedSiblingCycleMutantIsCaught",
          "TestReviewProgramsAgreeWithNode"
        ]
      },
      {
        "mutant": "D6",
        "file_line": "internal/native/emit_expressions.go:754",
        "change": "return fmt.Sprintf(\"(%s %s %s)\", left, cOperators[operator], right) -> return fmt.Sprintf(\"(%s %s %s)\", right, cOperators[operator], left)",
        "rows_failed": [
          "TestEnumCleanupMutant",
          "TestImportCycleRuntimeCalls",
          "TestInputAgreesWithNode",
          "TestNativeAgreesWithNode",
          "TestNestedSiblingCycleMutantIsCaught",
          "TestReviewProgramsAgreeWithNode",
          "TestStage3EnumFallthrough"
        ]
      },
      {
        "mutant": "D7",
        "file_line": "internal/native/emit_branches.go:104",
        "change": "e.line(\"if (%s) {\", unwrap(condition)) -> e.line(\"if (!(%s)) {\", unwrap(condition))",
        "rows_failed": [
          "TestImportCycleRuntimeCalls",
          "TestInputAgreesWithNode",
          "TestNestedSiblingCycleMutantIsCaught",
          "TestReviewProgramsAgreeWithNode"
        ]
      }
    ],
    "evidence": "source /workspace/adamic-tools/env.sh; python3 review/test-defend/internal-oracle-enums_open/run.py; D5-matrix.log: import_cycles_test.go:73: stdout differs",
    "bounded": true
  }
]

## Mutants

| id | base file:line | change | failed rows |
|---|---|---|---|
| D1 | internal/lower/fresh.go:69 | return "the write at " + -> return "the assignment at " + | TestFreshWriteProbesStayRefused |
| D2 | internal/native/runtime/array.c:407 | is outside an array of length -> is beyond an array of length | TestNativeAgreesWithNode, TestStage3EnumSparseArrayBoundary |
| D3 | internal/javascript/javascript.go:91 | is outside an array of length -> is beyond an array of length | TestNativeAgreesWithNode, TestStage3EnumSparseArrayBoundary, TestTypedArrayWriteStopIsPinned |
| D4 | internal/native/runtime/array.c:407, internal/native/runtime/typed_array.c:101, internal/javascript/javascript.go:91 | is outside an array of length -> is beyond an array of length; is outside an array of length -> is beyond an array of length; is outside an array of length -> is beyond an array of length | TestStage3EnumSparseArrayBoundary, TestTypedArrayWriteStopIsPinned |
| D5 | internal/lower/expression.go:784 | ast.KindMinusToken:            ir.Subtract, -> ast.KindMinusToken:            ir.Remainder, | TestImportCycleRuntimeCalls, TestNestedSiblingCycleMutantIsCaught, TestReviewProgramsAgreeWithNode |
| D6 | internal/native/emit_expressions.go:754 | return fmt.Sprintf("(%s %s %s)", left, cOperators[operator], right) -> return fmt.Sprintf("(%s %s %s)", right, cOperators[operator], left) | TestEnumCleanupMutant, TestImportCycleRuntimeCalls, TestInputAgreesWithNode, TestNativeAgreesWithNode, TestNestedSiblingCycleMutantIsCaught, TestReviewProgramsAgreeWithNode, TestStage3EnumFallthrough |
| D7 | internal/native/emit_branches.go:104 | e.line("if (%s) {", unwrap(condition)) -> e.line("if (!(%s)) {", unwrap(condition)) | TestImportCycleRuntimeCalls, TestInputAgreesWithNode, TestNestedSiblingCycleMutantIsCaught, TestReviewProgramsAgreeWithNode |

## Passing rows for D1

TestEnumCleanupMutant, TestEnumInitializationNode, TestEnumInitializationUnknownPinned, TestEnumNameEnumerationMutant, TestEnumSemanticMutants, TestFallthroughMutants, TestFieldReadinessRepresentation, TestFractionalPowersReachRuntime, TestFunctionValueBoundaryBoxing, TestImportCycleLoadTimeReads, TestImportCycleRuntimeCalls, TestImportedNonliteralConstCaseIsNotYet, TestInputAgreesWithNode, TestModuleNamespaceLiveBindingMutant, TestModuleNamespaceReadinessMutants, TestModuleNamespaceReadsMatchNode, TestNativeAgreesWithNode, TestNestedCycleRefusal, TestNestedSiblingCycleMutantIsCaught, TestNumericEnumNeverPathsPinned, TestNumericEnumNeverPinned, TestReviewProgramsAgreeWithNode, TestReviewProgramsNoLooseFiles, TestReviewProgramsRefuse, TestReviewProgramsSelfTest, TestStage3EnumBoundaries, TestStage3EnumFallthrough, TestStage3EnumSparseArrayBoundary, TestTypedArrayWriteStopIsPinned. TestNativeAgreesWithNode passed only the selected 14 fixture subcases, listed in matrix.json. Added review rows contain skips, listed in skipped.json; skipped subcases are not evidence of passing.

## Findings and limits

# Code, oracle, scope, and defense limits

Base: e77a4ae41f473c149aee910c51b73637686a804a. Audit base: 859dee825a4ef89f5c4ecc0f00e6620ac75c4994.

Code under test: Adamic's lowerer and native/JavaScript backends, including fresh-write refusal diagnostics, ordinary-array assignment bounds, arithmetic and conditional emission, and imported module functions. Node, fixtures, tests, agreement helpers, and expected answers were unchanged.

Oracles:
- TestStage3EnumSparseArrayBoundary uses a self-written exact empty stdout, panic stderr, and exit-70 pin for native and JavaScript. Node is run but only its successful exit is checked.
- TestFreshWriteProbesStayRefused uses self-written rejection class, cycle-capable rule label, and the location of each fixture's marked closing write. Its ordinary rejection path does not execute Node.
- TestImportCycleRuntimeCalls runs Node with a fresh runner importing the mutually recursive functions, pins Node's output, adds calls to the already-lowered IR, and compares JavaScript and native output. It also checks native heap accounting. The fixture's generic run only runs module-body logging, not the imported functions after initialization.

Coverage profiles use -coverpkg=./internal/lower,./internal/native,./internal/javascript. Each target was measured separately from its named subsumer. Exclusive covered blocks are leads; they do not prove unique behavior. Go coverage does not instrument runtime C. The sparse-array path differs by ordinary-array indexed assignment; fresh refusal differs by exact marked-write location; import calls differ by mutually recursive execution history after initialization.

The whole clean package timed out at 90.066 binary seconds with no preceding test failure. The narrowed clean matrix and selected generic fixtures pass. The inventory grew from 193 to 198 tests. All five added tests are replayed separately for every mutant. Uniqueness is bounded to the recorded rows and selected generic fixture subcases. Other package rows and other packages are unknown. No deletion recommendation follows from these limited attempts.

Mutants were planned before running the defense matrix. D1 changes the marked-write diagnostic prefix. D2 changes native ordinary-array panic wording. D3 changes JavaScript array-write panic wording. D4 changes the same wording consistently in ordinary-array C, typed-array C, and JavaScript. This is a single shared-diagnostic fault implemented in three constant sites; the diff is intentionally multi-file. D5 changes subtraction lowering to remainder to perturb recursive decrement. D6 swaps binary operands to perturb recursive decrement and comparisons. D7 flips the emitted conditional branch to perturb recursive termination. These use change-constant/option, swap-arguments, and flip-condition operations. No switch, oracle, test, or harness mutation was used.

Each standalone diff applies to the recorded main base. All seven variants passed go vet for lower/native/javascript. Native binaries in the matrix were rebuilt using each variant's separate ADAMIC_BUILD_CACHE_DIR. The matrix also supplies native product compilation evidence for C edits. ADAMIC_GATE_UNCACHED=1 prevents cached agreement observations.

Disk check: /tmp has an 8.8 GB total filesystem, so it cannot meet a 15 GB free requirement. It had 8.6 GB free initially, 8.6 GB after removing the earlier /tmp/defend-typed scratch directory, and 8.5 GB during this run. /workspace had 16 GB free. No repository or tools were deleted. No disk-related failure occurred.

Toolchain was warm, so no setup install ran. nproc=5, Go go1.27.1, Node v24.19.0. npm ci in stage3/api completed and reported 346 ms. No other node_modules directory is loaded by these target rows. Exact command wall times are in commands.json and new-commands.json; test-binary elapsed times are in the JSON logs. Building and execution are not separately instrumented in those totals.

Brief friction: the whole-package baseline cannot finish inside the budget, so the matrix is bounded. The original audit's subsumption rests on one coarse dropped-main mutant for two targets and one diagnostic mutant for fresh refusal. Native C reachability cannot be read from Go coverage. The fresh defense proves a diagnostic contract, not a new cycle-safety rejection. Witness failures in the raw matrix can be broken preconditions and are not proof that their own disagreement guard works. The review lane contains pending sidecars; its skipped subcases are recorded rather than treated as passes. The instruction to include added tests required an additional clean baseline and seven matrix replays. The disk threshold exceeds /tmp's total capacity.

Names and assertions: the sparse-array row checks a specific intentional panic boundary, and the import-cycle row actually invokes imported mutually recursive functions after initialization. Neither name promises a behavior absent from its assertions. Failure to find uniqueness in these attempts is not proof of redundancy. There are no cost rows or Node/native twins among the three assigned rows.


## Run times

| variant | matrix binary seconds | generic fixtures binary seconds | added tests binary seconds |
|---|---|---|---|
| clean | 9.974 | 1.811 | 34.402 |
| D1 | 10.213 | 1.554 | 33.43 |
| D2 | 23.372 | 1.726 | 49.532 |
| D3 | 9.747 | 1.549 | 34.02 |
| D4 | 12.611 | 1.619 | 50.436 |
| D5 | 9.463 | 1.589 | 33.822 |
| D6 | 9.817 | 1.643 | 32.073 |
| D7 | 9.697 | 1.63 | 33.002 |

Wall time, vet results, and exact selectors are in commands.json and new-commands.json. D1 is unique among observed rows, not a proof of whole-package or repository uniqueness. Diagnostic coverage is the defense; fresh-cycle acceptance was not mutated. No named prior subsumer caught the aimed faults in its assigned target. D5-D7 are also caught by TestReviewProgramsAgreeWithNode. Other witness failures may be broken preconditions.
