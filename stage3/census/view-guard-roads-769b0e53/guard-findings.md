# guard: every new first-error guard finding

Reasons and type names are verbatim. Source slices retain their text, with CRLF displayed as LF. Repeated observations are grouped only when site, source, target, reason and source text agree; every owner unit is listed. Road is the first new mechanism in the actual path; the exact reason preserves every subsequent mechanism.

## elements

### guard-0001: src/compiler/symbolWalker.ts:88:36

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Type as ObjectType
~~~~

Exact reason:

~~~~text
a view of Type as ObjectType in a graph program, which the cycle finder doesn't relate yet through its field symbol (Symbol as Symbol) through its field declarations (Declaration[] | undefined as Declaration[] | undefined) through its elements (Declaration as Declaration) through its field localSymbol (Symbol | undefined as Symbol | undefined) through its field valueDeclaration (Declaration | undefined as Declaration | undefined) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
type as ObjectType
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/symbolWalker.ts:30:1`

### guard-0002: src/compiler/utilities.ts:8039:52

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Symbol as TransientSymbol
~~~~

Exact reason:

~~~~text
a view of Symbol as TransientSymbol in a graph program, which the cycle finder doesn't relate yet through its field declarations (Declaration[] | undefined as Declaration[] | undefined) through its elements (Declaration as Declaration) through its field symbol (Symbol as Symbol) through its field valueDeclaration (Declaration | undefined as Declaration | undefined) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
symbol as TransientSymbol
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:8038:1`

## index values

### guard-0003: src/compiler/binder.ts:359:29

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as EnumDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as EnumDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as EnumDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/binder.ts:350:1`

### guard-0004: src/compiler/checker.ts:54229:12

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Declaration as FunctionDeclaration
~~~~

Exact reason:

~~~~text
a view of Declaration as FunctionDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
declaration as FunctionDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/checker.ts:54227:1`

### guard-0005: src/compiler/core.ts:1787:32

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Expression as LeftHandSideExpression
~~~~

Exact reason:

~~~~text
a view of Expression as LeftHandSideExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
test(value)
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/factory/parenthesizerRules.ts:675:1`

### guard-0006: src/compiler/emitter.ts:1158:18

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Declaration as LateBoundDeclaration
~~~~

Exact reason:

~~~~text
a view of Declaration as LateBoundDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
(_node): _node is LateBoundDeclaration => false
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/emitter.ts:1147:1`

### guard-0007: src/compiler/factory/nodeChildren.ts:20:17

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as SyntaxList
~~~~

Exact reason:

~~~~text
a view of Node as SyntaxList in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as SyntaxList
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/factory/nodeChildren.ts:14:1`

### guard-0008: src/compiler/factory/nodeTests.ts:1003:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as SpreadAssignment
~~~~

Exact reason:

~~~~text
a view of Node as SpreadAssignment in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ObjectLiteralExpression) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isSpreadAssignment(node: Node): node is SpreadAssignment {
    return node.kind === SyntaxKind.SpreadAssignment;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/factory/utilities.ts:1008:1`

### guard-0009: src/compiler/factory/nodeTests.ts:1014:1

Observations: {'head': 4}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as SourceFile
~~~~

Exact reason:

