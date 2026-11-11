# Paired overload-result census: task z5d0jye

Report only. No compiler or adaptation changes.

Parent compiler: `efe887763f26e96653546fae904d57539139f7ae`. Head compiler: `10b6c33daf5c5db84f5a423e28701104481d8b04`. The optimizer branch was checked before setup and still pointed at the requested head.

Both runs use exactly one adapted tree, produced by efe88776 stage3/apply.sh from TypeScript 6.0.3 `050880ce59e30b356b686bd3144efe24f875ebc8`. Both compilers were built from their own SHA with their recorded submodules. The same parent meter latent overlay and entry driver instrument each compiler. `LATENT_FULL=0`, `LATENT_ASSERT_NO_OUTPUT=1`; the tsc entry resolves 81 source files and has 0 checker diagnostics. Headers match exactly; both streams have completed relation trailers.

## Counts

| Measure | Parent | Head |
| --- | ---: | ---: |
| Units | 4562 | 4562 |
| Attempted units (including recovered panics) | 4058 | 4058 |
| Attempted units with no lowering first stop | 3158 | 3174 |
| Overload-result first stops | 110 | 67 |

| Paired delta | Raw | Distinct |
| --- | ---: | ---: |
| Cleared findings with no remaining finding at the identity | 10 | 10 |
| Cleared findings with another reason remaining | 0 | 0 |
| New findings | 2 | 2 |
| New sites | 2 | 2 |
| Changed unit first stops | 43 | 43 |
| Regressions | 0 | 0 |
| before NotYet | 525 | 518 |
| before Refused | 4466 | 4271 |
| after NotYet | 527 | 520 |
| after Refused | 4456 | 4261 |

Requested subset: 43/43 overload-result stops clear; 16 units then lower without another lowering finding. Other overload-result subset: 67/67 first stops unchanged.

A finding disappearing at a source identity is not proof that its enclosing unit lowers. Newly exposed reasons in an already-blocked unit are listed as changed first stops. Regressions are previously finding-free attempted units that now stop, previously finding-free refusal-scan sites that now find, or new panics. Raw and distinct access-chain pairing use stage3/census/pair without modifying the tool. The complete delta, reason transitions, source hashes and panics are in [delta.json](delta.json); [pair-report.md](pair-report.md) is the tool-generated report.

## The requested 43 units

