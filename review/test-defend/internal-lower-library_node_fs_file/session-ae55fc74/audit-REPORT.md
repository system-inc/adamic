Audited u035 at e2492670b06a4dce1deafe837158ce0366cf2bc4.
15 named tests grouped into 12 rows, including one four-member acceptance family.
Verdicts: 3 sacred, 7 subsumed, 1 overlapping, 1 untrue within this fixed mutant set.
20 production mutants, 4 confirmed survivors, 2 empty-answer probes; nproc=5.
Clean baseline passed; evidence branch test-audit/internal-lower-library_node_fs_file.

```json
[
  {
    "test": "TestNodeFSFile family",
    "package": "internal/lower",
    "file": "internal/lower/library_node_fs_file_test.go",
    "seconds": 0.97,
    "oracle": "Handwritten err == nil acceptance only. Four source-input wrappers share lowerSource and the same assertion; IR contents are not checked.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M02",
      "M03",
      "M04",
      "M05",
      "M06",
      "M08",
      "M10",
      "M17",
      "M18",
      "M20"
    ],
    "unique_kills": [
      "M01",
      "M02",
      "M06",
      "M08",
      "M17",
      "M18",
      "M20"
    ],
    "last_proven_fail": "M20: library_node_fs_file_test.go:93: /tmp/adamic-gate/TestNodeFSFileBufferBorrow1925176034/001/main.a:1:115: stage 0 can't lower node:fs.readSync: Buffer or Hash viewed as another object type (native host internal slots) yet",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M20 ADAMIC_BUILD_CACHE_DIR=/tmp/u035/cache/M20 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M20.log 2>&1; library_node_fs_file_test.go:93: /tmp/adamic-gate/TestNodeFSFileBufferBorrow1925176034/001/main.a:1:115: stage 0 can't lower node:fs.readSync: Buffer or Hash viewed as another object type (native host internal slots) yet",
    "members": [
      "TestNodeFSFileOptionsBorrow",
      "TestNodeFSFileNamespaceImport",
      "TestNodeFSFileBufferBorrow",
      "TestNodeFSFileScratchOptionsBorrow"
    ]
  },
  {
    "test": "TestNodeFSFileDoesNotAuthorizeMutableWidening",
    "package": "internal/lower",
    "file": "internal/lower/library_node_fs_file_test.go",
    "seconds": 0.349,
    "oracle": "Handwritten assertion: mutable widening refusal contains invariant-mutable",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M14"
    ],
    "unique_kills": [],
    "last_proven_fail": "M14: library_node_fs_file_test.go:25: want mutable widening refusal, got <nil>",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestEmptyNeverMapCannotGainWritableInhabitants"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 0.116,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M14 ADAMIC_BUILD_CACHE_DIR=/tmp/u035/cache/M14 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M14.log 2>&1; library_node_fs_file_test.go:25: want mutable widening refusal, got <nil>"
  },
  {
    "test": "TestNodeFSFileRefusesOptionEffects",
    "package": "internal/lower",
    "file": "internal/lower/library_node_fs_file_test.go",
    "seconds": 0.504,
    "oracle": "Handwritten assertion: option-effects refusal contains evaluated expressions",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M04",
      "M05",
      "M07"
    ],
    "unique_kills": [
      "M07"
    ],
    "last_proven_fail": "M07: library_node_fs_file_test.go:33: want effects refusal, got /tmp/adamic-gate/TestNodeFSFileRefusesOptionEffects1815665381/001/main.a:1:112: stage 0 can't lower statSync: unsupported options yet",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M07 ADAMIC_BUILD_CACHE_DIR=/tmp/u035/cache/M07 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M07.log 2>&1; library_node_fs_file_test.go:33: want effects refusal, got /tmp/adamic-gate/TestNodeFSFileRefusesOptionEffects1815665381/001/main.a:1:112: stage 0 can't lower statSync: unsupported options yet"
  },
  {
    "test": "TestNodeFSFileRefusesVoidValues",
    "package": "internal/lower",
    "file": "internal/lower/library_node_fs_file_test.go",
    "seconds": 0.569,
    "oracle": "Handwritten assertion: void-value refusal contains fs void calls used as values",
    "oracle_kind": "self",
    "kills": [
      "M05"
    ],
    "unique_kills": [],
    "last_proven_fail": "M05: library_node_fs_file_test.go:41: want void value refusal, got /tmp/adamic-gate/TestNodeFSFileRefusesVoidValues3798862361/001/main.a:1:49: stage 0 can't lower node:fs.closeSync yet",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestNodeFSFileRefusesOptionEffects"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 0.504,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M05 ADAMIC_BUILD_CACHE_DIR=/tmp/u035/cache/M05 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M05.log 2>&1; library_node_fs_file_test.go:41: want void value refusal, got /tmp/adamic-gate/TestNodeFSFileRefusesVoidValues3798862361/001/main.a:1:49: stage 0 can't lower node:fs.closeSync yet"
  },
  {
    "test": "TestNodeFSFileRefusesPinnedUnsupportedOverloads",
    "package": "internal/lower",
    "file": "internal/lower/library_node_fs_file_test.go",
    "seconds": 1.487,
    "oracle": "Handwritten assertion: NotYet type and member substring only; M09 demonstrates rejection reason can be wrong",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M10"
    ],
    "unique_kills": [],
    "last_proven_fail": "M10: library_node_fs_file_test.go:59: want named NotYet statSync, got <nil>",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestNodeFSFile family"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 0.97,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M10 ADAMIC_BUILD_CACHE_DIR=/tmp/u035/cache/M10 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M10.log 2>&1; library_node_fs_file_test.go:59: want named NotYet statSync, got <nil>"
  },
  {
    "test": "TestNodeFSFileQualifiedErrorType",
    "package": "internal/lower",
    "file": "internal/lower/library_node_fs_file_test.go",
    "seconds": 0.531,
    "oracle": "Handwritten assertion: Refused type plus no-optional-widening substring",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M04",
      "M05"
    ],
    "unique_kills": [],
    "last_proven_fail": "M05: library_node_fs_file_test.go:77: want optional host fields proven before the qualified cast, got /tmp/adamic-gate/TestNodeFSFileQualifiedErrorType497201411/001/main.a:1:80: stage 0 can't lower node:globals.ErrnoException.code yet",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestNodeFSFileRefusesOptionEffects"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 0.504,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M05 ADAMIC_BUILD_CACHE_DIR=/tmp/u035/cache/M05 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M05.log 2>&1; library_node_fs_file_test.go:77: want optional host fields proven before the qualified cast, got /tmp/adamic-gate/TestNodeFSFileQualifiedErrorType497201411/001/main.a:1:80: stage 0 can't lower node:globals.ErrnoException.code yet"
  },
  {
    "test": "TestNodeFSFileKeepsDetachedMethodRefusal",
    "package": "internal/lower",
    "file": "internal/lower/library_node_fs_file_test.go",
    "seconds": 0.509,
    "oracle": "Handwritten assertion: Refused type and isFile member substring",
    "oracle_kind": "self",
    "kills": [
      "M04",
      "M05",
      "M16"
    ],
    "unique_kills": [],
    "last_proven_fail": "M16: library_node_fs_file_test.go:86: want existing unbound-method refusal naming isFile, got /tmp/adamic-gate/TestNodeFSFileKeepsDetachedMethodRefusal864325006/001/main.a:1:74: Adamic 0.1 refuses a method read as a value (method would lose its object, and this with it); call it in an arrow that keeps the object: () => stat.isFile() (unbound-method)",
    "verdict": "overlapping",
    "subsumed_by": [
      "TestNodeFSFile family",
      "TestAMethodReadAsAValueIsRefused"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M16 ADAMIC_BUILD_CACHE_DIR=/tmp/u035/cache/M16 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M16.log 2>&1; library_node_fs_file_test.go:86: want existing unbound-method refusal naming isFile, got /tmp/adamic-gate/TestNodeFSFileKeepsDetachedMethodRefusal864325006/001/main.a:1:74: Adamic 0.1 refuses a method read as a value (method would lose its object, and this with it); call it in an arrow that keeps the object: () => stat.isFile() (unbound-method)"
  },
  {
    "test": "TestNodeFSFileScratchOverloadsAreNamed",
    "package": "internal/lower",
    "file": "internal/lower/library_node_fs_file_test.go",
    "seconds": 1.144,
    "oracle": "Handwritten assertion: NotYet type and overload member substring",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M04",
      "M05",
      "M19"
    ],
    "unique_kills": [
      "M19"
    ],
    "last_proven_fail": "M19: library_node_fs_file_test.go:119: rmSync: want named NotYet, got <nil>",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M19 ADAMIC_BUILD_CACHE_DIR=/tmp/u035/cache/M19 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M19.log 2>&1; library_node_fs_file_test.go:119: rmSync: want named NotYet, got <nil>"
  },
  {
    "test": "TestNodeLibraryNamesUnimplementedMembers",
    "package": "internal/lower",
    "file": "internal/lower/library_node_test.go",
    "seconds": 2.182,
    "oracle": "Handwritten assertion: NotYet type and qualified declaration name substring",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M04",
      "M05"
    ],
    "unique_kills": [],
    "last_proven_fail": "M05: library_node_test.go:31: want named NotYet node:fs.copyFileSync, got /tmp/adamic-gate/TestNodeLibraryNamesUnimplementedMembersnodefs.copyFileSync363952989/001/main.a:1:47: stage 0 can't lower reading copy yet",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestNodeFSFileRefusesOptionEffects"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 0.504,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M05 ADAMIC_BUILD_CACHE_DIR=/tmp/u035/cache/M05 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M05.log 2>&1; library_node_test.go:31: want named NotYet node:fs.copyFileSync, got /tmp/adamic-gate/TestNodeLibraryNamesUnimplementedMembersnodefs.copyFileSync363952989/001/main.a:1:47: stage 0 can't lower reading copy yet"
  },
  {
    "test": "TestNodeLibraryDoesNotRefuseTypeOnlyOrUserNames",
    "package": "internal/lower",
    "file": "internal/lower/library_node_test.go",
    "seconds": 0.952,
    "oracle": "Handwritten assertion: type-only imports and user names accepted; only checks err == nil",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M04",
      "M05"
    ],
    "unique_kills": [],
    "last_proven_fail": "M05: library_node_test.go:44: /tmp/adamic-gate/TestNodeLibraryDoesNotRefuseTypeOnlyOrUserNames2037322176/001/main.a:1:67: stage 0 can't lower node:console.Console.log yet",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestNodeFSFileRefusesOptionEffects"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [],
    "subsumer_seconds": 0.504,
    "vacuous": true,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M05 ADAMIC_BUILD_CACHE_DIR=/tmp/u035/cache/M05 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M05.log 2>&1; library_node_test.go:44: /tmp/adamic-gate/TestNodeLibraryDoesNotRefuseTypeOnlyOrUserNames2037322176/001/main.a:1:67: stage 0 can't lower node:console.Console.log yet"
  },
  {
    "test": "TestNodeLibraryDistinguishesReceiverOwners",
    "package": "internal/lower",
    "file": "internal/lower/library_node_test.go",
    "seconds": 0.328,
    "oracle": "Handwritten assertion: handwritten two distinct member names from AST symbol owners",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M04"
    ],
    "unique_kills": [],
    "last_proven_fail": "M04: library_node_test.go:75: missing distinct member node:fs.StatsBase.isFile in map[node:fs.isFile:true]",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestNodeFSFileRefusesOptionEffects"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P02"
    ],
    "subsumer_seconds": 0.504,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M04 ADAMIC_BUILD_CACHE_DIR=/tmp/u035/cache/M04 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M04.log 2>&1; library_node_test.go:75: missing distinct member node:fs.StatsBase.isFile in map[node:fs.isFile:true]"
  },
  {
    "test": "TestNodeLibraryQualifiedTypeAssertion",
    "package": "internal/lower",
    "file": "internal/lower/library_node_test.go",
    "seconds": 0.623,
    "oracle": "Handwritten assertion: Refused type plus no-optional-widening substring; fixed mutant plan did not isolate provenRelation optional-field branch",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=P01 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run ^TestNodeLibraryQualifiedTypeAssertion$ > P01-TestNodeLibraryQualifiedTypeAssertion.log 2>&1; library_node_test.go:85: want the qualified name checked without inventing optional fields, got <nil>"
  }
]
```

