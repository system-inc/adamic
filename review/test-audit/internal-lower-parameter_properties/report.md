Starting commit: 6da361da470ada7ae90c42e6d947f3154c509e8b; all four requested rows remain in the stated files.
Verdicts: three sacred, one subsumed, measured against 12 production mutants.
All four catch their own entry's empty-answer probe; checker-contract positive Lower subcase passes P1.
Two whole-package survivors change behavior: missing parameter stores, disabled unchecked-index checking.
Evidence is on test-audit/internal-lower-parameter_properties under review/test-audit/internal-lower-parameter_properties/; production restored.

CODE UNDER TEST, named before mutations: Adamic's parameter-property recognition, layout/store/default/reset handling, inherited override checking, named-function receiver handling, predicate-summary proof, and loader configuration/diagnostic collection. The upstream tsgo checker was not mutated. The complete pre-mutation reached-function list is reached-functions.txt, 354 functions across 85 files, measured with coverage.out. functions-coverage.txt also includes zero-hit functions. Each chosen mutant function was covered before mutation.

ORACLE, named before mutations: self-written Refused/NotYet/CheckError type checks, diagnostic fragments, and fixture locations. No selected row compares generated behavior against an independent external execution. CheckerContracts invokes tsgo in-process through Adamic Load, but asserts only the self-written CheckError expectation, not a specific tsc diagnostic. No external-authority value was supplied or checked. Node executions below are audit witnesses, not existing test oracles.

The four rows have different assertions and are not a family. No subprocess helper, setup-check, or witness row was identified. All four names appear in the actual go test -list output, and none moved from the stated files or vanished. The start commit is newer than the brief's 8de93800f4 reference.

Because the clean package binary fit in 90 seconds, every production mutant was run against the whole lower package. All 12 production runs completed without a package panic or timeout. unique_kills are from completed full-package runs with exactly one failed top-level row. P1/P2 panic-aborted the package and were rerun for all four selected rows alone; outside rows for those probes stay unknown. The rows are marked bounded for this slice and its probe reruns; matrix.json lists full production observation rows and exact failed rows, including outside-unit catches. Default skipped rows are unknown, and repo-wide uniqueness is left to central replay.

