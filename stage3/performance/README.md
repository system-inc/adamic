# Checker performance groundwork

This baseline compares stock TypeScript 6.0.3 on Node 24.19.0 with
`@typescript/native-preview` 7.0.0-dev.20260707.2, before Adamic has a native
TypeScript checker. It measures fresh-process `--noEmit` CLI wall time, including
startup and library loading. It does not establish any Adamic performance claim.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/performance-setup.log 2>&1
source /workspace/adamic-tools/env.sh # use the path setup prints
stage3/performance/run.sh /tmp/new-performance-results > /tmp/performance.log 2>&1
stage3/performance/run.sh /tmp/new-native-results --native /absolute/native-tsc > /tmp/native-performance.log 2>&1
stage3/performance/prove-mutant.sh /tmp/new-mutant-results /absolute/typescript/lib/tsc.js > /tmp/mutant.log 2>&1
```

The output directory must not exist. Prerequisites: Linux x64, Python 3.12 or
newer, Node at the pinned version, npm, git, GNU timeout, network access,
and a C compiler plus make if GNU time is absent. npm ci verifies package
integrities from the committed lockfile. Missing `/usr/bin/time` is handled by
building checksum-pinned GNU time 1.9 in the output directory. This worker could
not install system packages (permission denied under `/var/lib/apt`). Its exact
GNU time executable, version and hash are recorded in `time-tool.json`.

`pins.json` records input commits. `versions.json` records executable/compiler
SHA256 hashes and observed versions. `input-hashes.json` records source, generated
source, configuration and stock library bytes. Downloads and input preparation
are outside the timed intervals. Reruns fetch the same commits, regenerate
TypeScript's diagnostic map with its upstream script, and install the same
locked declaration dependencies. Libraries are supplied as explicit roots with
`noLib`, so both compilers read identical TypeScript 6.0.3 declarations. Neither
compiler uses its own differing implicit library set.

The driver cases reuse its existing header materializer. Three clean cases
exercise variance and generic inference; the const-enum error case is retained
as an exclusion and as the mutant witness. mitt checks its library source;
Zod checks its library source, excluding tests, benchmarks and compile.ts.
These are deliberately source-only projects with explicit strict ES2020,
NodeNext settings, not their upstream development/test configurations. The
TypeScript compiler workload runs `-p src/compiler --noEmit --pretty false`
with upstream settings, explicit shared ES2020 declarations, and composite and
incremental disabled to prevent persistent build state. Its original config is
saved next to the generated one. No upstream semantic source is edited.

Preflight compares stdout and stderr bytes and process exit status. It does
not normalize paths, messages, ordering, whitespace or status. Any discrepancy
excludes an input; a native discrepancy aborts the entire run before any compiler
is timed. Subsequent warmup, timed and memory runs must also reproduce preflight
output. Both Go modes are checked independently: `--singleThreaded` and no
concurrency flag for the default. GOMAXPROCS is inherited and recorded.

Each accepted input has two warmups per mode and ten measured fresh processes
per mode. Mode order is deterministically shuffled in each round (seed 0).
Wall time uses perf_counter and a blocking process wait; GNU timeout limits
individual invocations to 300 seconds without Python timeout polling latency.
SD is the sample standard deviation. Peak RSS is a separate GNU time `-v` run,
not the compiler's reported heap size. This is a shared cloud box with warm file
caches, no CPU isolation and no cold-cache claims. CPU model, nproc, quota, load
and full command argv are retained. Differences inside run variability should
not be interpreted as stable speedups.

The largest selected source workload gets `node --cpu-prof` and separate
`--extendedDiagnostics` invocations for Node and both Go modes. The raw V8
profile and top sampled self times are retained. Compiler phase times come from
extended diagnostics rather than guessed phase attribution from function names.
Emit is disabled; diagnostics may omit a phase instead of reporting zero.

The diagnostic mutant forwards to the stock Node oracle and replaces only the
first TS2567 message's `Enum` with `MUTATED`, preserving stderr and exit status.
`prove-mutant.sh` requires exit 1, exactly that changed output, unchanged stderr
and exit, and no timing commands or results. This tests rejection, not native
checker performance. No Adamic oracle fixture was added, so counts.md needs no
refresh. No whole-package tests or full gate are run for this measurement unit.

## Observed baseline

Base: ef3141e9b1152ab51b51497f8ce3a2799449c8a3. AMD EPYC 9V74
80-Core Processor; nproc 5; cgroup quota 400000/100000 (four CPUs).
Setup completed in 184.098s. Exact timing lines and the system package-manager
failure are retained under evidence/. Six clean inputs match all output bytes
and status. `061_constEnumErrors` has identical diagnostic bytes but Node exits
2 and both Go modes exit 1; it is excluded. mitt has one source file (3,723
bytes), Zod 123 (1,251,581 bytes); the compiler has 194,779 TypeScript lines in
extended diagnostics, plus its declarations, and is the largest workload.

Mean ± sample SD, seconds; ten runs per cell:

| Input | node | go-single | go-default |
|---|---:|---:|---:|
| 001_varianceCantBeStrictWhileStructureIsnt | 0.9606 ± 0.0800 | 0.2664 ± 0.0132 | 0.2564 ± 0.0130 |
| 024_genericTypeParameterEquivalence2 | 0.9341 ± 0.0336 | 0.2655 ± 0.0114 | 0.2562 ± 0.0168 |
| 056_genericCallInferenceInConditionalTypes1 | 0.9553 ± 0.0432 | 0.2742 ± 0.0197 | 0.2588 ± 0.0132 |
| mitt | 0.3672 ± 0.0107 | 0.0519 ± 0.0030 | 0.0529 ± 0.0031 |
| zod | 4.0239 ± 0.2145 | 1.1477 ± 0.0455 | 0.9325 ± 0.0530 |
| typescript-compiler | 8.3587 ± 0.4202 | 2.6831 ± 0.1292 | 1.7810 ± 0.0840 |

[Full table with min, max and peak RSS](evidence/results/table.md).
[Individual samples](evidence/results/results.json),
[versions and hashes](evidence/results/versions.json), and
[preflight exclusions](evidence/results/preflight.json) retain the observations.
With --native the comparison table adds a native column; the full table and
JSON add its measurements only after global preflight succeeds.

Largest-input phase diagnostics, seconds (separate diagnostic runs):

| Compiler | Parse | Bind | Check | Emit label | Total |
|---|---:|---:|---:|---:|---:|
| Node | 0.770 | 0.550 | 6.400 | 0.000 | 7.800 |
| Go single | 0.229 | 0.082 | 2.046 | 0.206 | 2.790 |
| Go default | 0.165 | 0.059 | 1.306 | 0.176 | 1.802 |

Checking dominates Node's measured phases (about 82% of diagnostics total).
The V8 profile sampled 8.750s: GC self time is 0.434s (5%); prominent checker
self samples include checkIdentifier (0.163s), isTypeRelatedTo (0.135s),
recursiveTypeRelatedTo (0.132s), and getFlowTypeOfReference (0.125s).
These are self samples, not inclusive phase totals. The exact
[profile summary](evidence/results/profile-summary.json) and gzip-compressed
[raw V8 profile](evidence/results/tsc.cpuprofile.gz) are retained; decompress
with gzip -dk to open it in a profiler. All extended diagnostics are in
`evidence/results/raw/typescript-compiler/*-extended.stdout`.
Go reports an Emit label despite --noEmit; no output files were produced.
This project retains isolatedDeclarations and emitDeclarationOnly settings,
so the label must not be read as a measurement of emitted-code performance.

Final commands run: run.sh /tmp/performance-results (exit 0),
prove-mutant.sh /tmp/performance-mutant-final with the pinned stock tsc.js
(exit 0), bash -n for both shell entrypoints, Python AST parsing, and focused
artifact verification. The mutant run prints REFUSED on 061_constEnumErrors,
then PASS: one diagnostic changed; stderr and exit preserved; global preflight
refused all timing. Its wrapper, exact bytes and report are committed.
The initial incomplete runs were superseded after missing GNU time, input
preparation issues and Python timeout polling were fixed; their timings are
not the retained baseline.

Not covered: an actual Adamic native checker, its performance, emit/watch/build
workloads, the whole 301-project driver, full upstream development/test
configurations, other operating systems or architectures, and cold caches.
The positive native slot was checked in preflight through a Node forwarder;
it was deliberately mutated on the error case and never timed.
