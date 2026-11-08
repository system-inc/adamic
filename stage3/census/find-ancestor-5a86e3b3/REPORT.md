# Find ancestor 5a86e3b3 interim checkpoint

Parent 72f23939 finished with 81 complete source records and actual exit 0 recovered from its retained kernel record after the old supervisor was cancelled. The replacement head entry run is still active. Entry counts and regression verdict are pending.

A and B have identical adapter tree IDs, apply keys and all 81 measured source hashes. The final report will reuse A streams for B and run the pairing tool on both applied trees. B apply and its checkpoint passed.

## Six requested checker callers, both A and B

Focused full-mode inventories use the same replay selector as pairing A. They retain the complete entry graph and admit no usable IR. A failed statement may contain multiple calls; its boundary is not evidence of independent execution of every call. These observations are separate from the entry first-error counts.

| Upstream line | Adapted call | Enclosing declaration | Parent boundary stop | Head boundary stop |
| ---: | --- | --- | --- | --- |
| 4748 | compiler/checker.ts:4751:17 | resolveExternalModule at compiler/checker.ts:4733:5 | enclosing statement: NotYet: writing a checked object field without its source-slot type certificate at compiler/binder.ts:2258:13 | enclosing statement: NotYet: writing a checked object field without its source-slot type certificate at compiler/binder.ts:2258:13 |
| 4767 | compiler/checker.ts:4770:21 | resolveExternalModule at compiler/checker.ts:4733:5 | enclosing statement: Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /tmp/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | callee helper orPredicate2 stop recorded in this observing unit; no caller-local failed boundary: NotYet: a checked predicate result returned from a function the checker infers a type predicate for, an ArrowFunction, whose uses the walk doesn't follow, which its direction walk doesn't read at compiler/core.ts:2463:17 |
| 4778 | compiler/checker.ts:4781:21 | resolveExternalModule at compiler/checker.ts:4733:5 | enclosing statement: Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /tmp/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | callee helper orPredicate2 stop recorded in this observing unit; no caller-local failed boundary: NotYet: a checked predicate result returned from a function the checker infers a type predicate for, an ArrowFunction, whose uses the walk doesn't follow, which its direction walk doesn't read at compiler/core.ts:2463:17 |
| 48433 | compiler/checker.ts:48437:47 | checkAliasSymbol at compiler/checker.ts:48398:5 | unit prologue: NotYet: a value of type AliasDeclarationNode at compiler/checker.ts:48398:31 | unit prologue: NotYet: a value of type AliasDeclarationNode at compiler/checker.ts:48398:31 |
| 50073 | compiler/checker.ts:50077:29 | getSymbolOfNameOrPropertyAccessExpression at compiler/checker.ts:49992:5 | enclosing statement: Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /tmp/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | callee helper orPredicate3 stop recorded in this observing unit; no caller-local failed boundary: NotYet: a checked predicate result returned from a function the checker infers a type predicate for, an ArrowFunction, whose uses the walk doesn't follow, which its direction walk doesn't read at compiler/core.ts:2467:17 |
| 50082 | compiler/checker.ts:50086:39 | getSymbolOfNameOrPropertyAccessExpression at compiler/checker.ts:49992:5 | enclosing statement: NotYet: reading result at compiler/checker.ts:50085:22 | callee helper orPredicate2 stop recorded in this observing unit; no caller-local failed boundary: NotYet: a checked predicate result returned from a function the checker infers a type predicate for, an ArrowFunction, whose uses the walk doesn't follow, which its direction walk doesn't read at compiler/core.ts:2463:17 |

A head helper stop is a real callee finding within the selected observing unit. Where the full-mode continuation records its boundary inside core.ts rather than at the caller, the table says so. Repeated calls can share that single helper finding; no independent per-call execution or whole-unit success is claimed.

### Source and first stops for each caller

#### Upstream 4748, adapted compiler/checker.ts:4751:17

Enclosing unit `resolveExternalModule` at `compiler/checker.ts:4733:5`.

Call source:

```typescript
findAncestor(location, orPredicate3(isImportDeclaration, isJSDocImportTag, isExportDeclaration))
```

