Adamic accepts this JavaScript-changing adaptation after its review and census (#bbssdcy, the tail pass's family 120, her Nov 22 21:49Z Map ruling for the sibling table).

The private table is allocated once per `transformSystemModule` call. Its two closures are `transformSourceFile` (one writer, creating the identifier once) and `onEmitNode` (one lookup, using the original node ID). An absent key returns undefined. No table reference observes length, order, holes, or the table object.

Ruling 165

> - **An assignment expression's value is its right side's value, and the right side's proofs (membership, narrowing, definedness) carry to the result.** Meaning. Oct 8 2026, @system_adamic; source: #bw3xg7c (0xj3wz00), #y5qmkr3 (item 5). Spec: What's in, Values and statements (narrowing by assignment).

Ruling 329

> - **Array writes are allowed only at an existing index; writing past the end, or at a negative or fractional index, panics in both backends; `push` appends; `length` can't be assigned.** Check, Refuse for `length`. Oct 5 2026, Kirk; source: #8vpywxp. Spec: Semantics that differ from JavaScript (Array writes). Test: internal/oracle/oracle_test.go:208 (`writes_past_end.a`); internal/conformance/testdata/checks/array_write_negative.ts, array_write_past_end.a and array_write_fractional.ts (panics, Nov 3); `length` assignment NONE (not found).

| Adapted `system.ts` | Owner | Symbol reference |
|---|---|---|
| 164 | `transformSystemModule` | Allocates the table |
| 212 | `transformSourceFile` | Writes once with `getOriginalNodeId(node)`, before `createSystemModuleBody` |
| 1776 | `onEmitNode` | Reads by `getOriginalNodeId(node)` before calling `previousOnEmitNode` |

The table entries reach two functions: `transformSourceFile` has one write and `onEmitNode` has one read. The writer runs when a source file is transformed. For that file, the emit notification reads the previously stored identifier. An unwritten ID reads as `undefined`, and a repeated write replaces only that ID.

The active identifier reaches four consumers. `transformSourceFile` passes it once to `createParameterDeclaration` at line 224. `createSystemModuleBody` reads it twice to create the module name at lines 373 and 374. `visitImportCallExpression` reads it once at line 1630. `substituteMetaProperty` reads it once at line 1976. Both writes set the active identifier to the exact object supplied by the factory or retrieved from the table. The existing context resets remain in place.
