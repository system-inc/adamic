# Writable range setters, adaptation 85

Adamic, task #aryr7ea, ruled Nov 18 under rulings.md:391. Type-only: the
emitted JavaScript doesn't change.

## The edit

`adapt.cjs` replaces the `ReadonlyTextRange` constraint on the type parameter
of tsc's four range setters in `src/compiler/utilities.ts` with `TextRange`:

```ts
export function setTextRangePos<T extends TextRange>(range: T, pos: number): T
export function setTextRangeEnd<T extends TextRange>(range: T, end: number): T
export function setTextRangePosEnd<T extends TextRange>(range: T, pos: number, end: number): T
export function setTextRangePosWidth<T extends TextRange>(range: T, pos: number, width: number): T
```

It parses utilities.ts with stock TypeScript 6.0.3 and refuses before any write
unless it finds exactly one declaration of each, with the parameters upstream
writes and one type parameter `T` constrained as upstream constrains it. A
constraint already adapted keeps its bytes, so a second run changes nothing.

## The premise

Each setter writes the range's `pos` or `end`. setTextRangePos and
setTextRangeEnd write through a cast, `(range as TextRange).pos = pos`, and
setTextRangePosEnd and setTextRangePosWidth write through them. A
`ReadonlyTextRange` parameter was tsc claiming it wouldn't write what it then
writes. Constrained to `TextRange`, each signature says what its body does.
After adaptation 80, every Node and NodeArray already is a `TextRange`, so the
only caller this could turn away is one that really hands over a readonly range
to be written.

Adamic ruled the first two by name. The other two are in this adaptation by
measurement (J-vtw1t3r): with only the first two constrained, setTextRangePosEnd
hands its own `T extends ReadonlyTextRange` to them, so the refusal moves one
call up, to utilities.ts:10667. All four are `@internal`, so no line of the
published `typescript.d.ts` changes and there's nothing to sanction.

This extends rulings.md:391 again, by name, for these four functions. Language
holds the row with 80's.

## Measured

MEASUREMENTS

## Re-running

`node adapt.cjs <tree>` uses stock typescript@6.0.3 through NODE_PATH or
`CENSUS_TYPESCRIPT`, as apply runs every adaptation.
