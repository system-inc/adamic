# Stacked findAncestor: paired latent census

Task 448a51k. Report only. No compiler or adaptation change is committed.

| Side | Compiler pin |
| --- | --- |
| Parent and apply | `1060360b7b0d15dd6f0c0e8ec7bf0254f9fff000` |
| Head | `63c11488f6eed25e9070c80e392a53cd3309c43d` |

The remote optimizer/find-ancestor-result-stacked tip was checked before setup and remained 63c11488. Its parent is road (B), 1060360b, on b5c0aa64 and main 3fbe3a2a. Both internal/lower and the rest of each compiler are built from their own pin. The pinned checker is shared unchanged. [Input pins and carried-directory hashes](input-pins.json) and [binary provenance](provenance.json) retain the checks.

One scratch checkout of the parent received both complete directories stage3/adapt/99-correctness-core-or-predicate and stage3/adapt/99-correctness-core-multimap from eeb87728. Their bytes were checked against that commit before apply. Apply succeeded, exit 0, on TypeScript 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8. Both sides use this single adapted tree. [Patch set](patch-set.md), [apply log](apply.log), and [source hashes](source-hashes.json).

The entry first-error comparison reaches 81 files from src/tsc/tsc.ts, with zero checker diagnostics on each side. All census runs set LATENT_ASSERT_NO_OUTPUT=1 and exercise both no-output guards. Full-file runs use LATENT_FULL=1 and LATENT_ONLY_FILE set to the selected adapted file, with src/compiler as the loader root. Their historical raw measurement label says checker-rejected, but their checker_rejected field is false and diagnostics are empty. No TypeScript program is compiled to native output or run.

The method follows efb78625 stage3/census/find-ancestor-5a86e3b3/REPORT.md: own-pin guarded entry overlays, access-chain pairing, and separate focused full-unit inventories for the six nested checker callers. The specifically named core-sites-af42d3dc report was absent from the fetched history and the available typescript/core-sites tip; the available core-sites-4186581e runner was read, and the exact full-mode flags requested here were used.

## Entry counts

| View | Cleared | New findings | Changed reasons | New sites | Unit first-stop changes | Regressions |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Raw | 33 | 7 | 0 | 7 | 15 | 5 |
| Distinct outermost access chains | 33 | 7 | 0 | 7 | 15 | 5 |

Cleared means the old finding disappeared with no remaining finding at that source identity. Changed reasons retain another finding there. New findings include replacement reasons, so new and changed are not disjoint. Raw counts use exact raw-position keys, not event multiplicity. Distinct counts use the stock TypeScript 6.0.3 outermost access chain. A different first stop in an already-blocked unit is reported separately from regressions. The pair tool flags newly stopped units, newly finding refusal-scan sites and new panics. These observations do not prove native correctness.

[Raw parent stream](parent-entry-run.jsonl), [raw head stream](head-entry-run.jsonl), [paired JSON](entry-paired.json), and [paired ledger](entry-paired.md).

Full-file evidence: [utilitiesPublic parent](parent-full-utilitiesPublic.jsonl), [utilitiesPublic head](head-full-utilitiesPublic.jsonl), [utilitiesPublic paired JSON](utilitiesPublic-paired.json), [core parent](parent-full-core.jsonl), [core head](head-full-core.jsonl), [core paired JSON](core-paired.json), and [complete full-site ledger](full-sites.json).

## Answers to the four questions

1. Yes. In both the full-file and focused probes, the attempted findAncestor overload/implementation units at 787 and 788 move from the prologue Refused at utilitiesPublic.ts:786:1, citing return undefined at 792:13, to no recorded finding. Exact old reason: `overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:792:13 isn't the value its predicate checked)`. New first stop: none. The generic declaration at 786 is deferred on both sides.

2. No, the old helper-arrow stops are not reproduced. Neither full core stream records a finding in orPredicate2 or orPredicate3; their generic declarations remain deferred. In the focused head callers, the former checked-predicate-result ArrowFunction reason is absent at core.ts:2463:17 and 2467:17, and the four formerly blocked call sites have no local failed boundary or callee-helper finding. Parent probes at 4770, 4781 and 50077 stop earlier at findAncestor overload validation, so absence of a parent helper finding does not prove that helper body lowered. The complete core streams are byte-for-byte identical.

