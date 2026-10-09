Merged concurrency, scaling, moves, fuzzing, WASI hooks, host promises and async coverage onto the runtime area line.
Code checkpoint: b9479aa4beb65409312eaae5f406e94d3a6d7dfb; all requested merges and conflict resolutions are committed.
Native/lower, full uncached oracle, TestWASI, full WASI oracle and every parallel TSan fixture passed.
All nine race mutants caught in every one of 50 groups; all thirteen runtime controls also caught 50/50.
Linux verification only; no macOS run or full repository go test ./... in this unit.

## Branch and commits

Own branch: codex/concurrency-area, based on area/runtime 4d86c305dda261768b35687d199c1b2188c7ab71. Also contains current-main snapshot c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06, merged in 3b7188c. No push to main or any area branch.

| Merge | Incoming tip | Merge commit |
|---|---|---|
| concurrency-stack | 5a1963eca6e2b688b7bdaf4668a5c19989defc27 | 524a20e |
| wasm-threads | 59016fcc72a715e7fe4bb78526fe64bb4e193aee | 5c92ed9 |
| host-promises-typeof | 1b99f54e392d7cfe63bcfc6e642d4607d42e45fc | 67d0172 |

All three were actual merges, not no-ops. Merge commit messages describe the resolutions file by file.

## Conflict resolutions

Concurrency stack:

- lower/refusals.go: preserve area parsed-file pragma checking, including overridden directives, then concurrency preflight.
- runtime/adamic.c: preserve WASI signal guards, with native pthread support, locked output and a single winning panic.
- runtime/adamic.h: retain inline ordinary unshared counting and exact counted hooks; shared counts dispatch atomically. Preserve area region/WASI definitions, optional-field and data-field fast paths using packed atomic cache loads. Unify string metadata into the full UTF-16 view including supplementary halves; preserve ASCII and known-length shortcuts.
- runtime/heap.c: preserve the area's noinline last-reference destruction split and early returns together with shared atomic reference handling, thread-local queues/slabs, remote frees and shutdown.
- runtime/object.c: retain the optional lookup helper behind inline hits and publish shape/index as one atomic cache word.
- runtime/string_index.c: retain area's full UTF-16 view and publish shared metadata with release/acquire CAS; losing candidates free both their view and metadata. Shared readers do not mutate cursor or plain length. Short/ASCII strings retain their fast paths.
- oracle/counts.md: preserve existing area rows and add concurrency rows before regeneration.
- oracle/oracle_test.go: preserve both registration sets and concurrency variants.
- Automatic overlap in string_append.c keeps adamic_reference_count and length + added <= capacity. Preserve area's string builders, borrow chains and loop bindings. Adapt the cache mutant to target the complete field helper and the cache-loser mutant to the unified view field.

WASI hooks:

- prelude.d.ts: keep one current parallelMap declaration and current compiler API declarations.
- refusals.go: retain area pragma checking omitted by the older incoming ancestor.
- counts.md: retain the merged area table pending regeneration.
- Keep all four hooks: WASI pool never starts, native-only -pthread, 64-bit widening before the share hash shift, and linkable WASI adamic_release. Native -O2 objects before/after this merge are byte-identical: 50/50, with per-object SHA-256 values in w7-objects.json.

Host promises/async:

- refusals.go: preserve area pragma validation and remove obsolete permanent async syntax refusals.
- emit.go: dispatch async before ordinary cProgram while preserving request-handler emission.
- heap.c: retain WASI external release export and add async object destruction.
- oracle_test.go: concatenate area and async fixture registrations.
- counts.md: retain existing area counts and add async coverage rows pending regeneration.

## Separate fixes and observed failures

| Commit | Finding and correction |
|---|---|
| 1fac1a8 | Empty-string sharing test assumed unknown length; preserve its already-known length. Lazy-cache runtime behavior unchanged; eager-building mutant still fails. |
| a702a1e | TestWASI exposed POSIX pipe use in the host implementation. Guard native pipe/default loop code; WASI requires explicit wake and wait hooks, otherwise a clear panic. Native objects remain identical, 51/51, in host-wasi-objects.json. |
| db40beb | Host settlement race mutant could crash on its invalid payload before reporting its race. Keep its count read observable and require an actual TSan data-race report. |
| 41f6722 | Regenerate Linux counts separately. |
| 083ffa0 | A field-cache mutant escaped once in the loaded native suite. Start four readers together before the field-only witness phase and increase reads from 1,024 to 10,000. No ordering is added inside the conflicting phase. Final rerun caught 150/150 actual races. |
| 2e8cf4e | Add a counted WASI custom-host completion proof and missing-root-release mutant. Control alloc/frees 3/3; mutant 3/0 with identical output and successful exit, caught by counts. |
| b9479aa | Large WASI moves emission exceeded the existing one-minute compilation deadline. Give only emission compilation five minutes; runtime execution deadlines and comparisons stay unchanged. Targeted compile passed in 97.27 seconds. |

An overlapping WASI run also hit the container memory limit. Final WASI verification ran in isolation and passed; maximum child RSS 3,303,408 KiB. Failed runs are retained alongside successful logs. No checks were weakened to accept crashes, timeouts, wrong bytes or missing race reports.

## Environment and commands

Linux amd64; nproc=5, cgroup quota=4 CPUs. Go 1.27.1, clang 20.1.8, Node 24.19.0; WASI SDK 27. Every toolchain command sources /workspace/adamic-tools/env.sh. Initial setup timings: Go 0s, clang 1s, Node 1s, submodules 1s, cache warm 30s. WASI setup: Go 0s, clang 0s, Node 0s, SDK ready 12s, submodules 12s, cache warm 169s. These are gate timings on a loaded machine, not benchmark measurements.