| Unit name | Unit location | Overload-result stop | Next first stop |
| --- | --- | --- | --- |
| createParenthesizerRules | src/compiler/factory/parenthesizerRules.ts:54:1 | cleared | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/transformers/destructuring.ts:192:9 |
| getExpandoInitializer | src/compiler/utilities.ts:3992:1 | cleared | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/transformers/destructuring.ts:192:9 |
| getNewTargetContainer | src/compiler/utilities.ts:3291:1 | cleared | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/binder.ts:2258:13 |
| getSuperContainer | src/compiler/utilities.ts:3333:1 | cleared | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3362:24 |
| getSuperContainer | src/compiler/utilities.ts:3335:1 | cleared | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3362:24 |
| getSuperContainer | src/compiler/utilities.ts:3336:1 | cleared | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3362:24 |
| getThisContainer | src/compiler/utilities.ts:3180:1 | cleared | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3245:24 |
| getThisContainer | src/compiler/utilities.ts:3182:1 | cleared | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3245:24 |
| getThisContainer | src/compiler/utilities.ts:3184:1 | cleared | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3245:24 |
| getThisContainer | src/compiler/utilities.ts:3186:1 | cleared | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3245:24 |
| getThisContainer | src/compiler/utilities.ts:3187:1 | cleared | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3245:24 |
| isAnonymousFunctionDefinition | src/compiler/utilities.ts:5483:1 | cleared | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/transformers/destructuring.ts:192:9 |
| isCompoundLikeAssignment | src/compiler/utilities.ts:4927:1 | cleared | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/transformers/destructuring.ts:192:9 |
| isExpression | src/compiler/utilitiesPublic.ts:2086:1 | cleared | Lowered |
| isLeftHandSideExpression | src/compiler/utilitiesPublic.ts:1995:1 | cleared | Lowered |
| isUnaryExpression | src/compiler/utilitiesPublic.ts:2039:1 | cleared | Lowered |
| nodeCanBeDecorated | src/compiler/utilities.ts:3469:1 | cleared | NotYet: checked predicate overload target NamedDeclaration & { name: DeclarationName; } at src/compiler/utilitiesPublic.ts:946:1 |
| nodeCanBeDecorated | src/compiler/utilities.ts:3471:1 | cleared | NotYet: checked predicate overload target NamedDeclaration & { name: DeclarationName; } at src/compiler/utilitiesPublic.ts:946:1 |
| nodeCanBeDecorated | src/compiler/utilities.ts:3473:1 | cleared | NotYet: checked predicate overload target NamedDeclaration & { name: DeclarationName; } at src/compiler/utilitiesPublic.ts:946:1 |
| nodeCanBeDecorated | src/compiler/utilities.ts:3475:1 | cleared | NotYet: checked predicate overload target NamedDeclaration & { name: DeclarationName; } at src/compiler/utilitiesPublic.ts:946:1 |
| nodeCanBeDecorated | src/compiler/utilities.ts:3477:1 | cleared | NotYet: checked predicate overload target NamedDeclaration & { name: DeclarationName; } at src/compiler/utilitiesPublic.ts:946:1 |
| nodeIsDecorated | src/compiler/utilities.ts:3523:1 | cleared | Lowered |
| nodeIsDecorated | src/compiler/utilities.ts:3525:1 | cleared | Lowered |
| nodeIsDecorated | src/compiler/utilities.ts:3527:1 | cleared | Lowered |
| nodeIsDecorated | src/compiler/utilities.ts:3529:1 | cleared | Lowered |
| skipOuterExpressions | src/compiler/factory/utilities.ts:656:1 | cleared | Lowered |
| skipOuterExpressions | src/compiler/factory/utilities.ts:658:1 | cleared | Lowered |
| skipOuterExpressions | src/compiler/factory/utilities.ts:660:1 | cleared | Lowered |
| skipParentheses | src/compiler/utilities.ts:5039:1 | cleared | Lowered |
| skipParentheses | src/compiler/utilities.ts:5041:1 | cleared | Lowered |
| skipParentheses | src/compiler/utilities.ts:5043:1 | cleared | Lowered |
| skipPartiallyEmittedExpressions | src/compiler/utilitiesPublic.ts:1431:1 | cleared | Lowered |
| skipPartiallyEmittedExpressions | src/compiler/utilitiesPublic.ts:1432:1 | cleared | Lowered |
| skipPartiallyEmittedExpressions | src/compiler/utilitiesPublic.ts:1433:1 | cleared | Lowered |
| transformES2017 | src/compiler/transformers/es2017.ts:119:1 | cleared | Refused: overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor at src/compiler/visitorPublic.ts:587:119 |
| transformES2018 | src/compiler/transformers/es2018.ts:154:1 | cleared | NotYet: writing a possibly absent optional own field at src/compiler/factory/emitNode.ts:262:13 |
| visitFunctionBody | src/compiler/visitorPublic.ts:508:1 | cleared | NotYet: a branded supertype inside a checked predicate's array target at src/compiler/visitorPublic.ts:526:9 |
| visitFunctionBody | src/compiler/visitorPublic.ts:513:1 | cleared | NotYet: a branded supertype inside a checked predicate's array target at src/compiler/visitorPublic.ts:526:9 |
| visitFunctionBody | src/compiler/visitorPublic.ts:518:1 | cleared | NotYet: a branded supertype inside a checked predicate's array target at src/compiler/visitorPublic.ts:526:9 |
| visitFunctionBody | src/compiler/visitorPublic.ts:519:18 | cleared | NotYet: a branded supertype inside a checked predicate's array target at src/compiler/visitorPublic.ts:526:9 |
| visitFunctionBody | src/compiler/visitorPublic.ts:520:18 | cleared | NotYet: a branded supertype inside a checked predicate's array target at src/compiler/visitorPublic.ts:526:9 |
| visitFunctionBody | src/compiler/visitorPublic.ts:521:18 | cleared | NotYet: a branded supertype inside a checked predicate's array target at src/compiler/visitorPublic.ts:526:9 |
| visitFunctionBody | src/compiler/visitorPublic.ts:522:1 | cleared | NotYet: a branded supertype inside a checked predicate's array target at src/compiler/visitorPublic.ts:526:9 |

## The other 67 overload-result units

