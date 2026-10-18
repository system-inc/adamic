Compared: the entire stage1/cohere/json package on exact base dbdf977f5b and V4 82a5dffd, sequentially on one four-CPU quota.
Commits: base dbdf977f5b8b0979d414ebe6f20e7aadd312e309; V4 82a5dffd7654f3eb2cb4652a12b99401210f442a. No compiler fix is asserted by this report.
Commands and outputs: both full-package runs pass with count=1 and JSON output; every ordinary agreement shard runs on both revisions.
Mutants: the package's existing port, width, comparison and corpus mutation checks ran as part of the complete package; no new compiler check was introduced.
Not covered: the remote Lunatic failure mechanism, its five unnamed shards as remote identities, optional external/profile lanes, or a full repository gate.

Both complete package runs use GOMAXPROCS=4, -parallel=4, -count=1, -json, -timeout=30m, and an outer 1860 s bound. They run one after the other, with no overlapping tests or compilation on this box. No -run selector or sampling variable filters the package. Both source trees have real pinned cohere and nested TypeScript worktrees, with compiler code taken from their own exact revision. No cohere implementation was copied. The box's cgroup cpu.max is 400000 100000 and memory.max is 17179869184; nproc reports five available CPUs but the quota is four. configuration.json records the non-secret selection settings, all unset. In particular, ADAMIC_TEST_SHARD and ADAMIC_JSON_MEMORY_UNITS are unset.

The separately owned corpus and Go oracle products were settled first on each tree. The owner phase uses the same GOMAXPROCS=4 and -parallel=4; commands.jsonl gives every command, exit and wall time. The full package commands are exactly:

    GOMAXPROCS=4 timeout 1860 go test ./stage1/cohere/json -count=1 -json -parallel=4 -timeout=30m

Outputs are base-package.jsonl and v4-package.jsonl, not a pipeline. analyze.py reads every JSON event into base-tests.json and v4-tests.json and emits shard-times.csv. That CSV includes every shard and a column for a base-below-60/V4-at-least-60 crossing. No completed or over-budget test was rerun to obtain these package results.

| Full package result | Exact base | V4 |
|---|---:|---:|
| Exit | 0 | 0 |
| Go package seconds | 304.514 | 339.365 |
| Passed test results, including nested tests | 2121 | 2122 |
| Skipped test results | 5 | 5 |
| Ordinary agreement shards passed | 2048 | 2048 |
| Failed shards | 0 | 0 |
| Assertion failures | 0 | 0 |
| Timeouts | 0 | 0 |
| Test leaves at or above 60 s | 0 | 0 |
| Slowest ordinary shard | _1643, 14.19 s | _1226, 15.62 s |
| Slowest test overall | TestCachedWidthMutantIsCaught, 15.50 s | TestPortMatchesGoCohereClassOrder_000, 37.11 s |

Every shard failing on either revision: none. There is no local output-mismatch input to reduce or print. This is an observed result on this box, not a claim that either remote red was noise or that Lunatic is fixed. A complete remote diagnostic and command/flags were requested while the measurements continued; test names alone do not identify whether those remote results were assertion, timeout or resource failures.

| Previously reported shard | Base seconds | V4 seconds | Result |
|---|---:|---:|---|
| _1894 | 0.29 | 0.33 | both pass |
| _1896 | 0.34 | 0.35 | both pass |
| _1898 | 0.35 | 0.29 | both pass |
| _2036 | 0.26 | 0.27 | both pass |
| _2040 | 0.32 | 0.32 | both pass |
| _2041 | 0.38 | 0.30 | both pass |

These are full-package observations with ordinary contention, not individually selected reruns. The five unnamed remote shard identities are unknown, but all 2048 local shard names were exercised on both revisions. Both runs use the pinned corpus: 3919 cases before two dedicated cases are partitioned out, 3917 ordinary cases exactly once across 2048 hash buckets. Each ordinary shard compares Go oracle, source Node, release native, ASan/UBSan/LeakSanitizer native, and JavaScript-backend results. The class-order and inventory owners also pass.

The 34.851 s total package difference is not a matched cold-compilation benchmark. Base fetched its existing port products; V4 rebuilt its newly keyed lowering and native products. V4 logged 1.54 s for lowering, 14.37 s for release compilation and 8.08 s for sanitizer compilation, and its first dedicated test took 37.11 s rather than base's 12.63 s. Fresh compilation measurements below treat both versions alike.

| Sum of ordinary-shard backend wall times | Base s | V4 s |
|---|---:|---:|
| Go oracle | 53.921 | 54.405 |
| Source Node | 362.445 | 363.114 |
| Native release | 47.572 | 48.038 |
| ASan/UBSan/LeakSanitizer | 212.617 | 215.495 |
| JavaScript backend | 168.061 | 200.720 |

These sums overlap in time, include process startup and contention, and are not package wall time or CPU time. They show a larger change in the JavaScript backend component than in native release or sanitizer execution. Shards whose observed V4 time rose past 60 s: none, including all dedicated owners and other package tests.

The five existing optional skips on each revision are TestUpstreamNumericSeparatorsAgree, TestExternalComparisonCatchesThreePrinterMutants, TestProfileSnapshotsAgree, TestProduct_JsonFindsUpstreamParityVerdict, and TestUpstreamRepositoryCorpusParity. ADAMIC_JSON_PRETTIER and profile snapshots were not configured. No ordinary native agreement shard was skipped.