```json
[
  {
    "test": "TestParameterPropertiesSoundness",
    "package": "internal/lower",
    "file": "internal/lower/parameter_properties_test.go:13",
    "seconds": 0.19,
    "oracle": "Self-written Refused type and reason fragments across six unsafe parameter-property sources. Does not execute accepted programs or inspect initialized field values.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M02",
      "M04",
      "M05",
      "M06"
    ],
    "unique_kills": [
      "M02",
      "M04",
      "M05"
    ],
    "last_proven_fail": "M06: parameter_properties_test.go:27: got <nil>, want refusal invariant-mutable",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": true,
    "matrix_rows": [
      "TestParameterPropertiesSoundness",
      "TestParameterPropertyCheckerContracts",
      "TestParameterPropertyCallbackReceiver",
      "TestPredicateSummaryParameterIndex"
    ],
    "evidence": "ADAMIC_MUTANT=M06 ADAMIC_BUILD_CACHE_DIR=/tmp/u041/cache/M06 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > review/test-audit/internal-lower-parameter_properties/M06.log 2>&1; parameter_properties_test.go:27: got <nil>, want refusal invariant-mutable",
    "subsumption_mutant_count": null
  },
  {
    "test": "TestParameterPropertyCheckerContracts",
    "package": "internal/lower",
    "file": "internal/lower/parameter_properties_test.go:33",
    "seconds": 0.122,
    "oracle": "Self-written load.CheckError type for readonly/private/protected writes and nil error for positive lowering. Negative cases accept any CheckError, not an exact diagnostic. The positive Lower subcase passes P1.",
    "oracle_kind": "self",
    "kills": [
      "M11"
    ],
    "unique_kills": [],
    "last_proven_fail": "M11: parameter_properties_test.go:47: got <nil>, want checker rejection",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestClassFeaturesPrivateChecker"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P2"
    ],
    "subsumer_seconds": 0.039,
    "vacuous": false,
    "vacuous_subcases": [
      "positive mutable parameter-property Lower subcase passes P1"
    ],
    "bounded": true,
    "matrix_rows": [
      "TestParameterPropertiesSoundness",
      "TestParameterPropertyCheckerContracts",
      "TestParameterPropertyCallbackReceiver",
      "TestPredicateSummaryParameterIndex"
    ],
    "evidence": "ADAMIC_MUTANT=M11 ADAMIC_BUILD_CACHE_DIR=/tmp/u041/cache/M11 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > review/test-audit/internal-lower-parameter_properties/M11.log 2>&1; parameter_properties_test.go:47: got <nil>, want checker rejection",
    "subsumption_mutant_count": 1
  },
  {
    "test": "TestParameterPropertyCallbackReceiver",
    "package": "internal/lower",
    "file": "internal/lower/parameter_properties_test.go:55",
    "seconds": 0.034,
    "oracle": "Self-written NotYet type and this parameter used as a value fragment; no exact location or runtime output.",
    "oracle_kind": "self",
    "kills": [
      "M07"
    ],
    "unique_kills": [
      "M07"
    ],
    "last_proven_fail": "M07: parameter_properties_test.go:60: got <nil>, want NotYet dynamic receiver",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": true,
    "matrix_rows": [
      "TestParameterPropertiesSoundness",
      "TestParameterPropertyCheckerContracts",
      "TestParameterPropertyCallbackReceiver",
      "TestPredicateSummaryParameterIndex"
    ],
    "evidence": "ADAMIC_MUTANT=M07 ADAMIC_BUILD_CACHE_DIR=/tmp/u041/cache/M07 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > review/test-audit/internal-lower-parameter_properties/M07.log 2>&1; parameter_properties_test.go:60: got <nil>, want NotYet dynamic receiver",
    "subsumption_mutant_count": null
  },
  {
    "test": "TestPredicateSummaryParameterIndex",
    "package": "internal/lower",
    "file": "internal/lower/predicate_refusals_test.go:13",
    "seconds": 0.093,
    "oracle": "Self-written helper argument/index-fix fragments plus fixture basename:5:10 location. No external expected-answer authority.",
    "oracle_kind": "self",
    "kills": [
      "M08",
      "M09",
      "M10"
    ],
    "unique_kills": [
      "M09",
      "M10"
    ],
    "last_proven_fail": "M10: predicate_refusals_test.go:44: want helper call path and parameter-index fix, got : Adamic 0.1 refuses an unproven type predicate; helper argument 0 does not occupy its predicate parameter; pass the tested value at the helper's predicate parameter index; return a boolean and narrow at the caller (adamic/no-type-predicate)",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P3"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": true,
    "matrix_rows": [
      "TestParameterPropertiesSoundness",
      "TestParameterPropertyCheckerContracts",
      "TestParameterPropertyCallbackReceiver",
      "TestPredicateSummaryParameterIndex"
    ],
    "evidence": "ADAMIC_MUTANT=M10 ADAMIC_BUILD_CACHE_DIR=/tmp/u041/cache/M10 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > review/test-audit/internal-lower-parameter_properties/M10.log 2>&1; predicate_refusals_test.go:44: want helper call path and parameter-index fix, got : Adamic 0.1 refuses an unproven type predicate; helper argument 0 does not occupy its predicate parameter; pass the tested value at the helper's predicate parameter index; return a boolean and narrow at the caller (adamic/no-type-predicate)",
    "subsumption_mutant_count": null
  }
]
```

