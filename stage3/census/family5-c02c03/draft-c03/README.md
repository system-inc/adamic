# Type constraint slot (C03, #pncjpqy)

Type-only adaptation for TypeScript v6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8. Adds exactly this internal member to `Type` in `src/compiler/types.ts`:

```ts
/** @internal */ constraint?: Type;
```

This follows adaptation 72's pattern: the compiled program's relation gates certify the views; it does not weaken optional-widening or assert absence. Scout 464502ff, C03/G03 and G12, identifies 14 first-error sites, all in checker.ts. Type is public, but this new member is `@internal` and is stripped from the public typescript.d.ts. The candidate is stacked on C02 and is measured on C02's tree.

The compiler allocator's Type constructor (utilities.ts:8491), createType (checker.ts:5521) and createTypeParameter (checker.ts:5575) establish these internal shapes. Every constraint write on a Type stores a Type: the marker types, thisType, inferred and instantiated type-parameter constraints, noConstraintType, substitution constraints, restrictive instantiation and synthetic parameters. The audited writes are checker.ts:2172, :2177, :13504, :16855, :16860, :16869, :17258, :17927, :20682 and :22949. AST TypeParameterDeclaration.constraint is a separate TypeNode field on AST nodes, not this Type field.

Escapes include TypeParameter, InterfaceType, substitution/thisType views, cache fields, arrays, callback parameters, services TypeObject and public type relations. They are not claimed to satisfy adaptation 70's no-write audit; adaptation 72's compiled-program certificate applies. A source carrying an incompatible constraint remains subject to the ordinary relation gate. This declaration describes the existing values; it changes neither JavaScript nor constructors.

Pins were searched in every adaptation's adapt.cjs, owners.json and targets.json. No complete Type or complete types.ts pin exists. Earlier adaptations pin particular other owners or members; this new adaptation pins the complete Type interface before and after with SHA-256 and exact text.

Unknown text, duplicate owners, half adaptations and mixed line endings are refused before writing. CRLF is preserved. Both commented and comment-free JavaScript emits must be byte-identical before writing. Replay writes nothing. mutant.cjs restores the exact original Type owner while retaining C02, and is not part of normal apply.

All commands, child exit receipts, complete census streams, reached stops, public API comparison and lane receipts belong on typescript/family5-c02c03-evidence, separate from this landing directory.