Fresh port compilation uses the identical measure-port.go.txt driver built separately against each revision's compiler packages. It loads the port, lowers it, emits C, and invokes native.Build for fresh release and sanitizer executables, with the runtime archives already warmed by the full tests. Temporary Go driver files were created only in hidden scratch directories and removed after compilation. No compiler-consuming package is added to the repository. Compiler binaries and large raw protocol/output files are local under artifacts/, excluded from the source commit; hashes, byte counts, timing records and all measurement sources are retained outside that directory.

| Fresh phase | Base s | V4 s |
|---|---:|---:|
| Load | 0.0612 | 0.0542 |
| Lower | 0.5409 | 0.5656 |
| Emit C | 0.1183 | 0.1625 |
| clang release build | 14.1492 | 14.0236 |
| clang sanitized build | 7.6841 | 7.9648 |
| C bytes | 1,878,008 | 2,033,942 |

These are single fresh compilation observations, not a statistical claim. The flags and runtime entry points match the real compiler. Load/lower/emit and clang phases are measured separately, and no test process runs alongside them.

The diagnostic corpus export ran after both full-package commands, restored its temporary _test.go immediately, and exported the exact 3919 pinned cases without sampling. Its 4.12 s test leaf passed. Release benchmarks run these cases in one process per observation, alternating base-first and V4-first over three rounds, with source Node's full stdout as the truth. These are bulk measurements, not new gate leaves.

One exploratory baseline bulk sanitizer run exceeded its 180 s process bound. It was killed by the benchmark's hard timeout, and measurement-driver.log retains the exception. This was not a failed package test or failed hash shard: all sanitizer hash-shard observations on both revisions already completed successfully. It was not rerun. bulk-sanitizer-limit.json records the limit; this diagnostic was not used to infer a V4 sanitizer regression. The matched sanitizer numbers are the complete package's per-shard observations above.

Setup: GOPROXY='https://proxy.golang.org|direct'; timeout 300 bash cloud/setup.sh; source the printed /workspace/adamic-tools/env.sh. setup.log retains Node/Go ready 0.023 s, markdown 0.062 s, submodules 0.073 s, clang 0.161 s, shared cache 1.156 s, Go build 14.313 s, deferred tests 14.382 s, warm cache 14.383 s, done 14.407 s. Versions: clang 20.1.8, Go 1.27.1, Node 24.19.0.

Final runtime observations:

| Workload | Base s | V4 s |
|---|---:|---:|
| Native release, median of three complete-corpus runs | 33.1375 | 32.7720 |
| JavaScript backend, one complete-corpus run each | 25.4655 | 48.1480 |
| Backend startup, empty protocol, median of ten precise samples | 0.037075 | 0.037802 |

Native release samples are base 32.3521 / 33.3372 / 33.1375 s and V4 32.7134 / 32.7720 / 32.8718 s. In these samples V4 native release is not slower. Native compilation likewise has no large observed increase. The JavaScript backend is materially slower on the full-corpus throughput sample (about 1.89 times base), with little startup difference. The bulk backend comparison is one observation per revision; the complete shard run independently shows increased backend wall-time sums. These results quantify a backend performance cost, not an observed local gate-budget violation.

All completed complete-corpus observations emit exactly 300,372,755 stdout bytes with SHA256 8571e190419a2e27c5796b3f1ff6485cfc824ddf161fcf48e9b3a55739ac614f, matching source Node. Every completed process exits zero with empty stderr. Raw protocol and output files remain under artifacts/. run-times.json, javascript-times.json, javascript-startup-precise.json, performance-summary.json and measurement-input.json retain timing, corpus and output identities. The initial file-output startup samples had coarse process-wait polling resolution; the final startup medians above use pipe completion and are retained separately as the precise series.

The emitted JavaScript diff identifies a concrete hot-path difference: V4 wraps scalar equality operands in adamicViewAdapterUnderlying, whose implementation performs an adapter-root WeakMap lookup even for numbers and strings. For example, character-code comparisons to 32 and 160 gain lookups on both sides. javascript-diff.patch retains the full emitted diff. This is distinct from a native lowering output mismatch.

An emitted-program-only experiment changes that helper to bypass the WeakMap lookup for primitive values, retaining it for objects and functions. It runs the same 3919-case protocol in 27.5511 s, versus the original V4 backend's 48.1480 s, with exactly the same source-Node output hash, exit zero and empty stderr. identity-ablation.patch and identity-ablation.json record the single helper change and observation. The original emitted program and all compiler sources remain unchanged. This experiment supports the primitive identity lookup as a major contributor to the local JavaScript throughput regression. It is not a proof of the remote red's cause, nor a committed compiler fix.

No local shard output differs only on V4, so the requested lowering-miscompile reduction/fix condition was not triggered. No local shard crosses the 60 s budget on V4, so there is no such shard list to supply beyond the empty result in package-summary.json and shard-times.csv. No compiler/loom-r-v4-3 compiler change or push was made. The previous compiler commit 82a5dffd is unchanged. The two exact comparison worktrees have no tracked modifications; source-provenance.json records their commits and only the provisioned node_modules symlink as untracked input. Counts and runtime C are unchanged. The remote Lunatic red remains unclassified without its complete failure diagnostic and invocation; neither repeated red is dismissed as noise.
