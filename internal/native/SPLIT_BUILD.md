# Split native build integration

Base: `runtime/area-take-batch` at `6c891cca313e614ae46de794814bc7c0db4cae01`.
Destination: `runtime/split-build`. No main push or force push.

## Merges and review

None of the requested branch tips was already in the area. Merged in requested order:

| Branch | Incoming tip | Merge |
| --- | --- | --- |
| codex/stable-emitter-names | aca41891 | bac02682 |
| devtools/stable-splitter | 99814057 | 94fd5756 |
| codex/outline-module-main | 5bbc282d | 22e8350e |

Compiler review approved each conflict resolution before its merge and the subsequent build-path changes. Resolutions retain the area's graph ownership, initialized slot cache, captured-cell readiness, conservative promise handling, async scheduler setup and draining, and global cleanup. Stable module ownership supersedes the area's groups of sixteen consecutive functions: it has stable source identity and selective shared declarations, rather than a giant repeated generated header. Outlined initializers compose with this ownership and retain source module order. Async startup remains in main.

## Cold sanitized probe

Probe: `stage1/cohere/markdownblocks/testdata/ast_probe.ts`, the native probe used by the port's tests. Setup used `GOPROXY=https://proxy.golang.org|direct`, `bash cloud/setup.sh`, `/workspace/adamic-tools/env.sh`, and Node v24.19.0.

Box: 5 available CPUs (`nproc` and Go `runtime.NumCPU`), cgroup quota 4 CPU equivalents (`cpu.max=400000 100000`), clang 20.1.8, LLVM commit `87f0227cb60147a26a1eeb4fb06e3b505e9c7261`, x86_64 Linux. Each round sets `ADAMIC_GATE_UNCACHED=1` and a new `XDG_CACHE_HOME`. Loading, lowering and C emission precede the timer; the timer includes a cold sanitized runtime archive, generated-program compilation and linking. Best of three, seconds:

| Build | Round 1 | Round 2 | Round 3 | Best | Runtime at best | Program and link at best |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Area, single unit | 11.801 | 10.610 | 11.275 | 10.610 | 7.994 | 2.616 |
| Merged, split, one job | 18.403 | 15.525 | 14.263 | 14.263 | 9.051 | 5.211 |
| Merged, split, five jobs | 9.691 | 9.681 | 10.183 | 9.681 | 7.655 | 2.025 |
| Runtime parallelism, development measurement | 4.875 | 4.717 | 4.805 | **4.717** | 2.739 | 1.978 |

| Committed defaults, parallel runtime and units | 7.111 | 6.806 | 5.357 | **5.357** | 3.109 | 2.248 |

Best-round load averages (1/5/15 minutes), before to after: area `3.97/7.77/8.35` to `3.51/7.55/8.27`; split one job `1.23/4.91/7.15` to `1.20/4.78/7.09`; split five jobs `1.16/4.66/7.02` to `1.14/4.54/6.96`; development `1.48/2.94/5.92` to `1.84/2.99/5.92`; committed defaults `5.65/6.36/7.30` to `5.60/6.34/7.28`. Raw JSON records carry every round's loads and exact clang version. No verification workloads ran during measurement.

## Changes beyond merges

Sanitized and counted generated native programs now split by default. Zero jobs selects `ADAMIC_NATIVE_JOBS` or the machine CPU count; explicit jobs takes precedence. `ADAMIC_NATIVE_SPLIT=0` retains an opt-out for baseline comparisons. Shipping release/profile builds and foreign targets keep their original single-unit path and ThinLTO flags. Both ordinary and checker-archive build dispatches follow this policy. Handwritten C harnesses retain their existing build path, because their macro/declaration contracts exceed the emitted-C splitter's scope.

Cold runtime archive compilation dominated the five-job probe: 7.655 of 9.681 seconds. The final build compiles its existing C files concurrently with the same job limit, preserves source/archive member order, cache keys, flags and atomic publication, and joins every worker before reporting failure or publishing. No runtime file or program behavior changes. A focused linked-state test holds archive order and shared state.

The splitter recognizes async frame typedefs as shared types rather than state definitions and keeps external parallel runtime ABI prototypes in the common header without renaming. The outliner recognizes area-added typed-array, promise and helper types so live values receive address parameters across chunk boundaries; a linked sanitized typed-array regression holds its state and cleanup. Async normalization clears module boundaries when Main moves into the synthetic async entry. The async abandonment test discovers its stable emitted callable instead of constructing the obsolete ordinal symbol; all control and leak-mutant assertions remain.

## Evidence

Raw output recordings, comparator, mutation runner and measurement helper are under `split_build_evidence/`. The recording overlay preserves ordinary oracle assertions and records stdout/stderr bytes and exit codes for every execution. Its comparator decodes recorded bytes without normalization.

The module-order mutant reverses outlined initializer order. Compilation and linking succeed, then `TestNativeAgreesWithNode/internal/oracle/testdata/import_cycles/order/main` fails on stdout: expected `shared,c,b,a,inline,main` lines; mutant emits `main,inline,a,b,c,shared` lines. Both exits are zero and stderr is empty.

Implementation revision: `ada33dcf`.

The area whole-oracle run passed in 1492.666 seconds. The integrated whole run reached its 30-minute limit without an assertion failure. Its 85 completed roots were retained; all remaining scopes passed separately with the same assertions, sanitizers and uncached settings: object fixture 243.362 seconds, twelve remaining roots (including counts) 400.226 seconds, remaining parallel-files schedules 448.244 seconds, and object schedules 216.357 seconds. `assemble.py` requires successful completion logs and replaces only those incomplete scopes.

Both recordings contain exactly 9,172 execution keys, with no additions or omissions and identical exit codes. All 4,113 normal fixture executions across 797 contexts have byte-identical stdout and stderr, including both backends and sanitized variants. Across the entire suite, 9,141 raw records are identical and 31 differ. The strict comparator reports failure for those 31; literal whole-record byte parity did not pass. No normalization was applied. `parity.json` lists every affected key and field.

The 31 differences comprise 14 intentional-fault sanitizer diagnostics (process/address/debug/cache paths), three invalid-mutant stdout records (stale/undefined payloads), three RSS/heap measurement records, five argument-test records echoing independently generated temporary filenames, and six transport/timing harness records (temporary executable paths, inspector UUID or elapsed time). Ordinary program results are unchanged. Compiler review explicitly approved the normal-program evidence with this limitation disclosed.

`TestCountsAreRecorded` passed. No oracle fixture was added; counts.md is unchanged. Flow and lowering suites passed, as did internal-package vet and focused split/runtime/async/outliner checks. Whole-repository vet remains blocked by the pre-existing undefined `volumeGenerated` in `stage1/cohere/lint/shipped_profile_test.go:47`, identical in the area baseline.


Native package coverage is complete across the initial run's 26 successful roots and an uncached completion run passing in 445.266 seconds; optional environment-gated tests retain their normal skips (listed in native-coverage.json). The initial whole native run hit its 30-minute limit during signal checks and its lookup performance bound failed under concurrent verification load. The completion run passed both those checks; the isolated lookup used 0.037217 CPU seconds against the 0.35 bound. Both logs are retained instead of presenting the initial run as a pass.