3. The exact caller coordinates and old/new stops are in the six-row table below. At 4751 the binder.ts:2258:13 source-slot certificate stop remains. At 48437 the checker.ts:48398:31 AliasDeclarationNode prologue stop remains. Calls at 4770, 4781 and 50077 lose the findAncestor return-at-795 refusal; 50086 loses its enclosing reading-result stop at checker.ts:50085:22. Those four head sites have no failed boundary containing the call. Earlier whole-unit stops still exist, as the focused table states; no whole-unit success is claimed.

Focused continuation also records later failures, distinct from the requested call sites: checked predicate overload target LiteralImportTypeNode at utilities.ts:2722:1 (boundary checker.ts:4787:18), the inferred-ArrowFunction predicate-result reason at checker.ts:4833:70, and a value of type HasJSDoc | undefined at checker.ts:50092:27. The first is present on both sides; the latter two are exposed on the head. Neither is a new stop in orPredicate2 or orPredicate3. Earlier whole-unit stops remain unchanged.

4. No. There are five newly finding refusal-scan sites, zero newly stopped units, and zero new panics. Entry regression entries: 5 distinct, 5 raw. The full-file regression and panic table below gives both additional comparisons. Changed first stops in already-blocked units are listed separately.

## Every unit whose first stop changes

### entry

| Unit | Location | Parent first stop | Head first stop | Interpretation |
| --- | --- | --- | --- | --- |
| transformES2016 | compiler/transformers/es2016.ts:21:1 | Refused: overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor at compiler/visitorPublic.ts:587:119 | Refused: overload 1 of visitEachChild parameter node cannot be served by implementation parameter node at compiler/visitorPublic.ts:585:48 | changed stop in blocked unit |
| transformES2019 | compiler/transformers/es2019.ts:17:1 | Refused: overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor at compiler/visitorPublic.ts:587:119 | Refused: overload 1 of visitEachChild parameter node cannot be served by implementation parameter node at compiler/visitorPublic.ts:585:48 | changed stop in blocked unit |
| transformES2020 | compiler/transformers/es2020.ts:42:1 | Refused: overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor at compiler/visitorPublic.ts:587:119 | Refused: overload 1 of visitEachChild parameter node cannot be served by implementation parameter node at compiler/visitorPublic.ts:585:48 | changed stop in blocked unit |
| transformES2021 | compiler/transformers/es2021.ts:25:1 | Refused: overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor at compiler/visitorPublic.ts:587:119 | Refused: overload 1 of visitEachChild parameter node cannot be served by implementation parameter node at compiler/visitorPublic.ts:585:48 | changed stop in blocked unit |
| transformLegacyDecorators | compiler/transformers/legacyDecorators.ts:88:1 | Refused: overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor at compiler/visitorPublic.ts:587:119 | Refused: overload 1 of visitEachChild parameter node cannot be served by implementation parameter node at compiler/visitorPublic.ts:585:48 | changed stop in blocked unit |
| isJSDocTypeExpressionOrChild | compiler/utilities.ts:1291:1 | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | finding-free | cleared first stop |
| getContainingFunction | compiler/utilities.ts:3125:1 | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | NotYet: writing a checked object field without its source-slot type certificate at compiler/binder.ts:1034:17 | changed stop in blocked unit |
| getContainingFunctionDeclaration | compiler/utilities.ts:3130:1 | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | NotYet: writing a checked object field without its source-slot type certificate at compiler/binder.ts:1034:17 | changed stop in blocked unit |
| getContainingClass | compiler/utilities.ts:3135:1 | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | finding-free | cleared first stop |
| getContainingFunctionOrClassStaticBlock | compiler/utilities.ts:3150:1 | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | NotYet: writing a checked object field without its source-slot type certificate at compiler/binder.ts:1034:17 | changed stop in blocked unit |
| getJSDocRoot | compiler/utilities.ts:4823:1 | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | finding-free | cleared first stop |
| isSideEffectImport | compiler/utilities.ts:12088:1 | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | finding-free | cleared first stop |
| findAncestor | compiler/utilitiesPublic.ts:787:1 | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:792:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | finding-free | cleared first stop |
| findAncestor | compiler/utilitiesPublic.ts:788:1 | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:792:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | finding-free | cleared first stop |
| isPartOfTypeOnlyImportOrExportDeclaration | compiler/utilitiesPublic.ts:1566:1 | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | NotYet: checked predicate overload target ExportDeclaration & { readonly isTypeOnly: true; readonly moduleSpecifier: Expression; } at compiler/utilitiesPublic.ts:1562:1 | changed stop in blocked unit |

