# Concurrency integration, October 6

This is the third pass after the accepted runtime and inline-count work. The compiler branch at `2e284e1` was merged, not rebased, by `a2ca470`. The three-argument ABI remains unchanged. The compiler's Shareable inference and conservative effect proofs were taken from that branch; no lowering or JavaScript source was edited in this pass.

`46bf339` adds `_DARWIN_C_SOURCE` beside `_POSIX_C_SOURCE` in parallel.c. Runtime harnesses set `ASAN_OPTIONS=detect_leaks` only on Linux. On Darwin they check unsanitized `Options{Malloc: true}` binaries with `leaks --atExit -- <binary>`, at both one and four threads. The existing release mutant also no longer passes a Linux-only ASan option on Darwin. Darwin runtime TSan controls and race mutants run when a compile-and-run clang probe succeeds; an unavailable toolchain or runtime logs its specific failure. This machine is Linux; none of these Darwin paths was executed here.

`9aa6c57` registers all 13 accepted compiler fixtures in the main oracle and removes the opt-in native gate. Every compiled fixture using parallelMap is compared with source Node and the JavaScript backend, with one thread and the unset/default override, under ASan/UBSan, release, sanitized size classes and separate Linux TSan. Linux sanitizer variants also run with leak detection enabled; Darwin gets the unsanitized malloc/leaks variant. Race observations are never cached. All 28 refusal fixtures still hold exact messages. Counted parallel fixtures use `ADAMIC_THREADS=1`, because peak liveness and work performed after an exception depend on scheduling.

`4bd2315` adds the file workload, renamed to `bench/parallel_files.a` in `6360919` under Kirk's standing rule: 4,096 deterministic in-memory source strings, pure tokenization into local arrays, per-file summaries using a fresh local Map, and ordered merging. Native, Node and Bun execute the same source. Node uses the existing oracle loader; Bun resolves the independent sequential shim through a temporary package under NODE_PATH, without changing the source. Bun 1.3.14 executes `.a` directly, verified without a preload hook; no Bun source loader is needed. The runner discovers both `.a` and existing `.ts` programs; `.a` sources always use the Node oracle loader. `136e89f` adds discovery and source registration. The benchmark runner interleaves native thread settings with Node and Bun in every round.

All versions print:

```text
4096 files, 20282626 UTF-16 units, 2621440 tokens, 556992 distinct per file, ordered digest 823772603
```

Its one-thread counted build records allocations 3,952,651, frees 3,952,651, retains 8,151,046, releases 8,654,931, peak live 8,839 and regions 0. The oracle also runs this program under every parallel native variant, including TSan.

## Reproduce on the 16-core Mac

From the repository root, with Go, Xcode clang, Node 24 and optionally Bun on PATH:

```sh
go run ./bench -only parallel_files -threads 1,2,4,8,16 -rounds 5 > /tmp/adamic-parallel-files.log 2>&1
cat /tmp/adamic-parallel-files.log
```

The runner prints machine, versions, load before/after, every round, best times, memory, output agreement and the one-thread counts. The same command with `-threads 1,2,4` measures this container's four quota-limited CPUs (`nproc` reports five).

## Mutants run in this pass

Each mutation compiles and is tested in isolation. A build failure or timeout is not accepted as a proof.

| Mutation | What caught it |
| --- | --- |
| Omit sharing the items | TSan data race in adamic_retain, exit 66 |
| Plain shared retain count | TSan data race in adamic_retain_slow, exit 66 |
| Non-thread-local emitted slot cache | TSan data race in adamic_object_field, exit 66 |
| Foreign free directly into the owner's lists | TSan data race in give_local, exit 66 |
| Reverse result slots | Harness index comparison, exit 3 |
| Skip revisiting reused shared containers | Lifecycle harness abort |
| Create a second worker for ADAMIC_THREADS=1 | One-thread harness abort |
| Free the remaining shared owner prematurely | ASan heap-use-after-free on dynamically built text |
| Reverse the generated program's map results after joining | Source Node comparison: stdout differs, at one thread and default; sanitizer and leak clean |
| Treat a captured let as an immutable binding | Exact captured_let refusal fixture: expected Refused, got nil; go test exit 1 |

The first seven and the premature-release proof ran in the full native suite too. The generated-code order mutant is retained as `TestParallelOracleCatchesResultOrder`. The captured-let mutant was made only in a detached scratch checkout by replacing `return node.Flags&ast.NodeFlagsConstant != 0` with `return true` in parallel.go. The branch's compiler files were not edited. Its first cold dependency build exceeded a 120-second budget, which was not counted as catching it; pointing that checkout's go.work at the installed checker and allowing a 600-second build produced the intended refusal failure in 0.315s. The unchanged refusal control passed in 0.324s.

