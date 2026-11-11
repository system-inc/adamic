# 99-cli-tracing-process-import: tracing reads process through a static node:process import

This is a JavaScript-changing adaptation in one file, tracing.ts. startTracing
names its trace files with `process.pid`, which it read through 65's
`declare const process: { pid: number };`, an ambient value declaration Adamic
refuses. Now process arrives through a static top-level
`import process from "node:process";`, added after the file's last import,
and 65's process declaration goes. The two `process.pid` reads in countPart
are unchanged; they bind to the import.

65's `declare const require: (specifier: "fs") => any;` stays as it is.
tracing's `require("fs")` is moving to a node:fs import in its own change,
which owns that line. The adapter accepts tracing.ts either way: with the
require line standing directly before the process declaration, as 65 leaves
it, or with it already gone.

No public API changes: tracingEnabled is `@internal`, and neither the
declaration nor the import reaches a declaration, so typescript.d.ts, built
with stripInternal, is unchanged, and so are the declarations without it.

## The import form

Two forms were candidates, `import process from "node:process";` and
`import * as process from "node:process";`. The default import is the one the
checker, Lower and Node agree on.

The checker. Adamic checks every program under internal/load/load.go's
compilerOptions: module ESNext, moduleResolution Bundler, verbatimModuleSyntax,
no esModuleInterop or allowSyntheticDefaultImports set. Stock 6.0.3 reads
both as true under those options (getESModuleInterop and
getAllowSyntheticDefaultImports), and the pinned tsgo checker has
esModuleInterop always on. @types/node 25.3.3, the declarations Adamic bundles
(load.NodeTypesVersion) and stage3/api pins, declares
`declare module "node:process"` with `export = process`, the global process.
In tsgo's getTargetOfModuleDefault, canHaveSyntheticDefault is true for an
ambient module with no syntactic default and no `__esModule`, so the default
import resolves to the export= value, the global `process` variable. Stock
6.0.3 under Adamic's options with @types/node gives both forms no diagnostic,
types `process.pid` as number, and resolves both aliases to that global
variable (declared at @types/node globals.d.ts:3 and in process.d.ts's
`global` block). The witnesses' checker row holds the same of tracing.ts
itself, in the whole tree with @types/node: no diagnostics before or after,
and both reads bound to the import, whose target is the global process.

Lower. Its symbol helper (internal/lower/locals.go:178) follows the alias to
that variable. processPath (internal/lower/process.go:16) admits a symbol
named process declared in the prelude or a Node library file, so the receiver
is `process`, and `process.pid` lowers to the pid ProcessCall
(internal/lower/library_node_process.go:346). nodeProcessPath's first branch,
a symbol whose declaration is the `node:process` module itself, is not the one
taken: that is how a namespace import of a module without export=, such as
node:perf_hooks or node:os, resolves. Because node:process is export=, the
namespace import resolves to the same global variable (resolveESModuleSymbol
returns the export= target: Process has no call or construct signature and no
default member). So Lower's code alone would read the two forms alike; the
repository's own fixtures mostly write the default form, and
internal/hostconfiguration/testdata/configuration.a writes it exactly.

Node. Under Node 24 as an ES module, the default import is
`globalThis.process` itself. The namespace import is a module namespace
object, and its `pid` is a copy Node took when it built the builtin's facade.
Both give the same pid while nothing redefines it, but only the default import
is the live object upstream's global read was. The witnesses show it: with
`process.pid` redefined before startTracing, the before tree and this
adaptation name the trace file with the new pid, and the namespace mutant
names it with the old one.

The three agree on the default import, so that is the form.

## What moves at load

Under the oracle's ES-module load, tracing.ts's import is the first
node:process import in tsc's sources; sys.ts imports node:fs, node:path,
node:os and node:crypto, and performanceCore.ts node:perf_hooks. Node builds a
builtin's ES module facade from the object's own properties, so linking
node:process runs process's getters once: stdout, stdin, stderr, report and
the others. After those other imports, it adds `process`,
`internal/process/finalization`, `net`, `internal/stream_base_commons`,
`internal/process/report`, and the cares_wrap, stream_wrap, tcp_wrap,
pipe_wrap and report bindings to Node's module list. The default and
namespace forms do the same. node:process can't fail to load on any Node, and
the stdin it makes holds nothing open: with stdin a pipe a shell keeps open,
a probe that imports node:process exits within a second. Under CommonJS, as
in upstream's bundled tsc, stock 6.0.3 writes the import as
`__importDefault(require("node:process"))`, which wraps process without
reading its properties. That is the observable difference, and it is Adamic's to
rule on; the rows record it.

## Native and wasm32

Native lowers process.pid: internal/lower/library_node_process.go lowers it
to a pid ProcessCall. wasm32-wasi refuses it by name at the build boundary
(internal/native/host_target.go: "wasm32-wasi refuses process.pid"), so this
moves wasm32's stop from lowering to a named target refusal and is not a
clear there.

## The order