~~~~text
a view of Node as SourceFile in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
// Top-level nodes
export function isSourceFile(node: Node): node is SourceFile {
    return node.kind === SyntaxKind.SourceFile;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/module/impliedNodeFormatDependent.ts:18:1`
- `src/compiler/utilities.ts:2135:1`
- `src/compiler/utilities.ts:5071:1`
- `src/compiler/utilitiesPublic.ts:1727:1`

### guard-0010: src/compiler/factory/nodeTests.ts:1026:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDocTypeExpression
~~~~

Exact reason:

~~~~text
a view of Node as JSDocTypeExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
// TODO(rbuckton): isInputFiles

// JSDoc Elements

export function isJSDocTypeExpression(node: Node): node is JSDocTypeExpression {
    return node.kind === SyntaxKind.JSDocTypeExpression;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:1291:1`

### guard-0011: src/compiler/factory/nodeTests.ts:1030:1

Observations: {'head': 3}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDocNameReference
~~~~

Exact reason:

~~~~text
a view of Node as JSDocNameReference in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isJSDocNameReference(node: Node): node is JSDocNameReference {
    return node.kind === SyntaxKind.JSDocNameReference;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:3629:1`

### guard-0012: src/compiler/factory/nodeTests.ts:1034:1

Observations: {'head': 4}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDocMemberName
~~~~

Exact reason:

~~~~text
a view of Node as JSDocMemberName in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isJSDocMemberName(node: Node): node is JSDocMemberName {
    return node.kind === SyntaxKind.JSDocMemberName;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:3629:1`

### guard-0013: src/compiler/factory/nodeTests.ts:1070:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDocFunctionType
~~~~

Exact reason:

~~~~text
a view of Node as JSDocFunctionType in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isJSDocFunctionType(node: Node): node is JSDocFunctionType {
    return node.kind === SyntaxKind.JSDocFunctionType;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:4473:1`

### guard-0014: src/compiler/factory/nodeTests.ts:1096:1

Observations: {'head': 4}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDocAugmentsTag
~~~~

Exact reason:

~~~~text
a view of Node as JSDocAugmentsTag in a graph program, which the cycle finder doesn't relate yet through its field original (Node | undefined as Node | undefined) through its field parent (Node as Node) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
// JSDoc Tags

export function isJSDocAugmentsTag(node: Node): node is JSDocAugmentsTag {
    return node.kind === SyntaxKind.JSDocAugmentsTag;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2882:1`
- `src/compiler/utilities.ts:3629:1`
- `src/compiler/utilities.ts:7453:1`
- `src/compiler/utilitiesPublic.ts:1111:1`

### guard-0015: src/compiler/factory/nodeTests.ts:1104:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDocClassTag
~~~~

Exact reason:

~~~~text
a view of Node as JSDocClassTag in a graph program, which the cycle finder doesn't relate yet through its field original (Node | undefined as Node | undefined) through its field parent (Node as Node) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isJSDocClassTag(node: Node): node is JSDocClassTag {
    return node.kind === SyntaxKind.JSDocClassTag;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1121:1`

### guard-0016: src/compiler/factory/nodeTests.ts:1112:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDocPublicTag
~~~~

Exact reason:

~~~~text
a view of Node as JSDocPublicTag in a graph program, which the cycle finder doesn't relate yet through its field original (Node | undefined as Node | undefined) through its field parent (Node as Node) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isJSDocPublicTag(node: Node): node is JSDocPublicTag {
    return node.kind === SyntaxKind.JSDocPublicTag;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1126:1`
- `src/compiler/utilitiesPublic.ts:1131:1`

### guard-0017: src/compiler/factory/nodeTests.ts:1116:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDocPrivateTag
~~~~

Exact reason:

~~~~text
a view of Node as JSDocPrivateTag in a graph program, which the cycle finder doesn't relate yet through its field original (Node | undefined as Node | undefined) through its field parent (Node as Node) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isJSDocPrivateTag(node: Node): node is JSDocPrivateTag {
    return node.kind === SyntaxKind.JSDocPrivateTag;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1136:1`
- `src/compiler/utilitiesPublic.ts:1141:1`

### guard-0018: src/compiler/factory/nodeTests.ts:1120:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDocProtectedTag
~~~~

Exact reason:

~~~~text
a view of Node as JSDocProtectedTag in a graph program, which the cycle finder doesn't relate yet through its field original (Node | undefined as Node | undefined) through its field parent (Node as Node) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isJSDocProtectedTag(node: Node): node is JSDocProtectedTag {
    return node.kind === SyntaxKind.JSDocProtectedTag;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1146:1`
- `src/compiler/utilitiesPublic.ts:1151:1`

### guard-0019: src/compiler/factory/nodeTests.ts:1124:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDocReadonlyTag
~~~~

Exact reason:

~~~~text
a view of Node as JSDocReadonlyTag in a graph program, which the cycle finder doesn't relate yet through its field original (Node | undefined as Node | undefined) through its field parent (Node as Node) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isJSDocReadonlyTag(node: Node): node is JSDocReadonlyTag {
    return node.kind === SyntaxKind.JSDocReadonlyTag;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1156:1`
- `src/compiler/utilitiesPublic.ts:1161:1`

### guard-0020: src/compiler/factory/nodeTests.ts:1128:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDocOverrideTag
~~~~

Exact reason:

~~~~text
a view of Node as JSDocOverrideTag in a graph program, which the cycle finder doesn't relate yet through its field original (Node | undefined as Node | undefined) through its field parent (Node as Node) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isJSDocOverrideTag(node: Node): node is JSDocOverrideTag {
    return node.kind === SyntaxKind.JSDocOverrideTag;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1165:1`

### guard-0021: src/compiler/factory/nodeTests.ts:1136:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDocDeprecatedTag
~~~~

Exact reason:

~~~~text
a view of Node as JSDocDeprecatedTag in a graph program, which the cycle finder doesn't relate yet through its field original (Node | undefined as Node | undefined) through its field parent (Node as Node) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isJSDocDeprecatedTag(node: Node): node is JSDocDeprecatedTag {
    return node.kind === SyntaxKind.JSDocDeprecatedTag;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1170:1`
- `src/compiler/utilitiesPublic.ts:1175:1`

### guard-0022: src/compiler/factory/nodeTests.ts:1148:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDocParameterTag
~~~~

Exact reason:

~~~~text
a view of Node as JSDocParameterTag in a graph program, which the cycle finder doesn't relate yet through its field original (Node | undefined as Node | undefined) through its field parent (Node as Node) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isJSDocParameterTag(node: Node): node is JSDocParameterTag {
    return node.kind === SyntaxKind.JSDocParameterTag;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1106:1`

### guard-0023: src/compiler/factory/nodeTests.ts:1152:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDocReturnTag
~~~~

Exact reason:

~~~~text
a view of Node as JSDocReturnTag in a graph program, which the cycle finder doesn't relate yet through its field original (Node | undefined as Node | undefined) through its field parent (Node as Node) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isJSDocReturnTag(node: Node): node is JSDocReturnTag {
    return node.kind === SyntaxKind.JSDocReturnTag;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1190:1`

### guard-0024: src/compiler/factory/nodeTests.ts:1156:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDocThisTag
~~~~

Exact reason:

~~~~text
a view of Node as JSDocThisTag in a graph program, which the cycle finder doesn't relate yet through its field original (Node | undefined as Node | undefined) through its field parent (Node as Node) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isJSDocThisTag(node: Node): node is JSDocThisTag {
    return node.kind === SyntaxKind.JSDocThisTag;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1185:1`

### guard-0025: src/compiler/factory/nodeTests.ts:1160:1

Observations: {'head': 4}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDocTypeTag
~~~~

Exact reason:

~~~~text
a view of Node as JSDocTypeTag in a graph program, which the cycle finder doesn't relate yet through its field original (Node | undefined as Node | undefined) through its field parent (Node as Node) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isJSDocTypeTag(node: Node): node is JSDocTypeTag {
    return node.kind === SyntaxKind.JSDocTypeTag;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:11450:1`
- `src/compiler/utilities.ts:4699:1`
- `src/compiler/utilitiesPublic.ts:1204:1`
- `src/compiler/utilitiesPublic.ts:1224:1`

### guard-0026: src/compiler/factory/nodeTests.ts:1164:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDocTemplateTag
~~~~

Exact reason:

~~~~text
a view of Node as JSDocTemplateTag in a graph program, which the cycle finder doesn't relate yet through its field original (Node | undefined as Node | undefined) through its field parent (Node as Node) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isJSDocTemplateTag(node: Node): node is JSDocTemplateTag {
    return node.kind === SyntaxKind.JSDocTemplateTag;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:6921:1`
- `src/compiler/utilitiesPublic.ts:1195:1`

### guard-0027: src/compiler/factory/nodeTests.ts:1180:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDocImplementsTag
~~~~

Exact reason:

~~~~text
a view of Node as JSDocImplementsTag in a graph program, which the cycle finder doesn't relate yet through its field original (Node | undefined as Node | undefined) through its field parent (Node as Node) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isJSDocImplementsTag(node: Node): node is JSDocImplementsTag {
    return node.kind === SyntaxKind.JSDocImplementsTag;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2882:1`
- `src/compiler/utilitiesPublic.ts:1116:1`

### guard-0028: src/compiler/factory/nodeTests.ts:1184:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDocSatisfiesTag
~~~~

Exact reason:

~~~~text
a view of Node as JSDocSatisfiesTag in a graph program, which the cycle finder doesn't relate yet through its field original (Node | undefined as Node | undefined) through its field parent (Node as Node) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isJSDocSatisfiesTag(node: Node): node is JSDocSatisfiesTag {
    return node.kind === SyntaxKind.JSDocSatisfiesTag;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:4699:1`
- `src/compiler/utilitiesPublic.ts:1199:1`

### guard-0029: src/compiler/factory/nodeTests.ts:235:1

Observations: {'head': 6}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as NumericLiteral
~~~~

Exact reason:

~~~~text
a view of Node as NumericLiteral in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
// Literals

export function isNumericLiteral(node: Node): node is NumericLiteral {
    return node.kind === SyntaxKind.NumericLiteral;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:12031:1`
- `src/compiler/utilities.ts:2037:1`
- `src/compiler/utilities.ts:4200:1`
- `src/compiler/utilities.ts:5360:1`
- `src/compiler/utilities.ts:5365:1`

### guard-0030: src/compiler/factory/nodeTests.ts:239:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as BigIntLiteral
~~~~

Exact reason:

~~~~text
a view of Node as BigIntLiteral in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isBigIntLiteral(node: Node): node is BigIntLiteral {
    return node.kind === SyntaxKind.BigIntLiteral;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:12031:1`
- `src/compiler/utilities.ts:2037:1`

### guard-0031: src/compiler/factory/nodeTests.ts:243:1

Observations: {'head': 8}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as StringLiteral
~~~~

Exact reason:

~~~~text
a view of Node as StringLiteral in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isStringLiteral(node: Node): node is StringLiteral {
    return node.kind === SyntaxKind.StringLiteral;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/factory/utilities.ts:579:1`
- `src/compiler/factory/utilities.ts:814:1`
- `src/compiler/transformers/utilities.ts:876:1`
- `src/compiler/transformers/utilities.ts:878:1`
- `src/compiler/transformers/utilities.ts:880:1`
- `src/compiler/utilities.ts:2722:1`
- `src/compiler/utilities.ts:8288:1`
- `src/compiler/utilitiesPublic.ts:1574:1`

### guard-0032: src/compiler/factory/nodeTests.ts:284:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as PlusToken
~~~~

Exact reason:

~~~~text
a view of Node as PlusToken in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isPlusToken(node: Node): node is PlusToken {
    return node.kind === SyntaxKind.PlusToken;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/factory/utilities.ts:1159:1`
- `src/compiler/factory/utilities.ts:1163:1`

### guard-0033: src/compiler/factory/nodeTests.ts:288:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as MinusToken
~~~~

Exact reason:

~~~~text
a view of Node as MinusToken in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isMinusToken(node: Node): node is MinusToken {
    return node.kind === SyntaxKind.MinusToken;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/factory/utilities.ts:1159:1`
- `src/compiler/factory/utilities.ts:1163:1`

### guard-0034: src/compiler/factory/nodeTests.ts:296:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ExclamationToken
~~~~

Exact reason:

~~~~text
a view of Node as ExclamationToken in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isExclamationToken(node: Node): node is ExclamationToken {
    return node.kind === SyntaxKind.ExclamationToken;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/factory/utilities.ts:1151:1`

### guard-0035: src/compiler/factory/nodeTests.ts:300:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as QuestionToken
~~~~

Exact reason:

~~~~text
a view of Node as QuestionToken in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isQuestionToken(node: Node): node is QuestionToken {
    return node.kind === SyntaxKind.QuestionToken;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/factory/utilities.ts:1151:1`
- `src/compiler/factory/utilities.ts:1163:1`

### guard-0036: src/compiler/factory/nodeTests.ts:318:1

Observations: {'head': 20}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as Identifier
~~~~

Exact reason:

~~~~text
a view of Node as Identifier in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
// Identifiers

export function isIdentifier(node: Node): node is Identifier {
    return node.kind === SyntaxKind.Identifier;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/checker.ts:54233:1`
- `src/compiler/factory/emitHelpers.ts:1486:1`
- `src/compiler/factory/utilities.ts:1155:1`
- `src/compiler/factory/utilities.ts:1167:1`
- `src/compiler/transformers/destructuring.ts:583:1`
- `src/compiler/transformers/utilities.ts:471:1`
- `src/compiler/transformers/utilities.ts:485:1`
- `src/compiler/transformers/utilities.ts:866:1`
- `src/compiler/utilities.ts:10589:1`
- `src/compiler/utilities.ts:10625:1`
- `src/compiler/utilities.ts:2743:1`
- `src/compiler/utilities.ts:3281:1`
- `src/compiler/utilities.ts:3822:1`
- `src/compiler/utilities.ts:3951:1`
- `src/compiler/utilities.ts:4039:1`
- `src/compiler/utilities.ts:4092:1`
- `src/compiler/utilities.ts:4097:1`
- `src/compiler/utilitiesPublic.ts:1426:1`
- `src/compiler/utilitiesPublic.ts:1574:1`
- `src/compiler/utilitiesPublic.ts:1581:1`

### guard-0037: src/compiler/factory/nodeTests.ts:322:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as PrivateIdentifier
~~~~

Exact reason:

~~~~text
a view of Node as PrivateIdentifier in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isPrivateIdentifier(node: Node): node is PrivateIdentifier {
    return node.kind === SyntaxKind.PrivateIdentifier;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/classFields.ts:3349:1`
- `src/compiler/utilitiesPublic.ts:1586:1`

### guard-0038: src/compiler/factory/nodeTests.ts:352:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ReadonlyKeyword
~~~~

Exact reason:

~~~~text
a view of Node as ReadonlyKeyword in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isReadonlyKeyword(node: Node): node is ReadonlyKeyword {
    return node.kind === SyntaxKind.ReadonlyKeyword;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/factory/utilities.ts:1159:1`

### guard-0039: src/compiler/factory/nodeTests.ts:393:1

Observations: {'head': 3}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as QualifiedName
~~~~

Exact reason:

~~~~text
a view of Node as QualifiedName in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
// Names

export function isQualifiedName(node: Node): node is QualifiedName {
    return node.kind === SyntaxKind.QualifiedName;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:5077:1`
- `src/compiler/utilities.ts:6812:1`
- `src/compiler/utilities.ts:7579:1`

### guard-0040: src/compiler/factory/nodeTests.ts:397:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ComputedPropertyName
~~~~

Exact reason:

~~~~text
a view of Node as ComputedPropertyName in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Declaration) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isComputedPropertyName(node: Node): node is ComputedPropertyName {
    return node.kind === SyntaxKind.ComputedPropertyName;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/destructuring.ts:218:1`
- `src/compiler/utilities.ts:5077:1`

### guard-0041: src/compiler/factory/nodeTests.ts:412:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as Decorator
~~~~

Exact reason:

~~~~text
a view of Node as Decorator in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as NamedDeclaration) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isDecorator(node: Node): node is Decorator {
    return node.kind === SyntaxKind.Decorator;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:3155:1`
- `src/compiler/utilitiesPublic.ts:1778:1`

### guard-0042: src/compiler/factory/nodeTests.ts:450:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as CallSignatureDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as CallSignatureDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isCallSignatureDeclaration(node: Node): node is CallSignatureDeclaration {
    return node.kind === SyntaxKind.CallSignature;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1239:1`

### guard-0043: src/compiler/factory/nodeTests.ts:468:1

Observations: {'head': 3}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as TypeReferenceNode
~~~~

Exact reason:

~~~~text
a view of Node as TypeReferenceNode in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isTypeReferenceNode(node: Node): node is TypeReferenceNode {
    return node.kind === SyntaxKind.TypeReference;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:10625:1`
- `src/compiler/utilities.ts:3822:1`
- `src/compiler/utilitiesPublic.ts:1426:1`

### guard-0044: src/compiler/factory/nodeTests.ts:480:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as TypeQueryNode
~~~~

Exact reason:

~~~~text
a view of Node as TypeQueryNode in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isTypeQueryNode(node: Node): node is TypeQueryNode {
    return node.kind === SyntaxKind.TypeQuery;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/declarations/diagnostics.ts:615:1`

### guard-0045: src/compiler/factory/nodeTests.ts:484:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as TypeLiteralNode
~~~~

Exact reason:

~~~~text
a view of Node as TypeLiteralNode in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isTypeLiteralNode(node: Node): node is TypeLiteralNode {
    return node.kind === SyntaxKind.TypeLiteral;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:8323:1`
- `src/compiler/utilitiesPublic.ts:1239:1`

### guard-0046: src/compiler/factory/nodeTests.ts:524:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ParenthesizedTypeNode
~~~~

Exact reason:

~~~~text
a view of Node as ParenthesizedTypeNode in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isParenthesizedTypeNode(node: Node): node is ParenthesizedTypeNode {
    return node.kind === SyntaxKind.ParenthesizedType;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:5033:1`

### guard-0047: src/compiler/factory/nodeTests.ts:544:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as LiteralTypeNode
~~~~

Exact reason:

~~~~text
a view of Node as LiteralTypeNode in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isLiteralTypeNode(node: Node): node is LiteralTypeNode {
    return node.kind === SyntaxKind.LiteralType;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2722:1`

### guard-0048: src/compiler/factory/nodeTests.ts:548:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ImportTypeNode
~~~~

Exact reason:

~~~~text
a view of Node as ImportTypeNode in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isImportTypeNode(node: Node): node is ImportTypeNode {
    return node.kind === SyntaxKind.ImportType;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/program.ts:937:1`
- `src/compiler/utilities.ts:2722:1`

### guard-0049: src/compiler/factory/nodeTests.ts:576:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ArrayLiteralExpression
~~~~

Exact reason:

~~~~text
a view of Node as ArrayLiteralExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
// Expression

export function isArrayLiteralExpression(node: Node): node is ArrayLiteralExpression {
    return node.kind === SyntaxKind.ArrayLiteralExpression;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:10758:1`

### guard-0050: src/compiler/factory/nodeTests.ts:580:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ObjectLiteralExpression
~~~~

Exact reason:

~~~~text
a view of Node as ObjectLiteralExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isObjectLiteralExpression(node: Node): node is ObjectLiteralExpression {
    return node.kind === SyntaxKind.ObjectLiteralExpression;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:3018:1`
- `src/compiler/utilities.ts:3104:1`

### guard-0051: src/compiler/factory/nodeTests.ts:584:1

Observations: {'head': 9}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as PropertyAccessExpression
~~~~

Exact reason:

~~~~text
a view of Node as PropertyAccessExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isPropertyAccessExpression(node: Node): node is PropertyAccessExpression {
    return node.kind === SyntaxKind.PropertyAccessExpression;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:11182:1`
- `src/compiler/utilities.ts:4102:1`
- `src/compiler/utilities.ts:4117:1`
- `src/compiler/utilities.ts:4130:1`
- `src/compiler/utilities.ts:4146:1`
- `src/compiler/utilities.ts:7533:1`
- `src/compiler/utilities.ts:7573:1`
- `src/compiler/utilitiesPublic.ts:1368:1`
- `src/compiler/utilitiesPublic.ts:1605:1`

### guard-0052: src/compiler/factory/nodeTests.ts:588:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ElementAccessExpression
~~~~

Exact reason:

~~~~text
a view of Node as ElementAccessExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isElementAccessExpression(node: Node): node is ElementAccessExpression {
    return node.kind === SyntaxKind.ElementAccessExpression;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:4137:1`
- `src/compiler/utilitiesPublic.ts:1372:1`

### guard-0053: src/compiler/factory/nodeTests.ts:592:1

Observations: {'head': 3}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as CallExpression
~~~~

Exact reason:

~~~~text
a view of Node as CallExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isCallExpression(node: Node): node is CallExpression {
    return node.kind === SyntaxKind.CallExpression;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/factory/emitHelpers.ts:1486:1`
- `src/compiler/utilities.ts:3992:1`
- `src/compiler/utilitiesPublic.ts:1376:1`

### guard-0054: src/compiler/factory/nodeTests.ts:608:1

Observations: {'head': 4}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ParenthesizedExpression
~~~~

Exact reason:

~~~~text
a view of Node as ParenthesizedExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isParenthesizedExpression(node: Node): node is ParenthesizedExpression {
    return node.kind === SyntaxKind.ParenthesizedExpression;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/factory/utilities.ts:1695:1`
- `src/compiler/factory/utilities.ts:617:1`
- `src/compiler/utilities.ts:10769:1`
- `src/compiler/utilities.ts:11112:1`

### guard-0055: src/compiler/factory/nodeTests.ts:612:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as FunctionExpression
~~~~

Exact reason:

~~~~text
a view of Node as FunctionExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isFunctionExpression(node: Node): node is FunctionExpression {
    return node.kind === SyntaxKind.FunctionExpression;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:4995:1`
- `src/compiler/utilitiesPublic.ts:997:1`

### guard-0056: src/compiler/factory/nodeTests.ts:628:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as VoidExpression
~~~~

Exact reason:

~~~~text
a view of Node as VoidExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isVoidExpression(node: Node): node is VoidExpression {
    return node.kind === SyntaxKind.VoidExpression;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:4200:1`

### guard-0057: src/compiler/factory/nodeTests.ts:636:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as PrefixUnaryExpression
~~~~

Exact reason:

~~~~text
a view of Node as PrefixUnaryExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isPrefixUnaryExpression(node: Node): node is PrefixUnaryExpression {
    return node.kind === SyntaxKind.PrefixUnaryExpression;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:12031:1`
- `src/compiler/utilities.ts:5365:1`

### guard-0058: src/compiler/factory/nodeTests.ts:644:1

Observations: {'head': 18}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as BinaryExpression
~~~~

Exact reason:

~~~~text
a view of Node as BinaryExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isBinaryExpression(node: Node): node is BinaryExpression {
    return node.kind === SyntaxKind.BinaryExpression;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:3018:1`
- `src/compiler/utilities.ts:3025:1`
- `src/compiler/utilities.ts:3423:1`
- `src/compiler/utilities.ts:3629:1`
- `src/compiler/utilities.ts:3920:1`
- `src/compiler/utilities.ts:3929:1`
- `src/compiler/utilities.ts:3966:1`
- `src/compiler/utilities.ts:4017:1`
- `src/compiler/utilities.ts:4039:1`
- `src/compiler/utilities.ts:4278:1`
- `src/compiler/utilities.ts:4292:1`
- `src/compiler/utilities.ts:7418:1`
- `src/compiler/utilities.ts:7428:1`
- `src/compiler/utilities.ts:7469:1`
- `src/compiler/utilities.ts:7471:1`
- `src/compiler/utilities.ts:7473:1`
- `src/compiler/utilities.ts:7585:1`

### guard-0059: src/compiler/factory/nodeTests.ts:668:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as OmittedExpression
~~~~

Exact reason:

~~~~text
a view of Node as OmittedExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isOmittedExpression(node: Node): node is OmittedExpression {
    return node.kind === SyntaxKind.OmittedExpression;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:10749:1`

### guard-0060: src/compiler/factory/nodeTests.ts:672:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ExpressionWithTypeArguments
~~~~

Exact reason:

~~~~text
a view of Node as ExpressionWithTypeArguments in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isExpressionWithTypeArguments(node: Node): node is ExpressionWithTypeArguments {
    return node.kind === SyntaxKind.ExpressionWithTypeArguments;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:7453:1`

### guard-0061: src/compiler/factory/nodeTests.ts:684:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as NonNullExpression
~~~~

Exact reason:

~~~~text
a view of Node as NonNullExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isNonNullExpression(node: Node): node is NonNullExpression {
    return node.kind === SyntaxKind.NonNullExpression;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1437:1`

### guard-0062: src/compiler/factory/nodeTests.ts:688:1

Observations: {'head': 3}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as MetaProperty
~~~~

Exact reason:

~~~~text
a view of Node as MetaProperty in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isMetaProperty(node: Node): node is MetaProperty {
    return node.kind === SyntaxKind.MetaProperty;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/parser.ts:499:1`
- `src/compiler/utilities.ts:2704:1`
- `src/compiler/utilities.ts:2715:1`

### guard-0063: src/compiler/factory/nodeTests.ts:716:1

Observations: {'head': 4}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as Block
~~~~

Exact reason:

~~~~text
a view of Node as Block in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
// Elements

export function isBlock(node: Node): node is Block {
    return node.kind === SyntaxKind.Block;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/binder.ts:422:1`
- `src/compiler/utilities.ts:12431:1`
- `src/compiler/utilitiesPublic.ts:2160:1`
- `src/compiler/utilitiesPublic.ts:2166:1`

### guard-0064: src/compiler/factory/nodeTests.ts:720:1

Observations: {'head': 6}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as VariableStatement
~~~~

Exact reason:

~~~~text
a view of Node as VariableStatement in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isVariableStatement(node: Node): node is VariableStatement {
    return node.kind === SyntaxKind.VariableStatement;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/esnext.ts:788:1`
- `src/compiler/utilities.ts:11049:1`
- `src/compiler/utilities.ts:12436:1`
- `src/compiler/utilities.ts:2749:1`
- `src/compiler/utilities.ts:3903:1`
- `src/compiler/utilities.ts:4520:1`

### guard-0065: src/compiler/factory/nodeTests.ts:728:1

Observations: {'head': 4}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ExpressionStatement
~~~~

Exact reason:

~~~~text
a view of Node as ExpressionStatement in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isExpressionStatement(node: Node): node is ExpressionStatement {
    return node.kind === SyntaxKind.ExpressionStatement;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/utilities.ts:536:1`
- `src/compiler/utilities.ts:10769:1`
- `src/compiler/utilities.ts:4489:1`
- `src/compiler/utilities.ts:4497:1`

### guard-0066: src/compiler/factory/nodeTests.ts:784:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as TryStatement
~~~~

Exact reason:

~~~~text
a view of Node as TryStatement in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isTryStatement(node: Node): node is TryStatement {
    return node.kind === SyntaxKind.TryStatement;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/utilities.ts:547:1`

### guard-0067: src/compiler/factory/nodeTests.ts:800:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as FunctionDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as FunctionDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isFunctionDeclaration(node: Node): node is FunctionDeclaration {
    return node.kind === SyntaxKind.FunctionDeclaration;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2738:1`
- `src/compiler/utilities.ts:6890:1`

### guard-0068: src/compiler/factory/nodeTests.ts:808:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as InterfaceDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as InterfaceDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isInterfaceDeclaration(node: Node): node is InterfaceDeclaration {
    return node.kind === SyntaxKind.InterfaceDeclaration;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:5227:1`
- `src/compiler/utilities.ts:8323:1`

### guard-0069: src/compiler/factory/nodeTests.ts:816:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as EnumDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as EnumDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isEnumDeclaration(node: Node): node is EnumDeclaration {
    return node.kind === SyntaxKind.EnumDeclaration;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:11049:1`

### guard-0070: src/compiler/factory/nodeTests.ts:832:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as NamespaceExportDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as NamespaceExportDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isNamespaceExportDeclaration(node: Node): node is NamespaceExportDeclaration {
    return node.kind === SyntaxKind.NamespaceExportDeclaration;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:8283:1`

### guard-0071: src/compiler/factory/nodeTests.ts:886:1

Observations: {'head': 3}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ExportAssignment
~~~~

Exact reason:

~~~~text
a view of Node as ExportAssignment in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as SourceFile) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isExportAssignment(node: Node): node is ExportAssignment {
    return node.kind === SyntaxKind.ExportAssignment;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:2133:1`
- `src/compiler/utilitiesPublic.ts:2143:1`
- `src/compiler/utilitiesPublic.ts:2148:1`

### guard-0072: src/compiler/factory/nodeTests.ts:943:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JsxFragment
~~~~

Exact reason:

~~~~text
a view of Node as JsxFragment in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isJsxFragment(node: Node): node is JsxFragment {
    return node.kind === SyntaxKind.JsxFragment;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:8959:1`

### guard-0073: src/compiler/factory/nodeTests.ts:995:1

Observations: {'head': 3}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as PropertyAssignment
~~~~

Exact reason:

~~~~text
a view of Node as PropertyAssignment in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ObjectLiteralExpression) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
// Property assignments

export function isPropertyAssignment(node: Node): node is PropertyAssignment {
    return node.kind === SyntaxKind.PropertyAssignment;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:3423:1`
- `src/compiler/utilities.ts:3951:1`
- `src/compiler/utilitiesPublic.ts:1004:1`

### guard-0074: src/compiler/factory/nodeTests.ts:999:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ShorthandPropertyAssignment
~~~~

Exact reason:

~~~~text
a view of Node as ShorthandPropertyAssignment in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ObjectLiteralExpression) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isShorthandPropertyAssignment(node: Node): node is ShorthandPropertyAssignment {
    return node.kind === SyntaxKind.ShorthandPropertyAssignment;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:3423:1`

### guard-0075: src/compiler/factory/utilities.ts:1695:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Expression as ParenthesizedExpression
~~~~

Exact reason:

~~~~text
a view of Expression as ParenthesizedExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
function isSyntheticParenthesizedExpression(node: Expression): node is ParenthesizedExpression {
    return isParenthesizedExpression(node)
        && nodeIsSynthesized(node)
        && !node.emitNode;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/factory/utilities.ts:1701:1`

### guard-0076: src/compiler/factory/utilities.ts:608:58

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Expression as BinaryExpression
~~~~

Exact reason:

~~~~text
a view of Expression as BinaryExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as BinaryExpression
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/factory/utilities.ts:607:1`

### guard-0077: src/compiler/factory/utilities.ts:617:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDocTypeAssertion
~~~~

Exact reason:

~~~~text
a view of Node as JSDocTypeAssertion in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isJSDocTypeAssertion(node: Node): node is JSDocTypeAssertion {
    return isParenthesizedExpression(node)
        && isInJSFile(node)
        && !!getJSDocTypeTag(node);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/factory/utilities.ts:631:1`

### guard-0078: src/compiler/factory/utilities.ts:631:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as AsExpression
~~~~

Exact reason:

~~~~text
a view of Node as AsExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isOuterExpression(node: Node, kinds: OuterExpressionKinds = OuterExpressionKinds.All): node is OuterExpression {
    switch (node.kind) {
        case SyntaxKind.ParenthesizedExpression:
            if (kinds & OuterExpressionKinds.ExcludeJSDocTypeAssertion && isJSDocTypeAssertion(node)) {
                return false;
            }
            return (kinds & OuterExpressionKinds.Parentheses) !== 0;
        case SyntaxKind.TypeAssertionExpression:
        case SyntaxKind.AsExpression:
            return (kinds & OuterExpressionKinds.TypeAssertions) !== 0;
        case SyntaxKind.SatisfiesExpression:
            return (kinds & (OuterExpressionKinds.TypeAssertions | OuterExpressionKinds.Satisfies)) !== 0;
        case SyntaxKind.ExpressionWithTypeArguments:
            return (kinds & OuterExpressionKinds.ExpressionsWithTypeArguments) !== 0;
        case SyntaxKind.NonNullExpression:
            return (kinds & OuterExpressionKinds.NonNullAssertions) !== 0;
        case SyntaxKind.PartiallyEmittedExpression:
            return (kinds & OuterExpressionKinds.PartiallyEmittedExpressions) !== 0;
    }
    return false;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/factory/utilities.ts:668:1`

### guard-0079: src/compiler/factory/utilities.ts:631:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ExpressionWithTypeArguments
~~~~

Exact reason:

~~~~text
a view of Node as ExpressionWithTypeArguments in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isOuterExpression(node: Node, kinds: OuterExpressionKinds = OuterExpressionKinds.All): node is OuterExpression {
    switch (node.kind) {
        case SyntaxKind.ParenthesizedExpression:
            if (kinds & OuterExpressionKinds.ExcludeJSDocTypeAssertion && isJSDocTypeAssertion(node)) {
                return false;
            }
            return (kinds & OuterExpressionKinds.Parentheses) !== 0;
        case SyntaxKind.TypeAssertionExpression:
        case SyntaxKind.AsExpression:
            return (kinds & OuterExpressionKinds.TypeAssertions) !== 0;
        case SyntaxKind.SatisfiesExpression:
            return (kinds & (OuterExpressionKinds.TypeAssertions | OuterExpressionKinds.Satisfies)) !== 0;
        case SyntaxKind.ExpressionWithTypeArguments:
            return (kinds & OuterExpressionKinds.ExpressionsWithTypeArguments) !== 0;
        case SyntaxKind.NonNullExpression:
            return (kinds & OuterExpressionKinds.NonNullAssertions) !== 0;
        case SyntaxKind.PartiallyEmittedExpression:
            return (kinds & OuterExpressionKinds.PartiallyEmittedExpressions) !== 0;
    }
    return false;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/factory/utilities.ts:668:1`

### guard-0080: src/compiler/factory/utilities.ts:631:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as NonNullExpression
~~~~

Exact reason:

~~~~text
a view of Node as NonNullExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isOuterExpression(node: Node, kinds: OuterExpressionKinds = OuterExpressionKinds.All): node is OuterExpression {
    switch (node.kind) {
        case SyntaxKind.ParenthesizedExpression:
            if (kinds & OuterExpressionKinds.ExcludeJSDocTypeAssertion && isJSDocTypeAssertion(node)) {
                return false;
            }
            return (kinds & OuterExpressionKinds.Parentheses) !== 0;
        case SyntaxKind.TypeAssertionExpression:
        case SyntaxKind.AsExpression:
            return (kinds & OuterExpressionKinds.TypeAssertions) !== 0;
        case SyntaxKind.SatisfiesExpression:
            return (kinds & (OuterExpressionKinds.TypeAssertions | OuterExpressionKinds.Satisfies)) !== 0;
        case SyntaxKind.ExpressionWithTypeArguments:
            return (kinds & OuterExpressionKinds.ExpressionsWithTypeArguments) !== 0;
        case SyntaxKind.NonNullExpression:
            return (kinds & OuterExpressionKinds.NonNullAssertions) !== 0;
        case SyntaxKind.PartiallyEmittedExpression:
            return (kinds & OuterExpressionKinds.PartiallyEmittedExpressions) !== 0;
    }
    return false;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/factory/utilities.ts:668:1`

### guard-0081: src/compiler/factory/utilities.ts:631:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ParenthesizedExpression
~~~~

Exact reason:

~~~~text
a view of Node as ParenthesizedExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isOuterExpression(node: Node, kinds: OuterExpressionKinds = OuterExpressionKinds.All): node is OuterExpression {
    switch (node.kind) {
        case SyntaxKind.ParenthesizedExpression:
            if (kinds & OuterExpressionKinds.ExcludeJSDocTypeAssertion && isJSDocTypeAssertion(node)) {
                return false;
            }
            return (kinds & OuterExpressionKinds.Parentheses) !== 0;
        case SyntaxKind.TypeAssertionExpression:
        case SyntaxKind.AsExpression:
            return (kinds & OuterExpressionKinds.TypeAssertions) !== 0;
        case SyntaxKind.SatisfiesExpression:
            return (kinds & (OuterExpressionKinds.TypeAssertions | OuterExpressionKinds.Satisfies)) !== 0;
        case SyntaxKind.ExpressionWithTypeArguments:
            return (kinds & OuterExpressionKinds.ExpressionsWithTypeArguments) !== 0;
        case SyntaxKind.NonNullExpression:
            return (kinds & OuterExpressionKinds.NonNullAssertions) !== 0;
        case SyntaxKind.PartiallyEmittedExpression:
            return (kinds & OuterExpressionKinds.PartiallyEmittedExpressions) !== 0;
    }
    return false;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/factory/utilities.ts:668:1`

### guard-0082: src/compiler/factory/utilities.ts:631:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as PartiallyEmittedExpression
~~~~

Exact reason:

~~~~text
a view of Node as PartiallyEmittedExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isOuterExpression(node: Node, kinds: OuterExpressionKinds = OuterExpressionKinds.All): node is OuterExpression {
    switch (node.kind) {
        case SyntaxKind.ParenthesizedExpression:
            if (kinds & OuterExpressionKinds.ExcludeJSDocTypeAssertion && isJSDocTypeAssertion(node)) {
                return false;
            }
            return (kinds & OuterExpressionKinds.Parentheses) !== 0;
        case SyntaxKind.TypeAssertionExpression:
        case SyntaxKind.AsExpression:
            return (kinds & OuterExpressionKinds.TypeAssertions) !== 0;
        case SyntaxKind.SatisfiesExpression:
            return (kinds & (OuterExpressionKinds.TypeAssertions | OuterExpressionKinds.Satisfies)) !== 0;
        case SyntaxKind.ExpressionWithTypeArguments:
            return (kinds & OuterExpressionKinds.ExpressionsWithTypeArguments) !== 0;
        case SyntaxKind.NonNullExpression:
            return (kinds & OuterExpressionKinds.NonNullAssertions) !== 0;
        case SyntaxKind.PartiallyEmittedExpression:
            return (kinds & OuterExpressionKinds.PartiallyEmittedExpressions) !== 0;
    }
    return false;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/factory/utilities.ts:668:1`

### guard-0083: src/compiler/factory/utilities.ts:631:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as SatisfiesExpression
~~~~

Exact reason:

~~~~text
a view of Node as SatisfiesExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isOuterExpression(node: Node, kinds: OuterExpressionKinds = OuterExpressionKinds.All): node is OuterExpression {
    switch (node.kind) {
        case SyntaxKind.ParenthesizedExpression:
            if (kinds & OuterExpressionKinds.ExcludeJSDocTypeAssertion && isJSDocTypeAssertion(node)) {
                return false;
            }
            return (kinds & OuterExpressionKinds.Parentheses) !== 0;
        case SyntaxKind.TypeAssertionExpression:
        case SyntaxKind.AsExpression:
            return (kinds & OuterExpressionKinds.TypeAssertions) !== 0;
        case SyntaxKind.SatisfiesExpression:
            return (kinds & (OuterExpressionKinds.TypeAssertions | OuterExpressionKinds.Satisfies)) !== 0;
        case SyntaxKind.ExpressionWithTypeArguments:
            return (kinds & OuterExpressionKinds.ExpressionsWithTypeArguments) !== 0;
        case SyntaxKind.NonNullExpression:
            return (kinds & OuterExpressionKinds.NonNullAssertions) !== 0;
        case SyntaxKind.PartiallyEmittedExpression:
            return (kinds & OuterExpressionKinds.PartiallyEmittedExpressions) !== 0;
    }
    return false;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/factory/utilities.ts:668:1`

### guard-0084: src/compiler/factory/utilities.ts:631:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as TypeAssertion
~~~~

Exact reason:

~~~~text
a view of Node as TypeAssertion in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isOuterExpression(node: Node, kinds: OuterExpressionKinds = OuterExpressionKinds.All): node is OuterExpression {
    switch (node.kind) {
        case SyntaxKind.ParenthesizedExpression:
            if (kinds & OuterExpressionKinds.ExcludeJSDocTypeAssertion && isJSDocTypeAssertion(node)) {
                return false;
            }
            return (kinds & OuterExpressionKinds.Parentheses) !== 0;
        case SyntaxKind.TypeAssertionExpression:
        case SyntaxKind.AsExpression:
            return (kinds & OuterExpressionKinds.TypeAssertions) !== 0;
        case SyntaxKind.SatisfiesExpression:
            return (kinds & (OuterExpressionKinds.TypeAssertions | OuterExpressionKinds.Satisfies)) !== 0;
        case SyntaxKind.ExpressionWithTypeArguments:
            return (kinds & OuterExpressionKinds.ExpressionsWithTypeArguments) !== 0;
        case SyntaxKind.NonNullExpression:
            return (kinds & OuterExpressionKinds.NonNullAssertions) !== 0;
        case SyntaxKind.PartiallyEmittedExpression:
            return (kinds & OuterExpressionKinds.PartiallyEmittedExpressions) !== 0;
    }
    return false;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/factory/utilities.ts:668:1`

### guard-0085: src/compiler/factory/utilitiesPublic.ts:14:1

Observations: {'head': 4}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ArrowFunction
~~~~

Exact reason:

~~~~text
a view of Node as ArrowFunction in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function canHaveModifiers(node: Node): node is HasModifiers {
    const kind = node.kind;
    return kind === SyntaxKind.TypeParameter
        || kind === SyntaxKind.Parameter
        || kind === SyntaxKind.PropertySignature
        || kind === SyntaxKind.PropertyDeclaration
        || kind === SyntaxKind.MethodSignature
        || kind === SyntaxKind.MethodDeclaration
        || kind === SyntaxKind.Constructor
        || kind === SyntaxKind.GetAccessor
        || kind === SyntaxKind.SetAccessor
        || kind === SyntaxKind.IndexSignature
        || kind === SyntaxKind.ConstructorType
        || kind === SyntaxKind.FunctionExpression
        || kind === SyntaxKind.ArrowFunction
        || kind === SyntaxKind.ClassExpression
        || kind === SyntaxKind.VariableStatement
        || kind === SyntaxKind.FunctionDeclaration
        || kind === SyntaxKind.ClassDeclaration
        || kind === SyntaxKind.InterfaceDeclaration
        || kind === SyntaxKind.TypeAliasDeclaration
        || kind === SyntaxKind.EnumDeclaration
        || kind === SyntaxKind.ModuleDeclaration
        || kind === SyntaxKind.ImportEqualsDeclaration
        || kind === SyntaxKind.ImportDeclaration
        || kind === SyntaxKind.ExportAssignment
        || kind === SyntaxKind.ExportDeclaration;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/parser.ts:476:1`
- `src/compiler/utilities.ts:1277:1`
- `src/compiler/utilities.ts:7342:1`
- `src/compiler/utilities.ts:7877:1`

### guard-0086: src/compiler/utilities.ts:11182:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Declaration | undefined as BinaryExpression
~~~~

Exact reason:

~~~~text
a view of Declaration | undefined as BinaryExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isExpandoPropertyDeclaration(declaration: Declaration | undefined): declaration is PropertyAccessExpression | ElementAccessExpression | BinaryExpression {
    return !!declaration && (isPropertyAccessExpression(declaration) || isElementAccessExpression(declaration) || isBinaryExpression(declaration));
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/declarations.ts:259:1`
- `src/compiler/utilitiesPublic.ts:1757:1`

### guard-0087: src/compiler/utilities.ts:11182:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Declaration | undefined as ElementAccessExpression
~~~~

Exact reason:

~~~~text
a view of Declaration | undefined as ElementAccessExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isExpandoPropertyDeclaration(declaration: Declaration | undefined): declaration is PropertyAccessExpression | ElementAccessExpression | BinaryExpression {
    return !!declaration && (isPropertyAccessExpression(declaration) || isElementAccessExpression(declaration) || isBinaryExpression(declaration));
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/declarations.ts:259:1`
- `src/compiler/utilitiesPublic.ts:1757:1`

### guard-0088: src/compiler/utilities.ts:12059:14

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Expression as ParenthesizedExpression
~~~~

Exact reason:

~~~~text
a view of Expression as ParenthesizedExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
o as ParenthesizedExpression
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:12057:1`

### guard-0089: src/compiler/utilities.ts:1995:18

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
LiteralLikeNode as StringLiteral
~~~~

Exact reason:

~~~~text
a view of LiteralLikeNode as StringLiteral in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as StringLiteral
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:1981:1`

### guard-0090: src/compiler/utilities.ts:2079:1

Observations: {'head': 3}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as AmbientModuleDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as AmbientModuleDeclaration in a graph program, which the cycle finder doesn't relate yet through its field original (Node | undefined as Node | undefined) through its field parent (Node as Node) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isAmbientModule(node: Node): node is AmbientModuleDeclaration {
    return isModuleDeclaration(node) && (node.name.kind === SyntaxKind.StringLiteral || isGlobalScopeAugmentation(node));
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2130:1`
- `src/compiler/utilities.ts:2135:1`
- `src/compiler/utilitiesPublic.ts:2143:1`

### guard-0091: src/compiler/utilities.ts:2130:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as AmbientModuleDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as AmbientModuleDeclaration in a graph program, which the cycle finder doesn't relate yet through its field original (Node | undefined as Node | undefined) through its field parent (Node as Node) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isExternalModuleAugmentation(node: Node): node is AmbientModuleDeclaration {
    return isAmbientModule(node) && isModuleAugmentationExternal(node);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2149:1`

### guard-0092: src/compiler/utilities.ts:2245:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ArrowFunction
~~~~

Exact reason:

~~~~text
a view of Node as ArrowFunction in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isDeclarationWithTypeParameterChildren(node: Node): node is DeclarationWithTypeParameterChildren {
    switch (node.kind) {
        case SyntaxKind.CallSignature:
        case SyntaxKind.ConstructSignature:
        case SyntaxKind.MethodSignature:
        case SyntaxKind.IndexSignature:
        case SyntaxKind.FunctionType:
        case SyntaxKind.ConstructorType:
        case SyntaxKind.JSDocFunctionType:
        case SyntaxKind.ClassDeclaration:
        case SyntaxKind.ClassExpression:
        case SyntaxKind.InterfaceDeclaration:
        case SyntaxKind.TypeAliasDeclaration:
        case SyntaxKind.JSDocTemplateTag:
        case SyntaxKind.FunctionDeclaration:
        case SyntaxKind.MethodDeclaration:
        case SyntaxKind.Constructor:
        case SyntaxKind.GetAccessor:
        case SyntaxKind.SetAccessor:
        case SyntaxKind.FunctionExpression:
        case SyntaxKind.ArrowFunction:
            return true;
        default:
            assertType<Exclude<SyntaxKind, DeclarationWithTypeParameterChildren["kind"]>>(node.kind);
            return false;
    }
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2233:1`

### guard-0093: src/compiler/utilities.ts:2568:26

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as NamedDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as NamedDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as NamedDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2538:1`

### guard-0094: src/compiler/utilities.ts:2571:61

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ArrowFunction
~~~~

Exact reason:

~~~~text
a view of Node as ArrowFunction in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ArrowFunction
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2538:1`

### guard-0095: src/compiler/utilities.ts:2632:79

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as SourceFile
~~~~

Exact reason:

~~~~text
a view of Node as SourceFile in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as SourceFile
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2631:1`

### guard-0096: src/compiler/utilities.ts:2641:1

Observations: {'head': 3}. New in head: True.

Source and target as named by the guard:

~~~~text
SourceFile as JsonSourceFile
~~~~

Exact reason:

~~~~text
a view of SourceFile as JsonSourceFile in a graph program, which the cycle finder doesn't relate yet through its field statements (NodeArray<Statement> as NodeArray<JsonObjectExpressionStatement>) through its index values (Statement as JsonObjectExpressionStatement) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isJsonSourceFile(file: SourceFile): file is JsonSourceFile {
    return file.scriptKind === ScriptKind.JSON;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/emitter.ts:516:1`
- `src/compiler/utilities.ts:3812:1`
- `src/compiler/utilities.ts:6659:1`

### guard-0097: src/compiler/utilities.ts:2700:53

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as CallExpression
~~~~

Exact reason:

~~~~text
a view of Node as CallExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
n as CallExpression
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2699:1`

### guard-0098: src/compiler/utilities.ts:2704:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ImportCall
~~~~

Exact reason:

~~~~text
a view of Node as ImportCall in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isImportCall(n: Node): n is ImportCall {
    if (n.kind !== SyntaxKind.CallExpression) return false;
    const e = (n as CallExpression).expression;
    return e.kind === SyntaxKind.ImportKeyword || (
        isMetaProperty(e)
        && e.keywordToken === SyntaxKind.ImportKeyword
        && e.name.escapedText === "defer"
    );
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/program.ts:959:1`
- `src/compiler/utilities.ts:3629:1`

### guard-0099: src/compiler/utilities.ts:2706:16

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as CallExpression
~~~~

Exact reason:

~~~~text
a view of Node as CallExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
n as CallExpression
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2704:1`

### guard-0100: src/compiler/utilities.ts:2727:1

Observations: {'head': 4}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as PrologueDirective
~~~~

Exact reason:

~~~~text
a view of Node as PrologueDirective in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isPrologueDirective(node: Node): node is PrologueDirective {
    return node.kind === SyntaxKind.ExpressionStatement
        && (node as ExpressionStatement).expression.kind === SyntaxKind.StringLiteral;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/factory/utilities.ts:584:1`
- `src/compiler/factory/utilities.ts:599:1`
- `src/compiler/transformers/esnext.ts:765:1`
- `src/compiler/utilities.ts:1143:1`

### guard-0101: src/compiler/utilities.ts:2729:13

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ExpressionStatement
~~~~

Exact reason:

~~~~text
a view of Node as ExpressionStatement in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ExpressionStatement
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2727:1`

### guard-0102: src/compiler/utilities.ts:2815:67

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as QualifiedName
~~~~

Exact reason:

~~~~text
a view of Node as QualifiedName in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node.parent as QualifiedName
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2786:1`

### guard-0103: src/compiler/utilities.ts:2970:17

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
TypeNode as ArrayTypeNode
~~~~

Exact reason:

~~~~text
a view of TypeNode as ArrayTypeNode in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ArrayTypeNode
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2968:1`

### guard-0104: src/compiler/utilities.ts:2973:35

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
TypeNode as TypeReferenceNode
~~~~

Exact reason:

~~~~text
a view of TypeNode as TypeReferenceNode in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as TypeReferenceNode
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2968:1`

### guard-0105: src/compiler/utilities.ts:3061:16

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Statement as LabeledStatement
~~~~

Exact reason:

~~~~text
a view of Statement as LabeledStatement in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node.statement as LabeledStatement
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:3053:1`

### guard-0106: src/compiler/utilities.ts:3389:59

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as CallExpression
~~~~

Exact reason:

~~~~text
a view of Node as CallExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
parent as CallExpression
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:3381:1`

### guard-0107: src/compiler/utilities.ts:3390:20

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as CallExpression
~~~~

Exact reason:

~~~~text
a view of Node as CallExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
parent as CallExpression
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:3381:1`

### guard-0108: src/compiler/utilities.ts:3435:21

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
TypeNode as TypeReferenceNode
~~~~

Exact reason:

~~~~text
a view of TypeNode as TypeReferenceNode in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as TypeReferenceNode
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:3432:1`

### guard-0109: src/compiler/utilities.ts:3438:44

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
TypeNode as ExpressionWithTypeArguments
~~~~

Exact reason:

~~~~text
a view of TypeNode as ExpressionWithTypeArguments in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ExpressionWithTypeArguments
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:3432:1`

### guard-0110: src/compiler/utilities.ts:3554:26

Observations: {'head': 6}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ClassDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as ClassDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ClassDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:3547:1`
- `src/compiler/utilities.ts:3549:1`
- `src/compiler/utilities.ts:3551:1`

### guard-0111: src/compiler/utilities.ts:3556:50

Observations: {'head': 6}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ClassExpression
~~~~

Exact reason:

~~~~text
a view of Node as ClassExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ClassExpression
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:3547:1`
- `src/compiler/utilities.ts:3549:1`
- `src/compiler/utilities.ts:3551:1`

### guard-0112: src/compiler/utilities.ts:3839:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as CallExpression
~~~~

Exact reason:

~~~~text
a view of Node as CallExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isRequireCall(callExpression: Node, requireStringLiteralLikeArgument: boolean): callExpression is CallExpression;
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/program.ts:959:1`
- `src/compiler/transformers/module/esnextAnd2015.ts:67:1`

### guard-0113: src/compiler/utilities.ts:3845:45

Observations: {'head': 6}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as CallExpression
~~~~

Exact reason:

~~~~text
a view of Node as CallExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
callExpression as CallExpression
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:3837:1`
- `src/compiler/utilities.ts:3839:1`
- `src/compiler/utilities.ts:3841:1`

### guard-0114: src/compiler/utilities.ts:3847:55

Observations: {'head': 3}. New in head: True.

Source and target as named by the guard:

~~~~text
LeftHandSideExpression as Identifier
~~~~

Exact reason:

~~~~text
a view of LeftHandSideExpression as Identifier in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
expression as Identifier
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:3837:1`
- `src/compiler/utilities.ts:3839:1`
- `src/compiler/utilities.ts:3841:1`

### guard-0115: src/compiler/utilities.ts:3864:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as VariableDeclarationInitializedTo<RequireOrImportCall>
~~~~

Exact reason:

~~~~text
a view of Node as VariableDeclarationInitializedTo<RequireOrImportCall> in a graph program, which the cycle finder doesn't relate yet through its field original (Node | undefined as Node | undefined) through its field parent (Node as Node) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/**
 * Returns true if the node is a VariableDeclaration initialized to a require call (see `isRequireCall`).
 * This function does not test if the node is in a JavaScript file or not.
 *
 * @internal
 */
export function isVariableDeclarationInitializedToRequire(node: Node): node is VariableDeclarationInitializedTo<RequireOrImportCall> {
    return isVariableDeclarationInitializedWithRequireHelper(node, /*allowAccessedRequire*/ false);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:3883:1`

### guard-0116: src/compiler/utilities.ts:3873:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as VariableDeclarationInitializedTo<AccessExpression | RequireOrImportCall>
~~~~

Exact reason:

~~~~text
a view of Node as VariableDeclarationInitializedTo<AccessExpression | RequireOrImportCall> in a graph program, which the cycle finder doesn't relate yet through its field original (Node | undefined as Node | undefined) through its field parent (Node as Node) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/**
 * Like {@link isVariableDeclarationInitializedToRequire} but allows things like `require("...").foo.bar` or `require("...")["baz"]`.
 *
 * @internal
 */
export function isVariableDeclarationInitializedToBareOrAccessedRequire(node: Node): node is VariableDeclarationInitializedTo<RequireOrImportCall | AccessExpression> {
    return isVariableDeclarationInitializedWithRequireHelper(node, /*allowAccessedRequire*/ true);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2285:1`
- `src/compiler/utilities.ts:3782:1`

### guard-0117: src/compiler/utilities.ts:3903:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as RequireVariableStatement
~~~~

Exact reason:

~~~~text
a view of Node as RequireVariableStatement in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isRequireVariableStatement(node: Node): node is RequireVariableStatement {
    return isVariableStatement(node)
        && node.declarationList.declarations.length > 0
        && every(node.declarationList.declarations, decl => isVariableDeclarationInitializedToRequire(decl));
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:10125:1`
- `src/compiler/utilities.ts:2290:1`

### guard-0118: src/compiler/utilities.ts:4146:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as PropertyAccessEntityNameExpression
~~~~

Exact reason:

~~~~text
a view of Node as PropertyAccessEntityNameExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/**
 * Any series of property and element accesses.
 *
 * @internal
 */
export function isBindableStaticAccessExpression(node: Node, excludeThisKeyword?: boolean): node is BindableStaticAccessExpression {
    return isPropertyAccessExpression(node) && (!excludeThisKeyword && node.expression.kind === SyntaxKind.ThisKeyword || isIdentifier(node.name) && isBindableStaticNameExpression(node.expression, /*excludeThisKeyword*/ true))
        || isBindableStaticElementAccessExpression(node, excludeThisKeyword);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:7561:1`

### guard-0119: src/compiler/utilities.ts:4515:21

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as PropertyAssignment
~~~~

Exact reason:

~~~~text
a view of Node as PropertyAssignment in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ObjectLiteralExpression) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as PropertyAssignment
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:4507:1`

### guard-0120: src/compiler/utilities.ts:4559:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ArrowFunction
~~~~

Exact reason:

~~~~text
a view of Node as ArrowFunction in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function canHaveJSDoc(node: Node): node is HasJSDoc {
    switch (node.kind) {
        case SyntaxKind.ArrowFunction:
        case SyntaxKind.BinaryExpression:
        case SyntaxKind.Block:
        case SyntaxKind.BreakStatement:
        case SyntaxKind.CallSignature:
        case SyntaxKind.CaseClause:
        case SyntaxKind.ClassDeclaration:
        case SyntaxKind.ClassExpression:
        case SyntaxKind.ClassStaticBlockDeclaration:
        case SyntaxKind.Constructor:
        case SyntaxKind.ConstructorType:
        case SyntaxKind.ConstructSignature:
        case SyntaxKind.ContinueStatement:
        case SyntaxKind.DebuggerStatement:
        case SyntaxKind.DoStatement:
        case SyntaxKind.ElementAccessExpression:
        case SyntaxKind.EmptyStatement:
        case SyntaxKind.EndOfFileToken:
        case SyntaxKind.EnumDeclaration:
        case SyntaxKind.EnumMember:
        case SyntaxKind.ExportAssignment:
        case SyntaxKind.ExportDeclaration:
        case SyntaxKind.ExportSpecifier:
        case SyntaxKind.ExpressionStatement:
        case SyntaxKind.ForInStatement:
        case SyntaxKind.ForOfStatement:
        case SyntaxKind.ForStatement:
        case SyntaxKind.FunctionDeclaration:
        case SyntaxKind.FunctionExpression:
        case SyntaxKind.FunctionType:
        case SyntaxKind.GetAccessor:
        case SyntaxKind.Identifier:
        case SyntaxKind.IfStatement:
        case SyntaxKind.ImportDeclaration:
        case SyntaxKind.ImportEqualsDeclaration:
        case SyntaxKind.IndexSignature:
        case SyntaxKind.InterfaceDeclaration:
        case SyntaxKind.JSDocFunctionType:
        case SyntaxKind.JSDocSignature:
        case SyntaxKind.LabeledStatement:
        case SyntaxKind.MethodDeclaration:
        case SyntaxKind.MethodSignature:
        case SyntaxKind.ModuleDeclaration:
        case SyntaxKind.NamedTupleMember:
        case SyntaxKind.NamespaceExportDeclaration:
        case SyntaxKind.ObjectLiteralExpression:
        case SyntaxKind.Parameter:
        case SyntaxKind.ParenthesizedExpression:
        case SyntaxKind.PropertyAccessExpression:
        case SyntaxKind.PropertyAssignment:
        case SyntaxKind.PropertyDeclaration:
        case SyntaxKind.PropertySignature:
        case SyntaxKind.ReturnStatement:
        case SyntaxKind.SemicolonClassElement:
        case SyntaxKind.SetAccessor:
        case SyntaxKind.ShorthandPropertyAssignment:
        case SyntaxKind.SpreadAssignment:
        case SyntaxKind.SwitchStatement:
        case SyntaxKind.ThrowStatement:
        case SyntaxKind.TryStatement:
        case SyntaxKind.TypeAliasDeclaration:
        case SyntaxKind.TypeParameter:
        case SyntaxKind.VariableDeclaration:
        case SyntaxKind.VariableStatement:
        case SyntaxKind.WhileStatement:
        case SyntaxKind.WithStatement:
            return true;
        default:
            return false;
    }
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1257:1`
- `src/compiler/utilitiesPublic.ts:2552:1`

### guard-0121: src/compiler/utilities.ts:4857:42

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as BinaryExpression
~~~~

Exact reason:

~~~~text
a view of Node as BinaryExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
parent as BinaryExpression
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:4852:1`

### guard-0122: src/compiler/utilities.ts:5026:17

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ParenthesizedTypeNode
~~~~

Exact reason:

~~~~text
a view of Node as ParenthesizedTypeNode in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ParenthesizedTypeNode
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:5023:1`

### guard-0123: src/compiler/utilities.ts:5160:20

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as Declaration
~~~~

Exact reason:

~~~~text
a view of Node as Declaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node.parent as Declaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:5151:1`

### guard-0124: src/compiler/utilities.ts:5189:10

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as BinaryExpression
~~~~

Exact reason:

~~~~text
a view of Node as BinaryExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node.parent as BinaryExpression
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:5187:1`

### guard-0125: src/compiler/utilities.ts:5360:1

Observations: {'head': 4}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as NoSubstitutionTemplateLiteral
~~~~

Exact reason:

~~~~text
a view of Node as NoSubstitutionTemplateLiteral in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isStringOrNumericLiteralLike(node: Node): node is StringLiteralLike | NumericLiteral {
    return isStringLiteralLike(node) || isNumericLiteral(node);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2357:1`
- `src/compiler/utilities.ts:2362:1`
- `src/compiler/utilities.ts:5110:1`
- `src/compiler/utilities.ts:5395:1`

### guard-0126: src/compiler/utilities.ts:5360:1

Observations: {'head': 4}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as NumericLiteral
~~~~

Exact reason:

~~~~text
a view of Node as NumericLiteral in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isStringOrNumericLiteralLike(node: Node): node is StringLiteralLike | NumericLiteral {
    return isStringLiteralLike(node) || isNumericLiteral(node);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2357:1`
- `src/compiler/utilities.ts:2362:1`
- `src/compiler/utilities.ts:5110:1`
- `src/compiler/utilities.ts:5395:1`

### guard-0127: src/compiler/utilities.ts:5360:1

Observations: {'head': 4}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as StringLiteral
~~~~

Exact reason:

~~~~text
a view of Node as StringLiteral in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isStringOrNumericLiteralLike(node: Node): node is StringLiteralLike | NumericLiteral {
    return isStringLiteralLike(node) || isNumericLiteral(node);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2357:1`
- `src/compiler/utilities.ts:2362:1`
- `src/compiler/utilities.ts:5110:1`
- `src/compiler/utilities.ts:5395:1`

### guard-0128: src/compiler/utilities.ts:5425:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as BigIntLiteral
~~~~

Exact reason:

~~~~text
a view of Node as BigIntLiteral in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isPropertyNameLiteral(node: Node): node is PropertyNameLiteral {
    switch (node.kind) {
        case SyntaxKind.Identifier:
        case SyntaxKind.StringLiteral:
        case SyntaxKind.NoSubstitutionTemplateLiteral:
        case SyntaxKind.NumericLiteral:
            return true;
        default:
            return false;
    }
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:4062:1`

### guard-0129: src/compiler/utilities.ts:5425:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as Identifier
~~~~

Exact reason:

~~~~text
a view of Node as Identifier in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isPropertyNameLiteral(node: Node): node is PropertyNameLiteral {
    switch (node.kind) {
        case SyntaxKind.Identifier:
        case SyntaxKind.StringLiteral:
        case SyntaxKind.NoSubstitutionTemplateLiteral:
        case SyntaxKind.NumericLiteral:
            return true;
        default:
            return false;
    }
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:4062:1`

### guard-0130: src/compiler/utilities.ts:5520:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ExportAssignment
~~~~

Exact reason:

~~~~text
a view of Node as ExportAssignment in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as SourceFile) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/**
 * Indicates whether a node is a potential source of an assigned name for a class, function, or arrow function.
 *
 * @internal
 */
export function isNamedEvaluationSource(node: Node): node is NamedEvaluationSource {
    switch (node.kind) {
        case SyntaxKind.PropertyAssignment:
            return !isProtoSetter((node as PropertyAssignment).name);
        case SyntaxKind.ShorthandPropertyAssignment:
            return !!(node as ShorthandPropertyAssignment).objectAssignmentInitializer;
        case SyntaxKind.VariableDeclaration:
            return isIdentifier((node as VariableDeclaration).name) && !!(node as VariableDeclaration).initializer;
        case SyntaxKind.Parameter:
            return isIdentifier((node as ParameterDeclaration).name) && !!(node as VariableDeclaration).initializer && !(node as BindingElement).dotDotDotToken;
        case SyntaxKind.BindingElement:
            return isIdentifier((node as BindingElement).name) && !!(node as VariableDeclaration).initializer && !(node as BindingElement).dotDotDotToken;
        case SyntaxKind.PropertyDeclaration:
            return !!(node as PropertyDeclaration).initializer;
        case SyntaxKind.BinaryExpression:
            switch ((node as BinaryExpression).operatorToken.kind) {
                case SyntaxKind.EqualsToken:
                case SyntaxKind.AmpersandAmpersandEqualsToken:
                case SyntaxKind.BarBarEqualsToken:
                case SyntaxKind.QuestionQuestionEqualsToken:
                    return isIdentifier((node as BinaryExpression).left);
            }
            break;
        case SyntaxKind.ExportAssignment:
            return true;
    }
    return false;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:5562:1`

### guard-0131: src/compiler/utilities.ts:5523:36

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as PropertyAssignment
~~~~

Exact reason:

~~~~text
a view of Node as PropertyAssignment in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ObjectLiteralExpression) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as PropertyAssignment
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:5520:1`

### guard-0132: src/compiler/utilities.ts:5525:23

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ShorthandPropertyAssignment
~~~~

Exact reason:

~~~~text
a view of Node as ShorthandPropertyAssignment in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ObjectLiteralExpression) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ShorthandPropertyAssignment
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:5520:1`

### guard-0133: src/compiler/utilities.ts:5637:75

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Expression as NewExpression
~~~~

Exact reason:

~~~~text
a view of Expression as NewExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
expression as NewExpression
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:5635:1`

### guard-0134: src/compiler/utilities.ts:5684:75

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Expression as NewExpression
~~~~

Exact reason:

~~~~text
a view of Expression as NewExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
expression as NewExpression
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:5682:1`

### guard-0135: src/compiler/utilities.ts:5690:17

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Expression as BinaryExpression
~~~~

Exact reason:

~~~~text
a view of Expression as BinaryExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
expression as BinaryExpression
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:5688:1`

### guard-0136: src/compiler/utilities.ts:6797:85

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as Identifier
~~~~

Exact reason:

~~~~text
a view of Node as Identifier in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as Identifier
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:6796:1`

### guard-0137: src/compiler/utilities.ts:7469:1

Observations: {'head': 4}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as AssignmentExpression<EqualsToken>
~~~~

Exact reason:

~~~~text
a view of Node as AssignmentExpression<EqualsToken> in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isAssignmentExpression(node: Node, excludeCompoundAssignment: true): node is AssignmentExpression<EqualsToken>;
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:4084:1`
- `src/compiler/utilities.ts:4933:1`
- `src/compiler/utilities.ts:7482:1`
- `src/compiler/utilitiesPublic.ts:1932:1`

### guard-0138: src/compiler/utilities.ts:7471:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as AssignmentExpression<AssignmentOperatorToken>
~~~~

Exact reason:

~~~~text
a view of Node as AssignmentExpression<AssignmentOperatorToken> in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isAssignmentExpression(node: Node, excludeCompoundAssignment?: false): node is AssignmentExpression<AssignmentOperatorToken>;
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:4708:1`

### guard-0139: src/compiler/utilities.ts:7493:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ExpressionWithTypeArguments
~~~~

Exact reason:

~~~~text
a view of Node as ExpressionWithTypeArguments in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isExpressionWithTypeArgumentsInClassExtendsClause(node: Node): node is ExpressionWithTypeArguments {
    return tryGetClassExtendingExpressionWithTypeArguments(node) !== undefined;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2882:1`

### guard-0140: src/compiler/utilities.ts:7498:1

Observations: {'head': 4}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as Identifier
~~~~

Exact reason:

~~~~text
a view of Node as Identifier in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isEntityNameExpression(node: Node): node is EntityNameExpression {
    return node.kind === SyntaxKind.Identifier || isPropertyAccessEntityNameExpression(node);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:3432:1`
- `src/compiler/utilities.ts:3929:1`
- `src/compiler/utilities.ts:4164:1`
- `src/compiler/utilities.ts:5171:1`

### guard-0141: src/compiler/utilities.ts:7528:79

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Expression as PropertyAccessExpression
~~~~

Exact reason:

~~~~text
a view of Expression as PropertyAccessExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as PropertyAccessExpression
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:7523:1`

### guard-0142: src/compiler/utilities.ts:7533:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as PropertyAccessEntityNameExpression
~~~~

Exact reason:

~~~~text
a view of Node as PropertyAccessEntityNameExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isPropertyAccessEntityNameExpression(node: Node): node is PropertyAccessEntityNameExpression {
    return isPropertyAccessExpression(node) && isIdentifier(node.name) && isEntityNameExpression(node.expression);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:7498:1`

### guard-0143: src/compiler/utilities.ts:7561:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as PropertyAccessEntityNameExpression
~~~~

Exact reason:

~~~~text
a view of Node as PropertyAccessEntityNameExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isPrototypeAccess(node: Node): node is BindableStaticAccessExpression {
    return isBindableStaticAccessExpression(node) && getElementOrPropertyAccessName(node) === "prototype";
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:3946:1`
- `src/compiler/utilities.ts:3966:1`

### guard-0144: src/compiler/utilities.ts:7567:63

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as QualifiedName
~~~~

Exact reason:

~~~~text
a view of Node as QualifiedName in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node.parent as QualifiedName
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:7566:1`

### guard-0145: src/compiler/utilities.ts:7585:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as InstanceofExpression
~~~~

Exact reason:

~~~~text
a view of Node as InstanceofExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isInstanceOfExpression(node: Node): node is InstanceofExpression {
    return isBinaryExpression(node) && node.operatorToken.kind === SyntaxKind.InstanceOfKeyword;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:7590:1`

### guard-0146: src/compiler/utilities.ts:7597:10

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ObjectLiteralExpression
~~~~

Exact reason:

~~~~text
a view of Node as ObjectLiteralExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
expression as ObjectLiteralExpression
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:7595:1`

### guard-0147: src/compiler/utilities.ts:7603:10

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ArrayLiteralExpression
~~~~

Exact reason:

~~~~text
a view of Node as ArrayLiteralExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
expression as ArrayLiteralExpression
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:7601:1`

### guard-0148: src/compiler/utilities.ts:8353:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ElementAccessExpression
~~~~

Exact reason:

~~~~text
a view of Node as ElementAccessExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isAccessExpression(node: Node): node is AccessExpression {
    return node.kind === SyntaxKind.PropertyAccessExpression || node.kind === SyntaxKind.ElementAccessExpression;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:3920:1`
- `src/compiler/utilities.ts:8372:1`

### guard-0149: src/compiler/utilities.ts:8422:25

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Expression as PostfixUnaryExpression
~~~~

Exact reason:

~~~~text
a view of Expression as PostfixUnaryExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as PostfixUnaryExpression
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:8418:1`

### guard-0150: src/compiler/utilities.ts:8426:25

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Expression as BinaryExpression
~~~~

Exact reason:

~~~~text
a view of Expression as BinaryExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as BinaryExpression
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:8418:1`

### guard-0151: src/compiler/utilities.ts:8430:25

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Expression as ConditionalExpression
~~~~

Exact reason:

~~~~text
a view of Expression as ConditionalExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ConditionalExpression
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:8418:1`

### guard-0152: src/compiler/utilitiesPublic.ts:1380:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as CallChain
~~~~

Exact reason:

~~~~text
a view of Node as CallChain in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isOptionalChain(node: Node): node is PropertyAccessChain | ElementAccessChain | CallChain | NonNullChain {
    const kind = node.kind;
    return !!(node.flags & NodeFlags.OptionalChain) &&
        (kind === SyntaxKind.PropertyAccessExpression
            || kind === SyntaxKind.ElementAccessExpression
            || kind === SyntaxKind.CallExpression
            || kind === SyntaxKind.NonNullExpression);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1390:1`
- `src/compiler/utilitiesPublic.ts:1416:1`

### guard-0153: src/compiler/utilitiesPublic.ts:1380:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ElementAccessChain
~~~~

Exact reason:

~~~~text
a view of Node as ElementAccessChain in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isOptionalChain(node: Node): node is PropertyAccessChain | ElementAccessChain | CallChain | NonNullChain {
    const kind = node.kind;
    return !!(node.flags & NodeFlags.OptionalChain) &&
        (kind === SyntaxKind.PropertyAccessExpression
            || kind === SyntaxKind.ElementAccessExpression
            || kind === SyntaxKind.CallExpression
            || kind === SyntaxKind.NonNullExpression);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1390:1`
- `src/compiler/utilitiesPublic.ts:1416:1`

### guard-0154: src/compiler/utilitiesPublic.ts:1380:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as NonNullChain
~~~~

Exact reason:

~~~~text
a view of Node as NonNullChain in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isOptionalChain(node: Node): node is PropertyAccessChain | ElementAccessChain | CallChain | NonNullChain {
    const kind = node.kind;
    return !!(node.flags & NodeFlags.OptionalChain) &&
        (kind === SyntaxKind.PropertyAccessExpression
            || kind === SyntaxKind.ElementAccessExpression
            || kind === SyntaxKind.CallExpression
            || kind === SyntaxKind.NonNullExpression);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1390:1`
- `src/compiler/utilitiesPublic.ts:1416:1`

### guard-0155: src/compiler/utilitiesPublic.ts:1380:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as PropertyAccessChain
~~~~

Exact reason:

~~~~text
a view of Node as PropertyAccessChain in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isOptionalChain(node: Node): node is PropertyAccessChain | ElementAccessChain | CallChain | NonNullChain {
    const kind = node.kind;
    return !!(node.flags & NodeFlags.OptionalChain) &&
        (kind === SyntaxKind.PropertyAccessExpression
            || kind === SyntaxKind.ElementAccessExpression
            || kind === SyntaxKind.CallExpression
            || kind === SyntaxKind.NonNullExpression);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1390:1`
- `src/compiler/utilitiesPublic.ts:1416:1`

### guard-0156: src/compiler/utilitiesPublic.ts:1390:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as CallChainRoot
~~~~

Exact reason:

~~~~text
a view of Node as CallChainRoot in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isOptionalChainRoot(node: Node): node is OptionalChainRoot {
    return isOptionalChain(node) && !isNonNullExpression(node) && !!node.questionDotToken;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1399:1`
- `src/compiler/utilitiesPublic.ts:1416:1`

### guard-0157: src/compiler/utilitiesPublic.ts:1390:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ElementAccessChainRoot
~~~~

Exact reason:

~~~~text
a view of Node as ElementAccessChainRoot in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isOptionalChainRoot(node: Node): node is OptionalChainRoot {
    return isOptionalChain(node) && !isNonNullExpression(node) && !!node.questionDotToken;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1399:1`
- `src/compiler/utilitiesPublic.ts:1416:1`

### guard-0158: src/compiler/utilitiesPublic.ts:1390:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as PropertyAccessChainRoot
~~~~

Exact reason:

~~~~text
a view of Node as PropertyAccessChainRoot in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isOptionalChainRoot(node: Node): node is OptionalChainRoot {
    return isOptionalChain(node) && !isNonNullExpression(node) && !!node.questionDotToken;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1399:1`
- `src/compiler/utilitiesPublic.ts:1416:1`

### guard-0159: src/compiler/utilitiesPublic.ts:1423:58

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as BinaryExpression
~~~~

Exact reason:

~~~~text
a view of Node as BinaryExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as BinaryExpression
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1422:1`

### guard-0160: src/compiler/utilitiesPublic.ts:1498:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as LiteralExpression
~~~~

Exact reason:

~~~~text
a view of Node as LiteralExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isLiteralExpression(node: Node): node is LiteralExpression {
    return isLiteralKind(node.kind);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/destructuring.ts:218:1`
- `src/compiler/utilitiesPublic.ts:2071:1`

### guard-0161: src/compiler/utilitiesPublic.ts:1600:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as PrivateIdentifierAutoAccessorPropertyDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as PrivateIdentifierAutoAccessorPropertyDeclaration in a graph program, which the cycle finder doesn't relate yet through its field original (Node | undefined as Node | undefined) through its field parent (Node as Node) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
// Private Identifiers
/** @internal */
export function isPrivateIdentifierClassElementDeclaration(node: Node): node is PrivateClassElementDeclaration {
    return (isPropertyDeclaration(node) || isMethodOrAccessor(node)) && isPrivateIdentifier(node.name);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:870:1`

### guard-0162: src/compiler/utilitiesPublic.ts:1600:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as PrivateIdentifierGetAccessorDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as PrivateIdentifierGetAccessorDeclaration in a graph program, which the cycle finder doesn't relate yet through its field original (Node | undefined as Node | undefined) through its field parent (Node as Node) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
// Private Identifiers
/** @internal */
export function isPrivateIdentifierClassElementDeclaration(node: Node): node is PrivateClassElementDeclaration {
    return (isPropertyDeclaration(node) || isMethodOrAccessor(node)) && isPrivateIdentifier(node.name);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:870:1`

### guard-0163: src/compiler/utilitiesPublic.ts:1600:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as PrivateIdentifierMethodDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as PrivateIdentifierMethodDeclaration in a graph program, which the cycle finder doesn't relate yet through its field original (Node | undefined as Node | undefined) through its field parent (Node as Node) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
// Private Identifiers
/** @internal */
export function isPrivateIdentifierClassElementDeclaration(node: Node): node is PrivateClassElementDeclaration {
    return (isPropertyDeclaration(node) || isMethodOrAccessor(node)) && isPrivateIdentifier(node.name);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:870:1`

### guard-0164: src/compiler/utilitiesPublic.ts:1600:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as PrivateIdentifierPropertyDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as PrivateIdentifierPropertyDeclaration in a graph program, which the cycle finder doesn't relate yet through its field original (Node | undefined as Node | undefined) through its field parent (Node as Node) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
// Private Identifiers
/** @internal */
export function isPrivateIdentifierClassElementDeclaration(node: Node): node is PrivateClassElementDeclaration {
    return (isPropertyDeclaration(node) || isMethodOrAccessor(node)) && isPrivateIdentifier(node.name);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:870:1`

### guard-0165: src/compiler/utilitiesPublic.ts:1600:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as PrivateIdentifierSetAccessorDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as PrivateIdentifierSetAccessorDeclaration in a graph program, which the cycle finder doesn't relate yet through its field original (Node | undefined as Node | undefined) through its field parent (Node as Node) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
// Private Identifiers
/** @internal */
export function isPrivateIdentifierClassElementDeclaration(node: Node): node is PrivateClassElementDeclaration {
    return (isPropertyDeclaration(node) || isMethodOrAccessor(node)) && isPrivateIdentifier(node.name);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:870:1`

### guard-0166: src/compiler/utilitiesPublic.ts:1647:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as AbstractKeyword
~~~~

Exact reason:

~~~~text
a view of Node as AbstractKeyword in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isModifier(node: Node): node is Modifier {
    return isModifierKind(node.kind);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1778:1`

### guard-0167: src/compiler/utilitiesPublic.ts:1647:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as AccessorKeyword
~~~~

Exact reason:

~~~~text
a view of Node as AccessorKeyword in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isModifier(node: Node): node is Modifier {
    return isModifierKind(node.kind);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1778:1`

### guard-0168: src/compiler/utilitiesPublic.ts:1647:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as AsyncKeyword
~~~~

Exact reason:

~~~~text
a view of Node as AsyncKeyword in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isModifier(node: Node): node is Modifier {
    return isModifierKind(node.kind);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1778:1`

### guard-0169: src/compiler/utilitiesPublic.ts:1647:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ConstKeyword
~~~~

Exact reason:

~~~~text
a view of Node as ConstKeyword in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isModifier(node: Node): node is Modifier {
    return isModifierKind(node.kind);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1778:1`

### guard-0170: src/compiler/utilitiesPublic.ts:1647:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as DeclareKeyword
~~~~

Exact reason:

~~~~text
a view of Node as DeclareKeyword in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isModifier(node: Node): node is Modifier {
    return isModifierKind(node.kind);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1778:1`

### guard-0171: src/compiler/utilitiesPublic.ts:1647:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as DefaultKeyword
~~~~

Exact reason:

~~~~text
a view of Node as DefaultKeyword in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isModifier(node: Node): node is Modifier {
    return isModifierKind(node.kind);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1778:1`

### guard-0172: src/compiler/utilitiesPublic.ts:1647:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ExportKeyword
~~~~

Exact reason:

~~~~text
a view of Node as ExportKeyword in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isModifier(node: Node): node is Modifier {
    return isModifierKind(node.kind);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1778:1`

### guard-0173: src/compiler/utilitiesPublic.ts:1647:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as InKeyword
~~~~

Exact reason:

~~~~text
a view of Node as InKeyword in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isModifier(node: Node): node is Modifier {
    return isModifierKind(node.kind);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1778:1`

### guard-0174: src/compiler/utilitiesPublic.ts:1647:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as OutKeyword
~~~~

Exact reason:

~~~~text
a view of Node as OutKeyword in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isModifier(node: Node): node is Modifier {
    return isModifierKind(node.kind);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1778:1`

### guard-0175: src/compiler/utilitiesPublic.ts:1647:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as OverrideKeyword
~~~~

Exact reason:

~~~~text
a view of Node as OverrideKeyword in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isModifier(node: Node): node is Modifier {
    return isModifierKind(node.kind);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1778:1`

### guard-0176: src/compiler/utilitiesPublic.ts:1647:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as PrivateKeyword
~~~~

Exact reason:

~~~~text
a view of Node as PrivateKeyword in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isModifier(node: Node): node is Modifier {
    return isModifierKind(node.kind);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1778:1`

### guard-0177: src/compiler/utilitiesPublic.ts:1647:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ProtectedKeyword
~~~~

Exact reason:

~~~~text
a view of Node as ProtectedKeyword in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isModifier(node: Node): node is Modifier {
    return isModifierKind(node.kind);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1778:1`

### guard-0178: src/compiler/utilitiesPublic.ts:1647:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as PublicKeyword
~~~~

Exact reason:

~~~~text
a view of Node as PublicKeyword in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isModifier(node: Node): node is Modifier {
    return isModifierKind(node.kind);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1778:1`

### guard-0179: src/compiler/utilitiesPublic.ts:1647:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ReadonlyKeyword
~~~~

Exact reason:

~~~~text
a view of Node as ReadonlyKeyword in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isModifier(node: Node): node is Modifier {
    return isModifierKind(node.kind);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1778:1`

### guard-0180: src/compiler/utilitiesPublic.ts:1647:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as StaticKeyword
~~~~

Exact reason:

~~~~text
a view of Node as StaticKeyword in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isModifier(node: Node): node is Modifier {
    return isModifierKind(node.kind);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1778:1`

### guard-0181: src/compiler/utilitiesPublic.ts:1651:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as Identifier
~~~~

Exact reason:

~~~~text
a view of Node as Identifier in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isEntityName(node: Node): node is EntityName {
    const kind = node.kind;
    return kind === SyntaxKind.QualifiedName
        || kind === SyntaxKind.Identifier;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/declarations/diagnostics.ts:615:1`

### guard-0182: src/compiler/utilitiesPublic.ts:1675:1

Observations: {'head': 6}. New in head: True.

Source and target as named by the guard:

~~~~text
Node | undefined as ArrowFunction
~~~~

Exact reason:

~~~~text
a view of Node | undefined as ArrowFunction in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
// Functions

export function isFunctionLike(node: Node | undefined): node is SignatureDeclaration {
    return !!node && isFunctionLikeKind(node.kind);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:12218:1`
- `src/compiler/utilities.ts:3066:1`
- `src/compiler/utilities.ts:3125:1`
- `src/compiler/utilities.ts:3140:1`
- `src/compiler/utilities.ts:4783:1`
- `src/compiler/utilitiesPublic.ts:603:1`

### guard-0183: src/compiler/utilitiesPublic.ts:1680:1

Observations: {'head': 3}. New in head: True.

Source and target as named by the guard:

~~~~text
Node | undefined as ArrowFunction
~~~~

Exact reason:

~~~~text
a view of Node | undefined as ArrowFunction in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isFunctionLikeOrClassStaticBlockDeclaration(node: Node | undefined): node is SignatureDeclaration | ClassStaticBlockDeclaration {
    return !!node && (isFunctionLikeKind(node.kind) || isClassStaticBlockDeclaration(node));
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2118:1`
- `src/compiler/utilities.ts:2203:1`
- `src/compiler/utilities.ts:3150:1`

### guard-0184: src/compiler/utilitiesPublic.ts:1685:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ArrowFunction
~~~~

Exact reason:

~~~~text
a view of Node as ArrowFunction in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isFunctionLikeDeclaration(node: Node): node is FunctionLikeDeclaration {
    return node && isFunctionLikeDeclarationKind(node.kind);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:3130:1`

### guard-0185: src/compiler/utilitiesPublic.ts:1732:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ClassElement
~~~~

Exact reason:

~~~~text
a view of Node as ClassElement in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
// Classes
export function isClassElement(node: Node): node is ClassElement {
    const kind = node.kind;
    return kind === SyntaxKind.Constructor
        || kind === SyntaxKind.PropertyDeclaration
        || kind === SyntaxKind.MethodDeclaration
        || kind === SyntaxKind.GetAccessor
        || kind === SyntaxKind.SetAccessor
        || kind === SyntaxKind.IndexSignature
        || kind === SyntaxKind.ClassStaticBlockDeclaration
        || kind === SyntaxKind.SemicolonClassElement;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:7198:1`
- `src/compiler/utilitiesPublic.ts:1794:1`

### guard-0186: src/compiler/utilitiesPublic.ts:1744:1

Observations: {'head': 10}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ClassDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as ClassDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isClassLike(node: Node): node is ClassLikeDeclaration {
    return node && (node.kind === SyntaxKind.ClassDeclaration || node.kind === SyntaxKind.ClassExpression);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/typeSerializer.ts:132:1`
- `src/compiler/utilities.ts:3135:1`
- `src/compiler/utilities.ts:3140:1`
- `src/compiler/utilities.ts:3155:1`
- `src/compiler/utilities.ts:5227:1`
- `src/compiler/utilities.ts:7453:1`
- `src/compiler/utilities.ts:8273:1`
- `src/compiler/utilities.ts:8323:1`

### guard-0187: src/compiler/utilitiesPublic.ts:1744:1

Observations: {'head': 10}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ClassExpression
~~~~

Exact reason:

~~~~text
a view of Node as ClassExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isClassLike(node: Node): node is ClassLikeDeclaration {
    return node && (node.kind === SyntaxKind.ClassDeclaration || node.kind === SyntaxKind.ClassExpression);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/typeSerializer.ts:132:1`
- `src/compiler/utilities.ts:3135:1`
- `src/compiler/utilities.ts:3140:1`
- `src/compiler/utilities.ts:3155:1`
- `src/compiler/utilities.ts:5227:1`
- `src/compiler/utilities.ts:7453:1`
- `src/compiler/utilities.ts:8273:1`
- `src/compiler/utilities.ts:8323:1`

### guard-0188: src/compiler/utilitiesPublic.ts:1782:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as TypeElement
~~~~

Exact reason:

~~~~text
a view of Node as TypeElement in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isTypeElement(node: Node): node is TypeElement {
    const kind = node.kind;
    return kind === SyntaxKind.ConstructSignature
        || kind === SyntaxKind.CallSignature
        || kind === SyntaxKind.PropertySignature
        || kind === SyntaxKind.MethodSignature
        || kind === SyntaxKind.IndexSignature
        || kind === SyntaxKind.GetAccessor
        || kind === SyntaxKind.SetAccessor
        || kind === SyntaxKind.NotEmittedTypeElement;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1794:1`

### guard-0189: src/compiler/utilitiesPublic.ts:1968:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as CallExpression
~~~~

Exact reason:

~~~~text
a view of Node as CallExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isCallLikeExpression(node: Node): node is CallLikeExpression {
    switch (node.kind) {
        case SyntaxKind.CallExpression:
        case SyntaxKind.NewExpression:
        case SyntaxKind.TaggedTemplateExpression:
        case SyntaxKind.Decorator:
        case SyntaxKind.JsxOpeningElement:
        case SyntaxKind.JsxSelfClosingElement:
        case SyntaxKind.JsxOpeningFragment:
            return true;
        case SyntaxKind.BinaryExpression:
            return (node as BinaryExpression).operatorToken.kind === SyntaxKind.InstanceOfKeyword;
        default:
            return false;
    }
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1964:1`

### guard-0190: src/compiler/utilitiesPublic.ts:1968:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as Decorator
~~~~

Exact reason:

~~~~text
a view of Node as Decorator in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as NamedDeclaration) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isCallLikeExpression(node: Node): node is CallLikeExpression {
    switch (node.kind) {
        case SyntaxKind.CallExpression:
        case SyntaxKind.NewExpression:
        case SyntaxKind.TaggedTemplateExpression:
        case SyntaxKind.Decorator:
        case SyntaxKind.JsxOpeningElement:
        case SyntaxKind.JsxSelfClosingElement:
        case SyntaxKind.JsxOpeningFragment:
            return true;
        case SyntaxKind.BinaryExpression:
            return (node as BinaryExpression).operatorToken.kind === SyntaxKind.InstanceOfKeyword;
        default:
            return false;
    }
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1964:1`

### guard-0191: src/compiler/utilitiesPublic.ts:1968:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as InstanceofExpression
~~~~

Exact reason:

~~~~text
a view of Node as InstanceofExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isCallLikeExpression(node: Node): node is CallLikeExpression {
    switch (node.kind) {
        case SyntaxKind.CallExpression:
        case SyntaxKind.NewExpression:
        case SyntaxKind.TaggedTemplateExpression:
        case SyntaxKind.Decorator:
        case SyntaxKind.JsxOpeningElement:
        case SyntaxKind.JsxSelfClosingElement:
        case SyntaxKind.JsxOpeningFragment:
            return true;
        case SyntaxKind.BinaryExpression:
            return (node as BinaryExpression).operatorToken.kind === SyntaxKind.InstanceOfKeyword;
        default:
            return false;
    }
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1964:1`

### guard-0192: src/compiler/utilitiesPublic.ts:1968:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JsxOpeningElement
~~~~

Exact reason:

~~~~text
a view of Node as JsxOpeningElement in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as JsxElement) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isCallLikeExpression(node: Node): node is CallLikeExpression {
    switch (node.kind) {
        case SyntaxKind.CallExpression:
        case SyntaxKind.NewExpression:
        case SyntaxKind.TaggedTemplateExpression:
        case SyntaxKind.Decorator:
        case SyntaxKind.JsxOpeningElement:
        case SyntaxKind.JsxSelfClosingElement:
        case SyntaxKind.JsxOpeningFragment:
            return true;
        case SyntaxKind.BinaryExpression:
            return (node as BinaryExpression).operatorToken.kind === SyntaxKind.InstanceOfKeyword;
        default:
            return false;
    }
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1964:1`

### guard-0193: src/compiler/utilitiesPublic.ts:1968:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JsxOpeningFragment
~~~~

Exact reason:

~~~~text
a view of Node as JsxOpeningFragment in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as JsxFragment) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isCallLikeExpression(node: Node): node is CallLikeExpression {
    switch (node.kind) {
        case SyntaxKind.CallExpression:
        case SyntaxKind.NewExpression:
        case SyntaxKind.TaggedTemplateExpression:
        case SyntaxKind.Decorator:
        case SyntaxKind.JsxOpeningElement:
        case SyntaxKind.JsxSelfClosingElement:
        case SyntaxKind.JsxOpeningFragment:
            return true;
        case SyntaxKind.BinaryExpression:
            return (node as BinaryExpression).operatorToken.kind === SyntaxKind.InstanceOfKeyword;
        default:
            return false;
    }
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1964:1`

### guard-0194: src/compiler/utilitiesPublic.ts:1968:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JsxSelfClosingElement
~~~~

Exact reason:

~~~~text
a view of Node as JsxSelfClosingElement in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isCallLikeExpression(node: Node): node is CallLikeExpression {
    switch (node.kind) {
        case SyntaxKind.CallExpression:
        case SyntaxKind.NewExpression:
        case SyntaxKind.TaggedTemplateExpression:
        case SyntaxKind.Decorator:
        case SyntaxKind.JsxOpeningElement:
        case SyntaxKind.JsxSelfClosingElement:
        case SyntaxKind.JsxOpeningFragment:
            return true;
        case SyntaxKind.BinaryExpression:
            return (node as BinaryExpression).operatorToken.kind === SyntaxKind.InstanceOfKeyword;
        default:
            return false;
    }
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1964:1`

### guard-0195: src/compiler/utilitiesPublic.ts:1968:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as NewExpression
~~~~

Exact reason:

~~~~text
a view of Node as NewExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isCallLikeExpression(node: Node): node is CallLikeExpression {
    switch (node.kind) {
        case SyntaxKind.CallExpression:
        case SyntaxKind.NewExpression:
        case SyntaxKind.TaggedTemplateExpression:
        case SyntaxKind.Decorator:
        case SyntaxKind.JsxOpeningElement:
        case SyntaxKind.JsxSelfClosingElement:
        case SyntaxKind.JsxOpeningFragment:
            return true;
        case SyntaxKind.BinaryExpression:
            return (node as BinaryExpression).operatorToken.kind === SyntaxKind.InstanceOfKeyword;
        default:
            return false;
    }
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1964:1`

### guard-0196: src/compiler/utilitiesPublic.ts:1968:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as TaggedTemplateExpression
~~~~

Exact reason:

~~~~text
a view of Node as TaggedTemplateExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isCallLikeExpression(node: Node): node is CallLikeExpression {
    switch (node.kind) {
        case SyntaxKind.CallExpression:
        case SyntaxKind.NewExpression:
        case SyntaxKind.TaggedTemplateExpression:
        case SyntaxKind.Decorator:
        case SyntaxKind.JsxOpeningElement:
        case SyntaxKind.JsxSelfClosingElement:
        case SyntaxKind.JsxOpeningFragment:
            return true;
        case SyntaxKind.BinaryExpression:
            return (node as BinaryExpression).operatorToken.kind === SyntaxKind.InstanceOfKeyword;
        default:
            return false;
    }
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1964:1`

### guard-0197: src/compiler/utilitiesPublic.ts:1979:21

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as BinaryExpression
~~~~

Exact reason:

~~~~text
a view of Node as BinaryExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as BinaryExpression
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1968:1`

### guard-0198: src/compiler/utilitiesPublic.ts:1995:1

Observations: {'head': 3}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as LeftHandSideExpression
~~~~

Exact reason:

~~~~text
a view of Node as LeftHandSideExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isLeftHandSideExpression(node: Node): node is LeftHandSideExpression {
    return isLeftHandSideExpressionKind(skipPartiallyEmittedExpressions(node).kind);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:7469:1`
- `src/compiler/utilities.ts:7471:1`
- `src/compiler/utilities.ts:7473:1`

### guard-0199: src/compiler/utilitiesPublic.ts:2064:21

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as PrefixUnaryExpression
~~~~

Exact reason:

~~~~text
a view of Node as PrefixUnaryExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
expr as PrefixUnaryExpression
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:2059:1`

### guard-0200: src/compiler/utilitiesPublic.ts:2065:18

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as PrefixUnaryExpression
~~~~

Exact reason:

~~~~text
a view of Node as PrefixUnaryExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
expr as PrefixUnaryExpression
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:2059:1`

### guard-0201: src/compiler/utilitiesPublic.ts:2086:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as Expression
~~~~

Exact reason:

~~~~text
a view of Node as Expression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/**
 * Determines whether a node is an expression based only on its kind.
 */
export function isExpression(node: Node): node is Expression {
    return isExpressionKind(skipPartiallyEmittedExpressions(node).kind);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:2170:1`

### guard-0202: src/compiler/utilitiesPublic.ts:2108:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as AsExpression
~~~~

Exact reason:

~~~~text
a view of Node as AsExpression in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isAssertionExpression(node: Node): node is AssertionExpression {
    const kind = node.kind;
    return kind === SyntaxKind.TypeAssertionExpression
        || kind === SyntaxKind.AsExpression;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:11450:1`

### guard-0203: src/compiler/utilitiesPublic.ts:2108:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as TypeAssertion
~~~~

Exact reason:

~~~~text
a view of Node as TypeAssertion in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isAssertionExpression(node: Node): node is AssertionExpression {
    const kind = node.kind;
    return kind === SyntaxKind.TypeAssertionExpression
        || kind === SyntaxKind.AsExpression;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:11450:1`

### guard-0204: src/compiler/utilitiesPublic.ts:2127:69

Observations: {'head': 3}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as LabeledStatement
~~~~

Exact reason:

~~~~text
a view of Node as LabeledStatement in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as LabeledStatement
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:2116:1`
- `src/compiler/utilitiesPublic.ts:2117:1`
- `src/compiler/utilitiesPublic.ts:2118:1`

### guard-0205: src/compiler/utilitiesPublic.ts:2398:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as NamedDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as NamedDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isDeclaration(node: Node): node is NamedDeclaration {
    if (node.kind === SyntaxKind.TypeParameter) {
        return (node.parent && node.parent.kind !== SyntaxKind.JSDocTemplateTag) || isInJSFile(node);
    }

    return isDeclarationKind(node.kind);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:5077:1`
- `src/compiler/utilities.ts:5110:1`

### guard-0206: src/compiler/utilitiesPublic.ts:2426:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as Block
~~~~

Exact reason:

~~~~text
a view of Node as Block in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
function isBlockStatement(node: Node): node is Block {
    if (node.kind !== SyntaxKind.Block) return false;
    if (node.parent !== undefined) {
        if (node.parent.kind === SyntaxKind.TryStatement || node.parent.kind === SyntaxKind.CatchClause) {
            return false;
        }
    }
    return !isFunctionBlock(node);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:2419:1`

### guard-0207: src/compiler/utilitiesPublic.ts:2489:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JsxOpeningElement
~~~~

Exact reason:

~~~~text
a view of Node as JsxOpeningElement in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as JsxElement) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isJsxOpeningLikeElement(node: Node): node is JsxOpeningLikeElement {
    const kind = node.kind;
    return kind === SyntaxKind.JsxOpeningElement
        || kind === SyntaxKind.JsxSelfClosingElement;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:8959:1`

### guard-0208: src/compiler/utilitiesPublic.ts:2489:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JsxSelfClosingElement
~~~~

Exact reason:

~~~~text
a view of Node as JsxSelfClosingElement in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isJsxOpeningLikeElement(node: Node): node is JsxOpeningLikeElement {
    const kind = node.kind;
    return kind === SyntaxKind.JsxOpeningElement
        || kind === SyntaxKind.JsxSelfClosingElement;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:8959:1`

### guard-0209: src/compiler/utilitiesPublic.ts:2552:1

Observations: {'head': 3}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ArrowFunction
~~~~

Exact reason:

~~~~text
a view of Node as ArrowFunction in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/**
 * True if has jsdoc nodes attached to it.
 *
 * @internal
 */
// TODO: GH#19856 Would like to return `node is Node & { jsDoc: JSDoc[] }` but it causes long compile times
export function hasJSDocNodes(node: Node): node is HasJSDoc {
    if (!canHaveJSDoc(node)) return false;

    const { jsDoc } = node as JSDocContainer;
    return !!jsDoc && jsDoc.length > 0;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:11112:1`
- `src/compiler/utilities.ts:12201:1`
- `src/compiler/utilities.ts:1239:1`

### guard-0210: src/compiler/utilitiesPublic.ts:2629:1

Observations: {'head': 4}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDocLink
~~~~

Exact reason:

~~~~text
a view of Node as JSDocLink in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isJSDocLinkLike(node: Node): node is JSDocLink | JSDocLinkCode | JSDocLinkPlain {
    return node.kind === SyntaxKind.JSDocLink || node.kind === SyntaxKind.JSDocLinkCode || node.kind === SyntaxKind.JSDocLinkPlain;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:3629:1`
- `src/compiler/utilitiesPublic.ts:2522:1`

### guard-0211: src/compiler/utilitiesPublic.ts:2629:1

Observations: {'head': 4}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDocLinkCode
~~~~

Exact reason:

~~~~text
a view of Node as JSDocLinkCode in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isJSDocLinkLike(node: Node): node is JSDocLink | JSDocLinkCode | JSDocLinkPlain {
    return node.kind === SyntaxKind.JSDocLink || node.kind === SyntaxKind.JSDocLinkCode || node.kind === SyntaxKind.JSDocLinkPlain;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:3629:1`
- `src/compiler/utilitiesPublic.ts:2522:1`

### guard-0212: src/compiler/utilitiesPublic.ts:2629:1

Observations: {'head': 4}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDocLinkPlain
~~~~

Exact reason:

~~~~text
a view of Node as JSDocLinkPlain in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isJSDocLinkLike(node: Node): node is JSDocLink | JSDocLinkCode | JSDocLinkPlain {
    return node.kind === SyntaxKind.JSDocLink || node.kind === SyntaxKind.JSDocLinkCode || node.kind === SyntaxKind.JSDocLinkPlain;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:3629:1`
- `src/compiler/utilitiesPublic.ts:2522:1`

### guard-0213: src/compiler/utilitiesPublic.ts:947:15

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as NamedDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as NamedDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as NamedDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:946:1`

## union arms

### guard-0214: src/compiler/binder.ts:372:39

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ExportDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as ExportDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ExportDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/binder.ts:350:1`

### guard-0215: src/compiler/core.ts:1782:35

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
ArrayLiteralExpression | FalseLiteral | JsonMinusNumericLiteral | NullLiteral | NumericLiteral | ObjectLiteralExpression | StringLiteral | TrueLiteral as ObjectLiteralExpression
~~~~

Exact reason:

~~~~text
a view of ArrayLiteralExpression | FalseLiteral | JsonMinusNumericLiteral | NullLiteral | NumericLiteral | ObjectLiteralExpression | StringLiteral | TrueLiteral as ObjectLiteralExpression in a graph program, which the cycle finder doesn't relate yet through a source arm (ArrayLiteralExpression as ObjectLiteralExpression)
~~~~

Source text:

~~~~typescript
test(value)
~~~~

Reading: **guard bug: checked tryCast alternative.** The site is tryCast's `test(value)`. The enclosing caller supplies the corresponding node-kind predicate before tryCast returns `ObjectLiteralExpression` (for example getTsConfigObjectLiteralExpression supplies isObjectLiteralExpression, and getNamespaceDeclarationNode supplies isNamespaceImport/isNamespaceExport). `ArrayLiteralExpression` is an excluded alternative, not a receiver admitted by that successful check. The leaf types are disjoint; demanding this cross-arm pair is a false refusal.

Owner units:

- `src/compiler/utilities.ts:3104:1`

### guard-0216: src/compiler/core.ts:1782:35

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
NamedImports | NamespaceImport as NamespaceImport
~~~~

Exact reason:

~~~~text
a view of NamedImports | NamespaceImport as NamespaceImport in a graph program, which the cycle finder doesn't relate yet through a source arm (NamedImports as NamespaceImport)
~~~~

Source text:

~~~~typescript
test(value)
~~~~

Reading: **guard bug: checked tryCast alternative.** The site is tryCast's `test(value)`. The enclosing caller supplies the corresponding node-kind predicate before tryCast returns `NamespaceImport` (for example getTsConfigObjectLiteralExpression supplies isObjectLiteralExpression, and getNamespaceDeclarationNode supplies isNamespaceImport/isNamespaceExport). `NamedImports` is an excluded alternative, not a receiver admitted by that successful check. The leaf types are disjoint; demanding this cross-arm pair is a false refusal.

Owner units:

- `src/compiler/utilities.ts:4425:1`

### guard-0217: src/compiler/core.ts:1787:32

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
NamedExports | NamespaceExport as NamedExports
~~~~

Exact reason:

~~~~text
a view of NamedExports | NamespaceExport as NamedExports in a graph program, which the cycle finder doesn't relate yet through a source arm (NamedExports as NamedExports) through its field parent (ExportDeclaration as ExportDeclaration) through its field parent (ModuleBlock | SourceFile as ModuleBlock | SourceFile) through a source arm (ModuleBlock as ModuleBlock | SourceFile) through an arm (ModuleBlock as ModuleBlock) through its field parent (ModuleDeclaration as ModuleDeclaration) through its field parent (SourceFile | ModuleBody as SourceFile | ModuleBody) through a source arm (Identifier as SourceFile | ModuleBody) through an arm (Identifier as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
test(value)
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/utilities.ts:179:1`

### guard-0218: src/compiler/factory/nodeTests.ts:1082:1

Observations: {'head': 3}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDoc
~~~~

Exact reason:

~~~~text
a view of Node as JSDoc in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as HasJSDoc) through an arm (Node as ArrowFunction) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isJSDoc(node: Node): node is JSDoc {
    return node.kind === SyntaxKind.JSDoc;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:4699:1`
- `src/compiler/utilities.ts:4765:1`
- `src/compiler/utilities.ts:4823:1`

### guard-0219: src/compiler/factory/nodeTests.ts:1132:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDocOverloadTag
~~~~

Exact reason:

~~~~text
a view of Node as JSDocOverloadTag in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as JSDoc) through its field parent (Node as HasJSDoc) through an arm (Node as ArrowFunction) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isJSDocOverloadTag(node: Node): node is JSDocOverloadTag {
    return node.kind === SyntaxKind.JSDocOverloadTag;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:4778:1`

### guard-0220: src/compiler/factory/nodeTests.ts:1144:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDocEnumTag
~~~~

Exact reason:

~~~~text
a view of Node as JSDocEnumTag in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as JSDoc) through its field parent (Node as HasJSDoc) through an arm (Node as ArrowFunction) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isJSDocEnumTag(node: Node): node is JSDocEnumTag {
    return node.kind === SyntaxKind.JSDocEnumTag;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1180:1`

### guard-0221: src/compiler/factory/nodeTests.ts:1192:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDocImportTag
~~~~

Exact reason:

~~~~text
a view of Node as JSDocImportTag in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as JSDoc) through its field parent (Node as HasJSDoc) through an arm (Node as ArrowFunction) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isJSDocImportTag(node: Node): node is JSDocImportTag {
    return node.kind === SyntaxKind.JSDocImportTag;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/program.ts:937:1`

### guard-0222: src/compiler/factory/nodeTests.ts:403:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as TypeParameterDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as TypeParameterDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as InferTypeNode | DeclarationWithTypeParameterChildren) through an arm (Node as ArrowFunction) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
// Signature elements

export function isTypeParameterDeclaration(node: Node): node is TypeParameterDeclaration {
    return node.kind === SyntaxKind.TypeParameter;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:1097:1`

### guard-0223: src/compiler/factory/nodeTests.ts:408:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ParameterDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as ParameterDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as SignatureDeclaration) through an arm (Node as ArrowFunction) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
// TODO(rbuckton): Rename to 'isParameterDeclaration'
export function isParameter(node: Node): node is ParameterDeclaration {
    return node.kind === SyntaxKind.Parameter;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:7294:1`
- `src/compiler/utilitiesPublic.ts:615:1`

### guard-0224: src/compiler/factory/nodeTests.ts:418:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as PropertySignature
~~~~

Exact reason:

~~~~text
a view of Node as PropertySignature in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as InterfaceDeclaration | TypeLiteralNode) through an arm (Node as InterfaceDeclaration) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
// TypeMember

export function isPropertySignature(node: Node): node is PropertySignature {
    return node.kind === SyntaxKind.PropertySignature;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:4783:1`

### guard-0225: src/compiler/factory/nodeTests.ts:422:1

Observations: {'head': 6}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as PropertyDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as PropertyDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ClassLikeDeclaration) through an arm (Node as ClassDeclaration) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isPropertyDeclaration(node: Node): node is PropertyDeclaration {
    return node.kind === SyntaxKind.PropertyDeclaration;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/classFields.ts:3354:1`
- `src/compiler/transformers/utilities.ts:612:1`
- `src/compiler/transformers/utilities.ts:618:1`
- `src/compiler/utilities.ts:7889:1`
- `src/compiler/utilitiesPublic.ts:1600:1`
- `src/compiler/utilitiesPublic.ts:1752:1`

### guard-0226: src/compiler/factory/nodeTests.ts:430:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as MethodDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as MethodDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ObjectLiteralExpression | ClassLikeDeclaration) through an arm (Node as ClassDeclaration) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isMethodDeclaration(node: Node): node is MethodDeclaration {
    return node.kind === SyntaxKind.MethodDeclaration;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:7889:1`

### guard-0227: src/compiler/factory/nodeTests.ts:434:1

Observations: {'head': 6}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ClassStaticBlockDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as ClassStaticBlockDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ClassDeclaration | ClassExpression) through an arm (Node as ClassDeclaration) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isClassStaticBlockDeclaration(node: Node): node is ClassStaticBlockDeclaration {
    return node.kind === SyntaxKind.ClassStaticBlockDeclaration;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/classFields.ts:3358:1`
- `src/compiler/transformers/classThis.ts:72:1`
- `src/compiler/transformers/namedEvaluation.ts:139:1`
- `src/compiler/transformers/utilities.ts:593:1`
- `src/compiler/utilities.ts:7198:1`
- `src/compiler/utilitiesPublic.ts:1680:1`

### guard-0228: src/compiler/factory/nodeTests.ts:446:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as SetAccessorDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as SetAccessorDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as InterfaceDeclaration | ObjectLiteralExpression | TypeLiteralNode | ClassLikeDeclaration) through an arm (Node as ClassDeclaration) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isSetAccessorDeclaration(node: Node): node is SetAccessorDeclaration {
    return node.kind === SyntaxKind.SetAccessor;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:8043:1`

### guard-0229: src/compiler/factory/nodeTests.ts:570:1

Observations: {'head': 3}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as BindingElement
~~~~

Exact reason:

~~~~text
a view of Node as BindingElement in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as BindingPattern) through an arm (Node as ArrayBindingPattern) through its field parent (Node as BindingElement | ParameterDeclaration | VariableDeclaration) through an arm (Node as ParameterDeclaration) through its field parent (Node as SignatureDeclaration) through an arm (Node as ArrowFunction) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isBindingElement(node: Node): node is BindingElement {
    return node.kind === SyntaxKind.BindingElement;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:3878:1`
- `src/compiler/utilitiesPublic.ts:1004:1`
- `src/compiler/utilitiesPublic.ts:644:1`

### guard-0230: src/compiler/factory/nodeTests.ts:792:1

Observations: {'head': 8}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as VariableDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as VariableDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as CatchClause | VariableDeclarationList) through an arm (Node as CatchClause) through its field parent (Node as TryStatement) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isVariableDeclaration(node: Node): node is VariableDeclaration {
    return node.kind === SyntaxKind.VariableDeclaration;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/declarations/diagnostics.ts:147:1`
- `src/compiler/utilities.ts:3031:1`
- `src/compiler/utilities.ts:3418:1`
- `src/compiler/utilities.ts:3896:1`
- `src/compiler/utilities.ts:4027:1`
- `src/compiler/utilities.ts:4319:1`
- `src/compiler/utilities.ts:8022:1`
- `src/compiler/utilitiesPublic.ts:1872:1`

### guard-0231: src/compiler/factory/nodeTests.ts:796:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as VariableDeclarationList
~~~~

Exact reason:

~~~~text
a view of Node as VariableDeclarationList in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ForInStatement | ForOfStatement | ForStatement | VariableStatement) through an arm (Node as ForInStatement) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isVariableDeclarationList(node: Node): node is VariableDeclarationList {
    return node.kind === SyntaxKind.VariableDeclarationList;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/esnext.ts:774:1`
- `src/compiler/utilitiesPublic.ts:2170:1`

### guard-0232: src/compiler/factory/nodeTests.ts:820:1

Observations: {'head': 7}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ModuleDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as ModuleDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isModuleDeclaration(node: Node): node is ModuleDeclaration {
    return node.kind === SyntaxKind.ModuleDeclaration;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2079:1`
- `src/compiler/utilities.ts:2084:1`
- `src/compiler/utilities.ts:2089:1`
- `src/compiler/utilities.ts:2099:1`
- `src/compiler/utilities.ts:2149:1`
- `src/compiler/utilities.ts:2313:1`
- `src/compiler/utilities.ts:4524:1`

### guard-0233: src/compiler/factory/nodeTests.ts:836:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ImportEqualsDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as ImportEqualsDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isImportEqualsDeclaration(node: Node): node is ImportEqualsDeclaration {
    return node.kind === SyntaxKind.ImportEqualsDeclaration;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/program.ts:959:1`

### guard-0234: src/compiler/factory/nodeTests.ts:840:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ImportDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as ImportDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isImportDeclaration(node: Node): node is ImportDeclaration {
    return node.kind === SyntaxKind.ImportDeclaration;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/program.ts:937:1`
- `src/compiler/utilities.ts:12088:1`

### guard-0235: src/compiler/factory/nodeTests.ts:870:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as NamespaceImport
~~~~

Exact reason:

~~~~text
a view of Node as NamespaceImport in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ImportClause) through its field parent (Node as ImportDeclaration | JSDocImportTag) through an arm (Node as ImportDeclaration) through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isNamespaceImport(node: Node): node is NamespaceImport {
    return node.kind === SyntaxKind.NamespaceImport;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:4425:1`

### guard-0236: src/compiler/factory/nodeTests.ts:874:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as NamespaceExport
~~~~

Exact reason:

~~~~text
a view of Node as NamespaceExport in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ExportDeclaration) through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isNamespaceExport(node: Node): node is NamespaceExport {
    return node.kind === SyntaxKind.NamespaceExport;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:1296:1`
- `src/compiler/utilities.ts:3766:1`

### guard-0237: src/compiler/factory/nodeTests.ts:882:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ImportSpecifier
~~~~

Exact reason:

~~~~text
a view of Node as ImportSpecifier in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as NamedImports) through its field parent (Node as ImportClause) through its field parent (Node as ImportDeclaration | JSDocImportTag) through an arm (Node as ImportDeclaration) through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isImportSpecifier(node: Node): node is ImportSpecifier {
    return node.kind === SyntaxKind.ImportSpecifier;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1532:1`

### guard-0238: src/compiler/factory/nodeTests.ts:890:1

Observations: {'head': 4}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ExportDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as ExportDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isExportDeclaration(node: Node): node is ExportDeclaration {
    return node.kind === SyntaxKind.ExportDeclaration;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/program.ts:937:1`
- `src/compiler/utilities.ts:1296:1`
- `src/compiler/utilities.ts:2318:1`
- `src/compiler/utilitiesPublic.ts:2133:1`

### guard-0239: src/compiler/factory/nodeTests.ts:898:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ExportSpecifier
~~~~

Exact reason:

~~~~text
a view of Node as ExportSpecifier in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as NamedExports) through its field parent (Node as ExportDeclaration) through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isExportSpecifier(node: Node): node is ExportSpecifier {
    return node.kind === SyntaxKind.ExportSpecifier;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1532:1`

### guard-0240: src/compiler/factory/nodeTests.ts:985:1

Observations: {'head': 4}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as HeritageClause
~~~~

Exact reason:

~~~~text
a view of Node as HeritageClause in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as InterfaceDeclaration | ClassLikeDeclaration) through an arm (Node as ClassDeclaration) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isHeritageClause(node: Node): node is HeritageClause {
    return node.kind === SyntaxKind.HeritageClause;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/declarations/diagnostics.ts:615:1`
- `src/compiler/utilities.ts:2882:1`
- `src/compiler/utilities.ts:3629:1`
- `src/compiler/utilities.ts:7453:1`

### guard-0241: src/compiler/parser.ts:10806:37

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
JsxTagNameExpression as Identifier
~~~~

Exact reason:

~~~~text
a view of JsxTagNameExpression as Identifier in a graph program, which the cycle finder doesn't relate yet through a source arm (Identifier as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
rhs as Identifier
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/parser.ts:10800:1`

### guard-0242: src/compiler/transformers/utilities.ts:632:13

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
ClassElement as PropertyDeclaration
~~~~

Exact reason:

~~~~text
a view of ClassElement as PropertyDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ClassLikeDeclaration) through an arm (Node as ClassDeclaration) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
member as PropertyDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/utilities.ts:630:1`

### guard-0243: src/compiler/utilities.ts:10814:32

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as TypeParameterDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as TypeParameterDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as InferTypeNode | DeclarationWithTypeParameterChildren) through an arm (Node as ArrowFunction) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as TypeParameterDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:10810:1`

### guard-0244: src/compiler/utilities.ts:11038:21

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ImportClause
~~~~

Exact reason:

~~~~text
a view of Node as ImportClause in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ImportDeclaration | JSDocImportTag) through an arm (Node as ImportDeclaration) through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ImportClause
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:11026:1`

### guard-0245: src/compiler/utilities.ts:11040:21

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ImportSpecifier
~~~~

Exact reason:

~~~~text
a view of Node as ImportSpecifier in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as NamedImports) through its field parent (Node as ImportClause) through its field parent (Node as ImportDeclaration | JSDocImportTag) through an arm (Node as ImportDeclaration) through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ImportSpecifier
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:11026:1`

### guard-0246: src/compiler/utilities.ts:11042:21

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ExportSpecifier
~~~~

Exact reason:

~~~~text
a view of Node as ExportSpecifier in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as NamedExports) through its field parent (Node as ExportDeclaration) through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ExportSpecifier
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:11026:1`

### guard-0247: src/compiler/utilities.ts:11459:70

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
ClassElement as ConstructorDeclaration
~~~~

Exact reason:

~~~~text
a view of ClassElement as ConstructorDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ClassLikeDeclaration) through an arm (Node as ClassDeclaration) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
member as ConstructorDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:11456:1`

### guard-0248: src/compiler/utilities.ts:2114:71

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ModuleDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as ModuleDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ModuleDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2112:1`

### guard-0249: src/compiler/utilities.ts:2274:1

Observations: {'head': 3}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ImportDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as ImportDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isAnyImportSyntax(node: Node): node is AnyImportSyntax {
    switch (node.kind) {
        case SyntaxKind.ImportDeclaration:
        case SyntaxKind.ImportEqualsDeclaration:
            return true;
        default:
            return false;
    }
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2285:1`
- `src/compiler/utilities.ts:2290:1`
- `src/compiler/utilities.ts:2318:1`

### guard-0250: src/compiler/utilities.ts:2274:1

Observations: {'head': 3}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ImportEqualsDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as ImportEqualsDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isAnyImportSyntax(node: Node): node is AnyImportSyntax {
    switch (node.kind) {
        case SyntaxKind.ImportDeclaration:
        case SyntaxKind.ImportEqualsDeclaration:
            return true;
        default:
            return false;
    }
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2285:1`
- `src/compiler/utilities.ts:2290:1`
- `src/compiler/utilities.ts:2318:1`

### guard-0251: src/compiler/utilities.ts:2318:1

Observations: {'head': 3}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ExportDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as ExportDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isAnyImportOrReExport(node: Node): node is AnyImportOrReExport {
    return isAnyImportSyntax(node) || isExportDeclaration(node);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2313:1`
- `src/compiler/utilitiesPublic.ts:2143:1`
- `src/compiler/utilitiesPublic.ts:2148:1`

### guard-0252: src/compiler/utilities.ts:2318:1

Observations: {'head': 3}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ImportDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as ImportDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isAnyImportOrReExport(node: Node): node is AnyImportOrReExport {
    return isAnyImportSyntax(node) || isExportDeclaration(node);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2313:1`
- `src/compiler/utilitiesPublic.ts:2143:1`
- `src/compiler/utilitiesPublic.ts:2148:1`

### guard-0253: src/compiler/utilities.ts:2318:1

Observations: {'head': 3}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ImportEqualsDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as ImportEqualsDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isAnyImportOrReExport(node: Node): node is AnyImportOrReExport {
    return isAnyImportSyntax(node) || isExportDeclaration(node);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:2313:1`
- `src/compiler/utilitiesPublic.ts:2143:1`
- `src/compiler/utilitiesPublic.ts:2148:1`

### guard-0254: src/compiler/utilities.ts:2994:1

Observations: {'head': 3}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as BindingElement
~~~~

Exact reason:

~~~~text
a view of Node as BindingElement in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as BindingPattern) through an arm (Node as ArrayBindingPattern) through its field parent (Node as BindingElement | ParameterDeclaration | VariableDeclaration) through an arm (Node as ParameterDeclaration) through its field parent (Node as SignatureDeclaration) through an arm (Node as ArrowFunction) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isVariableLike(node: Node): node is VariableLikeDeclaration {
    if (node) {
        switch (node.kind) {
            case SyntaxKind.BindingElement:
            case SyntaxKind.EnumMember:
            case SyntaxKind.Parameter:
            case SyntaxKind.PropertyAssignment:
            case SyntaxKind.PropertyDeclaration:
            case SyntaxKind.PropertySignature:
            case SyntaxKind.ShorthandPropertyAssignment:
            case SyntaxKind.VariableDeclaration:
                return true;
        }
    }
    return false;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:4652:1`
- `src/compiler/utilities.ts:4655:1`
- `src/compiler/utilities.ts:4656:1`

### guard-0255: src/compiler/utilities.ts:3076:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as GetAccessorDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as GetAccessorDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as InterfaceDeclaration | ObjectLiteralExpression | TypeLiteralNode | ClassLikeDeclaration) through an arm (Node as ClassDeclaration) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isObjectLiteralOrClassExpressionMethodOrAccessor(node: Node): node is MethodDeclaration | AccessorDeclaration {
    return (node.kind === SyntaxKind.MethodDeclaration || node.kind === SyntaxKind.GetAccessor || node.kind === SyntaxKind.SetAccessor) &&
        (node.parent.kind === SyntaxKind.ObjectLiteralExpression ||
            node.parent.kind === SyntaxKind.ClassExpression);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/binder.ts:3815:1`

### guard-0256: src/compiler/utilities.ts:3772:65

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ImportEqualsDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as ImportEqualsDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ImportEqualsDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:3771:1`

### guard-0257: src/compiler/utilities.ts:3788:65

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ImportEqualsDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as ImportEqualsDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ImportEqualsDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:3787:1`

### guard-0258: src/compiler/utilities.ts:4480:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDocCallbackTag
~~~~

Exact reason:

~~~~text
a view of Node as JSDocCallbackTag in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as JSDoc) through its field parent (Node as HasJSDoc) through an arm (Node as ArrowFunction) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isJSDocTypeAlias(node: Node): node is JSDocTypedefTag | JSDocCallbackTag | JSDocEnumTag {
    return node.kind === SyntaxKind.JSDocTypedefTag || node.kind === SyntaxKind.JSDocCallbackTag || node.kind === SyntaxKind.JSDocEnumTag;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:4485:1`
- `src/compiler/utilities.ts:6921:1`

### guard-0259: src/compiler/utilities.ts:4513:21

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as PropertyDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as PropertyDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ClassLikeDeclaration) through an arm (Node as ClassDeclaration) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as PropertyDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:4507:1`

### guard-0260: src/compiler/utilities.ts:5338:11

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
ArrowFunction | CallSignatureDeclaration | ConstructSignatureDeclaration | ConstructorDeclaration | ... 8 more ... | AccessorDeclaration as FunctionLikeDeclaration
~~~~

Exact reason:

~~~~text
a view of ArrowFunction | CallSignatureDeclaration | ConstructSignatureDeclaration | ConstructorDeclaration | ... 8 more ... | AccessorDeclaration as FunctionLikeDeclaration in a graph program, which the cycle finder doesn't relate yet through a source arm (ArrowFunction as FunctionLikeDeclaration) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through a source arm (AbstractKeyword as Modifier) through an arm (AbstractKeyword as AbstractKeyword) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (Block as HasJSDoc) through an arm (Block as Block) through its field nextContainer (HasLocals | undefined as HasLocals | undefined) through a source arm (CallSignatureDeclaration as HasLocals) through an arm (CallSignatureDeclaration as CallSignatureDeclaration) through its field name (PropertyName | undefined as PropertyName | undefined) through a source arm (Identifier as PropertyName) through an arm (Identifier as Identifier) through its field flowNode (FlowNode | undefined as FlowNode | undefined) through a source arm (FlowArrayMutation as FlowNode) through an arm (FlowArrayMutation as FlowArrayMutation) through its field node (BinaryExpression | CallExpression as BinaryExpression | CallExpression) through an arm (BinaryExpression | CallExpression as BinaryExpression) through a source arm (CallExpression as BinaryExpression)
~~~~

Source text:

~~~~typescript
node as FunctionLikeDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `CallExpression as BinaryExpression` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:5316:1`

### guard-0261: src/compiler/utilities.ts:5527:34

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as VariableDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as VariableDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as CatchClause | VariableDeclarationList) through an arm (Node as CatchClause) through its field parent (Node as TryStatement) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as VariableDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:5520:1`

### guard-0262: src/compiler/utilitiesPublic.ts:1449:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as JSDocPropertyLikeTag
~~~~

Exact reason:

~~~~text
a view of Node as JSDocPropertyLikeTag in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as JSDoc) through its field parent (Node as HasJSDoc) through an arm (Node as ArrowFunction) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isJSDocPropertyLikeTag(node: Node): node is JSDocPropertyLikeTag {
    return node.kind === SyntaxKind.JSDocPropertyTag || node.kind === SyntaxKind.JSDocParameterTag;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilities.ts:11055:1`

### guard-0263: src/compiler/utilitiesPublic.ts:1539:21

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ImportSpecifier
~~~~

Exact reason:

~~~~text
a view of Node as ImportSpecifier in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as NamedImports) through its field parent (Node as ImportClause) through its field parent (Node as ImportDeclaration | JSDocImportTag) through an arm (Node as ImportDeclaration) through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ImportSpecifier
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1536:1`

### guard-0264: src/compiler/utilitiesPublic.ts:1539:61

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ImportSpecifier
~~~~

Exact reason:

~~~~text
a view of Node as ImportSpecifier in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as NamedImports) through its field parent (Node as ImportClause) through its field parent (Node as ImportDeclaration | JSDocImportTag) through an arm (Node as ImportDeclaration) through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ImportSpecifier
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1536:1`

### guard-0265: src/compiler/utilitiesPublic.ts:1541:21

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as NamespaceImport
~~~~

Exact reason:

~~~~text
a view of Node as NamespaceImport in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ImportClause) through its field parent (Node as ImportDeclaration | JSDocImportTag) through an arm (Node as ImportDeclaration) through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as NamespaceImport
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1536:1`

### guard-0266: src/compiler/utilitiesPublic.ts:1543:21

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ImportClause
~~~~

Exact reason:

~~~~text
a view of Node as ImportClause in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ImportDeclaration | JSDocImportTag) through an arm (Node as ImportDeclaration) through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ImportClause
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1536:1`

### guard-0267: src/compiler/utilitiesPublic.ts:1545:21

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ImportEqualsDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as ImportEqualsDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ImportEqualsDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1536:1`

### guard-0268: src/compiler/utilitiesPublic.ts:1553:21

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ExportSpecifier
~~~~

Exact reason:

~~~~text
a view of Node as ExportSpecifier in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as NamedExports) through its field parent (Node as ExportDeclaration) through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ExportSpecifier
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1550:1`

### guard-0269: src/compiler/utilitiesPublic.ts:1553:61

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ExportSpecifier
~~~~

Exact reason:

~~~~text
a view of Node as ExportSpecifier in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as NamedExports) through its field parent (Node as ExportDeclaration) through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ExportSpecifier
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1550:1`

### guard-0270: src/compiler/utilitiesPublic.ts:1555:113

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ExportDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as ExportDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ExportDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1550:1`

### guard-0271: src/compiler/utilitiesPublic.ts:1555:21

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ExportDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as ExportDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ExportDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1550:1`

### guard-0272: src/compiler/utilitiesPublic.ts:1555:65

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as ExportDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as ExportDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ExportDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1550:1`

### guard-0273: src/compiler/utilitiesPublic.ts:1557:21

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as NamespaceExport
~~~~

Exact reason:

~~~~text
a view of Node as NamespaceExport in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ExportDeclaration) through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as NamespaceExport
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1550:1`

### guard-0274: src/compiler/utilitiesPublic.ts:1581:1

Observations: {'head': 4}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as GeneratedIdentifier
~~~~

Exact reason:

~~~~text
a view of Node as GeneratedIdentifier in a graph program, which the cycle finder doesn't relate yet through its field emitNode (EmitNode | undefined as EmitNode & { autoGenerate: AutoGenerateInfo; }) through an arm (EmitNode as EmitNode) through its field annotatedNodes (Node[] | undefined as Node[] | undefined) through its elements (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
// Identifiers

/** @internal */
export function isGeneratedIdentifier(node: Node): node is GeneratedIdentifier {
    return isIdentifier(node) && node.emitNode?.autoGenerate !== undefined;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/factory/utilities.ts:1570:1`
- `src/compiler/factory/utilities.ts:786:1`
- `src/compiler/transformers/utilities.ts:389:1`
- `src/compiler/transformers/utilities.ts:441:1`

### guard-0275: src/compiler/utilitiesPublic.ts:1586:1

Observations: {'head': 4}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as GeneratedPrivateIdentifier
~~~~

Exact reason:

~~~~text
a view of Node as GeneratedPrivateIdentifier in a graph program, which the cycle finder doesn't relate yet through its field emitNode (EmitNode | undefined as EmitNode & { autoGenerate: AutoGenerateInfo; }) through an arm (EmitNode as EmitNode) through its field annotatedNodes (Node[] | undefined as Node[] | undefined) through its elements (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isGeneratedPrivateIdentifier(node: Node): node is GeneratedPrivateIdentifier {
    return isPrivateIdentifier(node) && node.emitNode?.autoGenerate !== undefined;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/factory/utilities.ts:1570:1`
- `src/compiler/transformers/classFields.ts:3343:1`
- `src/compiler/transformers/utilities.ts:389:1`
- `src/compiler/transformers/utilities.ts:441:1`

### guard-0276: src/compiler/utilitiesPublic.ts:1748:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as GetAccessorDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as GetAccessorDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as InterfaceDeclaration | ObjectLiteralExpression | TypeLiteralNode | ClassLikeDeclaration) through an arm (Node as ClassDeclaration) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isAccessor(node: Node): node is AccessorDeclaration {
    return node && (node.kind === SyntaxKind.GetAccessor || node.kind === SyntaxKind.SetAccessor);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/checker.ts:54222:1`
- `src/compiler/utilities.ts:3574:1`

### guard-0277: src/compiler/utilitiesPublic.ts:1765:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as GetAccessorDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as GetAccessorDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as InterfaceDeclaration | ObjectLiteralExpression | TypeLiteralNode | ClassLikeDeclaration) through an arm (Node as ClassDeclaration) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
/** @internal */
export function isMethodOrAccessor(node: Node): node is MethodDeclaration | AccessorDeclaration {
    switch (node.kind) {
        case SyntaxKind.MethodDeclaration:
        case SyntaxKind.GetAccessor:
        case SyntaxKind.SetAccessor:
            return true;
        default:
            return false;
    }
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/utilities.ts:642:1`
- `src/compiler/utilitiesPublic.ts:1600:1`

### guard-0278: src/compiler/utilitiesPublic.ts:1798:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as GetAccessorDeclaration
~~~~

Exact reason:

~~~~text
a view of Node as GetAccessorDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as InterfaceDeclaration | ObjectLiteralExpression | TypeLiteralNode | ClassLikeDeclaration) through an arm (Node as ClassDeclaration) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isObjectLiteralElementLike(node: Node): node is ObjectLiteralElementLike {
    const kind = node.kind;
    return kind === SyntaxKind.PropertyAssignment
        || kind === SyntaxKind.ShorthandPropertyAssignment
        || kind === SyntaxKind.SpreadAssignment
        || kind === SyntaxKind.MethodDeclaration
        || kind === SyntaxKind.GetAccessor
        || kind === SyntaxKind.SetAccessor;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:2592:1`

### guard-0279: src/compiler/utilitiesPublic.ts:1860:1

Observations: {'head': 3}. New in head: True.

Source and target as named by the guard:

~~~~text
BindingOrAssignmentElement as BindingElement
~~~~

Exact reason:

~~~~text
a view of BindingOrAssignmentElement as BindingElement in a graph program, which the cycle finder doesn't relate yet through a source arm (ArrayLiteralExpression as BindingElement)
~~~~

Source text:

~~~~typescript
/**
 * Determines whether the BindingOrAssignmentElement is a BindingElement-like declaration
 *
 * @internal
 */
export function isDeclarationBindingElement(bindingElement: BindingOrAssignmentElement): bindingElement is VariableDeclaration | ParameterDeclaration | BindingElement {
    switch (bindingElement.kind) {
        case SyntaxKind.VariableDeclaration:
        case SyntaxKind.Parameter:
        case SyntaxKind.BindingElement:
            return true;
    }

    return false;
}
~~~~

Reading: **guard bug: checked pattern alternatives.** Guard bug. isDeclarationBindingElement returns true only for VariableDeclaration, Parameter, or BindingElement kinds. getInitializerOfBindingOrAssignmentElement and getTargetOfBindingOrAssignmentElement call it before reading the corresponding fields. An ArrayLiteralExpression is rejected by that test, so it does not enter this BindingElement view. The leaf pair is disjoint, but this is not an incompatible write through an admitted view.

Owner units:

- `src/compiler/factory/utilities.ts:865:1`
- `src/compiler/factory/utilities.ts:910:1`
- `src/compiler/transformers/destructuring.ts:386:1`

### guard-0280: src/compiler/utilitiesPublic.ts:1884:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
BindingOrAssignmentElementTarget as ArrayBindingPattern
~~~~

Exact reason:

~~~~text
a view of BindingOrAssignmentElementTarget as ArrayBindingPattern in a graph program, which the cycle finder doesn't relate yet through a source arm (ArrayBindingPattern as ArrayBindingPattern) through its field parent (BindingElement | ParameterDeclaration | VariableDeclaration as BindingElement | ParameterDeclaration | VariableDeclaration) through a source arm (BindingElement as BindingElement | ParameterDeclaration | VariableDeclaration) through an arm (BindingElement as BindingElement) through its field parent (BindingPattern as BindingPattern) through a source arm (ArrayBindingPattern as BindingPattern) through its field elements (NodeArray<ArrayBindingElement> as NodeArray<BindingElement> | NodeArray<ArrayBindingElement>) through an arm (NodeArray<ArrayBindingElement> as NodeArray<BindingElement>) through its index values (ArrayBindingElement as BindingElement) through a source arm (OmittedExpression as BindingElement)
~~~~

Source text:

~~~~typescript
/**
 * Determines whether a node is a BindingOrAssignmentPattern
 *
 * @internal
 */
export function isBindingOrAssignmentPattern(node: BindingOrAssignmentElementTarget): node is BindingOrAssignmentPattern {
    return isObjectBindingOrAssignmentPattern(node)
        || isArrayBindingOrAssignmentPattern(node);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `BindingElement | ParameterDeclaration | VariableDeclaration` to the same `BindingElement | ParameterDeclaration | VariableDeclaration` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `OmittedExpression as BindingElement` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/destructuring.ts:197:1`
- `src/compiler/transformers/destructuring.ts:218:1`

### guard-0281: src/compiler/utilitiesPublic.ts:1884:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
BindingOrAssignmentElementTarget as ArrayLiteralExpression
~~~~

Exact reason:

~~~~text
a view of BindingOrAssignmentElementTarget as ArrayLiteralExpression in a graph program, which the cycle finder doesn't relate yet through a source arm (ArrayBindingPattern as ArrayLiteralExpression)
~~~~

Source text:

~~~~typescript
/**
 * Determines whether a node is a BindingOrAssignmentPattern
 *
 * @internal
 */
export function isBindingOrAssignmentPattern(node: BindingOrAssignmentElementTarget): node is BindingOrAssignmentPattern {
    return isObjectBindingOrAssignmentPattern(node)
        || isArrayBindingOrAssignmentPattern(node);
}
~~~~

Reading: **guard bug: checked pattern alternatives.** Guard bug. isBindingOrAssignmentPattern delegates to the object-pattern and array-pattern kind switches. bindingOrAssignmentElementAssignsToName uses the checked result as a discriminated pattern union. An ArrayBindingPattern remains the array-binding alternative; it is not reinterpreted as this ArrayLiteralExpression alternative. The readonly kind and readonly elements declaration preserve that distinction. The guard has crossed source and target alternatives; the disjoint leaf pair is not a write the admitted alternative introduces.

Owner units:

- `src/compiler/transformers/destructuring.ts:197:1`
- `src/compiler/transformers/destructuring.ts:218:1`

### guard-0282: src/compiler/utilitiesPublic.ts:1884:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
BindingOrAssignmentElementTarget as ObjectBindingPattern
~~~~

Exact reason:

~~~~text
a view of BindingOrAssignmentElementTarget as ObjectBindingPattern in a graph program, which the cycle finder doesn't relate yet through a source arm (ArrayBindingPattern as ObjectBindingPattern)
~~~~

Source text:

~~~~typescript
/**
 * Determines whether a node is a BindingOrAssignmentPattern
 *
 * @internal
 */
export function isBindingOrAssignmentPattern(node: BindingOrAssignmentElementTarget): node is BindingOrAssignmentPattern {
    return isObjectBindingOrAssignmentPattern(node)
        || isArrayBindingOrAssignmentPattern(node);
}
~~~~

Reading: **guard bug: checked pattern alternatives.** Guard bug. isBindingOrAssignmentPattern delegates to the object-pattern and array-pattern kind switches. bindingOrAssignmentElementAssignsToName uses the checked result as a discriminated pattern union. An ArrayBindingPattern remains the array-binding alternative; it is not reinterpreted as this ObjectBindingPattern alternative. The readonly kind and readonly elements declaration preserve that distinction. The guard has crossed source and target alternatives; the disjoint leaf pair is not a write the admitted alternative introduces.

Owner units:

- `src/compiler/transformers/destructuring.ts:197:1`
- `src/compiler/transformers/destructuring.ts:218:1`

### guard-0283: src/compiler/utilitiesPublic.ts:1884:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
BindingOrAssignmentElementTarget as ObjectLiteralExpression
~~~~

Exact reason:

~~~~text
a view of BindingOrAssignmentElementTarget as ObjectLiteralExpression in a graph program, which the cycle finder doesn't relate yet through a source arm (ArrayBindingPattern as ObjectLiteralExpression)
~~~~

Source text:

~~~~typescript
/**
 * Determines whether a node is a BindingOrAssignmentPattern
 *
 * @internal
 */
export function isBindingOrAssignmentPattern(node: BindingOrAssignmentElementTarget): node is BindingOrAssignmentPattern {
    return isObjectBindingOrAssignmentPattern(node)
        || isArrayBindingOrAssignmentPattern(node);
}
~~~~

Reading: **guard bug: checked pattern alternatives.** Guard bug. isBindingOrAssignmentPattern delegates to the object-pattern and array-pattern kind switches. bindingOrAssignmentElementAssignsToName uses the checked result as a discriminated pattern union. An ArrayBindingPattern remains the array-binding alternative; it is not reinterpreted as this ObjectLiteralExpression alternative. The readonly kind and readonly elements declaration preserve that distinction. The guard has crossed source and target alternatives; the disjoint leaf pair is not a write the admitted alternative introduces.

Owner units:

- `src/compiler/transformers/destructuring.ts:197:1`
- `src/compiler/transformers/destructuring.ts:218:1`

### guard-0284: src/compiler/utilitiesPublic.ts:1894:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
BindingOrAssignmentElementTarget as ObjectBindingPattern
~~~~

Exact reason:

~~~~text
a view of BindingOrAssignmentElementTarget as ObjectBindingPattern in a graph program, which the cycle finder doesn't relate yet through a source arm (ArrayBindingPattern as ObjectBindingPattern)
~~~~

Source text:

~~~~typescript
/**
 * Determines whether a node is an ObjectBindingOrAssignmentPattern
 *
 * @internal
 */
export function isObjectBindingOrAssignmentPattern(node: BindingOrAssignmentElementTarget): node is ObjectBindingOrAssignmentPattern {
    switch (node.kind) {
        case SyntaxKind.ObjectBindingPattern:
        case SyntaxKind.ObjectLiteralExpression:
            return true;
    }

    return false;
}
~~~~

Reading: **guard bug: excluded discriminator arm.** This predicate checks node kinds and preserves their discriminated alternatives. The concrete recorded target is `ObjectBindingPattern`; the rejected `ArrayBindingPattern` has a different readonly kind. It cannot enter this particular target alternative. If the predicate also admits that source kind, it belongs to its own target alternative instead. The leaf types are disjoint, but the guard incorrectly requires every source alternative to fit this individual target, losing that correlation. No write through this target alternative is a store into the other kind.

Owner units:

- `src/compiler/utilitiesPublic.ts:1884:1`

### guard-0285: src/compiler/utilitiesPublic.ts:1894:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
BindingOrAssignmentElementTarget as ObjectLiteralExpression
~~~~

Exact reason:

~~~~text
a view of BindingOrAssignmentElementTarget as ObjectLiteralExpression in a graph program, which the cycle finder doesn't relate yet through a source arm (ArrayBindingPattern as ObjectLiteralExpression)
~~~~

Source text:

~~~~typescript
/**
 * Determines whether a node is an ObjectBindingOrAssignmentPattern
 *
 * @internal
 */
export function isObjectBindingOrAssignmentPattern(node: BindingOrAssignmentElementTarget): node is ObjectBindingOrAssignmentPattern {
    switch (node.kind) {
        case SyntaxKind.ObjectBindingPattern:
        case SyntaxKind.ObjectLiteralExpression:
            return true;
    }

    return false;
}
~~~~

Reading: **guard bug: excluded discriminator arm.** This predicate checks node kinds and preserves their discriminated alternatives. The concrete recorded target is `ObjectLiteralExpression`; the rejected `ArrayBindingPattern` has a different readonly kind. It cannot enter this particular target alternative. If the predicate also admits that source kind, it belongs to its own target alternative instead. The leaf types are disjoint, but the guard incorrectly requires every source alternative to fit this individual target, losing that correlation. No write through this target alternative is a store into the other kind.

Owner units:

- `src/compiler/utilitiesPublic.ts:1884:1`

### guard-0286: src/compiler/utilitiesPublic.ts:1921:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
BindingOrAssignmentElementTarget as ArrayBindingPattern
~~~~

Exact reason:

~~~~text
a view of BindingOrAssignmentElementTarget as ArrayBindingPattern in a graph program, which the cycle finder doesn't relate yet through a source arm (ArrayBindingPattern as ArrayBindingPattern) through its field parent (BindingElement | ParameterDeclaration | VariableDeclaration as BindingElement | ParameterDeclaration | VariableDeclaration) through a source arm (BindingElement as BindingElement | ParameterDeclaration | VariableDeclaration) through an arm (BindingElement as BindingElement) through its field parent (BindingPattern as BindingPattern) through a source arm (ArrayBindingPattern as BindingPattern) through its field elements (NodeArray<ArrayBindingElement> as NodeArray<BindingElement> | NodeArray<ArrayBindingElement>) through an arm (NodeArray<ArrayBindingElement> as NodeArray<BindingElement>) through its index values (ArrayBindingElement as BindingElement) through a source arm (OmittedExpression as BindingElement)
~~~~

Source text:

~~~~typescript
/**
 * Determines whether a node is an ArrayBindingOrAssignmentPattern
 *
 * @internal
 */
export function isArrayBindingOrAssignmentPattern(node: BindingOrAssignmentElementTarget): node is ArrayBindingOrAssignmentPattern {
    switch (node.kind) {
        case SyntaxKind.ArrayBindingPattern:
        case SyntaxKind.ArrayLiteralExpression:
            return true;
    }

    return false;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `BindingElement | ParameterDeclaration | VariableDeclaration` to the same `BindingElement | ParameterDeclaration | VariableDeclaration` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `OmittedExpression as BindingElement` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/utilitiesPublic.ts:1884:1`

### guard-0287: src/compiler/utilitiesPublic.ts:1921:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
BindingOrAssignmentElementTarget as ArrayLiteralExpression
~~~~

Exact reason:

~~~~text
a view of BindingOrAssignmentElementTarget as ArrayLiteralExpression in a graph program, which the cycle finder doesn't relate yet through a source arm (ArrayBindingPattern as ArrayLiteralExpression)
~~~~

Source text:

~~~~typescript
/**
 * Determines whether a node is an ArrayBindingOrAssignmentPattern
 *
 * @internal
 */
export function isArrayBindingOrAssignmentPattern(node: BindingOrAssignmentElementTarget): node is ArrayBindingOrAssignmentPattern {
    switch (node.kind) {
        case SyntaxKind.ArrayBindingPattern:
        case SyntaxKind.ArrayLiteralExpression:
            return true;
    }

    return false;
}
~~~~

Reading: **guard bug: excluded discriminator arm.** This predicate checks node kinds and preserves their discriminated alternatives. The concrete recorded target is `ArrayLiteralExpression`; the rejected `ArrayBindingPattern` has a different readonly kind. It cannot enter this particular target alternative. If the predicate also admits that source kind, it belongs to its own target alternative instead. The leaf types are disjoint, but the guard incorrectly requires every source alternative to fit this individual target, losing that correlation. No write through this target alternative is a store into the other kind.

Owner units:

- `src/compiler/utilitiesPublic.ts:1884:1`

### guard-0288: src/compiler/utilitiesPublic.ts:2625:1

Observations: {'head': 3}. New in head: True.

Source and target as named by the guard:

~~~~text
FileReference | Node as NoSubstitutionTemplateLiteral
~~~~

Exact reason:

~~~~text
a view of FileReference | Node as NoSubstitutionTemplateLiteral in a graph program, which the cycle finder doesn't relate yet through a source arm (FileReference as NoSubstitutionTemplateLiteral)
~~~~

Source text:

~~~~typescript
export function isStringLiteralLike(node: Node | FileReference): node is StringLiteralLike {
    return (node as Node).kind === SyntaxKind.StringLiteral || (node as Node).kind === SyntaxKind.NoSubstitutionTemplateLiteral;
}
~~~~

Reading: **guard bug: excluded discriminator arm.** This predicate checks node kinds and preserves their discriminated alternatives. The concrete recorded target is `NoSubstitutionTemplateLiteral`; the rejected `FileReference` has a different readonly kind. It cannot enter this particular target alternative. If the predicate also admits that source kind, it belongs to its own target alternative instead. The leaf types are disjoint, but the guard incorrectly requires every source alternative to fit this individual target, losing that correlation. No write through this target alternative is a store into the other kind.

Owner units:

- `src/compiler/transformers/utilities.ts:471:1`
- `src/compiler/utilities.ts:5360:1`
- `src/compiler/utilities.ts:6507:1`

### guard-0289: src/compiler/utilitiesPublic.ts:2625:1

Observations: {'head': 3}. New in head: True.

Source and target as named by the guard:

~~~~text
FileReference | Node as StringLiteral
~~~~

Exact reason:

~~~~text
a view of FileReference | Node as StringLiteral in a graph program, which the cycle finder doesn't relate yet through a source arm (FileReference as StringLiteral)
~~~~

Source text:

~~~~typescript
export function isStringLiteralLike(node: Node | FileReference): node is StringLiteralLike {
    return (node as Node).kind === SyntaxKind.StringLiteral || (node as Node).kind === SyntaxKind.NoSubstitutionTemplateLiteral;
}
~~~~

Reading: **guard bug: excluded discriminator arm.** This predicate checks node kinds and preserves their discriminated alternatives. The concrete recorded target is `StringLiteral`; the rejected `FileReference` has a different readonly kind. It cannot enter this particular target alternative. If the predicate also admits that source kind, it belongs to its own target alternative instead. The leaf types are disjoint, but the guard incorrectly requires every source alternative to fit this individual target, losing that correlation. No write through this target alternative is a store into the other kind.

Owner units:

- `src/compiler/transformers/utilities.ts:471:1`
- `src/compiler/utilities.ts:5360:1`
- `src/compiler/utilities.ts:6507:1`

## base pair check

None.

