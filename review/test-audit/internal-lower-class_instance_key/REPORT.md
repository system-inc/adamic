u030: all 14 assigned names exist in their original files; none moved or vanished.
Base: origin/main 8171b3173bdbfce1f7982d3c4f731279307ece37.
Clean enabled baseline: PASS, 32.623 test-binary seconds; nproc=5.
Verdicts: 7 sacred, 5 subsumed, 1 overlapping, 1 cannot-judge; five vacuous rows.
Evidence: test-audit/internal-lower-class_instance_key, review/test-audit/internal-lower-class_instance_key/.

```json
[
  {
    "test": "TestClassArgumentsUseCheckerIdentity",
    "package": "internal/lower",
    "file": "internal/lower/class_instance_key_test.go:5",
    "seconds": 0.22,
    "oracle": "Self-written class-instantiation counts for checker-equivalent and distinct arguments. Counts do not verify class contents or native behavior.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M02",
      "M03",
      "M12"
    ],
    "unique_kills": [],
    "last_proven_fail": "M12 class_instance_key_test.go:33: got 0 class instantiations, want 1",
    "verdict": "overlapping",
    "subsumed_by": [
      "TestInheritanceGenericMonomorphizations",
      "TestGenericFunctionPolymorphicRecursionIsRefused"
    ],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "evidence": "ADAMIC_MUTANT=M12 ADAMIC_BUILD_CACHE_DIR=/tmp/u030/cache/M12 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run ^TestClassArgumentsUseCheckerIdentity$; class_instance_key_test.go:33: got 0 class instantiations, want 1",
    "matrix_rows": [
      "TestClassArgumentsUseCheckerIdentity",
      "TestClassWrongOutputIteratorReceiver",
      "TestClassWrongOutputKeysRepair",
      "TestClassWrongOutputPrivateRepair",
      "TestClockGenericReturnsT01Shapes",
      "TestClockGenericReturnsT01DeclinesBrandedPrimitives",
      "TestClockGenericReturnsT01RejectsNullBeforeBody",
      "TestClockGenericReturnsT01RejectsIndexBeforeBody",
      "TestDefiniteAssignmentUsesReadiness",
      "TestDefiniteAssignmentSoundNeighbors",
      "TestNestedEmptyArrayElementKinds",
      "TestEmptyLiteralGenericReturnUsesSamePath",
      "TestEntriesAllocationProof",
      "TestEntriesRecordBoundaries"
    ],
    "subsumption_mutants": null
  },
  {
    "test": "TestClassWrongOutputIteratorReceiver",
    "package": "internal/lower",
    "file": "internal/lower/class_wrong_output_test.go:9",
    "seconds": 0.169,
    "oracle": "Self-written iterator-receiver refusal label plus accepted unchanged-protocol controls. Does not execute iterator output.",
    "oracle_kind": "self",
    "kills": [
      "M04"
    ],
    "unique_kills": [
      "M04"
    ],
    "last_proven_fail": "M04 class_wrong_output_test.go:24: unsafe protocol accepted: for (const value of new ScaledIterator()) { console.log(`${value}`); break; }: <nil>",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M04 ADAMIC_BUILD_CACHE_DIR=/tmp/u030/cache/M04 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; class_wrong_output_test.go:24: unsafe protocol accepted: for (const value of new ScaledIterator()) { console.log(`${value}`); break; }: <nil>",
    "subsumption_mutants": null
  },
  {
    "test": "TestClassWrongOutputKeysRepair",
    "package": "internal/lower",
    "file": "internal/lower/class_wrong_output_test.go:37",
    "seconds": 0.037,
    "oracle": "Self-written acceptance expectation for an explicit string-key copy. Checks only absence of an error.",
    "oracle_kind": "self",
    "kills": [
      "M05",
      "M19"
    ],
    "unique_kills": [],
    "last_proven_fail": "M19 class_wrong_output_test.go:45: explicit string-key copy refused: /tmp/adamic-gate/TestClassWrongOutputKeysRepair4132599371/001/main.a:1:16: stage 0 can't lower record allocation outside scalar storage yet",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestClassFeaturesPrivateStorage"
    ],
    "mutants_in_matrix": 19,
    "probe_kills": [],
    "subsumer_seconds": 0.053,
    "vacuous": true,
    "bounded": true,
    "evidence": "ADAMIC_MUTANT=M19 ADAMIC_BUILD_CACHE_DIR=/tmp/u030/cache/M19 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; class_wrong_output_test.go:45: explicit string-key copy refused: /tmp/adamic-gate/TestClassWrongOutputKeysRepair4132599371/001/main.a:1:16: stage 0 can't lower record allocation outside scalar storage yet",
    "matrix_rows": [
      "TestClassArgumentsUseCheckerIdentity",
      "TestClassWrongOutputIteratorReceiver",
      "TestClassWrongOutputKeysRepair",
      "TestClassWrongOutputPrivateRepair",
      "TestClockGenericReturnsT01Shapes",
      "TestClockGenericReturnsT01DeclinesBrandedPrimitives",
      "TestClockGenericReturnsT01RejectsNullBeforeBody",
      "TestClockGenericReturnsT01RejectsIndexBeforeBody",
      "TestDefiniteAssignmentUsesReadiness",
      "TestDefiniteAssignmentSoundNeighbors",
      "TestNestedEmptyArrayElementKinds",
      "TestEmptyLiteralGenericReturnUsesSamePath",
      "TestEntriesAllocationProof",
      "TestEntriesRecordBoundaries"
    ],
    "subsumption_mutants": 2
  },
  {
    "test": "TestClassWrongOutputPrivateRepair",
    "package": "internal/lower",
    "file": "internal/lower/class_wrong_output_test.go:49",
    "seconds": 0.036,
    "oracle": "Self-written acceptance expectation for instance delegation of private storage. Checks only absence of an error.",
    "oracle_kind": "self",
    "kills": [
      "M14"
    ],
    "unique_kills": [],
    "last_proven_fail": "M14 class_wrong_output_test.go:53: instance delegation refused: /tmp/adamic-gate/TestClassWrongOutputPrivateRepair2181356568/001/main.a:1:60: Adamic 0.1 refuses this escaping a constructor before every field is set (stored, passed, or a method called on it, which could read a field that holds undefined while its type says otherwise); assign every field first, then use this",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestDefiniteAssignmentSoundNeighbors"
    ],
    "mutants_in_matrix": 19,
    "probe_kills": [],
    "subsumer_seconds": 0.111,
    "vacuous": true,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M14 ADAMIC_BUILD_CACHE_DIR=/tmp/u030/cache/M14 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; class_wrong_output_test.go:53: instance delegation refused: /tmp/adamic-gate/TestClassWrongOutputPrivateRepair2181356568/001/main.a:1:60: Adamic 0.1 refuses this escaping a constructor before every field is set (stored, passed, or a method called on it, which could read a field that holds undefined while its type says otherwise); assign every field first, then use this",
    "subsumption_mutants": 1
  },
  {
    "test": "TestClockGenericReturnsT01Shapes",
    "package": "internal/lower",
    "file": "internal/lower/clock_generic_returns_t_01_test.go:14",
    "seconds": 0.121,
    "oracle": "Self-written shape acceptance and NotYet/Refused stop expectations. Unsupported subcases accept either stop type and can pass on a different later refusal.",
    "oracle_kind": "self",
    "kills": [
      "M09"
    ],
    "unique_kills": [
      "M09"
    ],
    "last_proven_fail": "M09 clock_generic_returns_t_01_test.go:40: want unsupported shape stop, got <nil>",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M09 ADAMIC_BUILD_CACHE_DIR=/tmp/u030/cache/M09 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; clock_generic_returns_t_01_test.go:40: want unsupported shape stop, got <nil>",
    "vacuous_subcases": [
      "nested_readonly_fields"
    ],
    "subsumption_mutants": null
  },
  {
    "test": "TestClockGenericReturnsT01DeclinesBrandedPrimitives",
    "package": "internal/lower",
    "file": "internal/lower/clock_generic_returns_t_01_test.go:60",
    "seconds": 0.062,
    "oracle": "Self-written NotYet expectation for branded primitive results. Does not check the reason or representation and can pass on a different later NotYet.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "M10 (supplemental): panic: runtime error: invalid memory address or nil pointer dereference; stack clock_generic_returns_t_01_test.go:66",
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "evidence": "ADAMIC_MUTANT=M10 ADAMIC_BUILD_CACHE_DIR=/tmp/u030/cache/M10 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run ^TestClockGenericReturnsT01DeclinesBrandedPrimitives$; panic: runtime error: invalid memory address or nil pointer dereference; stack clock_generic_returns_t_01_test.go:66",
    "supplemental_kills": [
      "M10"
    ],
    "cannot_judge_reason": "The only meaningful challenge to the branded type-argument guard removed condition terms, outside the fixed menu. It is supplemental. No admissible mutation meaningfully challenged that guard; no worthy verdict rests on M10.",
    "matrix_rows": [
      "TestClassArgumentsUseCheckerIdentity",
      "TestClassWrongOutputIteratorReceiver",
      "TestClassWrongOutputKeysRepair",
      "TestClassWrongOutputPrivateRepair",
      "TestClockGenericReturnsT01Shapes",
      "TestClockGenericReturnsT01DeclinesBrandedPrimitives",
      "TestClockGenericReturnsT01RejectsNullBeforeBody",
      "TestClockGenericReturnsT01RejectsIndexBeforeBody",
      "TestDefiniteAssignmentUsesReadiness",
      "TestDefiniteAssignmentSoundNeighbors",
      "TestNestedEmptyArrayElementKinds",
      "TestEmptyLiteralGenericReturnUsesSamePath",
      "TestEntriesAllocationProof",
      "TestEntriesRecordBoundaries"
    ],
    "subsumption_mutants": null
  },
  {
    "test": "TestClockGenericReturnsT01RejectsNullBeforeBody",
    "package": "internal/lower",
    "file": "internal/lower/clock_generic_returns_t_01_test.go:76",
    "seconds": 0.035,
    "oracle": "Self-written direct signature-proof expectation known=false for distinct null and undefined. Does not require a positive control.",
    "oracle_kind": "self",
    "kills": [
      "M07"
    ],
    "unique_kills": [
      "M07"
    ],
    "last_proven_fail": "M07 clock_generic_returns_t_01_test.go:100: null and undefined admitted with one representation: 4",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 19,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M07 ADAMIC_BUILD_CACHE_DIR=/tmp/u030/cache/M07 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; clock_generic_returns_t_01_test.go:100: null and undefined admitted with one representation: 4",
    "subsumption_mutants": null
  },
  {
    "test": "TestClockGenericReturnsT01RejectsIndexBeforeBody",
    "package": "internal/lower",
    "file": "internal/lower/clock_generic_returns_t_01_test.go:114",
    "seconds": 0.036,
    "oracle": "Self-written direct signature-proof expectation known=false for an index signature. Does not require a positive control.",
    "oracle_kind": "self",
    "kills": [
      "M08"
    ],
    "unique_kills": [
      "M08"
    ],
    "last_proven_fail": "M08 clock_generic_returns_t_01_test.go:138: indexed object admitted by the finite-shape proof: 4",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 19,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M08 ADAMIC_BUILD_CACHE_DIR=/tmp/u030/cache/M08 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; clock_generic_returns_t_01_test.go:138: indexed object admitted by the finite-shape proof: 4",
    "subsumption_mutants": null
  },
  {
    "test": "TestDefiniteAssignmentUsesReadiness",
    "package": "internal/lower",
    "file": "internal/lower/definite_assignment_test.go:8",
    "seconds": 0.109,
    "oracle": "Self-written counts of nonempty readiness tags on IR reads/properties. Does not verify tag text or runtime readiness behavior.",
    "oracle_kind": "self",
    "kills": [
      "M11",
      "M12"
    ],
    "unique_kills": [],
    "last_proven_fail": "M12 definite_assignment_test.go:42: 0 checked reads, want 1",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestReadinessElisionRequiresDominatingAssignment"
    ],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 0.382,
    "vacuous": false,
    "bounded": true,
    "evidence": "ADAMIC_MUTANT=M12 ADAMIC_BUILD_CACHE_DIR=/tmp/u030/cache/M12 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run ^TestDefiniteAssignmentUsesReadiness$; definite_assignment_test.go:42: 0 checked reads, want 1",
    "matrix_rows": [
      "TestClassArgumentsUseCheckerIdentity",
      "TestClassWrongOutputIteratorReceiver",
      "TestClassWrongOutputKeysRepair",
      "TestClassWrongOutputPrivateRepair",
      "TestClockGenericReturnsT01Shapes",
      "TestClockGenericReturnsT01DeclinesBrandedPrimitives",
      "TestClockGenericReturnsT01RejectsNullBeforeBody",
      "TestClockGenericReturnsT01RejectsIndexBeforeBody",
      "TestDefiniteAssignmentUsesReadiness",
      "TestDefiniteAssignmentSoundNeighbors",
      "TestNestedEmptyArrayElementKinds",
      "TestEmptyLiteralGenericReturnUsesSamePath",
      "TestEntriesAllocationProof",
      "TestEntriesRecordBoundaries"
    ],
    "subsumption_mutants": 2
  },
  {
    "test": "TestDefiniteAssignmentSoundNeighbors",
    "package": "internal/lower",
    "file": "internal/lower/definite_assignment_test.go:47",
    "seconds": 0.111,
    "oracle": "Self-written acceptance expectation for sound initialized neighbors and prose mentioning ts-ignore. Checks only absence of an error.",
    "oracle_kind": "self",
    "kills": [
      "M14"
    ],
    "unique_kills": [],
    "last_proven_fail": "M14 definite_assignment_test.go:57: sound neighbor: /tmp/adamic-gate/TestDefiniteAssignmentSoundNeighbors1620298715/003/main.a:1:40: Adamic 0.1 refuses this escaping a constructor before every field is set (stored, passed, or a method called on it, which could read a field that holds undefined while its type says otherwise); assign every field first, then use this",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestClassWrongOutputPrivateRepair"
    ],
    "mutants_in_matrix": 19,
    "probe_kills": [],
    "subsumer_seconds": 0.036,
    "vacuous": true,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M14 ADAMIC_BUILD_CACHE_DIR=/tmp/u030/cache/M14 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; definite_assignment_test.go:57: sound neighbor: /tmp/adamic-gate/TestDefiniteAssignmentSoundNeighbors1620298715/003/main.a:1:40: Adamic 0.1 refuses this escaping a constructor before every field is set (stored, passed, or a method called on it, which could read a field that holds undefined while its type says otherwise); assign every field first, then use this",
    "subsumption_mutants": 1
  },
  {
    "test": "TestNestedEmptyArrayElementKinds",
    "package": "internal/lower",
    "file": "internal/lower/empty_literal_test.go:9",
    "seconds": 0.112,
    "oracle": "Self-written IR element kinds and presence of an empty literal. Union case permits Number or String, rather than one exact layout.",
    "oracle_kind": "self",
    "kills": [
      "M12",
      "M15",
      "M16"
    ],
    "unique_kills": [
      "M16"
    ],
    "last_proven_fail": "M16 empty_literal_test.go:26: /tmp/adamic-gate/TestNestedEmptyArrayElementKindsstrings888584698/001/main.a:1:16: stage 0 can't lower an array of never yet",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "evidence": "ADAMIC_MUTANT=M16 ADAMIC_BUILD_CACHE_DIR=/tmp/u030/cache/M16 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; empty_literal_test.go:26: /tmp/adamic-gate/TestNestedEmptyArrayElementKindsstrings888584698/001/main.a:1:16: stage 0 can't lower an array of never yet",
    "matrix_rows": [
      "TestClassArgumentsUseCheckerIdentity",
      "TestClassWrongOutputIteratorReceiver",
      "TestClassWrongOutputKeysRepair",
      "TestClassWrongOutputPrivateRepair",
      "TestClockGenericReturnsT01Shapes",
      "TestClockGenericReturnsT01DeclinesBrandedPrimitives",
      "TestClockGenericReturnsT01RejectsNullBeforeBody",
      "TestClockGenericReturnsT01RejectsIndexBeforeBody",
      "TestDefiniteAssignmentUsesReadiness",
      "TestDefiniteAssignmentSoundNeighbors",
      "TestNestedEmptyArrayElementKinds",
      "TestEmptyLiteralGenericReturnUsesSamePath",
      "TestEntriesAllocationProof",
      "TestEntriesRecordBoundaries"
    ],
    "subsumption_mutants": null
  },
  {
    "test": "TestEmptyLiteralGenericReturnUsesSamePath",
    "package": "internal/lower",
    "file": "internal/lower/empty_literal_test.go:46",
    "seconds": 0.038,
    "oracle": "Self-written exact multiset of Number and Object generic empty-return IR element kinds.",
    "oracle_kind": "self",
    "kills": [
      "M15"
    ],
    "unique_kills": [],
    "last_proven_fail": "M15 empty_literal_test.go:52: /tmp/adamic-gate/TestEmptyLiteralGenericReturnUsesSamePath789301906/001/main.a:1:35: stage 0 can't lower an array of never yet",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestNestedEmptyArrayElementKinds"
    ],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 0.112,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M15 ADAMIC_BUILD_CACHE_DIR=/tmp/u030/cache/M15 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; empty_literal_test.go:52: /tmp/adamic-gate/TestEmptyLiteralGenericReturnUsesSamePath789301906/001/main.a:1:35: stage 0 can't lower an array of never yet",
    "subsumption_mutants": 1
  },
  {
    "test": "TestEntriesAllocationProof",
    "package": "internal/lower",
    "file": "internal/lower/entries_provenance_test.go:11",
    "seconds": 0.13,
    "oracle": "Self-written allocation-proof/Checked bits and exactly one entries call in IR. Does not execute enumeration output.",
    "oracle_kind": "self",
    "kills": [
      "M17",
      "M18",
      "M19"
    ],
    "unique_kills": [
      "M17",
      "M18"
    ],
    "last_proven_fail": "M19 entries_provenance_test.go:35: proven=true, checked=true",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M19 ADAMIC_BUILD_CACHE_DIR=/tmp/u030/cache/M19 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; entries_provenance_test.go:35: proven=true, checked=true",
    "subsumption_mutants": null
  },
  {
    "test": "TestEntriesRecordBoundaries",
    "package": "internal/lower",
    "file": "internal/lower/entries_provenance_test.go:52",
    "seconds": 0.067,
    "oracle": "Self-written NotYet type plus specific reason substrings for three record boundaries.",
    "oracle_kind": "self",
    "kills": [
      "M12",
      "M19",
      "M20"
    ],
    "unique_kills": [
      "M20"
    ],
    "last_proven_fail": "M20 entries_provenance_test.go:64: want explicit NotYet \"closed const literal origin\", got /tmp/adamic-gate/TestEntriesRecordBoundariesclosed_const_literal_origin3602173433/001/main.a:1:51: stage 0 can't lower an indexed mutation without a proven allocation origin yet",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "evidence": "ADAMIC_MUTANT=M20 ADAMIC_BUILD_CACHE_DIR=/tmp/u030/cache/M20 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; entries_provenance_test.go:64: want explicit NotYet \"closed const literal origin\", got /tmp/adamic-gate/TestEntriesRecordBoundariesclosed_const_literal_origin3602173433/001/main.a:1:51: stage 0 can't lower an indexed mutation without a proven allocation origin yet",
    "matrix_rows": [
      "TestClassArgumentsUseCheckerIdentity",
      "TestClassWrongOutputIteratorReceiver",
      "TestClassWrongOutputKeysRepair",
      "TestClassWrongOutputPrivateRepair",
      "TestClockGenericReturnsT01Shapes",
      "TestClockGenericReturnsT01DeclinesBrandedPrimitives",
      "TestClockGenericReturnsT01RejectsNullBeforeBody",
      "TestClockGenericReturnsT01RejectsIndexBeforeBody",
      "TestDefiniteAssignmentUsesReadiness",
      "TestDefiniteAssignmentSoundNeighbors",
      "TestNestedEmptyArrayElementKinds",
      "TestEmptyLiteralGenericReturnUsesSamePath",
      "TestEntriesAllocationProof",
      "TestEntriesRecordBoundaries"
    ],
    "subsumption_mutants": null
  }
]
```

