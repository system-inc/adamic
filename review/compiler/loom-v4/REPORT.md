Built: V4 checked callables with the source branch's V3 arrays, squashed for Loom.
Commits: source 8a4c135e4f3cf611d2fc7801f4e640d7804af7f9; initial main cd7880cf5145b33a5bf6bb2227e216d7fb9d63c0; landing base c0fd8945b872e57a341ccee5973e3f4d1b75c34b.
Commands and outputs: 268 focused top-level checks passed, seven carried skips; counts and changed-package vet green.
Mutants: 17 independently restored source/witness mutants caught; embedded IR and private-runtime mutants passed their behavioral catchers.
Not covered: the full gate and seven carried pending tests. The 30-minute first-green push target was missed.

The final commit is the pushed compiler/loom-v4 branch tip, one commit above its landing base. Main advanced only under stage3/census/tsc-closure during validation; its compiler, runtime and counts table did not change. The branch is rebased once onto that refreshed tip before pushing. No PR or Loom submission is made.

Code whitespace checks exclude raw review logs and .patch evidence: unified-diff context indentation and sanitizer output are retained verbatim. Two inherited trailing spaces in native test C strings were removed; the affected native probes, reader guard and native/oracle vet passed again. gofmt -l on all changed Go files was empty.

The three squash conflicts were resolved once:

- internal/flow/trace_test.go: use V3 sparse-array snapshots (length and present keys), retaining main's own-property descriptor check so object getters are never called by tracing.
- internal/ir/ir.go: retain main's SortedMap contract registry and newer fields; include all V3 array/V4 adapter metadata. Use SortedMap for V4's new optional-field map too, preserving deterministic gob products.
- internal/oracle/checked_views_v2_counts_test.go: retain main's per-row countCase registration, omitting the three V4 compile-time refusal rows. Existing V4 refusal readers preserve their Node controls. Port the V3 array count registrations to the same per-row form.

Test grain and products: split the large new array-read, array-mutant, array-union and native callable tables into top-level functions. Every executed focused leaf, product and count leaf is under 60 seconds on cpu.max=400000 100000 (4 CPUs); nproc=5. Process wall times include startup and helper setup. Go JSON timings are labeled as leaf elapsed times. Native runtime products are run ahead and their build times are recorded separately. Test-binary compilation is recorded in the build rows, not reported as a test leaf.

V4's ASan plus counted runs were absent from main's runtimeOptions. Declare that mode and three setup products in checked_views_v4_runtime_products_test.go. Also declare the receiver plus RegExp callback feature combination. Main's runtime-set check accepts a source-profile witness callback in addition to a persistent file, and still checks the actual emitted features exactly. The same source constant is used by that witness and the RegExp callback oracle. No other package's tests were edited to make room. No new Go package consumes the compiler, so cloud/fast-gate/compiler-dependencies.json needs no entry. No new product or test stores a fixed machine/build directory.

Conservative assumptions: the exact anchored legacy count name no longer exists on main. Use ^TestCountsAreRecorded to select main's generated shards. Existing upstream TypeScript checker-profile witnesses remain references, not new Adamic admissions; new Adamic programs are .a. No cohere implementation was copied.

Counts were written once. The requested anchored command ran no tests (5.561 s) and refused to write. The first sharded attempt failed because @types/node 25.3.3 was missing (52.428 s), also without writing. npm ci --ignore-scripts --prefix stage3/api installed the locked dependencies (0.738 s). The single successful update took 57.221 s; independent verification without -update-counts took 45.066 s. All 30 changed rows are below and in counts-diff.json. Twenty-six rows add the carried V3 sparse-array and checked-array fixtures. Three rows disappear because V4 now refuses their escapes before emission. The only changed existing numeric row is namespace_method_receiver.a: 7/7/9/16/5/0 becomes 7/7/7/14/5/0. Attribution from the lowering diff: an explicit this parameter is a declaration contract and is excluded from runtime argument slots; the fixture makes two method calls. This attribution is a code reading, not a separate counterfactual measurement.

