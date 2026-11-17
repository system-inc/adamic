# Readonly script target features, adaptation 77

Adamic, task #r11wppy item 1. This is type-only: the emitted JavaScript does
not change. It is the alias as its own change, as Adamic ruled on item 1.

## The edit

`adapt.cjs` inserts `readonly ` before the innermost `string[]` of the
internal alias `ScriptTargetFeatures` in `src/compiler/utilities.ts`:

```ts
/** @internal */
export type ScriptTargetFeatures = ReadonlyMap<string, ReadonlyMap<string, readonly string[]>>;
```

It parses utilities.ts with stock TypeScript 6.0.3 and refuses before any
write unless it finds exactly one type alias named `ScriptTargetFeatures`,
carrying `@internal`, whose type is upstream's
`ReadonlyMap<string, ReadonlyMap<string, string[]>>`. An innermost array
already readonly keeps its bytes, so a second run changes nothing. No source
coordinate is a rewrite address.

The directory sorts after 76-truthful-casts and before 78-variance-marker, so
every later adaptation and every consumer reads apply's tree with the
innermost array readonly.

## Why

The alias's inner `string[]` receives tsc's shared `emptyArray`, which the
item 1 change types `readonly never[]`. A readonly array seen as a writable
`string[]` is a view that can write a string where never is read, so the
alias has to say readonly too. Of the 97 listed sites in sys.ts's closure at
`03c140ec`, 17 sit in this alias. No other edit was needed in the fixpoint's
second pass (J-d01np2e).

## Public consequence

None. The alias is `@internal`, so the published declaration files do not
change, this directory carries no `api-sanctions.json`, and
`stage3/lane/sanctioned-api.json` is not regenerated.

## Proof

Run on a sealed box: a census before and after, and a mutant that drops the
`readonly`.

| Measurement | Before | After | Mutant |
|---|---:|---:|---:|
| Sites in the alias, of sys.ts's closure | TBD | TBD | TBD |
| The meter's latent total | TBD | TBD | TBD |
| Stock tsc errors new in after, over all upstream src files | | TBD | |
| Lane units failing, whole lane on the after tree | | TBD | |

## Re-running

`node adapt.cjs <tree>` uses stock typescript@6.0.3 through NODE_PATH or
`CENSUS_TYPESCRIPT`, as apply runs every adaptation.
