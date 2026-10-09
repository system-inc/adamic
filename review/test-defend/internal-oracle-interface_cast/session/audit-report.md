Starting main: 859dee825a4ef89f5c4ecc0f00e6620ac75c4994; all 15 names present, none moved or vanished.
15 rows: seven production/mixed rows and eight pure witnesses; no extra family grouping.
Clean scoped baseline passed in 8.625s; whole package exceeded 90s.
Bounded verdicts: four sacred, three subsumed; eight witness verdicts.
Evidence branch: test-audit/internal-oracle-interface_cast; nproc 5.

```json
[
  {
    "test": "TestInterfaceCastOracle",
    "package": "internal/oracle",
    "file": "internal/oracle/interface_cast_test.go",
    "seconds": 0.609,
    "oracle": "Node source observations plus self-written Adamic inserted-panic contracts, stdout and stderr",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M01",
      "M02"
    ],
    "unique_kills": [],
    "last_proven_fail": "M02: interface_cast_test.go:59: native sanitized: exit codes differ; exit 70 stdout \"\" stderr \"adamic: panic: cast failed: this Node is not a Identifier\\n\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestInterfaceCastImportedConstruction"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P01",
      "P03",
      "P04"
    ],
    "subsumer_seconds": 0.398,
    "vacuous": false,
    "bounded": true,
    "evidence": "ADAMIC_GATE_UNCACHED=1 ADAMIC_MUTANT=2 ADAMIC_BUILD_CACHE_DIR=/tmp/u061/cache/M02 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestInterfaceCastOracle|TestInterfaceCastChecksMalformedRead|TestInterfaceCastImportedConstruction|TestInterfaceCastScalarTags|TestJSONStringifyRefusals|TestJSONStringifyResultMayBeUndefined|TestLibraryMapSetIteratorCopiesRefused)$ => interface_cast_test.go:59: native sanitized: exit codes differ; exit 70 stdout \"\" stderr \"adamic: panic: cast failed: this Node is not a Identifier\\n\"",
    "matrix_rows": [
      "TestInterfaceCastOracle",
      "TestInterfaceCastChecksMalformedRead",
      "TestInterfaceCastImportedConstruction",
      "TestInterfaceCastScalarTags",
      "TestJSONStringifyRefusals",
      "TestJSONStringifyResultMayBeUndefined",
      "TestLibraryMapSetIteratorCopiesRefused"
    ]
  },
  {
    "test": "TestInterfaceCastChecksMalformedRead",
    "package": "internal/oracle",
    "file": "internal/oracle/interface_cast_test.go",
    "seconds": 0.489,
    "oracle": "Node source observations plus self-written Adamic inserted-panic contracts, stdout and stderr",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M01",
      "M02",
      "M04"
    ],
    "unique_kills": [
      "M04"
    ],
    "last_proven_fail": "M04: interface_cast_test.go:90: malformed view read: stderr differs; got oracle.run{stdout:[]uint8{}, stderr:[]uint8{0x61, 0x64, 0x61, 0x6d, 0x69, 0x63, 0x3a, 0x20, 0x70, 0x61, 0x6e, 0x69, 0x63, 0x3a, 0x20, 0x63, 0x6f, 0x6d, 0x70, 0x69, 0x6c, 0x65, 0x72, 0x20, 0x62, 0x75, 0x67, 0x3a, 0x20, 0x61, 0x20, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x20, 0x74, 0x68, 0x65, 0x20, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x65, 0x72, 0x20, 0x70, 0x72, 0x6f, 0x76, 0x65, 0x64, 0x20, 0x69, 0x73, 0x20, 0x74, 0x68, 0x65, 0x72, 0x65, 0x20, 0x69, 0x73, 0x20, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6e, 0x67, 0xa}, exitCode:70}",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P01",
      "P03",
      "P04"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "evidence": "ADAMIC_GATE_UNCACHED=1 ADAMIC_MUTANT=4 ADAMIC_BUILD_CACHE_DIR=/tmp/u061/cache/M04 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestInterfaceCastOracle|TestInterfaceCastChecksMalformedRead|TestInterfaceCastImportedConstruction|TestInterfaceCastScalarTags|TestJSONStringifyRefusals|TestJSONStringifyResultMayBeUndefined|TestLibraryMapSetIteratorCopiesRefused)$ => interface_cast_test.go:90: malformed view read: stderr differs; got oracle.run{stdout:[]uint8{}, stderr:[]uint8{0x61, 0x64, 0x61, 0x6d, 0x69, 0x63, 0x3a, 0x20, 0x70, 0x61, 0x6e, 0x69, 0x63, 0x3a, 0x20, 0x63, 0x6f, 0x6d, 0x70, 0x69, 0x6c, 0x65, 0x72, 0x20, 0x62, 0x75, 0x67, 0x3a, 0x20, 0x61, 0x20, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x20, 0x74, 0x68, 0x65, 0x20, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x65, 0x72, 0x20, 0x70, 0x72, 0x6f, 0x76, 0x65, 0x64, 0x20, 0x69, 0x73, 0x20, 0x74, 0x68, 0x65, 0x72, 0x65, 0x20, 0x69, 0x73, 0x20, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6e, 0x67, 0xa}, exitCode:70}",
    "matrix_rows": [
      "TestInterfaceCastOracle",
      "TestInterfaceCastChecksMalformedRead",
      "TestInterfaceCastImportedConstruction",
      "TestInterfaceCastScalarTags",
      "TestJSONStringifyRefusals",
      "TestJSONStringifyResultMayBeUndefined",
      "TestLibraryMapSetIteratorCopiesRefused"
    ]
  },
  {
    "test": "TestInterfaceCastRuntimeMutants",
    "package": "internal/oracle",
    "file": "internal/oracle/interface_cast_test.go",
    "seconds": 0.54,
    "oracle": "Node source observations and self-written expectation that disagreement detects the built-in mutant; W01 disabled the witnessed check",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W01: interface_cast_test.go:141: mutant survived",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [
      "P05"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^TestInterfaceCastRuntimeMutants$ => interface_cast_test.go:141: mutant survived",
    "witness_checks": [
      "W01",
      "W02"
    ],
    "probe_note": "P05 is the empty-answer W01 weakening, sharing the same diff and observations"
  },
  {
    "test": "TestInterfaceCastImportedConstruction",
    "package": "internal/oracle",
    "file": "internal/oracle/interface_cast_test.go",
    "seconds": 0.398,
    "oracle": "Node source stdout, stderr and exit status compared to native and JavaScript backends",
    "oracle_kind": "external-run",
    "kills": [
      "M01",
      "M02"
    ],
    "unique_kills": [],
    "last_proven_fail": "M02: interface_cast_test.go:175: exit codes differ",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestInterfaceCastChecksMalformedRead"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P01",
      "P03",
      "P04"
    ],
    "subsumer_seconds": 0.489,
    "vacuous": false,
    "bounded": true,
    "evidence": "ADAMIC_GATE_UNCACHED=1 ADAMIC_MUTANT=2 ADAMIC_BUILD_CACHE_DIR=/tmp/u061/cache/M02 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestInterfaceCastOracle|TestInterfaceCastChecksMalformedRead|TestInterfaceCastImportedConstruction|TestInterfaceCastScalarTags|TestJSONStringifyRefusals|TestJSONStringifyResultMayBeUndefined|TestLibraryMapSetIteratorCopiesRefused)$ => interface_cast_test.go:175: exit codes differ",
    "matrix_rows": [
      "TestInterfaceCastOracle",
      "TestInterfaceCastChecksMalformedRead",
      "TestInterfaceCastImportedConstruction",
      "TestInterfaceCastScalarTags",
      "TestJSONStringifyRefusals",
      "TestJSONStringifyResultMayBeUndefined",
      "TestLibraryMapSetIteratorCopiesRefused"
    ]
  },
  {
    "test": "TestInterfaceCastScalarTags",
    "package": "internal/oracle",
    "file": "internal/oracle/interface_cast_test.go",
    "seconds": 0.679,
    "oracle": "Node source observations plus self-written Adamic inserted-panic contracts, stdout and stderr; scalar row also witnesses omission detection",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M01",
      "M02"
    ],
    "unique_kills": [],
    "last_proven_fail": "M02: interface_cast_test.go:218: native: exit codes differ; exit 70 stdout \"casting\\n\" stderr \"adamic: panic: cast failed: this Base is not a Member\\n\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestInterfaceCastImportedConstruction"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P01",
      "P03",
      "P04"
    ],
    "subsumer_seconds": 0.398,
    "vacuous": false,
    "bounded": true,
    "evidence": "ADAMIC_GATE_UNCACHED=1 ADAMIC_MUTANT=2 ADAMIC_BUILD_CACHE_DIR=/tmp/u061/cache/M02 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestInterfaceCastOracle|TestInterfaceCastChecksMalformedRead|TestInterfaceCastImportedConstruction|TestInterfaceCastScalarTags|TestJSONStringifyRefusals|TestJSONStringifyResultMayBeUndefined|TestLibraryMapSetIteratorCopiesRefused)$ => interface_cast_test.go:218: native: exit codes differ; exit 70 stdout \"casting\\n\" stderr \"adamic: panic: cast failed: this Base is not a Member\\n\"",
    "matrix_rows": [
      "TestInterfaceCastOracle",
      "TestInterfaceCastChecksMalformedRead",
      "TestInterfaceCastImportedConstruction",
      "TestInterfaceCastScalarTags",
      "TestJSONStringifyRefusals",
      "TestJSONStringifyResultMayBeUndefined",
      "TestLibraryMapSetIteratorCopiesRefused"
    ],
    "witness_component": "W01 also fails the built-in scalar omission checks; production verdict uses only M-series kills"
  },
  {
    "test": "TestArrayFamilyMutants",
    "package": "internal/oracle",
    "file": "internal/oracle/library_array_mutant_test.go",
    "seconds": 1.534,
    "oracle": "Node source observations and self-written expectation that disagreement detects the built-in mutant; W01 disabled the witnessed check",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W01: library_array_mutant_test.go:122: mutant caught by \"\", want stdout differs",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [
      "P05"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^TestArrayFamilyMutants$ => library_array_mutant_test.go:122: mutant caught by \"\", want stdout differs",
    "witness_checks": [
      "W01",
      "W02"
    ],
    "probe_note": "P05 is the empty-answer W01 weakening, sharing the same diff and observations"
  },
  {
    "test": "TestArrayWithBoundsMutant",
    "package": "internal/oracle",
    "file": "internal/oracle/library_array_mutant_test.go",
    "seconds": 0.342,
    "oracle": "Node source observations and self-written expectation that disagreement detects the built-in mutant; W01 disabled the witnessed check",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W01: library_array_mutant_test.go:170: bounds mutant caught by \"\"",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [
      "P05"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^TestArrayWithBoundsMutant$ => library_array_mutant_test.go:170: bounds mutant caught by \"\"",
    "witness_checks": [
      "W01",
      "W02"
    ],
    "probe_note": "P05 is the empty-answer W01 weakening, sharing the same diff and observations"
  },
  {
    "test": "TestJSONStringifyRefusals",
    "package": "internal/oracle",
    "file": "internal/oracle/library_json_stringify_test.go",
    "seconds": 0.204,
    "oracle": "Self-written Adamic refusal text; duplicate TS1117 copied from TypeScript diagnostic authority, checked against cohere/TypeScript/tsc/internal/diagnostics/diagnosticMessages.json this session",
    "oracle_kind": [
      "self",
      "external-authority"
    ],
    "kills": [
      "M05",
      "M06",
      "M07",
      "M12"
    ],
    "unique_kills": [
      "M05",
      "M06",
      "M07",
      "M12"
    ],
    "last_proven_fail": "M12: library_json_stringify_test.go:44: want refusal containing \"callback must have a proven type\", got /tmp/adamic-gate/TestJSONStringifyRefusalsreplacer_function3460705154/001/probe.ts:1:19: stage 0 can't lower JSON.stringify replacer functions (the callback must have a runtime type for every visited value and its holder) yet",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P01",
      "P02"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "evidence": "ADAMIC_GATE_UNCACHED=1 ADAMIC_MUTANT=12 ADAMIC_BUILD_CACHE_DIR=/tmp/u061/cache/M12 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestInterfaceCastOracle|TestInterfaceCastChecksMalformedRead|TestInterfaceCastImportedConstruction|TestInterfaceCastScalarTags|TestJSONStringifyRefusals|TestJSONStringifyResultMayBeUndefined|TestLibraryMapSetIteratorCopiesRefused)$ => library_json_stringify_test.go:44: want refusal containing \"callback must have a proven type\", got /tmp/adamic-gate/TestJSONStringifyRefusalsreplacer_function3460705154/001/probe.ts:1:19: stage 0 can't lower JSON.stringify replacer functions (the callback must have a runtime type for every visited value and its holder) yet",
    "matrix_rows": [
      "TestInterfaceCastOracle",
      "TestInterfaceCastChecksMalformedRead",
      "TestInterfaceCastImportedConstruction",
      "TestInterfaceCastScalarTags",
      "TestJSONStringifyRefusals",
      "TestJSONStringifyResultMayBeUndefined",
      "TestLibraryMapSetIteratorCopiesRefused"
    ],
    "entry_note": "Duplicate subcase calls Load; P02-duplicate.log probes that entry separately. Other cases call Lower."
  },
  {
    "test": "TestJSONStringifyResultMayBeUndefined",
    "package": "internal/oracle",
    "file": "internal/oracle/library_json_stringify_test.go",
    "seconds": 0.043,
    "oracle": "Self-written expected undefined rejection from Adamic public declarations; diagnostic substring must contain undefined",
    "oracle_kind": "self",
    "kills": [
      "M10"
    ],
    "unique_kills": [
      "M10"
    ],
    "last_proven_fail": "M10: library_json_stringify_test.go:62: want the checker to require handling stringify's undefined result, got <nil>",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P02"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "evidence": "ADAMIC_GATE_UNCACHED=1 ADAMIC_MUTANT=10 ADAMIC_BUILD_CACHE_DIR=/tmp/u061/cache/M10 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestInterfaceCastOracle|TestInterfaceCastChecksMalformedRead|TestInterfaceCastImportedConstruction|TestInterfaceCastScalarTags|TestJSONStringifyRefusals|TestJSONStringifyResultMayBeUndefined|TestLibraryMapSetIteratorCopiesRefused)$ => library_json_stringify_test.go:62: want the checker to require handling stringify's undefined result, got <nil>",
    "matrix_rows": [
      "TestInterfaceCastOracle",
      "TestInterfaceCastChecksMalformedRead",
      "TestInterfaceCastImportedConstruction",
      "TestInterfaceCastScalarTags",
      "TestJSONStringifyRefusals",
      "TestJSONStringifyResultMayBeUndefined",
      "TestLibraryMapSetIteratorCopiesRefused"
    ]
  },
  {
    "test": "TestJSONStringifyOracleCatchesKeyOrder",
    "package": "internal/oracle",
    "file": "internal/oracle/library_json_stringify_test.go",
    "seconds": 0.319,
    "oracle": "Node source observations and self-written expectation that disagreement detects the built-in mutant; W01 disabled the witnessed check",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W01: library_json_stringify_test.go:89: want Node alone to catch key order as stdout differs, got \"\"",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [
      "P05"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^TestJSONStringifyOracleCatchesKeyOrder$ => library_json_stringify_test.go:89: want Node alone to catch key order as stdout differs, got \"\"",
    "witness_checks": [
      "W01",
      "W02"
    ],
    "probe_note": "P05 is the empty-answer W01 weakening, sharing the same diff and observations"
  },
  {
    "test": "TestLibraryMapSetIteratorCopiesRefused",
    "package": "internal/oracle",
    "file": "internal/oracle/library_map_set_iterator_test.go",
    "seconds": 0.792,
    "oracle": "Node copy.next failure plus self-written NotYet inherited-next refusal text",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M08",
      "M09"
    ],
    "unique_kills": [
      "M08",
      "M09"
    ],
    "last_proven_fail": "M09: library_map_set_iterator_test.go:42: want inherited-next refusal, got <nil>",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "evidence": "ADAMIC_GATE_UNCACHED=1 ADAMIC_MUTANT=9 ADAMIC_BUILD_CACHE_DIR=/tmp/u061/cache/M09 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestInterfaceCastOracle|TestInterfaceCastChecksMalformedRead|TestInterfaceCastImportedConstruction|TestInterfaceCastScalarTags|TestJSONStringifyRefusals|TestJSONStringifyResultMayBeUndefined|TestLibraryMapSetIteratorCopiesRefused)$ => library_map_set_iterator_test.go:42: want inherited-next refusal, got <nil>",
    "matrix_rows": [
      "TestInterfaceCastOracle",
      "TestInterfaceCastChecksMalformedRead",
      "TestInterfaceCastImportedConstruction",
      "TestInterfaceCastScalarTags",
      "TestJSONStringifyRefusals",
      "TestJSONStringifyResultMayBeUndefined",
      "TestLibraryMapSetIteratorCopiesRefused"
    ]
  },
  {
    "test": "TestLibraryMapSetMutants",
    "package": "internal/oracle",
    "file": "internal/oracle/library_map_set_test.go",
    "seconds": 2.449,
    "oracle": "Node source observations and self-written expectation that disagreement detects the built-in mutant; W01 disabled the witnessed check",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W01: library_map_set_test.go:111: mutant was not caught by stdout comparison: \"\"",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [
      "P05"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^TestLibraryMapSetMutants$ => library_map_set_test.go:111: mutant was not caught by stdout comparison: \"\"",
    "witness_checks": [
      "W01",
      "W02"
    ],
    "probe_note": "P05 is the empty-answer W01 weakening, sharing the same diff and observations"
  },
  {
    "test": "TestMathNumberOracleCatchesMutants",
    "package": "internal/oracle",
    "file": "internal/oracle/library_math_number_test.go",
    "seconds": 0.881,
    "oracle": "Node source observations and self-written expectation that disagreement detects the built-in mutant; W01 disabled the witnessed check",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W01: library_math_number_test.go:49: got \"\", want stdout differs",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [
      "P05"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^TestMathNumberOracleCatchesMutants$ => library_math_number_test.go:49: got \"\", want stdout differs",
    "witness_checks": [
      "W01",
      "W02"
    ],
    "probe_note": "P05 is the empty-answer W01 weakening, sharing the same diff and observations"
  },
  {
    "test": "TestLibraryStringMutants",
    "package": "internal/oracle",
    "file": "internal/oracle/library_string_test.go",
    "seconds": 1.5550000000000002,
    "oracle": "Node source observations and self-written expectation that disagreement detects the built-in mutant; W01 disabled the witnessed check",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W01: library_string_test.go:121: want only the Node comparison to kill mutant, got \"\"",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [
      "P05"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^TestLibraryStringMutants$ => library_string_test.go:121: want only the Node comparison to kill mutant, got \"\"",
    "witness_checks": [
      "W01",
      "W02"
    ],
    "probe_note": "P05 is the empty-answer W01 weakening, sharing the same diff and observations"
  },
  {
    "test": "TestLiteralOptionalOracleCatchesMutant",
    "package": "internal/oracle",
    "file": "internal/oracle/literal_optional_test.go",
    "seconds": 0.337,
    "oracle": "Node source observations and self-written expectation that disagreement detects the built-in mutant; W01 disabled the witnessed check",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W01: literal_optional_test.go:48: Node did not catch mutant",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [
      "P05"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^TestLiteralOptionalOracleCatchesMutant$ => literal_optional_test.go:48: Node did not catch mutant",
    "witness_checks": [
      "W01",
      "W02"
    ],
    "probe_note": "P05 is the empty-answer W01 weakening, sharing the same diff and observations"
  }
]
```

