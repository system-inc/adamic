## 8. Raising the assertion level writes a cache entry over the assertion it cached

Status: removed in our fork by JavaScript-changing adaptation
99-cli-debug-assertion-level, under Adamic's Nov 5 ruling (#0kpp2zc,
#b1y988t); draft issue, not filed.
Upstream: TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8.
Location: src/compiler/debug.ts:170, in setAssertionLevel's restore loop,
`(Debug as any)[key] = cachedFunc;`. The cache is written by
shouldAssertFunction at debug.ts:189 and 190. These are stock source
coordinates.

Below its level, shouldAssertFunction stores `{ level, assertion: Debug[name] }`
in assertionCache and replaces `Debug[name]` with noop. When setAssertionLevel
raises the level, its loop means to put the assertion back, and its own
condition compares `Debug[key]` with `cachedFunc.assertion`. It assigns
`cachedFunc`, the entry, instead. After that, `Debug.assertNode` (or any of the
six assertion functions) is a plain object, and every call site throws.

tsc itself sets the level once, from `NODE_ENV`, before any assertion runs
(sys.ts:1973), so the command line never raises it after a cached call. An API
consumer that calls `ts.Debug.setAssertionLevel` after the compiler has run at
None does. Debug is internal and absent from typescript.d.ts, but it is on the
published module object.

Node 24.14.1 counterexample, with `NODE_ENV` unset: call
`Debug.assertNode(undefined, isIdentifier)`, then
`Debug.setAssertionLevel(AssertionLevel.Normal)`, then the same call again.
Before: `Debug.assertNode` is the object with keys `level` and `assertion`, and
the second call throws `TypeError: Debug.assertNode is not a function`. The
same holds for assertEachNode, assertNotNode, assertOptionalNode,
assertOptionalToken and assertMissingNode. After: the second call runs the
assertion and throws `Debug Failure. False expression: Unexpected node.`, as
it does when the level was Normal from the start.

Our patch: stage3/adapt/99-cli-debug-assertion-level/adapt.cjs. Adamic refuses
both namespace writes and ruled adapt, don't lower, so the fork removes the
patching instead of correcting line 170:

```diff
-    const assertionCache: Partial<Record<AssertionKeys, { level: AssertionLevel; assertion: AnyFunction; } | undefined>> = {};
     export function setAssertionLevel(level: AssertionLevel): void {
-        const prevAssertionLevel = currentAssertionLevel;
         currentAssertionLevel = level;
-        ...the restore loop...
     }
     function shouldAssertFunction<K extends AssertionKeys>(level: AssertionLevel, name: K): boolean {
-        if (!shouldAssert(level)) {
-            assertionCache[name] = { level, assertion: Debug[name] };
-            (Debug as any)[name] = noop;
-            return false;
-        }
-        return true;
+        return shouldAssert(level);
     }
```

The imports of `noop` and `getOwnKeys`, which only those lines named, go too,
and shouldAssertFunction's JSDoc no longer describes a cache. Each disabled
assertion now costs a call and a comparison on every call, where upstream paid
them once. No declaration changes. The witnesses, the mutant and the
transpile diff are in the adaptation's README.

Oracle counter (Correctness, #0kpp2zc condition 1; typescript/debug-assertion-level-census
aa84f129, stage3/census/debug-assertion-level): across the whole lane oracle,
setAssertionLevel was called 9 times, every one from sys.ts's module
initialization, each from level 0 to 1, and all 9 with an empty
assertionCache. So no call found a cached assertion to restore, and no noop
swap ran. The instrumented oracle kept 106,366 passing plus the one sanctioned
failure, with identities exactly main's: the adaptation is unobservable across
tsc's baselines.

### Draft GitHub issue

Title: Debug.setAssertionLevel replaces a cached assertion with its cache entry, so the next call throws "is not a function"

Version: TypeScript 6.0.3, reproduced on Node 24.14.1.

Minimal reproduction:

```sh
mkdir tsc-assertion-level-repro
cd tsc-assertion-level-repro
npm install --no-audit --no-fund typescript@6.0.3
```

Save this as repro.cjs and run `node repro.cjs` with `NODE_ENV` unset:

```js
const ts = require("typescript");
// With NODE_ENV unset, loading typescript set the assertion level to None.
ts.Debug.assertNode(undefined, ts.isIdentifier);
ts.Debug.setAssertionLevel(ts.AssertionLevel.Normal);
console.log(typeof ts.Debug.assertNode, JSON.stringify(Object.keys(ts.Debug.assertNode)));
ts.Debug.assertNode(undefined, ts.isIdentifier);
```

Expected: raising the level restores the assertion. The log prints
`function []`, and the last call fails the assertion with
`Debug Failure. False expression: Unexpected node.`, as it does when
setAssertionLevel runs before the first call.

Actual:

```text
object ["level","assertion"]
TypeError: ts.Debug.assertNode is not a function
```

Cause: setAssertionLevel's restore loop (src/compiler/debug.ts:170) assigns
the cache entry, `(Debug as any)[key] = cachedFunc;`, where its condition on
the line above compares against `cachedFunc.assertion`.

Suggested upstream change: assign `cachedFunc.assertion`. Alternatively, have
shouldAssertFunction return `shouldAssert(level)` without replacing anything,
which removes the cache and the restore loop. That is what our fork does. This
draft makes no claim about which one upstream prefers.
