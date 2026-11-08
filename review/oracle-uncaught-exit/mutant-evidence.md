# Uncaught throws stop the oracle

Task 995spev. Commit parent: origin/main ce883cb24bc849579918fb39a3cbef8439f23246.
Initial proof runs used 93fa14ff0f189f13be19aec4f82974d4f663e4e8.

The entire oracle/adamic.mjs was read before editing. The only runtime change is
process.exit(70) after the uncaught handler's stderr.write returns. Its exitSignal
return, panicking guard and RangeError frame naming are unchanged.

Native timer_dispatch calls adamic_uncaught when its callback leaves adamic_thrown.
adamic_uncaught calls adamic_panic, which writes the panic line and calls _exit(70).
Performance head 801bc791336eb5f13fcb2b0dfac9ed66aae0c35a explicitly records 70
for timers_source/callback_throw.a. The observed fixed-runner result is also 70.
No exit-status contract was changed. docs/0.1.md contract item 5 says uncaught throws
exit 1; that differs from these implementation and fixture contracts and was left alone.

mutant-observations.json records complete stdout, stderr and status for every
source proof, raw Node stack, timer mode and mutant run on Node v24.19.0.
Before the fix, all three new fixtures print pending timer after the panic.
After it, all three have empty stdout, exit 70, and exactly this stderr line:

```
adamic: panic: Error: uncaught exit
```

Plain Node exits 1 with empty stdout and the Error stack recorded in the JSON.
The promise-job throw reaches uncaughtException with Error: uncaught exit and
origin unhandledRejection, observed through uncaughtExceptionMonitor. There is
no separate unhandledRejection handler in oracle/. It has the same old bug and
is fixed by the same one-line change.

TestOracleUncaughtExitStopsPendingTimers/two_timers is the regression assertion.
TestOracleUncaughtExitCodeMutantKeepsTimerAlive copies the real runtime bundle,
restores process.exitCode = 70 only in the uncaught handler, and applies that
same assertion. The mutant exits 70 with the correct panic line but prints
pending timer, so the stdout assertion rejects it. Compilation failures are
not accepted as a mutant kill. Every new Go test and subtest calls t.Parallel().
Tests launch only Node through the oracle's bounded execute helper.

The compiled regression test was also run against a private repository containing
the mutant runtime: TestOracleUncaughtExitStopsPendingTimers/two_timers failed
with test-process exit 1 because pending timer printed. The identical binary
and fixture passed with exit 0 after replacing only that runtime with the fixed
one. Full outputs: mutant-red.log.gz and mutant-control.log.gz.
The focused regression and built-in mutant run is in mutant-focused.log.gz.

The timer census covers the native timers/oracle.a modes, async_host/oracle.a
modes and watchfile/oracle.a modes, plus all six timers_source fixtures extracted
from Performance's exact head. CommonJS-only native probes use a createRequire
preload while still running through the fixed oracle/node.mjs. Every result is
listed in the JSON. New uncaught_exit fixtures are recorded there separately.

Observed fixed-runner timer exits, with complete output in mutant-observations.json:

| Fixture and mode | Exit |
| --- | ---: |
| `801bc791:timers_source/forwarded.a` | 0 |
| `801bc791:timers_source/fewer_parameters.a` | 0 |
| `801bc791:timers_source/clear_then_set.a` | 0 |
| `801bc791:timers_source/order.a` | 0 |
| `801bc791:timers_source/callback_throw.a` | 70 |
| `801bc791:timers_source/clear_kinds.a` | 0 |
| `internal/native/testdata/timers/oracle.a throw` | 70 |
| `internal/native/testdata/timers/oracle.a order` | 0 |
| `internal/native/testdata/timers/oracle.a interval` | 0 |
| `internal/native/testdata/timers/oracle.a cancel` | 0 |
| `internal/native/testdata/timers/oracle.a cancel-self` | 0 |
| `internal/native/testdata/timers/oracle.a foreign` | 0 |
| `internal/native/testdata/timers/oracle.a unref` | 0 |
| `internal/native/testdata/timers/oracle.a unref-live` | 0 |
| `internal/native/testdata/timers/oracle.a deadlines` | 0 |
| `internal/native/testdata/timers/oracle.a graph` | 0 |
| `internal/native/testdata/timers/oracle.a arity` | 0 |
| `internal/native/testdata/timers/oracle.a host` | 0 |
| `internal/native/testdata/timers/oracle.a turn-phase` | 0 |
| `internal/native/testdata/async_host/oracle.a early` | 0 |
| `internal/native/testdata/async_host/oracle.a late` | 0 |
| `internal/native/testdata/async_host/oracle.a reject` | 0 |
| `internal/native/testdata/watchfile/oracle.a throw` | 70 |
| `internal/native/testdata/watchfile/oracle.a interval` | 0 |
| `internal/native/testdata/watchfile/oracle.a listeners` | 0 |
| `internal/oracle/testdata/uncaught_exit/two_timers.mjs` | 70 |
| `internal/oracle/testdata/uncaught_exit/top_level.mjs` | 70 |
| `internal/oracle/testdata/uncaught_exit/promise_job.mjs` | 70 |

The all-extension testdata census also matches the Program census CSV, the lint
message catalog JSON, and RegExp extraction scripts. These are ledgers, message
text, or corpus extractors that expose or stub timer bindings, not timer runtime
fixtures. No additional timer program was found. Benchmark worker harnesses and
Node declaration files are likewise outside this fixture census.

The fixture files live under internal/oracle/testdata and are consumed directly
by the new tests, not registered as compiler fixtures. They need no counted
native records. compiler-dependencies.json describes linked compiler packages
and launchers; these tests add neither. No census list or declared compiler
input needs an entry for these files.

Logs are in /tmp/oracle-uncaught-evidence, with setup in
/tmp/oracle-uncaught-setup.log and fetch in /tmp/oracle-uncaught-fetch.log.
Broad validation was stopped at the user's revised instruction before commit.
Completed invocation exits are in mutant-validation-before-stop.json; failure logs
are in mutant-broad-run-failures.tar.gz. No interrupted invocation is called passing.

Observed whole-package results before stopping:

| Package | Exit and observation |
| --- | --- |
| internal/oracle | 1: 30-minute timeout; missing WASI_SYSROOT and the stage3/api TypeScript module |
| internal/native | 1: 30-minute timeout; missing WASI_SYSROOT in RegExp WASM tests; TestDebugSourcePathsSplit hit its 20-second deadline |
| internal/worker, initial environment | 1: missing WASI SDK/products |
| internal/worker, SDK configured | 1: three WASM mutant child checks hit 20-second deadlines |
| bench | 0 |
| internal/adamiclint/repowalk | 0 |
| internal/fuzz | 0 |

Other broad checks and queued isolated retries were stopped without claiming completion.
The pinned SDK was installed with cloud/setup.sh --wasi-sdk (exit 0), and
npm ci --ignore-scripts --no-audit --no-fund --prefix stage3/api passed, including
the exact TypeScript 6.0.3 assignment-proof typecheck helper. Source and unit-cache
inputs were also configured. These were environment changes, not repository code edits.

The user will run the whole oracle through Loom. After this commit is pushed,
only timer/uncaught fixture tests, the new mutant proof, and the stage3 runner
guards will run. Their fresh exit codes will be reported in the hand-back.

