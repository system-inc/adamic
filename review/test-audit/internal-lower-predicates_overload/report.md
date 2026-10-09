u042 audited at origin/main 7709c91213f476eba7e3dbfbf4ee65e6988038cb; all twelve requested names exist.
Verdicts: 4 sacred, 5 subsumed, 1 overlapping, 2 untrue under the fixed production plan.
Twenty mutations: nineteen killed; M18 survived with a changed helper output.
M01 is bounded after a Go panic; nineteen columns completed 239 tests; two rows accept empty answers.
Evidence on test-audit/internal-lower-predicates_overload under review/test-audit/internal-lower-predicates_overload/; production restored.

```json
[
  {
    "test": "TestPredicateOverloadRuntime",
    "package": "internal/lower",
    "file": "internal/lower/predicates_overload_test.go",
    "seconds": 5.981,
    "oracle": "Node runs original fixtures and validates handwritten stdout; native and backend JS match those validated values. Checked exit 70, diagnostic text and counters are self-written contracts.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M01",
      "M15",
      "M16",
      "M17"
    ],
    "unique_kills": [
      "M15",
      "M16",
      "M17"
    ],
    "last_proven_fail": "M17: predicates_overload_test.go:95: predicate counts: {Proven:2 Checked:0 Unobservable:0 Sites:[{Where:/tmp/adamic-gate/TestPredicateOverloadRuntimeoverload_some_empty3637780986/001/main.a:7:20 Function:some Overload:1 Directions:[{Direction:true Status:unobservable Reason:no narrowed read in the true region} {Direction:false Status:unobservable Reason:no narrowed read in the false region}]}]}, want [2 0 2] (proven, checked, unobservable)",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestPredicateOverloadRuntime",
      "TestIndirectPredicateOverloadIsPending",
      "TestPredicateBodyProof",
      "TestConditionAssertionAdmission",
      "TestPredicateCallbackContracts",
      "TestEveryNeedsCallbackEffects",
      "TestPredicateOverloadCallback",
      "TestPredicateUseRegions",
      "TestPredicateUsesBelongToEachCall",
      "TestUnprovenPredicateReturnsAreRefused",
      "TestPredicateBodiesAreProven",
      "TestPrimitiveAdmittingSlotsUseBoxes"
    ],
    "bounded_mutants": [
      "M01"
    ],
    "evidence": "ADAMIC_MUTANT=M17 ADAMIC_BUILD_CACHE_DIR=/tmp/u042/cache/M17 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M17.log 2>&1; predicates_overload_test.go:95: predicate counts: {Proven:2 Checked:0 Unobservable:0 Sites:[{Where:/tmp/adamic-gate/TestPredicateOverloadRuntimeoverload_some_empty3637780986/001/main.a:7:20 Function:some Overload:1 Directions:[{Direction:true Status:unobservable Reason:no narrowed read in the true region} {Direction:false Status:unobservable Reason:no narrowed read in the false region}]}]}, want [2 0 2] (proven, checked, unobservable)"
  },
  {
    "test": "TestIndirectPredicateOverloadIsPending",
    "package": "internal/lower",
    "file": "internal/lower/predicates_overload_test.go",
    "seconds": 0.049,
    "oracle": "Handwritten proof/admission/IR expectations; no external authority checked. Checks diagnostic substring, not diagnostic type.",
    "oracle_kind": "self",
    "kills": [
      "M06"
    ],
    "unique_kills": [],
    "last_proven_fail": "M06: predicates_overload_test.go:130: want explicit indirect-call capability gap, got /tmp/adamic-gate/TestIndirectPredicateOverloadIsPending3273718591/001/main.a:1:131: stage 0 can't lower an overloaded function as a value yet",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestPredicateOverloadCallback"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 0.081,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestPredicateOverloadRuntime",
      "TestIndirectPredicateOverloadIsPending",
      "TestPredicateBodyProof",
      "TestConditionAssertionAdmission",
      "TestPredicateCallbackContracts",
      "TestEveryNeedsCallbackEffects",
      "TestPredicateOverloadCallback",
      "TestPredicateUseRegions",
      "TestPredicateUsesBelongToEachCall",
      "TestUnprovenPredicateReturnsAreRefused",
      "TestPredicateBodiesAreProven",
      "TestPrimitiveAdmittingSlotsUseBoxes"
    ],
    "bounded_mutants": [
      "M01"
    ],
    "evidence": "ADAMIC_MUTANT=M06 ADAMIC_BUILD_CACHE_DIR=/tmp/u042/cache/M06 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M06.log 2>&1; predicates_overload_test.go:130: want explicit indirect-call capability gap, got /tmp/adamic-gate/TestIndirectPredicateOverloadIsPending3273718591/001/main.a:1:131: stage 0 can't lower an overloaded function as a value yet",
    "subsumption_mutants": 1
  },
  {
    "test": "TestPredicateBodyProof",
    "package": "internal/lower",
    "file": "internal/lower/predicates_proof_test.go",
    "seconds": 0.451,
    "oracle": "Handwritten proof/admission/IR expectations; no external authority checked. TypeScript 6.0.3 supplies positive input bodies, not expected Adamic proof results.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M05",
      "M08",
      "M09",
      "M10",
      "M11",
      "M12",
      "M13",
      "M14"
    ],
    "unique_kills": [
      "M09",
      "M10",
      "M11"
    ],
    "last_proven_fail": "M14: predicates_proof_test.go:70: want \"normal return\" refusal with caller fix, got {TaggedView:false}, <nil>",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P02"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestPredicateOverloadRuntime",
      "TestIndirectPredicateOverloadIsPending",
      "TestPredicateBodyProof",
      "TestConditionAssertionAdmission",
      "TestPredicateCallbackContracts",
      "TestEveryNeedsCallbackEffects",
      "TestPredicateOverloadCallback",
      "TestPredicateUseRegions",
      "TestPredicateUsesBelongToEachCall",
      "TestUnprovenPredicateReturnsAreRefused",
      "TestPredicateBodiesAreProven",
      "TestPrimitiveAdmittingSlotsUseBoxes"
    ],
    "bounded_mutants": [
      "M01"
    ],
    "evidence": "ADAMIC_MUTANT=M14 ADAMIC_BUILD_CACHE_DIR=/tmp/u042/cache/M14 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M14.log 2>&1; predicates_proof_test.go:70: want \"normal return\" refusal with caller fix, got {TaggedView:false}, <nil>",
    "vacuous_subcases": [
      "TestPredicateBodyProof/condition_assertion",
      "TestPredicateBodyProof/typeof",
      "TestPredicateBodyProof/helper",
      "TestPredicateBodyProof/diagnostic_work_on_failure",
      "TestPredicateBodyProof/assertion"
    ]
  },
  {
    "test": "TestConditionAssertionAdmission",
    "package": "internal/lower",
    "file": "internal/lower/predicates_proof_test.go",
    "seconds": 0.053,
    "oracle": "Handwritten proof/admission/IR expectations; no external authority checked.",
    "oracle_kind": "self",
    "kills": [
      "M13"
    ],
    "unique_kills": [],
    "last_proven_fail": "M13: predicates_proof_test.go:108: condition assertion must pass the production admission seam: /tmp/adamic-gate/TestConditionAssertionAdmission56825249/001/assert.a:2:57: Adamic 0.1 refuses a type predicate whose return is not proven (asserts cond needs a boolean parameter); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestPredicateBodyProof"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [],
    "subsumer_seconds": 0.451,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestPredicateOverloadRuntime",
      "TestIndirectPredicateOverloadIsPending",
      "TestPredicateBodyProof",
      "TestConditionAssertionAdmission",
      "TestPredicateCallbackContracts",
      "TestEveryNeedsCallbackEffects",
      "TestPredicateOverloadCallback",
      "TestPredicateUseRegions",
      "TestPredicateUsesBelongToEachCall",
      "TestUnprovenPredicateReturnsAreRefused",
      "TestPredicateBodiesAreProven",
      "TestPrimitiveAdmittingSlotsUseBoxes"
    ],
    "bounded_mutants": [
      "M01"
    ],
    "evidence": "ADAMIC_MUTANT=M13 ADAMIC_BUILD_CACHE_DIR=/tmp/u042/cache/M13 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M13.log 2>&1; predicates_proof_test.go:108: condition assertion must pass the production admission seam: /tmp/adamic-gate/TestConditionAssertionAdmission56825249/001/assert.a:2:57: Adamic 0.1 refuses a type predicate whose return is not proven (asserts cond needs a boolean parameter); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)",
    "subsumption_mutants": 1
  },
  {
    "test": "TestPredicateCallbackContracts",
    "package": "internal/lower",
    "file": "internal/lower/predicates_proof_test.go",
    "seconds": 0.263,
    "oracle": "Handwritten proof/admission/IR expectations; no external authority checked.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M04",
      "M05",
      "M06",
      "M07"
    ],
    "unique_kills": [],
    "last_proven_fail": "M07: predicates_proof_test.go:134: /tmp/adamic-gate/TestPredicateCallbackContractsinferred_arrow890367569/001/main.a:1:45: Adamic 0.1 refuses a type predicate whose return is not proven (there is no body proving this parameter); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)",
    "verdict": "overlapping",
    "subsumed_by": [
      "TestArrayPredicateCannotInventAnElementContract",
      "TestArrayPredicateCoexistsWithUnknownReflection",
      "TestArrayPredicatePreservesDeclaredElementContract",
      "TestCensusPredicateInteractionBoundary",
      "TestCensusPredicateInteractionEscapes",
      "TestCensusPredicateMarkerKeepsProofBoundaries",
      "TestEveryNeedsCallbackEffects",
      "TestIndirectPredicateOverloadIsPending",
      "TestPredicateBodiesAreProven",
      "TestPredicateBodyProof",
      "TestPredicateOverloadCallback",
      "TestPredicateOverloadRuntime",
      "TestUnprovenPredicateReturnsAreRefused"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestPredicateOverloadRuntime",
      "TestIndirectPredicateOverloadIsPending",
      "TestPredicateBodyProof",
      "TestConditionAssertionAdmission",
      "TestPredicateCallbackContracts",
      "TestEveryNeedsCallbackEffects",
      "TestPredicateOverloadCallback",
      "TestPredicateUseRegions",
      "TestPredicateUsesBelongToEachCall",
      "TestUnprovenPredicateReturnsAreRefused",
      "TestPredicateBodiesAreProven",
      "TestPrimitiveAdmittingSlotsUseBoxes"
    ],
    "bounded_mutants": [
      "M01"
    ],
    "evidence": "ADAMIC_MUTANT=M07 ADAMIC_BUILD_CACHE_DIR=/tmp/u042/cache/M07 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M07.log 2>&1; predicates_proof_test.go:134: /tmp/adamic-gate/TestPredicateCallbackContractsinferred_arrow890367569/001/main.a:1:45: Adamic 0.1 refuses a type predicate whose return is not proven (there is no body proving this parameter); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)",
    "vacuous_subcases": [
      "TestPredicateCallbackContracts/inferred_arrow",
      "TestPredicateCallbackContracts/inferred_named",
      "TestPredicateCallbackContracts/arrow",
      "TestPredicateCallbackContracts/named"
    ]
  },
  {
    "test": "TestEveryNeedsCallbackEffects",
    "package": "internal/lower",
    "file": "internal/lower/predicates_proof_test.go",
    "seconds": 0.05,
    "oracle": "Handwritten proof/admission/IR expectations; no external authority checked.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M04",
      "M05"
    ],
    "unique_kills": [],
    "last_proven_fail": "M05: predicates_proof_test.go:160: want array proof refusal without callback effects, got <nil>",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestPredicateCallbackContracts"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P03"
    ],
    "subsumer_seconds": 0.263,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestPredicateOverloadRuntime",
      "TestIndirectPredicateOverloadIsPending",
      "TestPredicateBodyProof",
      "TestConditionAssertionAdmission",
      "TestPredicateCallbackContracts",
      "TestEveryNeedsCallbackEffects",
      "TestPredicateOverloadCallback",
      "TestPredicateUseRegions",
      "TestPredicateUsesBelongToEachCall",
      "TestUnprovenPredicateReturnsAreRefused",
      "TestPredicateBodiesAreProven",
      "TestPrimitiveAdmittingSlotsUseBoxes"
    ],
    "bounded_mutants": [
      "M01"
    ],
    "evidence": "ADAMIC_MUTANT=M05 ADAMIC_BUILD_CACHE_DIR=/tmp/u042/cache/M05 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M05.log 2>&1; predicates_proof_test.go:160: want array proof refusal without callback effects, got <nil>",
    "subsumption_mutants": 3
  },
  {
    "test": "TestPredicateOverloadCallback",
    "package": "internal/lower",
    "file": "internal/lower/predicates_proof_test.go",
    "seconds": 0.081,
    "oracle": "Handwritten proof/admission/IR expectations; no external authority checked.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M05",
      "M06"
    ],
    "unique_kills": [],
    "last_proven_fail": "M06: predicates_proof_test.go:175: want lying overload callback argument refused at call site, got /tmp/adamic-gate/TestPredicateOverloadCallback705287358/002/main.a:7:58: Adamic 0.1 refuses a type predicate whose return is not proven (true return narrows to number, not number); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestPredicateCallbackContracts"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 0.263,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestPredicateOverloadRuntime",
      "TestIndirectPredicateOverloadIsPending",
      "TestPredicateBodyProof",
      "TestConditionAssertionAdmission",
      "TestPredicateCallbackContracts",
      "TestEveryNeedsCallbackEffects",
      "TestPredicateOverloadCallback",
      "TestPredicateUseRegions",
      "TestPredicateUsesBelongToEachCall",
      "TestUnprovenPredicateReturnsAreRefused",
      "TestPredicateBodiesAreProven",
      "TestPrimitiveAdmittingSlotsUseBoxes"
    ],
    "bounded_mutants": [
      "M01"
    ],
    "evidence": "ADAMIC_MUTANT=M06 ADAMIC_BUILD_CACHE_DIR=/tmp/u042/cache/M06 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M06.log 2>&1; predicates_proof_test.go:175: want lying overload callback argument refused at call site, got /tmp/adamic-gate/TestPredicateOverloadCallback705287358/002/main.a:7:58: Adamic 0.1 refuses a type predicate whose return is not proven (true return narrows to number, not number); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)",
    "subsumption_mutants": 3
  },
  {
    "test": "TestPredicateUseRegions",
    "package": "internal/lower",
    "file": "internal/lower/predicates_proof_test.go",
    "seconds": 0.526,
    "oracle": "Handwritten proof/admission/IR expectations; no external authority checked. Only the assignment-flow mutation reached the direction checker chain; no production kill was demonstrated in this plan.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P04"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestPredicateOverloadRuntime",
      "TestIndirectPredicateOverloadIsPending",
      "TestPredicateBodyProof",
      "TestConditionAssertionAdmission",
      "TestPredicateCallbackContracts",
      "TestEveryNeedsCallbackEffects",
      "TestPredicateOverloadCallback",
      "TestPredicateUseRegions",
      "TestPredicateUsesBelongToEachCall",
      "TestUnprovenPredicateReturnsAreRefused",
      "TestPredicateBodiesAreProven",
      "TestPrimitiveAdmittingSlotsUseBoxes"
    ],
    "bounded_mutants": [
      "M01"
    ],
    "evidence": "Probe only: ADAMIC_MUTANT=P04 ADAMIC_BUILD_CACHE_DIR=/tmp/u042/cache/P04 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run ^TestPredicateUseRegions$ > P04-TestPredicateUseRegions.log 2>&1; predicates_proof_test.go:233: directions 0, want 1",
    "vacuous_subcases": [
      "TestPredicateUseRegions/unused",
      "TestPredicateUseRegions/replacement"
    ]
  },
  {
    "test": "TestPredicateUsesBelongToEachCall",
    "package": "internal/lower",
    "file": "internal/lower/predicates_proof_test.go",
    "seconds": 0.049,
    "oracle": "Handwritten proof/admission/IR expectations; no external authority checked. Only the assignment-flow mutation reached the direction checker chain; no production kill was demonstrated in this plan.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P04"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestPredicateOverloadRuntime",
      "TestIndirectPredicateOverloadIsPending",
      "TestPredicateBodyProof",
      "TestConditionAssertionAdmission",
      "TestPredicateCallbackContracts",
      "TestEveryNeedsCallbackEffects",
      "TestPredicateOverloadCallback",
      "TestPredicateUseRegions",
      "TestPredicateUsesBelongToEachCall",
      "TestUnprovenPredicateReturnsAreRefused",
      "TestPredicateBodiesAreProven",
      "TestPrimitiveAdmittingSlotsUseBoxes"
    ],
    "bounded_mutants": [
      "M01"
    ],
    "evidence": "Probe only: ADAMIC_MUTANT=P04 ADAMIC_BUILD_CACHE_DIR=/tmp/u042/cache/P04 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run ^TestPredicateUsesBelongToEachCall$ > P04-TestPredicateUsesBelongToEachCall.log 2>&1; predicates_proof_test.go:274: call 1: directions 0, want 1"
  },
  {
    "test": "TestUnprovenPredicateReturnsAreRefused",
    "package": "internal/lower",
    "file": "internal/lower/predicates_test.go",
    "seconds": 0.51,
    "oracle": "Handwritten proof/admission/IR expectations; no external authority checked.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M02",
      "M03",
      "M04",
      "M05",
      "M08",
      "M12",
      "M14"
    ],
    "unique_kills": [
      "M03"
    ],
    "last_proven_fail": "M14: predicates_test.go:44: got <nil>, want a predicate refusal naming \"normal return\" and its rewrite",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestPredicateOverloadRuntime",
      "TestIndirectPredicateOverloadIsPending",
      "TestPredicateBodyProof",
      "TestConditionAssertionAdmission",
      "TestPredicateCallbackContracts",
      "TestEveryNeedsCallbackEffects",
      "TestPredicateOverloadCallback",
      "TestPredicateUseRegions",
      "TestPredicateUsesBelongToEachCall",
      "TestUnprovenPredicateReturnsAreRefused",
      "TestPredicateBodiesAreProven",
      "TestPrimitiveAdmittingSlotsUseBoxes"
    ],
    "bounded_mutants": [
      "M01"
    ],
    "evidence": "ADAMIC_MUTANT=M14 ADAMIC_BUILD_CACHE_DIR=/tmp/u042/cache/M14 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M14.log 2>&1; predicates_test.go:44: got <nil>, want a predicate refusal naming \"normal return\" and its rewrite"
  },
  {
    "test": "TestPredicateBodiesAreProven",
    "package": "internal/lower",
    "file": "internal/lower/predicates_test.go",
    "seconds": 0.341,
    "oracle": "Handwritten proof/admission/IR expectations; no external authority checked.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M02",
      "M04"
    ],
    "unique_kills": [],
    "last_proven_fail": "M04: predicates_test.go:67: function isText(x: string | undefined): x is string { return x !== undefined; }: /tmp/adamic-gate/TestPredicateBodiesAreProven384616363/003/main.a:1:55: Adamic 0.1 refuses a type predicate whose return is not proven (true return narrows to undefined, not string); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestUnprovenPredicateReturnsAreRefused"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [],
    "subsumer_seconds": 0.51,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestPredicateOverloadRuntime",
      "TestIndirectPredicateOverloadIsPending",
      "TestPredicateBodyProof",
      "TestConditionAssertionAdmission",
      "TestPredicateCallbackContracts",
      "TestEveryNeedsCallbackEffects",
      "TestPredicateOverloadCallback",
      "TestPredicateUseRegions",
      "TestPredicateUsesBelongToEachCall",
      "TestUnprovenPredicateReturnsAreRefused",
      "TestPredicateBodiesAreProven",
      "TestPrimitiveAdmittingSlotsUseBoxes"
    ],
    "bounded_mutants": [
      "M01"
    ],
    "evidence": "ADAMIC_MUTANT=M04 ADAMIC_BUILD_CACHE_DIR=/tmp/u042/cache/M04 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M04.log 2>&1; predicates_test.go:67: function isText(x: string | undefined): x is string { return x !== undefined; }: /tmp/adamic-gate/TestPredicateBodiesAreProven384616363/003/main.a:1:55: Adamic 0.1 refuses a type predicate whose return is not proven (true return narrows to undefined, not string); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)",
    "subsumption_mutants": 3
  },
  {
    "test": "TestPrimitiveAdmittingSlotsUseBoxes",
    "package": "internal/lower",
    "file": "internal/lower/primitive_slots_test.go",
    "seconds": 0.051,
    "oracle": "Handwritten proof/admission/IR expectations; no external authority checked.",
    "oracle_kind": "self",
    "kills": [
      "M20"
    ],
    "unique_kills": [
      "M20"
    ],
    "last_proven_fail": "M20: primitive_slots_test.go:26: {} binding is 4, want boxed union",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestPredicateOverloadRuntime",
      "TestIndirectPredicateOverloadIsPending",
      "TestPredicateBodyProof",
      "TestConditionAssertionAdmission",
      "TestPredicateCallbackContracts",
      "TestEveryNeedsCallbackEffects",
      "TestPredicateOverloadCallback",
      "TestPredicateUseRegions",
      "TestPredicateUsesBelongToEachCall",
      "TestUnprovenPredicateReturnsAreRefused",
      "TestPredicateBodiesAreProven",
      "TestPrimitiveAdmittingSlotsUseBoxes"
    ],
    "bounded_mutants": [
      "M01"
    ],
    "evidence": "ADAMIC_MUTANT=M20 ADAMIC_BUILD_CACHE_DIR=/tmp/u042/cache/M20 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M20.log 2>&1; primitive_slots_test.go:26: {} binding is 4, want boxed union"
  }
]
```

