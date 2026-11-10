> Ready to land: zero class (c), unchanged runtime export keys, green amended
> full apply, and all 398 final lane shards passing their guards.

## Follow-up: complete bundle hunk classification

The revised gate permits mechanical bundler fallout. Every outside-owner hunk
is classified in debug-display-bundle-classification.json.gz. The complete normal
before/after emitted patch, including the five owner edits, is retained in
debug-display-all-emitted-javascript-hunks.patch.gz.

| Bundle | (a) pure rename | (b) removed declaration | (c) stop |
| --- | ---: | ---: | ---: |
| _tsc.js | 772 | 2 | 0 |
| typescript.js | 1185 | 0 | 0 |
| run.js | 1247 | 0 | 0 |
| Total | 3204 | 2 | 0 |

The analyzer first replaces exactly the five owner texts with their after
texts, then inventories every residual hunk with three context lines. For (a),
only terminal decimal suffixes on AST Identifier nodes are normalized. Strings,
comments, numeric literals, punctuation and whitespace are untouched. Every
class-(a) hunk compares byte-identically after that transformation; both
normalized hashes are retained for each hunk. A mixed rename/deletion hunk
receives only one class, not two.

Only _tsc.js drops declarations. The reference list below is exhaustive and
resolved by TypeScript symbol identity in the original unadapted bundle;
declaration names themselves are excluded from the reference count.

| Removed function | Every original reference | Enclosing function | Inside a reviewed owner? |
| --- | --- | --- | --- |
| getJSDocModifierFlagsNoCache | _tsc.js:17073:51 | getEffectiveModifierFlagsNoCache | No |
| getEffectiveModifierFlagsNoCache | _tsc.js:1791:42 | enableDebugInfo | Yes |
| isArrayTypeNode | _tsc.js:1775:1231 | enableDebugInfo | Yes |

Both deletion hunks are (b), using the transitive closure authorized in the
follow-up. A removed declaration is class (b) when every reference to it is
inside a reviewed owner or inside another removed declaration that is itself
class (b). The classifier computes the least fixed point and records reference
edges and full chains; unanchored cycles fail closed.

Hunk 55 records getJSDocModifierFlagsNoCache ->
getEffectiveModifierFlagsNoCache -> enableDebugInfo. The direct reference at
17073 is inside the removed modifier helper, whose reference at 1791 is inside
enableDebugInfo. isArrayTypeNode's sole reference is directly in enableDebugInfo.
All three bundles compare byte-identically after identifier normalization and
these three declaration deletions. There are zero class-(c) hunks.

The runtime export comparison passes on Node v24.19.0. Loading each real
normal typescript.js through Node import and require gives byte-identical
sorted Object.keys arrays: 2258 module keys and 2256 keys on the separate ts
object/CommonJS exports. No key is added or removed. The complete arrays and
hashes are in debug-display-runtime-export-keys.json.gz; the executable check is
debug-display-export-keys.cjs.

Current result: ready to land. Classification and exports pass; the amended
full apply exits 0, all 746 applied source files match the classified normal
candidate, final replay makes zero edits, and all 398 final lane shards pass
their ordinary guards (106366 passes, one sanctioned API failure, zero skips).
The entire formatControlFlowGraph owner, including FlowGraphNode at
debug.ts:1021:29, is byte-identical to the baseline. The source adaptation and
pins are unchanged by the transitive-rule follow-up.

# Debugger-display family amendment to 99t

Adamic, task b5w3ycg, Nov 10 04:23Z. JavaScript-changing.
Measured baseline main: 7910ab94a42b0c7f62a79329d7f250b1b2045d05.
Landing base: 7f2c852df. The full apply validated 95afe50a6; the entire
stage3 tree is byte-identical between these main tips. The final main update
changes only the native runtime and its validation.
TypeScript 6.0.3 source: 050880ce59e30b356b686bd3144efe24f875ebc8.

## Ownership and behavior

The pin audit preceded edits. 99t actively pins the whole Debug.enableDebugInfo
owner, so this is an amendment here, with no new adaptation directory.
99n's attachFlowNodeDebugInfoWorker pin is proposed only; its active pin is
formatControlFlowGraph.hasNode. 40's enum-display edit concerns formatEnum;
its historical whole-file hashes do not gate these owners. 30's whole-files
helper checks the assertion-cache declaration rather than debug.ts's bytes.
The remaining debug assertion and reflection adapters cover other owners.
See debug-display-pin-audit.json.

