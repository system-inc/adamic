Built: latest area/stage3 adaptations and newest cumulative-2 feature tips in a never-pushed scratch tree.
Meter: 36/78 checker-clean (46.15%), 382 checker diagnostics; measured on a checker-rejected program.
1-to-10 list: debug.ts (1), diagnosticInformationMap.generated.ts (1), visitorPublic.ts (1), builderState.ts (2), factory/emitNode.ts (2), performanceCore.ts (2), scanner.ts (2), sourcemap.ts (2), transformers/declarations.ts (2), watch.ts (2), emitter.ts (3), executeCommandLine.ts (3), moduleSpecifiers.ts (3), parser.ts (3), transformers/ts.ts (3), transformers/utilities.ts (3), factory/nodeFactory.ts (4), transformers/es2018.ts (4), transformers/module/module.ts (4), transformers/module/system.ts (4), watchUtilities.ts (4), transformers/es2015.ts (5), transformers/esDecorators.ts (5), transformers/jsx.ts (5), transformers/module/esnextAnd2015.ts (5), utilities.ts (5), tracing.ts (6), transformer.ts (6), transformers/esnext.ts (6), core.ts (7), transformers/es2017.ts (8), commandLineParser.ts (9), transformers/classFields.ts (9), moduleNameResolver.ts (10); measured on a checker-rejected program.
Checks: guarded build/run and vet pass; 12 report mutants, both output-guard mutants and binder-only +1 NotYet mutant caught; logs in data/meter3/.
Limits: own-file diagnostics only, first lowering error per unit, no native correctness or full-gate claim.

# Checker meter

Every count and delta below is **measured on a checker-rejected program**. The denominator is the same 78 generated compiler source roots as the previous run. A zero means no diagnostic is attributed to that file in one whole-project check; imports and the project can still be rejected.

34 files have 1–10 diagnostics (141 diagnostic sites). Fixing every listed file would reach 70/78 on this fixed corpus, if no new diagnostics appeared. Diagnostic count is a work queue, not an estimate of implementation difficulty.

# Checker-clean files

- `src/compiler/_namespaces/ts.moduleSpecifiers.ts`
- `src/compiler/_namespaces/ts.performance.ts`
- `src/compiler/_namespaces/ts.ts`
- `src/compiler/binder.ts`
- `src/compiler/builderPublic.ts`
- `src/compiler/builderStatePublic.ts`
- `src/compiler/corePublic.ts`
- `src/compiler/expressionToTypeNode.ts`
- `src/compiler/factory/baseNodeFactory.ts`
- `src/compiler/factory/emitHelpers.ts`
- `src/compiler/factory/nodeChildren.ts`
- `src/compiler/factory/nodeConverters.ts`
- `src/compiler/factory/nodeTests.ts`
- `src/compiler/factory/parenthesizerRules.ts`
- `src/compiler/factory/utilities.ts`
- `src/compiler/factory/utilitiesPublic.ts`
- `src/compiler/path.ts`
- `src/compiler/performance.ts`
- `src/compiler/programDiagnostics.ts`
- `src/compiler/semver.ts`
- `src/compiler/symbolWalker.ts`
- `src/compiler/transformers/classThis.ts`
- `src/compiler/transformers/declarations/diagnostics.ts`
- `src/compiler/transformers/destructuring.ts`
- `src/compiler/transformers/es2016.ts`
- `src/compiler/transformers/es2019.ts`
- `src/compiler/transformers/es2020.ts`
- `src/compiler/transformers/es2021.ts`
- `src/compiler/transformers/legacyDecorators.ts`
- `src/compiler/transformers/module/impliedNodeFormatDependent.ts`
- `src/compiler/transformers/namedEvaluation.ts`
- `src/compiler/transformers/taggedTemplate.ts`
- `src/compiler/transformers/typeSerializer.ts`
- `src/compiler/tsbuild.ts`
- `src/compiler/types.ts`
- `src/compiler/utilitiesPublic.ts`

# Files with 1–10 diagnostics

Locations refer to the final adapted tree. Each cause is the first line of the actual checker message; full chains and byte spans are preserved in JSON and compressed raw data.

## src/compiler/debug.ts (1)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2322 | 171:21 | Type 'undefined' is not assignable to type '{ level: AssertionLevel; assertion: AnyFunction; }'. |

## src/compiler/diagnosticInformationMap.generated.ts (1)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS1484 | 4:30 | 'DiagnosticMessage' is a type and must be imported using a type-only import when 'verbatimModuleSyntax' is enabled. |

## src/compiler/visitorPublic.ts (1)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2345 | 424:59 | Argument of type 'ParameterDeclaration \| undefined' is not assignable to parameter of type 'ParameterDeclaration'. |

## src/compiler/builderState.ts (2)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2345 | 493:108 | Argument of type 'Path \| undefined' is not assignable to parameter of type 'Path'. |
| TS2322 | 512:23 | Type '(Path \| undefined)[]' is not assignable to type 'Path[]'. |

## src/compiler/factory/emitNode.ts (2)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2769 | 205:108 | No overload matches this call. |
| TS2769 | 218:110 | No overload matches this call. |

## src/compiler/performanceCore.ts (2)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2591 | 35:37 | Cannot find name 'require'. Do you need to install type definitions for node? Try `npm i --save-dev @types/node` and then add 'node' to the types field in your tsconfig. |
| TS2591 | 35:84 | Cannot find name 'perf_hooks'. Do you need to install type definitions for node? Try `npm i --save-dev @types/node` and then add 'node' to the types field in your tsconfig. |

## src/compiler/scanner.ts (2)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2532 | 491:22 | Object is possibly 'undefined'. |
| TS2322 | 491:45 | Type 'number \| undefined' is not assignable to type 'number'. |

## src/compiler/sourcemap.ts (2)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2322 | 83:25 | Type 'string \| undefined' is not assignable to type 'string'. |
| TS2345 | 216:58 | Argument of type 'string \| null \| undefined' is not assignable to parameter of type 'string \| null'. |

## src/compiler/transformers/declarations.ts (2)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2532 | 1465:40 | Object is possibly 'undefined'. |
| TS2345 | 1678:110 | Argument of type 'ExpressionWithTypeArguments \| undefined' is not assignable to parameter of type 'DeclarationDiagnosticProducing'. |

## src/compiler/watch.ts (2)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2375 | 756:11 | Type '{ getSourceFile: (fileName: string, languageVersionOrOptions: CreateSourceFileOptions \| ScriptTarget, onError?: ((message: string) => void) \| undefined, shouldCreateNewSourceFile?: boolean \| undefined) => SourceFile \| undefined; ... 17 more ...; jsDocParsingMode: JSDocParsingMode \| undefined; }' is not assignable to type 'CompilerHost' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties. |
| TS2375 | 845:5 | Type '{ useCaseSensitiveFileNames: () => boolean; getNewLine: () => string; getCurrentDirectory: () => string; getDefaultLibLocation: () => string; getDefaultLibFileName: (options: CompilerOptions) => string; ... 13 more ...; now: ... \| undefined; }' is not assignable to type 'ProgramHost<T>' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties. |

## src/compiler/emitter.ts (3)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2379 | 862:55 | Argument of type '{ hasGlobalName: (name: string) => boolean; onEmitNode: (hint: EmitHint, node: Node, emitCallback: (hint: EmitHint, node: Node) => void) => void; isEmitNotificationEnabled: ... \| undefined; substituteNode: ...; }' is not assignable to parameter of type 'PrintHandlers' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties. |
| TS2379 | 938:70 | Argument of type '{ hasGlobalName: (name: string) => boolean; onEmitNode: (hint: EmitHint, node: Node, emitCallback: (hint: EmitHint, node: Node) => void) => void; isEmitNotificationEnabled: ... \| undefined; substituteNode: ...; }' is not assignable to parameter of type 'PrintHandlers' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties. |
| TS2322 | 1138:5 | Type 'string \| undefined' is not assignable to type 'string'. |

## src/compiler/executeCommandLine.ts (3)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2488 | 314:26 | Type '[string, string \| number] \| undefined' must have a '[Symbol.iterator]()' method that returns an iterator. |
| TS2488 | 315:23 | Type '[string, string \| number] \| undefined' must have a '[Symbol.iterator]()' method that returns an iterator. |
| TS18048 | 1180:105 | 'count' is possibly 'undefined'. |

## src/compiler/moduleSpecifiers.ts (3)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2322 | 349:5 | Type '["ambient" \| "node_modules" \| "paths" \| "redirect" \| "relative" \| undefined, readonly string[] \| undefined, SourceFile, readonly ModulePath[] \| undefined, ModuleSpecifierCache \| undefined]' is not assignable to type 'readonly [kind?: "ambient" \| "node_modules" \| "paths" \| "redirect" \| "relative" \| undefined, specifiers?: readonly string[], moduleFile?: SourceFile, modulePaths?: readonly ModulePath[], cache?: ModuleSpecifierCache]'. |
| TS2488 | 870:14 | Type '[string, { path: string; isRedirect: boolean; isInNodeModules: boolean; }] \| undefined' must have a '[Symbol.iterator]()' method that returns an iterator. |
| TS2345 | 1411:38 | Argument of type 'undefined' is not assignable to parameter of type 'never'. |

## src/compiler/parser.ts (3)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2322 | 10458:13 | Type '(position: number) => Node \| undefined' is not assignable to type '(position: number) => Node'. |
| TS2322 | 10656:17 | Type '{ name: string \| undefined; path: string; }[]' is not assignable to type 'AmdDependency[]'. |
| TS2322 | 10794:9 | Type 'string \| undefined' is not assignable to type 'string'. |

## src/compiler/transformers/ts.ts (3)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2538 | 1351:45 | Type 'undefined' cannot be used as an index type. |
| TS18048 | 1352:97 | 'superStatementIndex' is possibly 'undefined'. |
| TS18048 | 1379:80 | 'superStatementIndex' is possibly 'undefined'. |

## src/compiler/transformers/utilities.ts (3)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2375 | 689:5 | Type '{ decorators: readonly Decorator[] \| undefined; parameters: (readonly Decorator[] \| undefined)[] \| undefined; }' is not assignable to type 'AllDecorators' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties. |
| TS2375 | 749:5 | Type '{ decorators: readonly Decorator[] \| undefined; parameters: (readonly Decorator[] \| undefined)[] \| undefined; getDecorators: readonly Decorator[] \| undefined; setDecorators: ... \| undefined; }' is not assignable to type 'AllDecorators' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties. |
| TS2375 | 773:5 | Type '{ decorators: readonly Decorator[] \| undefined; parameters: (readonly Decorator[] \| undefined)[] \| undefined; }' is not assignable to type 'AllDecorators' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties. |

## src/compiler/factory/nodeFactory.ts (4)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2412 | 1216:9 | Type 'undefined' is not assignable to type 'T["localSymbol"]' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the type of the target. |
| TS2379 | 1315:41 | Argument of type '{ flags: GeneratedIdentifierFlags; id: number; prefix: string \| GeneratedNamePart \| undefined; suffix: string \| undefined; }' is not assignable to parameter of type 'AutoGenerateInfo' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties. |
| TS2379 | 1403:41 | Argument of type '{ flags: GeneratedIdentifierFlags; id: number; prefix: string \| GeneratedNamePart \| undefined; suffix: string \| undefined; }' is not assignable to parameter of type 'AutoGenerateInfo' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties. |
| TS2412 | 5496:9 | Type 'JSDocTypeExpression \| undefined' is not assignable to type 'T["typeExpression"]' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the type of the target. |

## src/compiler/transformers/es2018.ts (4)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2532 | 534:35 | Object is possibly 'undefined'. |
| TS2322 | 537:17 | Type 'Expression \| undefined' is not assignable to type 'Expression'. |
| TS2322 | 540:80 | Type 'Expression \| undefined' is not assignable to type 'Expression'. |
| TS2345 | 636:29 | Argument of type 'Expression \| undefined' is not assignable to parameter of type 'Expression'. |

## src/compiler/transformers/module/module.ts (4)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2345 | 246:42 | Argument of type 'Expression \| undefined' is not assignable to parameter of type 'FileReference \| Node'. |
| TS2532 | 382:72 | Object is possibly 'undefined'. |
| TS2322 | 2262:13 | Type 'ExternalModuleInfo \| undefined' is not assignable to type 'ExternalModuleInfo'. |
| TS2322 | 2481:21 | Type '(Identifier \| undefined)[]' is not assignable to type 'ModuleExportName[]'. |

## src/compiler/transformers/module/system.ts (4)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2532 | 289:21 | Object is possibly 'undefined'. |
| TS2322 | 1773:13 | Type 'ExternalModuleInfo \| undefined' is not assignable to type 'ExternalModuleInfo'. |
| TS2322 | 1774:13 | Type 'Identifier \| undefined' is not assignable to type 'Identifier'. |
| TS2322 | 1776:13 | Type 'Identifier \| undefined' is not assignable to type 'Identifier'. |

