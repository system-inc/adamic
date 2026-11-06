# 99-cli-perf-hooks-import: performanceCore reads perf_hooks through a static node:perf_hooks import

Task #00qn5az, under Adamic's ruling of Nov 6. lane: run as part of the wave
Correctness cuts it into; Loom's run of that wave is the lane's proof.

This is a JavaScript-changing adaptation in one file, performanceCore.ts.
tryGetPerformance reads perf_hooks through a static top-level
`import * as perfHooks from "node:perf_hooks";`, in 93's idiom, where it
called `require("perf_hooks")`. 65's
`declare const require: (specifier: "perf_hooks") => any;` goes, since
nothing else in the file reads it. tryGetPerformance declares its return type
with the file's own Performance, and PerformanceTime's timeOrigin becomes
readonly; both are types only. core.ts is untouched: isNodeLikeSystem and its
`typeof require` test are 99q's. No public API changes: tryGetPerformance is
not exported and PerformanceTime is `@internal`, so typescript.d.ts, built
with stripInternal, is unchanged.

With isNodeLikeSystem false on the oracle's ES-module Node, as 99q has it
today, nothing observable changes there: the perf_hooks binding sits under
`if (isNodeLikeSystem())` and doesn't run, before or after. Under CommonJS,
as in upstream's bundled tsc, the binding runs and reads the same object
`require("perf_hooks")` returned. What moves is when perf_hooks loads:
eagerly, when performanceCore.ts is evaluated, where it loaded lazily, on the
first call, and only on the Node path. That is timing only, given the next
section.

## node:perf_hooks loads on every ruled configuration

Adamic's rule of Nov 6 on #00qn5az: lazy to eager is an unobservable move only
for a module whose load can't fail on any configuration we've ruled. The
ruled configurations are the oracle's own Node and the rows of
internal/hostconfiguration/configurations.json, which today has one,
node-without-inspector (Node v24.19.0 configured `--without-inspector`,
#cn4yapb).

- The oracle's Node (v24.21.0 on Golem): run. `import("node:perf_hooks")`
  loads, its performance is `globalThis.performance`, and the load adds
  eleven of Node's own modules (`perf_hooks`, nine under `internal/perf/`, and
  `internal/histogram`), none of them an inspector module, and no internal
  binding.
- node-without-inspector: not run. No such Node is reachable on Golem. The
  gates find it through `ADAMIC_HOST_NODE_WITHOUT_INSPECTOR`, which Golem's
  environment doesn't set. A `--without-inspector` build of v24.19.0 was
  configured in Golem's `~/node-build-v24.19.0` on Nov 2, and its make was
  terminated before it linked a binary. The witnesses read the switch through
  the gates' own reader, internal/hostconfiguration/configuration.cjs, and
  record the row as skipped.
- So the claim for node-without-inspector rests on Node's source: v24.19.0,
  whose tarball's sha256 matches the release's SHASUMS256.txt. None of the
  eleven modules the load adds names inspector: lib/perf_hooks.js, every file
  in lib/internal/perf/, and lib/internal/histogram.js. Neither do the C++
  sources behind its `performance` binding (src/node_perf.cc, node_perf.h,
  node_perf_common.h, histogram.cc and histogram.h). node.gyp lists
  src/node_perf.cc among its unconditional sources. Its
  `v8_enable_inspector==1` condition only includes
  src/inspector/node_inspector.gypi, and otherwise defines
  `HAVE_INSPECTOR=0`. perf_hooks doesn't depend on inspector, so a Node built
  without inspector loads it as any other Node does.

## The order

No other adaptation pins performanceCore.ts's text after 65 writes it. 65
writes the cast and the declaration this adaptation removes, and runs first.
32's census and 93's and 99-wasm's scope proofs name the file, but their
adapters don't read it. 99q's adapter names it only in isNodeLikeSystem's
comment. So the order among the 99s doesn't matter, and the name sorts with
the other `99-cli-` directories.

## The rows

