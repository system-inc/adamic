Unit u052 started from origin/main 83f3940ec77b8ba779da05ded113e9d9e7e710b6, nproc 5.
All 59 named functions exist in the listed files and group into 15 rows; none moved or vanished.
Full baseline timed out at 90.022 seconds; the selected slice passed in 42.193 seconds with benchmark enabled and no skips.
Twelve production mutants support 3 bounded sacred rows, 10 subsumed rows and 2 separately proven witnesses.
One survivor changes behavior; eight selected entry probes were caught. Evidence is on test-audit/internal-native-radix.

```json
[
  {
    "test": "TestToStringWithARadixMatchesNode",
    "package": "internal/native",
    "file": "internal/native/radix_test.go",
    "seconds": 0.715,
    "oracle": "Live Node Number.toString, exact text for every input.",
    "oracle_kind": "external-run",
    "kills": [
      "M01",
      "M02",
      "M03",
      "M11"
    ],
    "unique_kills": [
      "M01"
    ],
    "last_proven_fail": "M11: radix_test.go:82: /tmp/adamic-gate/TestToStringWithARadixMatchesNode2351002337/001/harness: exit status 1",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestToStringWithARadixMatchesNode",
      "TestToStringWithARadixOutOfRangePanics",
      "TestRecordsAgainstNode",
      "TestRecordReadMutants",
      "TestRecordBenchmark",
      "TestRegExpSearchNode",
      "TestRegExpLintPatternsNode",
      "TestRegExpBytecodeTest262",
      "TestRegExpNativeStepLimit",
      "TestRegExpIteratorResultShape",
      "TestRegExpBytecodePatternUnits",
      "TestRuntimeReleasePaths",
      "TestRuntimeStringEquality",
      "TestRecordMutants family",
      "TestRegExpBytecodeRandomNode family"
    ],
    "evidence": "ADAMIC_MUTANT=M11 ADAMIC_RECORD_BENCH=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u052/cache/M11 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestToStringWithARadixMatchesNode|TestToStringWithARadixOutOfRangePanics|TestRecordsAgainstNode|TestRecordReadMutants|TestRecordBenchmark|TestRegExpSearchNode|TestRegExpLintPatternsNode|TestRegExpBytecodeTest262|TestRegExpNativeStepLimit|TestRegExpIteratorResultShape|TestRegExpBytecodePatternUnits|TestRuntimeReleasePaths|TestRuntimeStringEquality|TestRecordMutantsUnit00|TestRecordMutantsUnit01|TestRecordMutantsUnit02|TestRecordMutantsUnit03|TestRecordMutantsUnit04|TestRecordMutantsUnit05|TestRegExpBytecodeRandomNodeUnit00|TestRegExpBytecodeRandomNodeUnit01|TestRegExpBytecodeRandomNodeUnit02|TestRegExpBytecodeRandomNodeUnit03|TestRegExpBytecodeRandomNodeUnit04|TestRegExpBytecodeRandomNodeUnit05|TestRegExpBytecodeRandomNodeUnit06|TestRegExpBytecodeRandomNodeUnit07|TestRegExpBytecodeRandomNodeUnit08|TestRegExpBytecodeRandomNodeUnit09|TestRegExpBytecodeRandomNodeUnit10|TestRegExpBytecodeRandomNodeUnit11|TestRegExpBytecodeRandomNodeUnit12|TestRegExpBytecodeRandomNodeUnit13|TestRegExpBytecodeRandomNodeUnit14|TestRegExpBytecodeRandomNodeUnit15|TestRegExpBytecodeRandomNodeUnit16|TestRegExpBytecodeRandomNodeUnit17|TestRegExpBytecodeRandomNodeUnit18|TestRegExpBytecodeRandomNodeUnit19|TestRegExpBytecodeRandomNodeUnit20|TestRegExpBytecodeRandomNodeUnit21|TestRegExpBytecodeRandomNodeUnit22|TestRegExpBytecodeRandomNodeUnit23|TestRegExpBytecodeRandomNodeUnit24|TestRegExpBytecodeRandomNodeUnit25|TestRegExpBytecodeRandomNodeUnit26|TestRegExpBytecodeRandomNodeUnit27|TestRegExpBytecodeRandomNodeUnit28|TestRegExpBytecodeRandomNodeUnit29|TestRegExpBytecodeRandomNodeUnit30|TestRegExpBytecodeRandomNodeUnit31|TestRegExpBytecodeRandomNodeUnit32|TestRegExpBytecodeRandomNodeUnit33|TestRegExpBytecodeRandomNodeUnit34|TestRegExpBytecodeRandomNodeUnit35|TestRegExpBytecodeRandomNodeUnit36|TestRegExpBytecodeRandomNodeUnit37|TestRegExpBytecodeRandomNodeUnit38|TestRegExpBytecodeRandomNodeUnit39)$ => radix_test.go:82: /tmp/adamic-gate/TestToStringWithARadixMatchesNode2351002337/001/harness: exit status 1",
    "members": [
      "TestToStringWithARadixMatchesNode"
    ],
    "entry_probes": {
      "P01": "fail"
    }
  },
  {
    "test": "TestToStringWithARadixOutOfRangePanics",
    "package": "internal/native",
    "file": "internal/native/radix_test.go",
    "seconds": 0.561,
    "oracle": "Live Node output and exception classification; native exit70 and exact diagnostic prefix.",
    "oracle_kind": "external-run",
    "kills": [
      "M02",
      "M03"
    ],
    "unique_kills": [],
    "last_proven_fail": "M03: radix_test.go:145: (255).toString(36.5): native exit 70 \"before\\n\", Node exit 0 \"before\\n73\\n\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestToStringWithARadixMatchesNode"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 0.715,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestToStringWithARadixMatchesNode",
      "TestToStringWithARadixOutOfRangePanics",
      "TestRecordsAgainstNode",
      "TestRecordReadMutants",
      "TestRecordBenchmark",
      "TestRegExpSearchNode",
      "TestRegExpLintPatternsNode",
      "TestRegExpBytecodeTest262",
      "TestRegExpNativeStepLimit",
      "TestRegExpIteratorResultShape",
      "TestRegExpBytecodePatternUnits",
      "TestRuntimeReleasePaths",
      "TestRuntimeStringEquality",
      "TestRecordMutants family",
      "TestRegExpBytecodeRandomNode family"
    ],
    "evidence": "ADAMIC_MUTANT=M03 ADAMIC_RECORD_BENCH=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u052/cache/M03 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestToStringWithARadixMatchesNode|TestToStringWithARadixOutOfRangePanics|TestRecordsAgainstNode|TestRecordReadMutants|TestRecordBenchmark|TestRegExpSearchNode|TestRegExpLintPatternsNode|TestRegExpBytecodeTest262|TestRegExpNativeStepLimit|TestRegExpIteratorResultShape|TestRegExpBytecodePatternUnits|TestRuntimeReleasePaths|TestRuntimeStringEquality|TestRecordMutantsUnit00|TestRecordMutantsUnit01|TestRecordMutantsUnit02|TestRecordMutantsUnit03|TestRecordMutantsUnit04|TestRecordMutantsUnit05|TestRegExpBytecodeRandomNodeUnit00|TestRegExpBytecodeRandomNodeUnit01|TestRegExpBytecodeRandomNodeUnit02|TestRegExpBytecodeRandomNodeUnit03|TestRegExpBytecodeRandomNodeUnit04|TestRegExpBytecodeRandomNodeUnit05|TestRegExpBytecodeRandomNodeUnit06|TestRegExpBytecodeRandomNodeUnit07|TestRegExpBytecodeRandomNodeUnit08|TestRegExpBytecodeRandomNodeUnit09|TestRegExpBytecodeRandomNodeUnit10|TestRegExpBytecodeRandomNodeUnit11|TestRegExpBytecodeRandomNodeUnit12|TestRegExpBytecodeRandomNodeUnit13|TestRegExpBytecodeRandomNodeUnit14|TestRegExpBytecodeRandomNodeUnit15|TestRegExpBytecodeRandomNodeUnit16|TestRegExpBytecodeRandomNodeUnit17|TestRegExpBytecodeRandomNodeUnit18|TestRegExpBytecodeRandomNodeUnit19|TestRegExpBytecodeRandomNodeUnit20|TestRegExpBytecodeRandomNodeUnit21|TestRegExpBytecodeRandomNodeUnit22|TestRegExpBytecodeRandomNodeUnit23|TestRegExpBytecodeRandomNodeUnit24|TestRegExpBytecodeRandomNodeUnit25|TestRegExpBytecodeRandomNodeUnit26|TestRegExpBytecodeRandomNodeUnit27|TestRegExpBytecodeRandomNodeUnit28|TestRegExpBytecodeRandomNodeUnit29|TestRegExpBytecodeRandomNodeUnit30|TestRegExpBytecodeRandomNodeUnit31|TestRegExpBytecodeRandomNodeUnit32|TestRegExpBytecodeRandomNodeUnit33|TestRegExpBytecodeRandomNodeUnit34|TestRegExpBytecodeRandomNodeUnit35|TestRegExpBytecodeRandomNodeUnit36|TestRegExpBytecodeRandomNodeUnit37|TestRegExpBytecodeRandomNodeUnit38|TestRegExpBytecodeRandomNodeUnit39)$ => radix_test.go:145: (255).toString(36.5): native exit 70 \"before\\n\", Node exit 0 \"before\\n73\\n\"",
    "members": [
      "TestToStringWithARadixOutOfRangePanics"
    ],
    "entry_probes": {
      "P01": "fail"
    },
    "subsumption_mutants": 2
  },
  {
    "test": "TestRecordsAgainstNode",
    "package": "internal/native",
    "file": "internal/native/record_test.go",
    "seconds": 6.995,
    "oracle": "Live Node fixture stdout and prototype list; self runtime own-key stop policy and balanced allocation counts.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M04",
      "M05",
      "M06",
      "M10",
      "M11"
    ],
    "unique_kills": [
      "M04",
      "M05",
      "M06"
    ],
    "last_proven_fail": "M11: record_test.go:73: native: exit status 23",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P07",
      "P08"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestToStringWithARadixMatchesNode",
      "TestToStringWithARadixOutOfRangePanics",
      "TestRecordsAgainstNode",
      "TestRecordReadMutants",
      "TestRecordBenchmark",
      "TestRegExpSearchNode",
      "TestRegExpLintPatternsNode",
      "TestRegExpBytecodeTest262",
      "TestRegExpNativeStepLimit",
      "TestRegExpIteratorResultShape",
      "TestRegExpBytecodePatternUnits",
      "TestRuntimeReleasePaths",
      "TestRuntimeStringEquality",
      "TestRecordMutants family",
      "TestRegExpBytecodeRandomNode family"
    ],
    "evidence": "ADAMIC_MUTANT=M11 ADAMIC_RECORD_BENCH=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u052/cache/M11 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestToStringWithARadixMatchesNode|TestToStringWithARadixOutOfRangePanics|TestRecordsAgainstNode|TestRecordReadMutants|TestRecordBenchmark|TestRegExpSearchNode|TestRegExpLintPatternsNode|TestRegExpBytecodeTest262|TestRegExpNativeStepLimit|TestRegExpIteratorResultShape|TestRegExpBytecodePatternUnits|TestRuntimeReleasePaths|TestRuntimeStringEquality|TestRecordMutantsUnit00|TestRecordMutantsUnit01|TestRecordMutantsUnit02|TestRecordMutantsUnit03|TestRecordMutantsUnit04|TestRecordMutantsUnit05|TestRegExpBytecodeRandomNodeUnit00|TestRegExpBytecodeRandomNodeUnit01|TestRegExpBytecodeRandomNodeUnit02|TestRegExpBytecodeRandomNodeUnit03|TestRegExpBytecodeRandomNodeUnit04|TestRegExpBytecodeRandomNodeUnit05|TestRegExpBytecodeRandomNodeUnit06|TestRegExpBytecodeRandomNodeUnit07|TestRegExpBytecodeRandomNodeUnit08|TestRegExpBytecodeRandomNodeUnit09|TestRegExpBytecodeRandomNodeUnit10|TestRegExpBytecodeRandomNodeUnit11|TestRegExpBytecodeRandomNodeUnit12|TestRegExpBytecodeRandomNodeUnit13|TestRegExpBytecodeRandomNodeUnit14|TestRegExpBytecodeRandomNodeUnit15|TestRegExpBytecodeRandomNodeUnit16|TestRegExpBytecodeRandomNodeUnit17|TestRegExpBytecodeRandomNodeUnit18|TestRegExpBytecodeRandomNodeUnit19|TestRegExpBytecodeRandomNodeUnit20|TestRegExpBytecodeRandomNodeUnit21|TestRegExpBytecodeRandomNodeUnit22|TestRegExpBytecodeRandomNodeUnit23|TestRegExpBytecodeRandomNodeUnit24|TestRegExpBytecodeRandomNodeUnit25|TestRegExpBytecodeRandomNodeUnit26|TestRegExpBytecodeRandomNodeUnit27|TestRegExpBytecodeRandomNodeUnit28|TestRegExpBytecodeRandomNodeUnit29|TestRegExpBytecodeRandomNodeUnit30|TestRegExpBytecodeRandomNodeUnit31|TestRegExpBytecodeRandomNodeUnit32|TestRegExpBytecodeRandomNodeUnit33|TestRegExpBytecodeRandomNodeUnit34|TestRegExpBytecodeRandomNodeUnit35|TestRegExpBytecodeRandomNodeUnit36|TestRegExpBytecodeRandomNodeUnit37|TestRegExpBytecodeRandomNodeUnit38|TestRegExpBytecodeRandomNodeUnit39)$ => record_test.go:73: native: exit status 23",
    "members": [
      "TestRecordsAgainstNode"
    ],
    "entry_probes": {
      "P07": "fail",
      "P08": "fail"
    },
    "vacuous_subcases": {
      "P07": [
        "semantics",
        "references",
        "iteration",
        "numeric",
        "proto-assignment"
      ],
      "P08": [
        "reads",
        "proto-assignment"
      ]
    }
  },
  {
    "test": "TestRecordReadMutants",
    "package": "internal/native",
    "file": "internal/native/record_test.go",
    "seconds": 0.805,
    "oracle": "Witness of exact stop contract and own-hit comparison, plus sanitizer exclusion.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W01: record_test.go:274: exact stop contract did not catch mutant",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestToStringWithARadixMatchesNode",
      "TestToStringWithARadixOutOfRangePanics",
      "TestRecordsAgainstNode",
      "TestRecordReadMutants",
      "TestRecordBenchmark",
      "TestRegExpSearchNode",
      "TestRegExpLintPatternsNode",
      "TestRegExpBytecodeTest262",
      "TestRegExpNativeStepLimit",
      "TestRegExpIteratorResultShape",
      "TestRegExpBytecodePatternUnits",
      "TestRuntimeReleasePaths",
      "TestRuntimeStringEquality",
      "TestRecordMutants family",
      "TestRegExpBytecodeRandomNode family"
    ],
    "evidence": "timeout 120 go test -json -overlay /tmp/u052-witness-overlay.json -count=1 -timeout 90s ./internal/native/ -run ^(TestRecordReadMutants|TestRecordMutantsUnit.*)$ => record_test.go:274: exact stop contract did not catch mutant",
    "members": [
      "TestRecordReadMutants"
    ],
    "entry_probes": {}
  },
  {
    "test": "TestRecordBenchmark",
    "package": "internal/native",
    "file": "internal/native/record_test.go",
    "seconds": 16.934,
    "oracle": "Live Node work checksum; self five timing-field validity checks, no performance threshold.",
    "oracle_kind": "external-run",
    "kills": [
      "M10"
    ],
    "unique_kills": [],
    "last_proven_fail": "M10: record_test.go:321: benchmark side 0: signal: aborted",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestRuntimeStringEquality"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P07",
      "P08"
    ],
    "subsumer_seconds": 0.232,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestToStringWithARadixMatchesNode",
      "TestToStringWithARadixOutOfRangePanics",
      "TestRecordsAgainstNode",
      "TestRecordReadMutants",
      "TestRecordBenchmark",
      "TestRegExpSearchNode",
      "TestRegExpLintPatternsNode",
      "TestRegExpBytecodeTest262",
      "TestRegExpNativeStepLimit",
      "TestRegExpIteratorResultShape",
      "TestRegExpBytecodePatternUnits",
      "TestRuntimeReleasePaths",
      "TestRuntimeStringEquality",
      "TestRecordMutants family",
      "TestRegExpBytecodeRandomNode family"
    ],
    "evidence": "ADAMIC_MUTANT=M10 ADAMIC_RECORD_BENCH=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u052/cache/M10 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestToStringWithARadixMatchesNode|TestToStringWithARadixOutOfRangePanics|TestRecordsAgainstNode|TestRecordReadMutants|TestRecordBenchmark|TestRegExpSearchNode|TestRegExpLintPatternsNode|TestRegExpBytecodeTest262|TestRegExpNativeStepLimit|TestRegExpIteratorResultShape|TestRegExpBytecodePatternUnits|TestRuntimeReleasePaths|TestRuntimeStringEquality|TestRecordMutantsUnit00|TestRecordMutantsUnit01|TestRecordMutantsUnit02|TestRecordMutantsUnit03|TestRecordMutantsUnit04|TestRecordMutantsUnit05|TestRegExpBytecodeRandomNodeUnit00|TestRegExpBytecodeRandomNodeUnit01|TestRegExpBytecodeRandomNodeUnit02|TestRegExpBytecodeRandomNodeUnit03|TestRegExpBytecodeRandomNodeUnit04|TestRegExpBytecodeRandomNodeUnit05|TestRegExpBytecodeRandomNodeUnit06|TestRegExpBytecodeRandomNodeUnit07|TestRegExpBytecodeRandomNodeUnit08|TestRegExpBytecodeRandomNodeUnit09|TestRegExpBytecodeRandomNodeUnit10|TestRegExpBytecodeRandomNodeUnit11|TestRegExpBytecodeRandomNodeUnit12|TestRegExpBytecodeRandomNodeUnit13|TestRegExpBytecodeRandomNodeUnit14|TestRegExpBytecodeRandomNodeUnit15|TestRegExpBytecodeRandomNodeUnit16|TestRegExpBytecodeRandomNodeUnit17|TestRegExpBytecodeRandomNodeUnit18|TestRegExpBytecodeRandomNodeUnit19|TestRegExpBytecodeRandomNodeUnit20|TestRegExpBytecodeRandomNodeUnit21|TestRegExpBytecodeRandomNodeUnit22|TestRegExpBytecodeRandomNodeUnit23|TestRegExpBytecodeRandomNodeUnit24|TestRegExpBytecodeRandomNodeUnit25|TestRegExpBytecodeRandomNodeUnit26|TestRegExpBytecodeRandomNodeUnit27|TestRegExpBytecodeRandomNodeUnit28|TestRegExpBytecodeRandomNodeUnit29|TestRegExpBytecodeRandomNodeUnit30|TestRegExpBytecodeRandomNodeUnit31|TestRegExpBytecodeRandomNodeUnit32|TestRegExpBytecodeRandomNodeUnit33|TestRegExpBytecodeRandomNodeUnit34|TestRegExpBytecodeRandomNodeUnit35|TestRegExpBytecodeRandomNodeUnit36|TestRegExpBytecodeRandomNodeUnit37|TestRegExpBytecodeRandomNodeUnit38|TestRegExpBytecodeRandomNodeUnit39)$ => record_test.go:321: benchmark side 0: signal: aborted",
    "members": [
      "TestRecordBenchmark"
    ],
    "entry_probes": {
      "P07": "fail",
      "P08": "fail"
    },
    "subsumption_mutants": 1
  },
  {
    "test": "TestRegExpSearchNode",
    "package": "internal/native",
    "file": "internal/native/regexp_search_test.go",
    "seconds": 0.356,
    "oracle": "Live Node captures, groups, indices and lastIndex.",
    "oracle_kind": "external-run",
    "kills": [
      "M07",
      "M08",
      "M10",
      "M11"
    ],
    "unique_kills": [],
    "last_proven_fail": "M11: regexp_search_test.go:34: native regex oracle ([]): exit status 1",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestRegExpLintPatternsNode"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P02",
      "P03"
    ],
    "subsumer_seconds": 1.634,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestToStringWithARadixMatchesNode",
      "TestToStringWithARadixOutOfRangePanics",
      "TestRecordsAgainstNode",
      "TestRecordReadMutants",
      "TestRecordBenchmark",
      "TestRegExpSearchNode",
      "TestRegExpLintPatternsNode",
      "TestRegExpBytecodeTest262",
      "TestRegExpNativeStepLimit",
      "TestRegExpIteratorResultShape",
      "TestRegExpBytecodePatternUnits",
      "TestRuntimeReleasePaths",
      "TestRuntimeStringEquality",
      "TestRecordMutants family",
      "TestRegExpBytecodeRandomNode family"
    ],
    "evidence": "ADAMIC_MUTANT=M11 ADAMIC_RECORD_BENCH=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u052/cache/M11 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestToStringWithARadixMatchesNode|TestToStringWithARadixOutOfRangePanics|TestRecordsAgainstNode|TestRecordReadMutants|TestRecordBenchmark|TestRegExpSearchNode|TestRegExpLintPatternsNode|TestRegExpBytecodeTest262|TestRegExpNativeStepLimit|TestRegExpIteratorResultShape|TestRegExpBytecodePatternUnits|TestRuntimeReleasePaths|TestRuntimeStringEquality|TestRecordMutantsUnit00|TestRecordMutantsUnit01|TestRecordMutantsUnit02|TestRecordMutantsUnit03|TestRecordMutantsUnit04|TestRecordMutantsUnit05|TestRegExpBytecodeRandomNodeUnit00|TestRegExpBytecodeRandomNodeUnit01|TestRegExpBytecodeRandomNodeUnit02|TestRegExpBytecodeRandomNodeUnit03|TestRegExpBytecodeRandomNodeUnit04|TestRegExpBytecodeRandomNodeUnit05|TestRegExpBytecodeRandomNodeUnit06|TestRegExpBytecodeRandomNodeUnit07|TestRegExpBytecodeRandomNodeUnit08|TestRegExpBytecodeRandomNodeUnit09|TestRegExpBytecodeRandomNodeUnit10|TestRegExpBytecodeRandomNodeUnit11|TestRegExpBytecodeRandomNodeUnit12|TestRegExpBytecodeRandomNodeUnit13|TestRegExpBytecodeRandomNodeUnit14|TestRegExpBytecodeRandomNodeUnit15|TestRegExpBytecodeRandomNodeUnit16|TestRegExpBytecodeRandomNodeUnit17|TestRegExpBytecodeRandomNodeUnit18|TestRegExpBytecodeRandomNodeUnit19|TestRegExpBytecodeRandomNodeUnit20|TestRegExpBytecodeRandomNodeUnit21|TestRegExpBytecodeRandomNodeUnit22|TestRegExpBytecodeRandomNodeUnit23|TestRegExpBytecodeRandomNodeUnit24|TestRegExpBytecodeRandomNodeUnit25|TestRegExpBytecodeRandomNodeUnit26|TestRegExpBytecodeRandomNodeUnit27|TestRegExpBytecodeRandomNodeUnit28|TestRegExpBytecodeRandomNodeUnit29|TestRegExpBytecodeRandomNodeUnit30|TestRegExpBytecodeRandomNodeUnit31|TestRegExpBytecodeRandomNodeUnit32|TestRegExpBytecodeRandomNodeUnit33|TestRegExpBytecodeRandomNodeUnit34|TestRegExpBytecodeRandomNodeUnit35|TestRegExpBytecodeRandomNodeUnit36|TestRegExpBytecodeRandomNodeUnit37|TestRegExpBytecodeRandomNodeUnit38|TestRegExpBytecodeRandomNodeUnit39)$ => regexp_search_test.go:34: native regex oracle ([]): exit status 1",
    "members": [
      "TestRegExpSearchNode"
    ],
    "entry_probes": {
      "P02": "fail",
      "P03": "fail"
    },
    "subsumption_mutants": 4
  },
  {
    "test": "TestRegExpLintPatternsNode",
    "package": "internal/native",
    "file": "internal/native/regexp_search_test.go",
    "seconds": 1.634,
    "oracle": "Live Node captures, groups, indices and lastIndex on lint patterns.",
    "oracle_kind": "external-run",
    "kills": [
      "M07",
      "M08",
      "M10",
      "M11"
    ],
    "unique_kills": [],
    "last_proven_fail": "M11: regexp_search_test.go:76: native regex oracle ([]): exit status 1",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestRegExpSearchNode"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P02",
      "P03"
    ],
    "subsumer_seconds": 0.356,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestToStringWithARadixMatchesNode",
      "TestToStringWithARadixOutOfRangePanics",
      "TestRecordsAgainstNode",
      "TestRecordReadMutants",
      "TestRecordBenchmark",
      "TestRegExpSearchNode",
      "TestRegExpLintPatternsNode",
      "TestRegExpBytecodeTest262",
      "TestRegExpNativeStepLimit",
      "TestRegExpIteratorResultShape",
      "TestRegExpBytecodePatternUnits",
      "TestRuntimeReleasePaths",
      "TestRuntimeStringEquality",
      "TestRecordMutants family",
      "TestRegExpBytecodeRandomNode family"
    ],
    "evidence": "ADAMIC_MUTANT=M11 ADAMIC_RECORD_BENCH=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u052/cache/M11 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestToStringWithARadixMatchesNode|TestToStringWithARadixOutOfRangePanics|TestRecordsAgainstNode|TestRecordReadMutants|TestRecordBenchmark|TestRegExpSearchNode|TestRegExpLintPatternsNode|TestRegExpBytecodeTest262|TestRegExpNativeStepLimit|TestRegExpIteratorResultShape|TestRegExpBytecodePatternUnits|TestRuntimeReleasePaths|TestRuntimeStringEquality|TestRecordMutantsUnit00|TestRecordMutantsUnit01|TestRecordMutantsUnit02|TestRecordMutantsUnit03|TestRecordMutantsUnit04|TestRecordMutantsUnit05|TestRegExpBytecodeRandomNodeUnit00|TestRegExpBytecodeRandomNodeUnit01|TestRegExpBytecodeRandomNodeUnit02|TestRegExpBytecodeRandomNodeUnit03|TestRegExpBytecodeRandomNodeUnit04|TestRegExpBytecodeRandomNodeUnit05|TestRegExpBytecodeRandomNodeUnit06|TestRegExpBytecodeRandomNodeUnit07|TestRegExpBytecodeRandomNodeUnit08|TestRegExpBytecodeRandomNodeUnit09|TestRegExpBytecodeRandomNodeUnit10|TestRegExpBytecodeRandomNodeUnit11|TestRegExpBytecodeRandomNodeUnit12|TestRegExpBytecodeRandomNodeUnit13|TestRegExpBytecodeRandomNodeUnit14|TestRegExpBytecodeRandomNodeUnit15|TestRegExpBytecodeRandomNodeUnit16|TestRegExpBytecodeRandomNodeUnit17|TestRegExpBytecodeRandomNodeUnit18|TestRegExpBytecodeRandomNodeUnit19|TestRegExpBytecodeRandomNodeUnit20|TestRegExpBytecodeRandomNodeUnit21|TestRegExpBytecodeRandomNodeUnit22|TestRegExpBytecodeRandomNodeUnit23|TestRegExpBytecodeRandomNodeUnit24|TestRegExpBytecodeRandomNodeUnit25|TestRegExpBytecodeRandomNodeUnit26|TestRegExpBytecodeRandomNodeUnit27|TestRegExpBytecodeRandomNodeUnit28|TestRegExpBytecodeRandomNodeUnit29|TestRegExpBytecodeRandomNodeUnit30|TestRegExpBytecodeRandomNodeUnit31|TestRegExpBytecodeRandomNodeUnit32|TestRegExpBytecodeRandomNodeUnit33|TestRegExpBytecodeRandomNodeUnit34|TestRegExpBytecodeRandomNodeUnit35|TestRegExpBytecodeRandomNodeUnit36|TestRegExpBytecodeRandomNodeUnit37|TestRegExpBytecodeRandomNodeUnit38|TestRegExpBytecodeRandomNodeUnit39)$ => regexp_search_test.go:76: native regex oracle ([]): exit status 1",
    "members": [
      "TestRegExpLintPatternsNode"
    ],
    "entry_probes": {
      "P02": "fail",
      "P03": "fail"
    },
    "subsumption_mutants": 4
  },
  {
    "test": "TestRegExpBytecodeTest262",
    "package": "internal/native",
    "file": "internal/native/regexp_test.go",
    "seconds": 11.938,
    "oracle": "Recorded test262 RegExp observations in matches.json.gz; first digit-class result checked against live Node in authority-check.json.",
    "oracle_kind": "external-authority",
    "kills": [
      "M07",
      "M08",
      "M10",
      "M11"
    ],
    "unique_kills": [],
    "last_proven_fail": "M11: regexp_test.go:53: native regex oracle ([]): exit status 1",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestRegExpSearchNode"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P02",
      "P03"
    ],
    "subsumer_seconds": 0.356,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestToStringWithARadixMatchesNode",
      "TestToStringWithARadixOutOfRangePanics",
      "TestRecordsAgainstNode",
      "TestRecordReadMutants",
      "TestRecordBenchmark",
      "TestRegExpSearchNode",
      "TestRegExpLintPatternsNode",
      "TestRegExpBytecodeTest262",
      "TestRegExpNativeStepLimit",
      "TestRegExpIteratorResultShape",
      "TestRegExpBytecodePatternUnits",
      "TestRuntimeReleasePaths",
      "TestRuntimeStringEquality",
      "TestRecordMutants family",
      "TestRegExpBytecodeRandomNode family"
    ],
    "evidence": "ADAMIC_MUTANT=M11 ADAMIC_RECORD_BENCH=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u052/cache/M11 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestToStringWithARadixMatchesNode|TestToStringWithARadixOutOfRangePanics|TestRecordsAgainstNode|TestRecordReadMutants|TestRecordBenchmark|TestRegExpSearchNode|TestRegExpLintPatternsNode|TestRegExpBytecodeTest262|TestRegExpNativeStepLimit|TestRegExpIteratorResultShape|TestRegExpBytecodePatternUnits|TestRuntimeReleasePaths|TestRuntimeStringEquality|TestRecordMutantsUnit00|TestRecordMutantsUnit01|TestRecordMutantsUnit02|TestRecordMutantsUnit03|TestRecordMutantsUnit04|TestRecordMutantsUnit05|TestRegExpBytecodeRandomNodeUnit00|TestRegExpBytecodeRandomNodeUnit01|TestRegExpBytecodeRandomNodeUnit02|TestRegExpBytecodeRandomNodeUnit03|TestRegExpBytecodeRandomNodeUnit04|TestRegExpBytecodeRandomNodeUnit05|TestRegExpBytecodeRandomNodeUnit06|TestRegExpBytecodeRandomNodeUnit07|TestRegExpBytecodeRandomNodeUnit08|TestRegExpBytecodeRandomNodeUnit09|TestRegExpBytecodeRandomNodeUnit10|TestRegExpBytecodeRandomNodeUnit11|TestRegExpBytecodeRandomNodeUnit12|TestRegExpBytecodeRandomNodeUnit13|TestRegExpBytecodeRandomNodeUnit14|TestRegExpBytecodeRandomNodeUnit15|TestRegExpBytecodeRandomNodeUnit16|TestRegExpBytecodeRandomNodeUnit17|TestRegExpBytecodeRandomNodeUnit18|TestRegExpBytecodeRandomNodeUnit19|TestRegExpBytecodeRandomNodeUnit20|TestRegExpBytecodeRandomNodeUnit21|TestRegExpBytecodeRandomNodeUnit22|TestRegExpBytecodeRandomNodeUnit23|TestRegExpBytecodeRandomNodeUnit24|TestRegExpBytecodeRandomNodeUnit25|TestRegExpBytecodeRandomNodeUnit26|TestRegExpBytecodeRandomNodeUnit27|TestRegExpBytecodeRandomNodeUnit28|TestRegExpBytecodeRandomNodeUnit29|TestRegExpBytecodeRandomNodeUnit30|TestRegExpBytecodeRandomNodeUnit31|TestRegExpBytecodeRandomNodeUnit32|TestRegExpBytecodeRandomNodeUnit33|TestRegExpBytecodeRandomNodeUnit34|TestRegExpBytecodeRandomNodeUnit35|TestRegExpBytecodeRandomNodeUnit36|TestRegExpBytecodeRandomNodeUnit37|TestRegExpBytecodeRandomNodeUnit38|TestRegExpBytecodeRandomNodeUnit39)$ => regexp_test.go:53: native regex oracle ([]): exit status 1",
    "members": [
      "TestRegExpBytecodeTest262"
    ],
    "entry_probes": {
      "P02": "fail",
      "P03": "fail"
    },
    "subsumption_mutants": 4
  },
  {
    "test": "TestRegExpNativeStepLimit",
    "package": "internal/native",
    "file": "internal/native/regexp_test.go",
    "seconds": 0.145,
    "oracle": "Self exit70 and instruction-step-limit diagnostic; does not verify exact consumed instruction count, M12 survives.",
    "oracle_kind": "self",
    "kills": [
      "M08"
    ],
    "unique_kills": [],
    "last_proven_fail": "M08: regexp_test.go:273: native catastrophic backtracking: exit=<nil> output=",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestRegExpBytecodePatternUnits"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P02"
    ],
    "subsumer_seconds": 0.199,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestToStringWithARadixMatchesNode",
      "TestToStringWithARadixOutOfRangePanics",
      "TestRecordsAgainstNode",
      "TestRecordReadMutants",
      "TestRecordBenchmark",
      "TestRegExpSearchNode",
      "TestRegExpLintPatternsNode",
      "TestRegExpBytecodeTest262",
      "TestRegExpNativeStepLimit",
      "TestRegExpIteratorResultShape",
      "TestRegExpBytecodePatternUnits",
      "TestRuntimeReleasePaths",
      "TestRuntimeStringEquality",
      "TestRecordMutants family",
      "TestRegExpBytecodeRandomNode family"
    ],
    "evidence": "ADAMIC_MUTANT=M08 ADAMIC_RECORD_BENCH=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u052/cache/M08 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestToStringWithARadixMatchesNode|TestToStringWithARadixOutOfRangePanics|TestRecordsAgainstNode|TestRecordReadMutants|TestRecordBenchmark|TestRegExpSearchNode|TestRegExpLintPatternsNode|TestRegExpBytecodeTest262|TestRegExpNativeStepLimit|TestRegExpIteratorResultShape|TestRegExpBytecodePatternUnits|TestRuntimeReleasePaths|TestRuntimeStringEquality|TestRecordMutantsUnit00|TestRecordMutantsUnit01|TestRecordMutantsUnit02|TestRecordMutantsUnit03|TestRecordMutantsUnit04|TestRecordMutantsUnit05|TestRegExpBytecodeRandomNodeUnit00|TestRegExpBytecodeRandomNodeUnit01|TestRegExpBytecodeRandomNodeUnit02|TestRegExpBytecodeRandomNodeUnit03|TestRegExpBytecodeRandomNodeUnit04|TestRegExpBytecodeRandomNodeUnit05|TestRegExpBytecodeRandomNodeUnit06|TestRegExpBytecodeRandomNodeUnit07|TestRegExpBytecodeRandomNodeUnit08|TestRegExpBytecodeRandomNodeUnit09|TestRegExpBytecodeRandomNodeUnit10|TestRegExpBytecodeRandomNodeUnit11|TestRegExpBytecodeRandomNodeUnit12|TestRegExpBytecodeRandomNodeUnit13|TestRegExpBytecodeRandomNodeUnit14|TestRegExpBytecodeRandomNodeUnit15|TestRegExpBytecodeRandomNodeUnit16|TestRegExpBytecodeRandomNodeUnit17|TestRegExpBytecodeRandomNodeUnit18|TestRegExpBytecodeRandomNodeUnit19|TestRegExpBytecodeRandomNodeUnit20|TestRegExpBytecodeRandomNodeUnit21|TestRegExpBytecodeRandomNodeUnit22|TestRegExpBytecodeRandomNodeUnit23|TestRegExpBytecodeRandomNodeUnit24|TestRegExpBytecodeRandomNodeUnit25|TestRegExpBytecodeRandomNodeUnit26|TestRegExpBytecodeRandomNodeUnit27|TestRegExpBytecodeRandomNodeUnit28|TestRegExpBytecodeRandomNodeUnit29|TestRegExpBytecodeRandomNodeUnit30|TestRegExpBytecodeRandomNodeUnit31|TestRegExpBytecodeRandomNodeUnit32|TestRegExpBytecodeRandomNodeUnit33|TestRegExpBytecodeRandomNodeUnit34|TestRegExpBytecodeRandomNodeUnit35|TestRegExpBytecodeRandomNodeUnit36|TestRegExpBytecodeRandomNodeUnit37|TestRegExpBytecodeRandomNodeUnit38|TestRegExpBytecodeRandomNodeUnit39)$ => regexp_test.go:273: native catastrophic backtracking: exit=<nil> output=",
    "members": [
      "TestRegExpNativeStepLimit"
    ],
    "entry_probes": {
      "P02": "fail"
    },
    "subsumption_mutants": 1
  },
  {
    "test": "TestRegExpIteratorResultShape",
    "package": "internal/native",
    "file": "internal/native/regexp_test.go",
    "seconds": 0.151,
    "oracle": "Self optional done-field expectations; checks native process exit status.",
    "oracle_kind": "self",
    "kills": [
      "M09",
      "M11"
    ],
    "unique_kills": [
      "M09"
    ],
    "last_proven_fail": "M11: regexp_test.go:304: iterator result shape: exit status 1",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P04"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestToStringWithARadixMatchesNode",
      "TestToStringWithARadixOutOfRangePanics",
      "TestRecordsAgainstNode",
      "TestRecordReadMutants",
      "TestRecordBenchmark",
      "TestRegExpSearchNode",
      "TestRegExpLintPatternsNode",
      "TestRegExpBytecodeTest262",
      "TestRegExpNativeStepLimit",
      "TestRegExpIteratorResultShape",
      "TestRegExpBytecodePatternUnits",
      "TestRuntimeReleasePaths",
      "TestRuntimeStringEquality",
      "TestRecordMutants family",
      "TestRegExpBytecodeRandomNode family"
    ],
    "evidence": "ADAMIC_MUTANT=M11 ADAMIC_RECORD_BENCH=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u052/cache/M11 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestToStringWithARadixMatchesNode|TestToStringWithARadixOutOfRangePanics|TestRecordsAgainstNode|TestRecordReadMutants|TestRecordBenchmark|TestRegExpSearchNode|TestRegExpLintPatternsNode|TestRegExpBytecodeTest262|TestRegExpNativeStepLimit|TestRegExpIteratorResultShape|TestRegExpBytecodePatternUnits|TestRuntimeReleasePaths|TestRuntimeStringEquality|TestRecordMutantsUnit00|TestRecordMutantsUnit01|TestRecordMutantsUnit02|TestRecordMutantsUnit03|TestRecordMutantsUnit04|TestRecordMutantsUnit05|TestRegExpBytecodeRandomNodeUnit00|TestRegExpBytecodeRandomNodeUnit01|TestRegExpBytecodeRandomNodeUnit02|TestRegExpBytecodeRandomNodeUnit03|TestRegExpBytecodeRandomNodeUnit04|TestRegExpBytecodeRandomNodeUnit05|TestRegExpBytecodeRandomNodeUnit06|TestRegExpBytecodeRandomNodeUnit07|TestRegExpBytecodeRandomNodeUnit08|TestRegExpBytecodeRandomNodeUnit09|TestRegExpBytecodeRandomNodeUnit10|TestRegExpBytecodeRandomNodeUnit11|TestRegExpBytecodeRandomNodeUnit12|TestRegExpBytecodeRandomNodeUnit13|TestRegExpBytecodeRandomNodeUnit14|TestRegExpBytecodeRandomNodeUnit15|TestRegExpBytecodeRandomNodeUnit16|TestRegExpBytecodeRandomNodeUnit17|TestRegExpBytecodeRandomNodeUnit18|TestRegExpBytecodeRandomNodeUnit19|TestRegExpBytecodeRandomNodeUnit20|TestRegExpBytecodeRandomNodeUnit21|TestRegExpBytecodeRandomNodeUnit22|TestRegExpBytecodeRandomNodeUnit23|TestRegExpBytecodeRandomNodeUnit24|TestRegExpBytecodeRandomNodeUnit25|TestRegExpBytecodeRandomNodeUnit26|TestRegExpBytecodeRandomNodeUnit27|TestRegExpBytecodeRandomNodeUnit28|TestRegExpBytecodeRandomNodeUnit29|TestRegExpBytecodeRandomNodeUnit30|TestRegExpBytecodeRandomNodeUnit31|TestRegExpBytecodeRandomNodeUnit32|TestRegExpBytecodeRandomNodeUnit33|TestRegExpBytecodeRandomNodeUnit34|TestRegExpBytecodeRandomNodeUnit35|TestRegExpBytecodeRandomNodeUnit36|TestRegExpBytecodeRandomNodeUnit37|TestRegExpBytecodeRandomNodeUnit38|TestRegExpBytecodeRandomNodeUnit39)$ => regexp_test.go:304: iterator result shape: exit status 1",
    "members": [
      "TestRegExpIteratorResultShape"
    ],
    "entry_probes": {
      "P04": "fail"
    }
  },
  {
    "test": "TestRegExpBytecodePatternUnits",
    "package": "internal/native",
    "file": "internal/native/regexp_test.go",
    "seconds": 0.199,
    "oracle": "Self two lone-surrogate capture spans [0,1].",
    "oracle_kind": "self",
    "kills": [
      "M07",
      "M08",
      "M11"
    ],
    "unique_kills": [],
    "last_proven_fail": "M11: regexp_test.go:318: native regex oracle ([]): exit status 1",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestRegExpSearchNode"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P02",
      "P03"
    ],
    "subsumer_seconds": 0.356,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestToStringWithARadixMatchesNode",
      "TestToStringWithARadixOutOfRangePanics",
      "TestRecordsAgainstNode",
      "TestRecordReadMutants",
      "TestRecordBenchmark",
      "TestRegExpSearchNode",
      "TestRegExpLintPatternsNode",
      "TestRegExpBytecodeTest262",
      "TestRegExpNativeStepLimit",
      "TestRegExpIteratorResultShape",
      "TestRegExpBytecodePatternUnits",
      "TestRuntimeReleasePaths",
      "TestRuntimeStringEquality",
      "TestRecordMutants family",
      "TestRegExpBytecodeRandomNode family"
    ],
    "evidence": "ADAMIC_MUTANT=M11 ADAMIC_RECORD_BENCH=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u052/cache/M11 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestToStringWithARadixMatchesNode|TestToStringWithARadixOutOfRangePanics|TestRecordsAgainstNode|TestRecordReadMutants|TestRecordBenchmark|TestRegExpSearchNode|TestRegExpLintPatternsNode|TestRegExpBytecodeTest262|TestRegExpNativeStepLimit|TestRegExpIteratorResultShape|TestRegExpBytecodePatternUnits|TestRuntimeReleasePaths|TestRuntimeStringEquality|TestRecordMutantsUnit00|TestRecordMutantsUnit01|TestRecordMutantsUnit02|TestRecordMutantsUnit03|TestRecordMutantsUnit04|TestRecordMutantsUnit05|TestRegExpBytecodeRandomNodeUnit00|TestRegExpBytecodeRandomNodeUnit01|TestRegExpBytecodeRandomNodeUnit02|TestRegExpBytecodeRandomNodeUnit03|TestRegExpBytecodeRandomNodeUnit04|TestRegExpBytecodeRandomNodeUnit05|TestRegExpBytecodeRandomNodeUnit06|TestRegExpBytecodeRandomNodeUnit07|TestRegExpBytecodeRandomNodeUnit08|TestRegExpBytecodeRandomNodeUnit09|TestRegExpBytecodeRandomNodeUnit10|TestRegExpBytecodeRandomNodeUnit11|TestRegExpBytecodeRandomNodeUnit12|TestRegExpBytecodeRandomNodeUnit13|TestRegExpBytecodeRandomNodeUnit14|TestRegExpBytecodeRandomNodeUnit15|TestRegExpBytecodeRandomNodeUnit16|TestRegExpBytecodeRandomNodeUnit17|TestRegExpBytecodeRandomNodeUnit18|TestRegExpBytecodeRandomNodeUnit19|TestRegExpBytecodeRandomNodeUnit20|TestRegExpBytecodeRandomNodeUnit21|TestRegExpBytecodeRandomNodeUnit22|TestRegExpBytecodeRandomNodeUnit23|TestRegExpBytecodeRandomNodeUnit24|TestRegExpBytecodeRandomNodeUnit25|TestRegExpBytecodeRandomNodeUnit26|TestRegExpBytecodeRandomNodeUnit27|TestRegExpBytecodeRandomNodeUnit28|TestRegExpBytecodeRandomNodeUnit29|TestRegExpBytecodeRandomNodeUnit30|TestRegExpBytecodeRandomNodeUnit31|TestRegExpBytecodeRandomNodeUnit32|TestRegExpBytecodeRandomNodeUnit33|TestRegExpBytecodeRandomNodeUnit34|TestRegExpBytecodeRandomNodeUnit35|TestRegExpBytecodeRandomNodeUnit36|TestRegExpBytecodeRandomNodeUnit37|TestRegExpBytecodeRandomNodeUnit38|TestRegExpBytecodeRandomNodeUnit39)$ => regexp_test.go:318: native regex oracle ([]): exit status 1",
    "members": [
      "TestRegExpBytecodePatternUnits"
    ],
    "entry_probes": {
      "P02": "fail",
      "P03": "fail"
    },
    "subsumption_mutants": 3
  },
  {
    "test": "TestRuntimeReleasePaths",
    "package": "internal/native",
    "file": "internal/native/runtime_profile_test.go",
    "seconds": 0.27,
    "oracle": "Self live allocation counts 1 then0, fixed stdout; ASan and UBSan.",
    "oracle_kind": "self",
    "kills": [
      "M11"
    ],
    "unique_kills": [],
    "last_proven_fail": "M11: runtime_profile_test.go:51: /tmp/adamic-gate/TestRuntimeReleasePaths181448853/001/release: exit status 1",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestRegExpIteratorResultShape"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P06"
    ],
    "subsumer_seconds": 0.151,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestToStringWithARadixMatchesNode",
      "TestToStringWithARadixOutOfRangePanics",
      "TestRecordsAgainstNode",
      "TestRecordReadMutants",
      "TestRecordBenchmark",
      "TestRegExpSearchNode",
      "TestRegExpLintPatternsNode",
      "TestRegExpBytecodeTest262",
      "TestRegExpNativeStepLimit",
      "TestRegExpIteratorResultShape",
      "TestRegExpBytecodePatternUnits",
      "TestRuntimeReleasePaths",
      "TestRuntimeStringEquality",
      "TestRecordMutants family",
      "TestRegExpBytecodeRandomNode family"
    ],
    "evidence": "ADAMIC_MUTANT=M11 ADAMIC_RECORD_BENCH=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u052/cache/M11 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestToStringWithARadixMatchesNode|TestToStringWithARadixOutOfRangePanics|TestRecordsAgainstNode|TestRecordReadMutants|TestRecordBenchmark|TestRegExpSearchNode|TestRegExpLintPatternsNode|TestRegExpBytecodeTest262|TestRegExpNativeStepLimit|TestRegExpIteratorResultShape|TestRegExpBytecodePatternUnits|TestRuntimeReleasePaths|TestRuntimeStringEquality|TestRecordMutantsUnit00|TestRecordMutantsUnit01|TestRecordMutantsUnit02|TestRecordMutantsUnit03|TestRecordMutantsUnit04|TestRecordMutantsUnit05|TestRegExpBytecodeRandomNodeUnit00|TestRegExpBytecodeRandomNodeUnit01|TestRegExpBytecodeRandomNodeUnit02|TestRegExpBytecodeRandomNodeUnit03|TestRegExpBytecodeRandomNodeUnit04|TestRegExpBytecodeRandomNodeUnit05|TestRegExpBytecodeRandomNodeUnit06|TestRegExpBytecodeRandomNodeUnit07|TestRegExpBytecodeRandomNodeUnit08|TestRegExpBytecodeRandomNodeUnit09|TestRegExpBytecodeRandomNodeUnit10|TestRegExpBytecodeRandomNodeUnit11|TestRegExpBytecodeRandomNodeUnit12|TestRegExpBytecodeRandomNodeUnit13|TestRegExpBytecodeRandomNodeUnit14|TestRegExpBytecodeRandomNodeUnit15|TestRegExpBytecodeRandomNodeUnit16|TestRegExpBytecodeRandomNodeUnit17|TestRegExpBytecodeRandomNodeUnit18|TestRegExpBytecodeRandomNodeUnit19|TestRegExpBytecodeRandomNodeUnit20|TestRegExpBytecodeRandomNodeUnit21|TestRegExpBytecodeRandomNodeUnit22|TestRegExpBytecodeRandomNodeUnit23|TestRegExpBytecodeRandomNodeUnit24|TestRegExpBytecodeRandomNodeUnit25|TestRegExpBytecodeRandomNodeUnit26|TestRegExpBytecodeRandomNodeUnit27|TestRegExpBytecodeRandomNodeUnit28|TestRegExpBytecodeRandomNodeUnit29|TestRegExpBytecodeRandomNodeUnit30|TestRegExpBytecodeRandomNodeUnit31|TestRegExpBytecodeRandomNodeUnit32|TestRegExpBytecodeRandomNodeUnit33|TestRegExpBytecodeRandomNodeUnit34|TestRegExpBytecodeRandomNodeUnit35|TestRegExpBytecodeRandomNodeUnit36|TestRegExpBytecodeRandomNodeUnit37|TestRegExpBytecodeRandomNodeUnit38|TestRegExpBytecodeRandomNodeUnit39)$ => runtime_profile_test.go:51: /tmp/adamic-gate/TestRuntimeReleasePaths181448853/001/release: exit status 1",
    "members": [
      "TestRuntimeReleasePaths"
    ],
    "entry_probes": {
      "P06": "fail"
    },
    "subsumption_mutants": 1
  },
  {
    "test": "TestRuntimeStringEquality",
    "package": "internal/native",
    "file": "internal/native/runtime_profile_test.go",
    "seconds": 0.232,
    "oracle": "Live Node strict string equality including undefined, exact six booleans.",
    "oracle_kind": "external-run",
    "kills": [
      "M10",
      "M11"
    ],
    "unique_kills": [],
    "last_proven_fail": "M11: runtime_profile_test.go:80: /tmp/adamic-gate/TestRuntimeStringEquality3741305124/002/equal: exit status 1",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestRegExpSearchNode"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P05",
      "P06"
    ],
    "subsumer_seconds": 0.356,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestToStringWithARadixMatchesNode",
      "TestToStringWithARadixOutOfRangePanics",
      "TestRecordsAgainstNode",
      "TestRecordReadMutants",
      "TestRecordBenchmark",
      "TestRegExpSearchNode",
      "TestRegExpLintPatternsNode",
      "TestRegExpBytecodeTest262",
      "TestRegExpNativeStepLimit",
      "TestRegExpIteratorResultShape",
      "TestRegExpBytecodePatternUnits",
      "TestRuntimeReleasePaths",
      "TestRuntimeStringEquality",
      "TestRecordMutants family",
      "TestRegExpBytecodeRandomNode family"
    ],
    "evidence": "ADAMIC_MUTANT=M11 ADAMIC_RECORD_BENCH=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u052/cache/M11 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestToStringWithARadixMatchesNode|TestToStringWithARadixOutOfRangePanics|TestRecordsAgainstNode|TestRecordReadMutants|TestRecordBenchmark|TestRegExpSearchNode|TestRegExpLintPatternsNode|TestRegExpBytecodeTest262|TestRegExpNativeStepLimit|TestRegExpIteratorResultShape|TestRegExpBytecodePatternUnits|TestRuntimeReleasePaths|TestRuntimeStringEquality|TestRecordMutantsUnit00|TestRecordMutantsUnit01|TestRecordMutantsUnit02|TestRecordMutantsUnit03|TestRecordMutantsUnit04|TestRecordMutantsUnit05|TestRegExpBytecodeRandomNodeUnit00|TestRegExpBytecodeRandomNodeUnit01|TestRegExpBytecodeRandomNodeUnit02|TestRegExpBytecodeRandomNodeUnit03|TestRegExpBytecodeRandomNodeUnit04|TestRegExpBytecodeRandomNodeUnit05|TestRegExpBytecodeRandomNodeUnit06|TestRegExpBytecodeRandomNodeUnit07|TestRegExpBytecodeRandomNodeUnit08|TestRegExpBytecodeRandomNodeUnit09|TestRegExpBytecodeRandomNodeUnit10|TestRegExpBytecodeRandomNodeUnit11|TestRegExpBytecodeRandomNodeUnit12|TestRegExpBytecodeRandomNodeUnit13|TestRegExpBytecodeRandomNodeUnit14|TestRegExpBytecodeRandomNodeUnit15|TestRegExpBytecodeRandomNodeUnit16|TestRegExpBytecodeRandomNodeUnit17|TestRegExpBytecodeRandomNodeUnit18|TestRegExpBytecodeRandomNodeUnit19|TestRegExpBytecodeRandomNodeUnit20|TestRegExpBytecodeRandomNodeUnit21|TestRegExpBytecodeRandomNodeUnit22|TestRegExpBytecodeRandomNodeUnit23|TestRegExpBytecodeRandomNodeUnit24|TestRegExpBytecodeRandomNodeUnit25|TestRegExpBytecodeRandomNodeUnit26|TestRegExpBytecodeRandomNodeUnit27|TestRegExpBytecodeRandomNodeUnit28|TestRegExpBytecodeRandomNodeUnit29|TestRegExpBytecodeRandomNodeUnit30|TestRegExpBytecodeRandomNodeUnit31|TestRegExpBytecodeRandomNodeUnit32|TestRegExpBytecodeRandomNodeUnit33|TestRegExpBytecodeRandomNodeUnit34|TestRegExpBytecodeRandomNodeUnit35|TestRegExpBytecodeRandomNodeUnit36|TestRegExpBytecodeRandomNodeUnit37|TestRegExpBytecodeRandomNodeUnit38|TestRegExpBytecodeRandomNodeUnit39)$ => runtime_profile_test.go:80: /tmp/adamic-gate/TestRuntimeStringEquality3741305124/002/equal: exit status 1",
    "members": [
      "TestRuntimeStringEquality"
    ],
    "entry_probes": {
      "P05": "fail",
      "P06": "fail"
    },
    "subsumption_mutants": 2
  },
  {
    "test": "TestRecordMutants family",
    "package": "internal/native",
    "file": "internal/native/record_test.go",
    "seconds": 0.748,
    "oracle": "Witness of Node comparison and sanitizer detection.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W02: record_test.go:236: Node comparison did not catch mutant",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestToStringWithARadixMatchesNode",
      "TestToStringWithARadixOutOfRangePanics",
      "TestRecordsAgainstNode",
      "TestRecordReadMutants",
      "TestRecordBenchmark",
      "TestRegExpSearchNode",
      "TestRegExpLintPatternsNode",
      "TestRegExpBytecodeTest262",
      "TestRegExpNativeStepLimit",
      "TestRegExpIteratorResultShape",
      "TestRegExpBytecodePatternUnits",
      "TestRuntimeReleasePaths",
      "TestRuntimeStringEquality",
      "TestRecordMutants family",
      "TestRegExpBytecodeRandomNode family"
    ],
    "evidence": "timeout 120 go test -json -overlay /tmp/u052-witness-overlay.json -count=1 -timeout 90s ./internal/native/ -run ^(TestRecordReadMutants|TestRecordMutantsUnit.*)$ => record_test.go:236: Node comparison did not catch mutant",
    "members": [
      "TestRecordMutantsUnit00",
      "TestRecordMutantsUnit01",
      "TestRecordMutantsUnit02",
      "TestRecordMutantsUnit03",
      "TestRecordMutantsUnit04",
      "TestRecordMutantsUnit05"
    ],
    "entry_probes": {}
  },
  {
    "test": "TestRegExpBytecodeRandomNode family",
    "package": "internal/native",
    "file": "internal/native/regexp_test.go",
    "seconds": 10.594,
    "oracle": "Live Node captures, groups, indices and lastIndex; forty wrappers share runRegexCases.",
    "oracle_kind": "external-run",
    "kills": [
      "M07",
      "M08",
      "M10",
      "M11"
    ],
    "unique_kills": [],
    "last_proven_fail": "M11: regexp_test.go:388: native regex oracle ([]): exit status 1",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestRegExpSearchNode"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P02",
      "P03"
    ],
    "subsumer_seconds": 0.356,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestToStringWithARadixMatchesNode",
      "TestToStringWithARadixOutOfRangePanics",
      "TestRecordsAgainstNode",
      "TestRecordReadMutants",
      "TestRecordBenchmark",
      "TestRegExpSearchNode",
      "TestRegExpLintPatternsNode",
      "TestRegExpBytecodeTest262",
      "TestRegExpNativeStepLimit",
      "TestRegExpIteratorResultShape",
      "TestRegExpBytecodePatternUnits",
      "TestRuntimeReleasePaths",
      "TestRuntimeStringEquality",
      "TestRecordMutants family",
      "TestRegExpBytecodeRandomNode family"
    ],
    "evidence": "ADAMIC_MUTANT=M11 ADAMIC_RECORD_BENCH=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u052/cache/M11 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestToStringWithARadixMatchesNode|TestToStringWithARadixOutOfRangePanics|TestRecordsAgainstNode|TestRecordReadMutants|TestRecordBenchmark|TestRegExpSearchNode|TestRegExpLintPatternsNode|TestRegExpBytecodeTest262|TestRegExpNativeStepLimit|TestRegExpIteratorResultShape|TestRegExpBytecodePatternUnits|TestRuntimeReleasePaths|TestRuntimeStringEquality|TestRecordMutantsUnit00|TestRecordMutantsUnit01|TestRecordMutantsUnit02|TestRecordMutantsUnit03|TestRecordMutantsUnit04|TestRecordMutantsUnit05|TestRegExpBytecodeRandomNodeUnit00|TestRegExpBytecodeRandomNodeUnit01|TestRegExpBytecodeRandomNodeUnit02|TestRegExpBytecodeRandomNodeUnit03|TestRegExpBytecodeRandomNodeUnit04|TestRegExpBytecodeRandomNodeUnit05|TestRegExpBytecodeRandomNodeUnit06|TestRegExpBytecodeRandomNodeUnit07|TestRegExpBytecodeRandomNodeUnit08|TestRegExpBytecodeRandomNodeUnit09|TestRegExpBytecodeRandomNodeUnit10|TestRegExpBytecodeRandomNodeUnit11|TestRegExpBytecodeRandomNodeUnit12|TestRegExpBytecodeRandomNodeUnit13|TestRegExpBytecodeRandomNodeUnit14|TestRegExpBytecodeRandomNodeUnit15|TestRegExpBytecodeRandomNodeUnit16|TestRegExpBytecodeRandomNodeUnit17|TestRegExpBytecodeRandomNodeUnit18|TestRegExpBytecodeRandomNodeUnit19|TestRegExpBytecodeRandomNodeUnit20|TestRegExpBytecodeRandomNodeUnit21|TestRegExpBytecodeRandomNodeUnit22|TestRegExpBytecodeRandomNodeUnit23|TestRegExpBytecodeRandomNodeUnit24|TestRegExpBytecodeRandomNodeUnit25|TestRegExpBytecodeRandomNodeUnit26|TestRegExpBytecodeRandomNodeUnit27|TestRegExpBytecodeRandomNodeUnit28|TestRegExpBytecodeRandomNodeUnit29|TestRegExpBytecodeRandomNodeUnit30|TestRegExpBytecodeRandomNodeUnit31|TestRegExpBytecodeRandomNodeUnit32|TestRegExpBytecodeRandomNodeUnit33|TestRegExpBytecodeRandomNodeUnit34|TestRegExpBytecodeRandomNodeUnit35|TestRegExpBytecodeRandomNodeUnit36|TestRegExpBytecodeRandomNodeUnit37|TestRegExpBytecodeRandomNodeUnit38|TestRegExpBytecodeRandomNodeUnit39)$ => regexp_test.go:388: native regex oracle ([]): exit status 1",
    "members": [
      "TestRegExpBytecodeRandomNodeUnit00",
      "TestRegExpBytecodeRandomNodeUnit01",
      "TestRegExpBytecodeRandomNodeUnit02",
      "TestRegExpBytecodeRandomNodeUnit03",
      "TestRegExpBytecodeRandomNodeUnit04",
      "TestRegExpBytecodeRandomNodeUnit05",
      "TestRegExpBytecodeRandomNodeUnit06",
      "TestRegExpBytecodeRandomNodeUnit07",
      "TestRegExpBytecodeRandomNodeUnit08",
      "TestRegExpBytecodeRandomNodeUnit09",
      "TestRegExpBytecodeRandomNodeUnit10",
      "TestRegExpBytecodeRandomNodeUnit11",
      "TestRegExpBytecodeRandomNodeUnit12",
      "TestRegExpBytecodeRandomNodeUnit13",
      "TestRegExpBytecodeRandomNodeUnit14",
      "TestRegExpBytecodeRandomNodeUnit15",
      "TestRegExpBytecodeRandomNodeUnit16",
      "TestRegExpBytecodeRandomNodeUnit17",
      "TestRegExpBytecodeRandomNodeUnit18",
      "TestRegExpBytecodeRandomNodeUnit19",
      "TestRegExpBytecodeRandomNodeUnit20",
      "TestRegExpBytecodeRandomNodeUnit21",
      "TestRegExpBytecodeRandomNodeUnit22",
      "TestRegExpBytecodeRandomNodeUnit23",
      "TestRegExpBytecodeRandomNodeUnit24",
      "TestRegExpBytecodeRandomNodeUnit25",
      "TestRegExpBytecodeRandomNodeUnit26",
      "TestRegExpBytecodeRandomNodeUnit27",
      "TestRegExpBytecodeRandomNodeUnit28",
      "TestRegExpBytecodeRandomNodeUnit29",
      "TestRegExpBytecodeRandomNodeUnit30",
      "TestRegExpBytecodeRandomNodeUnit31",
      "TestRegExpBytecodeRandomNodeUnit32",
      "TestRegExpBytecodeRandomNodeUnit33",
      "TestRegExpBytecodeRandomNodeUnit34",
      "TestRegExpBytecodeRandomNodeUnit35",
      "TestRegExpBytecodeRandomNodeUnit36",
      "TestRegExpBytecodeRandomNodeUnit37",
      "TestRegExpBytecodeRandomNodeUnit38",
      "TestRegExpBytecodeRandomNodeUnit39"
    ],
    "entry_probes": {
      "P02": "fail",
      "P03": "fail"
    },
    "subsumption_mutants": 4
  }
]
```

