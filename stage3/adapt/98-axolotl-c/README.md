# 98-c: sys.ts's own-file diagnostics, three rows

Task #7xw984y. lane: run as part of wave N (Loom).

This is a JavaScript-changing adaptation at one site: row 3's caught-value
read goes through 94's propertyOf. The other two edits erase.

This adaptation needs 94: adapt.cjs refuses a tree whose hostErrors.ts lacks
94's propertyOf, before reading anything else. It runs after 92, which pins
sys.ts's before and after hashes when it runs; 98 edits sys.ts later in the
order, so 92's pin is checked against 92's own output and is untouched by 98.
Rerunning 92 over a tree 98 has adapted is refused by that pin, as rerunning 47
over 94's tree is refused by 47's.

Input: microsoft/TypeScript v6.0.3, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`. Adamic base: 94's tip
`8a7cd267`, on main `4d63c0a6`.

sys.ts goes from 4 own-file diagnostics to 1. The one left, TS2322 at
1617:13 (nodeSystem.require against the @internal ModuleImportResult), and the
TS2375 on createSHA256Hash that fixing it unmasks, belong to 98-e.

## The rows

Rows are located by expression; sites.json records each with its occurrence
count. Line numbers are the meter's, never edit addresses.

| Row | Site | Edit | JavaScript |
| --- | --- | --- | --- |
| 1 | 1708:56, `JSON.stringify(cleanupPaths(profile))`, TS2345 | `JSON.stringify(cleanupPaths(profile))!` | erased |
| 2 | 1598:69, `(require("source-map-support") as typeof import("source-map-support")).install()`, TS2307 | `require("source-map-support").install()` | erased |
| 3 | 1281:43, `hitSystemWatcherLimit \|\|= e.code === "ENOSPC"`, TS18046 | `hitSystemWatcherLimit \|\|= propertyOf(e, "code") === "ENOSPC"` | changed |

Row 1 is a required result, as 33 did for sourcemap's
`JSON.stringify(toJSON())!`. profile is the Profiler.stop result the inspector
session hands its callback, a plain object parsed from the protocol's JSON
message with no toJSON hook, and cleanupPaths rewrites callFrame urls in place
and returns the same object; cleanupPaths has no other caller. JSON.stringify
of such an object returns a string or throws, never undefined. The assertion
is on the original call result, at its original evaluation point.

Row 2 drops the type-only assertion. The any the call now reads off comes from
@types/node 25.3.3's global `var require: NodeJS.Require`, whose call
signature is `(id: string): any`, from internal/load/node_types; it is not
adaptation 65's, whose require declarations are file-local to
performanceCore.ts and tracing.ts. Assigning sys.ts's `require` to a number
reports type `Require`, which is how this was found. The parentheses only
held the assertion and stock 6.0.3 emits the same bytes without them.

Row 3 reads the caught value with JavaScript's own semantics: the same value,
or the same TypeError type and message on null and undefined. `||=`'s short
circuit is kept: once the flag is true, the caught value is not read. The
import 47 added becomes `import { errorCode, propertyOf } from "./hostErrors.js";`,
the way 47 imports errorCode; per Adamic's ruling, errorCode stays where 47
used it and new caught-property reads go through propertyOf.

The adapter parses current text with stock typescript@6.0.3, checks the
caught binding is unknown by symbol identity as 47 does, pins each site's
shape and count (the profile write, cleanupPaths' single caller, the
assertion's parenthesized `as` around a require call, the `=== "ENOSPC"` and
`||=` around the read, the one hostErrors.js import), and refuses anything
else. It plans every edit, then proves the erasing edits erase: stock 6.0.3
transpiles the adapted text and the text with only row 3's edits, with
comments kept and removed, and they must match. Only then does it write. A
second run writes zero files.

## JavaScript

Stock 6.0.3 transpiles sys.ts before and after, comments kept and removed.
The only differences are two lines:

```diff
-import { errorCode } from "./hostErrors.js";
+import { errorCode, propertyOf } from "./hostErrors.js";
-                hitSystemWatcherLimit ||= e.code === "ENOSPC";
+                hitSystemWatcherLimit ||= propertyOf(e, "code") === "ENOSPC";
```

The difference at row 3: the same values, the same TypeError type and
message, one extra stack frame (propertyOf's).

## Witnesses and the mutant

witnesses.cjs extracts row 3's statement from the tree before 98 and the tree
after, transpiles each, and runs both with the flag false and true over a
thrown null, undefined, a string, an Error with code ENOSPC, one with code
EMFILE, one without code, and a getter-only code that checks its receiver.
14 of 14 agree on the flag's value or on the thrown type and exact message.
With the flag true, a thrown null throws on neither side.

The mutant is 94's: the helper's null and undefined check removed, in memory.
Thrown null and undefined with the flag false go red (the original throws
TypeError, the mutant gives false), 12 of 14, exit 1. The adapted sys.ts's
hash is checked after. outputs/ holds both runs.

## Measured

| Tree | Own-file diagnostics | Files with any | sys.ts |
| --- | ---: | ---: | ---: |
| 94 | 145 | 21 of 79 | 4 |
| 94 and 98-c | 142 | 21 of 79 | 1 |

Attributed as stage3/meter/report.py attributes them, with 94's own-file.py.
sys.ts loses 1281:43, 1598:69 and 1708:56 and keeps 1617:13; every other
file's diagnostic set is identical, text, line and column.

apply.sh passes; its row is `98-axolotl-c | 1 | 4 | 4`. Stock `npm ci` and
`npm run build` of the adapted tree exit 0, build-tests included, with no
error. Run over a tree without 94, the adapter exits 1 with "hostErrors.ts
lacks 94's propertyOf: adaptation 94 must run first" and writes nothing.

## Reproduction

```sh
unit=stage3/adapt/98-axolotl-c
export NODE_PATH=$PWD/stage3/api/node_modules
stage3/apply.sh /tmp/tsc-98c > /tmp/tsc-98c-apply.log 2>&1
node "$unit/adapt.cjs" /tmp/tsc-98c > /tmp/tsc-98c-idempotence.log 2>&1
node "$unit/witnesses.cjs" /tmp/tsc-94 /tmp/tsc-98c /tmp/tsc-98c-witnesses.json > /tmp/tsc-98c-witnesses.log 2>&1
node "$unit/witnesses.cjs" --mutant /tmp/tsc-94 /tmp/tsc-98c /tmp/tsc-98c-mutant.json > /tmp/tsc-98c-mutant.log 2>&1 # exits 1
```

/tmp/tsc-94 is a tree applied at 94's tip, without this directory. The census
and own-file commands are 94's README's. Logs are on evidence/compiler/axolotl-98c.
