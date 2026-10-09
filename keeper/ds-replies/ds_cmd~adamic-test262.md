Replayed all 16 qualifying mutants on current main with the five candidates skipped and fresh build caches. Every mutant retained a catcher outside the set; all 39 retained tests completed each run. Baseline passed in 55.051 binary seconds; replays took 810.728 wall seconds total, ranging from 34.341–68.393 seconds. Nothing was stale or broken; no panics, pins-only catches, witness-only catches or narrowing occurred. Source and tests are restored. Evidence pushed to `test-defend/deletion-set/cmd-adamic-test262`, commit `e36f12a6`, under `review/test-defend/deletion-set/cmd-adamic-test262/`. “Deletable” applies only to this recorded mutant set.

```json
{
  "package": "cmd/adamic-test262",
  "main": "b8bcadb2c493173855f19d7e5c508b34f5eeb5b6",
  "skipped": [
    "TestAdaptLeavesCheckoutText",
    "TestCompilerStartupMeasurement",
    "TestNativeCacheGeneratedC",
    "TestNodeCacheProgram",
    "TestParallelCachedMatchesSerial"
  ],
  "mutants": [
    {
      "mutant": "M18",
      "file_line": "cmd/adamic-test262/classify.go:65",
      "branch": "test-audit/cmd-adamic-test262-adapt",
      "candidates_failed": ["TestParallelCachedMatchesSerial"],
      "still_caught_by": ["TestClassifyAdapt", "TestMiniRunner"],
      "stale": false
    },
    {
      "mutant": "M02",
      "file_line": "cmd/adamic-test262/cache.go:60",
      "branch": "test-audit/cmd-adamic-test262-cache",
      "candidates_failed": ["TestNodeCacheProgram"],
      "still_caught_by": ["TestCacheKeyDimensions"],
      "stale": false
    },
    {
      "mutant": "M03",
      "file_line": "cmd/adamic-test262/cache.go:63",
      "branch": "test-audit/cmd-adamic-test262-cache",
      "candidates_failed": ["TestNativeCacheGeneratedC"],
      "still_caught_by": ["TestCacheKeyDimensions"],
      "stale": false
    },
    {
      "mutant": "M05",
      "file_line": "cmd/adamic-test262/cache.go:35",
      "branch": "test-audit/cmd-adamic-test262-cache",
      "candidates_failed": ["TestParallelCachedMatchesSerial"],
      "still_caught_by": ["TestCacheBypass", "TestCompilerCacheAcrossScratchDirectories", "TestEditCacheSeparation", "TestLoweringSourceEdit", "TestCompilerWorkerMatchesSubprocess", "TestCacheAtomicBytes", "TestWorkerLazyFallback", "TestOrderedProgress"],
      "stale": false
    },
    {
      "mutant": "M01",
      "file_line": "cmd/adamic-test262/classify.go:90",
      "branch": "test-audit/cmd-adamic-test262-corpus",
      "candidates_failed": ["TestParallelCachedMatchesSerial"],
      "still_caught_by": ["TestRunFilterAttemptLimit", "TestClassifyCorpus", "TestOrderedProgress", "TestMiniRunner"],
      "stale": false
    },
    {
      "mutant": "M05",
      "file_line": "cmd/adamic-test262/verdict.go:58",
      "branch": "test-audit/cmd-adamic-test262-corpus",
      "candidates_failed": ["TestParallelCachedMatchesSerial"],
      "still_caught_by": ["TestCompilerCacheAcrossScratchDirectories", "TestEditCacheSeparation", "TestLoweringSourceEdit", "TestWorkerLazyFallback", "TestMiniRunner", "TestUnavailableCacheStillRuns", "TestVerdictCorpus", "TestOrderedProgress", "TestLargeCompilerOutputIsComplete"],
      "stale": false
    },
    {
      "mutant": "M15",
      "file_line": "cmd/adamic-test262/run.go:390",
      "branch": "test-audit/cmd-adamic-test262-corpus",
      "candidates_failed": ["TestCompilerStartupMeasurement", "TestNativeCacheGeneratedC", "TestNodeCacheProgram", "TestParallelCachedMatchesSerial"],
      "still_caught_by": ["TestProgramCPUDeadline", "TestCompilerCacheProgram", "TestUnavailableCacheStillRuns", "TestMiniRunner", "TestLargeCompilerOutputIsComplete", "TestRunProgramFractionalCPUBudget", "TestRunCommandExactCaptureCapacity", "TestImportedInputsRunFresh", "TestWorkerLazyFallback"],
      "stale": false
    },
    {
      "mutant": "D4",
      "file_line": "cmd/adamic-test262/adapt.go:35",
      "branch": "test-defend/cmd-adamic-test262-adapt",
      "candidates_failed": ["TestAdaptLeavesCheckoutText"],
      "still_caught_by": ["TestAdaptClassKeepsVars", "TestAdaptVarToLet"],
      "stale": false
    },
    {
      "mutant": "D2",
      "file_line": "cmd/adamic-test262/cache.go:35",
      "branch": "test-defend/cmd-adamic-test262-cache",
      "candidates_failed": ["TestParallelCachedMatchesSerial"],
      "still_caught_by": ["TestCompilerWorkerMatchesSubprocess", "TestCacheAtomicBytes"],
      "stale": false
    },
    {
      "mutant": "D02",
      "file_line": "cmd/adamic-test262/cache.go:60",
      "branch": "test-defend/cmd-adamic-test262-cache",
      "candidates_failed": ["TestNodeCacheProgram"],
      "still_caught_by": ["TestCacheKeyDimensions"],
      "stale": false
    },
    {
      "mutant": "D03",
      "file_line": "cmd/adamic-test262/cache.go:63",
      "branch": "test-defend/cmd-adamic-test262-cache",
      "candidates_failed": ["TestNativeCacheGeneratedC"],
      "still_caught_by": ["TestCacheKeyDimensions"],
      "stale": false
    },
    {
      "mutant": "D04",
      "file_line": "cmd/adamic-test262/run.go:379",
      "branch": "test-defend/cmd-adamic-test262-cache",
      "candidates_failed": ["TestNativeCacheGeneratedC", "TestNodeCacheProgram", "TestParallelCachedMatchesSerial"],
      "still_caught_by": ["TestProgramCPUDeadline", "TestCompilerCacheProgram"],
      "stale": false
    },
    {
      "mutant": "D05",
      "file_line": "cmd/adamic-test262/run.go:392",
      "branch": "test-defend/cmd-adamic-test262-cache",
      "candidates_failed": ["TestCompilerStartupMeasurement", "TestNativeCacheGeneratedC", "TestNodeCacheProgram", "TestParallelCachedMatchesSerial"],
      "still_caught_by": ["TestProgramCPUDeadline", "TestRunProgramFractionalCPUBudget", "TestCompilerCacheProgram", "TestWorkerLazyFallback", "TestLargeCompilerOutputIsComplete", "TestRunCommandExactCaptureCapacity", "TestImportedInputsRunFresh", "TestUnavailableCacheStillRuns", "TestMiniRunner"],
      "stale": false
    },
    {
      "mutant": "D05",
      "file_line": "cmd/adamic-test262/run.go:392",
      "branch": "test-defend/cmd-adamic-test262-corpus",
      "candidates_failed": ["TestCompilerStartupMeasurement", "TestNativeCacheGeneratedC", "TestNodeCacheProgram", "TestParallelCachedMatchesSerial"],
      "still_caught_by": ["TestProgramCPUDeadline", "TestRunProgramFractionalCPUBudget", "TestCompilerCacheProgram", "TestImportedInputsRunFresh", "TestWorkerLazyFallback", "TestRunCommandExactCaptureCapacity", "TestUnavailableCacheStillRuns", "TestLargeCompilerOutputIsComplete", "TestMiniRunner"],
      "stale": false
    },
    {
      "mutant": "D06",
      "file_line": "cmd/adamic-test262/run.go:451",
      "branch": "test-defend/cmd-adamic-test262-corpus",
      "candidates_failed": ["TestCompilerStartupMeasurement", "TestNativeCacheGeneratedC", "TestNodeCacheProgram", "TestParallelCachedMatchesSerial"],
      "still_caught_by": ["TestProgramCPUDeadline", "TestOutputOverflowIsReported", "TestRunProgramFractionalCPUBudget", "TestCompilerCacheProgram", "TestCompilerWorkerMatchesSubprocess", "TestImportedInputsRunFresh", "TestRunCommandExactCaptureCapacity", "TestUnavailableCacheStillRuns", "TestMiniRunner", "TestLargeCompilerOutputIsComplete"],
      "stale": false
    },
    {
      "mutant": "D07",
      "file_line": "cmd/adamic-test262/run.go:396",
      "branch": "test-defend/cmd-adamic-test262-corpus",
      "candidates_failed": ["TestCompilerStartupMeasurement", "TestParallelCachedMatchesSerial"],
      "still_caught_by": ["TestCompilerCacheAcrossScratchDirectories", "TestEditCacheSeparation", "TestLoweringSourceEdit", "TestProgramCPUDeadline", "TestWorkerLazyFallback", "TestImportedInputsRunFresh", "TestMiniRunner", "TestLargeCompilerOutputIsComplete", "TestRunProgramFractionalCPUBudget", "TestRunCommandExactCaptureCapacity", "TestUnavailableCacheStillRuns", "TestOrderedProgress"],
      "stale": false
    }
  ],
  "keep": [],
  "deletable": [
    "TestAdaptLeavesCheckoutText",
    "TestCompilerStartupMeasurement",
    "TestNativeCacheGeneratedC",
    "TestNodeCacheProgram",
    "TestParallelCachedMatchesSerial"
  ]
}
```
