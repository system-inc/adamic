# checker.ts lowering scout: 75-minute rerun

First stop: **src/compiler/checker.ts:1708:51, Refused**.
Exact reason: `an unproven predicate argument for parameter nodeTest (argument "isExportSpecifier")`.
Full-mode inventory: pending; the first-error entry pass is complete.
Meter comparison: pending the companion first-error compiler-directory pass.
Branch: `scout/checker-lowering`; compiler and adapted inputs remain pinned to `a0ba9dd52b23326880b16a7898d446597cfb5a88`.

<!-- current-rerun-start -->
## Measured first-error result

The tsc-entry first-error census completed successfully, exit 0, under the shared
75-minute background deadline. All **81** resolved implementation-file records
passed the meter reporter's coverage and measurement-label checks. The entry
header has zero checker diagnostics. The first checker.ts observation in its
record's emitted order is a **refusal_scan** finding, in the unit beginning at
checker.ts:1486:1:

```text
src/compiler/checker.ts:1708:51
Refused
an unproven predicate argument for parameter nodeTest (argument "isExportSpecifier")
            const node = getParseTreeNode(nodeIn, isExportSpecifier);
```

The exact diagnostic text, with only the scratch directory prefix omitted:

```text
src/compiler/checker.ts:1708:51: Adamic 0.1 refuses an unproven predicate argument for parameter nodeTest (argument "isExportSpecifier"); pass a named function or arrow whose body proves both predicate branches; return a boolean and narrow at the caller (adamic/no-type-predicate)
```

The meter's owner rule assigns this prefix to **#1w5bve2 (Optimizer)**.
This is checker.ts's first census observation, not the entry program's first
production stop (which remains the separately observed core.ts generator).

Deduplicating `(kind, where, reason, text)` exactly as the meter does, the entry
scope has **207 NotYet**, **4,240 Refused**, **2,279 recovered panic sites**,
0 errors and 0 SkippedDependency sites. checker.ts has **20 NotYet + 1,538
Refused = 1,558 blocker sites**, across **208 distinct reasons**, in this
**first-error** pass. These are not full-mode counts. Recovered panics remain
outside the blocker ranking; no successful IR or backend output is claimed.

`scout/checker-lowering.jsonl.gz` now preserves the actual checker.ts findings
from the completed first-error entry stream, deduplicated by the meter key, with
`census` scope and source lines added. Original finding fields are retained.
The companion directory census and full entry census are still running.

## Background run and progress checks

The shared hard limit started **2026-11-03T21:02:27Z** and expires at
**2026-11-03T22:17:27Z**. `nohup timeout --signal=KILL 4500s bash
/tmp/checker-rerun.sh` bounds all census children. The supervisor records its
exit in `/tmp/checker-rerun/runner.status`. Separate logs, JSONL streams and
`LATENT_TIMES` ledgers live in `/tmp/checker-rerun/` under `first-entry`,
`first-directory` and `full-entry` prefixes. Both latent modes retain
`LATENT_ASSERT_NO_OUTPUT=1`; production sources remain untouched.

An independent monitor records log tails, per-file record counts and timed-unit
counts every 300 seconds in `/tmp/checker-rerun/progress.log`. The main census
buffers file records until all callers have run (`Latent`, in
stage3/census/latent/lower.go.txt:213-238), to resolve deferred-generic witness
metadata. Consequently an empty main log does not mean no work happened.
Timing rows showed continuing progress through types.ts, scanner.ts,
utilities.ts, factory files, parser.ts, checker.ts, transformers, program.ts,
resolutionCache.ts and executeCommandLine.ts before the entry pass completed.

The earlier 600-second attempt below is retained as history and is superseded
by these observations. No new compiler tests or mutants have been introduced;
this remains a measurement-only unit.
<!-- current-rerun-end -->

# Previous 600-second attempt (superseded)

The following section records the earlier attempt, before the authorized 75-minute rerun. Its unmeasured statuses do not describe the completed first-error rerun.

checker.ts first stop: **unmeasured**; no checker.ts finding was emitted.
Full mode: distinct reasons and total sites **unmeasured**, not zero.
Totals against meter: **unavailable**; baseline is 569 NotYet and 4,296 Refused.
Branch: `scout/checker-lowering`, based on `a0ba9dd52b23326880b16a7898d446597cfb5a88`.
Not covered: checker.ts first stop, full-mode reason inventory/examples, top-five lowering analysis, and census totals reconciliation.

## Why measurement stopped

The meter's guarded tsc-entry first-error executable hit this scout's explicit
600-second hard deadline and returned **124**. It wrote one header, with 81
resolved implementation sources, zero diagnostics and `checker_rejected: false`,
but **no per-file record and no finding**. Its log was empty. This was a timeout,
not a reported NotYet, Refused, panic, or checker failure.

