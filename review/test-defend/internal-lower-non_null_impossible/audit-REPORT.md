Audited f91994f019703ba25d2918cf529c0e0b0c05d93c; all fourteen requested functions exist, grouped into twelve rows.
Clean whole-package baseline: 44.571 binary seconds; restored selected-row control passes.
Seven sacred, four subsumed, one untrue; all verdicts use package-wide observed runs.
Fifteen menu mutants, one supplemental mutant and four separate entry probes; two survivors.
Production files restored; evidence only, for central replay.

```json
[
  {
    "test": "TestAdamicNullishAssertionsAreRefused",
    "package": "internal/lower",
    "file": "internal/lower/non_null_impossible_test.go:12",
    "seconds": 1.245,
    "oracle": "Handwritten Refused class, exact non-null What and Fix for nullish .a assertions.",
    "oracle_kind": "self",
    "kills": [
      "M01"
    ],
    "unique_kills": [],
    "last_proven_fail": "M01 non_null_impossible_test.go:53: want exactly-nullish Refused, got <nil>",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestNonNullAssertionOnPresentTypeIsErased"
    ],
    "mutants_in_matrix": 16,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 0.07,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "OPTIONAL_WIDENING_CONFIG=/workspace/adamic/review/test-audit/internal-lower-non_null_impossible/census-project/tsconfig.json OPTIONAL_WIDENING_OUTPUT=/tmp/u040-census.json ADAMIC_MUTANT=M01 ADAMIC_BUILD_CACHE_DIR=/tmp/u040/cache/M01 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; non_null_impossible_test.go:53: want exactly-nullish Refused, got <nil>",
    "subsumption_mutants": 1,
    "eligible_mutants_in_matrix": 15
  },
  {
    "test": "TestNonNullAssertionLowersToNullishPanic",
    "package": "internal/lower",
    "file": "internal/lower/non_null_test.go:14",
    "seconds": 0.074,
    "oracle": "Handwritten IR Coalesce/Panic presence, Number/MaybeNumber types and panic-message prefix/suffix.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M02",
      "M03"
    ],
    "unique_kills": [
      "M03"
    ],
    "last_proven_fail": "M03 non_null_test.go:35: unexpected panic text \"/tmp/adamic-gate/TestNonNullAssertionLowersToNullishPanic2853542428/001/main.ts:2:16: map.get('a')! is null or undefined\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 16,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "OPTIONAL_WIDENING_CONFIG=/workspace/adamic/review/test-audit/internal-lower-non_null_impossible/census-project/tsconfig.json OPTIONAL_WIDENING_OUTPUT=/tmp/u040-census.json ADAMIC_MUTANT=M03 ADAMIC_BUILD_CACHE_DIR=/tmp/u040/cache/M03 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; non_null_test.go:35: unexpected panic text \"/tmp/adamic-gate/TestNonNullAssertionLowersToNullishPanic2853542428/001/main.ts:2:16: map.get('a')! is null or undefined\"",
    "eligible_mutants_in_matrix": 15
  },
  {
    "test": "TestNonNullAssertionOnPresentTypeIsErased",
    "package": "internal/lower",
    "file": "internal/lower/non_null_test.go:39",
    "seconds": 0.07,
    "oracle": "Handwritten NumberConstant type and zero panic strings; does not check the numeric value. M04 changes 0 to 1 and passes.",
    "oracle_kind": "self",
    "kills": [
      "M01"
    ],
    "unique_kills": [],
    "last_proven_fail": "M01 non_null_test.go:43: /tmp/adamic-gate/TestNonNullAssertionOnPresentTypeIsErased1913261029/001/main.ts:1:15: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestNonNullAssertionLowersToNullishPanic"
    ],
    "mutants_in_matrix": 16,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 0.074,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "OPTIONAL_WIDENING_CONFIG=/workspace/adamic/review/test-audit/internal-lower-non_null_impossible/census-project/tsconfig.json OPTIONAL_WIDENING_OUTPUT=/tmp/u040-census.json ADAMIC_MUTANT=M01 ADAMIC_BUILD_CACHE_DIR=/tmp/u040/cache/M01 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; non_null_test.go:43: /tmp/adamic-gate/TestNonNullAssertionOnPresentTypeIsErased1913261029/001/main.ts:1:15: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case",
    "subsumption_mutants": 1,
    "eligible_mutants_in_matrix": 15
  },
  {
    "test": "TestOptionalIndexingMapShapeRefused",
    "package": "internal/lower",
    "file": "internal/lower/optional_indexing_boundary_test.go:9",
    "seconds": 0.068,
    "oracle": "Handwritten NotYet class and exact structural-Map diagnostic.",
    "oracle_kind": "self",
    "kills": [
      "M06"
    ],
    "unique_kills": [
      "M06"
    ],
    "last_proven_fail": "M06 optional_indexing_boundary_test.go:18: got <nil>, want structural Map receiver boundary",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 16,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "OPTIONAL_WIDENING_CONFIG=/workspace/adamic/review/test-audit/internal-lower-non_null_impossible/census-project/tsconfig.json OPTIONAL_WIDENING_OUTPUT=/tmp/u040-census.json ADAMIC_MUTANT=M06 ADAMIC_BUILD_CACHE_DIR=/tmp/u040/cache/M06 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; optional_indexing_boundary_test.go:18: got <nil>, want structural Map receiver boundary",
    "eligible_mutants_in_matrix": 15
  },
  {
    "test": "TestOptionalIndexingKeepsIndexSignaturesRefused",
    "package": "internal/lower",
    "file": "internal/lower/optional_indexing_test.go:9",
    "seconds": 0.067,
    "oracle": "Handwritten Refused class, exact index-signature What and Fix.",
    "oracle_kind": "self",
    "kills": [
      "M05"
    ],
    "unique_kills": [
      "M05"
    ],
    "last_proven_fail": "M05 optional_indexing_test.go:14: got /tmp/adamic-gate/TestOptionalIndexingKeepsIndexSignaturesRefused3460425717/001/main.a:1:21: Adamic 0.1 refuses an index signature; use a Map, want the existing index-signature refusal and fix",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 16,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "OPTIONAL_WIDENING_CONFIG=/workspace/adamic/review/test-audit/internal-lower-non_null_impossible/census-project/tsconfig.json OPTIONAL_WIDENING_OUTPUT=/tmp/u040-census.json ADAMIC_MUTANT=M05 ADAMIC_BUILD_CACHE_DIR=/tmp/u040/cache/M05 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; optional_indexing_test.go:14: got /tmp/adamic-gate/TestOptionalIndexingKeepsIndexSignaturesRefused3460425717/001/main.a:1:21: Adamic 0.1 refuses an index signature; use a Map, want the existing index-signature refusal and fix",
    "eligible_mutants_in_matrix": 15
  },
  {
    "test": "TestOptionalIndexingKeepsUnsupportedStorageNotYet",
    "package": "internal/lower",
    "file": "internal/lower/optional_indexing_test.go:18",
    "seconds": 0.23,
    "oracle": "Handwritten diagnostic classes/text for four negative cases; last continuation case checks only err == nil.",
    "oracle_kind": "self",
    "kills": [
      "M07"
    ],
    "unique_kills": [
      "M07"
    ],
    "last_proven_fail": "M07 optional_indexing_test.go:37: got <nil>, want NotYet containing \"an optional index that isn't a number\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 16,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "OPTIONAL_WIDENING_CONFIG=/workspace/adamic/review/test-audit/internal-lower-non_null_impossible/census-project/tsconfig.json OPTIONAL_WIDENING_OUTPUT=/tmp/u040-census.json ADAMIC_MUTANT=M07 ADAMIC_BUILD_CACHE_DIR=/tmp/u040/cache/M07 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; optional_indexing_test.go:37: got <nil>, want NotYet containing \"an optional index that isn't a number\"",
    "vacuous_subcases": [
      "index_after_ordinary_continuation"
    ],
    "eligible_mutants_in_matrix": 15
  },
  {
    "test": "TestOptionalWideningCensus",
    "package": "internal/lower",
    "file": "internal/lower/optional_widening_census_test.go:26",
    "seconds": 0.088,
    "oracle": "Only successful census construction/encoding; no expected site count or contents. Fixture-derived project: baseline one site; empty-answer probe zero sites still passes.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 16,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=P03 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run ^TestOptionalWideningCensus$; PASS with summary files=1 sites=0 (baseline sites=1)",
    "eligible_mutants_in_matrix": 15
  },
  {
    "test": "TestOptionalWideningRefused",
    "package": "internal/lower",
    "file": "internal/lower/optional_widening_test.go:14",
    "seconds": 0.639,
    "oracle": "Handwritten Refused diagnostics and own .refused snapshots; incompatible subcase checks a preparation-time TypeScript checker error.",
    "oracle_kind": "self",
    "kills": [
      "M08",
      "M11"
    ],
    "unique_kills": [
      "M11"
    ],
    "last_proven_fail": "M11 optional_widening_test.go:45: want optional-property refusal, got /workspace/adamic/internal/lower/testdata/optional_widening/argument.a:4:6: Adamic 0.1 refuses optional property y in { x: number; y?: number; } absent from structural source { x: number; }, which can hide fields; declare y on the source type, or build a fresh object with known fields (adamic/no-optional-view)",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 16,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "OPTIONAL_WIDENING_CONFIG=/workspace/adamic/review/test-audit/internal-lower-non_null_impossible/census-project/tsconfig.json OPTIONAL_WIDENING_OUTPUT=/tmp/u040-census.json ADAMIC_MUTANT=M11 ADAMIC_BUILD_CACHE_DIR=/tmp/u040/cache/M11 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; optional_widening_test.go:45: want optional-property refusal, got /workspace/adamic/internal/lower/testdata/optional_widening/argument.a:4:6: Adamic 0.1 refuses optional property y in { x: number; y?: number; } absent from structural source { x: number; }, which can hide fields; declare y on the source type, or build a fresh object with known fields (adamic/no-optional-view)",
    "preparation_only_subcases": [
      "incompatible"
    ],
    "eligible_mutants_in_matrix": 15
  },
  {
    "test": "TestOptionalWideningAllowed",
    "package": "internal/lower",
    "file": "internal/lower/optional_widening_test.go:71",
    "seconds": 0.133,
    "oracle": "Only err == nil for class/fresh/declared cases; no IR values or execution assertion.",
    "oracle_kind": "self",
    "kills": [
      "M08"
    ],
    "unique_kills": [],
    "last_proven_fail": "M08 optional_widening_test.go:81: /tmp/adamic-gate/TestOptionalWideningAllowedclass1934236940/001/main.a:3:42: Adamic 0.1 refuses optional property y in { x: number; y?: number; } absent from structural source Point, which can hide fields; declare y on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestOptionalWideningRefused"
    ],
    "mutants_in_matrix": 16,
    "probe_kills": [],
    "subsumer_seconds": 0.639,
    "vacuous": true,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "OPTIONAL_WIDENING_CONFIG=/workspace/adamic/review/test-audit/internal-lower-non_null_impossible/census-project/tsconfig.json OPTIONAL_WIDENING_OUTPUT=/tmp/u040-census.json ADAMIC_MUTANT=M08 ADAMIC_BUILD_CACHE_DIR=/tmp/u040/cache/M08 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; optional_widening_test.go:81: /tmp/adamic-gate/TestOptionalWideningAllowedclass1934236940/001/main.a:3:42: Adamic 0.1 refuses optional property y in { x: number; y?: number; } absent from structural source Point, which can hide fields; declare y on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "subsumption_mutants": 1,
    "eligible_mutants_in_matrix": 15
  },
  {
    "test": "TestOptionalWideningSpreadOverwrite",
    "package": "internal/lower",
    "file": "internal/lower/optional_widening_test.go:89",
    "seconds": 0.062,
    "oracle": "Only refuse returns nil for overwritten spread; no positive refusal case.",
    "oracle_kind": "self",
    "kills": [
      "M09"
    ],
    "unique_kills": [
      "M09"
    ],
    "last_proven_fail": "M09 optional_widening_test.go:105: /tmp/adamic-gate/TestOptionalWideningSpreadOverwrite1970651270/001/main.a:1:133: Adamic 0.1 refuses optional property y in { x: number; y?: number; } absent from structural source { x: number; }, which can hide fields; declare y on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 16,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "OPTIONAL_WIDENING_CONFIG=/workspace/adamic/review/test-audit/internal-lower-non_null_impossible/census-project/tsconfig.json OPTIONAL_WIDENING_OUTPUT=/tmp/u040-census.json ADAMIC_MUTANT=M09 ADAMIC_BUILD_CACHE_DIR=/tmp/u040/cache/M09 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; optional_widening_test.go:105: /tmp/adamic-gate/TestOptionalWideningSpreadOverwrite1970651270/001/main.a:1:133: Adamic 0.1 refuses optional property y in { x: number; y?: number; } absent from structural source { x: number; }, which can hide fields; declare y on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "eligible_mutants_in_matrix": 15
  },
  {
    "test": "TestOverloadInferenceWitnesses",
    "package": "internal/lower",
    "file": "internal/lower/overload_inference_test.go:12",
    "seconds": 0.063,
    "oracle": "Handwritten inferred type-map entries/absence for four binder cases; ordinary production test despite Witnesses name.",
    "oracle_kind": "self",
    "kills": [
      "M12",
      "M13",
      "M14"
    ],
    "unique_kills": [
      "M12",
      "M13"
    ],
    "last_proven_fail": "M14 overload_inference_test.go:59: lost the rigid binder behind undefined",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 16,
    "probe_kills": [
      "P04"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "OPTIONAL_WIDENING_CONFIG=/workspace/adamic/review/test-audit/internal-lower-non_null_impossible/census-project/tsconfig.json OPTIONAL_WIDENING_OUTPUT=/tmp/u040-census.json ADAMIC_MUTANT=M14 ADAMIC_BUILD_CACHE_DIR=/tmp/u040/cache/M14 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; overload_inference_test.go:59: lost the rigid binder behind undefined",
    "eligible_mutants_in_matrix": 15
  },
  {
    "test": "TestOverrideNativeSignature family",
    "package": "internal/lower",
    "file": "internal/lower/override_representation_test.go:11",
    "seconds": 0.117,
    "oracle": "Handwritten NotYet class and parameter/fix substrings, one checker over three fixtures.",
    "oracle_kind": "self",
    "kills": [
      "M15"
    ],
    "unique_kills": [],
    "last_proven_fail": "M15 override_representation_test.go:23: override_optional.a: want NotYet naming parameter \"factor\" and its repair, got /tmp/adamic-gate/TestOverrideOptionalRefusesNativeSignature646077512/001/main.a:6:5: stage 0 can't lower an override with a different native representation for parameter \"value\"; keep the base method's parameter form yet",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestInheritanceNativeSignatureLimits"
    ],
    "mutants_in_matrix": 16,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 0.307,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "OPTIONAL_WIDENING_CONFIG=/workspace/adamic/review/test-audit/internal-lower-non_null_impossible/census-project/tsconfig.json OPTIONAL_WIDENING_OUTPUT=/tmp/u040-census.json ADAMIC_MUTANT=M15 ADAMIC_BUILD_CACHE_DIR=/tmp/u040/cache/M15 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; override_representation_test.go:23: override_optional.a: want NotYet naming parameter \"factor\" and its repair, got /tmp/adamic-gate/TestOverrideOptionalRefusesNativeSignature646077512/001/main.a:6:5: stage 0 can't lower an override with a different native representation for parameter \"value\"; keep the base method's parameter form yet",
    "members": [
      "TestAbfe962OverrideDefaultAddedRefusesNativeSignature",
      "TestOverrideOptionalNumberRefusesNativeSignature",
      "TestOverrideOptionalRefusesNativeSignature"
    ],
    "subsumption_mutants": 1,
    "eligible_mutants_in_matrix": 15
  }
]
```