### utilitiesPublic

| Unit | Location | Parent first stop | Head first stop | Interpretation |
| --- | --- | --- | --- | --- |
| findAncestor | compiler/utilitiesPublic.ts:787:1 | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:792:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | finding-free | cleared first stop |
| findAncestor | compiler/utilitiesPublic.ts:788:1 | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:792:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | finding-free | cleared first stop |
| isPartOfTypeOnlyImportOrExportDeclaration | compiler/utilitiesPublic.ts:1566:1 | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | NotYet: checked predicate overload target ExportDeclaration & { readonly isTypeOnly: true; readonly moduleSpecifier: Expression; } at compiler/utilitiesPublic.ts:1562:1 | changed stop in blocked unit |

### core

| Unit | Location | Parent first stop | Head first stop | Interpretation |
| --- | --- | --- | --- | --- |
| None | | | | |

## Entry findAncestor status

The generic overload declaration at 786 is deferred; its absence of a direct lowering stop is not a successful body probe. The non-generic overload and implementation units at 787 and 788 exercise the prologue.

| Side | Unit | Status | Instantiated | First lowering stop |
| --- | --- | --- | --- | --- |
| parent | compiler/utilitiesPublic.ts:786:1 | deferred_generic_instantiation | True | finding-free |
| parent | compiler/utilitiesPublic.ts:787:1 | attempted | False | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:792:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 |
| parent | compiler/utilitiesPublic.ts:788:1 | attempted | False | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:792:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 |
| head | compiler/utilitiesPublic.ts:786:1 | deferred_generic_instantiation | True | finding-free |
| head | compiler/utilitiesPublic.ts:787:1 | attempted | False | finding-free |
| head | compiler/utilitiesPublic.ts:788:1 | attempted | False | finding-free |

## Six requested checker callers, parent against head

Focused full-mode inventories use the same replay selector as the prior report. They retain the complete entry graph and admit no usable IR. A failed statement may contain multiple calls; its boundary is not evidence of independent execution of every call. These observations are separate from the entry first-error counts.

| Upstream line | Adapted call | Enclosing declaration | Parent boundary stop | Head boundary stop |
| ---: | --- | --- | --- | --- |
| 4748 | compiler/checker.ts:4751:17 | resolveExternalModule at compiler/checker.ts:4733:5 | enclosing statement: NotYet: writing a checked object field without its source-slot type certificate at compiler/binder.ts:2258:13 | enclosing statement: NotYet: writing a checked object field without its source-slot type certificate at compiler/binder.ts:2258:13 |
| 4767 | compiler/checker.ts:4770:21 | resolveExternalModule at compiler/checker.ts:4733:5 | enclosing statement: Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | no boundary contains this site: no finding recorded; status attempted |
| 4778 | compiler/checker.ts:4781:21 | resolveExternalModule at compiler/checker.ts:4733:5 | enclosing statement: Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | no boundary contains this site: no finding recorded; status attempted |
| 48433 | compiler/checker.ts:48437:47 | checkAliasSymbol at compiler/checker.ts:48398:5 | unit prologue: NotYet: a value of type AliasDeclarationNode at compiler/checker.ts:48398:31 | unit prologue: NotYet: a value of type AliasDeclarationNode at compiler/checker.ts:48398:31 |
| 50073 | compiler/checker.ts:50077:29 | getSymbolOfNameOrPropertyAccessExpression at compiler/checker.ts:49992:5 | enclosing statement: Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | no boundary contains this site: no finding recorded; status attempted |
| 50082 | compiler/checker.ts:50086:39 | getSymbolOfNameOrPropertyAccessExpression at compiler/checker.ts:49992:5 | enclosing statement: NotYet: reading result at compiler/checker.ts:50085:22 | no boundary contains this site: no finding recorded; status attempted |

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