| ID | Starting-main file:line | Change | Failed rows |
|---|---|---|---|
| M01 | internal/native/view_fields.go:31 | `e.line("adamic_value %s = adamic_object_view(%s, %s, &%s, %d, %s, %s);", slot, object, cString(property.Name), e.cache(), property.Of, cString(name), cString(property.View)) -> e.line("adamic_value %s = adamic_object_view(%s, %s, &%s, %d, %s, %s);", slot, object, cString(property.View), e.cache(), property.Of, cString(name), cString(property.Name))` | TestInterfaceCastChecksMalformedRead, TestInterfaceCastImportedConstruction, TestInterfaceCastOracle, TestInterfaceCastScalarTags |
| M02 | internal/native/emit_expressions.go:435 | `e.binary(ir.Equal, expression.FieldType, field, e.value(allowed)) -> e.binary(ir.NotEqual, expression.FieldType, field, e.value(allowed))` | TestInterfaceCastChecksMalformedRead, TestInterfaceCastImportedConstruction, TestInterfaceCastOracle, TestInterfaceCastScalarTags |
| M03 | internal/native/native.go:35 | `character >= 0x20 -> character >= 0x21` | survivor |
| M04 | internal/lower/readiness.go:287 | `!program.CheckedFields[expression.Name] || expression.Method -> program.CheckedFields[expression.Name] || expression.Method` | TestInterfaceCastChecksMalformedRead |
| M05 | internal/lower/library_json_stringify.go:14 | `JSON.parse: its result's type can't be proven from the text -> JSON.parse is unsupported` | TestJSONStringifyRefusals |
| M06 | internal/lower/library_json_stringify.go:185 | `of == ir.Object || of == ir.Weak -> of != ir.Object && of != ir.Weak` | TestJSONStringifyRefusals |
| M07 | internal/lower/library_json_stringify.go:84 | `name == "toJSON" || name == "__proto__" -> name != "toJSON" && name != "__proto__"` | TestJSONStringifyRefusals |
| M08 | internal/lower/library_map_set.go:514 | `if isIterator(source) { -> if !isIterator(source) {` | TestLibraryMapSetIteratorCopiesRefused |
| M09 | internal/lower/library_map_set.go:501 | `property.Kind == ast.KindSpreadAssignment && isIterator(property.AsSpreadAssignment().Expression) -> property.Kind == ast.KindSpreadAssignment && !isIterator(property.AsSpreadAssignment().Expression)` | TestLibraryMapSetIteratorCopiesRefused |
| M10 | internal/load/prelude.d.ts:82 | `stringify(value?: unknown, replacer?: unknown, space?: unknown): string | undefined; -> stringify(value?: unknown, replacer?: unknown, space?: unknown): string;` | TestJSONStringifyResultMayBeUndefined |
| M11 | internal/load/load.go:61 | `ExactOptionalPropertyTypes: core.TSTrue, -> ExactOptionalPropertyTypes: core.TSFalse,` | survivor |
| M12 | internal/lower/library_json_stringify.go:33 | `JSON.stringify replacer functions (the callback must have a proven type for every visited value and its holder) -> JSON.stringify replacer functions (the callback must have a runtime type for every visited value and its holder)` | TestJSONStringifyRefusals |

