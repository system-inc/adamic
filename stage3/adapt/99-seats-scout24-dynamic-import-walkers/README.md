# Dynamic import and require walkers

Adamic ruled this adaptation on Nov 24 2026 (#bbssdcy, the scout pass's proposal 23, casts cause 6), family 85's twin shape; Compiler accepts the head after its review and census.

The transformer callers move to the value space walker. Its scanner excludes
JSDoc and accepts nonliteral arguments. The program keeps its existing call
and literal true flags. Its walker retains literal require and import calls,
literal import types, and JSDoc import tags. The two literal tests and both value space reads use
four `!` checks in `30-indexed-reads`' form, verified by stock tsc with
Adamic's loader options, including noUncheckedIndexedAccess. The require arm
tests the literal after `isRequireCall`'s boolean overload because its `true`
overload narrows `arguments` to an array and tuple intersection that Adamic's
lowering cannot take (review-v's read 2).

The new internal runtime key on `ts` is
`forEachDynamicImportOrRequireCallExpression`. It is hidden by stripInternal
under rows 406 and 408. Rows 12 and 130 govern the callback assertions and
the rigid generic body. Family 19's dispatch (proposal 17) shares the
erasure of the correlation between a runtime flag and a type parameter.

Value identity is claimed for tsc's own program and for callers within the
public declarations: on tsc's parsed nodes, stock and adapted hand out the
same objects in the same order, read them in the same order, and throw the
same errors at the same points. Three differences are reachable only by a
caller outside them, and Compiler read all three as outside the contract
under Adamic's Nov 23 rule (#0gfgm5r, Nov 25): a getter or Proxy on a node a
caller hands in sees two more reads of `node.arguments` and `[0]` on a
one-argument `require` in a JavaScript file, and can throw at a read stock
never makes; a direct call to the undeclared runtime member
`forEachDynamicImportOrRequireCall` with `false, false` on `require(name)`
calls back once in stock and never here; and `ts` gains the one runtime key
named above.

Run the witness with `NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/99-seats-scout24-dynamic-import-walkers/test.cjs <applied tree>`.
It reads this adapter's before and after Git snapshots. It compares callback
tuples and the scanner's JSDoc flag at each position, which can differ without
a value callback when the JSDoc branch contains only type space nodes.

## Resolved calls

Locations are in the fully adapted base. Signatures use stock TypeScript 6.0.3.

| Call | Flags | Before | After | Disposition |
| --- | --- | --- | --- | --- |
| `src/compiler/program.ts:3368:13` | `true, true` | `(file: SourceFile, includeTypeSpaceImports: true, requireStringLiteralLikeArgument: true, cb: (node: ImportTypeNode \| CallExpression \| JSDocImportTag, argument: StringLiteralLike) => void): void` | `(file: SourceFile, includeTypeSpaceImports: true, requireStringLiteralLikeArgument: true, cb: (node: ImportTypeNode \| CallExpression \| JSDocImportTag, argument: StringLiteralLike) => void): void` | Call text retained |
| `src/compiler/transformers/module/esnextAnd2015.ts:98:17` | `false, false` | `(file: SourceFile, includeTypeSpaceImports: false, requireStringLiteralLikeArgument: false, cb: (node: CallExpression, argument: Expression) => void): void` | `(file: SourceFile, cb: (node: CallExpression, argument: Expression) => void): void` | Redirected with identical callback text |
| `src/compiler/transformers/module/module.ts:245:13` | `false, false` | `(file: SourceFile, includeTypeSpaceImports: false, requireStringLiteralLikeArgument: false, cb: (node: CallExpression, argument: Expression) => void): void` | `(file: SourceFile, cb: (node: CallExpression, argument: Expression) => void): void` | Redirected with identical callback text |