Parent: enclosing statement: Refused: overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1.

Head: no boundary contains this site: no finding recorded; status attempted.

Focused unit first lowering stop (parent): NotYet: writing a checked object field without its source-slot type certificate at compiler/binder.ts:2258:13.

Focused unit first lowering stop (head): NotYet: writing a checked object field without its source-slot type certificate at compiler/binder.ts:2258:13.

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

Parent: enclosing statement: Refused: overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1.

Head: no boundary contains this site: no finding recorded; status attempted.

Focused unit first lowering stop (parent): NotYet: writing a checked object field without its source-slot type certificate at compiler/binder.ts:2258:13.

Focused unit first lowering stop (head): NotYet: writing a checked object field without its source-slot type certificate at compiler/binder.ts:2258:13.

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

Parent: enclosing statement: Refused: overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1.

Head: no boundary contains this site: no finding recorded; status attempted.

Focused unit first lowering stop (parent): NotYet: checked predicate membership over an open tag domain at compiler/checker.ts:50003:18.

Focused unit first lowering stop (head): NotYet: checked predicate membership over an open tag domain at compiler/checker.ts:50003:18.

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

Head: no boundary contains this site: no finding recorded; status attempted.

Focused unit first lowering stop (parent): NotYet: checked predicate membership over an open tag domain at compiler/checker.ts:50003:18.

Focused unit first lowering stop (head): NotYet: checked predicate membership over an open tag domain at compiler/checker.ts:50003:18.

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

## utilitiesPublic.ts findAncestor itself, parent against head

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

Parent focused observation: unit prologue: Refused: overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:792:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1.

Head focused observation: no boundary contains this site: no finding recorded; status attempted.

parent:
Source `compiler/utilitiesPublic.ts:784` through line 788:

```typescript
 * At that point findAncestor returns undefined.
 */
export function findAncestor<T extends Node>(node: Node | undefined, callback: (element: Node) => element is T): T | undefined;
export function findAncestor(node: Node | undefined, callback: (element: Node) => boolean | "quit"): Node | undefined;
export function findAncestor(node: Node | undefined, callback: (element: Node) => boolean | "quit"): Node | undefined {
```


## Resource index

[Resource index](resources.md) includes all retained status/time pairs. The census run table below gives own child exits, wall time and peak RSS for every requested run.


## Full-file inventories

These are separate full-mode measurements on the compiler directory, selecting exactly one adapted file with LATENT_ONLY_FILE. Continuation Boundary records are retained in raw streams but are not counted as refusal sites here. A deferred generic body with no finding is unexamined, not proven to lower.

### findAncestor

```typescript
export function findAncestor<T extends Node>(node: Node | undefined, callback: (element: Node) => element is T): T | undefined;
```

```typescript
export function findAncestor(node: Node | undefined, callback: (element: Node) => boolean | "quit"): Node | undefined;
```

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

| Side | Unit | Status | Instantiated |
| --- | --- | --- | --- |
| parent | compiler/utilitiesPublic.ts:786:1 | deferred_generic_instantiation | True |
| parent | compiler/utilitiesPublic.ts:787:1 | attempted | False |
| parent | compiler/utilitiesPublic.ts:788:1 | attempted | False |
| head | compiler/utilitiesPublic.ts:786:1 | deferred_generic_instantiation | True |
| head | compiler/utilitiesPublic.ts:787:1 | attempted | False |
| head | compiler/utilitiesPublic.ts:788:1 | attempted | False |

| Side | Every recorded site | Phase | Recording units |
| --- | --- | --- | --- |
| parent | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:792:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | lowering | compiler/utilitiesPublic.ts:787:1, compiler/utilitiesPublic.ts:788:1 |
| parent | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | lowering | compiler/utilitiesPublic.ts:1566:1 |
| head | No site finding recorded | | |