| ID | Origin file:line | Change | Failed production rows |
|---|---|---|---|
| M01 | internal/native/runtime/radix.c:69 | 0.5 * (radix_next_double(value) - value) → 1.0 * (radix_next_double(value) - value) | TestToStringWithARadixMatchesNode |
| M02 | internal/native/runtime/radix.c:151 | radix_result("0", 1) → radix_result("1", 1) | TestToStringWithARadixMatchesNode, TestToStringWithARadixOutOfRangePanics |
| M03 | internal/native/runtime/radix.c:143 | radix > 36 → radix > 35 | TestToStringWithARadixMatchesNode, TestToStringWithARadixOutOfRangePanics |
| M04 | internal/native/runtime/record.c:111 | return true; }  size_t adamic_record_size → return false; }  size_t adamic_record_size | TestRecordsAgainstNode |
| M05 | internal/native/runtime/record.c:132 | number >= UINT32_MAX → number > UINT32_MAX | TestRecordsAgainstNode |
| M06 | internal/native/runtime/record.c:147 | return a < b ? -1 : a > b ? 1 : 0; → return a < b ? 1 : a > b ? -1 : 0; | TestRecordsAgainstNode |
| M07 | internal/native/runtime/regexp.c:9 | regex_step_limit = limit; → regex_step_limit = 1; | TestRegExpBytecodePatternUnits, TestRegExpBytecodeRandomNode family, TestRegExpBytecodeTest262, TestRegExpLintPatternsNode, TestRegExpSearchNode |
| M08 | internal/native/runtime/regexp.c:82 | i->ranges[first].first <= c → i->ranges[first].first < c | TestRegExpBytecodePatternUnits, TestRegExpBytecodeRandomNode family, TestRegExpBytecodeTest262, TestRegExpLintPatternsNode, TestRegExpNativeStepLimit, TestRegExpSearchNode |
| M09 | internal/native/runtime/regexp.c:581 | strcmp(object->shape->names[k], "done") → strcmp(object->shape->names[k], "value") | TestRegExpIteratorResultShape |
| M10 | internal/native/runtime/string_build_impl.h:127 | memcmp(left->bytes, right->bytes, left->length) == 0 → memcmp(left->bytes, right->bytes, left->length) != 0 | TestRecordBenchmark, TestRecordsAgainstNode, TestRegExpBytecodeRandomNode family, TestRegExpBytecodeTest262, TestRegExpLintPatternsNode, TestRegExpSearchNode, TestRuntimeStringEquality |
| M11 | internal/native/runtime/heap.c:375 | release_last(value); → (void)value; | TestRecordsAgainstNode, TestRegExpBytecodePatternUnits, TestRegExpBytecodeRandomNode family, TestRegExpBytecodeTest262, TestRegExpIteratorResultShape, TestRegExpLintPatternsNode, TestRegExpSearchNode, TestRuntimeReleasePaths, TestRuntimeStringEquality, TestToStringWithARadixMatchesNode |
| M12 | internal/native/runtime/regexp.c:187 | *steps >= regex_step_limit → *steps > regex_step_limit |  |

