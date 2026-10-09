With both candidates skipped, the clean baseline passed in 30.6 seconds; all nine mutant replays completed in 30.2–30.9 seconds each, totaling 306.1 seconds including baseline. Every mutant retained a non-witness catcher; none was stale or panicked. Witness failures were excluded. I corrected the skip regex to match numbered family members; `matrix-selector.txt` was a file, not a branch. [Evidence](https://github.com/system-inc/adamic/tree/c08a7410218d01a332d12b18635c878774dd73bf/review/test-defend/deletion-set/internal-unicodeproperties) is pushed to the requested branch, and the clean detached main checkout is restored. “Deletable” applies only to this gathered mutant set.

```json
{
  "package": "internal/unicodeproperties",
  "main": "b8bcadb2c493173855f19d7e5c508b34f5eeb5b6",
  "skipped": [
    "TestCanonicalizeUnicodeNodeRange family",
    "TestNodeStringProperties"
  ],
  "mutants": [
    {
      "mutant": "M11",
      "file_line": "internal/unicodeproperties/canonicalize_tables.go:3022",
      "branch": "test-audit/internal-unicodeproperties-canonicalize",
      "candidates_failed": ["TestCanonicalizeUnicodeNodeRange family"],
      "still_caught_by": ["TestEquivalentsAreClosed", "TestUnicodeNodeShardCoverage"],
      "stale": false
    },
    {
      "mutant": "M05",
      "file_line": "internal/unicodeproperties/unicodeproperties.go:175",
      "branch": "test-audit/internal-unicodeproperties-node",
      "candidates_failed": ["TestNodeStringProperties"],
      "still_caught_by": ["TestKnownMembership", "TestRejectedNames", "TestStringPropertyCensus"],
      "stale": false
    },
    {
      "mutant": "M13",
      "file_line": "internal/unicodeproperties/tables.go:25712",
      "branch": "test-audit/internal-unicodeproperties-node",
      "candidates_failed": ["TestNodeStringProperties"],
      "still_caught_by": ["TestKnownMembership", "TestStringPropertyCensus"],
      "stale": false
    },
    {
      "mutant": "D2",
      "file_line": "internal/unicodeproperties/canonicalize_tables.go:22",
      "branch": "test-defend/internal-unicodeproperties-canonicalize",
      "candidates_failed": ["TestCanonicalizeUnicodeNodeRange family"],
      "still_caught_by": ["TestUnicodeNodeShardCoverage"],
      "stale": false
    },
    {
      "mutant": "D3",
      "file_line": "internal/unicodeproperties/canonicalize_tables.go:23",
      "branch": "test-defend/internal-unicodeproperties-canonicalize",
      "candidates_failed": ["TestCanonicalizeUnicodeNodeRange family"],
      "still_caught_by": ["TestUnicodeNodeShardCoverage"],
      "stale": false
    },
    {
      "mutant": "D4",
      "file_line": "internal/unicodeproperties/canonicalize_tables.go:24",
      "branch": "test-defend/internal-unicodeproperties-canonicalize",
      "candidates_failed": ["TestCanonicalizeUnicodeNodeRange family"],
      "still_caught_by": ["TestUnicodeNodeShardCoverage"],
      "stale": false
    },
    {
      "mutant": "D1",
      "file_line": "internal/unicodeproperties/tables.go:25713",
      "branch": "test-defend/internal-unicodeproperties-node",
      "candidates_failed": ["TestNodeStringProperties"],
      "still_caught_by": ["TestStringPropertyCensus"],
      "stale": false
    },
    {
      "mutant": "D2",
      "file_line": "internal/unicodeproperties/tables.go:27119",
      "branch": "test-defend/internal-unicodeproperties-node",
      "candidates_failed": ["TestNodeStringProperties"],
      "still_caught_by": ["TestStringPropertyCensus"],
      "stale": false
    },
    {
      "mutant": "D3",
      "file_line": "internal/unicodeproperties/tables.go:28061",
      "branch": "test-defend/internal-unicodeproperties-node",
      "candidates_failed": ["TestNodeStringProperties"],
      "still_caught_by": ["TestStringPropertyCensus"],
      "stale": false
    }
  ],
  "keep": [],
  "deletable": [
    "TestCanonicalizeUnicodeNodeRange family",
    "TestNodeStringProperties"
  ]
}
```