Compiler limits remain those documented in [the compiler worker's report](claims/concurrency-compiler.md): unknown closure calls and class dispatch are refused conservatively; freshness is not propagated through helper parameters, and mutable task-local captures can be refused. This pass does not add moves, async or await. It does not rerun the other five historical compiler mutants. Darwin compilation, Darwin leaks, Darwin TSan, the 16-core scaling run, cgroup v1 fallback and contended Weak throughput were not measured on this Linux host.

## Verification commands and observations

Output went directly to log files, never through a pipe. The toolchain was reused from the accepted inline-count pass: Go 1.27.1, clang 20.1.8, Node 24.19.0; its setup timing lines were Go/clang/Node/submodules ready at 0s, cache warm and done at 57s. `nproc` is 5; cgroup cpu.max is `400000 100000`, so the pool defaults to four executors. Bun 1.3.14 was installed for the three-runtime source comparison.

```sh
source /workspace/adamic-tools/env.sh
go test ./internal/native -run 'TestParallel|TestReleaseShared' -count=1 -v > /workspace/parallel-platform.log 2>&1
# exit 0, native 37.580s; every runtime control and all seven mutants passed
go test ./internal/oracle -run 'TestConcurrency|TestNativeAgreesWithNode/internal/oracle/testdata/concurrency' -count=1 -v > /workspace/concurrency-integration.log 2>&1
# exit 0, oracle 7.168s; 13 accepted fixtures, four native variants, two thread settings, 28 refusals
go test ./internal/oracle -run 'TestParallelOracleCatchesResultOrder|TestCountsAreRecorded' -count=1 -v -args -update-counts > /workspace/parallel-files-counts.log 2>&1
# exit 0, oracle 14.608s; result-order mutant caught only by Node
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./... > /workspace/concurrency-full-gate.log 2>&1
# exit 0, every package passed; native 350.951s, oracle 453.199s
# Unicode Node conformance took 1085.218s; large cohere ports made this a long gate.
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /workspace/concurrency-a-counts.log 2>&1
# exit 0, oracle 54.872s; unchanged numbers recorded under the renamed .a path
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m -parallel 4 ./internal/oracle > /workspace/concurrency-a-oracle.log 2>&1
# exit 0, oracle 359.405s; complete rerun after the .a rename, including separate TSan
```

The whole-repository run had already compiled the benchmark's former filename before Kirk's new naming rule arrived. Its source bytes were unchanged by the rename; the complete oracle and counts reruns above tested the final `.a` path. Whole-repository vet passed before and after the rename; bench vet passed again after removing the unnecessary Bun preload hook. gofmt and git diff --check were clean. Direct `bun bench/parallel_files.a` with the temporary shim on NODE_PATH exited zero and matched native and Node; `039d402` removes the redundant hook. Existing `.ts` programs were left unchanged.

The full package output and extracted runtime mutant findings are retained in [concurrency-integration-results.json](concurrency-integration-results.json). The benchmark's raw five-round output is [parallel_files.measurements.log](../bench/parallel_files.measurements.log). This pass did not repeat the already accepted main-versus-branch single-thread timings; no runtime execution code on Linux changed after the inline-count pass.

## File workload measurements

Final runner `039d402`, Linux amd64, AMD EPYC 9V74, five visible logical CPUs and a four-CPU quota. clang 20.1.8, Node 24.19.0, Bun 1.3.14. All tests finished before these timings. Load before was **0.11/3.16/6.23**, after **0.69/3.10/6.14** (1/5/15 minutes); the longer averages still include the completed gate. These are shared-container observations, not a noise-adjusted guarantee.

```sh
go run ./bench -only parallel_files -threads 1,2,4 -rounds 5 > bench/parallel_files.measurements.log 2>&1
# exit 0; all five rounds finished, same answer yes, counted allocations equal frees
```

| Runtime | Best wall time | Peak resident memory of that run |
| --- | ---: | ---: |
| Native, 1 thread | 0.744 s | 63.2 MiB |
| Native, 2 threads | 0.454 s | 63.4 MiB |
| Native, 4 threads | 0.312 s | 63.5 MiB |
| Node, sequential witness | 0.637 s | 223.6 MiB |
| Bun, sequential witness | 0.866 s | 217.7 MiB |

Four threads are 2.38x faster than one for the whole program, including generation and ordered merging. Native at one thread is 1.17x slower than Node here; native at four threads is 2.04x faster. Every round's stdout agreed. This is the new parallel program, not a regression measurement of programs that never call parallelMap.

Commits in this pass: `a2ca470` compiler merge (second parent `2e284e1`); `46bf339` Darwin feature macro and leak/TSan harness selection; `9aa6c57` native oracle integration and counts; `4bd2315` file workload and thread profiles; `6360919` `.a` rename; `136e89f` discovery, Node/Bun resolution and count-row registration; `039d402` direct Bun `.a` execution after capability verification. The final evidence commit adds this report, raw measurements and gate results. No PR was opened.
