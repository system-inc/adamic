# full-5dc3fc14d8: every guard finding in full mode

Reasons and type names are verbatim. Source slices retain their text, with CRLF displayed as LF. Repeated observations are grouped only when site, source, target, reason and source text agree; every owner unit is listed. Road is the first new mechanism in the actual path; the exact reason preserves every subsequent mechanism.

## elements

None.

## index values

<a id="full-5dc3fc14d8-0001"></a>

### full-5dc3fc14d8-0001: src/compiler/factory/nodeTests.ts:397:1

Observations: {'head': 1}. New in head: True.

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

<a id="full-5dc3fc14d8-0002"></a>

### full-5dc3fc14d8-0002: src/compiler/factory/nodeTests.ts:668:1

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

- `src/compiler/transformers/destructuring.ts:446:1`

<a id="full-5dc3fc14d8-0003"></a>

### full-5dc3fc14d8-0003: src/compiler/utilitiesPublic.ts:1498:1

Observations: {'head': 1}. New in head: True.

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

## union arms

<a id="full-5dc3fc14d8-0004"></a>

### full-5dc3fc14d8-0004: src/compiler/core.ts:140:1

Observations: {'head': 1}. New in head: True.

Source and target as named by the guard:

~~~~text
BindingOrAssignmentElement as OmittedExpression
~~~~

Exact reason:

~~~~text
a view of BindingOrAssignmentElement as OmittedExpression in a graph program, which the cycle finder doesn't relate yet through a source arm (ArrayLiteralExpression as OmittedExpression)
~~~~

Source text:

~~~~typescript
/**
 * Iterates through `array` by index and performs the callback on each element of array until the callback
 * returns a falsey value, then returns false.
 * If no such value is found, the callback is applied to each element of array and `true` is returned.
 *
 * @internal
 */
export function every<T, U extends T>(array: readonly T[], callback: (element: T, index: number) => element is U): array is readonly U[];
~~~~

Reading: **guard bug: checked every predicate.** Guard bug. flattenArrayBindingOrAssignmentPattern calls every(elements, isOmittedExpression). every returns true only after that predicate succeeds for every element; isOmittedExpression checks kind === SyntaxKind.OmittedExpression. Its resulting array view is readonly. ArrayLiteralExpression is not admitted to the OmittedExpression element alternative, and this readonly view cannot store an incompatible replacement element into the source. The leaf kinds are disjoint, but the unconditional source-arm check has ignored the checked element refinement.

Owner units:

- `src/compiler/transformers/destructuring.ts:446:1`

<a id="full-5dc3fc14d8-0005"></a>

### full-5dc3fc14d8-0005: src/compiler/utilitiesPublic.ts:1884:1

Observations: {'head': 3}. New in head: True.

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
- `src/compiler/transformers/destructuring.ts:522:1`

<a id="full-5dc3fc14d8-0006"></a>

### full-5dc3fc14d8-0006: src/compiler/utilitiesPublic.ts:1884:1

Observations: {'head': 3}. New in head: True.

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
- `src/compiler/transformers/destructuring.ts:522:1`

<a id="full-5dc3fc14d8-0007"></a>

### full-5dc3fc14d8-0007: src/compiler/utilitiesPublic.ts:1884:1

Observations: {'head': 3}. New in head: True.

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
- `src/compiler/transformers/destructuring.ts:522:1`

<a id="full-5dc3fc14d8-0008"></a>

### full-5dc3fc14d8-0008: src/compiler/utilitiesPublic.ts:1884:1

Observations: {'head': 3}. New in head: True.

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
- `src/compiler/transformers/destructuring.ts:522:1`

<a id="full-5dc3fc14d8-0009"></a>

### full-5dc3fc14d8-0009: src/compiler/utilitiesPublic.ts:1894:1

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

- `src/compiler/transformers/destructuring.ts:338:1`

<a id="full-5dc3fc14d8-0010"></a>

### full-5dc3fc14d8-0010: src/compiler/utilitiesPublic.ts:1894:1

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

- `src/compiler/transformers/destructuring.ts:338:1`

<a id="full-5dc3fc14d8-0011"></a>

### full-5dc3fc14d8-0011: src/compiler/utilitiesPublic.ts:1921:1

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

- `src/compiler/transformers/destructuring.ts:338:1`

<a id="full-5dc3fc14d8-0012"></a>

### full-5dc3fc14d8-0012: src/compiler/utilitiesPublic.ts:1921:1

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

- `src/compiler/transformers/destructuring.ts:338:1`

## base pair check

None.

