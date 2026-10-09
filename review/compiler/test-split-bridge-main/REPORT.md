Split TestBridge into 36 independently selectable tests for step 79, task #jy0v0am.
Use main's internal/buildcache.Get and lazy sync.Once products, with no batch 1 dependency.
Run every leaf separately, with an empty product cache and setup included; all passed below 60 seconds.
Catch planted shard, count, digest, rebuild and budget failures; retain every original semantic mutant.
Omit 3ebd13ea because main has no internal/skipcensus; no product changes or new Adamic fixtures.

The original TestBridge measured 397 seconds in Loom's supplied measurement. It is replaced by 36 roots: 11 shared checks and five analyses for each of sample, checker, parser, types and utilities. Sample mode runs 16 checks; compiler mode runs 31. TestBridgeUnitsCoverEveryPiece freezes the names, helpers, modes, rounds, positions and bindings, and proves exactly one shard owner for each applicable check for every shard count from 1 through 40.

Measurements use GOMAXPROCS=4 on Linux amd64. nproc reports 5; cgroup cpu.max is 400000 100000, a four-CPU quota. Go is 1.27.1, clang is 20.1.8, Node is 24.19.0. cloud/setup.sh succeeded; its timing lines are in bridge-main-setup.log. The initial sanitized archive build was included in TestBridgeABI's 45.65-second test elapsed time. The table uses complete go test invocation wall time, including lazy builds, with a new empty ADAMIC_BUILD_CACHE_DIR for each row. Go's dependency and toolchain caches were prepared by setup. No parallel measurements ran on the instance.

Each row ran go test ./bridge/tsgo -run '^<test>$' -count=1 -json -timeout=90s with ADAMIC_UNIT_BUDGET=1. Compiler rows also set ADAMIC_TSGO_CORPUS to the clean TypeScript 6.0.3 checkout at 050880ce59e30b356b686bd3144efe24f875ebc8. Budget enforcement begins before product setup. Other machines only log the budget; subprocesses retain a five-minute hang guard.

| Test | Mode | Invocation seconds | Test seconds |
|---|---|---:|---:|
| TestBridgeABI | sample | 4.741 | 2.65 |
| TestBridgeInputLength | sample | 4.518 | 2.60 |
| TestBridgeUnlinkedBuild | sample | 4.555 | 2.73 |
| TestBridgeUnlinkedC | sample | 4.593 | 2.66 |
| TestBridgeUnlinkedJavaScript | sample | 4.572 | 2.72 |
| TestBridgeOutputLength | sample | 8.180 | 6.31 |
| TestBridgeStaleHandle | sample | 8.138 | 6.04 |
| TestBridgeLinkage | sample | 4.636 | 2.66 |
| TestBridgeOutputFree | sample | 8.034 | 6.10 |
| TestBridgeRegion | sample | 14.966 | 12.94 |
| TestBridgeRegionOwnership | sample | 15.549 | 13.69 |
| TestBridgeOracleSample | sample | 14.038 | 12.12 |
| TestBridgeTimingRound1Sample | sample | 11.645 | 9.84 |
| TestBridgeTimingRound2Sample | sample | 12.590 | 10.47 |
| TestBridgeTimingRound3Sample | sample | 12.291 | 10.34 |
| TestBridgeWrongPositionSample | sample | 17.594 | 15.74 |
| TestBridgeOracleChecker | corpus | 14.674 | 12.73 |
| TestBridgeTimingRound1Checker | corpus | 13.607 | 11.68 |
| TestBridgeTimingRound2Checker | corpus | 13.164 | 11.08 |
| TestBridgeTimingRound3Checker | corpus | 12.334 | 10.48 |
| TestBridgeWrongPositionChecker | corpus | 14.814 | 12.90 |
| TestBridgeOracleParser | corpus | 14.972 | 13.01 |
| TestBridgeTimingRound1Parser | corpus | 12.598 | 10.54 |
| TestBridgeTimingRound2Parser | corpus | 12.473 | 10.66 |
| TestBridgeTimingRound3Parser | corpus | 12.644 | 10.53 |
| TestBridgeWrongPositionParser | corpus | 14.551 | 12.70 |
| TestBridgeOracleTypes | corpus | 15.455 | 13.55 |
| TestBridgeTimingRound1Types | corpus | 12.497 | 10.55 |
| TestBridgeTimingRound2Types | corpus | 12.537 | 10.60 |
| TestBridgeTimingRound3Types | corpus | 12.907 | 10.92 |
| TestBridgeWrongPositionTypes | corpus | 14.988 | 12.92 |
| TestBridgeOracleUtilities | corpus | 14.904 | 12.96 |
| TestBridgeTimingRound1Utilities | corpus | 12.607 | 10.50 |
| TestBridgeTimingRound2Utilities | corpus | 12.220 | 10.35 |
| TestBridgeTimingRound3Utilities | corpus | 12.216 | 10.35 |
| TestBridgeWrongPositionUtilities | corpus | 14.093 | 12.22 |
| TestBridgeUnitsCoverEveryPiece | sample | 1.948 | 0.01 |
| TestBridgeUnitsCoverEveryPiece | corpus | 1.831 | 0.01 |
| TestBridgeProductCacheIsVerified | sample | 2.003 | 0.00 |
| TestTSGoRequiresLink | sample | 2.088 | 0.04 |

