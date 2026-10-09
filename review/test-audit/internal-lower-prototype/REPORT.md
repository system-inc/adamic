u043: all 14 requested rows remain present at origin/main 5deb11d433947539ab6448558b8ab1f1904393a5.
Clean whole-package baseline passed in 24.211 binary seconds; warm tools; nproc=5.
19 production mutants plus M04 supplemental empty-hazard probe; whole-package matrices saved.
Verdicts: 7 sacred, 5 subsumed, 0 overlapping, 1 untrue, 1 cannot-judge.
Evidence pushed on test-audit/internal-lower-prototype under review/test-audit/internal-lower-prototype/.

```json
[
  {
    "test": "TestInheritedLibraryReadsNeverLoadOwnFields",
    "package": "internal/lower",
    "file": "internal/lower/prototype_test.go:20",
    "seconds": 3.226,
    "oracle": "Node enumerates Object.prototype; inherited-read refusal class and nil result are handwritten.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M01",
      "M07",
      "M08",
      "M09",
      "M10"
    ],
    "unique_kills": [],
    "last_proven_fail": "M10: prototype_test.go:64: /tmp/adamic-gate/TestInheritedLibraryReadsNeverLoadOwnFieldsfunctiontoLocaleStr2214049125/001/sweep.a:1:7: stage 0 can't lower a value of type () => number yet",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestPredicateOverloadRuntime"
    ],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "PProperty",
      "PElement"
    ],
    "subsumer_seconds": 5.094,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M10 ADAMIC_BUILD_CACHE_DIR=/tmp/u043/cache/M10 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M10.log 2>&1; prototype_test.go:64: /tmp/adamic-gate/TestInheritedLibraryReadsNeverLoadOwnFieldsfunctiontoLocaleStr2214049125/001/sweep.a:1:7: stage 0 can't lower a value of type () => number yet",
    "subsumption_kills": 5
  },
  {
    "test": "TestPrototypeMethodsAreRefusedWithReasons",
    "package": "internal/lower",
    "file": "internal/lower/prototype_test.go:86",
    "seconds": 0.322,
    "oracle": "Handwritten Refused class and nonempty repair text.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M03",
      "M06",
      "M09",
      "M10"
    ],
    "unique_kills": [
      "M06"
    ],
    "last_proven_fail": "M06: prototype_test.go:97: got /tmp/adamic-gate/TestPrototypeMethodsAreRefusedWithReasons15340168/002/main.a:1:21: Adamic 0.1 refuses toLocaleString on a number; , want a reasoned refusal",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "PLower"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M06 ADAMIC_BUILD_CACHE_DIR=/tmp/u043/cache/M06 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M06.log 2>&1; prototype_test.go:97: got /tmp/adamic-gate/TestPrototypeMethodsAreRefusedWithReasons15340168/002/main.a:1:21: Adamic 0.1 refuses toLocaleString on a number; , want a reasoned refusal"
  },
  {
    "test": "TestNullishPrototypeReadsAreRejectedByChecker",
    "package": "internal/lower",
    "file": "internal/lower/prototype_test.go:110",
    "seconds": 0.738,
    "oracle": "Node enumerates members; handwritten CheckError expectation from TypeScript checker runs. No lowering entry is reached.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 19,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": false,
    "evidence": "No lowering function reached; no checker/oracle mutation performed.",
    "cannot_judge_reason": "The requested row exercises load.Load and the external checker rather than internal/lower. The frozen production menu changes lowering only; a meaningful lowering mutant cannot reach it."
  },
  {
    "test": "TestPrototypeHazardsBehindObjectViewsAreNotYet",
    "package": "internal/lower",
    "file": "internal/lower/prototype_test.go:127",
    "seconds": 0.311,
    "oracle": "Handwritten NotYet class and reason substrings.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M03",
      "M07",
      "M17"
    ],
    "unique_kills": [
      "M17"
    ],
    "last_proven_fail": "M17: prototype_test.go:151: const value: {} = new RegExp('a'); console.log(value.toString());: got /tmp/adamic-gate/TestPrototypeHazardsBehindObjectViewsAreNotYet399609184/002/main.a:1:19: stage 0 can't lower RegExp with more than two arguments yet, want an explicit NotYet containing \"through an object view\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "PLower",
      "M04"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M17 ADAMIC_BUILD_CACHE_DIR=/tmp/u043/cache/M17 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M17.log 2>&1; prototype_test.go:151: const value: {} = new RegExp('a'); console.log(value.toString());: got /tmp/adamic-gate/TestPrototypeHazardsBehindObjectViewsAreNotYet399609184/002/main.a:1:19: stage 0 can't lower RegExp with more than two arguments yet, want an explicit NotYet containing \"through an object view\""
  },
  {
    "test": "TestUnrepresentedPrototypeCallsAreNotYet",
    "package": "internal/lower",
    "file": "internal/lower/prototype_test.go:175",
    "seconds": 0.232,
    "oracle": "Handwritten NotYet class only; alternate NotYet reasons can pass.",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M05"
    ],
    "unique_kills": [
      "M05"
    ],
    "last_proven_fail": "M05: prototype_test.go:190: const value = [1]; const kept = value.valueOf();: <nil>",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "PLower"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M05 ADAMIC_BUILD_CACHE_DIR=/tmp/u043/cache/M05 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M05.log 2>&1; prototype_test.go:190: const value = [1]; const kept = value.valueOf();: <nil>"
  },
  {
    "test": "TestIsPrototypeOfReadsExplainThePrototypeRefusal",
    "package": "internal/lower",
    "file": "internal/lower/prototype_test.go:195",
    "seconds": 0.068,
    "oracle": "Handwritten prototype-chain refusal and repair substring.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M02",
      "M08",
      "M10"
    ],
    "unique_kills": [
      "M02"
    ],
    "last_proven_fail": "M02: prototype_test.go:204: got /tmp/adamic-gate/TestIsPrototypeOfReadsExplainThePrototypeRefusal2580117112/001/main.a:1:36: Adamic 0.1 refuses inherited library member isPrototypeOf read as an own field; prototype members are not stored in an object's shape; call the method on its receiver, or wrap that call in an arrow (unbound-method)",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "PLower"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M02 ADAMIC_BUILD_CACHE_DIR=/tmp/u043/cache/M02 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M02.log 2>&1; prototype_test.go:204: got /tmp/adamic-gate/TestIsPrototypeOfReadsExplainThePrototypeRefusal2580117112/001/main.a:1:36: Adamic 0.1 refuses inherited library member isPrototypeOf read as an own field; prototype members are not stored in an object's shape; call the method on its receiver, or wrap that call in an arrow (unbound-method)"
  },
  {
    "test": "TestProvenRelationsRefuse",
    "package": "internal/lower",
    "file": "internal/lower/proven_relations_test.go:12",
    "seconds": 0.307,
    "oracle": "Handwritten Refused class and diagnostic substrings.",
    "oracle_kind": "self",
    "kills": [
      "M11",
      "M12"
    ],
    "unique_kills": [],
    "last_proven_fail": "M12: proven_relations_test.go:41: want refusal naming \"nominal ancestry for A\", got <nil>",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestCheckedCastProofAndElision"
    ],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "PLower"
    ],
    "subsumer_seconds": 0.149,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M12 ADAMIC_BUILD_CACHE_DIR=/tmp/u043/cache/M12 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M12.log 2>&1; proven_relations_test.go:41: want refusal naming \"nominal ancestry for A\", got <nil>",
    "subsumption_kills": 2
  },
  {
    "test": "TestProvenRelationsErase",
    "package": "internal/lower",
    "file": "internal/lower/proven_relations_test.go:48",
    "seconds": 0.103,
    "oracle": "Both native C and JavaScript artifacts must equal our own operand-only output. Shared wrong output can compare equal.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M07",
      "M11"
    ],
    "unique_kills": [],
    "last_proven_fail": "M11: proven_relations_test.go:59: /tmp/adamic-gate/TestProvenRelationsErase1993644486/002/main.a:1:105: Adamic 0.1 refuses an unproven relation from { readonly name: string; readonly extra: number; } to { readonly name: string; }: optional field  has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestDefaultTaggedInterfaceNeedsNoFlag"
    ],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "PLower"
    ],
    "subsumer_seconds": 0.04,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M11 ADAMIC_BUILD_CACHE_DIR=/tmp/u043/cache/M11 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M11.log 2>&1; proven_relations_test.go:59: /tmp/adamic-gate/TestProvenRelationsErase1993644486/002/main.a:1:105: Adamic 0.1 refuses an unproven relation from { readonly name: string; readonly extra: number; } to { readonly name: string; }: optional field  has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)",
    "subsumption_kills": 3
  },
  {
    "test": "TestReadinessElisionRequiresDominatingAssignment",
    "package": "internal/lower",
    "file": "internal/lower/readiness_test.go:10",
    "seconds": 0.354,
    "oracle": "Handwritten .a refusal text, one eager .ts assertion, and no delayed read checks in IR. Historical test name no longer describes its entire checker.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M03",
      "M07",
      "M10",
      "M13",
      "M14"
    ],
    "unique_kills": [
      "M13"
    ],
    "last_proven_fail": "M13: readiness_test.go:42: want one eager assertion, got ir.NonNullCheckCounts{Proven:0, Checked:2, Sites:[]ir.NonNullCheck{ir.NonNullCheck{Where:\"/tmp/adamic-gate/TestReadinessElisionRequiresDominatingAssignmentbranch-join3283925662/002/main.ts:1:50\", Expression:\"undefined!\", Proven:false}}}",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "PLower"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M13 ADAMIC_BUILD_CACHE_DIR=/tmp/u043/cache/M13 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M13.log 2>&1; readiness_test.go:42: want one eager assertion, got ir.NonNullCheckCounts{Proven:0, Checked:2, Sites:[]ir.NonNullCheck{ir.NonNullCheck{Where:\"/tmp/adamic-gate/TestReadinessElisionRequiresDominatingAssignmentbranch-join3283925662/002/main.ts:1:50\", Expression:\"undefined!\", Proven:false}}}"
  },
  {
    "test": "TestRegExpNativeRefusals",
    "package": "internal/lower",
    "file": "internal/lower/regexp_test.go:11",
    "seconds": 0.148,
    "oracle": "Handwritten acceptance of either NotYet or Refused, without checking the reason. M17 changes the first refusal reason while this row passes.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "PLower"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "All 19 production matrix runs passed this row; see matrix.json and M*.log."
  },
  {
    "test": "TestRegExpSourceNode",
    "package": "internal/lower",
    "file": "internal/lower/regexp_test.go:29",
    "seconds": 0.035,
    "oracle": "Node 24 computes RegExp.source for each input; lone-surrogate byte check is handwritten.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M18",
      "M19"
    ],
    "unique_kills": [
      "M18",
      "M19"
    ],
    "last_proven_fail": "M19: regexp_test.go:48: source differs: \"[/]\"/ got=\"[\\\\/]\" Node=\"[/]\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "PRegexSource"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M19 ADAMIC_BUILD_CACHE_DIR=/tmp/u043/cache/M19 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M19.log 2>&1; regexp_test.go:48: source differs: \"[/]\"/ got=\"[\\\\/]\" Node=\"[/]\""
  },
  {
    "test": "TestRepresentationClockSourceCheckedTypes",
    "package": "internal/lower",
    "file": "internal/lower/representation_clock_source_test.go:17",
    "seconds": 0.038,
    "oracle": "Handwritten representation kind/known pairs and checker-assignability preconditions.",
    "oracle_kind": "self",
    "kills": [
      "M09",
      "M10"
    ],
    "unique_kills": [],
    "last_proven_fail": "M10: representation_clock_source_test.go:69: callableSource: representation = (8, false), want (8, true)",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestPrototypeMethodsAreRefusedWithReasons"
    ],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "PRepresentation"
    ],
    "subsumer_seconds": 0.322,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M10 ADAMIC_BUILD_CACHE_DIR=/tmp/u043/cache/M10 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M10.log 2>&1; representation_clock_source_test.go:69: callableSource: representation = (8, false), want (8, true)",
    "subsumption_kills": 2,
    "vacuous_subcases": [
      "PRepresentation: source and lengthSource still return expected (0,false); objectSource, callableSource, arraySource and stringSource fail."
    ]
  },
  {
    "test": "TestSuppressionDirectivesAreRefused",
    "package": "internal/lower",
    "file": "internal/lower/suppression_directives_test.go:9",
    "seconds": 0.394,
    "oracle": "Handwritten Refused class, location, directive name, and repair text.",
    "oracle_kind": "self",
    "kills": [
      "M15",
      "M16"
    ],
    "unique_kills": [
      "M15",
      "M16"
    ],
    "last_proven_fail": "M16: suppression_directives_test.go:35: got /tmp/adamic-gate/TestSuppressionDirectivesAreRefused@ts-ignorenumber.a_@ts-i3400142785/001/main.a:1:1: Adamic 0.1 refuses @ts-ignore suppression directive; , want \"remove it and fix the type error\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "PLower"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M16 ADAMIC_BUILD_CACHE_DIR=/tmp/u043/cache/M16 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M16.log 2>&1; suppression_directives_test.go:35: got /tmp/adamic-gate/TestSuppressionDirectivesAreRefused@ts-ignorenumber.a_@ts-i3400142785/001/main.a:1:1: Adamic 0.1 refuses @ts-ignore suppression directive; , want \"remove it and fix the type error\""
  },
  {
    "test": "TestSuppressionDirectiveSoundNeighbors",
    "package": "internal/lower",
    "file": "internal/lower/suppression_directives_test.go:44",
    "seconds": 0.168,
    "oracle": "Handwritten absence of an error; it does not assert an IR answer exists.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M07"
    ],
    "unique_kills": [],
    "last_proven_fail": "M07: suppression_directives_test.go:56: sound neighbor: /tmp/adamic-gate/TestSuppressionDirectiveSoundNeighbors787477692/005/main.a:2:16: Adamic 0.1 refuses inherited library member n read as an own field; prototype members are not stored in an object's shape; call the method on its receiver, or wrap that call in an arrow (unbound-method)",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestDefaultTaggedInterfaceNeedsNoFlag"
    ],
    "mutants_in_matrix": 19,
    "probe_kills": [],
    "subsumer_seconds": 0.04,
    "vacuous": true,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M07 ADAMIC_BUILD_CACHE_DIR=/tmp/u043/cache/M07 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M07.log 2>&1; suppression_directives_test.go:56: sound neighbor: /tmp/adamic-gate/TestSuppressionDirectiveSoundNeighbors787477692/005/main.a:2:16: Adamic 0.1 refuses inherited library member n read as an own field; prototype members are not stored in an object's shape; call the method on its receiver, or wrap that call in an arrow (unbound-method) PLower-TestSuppressionDirectiveSoundNeighbors.log: passes with nil program and nil error.",
    "subsumption_kills": 2
  }
]
```

