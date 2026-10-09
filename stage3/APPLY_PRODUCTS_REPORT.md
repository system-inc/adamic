# Adapted TypeScript build products

Four products in `stage3/applyproducts` replace the monolithic apply miss. This changes `stage3/apply.py`, not only tests. Each top-level `TestProduct_Stage3AdaptedNN` and the Python CLI use the same Go recipe, through `buildcache.Product` or its tool interface `buildcache.Get`.

The boundaries are after adaptations 10, 40, and 70; the final product applies the remainder and packages the tree. Intermediate products carry the pristine snapshot, last snapshot, incremental table rows, stock parser install, and complete git snapshot refs. Intermediate tables are not written into the tree, so they cannot contaminate subsequent snapshots. Consumers restore the final verified archive. `--build-product` also uses this chain and copies its already packaged payloads for the existing publication hook.

## Cold measurements

Machine: `nproc=5`, cgroup `cpu.max=400000 100000` (4 CPUs). Go 1.27.1, Node 24.19.0, clang 20.1.8. Product, source mirror, and npm caches were empty. The installed toolchain and OS page cache were not flushed. Every measurement had a hard deadline before launch.

Setup command: `export GOPROXY='https://proxy.golang.org|direct'; timeout 600 bash cloud/setup.sh > /tmp/stage3-products-setup.log 2>&1`, followed by `source /workspace/adamic-tools/env.sh`. Timing lines: Go ready 0.078 s; Node ready 0.098 s; clang ready 0.473 s; markdown install step 1.323 s and ready 1.531 s; submodules ready 25.234 s; Go build ready 262.668 s; build cache warm 262.813 s; done 262.871 s.

The first measurement overlapped setup and was killed at 240 s during adaptation 71, exit 124. It is not the sizing measurement. The complete uncontended run used `timeout 420 python3 /tmp/stage3-measure-products.py uncontended > /tmp/stage3-products-uncontended.log 2>&1`. That scratch harness wraps subprocess calls in the original apply with a monotonic clock and fresh source/npm/product directories. It records every command, including checkout, snapshots, npm, adaptations, and tar. No separate tsc build runs during apply; semantic checking happens inside the adaptations.

| Before step | Cold seconds |
|---|---:|
| source fetch | 8.112 |
| source checkout | 3.571 |
| pinned checkout | 0.573 |
| 00-setup | 0.129 |
| API npm ci | 0.766 |
| 00-setup | 0.121 |
| 10-type-imports | 10.442 |
| 20-optional-declarations | 30.414 |
| 30-indexed-reads | 0.874 |
| 31-indexed-reads-checker | 1.945 |
| 32-indexed-reads-program | 1.348 |
| 33-indexed-reads-emit | 0.535 |
| 40-explicit-any | 1.897 |
| 41-explicit-any-remaining | 0.807 |
| 42-scanner-any | 0.267 |
| 43-any-returns | 0.454 |
| 45-regex-captures | 1.282 |
| 46-fix-pragma-empty-argument | 0.484 |
| 47-host-errors | 1.853 |
| 48-memoize | 0.323 |
| 50-temporary-scanner-implicit-returns | 0.266 |
| 51-temporary-scanner-fallthrough | 0.037 |
| 60-temporary-factory-local-symbol | 0.746 |
| 61-temporary-factory-type-expression | 0.711 |
| 62-temporary-tracing-write | 0.372 |
| 63-temporary-tracing-legend | 0.301 |
| 64-temporary-parser-range-read | 0.560 |
| 65-temporary-node-builtins | 0.676 |
| 70-readonly-views | 34.346 |
| 71-writable-views | 5.237 |
| 75-optional-widening | 2.682 |
| 76-truthful-casts | 0.549 |
| Other apply command | 0.005 |
| Git pin verification, snapshots, refs and diffs | 10.491 |
| Complete build | 123.298 |
| Packaging, including hashes and parts | 8.382 |
| Full tree copy for handoff | 0.766 |
| Complete build, packaging and copy | 132.445 |