Line numbers are the adapted tree's before this adaptation, from the
adapter's evidence. They are never used as edit addresses.

| Row | Site | Edit |
| --- | --- | --- |
| import | performanceCore.ts:2 | `import * as perfHooks from "node:perf_hooks";` added after the `./_namespaces/ts.js` import |
| binding | performanceCore.ts:35, in tryGetPerformance's try block | `const { performance } = require("perf_hooks") as { performance?: Performance \| undefined };` becomes `const { performance } = perfHooks;` |
| declare | performanceCore.ts:104, the last statement | `declare const require: (specifier: "perf_hooks") => any;` removed, with the LF lines 65 appended, so the file ends as upstream's does |
| return type | performanceCore.ts:30, tryGetPerformance's header | `function tryGetPerformance() {` becomes `function tryGetPerformance(): { shouldWriteNativeEvents: boolean; performance: Performance; } \| undefined {` |
| readonly | performanceCore.ts:16, in PerformanceTime | `timeOrigin: number;` becomes `readonly timeOrigin: number;` |

The binding keeps upstream's destructuring. That is how
internal/oracle/testdata/node_process_performance_core.a shapes it
(`const { performance } = nativeHooks;`), and it is the one perf_hooks
namespace use Lower admits. Lower refuses the namespace as a value anywhere
else, a cast around it included, so 65's cast goes with the require. The
binding stays under `if (isNodeLikeSystem())` in its try block, and
`if (performance)` stays as upstream wrote it. The file is CRLF, and so is
everything written.

## The return type and readonly timeOrigin

The cast gave Node's perf_hooks object the file's own type. Without it,
Node's Performance leaked into tryGetPerformance's inferred return, and the
latent census (below) showed three new refusals: 53:16, the browser branch's
object seen through Node's `mark` returning `PerformanceMark`; and 74:33 and
84:29, `Performance | Performance` seen as the file's PerformanceTime and
Performance, "whose readonly field timeOrigin becomes writable". The explicit
return type restores the local view. `readonly timeOrigin` makes that view
true of Node's object, whose timeOrigin is readonly, and tsc never writes
timeOrigin: in the adapted tree's src, the only uses outside src/lib are
PerformanceTime's member, the `typeof performance.timeOrigin` test in
tryGetPerformanceHooks, and performance.ts:184's read into a local. No
assignment, compound assignment, increment, delete, defineProperty,
Reflect.set or Object.assign touches it.

Without readonly, the census reads clean only because of Lower's gap
#20nf0sm: a shorthand property in a narrowed return skips
adamic/invariant-mutable. A probe on Adamic at main refuses
`return { p: r }` from R (`readonly t`) seen as W (`t`) at the return site,
and refuses `const w: W = r;`, but admits `return { p }` under an `if`.
tryGetPerformance returns `{ shouldWriteNativeEvents: false, performance }`
under `if (performance)`. So a clean census alone wouldn't prove the return
type sound by itself; with readonly, the view is honest.

### Census

Meter's first-error latent entry census (`STAGE3_METER_LATENT=first-error`),
the entry-census binary built as twice-daily.sh builds it, from main's
compiler, rooted at src/tsc/tsc.ts (checker-clean, 81 sources, every unit
attempted). To keep it short, a scratch copy of its overlay skips files not
named in `LATENT_SELECT_FILE` in the first-error file loop. Selecting
performanceCore.ts alone, and selecting the census's file prefix through it
(tsc.ts, corePublic, core, debug, semver, performanceCore, then
performance.ts, the one other file that reads its types), give the same
performanceCore.ts record, and no finding elsewhere lands in it.

| Tree | Stops in performanceCore.ts |
| --- | --- |
| main, before this adaptation | 28:1 an ambient value declaration; 35:37 a cast the runtime can't check (scan and lowering); 72:62 NotYet, a presence test of a method whose presence the receiver's construction doesn't prove; 104:1 an ambient value declaration |
| the import alone (e1e31c24) | 29:1; 53:16, 74:33 and 84:29 as above |
| this adaptation | 29:1 an ambient value declaration (`declare const performance`); 73:62 the same presence-test NotYet |

