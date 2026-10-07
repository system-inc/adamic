Fixed atomic owner-count queries, retained-parent cache writes and ASCII pool initialization; merged the shared leak checker.
Commits: runtime 210c226, developer-tools merge 6b71f8c (f6eef5df), leak-test migration 5c238f5, deterministic cache mutant proof 0ac21ce; base 51488ac.
Commands and outputs: focused native/TSan/guard PASS 32.115s; three TSan mutants and parallel ASCII bytes PASS 18.516s; async PASS 5.727s; fuzz campaign PASS 52.578s.
Mutants: plain parent units, plain owner count and ASCII header write each race in 3/3 TSan runs; leaked frame local and parallel result graph fail LSan and the counted rule.
Not covered: macOS execution and complete repository gate. Full native passed in 646.784s and full uncached oracle passed in 841.650s.

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
