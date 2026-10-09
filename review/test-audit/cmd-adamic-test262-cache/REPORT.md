u012: 14 requested rows present at origin/main 8171b3173bdbfce1f7982d3c4f731279307ece37; none moved or vanished.
Clean whole-package baseline passed in 42.720 binary seconds; nproc=5; warm toolchain, setup skipped.
20 vetted production mutants ran against all 37 top-level package tests: 18 killed, 2 survived.
Verdicts: 8 sacred, 5 subsumed, 1 subprocess helper; all 13 audited rows failed their direct-entry probes.
Evidence: test-audit/cmd-adamic-test262-cache, review/test-audit/cmd-adamic-test262-cache/.

```json
[
  {
    "test": "TestNodeCacheProgram",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/cache_test.go:17",
    "seconds": 0.129,
    "oracle": "Node executes the program; literal word and cache invocation counts are handwritten.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M02"
    ],
    "unique_kills": [],
    "last_proven_fail": "M02: cache_test.go:31: changed Node program served \"before\\n\", want \"after\\n\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestCacheKeyDimensions"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "PReuse",
      "PNodeKey"
    ],
    "subsumer_seconds": 0.008,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT=M02 ADAMIC_BUILD_CACHE_DIR=/tmp/u012/cache/M02 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M02.log 2>&1; cache_test.go:31: changed Node program served \"before\\n\", want \"after\\n\"",
    "subsumption_kills": 1
  },
  {
    "test": "TestNativeCacheGeneratedC",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/cache_test.go:36",
    "seconds": 0.124,
    "oracle": "clang runs generated C; literal word and cache invocation counts are handwritten.",
    "oracle_kind": "self",
    "kills": [
      "M03"
    ],
    "unique_kills": [],
    "last_proven_fail": "M03: cache_test.go:50: changed emitted C served \"before\\n\", want \"after\\n\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestCacheKeyDimensions"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "PReuse",
      "PNativeKey"
    ],
    "subsumer_seconds": 0.008,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT=M03 ADAMIC_BUILD_CACHE_DIR=/tmp/u012/cache/M03 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M03.log 2>&1; cache_test.go:50: changed emitted C served \"before\\n\", want \"after\\n\"",
    "subsumption_kills": 1
  },
  {
    "test": "TestCacheKeyDimensions",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/cache_test.go:55",
    "seconds": 0.008,
    "oracle": "Handwritten key equality, dimension separation, and length-boundary assertions.",
    "oracle_kind": "self",
    "kills": [
      "M02",
      "M03",
      "M04"
    ],
    "unique_kills": [
      "M04"
    ],
    "last_proven_fail": "M04: cache_test.go:79: key boundaries lost",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "PReuse",
      "PNodeKey",
      "PNativeKey",
      "PKey"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT=M04 ADAMIC_BUILD_CACHE_DIR=/tmp/u012/cache/M04 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M04.log 2>&1; cache_test.go:79: key boundaries lost"
  },
  {
    "test": "TestCacheAtomicBytes",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/cache_test.go:83",
    "seconds": 0.008,
    "oracle": "Handwritten exact stdout/stderr bytes, status, reuse count, and corruption rejection.",
    "oracle_kind": "self",
    "kills": [
      "M05",
      "M06",
      "M07",
      "M08",
      "M10"
    ],
    "unique_kills": [
      "M06"
    ],
    "last_proven_fail": "M06: cache_test.go:106: corrupt evidence reused",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "PReuse"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT=M06 ADAMIC_BUILD_CACHE_DIR=/tmp/u012/cache/M06 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M06.log 2>&1; cache_test.go:106: corrupt evidence reused"
  },
  {
    "test": "TestCacheBypass",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/cache_test.go:110",
    "seconds": 0.008,
    "oracle": "Handwritten invocation counts, bypass environment and transient-result rules.",
    "oracle_kind": "self",
    "kills": [
      "M05",
      "M07",
      "M08",
      "M09",
      "M10"
    ],
    "unique_kills": [],
    "last_proven_fail": "M10: cache_test.go:121: cache missing",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestEditCacheSeparation"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "PObserve"
    ],
    "subsumer_seconds": 1.483,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT=M10 ADAMIC_BUILD_CACHE_DIR=/tmp/u012/cache/M10 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M10.log 2>&1; cache_test.go:121: cache missing",
    "subsumption_kills": 5
  },
  {
    "test": "TestParallelCachedMatchesSerial",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/cache_test.go:135",
    "seconds": 4.445,
    "oracle": "Node/native agreement plus handwritten serial/parallel report equality. M17 extra attempt and M19 doubled pass counts preserve parity.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M05"
    ],
    "unique_kills": [],
    "last_proven_fail": "M05: cache_test.go:175: round 0 differs:",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestCacheAtomicBytes"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "PFilter"
    ],
    "subsumer_seconds": 0.008,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT=M05 ADAMIC_BUILD_CACHE_DIR=/tmp/u012/cache/M05 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M05.log 2>&1; cache_test.go:175: round 0 differs:",
    "subsumption_kills": 1
  },
  {
    "test": "TestOrderedProgress",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/cache_test.go:196",
    "seconds": 3.244,
    "oracle": "Node/native agreement plus handwritten progress lines and report counts.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M05",
      "M18",
      "M19"
    ],
    "unique_kills": [
      "M18"
    ],
    "last_proven_fail": "M18: cache_test.go:233: ordered progress: \" 49/51 pass=0 fail=0 refused=0 crashed=0 skipped=49\\n 51/51 pass=1 fail=0 refused=0 crashed=0 skipped=50\\n\", want \" 50/51 pass=0 fail=0 refused=0 crashed=0 skipped=50\\n 51/51 pass=1 fail=0 refused=0 crashed=0 skipped=50\\n\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "PFilter"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT=M18 ADAMIC_BUILD_CACHE_DIR=/tmp/u012/cache/M18 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M18.log 2>&1; cache_test.go:233: ordered progress: \" 49/51 pass=0 fail=0 refused=0 crashed=0 skipped=49\\n 51/51 pass=1 fail=0 refused=0 crashed=0 skipped=50\\n\", want \" 50/51 pass=0 fail=0 refused=0 crashed=0 skipped=50\\n 51/51 pass=1 fail=0 refused=0 crashed=0 skipped=50\\n\""
  },
  {
    "test": "TestUnavailableCacheStillRuns",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/cache_test.go:237",
    "seconds": 3.43,
    "oracle": "Node/native comparison inside attempt; test requires pass verdict despite an unavailable cache.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M12"
    ],
    "unique_kills": [
      "M12"
    ],
    "last_proven_fail": "M12: cache_test.go:254: unavailable result cache changed verdict: {Path:pass/pad.js Directory:pass Kind:crashed Reason:mkdir /tmp/adamic-gate/TestUnavailableCacheStillRuns3490122339/002/not-a-directory: not a directory Adaptations:map[]}",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "PAttempt"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT=M12 ADAMIC_BUILD_CACHE_DIR=/tmp/u012/cache/M12 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M12.log 2>&1; cache_test.go:254: unavailable result cache changed verdict: {Path:pass/pad.js Directory:pass Kind:crashed Reason:mkdir /tmp/adamic-gate/TestUnavailableCacheStillRuns3490122339/002/not-a-directory: not a directory Adaptations:map[]}"
  },
  {
    "test": "TestImportedInputsRunFresh",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/cache_test.go:258",
    "seconds": 3.466,
    "oracle": "Node/native runs plus handwritten freshness/reason inequality. Node import failure is expected; native outcome is not pinned exactly.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M11"
    ],
    "unique_kills": [],
    "last_proven_fail": "M11: cache_test.go:280: mutable imported input served stale native observation: [exit 1 vs 0 exit 1 vs 0]",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestCompilerCacheProgram"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "PAttempt"
    ],
    "subsumer_seconds": 0.217,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT=M11 ADAMIC_BUILD_CACHE_DIR=/tmp/u012/cache/M11 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M11.log 2>&1; cache_test.go:280: mutable imported input served stale native observation: [exit 1 vs 0 exit 1 vs 0]",
    "subsumption_kills": 1
  },
  {
    "test": "TestCompilerCacheProgram",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/cache_test.go:284",
    "seconds": 0.217,
    "oracle": "Handwritten literal word, key dimensions, and dependency detection; own compiler output.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M11"
    ],
    "unique_kills": [
      "M01"
    ],
    "last_proven_fail": "M01: cache_test.go:307: changed compiler program served \"before\\n\", want \"after\\n\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "PReuse",
      "PCompilerKey",
      "PDependent",
      "PCompile"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT=M01 ADAMIC_BUILD_CACHE_DIR=/tmp/u012/cache/M01 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M01.log 2>&1; cache_test.go:307: changed compiler program served \"before\\n\", want \"after\\n\"",
    "vacuous_subcases": [
      "PDependent: unnamed before/after native-output loop and four key-dimension loops passed; final reference assertion failed (PDependent.log cache_test.go:322). These earlier checks do not call dependentProgram."
    ]
  },
  {
    "test": "TestCompilerCacheAcrossScratchDirectories",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/cache_test.go:326",
    "seconds": 3.521,
    "oracle": "Node/native agreement plus handwritten cache-hit and scratch independence checks.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M05",
      "M07",
      "M08",
      "M09",
      "M10",
      "M13"
    ],
    "unique_kills": [
      "M13"
    ],
    "last_proven_fail": "M13: cache_test.go:350: worker C cache missed when only scratch compiler path changed: {Path:pass/pad.js Directory:pass Kind:pass Reason: Adaptations:map[]}",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "PAttempt"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT=M13 ADAMIC_BUILD_CACHE_DIR=/tmp/u012/cache/M13 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M13.log 2>&1; cache_test.go:350: worker C cache missed when only scratch compiler path changed: {Path:pass/pad.js Directory:pass Kind:pass Reason: Adaptations:map[]}"
  },
  {
    "test": "TestCompilerWorkerMatchesSubprocess",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/compiler_test.go:12",
    "seconds": 3.749,
    "oracle": "Exact output/status/class/reason comparison with our own separate adamic c subprocess. Shared compiler errors remain invisible.",
    "oracle_kind": "self",
    "kills": [
      "M05",
      "M16"
    ],
    "unique_kills": [
      "M16"
    ],
    "last_proven_fail": "M16: compiler_test.go:30: compiler differs: actual={Stdout: Stderr:/tmp/adamic-gate/TestCompilerWorkerMatchesSubprocess1951681831/004/program.a:93:7: error TS2322: Type 'string' is not assignable to type 'number'.",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "PWorker"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT=M16 ADAMIC_BUILD_CACHE_DIR=/tmp/u012/cache/M16 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M16.log 2>&1; compiler_test.go:30: compiler differs: actual={Stdout: Stderr:/tmp/adamic-gate/TestCompilerWorkerMatchesSubprocess1951681831/004/program.a:93:7: error TS2322: Type 'string' is not assignable to type 'number'."
  },
  {
    "test": "TestCompilerHangHelper",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/compiler_test.go:35",
    "seconds": 0.007,
    "oracle": "Subprocess entry only; parent TestCompilerWorkerTimeout supplies assertions. Standalone timing measures an immediate return.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "helper",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": false,
    "evidence": "timing-TestCompilerHangHelper-{1,2,3}.log; parent timeout evidence M14.log and M15.log",
    "parent": "TestCompilerWorkerTimeout"
  },
  {
    "test": "TestCompilerWorkerTimeout",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/compiler_test.go:43",
    "seconds": 0.029,
    "oracle": "Handwritten 20 ms deadline, TimedOut/status assertions, worker closure and process reaping for known hanging helper.",
    "oracle_kind": "self",
    "kills": [
      "M14",
      "M15"
    ],
    "unique_kills": [
      "M14",
      "M15"
    ],
    "last_proven_fail": "M15: compiler_test.go:66: deadline failed to stop and reap compiler: {Stdout: Stderr: Exit:-1 Signal: TimedOut:true}",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "PWorker"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT=M15 ADAMIC_BUILD_CACHE_DIR=/tmp/u012/cache/M15 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M15.log 2>&1; compiler_test.go:66: deadline failed to stop and reap compiler: {Stdout: Stderr: Exit:-1 Signal: TimedOut:true}"
  }
]
```

