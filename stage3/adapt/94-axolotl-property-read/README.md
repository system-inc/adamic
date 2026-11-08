# 94: one property read with JavaScript's own semantics

Task #7xw984y. Adaptation 94 adds one @internal helper, `propertyOf`, to
adaptation 47's `src/compiler/hostErrors.ts` in the external TypeScript 6.0.3
tree. It changes no call site. Later adaptations (98, and the heads 97 and 99)
read a property off an unknown value through it: a caught value, or a plugin
error. Adaptations apply in number order, so 94 runs after 47 has written
hostErrors.ts and before any adaptation that imports the helper.

Input: microsoft/TypeScript v6.0.3, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`. Adamic base: main at
`4d63c0a6093d84d7f1927e159d2bb7543e91577c`.

94 lands after Lower's fix that admits `throw new TypeError`; until then
stage 0 refuses the helper. See "What Adamic's lowering says" below.

## The helper

Plain TypeScript, no cast, no any, no non-null assertion:

```ts
/** @internal */
export function propertyOf(value: unknown, key: string): unknown {
    if (value === null || value === undefined) {
        throw new TypeError(`Cannot read properties of ${value === null ? "null" : "undefined"} (reading '${key}')`);
    }
    return Reflect.get(Object(value), key, value);
}
```

A plain read `value.key` throws TypeError on null and undefined, with the
message V8 gives; Node 24.22.2 gives `Cannot read properties of null (reading
'code')` for `null.code`, and the helper builds the same text without calling
String(). Any other value
is read through its object wrapper, `Object(value)`, which is the value itself
for an object or a function. The original value is passed as the receiver, so
a getter sees the same `this` a plain read gives it, a primitive included, and
a proxy's get trap sees the same receiver. Inherited properties, getters that
throw, and values of any type come back exactly as the plain read returns them.
The key is a string; the helper makes no claim about symbol keys, whose V8
message differs.

The helper is not added to the `ts` namespace barrel, like 47's two. Nothing
outside sys.ts and tracing.ts imports hostErrors.ts today.

## The adapter

adapt.cjs uses stock typescript@6.0.3's parser on the current text of
hostErrors.ts. It accepts exactly two shapes: 47's file, whose statements are
the exported function declarations errorCode and errorMessage and which names
propertyOf nowhere, or that file with this helper already appended byte for
byte. Anything else is refused before writing: a missing file (47 has not
run), another statement, other modifiers, or a propertyOf with other contents.
It plans the new text, parses it again, checks the three declarations, checks
the file has not changed while planning, and only then writes. A second run
reports zero files.

47's own adapter refuses a hostErrors.ts that differs from its text, so
running 47 again over a tree that 94 has adapted is refused. apply.sh runs
each adaptation once, in order, over a fresh tree, so this does not arise
there.

## Witnesses

witnesses.cjs reads hostErrors.ts from the adapted tree, transpiles it with
stock 6.0.3, and runs the actual propertyOf against a plain read written as
the call sites write it, `value.code`, `value.stack`, `value.message`. Each
value is built fresh for each read. Each case must agree on the value's type
and value, or on the thrown error's constructor name and exact message. Stack
strings are compared by their first line, since the frames below it name
different callers.

21 of 21 agree on Node 24.22.2:

| Value | Key | Both give |
| --- | --- | --- |
| null | code | TypeError, `Cannot read properties of null (reading 'code')` |
| undefined | code | TypeError, `Cannot read properties of undefined (reading 'code')` |
| the string "ENOSPC" | code | undefined |
| the string "ENOSPC" | length | 6 |
| the number 28 | code | undefined |
| `{ code: "ENOSPC" }` | code | "ENOSPC" |
| an object without code | code | undefined |
| `{ code: 28 }` | code | 28 |
| a getter-only code that checks its receiver | code | "ENOSPC" |
| an inherited code | code | "EMFILE" |
| an Error | message | "plugin failed" |
| an Error | stack | a string, first line `Error: plugin failed` |
| an object with only message | message, stack | "only a message", undefined |
| `{ message: 42 }` | message | 42 |
| an object with neither | stack, message | undefined, undefined |
| a function carrying code | code | "ENOSPC" |
| a proxy | code | its trap's value, receiver an object |
| a getter that throws RangeError | code | RangeError, `getter threw` |
| a getter on String.prototype | adamicWitness | `string:ENOSPC`: the receiver is the primitive |

The mutant removes the null and undefined check, so the helper reads off
`Object(null)` without throwing. `witnesses.cjs --mutant` makes that edit to
the transpiled source in memory and runs the same cases: the null and
undefined witnesses go red (the helper gives undefined where the plain read
throws TypeError), 19 of 21 agree, and it exits 1. The tree's hostErrors.ts
hash is checked before and after; the mutant never touches it.

outputs/witnesses.json and outputs/mutant.json are the two recorded runs.

## Measured

Own-file diagnostics come from the stage 3 census tool, attributed the way
stage3/meter/report.py attributes them: each diagnostic to the file of its
primary location, deduplicated across roots. own-file.py does this and prints
the table.

| Tree | Own-file diagnostics | Files with any | hostErrors.ts | sys.ts |
| --- | ---: | ---: | ---: | ---: |
| main | 145 | 21 of 79 | 0 | 4 |
| with 94 | 145 | 21 of 79 | 0 | 4 |

Every file's set of diagnostics is identical, text, line and column. The
census binary built at main's tip is byte-identical to the one built at
`7b9c0fe7`, where the control tree was measured; main's change between them
touches no stage3, internal/load or cloud file, and the control apply and the
94 apply differ in src/ only by hostErrors.ts.

apply.sh passes end to end. Its patch-set row is
`94-axolotl-property-read | 1 | 8 | 0`, one file, eight lines added, none
removed. Stock `npm ci` and `npm run build` of the adapted tree exit 0;
build-tsc, build-services, build-tsserver, build-typings-installer, the
declaration bundles and build-tests all finish.

The emitted JavaScript of every existing file is unchanged: no file but
hostErrors.ts changes, and that file only gains a function nobody calls yet.

## What Adamic's lowering says

Adamic's ruling (a): `throw new TypeError` is admitted at lowering, and the
helper keeps the TypeError, type and message both. docs/0.1.md's exceptions
row already says any value may be thrown. Lower admits it in #bbhwj30, and 94
stacks on that fix.

The helper builds its message without String(): the null and undefined
branch writes `value === null ? "null" : "undefined"`, the same text on Node
(the null and undefined witnesses hold it), because Lower cannot lower
String() of a value that is null or undefined. The message no longer converts.

What remains is the read: Reflect.get with a receiver is Lower's next cut.
Until it lands, the ruling expects hostErrors.ts to read NotYet on the meter,
never Refused.

Measured here, with the census built at main `4d63c0a6`, without #bbhwj30:
hostErrors.ts was `accepted` before 94 and is `Refused` with it, at
hostErrors.ts:14:15, `throwing a TypeError` (internal/lower/exceptions.go:58).
No other root's outcome moves. That is the gap #bbhwj30 closes; this head has
not measured a census with #bbhwj30 in it, so the NotYet outcome above is the
ruling's, not an observation.

47's errorCode stays only where 47 used it. Every new caught-property read
goes through propertyOf. Retiring errorCode is a later Should.

## Reproduction

From the Adamic repository with the setup environment sourced, the TypeScript
6.0.3 checkout in ADAMIC_TYPESCRIPT_SOURCE, and stage3/api installed from its
lockfile. Use fresh output paths.

```sh
unit=stage3/adapt/94-axolotl-property-read
export NODE_PATH=$PWD/stage3/api/node_modules
stage3/apply.sh /tmp/tsc-94 > /tmp/tsc-94-apply.log 2>&1
node "$unit/adapt.cjs" /tmp/tsc-94 > /tmp/tsc-94-idempotence.log 2>&1
node "$unit/witnesses.cjs" /tmp/tsc-94 /tmp/tsc-94-witnesses.json > /tmp/tsc-94-witnesses.log 2>&1
node "$unit/witnesses.cjs" --mutant /tmp/tsc-94 /tmp/tsc-94-mutant.json > /tmp/tsc-94-mutant.log 2>&1 # exits 1
go build -o /tmp/tsc-census ./stage3/census/tool > /tmp/tsc-census-build.log 2>&1
/tmp/tsc-census /tmp/tsc-94/src/compiler /tmp/tsc-94-census.jsonl > /tmp/tsc-94-census.log 2>&1
python3 "$unit/own-file.py" /tmp/tsc-94 /tmp/tsc-94-census.jsonl /tmp/tsc-94-own-file.json > /tmp/tsc-94-own-file.txt
```

The stock build runs in a copy of the adapted tree: `npm ci`, then
`npm run build`, each to its own log. Measured logs are on
evidence/compiler/axolotl-94.
