# Family 136 polling level overrides

Adamic ruled this adaptation on Nov 23 2026 (#bbssdcy, family 136) and accepts the head after its review and census.

This changes the true branch of `getCustomPollingBasedLevels` in
`src/compiler/sys.ts`. The scout and replay located the second spread at
`sys.ts:168:81`, in unit `sys.ts:138:1`, at `d1a5561a48c8`.

`getCustomLevels` creates a fresh local override object on the first truthy
environment value. Its only stores are `Number(customLevel)` for `Low`,
`Medium`, or `High`. Each key is either absent or holds a number. In
particular, `0` and NaN survive nullish coalescing, and an empty string causes
no store. A present `undefined` would make the two forms differ.

Both callers pass `defaultChunkLevels`, a `Levels` literal with exactly
`Low`, `Medium`, and `High` in that order.
`createPollingIntervalBasedLevels` reads exactly those fields. The replacement
enumerates the three keys in the same order while retaining the existing
no-custom branch and the `&&` short-circuit.

The merge sits inside `setCustomPollingValues`, the owner that
`99-correctness-polling-no-enum-write` pins whole by its text and digest
(its `adapt.cjs:17`). That adapter sorts first, so the apply runs it once,
before this edit. Its `contract.cjs` refuses a tree this adapter has
written. Run that contract on its own after snapshot instead: extract
`src/compiler/sys.ts` from
`refs/stage3/99-correctness-polling-no-enum-write-after` into a directory
and pass the directory.

The second-spread refusal is `docs/rulings.md` row 81, "An object literal
holds at most one spread, leading, followed by fields", and row 470,
"Refuse: `adamic/single-spread` in stage 0". This adapter pins the
producer, the relevant defaults, callers and consumer, and the exact merge
snapshot. It changes one object literal in the adapted TypeScript tree.

Run `test.cjs APPLIED_TREE` with Node and the pinned `stage3/api/node_modules`
on `NODE_PATH`, where `APPLIED_TREE` is a tree `stage3/apply.sh` wrote. The
test reads this adapter's before and after snapshots from the tree's
`refs/stage3/` refs, so a later adapter that edits `sys.ts` leaves it valid.
It runs the adapter on the before snapshot, checks its output equals the
after snapshot byte for byte, and runs the written merge beside the stock
spread.

```sh
NODE_PATH="$PWD/stage3/api/node_modules" \
  node stage3/adapt/99-seats-build22-136-level-overrides/test.cjs /tmp/lane-head
```
