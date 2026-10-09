Two previously untrue rows have a unique production kill in the bounded matrix.
The moved stage-3 row was not defended by three valid targeted attempts.
Evidence is committed on test-defend/internal-oracle-checked_views_flag_downcast; package-wide uniqueness remains bounded.

[
  {
    "test": "TestCheckedViewV2TupleIdentityMutant",
    "package": "internal/oracle",
    "prior_verdict": "untrue",
    "subsumed_by": [],
    "defense": "defended",
    "unique_mutant": "D1 internal/javascript/view_unions_untagged.go:87",
    "attempts": [
      {
        "mutant": "D1",
        "file_line": "internal/javascript/view_unions_untagged.go:87",
        "change": "drop fixed tuple array identity and length guard",
        "rows_failed": [
          "TestCheckedViewV2TupleIdentityMutant"
        ]
      }
    ],
    "evidence": "ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/checked-views-defense/cache/D1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^(Test.*View.*|TestInterfaceCast.*|TestCheckedCast.*|TestUncheckableCastAdmission|TestNarrowedUnion.*|TestNarrowedFieldUsesSharedReadiness)$ > D1-0.log 2>&1; checked_views_v2_migration_test.go:200: stderr differs",
    "bounded": true,
    "rows_passed": [
      "TestCheckedCastFailureContract",
      "TestCheckedCastFlushesOutput",
      "TestCheckedCastMutants",
      "TestCheckedCastRunsNoCatchOrFinally",
      "TestCheckedViewFlagDowncast",
      "TestCheckedViewInterfaces",
      "TestCheckedViewObjectPrimitiveSource",
      "TestCheckedViewObjects",
      "TestCheckedViewOptionalReadBoundary",
      "TestCheckedViewUntaggedArrayPending",
      "TestCheckedViewUntaggedCallableUnion",
      "TestCheckedViewUntaggedOptionalCallableControl",
      "TestCheckedViewUntaggedOwnClassData",
      "TestCheckedViewUntaggedRecursive",
      "TestCheckedViewUntaggedSourceDispatch",
      "TestCheckedViewUntaggedSourceFlows",
      "TestCheckedViewV2ArrayArmBoundary",
      "TestCheckedViewV2CallableProducerMutant",
      "TestCheckedViewV2MembershipMutant",
      "TestCheckedViewV2MovedStage3Results",
      "TestCheckedViewV2ReadsAfterWrites",
      "TestCheckedViewV2RepresentationMutants",
      "TestCheckedViewV2WrongFamilyMutant",
      "TestDefaultTaggedSourceViews",
      "TestFractionalPowersReachRuntime",
      "TestInUnionNarrowingIsRefused",
      "TestInterfaceCastChecksMalformedRead",
      "TestInterfaceCastImportedConstruction",
      "TestInterfaceCastOracle",
      "TestInterfaceCastRuntimeMutants",
      "TestInterfaceCastScalarTags",
      "TestLiteralOptionalOracleCatchesMutant",
      "TestLiteralUndefinedOracleCatchesMutant",
      "TestNarrowedFieldUsesSharedReadiness",
      "TestNarrowedUnionMemberCheck",
      "TestNarrowedUnionObjectTagRefusal",
      "TestNativeAgreesWithNode",
      "TestNonNullLiteralUnionInitializers",
      "TestParseIntMapIndexRadixAgreesWithNode",
      "TestPredicateDirectionCountsAreRecorded",
      "TestPredicateMiscompileRefusals",
      "TestRequiredViewFieldOperandOnce",
      "TestRequiredViewFieldPrimitive",
      "TestReviewProgramsAgreeWithNode",
      "TestReviewProgramsNoLooseFiles",
      "TestReviewProgramsRefuse",
      "TestReviewProgramsSelfTest",
      "TestTypeOfStringLiteralMutant",
      "TestUncheckableCastAdmission",
      "TestUnknownNarrowingMutants",
      "TestViewFieldInheritedStaticReadiness"
    ],
    "matrix_rows": [
      "TestCheckedCastFailureContract",
      "TestCheckedCastFlushesOutput",
      "TestCheckedCastMutants",
      "TestCheckedCastRunsNoCatchOrFinally",
      "TestCheckedViewFlagDowncast",
      "TestCheckedViewInterfaces",
      "TestCheckedViewObjectPrimitiveSource",
      "TestCheckedViewObjects",
      "TestCheckedViewOptionalReadBoundary",
      "TestCheckedViewUntaggedArrayPending",
      "TestCheckedViewUntaggedCallableUnion",
      "TestCheckedViewUntaggedOptionalCallableControl",
      "TestCheckedViewUntaggedOwnClassData",
      "TestCheckedViewUntaggedRecursive",
      "TestCheckedViewUntaggedSourceDispatch",
      "TestCheckedViewUntaggedSourceFlows",
      "TestCheckedViewV2ArrayArmBoundary",
      "TestCheckedViewV2CallableProducerMutant",
      "TestCheckedViewV2MembershipMutant",
      "TestCheckedViewV2MovedStage3Results",
      "TestCheckedViewV2ReadsAfterWrites",
      "TestCheckedViewV2RepresentationMutants",
      "TestCheckedViewV2TupleIdentityMutant",
      "TestCheckedViewV2WrongFamilyMutant",
      "TestDefaultTaggedSourceViews",
      "TestFractionalPowersReachRuntime",
      "TestInUnionNarrowingIsRefused",
      "TestInterfaceCastChecksMalformedRead",
      "TestInterfaceCastImportedConstruction",
      "TestInterfaceCastOracle",
      "TestInterfaceCastRuntimeMutants",
      "TestInterfaceCastScalarTags",
      "TestLiteralOptionalOracleCatchesMutant",
      "TestLiteralUndefinedOracleCatchesMutant",
      "TestNarrowedFieldUsesSharedReadiness",
      "TestNarrowedUnionMemberCheck",
      "TestNarrowedUnionObjectTagRefusal",
      "TestNativeAgreesWithNode",
      "TestNonNullLiteralUnionInitializers",
      "TestParseIntMapIndexRadixAgreesWithNode",
      "TestPredicateDirectionCountsAreRecorded",
      "TestPredicateMiscompileRefusals",
      "TestRequiredViewFieldOperandOnce",
      "TestRequiredViewFieldPrimitive",
      "TestReviewProgramsAgreeWithNode",
      "TestReviewProgramsNoLooseFiles",
      "TestReviewProgramsRefuse",
      "TestReviewProgramsSelfTest",
      "TestTypeOfStringLiteralMutant",
      "TestUncheckableCastAdmission",
      "TestUnknownNarrowingMutants",
      "TestViewFieldInheritedStaticReadiness"
    ]
  },
  {
    "test": "TestCheckedViewV2CallableProducerMutant",
    "package": "internal/oracle",
    "prior_verdict": "untrue",
    "subsumed_by": [],
    "defense": "defended",
    "unique_mutant": "D2 internal/javascript/view_unions_untagged.go:153",
    "attempts": [
      {
        "mutant": "D2",
        "file_line": "internal/javascript/view_unions_untagged.go:153",
        "change": "flip callable producer certification bypass",
        "rows_failed": [
          "TestCheckedViewV2CallableProducerMutant"
        ]
      }
    ],
    "evidence": "ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/checked-views-defense/cache/D2 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^(Test.*View.*|TestInterfaceCast.*|TestCheckedCast.*|TestUncheckableCastAdmission|TestNarrowedUnion.*|TestNarrowedFieldUsesSharedReadiness)$ > D2-0.log 2>&1; checked_views_v2_migration_test.go:231: exit codes differ",
    "bounded": true,
    "rows_passed": [
      "TestCheckedCastFailureContract",
      "TestCheckedCastFlushesOutput",
      "TestCheckedCastMutants",
      "TestCheckedCastRunsNoCatchOrFinally",
      "TestCheckedViewFlagDowncast",
      "TestCheckedViewInterfaces",
      "TestCheckedViewObjectPrimitiveSource",
      "TestCheckedViewObjects",
      "TestCheckedViewOptionalReadBoundary",
      "TestCheckedViewUntaggedArrayPending",
      "TestCheckedViewUntaggedCallableUnion",
      "TestCheckedViewUntaggedOptionalCallableControl",
      "TestCheckedViewUntaggedOwnClassData",
      "TestCheckedViewUntaggedRecursive",
      "TestCheckedViewUntaggedSourceDispatch",
      "TestCheckedViewUntaggedSourceFlows",
      "TestCheckedViewV2ArrayArmBoundary",
      "TestCheckedViewV2MembershipMutant",
      "TestCheckedViewV2MovedStage3Results",
      "TestCheckedViewV2ReadsAfterWrites",
      "TestCheckedViewV2RepresentationMutants",
      "TestCheckedViewV2TupleIdentityMutant",
      "TestCheckedViewV2WrongFamilyMutant",
      "TestDefaultTaggedSourceViews",
      "TestFractionalPowersReachRuntime",
      "TestInUnionNarrowingIsRefused",
      "TestInterfaceCastChecksMalformedRead",
      "TestInterfaceCastImportedConstruction",
      "TestInterfaceCastOracle",
      "TestInterfaceCastRuntimeMutants",
      "TestInterfaceCastScalarTags",
      "TestLiteralOptionalOracleCatchesMutant",
      "TestLiteralUndefinedOracleCatchesMutant",
      "TestNarrowedFieldUsesSharedReadiness",
      "TestNarrowedUnionMemberCheck",
      "TestNarrowedUnionObjectTagRefusal",
      "TestNativeAgreesWithNode",
      "TestNonNullLiteralUnionInitializers",
      "TestParseIntMapIndexRadixAgreesWithNode",
      "TestPredicateDirectionCountsAreRecorded",
      "TestPredicateMiscompileRefusals",
      "TestRequiredViewFieldOperandOnce",
      "TestRequiredViewFieldPrimitive",
      "TestReviewProgramsAgreeWithNode",
      "TestReviewProgramsNoLooseFiles",
      "TestReviewProgramsRefuse",
      "TestReviewProgramsSelfTest",
      "TestTypeOfStringLiteralMutant",
      "TestUncheckableCastAdmission",
      "TestUnknownNarrowingMutants",
      "TestViewFieldInheritedStaticReadiness"
    ],
    "matrix_rows": [
      "TestCheckedCastFailureContract",
      "TestCheckedCastFlushesOutput",
      "TestCheckedCastMutants",
      "TestCheckedCastRunsNoCatchOrFinally",
      "TestCheckedViewFlagDowncast",
      "TestCheckedViewInterfaces",
      "TestCheckedViewObjectPrimitiveSource",
      "TestCheckedViewObjects",
      "TestCheckedViewOptionalReadBoundary",
      "TestCheckedViewUntaggedArrayPending",
      "TestCheckedViewUntaggedCallableUnion",
      "TestCheckedViewUntaggedOptionalCallableControl",
      "TestCheckedViewUntaggedOwnClassData",
      "TestCheckedViewUntaggedRecursive",
      "TestCheckedViewUntaggedSourceDispatch",
      "TestCheckedViewUntaggedSourceFlows",
      "TestCheckedViewV2ArrayArmBoundary",
      "TestCheckedViewV2CallableProducerMutant",
      "TestCheckedViewV2MembershipMutant",
      "TestCheckedViewV2MovedStage3Results",
      "TestCheckedViewV2ReadsAfterWrites",
      "TestCheckedViewV2RepresentationMutants",
      "TestCheckedViewV2TupleIdentityMutant",
      "TestCheckedViewV2WrongFamilyMutant",
      "TestDefaultTaggedSourceViews",
      "TestFractionalPowersReachRuntime",
      "TestInUnionNarrowingIsRefused",
      "TestInterfaceCastChecksMalformedRead",
      "TestInterfaceCastImportedConstruction",
      "TestInterfaceCastOracle",
      "TestInterfaceCastRuntimeMutants",
      "TestInterfaceCastScalarTags",
      "TestLiteralOptionalOracleCatchesMutant",
      "TestLiteralUndefinedOracleCatchesMutant",
      "TestNarrowedFieldUsesSharedReadiness",
      "TestNarrowedUnionMemberCheck",
      "TestNarrowedUnionObjectTagRefusal",
      "TestNativeAgreesWithNode",
      "TestNonNullLiteralUnionInitializers",
      "TestParseIntMapIndexRadixAgreesWithNode",
      "TestPredicateDirectionCountsAreRecorded",
      "TestPredicateMiscompileRefusals",
      "TestRequiredViewFieldOperandOnce",
      "TestRequiredViewFieldPrimitive",
      "TestReviewProgramsAgreeWithNode",
      "TestReviewProgramsNoLooseFiles",
      "TestReviewProgramsRefuse",
      "TestReviewProgramsSelfTest",
      "TestTypeOfStringLiteralMutant",
      "TestUncheckableCastAdmission",
      "TestUnknownNarrowingMutants",
      "TestViewFieldInheritedStaticReadiness"
    ]
  },
  {
    "test": "TestCheckedViewV2MovedStage3Results",
    "package": "internal/oracle",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestCheckedViewV2ReadsAfterWrites"
    ],
    "defense": "not defended",
    "unique_mutant": null,
    "attempts": [
      {
        "mutant": "D3b",
        "file_line": "internal/lower/view_objects.go:35",
        "change": "flip structural cast receiver admission condition",
        "rows_failed": [
          "TestCheckedViewObjectPrimitiveSource"
        ]
      },
      {
        "mutant": "D4",
        "file_line": "internal/javascript/javascript.go:688",
        "change": "change bitmask AND operator to OR",
        "rows_failed": [
          "TestCheckedViewFlagDowncast",
          "TestCheckedViewV2MovedStage3Results",
          "TestNativeAgreesWithNode"
        ]
      },
      {
        "mutant": "D5",
        "file_line": "internal/native/runtime/view_unions_untagged.c:141",
        "change": "change open kind selector number storage option to boolean",
        "rows_failed": [
          "TestCheckedViewUntaggedOwnClassData"
        ]
      }
    ],
    "evidence": "ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/checked-views-defense/cache/D4 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^(Test.*View.*|TestInterfaceCast.*|TestCheckedCast.*|TestUncheckableCastAdmission|TestNarrowedUnion.*|TestNarrowedFieldUsesSharedReadiness)$ > D4-0.log 2>&1; checked_views_v2_migration_test.go:273: javascript: exit codes differ; oracle.run{stdout:[]uint8{0x74, 0x72, 0x75, 0x65, 0xa, 0x74, 0x72, 0x75, 0x65, 0xa, 0x6f, 0x75, 0x74, 0x65, 0x72, 0xa, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6e, 0x67, 0xa}, stderr:[]uint8{0x61, 0x64, 0x61, 0x6d, 0x69, 0x63, 0x3a, 0x20, 0x70, 0x61, 0x6e, 0x69, 0x63, 0x3a, 0x20, 0x63, 0x61, 0x73, 0x74, 0x20, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x3a, 0x20, 0x74, 0x68, 0x69, 0x73, 0x20, 0x54, 0x79, 0x70, 0x65, 0x20, 0x69, 0x73, 0x20, 0x6e, 0x6f, 0x74, 0x20, 0x61, 0x20, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x65, 0x64, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73, 0x54, 0x79, 0x70, 0x65, 0xa}, exitCode:70}",
    "bounded": true,
    "rows_passed": [],
    "matrix_rows": [
      "TestCheckedCastFailureContract",
      "TestCheckedCastFlushesOutput",
      "TestCheckedCastMutants",
      "TestCheckedCastRunsNoCatchOrFinally",
      "TestCheckedViewFlagDowncast",
      "TestCheckedViewInterfaces",
      "TestCheckedViewObjectPrimitiveSource",
      "TestCheckedViewObjects",
      "TestCheckedViewOptionalReadBoundary",
      "TestCheckedViewUntaggedArrayPending",
      "TestCheckedViewUntaggedCallableUnion",
      "TestCheckedViewUntaggedOptionalCallableControl",
      "TestCheckedViewUntaggedOwnClassData",
      "TestCheckedViewUntaggedRecursive",
      "TestCheckedViewUntaggedSourceDispatch",
      "TestCheckedViewUntaggedSourceFlows",
      "TestCheckedViewV2ArrayArmBoundary",
      "TestCheckedViewV2CallableProducerMutant",
      "TestCheckedViewV2MembershipMutant",
      "TestCheckedViewV2MovedStage3Results",
      "TestCheckedViewV2ReadsAfterWrites",
      "TestCheckedViewV2RepresentationMutants",
      "TestCheckedViewV2TupleIdentityMutant",
      "TestCheckedViewV2WrongFamilyMutant",
      "TestDefaultTaggedSourceViews",
      "TestFractionalPowersReachRuntime",
      "TestInUnionNarrowingIsRefused",
      "TestInterfaceCastChecksMalformedRead",
      "TestInterfaceCastImportedConstruction",
      "TestInterfaceCastOracle",
      "TestInterfaceCastRuntimeMutants",
      "TestInterfaceCastScalarTags",
      "TestLiteralOptionalOracleCatchesMutant",
      "TestLiteralUndefinedOracleCatchesMutant",
      "TestNarrowedFieldUsesSharedReadiness",
      "TestNarrowedUnionMemberCheck",
      "TestNarrowedUnionObjectTagRefusal",
      "TestNativeAgreesWithNode",
      "TestNonNullLiteralUnionInitializers",
      "TestParseIntMapIndexRadixAgreesWithNode",
      "TestPredicateDirectionCountsAreRecorded",
      "TestPredicateMiscompileRefusals",
      "TestRequiredViewFieldOperandOnce",
      "TestRequiredViewFieldPrimitive",
      "TestReviewProgramsAgreeWithNode",
      "TestReviewProgramsNoLooseFiles",
      "TestReviewProgramsRefuse",
      "TestReviewProgramsSelfTest",
      "TestTypeOfStringLiteralMutant",
      "TestUncheckableCastAdmission",
      "TestUnknownNarrowingMutants",
      "TestViewFieldInheritedStaticReadiness"
    ]
  }
]

