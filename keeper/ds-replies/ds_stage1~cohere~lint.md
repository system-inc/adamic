All three mutants remain caught with both candidates skipped. Missing `ADAMIC_TYPESCRIPT_SOURCE` required narrowing to 29 recorded catchers: baseline passed in 100.1 seconds; replays took 104.5, 105.5 and 12.5 seconds. No diffs were stale, and sources are restored. Evidence was [pushed](https://github.com/system-inc/adamic/tree/9ab382220d04ffd79796f4d14538ab7dfb4de220/review/test-defend/deletion-set/stage1-cohere-lint/session-b8bcadb2-20261009T1807).

```json
{
  "package": "stage1/cohere/lint",
  "main": "b8bcadb2c493173855f19d7e5c508b34f5eeb5b6",
  "skipped": [
    "TestNodeTableIsLinkOnlyFamily",
    "TestProfileCompilationBuildLower"
  ],
  "mutants": [
    {
      "mutant": "lint-D1",
      "file_line": "stage1/cohere/lint/lint.ts:83",
      "branch": "test-defend/stage1-cohere-lint-lint",
      "candidates_failed": ["TestNodeTableIsLinkOnlyFamily"],
      "still_caught_by": [
        "TestOwnedWitnesses_000",
        "TestOwnedWitnesses_001",
        "TestOwnedWitnesses_002",
        "TestOwnedWitnesses_003",
        "TestOwnedWitnesses_004",
        "TestOwnedWitnesses_005",
        "TestOwnedWitnesses_006",
        "TestOwnedWitnesses_007",
        "TestOwnedWitnesses_008",
        "TestOwnedWitnesses_009",
        "TestOwnedWitnesses_010",
        "TestOwnedWitnesses_011",
        "TestOwnedWitnesses_012",
        "TestOwnedWitnesses_013",
        "TestOwnedWitnesses_014",
        "TestOwnedWitnesses_015",
        "TestProfileCompilation_000"
      ],
      "stale": false
    },
    {
      "mutant": "lint-D2",
      "file_line": "stage1/cohere/lint/lint.ts:202",
      "branch": "test-defend/stage1-cohere-lint-lint",
      "candidates_failed": ["TestNodeTableIsLinkOnlyFamily"],
      "still_caught_by": ["TestOwnedWitnesses_004"],
      "stale": false
    },
    {
      "mutant": "profile-D1",
      "file_line": "internal/lower/lower.go:22",
      "branch": "test-defend/stage1-cohere-lint-profile_compilation_main",
      "candidates_failed": ["TestProfileCompilationBuildLower"],
      "still_caught_by": ["TestProduct_ProfileCompilationLowered"],
      "stale": false
    }
  ],
  "keep": [],
  "deletable": [
    "TestNodeTableIsLinkOnlyFamily",
    "TestProfileCompilationBuildLower"
  ],
  "bounded": true,
  "notes": [
    "Tests outside the 29-row narrowed selection remain unknown.",
    "The virtual NodeTable family name was expanded to its eight actual shard tests for skipping.",
    "Planted-failure failures were excluded as catchers.",
    "lint-D1 native child panics did not abort the Go binary; all 29 selected tests completed.",
    "lint-D2 and profile-D1 stopped after their first clean non-witness failure."
  ]
}
```
