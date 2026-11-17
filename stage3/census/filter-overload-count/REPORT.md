# Filter result-use count

Task #ev14ch9. Report only. No adapter, compiler, lane, API-sanction or ledger change.
Base: `03c140ec28a249b8b0f13eb24e34fc3d3df05266`.
TypeScript 6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`.

Fresh detached worktree; recursive shallow submodules initialized; stage3/api npm ci;
unchanged stage3/apply.sh exited 0. Scratch contains all caches and temporary products.
The completed adapted sources are hashed in evidence/source-sha256.json.

## Coverage and method

Reuses the checker traversal from `typescript/filter-overload-sites` at
`05a2ba742a413765fd4ca636c5dbcdae1f07e948`; supplement-original.cjs preserves it.
The testRunner/typings exclusions are removed. Both independent traversals resolve
157 calls to declarations named filter owned by src/compiler/core.ts. 713 src .ts
roots include compiler, services, server, harness, testRunner and the other src trees.
Outside src, AST traversal examines 20,523 .ts/.tsx/.mts/.cts files; 40 contain a
syntactic filter call. Separate checker programs for those files resolve zero calls
to core filter. Separate programs prevent test-fixture globals merging. Outside-src
programs use noLib; no claim of a diagnostic-free production program is made.
All selection is by AST plus TypeScript getResolvedSignature, never text matching.

Classes overlap: a writable slot can also be passed to a writable parameter and
identity-compared. Read-only means none of these non-read classes. Writable slot
includes an inferred mutable local, an explicit writable assertion, an assignment,
a property initializer, or a writable return type. An inferred slot can become
readonly automatically after changing an overload; its presence is not itself a
type-check failure. The write column includes writable parameters, including a
subsequent call to filter's mutable-input overload, without claiming the callee
actually mutates. There are zero direct push/splice/sort/index/length writes in the
recorded local result uses. This classifies result uses, not writes through the
original argument, arbitrary reentrant aliases or whole-program escape lifetimes.

## Count

| Overload | Result | Calls | Reads only | Writes / writable parameter | Writable slot / return | Identity vs argument |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| 1 | `U[]` | 5 | 2 | 0 | 3 | 0 |
| 2 | `T[]` | 60 | 16 | 7 | 41 | 2 |
| 5 | `U[] or undefined` | 11 | 4 | 2 | 5 | 0 |
| 6 | `T[] or undefined` | 13 | 4 | 2 | 8 | 0 |

Narrowing overloads 1 + 5: 16 calls, 6 reads-only, 2 writable-parameter sites,
8 writable slots/returns, no argument-identity comparison. Other overloads are
already readonly: 3 = 14 calls, 4 = 41, 7 = 9, 8 = 4.

The implementation returns readonly T[] or undefined. It returns its argument
when no predicate call rejects an element; the empty-array and undefined paths
also return the argument. No copying is proposed. The separate readonly boolean
call utilities.ts:4686:31 compares its result with jsDoc.tags; it is overload 8,
outside this writable-result table. The two additional identities below are real
writable-result callers and must also survive any implementation change.

filter is @internal. The public baseline tests/baselines/reference/api/typescript.d.ts
has no filter declaration. Thus filter itself does not trigger the public-line gate.

## Every non-read call

For each site, **may return argument as-is: yes** is the conservative static
answer: these are ordinary array/optional-array arguments, not contracts forcing
one element to fail the predicate. Empty inputs or all-matching elements reach
return array. This is not an assertion that a particular lane invocation exercised
that path. The earlier runtime census is not reused as current runtime evidence.
No site is cleared by assuming that filtering always allocates.

### src/compiler/utilities.ts:642:12 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(symbol.declarations || emptyArray, d => d.kind === kind)
```

writable type assertion : T[].

### src/compiler/utilities.ts:2773:12 — overload 6

Classes: store. May return argument as-is: **yes**.

```ts
filter(commentRanges, comment =>
        comment.end <= node.end && // Due to parse errors sometime empty parameter may get comments assigned to it that end up not in parameter range
        text.charCodeAt(comment.pos + 1) === CharacterCodes.asterisk &&
        text.charCodeAt(comment.pos + 2) === CharacterCodes.asterisk &&
        text.charCodeAt(comment.pos + 3) !== CharacterCodes.slash)
```

return : CommentRange[] | undefined.

### src/compiler/utilities.ts:7029:31 — overload 6

Classes: store. May return argument as-is: **yes**.

```ts
filter(getLeadingCommentRanges(text, node.pos), isPinnedCommentLocal)
```