The complete upstream debug.ts region 505 through 810 was read. Five owners
lose their debugger-display installation bodies:

| Owner | Result |
| --- | --- |
| attachFlowNodeDebugInfo | Returns flowNode unchanged |
| attachNodeArrayDebugInfo | Returns without touching array |
| attachFlowNodeDebugInfoWorker | Empty body; no property installation |
| attachNodeArrayDebugInfoWorker | Empty body; no property installation |
| enableDebugInfo | Keeps its early guard and installs nothing |

isDebugInfoEnabled retains its false initializer and is never set true. Its
remaining early-return reader is truthful: nothing has been installed.
The two attachment readers disappear with their installation bodies.
The prototype variable declarations, imports, signatures and surrounding
upstream comments remain to keep the edit inside its five owner boundaries.
Blank lines preserve the before-owner line counts and the restoring mutant's
525:37 coordinate. The worker declarations remain, with empty bodies, to keep
strict before/after owner recognition. No Object.create, Object.setPrototypeOf
or Object.defineProperties operation remains in these five bodies.

Other formatting helpers, formatControlFlowGraph, DebugTypeMapper and
attachDebugPrototypeIfDebug remain unchanged. They are outside this ruled
installation family. 99t's SourceFile check CW143 was inside enableDebugInfo
and is withdrawn with that body. Its other three owners and their checks
remain. The retained formatter witness supplies an explicit formatter fixture
for its type object; it no longer relies on automatic installation of getters.
The original 99t measurements below the amendment notice are historical.

Upstream runs the installation at tsc startup only when Debug.isDebugging is
set by sys.debugMode: NODE_INSPECTOR_IPC, VSCODE_INSPECTOR_OPTIONS, a recognized
--inspect/--debug execArgv, or process.recordreplay. Flow callers are upstream
binder.ts:496 and 582; array callers are nodeFactory.ts:1179, 1191 and 1205.
The installed members serve a JavaScript debugger and the worker's own
"__tsDebuggerDisplay" in array guard. They do not supply CLI output.
Under a JavaScript debugger, Node tsc installs displays on flow nodes and
NodeArrays, and getters on symbols, types, signatures and node constructors;
the adapted native tsc does not. A native binary has no JavaScript debugger.

## Measured gates

The complete lane oracle ran through all 398 ordinary shards: 106366 passes,
one sanctioned API failure, zero skips, and every ordinary shard guard passed.
The guarded-body counts, separately, are 0, 0 and 0 for flow attachment, array
attachment and enableDebugInfo. The debug-mode small-program control exits 0
and fires them 6801, 45429 and 1 times. Its ordinary-mode twin exits 0 with
three zeros. Instrumentation touched a private emitted-JavaScript view only.
The omitted Git metadata API attempt and the initial control missing
ignoreConfig are retained as preliminary failures, with passing corrections.
The full counter controller's exit was 1 for that first API attempt; the
398-unit reconciliation exits 0 and retains the corrected accepted inventory.

The Node witness runs the unmodified baseline CLI on 200 sorted single-file
compiler inputs, once with VSCODE_INSPECTOR_OPTIONS={} and once without it.
It compares raw stdout bytes, stderr bytes and exit directly. All 200 pairs
are identical. Separate instrumented replays are identical too and fire the
three bodies 1312455, 9088154 and 200 times. The fixed CLI flags are
ignoreConfig, pretty false, noEmit and skipLibCheck. Case identities are
recorded as input hashes. An interrupted controller left 64 completed pairs;
those exact pairs were verified and reused, and all 200 counted replays were
completed by the foreground resume controller, which exits 0.

adapt.cjs pins all eight whole owners of the amended 99t transformation.
It accepts only the complete before state or complete after state, refuses
mixed or changed owners before writes, preserves CRLF, bounds every replacement
inside its non-overlapping owner, and makes zero edits on replay. Its stock
6.0.3 emission proof uses verbatim imports to isolate the body boundaries,
with comments kept and removed. Declaration emission is equal with and without
stripInternal. The normal emitted bundles are checked separately.

See debug-display-PROOF.md and the adjacent JSON records for the remaining
census, normal JavaScript comparison, contract and full-apply evidence.

The five largest evidence records (bundle classification, runtime export keys, witness, final lane and count) are stored gzipped as written by their scripts, `gzip -n -9`, to keep the change small. Read each with `gzip -dc`.