The import moves every later line down by one, so 29:1 and 73:62 are main's
28:1 and 72:62. Nothing is new.

## The ruling

Adamic refuses an ambient value declaration, and performanceCore.ts reads
`require` through one. Adamic ruled that perf_hooks arrives the way 93 brings
in fs, path and os: a static `node:` namespace import, read through a
qualified or destructured use.

## JavaScript

Stock 6.0.3 transpiles performanceCore.ts before and after. With comments
removed, only these lines differ:

```diff
 import { isNodeLikeSystem } from "./_namespaces/ts.js";
+import * as perfHooks from "node:perf_hooks";
 function tryGetPerformance() {
     if (isNodeLikeSystem()) {
         try {
-            const { performance } = require("perf_hooks");
+            const { performance } = perfHooks;
```

With comments kept, the same two lines change. The return type and readonly
are types: stock 6.0.3 gives the planned text the same JavaScript with them
and without them, and the adapter asserts that. Upstream's esbuild bundles
write the import as `var perfHooks = __toESM(require("perf_hooks"));` at
module scope, in _tsc.js and typescript.js. On a host with no
`node:perf_hooks` module at all, the module would now fail to load, where
the require failed inside the try and was ignored. No ruled configuration is
such a host, and 93's `node:fs` import already makes the same true of sys.ts.

## The adapter

adapt.cjs parses performanceCore.ts with stock typescript@6.0.3. It finds the
imports, tryGetPerformance and the top-level `declare const require` by name,
never by line. The perf_hooks import must be the reviewed text, directly after
the one `./_namespaces/ts.js` import, or absent. The binding must be 65's
text or this adaptation's, in tryGetPerformance's try block under its
`isNodeLikeSystem()` guard. The require declaration must be 65's text as the
last statement, after its LF lines, or absent. tryGetPerformance's header,
from `function` to its body, must be upstream's text or the reviewed one with
the return type. PerformanceTime, found by name, must have one `timeOrigin`
member, upstream's text or the reviewed readonly one. Before, the file names
`require` exactly twice and never names `perfHooks`. After, it never names
`require` and names `perfHooks` exactly twice. On both sides it names
`timeOrigin` exactly twice, the member and the `typeof` test, so no other
member redeclares it writable.

The file is in exactly one of two states: every piece original, or every piece
adapted. Anything else is refused before writing. A second run writes zero
files without building a program. On an original file, the adapter
type-checks the tree with 47's compiler program, and checks by symbol that the
require declaration is read only in the binding being replaced, and that the
read resolves to the file's own declaration.

The adapter plans every edit and re-parses the planned text, which must be
CRLF throughout and exactly the adapted state. Stock 6.0.3 then transpiles the
before and planned texts, with comments kept and with comments removed. The
adapted JavaScript must equal upstream's with the import added and the binding
replaced, and equal the planned text's without the return type and readonly.
Stock 6.0.3's transpileDeclaration must give identical declarations before
and after with stripInternal. Without it, the one difference must be
`readonly` on PerformanceTime's timeOrigin. A planned program, 47's with the
planned text, must bind the `perfHooks` read to the import, resolve the
return type's Performance to the file's own interface with a readonly
timeOrigin, and leave the file's diagnostics unchanged apart from the
import's own. 47's program
has no Node types (stage3 never runs npm), so the specifier draws TS2591
`Cannot find name 'node:perf_hooks'`. That must be exactly what sys.ts's
`node:fs` import draws in the same program, with the name changed. Where
Node's types are present, the import must resolve to @types/node's
`node:perf_hooks`, which exports performance. Only then does the adapter check
that the file hasn't changed and write it.

## Witnesses and the mutant

