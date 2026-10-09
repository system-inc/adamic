u071: all 13 requested tests exist at origin/main 09fe4b54913753188a9357982bfd47cdf36ef97c; two form one readiness family.
Clean whole-package baseline cooked at 90.036 seconds without observed test failures; bounded baseline passed.
Four production mutants were caught; M3 is unique only within the bounded matrix, with no survivors.
Twelve grouped rows: eight witnesses, one setup-check, one bounded sacred and two subsumed.
Tools warm; WASI SDK installed and enabled; evidence saved on the requested audit branch.

```json
[
  {
    "test": "TestTypeOfConstructorMutant",
    "package": "internal/oracle",
    "file": "internal/oracle/typeof_dispatch_test.go:25",
    "seconds": 0.282,
    "oracle": "Node source execution; full stdout, stderr and exit-code disagreement, with clean sanitizer/leak preconditions",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: typeof_dispatch_test.go:56: want Node to catch stdout alone, got \"\"",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestTypeOfConstructorMutant",
      "TestTypeOfStringLiteralMutant",
      "TestTypeOfNullMutant",
      "TestTypeOfNullSlotPresenceMutant",
      "TestUnknownNarrowingMutants",
      "TestRequiredViewFieldPrimitive",
      "TestRequiredViewFieldOperandOnce",
      "TestDefaultTaggedSourceViews",
      "TestWASIShardUnion",
      "TestWASIShardPlantedDisagreement",
      "TestWASIShardPlantedFixture",
      "TestViewFieldReadiness family"
    ],
    "evidence": "Apply W1.diff at base, or use saved selector for production/probes; ADAMIC_MUTANT=W1 ADAMIC_BUILD_CACHE_DIR=/tmp/u071/cache/W1 ADAMIC_ORACLE_WASI=1 WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestTypeOfConstructorMutant|TestTypeOfStringLiteralMutant|TestTypeOfNullMutant|TestTypeOfNullSlotPresenceMutant|TestUnknownNarrowingMutants|TestRequiredViewFieldPrimitive|TestWASIShardPlantedDisagreement|TestWASIShardPlantedFixture)$' > W1.log 2>&1; typeof_dispatch_test.go:56: want Node to catch stdout alone, got \"\"",
    "mutants_supporting_subsumption": null,
    "unique_scope": "bounded matrix; package uniqueness unknown"
  },
  {
    "test": "TestTypeOfStringLiteralMutant",
    "package": "internal/oracle",
    "file": "internal/oracle/typeof_dispatch_test.go:62",
    "seconds": 0.229,
    "oracle": "Node source execution; full stdout, stderr and exit-code disagreement, with clean sanitizer/leak preconditions",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: typeof_dispatch_test.go:93: want Node to catch stdout alone, got \"\"",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestTypeOfConstructorMutant",
      "TestTypeOfStringLiteralMutant",
      "TestTypeOfNullMutant",
      "TestTypeOfNullSlotPresenceMutant",
      "TestUnknownNarrowingMutants",
      "TestRequiredViewFieldPrimitive",
      "TestRequiredViewFieldOperandOnce",
      "TestDefaultTaggedSourceViews",
      "TestWASIShardUnion",
      "TestWASIShardPlantedDisagreement",
      "TestWASIShardPlantedFixture",
      "TestViewFieldReadiness family"
    ],
    "evidence": "Apply W1.diff at base, or use saved selector for production/probes; ADAMIC_MUTANT=W1 ADAMIC_BUILD_CACHE_DIR=/tmp/u071/cache/W1 ADAMIC_ORACLE_WASI=1 WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestTypeOfConstructorMutant|TestTypeOfStringLiteralMutant|TestTypeOfNullMutant|TestTypeOfNullSlotPresenceMutant|TestUnknownNarrowingMutants|TestRequiredViewFieldPrimitive|TestWASIShardPlantedDisagreement|TestWASIShardPlantedFixture)$' > W1.log 2>&1; typeof_dispatch_test.go:93: want Node to catch stdout alone, got \"\"",
    "mutants_supporting_subsumption": null,
    "unique_scope": "bounded matrix; package uniqueness unknown"
  },
  {
    "test": "TestTypeOfNullMutant",
    "package": "internal/oracle",
    "file": "internal/oracle/typeof_null_test.go:33",
    "seconds": 0.163,
    "oracle": "Node source execution; full stdout, stderr and exit-code disagreement, with clean sanitizer/leak preconditions",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: typeof_null_test.go:56: want Node to catch stdout alone, got \"\"",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestTypeOfConstructorMutant",
      "TestTypeOfStringLiteralMutant",
      "TestTypeOfNullMutant",
      "TestTypeOfNullSlotPresenceMutant",
      "TestUnknownNarrowingMutants",
      "TestRequiredViewFieldPrimitive",
      "TestRequiredViewFieldOperandOnce",
      "TestDefaultTaggedSourceViews",
      "TestWASIShardUnion",
      "TestWASIShardPlantedDisagreement",
      "TestWASIShardPlantedFixture",
      "TestViewFieldReadiness family"
    ],
    "evidence": "Apply W1.diff at base, or use saved selector for production/probes; ADAMIC_MUTANT=W1 ADAMIC_BUILD_CACHE_DIR=/tmp/u071/cache/W1 ADAMIC_ORACLE_WASI=1 WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestTypeOfConstructorMutant|TestTypeOfStringLiteralMutant|TestTypeOfNullMutant|TestTypeOfNullSlotPresenceMutant|TestUnknownNarrowingMutants|TestRequiredViewFieldPrimitive|TestWASIShardPlantedDisagreement|TestWASIShardPlantedFixture)$' > W1.log 2>&1; typeof_null_test.go:56: want Node to catch stdout alone, got \"\"",
    "mutants_supporting_subsumption": null,
    "unique_scope": "bounded matrix; package uniqueness unknown"
  },
  {
    "test": "TestTypeOfNullSlotPresenceMutant",
    "package": "internal/oracle",
    "file": "internal/oracle/typeof_null_test.go:111",
    "seconds": 0.3,
    "oracle": "Node source execution; full stdout, stderr and exit-code disagreement, with clean sanitizer/leak preconditions",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: typeof_null_test.go:137: want Node to catch stdout alone, got \"\"",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestTypeOfConstructorMutant",
      "TestTypeOfStringLiteralMutant",
      "TestTypeOfNullMutant",
      "TestTypeOfNullSlotPresenceMutant",
      "TestUnknownNarrowingMutants",
      "TestRequiredViewFieldPrimitive",
      "TestRequiredViewFieldOperandOnce",
      "TestDefaultTaggedSourceViews",
      "TestWASIShardUnion",
      "TestWASIShardPlantedDisagreement",
      "TestWASIShardPlantedFixture",
      "TestViewFieldReadiness family"
    ],
    "evidence": "Apply W1.diff at base, or use saved selector for production/probes; ADAMIC_MUTANT=W1 ADAMIC_BUILD_CACHE_DIR=/tmp/u071/cache/W1 ADAMIC_ORACLE_WASI=1 WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestTypeOfConstructorMutant|TestTypeOfStringLiteralMutant|TestTypeOfNullMutant|TestTypeOfNullSlotPresenceMutant|TestUnknownNarrowingMutants|TestRequiredViewFieldPrimitive|TestWASIShardPlantedDisagreement|TestWASIShardPlantedFixture)$' > W1.log 2>&1; typeof_null_test.go:137: want Node to catch stdout alone, got \"\"",
    "mutants_supporting_subsumption": null,
    "unique_scope": "bounded matrix; package uniqueness unknown"
  },
  {
    "test": "TestUnknownNarrowingMutants",
    "package": "internal/oracle",
    "file": "internal/oracle/unknown_test.go:25",
    "seconds": 0.097,
    "oracle": "Node source execution; full stdout, stderr and exit-code disagreement, with clean sanitizer/leak preconditions",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: unknown_test.go:46: want Node stdout to catch skip_inner_typeof, got \"\"",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestTypeOfConstructorMutant",
      "TestTypeOfStringLiteralMutant",
      "TestTypeOfNullMutant",
      "TestTypeOfNullSlotPresenceMutant",
      "TestUnknownNarrowingMutants",
      "TestRequiredViewFieldPrimitive",
      "TestRequiredViewFieldOperandOnce",
      "TestDefaultTaggedSourceViews",
      "TestWASIShardUnion",
      "TestWASIShardPlantedDisagreement",
      "TestWASIShardPlantedFixture",
      "TestViewFieldReadiness family"
    ],
    "evidence": "Apply W1.diff at base, or use saved selector for production/probes; ADAMIC_MUTANT=W1 ADAMIC_BUILD_CACHE_DIR=/tmp/u071/cache/W1 ADAMIC_ORACLE_WASI=1 WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestTypeOfConstructorMutant|TestTypeOfStringLiteralMutant|TestTypeOfNullMutant|TestTypeOfNullSlotPresenceMutant|TestUnknownNarrowingMutants|TestRequiredViewFieldPrimitive|TestWASIShardPlantedDisagreement|TestWASIShardPlantedFixture)$' > W1.log 2>&1; unknown_test.go:46: want Node stdout to catch skip_inner_typeof, got \"\"",
    "mutants_supporting_subsumption": null,
    "unique_scope": "bounded matrix; package uniqueness unknown"
  },
  {
    "test": "TestRequiredViewFieldPrimitive",
    "package": "internal/oracle",
    "file": "internal/oracle/view_fields_test.go:13",
    "seconds": 0.805,
    "oracle": "Node runs three positive source values; negative expected exits and exact panic text are handwritten. Built-in mutants must disagree with those self pins.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: view_fields_test.go:72: drop field check mutant escaped independent assertion",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestTypeOfConstructorMutant",
      "TestTypeOfStringLiteralMutant",
      "TestTypeOfNullMutant",
      "TestTypeOfNullSlotPresenceMutant",
      "TestUnknownNarrowingMutants",
      "TestRequiredViewFieldPrimitive",
      "TestRequiredViewFieldOperandOnce",
      "TestDefaultTaggedSourceViews",
      "TestWASIShardUnion",
      "TestWASIShardPlantedDisagreement",
      "TestWASIShardPlantedFixture",
      "TestViewFieldReadiness family"
    ],
    "evidence": "Apply W1.diff at base, or use saved selector for production/probes; ADAMIC_MUTANT=W1 ADAMIC_BUILD_CACHE_DIR=/tmp/u071/cache/W1 ADAMIC_ORACLE_WASI=1 WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestTypeOfConstructorMutant|TestTypeOfStringLiteralMutant|TestTypeOfNullMutant|TestTypeOfNullSlotPresenceMutant|TestUnknownNarrowingMutants|TestRequiredViewFieldPrimitive|TestWASIShardPlantedDisagreement|TestWASIShardPlantedFixture)$' > W1.log 2>&1; view_fields_test.go:72: drop field check mutant escaped independent assertion",
    "mutants_supporting_subsumption": null,
    "unique_scope": "bounded matrix; package uniqueness unknown"
  },
  {
    "test": "TestRequiredViewFieldOperandOnce",
    "package": "internal/oracle",
    "file": "internal/oracle/view_fields_test.go:93",
    "seconds": 0.335,
    "oracle": "Node source pins full output 7/1 and backend agreement. M2 fails on wrong field lookup, not duplicate evaluation; failure label overstates its cause.",
    "oracle_kind": "external-run",
    "kills": [
      "M2"
    ],
    "unique_kills": [],
    "last_proven_fail": "M2: view_fields_test.go:118: operand evaluated more than once: exit codes differ; got oracle.run{stdout:[]uint8{}, stderr:[]uint8{0x61, 0x64, 0x61, 0x6d, 0x69, 0x63, 0x3a, 0x20, 0x70, 0x61, 0x6e, 0x69, 0x63, 0x3a, 0x20, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x20, 0x72, 0x65, 0x61, 0x64, 0x20, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x3a, 0x20, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x20, 0x69, 0x73, 0x20, 0x6e, 0x6f, 0x74, 0x20, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x3b, 0x20, 0x65, 0x78, 0x70, 0x65, 0x63, 0x74, 0x65, 0x64, 0x20, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x2c, 0x20, 0x66, 0x6f, 0x75, 0x6e, 0x64, 0x20, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6e, 0x67, 0xa}, exitCode:70}",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestDefaultTaggedSourceViews"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "PC",
      "PJS"
    ],
    "subsumer_seconds": 0.755,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestTypeOfConstructorMutant",
      "TestTypeOfStringLiteralMutant",
      "TestTypeOfNullMutant",
      "TestTypeOfNullSlotPresenceMutant",
      "TestUnknownNarrowingMutants",
      "TestRequiredViewFieldPrimitive",
      "TestRequiredViewFieldOperandOnce",
      "TestDefaultTaggedSourceViews",
      "TestWASIShardUnion",
      "TestWASIShardPlantedDisagreement",
      "TestWASIShardPlantedFixture",
      "TestViewFieldReadiness family"
    ],
    "evidence": "Apply M2.diff at base, or use saved selector for production/probes; ADAMIC_MUTANT=M2 ADAMIC_BUILD_CACHE_DIR=/tmp/u071/cache/M2 ADAMIC_ORACLE_WASI=1 WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestTypeOfConstructorMutant|TestTypeOfStringLiteralMutant|TestTypeOfNullMutant|TestTypeOfNullSlotPresenceMutant|TestUnknownNarrowingMutants|TestRequiredViewFieldPrimitive|TestRequiredViewFieldOperandOnce|TestNarrowedFieldUsesSharedReadiness|TestViewFieldInheritedStaticReadiness|TestDefaultTaggedSourceViews|TestWASIShardUnion|TestWASIShardPlantedDisagreement|TestWASIShardPlantedFixture)$' > M2.log 2>&1; view_fields_test.go:118: operand evaluated more than once: exit codes differ; got oracle.run{stdout:[]uint8{}, stderr:[]uint8{0x61, 0x64, 0x61, 0x6d, 0x69, 0x63, 0x3a, 0x20, 0x70, 0x61, 0x6e, 0x69, 0x63, 0x3a, 0x20, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x20, 0x72, 0x65, 0x61, 0x64, 0x20, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x3a, 0x20, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x20, 0x69, 0x73, 0x20, 0x6e, 0x6f, 0x74, 0x20, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x3b, 0x20, 0x65, 0x78, 0x70, 0x65, 0x63, 0x74, 0x65, 0x64, 0x20, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x2c, 0x20, 0x66, 0x6f, 0x75, 0x6e, 0x64, 0x20, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6e, 0x67, 0xa}, exitCode:70}",
    "mutants_supporting_subsumption": 1,
    "unique_scope": "bounded matrix; package uniqueness unknown"
  },
  {
    "test": "TestDefaultTaggedSourceViews",
    "package": "internal/oracle",
    "file": "internal/oracle/view_fields_test.go:161",
    "seconds": 0.755,
    "oracle": "Node runs positive source cases. Negative field/literal panic texts and migrated eager non-null checks are handwritten self pins. Migrated source Node only must not exit 70.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M1",
      "M2",
      "M3",
      "M4"
    ],
    "unique_kills": [
      "M3"
    ],
    "last_proven_fail": "M4: view_fields_test.go:180: want eager checked undefined! with stdout \"\", got oracle.run{stdout:[]uint8{0x6f, 0x6b, 0x6f, 0x6b, 0xa}, stderr:[]uint8{}, exitCode:0}",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "PLower",
      "PC",
      "PJS"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestTypeOfConstructorMutant",
      "TestTypeOfStringLiteralMutant",
      "TestTypeOfNullMutant",
      "TestTypeOfNullSlotPresenceMutant",
      "TestUnknownNarrowingMutants",
      "TestRequiredViewFieldPrimitive",
      "TestRequiredViewFieldOperandOnce",
      "TestDefaultTaggedSourceViews",
      "TestWASIShardUnion",
      "TestWASIShardPlantedDisagreement",
      "TestWASIShardPlantedFixture",
      "TestViewFieldReadiness family"
    ],
    "evidence": "Apply M4.diff at base, or use saved selector for production/probes; ADAMIC_MUTANT=M4 ADAMIC_BUILD_CACHE_DIR=/tmp/u071/cache/M4 ADAMIC_ORACLE_WASI=1 WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestTypeOfConstructorMutant|TestTypeOfStringLiteralMutant|TestTypeOfNullMutant|TestTypeOfNullSlotPresenceMutant|TestUnknownNarrowingMutants|TestRequiredViewFieldPrimitive|TestRequiredViewFieldOperandOnce|TestNarrowedFieldUsesSharedReadiness|TestViewFieldInheritedStaticReadiness|TestDefaultTaggedSourceViews|TestWASIShardUnion|TestWASIShardPlantedDisagreement|TestWASIShardPlantedFixture)$' > M4.log 2>&1; view_fields_test.go:180: want eager checked undefined! with stdout \"\", got oracle.run{stdout:[]uint8{0x6f, 0x6b, 0x6f, 0x6b, 0xa}, stderr:[]uint8{}, exitCode:0}",
    "mutants_supporting_subsumption": null,
    "unique_scope": "bounded matrix; package uniqueness unknown"
  },
  {
    "test": "TestWASIShardUnion",
    "package": "internal/oracle",
    "file": "internal/oracle/wasi_shards_test.go:73",
    "seconds": 0.011,
    "oracle": "Self: fixture registry union, identity uniqueness and shard construction; construction loop removed in S1.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S1: wasi_shards_test.go:76: missing shard-000 [origin line; raw S1.log line 73]",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "PShards"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestTypeOfConstructorMutant",
      "TestTypeOfStringLiteralMutant",
      "TestTypeOfNullMutant",
      "TestTypeOfNullSlotPresenceMutant",
      "TestUnknownNarrowingMutants",
      "TestRequiredViewFieldPrimitive",
      "TestRequiredViewFieldOperandOnce",
      "TestDefaultTaggedSourceViews",
      "TestWASIShardUnion",
      "TestWASIShardPlantedDisagreement",
      "TestWASIShardPlantedFixture",
      "TestViewFieldReadiness family"
    ],
    "evidence": "Apply S1.diff at base, or use saved selector for production/probes; ADAMIC_MUTANT=S1 ADAMIC_BUILD_CACHE_DIR=/tmp/u071/cache/S1 ADAMIC_ORACLE_WASI=1 WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestWASIShardUnion)$' > S1.log 2>&1; wasi_shards_test.go:76: missing shard-000 [origin line; raw S1.log line 73]",
    "mutants_supporting_subsumption": null,
    "unique_scope": "bounded matrix; package uniqueness unknown"
  },
  {
    "test": "TestWASIShardPlantedDisagreement",
    "package": "internal/oracle",
    "file": "internal/oracle/wasi_shards_test.go:103",
    "seconds": 0.008,
    "oracle": "Self: synthetic agreed/disagreed stdout, planted fixture identity and exactly one rejection.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: wasi_shards_test.go:117: planted fixture lost: <nil>",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestTypeOfConstructorMutant",
      "TestTypeOfStringLiteralMutant",
      "TestTypeOfNullMutant",
      "TestTypeOfNullSlotPresenceMutant",
      "TestUnknownNarrowingMutants",
      "TestRequiredViewFieldPrimitive",
      "TestRequiredViewFieldOperandOnce",
      "TestDefaultTaggedSourceViews",
      "TestWASIShardUnion",
      "TestWASIShardPlantedDisagreement",
      "TestWASIShardPlantedFixture",
      "TestViewFieldReadiness family"
    ],
    "evidence": "Apply W1.diff at base, or use saved selector for production/probes; ADAMIC_MUTANT=W1 ADAMIC_BUILD_CACHE_DIR=/tmp/u071/cache/W1 ADAMIC_ORACLE_WASI=1 WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestTypeOfConstructorMutant|TestTypeOfStringLiteralMutant|TestTypeOfNullMutant|TestTypeOfNullSlotPresenceMutant|TestUnknownNarrowingMutants|TestRequiredViewFieldPrimitive|TestWASIShardPlantedDisagreement|TestWASIShardPlantedFixture)$' > W1.log 2>&1; wasi_shards_test.go:117: planted fixture lost: <nil>",
    "mutants_supporting_subsumption": null,
    "unique_scope": "bounded matrix; package uniqueness unknown"
  },
  {
    "test": "TestWASIShardPlantedFixture",
    "package": "internal/oracle",
    "file": "internal/oracle/wasi_shards_test.go:133",
    "seconds": 0.25,
    "oracle": "Node source compared with compiled WASI fixture; parent expects nonzero child, owning shard and stdout-difference text. Full byte comparison is exercised.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: wasi_shards_test.go:174: planted fixture was not caught in shard-001/dedication/dedication.a: <nil>",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestTypeOfConstructorMutant",
      "TestTypeOfStringLiteralMutant",
      "TestTypeOfNullMutant",
      "TestTypeOfNullSlotPresenceMutant",
      "TestUnknownNarrowingMutants",
      "TestRequiredViewFieldPrimitive",
      "TestRequiredViewFieldOperandOnce",
      "TestDefaultTaggedSourceViews",
      "TestWASIShardUnion",
      "TestWASIShardPlantedDisagreement",
      "TestWASIShardPlantedFixture",
      "TestViewFieldReadiness family"
    ],
    "evidence": "Apply W1.diff at base, or use saved selector for production/probes; ADAMIC_MUTANT=W1 ADAMIC_BUILD_CACHE_DIR=/tmp/u071/cache/W1 ADAMIC_ORACLE_WASI=1 WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestTypeOfConstructorMutant|TestTypeOfStringLiteralMutant|TestTypeOfNullMutant|TestTypeOfNullSlotPresenceMutant|TestUnknownNarrowingMutants|TestRequiredViewFieldPrimitive|TestWASIShardPlantedDisagreement|TestWASIShardPlantedFixture)$' > W1.log 2>&1; wasi_shards_test.go:174: planted fixture was not caught in shard-001/dedication/dedication.a: <nil>",
    "mutants_supporting_subsumption": null,
    "unique_scope": "bounded matrix; package uniqueness unknown"
  },
  {
    "test": "TestViewFieldReadiness family",
    "package": "internal/oracle",
    "file": "internal/oracle/view_fields_test.go:125",
    "seconds": 0.512,
    "oracle": "Handwritten eager non-null panic prefix/suffix and empty stdout; inserted-check count >0; native/JavaScript agree. Source Node only must not exit 70, so other source failures could pass that health pin.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M1",
      "M4"
    ],
    "unique_kills": [],
    "last_proven_fail": "M4: view_fields_test.go:146: want eager checked undefined! with stdout \"\", got oracle.run{stdout:[]uint8{0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0xa, 0x6f, 0x6b, 0x6f, 0x6b, 0xa}, stderr:[]uint8{}, exitCode:0}",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestDefaultTaggedSourceViews"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "PLower",
      "PC",
      "PJS"
    ],
    "subsumer_seconds": 0.755,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestTypeOfConstructorMutant",
      "TestTypeOfStringLiteralMutant",
      "TestTypeOfNullMutant",
      "TestTypeOfNullSlotPresenceMutant",
      "TestUnknownNarrowingMutants",
      "TestRequiredViewFieldPrimitive",
      "TestRequiredViewFieldOperandOnce",
      "TestDefaultTaggedSourceViews",
      "TestWASIShardUnion",
      "TestWASIShardPlantedDisagreement",
      "TestWASIShardPlantedFixture",
      "TestViewFieldReadiness family"
    ],
    "evidence": "Apply M4.diff at base, or use saved selector for production/probes; ADAMIC_MUTANT=M4 ADAMIC_BUILD_CACHE_DIR=/tmp/u071/cache/M4 ADAMIC_ORACLE_WASI=1 WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestTypeOfConstructorMutant|TestTypeOfStringLiteralMutant|TestTypeOfNullMutant|TestTypeOfNullSlotPresenceMutant|TestUnknownNarrowingMutants|TestRequiredViewFieldPrimitive|TestRequiredViewFieldOperandOnce|TestNarrowedFieldUsesSharedReadiness|TestViewFieldInheritedStaticReadiness|TestDefaultTaggedSourceViews|TestWASIShardUnion|TestWASIShardPlantedDisagreement|TestWASIShardPlantedFixture)$' > M4.log 2>&1; view_fields_test.go:146: want eager checked undefined! with stdout \"\", got oracle.run{stdout:[]uint8{0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0xa, 0x6f, 0x6b, 0x6f, 0x6b, 0xa}, stderr:[]uint8{}, exitCode:0}",
    "mutants_supporting_subsumption": 2,
    "unique_scope": "bounded matrix; package uniqueness unknown",
    "members": [
      {
        "test": "TestNarrowedFieldUsesSharedReadiness",
        "file": "internal/oracle/view_fields_test.go:125"
      },
      {
        "test": "TestViewFieldInheritedStaticReadiness",
        "file": "internal/oracle/view_fields_test.go:151"
      }
    ]
  }
]
```