| Counts row | Before | After | Attribution |
|---|---|---|---|
| internal/oracle/testdata/library_array_holes_length.a | absent | 17/17/3/19/6/0 | new V3 fixture |
| internal/oracle/testdata/library_array_holes_range.a | absent | 24/24/12/27/4/0 | new V3 fixture |
| internal/oracle/testdata/library_array_holes_scanner_probe.a | absent | 4/4/0/3/4/0 | new V3 fixture |
| internal/oracle/testdata/namespace_method_receiver.a | 7/7/9/16/5/0 | 7/7/7/14/5/0 | declaration-only this excluded from runtime arguments |
| stage3/interface-downcasts/lane2/array-alias.a | absent | 3/1/9/8/3/0 | new V3 fixture |
| stage3/interface-downcasts/lane2/array-boolean.a | absent | 2/0/4/3/2/0 | new V3 fixture |
| stage3/interface-downcasts/lane2/array-bounds.a | absent | 5/5/5/10/5/0 | new V3 fixture |
| stage3/interface-downcasts/lane2/array-element.a | absent | 4/4/4/7/4/0 | new V3 fixture |
| stage3/interface-downcasts/lane2/array-field.a | absent | 4/4/3/6/4/0 | new V3 fixture |
| stage3/interface-downcasts/lane2/array-holes-view.a | absent | 7/7/9/15/5/0 | new V3 fixture |
| stage3/interface-downcasts/lane2/array-iteration-bad.a | absent | 2/0/4/2/2/0 | new V3 fixture |
| stage3/interface-downcasts/lane2/array-iteration.a | absent | 6/6/3/8/4/0 | new V3 fixture |
| stage3/interface-downcasts/lane2/array-lazy.a | absent | 4/4/5/7/4/0 | new V3 fixture |
| stage3/interface-downcasts/lane2/array-length-kind.a | absent | 2/0/2/1/2/0 | new V3 fixture |
| stage3/interface-downcasts/lane2/array-map-bad.a | absent | 4/0/4/1/4/0 | new V3 fixture |
| stage3/interface-downcasts/lane2/array-map.a | absent | 7/7/5/12/6/0 | new V3 fixture |
| stage3/interface-downcasts/lane2/array-missing.a | absent | 1/0/2/1/1/0 | new V3 fixture |
| stage3/interface-downcasts/lane2/array-pop.a | absent | 6/6/5/11/6/0 | new V3 fixture |
| stage3/interface-downcasts/lane2/array-second.a | absent | 2/0/6/3/2/0 | new V3 fixture |
| stage3/interface-downcasts/lane2/array-string-literal.a | absent | 2/0/4/2/2/0 | new V3 fixture |
| stage3/interface-downcasts/lane2/array-undefined.a | absent | 2/0/4/1/2/0 | new V3 fixture |
| stage3/interface-downcasts/lane2/generic-array-element.a | absent | 2/1/1/2/2/0 | new V3 fixture |
| stage3/interface-downcasts/lane2/non-array.a | absent | 1/0/2/1/1/0 | new V3 fixture |
| stage3/interface-downcasts/lane2/nullable-array-element.a | absent | 1/0/4/2/1/0 | new V3 fixture |
| stage3/interface-downcasts/lane2/nullable-array-missing.a | absent | 0/0/1/1/0/0 | new V3 fixture |
| stage3/interface-downcasts/lane2/nullable-array-string.a | absent | 2/0/4/2/2/0 | new V3 fixture |
| stage3/interface-downcasts/lane2/object-element.a | absent | 3/3/9/10/3/0 | new V3 fixture |
| stage3/interface-downcasts/untagged/fixtures/callable-union-nested.a | 3/0/6/4/3/0 | absent | V4 compile-time escape refusal |
| stage3/interface-downcasts/untagged/fixtures/callable-union-wrong.a | 2/0/5/4/2/0 | absent | V4 compile-time escape refusal |
| stage3/interface-downcasts/v2/callable-producer-wrong.a | 2/0/5/4/2/0 | absent | V4 compile-time escape refusal |

Runtime files changed against landing main:

- internal/native/runtime/adamic.h
- internal/native/runtime/array.c
- internal/native/runtime/array_holes.c
- internal/native/runtime/closure.c
- internal/native/runtime/heap.c
- internal/native/runtime/map.c
- internal/native/runtime/object.c
- internal/native/runtime/regexp_replace.c
- internal/native/runtime/union.c
- internal/native/runtime/view_arrays.c
- internal/native/runtime/view_arrays.h
- internal/native/runtime/view_callables.c
- internal/native/runtime/view_callables.h
- internal/native/runtime/view_unions_untagged.c

Exact validation commands are in run-focused.py, run-remaining.py, run-fixed.py and run-source-mutants.py. No whole package test run or full gate was used. Focused tests were compiled with go test -c, then selected individually with -test.run ^NAME$, -test.count=1, -test.timeout=90s and a 95-second outer kill. The remaining pass used four independent processes with -test.parallel=1; completed tests were excluded. One in-flight serial test was interrupted during the switch and is recorded in focused.log. It was subsequently completed; no test over 60 seconds was rerun. Registry failures were rerun after their declarations were fixed. All original attempts remain in the JSONL/logs.

Final restored controls: timeout 120 go test ./internal/ir ./internal/oracle ./stage1/cohere/json -run <exact source-mutant catcher names and TestCallTargetReaders/TestProduct_CallTargetExports> -count=1 -timeout 90s -parallel 4 -json, 10.034 s wall. Final go vet ./internal/flow ./internal/fresh ./internal/ir ./internal/javascript ./internal/load ./internal/lower ./internal/native ./internal/oracle ./stage1/cohere/json passed with no diagnostics, 2.396 s. The earlier vet caught missing header setup in two split native probe helpers; it was restored before the focused native tests ran. /usr/bin/time was unavailable, so Bash time was used.