The companion compiler-directory first-error census had likewise written only
its header. After the entry timeout I sent SIGTERM to that scout-owned process,
following the unit's instruction to stop if the census cannot run as the meter
runs it. The companion returned **143**. Full mode was not started. No census
coverage or blocker total can be inferred from either header.

The drivers, overlay generation, inputs and mode flags reproduce the meter's
measurement path. This was not a complete invocation of `twice-daily.sh`: I used
one pinned checkout, one adapted tree, and separate commands instead of paired
main/area copies. The entry census deadline was 600 seconds; the daily meter's
outer deadline is 7,200 seconds. **This does not establish that a full daily run
would fail.** No attempt was made to change the compiler, filter its traversal,
remove declarations, or bypass the measurement guards to obtain results.

## Observed preparation and checker result

- Fetched origin and created the requested branch at exactly
  `a0ba9dd52b23326880b16a7898d446597cfb5a88`.
- Read CLAUDE.md, stage3/README.md, stage3/meter/README.md and
  stage3/census/latent/README.md before preparing measurements. Also read the
  repository introduction, language contract and memory model.
- `bash cloud/setup.sh` passed, exit 0, within a 900-second deadline.
  Environment sourced: `/workspace/adamic-tools/env.sh`.
- Installed toolchain: Go 1.27.1, clang 20.1.8, Node 24.19.0,
  TypeScript 6.0.3. `nproc` printed **5**; `cpu.max` was `400000 100000`.
- Recorded dependency pins: cohere
  `06af2e399c6c1fb93fd570b774027942747400b4`, nested TypeScript
  `d92d9bfee114c80be2c375d72edae966176e3a4f`.
- Cloned upstream TypeScript v6.0.3 at
  `050880ce59e30b356b686bd3144efe24f875ebc8`; installed stage3 API packages
  with the meter's `npm ci --ignore-scripts --no-audit --no-fund`.
- `stage3/apply.sh /tmp/checker-adapted` passed, exit 0, within a 900-second
  deadline. Its output ends with `/tmp/checker-adapted`.
- Adapted checker.ts SHA-256:
  `213d3a7a2efe1332bfd4ec580434a77db36f227d75d31f685896940e59d73369`.
- An early overlay attempt, before setup finished, failed because
  `cohere/TypeScript/tsc/go.mod` was absent. The logged Go error was
  `go: cannot load module ../../../cohere/TypeScript/tsc listed in go.work file: open ../../../cohere/TypeScript/tsc/go.mod: no such file or directory`.
  After setup completed, generation in a new scratch directory and all three
  meter driver builds passed, exit 0. Production sources were untouched.

Setup timing lines, elapsed seconds:

```text
go ready 0.106
node ready 0.124
clang ready 0.779
tsc ready 0.889
markdown dependencies installed: step-duration 1.036
markdown dependencies ready 2.037
submodules ready 702.690
shared cache off 702.692
go build ready 869.042
test binaries deferred 869.150
build cache warm 869.153
done 869.188
```

The ordinary meter census of `src/tsc/tsc.ts` completed, exit 0. Its root check
and aggregate check both reached lowering, so both passed the checker. They
reported the same **entry-program** first production stop:

```text
src/compiler/core.ts:332:1: Refused: a generator function
export function* mapIterator<T, U>(iter: Iterable<T>, mapFn: (x: T) => U): Generator<U, void, unknown> {
```

Exact fix from the row:

```text
use an explicit iterator object; suspended frames need ownership and cancellation rules before generators can be compiled without a collector (docs/user-iterators.md)
```

The log printed root `Refused 9.179s` and whole-program `Refused 8.770s`.
**This core.ts stop is not checker.ts's first stop.** The ordinary census does
not enumerate checker.ts's blockers after this entry-program refusal.

## Baseline comparison, with observed and read values separated

Read from committed
`stage3/meter/runs/20261103T131714Z.YWfwfQ/main/report.json`:
`latent_lowering.totals` is NotYet **569**, Refused **4,296**,
SkippedDependency **0**, error **0**, panic **0**. Its compiler-mode.json says
first-error mode, per-ref compiler
`0080ee95cf1c71dfadd636fef55189467033dfa0`.

That baseline's checker.ts own-file diagnostic count is **0**, but its tsc entry
has **2** program.ts TS18046 diagnostics and no entry lowering census. The new
ordinary entry measurement and guarded header have **0** diagnostics. These are
observed checker outcomes, not lowering totals.

There are **no measured new complete lowering totals**. A numeric delta against
569/4,296 is undefined, not -569/-4,296 and not a match. The reason this scout
cannot reconcile totals is the incomplete census after timeout and stop.

