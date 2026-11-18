# Type constraint cache (#pncjpqy)

Family 5 candidate C03 declares `/** @internal */ constraint?: Type;` on
`src/compiler/types.ts`'s `Type`. `InterfaceType` inherits the member.
This was the proposed compiler-internal cache shape under the delegation. The refusal scan rejected the candidate.
[Family 5 scout, C03](https://github.com/system-inc/adamic/blob/464502ff430b1c88b2e79ebf6b21badcb09e02af/stage3/typescript/family5-scout/REPORT.md#c03--g03-g12)

The allocator's Type constructor supplies the objects used by `Type` and
`TypeParameter` (checker.ts:5521/5575 and utilities.ts:8491 in the scout).
Every write to the constraint cache stores a `Type`, including the parameter,
substitution and this-type caches at checker.ts:16855–16869, 17258 and 17927.
The AST's `TypeParameterDeclaration.constraint` is a separate `TypeNode` field.

This follows adaptation 72's pattern. Adamic judges each escape at the
relation site. G03 names escapes from `Type` to `TypeParameter` through the
arguments in `getBaseConstructorTypeOfClass`, `getImplementsTypes`,
`createSignatureTypeMapper`, `structuredTypeRelatedToWorker`,
`instantiateSignatureInContextOf`, `typeToTypeNodeWorker`,
`createMappedTypeNodeFromType` and `attachSymbolToLeftmostIdentifier`, the
return in `getDeclaredTypeOfTypeParameter`, and the assignment of
`append(result, getDeclaredTypeOfSymbol(symbol))`. G12 names the
`InterfaceType` to `TypeParameter` assignments in
`getTypeOfPropertyInBaseClass` and `checkSuperExpression`.

`owners.json` pins the whole interface before and after with SHA-256.
Adaptation 75 owns three other cache members on the same interface. The
adapter rejects owner drift, half-adapted text and mixed line endings,
preserves CRLF, and checks byte equality of emitted JavaScript with and
without comments before writing. A replay writes zero files. The mutant
restores the original owner.