Every focused top-level test's final seconds follow. Additional subtest and count-leaf seconds are preserved in final-results.json and counts-verification.jsonl. Positive fixtures are held to source Node, release native, ASan/UBSan and backend Node; finished programs also run the leak checks where the source tests request them. Deliberate checked-view negatives pin exit 70 and exact output/diagnostics rather than claiming equality to unchecked Node. Original array checker witnesses are Node reference controls.

| Package | Test | Seconds | Measurement | Result |
|---|---|---:|---|---|
| internal/flow | TestArrayHolesAliasingIsFresh | 0.008 | process wall | pass |
| internal/flow | TestArrayHolesOperandMutations | 0.008 | process wall | pass |
| internal/flow | TestArrayHolesThrowAndIdentityRead | 0.016 | process wall | pass |
| internal/flow | TestArrayHolesTraceSnapshots | 0.064 | process wall | pass |
| internal/fresh | TestArrayHolesBackEdgeRemainsUnproven | 0.008 | process wall | pass |
| internal/fresh | TestArrayHolesOperandsStillJudgeCycles | 0.008 | process wall | pass |
| internal/fresh | TestArrayHolesResultsAreFresh | 0.008 | process wall | pass |
| internal/fresh | TestArrayRangeErrorIdentityDoesNotEscape | 0.007 | process wall | pass |
| internal/ir | TestCallTargetReaders | 7.420 | Go test leaf | pass |
| internal/ir | TestCallTargetsIncludeEveryDescendant | 0.115 | process wall | pass |
| internal/ir | TestClosureTargetsBoundOnlyProvenValues | 0.016 | process wall | pass |
| internal/ir | TestClosureTargetsPreserveBoundaryValue | 0.016 | process wall | pass |
| internal/ir | TestProduct_CallTargetExports | 2.430 | Go test leaf | pass |
| internal/javascript | TestViewArrayFieldPresence | 0.114 | process wall | pass |
| internal/javascript | TestViewArraysNode | 0.314 | process wall | pass |
| internal/javascript | TestViewCallableBoxingUnknownProducer | 0.064 | process wall | pass |
| internal/javascript | TestViewCallableProducerCertificateNode | 0.164 | process wall | pass |
| internal/javascript | TestViewCallableShapeNode | 0.315 | process wall | pass |
| internal/javascript | TestViewCallablesNode | 0.214 | process wall | pass |
| internal/lower | TestArrayMembershipUsesUnionMatcher | 0.008 | process wall | pass |
| internal/lower | TestLibraryLanguageBoundaries | 0.216 | process wall | pass |
| internal/lower | TestMixedArrayElementStorageRemainsRefused | 0.064 | process wall | pass |
| internal/lower | TestMixedUnionContractFailureDoesNotCertifyRetry | 0.064 | process wall | pass |
| internal/lower | TestMixedUnionContractGraph | 0.115 | process wall | pass |
| internal/lower | TestMixedUnionContractPhantomBrandUsesPrimitiveBase | 0.064 | process wall | pass |
| internal/lower | TestMixedUnionContractPhantomVoidIsUndefined | 0.064 | process wall | pass |
| internal/lower | TestMixedUnionContractRecursiveMember | 0.064 | process wall | pass |
| internal/lower | TestMixedUnionContractUnknownMemberFails | 0.064 | process wall | pass |
| internal/lower | TestPrepareViewCallableRead | 0.166 | process wall | pass |
| internal/lower | TestViewArrayContract | 0.164 | process wall | pass |
| internal/lower | TestViewArraySourceCertificateOmission | 0.064 | process wall | pass |
| internal/lower | TestViewArrayWritesNeedSourceCertificate | 0.064 | process wall | pass |
| internal/lower | TestViewCallableAggregateReturnedDemand | 0.008 | process wall | pass |
| internal/lower | TestViewCallableShapeContract | 0.164 | process wall | pass |
| internal/lower | TestViewCallableSignature | 0.008 | process wall | pass |
| internal/native | TestViewAdapterCacheBounded | 1.417 | process wall | pass |
| internal/native | TestViewAdapterCacheBytes | 0.616 | process wall | pass |
| internal/native | TestViewAdapterCacheCanonical | 1.469 | process wall | pass |
| internal/native | TestViewAdapterCacheDying | 1.118 | process wall | pass |
| internal/native | TestViewAdapterCacheIdentity | 1.367 | process wall | pass |
| internal/native | TestViewAdapterCacheIntern | 0.715 | process wall | pass |
| internal/native | TestViewAdapterCacheMany | 0.769 | process wall | pass |
| internal/native | TestViewAdapterCacheOrder | 1.067 | process wall | pass |
| internal/native | TestViewAdapterCacheRecheck | 1.069 | process wall | pass |
| internal/native | TestViewAdapterCacheWeakLeakRoot | 0.869 | process wall | pass |
| internal/native | TestViewAdapterPoolTSanPending | 0.016 | process wall | skip |
| internal/native | TestViewAdapterProgramUnderlyingPending | 0.007 | process wall | skip |
| internal/native | TestViewAdapterRegionAdoptionPending | 0.007 | process wall | skip |
| internal/native | TestViewAdapterRuntimeStaticsPending | 0.007 | process wall | skip |
| internal/native | TestViewCallableBoxingUnknownProducer | 1.960 | Go test leaf | pass |
| internal/native | TestViewCallableProducerCertificateNativeOptionalAbsent | 1.882 | process wall | pass |
| internal/native | TestViewCallableProducerCertificateNativeUnrecorded | 1.772 | process wall | pass |
| internal/native | TestViewCallableProducerCertificateNativeValid | 1.727 | process wall | pass |
| internal/native | TestViewCallableProducerCertificateNativeWrongArity | 1.925 | process wall | pass |
| internal/native | TestViewCallableProducerCertificateNativeWrongResult | 1.819 | process wall | pass |
| internal/native | TestViewCallableShapeNativeNumber | 0.270 | Go test leaf | pass |
| internal/native | TestViewCallableShapeNativeOptionalAbsent | 1.960 | Go test leaf | pass |
| internal/native | TestViewCallableShapeNativeRequiredAbsent | 1.950 | Go test leaf | pass |
| internal/native | TestViewCallableShapeNativeUnknownParameter | 1.970 | Go test leaf | pass |
| internal/native | TestViewCallableShapeNativeUnknownResult | 0.250 | Go test leaf | pass |
| internal/native | TestViewCallableShapeNativeUnknownSignature | 0.240 | Go test leaf | pass |
| internal/native | TestViewCallableShapeNativeValid | 0.290 | Go test leaf | pass |
| internal/native | TestViewCallableShapeNativeWrongArity | 0.280 | Go test leaf | pass |
| internal/native | TestViewCallableShapeNativeWrongParameter | 0.270 | Go test leaf | pass |
| internal/native | TestViewCallableShapeNativeWrongResult | 0.260 | Go test leaf | pass |
| internal/oracle | TestArrayElementKindRuntimeMutant | 6.959 | process wall | pass |
| internal/oracle | TestArrayHolesAbsentSlotMutant | 7.087 | process wall | pass |
| internal/oracle | TestArrayHolesMilestone | 4.604 | process wall | pass |
| internal/oracle | TestArrayHolesRangeErrorMutant | 9.220 | process wall | pass |
| internal/oracle | TestArrayViewHolesAbsenceMutant | 5.575 | process wall | pass |
| internal/oracle | TestArrayViewMissingReceiverMutant | 9.861 | process wall | pass |
| internal/oracle | TestArrayViewPhysicalWriteMutant | 10.341 | process wall | pass |
| internal/oracle | TestCheckedViewArrayReadMutantsArrayBoolean | 4.478 | process wall | pass |
| internal/oracle | TestCheckedViewArrayReadMutantsArrayIterationBad | 4.160 | process wall | pass |
| internal/oracle | TestCheckedViewArrayReadMutantsArrayMapBad | 2.825 | process wall | pass |
| internal/oracle | TestCheckedViewArrayReadMutantsArrayMissing | 3.029 | process wall | pass |
| internal/oracle | TestCheckedViewArrayReadMutantsArraySecond | 2.876 | process wall | pass |
| internal/oracle | TestCheckedViewArrayReadMutantsArrayStringLiteral | 2.974 | process wall | pass |
| internal/oracle | TestCheckedViewArrayReadMutantsArrayUndefined | 2.737 | process wall | pass |
| internal/oracle | TestCheckedViewArrayReadMutantsNonArray | 2.980 | process wall | pass |
| internal/oracle | TestCheckedViewArrayUnionMembershipBoolean | 3.083 | process wall | pass |
| internal/oracle | TestCheckedViewArrayUnionMembershipBooleanGood | 3.037 | process wall | pass |
| internal/oracle | TestCheckedViewArrayUnionMembershipBoxedMixed | 2.941 | process wall | pass |
| internal/oracle | TestCheckedViewArrayUnionMembershipMaybeNumber | 3.031 | process wall | pass |
| internal/oracle | TestCheckedViewArrayUnionMembershipNumber | 2.897 | process wall | pass |
| internal/oracle | TestCheckedViewArrayUnionMembershipSparseMax | 2.985 | process wall | pass |
| internal/oracle | TestCheckedViewArrayUnionMembershipSparseWrong | 2.883 | process wall | pass |
| internal/oracle | TestCheckedViewArrayUnionMembershipString | 2.936 | process wall | pass |
| internal/oracle | TestCheckedViewArrayUnionMembershipTupleObject | 2.925 | process wall | pass |
| internal/oracle | TestCheckedViewArrayUnionNarrowing | 2.874 | process wall | pass |
| internal/oracle | TestCheckedViewArraysArrayAlias | 2.985 | process wall | pass |
| internal/oracle | TestCheckedViewArraysArrayBoolean | 3.033 | process wall | pass |
| internal/oracle | TestCheckedViewArraysArrayBounds | 3.082 | process wall | pass |
| internal/oracle | TestCheckedViewArraysArrayElement | 2.929 | process wall | pass |
| internal/oracle | TestCheckedViewArraysArrayField | 2.937 | process wall | pass |
| internal/oracle | TestCheckedViewArraysArrayHolesView | 2.874 | process wall | pass |
| internal/oracle | TestCheckedViewArraysArrayIteration | 3.187 | process wall | pass |
| internal/oracle | TestCheckedViewArraysArrayIterationBad | 3.033 | process wall | pass |
| internal/oracle | TestCheckedViewArraysArrayLazy | 3.033 | process wall | pass |
| internal/oracle | TestCheckedViewArraysArrayLengthKind | 3.076 | process wall | pass |
| internal/oracle | TestCheckedViewArraysArrayMap | 3.278 | process wall | pass |
| internal/oracle | TestCheckedViewArraysArrayMapBad | 2.976 | process wall | pass |
| internal/oracle | TestCheckedViewArraysArrayMissing | 2.930 | process wall | pass |
| internal/oracle | TestCheckedViewArraysArrayPop | 3.530 | process wall | pass |
| internal/oracle | TestCheckedViewArraysArraySecond | 3.258 | process wall | pass |
| internal/oracle | TestCheckedViewArraysArrayStringLiteral | 3.679 | process wall | pass |
| internal/oracle | TestCheckedViewArraysArrayUndefined | 4.065 | process wall | pass |
| internal/oracle | TestCheckedViewArraysGenericArrayElement | 4.334 | process wall | pass |
| internal/oracle | TestCheckedViewArraysNonArray | 4.269 | process wall | pass |
| internal/oracle | TestCheckedViewArraysNullableArrayElement | 4.297 | process wall | pass |
| internal/oracle | TestCheckedViewArraysNullableArrayMissing | 4.292 | process wall | pass |
| internal/oracle | TestCheckedViewArraysNullableArrayString | 4.340 | process wall | pass |
| internal/oracle | TestCheckedViewArraysObjectElement | 4.521 | process wall | pass |
| internal/oracle | TestCheckedViewObjectPrimitiveSource | 8.155 | process wall | pass |
| internal/oracle | TestCheckedViewUntaggedArraySource | 5.286 | process wall | pass |
| internal/oracle | TestCheckedViewUntaggedCallableUnion | 4.531 | process wall | pass |
| internal/oracle | TestCheckedViewUntaggedOptionalCallableControl | 2.985 | process wall | pass |
| internal/oracle | TestCheckedViewUntaggedOwnClassData | 2.885 | process wall | pass |
| internal/oracle | TestCheckedViewUntaggedRecursive | 2.886 | process wall | pass |
| internal/oracle | TestCheckedViewUntaggedSourceDispatch | 3.795 | process wall | pass |
| internal/oracle | TestCheckedViewUntaggedSourceFlows | 3.816 | process wall | pass |
| internal/oracle | TestCheckedViewV2ArrayArmAdmission | 2.888 | process wall | pass |
| internal/oracle | TestCheckedViewV2CallableProducerMutant | 3.088 | process wall | pass |
| internal/oracle | TestCheckedViewV2MembershipMutant | 2.880 | process wall | pass |
| internal/oracle | TestCheckedViewV2MovedStage3Results | 3.837 | process wall | pass |
| internal/oracle | TestCheckedViewV2ReadsAfterWrites | 8.886 | process wall | pass |
| internal/oracle | TestCheckedViewV2RepresentationMutants | 2.698 | process wall | pass |
| internal/oracle | TestCheckedViewV2TupleIdentityMutant | 3.336 | process wall | pass |
| internal/oracle | TestCheckedViewV2WrongFamilyMutant | 3.336 | process wall | pass |
| internal/oracle | TestNarrowedUnionMemberCheck | 4.351 | process wall | pass |
| internal/oracle | TestNarrowedUnionObjectTagMisfit | 3.508 | process wall | pass |
| internal/oracle | TestNarrowedUnionObjectTagPassing | 3.503 | process wall | pass |
| internal/oracle | TestOriginalArrayWitnessExtensions | 0.036 | process wall | pass |
| internal/oracle | TestOriginalArrayWitnessNode | 40.172 | process wall | pass |
| internal/oracle | TestProduct_Adamic | 5.052 | process wall | pass |
| internal/oracle | TestProduct_NativeRuntimesCounted0 | 6.672 | process wall | pass |
| internal/oracle | TestProduct_NativeRuntimesCounted1 | 6.781 | process wall | pass |
| internal/oracle | TestProduct_NativeRuntimesCounted2 | 5.273 | process wall | pass |
| internal/oracle | TestProduct_NativeRuntimesRelease0 | 6.766 | process wall | pass |
| internal/oracle | TestProduct_NativeRuntimesRelease1 | 7.064 | process wall | pass |
| internal/oracle | TestProduct_NativeRuntimesRelease2 | 2.560 | Go test leaf | pass |
| internal/oracle | TestProduct_NativeRuntimesSanitized0 | 10.460 | process wall | pass |
| internal/oracle | TestProduct_NativeRuntimesSanitized1 | 10.255 | process wall | pass |
| internal/oracle | TestProduct_NativeRuntimesSanitized2 | 7.863 | process wall | pass |
| internal/oracle | TestProduct_V4SanitizedCountedRuntimes0 | 9.831 | process wall | pass |
| internal/oracle | TestProduct_V4SanitizedCountedRuntimes1 | 10.103 | process wall | pass |
| internal/oracle | TestProduct_V4SanitizedCountedRuntimes2 | 2.460 | Go test leaf | pass |
| internal/oracle | TestRuntimeProductKeyHoldsEveryPart | 7.192 | process wall | pass |
| internal/oracle | TestRuntimeSetsAreEachUsed | 0.750 | Go test leaf | pass |
| internal/oracle | TestV4DestructuredCallableAdamicRefusal | 3.130 | process wall | pass |
| internal/oracle | TestV4DestructuredCallableArgument | 6.716 | process wall | pass |
| internal/oracle | TestV4DestructuredCallableReadOnly | 5.374 | process wall | pass |
| internal/oracle | TestV4DirectArgument | 3.751 | process wall | pass |
| internal/oracle | TestV4DirectArrayArgument | 3.993 | process wall | pass |
| internal/oracle | TestV4DirectBoxed | 3.644 | process wall | pass |
| internal/oracle | TestV4DirectCalleeOrder | 0.300 | Go test leaf | pass |
| internal/oracle | TestV4DirectClassResult | 3.536 | process wall | pass |
| internal/oracle | TestV4DirectCompatible | 3.559 | process wall | pass |
| internal/oracle | TestV4DirectMethod | 3.537 | process wall | pass |
| internal/oracle | TestV4DirectNull | 3.639 | process wall | pass |
| internal/oracle | TestV4DirectObjectArgument | 3.843 | process wall | pass |
| internal/oracle | TestV4DirectOptionalReceiver | 3.561 | process wall | pass |
| internal/oracle | TestV4DirectProducerRefusal | 0.240 | Go test leaf | pass |
| internal/oracle | TestV4DirectRecursiveRefusal | 0.130 | Go test leaf | pass |
| internal/oracle | TestV4DirectResult | 4.151 | process wall | pass |
| internal/oracle | TestV4DirectUnknownProducerParameter | 3.909 | process wall | pass |
| internal/oracle | TestV4DirectUnused | 3.760 | process wall | pass |
| internal/oracle | TestV4DirectVoid | 3.994 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterApplyArgument | 6.682 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterApplyResult | 6.761 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterArgument | 6.429 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterBindAllArguments | 5.542 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterBindBoxed | 5.404 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterBindCompatible | 5.458 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterBindCountAndSurface | 5.457 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterBindNameOwnership | 5.350 | Go test leaf | pass |
| internal/oracle | TestV4EscapeAdapterBindPrefix | 6.129 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterBindReceiverAndCapture | 5.578 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterBindRemainingArgument | 6.061 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterBindResult | 6.012 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterBlameApply | 6.122 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterBlameArgument | 5.953 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterBlameBound | 5.975 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterBlameCall | 5.929 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterBlameNestedResult | 5.606 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterBlameResult | 5.918 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterBlameRuntimeCallback | 0.016 | process wall | skip |
| internal/oracle | TestV4EscapeAdapterBoundAliasCallRefusal | 3.453 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterBoundApplyRefusal | 3.426 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterBoundBindRefusal | 3.337 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterBoundCallRefusal | 3.337 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterBoundHelperApplyRefusal | 3.192 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterBoundHelperBindRefusal | 3.265 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterBoundViewRefusal | 3.454 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterBoxed | 5.680 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterCallApplyCount | 6.059 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterCallApplyReceiver | 5.423 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterCallApplyZeroReceiver | 5.908 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterCallArgument | 6.539 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterCallReceiverCheck | 7.261 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterCallReceiverRefusal | 3.435 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterCallRefusal | 3.138 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterCallResult | 7.130 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterCollectionKeyContract | 5.258 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterConstructSignatureRefusal | 3.138 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterConstructorIdentity | 3.291 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterCounted | 5.354 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterDynamicApplyRefusal | 3.240 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterIdentity | 5.600 | Go test leaf | pass |
| internal/oracle | TestV4EscapeAdapterLength | 9.304 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterName | 5.580 | Go test leaf | pass |
| internal/oracle | TestV4EscapeAdapterNoStacking | 8.709 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterNull | 5.261 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterNullablePrimitiveCallRefusal | 3.083 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterPrimitiveApplyRefusal | 3.097 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterPrimitiveCallRefusal | 3.092 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterResult | 5.528 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterReturned | 5.256 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterScalarIdentityControls | 3.885 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterSurfaceAnonymousName | 8.543 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterSurfaceComputedName | 8.214 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterSurfaceDefaults | 8.301 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterSurfaceLengthRefusal | 3.348 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterSurfaceNameRefusal | 3.447 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterSurfaceRegExpCallback | 13.570 | Go test leaf | pass |
| internal/oracle | TestV4EscapeAdapterSurfaceReturned | 8.028 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterSurfaceSortCallback | 5.460 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterUncheckableRefusal | 3.609 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterUnusedMisfit | 8.614 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterVoid | 8.909 | process wall | pass |
| internal/oracle | TestV4EscapeAdapterWithoutConstructSignatureRefusal | 0.090 | Go test leaf | pass |
| internal/oracle | TestV4EscapeArgumentRefusal | 0.120 | Go test leaf | pass |
| internal/oracle | TestV4EscapeCompatible | 5.547 | process wall | pass |
| internal/oracle | TestV4EscapeMixedProducerRefusal | 0.150 | Go test leaf | pass |
| internal/oracle | TestV4EscapeOrdinaryRead | 4.802 | process wall | pass |
| internal/oracle | TestV4EscapePassedRefusal | 3.200 | Go test leaf | pass |
| internal/oracle | TestV4EscapeResultRefusal | 0.130 | Go test leaf | pass |
| internal/oracle | TestV4EscapeReturnedRefusal | 0.130 | Go test leaf | pass |
| internal/oracle | TestV4EscapeTypeScriptBoundary | 5.582 | process wall | pass |
| internal/oracle | TestV4ExistingExplicitMethodReceiver | 5.282 | process wall | pass |
| internal/oracle | TestV4GenericErasedReadRefusal | 3.192 | process wall | pass |
| internal/oracle | TestV4GenericProducerInstantiation | 5.624 | process wall | pass |
| internal/oracle | TestV4GenericViewCallInstantiation | 3.941 | process wall | pass |
| internal/oracle | TestV4HigherAdamicEscapeRefusal | 3.285 | process wall | pass |
| internal/oracle | TestV4HigherCallbackArgument | 5.720 | process wall | pass |
| internal/oracle | TestV4HigherCallbackArgumentOrder | 5.424 | process wall | pass |
| internal/oracle | TestV4HigherCallbackNeverCalled | 5.500 | process wall | pass |
| internal/oracle | TestV4HigherCallbackResult | 5.823 | process wall | pass |
| internal/oracle | TestV4HigherReturnedCallable | 5.585 | process wall | pass |
| internal/oracle | TestV4IteratorPackedReceiver | 5.568 | process wall | pass |
| internal/oracle | TestV4Lane5BoxedAdamicRefusal | 2.450 | Go test leaf | pass |
| internal/oracle | TestV4Lane5BoxedParameter | 5.884 | process wall | pass |
| internal/oracle | TestV4Lane5BoxedParameterMembers | 6.116 | process wall | pass |
| internal/oracle | TestV4Lane5BoxedResult | 5.760 | process wall | pass |
| internal/oracle | TestV4Lane5BoxedResultMembers | 5.693 | process wall | pass |
| internal/oracle | TestV4Lane5FactoryClone | 4.310 | process wall | pass |
| internal/oracle | TestV4Lane5FactoryIdentifier | 4.159 | process wall | pass |
| internal/oracle | TestV4Lane5FactoryImport | 4.326 | process wall | pass |
| internal/oracle | TestV4Lane5FactoryModifier | 4.103 | process wall | pass |
| internal/oracle | TestV4Lane5FactoryReadonlyTargetControl | 0.008 | process wall | skip |
| internal/oracle | TestV4Lane5FactoryResult | 3.728 | process wall | pass |
| internal/oracle | TestV4Lane5FactoryStringLiteral | 4.094 | process wall | pass |
| internal/oracle | TestV4Lane5FactoryUniqueName | 4.166 | process wall | pass |
| internal/oracle | TestV4Lane5FactoryWiderWritePending | 0.008 | process wall | skip |
| internal/oracle | TestV4Lane5FactoryWiderWriteRefusal | 2.460 | Go test leaf | pass |
| internal/oracle | TestV4Lane5FactoryYield | 4.044 | process wall | pass |
| internal/oracle | TestV4Lane5SkipAdapterMutant | 6.690 | process wall | pass |
| internal/oracle | TestV4MixedCallableFieldArgument | 3.756 | process wall | pass |
| internal/oracle | TestV4MixedCallableFieldBoxing | 3.440 | process wall | pass |
| internal/oracle | TestV4MixedLongFieldDiagnostic | 3.398 | process wall | pass |
| internal/oracle | TestV4UnrelatedArrayFieldWrite | 4.149 | process wall | pass |
| internal/oracle | TestViewAdapterBaselineCountsClosures | 3.372 | process wall | pass |
| internal/oracle | TestViewAdapterBaselineCountsMethods | 3.299 | process wall | pass |
| internal/oracle | TestViewAdapterBaselineCountsReceiverRest | 3.797 | process wall | pass |
| stage1/cohere/json | TestPortElementMetadataEmission | 0.990 | Go test leaf | pass |

