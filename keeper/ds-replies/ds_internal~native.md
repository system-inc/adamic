The baseline passed in 135 seconds; 24 replays took 1,453 seconds. No diffs were stale and no Go panics occurred. Keep the parser-thunk test; retain the number-format test pending resolution of one broken C mutant. The other five meet the replay’s deletable criterion. Source is restored, and [evidence](https://github.com/system-inc/adamic/blob/56569b91c8aa5d705c6a6a6e4a23d5b05ce30983/review/test-defend/deletion-set/internal-native/report.json) was pushed to `test-defend/deletion-set/internal-native`.

```json
{
  "package": "internal/native",
  "main": "7b9d4272c28f59530ab13daa5c49067e47933b06",
  "skipped": [
    "TestNumbersFormatExactlyAsJavaScriptDoes",
    "TestParserHasNoUnusedOptionalMethodThunks",
    "TestRecordBenchmark",
    "TestRegExpLintPatternsNode",
    "TestRegExpNativeStepLimit",
    "TestRegExpSearchNode",
    "TestRuntimeReleasePaths"
  ],
  "mutants": [
    {
      "replay": "R001",
      "mutant": "M15",
      "file_line": "internal/native/emit_objects.go:344",
      "branch": "test-audit/internal-native-arguments_length",
      "candidates_failed": ["TestParserHasNoUnusedOptionalMethodThunks"],
      "still_caught_by": [],
      "witness_failures": ["TestOptionalMethodThunksMatchNode"],
      "stale": false
    },
    {
      "replay": "R002",
      "mutant": "M01",
      "file_line": "internal/native/runtime/number.c:20",
      "branch": "test-audit/internal-native-number",
      "candidates_failed": ["TestNumbersFormatExactlyAsJavaScriptDoes"],
      "still_caught_by": ["TestToExponentialAndToPrecisionMatchNode"],
      "stale": false
    },
    {
      "replay": "R003",
      "mutant": "M02",
      "file_line": "internal/native/runtime/number.c:75",
      "branch": "test-audit/internal-native-number",
      "candidates_failed": ["TestNumbersFormatExactlyAsJavaScriptDoes"],
      "still_caught_by": ["TestToExponentialAndToPrecisionMatchNode"],
      "stale": false
    },
    {
      "replay": "R004",
      "mutant": "M03",
      "file_line": "internal/native/runtime/number.c:93",
      "branch": "test-audit/internal-native-number",
      "candidates_failed": ["TestNumbersFormatExactlyAsJavaScriptDoes"],
      "still_caught_by": ["TestTypedArrayRuntime"],
      "stale": false
    },
    {
      "replay": "R005",
      "mutant": "M07",
      "file_line": "internal/native/runtime/regexp.c:9",
      "branch": "test-audit/internal-native-radix",
      "candidates_failed": ["TestRegExpLintPatternsNode", "TestRegExpSearchNode"],
      "still_caught_by": ["TestRegExpBytecodeRandomNodeUnit39"],
      "stale": false
    },
    {
      "replay": "R006",
      "mutant": "M08",
      "file_line": "internal/native/runtime/regexp.c:82",
      "branch": "test-audit/internal-native-radix",
      "candidates_failed": [
        "TestRegExpLintPatternsNode",
        "TestRegExpNativeStepLimit",
        "TestRegExpSearchNode"
      ],
      "still_caught_by": ["TestRegExpBytecodeRandomNodeUnit32"],
      "stale": false
    },
    {
      "replay": "R007",
      "mutant": "M10",
      "file_line": "internal/native/runtime/string_build_impl.h:127",
      "branch": "test-audit/internal-native-radix",
      "candidates_failed": [
        "TestRecordBenchmark",
        "TestRegExpLintPatternsNode",
        "TestRegExpSearchNode"
      ],
      "still_caught_by": [
        "TestRegExpBytecodeRandomNodeUnit28",
        "TestRegExpBytecodeRandomNodeUnit27",
        "TestRegExpBytecodeRandomNodeUnit26"
      ],
      "stale": false
    },
    {
      "replay": "R008",
      "mutant": "M11",
      "file_line": "internal/native/runtime/heap.c:375",
      "branch": "test-audit/internal-native-radix",
      "candidates_failed": [
        "TestRegExpLintPatternsNode",
        "TestRegExpSearchNode",
        "TestRuntimeReleasePaths"
      ],
      "still_caught_by": ["TestSizeClassesShareTheirChunks"],
      "stale": false
    },
    {
      "replay": "R009",
      "mutant": "D3",
      "file_line": "internal/native/emit_objects.go:345",
      "branch": "test-defend/internal-native-arguments_length",
      "candidates_failed": ["TestParserHasNoUnusedOptionalMethodThunks"],
      "still_caught_by": [],
      "witness_failures": ["TestOptionalMethodThunksMatchNode"],
      "stale": false
    },
    {
      "replay": "R010",
      "mutant": "D4",
      "file_line": "internal/native/emit_objects.go:327",
      "branch": "test-defend/internal-native-arguments_length",
      "candidates_failed": ["TestParserHasNoUnusedOptionalMethodThunks"],
      "still_caught_by": [],
      "witness_failures": ["TestOptionalMethodThunksMatchNode"],
      "stale": false
    },
    {
      "replay": "R011",
      "mutant": "D5",
      "file_line": "internal/native/emit_objects.go:344",
      "branch": "test-defend/internal-native-arguments_length",
      "candidates_failed": ["TestParserHasNoUnusedOptionalMethodThunks"],
      "still_caught_by": [],
      "witness_failures": ["TestOptionalMethodThunksMatchNode"],
      "stale": false
    },
    {
      "replay": "R012",
      "mutant": "D1",
      "file_line": "internal/native/runtime/dtoa.c:619",
      "branch": "test-defend/internal-native-number",
      "candidates_failed": ["TestNumbersFormatExactlyAsJavaScriptDoes"],
      "still_caught_by": ["TestToExponentialAndToPrecisionMatchNode"],
      "stale": false
    },
    {
      "replay": "R013",
      "mutant": "D2",
      "file_line": "internal/native/runtime/dtoa.c:591",
      "branch": "test-defend/internal-native-number",
      "candidates_failed": ["TestNumbersFormatExactlyAsJavaScriptDoes"],
      "still_caught_by": [],
      "stale": false,
      "broken": "C compilation fails: dtoa.c:427:13: unused function 'bignum_less' [-Werror,-Wunused-function]",
      "build_failure_rows": ["TestDecodeASCIIUnit42", "TestDecodeASCIIUnit41"]
    },
    {
      "replay": "R014",
      "mutant": "D3",
      "file_line": "internal/native/runtime/dtoa.c:596",
      "branch": "test-defend/internal-native-number",
      "candidates_failed": ["TestNumbersFormatExactlyAsJavaScriptDoes"],
      "still_caught_by": ["TestPowersOfTwoFormatAsJavaScriptDoes"],
      "stale": false
    },
    {
      "replay": "R015",
      "mutant": "D3",
      "file_line": "internal/native/runtime/regexp.c:188",
      "branch": "test-defend/internal-native-radix",
      "candidates_failed": ["TestRegExpNativeStepLimit"],
      "still_caught_by": ["TestRegExpNativeStepLimitBoundary"],
      "stale": false
    },
    {
      "replay": "R016",
      "mutant": "D5",
      "file_line": "internal/native/runtime/regexp.c:9",
      "branch": "test-defend/internal-native-radix",
      "candidates_failed": ["TestRegExpNativeStepLimit"],
      "still_caught_by": ["TestRegExpNativeStepLimitBoundary"],
      "stale": false
    },
    {
      "replay": "R017",
      "mutant": "D7",
      "file_line": "internal/native/runtime/regexp.c:464",
      "branch": "test-defend/internal-native-radix",
      "candidates_failed": ["TestRegExpSearchNode"],
      "still_caught_by": ["TestRegExpBytecodeRandomNodeUnit26"],
      "stale": false
    },
    {
      "replay": "R018",
      "mutant": "D1",
      "file_line": "internal/native/runtime/regexp.c:115",
      "branch": "test-defend/internal-native-radix",
      "candidates_failed": ["TestRegExpLintPatternsNode", "TestRegExpSearchNode"],
      "still_caught_by": ["TestRegExpBytecodeTest262"],
      "stale": false
    },
    {
      "replay": "R019",
      "mutant": "D2",
      "file_line": "internal/native/runtime/regexp.c:464",
      "branch": "test-defend/internal-native-radix",
      "candidates_failed": ["TestRegExpSearchNode"],
      "still_caught_by": ["TestRegExpBytecodeRandomNodeUnit11"],
      "stale": false
    },
    {
      "replay": "R020",
      "mutant": "D3",
      "file_line": "internal/regexp/native_search.go:42",
      "branch": "test-defend/internal-native-radix",
      "candidates_failed": ["TestRegExpSearchNode"],
      "still_caught_by": [
        "TestRegExpBytecodeRandomNodeUnit18",
        "TestRegExpBytecodeRandomNodeUnit17",
        "TestRegExpBytecodeRandomNodeUnit21"
      ],
      "stale": false
    },
    {
      "replay": "R021",
      "mutant": "D5",
      "file_line": "internal/regexp/native_search.go:75",
      "branch": "test-defend/internal-native-radix",
      "candidates_failed": ["TestRegExpLintPatternsNode", "TestRegExpSearchNode"],
      "still_caught_by": ["TestRegExpBytecodeRandomNodeUnit31"],
      "stale": false
    },
    {
      "replay": "R022",
      "mutant": "D6",
      "file_line": "internal/native/runtime/regexp.c:291",
      "branch": "test-defend/internal-native-radix",
      "candidates_failed": ["TestRegExpLintPatternsNode"],
      "still_caught_by": ["TestRegExpBytecodeRandomNodeUnit39"],
      "stale": false
    },
    {
      "replay": "R023",
      "mutant": "D02",
      "file_line": "internal/native/runtime/record.c:207",
      "branch": "test-defend/internal-native-radix",
      "candidates_failed": ["TestRecordBenchmark"],
      "still_caught_by": ["TestRecordsAgainstNode"],
      "stale": false
    },
    {
      "replay": "R024",
      "mutant": "D03",
      "file_line": "internal/native/runtime/map.c:222",
      "branch": "test-defend/internal-native-radix",
      "candidates_failed": ["TestRecordBenchmark"],
      "still_caught_by": ["TestRecordsAgainstNode"],
      "stale": false
    }
  ],
  "keep": [
    {
      "test": "TestParserHasNoUnusedOptionalMethodThunks",
      "because": "R001/M15, R009/D3, R010/D4 and R011/D5 lose their last non-witness catcher without it."
    },
    {
      "test": "TestNumbersFormatExactlyAsJavaScriptDoes",
      "because": "R013/D2 is broken on current main; deletion remains unresolved."
    }
  ],
  "deletable": [
    "TestRecordBenchmark",
    "TestRegExpLintPatternsNode",
    "TestRegExpNativeStepLimit",
    "TestRegExpSearchNode",
    "TestRuntimeReleasePaths"
  ],
  "panicking_tests": [],
  "notes": [
    "Baseline: 190 top-level passes and 85 default skips.",
    "Caught runs stopped after a clean non-witness failure; later rows are unknown.",
    "All four witness-only runs completed the package.",
    "Repeated mutant IDs are disambiguated by replay IDs and source paths in mutant-list.json."
  ]
}
```