The table reports the entire top-level test time; builder-only times were 31.11, 43.59, 53.79, and 22.92 seconds. Each after measurement starts one fresh Go process, with a missing product and already prepared predecessors. Product 10 starts with no source mirror or npm cache and fetches the pin from GitHub. Product 99 includes packaging. Each command is `ADAMIC_BUILD_CACHE_DIR=/tmp/adamic-gate/stage3-products-final timeout 90 go test ./stage3/applyproducts -run '^TestProduct_Stage3AdaptedNN$' -count=1 -timeout 90s -v > /tmp/stage3-products-finalNN.log 2>&1`.

| After product | Work | Cold seconds | Margin below 60 s |
|---|---|---:|---:|
| Stage3Adapted10 | Fetch source, generate diagnostics, install parser, apply 00 and 10 | 31.28 | 28.72 |
| Stage3Adapted40 | Copy predecessor, apply 20 through 40 | 43.71 | 16.29 |
| Stage3Adapted70 | Copy predecessor, apply 41 through 70 | 53.90 | 6.10 |
| Stage3Adapted99 | Copy predecessor, apply 71 through 76 and package | 23.08 | 36.92 |

The intermediate tree copies use ordinary independent files, so later adaptations cannot alter a predecessor. Keys include cumulative adaptation directories, the pinned source manifest, API manifest/lock, Python and Go recipes, Node/npm/Python/git/tar versions, Go runtime, build flags, Node options, locale/timezone, npm registry, and parser overrides. Source-mirror and cache directory locations are absent from keys; a supplied source mirror must resolve the pinned tag to the exact commit. npm reads separate empty user/global config files. The detached Python fetch runs Go with `GOWORK=off`, so it does not require the cohere submodule.

## Equality and checks

The original unmodified cold apply and the final product have the same whole worktree hash:

`f2fabd4de49a6f02fc8c3b1ebd7479744d1b4085d1d1caea307429010c23c899`

The independent comparison hashes every non-.git path, mode, directory, symlink target, and file byte in sorted order. All adaptation snapshot refs also match exactly. Clone-local `.git` configuration and reflog timestamps are inherently different between fresh applies; those metadata bytes are excluded, not compiler source or generated outputs. Commands: `timeout 90 python3 /tmp/stage3-compare-products.py`, logs `/tmp/stage3-products-hash.log` and `/tmp/stage3-products-hash-mutant.log`.

Focused Python cache/shard cases: `timeout 90 python3 stage3/test_apply.py ProductTests ProofShardTests`, exit 0, 17 cases in 9.012 s. Detached checkout: `ADAMIC_BUILD_CACHE_DIR=/tmp/adamic-gate/stage3-products-final timeout 90 python3 stage3/test_apply.py ApplyCheckoutTests.test_ordinary_apply_leaves_git_clean`, exit 0, 14.713 s. The four key checks and process-group deadline proof passed in 0.790 s, using `-run '^(TestStage3Adapted.*TracksInputs|TestProductDeadlineKillsDescendants)$'`. The four key checks can also run with `go test ./stage3/applyproducts -run '^TestStage3Adapted(10|40|70|99)TracksInputs$' -count=1 -timeout 90s -v`; they pass and compare keys after copying the inputs to a different root, preserving symlinks and modes.

| Mutant | Catcher |
|---|---|
| Product 10 drops source.json | TestStage3Adapted10TracksInputs: changed source pin reused product key |
| Product 40 drops source.json | TestStage3Adapted40TracksInputs: same assertion |
| Product 70 drops source.json | TestStage3Adapted70TracksInputs: same assertion |
| Product 99 drops source.json | TestStage3Adapted99TracksInputs: same assertion |
| Kill only the worker parent at its deadline | TestProductDeadlineKillsDescendants: deadline left the worker's child running |
| Omit prepared parser API exposure | ProductTests.test_fetch_exposes_prepared_parser: prepared parser was not exposed |
| Append bytes to real reference core.ts | Whole-tree equality assertion |
| Set STAGE3_APPLY_MUTANT=1 to force --write-table | Detached checkout cleanliness assertion |