Independent source and witness mutant catchers:

| Mutant | Catcher test | Seconds | Observation |
|---|---|---:|---|
| escape-Argument | TestV4EscapeArgumentRefusal | 23.889 | exit 1; missing escaping-read refusal after clean Node-equivalent execution |
| escape-Result | TestV4EscapeResultRefusal | 6.303 | exit 1; missing escaping-read refusal after clean Node-equivalent execution |
| escape-Passed | TestV4EscapePassedRefusal | 6.073 | exit 1; missing escaping-read refusal after clean Node-equivalent execution |
| escape-Returned | TestV4EscapeReturnedRefusal | 6.103 | exit 1; missing escaping-read refusal after clean Node-equivalent execution |
| escape-MixedProducer | TestV4EscapeMixedProducerRefusal | 6.168 | exit 1; missing escaping-read refusal after clean Node-equivalent execution |
| unsupported-domain | TestV4DirectRecursiveRefusal | 16.639 | exit 1; recursive domain refusal missing |
| unsupported-producer | TestV4DirectProducerRefusal | 15.988 | exit 1; recursive producer refusal missing |
| surface-borrow | TestV4EscapeAdapterName | 12.228 | exit 1; ASan heap-use-after-free in adapter root access |
| bind-name-borrow | TestV4EscapeAdapterBindNameOwnership | 12.627 | exit 1; ASan heap-use-after-free in string output |
| identity-root | TestV4EscapeAdapterIdentity | 33.983 | exit 1; Node stdout comparison, clean exit 0 |
| metadata-inline | TestPortElementMetadataEmission | 10.011 | exit 1; 2,885 inline stores violate the zero-store bound |
| metadata-omitted | TestPortElementMetadataEmission | 9.244 | exit 1; zero runtime calls violate the 2,885-certificate assertion |
| native-order | TestV4DirectCalleeOrder | 11.051 | exit 1; argument output missing; premature exit 70 |
| javascript-order | TestV4DirectCalleeOrder | 11.739 | exit 1; argument output missing; premature exit 70 |
| construct-no-signature | TestV4EscapeAdapterWithoutConstructSignatureRefusal | 3.393 | exit 1; checker refusal needs path and fix |
| view-result-wrong | TestV4Lane5BoxedAdamicRefusal | 5.892 | exit 1; .a keeps the unproven-read refusal |
| factory-165-wrong-arity | TestV4Lane5FactoryWiderWriteRefusal | 5.804 | exit 1; wider-write keeps path and fix |

