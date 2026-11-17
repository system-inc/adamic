# getRestrictiveTypeParameter: build the type parameter, then store it

JavaScript-changing and value-identical, task #msn1j8r. Adamic ruled it Nov 17, 2026,
under Compiler's delegation.

checker.ts's getRestrictiveTypeParameter makes a type parameter's restrictive
instantiation in one comma expression. It stores a fresh type parameter in
`tp.restrictiveInstantiation`, then sets that one's constraint through
`(tp.restrictiveInstantiation as TypeParameter).constraint = noConstraintType`,
because the cache field is declared `Type` on every Type. That cast is one of the 29
`Type` to `TypeParameter` stops on #msn1j8r (adapted checker.ts:20671 at 03c140ec),
and it is the only one an upstream edit can take away without a check. The other 28
wait on #b5w3ycg's checked views.

The edit, in the one owner pinned whole in owners.json: the fresh type parameter is
built in a local, its constraint is set, and then it is stored. The early return of
`tp` and the cached return are the same tests in the same order (`!x` for `x ||`).
No cast is left, and nothing reads `tp.restrictiveInstantiation` as a TypeParameter.

**Premise:** nothing reads `tp.restrictiveInstantiation` between upstream's store and
its constraint write. In upstream, those two writes are adjacent operands of one comma
expression, with only the cast read between them, so the new order (constraint first,
then the store) is observable by nothing else. createTypeParameter runs before both,
as it did. The inferred return type stays `Type`, which check.cjs (b) shows on both
trees.

No public declaration moves: the function, the field and createTypeParameter are all
internal to createTypeChecker or `@internal`.

- `adapt.cjs` finds the one function declaration of that name anywhere in
  checker.ts, refuses before any write unless its text is the pinned before or after
  (SHA-256 and exact text), keeps CRLF, and makes no write on replay.
- `contract.cjs TREE` checks LF and CRLF, a zero-edit replay that keeps bytes and
  mtime, and refusal of a renamed, duplicated or drifted owner (three drifts, each a
  one-token change), with the file untouched.
- `check.cjs BEFORE AFTER` is run on two complete applies, before without this
  directory. It checks (a) that only this owner differs under src, and (b) that stock
  6.0.3 reports the same diagnostics over the 79 src/compiler roots and the same
  return type for the owner.
