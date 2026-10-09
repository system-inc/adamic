u078: 95 listed tests present at base 60397548dd8a, grouped into 16 rows.
Bounded verdicts: 2 sacred, 1 subsumed, 5 witnesses, 3 setup checks, 5 cannot-judge.
Four valid production mutants were caught; all standalone diffs apply to the starting commit.
Both composition setup wrappers pass the empty-setup probe; complete printer-family timing is over budget.
Evidence branch: test-audit/stage1-cohere-css-composition_shards; production sources restored.

```json
[
  {
    "test": "TestCompositionMatchesGo setup family",
    "package": "stage1/cohere/css",
    "file": "stage1/cohere/css/composition_shards_test.go:147",
    "files": [
      "stage1/cohere/css/composition_shards_test.go:147",
      "stage1/cohere/css/css_test.go:420"
    ],
    "members": [
      "TestCompositionMatchesGo_Setup",
      "TestCompositionMatchesGo"
    ],
    "seconds": 5.954,
    "timing_samples": [
      5.536,
      5.954,
      6.151
    ],
    "oracle": "Constructed products must have nonempty identities; empty setup return bypasses this guard.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "SSetup: css_test.go:421: composition setup did not complete",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": [],
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestCompositionMatchesGo setup family"
    ],
    "evidence": "ADAMIC_MUTANT=SSetup; /tmp/u078-mutant=SSetup; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/css/ -run ^(TestCompositionMatchesGo_Setup|TestCompositionMatchesGo)$; css_test.go:421: composition setup did not complete; log=SSetup-TestCompositionMatchesGo_setup_family.log",
    "limitations": "Weakened guard/construction failed in this session.",
    "probe_evidence": [
      {
        "id": "PSetup",
        "log": "PSetup-TestCompositionMatchesGo_Setup.log",
        "command": [
          "timeout",
          "120",
          "go",
          "test",
          "-json",
          "-count=1",
          "-timeout",
          "90s",
          "./stage1/cohere/css/",
          "-run",
          "^(TestCompositionMatchesGo_Setup)$"
        ],
        "failing_output": null,
        "results": {
          "TestCompositionMatchesGo_Setup": "pass"
        }
      },
      {
        "id": "PSetup",
        "log": "PSetup-TestCompositionMatchesGo.log",
        "command": [
          "timeout",
          "120",
          "go",
          "test",
          "-json",
          "-count=1",
          "-timeout",
          "90s",
          "./stage1/cohere/css/",
          "-run",
          "^(TestCompositionMatchesGo)$"
        ],
        "failing_output": null,
        "results": {
          "TestCompositionMatchesGo": "pass"
        }
      }
    ]
  },
  {
    "test": "TestCompositionMatchesGo family",
    "package": "stage1/cohere/css",
    "file": "stage1/cohere/css/composition_shards_test.go:240",
    "files": [
      "stage1/cohere/css/composition_shards_test.go:240",
      "stage1/cohere/css/composition_shards_test.go:245",
      "stage1/cohere/css/composition_shards_test.go:250",
      "stage1/cohere/css/composition_shards_test.go:255",
      "stage1/cohere/css/composition_shards_test.go:260",
      "stage1/cohere/css/composition_shards_test.go:265",
      "stage1/cohere/css/composition_shards_test.go:270",
      "stage1/cohere/css/composition_shards_test.go:275",
      "stage1/cohere/css/composition_shards_test.go:280",
      "stage1/cohere/css/composition_shards_test.go:285",
      "stage1/cohere/css/composition_shards_test.go:290",
      "stage1/cohere/css/composition_shards_test.go:295",
      "stage1/cohere/css/composition_shards_test.go:300",
      "stage1/cohere/css/composition_shards_test.go:305",
      "stage1/cohere/css/composition_shards_test.go:310",
      "stage1/cohere/css/composition_shards_test.go:315"
    ],
    "members": [
      "TestCompositionMatchesGo_000",
      "TestCompositionMatchesGo_001",
      "TestCompositionMatchesGo_002",
      "TestCompositionMatchesGo_003",
      "TestCompositionMatchesGo_004",
      "TestCompositionMatchesGo_005",
      "TestCompositionMatchesGo_006",
      "TestCompositionMatchesGo_007",
      "TestCompositionMatchesGo_008",
      "TestCompositionMatchesGo_009",
      "TestCompositionMatchesGo_010",
      "TestCompositionMatchesGo_011",
      "TestCompositionMatchesGo_012",
      "TestCompositionMatchesGo_013",
      "TestCompositionMatchesGo_014",
      "TestCompositionMatchesGo_015"
    ],
    "seconds": 38.263,
    "timing_samples": [
      38.263,
      39.244,
      37.493
    ],
    "oracle": "Independent Go cohere composition trees and errors, compared byte for byte.",
    "oracle_kind": "external-run",
    "kills": [
      "M1",
      "M2",
      "M3"
    ],
    "unique_kills": [
      "M2",
      "M3"
    ],
    "last_proven_fail": "M3: composition_shards_test.go:138: native ASan/UBSan: line 202, byte 300: \"\\\":\\\"value-paren\\\",\\\"value\\\":\\\")\\\"},\\\"groups\\\":[\\\"\\\"],\\\"open\\\":{\\\"parenType\\\":\\\"\\\",\\\"raws\\\":{\\\"after\\\":\\\"\\\",\\\"before\\\":\\\"\\\"},\\\"source\\\":{\\\"end\\\":{\\\"column\\\":3,\\\"line\\\":1},\\\"endOffset\\\":11,\\\"start\\\":{\", Go cohere \"\\\":\\\"value-paren\\\",\\\"value\\\":\\\")\\\"},\\\"groups\\\":[\\\"foo.css\\\"],\\\"open\\\":{\\\"parenType\\\":\\\"\\\",\\\"raws\\\":{\\\"after\\\":\\\"\\\",\\\"before\\\":\\\"\\\"},\\\"source\\\":{\\\"end\\\":{\\\"column\\\":3,\\\"line\\\":1},\\\"endOffset\\\":11,\\\"s\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": [
      "M1",
      "M2",
      "M3"
    ],
    "probe_kills": [
      "PCompose"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestCompositionMatchesGo family",
      "TestCSSThroughput",
      "TestCSSPrinterOptimizedMatchesGo"
    ],
    "evidence": "ADAMIC_MUTANT=M3; /tmp/u078-mutant=M3; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/css/ -run ^(TestCompositionMatchesGo_000|TestCompositionMatchesGo_001|TestCompositionMatchesGo_002|TestCompositionMatchesGo_003|TestCompositionMatchesGo_004|TestCompositionMatchesGo_005|TestCompositionMatchesGo_006|TestCompositionMatchesGo_007|TestCompositionMatchesGo_008|TestCompositionMatchesGo_009|TestCompositionMatchesGo_010|TestCompositionMatchesGo_011|TestCompositionMatchesGo_012|TestCompositionMatchesGo_013|TestCompositionMatchesGo_014|TestCompositionMatchesGo_015)$; composition_shards_test.go:138: native ASan/UBSan: line 202, byte 300: \"\\\":\\\"value-paren\\\",\\\"value\\\":\\\")\\\"},\\\"groups\\\":[\\\"\\\"],\\\"open\\\":{\\\"parenType\\\":\\\"\\\",\\\"raws\\\":{\\\"after\\\":\\\"\\\",\\\"before\\\":\\\"\\\"},\\\"source\\\":{\\\"end\\\":{\\\"column\\\":3,\\\"line\\\":1},\\\"endOffset\\\":11,\\\"start\\\":{\", Go cohere \"\\\":\\\"value-paren\\\",\\\"value\\\":\\\")\\\"},\\\"groups\\\":[\\\"foo.css\\\"],\\\"open\\\":{\\\"parenType\\\":\\\"\\\",\\\"raws\\\":{\\\"after\\\":\\\"\\\",\\\"before\\\":\\\"\\\"},\\\"source\\\":{\\\"end\\\":{\\\"column\\\":3,\\\"line\\\":1},\\\"endOffset\\\":11,\\\"s\"; log=M3-TestCompositionMatchesGo_family.log",
    "limitations": "Unique only within the declared bounded matrix, not proven package unique.",
    "probe_evidence": [
      {
        "id": "PCompose",
        "log": "PCompose-TestCompositionMatchesGo_family.log",
        "command": [
          "timeout",
          "120",
          "go",
          "test",
          "-json",
          "-count=1",
          "-timeout",
          "90s",
          "./stage1/cohere/css/",
          "-run",
          "^(TestCompositionMatchesGo_000|TestCompositionMatchesGo_001|TestCompositionMatchesGo_002|TestCompositionMatchesGo_003|TestCompositionMatchesGo_004|TestCompositionMatchesGo_005|TestCompositionMatchesGo_006|TestCompositionMatchesGo_007|TestCompositionMatchesGo_008|TestCompositionMatchesGo_009|TestCompositionMatchesGo_010|TestCompositionMatchesGo_011|TestCompositionMatchesGo_012|TestCompositionMatchesGo_013|TestCompositionMatchesGo_014|TestCompositionMatchesGo_015)$"
        ],
        "failing_output": "composition_shards_test.go:138: native ASan/UBSan: line 1, byte 1: \"\", Go cohere \"case 0\"",
        "results": {
          "TestCompositionMatchesGo_008": "fail",
          "TestCompositionMatchesGo_009": "fail",
          "TestCompositionMatchesGo_010": "fail",
          "TestCompositionMatchesGo_011": "fail",
          "TestCompositionMatchesGo_000": "fail",
          "TestCompositionMatchesGo_007": "fail",
          "TestCompositionMatchesGo_004": "fail",
          "TestCompositionMatchesGo_005": "fail",
          "TestCompositionMatchesGo_014": "fail",
          "TestCompositionMatchesGo_006": "fail",
          "TestCompositionMatchesGo_015": "fail",
          "TestCompositionMatchesGo_002": "fail",
          "TestCompositionMatchesGo_003": "fail",
          "TestCompositionMatchesGo_013": "fail",
          "TestCompositionMatchesGo_001": "fail",
          "TestCompositionMatchesGo_012": "fail"
        }
      }
    ]
  },
  {
    "test": "TestCSSPrinterAgreesWithGo family",
    "package": "stage1/cohere/css",
    "file": "stage1/cohere/css/css_printer_parallel_test.go:146",
    "files": [
      "stage1/cohere/css/css_printer_parallel_test.go:146",
      "stage1/cohere/css/css_printer_parallel_test.go:147",
      "stage1/cohere/css/css_printer_parallel_test.go:148",
      "stage1/cohere/css/css_printer_parallel_test.go:149",
      "stage1/cohere/css/css_printer_parallel_test.go:150",
      "stage1/cohere/css/css_printer_parallel_test.go:151",
      "stage1/cohere/css/css_printer_parallel_test.go:152",
      "stage1/cohere/css/css_printer_parallel_test.go:153",
      "stage1/cohere/css/css_printer_parallel_test.go:154",
      "stage1/cohere/css/css_printer_parallel_test.go:155",
      "stage1/cohere/css/css_printer_parallel_test.go:156",
      "stage1/cohere/css/css_printer_parallel_test.go:157",
      "stage1/cohere/css/css_printer_parallel_test.go:158",
      "stage1/cohere/css/css_printer_parallel_test.go:159",
      "stage1/cohere/css/css_printer_parallel_test.go:160",
      "stage1/cohere/css/css_printer_parallel_test.go:161",
      "stage1/cohere/css/css_printer_parallel_test.go:162",
      "stage1/cohere/css/css_printer_parallel_test.go:163",
      "stage1/cohere/css/css_printer_parallel_test.go:164",
      "stage1/cohere/css/css_printer_parallel_test.go:165",
      "stage1/cohere/css/css_printer_parallel_test.go:166",
      "stage1/cohere/css/css_printer_parallel_test.go:167",
      "stage1/cohere/css/css_printer_parallel_test.go:168",
      "stage1/cohere/css/css_printer_parallel_test.go:169",
      "stage1/cohere/css/css_printer_parallel_test.go:170",
      "stage1/cohere/css/css_printer_parallel_test.go:171",
      "stage1/cohere/css/css_printer_parallel_test.go:172",
      "stage1/cohere/css/css_printer_parallel_test.go:173",
      "stage1/cohere/css/css_printer_parallel_test.go:174",
      "stage1/cohere/css/css_printer_parallel_test.go:175",
      "stage1/cohere/css/css_printer_parallel_test.go:176",
      "stage1/cohere/css/css_printer_parallel_test.go:177",
      "stage1/cohere/css/css_printer_parallel_test.go:178",
      "stage1/cohere/css/css_printer_parallel_test.go:179",
      "stage1/cohere/css/css_printer_parallel_test.go:180",
      "stage1/cohere/css/css_printer_parallel_test.go:181",
      "stage1/cohere/css/css_printer_parallel_test.go:182",
      "stage1/cohere/css/css_printer_parallel_test.go:183",
      "stage1/cohere/css/css_printer_parallel_test.go:184",
      "stage1/cohere/css/css_printer_parallel_test.go:185",
      "stage1/cohere/css/css_printer_parallel_test.go:186",
      "stage1/cohere/css/css_printer_parallel_test.go:187",
      "stage1/cohere/css/css_printer_parallel_test.go:188",
      "stage1/cohere/css/css_printer_parallel_test.go:189",
      "stage1/cohere/css/css_printer_parallel_test.go:190",
      "stage1/cohere/css/css_printer_parallel_test.go:191",
      "stage1/cohere/css/css_printer_parallel_test.go:192",
      "stage1/cohere/css/css_printer_parallel_test.go:193",
      "stage1/cohere/css/css_printer_parallel_test.go:194",
      "stage1/cohere/css/css_printer_parallel_test.go:195",
      "stage1/cohere/css/css_printer_parallel_test.go:196",
      "stage1/cohere/css/css_printer_parallel_test.go:197",
      "stage1/cohere/css/css_printer_parallel_test.go:198",
      "stage1/cohere/css/css_printer_parallel_test.go:199",
      "stage1/cohere/css/css_printer_parallel_test.go:200",
      "stage1/cohere/css/css_printer_parallel_test.go:201",
      "stage1/cohere/css/css_printer_parallel_test.go:202",
      "stage1/cohere/css/css_printer_parallel_test.go:203",
      "stage1/cohere/css/css_printer_parallel_test.go:204",
      "stage1/cohere/css/css_printer_parallel_test.go:205",
      "stage1/cohere/css/css_printer_parallel_test.go:206",
      "stage1/cohere/css/css_printer_parallel_test.go:207",
      "stage1/cohere/css/css_printer_parallel_test.go:208",
      "stage1/cohere/css/css_printer_parallel_test.go:209"
    ],
    "members": [
      "TestCSSPrinterAgreesWithGo_000",
      "TestCSSPrinterAgreesWithGo_001",
      "TestCSSPrinterAgreesWithGo_002",
      "TestCSSPrinterAgreesWithGo_003",
      "TestCSSPrinterAgreesWithGo_004",
      "TestCSSPrinterAgreesWithGo_005",
      "TestCSSPrinterAgreesWithGo_006",
      "TestCSSPrinterAgreesWithGo_007",
      "TestCSSPrinterAgreesWithGo_008",
      "TestCSSPrinterAgreesWithGo_009",
      "TestCSSPrinterAgreesWithGo_010",
      "TestCSSPrinterAgreesWithGo_011",
      "TestCSSPrinterAgreesWithGo_012",
      "TestCSSPrinterAgreesWithGo_013",
      "TestCSSPrinterAgreesWithGo_014",
      "TestCSSPrinterAgreesWithGo_015",
      "TestCSSPrinterAgreesWithGo_016",
      "TestCSSPrinterAgreesWithGo_017",
      "TestCSSPrinterAgreesWithGo_018",
      "TestCSSPrinterAgreesWithGo_019",
      "TestCSSPrinterAgreesWithGo_020",
      "TestCSSPrinterAgreesWithGo_021",
      "TestCSSPrinterAgreesWithGo_022",
      "TestCSSPrinterAgreesWithGo_023",
      "TestCSSPrinterAgreesWithGo_024",
      "TestCSSPrinterAgreesWithGo_025",
      "TestCSSPrinterAgreesWithGo_026",
      "TestCSSPrinterAgreesWithGo_027",
      "TestCSSPrinterAgreesWithGo_028",
      "TestCSSPrinterAgreesWithGo_029",
      "TestCSSPrinterAgreesWithGo_030",
      "TestCSSPrinterAgreesWithGo_031",
      "TestCSSPrinterAgreesWithGo_032",
      "TestCSSPrinterAgreesWithGo_033",
      "TestCSSPrinterAgreesWithGo_034",
      "TestCSSPrinterAgreesWithGo_035",
      "TestCSSPrinterAgreesWithGo_036",
      "TestCSSPrinterAgreesWithGo_037",
      "TestCSSPrinterAgreesWithGo_038",
      "TestCSSPrinterAgreesWithGo_039",
      "TestCSSPrinterAgreesWithGo_040",
      "TestCSSPrinterAgreesWithGo_041",
      "TestCSSPrinterAgreesWithGo_042",
      "TestCSSPrinterAgreesWithGo_043",
      "TestCSSPrinterAgreesWithGo_044",
      "TestCSSPrinterAgreesWithGo_045",
      "TestCSSPrinterAgreesWithGo_046",
      "TestCSSPrinterAgreesWithGo_047",
      "TestCSSPrinterAgreesWithGo_048",
      "TestCSSPrinterAgreesWithGo_049",
      "TestCSSPrinterAgreesWithGo_050",
      "TestCSSPrinterAgreesWithGo_051",
      "TestCSSPrinterAgreesWithGo_052",
      "TestCSSPrinterAgreesWithGo_053",
      "TestCSSPrinterAgreesWithGo_054",
      "TestCSSPrinterAgreesWithGo_055",
      "TestCSSPrinterAgreesWithGo_056",
      "TestCSSPrinterAgreesWithGo_057",
      "TestCSSPrinterAgreesWithGo_058",
      "TestCSSPrinterAgreesWithGo_059",
      "TestCSSPrinterAgreesWithGo_060",
      "TestCSSPrinterAgreesWithGo_061",
      "TestCSSPrinterAgreesWithGo_062",
      "TestCSSPrinterAgreesWithGo_063"
    ],
    "seconds": null,
    "timing_samples": [],
    "oracle": "Independent Go cohere printer. Agreement leaves also compare Prettier and its Go fork. Mixed agreement and planted-mutant leaves share one checker; witness verdict rests only on weakened comparison leaves.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: css_printer_parallel_test.go:209: narrow native ASan/UBSan printer mutant groups ignore the remaining width survived",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": [],
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCSSPrinterAgreesWithGo family"
    ],
    "evidence": "ADAMIC_MUTANT=W1; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/css/ -run ^TestCSSPrinterAgreesWithGo_063$; css_printer_parallel_test.go:209: narrow native ASan/UBSan printer mutant groups ignore the remaining width survived; log=bounded-printer-063-W1.log",
    "limitations": "Weakened guard/construction failed in this session. Full 64-member runtime exceeded 90 seconds twice. Witness proved on member 063 after a passing clean single-leaf run; agreement members and unfinished checks are not given production verdicts.",
    "probe_evidence": []
  },
  {
    "test": "TestCompositionMatchesGoUnion",
    "package": "stage1/cohere/css",
    "file": "stage1/cohere/css/composition_shards_test.go:152",
    "files": [
      "stage1/cohere/css/composition_shards_test.go:152"
    ],
    "members": [
      "TestCompositionMatchesGoUnion"
    ],
    "seconds": 22.92,
    "timing_samples": [
      23.294,
      21.873,
      22.92
    ],
    "oracle": "Node port mutant comparisons plus own synthetic shard coverage and planted firstDifference checks.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: composition_shards_test.go:189: planted failure caught by 0 shards, union 24594/24594",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": [],
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCompositionMatchesGoUnion"
    ],
    "evidence": "ADAMIC_MUTANT=W1; /tmp/u078-mutant=W1; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/css/ -run ^(TestCompositionMatchesGoUnion)$; composition_shards_test.go:189: planted failure caught by 0 shards, union 24594/24594; log=W1-TestCompositionMatchesGoUnion.log",
    "limitations": "Weakened guard/construction failed in this session.",
    "probe_evidence": []
  },
  {
    "test": "TestCSSPrinterAgreesWithGoUnion",
    "package": "stage1/cohere/css",
    "file": "stage1/cohere/css/css_printer_parallel_test.go:119",
    "files": [
      "stage1/cohere/css/css_printer_parallel_test.go:119"
    ],
    "members": [
      "TestCSSPrinterAgreesWithGoUnion"
    ],
    "seconds": 5.422,
    "timing_samples": [
      5.422,
      5.939,
      5.178
    ],
    "oracle": "Handwritten shard count, mode coverage and partition union assertions.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "SMode: css_printer_parallel_test.go:125: incorrect mode/check group in shard 8",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": [],
    "probe_kills": [
      "PMode"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestCSSPrinterAgreesWithGoUnion"
    ],
    "evidence": "ADAMIC_MUTANT=SMode; /tmp/u078-mutant=SMode; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/css/ -run ^(TestCSSPrinterAgreesWithGoUnion)$; css_printer_parallel_test.go:125: incorrect mode/check group in shard 8; log=SMode-TestCSSPrinterAgreesWithGoUnion.log",
    "limitations": "Weakened guard/construction failed in this session.",
    "probe_evidence": [
      {
        "id": "PMode",
        "log": "PMode-TestCSSPrinterAgreesWithGoUnion.log",
        "command": [
          "timeout",
          "120",
          "go",
          "test",
          "-json",
          "-count=1",
          "-timeout",
          "90s",
          "./stage1/cohere/css/",
          "-run",
          "^(TestCSSPrinterAgreesWithGoUnion)$"
        ],
        "failing_output": "css_printer_parallel_test.go:125: enumerated 0 shards, want 64",
        "results": {
          "TestCSSPrinterAgreesWithGoUnion": "fail"
        }
      }
    ]
  },
  {
    "test": "TestCSSThroughput",
    "package": "stage1/cohere/css",
    "file": "stage1/cohere/css/css_test.go:425",
    "files": [
      "stage1/cohere/css/css_test.go:425"
    ],
    "members": [
      "TestCSSThroughput"
    ],
    "seconds": 41.684,
    "timing_samples": [
      42.415,
      40.01,
      41.684
    ],
    "oracle": "PostCSS counts checked against native count/checksum, plus Node executes own source. Count agreement does not prove tree agreement.",
    "oracle_kind": "external-run",
    "kills": [
      "M1"
    ],
    "unique_kills": [],
    "last_proven_fail": "M1: css_test.go:465: PostCSS Node checksum \"54790 of 245940 stylesheets parsed, 129910 nodes\\n\" differs from \"245940 of 245940 stylesheets parsed, 0 nodes\\n\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestCompositionMatchesGo family"
    ],
    "mutants_in_matrix": [
      "M1"
    ],
    "probe_kills": [
      "PMain"
    ],
    "subsumer_seconds": 38.263,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestCompositionMatchesGo family",
      "TestCSSThroughput",
      "TestCSSPrinterOptimizedMatchesGo"
    ],
    "evidence": "ADAMIC_MUTANT=M1; /tmp/u078-mutant=M1; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/css/ -run ^(TestCSSThroughput)$; css_test.go:465: PostCSS Node checksum \"54790 of 245940 stylesheets parsed, 129910 nodes\\n\" differs from \"245940 of 245940 stylesheets parsed, 0 nodes\\n\"; log=M1-TestCSSThroughput.log",
    "limitations": "Verdict rests on 1 caught production mutants; it is not deletion advice. The second round rejects the empty answer after PostCSS establishes a nonempty checksum.",
    "vacuous_subcases": [
      "PMain native round 1 accepts empty output",
      "PMain Node source round 1 accepts empty output"
    ],
    "probe_evidence": [
      {
        "id": "PMain",
        "log": "PMain-TestCSSThroughput.log",
        "command": [
          "timeout",
          "120",
          "go",
          "test",
          "-json",
          "-count=1",
          "-timeout",
          "90s",
          "./stage1/cohere/css/",
          "-run",
          "^(TestCSSThroughput)$"
        ],
        "failing_output": "css_test.go:465: native checksum \"\" differs from \"54790 of 245940 stylesheets parsed, 129910 nodes\\n\"",
        "results": {
          "TestCSSThroughput": "fail"
        }
      }
    ]
  },
  {
    "test": "TestTheCanonicalRangeChecksCanFail",
    "package": "stage1/cohere/css",
    "file": "stage1/cohere/css/css_test.go:474",
    "files": [
      "stage1/cohere/css/css_test.go:474"
    ],
    "members": [
      "TestTheCanonicalRangeChecksCanFail"
    ],
    "seconds": 18.437,
    "timing_samples": [
      17.826,
      18.437,
      19.26
    ],
    "oracle": "Own planted IR range corruption must appear in rendered mismatch labels.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "WRange: css_test.go:486: raw native: exit 0, stdout \"missed\\n\", stderr \"\"",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": [],
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestTheCanonicalRangeChecksCanFail"
    ],
    "evidence": "ADAMIC_MUTANT=WRange; /tmp/u078-mutant=WRange; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/css/ -run ^(TestTheCanonicalRangeChecksCanFail)$; css_test.go:486: raw native: exit 0, stdout \"missed\\n\", stderr \"\"; log=WRange-TestTheCanonicalRangeChecksCanFail.log",
    "limitations": "Weakened guard/construction failed in this session.",
    "probe_evidence": []
  },
  {
    "test": "TestEachGapStandsWhereGapsMdSaysItDoes",
    "package": "stage1/cohere/css",
    "file": "stage1/cohere/css/gaps_test.go:13",
    "files": [
      "stage1/cohere/css/gaps_test.go:13"
    ],
    "members": [
      "TestEachGapStandsWhereGapsMdSaysItDoes"
    ],
    "seconds": 0.16,
    "timing_samples": [
      0.166,
      0.16,
      0.152
    ],
    "oracle": "Node execution is checked against handwritten stdout; Adamic refusal text is handwritten and has no independent authority.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": [],
    "probe_kills": [
      "PLower"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestEachGapStandsWhereGapsMdSaysItDoes"
    ],
    "evidence": "Clean timing logs and entry probes only; no admissible production kill.",
    "limitations": "No admissible production mutant of this compiler or dependency-port boundary was built within the four port-mutant menu.",
    "probe_evidence": [
      {
        "id": "PLower",
        "log": "PLower-TestEachGapStandsWhereGapsMdSaysItDoes.log",
        "command": [
          "timeout",
          "120",
          "go",
          "test",
          "-json",
          "-count=1",
          "-timeout",
          "90s",
          "./stage1/cohere/css/",
          "-run",
          "^(TestEachGapStandsWhereGapsMdSaysItDoes)$"
        ],
        "failing_output": "gaps_test.go:35: gap closed: update GAPS.md and remove the workaround or unblock native composition",
        "results": {
          "TestEachGapStandsWhereGapsMdSaysItDoes": "fail"
        }
      }
    ]
  },
  {
    "test": "TestClosedEmptyArrayUnionGap",
    "package": "stage1/cohere/css",
    "file": "stage1/cohere/css/gaps_test.go:48",
    "files": [
      "stage1/cohere/css/gaps_test.go:48"
    ],
    "members": [
      "TestClosedEmptyArrayUnionGap"
    ],
    "seconds": 0.364,
    "timing_samples": [
      0.315,
      0.364,
      0.374
    ],
    "oracle": "Node, native and own JavaScript backend must print handwritten 0; sanitizer leak report must be empty.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": [],
    "probe_kills": [
      "PLower"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestClosedEmptyArrayUnionGap"
    ],
    "evidence": "Clean timing logs and entry probes only; no admissible production kill.",
    "limitations": "No admissible production mutant of this compiler or dependency-port boundary was built within the four port-mutant menu.",
    "probe_evidence": [
      {
        "id": "PLower",
        "log": "PLower-TestClosedEmptyArrayUnionGap.log",
        "command": [
          "timeout",
          "120",
          "go",
          "test",
          "-json",
          "-count=1",
          "-timeout",
          "90s",
          "./stage1/cohere/css/",
          "-run",
          "^(TestClosedEmptyArrayUnionGap)$"
        ],
        "failing_output": null,
        "results": {
          "TestClosedEmptyArrayUnionGap": "fail"
        }
      }
    ]
  },
  {
    "test": "TestClosedParserRegexGap",
    "package": "stage1/cohere/css",
    "file": "stage1/cohere/css/gaps_test.go:68",
    "files": [
      "stage1/cohere/css/gaps_test.go:68"
    ],
    "members": [
      "TestClosedParserRegexGap"
    ],
    "seconds": 4.559,
    "timing_samples": [
      4.559,
      4.398,
      4.585
    ],
    "oracle": "Node, native and own JavaScript backend must print handwritten Parsed/Ok status labels. Status-only result is weaker than a tree comparison.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": [],
    "probe_kills": [
      "PLower"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestClosedParserRegexGap"
    ],
    "evidence": "Clean timing logs and entry probes only; no admissible production kill.",
    "limitations": "No admissible production mutant of this compiler or dependency-port boundary was built within the four port-mutant menu.",
    "probe_evidence": [
      {
        "id": "PLower",
        "log": "PLower-TestClosedParserRegexGap.log",
        "command": [
          "timeout",
          "120",
          "go",
          "test",
          "-json",
          "-count=1",
          "-timeout",
          "90s",
          "./stage1/cohere/css/",
          "-run",
          "^(TestClosedParserRegexGap)$"
        ],
        "failing_output": null,
        "results": {
          "TestClosedParserRegexGap": "fail"
        }
      }
    ]
  },
  {
    "test": "TestClosedOptionalBooleanConditionGap",
    "package": "stage1/cohere/css",
    "file": "stage1/cohere/css/gaps_test.go:88",
    "files": [
      "stage1/cohere/css/gaps_test.go:88"
    ],
    "members": [
      "TestClosedOptionalBooleanConditionGap"
    ],
    "seconds": 0.363,
    "timing_samples": [
      0.331,
      0.379,
      0.363
    ],
    "oracle": "Node, native and own JavaScript backend must print handwritten important; sanitizer leak report must be empty.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": [],
    "probe_kills": [
      "PLower"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestClosedOptionalBooleanConditionGap"
    ],
    "evidence": "Clean timing logs and entry probes only; no admissible production kill.",
    "limitations": "No admissible production mutant of this compiler or dependency-port boundary was built within the four port-mutant menu.",
    "probe_evidence": [
      {
        "id": "PLower",
        "log": "PLower-TestClosedOptionalBooleanConditionGap.log",
        "command": [
          "timeout",
          "120",
          "go",
          "test",
          "-json",
          "-count=1",
          "-timeout",
          "90s",
          "./stage1/cohere/css/",
          "-run",
          "^(TestClosedOptionalBooleanConditionGap)$"
        ],
        "failing_output": null,
        "results": {
          "TestClosedOptionalBooleanConditionGap": "fail"
        }
      }
    ]
  },
  {
    "test": "TestComposedMemoryChecksCanFail",
    "package": "stage1/cohere/css",
    "file": "stage1/cohere/css/memory_checks_test.go:17",
    "files": [
      "stage1/cohere/css/memory_checks_test.go:17"
    ],
    "members": [
      "TestComposedMemoryChecksCanFail"
    ],
    "seconds": 64.836,
    "timing_samples": [
      64.836,
      67.634,
      64.328
    ],
    "oracle": "Go printer bytes must survive ordinary runs; own planted C defects must trigger named ASan/UBSan/LSan reports.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "WMemory: memory_checks_test.go:83: memory mutant survived: 0",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": [],
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestComposedMemoryChecksCanFail"
    ],
    "evidence": "ADAMIC_MUTANT=WMemory; /tmp/u078-mutant=WMemory; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/css/ -run ^(TestComposedMemoryChecksCanFail)$; memory_checks_test.go:83: memory mutant survived: 0; log=WMemory-TestComposedMemoryChecksCanFail.log",
    "limitations": "Weakened guard/construction failed in this session.",
    "probe_evidence": []
  },
  {
    "test": "TestCSSPrinterOptimizedMatchesGo",
    "package": "stage1/cohere/css",
    "file": "stage1/cohere/css/optimized_test.go:14",
    "files": [
      "stage1/cohere/css/optimized_test.go:14"
    ],
    "members": [
      "TestCSSPrinterOptimizedMatchesGo"
    ],
    "seconds": 43.842,
    "timing_samples": [
      42.065,
      43.842,
      44.079
    ],
    "oracle": "Independent Go cohere printer bytes, default and narrow modes, compared with optimized native artifact.",
    "oracle_kind": "external-run",
    "kills": [
      "M1",
      "M4"
    ],
    "unique_kills": [
      "M4"
    ],
    "last_proven_fail": "M4: optimized_test.go:29: line 2, byte 10: \"\\\"rules:\\\\nselectors,\\\\ndeclarations,\\\\nempty rules,\\\\nsemicolons;\\\\n\\\"\", Go cohere \"\\\"rules:\\\\n  selectors,\\\\n  declarations,\\\\n  empty rules,\\\\n  semicolons;\\\\n\\\"\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": [
      "M1",
      "M2",
      "M4"
    ],
    "probe_kills": [
      "PPrint"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestCompositionMatchesGo family",
      "TestCSSThroughput",
      "TestCSSPrinterOptimizedMatchesGo"
    ],
    "evidence": "ADAMIC_MUTANT=M4; /tmp/u078-mutant=M4; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/css/ -run ^(TestCSSPrinterOptimizedMatchesGo)$; optimized_test.go:29: line 2, byte 10: \"\\\"rules:\\\\nselectors,\\\\ndeclarations,\\\\nempty rules,\\\\nsemicolons;\\\\n\\\"\", Go cohere \"\\\"rules:\\\\n  selectors,\\\\n  declarations,\\\\n  empty rules,\\\\n  semicolons;\\\\n\\\"\"; log=M4-TestCSSPrinterOptimizedMatchesGo.log",
    "limitations": "Unique only within the declared bounded matrix, not proven package unique.",
    "probe_evidence": [
      {
        "id": "PPrint",
        "log": "PPrint-TestCSSPrinterOptimizedMatchesGo.log",
        "command": [
          "timeout",
          "120",
          "go",
          "test",
          "-json",
          "-count=1",
          "-timeout",
          "90s",
          "./stage1/cohere/css/",
          "-run",
          "^(TestCSSPrinterOptimizedMatchesGo)$"
        ],
        "failing_output": "optimized_test.go:29: line 1, byte 1: \"\", Go cohere \"case 0\"",
        "results": {
          "TestCSSPrinterOptimizedMatchesGo": "fail"
        }
      }
    ]
  },
  {
    "test": "TestCSSParserOptimizedMatchesNode",
    "package": "stage1/cohere/css",
    "file": "stage1/cohere/css/optimized_test.go:36",
    "files": [
      "stage1/cohere/css/optimized_test.go:36"
    ],
    "members": [
      "TestCSSParserOptimizedMatchesNode"
    ],
    "seconds": 25.034,
    "timing_samples": [
      26.35,
      25.034,
      24.58
    ],
    "oracle": "Node executes the same port source. Port-source mutations change this oracle too, so no production kill is counted here.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": [],
    "probe_kills": [
      "PC"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestCSSParserOptimizedMatchesNode"
    ],
    "evidence": "Clean timing logs and entry probes only; no admissible production kill.",
    "limitations": "No admissible production mutant of this compiler or dependency-port boundary was built within the four port-mutant menu.",
    "probe_evidence": [
      {
        "id": "PC",
        "log": "PC-TestCSSParserOptimizedMatchesNode.log",
        "command": [
          "timeout",
          "120",
          "go",
          "test",
          "-json",
          "-count=1",
          "-timeout",
          "90s",
          "./stage1/cohere/css/",
          "-run",
          "^(TestCSSParserOptimizedMatchesNode)$"
        ],
        "failing_output": "optimized_test.go:42: native: clang failed: exit status 1",
        "results": {
          "TestCSSParserOptimizedMatchesNode": "fail"
        }
      }
    ]
  },
  {
    "test": "TestThePortParsesAsGoCohereDoesUnion",
    "package": "stage1/cohere/css",
    "file": "stage1/cohere/css/parser_shards_test.go:417",
    "files": [
      "stage1/cohere/css/parser_shards_test.go:417"
    ],
    "members": [
      "TestThePortParsesAsGoCohereDoesUnion"
    ],
    "seconds": 2.755,
    "timing_samples": [
      2.755,
      2.406,
      2.77
    ],
    "oracle": "Own shard assignments, fixture coverage and variant identities.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "SParser: parser_shards_test.go:425: union count 73782, live enumeration 98376",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": [],
    "probe_kills": [
      "PParserPlan"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestThePortParsesAsGoCohereDoesUnion"
    ],
    "evidence": "ADAMIC_MUTANT=SParser; /tmp/u078-mutant=SParser; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/css/ -run ^(TestThePortParsesAsGoCohereDoesUnion)$; parser_shards_test.go:425: union count 73782, live enumeration 98376; log=SParser-TestThePortParsesAsGoCohereDoesUnion.log",
    "limitations": "Weakened guard/construction failed in this session.",
    "probe_evidence": [
      {
        "id": "PParserPlan",
        "log": "PParserPlan-TestThePortParsesAsGoCohereDoesUnion.log",
        "command": [
          "timeout",
          "120",
          "go",
          "test",
          "-json",
          "-count=1",
          "-timeout",
          "90s",
          "./stage1/cohere/css/",
          "-run",
          "^(TestThePortParsesAsGoCohereDoesUnion)$"
        ],
        "failing_output": "parser_shards_test.go:423: shard count 0",
        "results": {
          "TestThePortParsesAsGoCohereDoesUnion": "fail"
        }
      }
    ]
  },
  {
    "test": "TestCSSParserPlantedDisagreement",
    "package": "stage1/cohere/css",
    "file": "stage1/cohere/css/parser_shards_test.go:480",
    "files": [
      "stage1/cohere/css/parser_shards_test.go:480"
    ],
    "members": [
      "TestCSSParserPlantedDisagreement"
    ],
    "seconds": 7.021,
    "timing_samples": [
      7.414,
      6.499,
      7.021
    ],
    "oracle": "Subprocess should report exactly the planted shard failure; its agreement guard must detect altered bytes.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: parser_shards_test.go:507: expected only TestThePortParsesAsGoCohereDoes_120 to catch pinned/cohere@7945d102a6c18dd36adf9114a758ce646e8b2359/0:C: exit 0",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": [],
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCSSParserPlantedDisagreement"
    ],
    "evidence": "ADAMIC_MUTANT=W1; /tmp/u078-mutant=W1; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/css/ -run ^(TestCSSParserPlantedDisagreement)$; parser_shards_test.go:507: expected only TestThePortParsesAsGoCohereDoes_120 to catch pinned/cohere@7945d102a6c18dd36adf9114a758ce646e8b2359/0:C: exit 0; log=W1-TestCSSParserPlantedDisagreement.log",
    "limitations": "Weakened guard/construction failed in this session.",
    "probe_evidence": []
  }
]
```

