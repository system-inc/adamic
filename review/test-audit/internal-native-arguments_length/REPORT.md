u045: base e8cb7ac7478d52c3cc847a1163518b5849f7cc87; nproc 5.
All 14 requested functions exist in their named files; no family grouping was needed.
Whole-package baseline cooked at 90.023s; clean scoped baseline passed in 18.635s.
Bounded verdicts: 10 sacred, 4 witness; no package-wide uniqueness is claimed.
20 production mutants, 2 weakened checks, 10 entry probes; production restored.

```json
[
  {
    "row_id": "R1",
    "test": "TestNoReaderCallingConvention",
    "package": "internal/native",
    "file": "internal/native/arguments_length_test.go",
    "seconds": 0.082,
    "oracle": "Handwritten production-plan or emitted-C assertions",
    "oracle_kind": "self",
    "kills": [
      "M01"
    ],
    "unique_kills": [
      "M01"
    ],
    "last_proven_fail": "M01: arguments_length_test.go:38: counted convention: got true, want false",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P_C"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNoReaderCallingConvention",
      "TestBorrowChainDeclarations",
      "TestBorrowChainTargets",
      "TestPassThroughsAreNotConsumers",
      "TestCaseMappingMatchesNode",
      "TestCaseTablesMatchNodesUnicode",
      "TestInheritanceMemoryPlans",
      "TestClosureConventionDropCount",
      "TestClosureConventionRuntimeDropCount",
      "TestClosureConventionRuntimeFeaturesIgnoreLiterals",
      "TestClosureConventionWrongOrder",
      "TestParserHasNoUnusedOptionalMethodThunks",
      "TestOptionalMethodThunksMatchNode",
      "TestArithmeticIsNeverFused"
    ],
    "evidence": "ADAMIC_MUTANT=M01 ADAMIC_BUILD_CACHE_DIR=/tmp/u045/cache/M01 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestNoReaderCallingConvention|TestBorrowChainDeclarations|TestBorrowChainTargets|TestPassThroughsAreNotConsumers|TestCaseMappingMatchesNode|TestCaseTablesMatchNodesUnicode|TestInheritanceMemoryPlans|TestClosureConventionDropCount|TestClosureConventionRuntimeDropCount|TestClosureConventionRuntimeFeaturesIgnoreLiterals|TestClosureConventionWrongOrder|TestParserHasNoUnusedOptionalMethodThunks|TestOptionalMethodThunksMatchNode|TestArithmeticIsNeverFused)$' > M01.log 2>&1; arguments_length_test.go:38: counted convention: got true, want false",
    "vacuous_subcases": [
      "TestNoReaderCallingConvention/closure_convention_plain.a"
    ]
  },
  {
    "row_id": "R2",
    "test": "TestBorrowChainDeclarations",
    "package": "internal/native",
    "file": "internal/native/borrow_chain_test.go",
    "seconds": 0.149,
    "oracle": "Handwritten production-plan or emitted-C assertions",
    "oracle_kind": "self",
    "kills": [
      "M07",
      "M08",
      "M10",
      "M11"
    ],
    "unique_kills": [
      "M07",
      "M10",
      "M11"
    ],
    "last_proven_fail": "M11: borrow_chain_test.go:39: walk: parent did not borrow",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P_BORROW"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNoReaderCallingConvention",
      "TestBorrowChainDeclarations",
      "TestBorrowChainTargets",
      "TestPassThroughsAreNotConsumers",
      "TestCaseMappingMatchesNode",
      "TestCaseTablesMatchNodesUnicode",
      "TestInheritanceMemoryPlans",
      "TestClosureConventionDropCount",
      "TestClosureConventionRuntimeDropCount",
      "TestClosureConventionRuntimeFeaturesIgnoreLiterals",
      "TestClosureConventionWrongOrder",
      "TestParserHasNoUnusedOptionalMethodThunks",
      "TestOptionalMethodThunksMatchNode",
      "TestArithmeticIsNeverFused"
    ],
    "evidence": "ADAMIC_MUTANT=M11 ADAMIC_BUILD_CACHE_DIR=/tmp/u045/cache/M11 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestNoReaderCallingConvention|TestBorrowChainDeclarations|TestBorrowChainTargets|TestPassThroughsAreNotConsumers|TestCaseMappingMatchesNode|TestCaseTablesMatchNodesUnicode|TestInheritanceMemoryPlans|TestClosureConventionDropCount|TestClosureConventionRuntimeDropCount|TestClosureConventionRuntimeFeaturesIgnoreLiterals|TestClosureConventionWrongOrder|TestParserHasNoUnusedOptionalMethodThunks|TestOptionalMethodThunksMatchNode|TestArithmeticIsNeverFused)$' > M11.log 2>&1; borrow_chain_test.go:39: walk: parent did not borrow"
  },
  {
    "row_id": "R3",
    "test": "TestBorrowChainTargets",
    "package": "internal/native",
    "file": "internal/native/borrow_chain_test.go",
    "seconds": 0.008,
    "oracle": "Handwritten production-plan or emitted-C assertions",
    "oracle_kind": "self",
    "kills": [
      "M08",
      "M09"
    ],
    "unique_kills": [
      "M09"
    ],
    "last_proven_fail": "M09: borrow_chain_test.go:62: unknown callback considered read-only",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P_CHAIN"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNoReaderCallingConvention",
      "TestBorrowChainDeclarations",
      "TestBorrowChainTargets",
      "TestPassThroughsAreNotConsumers",
      "TestCaseMappingMatchesNode",
      "TestCaseTablesMatchNodesUnicode",
      "TestInheritanceMemoryPlans",
      "TestClosureConventionDropCount",
      "TestClosureConventionRuntimeDropCount",
      "TestClosureConventionRuntimeFeaturesIgnoreLiterals",
      "TestClosureConventionWrongOrder",
      "TestParserHasNoUnusedOptionalMethodThunks",
      "TestOptionalMethodThunksMatchNode",
      "TestArithmeticIsNeverFused"
    ],
    "evidence": "ADAMIC_MUTANT=M09 ADAMIC_BUILD_CACHE_DIR=/tmp/u045/cache/M09 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestNoReaderCallingConvention|TestBorrowChainDeclarations|TestBorrowChainTargets|TestPassThroughsAreNotConsumers|TestCaseMappingMatchesNode|TestCaseTablesMatchNodesUnicode|TestInheritanceMemoryPlans|TestClosureConventionDropCount|TestClosureConventionRuntimeDropCount|TestClosureConventionRuntimeFeaturesIgnoreLiterals|TestClosureConventionWrongOrder|TestParserHasNoUnusedOptionalMethodThunks|TestOptionalMethodThunksMatchNode|TestArithmeticIsNeverFused)$' > M09.log 2>&1; borrow_chain_test.go:62: unknown callback considered read-only"
  },
  {
    "row_id": "R4",
    "test": "TestPassThroughsAreNotConsumers",
    "package": "internal/native",
    "file": "internal/native/borrow_consumes_test.go",
    "seconds": 0.01,
    "oracle": "Handwritten production-plan or emitted-C assertions",
    "oracle_kind": "self",
    "kills": [
      "M05"
    ],
    "unique_kills": [
      "M05"
    ],
    "last_proven_fail": "M05: borrow_consumes_test.go:39: ir.Coalesce hands an operand on as its own value, but consumes says it may lend it",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P_CONSUMES"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNoReaderCallingConvention",
      "TestBorrowChainDeclarations",
      "TestBorrowChainTargets",
      "TestPassThroughsAreNotConsumers",
      "TestCaseMappingMatchesNode",
      "TestCaseTablesMatchNodesUnicode",
      "TestInheritanceMemoryPlans",
      "TestClosureConventionDropCount",
      "TestClosureConventionRuntimeDropCount",
      "TestClosureConventionRuntimeFeaturesIgnoreLiterals",
      "TestClosureConventionWrongOrder",
      "TestParserHasNoUnusedOptionalMethodThunks",
      "TestOptionalMethodThunksMatchNode",
      "TestArithmeticIsNeverFused"
    ],
    "evidence": "ADAMIC_MUTANT=M05 ADAMIC_BUILD_CACHE_DIR=/tmp/u045/cache/M05 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestNoReaderCallingConvention|TestBorrowChainDeclarations|TestBorrowChainTargets|TestPassThroughsAreNotConsumers|TestCaseMappingMatchesNode|TestCaseTablesMatchNodesUnicode|TestInheritanceMemoryPlans|TestClosureConventionDropCount|TestClosureConventionRuntimeDropCount|TestClosureConventionRuntimeFeaturesIgnoreLiterals|TestClosureConventionWrongOrder|TestParserHasNoUnusedOptionalMethodThunks|TestOptionalMethodThunksMatchNode|TestArithmeticIsNeverFused)$' > M05.log 2>&1; borrow_consumes_test.go:39: ir.Coalesce hands an operand on as its own value, but consumes says it may lend it"
  },
  {
    "row_id": "R5",
    "test": "TestCaseMappingMatchesNode",
    "package": "internal/native",
    "file": "internal/native/case_test.go",
    "seconds": 12.626,
    "oracle": "Native WTF-8 mapping compared byte-for-byte with Node toUpperCase/toLowerCase; handwritten sweep line counts",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M18",
      "M19"
    ],
    "unique_kills": [
      "M18",
      "M19"
    ],
    "last_proven_fail": "M19: case_test.go:234: native \"41cea3 61cf83 41cea3\"; Node   \"41cea3 61cf82 41cea3\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P_CASE_LOWER",
      "P_CASE_UPPER"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNoReaderCallingConvention",
      "TestBorrowChainDeclarations",
      "TestBorrowChainTargets",
      "TestPassThroughsAreNotConsumers",
      "TestCaseMappingMatchesNode",
      "TestCaseTablesMatchNodesUnicode",
      "TestInheritanceMemoryPlans",
      "TestClosureConventionDropCount",
      "TestClosureConventionRuntimeDropCount",
      "TestClosureConventionRuntimeFeaturesIgnoreLiterals",
      "TestClosureConventionWrongOrder",
      "TestParserHasNoUnusedOptionalMethodThunks",
      "TestOptionalMethodThunksMatchNode",
      "TestArithmeticIsNeverFused"
    ],
    "evidence": "ADAMIC_MUTANT=M19 ADAMIC_BUILD_CACHE_DIR=/tmp/u045/cache/M19 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestNoReaderCallingConvention|TestBorrowChainDeclarations|TestBorrowChainTargets|TestPassThroughsAreNotConsumers|TestCaseMappingMatchesNode|TestCaseTablesMatchNodesUnicode|TestInheritanceMemoryPlans|TestClosureConventionDropCount|TestClosureConventionRuntimeDropCount|TestClosureConventionRuntimeFeaturesIgnoreLiterals|TestClosureConventionWrongOrder|TestParserHasNoUnusedOptionalMethodThunks|TestOptionalMethodThunksMatchNode|TestArithmeticIsNeverFused)$' > M19.log 2>&1; case_test.go:234: native \"41cea3 61cf83 41cea3\"; Node   \"41cea3 61cf82 41cea3\""
  },
  {
    "row_id": "R6",
    "test": "TestCaseTablesMatchNodesUnicode",
    "package": "internal/native",
    "file": "internal/native/case_test.go",
    "seconds": 0.042,
    "oracle": "Node process.versions.unicode compared with production Unicode metadata",
    "oracle_kind": "external-run",
    "kills": [
      "M20"
    ],
    "unique_kills": [
      "M20"
    ],
    "last_proven_fail": "M20: case_test.go:335: case_tables.h is Unicode 16.0.0 and Node is Unicode 17.0: update case_generate.go and run go generate",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestNoReaderCallingConvention",
      "TestBorrowChainDeclarations",
      "TestBorrowChainTargets",
      "TestPassThroughsAreNotConsumers",
      "TestCaseMappingMatchesNode",
      "TestCaseTablesMatchNodesUnicode",
      "TestInheritanceMemoryPlans",
      "TestClosureConventionDropCount",
      "TestClosureConventionRuntimeDropCount",
      "TestClosureConventionRuntimeFeaturesIgnoreLiterals",
      "TestClosureConventionWrongOrder",
      "TestParserHasNoUnusedOptionalMethodThunks",
      "TestOptionalMethodThunksMatchNode",
      "TestArithmeticIsNeverFused"
    ],
    "evidence": "ADAMIC_MUTANT=M20 ADAMIC_BUILD_CACHE_DIR=/tmp/u045/cache/M20 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestNoReaderCallingConvention|TestBorrowChainDeclarations|TestBorrowChainTargets|TestPassThroughsAreNotConsumers|TestCaseMappingMatchesNode|TestCaseTablesMatchNodesUnicode|TestInheritanceMemoryPlans|TestClosureConventionDropCount|TestClosureConventionRuntimeDropCount|TestClosureConventionRuntimeFeaturesIgnoreLiterals|TestClosureConventionWrongOrder|TestParserHasNoUnusedOptionalMethodThunks|TestOptionalMethodThunksMatchNode|TestArithmeticIsNeverFused)$' > M20.log 2>&1; case_test.go:335: case_tables.h is Unicode 16.0.0 and Node is Unicode 17.0: update case_generate.go and run go generate"
  },
  {
    "row_id": "R7",
    "test": "TestInheritanceMemoryPlans",
    "package": "internal/native",
    "file": "internal/native/class_inheritance_test.go",
    "seconds": 0.042,
    "oracle": "Handwritten production-plan or emitted-C assertions; empty element-borrow plan passes while empty region and reuse plans fail",
    "oracle_kind": "self",
    "kills": [
      "M12",
      "M14"
    ],
    "unique_kills": [
      "M12",
      "M14"
    ],
    "last_proven_fail": "M14: class_inheritance_test.go:69: ChildReader_replace did not join consumed argument",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P_REGIONS",
      "P_REUSE"
    ],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestNoReaderCallingConvention",
      "TestBorrowChainDeclarations",
      "TestBorrowChainTargets",
      "TestPassThroughsAreNotConsumers",
      "TestCaseMappingMatchesNode",
      "TestCaseTablesMatchNodesUnicode",
      "TestInheritanceMemoryPlans",
      "TestClosureConventionDropCount",
      "TestClosureConventionRuntimeDropCount",
      "TestClosureConventionRuntimeFeaturesIgnoreLiterals",
      "TestClosureConventionWrongOrder",
      "TestParserHasNoUnusedOptionalMethodThunks",
      "TestOptionalMethodThunksMatchNode",
      "TestArithmeticIsNeverFused"
    ],
    "evidence": "ADAMIC_MUTANT=M14 ADAMIC_BUILD_CACHE_DIR=/tmp/u045/cache/M14 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestNoReaderCallingConvention|TestBorrowChainDeclarations|TestBorrowChainTargets|TestPassThroughsAreNotConsumers|TestCaseMappingMatchesNode|TestCaseTablesMatchNodesUnicode|TestInheritanceMemoryPlans|TestClosureConventionDropCount|TestClosureConventionRuntimeDropCount|TestClosureConventionRuntimeFeaturesIgnoreLiterals|TestClosureConventionWrongOrder|TestParserHasNoUnusedOptionalMethodThunks|TestOptionalMethodThunksMatchNode|TestArithmeticIsNeverFused)$' > M14.log 2>&1; class_inheritance_test.go:69: ChildReader_replace did not join consumed argument",
    "vacuous_entries": [
      "planElementBorrows"
    ]
  },
  {
    "row_id": "R8",
    "test": "TestClosureConventionDropCount",
    "package": "internal/native",
    "file": "internal/native/closure_convention_test.go",
    "seconds": 0.163,
    "oracle": "Clang rejects planted typed-ABI violations; handwritten diagnostic substrings",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W_BUILD: closure_convention_test.go:52: drop-count mutant was not rejected for typed arity: <nil>",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestNoReaderCallingConvention",
      "TestBorrowChainDeclarations",
      "TestBorrowChainTargets",
      "TestPassThroughsAreNotConsumers",
      "TestCaseMappingMatchesNode",
      "TestCaseTablesMatchNodesUnicode",
      "TestInheritanceMemoryPlans",
      "TestClosureConventionDropCount",
      "TestClosureConventionRuntimeDropCount",
      "TestClosureConventionRuntimeFeaturesIgnoreLiterals",
      "TestClosureConventionWrongOrder",
      "TestParserHasNoUnusedOptionalMethodThunks",
      "TestOptionalMethodThunksMatchNode",
      "TestArithmeticIsNeverFused"
    ],
    "evidence": "ADAMIC_MUTANT=W_BUILD ADAMIC_BUILD_CACHE_DIR=/tmp/u045/cache/W_BUILD timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestClosureConventionDropCount|TestClosureConventionRuntimeDropCount|TestClosureConventionWrongOrder)$' > W_BUILD.log 2>&1; closure_convention_test.go:52: drop-count mutant was not rejected for typed arity: <nil>",
    "witness_check": "W_BUILD"
  },
  {
    "row_id": "R9",
    "test": "TestClosureConventionRuntimeDropCount",
    "package": "internal/native",
    "file": "internal/native/closure_convention_test.go",
    "seconds": 0.086,
    "oracle": "Clang rejects planted typed-ABI violations; handwritten diagnostic substrings",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W_BUILD: closure_convention_test.go:81: runtime drop-count mutant escaped the typed convention: <nil>",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestNoReaderCallingConvention",
      "TestBorrowChainDeclarations",
      "TestBorrowChainTargets",
      "TestPassThroughsAreNotConsumers",
      "TestCaseMappingMatchesNode",
      "TestCaseTablesMatchNodesUnicode",
      "TestInheritanceMemoryPlans",
      "TestClosureConventionDropCount",
      "TestClosureConventionRuntimeDropCount",
      "TestClosureConventionRuntimeFeaturesIgnoreLiterals",
      "TestClosureConventionWrongOrder",
      "TestParserHasNoUnusedOptionalMethodThunks",
      "TestOptionalMethodThunksMatchNode",
      "TestArithmeticIsNeverFused"
    ],
    "evidence": "ADAMIC_MUTANT=W_BUILD ADAMIC_BUILD_CACHE_DIR=/tmp/u045/cache/W_BUILD timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestClosureConventionDropCount|TestClosureConventionRuntimeDropCount|TestClosureConventionWrongOrder)$' > W_BUILD.log 2>&1; closure_convention_test.go:81: runtime drop-count mutant escaped the typed convention: <nil>",
    "witness_check": "W_BUILD"
  },
  {
    "row_id": "R10",
    "test": "TestClosureConventionRuntimeFeaturesIgnoreLiterals",
    "package": "internal/native",
    "file": "internal/native/closure_convention_test.go",
    "seconds": 0.023,
    "oracle": "Handwritten absence of unused feature macros in built runtime header; empty emitted C passes",
    "oracle_kind": "self",
    "kills": [
      "M02"
    ],
    "unique_kills": [
      "M02"
    ],
    "last_proven_fail": "M02: closure_convention_test.go:99: a source literal selected unused runtime support: #define ADAMIC_REGEXP_REPLACE_CALLBACK 1",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P_LIBRARY"
    ],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestNoReaderCallingConvention",
      "TestBorrowChainDeclarations",
      "TestBorrowChainTargets",
      "TestPassThroughsAreNotConsumers",
      "TestCaseMappingMatchesNode",
      "TestCaseTablesMatchNodesUnicode",
      "TestInheritanceMemoryPlans",
      "TestClosureConventionDropCount",
      "TestClosureConventionRuntimeDropCount",
      "TestClosureConventionRuntimeFeaturesIgnoreLiterals",
      "TestClosureConventionWrongOrder",
      "TestParserHasNoUnusedOptionalMethodThunks",
      "TestOptionalMethodThunksMatchNode",
      "TestArithmeticIsNeverFused"
    ],
    "evidence": "ADAMIC_MUTANT=M02 ADAMIC_BUILD_CACHE_DIR=/tmp/u045/cache/M02 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestNoReaderCallingConvention|TestBorrowChainDeclarations|TestBorrowChainTargets|TestPassThroughsAreNotConsumers|TestCaseMappingMatchesNode|TestCaseTablesMatchNodesUnicode|TestInheritanceMemoryPlans|TestClosureConventionDropCount|TestClosureConventionRuntimeDropCount|TestClosureConventionRuntimeFeaturesIgnoreLiterals|TestClosureConventionWrongOrder|TestParserHasNoUnusedOptionalMethodThunks|TestOptionalMethodThunksMatchNode|TestArithmeticIsNeverFused)$' > M02.log 2>&1; closure_convention_test.go:99: a source literal selected unused runtime support: #define ADAMIC_REGEXP_REPLACE_CALLBACK 1",
    "vacuous_entries": [
      "C"
    ]
  },
  {
    "row_id": "R11",
    "test": "TestClosureConventionWrongOrder",
    "package": "internal/native",
    "file": "internal/native/closure_convention_test.go",
    "seconds": 0.105,
    "oracle": "Clang rejects planted typed-ABI violations; handwritten diagnostic substrings",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W_BUILD: closure_convention_test.go:126: swapped definition escaped the header-derived declaration: <nil>",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestNoReaderCallingConvention",
      "TestBorrowChainDeclarations",
      "TestBorrowChainTargets",
      "TestPassThroughsAreNotConsumers",
      "TestCaseMappingMatchesNode",
      "TestCaseTablesMatchNodesUnicode",
      "TestInheritanceMemoryPlans",
      "TestClosureConventionDropCount",
      "TestClosureConventionRuntimeDropCount",
      "TestClosureConventionRuntimeFeaturesIgnoreLiterals",
      "TestClosureConventionWrongOrder",
      "TestParserHasNoUnusedOptionalMethodThunks",
      "TestOptionalMethodThunksMatchNode",
      "TestArithmeticIsNeverFused"
    ],
    "evidence": "ADAMIC_MUTANT=W_BUILD ADAMIC_BUILD_CACHE_DIR=/tmp/u045/cache/W_BUILD timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestClosureConventionDropCount|TestClosureConventionRuntimeDropCount|TestClosureConventionWrongOrder)$' > W_BUILD.log 2>&1; closure_convention_test.go:126: swapped definition escaped the header-derived declaration: <nil>",
    "witness_check": "W_BUILD"
  },
  {
    "row_id": "R12",
    "test": "TestParserHasNoUnusedOptionalMethodThunks",
    "package": "internal/native",
    "file": "internal/native/closure_convention_test.go",
    "seconds": 6.187,
    "oracle": "Handwritten absence of three optional-thunk declarations and IR fixture guards; empty C output passes",
    "oracle_kind": "self",
    "kills": [
      "M15"
    ],
    "unique_kills": [
      "M15"
    ],
    "last_proven_fail": "M15: closure_convention_test.go:193: unused optional method thunk: Parser_type",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestNoReaderCallingConvention",
      "TestBorrowChainDeclarations",
      "TestBorrowChainTargets",
      "TestPassThroughsAreNotConsumers",
      "TestCaseMappingMatchesNode",
      "TestCaseTablesMatchNodesUnicode",
      "TestInheritanceMemoryPlans",
      "TestClosureConventionDropCount",
      "TestClosureConventionRuntimeDropCount",
      "TestClosureConventionRuntimeFeaturesIgnoreLiterals",
      "TestClosureConventionWrongOrder",
      "TestParserHasNoUnusedOptionalMethodThunks",
      "TestOptionalMethodThunksMatchNode",
      "TestArithmeticIsNeverFused"
    ],
    "evidence": "ADAMIC_MUTANT=M15 ADAMIC_BUILD_CACHE_DIR=/tmp/u045/cache/M15 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestNoReaderCallingConvention|TestBorrowChainDeclarations|TestBorrowChainTargets|TestPassThroughsAreNotConsumers|TestCaseMappingMatchesNode|TestCaseTablesMatchNodesUnicode|TestInheritanceMemoryPlans|TestClosureConventionDropCount|TestClosureConventionRuntimeDropCount|TestClosureConventionRuntimeFeaturesIgnoreLiterals|TestClosureConventionWrongOrder|TestParserHasNoUnusedOptionalMethodThunks|TestOptionalMethodThunksMatchNode|TestArithmeticIsNeverFused)$' > M15.log 2>&1; closure_convention_test.go:193: unused optional method thunk: Parser_type",
    "vacuous_entries": [
      "C"
    ]
  },
  {
    "row_id": "R13",
    "test": "TestOptionalMethodThunksMatchNode",
    "package": "internal/native",
    "file": "internal/native/closure_convention_test.go",
    "seconds": 0.417,
    "oracle": "Node/native agreement plus handwritten convention/thunk checks; planted missing-thunk disagreement witness",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W_NODE: closure_convention_test.go:255: Node failed to catch the omitted required method thunk",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestNoReaderCallingConvention",
      "TestBorrowChainDeclarations",
      "TestBorrowChainTargets",
      "TestPassThroughsAreNotConsumers",
      "TestCaseMappingMatchesNode",
      "TestCaseTablesMatchNodesUnicode",
      "TestInheritanceMemoryPlans",
      "TestClosureConventionDropCount",
      "TestClosureConventionRuntimeDropCount",
      "TestClosureConventionRuntimeFeaturesIgnoreLiterals",
      "TestClosureConventionWrongOrder",
      "TestParserHasNoUnusedOptionalMethodThunks",
      "TestOptionalMethodThunksMatchNode",
      "TestArithmeticIsNeverFused"
    ],
    "evidence": "ADAMIC_MUTANT=W_NODE ADAMIC_BUILD_CACHE_DIR=/tmp/u045/cache/W_NODE timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestOptionalMethodThunksMatchNode)$' > W_NODE.log 2>&1; closure_convention_test.go:255: Node failed to catch the omitted required method thunk",
    "witness_check": "W_NODE"
  },
  {
    "row_id": "R14",
    "test": "TestArithmeticIsNeverFused",
    "package": "internal/native",
    "file": "internal/native/contract_test.go",
    "seconds": 0.29,
    "oracle": "Handwritten 0 0 0 output; separately built clang fast-contraction control proves FMA is observable",
    "oracle_kind": "self",
    "kills": [
      "M04"
    ],
    "unique_kills": [
      "M04"
    ],
    "last_proven_fail": "M04: contract_test.go:86: sanitize false: got \"5.5511151231257827e-17 -5.5511151231257827e-17 0\\n\", want \"0 0 0\\n\", as JavaScript computes it",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P_BUILD"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNoReaderCallingConvention",
      "TestBorrowChainDeclarations",
      "TestBorrowChainTargets",
      "TestPassThroughsAreNotConsumers",
      "TestCaseMappingMatchesNode",
      "TestCaseTablesMatchNodesUnicode",
      "TestInheritanceMemoryPlans",
      "TestClosureConventionDropCount",
      "TestClosureConventionRuntimeDropCount",
      "TestClosureConventionRuntimeFeaturesIgnoreLiterals",
      "TestClosureConventionWrongOrder",
      "TestParserHasNoUnusedOptionalMethodThunks",
      "TestOptionalMethodThunksMatchNode",
      "TestArithmeticIsNeverFused"
    ],
    "evidence": "ADAMIC_MUTANT=M04 ADAMIC_BUILD_CACHE_DIR=/tmp/u045/cache/M04 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestNoReaderCallingConvention|TestBorrowChainDeclarations|TestBorrowChainTargets|TestPassThroughsAreNotConsumers|TestCaseMappingMatchesNode|TestCaseTablesMatchNodesUnicode|TestInheritanceMemoryPlans|TestClosureConventionDropCount|TestClosureConventionRuntimeDropCount|TestClosureConventionRuntimeFeaturesIgnoreLiterals|TestClosureConventionWrongOrder|TestParserHasNoUnusedOptionalMethodThunks|TestOptionalMethodThunksMatchNode|TestArithmeticIsNeverFused)$' > M04.log 2>&1; contract_test.go:86: sanitize false: got \"5.5511151231257827e-17 -5.5511151231257827e-17 0\\n\", want \"0 0 0\\n\", as JavaScript computes it"
  }
]
```

