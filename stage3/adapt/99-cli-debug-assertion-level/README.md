# 99-cli-debug-assertion-level: debug.ts checks the assertion level on each call

Tasks #0kpp2zc and #b1y988t, under Adamic's ruling of Nov 5. lane: run as part
of the wave Correctness cuts it into; Loom's run of that wave is the lane's
proof.

This is a JavaScript-changing adaptation in one file, debug.ts.
shouldAssertFunction no longer caches an assertion and replaces it with
`noop` below its level. It returns `shouldAssert(level)` on every call.
setAssertionLevel no longer writes cached assertions back. It stores the
level and nothing else. The cache declaration goes, and so do the imports of
`noop` and `getOwnKeys`, which only the removed lines named. No public API
changes: Debug is `@internal`, shouldAssertFunction and the cache are not
exported, and setAssertionLevel keeps its signature.

Input: microsoft/TypeScript v6.0.3, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`. Adamic base: main `fd30fac4`.

## The rows

Line numbers are the adapted tree's before this adaptation, from the
adapter's evidence. They are never used as edit addresses.

| Row | Site | Edit |
| --- | --- | --- |
| import | debug.ts:16 and 71, the `./_namespaces/ts.js` named import | `getOwnKeys,` and `noop,` removed |
| cache | debug.ts:155, `const assertionCache` | removed, with the blank line above it |
| set | debug.ts:161, setAssertionLevel | the body becomes `currentAssertionLevel = level;` |
| doc | debug.ts:181, shouldAssertFunction's JSDoc | reworded: it no longer caches or replaces anything |
| should | debug.ts:187, shouldAssertFunction | the body becomes `return shouldAssert(level);` |

The cache declaration is upstream's with adaptation 30's `| undefined` on the
entry type. 76 turned AssertionKeys into its six literal names. This
adaptation leaves both AssertionKeys and AnyFunction alone. shouldAssertFunction
keeps `<K extends AssertionKeys>(level: AssertionLevel, name: K): boolean`.
debug.ts is CRLF, and so is everything written.

## The ruling

Adamic refuses the two writes, which use the namespace object as a value and
write through `any`: `(Debug as any)[name] = noop` in shouldAssertFunction
and `(Debug as any)[key] = cachedFunc` in setAssertionLevel. Adamic ruled
adapt, don't lower, and accepted exactly this edit. The level check moves from
a one-time patch of the namespace to a comparison on each call. That costs
each disabled assertion one call and one comparison, where upstream paid them
once. The caller's arguments are evaluated either way.

## The upstream bug

setAssertionLevel's restore loop writes `cachedFunc`, the cache entry
`{ level, assertion }`, back into `Debug[key]` instead of
`cachedFunc.assertion` (debug.ts:170 upstream). So an assertion that ran at
None (cached, replaced with noop) and is called again after the level rises
to Normal throws `TypeError: Debug.assertNode is not a function`. tsc itself
sets the level once, from `NODE_ENV`, before any assertion runs (sys.ts:1973
upstream), so the CLI never reaches the restore. A consumer that raises the
level through `ts.Debug.setAssertionLevel` does, and this adaptation removes
the loop. LEDGER.md entry 8 has the counterexample and the drafted issue.

## The adapter

adapt.cjs parses debug.ts with stock typescript@6.0.3. It finds namespace
Debug, and in it setAssertionLevel, shouldAssert and shouldAssertFunction by
name, never by line. Each function's head must be the reviewed signature,
and each body must be exactly upstream's text or exactly 99's text.
shouldAssert must still be `return currentAssertionLevel >= level;`, which
the adapted shouldAssertFunction relies on. shouldAssertFunction's one leading
comment must be upstream's JSDoc or 99's. The cache must be upstream's
declaration or absent. `noop` and `getOwnKeys` must both be value specifiers
of the one `./_namespaces/ts.js` named import, or both gone.

shouldAssertFunction must be called exactly six times, once inside each of
the six assertion functions, as `shouldAssertFunction(AssertionLevel.Normal,
"<its own name>")`. `Debug as any` must appear exactly twice in the namespace
before (the two writes) and never after. After, the file names none of
`assertionCache`, `prevAssertionLevel`, `cachedFunc`, `noop` or `getOwnKeys`
as an identifier.

The whole file is in exactly one of two states: every piece upstream's, or
every piece 99's. Anything else is refused before writing. A second run
writes zero files without building a program. On an upstream file, the
adapter type-checks the tree with 47's compiler program. By symbol identity,
it then checks that the cache, `noop` and `getOwnKeys` are read only inside
the two bodies being replaced, so removing the declaration and the two
imports leaves nothing unbound.

The adapter plans every edit and re-parses the planned text, which must be
CRLF throughout and exactly the adapted state. Stock 6.0.3 then transpiles
the before and planned texts, with comments kept and with comments removed.
The adapted JavaScript must equal upstream's with the three pieces replaced
and the two names dropped from the import. Stock 6.0.3's transpileDeclaration
must give identical declarations before and after, with stripInternal (as
typescript.d.ts is built, where both are `export {};`) and without it. Only
then does the adapter check that the file hasn't changed and write it.

## JavaScript

Stock 6.0.3 transpiles debug.ts before and after. With comments removed, only
these lines differ (the import line is cut here, and loses `noop` further
along too):

```diff
-import { AssertionLevel, ..., getEmitFlags, getOwnKeys, getParseTreeNode, ...
+import { AssertionLevel, ..., getEmitFlags, getParseTreeNode, ...
-    const assertionCache = {};
     function setAssertionLevel(level) {
-        const prevAssertionLevel = currentAssertionLevel;
         currentAssertionLevel = level;
-        if (level > prevAssertionLevel) {
-            for (const key of getOwnKeys(assertionCache)) {
-                const cachedFunc = assertionCache[key];
-                if (cachedFunc !== undefined && Debug[key] !== cachedFunc.assertion && level >= cachedFunc.level) {
-                    Debug[key] = cachedFunc;
-                    assertionCache[key] = undefined;
-                }
-            }
-        }
     }
     function shouldAssertFunction(level, name) {
-        if (!shouldAssert(level)) {
-            assertionCache[name] = { level, assertion: Debug[name] };
-            Debug[name] = noop;
-            return false;
-        }
-        return true;
+        return shouldAssert(level);
     }
```

With comments kept, the restore comment goes too and the JSDoc is reworded.

## Witnesses and the mutant

witnesses.cjs extracts debug.ts's own statements (LogLevel, LoggingHost and
namespace Debug) whole from the tree before 99 and the tree after it. Stock
6.0.3 transpiles each side, and the imports are bound to stock 6.0.3's own
values. Each case starts a fresh namespace and sets the level with sys.ts's
own `Debug.setAssertionLevel(...)` call, extracted from the same tree, over a
sys whose getEnvironmentVariable reads `environment[name] || ""` as Node's
does. Then it calls an assertion from a tsc-style call site,
`Debug.assertNode(...)`, and records what the call returns or throws (type
and exact message), and whether the namespace member is still the assertion.

Each of the six assertion functions runs five rows. 30 of 30 are as ruled on
Node 24.14.1, and the six assertions give the same answer (assertNode shown):

| Row | Before | After |
| --- | --- | --- |
| None (`NODE_ENV` unset), failing | returns; Debug.assertNode is now noop | returns; Debug.assertNode is unchanged |
| Normal (`NODE_ENV=development`), passing | returns | returns |
| Normal, failing | `Debug Failure. False expression: Unexpected node.` with the verbose line | the same message |
| None, one failing call, then `setAssertionLevel(Normal)`, failing | `TypeError: Debug.assertNode is not a function`; the member is the object `{ level, assertion }` | the same Debug Failure as at Normal |
| the same, passing | the same TypeError | returns |

The verbose line for assertNode reads
`Verbose Debug Information: Node Identifier did not pass test 'isStringLiteral'.`
The other five assertions have their own lines, identical on both sides.

18 rows move, three per assertion. The None row moves only because the
namespace member is no longer replaced. The two raised rows move from the
TypeError to the assertion running and asserting. The other twelve don't
move.

The mutant puts the noop patch back, in memory, in the after tree's
shouldAssertFunction: `if (!shouldAssert(level)) { (Debug as any)[name] = noop; return false; } return true;`.
The raised failing row then goes red for all six assertions. The original
throws the TypeError, and the mutant returns without asserting, because
nothing restores the assertion. 24 of 30 are as ruled, and the run exits 1.
Outputs go to the paths given on the command line, never this directory,
because every file in this directory feeds apply's input key.

Refused, writing nothing (debug.ts's hash checked after each run):

| Tree | Adapter |
| --- | --- |
| setAssertionLevel adapted, the rest upstream | `half-adapted debug.ts: setAssertionLevel adapted, shouldAssertFunction original, doc comment original, assertionCache original, imports original` |
| adapted, the `noop` import kept | `half-adapted import: getOwnKeys adapted, noop original` |
| adapted, both imports kept | `half-adapted debug.ts: ... imports original` |
| adapted, upstream's JSDoc kept | `half-adapted debug.ts: ... doc comment original ...` |
| upstream with line 170 fixed to `cachedFunc.assertion` | `uninventoried setAssertionLevel body: ...` |
| upstream with a seventh shouldAssertFunction call, in Debug.type | `type: the assertion key is not its own name` |
| upstream with `noop` read in getFunctionName | `noop is read outside the replaced bodies, at line 369` |
| upstream with `getOwnKeys` read in formatEnum | `getOwnKeys is read outside the replaced bodies, at line 391` |

## Measured

debug.ts's own diagnostics, with stock 6.0.3 over src/compiler, before and
after: three under the tree's own tsconfig and three under 47's strict
options. They are the same three rows both times (TS2339 `captureStackTrace`
twice and TS2584 `console`), 21 lines higher after.

Not checked: upstream's `hereby lint` (eslint, `--max-warnings 0`). Its
`@typescript-eslint/no-unused-vars` would now warn on shouldAssertFunction's
`name` parameter, which the ruled signature keeps. 76 already left
`type MatchingKeys` imported and unused in the same file.

Not run on this Mac: Adamic's own checker, the oracle, stock `npm ci` and
`npm run build` of the adapted tree. Correctness's counter numbers are
pending (LEDGER.md entry 8).

## Reproduction

```sh
unit=stage3/adapt/99-cli-debug-assertion-level
export NODE_PATH=$PWD/stage3/api/node_modules
stage3/apply.sh /tmp/tsc-99d > /tmp/tsc-99d-apply.log 2>&1
node "$unit/adapt.cjs" /tmp/tsc-99d > /tmp/tsc-99d-idempotence.log 2>&1
node "$unit/witnesses.cjs" /tmp/tsc-main /tmp/tsc-99d /tmp/tsc-99d-witnesses.json > /tmp/tsc-99d-witnesses.log 2>&1
node "$unit/witnesses.cjs" --mutant /tmp/tsc-main /tmp/tsc-99d /tmp/tsc-99d-mutant.json > /tmp/tsc-99d-mutant.log 2>&1 # exits 1
```

/tmp/tsc-main is a tree applied at main without this directory.
