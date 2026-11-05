# Debug assertion level census

Measured 2026-11-05 on Linux x64, Node v24.19.0. Report only, not for landing.
Base: origin/main `fd30fac46130d994f0cd470652d90dffa84f9518`. TypeScript v6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`.

## Result

| Observation across the complete instrumented oracle | Count |
| --- | ---: |
| Calls to setAssertionLevel | 9 |
| Calls with a non-empty assertionCache | **0** |
| shouldAssertFunction cache writes | 0 |
| Noop swaps | **0** |

All nine calls changed level 0 (None) to 1 (Normal), before any test was active. Both own-key count and defined-entry count were zero. There are **no tests** in which setAssertionLevel was called with a non-empty cache, and no later level changes in this run. The nine calls are the parallel host plus its eight worker processes, once each. No extra setter calls were observed during install/build or VM evaluation.

Deduplicated top stack frames, shared by all nine calls (paths abbreviated to the instrumented tree):

```text
Object.setAssertionLevel (built/local/run.js:1602:20)
src/compiler/sys.ts (built/local/run.js:6525:13)
__init (built/local/run.js:25:56)
```

The source caller is `src/compiler/sys.ts:1974`, its module initialization call. The full stacks, argv, process IDs, old/new levels, own-key counts, defined-entry counts, and active-test fields are in [counters.jsonl.gz](counters.jsonl.gz).

This oracle provides no observed post-cache level change. It also performs zero noop swaps: the harness initializes at Normal and all six assertions guarded by shouldAssertFunction require Normal. Therefore this measurement does not exercise restoration after caching or prove equivalence for arbitrary clients, other environments, or production-level runs. No adaptation was implemented.

## Oracle identity check

Both runs used this base checkout's unmodified `stage3/oracle/run.sh`, the same oracle invoked by `stage3/lane/run.sh`: all runners, eight workers, no filter, default timeout. Both completed install/build with exit 0 and tests/oracle with exit 1, as required for the sanctioned failure.

| Run | Passing | Failing | Pending | Tests wall seconds |
| --- | ---: | ---: | ---: | ---: |
| Main adapted reference | 106,366 | 1 | 0 | 360.340 |
| Instrumented scratch copy | 106,366 | 1 | 0 | 282.377 |

Exact failure in both:

```text
unittests:: Public APIs for typescript.d.ts should be acknowledged when they change
```

Both have exactly `baseline_diffs == ["api/typescript.d.ts"]`. Both pass the unchanged lane checker, including the declaration sanction. For checking the separately invoked oracle, scratch lane evidence directories link each adapted tree and oracle output, and supply execution metadata for the successful common apply and each observed oracle exit. The full lane wrapper itself was not rerun, since instrumentation must be inserted between apply and the oracle.

**Exact title/outcome multisets match: zero additions and zero removals.** The observer records original Mocha pass/fail/pending events in workers and the host's replay. Each title/outcome occurrence is emitted twice. Both raw multisets match exactly (212,732 pass and 2 fail events); all multiplicities are even. Dividing by two preserves duplicate titles and yields 106,366 passing and one failing occurrence. The normalized multiset was retained in scratch. SHA-256 of its uncompressed JSONL bytes: `0ae084170b7b567dec37748821263f079307c9612c3a739a9e1f655ee574fc3b`. Instrumentation did not change pass/fail identities.

The setup Go cache warmup overlapped the reference build/tests, so wall times are context only and are not a performance comparison. Setup exited 0 in 559.917 seconds. Both suites completed without timeout failures.

## Procedure and counter semantics

A fresh clone of system-inc/adamic at origin/main supplied the source. Read `stage3/lane/README.md` and both complete adapted Debug functions before instrumenting. No committed source or runner was changed. Applied with:

```sh
export GOPROXY='https://proxy.golang.org|direct'
timeout --kill-after=60s 1800 bash cloud/setup.sh > /workspace/scratch/setup.log 2>&1
# Installed stage3/api's pinned package-lock.json with npm ci.
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-source
stage3/apply.sh /workspace/scratch/debug-assertion-adapted > /workspace/scratch/apply.log 2>&1
```

The TypeScript input was independently cloned at v6.0.3 and its HEAD checked against stage3/source.json. After apply exited 0, copied the complete output using `shutil.copytree(..., symlinks=True)` to `/workspace/scratch/debug-assertion-instrumented`. Exactly three calls were inserted into that copy's debug.ts, preserving CRLF and the existing assignments:

```ts
// First statement of setAssertionLevel, before the level assignment:
(globalThis as any).__debugAssertionCensus({ event: "setAssertionLevel", oldLevel: currentAssertionLevel, newLevel: level, cacheKeys: Object.keys(assertionCache).length, cacheEntries: Object.keys(assertionCache).filter(key => (assertionCache as any)[key] !== undefined).length });
// Immediately after assertionCache[name] is assigned:
(globalThis as any).__debugAssertionCensus({ event: "cache", name, level, currentLevel: currentAssertionLevel });
// Immediately after Debug[name] is assigned noop:
(globalThis as any).__debugAssertionCensus({ event: "noop", name, level, currentLevel: currentAssertionLevel });
```

Each emitted event appends one JSON line synchronously to the file in `DEBUG_ASSERTION_CENSUS`. Logging errors are not swallowed. `cacheEntries` counts defined cached records, while `cacheKeys` also includes keys cleared by assignment to undefined during restoration. They coincide at zero in every measured call. Stack capture is after the emitter, so its first frame is the instrumented Debug function. Null test means no runnable is active; it is not a missing title for an active test.

Both runs used the same scratch preload below for identity observation. Only the instrumented copy contains the Debug calls. The VM shim exposes the same emitter if the harness evaluates instrumented compiler code in a fresh context.

```js
const fs = require('node:fs');
const path = require('node:path');
const {createRequire} = require('node:module');
let active = null;
globalThis.__debugAssertionCensus = row => {
  if (process.env.DEBUG_ASSERTION_CENSUS) fs.appendFileSync(process.env.DEBUG_ASSERTION_CENSUS, JSON.stringify({...row, pid:process.pid, test:active, argv:process.argv, stack:new Error().stack.split('\n').slice(2)})+'\n');
};
if (['run.js','_mocha','mocha.js'].includes(path.basename(process.argv[1] || ''))) {
  const Mocha = createRequire(path.join(process.cwd(),'package.json'))('mocha');
  const originalRun = Mocha.Runnable.prototype.run;
  Mocha.Runnable.prototype.run = function (...args) {
    active = this.titlePath().filter(Boolean).join(' ');
    return Reflect.apply(originalRun,this,args);
  };
  const originalEmit = Mocha.Runner.prototype.emit;
  Mocha.Runner.prototype.emit = function (...args) {
    const [event, runnable] = args;
    if (process.env.CENSUS_IDENTITIES && ['pass','fail','pending'].includes(event)) {
      fs.appendFileSync(process.env.CENSUS_IDENTITIES, JSON.stringify({event,title:runnable.titlePath().filter(Boolean).join(' '),pid:process.pid})+'\n');
    }
    if (event === 'test end' || event === 'hook end') active = null;
    return Reflect.apply(originalEmit,this,args);
  };
}
const vm = require('node:vm');
const originalRunInNewContext = vm.runInNewContext;
vm.runInNewContext = function(code, context, ...rest) {
  if (context && typeof code === 'string' && code.includes('__debugAssertionCensus')) {
    Object.defineProperty(context, '__debugAssertionCensus', {value:globalThis.__debugAssertionCensus, configurable:true});
  }
  return Reflect.apply(originalRunInNewContext,this,[code,context,...rest]);
};
```

Commands from the base checkout:

```sh
NODE_OPTIONS=--require=/workspace/scratch/census-preload.cjs \
 CENSUS_IDENTITIES=/workspace/scratch/reference-identities.jsonl \
 stage3/oracle/run.sh /workspace/scratch/debug-assertion-adapted /workspace/scratch/reference-oracle \
 > /workspace/scratch/reference-oracle.log 2>&1
