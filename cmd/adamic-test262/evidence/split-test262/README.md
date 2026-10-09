# Test262 independent units on four CPUs

Base: origin/main `54cbc125422d` (full SHA in coverage.json). Branch: devtools/split-test262.
Instance quota: cpu.max `400000 100000`; nproc reports 5, but the quota is four CPUs.
GOMAXPROCS=4. Go 1.27.1; clang 20.1.8; Node v24.19.0. Measurements ran on one instance.

Instrument: Python `time.perf_counter()` around a fresh test executable process, plus Go's
`-test.v -test.count=1` terminal elapsed time. Setup is included in wall time. Setup below is
Go parent elapsed minus its selected child elapsed; cleanup/process overhead remains in wall.
The compiler build is inside the measured parent, never TestMain. No build time is subtracted.

Initial measurements of the unmodified tests, before editing, including the first compiler/runtime
materialization: large output 18.357 s wall / 18.32 s Go; parallel cache 4.252 s wall / 4.22 s Go.
See baseline.txt. Both were already below 30 s in isolation here, unlike the supplied whole-gate run.

Final measurements used SHA-256-addressed test executables copied from a local product store and
verified before running (products.json). Go's content-addressed build cache and the keyed native
runtime library were available. This was a local product fetch, not a remote/fleet-store fetch.
A unique ADAMIC_SPLIT_TEST262_MEASUREMENT value per process changes the observation-cache context,
so the original test also starts with fresh observations. New children additionally own private
observation caches. The intentional warm child's priming is included in its time. No other test
workload ran concurrently with final-before.txt or final-after.txt.

| TestLargeCompilerOutputIsComplete | Go unit s | Setup s | Process wall s |
| --- | ---: | ---: | ---: |
| Before (unsplit) | 9.22 | included | 9.259 |
| After /large.js | 5.93 | 3.21 | 9.180 |

| TestParallelCachedMatchesSerial | Go unit s | Setup s | Process wall s |
| --- | ---: | ---: | ---: |
| Before (unsplit) | 4.62 | included | 4.654 |
| After /cold | 0.33 | 4.14 | 4.516 |
| After /warm | 0.37 | 4.02 | 4.440 |
| After /limit | 0.25 | 4.09 | 4.384 |

No measured unit or parent setup exceeded 30 s. large.js remains one indivisible input: dividing
its 1,800 assertions would weaken the requirement that one emitted program exceeds 256 KiB.
The gate now independently enumerates/selects this case and the three cache-comparison cases.
Gate timing weights use the final child measurements, final large parent, and the observed full
parallel parent (4.70 s in after.txt); these are estimates for planning, not performance assertions.

## Shared build and independence

`buildTest262Compiler(directory string) error` in split_setup_test.go builds the compiler once per
process via sync.OnceValues. It takes no test state and already has buildcache.Product's callback
signature. Only its caller needs to move behind Product(t, Inputs{...}, buildTest262Compiler).
TestMain only cleans up the product after m.Run; it does no setup. There is no separate runner
build inside the tests: compiler workers use the fetched test executable. RuntimeLibrary already
builds once per content/toolchain/flags key and takes no test state.

Each child has fresh work and observations. The serial subprocess reference is computed once in
the parent. cold compares the first cached parallel report, JSON, table and progress against it;
warm primes its own cache, then compares the second report in the same way. limit compares serial
and parallel limit=1 reports with its own cache. Selecting any child alone needs no sibling's work.

## Coverage and planted failures

coverage.json verifies original and current fixture bytes by SHA-256 and names/counts:

- large.js: one unchanged input, 1,800 `assert.sameValue(1, 1)` assertions; pass count=1 and the
  emitted C >256 KiB with final newline checks remain.
- Mini corpus: five unchanged cases: crash/stderr.js, fail/uncaught.js, pass/pad.js,
  refuse/var.js, skip/negative.js. Both cold/warm comparisons cover all five; limit=1 retains
  the same five-file report and one attempted selection (crash/stderr.js). The parent pins
  the exact five names and still requires every verdict class.
- Original comparison cases cold round=0, warm round=1, and limit equality map one-to-one to
  cold, warm, and limit children. warm adds its independent priming execution; the union of
  inputs and assertions is unchanged. No fixture bytes changed.

Temporary production mutants were compiled into separate executables, with run.go restored
immediately after each build. Every run selected only the named unit and exited 1:

- Restored the old compiler capture limit (16 MiB -> 256 KiB): /large.js failed with
  `command output exceeded capture limit` (mutant-large.txt).
- Incremented report.Pass when jobs>1: /cold and /warm failed JSON/table/progress comparison;
  /limit failed `parallel limit selected different tests` (mutant-parallel.txt).

Clean full cmd/adamic-test262 package PASS (36.212 s total across all tests); cmd/adamic-gate
package PASS, including exact child enumeration and existing selector/coverage tests. That package
aggregate is not a test unit. git diff --check passed. No mutants are in the committed source.

## Repeating the measurement

Source /workspace/adamic-tools/env.sh; build with GOMAXPROCS=4:
`go test -c -o /workspace/scratch/split-test262/after.test ./cmd/adamic-test262`.
Run measure.py from cmd/adamic-test262, supplying executable, label and anchored Go selectors,
e.g. `^TestParallelCachedMatchesSerial$/^warm$`. The included harness writes logs under
/workspace/scratch/split-test262. Baseline binary was built before any source edits at the base.
The final scripts stage each executable under its SHA-256, copy by that address, and verify the
copy; hashes are in products.json. Submodules were pinned to the base gitlinks.
