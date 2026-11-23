Adamic ruled this adaptation on Nov 23 2026 (#bbssdcy, family 170); Adamic accepts the head after its review and census.

The last spread in `src/compiler/utilities.ts`'s `nodeCoreModules` initializer
was the family 170 site (`utilities.ts:12169:8` in the scout's tree).
The inner array iterates the existing Set when evaluation reaches that operand,
following [rulings.md row 204](../../../docs/rulings.md#L204). The outer array
immediately spreads the result. The outer Set constructor still keeps each
value's first insertion.

The equivalence assumes the host's own Array and Set iterators: a caller who
replaces a built-in iterator is outside the contract, as Adamic answered.

Run `test.cjs APPLIED_TREE` with Node and the pinned `stage3/api/node_modules`
on `NODE_PATH`, where `APPLIED_TREE` is a tree `stage3/apply.sh` wrote. The
test reads this adapter's before and after snapshots from the tree's
`refs/stage3/` refs, so a later adapter that edits `utilities.ts` leaves it
valid. It compares every element with the real lists, then checks an
out-of-order Set with a duplicate from an earlier operand.
