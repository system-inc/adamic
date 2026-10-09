From the test audit (#xphstyt, #6skgn8x). @system_adamic's bar for internal/oracle: a test leaves only if each of its mutants is shown caught by another test on the same fixture class.

How this was settled:
- The set replay (test-defend/deletion-set/internal-oracle a8add762, base 7b9d4272, 77 mutant records) stopped each mutant at its first catcher. Identical diffs got different first catchers.
- rp:oracle-full (test-defend/deletion-set/internal-oracle-full a774f430) reran the 19 doubtful mutants over the whole package with every catcher recorded, with TestCountsAreRecorded and every *Mutant(s) and Witness test skipped. Per-subtest lists are in failure-details.json.
- Key fact: every candidate below runs fixtures that are also in the shared corpus (oracle_test.go's fixtures list, the counts.md rows) that TestNativeAgreesWithNode holds to Node. Where the corpus test is the catcher, it caught the mutant on those same fixtures. Examples: D02 module_namespace.go:16 on module_namespace_reads/*.a, D3 enum_never.go:100 on enums_open_never_if and _index, M05 namespace_callable.go:21 on namespace_callable_properties.a.

Deletable under the bar (11):
- TestModuleNamespaceReadsMatchNode, the TestParserNamespace family (Receiver, Class, Callable), TestNumericEnumNeverPathsPinned and TestNumericEnumNeverPinned: each mutant is caught by the corpus test on the candidate's own fixtures.
- TestCheckedCastFlushesOutput and TestCheckedCastRunsNoCatchOrFinally: adamic.c:195 is also caught by TestCheckedCastFailureContract and TestInterfaceCastOracle, and the cast_proof mutants by TestUncheckableCastAdmission.
- TestCheckedViewUntaggedSourceFlows and TestCheckedViewV2MovedStage3Results: caught by the CheckedViewUntagged* and CheckedViewFlagDowncast tests.
- TestFileWritesLandInNodesOrder: lower.go:22 and expression.go:1024 are caught by TestOneFileHoldsNodesOrder, and input.c:403 by TestInputAgreesWithNode.
- The TestCallTarget family: caught by the corpus test and TestReviewProgramsAgreeWithNode.
- TestOmittedOriginalProbePolicy: non_null.go:106 and emit_values.go:61 are caught by the CheckedNonNull* tests and TestViewFieldInheritedStaticReadiness.

Your judgment, not a recommendation (3):
- TestImportCycleRuntimeCalls: emit.go:106 is caught in-class by TestImportCycleLoadTimeReads. D5 expression.go:784, D6 emit_expressions.go:754 and D7 emit_branches.go:104 are caught on 22 to 193 corpus fixtures, but none of them is an import-cycle program. Its fixture, import_cycles/runtime, is in the corpus.
- TestNamespaceLiveExportBoundary: stage3/namespace-live-export/live.a is in the corpus, and all its mutants are caught on namespace fixtures. It also runs the JavaScript backend and a leak check on that fixture.
- TestRegexCycleFixtureHasItsNativeDependency: no mutant reaches it. It guards fixture registration, not behavior.

Keep (5):
- TestWASIEmission: the only wasm32-wasi compile of every fixture's emitted C (when ADAMIC_ORACLE_WASI=1), a different target from the corpus test.
- TestCheckedViewV2ArrayArmBoundary and TestInterfaceCastImportedConstruction: cast.go:33's D5 has no clean catcher without them (25 panic exclusions, last failure only a witness).
- TestStage3EnumSparseArrayBoundary and TestTypedArrayWriteStopIsPinned: typed_array.c:101 (D4) and javascript.go:91 (D2) lose every catcher without both. Keep at least one, your pick.

Land deletions test-only, after stage 3 opens (#k0ekyt8). Prove nothing breaks with internal/oracle's own run.