assignment : CommentRange[] | undefined.

### src/compiler/factory/utilities.ts:761:12 — overload 5

Classes: store. May return argument as-is: **yes**.

```ts
filter(getEmitHelpers(sourceFile), helper => !helper.scoped)
```

return : UnscopedEmitHelper[] | undefined.

### src/compiler/commandLineParser.ts:2550:16 — overload 1

Classes: store. May return argument as-is: **yes**.

```ts
filter(elements.map(element => convertPropertyValueToJson(element, elementOption)), v => v !== undefined)
```

return : AdamicJsonRecoveryValue[] | undefined.

### src/compiler/commandLineParser.ts:3909:12 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(map<TValues[number], CompilerOptionsValue>(values, (v, index) => convertJsonOption(option.element, v, basePath, errors, propertyAssignment, valueExpression?.elements[index], sourceFile)), v => option.listPreserveFalsyValues ? true : !!v)
```

return : any[].

### src/compiler/checker.ts:6586:39 — overload 5

Classes: store. May return argument as-is: **yes**.

```ts
filter(symbol.declarations, isClassLike)
```

variable classDeclarations : ClassLikeDeclaration[] | undefined.

### src/compiler/checker.ts:6604:27 — overload 6

Classes: store. May return argument as-is: **yes**.

```ts
filter(symbol.declarations, isDeclKind)
```

variable decls : Declaration[] | undefined.

### src/compiler/checker.ts:7627:47 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(indexInfo.components, e => {
                            // skip late bound props that contribute to the index signature - they'll be created by property creation anyway
                            return !hasLateBindableName(e);
                        })
```

variable newComponents : ElementWithComputedPropertyName[].

### src/compiler/checker.ts:9417:28 — overload 1

Classes: store. May return argument as-is: **yes**.

```ts
filter(map(statement.declarationList.declarations, getNameOfDeclaration), isIdentifierAndNotUndefined)
```

return : Identifier[].

### src/compiler/checker.ts:9419:24 — overload 1

Classes: store. May return argument as-is: **yes**.

```ts
filter([getNameOfDeclaration(statement as DeclarationStatement)], isIdentifierAndNotUndefined)
```

return : Identifier[].

### src/compiler/checker.ts:9433:43 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(statements, s => !!(getEffectiveModifierFlags(s) & ModifierFlags.Export))
```

variable excessExports : Statement[].

### src/compiler/checker.ts:9474:33 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(statements, d => isExportDeclaration(d) && !d.moduleSpecifier && !!d.exportClause && isNamedExports(d.exportClause))
```

writable type assertion : ExportDeclaration[].

### src/compiler/checker.ts:9476:40 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(statements, d => !isExportDeclaration(d) || !!d.moduleSpecifier || !d.exportClause)
```

variable nonExports : Statement[].

### src/compiler/checker.ts:9488:35 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(statements, d => isExportDeclaration(d) && !!d.moduleSpecifier && !!d.exportClause && isNamedExports(d.exportClause))
```

writable type assertion : ExportDeclaration[].

### src/compiler/checker.ts:9521:55 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(indices, i => nodeHasName(statements[i]!, name))
```

variable associatedIndices : number[].

### src/compiler/checker.ts:10025:24 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(exports, m => isNamespaceMember(m) && isIdentifierText(m.escapedName as string, ScriptTarget.ESNext))
```

return : Symbol[].

### src/compiler/checker.ts:10094:37 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(getPropertiesOfType(getTypeOfSymbol(symbol)), p => !!(p.flags & SymbolFlags.EnumMember))
```

variable memberProps : Symbol[].

### src/compiler/checker.ts:10162:35 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(getPropertiesOfType(type), isNamespaceMember)
```

variable props : Symbol[].

### src/compiler/checker.ts:10322:43 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(symbolProps, s => !isHashPrivate(s))
```

variable publicSymbolProps : Symbol[].

### src/compiler/checker.ts:16748:77 — overload 6

Classes: write. May return argument as-is: **yes**.

```ts
filter(siblingSymbols, s => s !== indexSymbol)
```

parameter : Symbol[] | undefined.

writable parameter, src/compiler/checker.ts:16748:40:

```ts
concatenate(computedPropertySymbols, filter(siblingSymbols, s => s !== indexSymbol))
```

### src/compiler/checker.ts:18223:27 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(types, isPatternLiteralType)
```

writable type assertion : (TemplateLiteralType | StringMappingType)[].

