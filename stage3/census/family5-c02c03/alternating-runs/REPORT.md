# C02/main: executable bytes and alternating lane runs

**All 19 JavaScript bundles are byte-identical. All five new lane runs are green: 106,366 passing, one sanctioned failure, real lane exit 0. Nothing was committed or pushed.**

Main: `fbe5b72ba115a8f3060d4b5229fd578574f4505f`. C02: `28a2e22015d0cf51667fcefa7acb246cab553817`. Both use TypeScript v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`.

## The lane actually executes

The checked-in chain is `stage3/lane/run.sh` → lane/run.py → oracle/run.sh → oracle/run.py. The oracle runs `npm run build` and then `npm test -- --workers=8 --lint=false --no-colors`. package.json selects `hereby runtests-parallel --light=false`. Herebyfile.mjs:532 names `./built/local/run.js`; scripts/build/tests.mjs:133 passes that bundle to Node, and src/testRunner/parallel/host.ts:286 forks `process.argv[1]` for the eight workers. Thus the host and each worker execute the same **run.js**, which bundles the compiler, harness and test-runner code. There is no separate harness.js bundle.

The local compiler/server loaders and their payloads are included below: tsc.js→_tsc.js; tsserver.js→_tsserver.js; typingsInstaller.js→_typingsInstaller.js; tsserverlibrary.js→typescript.js; watchGuard.js is covered too. The nine tracked last-known-good lib/*.js bundles are included as well. Herebyfile.mjs:194 emits linked source maps with sourcesContent:false and external packages.

Beyond the bundles, all six compared launch/build support files match, and **4,776 external JavaScript/native/config dependency files match byte for byte**. The stage3 lane/oracle/observer scripts also match at the two repository shas. See [support comparison](support-comparison.json) and [executor comparison](executor-comparison.json).

## Executable byte comparison

Primary inputs: the untouched original C02 full-lane tree `/workspace/family5-c02-lane/adapted-tree` and the completed main control `/workspace/family5-main-lane-alone/adapted-tree`. The C02 tree used here was not modified by subsequent targeted tests or shared-path input selection. Direct byte equality was tested, not just hash equality. The full [1,499-file manifest](artifact-manifest.json) includes every regular file below built/local and lib. There are no missing files between sides.

| JavaScript file | Bytes on each side | Result |
|---|---:|---|
| `built/local/_tsc.js` | 6,232,191 | identical |
| `built/local/_tsserver.js` | 27,889 | identical |
| `built/local/_typingsInstaller.js` | 10,364 | identical |
| `built/local/run.js` | 13,176,893 | identical |
| `built/local/tsc.js` | 267 | identical |
| `built/local/tsserver.js` | 272 | identical |
| `built/local/tsserverlibrary.js` | 1,013 | identical |
| `built/local/typescript.js` | 9,142,558 | identical |
| `built/local/typingsInstaller.js` | 280 | identical |
| `built/local/watchGuard.js` | 2,314 | identical |
| `lib/_tsc.js` | 6,239,091 | identical |
| `lib/_tsserver.js` | 27,889 | identical |
| `lib/_typingsInstaller.js` | 10,364 | identical |
| `lib/tsc.js` | 267 | identical |
| `lib/tsserver.js` | 272 | identical |
| `lib/tsserverlibrary.js` | 1,013 | identical |
| `lib/typescript.js` | 9,144,216 | identical |
| `lib/typingsInstaller.js` | 280 | identical |
| `lib/watchGuard.js` | 2,314 | identical |

**The executable JavaScript code under test is identical. There is no C02 JavaScript implementation delta to explain red versus green; the fresh alternating sequence supports execution nondeterminism/order as the explanation for the earlier reds.** This statement concerns executable code. The build directories contain the 18 non-executable differences below. In total, 1,481 of the 1,499 artifacts match, including public typescript.d.ts and every standard-library declaration. Every new run's 19 JavaScript files was also re-hashed after execution and matched this initial snapshot.

## Every differing artifact

Offsets are zero-based byte offsets. Both sizes are in bytes. [Hunks for every file](artifact-hunks.md) show the exact first changed context; declaration files have unified source hunks.

| File | Main bytes | C02 bytes | First different byte |
|---|---:|---:|---:|
| `built/local/_tsc.js.map` | 2997654 | 2997654 | 64983 |
| `built/local/compiler/tsconfig.tsbuildinfo` | 53582 | 53582 | 25910 |
| `built/local/compiler/types.d.ts` | 408078 | 408209 | 26127 |
| `built/local/compiler/types.d.ts.map` | 275854 | 275989 | 16640 |
| `built/local/deprecatedCompat/tsconfig.tsbuildinfo` | 47087 | 47087 | 36802 |
| `built/local/harness/tsconfig.tsbuildinfo` | 92987 | 92987 | 40974 |
| `built/local/jsTyping/tsconfig.tsbuildinfo` | 47708 | 47708 | 24549 |
| `built/local/run.js.map` | 6316408 | 6316408 | 90951 |
| `built/local/server/tsconfig.tsbuildinfo` | 80287 | 80287 | 36387 |
| `built/local/services/tsconfig.tsbuildinfo` | 62402 | 62402 | 22587 |
| `built/local/testRunner/tsconfig.tsbuildinfo` | 167608 | 167608 | 66662 |
| `built/local/tsc/tsconfig.tsbuildinfo` | 20556 | 20556 | 12086 |
| `built/local/tsserver/tsconfig.tsbuildinfo` | 80453 | 80453 | 36872 |
| `built/local/typescript.internal.d.ts` | 1546565 | 1546700 | 417061 |
| `built/local/typescript.js.map` | 4430249 | 4430249 | 93969 |
| `built/local/typescript/tsconfig.tsbuildinfo` | 50910 | 50910 | 22307 |
| `built/local/typingsInstaller/tsconfig.tsbuildinfo` | 79935 | 79935 | 36718 |
| `built/local/typingsInstallerCore/tsconfig.tsbuildinfo` | 48370 | 48370 | 24934 |

The differences are completely characterized:

- **Two internal declaration files** gain the requested internal optional Node callable. Public typescript.d.ts stays identical.
- **Three JavaScript source maps** differ only in mappings. Decoding confirms identical generated coordinates, source paths and symbol names; references into compiler/types.ts move by **+1 source line**, with no column or name change. Changed segment counts: _tsc.js.map 2,319; run.js.map 5,655; typescript.js.map 5,580.
- **One declaration source map**, compiler/types.d.ts.map, records that added member.
- **Twelve tsbuildinfo files** differ only in SHA-256 source/declaration hashes; every changed JSON leaf is a hash. There is no embedded-path delta.

[Decoded map explanation](source-map-explanation.json), [every changed build-info field](build-info-explanation.json), and the complete [artifact hunks](artifact-hunks.md). Source maps can be read for error-stack reporting; they are not claimed to be identical.

The declaration hunk is:

```diff
     /** @internal */ emitNode?: EmitNode | undefined;
