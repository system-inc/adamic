# Entry progress and cached IR field answers, task 20pb346

Measured base: `31d7a244808fa8092643926d21ab9d310a54cce3` (origin/main at
preparation). `environment.json` records the checker submodule pins, Go and Node
versions, adapted product key, TypeScript source pin, and binary hashes.

The entry pass now sets `LATENT_TIMES` to its own
`main/tsc/latent-times.jsonl`. Both entry and compiler timing files are copied to
the area alias and compressed after reporting. A checker-blocked first-error
entry produces an empty timing file, since it attempts no units. The existing
unit timer encodes directly to an `os.File`, without a user-space buffer.
`TestLatentUnitTimesVisibleBeforePassReturns20pb346` reads the first completed
unit from the still-open file while the test's simulated pass is blocked before
its next unit. No filesystem durability claim is made; visibility to readers is
the progress contract.

`latent.jsonl` remains buffered. Its deferred generic `instantiated` fields can
be changed by callers in later files. The existing order is preserved exactly;
no traversal or output sorting was changed. The meter README records this
finding and directs live readers to the matching timing file.

The generator replaces per-field reflection/name/type tests with immutable plans
indexed by `reflect.Type`, initialized once. Each full struct copy looks up its
plan once, then indexes the plan for its fields. Types without skipped fields
use the zero plan. `argumentFacts` still requires nil and retains the populated
cache panic text. `methodTables` and `indexes` are still left empty, and generated
type pointer sharing is unchanged. The tested helper source is included in the
scratch lower overlay so per-ref compiler checkouts do not need the new helper
package installed in their source tree. Concurrent reads require no locks;
focused tests also pass under the race detector.

## Completed evidence

Setup ran with `GOPROXY='https://proxy.golang.org|direct'` and
`timeout --kill-after=60s 1800 bash cloud/setup.sh`, finishing successfully in
521.900 seconds. Recursive submodules were initialized, and `npm ci` installed
the stage3/api lockfile packages. Main's meter Python suite passes all 65 tests
with real census binaries and no skips; the changed suite passes all 66 with
real binaries and no skips. The complete latent and meter Go suites pass, the
pair tool passes all 10 tests, and focused plan/timing/generator race tests pass.
Adjacent `*-tests.txt` files contain the captured outputs. The earlier missing
package/tool setup failures are not claimed as pre-existing test failures.

All new Go tests call `t.Parallel()`, have unique names, and start no Go or
compiler subprocesses. The updated existing Python generator probe compiles its
fixture, as before; its generated Go test functions also call `t.Parallel()`.
`cloud/fast-gate/compiler-dependencies.json` was reviewed. No package's compiler
use changes, so no manifest change is required.

Both entry runs had a 600-second wall budget; both capped compiler runs had a
120-second wall budget. The baseline entry was given the existing timing flag
externally so both sides have the same observation mechanism. The run JSON files
record actual wall times and expected SIGTERM exits at their caps. The entry
growth JSONL files contain samples at every 60 seconds through 600 seconds.
Canonical latent output remains at 6,139 bytes throughout both entry windows.
The entry timing streams share the first 1,528 unit identities (before: 1,528,
after: 1,568); both capped compiler timing streams contain the same 98 unit
identities. The capped canonical outputs are byte-identical. Because the actual
per-unit latent records remain buffered until return, these capped runs cannot
prove those unreturned records equal by bytes. `identity-proof.json` distinguishes
these timing-prefix comparisons from completed-output comparisons.

| Seconds | Before timing bytes | After timing bytes |
| ---: | ---: | ---: |
| 60 | 33,453 | 37,226 |
| 120 | 47,799 | 52,094 |
| 180 | 53,494 | 234,970 |
| 240 | 234,603 | 249,696 |
| 300 | 246,908 | 290,661 |
| 360 | 278,742 | 298,011 |
| 420 | 291,959 | 310,650 |
| 480 | 298,470 | 317,648 |
| 540 | 310,261 | 323,517 |
| 600 | 316,807 | 326,237 |

