All four mutants remain caught outside the candidate family. Baselines passed in 608.6s and 493.3s; replays took 345.7s, 244.5s, 122.7s and 375.6s, stopping after ordinary catches. No stale diffs or Go test-binary panics occurred. Repository corpus opt-ins were enabled for D1/D3. Sources are restored; [evidence](https://github.com/system-inc/adamic/tree/8f96d2434d66b3dcc5ea47f07d491292e3aa9bdf/review/test-defend/deletion-set/stage1-cohere-typeaware) was pushed to the requested branch.

```json
{
  "package": "stage1/cohere/typeaware",
  "main": "b8bcadb2c493173855f19d7e5c508b34f5eeb5b6",
  "skipped": ["TestVolumeProfileCorpora family"],
  "mutants": [
    {
      "mutant": "M1",
      "file_line": "stage1/cohere/typeaware/shadow.ts:106",
      "branch": "test-audit/stage1-cohere-typeaware-volume_profile_Controls",
      "candidates_failed": ["TestVolumeProfileCorpora family"],
      "still_caught_by": [
        "TestVolumeAgreementAndMutants_014",
        "TestVolumeAgreementAndMutants_030"
      ],
      "stale": false
    },
    {
      "mutant": "M4",
      "file_line": "stage1/cohere/typeaware/diagnostic.ts:17",
      "branch": "test-audit/stage1-cohere-typeaware-volume_profile_Controls",
      "candidates_failed": ["TestVolumeProfileCorpora family"],
      "still_caught_by": ["TestSixRuleAgreementAndMutants_015"],
      "stale": false
    },
    {
      "mutant": "D1",
      "file_line": "stage1/cohere/typeaware/shadow.ts:70",
      "branch": "test-defend/stage1-cohere-typeaware-volume_profile_Controls",
      "candidates_failed": ["TestVolumeProfileCorpora family"],
      "still_caught_by": ["TestVolumeAgreementRepository_017"],
      "stale": false
    },
    {
      "mutant": "D3",
      "file_line": "stage1/cohere/typeaware/shadow.ts:66",
      "branch": "test-defend/stage1-cohere-typeaware-volume_profile_Controls",
      "candidates_failed": ["TestVolumeProfileCorpora family"],
      "still_caught_by": ["TestVolumeAgreementRepository_017"],
      "stale": false
    }
  ],
  "keep": [],
  "deletable": ["TestVolumeProfileCorpora family"],
  "notes": [
    "Decision is limited to the gathered mutants.",
    "The requested skip expression retains TestVolumeProfileCorporaUnion, which checks enumeration only."
  ]
}
```
