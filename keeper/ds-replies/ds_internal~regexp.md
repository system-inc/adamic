With all three candidates skipped, the clean baseline passed and all 13 gathered mutants remained caught. Replay wall time totaled 20.13s; the restored baseline passed in 1.22s wall time. No diffs were stale or failed to build. D08 panicked in `TestMatcherOct6Mutants`, aborting that run; only its observed failure is listed. Production and tests are restored. The matrix, inventory and logs were pushed to `test-defend/deletion-set/internal-regexp` under `review/test-defend/deletion-set/internal-regexp/`, commit `04b7d5c5`.

```json
{
  "package": "internal/regexp",
  "main": "4c075a43e2216c335fe70efbf707f52fdb141d3c",
  "skipped": ["TestFlags", "TestMatcherStepLimit", "TestQuantifierBounds"],
  "mutants": [
    {
      "mutant": "M11",
      "file_line": "internal/regexp/canonicalize.go:13",
      "branch": "test-audit/internal-regexp-matcher_oracle",
      "candidates_failed": ["TestMatcherStepLimit"],
      "still_caught_by": ["TestMatcherCanonicalizeNode", "TestMatcherNodeControls", "TestMatcherOct6LoopsNode", "TestMatcherOct6Mutants", "TestMatcherOct6Node", "TestMatcherRandomNode", "TestMatcherStepLimitBoundary", "TestMatcherTest262Executions", "TestMatcherUTF16PatternsNode"],
      "stale": false
    },
    {
      "mutant": "M14",
      "file_line": "internal/regexp/sets.go:293",
      "branch": "test-audit/internal-regexp-matcher_oracle",
      "candidates_failed": ["TestMatcherStepLimit"],
      "still_caught_by": ["TestMatcherCanonicalizeNode", "TestMatcherNodeControls", "TestMatcherOct6LoopsNode", "TestMatcherOct6Mutants", "TestMatcherOct6Node", "TestMatcherPropertyProviderStrings", "TestMatcherRandomNode", "TestMatcherStepLimitBoundary", "TestMatcherTest262Executions", "TestMatcherUTF16PatternsNode"],
      "stale": false
    },
    {
      "mutant": "M15",
      "file_line": "internal/regexp/parser.go:110",
      "branch": "test-audit/internal-regexp-matcher_oracle",
      "candidates_failed": ["TestFlags"],
      "still_caught_by": ["TestMatcherCanonicalizeNode", "TestMatcherNodeControls", "TestMatcherOct6LoopsNode", "TestMatcherOct6Node", "TestMatcherPropertyProviderStrings", "TestMatcherTest262Executions", "TestMatcherUTF16PatternsNode", "TestNodeAgreement", "TestParse", "TestUnicodeDecimalEscape"],
      "stale": false
    },
    {
      "mutant": "M1",
      "file_line": "internal/regexp/parser.go:88",
      "branch": "test-audit/internal-regexp-parser",
      "candidates_failed": ["TestFlags"],
      "still_caught_by": ["TestNodeAgreement"],
      "stale": false
    },
    {
      "mutant": "M2",
      "file_line": "internal/regexp/parser.go:109",
      "branch": "test-audit/internal-regexp-parser",
      "candidates_failed": ["TestFlags"],
      "still_caught_by": ["TestNodeAgreement"],
      "stale": false
    },
    {
      "mutant": "M4",
      "file_line": "internal/regexp/parser.go:329",
      "branch": "test-audit/internal-regexp-parser",
      "candidates_failed": ["TestQuantifierBounds"],
      "still_caught_by": ["TestMatcherNodeControls", "TestMatcherOct6LoopsNode", "TestMatcherOct6Mutants", "TestMatcherOct6Node", "TestMatcherRandomNode", "TestMatcherTest262Executions", "TestMatcherUTF16PatternsNode"],
      "stale": false
    },
    {
      "mutant": "D4",
      "file_line": "internal/regexp/matcher.go:485",
      "branch": "test-defend/internal-regexp-matcher_oracle",
      "candidates_failed": ["TestMatcherStepLimit"],
      "still_caught_by": ["TestMatcherNodeControls", "TestMatcherOct6Mutants", "TestMatcherOct6Node", "TestMatcherRandomNode", "TestMatcherTest262Executions"],
      "stale": false
    },
    {
      "mutant": "D04",
      "file_line": "internal/regexp/parser.go:88",
      "branch": "test-defend/internal-regexp-parser",
      "candidates_failed": ["TestFlags"],
      "still_caught_by": ["TestNodeAgreement"],
      "stale": false
    },
    {
      "mutant": "D05",
      "file_line": "internal/regexp/parser.go:109",
      "branch": "test-defend/internal-regexp-parser",
      "candidates_failed": ["TestFlags"],
      "still_caught_by": ["TestNodeAgreement"],
      "stale": false
    },
    {
      "mutant": "D06",
      "file_line": "internal/regexp/parser.go:107",
      "branch": "test-defend/internal-regexp-parser",
      "candidates_failed": ["TestFlags"],
      "still_caught_by": ["TestNodeAgreement"],
      "stale": false
    },
    {
      "mutant": "D07",
      "file_line": "internal/regexp/parser.go:317",
      "branch": "test-defend/internal-regexp-parser",
      "candidates_failed": ["TestQuantifierBounds"],
      "still_caught_by": ["TestMatcherRandomNode", "TestMatcherTest262Executions", "TestNodeAgreement", "TestParse"],
      "stale": false
    },
    {
      "mutant": "D08",
      "file_line": "internal/regexp/parser.go:330",
      "branch": "test-defend/internal-regexp-parser",
      "candidates_failed": ["TestMatcherStepLimit", "TestQuantifierBounds"],
      "still_caught_by": ["TestMatcherOct6Mutants"],
      "stale": false
    },
    {
      "mutant": "D09",
      "file_line": "internal/regexp/parser.go:329",
      "branch": "test-defend/internal-regexp-parser",
      "candidates_failed": ["TestQuantifierBounds"],
      "still_caught_by": ["TestMatcherNodeControls", "TestMatcherOct6LoopsNode", "TestMatcherOct6Mutants", "TestMatcherOct6Node", "TestMatcherRandomNode", "TestMatcherTest262Executions", "TestMatcherUTF16PatternsNode"],
      "stale": false
    }
  ],
  "keep": [],
  "deletable": ["TestFlags", "TestMatcherStepLimit", "TestQuantifierBounds"]
}
```
