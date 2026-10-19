Rebased the closure-target cycle rule onto pinned main, retaining main's newer lowering guards.
Commits: delivery SHA in final response; base dbdf977f5b8b0979d414ebe6f20e7aadd312e309; source 20fde75fa96b48f3dea2e3612b418acae666332f.
Checks: 35 focused witnesses/guards, all 298 lowering top-level tests in 15 bounded shards, full flow corpus, readers, affected fixtures, counts and vet pass.
Mutants: all six source-report mutants fail at their intended assertions; every production file is restored.
Not covered: full gate, unrelated oracle/consumer leaves, WASI, benchmarks, full pristine-source cycle ledger or admission census; one existing mixed-union array child skips pending views-v3. No ahra, PR or Loom submission.

Applied only the source commit's own change using `git cherry-pick -n 20fde75fa96b48f3dea2e3612b418acae666332f` onto the exact pinned base. No conflicts or conflicted hunks occurred. Main's generated async cycle identity audit remains intact and TestAsyncGeneratedIdentityCannotBeClaimedBySource passes in the complete lowering run. The source's compiler patch passes reverse-apply checking on the rebased tree, and source/rebased stable patch IDs are identical (excluding generated counts and review evidence). There are no additional production/compiler edits. The cache's private contents map remains SortedMap, as in the source.

Conservative assumption: unknown closure producers remain cycle-capable. The function-slot decision uses actual producer flow and captured cells; no function-type compatibility guess supplies a bounded target set. The previously stale cwd refusal expectation is already corrected in the source change. Its exact program is stage3/namespace-init-sys/cwd.a: memoize(() => process.cwd()), a string cache, Debug.enabled initialization, and a comparison with process.cwd(). The old type-matched rule incorrectly put the returned memoizer closure into callback's target set and refused callback as cycle-capable. The actual callback captures no environment; the returned closure owns callback and a string cache, which cannot reach it. Node, JavaScript, release native and ASan/UBSan plus leak checks agree on true. Unsupported realpath remains independently stopped at node:fs.native; executable namespace reads before initialization remain rejected.

Counts were restored from pinned main and regenerated once successfully on Linux with house clang 20.1.8 using the documented unanchored -run TestCountsAreRecorded -args -update-counts command. The first anchored attempt selected no tests and wrote no table; its failed writer output is preserved. Regeneration adds eleven rows, changes six existing rows and removes none. All numeric changes from the source report match. Main's newer instantiation_find.a is the sole additional changed row: 13/13/28/31/7/0 becomes 13/13/22/25/7/0. Observation: allocations/frees/peak/regions stay unchanged and six retain/release pairs disappear. Inference: the now bounded getPattern callback flow improves borrowing. Its dedicated TestInstantiationFind independently passes uncached Node/JavaScript/release/sanitized native and leak agreement. The five source-report affected existing fixtures also pass uncached; their actual fixture bodies ran, not just family wrappers.

Column order below: allocations/frees/retains/releases/peak/regions.

| Fixture | Pinned main | Rebased |
|---|---|---|
| internal/oracle/testdata/borrow_chain_unknown.a | 8/8/3/8/7/0 | 8/8/2/7/7/0 |
| internal/oracle/testdata/call_targets_element.a | 45/43/32/60/7/2 | 45/43/30/58/7/2 |
| internal/oracle/testdata/generic_functions.a | 35/35/32/55/8/0 | 35/35/29/52/8/0 |
| internal/oracle/testdata/getters_census_binary_function.a | new | 39/39/37/62/25/0 |
| internal/oracle/testdata/getters_census_optional_boolean_function.a | new | 36/36/34/56/24/0 |
| internal/oracle/testdata/getters_census_optional_comment_function.a | new | 33/33/37/56/23/0 |
| internal/oracle/testdata/getters_census_optional_type_function.a | new | 36/36/40/62/24/0 |
| internal/oracle/testdata/getters_census_primary_function.a | new | 33/33/31/50/23/0 |
| internal/oracle/testdata/getters_census_unary_function.a | new | 36/36/34/56/24/0 |
| internal/oracle/testdata/getters_census_unary_node_function.a | new | 36/36/34/56/24/0 |
| internal/oracle/testdata/getters_census_update_comment_function.a | new | 36/36/40/62/24/0 |
| internal/oracle/testdata/getters_census_update_node_function.a | new | 39/39/40/65/25/0 |
| internal/oracle/testdata/getters_census_update_type_function.a | new | 39/39/43/68/25/0 |
| internal/oracle/testdata/instantiation_find.a | 13/13/28/31/7/0 | 13/13/22/25/7/0 |
| internal/oracle/testdata/route_targets_sort_values.a | 84/84/92/142/17/0 | 84/84/96/150/17/0 |
| internal/oracle/testdata/user_iterators.a | 663/663/455/904/64/0 | 663/663/442/891/64/0 |
| stage3/namespace-init-sys/cwd.a | new | 6/6/8/12/6/0 |

