# Primitive cast retypes

Adamic ruled these adaptations on Nov 3 2026 (#b5w3ycg, the primitive targets); Compiler accepts this type-only head under the split, after its review and census.

- `utilities.ts:3444:14`, `getEntityNameFromTypeNode`, and its adjacent case at `3445:14`. Its internal parameter admits EntityName alongside TypeNode, as the existing TODO describes. The two case labels narrow directly. Stock tsc emits `return node;` from both the original assertion and the new return. The four callers keep their original arguments.

The resolver retype was removed because the reader reproduced the refusal from `Extension | undefined`: primitive targets are refused before the lowering checks whether the cast drops `undefined`.

The primitive rule at `docs/rulings.md:11`, the undefined-removal rule at `:30`, and the temporary adaptation rule at `:410` govern this change. The adapter requires the owner's `@internal` JSDoc tag on both inputs and pins the full before and after AST owner text. Stock JavaScript and declarations emitted with `stripInternal` stay byte-identical.