| ID | Origin file:line | Change | Raw top-level failures |
|---|---|---|---|
| M1 | internal/lower/non_null.go:96 | change non-null assertion diagnostic suffix | TestViewFieldInheritedStaticReadiness, TestNarrowedFieldUsesSharedReadiness, TestDefaultTaggedSourceViews |
| M2 | internal/native/view_fields.go:31 | swap checked-field lookup name and diagnostic name | TestRequiredViewFieldOperandOnce, TestRequiredViewFieldPrimitive, TestDefaultTaggedSourceViews |
| M3 | internal/javascript/view_fields.go:22 | drop ordinary checked-field literal restrictions | TestDefaultTaggedSourceViews |
| M4 | internal/native/emit_branches.go:146 | treat literal undefined/null as present in coalescing | TestDefaultTaggedSourceViews, TestNarrowedFieldUsesSharedReadiness |

Survivors: none. All four fixed-menu production mutants changed exercised behavior and were caught. W1 disables disagreement at internal/oracle/oracle_test.go:716; all eight witnesses fail. S1 drops the entire distribution loop at internal/oracle/wasi_shards_test.go:30; TestWASIShardUnion fails with missing shard-000. Its raw log line 73 maps to origin line 76 after the three-line loop deletion. Probe diffs and logs are separate and never count toward production kills or uniqueness.