## src/compiler/watchUtilities.ts (4)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2375 | 125:5 | Type '{ useCaseSensitiveFileNames: boolean; fileExists: (fileName: string) => boolean; readFile: (path: string, encoding: string \| undefined) => string \| undefined; directoryExists: ... \| undefined; ... 7 more ...; realpath: ... \| undefined; }' is not assignable to type 'CachedDirectoryStructureHost' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties. |
| TS2684 | 736:13 | The 'this' context of type '((file: string, callback: FileWatcherCallback, pollingInterval: PollingInterval, options: WatchOptions \| undefined, detailInfo1: X, detailInfo2?: Y \| undefined) => FileWatcher) \| ...' is not assignable to method's 'this' of type '(this: undefined, file: string, callback: FileWatcherCallback, pollingInterval: PollingInterval, options: WatchOptions \| undefined, detailInfo1: X, detailInfo2?: Y \| undefined) => FileWatcher'. |
| TS2684 | 811:14 | The 'this' context of type '((file: string, callback: FileWatcherCallback, pollingInterval: PollingInterval, options: WatchOptions \| undefined, detailInfo1: X, detailInfo2?: Y \| undefined) => FileWatcher) \| ...' is not assignable to method's 'this' of type '(this: undefined, file: string, callback: FileWatcherCallback, pollingInterval: PollingInterval, options: WatchOptions \| undefined, detailInfo1: X, detailInfo2?: Y \| undefined) => FileWatcher'. |
| TS2556 | 818:48 | A spread argument must either have a tuple type or be passed to a rest parameter. |

## src/compiler/transformers/es2015.ts (5)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2532 | 2803:38 | Object is possibly 'undefined'. |
| TS2532 | 3046:64 | Object is possibly 'undefined'. |
| TS2345 | 4437:29 | Argument of type 'Statement \| undefined' is not assignable to parameter of type 'Statement'. |
| TS2532 | 4691:34 | Object is possibly 'undefined'. |
| TS2532 | 4693:13 | Object is possibly 'undefined'. |

## src/compiler/transformers/esDecorators.ts (5)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2322 | 351:9 | Type '{ kind: "class"; next: LexicalEnvironmentStackEntry \| undefined; classInfo: ClassInfo \| undefined; savedPendingExpressions: Expression[] \| undefined; }' is not assignable to type 'LexicalEnvironmentStackEntry \| undefined'. |
| TS2322 | 398:13 | Type '{ kind: "other"; next: ClassElementLexicalEnvironmentStackEntry \| ClassLexicalEnvironmentStackEntry \| PropertyNameLexicalEnvironmentStackEntry \| undefined; depth: number; savedPendingExpressions: ... \| undefined; }' is not assignable to type 'LexicalEnvironmentStackEntry \| undefined'. |
| TS2538 | 1164:45 | Type 'undefined' cannot be used as an index type. |
| TS18048 | 1165:97 | 'superStatementIndex' is possibly 'undefined'. |
| TS18048 | 1192:80 | 'superStatementIndex' is possibly 'undefined'. |

## src/compiler/transformers/jsx.ts (5)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2488 | 174:24 | Type '[string, Map<string, ImportSpecifier>] \| undefined' must have a '[Symbol.iterator]()' method that returns an iterator. |
| TS18046 | 187:172 | 's' is of type 'unknown'. |
| TS18046 | 187:188 | 's' is of type 'unknown'. |
| TS2345 | 309:58 | Argument of type 'JsxChild \| undefined' is not assignable to parameter of type 'JsxChild'. |
| TS2345 | 502:62 | Argument of type 'Expression \| undefined' is not assignable to parameter of type 'Node'. |

## src/compiler/transformers/module/esnextAnd2015.ts (5)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2345 | 100:46 | Argument of type 'Expression \| undefined' is not assignable to parameter of type 'FileReference \| Node'. |
| TS2345 | 163:91 | Argument of type 'CallExpression \| undefined' is not assignable to parameter of type 'TextRange'. |
| TS2345 | 194:37 | Argument of type 'Expression \| undefined' is not assignable to parameter of type 'FileReference \| Node'. |
| TS2345 | 196:81 | Argument of type 'Expression \| undefined' is not assignable to parameter of type 'Expression'. |
| TS2532 | 251:22 | Object is possibly 'undefined'. |

## src/compiler/utilities.ts (5)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2488 | 1223:22 | Type '[string, CommentDirective] \| undefined' must have a '[Symbol.iterator]()' method that returns an iterator. |
| TS2488 | 1224:19 | Type '[string, CommentDirective] \| undefined' must have a '[Symbol.iterator]()' method that returns an iterator. |
| TS2322 | 7721:17 | Type 'number \| undefined' is not assignable to type 'number'. |
| TS2322 | 7725:17 | Type 'number \| undefined' is not assignable to type 'number'. |
| TS2345 | 11201:34 | Argument of type 'string' is not assignable to parameter of type '{ [Symbol.replace](string: string, replacer: (substring: string, ...args: any[]) => string): string; }'. |

## src/compiler/tracing.ts (6)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2591 | 40:27 | Cannot find name 'fs'. Do you need to install type definitions for node? Try `npm i --save-dev @types/node` and then add 'node' to the types field in your tsconfig. |
| TS2591 | 63:22 | Cannot find name 'require'. Do you need to install type definitions for node? Try `npm i --save-dev @types/node` and then add 'node' to the types field in your tsconfig. |
| TS18046 | 66:81 | 'e' is of type 'unknown'. |
| TS2591 | 82:50 | Cannot find name 'process'. Do you need to install type definitions for node? Try `npm i --save-dev @types/node` and then add 'node' to the types field in your tsconfig. |
| TS2591 | 83:39 | Cannot find name 'process'. Do you need to install type definitions for node? Try `npm i --save-dev @types/node` and then add 'node' to the types field in your tsconfig. |
| TS2412 | 122:13 | Type 'undefined' is not assignable to type 'string' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the type of the target. |

## src/compiler/transformer.ts (6)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2532 | 367:9 | Object is possibly 'undefined'. |
| TS2532 | 395:9 | Object is possibly 'undefined'. |
| TS2322 | 563:9 | Type 'VariableDeclaration[] \| undefined' is not assignable to type 'VariableDeclaration[]'. |
| TS2322 | 564:9 | Type 'FunctionDeclaration[] \| undefined' is not assignable to type 'FunctionDeclaration[]'. |
| TS2322 | 565:9 | Type 'Statement[] \| undefined' is not assignable to type 'Statement[]'. |
| TS2322 | 614:9 | Type 'Identifier[] \| undefined' is not assignable to type 'Identifier[]'. |

## src/compiler/transformers/esnext.ts (6)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2345 | 183:34 | Argument of type 'Statement \| undefined' is not assignable to parameter of type 'Statement'. |
| TS2345 | 205:52 | Argument of type '(ExportSpecifier \| undefined)[]' is not assignable to parameter of type 'readonly ExportSpecifier[]'. |
| TS2345 | 344:44 | Argument of type 'Statement \| undefined' is not assignable to parameter of type 'Statement'. |
| TS2345 | 384:36 | Argument of type 'Statement \| undefined' is not assignable to parameter of type 'Node'. |
| TS2345 | 767:34 | Argument of type 'Statement \| undefined' is not assignable to parameter of type 'Node'. |
| TS2345 | 767:70 | Argument of type 'Statement \| undefined' is not assignable to parameter of type 'Statement'. |

## src/compiler/core.ts (7)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2345 | 992:21 | Argument of type 'T \| undefined' is not assignable to parameter of type 'T'. |
| TS2740 | 1637:11 | Type '{ has(element: TElement): boolean; add(element: TElement): Set<TElement>; delete(element: TElement): boolean; clear(): void; ... 6 more ...; [Symbol.toStringTag]: string; }' is missing the following properties from type 'Set<TElement>': union, intersection, difference, symmetricDifference, and 3 more. |
| TS2345 | 1721:28 | Argument of type 'TElement \| undefined' is not assignable to parameter of type 'TElement'. |
| TS2591 | 2591:19 | Cannot find name 'process'. Do you need to install type definitions for node? Try `npm i --save-dev @types/node` and then add 'node' to the types field in your tsconfig. |
| TS2591 | 2592:14 | Cannot find name 'process'. Do you need to install type definitions for node? Try `npm i --save-dev @types/node` and then add 'node' to the types field in your tsconfig. |
| TS2591 | 2593:14 | Cannot find name 'process'. Do you need to install type definitions for node? Try `npm i --save-dev @types/node` and then add 'node' to the types field in your tsconfig. |
| TS2591 | 2594:19 | Cannot find name 'require'. Do you need to install type definitions for node? Try `npm i --save-dev @types/node` and then add 'node' to the types field in your tsconfig. |

## src/compiler/transformers/es2017.ts (8)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2322 | 325:21 | Type 'Set<__String \| undefined>' is not assignable to type 'Set<__String>'. |
| TS18048 | 327:17 | 'catchClauseUnshadowedNames' is possibly 'undefined'. |
| TS2532 | 599:86 | Object is possibly 'undefined'. |
| TS18048 | 764:38 | 'outerParameter' is possibly 'undefined'. |
| TS18048 | 765:25 | 'originalParameter' is possibly 'undefined'. |
| TS18048 | 765:58 | 'originalParameter' is possibly 'undefined'. |
| TS18048 | 767:76 | 'outerParameter' is possibly 'undefined'. |
| TS18048 | 770:44 | 'outerParameter' is possibly 'undefined'. |

## src/compiler/commandLineParser.ts (9)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2345 | 1874:96 | Argument of type 'string \| undefined' is not assignable to parameter of type 'string'. |
| TS2345 | 2077:79 | Argument of type 'string \| undefined' is not assignable to parameter of type 'string'. |
| TS18046 | 2301:91 | 'e' is of type 'unknown'. |
| TS2375 | 2663:9 | Type '{ showConfig: undefined; configFile: undefined; configFilePath: undefined; help: undefined; init: undefined; listFiles: undefined; listEmittedFiles: undefined; project: undefined; build: undefined; version: undefined; }' is not assignable to type 'CompilerOptions' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties. |
| TS2322 | 2677:9 | Type '{ prepend?: boolean \| undefined; circular?: boolean \| undefined; path: string; originalPath: undefined; }[] \| undefined' is not assignable to type 'readonly ProjectReference[] \| undefined'. |
| TS2345 | 2689:66 | Argument of type 'Set<string \| undefined>' is not assignable to parameter of type 'Set<string>'. |
| TS2322 | 2952:13 | Type 'string \| undefined' is not assignable to type 'string'. |
| TS2322 | 3449:55 | Type '(string \| undefined)[]' is not assignable to type 'string[]'. |
| TS2322 | 4014:5 | Type '(string \| undefined)[]' is not assignable to type 'string[]'. |

## src/compiler/transformers/classFields.ts (9)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2532 | 1835:21 | Object is possibly 'undefined'. |
| TS2538 | 2251:45 | Type 'undefined' cannot be used as an index type. |
| TS18048 | 2252:97 | 'superStatementIndex' is possibly 'undefined'. |
| TS18048 | 2253:27 | 'superStatementIndex' is possibly 'undefined'. |
| TS2345 | 2297:52 | Argument of type 'Node \| undefined' is not assignable to parameter of type 'Node'. |
| TS2345 | 2376:56 | Argument of type 'Node \| undefined' is not assignable to parameter of type 'Node'. |
| TS2375 | 2745:16 | Type 'PrivateEnvironment<{ className: undefined; weakSetName: undefined; }, PrivateIdentifierInfo>' is not assignable to type 'PrivateEnvironment<PrivateEnvironmentData, PrivateIdentifierInfo>' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties. |
| TS2345 | 2864:52 | Argument of type '{ kind: PrivateIdentifierKind.Accessor; getterName: Identifier; setterName: undefined; brandCheckIdentifier: Identifier; isStatic: boolean; isValid: boolean; }' is not assignable to parameter of type 'PrivateIdentifierInfo'. |
| TS2345 | 2896:52 | Argument of type '{ kind: PrivateIdentifierKind.Accessor; getterName: undefined; setterName: Identifier; brandCheckIdentifier: Identifier; isStatic: boolean; isValid: boolean; }' is not assignable to parameter of type 'PrivateIdentifierInfo'. |

## src/compiler/moduleNameResolver.ts (10)

| Code | Line:column | One-line cause |
| --- | --- | --- |
| TS2375 | 132:13 | Type '{ name: string; subModuleName: string; version: string; peerDependencies: string \| undefined; }' is not assignable to type 'PackageId' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties. |
| TS2322 | 286:9 | Type '{ resolvedFileName: string; originalPath: string \| undefined; extension: string; isExternalLibraryImport: boolean \| undefined; packageId: PackageId \| undefined; resolvedUsingTsExtension: boolean; } \| undefined' is not assignable to type 'ResolvedModuleFull \| undefined'. |
| TS2322 | 474:36 | Type 'MapLike<string[]> \| undefined' is not assignable to type 'MapLike<string[]>'. |
| TS2375 | 627:9 | Type '{ primary: boolean; resolvedFileName: string; originalPath: string \| undefined; packageId: PackageId \| undefined; isExternalLibraryImport: boolean; }' is not assignable to type 'ResolvedTypeReferenceDirective' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties. |
| TS2375 | 635:5 | Type '{ resolvedTypeReferenceDirective: ResolvedTypeReferenceDirective \| undefined; failedLookupLocations: string[] \| undefined; affectingLocations: ... \| undefined; resolutionDiagnostics: ... \| undefined; }' is not assignable to type 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties. |
| TS2345 | 642:143 | Argument of type 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations \| undefined' is not assignable to parameter of type 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations'. |
| TS2345 | 644:144 | Argument of type 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations \| undefined' is not assignable to parameter of type 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations'. |
| TS2345 | 647:35 | Argument of type 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations \| undefined' is not assignable to parameter of type 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations'. |
| TS2322 | 648:5 | Type 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations \| undefined' is not assignable to type 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations'. |
| TS2322 | 2772:21 | Type 'SearchResult<{ path: string; extension: string; packageId: PackageId \| undefined; originalPath: string \| undefined; resolvedUsingTsExtension: boolean \| undefined; }>' is not assignable to type 'SearchResult<Resolved>'. |