| ID | origin/main file:line | Change | Failed rows |
| --- | --- | --- | --- |
| M01 | cmd/adamic-test262/cache.go:56 | omit compiler source key dimension | TestCompilerCacheProgram |
| M02 | cmd/adamic-test262/cache.go:60 | omit Node source key dimension | TestCacheKeyDimensions, TestNodeCacheProgram |
| M03 | cmd/adamic-test262/cache.go:63 | omit native C key dimension | TestNativeCacheGeneratedC, TestCacheKeyDimensions |
| M04 | cmd/adamic-test262/cache.go:49 | drop key part length boundaries | TestCacheKeyDimensions |
| M05 | cmd/adamic-test262/cache.go:35 | swap stdout and stderr in stored execution | TestCacheBypass, TestParallelCachedMatchesSerial, TestCompilerCacheAcrossScratchDirectories, TestEditCacheSeparation, TestLoweringSourceEdit, TestCacheAtomicBytes, TestWorkerLazyFallback, TestOrderedProgress, TestCompilerWorkerMatchesSubprocess |
| M06 | cmd/adamic-test262/cache.go:84 | accept envelope without digest validation | TestCacheAtomicBytes |
| M07 | cmd/adamic-test262/cache.go:90 | invert reusable eligibility | TestCacheBypass, TestCompilerCacheAcrossScratchDirectories, TestEditCacheSeparation, TestLoweringSourceEdit, TestCacheAtomicBytes |
| M08 | cmd/adamic-test262/cache.go:90 | invert timeout eligibility | TestCacheBypass, TestCompilerCacheAcrossScratchDirectories, TestEditCacheSeparation, TestLoweringSourceEdit, TestCacheAtomicBytes |
| M09 | cmd/adamic-test262/cache.go:67 | invert bypass selector | TestCacheBypass, TestCompilerCacheAcrossScratchDirectories, TestEditCacheSeparation, TestLoweringSourceEdit |
| M10 | cmd/adamic-test262/cache.go:109 | drop atomic cache publication | TestCacheBypass, TestCompilerCacheAcrossScratchDirectories, TestEditCacheSeparation, TestLoweringSourceEdit, TestCacheAtomicBytes |
| M11 | cmd/adamic-test262/run.go:358 | require both dependency forms | TestCompilerCacheProgram, TestImportedInputsRunFresh |
| M12 | cmd/adamic-test262/run.go:239 | drop cache mkdir fallback directory assignment | TestUnavailableCacheStillRuns |
| M13 | cmd/adamic-test262/run.go:276 | invert stable worker command selection | TestCompilerCacheAcrossScratchDirectories |
| M14 | cmd/adamic-test262/compiler.go:146 | drop timeout indication | TestCompilerWorkerTimeout |
| M15 | cmd/adamic-test262/compiler.go:91 | retain command after close | TestCompilerWorkerTimeout |
| M16 | cmd/adamic-test262/compiler.go:58 | report checker diagnostics as successful compilation | TestCompilerWorkerMatchesSubprocess |
| M17 | cmd/adamic-test262/run.go:154 | allow one extra attempted test |  |
| M18 | cmd/adamic-test262/run.go:213 | change progress interval | TestOrderedProgress |
| M19 | cmd/adamic-test262/run.go:552 | double pass counter increment | TestWorkerLazyFallback, TestOrderedProgress, TestMiniRunner, TestLargeCompilerOutputIsComplete |
| M20 | cmd/adamic-test262/cache.go:137 | include bypass environment in cache context |  |

