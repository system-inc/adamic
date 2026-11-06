# 99-host-require: isNodeLikeSystem drops its require test, and performanceCore imports node:perf_hooks

Task #00qn5az, under Adamic's ruling of Nov 6. lane: run as part of the wave
Correctness cuts it into; Loom's run of that wave is the lane's proof.

This is a JavaScript-changing adaptation in two files. In core.ts,
isNodeLikeSystem no longer tests `typeof require !== "undefined"`, and 30's
`declare const require: unknown;` goes, since nothing else in core.ts reads
it. In performanceCore.ts, tryGetPerformance reads perf_hooks through a
static top-level `import * as perfHooks from "node:perf_hooks";`, in 93's
idiom, where it called `require("perf_hooks")`, and 65's
`declare const require: (specifier: "perf_hooks") => any;` goes, since
nothing else in performanceCore.ts reads it. No public API changes:
isNodeLikeSystem is `@internal` and keeps its signature, and tryGetPerformance
is not exported.

Input: microsoft/TypeScript v6.0.3, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`. Adamic base: main `521ca101`.

## The order

The directory is named to run after 99-correctness-axolotl-q. 99q pins
isNodeLikeSystem's whole text by hash, with the require test in it, and it
already turns `!(process as any).browser` into `!process.browser`, read
through 30's declared `browser?: unknown`. So the ruling's `process.browser`
edit is on main before this adaptation runs, and this adaptation takes 99q's
text as its before state and leaves that line alone. A name under
`99-cli-` would sort before 99q, and 99q would refuse the tree.
99-wasm-crypto-import runs after this one and touches only sys.ts.

## The rows

Line numbers are the adapted tree's before this adaptation, from the
adapter's evidence. They are never used as edit addresses.

| Row | Site | Edit |
| --- | --- | --- |
| test | core.ts:2616, isNodeLikeSystem | `&& typeof require !== "undefined"` removed; the `!process.browser` line ends the expression |
| declare | core.ts:2614, after the process view on the same line | ` declare const require: unknown;` removed |
| import | performanceCore.ts:2 | `import * as perfHooks from "node:perf_hooks";` added after the `./_namespaces/ts.js` import |
| binding | performanceCore.ts:35, in tryGetPerformance's try block | `const { performance } = require("perf_hooks") as { performance?: Performance \| undefined };` becomes `const { performance } = perfHooks;` |
| declare | performanceCore.ts:104, the last statement | `declare const require: (specifier: "perf_hooks") => any;` removed, with the LF lines 65 appended, so the file ends as upstream's does |

The binding keeps upstream's destructuring, which is how
internal/oracle/testdata/node_process_performance_core.a shapes it
(`const { performance } = nativeHooks;`) and the one perf_hooks namespace use
Lower admits. Lower refuses the namespace as a value anywhere else, a cast
around it included, so 65's cast goes with the require. The binding stays
under `if (isNodeLikeSystem())` in its try block, and `if (performance)` stays
as upstream wrote it. Both files are CRLF, and so is everything written.

## The ruling

Adamic refuses an ambient value declaration, and both files read `require`
through one. Adamic ruled one JavaScript-changing adaptation that takes the
ambient require out of both: isNodeLikeSystem decides from `process` alone,
and perf_hooks arrives the way 93 brings in fs, path and os.

## JavaScript

Stock 6.0.3 transpiles both files before and after. With comments removed,
only these lines differ:

```diff
     return typeof process !== "undefined"
         && !!process.nextTick
-        && !process.browser
-        && typeof require !== "undefined";
+        && !process.browser;
```

```diff
 import { isNodeLikeSystem } from "./_namespaces/ts.js";
