# Hidden bytes by roadmap step

Generated from the pinned ranking. Unplaced is included to conserve every credited byte.

| Rank | Step | Title | Hidden bytes | Reasons |
|---:|---:|---|---:|---:|
| 1 | unplaced | unplaced | 2,499,741 | 1302 |
| 2 | 30 | long tail | 394,985 | 123 |
| 3 | 22 | object semantics | 299,265 | 82 |
| 4 | 11 | branded types | 280,930 | 16 |
| 5 | 16 | generics | 106,161 | 70 |
| 6 | 17 | counted and uncounted unions | 32,976 | 113 |
| 7 | 20 | iteration | 28,342 | 6 |
| 8 | 19 | Maps and Sets with any key | 6,770 | 13 |
| 9 | 18 | optional calls | 5,095 | 1 |
| 10 | 23 | regex | 615 | 1 |
| 11 | 15 | namespaces | 0 | 0 |
| 12 | 21 | exceptions | 0 | 0 |

## Unplaced: unplaced

- 1,613,177 bytes: checker: checker-rejected body
- 129,056 bytes: NotYet: a function returning CapturedThis
- 117,867 bytes: NotYet: a value of type InitializedVariableDeclaration

## 30: long tail

- 62,495 bytes: Refused: overload 1 of writeTokenText result void cannot be served by implementation result number
- 59,499 bytes: Refused: overload 1 of visitBindingElement parameter elem cannot be served by implementation parameter elem
- 41,822 bytes: Refused: overload 1 of transformFunctionBody result Block cannot be served by implementation result ConciseBody

## 22: object semantics

- 117,289 bytes: NotYet: a method call through a structural signature in a program with statics; use typeof the declaring class
- 113,387 bytes: NotYet: a computed field name
- 34,480 bytes: Refused: a method in object destructuring

## 11: branded types

- 168,577 bytes: NotYet: a value of type __String
- 61,844 bytes: NotYet: a value of type Path
- 20,619 bytes: NotYet: a value of type ResolvedConfigFileName

## 16: generics

- 24,802 bytes: NotYet: a function returning T | undefined
- 18,006 bytes: NotYet: a function returning T
- 7,647 bytes: NotYet: a generic function as a value

## 17: counted and uncounted unions

- 6,424 bytes: NotYet: a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables
- 6,421 bytes: NotYet: a value of type string | null
- 2,295 bytes: NotYet: a value of type AccessorDeclaration & { readonly name: BigIntLiteral | ComputedPropertyName | Identifier | NoSubstitutionTemplateLiteral | NumericLiteral | StringLiteral; }

## 20: iteration

- 16,604 bytes: NotYet: for...of over an object
- 10,540 bytes: NotYet: for...in without a proven fixed plain-object origin (arrays, prototypes and absent synthetic fields cannot be enumerated soundly)
- 1,106 bytes: NotYet: a for...of destructuring an object

## 19: Maps and Sets with any key

- 2,739 bytes: NotYet: a Set of __String (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far)
- 2,391 bytes: NotYet: a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions
- 931 bytes: NotYet: a Set of Path (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far)

## 18: optional calls

- 5,095 bytes: NotYet: a call through ?. (an optional call)

## 23: regex

- 615 bytes: NotYet: RegExp with a nonconstant pattern

## 15: namespaces

No matching diagnostic reasons; zero is not evidence that this feature compiles.

## 21: exceptions

No matching diagnostic reasons; zero is not evidence that this feature compiles.

## Every unplaced reason

