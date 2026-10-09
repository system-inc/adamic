u039: all ten requested functions exist at f91994f019703ba25d2918cf529c0e0b0c05d93c; none moved or vanished.
Three cycle cases form one family, leaving eight audit rows.
Clean whole-package baseline passed in 21.589 s; nproc=5; warm toolchain setup skipped.
Four production mutants: two sacred rows, two subsumed rows; one proven witness; three cannot-judge rows.
Rest support is vacuous under PLower; every production mutant was killed; no production changes retained.

```json
[
  {
    "test": "TestNestedFunctionGapsAreLoud",
    "package": "internal/lower",
    "file": "internal/lower/nested_functions_test.go:16",
    "seconds": 0.093,
    "oracle": "Self-written NotYet type and diagnostic substrings for block, generic-value and dynamic-this gaps.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "PLower"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=PLower ADAMIC_BUILD_CACHE_DIR=/tmp/u039/cache/PLower timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run ^TestNestedFunctionGapsAreLoud$ > PLower-TestNestedFunctionGapsAreLoud.log 2>&1; nested_functions_test.go:27: want NotYet \"block-scoped nested\", got <nil>; nested_functions_test.go:27: want NotYet \"generic function as a value\", got <nil>; nested_functions_test.go:27: want NotYet \"dynamic this\", got <nil>; no production kills in M01-M04.",
    "reason": "Four compiler mutants exhausted the stated cap; no meaningful change to this row's specific guard or rest-parameter handling was exercised. No untrue verdict without that honest try."
  },
  {
    "test": "TestNestedCycle family",
    "package": "internal/lower",
    "file": "internal/lower/nested_functions_test.go:33",
    "seconds": 0.066,
    "oracle": "Self-written Refused type and adamic/cycle-capable substring; all three cases use the same checker with different source inputs.",
    "oracle_kind": "self",
    "kills": [
      "M03"
    ],
    "unique_kills": [],
    "last_proven_fail": "M03: nested_functions_test.go:44: want cycle refusal, got <nil>",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestLiteralMethodCapturesCannotMakeCycles"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "PLower"
    ],
    "subsumer_seconds": 0.038,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M03 ADAMIC_BUILD_CACHE_DIR=/tmp/u039/cache/M03 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M03.log 2>&1; M03: nested_functions_test.go:44: want cycle refusal, got <nil>",
    "members": [
      "TestNestedFunctionCycleIsRefused",
      "TestNestedEnvironmentCycleIncludesDisjointSlots",
      "TestNestedCallbackCycleIsRefused"
    ],
    "subsumption_mutants": 1
  },
  {
    "test": "TestNestedEnvironmentHasOneAllocationSite",
    "package": "internal/lower",
    "file": "internal/lower/nested_functions_test.go:48",
    "seconds": 0.036,
    "oracle": "Self-written IR invariant: exactly one allocation, two cells, complete layout and linked EnvironmentCell flags.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M02"
    ],
    "unique_kills": [
      "M02"
    ],
    "last_proven_fail": "M02: nested_functions_test.go:78: want one environment allocation, got 0",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "PLower"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M02 ADAMIC_BUILD_CACHE_DIR=/tmp/u039/cache/M02 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M02.log 2>&1; M02: nested_functions_test.go:78: want one environment allocation, got 0"
  },
  {
    "test": "TestNestedCapturedParametersAreOwned",
    "package": "internal/lower",
    "file": "internal/lower/nested_functions_test.go:99",
    "seconds": 0.034,
    "oracle": "Self-written IR invariant: a captured string parameter exists and is not Borrowed.",
    "oracle_kind": "self",
    "kills": [
      "M01"
    ],
    "unique_kills": [],
    "last_proven_fail": "M01: nested_functions_test.go:115: missing captured parameter",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestNestedEnvironmentHasOneAllocationSite"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "PLower"
    ],
    "subsumer_seconds": 0.036,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M01 ADAMIC_BUILD_CACHE_DIR=/tmp/u039/cache/M01 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M01.log 2>&1; M01: nested_functions_test.go:115: missing captured parameter",
    "subsumption_mutants": 1
  },
  {
    "test": "TestClosedFrameInputRejectsMutation",
    "package": "internal/lower",
    "file": "internal/lower/nested_functions_test.go:119",
    "seconds": 0.035,
    "oracle": "Self-written proof predicate accepts a literal input and rejects the built-in SetProperty mutation; W01 alone decides witness verdict.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W01: nested_functions_test.go:144: mutable graph accepted as closed",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "PClosedFrame"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=W01 ADAMIC_BUILD_CACHE_DIR=/tmp/u039/cache/W01 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run ^(TestNestedFunctionGapsAreLoud|TestNestedFunctionCycleIsRefused|TestNestedEnvironmentHasOneAllocationSite|TestNestedEnvironmentCycleIncludesDisjointSlots|TestNestedCapturedParametersAreOwned|TestClosedFrameInputRejectsMutation|TestNestedRebindingNotYet|TestNestedCallbackCycleIsRefused|TestNestedBodylessDeclarationsAreLoud|TestNestedRestIsSupported)$ > W01.log 2>&1; W01: nested_functions_test.go:144: mutable graph accepted as closed"
  },
  {
    "test": "TestNestedRebindingNotYet",
    "package": "internal/lower",
    "file": "internal/lower/nested_functions_test.go:150",
    "seconds": 0.037,
    "oracle": "Self-written lowering NotYet type and rebinding substring; TS2630 is bypassed, not used as the expected answer.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "PDeclareModule",
      "PStatements"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=PDeclareModule ADAMIC_BUILD_CACHE_DIR=/tmp/u039/cache/PDeclareModule timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run ^(TestNestedFunctionGapsAreLoud|TestNestedFunctionCycleIsRefused|TestNestedEnvironmentHasOneAllocationSite|TestNestedEnvironmentCycleIncludesDisjointSlots|TestNestedCapturedParametersAreOwned|TestClosedFrameInputRejectsMutation|TestNestedRebindingNotYet|TestNestedCallbackCycleIsRefused|TestNestedBodylessDeclarationsAreLoud|TestNestedRestIsSupported)$ > PDeclareModule.log 2>&1; nested_functions_test.go:177: want canonical NotYet, got /workspace/adamic/internal/oracle/refusals/nested_rebinding.a:7:20: stage 0 can't lower reading outer yet; no production kills in M01-M04.",
    "reason": "Four compiler mutants exhausted the stated cap; no meaningful change to this row's specific guard or rest-parameter handling was exercised. No untrue verdict without that honest try."
  },
  {
    "test": "TestNestedBodylessDeclarationsAreLoud",
    "package": "internal/lower",
    "file": "internal/lower/nested_functions_test.go:197",
    "seconds": 0.044,
    "oracle": "Self-written NotYet type and exact bodyless diagnostic substrings on two direct entries.",
    "oracle_kind": "self",
    "kills": [
      "M04"
    ],
    "unique_kills": [
      "M04"
    ],
    "last_proven_fail": "M04: nested_functions_test.go:229: want NotYet \"a nested function declaration without an implementation\", got <nil>",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "PNestedDeclarations",
      "PLowerBody"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M04 ADAMIC_BUILD_CACHE_DIR=/tmp/u039/cache/M04 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M04.log 2>&1; M04: nested_functions_test.go:229: want NotYet \"a nested function declaration without an implementation\", got <nil>",
    "vacuous_subcases": {
      "PNestedDeclarations": [],
      "PLowerBody": []
    }
  },
  {
    "test": "TestNestedRestIsSupported",
    "package": "internal/lower",
    "file": "internal/lower/nested_functions_test.go:235",
    "seconds": 0.035,
    "oracle": "Checks only successful lowering, discards returned IR; PLower passes with nil output and nil error.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=PLower ADAMIC_BUILD_CACHE_DIR=/tmp/u039/cache/PLower timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run ^TestNestedRestIsSupported$ > PLower-TestNestedRestIsSupported.log 2>&1; --- PASS: TestNestedRestIsSupported (0.03s); no production kills in M01-M04.",
    "reason": "Four compiler mutants exhausted the stated cap; no meaningful change to this row's specific guard or rest-parameter handling was exercised. No untrue verdict without that honest try."
  }
]
```