| ID | Starting file:line | Change | Failed rows |
| --- | --- | --- | --- |
| M01 | internal/lower/prototype.go:41 | invert inherited-library declaration recognition | TestUndecidedCycleReadsUseReadyChecks, TestParameterPropertyCallbackReceiver, TestParameterPropertyCheckerContracts, TestClosedFrameInputRejectsMutation, TestOptionalIndexingKeepsUnsupportedStorageNotYet, TestNamespaceTypesErase, TestParserFactoryBindingHoisting, TestViewObjectContractsAreAvailableToEraser, TestLiteralMethodViewsDoNotLoseThis, TestDefaultTaggedInterfaceNeedsNoFlag, TestReadonlyFieldsAreJudgedByTheirConstructorsWrites, TestNumericEnumsAreOpen, TestDefiniteAssignmentUsesReadiness, TestDefiniteAssignmentSoundNeighbors, TestTasteRepresentationLimitsStayExplicit, TestSuppressionDirectiveSoundNeighbors, TestProvenRelationsErase, TestIsPrototypeOfReadsExplainThePrototypeRefusal, TestPredicateOverloadRuntime, TestClassFeaturesAccessorRefusals, TestClassWrongOutputPrivateRepair, TestPrototypeMethodsAreRefusedWithReasons, TestPrototypeHazardsBehindObjectViewsAreNotYet, TestClassWrongOutputKeysRepair, TestInheritanceGenericFactoryLayouts, TestClassWrongOutputIteratorReceiver, TestInheritanceConditionalThisRules, TestInheritanceGenericMonomorphizations, TestInheritanceRefusesThisBeforeSuperReturns, TestInheritanceKeepsNominalTupleDestructuring, TestInheritanceAllowsSoundOverrides, TestClassFeaturesAccessorCaptureCycle, TestClassFeaturesNarrowedAccessor, TestClassFeaturesPrivateStorage, TestClassFeaturesStaticDeclarationsExecute, TestClassFeaturesReadonlyChecker, TestArgumentsLengthReadNeighbors, TestInheritedLibraryReadsNeverLoadOwnFields, TestOptionalWideningAllowed, TestTscNamespaceDeclarationShapes, TestAMethodReadAsAValueIsRefused, TestAViewThatCantWriteIsNotRefused, TestWhatZeroOneRefusesIsRefusedWithAFix, TestDefaultTaggedInterfaceAdmission, TestEnumInitializationReach, TestFlagEnumsDomain, TestReadinessElisionRequiresDominatingAssignment, TestCheckedCastProofAndElision |
| M02 | internal/lower/prototype.go:71 | select constructor instead of isPrototypeOf diagnostic | TestIsPrototypeOfReadsExplainThePrototypeRefusal |
| M03 | internal/lower/prototype.go:91 | invert prototype-chain call refusal | TestLibraryMethodValues, TestTypedArraysLower, TestPrototypeMethodsAreRefusedWithReasons, TestTypedArrayGaps, TestUnrepresentedPrototypeCallsAreNotYet, TestPrototypeHazardsBehindObjectViewsAreNotYet, TestEntriesRecordBoundaries, TestReadinessElisionRequiresDominatingAssignment, TestInheritanceNativeSignatureNeighbors |
| M04 supplemental probe | internal/lower/prototype.go:264 | discard whole-program hazard result | TestPrototypeHazardsBehindObjectViewsAreNotYet |
| M05 | internal/lower/prototype.go:151 | invert valueOf representation compatibility | TestUnrepresentedPrototypeCallsAreNotYet |
| M06 | internal/lower/prototype.go:110 | clear locale-sensitive number repair | TestPrototypeMethodsAreRefusedWithReasons |
| M07 | internal/lower/object.go:382 | invert property inherited guard | TestUndecidedCycleReadsUseReadyChecks, TestViewObjectContractsAreAvailableToEraser, TestTasteRepresentationLimitsStayExplicit, TestProvenRelationsErase, TestSuppressionDirectiveSoundNeighbors, TestPrototypeHazardsBehindObjectViewsAreNotYet, TestParameterPropertyCallbackReceiver, TestParameterPropertyCheckerContracts, TestLiteralMethodViewsDoNotLoseThis, TestDefaultTaggedInterfaceNeedsNoFlag, TestOptionalIndexingKeepsUnsupportedStorageNotYet, TestClosedFrameInputRejectsMutation, TestClassFeaturesAccessorRefusals, TestClassWrongOutputPrivateRepair, TestReadonlyFieldsAreJudgedByTheirConstructorsWrites, TestClassWrongOutputKeysRepair, TestClassFeaturesPrivateStorage, TestNamespaceTypesErase, TestClassWrongOutputIteratorReceiver, TestInheritanceGenericFactoryLayouts, TestClassFeaturesReadonlyChecker, TestInheritanceConditionalThisRules, TestInheritanceGenericMonomorphizations, TestInheritanceRefusesThisBeforeSuperReturns, TestInheritanceKeepsNominalTupleDestructuring, TestInheritanceAllowsSoundOverrides, TestClassFeaturesNarrowedAccessor, TestClassFeaturesAccessorCaptureCycle, TestClassFeaturesStaticDeclarationsExecute, TestNumericEnumsAreOpen, TestDefiniteAssignmentUsesReadiness, TestDefiniteAssignmentSoundNeighbors, TestArrayPredicateCoexistsWithUnknownReflection, TestArgumentsLengthReadNeighbors, TestPredicateOverloadRuntime, TestUnknownReflectionRefusals, TestReadinessElisionRequiresDominatingAssignment, TestInheritedLibraryReadsNeverLoadOwnFields, TestOptionalWideningAllowed, TestDefaultTaggedInterfaceAdmission, TestCheckedCastProofAndElision, TestEnumInitializationReach, TestAMethodReadAsAValueIsRefused, TestAViewThatCantWriteIsNotRefused, TestFlagEnumsDomain, TestWhatZeroOneRefusesIsRefusedWithAFix |
| M08 | internal/lower/object.go:1788 | invert element inherited guard | TestNumericEnumNeverProof, TestCensusSmallFiniteKeyRead, TestIsPrototypeOfReadsExplainThePrototypeRefusal, TestUnknownArrayPredicateRefusesUnrepresentedObservations, TestOptionalIndexingKeepsUnsupportedStorageNotYet, TestAViewThatCantWriteIsNotRefused, TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat, TestPredicateOverloadRuntime, TestInheritedLibraryReadsNeverLoadOwnFields |
| M09 | internal/lower/expression.go:83 | mark array representation unknown | TestNestedRestIsSupported, TestTypedArraysLower, TestRepresentationClockSourceCheckedTypes, TestPrototypeMethodsAreRefusedWithReasons, TestPrimitiveAdmittingSlotsUseBoxes, TestOptionalIndexingKeepsUnsupportedStorageNotYet, TestInheritanceRefusesGrowingGenericClasses, TestInheritanceGenericMonomorphizations, TestIteratorMapperIndexHasNumberRepresentation, TestLibraryMethodValueBoundaries, TestGenericJSONUnionArrayIsNotYet, TestGenericFunctionPolymorphicRecursionIsRefused, TestReadonlyFieldsAreJudgedByTheirConstructorsWrites, TestNumericEnumNeverProof, TestParserFactoryBindingHoisting, TestFlagEnumAliasBoundaries, TestNumericEnumsAreOpen, TestEmptyLiteralGenericReturnUsesSamePath, TestPredicateOverloadRuntime, TestArrayPredicatePreservesDeclaredElementContract, TestArgumentsLengthReadNeighbors, TestInheritedLibraryReadsNeverLoadOwnFields, TestClassArgumentsUseCheckerIdentity, TestInheritanceRefusesUnsoundOverrides, TestLibraryMapSetGapsStayRefused, TestEnumSlotViews, TestFlagEnumsDomain, TestFlagEnumsOpen, TestNestedEmptyArrayElementKinds, TestATupleSeenAsAnArrayIsNotYet, TestAMutableLocationSeenWiderIsRefused, TestWhatZeroOneRefusesIsRefusedWithAFix, TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat |
| M10 | internal/lower/expression.go:95 | mark callable representation unknown | TestNamespaceAmbientHostInitialization, TestLibraryMethodValueBoundaries, TestConstructorCacheOneArgumentRemainsNotYet, TestLibraryMethodValues, TestInheritanceConditionalThisRules, TestInheritanceRefusesThisBeforeSuperReturns, TestLiteralMethodViewsDoNotLoseThis, TestParameterPropertyCallbackReceiver, TestRepresentationClockSourceCheckedTypes, TestIsPrototypeOfReadsExplainThePrototypeRefusal, TestPrototypeMethodsAreRefusedWithReasons, TestPredicateOverloadCallback, TestPredicateBodiesAreProven, TestPredicateCallbackContracts, TestCensusMarkerZeroCallIsNotAssumedSafe, TestArgumentsLengthReadNeighbors, TestNestedCapturedParametersAreOwned, TestParameterPropertiesSoundness, TestOptionalIndexingMapShapeRefused, TestNestedRestIsSupported, TestNestedRebindingNotYet, TestNestedCallbackCycleIsRefused, TestClosedFrameInputRejectsMutation, TestNestedEnvironmentCycleIncludesDisjointSlots, TestNestedEnvironmentHasOneAllocationSite, TestNestedFunctionCycleIsRefused, TestNestedFunctionGapsAreLoud, TestParserFactoryBindingHoisting, TestATupleSeenAsAnArrayIsNotYet, TestWhatZeroOneRefusesIsRefusedWithAFix, TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat, TestLibraryStringRefusals, TestPredicateOverloadRuntime, TestClassArgumentsUseCheckerIdentity, TestLibraryLanguageBoundaries, TestIteratorDescriptorReasons, TestDefaultTaggedInterfaceAdmission, TestUnknownReflectionRefusals, TestReadinessElisionRequiresDominatingAssignment, TestInheritedLibraryReadsNeverLoadOwnFields, TestEnumSlotViews, TestCensusOverloadReturnProof, TestEnumInitializationReach |
| M11 | internal/lower/proven_relations.go:22 | invert optional relation failure | TestDefaultTaggedInterfaceNeedsNoFlag, TestNumericEnumsAreOpen, TestEmptyLiteralGenericReturnUsesSamePath, TestTypedArrayViewsCannotChangeRepresentation, TestProvenRelationsErase, TestDefaultTaggedInterfaceAdmission, TestProvenRelationsRefuse, TestATupleSeenAsAnArrayIsNotYet, TestCheckedCastProofAndElision, TestUncheckableCastsStayRefused, TestOptionalWideningRefused |
| M12 | internal/lower/proven_relations.go:25 | drop nominal relation refusal block | TestCheckedCastProofAndElision, TestUncheckableCastsStayRefused, TestProvenRelationsRefuse |
| M13 | internal/lower/non_null.go:114 | double eager assertion count | TestReadinessElisionRequiresDominatingAssignment |
| M14 | internal/lower/refusals.go:22 | clear non-null repair | TestAdamicNullishAssertionsAreRefused, TestWhatZeroOneRefusesIsRefusedWithAFix, TestReadinessElisionRequiresDominatingAssignment |
| M15 | internal/lower/refusals.go:42 | require two suppression directives | TestSuppressionDirectivesAreRefused |
| M16 | internal/lower/refusals.go:49 | clear suppression repair | TestSuppressionDirectivesAreRefused |
| M17 | internal/lower/regexp.go:32 | reject any RegExp constructor argument | TestPrototypeHazardsBehindObjectViewsAreNotYet |
| M18 | internal/lower/regexp.go:222 | change empty RegExp source | TestRegExpSourceNode |
| M19 | internal/lower/regexp.go:264 | invert slash character-class condition | TestRegExpSourceNode |
| M20 | internal/lower/regexp.go:225 | use u for nested character sets |  |