NODE_OPTIONS=--require=/workspace/scratch/census-preload.cjs \
 CENSUS_IDENTITIES=/workspace/scratch/instrumented-identities.jsonl \
 DEBUG_ASSERTION_CENSUS=/workspace/scratch/debug-assertion-counters.jsonl \
 stage3/oracle/run.sh /workspace/scratch/debug-assertion-instrumented /workspace/scratch/instrumented-oracle \
 > /workspace/scratch/instrumented-oracle.log 2>&1
python3 stage3/lane/check.py /workspace/scratch/reference-lane > /workspace/scratch/reference-lane-check.log 2>&1
python3 stage3/lane/check.py /workspace/scratch/instrumented-lane > /workspace/scratch/instrumented-lane-check.log 2>&1
```

## Separate positive control

To verify zero counts reflect the oracle rather than a dead instrumentation path, a separate process loaded the instrumented built/local/typescript.js under NODE_ENV=production, called `ts.Debug.assertNode(undefined)`, then `ts.Debug.setAssertionLevel(ts.Debug.getAssertionLevel())`. It exited 0 and logged: initialization 0 to 0 with empty cache, one cache write, one noop swap, and another setter call 0 to 0 with one cached entry/key. Its output is retained separately in scratch. These four events are excluded from all oracle counts. This control did not modify any oracle artifact or test.

## Retained evidence

This commit contains only REPORT.md and counters.jsonl.gz. Setup/apply logs, both oracle reports, original test logs, API baseline diffs, normalized identities, positive-control output, and lane checker reports/verdict logs remain in /workspace/scratch. All census instrumentation remains uncommitted in scratch.