| ID | Origin/main file:line | Change | Observed failed rows | Unknown rows |
|---|---|---|---|---|
| [M01](M01.diff) | internal/lower/generic.go:167 | Negate checker identity comparison | TestClassArgumentsUseCheckerIdentity, TestGenericFunctionPolymorphicRecursionIsRefused, TestGenericJSONUnionArrayIsNotYet, TestInheritanceGenericMonomorphizations, TestInheritanceRefusesGrowingGenericClasses | 0 |
| [M02](M02.diff) | internal/lower/generic.go:172 | Drop type representative registration statement | TestClassArgumentsUseCheckerIdentity, TestGenericFunctionPolymorphicRecursionIsRefused, TestGenericJSONUnionArrayIsNotYet, TestGenericUnionFixtureHasSeparateInstances | 0 |
| [M03](M03.diff) | internal/lower/class.go:801 | Drop whole class-key argument loop | TestClassArgumentsUseCheckerIdentity, TestInheritanceGenericFactoryLayouts, TestInheritanceGenericMonomorphizations, TestInheritanceRefusesGrowingGenericClasses, TestWhatZeroOneRefusesIsRefusedWithAFix | 0 |
| [M04](M04.diff) | internal/lower/iteration.go:223 | Negate receiver-overrides condition | TestClassWrongOutputIteratorReceiver | 0 |
| [M05](M05.diff) | internal/lower/class_features.go:83 | Flip fresh literal condition | TestClassFeaturesPrivateStorage, TestClassWrongOutputKeysRepair | 87 |
| [M06](M06.diff) | internal/lower/class_static_private.go:9 | Flip private-identifier exclusion |  | 0 |
| [M07](M07.diff) | internal/lower/clock_generic_returns_t_01.go:16 | Union member bound !=2 to <2 | TestClockGenericReturnsT01RejectsNullBeforeBody | 0 |
| [M08](M08.diff) | internal/lower/clock_generic_returns_t_01.go:76 | Index-signature condition !=0 to <0 | TestClockGenericReturnsT01RejectsIndexBeforeBody | 0 |
| [M09](M09.diff) | internal/lower/clock_generic_returns_t_01.go:74 | Call-signature condition !=0 to <0 | TestClockGenericReturnsT01Shapes | 0 |
| [M10](M10.diff) | internal/lower/clock_generic_returns_t_01.go:42 | SUPPLEMENTAL: remove guard terms before GetTypeArguments | TestClockGenericReturnsT01DeclinesBrandedPrimitives | 153 |
| [M11](M11.diff) | internal/lower/lower.go:83 | Drop readiness pass call | TestDefiniteAssignmentUsesReadiness, TestPredicateOverloadRuntime, TestReadinessElisionRequiresDominatingAssignment, TestUndecidedCycleReadsUseReadyChecks | 0 |
| [M12](M12.diff) | internal/lower/readiness.go:45 | Flip variable exclamation-token condition | TestClassArgumentsUseCheckerIdentity, TestClassFeaturesAccessorCaptureCycle, TestClassFeaturesAccessorRefusals, TestDefaultTaggedInterfaceAdmission, TestDefiniteAssignmentUsesReadiness, TestEntriesRecordBoundaries, TestFunctionValueUnionViewsStayNotYet, TestInheritanceConditionalThisRules, TestInheritanceGenericFactoryLayouts, TestInheritanceGenericMonomorphizations, TestInheritanceGenericNominalConstraints, TestInheritanceHasClassIdentity, TestInheritanceRefusesGrowingGenericClasses, TestInheritanceRefusesThisBeforeSuperReturns, TestIsPrototypeOfReadsExplainThePrototypeRefusal, TestIteratorDescriptorReasons, TestIteratorGapsAreExplicit, TestIteratorMapperIndexHasNumberRepresentation, TestLibraryMapSetIteratorCopyTypesRefused, TestLibraryMethodValueBoundaries, TestLibraryMethodValueSafety, TestLiteralMethodCapturesCannotMakeCycles, TestNestedEmptyArrayElementKinds, TestNestedEnvironmentCycleIncludesDisjointSlots, TestNestedFunctionGapsAreLoud, TestNodeFSFileRefusesOptionEffects, TestNodeFSFileRefusesVoidValues, TestNonNullAssertionOnPresentTypeIsErased, TestNumericEnumNeverProof, TestOptionalIndexingKeepsUnsupportedStorageNotYet, TestOverloadedShorthandFunctionValueStaysNotYet, TestParameterPropertiesSoundness, TestParameterPropertyCallbackReceiver, TestPredicateOverloadRuntime, TestPrimitiveAdmittingSlotsUseBoxes, TestPrototypeHazardsBehindObjectViewsAreNotYet, TestPrototypeMethodsAreRefusedWithReasons, TestReadinessElisionRequiresDominatingAssignment, TestReadonlyFieldsAreJudgedByTheirConstructorsWrites, TestRegExpNativeRefusals, TestTypedArrayGaps, TestTypedArrayViewsCannotChangeRepresentation, TestUndecidedCycleReadsUseReadyChecks, TestUnrepresentedPrototypeCallsAreNotYet | 33 |
| [M13](M13.diff) | internal/lower/object.go:577 | Drop property readiness-origin assignment |  | 0 |
| [M14](M14.diff) | internal/lower/class.go:769 | Flip property-declaration permission in useOfThis | TestAMethodReadAsAValueIsRefused, TestAViewThatCantWriteIsNotRefused, TestClassFeaturesPrivateStorage, TestClassFeaturesReadonlyChecker, TestClassWrongOutputPrivateRepair, TestConstructorCacheOneArgumentRemainsNotYet, TestDefiniteAssignmentSoundNeighbors, TestInheritanceGenericMonomorphizations, TestInheritanceRefusesGrowingGenericClasses, TestReadonlyFieldsAreJudgedByTheirConstructorsWrites, TestWhatZeroOneRefusesIsRefusedWithAFix | 0 |
| [M15](M15.diff) | internal/lower/empty_literal.go:13 | Negate array-type condition | TestATupleSeenAsAnArrayIsNotYet, TestEmptyLiteralGenericReturnUsesSamePath, TestNestedEmptyArrayElementKinds, TestPredicateOverloadRuntime, TestReadonlyFieldsAreJudgedByTheirConstructorsWrites, TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat, TestWhatZeroOneRefusesIsRefusedWithAFix | 0 |
| [M16](M16.diff) | internal/lower/object.go:306 | Drop implied empty-literal target context append | TestNestedEmptyArrayElementKinds | 0 |
| [M17](M17.diff) | internal/lower/entries_provenance.go:58 | Negate enumeration binding-written condition | TestEntriesAllocationProof | 0 |
| [M18](M18.diff) | internal/lower/entries_provenance.go:89 | Written detection true to false | TestEntriesAllocationProof | 0 |
| [M19](M19.diff) | internal/lower/entries_records.go:117 | Return negated record-written flag | TestATupleSeenAsAnArrayIsNotYet, TestArgumentsLengthReadNeighbors, TestClassFeaturesAccessorRefusals, TestClassFeaturesPrivateStorage, TestClassWrongOutputKeysRepair, TestDefaultTaggedInterfaceAdmission, TestEntriesAllocationProof, TestEntriesRecordBoundaries, TestEnumInitializationReach, TestInheritanceGenericSoundness, TestIteratorDescriptorReasons, TestIteratorDestructuringDoesNotLieAboutExhaustion, TestIteratorSymbolKeysAreNotStringKeys, TestLibraryLanguageBoundaries, TestLibraryMapSetGapsStayRefused, TestLiteralMethodCapturesCannotMakeCycles, TestLiteralMethodViewsDoNotLoseThis, TestOptionalIndexingMapShapeRefused, TestOptionalWideningAllowed, TestParserFactoryBindingHoisting, TestPrototypeHazardsBehindObjectViewsAreNotYet, TestViewObjectContractsAreAvailableToEraser, TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat, TestWhatZeroOneRefusesIsRefusedWithAFix | 0 |
| [M20](M20.diff) | internal/lower/entries_records.go:213 | Change indexed-write diagnostic reason constant | TestEntriesRecordBoundaries | 0 |