65 writes the declaration this adaptation removes, and runs first. 47 adds
tracing.ts's `./hostErrors.js` import, which this import follows. 62, 63 and
99-correctness-tracing-assert edit dumpTypes, dumpLegend and stopTracing,
and none reads the imports, countPart or the end of the file. The adapter
needs sys.ts's `node:fs` import, which 93 adds, for its diagnostic comparison.
The node:fs change for the require line will remove the line directly before
this one; either order works, since this adapter accepts the file with that
line gone, and then gives back upstream's own ending. One thing for that
change to know: 65's adapter appends both of its tracing lines whenever the
file has no `declare const require:`, so once the require line is gone, a
rerun of 65's check on the final tree would see tracing.ts change.

## The rows

Line numbers are the adapted tree's before this adaptation, from the
adapter's evidence. They are never used as edit addresses.

| Row | Site | Edit |
| --- | --- | --- |
| import | tracing.ts:26 | `import process from "node:process";` added after the `./hostErrors.js` import |
| declaration | tracing.ts:379, the last statement | `declare const process: { pid: number };` removed, with the LF line 65 appended for it |
| reads | tracing.ts:83 and 84, startTracing's countPart | unchanged; `process.pid` now reads the import |

The file is CRLF, and so is the import. 65's lines are LF; removing the
declaration from the end of the require line to the end of the file leaves
the file ending as 65 wrote it without the process line.

## JavaScript

Stock 6.0.3 transpiles tracing.ts before and after. With comments removed,
and with comments kept, only this line differs:

```diff
 import { errorMessage } from "./hostErrors.js";
+import process from "node:process";
```

65's declaration is types only, so its removal emits nothing. Stock 6.0.3's
CommonJS output reads `node_process_1.default.pid`, where
`node_process_1 = __importDefault(require("node:process"))`, which is
process itself.

## The adapter

adapt.cjs parses tracing.ts with stock typescript@6.0.3. It finds the
imports, startTracing's countPart and the top-level process declaration by
name, never by line. The first three imports must be the `./_namespaces/ts.js`
import, upstream's `./_namespaces/ts.performance.js` namespace import and
47's `./hostErrors.js` import, in that order, as the first three statements.
The node:process import must be the reviewed text directly after them, or
absent. The process declaration must be 65's text as the last statement,
followed by its LF, and preceded either by 65's require line and an LF, or,
with no require declaration left, by the file's own CRLF and 65's blank LF
line; or absent. countPart must be upstream's text in both states. Both
states name `process` exactly three times: the declaration or the import
binding, and the two reads, each the receiver of `.pid` inside countPart.

The file is in exactly one of two states: every piece original, or every piece
adapted. Anything else is refused before writing. A second run writes zero
files without building a program. On an original file, the adapter
type-checks the tree with 47's compiler program and checks by symbol that the
two reads resolve to the file's own declaration.

The adapter plans both edits and re-parses the planned text, which must keep
the file's line breaks, keep any require declaration, and be exactly the
adapted state. Stock 6.0.3 then transpiles the before and planned texts, with
comments kept and removed: the adapted JavaScript must equal upstream's with
the import added after the `./hostErrors.js` import, built from the before
output. transpileDeclaration must give identical declarations before and
after, with stripInternal and without it. A planned program, 47's with the
planned text, must bind both reads to the import, and leave the file's
diagnostics unchanged apart from the import's own. 47's program has no Node
types (stage3 never runs npm), so the specifier draws TS2591
`Cannot find name 'node:process'`, which must be exactly what sys.ts's
`node:fs` import draws in the same program, with the name changed. Where Node's
types are present, the import must resolve to @types/node's `node:process`,
its target must be the global process, from @types/node, with a number pid.
Only then does the adapter check that the file hasn't changed and write it.

## Witnesses and the mutants

witnesses.cjs reads the tree before this adaptation and the tree after it.
Each runtime probe loads both sides into one fresh Node process, so they read
one pid and their trace file names compare byte for byte. The ES-module rows
load the whole of src/compiler through `_namespaces/ts.ts` the way
oracle/node.mjs loads a source (stripTypeScriptTypes in transform mode, format
module), with tsc's `.js` specifiers resolved to their `.ts` sources as
stage3/drivers/tsc-entry/node.mjs does, and call
`tracingEnabled.startTracing(mode, directory)`, then list the directory. As
the oracle loads it there is no require, and startTracing stops at
`require("fs")` before it reads process.pid, the same on both sides; that row
is recorded. The ruled ES-module rows give the probe a require after the
compiler has loaded, standing in for the node:fs import. The CommonJS rows
load stock 6.0.3's CommonJS output of tracing.ts alone, with its imports bound
to stand-ins for Debug, combinePaths, timestamp and errorMessage. The checker
row builds 47's program with @types/node 25.3.3 from stage3/api.

Eight rows are ruled, and eight of eight hold on Node v24.14.1:

| Row | Before and after |
| --- | --- |
| tracing, ES module, build | `trace.<pid>-1.json`, the same bytes on both sides |
| tracing, ES module, server | `trace.<pid>.json`, the same |
| tracing, ES module, project | `trace.json`, the same |
| tracing, ES module, server, process.pid redefined first | `trace.2147480001.json`, the same |
| tracing, CommonJS, build | `trace.<pid>-1.json`, the same |
| tracing, CommonJS, server | `trace.<pid>.json`, the same |
| node:process, ES module | the default import is `globalThis.process`, with its pid |
| checker with @types/node 25.3.3 | tracing.ts has no diagnostics on either side; before, the reads bind to the file's own declaration; after, to the import, whose target is the global process |