| ID | Origin file:line | Change | Failed rows in bounded matrix |
|---|---|---|---|
| M1 | stage1/cohere/css/parser.ts:703 | drop parser.parse invocation; constructor root still returned | TestCompositionMatchesGo family, TestCSSThroughput, TestCSSPrinterOptimizedMatchesGo |
| M2 | stage1/cohere/css/compose.ts:252 | clear composed import marker | TestCompositionMatchesGo family |
| M3 | stage1/cohere/css/tree.ts:146 | render literal tree nodes as empty strings | TestCompositionMatchesGo family |
| M4 | stage1/cohere/css/print_doc.ts:132 | drop document indentation increment | TestCSSPrinterOptimizedMatchesGo |

Survivors: none in the declared production matrix. M2 survives the optimized printer row but is caught by composition. Kills outside the declared rows remain unknown.

Special checks, separate from production kills: W1 disables firstDifference at css_test.go:277; WMemory disables native.go:103 sanitizer flags; WRange drops mismatch reports at nodes.ts:102 and tree.ts:179. SSetup erases the constructed oracle path at composition_shards_test.go:94; SMode changes count/8 to count/4 at printer_shards_test.go:293; SParser starts at variant 0 instead of -1 at parser_shards_test.go:77. Each fails its declared witness or construction row; checks.json contains commands, failed members and unknown members. None counts toward sacred/subsumed.