CODE UNDER TEST: Go lowering in internal/lower, principally Lower, Node symbol naming, fs overload recognition and relation/refusal checking. ORACLE: handwritten acceptance/diagnostic/member-name assertions in the two unchanged test files. All are self. TypeScript checking is preparation, not an independently executed expected-answer oracle.

All 276 covered production functions are listed with starting-commit source locations in reached-functions.txt. Reach was measured before mutations with the unit coverage run. The fixed code-derived plan is plan.json; standalone diff MNN.diff contains each change without selector scaffolding. The compile-once scratch source is preserved in selector.diff.

| ID | Starting-commit file:line | Change | Failed rows |
|---|---|---|---|
| M01 | internal/lower/library_node_fs_file.go:466 | index := 1 -> index := 0 | TestNodeFSFile family |
| M02 | internal/lower/library_node_fs_buffer.go:34 | len(call.Arguments.Nodes) != 5 -> len(call.Arguments.Nodes) != 4 | TestNodeFSFile family |
| M03 | internal/lower/library_node.go:35 | !load.IsNodeLibrary(file) -> load.IsNodeLibrary(file) | TestAMethodReadAsAValueIsRefused, TestAMutableLocationSeenWiderIsRefused, TestATupleSeenAsAnArrayIsNotYet, TestAbfe962OverrideDefaultAddedRefusesNativeSignature, TestAdamicNullishAssertionsAreRefused, TestArgumentsLengthReadNeighbors, TestArgumentsLengthRefusalFixtures, TestArgumentsLengthRefusals, TestArrayPredicateCoexistsWithUnknownReflection, TestArrayPredicateDoesNotMisclassifyNativeTuples, TestArrayPredicatePreservesDeclaredElementContract, TestCallableNamespaceLimitsStayLoud, TestCallableNamespaceReceiverStaysLoud, TestCensusBooleanDeadBranch, TestCensusMarkerResultIsAssignable, TestCensusMarkerZeroCallIsNotAssumedSafe, TestCensusOverloadBinderGuards, TestCensusOverloadRelation, TestCensusOverloadReturnProof, TestCensusRestMutableElements, TestCensusSmallFiniteKeyRead, TestCheckPragmaNeighborsCompile, TestCheckedCastProofAndElision, TestClassArgumentsUseCheckerIdentity, TestClassFeaturesAccessorCaptureCycle, TestClassFeaturesAccessorRefusals, TestClassFeaturesNarrowedAccessor, TestClassFeaturesPrivateStorage, TestClassFeaturesReadonlyChecker, TestClassFeaturesStaticDeclarationsExecute, TestClassFeaturesStaticInterfaceCycle, TestClassFeaturesStaticParentCycle, TestClassWrongOutputIteratorReceiver, TestClassWrongOutputKeysRepair, TestClassWrongOutputPrivateRepair, TestClockGenericReturnsT01Shapes, TestClosedFrameInputRejectsMutation, TestConditionAssertionAdmission, TestConsoleLowersToWriteLine, TestConstEnumErasesRuntimeObject, TestConstructorCacheOneArgumentRemainsNotYet, TestDebugNamespaceMergedCapabilities, TestDefaultTaggedInterfaceAdmission, TestDefaultTaggedInterfaceNeedsNoFlag, TestDefiniteAssignmentSoundNeighbors, TestDefiniteAssignmentUsesReadiness, TestEmptyLiteralGenericReturnUsesSamePath, TestEmptyNeverMapCannotGainWritableInhabitants, TestEntriesAllocationProof, TestEntriesRecordBoundaries, TestEnumIdentityAcrossModules, TestEnumInitializationReach, TestEnumNameEnumeration, TestEnumNeverDefault, TestEnumSlotViews, TestEnumSwitchExhaustiveness, TestFlagEnumAliasBoundaries, TestFlagEnumInlineIteration, TestFlagEnumLiteralSpellings, TestFlagEnumMemberAliases, TestFlagEnumsDomain, TestFlagEnumsOpen, TestGenericFunctionPolymorphicRecursionIsRefused, TestGenericIteratorViewsPreserveNativeArguments, TestGenericJSONUnionArrayIsNotYet, TestGenericUnionFixtureHasSeparateInstances, TestInheritanceAllowsSoundOverrides, TestInheritanceConditionalThisRules, TestInheritanceGenericFactoryLayouts, TestInheritanceGenericMonomorphizations, TestInheritanceGenericNominalConstraints, TestInheritanceGenericSoundness, TestInheritanceGenericViewsKeepNominalArguments, TestInheritanceHasClassIdentity, TestInheritanceKeepsNominalTupleDestructuring, TestInheritanceNativeSignatureLimits, TestInheritanceNativeSignatureNeighbors, TestInheritanceRefusesFalseNominalViews, TestInheritanceRefusesGrowingGenericClasses, TestInheritanceRefusesThisBeforeSuperReturns, TestInheritanceRefusesUnsoundOverrides, TestInheritanceRejectsUnsupportedConstructorShapes, TestInputSpreadArgumentsAreNotYet, TestInputSpreadCoverage, TestIsPrototypeOfReadsExplainThePrototypeRefusal, TestIteratorDescriptorReasons, TestIteratorDestructuringDoesNotLieAboutExhaustion, TestIteratorMapperIndexHasNumberRepresentation, TestIteratorSymbolKeysAreNotStringKeys, TestIteratorViewsCannotEraseReceivers, TestIteratorViewsCannotHideReturn, TestLibraryLanguageBoundaries, TestLibraryMapSetGapsStayRefused, TestLibraryMapSetIteratorCopyTypesRefused, TestLibraryMethodValueBoundaries, TestLibraryMethodValues, TestLibraryStringRefusals, TestLiteralMethodCapturesCannotMakeCycles, TestLiteralMethodViewsDoNotLoseThis, TestModuleNamespaceInitializedReadProof, TestModuleNamespaceLimitsStayLoud, TestNamespaceAmbientHostInitialization, TestNamespaceLimitsStayLoud, TestNamespaceReturnedAssignmentLimits, TestNamespaceTypesErase, TestNestedCallbackCycleIsRefused, TestNestedCapturedParametersAreOwned, TestNestedEnvironmentCycleIncludesDisjointSlots, TestNestedEnvironmentHasOneAllocationSite, TestNestedFunctionCycleIsRefused, TestNestedFunctionGapsAreLoud, TestNestedRestIsSupported, TestNodeBufferRefusals, TestNodeFSFile family, TestNodeFSFileDoesNotAuthorizeMutableWidening, TestNodeFSFileQualifiedErrorType, TestNodeFSFileRefusesOptionEffects, TestNodeFSFileRefusesPinnedUnsupportedOverloads, TestNodeFSFileScratchOverloadsAreNamed, TestNodeLibraryDistinguishesReceiverOwners, TestNodeLibraryDoesNotRefuseTypeOnlyOrUserNames, TestNodeLibraryNamesUnimplementedMembers, TestNonNullAssertionLowersToNullishPanic, TestNumericEnumLiteralPromises, TestNumericEnumNeverProof, TestNumericEnumsAreOpen, TestObjectRefusalsExplainSoundness, TestOptionalFunctionValueRelation, TestOptionalIndexingKeepsUnsupportedStorageNotYet, TestOptionalIndexingMapShapeRefused, TestOptionalWideningAllowed, TestOptionalWideningRefused, TestOptionalWideningSpreadOverwrite, TestOverloadedShorthandFunctionValueStaysNotYet, TestOverrideOptionalNumberRefusesNativeSignature, TestOverrideOptionalRefusesNativeSignature, TestParameterPropertiesSoundness, TestParameterPropertyCallbackReceiver, TestParameterPropertyCheckerContracts, TestPredicateBodiesAreProven, TestPredicateCallbackContracts, TestPredicateOverloadCallback, TestPredicateOverloadRuntime, TestPrimitiveAdmittingSlotsUseBoxes, TestPrototypeHazardsBehindObjectViewsAreNotYet, TestPrototypeMethodsAreRefusedWithReasons, TestProvenRelationsErase, TestProvenRelationsRefuse, TestReadinessElisionRequiresDominatingAssignment, TestReadonlyFieldsAreJudgedByTheirConstructorsWrites, TestSuppressionDirectiveSoundNeighbors, TestTasteRepresentationLimitsStayExplicit, TestTracingNamespaceEscapeStaysNotYet, TestTscNamespaceDeclarationShapes, TestTypedArrayGaps, TestTypedArraysLower, TestUncheckableCastsStayRefused, TestUndecidedCycleReadsUseReadyChecks, TestUnknownReflectionRefusals, TestViewObjectContractsAreAvailableToEraser, TestViewObjectWritesNeedSourceCertificate, TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat, TestWhatZeroOneRefusesIsRefusedWithAFix |
| M04 | internal/lower/library_node.go:45 | owners = append([]string{parent.Name().Text()}, owners...) -> drop statement | TestNodeBufferRefusals, TestNodeFSFile family, TestNodeFSFileKeepsDetachedMethodRefusal, TestNodeFSFileQualifiedErrorType, TestNodeFSFileRefusesOptionEffects, TestNodeFSFileScratchOverloadsAreNamed, TestNodeLibraryDistinguishesReceiverOwners, TestNodeLibraryDoesNotRefuseTypeOnlyOrUserNames, TestNodeLibraryNamesUnimplementedMembers |
| M05 | internal/lower/library_node.go:68 | !implementedNodeMembers[name] -> implementedNodeMembers[name] | TestNamespaceAmbientHostInitialization, TestNodeBufferRefusals, TestNodeFSFile family, TestNodeFSFileKeepsDetachedMethodRefusal, TestNodeFSFileQualifiedErrorType, TestNodeFSFileRefusesOptionEffects, TestNodeFSFileRefusesVoidValues, TestNodeFSFileScratchOverloadsAreNamed, TestNodeLibraryDoesNotRefuseTypeOnlyOrUserNames, TestNodeLibraryNamesUnimplementedMembers |
| M06 | internal/lower/library_node_fs_file.go:101 | !discarded && !returned -> !discarded \|\| !returned | TestNodeFSFile family |
| M07 | internal/lower/library_node_fs_file.go:171 | "fs option literals containing evaluated expressions; bind a plain options object first" -> "unsupported options" | TestNodeFSFileRefusesOptionEffects |
| M08 | internal/lower/library_node_fs_file.go:192 | l.result.Strings[text.Index] == "utf8" -> l.result.Strings[text.Index] == "ascii" | TestNodeFSFile family |
| M09 | internal/lower/library_node_fs_file.go:270 | args[1].Type() != ir.String -> args[1].Type() == ir.String |  |
| M10 | internal/lower/library_node_fs_file.go:358 | !ok \|\| b.Value -> !ok \|\| !b.Value | TestNodeFSFile family, TestNodeFSFileRefusesPinnedUnsupportedOverloads |
| M11 | internal/lower/library_node_fs_file.go:230 | fallback = number(100) -> fallback = number(101) |  |
| M12 | internal/lower/library_node_fs_file.go:81 | operation, of = "exists", ir.Boolean -> operation, of = "stat", ir.Boolean |  |
| M13 | internal/lower/library_node_fs_file.go:348 | throws := ir.Expression(ir.BooleanConstant{Value: true}) -> throws := ir.Expression(ir.BooleanConstant{Value: false}) |  |
| M14 | internal/lower/invariance.go:47 | from == to \|\| visited[[2]*checker.Type{from, to}] -> from != to \|\| visited[[2]*checker.Type{from, to}] | TestAMutableLocationSeenWiderIsRefused, TestCensusMarkerResultIsAssignable, TestCensusOverloadRelation, TestCensusRestMutableElements, TestEmptyNeverMapCannotGainWritableInhabitants, TestEnumIdentityAcrossModules, TestEnumSlotViews, TestFlagEnumAliasBoundaries, TestFlagEnumsOpen, TestInheritanceRefusesUnsoundOverrides, TestLibraryStringRefusals, TestNodeFSFileDoesNotAuthorizeMutableWidening, TestNumericEnumLiteralPromises, TestOptionalFunctionValueRelation, TestParameterPropertiesSoundness, TestProvenRelationsRefuse, TestUncheckableCastsStayRefused |
| M15 | internal/lower/optional_widening.go:196 | "declare " + found.property + " on the source type, or build a fresh object with known fields (adamic/no-optional-widening)" -> "declare " + found.property | TestOptionalWideningRefused |
| M16 | internal/lower/refusals.go:205 | "a method read as a value (" + symbol.Name -> "a method read as a value (" + "method" | TestAMethodReadAsAValueIsRefused, TestNodeFSFileKeepsDetachedMethodRefusal, TestWhatZeroOneRefusesIsRefusedWithAFix |
| M17 | internal/lower/library_node_fs_file.go:474 | len(call.Arguments.Nodes) > index && -> len(call.Arguments.Nodes) > index+1 && | TestNodeFSFile family |
| M18 | internal/lower/library_node_fs_file.go:70 | len(call.Arguments.Nodes) != 5 -> len(call.Arguments.Nodes) != 4 | TestNodeFSFile family |
| M19 | internal/lower/library_node_fs_file.go:66 | operation = "rm" -> operation = "unlink" | TestNodeFSFileScratchOverloadsAreNamed |
| M20 | internal/lower/library_node_fs_buffer.go:42 | load.IsNodeLibrary(file) && -> !load.IsNodeLibrary(file) && | TestNodeFSFile family |

