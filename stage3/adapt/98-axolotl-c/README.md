# 98-c: sys.ts's own-file diagnostics (work in progress, stopped)

Task #7xw984y. This adaptation needs 94: its caught-value read goes through
94's propertyOf, and adapt.cjs refuses a tree whose hostErrors.ts lacks it.
It runs after 92, which pins sys.ts's before and after hashes when it runs;
98 edits sys.ts later in the order, so 92's pin is checked against 92's own
output and is untouched by 98. Rerunning 92 over a tree 98 has adapted is
refused by that pin, as rerunning 47 over 94's tree is refused by 47's.

Input: microsoft/TypeScript v6.0.3, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`. Adamic base: 94's tip
`8a7cd267`, on main `4d63c0a6`.

This is not ready to land. It stops at three places, each needing a ruling:
row 2 is held because the site does not see 65's require; row 4 as ruled
breaks the stock build in the test harness; and row 4 unmasks a fifth sys.ts
diagnostic no row covers. sys.ts measures 4 to 2, not 0.

## The rows

Rows are located by expression; sites.json records each with its occurrence
count, and the line numbers are the meter's, never edit addresses.

| Row | Site | Action | JavaScript |
| --- | --- | --- | --- |
| 1 | sys.ts 1708:56, `JSON.stringify(cleanupPaths(profile))`, TS2345 | `!`, a required result as 33 did for sourcemap | erased |
| 2 | sys.ts 1598:69, `(require("source-map-support") as typeof import("source-map-support")).install()`, TS2307 | held, validated unchanged | unchanged |
| 3 | sys.ts 1281:43, `e.code` in `hitSystemWatcherLimit \|\|= e.code === "ENOSPC"`, TS18046 | `propertyOf(e, "code")`, imported beside 47's errorCode | changed |
| 4 | types.ts @internal ModuleImportResult, failure member's `error`, TS2322 at sys.ts 1617:13 | widened to `unknown` | erased |

Row 1: profile is the Profiler.stop result the inspector session hands its
callback, a plain object parsed from the protocol's JSON message with no
toJSON hook, and cleanupPaths rewrites callFrame urls in place and returns
the same object; cleanupPaths has no other caller. JSON.stringify of such an
object returns a string or throws, never undefined.

Row 3: the import 47 added becomes
`import { errorCode, propertyOf } from "./hostErrors.js";`. The `||=`'s short
circuit is kept: once the flag is true, the caught value is not read.

## Where it stops

**Row 2, held.** The ruling removes the type-only assertion so the call reads
off 65's explicit-any require, and says to stop if that isn't what the site
sees. It isn't: assigning sys.ts's `require` to a number reports type
`Require`, @types/node 25.3.3's global `var require: NodeJS.Require` with
`(id: string): any`, from internal/load/node_types. 65 declares require only
in performanceCore.ts and tracing.ts, file-local. Removing the assertion in a
scratch copy clears TS2307 (sys.ts then reads 1, the unmasked diagnostic
below), so this is one ruling away, but not this head's to make.

**Row 4 breaks the stock build.** `npm ci` passes; `npm run build` fails in
build-tests with 24 errors, all in src/harness and src/testRunner.
TestServerHost (testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:418
and :419) and SessionServerHost (harness/harnessLanguageService.ts:613) declare
require and importPlugin with the compiler's ModuleImportResult, and those
classes are ServerHosts, whose require returns the server's public
ModuleImportResult (src/server/types.ts:24), `error: { stack?: string;
message?: string; }`. The compiler's type must stay assignable to the
server's; `unknown` is not, so TS2416 at those properties and TS2345, TS2322
and TS2430 wherever a TestServerHost or SessionServerHost is passed as a
ServerHost. Fixing it means widening the server's public type too (a public
API change) or changing the test hosts' types; neither is ruled.

**Row 4's project.ts readers are not edited.** The ruling has project.ts's
two readers, importServicePluginSync (:507) and importServicePluginAsync
(:537), read through propertyOf. Their `result` is the server's public
ModuleImportResult, imported from `./_namespaces/ts.server.js`, not the
compiler's, so widening types.ts does not retype them, and propertyOf is not
reachable through the `ts` namespace: hostErrors.ts is in no barrel, and only
sys.ts and tracing.ts import it. Reaching it would add hostErrors.ts to
src/compiler/_namespaces/ts.ts, which is not ruled.

**A fifth diagnostic is unmasked.** With row 4, sys.ts gains TS2375 at
1524:15: the nodeSystem literal is not assignable to System under
exactOptionalPropertyTypes, because `createSHA256Hash: _crypto ?
createSHA256Hash : undefined` (sys.ts:1574) gives `undefined` to
`createSHA256Hash?(data: string): string`. Before, the TS2322 on the literal's
require property was reported instead. No row covers it.

Every ModuleImportResult use in src/: compiler/sys.ts:32 (import) and :1447
(System.require); server/types.ts:24 (the public type) and its users
server/project.ts:149, :502, :526, with the readers :507 and :537;
harness/harnessLanguageService.ts:613; testRunner/unittests/helpers/
virtualFileSystemWithWatch.ts:49, :418, :419; testRunner/unittests/tsserver/
pluginsAsync.ts's six importPlugin stubs, typed with ts.server's. No other
file reads its error's properties; services reads none.

## The types.ts diff

```diff
 /** @internal */
 export type ModuleImportResult<T = {}> =
     | { module: T; modulePath?: string; error: undefined; }
