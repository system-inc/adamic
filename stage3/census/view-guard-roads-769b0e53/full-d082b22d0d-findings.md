# full-d082b22d0d: every guard finding in full mode

Reasons and type names are verbatim. Source slices retain their text, with CRLF displayed as LF. Repeated observations are grouped only when site, source, target, reason and source text agree; every owner unit is listed. Road is the first new mechanism in the actual path; the exact reason preserves every subsequent mechanism.

## elements

None.

## index values

<a id="full-d082b22d0d-0001"></a>

### full-d082b22d0d-0001: src/compiler/transformers/module/esnextAnd2015.ts:151:46

Observations: {'head': 4}. New in head: True.

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
node as ExportAssignment
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/module/esnextAnd2015.ts:143:5`
- `src/compiler/transformers/module/esnextAnd2015.ts:67:1`

<a id="full-d082b22d0d-0002"></a>

### full-d082b22d0d-0002: src/compiler/utilities.ts:2704:1

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

- `src/compiler/transformers/module/esnextAnd2015.ts:67:1`
- `src/compiler/transformers/module/esnextAnd2015.ts:90:5`

<a id="full-d082b22d0d-0003"></a>

### full-d082b22d0d-0003: src/compiler/utilities.ts:3839:1

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

- `src/compiler/transformers/module/esnextAnd2015.ts:67:1`
- `src/compiler/transformers/module/esnextAnd2015.ts:90:5`

## union arms

<a id="full-d082b22d0d-0004"></a>

### full-d082b22d0d-0004: src/compiler/transformers/module/esnextAnd2015.ts:149:111

Observations: {'head': 4}. New in head: True.

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

- `src/compiler/transformers/module/esnextAnd2015.ts:143:5`
- `src/compiler/transformers/module/esnextAnd2015.ts:67:1`

<a id="full-d082b22d0d-0005"></a>

### full-d082b22d0d-0005: src/compiler/transformers/module/esnextAnd2015.ts:153:36

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
node as ExportDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/module/esnextAnd2015.ts:143:5`
- `src/compiler/transformers/module/esnextAnd2015.ts:67:1`

<a id="full-d082b22d0d-0006"></a>

### full-d082b22d0d-0006: src/compiler/transformers/module/esnextAnd2015.ts:156:47

Observations: {'head': 4}. New in head: True.

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
node as ImportDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/module/esnextAnd2015.ts:143:5`
- `src/compiler/transformers/module/esnextAnd2015.ts:67:1`

<a id="full-d082b22d0d-0007"></a>

### full-d082b22d0d-0007: src/compiler/utilitiesPublic.ts:2625:1

Observations: {'head': 6}. New in head: True.

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

- `src/compiler/transformers/module/esnextAnd2015.ts:188:5`
- `src/compiler/transformers/module/esnextAnd2015.ts:67:1`
- `src/compiler/transformers/module/esnextAnd2015.ts:90:5`

<a id="full-d082b22d0d-0008"></a>

### full-d082b22d0d-0008: src/compiler/utilitiesPublic.ts:2625:1

Observations: {'head': 6}. New in head: True.

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

- `src/compiler/transformers/module/esnextAnd2015.ts:188:5`
- `src/compiler/transformers/module/esnextAnd2015.ts:67:1`
- `src/compiler/transformers/module/esnextAnd2015.ts:90:5`

## base pair check

None.