witnesses.cjs reads the tree before this adaptation and the tree after it,
runs each probe in a fresh Node process, and records what the probe prints.
performanceCore.ts runs whole, with its `./_namespaces/ts.js` import bound to
the same tree's own isNodeLikeSystem, extracted from core.ts with its ambient
declarations. The probe reads tryGetNativePerformanceHooks() and timestamp().
The ES-module rows load the source the way oracle/node.mjs does: Node's
stripTypeScriptTypes in transform mode, format module, so the program has no
`require`. The CommonJS rows load stock 6.0.3's CommonJS output, where
`require` is defined as in upstream's bundled tsc. The src/compiler row loads
the whole of src/compiler as ES modules through `_namespaces/ts.ts`, with
tsc's `.js` specifiers resolved to their `.ts` sources as
stage3/drivers/tsc-entry/node.mjs does. The import row imports
node:perf_hooks alone. The return type and readonly strip away on both
sides; the witnesses require the after tree to carry both and the before
tree neither.

Four rows are ruled, and four of four hold on Node 24.21.0, with each row the
same on both sides:

| Row | Before and after |
| --- | --- |
| performanceCore, ES module | the browser global: shouldWriteNativeEvents true, the perf_hooks performance object, timestamp finite and monotonic |
| performanceCore, CommonJS | the perf_hooks path: shouldWriteNativeEvents false, the same object |
| src/compiler, ES module | loads; isNodeLikeSystem false; `sys` undefined; shouldWriteNativeEvents true |
| node:perf_hooks, ES module | loads; its performance is `globalThis.performance`; no inspector module added |

Two rows are recorded, not ruled:

| Row | Before | After |
| --- | --- | --- |
| node:perf_hooks, node-without-inspector | skipped: `ADAMIC_HOST_NODE_WITHOUT_INSPECTOR` unset | the same |
| performanceCore, ES module, without the require test | the browser global: the require throws inside the try and is ignored | the perf_hooks path: shouldWriteNativeEvents false, the same object |

The last row extracts isNodeLikeSystem with its `typeof require` test removed
in memory, as Correctness's change to 99q would have it. It shows what this
adaptation does once isNodeLikeSystem is true on an ES-module Node, which is
hers to rule on. It moves no ruled row.

The mutant reads perf_hooks outside the guard, in memory, in the after tree:
`if (isNodeLikeSystem() || perfHooks.performance) {`. The ES-module
performanceCore row and the src/compiler row then go red, because
tryGetPerformance takes the perf_hooks path with isNodeLikeSystem false. 2 of
4 are as ruled, and the run exits 1. Outputs go to the paths given on the
command line, never this directory, because every file in this directory
feeds apply's input key.

