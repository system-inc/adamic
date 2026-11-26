Adamic accepts this JavaScript-changing adaptation after its review and census (#bbssdcy, the tail pass's family 120, her Nov 22 21:49Z Map ruling for the sibling table).

The private table is allocated once per `transformSystemModule` call. Its two closures are `transformSourceFile` (one writer, creating the identifier once) and `onEmitNode` (one lookup, using the original node ID). An absent key returns undefined. No table reference observes length, order, holes, or the table object.

Ruling 165

> - **An assignment expression's value is its right side's value, and the right side's proofs (membership, narrowing, definedness) carry to the result.** Meaning. Oct 8 2026, @system_adamic; source: #bw3xg7c (0xj3wz00), #y5qmkr3 (item 5). Spec: What's in, Values and statements (narrowing by assignment).

Ruling 329

> - **Array writes are allowed only at an existing index; writing past the end, or at a negative or fractional index, panics in both backends; `push` appends; `length` can't be assigned.** Check, Refuse for `length`. Oct 5 2026, Kirk; source: #8vpywxp. Spec: Semantics that differ from JavaScript (Array writes). Test: internal/oracle/oracle_test.go:208 (`writes_past_end.a`); internal/conformance/testdata/checks/array_write_negative.ts, array_write_past_end.a and array_write_fractional.ts (panics, Nov 3); `length` assignment NONE (not found).

| Adapted `system.ts`, before this head | After this head | Owner | Symbol reference |
|---|---|---|---|
| 164 | 164 | `transformSystemModule` | Allocates the table |
| 212 | 212 to 214 | `transformSourceFile` | Writes once with `getOriginalNodeId(node)`, before `createSystemModuleBody` |
| 1776 | 1778 | `onEmitNode` | Reads by `getOriginalNodeId(node)` before calling `previousOnEmitNode` |

The table entries reach two functions: `transformSourceFile` has one write and `onEmitNode` has one read. The writer runs when a source file is transformed. For that file, the emit notification reads the previously stored identifier. An unwritten ID reads as `undefined`, and a repeated write replaces only that ID.

The active identifier reaches four consumers. `transformSourceFile` passes it once to `createParameterDeclaration` at line 224. `createSystemModuleBody` reads it twice to create the module name at lines 373 and 374. `visitImportCallExpression` reads it once at line 1630. `substituteMetaProperty` reads it once at line 1976. Both writes set the active identifier to the exact object supplied by the factory or retrieved from the table. The existing context resets remain in place. These lines are before this head; after it each sits two lines lower (226, 375 and 376, 1632, 1978).

`33-indexed-reads-emit/verify-system-context.cjs` pins this table's text from before this head, `contextObjectMap[id] = factory.createUniqueName("context")` in its producers table (`:72`) and the map name at `:73` that its `onEmitNode` row (`:83`) reads as `contextObjectMap[id]`, which main no longer has after this head, as it already hasn't for `moduleInfoMap` since 47 landed; only 33's own `system-type-mutants.cjs` (`:58`, `:74`) runs it, on trees 33 builds, and no lane test reads it.

Exposed stop: with `:212:25` cleared, tail family 120's one unit, `transformSystemModule`, stops next at `transformers/module/system.ts:384:9`, "overload 1 of visitNode parameter node cannot be served by implementation parameter node" (meter row 13, the rank 2 family "overload n of name parameter p cannot be served by implementation parameter q"), the `visitNode` call on `moduleInfo.externalHelpersImportDeclaration` in `createSystemModuleBody`, in upstream's bytes, two lines lower than stock `:382` after this head's write, and Apple's `99-apple-visit-node-parameter` head is its road (census by unit J-tkegpeh and J-z7hyqr3 at `25f3ed7f`; Adamic read it as an exposure at 2026-11-26 21:34:24Z on #4939pv9).
