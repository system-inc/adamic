Defended 3 rows in bounded direct-consumer matrices.
All other rows remain kept; incomplete defenses are cannot-judge.
Evidence includes standalone diffs, compile/apply logs, coverage pairs and exact passed-row lists.

```json
[
  {
    "test": "TestToStringWithARadixOutOfRangePanics",
    "package": "internal/native",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestToStringWithARadixMatchesNode"
    ],
    "defense": "defended",
    "unique_mutant": "D1 internal/native/runtime/radix.c:145",
    "attempts": [
      {
        "mutant": "D1",
        "file_line": "internal/native/runtime/radix.c:145",
        "change": "adamic_panic(message, sizeof message - 1); -> adamic_panic(message, sizeof message - 2);",
        "rows_failed": [
          "TestToStringWithARadixOutOfRangePanics"
        ],
        "intent": "Out-of-range diagnostics: write one byte fewer; valid-radix subsumer never enters the RangeError branch."
      }
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestToStringWithARadixMatchesNode|TestToStringWithARadixOutOfRangePanics)$'; radix_test.go:148: (255).toString(1): native stderr \"adamic: panic: RangeError: toString() radix argument must be between 2 and 3\\n\", Node stderr \"[eval]:1\\nconsole.log('before'); console.log((255).toString(1));\\n                                         ^\\n\\nRangeError: toString() radix argument must be between 2 and 36\\n    at Number.toString (<anonymous>)\\n    at [eval]:1:42\\n    at runScriptInThisContext (node:internal/vm:219:10)\\n    at node:internal/process/execution:451:12\\n    at [eval]-wrapper:6:24\\n    at runScriptInContext (node:internal/process/execution:449:60)\\n    at evalFunction (node:internal/process/execution:283:30)\\n    at evalTypeScript (node:internal/process/execution:295:3)\\n    at node:internal/main/eval_string:71:3\\n\\nNode.js v24.19.0\\n\"",
    "bounded": true,
    "reason": "Unique within completed direct-consumer matrix; other package suites are unknown.",
    "rows_passed": [
      "TestToStringWithARadixMatchesNode"
    ],
    "code_under_test": "internal/native/radix_test.go exercises production runtime and regexp compilation; see code-and-oracle.md",
    "oracle": "Live Node output and exception classification; native exit70 and exact diagnostic prefix."
  },
  {
    "test": "TestRecordBenchmark",
    "package": "internal/native",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestRuntimeStringEquality"
    ],
    "defense": "cannot-judge",
    "unique_mutant": null,
    "attempts": [
      {
        "mutant": "D2",
        "file_line": "internal/native/runtime/string_build_impl.h:125",
        "change": "return left == right; -> return false;",
        "rows_failed": [
          "TestRuntimeStringEquality"
        ],
        "intent": "Undefined equals itself; regex comparisons pass nonnull strings."
      }
    ],
    "evidence": "No isolated targeted kill established; see coverage-differences.json and bounded matrix logs.",
    "bounded": true,
    "reason": "Seven-mutant budget exhausted before three dedicated attempts. Exploratory shared-path runs do not establish redundancy.",
    "rows_passed": [],
    "code_under_test": "internal/native/record_test.go exercises production runtime and regexp compilation; see code-and-oracle.md",
    "oracle": "Live Node work checksum; self five timing-field validity checks, no performance threshold."
  },
  {
    "test": "TestRegExpSearchNode",
    "package": "internal/native",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestRegExpLintPatternsNode"
    ],
    "defense": "cannot-judge",
    "unique_mutant": null,
    "attempts": [
      {
        "mutant": "D7",
        "file_line": "internal/native/runtime/regexp.c:464",
        "change": "\t\t\tat--; -> \t\t\t(void)at;",
        "rows_failed": [
          "TestRegExpBytecodeRandomNode family",
          "TestRegExpSearchNode"
        ],
        "intent": "Drop low-surrogate lastIndex rewind; search explicitly starts inside a surrogate pair."
      }
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestRegexProgramsKeepCheckedFieldReads|TestRegExpSearchNode|TestRegExpLintPatternsNode|TestRegExpBytecodeTest262|TestRegExpNativeStepLimit|TestRegExpNativeStepLimitBoundary|TestRegExpIteratorResultShape|TestRegExpBytecodePatternUnits|TestRegExpBytecodeRandomNodeUnit00|TestRegExpBytecodeRandomNodeUnit01|TestRegExpBytecodeRandomNodeUnit02|TestRegExpBytecodeRandomNodeUnit03|TestRegExpBytecodeRandomNodeUnit04|TestRegExpBytecodeRandomNodeUnit05|TestRegExpBytecodeRandomNodeUnit06|TestRegExpBytecodeRandomNodeUnit07|TestRegExpBytecodeRandomNodeUnit08|TestRegExpBytecodeRandomNodeUnit09|TestRegExpBytecodeRandomNodeUnit10|TestRegExpBytecodeRandomNodeUnit11|TestRegExpBytecodeRandomNodeUnit12|TestRegExpBytecodeRandomNodeUnit13|TestRegExpBytecodeRandomNodeUnit14|TestRegExpBytecodeRandomNodeUnit15|TestRegExpBytecodeRandomNodeUnit16|TestRegExpBytecodeRandomNodeUnit17|TestRegExpBytecodeRandomNodeUnit18|TestRegExpBytecodeRandomNodeUnit19|TestRegExpBytecodeRandomNodeUnit20|TestRegExpBytecodeRandomNodeUnit21|TestRegExpBytecodeRandomNodeUnit22|TestRegExpBytecodeRandomNodeUnit23|TestRegExpBytecodeRandomNodeUnit24|TestRegExpBytecodeRandomNodeUnit25|TestRegExpBytecodeRandomNodeUnit26|TestRegExpBytecodeRandomNodeUnit27|TestRegExpBytecodeRandomNodeUnit28|TestRegExpBytecodeRandomNodeUnit29|TestRegExpBytecodeRandomNodeUnit30|TestRegExpBytecodeRandomNodeUnit31|TestRegExpBytecodeRandomNodeUnit32|TestRegExpBytecodeRandomNodeUnit33|TestRegExpBytecodeRandomNodeUnit34|TestRegExpBytecodeRandomNodeUnit35|TestRegExpBytecodeRandomNodeUnit36|TestRegExpBytecodeRandomNodeUnit37|TestRegExpBytecodeRandomNodeUnit38|TestRegExpBytecodeRandomNodeUnit39)$'; regexp_search_test.go:34: native regex oracle ([]): exit status 1",
    "bounded": true,
    "reason": "Seven-mutant budget exhausted before three dedicated attempts. Exploratory shared-path runs do not establish redundancy.",
    "rows_passed": [],
    "code_under_test": "internal/native/regexp_search_test.go exercises production runtime and regexp compilation; see code-and-oracle.md",
    "oracle": "Live Node captures, groups, indices and lastIndex."
  },
  {
    "test": "TestRegExpLintPatternsNode",
    "package": "internal/native",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestRegExpSearchNode"
    ],
    "defense": "cannot-judge",
    "unique_mutant": null,
    "attempts": [
      {
        "mutant": "D7",
        "file_line": "internal/native/runtime/regexp.c:464",
        "change": "\t\t\tat--; -> \t\t\t(void)at;",
        "rows_failed": [
          "TestRegExpBytecodeRandomNode family",
          "TestRegExpSearchNode"
        ],
        "intent": "Drop low-surrogate lastIndex rewind; search explicitly starts inside a surrogate pair."
      }
    ],
    "evidence": "No isolated targeted kill established; see coverage-differences.json and bounded matrix logs.",
    "bounded": true,
    "reason": "Seven-mutant budget exhausted before three dedicated attempts. Exploratory shared-path runs do not establish redundancy.",
    "rows_passed": [],
    "code_under_test": "internal/native/regexp_search_test.go exercises production runtime and regexp compilation; see code-and-oracle.md",
    "oracle": "Live Node captures, groups, indices and lastIndex on lint patterns."
  },
  {
    "test": "TestRegExpBytecodeTest262",
    "package": "internal/native",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestRegExpSearchNode"
    ],
    "defense": "cannot-judge",
    "unique_mutant": null,
    "attempts": [
      {
        "mutant": "D6",
        "file_line": "internal/regexp/parser.go:63",
        "change": "byte(0x80|c&0x3f) -> byte(0x80|c&0x3e)",
        "rows_failed": [
          "TestRegExpBytecodePatternUnits"
        ],
        "intent": "Off-by-one UTF16 pattern encoding: clear the low bit of the final surrogate byte; exact pattern identity differs from JSON text."
      },
      {
        "mutant": "D7",
        "file_line": "internal/native/runtime/regexp.c:464",
        "change": "\t\t\tat--; -> \t\t\t(void)at;",
        "rows_failed": [
          "TestRegExpBytecodeRandomNode family",
          "TestRegExpSearchNode"
        ],
        "intent": "Drop low-surrogate lastIndex rewind; search explicitly starts inside a surrogate pair."
      }
    ],
    "evidence": "No isolated targeted kill established; see coverage-differences.json and bounded matrix logs.",
    "bounded": true,
    "reason": "Seven-mutant budget exhausted before three dedicated attempts. Exploratory shared-path runs do not establish redundancy.",
    "rows_passed": [],
    "code_under_test": "internal/native/regexp_test.go exercises production runtime and regexp compilation; see code-and-oracle.md",
    "oracle": "Recorded test262 RegExp observations in matches.json.gz; first digit-class result checked against live Node in authority-check.json."
  },
  {
    "test": "TestRegExpNativeStepLimit",
    "package": "internal/native",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestRegExpBytecodePatternUnits"
    ],
    "defense": "not defended",
    "unique_mutant": null,
    "attempts": [
      {
        "mutant": "D3",
        "file_line": "internal/native/runtime/regexp.c:188",
        "change": "regexp: instruction step limit exceeded -> regexp: instruction budget exceeded",
        "rows_failed": [
          "TestRegExpNativeStepLimit",
          "TestRegExpNativeStepLimitBoundary"
        ],
        "intent": "Diagnostic text: catastrophic-budget row and newly added exact boundary row."
      },
      {
        "mutant": "D4",
        "file_line": "internal/native/runtime/regexp.c:187",
        "change": "*steps >= regex_step_limit -> *steps > regex_step_limit",
        "rows_failed": [
          "TestRegExpNativeStepLimitBoundary"
        ],
        "intent": "Exact consumed-instruction boundary; new boundary row may catch what catastrophic row does not."
      },
      {
        "mutant": "D5",
        "file_line": "internal/native/runtime/regexp.c:9",
        "change": "regex_step_limit = limit; -> regex_step_limit = 0;",
        "rows_failed": [
          "TestRegExpNativeStepLimit",
          "TestRegExpNativeStepLimitBoundary"
        ],
        "intent": "Drop step-limit configuration, replacing assignment with zero. Catastrophic backtracking must be interrupted."
      }
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestRegexProgramsKeepCheckedFieldReads|TestRegExpSearchNode|TestRegExpLintPatternsNode|TestRegExpBytecodeTest262|TestRegExpNativeStepLimit|TestRegExpNativeStepLimitBoundary|TestRegExpIteratorResultShape|TestRegExpBytecodePatternUnits|TestRegExpBytecodeRandomNodeUnit00|TestRegExpBytecodeRandomNodeUnit01|TestRegExpBytecodeRandomNodeUnit02|TestRegExpBytecodeRandomNodeUnit03|TestRegExpBytecodeRandomNodeUnit04|TestRegExpBytecodeRandomNodeUnit05|TestRegExpBytecodeRandomNodeUnit06|TestRegExpBytecodeRandomNodeUnit07|TestRegExpBytecodeRandomNodeUnit08|TestRegExpBytecodeRandomNodeUnit09|TestRegExpBytecodeRandomNodeUnit10|TestRegExpBytecodeRandomNodeUnit11|TestRegExpBytecodeRandomNodeUnit12|TestRegExpBytecodeRandomNodeUnit13|TestRegExpBytecodeRandomNodeUnit14|TestRegExpBytecodeRandomNodeUnit15|TestRegExpBytecodeRandomNodeUnit16|TestRegExpBytecodeRandomNodeUnit17|TestRegExpBytecodeRandomNodeUnit18|TestRegExpBytecodeRandomNodeUnit19|TestRegExpBytecodeRandomNodeUnit20|TestRegExpBytecodeRandomNodeUnit21|TestRegExpBytecodeRandomNodeUnit22|TestRegExpBytecodeRandomNodeUnit23|TestRegExpBytecodeRandomNodeUnit24|TestRegExpBytecodeRandomNodeUnit25|TestRegExpBytecodeRandomNodeUnit26|TestRegExpBytecodeRandomNodeUnit27|TestRegExpBytecodeRandomNodeUnit28|TestRegExpBytecodeRandomNodeUnit29|TestRegExpBytecodeRandomNodeUnit30|TestRegExpBytecodeRandomNodeUnit31|TestRegExpBytecodeRandomNodeUnit32|TestRegExpBytecodeRandomNodeUnit33|TestRegExpBytecodeRandomNodeUnit34|TestRegExpBytecodeRandomNodeUnit35|TestRegExpBytecodeRandomNodeUnit36|TestRegExpBytecodeRandomNodeUnit37|TestRegExpBytecodeRandomNodeUnit38|TestRegExpBytecodeRandomNodeUnit39)$'; regexp_test.go:273: native catastrophic backtracking: exit=exit status 70 output=adamic: panic: regexp: instruction budget exceeded",
    "bounded": true,
    "reason": "Three attempts: changed diagnostic and removed limit configuration were also caught by the new boundary row; off-by-one limit was caught only by that boundary row.",
    "rows_passed": [],
    "code_under_test": "internal/native/regexp_test.go exercises production runtime and regexp compilation; see code-and-oracle.md",
    "oracle": "Self exit70 and instruction-step-limit diagnostic; does not verify exact consumed instruction count, M12 survives."
  },
  {
    "test": "TestRegExpBytecodePatternUnits",
    "package": "internal/native",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestRegExpSearchNode"
    ],
    "defense": "defended",
    "unique_mutant": "D6 internal/regexp/parser.go:63",
    "attempts": [
      {
        "mutant": "D6",
        "file_line": "internal/regexp/parser.go:63",
        "change": "byte(0x80|c&0x3f) -> byte(0x80|c&0x3e)",
        "rows_failed": [
          "TestRegExpBytecodePatternUnits"
        ],
        "intent": "Off-by-one UTF16 pattern encoding: clear the low bit of the final surrogate byte; exact pattern identity differs from JSON text."
      }
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestRegexProgramsKeepCheckedFieldReads|TestRegExpSearchNode|TestRegExpLintPatternsNode|TestRegExpBytecodeTest262|TestRegExpNativeStepLimit|TestRegExpNativeStepLimitBoundary|TestRegExpIteratorResultShape|TestRegExpBytecodePatternUnits|TestRegExpBytecodeRandomNodeUnit00|TestRegExpBytecodeRandomNodeUnit01|TestRegExpBytecodeRandomNodeUnit02|TestRegExpBytecodeRandomNodeUnit03|TestRegExpBytecodeRandomNodeUnit04|TestRegExpBytecodeRandomNodeUnit05|TestRegExpBytecodeRandomNodeUnit06|TestRegExpBytecodeRandomNodeUnit07|TestRegExpBytecodeRandomNodeUnit08|TestRegExpBytecodeRandomNodeUnit09|TestRegExpBytecodeRandomNodeUnit10|TestRegExpBytecodeRandomNodeUnit11|TestRegExpBytecodeRandomNodeUnit12|TestRegExpBytecodeRandomNodeUnit13|TestRegExpBytecodeRandomNodeUnit14|TestRegExpBytecodeRandomNodeUnit15|TestRegExpBytecodeRandomNodeUnit16|TestRegExpBytecodeRandomNodeUnit17|TestRegExpBytecodeRandomNodeUnit18|TestRegExpBytecodeRandomNodeUnit19|TestRegExpBytecodeRandomNodeUnit20|TestRegExpBytecodeRandomNodeUnit21|TestRegExpBytecodeRandomNodeUnit22|TestRegExpBytecodeRandomNodeUnit23|TestRegExpBytecodeRandomNodeUnit24|TestRegExpBytecodeRandomNodeUnit25|TestRegExpBytecodeRandomNodeUnit26|TestRegExpBytecodeRandomNodeUnit27|TestRegExpBytecodeRandomNodeUnit28|TestRegExpBytecodeRandomNodeUnit29|TestRegExpBytecodeRandomNodeUnit30|TestRegExpBytecodeRandomNodeUnit31|TestRegExpBytecodeRandomNodeUnit32|TestRegExpBytecodeRandomNodeUnit33|TestRegExpBytecodeRandomNodeUnit34|TestRegExpBytecodeRandomNodeUnit35|TestRegExpBytecodeRandomNodeUnit36|TestRegExpBytecodeRandomNodeUnit37|TestRegExpBytecodeRandomNodeUnit38|TestRegExpBytecodeRandomNodeUnit39)$'; regexp_test.go:366: native regex oracle ([]): exit status 1",
    "bounded": true,
    "reason": "Unique within completed direct-consumer matrix; other package suites are unknown.",
    "rows_passed": [
      "TestRegExpBytecodeRandomNode family",
      "TestRegExpBytecodeTest262",
      "TestRegExpIteratorResultShape",
      "TestRegExpLintPatternsNode",
      "TestRegExpNativeStepLimit",
      "TestRegExpNativeStepLimitBoundary",
      "TestRegExpSearchNode",
      "TestRegexProgramsKeepCheckedFieldReads"
    ],
    "code_under_test": "internal/native/regexp_test.go exercises production runtime and regexp compilation; see code-and-oracle.md",
    "oracle": "Self two lone-surrogate capture spans [0,1]."
  },
  {
    "test": "TestRuntimeReleasePaths",
    "package": "internal/native",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestRegExpIteratorResultShape"
    ],
    "defense": "cannot-judge",
    "unique_mutant": null,
    "attempts": [],
    "evidence": "No isolated targeted kill established; see coverage-differences.json and bounded matrix logs.",
    "bounded": true,
    "reason": "Seven-mutant budget exhausted before three dedicated attempts. Exploratory shared-path runs do not establish redundancy.",
    "rows_passed": [],
    "code_under_test": "internal/native/runtime_profile_test.go exercises production runtime and regexp compilation; see code-and-oracle.md",
    "oracle": "Self live allocation counts 1 then0, fixed stdout; ASan and UBSan."
  },
  {
    "test": "TestRuntimeStringEquality",
    "package": "internal/native",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestRegExpSearchNode"
    ],
    "defense": "defended",
    "unique_mutant": "D2 internal/native/runtime/string_build_impl.h:125",
    "attempts": [
      {
        "mutant": "D2",
        "file_line": "internal/native/runtime/string_build_impl.h:125",
        "change": "return left == right; -> return false;",
        "rows_failed": [
          "TestRuntimeStringEquality"
        ],
        "intent": "Undefined equals itself; regex comparisons pass nonnull strings."
      }
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestRegexProgramsKeepCheckedFieldReads|TestRecordsAgainstNode|TestRecordReadMutants|TestRecordBenchmark|TestRecordMutantsUnit00|TestRecordMutantsUnit01|TestRecordMutantsUnit02|TestRecordMutantsUnit03|TestRecordMutantsUnit04|TestRecordMutantsUnit05|TestRegExpSearchNode|TestRegExpLintPatternsNode|TestRegExpBytecodeTest262|TestRegExpNativeStepLimit|TestRegExpNativeStepLimitBoundary|TestRegExpIteratorResultShape|TestRegExpBytecodePatternUnits|TestRegExpBytecodeRandomNodeUnit00|TestRegExpBytecodeRandomNodeUnit01|TestRegExpBytecodeRandomNodeUnit02|TestRegExpBytecodeRandomNodeUnit03|TestRegExpBytecodeRandomNodeUnit04|TestRegExpBytecodeRandomNodeUnit05|TestRegExpBytecodeRandomNodeUnit06|TestRegExpBytecodeRandomNodeUnit07|TestRegExpBytecodeRandomNodeUnit08|TestRegExpBytecodeRandomNodeUnit09|TestRegExpBytecodeRandomNodeUnit10|TestRegExpBytecodeRandomNodeUnit11|TestRegExpBytecodeRandomNodeUnit12|TestRegExpBytecodeRandomNodeUnit13|TestRegExpBytecodeRandomNodeUnit14|TestRegExpBytecodeRandomNodeUnit15|TestRegExpBytecodeRandomNodeUnit16|TestRegExpBytecodeRandomNodeUnit17|TestRegExpBytecodeRandomNodeUnit18|TestRegExpBytecodeRandomNodeUnit19|TestRegExpBytecodeRandomNodeUnit20|TestRegExpBytecodeRandomNodeUnit21|TestRegExpBytecodeRandomNodeUnit22|TestRegExpBytecodeRandomNodeUnit23|TestRegExpBytecodeRandomNodeUnit24|TestRegExpBytecodeRandomNodeUnit25|TestRegExpBytecodeRandomNodeUnit26|TestRegExpBytecodeRandomNodeUnit27|TestRegExpBytecodeRandomNodeUnit28|TestRegExpBytecodeRandomNodeUnit29|TestRegExpBytecodeRandomNodeUnit30|TestRegExpBytecodeRandomNodeUnit31|TestRegExpBytecodeRandomNodeUnit32|TestRegExpBytecodeRandomNodeUnit33|TestRegExpBytecodeRandomNodeUnit34|TestRegExpBytecodeRandomNodeUnit35|TestRegExpBytecodeRandomNodeUnit36|TestRegExpBytecodeRandomNodeUnit37|TestRegExpBytecodeRandomNodeUnit38|TestRegExpBytecodeRandomNodeUnit39|TestRuntimeStringEquality|TestStringsMatchJavaScript)$'; runtime_profile_test.go:81: sanitize false: \"1 1 0 0 0 0\\n\"; Node \"1 1 0 1 0 0\\n\"",
    "bounded": true,
    "reason": "Unique within completed direct-consumer matrix; other package suites are unknown.",
    "rows_passed": [
      "TestRecordBenchmark",
      "TestRecordMutants family",
      "TestRecordReadMutants",
      "TestRecordsAgainstNode",
      "TestRegExpBytecodePatternUnits",
      "TestRegExpBytecodeRandomNode family",
      "TestRegExpBytecodeTest262",
      "TestRegExpIteratorResultShape",
      "TestRegExpLintPatternsNode",
      "TestRegExpNativeStepLimit",
      "TestRegExpNativeStepLimitBoundary",
      "TestRegExpSearchNode",
      "TestRegexProgramsKeepCheckedFieldReads",
      "TestStringsMatchJavaScript"
    ],
    "code_under_test": "internal/native/runtime_profile_test.go exercises production runtime and regexp compilation; see code-and-oracle.md",
    "oracle": "Live Node strict string equality including undefined, exact six booleans."
  },
  {
    "test": "TestRegExpBytecodeRandomNode family",
    "package": "internal/native",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestRegExpSearchNode"
    ],
    "defense": "cannot-judge",
    "unique_mutant": null,
    "attempts": [
      {
        "mutant": "D6",
        "file_line": "internal/regexp/parser.go:63",
        "change": "byte(0x80|c&0x3f) -> byte(0x80|c&0x3e)",
        "rows_failed": [
          "TestRegExpBytecodePatternUnits"
        ],
        "intent": "Off-by-one UTF16 pattern encoding: clear the low bit of the final surrogate byte; exact pattern identity differs from JSON text."
      },
      {
        "mutant": "D7",
        "file_line": "internal/native/runtime/regexp.c:464",
        "change": "\t\t\tat--; -> \t\t\t(void)at;",
        "rows_failed": [
          "TestRegExpBytecodeRandomNode family",
          "TestRegExpSearchNode"
        ],
        "intent": "Drop low-surrogate lastIndex rewind; search explicitly starts inside a surrogate pair."
      }
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestRegexProgramsKeepCheckedFieldReads|TestRegExpSearchNode|TestRegExpLintPatternsNode|TestRegExpBytecodeTest262|TestRegExpNativeStepLimit|TestRegExpNativeStepLimitBoundary|TestRegExpIteratorResultShape|TestRegExpBytecodePatternUnits|TestRegExpBytecodeRandomNodeUnit00|TestRegExpBytecodeRandomNodeUnit01|TestRegExpBytecodeRandomNodeUnit02|TestRegExpBytecodeRandomNodeUnit03|TestRegExpBytecodeRandomNodeUnit04|TestRegExpBytecodeRandomNodeUnit05|TestRegExpBytecodeRandomNodeUnit06|TestRegExpBytecodeRandomNodeUnit07|TestRegExpBytecodeRandomNodeUnit08|TestRegExpBytecodeRandomNodeUnit09|TestRegExpBytecodeRandomNodeUnit10|TestRegExpBytecodeRandomNodeUnit11|TestRegExpBytecodeRandomNodeUnit12|TestRegExpBytecodeRandomNodeUnit13|TestRegExpBytecodeRandomNodeUnit14|TestRegExpBytecodeRandomNodeUnit15|TestRegExpBytecodeRandomNodeUnit16|TestRegExpBytecodeRandomNodeUnit17|TestRegExpBytecodeRandomNodeUnit18|TestRegExpBytecodeRandomNodeUnit19|TestRegExpBytecodeRandomNodeUnit20|TestRegExpBytecodeRandomNodeUnit21|TestRegExpBytecodeRandomNodeUnit22|TestRegExpBytecodeRandomNodeUnit23|TestRegExpBytecodeRandomNodeUnit24|TestRegExpBytecodeRandomNodeUnit25|TestRegExpBytecodeRandomNodeUnit26|TestRegExpBytecodeRandomNodeUnit27|TestRegExpBytecodeRandomNodeUnit28|TestRegExpBytecodeRandomNodeUnit29|TestRegExpBytecodeRandomNodeUnit30|TestRegExpBytecodeRandomNodeUnit31|TestRegExpBytecodeRandomNodeUnit32|TestRegExpBytecodeRandomNodeUnit33|TestRegExpBytecodeRandomNodeUnit34|TestRegExpBytecodeRandomNodeUnit35|TestRegExpBytecodeRandomNodeUnit36|TestRegExpBytecodeRandomNodeUnit37|TestRegExpBytecodeRandomNodeUnit38|TestRegExpBytecodeRandomNodeUnit39)$'; regexp_test.go:451: native regex oracle ([]): exit status 1",
    "bounded": true,
    "reason": "Seven-mutant budget exhausted before three dedicated attempts. Exploratory shared-path runs do not establish redundancy.",
    "rows_passed": [],
    "code_under_test": "internal/native/regexp_test.go exercises production runtime and regexp compilation; see code-and-oracle.md",
    "oracle": "Live Node captures, groups, indices and lastIndex; forty wrappers share runRegexCases."
  }
]
```