Survivors: M20

The CUT is Adamic lowering: inherited-member/property/element guards, prototype dispatch and hazard detection, representation mapping, relation proofs, eager assertion accounting, directive refusal, and RegExp source escaping. The oracle is Node for prototype membership and RegExp sources, and handwritten expected error classes, diagnostic text, IR facts, and own-artifact equality elsewhere. The reached-function inventory was generated before planting mutants from slice coverage. Every production standalone diff was vetted separately; the source switch was built once and uses per-mutant ADAMIC_BUILD_CACHE_DIR directories. No .a port or external checker was mutated.

The brief cites 8de93800f4 but requires fresh origin/main. This audit uses 5deb11d433947539ab6448558b8ab1f1904393a5, and all file:line references and diffs use that commit. None of the 14 rows moved or vanished. Their bodies have different assertions and are separate rows, not input-only wrappers sharing one family checker. No requested witness, setup-only check, or subprocess helper was found.

The whole baseline fit under 90 seconds, so production matrices were not narrowed just because internal/lower is named as a big package. Package-unique kills require one failed top-level row in a complete whole-package run. Kills outside these runs and repo-wide uniqueness remain unmeasured. Out-of-slice subsumers were timed when needed; choosing a subsumer is a small-matrix hint, not a deletion recommendation.

