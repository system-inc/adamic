# Family 85: getOriginalNodeWithoutPredicate

Adamic ruled this adaptation's shape on Nov 23 2026 (#bbssdcy, family 85); she accepts the head after its review and census.

In her Nov 23 05:57Z comment in #bbssdcy, Adamic confirmed that the separate
internal helpers with redirected callers are the ruled shape. The original-node
helper carries row 398's two internal overloads. She approved the runtime keys
`changeExtensionString` and `getOriginalNodeWithoutPredicate` on `ts` as internal
exports that the `stripInternal` declaration graph hides. She waived row 398's
SDK probe for family 85. The `stripInternal` emit and the lane's `api_check`
stand in.

It applies rulings 12, 14,
127, 397 and 398 against Adamic 0efbdb083a60086fec3a29687a162bbf21105336.
The public and generic declarations stay byte-identical in source. The twin is
internal. The following positions are in the fully adapted baseline.

| Call | Stock TypeScript 6.0.3 resolved signature | Disposition |
| --- | --- | --- |
| `src/compiler/checker.ts:6526:160` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/checker.ts:6542:88` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/checker.ts:8630:42` | `(node: Node \| undefined): Node \| undefined` | Moved (no predicate) |
| `src/compiler/checker.ts:8680:57` | `(node: Node \| undefined): Node \| undefined` | Moved (no predicate) |
| `src/compiler/emitter.ts:5047:40` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/emitter.ts:5047:79` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/emitter.ts:5253:68` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/emitter.ts:5258:68` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/emitter.ts:5955:17` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/emitter.ts:5958:49` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/emitter.ts:5966:24` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/emitter.ts:5967:20` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/factory/utilities.ts:685:23` | `(node: Node, nodeTest: (node: Node) => node is SourceFile): SourceFile` | Left (predicate) |
| `src/compiler/factory/utilities.ts:692:23` | `(node: Node, nodeTest: (node: Node) => node is SourceFile): SourceFile` | Left (predicate) |
| `src/compiler/factory/utilities.ts:729:39` | `(node: Node, nodeTest: (node: Node) => node is SourceFile): SourceFile` | Left (predicate) |
| `src/compiler/factory/utilities.ts:776:27` | `(node: Node, nodeTest: (node: Node) => node is SourceFile): SourceFile` | Left (predicate) |
| `src/compiler/transformers/classFields.ts:1092:47` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/transformers/classFields.ts:1433:39` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/transformers/classFields.ts:1713:26` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/transformers/classFields.ts:1749:43` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/transformers/classFields.ts:2298:52` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/transformers/classFields.ts:2349:107` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/transformers/classFields.ts:2350:103` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/transformers/classFields.ts:2377:56` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/transformers/classFields.ts:2463:38` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/transformers/classFields.ts:2528:39` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/transformers/classFields.ts:2584:38` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/transformers/classFields.ts:3180:26` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/transformers/es2015.ts:610:30` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/transformers/es2017.ts:660:32` | `(node: Node, nodeTest: (node: Node) => node is FunctionLikeDeclaration): FunctionLikeDeclaration` | Left (predicate) |
| `src/compiler/transformers/es2017.ts:741:26` | `(node: Node, nodeTest: (node: Node) => node is SignatureDeclaration): SignatureDeclaration` | Left (predicate) |
| `src/compiler/transformers/esDecorators.ts:1058:35` | `(node: Node, nodeTest: (node: Node) => node is ClassLikeDeclaration): ClassLikeDeclaration` | Left (predicate) |
| `src/compiler/transformers/generators.ts:2068:30` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/transformers/jsx.ts:349:34` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/transformers/module/system.ts:942:21` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/transformers/namedEvaluation.ts:59:22` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/transformers/ts.ts:2598:16` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/transformers/ts.ts:2602:16` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/transformers/ts.ts:2730:38` | `(node: Node, nodeTest: (node: Node) => node is AccessExpression): AccessExpression` | Left (predicate) |
| `src/compiler/transformers/utilities.ts:110:12` | `(node: Node): Node` | Moved (no predicate) |
| `src/compiler/transformers/utilities.ts:335:31` | `(node: Node, nodeTest: (node: Node) => node is FunctionDeclaration): FunctionDeclaration` | Left (predicate) |
| `src/services/codefixes/convertToAsyncFunction.ts:404:26` | `(node: Node): Node` | Moved (no predicate) |
| `src/services/codefixes/convertToAsyncFunction.ts:870:30` | `(node: Node): Node` | Left (other declaration) |
| `src/services/textChanges.ts:1279:108` | `(node: Node): Node` | Moved (no predicate) |
| `src/services/textChanges.ts:1280:90` | `(node: Node): Node` | Moved (no predicate) |
| `src/services/utilities.ts:2627:85` | `(node: Node): Node` | Moved (no predicate) |