Brief problems, costs and limits:

- The brief says 14 rows but lists 95 Test functions. The shared-checker and distinct-assertion rules produce 16 rows here. row-members.json lists every member; no supplied name moved or vanished.
- The supplied historical file commit 8de93800f4 is not current origin/main. All locations and diffs use the fetched base 60397548dd8a9494a7d2aaa874b7648b8e625607.
- Whole-package clean testing cooked at the 90-second binary limit. No observed test assertion failed before that timeout. Production mutation testing therefore uses source-call bounds and runs each reached row separately.
- The 64-member printer family cooked on both cold baseline and warm timing retry. Its full median is null, not an invented lower-bound median. A single clean member 063 followed by its weakened-check failure proves the bounded witness.
- There are 26 requested printer leaves without a completed clean result. They are explicitly unknown in baseline-coverage.json. The family witness verdict does not claim production quality for those leaves or its agreement members.
- The 48 printer mutant leaves share a checker with the agreement leaves. Their witness guard is conditional on ownership of a pinned witness; some other leaves check termination only. Whole-family witness status does not prove each leaf can detect disagreement.
- The broad W1 printer run cooked after real assertion failures; unfinished members stay unknown in checks.json. The single-leaf replay removes reliance on the timeout as evidence.
- TestCSSParserOptimizedMatchesNode executes the same port source on Node as its expected oracle. A port-source change would alter both implementation and oracle, so no such production kill is counted. PC probes only native C emission and leaves Node source intact.
- The open-gap and three closed-gap rows check compiler or dependency-port snippets rather than the selected CSS port behaviors. No admissible production compiler/dependency mutant was built for them within the four port-mutant budget; their verdict is cannot-judge. PLower is only an empty-answer probe.
- The maximum-four rebuild guidance and the aim of three mutants per row conflict for this mixed slice. Four production mutants cover parser, composition, canonical rendering and document printing. No verdict treats empty probes or guard edits as production mutants.
- The function inventory is a conservative source declaration inventory, not dynamic TypeScript coverage. Exact transitive runtime reach was not proved. Entry/import reasoning defines the declared matrix bounds.
- Throughput compares aggregate counts/checksum, not full trees. Its native-first expected value allows empty native and Node answers to pass round 1 under PMain; round 2 fails after PostCSS provides the nonempty checksum.
- Both setup wrappers return successfully when compositionSetup returns its zero value at entry. Breaking construction instead makes their normal guard fail. They are setup-check rows with vacuous=true, not production-unique tests.
- The first SSetup edit left oracle unused and failed go vet. It was replaced by oracle[:0], preserving variable use. The first WRange edit used constant false and lost TypeScript field narrowing; it was replaced by dropping the reporting statements. Invalid attempts are preserved and support no verdict.
- Automatic approval review initially rejected a resume over a possible leftover source mutation. Read-only verification showed equal HEAD/origin/main hashes and no production diff; the subsequent safer resume was approved. No authorization remains blocked.
- Timing medians are the package binary elapsed line from three separate count=1 runs. Some singleton rows ran concurrently with one other process, so their costs include contention on 5 CPUs. The printer family has no three successful full timings.
- Warm tooling did not include CSS fixture or npm dependencies. API npm ci, pinned Prettier fixtures, PostCSS and Prettier installation were performed before baseline; throughput opt-in and both library comparison variables were enabled. No completed requested row skipped.
- The WRange weakened run fails first on the raw native check; it does not independently prove that the later composed branch catches its disabled report. Its standalone diff compiles both modules, but only the observed raw failure supports the witness verdict.
- No repo-wide replay, package uniqueness outside the declared bounds, full untimed CSS run, or extra compiler mutant campaign was performed. Central replay can use every standalone diff. No main push or PR was made.
- The suggested 30-minute port budget was exceeded. Cold/warm 90-second family attempts, repeated native products, three-run timing requirements and special-edit validation consumed the extra time. All timed-out runs were stopped by their binary deadline rather than extended.