| ID | origin/main file:line | Change | Failed top-level tests |
|---|---|---|---|
| M01 | internal/lower/nested_functions.go:270 | change constant: l.result.Locals[local].EnvironmentCell = true -> l.result.Locals[local].EnvironmentCell = false | TestNestedCapturedParametersAreOwned, TestNestedEnvironmentHasOneAllocationSite |
| M02 | internal/lower/nested_functions.go:275 | drop statement: function.Body = append([]ir.Statement{ir.AllocateEnvironment{Cells: slices.Clone(function.FrameEnvironment)}}, function.Body...) -> (removed) | TestNestedEnvironmentHasOneAllocationSite |
| M03 | internal/lower/cycles.go:412 | drop whole loop: for _, local := range f.l.result.Functions[closure.function].Environment { 					queue = append(queue, cycleNode{cell: local + 1}) 				} -> (removed) | TestLiteralMethodCapturesCannotMakeCycles, TestNamespaceAmbientHostInitialization, TestNestedCallbackCycleIsRefused, TestNestedEnvironmentCycleIncludesDisjointSlots, TestNestedFunctionCycleIsRefused, TestWhatZeroOneRefusesIsRefusedWithAFix |
| M04 | internal/lower/nested_functions.go:41 | flip condition: if !implemented { -> if implemented { | TestNestedBodylessDeclarationsAreLoud |
| W01 | internal/lower/closed_frame_inputs.go:66 | witness weakening: return safe && called -> return safe \|\| called | TestClosedFrameInputRejectsMutation |

Survivors: none. W01 is a witness weakening, not a production mutant.

The supplied file reference is 8de93800f4, but fetching origin/main started this audit at f91994f019703ba25d2918cf529c0e0b0c05d93c. The test file is unchanged between those commits. All source lines and diffs refer to the actual starting commit.

The brief asks for about three mutants per row and also limits compiler mutants with separate rebuilds to four. I interpreted the compiler rebuild guidance conservatively as a four-mutant limit. The switch avoids recompiling Go for each run, and these ten rows inspect lowering rather than native output, so that interpretation limited coverage more than necessary. It leaves the gap, rebinding and rest guards without an honest targeted attempt at their specific behavior. Those rows are cannot-judge, not untrue. The fixed plan was saved before test outcomes. The initial witness weakening `return called` failed vet because `safe` became unused. Before running tests it was replaced by `return safe || called`; both weakening and correction are recorded here.

The three nested cycle tests differ only in source input and use the same Refused/cycle-capable comparison. They are grouped as TestNestedCycle family, with members named in results.json. Their separate timings are retained, plus three runs of the grouped family. The family is subsumed by TestLiteralMethodCapturesCannotMakeCycles on one observed mutant; this is a small-matrix hint, not a deletion recommendation. CapturedParameters is subsumed by EnvironmentHasOneAllocationSite on one observed mutant.

The whole package fit the budget on all four production runs, so no narrowing was needed. The enabled package matrix is complete; repo-wide kills are unknown. The baseline skipped TestOriginalCycleLedger (external pristine corpus not configured), TestOptionalWideningCensus (external project inventory not configured), and one TestMixedUnionContractGraph subcase with an explicit compiler/views-v3 prerequisite. No requested test skipped. Those skipped checks remain unknown; no installable tool was missing for these ten rows.

PLower panicked in IR-inspection rows. All ten functions were rerun alone for that probe; only their observed results are recorded. Probe failures never count as production kills. ClosedFrameInput's PClosedFrame can fail during its production lowering precondition; only W01's mutable-graph assertion establishes the witness verdict. Two direct bodyless entries each reject their own empty probe; their unaffected sibling subcases pass, which is expected because they call different entries.

Warm setup took zero setup-script seconds. npm ci succeeded before baseline; its wall duration was not separately recorded. Clean test-binary compilation took 1.871 s. Per-mutant vet build validation is in validation.json. Production build-and-run wall times were M01: 29.372 s, M02: 23.372 s, M03: 23.585 s, M04: 24.050 s; their binary run times were M01: 22.034 s, M02: 21.663 s, M03: 21.897 s, M04: 22.357 s. Compilation/startup overhead is included in wall times and was not separately timed per mutant. All mutation, witness and probe runs total 131.746 wall seconds; 36 clean timing invocations total 61.901 wall seconds. Baseline binary time was 21.589 s; clean coverage run was 0.215 s. Four mutants rather than twenty, the three unjudged guard paths, the skipped corpora, and repo-wide replay are not covered. No survivors exist in the four-mutant matrix.
