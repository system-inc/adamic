u013 audited at origin/main 7b18d0576930caca4e22ce2eef92fcf563af52d0; all fifteen names exist.
Verdicts: 8 sacred, 5 subsumed, 1 cannot-judge, 1 helper.
Full matrix: 37 top-level tests per column; 15 admissible mutations plus 2 supplemental; no survivors or skips.
Three-run medians: 0.010 to 5.156 seconds; nproc=5; no scoped row passed its own empty-answer probe.
Evidence: test-audit/cmd-adamic-test262-corpus, review/test-audit/cmd-adamic-test262-corpus/; production restored.

```json
[
  {
    "test": "TestClassifyCorpus",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/corpus_test.go",
    "seconds": 0.013,
    "oracle": "self: handwritten .want classification and negative metadata; skip reasons checked by substring only",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M02",
      "M03"
    ],
    "unique_kills": [
      "M02"
    ],
    "last_proven_fail": "M03: corpus_test.go:49: testdata/corpus/negative_runtime.js: negative \"runtime \", want \"runtime TypeError\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT=M03 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M03.log 2>&1; corpus_test.go:49: testdata/corpus/negative_runtime.js: negative \"runtime \", want \"runtime TypeError\"",
    "supplemental_kills": [],
    "admissible_mutants_in_matrix": 15
  },
  {
    "test": "TestVerdictCorpus",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/corpus_test.go",
    "seconds": 0.012,
    "oracle": "self: JSON executions and expected outcome labels; Node is not run by this row",
    "oracle_kind": "self",
    "kills": [
      "M05"
    ],
    "unique_kills": [],
    "last_proven_fail": "M05: corpus_test.go:88: testdata/verdicts/both_pass.json (both succeed and match): got fail (stderr disagrees), want pass",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestEditCacheSeparation"
    ],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "P02"
    ],
    "subsumer_seconds": 2.409,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT=M05 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M05.log 2>&1; corpus_test.go:88: testdata/verdicts/both_pass.json (both succeed and match): got fail (stderr disagrees), want pass",
    "supplemental_kills": [
      "M06"
    ],
    "admissible_mutants_in_matrix": 15,
    "subsumption_mutants": 1
  },
  {
    "test": "TestNormalizeReason",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/corpus_test.go",
    "seconds": 0.01,
    "oracle": "self: handwritten diagnostic grouping and panic labels; no tsc comparison",
    "oracle_kind": "self",
    "kills": [
      "M07",
      "M08"
    ],
    "unique_kills": [
      "M07",
      "M08"
    ],
    "last_proven_fail": "M08: corpus_test.go:118: compiler panic: kind refused, want crashed",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "P03"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT=M08 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M08.log 2>&1; corpus_test.go:118: compiler panic: kind refused, want crashed",
    "supplemental_kills": [],
    "admissible_mutants_in_matrix": 15
  },
  {
    "test": "TestRewriteHarnessCalls",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/corpus_test.go",
    "seconds": 0.01,
    "oracle": "self: handwritten strings and substring checks",
    "oracle_kind": "self",
    "kills": [
      "M04"
    ],
    "unique_kills": [
      "M04"
    ],
    "last_proven_fail": "M04: corpus_test.go:133: missed throws:",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "P04"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT=M04 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M04.log 2>&1; corpus_test.go:133: missed throws:",
    "supplemental_kills": [],
    "admissible_mutants_in_matrix": 15
  },
  {
    "test": "TestFrontmatterShapes",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/corpus_test.go",
    "seconds": 0.011,
    "oracle": "self: handwritten frontmatter fields and classification",
    "oracle_kind": "self",
    "kills": [
      "M03"
    ],
    "unique_kills": [],
    "last_proven_fail": "M03: corpus_test.go:181: negative runtime",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestClassifyCorpus"
    ],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "P05"
    ],
    "subsumer_seconds": 0.013,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT=M03 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M03.log 2>&1; corpus_test.go:181: negative runtime",
    "supplemental_kills": [],
    "admissible_mutants_in_matrix": 15,
    "subsumption_mutants": 1
  },
  {
    "test": "TestMiniRunner",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/corpus_test.go",
    "seconds": 4.217,
    "oracle": "Node runs and is compared with native execution; self-written aggregate counts and refusal label",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M01",
      "M05",
      "M15"
    ],
    "unique_kills": [],
    "last_proven_fail": "M15: corpus_test.go:210: pass fixture: 0 reasons fail=[] refused=[] crash=[{pass/pad.js: command output exceeded capture limit 1}]",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestParallelCachedMatchesSerial"
    ],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "P06"
    ],
    "subsumer_seconds": 8.035,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT=M15 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M15.log 2>&1; corpus_test.go:210: pass fixture: 0 reasons fail=[] refused=[] crash=[{pass/pad.js: command output exceeded capture limit 1}]",
    "supplemental_kills": [],
    "admissible_mutants_in_matrix": 15,
    "subsumption_mutants": 3
  },
  {
    "test": "TestLargeCompilerOutputIsComplete",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/corpus_test.go",
    "seconds": 4.137,
    "oracle": "Node runs and is compared with native execution; self-written minimum C size and final newline",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M05",
      "M15",
      "M17"
    ],
    "unique_kills": [
      "M17"
    ],
    "last_proven_fail": "M17: corpus_test.go:251: large program: {Path:large.js Pass:0 Fail:0 Refused:0 Crashed:1 Skipped:0 Unrun:0 Total:1 Directories:[{Path:. Pass:0 Fail:0 Refused:0 Crashed:1 Skipped:0 Unrun:0 Total:1}] RefusalReasons:[] SkipReasons:[] CrashReasons:[{Reason:large.js: command output exceeded capture limit Count:1}] FailReasons:[] Passes:[] Adaptations:[] directories:map[.:0x2da5c760e280] refusalReasons:map[] skipReasons:map[] crashReasons:map[large.js: command output exceeded capture limit:1] failReasons:map[] adaptations:map[]}",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "P06"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT=M17 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M17.log 2>&1; corpus_test.go:251: large program: {Path:large.js Pass:0 Fail:0 Refused:0 Crashed:1 Skipped:0 Unrun:0 Total:1 Directories:[{Path:. Pass:0 Fail:0 Refused:0 Crashed:1 Skipped:0 Unrun:0 Total:1}] RefusalReasons:[] SkipReasons:[] CrashReasons:[{Reason:large.js: command output exceeded capture limit Count:1}] FailReasons:[] Passes:[] Adaptations:[] directories:map[.:0x2da5c760e280] refusalReasons:map[] skipReasons:map[] crashReasons:map[large.js: command output exceeded capture limit:1] failReasons:map[] adaptations:map[]}",
    "supplemental_kills": [],
    "admissible_mutants_in_matrix": 15
  },
  {
    "test": "TestOutputOverflowIsReported",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/corpus_test.go",
    "seconds": 0.01,
    "oracle": "self: 6 consumed bytes, abc retained, overflow flag",
    "oracle_kind": "self",
    "kills": [
      "M14"
    ],
    "unique_kills": [
      "M14"
    ],
    "last_proven_fail": "M14: corpus_test.go:267: overflow: written=6 error=<nil> buffer={buf:{buf:[97 98 99] off:0 lastRead:0} limit:3 exceeded:false}",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "P08"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT=M14 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M14.log 2>&1; corpus_test.go:267: overflow: written=6 error=<nil> buffer={buf:{buf:[97 98 99] off:0 lastRead:0} limit:3 exceeded:false}",
    "supplemental_kills": [],
    "admissible_mutants_in_matrix": 15
  },
  {
    "test": "TestEditCacheSeparation",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/edit_test.go",
    "seconds": 2.409,
    "oracle": "Node runs and is compared with native execution; self-written cache hit invariants",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M05",
      "M09"
    ],
    "unique_kills": [],
    "last_proven_fail": "M09: edit_test.go:43: lowering native result hit after changed identity",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestLoweringSourceEdit"
    ],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "P07"
    ],
    "subsumer_seconds": 3.763,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT=M09 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M09.log 2>&1; edit_test.go:43: lowering native result hit after changed identity",
    "supplemental_kills": [],
    "admissible_mutants_in_matrix": 15,
    "subsumption_mutants": 2
  },
  {
    "test": "TestNodeHarnessIdentity",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/edit_test.go",
    "seconds": 0.011,
    "oracle": "self: identity equality and inequality after fixture source edits",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "M13 (supplemental): edit_test.go:55: compiler source invalidates Node harness",
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "P09"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT=M13 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M13.log 2>&1; edit_test.go:55: compiler source invalidates Node harness",
    "supplemental_kills": [
      "M13"
    ],
    "limitation": "Only supplemental M13 caught this row. It cannot support a verdict under the fixed menu.",
    "vacuous_subcases": [
      "compiler-only source edit preserves Node harness identity: the empty identity passes this first assertion; execution edit then fails"
    ],
    "admissible_mutants_in_matrix": 15
  },
  {
    "test": "TestWorkerLazyFallback",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/edit_test.go",
    "seconds": 4.651,
    "oracle": "Node runs and is compared with native execution; fallback itself checks only exit zero and nonempty C, not equivalence of C",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M05",
      "M15",
      "M16"
    ],
    "unique_kills": [
      "M16"
    ],
    "last_proven_fail": "M16: edit_test.go:88: lazy fallback lost compiler behavior: {Stdout: Stderr:fork/exec /tmp/adamic-gate/TestWorkerLazyFallback2192421139/001/adamic: no such file or directory Exit:-1 Signal: TimedOut:false}",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "P06",
      "P11"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT=M16 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M16.log 2>&1; edit_test.go:88: lazy fallback lost compiler behavior: {Stdout: Stderr:fork/exec /tmp/adamic-gate/TestWorkerLazyFallback2192421139/001/adamic: no such file or directory Exit:-1 Signal: TimedOut:false}",
    "supplemental_kills": [],
    "admissible_mutants_in_matrix": 15
  },
  {
    "test": "TestLoweringSourceEdit",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/edit_test.go",
    "seconds": 3.763,
    "oracle": "Node runs and is compared with native execution; self-written cache invalidation invariants",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M05",
      "M09",
      "M12"
    ],
    "unique_kills": [
      "M12"
    ],
    "last_proven_fail": "M12: edit_test.go:132: one-byte lowering edit did not change native identity",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "P07"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT=M12 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M12.log 2>&1; edit_test.go:132: one-byte lowering edit did not change native identity",
    "supplemental_kills": [],
    "admissible_mutants_in_matrix": 15
  },
  {
    "test": "TestRunnerLocationHelper",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/edit_test.go",
    "seconds": 0.012,
    "oracle": "helper subprocess entry for TestRunnerLocationIdentity; default timing measures immediate return",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "helper",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "Three default runs return immediately; parent subprocess evidence is in TestRunnerLocationIdentity",
    "supplemental_kills": [],
    "parent": "TestRunnerLocationIdentity"
  },
  {
    "test": "TestRunnerLocationIdentity",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/edit_test.go",
    "seconds": 5.1,
    "oracle": "self: byte-for-byte equality of helper output after executable relocation",
    "oracle_kind": "self",
    "kills": [
      "M11"
    ],
    "unique_kills": [
      "M11"
    ],
    "last_proven_fail": "M11: edit_test.go:181: relocating identical runner bytes invalidates observations: f2439501b2f73ac7ef11651b9e2d698c6db557b3f0ac2e95abdeef3cb3e1f9dc c05dd1a9c885fef124f5676929d753f835fb678900b9f1cc2739395d4255da55",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "P10"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT=M11 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M11.log 2>&1; edit_test.go:181: relocating identical runner bytes invalidates observations: f2439501b2f73ac7ef11651b9e2d698c6db557b3f0ac2e95abdeef3cb3e1f9dc c05dd1a9c885fef124f5676929d753f835fb678900b9f1cc2739395d4255da55",
    "supplemental_kills": [],
    "admissible_mutants_in_matrix": 15
  },
  {
    "test": "TestCompilerStartupMeasurement",
    "package": "cmd/adamic-test262",
    "file": "cmd/adamic-test262/performance_test.go",
    "seconds": 5.156,
    "oracle": "self: subprocess Adamic C must equal internal/native.C output from Adamic Load and Lower; startup loop has no performance threshold",
    "oracle_kind": "self",
    "kills": [
      "M15"
    ],
    "unique_kills": [],
    "last_proven_fail": "M15: performance_test.go:39: command output exceeded capture limit",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestLargeCompilerOutputIsComplete"
    ],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "P12"
    ],
    "subsumer_seconds": 4.137,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT=M15 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M15.log 2>&1; performance_test.go:39: command output exceeded capture limit",
    "supplemental_kills": [],
    "admissible_mutants_in_matrix": 15,
    "subsumption_mutants": 1
  }
]
```

