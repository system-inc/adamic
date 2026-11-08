# Array copy census for tsc

Task 5nrs3q2. Report only. No compiler changes are committed.

Base origin/main: `0816e18511b5e1ef4383543ee81c526fc36acfaf`. Stock source: `050880ce59e30b356b686bd3144efe24f875ebc8` (TypeScript 6.0.3). Adapted product key: `0145bc21ed5011be8920401d7a7c2d375dd3938eaec6ff5be59df1303cb20d85`. Analysis API: TypeScript 6.0.3. Node: v24.19.0.

## Headline static counts

| Tree | Object in loop | Object outside loop | Mixed in loop | Mixed outside loop | Combined in loop | Combined outside loop | Unresolved in / out |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Adapted | 85 | 676 | 2 | 21 | 87 | 697 | 7 / 20 |
| Stock | 85 | 680 | 2 | 18 | 87 | 698 | 7 / 23 |

Explicitly named family subtotal (additional audited helpers excluded):

| Tree | In loop | Outside loop |
| --- | --- | --- |
| Adapted | 78 | 669 |
| Stock | 78 | 666 |

Combined totals include object-only and mixed elements. Unresolved elements are excluded from headline totals and runtime instrumentation. These are potential allocation sites, not a claim that every helper invocation allocates.

## Definitions and audit decisions

Each tree is loaded as one Program using its own `src/compiler/tsconfig.json`, including its extends, roots, references and compiler options. The config filename is supplied to the API so project-relative type resolution is preserved. Only non-declaration source files under `src/compiler` contribute sites. External modules are available to the checker but are not census files. The JSON includes diagnostic details, every accepted site, excluded primitive sites, types and source offsets.

Array recognition uses checker Array/ReadonlyArray/tuple types, NodeArray and sorted array interfaces, base types and constrained type parameters. Iterable inputs additionally recognize Iterable, iterator, generator and set element types through their checker type arguments and base types. This does not classify arbitrary array-like objects as arrays. A spread site is one spread element inside an array literal: two spreads in one literal are two source-copy sites. `Array.from` and `arrayFrom` accept iterable sources. Native method calls require a recognized array receiver. Helper identity is resolved through checker symbols and import aliases back to core.ts.

`object` means the element type contains object constituents only. `mixed` means a union with both object and primitive constituents; null, undefined and void count with primitive constituents. Nested arrays are objects. Branded primitive intersections such as Path and IncrementalBuildInfoFileId remain primitive. `any`, `unknown`, unconstrained generic elements, broad generic constraints such as T extends {}, and other unresolved types have their own columns. This avoids asserting that every instantiation of a generic helper copies objects: the constraint {} permits primitive values too. A mapping may produce primitive output: eligibility is based on its source element type, as requested.

A syntactic loop contributes one scope only within its body. An inline callback passed to forEach, map, filter, some, every, find or reduce contributes one scope; core helper callbacks use the requested forEach, some, every, map, filter, find and mapDefined names. Checker-resolved named functions used as these callback arguments also contribute one scope. Named callback declarations are counted once even when referenced more than once; their JSON flag identifies that evidence. This is a static callback-use classification, not context-sensitive call-graph expansion. Loop depth follows lexical ancestry and callback boundaries, not dynamic recursion or callers. A non-callback function boundary stops outer scope inheritance. A site with an outer lexical loop or iteration callback beyond that boundary is marked `uncertainOuterLoop`, an overlapping annotation on its definite in/out classification; it is not automatically promoted into a loop. No whole-program call-graph proof of nested function invocation is attempted.

The core helper audit includes the explicitly requested families plus other exported helpers that return fresh arrays from array inputs. `requestedFamilyTotals` in JSON excludes the additional helpers listed below. Scalar factories, iterator-returning helpers, Map-returning helpers, and mutators that never copy an existing array are excluded. In particular append, appendIfUnique, filterMutate, pushIfUnique, toArray and singleOrMany do not copy an existing input array.

| Helper family | Copy behavior |
| --- | --- |
| map, mapDefined, flatten, flatMapToMutable, arrayFrom | Fresh result when an array result is returned; map can return undefined. |
| filter, sameMap, sameFlatMap, compact | Can return the input unchanged; count runtime copies only when result identity differs. |
| concatenate, combine | Can return an input unchanged. concatenate copies when both inputs are nonempty; combine can also mutate and return an input. |
| flatMap | Returns shared emptyArray when there is no result, otherwise a fresh array. |
| toSorted, sortAndDeduplicate | Shared emptyArray for empty input; otherwise fresh. sortAndDeduplicate has internal intermediate arrays; the call site records final result length. |
| mapAllOrFail | Fresh result on success; may return undefined after constructing an abandoned partial result. Only returned arrays are measured. |
| Additional: zipWith, intersperse, spanMap, indicesOf, deduplicate, compact, relativeComplement, combine, addRange, replaceElement, getOwnValues, arrayToNumericMap, group, cartesianProduct, takeWhile, skipWhile | Audited in core.ts. Eligibility still requires an array source with object or mixed element type. addRange uses its second argument and is excluded when the checker proves its destination is already defined; relativeComplement uses arrayB; concatenate, combine and zipWith inspect both array sources. Input aliases and undefined returns are excluded from runtime copies. |