Parent: enclosing statement: NotYet: writing a checked object field without its source-slot type certificate at compiler/binder.ts:2258:13.

Head: enclosing statement: NotYet: writing a checked object field without its source-slot type certificate at compiler/binder.ts:2258:13.

Focused unit first lowering stop (parent): NotYet: writing a checked object field without its source-slot type certificate at compiler/binder.ts:2258:13.

Focused unit first lowering stop (head): NotYet: writing a checked object field without its source-slot type certificate at compiler/binder.ts:2258:13.

Source `compiler/binder.ts:2256` through line 2260:

```typescript
    function addToContainerChain(next: HasLocals) {
        if (lastContainer) {
            lastContainer.nextContainer = next;
        }

```

Source `compiler/checker.ts:4749` through line 4753:

```typescript
                isVariableDeclarationInitializedToBareOrAccessedRequire(location) && getModuleSpecifierOfBareOrAccessedRequire(location) ||
                findAncestor(location, isImportCall)?.arguments[0] ||
                findAncestor(location, orPredicate3(isImportDeclaration, isJSDocImportTag, isExportDeclaration))?.moduleSpecifier ||
                findAncestor(location, isExternalModuleImportEqualsDeclaration)?.moduleReference.expression;
        const mode = contextSpecifier && isStringLiteralLike(contextSpecifier)
```

#### Upstream 4767, adapted compiler/checker.ts:4770:21

Enclosing unit `resolveExternalModule` at `compiler/checker.ts:4733:5`.

Call source:

```typescript
findAncestor(location, orPredicate2(isImportEqualsDeclaration, isExportDeclaration))
```

Parent: enclosing statement: Refused: overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at /tmp/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1.

Head: callee helper orPredicate2 stop recorded in this observing unit; no caller-local failed boundary: NotYet: a checked predicate result returned from a function the checker infers a type predicate for, an ArrowFunction, whose uses the walk doesn't follow, which its direction walk doesn't read at compiler/core.ts:2463:17.

Focused unit first lowering stop (parent): NotYet: writing a checked object field without its source-slot type certificate at compiler/binder.ts:2258:13.

Focused unit first lowering stop (head): NotYet: writing a checked object field without its source-slot type certificate at compiler/binder.ts:2258:13.

Source `compiler/core.ts:2461` through line 2465:

```typescript
/** @internal */
export function orPredicate2<P, R1 extends P, R2 extends P>(f1: (p1: P) => p1 is R1, f2: (p2: P) => p2 is R2): (p: P) => p is R1 | R2 {
    return p => f1(p) || f2(p);
}
/** @internal */
```

Shared helper finding; a single recorded finding can serve more than one caller of that helper in this unit. This is not independent per-call execution evidence.

Source `compiler/checker.ts:4768` through line 4772:

```typescript
            if (resolvedModule.resolvedUsingTsExtension && isDeclarationFileName(moduleReference)) {
                const importOrExport = findAncestor(location, isImportDeclaration)?.importClause ||
                    findAncestor(location, orPredicate2(isImportEqualsDeclaration, isExportDeclaration));
                if (errorNode && importOrExport && !importOrExport.isTypeOnly || findAncestor(location, isImportCall)) {
                    error(
```

#### Upstream 4778, adapted compiler/checker.ts:4781:21

Enclosing unit `resolveExternalModule` at `compiler/checker.ts:4733:5`.

Call source:

```typescript
findAncestor(location, orPredicate2(isImportEqualsDeclaration, isExportDeclaration))
```

Parent: enclosing statement: Refused: overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at /tmp/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1.

Head: callee helper orPredicate2 stop recorded in this observing unit; no caller-local failed boundary: NotYet: a checked predicate result returned from a function the checker infers a type predicate for, an ArrowFunction, whose uses the walk doesn't follow, which its direction walk doesn't read at compiler/core.ts:2463:17.

Focused unit first lowering stop (parent): NotYet: writing a checked object field without its source-slot type certificate at compiler/binder.ts:2258:13.

Focused unit first lowering stop (head): NotYet: writing a checked object field without its source-slot type certificate at compiler/binder.ts:2258:13.