CODE UNDER TEST: checked-view lowering in internal/lower, native/runtime emission in internal/native, and generated JavaScript checked-view dispatch in internal/javascript. No oracle, harness or test changes.
ORACLE: source on Node, full handwritten expected panic stderr and exit 70, builtin IR counterfactuals compared with Node, plus sanitized native execution and leak checks. MovedStage3Results compares all backend outputs to source Node. Handwritten panic strings are self oracles.

| Mutant | origin/main file:line | Change | Failed top-level rows |
|---|---|---|---|
| D1 | internal/javascript/view_unions_untagged.go:87 | drop fixed tuple array identity and length guard | TestCheckedViewV2TupleIdentityMutant |
| D2 | internal/javascript/view_unions_untagged.go:153 | flip callable producer certification bypass | TestCheckedViewV2CallableProducerMutant |
| D3b | internal/lower/view_objects.go:35 | flip structural cast receiver admission condition | TestCheckedViewObjectPrimitiveSource |
| D4 | internal/javascript/javascript.go:688 | change bitmask AND operator to OR | TestCheckedViewFlagDowncast, TestCheckedViewV2MovedStage3Results, TestNativeAgreesWithNode |
| D5 | internal/native/runtime/view_unions_untagged.c:141 | change open kind selector number storage option to boolean | TestCheckedViewUntaggedOwnClassData |