M04 was in the initial fixed change-constant menu, but its unconditional empty hazard answer is reported as a supplemental empty-answer probe. It is excluded from production kills, unique kills, and verdicts. Direct-entry probes are separate P*.diff files: Lower, property, elementAccess, representation, and escapeRegexSource. They ran callers separately to prevent a nil-program panic from hiding later rows. Probe panics are observed failures. No probe was planted in load.Load, because it is preparation for the lowering rows; the nullish row itself reaches only that checker path and has vacuity null.

The nullish row cannot be judged by a lowering mutant. Its time and baseline behavior are measured, but no mutation in the external checker was performed. The suppression sound-neighbor row accepts nil program/nil error, demonstrating vacuity. Representation's two unresolved-parameter checks pass its empty probe while its four positive representations fail. Most refusal rows use our own expected text or error classes. RegExpNativeRefusals accepts either Refused or NotYet without pinning the reason; the M17 witness demonstrates an alternate refusal passing this row. Erasure compares two outputs of the same lowerer, so shared wrong artifacts can agree.

Baseline skips outside the requested rows were TestOriginalCycleLedger (requires pinned pristine TypeScript 6.0.3 with generated diagnostics), TestOptionalWideningCensus (requires a caller-selected project config and output path), and one TestMixedUnionContractGraph subcase. No requested row skipped. These are corpus/inventory opt-ins rather than missing tools for requested rows; their corpus runs remain uncovered. stage3/api npm ci ran first and reported three packages in 357 ms. Tools were warm and cloud/setup.sh was not run. Fetch/list/baseline compilation wall timings were not separately captured.

