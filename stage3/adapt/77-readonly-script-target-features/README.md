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

The alias's inner `string[]` is where getScriptTargetFeatures' map literal
lands: its value is a `Map<string, Map<string, string[]> | Map<string,
never[]> | Map<string, string[] | never[]>>`, the entries' own types joined,
because the shared `emptyArray` and the literal arrays beside it are inferred
apart. Seen as the alias with a writable inner array, that's a view that can
write a string where never is read, refused at utilities.ts:1387. With the
inner array readonly it's a read-only view, and the stop goes.

## The count: one stop, not seventeen

The readonly fixpoint's second pass (J-d01np2e) put 17 listed sites under
this alias. It matched a site to a slot by the diagnostic's span, and the
alias's diagnostic spans the whole literal, so it took in the 16 sites inside
it. Measured by scan-all over sys.ts's closure (J-pfrgr6m, J-b7xxqdq), this
adaptation clears exactly one distinct stop, utilities.ts:1387, and adds none.

The other 16 are still stopped, each at an inner entry of the same literal
(utilities.ts:1417, 1420, 1466, 1469, 1472, 1475, 1715, 1716, 1778, 1885,
1914, 1915, 1928, 1929, 1963 and 1966). Each is an inner `new Map` (or, three
of them, an `emptyArray`) seen as the union tsc infers from its siblings,
`Map<string, string[]> | Map<string, never[]> | ...`, never as the alias. That's
the shape Adamic's ruling at 08:05Z on #r11wppy calls an inferred union, which
waits on the internal change's census. An explicit type argument on the outer
`new Map` was measured and reaches none of them (J-b7xxqdq: 0 gone, 0 new), since
`Object.entries` infers its element type from the literal before any context.

## Public consequence

None. The alias is `@internal`, so the published declaration files do not
change, this directory carries no `api-sanctions.json`, and
`stage3/lane/sanctioned-api.json` is not regenerated.

## Proof

Sealed, by scan-all's instrumented probe over sys.ts's closure in apply's tree,
with stock TypeScript 6.0.3:

| Measurement | Before (the mutant) | After |
|---|---:|---:|
| Distinct stops, every rule | 3,538 | 3,537 |
| utilities.ts:1387, the literal seen as the alias | stopped | gone |
| Stops new with 77 | | 0 |
| Stock tsc errors over src/compiler | 1 | 1, the same one |

"Before" is apply's tree with this adaptation's one `readonly ` taken back out,
so it is also the mutant that drops it, and it brings the stop back. The adapted
tree's product test and stage3/lane pass with 77 (J-2kx5eqg).

## Re-running

`node adapt.cjs <tree>` uses stock typescript@6.0.3 through NODE_PATH or
`CENSUS_TYPESCRIPT`, as apply runs every adaptation.