M12, regexp.c:187: >= changed to >. The existing slice passes. step-witness.log.gz proves a real change: limit1 with a two-instruction program exits70 in the baseline, but the mutant exits0 and prints "matched 1". Reproduce while switch.diff is applied with ADAMIC_BUILD_CACHE_DIR=/tmp/u052/cache/step-witness go run ./review/test-audit/internal-native-radix/step-witness. This is an unguarded exact-boundary behavior in the bounded slice, not an equivalent candidate.

The brief says 15 rows and lists 59 functions. Grouping six record wrappers and forty random-regexp wrappers reconciles these counts. The reference commit 8de93800f4 was older than origin/main; all reported mutation coordinates use the actual starting commit above.

The full-package timeout consumed 90 seconds. Its skips are recorded in baseline-skips.json. None of the selected rows skipped after ADAMIC_RECORD_BENCH=1. Unrelated WASI and split-TSGo opt-ins were not provisioned for this slice. The narrowed matrix runs exactly the 59 functions in scope.json, grouped into the 15 rows in every matrix_rows field. Kills outside this set are unknown. Sacred and subsumption labels are bounded to this matrix, not proven package-wide or repository-wide.

The warm workspace's cohere submodule/dependency tree was absent from the scratch worktree. Initial attempted matrix commands failed before tests ran. A misplaced dependency symlink and then path-sensitive cold Go compilation cost roughly three minutes. Those setup failures are not kills. The final matrix used the original checkout after clean timing finished, with a runtime selector; production files were restored before commit. The scratch worktree was left detached with the selector for local review.