Source `compiler/core.ts:2461` through line 2465:

```typescript
/** @internal */
export function orPredicate2<P, R1 extends P, R2 extends P>(f1: (p1: P) => p1 is R1, f2: (p2: P) => p2 is R2): (p: P) => p is R1 | R2 {
    return p => f1(p) || f2(p);
}
/** @internal */
```

Shared helper finding; a single recorded finding can serve more than one caller of that helper in this unit. This is not independent per-call execution evidence.

Source `compiler/checker.ts:4779` through line 4783:

```typescript
            else if (resolvedModule.resolvedUsingTsExtension && !shouldAllowImportingTsExtension(compilerOptions, currentSourceFile.fileName)) {
                const importOrExport = findAncestor(location, isImportDeclaration)?.importClause ||
                    findAncestor(location, orPredicate2(isImportEqualsDeclaration, isExportDeclaration));
                if (errorNode && !(importOrExport?.isTypeOnly || findAncestor(location, isImportTypeNode))) {
                    const tsExtension = Debug.checkDefined(tryExtractTSExtension(moduleReference));
```

#### Upstream 48433, adapted compiler/checker.ts:48437:47

Enclosing unit `checkAliasSymbol` at `compiler/checker.ts:48398:5`.

Call source:

```typescript
findAncestor(node, orPredicate2(isImportDeclaration, isImportEqualsDeclaration))
```

Parent: unit prologue: NotYet: a value of type AliasDeclarationNode at compiler/checker.ts:48398:31.

Head: unit prologue: NotYet: a value of type AliasDeclarationNode at compiler/checker.ts:48398:31.

Focused unit first lowering stop (parent): NotYet: a value of type AliasDeclarationNode at compiler/checker.ts:48398:31.

Focused unit first lowering stop (head): NotYet: a value of type AliasDeclarationNode at compiler/checker.ts:48398:31.

Source `compiler/checker.ts:48396` through line 48400:

```typescript
    }

    function checkAliasSymbol(node: AliasDeclarationNode) {
        let symbol = getSymbolOfDeclaration(node);
        const target = resolveAlias(symbol);
```

Source `compiler/checker.ts:48435` through line 48439:

```typescript
                else {
                    Debug.assert(node.kind !== SyntaxKind.VariableDeclaration);
                    const importDeclaration = findAncestor(node, orPredicate2(isImportDeclaration, isImportEqualsDeclaration));
                    const moduleSpecifier = (importDeclaration && tryGetModuleSpecifierFromDeclaration(importDeclaration)?.text) ?? "...";
                    const importedIdentifier = unescapeLeadingUnderscores(isIdentifier(errorNode) ? errorNode.escapedText : symbol.escapedName);
```

#### Upstream 50073, adapted compiler/checker.ts:50077:29

Enclosing unit `getSymbolOfNameOrPropertyAccessExpression` at `compiler/checker.ts:49992:5`.

Call source:

```typescript
findAncestor(name, orPredicate3(isJSDocLinkLike, isJSDocNameReference, isJSDocMemberName))
```

Parent: enclosing statement: Refused: overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at /tmp/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1.

Head: callee helper orPredicate3 stop recorded in this observing unit; no caller-local failed boundary: NotYet: a checked predicate result returned from a function the checker infers a type predicate for, an ArrowFunction, whose uses the walk doesn't follow, which its direction walk doesn't read at compiler/core.ts:2467:17.

Focused unit first lowering stop (parent): NotYet: checked predicate membership over an open tag domain at compiler/checker.ts:50003:18.

Focused unit first lowering stop (head): NotYet: checked predicate membership over an open tag domain at compiler/checker.ts:50003:18.

Source `compiler/core.ts:2465` through line 2469:

```typescript
/** @internal */
export function orPredicate3<P, R1 extends P, R2 extends P, R3 extends P>(f1: (p1: P) => p1 is R1, f2: (p2: P) => p2 is R2, f3: (p3: P) => p3 is R3): (p: P) => p is R1 | R2 | R3 {
    return p => f1(p) || f2(p) || f3(p);
}
/** @internal */
```

