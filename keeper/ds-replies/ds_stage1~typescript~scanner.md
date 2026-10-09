Replayed seven mutants with both candidates skipped. Baseline passed in 22.0 test-binary seconds; replays totaled 195.7 wall seconds. Five mutants lost their last catchers; two remain caught by `TestPerformance`. Nothing was stale, broken or panicking. Tree restored; evidence pushed to `test-defend/deletion-set/stage1-typescript-scanner` under `review/test-defend/deletion-set/stage1-typescript-scanner/`.

```json
{
  "package": "stage1/typescript/scanner",
  "main": "7b9d4272c28f59530ab13daa5c49067e47933b06",
  "skipped": [
    "TestProfileSnapshotsAgree",
    "TestScannerAgreesWithTypescriptGo family"
  ],
  "mutants": [
    {
      "mutant": "M3",
      "file_line": "stage1/typescript/scanner/characters.ts:250",
      "branch": "test-audit/stage1-typescript-scanner",
      "candidates_failed": [
        "TestProfileSnapshotsAgree",
        "TestScannerAgreesWithTypescriptGo family"
      ],
      "still_caught_by": [],
      "stale": false
    },
    {
      "mutant": "M1",
      "file_line": "stage1/typescript/scanner/scanner.ts:74",
      "branch": "test-audit/stage1-typescript-scanner",
      "candidates_failed": [
        "TestProfileSnapshotsAgree",
        "TestScannerAgreesWithTypescriptGo family"
      ],
      "still_caught_by": [],
      "stale": false
    },
    {
      "mutant": "D02",
      "file_line": "stage1/typescript/scanner/scanner.ts:857",
      "branch": "test-defend/stage1-typescript-scanner",
      "candidates_failed": [
        "TestProfileSnapshotsAgree",
        "TestScannerAgreesWithTypescriptGo family"
      ],
      "still_caught_by": [],
      "stale": false
    },
    {
      "mutant": "D01",
      "file_line": "stage1/typescript/scanner/scanner.ts:1009",
      "branch": "test-defend/stage1-typescript-scanner",
      "candidates_failed": [
        "TestProfileSnapshotsAgree",
        "TestScannerAgreesWithTypescriptGo family"
      ],
      "still_caught_by": [],
      "stale": false
    },
    {
      "mutant": "D03",
      "file_line": "stage1/typescript/scanner/scanner.ts:836",
      "branch": "test-defend/stage1-typescript-scanner",
      "candidates_failed": [
        "TestProfileSnapshotsAgree",
        "TestScannerAgreesWithTypescriptGo family"
      ],
      "still_caught_by": [],
      "stale": false
    },
    {
      "mutant": "D3",
      "file_line": "internal/native/runtime/heap.c:149",
      "branch": "test-defend/stage1-typescript-scanner",
      "candidates_failed": ["TestProfileSnapshotsAgree"],
      "still_caught_by": ["TestPerformance"],
      "stale": false
    },
    {
      "mutant": "D4",
      "file_line": "internal/native/runtime/heap.c:152",
      "branch": "test-defend/stage1-typescript-scanner",
      "candidates_failed": ["TestProfileSnapshotsAgree"],
      "still_caught_by": ["TestPerformance"],
      "stale": false
    }
  ],
  "keep": [
    {
      "test": "TestProfileSnapshotsAgree",
      "because": "M1, M3, D01, D02 and D03 lose their last catchers without the set. Preserve at least one candidate."
    },
    {
      "test": "TestScannerAgreesWithTypescriptGo family",
      "because": "M1, M3, D01, D02 and D03 lose their last catchers without the set. Preserve at least one candidate."
    }
  ],
  "deletable": [],
  "notes": [
    "Prior evidence shows either candidate catches all five lost mutants. This replay establishes collective coverage loss, not that both candidates must remain.",
    "Remaining profile and performance tests were enabled. Every replay completed.",
    "No witness-only failures or package panics occurred."
  ]
}
```