All 17 mutants restored exact source bytes. None is counted as a build-warning kill. The original JavaScript-order script's mutation seam had moved; its blocked attempt is retained, and javascript-order-fixed.patch targets the current direct-call emitter. Source patches are .patch, never compilable Go. Embedded mutant observations, including native and JavaScript argument/result, aggregate, array-presence/storage/literal/hole, identity/surface, and adapter-cache controls, are collected in mutant-observations.txt with the full per-test logs beside it. Runtime adapter mutants include skip-intern (allocation bound), skip-recheck (winner selection), dying-retain (ASan use-after-free), release-before-remove (lifecycle order), canonical-factory (required exit 70), strong-leak-root (missing required LSan report), and identity-overread (ASan global-buffer-overflow).

Carried pending coverage: TestViewAdapterRuntimeStaticsPending, TestViewAdapterPoolTSanPending, TestViewAdapterRegionAdoptionPending, TestViewAdapterProgramUnderlyingPending, TestV4EscapeAdapterBlameRuntimeCallback, TestV4Lane5FactoryWiderWritePending and TestV4Lane5FactoryReadonlyTargetControl. No whole-gate, ThreadSanitizer pool, region adoption or pending higher-order control is claimed.

Setup: initial cloud/setup.sh was stopped with exit 143 while Git's nested historical lookup fetched over 550 MB. An explicit depth-1, blob-filtered fetch of d92d9bfee114c80be2c375d72edae966176e3a4f and detached checkout repaired the submodule setup; setup was rerun successfully. GOPROXY=https://proxy.golang.org|direct and /workspace/adamic-tools/env.sh were used. Setup timing lines:

```text
setup: node ready (0.022s)
setup: go ready (0.024s)
setup: submodules ready (0.063s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.009s
setup: markdown dependencies ready (0.073s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.151s)
setup: shared cache ready (5.151s)
setup: go build ready (322.790s)
setup: test binaries deferred (use --warm-tests) (322.865s)
setup: build cache warm (322.866s)
setup: build-flags commit=cd7880cf5145b33a5bf6bb2227e216d7fb9d63c0 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.06 0.03 0.01 1/152 1553 load-after=9.65 5.89 2.41 1/190 10845
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (322.893s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.ObeLJB
```
