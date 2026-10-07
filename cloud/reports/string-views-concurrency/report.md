Fixed atomic owner-count queries, retained-parent cache writes and ASCII pool initialization; merged and applied shared leak checks to async, fuzz and record tests.
Commits: runtime 210c226, developer-tools merge 6b71f8c (f6eef5df), leak-test migration 5c238f5, deterministic cache mutant proof 0ac21ce, record dependency merge 4601e02 (754e666), record migration a29a1b9, const record audit and single signal sender 01ef973; base 51488ac.
Commands and outputs: full native PASS 661.420s, full uncached oracle PASS 932.808s; focused TSan/views PASS 18.516s, async PASS 5.727s, fuzz PASS 52.578s, records PASS 116.681s, final statics/signal PASS 62.121s.
Mutants: plain parent units/count and ASCII writes, cache CAS removal and signal unsafe flush caught by TSan; frame-local, parallel-result and record-key leaks caught by LSan and actual counted builds; Node and ASan/UBSan catch the behavioral and retention mutants detailed below.
Not covered: actual macOS execution, leaks --atExit on macOS, and the complete repository gate. Full native and uncached oracle are green.

## Runtime fix

The merged view code still read owner->heap.references plainly in two decisions. These reads race with the shared atomic count updates. All owner decisions now use one adamic_reference_count snapshot. Retention continues through adamic_retain's atomic shared path; new view owner fields are set before publication. A zero-count region is borrowed storage, not an immortal owner.

A whole Unicode slice returns the input itself. The previous units assignment therefore wrote that shared parent's plain cache, also racing with length readers. Propagate units only into a newly allocated view. Whole-input shared metadata is already available through string_index.c's complete index publication with release/acquire CAS; no additional parent field is written. Immortal input headers are left unchanged too.

ASCII bytes are const and both byte/header arrays have complete C aggregate initializers before any worker can start. Headers have zero counts, known unit length and the literal index marker, so the runtime never mutates their fields. Both arrays are classified in docs/runtime-statics.md; the const byte array is excluded from the mutable inventory and the headers remain explicitly listed.

## Race and correctness evidence

TestParallelMap/tsan/strings passed at one and four workers, three executions per setting. TestRuntimeStaticsAreListed, TestRuntimeStringViews and the 11,460-answer Node UTF-16 sweep passed. TestParallelStringViews compares all 128 byte values, including NUL, against Node at one/four workers under ASan, ASan with slabs, counted, malloc and TSan builds. First worker accesses rendezvous before indexing. TSan repeats each setting three times.

| Durable mutant in TestParallelChecksCatchMutants | Catcher | Observed |
|---|---|---|
| view_parent_units: restore plain units assignment on retained input | TestParallelMap strings harness under TSan | Explicit data race in adamic_string_slice, all 3 attempts, exit 66 |
| view_owner_count: use owner->heap.references directly | Same strings harness under TSan | Explicit data race in string_share.c/heap.c, all 3 attempts, exit 66 |
| ascii_character_write: write units into the static header on every lookup | Parallel string_views.c harness under TSan | Explicit data race in ascii_character, all 3 attempts, exit 66 |

The runner copies the runtime before each mutant and requires successful strict C compilation and a positive TSan race. Warnings, compiler refusals, crashes without a race report and timeouts do not count. No working-tree mutant remains.

## Shared leak checks

Merged origin/devtools/stage1-leaks at f6eef5df9f951df90044d4608c07dd5ce6529190 with git merge, not cherry-pick. Resolved conflicts in internal/oracle/oracle_test.go and input_test.go by keeping the shared-helper implementation and the integration branch's surrounding tests. The merge also carries developer-tools' stage-1 port updates and its Darwin counted-build oracle rule.

Moved these named tests:

- internal/oracle/TestAsyncThrowReleasesFrame, including control and missing-local-release: leakcheck.Report.
- internal/oracle/TestAsyncTypeOfWrongKindMutant: ordinary native run and leakcheck.Report for the finished wrong-kind program.
- internal/oracle/TestAsyncCompilerChecksCatchMutants: ordinary native sanitizer runs; leakcheck.Report for the finished wrong-resumed-value program. The missing-parameter-retain mutant stops with ASan use-after-free and is not a finished program eligible for a leak check.
- internal/fuzz/TestMovesCampaign and TestParallelRunnerAgreesAndCanFail: Checkout's shared leakcheck.Check path, inherited by other Checkout.Try users too.