### Every changed site in utilitiesPublic.ts

| Category | Old | New |
| --- | --- | --- |
| lowered | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:792:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | none |
| lowered | Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) at compiler/utilitiesPublic.ts:786:1 | none |
| new_findings | none | NotYet: checked predicate overload target ExportDeclaration & { readonly isTypeOnly: true; readonly moduleSpecifier: Expression; } at compiler/utilitiesPublic.ts:1562:1 |

### orPredicate2

```typescript
export function orPredicate2<P, R1 extends P, R2 extends P>(f1: (p1: P) => p1 is R1, f2: (p2: P) => p2 is R2): (p: P) => p is R1 | R2 {
    return p => f1(p) || f2(p);
}
```

| Side | Unit | Status | Instantiated |
| --- | --- | --- | --- |
| parent | compiler/core.ts:2462:1 | deferred_generic_instantiation | False |
| head | compiler/core.ts:2462:1 | deferred_generic_instantiation | False |

| Side | Every recorded site | Phase | Recording units |
| --- | --- | --- | --- |
| parent | No site finding recorded | | |
| head | No site finding recorded | | |

### orPredicate3

```typescript
export function orPredicate3<P, R1 extends P, R2 extends P, R3 extends P>(f1: (p1: P) => p1 is R1, f2: (p2: P) => p2 is R2, f3: (p3: P) => p3 is R3): (p: P) => p is R1 | R2 | R3 {
    return p => f1(p) || f2(p) || f3(p);
}
```

| Side | Unit | Status | Instantiated |
| --- | --- | --- | --- |
| parent | compiler/core.ts:2466:1 | deferred_generic_instantiation | False |
| head | compiler/core.ts:2466:1 | deferred_generic_instantiation | False |

| Side | Every recorded site | Phase | Recording units |
| --- | --- | --- | --- |
| parent | No site finding recorded | | |
| head | No site finding recorded | | |

### or

```typescript
export function or<T extends unknown[], U>(...fs: [(...args: T) => U, ...((...args: T) => U)[]]): (...args: T) => U;
```

```typescript
export function or<T extends unknown[], U>(...fs: ((...args: T) => U)[]): (...args: T) => U | undefined;
```

```typescript
export function or<T extends unknown[], U>(...fs: ((...args: T) => U)[]): (...args: T) => U | undefined {
    return (...args) => {
        let lastResult: U | undefined;
        for (const f of fs) {
            lastResult = f(...args);
            if (lastResult) {
                return lastResult;
            }
        }
        return lastResult;
    };
}
```

| Side | Unit | Status | Instantiated |
| --- | --- | --- | --- |
| parent | compiler/core.ts:2470:1 | deferred_generic_instantiation | False |
| parent | compiler/core.ts:2472:1 | deferred_generic_instantiation | False |
| parent | compiler/core.ts:2474:1 | deferred_generic_instantiation | False |
| head | compiler/core.ts:2470:1 | deferred_generic_instantiation | False |
| head | compiler/core.ts:2472:1 | deferred_generic_instantiation | False |
| head | compiler/core.ts:2474:1 | deferred_generic_instantiation | False |

| Side | Every recorded site | Phase | Recording units |
| --- | --- | --- | --- |
| parent | No site finding recorded | | |
| head | No site finding recorded | | |

### Every changed site in core.ts

| Category | Old | New |
| --- | --- | --- |
| None | | |

### Helper sites across entry and focused contexts

| Context | Side | Helper | Site and exact reason | Phase | Recording unit |
| --- | --- | --- | --- | --- | --- |
| entry | parent | or | NotYet: overload 1 of or with a non-array rest parameter at compiler/core.ts:2470:1 | lowering | compiler/utilities.ts:10054:1 |
| entry | head | or | NotYet: overload 1 of or with a non-array rest parameter at compiler/core.ts:2470:1 | lowering | compiler/utilities.ts:10054:1 |


## Regressions and panics