The first C function inventory parser missed pointer-return definitions. runtime-function-inventory-complete.json corrects that parser and lists 658 runtime definitions as a conservative static superset. This is not dynamic C coverage and does not claim every listed function ran. reached-go-functions.txt lists seven Go functions with nonzero coverage; go.cover is the measured profile. The exact dynamic C call graph was not measured, so the requested exhaustive reached-function inventory remains a limitation. Mutants were fixed from the read production code before their outcomes were checked.

Native runtime tests call multiple public APIs. The brief's single-entry examples do not define how to combine their vacuity results. This audit reports eight selected public-entry probes, their individual outcomes in entry_probes, and vacuous_subcases for record modes that passed an empty-get or empty-keys probe. vacuous=false means the listed relevant probes were caught, not that every other public API received a probe. Other entry probes remain uncovered. Witness rows have vacuous=null and no production kills. Their weakened comparison/sanitizer runs are recorded separately in W01-W02.diff and witness.log.gz. Production-mutant failures of witness preconditions are retained in matrix.json but excluded from verdicts, uniqueness and subsumption.

The fixed-menu comparator mutant M06 changes the two comparator return constants. The switched build adds selector instrumentation, which is not itself a mutant. Each standalone production or probe diff applies to the starting sources and compiles with the runtime's C11 warning, count and sanitizer flags. Twenty standalone validation cases passed. The witness Go overlay also compiled and ran.