M06 and M13 are supplemental and excluded from kills, unique kills and verdicts. Subsumption rests only on each row's listed admissible kills.

| ID | File:line at starting origin/main | Change | Failed rows |
|---|---|---|---|
| M01 | cmd/adamic-test262/classify.go:90 | return "negative parse or early error" -> return "" | TestClassifyCorpus, TestMiniRunner, TestOrderedProgress, TestParallelCachedMatchesSerial |
| M02 | cmd/adamic-test262/classify.go:156 | case "Proxy", "Reflect" -> case "Reflect" | TestClassifyCorpus |
| M03 | cmd/adamic-test262/frontmatter.go:135 | parsed.NegativeType = value ->  | TestClassifyCorpus, TestFrontmatterShapes |
| M04 | cmd/adamic-test262/rewrite.go:112 | builder.WriteByte('"') ->  | TestRewriteHarnessCalls |
| M05 | cmd/adamic-test262/verdict.go:58 | input.Native.Exit == 0 && input.Node.Stdout -> input.Native.Exit != 0 && input.Node.Stdout | TestCompilerCacheAcrossScratchDirectories, TestEditCacheSeparation, TestLargeCompilerOutputIsComplete, TestLoweringSourceEdit, TestMiniRunner, TestOrderedProgress, TestParallelCachedMatchesSerial, TestUnavailableCacheStillRuns, TestVerdictCorpus, TestWorkerLazyFallback |
| M06 | cmd/adamic-test262/verdict.go:58 | Supplemental:  && input.Node.Stdout == input.Native.Stdout && input.Node.Stderr ->  && input.Node.Stderr | TestVerdictCorpus |
| M07 | cmd/adamic-test262/reason.go:54 | "'…'" -> "''" | TestNormalizeReason |
| M08 | cmd/adamic-test262/reason.go:89 | return "crashed", "compiler panicked" -> return "refused", "compiler panicked" | TestNormalizeReason |
| M09 | cmd/adamic-test262/cache.go:63 | cacheKey("test262-native-v1", code, library, command, context) -> cacheKey("test262-native-v1", code, library, command, "") | TestCacheKeyDimensions, TestEditCacheSeparation, TestLoweringSourceEdit |
| M10 | cmd/adamic-test262/cache.go:56 | cacheKey("test262-compiler-v1", program, compiler, command, context) -> cacheKey("test262-compiler-v1", program, compiler, command, "") | TestCompilerCacheProgram |
| M11 | cmd/adamic-test262/cache.go:173 | parts = append(parts, "runner", cacheKey(string(contents))) -> parts = append(parts, path, cacheKey(string(contents))) | TestRunnerLocationIdentity |
| M12 | cmd/adamic-test262/cache.go:231 | parts = append(parts, name, string(contents)) -> parts = append(parts, name, string(contents[:0])) | TestLoweringSourceEdit |
| M13 | cmd/adamic-test262/cache.go:200 | Supplemental: name == "compiler.go" \|\| strings.HasSuffix -> strings.HasSuffix | TestNodeHarnessIdentity |
| M14 | cmd/adamic-test262/run.go:445 | buffer.exceeded = true ->  | TestOutputOverflowIsReported |
| M15 | cmd/adamic-test262/run.go:390 | stdout.limit = limit -> stdout.limit = 0 | TestCompilerCacheProgram, TestCompilerStartupMeasurement, TestCompilerWorkerMatchesSubprocess, TestImportedInputsRunFresh, TestLargeCompilerOutputIsComplete, TestMiniRunner, TestNativeCacheGeneratedC, TestNodeCacheProgram, TestParallelCachedMatchesSerial, TestProgramCPUDeadline, TestUnavailableCacheStillRuns, TestWorkerLazyFallback |
| M16 | cmd/adamic-test262/run.go:661 | Drop the entire e.fallback.once.Do(func(){...}) statement | TestWorkerLazyFallback |
| M17 | cmd/adamic-test262/run.go:390 | stdout.limit = limit -> stdout.limit = min(limit, outputLimit) | TestLargeCompilerOutputIsComplete |