P1 at lower.go:20 returns nil,nil at Lower entry. P2 at clock_generic_returns_t_01.go:14 returns 0,false at the direct signature-proof entry. Their standalone diffs use tautological pointer guards solely to avoid vet's unreachable-code rejection. Results are recorded per assigned entry, not across unrelated entries. Probe failures are excluded from production kills.

Survivors:
- M06: privateStaticMember(#secret) changes present=true to false; public field changes present=false to true. Actual function-output witness in survivor-original.log and survivor-M06.log. Unguarded member classification; downstream runtime consequences were not evaluated.
- M13: IR property n readiness changes from "new Box().n" to "n". Actual Lower-output witness in survivor-original.log and survivor-M13.log. Unguarded diagnostic-origin metadata change; checked-read counts remain unchanged.
Both witnesses used `timeout 120 go test -v -count=1 -timeout 90s ./internal/lower/ -run ^TestAuditU030SurvivorWitness$` before and after applying the standalone diff. No equivalent candidate among the observed survivors.

Brief interpretation, mistakes, and costs:
- The reference commit 8de93800f4 was older than fetched origin/main. All 14 names were checked against the actual starting test list and remained in their named files.
- env.sh worked, but listing the lower package required a cold build, approximately 47.18 seconds. No cloud/setup.sh was run.
- The default package baseline skips two unrelated inventories. Both were enabled by fetching pristine TypeScript v6.0.3, verifying the pin, and generating diagnostics with upstream's script. Source cloning took about 9.27 seconds. Tests load no extra Node dependency directory for this setup; the generator uses Node builtins.
- The enabled baseline still skips TestMixedUnionContractGraph/interface_Node_{readonly_ready:boolean}_type_Target=Node|readonly_Node[];, whose body explicitly awaits compiler/views-v3. No installable tool enables that implementation gap.
- There are 239 top-level test functions and 237 package rows after grouping the three assertOverrideParameterRefusal wrappers. Their bodies call one identical checker with different fixture inputs. Families and members are in families.json. The 14 assigned functions retain separate rows because they have distinct test bodies/assertions rather than input-only checker wrappers. Shared lowerSource prepares and invokes lowering; using that helper alone does not establish an input-only wrapper family.
- The unconditional entry return probes failed go vet as unreachable code. Tautological pointer guards preserve an unconditional empty answer while passing vet. The initial failure is saved in probe-first-vet.log. This added one failed validation and a resume.
- M10 removes condition terms, which does not literally match the supplied menu. It is supplemental and excluded from kills, unique kills, subsumption, and worthy verdicts. The branded row can fail on M10, but the admissible menu did not meaningfully attack its guard, so its verdict is cannot-judge. The matrix preserves the supplemental column and matrix-kinds.json labels it. There are 19 admissible production mutants, one supplemental mutation, and two probes, not 22 production mutants.
- Actual Go binary panics happened for M05, M10, and M12. All 14 assigned rows were rerun alone for each. Outside unfinished rows remain unknown: M05 87, M10 153, M12 33 grouped rows. A failed child establishes row failure even when a panic prevents the parent completion event.
- The original detector incorrectly treated quoted native stderr containing panic: as a Go test binary panic on M11. That run completed normally with every row resolved; 14 unnecessary reruns cost about 25 seconds. The saved runner's detector is corrected to match an actual panic output line. M11 is not an aborted matrix.
- bounded=true indicates an assigned row's results include an aborted package run or the supplemental aborted guard run. matrix_rows lists the isolated replay set of all 14 assigned rows. Completed outside-row failures from the original run are retained in the full matrix and disqualify uniqueness; unknowns are never treated as passes. The seven sacred rows each have an admissible unique kill from a completed whole-package run.
- Direct negative signature proofs correctly decline their inputs even with no implementation. Their own P2 entry probe passes, so both are vacuous under the prescribed definition. Acceptance-only Lower rows also pass P1. These are findings independent of their verdicts.
- Shapes admits its positive nested-readonly-fields subcase on P1 and fails its negative subcases. The iterator row fails its first negative case on P1 before its positive controls execute; no claim is made that those unexecuted controls passed.
- Whole-package tests include TestUndecidedCycleReadsUseReadyChecks, which emits C, builds sanitized native products, and runs Node. Every mutant received its own ADAMIC_BUILD_CACHE_DIR. Assigned rows themselves do not build native products. Individual native rebuild durations were not instrumented; package binary times bound their cost. The brief's four-mutant fallback for native rebuilds and twenty-mutant Go switch menu are ambiguous for these incidental whole-package builds: this run used twenty switched Go changes and reports that limit rather than claiming four native rebuilds.
- Reciprocal subsumption names the other assigned row for PrivateRepair and SoundNeighbors, each resting on M14 alone. KeysRepair rests on two admissible kills; UsesReadiness on two; GenericReturn on one. These are finite-matrix hints, not deletion recommendations. Outside subsumers were timed three times after restoration. No claim is made that the chosen subsumer is globally fastest among every candidate.
- The audit took approximately 34 minutes, beyond the requested approximate 20-minute target. All individual test binaries fit their 90-second budget; no cooked run required coverage narrowing. Whole-package repetitions and panic reruns account for most of the cost.

Setup/build/run costs:
Warm tool verification: under one second, Go 1.27.1, nproc=5. npm ci reported 326 ms. Initial list/build wall about 47.18 s. Default baseline 30.261 s and enabled baseline 32.623 s in the test binary; Node diagnostics generation under one second. Reach coverage command 8.746 s wall, listing 422 reached functions.
All 42 assigned isolated timing commands: 76.944 s wall. Median binary samples are in results.json and raw logs. Production/probe vet commands: 11.169 s wall, plus the initial rejected P1 validation (0.491 s). The switched Go binary built once in 7.573 s.
Twenty matrix commands (including supplemental M10): 634.767 s wall. Isolated mutation reruns: 101.141 s wall, including the unnecessary M11 reruns. Fourteen entry-probe commands: 26.767 s wall. Survivor witnesses and outside-subsumer timings: 37.576 s wall. Detailed timings preserve every command and exit.
No other Go packages were tested, no repo-wide uniqueness was claimed, no PR was opened, and main was not pushed. The whole-package native cycle tests ran, but their individual rebuild times were not separately measured. Supplemental observation tests are not mutation kills and were removed. All production sources and submodule state are restored. Every standalone diff applies to the starting base and passed go vet ./internal/lower/; selector scaffolding is saved separately as selector.txt.
