Unit u103, starting origin/main a467d1a1571c43e0f01fc4efcb43890280e4a1ac.
All 141 supplied names exist, grouped into 11 rows under the family rules.
Clean whole package and full agreement family exceed 90s; bounded landing sample passes.
Verdicts: 1 slow-worthy, 1 sacred, 2 subsumed, 4 setup-check, 3 witness.
M01 survives the sampled matrix with a changed-output witness; setup and product rows are vacuous.

```json
[
  {
    "test": "TestPortMatchesGoCohereSplit_Setup",
    "package": "stage1/cohere/json",
    "file": "stage1/cohere/json/port_matches_prepared_split_test.go",
    "seconds": 2.759,
    "oracle": "Self: preparation returns successfully and publishes shared-ready state; no product is executed.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "H01: port_matches_prepared_split_test.go:106: shared preparation failed",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestPortMatchesGoCohereSplit_Setup"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u103/cache/H01 ADAMIC_GATE_SAMPLE=a467d1a1571c43e0f01fc4efcb43890280e4a1ac timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/json/ -run '^(TestPortMatchesGoCohereSplit_Setup)$' > H01.log; port_matches_prepared_split_test.go:106: shared preparation failed",
    "members": [
      "TestPortMatchesGoCohereSplit_Setup"
    ],
    "observed_assertion": "port_matches_prepared_split_test.go:106: shared preparation failed",
    "construction_or_witness_kills": [
      "H01"
    ]
  },
  {
    "test": "TestPortMatchesGoCohereSplitUnion",
    "package": "stage1/cohere/json",
    "file": "stage1/cohere/json/port_matches_prepared_split_test.go",
    "seconds": 0.448,
    "oracle": "Self: exactly one partition must reject a synthetic changed answer; W01 weakens comparisonError.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W01: port_matches_prepared_split_test.go:273: planted failure caught by 0 shards",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestPortMatchesGoCohereSplitUnion"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u103/cache/W01 ADAMIC_GATE_SAMPLE=a467d1a1571c43e0f01fc4efcb43890280e4a1ac timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/json/ -run '^(TestPortMatchesGoCohereSplitUnion)$' > W01.log; port_matches_prepared_split_test.go:273: planted failure caught by 0 shards",
    "members": [
      "TestPortMatchesGoCohereSplitUnion"
    ],
    "observed_assertion": "port_matches_prepared_split_test.go:273: planted failure caught by 0 shards",
    "construction_or_witness_kills": [
      "W01"
    ]
  },
  {
    "test": "TestPortMatchesGoCohere family",
    "package": "stage1/cohere/json",
    "file": "stage1/cohere/json/port_matches_prepared_split_test.go",
    "seconds": null,
    "oracle": "Executed Go cohere batch stdout compared byte-for-byte with source Node, native release, ASan/UBSan, leak pass and emitted JavaScript. Bounded 148-case landing sample.",
    "oracle_kind": "external-run",
    "kills": [
      "M02",
      "M03",
      "M04"
    ],
    "unique_kills": [
      "M04"
    ],
    "last_proven_fail": "M04: port_matches_prepared_split_test.go:534: Node case 1 cohere/internal/lint/rules/tailwind/collapse/testdata/candidate_fixtures.json byte 317: got \"\\\\n    \\\"fade-in\\\",\\\\n    \\\"slide-in-from-top\\\"\\\\n  ],\\\\n  \\\"repositoryVariantMarkers\\\": [\\\\n    \\\"sm\\\",\\\\n    \\\"md\\\",\\\\n    \\\"lg\\\",\\\\n    \\\"xl\\\",\\\\n    \\\"2xl\\\"\\\\n  ],\\\\n  \\\"classNameOccurrences\\\": 25510,\\\\n  \\\"repositoryClassCount\\\": 1229,\\\\n  \\\"ambiguousCount\\\": 2760,\\\\n  \\\"ambiguousRepositoryCount\\\": 178,\\\\n  \\\"util\"; Go \"\\\\n    \\\"fade-in\\\",\\\\n    \\\"slide-in-from-top\\\"\\\\n  ],\\\\n  \\\"repositoryVariantMarkers\\\": [\\\"sm\\\", \\\"md\\\", \\\"lg\\\", \\\"xl\\\", \\\"2xl\\\"],\\\\n  \\\"classNameOccurrences\\\": 25510,\\\\n  \\\"repositoryClassCount\\\": 1229,\\\\n  \\\"ambiguousCount\\\": 2760,\\\\n  \\\"ambiguousRepositoryCount\\\": 178,\\\\n  \\\"utilityRoots\\\": {\\\\n    \\\"sr-only\\\": [\"",
    "verdict": "slow-worthy",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "PPort"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestAdditionalJSONBoundaries",
      "TestPortMatchesGoCohere family",
      "TestSingleFileStdoutDriver"
    ],
    "evidence": "ADAMIC_MUTANT=M04 ADAMIC_BUILD_CACHE_DIR=/tmp/u103/cache/switch ADAMIC_GATE_SAMPLE=a467d1a1571c43e0f01fc4efcb43890280e4a1ac timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/json/ -run '^(TestPortMatchesGoCohere_[0-9]{3}|TestAdditionalJSONBoundaries|TestSingleFileStdoutDriver)$'; selector file /tmp/u103/selector contains M04 > M04.log; port_matches_prepared_split_test.go:534: Node case 1 cohere/internal/lint/rules/tailwind/collapse/testdata/candidate_fixtures.json byte 317: got \"\\\\n    \\\"fade-in\\\",\\\\n    \\\"slide-in-from-top\\\"\\\\n  ],\\\\n  \\\"repositoryVariantMarkers\\\": [\\\\n    \\\"sm\\\",\\\\n    \\\"md\\\",\\\\n    \\\"lg\\\",\\\\n    \\\"xl\\\",\\\\n    \\\"2xl\\\"\\\\n  ],\\\\n  \\\"classNameOccurrences\\\": 25510,\\\\n  \\\"repositoryClassCount\\\": 1229,\\\\n  \\\"ambiguousCount\\\": 2760,\\\\n  \\\"ambiguousRepositoryCount\\\": 178,\\\\n  \\\"util\"; Go \"\\\\n    \\\"fade-in\\\",\\\\n    \\\"slide-in-from-top\\\"\\\\n  ],\\\\n  \\\"repositoryVariantMarkers\\\": [\\\"sm\\\", \\\"md\\\", \\\"lg\\\", \\\"xl\\\", \\\"2xl\\\"],\\\\n  \\\"classNameOccurrences\\\": 25510,\\\\n  \\\"repositoryClassCount\\\": 1229,\\\\n  \\\"ambiguousCount\\\": 2760,\\\\n  \\\"ambiguousRepositoryCount\\\": 178,\\\\n  \\\"utilityRoots\\\": {\\\\n    \\\"sr-only\\\": [\"",
    "members": [
      "TestPortMatchesGoCohere_000",
      "TestPortMatchesGoCohere_001",
      "TestPortMatchesGoCohere_002",
      "TestPortMatchesGoCohere_003",
      "TestPortMatchesGoCohere_004",
      "TestPortMatchesGoCohere_005",
      "TestPortMatchesGoCohere_006",
      "TestPortMatchesGoCohere_007",
      "TestPortMatchesGoCohere_008",
      "TestPortMatchesGoCohere_009",
      "TestPortMatchesGoCohere_010",
      "TestPortMatchesGoCohere_011",
      "TestPortMatchesGoCohere_012",
      "TestPortMatchesGoCohere_013",
      "TestPortMatchesGoCohere_014",
      "TestPortMatchesGoCohere_015",
      "TestPortMatchesGoCohere_016",
      "TestPortMatchesGoCohere_017",
      "TestPortMatchesGoCohere_018",
      "TestPortMatchesGoCohere_019",
      "TestPortMatchesGoCohere_020",
      "TestPortMatchesGoCohere_021",
      "TestPortMatchesGoCohere_022",
      "TestPortMatchesGoCohere_023",
      "TestPortMatchesGoCohere_024",
      "TestPortMatchesGoCohere_025",
      "TestPortMatchesGoCohere_026",
      "TestPortMatchesGoCohere_027",
      "TestPortMatchesGoCohere_028",
      "TestPortMatchesGoCohere_029",
      "TestPortMatchesGoCohere_030",
      "TestPortMatchesGoCohere_031",
      "TestPortMatchesGoCohere_032",
      "TestPortMatchesGoCohere_033",
      "TestPortMatchesGoCohere_034",
      "TestPortMatchesGoCohere_035",
      "TestPortMatchesGoCohere_036",
      "TestPortMatchesGoCohere_037",
      "TestPortMatchesGoCohere_038",
      "TestPortMatchesGoCohere_039",
      "TestPortMatchesGoCohere_040",
      "TestPortMatchesGoCohere_041",
      "TestPortMatchesGoCohere_042",
      "TestPortMatchesGoCohere_043",
      "TestPortMatchesGoCohere_044",
      "TestPortMatchesGoCohere_045",
      "TestPortMatchesGoCohere_046",
      "TestPortMatchesGoCohere_047",
      "TestPortMatchesGoCohere_048",
      "TestPortMatchesGoCohere_049",
      "TestPortMatchesGoCohere_050",
      "TestPortMatchesGoCohere_051",
      "TestPortMatchesGoCohere_052",
      "TestPortMatchesGoCohere_053",
      "TestPortMatchesGoCohere_054",
      "TestPortMatchesGoCohere_055",
      "TestPortMatchesGoCohere_056",
      "TestPortMatchesGoCohere_057",
      "TestPortMatchesGoCohere_058",
      "TestPortMatchesGoCohere_059",
      "TestPortMatchesGoCohere_060",
      "TestPortMatchesGoCohere_061",
      "TestPortMatchesGoCohere_062",
      "TestPortMatchesGoCohere_063",
      "TestPortMatchesGoCohere_064",
      "TestPortMatchesGoCohere_065",
      "TestPortMatchesGoCohere_066",
      "TestPortMatchesGoCohere_067",
      "TestPortMatchesGoCohere_068",
      "TestPortMatchesGoCohere_069",
      "TestPortMatchesGoCohere_070",
      "TestPortMatchesGoCohere_071",
      "TestPortMatchesGoCohere_072",
      "TestPortMatchesGoCohere_073",
      "TestPortMatchesGoCohere_074",
      "TestPortMatchesGoCohere_075",
      "TestPortMatchesGoCohere_076",
      "TestPortMatchesGoCohere_077",
      "TestPortMatchesGoCohere_078",
      "TestPortMatchesGoCohere_079",
      "TestPortMatchesGoCohere_080",
      "TestPortMatchesGoCohere_081",
      "TestPortMatchesGoCohere_082",
      "TestPortMatchesGoCohere_083",
      "TestPortMatchesGoCohere_084",
      "TestPortMatchesGoCohere_085",
      "TestPortMatchesGoCohere_086",
      "TestPortMatchesGoCohere_087",
      "TestPortMatchesGoCohere_088",
      "TestPortMatchesGoCohere_089",
      "TestPortMatchesGoCohere_090",
      "TestPortMatchesGoCohere_091",
      "TestPortMatchesGoCohere_092",
      "TestPortMatchesGoCohere_093",
      "TestPortMatchesGoCohere_094",
      "TestPortMatchesGoCohere_095",
      "TestPortMatchesGoCohere_096",
      "TestPortMatchesGoCohere_097",
      "TestPortMatchesGoCohere_098",
      "TestPortMatchesGoCohere_099",
      "TestPortMatchesGoCohere_100",
      "TestPortMatchesGoCohere_101",
      "TestPortMatchesGoCohere_102",
      "TestPortMatchesGoCohere_103",
      "TestPortMatchesGoCohere_104",
      "TestPortMatchesGoCohere_105",
      "TestPortMatchesGoCohere_106",
      "TestPortMatchesGoCohere_107",
      "TestPortMatchesGoCohere_108",
      "TestPortMatchesGoCohere_109",
      "TestPortMatchesGoCohere_110",
      "TestPortMatchesGoCohere_111",
      "TestPortMatchesGoCohere_112",
      "TestPortMatchesGoCohere_113",
      "TestPortMatchesGoCohere_114",
      "TestPortMatchesGoCohere_115",
      "TestPortMatchesGoCohere_116",
      "TestPortMatchesGoCohere_117",
      "TestPortMatchesGoCohere_118",
      "TestPortMatchesGoCohere_119",
      "TestPortMatchesGoCohere_120",
      "TestPortMatchesGoCohere_121",
      "TestPortMatchesGoCohere_122",
      "TestPortMatchesGoCohere_123",
      "TestPortMatchesGoCohere_124",
      "TestPortMatchesGoCohere_125",
      "TestPortMatchesGoCohere_126",
      "TestPortMatchesGoCohere_127"
    ],
    "observed_assertion": "port_matches_prepared_split_test.go:535: Node case 1 cohere/internal/lint/rules/tailwind/collapse/testdata/candidate_fixtures.json byte 317: got \"\\\\n    \\\"fade-in\\\",\\\\n    \\\"slide-in-from-top\\\"\\\\n  ],\\\\n  \\\"repositoryVariantMarkers\\\": [\\\\n    \\\"sm\\\",\\\\n    \\\"md\\\",\\\\n    \\\"lg\\\",\\\\n    \\\"xl\\\",\\\\n    \\\"2xl\\\"\\\\n  ],\\\\n  \\\"classNameOccurrences\\\": 25510,\\\\n  \\\"repositoryClassCount\\\": 1229,\\\\n  \\\"ambiguousCount\\\": 2760,\\\\n  \\\"ambiguousRepositoryCount\\\": 178,\\\\n  \\\"util\"; Go \"\\\\n    \\\"fade-in\\\",\\\\n    \\\"slide-in-from-top\\\"\\\\n  ],\\\\n  \\\"repositoryVariantMarkers\\\": [\\\"sm\\\", \\\"md\\\", \\\"lg\\\", \\\"xl\\\", \\\"2xl\\\"],\\\\n  \\\"classNameOccurrences\\\": 25510,\\\\n  \\\"repositoryClassCount\\\": 1229,\\\\n  \\\"ambiguousCount\\\": 2760,\\\\n  \\\"ambiguousRepositoryCount\\\": 178,\\\\n  \\\"utilityRoots\\\": {\\\\n    \\\"sr-only\\\": [\"",
    "sample_seconds": 35.658,
    "full_corpus_over_budget": true,
    "vacuous_subcases": [
      "TestPortMatchesGoCohere_033",
      "TestPortMatchesGoCohere_097",
      "TestPortMatchesGoCohere_060",
      "TestPortMatchesGoCohere_050",
      "TestPortMatchesGoCohere_049",
      "TestPortMatchesGoCohere_045",
      "TestPortMatchesGoCohere_044",
      "TestPortMatchesGoCohere_041",
      "TestPortMatchesGoCohere_040",
      "TestPortMatchesGoCohere_039",
      "TestPortMatchesGoCohere_035",
      "TestPortMatchesGoCohere_093",
      "TestPortMatchesGoCohere_090",
      "TestPortMatchesGoCohere_095",
      "TestPortMatchesGoCohere_076",
      "TestPortMatchesGoCohere_075",
      "TestPortMatchesGoCohere_085",
      "TestPortMatchesGoCohere_072",
      "TestPortMatchesGoCohere_068",
      "TestPortMatchesGoCohere_070",
      "TestPortMatchesGoCohere_114",
      "TestPortMatchesGoCohere_125",
      "TestPortMatchesGoCohere_124",
      "TestPortMatchesGoCohere_123",
      "TestPortMatchesGoCohere_105",
      "TestPortMatchesGoCohere_119",
      "TestPortMatchesGoCohere_102",
      "TestPortMatchesGoCohere_109",
      "TestPortMatchesGoCohere_100",
      "TestPortMatchesGoCohere_108",
      "TestPortMatchesGoCohere_016",
      "TestPortMatchesGoCohere_031",
      "TestPortMatchesGoCohere_013",
      "TestPortMatchesGoCohere_012",
      "TestPortMatchesGoCohere_021",
      "TestPortMatchesGoCohere_010",
      "TestPortMatchesGoCohere_019",
      "TestPortMatchesGoCohere_029",
      "TestPortMatchesGoCohere_018",
      "TestPortMatchesGoCohere_005",
      "TestPortMatchesGoCohere_003",
      "TestPortMatchesGoCohere_004",
      "TestPortMatchesGoCohere_002",
      "TestPortMatchesGoCohere_025"
    ],
    "vacuous_subcases_note": "These members have empty sampled hash buckets; all nonempty members reject PPort."
  },
  {
    "test": "TestProduct_JSON family",
    "package": "stage1/cohere/json",
    "file": "stage1/cohere/json/port_matches_prepared_split_test.go",
    "seconds": 1.245,
    "oracle": "Self: shared builder succeeds for four recipes. Members do not assert a nonempty path or execute the returned product.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "H02: port_matches_prepared_split_test.go:611: open /tmp/u103/cache/H02/1acbc60b4f93699b7d9dc1d20afb8af1163d5814cbf80d27fcddb396ef0fcc0e/main.c: no such file or directory",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_JSON family"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u103/cache/H02 ADAMIC_GATE_SAMPLE=a467d1a1571c43e0f01fc4efcb43890280e4a1ac timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/json/ -run '^(TestProduct_JSON(GoOracle|LoweredPort|NativeRelease|NativeSanitized))$' > H02.log; port_matches_prepared_split_test.go:611: open /tmp/u103/cache/H02/1acbc60b4f93699b7d9dc1d20afb8af1163d5814cbf80d27fcddb396ef0fcc0e/main.c: no such file or directory",
    "members": [
      "TestProduct_JSONGoOracle",
      "TestProduct_JSONLoweredPort",
      "TestProduct_JSONNativeRelease",
      "TestProduct_JSONNativeSanitized"
    ],
    "observed_assertion": "port_matches_prepared_split_test.go:611: open /tmp/u103/cache/H02/1acbc60b4f93699b7d9dc1d20afb8af1163d5814cbf80d27fcddb396ef0fcc0e/main.c: no such file or directory",
    "construction_or_witness_kills": [
      "H02"
    ]
  },
  {
    "test": "TestThreePortMutantsAreCaught",
    "package": "stage1/cohere/json",
    "file": "stage1/cohere/json/port_test.go",
    "seconds": 9.293,
    "oracle": "Executed Go cohere answers; deliberate port mutants must disagree in successful formatter output. W02 forces the inline equality comparison to report agreement.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W02: port_test.go:506: Node failed to catch wrong filename parser",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestThreePortMutantsAreCaught"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u103/cache/W02 ADAMIC_GATE_SAMPLE=a467d1a1571c43e0f01fc4efcb43890280e4a1ac timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/json/ -run '^(TestThreePortMutantsAreCaught)$' > W02.log; port_test.go:506: Node failed to catch wrong filename parser",
    "members": [
      "TestThreePortMutantsAreCaught"
    ],
    "observed_assertion": "port_test.go:506: Node failed to catch wrong filename parser",
    "construction_or_witness_kills": [
      "W02"
    ]
  },
  {
    "test": "TestAdditionalJSONBoundaries",
    "package": "stage1/cohere/json",
    "file": "stage1/cohere/json/port_test.go",
    "seconds": 8.97,
    "oracle": "Executed Go cohere answers for fixed parser/formatter boundaries; source Node and sanitized native output must match.",
    "oracle_kind": "external-run",
    "kills": [
      "M02",
      "M03"
    ],
    "unique_kills": [],
    "last_proven_fail": "M03: port_test.go:562: Node boundaries case 0 boundary/0/probe.json byte 7: got \"ok\\t[\\\\n // dangling\\\\n]\\\\n\"; Go \"ok\\t[\\\\n  // dangling\\\\n]\\\\n\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestPortMatchesGoCohere family"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "PPort"
    ],
    "subsumer_seconds": 35.658,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestAdditionalJSONBoundaries",
      "TestPortMatchesGoCohere family",
      "TestSingleFileStdoutDriver"
    ],
    "evidence": "ADAMIC_MUTANT=M03 ADAMIC_BUILD_CACHE_DIR=/tmp/u103/cache/switch ADAMIC_GATE_SAMPLE=a467d1a1571c43e0f01fc4efcb43890280e4a1ac timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/json/ -run '^(TestPortMatchesGoCohere_[0-9]{3}|TestAdditionalJSONBoundaries|TestSingleFileStdoutDriver)$'; selector file /tmp/u103/selector contains M03 > M03.log; port_test.go:562: Node boundaries case 0 boundary/0/probe.json byte 7: got \"ok\\t[\\\\n // dangling\\\\n]\\\\n\"; Go \"ok\\t[\\\\n  // dangling\\\\n]\\\\n\"",
    "members": [
      "TestAdditionalJSONBoundaries"
    ],
    "observed_assertion": "port_test.go:562: Node boundaries case 0 boundary/0/probe.json byte 7: got \"ok\\t[\\\\n // dangling\\\\n]\\\\n\"; Go \"ok\\t[\\\\n  // dangling\\\\n]\\\\n\"",
    "subsumption_basis_mutants": 2
  },
  {
    "test": "TestSingleFileStdoutDriver",
    "package": "stage1/cohere/json",
    "file": "stage1/cohere/json/port_test.go",
    "seconds": 8.941,
    "oracle": "Executed Go cohere answers for two single-file fixtures; source Node and sanitized native stdout must match.",
    "oracle_kind": "external-run",
    "kills": [
      "M03"
    ],
    "unique_kills": [],
    "last_proven_fail": "M03: port_test.go:612: Node file driver case 1 package.json byte 1: got \" \\\"text\\\": \\\"\u00e9\ud83d\ude00\\\",\"; Go \"  \\\"text\\\": \\\"\u00e9\ud83d\ude00\\\",\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestAdditionalJSONBoundaries"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "PPort"
    ],
    "subsumer_seconds": 8.97,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestAdditionalJSONBoundaries",
      "TestPortMatchesGoCohere family",
      "TestSingleFileStdoutDriver"
    ],
    "evidence": "ADAMIC_MUTANT=M03 ADAMIC_BUILD_CACHE_DIR=/tmp/u103/cache/switch ADAMIC_GATE_SAMPLE=a467d1a1571c43e0f01fc4efcb43890280e4a1ac timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/json/ -run '^(TestPortMatchesGoCohere_[0-9]{3}|TestAdditionalJSONBoundaries|TestSingleFileStdoutDriver)$'; selector file /tmp/u103/selector contains M03 > M03.log; port_test.go:612: Node file driver case 1 package.json byte 1: got \" \\\"text\\\": \\\"\u00e9\ud83d\ude00\\\",\"; Go \"  \\\"text\\\": \\\"\u00e9\ud83d\ude00\\\",\"",
    "members": [
      "TestSingleFileStdoutDriver"
    ],
    "observed_assertion": "port_test.go:612: Node file driver case 1 package.json byte 1: got \" \\\"text\\\": \\\"\u00e9\ud83d\ude00\\\",\"; Go \"  \\\"text\\\": \\\"\u00e9\ud83d\ude00\\\",\"",
    "subsumption_basis_mutants": 1
  },
  {
    "test": "TestProgressGuard",
    "package": "stage1/cohere/json",
    "file": "stage1/cohere/json/progress_test.go",
    "seconds": 5.014,
    "oracle": "Self: named guard timeout reasons and exact eight-byte stdout/stderr progress fixtures. Directly checks childguard, not formatter output.",
    "oracle_kind": "self",
    "kills": [
      "M05"
    ],
    "unique_kills": [
      "M05"
    ],
    "last_proven_fail": "M05: progress_test.go:57: stderr progress was killed: \"xxxxxx\" stalled: no first output for 3s after 3.000702945s, load 2.45 (/usr/bin/sh)",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "probe_kills": [
      "PGuard"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestProgressGuard"
    ],
    "evidence": "ADAMIC_MUTANT=M05 ADAMIC_BUILD_CACHE_DIR=/tmp/u103/cache/M05-bounded ADAMIC_GATE_SAMPLE=a467d1a1571c43e0f01fc4efcb43890280e4a1ac timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/json/ -run '^TestProgressGuard$/(progress_outlasts_startup|stderr_is_progress|never_prints|overall_backstop)$' > M05-bounded.log; progress_test.go:57: stderr progress was killed: \"xxxxxx\" stalled: no first output for 3s after 3.000702945s, load 2.45 (/usr/bin/sh)",
    "members": [
      "TestProgressGuard"
    ],
    "observed_assertion": "progress_test.go:57: stderr progress was killed: \"xxxxxx\" stalled: no first output for 3s after 3.000702945s, load 2.45 (/usr/bin/sh)",
    "bounded_subcases": [
      "never prints",
      "progress outlasts startup",
      "stderr is progress",
      "overall backstop"
    ],
    "unknown_subcases": [
      "never returns under M05 (over budget)"
    ]
  },
  {
    "test": "TestRepositoryCorpusMutants",
    "package": "stage1/cohere/json",
    "file": "stage1/cohere/json/repository_test.go",
    "seconds": 0.059,
    "oracle": "Git runs on private repositories; self-written rejection messages must name seven planted corpus faults. W03 disables pin comparison.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W03: repository_test.go:251: mutant survived or lost its name: <nil>",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestRepositoryCorpusMutants"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u103/cache/W03 ADAMIC_GATE_SAMPLE=a467d1a1571c43e0f01fc4efcb43890280e4a1ac timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/json/ -run '^(TestRepositoryCorpusMutants)$' > W03.log; repository_test.go:251: mutant survived or lost its name: <nil>",
    "members": [
      "TestRepositoryCorpusMutants"
    ],
    "observed_assertion": "repository_test.go:251: mutant survived or lost its name: <nil>",
    "construction_or_witness_kills": [
      "W03"
    ]
  },
  {
    "test": "TestRepositoryLandingWithoutPinEdit",
    "package": "stage1/cohere/json",
    "file": "stage1/cohere/json/repository_test.go",
    "seconds": 0.032,
    "oracle": "Git runs a later private HEAD; self-written tracked-count and unchanged-pin assertions.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "H03: repository_test.go:270: landing changed pin or parity: count 0: <nil>",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [
      "PCorpus"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestRepositoryLandingWithoutPinEdit"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u103/cache/H03 ADAMIC_GATE_SAMPLE=a467d1a1571c43e0f01fc4efcb43890280e4a1ac timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/json/ -run '^(TestRepositoryLandingWithoutPinEdit)$' > H03.log; repository_test.go:270: landing changed pin or parity: count 0: <nil>",
    "members": [
      "TestRepositoryLandingWithoutPinEdit"
    ],
    "observed_assertion": "repository_test.go:270: landing changed pin or parity: count 0: <nil>",
    "construction_or_witness_kills": [
      "H03"
    ]
  },
  {
    "test": "TestRepositoryRequiresGit",
    "package": "stage1/cohere/json",
    "file": "stage1/cohere/json/repository_test.go",
    "seconds": 0.01,
    "oracle": "Git runs without repository metadata; self-written named metadata rejection must fire. The assertion checks only the named wrapper substring; another Git failure carrying it could also pass.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "H04: repository_test.go:285: export without Git silently fell back to walking: JSON corpus unavailable: git rev-parse --show-toplevel: exit status 128: fatal: not a git repository (or any parent up to mount point /)",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [
      "PCorpus"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestRepositoryRequiresGit"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u103/cache/H04 ADAMIC_GATE_SAMPLE=a467d1a1571c43e0f01fc4efcb43890280e4a1ac timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/json/ -run '^(TestRepositoryRequiresGit)$' > H04.log; repository_test.go:285: export without Git silently fell back to walking: JSON corpus unavailable: git rev-parse --show-toplevel: exit status 128: fatal: not a git repository (or any parent up to mount point /)",
    "members": [
      "TestRepositoryRequiresGit"
    ],
    "observed_assertion": "repository_test.go:285: export without Git silently fell back to walking: JSON corpus unavailable: git rev-parse --show-toplevel: exit status 128: fatal: not a git repository (or any parent up to mount point /)",
    "construction_or_witness_kills": [
      "H04"
    ]
  }
]
```

