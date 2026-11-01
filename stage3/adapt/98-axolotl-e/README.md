# 98-e: sys.ts's last own-file diagnostic, and the one it unmasks

Task #7xw984y. lane: run as part of wave N (Loom).

This is a JavaScript-changing adaptation at the two plugin-error readers in
src/server/project.ts, which now read through 94's propertyOf, together with
the two lines that carry the helper to them: project.ts's named import and
one re-export in the compiler's ts namespace barrel. Every other edit erases.
It is also a declared public API change of two members, listed in
api-sanctions.json.

It needs 94 and 98-c, and runs after both: adapt.cjs refuses a tree whose
hostErrors.ts lacks 94's propertyOf, or whose sys.ts lacks 98-c's rows.
It edits src/compiler/sys.ts, src/compiler/types.ts,
src/compiler/_namespaces/ts.ts, src/server/types.ts and src/server/project.ts.

Input: microsoft/TypeScript v6.0.3, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`. Adamic base: 98-c's tip
`3fb29032`, on 94's `8a7cd267`, on main `4d63c0a6`. 97b and 97d are not under
this branch; see "With 97d" below.

sys.ts goes from 1 own-file diagnostic to 0, and nothing new appears anywhere.

## Row 4: sys.ts 1617:13, ModuleImportResult (ruling i)

nodeSystem.require returns `{ module: undefined, modulePath: undefined, error }`
from its catch, and importServicePluginAsync stores its catch's value the same
way; the error is whatever was thrown. Both ModuleImportResult failure members
widen their error to `unknown`:

```diff
 /** @internal */
 export type ModuleImportResult<T = {}> =
     | { module: T; modulePath?: string; error: undefined; }
-    | { module: undefined; modulePath?: undefined; error: { stack?: string; message?: string; }; };
+    | { module: undefined; modulePath?: undefined; error: unknown; };
```

That is src/compiler/types.ts, @internal. The server's is public:

```diff
-export type ModuleImportResult = { module: {}; error: undefined; } | { module: undefined; error: { stack?: string; message?: string; }; };
+export type ModuleImportResult = { module: {}; error: undefined; } | { module: undefined; error: unknown; };
```

The test hosts follow without an edit. TestServerHost
(virtualFileSystemWithWatch.ts:418, :419) and SessionServerHost
(harnessLanguageService.ts:613) declare require and importPlugin with the
compiler's type, which must stay assignable to the server's; with both
widened it is. pluginsAsync.ts's six importPlugin stubs return
ts.server.ModuleImportResult and need nothing.

project.ts's two readers read the plugin error with JavaScript's semantics, so
the logged text is the same:

```diff
-            const err = result.error.stack || result.error.message || JSON.stringify(result.error);
+            const err = propertyOf(result.error, "stack") || propertyOf(result.error, "message") || JSON.stringify(result.error);
```

in importServicePluginSync and in importServicePluginAsync. project.ts imports
compiler internals by name from `./_namespaces/ts.js`; `propertyOf` joins that
list in its sorted place, after `ProjectReference`. No other file reads
ModuleImportResult's error properties; 98-c's README lists every use.

To reach it, the compiler's barrel gains one line after `export * from "../sys.js";`:

```diff
 export * from "../sys.js";
+export * from "../hostErrors.js";
```

This is not the propertyOf-only export the ruling describes. A named
re-export, `/** @internal */ export { propertyOf } from "../hostErrors.js";`,
was tried first: tsc compiles it, but upstream's scripts/dtsBundler.mjs fails
dts-services with "Failed to find namespace for propertyOf", and every line of
the barrel is `export *`. So the barrel exports hostErrors.ts whole, and with
it 47's errorCode and errorMessage. All three functions are @internal, the
adapter checks that hostErrors.ts holds exactly these three and each is
@internal, and none collides with another name in the namespace. The internal
d.ts gains the three as @internal declarations; the public d.ts gains nothing
(below). If the barrel should carry propertyOf alone, the bundler is what
would need to change.

## Row 5: sys.ts 1524:15, System.createSHA256Hash (admitted by name)

Widening row 4 stops the TS2322 on nodeSystem's require property, and the
checker then reports the literal as a whole: `createSHA256Hash: _crypto ?
createSHA256Hash : undefined` (sys.ts:1574) stores undefined into
`createSHA256Hash?(data: string): string` under exactOptionalPropertyTypes.
The member becomes a property in adaptation 32's idiom, as 97d's nineteen did:

```diff
     /** This must be cryptographically secure. Only implement this method using `crypto.createHash("sha256")`. */
