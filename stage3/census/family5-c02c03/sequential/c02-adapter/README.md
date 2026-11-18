# Node line and character slot (C02, #pncjpqy)

Type-only adaptation for TypeScript v6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8. Adds exactly this internal member to `Node` in `src/compiler/types.ts`:

```ts
/** @internal */ getPositionOfLineAndCharacter?: ((line: number, character: number, allowEdits?: true) => number) | undefined;
```

This follows adaptation 72's pattern: the compiled program's relation gates certify the views; this is not an absence assertion or an unchecked exemption. Scout 464502ff, C02/G02, identifies 14 first-error sites. Node is public, but this member is `@internal` and is stripped from the public typescript.d.ts.

Compiler SourceFiles are constructed through parser.createSourceFile, the base node factory, and the compiler's objectAllocator Node constructor; they omit this method. Services' `SourceFileObject` implements the matching callable, including `allowEdits?: true`. Scanner's getPositionOfLineAndCharacter worker calls it only after testing its presence (scout scanner.ts:468; current source coordinates are recorded with evidence). Thus the shared Node declaration states the established internal shape without changing runtime construction or calls.

Escapes include Node's AST subtypes, factory create/clone/update operations, SourceFileLike views, and service/API boundaries. No claim is made that these escapes satisfy adaptation 70's no-write audit: adaptation 72's compiled-program certificate applies. Structural sources that really carry an incompatible callable remain subject to the ordinary relation gates.

Pins were searched in every adaptation's adapt.cjs, owners.json and targets.json. Adaptation 80 pins Node's heritage only; 71 pins individual members. Neither pins this complete interface. There is no existing whole Node or whole types.ts pin, so this is a new adaptation, matching the ownership distinction in 99-correctness-sourcefile-like-position.

owners.json pins the complete Node interface before and after with SHA-256 and exact text. The adapter rejects drift, duplicate owners, half adaptations and mixed line endings, and preserves CRLF. Before writing it checks both commented and comment-free JavaScript are byte-identical. Replay writes nothing. mutant.cjs restores the exact original owner and is not part of normal apply.

Proof commands, actual child exits, complete census streams, reached stops, public API comparison and lane receipts belong on typescript/family5-c02c03-evidence, separate from this landing directory.