The fuzz runner preserves its isolated environment, process-group deadline and both default/one-worker leak witnesses. Leak detection is selected inside internal/leakcheck: Linux reruns under LSan; Darwin builds counted, checks allocations == frees + in-regions, then uses leaks --atExit for malloc storage outside counts. Native comparison sets detect_leaks=0 only on Linux. No migrated oracle test sets ASAN_OPTIONS itself.

| Leak mutant | Linux check | Counted rule, exercised with actual Linux builds |
|---|---|---|
| Throw frame missing local release | LeakSanitizer reports memory leaks | 11 allocations, 10 frees, 0 in regions: 1 leaked value |
| Missing parallel results release | LeakSanitizer reports leaked result graph, default and one worker | 19 allocations, 2 frees, 0 in regions: 17 leaked values, both worker settings |

Controls pass both checks. TestParallelLeakCheckCatchesMutant is durable in internal/fuzz; its task results are dynamically allocated strings, so literal immortality cannot hide the leak. The counted-rule checks run on Linux as well as Darwin, proving Darwin's accounting predicate without claiming a macOS execution. The resumed-value and typeof mutants remain caught only by independent Node stdout, and missing frame-parameter retention remains caught by ASan use-after-free.

## Commands and logs

Source /workspace/adamic-tools/env.sh before every Go command. Setup reported Go, clang, Node and submodules ready at 1s, cache warm/done 204s, nproc 5, cgroup cpu.max 400000 100000, 17.6 GB. No setup failure.

```sh
bash cloud/setup.sh > /tmp/string-views-concurrency-setup.log 2>&1
go test ./internal/native -run 'TestParallelMap/tsan/strings|TestRuntimeStaticsAreListed|TestRuntimeStringViews|TestStringsMatchJavaScript' -count=1 -v -timeout 30m > /tmp/string-views-concurrency-focus.log 2>&1
go test ./internal/native -run 'TestParallelChecksCatchMutants/(view_parent_units|view_owner_count|ascii_character_write)|TestParallelStringViews' -count=1 -v -timeout 30m > /tmp/string-views-concurrency-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestAsyncThrowReleasesFrame|TestAsyncTypeOfWrongKindMutant|TestAsyncCompilerChecksCatchMutants' -count=1 -v -timeout 30m > /tmp/string-views-concurrency-async-leaks.log 2>&1
go test ./internal/fuzz -run 'TestMovesCampaign|TestParallelRunnerAgreesAndCanFail' -count=1 -v -timeout 30m > /tmp/string-views-concurrency-fuzz-leaks.log 2>&1
go test ./internal/fuzz -run '^TestParallelLeakCheckCatchesMutant$' -count=1 -v -timeout 30m > /tmp/string-views-concurrency-fuzz-mutant.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/native ./internal/oracle -count=1 -timeout 30m > /tmp/string-views-concurrency-final-gate.log 2>&1
go vet ./internal/native ./internal/oracle ./internal/fuzz ./internal/leakcheck > /tmp/string-views-concurrency-vet.log 2>&1
```

Focused async and fuzz results above are uncached where applicable. Fuzz campaign: 8 move seeds, 3 agreed and 5 expected refusals; TSan ready. The explicit leaked-result proof passed in 1.230s. Vet, formatting and diff checks passed. Logs are compressed beside this report.

## Full gate and cache mutant proof

The first full run passed the uncached oracle in 841.650s. Native failed only TestRuntimeStaticsProtectionMutants/shared_string_index: its small BMP cache could publish before competing builders reached publication, allowing the CAS-removal mutant to survive one execution. This was a weakness in the negative proof, not a new production change.

Commit 0ac21ce uses the existing four-builder cache gate and cache_race.c fixture for this mutant. All four workers finish their cache candidates before publication, forcing competition. The gate exists only in the mutant runtime snapshot. The corrected CAS-removal proof produced an explicit TSan data race in all three executions; the statics BMP control and TestParallelCachePublication/tsan passed (14.530s combined). Production string_index.c is unchanged.

```sh
go test ./internal/native -run 'TestRuntimeStaticsProtectionMutants/shared_string_index|TestRuntimeStaticsParallel/string_bmp|TestParallelCachePublication/tsan' -count=1 -v -timeout 30m > /tmp/string-views-concurrency-cache-proof.log 2>&1
go test ./internal/native -count=1 -timeout 30m > /tmp/string-views-concurrency-native-final.log 2>&1
```

