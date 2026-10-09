u060: all 15 assigned names exist; no families, moves or vanished functions.
Base: 859dee825a4ef89f5c4ecc0f00e6620ac75c4994.
Bounded verdicts: 2 sacred, 8 subsumed, 5 witness.
All ten production rows rejected their Lower probe; witness vacuity was not judged by production probes.
Evidence: review/test-audit/internal-oracle-enums_open/ on test-audit/internal-oracle-enums_open.

```json
[
  {
    "test": "TestNumericEnumNeverPathsPinned",
    "package": "internal/oracle",
    "file": "internal/oracle/enums_open_test.go:19",
    "seconds": 0.195,
    "oracle": "Self-pinned before/lookup stdout, numeric-enum panic stderr and exit 70; native and JavaScript must match the full triple.",
    "oracle_kind": "self",
    "kills": [
      "M1"
    ],
    "unique_kills": [],
    "last_proven_fail": "M1: enums_open_test.go:39: native: exit codes differ; {stdout:[] stderr:[] exitCode:0}",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestNumericEnumNeverPinned"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1",
      "P2",
      "P3"
    ],
    "subsumer_seconds": 0.088,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNumericEnumNeverPathsPinned",
      "TestNumericEnumNeverPinned",
      "TestStage3EnumBoundaries",
      "TestStage3EnumSparseArrayBoundary",
      "TestEnumNameEnumerationMutant",
      "TestEnumSemanticMutants",
      "TestEnumCleanupMutant",
      "TestFallthroughMutants",
      "TestFieldReadinessRepresentation",
      "TestFreshWriteProbesStayRefused",
      "TestFunctionValueBoundaryBoxing",
      "TestImportCycleRuntimeCalls",
      "TestImportCycleLoadTimeReads",
      "TestImportedNonliteralConstCaseIsNotYet",
      "TestInputAgreesWithNode"
    ],
    "evidence": "ADAMIC_MUTANT=M1 ADAMIC_BUILD_CACHE_DIR=/tmp/u060/cache/M1 ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestNumericEnumNeverPathsPinned|TestNumericEnumNeverPinned|TestStage3EnumBoundaries|TestStage3EnumSparseArrayBoundary|TestEnumNameEnumerationMutant|TestEnumSemanticMutants|TestEnumCleanupMutant|TestFallthroughMutants|TestFieldReadinessRepresentation|TestFreshWriteProbesStayRefused|TestFunctionValueBoundaryBoxing|TestImportCycleRuntimeCalls|TestImportCycleLoadTimeReads|TestImportedNonliteralConstCaseIsNotYet|TestInputAgreesWithNode)$' > review/test-audit/internal-oracle-enums_open/M1.log 2>&1; enums_open_test.go:39: native: exit codes differ; {stdout:[] stderr:[] exitCode:0}"
  },
  {
    "test": "TestNumericEnumNeverPinned",
    "package": "internal/oracle",
    "file": "internal/oracle/enums_open_test.go:46",
    "seconds": 0.088,
    "oracle": "Self-pinned Adamic panic triple; separately runs source on Node and checks its successful stdout.",
    "oracle_kind": [
      "self",
      "external-run"
    ],
    "kills": [
      "M1"
    ],
    "unique_kills": [],
    "last_proven_fail": "M1: enums_open_test.go:60: native: exit codes differ; exit 0, stdout \"\", stderr \"\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestStage3EnumSparseArrayBoundary"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1",
      "P2",
      "P3"
    ],
    "subsumer_seconds": 0.098,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNumericEnumNeverPathsPinned",
      "TestNumericEnumNeverPinned",
      "TestStage3EnumBoundaries",
      "TestStage3EnumSparseArrayBoundary",
      "TestEnumNameEnumerationMutant",
      "TestEnumSemanticMutants",
      "TestEnumCleanupMutant",
      "TestFallthroughMutants",
      "TestFieldReadinessRepresentation",
      "TestFreshWriteProbesStayRefused",
      "TestFunctionValueBoundaryBoxing",
      "TestImportCycleRuntimeCalls",
      "TestImportCycleLoadTimeReads",
      "TestImportedNonliteralConstCaseIsNotYet",
      "TestInputAgreesWithNode"
    ],
    "evidence": "ADAMIC_MUTANT=M1 ADAMIC_BUILD_CACHE_DIR=/tmp/u060/cache/M1 ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestNumericEnumNeverPathsPinned|TestNumericEnumNeverPinned|TestStage3EnumBoundaries|TestStage3EnumSparseArrayBoundary|TestEnumNameEnumerationMutant|TestEnumSemanticMutants|TestEnumCleanupMutant|TestFallthroughMutants|TestFieldReadinessRepresentation|TestFreshWriteProbesStayRefused|TestFunctionValueBoundaryBoxing|TestImportCycleRuntimeCalls|TestImportCycleLoadTimeReads|TestImportedNonliteralConstCaseIsNotYet|TestInputAgreesWithNode)$' > review/test-audit/internal-oracle-enums_open/M1.log 2>&1; enums_open_test.go:60: native: exit codes differ; exit 0, stdout \"\", stderr \"\""
  },
  {
    "test": "TestStage3EnumBoundaries",
    "package": "internal/oracle",
    "file": "internal/oracle/enums_stage3_test.go:38",
    "seconds": 0.087,
    "oracle": "TypeScript-Go checker errors must have TS2532; Adamic refusal must have invariant-mutable. Error class plus substring only, not full diagnostic location.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M3"
    ],
    "unique_kills": [],
    "last_proven_fail": "M3: enums_stage3_test.go:62: want boundary invariant-mutable, got /workspace/adamic/stage3/fixtures/enums/06_set_node_flags.a:79:10: Adamic 0.1 refuses a value of type T seen as Mutable<T>, which can write T[\"flags\"] where NodeFlags is read;",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestFreshWriteProbesStayRefused"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 0.678,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNumericEnumNeverPathsPinned",
      "TestNumericEnumNeverPinned",
      "TestStage3EnumBoundaries",
      "TestStage3EnumSparseArrayBoundary",
      "TestEnumNameEnumerationMutant",
      "TestEnumSemanticMutants",
      "TestEnumCleanupMutant",
      "TestFallthroughMutants",
      "TestFieldReadinessRepresentation",
      "TestFreshWriteProbesStayRefused",
      "TestFunctionValueBoundaryBoxing",
      "TestImportCycleRuntimeCalls",
      "TestImportCycleLoadTimeReads",
      "TestImportedNonliteralConstCaseIsNotYet",
      "TestInputAgreesWithNode"
    ],
    "evidence": "ADAMIC_MUTANT=M3 ADAMIC_BUILD_CACHE_DIR=/tmp/u060/cache/M3 ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestNumericEnumNeverPathsPinned|TestNumericEnumNeverPinned|TestStage3EnumBoundaries|TestStage3EnumSparseArrayBoundary|TestEnumNameEnumerationMutant|TestEnumSemanticMutants|TestEnumCleanupMutant|TestFallthroughMutants|TestFieldReadinessRepresentation|TestFreshWriteProbesStayRefused|TestFunctionValueBoundaryBoxing|TestImportCycleRuntimeCalls|TestImportCycleLoadTimeReads|TestImportedNonliteralConstCaseIsNotYet|TestInputAgreesWithNode)$' > review/test-audit/internal-oracle-enums_open/M3.log 2>&1; enums_stage3_test.go:62: want boundary invariant-mutable, got /workspace/adamic/stage3/fixtures/enums/06_set_node_flags.a:79:10: Adamic 0.1 refuses a value of type T seen as Mutable<T>, which can write T[\"flags\"] where NodeFlags is read;",
    "probe_scope": "Lower subcase only; two checker-error subcases did not reach Lower and were not entry-probed."
  },
  {
    "test": "TestStage3EnumSparseArrayBoundary",
    "package": "internal/oracle",
    "file": "internal/oracle/enums_stage3_test.go:68",
    "seconds": 0.098,
    "oracle": "Self-pinned native and JavaScript array-bounds panic triple. Node side checks only successful exit, not its stdout.",
    "oracle_kind": [
      "self",
      "external-run"
    ],
    "kills": [
      "M1"
    ],
    "unique_kills": [],
    "last_proven_fail": "M1: enums_stage3_test.go:82: native: exit codes differ; {stdout:[] stderr:[] exitCode:0}",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestNumericEnumNeverPinned"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1",
      "P2",
      "P3"
    ],
    "subsumer_seconds": 0.088,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNumericEnumNeverPathsPinned",
      "TestNumericEnumNeverPinned",
      "TestStage3EnumBoundaries",
      "TestStage3EnumSparseArrayBoundary",
      "TestEnumNameEnumerationMutant",
      "TestEnumSemanticMutants",
      "TestEnumCleanupMutant",
      "TestFallthroughMutants",
      "TestFieldReadinessRepresentation",
      "TestFreshWriteProbesStayRefused",
      "TestFunctionValueBoundaryBoxing",
      "TestImportCycleRuntimeCalls",
      "TestImportCycleLoadTimeReads",
      "TestImportedNonliteralConstCaseIsNotYet",
      "TestInputAgreesWithNode"
    ],
    "evidence": "ADAMIC_MUTANT=M1 ADAMIC_BUILD_CACHE_DIR=/tmp/u060/cache/M1 ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestNumericEnumNeverPathsPinned|TestNumericEnumNeverPinned|TestStage3EnumBoundaries|TestStage3EnumSparseArrayBoundary|TestEnumNameEnumerationMutant|TestEnumSemanticMutants|TestEnumCleanupMutant|TestFallthroughMutants|TestFieldReadinessRepresentation|TestFreshWriteProbesStayRefused|TestFunctionValueBoundaryBoxing|TestImportCycleRuntimeCalls|TestImportCycleLoadTimeReads|TestImportedNonliteralConstCaseIsNotYet|TestInputAgreesWithNode)$' > review/test-audit/internal-oracle-enums_open/M1.log 2>&1; enums_stage3_test.go:82: native: exit codes differ; {stdout:[] stderr:[] exitCode:0}"
  },
  {
    "test": "TestEnumNameEnumerationMutant",
    "package": "internal/oracle",
    "file": "internal/oracle/enums_stage3_test.go:92",
    "seconds": 0.089,
    "oracle": "Actual Node source stdout comparison must detect built-in enum-field order swap.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: enums_stage3_test.go:134: want Node stdout catch, got \"\"",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestEnumNameEnumerationMutant",
      "TestEnumSemanticMutants",
      "TestFallthroughMutants",
      "TestFieldReadinessRepresentation"
    ],
    "evidence": "ADAMIC_MUTANT=W1 ADAMIC_BUILD_CACHE_DIR=/tmp/u060/cache/W1 ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestEnumNameEnumerationMutant|TestEnumSemanticMutants|TestFallthroughMutants|TestFieldReadinessRepresentation)$' > review/test-audit/internal-oracle-enums_open/W1.log 2>&1; enums_stage3_test.go:134: want Node stdout catch, got \"\"",
    "witness_kills": [
      "W1"
    ]
  },
  {
    "test": "TestEnumSemanticMutants",
    "package": "internal/oracle",
    "file": "internal/oracle/enums_test.go:26",
    "seconds": 0.878,
    "oracle": "Actual Node source stdout comparison must detect six built-in semantic mutants after clean native exits.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: fallthrough_test.go:83: Node \"1: one two \\n2: two \\n3: three default four \\n4: four \\n5: default four \\n6 undefined\\n1132\\n\"; native exit 70 stderr \"adamic: panic: compiler bug: a function ended without returning\\n\"",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestEnumNameEnumerationMutant",
      "TestEnumSemanticMutants",
      "TestFallthroughMutants",
      "TestFieldReadinessRepresentation"
    ],
    "evidence": "ADAMIC_MUTANT=W1 ADAMIC_BUILD_CACHE_DIR=/tmp/u060/cache/W1 ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestEnumNameEnumerationMutant|TestEnumSemanticMutants|TestFallthroughMutants|TestFieldReadinessRepresentation)$' > review/test-audit/internal-oracle-enums_open/W1.log 2>&1; fallthrough_test.go:83: Node \"1: one two \\n2: two \\n3: three default four \\n4: four \\n5: default four \\n6 undefined\\n1132\\n\"; native exit 70 stderr \"adamic: panic: compiler bug: a function ended without returning\\n\"",
    "witness_kills": [
      "W1"
    ]
  },
  {
    "test": "TestEnumCleanupMutant",
    "package": "internal/oracle",
    "file": "internal/oracle/enums_test.go:115",
    "seconds": 0.451,
    "oracle": "LeakSanitizer must detect omitted cleanup, while actual Node stdout remains unchanged.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W2: enums_test.go:148: enum cleanup mutant not caught: exit 0, stderr",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestEnumCleanupMutant"
    ],
    "evidence": "ADAMIC_MUTANT=W2 ADAMIC_BUILD_CACHE_DIR=/tmp/u060/cache/W2 ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^TestEnumCleanupMutant$' > review/test-audit/internal-oracle-enums_open/W2.log 2>&1; enums_test.go:148: enum cleanup mutant not caught: exit 0, stderr",
    "witness_kills": [
      "W2"
    ]
  },
  {
    "test": "TestFallthroughMutants",
    "package": "internal/oracle",
    "file": "internal/oracle/fallthrough_test.go:22",
    "seconds": 0.128,
    "oracle": "Actual Node stdout comparisons for isolated cases and continue-as-break; self-pinned compiler missing-return panic for the third subcase.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: fallthrough_test.go:83: Node \"1: one two \\n2: two \\n3: three default four \\n4: four \\n5: default four \\n6 undefined\\n1132\\n\"; native exit 70 stderr \"adamic: panic: compiler bug: a function ended without returning\\n\"",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestEnumNameEnumerationMutant",
      "TestEnumSemanticMutants",
      "TestFallthroughMutants",
      "TestFieldReadinessRepresentation"
    ],
    "evidence": "ADAMIC_MUTANT=W1 ADAMIC_BUILD_CACHE_DIR=/tmp/u060/cache/W1 ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestEnumNameEnumerationMutant|TestEnumSemanticMutants|TestFallthroughMutants|TestFieldReadinessRepresentation)$' > review/test-audit/internal-oracle-enums_open/W1.log 2>&1; fallthrough_test.go:83: Node \"1: one two \\n2: two \\n3: three default four \\n4: four \\n5: default four \\n6 undefined\\n1132\\n\"; native exit 70 stderr \"adamic: panic: compiler bug: a function ended without returning\\n\"",
    "witness_kills": [
      "W1"
    ]
  },
  {
    "test": "TestFieldReadinessRepresentation",
    "package": "internal/oracle",
    "file": "internal/oracle/field_readiness_test.go:9",
    "seconds": 0.532,
    "oracle": "Self-pinned readiness panic and written zero; built-in removed-readiness mutant must disagree. No external authority.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: field_readiness_test.go:41: dropping readiness check escaped pinned assertion",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestEnumNameEnumerationMutant",
      "TestEnumSemanticMutants",
      "TestFallthroughMutants",
      "TestFieldReadinessRepresentation"
    ],
    "evidence": "ADAMIC_MUTANT=W1 ADAMIC_BUILD_CACHE_DIR=/tmp/u060/cache/W1 ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestEnumNameEnumerationMutant|TestEnumSemanticMutants|TestFallthroughMutants|TestFieldReadinessRepresentation)$' > review/test-audit/internal-oracle-enums_open/W1.log 2>&1; field_readiness_test.go:41: dropping readiness check escaped pinned assertion",
    "witness_kills": [
      "W1"
    ]
  },
  {
    "test": "TestFreshWriteProbesStayRefused",
    "package": "internal/oracle",
    "file": "internal/oracle/fresh_test.go:19",
    "seconds": 0.678,
    "oracle": "Self-pinned marked cycle-write site and adamic/cycle-capable label. Node and leak checks run only if lowering accepts a probe. M3 catches removal of the label, not acceptance of a cycle.",
    "oracle_kind": "self",
    "kills": [
      "M3"
    ],
    "unique_kills": [],
    "last_proven_fail": "M3: fresh_test.go:53: refused, but not naming the marked write (\"the write at /workspace/adamic/internal/oracle/testdata/fresh_refused/ctor_wrapped.a:17:\"):",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestStage3EnumBoundaries"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 0.087,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNumericEnumNeverPathsPinned",
      "TestNumericEnumNeverPinned",
      "TestStage3EnumBoundaries",
      "TestStage3EnumSparseArrayBoundary",
      "TestEnumNameEnumerationMutant",
      "TestEnumSemanticMutants",
      "TestEnumCleanupMutant",
      "TestFallthroughMutants",
      "TestFieldReadinessRepresentation",
      "TestFreshWriteProbesStayRefused",
      "TestFunctionValueBoundaryBoxing",
      "TestImportCycleRuntimeCalls",
      "TestImportCycleLoadTimeReads",
      "TestImportedNonliteralConstCaseIsNotYet",
      "TestInputAgreesWithNode"
    ],
    "evidence": "ADAMIC_MUTANT=M3 ADAMIC_BUILD_CACHE_DIR=/tmp/u060/cache/M3 ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestNumericEnumNeverPathsPinned|TestNumericEnumNeverPinned|TestStage3EnumBoundaries|TestStage3EnumSparseArrayBoundary|TestEnumNameEnumerationMutant|TestEnumSemanticMutants|TestEnumCleanupMutant|TestFallthroughMutants|TestFieldReadinessRepresentation|TestFreshWriteProbesStayRefused|TestFunctionValueBoundaryBoxing|TestImportCycleRuntimeCalls|TestImportCycleLoadTimeReads|TestImportedNonliteralConstCaseIsNotYet|TestInputAgreesWithNode)$' > review/test-audit/internal-oracle-enums_open/M3.log 2>&1; fresh_test.go:53: refused, but not naming the marked write (\"the write at /workspace/adamic/internal/oracle/testdata/fresh_refused/ctor_wrapped.a:17:\"):"
  },
  {
    "test": "TestFunctionValueBoundaryBoxing",
    "package": "internal/oracle",
    "file": "internal/oracle/function_values_test.go:29",
    "seconds": 0.05,
    "oracle": "Self-pinned successful lowering, four callable arguments and boxed union return shapes. M2 fails during Lower with NotYet, before the structural assertions; those individual assertions remain unproven.",
    "oracle_kind": "self",
    "kills": [
      "M2"
    ],
    "unique_kills": [
      "M2"
    ],
    "last_proven_fail": "M2: function_values_test.go:37: /workspace/adamic/internal/oracle/testdata/function_values_boundary.a:16:16: stage 0 can't lower a BinaryExpression with a union of differently held members and a value yet",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNumericEnumNeverPathsPinned",
      "TestNumericEnumNeverPinned",
      "TestStage3EnumBoundaries",
      "TestStage3EnumSparseArrayBoundary",
      "TestEnumNameEnumerationMutant",
      "TestEnumSemanticMutants",
      "TestEnumCleanupMutant",
      "TestFallthroughMutants",
      "TestFieldReadinessRepresentation",
      "TestFreshWriteProbesStayRefused",
      "TestFunctionValueBoundaryBoxing",
      "TestImportCycleRuntimeCalls",
      "TestImportCycleLoadTimeReads",
      "TestImportedNonliteralConstCaseIsNotYet",
      "TestInputAgreesWithNode"
    ],
    "evidence": "ADAMIC_MUTANT=M2 ADAMIC_BUILD_CACHE_DIR=/tmp/u060/cache/M2 ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestNumericEnumNeverPathsPinned|TestNumericEnumNeverPinned|TestStage3EnumBoundaries|TestStage3EnumSparseArrayBoundary|TestEnumNameEnumerationMutant|TestEnumSemanticMutants|TestEnumCleanupMutant|TestFallthroughMutants|TestFieldReadinessRepresentation|TestFreshWriteProbesStayRefused|TestFunctionValueBoundaryBoxing|TestImportCycleRuntimeCalls|TestImportCycleLoadTimeReads|TestImportedNonliteralConstCaseIsNotYet|TestInputAgreesWithNode)$' > review/test-audit/internal-oracle-enums_open/M2.log 2>&1; function_values_test.go:37: /workspace/adamic/internal/oracle/testdata/function_values_boundary.a:16:16: stage 0 can't lower a BinaryExpression with a union of differently held members and a value yet"
  },
  {
    "test": "TestImportCycleRuntimeCalls",
    "package": "internal/oracle",
    "file": "internal/oracle/import_cycles_test.go:26",
    "seconds": 0.167,
    "oracle": "Actual independent Node runner invokes exports after module bodies; compares full result triples to native, released native and generated JavaScript, plus leak checks.",
    "oracle_kind": "external-run",
    "kills": [
      "M1"
    ],
    "unique_kills": [],
    "last_proven_fail": "M1: import_cycles_test.go:73: stdout differs",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestNumericEnumNeverPinned"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1",
      "P2",
      "P3"
    ],
    "subsumer_seconds": 0.088,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNumericEnumNeverPathsPinned",
      "TestNumericEnumNeverPinned",
      "TestStage3EnumBoundaries",
      "TestStage3EnumSparseArrayBoundary",
      "TestEnumNameEnumerationMutant",
      "TestEnumSemanticMutants",
      "TestEnumCleanupMutant",
      "TestFallthroughMutants",
      "TestFieldReadinessRepresentation",
      "TestFreshWriteProbesStayRefused",
      "TestFunctionValueBoundaryBoxing",
      "TestImportCycleRuntimeCalls",
      "TestImportCycleLoadTimeReads",
      "TestImportedNonliteralConstCaseIsNotYet",
      "TestInputAgreesWithNode"
    ],
    "evidence": "ADAMIC_MUTANT=M1 ADAMIC_BUILD_CACHE_DIR=/tmp/u060/cache/M1 ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestNumericEnumNeverPathsPinned|TestNumericEnumNeverPinned|TestStage3EnumBoundaries|TestStage3EnumSparseArrayBoundary|TestEnumNameEnumerationMutant|TestEnumSemanticMutants|TestEnumCleanupMutant|TestFallthroughMutants|TestFieldReadinessRepresentation|TestFreshWriteProbesStayRefused|TestFunctionValueBoundaryBoxing|TestImportCycleRuntimeCalls|TestImportCycleLoadTimeReads|TestImportedNonliteralConstCaseIsNotYet|TestInputAgreesWithNode)$' > review/test-audit/internal-oracle-enums_open/M1.log 2>&1; import_cycles_test.go:73: stdout differs"
  },
  {
    "test": "TestImportCycleLoadTimeReads",
    "package": "internal/oracle",
    "file": "internal/oracle/import_cycles_test.go:83",
    "seconds": 0.417,
    "oracle": "Actual Node source stdout, stderr and exit comparisons, with fixture preconditions for initialization errors.",
    "oracle_kind": "external-run",
    "kills": [
      "M1"
    ],
    "unique_kills": [],
    "last_proven_fail": "M1: import_cycles_test.go:117: exit codes differ",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestNumericEnumNeverPinned"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1",
      "P2",
      "P3"
    ],
    "subsumer_seconds": 0.088,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNumericEnumNeverPathsPinned",
      "TestNumericEnumNeverPinned",
      "TestStage3EnumBoundaries",
      "TestStage3EnumSparseArrayBoundary",
      "TestEnumNameEnumerationMutant",
      "TestEnumSemanticMutants",
      "TestEnumCleanupMutant",
      "TestFallthroughMutants",
      "TestFieldReadinessRepresentation",
      "TestFreshWriteProbesStayRefused",
      "TestFunctionValueBoundaryBoxing",
      "TestImportCycleRuntimeCalls",
      "TestImportCycleLoadTimeReads",
      "TestImportedNonliteralConstCaseIsNotYet",
      "TestInputAgreesWithNode"
    ],
    "evidence": "ADAMIC_MUTANT=M1 ADAMIC_BUILD_CACHE_DIR=/tmp/u060/cache/M1 ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestNumericEnumNeverPathsPinned|TestNumericEnumNeverPinned|TestStage3EnumBoundaries|TestStage3EnumSparseArrayBoundary|TestEnumNameEnumerationMutant|TestEnumSemanticMutants|TestEnumCleanupMutant|TestFallthroughMutants|TestFieldReadinessRepresentation|TestFreshWriteProbesStayRefused|TestFunctionValueBoundaryBoxing|TestImportCycleRuntimeCalls|TestImportCycleLoadTimeReads|TestImportedNonliteralConstCaseIsNotYet|TestInputAgreesWithNode)$' > review/test-audit/internal-oracle-enums_open/M1.log 2>&1; import_cycles_test.go:117: exit codes differ"
  },
  {
    "test": "TestImportedNonliteralConstCaseIsNotYet",
    "package": "internal/oracle",
    "file": "internal/oracle/imported_const_case_test.go:21",
    "seconds": 0.144,
    "oracle": "Actual Node source succeeds with pinned stdout; Adamic NotYet type, path, position and full wording are self-pinned.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M4"
    ],
    "unique_kills": [
      "M4"
    ],
    "last_proven_fail": "M4: imported_const_case_test.go:35: got /workspace/adamic/internal/oracle/testdata/imported_const_cases/nonliteral.a:3:7: stage 0 can't lower a case label using const WIDE with type number (not a number or string literal type) now; want /workspace/adamic/internal/oracle/testdata/imported_const_cases/nonliteral.a:3:7: stage 0 can't lower a case label using const WIDE with type number (not a number or string literal type) yet",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNumericEnumNeverPathsPinned",
      "TestNumericEnumNeverPinned",
      "TestStage3EnumBoundaries",
      "TestStage3EnumSparseArrayBoundary",
      "TestEnumNameEnumerationMutant",
      "TestEnumSemanticMutants",
      "TestEnumCleanupMutant",
      "TestFallthroughMutants",
      "TestFieldReadinessRepresentation",
      "TestFreshWriteProbesStayRefused",
      "TestFunctionValueBoundaryBoxing",
      "TestImportCycleRuntimeCalls",
      "TestImportCycleLoadTimeReads",
      "TestImportedNonliteralConstCaseIsNotYet",
      "TestInputAgreesWithNode"
    ],
    "evidence": "ADAMIC_MUTANT=M4 ADAMIC_BUILD_CACHE_DIR=/tmp/u060/cache/M4 ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestNumericEnumNeverPathsPinned|TestNumericEnumNeverPinned|TestStage3EnumBoundaries|TestStage3EnumSparseArrayBoundary|TestEnumNameEnumerationMutant|TestEnumSemanticMutants|TestEnumCleanupMutant|TestFallthroughMutants|TestFieldReadinessRepresentation|TestFreshWriteProbesStayRefused|TestFunctionValueBoundaryBoxing|TestImportCycleRuntimeCalls|TestImportCycleLoadTimeReads|TestImportedNonliteralConstCaseIsNotYet|TestInputAgreesWithNode)$' > review/test-audit/internal-oracle-enums_open/M4.log 2>&1; imported_const_case_test.go:35: got /workspace/adamic/internal/oracle/testdata/imported_const_cases/nonliteral.a:3:7: stage 0 can't lower a case label using const WIDE with type number (not a number or string literal type) now; want /workspace/adamic/internal/oracle/testdata/imported_const_cases/nonliteral.a:3:7: stage 0 can't lower a case label using const WIDE with type number (not a number or string literal type) yet"
  },
  {
    "test": "TestInputAgreesWithNode",
    "package": "internal/oracle",
    "file": "internal/oracle/input_test.go:106",
    "seconds": 0.428,
    "oracle": "Actual Node source triples plus written file names, bytes and permissions; native and JavaScript must agree, with explicit permission-denial checks and leak checks.",
    "oracle_kind": "external-run",
    "kills": [
      "M1"
    ],
    "unique_kills": [],
    "last_proven_fail": "M1: input_test.go:169: stdout differs",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestNumericEnumNeverPinned"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1",
      "P2",
      "P3"
    ],
    "subsumer_seconds": 0.088,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNumericEnumNeverPathsPinned",
      "TestNumericEnumNeverPinned",
      "TestStage3EnumBoundaries",
      "TestStage3EnumSparseArrayBoundary",
      "TestEnumNameEnumerationMutant",
      "TestEnumSemanticMutants",
      "TestEnumCleanupMutant",
      "TestFallthroughMutants",
      "TestFieldReadinessRepresentation",
      "TestFreshWriteProbesStayRefused",
      "TestFunctionValueBoundaryBoxing",
      "TestImportCycleRuntimeCalls",
      "TestImportCycleLoadTimeReads",
      "TestImportedNonliteralConstCaseIsNotYet",
      "TestInputAgreesWithNode"
    ],
    "evidence": "ADAMIC_MUTANT=M1 ADAMIC_BUILD_CACHE_DIR=/tmp/u060/cache/M1 ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestNumericEnumNeverPathsPinned|TestNumericEnumNeverPinned|TestStage3EnumBoundaries|TestStage3EnumSparseArrayBoundary|TestEnumNameEnumerationMutant|TestEnumSemanticMutants|TestEnumCleanupMutant|TestFallthroughMutants|TestFieldReadinessRepresentation|TestFreshWriteProbesStayRefused|TestFunctionValueBoundaryBoxing|TestImportCycleRuntimeCalls|TestImportCycleLoadTimeReads|TestImportedNonliteralConstCaseIsNotYet|TestInputAgreesWithNode)$' > review/test-audit/internal-oracle-enums_open/M1.log 2>&1; input_test.go:169: stdout differs"
  }
]
```