Survivors: none. All fifteen admissible mutations and both supplemental mutations were caught. survivor-clean.log and survivor-M02.log are an extra behavior witness for the killed M02 mutation, not survivor claims.

The starting commit is 7b18d0576930caca4e22ce2eef92fcf563af52d0, fetched from origin/main. All fifteen requested names exist in their named files. The older commit printed in the brief is not the starting commit.

The whole package has 37 top-level tests. The baseline passed with ADAMIC_TEST262_MEASURE=1 and no skipped rows. Each mutation ran the whole package, including the rows outside this unit. Repo-wide uniqueness was not checked.

The scoped tests have distinct assertions. Shared preparation and execution helpers do not make these tests one family under the brief's exception for distinct assertions. TestRunnerLocationHelper is only a subprocess entry for TestRunnerLocationIdentity. Its three default timings measure immediate return.

The code under test is the Go test262 runner, not Node, clang, internal/lower or internal/native.C. CompilerStartupMeasurement compares subprocess Adamic against in-process Adamic. That reference is self, despite the subprocess invocation. Mutating the shared lowering or C emitter would change both answers, so this audit left them intact.

Classification, verdict JSON, diagnostic normalization, rewriting and frontmatter expectations are handwritten. Diagnostic text containing a TS code does not make normalization an external-authority oracle. No external authority value was checked in this unit.