### src/compiler/checker.ts:18565:26 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(types, t => !!(t.flags & TypeFlags.StringLiteral))
```

variable literals : Type[].

### src/compiler/checker.ts:19962:87 — overload 6

Classes: store. May return argument as-is: **yes**.

```ts
filter(allOuterTypeParameters, tp => isTypeParameterPossiblyReferenced(tp, node))
```

variable outerTypeParameters : TypeParameter[] | undefined.

### src/compiler/checker.ts:20763:17 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(typeParameters, tp => some(allDeclarations, d => isTypeParameterPossiblyReferenced(tp, d)))
```

assignment : TypeParameter[] | undefined.

### src/compiler/checker.ts:26375:28 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(context.inferences, hasInferenceCandidates)
```

variable inferences : InferenceInfo[].

### src/compiler/checker.ts:27086:34 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(sources, t => !contains(matchedSources, t))
```

contextual writable slot : Type[].

### src/compiler/checker.ts:27087:34 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(targets, t => !contains(matchedTargets, t))
```

contextual writable slot : Type[].

### src/compiler/checker.ts:27577:36 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(candidates, isObjectOrArrayLiteralType)
```

variable objectLiterals : Type[].

### src/compiler/checker.ts:27580:36 — overload 2

Classes: write. May return argument as-is: **yes**.

```ts
filter(candidates, t => !isObjectOrArrayLiteralType(t))
```

parameter : Type[].

writable parameter, src/compiler/checker.ts:27580:24:

```ts
concatenate(filter(candidates, t => !isObjectOrArrayLiteralType(t)), [literalsType])
```

### src/compiler/checker.ts:28532:30 — overload 2

Classes: store, write, identity. May return argument as-is: **yes**.

```ts
filter(types, f)
```

variable filtered : Type[].

identity, src/compiler/checker.ts:28533:17:

```ts
filtered === types
```

writable parameter, src/compiler/checker.ts:28555:20:

```ts
getUnionTypeFromSortedList(filtered, (type as UnionType).objectFlags & (ObjectFlags.PrimitiveUnion | ObjectFlags.ContainsIntersections), /*aliasSymbol*/ undefined, /*aliasTypeArguments*/ undefined, newOrigin)
```

### src/compiler/checker.ts:28545:40 — overload 2

Classes: store, write. May return argument as-is: **yes**.

```ts
filter(originTypes, t => !!(t.flags & TypeFlags.Union) || f(t))
```

variable originFiltered : Type[].

writable parameter, src/compiler/checker.ts:28550:33:

```ts
createOriginUnionOrIntersectionType(TypeFlags.Union, originFiltered)
```

### src/compiler/checker.ts:35256:25 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(props, prop => isValidPropertyAccessForCompletions(parent, containingType, prop))
```

assignment : Symbol[].

### src/compiler/checker.ts:45744:39 — overload 2

Classes: store, identity. May return argument as-is: **yes**.

```ts
filter(arrayTypes, t => !(t.flags & TypeFlags.StringLike))
```

variable filteredTypes : Type[].

identity, src/compiler/checker.ts:45745:21:

```ts
filteredTypes !== arrayTypes
```

### src/compiler/checker.ts:47509:16 — overload 5

Classes: store. May return argument as-is: **yes**.

```ts
filter(symbol.declarations, (d: Declaration): d is ClassDeclaration | InterfaceDeclaration => d.kind === SyntaxKind.ClassDeclaration || d.kind === SyntaxKind.InterfaceDeclaration)
```

return : (InterfaceDeclaration | ClassDeclaration)[] | undefined.

### src/compiler/checker.ts:51301:28 — overload 6

Classes: store. May return argument as-is: **yes**.

```ts
filter(getExportSymbolOfValueSymbolIfExported(symbol).declarations, declaration => {
                        switch (declaration.kind) {
                            case SyntaxKind.VariableDeclaration:
                            case SyntaxKind.Parameter:
                            case SyntaxKind.BindingElement:
                            case SyntaxKind.PropertyDeclaration:
                            case SyntaxKind.PropertyAssignment:
                            case SyntaxKind.ShorthandPropertyAssignment:
                            case SyntaxKind.EnumMember:
                            case SyntaxKind.ObjectLiteralExpression:
                            case SyntaxKind.FunctionDeclaration:
                            case SyntaxKind.FunctionExpression:
                            case SyntaxKind.ArrowFunction:
                            case SyntaxKind.ClassDeclaration:
                            case SyntaxKind.ClassExpression:
                            case SyntaxKind.EnumDeclaration:
                            case SyntaxKind.MethodDeclaration:
                            case SyntaxKind.GetAccessor:
                            case SyntaxKind.SetAccessor:
                            case SyntaxKind.ModuleDeclaration:
                                return true;
                        }
                        return false;
                    })