-    | { module: undefined; modulePath?: undefined; error: { stack?: string; message?: string; }; };
+    | { module: undefined; modulePath?: undefined; error: unknown; };
```

ModuleImportResult is @internal in types.ts; the adapter checks the tag.
project.ts has no diff.

## JavaScript

This adaptation changes JavaScript at one site. Stock 6.0.3 transpiles
sys.ts and types.ts before and after, with comments kept and removed: types.ts
is byte-identical, and sys.ts differs at exactly two lines, the hostErrors.js
import and row 3's read. Row 1's `!` erases. The adapter itself proves the
erasing edits erase before writing: it transpiles the adapted text and the
text with only the JavaScript-changing edits, and they must match.

The difference at row 3: the same values, the same TypeError type and
message, one extra stack frame (propertyOf's).

witnesses.cjs extracts row 3's statement from the tree before 98 and the tree
after, transpiles each, and runs both with the flag false and true over a
thrown null, undefined, a string, an object with code ENOSPC, one with code
EMFILE, one without code, and a getter-only code. 14 of 14 agree; with the
flag true, a thrown null does not throw on either side. The mutant removes
94's null and undefined check from the helper: thrown null and undefined with
the flag false go red, 12 of 14, exit 1. outputs/ holds both runs. The reader
witnesses were not run, since project.ts is not edited.

## Measured

| Tree | Own-file diagnostics | Files with any | sys.ts |
| --- | ---: | ---: | ---: |
| 94 | 145 | 21 of 79 | 4 |
| 94 and 98-c | 143 | 21 of 79 | 2 |

Attributed as stage3/meter/report.py attributes them, with 94's own-file.py.
Every other file's diagnostic set is identical. sys.ts loses 1281:43, 1617:13
and 1708:56, keeps 1598:69, and gains 1524:15.

apply.sh passes; its row is `98-axolotl-c | 2 | 4 | 4`. A second adapter run
writes zero files and leaves both hashes. Stock `npm run build` fails as
above.

## Reproduction

```sh
unit=stage3/adapt/98-axolotl-c
export NODE_PATH=$PWD/stage3/api/node_modules
stage3/apply.sh /tmp/tsc-98c > /tmp/tsc-98c-apply.log 2>&1
node "$unit/adapt.cjs" /tmp/tsc-98c > /tmp/tsc-98c-idempotence.log 2>&1
node "$unit/witnesses.cjs" /tmp/tsc-94 /tmp/tsc-98c /tmp/tsc-98c-witnesses.json > /tmp/tsc-98c-witnesses.log 2>&1
node "$unit/witnesses.cjs" --mutant /tmp/tsc-94 /tmp/tsc-98c /tmp/tsc-98c-mutant.json > /tmp/tsc-98c-mutant.log 2>&1 # exits 1
```

/tmp/tsc-94 is a tree applied without this directory. The census and
own-file commands are 94's. Logs are on evidence/compiler/axolotl-98c-wip.