| ID | File:line at base | Change | Observed top-level failures |
|---|---|---|---|
| M1 | internal/native/emit.go:106 | drop emitter.block(program.Main, nil) | TestNumericEnumNeverPathsPinned, TestNumericEnumNeverPinned, TestStage3EnumSparseArrayBoundary, TestEnumCleanupMutant, TestFallthroughMutants, TestFieldReadinessRepresentation, TestImportCycleRuntimeCalls, TestImportCycleLoadTimeReads, TestInputAgreesWithNode |
| M2 | internal/lower/expression.go:684 | flip union boxing value.Type() != ir.Union to == | TestFunctionValueBoundaryBoxing |
| M3 | internal/lower/diagnostics.go:29 | final Refused format %s to %.0s, hiding Fix | TestStage3EnumBoundaries, TestFreshWriteProbesStayRefused |
| M4 | internal/lower/diagnostics.go:17 | NotYet suffix yet to now | TestImportedNonliteralConstCaseIsNotYet |
| W1 | internal/oracle/oracle_test.go:716 | disagreement returns empty; witness-only harness weakening | TestEnumNameEnumerationMutant, TestEnumSemanticMutants, TestFallthroughMutants, TestFieldReadinessRepresentation |
| W2 | internal/oracle/oracle_test.go:512 | force ASAN_OPTIONS=detect_leaks=0; witness-only harness weakening | TestEnumCleanupMutant |
| P1 | internal/lower/lower.go:20 | Lower returns nil,nil; probe | TestNumericEnumNeverPathsPinned, TestNumericEnumNeverPinned, TestStage3EnumBoundaries, TestStage3EnumSparseArrayBoundary, TestEnumNameEnumerationMutant, TestEnumSemanticMutants, TestEnumCleanupMutant, TestFallthroughMutants, TestFreshWriteProbesStayRefused, TestFunctionValueBoundaryBoxing, TestImportCycleRuntimeCalls, TestImportCycleLoadTimeReads, TestImportedNonliteralConstCaseIsNotYet, TestInputAgreesWithNode |
| P2 | internal/native/emit.go:22 | C returns empty string; probe | TestNumericEnumNeverPathsPinned, TestNumericEnumNeverPinned, TestStage3EnumSparseArrayBoundary, TestEnumNameEnumerationMutant, TestEnumSemanticMutants, TestEnumCleanupMutant, TestFallthroughMutants, TestFieldReadinessRepresentation, TestImportCycleRuntimeCalls, TestImportCycleLoadTimeReads, TestInputAgreesWithNode |
| P3 | internal/javascript/javascript.go:24 | JavaScript returns empty string; probe | TestNumericEnumNeverPathsPinned, TestNumericEnumNeverPinned, TestStage3EnumSparseArrayBoundary, TestFieldReadinessRepresentation, TestImportCycleRuntimeCalls, TestImportCycleLoadTimeReads, TestInputAgreesWithNode |