Build after each merge: go build ./cmd/adamic. Native/lower and filtered oracle were run after each merge; initial failures and subsequent fixes are in step*.log.gz and the corresponding fix logs.

Final commands and outputs:

```sh
go test -count=1 -timeout 30m ./internal/native ./internal/lower
# native 193.063s; lower 26.459s, PASS
ADAMIC_GATE_UNCACHED=1 go test -count=1 -parallel 4 -timeout 30m ./internal/oracle
# PASS 466.162s after merging main

go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
# PASS 97.109s

go test -count=50 -timeout 60m -v -run '^TestParallel(ChecksCatchMutants|ScalingGuardMutants)$' ./internal/native
# PASS 433.486s

go test -count=1 -timeout 20m -v -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^concurrency$/^accepted$/.*$/^tsan$|^TestMovesAgreesWithNode$/^tsan$|^TestNativeAgreesWithNode$/^bench$/^parallel_files.a$/^tsan$' ./internal/oracle
# PASS 202.800s: 15 fixtures, threads=1/default, three runs each, 90 executions
python3 internal/oracle/testdata/moves/prove.py
# PASS refusal, diagnostics, count and race mutants
go test -count=1 -timeout 15m -v -run '^TestParallelRunnerAgreesAndCanFail$' ./internal/fuzz
# PASS 18.012s

PATH="/workspace/adamic-tools/wasi-sdk/bin:$PATH" ADAMIC_TEST_WASI=1 ADAMIC_ORACLE_WASI=1 go test -count=1 -timeout 15m -v -run '^TestWASI$' ./internal/native
# PASS 71.730s
ADAMIC_TEST_WASI=1 ADAMIC_ORACLE_WASI=1 go test -count=1 -timeout 10m -v -run '^TestWASIHostPromises$' ./internal/native
# PASS 11.163s, including counted leak mutant
ADAMIC_GATE_UNCACHED=1 ADAMIC_TEST_WASI=1 ADAMIC_ORACLE_WASI=1 go test -count=1 -timeout 30m -v -run '^TestWASI' ./internal/oracle
# PASS 448.180s; wrapper elapsed 449.878s
python3 internal/native/testdata/parallel/check_objects.py 1fac1a8
python3 internal/native/testdata/parallel/check_objects.py 67d0172
# respectively 50/50 and 51/51 identical native objects
```

Object checks also confirm no TSan hooks in release/ASan objects and reject ADAMIC_TSAN_TEST without TSan. Moves refusals and fix files are unchanged from the stack tip. Full oracle includes async/host coverage and its existing byte, ownership and leak mutants.

## Mutant rates on the final tree

| Race mutant | Groups caught | Actual TSan reports / executions |
|---|---|---|
| skip_items_share | 50/50 | 150/150 |
| lazy_cache | 50/50 | 150/150 |
| plain_shared_count | 50/50 | 150/150 |
| field_cache | 50/50 | 150/150 |
| remote_free | 50/50 | 150/150 |
| moves aliased element | 50/50 | 50/50 |
| moves flattened nested graph | 50/50 | 50/50 |
| host direct settlement | 50/50 | 150/150 |
| host queue without lock | 50/50 | 150/150 |

Each grouped runtime/host check runs three times and accepts only an actual TSan race diagnostic. Moves binaries built by the real mutant harness were each repeated 50 times. Complete host/moves run logs are archived; runtime group timings and reports are in committed-mutants.log.gz and runtime-rates.json.

Other runtime controls each caught 50/50: result_order (output comparison), reused_graph, oversized_slot, fixed_grain, eager_strings, one_worker and pointer_guard (contract assertions), loser_free (LeakSanitizer). Mean seconds per group, respectively: 0.151, 0.111, 0.118, 0.116, 0.117, 0.148, 0.115, 0.337. Race means: skip share 0.468, lazy cache 1.967, plain count 1.813, field cache 0.373, remote free 2.834. Host repeated groups took 30.917s/29.502s total; moves repeated binaries 6.704s/7.445s.

Additional host checkpoint, wake timeout, registry/settle/cancel/exit-cycle leaks, payload/status/encoding/token/hook contract mutants passed their checks. Async wrong resumed value/typeof fail the Node comparison, missing parameter retain fails ASan, and missing throw-local release fails leak checking. Moves after-use/diagnostic/fix/path mutants fail their refusal tests; gratuitous sharing changes the counted graph from plain/shared 2050/0 to 1/2049 and fails its guard. WASI byte/exit runner mutants are included in the final green WASI gate.

## Counts

441 rows after regeneration, versus 413 on the area base. Every one of the 413 existing rows is exactly unchanged: no missing rows, no increases, no changed area rows. The 28 additions are 13 parallel fixtures, parallel_files, one moves fixture and 13 async fixtures, listed in counts.json.

Six imported concurrency rows decrease relative to the stack's old table: moves objects retains/releases 8194/6153 to 6145/4104; fresh 11/26 to 10/25; identity 14/20 to 13/19; large 6155/6165 to 6154/6164; recursive_tree 11/17 to 9/15; parallel_files 8151046/8654931 to 8151045/8654930. These are the area's existing borrow/release optimizations applied to newly imported programs; allocations and frees are unchanged. Structured before/after values are in counts.json.

## Limits

This is Linux verification; no macOS measurement or test was run. The requested native/lower, full uncached oracle and WASI gates passed; a full repository go test ./... was not run. WASI host promises require host-provided wake and wait hooks; the default POSIX loop is unavailable there and explicitly refused. Finite 50-run results are observations, not a proof that a race witness can never miss on another scheduler. All requested conflicts are resolved and no gate remains running.
