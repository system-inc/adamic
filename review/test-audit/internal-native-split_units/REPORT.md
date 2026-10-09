Unit u053 at 0942c5169d0ea736d9dfaa19881af1ad8adad162. All ten requested names exist in their listed files.
The clean full package timed out at 90.053s; the ten-row clean slice passed in 5.964s. No assigned row skipped.
Four production mutants: two bounded sacred rows and four subsumed rows. Four construction rows are setup-checks.
Twenty-five empty-entry probes are separate; length and unit-count probes cooked on the JavaScript row and remain unknown there.
Evidence is on test-audit/internal-native-split_units under review/test-audit/internal-native-split_units/.

```json
[
  {
    "test": "TestSplitUnitCoverage",
    "package": "internal/native",
    "file": "internal/native/split_units_test.go",
    "seconds": 0.011,
    "oracle": "Frozen range counts and AST assertions over real wrapper names, targets, helpers and indices. Self-written construction contract.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S1:     split_units_test.go:65: invalid coverage at {first:1 last:2049} after 0",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P23"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestSplitUnitCoverage",
      "TestShardSelection",
      "TestRetainedSplitCoverage",
      "TestRetainedTopLevelCoverage",
      "TestStringBuildingMatchesNode",
      "TestStringIndexMatchesNode",
      "TestStringViewAfterAppendMatchesNode",
      "TestStringIndexCacheStatesMatchNode",
      "TestStringsMatchJavaScript",
      "TestRuntimeStringViews"
    ],
    "evidence": "ADAMIC_MUTANT=S1 ADAMIC_BUILD_CACHE_DIR=/tmp/u053/cache/S1 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestSplitUnitCoverage|TestShardSelection|TestRetainedSplitCoverage|TestRetainedTopLevelCoverage|TestStringBuildingMatchesNode|TestStringIndexMatchesNode|TestStringViewAfterAppendMatchesNode|TestStringIndexCacheStatesMatchNode|TestStringsMatchJavaScript|TestRuntimeStringViews)$ > logs/S1.log 2>&1;     split_units_test.go:65: invalid coverage at {first:1 last:2049} after 0",
    "construction_kills": [
      "S1"
    ],
    "entry_probe_results": {
      "unitRanges": "fail"
    },
    "subsumption_mutants": null
  },
  {
    "test": "TestShardSelection",
    "package": "internal/native",
    "file": "internal/native/split_units_test.go",
    "seconds": 0.007,
    "oracle": "Self-written rejected-shard list and default (0,1). No valid explicit shard checked. S2 rejects valid 0/1 but this row passes.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "P24:     split_units_test.go:349: invalid shard accepted: \"-1/2\"",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P24"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestSplitUnitCoverage",
      "TestShardSelection",
      "TestRetainedSplitCoverage",
      "TestRetainedTopLevelCoverage",
      "TestStringBuildingMatchesNode",
      "TestStringIndexMatchesNode",
      "TestStringViewAfterAppendMatchesNode",
      "TestStringIndexCacheStatesMatchNode",
      "TestStringsMatchJavaScript",
      "TestRuntimeStringViews"
    ],
    "evidence": "ADAMIC_MUTANT=P24 timeout 120 go test -overlay /tmp/u053/P24-overlay.json -json -count=1 -timeout 90s ./internal/native/ -run ^(TestSplitUnitCoverage|TestShardSelection|TestRetainedSplitCoverage|TestRetainedTopLevelCoverage|TestStringBuildingMatchesNode|TestStringIndexMatchesNode|TestStringViewAfterAppendMatchesNode|TestStringIndexCacheStatesMatchNode|TestStringsMatchJavaScript|TestRuntimeStringViews)$ > logs/P24.log 2>&1;     split_units_test.go:349: invalid shard accepted: \"-1/2\"",
    "construction_kills": [],
    "entry_probe_results": {
      "parseTestShard": "fail"
    },
    "subsumption_mutants": null
  },
  {
    "test": "TestRetainedSplitCoverage",
    "package": "internal/native",
    "file": "internal/native/split_units_test.go",
    "seconds": 0.014,
    "oracle": "Self-written partition, one-owner-per-piece, frozen WASI digest/count, target/cache lists and record-mutant membership.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S3:     split_units_test.go:404: piece 0 of 90 has 0 owners across 1 shards",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P23",
      "P24",
      "P25"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestSplitUnitCoverage",
      "TestShardSelection",
      "TestRetainedSplitCoverage",
      "TestRetainedTopLevelCoverage",
      "TestStringBuildingMatchesNode",
      "TestStringIndexMatchesNode",
      "TestStringViewAfterAppendMatchesNode",
      "TestStringIndexCacheStatesMatchNode",
      "TestStringsMatchJavaScript",
      "TestRuntimeStringViews"
    ],
    "evidence": "ADAMIC_MUTANT=S3 ADAMIC_BUILD_CACHE_DIR=/tmp/u053/cache/S3 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestSplitUnitCoverage|TestShardSelection|TestRetainedSplitCoverage|TestRetainedTopLevelCoverage|TestStringBuildingMatchesNode|TestStringIndexMatchesNode|TestStringViewAfterAppendMatchesNode|TestStringIndexCacheStatesMatchNode|TestStringsMatchJavaScript|TestRuntimeStringViews)$ > logs/S3.log 2>&1;     split_units_test.go:404: piece 0 of 90 has 0 owners across 1 shards",
    "construction_kills": [
      "S2",
      "S3"
    ],
    "entry_probe_results": {
      "unitRanges": "fail",
      "parseTestShard": "fail",
      "testShard.owns": "fail"
    },
    "subsumption_mutants": null
  },
  {
    "test": "TestRetainedTopLevelCoverage",
    "package": "internal/native",
    "file": "internal/native/split_units_test.go",
    "seconds": 0.01,
    "oracle": "Self-written AST wrapper-name, helper and unit-index assertions.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S4:     split_units_test.go:477: wrong unit in TestRecordMutantsUnit00",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestSplitUnitCoverage",
      "TestShardSelection",
      "TestRetainedSplitCoverage",
      "TestRetainedTopLevelCoverage",
      "TestStringBuildingMatchesNode",
      "TestStringIndexMatchesNode",
      "TestStringViewAfterAppendMatchesNode",
      "TestStringIndexCacheStatesMatchNode",
      "TestStringsMatchJavaScript",
      "TestRuntimeStringViews"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestSplitUnitCoverage|TestShardSelection|TestRetainedSplitCoverage|TestRetainedTopLevelCoverage|TestStringBuildingMatchesNode|TestStringIndexMatchesNode|TestStringViewAfterAppendMatchesNode|TestStringIndexCacheStatesMatchNode|TestStringsMatchJavaScript|TestRuntimeStringViews)$ > logs/S4.log 2>&1;     split_units_test.go:477: wrong unit in TestRecordMutantsUnit00",
    "construction_kills": [
      "S4"
    ],
    "entry_probe_results": {},
    "subsumption_mutants": null
  },
  {
    "test": "TestStringBuildingMatchesNode",
    "package": "internal/native",
    "file": "internal/native/string_build_test.go",
    "seconds": 0.226,
    "oracle": "Executes Node and compares all 2000 stateful operation outputs; WTF-8 serialization is handwritten. M2 fails on ASan before comparison.",
    "oracle_kind": "external-run",
    "kills": [
      "M2",
      "M3"
    ],
    "unique_kills": [],
    "last_proven_fail": "M3:     string_build_test.go:116: operation 196: native  127758; Node  127757",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestStringViewAfterAppendMatchesNode"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P2",
      "P5",
      "P6",
      "P9",
      "P14",
      "P15",
      "P17",
      "P19",
      "P22"
    ],
    "subsumer_seconds": 0.249,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestSplitUnitCoverage",
      "TestShardSelection",
      "TestRetainedSplitCoverage",
      "TestRetainedTopLevelCoverage",
      "TestStringBuildingMatchesNode",
      "TestStringIndexMatchesNode",
      "TestStringViewAfterAppendMatchesNode",
      "TestStringIndexCacheStatesMatchNode",
      "TestStringsMatchJavaScript",
      "TestRuntimeStringViews"
    ],
    "evidence": "ADAMIC_MUTANT=M3 ADAMIC_BUILD_CACHE_DIR=/tmp/u053/cache/M3 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestSplitUnitCoverage|TestShardSelection|TestRetainedSplitCoverage|TestRetainedTopLevelCoverage|TestStringBuildingMatchesNode|TestStringIndexMatchesNode|TestStringViewAfterAppendMatchesNode|TestStringIndexCacheStatesMatchNode|TestStringsMatchJavaScript|TestRuntimeStringViews)$ > logs/M3.log 2>&1;     string_build_test.go:116: operation 196: native  127758; Node  127757",
    "construction_kills": [],
    "entry_probe_results": {
      "adamic_string_append": "fail",
      "adamic_string_char_code_at": "fail",
      "adamic_string_code_point_at": "fail",
      "adamic_string_concat": "fail",
      "adamic_string_index_of_from": "fail",
      "adamic_string_length": "fail",
      "adamic_string_repeat": "fail",
      "adamic_string_slice": "fail",
      "adamic_string_units": "fail"
    },
    "subsumption_mutants": 2
  },
  {
    "test": "TestStringIndexMatchesNode",
    "package": "internal/native",
    "file": "internal/native/string_index_test.go",
    "seconds": 5.431,
    "oracle": "Executes Node and compares complete output lines for four patterns, lengths, shifts and read orders; WTF-8 serialization is handwritten.",
    "oracle_kind": "external-run",
    "kills": [
      "M3"
    ],
    "unique_kills": [],
    "last_proven_fail": "M3:     string_index_test.go:278: line 1:",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestStringIndexCacheStatesMatchNode"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P3",
      "P5",
      "P6",
      "P9",
      "P13",
      "P15",
      "P19"
    ],
    "subsumer_seconds": 0.197,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestSplitUnitCoverage",
      "TestShardSelection",
      "TestRetainedSplitCoverage",
      "TestRetainedTopLevelCoverage",
      "TestStringBuildingMatchesNode",
      "TestStringIndexMatchesNode",
      "TestStringViewAfterAppendMatchesNode",
      "TestStringIndexCacheStatesMatchNode",
      "TestStringsMatchJavaScript",
      "TestRuntimeStringViews"
    ],
    "evidence": "ADAMIC_MUTANT=M3 ADAMIC_BUILD_CACHE_DIR=/tmp/u053/cache/M3 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestSplitUnitCoverage|TestShardSelection|TestRetainedSplitCoverage|TestRetainedTopLevelCoverage|TestStringBuildingMatchesNode|TestStringIndexMatchesNode|TestStringViewAfterAppendMatchesNode|TestStringIndexCacheStatesMatchNode|TestStringsMatchJavaScript|TestRuntimeStringViews)$ > logs/M3.log 2>&1;     string_index_test.go:278: line 1:",
    "construction_kills": [],
    "entry_probe_results": {
      "adamic_string_at": "fail",
      "adamic_string_char_code_at": "fail",
      "adamic_string_code_point_at": "fail",
      "adamic_string_concat": "fail",
      "adamic_string_index_of": "fail",
      "adamic_string_length": "fail",
      "adamic_string_slice": "fail"
    },
    "subsumption_mutants": 1
  },
  {
    "test": "TestStringViewAfterAppendMatchesNode",
    "package": "internal/native",
    "file": "internal/native/string_index_test.go",
    "seconds": 0.249,
    "oracle": "Executes Node and compares complete UTF-16 view output; an in-place pointer fact is a handwritten expected 1. M2 fails on ASan before comparison.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M2",
      "M3"
    ],
    "unique_kills": [],
    "last_proven_fail": "M3:     string_index_test.go:404: UTF-16 view after append differs from Node",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestStringBuildingMatchesNode"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1",
      "P2",
      "P3",
      "P5",
      "P6",
      "P15",
      "P19",
      "P22"
    ],
    "subsumer_seconds": 0.226,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestSplitUnitCoverage",
      "TestShardSelection",
      "TestRetainedSplitCoverage",
      "TestRetainedTopLevelCoverage",
      "TestStringBuildingMatchesNode",
      "TestStringIndexMatchesNode",
      "TestStringViewAfterAppendMatchesNode",
      "TestStringIndexCacheStatesMatchNode",
      "TestStringsMatchJavaScript",
      "TestRuntimeStringViews"
    ],
    "evidence": "ADAMIC_MUTANT=M3 ADAMIC_BUILD_CACHE_DIR=/tmp/u053/cache/M3 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestSplitUnitCoverage|TestShardSelection|TestRetainedSplitCoverage|TestRetainedTopLevelCoverage|TestStringBuildingMatchesNode|TestStringIndexMatchesNode|TestStringViewAfterAppendMatchesNode|TestStringIndexCacheStatesMatchNode|TestStringsMatchJavaScript|TestRuntimeStringViews)$ > logs/M3.log 2>&1;     string_index_test.go:404: UTF-16 view after append differs from Node",
    "construction_kills": [],
    "entry_probe_results": {
      "adamic_string_allocate": "fail",
      "adamic_string_append": "fail",
      "adamic_string_at": "fail",
      "adamic_string_char_code_at": "fail",
      "adamic_string_code_point_at": "fail",
      "adamic_string_length": "fail",
      "adamic_string_slice": "fail",
      "adamic_string_units": "fail"
    },
    "subsumption_mutants": 2
  },
  {
    "test": "TestStringIndexCacheStatesMatchNode",
    "package": "internal/native",
    "file": "internal/native/string_index_test.go",
    "seconds": 0.197,
    "oracle": "Executes Node and compares complete indexOf and backward char-code output across literal, stack and heap cache states.",
    "oracle_kind": "external-run",
    "kills": [
      "M3"
    ],
    "unique_kills": [],
    "last_proven_fail": "M3:     string_index_test.go:453: cache states differ: native \"96 96 96\\n90 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 \\n90 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 \\n33 90 65 56833 55357 233 65 5683",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestStringIndexMatchesNode"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P4",
      "P9",
      "P13",
      "P15"
    ],
    "subsumer_seconds": 5.431,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestSplitUnitCoverage",
      "TestShardSelection",
      "TestRetainedSplitCoverage",
      "TestRetainedTopLevelCoverage",
      "TestStringBuildingMatchesNode",
      "TestStringIndexMatchesNode",
      "TestStringViewAfterAppendMatchesNode",
      "TestStringIndexCacheStatesMatchNode",
      "TestStringsMatchJavaScript",
      "TestRuntimeStringViews"
    ],
    "evidence": "ADAMIC_MUTANT=M3 ADAMIC_BUILD_CACHE_DIR=/tmp/u053/cache/M3 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestSplitUnitCoverage|TestShardSelection|TestRetainedSplitCoverage|TestRetainedTopLevelCoverage|TestStringBuildingMatchesNode|TestStringIndexMatchesNode|TestStringViewAfterAppendMatchesNode|TestStringIndexCacheStatesMatchNode|TestStringsMatchJavaScript|TestRuntimeStringViews)$ > logs/M3.log 2>&1;     string_index_test.go:453: cache states differ: native \"96 96 96\\n90 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 65 56833 55357 233 \\n90 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 65 56832 55357 233 \\n33 90 65 56833 55357 233 65 5683",
    "construction_kills": [],
    "entry_probe_results": {
      "adamic_string_char_code": "fail",
      "adamic_string_concat": "fail",
      "adamic_string_index_of": "fail",
      "adamic_string_length": "fail"
    },
    "subsumption_mutants": 1
  },
  {
    "test": "TestStringsMatchJavaScript",
    "package": "internal/native",
    "file": "internal/native/string_test.go",
    "seconds": 0.299,
    "oracle": "Executes Node and compares every answer plus exact answer count across 15 operation labels.",
    "oracle_kind": "external-run",
    "kills": [
      "M4"
    ],
    "unique_kills": [
      "M4"
    ],
    "last_proven_fail": "M4:     string_test.go:200: slice 61 - c008000000000000 7ff8000000000001: native \"a\", Node \"\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P5",
      "P6",
      "P7",
      "P8",
      "P9",
      "P10",
      "P11",
      "P13",
      "P16",
      "P19",
      "P20",
      "P21"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestSplitUnitCoverage",
      "TestShardSelection",
      "TestRetainedSplitCoverage",
      "TestRetainedTopLevelCoverage",
      "TestStringBuildingMatchesNode",
      "TestStringIndexMatchesNode",
      "TestStringViewAfterAppendMatchesNode",
      "TestStringIndexCacheStatesMatchNode",
      "TestStringsMatchJavaScript",
      "TestRuntimeStringViews"
    ],
    "evidence": "ADAMIC_MUTANT=M4 ADAMIC_BUILD_CACHE_DIR=/tmp/u053/cache/M4 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestSplitUnitCoverage|TestShardSelection|TestRetainedSplitCoverage|TestRetainedTopLevelCoverage|TestStringBuildingMatchesNode|TestStringIndexMatchesNode|TestStringViewAfterAppendMatchesNode|TestStringIndexCacheStatesMatchNode|TestStringsMatchJavaScript|TestRuntimeStringViews)$ > logs/M4.log 2>&1;     string_test.go:200: slice 61 - c008000000000000 7ff8000000000001: native \"a\", Node \"\"",
    "construction_kills": [],
    "entry_probe_results": {
      "adamic_string_char_code_at": "fail",
      "adamic_string_code_point_at": "fail",
      "adamic_string_code_points": "fail",
      "adamic_string_compare": "fail",
      "adamic_string_concat": "fail",
      "adamic_string_ends_with": "fail",
      "adamic_string_equal": "fail",
      "adamic_string_index_of": "fail",
      "adamic_string_length": "over-budget",
      "adamic_string_pad": "fail",
      "adamic_string_slice": "fail",
      "adamic_string_starts_with": "fail",
      "adamic_string_trim": "fail"
    },
    "subsumption_mutants": null
  },
  {
    "test": "TestRuntimeStringViews",
    "package": "internal/native",
    "file": "internal/native/string_views_test.go",
    "seconds": 0.303,
    "oracle": "Executes Node for output bytes; handwritten reference, allocation, view-offset and eightfold-storage invariants decide ownership. M1 is caught by that invariant before Node comparison.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M1"
    ],
    "unique_kills": [
      "M1"
    ],
    "last_proven_fail": "M1:     string_views_test.go:75: /tmp/adamic-gate/TestRuntimeStringViews2032290328/001/views: exit status 1",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1",
      "P3",
      "P18",
      "P19"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestSplitUnitCoverage",
      "TestShardSelection",
      "TestRetainedSplitCoverage",
      "TestRetainedTopLevelCoverage",
      "TestStringBuildingMatchesNode",
      "TestStringIndexMatchesNode",
      "TestStringViewAfterAppendMatchesNode",
      "TestStringIndexCacheStatesMatchNode",
      "TestStringsMatchJavaScript",
      "TestRuntimeStringViews"
    ],
    "evidence": "ADAMIC_MUTANT=M1 ADAMIC_BUILD_CACHE_DIR=/tmp/u053/cache/M1 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestSplitUnitCoverage|TestShardSelection|TestRetainedSplitCoverage|TestRetainedTopLevelCoverage|TestStringBuildingMatchesNode|TestStringIndexMatchesNode|TestStringViewAfterAppendMatchesNode|TestStringIndexCacheStatesMatchNode|TestStringsMatchJavaScript|TestRuntimeStringViews)$ > logs/M1.log 2>&1;     string_views_test.go:75: /tmp/adamic-gate/TestRuntimeStringViews2032290328/001/views: exit status 1",
    "construction_kills": [],
    "entry_probe_results": {
      "adamic_string_allocate": "fail",
      "adamic_string_at": "fail",
      "adamic_string_share": "fail",
      "adamic_string_slice": "fail"
    },
    "subsumption_mutants": null
  }
]
```

