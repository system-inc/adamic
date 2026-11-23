# The dictionary view sees only records

Task #2v3xtgm, Compiler's ruling (b). JavaScript-changing, value-identical adaptation: Node's every answer
at both owners stays the same, and tsc's `object as MapLike<unknown>` casts there only ever see records.

## The lie

Adamic lowers tsc's `object as MapLike<unknown>` as the checked dictionary view
(`internal/lower/cast_dictionary_view.go`): at the cast, the value must be a record whose values are
references, or the program stops with a run-time NotYet. tsc's types promise that, since `object` is all the
cast says, but at two owners Node hands the cast other kinds:

- **utilities.ts, isJsonEqual.** Its one caller, optionsHaveChanges, hands it every option value
  (program.ts:2579 among others), so `paths`' arrays, `lib`, `types`, `rootDirs` and `plugins` reach
  `equalOwnProperties(a as MapLike<unknown>, b as MapLike<unknown>, isJsonEqual)` as arrays. In Adamic, the
  first option array compared stops the program, wherever tsc compares options: a program reusing an old one,
  a builder made from an old state, the watch host.
- **moduleNameResolver.ts, loadModuleFromExports.** A package.json's `exports` is typed `object`, but JSON
  makes it `true`, a number, a string or an array as readily. At 2629,
  `hasProperty(exports as MapLike<unknown>, ".")` is reached by `true` and a nonzero number. At 2637,
  `allKeysStartWithDot(exports as MapLike<unknown>)` is reached by every truthy kind. (Here and below,
  `exports` is short for `scope.contents.packageJsonContent.exports`, which the source spells out.)

## The edit

`adapt.cjs` pins both owners whole, before and after, with their CRLF bytes and SHA-256 hashes
(`pins.json`), each found by its name through stock 6.0.3's AST, never by a coordinate. It refuses an owner
that is neither pin, a duplicate or a hash drift, and a second run writes nothing. Each file's JavaScript
changes only inside its owner, and no declaration changes under either value of `stripInternal`:
isJsonEqual is `@internal` and keeps its signature, and loadModuleFromExports isn't exported. Its directory
sorts after every adaptation that edits either file (99-getter-callable-views and
99-readonly-inferred-bindings edit utilities.ts), so it plans over the text they leave.

isJsonEqual, before:

```ts
export function isJsonEqual(a: unknown, b: unknown): boolean {
    // eslint-disable-next-line no-restricted-syntax
    return a === b || typeof a === "object" && a !== null && typeof b === "object" && b !== null && equalOwnProperties(a as MapLike<unknown>, b as MapLike<unknown>, isJsonEqual);
}
```

After (its comment shortened here; `pins.json` holds every byte):

```ts
export function isJsonEqual(a: unknown, b: unknown): boolean {
    if (a === b) return true;
    // eslint-disable-next-line no-restricted-syntax
    if (typeof a !== "object" || a === null || typeof b !== "object" || b === null) return false;
    // (why, and the premise)
    if (isArray(a) && isArray(b)) return a.length === b.length && every(a, (element, index) => isJsonEqual(element, b[index]));
    if (isArray(a)) {
        const right = b as MapLike<unknown>;
        for (let index = 0; index < a.length; index++) {
            if (!hasProperty(right, String(index)) || !isJsonEqual(a[index], right[String(index)])) return false;
        }
        for (const key in right) {
            if (hasProperty(right, key) && key !== "length" && !isIndexBelow(key, a.length)) return false;
        }
        return true;
    }
    if (isArray(b)) {
        const left = a as MapLike<unknown>;
        for (const key in left) {
            if (!hasProperty(left, key)) continue;
            if (key === "length") {
                if (!isJsonEqual(left[key], b.length)) return false;
            }
            else if (!isIndexBelow(key, b.length) || !isJsonEqual(left[key], b[Number(key)])) return false;
        }
        for (let index = 0; index < b.length; index++) {
            if (!hasProperty(left, String(index))) return false;
        }
        return true;
    }
    return equalOwnProperties(a as MapLike<unknown>, b as MapLike<unknown>, isJsonEqual);

    // (isIndexBelow: an integer below the length, spelled as the array spells its own indices)
    function isIndexBelow(key: string, length: number) {
        const index = Number(key);
        return Number.isInteger(index) && index >= 0 && index < length && String(index) === key;
    }
}
```

The record side is read only by `for...in`, by element reads and through `hasProperty`, whose one use of its
parameter is `hasOwnProperty.call(map, key)`: each is a use the view admits (`dictionaryViewUse`). The brief
named `hasOwnProperty.call` itself, but utilities.ts doesn't import core's `hasOwnProperty`, and importing it
would change the file outside the owner. `hasProperty`, `isArray` and `every` are imported and read elsewhere in
the file already, so no import's elision moves.

