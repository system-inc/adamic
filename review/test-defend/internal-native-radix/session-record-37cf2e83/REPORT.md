TestRecordBenchmark was not defended after three aimed attempts.
The answer-preserving cost mutant passed; two semantic mutants also failed TestRecordsAgainstNode.
Evidence is bounded to 60 current top-level functions; no test deletion is recommended.

[
  {
    "test": "TestRecordBenchmark",
    "package": "internal/native",
    "prior_verdict": "subsumed",
    "subsumed_by": "TestRuntimeStringEquality",
    "defense": "not defended",
    "unique_mutant": null,
    "attempts": [
      {
        "mutant": "D01",
        "file_line": "internal/native/runtime/map.c:72",
        "change": "Repeat string hashing eight times, increasing work while retaining consistent equal-key hashes; cost-rule repeat-loop attempt",
        "rows_failed": []
      },
      {
        "mutant": "D02",
        "file_line": "internal/native/runtime/record.c:207",
        "change": "Off-by-one iterator exhaustion bound drops the last snapshot key",
        "rows_failed": [
          "TestRecordBenchmark",
          "TestRecordsAgainstNode"
        ]
      },
      {
        "mutant": "D03",
        "file_line": "internal/native/runtime/map.c:222",
        "change": "Drop deletion count decrement",
        "rows_failed": [
          "TestRecordBenchmark",
          "TestRecordsAgainstNode"
        ]
      }
    ],
    "evidence": "ADAMIC_RECORD_BENCH=1 ADAMIC_BUILD_CACHE_DIR=/tmp/record-defense/cache/<D01-D03> timeout 330 go test -json -count=1 -timeout 300s ./internal/native/ -run \"^(TestRegExp.*|TestRecord.*|TestRuntimeStringEquality|TestMapHash.*|TestLibraryMapSetIteratorResources)$\"; see matrix.json and D01-D03.log for exact observed failures; D02 and D03: record_test.go:341: benchmark work differs from Node; TestRecordsAgainstNode: record_test.go:83: Node comparison differs.",
    "bounded": true
  }
]

Code under test and oracle are named in code-and-oracle.md. Coverage commands and full outputs are benchmark-coverage.log and subsumer-coverage.log; coverage-difference.json compares the profiles. Standalone D01.diff, D02.diff and D03.diff apply to starting origin/main 37cf2e8362d65a1a4196426ff38fe7020de5cac6. Each changed C unit passed its compile check; matrix builds used the runtime's normal compiler flags. D01's additional loop is explicitly allowed by the cost-row instruction.

| Mutant | Origin file:line | Change | Failed rows |
| --- | --- | --- | --- |
| D01 | internal/native/runtime/map.c:72 | Repeat string hashing eight times | None |
| D02 | internal/native/runtime/record.c:207 | Stop record iteration one key early | TestRecordBenchmark, TestRecordsAgainstNode |
| D03 | internal/native/runtime/map.c:222 | Drop deletion count decrement | TestRecordBenchmark, TestRecordsAgainstNode |

D01 survivor witness: the production hash loop executes eight passes rather than one. The benchmark's entire work checksum continues to agree with Node at every tested size and round. It passed at 51.49 s versus baseline 30.66 s. Those times are observations under different shared-worker load, not controlled performance evidence. Its checks do not impose a cost threshold.

D02 failure: record_test.go:341: benchmark work differs from Node. Native size1000 output was 748501 1000 499 500; Node was 749500 1000 500 500.
D03 failure: record_test.go:341: benchmark work differs from Node. Native size1000 output was 749500 1000 500 1000; Node was 749500 1000 500 500.
Both also caused record_test.go:83: Node comparison differs in TestRecordsAgainstNode. The prior named subsumer TestRuntimeStringEquality passed all three attempts. Complete pass/fail function lists are in matrix.json; all 60 top-level functions reached a terminal result in each mutant run, with no unknown rows or skips. Failures occurred in fixture subprocesses, not a panic aborting the Go test binary.

Timing: warm setup skipped setup.sh, nproc=5. npm ci output is npm-ci.log. Initial clean bounded baseline was 78.584 binary seconds. Per-test coverage commands were 32.794 s (benchmark) and 0.546 s (subsumer). Native rebuilds are included in the binary times; they were not separately instrumented. Matrix command wall times and binary times:
- D01: 193.261 wall seconds, 182.556 binary seconds.
- D02: 134.406 wall seconds, 116.972 binary seconds.
- D03: 168.470 wall seconds, 151.061 binary seconds.

Final clean added-neighbor baseline: see neighbors-clean.log. Production source restoration and all three git apply --check validations succeeded.

Every instruction ambiguity, setup friction and owner finding is recorded in friction-and-limits.md. The benchmark explicitly observes timings without a pass threshold; if intended as a performance regression guard, that promise is absent from the assertions. No full-package uniqueness, controlled timing comparison, dynamic C coverage or tests outside the bounded selector are claimed.

Final clean added-neighbor run passed in 120.568 binary seconds, with no failures or skips.