| id | origin/main file:line | one-line change | failed rows |
|---|---|---|---|
| M1 | internal/native/runtime/string_share.c:23 | `#define SHARE_FRACTION 8` -> `#define SHARE_FRACTION 7` | TestRuntimeStringViews |
| M2 | internal/native/runtime/string_append.c:88 | `result->units = units + 1;` -> `result->units = units + 2;` | TestStringBuildingMatchesNode, TestStringViewAfterAppendMatchesNode |
| M3 | internal/native/runtime/string_index.c:98 | `index->view[at++] = (uint16_t)(0xdc00 +` -> `index->view[at++] = (uint16_t)(0xdc01 +` | TestStringBuildingMatchesNode, TestStringIndexMatchesNode, TestStringViewAfterAppendMatchesNode, TestStringIndexCacheStatesMatchNode |
| M4 | internal/native/runtime/string_slice_impl.h:6 | `index = isnan(index) ? 0 : trunc(index);` -> `index = isnan(index) ? 1 : trunc(index);` | TestStringsMatchJavaScript |
| S1 | internal/native/split_units_test.go:35 | `for first := 0; first < total; first += width {` -> `for first := 1; first < total; first += width {` | TestSplitUnitCoverage |
| S2 | internal/native/split_units_test.go:318 | `count < 1 || index < 0` -> `count < 2 || index < 0` | TestRetainedSplitCoverage |
| S3 | internal/native/split_units_test.go:331 | `return piece%s.count == s.index` -> `return piece%s.count != s.index` | TestRetainedSplitCoverage |
| S4 | internal/native/record_test.go:352 | `runRecordMutantUnit(t, 0)` -> `runRecordMutantUnit(t, 1)` | TestRetainedTopLevelCoverage |