+import * as perfHooks from "node:perf_hooks";
 function tryGetPerformance() {
     if (isNodeLikeSystem()) {
         try {
-            const { performance } = require("perf_hooks");
+            const { performance } = perfHooks;
```

What changes when tsc runs:

- On a host with `process.nextTick` and no `require`, tsc now takes the Node
  path. The oracle's ES-module Node is such a host: isNodeLikeSystem goes from
  false to true there, so sys.ts's getNodeSystem now runs there.
- On that host, tryGetPerformance now takes the perf_hooks path, where it took
  the browser global before. The performance object is the same one
  (`globalThis.performance === perf_hooks.performance` on Node), and
  shouldWriteNativeEvents goes from true to false, which is upstream's answer
  on Node. So performance.ts's enable writes native marks there only under a
  profiler or a debugger, as it does under CommonJS.
- Under CommonJS, where `require` is defined, as in upstream's bundled tsc,
  nothing changes: isNodeLikeSystem was already true and tryGetPerformance
  already took the perf_hooks path.
- perf_hooks loads when performanceCore.ts is evaluated, not on the first
  call. On Node that is unobservable except in timing. On a host without a
  `node:perf_hooks` module, the module now fails to load, where the require
  failed inside the try and was ignored. 93's `node:fs` import already makes
  that true of sys.ts.

## What the change exposes (not fixed here)

With isNodeLikeSystem true on a host without CommonJS globals, the rest of
getNodeSystem runs there, and it still reads them:

- `__filename`: isFileSystemCaseSensitive (sys.ts:1722) and executingFilePath
  (sys.ts:1492), both run unconditionally when getNodeSystem runs. Loading the
  whole adapted src/compiler as ES modules now throws
  `ReferenceError: __filename is not defined in ES module scope` at
  sys.ts:1722, during sys.ts's module evaluation. Before this adaptation the
  same load completed with `sys` undefined.
- `require(`: four calls remain in what sys.ts and tsc.ts reach. They are
  listed with the rest of the census on cli/host-require-adapt-evidence.
  sys.ts:1591 (source-map-support, tsc.ts:16 under `NODE_ENV=development`,
  caught), sys.ts:1644 (inspector, under `--generateCpuProfile`, not caught),
  tracing.ts:64 (fs, under `--generateTrace`, caught and rethrown), and
  sys.ts:1613 (System.require, which nothing in src/compiler or src/tsc calls;
  only src/server does).

## The adapter

adapt.cjs parses both files with stock typescript@6.0.3. It finds
isNodeLikeSystem, tryGetPerformance, the imports and the top-level
`declare const` statements by name, never by line. isNodeLikeSystem's head
must be the reviewed signature and its body exactly 99q's text or exactly this
adaptation's. The process view must be 30's in both states. Each require
declaration must be its reviewed text at its reviewed place (core.ts's after
the process view on the same line, performanceCore.ts's last, after its LF
lines), or absent. The perf_hooks import must be the reviewed text, directly
after the one `./_namespaces/ts.js` import, or absent. The binding must be the
reviewed text in tryGetPerformance's try block under its isNodeLikeSystem
guard. Before, each file names `require` exactly twice and performanceCore.ts
never names `perfHooks`. After, neither file names `require`, and
`perfHooks` appears exactly twice.

The two files are in exactly one of two states: every piece original, or
every piece adapted. Anything else is refused before writing. A second run
writes zero files without building a program. On an original tree, the
adapter type-checks the tree with 47's compiler program and checks by symbol
that each require declaration is read only at the site being replaced, and
that the read resolves to the file's own declaration.

The adapter plans every edit and re-parses the planned texts, which must be
CRLF throughout and exactly the adapted state. Stock 6.0.3 then transpiles the
before and planned texts, with comments kept and with comments removed. The
adapted JavaScript must equal upstream's with the require test removed, the
import added and the binding replaced. Stock 6.0.3's transpileDeclaration must
give identical declarations before and after, with stripInternal and without
it. A planned program, 47's with the two planned texts, must bind the
`perfHooks` read to the import. core.ts's diagnostics must be unchanged, and
performanceCore.ts's must be unchanged plus at most the import's own. 47's
program has no Node types (stage3 never runs npm), so the specifier draws
TS2591 `Cannot find name 'node:perf_hooks'`, which must be exactly what
sys.ts's `node:fs` import draws in the same program, with the name changed.
Where Node's types are present, the import must resolve to @types/node's
`node:perf_hooks`, which exports performance. Only then does the adapter check
that the files haven't changed and write them.

## Witnesses and the mutant

witnesses.cjs reads the tree before this adaptation and the tree after it.
It runs each probe in a fresh Node process and records what the probe prints.
The ES-module rows load the source the way oracle/node.mjs does: Node's
stripTypeScriptTypes in transform mode, format module, so the program has no
`require`, `module` or `__filename`. The CommonJS rows load stock 6.0.3's
CommonJS output, where `require` is defined as in upstream's bundled tsc.
isNodeLikeSystem is extracted from core.ts with its ambient declarations.
performanceCore.ts runs whole, with its `./_namespaces/ts.js` import bound to
the same tree's extracted isNodeLikeSystem, and the probe reads
tryGetNativePerformanceHooks() and timestamp(). The last row loads the whole
of src/compiler as ES modules through `_namespaces/ts.ts`, with tsc's `.js`
specifiers resolved to their `.ts` sources as stage3/drivers/tsc-entry/node.mjs
does. Its outcome is recorded, not ruled.

Five rows are ruled. Five of five hold on Node 24.21.0:

| Row | Before | After |
| --- | --- | --- |
| ES module, Node | isNodeLikeSystem false, `require` undefined | true |
| ES module, `process.browser` set | false | false |
| CommonJS, Node | true, `require` a function | true |
| performanceCore, ES module | the browser global: shouldWriteNativeEvents true, the perf_hooks performance object | the perf_hooks path: shouldWriteNativeEvents false, the same object |
| performanceCore, CommonJS | the perf_hooks path: false, the same object | the same |
| src/compiler, ES module (recorded) | loads; isNodeLikeSystem false; `sys` undefined | `ReferenceError: __filename is not defined in ES module scope`, at isFileSystemCaseSensitive, sys.ts:1722 |

In both performanceCore rows, `performanceTime` is the performance object and
timestamp() is finite and monotonic, before and after.

The mutant puts `&& typeof require !== "undefined"` back in the after tree's
isNodeLikeSystem, in memory, in the extracted function and in the whole-graph
load. The ES-module Node row and the performanceCore ES-module row then go
red: isNodeLikeSystem answers false and tryGetPerformance takes the browser
global again. The whole-graph row loads with `sys` undefined, as before. 3 of
5 are as ruled, and the run exits 1. Outputs go to the paths given on the
command line, never this directory, because every file in this directory
feeds apply's input key.

Refused, writing nothing (both files' hashes checked after each run):

| Tree | Adapter |
| --- | --- |
| core.ts adapted, performanceCore.ts original | `half-adapted tree: core.ts adapted, performanceCore.ts original` |
| core.ts original, performanceCore.ts adapted | `half-adapted tree: core.ts original, performanceCore.ts adapted` |
| the require test removed, its declaration kept | `half-adapted core.ts: isNodeLikeSystem body adapted, require declaration original` |
| the import added, the rest original | `half-adapted performanceCore.ts: import adapted, binding original, require declaration original` |
| adapted, 65's declaration kept | `half-adapted performanceCore.ts: import adapted, binding adapted, require declaration original` |
| 99q's before text, `(process as any).browser` | `uninventoried isNodeLikeSystem body: ...` |
| original, with `require` read in another core.ts function | `core.ts names require outside the reviewed sites` |
| original, with `const perfHooks = 0;` in performanceCore.ts | `performanceCore.ts names perfHooks outside the reviewed sites` |
| adapted, the binding moved out of its try block | `the perf_hooks binding left its try block` |

## Measured

The whole chain, stage3/apply.sh on main `521ca101` with this directory,
exits 0. Its row is `99-host-require | 2 | 4 | 6`, 99-wasm-crypto-import
still applies after it, and the total goes from 6,193 added and 5,823 removed
to 6,192 and 5,824. A rerun of the adapter on that tree writes zero files.

Under 47's program, core.ts has no diagnostics before or after.
performanceCore.ts has none before and the one TS2591 on its import after, as
sys.ts has on each of its four `node:` imports.

Upstream's `npm ci` and `npm run build` of the adapted tree exit 0, so tsc's
own build type-checks the binding against @types/node's perf_hooks. The
bundles read `var perfHooks = __toESM(require("perf_hooks"));` at module
scope, in _tsc.js and in typescript.js, and _tsc.js has no `typeof require`.
The built tsc runs `--extendedDiagnostics --noEmit` on a one-line file and
exits 0.

Not run: Adamic's own checker, the oracle's full baseline run, and upstream's
`hereby lint`.

## Reproduction

```sh
unit=stage3/adapt/99-host-require
export NODE_PATH=$PWD/stage3/api/node_modules
stage3/apply.sh /tmp/tsc-99h > /tmp/tsc-99h-apply.log 2>&1
node "$unit/adapt.cjs" /tmp/tsc-99h > /tmp/tsc-99h-idempotence.log 2>&1
node "$unit/witnesses.cjs" /tmp/tsc-main /tmp/tsc-99h /tmp/tsc-99h-witnesses.json > /tmp/tsc-99h-witnesses.log 2>&1
node "$unit/witnesses.cjs" --mutant /tmp/tsc-main /tmp/tsc-99h /tmp/tsc-99h-mutant.json > /tmp/tsc-99h-mutant.log 2>&1 # exits 1
```

/tmp/tsc-main is a tree applied at main without this directory. Each
witnesses run writes its probe files beside its output, in a new directory
named after it with `.probes` appended.
