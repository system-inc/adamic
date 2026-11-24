# Family 62, NodeArray range

This covers the refusal at `utilities.ts:10657:6` reached from
`visitorPublic.ts:420:1` in `addDefaultValueAssignmentsIfNeeded`. The body
change follows Adamic's Nov 22 ruling on #bbssdcy. Rulings 89 and 95 govern
augmented arrays and mutable slots. Ruling 400's Nov 17 and 18 extensions
provide the writable types used here. Those extensions and ruling 401 do not
authorize this JavaScript edit. The Nov 22 split authorizes it by value identity.

## Value identity

The adapted `src/compiler/factory/utilitiesPublic.ts:10` has:

```ts
export function setTextRange<T extends TextRange>(range: T, location: Readonly<TextRange> | undefined): T {
    return location ? setTextRangePosEnd(range, location.pos, location.end) : range;
}
```

After `85-writable-range-setters`, adapted `src/compiler/utilities.ts:10647`,
`:10657`, and `:10667` have:

```ts
export function setTextRangePos<T extends TextRange>(range: T, pos: number): T {
    (range as TextRange).pos = pos;
    return range;
}
export function setTextRangeEnd<T extends TextRange>(range: T, end: number): T {
    (range as TextRange).end = end;
    return range;
}
export function setTextRangePosEnd<T extends TextRange>(range: T, pos: number, end: number): T {
    return setTextRangeEnd(setTextRangePos(range, pos), end);
}
```

In `visitorPublic.ts:400` the guarded `if (nodes)` passes a defined
`NodeArray` to `nodesVisitor`. Its overload returns
`NodeArray<TOut> | (TInArray & undefined)`, with the second member impossible
for that input. The worker called at line 412 takes a required `NodeArray`
and reads `parameters.length` before reaching the edited branch. The
`parameters` binding remains fixed throughout the worker. At the old
`setTextRange` call on line 431 the undefined branch is therefore unreachable.

Both versions call `context.factory.createNodeArray(result, parameters.hasTrailingComma)`
once, then read `parameters.pos` and `parameters.end` in that order. The
two locals preserve the original argument evaluation order even if a range
write can be observed. The result of the factory call receives a `pos` write
followed by an `end` write. Each setter returns its input by identity and makes
only its named write. The direct writes return the same factory result.

## Output test

With the locked `stage3/api` packages on `NODE_PATH`, run
`node stage3/adapt/99-seats-build22-62-node-array-range/test.cjs <tree>`.
The test takes a tree written by `stage3/apply.sh` and reads only this
adapter's before and after refs under `refs/stage3/`. Later adapters can
edit the working tree without affecting the test. It applies the adapter
in scratch, checks its output against the after ref, and evaluates the
transpiled old and new workers with the same factory harness. It checks
factory identity, read and write order, distinct endpoints, and identity
on the no-result path. It verifies zero edits and identical bytes on
reapply. Replacing the end value with the start value makes it fail.
