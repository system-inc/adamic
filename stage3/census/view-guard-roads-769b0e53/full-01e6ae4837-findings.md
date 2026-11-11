# full-01e6ae4837: every guard finding in full mode

Reasons and type names are verbatim. Source slices retain their text, with CRLF displayed as LF. Repeated observations are grouped only when site, source, target, reason and source text agree; every owner unit is listed. Road is the first new mechanism in the actual path; the exact reason preserves every subsequent mechanism.

## elements

None.

## index values

<a id="full-01e6ae4837-0001"></a>

### full-01e6ae4837-0001: src/compiler/debug.ts:281:5

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node | undefined as Statement
~~~~

Exact reason:

~~~~text
a view of Node | undefined as Statement in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function assertNode<T extends Node, U extends T>(node: T | undefined, test: (node: T) => node is U, message?: string, stackCrawlMark?: AnyFunction): asserts node is U;
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/esnext.ts:339:5`
- `src/compiler/transformers/esnext.ts:395:9`

<a id="full-01e6ae4837-0002"></a>

### full-01e6ae4837-0002: src/compiler/transformers/esnext.ts:410:50

Observations: {'head': 4}. New in head: True.

Source and target as named by the guard:

~~~~text
Statement as ExportAssignment
~~~~

Exact reason:

~~~~text
a view of Statement as ExportAssignment in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as SourceFile) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ExportAssignment
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/esnext.ts:339:5`
- `src/compiler/transformers/esnext.ts:400:9`

<a id="full-01e6ae4837-0003"></a>

### full-01e6ae4837-0003: src/compiler/transformers/esnext.ts:412:50

Observations: {'head': 4}. New in head: True.

Source and target as named by the guard:

~~~~text
Statement as ClassDeclaration
~~~~

Exact reason:

~~~~text
a view of Statement as ClassDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ClassDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/esnext.ts:339:5`
- `src/compiler/transformers/esnext.ts:400:9`

<a id="full-01e6ae4837-0004"></a>

### full-01e6ae4837-0004: src/compiler/utilities.ts:2727:1

Observations: {'head': 1}. New in head: True.

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

- `src/compiler/transformers/esnext.ts:765:1`

<a id="full-01e6ae4837-0005"></a>

### full-01e6ae4837-0005: src/compiler/utilitiesPublic.ts:2419:1

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as Statement
~~~~

Exact reason:

~~~~text
a view of Node as Statement in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isStatement(node: Node): node is Statement {
    const kind = node.kind;
    return isStatementKindButNotDeclarationKind(kind)
        || isDeclarationStatementKind(kind)
        || isBlockStatement(node);
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/esnext.ts:339:5`
- `src/compiler/transformers/esnext.ts:395:9`

## union arms

<a id="full-01e6ae4837-0006"></a>

### full-01e6ae4837-0006: src/compiler/factory/nodeTests.ts:796:1

Observations: {'head': 1}. New in head: True.

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

<a id="full-01e6ae4837-0007"></a>

### full-01e6ae4837-0007: src/compiler/utilitiesPublic.ts:1581:1

Observations: {'head': 1}. New in head: True.

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

- `src/compiler/transformers/esnext.ts:636:5`

## base pair check

None.

