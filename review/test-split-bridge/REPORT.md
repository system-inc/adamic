Built: bridge test pieces and hash-keyed shared products for roadmap step 38.
Base: 54cbc125422d4e1d64c1ffe782445b2cbc2bc5b8; delivery SHA accompanies the push.
Checks: every default and configured corpus unit passes in its mode; maximum complete invocation 3.283 seconds.
Mutants: all seven existing mutants retained; eight added failure variants caught, with one native-input failure isolated to shard 3/4.
Limits: internal/buildcache is not on the observed main; native-package tests, full gate and other platforms were not run.

## Before and after

Only `TestBridge` exceeded 30 seconds in the six-test default baseline: 207.150 seconds (207.157 for its complete package invocation). The optional pristine TypeScript 6.0.3 corpus also passed on the original test, in 56.260 seconds (56.274 including package startup). The older Home measurement in the brief is a separate observation.

Measurements select each top-level test separately with `go test -json`, `-count=1`, `-parallel=4`, `GOMAXPROCS=4`, and `ADAMIC_GATE_UNCACHED=1`. After measurements use a 30-second timeout and a new process per root. `nproc` reports 5, while `cpu.max=400000 100000` gives four CPUs. Go compilation artifacts and OS page caches are retained; query observations and oracle comparisons are never cached.

| Test | Before, s | After maximum test, s | After maximum including product fetch, s | Units |
| --- | ---: | ---: | ---: | ---: |
| TestBridge sample | 207.150 | 0.170 | 1.994 | 16 |
| bridge/tsgo/TestTSGoRequiresLink | 0.040 | 0.040 | 0.053 | 1 |
| bridge/tsgo/checker/TestFactEncoding | 0.000 | 0.000 | 0.005 | 1 |
| bridge/tsgo/checker/TestShapeAndNameFacts | 0.010 | 0.010 | 0.016 | 1 |
| bridge/tsgo/checker/TestExactIndexMatchesCompilerNodes | 0.050 | 0.050 | 0.060 | 1 |
| bridge/tsgo/checker/TestAdamicRootKeepsConfigDeclarations | 0.050 | 0.050 | 0.060 | 1 |
| TestBridge compiler corpus | 56.260 | 1.260 | 3.283 | 31 |

There are 43 selectable roots, including the four unchanged checker-package tests. Default mode passes 23 and skips 20 compiler-only roots; compiler-corpus mode passes 38 and skips five sample-only roots. Every root passes in an applicable mode. Maximum test-event times are 0.170 and 1.260 seconds; maximum complete invocations are 1.994 and 3.283 seconds. No complete after invocation exceeds 30 seconds.

[before.json](before.json), [after.json](after.json), and [after-corpus.json](after-corpus.json) contain every selected command, result and raw-log path. [timings.csv](timings.csv) contains the before/after table. The corpus commit is `050880ce59e30b356b686bd3144efe24f875ebc8`, with a clean checkout and source hashes recorded in [environment.json](environment.json).

## Shared products

`TestMain` retrieves one immutable bundle before test timing. `ADAMIC_TSGO_PREPARE=1` prepares that bundle without running test units. The final preparation passed in 56.799 seconds with uncached native compilation; subsequent units fetch the completed product by its input hash. There are 19 products: ordinary and instrumented archives, four mutant archives, compiler/oracle binaries, ABI drivers, native query/region executables, a region-mutant compiler and a linkage-mutant test binary.

The bundle key is `68ae6a6782cdc60bed1d9bfa02a06de7ed4027ccce7a460f9a4acc928d8b6654`. Go validates dependency actions and supplies their build IDs; the key also includes all bridge test sources, fixtures used to compile executables, headers, build arguments, native compilation environment flags, Go/clang identities and Go compiler environment. A second key calculation rejects inputs changing during preparation. Products publish atomically under a cross-process lock. Every retrieved product is verified against its SHA-256 digest; corrupt or incomplete bundles fail. `ADAMIC_BUILD_CACHE_DIR` selects the local store, and `ADAMIC_BUILD_CACHE=off` prepares once per package process without reuse.

Origin/main was still `54cbc125` when the shared-helper decision was made, so it did not contain `internal/buildcache`. The `devtools/buildcache` branch at `64ea92b6` was read for its API but never merged or copied. The local cache is isolated in `product_cache_test.go`. Replacing its single retrieval call in `bridgeProductGet` with `buildcache.Get(buildcache.Inputs{Name: "tsgo-test-products", Flags: []string{key}}, build)` plus the import switches the backend; verification stays outside that call, and its proof test uses the same adapter. This preserves the later one-line swap requested by Compiler.

## Coverage and shard ownership

The original test has 11 non-query pieces and five query analyses per active file: sanitizer comparison, three independent timing rounds, and wrong-position mutation. Default mode therefore has 16 pieces; the four-file compiler mode has 31. Generated registrations expose each as a test root. `ADAMIC_TEST_SHARD=i/n` assigns registration index modulo n, so each active piece has exactly one owner.

`TestBridgeUnitsCoverEveryPiece` pins every registration, compiled test binding, helper identity, file and round. It checks every original query position and exercises the actual shard selector for every shard count from 1 through 40. Missing a piece, missing a query, or dropping shard filtering independently fails this pin. No original fixture, production source or checker test was changed.

