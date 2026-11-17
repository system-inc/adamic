# Construction families and representation context

These are construction families, not claimed origins for the seven untraced AST inputs/results. All coordinates refer to the adapted tree.

## factory/nodeFactory.ts:1169 createNodeArray

```typescript
function createNodeArray<T extends Node>(elements?: readonly T[], hasTrailingComma?: boolean): NodeArray<T> {
        if (elements === undefined || elements === emptyArray) {
            elements = [];
        }
        else if (isNodeArray(elements)) {
            if (hasTrailingComma === undefined || elements.hasTrailingComma === hasTrailingComma) {
                // Ensure the transform flags have been aggregated for this NodeArray
                if (elements.transformFlags === undefined) {
                    aggregateChildrenFlags(elements as MutableNodeArray<T>);
                }
                Debug.attachNodeArrayDebugInfo(elements);
                return elements;
            }

            // This *was* a `NodeArray`, but the `hasTrailingComma` option differs. Recreate the
            // array with the same elements, text range, and transform flags but with the updated
            // value for `hasTrailingComma`
            const array = elements.slice() as MutableNodeArray<T>;
            array.pos = elements.pos;
            array.end = elements.end;
            array.hasTrailingComma = hasTrailingComma;
            array.transformFlags = elements.transformFlags;
            Debug.attachNodeArrayDebugInfo(array);
            return array;
        }

        // Since the element list of a node array is typically created by starting with an empty array and
        // repeatedly calling push(), the list may not have the optimal memory layout. We invoke slice() for
        // small arrays (1 to 4 elements) to give the VM a chance to allocate an optimal representation.
        const length = elements.length;
        const array = (length >= 1 && length <= 4 ? elements.slice() : elements) as MutableNodeArray<T>;
        array.pos = -1;
        array.end = -1;
        array.hasTrailingComma = !!hasTrailingComma;
        array.transformFlags = TransformFlags.None;
        aggregateChildrenFlags(array);
        Debug.attachNodeArrayDebugInfo(array);
        return array;
    }
```

## factory/nodeFactory.ts:2778 createObjectBindingPattern

```typescript
function createObjectBindingPattern(elements: readonly BindingElement[]) {
        const node = createBaseNode<ObjectBindingPattern>(SyntaxKind.ObjectBindingPattern);
        node.elements = createNodeArray(elements);
        node.transformFlags |= propagateChildrenFlags(node.elements) |
            TransformFlags.ContainsES2015 |
            TransformFlags.ContainsBindingPattern;
        if (node.transformFlags & TransformFlags.ContainsRestOrSpread) {
            node.transformFlags |= TransformFlags.ContainsES2018 |
                TransformFlags.ContainsObjectRestOrSpread;
        }
        return node;
    }
```

## factory/nodeFactory.ts:2799 createArrayBindingPattern

```typescript
function createArrayBindingPattern(elements: readonly ArrayBindingElement[]) {
        const node = createBaseNode<ArrayBindingPattern>(SyntaxKind.ArrayBindingPattern);
        node.elements = createNodeArray(elements);
        node.transformFlags |= propagateChildrenFlags(node.elements) |
            TransformFlags.ContainsES2015 |
            TransformFlags.ContainsBindingPattern;
        return node;
    }
```

## factory/nodeFactory.ts:2869 createObjectLiteralExpression

```typescript
function createObjectLiteralExpression(properties?: readonly ObjectLiteralElementLike[], multiLine?: boolean) {
        const node = createBaseDeclaration<ObjectLiteralExpression>(SyntaxKind.ObjectLiteralExpression);
        node.properties = createNodeArray(properties);
        node.multiLine = multiLine;
        node.transformFlags |= propagateChildrenFlags(node.properties);

        node.jsDoc = undefined; // initialized by parser (JsDocContainer)
        return node;
    }
```

## factory/nodeFactory.ts:5170 createJSDocSignature

```typescript
function createJSDocSignature(typeParameters: readonly JSDocTemplateTag[] | undefined, parameters: readonly JSDocParameterTag[], type?: JSDocReturnTag): JSDocSignature {
        const node = createBaseDeclaration<JSDocSignature>(SyntaxKind.JSDocSignature);
        node.typeParameters = asNodeArray(typeParameters);
        node.parameters = createNodeArray(parameters);
        node.type = type;

        node.jsDoc = undefined; // initialized by parser (JsDocContainer)
        node.locals = undefined; // initialized by binder (LocalsContainer)
        node.nextContainer = undefined; // initialized by binder (LocalsContainer)
        return node;
    }
```