All file locations are against the starting origin/main. M01 failed rows come from individual bounded reruns; kills outside them are unknown.

| ID | File:line | Change | Failed rows |
|---|---|---|---|
| M01 | internal/lower/predicates.go:82 | if changed != nil { -> if changed == nil { | TestEveryNeedsCallbackEffects, TestPredicateBodiesAreProven, TestPredicateBodyProof, TestPredicateCallbackContracts, TestPredicateOverloadCallback, TestPredicateOverloadRuntime, TestUnprovenPredicateReturnsAreRefused |
| M02 | internal/lower/predicates.go:170 | literal && truth != (expression.Kind == ast.KindTrueKeyword) -> literal && truth == (expression.Kind == ast.KindTrueKeyword) | TestArrayPredicatePreservesDeclaredElementContract, TestPredicateBodiesAreProven, TestUnprovenPredicateReturnsAreRefused |
| M03 | internal/lower/predicates.go:318 | case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression: 		// A property read may dispatch a getter through a structural view. 		return true -> case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression: 		// A property read may dispatch a getter through a structural view. 		return false | TestUnprovenPredicateReturnsAreRefused |
| M04 | internal/lower/predicates.go:128 | flags = ast.FlowFlagsTrueCondition -> flags = ast.FlowFlagsFalseCondition | TestArrayPredicateCoexistsWithUnknownReflection, TestArrayPredicatePreservesDeclaredElementContract, TestEveryNeedsCallbackEffects, TestPredicateBodiesAreProven, TestPredicateCallbackContracts, TestUnprovenPredicateReturnsAreRefused |
| M05 | internal/lower/predicates.go:357 | if original == nil { -> if original != nil { | TestArrayPredicateCannotInventAnElementContract, TestCensusPredicateInteractionBoundary, TestCensusPredicateInteractionEscapes, TestCensusPredicateMarkerKeepsProofBoundaries, TestEveryNeedsCallbackEffects, TestPredicateBodyProof, TestPredicateCallbackContracts, TestPredicateOverloadCallback, TestUnprovenPredicateReturnsAreRefused |
| M06 | internal/lower/predicates.go:477 | if signature == nil { 		return nil -> if signature != nil { 		return nil | TestIndirectPredicateOverloadIsPending, TestPredicateCallbackContracts, TestPredicateOverloadCallback |
| M07 | internal/lower/predicates.go:472 | return found && safe -> return found && !safe | TestCensusPredicateInteractionEscapes, TestPredicateCallbackContracts |
| M08 | internal/lower/predicates_proof.go:62 | proof.TaggedView = true -> proof.TaggedView = false | TestPredicateBodyProof, TestUnprovenPredicateReturnsAreRefused |
| M09 | internal/lower/predicates_proof.go:99 | changed = true -> changed = false | TestPredicateBodyProof |
| M10 | internal/lower/predicates_proof.go:295 | (t&predicateTrue)<<1 -> (t&predicateTrue)>>1 | TestPredicateBodyProof |
| M11 | internal/lower/predicates_proof.go:305 | if and { -> if !and { | TestPredicateBodyProof |
| M12 | internal/lower/predicates_proof.go:366 | yes = v.cells[path.cell] == constant.Text() -> yes = v.cells[path.cell] != constant.Text() | TestPredicateBodyProof, TestPredicateSummaryParameterIndex, TestUnprovenPredicateReturnsAreRefused |
| M13 | internal/lower/predicates_proof.go:618 | case ast.KindThrowStatement: 		return true -> case ast.KindThrowStatement: 		return false | TestConditionAssertionAdmission, TestPredicateBodyProof |
| M14 | internal/lower/predicates_proof.go:545 | cells: []string{"truthy", "falsy"} -> cells: []string{"truthy"} | TestPredicateBodyProof, TestUnprovenPredicateReturnsAreRefused |
| M15 | internal/lower/predicates_proof.go:715 | return proof.proveBody(overload.Type()) == nil -> return proof.proveBody(overload.Type()) != nil | TestPredicateOverloadRuntime |
| M16 | internal/lower/predicates_proof.go:1119 | ir.Binary{Operator: ir.NotEqual, Left: returned, Right: membership} -> ir.Binary{Operator: ir.Equal, Left: returned, Right: membership} | TestPredicateOverloadRuntime |
| M17 | internal/lower/predicates_proof.go:1226 | counts.Unobservable++ ->  | TestPredicateOverloadRuntime |
| M18 | internal/lower/predicates_proof.go:1038 | if flow.Flags&ast.FlowFlagsAssignment != 0 && l.predicateSameReference(flow.Node, reference) && predicateWriteOnly(flow.Node) { 		return 0 -> if flow.Flags&ast.FlowFlagsAssignment != 0 && l.predicateSameReference(flow.Node, reference) && predicateWriteOnly(flow.Node) { 		return predicateEither |  |
| M19 | internal/lower/expression.go:49 | if flags&(checker.TypeFlagsUnknown\|checker.TypeFlagsNonPrimitive) != 0 { 		return ir.Union, true -> if flags&(checker.TypeFlagsUnknown\|checker.TypeFlagsNonPrimitive) != 0 { 		return ir.Object, true | TestUnknownReflectionRefusals |
| M20 | internal/lower/expression.go:90 | // the runtime brand with the same boxes used for scalar/reference unions. 		return ir.Union, true -> // the runtime brand with the same boxes used for scalar/reference unions. 		return ir.Object, true | TestPrimitiveAdmittingSlotsUseBoxes |

Survivor M18: assignment helper directions 0 -> 3; entry directions 0 -> 0. witness-fixed-clean.log and witness-fixed-M18.log record actual outputs. This shows an unguarded helper result, not a demonstrated whole-checker or native miscompile.

The starting origin/main is 7709c91213f476eba7e3dbfbf4ee65e6988038cb, not the older commit printed in the brief. All twelve requested names remain in their named files. The clean package passed: 239 top-level tests. Warm env.sh worked, setup was skipped, and stage3/api npm ci ran before the baseline.

The code under test is Adamic's Go lowering: predicate flow and body proof, callback admission, overload result checks, directional-use tracking and primitive-admitting storage. Node, TypeScript's checker, native emission, JavaScript emission, tests and fixture expectations were not mutated. Native products received a distinct ADAMIC_BUILD_CACHE_DIR for every compiler mutation.

The runtime row validates source output with Node and compares native and backend JavaScript against the same validated values. Its checked failure exit 70, complete stderr and proof counters are handwritten contracts. The other rows use handwritten expected diagnostics, admission outcomes, proof flags or IR slots. TypeScript 6.0.3 supplied some positive source bodies, not an independently computed expected Adamic proof result. No external-authority answer was checked.

These are separate rows because their assertions differ. PredicateOverloadCallback's edited lying source is a negative admission input, like the negative cases in CallbackContracts. It is not a witness of a separate comparison harness: both inputs exercise production lowering and expected admission.

The mutation menu was frozen before any kill result. All twenty mutations are condition flips, constant/option changes or statement drops in production. There are no supplemental verdict mutations. The function inventory conservatively includes every production function; it is not an exact dynamic coverage list. Direct proof and admission entries bypass much of Lower's orchestration, so they received their own probes.

M01 caused a real Go panic. Only seven top-level tests completed in its whole-package run. Each requested row was then run alone. M01 kills outside those twelve rows remain unknown. The other nineteen columns completed all 239 tests. The report marks the combined matrix bounded and lists the bounded column explicitly; unique kills in complete columns are package-wide observations.

The initial panic detector also matched quoted native panic text in M15 and M16 failure messages. It conservatively reran twelve scoped rows for each. Completion checks establish that these were complete columns, not Go aborts. This unnecessary work cost time. The AST inventory generator also initially failed on an unnamed receiver, before any mutation test ran; the corrected generator handles it.

Empty-answer probes are not verdict mutations. Lower returning nil,nil is accepted by PredicateBodiesAreProven; refuse returning nil is accepted by ConditionAssertionAdmission. Those two rows are vacuous under their own entries. Other rows reject empty answers, sometimes through a nil-answer panic. Passing subcases are listed separately for body proof, callbacks and use regions.

The two direction rows have no production kill in this fixed plan. Only M18, an assignment-flow helper change, exercised their distinctive chain. Their P04 empty-direction probe fails, proving that they can reject zero directions. Under the brief's production-mutant rule their verdict is untrue for this plan. This is a narrow result from one relevant helper fault, not a claim that those tests cannot fail or should be deleted.

Subsumption is only a hint. IndirectOverload and ConditionAdmission each rest on one shared kill. EveryNeedsCallbackEffects, OverloadCallback and PredicateBodiesAreProven each rest on three. The subsumer medians are reported even when the subsumer is slower; the brief leaves the keep decision to the defender wave. CallbackContracts overlaps multiple rows without one covering its whole kill set.

Diagnostic assertions have different strength: UnprovenPredicateReturnsAreRefused checks Refused type, reason and fix; IndirectPredicateOverloadIsPending checks a substring only. M06 changes its observed refusal to another capability gap, so that catch proves the diagnostic distinction rather than successful execution.

The survivor witness initially used a numeric console argument, which the .a prelude rejects. That attempt produced a checker error before the mutated helper ran. The corrected fixture uses toString; both attempts are preserved. The final witness records helper and entry outputs separately so a private helper change is not mistaken for a demonstrated miscompile.

The approximate twenty-minute budget was exceeded modestly by the complete matrix and additional panic-marker replays. No individual test run timed out. Reported row medians are binary elapsed values from three -count=1 invocations, not Go command wall time. Build and validation timing are recorded separately where measured.

No tests in other packages were run. Repo-wide uniqueness, exhaustive dynamic reachability and a native miscompile from the surviving helper change were not established. Production sources were restored before committing. Every standalone diff applies to the starting source and passes go vet ./internal/lower/.

Measured timings:

```json
{
  "start_commit": "7709c91213f476eba7e3dbfbf4ee65e6988038cb",
  "nproc": 5,
  "setup": "skipped, warm env works",
  "npm_seconds": 0.6006796199981181,
  "baseline_binary_seconds": 36.405,
  "baseline_command_seconds": 38.2770248170018,
  "baseline_skips": [
    "TestOriginalCycleLedger",
    "TestOptionalWideningCensus",
    "TestMixedUnionContractGraph/interface_Node_{readonly_ready:boolean}_type_Target=Node|readonly_Node[];"
  ],
  "timing_binary_seconds": 25.534,
  "matrix_and_replay_command_seconds": 798.6642764399949,
  "matrix_and_replay_binary_seconds": 662.83,
  "probe_command_seconds": 31.696068991004722,
  "standalone_vet_seconds": 10.436739956006932,
  "clean_build_seconds": 2.2822656300013477,
  "switch_build_seconds": "not separately timed; included in matrix command overhead",
  "finished_utc": "2026-10-09T10:47:48.044317+00:00",
  "verdict_counts": {
    "sacred": 4,
    "subsumed": 5,
    "overlapping": 1,
    "untrue": 2
  }
}
```

Baseline skips outside this unit: TestOriginalCycleLedger, TestOptionalWideningCensus, TestMixedUnionContractGraph/interface_Node_{readonly_ready:boolean}_type_Target=Node|readonly_Node[];. Reasons and uniqueness limits are in limitations.md.