Brief ambiguities, corrections and costs:
- The brief cites files at 8de93800f4; current origin/main is the recorded base. All 13 names remained in their named files; none moved or vanished.
- The two readiness wrappers call the same assertMigratedNonNullCheck with different fixture/expression inputs and no additional assertions. They are one family, timed together three times. It has two members, not two independent uniqueness rows.
- A family verdict must follow grouped kills. M4 catches one readiness member and not the inherited-static member; that still counts as one family kill. No member subsumes its own family.
- RequiredViewFieldPrimitive combines normal controls with built-in check-removal mutants. The whole top-level row is a witness. Its M2 production failure is a broken precondition and is excluded from production kills.
- Constructor, string, null, slot-presence and unknown rows are also witnesses. Their existing mutants remain unchanged. Their verdict rests solely on the allowed weakened comparison, not compiler-mutant failures.
- The WASI union is suite construction. S1 removes the whole loop, keeping standalone Go vet valid without unused ordinal or row variables. PShards is a separate empty-construction probe.
- The at-most-four compiler rebuilding instruction conflicts with a target of three mutants per row. Four spread mutations were selected from reached functions before checking failures; no claim extends beyond this menu.
- The package run exceeded its 90-second test budget. All matrices ran the 13 requested test functions, grouped into 12 rows. Kills outside this slice and package/repository uniqueness are unknown. Sacred here is explicitly bounded.
- No clean top-level or subcase skip was observed in the requested slice. WASI was enabled with an installed SDK, including the real planted-fixture child. The interrupted whole-package baseline is not a complete skip inventory.
- The names NarrowedFieldUsesSharedReadiness and ViewFieldInheritedStaticReadiness no longer describe the executed boundary precisely: their current sources are .ts and stop at an eager non-null initializer. The oracle is partly handwritten check text and metadata, not a Node prediction of Adamic's extra checks.
- Migrated source health checks reject only Node exit 70. Other source failure codes would pass that check. Native expectations do pin stdout and diagnostic prefix/suffix; backend-to-backend agreement by itself is a self oracle.
- OperandOnce compares complete output with Node, including the counter, but M2 catches a wrong lookup name, not repeated operand evaluation. Its fixed failure label says more than the particular failing observation establishes.
- M1 is diagnostic text. It demonstrates exact-message enforcement, not changed runtime control flow. M4 does change control flow and observed stdout/exit. Subsumption rests on two mutants for the readiness family and one for OperandOnce; it is a hint, not a deletion verdict.
- No witness production entry was empty-probed because the brief directs witness judgment through the comparison it guards. Every production entry used by the three production rows was probed; individual family members were run separately to avoid one Lower nil panic hiding the other.
- Lower's empty answer leads to Go panics in callers. Those isolated failures are recorded as probe kills only; they do not prove useful production discrimination. PC/PJS also fail the selected rows. No positive subcase accepting an empty answer was observed.
- Go coverage lists every reached CUT function before mutation selection. Runtime C functions were not instrumented, so that inventory does not claim exact C reachability.
- An initial timing command used unavailable /usr/bin/time and did not run installation/baseline. It was corrected to Bash time before the actual baseline. Both actual installation durations are in wasi-install.log.
- Warm Go/Node/clang tools did not include the WASI SDK. Only the optional SDK was installed, not a fresh cloud setup. stage3/api npm ci ran before baseline; requested tests use repository Node runners rather than another node_modules directory.
- Native caches were distinct per mutant. Matrix wall durations combine product rebuilding and execution; exclusive native rebuild duration was not isolated. Oracle caches retain keyed source observations. Production source/IR and emitted products are regenerated and rechecked.
- No full repository gate or other package test suite was run. No oracle or fixture was mutated in the production menu. W1/S1/PShards are explicitly allowed witness/construction edits, kept separate. No pull request or main push is part of this audit.

Timing: setup.sh skipped (warm tools), nproc 5. SDK download 1.824 s, extraction 2.509 s. npm: added 3 packages in 423ms. Whole-package baseline binary 90.036 s, wall 92.838 s; bounded baseline wall 5.32 s. Coverage wall 10.565 s. The 36 isolated timing commands totaled 82.997 s wall; medians of own binary lines are in the JSON. Selector binary build 13.631 s. Cold per-mutant matrix/rebuild command walls: M1 5.62 s, M2 5.336 s, M3 5.764 s, M4 5.732 s. Vet, witnesses, construction check, selector build, matrices, probes and final verification together took 88.011 s command wall. Final bounded baseline passed in 4.233 s wall. Human/tool overhead not included. Full timing records: clean-runs.json and audit-runs.json.