Survivors:

M03: equivalent candidate. native.C changes ADAMIC_STRING("a b") to ADAMIC_STRING("a\040b"); rebuilt programs both print "a b\n".

M11: unguarded by this bounded set. Load changes from TS2375 rejection to acceptance of `{p: undefined}` when ExactOptionalPropertyTypes is false. Outside-set protection remains unknown.

The named files came from an older commit. Fresh origin/main was 859dee82. Discovery verified all 15 names and their bodies; no name moved or vanished. The code inventory lists 693 covered functions in load, lower, native and javascript. Runtime C calls were exercised by existing sanitized builds but not exhaustively instrumented. No runtime C source was mutated.

Eight rows plant built-in IR or generated-C failures. Production mutations cannot establish their witness verdicts: disabling disagreement does. The scalar-tag row has both conformance assertions and built-in omission checks, so it retains a production verdict with separate witness evidence. The top-level bodies have distinct assertions; there are no additional input-only wrappers to group together.

W01 disables the entire comparison and fails all eight pure witnesses. W02 disables only stdout and fails seven: LiteralOptional still detects the planted exit-code change. P05 is the empty-answer W01 check, with the same physical diff and observations; it is not another independent mutant. Witness kills and uniqueness remain empty. Witness bounded=false refers to a complete isolated check of the row, not a claim about package-wide uniqueness.