# Totals by code

| Code | Diagnostics |
| --- | ---: |
| TS1484 | 1 |
| TS18046 | 8 |
| TS18048 | 52 |
| TS2304 | 6 |
| TS2307 | 1 |
| TS2322 | 48 |
| TS2339 | 3 |
| TS2345 | 95 |
| TS2375 | 21 |
| TS2379 | 14 |
| TS2412 | 28 |
| TS2420 | 1 |
| TS2488 | 11 |
| TS2532 | 20 |
| TS2538 | 3 |
| TS2556 | 2 |
| TS2591 | 54 |
| TS2684 | 2 |
| TS2722 | 1 |
| TS2740 | 1 |
| TS2769 | 8 |
| TS7006 | 1 |
| TS7031 | 1 |

# All files

| File | Checker diagnostics | NotYet | Refused |
| --- | ---: | ---: | ---: |
| src/compiler/_namespaces/ts.moduleSpecifiers.ts | 0 | 0 | 0 |
| src/compiler/_namespaces/ts.performance.ts | 0 | 0 | 0 |
| src/compiler/_namespaces/ts.ts | 0 | 0 | 0 |
| src/compiler/binder.ts | 0 | 4 | 255 |
| src/compiler/builder.ts | 17 | 23 | 85 |
| src/compiler/builderPublic.ts | 0 | 6 | 0 |
| src/compiler/builderState.ts | 2 | 1 | 15 |
| src/compiler/builderStatePublic.ts | 0 | 0 | 0 |
| src/compiler/checker.ts | 89 | 5 | 13 |
| src/compiler/commandLineParser.ts | 9 | 49 | 82 |
| src/compiler/core.ts | 7 | 225 | 146 |
| src/compiler/corePublic.ts | 0 | 0 | 1 |
| src/compiler/debug.ts | 1 | 14 | 104 |
| src/compiler/diagnosticInformationMap.generated.ts | 1 | 1 | 0 |
| src/compiler/emitter.ts | 3 | 10 | 382 |
| src/compiler/executeCommandLine.ts | 3 | 15 | 23 |
| src/compiler/expressionToTypeNode.ts | 0 | 3 | 39 |
| src/compiler/factory/baseNodeFactory.ts | 0 | 1 | 0 |
| src/compiler/factory/emitHelpers.ts | 0 | 5 | 10 |
| src/compiler/factory/emitNode.ts | 2 | 24 | 8 |
| src/compiler/factory/nodeChildren.ts | 0 | 4 | 2 |
| src/compiler/factory/nodeConverters.ts | 0 | 1 | 11 |
| src/compiler/factory/nodeFactory.ts | 4 | 7 | 6 |
| src/compiler/factory/nodeTests.ts | 0 | 0 | 227 |
| src/compiler/factory/parenthesizerRules.ts | 0 | 2 | 22 |
| src/compiler/factory/utilities.ts | 0 | 42 | 120 |
| src/compiler/factory/utilitiesPublic.ts | 0 | 1 | 2 |
| src/compiler/moduleNameResolver.ts | 10 | 44 | 47 |
| src/compiler/moduleSpecifiers.ts | 3 | 11 | 30 |
| src/compiler/parser.ts | 3 | 38 | 170 |
| src/compiler/path.ts | 0 | 31 | 8 |
| src/compiler/performance.ts | 0 | 6 | 0 |
| src/compiler/performanceCore.ts | 2 | 1 | 1 |
| src/compiler/program.ts | 13 | 19 | 43 |
| src/compiler/programDiagnostics.ts | 0 | 0 | 29 |
| src/compiler/resolutionCache.ts | 12 | 12 | 8 |
| src/compiler/scanner.ts | 2 | 30 | 35 |
| src/compiler/semver.ts | 0 | 5 | 9 |
| src/compiler/sourcemap.ts | 2 | 10 | 27 |
| src/compiler/symbolWalker.ts | 0 | 1 | 7 |
| src/compiler/sys.ts | 57 | 14 | 27 |
| src/compiler/tracing.ts | 6 | 3 | 9 |
| src/compiler/transformer.ts | 6 | 2 | 3 |
| src/compiler/transformers/classFields.ts | 9 | 4 | 3 |
| src/compiler/transformers/classThis.ts | 0 | 4 | 4 |
| src/compiler/transformers/declarations.ts | 2 | 2 | 8 |
| src/compiler/transformers/declarations/diagnostics.ts | 0 | 2 | 15 |
| src/compiler/transformers/destructuring.ts | 0 | 8 | 29 |
| src/compiler/transformers/es2015.ts | 5 | 0 | 0 |
| src/compiler/transformers/es2016.ts | 0 | 1 | 11 |
| src/compiler/transformers/es2017.ts | 8 | 1 | 1 |
| src/compiler/transformers/es2018.ts | 4 | 0 | 0 |
| src/compiler/transformers/es2019.ts | 0 | 1 | 1 |
| src/compiler/transformers/es2020.ts | 0 | 1 | 12 |
| src/compiler/transformers/es2021.ts | 0 | 1 | 0 |
| src/compiler/transformers/esDecorators.ts | 5 | 0 | 0 |
| src/compiler/transformers/esnext.ts | 6 | 0 | 1 |
| src/compiler/transformers/generators.ts | 27 | 1 | 1 |
| src/compiler/transformers/jsx.ts | 5 | 0 | 0 |
| src/compiler/transformers/legacyDecorators.ts | 0 | 1 | 19 |
| src/compiler/transformers/module/esnextAnd2015.ts | 5 | 0 | 0 |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts | 0 | 0 | 0 |
| src/compiler/transformers/module/module.ts | 4 | 0 | 0 |
| src/compiler/transformers/module/system.ts | 4 | 0 | 0 |
| src/compiler/transformers/namedEvaluation.ts | 0 | 15 | 7 |
| src/compiler/transformers/taggedTemplate.ts | 0 | 1 | 3 |
| src/compiler/transformers/ts.ts | 3 | 0 | 0 |
| src/compiler/transformers/typeSerializer.ts | 0 | 1 | 11 |
| src/compiler/transformers/utilities.ts | 3 | 19 | 50 |
| src/compiler/tsbuild.ts | 0 | 1 | 0 |
| src/compiler/tsbuildPublic.ts | 14 | 44 | 51 |
| src/compiler/types.ts | 0 | 1 | 11 |
| src/compiler/utilities.ts | 5 | 238 | 554 |
| src/compiler/utilitiesPublic.ts | 0 | 30 | 190 |
| src/compiler/visitorPublic.ts | 1 | 33 | 18 |
| src/compiler/watch.ts | 2 | 18 | 41 |
| src/compiler/watchPublic.ts | 12 | 5 | 2 |
| src/compiler/watchUtilities.ts | 4 | 6 | 4 |

# Lowering and variance

Unique lowering sites: {'NotYet': 1109, 'Refused': 3053, 'SkippedDependency': 4, 'error': 0, 'panic': 0}. Functions attempted: 2677; bodies skipped for own-body checker diagnostics: 78. Refusal scanning and independent top-level-unit lowering continue beyond failures. This remains a measurement-only overlay; no usable IR or production loader output is allowed.

Variance: 676 Refused sites across 408 nearest named owning declarations. Families: {'method-parameter-bivariance': 64, 'mutable-invariance': 612}. JSON retains the complete per-declaration ledger and attempting-unit context.