## factory/nodeFactory.ts:7137 asNodeArray

```typescript
function asNodeArray<T extends Node>(array: readonly T[] | undefined): NodeArray<T> | undefined {
        return array ? createNodeArray(array) : undefined;
    }
```

## visitorPublic.ts:223 visitNodes

```typescript
export function visitNodes(
    nodes: NodeArray<Node> | undefined,
    visitor: Visitor,
    test?: (node: Node) => boolean,
    start?: number,
    count?: number,
): NodeArray<Node> | undefined {
    if (nodes === undefined) {
        // If the input type is undefined, then the output type can be undefined.
        return nodes;
    }

    // Ensure start and count have valid values
    const length = nodes.length;
    if (start === undefined || start < 0) {
        start = 0;
    }

    if (count === undefined || count > length - start) {
        count = length - start;
    }

    let hasTrailingComma: boolean | undefined;
    let pos = -1;
    let end = -1;
    if (start > 0 || count < length) {
        // Since this is a fragment of a node array, we do not copy over the previous location
        // and will only copy over `hasTrailingComma` if we are including the last element.
        hasTrailingComma = nodes.hasTrailingComma && start + count === length;
    }
    else {
        pos = nodes.pos;
        end = nodes.end;
        hasTrailingComma = nodes.hasTrailingComma;
    }

    const updated = visitArrayWorker(nodes, visitor, test, start, count);
    if (updated !== nodes as readonly Node[]) {
        // TODO(rbuckton): Remove dependency on `ts.factory` in favor of a provided factory.
        const updatedArray = factory.createNodeArray(updated, hasTrailingComma);
        setTextRangePosEnd(updatedArray, pos, end);
        return updatedArray;
    }

    // If we are here, updated === nodes. This means that it's still a NodeArray,
    // and also that its contents passed the tests in visitArrayWorker, so has contents
    // of type TOut.
    return nodes;
}
```

## visitorPublic.ts:328 visitArrayWorker

```typescript
function visitArrayWorker(
    nodes: readonly Node[],
    visitor: Visitor | undefined,
    test: ((node: Node) => boolean) | undefined,
    start: number,
    count: number,
): readonly Node[] {
    let updated: Node[] | undefined;

    const length = nodes.length;
    if (start > 0 || count < length) {
        // If we are not visiting all of the original nodes, we must always create a new array.
        updated = [];
    }

    // Visit each original node.
    for (let i = 0; i < count; i++) {
        const node = nodes[i + start];
        const visited = node !== undefined ? (visitor ? visitor(node) : node) : undefined;
        if (updated !== undefined || visited === undefined || visited !== node) {
            if (updated === undefined) {
                // Ensure we have a copy of `nodes`, up to the current index.
                updated = nodes.slice(0, i);
                Debug.assertEachNode(updated, test);
            }
            if (visited) {
                if (isArray(visited)) {
                    for (const visitedNode of visited) {
                        Debug.assertNode(visitedNode, test);
                        updated.push(visitedNode);
                    }
                }
                else {
                    Debug.assertNode(visited, test);
                    updated.push(visited);
                }
            }
        }
    }

    if (updated) {
        // If we have an updated array, then all items will have been tested.
        return updated;
    }

    // If we are going to return the original array, ensure it passes the test.
    Debug.assertEachNode(nodes, test);
    return nodes;
}
```

## utilities.ts:646 createSymbolTable

```typescript
export function createSymbolTable(symbols?: readonly Symbol[]): SymbolTable {
    const result = new Map<__String, Symbol>();
    if (symbols) {
        for (const symbol of symbols) {
            result.set(symbol.escapedName, symbol);
        }
    }
    return result;
}
```

## utilities.ts:8507 Node

```typescript
function Node(this: Mutable<Node>, kind: SyntaxKind, pos: number, end: number): void {
    // Note: if modifying this, be sure to update NodeObject in src/services/services.ts
    this.pos = pos;
    this.end = end;
    this.kind = kind;
    this.id = 0;
    this.flags = NodeFlags.None;
    this.modifierFlagsCache = ModifierFlags.None;
    this.transformFlags = TransformFlags.None;
    this.parent = undefined!;
    this.original = undefined;
    this.emitNode = undefined;
}
```

## utilities.ts:8521 Token