Survivors: none among M1-M4 or S1-S4. Probe passes are not survivors. S2 survives TestShardSelection while TestRetainedSplitCoverage catches its rejection of valid 0/1.

Friction and limitations:

- origin/main moved from the brief's historical 8de93800f4 to 0942c516. The requested names and files remained present.
- The full package spent its 90-second budget before completing. Matrix results support only the listed ten rows; other callers through generated C remain unknown. The reached-function inventory is conservative for C and measured for Go. Its explicit inventory was assembled after the fixed menu, contrary to the requested ordering.
- Multiple C API entries are called by each string harness. entry_probe_results records each direct entry separately. vacuous false means at least one own empty entry was observed failing; setup wrapper discovery has no empty production entry and remains null. A passing unrelated or preparation probe does not determine vacuity.
- The setup tests exercise code in _test.go. S1-S4 are construction mutations, not production mutations. TestShardSelection's setup-check proof is its own P24 empty-construction failure, not a production kill. Empty probes never support sacred or subsumed.
- P15 and P22 hung the JavaScript harness; each was rerun over every assigned row alone, retaining the same 90-second test budget. The repeated JavaScript timeout is over-budget, not a kill. P24's zero-count shard caused a Go divide-by-zero panic; each row was rerun alone. Go test JSON attached the panic to a currently active different test, so the stack and alone runs settle attribution.
- The first probe-definition matcher matched calls as definitions and then missed saving setup sources. Both issues were fixed before source mutation. The one-line owns probe initially lacked a statement separator and did not compile; it was corrected and revalidated. One corrected P25 run overlapped the physical S4 wrapper edit, so its confounded log was excluded and P25 was rerun after restoration.
- The switched C helpers perform getenv checks in hot paths. Timing uses original sources, not the instrumented matrix. Runtime source/header content is keyed in the runtime library cache; the selector executes inside every native product. No per-mutant native rebuild is needed with this switch. The clean switched run took 14.832 wall seconds, including library builds and execution; build-only time was not separately isolated.
- npm ci installed three dependencies in 0.389s. The warm Go/clang/Node environment was reused, setup.sh was not run, and nproc was 5. Three-per-row timing runs took 76.593 wall seconds. Full baseline binary time was 90.053s; narrow baseline 5.964s. Per-run wall times and standalone validation times are in timings.json.
- All assigned rows ran without skips. Baseline skips outside the assigned slice are recorded below. No WASI/corpus opt-in installation was attempted for outside-slice rows. No other packages, repo-wide replay, performance scaling at 100000 operations, or complete C branch coverage was audited. Four production mutants are a small sample; subsumption rests on one or two catches and is not a deletion recommendation.

