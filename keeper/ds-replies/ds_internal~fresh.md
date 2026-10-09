The skipped-candidate baseline passed in 15.6 seconds; all three replays remained caught outside the deletion set, taking 20.3, 19.6 and 20.2 seconds. No diffs were stale or runs panicked. The missing audit branch was located as `test-audit/fresh` through the defender report. Sources are restored; complete logs and matrix are pushed to `test-defend/deletion-set/internal-fresh` under `review/test-defend/deletion-set/internal-fresh/`.

```json
{
  "package": "internal/fresh",
  "main": "b8bcadb2c493173855f19d7e5c508b34f5eeb5b6",
  "skipped": ["TestNodeFSFileDoesNotEscapeBorrowedObjects"],
  "mutants": [
    {
      "mutant": "M2",
      "file_line": "internal/fresh/node_fs_file.go:9",
      "branch": "test-audit/fresh",
      "candidates_failed": ["TestNodeFSFileDoesNotEscapeBorrowedObjects"],
      "still_caught_by": [
        "TestFreshWrites____oracle_testdata_node_fs_file_close_a_93025b0d712c",
        "TestFreshWrites____oracle_testdata_node_fs_file_date_a_e4096c088001",
        "TestFreshWrites____oracle_testdata_node_fs_file_mkdir_a_b27dc23fc728",
        "TestFreshWrites____oracle_testdata_node_fs_file_write_buffer_a_bab39c7b6026",
        "TestFreshWrites____oracle_testdata_node_fs_file_write_file_a_1f61ee3992e2",
        "TestFreshWrites____oracle_testdata_unknown_narrowing_host_a_e5af072b53a5",
        "TestNodeFSFileOperationsAreKnown",
        "TestNodeFSFileResultsAreFresh",
        "TestStatAtimeSelfCycleRemainsUnproven"
      ],
      "stale": false
    },
    {
      "mutant": "M18",
      "file_line": "internal/fresh/fresh.go:54",
      "branch": "test-audit/fresh",
      "candidates_failed": ["TestNodeFSFileDoesNotEscapeBorrowedObjects"],
      "still_caught_by": [
        "TestClassMethodOutsideSelfCycleRemainsUnproven",
        "TestFreshWrites____flow_testdata_mutations_a_5514557d1b47",
        "TestFutureBufferOperationRemainsUnknown",
        "TestFutureRegexMethodRemainsUnknown",
        "TestMethodKeepsArgument",
        "TestNodeBufferHashUpdateKeepsAlias",
        "TestNodeFSFileOperandsStillJudgeWrites",
        "TestNodeFSFileResultsAreFresh",
        "TestRegexOperandsStillJudgeCycleWrites",
        "TestRegexOperationsDoNotPoisonTreeWrites",
        "TestReplacementCallbackEscapeRemainsUnproven",
        "TestStatAtimeSelfCycleRemainsUnproven"
      ],
      "stale": false
    },
    {
      "mutant": "D4",
      "file_line": "internal/fresh/fresh.go:1333",
      "branch": "test-defend/internal-fresh",
      "candidates_failed": ["TestNodeFSFileDoesNotEscapeBorrowedObjects"],
      "still_caught_by": [
        "TestFreshWrites____oracle_testdata_node_fs_file_close_a_93025b0d712c",
        "TestFreshWrites____oracle_testdata_node_fs_file_date_a_e4096c088001",
        "TestFreshWrites____oracle_testdata_node_fs_file_mkdir_a_b27dc23fc728",
        "TestFreshWrites____oracle_testdata_node_fs_file_write_buffer_a_bab39c7b6026",
        "TestFreshWrites____oracle_testdata_node_fs_file_write_file_a_1f61ee3992e2",
        "TestFreshWrites____oracle_testdata_unknown_narrowing_host_a_e5af072b53a5",
        "TestNodeFSFileOperationsAreKnown",
        "TestNodeFSFileResultsAreFresh",
        "TestStatAtimeSelfCycleRemainsUnproven"
      ],
      "stale": false
    }
  ],
  "keep": [],
  "deletable": ["TestNodeFSFileDoesNotEscapeBorrowedObjects"]
}
```