Static potential sites include conditional helpers even when static analysis cannot prove their allocating branch is taken. In particular sameMap returning its input contributes no runtime copy. All methods, including zero-count methods, have method rows; all compiler files, including zero-site files, have file rows.

## Adapted tree

79 root files; 79 compiler census files; 0 pre-emit diagnostics.

### By method

| Name | Object in | Object out | Mixed in | Mixed out | Unresolved in | Unresolved out | Uncertain outer loop |
| --- | --- | --- | --- | --- | --- | --- | --- |
| .concat | 1 | 3 | 0 | 0 | 0 | 0 | 0 |
| .filter | 0 | 34 | 0 | 2 | 0 | 0 | 0 |
| .flat | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| .flatMap | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| .map | 12 | 55 | 0 | 4 | 0 | 2 | 0 |
| .slice | 20 | 59 | 0 | 2 | 6 | 8 | 0 |
| .toReversed | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| .toSorted | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| Array.from | 0 | 1 | 0 | 0 | 0 | 0 | 0 |
| helper:addRange | 8 | 12 | 0 | 0 | 1 | 0 | 0 |
| helper:arrayFrom | 2 | 29 | 0 | 0 | 0 | 1 | 0 |
| helper:arrayToNumericMap | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| helper:cartesianProduct | 0 | 1 | 0 | 0 | 0 | 0 | 0 |
| helper:combine | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| helper:compact | 0 | 2 | 0 | 0 | 0 | 0 | 0 |
| helper:concatenate | 10 | 51 | 0 | 0 | 0 | 1 | 0 |
| helper:deduplicate | 0 | 2 | 0 | 0 | 0 | 0 | 0 |
| helper:filter | 5 | 97 | 0 | 4 | 0 | 0 | 0 |
| helper:flatMap | 2 | 14 | 0 | 1 | 0 | 0 | 0 |
| helper:flatMapToMutable | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| helper:flatten | 0 | 15 | 0 | 2 | 0 | 0 | 0 |
| helper:getOwnValues | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| helper:group | 0 | 2 | 0 | 0 | 0 | 0 | 0 |
| helper:indicesOf | 1 | 0 | 0 | 0 | 0 | 1 | 0 |
| helper:intersperse | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| helper:map | 12 | 123 | 0 | 2 | 0 | 1 | 0 |
| helper:mapAllOrFail | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| helper:mapDefined | 3 | 39 | 0 | 0 | 0 | 1 | 0 |
| helper:relativeComplement | 0 | 1 | 0 | 0 | 0 | 0 | 0 |
| helper:replaceElement | 0 | 2 | 0 | 0 | 0 | 0 | 0 |
| helper:sameFlatMap | 0 | 1 | 0 | 0 | 0 | 0 | 0 |
| helper:sameMap | 0 | 31 | 0 | 0 | 0 | 0 | 0 |
| helper:skipWhile | 0 | 1 | 0 | 0 | 0 | 0 | 0 |
| helper:sortAndDeduplicate | 1 | 2 | 0 | 0 | 0 | 0 | 0 |
| helper:spanMap | 0 | 2 | 0 | 0 | 0 | 0 | 0 |
| helper:takeWhile | 0 | 1 | 0 | 0 | 0 | 0 | 0 |
| helper:toSorted | 0 | 4 | 0 | 0 | 0 | 1 | 0 |
| helper:zipWith | 0 | 1 | 0 | 1 | 0 | 0 | 0 |
| spread | 8 | 91 | 2 | 3 | 0 | 4 | 0 |

### By file