```typescript
function Token(this: Mutable<Node>, kind: SyntaxKind, pos: number, end: number): void {
    // Note: if modifying this, be sure to update TokenOrIdentifierObject in src/services/services.ts
    this.pos = pos;
    this.end = end;
    this.kind = kind;
    this.id = 0;
    this.flags = NodeFlags.None;
    this.transformFlags = TransformFlags.None;
    this.parent = undefined!;
    this.emitNode = undefined;
}
```

## parser.ts:476 isAnExternalModuleIndicatorNode

```typescript
function isAnExternalModuleIndicatorNode(node: Node) {
    return canHaveModifiers(node) && hasModifierOfKind(node, SyntaxKind.ExportKeyword)
            || isImportEqualsDeclaration(node) && isExternalModuleReference(node.moduleReference)
            || isImportDeclaration(node)
            || isExportAssignment(node)
            || isExportDeclaration(node) ? node : undefined;
}
```

## parser.ts:2567 createNodeArray

```typescript
function createNodeArray<T extends Node>(elements: T[], pos: number, end?: number, hasTrailingComma?: boolean): NodeArray<T> {
        const array = factory.createNodeArray(elements, hasTrailingComma);
        setTextRangePosEnd(array, pos, end ?? scanner.getTokenFullStart());
        return array;
    }
```

## factory/baseNodeFactory.ts:53 createBaseTokenNode

```typescript
function createBaseTokenNode(kind: SyntaxKind): Node {
        return new (TokenConstructor || (TokenConstructor = objectAllocator.getTokenConstructor()))(kind, /*pos*/ -1, /*end*/ -1);
    }
```

## factory/baseNodeFactory.ts:57 createBaseNode

```typescript
function createBaseNode(kind: SyntaxKind): Node {
        return new (NodeConstructor || (NodeConstructor = objectAllocator.getNodeConstructor()))(kind, /*pos*/ -1, /*end*/ -1);
    }
```

## types.ts:942 (context excerpt)

```typescript
export interface Node extends ReadonlyTextRange {
    readonly kind: SyntaxKind;
    readonly flags: NodeFlags;
    /** @internal */ modifierFlagsCache: ModifierFlags;
    /** @internal */ readonly transformFlags: TransformFlags; // Flags for transforms
    /** @internal */ id?: NodeId | undefined; // Unique id (used to look up NodeLinks)
    readonly parent: Node; // Parent node (initialized by binding)
    /** @internal */ original?: Node | undefined; // The original node if this is an updated node.
    /** @internal */ emitNode?: EmitNode | undefined; // Associated EmitNode (initialized by transforms)
    // NOTE: `symbol` and `localSymbol` have been moved to `Declaration`
    //       `locals` and `nextContainer` have been moved to `LocalsContainer`
    //       `flowNode` has been moved to `FlowContainer`
    //       see: https://github.com/microsoft/TypeScript/pull/51682
}

export interface JSDocContainer extends Node {
    _jsdocContainerBrand: undefined;
    /** @internal */ jsDoc?: JSDocArray | undefined; // JSDoc that directly precedes this node
}

/** @internal */
export interface JSDocArray extends Array<JSDoc> {
    jsDocCache?: readonly JSDocTag[] | undefined; // Cache for getJSDocTags
}

export interface LocalsContainer extends Node {
    _localsContainerBrand: undefined;
    /** @internal */ locals?: SymbolTable | undefined; // Locals associated with node (initialized by binding)
    /** @internal */ nextContainer?: HasLocals | undefined; // Next container in declaration order (initialized by binding)
}

export interface FlowContainer extends Node {
```

## types.ts:1589 (context excerpt)

```typescript
export interface NodeArray<T extends Node> extends ReadonlyArray<T>, ReadonlyTextRange {
    readonly hasTrailingComma: boolean;
    /** @internal */ transformFlags: TransformFlags; // Flags for transforms, possibly undefined
}

// TODO(rbuckton): Constraint 'TKind' to 'TokenSyntaxKind'
export interface Token<TKind extends SyntaxKind> extends Node {
    readonly kind: TKind;
}

export type EndOfFileToken = Token<SyntaxKind.EndOfFileToken> & JSDocContainer;

```

## types.ts:6048 (context excerpt)

