Measured WASI units 14 and 21 cold/warm, found shared setup dominates, and added test-only phase logging instead of splitting single-fixture units.
Commit: delivery SHA reported with compiler/wasi-grain; branch starts from current origin/main 2b1be38362046455e0e5454d8b0e674a8630e95d.
Commands: isolated -timeout 90s WASI runs, shard ownership audits and call-target reader guard passed; exact commands and observations below.
Mutants: none; no semantic check, compiler behavior, fixture, or ownership rule changed.
Not covered: Loom pool load or its exact 100.0/49.0-second delay; the full native package and full gate were not run.

# Task #psh9dvk

## Initial main measurements, before editing

Each command ran alone, sequentially, with ADAMIC_TEST_WASI=1, the installed SDK clang first on PATH, and -count=1 -v -timeout 90s. Outer timeout was 150 seconds with a five-second kill grace. Cold means a separate initially empty ADAMIC_BUILD_CACHE_DIR for each unit. The installed toolchain and Go compilation cache populated by required setup remain available; this is not a fresh OS/toolchain installation benchmark. Warm immediately repeats the unit against that same product cache.

| Unit | Fixture | Cold unit | Warm unit | Cold fixture | Warm fixture | Cold runtime build/cache | Cold compiler build/cache | Cold link |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| TestWASIUnit14 | regions.a | 7.89s | 0.33s | 0.33s | 0.17s | 5.15s | 2.22s | 0.12s |
| TestWASIUnit21 | maps_and_text.a | 9.13s | 0.43s | 0.29s | 0.19s | 5.70s | 2.98s | 0.11s |

The runtime and compiler phases already have exactly identical cache keys across these two units. Both warm runs report runtime, compiler and link hits. The cold total is dominated by the shared 58-translation-unit strict WASI runtime and compiler setup, not by fixture work.

Main's table has 35 fixtures and 36 units: each fixture index belongs to exactly its corresponding unit and the extra request benchmark belongs to unit 35. Unit 14 already owns only regions.a; unit 21 already owns only maps_and_text.a. Splitting these units cannot divide a fixture batch and would repeat setup or cache lookup/wait. No rebalance or fixture change is justified by the observations.

## Delivered diagnostics and verification

Only internal/native/wasm_test.go changes: logs now distinguish the toolchain probe, runtime-input/cache identity setup, counted/uncounted runtime build or cache wait, compiler build/cache, C generation, link/cache, source Node, and WASI execution. Both cold building and cache waiting remain included in their phase rather than being silently charged to a fixture. The counted request runtime also uses the diagnostic. Timing does not bypass any oracle comparison.

After logging was added, the same isolated cold/warm measurements passed:

| Phase | Unit 14 cold | Unit 14 warm | Unit 21 cold | Unit 21 warm |
| --- | ---: | ---: | ---: | ---: |
| Toolchain probe | 0.117s | 0.165s | 0.187s | 0.100s |
| Runtime inputs/cache identity | 0.044s | 0.058s | 0.111s | 0.066s |
| Runtime build/cache | 5.804s | 0.001s | 6.481s | 0.002s |
| Compiler build/cache | 7.279s | 0.032s | 3.253s | 0.077s |
| Generate | 0.053s | 0.087s | 0.111s | 0.089s |
| Link/cache | 0.127s | 0.002s | 0.134s | 0.002s |
| Source Node | 0.085s | 0.096s | 0.088s | 0.074s |
| WASI execution | 0.040s | 0.083s | 0.075s | 0.037s |
| Fixture subtotal | 0.31s | 0.27s | 0.41s | 0.20s |
| Whole unit | 13.55s | 0.53s | 10.44s | 0.45s |

These observed totals leave at least 16 seconds below the 30-second budget on this runner, including cold shared setup. No claim is made that the same bound holds under Loom's reported load. The exact pool overrun has not been reproduced or attributed to a particular phase; shared setup/cache contention is an inference suggested by these results. The delivered phase logs let the next pool run settle that question.

A separate cross-unit reuse observation ran unit 21 against the cache populated by unit 14: runtime and compiler were hits, link was a miss, and unit 21 passed in 0.73s with a 0.45s fixture. This demonstrates the products are shared rather than fixture-specific. Changing the test binary changes the native-test cache identity, so final diagnostic measurements use a fresh cache again instead of claiming the old products are warm.

## Exact commands and coverage

Every run sources /workspace/adamic-tools/env.sh and prepends /workspace/adamic-tools/wasi-sdk/bin to PATH for WASI commands. ADAMIC_BUILD_LOG points to the corresponding evidence file. Initial product cache directories: /tmp/wasi-grain-cache14 and /tmp/wasi-grain-cache21. Final directories: /tmp/wasi-grain-diagnostic-cache and /tmp/wasi-grain-diagnostic-cache21. The cross-unit run uses unit 14's final cache.

```
ADAMIC_TEST_WASI=1 timeout --kill-after=5s 150 go test ./internal/native -run '^TestWASIUnit14$' -count=1 -v -timeout 90s
ADAMIC_TEST_WASI=1 timeout --kill-after=5s 150 go test ./internal/native -run '^TestWASIUnit21$' -count=1 -v -timeout 90s
timeout 150 go test ./internal/native -run '^Test(ShardSelection|RetainedSplitCoverage|RetainedTopLevelCoverage)$' -count=1 -v -timeout 90s
timeout 180 go test ./internal/ir -run TestCallTargetReaders -count=1 -timeout 90s
```

All cold and warm WASI runs passed without skips. Ownership audits passed in 0.047s, checking the retained fixture membership digest, top-level unit/helper indices and exact-one-owner shard union. Reader guard passed in 9.564s. No new fixture or test leaf was added; counts.md is unchanged and a counts refresh is not applicable. Fixture membership/order and shard modulus remain unchanged, preserving all 35 fixtures plus the request benchmark exactly once. No compiler or runtime production file was edited.

Integration lane checks run after commit and before push with the required command:

```
git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -
```

The delivery response reports its output. Evidence logs and timings.json are under review/compiler/wasi-grain/.

## Toolchain setup

GOPROXY=https://proxy.golang.org|direct was exported before bounded cloud/setup.sh. Native setup completed in 44.547s: Node 0.047s, Go 0.050s, markdown 0.099s, submodules 0.117s, clang 0.202s, build 44.364s, cache 44.511s. WASI SDK was absent, so bounded cloud/setup.sh --wasi-sdk installed SDK 27 and completed in 14.751s: Node 0.035s, Go 0.041s, submodules 0.115s, markdown 0.202s, clang 0.208s, WASI SDK 4.574s, build 14.516s, cache 14.664s. Exact timing lines are in setup.log and sdk-setup.log. Go 1.27.1; Node 24.19.0; clang 20.1.8; SDK clang 20.1.8-wasi-sdk. nproc=5; cpu.max=400000 100000 (four CPU equivalents).