Each input-drop mutant runs as its own `go test` invocation with `-timeout 90s`; the recipe is restored afterward. Logs: `/tmp/stage3-products-mutant-input-NN.log`, `/tmp/stage3-products-mutant-parent-only.log`, `/tmp/stage3-products-mutant-api-exposure.log`; aggregate `/tmp/stage3-products-all-mutants.log`. The final table mutant log is `/tmp/stage3-products-table-mutant-final.log`; it failed the intended assertion in 12.758 s. The first hash-mutant attempt overlapped a recipe mutation, accidentally requested a different cache key, and was stopped with all descendants. It is discarded. The corrected byte mutant uses the already verified product directory directly and restores the reference bytes in a finally block.

## Baseline and integration validation

From the repository root, the full lane was launched with `ADAMIC_BUILD_CACHE_DIR=/tmp/adamic-gate/stage3-products-final timeout 600 bash stage3/lane/run.sh /tmp/adamic-gate/stage3-products-baseline-lane > /tmp/stage3-products-baseline-lane.log 2>&1`. The first eight-worker oracle attempt left a worker without counts; the cgroup reported one OOM kill. Its evidence is retained in `oracle-eight-worker-failed`. The full unchanged tree was retried with eight workers and `NODE_OPTIONS=--max-old-space-size=1024`, using `timeout 600 bash stage3/oracle/run.sh /tmp/adamic-gate/stage3-products-baseline-lane/adapted-tree /tmp/adamic-gate/stage3-products-baseline-lane/oracle --workers 8 --limit-seconds 540 > /tmp/stage3-products-baseline-retry.log 2>&1`. There was no additional OOM kill. Install took 3.337 s, build 2.191 s, full tests 413.058 s; complete retry 418.661 s.

Observed counts: **106366 passing, 1 failing, 0 pending**, exactly the stock-tsc lane expectations. The only baseline difference is the sanctioned `api/typescript.d.ts` widening. Oracle exit 1 is expected for that exception. The first check lacked the parser at the legacy cache path; the final implementation exposes the prepared parser. The root check with the final immutable parser was `STAGE3_CACHE=$(cat /tmp/stage3-products-final-path) timeout 90 python3 stage3/lane/check.py /tmp/adamic-gate/stage3-products-baseline-lane > /tmp/stage3-products-baseline-check-final.log 2>&1`: exit 0, `PASS stage3 landing lane [causes: timeout=0, baseline-content=1, baseline-missing=0, exception=0, other=0]`.

The full oracle tree came from commit 8dd02d807a618a442817287a4a7903cfc0f43db7. Commit 2888e545400f3051a5779557f246f5def87e78f9 adds parser exposure and process-group deadlines; its final product has the same independently verified tree hash. The full oracle was not repeated after these consumer/deadline fixes.

Root integration lane checks passed after the report commit with the prescribed `git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -`, bounded by an outer `timeout 90`, with output in `/tmp/stage3-products-lane-checks-final.log`. Final output: `lane checks 1.0 s: gofmt and tools on 3 Go files, t.Parallel on 1 test packages; vet 1 packages`, exit 0.

No new .a fixtures or counts rows were added. No whole Go package confirmation or full gate is run. No PR is opened. Remote store publication is still the existing developer-tools hook; this unit pushes source only.

Transient public GitHub authentication failures were retried. An earlier product-one proof used the measured pin-verified mirror during that failure; the final cold product-one measurement above fetched a fresh mirror. Superseded scratch trees exhausted the 8.8 GB disk once; only this task's named obsolete directories were removed, and the final cold measurements were repeated. An npm config experiment failed by double-loading /dev/null; the final recipe uses two different empty config files.

The requested first-green-commit-within-30-minutes target was missed during cold measurements and the oracle OOM retry. All commits on this branch carry the final `Task: #hefhych` trailer.