-    createSHA256Hash?(data: string): string;
+    createSHA256Hash?: { createSHA256Hash(data: string): string; }["createSHA256Hash"] | undefined;
```

The indexed method type keeps the method's parameter bivariance. The JSDoc
stays where it is. Its readers are an optional call guard
(editorServices.ts:2819 and :2827) and the test host's own method
(virtualFileSystemWithWatch.ts:1098), which still implements it.

Adaptation 20's README records the Oct 19 rule that no method-to-property
conversion is sanctioned, and 97d's exception to it for nineteen members by
name. This head extends 97d's exception by exactly this one member,
System.createSHA256Hash, named in this directory's api-sanctions.json; it
does not edit 20's README, which is not this head's. Whether the lane's
generator needs 20's README to name it too is a routing question, below.

## The adapter

adapt.cjs parses current text with stock typescript@6.0.3 and needs no checker:
each site is pinned by its parsed shape and text. It checks both
ModuleImportResult aliases (the compiler's @internal, the server's not), their
two-member unions, their success members' exact text and their failure
members' error types; the barrel's sys.js anchor and that no hostErrors.js
export exists yet, or exactly this one; hostErrors.ts's three @internal
functions; project.ts's one named import from ./_namespaces/ts.js, sorted
where propertyOf goes, with ProjectReference and its comma before it; each
reader's `err` initializer, and that no other `.error.stack` or
`.error.message` read is in project.ts; System and its createSHA256Hash
member, neither @internal; and the nodeSystem store. Anything else is refused.

It plans every file, then proves the erasing edits erase: stock 6.0.3
transpiles each adapted file and the same file with only its
JavaScript-changing edits, comments kept and removed, and they must match.
Only then does it write. A second run writes zero files. Line breaks are
preserved: every edit is a splice into the file's own text.

## JavaScript

Stock 6.0.3 transpiles all five files before and after, comments kept and
removed. compiler/types.ts, server/types.ts and sys.ts are byte-identical.
The barrel differs by the one `export *` line. project.ts differs at three
lines: the import from ./_namespaces/ts.js gains propertyOf, and the two
readers. The difference at each reader: the same values, the same TypeError
type and message, one extra stack frame per read (propertyOf's).

## Witnesses and mutants

witnesses.cjs extracts each reader's `err` initializer from the tree before
98-e and the tree after, transpiles each, and runs both over a failed result
whose error is null, undefined, a string, an Error with a stack, an object
with only message, one with neither, one whose empty stack falls through to
its message, and a number. 16 of 16 agree on the logged text or on the thrown
type and exact message (`Cannot read properties of null (reading 'stack')`).
In the code itself the readers run only after `if (!result.error)` has
skipped null and undefined; the witnesses hold the semantics anyway.

Mutants, each recorded and reverted:

- The server's widening undone, in the built tree: `npm run build` fails in
  build-tests with 24 errors, the same TS2416, TS2345, TS2322 and TS2430 at
  the test hosts that 98-c's work in progress recorded. Restored through the
  adapter, all five hashes match and the build is green again.
- Row 5 undone, in a scratch copy: `adamic types` on sys.ts reports TS2375 at
  sys.ts:1524:15 again, so sys.ts is back to 1. Restored, sys.ts has none.
- 94's: the helper's null and undefined check removed, in memory. Both
  readers' null and undefined witnesses go red, 12 of 16, exit 1.

outputs/ holds the witness run and the helper mutant's run.

## Declarations

Stock `npm run build` of the tree before 98-e (98-c's) and after. Only the two
declared members move in typescript.d.ts:

```diff
         } | {
             module: undefined;
-            error: {
-                stack?: string;
-                message?: string;
-            };
+            error: unknown;
         };
```

```diff
         /** This must be cryptographically secure. Only implement this method using `crypto.createHash("sha256")`. */
