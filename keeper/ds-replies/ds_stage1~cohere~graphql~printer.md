The candidate-skipped baseline passed in 53.9 seconds wall time. All seven qualifying mutants retained non-witness catchers; replays took 240.6 seconds total and stopped after clean catches. No stale diffs or panics occurred. Production sources and tests are restored. The corrected inventory, matrix, and logs are [published](/workspace/adamic/review/test-defend/deletion-set/stage1-cohere-graphql-printer/matrix.json) on `test-defend/deletion-set/stage1-cohere-graphql-printer`, commit `c95fcd2c`. Both candidates are deletable under this recorded-mutant criterion.

```json
{
  "package": "stage1/cohere/graphql/printer",
  "main": "b8bcadb2c493173855f19d7e5c508b34f5eeb5b6",
  "skipped": [
    "TestPrinterThroughput",
    "TestPrinterWhitespaceGap family"
  ],
  "mutants": [
    {
      "mutant": "M1",
      "file_line": "stage1/cohere/graphql/printer/printer.ts:436",
      "branch": "test-audit/stage1-cohere-graphql-printer-grain_mutant",
      "candidates_failed": ["TestPrinterThroughput"],
      "still_caught_by": ["TestPrinterAsGoCohere_003"],
      "stale": false
    },
    {
      "mutant": "M2",
      "file_line": "stage1/cohere/graphql/printer/printer.ts:47",
      "branch": "test-audit/stage1-cohere-graphql-printer-grain_mutant",
      "candidates_failed": ["TestPrinterThroughput"],
      "still_caught_by": ["TestPrinterAsGoCohere_003"],
      "witness_failures": ["TestPrinterShardPlantedDisagreement"],
      "stale": false
    },
    {
      "mutant": "M3",
      "file_line": "stage1/cohere/graphql/printer/doc.ts:183",
      "branch": "test-audit/stage1-cohere-graphql-printer-grain_mutant",
      "candidates_failed": ["TestPrinterThroughput"],
      "still_caught_by": ["TestPrinterAsGoCohere_002"],
      "witness_failures": ["TestPrinterShardPlantedDisagreement"],
      "stale": false
    },
    {
      "mutant": "M4",
      "file_line": "stage1/cohere/graphql/printer/doc.ts:94",
      "branch": "test-audit/stage1-cohere-graphql-printer-grain_mutant",
      "candidates_failed": ["TestPrinterThroughput"],
      "still_caught_by": ["TestPrinterAsGoCohere_003"],
      "stale": false
    },
    {
      "mutant": "M04",
      "file_line": "stage1/cohere/graphql/parser.ts:1228",
      "branch": "test-audit/stage1-cohere-graphql-printer-preflight_units",
      "candidates_failed": ["TestPrinterWhitespaceGap family"],
      "still_caught_by": [
        "TestPrinterAsGoCohere_000",
        "TestPrinterAsGoCohere_001",
        "TestPrinterAsGoCohere_002",
        "TestPrinterAsGoCohere_003"
      ],
      "witness_failures": ["TestPrinterShardPlantedDisagreement"],
      "stale": false
    },
    {
      "mutant": "D01",
      "file_line": "stage1/cohere/graphql/parser.ts:226",
      "branch": "test-defend/stage1-cohere-graphql-printer-preflight_units",
      "candidates_failed": ["TestPrinterWhitespaceGap family"],
      "still_caught_by": ["TestPrinterAsGoCohere_000"],
      "stale": false
    },
    {
      "mutant": "D02",
      "file_line": "stage1/cohere/graphql/lexer.ts:74",
      "branch": "test-defend/stage1-cohere-graphql-printer-preflight_units",
      "candidates_failed": ["TestPrinterWhitespaceGap family"],
      "still_caught_by": ["TestPrinterAsGoCohere_000"],
      "stale": false
    }
  ],
  "keep": [],
  "deletable": [
    "TestPrinterThroughput",
    "TestPrinterWhitespaceGap family"
  ]
}
```