loadModuleFromExports, two conditions, each given a guard in front of it:

| Site | Before | After |
|---|---|---|
| 2629 | `else if (hasProperty(exports as MapLike<unknown>, "."))` | `else if (typeof exports === "object" && hasProperty(exports as MapLike<unknown>, "."))` |
| 2637 | `else if (allKeysStartWithDot(exports as MapLike<unknown>))` | `else if (typeof exports === "object" && !isArray(exports) && allKeysStartWithDot(exports as MapLike<unknown>))` |

2630's `(exports as MapLike<unknown>)["."]` stays: only a record that owns `"."` reaches it. The inner
`typeof exports !== "object"` branch under 2637 can no longer be taken, and stays as upstream wrote it.

## Why every answer is the same

### isJsonEqual, by the kinds of its two sides

Node's answer is equalOwnProperties' (core.ts): every own enumerable key of the left side must be owned by the
right, with equal values compared left first, then every own enumerable key of the right must be owned by the
left, by presence alone. An array's own enumerable keys are its indices, and its own keys add `"length"`, which
`for...in` never visits but `hasOwnProperty` finds.

| a, b | Node before | After |
|---|---|---|
| `a === b` | true | true, the first line |
| either side not an object, or null | false | false, the second line |
| two arrays | each index of a owned by b with isJsonEqual(a[i], b[i]), then each index of b owned by a: for dense arrays, equal lengths and equal elements | `a.length === b.length` and each pair isJsonEqual, left first |
| an array, then a record | each index i of a owned by the record with isJsonEqual(a[i], record[i]), then each own enumerable key of the record owned by a: an index below a's length, or `"length"` | the same two walks, in the same order |
| a record, then an array | each own enumerable key k of the record owned by the array, an index below its length or `"length"`, with isJsonEqual(record[k], array[k]) (`array["length"]` is its length), then each index of the array owned by the record | the same two walks, in the same order |
| two records | equalOwnProperties | the same call |

The answer is asymmetric: `[]` against `{ length: 5 }` is true, since the second walk tests `"length"` for
presence alone, and `{ length: 5 }` against `[]` is false, since the first compares 5 with 0. So every
recursive call keeps the left side first. Every comparison is pure, so the order of the walks changes no answer;
it is kept anyway.

### loadModuleFromExports at 2629, subpath `"."`

The function returns undefined, untraced, on a falsy `exports` (2620), and 2626 sends a string, an array, and a
record with no key starting with `"."` to mainExport before 2629.

| exports | 2629 before | 2629 after | Outcome, both |
|---|---|---|---|
| false, 0, `""`, null | not reached | not reached | undefined, no trace |
| a string, an array, a record with no dot key | not reached | not reached | mainExport is exports |
| true, a nonzero number | `hasProperty` is false: a primitive's wrapper owns no `"."` | `typeof` is false | mainExport is undefined: the trace at the end, then undefined |
| a record with a dot key | `hasProperty(exports, ".")` | true, then the same | the same |

### loadModuleFromExports at 2637, any other subpath

| exports | 2637 before | Before, then | 2637 after | After, then |
|---|---|---|---|---|
| true, a nonzero number | true: `getOwnKeys` of a primitive is `[]`, and `every` of `[]` is true | the inner `typeof !== "object"`: the trace, then undefined | false | the trace at the end, then undefined |
| a nonempty string | false: its key `"0"` doesn't start with `"."` | the trace at the end, then undefined | false | the same |
| a nonempty array | false: its key `"0"` | the trace at the end, then undefined | false, by `isArray` | the same |
| an empty array | true: no keys | loadModuleFromExportsOrImports with `[]`: `hasProperty([], subpath)` is false, since every subpath starts with `"."` (2596, 3152) and none is `"length"`; `getOwnKeys([])` is `[]`, so no expanding key; it returns undefined, so the trace at the end, then undefined | false, by `isArray` | the trace at the end, then undefined |
| a record | computed | computed | true, true, then the same | the same |

Both traces are the same message with the same two arguments (`Export_specifier_0_does_not_exist_in_package_json_scope_at_path_1`,
subpath and the package directory), and each returns `toSearchResult(undefined)`, undefined. The empty array's
walk calls `getLoadModuleFromTargetExportOrImport`, which only makes a closure, so it leaves nothing behind.

## The premise

The two-array branch compares by length and element, which is equalOwnProperties' answer only for arrays that
hold every index below their length and no other own enumerable key. `[, ]` against `[undefined]` is false in
Node, since the hole is no key, and true after. The mixed branches lean on the same premise for the array's own
keys. JSON makes no hole and no extra key, and neither does the command line's parse
(`convertJsonOption`, `parseListTypeOption`), so it holds for every option value tsc builds. It is counted, not
assumed, below.

