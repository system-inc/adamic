# Readonly script target entries, adaptation 99-readonly-script-target-entries

Adamic, task #r11wppy, the ruling on class 2, form 1: the minimal edit. This
is type-only: the emitted JavaScript does not change. It finishes what
adaptation 77 began in the same function: 77 made the alias's innermost array
readonly, and this gives the literal's `Object.entries` calls that need it the
alias's types to be judged against.

## Why

`getScriptTargetFeatures` builds its value as
`new Map(Object.entries({ Array: new Map(Object.entries({...})), ... }))`.
`Object.entries` infers its element type from the literal it's given, before
any context reaches it. So tsc infers each inner map's value type from that
map's own entries, `string[]` for one, `never[]` for one that holds only the
shared `emptyArray`, `string[] | never[]` for one that holds both. The outer
object then joins its siblings into a union, `Map<string, string[]> |
Map<string, never[]> | ...`, and each inner map is judged against that union,
not against the alias. Adaptation 77's README lists the 16 stops that shape
left in the literal after 77 cleared the alias's own.

Two arguments clear them. The outer call gets the alias's entry type,
`ReadonlyMap<string, readonly string[]>`, which gives each inner map a slot of
the alias's type to be judged against instead of the siblings' union: that
clears 13 stops. The other 3 sit in the three inner maps whose own literal
mixes a bare `emptyArray` with a real array literal: `Promise` (`es2015:
emptyArray` beside `es2018: [...]`), and `BigInt64Array` and `BigUint64Array`
(`es2020: emptyArray` beside `es2022: [...]`). Their `Object.entries` infers
`string[] | never[]`, and the bare `emptyArray` is judged against that union,
so each of the three gets `readonly string[]` on its own call. Every other
inner map's inference already lands on a single type, and an argument there
changes nothing, so it gets none.

## The edit

`adapt.cjs` inserts a type argument right after `Object.entries` in 4 calls
inside `getScriptTargetFeatures` in `src/compiler/utilities.ts`: the outer one,
and the inner one of each of the three mixed maps.

```ts
export const getScriptTargetFeatures: () => ScriptTargetFeatures = /* @__PURE__ */ memoize((): ScriptTargetFeatures =>
    new Map(Object.entries<ReadonlyMap<string, readonly string[]>>({
        Array: new Map(Object.entries({
        ...
        Promise: new Map(Object.entries<readonly string[]>({
            es2015: emptyArray,
            es2018: [
```

It parses utilities.ts with stock TypeScript 6.0.3 and refuses before any
write unless it finds:

- exactly one `ScriptTargetFeatures` alias, of adaptation 77's shape
  `ReadonlyMap<string, ReadonlyMap<string, readonly string[]>>`, since the
  outer argument is its entry type;
- exactly one variable declaration named `getScriptTargetFeatures`, the only
  declaration of an exported `@internal` const typed
  `() => ScriptTargetFeatures`, whose initializer is upstream's
  `memoize((): ScriptTargetFeatures => new Map(Object.entries({...})))`;
- an outer literal whose every member is a property assignment whose value is
  `new Map(Object.entries({...}))`;
- exactly three mixed inner maps, and they are `Promise`, `BigInt64Array` and
  `BigUint64Array`;
- no `Object.entries` call in the function besides the outer one and one per
  inner map.

An inner map is mixed when its own literal has at least one property whose
value is the identifier `emptyArray` and at least one whose value is an array
literal. A fourth mixed map upstream refuses, naming it: it would be a literal
Adamic did not rule on, and whether its stop is the same stop is for a
measurement to say, not for the script to guess.

A call that already carries exactly the type argument it would get keeps its
bytes, so a second run changes nothing; a call carrying any other type
argument refuses, and so does an inner call that does not mix but carries a
type argument. It prints the mixed maps' names and which calls were edited.
No source coordinate is a rewrite address.

The directory sorts after 77-readonly-script-target-features, whose alias the
outer argument names, and after 99-readonly-inferred-bindings, and before
99-wasm-crypto-import. No other adaptation reads or pins the text of
`getScriptTargetFeatures`.

## Public consequence

None. `getScriptTargetFeatures` is `@internal`, and a type argument erases, so
no line of the emitted JavaScript and no line of the published declaration
files moves. This directory carries no `api-sanctions.json`, and
`stage3/lane/sanctioned-api.json` is not regenerated.

## Proof

Measured by scan-all's instrumented probe over sys.ts's closure on main's
apply tree, with stock TypeScript 6.0.3 (J-ezkkktv). "After" is main with the
outer argument and the three mixed maps' arguments, the edit this adaptation
makes.

| Measurement | Before | After |
|---|---:|---:|
| Stops in getScriptTargetFeatures | 16 | 0 |
| Stops new against before | | 0 |
| Stock tsc errors over src/compiler | 1 | 1, the same |

On this adaptation's head, with the directory present:

| Measurement | Before | After |
|---|---:|---:|
| Distinct stops, every rule | TBD | TBD |
| Stops in getScriptTargetFeatures | TBD | TBD |
| Stops new against before | | TBD |
| Stock tsc errors over src/compiler | TBD | TBD |

## Mutants

| Mutant | Brings back | On the head |
|---|---:|---:|
| Take out the outer argument, keep the three inner | 13 | TBD |
| Take out `Promise`'s argument | its 1 | TBD |
| Take out `BigInt64Array`'s argument | its 1 | TBD |
| Take out `BigUint64Array`'s argument | its 1 | TBD |

## Re-running

`node adapt.cjs <tree>` uses stock typescript@6.0.3 through NODE_PATH or
`CENSUS_TYPESCRIPT`, as apply runs every adaptation.