The integration tests use Node agreement and also handwritten counts, labels or cache invariants. WorkerLazyFallback's fallback assertion checks exit zero and nonempty C, without comparing that C against another implementation. Classification reasons use substring matching. These are limits of the assertions, not inferred correctness claims.

The mutation plan was frozen before any kill result. It spans production classification, metadata, rewriting, comparison, normalization, identities, buffering and fallback. M06 and M13 mistakenly drop predicates, outside the requested menu. They are supplemental. Their kills are reported separately and do not support a verdict. NodeHarnessIdentity consequently has no admissible demonstrated kill and is cannot-judge, rather than claiming a sacred verdict from the supplemental result.

The initially written switch generator treated Go byte offsets as Python character offsets and omitted variadic forwarding. Both caused build failures before mutant test runs. The corrected generator preserves the original functions and dispatches to code-derived variants through ADAMIC_MUTANT. Failed build attempts remain logged. These implementation mistakes cost time; they were not test failures.

The before-mutation function inventory conservatively listed all production functions. reached-functions.json adds a static transitive graph from immutable source. Method selectors are resolved by name, so this graph is an overapproximation. It is not a coverage profile or proof that every listed branch ran.

Timings use the test binary's package elapsed field from three independent -count=1 runs. This excludes Go command compilation time. Command wall times are separately recorded for the baseline and matrix. Warm medians can hide cold costs: LargeCompilerOutputIsComplete took 12.526, 4.026 and 4.137 seconds in its isolated runs.