```

return : Declaration[] | undefined.

### src/compiler/checker.ts:51492:55 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(info.components, e => {
                                    // skip late bound props that contribute to the index signature - they'll be preserved via other means
                                    return !hasLateBindableName(e);
                                })
```

variable newComponents : ElementWithComputedPropertyName[].

### src/compiler/transformers/declarations.ts:430:40 — overload 6

Classes: store. May return argument as-is: **yes**.

```ts
filter(symbol.declarations, d => getSourceFileOfNode(d) !== containingFile)
```

variable augmentingDeclarations : Declaration[] | undefined.

### src/compiler/emitter.ts:688:15 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(fileNames, file => !(options.noEmitForJsFiles && fileExtensionIsOneOf(file, supportedJSExtensionsFlat)) && !isDeclarationFileName(file))
```

return : string[].

### src/compiler/program.ts:2150:30 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(files, file => sourceFileMayBeEmitted(file, program))
```

variable emittedFiles : SourceFile[].

### src/compiler/program.ts:2910:35 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(bindDiagnostics, d => plainJSErrors.has(d.code))
```

assignment : DiagnosticWithLocation[].

### src/compiler/program.ts:2911:36 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(checkDiagnostics, d => plainJSErrors.has(d.code))
```

assignment : Diagnostic[].

### src/compiler/executeCommandLine.ts:177:25 — overload 2

Classes: store, write. May return argument as-is: **yes**.

```ts
filter(optionDeclarations.concat(tscBuildOption), option => option.showInHelp !== false)
```

variable helpOptions : CommandLineOption[].

writable parameter, src/compiler/executeCommandLine.ts:181:9:

```ts
filter(helpOptions, v => !!v.showInSimplifiedHelpView)
```

### src/compiler/executeCommandLine.ts:181:9 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(helpOptions, v => !!v.showInSimplifiedHelpView)
```

return : CommandLineOption[] | SortedReadonlyArray<CommandLineOption>.

### src/services/transpile.ts:243:9 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(optionDeclarations, o => typeof o.type === "object" && !forEachEntry(o.type, v => typeof v !== "number"))
```

writable type assertion : CommandLineOptionOfCustomType[].

### src/services/services.ts:780:81 — overload 5

Classes: write. May return argument as-is: **yes**.

```ts
filter(this.declarations, isGetAccessor)
```

parameter : Declaration[] | undefined.

writable parameter, src/services/services.ts:780:54:

```ts
getJsDocTagsOfDeclarations(filter(this.declarations, isGetAccessor), checker)
```

### src/services/services.ts:790:81 — overload 5

Classes: write. May return argument as-is: **yes**.

```ts
filter(this.declarations, isSetAccessor)
```

parameter : Declaration[] | undefined.

writable parameter, src/services/services.ts:790:54:

```ts
getJsDocTagsOfDeclarations(filter(this.declarations, isSetAccessor), checker)
```

### src/services/services.ts:3603:11 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(contextualType.types, t => !checker.isTypeInvalidDueToUnionDiscriminant(t, node.parent))
```

variable filteredTypes : Type[].

### src/services/callHierarchy.ts:482:19 — overload 5

Classes: store. May return argument as-is: **yes**.

```ts
filter(FindAllReferences.findReferenceOrRenameEntries(program, cancellationToken, program.getSourceFiles(), location, /*position*/ 0, { use: FindAllReferences.FindReferencesUse.References }, convertEntryToCallSite), isDefined)
```

variable calls : CallSite[] | undefined.

### src/services/codefixes/fixAddMissingMember.ts:773:12 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(targetProps, targetProp => isIdentifierText(targetProp.name, target, LanguageVariant.JSX) && !((targetProp.flags & SymbolFlags.Optional || getCheckFlags(targetProp) & CheckFlags.Partial) || seenNames.has(targetProp.escapedName)))
```

return : Symbol[].

### src/services/codefixes/fixAddMissingParam.ts:143:46 — overload 5

Classes: store. May return argument as-is: **yes**.

```ts
filter(type.symbol.declarations, isConvertibleSignatureDeclaration)
```

variable convertibleSignatureDeclarations : ConvertibleSignatureDeclaration[] | undefined.