Runtime C: the source's own delta and this rebased delta under internal/native/runtime both contain zero changed lines. No runtime lines need clearance. See runtime-delta-comparison.json. No new runtime feature set or compiler-consuming package was added; existing declarations and exact runtime witnesses remain, and TestRuntimeSetsAreEachUsed passes. New refused fixtures remain under cycle_closure_targets_refused/, outside flow's top-level fixture glob. No cohere code was copied.

Every recorded production test result below passes, with the existing mixed-union child skip noted above. Full per-test timings are in test-seconds.json; the 35 own focused events are also listed below. Counts' successful regeneration contains 1102 passing events. The flow corpus runs with ADAMIC_UNIT_BUDGET=1, GOMAXPROCS=4 and -parallel=4. Full lowering is selected from go test -list and split into 20-top-level-test shards, each with a 90-second Go timeout and 120-second outer limit. Setup builds every package; consumer tests focus on the call-target, borrowing, loop-call, flow and JavaScript paths, with the full flow corpus additionally exercised. Unrelated package tests are left to Loom as requested.

| Log | Passing events | Skipped events | Maximum event seconds |
|---|---:|---:|---:|
| affected-fixtures.log | 134 | 0 | 11.55 |
| consumers.log | 6 | 0 | 0.14 |
| counts-update.log | 1102 | 0 | 19.66 |
| flow-corpus.log | 856 | 0 | 29.53 |
| focused.log | 35 | 0 | 55.21 |
| instantiation-find.log | 1 | 0 | 4.37 |
| lower-shard-0.log | 27 | 0 | 7.79 |
| lower-shard-1.log | 68 | 0 | 1.48 |
| lower-shard-10.log | 31 | 0 | 0.4 |
| lower-shard-11.log | 101 | 0 | 10.42 |
| lower-shard-12.log | 351 | 0 | 1.59 |
| lower-shard-13.log | 71 | 0 | 1.02 |
| lower-shard-14.log | 50 | 1 | 0.91 |
| lower-shard-2.log | 43 | 0 | 0.68 |
| lower-shard-3.log | 56 | 0 | 11.02 |
| lower-shard-4.log | 74 | 0 | 4.39 |
| lower-shard-5.log | 54 | 0 | 1.63 |
| lower-shard-6.log | 118 | 0 | 5.08 |
| lower-shard-7.log | 127 | 0 | 13.59 |
| lower-shard-8.log | 180 | 0 | 2.15 |
| lower-shard-9.log | 88 | 0 | 1.94 |
| readers.log | 5 | 0 | 49.25 |
| table-and-families.log | 3 | 0 | 0.02 |

Every mutant was caught by an assertion, never by a build error or timeout. Sources are saved only as .patch files. The runners use a 90-second Go timeout, 120-second subprocess limit, and restore originals in finally blocks.

| Mutant | Catcher | Assertion |
|---|---|---|
| omit-boundary-inputs | TestClosureTargetsKnownCallerDoesNotHideUnknownCaller | known caller hid unbounded caller |
| unknown-is-empty | TestClosureTargetsBoundOnlyProvenValues | bounded without proof |
| omit-captures | TestClosureCycleDirect | want cycle refusal |
| drop-return-target | TestClosureTargetsFollowParametersReturnsAndMapValues | want both returned targets |
| omit-map-captures | TestClosureCycleMap | want cycle refusal |
| type-matched-captures | TestNamespaceAmbientCwdInitialization and TestClosureCycleAmbientCwd | adamic/cycle-capable |

Own focused test timings:

| Test | Seconds |
|---|---:|
| TestEveryMapAProgramHoldsIsSorted | 0 |
| TestClosureTargetsKnownCallerDoesNotHideUnknownCaller | 0 |
| TestClosureTargetsFollowParametersReturnsAndMapValues | 0 |
| TestClosureTargetsBoundOnlyProvenValues | 0 |
| TestNativeAgreementRowsNameRegisteredPrograms | 0.13 |
| TestEveryTestIsParallelOrSaysWhy | 1.08 |
| TestNamespaceAmbientContextsDoNotExecute | 0.5 |
| TestNamespaceAmbientHostInitialization | 1.54 |
| TestNamespaceAmbientCwdInitialization | 1.77 |
| TestClosureCycleDirect | 37.58 |
| TestGettersCensusUpdateTypeFunction | 17.62 |
| TestGettersCensusOptionalCommentFunction | 55.21 |
| TestGettersCensusUpdateCommentFunction | 55.21 |
| TestGettersCensusOptionalBooleanFunction | 55.21 |
| TestGettersCensusClosureUnionNext | 0.59 |
| TestGettersCensusUnaryNodeFunction | 1.23 |
| TestGettersCensusUpdateNodeFunction | 1.24 |
| TestGettersCensusOptionalTypeFunction | 1.34 |
| TestGettersCensusCallbackPairMethods | 0.25 |
| TestGettersCensusHigherCallbackPairMethods | 0.29 |
| TestGettersCensusSnapshotNext | 0.26 |
| TestGettersCensusClosureNumberNext | 0.23 |
| TestGettersCensusClassMap | 0.9 |
| TestGettersCensusUnaryFunction | 0.85 |
| TestGettersCensusClosureNumberMethods | 0.26 |
| TestGettersCensusThrowingObject | 0.96 |
| TestGettersCensusLazyObject | 0.53 |
| TestClosureCycleMap | 0.25 |
| TestGettersCensusOverloadedFunction | 0.35 |
| TestGettersCensusPrimaryFunction | 1.21 |
| TestClosureCycleJoined | 0.28 |
| TestGettersCensusBinaryFunction | 0.87 |
| TestRuntimeSetsAreEachUsed | 1.88 |
| TestGettersCensusBinaryInline | 19.02 |
| TestClosureCycleAmbientCwd | 19.35 |

The initial focused attempt reached its 180-second outer limit during cold Go compilation, with no test events; the subsequent attempt passed all 35 events. The unmatched counts attempt measured and wrote no rows. Both attempts and successful retries are recorded rather than hidden. No test leaf exceeded 60 seconds, and no over-budget leaf was rerun.

Toolchain setup passed. nproc=5; cgroup cpu.max=400000 100000 (four CPUs); Go 1.27.1, Node 24.19.0, clang 20.1.8. GOPROXY=https://proxy.golang.org|direct; source /workspace/adamic-tools/env.sh; GOMAXPROCS=4 for checks. Setup timing lines:

```text
setup: go ready (0.076s)
setup: node ready (0.087s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.021s
setup: markdown dependencies ready (0.211s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.421s)
setup: submodules ready (2.704s)
setup: shared cache ready (4.020s)
setup: go build ready (454.360s)
setup: test binaries deferred (use --warm-tests) (455.145s)
setup: build cache warm (455.148s)
setup: build-flags commit=dbdf977f5b8b0979d414ebe6f20e7aadd312e309 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.00 0.00 0.00 1/140 864 load-after=7.29 6.83 3.39 2/184 8705
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (455.189s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.zEJoxb
```

All test output goes directly to the named log, never through a pipe. commands.json records commands, exits and wall seconds, including both unsuccessful attempts. Fetch/apply/setup commands:

```text
timeout 120 git fetch origin
timeout 120 git fetch origin 20fde75fa96b48f3dea2e3612b418acae666332f
git checkout -b compiler/loom-r-cycle-targets dbdf977f5b
timeout 120 git cherry-pick -n 20fde75fa96b48f3dea2e3612b418acae666332f
git restore --source=dbdf977f5b --staged --worktree internal/oracle/counts.md
export GOPROXY='https://proxy.golang.org|direct'
timeout 600 bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
export GOMAXPROCS=4
```

Verification commands and outputs:

```text
timeout 180 go test ./internal/ir ./internal/lower ./internal/oracle ./cmd/adamic-gate -run '^(TestEveryMapAProgramHoldsIsSorted|TestEveryTestIsParallelOrSaysWhy|TestNativeAgreementRowsNameRegisteredPrograms|TestRuntimeSetsAreEachUsed|TestClosureTargets.*|TestNamespaceAmbient.*|TestClosureCycle.*|TestGettersCensus.*)$' -count=1 -parallel=4 -json -timeout 90s
  exit=124 wall=180.065s log=focused-initial.log
timeout 300 go test ./internal/ir ./internal/lower ./internal/oracle ./cmd/adamic-gate -run '^(TestEveryMapAProgramHoldsIsSorted|TestEveryTestIsParallelOrSaysWhy|TestNativeAgreementRowsNameRegisteredPrograms|TestRuntimeSetsAreEachUsed|TestClosureTargets.*|TestNamespaceAmbient.*|TestClosureCycle.*|TestGettersCensus.*)$' -count=1 -parallel=4 -json -timeout 90s
  exit=0 wall=129.208s log=focused.log
timeout 600 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -parallel=4 -json -timeout 570s -args -update-counts
  exit=1 wall=5.396s log=counts-unmatched.log
timeout 600 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -parallel=4 -json -timeout 570s -args -update-counts
  exit=0 wall=106.457s log=counts-update.log
timeout 120 go test ./internal/lower -list '^Test' -timeout 90s
  exit=0 wall=5.561s log=lower-list.log
timeout 120 go test ./internal/lower -run '^(TestProduct_LowerAgreementNative|TestAgreementRejectsWrongLoweredOutput|TestAgreementRejectsEmptyAnswer|TestAgreementAcceptsComputedAnswer|TestAgreementAcceptsExpectedPanic|TestAgreementEnumConstantWitness|TestAgreementInheritedFieldWitness|TestAgreementParameterPropertyWitness|TestAgreementStaticInitializerWitness|TestAgreementParseIntRadixWitness|TestNativeAgreementRejectsWrongOutput|TestNativeAgreementRejectsWrongExit|TestNativeAgreementRowsNameRegisteredPrograms|TestMixedTupleSpreadIsNotYet|TestArgumentsLengthReadNeighbors|TestArgumentsLengthRefusalFixtures|TestArgumentsCountLocalsAreNumberedTheSameEveryTime|TestArgumentsLengthReadKeepsReaderFact|TestArrayPredicatePreservesDeclaredElementContract|TestArrayPredicateDoesNotMisclassifyNativeTuples)$' -count=1 -parallel=4 -json -timeout 90s
  exit=0 wall=11.415s log=lower-shard-0.log
timeout 120 go test ./internal/lower -run '^(TestUnknownArrayPredicateRefusesUnrepresentedObservations|TestArrayPredicateCoexistsWithUnknownReflection|TestAsyncGeneratedIdentityCannotBeClaimedBySource|TestAsyncGapsNameTheMissingPiece|TestAsyncTypeOfCompoundFunctionValueIsNotYet|TestAsyncGeneratorsRemainRefused|TestUncheckableCastsStayRefused|TestCheckedCastProofAndElision|TestCensusMarkerResultIsAssignable|TestCensusMarkerZeroCallIsNotAssumedSafe|TestOptionalFunctionValueRelation|TestOverloadedShorthandFunctionValueStaysNotYet|TestCensusOverloadBinderGuards|TestCensusOverloadReturnProof|TestCensusPredicateInteractionBoundary|TestCensusPredicateInteractionEscapes|TestCensusPredicateMarkerKeepsProofBoundaries|TestCensusPredicateMarkerCallStaysNotYet|TestCensusOverloadRelation|TestCensusSmallFiniteKeyRead)$' -count=1 -parallel=4 -json -timeout 90s
  exit=0 wall=5.868s log=lower-shard-1.log
timeout 120 go test ./internal/lower -run '^(TestCensusRestMutableElements|TestCensusBooleanDeadBranch|TestCheckPragmasAreRefused|TestCheckPragmaNeighborsCompile|TestTsgoHonorsNoCheckInAdamicFiles|TestClassFeaturesPrivateChecker|TestClassFeaturesPrivateStorage|TestClassFeaturesAccessorRefusals|TestClassFeaturesStaticDeclarationsExecute|TestClassFeaturesStaticSoundness|TestClassFeaturesAccessorCaptureCycle|TestClassFeaturesNarrowedAccessor|TestClassFeaturesStaticInterfaceCycle|TestClassFeaturesThrowingAccessorSpread|TestInheritanceRefusesUnsoundOverrides|TestInheritanceKeepsCheckerConstructorRules|TestInheritanceAllowsSoundOverrides|TestInheritanceHasClassIdentity|TestInheritanceCycleFinderIncludesInheritedFields|TestInheritanceRejectsUnsupportedConstructorShapes)$' -count=1 -parallel=4 -json -timeout 90s
  exit=0 wall=4.952s log=lower-shard-2.log
timeout 120 go test ./internal/lower -run '^(TestInheritanceRefusesFalseNominalViews|TestInheritanceRefusesThisBeforeSuperReturns|TestInheritanceKeepsNominalTupleDestructuring|TestInheritanceGenericMonomorphizations|TestInheritanceRefusesGrowingGenericClasses|TestInheritanceGenericSoundness|TestInheritanceConditionalThisRules|TestInheritanceGenericViewsKeepNominalArguments|TestInheritanceGenericNominalConstraints|TestInheritanceNativeSignatureNeighbors|TestInheritanceNativeSignatureLimits|TestClassArgumentsUseCheckerIdentity|TestClassStaticInitializerCallIsEmitted|TestClassStaticSideEffectMatchesNode|TestClassWrongOutputIteratorReceiver|TestClassWrongOutputKeysRepair|TestClassWrongOutputPrivateRepair|TestClockGenericReturnsT01Shapes|TestClockGenericReturnsT01DeclinesBrandedPrimitives|TestClockGenericReturnsT01RejectsNullBeforeBody)$' -count=1 -parallel=4 -json -timeout 90s
  exit=0 wall=15.405s log=lower-shard-3.log
timeout 120 go test ./internal/lower -run '^(TestClockGenericReturnsT01RejectsIndexBeforeBody|TestConstructorCacheOneArgumentRemainsNotYet|TestDefiniteAssignmentUsesReadiness|TestNestedEmptyArrayElementKinds|TestEmptyLiteralGenericReturnUsesSamePath|TestEntriesAllocationProof|TestEntriesRecordBoundaries|TestFlagEnumsOpen|TestFlagEnumsDomain|TestEnumNeverDefault|TestFlagEnumLiteralSpellings|TestFlagEnumMemberAliases|TestEnumNameEnumeration|TestFlagEnumInlineIteration|TestFlagEnumAliasBoundaries|TestEnumMemberValuesAndReverseNameMatchNode|TestEnumFlagProofsWithoutObservableLoweringEffect|TestEnumReverseMappingUsesSingleSlot|TestEnumInitializationReach|TestEnumInitializationGraphMemo)$' -count=1 -parallel=4 -json -timeout 90s
  exit=0 wall=9.166s log=lower-shard-4.log
timeout 120 go test ./internal/lower -run '^(TestEnumNamespaceSharedCycle|TestNumericEnumsAreOpen|TestStringEnumsStayClosed|TestNumericEnumNeverProof|TestNumericEnumLiteralPromises|TestEnumSlotViews|TestEnumSwitchExhaustiveness|TestEnumLimitsStayLoud|TestConstEnumErasesRuntimeObject|TestEnumIdentityAcrossModules|TestReadonlyFieldsAreJudgedByTheirConstructorsWrites|TestParseIntMapUsesIndexRadix|TestFSOpenStringFlagsLower|TestFSRemoveDefaultRetryDelayLowers|TestFSExistsOperation|TestFSStatThrowsByDefault|TestFunctionValueUnionViewsStayNotYet|TestGenericFunctionPolymorphicRecursionIsRefused|TestGenericUnionFixtureHasSeparateInstances|TestGenericJSONUnionArrayIsNotYet)$' -count=1 -parallel=4 -json -timeout 90s
  exit=0 wall=8.961s log=lower-shard-5.log
timeout 120 go test ./internal/lower -run '^(TestUndecidedCycleReadsUseReadyChecks|TestInputSpreadCoverage|TestDefaultTaggedInterfaceAdmission|TestIteratorGapsAreExplicit|TestIteratorViewsCannotHideReturn|TestIteratorViewsCannotEraseReceivers|TestIteratorDestructuringDoesNotLieAboutExhaustion|TestGeneratorsAreRefusedEvenWithoutYield|TestLiteralMethodCapturesCannotMakeCycles|TestLiteralMethodViewsDoNotLoseThis|TestLiteralMethodSignatureViewsDoNotLoseThis|TestIteratorDescriptorReasons|TestRepresentedMethodReplacementIsNotYet|TestDestructuredMethodsCannotLoadOwnSlots|TestGenericIteratorViewsPreserveNativeArguments|TestIteratorMapperIndexHasNumberRepresentation|TestIteratorSymbolKeysAreNotStringKeys|TestLibraryLanguageBoundaries|TestLibraryMapSetGapsStayRefused|TestLibraryMapSetIteratorCopyTypesRefused)$' -count=1 -parallel=4 -json -timeout 90s
  exit=0 wall=8.865s log=lower-shard-6.log
timeout 120 go test ./internal/lower -run '^(TestLibraryMethodValues|TestLibraryMethodValueBoundaries|TestLibraryMethodValueSafety|TestNodeBufferRefusals|TestNodeFSFileOptionsBorrow|TestNodeFSFileDoesNotAuthorizeMutableWidening|TestNodeFSFileRefusesOptionEffects|TestNodeFSFileRefusesVoidValues|TestNodeFSFileRefusesPinnedUnsupportedOverloads|TestNodeFSFileNamespaceImport|TestNodeFSFileQualifiedErrorType|TestNodeFSFileKeepsDetachedMethodRefusal|TestNodeFSFileBufferBorrow|TestNodeFSFileScratchOptionsBorrow|TestNodeFSFileScratchOverloadsAreNamed|TestNodeHostConsumesArgument|TestNodeHostOptionsRefusalBoundaries|TestNodeHostOptionsConsumptionLowers|TestNodeHostOptionsReferencesAcrossFiles|TestNodeHostRmPresentUndefinedRefused)$' -count=1 -parallel=4 -json -timeout 90s
  exit=0 wall=21.01s log=lower-shard-7.log
timeout 120 go test ./internal/lower -run '^(TestNodeLibraryNamesUnimplementedMembers|TestNodeLibraryDoesNotRefuseTypeOnlyOrUserNames|TestNodeLibraryDistinguishesReceiverOwners|TestNodeLibraryQualifiedTypeAssertion|TestObjectRefusalsExplainSoundness|TestObjectUnprovenShapesStayNotYet|TestLibraryRegexOffsetRequiresNoCaptures|TestLibrarySpreadRegExpUnion|TestLibrarySpreadErrorView|TestLibrarySpreadPlainProducers|TestLibrarySpreadMapView|TestLibrarySpreadSetView|TestLibrarySpreadDateView|TestLibrarySpreadTypedArrayView|TestLibraryStringRefusals|TestConsoleLowersToWriteLine|TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat|TestWhatZeroOneRefusesIsRefusedWithAFix|TestAMutableLocationSeenWiderIsRefused|TestAViewThatCantWriteIsNotRefused)$' -count=1 -parallel=4 -json -timeout 90s
  exit=0 wall=9.895s log=lower-shard-8.log
timeout 120 go test ./internal/lower -run '^(TestATupleSeenAsAnArrayIsNotYet|TestAMethodReadAsAValueIsRefused|TestModuleNamespaceLimitsStayLoud|TestModuleNamespaceInitializedReadProof|TestCallableNamespaceLimitsStayLoud|TestCallableNamespaceReceiverStaysLoud|TestNamespaceClassEarlyConstructionStaysLoud|TestNamespaceReceiverArrowInheritsScope|TestNamespaceAmbientHostInitialization|TestNamespaceAmbientCwdInitialization|TestNamespaceAmbientContextsDoNotExecute|TestNamespaceCallGraphLinearWork|TestNamespaceCallGraphCycleUnion|TestNamespaceEnumInitializationIndependentOfModuleAnalysis|TestParserFactoryBindingHoisting|TestMissingBindingSymbolHasStructuredLocation|TestTscNamespaceDeclarationShapes|TestNamespaceLimitsStayLoud|TestNamespaceReceiverRefusal|TestNamespaceTypesErase)$' -count=1 -parallel=4 -json -timeout 90s
  exit=0 wall=6.427s log=lower-shard-9.log
timeout 120 go test ./internal/lower -run '^(TestTracingNamespaceEscapeStaysNotYet|TestDebugNamespaceMergedCapabilities|TestNamespaceReturnedAssignmentLimits|TestNamespaceInitializationReachability|TestNamespaceClosedCallGraphEdges|TestEmptyNeverMapCannotGainWritableInhabitants|TestNestedFunctionGapsAreLoud|TestNestedFunctionCycleIsRefused|TestNestedEnvironmentHasOneAllocationSite|TestNestedEnvironmentCycleIncludesDisjointSlots|TestNestedCapturedParametersAreOwned|TestClosedFrameInputRejectsMutation|TestNestedRebindingNotYet|TestNestedCallbackCycleIsRefused|TestNestedBodylessDeclarationsAreLoud|TestNestedRestIsSupported|TestNonNullAssertionLowersToNullishPanic|TestNonNullAssertionOnPresentTypeIsErased|TestOptionalIndexingMapShapeRefused|TestOptionalIndexingKeepsIndexSignaturesRefused)$' -count=1 -parallel=4 -json -timeout 90s
  exit=0 wall=4.868s log=lower-shard-10.log
timeout 120 go test ./internal/lower -run '^(TestOptionalIndexingKeepsUnsupportedStorageNotYet|TestOptionalWideningRefused|TestOptionalWideningSpreadOverwrite|TestOverloadInferenceWitnesses|TestAbfe962OverrideDefaultAddedRefusesNativeSignature|TestOverrideOptionalNumberRefusesNativeSignature|TestOverrideOptionalRefusesNativeSignature|TestParallelReadonlyCheckerRepresentation|TestParallelGlobalMarkingRoots|TestParameterPropertiesSoundness|TestParameterPropertyCheckerContracts|TestParameterPropertyCallbackReceiver|TestParameterPropertyValueMatchesNode|TestPredicateSummaryParameterIndex|TestPredicateOverloadRuntime|TestIndirectPredicateOverloadIsPending|TestPredicateBodyProof|TestConditionAssertionAdmission|TestPredicateCallbackContracts|TestEveryNeedsCallbackEffects)$' -count=1 -parallel=4 -json -timeout 90s
  exit=0 wall=14.73s log=lower-shard-11.log
timeout 120 go test ./internal/lower -run '^(TestPredicateUseRegions|TestPredicateUsesBelongToEachCall|TestUnprovenPredicateReturnsAreRefused|TestPredicateBodiesAreProven|TestPrimitiveAdmittingSlotsUseBoxes|TestInheritedLibraryReadsNeverLoadOwnFieldsObject|TestInheritedLibraryReadsNeverLoadOwnFieldsArray|TestInheritedLibraryReadsNeverLoadOwnFieldsTuple|TestInheritedLibraryReadsNeverLoadOwnFieldsNumber|TestInheritedLibraryReadsNeverLoadOwnFieldsString|TestInheritedLibraryReadsNeverLoadOwnFieldsBoolean|TestInheritedLibraryReadsNeverLoadOwnFieldsClass|TestInheritedLibraryReadsNeverLoadOwnFieldsMap|TestInheritedLibraryReadsNeverLoadOwnFieldsSet|TestInheritedLibraryReadsNeverLoadOwnFieldsFunction|TestInheritedLibraryReadsNeverLoadOwnFieldsError|TestInheritedLibraryReadsNeverLoadOwnFieldsUnion|TestInheritedLibraryReadReceiversEachHaveOneTest|TestPrototypeMethodsAreRefusedWithReasons|TestNullishPrototypeReadsAreRejectedByChecker)$' -count=1 -parallel=4 -json -timeout 90s
  exit=0 wall=10.462s log=lower-shard-12.log
timeout 120 go test ./internal/lower -run '^(TestPrototypeHazardsBehindObjectViewsAreNotYet|TestUnrepresentedPrototypeCallsAreNotYet|TestIsPrototypeOfReadsExplainThePrototypeRefusal|TestProvenRelationsRefuse|TestProvenRelationsErase|TestReadinessElisionRequiresDominatingAssignment|TestRegExpNativeRefusals|TestRegExpSourceNode|TestRegExpUnicodeClassSourceAgreesWithNode|TestRegExpSlashClassSourceAgreesWithNode|TestRegExpEscapedSlashSourceAgreesWithNode|TestRepresentationClockSourceCheckedTypes|TestPrivateAndPublicStaticsAgreeWithNode|TestNamespaceFactoryBindingsAgreeWithNode|TestReadinessErrorIncludesReceiverExpression|TestSuppressionDirectivesAreRefused|TestSuppressionDirectiveSoundNeighbors|TestSwitchConditionalBreakFallsThrough|TestSwitchEmptyElseFallsThrough|TestSwitchEmptyCaseFallsThrough)$' -count=1 -parallel=4 -json -timeout 90s
  exit=0 wall=6.718s log=lower-shard-13.log
timeout 120 go test ./internal/lower -run '^(TestSwitchEmptyBlockFallsThrough|TestTasteRepresentationLimitsStayExplicit|TestTypedArraysLower|TestTypedArrayGaps|TestTypedArrayViewsCannotChangeRepresentation|TestTypedArrayFromArray|TestTypedArrayFromEmptyArray|TestUnknownReflectionRefusals|TestViewObjectContractsAreAvailableToEraser|TestViewObjectWritesNeedSourceCertificate|TestMixedUnionContractGraph|TestMixedUnionContractFailureDoesNotCertifyRetry|TestMixedUnionContractUnknownMemberFails|TestMixedUnionContractRecursiveMember|TestMixedUnionContractPhantomBrandUsesPrimitiveBase|TestMixedUnionContractPhantomVoidIsUndefined|TestUntaggedViewMemberTags|TestUntaggedViewStructuralFallback)$' -count=1 -parallel=4 -json -timeout 90s
  exit=0 wall=5.433s log=lower-shard-14.log
timeout 180 go test ./internal/oracle -run '^(TestNativeAgreesWithNode_[0-9]+|TestNativeAgreesWithNode_user_iterators)$/^internal$/^oracle$/^testdata$/^(route_targets_sort_values|borrow_chain_unknown|call_targets_element|generic_functions|instantiation_find|user_iterators)\.a$' -count=1 -parallel=4 -json -timeout 90s
  exit=0 wall=16.197s log=affected-fixtures.log
timeout 180 go test ./internal/oracle -run '^TestInstantiationFind$' -count=1 -parallel=4 -json -timeout 90s
  exit=0 wall=9.473s log=instantiation-find.log
timeout 600 python3 /workspace/adamic/review/compiler/loom-r-cycle-targets/run-mutants.py
  exit=0 wall=103.765s log=mutants.log
timeout 180 python3 /workspace/adamic/review/compiler/loom-r-cycle-targets/run-admission-mutant.py
  exit=0 wall=29.487s log=admission-mutant.log
timeout 180 go test ./internal/ir -run '^(TestCallTargetReaders|TestCallTargetsIncludeEveryDescendant|TestEveryMapAProgramHoldsIsSorted|TestAProgramIsTheSameBytesEveryTimeGobWritesIt|TestASortedMapIsNilWhereAPlainMapIs)$' -count=1 -parallel=4 -json -timeout 90s
  exit=0 wall=53.952s log=readers.log
timeout 180 go vet ./internal/ir ./internal/lower ./internal/oracle
  exit=0 wall=36.328s log=vet.log
timeout 180 go test ./internal/native ./internal/flow ./internal/fresh ./internal/javascript -run '^(TestCallTargetsElementBorrowPlan|TestBorrowChainTargets|TestLoopCallCoverage|TestDebuggerHasNoFlowInstruction|TestRegexOperandsStillJudgeCycleWrites|TestViewMixedUnionSelectsNumber)$' -count=1 -parallel=4 -json -timeout 90s
  exit=0 wall=6.79s log=consumers.log
timeout 600 go test ./internal/flow -run '^(TestFlowCorpus.*|TestFlowProgram.*|TestFlowSingleAssignment.*|TestFlowMutationRanges.*|TestFlowGraphPaths.*|TestFlowLiveness.*)$' -count=1 -parallel=4 -json -timeout 570s
  exit=0 wall=110.92s log=flow-corpus.log
timeout 180 go test ./internal/oracle -run '^(TestCountsAreRecordedTable|TestFamiliesOwnEveryCaseOnce|TestFamilyWrappersAreCurrent)$' -count=1 -parallel=4 -json -timeout 90s
  exit=0 wall=4.408s log=table-and-families.log
```

Mutation runner source files record each exact inner go test command and mutation site. The final source-delta-retained.json and runtime-delta-comparison.json were rechecked after all mutations. Strict whitespace checks pass for docs/internal; patch evidence preserves unified-diff context whitespace. Historical review files inherited from the source commit remain historical; this directory is the evidence for this rebase.

Delivery commands:

```text
timeout 120 git commit -F review/compiler/loom-r-cycle-targets/commit-message.txt
timeout 120 git push origin HEAD:refs/heads/compiler/loom-r-cycle-targets
timeout 120 git ls-remote origin refs/heads/compiler/loom-r-cycle-targets
```
