Unit u145 audited from a7448d73cd17f16362b6cbc5c5c111080da64e43.
All 109 names present, grouped into 14 rows.
Four native-built production mutants were caught; no unit survivors.
Findings: 2 bounded sacred, 4 setup-check, 4 witness, 1 subsumed, 1 helper, 2 cannot-judge.
Evidence pushed on test-audit/stage1-cohere-typeaware-inspect_shards under review/test-audit/stage1-cohere-typeaware-inspect_shards/.

```json
[
  {
    "test": "TestInspectRequestRefusals",
    "package": "stage1/cohere/typeaware",
    "file": "stage1/cohere/typeaware/inspect_shards_test.go:17",
    "seconds": 32.358,
    "oracle": "Self-written exit 0 validity control and exit 70 plus named refusal substrings. Validity does not check returned facts or their count.",
    "oracle_kind": "self",
    "kills": [
      "M4"
    ],
    "unique_kills": [
      "M4"
    ],
    "last_proven_fail": "M4: inspect_shards_test.go:73: wrong-kind-mutant-run: exit status 70",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "probe_kills": [
      "P2"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestFactsDecoderGuards",
      "TestInspectRequestRefusals",
      "TestSixRuleAgreementAndMutants_000",
      "TestSixRuleAgreementAndMutants_033"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestInspectRequestRefusals|TestSixRuleAgreementAndMutants_000)$' > M4.log 2>&1; inspect_shards_test.go:73: wrong-kind-mutant-run: exit status 70",
    "control_kills": [],
    "timing_observations": [
      70.449,
      30.81,
      32.358
    ],
    "vacuous_subcases": [
      "TestInspectRequestRefusals/shard-003",
      "TestInspectRequestRefusals/shard-000",
      "TestInspectRequestRefusals/shard-005"
    ]
  },
  {
    "test": "TestShadowIndexMissingBinding_000",
    "package": "stage1/cohere/typeaware",
    "file": "stage1/cohere/typeaware/profile_test.go:51",
    "seconds": 4.294,
    "oracle": "Self-written exit 70 and missing binding index substring for a native built-in mutant; fresh guard weakening proves the witness.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W5: profile_test.go:64: missing binding index was not refused: <nil>",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestShadowIndexMissingBinding_000"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u145/cache/W5 ADAMIC_TYPESCRIPT_SOURCE=/tmp/u145/typescript ADAMIC_TYPEAWARE_BENCH=1 timeout 120 go test -overlay /tmp/u145/W5.json -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run ^TestShadowIndexMissingBinding_000$ > W5-fresh.log 2>&1; profile_test.go:64: missing binding index was not refused: <nil>",
    "control_kills": [
      "W5"
    ],
    "timing_observations": [
      61.738,
      4.228,
      4.294
    ]
  },
  {
    "test": "TestSharedProductPublication",
    "package": "stage1/cohere/typeaware",
    "file": "stage1/cohere/typeaware/shared_test.go:109",
    "seconds": 0.013,
    "oracle": "Hand-written expected construction results, errors or process markers",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S1: shared_test.go:154: different keys shared a product: /tmp/adamic-gate/typeaware-products-276804829/product-2149593182/ready, <nil>",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestSharedProductPublication"
    ],
    "evidence": "timeout 120 go test -overlay /tmp/u145/controls-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestSharedProductPublication|TestSixBuildCallbackUsesProductDirectory|TestSixShardUnionRejectsLossAndDuplication|TestSixShardPlantedDisagreement|TestTypeAwareUnitDeadline|TestTypeAwareContextKillsCompilerGroup|TestSixPinnedFlags|TestPinnedTypeFlags|TestTypeAwareNativeBuildChild)$' > S1.log 2>&1; shared_test.go:154: different keys shared a product: /tmp/adamic-gate/typeaware-products-276804829/product-2149593182/ready, <nil>",
    "control_kills": [
      "S1"
    ],
    "timing_observations": [
      0.016,
      0.012,
      0.013
    ]
  },
  {
    "test": "TestSixBuildCallbackUsesProductDirectory",
    "package": "stage1/cohere/typeaware",
    "file": "stage1/cohere/typeaware/six_builds_test.go:342",
    "seconds": 0.041,
    "oracle": "Hand-written expected construction results, errors or process markers",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S2: six_builds_test.go:368: callback wrote outside product directory: <nil>",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestSixBuildCallbackUsesProductDirectory"
    ],
    "evidence": "timeout 120 go test -overlay /tmp/u145/controls-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestSharedProductPublication|TestSixBuildCallbackUsesProductDirectory|TestSixShardUnionRejectsLossAndDuplication|TestSixShardPlantedDisagreement|TestTypeAwareUnitDeadline|TestTypeAwareContextKillsCompilerGroup|TestSixPinnedFlags|TestPinnedTypeFlags|TestTypeAwareNativeBuildChild)$' > S2.log 2>&1; six_builds_test.go:368: callback wrote outside product directory: <nil>",
    "control_kills": [
      "S2"
    ],
    "timing_observations": [
      0.041,
      0.15,
      0.025
    ]
  },
  {
    "test": "TestSixShardUnionRejectsLossAndDuplication",
    "package": "stage1/cohere/typeaware",
    "file": "stage1/cohere/typeaware/six_shards_test.go:232",
    "seconds": 0.019,
    "oracle": "Hand-written expected construction results, errors or process markers",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: six_shards_test.go:254: invalid union accepted: [{left [a] <nil>} {right [c] <nil>}]",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestSixShardUnionRejectsLossAndDuplication"
    ],
    "evidence": "timeout 120 go test -overlay /tmp/u145/controls-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestSharedProductPublication|TestSixBuildCallbackUsesProductDirectory|TestSixShardUnionRejectsLossAndDuplication|TestSixShardPlantedDisagreement|TestTypeAwareUnitDeadline|TestTypeAwareContextKillsCompilerGroup|TestSixPinnedFlags|TestPinnedTypeFlags|TestTypeAwareNativeBuildChild)$' > W1.log 2>&1; six_shards_test.go:254: invalid union accepted: [{left [a] <nil>} {right [c] <nil>}]",
    "control_kills": [
      "W1"
    ],
    "timing_observations": [
      0.013,
      0.024,
      0.019
    ]
  },
  {
    "test": "TestSixShardPlantedDisagreement",
    "package": "stage1/cohere/typeaware",
    "file": "stage1/cohere/typeaware/six_shards_test.go:268",
    "seconds": 0.075,
    "oracle": "Hand-written expected construction results, errors or process markers",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W2: six_shards_test.go:308: holding shard failed to catch planted case: <nil>",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestSixShardPlantedDisagreement"
    ],
    "evidence": "timeout 120 go test -overlay /tmp/u145/controls-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestSharedProductPublication|TestSixBuildCallbackUsesProductDirectory|TestSixShardUnionRejectsLossAndDuplication|TestSixShardPlantedDisagreement|TestTypeAwareUnitDeadline|TestTypeAwareContextKillsCompilerGroup|TestSixPinnedFlags|TestPinnedTypeFlags|TestTypeAwareNativeBuildChild)$' > W2.log 2>&1; six_shards_test.go:308: holding shard failed to catch planted case: <nil>",
    "control_kills": [
      "W2"
    ],
    "timing_observations": [
      0.075,
      0.078,
      0.056
    ]
  },
  {
    "test": "TestSixRuleAgreementAndMutants family",
    "package": "stage1/cohere/typeaware",
    "file": [
      "stage1/cohere/typeaware/six_top_shards_test.go:7",
      "stage1/cohere/typeaware/six_top_shards_test.go:8",
      "stage1/cohere/typeaware/six_top_shards_test.go:9",
      "stage1/cohere/typeaware/six_top_shards_test.go:10",
      "stage1/cohere/typeaware/six_top_shards_test.go:11",
      "stage1/cohere/typeaware/six_top_shards_test.go:12",
      "stage1/cohere/typeaware/six_top_shards_test.go:13",
      "stage1/cohere/typeaware/six_top_shards_test.go:14",
      "stage1/cohere/typeaware/six_top_shards_test.go:15",
      "stage1/cohere/typeaware/six_top_shards_test.go:16",
      "stage1/cohere/typeaware/six_top_shards_test.go:17",
      "stage1/cohere/typeaware/six_top_shards_test.go:18",
      "stage1/cohere/typeaware/six_top_shards_test.go:19",
      "stage1/cohere/typeaware/six_top_shards_test.go:20",
      "stage1/cohere/typeaware/six_top_shards_test.go:21",
      "stage1/cohere/typeaware/six_top_shards_test.go:22",
      "stage1/cohere/typeaware/six_top_shards_test.go:23",
      "stage1/cohere/typeaware/six_top_shards_test.go:24",
      "stage1/cohere/typeaware/six_top_shards_test.go:25",
      "stage1/cohere/typeaware/six_top_shards_test.go:26",
      "stage1/cohere/typeaware/six_top_shards_test.go:27",
      "stage1/cohere/typeaware/six_top_shards_test.go:28",
      "stage1/cohere/typeaware/six_top_shards_test.go:29",
      "stage1/cohere/typeaware/six_top_shards_test.go:30",
      "stage1/cohere/typeaware/six_top_shards_test.go:31",
      "stage1/cohere/typeaware/six_top_shards_test.go:32",
      "stage1/cohere/typeaware/six_top_shards_test.go:33",
      "stage1/cohere/typeaware/six_top_shards_test.go:34",
      "stage1/cohere/typeaware/six_top_shards_test.go:35",
      "stage1/cohere/typeaware/six_top_shards_test.go:36",
      "stage1/cohere/typeaware/six_top_shards_test.go:37",
      "stage1/cohere/typeaware/six_top_shards_test.go:38",
      "stage1/cohere/typeaware/six_top_shards_test.go:39",
      "stage1/cohere/typeaware/six_top_shards_test.go:40",
      "stage1/cohere/typeaware/six_top_shards_test.go:41",
      "stage1/cohere/typeaware/six_top_shards_test.go:42",
      "stage1/cohere/typeaware/six_top_shards_test.go:43",
      "stage1/cohere/typeaware/six_top_shards_test.go:44",
      "stage1/cohere/typeaware/six_top_shards_test.go:45",
      "stage1/cohere/typeaware/six_top_shards_test.go:46",
      "stage1/cohere/typeaware/six_top_shards_test.go:47",
      "stage1/cohere/typeaware/six_top_shards_test.go:48",
      "stage1/cohere/typeaware/six_top_shards_test.go:49",
      "stage1/cohere/typeaware/six_top_shards_test.go:50",
      "stage1/cohere/typeaware/six_top_shards_test.go:51",
      "stage1/cohere/typeaware/six_top_shards_test.go:52",
      "stage1/cohere/typeaware/six_top_shards_test.go:53",
      "stage1/cohere/typeaware/six_top_shards_test.go:54",
      "stage1/cohere/typeaware/six_top_shards_test.go:55",
      "stage1/cohere/typeaware/six_top_shards_test.go:56",
      "stage1/cohere/typeaware/six_top_shards_test.go:57",
      "stage1/cohere/typeaware/six_top_shards_test.go:58",
      "stage1/cohere/typeaware/six_top_shards_test.go:59",
      "stage1/cohere/typeaware/six_top_shards_test.go:60",
      "stage1/cohere/typeaware/six_top_shards_test.go:61",
      "stage1/cohere/typeaware/six_top_shards_test.go:62",
      "stage1/cohere/typeaware/six_top_shards_test.go:63",
      "stage1/cohere/typeaware/six_top_shards_test.go:64",
      "stage1/cohere/typeaware/six_top_shards_test.go:65",
      "stage1/cohere/typeaware/six_top_shards_test.go:66",
      "stage1/cohere/typeaware/six_top_shards_test.go:67",
      "stage1/cohere/typeaware/six_top_shards_test.go:68",
      "stage1/cohere/typeaware/six_top_shards_test.go:69",
      "stage1/cohere/typeaware/six_top_shards_test.go:70",
      "stage1/cohere/typeaware/six_top_shards_test.go:71",
      "stage1/cohere/typeaware/six_top_shards_test.go:72",
      "stage1/cohere/typeaware/six_top_shards_test.go:73",
      "stage1/cohere/typeaware/six_top_shards_test.go:74",
      "stage1/cohere/typeaware/six_top_shards_test.go:75",
      "stage1/cohere/typeaware/six_top_shards_test.go:76",
      "stage1/cohere/typeaware/six_top_shards_test.go:77",
      "stage1/cohere/typeaware/six_top_shards_test.go:78",
      "stage1/cohere/typeaware/six_top_shards_test.go:79",
      "stage1/cohere/typeaware/six_top_shards_test.go:80",
      "stage1/cohere/typeaware/six_top_shards_test.go:81",
      "stage1/cohere/typeaware/six_top_shards_test.go:82",
      "stage1/cohere/typeaware/six_top_shards_test.go:83",
      "stage1/cohere/typeaware/six_top_shards_test.go:84",
      "stage1/cohere/typeaware/six_top_shards_test.go:85",
      "stage1/cohere/typeaware/six_top_shards_test.go:86",
      "stage1/cohere/typeaware/six_top_shards_test.go:87",
      "stage1/cohere/typeaware/six_top_shards_test.go:88",
      "stage1/cohere/typeaware/six_top_shards_test.go:89",
      "stage1/cohere/typeaware/six_top_shards_test.go:90",
      "stage1/cohere/typeaware/six_top_shards_test.go:91",
      "stage1/cohere/typeaware/six_top_shards_test.go:92",
      "stage1/cohere/typeaware/six_top_shards_test.go:93",
      "stage1/cohere/typeaware/six_top_shards_test.go:94",
      "stage1/cohere/typeaware/six_top_shards_test.go:95",
      "stage1/cohere/typeaware/six_top_shards_test.go:96",
      "stage1/cohere/typeaware/six_top_shards_test.go:97",
      "stage1/cohere/typeaware/six_top_shards_test.go:98",
      "stage1/cohere/typeaware/six_top_shards_test.go:99",
      "stage1/cohere/typeaware/six_top_shards_test.go:100",
      "stage1/cohere/typeaware/six_top_shards_test.go:101",
      "stage1/cohere/typeaware/six_top_shards_test.go:6"
    ],
    "seconds": null,
    "oracle": "Exact stdout bytes from Go cohere; sanitizer stderr and exits; hand-written reporting controls. Only reached leaf 000 was replayed under production mutants.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M1",
      "M2",
      "M3"
    ],
    "unique_kills": [
      "M2"
    ],
    "last_proven_fail": "M3: suite_test.go:201: generated-native: exit status 70",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P2",
      "P3"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestFactsDecoderGuards",
      "TestInspectRequestRefusals",
      "TestSixRuleAgreementAndMutants_000",
      "TestSixRuleAgreementAndMutants_033"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestFactsDecoderGuards|TestSixRuleAgreementAndMutants_000)$' > M3.log 2>&1; suite_test.go:201: generated-native: exit status 70",
    "control_kills": [],
    "timing_observations": [
      null,
      null,
      null
    ],
    "members": [
      "TestSixRuleAgreementAndMutants_000",
      "TestSixRuleAgreementAndMutants_001",
      "TestSixRuleAgreementAndMutants_002",
      "TestSixRuleAgreementAndMutants_003",
      "TestSixRuleAgreementAndMutants_004",
      "TestSixRuleAgreementAndMutants_005",
      "TestSixRuleAgreementAndMutants_006",
      "TestSixRuleAgreementAndMutants_007",
      "TestSixRuleAgreementAndMutants_008",
      "TestSixRuleAgreementAndMutants_009",
      "TestSixRuleAgreementAndMutants_010",
      "TestSixRuleAgreementAndMutants_011",
      "TestSixRuleAgreementAndMutants_012",
      "TestSixRuleAgreementAndMutants_013",
      "TestSixRuleAgreementAndMutants_014",
      "TestSixRuleAgreementAndMutants_015",
      "TestSixRuleAgreementAndMutants_016",
      "TestSixRuleAgreementAndMutants_017",
      "TestSixRuleAgreementAndMutants_018",
      "TestSixRuleAgreementAndMutants_019",
      "TestSixRuleAgreementAndMutants_020",
      "TestSixRuleAgreementAndMutants_021",
      "TestSixRuleAgreementAndMutants_022",
      "TestSixRuleAgreementAndMutants_023",
      "TestSixRuleAgreementAndMutants_024",
      "TestSixRuleAgreementAndMutants_025",
      "TestSixRuleAgreementAndMutants_026",
      "TestSixRuleAgreementAndMutants_027",
      "TestSixRuleAgreementAndMutants_028",
      "TestSixRuleAgreementAndMutants_029",
      "TestSixRuleAgreementAndMutants_030",
      "TestSixRuleAgreementAndMutants_031",
      "TestSixRuleAgreementAndMutants_032",
      "TestSixRuleAgreementAndMutants_033",
      "TestSixRuleAgreementAndMutants_034",
      "TestSixRuleAgreementAndMutants_035",
      "TestSixRuleAgreementAndMutants_036",
      "TestSixRuleAgreementAndMutants_037",
      "TestSixRuleAgreementAndMutants_038",
      "TestSixRuleAgreementAndMutants_039",
      "TestSixRuleAgreementAndMutants_040",
      "TestSixRuleAgreementAndMutants_041",
      "TestSixRuleAgreementAndMutants_042",
      "TestSixRuleAgreementAndMutants_043",
      "TestSixRuleAgreementAndMutants_044",
      "TestSixRuleAgreementAndMutants_045",
      "TestSixRuleAgreementAndMutants_046",
      "TestSixRuleAgreementAndMutants_047",
      "TestSixRuleAgreementAndMutants_048",
      "TestSixRuleAgreementAndMutants_049",
      "TestSixRuleAgreementAndMutants_050",
      "TestSixRuleAgreementAndMutants_051",
      "TestSixRuleAgreementAndMutants_052",
      "TestSixRuleAgreementAndMutants_053",
      "TestSixRuleAgreementAndMutants_054",
      "TestSixRuleAgreementAndMutants_055",
      "TestSixRuleAgreementAndMutants_056",
      "TestSixRuleAgreementAndMutants_057",
      "TestSixRuleAgreementAndMutants_058",
      "TestSixRuleAgreementAndMutants_059",
      "TestSixRuleAgreementAndMutants_060",
      "TestSixRuleAgreementAndMutants_061",
      "TestSixRuleAgreementAndMutants_062",
      "TestSixRuleAgreementAndMutants_063",
      "TestSixRuleAgreementAndMutants_064",
      "TestSixRuleAgreementAndMutants_065",
      "TestSixRuleAgreementAndMutants_066",
      "TestSixRuleAgreementAndMutants_067",
      "TestSixRuleAgreementAndMutants_068",
      "TestSixRuleAgreementAndMutants_069",
      "TestSixRuleAgreementAndMutants_070",
      "TestSixRuleAgreementAndMutants_071",
      "TestSixRuleAgreementAndMutants_072",
      "TestSixRuleAgreementAndMutants_073",
      "TestSixRuleAgreementAndMutants_074",
      "TestSixRuleAgreementAndMutants_075",
      "TestSixRuleAgreementAndMutants_076",
      "TestSixRuleAgreementAndMutants_077",
      "TestSixRuleAgreementAndMutants_078",
      "TestSixRuleAgreementAndMutants_079",
      "TestSixRuleAgreementAndMutants_080",
      "TestSixRuleAgreementAndMutants_081",
      "TestSixRuleAgreementAndMutants_082",
      "TestSixRuleAgreementAndMutants_083",
      "TestSixRuleAgreementAndMutants_084",
      "TestSixRuleAgreementAndMutants_085",
      "TestSixRuleAgreementAndMutants_086",
      "TestSixRuleAgreementAndMutants_087",
      "TestSixRuleAgreementAndMutants_088",
      "TestSixRuleAgreementAndMutants_089",
      "TestSixRuleAgreementAndMutants_090",
      "TestSixRuleAgreementAndMutants_091",
      "TestSixRuleAgreementAndMutants_092",
      "TestSixRuleAgreementAndMutants_093",
      "TestSixRuleAgreementAndMutants_094",
      "TestSixRuleAgreementAndMutantsUnion"
    ],
    "timing_status": "Complete family timed out at 90 s; three successful complete-family observations unavailable."
  },
  {
    "test": "TestSixRuleAgreementAndMutants_Setup",
    "package": "stage1/cohere/typeaware",
    "file": "stage1/cohere/typeaware/six_top_shards_test.go:104",
    "seconds": 15.758,
    "oracle": "Hand-written expected construction results, errors or process markers",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S3: suite_test.go:497: enumerated shard count 95 differs from declared 94",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestSixRuleAgreementAndMutants_Setup"
    ],
    "evidence": "timeout 120 go test -overlay /tmp/u145/S3.json -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^TestSixRuleAgreementAndMutants_Setup$' > S3.log 2>&1; suite_test.go:497: enumerated shard count 95 differs from declared 94",
    "control_kills": [
      "S3"
    ],
    "timing_observations": [
      16.606,
      15.758,
      12.862
    ]
  },
  {
    "test": "TestSixPinnedFlags",
    "package": "stage1/cohere/typeaware",
    "file": "stage1/cohere/typeaware/suite_test.go:500",
    "seconds": 0.012,
    "oracle": "Pinned Go cohere TypeScript checker TypeFlags constants. Checked union=1<<27=134217728 and mask sum=334017 against source and successful tests. These rows do not execute port constants.",
    "oracle_kind": "external-authority",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestSixPinnedFlags"
    ],
    "evidence": null,
    "control_kills": [],
    "timing_observations": [
      0.014,
      0.011,
      0.012
    ],
    "reason": "A mutation of the Go constants would change the authority; no permitted port mutation can reach these assertions."
  },
  {
    "test": "TestFactsDecoderGuards",
    "package": "stage1/cohere/typeaware",
    "file": "stage1/cohere/typeaware/suite_test.go:516",
    "seconds": 4.107,
    "oracle": "Self-written valid stdout 64 plus exit 70 and named malformed-frame errors. Strict/present boolean inversion M2 passes this row.",
    "oracle_kind": "self",
    "kills": [
      "M1",
      "M3"
    ],
    "unique_kills": [],
    "last_proven_fail": "M3: suite_test.go:545: decoder-valid: exit status 70",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestSixRuleAgreementAndMutants family"
    ],
    "mutants_in_matrix": 3,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestFactsDecoderGuards",
      "TestInspectRequestRefusals",
      "TestSixRuleAgreementAndMutants_000",
      "TestSixRuleAgreementAndMutants_033"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestFactsDecoderGuards|TestSixRuleAgreementAndMutants_000)$' > M3.log 2>&1; suite_test.go:545: decoder-valid: exit status 70",
    "control_kills": [],
    "timing_observations": [
      4.107,
      6.077,
      4.013
    ],
    "subsumption_basis_mutants": 2,
    "subsumer_timing_status": "Complete family exceeded 90 s; no valid median."
  },
  {
    "test": "TestTypeAwareUnitDeadline",
    "package": "stage1/cohere/typeaware",
    "file": "stage1/cohere/typeaware/setup_deadline_typeaware_test.go:22",
    "seconds": 1.3639999999999999,
    "oracle": "Hand-written expected construction results, errors or process markers",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W3: setup_deadline_typeaware_test.go:64: cooked unit left its subprocess alive: <nil>",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestTypeAwareUnitDeadline"
    ],
    "evidence": "timeout 120 go test -overlay /tmp/u145/controls-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestSharedProductPublication|TestSixBuildCallbackUsesProductDirectory|TestSixShardUnionRejectsLossAndDuplication|TestSixShardPlantedDisagreement|TestTypeAwareUnitDeadline|TestTypeAwareContextKillsCompilerGroup|TestSixPinnedFlags|TestPinnedTypeFlags|TestTypeAwareNativeBuildChild)$' > W3.log 2>&1; setup_deadline_typeaware_test.go:64: cooked unit left its subprocess alive: <nil>",
    "control_kills": [
      "W3"
    ],
    "timing_observations": [
      1.445,
      1.3639999999999999,
      1.274
    ]
  },
  {
    "test": "TestTypeAwareNativeBuildChild",
    "package": "stage1/cohere/typeaware",
    "file": "stage1/cohere/typeaware/typeaware_commands_test.go:113",
    "seconds": null,
    "oracle": "Subprocess native builder, no independent oracle",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "helper",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestTypeAwareNativeBuildChild"
    ],
    "evidence": null,
    "control_kills": [],
    "timing_observations": [
      null,
      null,
      null
    ],
    "parents": [
      "TestInspectRequestRefusals",
      "TestShadowIndexMissingBinding_000",
      "TestSixRuleAgreementAndMutants family",
      "TestSixRuleAgreementAndMutants_Setup"
    ]
  },
  {
    "test": "TestTypeAwareContextKillsCompilerGroup",
    "package": "stage1/cohere/typeaware",
    "file": "stage1/cohere/typeaware/typeaware_commands_test.go:134",
    "seconds": 1.324,
    "oracle": "Hand-written expected construction results, errors or process markers",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W4: typeaware_commands_test.go:163: context left compiler descendant alive: <nil>",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestTypeAwareContextKillsCompilerGroup"
    ],
    "evidence": "timeout 120 go test -overlay /tmp/u145/controls-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run '^(TestSharedProductPublication|TestSixBuildCallbackUsesProductDirectory|TestSixShardUnionRejectsLossAndDuplication|TestSixShardPlantedDisagreement|TestTypeAwareUnitDeadline|TestTypeAwareContextKillsCompilerGroup|TestSixPinnedFlags|TestPinnedTypeFlags|TestTypeAwareNativeBuildChild)$' > W4.log 2>&1; typeaware_commands_test.go:163: context left compiler descendant alive: <nil>",
    "control_kills": [
      "W4"
    ],
    "timing_observations": [
      1.324,
      1.331,
      1.321
    ]
  },
  {
    "test": "TestPinnedTypeFlags",
    "package": "stage1/cohere/typeaware",
    "file": "stage1/cohere/typeaware/typeaware_test.go:636",
    "seconds": 0.013,
    "oracle": "Pinned Go cohere TypeScript checker TypeFlags constants. Checked union=1<<27=134217728 and mask sum=334017 against source and successful tests. These rows do not execute port constants.",
    "oracle_kind": "external-authority",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestPinnedTypeFlags"
    ],
    "evidence": null,
    "control_kills": [],
    "timing_observations": [
      0.018,
      0.013,
      0.012
    ],
    "reason": "A mutation of the Go constants would change the authority; no permitted port mutation can reach these assertions."
  }
]
```