Friction and limits:
- The audit labels were bounded, not package-wide. The current package has 282 Test functions; all were enumerated. The current exact step-limit boundary row was added to regexp matrices.
- The full clean run timed out at90.168 seconds with no observed individual failure. A broader direct-runtime baseline passed; D1's first broad run cooked at90.060 seconds. Its failures do not prove uniqueness and are retained separately. The subsequent per-function matrices completed or are explicitly flagged.
- Go -coverprofile with -coverpkg=./internal/native,./internal/regexp measures compiler and build code, not embedded C run in child binaries. All requested row/subsumer profiles are saved, with exact exclusive Go lines. No C coverage claim is made. Runtime leads use semantic inputs and assertions instead.
- Seven mutation slots cannot supply three dedicated attempts for ten rows. They were spent on invalid-radix diagnostics, undefined equality, three step-limit contracts, raw UTF16 pattern identity and low-surrogate lastIndex. Rows with fewer than three dedicated attempts are cannot-judge, not not defended. They remain candidates to keep.
- Some rows differ in input despite sharing a checker. The random wrappers are one family. A family member is never treated as its subsumer.
- The supplied radix audit excerpt truncates commands. Complete prior report, row data, scopes, plans and limits were fetched and saved.
- Every complete matrix lists its current test names and passed rows. Decoder corpora, unrelated compiler products, WASI and other emitted-program consumers were not rerun under mutants. No package-wide unique kill is asserted beyond the bounded direct-consumer scope.
Name/assertion findings for rows not defended:
- TestRecordBenchmark measures and logs timing, validates five nonnegative non-NaN durations, and compares Node workload results. It asserts no performance threshold and accepts Infinity. Its name does not establish protection against a slowdown.
- TestRegExpNativeStepLimit checks catastrophic-backtracking interruption, exit70 and a diagnostic substring. It does not check exact instruction count. The newer boundary row guards that separate promise.
- Search, Lint, Test262 and Random compare results for their own supplied inputs, including captures/groups/lastIndex. No missing name promise was established; insufficient mutants cannot justify weakening them.
- RuntimeReleasePaths checks shared ownership, live counts and a100000-object chain with two sanitizer settings. Its assertions do address the named release behavior. No exclusive production defect was tested within the budget.
No test or oracle was edited. All source edits are restored before committing evidence.