## Witness

`node witness.cjs TREE` holds the tree to both after owners, then:

- **isJsonEqual:** runs the pinned before and after bodies, each beside the tree's own core.ts
  `equalOwnProperties`, `equateValues`, `every`, `isArray` and `hasProperty`, over every ordered pair of a
  generated table: dense arrays of lengths 0 to 3, records keyed as each array is, with and without `"length"`
  (equal and one more), records keyed the way an index almost is (`"00"`, `"-0"`, `"0.5"`, `"1e0"`, `" 0"`,
  `"+0"`, `"4294967294"`), each nested one level, and strings, numbers, booleans, NaN, -0, null and undefined.
  Every answer compares by `Object.is`. It prints the count of pairs of each kind, requires a pair whose answer
  is asymmetric, and checks six named answers by value (the table above's `[]` and `{ length: 5 }` both ways
  among them). Its control runs arrays with a hole or an extra key, and requires that the after body answers
  some of them differently: the premise is load-bearing, and the count looks for exactly these.
- **loadModuleFromExports:** runs both pinned owners beside the tree's own helpers, the real
  `loadModuleFromExportsOrImports` included, with `trace` and `getLoadModuleFromTargetExportOrImport` recording
  every call, over 28 kinds of `exports` at five subpaths, traced and not. Every trace, target load, factory
  call and result compares. It also reads each guarded condition alone over each kind: 2629's must answer the
  same for every kind that reaches it, and 2637's must differ for exactly `true`, the nonzero numbers and the
  empty array, the rows the table above sends to the same trace.

`WITNESS_MUTANT=symmetric` swaps the argument order in isJsonEqual's two mixed branches, and must fail (the
nested `[{ length: 5 }]` against `{ "0": [] }` turns true). `WITNESS_MUTANT=object-guard-inverted` makes 2629's
guard `!== "object"`, and must fail.

## The premise's count

`WITNESS_COUNT=<a tsc tree> node witness.cjs TREE` adds the count. It compiles an in-memory copy of stock
6.0.3's `lib/typescript.js` under its own path, so it finds its lib files, with isJsonEqual's body opening on a
call that counts what it receives: each array, its holes and its keys beside its indices, each pair's kinds, and
each pair the after body answers differently from the before one. Then, for each `src/*/tsconfig.json` in the
tree, it parses the tsconfig twice, builds a program from the first and one from the second reusing it (the
module resolution and program structure comparisons), a builder from each, the second from the first's state
(the diagnostics, emit and declaration path comparisons), and, where the builder writes its buildinfo, a third
from the buildinfo read back, whose old options are `JSON.parse`'s own. It prints each project's counts and
passes only with isJsonEqual reached, 0 holes, 0 extra keys and 0 differing answers.

stage3 names no corpus of its own for this, so the count runs over the tsc repository's own projects; the lane's
oracle programs would be the wider count.

| Measurement | Value |
|---|---:|
| isJsonEqual calls | `<at landing>` |
| Arrays received | `<at landing>` |
| Holes | `<at landing>` |
| Keys beside indices | `<at landing>` |
| Array and record pairs | `<at landing>` |
| Answers differing | `<at landing>` |

## Check

`node check.cjs TREE`: stock tsc 6.0.3 over src/compiler reports the same diagnostics on the adapted tree and on
the tree with both before owners put back, and the after owners' eight casts each take a value stock tsc types
`object`, the source type the view takes.

## Measured

The census over utilities.ts's and moduleNameResolver.ts's closures, with and without this directory on one box:

| Measurement | Without | With |
|---|---:|---:|
| Distinct stops | `<at landing>` | `<at landing>` |
| Sites cleared | | `<at landing>` |
| Sites new | | `<at landing>` |

## What this leaves

- The view admits each cast statically; the lie is a run-time stop, so the census may count no site cleared.
  What it can count is a new stop: isJsonEqual now narrows an `object` by `isArray` and reads it as an array,
  calls `Number`, `String` and `Number.isInteger`, and captures a narrowed `b` in `every`'s callback, and
  loadModuleFromExports narrows a property read by `isArray`. Each lowers or stops by Adamic's own rules, and the
  census says which.
- No public declaration moves, so `api-sanctions.json`, `stage3/lane/sanctioned-api.json` and the lane's counts
  don't.

## Re-running

`node adapt.cjs <tree>`, `node check.cjs <tree>` and `node witness.cjs <tree>` use stock typescript@6.0.3
through NODE_PATH, as apply runs every adaptation.
