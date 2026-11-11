# full-e082ecd054: every guard finding in full mode

Reasons and type names are verbatim. Source slices retain their text, with CRLF displayed as LF. Repeated observations are grouped only when site, source, target, reason and source text agree; every owner unit is listed. Road is the first new mechanism in the actual path; the exact reason preserves every subsequent mechanism.

## elements

None.

## index values

<a id="full-e082ecd054-0001"></a>

### full-e082ecd054-0001: src/compiler/factory/nodeTests.ts:243:1

Observations: {'head': 3}. New in head: True.

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

- `src/compiler/transformers/utilities.ts:876:1`
- `src/compiler/transformers/utilities.ts:878:1`
- `src/compiler/transformers/utilities.ts:880:1`

<a id="full-e082ecd054-0002"></a>

### full-e082ecd054-0002: src/compiler/transformers/utilities.ts:253:22

Observations: {'head': 2}. New in head: True.

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

- `src/compiler/transformers/utilities.ts:179:1`

<a id="full-e082ecd054-0003"></a>

### full-e082ecd054-0003: src/compiler/transformers/utilities.ts:255:36

Observations: {'head': 2}. New in head: True.

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

- `src/compiler/transformers/utilities.ts:179:1`

<a id="full-e082ecd054-0004"></a>

### full-e082ecd054-0004: src/compiler/transformers/utilities.ts:278:130

Observations: {'head': 2}. New in head: True.

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

- `src/compiler/transformers/utilities.ts:179:1`

<a id="full-e082ecd054-0005"></a>

### full-e082ecd054-0005: src/compiler/transformers/utilities.ts:284:39

Observations: {'head': 2}. New in head: True.

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

- `src/compiler/transformers/utilities.ts:179:1`

## union arms

<a id="full-e082ecd054-0006"></a>

### full-e082ecd054-0006: src/compiler/core.ts:1787:32

Observations: {'head': 2}. New in head: True.

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
- `src/compiler/transformers/utilities.ts:303:5`

<a id="full-e082ecd054-0007"></a>

