# Find ancestor: paired latent first-error census

Task 448a51k. Report only. No compiler or adaptation change is committed.

| Compiler | Pin |
| --- | --- |
| Parent and apply | `72f23939c773ba0045aeed5d6afb9b025b6473a0` |
| Head | `5a86e3b3588914eb367f800eb7339db8d1bd70f9` |

The head admits a predicate overload result when a const holds the callback answer and checker flow proves the tested answer boolean, subject to adjacency and binding guards. The adapted input comes from the unmodified parent apply, using TypeScript 6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`. Both compiler binaries and their guarded overlays were built from their own pins; checker dependency `06af2e399c6c1fb93fd570b774027942747400b4` is shared unchanged.

The entry is `src/tsc/tsc.ts`, with 81 reached source files and 0 checker diagnostics on each compiler. Entry counts use latent **first-error** mode. `LATENT_ASSERT_NO_OUTPUT=1` checked both no-output guards. No full-core census, second pairing, native build or full gate ran. Focused full-unit inventories use pairing A's replay selector after the entry pass, and are reported separately.

## Pairings A and B

A is the unchanged parent apply. For B, a separate scratch parent explicitly received both complete directories from `eeb87728b743235cf1df68dfed17241f63416734` and applied successfully. Both directories already exist unchanged on main 72f23939: their Git tree IDs, apply product keys, and all 81 measured source hashes are identical. [Equivalence proof](AB-equivalence.json) and [B source hashes](B-source-hashes.json) retain the checks. Thus B reuses A census streams and focused observations on identical bytes, while the pair tool runs separately over the independently applied B tree. Both paired JSON results are exactly equal. No duplicate parent census was run.

The initially requested e6887750 head was stopped with SIGTERM after the user replaced it. Its partial stream is archived under `superseded/` and contributes no counts or verdict. The retained parent is the original 72f23939 run. The replacement head was checked out and rebuilt from 5a86e3b3. A staging-only sparse-checkout exclusion after renaming the directory was corrected before starting the replacement head; [initial resume log](resume-initial.log) is retained.

## Counts, each pairing

The following counts apply independently to both A and B.

| Pairing | View | Cleared findings | New findings | Changed findings | New sites | First-stop unit changes | Regression entries |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |
| A | Raw | 31 | 7 | 1 | 7 | 13 | 5 |
| A | Distinct outermost access chains | 31 | 7 | 1 | 7 | 13 | 5 |
| B | Raw | 31 | 7 | 1 | 7 | 13 | 5 |
| B | Distinct outermost access chains | 31 | 7 | 1 | 7 | 13 | 5 |

Raw means pair.py exact raw-position keys, not event multiplicity. Distinct uses the stock TypeScript 6.0.3 outermost-access-chain normalization. New findings include replacement reasons, so new and changed are not disjoint. A newly stopped unit was finding-free during latent lowering before and now has a lowering stop. A different first stop in an already-blocked unit is listed separately. This is a lowering observation, not proof of previous native correctness.

## Every newly stopped unit (0)

| Unit | Unit location | Head stop site | Exact head reason |
| --- | --- | --- | --- |
| None | | | |

## Every changed first stop in an already-blocked unit (9)

| Unit | Unit location | Parent site and exact reason | Head site and exact reason |
| --- | --- | --- | --- |
| transformES2016 | compiler/transformers/es2016.ts:21:1 | Refused: overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor at compiler/visitorPublic.ts:587:119 | Refused: overload 1 of visitEachChild parameter node cannot be served by implementation parameter node at compiler/visitorPublic.ts:585:48 |
| transformES2019 | compiler/transformers/es2019.ts:17:1 | Refused: overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor at compiler/visitorPublic.ts:587:119 | Refused: overload 1 of visitEachChild parameter node cannot be served by implementation parameter node at compiler/visitorPublic.ts:585:48 |
| transformES2020 | compiler/transformers/es2020.ts:42:1 | Refused: overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor at compiler/visitorPublic.ts:587:119 | Refused: overload 1 of visitEachChild parameter node cannot be served by implementation parameter node at compiler/visitorPublic.ts:585:48 |
| transformES2021 | compiler/transformers/es2021.ts:25:1 | Refused: overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor at compiler/visitorPublic.ts:587:119 | Refused: overload 1 of visitEachChild parameter node cannot be served by implementation parameter node at compiler/visitorPublic.ts:585:48 |
| transformLegacyDecorators | compiler/transformers/legacyDecorators.ts:88:1 | Refused: overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor at compiler/visitorPublic.ts:587:119 | Refused: overload 1 of visitEachChild parameter node cannot be served by implementation parameter node at compiler/visitorPublic.ts:585:48 |
| getContainingFunction | compiler/utilities.ts:3125:1 | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /tmp/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | NotYet: writing a checked object field without its source-slot type certificate at compiler/binder.ts:1034:17 |
| getContainingFunctionDeclaration | compiler/utilities.ts:3130:1 | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /tmp/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | NotYet: writing a checked object field without its source-slot type certificate at compiler/binder.ts:1034:17 |
| getContainingFunctionOrClassStaticBlock | compiler/utilities.ts:3150:1 | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /tmp/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | NotYet: writing a checked object field without its source-slot type certificate at compiler/binder.ts:1034:17 |
| isPartOfTypeOnlyImportOrExportDeclaration | compiler/utilitiesPublic.ts:1566:1 | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /tmp/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | NotYet: checked predicate overload target ExportDeclaration & { readonly isTypeOnly: true; readonly moduleSpecifier: Expression; } at compiler/utilitiesPublic.ts:1562:1 |

## Cleared unit first stops (4)

- `isJSDocTypeExpressionOrChild` at `compiler/utilities.ts:1291:1`: Refused: overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at /tmp/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 -> finding-free.
- `getContainingClass` at `compiler/utilities.ts:3135:1`: Refused: overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at /tmp/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 -> finding-free.
- `getJSDocRoot` at `compiler/utilities.ts:4823:1`: Refused: overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at /tmp/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 -> finding-free.
- `isSideEffectImport` at `compiler/utilities.ts:12088:1`: Refused: overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at /tmp/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 -> finding-free.

## Entry unit containing all six checker calls

- parent: `compiler/checker.ts:1489:1`: NotYet: a value of type ParameterPropertyDeclaration at compiler/checker.ts:2971:55.
- head: `compiler/checker.ts:1489:1`: NotYet: a value of type ParameterPropertyDeclaration at compiler/checker.ts:2971:55.

## Entry observations of utilitiesPublic.ts findAncestor

Deferred declarations have no measured lowering stop and are not counted as lowered bodies.

| Compiler | Unit | Status | Instantiated | First lowering stop |
| --- | --- | --- | --- | --- |
| parent | compiler/utilitiesPublic.ts:786:1 | deferred_generic_instantiation | True | finding-free |
| parent | compiler/utilitiesPublic.ts:787:1 | attempted | False | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /tmp/fa-adapted/src/compiler/utilitiesPublic.ts:792:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 |
| parent | compiler/utilitiesPublic.ts:788:1 | attempted | False | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /tmp/fa-adapted/src/compiler/utilitiesPublic.ts:792:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 |
| head | compiler/utilitiesPublic.ts:786:1 | deferred_generic_instantiation | True | finding-free |
| head | compiler/utilitiesPublic.ts:787:1 | attempted | False | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /tmp/fa-adapted/src/compiler/utilitiesPublic.ts:792:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 |
| head | compiler/utilitiesPublic.ts:788:1 | attempted | False | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /tmp/fa-adapted/src/compiler/utilitiesPublic.ts:792:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 |

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

Earlier whole-unit stop source:
Source `compiler/binder.ts:2256` through line 2260:

```typescript
    function addToContainerChain(next: HasLocals) {
        if (lastContainer) {
            lastContainer.nextContainer = next;
        }

```

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

Earlier whole-unit stop source:
Source `compiler/binder.ts:2256` through line 2260:

```typescript
    function addToContainerChain(next: HasLocals) {
        if (lastContainer) {
            lastContainer.nextContainer = next;
        }

```

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

Earlier whole-unit stop source:
Source `compiler/checker.ts:50001` through line 50005:

```typescript
        ) {
            // Check if this is a special property assignment
            if (!isPrivateIdentifier(name) && !isJSDocMemberName(name) && !isThisPropertyAndThisTyped(name.parent as PropertyAccessExpression)) {
                const specialPropertyAssignmentSymbol = getSpecialPropertyAssignmentSymbolFromEntityName(name);
                if (specialPropertyAssignmentSymbol) {
```

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

Earlier whole-unit stop source:
Source `compiler/checker.ts:50001` through line 50005:

```typescript
        ) {
            // Check if this is a special property assignment
            if (!isPrivateIdentifier(name) && !isJSDocMemberName(name) && !isThisPropertyAndThisTyped(name.parent as PropertyAccessExpression)) {
                const specialPropertyAssignmentSymbol = getSpecialPropertyAssignmentSymbolFromEntityName(name);
                if (specialPropertyAssignmentSymbol) {
```

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


## All retained run resources

[Resource index](resources.md) includes every retained status/time pair, including focused observations and the cancelled superseded head. The parent entry peak is unavailable for the reason recorded in the kernel-exit evidence.



## Panics and regression details

- before: 0 raw-position panic findings.
- after: 0 raw-position panic findings.
- previously_finding_free_scan_site: Refused: a function taking Node seen as one taking NonNullable<BigIntLiteral | Identifier | NumericLiteral | PrivateIdentifier | StringLiteralLike> (tsc relates a method's parameters both ways), so it can be handed what it can't take at compiler/transformers/esDecorators.ts:1931:42
- previously_finding_free_scan_site: Refused: a function taking Node seen as one taking NonNullable<BigIntLiteral | Identifier | NoSubstitutionTemplateLiteral | NumericLiteral | PrivateIdentifier | StringLiteral> (tsc relates a method's parameters both ways), so it can be handed what it can't take at compiler/transformers/esDecorators.ts:1954:32
- previously_finding_free_scan_site: Refused: a function taking (hint: EmitHint, node: Node) => void seen as one taking (hint: EmitHint, node: Node) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take at compiler/transformers/es2017.ts:163:26
- previously_finding_free_scan_site: Refused: a function taking (hint: EmitHint, node: Node) => void seen as one taking (hint: EmitHint, node: Node) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take at compiler/transformers/es2018.ts:168:26
- previously_finding_free_scan_site: Refused: a function taking Node seen as one taking NonNullable<MethodDeclaration | ShorthandPropertyAssignment | AccessorDeclaration> (tsc relates a method's parameters both ways), so it can be handed what it can't take at compiler/transformers/es2018.ts:499:40

## Time and peak memory

| Run | Child exit | Wall seconds | Peak RSS MiB |
| --- | ---: | ---: | ---: |
| parent | 0 | 3306.351 estimated | unavailable |
| head | 0 | 3426.550 | 8321.0 |

Head peak RSS is Linux child `getrusage` in KiB converted to MiB. Parent peak RSS is unavailable: cancelling the superseded controller also ended its wrapper, while the compiler continued in its own session and finished. Its actual exit 0 is recovered from the retained Linux dead-task record, field 52 of `/proc/PID/stat`; [kernel exit evidence](parent-kernel-exit.json) and [recovery script](recover-parent.py) retain the exact record and checks. Parent wall time is an estimate from its kernel start and stream final-write timestamp. Its 81 file records and source hashes are complete and verified. The parent is reused rather than rerun. Wall observations include up to one second of polling resolution. The parent is reused from the original launch. The replacement head overlaps its final files; the original [launch record](head-start.json) belongs to the superseded head and is retained only as historical evidence. B reuses the same measurements, so it has no additional run time or peak memory. These are shared-host observations, not benchmarks. Wrapper status and resource files retain child exit, wall time and peak RSS where the wrapper survived. The bounded runner stops a child after 30 minutes without output/progress. Unit timing streams provide progress even while the census buffers file records.

## Evidence and validation

- A: [parent stream](parent-run.jsonl), [head stream](head-run.jsonl), [paired JSON](paired.json), and [paired ledger](paired.md).
- B: [reused parent stream](B-parent-run.jsonl), [reused head stream](B-head-run.jsonl), [independent paired JSON](B-paired.json), and [paired ledger](B-paired.md).
- [Newly stopped units](newly-stopped.json), [changed first stops](changed-first-stops.json), and [named caller observations and source quotations](named-sites.json).
- [Applied patch set](patch-set.md), [all measured source hashes](source-hashes.json), [compiler/source/binary provenance](provenance.json), and [guarded overlays](overlays/). Go evidence uses `.go.txt` filenames.
- [Parent timings](parent-run-units.jsonl), [head timings](head-run-units.jsonl), run logs and `*.status.json` / `*.time.json` files retain commands, own child exits and resources.
- The independent audit below recounts raw keys, unit first stops and newly stopped units, verifies every input and compiler source hash, and checks the 81-file scope against the measurement header. No measured source was removed or changed. Only unused upstream tests were removed after apply to reserve RAM.

This runtime started with fresh temporary directories. Cloud setup initialized and verified the pinned checker, and both measurement drivers were rebuilt. The API packages were installed with `npm ci`. The parent apply and both own-pin builds exited 0. All 15 existing paired-census tests passed, including its normalization and regression checks. Logs are retained and never piped into head or tail.

The original A apply checkpoint and the replacement-head B apply checkpoint use separate suffixed refs ([A record](checkpoint-apply.json), [B record](checkpoint-B-apply.json)), allowing this final branch to contain one commit directly on 72f23939 without any force push. No main branch is pushed.
