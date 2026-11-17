# Filter freeze evidence — #ev14ch9

Report only, on Adamic `cf468df0747075e50d326a97ff03e8530d7a1a88`.
TypeScript v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`; Node v24.19.0, Linux x64.
No adaptation change or landing branch. Only this report directory is committed.

Observed writes: **0**. The complete four-worker corpus matches **106,366 / 1 / 0**
(passing / failing / pending), with only the expected API baseline failure. The lane
verdict remains **FAIL** because its guard requires eight workers. Four sites never
take an array identity path in these workloads.

## Census and measurement

Repeated the original census's stock TypeScript `getResolvedSignature` traversal on the
complete output of unchanged `stage3/apply.sh` (713 src roots). It resolves 157 core filter calls,
including exactly the same 16 narrowing calls as the previous census at
`03c140ec28a249b8b0f13eb24e34fc3d3df05266`: overload 1 has five, overload 5 has eleven.
No added or removed site, changed quoted call, changed position, or changed overload.
See [set comparison](evidence/set-comparison.json), [resolved calls](evidence/resolved-calls.json)
and [site list](evidence/sites.json). The census scripts were copied from
`typescript/filter-overload-count` at `888d2a68c6418d3c760ef550f9b1d49d83377b15`.

Only those 16 call expressions are wrapped in a private adapted copy. Each evaluates
its argument once, counts entry before executing the original filter, counts completion
when it returns, counts `result === argument`, and
on an array identity records provenance in a WeakMap and calls `Object.freeze(result)`.
Both aliases still denote the actual original array. Undefined identities are counted
separately and cannot be frozen arrays. Nothing copies, proxies, or substitutes an array.
The other 141 filter call expressions and filter itself are untouched.
A second checker traversal confirms all 157 calls retain their original resolved
overloads, argument types, and result types: [type preservation](evidence/type-preservation.json).
The self and lane instrumented source hashes match.

A Node preload installs a V8 inspector session with pause-on-all-exceptions, so it
observes caught as well as escaping freeze TypeErrors. It records the exception's
stack, writer frames, current Mocha test/hook title, file when available (or CLI argv), and
freezing sites found by looking up local array identities in the provenance WeakMap.
The observer requests 100-frame Error stacks and also records inspector frames.
It resumes the original exception immediately; production and Mocha keep their normal
handling. Harness VM loads of the instrumented compiler receive the same observer.
Both compiler bundles contain `"use strict"`. Normal exit and worker termination flush
per-process counters. Zero-count sites are retained, not discarded.

## 1. Writes after the call

**Zero observed post-call array writes, at every site, in self-compilation and the canonical whole lane. No test fails because of a freeze throw.** Per-site zeros appear in the table below; there is no production writer to name. This is a workload observation, not a proof for all inputs. The historical eight-worker run also records zero freeze writes.

The observer's broad message match caught the same six deliberate decorator-test TypeErrors
in each completed whole lane (12 observations total):
`Cannot add initializers after decoration has completed`. All six distinct errors come from the explicit
throw in `src/compiler/factory/emitHelpers.ts:783`, before any `extraInitializers.push`,
and all six tests assert that error. They are unrelated to Object.freeze, have no frozen
array aliases in their writer scopes, and are retained with their full stacks and
classification in [counts](evidence/counts.json) and [source proof](evidence/decorator-exception-proof.json).
The raw observer field `freezeThrows` counts message candidates, including these six per run;
`freeze_write_throws` in the aggregate records the actual freeze-write classification.
No candidate is silently dropped or left unclassified.

In the eight-worker lane, checker.ts:9417 and :9419 are never called; transformers/ts.ts:1088
and services/callHierarchy.ts:482 are called but never take an identity path. Thus these
runs do not exercise an array alias at those four sites. The final repeat's coverage
is reported with its counts below.

## 2. Calls and identity paths

Canonical numbers: **calls / identity results / array identities** in each cell.
Array identities are the copy allocation count; undefined identities are excluded from that number.

| Site | Overload | Self | Whole lane, 4 workers | Writes, self / lane | Writer |
| --- | ---: | ---: | ---: | ---: | --- |
| `src/compiler/factory/utilities.ts:761:12` | 5 | 0 / 0 / 0 | 484 / 484 / 156 | 0 / 0 | none |
| `src/compiler/commandLineParser.ts:2550:16` | 1 | 7 / 7 / 7 | 5,968 / 5,967 / 5,967 | 0 / 0 | none |
| `src/compiler/commandLineParser.ts:3157:32` | 1 | 2 / 0 / 0 | 2,245 / 30 / 30 | 0 / 0 | none |
| `src/compiler/checker.ts:6586:39` | 5 | 0 / 0 / 0 | 16 / 13 / 13 | 0 / 0 | none |
| `src/compiler/checker.ts:9417:28` | 1 | 0 / 0 / 0 | 0 / 0 / 0 | 0 / 0 | none |
| `src/compiler/checker.ts:9419:24` | 1 | 0 / 0 / 0 | 0 / 0 / 0 | 0 / 0 | none |
| `src/compiler/checker.ts:10370:48` | 1 | 0 / 0 / 0 | 1,840 / 1,621 / 1,621 | 0 / 0 | none |
| `src/compiler/checker.ts:16764:27` | 5 | 1,357 / 1,357 / 1,355 | 238,252 / 238,142 / 232,451 | 0 / 0 | none |
| `src/compiler/checker.ts:47509:16` | 5 | 5 / 4 / 4 | 35,872 / 1,873 / 1,873 | 0 / 0 | none |
| `src/compiler/transformers/ts.ts:1088:38` | 5 | 0 / 0 / 0 | 88 / 0 / 0 | 0 / 0 | none |
| `src/services/services.ts:745:94` | 5 | 0 / 0 / 0 | 33 / 5 / 5 | 0 / 0 | none |
| `src/services/services.ts:755:94` | 5 | 0 / 0 / 0 | 27 / 4 / 4 | 0 / 0 | none |
| `src/services/services.ts:780:81` | 5 | 0 / 0 / 0 | 37 / 6 / 6 | 0 / 0 | none |
| `src/services/services.ts:790:81` | 5 | 0 / 0 / 0 | 30 / 4 / 4 | 0 / 0 | none |
| `src/services/callHierarchy.ts:482:19` | 5 | 0 / 0 / 0 | 64 / 0 / 0 | 0 / 0 | none |
| `src/services/codefixes/fixAddMissingParam.ts:143:46` | 5 | 0 / 0 / 0 | 53 / 53 / 53 | 0 / 0 | none |
| **Total** | | **1,371 / 1,368 / 1,366** | **285,009 / 248,202 / 242,183** | **0 / 0** | |

Self has 2 undefined identities; the canonical lane has 6,019.
Every entry completes at every site in both canonical runs; no predicate-throw invocation is missing from the call totals.

Canonical coverage: uncalled sites: `src/compiler/checker.ts:9417:28`, `src/compiler/checker.ts:9419:24`.
Called sites with no array identity: `src/compiler/transformers/ts.ts:1088:38`, `src/services/callHierarchy.ts:482:19`.
These workloads therefore do not establish the no-write premise on every possible identity path at all 16 sites.

Historical eight-worker run (**completed calls / identities / array identities**; entry count unavailable):

| Site | Historical counts | Writes | Writer |
| --- | ---: | ---: | --- |
| `src/compiler/factory/utilities.ts:761:12` | 484 / 484 / 156 | 0 | none |
| `src/compiler/commandLineParser.ts:2550:16` | 5,968 / 5,967 / 5,967 | 0 | none |
| `src/compiler/commandLineParser.ts:3157:32` | 2,245 / 30 / 30 | 0 | none |
| `src/compiler/checker.ts:6586:39` | 16 / 13 / 13 | 0 | none |
| `src/compiler/checker.ts:9417:28` | 0 / 0 / 0 | 0 | none |
| `src/compiler/checker.ts:9419:24` | 0 / 0 / 0 | 0 | none |
| `src/compiler/checker.ts:10370:48` | 1,840 / 1,621 / 1,621 | 0 | none |
| `src/compiler/checker.ts:16764:27` | 238,251 / 238,141 / 232,450 | 0 | none |
| `src/compiler/checker.ts:47509:16` | 35,872 / 1,873 / 1,873 | 0 | none |
| `src/compiler/transformers/ts.ts:1088:38` | 88 / 0 / 0 | 0 | none |
| `src/services/services.ts:745:94` | 33 / 5 / 5 | 0 | none |
| `src/services/services.ts:755:94` | 27 / 4 / 4 | 0 | none |
| `src/services/services.ts:780:81` | 37 / 6 / 6 | 0 | none |
| `src/services/services.ts:790:81` | 30 / 4 / 4 | 0 | none |
| `src/services/callHierarchy.ts:482:19` | 64 / 0 / 0 | 0 | none |
| `src/services/codefixes/fixAddMissingParam.ts:143:46` | 53 / 53 / 53 | 0 | none |
| **Total** | **285,008 / 248,201 / 242,182** | **0** | |

All counts and per-process metadata: [counts.json](evidence/counts.json). All candidate exception stacks, classifications and writer contexts: [throws.json](evidence/throws.json).

Identity includes undefined. `frozen_arrays` counts array identities only, including
empty arrays and repeated returns of the same array. Those are the observed allocations
a copy-on-identity implementation would add, assuming it copies every array identity;
undefined identities do not require an array allocation. Counts measure these workloads,
not all possible uses or unique arrays. Wall times include instrumentation overhead
and are not a copying-performance benchmark.

## 3. Identity as a premise

The repeated static count finds **0 direct comparisons with the argument** at overloads
1 and 5: [static identity evidence](evidence/static-identity.json).
**A result stored in a slot and compared later isn't something a static count sees.**
This local count is not a whole-program identity proof. The full unfiltered lane result
below is the runtime witness for the exercised behavior under freezing. Freezing
preserves identity, so that witness does not itself prove that changing identity by
copying would preserve all behavior.

## 4. What ran and control

The corrected scratch control imports the actual adapted `built/local/typescript.js` filter.
Its seven normal calls include five array identities, one undefined identity, and one
nonidentity result. Push, splice, sort, an index assignment and a length assignment
are each attempted through both argument and result aliases: **10 TypeErrors**.
The V8 observer captures all ten despite the control catching every one, maps each
to the control's freezing site, and records its writer stack. The nonidentity input
and result remain mutable. A separate throwing-predicate control records one entry and zero completions, proving
that an abrupt filter call is counted. Control exits 0. It is a separate scratch program, never
added to either measured compiler tree or corpus. The executed scratch program was removed after the control. Its reproduction source is retained
here as evidence, with no deliberate write added to a production caller.

Exact commands (working directory shown by `cd`; source `environment.sh` exports the
scratch temp directories, Go cache, TypeScript source and Node compile-cache setting):

```sh
git -C /workspace/adamic fetch origin main typescript/filter-overload-count
git -C /workspace/adamic worktree add --detach /workspace/scratch/filter-freeze/repo cf468df0747075e50d326a97ff03e8530d7a1a88
cd /workspace/scratch/filter-freeze/repo
git submodule update --init --recursive --depth 1
npm ci --prefix stage3/api
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /workspace/scratch/filter-freeze/typescript
source /workspace/scratch/filter-freeze/env.sh
npm ci --prefix stage3/api
bash stage3/apply.sh /workspace/scratch/filter-freeze/adapted
node stage3/census/filter-freeze/resolve.cjs /workspace/scratch/filter-freeze/adapted stage3/census/filter-freeze/evidence/resolved-calls.json
node stage3/census/filter-freeze/contexts.cjs /workspace/scratch/filter-freeze/adapted stage3/census/filter-freeze/evidence/contexts.json
node stage3/census/filter-freeze/instrument.cjs /workspace/scratch/filter-freeze/adapted stage3/census/filter-freeze/evidence/instrumented-sha256.json
node stage3/census/filter-freeze/resolve.cjs /workspace/scratch/filter-freeze/adapted stage3/census/filter-freeze/evidence/instrumented-resolved-calls.json
node stage3/census/filter-freeze/validate-types.cjs
```

The reproduction scripts were extracted from the previous census commit with
`git archive 888d2a68c6418d3c760ef550f9b1d49d83377b15 stage3/census/filter-overload-count`.
The private adapted tree's `npm ci --no-audit --no-fund` completed during apply's
final source steps, before packaging; the lane's own install ran again normally.
No source instrumenting began until apply exited 0 and the checker comparison passed.

Self-compile (the full bootstrap command and environment are in [run-self.sh](run-self.sh)):

```sh
cd /workspace/scratch/filter-freeze/adapted
npm ci --no-audit --no-fund
FILTER_FREEZE_OUT=/workspace/scratch/filter-freeze/build-entry-counts NODE_OPTIONS=--require=/workspace/scratch/filter-freeze/repo/stage3/census/filter-freeze/observe.cjs npm run build
FILTER_FREEZE_OUT=/workspace/scratch/filter-freeze/self-entry-counts NODE_OPTIONS=--require=/workspace/scratch/filter-freeze/repo/stage3/census/filter-freeze/observe.cjs node built/local/tsc.js -b src/compiler src/tsc --force --pretty false
```

The adapted tsc actually ran, with force rebuilding both projects despite bootstrap's
existing build information. This project configuration emits declarations only.
The 16-site bootstrap count is separately retained and is zero; it is excluded from
self-compile numbers. Build and self-compile each exit 0.

Control and whole lane:

```sh
cd /workspace/scratch/filter-freeze/repo
cp stage3/census/filter-freeze/control.cjs /workspace/scratch/filter-freeze/control.cjs
FILTER_FREEZE_OUT=/workspace/scratch/filter-freeze/control-entry-counts node --require ./stage3/census/filter-freeze/observe.cjs /workspace/scratch/filter-freeze/control.cjs /workspace/scratch/filter-freeze/adapted/built/local/typescript.js
rm /workspace/scratch/filter-freeze/control.cjs
FILTER_FREEZE_OUT=/workspace/scratch/filter-freeze/lane-counts NODE_OPTIONS=--require=/workspace/scratch/filter-freeze/repo/stage3/census/filter-freeze/observe.cjs python3 stage3/census/filter-freeze/run-lane.py /workspace/scratch/filter-freeze/lane
```

`run-lane.py` executes the pinned `stage3/lane/run.py` in memory, adding only a hook
between successful apply and the normal oracle to instrument that private output.
The installed lane, oracle, adapters, expected counts, timeout and API sanctions are
not edited. The oracle runs its normal install, build and `npm test -- --workers=8
--lint=false --no-colors`, with all runners and no test filter. The resulting
execution, oracle reports and failure stacks are retained below.

The initial eight-worker lane completed with **106,354 passing, 3 failing, 0 pending**.
Install/build exit 0; tests/oracle exit 1; lane exit 1. Its test phase took 1475.193 s.
Only `api/typescript.d.ts` differs. The expected public API baseline failure is joined
by two 40,000 ms hook timeouts:

- `compiler tests ... tests/cases/compiler/largeControlFlowGraph.ts "before all" hook for "Correct errors for tests/cases/compiler/largeControlFlowGraph.ts"` (40,803.883 ms).
- `compiler tests ... tests/cases/compiler/conditionalTypeDiscriminatingLargeUnionRegularTypeFetchingSpeedReasonable.ts "before all" hook for "Correct errors for tests/cases/compiler/conditionalTypeDiscriminatingLargeUnionRegularTypeFetchingSpeedReasonable.ts"` (46,046.547 ms).

These are timeouts, not freeze throws. Exact full names and stacks are in the
[initial oracle report](evidence/lane8-oracle-report.json), and the unfiltered initial
lane verdict is retained in [lane8-verdict.txt](evidence/lane8-verdict.txt).
This first run is not a clean lane witness. Its original counter increments on return,
so its historical `calls` values are **completed calls**, not observed entries; a predicate
throw could be missed. Array-identity and freeze-write observations remain recorded.
The original observer, wrapper and control source are retained as `evidence/return-counter-*`.
An initial four-worker repeat was cancelled with SIGTERM upon finding this counter bug;
its partial evidence is labeled `aborted4`, excluded from the canonical counts and witness.
The corrected self-compile and canonical four-worker whole lane observe entries and
completions separately. No historical result is silently substituted.

The canonical fresh whole lane uses the corrected entry/completion counter and four workers, with the same tests and 40,000 ms timeout:

```sh
FILTER_FREEZE_WORKERS=4 FILTER_FREEZE_OUT=/workspace/scratch/filter-freeze/lane-entry4-counts NODE_OPTIONS=--require=/workspace/scratch/filter-freeze/repo/stage3/census/filter-freeze/observe.cjs python3 stage3/census/filter-freeze/run-lane.py /workspace/scratch/filter-freeze/lane-entry4
```

This changes only the normal oracle `--workers=4` argument, not test selection or timeout.

Canonical four-worker result: **106,366 passing, 1 failing, 0 pending**.
The stage3 lane verdict is **FAIL**. Apply exit 0; install/build exits 0/0; tests/oracle exits 1/1.
Tests take 1545.992 s; whole lane takes 1652.851 s.
All runners run; `tests` is null; the timeout remains 40,000 ms. Baseline paths: api/typescript.d.ts.

Exact verdict: `FAIL stage3 landing lane: workers: expected 8, observed 4 [causes: timeout=0, baseline-content=1, baseline-missing=0, exception=0, other=0]`.

Failures:

- `unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`: The baseline file api/typescript.d.ts has changed. (Run "hereby baseline-accept" if the new baseline is correct.)

See [canonical lane report](evidence/lane4-report.json), [oracle report and failure stacks](evidence/lane4-oracle-report.json), and [verdict](evidence/lane4-verdict.txt).
The two initial timeout cases complete successfully in this unfiltered repeat. The corpus
matches the lane's expected 106,366/1/0 result, exact failure title and sole baseline path;
the API comparison accepts all 276 composed declarations as sanctioned. The **only** lane
guard error is `workers: expected 8, observed 4`. Thus this is a complete corpus witness
under freezing, while the unchanged lane verdict remains FAIL on its worker policy.
No expectation or guard was relaxed to make it green.

Raw per-process counters and exception JSONL are archived separately for self, canonical lane, control, initial lane, and cancelled repeat. Bootstrap counts are separate and zero. The source hash table and strict-bundle proof identify the actual private compiler sources and outputs. Only this report directory is staged for the authorized evidence branch; no adaptation is committed.


Aggregation command:

```sh
python3 stage3/census/filter-freeze/aggregate.py self /workspace/scratch/filter-freeze/self-entry-counts lane4 /workspace/scratch/filter-freeze/lane-entry4-counts lane8 /workspace/scratch/filter-freeze/lane-counts control /workspace/scratch/filter-freeze/control-entry-counts bootstrap /workspace/scratch/filter-freeze/build-entry-counts
```