```typescript
export interface Symbol {
    flags: SymbolFlags;                     // Symbol flags
    escapedName: __String;                  // Name of symbol
    declarations?: Declaration[] | undefined;           // Declarations associated with this symbol
    valueDeclaration?: Declaration | undefined;         // First value declaration of the symbol
    members?: SymbolTable | undefined;                  // Class, interface or object literal instance members
    exports?: SymbolTable | undefined;                  // Module exports
    globalExports?: SymbolTable;            // Conditional global UMD exports
    /** @internal */ id: SymbolId;          // Unique id (used to look up SymbolLinks)
    /** @internal */ mergeId: number;       // Merge id (used to look up merged symbol)
    /** @internal */ parent?: Symbol | undefined;       // Parent symbol
    /** @internal */ exportSymbol?: Symbol | undefined; // Exported symbol associated with this symbol
    /** @internal */ constEnumOnlyModule: boolean | undefined; // True if module contains only const enums or other modules with only const enums
    /** @internal */ isReferenced?: SymbolFlags | undefined; // True if the symbol is referenced elsewhere. Keeps track of the meaning of a reference in case a symbol is both a type parameter and parameter.
```

## types.ts:6216 (context excerpt)

```typescript
export type SymbolTable = Map<__String, Symbol>;
```

## types.ts:10440 (context excerpt)

```typescript
export interface ReadonlyPragmaMap extends ReadonlyMap<string, PragmaPseudoMap[keyof PragmaPseudoMap] | PragmaPseudoMap[keyof PragmaPseudoMap][]> {
    get<TKey extends keyof PragmaPseudoMap>(key: TKey): PragmaPseudoMap[TKey] | PragmaPseudoMap[TKey][];
    forEach(action: <TKey extends keyof PragmaPseudoMap>(value: PragmaPseudoMap[TKey] | PragmaPseudoMap[TKey][], key: TKey, map: ReadonlyPragmaMap) => void): void;
}

/**
 * A strongly-typed es6 map of pragma entries, the values of which are either a single argument
 * value (if only one was found), or an array of multiple argument values if the pragma is present
 * in multiple places
 *
 * @internal
 */
export interface PragmaMap extends Map<string, PragmaPseudoMap[keyof PragmaPseudoMap] | PragmaPseudoMap[keyof PragmaPseudoMap][]>, ReadonlyPragmaMap {
    set<TKey extends keyof PragmaPseudoMap>(key: TKey, value: PragmaPseudoMap[TKey] | PragmaPseudoMap[TKey][]): this;
    get<TKey extends keyof PragmaPseudoMap>(key: TKey): PragmaPseudoMap[TKey] | PragmaPseudoMap[TKey][];
    forEach(action: <TKey extends keyof PragmaPseudoMap>(value: PragmaPseudoMap[TKey] | PragmaPseudoMap[TKey][], key: TKey, map: PragmaMap) => void): void;
}
```

## core.ts:19 (context excerpt)

```typescript
export const emptyMap: ReadonlyMap<never, never> = new Map<never, never>();
```

## parser.ts:1588 (context excerpt)

```typescript
            const result = parseJsonText(fileName, sourceText, languageVersion, syntaxCursor, setParentNodes);
            convertToJson(result, result.statements[0]?.expression, result.parseDiagnostics, /*returnValue*/ false, /*jsonConversionNotifier*/ undefined);
            result.referencedFiles = emptyArray;
            result.typeReferenceDirectives = emptyArray;
            result.libReferenceDirectives = emptyArray;
            result.amdDependencies = emptyArray;
            result.hasNoDefaultLib = false;
            result.pragmas = emptyMap as ReadonlyPragmaMap;
            return result;
        }
```

## parser.ts:10564 (context excerpt)

```typescript
export function processCommentPragmas(context: PragmaContext, sourceText: string): void {
    const pragmas: PragmaPseudoMapEntry[] = [];

    for (const range of getLeadingCommentRanges(sourceText, 0) || emptyArray) {
        const comment = sourceText.substring(range.pos, range.end);
        extractPragmas(pragmas, range, comment);
    }

    context.pragmas = new Map() as PragmaMap;
    for (const pragma of pragmas) {
        if (context.pragmas.has(pragma.name)) {
            const currentValue = context.pragmas.get(pragma.name);
            if (currentValue instanceof Array) {
                currentValue.push(pragma.args);
            }
            else {
                context.pragmas.set(pragma.name, [currentValue, pragma.args]);
            }
            continue;
        }
        context.pragmas.set(pragma.name, pragma.args);
    }
}

```

## utilities.ts:8554 (context excerpt)

```typescript
export const objectAllocator: ObjectAllocator = {
    getNodeConstructor: (): ReturnType<ObjectAllocator["getNodeConstructor"]> => Node as any,
    getTokenConstructor: () => Token as any,
    getIdentifierConstructor: () => Identifier as any,
    getPrivateIdentifierConstructor: () => Node as any,
```
