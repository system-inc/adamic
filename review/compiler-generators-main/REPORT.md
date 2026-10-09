Built synchronous generator state machines over ordinary counted frames toward step 20, task #qk8rztp.
Docs first: a918e7fc; lowering: 5d8f670a; main ee8b5215 merged by ba05fa10; latest main b5245943 merged before delivery; delivery tip is in the handoff.
Focused Node comparisons, both backends, sanitizers, leak counts, flow, freshness, a-check, records and call-target readers pass.
All seven required mutants are caught; an eighth catches missing type metadata for a retained destructured parameter.
Opaque protocols, unsupported parameter patterns and other named suspension gaps remain refused; native tsc does not execute.

The factory retains parameters, captures and this before returning. Defaults and supported destructuring execute at the call. A resume closure owns the frame, and protocol methods own that closure. There is no owning frame-to-iterator back-reference. Yielded, returned and sent values have separate storage. Resume states handle first-argument discard, reentrancy, return and Error throw completions, yielding finally blocks, IteratorClose and delegation through arrays, Sets and known generators. Dropping the iterator releases its frame without running finally.

Fifteen independently written .a fixtures cover the requested behavior, including array-based mapIterator and checker-generator reductions. The reductions do not copy TypeScript or cohere source. Two existing regexp fixtures also agree with Node and pass sanitizers and leak checks. All 58 added or touched top-level test functions are recorded in [test-seconds.md](test-seconds.md); the final run's maximum is 10.17 seconds. A preceding uncached run's maximum was 14.67 seconds. No test is skipped. The existing, untouched records test took 65.667 seconds on one run; its later update took 55.372 seconds.

Commands ran from the repository root with /workspace/adamic-tools/env.sh, GOPROXY='https://proxy.golang.org|direct', and GOTMPDIR=/tmp/generators-go-build:

```
go test ./internal/oracle -run '^TestGenerator' -timeout 60s -count=1 -v
# PASS, 15.476 s; Node, JavaScript, native ASan/UBSan, leaks, release builds.
go test ./internal/lower ./internal/flow ./internal/fresh -run '^Test(Generator|AsyncGenerator)' -timeout 60s -count=1 -v
# PASS; separate top-level tests for all 15 flow and freshness fixtures.
go test ./internal/ir -run '^TestCallTargetReaders$' -timeout 60s -count=1 -v
# PASS, 9.487 s.
go test ./internal/oracle -run '^TestCountsAreRecorded$' -timeout 90s -count=1 -v
# PASS, 65.667 s.
go test ./internal/oracle -run '^TestCountsAreRecorded$' -timeout 90s -args -update-counts
# PASS, 55.372 s; final defaults/destructuring row refreshed.
python3 review/compiler-generators-main/a-check.py
# PASS, 16 files, 11.9 s; uses integration's a-check implementation.
python3 review/compiler-generators-main/run-mutants.py
# Six source overlays caught; mutant sources are .go.txt, never .go.
go build ./cmd/adamic
# Exit 0.
go vet ./internal/ir ./internal/lower ./internal/javascript ./internal/native ./internal/flow ./internal/fresh ./internal/oracle
# Exit 0, no findings.
```

| Mutant | Witness |
|---|---|
| Skip finally on return | TestGeneratorCancellation disagrees with Node stdout |
| Do not forward delegate return | TestGeneratorDelegate disagrees with Node stdout |
| Keep a local borrow across yield | TestGeneratorLocalBorrowMutant: ASan heap-use-after-free |
| Borrow a parameter past the caller's release | TestGeneratorParameterBorrowMutant: ASan heap-use-after-free |
| Defer a parameter default until next | TestGeneratorDefaults disagrees with Node stdout |
| Leak the first next argument into the body | TestGeneratorBasic disagrees with Node stdout |
| Admit a cycle-capable frame slot | TestGeneratorCycleFrame: LeakSanitizer |
| Forget the retained whole destructured parameter's type | TestGeneratorDestructuredParameterCycle loses its named refusal |

The source overlay tests fail only with their mutation. Their baseline leaves pass. No mutant is counted as caught by a Go or clang compilation failure. The two C ownership mutants preserve the normal program's SourceFlags and are caught by ASan. Compressed observations and mutation sources are beside this report.

