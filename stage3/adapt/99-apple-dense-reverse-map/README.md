# makeReverseMap builds densely

Task #s59veah. JavaScript-changing, value-identical adaptation. Adamic's word (#s59veah,
Nov 20): tsc's source changes to build the map densely, with its values read by index
identical, and rulings.md:318 (array writes only at an existing index) is untouched.

## The edit

`src/compiler/scanner.ts`'s `makeReverseMap` is pinned whole, before and after, with
its CRLF bytes and SHA-256 hashes (`pins.json`). Upstream writes `result[value] = name`
into an empty array, so the array grows by writes past its end and keeps holes where
no entry names an index. Rulings.md:318 stops the first such write. The adapted
owner first sizes the array to the largest value plus one, filled with `undefined` by
`push`. It then writes each entry at an index the array already holds:

```ts
function makeReverseMap<T>(source: Map<T, number>): (T | undefined)[] {
    let length = 0;
    source.forEach(value => {
        if (value >= length) length = value + 1;
    });
    const result: (T | undefined)[] = [];
    while (result.length < length) result.push(undefined);
    source.forEach((value, name) => {
        result[value] = name;
    });
    return result;
}
```

The adapter refuses an owner that is neither pin, a duplicate, or a hash drift, and a
second run writes nothing. The file's JavaScript changes only inside the owner, and its
declarations don't change, since the owner is internal and unexported.

## Why every read is the same

- **A hole now holds undefined.** An index no entry names read `undefined` from the
  hole, and now reads the `undefined` the fill pushed.
- **A written index holds the same name.** Both builds write the map's entries in one
  order, so a name written twice keeps its last write in both.
- **The length is the same.** It is the largest value plus one in both builds.

That holds for any map whose values are non-negative integers. Both of tsc's maps hold
only enum members: `textToToken`'s SyntaxKind tokens, and `charCodeToRegExpFlag`'s
RegularExpressionFlags bits, the largest `Sticky`, 1 << 7.

What could tell the builds apart is reading something other than an index:
`index in array`, an iteration that skips holes, or `Object.keys`. The two arrays have
one reader each, and each reads by index (`tokenToString`'s `tokenStrings[t]` and
`regularExpressionFlagToCharacterCode`'s `regExpFlagCharCodes[f]`), the only uses of
either name in src.

The declared result widens from `T[]` to `(T | undefined)[]`, which is the truth about
the holes. Both readers already return `string | undefined` and
`CharacterCodes | undefined`, so stock tsc's diagnostics are unchanged.

## Witness

`node witness.cjs TREE` holds the adapted tree to the dense owner and runs both pinned
bodies on tsc's own two maps. Each map is recovered from stock 6.0.3's
`tokenToString` and `regularExpressionFlagToCharacterCode`, which are upstream's sparse
builds read by index. For each map it checks five things:

- the lengths are equal;
- every index up to the length plus 16 reads the same value by `Object.is`;
- the dense build reads stock's answer at every index from 0 to 1023;
- the dense build has no hole;
- the dense body, run with each write guarded as :318 guards it (a write lands only at an
  index the array holds), never trips the guard, while upstream's body, guarded the
  same, stops at its first write.

`WITNESS_MUTANT=off-by-one` sizes the array to the largest value instead of one past
it, and the witness must fail.

## The count

The count of tsc sites under rulings.md:318 is on #s59veah. J-2xxcmgy's static
candidates over the pristine 050880ce src/compiler are attached there, and
Correctness's instrumented run over the lane oracle is the dynamic count of record.
This adaptation owns makeReverseMap alone.