Old checks map to these tests. X means each of Sample, Checker, Parser, Types and Utilities, selected by corpus mode. All program roots remain in every file-specific invocation.

| Old TestBridge check | New test |
|---|---|
| C ABI: 100 repeated queries, exact-length inputs, outputs surviving release, cleared frees, invalid positions, zero/stale handles, double release, distinct new handle | TestBridgeABI |
| Input view length off by one, ASan heap-buffer-overflow | TestBridgeInputLength |
| Unlinked build command refused | TestBridgeUnlinkedBuild |
| Unlinked C command refused | TestBridgeUnlinkedC |
| Unlinked JavaScript command refused | TestBridgeUnlinkedJavaScript |
| Full manifest equals independent Go checker under ASan/UBSan/LSan | TestBridgeOracleX |
| Optimized native and direct Go checker match baseline in rounds 1, 2, 3 | TestBridgeTimingRound1X, TestBridgeTimingRound2X, TestBridgeTimingRound3X |
| Output string length plus one, ASan heap-buffer-overflow | TestBridgeOutputLength |
| Released handle retained, stale-handle assertion | TestBridgeStaleHandle |
| Type from source-file node, Go oracle mismatch | TestBridgeWrongPositionX |
| Link opt-in guard removed, lowering refusal test fails | TestBridgeLinkage |
| C output free removed, LSan leak report | TestBridgeOutputFree |
| UTF-16 lengths match Go oracle, exactly one region | TestBridgeRegion |
| Region result allocated on heap, LSan leak report | TestBridgeRegionOwnership |

union.json records source hashes and output hashes: 3,261 sample bytes and 54,982 compiler bytes. The old whole-manifest Go output equals concatenated per-file Go output and whole-manifest sanitized native output. The coverage count also preserves all 400 original byte positions per compiler file and every sample byte.

Additional planted failures:

| Mutant | Check that catches it |
|---|---|
| Remove ABI case | TestBridgeUnitsCoverEveryPiece: bridge coverage count |
| Remove one query position | TestBridgeUnitsCoverEveryPiece: query coverage count |
| Disable shard ownership filter | TestBridgeUnitsCoverEveryPiece: shard coverage |
| Disable bundle digest verification | TestBridgeProductCacheIsVerified: corrupt product was accepted |
| Disable bundle product count | TestBridgeProductCacheIsVerified: wrong product count was accepted |
| Rebuild cached bundle | TestBridgeProductCacheIsVerified: product was rebuilt |
| Set unit start 61 seconds earlier | TestBridgeABI fails only with ADAMIC_UNIT_BUDGET=1; with 0 it passes and logs |
| Change only OracleSample's native manifest position 0 to 14 | Only TestBridgeOracleSample fails, only on shard 3/4; TimingRound1Sample passes on shard 0/4; other shards skip |
| Corrupt cached API executable | TestBridgeABI refuses product digest: api; restored executable passes |

The original input length, output length, stale handle, wrong position, linkage, output free and region ownership mutants ran in their corresponding green leaves; each requires the specific sanitizer report, assertion or oracle mismatch, rather than accepting a build failure as evidence.

Provenance: bridge-only changes from e1137975 and the conditional-budget/hang-guard changes from 93cd7227 were applied on main. The LF evidence and invocation-budget requirements from 4f625473 were retained in new measurements; no pycache is included, following c48b8c3d. Neither f45b69dd nor the unlanded shared-buildcache commits are dependencies. Main's buildcache adapter uses only Get and Inputs. Products are keyed by Go dependency BuildIDs, bridge test and fixture bytes, tools, flags and relevant build environment; the single-product digest is verified before execution.

No whole package or full gate was run. No new .a fixture was added, so oracle counts did not change. The measured platform is Linux amd64; other platforms and an empty Go/toolchain cache were not measured. Exact commands and raw logs are included in results JSON and compressed JSONL evidence. The integration lane output is in lane-checks.log.

The complete table was measured on main 7f409c9f. The delivery base advanced to 153db5b1 afterward through test-only changes outside bridge/tsgo, without conflicts. An additional ABI invocation on that base passed in 12.051 test seconds, including archive and driver builds. The bridge test source hashes are unchanged.