Setup, building and running: {
  "base": "60397548dd8a9494a7d2aaa874b7648b8e625607",
  "nproc": 5,
  "setup": "warm env.sh, cloud/setup.sh skipped",
  "dependency_install": {
    "npm_ci_api_reported_seconds": 0.421,
    "fixture_clone_wall": 1.951,
    "postcss_install_wall": 0.715,
    "prettier_install_wall": 0.693
  },
  "clean_controller_wall_sum": 489.9258430469981,
  "timing_controller_wall_sum": 890.9040042909983,
  "audit_and_followup_wall_sum": 744.3220777019988,
  "native_mutant_build_wall": {
    "M1": 14.449862155999654,
    "M2": 14.665385534999587,
    "M3": 14.959863443999893,
    "M4": 25.56399897799929
  },
  "compiler_build_wall": 5.576399978999689,
  "conditions": "Go test -count=1 per invocation. Setup/composition timing families sequential; remaining singleton timing rows used two concurrent processes on 5 CPUs. Durations are each test binary package line, not command wall time. The third composition timing log completed while its controller was paused; its binary elapsed is valid, controller wall unavailable. Successful build validations were resumed rather than repeated after two invalid special edits. Total wall sums count the failed validation attempts.",
  "overall_elapsed": "Approximately 45 minutes before publishing; exceeded the suggested 30-minute port budget."
}