The whole package reached 90.097s without an earlier reported failure. The scoped uncached baseline passed in 8.625s. All production matrices are bounded to the seven production/mixed rows. All named rows were measured alone three times with -count=1 and ADAMIC_GATE_UNCACHED=1. No scoped row skipped. Outside-scope package tests and opt-ins were not replayed per mutant, and uniqueness outside the bounded rows remains unknown.

The oracle caches observations, so selector runs used ADAMIC_GATE_UNCACHED=1 as well as per-mutant ADAMIC_BUILD_CACHE_DIR values. Runtime libraries have their own source-content cache. The switch preserves identical source between selector runs while emitting different generated products. Native products were rebuilt by the normal uncached harness; the runtime sources and oracle were untouched by production mutations. Twelve mutants, including four that can change native emission and eight checking/lowering changes, stay below the 20-mutant cap. This does not reach the aspirational three per row.

The driver conservatively reran rows when logs contained panic text. Some of that text was an expected native Adamic panic rather than a Go test-binary abort, so those isolation reruns cost unnecessary time. Their actual results and commands are retained. Go panic probes were isolated, including the duplicate-key Load entry. The JSON refusal row mixes Load for duplicate diagnostics and Lower for its other cases; it therefore received both applicable entry probes.

Refusal and inserted-panic strings are self-written contracts. Duplicate TS1117 is external-authority evidence: its numeric value was checked against TypeScript's diagnosticMessages.json this session. Node supplies independent execution observations, but Adamic's inserted panic policy differs from running the unchecked source. The result-may-be-undefined row checks a diagnostic substring, not a complete diagnostic. The two broad interface mutants provide only a small subsumption hint: every subsumed row caught M01 and M02, and another named row caught both. This is not a deletion recommendation.

M03 changes emitted C spelling while preserving observed native behavior, so it remains an equivalent candidate. M11 changes acceptance of an explicit undefined optional property and survives this bounded set. Central replay may find protection elsewhere. The survivor source, commands and outputs are retained. No other package's tests were run, no PR was opened, and main was not pushed. Production sources were restored and final vet passed.

Measured timings: core setup 0s; npm installation and initial Go/coverage compilation were not separately timed. Row timing runs total 33.517 binary seconds. Production/witness/probe matrix commands total 72.79 wall seconds plus 41.992s isolation reruns. Isolated standalone witness checks add 24.645s. Standalone production vet checks took 4.605s; other replay compilation checks took 1.187s. Runtime rebuilding is included in these command wall times and was not separately instrumented.