Three rows are recorded, not ruled: the ES-module load as the oracle has it
(both sides throw `tracing requires having fs`), the modules and bindings
node:process adds after sys.ts's imports (above), and the stdin probe (exits
within a second).

`--mutant namespace` writes the after tree's import as
`import * as process from "node:process";` in memory. The redefined-pid row
goes red, the adapted side naming `trace.<real pid>.json`, and so does the
checker row (the reads bind to a namespace import): six of eight, exit 1.
`--mutant undo` undoes the change in memory: the import goes and 65's
declaration comes back. The runtime rows still hold, since the JavaScript is
the before tree's, and the checker row goes red: the reads bind to the file's
own declaration again. Seven of eight, exit 1. Outputs go to the paths given
on the command line, never this directory, because every file in this
directory feeds apply's input key.

Refused, writing nothing (the file's hash checked after each run):

| tracing.ts | Adapter |
| --- | --- |
| the import added, the declaration kept | `half-adapted tracing.ts: import adapted, process declaration original` |
| the declaration removed, no import | `half-adapted tracing.ts: import original, process declaration adapted` |
| adapted, the import a namespace import | `unreviewed node:process import` |
| adapted, the import `import { pid } from "node:process";` | `unreviewed node:process import` |
| original, with `void process.pid;` in startTracing's body | `tracing.ts names process outside the reviewed sites` |
| original, countPart reading `process.ppid` | `unreviewed countPart` |
| original, a statement after the declaration | `tracing.ts has a line break that is not CRLF` |
| original, the declaration's line break CRLF | `tracing.ts has a line break that is not CRLF` |

Accepted: the original with 65's require line already gone. The adapter
writes the import, removes the declaration, and the file ends `}\r\n`, as
upstream's does.

## Measured

On a Mac, Node v24.14.1, stage3/api installed with `npm ci` (its digest
matches node_modules.sha256). A tree applied with stage3/apply.py
`--adapt-checkout` at main, without this directory, then this adapter on a
copy: one file written, the evidence above; a rerun writes zero files and
leaves tracing.ts's hash unchanged. The whole chain with this directory,
`--adapt-checkout` on a fresh checkout of the pinned commit, exits 0 with the
row `99-cli-tracing-process-import | 1 | 1 | 1`; its src differs from main's
applied tree only in tracing.ts, which is byte for byte the adapter's output
on the copy, and a rerun of the adapter on it writes zero files. Under 47's program, tracing.ts has no
diagnostics before and the one TS2591 on its import after, as sys.ts has on
its `node:` imports. The witnesses hold eight of eight, each mutant goes red
and exits 1, and each refused state left the file's hash unchanged.

65's check.cjs no longer passes on main's full tree: it fails on
performanceCore.ts (`source site drift`, since
99-cli-perf-hooks-import removed the cast it pins), before and after this
adaptation alike. Run on core.ts and performanceCore.ts as 65 leaves them
with tracing.ts from the tree after this adaptation, it passes
(`idempotent`, `runtime_change_mutant_caught`) and leaves tracing.ts's hash
unchanged, as it does with main's tracing.ts.

Not run: Go (the Adamic census, the meter and Lower's tests), the oracle's
full baseline run, upstream's build and `hereby lint`, and a Node built
without inspector.

## Reproduction

```sh
unit=stage3/adapt/99-cli-tracing-process-import
export NODE_PATH=$PWD/stage3/api/node_modules
export CENSUS_TYPESCRIPT=$NODE_PATH/typescript
# /tmp/tsc-main: a checkout of the pinned commit adapted at main without this directory
python3 stage3/apply.py --adapt-checkout /tmp/tsc-main > /tmp/tsc-main-apply.log 2>&1
cp -R /tmp/tsc-main /tmp/tsc-99t
node "$unit/adapt.cjs" /tmp/tsc-99t > /tmp/tsc-99t-adapt.log 2>&1
node "$unit/adapt.cjs" /tmp/tsc-99t > /tmp/tsc-99t-idempotence.log 2>&1
node "$unit/witnesses.cjs" /tmp/tsc-main /tmp/tsc-99t /tmp/tsc-99t-witnesses.json > /tmp/tsc-99t-witnesses.log 2>&1
node "$unit/witnesses.cjs" --mutant namespace /tmp/tsc-main /tmp/tsc-99t /tmp/tsc-99t-namespace.json > /tmp/tsc-99t-namespace.log 2>&1 # exits 1
node "$unit/witnesses.cjs" --mutant undo /tmp/tsc-main /tmp/tsc-99t /tmp/tsc-99t-undo.json > /tmp/tsc-99t-undo.log 2>&1 # exits 1
```

Each witnesses run writes its probe files beside its output, in a new
directory named after it with `.probes` appended.