Outside-slice baseline skipped rows: TestMeasureClangUnits, TestNormalizeLongMeasurements, TestRecordBenchmark, TestSplitTSGoAgreesUnit00, TestSplitTSGoAgreesUnit01, TestWASIUnit00, TestWASIUnit01, TestWASIUnit02, TestWASIUnit03, TestWASIUnit04, TestWASIUnit05, TestWASIUnit06, TestWASIUnit07, TestWASIUnit08, TestWASIUnit09, TestWASIUnit10, TestWASIUnit11, TestWASIUnit12, TestWASIUnit13, TestWASIUnit14, TestWASIUnit15, TestWASIUnit16, TestWASIUnit17, TestWASIUnit18, TestWASIUnit19, TestWASIUnit20, TestWASIUnit21, TestWASIUnit22, TestWASIUnit23, TestWASIUnit24, TestWASIUnit25, TestWASIUnit26, TestWASIUnit27, TestWASIUnit28, TestWASIUnit29, TestWASIUnit30, TestWASIUnit31, TestWASIUnit32, TestWASIUnit33, TestWASIUnit34, TestWASIUnit35.

Restored production sources: the same ten-row clean slice passed again in 5.919s. Standalone M1-M4, S1-S4 and P1-P25 all apply to the starting commit; all final standalone validations passed. P25's initial invalid validation log was replaced by its corrected successful validation.
