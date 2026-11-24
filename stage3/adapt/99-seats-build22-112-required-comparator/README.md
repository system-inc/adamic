# Required string comparator in matchFiles

Adamic reads this JavaScript-changing head after review and census under
her Nov 22 ruling on #bbssdcy for Seats lane build22-112, family 112.
The recorded refusal is `core.ts:1049:82`, "a comparator that isn't a function", in
`utilities.ts:9828:1`. This applies rows 322 and 335 of
`docs/rulings.md`: sort requires an explicit comparator and remains stable
as in Node.

`matchFiles` calls its nested `visitDirectory`, which passes the imported
`compareStringsCaseSensitive` function directly for both `files` and
`directories`. Its operands accept strings or undefined and it returns
the numeric comparison result. On these call paths the comparator is
always present.

The fully adapted base has this `toSorted` body.

```ts
{
    return (array.length === 0 ? emptyArray as readonly T[] : array.slice().sort(comparer)) as unknown as SortedReadonlyArray<T>;
}
```

The private string helper uses that same shared `emptyArray` on the
empty branch. Otherwise it sorts a fresh slice and leaves the input in
its original order. The declared comparator accepts two strings and
returns a number. The existing function accepts that parameter view
directly. The generic `toSorted` and its other callers keep their
existing behavior.

The adapter locates both loops under their syntax-tree owners and
checks the exact call text and the two core declarations before it
writes. A second run checks the already-adapted snapshots.

## Capture and snapshot guards

The name guard searches all of utilities.ts, including comments. The
original input must have no occurrence of toSortedStrings. Before writing
the rewritten tree, and on an already-adapted tree, the only three
occurrences must be the final top-level declaration's name and the two
callee tokens found by the loop checks. The AST check also rejects
identifiers that spell the name with Unicode escapes.

The binding guard pins the named value import from ./_namespaces/ts.js
and the namespace's value star export from ../core.js by their AST
nodes. A three-file checker follows the original callees and comparators
to their pinned declarations in core.ts. Before writing the rewritten
text and on reapplication it follows the comparators and the helper's
emptyArray reference to core. Binding errors name the affected symbol
and its use. The host serves only utilities.ts, _namespaces/ts.ts and
core.ts from memory. With noResolve, no library and no
ambient types, the namespace's other exports stay unresolved.

The exact snapshots cover the three core declarations, the two direct
for-of expressions in matchFiles's nested visitDirectory and the
generated helper declaration. The first write also pins the EOF gap to
one line ending. The checks have these boundaries.

| Input | Result or expected error |
| --- | --- |
| Local shadow of original sort | Reject before writing with `toSorted binding at visitDirectory files sort` |
| Local shadow of comparator on base input | Reject before writing with `compareStringsCaseSensitive binding at visitDirectory files sort` |
| Local shadow of comparator on adapted input | Reject before writing with `compareStringsCaseSensitive binding at visitDirectory files sort` |
| Local helper shadow in base or adapted input | Reject before writing with `unexpected toSortedStrings occurrence in utilities.ts` |
| Helper name in a comment | Reject before writing with `unexpected toSortedStrings occurrence in utilities.ts` |
| Unicode-escaped local helper shadow | Reject before writing with `unexpected toSortedStrings identifier in utilities.ts` |
| Duplicate loop with the same first argument | Reject before writing with `expected one visitDirectory files sort, found 2` |
| Target sort taken out of its direct for-of | Reject before writing with `expected one visitDirectory files sort, found 0` |
| Target files comparator replaced with undefined | Reject before writing with `snapshot drift in files sort` |
| Trailing comment on base input | Reject before writing with `snapshot drift at utilities.ts end of file` |
| Partially adapted input | Reject before writing with `expected one toSortedStrings, found 0` |
| Third direct loop sorting files.slice() | Accept and preserve its generic toSorted for separate review |
| Trailing comment after the adapted helper | Accept and preserve the bytes because the helper is still the final statement |
| Namespace retarget of original sort | Reject with `toSorted binding at visitDirectory files sort` |
| Namespace retarget of comparator on adapted input | Reject with `compareStringsCaseSensitive binding at visitDirectory files sort` |
| Namespace without its core re-export | Reject with `_namespaces/ts.ts core value star export` |
| Namespace retarget of emptyArray on adapted input | Reject with `emptyArray binding at toSortedStrings empty branch` |

## Adapter test

Run this command from the Adamic checkout after a full stage3 apply.

    NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/99-seats-build22-112-required-comparator/test.cjs <applied-tree>

The test reads this adapter's before and after refs from the applied
tree and runs the production adapter next to the test in a temporary
tree. It drives the generated helper beside the original toSorted in
one VM context and checks shared empty identity, comparator elision,
stable ties and fresh nonempty copies with unchanged input. It checks
the after bytes, reapplication and the guard inventory, including the
expected stderr message and unchanged bytes on every rejection.