| ID | Origin file:line | Change | Observed failed rows |
|---|---|---|---|
| M1 | stage1/cohere/typeaware/frames.ts:15 | value = value * 9 + digit; | TestFactsDecoderGuards, TestSixRuleAgreementAndMutants_000 |
| M2 | stage1/cohere/typeaware/frames.ts:61 | return value === 0; | TestSixRuleAgreementAndMutants_000 |
| M3 | stage1/cohere/typeaware/facts.ts:41 | tuple < 0 | TestFactsDecoderGuards, TestSixRuleAgreementAndMutants_000 |
| M4 | stage1/cohere/typeaware/testdata/fact_cost.ts:9 | count < 3 | TestInspectRequestRefusals |
| S1 | stage1/cohere/typeaware/shared_test.go:43 | products.LoadOrStore("one-key", &product{}) | TestSharedProductPublication |
| S2 | stage1/cohere/typeaware/six_builds_test.go:41 | args[i+1] = command.Args[i+2] | TestSixBuildCallbackUsesProductDirectory |
| W1 | stage1/cohere/typeaware/six_shards_test.go:42 | func sixUnion(expected []string, shards []sixShard) error {  if len(expected) >= 0 { return nil } | TestSixShardUnionRejectsLossAndDuplication |
| W2 | stage1/cohere/typeaware/six_shards_test.go:177 | if false { | TestSixShardPlantedDisagreement |
| W3 | stage1/cohere/typeaware/typeaware_budget_test.go:28 | 		// Command groups retained. | TestTypeAwareUnitDeadline |
| W4 | stage1/cohere/typeaware/typeaware_commands_test.go:33 | syscall.Kill(command.Process.Pid, syscall.SIGKILL) | TestTypeAwareContextKillsCompilerGroup |
| S3 | stage1/cohere/typeaware/suite_test.go:129 | const testSixRuleAgreementAndMutantsShards = 94 | TestSixRuleAgreementAndMutants_Setup |
| W5 | stage1/cohere/typeaware/profile_test.go:34 | source = strings.Replace(source, from, to, 1)  source = strings.ReplaceAll(source, "?? panic('missing binding index')", "?? []") | TestShadowIndexMissingBinding_000 |

S and W IDs are construction/check weakening controls, never production kills. W5 is the permitted witness harness edit: its source-copy helper disables the missing-binding guard before building the built-in mutant. Its native fresh-cache result, not its stale-cache pass, decides the witness. P1 returns empty Types at the decoder API; P2 and P3 empty the native request and Six driver entries. P1 is a nested probe for the Six family, excluded from that family's probe_kills.

Survivors: none among the four production mutants in the executed bounded matrix. M2 survived TestFactsDecoderGuards but was killed by the Six agreement leaf.

Brief ambiguities and execution costs:

- The historical 8de93800f4 file list is stale. Origin/main was a7448d73cd17f16362b6cbc5c5c111080da64e43. TestTypeAwareUnitDeadline now lives in setup_deadline_typeaware_test.go; no names vanished.
- Ninety-five input/recipe wrappers and their coverage union are one family. Positive agreement, negative controls and built-in mutants coexist in this family. Production findings here rest on the positive reached agreement leaf, not failed mutant preconditions.
- Whole-package baseline exceeded 90 seconds during the first row's shared archive setup. The complete isolated Six family also exceeded 90 seconds. Narrowed clean baselines passed; no assertion-red baseline was audited. Cold Six setup also timed out once, then three warm isolated setup observations passed. Family median is null, not a guessed timing from leaf 000.
- The flags rows inspect Go authority constants and do not execute TypeScript port constants. Mutating their Go constants would mutate the authority. Their verdict is cannot-judge. Union 1<<27 and the 334017 mask were checked against the pinned Go source and successful tests.
- A proposed boolean acceptance mutant failed TypeScript narrowing, TS2367. It was discarded before the matrix and replaced with the independently frozen boolean-return constant change. An unconditional union control also failed go vet for unreachable code and was revised before use.
- A scratch guard edit initially read a cached native missing-binding mutant because the recipe declared only the original from/to anchor. Fresh ADAMIC_BUILD_CACHE_DIR=/tmp/u145/cache/W5 forced a native rebuild and caught the weakening in 66.162 seconds. The cached pass is not a survivor or witness verdict.
- The combined P2 probe timed out after the request row completed while the cost leaf was preparing products. The cost leaf was rerun alone and failed in 46.904 seconds. Its first timeout remains in the logs.
- The request validity check accepts an empty answer. Only refusal checks reject it. Vacuous subcases are request-valid, wrong-kind/mutant and unknown-question/mutant; the released-name subcase uses another entry and is excluded.
- The decoder checks root flags and malformed-frame refusal messages but does not detect strict/present boolean inversion. That limitation is demonstrated by M2 passing it and failing the Go agreement leaf.
- Exhaustive dynamic function reachability for the complete timed-out Six family was not established. The explicit decoder and request entries and their static functions are listed in code-under-test.md. No full-package or repo-wide uniqueness claim is made.
- All unexecuted results remain unknown. Bounded uniqueness and the two-mutant subsumption hint require central replay before broader conclusions. The subsumer has no complete-family median, so subsumer_seconds is null.

Timing and coverage:

Warm toolchain setup 0 seconds; nproc 5; mandatory npm ci 0.831 seconds. Pinned TypeScript Git checkout 13.328 seconds. Standalone native rebuilds: M1 0.904 s, M2 0.87 s, M3 0.869 s, M4 5.731 s. Compiler and checker archive builds 3.604 seconds. Matrix command wall times: M1 87.786 s, M2 29.167 s, M3 22.981 s, M4 45.532 s, P1 18.538 s. These include preparation; binary timing lines remain in logs. Session took about 30 minutes, including cold native setup, timeout narrowing and probes.

Skipped row: TestTypeAwareNativeBuildChild at top level, a subprocess helper; parents enable it during native builds. Other observed optional skips are listed verbatim in skips.json; compiler corpus and ADAMIC_TYPEAWARE_BENCH were enabled. Complete family timings, unexecuted family members under mutants, other package rows and repo-wide uniqueness were not covered. Production files and scratch switch are reverted; standalone diffs, replay scripts, logs and observed matrices are retained.