| ID | origin/main file:line | Change | Every observed failed row |
| --- | --- | --- | --- |
| M01 | internal/lower/parameter_properties.go:9 | flip parameter-kind condition | TestAbfe962OverrideDefaultAddedRefusesNativeSignature, TestClassFeaturesPrivateStorage, TestClassFeaturesReadonlyChecker, TestConstructorCacheOneArgumentRemainsNotYet, TestInheritanceAllowsSoundOverrides, TestInheritanceConditionalThisRules, TestInheritanceGenericFactoryLayouts, TestInheritanceGenericMonomorphizations, TestInheritanceGenericSoundness, TestInheritanceNativeSignatureLimits, TestInheritanceNativeSignatureNeighbors, TestInheritanceRefusesFalseNominalViews, TestInheritanceRefusesGrowingGenericClasses, TestInheritanceRefusesUnsoundOverrides, TestOverrideOptionalNumberRefusesNativeSignature, TestOverrideOptionalRefusesNativeSignature, TestParameterPropertiesSoundness, TestReadonlyFieldsAreJudgedByTheirConstructorsWrites, TestWhatZeroOneRefusesIsRefusedWithAFix |
| M02 | internal/lower/parameter_properties.go:20 | drop parameter-member append | TestParameterPropertiesSoundness |
| M03 | internal/lower/parameter_properties.go:28 | return early: nil, nil |  |
| M04 | internal/lower/parameter_properties.go:50 | return early: nil | TestParameterPropertiesSoundness |
| M05 | internal/lower/parameter_properties.go:76 | drop inherited availability reset | TestParameterPropertiesSoundness |
| M06 | internal/lower/class_inheritance.go:48 | return early: nil | TestAbfe962OverrideDefaultAddedRefusesNativeSignature, TestClassFeaturesAccessorRefusals, TestClassFeaturesStaticSoundness, TestInheritanceGenericSoundness, TestInheritanceNativeSignatureLimits, TestInheritanceRefusesFalseNominalViews, TestInheritanceRefusesUnsoundOverrides, TestOverrideOptionalNumberRefusesNativeSignature, TestOverrideOptionalRefusesNativeSignature, TestParameterPropertiesSoundness |
| M07 | internal/lower/expression.go:1238 | drop named-function this-parameter guard loop | TestParameterPropertyCallbackReceiver |
| M08 | internal/lower/predicates_proof.go:404 | off-by-one helper predicate parameter index | TestPredicateBodyProof, TestPredicateSummaryParameterIndex |
| M09 | internal/lower/predicates_proof.go:403 | drop claim declaration and parameter-index guard block | TestPredicateSummaryParameterIndex |
| M10 | internal/lower/predicates_proof.go:39 | drop predicate diagnostic location assignment | TestPredicateSummaryParameterIndex |
| M11 | internal/load/load.go:233 | drop semantic diagnostic collection | TestClassFeaturesPrivateChecker, TestClassFeaturesReadonlyChecker, TestInheritanceKeepsCheckerConstructorRules, TestInheritedLibraryReadsNeverLoadOwnFields, TestInputSpreadCoverage, TestModuleNamespaceLimitsStayLoud, TestNullishPrototypeReadsAreRejectedByChecker, TestOptionalWideningRefused, TestParameterPropertyCheckerContracts, TestTsgoHonorsNoCheckInAdamicFiles |
| M12 | internal/load/load.go:60 | disable unchecked-index option |  |
| P1 | internal/lower/lower.go:20 | return early: nil, nil | TestParameterPropertiesSoundness, TestParameterPropertyCallbackReceiver |
| P2 | internal/load/load.go:80 | return early: nil, nil | TestParameterPropertiesSoundness, TestParameterPropertyCheckerContracts, TestParameterPropertyCallbackReceiver, TestPredicateSummaryParameterIndex |
| P3 | internal/lower/predicates_proof.go:47 | return early: predicateProof{}, nil | TestArrayPredicateCannotInventAnElementContract, TestCensusPredicateInteractionBoundary, TestCensusPredicateInteractionEscapes, TestCensusPredicateMarkerKeepsProofBoundaries, TestEveryNeedsCallbackEffects, TestPredicateBodyProof, TestPredicateCallbackContracts, TestPredicateOverloadCallback, TestPredicateSummaryParameterIndex, TestUnprovenPredicateReturnsAreRefused |