All production matrices ran the 14 names in names.json. Beyond that bounded set, catches are unknown. Static caller evidence is in static-test-callers.log, reached Go functions in reached-functions.txt (187), and the C entry/call graph in runtime-functions.txt.

Code under test: native C emission and calling conventions, borrow/region/reuse analysis, runtime feature selection, native build flags, the C case runtime and its production version metadata. Oracles: actual Node mappings/version/output, handwritten plan/C marker/numeric expectations, and clang rejection plus handwritten diagnostic labels for witnesses. No oracle, harness or original test was mutated for production results. The two tagged witness diffs are the explicitly allowed weakened-check exception.

Witnesses: W_BUILD ignores clang failures in the Build adapter. DropCount, RuntimeDropCount and WrongOrder all fail. W_NODE makes the planted missing-thunk comparison always report agreement; OptionalMethodThunksMatchNode fails. Production-mutant failures of these four witnesses are recorded as precondition failures, never production kills.

| ID | Base file:line | Change | Eligible failed rows | Witness precondition failures, excluded |
|---|---|---|---|---|
| M01 | internal/native/emit.go:52 | program.ClosureConventionNeeded() -> !program.ClosureConventionNeeded() | R1 | R8, R11, R13 |
| M02 | internal/native/emit.go:40 | if regexCallbacks { -> if !regexCallbacks { | R10 | R9 |
| M03 | internal/native/library.go:49 | strings.Contains(source, "#define "+feature+" 1\n") -> strings.Contains(source, feature) |  |  |
| M04 | internal/native/native.go:86 | "-ffp-contract=off" -> "-ffp-contract=fast" | R14 |  |
| M05 | internal/native/borrow.go:68 | Pass-through exclusions return true instead of false | R4 |  |
| M06 | internal/native/borrow.go:101 | default: 		return false 	} 	return true -> default: 		return true 	} 	return true |  |  |
| M07 | internal/native/borrow.go:168 | expression.Method \|\| expression.Optional \|\| expression.Object.Type() == ir.Weak -> expression.Method \|\| !expression.Optional \|\| expression.Object.Type() == ir.Weak | R2 |  |
| M08 | internal/native/borrow.go:200 | ok && names[write.Name] -> ok && !names[write.Name] | R2, R3 |  |
| M09 | internal/native/borrow.go:214 | if targets.Unknown { 						safe = false -> if false && targets.Unknown { 						safe = false | R3 |  |
| M10 | internal/native/element_borrow.go:40 | if element \|\| chain { -> if element \|\| false && chain { | R2 |  |
| M11 | internal/native/element_borrow.go:267 | local.Global \|\| local.Captured \|\| local.Function != function \|\| !lendable(local.Type) \|\| assigned[declare.Local] -> local.Global \|\| !local.Captured \|\| local.Function != function \|\| !lendable(local.Type) \|\| assigned[declare.Local] | R2 |  |
| M12 | internal/native/region.go:68 | !function.Closure && function.RestElement == 0 && function.Returns == ir.Object -> !function.Closure && function.RestElement == 0 && function.Returns != ir.Object | R7 |  |
| M13 | internal/native/region.go:372 | if !known \|\| escapes { -> if !known \|\| !escapes { |  |  |
| M14 | internal/native/reuse.go:126 | consumed = consumed \|\| plan.consumed[program.Functions[target].Parameters[position]] -> consumed = consumed && plan.consumed[program.Functions[target].Parameters[position]] | R7 |  |
| M15 | internal/native/emit_objects.go:344 | valueType == ir.Union \|\| valueType == ir.MaybeBoolean && !needed -> valueType == ir.Union \|\| valueType == ir.MaybeBoolean && needed | R12 | R13 |
| M16 | internal/native/emit_objects.go:392 | "(argument_count > %d ? %s : %s)" -> "(argument_count >= %d ? %s : %s)" |  |  |
| M17 | internal/native/emit_functions.go:21 | count = ", size_t argument_count" -> (drop statement) |  | R8, R11 |
| M18 | internal/native/runtime/case.c:198 | character - 'a' + 'A' -> character - 'a' + 'B' | R5 |  |
| M19 | internal/native/runtime/case.c:172 | encode(0x3c2, out) -> encode(0x3c3, out) | R5 |  |
| M20 | internal/native/runtime/case_tables.h:8 | #define CASE_UNICODE_VERSION "17.0.0" -> #define CASE_UNICODE_VERSION "16.0.0" | R6 |  |

Survivors in the eligible bounded production matrix:

- M03: observed featureFlags("// ADAMIC_NODE_HOST\n") changes [] to [-DADAMIC_NODE_HOST=1]. This comment selects a runtime feature.
- M06: observed consumes(ir.Call{}) changes false to true. The table-based test does not cover this unknown-operation default.
- M13: controlled object-parameter summaries change safe/storing/mixed from false/true/true to true/false/true. Mixed-target truth remains the same, exposing the weakness of the mixed-summary assertion.
- M16: emitted method-thunk text changes its outer argument_count > 0 to >= 0; an inner > 0 guard remains. No native execution difference was established. Treat it as an execution-equivalent candidate, not a demonstrated semantic bug.
- M17: counted closure signature loses its size_t argument_count parameter. No eligible regular row catches it, but DropCount and WrongOrder fail their valid-build precondition. This is not an all-tests survivor; those catches are excluded by the witness rule.

Empty-answer observations:

- P_C: NoReader fails its counted cases; closure_convention_plain.a passes. ParserHasNoUnusedOptionalMethodThunks and RuntimeFeaturesIgnoreLiterals both pass empty C.
- P_BORROW: BorrowChainDeclarations fails its positive walk expectations; InheritanceMemoryPlans passes an empty borrow plan.
- P_CHAIN and P_CONSUMES: their rows fail their positive expectations.
- P_REGIONS and P_REUSE: InheritanceMemoryPlans fails an empty region or reuse plan.
- P_LIBRARY: RuntimeFeaturesIgnoreLiterals fails when no library path is returned.
- P_BUILD: ArithmeticIsNeverFused fails when Build returns success without producing a binary.
- P_CASE_LOWER and P_CASE_UPPER: CaseMappingMatchesNode fails when either runtime entry returns NULL.
- Unicode version metadata has no callable production entry to short-circuit; it was not probed, and its vacuous value is null. Witness vacuity is also null.
- For a composite row, vacuous=true denotes a passing empty probe on at least one direct entry. vacuous_entries names that entry; it does not erase the row's checks on other entries.

Brief ambiguities, limits and costs:

- Fresh origin/main is e8cb7ac7, not the historical 8de93800f4. Fresh main governed; no requested name moved or vanished.
- Three mutants per 14 rows would be 42; the cap of 20 governed. Four rows are witnesses and were judged by the two weakened checks instead.
- The whole package cannot fit the 90-second budget. The bounded matrix includes all requested rows, and caller searches show other package tests reach the mutated functions. Sacred and unique_kills are bounded findings only; central replay must settle package and repo uniqueness.
- Wrapper grouping was checked against actual bodies. The closure witnesses assert different properties and remain separate. The points/contexts sweeps are already subtests of one top-level row.
- M10 initially left a chain binding unused. Before any matrix ran, its condition was changed to element || false && chain, preserving the deliberate disabled-chain behavior and a compilable binding. M17's standalone diff was later cleaned of blank-line trailing whitespace without changing behavior.
- Runtime C can read the selector; production header metadata cannot. M20 required its own embedded-source rebuild. Its command wall and binary durations are logged rather than conflated with a timed standalone rebuild.
- The Flags mutation changes compiler options, which rebuilds runtime variants. Native Build writes fresh fixture executables; runtime libraries are keyed by source/header bytes, compiler and flags. Every matrix also received ADAMIC_BUILD_CACHE_DIR=/tmp/u045/cache/ID.
- Three vacuity findings concern one entry of a composite check. The entry names are reported explicitly. No safe callable empty-entry probe was added for embedded Unicode metadata.
- The optional-method row combines positive Node comparisons and a planted failure. Under the brief's witness rule, only W_NODE decides its verdict; production precondition failures do not count.
- No production matrix cooked or panicked. The whole-package baseline timed out, rather than producing a red assertion baseline; the narrowed clean baseline passed.
- 86 default baseline skips were observed, listed individually in skipped-baseline.json. They include the WASI families without a configured SDK and opt-in measurement/archive rows. These outside-scope integrations were not provisioned or enabled after narrowing. Installable benchmark opt-ins were left uncovered because the 610-measurement and other inventory runs exceed this unit's budget. No scoped function skipped.
- M16 has a generated-C textual difference but no demonstrated runtime difference; it is not labeled an unguarded semantic failure. M17 has observed witness precondition catches; it is not labeled globally unguarded.
- No other package tests ran. Independent replay diffs all apply to the starting sources and compile. Production sources and original tests are restored; no PR or main push was made.

Timing: warm toolchain worked, setup skipped (0 seconds); nproc 5. npm ci reported 0.394s. Whole baseline: 90.023s cooked. Scoped baseline: 18.635s binary, 25.448s command wall including coverage build. All three isolated measurements: 146.763s command wall. Switched Go test build: 6.669s. Matrices, weakened checks and probes: 412.116s command wall, 336.846s binary elapsed. Independent final diff validation: 5.449s. M20 command/binary/rebuild-overhead observation: [{'wall': 30.084431933002634, 'binary_seconds': 23.355, 'overhead_including_embedded_rebuild': 6.729431933002633}]. Total session approximately 25 minutes, including source review, preparation and survivor observations.