| Scope | Regressions, raw | Regressions, distinct | Parent panics | Head panics |
| --- | ---: | ---: | ---: | ---: |
| entry | 5 | 5 | 0 | 0 |
| utilitiesPublic | 0 | 0 | 0 | 0 |
| core | 0 | 0 | 0 | 0 |

- entry previously_finding_free_scan_site: Refused: a function taking Node seen as one taking NonNullable<BigIntLiteral | Identifier | NumericLiteral | PrivateIdentifier | StringLiteralLike> (tsc relates a method's parameters both ways), so it can be handed what it can't take at compiler/transformers/esDecorators.ts:1931:42

- entry previously_finding_free_scan_site: Refused: a function taking Node seen as one taking NonNullable<BigIntLiteral | Identifier | NoSubstitutionTemplateLiteral | NumericLiteral | PrivateIdentifier | StringLiteral> (tsc relates a method's parameters both ways), so it can be handed what it can't take at compiler/transformers/esDecorators.ts:1954:32

- entry previously_finding_free_scan_site: Refused: a function taking (hint: EmitHint, node: Node) => void seen as one taking (hint: EmitHint, node: Node) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take at compiler/transformers/es2017.ts:163:26

- entry previously_finding_free_scan_site: Refused: a function taking (hint: EmitHint, node: Node) => void seen as one taking (hint: EmitHint, node: Node) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take at compiler/transformers/es2018.ts:168:26

- entry previously_finding_free_scan_site: Refused: a function taking Node seen as one taking NonNullable<MethodDeclaration | ShorthandPropertyAssignment | AccessorDeclaration> (tsc relates a method's parameters both ways), so it can be handed what it can't take at compiler/transformers/es2018.ts:499:40

## Every census run: time and peak memory

| Run | Own child exit | Wall seconds | Peak RSS MiB |
| --- | ---: | ---: | ---: |
| parent-entry-run | 0 | 3714.281 | 5335.9 |
| head-entry-run | 0 | 3730.644 | 5584.7 |
| parent-full-utilitiesPublic | 0 | 167.020 | 1241.7 |
| head-full-utilitiesPublic | 0 | 166.404 | 1203.7 |
| parent-full-core | 0 | 27.033 | 803.7 |
| head-full-core | 0 | 27.027 | 745.5 |
| parent-focus-1 | 0 | 16.087 | 744.2 |
| parent-focus-2 | 0 | 9.008 | 529.5 |
| parent-focus-3 | 0 | 15.119 | 742.3 |
| parent-focus-4 | 0 | 9.015 | 518.9 |
| head-focus-1 | 0 | 23.033 | 808.1 |
| head-focus-2 | 0 | 8.007 | 530.1 |
| head-focus-3 | 0 | 23.026 | 821.4 |
| head-focus-4 | 0 | 10.022 | 568.5 |

RSS is Linux child getrusage peak RSS, converted from KiB to MiB. Wall times include up to one second of wrapper polling. Entry runs overlap and use identical GOMEMLIMIT=5GiB and GOGC=70 settings; these are shared-host observations, not benchmarks. Each run retains its exact command, own exit, unit timings and log. The wrapper terminates a process group after 30 minutes without output or unit progress, and cleans up remaining group members after exit.

## Validation and checkpoints

Setup, npm ci, apply and all six guarded binary builds completed successfully. All 15 existing paired-census tests passed. Source hashes, equal measurement headers, equal unit identities, raw finding totals, new-finding counts, first-stop changes, compiler pins and binary hashes were independently checked by finish.py. Go evidence uses .go.txt. No measured source was removed or changed.

The initial controller stopped before any census when the entry binary output path collided with its overlay directory. The already built binaries were moved from the directory to explicit -entry-bin paths, and the runner was corrected. This is retained in controller-initial.log; no partial census from that attempt contributes to counts.

A checkpoint was pushed after successful apply and after every census run, to separate typescript/find-ancestor-stacked-census-checkpoint-* refs. Each checkpoint has a JSON receipt and push log beside this report. This allows the final branch to contain one report-only commit directly on 1060360b without force-pushing. No main ref is pushed.
