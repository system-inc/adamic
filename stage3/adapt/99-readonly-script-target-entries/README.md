# Readonly script target entries, adaptation 99-readonly-script-target-entries

Adamic, task #r11wppy, the ruling on class 2. This is type-only: the emitted
JavaScript does not change. It finishes what adaptation 77 began in the same
function: 77 made the alias's innermost array readonly, and this gives the
literal's `Object.entries` calls the alias's types to be judged against.

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

An explicit type argument closes each inference at the alias's types. The
outer call gets the alias's entry type, so every property of the outer literal
is judged as a `ReadonlyMap<string, readonly string[]>`; each inner call gets
`readonly string[]`, so every entry of an inner literal is judged as a
read-only view. Nothing else moves.

## The edit

`adapt.cjs` inserts a type argument right after `Object.entries` in 55 calls
inside `getScriptTargetFeatures` in `src/compiler/utilities.ts`: the outer one,
and the one in each of the outer literal's 54 inner maps.

```ts
export const getScriptTargetFeatures: () => ScriptTargetFeatures = /* @__PURE__ */ memoize((): ScriptTargetFeatures =>
    new Map(Object.entries<ReadonlyMap<string, readonly string[]>>({
        Array: new Map(Object.entries<readonly string[]>({
            es2015: [
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
  `new Map(Object.entries({...}))`, and exactly 54 of them;
- no `Object.entries` call in the function besides those 55.

A call that already carries exactly the type argument it would get keeps its
bytes, so a second run changes nothing; a call carrying any other type
argument refuses. It prints the count of inner calls and which were edited.
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
apply tree, with stock TypeScript 6.0.3 (J-61zs7wd). "Outer" is main with only
the outer argument; "both" is main with the outer argument and all 54 inner
ones, the edit this adaptation makes.

| Measurement | Main | Outer | Both |
|---|---:|---:|---:|
| Stops in getScriptTargetFeatures | 16 | 3 | 0 |
| Stops new against main | | 0 | 0 |
| Stock tsc errors over src/compiler | 1 | 1, the same | 1, the same |

On this adaptation's head, with the directory present:

| Measurement | Before | After |
|---|---:|---:|
| Distinct stops, every rule | TBD | TBD |
| Stops in getScriptTargetFeatures | TBD | TBD |
| Stops new against before | | TBD |
| Stock tsc errors over src/compiler | TBD | TBD |

## Mutants

Adamic named two:

| Mutant | Brings back |
|---|---:|
| Take out the outer argument, keep the 54 inner | 13 |
| Take out one inner argument | that map's 1 |

## Re-running

`node adapt.cjs <tree>` uses stock typescript@6.0.3 through NODE_PATH or
`CENSUS_TYPESCRIPT`, as apply runs every adaptation.