+    /** @internal */ getPositionOfLineAndCharacter?: ((line: number, character: number, allowEdits?: true) => number) | undefined;
 }
```

## Four recorded lanes per side

There were **two earlier C02 lanes and one earlier main lane**, not two main lanes. The requested fresh order **C02, main, C02, main** was run exactly. One additional final main control makes the requested totals **four C02 and four main records**. The initial historical C02 run is explicitly marked concurrent; it is not an idle-box measurement. All five new runs were sequential and alone on the same box, with eight workers, the standard 40,000 ms test/hook timeout, no test filter, all runners, and Node v24.19.0. No product test or census ran alongside them. Pre-run process snapshots and exact commands are retained per run.

| Record | Side | Conditions | Passing | Failing | Pending | Real lane exit | Every failure, seconds |
|---|---|---|---:|---:|---:|---:|---|
| C02-earlier-1 | c02 | CONCURRENT historical | 106359 | 3 | 0 | **1** | NoTruncation body 46.964s; relationComplexityError before-all 55.109s; API 4.085s |
| C02-earlier-2 | c02 | idle historical | 106354 | 3 | 0 | **1** | recursiveConditionalTypes before-all 42.168s; templateLiteralTypes1 before-all 43.073s; API 2.481s |
| main-earlier-1 | main | idle historical | 106366 | 1 | 0 | **0** | API 2.486s |
| new-1 | c02 | idle; requested alternation | 106366 | 1 | 0 | **0** | API 2.440s |
| new-2 | main | idle; requested alternation | 106366 | 1 | 0 | **0** | API 2.413s |
| new-3 | c02 | idle; requested alternation | 106366 | 1 | 0 | **0** | API 1.965s |
| new-4 | main | idle; requested alternation | 106366 | 1 | 0 | **0** | API 2.504s |
| new-5 | main | idle; extra count control | 106366 | 1 | 0 | **0** | API 1.946s |

API = `unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`, the sanctioned baseline-content failure. All npm test children exited 1 because this failure is expected; the validated lane exits 0 only when the exact required counts and sanctioned failure match. [All run records](all-runs.json) retain the individual install/build/test phase exits and durations.

The first historical C02 hook failure skipped six suite bodies and added one failed hook: net five missing reported results. The second historical C02 run lost two such suites: net ten. All new runs report the full **106,367** total.

## Every failure: exact case or hook name and observed time


### C02-earlier-1 (c02)

- `fourslash tests tests/cases/fourslash/codeFixClassImplementInterfaceNoTruncation.ts fourslash test codeFixClassImplementInterfaceNoTruncation.ts runs correctly` — 46964.014760 ms; cause `timeout`; timeout 40000 ms.
- `compiler tests compiler tests for tests/cases/compiler/relationComplexityError.ts "before all" hook for "Correct errors for tests/cases/compiler/relationComplexityError.ts"` — 55109.133596 ms; cause `timeout`; timeout 40000 ms.
- `unittests:: Public APIs for typescript.d.ts should be acknowledged when they change` — 4085.339037 ms; cause `baseline-content`.

### C02-earlier-2 (c02)

- `compiler tests compiler tests for tests/cases/compiler/recursiveConditionalTypes.ts "before all" hook for "Correct errors for tests/cases/compiler/recursiveConditionalTypes.ts"` — 42167.509889 ms; cause `timeout`; timeout 40000 ms.
- `conformance tests conformance tests for tests/cases/conformance/types/literal/templateLiteralTypes1.ts "before all" hook for "Correct errors for tests/cases/conformance/types/literal/templateLiteralTypes1.ts"` — 43073.104485 ms; cause `timeout`; timeout 40000 ms.
- `unittests:: Public APIs for typescript.d.ts should be acknowledged when they change` — 2480.736526 ms; cause `baseline-content`.

### main-earlier-1 (main)

- `unittests:: Public APIs for typescript.d.ts should be acknowledged when they change` — 2485.732159 ms; cause `baseline-content`.

### new-1 (c02)

- `unittests:: Public APIs for typescript.d.ts should be acknowledged when they change` — 2440.253907 ms; cause `baseline-content`.

### new-2 (main)

- `unittests:: Public APIs for typescript.d.ts should be acknowledged when they change` — 2412.630164 ms; cause `baseline-content`.

### new-3 (c02)

- `unittests:: Public APIs for typescript.d.ts should be acknowledged when they change` — 1965.275734 ms; cause `baseline-content`.

### new-4 (main)

- `unittests:: Public APIs for typescript.d.ts should be acknowledged when they change` — 2503.786440 ms; cause `baseline-content`.

### new-5 (main)

- `unittests:: Public APIs for typescript.d.ts should be acknowledged when they change` — 1945.746219 ms; cause `baseline-content`.

## Interpretation and measurement notes

New C02 results: **two green out of two**. New main results: **three green out of three**. Including the requested earlier records, C02 is **two green/two red** (one red was concurrent), and main is **four green/zero red**. The earlier C02 reds did not reproduce in the requested alternating sequence. The JavaScript identity is established independently of the timing observations; this small series does not identify a particular scheduler, GC or cache mechanism.

All five fresh output trees began without .parallelperf.json and without mytest.config; their scheduling-input snapshots are retained. Each output was fresh. After a run ended, complete logs/reports and all built-artifact hashes were preserved before its generated tree was removed to keep space available. The two original comparison trees remain available at the paths above.

An environment-context refresh terminated the original coordinator during new-2, while its independently sessioned main lane continued and completed. Its real exit **0** was recovered from the retained kernel wait status of the unique session-leader process PID 65310, started at the recorded run time. [Raw exit recovery](run-2-main/exit-recovery.json) preserves /proc/PID/stat field 52 and its decoding; the exit was not inferred from the green verdict. The lane's internal wall time is used for that record because the parent wall clock was lost. There was an idle gap before new-3 while the coordinator was recovered. The resumed coordinator was detached and recorded durable child PIDs. No two lane workloads overlapped.

No adaptations, ledger entries, internal/ code, test timeouts or worker counts were changed. **No commits or pushes were made.** The worktree was restored to its original commit 16fbb39cdeaefa79e9f23b3ab0f02cf8a32ad365. This report and its evidence are local only.