All program roots remain present when a query file is split, preserving the original checker context. Fresh direct-oracle baselines in each timing and wrong-position unit avoid sharing successful observations. They add direct observations while retaining the original comparisons.

Observed before/after totals are identical: sample mode compares 162 positions and 3,261 bytes; compiler mode compares 1,600 positions and 54,982 bytes. The five query analyses cover 810 and 8,000 positions respectively. The region witness remains `6; adamic: counts: allocations 9 frees 8 retains 5 releases 13 peak 8 regions 1`. [audit.json](audit.json) records these checks and the 19 product hashes.

## Mutants

All original checks still catch their mutants: input length and output length through ASan heap-buffer-overflow; a retained handle through the stale-handle assertion; a source-file type through the Go oracle; the missing link guard through `TestTSGoRequiresLink`; the missing output free through LSan; and heap allocation in the region entry through LSan. Archive or compiler build failures are rejected during shared preparation and do not count as mutant catchers.

Added probes use Go overlays without changing the checkout:

| Probe | Catcher / result | Other passes | Skips |
| --- | --- | ---: | ---: |
| changed-native-query-0 (0/4) | all selected units pass: PASS | 6 | 32 |
| changed-native-query-1 (1/4) | all selected units pass: PASS | 6 | 32 |
| changed-native-query-2 (2/4) | all selected units pass: PASS | 6 | 32 |
| changed-native-query-3 (3/4) | TestBridgeOracleSample: native oracle mismatch | 5 | 32 |
| missing-piece (unsharded) | TestBridgeUnitsCoverEveryPiece: bridge coverage count | 0 | 0 |
| missing-query (unsharded) | TestBridgeUnitsCoverEveryPiece: query coverage count | 0 | 0 |
| drop-shard-filter (unsharded) | TestBridgeUnitsCoverEveryPiece: shard coverage | 0 | 0 |
| accept-corrupt-product (unsharded) | TestBridgeProductCacheIsVerified: corrupt product was accepted | 0 | 0 |
| accept-extra-product (unsharded) | TestBridgeProductCacheIsVerified: wrong product count was accepted | 0 | 0 |
| rebuild-product (unsharded) | TestBridgeProductCacheIsVerified: product was rebuilt | 0 | 0 |
| over-budget (unsharded) | TestBridgeABI: bridge unit exceeded 30 seconds | 0 | 0 |
| poison-build-callback (unsharded) | all selected units pass: PASS | 18 | 20 |

The planted input changes the first query from byte 0 to byte 14 only for the native sanitizer observation of the sample fixture. Both inputs are valid; the oracle comparison disagrees. Shards 0/4, 1/4 and 2/4 pass, and only `TestBridgeOracleSample` fails in shard 3/4. The coverage and cache proofs continue to pass. This establishes failure ownership independently of the count assertion.

The poisoned build callback deliberately errors if preparation runs after the fetch. All 18 selected active bridge checks pass with that callback, proving units use the fetched products. This is a reuse probe, not a killed semantic mutant. The cache proof independently catches rebuilding a product, accepting corrupt bytes, and accepting an extra product. The elapsed-clock overlay proves the unit budget assertion can fail.

## Commands and evidence

- `export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh`: passed. Timing lines: Node 0.027 s; Go 0.030 s; submodules 0.086 s; markdown dependencies 0.088 s; clang 0.187 s; build 11.076 s; deferred tests 11.332 s; warm cache 11.334 s; done 11.407 s. Environment: `/workspace/adamic-tools/env.sh`.
- `python3 review/test-split-bridge/measure.py before`: all six original tests pass, measured separately.
- Original `TestBridge` with `ADAMIC_TSGO_CORPUS=/tmp/test-split-bridge-typescript`: passes all 1,600 original positions and mutants on the pinned pristine source.
- `ADAMIC_TSGO_PREPARE=1 ADAMIC_GATE_UNCACHED=1 GOMAXPROCS=4 go test ./bridge/tsgo -run '^$' -v -count=1 -timeout=30m`: prepares the final shared products once, before tests.
- `python3 review/test-split-bridge/measure.py after` and `after-corpus`: each measures all 43 roots separately; all applicable roots pass below the budget.
- `python3 review/test-split-bridge/mutants.py`: eight failing variants caught, three non-owner shards pass, and the poisoned-build reuse probe passes.
- `python3 review/test-split-bridge/audit.py`: all coverage, byte totals, region counts, unchanged production/fixture inputs, shard ownership and product hashes verified.

[evidence.tar.gz](evidence.tar.gz) preserves JSONL test logs, overlays, product build logs and setup output. Extract it into `/tmp` to restore recorded log paths. Build binaries/archives are excluded; their hashes are in audit.json. Regenerate registrations with `generate.py`, then gofmt.

No new oracle fixtures were added, so counts.md did not change. No whole package or full gate was run: measurements selected one root per process, and mutant probes selected this unit's bridge roots. The native-package tests mentioned in the brief and other platforms remain outside this unit. Only bridge test files and review/test helpers changed. No cohere source was copied. No main, area or other worker branch was merged into or pushed.