Survivors:

- M17: u012_observation_test.go:18: total=5 skipped=1 unrun=3 attempted=1; mutant u012_observation_test.go:18: total=5 skipped=1 unrun=2 attempted=2. See witness-clean-limit.log and witness-M17.log, temporary observation harness saved in survivor-witness.go.txt.
- M20: u012_observation_test.go:28: context_equal=true node_equal=true; mutant u012_observation_test.go:28: context_equal=false node_equal=false. See witness-clean-context.log and witness-M20.log, temporary observation harness saved in survivor-witness.go.txt.

The brief cites 8de93800f4, but required a fresh origin/main checkout. All file lines and standalone diffs use 8171b3173bdbfce1f7982d3c4f731279307ece37. All requested names remain in the two cited files. No requested wrappers share a family checker; their assertions differ. The helper has no independent verdict and its immediate-return timing does not price the parent workload.

The toolchain was warm, but the initial list build exceeded the 90-second cap and was stopped. Repository cohere submodules and stage3/api npm dependencies still required installation. The retry list and whole-package clean baseline succeeded. ADAMIC_TEST262_MEASURE=1 enabled the measurement row. No baseline rows remained skipped. Setup and dependency wall timings were not captured separately; installation output is retained. Every later whole-package mutant run completed below 90 seconds.