| Name | Object in | Object out | Mixed in | Mixed out | Unresolved in | Unresolved out | Uncertain outer loop |
| --- | --- | --- | --- | --- | --- | --- | --- |
| src/compiler/_namespaces/ts.moduleSpecifiers.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/_namespaces/ts.performance.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/_namespaces/ts.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/binder.ts | 0 | 2 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/builder.ts | 8 | 10 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/builderPublic.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/builderState.ts | 0 | 3 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/builderStatePublic.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/checker.ts | 40 | 336 | 0 | 7 | 1 | 0 | 0 |
| src/compiler/commandLineParser.ts | 0 | 19 | 0 | 2 | 0 | 2 | 0 |
| src/compiler/core.ts | 0 | 1 | 0 | 0 | 6 | 13 | 0 |
| src/compiler/corePublic.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/debug.ts | 0 | 4 | 0 | 1 | 0 | 0 | 0 |
| src/compiler/diagnosticInformationMap.generated.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/emitter.ts | 0 | 3 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/executeCommandLine.ts | 0 | 13 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/expressionToTypeNode.ts | 1 | 9 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/factory/baseNodeFactory.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/factory/emitHelpers.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/factory/emitNode.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/factory/nodeChildren.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/factory/nodeConverters.ts | 0 | 4 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/factory/nodeFactory.ts | 0 | 13 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/factory/nodeTests.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/factory/parenthesizerRules.ts | 0 | 5 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/factory/utilities.ts | 0 | 1 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/factory/utilitiesPublic.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/hostErrors.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/moduleNameResolver.ts | 0 | 2 | 0 | 1 | 0 | 4 | 0 |
| src/compiler/moduleSpecifiers.ts | 0 | 3 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/parser.ts | 1 | 4 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/path.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/performance.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/performanceCore.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/program.ts | 1 | 37 | 0 | 4 | 0 | 0 | 0 |
| src/compiler/programDiagnostics.ts | 0 | 4 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/resolutionCache.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/scanner.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/semver.ts | 1 | 1 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/sourcemap.ts | 1 | 3 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/symbolWalker.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/sys.ts | 0 | 1 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/tracing.ts | 4 | 1 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformer.ts | 0 | 9 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/classFields.ts | 0 | 12 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/classThis.ts | 0 | 1 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/declarations.ts | 4 | 29 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/declarations/diagnostics.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/destructuring.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/es2015.ts | 11 | 13 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/es2016.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/es2017.ts | 0 | 3 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/es2018.ts | 1 | 4 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/es2019.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/es2020.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/es2021.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/esDecorators.ts | 0 | 2 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/esnext.ts | 0 | 4 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/generators.ts | 0 | 5 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/jsx.ts | 4 | 11 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/legacyDecorators.ts | 0 | 6 | 0 | 1 | 0 | 0 | 0 |
| src/compiler/transformers/module/esnextAnd2015.ts | 0 | 4 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts | 0 | 1 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/module/module.ts | 1 | 16 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/module/system.ts | 0 | 2 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/namedEvaluation.ts | 0 | 4 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/taggedTemplate.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/ts.ts | 0 | 13 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/typeSerializer.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/utilities.ts | 0 | 3 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/tsbuild.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/tsbuildPublic.ts | 0 | 1 | 2 | 2 | 0 | 0 | 0 |
| src/compiler/types.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/utilities.ts | 5 | 35 | 0 | 0 | 0 | 1 | 0 |
| src/compiler/utilitiesPublic.ts | 0 | 14 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/visitorPublic.ts | 2 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/watch.ts | 0 | 4 | 0 | 3 | 0 | 0 | 0 |
| src/compiler/watchPublic.ts | 0 | 1 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/watchUtilities.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |

### Ten deepest in-loop sites

