# Family 25 call-written reads: the test's mutants

`test.cjs` holds two checks, and each mutant below turns exactly one of them
red while the other stays green. Run the test from the repository root with
TypeScript on `NODE_PATH`, as `stage3/test_apply.py`'s `Family25Tests` runs it:

```sh
NODE_PATH=stage3/api/node_modules node --test --test-reporter=tap \
    stage3/adapt/99-seats-family25-call-written-reads/test.cjs
```

## The four inventory mutants

Each edits this directory's `README.md`, in the table under "Exposed stops in
scoped lowering", and turns `README lists the scoped exposed stops` red:

1. `location-406-47`: delete the row for `core.ts:406:47`.
2. `location-409-48`: delete the row for `core.ts:409:48`.
3. `wording-406-47`: empty the census reason cell of the `core.ts:406:47` row.
4. `wording-409-48`: empty the census reason cell of the `core.ts:409:48` row.

The first two leave no row for the location, so the count of rows is 0, not 1.
The last two leave the row with a reason that isn't the census's words.

## The assignment-position mutant

In `adapt.cjs`, change the `assignment position` edit's adapted text from
`currentSymbol.lastAssignmentPos < 0` to `currentSymbol.lastAssignmentPos <= 0`.
It turns `assignment position preserves the result and getter order` red: with
reads `undefined, -1, 0` the adapted function returns true where tsc's returns
false.

All five were seen red by the lane's round 5 (review-e's read), and are run
again on the recut by the Seats tip test kit, which applies each one to a copy
of this directory and requires its named check red and the other green.