Shared helper finding; a single recorded finding can serve more than one caller of that helper in this unit. This is not independent per-call execution evidence.

Source `compiler/checker.ts:50075` through line 50079:

```typescript
            }

            const isJSDoc = findAncestor(name, orPredicate3(isJSDocLinkLike, isJSDocNameReference, isJSDocMemberName));
            const meaning = isJSDoc ? SymbolFlags.Type | SymbolFlags.Namespace | SymbolFlags.Value : SymbolFlags.Value;
            if (name.kind === SyntaxKind.Identifier) {
```

#### Upstream 50082, adapted compiler/checker.ts:50086:39

Enclosing unit `getSymbolOfNameOrPropertyAccessExpression` at `compiler/checker.ts:49992:5`.

Call source:

```typescript
findAncestor(name, orPredicate2(isClassLike, isInterfaceDeclaration))
```

Parent: enclosing statement: NotYet: reading result at compiler/checker.ts:50085:22.

Head: callee helper orPredicate2 stop recorded in this observing unit; no caller-local failed boundary: NotYet: a checked predicate result returned from a function the checker infers a type predicate for, an ArrowFunction, whose uses the walk doesn't follow, which its direction walk doesn't read at compiler/core.ts:2463:17.

Focused unit first lowering stop (parent): NotYet: checked predicate membership over an open tag domain at compiler/checker.ts:50003:18.

Focused unit first lowering stop (head): NotYet: checked predicate membership over an open tag domain at compiler/checker.ts:50003:18.

Source `compiler/core.ts:2461` through line 2465:

```typescript
/** @internal */
export function orPredicate2<P, R1 extends P, R2 extends P>(f1: (p1: P) => p1 is R1, f2: (p2: P) => p2 is R2): (p: P) => p is R1 | R2 {
    return p => f1(p) || f2(p);
}
/** @internal */
```

Shared helper finding; a single recorded finding can serve more than one caller of that helper in this unit. This is not independent per-call execution evidence.

Source `compiler/checker.ts:50084` through line 50088:

```typescript
                const result = resolveEntityName(name, meaning, /*ignoreErrors*/ true, /*dontResolveAlias*/ true, getHostSignatureFromJSDoc(name));
                if (!result && isJSDoc) {
                    const container = findAncestor(name, orPredicate2(isClassLike, isInterfaceDeclaration));
                    if (container) {
                        return resolveJSDocMemberName(name, /*ignoreErrors*/ true, getSymbolOfDeclaration(container));
```

## utilitiesPublic.ts findAncestor itself, both A and B

Implementation at `compiler/utilitiesPublic.ts:788:1`.

```typescript
export function findAncestor(node: Node | undefined, callback: (element: Node) => boolean | "quit"): Node | undefined {
    while (node) {
        const result = callback(node);
        if (result === "quit") {
            return undefined;
        }
        else if (result) {
            return node;
        }
        node = node.parent;
    }
    return undefined;
}
```

Parent focused observation: unit prologue: Refused: overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at /tmp/fa-adapted/src/compiler/utilitiesPublic.ts:792:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1.

Head focused observation: unit prologue: Refused: overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at /tmp/fa-adapted/src/compiler/utilitiesPublic.ts:792:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1.

parent:
Source `compiler/utilitiesPublic.ts:784` through line 788:

```typescript
 * At that point findAncestor returns undefined.
 */
export function findAncestor<T extends Node>(node: Node | undefined, callback: (element: Node) => element is T): T | undefined;
export function findAncestor(node: Node | undefined, callback: (element: Node) => boolean | "quit"): Node | undefined;
export function findAncestor(node: Node | undefined, callback: (element: Node) => boolean | "quit"): Node | undefined {
```

head:
Source `compiler/utilitiesPublic.ts:784` through line 788:

```typescript
 * At that point findAncestor returns undefined.
 */
export function findAncestor<T extends Node>(node: Node | undefined, callback: (element: Node) => element is T): T | undefined;
export function findAncestor(node: Node | undefined, callback: (element: Node) => boolean | "quit"): Node | undefined;
export function findAncestor(node: Node | undefined, callback: (element: Node) => boolean | "quit"): Node | undefined {
```