### src/services/completions.ts:2332:39 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter((effectiveType as UnionType).types, type => checker.getSignaturesOfType(type, SignatureKind.Call).length > 0)
```

variable functionTypes : Type[].

### src/services/completions.ts:4532:40 — overload 2

Classes: write. May return argument as-is: **yes**.

```ts
filter(members, s => !existingMemberEscapedNames.has(s.escapedName))
```

parameter : Symbol[].

writable parameter, src/services/completions.ts:4532:19:

```ts
concatenate(symbols, filter(members, s => !existingMemberEscapedNames.has(s.escapedName)))
```

### src/services/completions.ts:4695:19 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(typeChecker.getTypeAtLocation(importAttributes).getApparentProperties(), attr => !existing.has(attr.escapedName))
```

assignment : Symbol[].

### src/services/completions.ts:5624:9 — overload 2

Classes: write. May return argument as-is: **yes**.

```ts
filter(
            contextualType.flags & TypeFlags.Union ?
                (contextualType as UnionType).types :
                [contextualType],
            t => !checker.getPromisedTypeOfPromise(t),
        )
```

parameter : Type[].

writable parameter, src/services/completions.ts:5623:43:

```ts
checker.getUnionType(
        filter(
            contextualType.flags & TypeFlags.Union ?
                (contextualType as UnionType).types :
                [contextualType],
            t => !checker.getPromisedTypeOfPromise(t),
        ),
    )
```

### src/services/completions.ts:5637:30 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(properties, hasDeclarationOtherThanSelf)
```

return : Symbol[].

### src/services/stringCompletions.ts:614:18 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(type.getApparentProperties(), prop => !(prop.valueDeclaration && isPrivateIdentifierClassElementDeclaration(prop.valueDeclaration)))
```

property : Symbol[].

### src/services/goToDefinition.ts:628:68 — overload 6

Classes: store, write. May return argument as-is: **yes**.

```ts
filter(symbol.declarations, declarationFilter)
```

variable filteredDeclarations : Declaration[] | undefined.

writable parameter, src/services/goToDefinition.ts:634:29:

```ts
filter(filteredDeclarations, d => !isExpandoDeclaration(d))
```

### src/services/goToDefinition.ts:634:29 — overload 6

Classes: store. May return argument as-is: **yes**.

```ts
filter(filteredDeclarations, d => !isExpandoDeclaration(d))
```

variable withoutExpandos : Declaration[] | undefined.

### src/server/project.ts:1659:16 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(include, i => !existing.includes(i))
```

return : string[].

### src/server/session.ts:2856:33 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
filter(projectItems, item => tryAddSeenItem(item) && !getMappedLocationForProject(documentSpanLocation(item), project))
```

variable unseenItems : NavigateToItem[].

### src/harness/fourslashImpl.ts:2394:23 — overload 2

Classes: store, write. May return argument as-is: **yes**.

```ts
ts.filter(this.getCompilerTestFiles(), f => !ts.endsWith(f.unitName, ".json"))
```

variable files : { content: string; unitName: string; }[].

writable parameter, src/harness/fourslashImpl.ts:2395:24:

```ts
this.getSyntacticDiagnosticBaselineText(files)
```

writable parameter, src/harness/fourslashImpl.ts:2398:15:

```ts
this.getSemanticDiagnosticBaselineText(files)
```

### src/harness/fourslashImpl.ts:3285:30 — overload 2

Classes: store. May return argument as-is: **yes**.

```ts
ts.filter(actual, f => kind === undefined ? true : f.kind === kind)
```

variable filterActual : ts.OutliningSpan[].

## Decision gate

The count does not establish an all-read overload 1 or 5. Road (a) needs truthful
local treatment of writable returns and writable parameter contracts. Road (b)
must retain required narrowed element types at callers. A no-emit virtual checker
comparison of those roads follows this count checkpoint; no stage3 adaptation has
been built. The existing generic narrowing/array-result refusal must also be
resolved before claiming a readonly declaration removes core.ts:261:1.

## Reproduction

With the unchanged completed adapted tree TREE:

```sh
node stage3/census/filter-overload-count/contexts.cjs TREE contexts.json
node stage3/census/filter-overload-count/supplement.cjs TREE calls.json
node stage3/census/filter-overload-count/outside.cjs TREE outside.json
```

Raw rows retain all read sites, result types, contexts, local symbol references,
and the exact declaration selection. No Loom submission has been made.

Count audit correction: overload-2 calls checker.ts:27086:34 and :27087:34
occupy writable tuple-result elements (`[Type[], Type[]]`). The final table includes
them as stores. This correction changes overload 2 from 18 reads / 39 stores to
16 reads / 41 stores; all call totals and narrowing-overload counts stay unchanged.