Read from adaptation 97's code: it rewrites the two caught `e.message` reads in
program.ts through `String(propertyOf(e, "message"))`. This explains the intended
removal of those two baseline checker errors; that causal explanation is a code
reading, not an isolated before/after experiment performed by this scout.

## checker.ts rows and top-five table

`checker-lowering.jsonl.gz` is a valid gzip containing **empty JSONL**. Neither
stream emitted a checker.ts record, so there are no per-site rows to preserve.
An empty artifact denotes missing observations here, not a measured zero.
No reason counts, examples, owners, or lowering recommendations are asserted.

| Rank | checker.ts reason | Sites | Owner / required work |
| --- | --- | --- | --- |
| 1-5 | Unavailable: census timed out before per-file output | Unmeasured | Unmeasured |

## Commands and retained scratch evidence

All long command output was redirected to files, never piped. Builds were under
one 1,200-second timeout including waiting for setup's lock. The successful
build sequence was:

```sh
source /workspace/adamic-tools/env.sh
python3 stage3/census/latent/make_overlay.py /workspace/adamic /tmp/checker-overlay-ready
gofmt -w /tmp/checker-overlay-ready/*.go
python3 stage3/meter/entry_overlay.py /tmp/checker-overlay-ready /tmp/checker-entry-overlay-ready
gofmt -w /tmp/checker-entry-overlay-ready/*.go
go build -buildvcs=false -overlay=/tmp/checker-overlay-ready/overlay.json -o /tmp/checker-latent ./stage3/census/latent/tool
go build -buildvcs=false -overlay=/tmp/checker-entry-overlay-ready/overlay.json -o /tmp/checker-entry ./stage3/census/latent/tool
go build -o /tmp/checker-census ./stage3/census/tool
```

Measurement invocations, each with the shown log redirection:

```sh
LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 timeout --kill-after=30s 1800s /tmp/checker-latent /tmp/checker-adapted/src/compiler /tmp/checker-first-directory.jsonl > /tmp/checker-first-directory.log 2>&1
timeout --kill-after=30s 600s /tmp/checker-census /tmp/checker-adapted/src/tsc/tsc.ts /tmp/checker-entry-census.jsonl > /tmp/checker-entry-census.log 2>&1
LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 timeout --kill-after=30s 600s /tmp/checker-entry /tmp/checker-adapted/src/tsc/tsc.ts /tmp/checker-first-entry.jsonl > /tmp/checker-first-entry.log 2>&1
```

| Artifact | Contents / outcome |
| --- | --- |
| `/tmp/checker-fetch.log` | Fetch and branch creation, exit 0, 120-second fetch limit |
| `/tmp/checker-setup.log` | Successful required setup |
| `/tmp/checker-upstream.log` | Pinned upstream clone, 600-second limit |
| `/tmp/checker-packages.log` | Locked API install, 300-second limit |
| `/tmp/checker-apply.log` | Successful adaptations |
| `/tmp/checker-overlay.log` | Early missing-dependency failure |
| `/tmp/checker-overlay-ready.log` | Successful overlay retry |
| `/tmp/checker-build.log` | Successful bounded build sequence |
| `/tmp/checker-latent-build.log` | Successful latent driver build |
| `/tmp/checker-entry-overlay-ready.log` | Entry overlay generation |
| `/tmp/checker-entry-build.log` | Successful entry driver build |
| `/tmp/checker-census-build.log` | Successful ordinary driver build |
| `/tmp/checker-entry-census.jsonl` | Two ordinary entry observations |
| `/tmp/checker-entry-census.log` | Ordinary root and aggregate results |
| `/tmp/checker-first-entry.jsonl` | 4,834 bytes, one clean header, 81 source paths, no per-file records |
| `/tmp/checker-first-entry.log` | Empty; process timed out, exit 124 |
| `/tmp/checker-first-directory.jsonl` | 172 bytes, one header, no per-file records |
| `/tmp/checker-first-directory.log` | Empty; stopped after entry timeout, exit 143 |

SHA-256 of retained raw streams:

```text
89ec663211a52d0a17f0aac5c0955b27db025b1d8b9fca27c04217e80c9cec04  checker-first-entry.jsonl
2a1ff7f7ced2cbbf79399f1c2ca2f1bd54845c791873b3fa8a8c2a641eb126b8  checker-first-directory.jsonl
c33c001fcc954cf0395b36b5eb312b2ed959be90d682e7a7ea8a0fb2dc3c501e  checker-entry-census.jsonl
```

No implementation, adaptation, test, or compiler source was changed. No tests,
mutants, backend runs or full gate were run for this measurement-only unit.
Validation is limited to inspecting command exits, parsing both incomplete JSONL
streams, checking their record counts, and verifying the empty gzip artifact.
The no-output guards were enabled on both actual latent invocations. No pull
request was opened.