Counts added 15 fixture rows, each with allocations equal to frees. The final defaults row is 29 allocations and 29 frees, after adding call-time destructuring. The two existing regexp rows changed because iterator-result value reads now own their boxed field value rather than using the former borrowed static read. regexp.a gained one retain and release (383/418), with allocations/frees still 437/437. regexp_tree.ts gained four (85/89), still 91/91. [count-comparison/RESULT.json](count-comparison/RESULT.json) compares untouched main 10328fb5 with this compiler; the compatibility oracle leaves check Node and leaks.

The additional flow and freshness probes found unrecorded generated frame writes. Those writes now have lowering sites, and freshness interprets GeneratorFrame's ordinary object literal. The same cycle proof follows the retained whole destructured parameter, including unbound fields. A first attempted destructuring probe used binding defaults that main already refuses. It was an additional probe, not an existing fixture; supported plain destructuring now proves call-time reads, while the unsupported patterns retain their diagnostics in docs/generators.md.

Local tsc replay merged compiler/scratch-tsc-entry-maplike 4cf6791a into the implementation in a detached worktree, merge ffd9a59121eb34385cae9455cce93f97fd797783. That merge is not pushed. [scratch-conflicts.json](scratch-conflicts.json) names all 12 conflicts and their resolutions. Source conflicts retain generator lowering alongside the scratch's caught-value, sparse and typed-array guards, record deletion, async support and null handling; test metadata retains current main's structure. The delivery merges with ee8b5215 and b5245943 have no conflicts. b5245943 changes only unrelated stage 1 tests.

```
bash stage3/apply.sh
# Creates /tmp/generators-tsc-adapted from the pinned TypeScript source.
GOWORK=/tmp/generators-replay.work go build -buildvcs=false -o /tmp/generators-replay-adamic ./cmd/adamic
ADAMIC_NATIVE_SPLIT=0 /tmp/generators-replay-adamic build /tmp/generators-tsc-adapted/src/tsc/tsc.ts -o /tmp/generators-tsc-0
ADAMIC_NATIVE_SPLIT=1 /tmp/generators-replay-adamic build /tmp/generators-tsc-adapted/src/tsc/tsc.ts -o /tmp/generators-tsc-1
```

Both layouts exit 1 at core.ts:356:37 with:

```
Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)
```

This is sameMap's unchecked cast and lies past the former core.ts:332:1 generator refusal. Observation: the refusal walk reaches that later cast. It does not establish that all 13 TypeScript generator bodies lower or that the full Iterable<T>-parameter mapIterator has a proved protocol. The fixture reduction uses an array whose representation is known. The replay initially stopped on missing host declarations; installing the lockfile-pinned @types/node 25.3.3, undici-types 7.18.2, @types/source-map-support 0.5.10 and source-map 0.6.1 resolved setup without editing compiler source or diagnostic overlays.

Machine: Linux AMD EPYC, cgroup cpu.max 400000 100000 (4 CPUs), nproc 5. Setup passed: Go 0.021 s, Node 0.021 s, submodules 0.065 s, markdown 0.072 s, clang 0.167 s, build 34.447 s, cache warm 34.629 s, done 34.658 s. Go 1.27.1, Node 24.19.0, clang 20.1.8. The printed environment was /workspace/adamic-tools/env.sh. The cache exhausted the filesystem during the baseline comparison and a later main merge. Removing named, old rebuildable Go cache artifacts and using /tmp for build temporaries resolved both; the failed local checkout was discarded before retrying the clean main merge. No user work was discarded.

Conservative choices and remaining design questions: opaque structural receivers and protocol replacements need a calling-convention proof; used yields whose next type excludes undefined need a resume-presence proof; arbitrary primitive throw values need represented exception completions; nested closures over body locals need a safe frame reference. These stay stopped, with source paths, rather than trusting their annotations. Async generators, cycle-capable slots without weak, unsupported parameter patterns and the other suspension paths listed in docs/generators.md also stay stopped. No full gate or whole package suite was run.

The required integration lane command was run after committing, and is run again on the final committed changes. Its complete output is recorded in lane-checks.txt.
