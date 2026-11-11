# full-7cabb39c93: every guard finding in full mode

Reasons and type names are verbatim. Source slices retain their text, with CRLF displayed as LF. Repeated observations are grouped only when site, source, target, reason and source text agree; every owner unit is listed. Road is the first new mechanism in the actual path; the exact reason preserves every subsequent mechanism.

## elements

None.

## index values

### full-7cabb39c93-0001: src/compiler/factory/nodeChildren.ts:20:17

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

## union arms

None.

## base pair check

None.