Commands and validation:
The matrix command is ADAMIC_MUTANT=<id> ADAMIC_BUILD_CACHE_DIR=/tmp/u041/cache/<id> timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > <id>.log 2>&1. Each panic-isolated rerun uses -run '^TestName$' and the same mutant cache. All output is preserved in raw logs; rows.json records exact commands and failing lines. One scratch switch was compiled and its no-mutant selected baseline passed. switch.diff records the switched code. Production was then restored exactly and the whole package passed again.

Every standalone diff applies to the starting Git index and passed go vet ./<mutated package>/ using a Go overlay: the single candidate diff is overlaid onto origin/main, and all other scratch-mutated files are overlaid with their original contents. standalone-validation.json records all 15 apply/vet exits and wall times. No test or oracle edit was used. No supplemental insertion mutant was used. M09 drops the claim declaration with its guard so no unused local remains. The compiled native tests receive separate ADAMIC_BUILD_CACHE_DIR values; selected rows only inspect lowering/checker results and do not build native products. Individual native rebuild times in the whole package were not separated from binary run time.

Survivors, both non-equivalent:
- M03: ADAMIC_MUTANT= go run ./review/test-audit/internal-lower-parameter_properties/witness store prints 7. ADAMIC_MUTANT=M03 with the same command prints 0. The witness lowers a constructor parameter property and runs its generated JavaScript through unchanged oracle/node.mjs. Whole-package tests all survived removal of parameterPropertyStores. This is unguarded behavior within the observed package matrix.
- M12: ADAMIC_MUTANT= go run ./review/test-audit/internal-lower-parameter_properties/witness index prints TS2322 rejecting number | undefined assigned to number. ADAMIC_MUTANT=M12 prints load error: <nil>. Whole-package tests all survived disabling NoUncheckedIndexedAccess. This is unguarded loader-option behavior within the observed package matrix; central replay may find other-package catches.

Subsumption:
CheckerContracts catches only M11. TestClassFeaturesPrivateChecker catches M11 too, recorded in the same whole-package log: class_features_test.go:46: want tsc private scope diagnostic, got <nil>. Its body was read in this session, and three clean isolated runs after the matrix/restored package finished measured 0.041, 0.038, 0.039 s, median 0.039 s. Subsumption rests on one mutant and is a retention hint, not a deletion recommendation. Earlier measurements taken concurrently with the matrix are preserved separately and are not used for the reported median.