Coverage and semantic differences:
- Tuple: 0 exclusive Go source lines in coverage-diffs.json; Go coverage does not instrument generated JavaScript or runtime C branches.
- Producer: 4 exclusive Go source lines in coverage-diffs.json; Go coverage does not instrument generated JavaScript or runtime C branches.
- Moved: 1378 exclusive Go source lines in coverage-diffs.json; Go coverage does not instrument generated JavaScript or runtime C branches.
- Tuple identity: a plain object with fields 0 and 1, rather than a true tuple. D1 deletes the generated JavaScript tuple identity/length guard. The later narrowed-read guard still rejects it, but its stderr differs from the row's exact panic contract. This is a unique boundary-message defense, not proof that D1 admits a wrong tuple.
- Callable producer: the wrong closure accepts number while the union includes a literal-number or string producer. D2 bypasses certification and changes the checked panic to successful execution. Other scalar representation mismatches continue to be rejected.
- Moved stage-3: reduced TypeScript compiler functions exercise untagged structural downcasts, recursive predicate reads, truthy loops and bitmask guards. D3b targets structural admission; D4 targets bitwise AND; D5 removes the numeric open-kind fast path. Their observed failure sets are above.

Brief issues, costs, and limits:
- The audit's untrue verdicts concern witnesses under a globally disabled disagreement harness. That edit also disables their normal expected-panic assertions and Node counterfactual comparisons. This session proves production sensitivity without weakening that harness. These results do not redo a weakened-check witness proof.
- The whole-package baseline cooked after 90.106 seconds, with no ordinary assertion failures. Matrix selections are in matrix.json. All six newly added rows were included. Uniqueness outside the selections is unknown.
- The original corpus selector matched only the parent TestNativeAgreesWithNode. The corrected selector explicitly includes internal/oracle/testdata and its fixture children. Empty parent-only runs were excluded from the useful proof; both commands and their zero-child observations remain recorded.
- A first unconditional early-return candidate failed go vet for unreachable code and was rejected. It is not a counted attempt or valid replay mutant; D3b flips the existing receiver admission condition instead.
- Executed fixture children and their skips are recorded in matrix.json and the raw logs. Existing pending review cases and the views-v3 array subcases skipped; matrix logs record their exact names and reasons. No skipped row is claimed to pass.
- The requested 15 GB free space threshold cannot be met on the 8.8 GB /tmp filesystem. /workspace is a separate filesystem, so authorized /tmp cleanup cannot free workspace blocks. Earlier named scratch was removed; builds had ample available space and showed no disk-full failures.
- MovedStage3Results' name promises execution of five migrated fixture results, which it checks against Node across three backend modes with leak assertions. No name-versus-assertion gap was found. TupleIdentityMutant and CallableProducerMutant check both original panic and changed IR counterfactual behavior; their main assertions match their names.
- The independent observation initially used CommonJS on emitted ESM and failed before running the program. That rejected invocation is saved separately. Corrected observations use the unchanged oracle/node.mjs loader and show D1 exit 70 with narrowed-member stderr and D2 exit 0 with true stdout.
- There were no twin or cost rows among these three. No tests were deleted, rewritten, or weakened.
- Five valid production mutants are saved as standalone diffs. Go candidates passed go vet; the runtime C mutant compiled through the native test builds. Own per-mutant build caches were used. Full repository gates and other package suites were not run.

Timing: setup skipped because tools were warm; nproc 5. npm-ci.log captures installation. matrix.json records each command wall time and native rebuild-inclusive time. Clean package own-binary wall was 90.106 seconds. Native rebuilding and execution are not separated by the harness. Total recorded matrix/coverage wall seconds: 657.464. Four isolated profile times appear in the corresponding coverage logs.

The complete NativeAgreesWithNode clean family also cooked at 90 seconds (91.73 command wall seconds including compilation), with no ordinary assertion failures. The already completed reached-fixture matrix is the bound used for uniqueness.
Raw logs and coverage profiles are losslessly compressed as .log.gz and .cover.gz. Decompress a profile before go tool cover. The full list of completed top-level passes for each unique mutant is in rows.json; the exact fixture selections and subtest passes are in matrix.json.