-        createSHA256Hash?(data: string): string;
+        createSHA256Hash?:
+            | {
+                createSHA256Hash(data: string): string;
+            }["createSHA256Hash"]
+            | undefined;
```

The lane's normalize-api.cjs finds exactly two changed declarations,
`TypeAliasDeclaration:["ts","server","ModuleImportResult"]` and
`member:["ts","System","createSHA256Hash"]`, the latter from MethodSignature
to PropertySignature. tsserverlibrary.d.ts is byte-identical: it is a 17-line
re-export of typescript. Neither public file names propertyOf, errorCode or
errorMessage. typescript.internal.d.ts carries the same two members, the
compiler's internal ModuleImportResult's error, and the three hostErrors.ts
functions as @internal declarations; tsserverlibrary.internal.d.ts is
byte-identical.

api-sanctions.json holds the two public rows, each with its owner path in
generate-api.cjs's ownerPath() convention, which drops only the `ts`
namespace:

- `["System", "createSHA256Hash"]`, line 8487 of the pinned
  tests/baselines/reference/api/typescript.d.ts, kind MethodSignature,
  `"replace": "member"`, like 97d's nineteen.
- `["server", "ModuleImportResult", "error"]`, line 2707, kind
  PropertySignature, `"replace": "type"`, from `{ stack?: string; message?:
  string; }` to `unknown`. The reviewer's shape named the path
  `["ModuleImportResult", "error"]`; ownerPath() keeps the `server` namespace,
  and in the pinned file it matches two nodes, the success member's
  `error: undefined` at 2704 and this one, so the row also carries its line.
  generate-api.cjs's owned() refuses a path that matches twice before it reads
  a line, so the generator needs a way to take this row either way.

## Routing

generate-api.cjs does not read any adaptation's api-sanctions.json; it is
hard-coded to its proof files and asserts 222 sanctioned owner lines. As
97d's README says of its nineteen, the lane's sanctioned-api.json will not
include these two until the generator reads them, and the lane's api_check
will report them as unsanctioned declarations until then. Whether it also
needs 20's README to name System.createSHA256Hash is the generator's design,
which this head has not changed.

## With 97d

97d, on cli/axolotl-97d, is not under this branch. Its nineteen members are
others than createSHA256Hash, and its edits are in types.ts, watchPublic.ts,
tsbuildPublic.ts and program.ts. This adapter pins only its own sites by name
and text, so it should apply either way, but it has not been run on a tree
with 97b and 97d applied.

## Measured

| Tree | Own-file diagnostics | Files with any | sys.ts |
| --- | ---: | ---: | ---: |
| 98-c | 142 | 21 of 79 | 1 |
| 98-c and 98-e | 141 | 20 of 79 | 0 |

Attributed as stage3/meter/report.py attributes them, with 94's own-file.py.
sys.ts loses 1617:13 and gains nothing; every other file's diagnostic set is
identical, text, line and column. src/server is outside the meter; its check
is the stock build.

apply.sh passes; its row is `98-axolotl-e | 5 | 8 | 6`. Stock `npm ci` and
`npm run build` exit 0, build-tests included, with no error.

Not run: the sharded lane oracle. It runs in Loom as part of the wave.

## Reproduction

```sh
unit=stage3/adapt/98-axolotl-e
export NODE_PATH=$PWD/stage3/api/node_modules
stage3/apply.sh /tmp/tsc-98e > /tmp/tsc-98e-apply.log 2>&1
node "$unit/adapt.cjs" /tmp/tsc-98e > /tmp/tsc-98e-idempotence.log 2>&1
node "$unit/witnesses.cjs" /tmp/tsc-98c /tmp/tsc-98e /tmp/tsc-98e-witnesses.json > /tmp/tsc-98e-witnesses.log 2>&1
node "$unit/witnesses.cjs" --mutant /tmp/tsc-98c /tmp/tsc-98e /tmp/tsc-98e-mutant.json > /tmp/tsc-98e-mutant.log 2>&1 # exits 1
```

/tmp/tsc-98c is a tree applied at 98-c's tip, without this directory. The
census and own-file commands are 94's README's; the declarations come from
`npm ci && npm run build` in each tree, then `diff -u` of built/local's four
d.ts files. Logs, d.ts diffs, the normalizer's output and proofs/api-lines.md
are on evidence/compiler/axolotl-98e.