The slice did not require a bounded matrix: every production mutant ran all 37 top-level package tests. Empty probes ran their direct entry callers, with results in probe-rows.json and probe-matrix.json. Probe failures are excluded from production kills and uniqueness. No agreement witnesses or construction-only checks occur in these requested rows. Corrupt-cache inputs exercise production cache validation. A separate compiler subprocess is our own oracle, so its classification is self, despite crossing a process boundary. No outside-authority values were claimed or checked.

The 20-mutant cap gives fewer than three mutants per row. Subsumption is only a hint over the reported caught mutants, not grounds to delete tests. The fastest measured bypass subsumer was selected from both out-of-slice candidates and the requested scratch-directory row. Parallel report parity accepts the M17 extra attempt and M19 doubled count; imported-input checks assert freshness without pinning every expected native outcome. M20 changes cache identity without changing execution output, but its before/after key observation proves a real change.

The initial matrix driver was interrupted after M06 completed to narrow probes to direct callers. M06's full package event is intact; its wall timing is unknown. Continuation did not narrow any production matrix. All standalone production diffs passed go vet individually. Production source was restored before final clean package run, go vet, and git diff --check. Probe diffs are separate P files and are not production mutants. Temporary observation tests were removed; their source is retained as text.

Coverage and reached-functions.txt document the code read before the fixed menu. This audit covers cache identity/storage/fallback/dependencies, worker transport/timeout, limits, and progress/count accounting. It does not settle repo-wide uniqueness, shared compiler correctness, every error path, or mutations in lower/native code. No other package's suite was run. Per-mutant native runner products used separate ADAMIC_BUILD_CACHE_DIR directories; there was no compiler-lowering mutation requiring a cold compiler rebuild. Standalone diffs contain no runtime selector. The switch source and orchestration scripts are retained only as text.

Timing: production matrix binary total 649.515 seconds; clean timing/coverage orchestration 169.705 seconds; individual mutant vet total 6.868 seconds. Switch test-binary build was 6.165 seconds. finish-runs.json and mutant-runs.json retain measured command wall times, including probes and survivor observations. The audit exceeded the soft 20-minute budget because it retained whole-package matrices and three runs of every row.

Push encountered prior remote audit 37d06a26, which stopped on a red baseline. Its evidence is preserved under prior-blocked-audit/. The earlier branch history was merged without importing production edits; this audit began clean at the stated main commit. Evidence-file unified diff context can trigger git diff --check whitespace diagnostics when stored as text; final source diff check was clean.