| ID | Origin file:line | Change | Failed rows |
|---|---|---|---|
| M01 | internal/lower/non_null.go:106 | ".ts" -> ".a" | TestAdamicNullishAssertionsAreRefused, TestNonNullAssertionLowersToNullishPanic, TestNonNullAssertionOnPresentTypeIsErased, TestPredicateOverloadRuntime, TestReadinessElisionRequiresDominatingAssignment, TestWhatZeroOneRefusesIsRefusedWithAFix |
| M02 | internal/lower/non_null.go:81 | value.Type() == ir.Number \|\| value.Type() == ir.Boolean -> value.Type() == ir.MaybeNumber \|\| value.Type() == ir.Boolean | TestNonNullAssertionLowersToNullishPanic, TestReadinessElisionRequiresDominatingAssignment |
| M03 | internal/lower/non_null.go:96 | "non-null assertion failed at " -> "" | TestNonNullAssertionLowersToNullishPanic |
| M04 supplemental | internal/lower/non_null.go:83 | return value, nil -> return ir.NumberConstant{Value: 1}, nil |  |
| M05 | internal/lower/refusals.go:28 | "use a Map, which keeps keys in the order they were added" -> "use a Map" | TestOptionalIndexingKeepsIndexSignaturesRefused |
| M06 | internal/lower/optional_indexing_map.go:99 | return hazard -> return false | TestOptionalIndexingMapShapeRefused |
| M07 | internal/lower/optional_indexing.go:35 | index.Type() != ir.Number -> index.Type() != ir.String | TestOptionalIndexingKeepsUnsupportedStorageNotYet |
| M08 | internal/lower/optional_widening.go:86 | !isClassInstance(source) -> isClassInstance(source) | TestClassWrongOutputIteratorReceiver, TestIteratorMapperIndexHasNumberRepresentation, TestOptionalWideningAllowed, TestOptionalWideningRefused |
| M09 | internal/lower/optional_widening.go:81 | skip[property.Name] -> false | TestOptionalWideningSpreadOverwrite |
| M10 | internal/lower/optional_widening.go:179 | target = l.impliedTarget(node) -> target = nil |  |
| M11 | internal/lower/optional_widening.go:196 | " on the source type, or build a fresh object with known fields (adamic/no-optional-widening)" -> " on the source type, or build a fresh object with known fields (adamic/no-optional-view)" | TestOptionalWideningRefused |
| M12 | internal/lower/generic.go:259 | len(present) == 1 && given.Flags()&checker.TypeFlagsUndefined == 0 -> len(present) == 0 && given.Flags()&checker.TypeFlagsUndefined == 0 | TestOverloadInferenceWitnesses |
| M13 | internal/lower/generic.go:269 | missing != nil -> missing == nil | TestOverloadInferenceWitnesses |
| M14 | internal/lower/generic.go:288 | if _, isSet := into[declared]; !isSet { -> if _, isSet := into[declared]; isSet { | TestAMutableLocationSeenWiderIsRefused, TestCensusOverloadBinderGuards, TestDefaultTaggedInterfaceAdmission, TestEmptyLiteralGenericReturnUsesSamePath, TestGenericFunctionPolymorphicRecursionIsRefused, TestGenericJSONUnionArrayIsNotYet, TestGenericUnionFixtureHasSeparateInstances, TestOverloadInferenceWitnesses, TestPredicateOverloadRuntime |
| M15 | internal/lower/class_inheritance.go:140 | checkABI && (!knownA \|\| !knownB \|\| a != b) -> checkABI && (!knownA \|\| !knownB \|\| a == b) | TestEnumSlotViews, TestInheritanceAllowsSoundOverrides, TestInheritanceNativeSignatureLimits, TestInheritanceNativeSignatureNeighbors, TestInheritanceRefusesFalseNominalViews, TestInheritanceRefusesUnsoundOverrides, TestOverrideNativeSignature family |
| M16 | internal/lower/class_inheritance.go:98 | !l.classAssignable(accepts, override) -> !l.classAssignable(override, accepts) | TestClassFeaturesAccessorRefusals |

Survivors:

M04 supplemental: generated assignment changes from 0 to 1; all package tests pass. See M04.witness.diff and witness-commands.json. No verdict rests on this supplemental mutant.

M10: equivalent candidate. Four independent inputs emit identical output before and after; see M10.equivalent-candidate.json. No unguarded behavior claim.

Brief feedback, unit u040

1. The historical file reference is 8de93800f4, but the required fetch selected f91994f019703ba25d2918cf529c0e0b0c05d93c. All fourteen names still exist in their listed files. The actual fetched commit is recorded so replay is unambiguous.
2. Fourteen named functions become twelve rows. The three override wrappers use one checker with different fixtures even though the first has a different name prefix. Shared checker semantics, rather than the prefix alone, settled the grouping.
3. TestOverloadInferenceWitnesses is not a witness under this brief: it tests production inference directly. The name is misleading for classification.
4. The opt-in census requires a project and output path, neither supplied in the brief. This run enabled it using a strict, one-file TypeScript project copied from the repository's initializer fixture. Its results describe that project only, not a larger upstream inventory.
5. Lower is not the only production entry. SpreadOverwrite calls refuse, Census calls optionalAtSite, and InferenceWitnesses calls inferTypes. They require separate probes. Preparation Load is not a probe target for lowering.
6. The census tests successful enumeration and encoding but has no site-count or content oracle. M08 changes its report from one site to zero without failing it. The standalone empty-answer probe independently tests this omission.
7. The present-type assertion checks NumberConstant and absence of panic strings but not its value. M04 changes generated assignment 0 to 1 with a green whole package. This is a proven survivor, not an equivalent candidate.
8. About three mutants per twelve rows would exceed the twenty-mutant limit. Sixteen were fixed before outcomes because the clean whole package took 44.571 binary seconds and each additional whole-package run consumed roughly another minute of the twenty-minute budget.
9. A compiler build and a test binary have different clocks. This report separates test-binary medians from command wall time. Go compilation, linking and native products inside tests cannot be divided precisely without extra instrumentation; no pure build time is inferred from subtraction.
10. Independent M10 witness attempts covered conditional, coalescing, inferred array and inferred arrow expressions. None changed output. It is an equivalent candidate rather than an unguarded-behavior finding.
11. Standalone empty-answer diffs replace the entry body and remove newly unused imports. The switched scratch source instead conditionally returns at entry. Both forms are preserved, and every standalone diff passes origin apply checks and Go vet through a source overlay.
12. The first independent M04 source used console.log(number), which this project's library rejects because it expects a string. The corrected source formats the value as a string; both failed attempts and successful commands remain in the record.

13. M04 replaces a variable expression in an existing early return with a constant. The menu does not explicitly authorize changing return expressions. Conservatively marked supplemental, although it inserts no statement. No sacred or subsumed verdict rests on it. There are fifteen menu mutants and one supplemental mutant.

Timing:

```json
{
  "setup_seconds": 0,
  "nproc": 5,
  "npm_reported_seconds": 0.844,
  "baseline_binary_seconds": 44.571,
  "isolated_timing_and_coverage_command_seconds": 113.91132990799815,
  "coverage_command_seconds": 8.320914790998359,
  "matrix_command_seconds": 834.4859189309973,
  "matrix_binary_seconds": 733.764,
  "standalone_vet_seconds": 29.515807594012585,
  "cli_build_seconds": 32.42051046399865,
  "subsumer_timing_command_seconds": 13.325322484997741,
  "restored_control_seconds": 4.879109001998586,
  "restored_control_exit": 0
}
```

Limits: no other packages tested, no repository-wide uniqueness claim, no native output execution for selected rows (their own oracles are self assertions). Census covers a fixture-derived single-file project. The full baseline skips TestOriginalCycleLedger and TestMixedUnionContractGraph/interface_Node_{readonly_ready:boolean}_type_Target=Node|readonly_Node[]; neither belongs to this unit. All requested rows ran. Production runs have no panics/timeouts; P01 nil-IR probes panic in two individually isolated rows, recorded as failures. Subsumption rests on one eligible caught mutant per subsumed row.
