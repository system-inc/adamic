Adamic ruled this adaptation on Nov 24 2026 (#bbssdcy, the scout pass's proposal 4, family 1 E3 with casts cause 1c), and Compiler chose its shape on Nov 24 at 18:04Z (road 1); Compiler accepts the head after its review and census.

In the fully adapted base, `ReadonlyPragmaMap.forEach` at
`src/compiler/types.ts:10468:5` is deleted, so the interface inherits
`ReadonlyMap.forEach`. `PragmaMap.forEach` at `:10481:5` is replaced
with the nongeneric `Map` signature over the written out value union
from its extends clause. Round 3's two deletions produced TS2320 because
the inherited callbacks take different map types as their third
parameter. The explicit `PragmaMap` member resolves that collision
while preserving both extends clauses.

Both interfaces carry `@internal`. The published TypeScript 6.0.3
`typescript.d.ts` contains neither name. The stock 6.0.3 entry program
with the meter options reports zero diagnostics on base and on the
exact in memory candidate.

The only reader is `processPragmasIntoFields` at `parser.ts:10611:22`
on a `PragmaMap`. Its callback previously received
`PragmaPseudoMap[TKey] | PragmaPseudoMap[TKey][]` and `TKey`.
It now receives the map's value union and `string`. The resolved
signature takes a callback with value, key and `Map<string, V>` inputs
and an optional `thisArg?: any`, with `V` the same value union.

The generic `get` and `set` methods stay in place under Compiler's
terms. Loom's paired census measures E3 and the 1c cast at
`parser.ts:10585:23` with those methods present. The parser source,
including the cast and the callback body, retains its bytes.

Rows 12 and 130 of `docs/rulings.md` govern callback signature casts and
proof of generic bodies for every type argument.

The adapter finds each `forEach` member by its property name, refuses a computed member name, and `test.cjs` pins this behavior.

The census by unit at main ef741d4f, both passes (J-pgvfbyj entry, J-0n8wyn8 directory), moves one first stop with 0 units newly stopping and 0 cleared, and the move is an exposure: family 1 E3's unit `processPragmasIntoFields` goes from NotYet at `parser.ts:10611:31`, "a value of type PragmaPseudoMap[TKey][] | PragmaPseudoMap[TKey], a union whose member PragmaPseudoMap[TKey] has no representation: an unresolved indexed access", to Refused at `parser.ts:10619:25`, the `reference` case's upstream cast, "a cast the runtime can't check", owner #b5w3ycg; in the same unit the `amd-module` case's two casts at `:10655:47` and `:10659:43` go from NotYet to Refused, "a cast the runtime can't check" over a writable field `arguments` some member can't hold every value of, owner #b5w3ycg; and casts cause 1c's `processCommentPragmas` does not clear, still Refused at `:10585:23`, "a cast the runtime can't check".
