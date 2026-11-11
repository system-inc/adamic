# full-91560bd618: every guard finding in full mode

Reasons and type names are verbatim. Source slices retain their text, with CRLF displayed as LF. Repeated observations are grouped only when site, source, target, reason and source text agree; every owner unit is listed. Road is the first new mechanism in the actual path; the exact reason preserves every subsequent mechanism.

## elements

None.

## index values

<a id="full-91560bd618-0001"></a>

### full-91560bd618-0001: src/compiler/factory/nodeTests.ts:544:1

Observations: {'head': 2}. New in head: True.

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

- `src/compiler/transformers/typeSerializer.ts:132:1`
- `src/compiler/transformers/typeSerializer.ts:402:5`

<a id="full-91560bd618-0002"></a>

### full-91560bd618-0002: src/compiler/transformers/typeSerializer.ts:308:59

Observations: {'head': 4}. New in head: True.

Source and target as named by the guard:

~~~~text
TypeNode as LiteralTypeNode
~~~~

Exact reason:

~~~~text
a view of TypeNode as LiteralTypeNode in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as LiteralTypeNode
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/typeSerializer.ts:132:1`
- `src/compiler/transformers/typeSerializer.ts:271:5`

<a id="full-91560bd618-0003"></a>

### full-91560bd618-0003: src/compiler/transformers/typeSerializer.ts:320:51

Observations: {'head': 4}. New in head: True.

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

- `src/compiler/transformers/typeSerializer.ts:132:1`
- `src/compiler/transformers/typeSerializer.ts:271:5`

<a id="full-91560bd618-0004"></a>

### full-91560bd618-0004: src/compiler/transformers/typeSerializer.ts:332:22

Observations: {'head': 4}. New in head: True.

Source and target as named by the guard:

~~~~text
TypeNode as TypeOperatorNode
~~~~

Exact reason:

~~~~text
a view of TypeNode as TypeOperatorNode in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as TypeOperatorNode
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/typeSerializer.ts:132:1`
- `src/compiler/transformers/typeSerializer.ts:271:5`

<a id="full-91560bd618-0005"></a>

### full-91560bd618-0005: src/compiler/transformers/typeSerializer.ts:333:47

Observations: {'head': 4}. New in head: True.

Source and target as named by the guard:

~~~~text
TypeNode as TypeOperatorNode
~~~~

Exact reason:

~~~~text
a view of TypeNode as TypeOperatorNode in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as TypeOperatorNode
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/typeSerializer.ts:132:1`
- `src/compiler/transformers/typeSerializer.ts:271:5`

## union arms

<a id="full-91560bd618-0006"></a>

### full-91560bd618-0006: src/compiler/transformers/typeSerializer.ts:293:25

Observations: {'head': 4}. New in head: True.

Source and target as named by the guard:

~~~~text
TypeNode as TypePredicateNode
~~~~

Exact reason:

~~~~text
a view of TypeNode as TypePredicateNode in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as JSDocTypeExpression | SignatureDeclaration) through an arm (Node as ArrowFunction) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as TypePredicateNode
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/typeSerializer.ts:132:1`
- `src/compiler/transformers/typeSerializer.ts:271:5`

<a id="full-91560bd618-0007"></a>

### full-91560bd618-0007: src/compiler/transformers/typeSerializer.ts:615:16

Observations: {'head': 4}. New in head: True.

Source and target as named by the guard:

~~~~text
PropertyAccessExpression as PropertyAccessEntityNameExpression
~~~~

Exact reason:

~~~~text
a view of PropertyAccessExpression as PropertyAccessEntityNameExpression in a graph program, which the cycle finder doesn't relate yet through its field expression (LeftHandSideExpression as EntityNameExpression) through an arm (LeftHandSideExpression as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
factory.createPropertyAccessExpression(serializeEntityNameAsExpression(node.left), node.right) as PropertyAccessEntityNameExpression
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/typeSerializer.ts:132:1`
- `src/compiler/transformers/typeSerializer.ts:614:5`

## base pair check

None.