### full-e082ecd054-0007: src/compiler/factory/nodeTests.ts:422:1

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
export function isPropertyDeclaration(node: Node): node is PropertyDeclaration {
    return node.kind === SyntaxKind.PropertyDeclaration;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/utilities.ts:612:1`
- `src/compiler/transformers/utilities.ts:618:1`

<a id="full-e082ecd054-0008"></a>

### full-e082ecd054-0008: src/compiler/factory/nodeTests.ts:894:1

Observations: {'head': 3}. New in head: True.

Source and target as named by the guard:

~~~~text
Node as NamedExports
~~~~

Exact reason:

~~~~text
a view of Node as NamedExports in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ExportDeclaration) through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
export function isNamedExports(node: Node): node is NamedExports {
    return node.kind === SyntaxKind.NamedExports;
}
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/utilities.ts:179:1`
- `src/compiler/transformers/utilities.ts:303:5`

<a id="full-e082ecd054-0009"></a>

### full-e082ecd054-0009: src/compiler/transformers/utilities.ts:201:38

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Statement as ImportDeclaration
~~~~

Exact reason:

~~~~text
a view of Statement as ImportDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ImportDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/utilities.ts:179:1`

<a id="full-e082ecd054-0010"></a>

### full-e082ecd054-0010: src/compiler/transformers/utilities.ts:202:70

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Statement as ImportDeclaration
~~~~

Exact reason:

~~~~text
a view of Statement as ImportDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ImportDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/utilities.ts:179:1`

<a id="full-e082ecd054-0011"></a>

### full-e082ecd054-0011: src/compiler/transformers/utilities.ts:205:76

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Statement as ImportDeclaration
~~~~

Exact reason:

~~~~text
a view of Statement as ImportDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ImportDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/utilities.ts:179:1`

<a id="full-e082ecd054-0012"></a>

### full-e082ecd054-0012: src/compiler/transformers/utilities.ts:211:22

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Statement as ImportEqualsDeclaration
~~~~

Exact reason:

~~~~text
a view of Statement as ImportEqualsDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ImportEqualsDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/utilities.ts:179:1`

<a id="full-e082ecd054-0013"></a>

### full-e082ecd054-0013: src/compiler/transformers/utilities.ts:213:42

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Statement as ImportEqualsDeclaration
~~~~

Exact reason:

~~~~text
a view of Statement as ImportEqualsDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ImportEqualsDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/utilities.ts:179:1`

<a id="full-e082ecd054-0014"></a>

### full-e082ecd054-0014: src/compiler/transformers/utilities.ts:219:22

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Statement as ExportDeclaration
~~~~

Exact reason:

~~~~text
a view of Statement as ExportDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ExportDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/utilities.ts:179:1`

<a id="full-e082ecd054-0015"></a>

### full-e082ecd054-0015: src/compiler/transformers/utilities.ts:220:27

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Statement as ExportDeclaration
~~~~

Exact reason:

~~~~text
a view of Statement as ExportDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ExportDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/utilities.ts:179:1`

<a id="full-e082ecd054-0016"></a>

### full-e082ecd054-0016: src/compiler/transformers/utilities.ts:222:46

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Statement as ExportDeclaration
~~~~

Exact reason:

~~~~text
a view of Statement as ExportDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ExportDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/utilities.ts:179:1`

<a id="full-e082ecd054-0017"></a>

### full-e082ecd054-0017: src/compiler/transformers/utilities.ts:228:46

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Statement as ExportDeclaration
~~~~

Exact reason:

~~~~text
a view of Statement as ExportDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ExportDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/utilities.ts:179:1`

<a id="full-e082ecd054-0018"></a>

### full-e082ecd054-0018: src/compiler/transformers/utilities.ts:229:45

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Statement as ExportDeclaration
~~~~

Exact reason:

~~~~text
a view of Statement as ExportDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ExportDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/utilities.ts:179:1`

<a id="full-e082ecd054-0019"></a>

### full-e082ecd054-0019: src/compiler/transformers/utilities.ts:230:66

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Statement as ExportDeclaration
~~~~

Exact reason:

~~~~text
a view of Statement as ExportDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ExportDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/utilities.ts:179:1`

<a id="full-e082ecd054-0020"></a>

### full-e082ecd054-0020: src/compiler/transformers/utilities.ts:231:76

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Statement as ExportDeclaration
~~~~

Exact reason:

~~~~text
a view of Statement as ExportDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ExportDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/utilities.ts:179:1`

<a id="full-e082ecd054-0021"></a>

### full-e082ecd054-0021: src/compiler/transformers/utilities.ts:248:58

Observations: {'head': 2}. New in head: True.

Source and target as named by the guard:

~~~~text
Statement as ExportDeclaration
~~~~

Exact reason:

~~~~text
a view of Statement as ExportDeclaration in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as ModuleBlock | SourceFile) through an arm (Node as ModuleBlock) through its field parent (Node as ModuleDeclaration) through its field parent (Node as SourceFile | ModuleBody) through an arm (Node as Identifier) through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)
~~~~

Source text:

~~~~typescript
node as ExportDeclaration
~~~~

Reading: **guard bug: identity/union cross-product.** The reported path first relates `Modifier` to the same `Modifier` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `AccessorKeyword as AbstractKeyword` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless.

Owner units:

- `src/compiler/transformers/utilities.ts:179:1`

<a id="full-e082ecd054-0022"></a>

### full-e082ecd054-0022: src/compiler/transformers/utilities.ts:632:13

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

<a id="full-e082ecd054-0023"></a>

### full-e082ecd054-0023: src/compiler/transformers/utilities.ts:716:47

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

- `src/compiler/transformers/utilities.ts:703:1`

<a id="full-e082ecd054-0024"></a>

### full-e082ecd054-0024: src/compiler/utilitiesPublic.ts:1581:1

Observations: {'head': 2}. New in head: True.

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

- `src/compiler/transformers/utilities.ts:389:1`
- `src/compiler/transformers/utilities.ts:441:1`

<a id="full-e082ecd054-0025"></a>

### full-e082ecd054-0025: src/compiler/utilitiesPublic.ts:1586:1

Observations: {'head': 2}. New in head: True.

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

- `src/compiler/transformers/utilities.ts:389:1`
- `src/compiler/transformers/utilities.ts:441:1`

## base pair check

None.