| Location | Depth | Method | Source type | Enclosing function chain | Expression |
| --- | --- | --- | --- | --- | --- |
| src/compiler/builder.ts:2181:35 | 2 | helper:addRange | readonly Diagnostic[] | createBuilderProgram > getSemanticDiagnosticsOfNextAffectedFile > <anonymous> | addRange(<br>                        diagnostics,<br>                        getSemanticDiagnosticsOfFile(<br>                            state,<br>                            sourceFile,<br>                            cancellationToken,<br>                            semanticDiagnosticsPerFile,<br>                        ),<br>                    ) |
| src/compiler/checker.ts:14335:70 | 2 | helper:mapDefined | Signature[] | createTypeChecker > getUnionSignatures | mapDefined(unionSignatures, sig => sig.thisParameter && getTypeOfSymbol(sig.thisParameter)) |
| src/compiler/checker.ts:18174:116 | 2 | .map | Type[] | createTypeChecker > removeSubtypes | types.map(t => t.id) |
| src/compiler/checker.ts:51492:55 | 2 | helper:filter | ElementWithComputedPropertyName[] | createTypeChecker > createResolver > <anonymous> | filter(info.components, e => {<br>                                    // skip late bound props that contribute to the index signature - they'll be preserved via other means<br>                                    return !hasLateBindableName(e);<br>                                }) |
| src/compiler/checker.ts:51496:48 | 2 | helper:map | ElementWithComputedPropertyName[] | createTypeChecker > createResolver > <anonymous> | map(newComponents, e => {<br>                                    trackComputedName(e.name.expression as EntityNameExpression);<br>                                    const mods = infoList === staticInfos ? [factory.createModifier(SyntaxKind.StaticKeyword)] as Modifier[] : undefined;<br>                                    return factory.createPropertyDeclaration(<br>                                        append(mods, info.isReadonly ? factory.createModifier(SyntaxKind.ReadonlyKeyword) : undefined),<br>                                        e.name,<br>                                        (isPropertySignature(e) \|\| isPropertyDeclaration(e) \|\| isMethodSignature(e) \|\| isMethodDeclaration(e) \|\| isGetAccessor(e) \|\| isSetAccessor(e)) && e.questionToken ? factory.createToken(SyntaxKind.QuestionToken) : undefined,<br>                                        nodeBuilder.typeToTypeNode(getTypeOfSymbol(e.symbol), enclosing, flags, internalFlags, tracker),<br>                                        /*initializer*/ undefined,<br>                                    );<br>                                }) |
| src/compiler/builder.ts:1536:21 | 1 | .map | DiagnosticRelatedInformation[] | getBuildInfo > toReusableDiagnostic > <anonymous> | relatedInformation.map(r => toReusableDiagnosticRelatedInformation(r, diagnosticFilePath)) |
| src/compiler/builder.ts:1572:74 | 1 | .slice | DiagnosticMessageChain[] | getBuildInfo > toReusableDiagnosticMessageChainArray > <anonymous> | array.slice(0, index - 1) |
| src/compiler/builder.ts:1665:38 | 1 | .map | DiagnosticMessageChain[] | computeSignatureWithDiagnostics > flattenDiagnosticMessageText | diagnostic.next.map(flattenDiagnosticMessageText) |
| src/compiler/builder.ts:2061:35 | 1 | helper:addRange | readonly Diagnostic[] | createBuilderProgram > emit | addRange(diagnostics, affectedEmitResult.result.diagnostics) |
| src/compiler/builder.ts:2127:48 | 1 | helper:addRange | readonly Diagnostic[] | createBuilderProgram > getDeclarationDiagnostics | addRange(diagnostics, affectedEmitResult.result.diagnostics) |

Uncertain outer-loop sites: 0. Named callback sites: 8.

Explicitly named family subtotal: {"object":{"inLoop":76,"notInLoop":649,"uncertainOuterLoop":0},"mixed":{"inLoop":2,"notInLoop":20,"uncertainOuterLoop":0},"unresolved":{"inLoop":6,"notInLoop":19,"uncertainOuterLoop":0}}.

## Stock tree

78 root files; 78 compiler census files; 0 pre-emit diagnostics.

### By method

| Name | Object in | Object out | Mixed in | Mixed out | Unresolved in | Unresolved out | Uncertain outer loop |
| --- | --- | --- | --- | --- | --- | --- | --- |
| .concat | 1 | 3 | 0 | 0 | 0 | 0 | 0 |
| .filter | 0 | 34 | 0 | 2 | 0 | 0 | 0 |
| .flat | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| .flatMap | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| .map | 12 | 54 | 0 | 4 | 0 | 2 | 0 |
| .slice | 20 | 59 | 0 | 2 | 6 | 10 | 0 |
| .toReversed | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| .toSorted | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| Array.from | 0 | 1 | 0 | 0 | 0 | 0 | 0 |
| helper:addRange | 8 | 12 | 0 | 0 | 1 | 0 | 0 |
| helper:arrayFrom | 2 | 29 | 0 | 0 | 0 | 1 | 0 |
| helper:arrayToNumericMap | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| helper:cartesianProduct | 0 | 1 | 0 | 0 | 0 | 0 | 0 |
| helper:combine | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| helper:compact | 0 | 2 | 0 | 0 | 0 | 0 | 0 |
| helper:concatenate | 10 | 51 | 0 | 0 | 0 | 1 | 0 |
| helper:deduplicate | 0 | 2 | 0 | 0 | 0 | 0 | 0 |
| helper:filter | 5 | 97 | 0 | 3 | 0 | 1 | 0 |
| helper:flatMap | 2 | 14 | 0 | 1 | 0 | 0 | 0 |
| helper:flatMapToMutable | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| helper:flatten | 0 | 15 | 0 | 2 | 0 | 0 | 0 |
| helper:getOwnValues | 0 | 4 | 0 | 0 | 0 | 0 | 0 |
| helper:group | 0 | 2 | 0 | 0 | 0 | 0 | 0 |
| helper:indicesOf | 1 | 0 | 0 | 0 | 0 | 1 | 0 |
| helper:intersperse | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| helper:map | 12 | 124 | 0 | 2 | 0 | 1 | 0 |
| helper:mapAllOrFail | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| helper:mapDefined | 3 | 38 | 0 | 0 | 0 | 1 | 0 |
| helper:relativeComplement | 0 | 1 | 0 | 0 | 0 | 0 | 0 |
| helper:replaceElement | 0 | 2 | 0 | 0 | 0 | 0 | 0 |
| helper:sameFlatMap | 0 | 1 | 0 | 0 | 0 | 0 | 0 |
| helper:sameMap | 0 | 31 | 0 | 0 | 0 | 0 | 0 |
| helper:skipWhile | 0 | 1 | 0 | 0 | 0 | 0 | 0 |
| helper:sortAndDeduplicate | 1 | 2 | 0 | 0 | 0 | 0 | 0 |
| helper:spanMap | 0 | 2 | 0 | 0 | 0 | 0 | 0 |
| helper:takeWhile | 0 | 1 | 0 | 0 | 0 | 0 | 0 |
| helper:toSorted | 0 | 4 | 0 | 0 | 0 | 1 | 0 |
| helper:zipWith | 0 | 2 | 0 | 0 | 0 | 0 | 0 |
| spread | 8 | 91 | 2 | 2 | 0 | 4 | 0 |

### By file

| Name | Object in | Object out | Mixed in | Mixed out | Unresolved in | Unresolved out | Uncertain outer loop |
| --- | --- | --- | --- | --- | --- | --- | --- |
| src/compiler/_namespaces/ts.moduleSpecifiers.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/_namespaces/ts.performance.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/_namespaces/ts.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/binder.ts | 0 | 2 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/builder.ts | 8 | 9 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/builderPublic.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/builderState.ts | 0 | 3 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/builderStatePublic.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/checker.ts | 40 | 336 | 0 | 6 | 1 | 0 | 0 |
| src/compiler/commandLineParser.ts | 0 | 19 | 0 | 1 | 0 | 3 | 0 |
| src/compiler/core.ts | 0 | 1 | 0 | 0 | 6 | 15 | 0 |
| src/compiler/corePublic.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/debug.ts | 0 | 5 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/diagnosticInformationMap.generated.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/emitter.ts | 0 | 3 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/executeCommandLine.ts | 0 | 13 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/expressionToTypeNode.ts | 1 | 9 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/factory/baseNodeFactory.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/factory/emitHelpers.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/factory/emitNode.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/factory/nodeChildren.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/factory/nodeConverters.ts | 0 | 4 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/factory/nodeFactory.ts | 0 | 13 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/factory/nodeTests.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/factory/parenthesizerRules.ts | 0 | 5 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/factory/utilities.ts | 0 | 1 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/factory/utilitiesPublic.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/moduleNameResolver.ts | 0 | 2 | 0 | 1 | 0 | 4 | 0 |
| src/compiler/moduleSpecifiers.ts | 0 | 3 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/parser.ts | 1 | 4 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/path.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/performance.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/performanceCore.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/program.ts | 1 | 37 | 0 | 4 | 0 | 0 | 0 |
| src/compiler/programDiagnostics.ts | 0 | 4 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/resolutionCache.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/scanner.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/semver.ts | 1 | 1 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/sourcemap.ts | 1 | 3 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/symbolWalker.ts | 0 | 4 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/sys.ts | 0 | 1 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/tracing.ts | 4 | 1 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformer.ts | 0 | 9 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/classFields.ts | 0 | 12 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/classThis.ts | 0 | 1 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/declarations.ts | 4 | 29 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/declarations/diagnostics.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/destructuring.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/es2015.ts | 11 | 13 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/es2016.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/es2017.ts | 0 | 3 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/es2018.ts | 1 | 4 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/es2019.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/es2020.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/es2021.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/esDecorators.ts | 0 | 2 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/esnext.ts | 0 | 4 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/generators.ts | 0 | 5 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/jsx.ts | 4 | 11 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/legacyDecorators.ts | 0 | 6 | 0 | 1 | 0 | 0 | 0 |
| src/compiler/transformers/module/esnextAnd2015.ts | 0 | 4 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts | 0 | 1 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/module/module.ts | 1 | 16 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/module/system.ts | 0 | 2 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/namedEvaluation.ts | 0 | 4 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/taggedTemplate.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/ts.ts | 0 | 13 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/typeSerializer.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/transformers/utilities.ts | 0 | 3 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/tsbuild.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/tsbuildPublic.ts | 0 | 1 | 2 | 2 | 0 | 0 | 0 |
| src/compiler/types.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/utilities.ts | 5 | 35 | 0 | 0 | 0 | 1 | 0 |
| src/compiler/utilitiesPublic.ts | 0 | 14 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/visitorPublic.ts | 2 | 0 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/watch.ts | 0 | 4 | 0 | 3 | 0 | 0 | 0 |
| src/compiler/watchPublic.ts | 0 | 1 | 0 | 0 | 0 | 0 | 0 |
| src/compiler/watchUtilities.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |

### Ten deepest in-loop sites

| Location | Depth | Method | Source type | Enclosing function chain | Expression |
| --- | --- | --- | --- | --- | --- |
| src/compiler/builder.ts:2158:35 | 2 | helper:addRange | readonly Diagnostic[] | createBuilderProgram > getSemanticDiagnosticsOfNextAffectedFile > <anonymous> | addRange(<br>                        diagnostics,<br>                        getSemanticDiagnosticsOfFile(<br>                            state,<br>                            sourceFile,<br>                            cancellationToken,<br>                            semanticDiagnosticsPerFile,<br>                        ),<br>                    ) |
| src/compiler/checker.ts:14332:70 | 2 | helper:mapDefined | Signature[] | createTypeChecker > getUnionSignatures | mapDefined(unionSignatures, sig => sig.thisParameter && getTypeOfSymbol(sig.thisParameter)) |
| src/compiler/checker.ts:18171:116 | 2 | .map | Type[] | createTypeChecker > removeSubtypes | types.map(t => t.id) |
| src/compiler/checker.ts:51478:55 | 2 | helper:filter | ElementWithComputedPropertyName[] | createTypeChecker > createResolver > <anonymous> | filter(info.components, e => {<br>                                    // skip late bound props that contribute to the index signature - they'll be preserved via other means<br>                                    return !hasLateBindableName(e);<br>                                }) |
| src/compiler/checker.ts:51482:48 | 2 | helper:map | ElementWithComputedPropertyName[] | createTypeChecker > createResolver > <anonymous> | map(newComponents, e => {<br>                                    trackComputedName(e.name.expression as EntityNameExpression);<br>                                    const mods = infoList === staticInfos ? [factory.createModifier(SyntaxKind.StaticKeyword)] as Modifier[] : undefined;<br>                                    return factory.createPropertyDeclaration(<br>                                        append(mods, info.isReadonly ? factory.createModifier(SyntaxKind.ReadonlyKeyword) : undefined),<br>                                        e.name,<br>                                        (isPropertySignature(e) \|\| isPropertyDeclaration(e) \|\| isMethodSignature(e) \|\| isMethodDeclaration(e) \|\| isGetAccessor(e) \|\| isSetAccessor(e)) && e.questionToken ? factory.createToken(SyntaxKind.QuestionToken) : undefined,<br>                                        nodeBuilder.typeToTypeNode(getTypeOfSymbol(e.symbol), enclosing, flags, internalFlags, tracker),<br>                                        /*initializer*/ undefined,<br>                                    );<br>                                }) |
| src/compiler/builder.ts:1513:21 | 1 | .map | DiagnosticRelatedInformation[] | getBuildInfo > toReusableDiagnostic > <anonymous> | relatedInformation.map(r => toReusableDiagnosticRelatedInformation(r, diagnosticFilePath)) |
| src/compiler/builder.ts:1549:74 | 1 | .slice | DiagnosticMessageChain[] | getBuildInfo > toReusableDiagnosticMessageChainArray > <anonymous> | array.slice(0, index - 1) |
| src/compiler/builder.ts:1642:38 | 1 | .map | DiagnosticMessageChain[] | computeSignatureWithDiagnostics > flattenDiagnosticMessageText | diagnostic.next.map(flattenDiagnosticMessageText) |
| src/compiler/builder.ts:2038:35 | 1 | helper:addRange | readonly Diagnostic[] | createBuilderProgram > emit | addRange(diagnostics, affectedEmitResult.result.diagnostics) |
| src/compiler/builder.ts:2104:48 | 1 | helper:addRange | readonly Diagnostic[] | createBuilderProgram > getDeclarationDiagnostics | addRange(diagnostics, affectedEmitResult.result.diagnostics) |

Uncertain outer-loop sites: 0. Named callback sites: 8.

Explicitly named family subtotal: {"object":{"inLoop":76,"notInLoop":648,"uncertainOuterLoop":0},"mixed":{"inLoop":2,"notInLoop":18,"uncertainOuterLoop":0},"unresolved":{"inLoop":6,"notInLoop":22,"uncertainOuterLoop":0}}.

## Runtime: stock tsc checking its own compiler project

One measured noEmit run per variant, under Node v24.19.0. Both runs exited 0. Total instrumented calls: **108673**. Actual returned fresh-array/site executions: **83154**. Total result lengths: **152419 elements**.

20 of 87 selected sites ran; 13 returned fresh arrays. Thrown evaluations: 0.

Explicitly named family runtime subtotal: {"calls":85348,"executions":73940,"elements":141741}. Additional audited helpers: {"calls":23325,"executions":9214,"elements":10678}.

| Variant | Peak resident KiB | Peak resident MiB |
| --- | --- | --- |
| Without counters | 702852 | 686.38 |
| With counters | 703072 | 686.59 |
| Difference | 220 | 0.21 |

Peak resident memory is the completed Node child ru_maxrss from Linux wait4, including its exit-time serialization, not the Python monitor, runner or build process. The baseline has no counter or memory preload. One run per variant means the difference includes ordinary run-to-run variation. Counter overhead includes the preload, scalar table, wrapper calls and per-call argument arrays/thunks. No copied result or source array is retained in the counter table. runtime.json also retains an earlier resourceUsage sample for cross-checking; the table uses the complete-lifetime wait4 measurement.

### Top twenty sites by fresh-copy executions

| Location | Method | Calls | Copy executions | Elements | Mean length | Expression |
| --- | --- | --- | --- | --- | --- | --- |
| src/compiler/checker.ts:18816:34 | .slice | 65319 | 65319 | 131159 | 2.008 | types.slice() |
| src/compiler/checker.ts:15700:28 | helper:addRange | 22966 | 9182 | 10646 | 1.159 | addRange(declarations, prop.declarations) |
| src/compiler/checker.ts:26361:49 | .slice | 8073 | 8073 | 9883 | 1.224 | inference.candidates.slice() |
| src/compiler/visitorPublic.ts:350:27 | .slice | 379 | 379 | 463 | 1.222 | nodes.slice(0, i) |
| src/compiler/checker.ts:26362:61 | .slice | 62 | 62 | 62 | 1.000 | inference.contraCandidates.slice() |
| src/compiler/checker.ts:24293:30 | .slice | 40 | 40 | 18 | 0.450 | properties.slice(0, i) |
| src/compiler/checker.ts:16479:92 | helper:map | 35 | 35 | 71 | 2.029 | map(signature.compositeSignatures, getReturnTypeOfSignature) |
| src/compiler/utilities.ts:4668:22 | helper:addRange | 32 | 32 | 32 | 1.000 | addRange(result, filterOwnedJSDocTags(hostNode, node.jsDoc!)) |
| src/compiler/checker.ts:14491:73 | helper:map | 10 | 10 | 35 | 3.500 | map(types, t => getIndexTypeOfType(t, indexType)!) |
| src/compiler/checker.ts:13086:32 | helper:concatenate | 18 | 8 | 18 | 2.250 | concatenate(outerTypeParameters, getInferTypeParameters(node as ConditionalTypeNode)) |
| src/compiler/checker.ts:18150:62 | helper:map | 5 | 5 | 17 | 3.400 | map(types, t => t === source ? neverType : t) |
| src/compiler/checker.ts:20223:42 | helper:concatenate | 5 | 5 | 11 | 2.200 | concatenate(leftProp.declarations, rightProp.declarations) |
| src/compiler/checker.ts:14354:193 | helper:map | 4 | 4 | 4 | 1.000 | map(results, sig => combineSignaturesOfUnionMembers(sig, signature)) |
| src/compiler/builder.ts:1513:21 | .map | 0 | 0 | 0 | n/a | relatedInformation.map(r => toReusableDiagnosticRelatedInformation(r, diagnosticFilePath)) |
| src/compiler/builder.ts:1549:74 | .slice | 0 | 0 | 0 | n/a | array.slice(0, index - 1) |
| src/compiler/builder.ts:1642:38 | .map | 0 | 0 | 0 | n/a | diagnostic.next.map(flattenDiagnosticMessageText) |
| src/compiler/builder.ts:2038:35 | helper:addRange | 0 | 0 | 0 | n/a | addRange(diagnostics, affectedEmitResult.result.diagnostics) |
| src/compiler/builder.ts:2104:48 | helper:addRange | 78 | 0 | 0 | n/a | addRange(diagnostics, affectedEmitResult.result.diagnostics) |
| src/compiler/builder.ts:2158:35 | helper:addRange | 0 | 0 | 0 | n/a | addRange(<br>                        diagnostics,<br>                        getSemanticDiagnosticsOfFile(<br>                            state,<br>                            sourceFile,<br>                            cancellationToken,<br>                            semanticDiagnosticsPerFile,<br>                        ),<br>                    ) |
| src/compiler/builder.ts:2208:27 | helper:addRange | 249 | 0 | 0 | n/a | addRange(diagnostics, getSemanticDiagnosticsOfFile(state, sourceFile, cancellationToken)) |

### Top twenty sites by call executions

| Location | Method | Call executions | Fresh copies | Elements | Mean fresh length |
| --- | --- | --- | --- | --- | --- |
| src/compiler/checker.ts:18816:34 | .slice | 65319 | 65319 | 131159 | 2.008 |
| src/compiler/checker.ts:15700:28 | helper:addRange | 22966 | 9182 | 10646 | 1.159 |
| src/compiler/checker.ts:26361:49 | .slice | 8073 | 8073 | 9883 | 1.224 |
| src/compiler/checker.ts:14103:34 | helper:concatenate | 2740 | 0 | 0 | n/a |
| src/compiler/checker.ts:14104:39 | helper:concatenate | 2740 | 0 | 0 | n/a |
| src/compiler/checker.ts:14106:30 | helper:concatenate | 2740 | 0 | 0 | n/a |
| src/compiler/checker.ts:14106:54 | helper:filter | 2740 | 0 | 0 | n/a |
| src/compiler/transformers/declarations.ts:1853:49 | helper:filter | 438 | 0 | 0 | n/a |
| src/compiler/visitorPublic.ts:350:27 | .slice | 379 | 379 | 463 | 1.222 |
| src/compiler/builder.ts:2208:27 | helper:addRange | 249 | 0 | 0 | n/a |
| src/compiler/builder.ts:2104:48 | helper:addRange | 78 | 0 | 0 | n/a |
| src/compiler/checker.ts:26362:61 | .slice | 62 | 62 | 62 | 1.000 |
| src/compiler/checker.ts:24293:30 | .slice | 40 | 40 | 18 | 0.450 |
| src/compiler/checker.ts:16479:92 | helper:map | 35 | 35 | 71 | 2.029 |
| src/compiler/utilities.ts:4668:22 | helper:addRange | 32 | 32 | 32 | 1.000 |
| src/compiler/checker.ts:13086:32 | helper:concatenate | 18 | 8 | 18 | 2.250 |
| src/compiler/checker.ts:14491:73 | helper:map | 10 | 10 | 35 | 3.500 |
| src/compiler/checker.ts:18150:62 | helper:map | 5 | 5 | 17 | 3.400 |
| src/compiler/checker.ts:20223:42 | helper:concatenate | 5 | 5 | 11 | 2.200 |
| src/compiler/checker.ts:14354:193 | helper:map | 4 | 4 | 4 | 1.000 |

The full per-site runtime JSON includes zero-execution sites. `calls` measures entered wrapper evaluations, including any thrown evaluations; `executions` excludes unchanged helper inputs, undefined results and shared emptyArray results. The JSON records totalThrows and per-site throws. An array literal is wrapped after construction. Each eligible spread site receives the full result length, including explicit entries and other spreads. Multiple eligible spreads in one literal are separate site observations of the same new array, so their lengths must not be interpreted as distinct backing-store allocations. These are site-result counts: nested helper layers can observe the same backing array at more than one site. Result length is measured, not unique backing-array allocations, the number of primitive/object slots written internally, or retained graph bytes. No non-loop sites or unresolved generic implementation sites are instrumented. A single run can leave valid static sites cold.

The transform is applied only to a scratch copy of stock src, before bundling. The baseline is bundled from untouched pinned stock source. Both use the official esbuild bundle settings (CJS, Node, ES2020/node14.17, external packages, linked source maps) and identical copied built standard libraries. The source printer rewrites only scratch files; the adapted tree is never instrumented. Both execute the same stock src/compiler/tsconfig.json using --noEmit, --pretty false and a scratch --tsBuildInfoFile. This is normal project mode rather than build mode, and the metadata is removed between runs. The stock compiler itself is also built using npm run build:compiler.

Child commands, exit codes and durations are recorded in run-steps.json; full child output is in the adjacent logs. The bounded runtime phase completed in 56.993 seconds.

## Reproduction and validation

```sh
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh > /tmp/setup.log 2>&1
# Inspect and preserve each child exit code before continuing.
npm ci --prefix stage3/api > /tmp/api.log 2>&1
export ADAMIC_TYPESCRIPT_SOURCE=/absolute/path/to/pinned/TypeScript
bash stage3/apply.sh /absolute/path/to/new/adapted-tree > /tmp/apply.log 2>&1
# Install stock dependencies, then run npm run build:compiler in the pinned tree.
node stage3/census/array-copies/validate.cjs stage3/api/node_modules/typescript/lib/typescript.js > /tmp/validation.log 2>&1
node stage3/census/array-copies/census.cjs stage3/api/node_modules/typescript/lib/typescript.js "$ADAMIC_TYPESCRIPT_SOURCE" stage3/census/array-copies/stock.json > /tmp/stock.log 2>&1
node stage3/census/array-copies/census.cjs stage3/api/node_modules/typescript/lib/typescript.js /absolute/path/to/adapted-tree stage3/census/array-copies/adapted.json > /tmp/adapted.log 2>&1
node stage3/census/array-copies/run.cjs "$ADAMIC_TYPESCRIPT_SOURCE" /absolute/path/to/new/runtime-scratch stage3/census/array-copies > /tmp/runtime.log 2>&1
node stage3/census/array-copies/report.cjs
```

Focused integration validation checks branded primitives, broad generic constraints, mixed unions, Map and custom iterables, inline and named callback depth, function-boundary uncertainty, optional-call argument evaluation, full literal lengths with multiple spreads, unique site IDs, preservation of program output, and sameMap identity and changed-result counts. verify-results.cjs checks the complete source hashes, method/file total reconciliation, exact runtime site coverage, numeric totals, and matching baseline/counted output. Its result is saved in verification.json. Baseline and counted project runs must both finish with exit 0. JSON data retains unresolved sites rather than hiding checker limitations. Provenance and source hashes are in provenance.json.