The two published entry CPU profiles were repeated in a quiet window with no
builds, tests, or other census passes overlapping. Both use `GOMEMLIMIT=8GiB`
and a 180-second wall budget. They start immediately before `lower.Latent` and
stop after 120 seconds while the census continues. Their sample totals differ and
they cover different completed-unit prefixes; these shares are not a per-call
microbenchmark. Profile instrumentation exists only in the scratch drivers.

| Profile | Duration | Sampled CPU | Private helper self | Private helper cumulative |
| --- | ---: | ---: | ---: | ---: |
| Before | 120.00 s | 163.15 s | 0.78% | 10.83% |
| After | 120.00 s | 166.65 s | 0.14% | 0.48% |

See `profile-shares.json`, both raw CPU profiles, and their text summaries.
The initial profiles from the 600-second growth runs are retained separately as
`*-entry-initial-cpu.pprof` and `initial-profile-shares.json`; the initial
baseline window overlapped some build/test work and is not the published comparison.

`completed-pass-proof.py` ran actual before/after entry and compiler binaries
on a small cross-file generic fixture. Both completed streams are byte-identical,
including a true late `called` witness and a false `unused` witness. Compiler:
4,371 bytes, 6 units. Entry: 5,928 bytes, 8 units. See
`completed-fixture-proof.json` and the compressed streams. An initial fixture
lacked an import route from the compiler's first file to its caller, causing the
same source-route panic on both binaries. Adding that route made the witness
assertion meaningful.

The completed-fixture reports are also byte-identical: `report.json` is 13,783
bytes and `report.md` is 4,077 bytes. Independent ordinary-census elapsed seconds
and compiler unit timing values naturally differ, so report comparison holds
those unchanged inputs fixed after checking ordinary census records excluding
`seconds`. Independent entry timing sidecars remain different and are ignored.
The new script integration test additionally verifies that a deliberately invalid
entry timing file neither changes nor breaks the JSON or Markdown reports.
`report.py` itself is unchanged.

## Mutants

Each patch and captured failure has `mutant` in its filename. Run
`python3 stage3/meter/evidence/unit-visibility-20pb346/check-mutants.py` with the
provisioned tools on PATH. The runner restores each source file in a finally
block and rejects a surviving mutant.

| Mutant | Named test turned red |
| --- | --- |
| Entry timing flag removed | `test_entry_times_sidecar_preserves_report_20pb346` |
| An extra field skipped | `TestLatentPrivateIRPlanPreservesUnlistedFields20pb346` |
| Populated privateNil panic removed | `TestLatentPrivateIRPlanRejectsPopulatedNilCache20pb346` |

All three were killed. See `mutant-results.json`, the three patches, and their
failure outputs.

## Full compiler identity

The default-memory baseline was killed by SIGKILL after 16,164.303 seconds
(4 hours 29 minutes), with 2,334 completed timed units. The container has a
16 GiB memory cap, reached that peak, and recorded an OOM kill. This is
consistent with an OOM kill; direct kernel attribution to that PID was not
available. The latent output still contained only its 165-byte metadata line.
The subsequently started after pass was stopped so a controlled pair could
use matching memory settings.

A retry used `GOMEMLIMIT=8GiB` and a six-hour per-side budget. The baseline
continued making progress beyond two hours, but was deliberately stopped
without completing. The queued after retry was not started. The failed default
attempt, interrupted after attempt, and incomplete controlled retry are saved.
No completed full-corpus byte identity or matched full-pass wall times are
claimed. The user subsequently replaced this requirement with bounded disjoint-piece
comparisons; those results are recorded in `pieces/`. No more whole passes
will be run. The completed fixture above is
a separate, smaller proof and does not replace it.

After rebasing, the complete Go suites, all 66 meter Python tests, all 10 pair
tests, and the focused race tests pass again. See `final-*-tests.txt`.

Final integration base: `d4c6b07aca1a7c55785226563c2ad79d552c1adf`.
Measurements remain pinned to the preparation base above.

Task: #20pb346
