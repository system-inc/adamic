# Flatten JSDoc parameter node lists

Family 44, core.ts:406:47 from utilities.ts:6917:1 and utilitiesPublic.ts:1324:1. Adamic reads this new shape under her Nov 22 ruling on #bbssdcy and her Nov 23 07:21Z answer, “The flatten helper: (i) one head,” with the U extends Node bound accepted and the worker move conditional on the named first stops. See docs/rulings.md rows 68, 141, 142, 158, 397 and 398.

The getJSDocTypeParameterDeclarations callback and both getEffectiveTypeParameterDeclarations callbacks return NodeArray<TypeParameterDeclaration> or undefined. The helper keeps increasing indexes, one callback per input, addRange for each present list and the shared empty result.

Compiler-scope replay at ac44317c selected utilities.ts:6917:1 and utilitiesPublic.ts:1324:1. Both base units first stop at core.ts:406:47. The two head units were attempted and neither first stop requires moving getJSDocTagsWorker. The utilities.ts owner has no first stop. The utilitiesPublic.ts owner, now at :1325:1, exposes “a logical operand held as value where its result is held as boolean | undefined” at :1342:9 on the existing canHaveIllegalTypeParameters(node) && node.typeParameters test. A full selected-unit replay also reproduces that same refusal later in the base unit, at utilitiesPublic.ts:1341:9. This is the logical-operand storage family; the committed Nov 22 reason list does not assign it a numbered family.

The worker stays on generic flatMap and keeps its existing allocation path. All other generic callers remain on flatMap, including checker.ts and filterOwnedJSDocTags.

Run node test.cjs <typescript package> <adapted tree> <base|head>. It splices these functions by their syntax trees into the stock 6.0.3 bundle. The tests check four callback indexes and exact node identity, shared empty output, callback throw order, and direct template, overload and typedef parameter lists. The scratch control skips the second mapped list and must fail the four-callback test with length 1 instead of 3.

The JavaScript diff and test output are in the patch evidence.
