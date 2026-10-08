Partial result: the catalog and WASI phases are split and their unions are preserved, but the 30-second unit budget is not met in this measurement. Both local phase walls increased.

The measured compiler tree is `89cbe74a5009047b0fb2fd31c323794d00dbcc12`, the tools branch base. The historical main `73352e87` measurements (catalog 296.949 s, WASI 205 s) were not replayed. This is a phase-only comparison; the whole gate was not run.

Instrument: Python `time.monotonic`, including process startup, compilation, checkout, test execution and cleanup. Host `95dd530494ea`, affinity CPUs 0–3, cgroup `cpu.max=400000 100000` (four CPUs). Go 1.27.1, native clang 20.1.8, WASI SDK 27 / clang 20.1.8-wasi-sdk, Node 24.19.0.

Each phase started with an empty shared Go build cache. Compiler artifacts were reused within that phase, while test results stayed uncached (`-count=1`, `ADAMIC_GATE_UNCACHED=1`). Later units were not measured independently with empty compiler caches; OS page cache was not reset. This does not certify the requested independent cold budget for every unit.

The original default four-worker cold catalog filled the 8.8 GiB writable disk and was stopped without a verdict. The valid legacy catalog comparison explicitly uses `--jobs 2`. The split scheduler uses the runner default of two workers on this four-CPU box, one Go worker per unit. Local phase wall includes bounded queueing; slowest-unit wall is reported separately.

| Phase | Before wall (s) | After wall (s) | Slowest new unit (s) | Units | Over 30 s |
| --- | ---: | ---: | ---: | ---: | ---: |
| catalog | 707.912 | 892.943 | 757.906 | 16 | 8 |
| wasi | 240.724 | 797.897 | 551.413 | 36 | 2 |

Count and name comparisons in [summary.json](summary.json) match the legacy phase exactly. All 16 catalog units ran: 8 caught, 3 stale, 5 skipped in both versions; no uncaught entries. All 36 WASI units passed their selected fixture, including `requests`, and no extra fixture ran in any unit. Budget overruns make the split phases red even when their semantic checks pass. Per-unit commands, names, scratch paths, exit codes and walls are in [catalog fast.json](after-catalog/fast.json) and [WASI fast.json](after-wasi/fast.json). Production whole-gate runs keep `full.json` and publish the same ledger in `fast.json`, marked `gate_kind: full`.

Over-budget units in this run:

| Phase | Unit | Wall (s) |
| --- | --- | ---: |
| catalog | `01 shared-slice-append` | 750.966 |
| catalog | `02 liveness-throw` | 757.906 |
| catalog | `03 defined-lent` | 38.292 |
| catalog | `04 borrowed-array-move` | 35.497 |
| catalog | `05 spread-method-reuse` | 37.600 |
| catalog | `06 constructor-capture-region` | 35.819 |
| catalog | `07 borrowed-element-reads` | 30.865 |
| catalog | `09 literal-undefined-field` | 42.758 |
| wasi | `internal/load/testdata/0.1/compile/01_hello.ts` | 551.413 |
| wasi | `internal/load/testdata/0.1/compile/02_fizzbuzz.ts` | 550.206 |

Real failure probes used the production checker and native package with warm compiler artifacts, without editing compiler tests: Go test control was forced to exit 2 only for catalog entry 01; Node was forced to exit 17 only for the hello fixture. The affected units were named `catalog/01 shared-slice-append` and `wasi/internal/load/testdata/0.1/compile/01_hello.ts`; paired entry 02 / fizzbuzz remained semantically green. See [catalog probe](planted-catalog/fast.json) and [WASI probe](planted-wasi/fast.json).

`run_test.py`: 51 tests green. All 30 behavior-changing mutants were actually run and caught; [mutants.py](mutants.py), [mutants.json](mutants.json), and [run_test.log](run_test.log) retain commands and results. Checks cover exact fixture selection, required passes, inventory refusal, catalog verdicts, separate scratch, CPU quota/concurrency, unit coverage and count, named failures, aggregate exit, the 30-second boundary and full/fast JSON publication.

The measurement and failure scripts used on this box are retained as [measure-used.py](measure-used.py) and [plant-used.py](plant-used.py), including the explicit jobs override. For replay, prepare `/tmp/split-before-run.py` with `git show 89cbe74a5009047b0fb2fd31c323794d00dbcc12:cloud/fast-gate/run.py`, source `/workspace/adamic-tools/env.sh`, and run each `before/after catalog/wasi` phase sequentially. The catalog plant cache was a hard-link snapshot of the split catalog build cache, taken after entry 01 completed; the WASI probe reused the split WASI cache before its deletion. Plant scripts live under `/tmp/split-plant.py` for the retained measurement driver. The final precision-only check and full/fast alias edits were verified by the final test/mutant run; they do not change measured commands or concurrency.
