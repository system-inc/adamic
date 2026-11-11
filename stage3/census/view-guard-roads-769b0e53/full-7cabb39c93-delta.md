# Paired latent census

| Measure | Raw findings | Distinct reads |
| --- | ---: | ---: |
| Lowered (no remaining finding) | 0 | 0 |
| Moved by reason | 0 | 0 |
| New findings | 1 | 1 |
| New sites | 1 | 1 |
| Units whose first stop moved | 1 | 1 |
| Regressions | 1 | 1 |
| after NotYet | 1 | 1 |

| Reason | Raw before | Raw after | Reads before | Reads after | Reads lowered |
| --- | ---: | ---: | ---: | ---: | ---: |
| a view of Node as SyntaxList in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node \| undefined as Node \| undefined) through its field emitNode (EmitNode \| undefined as EmitNode \| undefined) through its field externalHelpersModuleName (Identifier \| undefined as Identifier \| undefined) through its field jsDoc (JSDocArray \| undefined as JSDocArray \| undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> \| undefined as NodeArray<Modifier> \| undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword) | 0 | 1 | 0 | 1 | 0 |

## Raw Findings

Moved by reason:


New findings:

- `/workspace/roads-adapted/src/compiler/factory/nodeChildren.ts:20:17`: NotYet: a view of Node as SyntaxList in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)

Units whose first stop moved:

- `/workspace/roads-adapted/src/compiler/factory/nodeChildren.ts:20:17`: NotYet: a view of Node as SyntaxList in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)

Panics: 0 before, 0 after.

## Distinct Reads

Moved by reason:


New findings:

- `/workspace/roads-adapted/src/compiler/factory/nodeChildren.ts:20:17`: NotYet: a view of Node as SyntaxList in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)

Units whose first stop moved:

- `/workspace/roads-adapted/src/compiler/factory/nodeChildren.ts:20:17`: NotYet: a view of Node as SyntaxList in a graph program, which the cycle finder doesn't relate yet through its field parent (Node as Node) through its field original (Node | undefined as Node | undefined) through its field emitNode (EmitNode | undefined as EmitNode | undefined) through its field externalHelpersModuleName (Identifier | undefined as Identifier | undefined) through its field jsDoc (JSDocArray | undefined as JSDocArray | undefined) through its index values (JSDoc as JSDoc) through its field parent (HasJSDoc as HasJSDoc) through a source arm (ArrowFunction as HasJSDoc) through an arm (ArrowFunction as ArrowFunction) through its field modifiers (NodeArray<Modifier> | undefined as NodeArray<Modifier> | undefined) through its index values (Modifier as Modifier) through an arm (Modifier as AbstractKeyword) through a source arm (AccessorKeyword as AbstractKeyword)

Panics: 0 before, 0 after.

The JSON retains every lowered site, move, new finding, unit stop and panic.