| Unit name | Unit location | Overload-result stop | Next first stop |
| --- | --- | --- | --- |
| commandLineOptionsToMap | src/compiler/commandLineParser.ts:2307:1 | unchanged | Refused: overload 2 of arrayToMap result Map<K, V2> cannot be served by implementation result Map<K, V1 \| V2> at src/compiler/core.ts:1400:1 |
| comparePathsWorker | src/compiler/path.ts:884:1 | unchanged | Refused: overload 1 of getPathComponents result PathPathComponents cannot be served by implementation result string[] at src/compiler/path.ts:492:1 |
| containsPath | src/compiler/path.ts:963:1 | unchanged | Refused: overload 1 of getPathComponents result PathPathComponents cannot be served by implementation result string[] at src/compiler/path.ts:492:1 |
| containsPath | src/compiler/path.ts:965:1 | unchanged | Refused: overload 1 of getPathComponents result PathPathComponents cannot be served by implementation result string[] at src/compiler/path.ts:492:1 |
| containsPath | src/compiler/path.ts:967:1 | unchanged | Refused: overload 1 of getPathComponents result PathPathComponents cannot be served by implementation result string[] at src/compiler/path.ts:492:1 |
| convertCompilerOptionsFromJson | src/compiler/commandLineParser.ts:3755:1 | unchanged | Refused: overload 1 of convertOptionsFromJson result WatchOptions \| undefined cannot be served by implementation result CompilerOptions \| TypeAcquisition \| WatchOptions \| undefined at src/compiler/commandLineParser.ts:3797:1 |
| convertToTSConfig | src/compiler/commandLineParser.ts:2660:1 | unchanged | Refused: overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:261:1 |
| convertTypeAcquisitionFromJson | src/compiler/commandLineParser.ts:3761:1 | unchanged | Refused: overload 1 of convertOptionsFromJson result WatchOptions \| undefined cannot be served by implementation result CompilerOptions \| TypeAcquisition \| WatchOptions \| undefined at src/compiler/commandLineParser.ts:3797:1 |
| createBuilderProgram | src/compiler/builder.ts:1686:1 | unchanged | Refused: overload 1 of createBuilderProgram result SemanticDiagnosticsBuilderProgram cannot be served by implementation result BuilderProgram \| undefined at src/compiler/builder.ts:1686:1 |
| createBuilderProgram | src/compiler/builder.ts:1691:1 | unchanged | Refused: overload 1 of createBuilderProgram result SemanticDiagnosticsBuilderProgram cannot be served by implementation result BuilderProgram \| undefined at src/compiler/builder.ts:1686:1 |
| createBuilderProgram | src/compiler/builder.ts:1696:1 | unchanged | Refused: overload 1 of createBuilderProgram result SemanticDiagnosticsBuilderProgram cannot be served by implementation result BuilderProgram \| undefined at src/compiler/builder.ts:1686:1 |
| createBuilderProgramUsingIncrementalBuildInfo | src/compiler/builder.ts:2277:1 | unchanged | Refused: overload 2 of arrayToMap result Map<K, V2> cannot be served by implementation result Map<K, V1 \| V2> at src/compiler/core.ts:1400:1 |
| createCommentDirectivesMap | src/compiler/utilities.ts:1210:1 | unchanged | Refused: overload 1 of arrayFrom result U[] cannot be served by implementation result (T \| U)[] at src/compiler/core.ts:1334:1 |
| createDiagnosticCollection | src/compiler/utilities.ts:6058:1 | unchanged | Refused: overload 1 of getDiagnostics result DiagnosticWithLocation[] cannot be served by implementation result Diagnostic[] at src/compiler/utilities.ts:6120:5 |
| createDiagnosticForInvalidCustomType | src/compiler/commandLineParser.ts:1873:1 | unchanged | Refused: overload 1 of arrayFrom result U[] cannot be served by implementation result (T \| U)[] at src/compiler/core.ts:1334:1 |
| createEmitAndSemanticDiagnosticsBuilderProgram | src/compiler/builderPublic.ts:201:1 | unchanged | Refused: overload 1 of createBuilderProgram result SemanticDiagnosticsBuilderProgram cannot be served by implementation result BuilderProgram \| undefined at src/compiler/builder.ts:1686:1 |
| createEmitAndSemanticDiagnosticsBuilderProgram | src/compiler/builderPublic.ts:207:1 | unchanged | Refused: overload 1 of createBuilderProgram result SemanticDiagnosticsBuilderProgram cannot be served by implementation result BuilderProgram \| undefined at src/compiler/builder.ts:1686:1 |
| createEmitAndSemanticDiagnosticsBuilderProgram | src/compiler/builderPublic.ts:215:1 | unchanged | Refused: overload 1 of createBuilderProgram result SemanticDiagnosticsBuilderProgram cannot be served by implementation result BuilderProgram \| undefined at src/compiler/builder.ts:1686:1 |
| createPrinter | src/compiler/emitter.ts:1211:1 | unchanged | Refused: overload 1 of writeTokenText result void cannot be served by implementation result number at src/compiler/emitter.ts:4949:5 |
| createSemanticDiagnosticsBuilderProgram | src/compiler/builderPublic.ts:162:1 | unchanged | Refused: overload 1 of createBuilderProgram result SemanticDiagnosticsBuilderProgram cannot be served by implementation result BuilderProgram \| undefined at src/compiler/builder.ts:1686:1 |
| createSemanticDiagnosticsBuilderProgram | src/compiler/builderPublic.ts:168:1 | unchanged | Refused: overload 1 of createBuilderProgram result SemanticDiagnosticsBuilderProgram cannot be served by implementation result BuilderProgram \| undefined at src/compiler/builder.ts:1686:1 |
| createSemanticDiagnosticsBuilderProgram | src/compiler/builderPublic.ts:176:1 | unchanged | Refused: overload 1 of createBuilderProgram result SemanticDiagnosticsBuilderProgram cannot be served by implementation result BuilderProgram \| undefined at src/compiler/builder.ts:1686:1 |
| decodeMappings | src/compiler/sourcemap.ts:464:1 | unchanged | Refused: overload 1 of captureMapping result Required<Mapping> cannot be served by implementation result Mapping at src/compiler/sourcemap.ts:547:5 |
| emitDetachedComments | src/compiler/utilities.ts:7010:1 | unchanged | Refused: overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:261:1 |
| filterSemanticDiagnostics | src/compiler/program.ts:5099:1 | unchanged | Refused: overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:261:1 |
| generateOptionOutput | src/compiler/executeCommandLine.ts:244:1 | unchanged | Refused: overload 1 of arrayFrom result U[] cannot be served by implementation result (T \| U)[] at src/compiler/core.ts:1334:1 |
| getAllSuperTypeNodes | src/compiler/utilities.ts:5227:1 | unchanged | Refused: overload 1 of concatenate result T[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:656:1 |
| getBuildInfo | src/compiler/builder.ts:1258:1 | unchanged | Refused: overload 1 of arrayFrom result U[] cannot be served by implementation result (T \| U)[] at src/compiler/core.ts:1334:1 |
| getCommonSourceDirectoryOfConfig | src/compiler/emitter.ts:685:1 | unchanged | Refused: overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:261:1 |
| getConditions | src/compiler/moduleNameResolver.ts:755:1 | unchanged | Refused: overload 1 of concatenate result T[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:656:1 |
| getDeclarationDiagnostics | src/compiler/transformers/declarations.ts:221:1 | unchanged | Refused: overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:261:1 |
| getDecorators | src/compiler/utilitiesPublic.ts:1024:1 | unchanged | Refused: overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:261:1 |
| getFileNamesFromConfigSpecs | src/compiler/commandLineParser.ts:3950:1 | unchanged | Refused: overload 1 of getSupportedExtensions result readonly Extension[][] cannot be served by implementation result readonly string[][] at src/compiler/utilities.ts:9990:1 |
| getFilesInErrorForSummary | src/compiler/watch.ts:233:1 | unchanged | Refused: overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:261:1 |
| getImportedHelpers | src/compiler/factory/utilities.ts:760:1 | unchanged | Refused: overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:261:1 |
| getInitializedVariables | src/compiler/utilities.ts:8017:1 | unchanged | Refused: overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:261:1 |
| getJSDocCommentRanges | src/compiler/utilities.ts:2761:1 | unchanged | Refused: overload 1 of concatenate result T[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:656:1 |
| getNormalizedPathComponents | src/compiler/path.ts:622:1 | unchanged | Refused: overload 1 of getPathComponents result PathPathComponents cannot be served by implementation result string[] at src/compiler/path.ts:492:1 |
| getOptionsForHelp | src/compiler/executeCommandLine.ts:176:1 | unchanged | Refused: overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:261:1 |
| getParseTreeNode | src/compiler/utilitiesPublic.ts:817:1 | unchanged | Refused: overload 2 of getParseTreeNode result T \| undefined cannot be served by implementation result Node \| undefined at src/compiler/utilitiesPublic.ts:826:1 |
| getParseTreeNode | src/compiler/utilitiesPublic.ts:827:1 | unchanged | Refused: overload 2 of getParseTreeNode result T \| undefined cannot be served by implementation result Node \| undefined at src/compiler/utilitiesPublic.ts:826:1 |
| getPathComponents | src/compiler/path.ts:492:1 | unchanged | Refused: overload 1 of getPathComponents result PathPathComponents cannot be served by implementation result string[] at src/compiler/path.ts:492:1 |
| getPathComponents | src/compiler/path.ts:494:1 | unchanged | Refused: overload 1 of getPathComponents result PathPathComponents cannot be served by implementation result string[] at src/compiler/path.ts:492:1 |
| getPathComponents | src/compiler/path.ts:495:1 | unchanged | Refused: overload 1 of getPathComponents result PathPathComponents cannot be served by implementation result string[] at src/compiler/path.ts:492:1 |
| getPathComponentsRelativeTo | src/compiler/path.ts:1010:1 | unchanged | Refused: overload 1 of getPathComponents result PathPathComponents cannot be served by implementation result string[] at src/compiler/path.ts:492:1 |
| getProperties | src/compiler/transformers/utilities.ts:585:1 | unchanged | Refused: overload 1 of getProperties result readonly InitializedPropertyDeclaration[] cannot be served by implementation result readonly PropertyDeclaration[] at src/compiler/transformers/utilities.ts:585:1 |
| getProperties | src/compiler/transformers/utilities.ts:587:1 | unchanged | Refused: overload 1 of getProperties result readonly InitializedPropertyDeclaration[] cannot be served by implementation result readonly PropertyDeclaration[] at src/compiler/transformers/utilities.ts:585:1 |
| getProperties | src/compiler/transformers/utilities.ts:589:1 | unchanged | Refused: overload 1 of getProperties result readonly InitializedPropertyDeclaration[] cannot be served by implementation result readonly PropertyDeclaration[] at src/compiler/transformers/utilities.ts:585:1 |
| getRequiresAtTopOfFile | src/compiler/utilities.ts:10125:1 | unchanged | Refused: overload 1 of concatenate result T[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:656:1 |
| getSemanticDiagnosticsOfFile | src/compiler/builder.ts:1058:1 | unchanged | Refused: overload 1 of concatenate result T[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:656:1 |
| getSemanticJsxChildren | src/compiler/utilities.ts:6044:1 | unchanged | Refused: overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:261:1 |
| getSourceFilesToEmit | src/compiler/utilities.ts:6632:1 | unchanged | Refused: overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:261:1 |
| getStaticPropertiesAndClassStaticBlock | src/compiler/transformers/utilities.ts:598:1 | unchanged | Refused: overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:261:1 |
| getStaticPropertiesAndClassStaticBlock | src/compiler/transformers/utilities.ts:600:1 | unchanged | Refused: overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:261:1 |
| getStaticPropertiesAndClassStaticBlock | src/compiler/transformers/utilities.ts:602:1 | unchanged | Refused: overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:261:1 |
| getSupportedExtensions | src/compiler/utilities.ts:9990:1 | unchanged | Refused: overload 1 of getSupportedExtensions result readonly Extension[][] cannot be served by implementation result readonly string[][] at src/compiler/utilities.ts:9990:1 |
| getSupportedExtensions | src/compiler/utilities.ts:9992:1 | unchanged | Refused: overload 1 of getSupportedExtensions result readonly Extension[][] cannot be served by implementation result readonly string[][] at src/compiler/utilities.ts:9990:1 |
| getSupportedExtensions | src/compiler/utilities.ts:9994:1 | unchanged | Refused: overload 1 of getSupportedExtensions result readonly Extension[][] cannot be served by implementation result readonly string[][] at src/compiler/utilities.ts:9990:1 |
| guessDirectorySymlink | src/compiler/utilities.ts:9559:1 | unchanged | Refused: overload 1 of getPathComponents result PathPathComponents cannot be served by implementation result string[] at src/compiler/path.ts:492:1 |
| injectClassNamedEvaluationHelperBlockIfMissing | src/compiler/transformers/namedEvaluation.ts:179:1 | unchanged | Refused: overload 1 of injectClassNamedEvaluationHelperBlockIfMissing result Extract<ClassDeclaration, Pick<...>> \| Extract<...> cannot be served by implementation result ClassLikeDeclaration at src/compiler/transformers/namedEvaluation.ts:173:1 |
| isSupportedSourceFileName | src/compiler/utilities.ts:10146:1 | unchanged | Refused: overload 1 of getSupportedExtensions result readonly Extension[][] cannot be served by implementation result readonly string[][] at src/compiler/utilities.ts:9990:1 |
| matchesExclude | src/compiler/commandLineParser.ts:4077:1 | unchanged | Refused: overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:261:1 |
| normalizePathForCJSResolution | src/compiler/moduleNameResolver.ts:1966:1 | unchanged | Refused: overload 1 of getPathComponents result PathPathComponents cannot be served by implementation result string[] at src/compiler/path.ts:492:1 |
| printAllHelp | src/compiler/executeCommandLine.ts:513:1 | unchanged | Refused: overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:261:1 |
| printBuildHelp | src/compiler/executeCommandLine.ts:523:1 | unchanged | Refused: overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:261:1 |
| repopulateDiagnostics | src/compiler/builder.ts:520:1 | unchanged | Refused: overload 1 of sameMap result U[] cannot be served by implementation result readonly U[] \| undefined at src/compiler/core.ts:342:1 |
| transformNamedEvaluation | src/compiler/transformers/namedEvaluation.ts:470:1 | unchanged | Refused: overload 1 of transformNamedEvaluation result Extract<AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; }, Pick<...>> \| ... 7 more ... \| Extract<...> cannot be served by implementation result BinaryExpression \| BindingElement \| ExportAssignment \| ParameterDeclaration \| PropertyAssignment \| PropertyDeclaration \| ShorthandPropertyAssignment \| VariableDeclaration at src/compiler/transformers/namedEvaluation.ts:469:1 |

## All changed lowering first stops

| Unit | Unit location | Parent first stop | Head first stop |
| --- | --- | --- | --- |
| skipPartiallyEmittedExpressions | src/compiler/utilitiesPublic.ts:1431:1 | Refused: overload 1 of skipPartiallyEmittedExpressions result Expression cannot be served by implementation result Node at src/compiler/utilitiesPublic.ts:1431:1 | Lowered |
| skipPartiallyEmittedExpressions | src/compiler/utilitiesPublic.ts:1432:1 | Refused: overload 1 of skipPartiallyEmittedExpressions result Expression cannot be served by implementation result Node at src/compiler/utilitiesPublic.ts:1431:1 | Lowered |
| skipPartiallyEmittedExpressions | src/compiler/utilitiesPublic.ts:1433:1 | Refused: overload 1 of skipPartiallyEmittedExpressions result Expression cannot be served by implementation result Node at src/compiler/utilitiesPublic.ts:1431:1 | Lowered |
| isLeftHandSideExpression | src/compiler/utilitiesPublic.ts:1995:1 | Refused: overload 1 of skipPartiallyEmittedExpressions result Expression cannot be served by implementation result Node at src/compiler/utilitiesPublic.ts:1431:1 | Lowered |
| isUnaryExpression | src/compiler/utilitiesPublic.ts:2039:1 | Refused: overload 1 of skipPartiallyEmittedExpressions result Expression cannot be served by implementation result Node at src/compiler/utilitiesPublic.ts:1431:1 | Lowered |
| isExpression | src/compiler/utilitiesPublic.ts:2086:1 | Refused: overload 1 of skipPartiallyEmittedExpressions result Expression cannot be served by implementation result Node at src/compiler/utilitiesPublic.ts:1431:1 | Lowered |
| getThisContainer | src/compiler/utilities.ts:3180:1 | Refused: overload 1 of getThisContainer result ThisContainer cannot be served by implementation result ArrowFunction \| ComputedPropertyName \| ThisContainer at src/compiler/utilities.ts:3180:1 | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3245:24 |
| getThisContainer | src/compiler/utilities.ts:3182:1 | Refused: overload 1 of getThisContainer result ThisContainer cannot be served by implementation result ArrowFunction \| ComputedPropertyName \| ThisContainer at src/compiler/utilities.ts:3180:1 | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3245:24 |
| getThisContainer | src/compiler/utilities.ts:3184:1 | Refused: overload 1 of getThisContainer result ThisContainer cannot be served by implementation result ArrowFunction \| ComputedPropertyName \| ThisContainer at src/compiler/utilities.ts:3180:1 | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3245:24 |
| getThisContainer | src/compiler/utilities.ts:3186:1 | Refused: overload 1 of getThisContainer result ThisContainer cannot be served by implementation result ArrowFunction \| ComputedPropertyName \| ThisContainer at src/compiler/utilities.ts:3180:1 | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3245:24 |
| getThisContainer | src/compiler/utilities.ts:3187:1 | Refused: overload 1 of getThisContainer result ThisContainer cannot be served by implementation result ArrowFunction \| ComputedPropertyName \| ThisContainer at src/compiler/utilities.ts:3180:1 | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3245:24 |
| getNewTargetContainer | src/compiler/utilities.ts:3291:1 | Refused: overload 1 of getThisContainer result ThisContainer cannot be served by implementation result ArrowFunction \| ComputedPropertyName \| ThisContainer at src/compiler/utilities.ts:3180:1 | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/binder.ts:2258:13 |
| getSuperContainer | src/compiler/utilities.ts:3333:1 | Refused: overload 1 of getSuperContainer result SuperContainer \| undefined cannot be served by implementation result SuperContainerOrFunctions \| undefined at src/compiler/utilities.ts:3333:1 | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3362:24 |
| getSuperContainer | src/compiler/utilities.ts:3335:1 | Refused: overload 1 of getSuperContainer result SuperContainer \| undefined cannot be served by implementation result SuperContainerOrFunctions \| undefined at src/compiler/utilities.ts:3333:1 | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3362:24 |
| getSuperContainer | src/compiler/utilities.ts:3336:1 | Refused: overload 1 of getSuperContainer result SuperContainer \| undefined cannot be served by implementation result SuperContainerOrFunctions \| undefined at src/compiler/utilities.ts:3333:1 | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3362:24 |
| nodeCanBeDecorated | src/compiler/utilities.ts:3469:1 | Refused: overload 1 of nodeCanBeDecorated result true cannot be served by implementation result boolean at src/compiler/utilities.ts:3469:1 | NotYet: checked predicate overload target NamedDeclaration & { name: DeclarationName; } at src/compiler/utilitiesPublic.ts:946:1 |
| nodeCanBeDecorated | src/compiler/utilities.ts:3471:1 | Refused: overload 1 of nodeCanBeDecorated result true cannot be served by implementation result boolean at src/compiler/utilities.ts:3469:1 | NotYet: checked predicate overload target NamedDeclaration & { name: DeclarationName; } at src/compiler/utilitiesPublic.ts:946:1 |
| nodeCanBeDecorated | src/compiler/utilities.ts:3473:1 | Refused: overload 1 of nodeCanBeDecorated result true cannot be served by implementation result boolean at src/compiler/utilities.ts:3469:1 | NotYet: checked predicate overload target NamedDeclaration & { name: DeclarationName; } at src/compiler/utilitiesPublic.ts:946:1 |
| nodeCanBeDecorated | src/compiler/utilities.ts:3475:1 | Refused: overload 1 of nodeCanBeDecorated result true cannot be served by implementation result boolean at src/compiler/utilities.ts:3469:1 | NotYet: checked predicate overload target NamedDeclaration & { name: DeclarationName; } at src/compiler/utilitiesPublic.ts:946:1 |
| nodeCanBeDecorated | src/compiler/utilities.ts:3477:1 | Refused: overload 1 of nodeCanBeDecorated result true cannot be served by implementation result boolean at src/compiler/utilities.ts:3469:1 | NotYet: checked predicate overload target NamedDeclaration & { name: DeclarationName; } at src/compiler/utilitiesPublic.ts:946:1 |
| nodeIsDecorated | src/compiler/utilities.ts:3523:1 | Refused: overload 1 of nodeCanBeDecorated result true cannot be served by implementation result boolean at src/compiler/utilities.ts:3469:1 | Lowered |
| nodeIsDecorated | src/compiler/utilities.ts:3525:1 | Refused: overload 1 of nodeCanBeDecorated result true cannot be served by implementation result boolean at src/compiler/utilities.ts:3469:1 | Lowered |
| nodeIsDecorated | src/compiler/utilities.ts:3527:1 | Refused: overload 1 of nodeCanBeDecorated result true cannot be served by implementation result boolean at src/compiler/utilities.ts:3469:1 | Lowered |
| nodeIsDecorated | src/compiler/utilities.ts:3529:1 | Refused: overload 1 of nodeCanBeDecorated result true cannot be served by implementation result boolean at src/compiler/utilities.ts:3469:1 | Lowered |
| getExpandoInitializer | src/compiler/utilities.ts:3992:1 | Refused: overload 1 of skipParentheses result Expression cannot be served by implementation result Node at src/compiler/utilities.ts:5039:1 | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/transformers/destructuring.ts:192:9 |
| isCompoundLikeAssignment | src/compiler/utilities.ts:4927:1 | Refused: overload 1 of skipParentheses result Expression cannot be served by implementation result Node at src/compiler/utilities.ts:5039:1 | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/transformers/destructuring.ts:192:9 |
| skipParentheses | src/compiler/utilities.ts:5039:1 | Refused: overload 1 of skipParentheses result Expression cannot be served by implementation result Node at src/compiler/utilities.ts:5039:1 | Lowered |
| skipParentheses | src/compiler/utilities.ts:5041:1 | Refused: overload 1 of skipParentheses result Expression cannot be served by implementation result Node at src/compiler/utilities.ts:5039:1 | Lowered |
| skipParentheses | src/compiler/utilities.ts:5043:1 | Refused: overload 1 of skipParentheses result Expression cannot be served by implementation result Node at src/compiler/utilities.ts:5039:1 | Lowered |
| isAnonymousFunctionDefinition | src/compiler/utilities.ts:5483:1 | Refused: overload 1 of skipOuterExpressions result T cannot be served by implementation result Node at src/compiler/factory/utilities.ts:654:1 | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/transformers/destructuring.ts:192:9 |
| createParenthesizerRules | src/compiler/factory/parenthesizerRules.ts:54:1 | Refused: overload 1 of parenthesizeConciseBodyOfArrowFunction result Expression cannot be served by implementation result ConciseBody at src/compiler/factory/parenthesizerRules.ts:461:5 | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/transformers/destructuring.ts:192:9 |
| skipOuterExpressions | src/compiler/factory/utilities.ts:656:1 | Refused: overload 1 of skipOuterExpressions result T cannot be served by implementation result Node at src/compiler/factory/utilities.ts:654:1 | Lowered |
| skipOuterExpressions | src/compiler/factory/utilities.ts:658:1 | Refused: overload 1 of skipOuterExpressions result T cannot be served by implementation result Node at src/compiler/factory/utilities.ts:654:1 | Lowered |
| skipOuterExpressions | src/compiler/factory/utilities.ts:660:1 | Refused: overload 1 of skipOuterExpressions result T cannot be served by implementation result Node at src/compiler/factory/utilities.ts:654:1 | Lowered |
| visitFunctionBody | src/compiler/visitorPublic.ts:508:1 | Refused: overload 1 of visitFunctionBody result Block cannot be served by implementation result ConciseBody \| undefined at src/compiler/visitorPublic.ts:508:1 | NotYet: a branded supertype inside a checked predicate's array target at src/compiler/visitorPublic.ts:526:9 |
| visitFunctionBody | src/compiler/visitorPublic.ts:513:1 | Refused: overload 1 of visitFunctionBody result Block cannot be served by implementation result ConciseBody \| undefined at src/compiler/visitorPublic.ts:508:1 | NotYet: a branded supertype inside a checked predicate's array target at src/compiler/visitorPublic.ts:526:9 |
| visitFunctionBody | src/compiler/visitorPublic.ts:518:1 | Refused: overload 1 of visitFunctionBody result Block cannot be served by implementation result ConciseBody \| undefined at src/compiler/visitorPublic.ts:508:1 | NotYet: a branded supertype inside a checked predicate's array target at src/compiler/visitorPublic.ts:526:9 |
| visitFunctionBody | src/compiler/visitorPublic.ts:519:18 | Refused: overload 1 of visitFunctionBody result Block cannot be served by implementation result ConciseBody \| undefined at src/compiler/visitorPublic.ts:508:1 | NotYet: a branded supertype inside a checked predicate's array target at src/compiler/visitorPublic.ts:526:9 |
| visitFunctionBody | src/compiler/visitorPublic.ts:520:18 | Refused: overload 1 of visitFunctionBody result Block cannot be served by implementation result ConciseBody \| undefined at src/compiler/visitorPublic.ts:508:1 | NotYet: a branded supertype inside a checked predicate's array target at src/compiler/visitorPublic.ts:526:9 |
| visitFunctionBody | src/compiler/visitorPublic.ts:521:18 | Refused: overload 1 of visitFunctionBody result Block cannot be served by implementation result ConciseBody \| undefined at src/compiler/visitorPublic.ts:508:1 | NotYet: a branded supertype inside a checked predicate's array target at src/compiler/visitorPublic.ts:526:9 |
| visitFunctionBody | src/compiler/visitorPublic.ts:522:1 | Refused: overload 1 of visitFunctionBody result Block cannot be served by implementation result ConciseBody \| undefined at src/compiler/visitorPublic.ts:508:1 | NotYet: a branded supertype inside a checked predicate's array target at src/compiler/visitorPublic.ts:526:9 |
| transformES2017 | src/compiler/transformers/es2017.ts:119:1 | Refused: overload 1 of transformAsyncFunctionBody result Block cannot be served by implementation result ConciseBody at src/compiler/transformers/es2017.ts:736:5 | Refused: overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor at src/compiler/visitorPublic.ts:587:119 |
| transformES2018 | src/compiler/transformers/es2018.ts:154:1 | Refused: overload 1 of transformFunctionBody result Block cannot be served by implementation result ConciseBody at src/compiler/transformers/es2018.ts:1220:5 | NotYet: writing a possibly absent optional own field at src/compiler/factory/emitNode.ts:262:13 |

## Every regression

raw_findings: 0.
distinct_reads: 0.

## Time and peak memory

Both measurements start after setup and both builds finish. Head overlaps the already-running parent to stay within the environment lifecycle; these timings include that overlap. Each uses a separate output stream and timing wrapper. The parent census exited 0; its outer shell later hit a script-edit syntax error. The completed stream was validated and reused without a compiler rerun. See harness-events.json and overload-parent-measure-original-shell.log. See measurement-scheduling.json. Peak RSS is the child process high-water mark from Linux getrusage, in KiB (MiB below). Each Python wrapper waits for its child, writes its actual exit code and exits with that code. Times include checking, census work and stream writes. Per-unit wall/CPU streams are retained separately.

| Run | Wall seconds | User seconds | System seconds | Peak RSS MiB | Child exit |
| --- | ---: | ---: | ---: | ---: | ---: |
| parent tsc entry first-error | 2654.361 | 3048.067 | 62.116 | 3907.71 | 0 |
| head tsc entry first-error | 2813.995 | 3240.638 | 67.978 | 3865.29 | 0 |

Machine configuration is in [box.json](box.json). Setup, apply, compiler build, measurement and pairing logs are retained beside this report; streams over 1 MB are gzip compressed. [provenance.json](provenance.json) records SHAs and dependency pins; [adapted-source-hashes.json](adapted-source-hashes.json) records the shared tree bytes. Binary digests and all child timing records are retained. Final source-byte, completed-reach and archived-stream checks are in [validation.json](validation.json). Go evidence sources use `.go.txt` filenames.

## Reproduction

```sh
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh
npm ci --prefix /workspace/overload-parent/stage3/api
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/overload-upstream
bash /workspace/overload-parent/stage3/apply.sh /workspace/overload-adapted
bash build.sh parent
bash build.sh head
bash measure.sh parent
bash measure.sh head
python3 /workspace/overload-parent/stage3/census/pair/pair.py parent-latent.jsonl.gz head-latent.jsonl.gz /workspace/overload-adapted --typescript /workspace/overload-parent/stage3/api/node_modules/typescript --json delta.json --report pair-report.md
```

The helper scripts retain the exact absolute scratch paths used here. Recreate the two pinned worktrees and initialize their exact submodules first. `run.py` is the timing wrapper. This is observational lowering evidence with the latent census's existing eligibility and first-error limits, and does not produce native output.


## Newly exposed findings

The two new findings are lowering first stops in already-blocked units. They are included in the 43 changed first stops, separately from regressions:

- `src/compiler/visitorPublic.ts:526:9`: NotYet: a branded supertype inside a checked predicate's array target. First exposed in `visitFunctionBody` at `src/compiler/visitorPublic.ts:508:1`.
- `src/compiler/factory/emitNode.ts:262:13`: NotYet: writing a possibly absent optional own field. First exposed in `transformES2018` at `src/compiler/transformers/es2018.ts:154:1`.

Panics: 0 parent, 0 head. No previously finding-free attempted unit gained a first stop, and no previously finding-free refusal-scan site gained a finding.