Brief friction, ambiguity, and limits:
- Fresh origin/main is 6da361da, while the supplied file reference was 8de93800f4. Names and source positions were verified against fresh main, not assumed from that reference. None moved or vanished.
- The brief calls this a slice but asks for whole-package runs when the baseline fits. Whole production runs were feasible here and were used; full-package unique kills are distinguished from the bounded probe reruns and still-skipped rows.
- CheckerContracts has two code-under-test entries. It passes P1 as a whole because negative Load checks still reject and its positive Lower call only checks nil error. It fails P2. Aggregate vacuous=false, with the positive Lower subcase explicitly listed as vacuous. Lower-only rows are never judged by Load preparation failures; the direct predicate row is judged only by P3.
- M03 returns an empty statement list from parameterPropertyStores, a production helper rather than the row's Lower entry. It is an allowed return-early mutant, independently shown to change output 7 to 0. P1, not M03, is the Lower empty-answer probe. No verdict or unique claim rests on this survivor.
- M09 needed to drop the helper claim declaration with its guard to keep the standalone diff compileable. Early-return standalone diffs use an if true block so the unchanged following code remains compile-checkable without unreachable-code vet errors.
- The checker oracle accepts any load.CheckError for each invalid source. A different diagnostic of that type can satisfy its negative case. Its positive lowering case cannot distinguish an empty result from a working program, demonstrated by P1. The other rows verify named refusal types and reasons, but none executes parameter-property success behavior.
- Default baseline skips outside the four rows: TestOriginalCycleLedger (a particular pristine TypeScript Git corpus with generated diagnostics), TestOptionalWideningCensus (a specified project and output destination), and the array-contract subcase of TestMixedUnionContractGraph (awaits compiler/views-v3). Those separate corpus/project workflows were not provisioned for this unit; their catches remain unknown. All four selected rows ran with no skips.
- The first extra subsumer timings overlapped a running matrix. They were preserved and replaced by three runs after all package work finished to avoid CPU-contention distortion. Selected-row medians were measured before matrix execution.
- The menu contains 12 production mutants, approximately three per row. It is a finite experiment and not exhaustive proof. The subsumption verdict rests on one observed kill.
- Individual native build times and exact isolated Go compile time were not instrumented. Whole command wall times and test-binary elapsed times are saved; no precise native-build-time claim is made. Compiler-cache separation was set for every mutant run.
- No other package tests were run. Upstream checker code, the test harness, and Node runner were not changed. No production fix, new gate test, PR, or main push was made. Repo-wide uniqueness and policy for skipped external corpus checks remain for central replay.

Timing:
Warm setup skipped; env.sh worked with Go 1.27.1 and nproc 5. npm ci command took 0.310 s. Fetch/branch/toolchain confirmation command took 2.380 s. Clean whole-package binary 22.144 s; restored whole-package binary 21.345 s. Coverage command took 4.976 s. Twelve isolated selected-row test binaries total 1.301 s, with individual runs and medians in timings.json. Standalone validations total 6.965 s. Switched clean-check command took 7.796 s including its build/run. Matrix plus switched clean check and panic reruns total 321.610 command wall seconds. No test binary cooked. Per-mutant command walls appear below; raw JSON logs include package binary times. No toolchain setup or stage1 rebuild was needed.

```json
{
  "clean-switch": {
    "whole_command_wall_seconds": 7.796,
    "isolated_wall_seconds": 0
  },
  "M01": {
    "whole_command_wall_seconds": 23.597,
    "isolated_wall_seconds": 0
  },
  "M02": {
    "whole_command_wall_seconds": 23.528,
    "isolated_wall_seconds": 0
  },
  "M03": {
    "whole_command_wall_seconds": 23.431,
    "isolated_wall_seconds": 0
  },
  "M04": {
    "whole_command_wall_seconds": 23.883,
    "isolated_wall_seconds": 0
  },
  "M05": {
    "whole_command_wall_seconds": 26.613,
    "isolated_wall_seconds": 0
  },
  "M06": {
    "whole_command_wall_seconds": 23.71,
    "isolated_wall_seconds": 0
  },
  "M07": {
    "whole_command_wall_seconds": 23.158,
    "isolated_wall_seconds": 0
  },
  "M08": {
    "whole_command_wall_seconds": 23.544,
    "isolated_wall_seconds": 0
  },
  "M09": {
    "whole_command_wall_seconds": 24.657,
    "isolated_wall_seconds": 0
  },
  "M10": {
    "whole_command_wall_seconds": 23.318,
    "isolated_wall_seconds": 0
  },
  "M11": {
    "whole_command_wall_seconds": 11.347,
    "isolated_wall_seconds": 0
  },
  "M12": {
    "whole_command_wall_seconds": 23.547,
    "isolated_wall_seconds": 0
  },
  "P1": {
    "whole_command_wall_seconds": 1.695,
    "isolated_wall_seconds": 6.788
  },
  "P2": {
    "whole_command_wall_seconds": 1.615,
    "isolated_wall_seconds": 6.5409999999999995
  },
  "P3": {
    "whole_command_wall_seconds": 22.842,
    "isolated_wall_seconds": 0
  }
}
```