| Exact family | Sites |
| --- | ---: |
| NotYet: .length on a union of differently held members | 1 |
| NotYet: .length on a value | 1 |
| NotYet: ?. to a boolean \| undefined, which would be boolean \| undefined \| undefined | 1 |
| NotYet: ?.[] on a value | 1 |
| NotYet: ?? whose sides have different types | 1 |
| NotYet: Object.assign on a shape not proven by a plain literal or its const binding | 1 |
| NotYet: Object.entries on a shape not proven by a plain literal or its const binding | 3 |
| NotYet: RegExp with a nonconstant pattern | 1 |
| NotYet: String as a value outside equality or typeof (overloaded calls and static properties need their own representation) | 1 |
| NotYet: a BinaryExpression with a string and a number | 2 |
| NotYet: a BinaryExpression with a union of differently held members and a union of differently held members | 1 |
| NotYet: a ClassExpression | 1 |
| NotYet: a Map of T | 1 |
| NotYet: a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions | 6 |
| NotYet: a NonNullExpression | 84 |
| NotYet: a PostfixUnaryExpression | 1 |
| NotYet: a PrefixUnaryExpression on a number | 1 |
| NotYet: a SpreadElement | 1 |
| NotYet: a YieldExpression as a statement | 1 |
| NotYet: a boolean \| undefined variable a function value captures | 1 |
| NotYet: a call returning T | 1 |
| NotYet: a call through ?. (an optional call) | 14 |
| NotYet: a case whose type differs from the switch's | 4 |
| NotYet: a computed field name | 4 |
| NotYet: a conditional whose branches have different types | 2 |
| NotYet: a declaration directly in a case (wrap the case in a block) | 4 |
| NotYet: a destructured parameter beside a parameter with a default | 1 |
| NotYet: a field from a boolean \| undefined variable | 1 |
| NotYet: a function returning (AssignmentExpression<EqualsToken> & { readonly left: GeneratedIdentifier; }) \| undefined | 1 |
| NotYet: a function returning (ConstructorDeclaration & { body: Block; }) \| undefined | 1 |
| NotYet: a function returning AnyValidImportOrReExport | 1 |
| NotYet: a function returning AnyValidImportOrReExport \| undefined | 1 |
| NotYet: a function returning CanonicalKey | 1 |
| NotYet: a function returning ClassNamedEvaluationHelperBlock | 1 |
| NotYet: a function returning ClassThisAssignmentBlock | 1 |
| NotYet: a function returning CompilerOptionsValue | 3 |
| NotYet: a function returning ExpressionWithTypeArguments & { readonly expression: Identifier \| PropertyAccessEntityNameExpression; } | 1 |
| NotYet: a function returning Extract<AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; }, Pick<...>> \| ... 7 more ... \| Extract<...> | 1 |
| NotYet: a function returning Extract<ClassDeclaration, Pick<...>> \| Extract<...> | 2 |
| NotYet: a function returning HasJSDoc \| undefined | 1 |
| NotYet: a function returning MemberName \| (Expression & (NumericLiteral \| StringLiteralLike)) | 1 |
| NotYet: a function returning ModeAwareCacheKey | 2 |
| NotYet: a function returning Node \| (TIn & undefined) \| (TVisited & undefined) | 1 |
| NotYet: a function returning NodeArray<Node> \| (TInArray & undefined) | 1 |
| NotYet: a function returning NodeArray<TOut> \| (TInArray & undefined) | 1 |
| NotYet: a function returning PackageJson[K] \| undefined | 4 |
| NotYet: a function returning Path | 5 |
| NotYet: a function returning Path \| undefined | 2 |
| NotYet: a function returning PathPathComponents | 1 |
| NotYet: a function returning ResolvedConfigFileName | 3 |
| NotYet: a function returning ResolvedConfigFilePath | 1 |
| NotYet: a function returning T | 50 |
| NotYet: a function returning T \| T[] | 1 |
| NotYet: a function returning T \| T[] \| undefined | 3 |
| NotYet: a function returning T \| readonly T[] | 1 |
| NotYet: a function returning T \| readonly T[] \| undefined | 3 |
| NotYet: a function returning T \| undefined | 72 |
| NotYet: a function returning T1 & T2 | 1 |
| NotYet: a function returning TEntry \| undefined | 1 |
| NotYet: a function returning TOut | 1 |
| NotYet: a function returning TOut \| (TIn & undefined) \| (TVisited & undefined) | 1 |
| NotYet: a function returning TOut \| undefined | 1 |
| NotYet: a function returning TPrivateEntry \| undefined | 1 |
| NotYet: a function returning TResult | 1 |
| NotYet: a function returning U | 2 |
| NotYet: a function returning U \| undefined | 17 |
| NotYet: a function returning V | 2 |
| NotYet: a function returning __String | 10 |
| NotYet: a function returning __String \| undefined | 4 |
| NotYet: a function returning any | 6 |
| NotYet: a function returning object | 2 |
| NotYet: a function returning object \| undefined | 1 |
| NotYet: a function returning readonly Node[] \| (TInArray & undefined) | 1 |
| NotYet: a function returning readonly TOut[] \| (TInArray & undefined) | 1 |
| NotYet: a function returning string \| object | 1 |
| NotYet: a function returning unknown | 1 |
| NotYet: a function value returning boolean \| undefined | 2 |
| NotYet: a function value returning union of differently held members | 8 |
| NotYet: a function value taking string \| DiagnosticMessageChain \| undefined | 1 |
| NotYet: a function value taking string \| string[] | 1 |
| NotYet: a function value with an optional parameter | 10 |
| NotYet: a function with an optional or rest parameter, as a value | 5 |
| NotYet: a function without a body | 198 |
| NotYet: a generic function as a value | 15 |
| NotYet: a method call through a structural signature in a program with statics; use typeof the declaring class | 123 |
| NotYet: a namespace merged with a function; callable object properties, identity and receivers are not represented | 1 |
| NotYet: a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports | 5 |
| NotYet: a narrowed scalar in a boxed union field | 1 |
| NotYet: a parameter that isn't a plain name | 24 |
| NotYet: a tagged template other than the intrinsic String.raw | 2 |
| NotYet: a template interpolating an object, an array, a map, a function or undefined | 2 |
| NotYet: a union of differently held members variable a function value captures | 1 |
| NotYet: a value of type "" \| ResolvedConfigFileName \| undefined | 1 |
| NotYet: a value of type (AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; } & BinaryExpression) \| (... & ... 1 more ... & BinaryExpression) | 1 |
| NotYet: a value of type (ConstructorDeclaration & { body: Block; }) \| undefined | 1 |
| NotYet: a value of type (EmitNode & { autoGenerate: AutoGenerateInfo; }) \| (EmitNode & { autoGenerate: AutoGenerateInfo; }) | 2 |
| NotYet: a value of type (ModuleDeclaration & { name: StringLiteral; }) \| undefined | 1 |
| NotYet: a value of type AccessExpression \| RequireOrImportCall | 1 |
| NotYet: a value of type AccessorDeclaration & { readonly name: BigIntLiteral \| ComputedPropertyName \| Identifier \| NoSubstitutionTemplateLiteral \| NumericLiteral \| StringLiteral; } | 1 |
| NotYet: a value of type BindableStaticNameExpression | 2 |
| NotYet: a value of type BindingElement & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type CompilerOptionsValue | 4 |
| NotYet: a value of type EmitNode & { autoGenerate: AutoGenerateInfo; } | 1 |
| NotYet: a value of type EntityNameExpression \| (LeftHandSideExpression & BindableStaticNameExpression) | 1 |
| NotYet: a value of type ExportAssignment & { readonly expression: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type ExpressionWithTypeArguments & { readonly expression: Identifier \| PropertyAccessEntityNameExpression; } | 1 |
| NotYet: a value of type HasJSDoc | 2 |
| NotYet: a value of type HasJSDoc \| undefined | 1 |
| NotYet: a value of type IncludeTypeSpaceImports | 1 |
| NotYet: a value of type IncrementalBuildInfoFileId | 1 |
| NotYet: a value of type IncrementalBuildInfoFilePendingEmit | 1 |
| NotYet: a value of type IncrementalMultiFileEmitBuildInfoFileInfo | 1 |
| NotYet: a value of type JSDocImportTag \| CanHaveModuleSpecifier | 1 |
| NotYet: a value of type K | 3 |
| NotYet: a value of type NamedEvaluation | 1 |
| NotYet: a value of type NodeArray<Expression> & readonly [BindableStaticNameExpression, NumericLiteral \| StringLiteralLike, ObjectLiteralExpression] & Readonly<...> | 1 |
| NotYet: a value of type NonNullable<T> | 5 |
| NotYet: a value of type ParameterDeclaration & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type Path | 21 |
| NotYet: a value of type PropertyAssignment & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type PropertyDeclaration & { readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type RequireOrImportCall | 1 |
| NotYet: a value of type ResolvedConfigFileName | 8 |
| NotYet: a value of type ResolvedConfigFileName \| undefined | 1 |
| NotYet: a value of type ResolvedConfigFilePath | 18 |
| NotYet: a value of type ResolvedModuleWithFailedLookupLocations & ResolvedTypeReferenceDirectiveWithFailedLookupLocations | 1 |
| NotYet: a value of type ShorthandPropertyAssignment & { readonly objectAssignmentInitializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type SourceFile | 1 |
| NotYet: a value of type T | 28 |
| NotYet: a value of type T \| Program | 4 |
| NotYet: a value of type T \| T[] | 2 |
| NotYet: a value of type T \| null \| undefined | 1 |
| NotYet: a value of type T \| readonly T[] | 1 |
| NotYet: a value of type T \| undefined | 11 |
| NotYet: a value of type T1 | 1 |
| NotYet: a value of type TData | 1 |
| NotYet: a value of type TEntry | 1 |
| NotYet: a value of type TInArray | 2 |
| NotYet: a value of type T["kind"] | 1 |
| NotYet: a value of type TypeNode & LiteralTypeNode & { readonly literal: StringLiteral; } | 1 |
| NotYet: a value of type TypeParameterDeclaration & { parent: JSDocTemplateTag; } | 1 |
| NotYet: a value of type U | 1 |
| NotYet: a value of type U \| readonly U[] \| undefined | 1 |
| NotYet: a value of type V | 1 |
| NotYet: a value of type VariableDeclaration & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type WatchFactoryHost & { trace?(s: string): void; } | 1 |
| NotYet: a value of type WrappedExpression<AnonymousFunctionDefinition> | 2 |
| NotYet: a value of type WrappedExpression<T> | 1 |
| NotYet: a value of type __String | 28 |
| NotYet: a value of type __String & string | 1 |
| NotYet: a value of type any | 33 |
| NotYet: a value of type false \| RegExpExecArray \| null | 1 |
| NotYet: a value of type never | 1 |
| NotYet: a value of type object | 6 |
| NotYet: a value of type object \| undefined | 1 |
| NotYet: a value of type string \| (void & { __escapedIdentifier: void; }) \| (string & { __escapedIdentifier: void; }) | 3 |
| NotYet: a value of type string \| null \| undefined | 2 |
| NotYet: a value of type unknown | 11 |
| NotYet: an ElementAccessExpression | 6 |
| NotYet: an array of T | 7 |
| NotYet: an array of never | 19 |
| NotYet: an array of unknown | 1 |
| NotYet: an assignment value to a member | 2 |
| NotYet: an enum inside a function or block; declare it at module scope | 3 |
| NotYet: an optional chain longer than one step | 4 |
| NotYet: assigning a field of a value | 2 |
| NotYet: assigning to an ObjectLiteralExpression | 1 |
| NotYet: for...in over an array (holes and own enumerable properties are not represented; use for...of for elements) | 1 |
| NotYet: for...in without a proven fixed plain-object origin (arrays, prototypes and absent synthetic fields cannot be enumerated soundly) | 4 |
| NotYet: for...of over an object | 11 |
| NotYet: lastIndexOf with these arguments | 2 |
| NotYet: new Map from something that isn't [key, value] pairs | 1 |
| NotYet: new a ParenthesizedExpression | 3 |
| NotYet: new an Identifier | 9 |
| NotYet: optional chaining to .size on a value | 4 |
| NotYet: reading Error | 1 |
| NotYet: reading getOptionsNameMap | 2 |
| NotYet: regex replacement other than a string | 6 |
| NotYet: spreading an array of other elements | 1 |
| NotYet: storing any in a field | 1 |
| NotYet: this outside a method | 7 |
| NotYet: writing a possibly absent optional own field | 6 |
| Refused: Object.defineProperty | 1 |
| Refused: a cast the runtime can't check | 71 |
| Refused: a definite assignment assertion ! | 11 |
| Refused: a flag initializer outside the non-negative int32 bound | 3 |
| Refused: a function taking () => T seen as one taking () => T (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 |
| Refused: a function taking (node: Node) => T \| undefined seen as one taking (node: Node) => T \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking (symbol: Symbol) => boolean seen as one taking ((symbol: Symbol) => boolean) \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking (value: never, key: never, map: ReadonlyMap<never, never>) => void seen as one taking <TKey extends keyof PragmaPseudoMap>(value: PragmaPseudoMap[TKey][] \| PragmaPseudoMap[TKey], key: TKey, map: ReadonlyPragmaMap) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking BinaryExpressionStateMachine<TOuterState, TState, TResult> seen as one taking BinaryExpressionStateMachine<TOuterState, TState, TResult> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 6 |
| Refused: a function taking BinaryOperator seen as one taking SyntaxKind (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking CallExpression \| (IncludeTypeSpaceImports extends false ? never : ImportTypeNode \| JSDocImportTag) seen as one taking CallExpression \| ImportTypeNode \| JSDocImportTag (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking Expression seen as one taking Expression \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking Expression[] seen as one taking readonly Expression[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking LogLevel seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking Node seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking Node \| undefined seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking NodeArray<Expression> seen as one taking readonly Expression[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking NodeArray<T> seen as one taking NodeArray<T> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking NodeArray<T> seen as one taking NodeArray<T> \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking NodeArray<TypeNode> \| undefined seen as one taking readonly TypeNode[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking OrdinalParentheizerRuleSelector<Node> \| undefined seen as one taking ParenthesizerRuleOrSelector<T> \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking ParenthesizerRule<Node> \| undefined seen as one taking ParenthesizerRuleOrSelector<T> \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking ParenthesizerRuleOrSelector<Node> \| undefined seen as one taking ParenthesizerRuleOrSelector<T> \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking PollingInterval seen as one taking number \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking SourceFile seen as one taking [file: SourceFile, options: CompilerOptions] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking T seen as one taking Expression (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking T seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking TIn seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking TypeNode seen as one taking Node (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking Visitor seen as one taking Visitor<TIn, Node \| undefined> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [callback: (...args: any[]) => void, ms: number, ...args: any[]] seen as one taking (...args: any[]) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [fileName: string, languageVersionOrOptions: CreateSourceFileOptions \| ScriptTarget, onError?: ((message: string) => void) \| undefined, shouldCreateNewSourceFile?: boolean \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking [fileName: string, text: string, writeByteOrderMark: boolean, onError?: ((message: string) => void) \| undefined, sourceFiles?: readonly SourceFile[] \| undefined, data?: WriteFileCallbackData \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 |
| Refused: a function taking [fileName: string] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [path: string, callback: DirectoryWatcherCallback, recursive?: boolean \| undefined, options?: WatchOptions \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [path: string, callback: FileWatcherCallback, pollingInterval?: number \| undefined, options?: WatchOptions \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [timeoutId: any] seen as one taking unknown (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [writeByteOrderMark: boolean, onError?: ((message: string) => void) \| undefined, sourceFiles?: readonly SourceFile[] \| undefined, data?: WriteFileCallbackData \| undefined] seen as one taking boolean (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking boolean seen as one taking boolean \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking never seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking number seen as one taking any[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking number seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 |
| Refused: a function taking readonly ParameterDeclaration[] seen as one taking readonly ParameterDeclaration[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking string seen as one taking [fileName: string, text: string, writeByteOrderMark: boolean, onError?: ((message: string) => void) \| undefined, sourceFiles?: readonly SourceFile[] \| undefined, data?: WriteFileCallbackData \| undefined] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 |
| Refused: a function taking string seen as one taking [fileName: string] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking string seen as one taking any[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking string \| undefined seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a generator function | 5 |
| Refused: a method in object destructuring | 3 |
| Refused: a method read as a value (add would lose its object, and this with it) | 1 |
| Refused: a method read as a value (afterProgramCreate would lose its object, and this with it) | 3 |
| Refused: a method read as a value (afterProgramEmitAndDiagnostics would lose its object, and this with it) | 3 |
| Refused: a method read as a value (base64decode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (base64encode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (clearScreen would lose its object, and this with it) | 1 |
| Refused: a method read as a value (clearTimeout would lose its object, and this with it) | 2 |
| Refused: a method read as a value (compare would lose its object, and this with it) | 1 |
| Refused: a method read as a value (convertToArrayAssignmentElement would lose its object, and this with it) | 1 |
| Refused: a method read as a value (convertToObjectAssignmentElement would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createDirectory would lose its object, and this with it) | 2 |
| Refused: a method read as a value (createIntersectionTypeNode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocClassTag would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocDeprecatedTag would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocLink would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocLinkCode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocLinkPlain would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocOverrideTag would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocPrivateTag would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocProtectedTag would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocPublicTag would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createJSDocReadonlyTag would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createUnionTypeNode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (deleteFile would lose its object, and this with it) | 1 |
| Refused: a method read as a value (directoryExists would lose its object, and this with it) | 3 |
| Refused: a method read as a value (emit would lose its object, and this with it) | 1 |
| Refused: a method read as a value (emitBuildInfo would lose its object, and this with it) | 1 |
| Refused: a method read as a value (emitNextAffectedFile would lose its object, and this with it) | 1 |
| Refused: a method read as a value (enableCPUProfiler would lose its object, and this with it) | 2 |
| Refused: a method read as a value (fileExists would lose its object, and this with it) | 4 |
| Refused: a method read as a value (fill would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getAllDependencies would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getBuildInfo would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getCanonicalFileName would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getCurrentDirectory would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getDeclarationDiagnostics would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getDirectories would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getEnvironmentVariable would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getGlobalTypingsCacheLocation would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getModifiedTime would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getNearestAncestorDirectoryWithPackageJson would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getOrCreateCacheForModuleName would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getParsedCommandLine would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getSemanticDiagnostics would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getSemanticDiagnosticsOfNextAffectedFile would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getSourceFile would lose its object, and this with it) | 4 |
| Refused: a method read as a value (globalCacheResolutionModuleName would lose its object, and this with it) | 1 |
| Refused: a method read as a value (hasChangedEmitSignature would lose its object, and this with it) | 1 |
| Refused: a method read as a value (hasOwnProperty would lose its object, and this with it) | 1 |
| Refused: a method read as a value (liftToBlock would lose its object, and this with it) | 5 |
| Refused: a method read as a value (log would lose its object, and this with it) | 1 |
| Refused: a method read as a value (nonEscapingWrite would lose its object, and this with it) | 1 |
| Refused: a method read as a value (now would lose its object, and this with it) | 3 |
| Refused: a method read as a value (onWatchStatusChange would lose its object, and this with it) | 3 |
| Refused: a method read as a value (parenthesizeBranchOfConditionalExpression would lose its object, and this with it) | 2 |
| Refused: a method read as a value (parenthesizeCheckTypeOfConditionalType would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeConciseBodyOfArrowFunction would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeConditionOfConditionalExpression would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeConstituentTypeOfIntersectionType would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeConstituentTypeOfUnionType would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeElementTypeOfTupleType would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeExpressionForDisallowedComma would lose its object, and this with it) | 12 |
| Refused: a method read as a value (parenthesizeExpressionOfComputedPropertyName would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeExpressionOfExportDefault would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeExpressionOfExpressionStatement would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeExpressionOfNew would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeExtendsTypeOfConditionalType would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeLeadingTypeArgument would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeLeftSideOfAccess would lose its object, and this with it) | 7 |
| Refused: a method read as a value (parenthesizeNonArrayTypeOfPostfixType would lose its object, and this with it) | 2 |
| Refused: a method read as a value (parenthesizeOperandOfPostfixUnary would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeOperandOfPrefixUnary would lose its object, and this with it) | 6 |
| Refused: a method read as a value (parenthesizeOperandOfReadonlyTypeOperator would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeOperandOfTypeOperator would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeTypeOfOptionalType would lose its object, and this with it) | 1 |
| Refused: a method read as a value (readFile would lose its object, and this with it) | 5 |
| Refused: a method read as a value (realpath would lose its object, and this with it) | 1 |
| Refused: a method read as a value (releaseProgram would lose its object, and this with it) | 1 |
| Refused: a method read as a value (remove would lose its object, and this with it) | 1 |
| Refused: a method read as a value (repeat would lose its object, and this with it) | 1 |
| Refused: a method read as a value (replace would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportCyclicStructureError would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportInaccessibleThisError would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportInaccessibleUniqueSymbolError would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportInferenceFallback would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportLikelyUnsafeImportRequiredError would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportNonSerializableProperty would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportNonlocalAugmentation would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportPrivateInBaseOfClassExpression would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportTruncationError would lose its object, and this with it) | 1 |
| Refused: a method read as a value (setModifiedTime would lose its object, and this with it) | 1 |
| Refused: a method read as a value (setPrototypeOf would lose its object, and this with it) | 2 |
| Refused: a method read as a value (setTimeout would lose its object, and this with it) | 2 |
| Refused: a method read as a value (toKey would lose its object, and this with it) | 2 |
| Refused: a method read as a value (toString would lose its object, and this with it) | 1 |
| Refused: a method read as a value (trace would lose its object, and this with it) | 5 |
| Refused: a method read as a value (trackSymbol would lose its object, and this with it) | 2 |
| Refused: a method read as a value (useCaseSensitiveFileNames would lose its object, and this with it) | 2 |
| Refused: a method read as a value (watchDirectory would lose its object, and this with it) | 2 |
| Refused: a method read as a value (watchFile would lose its object, and this with it) | 2 |
| Refused: a method read as a value (writeFile would lose its object, and this with it) | 2 |
| Refused: a method read as a value (writeOutputIsTTY would lose its object, and this with it) | 1 |
| Refused: a non-exhaustive enum switch; missing ModuleKind.CommonJS | 1 |
| Refused: a non-exhaustive enum switch; missing ModuleResolutionKind.Classic | 1 |
| Refused: a non-exhaustive enum switch; missing SyntaxKind.Unknown | 38 |
| Refused: a number outside the proven flag domain assigned to BuilderFileEmit.None; its domain is a closed union of non-negative int32 bit subsets | 3 |
| Refused: a number outside the proven flag domain assigned to CheckFlags.None; its domain is a closed union of non-negative int32 bit subsets | 2 |
| Refused: a number outside the proven flag domain assigned to Connection.None; its domain is a closed union of non-negative int32 bit subsets | 3 |
| Refused: a number outside the proven flag domain assigned to Connection; its domain is a closed union of non-negative int32 bit subsets | 3 |
| Refused: a number outside the proven flag domain assigned to EmitFlags.None; its domain is a closed union of non-negative int32 bit subsets | 3 |
| Refused: a number outside the proven flag domain assigned to EmitFlags.SingleLine; its domain is a closed union of non-negative int32 bit subsets | 1 |
| Refused: a number outside the proven flag domain assigned to EscapeSequenceScanningFlags.String; its domain is a closed union of non-negative int32 bit subsets | 2 |
| Refused: a number outside the proven flag domain assigned to Extensions.TypeScript; its domain is a closed union of non-negative int32 bit subsets | 7 |
| Refused: a number outside the proven flag domain assigned to Extensions; its domain is a closed union of non-negative int32 bit subsets | 1 |
| Refused: a number outside the proven flag domain assigned to GetLiteralTextFlags.None; its domain is a closed union of non-negative int32 bit subsets | 1 |
| Refused: a number outside the proven flag domain assigned to InternalEmitFlags.None; its domain is a closed union of non-negative int32 bit subsets | 3 |
| Refused: a number outside the proven flag domain assigned to LexicalEnvironmentFlags.None; its domain is a closed union of non-negative int32 bit subsets | 2 |
| Refused: a number outside the proven flag domain assigned to ListFormat.None; its domain is a closed union of non-negative int32 bit subsets | 1 |
| Refused: a number outside the proven flag domain assigned to NodeFlags.NestedNamespace; its domain is a closed union of non-negative int32 bit subsets | 1 |
| Refused: a number outside the proven flag domain assigned to NodeFlags.None; its domain is a closed union of non-negative int32 bit subsets | 8 |
| Refused: a number outside the proven flag domain assigned to NodeResolutionFeatures.None; its domain is a closed union of non-negative int32 bit subsets | 2 |
| Refused: a number outside the proven flag domain assigned to ObjectFlags.None; its domain is a closed union of non-negative int32 bit subsets | 3 |
| Refused: a number outside the proven flag domain assigned to TokenFlags.None; its domain is a closed union of non-negative int32 bit subsets | 3 |
| Refused: a type predicate | 589 |
| Refused: a value of type (baseDir: string, moduleName: string) => { module: any; modulePath: string; error: undefined; } \| { module: undefined; modulePath: undefined; error: unknown; } seen as (baseDir: string, moduleName: string) => ModuleImportResult, which can write string \| undefined where string is read | 1 |
| Refused: a value of type (left: MappedPosition, right: MappedPosition) => boolean seen as EqualityComparer<SourceMappedPosition> \| undefined, which can write string \| undefined where string is read | 1 |
| Refused: a value of type (node: CommentRange) => boolean seen as (value: SynthesizedComment) => boolean, which can write number where -1 is read | 1 |
| Refused: a value of type (sourceFile: SourceFile \| undefined, cancellationToken: CancellationToken \| undefined) => readonly DiagnosticWithLocation[] seen as (sourceFile?: SourceFile \| undefined, cancellationToken?: CancellationToken \| undefined) => readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 1 |
| Refused: a value of type (symbol: Symbol) => { visitedTypes: Type[]; visitedSymbols: Symbol[]; } seen as (root: Symbol) => { visitedTypes: readonly Type[]; visitedSymbols: readonly Symbol[]; }, which can write readonly Type[] where Type[] is read | 1 |
| Refused: a value of type (symbolAccessibilityResult: SymbolAccessibilityResult) => { diagnosticMessage: DiagnosticMessage; errorNode: DeclarationDiagnosticProducing; typeName: DeclarationName \| undefined; } \| undefined seen as (symbolAccessibilityResult: SymbolAccessibilityResult) => SymbolAccessibilityDiagnostic \| undefined, which can write Node where DeclarationDiagnosticProducing is read | 1 |
| Refused: a value of type (type: Type) => { visitedTypes: Type[]; visitedSymbols: Symbol[]; } seen as (root: Type) => { visitedTypes: readonly Type[]; visitedSymbols: readonly Symbol[]; }, which can write readonly Type[] where Type[] is read | 1 |
| Refused: a value of type AbstractKeyword \| AccessorKeyword \| AsyncKeyword \| ConstKeyword \| DeclareKeyword \| Decorator \| ... 9 more ... \| StaticKeyword seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type AccessorDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type AmbientModuleDeclaration \| undefined seen as (AmbientModuleDeclaration & { name: StringLiteral; }) \| undefined, whose readonly field name becomes writable: a readonly field may hold something narrower than ModuleName, which a write of StringLiteral would replace | 1 |
| Refused: a value of type ArrayBindingPattern seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ArrayLiteralExpression \| AssignmentExpression<EqualsToken> \| BindingElement \| ElementAccessExpression \| ... 8 more ... \| VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type BigIntLiteral \| Identifier \| NoSubstitutionTemplateLiteral \| NumericLiteral \| PrivateIdentifier \| StringLiteral seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type BigIntLiteral \| NoSubstitutionTemplateLiteral \| NumericLiteral \| StringLiteral seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type BinaryExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 4 |
| Refused: a value of type BindingElement seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 5 |
| Refused: a value of type BindingOrAssignmentElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type Block seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type BlockLike seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type BuilderProgram seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what BuilderProgram can't hold | 1 |
| Refused: a value of type BuilderProgram \| Program seen as BuilderProgram, which can write SourceFile \| undefined where SourceFile is read | 1 |
| Refused: a value of type BuilderProgramState seen as ReusableBuilderProgramState, which can write Map<Path, readonly Diagnostic[] \| readonly ReusableDiagnostic[]> where Map<Path, readonly Diagnostic[]> is read | 9 |
| Refused: a value of type BuilderProgramStateWithDefinedProgram seen as BuilderProgramState, which can write Program \| undefined where Program is read | 10 |
| Refused: a value of type CallExpression seen as RequireOrImportCall, whose readonly field expression becomes writable: a readonly field may hold something narrower than LeftHandSideExpression, which a write of LeftHandSideExpression & Identifier would replace | 1 |
| Refused: a value of type CallExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Children seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ClassDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ClassDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ClassElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ClassStaticBlockDeclaration seen as Mutable<ClassStaticBlockDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type CommandLineOption seen as CommandLineOptionOfCustomType, which can write Map<string, string \| number> where "boolean" is read | 1 |
| Refused: a value of type CommandLineOption seen as CommandLineOptionOfListType, which can write "list" \| "listOrElement" where "boolean" is read | 1 |
| Refused: a value of type CommandLineOption \| undefined seen as TsConfigOnlyOption, which can write "object" where "boolean" is read | 1 |
| Refused: a value of type CommandLineOptionOfBooleanType \| CommandLineOptionOfCustomType \| CommandLineOptionOfNumberType \| CommandLineOptionOfStringType \| TsConfigOnlyOption seen as CommandLineOptionOfCustomType, which can write Map<string, string \| number> where "boolean" is read | 1 |
| Refused: a value of type CompilerHost seen as CompilerHostLikeForCache, which can write WriteFileCallback \| undefined where WriteFileCallback is read | 1 |
| Refused: a value of type CompilerOptions & { types: string[]; } seen as CompilerOptions, which can write string[] \| undefined where string[] is read | 1 |
| Refused: a value of type CompilerOptionsValue seen as TsConfigSourceFile \| CompilerOptionsValue, which can write string \| number where string is read | 4 |
| Refused: a value of type CompilerOptionsValue seen as string[], which can write string where PluginImport is read | 1 |
| Refused: a value of type ComputedPropertyName seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ConstructorDeclaration seen as Mutable<ConstructorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Declaration seen as T, a type parameter whose constraint Declaration can be written, so it can write what Declaration can't hold | 2 |
| Refused: a value of type DestructuringAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Diagnostic seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 2 |
| Refused: a value of type DiagnosticMessageChain seen as T, a type parameter whose constraint DiagnosticMessageChain \| ReusableDiagnosticMessageChain can be written, so it can write what DiagnosticMessageChain can't hold | 3 |
| Refused: a value of type DiagnosticWithDetachedLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where undefined is read | 4 |
| Refused: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 15 |
| Refused: a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where SourceFile is read | 13 |
| Refused: a value of type DiagnosticWithLocation \| undefined seen as Diagnostic \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 |
| Refused: a value of type DiagnosticWithLocation[] seen as Diagnostic[], which can write Diagnostic where DiagnosticWithLocation is read | 3 |
| Refused: a value of type DiagnosticWithLocation[] \| undefined seen as DiagnosticRelatedInformation[] \| undefined, which can write DiagnosticRelatedInformation where DiagnosticWithLocation is read | 2 |
| Refused: a value of type ElementAccessExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type EvaluatorResult<number> seen as EvaluatorResult<string \| number \| undefined>, which can write string \| number \| undefined where number is read | 16 |
| Refused: a value of type EvaluatorResult<string \| undefined> seen as EvaluatorResult<string \| number \| undefined>, which can write string \| number \| undefined where string \| undefined is read | 1 |
| Refused: a value of type EvaluatorResult<string> seen as EvaluatorResult<string \| number \| undefined>, which can write string \| number \| undefined where string is read | 2 |
| Refused: a value of type EvaluatorResult<string> seen as EvaluatorResult<string \| undefined>, which can write string \| undefined where string is read | 1 |
| Refused: a value of type EvaluatorResult<undefined> seen as EvaluatorResult<string \| number \| undefined>, which can write string \| number \| undefined where undefined is read | 1 |
| Refused: a value of type EvaluatorResult<undefined> seen as EvaluatorResult<string \| undefined>, which can write string \| undefined where undefined is read | 1 |
| Refused: a value of type Expression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 12 |
| Refused: a value of type Expression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type Expression \| GeneratedIdentifier seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type Expression \| Identifier seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ExpressionWithTypeArguments seen as ExpressionWithTypeArguments & { expression: Identifier \| PropertyAccessEntityNameExpression; }, whose readonly field expression becomes writable: a readonly field may hold something narrower than LeftHandSideExpression, which a write of LeftHandSideExpression & (Identifier \| PropertyAccessEntityNameExpression) would replace | 1 |
| Refused: a value of type Extension[] seen as string[], which can write string where Extension is read | 1 |
| Refused: a value of type FlowArrayMutation \| FlowAssignment seen as FlowNode, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 5 |
| Refused: a value of type FlowArrayMutation \| FlowAssignment \| FlowCall \| FlowCondition \| FlowLabel \| FlowReduceLabel \| FlowStart \| FlowUnreachable seen as FlowNode, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 1 |
| Refused: a value of type FlowCall seen as FlowNode, which can write BinaryExpression \| CallExpression where CallExpression is read | 2 |
| Refused: a value of type FlowNode seen as FlowArrayMutation \| FlowAssignment, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 1 |
| Refused: a value of type FlowNode seen as FlowCondition, which can write Expression where BinaryExpression \| CallExpression is read | 1 |
| Refused: a value of type FlowNode seen as FlowLabel, which can write undefined where BinaryExpression \| CallExpression is read | 3 |
| Refused: a value of type FlowNode seen as FlowNode \| undefined, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 14 |
| Refused: a value of type FlowNode seen as FlowNode[] \| FlowNode \| undefined, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 5 |
| Refused: a value of type FlowNode seen as FlowReduceLabel, which can write FlowReduceLabelData where BinaryExpression \| CallExpression is read | 1 |
| Refused: a value of type FlowNode seen as FlowSwitchClause, which can write FlowSwitchClauseData where BinaryExpression \| CallExpression is read | 1 |
| Refused: a value of type FunctionDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type FutureSourceFile \| SourceFile seen as Pick<SourceFile, "fileName" \| "impliedNodeFormat" \| "packageJsonScope">, whose readonly field fileName becomes writable: a readonly field may hold something narrower than string, which a write of string would replace | 2 |
| Refused: a value of type FutureSourceFile \| SourceFile seen as Pick<SourceFile, "fileName" \| "impliedNodeFormat">, whose readonly field fileName becomes writable: a readonly field may hold something narrower than string, which a write of string would replace | 3 |
| Refused: a value of type GeneratedIdentifier seen as GeneratedIdentifier \| Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as GeneratedIdentifier \| GeneratedPrivateIdentifier \| Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as Identifier \| PrivateIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as Node \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 4 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier \| Identifier \| PrivateIdentifier seen as Identifier \| PrivateIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier \| Node seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 5 |
| Refused: a value of type GeneratedIdentifier \| Identifier seen as Identifier \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GetAccessorDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type GetAccessorDeclaration \| SetAccessorDeclaration seen as Mutable<AccessorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Identifier seen as Mutable<Identifier>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Identifier seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Identifier seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type Identifier seen as T, a type parameter whose constraint Node can be written, so it can write what Identifier can't hold | 1 |
| Refused: a value of type Identifier seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type Identifier seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type Identifier \| JsxNamespacedName \| LiteralExpression \| PrivateIdentifier seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Identifier[][] seen as ModuleExportName[][], which can write ModuleExportName[] where Identifier[] is read | 4 |
| Refused: a value of type ImportTypeNode \| undefined seen as ValidImportTypeNode \| undefined, whose readonly field argument becomes writable: a readonly field may hold something narrower than TypeNode, which a write of LiteralTypeNode & { literal: StringLiteral; } would replace | 1 |
| Refused: a value of type JSDocNullableType seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JsxExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type JsxOpeningFragment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JsxTagNameExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type LeftHandSideExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 4 |
| Refused: a value of type LeftHandSideExpression \| UnaryExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type LiteralLikeNode seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Map<CharacterCodes, RegularExpressionFlags> seen as Map<CharacterCodes, number>, which can write number where RegularExpressionFlags is read | 1 |
| Refused: a value of type Map<Path, Diagnostic[]> seen as Map<Path, readonly Diagnostic[]> \| undefined, which can write readonly Diagnostic[] where Diagnostic[] is read | 1 |
| Refused: a value of type Map<Path, FileWatcher> seen as Map<string, FileWatcher>, which can write string where Path is read | 1 |
| Refused: a value of type Map<Path, string[]> seen as InvokeMap, which can write true \| string[] where string[] is read | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, BuildInfoCacheEntry> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where BuildInfoCacheEntry is read | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, ConfigFileCacheEntry> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where ConfigFileCacheEntry is read | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, Map<Path, Date>> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where Map<Path, Date> is read | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, ProgramUpdateLevel> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where ProgramUpdateLevel is read | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, Set<string> \| undefined> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where Set<string> \| undefined is read | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, T> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where T is read | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, UpToDateStatus> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where UpToDateStatus is read | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, readonly Diagnostic[]> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where readonly Diagnostic[] is read | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, true> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where true is read | 1 |
| Refused: a value of type Map<string, ImportsNotUsedAsValues> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ImportsNotUsedAsValues is read | 1 |
| Refused: a value of type Map<string, JsxEmit> seen as Map<string, string \| number>, which can write string \| number where JsxEmit is read | 1 |
| Refused: a value of type Map<string, Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>> seen as ScriptTargetFeatures, which can write string where never is read | 1 |
| Refused: a value of type Map<string, ModuleDetectionKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ModuleDetectionKind is read | 1 |
| Refused: a value of type Map<string, ModuleResolutionKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ModuleResolutionKind is read | 1 |
| Refused: a value of type Map<string, NewLineKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where NewLineKind is read | 1 |
| Refused: a value of type Map<string, PollingWatchKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where PollingWatchKind is read | 1 |
| Refused: a value of type Map<string, SyntaxKind> seen as Map<string, number>, which can write number where SyntaxKind is read | 1 |
| Refused: a value of type Map<string, WatchDirectoryKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where WatchDirectoryKind is read | 1 |
| Refused: a value of type Map<string, WatchFileKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where WatchFileKind is read | 1 |
| Refused: a value of type Map<string, never[]> seen as Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>, which can write string[] where never[] is read | 10 |
| Refused: a value of type Map<string, string> seen as Map<string, string \| number>, which can write string \| number where string is read | 1 |
| Refused: a value of type Map<string, string[] \| never[]> seen as Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>, which can write string where never is read | 3 |
| Refused: a value of type MethodDeclaration seen as Mutable<MethodDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type MethodDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type MethodDeclaration \| PropertyAssignment \| AccessorDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type MethodDeclaration \| PropertyDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type MissingDeclaration seen as Mutable<MissingDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ModifierLike seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ModuleDeclaration seen as Mutable<ModuleDeclaration \| SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ModuleResolutionState & { host: GetPackageJsonEntrypointsHost; } seen as ModuleResolutionState, which can write ModuleResolutionHost where ModuleResolutionHost & GetPackageJsonEntrypointsHost is read | 1 |
| Refused: a value of type NamespaceExportDeclaration seen as Mutable<NamespaceExportDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Node seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 5 |
| Refused: a value of type Node seen as Node \| SourceMapRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Node seen as Node \| TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Node seen as SyntaxList, whose readonly field kind becomes writable: a readonly field may hold something narrower than SyntaxKind, which a write of SyntaxKind.SyntaxList would replace | 1 |
| Refused: a value of type Node seen as T, a type parameter whose constraint Node can be written, so it can write what Node can't hold | 4 |
| Refused: a value of type Node seen as T, a type parameter whose constraint Node \| undefined can be written, so it can write what Node can't hold | 1 |
| Refused: a value of type Node seen as TemplateLiteralTypeNode, whose readonly field kind becomes writable: a readonly field may hold something narrower than SyntaxKind, which a write of SyntaxKind.TemplateLiteralType would replace | 1 |
| Refused: a value of type Node seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 29 |
| Refused: a value of type Node \| NodeArray<Node> seen as NodeArray<Node>, whose readonly field transformFlags becomes writable: a readonly field may hold something narrower than TransformFlags, which a write of TransformFlags would replace | 1 |
| Refused: a value of type Node \| SourceMapRange seen as SourceMapRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Node \| TextRange seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Node \| undefined seen as EmitNode \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of EmitFlags would replace | 1 |
| Refused: a value of type NodeArray<ClassElement> seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 4 |
| Refused: a value of type NodeArray<Expression> seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type NodeArray<Statement> seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type NodeArray<Statement> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type NodeArray<T> seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type NonNullExpression seen as Mutable<NonNullExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type NonNullable<T> seen as T, a type parameter whose constraint any can be written, so it can write what NonNullable<T> can't hold | 6 |
| Refused: a value of type ObjectBindingOrAssignmentPattern seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ObjectBindingPattern seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type OptionalChain seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type OptionalTypeNode seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ParameterDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ParameterDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ParameterDeclaration \| PropertyDeclaration \| PropertySignature \| SignatureDeclaration seen as Mutable<ParameterDeclaration \| PropertyDeclaration \| PropertySignature \| SignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ParameterDeclaration \| PropertyDeclaration \| PropertySignature \| SignatureDeclaration seen as Mutable<ParameterDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ParameterDeclaration \| VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ParameterDeclaration \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type ParenthesizedExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ParenthesizedExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ParseConfigHost seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where boolean is read | 1 |
| Refused: a value of type Path[] seen as string[], which can write string where Path is read | 1 |
| Refused: a value of type PostfixUnaryExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type PostfixUnaryExpression \| PrefixUnaryExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type PrivateIdentifier seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type PropertyAccessExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type PropertyAssignment seen as Mutable<PropertyAssignment \| ShorthandPropertyAssignment>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type PropertyAssignment seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type PropertyName seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type PropertySignature seen as Mutable<PropertySignature>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type QualifiedName seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Readonly<BuilderState> \| undefined seen as BuilderState \| undefined, whose readonly field fileInfos becomes writable: a readonly field may hold something narrower than Map<Path, FileInfo>, which a write of Map<Path, FileInfo> would replace | 1 |
| Refused: a value of type ReferencedFile & { kind: FileIncludeKind.LibReferenceDirective; } seen as ReferencedFile, which can write ReferencedFileKind where FileIncludeKind.LibReferenceDirective is read | 1 |
| Refused: a value of type ReportFileInError[] seen as (ReportFileInError \| undefined)[], which can write ReportFileInError \| undefined where ReportFileInError is read | 1 |
| Refused: a value of type ResolvedModuleFull \| undefined seen as { path: string; originalPath: string \| true; extension: string; packageId: PackageId \| undefined; resolvedUsingTsExtension: boolean \| undefined; } \| undefined, whose readonly field originalPath becomes writable: a readonly field may hold something narrower than string \| undefined, which a write of string \| true would replace | 1 |
| Refused: a value of type SearchResult<Resolved> seen as { value: { resolved: Resolved; isExternalLibraryImport: true; } \| undefined; } \| undefined, which can write { resolved: Resolved; isExternalLibraryImport: true; } \| undefined where Resolved \| undefined is read | 1 |
| Refused: a value of type SearchResult<undefined> seen as SearchResult<Resolved>, which can write Resolved \| undefined where undefined is read | 6 |
| Refused: a value of type SetAccessorDeclaration seen as Mutable<SetAccessorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type SetAccessorDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ShorthandPropertyAssignment seen as Mutable<PropertyAssignment \| ShorthandPropertyAssignment>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ShorthandPropertyAssignment seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type SourceFile seen as EmitNode \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of EmitFlags would replace | 2 |
| Refused: a value of type SourceFile seen as Mutable<ModuleDeclaration \| SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type SourceFile seen as Mutable<SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type SourceFile seen as SourceFile \| SourceFileLike, which can write readonly number[] \| undefined where readonly number[] is read | 3 |
| Refused: a value of type SourceFile seen as SourceFileLike, which can write readonly number[] \| undefined where readonly number[] is read | 27 |
| Refused: a value of type SourceFile \| undefined seen as FileReasonToChainCache \| undefined, which can write DiagnosticMessageChain[] \| undefined where RedirectInfo \| undefined is read | 1 |
| Refused: a value of type SourceMapSource seen as SourceFileLike, which can write readonly number[] \| undefined where readonly number[] is read | 1 |
| Refused: a value of type StringLiteral seen as Mutable<StringLiteral>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type StringLiteral seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Symbol \| undefined seen as Declaration \| undefined, which can write number \| undefined where number is read | 2 |
| Refused: a value of type Symbol[] seen as unknown[], which can write unknown where Symbol is read | 2 |
| Refused: a value of type System seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where boolean is read | 1 |
| Refused: a value of type T seen as Mutable<T>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of T["pos"] would replace | 6 |
| Refused: a value of type T seen as T, a type parameter whose constraint Node can be written, so it can write what T can't hold | 1 |
| Refused: a value of type T seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type T[] seen as (T \| undefined)[], which can write T \| undefined where T is read | 1 |
| Refused: a value of type T[][] seen as (readonly T[])[], which can write readonly T[] where T[] is read | 1 |
| Refused: a value of type TaggedTemplateExpression seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type TemplateLiteralLikeNode seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Token<SyntaxKind> seen as T, a type parameter whose constraint Node can be written, so it can write what Token<SyntaxKind> can't hold | 2 |
| Refused: a value of type TsConfigSourceFile \| CompilerOptionsValue seen as CompilerOptionsValue, which can write string \| number where string is read | 2 |
| Refused: a value of type TsConfigSourceFile \| CompilerOptionsValue seen as string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined, which can write string \| number where string is read | 1 |
| Refused: a value of type TsConfigSourceFile \| CompilerOptionsValue seen as string[], which can write string where PluginImport is read | 1 |
| Refused: a value of type TypeNode seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Type[] seen as unknown[], which can write unknown where Type is read | 2 |
| Refused: a value of type VariableDeclaration \| DestructuringAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type VariableDeclarationList seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type VisitEachChildTable seen as Record<SyntaxKind, VisitEachChildFunction<any> \| undefined>, which can write VisitEachChildFunction<any> \| undefined where VisitEachChildFunction<QualifiedName> is read | 1 |
| Refused: a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read | 1 |
| Refused: a value of type any seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what any can't hold | 2 |
| Refused: a value of type any seen as T, a type parameter whose constraint Node can be written, so it can write what any can't hold | 2 |
| Refused: a value of type never seen as T, a type parameter whose constraint any can be written, so it can write what never can't hold | 1 |
| Refused: a value of type never seen as U, a type parameter whose constraint {} can be written, so it can write what never can't hold | 1 |
| Refused: a value of type never[] seen as (JSDoc \| JSDocTag)[], which can write JSDoc \| JSDocTag where never is read | 1 |
| Refused: a value of type never[] seen as (JSDocCallbackTag \| JSDocEnumTag \| JSDocTypedefTag)[], which can write JSDocCallbackTag \| JSDocEnumTag \| JSDocTypedefTag where never is read | 1 |
| Refused: a value of type never[] seen as (SourceMapRange \| undefined)[], which can write SourceMapRange \| undefined where never is read | 1 |
| Refused: a value of type never[] seen as ChildDirectoryWatcher[], which can write ChildDirectoryWatcher where never is read | 2 |
| Refused: a value of type never[] seen as CommentRange[], which can write CommentRange where never is read | 1 |
| Refused: a value of type never[] seen as Comparator[][], which can write Comparator[] where never is read | 1 |
| Refused: a value of type never[] seen as Declaration[], which can write Declaration where never is read | 1 |
| Refused: a value of type never[] seen as DiagnosticArguments, which can write string \| number \| boolean \| readonly string[] \| SourceFile \| undefined where never is read | 3 |
| Refused: a value of type never[] seen as Diagnostic[], which can write Diagnostic where never is read | 5 |
| Refused: a value of type never[] seen as FlowNode[], which can write FlowNode where never is read | 1 |
| Refused: a value of type never[] seen as JSDocImportTag[], which can write JSDocImportTag where never is read | 1 |
| Refused: a value of type never[] seen as ProjectReference[], which can write ProjectReference where never is read | 1 |
| Refused: a value of type never[] seen as RequireOrImportCall[], which can write RequireOrImportCall where never is read | 1 |
| Refused: a value of type never[] seen as ResolvedConfigFileName[], which can write ResolvedConfigFileName where never is read | 3 |
| Refused: a value of type never[] seen as ResolvedProjectReference[], which can write ResolvedProjectReference where never is read | 1 |
| Refused: a value of type never[] seen as SourceFile[], which can write SourceFile where never is read | 2 |
| Refused: a value of type never[] seen as SourceMappedPosition[], which can write SourceMappedPosition where never is read | 1 |
| Refused: a value of type never[] seen as StringLiteralLike[], which can write StringLiteralLike where never is read | 1 |
| Refused: a value of type never[] seen as TransformerFactory<Bundle \| SourceFile>[], which can write TransformerFactory<Bundle \| SourceFile> where never is read | 1 |
| Refused: a value of type never[] seen as string[] \| never[], which can write string where never is read | 3 |
| Refused: a value of type never[] seen as string[], which can write string where never is read | 8 |
| Refused: a value of type never[] \| SortedArray<DiagnosticWithLocation> seen as Diagnostic[], which can write Diagnostic where never is read | 1 |
| Refused: a value of type never[][] seen as string[][], which can write string where never is read | 1 |
| Refused: a value of type number[] seen as string \| (string \| number)[] \| undefined, which can write string \| number where number is read | 1 |
| Refused: a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[] \| undefined, which can write SourceFile \| undefined where SourceFile is read | 2 |
| Refused: a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 3 |
| Refused: a value of type readonly Extension[][] seen as readonly string[][], which can write string where Extension is read | 3 |
| Refused: a value of type readonly T[] \| undefined seen as unknown[], which can write unknown where T is read | 1 |
| Refused: a value of type readonly string[] \| undefined seen as RegExp[] \| undefined, which can write RegExp where string is read | 1 |
| Refused: a value of type string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined seen as string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined, which can write string \| number where string is read | 1 |
| Refused: a value of type string \| number \| boolean \| string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> seen as CompilerOptionsValue, which can write string \| number where string is read | 1 |
| Refused: a value of type string[] seen as CompilerOptionsValue, which can write string \| number where string is read | 4 |
| Refused: a value of type string[] seen as string \| (string \| number)[] \| undefined, which can write string \| number where string is read | 1 |
| Refused: a value of type string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] seen as unknown[], which can write unknown where string is read | 2 |
| Refused: a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"CheckMode", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof CheckMode is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"EmitFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof EmitFlags is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"FlowFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof FlowFlags is read | 2 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"GeneratedIdentifierFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof GeneratedIdentifierFlags is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"ModifierFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ModifierFlags is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"NodeCheckFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof NodeCheckFlags is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"NodeFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof NodeFlags is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"ObjectFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ObjectFlags is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"RelationComparisonResult", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof RelationComparisonResult is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"ScriptKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ScriptKind is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"SignatureCheckMode", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SignatureCheckMode is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"SignatureFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SignatureFlags is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"SnippetKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SnippetKind is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"SymbolFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SymbolFlags is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"SyntaxKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SyntaxKind is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"TransformFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TransformFlags is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"TypeFacts", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TypeFacts is read | 1 |
| Refused: a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"TypeFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TypeFlags is read | 1 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what undefined can't hold | 1 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint Declaration can be written, so it can write what undefined can't hold | 1 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint Node can be written, so it can write what undefined can't hold | 6 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint WatchedFileWithIsClosed can be written, so it can write what undefined can't hold | 3 |
| Refused: a value of type { (fileName: string): DiagnosticWithLocation[]; (): Diagnostic[]; } seen as { (): Diagnostic[]; (fileName: string): DiagnosticWithLocation[]; }, which can write Diagnostic where DiagnosticWithLocation is read | 1 |
| Refused: a value of type { affectedFile: SourceFile; emitKind: BuilderFileEmit.Js \| BuilderFileEmit.JsMap \| BuilderFileEmit.JsInlineMap \| BuilderFileEmit.DtsErrors \| ... 6 more ... \| BuilderFileEmit.All; } seen as { affectedFile: Program \| SourceFile \| undefined; emitKind: BuilderFileEmit; }, which can write Program \| SourceFile \| undefined where SourceFile is read | 1 |
| Refused: a value of type { all?: boolean; allowImportingTsExtensions?: boolean; allowJs?: boolean; allowNonTsExtensions?: boolean; allowArbitraryExtensions?: boolean; allowSyntheticDefaultImports?: boolean; allowUmdGlobalAccess?: boolean; ... 129 more ...; moduleResolution: ModuleResolutionKind; } seen as CompilerOptions, which can write ModuleResolutionKind \| undefined where ModuleResolutionKind is read | 1 |
| Refused: a value of type { arguments: { name?: string; } & { path: string; }; range: CommentRange; } \| { arguments: { name: string; }; range: CommentRange; } \| { arguments: { factory: string; }; range: CommentRange; } \| ... 5 more ... \| ... seen as ({ arguments: { name?: string; } & { path: string; }; range: CommentRange; } \| { arguments: { name: string; }; range: CommentRange; } \| { arguments: { factory: string; }; range: CommentRange; } \| ... 5 more ... \| ...)[] \| ... 8 more ... \| ..., which can write { name?: string; } & { path: string; } where never is read | 1 |
| Refused: a value of type { compilerOptions: CompilerOptions; traceEnabled: boolean; affectingLocations: string[] \| undefined; resultFromCache?: ResolvedModuleWithFailedLookupLocations; ... 9 more ...; host: GetPackageJsonEntrypointsHost; } seen as ModuleResolutionState & { host: GetPackageJsonEntrypointsHost; }, which can write string[] \| undefined where never[] is read | 1 |
| Refused: a value of type { ending: ModuleSpecifierEnding; value: string; }[] seen as { ending: ModuleSpecifierEnding \| undefined; value: string; }[], which can write ModuleSpecifierEnding \| undefined where ModuleSpecifierEnding is read | 1 |
| Refused: a value of type { host: ModuleResolutionHost; traceEnabled: boolean; failedLookupLocations: string[] \| undefined; affectingLocations: string[] \| undefined; resultFromCache?: ResolvedModuleWithFailedLookupLocations; ... 8 more ...; reportDiagnostic: ...; } seen as ModuleResolutionState, which can write CompilerOptions where { all?: boolean; allowImportingTsExtensions?: boolean; allowJs?: boolean; allowNonTsExtensions?: boolean; allowArbitraryExtensions?: boolean; allowSyntheticDefaultImports?: boolean; allowUmdGlobalAccess?: boolean; ... 129 more ...; moduleResolution: ModuleResolutionKind; } is read | 1 |
| Refused: a value of type { id: number; flowNode: FlowNode; edges: never[]; text: string; lane: number; endLane: number; level: number; circular: false; } seen as FlowGraphNode, which can write FlowGraphEdge[] where never[] is read | 1 |
| Refused: a value of type { kind: "ambient" \| "node_modules" \| "paths" \| "redirect" \| "relative" \| undefined; moduleSpecifiers: readonly string[]; computedWithoutCache: false; } \| undefined seen as ModuleSpecifierResult \| undefined, which can write boolean where false is read | 1 |
| Refused: a value of type { major: number; minor: number; patch: number; prerelease: string; build: string; } seen as { major: string \| number; minor: number; patch: number; prerelease: string \| readonly string[]; build: string \| readonly string[]; }, which can write string \| number where number is read | 1 |
| Refused: a value of type { matchableStringSet: Set<string> \| undefined; patterns: Pattern[] \| undefined; } seen as ParsedPatterns, which can write ReadonlySet<string> \| undefined where Set<string> \| undefined is read | 1 |
| Refused: a value of type { path: string; originalPath: string \| true; extension: string; packageId: PackageId \| undefined; resolvedUsingTsExtension: boolean \| undefined; } \| undefined seen as Resolved \| undefined, which can write string \| true \| undefined where string \| true is read | 1 |
| Refused: a value of type { readonly args: readonly [{ readonly name: "types"; readonly optional: true; readonly captureSpan: true; }, { readonly name: "lib"; readonly optional: true; readonly captureSpan: true; }, { readonly name: "path"; readonly optional: true; readonly captureSpan: true; }, ..., ..., ...]; readonly kind: PragmaKindFlags.... seen as PragmaDefinition<string, string, string, string>, whose readonly field args becomes writable: a readonly field may hold something narrower than readonly [{ readonly name: "types"; readonly optional: true; readonly captureSpan: true; }, { readonly name: "lib"; readonly optional: true; readonly captureSpan: true; }, { readonly name: "path"; readonly optional: true; readonly captureSpan: true; }, ..., ..., ...], which a write of readonly [PragmaArgumentSpecification<string>] \| readonly [PragmaArgumentSpecification<string>, PragmaArgumentSpecification<string>] \| ... \| ... \| undefined would replace | 2 |
| Refused: a value of type { resolved: Resolved; isExternalLibraryImport: true; } \| undefined seen as { resolved: Resolved; isExternalLibraryImport: boolean; } \| undefined, which can write boolean where true is read | 1 |
| Refused: a value of type { value: never; done: true; } seen as IteratorResult<Mapping, any>, which can write any where never is read | 14 |
| Refused: a value of type { value: { resolved: Resolved; isExternalLibraryImport: false; }; } \| undefined seen as SearchResult<{ resolved: Resolved; isExternalLibraryImport: boolean; }>, which can write { resolved: Resolved; isExternalLibraryImport: boolean; } \| undefined where { resolved: Resolved; isExternalLibraryImport: false; } is read | 2 |
| Refused: a value of type { value: { resolved: Resolved; isExternalLibraryImport: true; } \| undefined; } \| undefined seen as SearchResult<{ resolved: Resolved; isExternalLibraryImport: boolean; }>, which can write { resolved: Resolved; isExternalLibraryImport: boolean; } \| undefined where { resolved: Resolved; isExternalLibraryImport: true; } \| undefined is read | 1 |
| Refused: a value without nominal ancestry seen as BinaryExpressionStateMachine<TOuterState, TState, TResult> | 12 |
| Refused: an arbitrary number or a value from another enum assigned to AccessKind.Read; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to CharacterCodes.EOF; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to CharacterCodes.plus; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to CharacterCodes.slash; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to EmitOnly.Dts; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to Extension.Mjs; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to Extension.Ts; its members are a closed union | 3 |
| Refused: an arbitrary number or a value from another enum assigned to ForegroundColorEscapeSequences.Grey; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to GeneratedIdentifierFlags.None; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to InternalSymbolName.Call; its members are a closed union | 3 |
| Refused: an arbitrary number or a value from another enum assigned to InvalidPosition; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to ModifierFlags.None; its members are a closed union | 15 |
| Refused: an arbitrary number or a value from another enum assigned to ModuleKind.ESNext; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to OperatorPrecedence.Invalid; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to OuterExpressionKinds.Parentheses; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to ParsingContext.SourceElements; its members are a closed union | 4 |
| Refused: an arbitrary number or a value from another enum assigned to PipelinePhase.Notification; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SymbolFlags.None; its members are a closed union | 31 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ArrayBindingPattern; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ArrayLiteralExpression; its members are a closed union | 9 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ArrayType; its members are a closed union | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ArrowFunction; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.AsExpression; its members are a closed union | 4 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.AwaitExpression; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.BinaryExpression; its members are a closed union | 56 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.BindingElement; its members are a closed union | 12 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.Block; its members are a closed union | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.BreakStatement; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.Bundle; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.CallExpression; its members are a closed union | 26 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.CallSignature; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.CaseBlock; its members are a closed union | 4 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.CaseClause; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.CatchClause; its members are a closed union | 5 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ClassDeclaration; its members are a closed union | 6 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ClassExpression; its members are a closed union | 7 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ClassStaticBlockDeclaration; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.CommaListExpression; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.CommaToken; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ComputedPropertyName; its members are a closed union | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ConditionalExpression; its members are a closed union | 9 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ConditionalType; its members are a closed union | 5 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ConstructSignature; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.Constructor; its members are a closed union | 5 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ConstructorType; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ContinueStatement; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.DebuggerStatement; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.Decorator; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.DefaultClause; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.DeleteExpression; its members are a closed union | 4 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.DoStatement; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ElementAccessExpression; its members are a closed union | 12 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.EnumDeclaration; its members are a closed union | 6 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.EnumMember; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ExportAssignment; its members are a closed union | 8 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ExportDeclaration; its members are a closed union | 15 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ExportSpecifier; its members are a closed union | 7 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ExpressionStatement; its members are a closed union | 4 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ExpressionWithTypeArguments; its members are a closed union | 13 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ExternalModuleReference; its members are a closed union | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ForInStatement; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ForOfStatement; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ForStatement; its members are a closed union | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.FunctionDeclaration; its members are a closed union | 8 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.FunctionExpression; its members are a closed union | 6 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.FunctionType; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.GetAccessor; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.HeritageClause; its members are a closed union | 4 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.Identifier; its members are a closed union | 27 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.IfStatement; its members are a closed union | 4 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ImportAttribute; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ImportAttributes; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ImportClause; its members are a closed union | 6 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ImportDeclaration; its members are a closed union | 5 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ImportEqualsDeclaration; its members are a closed union | 6 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ImportSpecifier; its members are a closed union | 7 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ImportType; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.IndexSignature; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.IndexedAccessType; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.InferType; its members are a closed union | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.InterfaceDeclaration; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.IntersectionType; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDoc; its members are a closed union | 5 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocCallbackTag; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocClassTag; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocEnumTag; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocFunctionType; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocImportTag; its members are a closed union | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocNameReference; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocNonNullableType; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocNullableType; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocOptionalType; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocOverloadTag; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocParameterTag; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocReturnTag; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocSatisfiesTag; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocSeeTag; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocSignature; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocTemplateTag; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocTypeExpression; its members are a closed union | 8 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocTypeLiteral; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocTypedefTag; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JsxAttribute; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JsxAttributes; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JsxClosingElement; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JsxElement; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JsxExpression; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JsxFragment; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JsxNamespacedName; its members are a closed union | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JsxOpeningElement; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JsxSelfClosingElement; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JsxSpreadAttribute; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JsxText; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.LabeledStatement; its members are a closed union | 5 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.LiteralType; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.MappedType; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.MetaProperty; its members are a closed union | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.MethodDeclaration; its members are a closed union | 5 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.MethodSignature; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.MinusToken; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ModuleBlock; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ModuleDeclaration; its members are a closed union | 7 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.NamedExports; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.NamedImports; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.NamedTupleMember; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.NamespaceExport; its members are a closed union | 5 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.NamespaceExportDeclaration; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.NamespaceImport; its members are a closed union | 4 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.NewExpression; its members are a closed union | 5 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.NoSubstitutionTemplateLiteral; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.NonNullExpression; its members are a closed union | 4 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.NumericLiteral; its members are a closed union | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ObjectBindingPattern; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ObjectLiteralExpression; its members are a closed union | 7 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.OptionalType; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.Parameter; its members are a closed union | 19 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ParenthesizedExpression; its members are a closed union | 8 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ParenthesizedType; its members are a closed union | 5 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.PartiallyEmittedExpression; its members are a closed union | 4 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.PostfixUnaryExpression; its members are a closed union | 6 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.PrefixUnaryExpression; its members are a closed union | 20 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.PrivateIdentifier; its members are a closed union | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.PropertyAccessExpression; its members are a closed union | 15 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.PropertyAssignment; its members are a closed union | 5 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.PropertyDeclaration; its members are a closed union | 9 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.PropertySignature; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.QualifiedName; its members are a closed union | 6 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ReturnStatement; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.SatisfiesExpression; its members are a closed union | 6 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.SetAccessor; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ShorthandPropertyAssignment; its members are a closed union | 5 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.SourceFile; its members are a closed union | 11 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.SpreadAssignment; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.SpreadElement; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.StringLiteral; its members are a closed union | 10 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.SwitchStatement; its members are a closed union | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TaggedTemplateExpression; its members are a closed union | 6 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TemplateExpression; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TemplateHead; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TemplateLiteralTypeSpan; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TemplateSpan; its members are a closed union | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ThrowStatement; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TryStatement; its members are a closed union | 5 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TupleType; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TypeAliasDeclaration; its members are a closed union | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TypeAssertionExpression; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TypeLiteral; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TypeOfExpression; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TypeOperator; its members are a closed union | 4 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TypeParameter; its members are a closed union | 5 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TypePredicate; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TypeQuery; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TypeReference; its members are a closed union | 6 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.UnionType; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.VariableDeclaration; its members are a closed union | 9 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.VariableDeclarationList; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.VariableStatement; its members are a closed union | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.VoidExpression; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.WhileStatement; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.WithStatement; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.YieldExpression; its members are a closed union | 3 |
| Refused: an arbitrary number or a value from another enum assigned to TempFlags.Auto; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to TransformFlags.None; its members are a closed union | 2 |
| Refused: an arbitrary number or a value from another enum assigned to TypeReferenceSerializationKind.Unknown; its members are a closed union | 1 |
| Refused: an import cycle | 1 |
| Refused: an index signature | 8 |
| Refused: arguments | 1 |
| Refused: arithmetic assigned back into an enum; the result need not be one of its members | 22 |
| Refused: debugger | 1 |
| Refused: delete | 1 |
| Refused: in | 2 |
| Refused: inherited library member compare read as an own field | 1 |
| Refused: inherited library member hasOwnProperty read as an own field | 1 |
| Refused: inherited library member prototype read as an own field | 1 |
| Refused: inherited library member replace read as an own field | 1 |
| Refused: instantiating a generic function makes a value of type T \| undefined written where T is read | 1 |
| Refused: the non-null assertion ! | 625 |
| Refused: this in a namespace function; a qualified call and a detached call have different receivers | 1 |
| Refused: var | 1 |
| Refused: yield (generators) | 5 |
| SkippedDependency: a dependency function whose body has checker diagnostics (measurement skipped) | 4 |

# Deltas

Source adaptations and compiler commits both changed; deltas do not isolate causal feature contributions.

- delta_vs_cumulative_2: checker -1042, checker-clean files +10, lowering {'NotYet': -39, 'Refused': 823, 'SkippedDependency': -1, 'error': -1, 'panic': 0}.
- delta_vs_cumulative_2_nested: checker -1042, checker-clean files +10, lowering {'NotYet': -37, 'Refused': 823, 'SkippedDependency': -1, 'error': -1, 'panic': 0}.

# Pinned provenance and reproduction

Integration area/stage3: `634ef061fc72c061e2de1606d5c8faebec4411f6`. Scratch head: `237d7872fa69f21139c6f19be5e52fa5bcca5758` on `scratch/latent-meter3`, never pushed. Feature pins:

- `codex/taste-not-soundness`: `d2c05df34443d9c0ae4fe6639beef2c5aee8a4ad`
- `codex/flag-enums`: `246ecc073993de6a5f1bcb64346cde7938deec43`
- `codex/namespaces-tsc`: `90f1baca0fce5285ddfa2f8801070dc9d7c9108c`
- `codex/fallthrough-and-implicit-returns`: `5f77d3315f3a6472ee2a6cc28e23e2f7cb1eb8ff`
- `codex/nested-functions`: `e7587d3ecc702d0b2d757b67b0f878a1bd7e7783`

TypeScript 6.0.3 `050880ce59e30b356b686bd3144efe24f875ebc8`. Adaptations: 10-type-imports, 20-optional-declarations, 30-indexed-reads, 31-indexed-reads-checker, 32-indexed-reads-program, 33-indexed-reads-emit, 40-explicit-any, 45-regex-captures, 46-fix-pragma-empty-argument. The adapter files and apply script exactly match the integration tip; their hash ledger and patch-set are archived.

Build/run durations (seconds): `{'overlay': 0.022, 'build': 0.066, 'run': 166.824}`. Toolchain setup: Go 0s, clang 0s, Node 0s, submodules 0s, warm cache 29s; total 29s; nproc 5 (CPU quota 4).

```sh
source /workspace/adamic-tools/env.sh
bash stage3/apply.sh /tmp/tsc-latent3-adapted > /tmp/latent3-apply.log 2>&1
python3 stage3/census/latent/run_comparisons.py /workspace/adamic /tmp/tsc-latent3-adapted /tmp/latent3-runs /tmp/latent3-worktrees.json > /tmp/latent3-comparisons.log 2>&1
python3 stage3/census/latent/meter3.py /tmp/latent3-runs /tmp/tsc-latent3-adapted > /tmp/latent3-summary.log 2>&1
python3 stage3/census/latent/audit_meter3.py /tmp/tsc-latent3-adapted > /tmp/latent3-audit.log 2>&1
python3 stage3/census/latent/audit_corpus.py /tmp/latent3-runs /tmp/tsc-latent3-adapted meter-3 > /tmp/latent3-corpus-audit.log 2>&1
GOWORK=/tmp/latent3-runs/meter-3.go.work python3 stage3/census/latent/audit_output_guards.py /tmp/latent3-tree /tmp/latent3-runs/meter-3-overlay /tmp/latent3-output-guards > /tmp/latent3-output-guards.log 2>&1
GOWORK=/tmp/latent3-runs/meter-3.go.work go vet -overlay=/tmp/latent3-runs/meter-3-overlay/overlay.json ./stage3/census/latent/tool > /tmp/latent3-vet.log 2>&1
```

Scratch merges were resolved to preserve taste labels/void, fallthrough and implicit returns, newest flags, namespace/parameter-property support, and nested captured storage. The previous integration enum helper was replaced by newest enums; a stale enumElement call found by the first build was removed. No compiler source edit is delivered. Integration diffs, pinned ancestry, dependency pins, loader options, adaptation hashes and failed-build log are archived in data/meter3/. Backends in the scratch merge are unused and their correctness is not claimed.

Dedicated mutants remove a cheap-win file or diagnostic, corrupt its code/line/cause, change the threshold count, ratio denominator, totals by code or source hash. The planted real-corpus NotYet must add exactly one binder.ts site with all other 77 files unchanged. Both output guards run in the baseline and corpus-mutant invocations; deliberate non-nil IR and permissive-loader mutants were also caught by their guard panics. Vet passes for the overlaid measurement driver. No full gate, native program, adaptation oracle suite or semantic proof was run for this measurement update.
