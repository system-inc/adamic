# 97-e: program.ts's two caught-message reads

Task #7xw984y. lane: run as part of the Axolotl wave Correctness cuts it into;
Loom's run of that wave is the lane's proof.

This is a JavaScript-changing adaptation at exactly two lines. At both sites,
the caught value's `message` is read through 94's propertyOf and converted
with the global String. No public API changes. types.ts is not edited:
WriteFileCallback and CompilerHost.getSourceFile keep
`onError?: (message: string) => void`, and onError now receives a string
every time it is called.

This adaptation needs 94. adapt.cjs refuses a tree whose hostErrors.ts lacks
94's propertyOf, before it reads anything else.

Input: microsoft/TypeScript v6.0.3, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`. Adamic base: 94's tip
`8a7cd267`, on main `4d63c0a6`.

## The rows

Line numbers come from the meter. They are never used as edit addresses.

| Row | Site | Edit |
| --- | --- | --- |
| 1 | program.ts:406:25, createGetSourceFile, TS18046 | `onError(e.message)` becomes `onError(String(propertyOf(e, "message")))` |
| 2 | program.ts:441:25, createWriteFileMeasuringIO, TS18046 | the same |
| import | after program.ts:326's ts.performance.js import | `import { propertyOf } from "./hostErrors.js";` |

The import goes where 47 put its own in tracing.ts, directly after the
ts.performance.js import. program.ts is CRLF, and so is the inserted line.

## The ruling

Adamic, Nov 1, #axochk: "convert, don't cast." Widening onError to unknown
was measured, and it breaks five places in the tree. A cast to string is
refused too: rulings.md refuses a downcast to a primitive, and Adamic's
lowering refuses it today. So the read is converted:
`onError(String(propertyOf(e, "message")))` at both sites.

The conversion keeps the meaning, which is the message as text. It turns a
crash into the diagnostic tsc meant to report. It has the same shape as
98-d's read-once in appendSourceMap: a JavaScript-changing repair at the one
place where upstream mishandles caller data.

The witnesses are agreement fixtures. Native and Node agree everywhere, so
there is no negative witness.

## Waits

- **94:** propertyOf's own `throw new TypeError` waits on Lower's fix
  (#bbhwj30). See 94's README.
- **String of an unknown:** Adamic's lowering doesn't lower `String()` of an
  unknown yet ("ToPrimitive is not lowered"). Lower adds it as a should.

## The adapter

adapt.cjs parses the current text with stock typescript@6.0.3 and type-checks
it with 47's compiler program. It locates each site by function name, never by
line: the named function's one catch clause, its binding `e` with no type
annotation, unknown by the checker, and read exactly once (by symbol
identity, as 47 does). The read must be onError's one argument. That call is
the only statement in an `if (onError)` block with no else. onError has to be
the returned callback's parameter, and its one parameter must still be
declared `string`. The import has to be absent, or exactly 97e's line
directly after the ts.performance.js import.

String must be the global String. program.ts may declare no binding named
String: no variable, parameter, function, class, enum, namespace or import.
At each site, the value String must resolve only to the default library's
declarations.

The whole file is in exactly one of two states. In the first, both reads are
upstream's `e.message` and there is no import. In the second, both are
`String(propertyOf(e, "message"))` and the import is present. Anything else
is refused before writing. That covers half-adapted files, the withdrawn
`propertyOf(e, "message") as string` and a bare propertyOf call, each refused
as uninventoried. propertyOf may be named nowhere else in the file.

The adapter plans every edit and re-parses the planned text, which must be
CRLF throughout and be exactly the adapted state. Stock 6.0.3 then transpiles
the before and planned texts, with comments kept and with comments removed.
The adapted JavaScript must equal upstream's with the two reads converted and
the import line added. Only then does it check that the file hasn't changed
and write it. A second run writes zero files.

## JavaScript

Stock 6.0.3 transpiles program.ts before and after, with comments kept and
with comments removed. Only three lines differ:

```diff
 import * as performance from "./_namespaces/ts.performance.js";