Survivors:

M09: openSync('x','r') lowers successfully before; mutant refuses it as numeric fs open flags. Witness logs and source are saved.
M11: rmSync('x',{retryDelay:100}) lowers successfully before; mutant refuses it as outside driver defaults. Witness logs and source are saved.
M12: existsSync('x') has Main.Value.Operation exists before; stat afterward, with a new throw argument. Witness logs save both IRs.
M13: statSync('x') has its default throw argument true before and false afterward. Witness logs save both IRs.

Probes: P01 returns nil,nil at Lower entry. P02 returns an empty string at nodeLibraryMember entry. Neither is counted in production kills. P01 panicked in an unrelated whole-package IR-inspection test; all assigned tests were independently probed. Five positive tests passed P01, grouped into two vacuous rows. All remaining assigned tests failed their own probe. No positive/negative mixed subcase vacuity was observed. Probe standalone diffs were independently vetted.

Brief ambiguities, mistakes and costs:

- The expanded family rule was ambiguous for inline assertions. I initially kept 15 rows and asked for clarification. With no answer, I applied its input-only criterion to the four NodeFSFile acceptance wrappers, grouped them as one family, and measured that family three times. Its explicit members are in results.json and families.json. Other rows assert different diagnostic types, text, or names.
- I mistakenly overlapped initial timing measurements with standalone mutation validation. One initial timing run failed to build. All 45 individual timings were repeated against fixed source with the selector inactive. Contaminated logs are retained separately and excluded from verdicts.
- The 20-mutant cap is below three mutants for each of the 12 grouped rows. The plan was fixed before matrix outcomes, rather than adjusted toward uncaught tests.
- The qualified-type-assertion row caught no production mutant. The plan challenged mutable relations and assignment-site optional widening, but did not isolate provenRelation optional-field checking. Its untrue verdict is limited to these 20 mutants. P01 made it fail, proving that this observation is not a claim that the test can never fail.
- A name-only NotYet oracle can accept the wrong rejection reason. M09 survived the pinned-unsupported-overload row, while its separate valid-string-open witness demonstrably changed behavior.
- Accepted compilation is weaker than inspecting IR. The four-member acceptance family passes an empty Lower result and misses the M12/M13 IR changes, despite catching seven package-unique mutants.
- No production whole-package run cooked. P01 panicked; per-row probe runs supplied complete assigned-row observations. Probe effects on later unrelated tests remain unknown.
- The two unrelated top-level tests TestOriginalCycleLedger and TestOptionalWideningCensus skipped. The former requires a pristine pinned TypeScript corpus with generated diagnostics; the latter requires an explicitly selected project/output path. Neither is a requested row. An unrelated MixedUnionContractGraph subcase also explicitly skips an unimplemented compiler feature. No requested test skipped.
- Individual timing runs and matrix runs overlapped, so the machine was not idle. Each timed Go test invocation selected only that row; timing is the binary printed ok line, not command startup.
- Subsumption is a hint supported only by each row's listed kills, never a deletion recommendation. All production matrix runs used the whole active package. Other packages and repository-wide uniqueness were not tested.

Timing: warm toolchain setup skipped (0 seconds); npm ci reported 0.494 seconds. Clean baseline binary 28.260 seconds; inactive selector baseline 33.908 seconds. Explicit go test -c build 3.524 seconds. Matrix commands total 696.706 wall seconds, including Go startup/build overhead; binary elapsed total 648.526 seconds. Corrected 45 individual timing commands total 175.618 wall seconds. Additional family, companion-subsumer and probe timing logs are included. Initial build cost and individual standalone-vet build costs were not separately timed. No native products were built by these rows, but each mutant still used its own ADAMIC_BUILD_CACHE_DIR.

Production sources restored exactly to the starting commit. No main push or pull request.