The initial switch instrumentation had a syntax error in a guarded dropped block. This was corrected before the first matrix; initial-switch-error.log preserves that failed build preparation. The switch also selects M02's constant exactly as its standalone diff specifies. Production source is restored before final clean package run and vet. Evidence .diff files naturally contain whitespace-bearing context; whitespace diagnostics on them as newly added text are distinct from the restored source diff check.

The audit does not measure mutation adequacy for the entire compiler, error-message authority against ECMAScript, or repo-wide uniqueness. No other package's suite was run. The 20-change cap and supplemental-probe exclusion leave 19 production mutants, below three per row. Untrue means no fixed-menu production mutant made that row fail in this session, not that no possible mutant can ever fail it. Expected values without checked outside authority are classified self.

Measured timing: switch build 8.250688137999987 seconds; standalone vet 2.517 seconds; matrix/probe command wall total 591.308 seconds; clean timings/coverage command wall total 110.548 seconds. Individual rebuild/run timings are in mutant-runs.json; all three clean timings per requested row are retained.

Standalone probe diffs were also vetted individually. Their local boolean is always true and exists to retain the original body without static unreachable-code warnings. The runtime-switch probe results implement the same entry return. Probe vet timings are saved in probe-vet-runs.json.

M20 survivor witness: clean lowering and Node produce "[[/]\\/" for input "[[/]/" with u; mutant lowering produces "[[/]/". witness-clean-source.log and witness-M20.log show the exact observations and successful Node construction. This is unguarded behavior rather than an equivalent candidate. M17 witness: the first native-refusal case changes from "RegExp with a nonconstant pattern" to "RegExp with more than two arguments", while TestRegExpNativeRefusals passes its matrix run. witness-clean-refusal.log and witness-M17.log preserve the observations.

Final clean package run, final go vet, and restored source diff check passed; finish-runs.json records their command wall times. The soft 20-minute budget was exceeded to complete whole-package matrices, direct-entry probes, three timings per row, survivor witnesses, and outside-subsumer timings. No matrix exceeded its 90-second binary budget.