Refused, writing nothing (the file's hash checked after each run):

| performanceCore.ts | Adapter |
| --- | --- |
| the import added, the rest original | `half-adapted performanceCore.ts: import adapted, binding original, require declaration original, return type original, timeOrigin original` |
| the binding adapted, the rest original | `half-adapted performanceCore.ts: import original, binding adapted, require declaration original, return type original, timeOrigin original` |
| adapted, 65's declaration kept | `half-adapted performanceCore.ts: import adapted, binding adapted, require declaration original, return type adapted, timeOrigin adapted` |
| upstream's cast, as before 65 | `uninventoried perf_hooks binding: ...` |
| original, with `require` read in another function | `performanceCore.ts names require outside the reviewed sites` |
| original, with `const perfHooks = 0;` | `performanceCore.ts names perfHooks outside the reviewed sites` |
| adapted, the binding moved out of its try block | `the perf_hooks binding left its try block` |
| adapted, the import a named import | `unreviewed perf_hooks import` |
| the import, binding and declaration adapted, as e1e31c24 wrote it | `half-adapted performanceCore.ts: import adapted, binding adapted, require declaration adapted, return type original, timeOrigin original` |
| adapted, timeOrigin writable (the return type alone) | `half-adapted performanceCore.ts: import adapted, binding adapted, require declaration adapted, return type adapted, timeOrigin original` |
| adapted, without the return type (readonly alone) | `half-adapted performanceCore.ts: import adapted, binding adapted, require declaration adapted, return type original, timeOrigin adapted` |
| original, with the return type | `half-adapted performanceCore.ts: import original, binding original, require declaration original, return type adapted, timeOrigin original` |
| original, with readonly timeOrigin | `half-adapted performanceCore.ts: import original, binding original, require declaration original, return type original, timeOrigin adapted` |
| adapted, the return type's performance a PerformanceTime | `uninventoried tryGetPerformance header: ...` |
| adapted, `timeOrigin: number;` added to Performance | `performanceCore.ts names timeOrigin outside the reviewed sites` |

## Measured

The whole chain, stage3/apply.sh at main with this directory, exits 0, with
the row `99-cli-perf-hooks-import | 1 | 4 | 5`. A rerun of the adapter on that
tree writes zero files. The witnesses hold four of four as ruled, and the
mutant goes red at two of four and exits 1. Under 47's program,
performanceCore.ts has no diagnostics before and the one TS2591 on its import
after, as sys.ts has on each of its four `node:` imports. Adamic's own checker,
which reads @types/node, finds the tsc entry clean with this adaptation
applied (0 diagnostics, the census's header), so Node's Performance is
assignable to the file's own. The refused states above were each run against
the adapter, and each left the file's hash unchanged.

Before the return type and readonly were added, upstream's `npm ci` and
`npm run build` of the adapted tree exited 0. They were not rerun after.

Not run: the oracle's full baseline run, upstream's `hereby lint`, upstream's
build after the return type and readonly, the unfiltered latent census of all
81 sources, and a Node built without inspector.

## Reproduction

```sh
unit=stage3/adapt/99-cli-perf-hooks-import
export NODE_PATH=$PWD/stage3/api/node_modules
stage3/apply.sh /tmp/tsc-99p > /tmp/tsc-99p-apply.log 2>&1
node "$unit/adapt.cjs" /tmp/tsc-99p > /tmp/tsc-99p-idempotence.log 2>&1
node "$unit/witnesses.cjs" /tmp/tsc-main /tmp/tsc-99p /tmp/tsc-99p-witnesses.json > /tmp/tsc-99p-witnesses.log 2>&1
node "$unit/witnesses.cjs" --mutant /tmp/tsc-main /tmp/tsc-99p /tmp/tsc-99p-mutant.json > /tmp/tsc-99p-mutant.log 2>&1 # exits 1
```

The census, from a compiler checkout at main, on Golem. The selecting overlay
adds three lines at the top of the first-error loop in the generated
internal/lower latent hook, `for _, file := range program.Files() {`: skip
the file unless its name is one of `LATENT_SELECT_FILE`'s newline-separated
paths.

```sh
source ~/.adamic-tools/env.sh
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/latent-overlay > /tmp/latent-overlay.log 2>&1
gofmt -w /tmp/latent-overlay/*.go
python3 stage3/meter/entry_overlay.py /tmp/latent-overlay /tmp/entry-overlay > /tmp/entry-overlay.log 2>&1
# /tmp/select-overlay: /tmp/entry-overlay with the three-line skip in its lower latent hook
go build -buildvcs=false -overlay=/tmp/select-overlay/overlay.json -o /tmp/select-census ./stage3/census/latent/tool > /tmp/select-build.log 2>&1
LATENT_SELECT_FILE=/tmp/tsc-99p/src/compiler/performanceCore.ts LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 \
  /tmp/select-census /tmp/tsc-99p/src/tsc/tsc.ts /tmp/tsc-99p-latent.jsonl > /tmp/tsc-99p-latent.log 2>&1
```

For the prefix run, `LATENT_SELECT_FILE` names src/tsc/tsc.ts and
src/compiler's corePublic.ts, core.ts, debug.ts, semver.ts, performanceCore.ts
and performance.ts. Run both on /tmp/tsc-main too.

/tmp/tsc-main is a tree applied at main without this directory. Each
witnesses run writes its probe files beside its output, in a new directory
named after it with `.probes` appended. Set
`ADAMIC_HOST_NODE_WITHOUT_INSPECTOR` to a Node built `--without-inspector`
to rule that row too.
