# Token source-map ranges as a record

This applies Adamic's Nov 22 ruling on #bbssdcy for family 163, the string-keyed record, and her Nov 23 answer, the getter's undefined guard. It is a rewrite of the shape she ruled, applied exactly, so Compiler accepts it under the widened split after its review and census (Adamic, Nov 23 07:21Z).

A `null` written to `node.emitNode.tokenSourceMapRanges` by untyped code, which the original optional index would have short-circuited and the record lookup now throws on, is outside the ruling (Adamic, Nov 23 07:21Z): no checked writer allocates or assigns `null`, and the property is not in the stripped API.

Family 163 stops at the source table in the `for...in` of
`factory/nodeFactory.ts:7541:23` (`mergeTokenSourceMapRanges`). The table is
sparse by token kind. Both producers allocate an empty record. The three
token-kind accesses use `String(token)`, the decimal property key Node uses for
these numeric enum values. The merge keeps its keyed loop and updates the
existing destination in place.

Rows 188 and 283 of `docs/rulings.md` govern own enumerable key order, row 300
governs own-key records and dynamic reads, and row 405 preserves an entry whose
value is undefined. Row 320 governs native array writes. The original setter
can grow a Node array at an unwritten index, which Adamic's array writes refuse
at runtime.

## Readers and writers

These are stock TypeScript 6.0.3 checker references in the fully adapted base tree.

| Adapted base location | Use and Node behavior |
|---|---|
| `types.ts:8361` | The optional internal `EmitNode.tokenSourceMapRanges` property is declared as `(SourceMapRange \| undefined)[]`. The store can be absent, and an allocated store can have holes. |
| `factory/emitNode.ts:156 to 157` | The setter reads the table, initializes it if absent, and writes the token key. A write of undefined creates an own key, an overwrite replaces its value, and other unwritten indices stay absent. |
| `factory/emitNode.ts:148` | The getter reads the latest value at the token key. Both a hole and a present undefined return undefined. |
| `emitter.ts:6274` | The emitter reads the latest value at the token key. Both a hole and a present undefined take its existing fallback. |
| `factory/nodeFactory.ts:7442,7490 to 7491` | `mergeEmitNode` reads the source property and the current destination, then assigns the helper's returned table. The guard is on the source table's presence. |
| `factory/nodeFactory.ts:7539 to 7544` | The private helper allocates a missing destination, enumerates the source's present keys in ascending index order and copies their current values. A source key holding undefined overwrites the destination. An unwritten source key is skipped, retaining any destination entry at that key. |

The source and destination references use no array method, length, spread, or
`for...of`. The emit node is internal and the public TypeScript 6.0.3
declaration contains no token table property.

The Nov 23 ruling gives the getter an explicit undefined guard. It evaluates
the optional property chain once and converts the key only when the record
exists. The open-record optional-index refusal is tracked separately as a
Compiler Should beside family 68.