The final full native rerun passed in 646.784s. The full oracle result remains valid for these fixes: the later change only strengthens the native mutant harness.

## Record leak helper follow-up

The requested record tests were absent from this feature branch and its 51488ac base. Located them on origin/codex/runtime-records at 754e666 and merged that branch (4601e02) to bring their runtime and fixtures here. The merge was conflict-free and added six record files/declarations without changing compiler files.

Moved TestRecordsAgainstNode, TestRecordMutants and TestRecordReadMutants onto internal/leakcheck. Successful controls use Report; isolated mutant libraries use Check with an optional BuildCounted callback. The callback compiles the same changed runtime for the Darwin counted check; compiling the production runtime would accidentally check the control. The tests now use package native_test and public native.RuntimeLibrary/Flags/RuntimeLinkFlags, avoiding an import cycle through leakcheck's native builder. Ordinary runs set detect_leaks=0 only on Linux. Panics and ASan/UBSan stopping mutants keep their diagnostic checks and are not checked as finished programs.

Two distinct runs must not be confused:

- Early isolated checkout run, record-tests.log: FAILED to compile, reporting missing context/syscall/time imports. Imports were added while that build was running; that run supplies no correctness evidence.
- Fresh feature-branch run, record-final.log: PASS, package exit 0, 116.681s. All three requested tests and all nine mutants passed.

Every feature-branch control remained byte-identical to Node and balanced, including 3,500,004 allocations/frees for the million-key workload. Three ordering mutants were caught by Node, three read-contract mutants by exact behavior, stored-key freeing by ASan, null-slot access by UBSan, and overwrite-key leakage by both shared leak checking and an actual counted mutant build. The leak mutant reported 680,000 bytes in 10,000 allocations under Linux LeakSanitizer; its counted run had 350,024 allocations, 340,024 frees, 0 in regions, proving 10,000 leaked values. Actual macOS execution remains untested.

```sh
go test ./internal/native -run '^TestRecordsAgainstNode$|^TestRecordMutants$|^TestRecordReadMutants$' -count=1 -v -timeout 15m > /tmp/string-views-concurrency-record-final.log 2>&1
go vet ./internal/native ./internal/oracle ./internal/fuzz ./internal/leakcheck > /tmp/string-views-concurrency-record-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/native ./internal/oracle -count=1 -timeout 30m > /tmp/string-views-concurrency-record-full-gate.log 2>&1
```

Vet and diff checks passed. The full uncached oracle after record integration passed in 932.808s. Native failed in 1172.289s on two statics checks, described below; its final rerun passed in 661.420s.

Automatic approval review initially rejected recording the PASS because it conflated the failed isolated record-tests.log with the passing feature-branch record-final.log. Both logs were inspected separately, and both outcomes are disclosed here.

## Integrated statics gate follow-up

The integrated full uncached oracle passed in 932.808s. The accompanying full native run failed two checks: record.c had not yet received the runtime-file audit marker, and handler_buffer_mutant was caught twice but terminated without a race report on the third execution. No other native failures were reported. Both failures remain in record-full-gate.log.

record.c has only compile-time const storage: const field-name pointers, reference masks, shapes and diagnostic byte arrays. Added its runtime-file audit and that classification to docs/runtime-statics.md; no production runtime change was needed.

The signal fixture let all three pool workers raise the terminating signal. A later sender can terminate the process after the stop loop restores the default handler while TSan is still diagnosing the first handler. Changed only the test fixture to select one sender using a relaxed atomic flag before its writes. Other workers continue printing; the selection introduces no ordering between later output accesses. Production signal handling is unchanged. The original unchanged focused rerun passed, and the strengthened fixture then passed the complete signal suite and statics guard in 62.121s. Handler-buffer and exit-flush mutants each produced explicit TSan races in all three executions.

```sh
go test ./internal/native -run '^TestRuntimeStaticsAreListed$|^TestRuntimeStaticsSignalAndExit$' -count=1 -v -timeout 10m > /tmp/string-views-concurrency-record-statics-final.log 2>&1
go test ./internal/native -count=1 -timeout 30m > /tmp/string-views-concurrency-native-integrated-final.log 2>&1
```

The final full native rerun passed in 661.420s. The full oracle result applies to the final production code: subsequent changes only audit docs and a native test fixture.

Final validation: full native exit 0 (661.420s), full uncached oracle exit 0 (932.808s), targeted fuzz and leak mutants exit 0, vet/format/diff checks clean. Only codex/string-views-concurrency was pushed; no pull request was opened.
