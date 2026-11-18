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

Measured on main 17a20029. Each census is the latent census over the applied
tree's src/compiler in first-error mode, counting unique (kind, where, reason)
sites as the meter does.

- **The edit.** adapt.cjs changes exactly the four constraints, a second run
  writes nothing, and a moved signature (setTextRangeEnd with a third
  parameter) is refused by name before any write (J-s3f6ymy).
- **Stock tsc.** 2,765 diagnostics over all 713 of upstream's src files,
  harness and tests included, before and after, with 0 new and 0 gone. So no
  caller is turned away: TypeScript's own assignability ignores `readonly`
  (J-s3f6ymy, and J-3t0rq83 for two and for four setters).
- **Emitted JavaScript.** 604 files before and after, 0 differing (J-s3f6ymy).
- **The family.** 2 sites on main, at the casts in setTextRangePos and
  setTextRangeEnd (utilities.ts:10647 and :10657), and 0 with this
  adaptation. Total sites 2,950 before, 2,951 after (J-emhqm5b before,
  J-a3n1f59 after).
- **What appears, each already there behind a cleared stop.** Three NotYet:
  - "a method sharing its name with a getter" at the `factory.createAssignment`
    calls, factory/emitHelpers.ts:315:25 and visitorPublic.ts:485:37. A
    full-mode census of those files on main shows both (J-my8njda), and the
    mutant below hides both again.
  - "an array going where a plain object type is held" at setTextRangeEnd's
    cast, utilities.ts:10657:6. It's the next check at that site, reached when
    visitorPublic's run brings a NodeArray to the setter (J-my8njda).
- **The mutant.** `ReadonlyTextRange` back on setTextRangePos alone brings its
  site back, utilities.ts:10647:6, and hides the two `createAssignment` stops
  again (J-x7nyzap).
- **The lane.** stage3/lane and stage3/lane/adaptedtree green (J-m8q03dk,
  J-0mnptpr).

## Re-running

`node adapt.cjs <tree>` uses stock typescript@6.0.3 through NODE_PATH or
`CENSUS_TYPESCRIPT`, as apply runs every adaptation.