The opt-in startup measurement was enabled throughout. Its 30 interleaved compilations have no asserted performance threshold. It still checks C equivalence and can fail under output capture mutations and its empty-answer probe.

Empty-answer probes are separate from mutation verdicts. They run only the scoped rows that call their respective entries. NodeHarnessIdentity's compiler-only edit assertion accepts an empty identity before its runner-edit assertion rejects it. RunnerLocationIdentity rejects an empty prepareCache answer through its helper's error path. Such probe failures show rejection of no answer, not semantic correctness of a working answer.

There are no compiler or port-source mutants, so this unit does not require fresh native product cache keys or per-port rebuilds. Go runner variants are compiled into one switched test package. All standalone production diffs have no switch, apply to the starting index, and are individually checked with go vet.

The approximate 20-minute budget was exceeded. Full-package columns took tens of seconds each, on top of 45 isolated timing invocations, the clean baseline, reading the production code and repairing instrumentation. The matrix was not narrowed simply to manufacture uniqueness or meet the unit budget. Exact measured totals are saved in the final report.

Subsumption is only a hint from fifteen admissible mutations. Supplemental mutations and probes do not count. No test deletion is proposed. No other packages or repo-wide replay were run, and no external oracle implementation was mutated.

Measured totals:

```json
{
  "setup": "warm environment worked, cloud/setup.sh skipped",
  "npm_seconds": 0.7175376849991153,
  "nproc": 5,
  "baseline_binary_seconds": 56.101,
  "baseline_command_seconds": 57.92519243300194,
  "timing_binary_seconds": 121.68300000000002,
  "matrix_binary_seconds": 801.7310000000001,
  "matrix_command_seconds": 847.5097196030001,
  "matrix_frontend_and_build_overhead_seconds": 45.77871960300001,
  "probe_command_seconds": 50.93034880099731,
  "clean_build_seconds": 4.362867169002129,
  "standalone_vet_seconds": "not separately measured",
  "finished_utc": "2026-10-09T10:17:18.966997+00:00"
}
```