| ID | Starting-commit file:line | Change | Failed rows |
|---|---|---|---|
| M01 (production) | stage1/cohere/json/formatter.ts:8 | base === 'package-lock.json' → base === 'lock.json' |  |
| M02 (production) | stage1/cohere/json/parser.ts:115 | this.text.slice(0, this.position).split('\n').length → this.text.slice(0, this.position).split('\n').length + 1 | TestPortMatchesGoCohere family, TestAdditionalJSONBoundaries |
| M03 (production) | stage1/cohere/json/doc.ts:176 | command.indent + 2 → command.indent + 1 | TestPortMatchesGoCohere family, TestAdditionalJSONBoundaries, TestSingleFileStdoutDriver |
| M04 (production) | stage1/cohere/json/width.ts:79 | return text.length; → return text.length + 1; | TestPortMatchesGoCohere family |
| M05 (production) | internal/childguard/childguard.go:59 | if len(p) > 0 { → if len(p) == 0 { | TestProgressGuard |
| H01 (setup) | stage1/cohere/json/port_matches_prepared_split_test.go:155 | portMatchesShared.ready = true → portMatchesShared.ready = false | TestPortMatchesGoCohereSplit_Setup |
| H02 (setup) | stage1/cohere/json/port_matches_prepared_split_test.go:417 | os.WriteFile(filepath.Join(dir, "main.c"), → os.WriteFile(filepath.Join(dir, "missing-main.c"), | TestProduct_JSON family |
| H03 (setup) | stage1/cohere/json/repository_test.go:131 | return pin, len(tracked), corpusPinError(expected, pin) → return pin, len(tracked) % 3, corpusPinError(expected, pin) | TestRepositoryLandingWithoutPinEdit |
| H04 (setup) | stage1/cohere/json/repository_test.go:28 | JSON corpus requires usable Git metadata: git %s: %w: %s → JSON corpus unavailable: git %s: %w: %s | TestRepositoryRequiresGit |
| W01 (witness) | stage1/cohere/json/port_test.go:202 | if string(result.stdout) == expected { → if true { | TestPortMatchesGoCohereSplitUnion |
| W02 (witness) | stage1/cohere/json/port_test.go:505 | if string(side.result.stdout) == expected { → if true { | TestThreePortMutantsAreCaught |
| W03 (witness) | stage1/cohere/json/repository_test.go:131 | return pin, len(tracked), corpusPinError(expected, pin) → return pin, len(tracked), nil | TestRepositoryCorpusMutants |

Survivor M01: restored source Node outputs 'ok\t[\\n  1,\\n  2\\n]\\n\n'; the rebuilt standalone sanitized mutant outputs 'ok\t[1, 2]\\n\n'. This is unguarded in the observed matrix. Full-corpus coverage of it is unknown. Both witness commands exit zero.

Brief ambiguities, corrections, costs and limitations:

- The brief names 14 rows and 141 functions. Applying its own shared-checker family rule yields 11 rows: 128 portMatchesRun wrappers form one family and four portMatchesProduct recipe wrappers form another. Setup, the planted union witness, and the other seven functions have different assertions and remain separate. family-members.json lists every member. All 141 names still exist in the four supplied files; none moved or vanished. Thousands of four-digit wrappers now exist but use jsonPortTopShard, a different checker with additional assertions and optional benchmark work. They were not silently added to this slice.
- The reference commit 8de93800f4 is stale. The mandated fetch selected a467d1a1571c43e0f01fc4efcb43890280e4a1ac. All standalone diffs apply to that starting commit, and mutant locations use it. Raw matrix failure lines in prepared_split differ by one because the probe guard precedes wrapper declarations. rows.json keeps both observed_assertion and mapped origin/main failure locations.
- I initially wrote start-commands.json inside the checkout. fileCorpus walks every JSON file, including ignored/untracked review files, so that contaminated attempt failed on my evidence file. It is retained as a failed setup attempt, not a legitimate baseline or mutant kill. Evidence was moved outside the corpus under /tmp/u103/evidence, accessed through a directory symlink that WalkDir does not follow. The corrected whole-package baseline had no assertion failures before its 90s timeout.
- The clean requested full scope and the full 128-member family alone also exceeded 90s without assertion failures. The existing landing-sample switch ADAMIC_GATE_SAMPLE=a467d1a1571c43e0f01fc4efcb43890280e4a1ac selects 114/3602 physical files at stride 32, offset 1, retaining generated controls. The resulting 148 cases across 128 wrappers passed cleanly. No corpus file, pin or oracle was edited to obtain that sample. This is an extra input bound beyond the supplied test slice; unsampled input kills remain unknown.
- The complete family does not have a measured three-run median below the limit. Its seconds field is null, full_corpus_over_budget is true, and sample_seconds records the three sampled binary runs. Its slow-worthy verdict rests on the full family exceeding 90s and its bounded unique M04 catch. All other row medians are three independent -count=1 binary elapsed readings. Subsumption against this family reports its sampled median, not an invented full-corpus cost.
- Every verdict and unique_kills is bounded to the listed matrix rows and input sample. Package-wide and repository-wide uniqueness are not claimed. The central replay can apply the production diffs without switches and settle the rest. The initial whole-package run is the only whole-package coverage attempt; no other package suite was run.
- The five production mutants were fixed from read production functions before mutant outcomes, spanning filename selection, parser message location, document indentation, width and child progress. Four native port mutants respect the rebuild cap; the fifth is a direct Go guard mutant. This is fewer than the suggested three mutants per row. Construction and witness changes are exceptions judged separately and do not inflate production kills.
- M01's survival is real: its separately built sanitized native product changes package-lock.json formatting from pretty to compact. The sampled matrix misses this behavior. That does not establish that the full corpus is unguarded, nor does it make M01 equivalent. The witness retains the original source Node runtime and does not mutate Go cohere.
- PSetup makes portMatchesPrepare return before constructing anything; the setup test still passes. PProducts makes portMatchesProduct return an empty path; all four product members still pass. These rows are vacuous even though H01 and H02 show they can fail on other construction faults. A successful build return alone is a weak product oracle. The lowered product member also passes H02 even though the expected main.c file was never written; the native product members catch the missing artifact.
- PPort probes the formatter's public format(name,text) entry, which the batch and single-file main driver invoke. All nonempty sampled agreement buckets and both fixed execution rows reject its empty string answer. Empty sampled buckets pass and are listed in vacuous_subcases. The complete family is not vacuous.
- PCorpus returns the empty validation result. RequiresGit rejects it directly. LandingWithoutPinEdit fails earlier at newRepositoryFixture's baseline count check, so that probe proves rejection of an empty construction, not the later landing assertion. H03 independently reaches the later landing assertion with count 0 after a valid count-2 fixture baseline.
- ProgressGuard directly specifies childguard behavior through executable child fixtures and self-written expectations. It is not classified as a port witness. M05 makes the watched writer stop recording real output, causing talking children to be killed at the first-output deadline. Its never-returns case loses the stall deadline and cooks the full row at 90s. The terminating four-subcase bounded rerun fails in both talking-child subcases; never-returns remains unknown for M05. No stalled run was allowed past the binary budget.
- SplitUnion, ThreePortMutantsAreCaught and RepositoryCorpusMutants are witnesses. W01 forces their shared comparison to report agreement; W02 forces the mutant witness's inline equality comparison to report agreement; W03 disables the corpus-pin comparison. Each witness fails under its own weakening. Production-mutant preconditions are excluded from their kills. W03 leaves fixture baseline validation valid and catches surviving one-byte/missing-provisioned mutations, rather than breaking a prerequisite.
- AdditionalJSONBoundaries is subsumed on M02 and M03 by the agreement family. SingleFileStdoutDriver is subsumed on M03 by AdditionalJSONBoundaries, the faster observed subsumer. These hints rest on two and one kills respectively, not exhaustive mutation coverage or deletion advice.
- Go cohere remains the external formatter oracle and is never mutated. Product/setup and guard expectations are self-written; Git-backed repository rows also execute Git and compare their own count/message invariants. No hand-copied outside authority is claimed. The parser-message mutant proves a diagnostic-location assertion, not semantic parser correctness. Formatter comparisons check full byte output, exit code and stderr, not only answer counts.
- Optional Prettier 3.9.6 was available outside the checkout and enabled for whole-package attempts; its dependencies were refreshed afterward, recorded separately. No requested row requires it and no requested row skipped in the successful bounded baseline. Whole-package skip/completion status after timeout remains unknown. ADAMIC_JSON_BENCH is optional repeated timing work, not a missing SDK check; it was not enabled.
- The source inventory includes all port declarations and childguard functions plus construction helper declarations as an upper bound. It is not an exact dynamic transitive TypeScript call graph. callers.txt records the shared checker/entry calls. Full compiler and runtime behavior beyond these mutants was not audited.
- Native alternatives read one selector file at process start from a helper inside an existing copied port module, so no copied-file-list edit was needed. Source changes were compiled once per immutable product, then selected at runtime. Standalone production diffs contain no selector. M05 uses its own ADAMIC_BUILD_CACHE_DIR; native standalone builds and probe builds also use independent directories. ADAMIC_NATIVE_SPLIT=1 enables splitting and is not treated as a cache key. Unchanged compiler action-cache units may be reused; measured build times are actual warm-workspace command times, not asserted cold clang work.
- All five production diffs, five probe diffs and seven allowed harness diffs apply to the starting commit. Native production and PPort probe diffs passed their own sanitized port builds. Go production/harness/probe diffs passed the corresponding go vet checks. switch-clean validates the instrumented source before selections. All switches and harness edits were restored before the final clean bounded run.

Timing and coverage:

Warm tool setup: skipped, 0s; nproc=5. stage3/api npm ci: 0.444s. Native standalone builds: validate-M01 7.693s, validate-M02 2.329s, validate-M03 1.784s, validate-M04 1.756s. Instrumented clean port run: 58.344s including product preparation and execution. The clean sampled scope binary passed in 46.315s. Three-run timing commands total 280.046 shell seconds. Detailed per-command builds, matrix runs, probes and allowed harness checks are in mutant-commands.json and finish-commands.json.

Not covered: full-package completion, full-family median, unsampled inputs, four-digit checker family, exact transitive source coverage and repo-wide uniqueness. No production source change is retained and no pull request is opened. Large raw logs are compressed without changing their contents; decompress *.log.gz to recover original command redirections.

Restored scoped binary: 37.466s, all 141 requested member tests pass and none skip. Production matrix command times: M01 39.858s, M02 39.479s, M03 13.711s, M04 18.566s, M05 91.809s. Native probe build: 7.8s. Optional dependency refresh: 0.483s. The binary timeout fired at 90s; additional stack-dump/package-reporting time is not further test execution.