Subsumption is a hint based on twelve production mutants, with each row's actual caught-mutant count in subsumption_mutants. No deletion is recommended. The step-limit oracle checks exit70 and diagnostic, not exact steps; it misses M12. The benchmark compares work checksum and timing-field shape, not a performance threshold. Its M10 catch was a native abort, which also fails the semantic record row. test262 values are recorded outside-authority observations, not live test262 execution; one digit-class capture [0,10] and lastIndex0 were checked against Node during this session.

Warm setup skipped setup.sh. Fetch/toolchain check took 2.987 seconds; npm ci took 0.296 seconds. Full clean baseline took 90.022 binary seconds; slice baseline 42.193. The 45 separate clean timing runs totaled 157.795 binary seconds. Standalone C apply/compile validation took 4.484 seconds. Twelve matrix commands totaled 711.893 wall seconds, executed with concurrency two. Per-mutant binary and outer-command timings are in build-and-run-times.json. The first two commands exceeded 90 wall seconds including compilation, while every test binary finished below 90 seconds. Native clang rebuild times were not separately instrumented and remain part of binary elapsed times. The witness run took 26.594 binary seconds. Total session elapsed was about 24 minutes, including reporting and push.

Not covered: tests outside the slice, central repo-wide uniqueness, all public-entry empty probes, exact dynamic C reachability, separate native rebuild timing, and unrelated opt-in platform suites. No source change, test deletion, PR or main push is included.