+import { propertyOf } from "./hostErrors.js";
-                onError(e.message);
+                onError(String(propertyOf(e, "message")));
-                onError(e.message);
+                onError(String(propertyOf(e, "message")));
```

## Witnesses and the mutant

witnesses.cjs extracts createGetSourceFile and createWriteFileMeasuringIO
whole from the tree before 97e and the tree after it. Stock 6.0.3 transpiles
each side, and both run under Node with the callback (readFile, or
actualWriteFile) throwing each value. performance's marks are no-ops,
createSourceFile returns its text, and writeFileEnsuringDirectories calls the
writer once, as its try does. After 97e, propertyOf is the after tree's own,
transpiled from its hostErrors.ts. Each side records what onError receives
and what the call returns, or the thrown error's type, exact message and
frames above the harness.

Then the witness runs what tsc's own callbacks do with the message. Each
site's consumer formats it with stock 6.0.3's formatStringFromArgs and stock
Diagnostics:

- getSourceFile's consumer (program.ts's findSourceFile) uses
  `Cannot_read_file_0_Colon_1`.
- writeFile's consumer (utilities.ts's writeFile) uses
  `Could_not_write_file_0_Colon_1`.

formatStringFromArgs fills each slot through
`Debug.checkDefined(args[+index])` (utilities.ts:8584 on 050880ce), so an
undefined message fails with `Debug Failure.`. Debug.fail hides
checkDefined's own frame, so the witness names the failure by the stock line
its top frame points at.

Every row must match its ruled before and after. 14 of 14 do on Node 24.14.1,
and both sites give the same answer:

| Thrown | onError before | onError after | Consumer before | Consumer after |
| --- | --- | --- | --- | --- |
| null | TypeError `Cannot read properties of null (reading 'message')` | the same TypeError | not reached | not reached |
| undefined | TypeError `Cannot read properties of undefined (reading 'message')` | the same TypeError | not reached | not reached |
| the string "disk failed" | undefined | "undefined" | `Debug Failure.` | `Cannot read file 'a.ts': undefined.` |
| the number 5 | undefined | "undefined" | `Debug Failure.` | the same text |
| an Error "disk failed" | "disk failed" | "disk failed" | `Cannot read file 'a.ts': disk failed.` | the same |
| `{ message: 42 }` | the number 42 | the string "42" | `Cannot read file 'a.ts': 42.` | the same |
| `{}` | undefined | "undefined" | `Debug Failure.` | `Cannot read file 'a.ts': undefined.` |

The consumer column shows getSourceFile's text. writeFile's reads
`Could not write file 'a.js': ...` in the same rows.

Eight rows move, four values at both sites:

- **A thrown string, a number and `{}`:** onError's undefined becomes the
  string "undefined". The consumer's Debug Failure becomes the diagnostic
  text. That is the crash turned into the diagnostic tsc meant.
- **`{ message: 42 }`:** onError's 42 becomes "42". The consumer's text is
  unchanged, because formatStringFromArgs already concatenated the number.

The other six rows don't move. A thrown null or undefined throws the same
TypeError with the same message, now with `propertyOf` as the top frame
above the same callback (`eval` under the harness's new Function). An
Error's message is unchanged.

The mutant makes the helper return undefined on null, in memory, ahead of its
guard. A thrown null then goes red at both sites. The original throws
TypeError, while the mutant hands onError "undefined" and returns normally.
12 of 14 are as ruled, and the run exits 1. Both runs, `outputs/witnesses.json` and `outputs/mutant.json`, live on the branch `cli/axolotl-97e-evidence` (25cbb64c) at `stage3/adapt/97-cli-axolotl-e/outputs/`, not here, because every file in this directory feeds apply's input key.

Refused, writing nothing (program.ts's hash checked after each run):

| Tree | Adapter |
| --- | --- |
| one site converted, the other left, import present | `half-adapted program.ts: createGetSourceFile adapted, createWriteFileMeasuringIO original, import adapted` |
| both sites converted, the import dropped | `half-adapted program.ts: ... import original` |
| `String(` dropped, the reads bare propertyOf calls | `uninventoried onError argument: propertyOf(e, "message")` |
| the withdrawn `as string` | `uninventoried onError argument, the withdrawn cast: propertyOf(e, "message") as string` |
| a local `const String` in program.ts | `program.ts declares its own String` |
| no 94 | `hostErrors.ts lacks 94's propertyOf: adaptation 94 must run first` |

The strict check (Measured, below) run on two of those trees:

- **`String(` dropped:** TS2345 comes back at both sites,
  `Argument of type 'unknown' is not assignable to parameter of type 'string'`.
- **Import dropped:** TS2304 `Cannot find name 'propertyOf'` at both sites.

## Measured

A whole-src/compiler stock check with Adamic's strict options was run on a
94-applied tree, before and after. It loses program.ts:406:25 and 441:25,
both TS18046, and adds nothing: 141 diagnostics become 139. Every other
diagnostic is identical in text and column, and program.ts's are identical
once the inserted import line is subtracted.

`adamic types` on Golem, each tree built by apply.sh from its branch (main 4d63c0a6 with 94, then with 94 and 97e):

| Tree | Whole program | program.ts own-file | patch-set.md row |
| --- | ---: | ---: | --- |
| 94 | 144 | 11 | `94-axolotl-property-read \| 1 \| 8 \| 0` |
| 94 and 97e | 142 | 9 | `97-cli-axolotl-e \| 1 \| 3 \| 2` |

The two removed are exactly program.ts 406:25 and 441:25. Every other diagnostic in every other file is
identical, and program.ts's other nine are the same rows one line lower for the import.

Not run: stock `npm ci` and `npm run build` of the adapted tree. The lane builds it in Correctness' wave.

## Reproduction

```sh
unit=stage3/adapt/97-cli-axolotl-e
export NODE_PATH=$PWD/stage3/api/node_modules
stage3/apply.sh /tmp/tsc-97e > /tmp/tsc-97e-apply.log 2>&1
node "$unit/adapt.cjs" /tmp/tsc-97e > /tmp/tsc-97e-idempotence.log 2>&1
node "$unit/witnesses.cjs" /tmp/tsc-94 /tmp/tsc-97e /tmp/tsc-97e-witnesses.json > /tmp/tsc-97e-witnesses.log 2>&1
node "$unit/witnesses.cjs" --mutant /tmp/tsc-94 /tmp/tsc-97e /tmp/tsc-97e-mutant.json > /tmp/tsc-97e-mutant.log 2>&1 # exits 1
```

/tmp/tsc-94 is a tree applied at 94's tip, without this directory. The census
and own-file commands are 94's README's.