Survivors: none among the four production mutations in the bounded matrix. M1 witness failures are broken preconditions and are excluded from kill sets. W1/W2 and probes are excluded from production kills and uniqueness.

Brief ambiguities, costs and limitations:

- The old file reference is 8de93800f4; the required fresh main was 859dee825a4ef89f5c4ecc0f00e6620ac75c4994. Every line reference is against this actual base. All 15 names exist and remain in the named files.
- The whole-package baseline consumed 90.146 seconds without a preceding test failure, then timed out. The target slice passed clean with coverage in 6.476 seconds, and restored source passed with ADAMIC_GATE_UNCACHED=1 in 8.328 seconds. A completed whole-package baseline and kills outside the assigned slice remain unknown.
- The package is the oracle harness, while the production code under test is lowering and the backends. Mutating disagreement is permitted only for witnesses. Node, TypeScript-Go and LeakSanitizer were not changed; W2 changes harness detector configuration only.
- Five rows contain built-in mutants. Their failures under M1 are not proof that their guard works. W1 makes all four relevant witnesses fail; W2 makes cleanup fail. Fallthrough missing_implicit_return passes W1, because it guards a different runtime panic and does not call disagreement. That guard was not separately weakened.
- Four compiler mutants were used under the rebuild cap, rather than attempting the conflicting approximate three-per-row target. Two are diagnostic mutations. M3 proves refusal-label checking, not refusal safety; M4 proves wording checking. No cycle-proof guard was disabled.
- Six ordinary rows have only M1 in their kill set, and two have only M3. Their subsumption verdicts rest on one coarse mutant each. M1 drops a main emission call but retains C assembly, function bodies and global cleanup, so it is distinct from the empty-C probe. These small-set results are hints, not deletion recommendations.
- M2 is unique to the boxing row but is caught by Lower returning NotYet before its IR assertions. This proves that the row rejects that lowering regression. It cannot prove the individual argument/return shape assertions can fail.
- Fresh refusal runs do not actually run Node; the Node and leak branch runs only after unexpected acceptance. They are self-oracled. SparseArrayBoundary checks only Node exit success, while the Adamic panic expectation compares the full result triple. Stage3 checker cases check error class and code substring, not a full diagnostic location.
- Nil Lower aborts the binary. Every assigned row was rerun alone, and matrix.json uses only these isolated P1 observations. FieldReadiness does not call Lower, so its P1 pass is not a vacuity finding. Witness vacuity is null because their comparison/detector entry was judged through weakening, not a separate production entry probe.
- Empty C is rejected by native compilation/linking before runtime assertions; empty JavaScript actually runs as an empty module. These are probes only. Stage3 boundaries probe only the Lower subcase; the two upstream checker-error subcases do not reach Lower.
- Standalone early-return probes and W1 initially failed go vet for unreachable code. Their replay diffs now replace the unused bodies, with the same evaluated return behavior. Lower also drops the two imports used only by its removed body. Drafts and failed vet logs remain labeled evidence; all final replay diffs pass go vet.
- A validation call initially omitted env.sh and invoked a different go command. Those invalid-environment outputs were excluded, preserved separately, and all checks were rerun with the required toolchain.
- Go coverage records 698 reached functions across load/lower/native/javascript. It does not enumerate runtime C function coverage. Clean timings permit normal oracle caches; matrices and witness runs disable observation caches. Timings therefore describe normal warm invocation cost, not fully cold execution.
- Native rebuild measurements are for one representative enum fixture per variant, not the sum of all fixture builds in each matrix run. Matrix wall times include remaining builds and execution; those two components were not separately instrumented.

Setup: warm env.sh worked; no setup install; nproc 5. npm ci reported 0.438 seconds. Shared switched Go test compilation took 11.398 seconds. Representative native.Build times: M1 0.169s, M2 0.159s, M3 0.154s, M4 0.161s. Exact per-row three-run medians, matrix wall times, command exits and isolated-probe durations are preserved in timings.json and commands.json.

Not covered: other packages, full-package/repository uniqueness, exhaustive mutation coverage, fresh cycle acceptance, boxing assertion branches under a successful mutated Lower, Fallthrough missing-return witness weakening, upstream checker entry probes and C call coverage. Production sources were restored. No main push or pull request.