| Kind | Exact reason | Bytes | Rule |
|---|---|---:|---|
| NotYet | a function returning CapturedThis | 129,056 | insufficient-diagnostic |
| NotYet | a value of type InitializedVariableDeclaration | 117,867 | insufficient-diagnostic |
| NotYet | a value of type PrivateIdentifierInExpression | 91,578 | insufficient-diagnostic |
| NotYet | a function returning ImmediatelyInvokedArrowFunction | 87,684 | insufficient-diagnostic |
| NotYet | a value of type ParameterPropertyDeclaration | 59,747 | insufficient-diagnostic |
| NotYet | a value of type PrimitiveLiteral | 37,568 | insufficient-diagnostic |
| NotYet | a value of type any | 29,361 | insufficient-diagnostic |
| NotYet | an array of never | 12,123 | insufficient-diagnostic |
| NotYet | a value of type ReferencedFile & { kind: FileIncludeKind.LibReferenceDirective; } | 10,307 | insufficient-diagnostic |
| NotYet | a case whose type differs from the switch's | 7,392 | insufficient-diagnostic |
| NotYet | a function returning any | 7,290 | insufficient-diagnostic |
| NotYet | a value of type CompilerOptionsValue | 7,192 | insufficient-diagnostic |
| NotYet | reading result | 6,920 | insufficient-diagnostic |
| NotYet | an ElementAccessExpression | 6,704 | insufficient-diagnostic |
| NotYet | reading cachedPackageJson | 5,940 | insufficient-diagnostic |
| NotYet | reading options | 5,011 | insufficient-diagnostic |
| NotYet | reading left | 4,120 | insufficient-diagnostic |
| NotYet | assigning to an Identifier | 3,435 | insufficient-diagnostic |
| NotYet | a value of type ModuleResolutionState & { host: GetPackageJsonEntrypointsHost; } | 2,855 | insufficient-diagnostic |
| NotYet | reading importDeclaration | 2,788 | insufficient-diagnostic |
| NotYet | reading isPerformanceEnabled | 2,530 | insufficient-diagnostic |
| NotYet | reading kind | 2,507 | insufficient-diagnostic |
| NotYet | checked view field left of type LeftHandSideExpression | 2,419 | insufficient-diagnostic |
| NotYet | a function returning CompilerOptionsValue | 2,379 | insufficient-diagnostic |
| NotYet | reading commandLineOptions | 2,270 | insufficient-diagnostic |
| NotYet | reading declBlocked | 2,270 | insufficient-diagnostic |
| NotYet | an optional chain longer than one step | 2,229 | insufficient-diagnostic |
| NotYet | reading performance | 2,120 | insufficient-diagnostic |
| NotYet | reading host | 2,110 | insufficient-diagnostic |
| NotYet | ?. to a number, which would be number &#124; undefined | 2,099 | insufficient-diagnostic |
| NotYet | reading hostNode | 1,714 | insufficient-diagnostic |
| NotYet | reading tripleSlash | 1,699 | insufficient-diagnostic |
| NotYet | reading directoryExists | 1,647 | insufficient-diagnostic |
| NotYet | reading autoGenerate | 1,610 | insufficient-diagnostic |
| NotYet | checked view field expression of type LeftHandSideExpression | 1,583 | insufficient-diagnostic |
| NotYet | a value of type IncrementalBuildInfoFileId | 1,572 | insufficient-diagnostic |
| NotYet | reading normalized | 1,571 | insufficient-diagnostic |
| NotYet | checked view field left of type Expression | 1,554 | insufficient-diagnostic |
| NotYet | reading cache | 1,441 | insufficient-diagnostic |
| NotYet | a value of type IncludeTypeSpaceImports | 1,425 | insufficient-diagnostic |
| NotYet | reading globalCache | 1,400 | insufficient-diagnostic |
| NotYet | reading scope | 1,377 | insufficient-diagnostic |
| NotYet | a value of type ParameterDeclaration & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1,351 | insufficient-diagnostic |
| NotYet | a function returning ModeAwareCacheKey | 1,334 | insufficient-diagnostic |
| NotYet | reading existing | 1,327 | insufficient-diagnostic |
| NotYet | reading state | 1,313 | insufficient-diagnostic |
| NotYet | a function returning RedirectsCacheKey | 1,306 | insufficient-diagnostic |
| NotYet | a value of type BindingElement & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1,299 | insufficient-diagnostic |
| NotYet | reading outputFile | 1,289 | insufficient-diagnostic |
| NotYet | reading alternateResultMessage | 1,274 | insufficient-diagnostic |
| NotYet | reading expression | 1,255 | insufficient-diagnostic |
| NotYet | reading updated | 1,241 | insufficient-diagnostic |
| NotYet | checked view field expression of type Expression | 1,199 | insufficient-diagnostic |
| NotYet | reading parseTreeNode | 1,137 | insufficient-diagnostic |
| NotYet | a value of type VariableDeclaration & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1,115 | insufficient-diagnostic |
| NotYet | a tuple literal leaving out an element of type ModuleSpecifierEnding | 1,089 | insufficient-diagnostic |
| NotYet | reading specifiers | 1,078 | insufficient-diagnostic |
| NotYet | reading statement | 1,077 | insufficient-diagnostic |
| NotYet | reading str | 1,070 | insufficient-diagnostic |
| NotYet | reading text | 1,054 | insufficient-diagnostic |
| NotYet | reading configParseResult | 1,041 | insufficient-diagnostic |
| NotYet | reading diag | 1,032 | insufficient-diagnostic |
| NotYet | reading declarationFile | 1,009 | insufficient-diagnostic |
| NotYet | checked view field left of type EntityName | 1,005 | insufficient-diagnostic |
| NotYet | reading leftThisArg | 1,005 | insufficient-diagnostic |
| NotYet | a value of type WrappedExpression<AnonymousFunctionDefinition> | 992 | insufficient-diagnostic |
| NotYet | reading assignClassAliasInStaticBlock | 976 | insufficient-diagnostic |
| NotYet | reading parent | 957 | insufficient-diagnostic |
| NotYet | reading currentWriterIndentSpacing | 946 | insufficient-diagnostic |
| NotYet | reading valueDeclaration | 937 | insufficient-diagnostic |
| NotYet | a value of type ShorthandPropertyAssignment & { readonly objectAssignmentInitializer: WrappedExpression<AnonymousFunctionDefinition>; } | 933 | insufficient-diagnostic |
| NotYet | reading expr | 931 | insufficient-diagnostic |
| NotYet | a value of type ExportAssignment & { readonly expression: WrappedExpression<AnonymousFunctionDefinition>; } | 913 | insufficient-diagnostic |
| NotYet | reading spacesToEmit | 900 | insufficient-diagnostic |
| NotYet | reading initializer | 879 | insufficient-diagnostic |
| NotYet | reading nodeModulesFolderExists | 877 | insufficient-diagnostic |
| NotYet | a call spreading elements that cannot be packed | 848 | insufficient-diagnostic |
| NotYet | reading configFile | 833 | insufficient-diagnostic |
| NotYet | a value of type never | 823 | insufficient-diagnostic |
| NotYet | reading elem | 814 | insufficient-diagnostic |
| NotYet | a value of type PropertyDeclaration & { readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 809 | insufficient-diagnostic |
| NotYet | reading buildOptions | 795 | insufficient-diagnostic |
| NotYet | reading v | 787 | insufficient-diagnostic |
| NotYet | checked view field parent of type ObjectLiteralExpression | 764 | insufficient-diagnostic |
| NotYet | a value of type PropertyAssignment & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 751 | insufficient-diagnostic |
| NotYet | a value of type SourceFile | 751 | insufficient-diagnostic |
| NotYet | a function returning ClassNamedEvaluationHelperBlock | 728 | insufficient-diagnostic |
| NotYet | reading alternateResult | 721 | insufficient-diagnostic |
| NotYet | checked view field parent of type SignatureDeclaration | 718 | insufficient-diagnostic |
| NotYet | reading ext | 706 | insufficient-diagnostic |
| NotYet | checked view field parent of type ClassLikeDeclaration | 702 | insufficient-diagnostic |
| NotYet | a function returning CanonicalKey | 681 | insufficient-diagnostic |
| NotYet | reading file | 673 | insufficient-diagnostic |
| NotYet | reading conditionSets | 669 | insufficient-diagnostic |
| NotYet | reading encodeURI | 654 | insufficient-diagnostic |
| NotYet | reading transform | 643 | insufficient-diagnostic |
| NotYet | reading tsPriority | 640 | insufficient-diagnostic |
| NotYet | reading token | 636 | insufficient-diagnostic |
| NotYet | reading exportSpecifiers | 633 | insufficient-diagnostic |
| NotYet | a function returning ClassThisAssignmentBlock | 626 | insufficient-diagnostic |
| NotYet | reading memberName | 618 | insufficient-diagnostic |
| NotYet | reading filesForEmit | 614 | insufficient-diagnostic |
| NotYet | reading length | 607 | insufficient-diagnostic |
| NotYet | reading ts | 600 | insufficient-diagnostic |
| NotYet | reading lineNumber | 599 | insufficient-diagnostic |
| NotYet | checked view field escapedText of type __String | 596 | insufficient-diagnostic |
| NotYet | reading flags | 594 | insufficient-diagnostic |
| NotYet | reading objectFlags | 593 | insufficient-diagnostic |
| NotYet | reading aParts | 578 | insufficient-diagnostic |
| NotYet | reading filesInError | 562 | insufficient-diagnostic |
| NotYet | reading terminalWidth | 550 | insufficient-diagnostic |
| NotYet | reading resolvedProject | 540 | insufficient-diagnostic |
| NotYet | reading temp | 524 | insufficient-diagnostic |
| NotYet | reading originalCreateDirectory | 520 | insufficient-diagnostic |
| NotYet | reading parseNode | 516 | insufficient-diagnostic |
| NotYet | reading containingDirectory | 512 | insufficient-diagnostic |
| NotYet | a tuple element of type string &#124; number &#124; boolean &#124; readonly string[] &#124; SourceFile &#124; undefined | 509 | insufficient-diagnostic |
| NotYet | reading path | 508 | insufficient-diagnostic |
| NotYet | reading fromComponents | 496 | insufficient-diagnostic |
| NotYet | a value of type WatchFactoryHost & { trace?(s: string): void; } | 474 | insufficient-diagnostic |
| NotYet | reading minor | 474 | insufficient-diagnostic |
| NotYet | reading diff | 473 | insufficient-diagnostic |
| NotYet | reading firstNonzeroSegment | 471 | insufficient-diagnostic |
| NotYet | reading rightExpression | 467 | insufficient-diagnostic |
| NotYet | reading commonSourceDirectory | 466 | insufficient-diagnostic |
| NotYet | reading members | 462 | insufficient-diagnostic |
| NotYet | reading suggestion | 456 | insufficient-diagnostic |
| NotYet | reading value | 453 | insufficient-diagnostic |
| NotYet | assigning a field of a value | 446 | insufficient-diagnostic |
| NotYet | assigning to an ObjectLiteralExpression | 446 | insufficient-diagnostic |
| NotYet | reading leadingComments | 439 | insufficient-diagnostic |
| NotYet | a value of type IncrementalMultiFileEmitBuildInfoFileInfo | 434 | insufficient-diagnostic |
| NotYet | reading compareResult | 434 | insufficient-diagnostic |
| NotYet | reading next | 433 | insufficient-diagnostic |
| NotYet | reading onlyRecordFailuresForPackageFile | 428 | insufficient-diagnostic |
| NotYet | reading existingFlags | 424 | insufficient-diagnostic |
| NotYet | reading mapping | 422 | insufficient-diagnostic |
| NotYet | reading declarationTransform | 398 | insufficient-diagnostic |
| NotYet | reading loadPackageJsonMainState | 395 | insufficient-diagnostic |
| NotYet | reading entityName | 392 | insufficient-diagnostic |
| NotYet | reading watcher | 391 | insufficient-diagnostic |
| NotYet | reading jsxImportSourcePragma | 387 | insufficient-diagnostic |
| NotYet | reading commandLine | 384 | insufficient-diagnostic |
| NotYet | reading commonResolved | 383 | insufficient-diagnostic |
| NotYet | reading numParameters | 382 | insufficient-diagnostic |
| NotYet | a function returning T1 & T2 | 374 | insufficient-diagnostic |
| NotYet | reading otherOption | 374 | insufficient-diagnostic |
| NotYet | reading operand | 371 | insufficient-diagnostic |
| NotYet | reading addUndefined | 370 | insufficient-diagnostic |
| NotYet | reading recursiveKeys | 366 | insufficient-diagnostic |
| NotYet | reading emitNode | 360 | insufficient-diagnostic |
| NotYet | reading resolvedFromFile | 358 | insufficient-diagnostic |
| NotYet | reading getCanonicalFileName | 357 | insufficient-diagnostic |
| NotYet | reading externalHelpersImportDeclaration | 354 | insufficient-diagnostic |
| NotYet | reading start | 352 | insufficient-diagnostic |
| NotYet | reading suffix | 345 | insufficient-diagnostic |
| NotYet | reading hooks | 344 | insufficient-diagnostic |
| NotYet | reading firstArgument | 343 | insufficient-diagnostic |
| NotYet | optional chaining to .size on a value | 341 | insufficient-diagnostic |
| NotYet | reading gutterWidth | 338 | insufficient-diagnostic |
| NotYet | a call returning any | 336 | insufficient-diagnostic |
| NotYet | reading trailingComments | 334 | insufficient-diagnostic |
| NotYet | reading compilerOptions | 332 | insufficient-diagnostic |
| NotYet | reading ownOutputFilePath | 320 | insufficient-diagnostic |
| NotYet | reading lastDecorator | 317 | insufficient-diagnostic |
| NotYet | reading caches | 316 | insufficient-diagnostic |
| NotYet | reading mainExport | 313 | insufficient-diagnostic |
| NotYet | reading checkFlags | 312 | insufficient-diagnostic |
| NotYet | reading helper | 312 | insufficient-diagnostic |
| NotYet | reading e | 310 | insufficient-diagnostic |
| NotYet | reading heritageClause | 310 | insufficient-diagnostic |
| NotYet | assigning an element of a value | 307 | insufficient-diagnostic |
| NotYet | checked view field type of type TypeNode | 307 | insufficient-diagnostic |
| NotYet | reading helpers | 305 | insufficient-diagnostic |
| NotYet | reading assignedName | 304 | insufficient-diagnostic |
| NotYet | reading rhsValue | 302 | insufficient-diagnostic |
| NotYet | reading target | 297 | insufficient-diagnostic |
| NotYet | reading map | 293 | insufficient-diagnostic |
| NotYet | reading container | 288 | insufficient-diagnostic |
| NotYet | reading helpOptions | 284 | insufficient-diagnostic |
| NotYet | reading capturedLeft | 282 | insufficient-diagnostic |
| NotYet | reading parentComponents | 281 | insufficient-diagnostic |
| NotYet | .length on a value | 280 | insufficient-diagnostic |
| NotYet | reading classThis | 279 | insufficient-diagnostic |
| NotYet | reading sourceFile | 278 | insufficient-diagnostic |
| NotYet | reading getCommonSourceDirectory | 276 | insufficient-diagnostic |
| NotYet | a value of type HasJSDoc | 271 | insufficient-diagnostic |
| NotYet | reading declarations | 269 | insufficient-diagnostic |
| NotYet | reading newParsedCommandLine | 265 | insufficient-diagnostic |
| NotYet | reading patterns | 264 | insufficient-diagnostic |
| NotYet | reading currentDirectory | 263 | insufficient-diagnostic |
| NotYet | a value of type TypeParameterDeclaration & { parent: JSDocTemplateTag; } | 257 | insufficient-diagnostic |
| NotYet | checked view field label of type Identifier | 257 | insufficient-diagnostic |
| NotYet | a number &#124; undefined argument to substring | 255 | insufficient-diagnostic |
| NotYet | reading candidateExists | 252 | insufficient-diagnostic |
| NotYet | reading originalReadFile | 251 | insufficient-diagnostic |
| NotYet | reading candidate | 248 | insufficient-diagnostic |
| NotYet | reading statements | 247 | insufficient-diagnostic |
| NotYet | checked view field parent of type NamedImports | 244 | insufficient-diagnostic |
| NotYet | reading directory | 239 | insufficient-diagnostic |
| NotYet | a call spreading something other than an array | 238 | insufficient-diagnostic |
| NotYet | reading decorationStatements | 238 | insufficient-diagnostic |
| NotYet | reading buildInfo | 235 | insufficient-diagnostic |
| NotYet | reading tokenSourceMapRanges | 235 | insufficient-diagnostic |
| NotYet | reading cachedDiagnostics | 228 | insufficient-diagnostic |
| NotYet | reading jsFilePath | 228 | insufficient-diagnostic |
| NotYet | reading index | 227 | insufficient-diagnostic |
| NotYet | reading internalFlags | 223 | insufficient-diagnostic |
| NotYet | reading blockedByExports | 222 | insufficient-diagnostic |
| NotYet | reading nodeModulesAtTypesExists | 218 | insufficient-diagnostic |
| NotYet | reading sourceMappings | 218 | insufficient-diagnostic |
| NotYet | a value of type BindableStaticNameExpression | 215 | insufficient-diagnostic |
| NotYet | reading sharedLength | 215 | insufficient-diagnostic |
| NotYet | reading directoryStart | 214 | insufficient-diagnostic |
| NotYet | a value of type ResolvedModuleWithFailedLookupLocations & ResolvedTypeReferenceDirectiveWithFailedLookupLocations | 212 | insufficient-diagnostic |
| NotYet | reading react | 209 | insufficient-diagnostic |
| NotYet | reading configFileText | 208 | insufficient-diagnostic |
| NotYet | reading extensionlessPriority | 205 | insufficient-diagnostic |
| NotYet | Array.isArray on a tuple or an erased object/any/unknown view | 204 | insufficient-diagnostic |
| NotYet | reading filePath | 204 | insufficient-diagnostic |
| NotYet | reading exportContainer | 203 | insufficient-diagnostic |
| NotYet | checked view field elements of type NodeArray<Expression> | 202 | insufficient-diagnostic |
| NotYet | reading rawText | 198 | insufficient-diagnostic |
| NotYet | reading varStatement | 198 | insufficient-diagnostic |
| NotYet | reading i | 196 | insufficient-diagnostic |
| NotYet | reading pragma | 192 | insufficient-diagnostic |
| NotYet | reading includeRe | 191 | insufficient-diagnostic |
| NotYet | reading buildOrder | 190 | insufficient-diagnostic |
| NotYet | reading right | 190 | insufficient-diagnostic |
| NotYet | reading pos | 189 | insufficient-diagnostic |
| NotYet | checked view field parent of type NamedExports | 184 | insufficient-diagnostic |
| NotYet | checked view field operand of type UnaryExpression | 183 | insufficient-diagnostic |
| NotYet | reading isDosStyle | 182 | insufficient-diagnostic |
| NotYet | reading resolvedFileName | 176 | insufficient-diagnostic |
| NotYet | checked view field checkType of type TypeNode | 175 | insufficient-diagnostic |
| NotYet | reading args | 175 | insufficient-diagnostic |
| NotYet | reading argument | 175 | insufficient-diagnostic |
| NotYet | reading res | 174 | insufficient-diagnostic |
| NotYet | reading mapped | 173 | insufficient-diagnostic |
| NotYet | reading memoryUsed | 172 | insufficient-diagnostic |
| NotYet | reading exportName | 170 | insufficient-diagnostic |
| NotYet | reading packageRootPath | 170 | insufficient-diagnostic |
| NotYet | reading arg | 165 | insufficient-diagnostic |
| NotYet | reading firstParameterIsThis | 165 | insufficient-diagnostic |
| NotYet | a boolean &#124; undefined variable a function value captures | 164 | insufficient-diagnostic |
| NotYet | reading relativeFileName | 163 | insufficient-diagnostic |
| NotYet | reading packageDirectoryExists | 162 | insufficient-diagnostic |
| NotYet | reading extensions | 161 | insufficient-diagnostic |
| NotYet | reading startsOnNewLine | 161 | insufficient-diagnostic |
| NotYet | reading oldEmitKind | 159 | insufficient-diagnostic |
| NotYet | checked view field typeName of type EntityName | 157 | insufficient-diagnostic |
| NotYet | reading snippetElement | 157 | insufficient-diagnostic |
| NotYet | checked view field parent of type BindingPattern | 156 | insufficient-diagnostic |
| NotYet | reading hasArguments | 155 | insufficient-diagnostic |
| NotYet | reading targetIndex | 155 | insufficient-diagnostic |
| NotYet | reading constantValue | 153 | insufficient-diagnostic |
| NotYet | reading jsxImportSourcePragmas | 153 | insufficient-diagnostic |
| NotYet | reading shouldWriteNativeEvents | 153 | insufficient-diagnostic |
| NotYet | a function returning this | 151 | insufficient-diagnostic |
| NotYet | reading diagnostics | 151 | insufficient-diagnostic |
| NotYet | reading dirPath | 151 | insufficient-diagnostic |
| NotYet | reading name | 151 | insufficient-diagnostic |
| NotYet | reading parameters | 151 | insufficient-diagnostic |
| NotYet | a value of type T1 | 149 | insufficient-diagnostic |
| NotYet | reading nextKey | 147 | insufficient-diagnostic |
| NotYet | reading aComponents | 143 | insufficient-diagnostic |
| NotYet | reading lastPart | 143 | insufficient-diagnostic |
| NotYet | reading sourceMapRange | 143 | insufficient-diagnostic |
| NotYet | reading declarationFilePath | 141 | insufficient-diagnostic |
| NotYet | reading existingSpecifier | 141 | insufficient-diagnostic |
| NotYet | reading queue | 141 | insufficient-diagnostic |
| Refused | a type argument makes a value of type (left: MappedPosition, right: MappedPosition) => boolean seen as EqualityComparer<SourceMappedPosition> &#124; undefined, which can write string &#124; undefined where string is read | 141 | insufficient-diagnostic |
| NotYet | reading leftmost | 137 | insufficient-diagnostic |
| NotYet | reading commentRange | 135 | insufficient-diagnostic |
| NotYet | reading decorator | 135 | insufficient-diagnostic |
| NotYet | reading line | 133 | insufficient-diagnostic |
| NotYet | reading dir | 132 | insufficient-diagnostic |
| NotYet | checked view field parent of type TemplateExpression | 131 | insufficient-diagnostic |
| NotYet | reading annotatedNodes | 131 | insufficient-diagnostic |
| NotYet | reading directoryPath | 130 | insufficient-diagnostic |
| NotYet | checked view field parent of type CaseBlock | 128 | insufficient-diagnostic |
| NotYet | reading jsxRuntimePragmas | 128 | insufficient-diagnostic |
| NotYet | reading sourceMap | 127 | insufficient-diagnostic |
| NotYet | checked view field parent of type TryStatement | 126 | insufficient-diagnostic |
| NotYet | reading parameter | 123 | insufficient-diagnostic |
| NotYet | reading last | 122 | insufficient-diagnostic |
| NotYet | reading leadingCommentRanges | 122 | insufficient-diagnostic |
| NotYet | reading refPath | 121 | insufficient-diagnostic |
| NotYet | reading singleLine | 121 | insufficient-diagnostic |
| NotYet | reading current | 119 | insufficient-diagnostic |
| NotYet | reading trailingParts | 119 | insufficient-diagnostic |
| NotYet | an array of unknown with erased element storage (retain its declared element type before reading elements) | 118 | insufficient-diagnostic |
| NotYet | reading sourceIndex | 118 | insufficient-diagnostic |
| NotYet | reading existingPath | 115 | insufficient-diagnostic |
| NotYet | reading remainingPaths | 115 | insufficient-diagnostic |
| NotYet | reading basePath | 112 | insufficient-diagnostic |
| NotYet | checked view field _optionalChainBrand of type void | 110 | insufficient-diagnostic |
| NotYet | checked view field parent of type SwitchStatement | 110 | insufficient-diagnostic |
| NotYet | reading d | 110 | insufficient-diagnostic |
| NotYet | reading verbatimFromExports | 108 | insufficient-diagnostic |
| NotYet | reading optionsOfCurCategory | 105 | insufficient-diagnostic |
| NotYet | reading moduleResolutionState | 104 | insufficient-diagnostic |
| NotYet | reading parts | 104 | insufficient-diagnostic |
| NotYet | reading excludeRegex | 102 | insufficient-diagnostic |
| NotYet | reading sourceMapUrlPos | 102 | insufficient-diagnostic |
| NotYet | reading classExpression | 101 | insufficient-diagnostic |
| NotYet | reading cached | 100 | insufficient-diagnostic |
| NotYet | reading content | 100 | insufficient-diagnostic |
| NotYet | reading real | 100 | insufficient-diagnostic |
| NotYet | reading thisParameter | 100 | insufficient-diagnostic |
| NotYet | a function returning AnyValidImportOrReExport | 96 | insufficient-diagnostic |
| NotYet | checked view field parent of type ImportClause | 95 | insufficient-diagnostic |
| NotYet | checked view field typeExpression of type JSDocTypeExpression | 95 | insufficient-diagnostic |
| NotYet | a number &#124; undefined argument to slice | 94 | insufficient-diagnostic |
| NotYet | reading onlyRecordFailuresForIndex | 94 | insufficient-diagnostic |
| NotYet | reading targetNode | 94 | insufficient-diagnostic |
| NotYet | reading resolved | 93 | insufficient-diagnostic |
| NotYet | reading searchPath | 92 | insufficient-diagnostic |
| NotYet | checked predicate overload target Expression | 91 | insufficient-diagnostic |
| NotYet | reading argumentsList | 91 | insufficient-diagnostic |
| NotYet | reading childComponents | 91 | insufficient-diagnostic |
| NotYet | reading useCaseSensitiveFileNames | 91 | insufficient-diagnostic |
| NotYet | reading nameParts | 90 | insufficient-diagnostic |
| NotYet | a value of type IncrementalBuildInfoFilePendingEmit | 89 | insufficient-diagnostic |
| NotYet | reading exportedNamesStorageRef | 89 | insufficient-diagnostic |
| NotYet | reading ambientModuleDeclare | 88 | insufficient-diagnostic |
| NotYet | a value of type TypeNode & LiteralTypeNode & { readonly literal: StringLiteral; } | 87 | insufficient-diagnostic |
| NotYet | reading constructor | 87 | insufficient-diagnostic |
| NotYet | reading exportStarFunction | 87 | insufficient-diagnostic |
| NotYet | reading extensionless | 86 | insufficient-diagnostic |
| NotYet | reading visited | 86 | insufficient-diagnostic |
| NotYet | reading constructorDeclaration | 84 | insufficient-diagnostic |
| NotYet | reading parentDirectory | 84 | insufficient-diagnostic |
| NotYet | reading newDirectory | 83 | insufficient-diagnostic |
| NotYet | reading toComponents | 83 | insufficient-diagnostic |
| NotYet | reading pattern | 82 | insufficient-diagnostic |
| NotYet | assigning to a NonNullExpression | 81 | insufficient-diagnostic |
| NotYet | reading impliedNodeFormat | 79 | insufficient-diagnostic |
| NotYet | reading decorators | 78 | insufficient-diagnostic |
| NotYet | reading prerelease | 77 | insufficient-diagnostic |
| NotYet | checked view field parent of type TemplateLiteralTypeNode | 76 | insufficient-diagnostic |
| NotYet | reading metadata | 76 | insufficient-diagnostic |
| NotYet | reading operation | 76 | insufficient-diagnostic |
| NotYet | reading packageResult | 76 | insufficient-diagnostic |
| NotYet | reading base64SourceMapText | 75 | insufficient-diagnostic |
| NotYet | reading containingDirectoryPath | 75 | insufficient-diagnostic |
| NotYet | reading excludeRe | 75 | insufficient-diagnostic |
| NotYet | reading sourceMapText | 75 | insufficient-diagnostic |
| NotYet | checked view field parent of type Declaration | 74 | insufficient-diagnostic |
| NotYet | reading fromPaths | 74 | insufficient-diagnostic |
| NotYet | reading key | 74 | insufficient-diagnostic |
| NotYet | checked view field expression of type UnaryExpression | 73 | insufficient-diagnostic |
| NotYet | checked view field parent of type SourceFile | 73 | insufficient-diagnostic |
| NotYet | checked view field name of type Identifier | 70 | insufficient-diagnostic |
| NotYet | checked view field initializer of type ForInitializer | 69 | insufficient-diagnostic |
| NotYet | reading fileRef | 69 | insufficient-diagnostic |
| NotYet | reading ownedTags | 68 | insufficient-diagnostic |
| NotYet | reading typeAlias | 68 | insufficient-diagnostic |
| NotYet | checked view field condition of type Expression | 67 | insufficient-diagnostic |
| NotYet | checked view field operand of type LeftHandSideExpression | 66 | insufficient-diagnostic |
| NotYet | checked view field parent of type ExportDeclaration | 65 | insufficient-diagnostic |
| NotYet | checked view field tryBlock of type Block | 65 | insufficient-diagnostic |
| NotYet | checked view field tag of type LeftHandSideExpression | 64 | insufficient-diagnostic |
| NotYet | reading comparer | 63 | insufficient-diagnostic |
| NotYet | reading compilerHost | 63 | insufficient-diagnostic |
| NotYet | checked view field argument of type TypeNode | 62 | insufficient-diagnostic |
| NotYet | reading build | 62 | insufficient-diagnostic |
| NotYet | reading cloned | 62 | insufficient-diagnostic |
| NotYet | reading declaration | 62 | insufficient-diagnostic |
| NotYet | reading mainResolution | 62 | insufficient-diagnostic |
| NotYet | reading message | 62 | insufficient-diagnostic |
| NotYet | reading negative | 62 | insufficient-diagnostic |
| NotYet | reading relativePath | 62 | insufficient-diagnostic |
| NotYet | reading changed | 61 | insufficient-diagnostic |
| NotYet | reading objectType | 61 | insufficient-diagnostic |
| NotYet | reading previous | 60 | insufficient-diagnostic |
| NotYet | reading newItem | 59 | insufficient-diagnostic |
| NotYet | ?.[] on a value | 58 | insufficient-diagnostic |
| NotYet | reading body | 57 | insufficient-diagnostic |
| NotYet | a value of type UnaryExpression & NumericLiteral | 56 | insufficient-diagnostic |
| NotYet | reading decl | 56 | insufficient-diagnostic |
| NotYet | reading imports | 55 | insufficient-diagnostic |
| NotYet | reading pathComponents | 55 | insufficient-diagnostic |
| NotYet | storing Path &#124; undefined in a field | 55 | insufficient-diagnostic |
| NotYet | storing any in a field | 55 | insufficient-diagnostic |
| NotYet | reading packageFileResult | 54 | insufficient-diagnostic |
| NotYet | a value of type CanonicalKey | 53 | insufficient-diagnostic |
| NotYet | a value of type EmitNode & { autoGenerate: AutoGenerateInfo; } | 53 | insufficient-diagnostic |
| NotYet | checked view field elementType of type TypeNode | 53 | insufficient-diagnostic |
| NotYet | reading segments | 52 | insufficient-diagnostic |
| NotYet | checked view field parent of type NamedDeclaration | 51 | insufficient-diagnostic |
| NotYet | checked view field statements of type NodeArray<Statement> | 51 | insufficient-diagnostic |
| NotYet | a value of type __String & string | 50 | insufficient-diagnostic |
| NotYet | reading ancestor | 50 | insufficient-diagnostic |
| NotYet | reading readFileWithCache | 50 | insufficient-diagnostic |
| NotYet | reading accessModifier | 49 | insufficient-diagnostic |
| NotYet | reading setReadFileCache | 49 | insufficient-diagnostic |
| NotYet | reading type | 49 | insufficient-diagnostic |
| NotYet | an array of any | 48 | insufficient-diagnostic |
| NotYet | checked view field _children of type readonly Node[] | 48 | insufficient-diagnostic |
| NotYet | reading version | 48 | insufficient-diagnostic |
| NotYet | a value of type RequireOrImportCall | 47 | insufficient-diagnostic |
| NotYet | reading referencedFileName | 47 | insufficient-diagnostic |
| NotYet | reading components | 46 | insufficient-diagnostic |
| NotYet | reading dist | 46 | insufficient-diagnostic |
| NotYet | reading min | 46 | insufficient-diagnostic |
| NotYet | reading returnStatement | 46 | insufficient-diagnostic |
| NotYet | reading tempVar | 46 | insufficient-diagnostic |
| NotYet | a value of type CompilerHost & ReadBuildProgramHost | 45 | insufficient-diagnostic |
| NotYet | reading mappings | 45 | insufficient-diagnostic |
| NotYet | reading helperCall | 44 | insufficient-diagnostic |
| NotYet | reading varDecl | 41 | insufficient-diagnostic |
| NotYet | reading data | 39 | insufficient-diagnostic |
| NotYet | reading jsonText | 38 | insufficient-diagnostic |
| NotYet | reading sig | 37 | insufficient-diagnostic |
| NotYet | reading subStitution | 36 | insufficient-diagnostic |
| NotYet | reading moduleFileToTry | 34 | insufficient-diagnostic |
| NotYet | reading Debug | 31 | insufficient-diagnostic |
| NotYet | reading newValue | 30 | insufficient-diagnostic |
| NotYet | reading errorMessage | 28 | insufficient-diagnostic |
| NotYet | reading exitStatus | 24 | insufficient-diagnostic |
| NotYet | reading tag | 24 | insufficient-diagnostic |
| NotYet | reading values | 22 | insufficient-diagnostic |
| NotYet | .hasTrailingComma on a value | 0 | insufficient-diagnostic |
| NotYet | JSON.stringify a union containing containers without runtime element metadata | 0 | insufficient-diagnostic |
| NotYet | a [node: Node] seen as a Node (a tuple is held as an object, not an array, so far; write it as an array where it's made, or copy it into one: [pair[0], pair[1]]) | 0 | insufficient-diagnostic |
| NotYet | a call to a PropertyAccessExpression | 0 | insufficient-diagnostic |
| NotYet | a call to an Identifier | 0 | insufficient-diagnostic |
| NotYet | a function returning A | 0 | insufficient-diagnostic |
| NotYet | a function returning BindableAccessExpression | 0 | insufficient-diagnostic |
| NotYet | a function returning BindableStaticAccessExpression | 0 | insufficient-diagnostic |
| NotYet | a function returning Declaration & HasModifiers | 0 | insufficient-diagnostic |
| NotYet | a function returning IncrementalBuildInfoFileId | 0 | insufficient-diagnostic |
| NotYet | a function returning IncrementalBuildInfoFileIdListId | 0 | insufficient-diagnostic |
| NotYet | a function returning ReusableDiagnosticMessageChain | 0 | insufficient-diagnostic |
| NotYet | a function returning SyntheticSuper | 0 | insufficient-diagnostic |
| NotYet | a library method value outside a const alias, typed call/apply, or supported map callback (its receiver and callable ABI are not proven); wrap the call in an arrow | 0 | insufficient-diagnostic |
| NotYet | a value of type AliasDeclarationNode | 0 | insufficient-diagnostic |
| NotYet | a value of type AnonymousFunctionDefinition | 0 | insufficient-diagnostic |
| NotYet | a value of type BindableAccessExpression | 0 | insufficient-diagnostic |
| NotYet | a value of type BindableObjectDefinePropertyCall | 0 | insufficient-diagnostic |
| NotYet | a value of type BindableStaticAccessExpression | 0 | insufficient-diagnostic |
| NotYet | a value of type Canonicalized | 0 | insufficient-diagnostic |
| NotYet | a value of type Child | 0 | insufficient-diagnostic |
| NotYet | a value of type Declaration & Expression | 0 | insufficient-diagnostic |
| NotYet | a value of type ElementWithComputedPropertyName | 0 | insufficient-diagnostic |
| NotYet | a value of type EndOfFileToken | 0 | insufficient-diagnostic |
| NotYet | a value of type ExportDeclaration & { readonly exportClause: NamedExports; } | 0 | insufficient-diagnostic |
| NotYet | a value of type Expression & Declaration | 0 | insufficient-diagnostic |
| NotYet | a value of type ImmediatelyInvokedArrowFunction | 0 | insufficient-diagnostic |
| NotYet | a value of type ImportEqualsDeclaration & { moduleReference: ExternalModuleReference; } | 0 | insufficient-diagnostic |
| NotYet | a value of type IncrementalBuildInfoFileIdListId | 0 | insufficient-diagnostic |
| NotYet | a value of type LeftHandSideExpression & Identifier | 0 | insufficient-diagnostic |
| NotYet | a value of type ModeAwareCacheKey | 0 | insufficient-diagnostic |
| NotYet | a value of type NamedDeclaration & { name: DeclarationName; } | 0 | insufficient-diagnostic |
| NotYet | a value of type RedirectsCacheKey | 0 | insufficient-diagnostic |
| NotYet | a value of type ReplaceableIndexedAccessType | 0 | insufficient-diagnostic |
| NotYet | a value of type ReusableDiagnosticMessageChain | 0 | insufficient-diagnostic |
| NotYet | a value of type Source | 0 | insufficient-diagnostic |
| NotYet | a value of type SourceFileOrString | 0 | insufficient-diagnostic |
| NotYet | a value of type ThisCapturingVariableDeclaration | 0 | insufficient-diagnostic |
| NotYet | a value of type TransformedSuperCall | 0 | insufficient-diagnostic |
| NotYet | a value of type VariableDeclaration & { name: Identifier; } | 0 | insufficient-diagnostic |
| NotYet | a value of type VariableDeclarationList & { _usingBrand: void; } | 0 | insufficient-diagnostic |
| NotYet | an array of Child | 0 | insufficient-diagnostic |
| NotYet | an array of ElementWithComputedPropertyName | 0 | insufficient-diagnostic |
| NotYet | an array of InitializedPropertyDeclaration | 0 | insufficient-diagnostic |
| NotYet | an array of ParameterPropertyDeclaration | 0 | insufficient-diagnostic |
| NotYet | an array of ReusableDiagnosticMessageChain | 0 | insufficient-diagnostic |
| NotYet | checked predicate overload target CallExpression & { expression: Identifier; arguments: [StringLiteralLike]; } & { expression: Identifier; arguments: ...; } | 0 | insufficient-diagnostic |
| NotYet | checked predicate overload target IterationStatement | 0 | insufficient-diagnostic |
| NotYet | checked predicate overload target Statement | 0 | insufficient-diagnostic |
| NotYet | checked predicate overload target Type | 0 | insufficient-diagnostic |
| NotYet | checked predicate overload target readonly Decorator[] | 0 | insufficient-diagnostic |
| NotYet | checked predicate overload target readonly Modifier[] | 0 | insufficient-diagnostic |
| NotYet | checked predicate overload target readonly ShorthandPropertyAssignment[] | 0 | insufficient-diagnostic |
| NotYet | checked predicate overload target readonly TypeNode[] | 0 | insufficient-diagnostic |
| NotYet | checked view field exprName of type EntityName | 0 | insufficient-diagnostic |
| NotYet | checked view field expression of type JsonObjectExpression | 0 | insufficient-diagnostic |
| NotYet | checked view field expression of type StringLiteral | 0 | insufficient-diagnostic |
| NotYet | checked view field head of type TemplateHead | 0 | insufficient-diagnostic |
| NotYet | checked view field left of type BindableAccessExpression | 0 | insufficient-diagnostic |
| NotYet | checked view field left of type BindableStaticAccessExpression | 0 | insufficient-diagnostic |
| NotYet | checked view field members of type NodeArray<TypeElement> | 0 | insufficient-diagnostic |
| NotYet | checked view field modifiers of type undefined | 0 | insufficient-diagnostic |
| NotYet | checked view field objectType of type TypeNode | 0 | insufficient-diagnostic |
| NotYet | checked view field openingElement of type JsxOpeningElement | 0 | insufficient-diagnostic |
| NotYet | checked view field openingFragment of type JsxOpeningFragment | 0 | insufficient-diagnostic |
| NotYet | checked view field operatorToken of type Token<SyntaxKind.InstanceOfKeyword> | 0 | insufficient-diagnostic |
| NotYet | checked view field parent of type EnumDeclaration | 0 | insufficient-diagnostic |
| NotYet | checked view field parent of type HasJSDoc | 0 | insufficient-diagnostic |
| NotYet | checked view field parent of type ImportAttributes | 0 | insufficient-diagnostic |
| NotYet | checked view field parent of type ImportEqualsDeclaration | 0 | insufficient-diagnostic |
| NotYet | checked view field parent of type JSDoc | 0 | insufficient-diagnostic |
| NotYet | checked view field parent of type JsxAttributes | 0 | insufficient-diagnostic |
| NotYet | checked view field parent of type JsxElement | 0 | insufficient-diagnostic |
| NotYet | checked view field parent of type ModuleDeclaration | 0 | insufficient-diagnostic |
| NotYet | checked view field parent of type ObjectTypeDeclaration | 0 | insufficient-diagnostic |
| NotYet | checked view field properties of type NodeArray<JsxAttributeLike> | 0 | insufficient-diagnostic |
| NotYet | checked view field sourceFiles of type readonly SourceFile[] | 0 | insufficient-diagnostic |
| NotYet | checked view field tagName of type JsxTagNameExpression | 0 | insufficient-diagnostic |
| NotYet | checked view field type of type Type | 0 | insufficient-diagnostic |
| NotYet | checked view field typeParameter of type TypeParameterDeclaration | 0 | insufficient-diagnostic |
| NotYet | checked view field types of type NodeArray<TypeNode> | 0 | insufficient-diagnostic |
| NotYet | reading absoluteSourceFilePath | 0 | insufficient-diagnostic |
| NotYet | reading abstractSignatures | 0 | insufficient-diagnostic |
| NotYet | reading accessorDeclarations | 0 | insufficient-diagnostic |
| NotYet | reading accessorType | 0 | insufficient-diagnostic |
| NotYet | reading activeLabel | 0 | insufficient-diagnostic |
| NotYet | reading actualFileName | 0 | insufficient-diagnostic |
| NotYet | reading add | 0 | insufficient-diagnostic |
| NotYet | reading affectedSourceFile | 0 | insufficient-diagnostic |
| NotYet | reading afterImportPos | 0 | insufficient-diagnostic |
| NotYet | reading afterImportTagPos | 0 | insufficient-diagnostic |
| NotYet | reading alias | 0 | insufficient-diagnostic |
| NotYet | reading allAccessors | 0 | insufficient-diagnostic |
| NotYet | reading allComponentComputedNamesSerializable | 0 | insufficient-diagnostic |
| NotYet | reading allDiagnostics | 0 | insufficient-diagnostic |
| NotYet | reading allTypeFlags | 0 | insufficient-diagnostic |
| NotYet | reading alreadyTransformed | 0 | insufficient-diagnostic |
| NotYet | reading alternateForm | 0 | insufficient-diagnostic |
| NotYet | reading annotationSymbol | 0 | insufficient-diagnostic |
| NotYet | reading antecedent | 0 | insufficient-diagnostic |
| NotYet | reading antecedents | 0 | insufficient-diagnostic |
| NotYet | reading applicableByArity | 0 | insufficient-diagnostic |
| NotYet | reading array | 0 | insufficient-diagnostic |
| NotYet | reading arrayLiteral | 0 | insufficient-diagnostic |
| NotYet | reading arrayType | 0 | insufficient-diagnostic |
| NotYet | reading assertionNode | 0 | insufficient-diagnostic |
| NotYet | reading assignment | 0 | insufficient-diagnostic |
| NotYet | reading associatedName | 0 | insufficient-diagnostic |
| NotYet | reading associatedNames | 0 | insufficient-diagnostic |
| NotYet | reading assumeInitialized | 0 | insufficient-diagnostic |
| NotYet | reading asteriskToken | 0 | insufficient-diagnostic |
| NotYet | reading attributesType | 0 | insufficient-diagnostic |
| NotYet | reading attrs | 0 | insufficient-diagnostic |
| NotYet | reading awaitToken | 0 | insufficient-diagnostic |
| NotYet | reading awaitedType | 0 | insufficient-diagnostic |
| NotYet | reading b1 | 0 | insufficient-diagnostic |
| NotYet | reading backingField | 0 | insufficient-diagnostic |
| NotYet | reading bailedEarly | 0 | insufficient-diagnostic |
| NotYet | reading baseCandidates | 0 | insufficient-diagnostic |
| NotYet | reading baseConstructorType | 0 | insufficient-diagnostic |
| NotYet | reading baseIndexedAccess | 0 | insufficient-diagnostic |
| NotYet | reading baseObjectType | 0 | insufficient-diagnostic |
| NotYet | reading baseSymbol | 0 | insufficient-diagnostic |
| NotYet | reading baseType | 0 | insufficient-diagnostic |
| NotYet | reading baseTypeNode | 0 | insufficient-diagnostic |
| NotYet | reading baseTypeNodes | 0 | insufficient-diagnostic |
| NotYet | reading baseTypes | 0 | insufficient-diagnostic |
| NotYet | reading bases | 0 | insufficient-diagnostic |
| NotYet | reading best | 0 | insufficient-diagnostic |
| NotYet | reading bestMatch | 0 | insufficient-diagnostic |
| NotYet | reading binaryExpression | 0 | insufficient-diagnostic |
| NotYet | reading bindingElement | 0 | insufficient-diagnostic |
| NotYet | reading bindingList | 0 | insufficient-diagnostic |
| NotYet | reading block | 0 | insufficient-diagnostic |
| NotYet | reading boundTag | 0 | insufficient-diagnostic |
| NotYet | reading brandCheckIdentifier | 0 | insufficient-diagnostic |
| NotYet | reading buildInfoPath | 0 | insufficient-diagnostic |
| NotYet | reading builderProgram | 0 | insufficient-diagnostic |
| NotYet | reading bundle | 0 | insufficient-diagnostic |
| NotYet | reading byteOrderMarkIndicator | 0 | insufficient-diagnostic |
| NotYet | reading c | 0 | insufficient-diagnostic |
| NotYet | reading cacheAssignment | 0 | insufficient-diagnostic |
| NotYet | reading cacheKey | 0 | insufficient-diagnostic |
| NotYet | reading cachedResolvedSignatures | 0 | insufficient-diagnostic |
| NotYet | reading cachedTypes | 0 | insufficient-diagnostic |
| NotYet | reading call | 0 | insufficient-diagnostic |
| NotYet | reading callArgument | 0 | insufficient-diagnostic |
| NotYet | reading callExpr | 0 | insufficient-diagnostic |
| NotYet | reading callSignatures | 0 | insufficient-diagnostic |
| NotYet | reading callTarget | 0 | insufficient-diagnostic |
| NotYet | reading callee | 0 | insufficient-diagnostic |
| NotYet | reading candidateDirectories | 0 | insufficient-diagnostic |
| NotYet | reading canonicalFileName | 0 | insufficient-diagnostic |
| NotYet | reading captureNewTargetStatement | 0 | insufficient-diagnostic |
| NotYet | reading captureThisStatement | 0 | insufficient-diagnostic |
| NotYet | reading capturedBindings | 0 | insufficient-diagnostic |
| NotYet | reading chain | 0 | insufficient-diagnostic |
| NotYet | reading check | 0 | insufficient-diagnostic |
| NotYet | reading checkDiagnostics | 0 | insufficient-diagnostic |
| NotYet | reading checkType | 0 | insufficient-diagnostic |
| NotYet | reading child | 0 | insufficient-diagnostic |
| NotYet | reading childFieldType | 0 | insufficient-diagnostic |
| NotYet | reading childPropName | 0 | insufficient-diagnostic |
| NotYet | reading childrenNameType | 0 | insufficient-diagnostic |
| NotYet | reading childrenPropName | 0 | insufficient-diagnostic |
| NotYet | reading childrenTargetType | 0 | insufficient-diagnostic |
| NotYet | reading classDeclaration | 0 | insufficient-diagnostic |
| NotYet | reading classDeclarations | 0 | insufficient-diagnostic |
| NotYet | reading classFunction | 0 | insufficient-diagnostic |
| NotYet | reading classInstanceType | 0 | insufficient-diagnostic |
| NotYet | reading classNamedEvaluationHelperBlock | 0 | insufficient-diagnostic |
| NotYet | reading classThisAssignmentBlock | 0 | insufficient-diagnostic |
| NotYet | reading classType | 0 | insufficient-diagnostic |
| NotYet | reading clause | 0 | insufficient-diagnostic |
| NotYet | reading clauses | 0 | insufficient-diagnostic |
| NotYet | reading clone | 0 | insufficient-diagnostic |
| NotYet | reading closingLineTerminatorCount | 0 | insufficient-diagnostic |
| NotYet | reading collidingSymbol | 0 | insufficient-diagnostic |
| NotYet | reading columns | 0 | insufficient-diagnostic |
| NotYet | reading commentEnd | 0 | insufficient-diagnostic |
| NotYet | reading commentText | 0 | insufficient-diagnostic |
| NotYet | reading comments | 0 | insufficient-diagnostic |
| NotYet | reading commonDir | 0 | insufficient-diagnostic |
| NotYet | reading commonJSPropertyAccess | 0 | insufficient-diagnostic |
| NotYet | reading compilerOptionsProperty | 0 | insufficient-diagnostic |
| NotYet | reading computedPropertyName | 0 | insufficient-diagnostic |
| NotYet | reading conditionPrecedence | 0 | insufficient-diagnostic |
| NotYet | reading conditional | 0 | insufficient-diagnostic |
| NotYet | reading conditionalType | 0 | insufficient-diagnostic |
| NotYet | reading config | 0 | insufficient-diagnostic |
| NotYet | reading conflictingSymbolInfo | 0 | insufficient-diagnostic |
| NotYet | reading constraint | 0 | insufficient-diagnostic |
| NotYet | reading constraints | 0 | insufficient-diagnostic |
| NotYet | reading constructSignatures | 0 | insufficient-diagnostic |
| NotYet | reading constructorFunction | 0 | insufficient-diagnostic |
| NotYet | reading constructorLikeName | 0 | insufficient-diagnostic |
| NotYet | reading constructorSymbol | 0 | insufficient-diagnostic |
| NotYet | reading containingCall | 0 | insufficient-diagnostic |
| NotYet | reading containingClassDecl | 0 | insufficient-diagnostic |
| NotYet | reading containingMethod | 0 | insufficient-diagnostic |
| NotYet | reading context | 0 | insufficient-diagnostic |
| NotYet | reading contextFile | 0 | insufficient-diagnostic |
| NotYet | reading contextualType | 0 | insufficient-diagnostic |
| NotYet | reading ctor | 0 | insufficient-diagnostic |
| NotYet | reading currentGlobalDiagnostics | 0 | insufficient-diagnostic |
| NotYet | reading currentNamespace | 0 | insufficient-diagnostic |
| NotYet | reading currentNode | 0 | insufficient-diagnostic |
| NotYet | reading cwd | 0 | insufficient-diagnostic |
| NotYet | reading declarationName | 0 | insufficient-diagnostic |
| NotYet | reading declaredType | 0 | insufficient-diagnostic |
| NotYet | reading declaringClass | 0 | insufficient-diagnostic |
| NotYet | reading defaultConstraint | 0 | insufficient-diagnostic |
| NotYet | reading defaultExportSymbol | 0 | insufficient-diagnostic |
| NotYet | reading defaultMessage | 0 | insufficient-diagnostic |
| NotYet | reading defaultOnlyType | 0 | insufficient-diagnostic |
| NotYet | reading defaultReplaced | 0 | insufficient-diagnostic |
| NotYet | reading defaultType | 0 | insufficient-diagnostic |
| NotYet | reading deprecatedTag | 0 | insufficient-diagnostic |
| NotYet | reading derived | 0 | insufficient-diagnostic |
| NotYet | reading diagName | 0 | insufficient-diagnostic |
| NotYet | reading diagnostic | 0 | insufficient-diagnostic |
| NotYet | reading diagnosticStart | 0 | insufficient-diagnostic |
| NotYet | reading directlyRelated | 0 | insufficient-diagnostic |
| NotYet | reading directoryWatcher | 0 | insufficient-diagnostic |
| NotYet | reading discriminant | 0 | insufficient-diagnostic |
| NotYet | reading discriminantCombinations | 0 | insufficient-diagnostic |
| NotYet | reading discriminantType | 0 | insufficient-diagnostic |
| NotYet | reading discriminated | 0 | insufficient-diagnostic |
| NotYet | reading disposeScope | 0 | insufficient-diagnostic |
| NotYet | reading distributiveConstraint | 0 | insufficient-diagnostic |
| NotYet | reading doneType | 0 | insufficient-diagnostic |
| NotYet | reading downleveledImport | 0 | insufficient-diagnostic |
| NotYet | reading dupFile | 0 | insufficient-diagnostic |
| NotYet | reading effectiveArgs | 0 | insufficient-diagnostic |
| NotYet | reading element | 0 | insufficient-diagnostic |
| NotYet | reading elementAccess | 0 | insufficient-diagnostic |
| NotYet | reading elementFlags | 0 | insufficient-diagnostic |
| NotYet | reading elementTypes | 0 | insufficient-diagnostic |
| NotYet | reading elements | 0 | insufficient-diagnostic |
| NotYet | reading elideImport | 0 | insufficient-diagnostic |
| NotYet | reading emitAsSingleStatement | 0 | insufficient-diagnostic |
| NotYet | reading emitComments | 0 | insufficient-diagnostic |
| NotYet | reading emitExplicitInitializer | 0 | insufficient-diagnostic |
| NotYet | reading emitFileKey | 0 | insufficient-diagnostic |
| NotYet | reading emitResult | 0 | insufficient-diagnostic |
| NotYet | reading emitSignature | 0 | insufficient-diagnostic |
| NotYet | reading emitSourceMaps | 0 | insufficient-diagnostic |
| NotYet | reading emitSuperHelpers | 0 | insufficient-diagnostic |
| NotYet | reading emittedAsTopLevel | 0 | insufficient-diagnostic |
| NotYet | reading emittedCondition | 0 | insufficient-diagnostic |
| NotYet | reading emittedExpression | 0 | insufficient-diagnostic |
| NotYet | reading emittedOperand | 0 | insufficient-diagnostic |
| NotYet | reading enclosingBlockScopeContainer | 0 | insufficient-diagnostic |
| NotYet | reading enclosingClass | 0 | insufficient-diagnostic |
| NotYet | reading enclosingContainer | 0 | insufficient-diagnostic |
| NotYet | reading enclosingDeclaration | 0 | insufficient-diagnostic |
| NotYet | reading end | 0 | insufficient-diagnostic |
| NotYet | reading endLabel | 0 | insufficient-diagnostic |
| NotYet | reading entry | 0 | insufficient-diagnostic |
| NotYet | reading enumDeclaration | 0 | insufficient-diagnostic |
| NotYet | reading enumResult | 0 | insufficient-diagnostic |
| NotYet | reading enumStatement | 0 | insufficient-diagnostic |
| NotYet | reading envVarStatement | 0 | insufficient-diagnostic |
| NotYet | reading equalsToken | 0 | insufficient-diagnostic |
| NotYet | reading error | 0 | insufficient-diagnostic |
| NotYet | reading errorInfo | 0 | insufficient-diagnostic |
| NotYet | reading errorNode | 0 | insufficient-diagnostic |
| NotYet | reading errorRecord | 0 | insufficient-diagnostic |
| NotYet | reading errorSpan | 0 | insufficient-diagnostic |
| NotYet | reading esDecorateStatement | 0 | insufficient-diagnostic |
| NotYet | reading escapedText | 0 | insufficient-diagnostic |
| NotYet | reading evaluated | 0 | insufficient-diagnostic |
| NotYet | reading exception | 0 | insufficient-diagnostic |
| NotYet | reading exclamationToken | 0 | insufficient-diagnostic |
| NotYet | reading excludedProperties | 0 | insufficient-diagnostic |
| NotYet | reading existingPending | 0 | insufficient-diagnostic |
| NotYet | reading existingProp | 0 | insufficient-diagnostic |
| NotYet | reading exitNonUserCodeStatement | 0 | insufficient-diagnostic |
| NotYet | reading exportClause | 0 | insufficient-diagnostic |
| NotYet | reading exportDecl | 0 | insufficient-diagnostic |
| NotYet | reading exportEquals | 0 | insufficient-diagnostic |
| NotYet | reading exportEqualsSymbol | 0 | insufficient-diagnostic |
| NotYet | reading exportStars | 0 | insufficient-diagnostic |
| NotYet | reading exportStatement | 0 | insufficient-diagnostic |
| NotYet | reading exported | 0 | insufficient-diagnostic |
| NotYet | reading exportedTypeSymbol | 0 | insufficient-diagnostic |
| NotYet | reading exports | 0 | insufficient-diagnostic |
| NotYet | reading exprType | 0 | insufficient-diagnostic |
| NotYet | reading expressionPrecedence | 0 | insufficient-diagnostic |
| NotYet | reading expressionResult | 0 | insufficient-diagnostic |
| NotYet | reading extendedConstraint | 0 | insufficient-diagnostic |
| NotYet | reading extendsRaw | 0 | insufficient-diagnostic |
| NotYet | reading extendsType | 0 | insufficient-diagnostic |
| NotYet | reading externalHelpersModuleReference | 0 | insufficient-diagnostic |
| NotYet | reading fakeScope | 0 | insufficient-diagnostic |
| NotYet | reading fakeSignature | 0 | insufficient-diagnostic |
| NotYet | reading fakespace | 0 | insufficient-diagnostic |
| NotYet | reading falseSubtype | 0 | insufficient-diagnostic |
| NotYet | reading falseType | 0 | insufficient-diagnostic |
| NotYet | reading fileInfos | 0 | insufficient-diagnostic |
| NotYet | reading fileName | 0 | insufficient-diagnostic |
| NotYet | reading fileSystemEntryExists | 0 | insufficient-diagnostic |
| NotYet | reading filesDuplicates | 0 | insufficient-diagnostic |
| NotYet | reading filtered | 0 | insufficient-diagnostic |
| NotYet | reading filteredTypes | 0 | insufficient-diagnostic |
| NotYet | reading finished | 0 | insufficient-diagnostic |
| NotYet | reading first | 0 | insufficient-diagnostic |
| NotYet | reading firstDecl | 0 | insufficient-diagnostic |
| NotYet | reading firstDeclaration | 0 | insufficient-diagnostic |
| NotYet | reading firstDecorator | 0 | insufficient-diagnostic |
| NotYet | reading firstEnumMember | 0 | insufficient-diagnostic |
| NotYet | reading firstFile | 0 | insufficient-diagnostic |
| NotYet | reading firstIdentifier | 0 | insufficient-diagnostic |
| NotYet | reading fixed | 0 | insufficient-diagnostic |
| NotYet | reading flowType | 0 | insufficient-diagnostic |
| NotYet | reading fn | 0 | insufficient-diagnostic |
| NotYet | reading following | 0 | insufficient-diagnostic |
| NotYet | reading forInitializer | 0 | insufficient-diagnostic |
| NotYet | reading forStatement | 0 | insufficient-diagnostic |
| NotYet | reading forcedLookupLocation | 0 | insufficient-diagnostic |
| NotYet | reading fragment | 0 | insufficient-diagnostic |
| NotYet | reading freshType | 0 | insufficient-diagnostic |
| NotYet | reading freshTypeParameter | 0 | insufficient-diagnostic |
| NotYet | reading fromCache | 0 | insufficient-diagnostic |
| NotYet | reading fullName | 0 | insufficient-diagnostic |
| NotYet | reading func | 0 | insufficient-diagnostic |
| NotYet | reading functionFlags | 0 | insufficient-diagnostic |
| NotYet | reading functionName | 0 | insufficient-diagnostic |
| NotYet | reading generatedName | 0 | insufficient-diagnostic |
| NotYet | reading generatorFunc | 0 | insufficient-diagnostic |
| NotYet | reading getAccessorType | 0 | insufficient-diagnostic |
| NotYet | reading getFunc | 0 | insufficient-diagnostic |
| NotYet | reading getModifiedTime | 0 | insufficient-diagnostic |
| NotYet | reading getter | 0 | insufficient-diagnostic |
| NotYet | reading grandParent | 0 | insufficient-diagnostic |
| NotYet | reading graphNode | 0 | insufficient-diagnostic |
| NotYet | reading hasDefaultClause | 0 | insufficient-diagnostic |
| NotYet | reading hasEmptyObject | 0 | insufficient-diagnostic |
| NotYet | reading hasExistingReasonToReportErrorOn | 0 | insufficient-diagnostic |
| NotYet | reading hasExtends | 0 | insufficient-diagnostic |
| NotYet | reading hasInstanceProperty | 0 | insufficient-diagnostic |
| NotYet | reading hasPrivateIdentifier | 0 | insufficient-diagnostic |
| NotYet | reading hasSyntheticDefault | 0 | insufficient-diagnostic |
| NotYet | reading hostSourceFileInfo | 0 | insufficient-diagnostic |
| NotYet | reading id | 0 | insufficient-diagnostic |
| NotYet | reading identifier | 0 | insufficient-diagnostic |
| NotYet | reading ids | 0 | insufficient-diagnostic |
| NotYet | reading ifStatement | 0 | insufficient-diagnostic |
| NotYet | reading iife | 0 | insufficient-diagnostic |
| NotYet | reading implDecl | 0 | insufficient-diagnostic |
| NotYet | reading implementsTypeNodes | 0 | insufficient-diagnostic |
| NotYet | reading importAttributesArgument | 0 | insufficient-diagnostic |
| NotYet | reading importClause | 0 | insufficient-diagnostic |
| NotYet | reading importDecl | 0 | insufficient-diagnostic |
| NotYet | reading importOrExport | 0 | insufficient-diagnostic |
| NotYet | reading importSymbol | 0 | insufficient-diagnostic |
| NotYet | reading importedFileNames | 0 | insufficient-diagnostic |
| NotYet | reading inAmbientContextOrInterface | 0 | insufficient-diagnostic |
| NotYet | reading inTupleContext | 0 | insufficient-diagnostic |
| NotYet | reading includes | 0 | insufficient-diagnostic |
| NotYet | reading indexConstraint | 0 | insufficient-diagnostic |
| NotYet | reading indexInfo | 0 | insufficient-diagnostic |
| NotYet | reading indexInfos | 0 | insufficient-diagnostic |
| NotYet | reading indexOfParameter | 0 | insufficient-diagnostic |
| NotYet | reading indexSymbol | 0 | insufficient-diagnostic |
| NotYet | reading indexType | 0 | insufficient-diagnostic |
| NotYet | reading indexedAccessType | 0 | insufficient-diagnostic |
| NotYet | reading indexedType | 0 | insufficient-diagnostic |
| NotYet | reading inference | 0 | insufficient-diagnostic |
| NotYet | reading inferences | 0 | insufficient-diagnostic |
| NotYet | reading inferredProp | 0 | insufficient-diagnostic |
| NotYet | reading info | 0 | insufficient-diagnostic |
| NotYet | reading init | 0 | insufficient-diagnostic |
| NotYet | reading initialLocationForSecondaryLookup | 0 | insufficient-diagnostic |
| NotYet | reading initialType | 0 | insufficient-diagnostic |
| NotYet | reading initializerStatement | 0 | insufficient-diagnostic |
| NotYet | reading initializerWithoutParens | 0 | insufficient-diagnostic |
| NotYet | reading inlinable | 0 | insufficient-diagnostic |
| NotYet | reading inner | 0 | insufficient-diagnostic |
| NotYet | reading innerExpression | 0 | insufficient-diagnostic |
| NotYet | reading innerModuleSymbol | 0 | insufficient-diagnostic |
| NotYet | reading insertIndex | 0 | insufficient-diagnostic |
| NotYet | reading instantiated | 0 | insufficient-diagnostic |
| NotYet | reading instantiatedTemplateType | 0 | insufficient-diagnostic |
| NotYet | reading instantiation | 0 | insufficient-diagnostic |
| NotYet | reading instantiationExpressionType | 0 | insufficient-diagnostic |
| NotYet | reading instantiations | 0 | insufficient-diagnostic |
| NotYet | reading internalModuleReference | 0 | insufficient-diagnostic |
| NotYet | reading intrinsicAttribs | 0 | insufficient-diagnostic |
| NotYet | reading intrinsicAttributes | 0 | insufficient-diagnostic |
| NotYet | reading intrinsicElementsType | 0 | insufficient-diagnostic |
| NotYet | reading intrinsicType | 0 | insufficient-diagnostic |
| NotYet | reading intrinsics | 0 | insufficient-diagnostic |
| NotYet | reading introducesError | 0 | insufficient-diagnostic |
| NotYet | reading invocation | 0 | insufficient-diagnostic |
| NotYet | reading invokedExpression | 0 | insufficient-diagnostic |
| NotYet | reading isAmbient | 0 | insufficient-diagnostic |
| NotYet | reading isAnonymous | 0 | insufficient-diagnostic |
| NotYet | reading isCallToReadHelper | 0 | insufficient-diagnostic |
| NotYet | reading isCallbackTag | 0 | insufficient-diagnostic |
| NotYet | reading isCapturedInFunction | 0 | insufficient-diagnostic |
| NotYet | reading isClassWithConstructorReference | 0 | insufficient-diagnostic |
| NotYet | reading isConfigIdentical | 0 | insufficient-diagnostic |
| NotYet | reading isDeferredMappedIndex | 0 | insufficient-diagnostic |
| NotYet | reading isDerivedClass | 0 | insufficient-diagnostic |
| NotYet | reading isEmpty | 0 | insufficient-diagnostic |
| NotYet | reading isEsmCjsRef | 0 | insufficient-diagnostic |
| NotYet | reading isExportAssignmentCompatibleSymbolName | 0 | insufficient-diagnostic |
| NotYet | reading isExportEquals | 0 | insufficient-diagnostic |
| NotYet | reading isExternalImportAlias | 0 | insufficient-diagnostic |
| NotYet | reading isFinite | 0 | insufficient-diagnostic |
| NotYet | reading isFromNodeModulesSearch | 0 | insufficient-diagnostic |
| NotYet | reading isGenerator | 0 | insufficient-diagnostic |
| NotYet | reading isGlobal | 0 | insufficient-diagnostic |
| NotYet | reading isIllegalExportDefaultInCJS | 0 | insufficient-diagnostic |
| NotYet | reading isImmediatelyInvoked | 0 | insufficient-diagnostic |
| NotYet | reading isImportTypeWithQualifier | 0 | insufficient-diagnostic |
| NotYet | reading isJSDoc | 0 | insufficient-diagnostic |
| NotYet | reading isJSObjectLiteralInitializer | 0 | insufficient-diagnostic |
| NotYet | reading isJsFileFromNodeModules | 0 | insufficient-diagnostic |
| NotYet | reading isKnownProperty | 0 | insufficient-diagnostic |
| NotYet | reading isLeftNaN | 0 | insufficient-diagnostic |
| NotYet | reading isLengthPushOrUnshift | 0 | insufficient-diagnostic |
| NotYet | reading isMarkdownOrJSDocLink | 0 | insufficient-diagnostic |
| NotYet | reading isOptionalParameter | 0 | insufficient-diagnostic |
| NotYet | reading isPropertyName | 0 | insufficient-diagnostic |
| NotYet | reading isReservedWord | 0 | insufficient-diagnostic |
| NotYet | reading isSimilarNode | 0 | insufficient-diagnostic |
| NotYet | reading isSyncImport | 0 | insufficient-diagnostic |
| NotYet | reading isTopLevel | 0 | insufficient-diagnostic |
| NotYet | reading isTypeOnly | 0 | insufficient-diagnostic |
| NotYet | reading isUsed | 0 | insufficient-diagnostic |
| NotYet | reading isValue | 0 | insufficient-diagnostic |
| NotYet | reading isVoidPromiseError | 0 | insufficient-diagnostic |
| NotYet | reading isZero | 0 | insufficient-diagnostic |
| NotYet | reading issuedDiagnostic | 0 | insufficient-diagnostic |
| NotYet | reading item | 0 | insufficient-diagnostic |
| NotYet | reading iteratedType | 0 | insufficient-diagnostic |
| NotYet | reading iterator | 0 | insufficient-diagnostic |
| NotYet | reading iteratorValueStatement | 0 | insufficient-diagnostic |
| NotYet | reading jsDoc | 0 | insufficient-diagnostic |
| NotYet | reading jsDocNamespaceNode | 0 | insufficient-diagnostic |
| NotYet | reading jsdocAliasDecl | 0 | insufficient-diagnostic |
| NotYet | reading jsdocParameters | 0 | insufficient-diagnostic |
| NotYet | reading jsdocTypeLiteral | 0 | insufficient-diagnostic |
| NotYet | reading jsxChildrenPropertyName | 0 | insufficient-diagnostic |
| NotYet | reading jsxFactoryNamespace | 0 | insufficient-diagnostic |
| NotYet | reading jsxFactorySymbol | 0 | insufficient-diagnostic |
| NotYet | reading jsxFragPragma | 0 | insufficient-diagnostic |
| NotYet | reading jsxFragPragmas | 0 | insufficient-diagnostic |
| NotYet | reading jsxFragmentFactoryName | 0 | insufficient-diagnostic |
| NotYet | reading jsxSpecific | 0 | insufficient-diagnostic |
| NotYet | reading keyAttr | 0 | insufficient-diagnostic |
| NotYet | reading keyPropertyName | 0 | insufficient-diagnostic |
| NotYet | reading labeledElementDeclaration | 0 | insufficient-diagnostic |
| NotYet | reading labeledElementDeclarations | 0 | insufficient-diagnostic |
| NotYet | reading lastChild | 0 | insufficient-diagnostic |
| NotYet | reading lastJSDocParam | 0 | insufficient-diagnostic |
| NotYet | reading lastLeft | 0 | insufficient-diagnostic |
| NotYet | reading lastStatement | 0 | insufficient-diagnostic |
| NotYet | reading lateSymbol | 0 | insufficient-diagnostic |
| NotYet | reading leadingLineTerminatorCount | 0 | insufficient-diagnostic |
| NotYet | reading leadingNewlines | 0 | insufficient-diagnostic |
| NotYet | reading leftSpread | 0 | insufficient-diagnostic |
| NotYet | reading leftTarget | 0 | insufficient-diagnostic |
| NotYet | reading leftmostExpressionKind | 0 | insufficient-diagnostic |
| NotYet | reading len | 0 | insufficient-diagnostic |
| NotYet | reading lex | 0 | insufficient-diagnostic |
| NotYet | reading lexicallyScopedSymbol | 0 | insufficient-diagnostic |
| NotYet | reading lhsExpr | 0 | insufficient-diagnostic |
| NotYet | reading limitedConstraint | 0 | insufficient-diagnostic |
| NotYet | reading lineText | 0 | insufficient-diagnostic |
| NotYet | reading lines | 0 | insufficient-diagnostic |
| NotYet | reading linesAfterDot | 0 | insufficient-diagnostic |
| NotYet | reading linesBeforeDot | 0 | insufficient-diagnostic |
| NotYet | reading links | 0 | insufficient-diagnostic |
| NotYet | reading list | 0 | insufficient-diagnostic |
| NotYet | reading literal | 0 | insufficient-diagnostic |
| NotYet | reading literalProp | 0 | insufficient-diagnostic |
| NotYet | reading literalType | 0 | insufficient-diagnostic |
| NotYet | reading literalValue | 0 | insufficient-diagnostic |
| NotYet | reading literals | 0 | insufficient-diagnostic |
| NotYet | reading local | 0 | insufficient-diagnostic |
| NotYet | reading localDeclarationSymbol | 0 | insufficient-diagnostic |
| NotYet | reading localJsxNamespace | 0 | insufficient-diagnostic |
| NotYet | reading localProps | 0 | insufficient-diagnostic |
| NotYet | reading localSymbol | 0 | insufficient-diagnostic |
| NotYet | reading location | 0 | insufficient-diagnostic |
| NotYet | reading loopResultName | 0 | insufficient-diagnostic |
| NotYet | reading major | 0 | insufficient-diagnostic |
| NotYet | reading mappedSource | 0 | insufficient-diagnostic |
| NotYet | reading mappedTypeNode | 0 | insufficient-diagnostic |
| NotYet | reading match | 0 | insufficient-diagnostic |
| NotYet | reading matching | 0 | insufficient-diagnostic |
| NotYet | reading meaning | 0 | insufficient-diagnostic |
| NotYet | reading member | 0 | insufficient-diagnostic |
| NotYet | reading memberDecoratorsAssignment | 0 | insufficient-diagnostic |
| NotYet | reading memberProps | 0 | insufficient-diagnostic |
| NotYet | reading metaPropertySymbol | 0 | insufficient-diagnostic |
| NotYet | reading metadataReference | 0 | insufficient-diagnostic |
| NotYet | reading method | 0 | insufficient-diagnostic |
| NotYet | reading methodSignatures | 0 | insufficient-diagnostic |
| NotYet | reading methodType | 0 | insufficient-diagnostic |
| NotYet | reading missingNode | 0 | insufficient-diagnostic |
| NotYet | reading modifier | 0 | insufficient-diagnostic |
| NotYet | reading modifiers | 0 | insufficient-diagnostic |
| NotYet | reading modifiersWithoutAccessor | 0 | insufficient-diagnostic |
| NotYet | reading moduleBlock | 0 | insufficient-diagnostic |
| NotYet | reading moduleExports | 0 | insufficient-diagnostic |
| NotYet | reading moduleName | 0 | insufficient-diagnostic |
| NotYet | reading moduleReference | 0 | insufficient-diagnostic |
| NotYet | reading moduleSpecifier | 0 | insufficient-diagnostic |
| NotYet | reading moduleSpecifiers | 0 | insufficient-diagnostic |
| NotYet | reading moduleStatement | 0 | insufficient-diagnostic |
| NotYet | reading moduleSym | 0 | insufficient-diagnostic |
| NotYet | reading moduleSymbol | 0 | insufficient-diagnostic |
| NotYet | reading moduleTag | 0 | insufficient-diagnostic |
| NotYet | reading multiLine | 0 | insufficient-diagnostic |
| NotYet | reading nameStr | 0 | insufficient-diagnostic |
| NotYet | reading nameText | 0 | insufficient-diagnostic |
| NotYet | reading nameType | 0 | insufficient-diagnostic |
| NotYet | reading named | 0 | insufficient-diagnostic |
| NotYet | reading namedBindings | 0 | insufficient-diagnostic |
| NotYet | reading names | 0 | insufficient-diagnostic |
| NotYet | reading namespaceDeclaration | 0 | insufficient-diagnostic |
| NotYet | reading namespaceName | 0 | insufficient-diagnostic |
| NotYet | reading narrowedType | 0 | insufficient-diagnostic |
| NotYet | reading needCheckWidenedType | 0 | insufficient-diagnostic |
| NotYet | reading needCompilerDiagnostic | 0 | insufficient-diagnostic |
| NotYet | reading needsOutParam | 0 | insufficient-diagnostic |
| NotYet | reading needsParens | 0 | insufficient-diagnostic |
| NotYet | reading neverProp | 0 | insufficient-diagnostic |
| NotYet | reading newBaseType | 0 | insufficient-diagnostic |
| NotYet | reading newConstraint | 0 | insufficient-diagnostic |
| NotYet | reading newConstraintParam | 0 | insufficient-diagnostic |
| NotYet | reading newMapper | 0 | insufficient-diagnostic |
| NotYet | reading newName | 0 | insufficient-diagnostic |
| NotYet | reading newParam | 0 | insufficient-diagnostic |
| NotYet | reading newParametersArray | 0 | insufficient-diagnostic |
| NotYet | reading newReturnType | 0 | insufficient-diagnostic |
| NotYet | reading newRoot | 0 | insufficient-diagnostic |
| NotYet | reading newStatements | 0 | insufficient-diagnostic |
| NotYet | reading newSymbol | 0 | insufficient-diagnostic |
| NotYet | reading newTypes | 0 | insufficient-diagnostic |
| NotYet | reading newVarStatement | 0 | insufficient-diagnostic |
| NotYet | reading newVariableDeclaration | 0 | insufficient-diagnostic |
| NotYet | reading noInferSymbol | 0 | insufficient-diagnostic |
| NotYet | reading node | 0 | insufficient-diagnostic |
| NotYet | reading nodeConstructors | 0 | insufficient-diagnostic |
| NotYet | reading nodeContextFlags | 0 | insufficient-diagnostic |
| NotYet | reading nodes | 0 | insufficient-diagnostic |
| NotYet | reading nonAwaitStatement | 0 | insufficient-diagnostic |
| NotYet | reading nonPrologueStart | 0 | insufficient-diagnostic |
| NotYet | reading nonValueSymbol | 0 | insufficient-diagnostic |
| NotYet | reading normalizedElements | 0 | insufficient-diagnostic |
| NotYet | reading ns | 0 | insufficient-diagnostic |
| NotYet | reading nullishSemantics | 0 | insufficient-diagnostic |
| NotYet | reading numNodes | 0 | insufficient-diagnostic |
| NotYet | reading o1 | 0 | insufficient-diagnostic |
| NotYet | reading objectLiteral | 0 | insufficient-diagnostic |
| NotYet | reading objectLiterals | 0 | insufficient-diagnostic |
| NotYet | reading objectProperties | 0 | insufficient-diagnostic |
| NotYet | reading objectTypes | 0 | insufficient-diagnostic |
| NotYet | reading offset | 0 | insufficient-diagnostic |
| NotYet | reading ok | 0 | insufficient-diagnostic |
| NotYet | reading oldOptions | 0 | insufficient-diagnostic |
| NotYet | reading oldProgram | 0 | insufficient-diagnostic |
| NotYet | reading oldResolution | 0 | insufficient-diagnostic |
| NotYet | reading oldSourceFiles | 0 | insufficient-diagnostic |
| NotYet | reading oldTextPrefix | 0 | insufficient-diagnostic |
| NotYet | reading openBracePosition | 0 | insufficient-diagnostic |
| NotYet | reading openBracketPosition | 0 | insufficient-diagnostic |
| NotYet | reading openParenPosition | 0 | insufficient-diagnostic |
| NotYet | reading operandConstraint | 0 | insufficient-diagnostic |
| NotYet | reading operandPrecedence | 0 | insufficient-diagnostic |
| NotYet | reading operatorToken | 0 | insufficient-diagnostic |
| NotYet | reading optionsType | 0 | insufficient-diagnostic |
| NotYet | reading origTypeParameter | 0 | insufficient-diagnostic |
| NotYet | reading origin | 0 | insufficient-diagnostic |
| NotYet | reading original | 0 | insufficient-diagnostic |
| NotYet | reading originalClass | 0 | insufficient-diagnostic |
| NotYet | reading originalClassDecl | 0 | insufficient-diagnostic |
| NotYet | reading originalFile | 0 | insufficient-diagnostic |
| NotYet | reading originalModuleSpecifier | 0 | insufficient-diagnostic |
| NotYet | reading other | 0 | insufficient-diagnostic |
| NotYet | reading otherFiles | 0 | insufficient-diagnostic |
| NotYet | reading outer | 0 | insufficient-diagnostic |
| NotYet | reading outerTypeParameters | 0 | insufficient-diagnostic |
| NotYet | reading overloadSignatures | 0 | insufficient-diagnostic |
| NotYet | reading ownKey | 0 | insufficient-diagnostic |
| NotYet | reading packageJsonMap | 0 | insufficient-diagnostic |
| NotYet | reading param | 0 | insufficient-diagnostic |
| NotYet | reading paramSymbol | 0 | insufficient-diagnostic |
| NotYet | reading paramTypeProperty | 0 | insufficient-diagnostic |
| NotYet | reading parameterIndex | 0 | insufficient-diagnostic |
| NotYet | reading parameterNode | 0 | insufficient-diagnostic |
| NotYet | reading parameterRange | 0 | insufficient-diagnostic |
| NotYet | reading parameterSymbol | 0 | insufficient-diagnostic |
| NotYet | reading parameterTypeOfTypeTag | 0 | insufficient-diagnostic |
| NotYet | reading parametersWithPropertyAssignments | 0 | insufficient-diagnostic |
| NotYet | reading params | 0 | insufficient-diagnostic |
| NotYet | reading parentFile | 0 | insufficient-diagnostic |
| NotYet | reading parentSymbol | 0 | insufficient-diagnostic |
| NotYet | reading parentType | 0 | insufficient-diagnostic |
| NotYet | reading parenthesizerRule | 0 | insufficient-diagnostic |
| NotYet | reading parsed | 0 | insufficient-diagnostic |
| NotYet | reading parsedCommandLine | 0 | insufficient-diagnostic |
| NotYet | reading pendingKind | 0 | insufficient-diagnostic |
| NotYet | reading pipelinePhase | 0 | insufficient-diagnostic |
| NotYet | reading pollScheduled | 0 | insufficient-diagnostic |
| NotYet | reading postSuper | 0 | insufficient-diagnostic |
| NotYet | reading potentiallyUnusedIdentifiers | 0 | insufficient-diagnostic |
| NotYet | reading precedingLineBreak | 0 | insufficient-diagnostic |
| NotYet | reading prevNodeIndex | 0 | insufficient-diagnostic |
| NotYet | reading prevSignature | 0 | insufficient-diagnostic |
| NotYet | reading prevStatement | 0 | insufficient-diagnostic |
| NotYet | reading previousGlobalDiagnostics | 0 | insufficient-diagnostic |
| NotYet | reading primaryDeclaration | 0 | insufficient-diagnostic |
| NotYet | reading primaryTypes | 0 | insufficient-diagnostic |
| NotYet | reading privateProp | 0 | insufficient-diagnostic |
| NotYet | reading program | 0 | insufficient-diagnostic |
| NotYet | reading programDiagnosticsInFile | 0 | insufficient-diagnostic |
| NotYet | reading prologueStatementCount | 0 | insufficient-diagnostic |
| NotYet | reading promise | 0 | insufficient-diagnostic |
| NotYet | reading prop | 0 | insufficient-diagnostic |
| NotYet | reading propContext | 0 | insufficient-diagnostic |
| NotYet | reading propName | 0 | insufficient-diagnostic |
| NotYet | reading propNode | 0 | insufficient-diagnostic |
| NotYet | reading propType | 0 | insufficient-diagnostic |
| NotYet | reading property | 0 | insufficient-diagnostic |
| NotYet | reading propertyAccess | 0 | insufficient-diagnostic |
| NotYet | reading propertyName | 0 | insufficient-diagnostic |
| NotYet | reading propertyOriginalNode | 0 | insufficient-diagnostic |
| NotYet | reading props | 0 | insufficient-diagnostic |
| NotYet | reading proto | 0 | insufficient-diagnostic |
| NotYet | reading prototype | 0 | insufficient-diagnostic |
| NotYet | reading prototypePropertyType | 0 | insufficient-diagnostic |
| NotYet | reading prototypeSymbol | 0 | insufficient-diagnostic |
| NotYet | reading qualifiedName | 0 | insufficient-diagnostic |
| NotYet | reading questionDotToken | 0 | insufficient-diagnostic |
| NotYet | reading questionToken | 0 | insufficient-diagnostic |
| NotYet | reading quick | 0 | insufficient-diagnostic |
| NotYet | reading r | 0 | insufficient-diagnostic |
| NotYet | reading r1 | 0 | insufficient-diagnostic |
| NotYet | reading raw | 0 | insufficient-diagnostic |
| NotYet | reading rawName | 0 | insufficient-diagnostic |
| NotYet | reading reachable | 0 | insufficient-diagnostic |
| NotYet | reading reactExports | 0 | insufficient-diagnostic |
| NotYet | reading readExpression | 0 | insufficient-diagnostic |
| NotYet | reading realDeclarationPath | 0 | insufficient-diagnostic |
| NotYet | reading recursionIdentity | 0 | insufficient-diagnostic |
| NotYet | reading recursiveInnerModule | 0 | insufficient-diagnostic |
| NotYet | reading redirect | 0 | insufficient-diagnostic |
| NotYet | reading reduced | 0 | insufficient-diagnostic |
| NotYet | reading reducedTypes | 0 | insufficient-diagnostic |
| NotYet | reading reexports | 0 | insufficient-diagnostic |
| NotYet | reading ref | 0 | insufficient-diagnostic |
| NotYet | reading reference | 0 | insufficient-diagnostic |
| NotYet | reading referenceRedirect | 0 | insufficient-diagnostic |
| NotYet | reading referenced | 0 | insufficient-diagnostic |
| NotYet | reading referencedMap | 0 | insufficient-diagnostic |
| NotYet | reading referencedName | 0 | insufficient-diagnostic |
| NotYet | reading regularNew | 0 | insufficient-diagnostic |
| NotYet | reading regularType | 0 | insufficient-diagnostic |
| NotYet | reading related | 0 | insufficient-diagnostic |
| NotYet | reading relatedDiagnostics | 0 | insufficient-diagnostic |
| NotYet | reading relatedInfo | 0 | insufficient-diagnostic |
| NotYet | reading relevantConstraint | 0 | insufficient-diagnostic |
| NotYet | reading relevantTypeParameter | 0 | insufficient-diagnostic |
| NotYet | reading remainingMembers | 0 | insufficient-diagnostic |
| NotYet | reading remove | 0 | insufficient-diagnostic |
| NotYet | reading replacements | 0 | insufficient-diagnostic |
| NotYet | reading requiresAddingUndefined | 0 | insufficient-diagnostic |
| NotYet | reading resolution | 0 | insufficient-diagnostic |
| NotYet | reading resolutions | 0 | insufficient-diagnostic |
| NotYet | reading resolutionsChanged | 0 | insufficient-diagnostic |
| NotYet | reading resolvedRequire | 0 | insufficient-diagnostic |
| NotYet | reading resolvedTypeReferenceDirective | 0 | insufficient-diagnostic |
| NotYet | reading resolvedTypeSymbol | 0 | insufficient-diagnostic |
| NotYet | reading resolvedValueSymbol | 0 | insufficient-diagnostic |
| NotYet | reading resolver | 0 | insufficient-diagnostic |
| NotYet | reading rest | 0 | insufficient-diagnostic |
| NotYet | reading restParamSymbol | 0 | insufficient-diagnostic |
| NotYet | reading restType | 0 | insufficient-diagnostic |
| NotYet | reading resultFromDts | 0 | insufficient-diagnostic |
| NotYet | reading results | 0 | insufficient-diagnostic |
| NotYet | reading returnMethod | 0 | insufficient-diagnostic |
| NotYet | reading returnType | 0 | insufficient-diagnostic |
| NotYet | reading returnTypeNode | 0 | insufficient-diagnostic |
| NotYet | reading returnTypeProperty | 0 | insufficient-diagnostic |
| NotYet | reading reusable | 0 | insufficient-diagnostic |
| NotYet | reading reversed | 0 | insufficient-diagnostic |
| NotYet | reading root | 0 | insufficient-diagnostic |
| NotYet | reading rootExpr | 0 | insufficient-diagnostic |
| NotYet | reading rootNames | 0 | insufficient-diagnostic |
| NotYet | reading rootSymbol | 0 | insufficient-diagnostic |
| NotYet | reading s | 0 | insufficient-diagnostic |
| NotYet | reading savedInStrictMode | 0 | insufficient-diagnostic |
| NotYet | reading savedPreserveSourceNewlines | 0 | insufficient-diagnostic |
| NotYet | reading seen | 0 | insufficient-diagnostic |
| NotYet | reading seenSymbols | 0 | insufficient-diagnostic |
| NotYet | reading semanticDiagnostics | 0 | insufficient-diagnostic |
| NotYet | reading semanticDiagnosticsPerFile | 0 | insufficient-diagnostic |
| NotYet | reading separatingLineTerminatorCount | 0 | insufficient-diagnostic |
| NotYet | reading serializationKind | 0 | insufficient-diagnostic |
| NotYet | reading serializedName | 0 | insufficient-diagnostic |
| NotYet | reading setAccessor | 0 | insufficient-diagnostic |
| NotYet | reading setFunc | 0 | insufficient-diagnostic |
| NotYet | reading setProp | 0 | insufficient-diagnostic |
| NotYet | reading setter | 0 | insufficient-diagnostic |
| NotYet | reading setterModifiers | 0 | insufficient-diagnostic |
| NotYet | reading shouldEmitDetachedComment | 0 | insufficient-diagnostic |
| NotYet | reading shouldEmitDotDot | 0 | insufficient-diagnostic |
| NotYet | reading shouldResolveAlias | 0 | insufficient-diagnostic |
| NotYet | reading shouldResolveFactoryReference | 0 | insufficient-diagnostic |
| NotYet | reading signature | 0 | insufficient-diagnostic |
| NotYet | reading signatureDeclaration | 0 | insufficient-diagnostic |
| NotYet | reading signatureNode | 0 | insufficient-diagnostic |
| NotYet | reading signaturesWithCorrectTypeArgumentArity | 0 | insufficient-diagnostic |
| NotYet | reading skipCaching | 0 | insufficient-diagnostic |
| NotYet | reading skipped | 0 | insufficient-diagnostic |
| NotYet | reading sortedIndex | 0 | insufficient-diagnostic |
| NotYet | reading source | 0 | insufficient-diagnostic |
| NotYet | reading sourceDiscriminantTypes | 0 | insufficient-diagnostic |
| NotYet | reading sourceExtends | 0 | insufficient-diagnostic |
| NotYet | reading sourceFilePath | 0 | insufficient-diagnostic |
| NotYet | reading sourceFiles | 0 | insufficient-diagnostic |
| NotYet | reading sourceOrigin | 0 | insufficient-diagnostic |
| NotYet | reading sourceParams | 0 | insufficient-diagnostic |
| NotYet | reading sourceProp | 0 | insufficient-diagnostic |
| NotYet | reading sourceSize | 0 | insufficient-diagnostic |
| NotYet | reading sourceTypes | 0 | insufficient-diagnostic |
| NotYet | reading sourceUnionOrIntersection | 0 | insufficient-diagnostic |
| NotYet | reading sources | 0 | insufficient-diagnostic |
| NotYet | reading specialPropertyAssignmentKind | 0 | insufficient-diagnostic |
| NotYet | reading specifier | 0 | insufficient-diagnostic |
| NotYet | reading specifierType | 0 | insufficient-diagnostic |
| NotYet | reading spread | 0 | insufficient-diagnostic |
| NotYet | reading spreadElement | 0 | insufficient-diagnostic |
| NotYet | reading spreadType | 0 | insufficient-diagnostic |
| NotYet | reading startPos | 0 | insufficient-diagnostic |
| NotYet | reading stat | 0 | insufficient-diagnostic |
| NotYet | reading stateVariable | 0 | insufficient-diagnostic |
| NotYet | reading statementExpression | 0 | insufficient-diagnostic |
| NotYet | reading statementOffset | 0 | insufficient-diagnostic |
| NotYet | reading statementsArray | 0 | insufficient-diagnostic |
| NotYet | reading staticBlock | 0 | insufficient-diagnostic |
| NotYet | reading staticBlocks | 0 | insufficient-diagnostic |
| NotYet | reading staticType | 0 | insufficient-diagnostic |
| NotYet | reading subcontext | 0 | insufficient-diagnostic |
| NotYet | reading subsequentName | 0 | insufficient-diagnostic |
| NotYet | reading substitute | 0 | insufficient-diagnostic |
| NotYet | reading substitutionType | 0 | insufficient-diagnostic |
| NotYet | reading suggestedExt | 0 | insufficient-diagnostic |
| NotYet | reading suggestedType | 0 | insufficient-diagnostic |
| NotYet | reading superContainerFlags | 0 | insufficient-diagnostic |
| NotYet | reading superPath | 0 | insufficient-diagnostic |
| NotYet | reading superProperty | 0 | insufficient-diagnostic |
| NotYet | reading superStatement | 0 | insufficient-diagnostic |
| NotYet | reading superStatementIndices | 0 | insufficient-diagnostic |
| NotYet | reading swappedMode | 0 | insufficient-diagnostic |
| NotYet | reading switchStatement | 0 | insufficient-diagnostic |
| NotYet | reading sym | 0 | insufficient-diagnostic |
| NotYet | reading symbol | 0 | insufficient-diagnostic |
| NotYet | reading symbolExport | 0 | insufficient-diagnostic |
| NotYet | reading symbolFromSymbolTable | 0 | insufficient-diagnostic |
| NotYet | reading symbols | 0 | insufficient-diagnostic |
| NotYet | reading symlinkedDirectories | 0 | insufficient-diagnostic |
| NotYet | reading synthType | 0 | insufficient-diagnostic |
| NotYet | reading syntheticArgsSymbol | 0 | insufficient-diagnostic |
| NotYet | reading t | 0 | insufficient-diagnostic |
| NotYet | reading tagExpression | 0 | insufficient-diagnostic |
| NotYet | reading tagName | 0 | insufficient-diagnostic |
| NotYet | reading tagNameDeclaration | 0 | insufficient-diagnostic |
| NotYet | reading targetDeclarationKind | 0 | insufficient-diagnostic |
| NotYet | reading targetFile | 0 | insufficient-diagnostic |
| NotYet | reading targetFlags | 0 | insufficient-diagnostic |
| NotYet | reading targetMode | 0 | insufficient-diagnostic |
| NotYet | reading targetOrigin | 0 | insufficient-diagnostic |
| NotYet | reading targetParam | 0 | insufficient-diagnostic |
| NotYet | reading targetProp | 0 | insufficient-diagnostic |
| NotYet | reading targetProperty | 0 | insufficient-diagnostic |
| NotYet | reading targetPropertySymbol | 0 | insufficient-diagnostic |
| NotYet | reading targetSymbol | 0 | insufficient-diagnostic |
| NotYet | reading targetType | 0 | insufficient-diagnostic |
| NotYet | reading targetUnionOrIntersection | 0 | insufficient-diagnostic |
| NotYet | reading targetValue | 0 | insufficient-diagnostic |
| NotYet | reading targets | 0 | insufficient-diagnostic |
| NotYet | reading tempSources | 0 | insufficient-diagnostic |
| NotYet | reading templateType | 0 | insufficient-diagnostic |
| NotYet | reading templates | 0 | insufficient-diagnostic |
| NotYet | reading testedNode | 0 | insufficient-diagnostic |
| NotYet | reading testedSymbol | 0 | insufficient-diagnostic |
| NotYet | reading texts | 0 | insufficient-diagnostic |
| NotYet | reading thenFunction | 0 | insufficient-diagnostic |
| NotYet | reading thisContainer | 0 | insufficient-diagnostic |
| NotYet | reading thisParam | 0 | insufficient-diagnostic |
| NotYet | reading thisType | 0 | insufficient-diagnostic |
| NotYet | reading timerToInvalidateFailedLookupResolutions | 0 | insufficient-diagnostic |
| NotYet | reading timerToUpdateChildWatches | 0 | insufficient-diagnostic |
| NotYet | reading timerToUpdateProgram | 0 | insufficient-diagnostic |
| NotYet | reading toWatch | 0 | insufficient-diagnostic |
| NotYet | reading tokenPos | 0 | insufficient-diagnostic |
| NotYet | reading tokenText | 0 | insufficient-diagnostic |
| NotYet | reading trailingNewlines | 0 | insufficient-diagnostic |
| NotYet | reading trueCondition | 0 | insufficient-diagnostic |
| NotYet | reading trueType | 0 | insufficient-diagnostic |
| NotYet | reading tryStatement | 0 | insufficient-diagnostic |
| NotYet | reading tupleType | 0 | insufficient-diagnostic |
| NotYet | reading typeArguments | 0 | insufficient-diagnostic |
| NotYet | reading typeCopy | 0 | insufficient-diagnostic |
| NotYet | reading typeFromTypeNode | 0 | insufficient-diagnostic |
| NotYet | reading typeKey | 0 | insufficient-diagnostic |
| NotYet | reading typeKind | 0 | insufficient-diagnostic |
| NotYet | reading typeLiteralNode | 0 | insufficient-diagnostic |
| NotYet | reading typeLiteralSymbol | 0 | insufficient-diagnostic |
| NotYet | reading typeName | 0 | insufficient-diagnostic |
| NotYet | reading typeNode | 0 | insufficient-diagnostic |
| NotYet | reading typeNodes | 0 | insufficient-diagnostic |
| NotYet | reading typeOnlyDeclaration | 0 | insufficient-diagnostic |
| NotYet | reading typeOnlyDeclarationIsExportStar | 0 | insufficient-diagnostic |
| NotYet | reading typeParameter | 0 | insufficient-diagnostic |
| NotYet | reading typeParameters | 0 | insufficient-diagnostic |
| NotYet | reading typeProperty | 0 | insufficient-diagnostic |
| NotYet | reading typeReferenceResolutions | 0 | insufficient-diagnostic |
| NotYet | reading typeReferenceResolutionsChanged | 0 | insufficient-diagnostic |
| NotYet | reading typeRoots | 0 | insufficient-diagnostic |
| NotYet | reading typeSet | 0 | insufficient-diagnostic |
| NotYet | reading typeSymbol | 0 | insufficient-diagnostic |
| NotYet | reading typeVariable | 0 | insufficient-diagnostic |
| NotYet | reading typedefTag | 0 | insufficient-diagnostic |
| NotYet | reading types | 0 | insufficient-diagnostic |
| NotYet | reading undefinedStrippedTarget | 0 | insufficient-diagnostic |
| NotYet | reading uniqueFilled | 0 | insufficient-diagnostic |
| NotYet | reading unmatched | 0 | insufficient-diagnostic |
| NotYet | reading unwidenedType | 0 | insufficient-diagnostic |
| NotYet | reading unwrappedExpr | 0 | insufficient-diagnostic |
| NotYet | reading updatedDecl | 0 | insufficient-diagnostic |
| NotYet | reading valid | 0 | insufficient-diagnostic |
| NotYet | reading validatedFilesSpec | 0 | insufficient-diagnostic |
| NotYet | reading validatedFilesSpecBeforeSubstitution | 0 | insufficient-diagnostic |
| NotYet | reading valueParam | 0 | insufficient-diagnostic |
| NotYet | reading valueSymbol | 0 | insufficient-diagnostic |
| NotYet | reading valueType | 0 | insufficient-diagnostic |
| NotYet | reading varName | 0 | insufficient-diagnostic |
| NotYet | reading variable | 0 | insufficient-diagnostic |
| NotYet | reading variableDeclarator | 0 | insufficient-diagnostic |
| NotYet | reading variableList | 0 | insufficient-diagnostic |
| NotYet | reading variances | 0 | insufficient-diagnostic |
| NotYet | reading verbatimTargetName | 0 | insufficient-diagnostic |
| NotYet | reading visibilityResult | 0 | insufficient-diagnostic |
| NotYet | reading visibleDefaultBinding | 0 | insufficient-diagnostic |
| NotYet | reading visitedAccessorName | 0 | insufficient-diagnostic |
| NotYet | reading visitedBindings | 0 | insufficient-diagnostic |
| NotYet | reading visitedNode | 0 | insufficient-diagnostic |
| NotYet | reading watchDirectoryKind | 0 | insufficient-diagnostic |
| NotYet | reading watchFileKind | 0 | insufficient-diagnostic |
| NotYet | reading widened | 0 | insufficient-diagnostic |
| NotYet | reading widenedTypes | 0 | insufficient-diagnostic |
| NotYet | reading yieldType | 0 | insufficient-diagnostic |
| Refused | an unproven relation from "" to T: the source is not assignable to the target | 0 | insufficient-diagnostic |
| Refused | an unproven relation from Declaration to T: the source is not assignable to the target | 0 | insufficient-diagnostic |
| Refused | an unproven relation from Identifier &#124; MissingDeclaration &#124; NumericLiteral &#124; StringLiteral &#124; TemplateLiteralLikeNode &#124; Token<...> to T: the source is not assignable to the target | 0 | insufficient-diagnostic |
| Refused | an unproven relation from string to T: the source is not assignable to the target | 0 | insufficient-diagnostic |
| checker | checker-rejected body | 1,613,177 | measurement-cause |
| SkippedDependency | a dependency function whose body has checker diagnostics (measurement skipped) | 87,831 | measurement-cause |
| panic | statement panic: Unhandled case in Node.Text: *ast.ComputedPropertyName | 3,647 | measurement-cause |
| conflict | [["NotYet", "a function returning T &#124; undefined"], ["NotYet", "a value of type Path"]] | 228 | measurement-cause |
| conflict | [["NotYet", "an array of InitializedPropertyDeclaration"], ["NotYet", "an array of ParameterPropertyDeclaration"], ["NotYet", "an array of T"], ["NotYet", "an array of unknown with erased element storage (retain its declared element type before reading elements)"]] | 98 | measurement-cause |
| conflict | [["NotYet", "an array of ElementWithComputedPropertyName"], ["NotYet", "an array of T"]] | 90 | measurement-cause |
| conflict | [["NotYet", "a value of type U &#124; undefined"], ["NotYet", "an array of ParameterPropertyDeclaration"]] | 49 | measurement-cause |
| conflict | [["NotYet", "a SpreadElement"], ["NotYet", "an array of T"]] | 44 | measurement-cause |
| conflict | [["NotYet", "an array of U"], ["NotYet", "an array of __String"]] | 43 | measurement-cause |
| conflict | [["NotYet", "a generic function as a value"], ["NotYet", "a value of type T"]] | 40 | measurement-cause |
| conflict | [["NotYet", "a rest parameter that isn't an array"], ["NotYet", "a value of type T"]] | 39 | measurement-cause |
| conflict | [["NotYet", "an array of T"], ["NotYet", "slice with an index that isn't a number"]] | 31 | measurement-cause |
| conflict | [["NotYet", "a value of type (GetAccessorDeclaration &#124; MethodDeclaration &#124; PropertyAssignment &#124; SetAccessorDeclaration &#124; ShorthandPropertyAssignment &#124; SpreadAssignment)[][] &#124; ... where an array goes"], ["NotYet", "a value of type DiagnosticRelatedInformation[][] &#124; readonly (DiagnosticRelatedInformation &#124; readonly DiagnosticRelatedInformation[] &#124; undefined)[] where an array goes"], ["NotYet", "a value of type Diagnostic[][] &#124; readonly (Diagnostic &#124; readonly Diagnostic[] &#124; undefined)[] where an array goes"], ["NotYet", "a value of type Extension[][] &#124; readonly (readonly Extension[] &#124; Extension &#124; undefined)[] where an array goes"], ["NotYet", "a value of type T &#124; T[] &#124; readonly T[] &#124; undefined"], ["NotYet", "a value of type VariableDeclaration[][] &#124; readonly (readonly VariableDeclaration[] &#124; VariableDeclaration &#124; undefined)[] where an array goes"], ["NotYet", "a value of type string[][] &#124; readonly (string &#124; readonly string[] &#124; undefined)[] where an array goes"]] | 29 | measurement-cause |
| panic | runtime error: invalid memory address or nil pointer dereference | 0 | measurement-cause |
| panic | statement panic: Unhandled case in Node.Text: *ast.BindingPattern | 0 | measurement-cause |
