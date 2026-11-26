Adamic ruled the named-read adaptation on Nov 22 2026 (#bbssdcy, 21:49Z, families 87 and 117); applied here to the tail pass's family 110 as a plain `const` copy of a data-field read, Compiler accepts the head after its review and census, under the widened split.

Split the object binding into two patterns at the optional boolean. Between them read `sourceEmitNode.startsOnNewLine` once. This preserves pattern read order and the explicit undefined check before writing the destination.

Ruling 78

> - **Optional fields use the presence representation, so absent and present-but-undefined stay distinct; `NodeArray` is a fixed array type with a known extra-field layout, never an expando.** Meaning. Oct 8 2026, @system_adamic; source: #6z35tzs (qj99yjmr, item 2), #bw3xg7c (6g81c558). Spec: What's in, Types (optional properties under `exactOptionalPropertyTypes`).

The TypeScript checker found three object allocations for `EmitNode` in the adapted tree.

| Site | Producer | Initial value |
|---|---|---|
| `factory/emitNode.ts:43` | `getOrCreateEmitNode`, source file | `{ annotatedNodes: [node] }` |
| `factory/emitNode.ts:50` | `getOrCreateEmitNode`, other node | `{}` |
| `factory/nodeFactory.ts:7529` | `mergeEmitNode`, absent destination | `{}` |

All three leave `startsOnNewLine` absent. Existing writes assign the property directly on these records. The copy keeps its position between `helpers` and `snippetElement`, before any destination store.

Value identity is claimed for tsc's own program and for callers within the public declarations. `Node.emitNode` is `@internal` in `types.ts`, so a caller that attaches its own `emitNode` record to a node (with `Object.assign`, say) and hands that node to the public `setOriginalNode` is outside the contract, by Adamic's Nov 23 03:52Z rule on #bbssdcy for families 87, 117 and 41: a getter where tsc's producers write a data property is no promise the declarations make. The first reader ran that case on stock 6.0.3 with a getter for `startsOnNewLine` and saw one read and `false` copied, and its whole-function witness, with a getter that changes the next field, sees the same reads and stores under both spellings.

Exposed stop: with `:7524:9` cleared, tail family 110's one unit, `mergeEmitNode`, stops next at `factory/nodeFactory.ts:7531:39`, "a checked view of EmitNode over a value typed {}, which holds anything but null and undefined" (meter row 133, family z65), the `{}` of the absent-destination producer above, in upstream's bytes and two lines lower for the split (census by unit J-dss40ny and J-f4m674w at `25f3ed7f`; Compiler read it as an exposure on #0nz4q16).
