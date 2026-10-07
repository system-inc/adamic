Whole program: main: 0/78; area: 1/79
Own file: main: 22/78; area: 25/79

Stage 3 meter 20261007T062959Z

Main: origin/main at b8fb957aa839a9e8cb0b54279dd9864fa317bd30. Area: origin/area/stage3 at e39a299323cad76aea44ec71ca7b740130323d4c.
Both measured with Adamic e39a299323cad76aea44ec71ca7b740130323d4c and ordinary stage 0 options.

| File | Main whole program | Main own file | Area whole program | Area own file |
| --- | --- | --- | --- | --- |
| src/compiler/_namespaces/ts.moduleSpecifiers.ts | fail | pass | fail | pass |
| src/compiler/_namespaces/ts.performance.ts | fail | pass | fail | pass |
| src/compiler/_namespaces/ts.ts | fail | pass | fail | pass |
| src/compiler/binder.ts | fail | fail | fail | fail |
| src/compiler/builder.ts | fail | fail | fail | fail |
| src/compiler/builderPublic.ts | fail | pass | fail | pass |
| src/compiler/builderState.ts | fail | fail | fail | fail |
| src/compiler/builderStatePublic.ts | fail | pass | fail | pass |
| src/compiler/checker.ts | fail | fail | fail | fail |
| src/compiler/commandLineParser.ts | fail | fail | fail | fail |
| src/compiler/core.ts | fail | fail | fail | fail |
| src/compiler/corePublic.ts | fail | fail | fail | fail |
| src/compiler/debug.ts | fail | fail | fail | fail |
| src/compiler/diagnosticInformationMap.generated.ts | fail | fail | fail | pass |
| src/compiler/diagnosticMessages.generated.json | fail | fail | fail | fail |
| src/compiler/diagnosticMessages.json | fail | fail | fail | fail |
| src/compiler/emitter.ts | fail | fail | fail | fail |
| src/compiler/executeCommandLine.ts | fail | fail | fail | fail |
| src/compiler/expressionToTypeNode.ts | fail | fail | fail | fail |
| src/compiler/factory/baseNodeFactory.ts | fail | pass | fail | pass |
| src/compiler/factory/emitHelpers.ts | fail | fail | fail | fail |
| src/compiler/factory/emitNode.ts | fail | fail | fail | fail |
| src/compiler/factory/nodeChildren.ts | fail | pass | fail | pass |
| src/compiler/factory/nodeConverters.ts | fail | pass | fail | pass |
| src/compiler/factory/nodeFactory.ts | fail | fail | fail | fail |
| src/compiler/factory/nodeTests.ts | fail | pass | fail | pass |
| src/compiler/factory/parenthesizerRules.ts | fail | fail | fail | fail |
| src/compiler/factory/utilities.ts | fail | fail | fail | fail |
| src/compiler/factory/utilitiesPublic.ts | fail | pass | fail | pass |
| src/compiler/hostErrors.ts | absent | absent | pass | pass |
| src/compiler/moduleNameResolver.ts | fail | fail | fail | fail |
| src/compiler/moduleSpecifiers.ts | fail | fail | fail | fail |
| src/compiler/parser.ts | fail | fail | fail | fail |
| src/compiler/path.ts | fail | fail | fail | fail |
| src/compiler/performance.ts | fail | pass | fail | pass |
| src/compiler/performanceCore.ts | fail | fail | fail | fail |
| src/compiler/program.ts | fail | fail | fail | fail |
| src/compiler/programDiagnostics.ts | fail | pass | fail | pass |
| src/compiler/resolutionCache.ts | fail | fail | fail | fail |
| src/compiler/scanner.ts | fail | fail | fail | fail |
| src/compiler/semver.ts | fail | pass | fail | pass |
| src/compiler/sourcemap.ts | fail | fail | fail | fail |
| src/compiler/symbolWalker.ts | fail | pass | fail | pass |
| src/compiler/sys.ts | fail | fail | fail | fail |
| src/compiler/tracing.ts | fail | fail | fail | fail |
| src/compiler/transformer.ts | fail | fail | fail | fail |
| src/compiler/transformers/classFields.ts | fail | fail | fail | fail |
| src/compiler/transformers/classThis.ts | fail | pass | fail | pass |
| src/compiler/transformers/declarations.ts | fail | fail | fail | fail |
| src/compiler/transformers/declarations/diagnostics.ts | fail | fail | fail | fail |
| src/compiler/transformers/destructuring.ts | fail | fail | fail | fail |
| src/compiler/transformers/es2015.ts | fail | fail | fail | fail |
| src/compiler/transformers/es2016.ts | fail | pass | fail | pass |
| src/compiler/transformers/es2017.ts | fail | fail | fail | fail |
| src/compiler/transformers/es2018.ts | fail | fail | fail | fail |
| src/compiler/transformers/es2019.ts | fail | pass | fail | pass |
| src/compiler/transformers/es2020.ts | fail | pass | fail | pass |
| src/compiler/transformers/es2021.ts | fail | pass | fail | pass |
| src/compiler/transformers/esDecorators.ts | fail | fail | fail | fail |
| src/compiler/transformers/esnext.ts | fail | fail | fail | fail |
| src/compiler/transformers/generators.ts | fail | fail | fail | fail |
| src/compiler/transformers/jsx.ts | fail | fail | fail | fail |
| src/compiler/transformers/legacyDecorators.ts | fail | fail | fail | fail |
| src/compiler/transformers/module/esnextAnd2015.ts | fail | fail | fail | fail |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts | fail | pass | fail | pass |
| src/compiler/transformers/module/module.ts | fail | fail | fail | fail |
| src/compiler/transformers/module/system.ts | fail | fail | fail | fail |
| src/compiler/transformers/namedEvaluation.ts | fail | pass | fail | pass |
| src/compiler/transformers/taggedTemplate.ts | fail | fail | fail | fail |
| src/compiler/transformers/ts.ts | fail | fail | fail | fail |
| src/compiler/transformers/typeSerializer.ts | fail | pass | fail | pass |
| src/compiler/transformers/utilities.ts | fail | fail | fail | fail |
| src/compiler/tsbuild.ts | fail | fail | fail | fail |
| src/compiler/tsbuildPublic.ts | fail | fail | fail | fail |
| src/compiler/tsconfig.json | fail | fail | fail | fail |
| src/compiler/types.ts | fail | fail | fail | fail |
| src/compiler/utilities.ts | fail | fail | fail | fail |
| src/compiler/utilitiesPublic.ts | fail | fail | fail | fail |
| src/compiler/visitorPublic.ts | fail | fail | fail | pass |
| src/compiler/watch.ts | fail | fail | fail | fail |
| src/compiler/watchPublic.ts | fail | fail | fail | fail |
| src/compiler/watchUtilities.ts | fail | fail | fail | fail |

### Unowned

measured on a checker-rejected program.
OWNER BLANK rows go to @system_adamic. Counts are grouped by exact reason.

| Reason | Owner | Main NotYet | Main Refused | Area NotYet | Area Refused |
| --- | --- | ---: | ---: | ---: | ---: |
| a value as a condition | OWNER BLANK | 0 | 56 | 0 | 56 |
| a function returning T | OWNER BLANK | 50 | 0 | 50 | 0 |
| reading CharacterCodes | OWNER BLANK | 37 | 0 | 37 | 0 |
| a BinaryExpression with a value and a value | OWNER BLANK | 32 | 0 | 32 | 0 |
| a value of type T | OWNER BLANK | 30 | 0 | 30 | 0 |
| a value of type any | OWNER BLANK | 30 | 0 | 30 | 0 |
| a field of type boolean &#124; undefined | OWNER BLANK | 25 | 0 | 25 | 0 |
| a cast the runtime can&#x27;t check | OWNER BLANK | 0 | 23 | 0 | 23 |
| a parameter that isn&#x27;t a plain name | OWNER BLANK | 22 | 0 | 22 | 0 |
| reading ModifierFlags | OWNER BLANK | 22 | 0 | 22 | 0 |

<details>
<summary>348 more unowned reasons</summary>

| Reason | Owner | Main NotYet | Main Refused | Area NotYet | Area Refused |
| --- | --- | ---: | ---: | ---: | ---: |
| reading Extension | OWNER BLANK | 20 | 0 | 21 | 0 |
| reading NodeFlags | OWNER BLANK | 20 | 0 | 20 | 0 |
| a value of type Path | OWNER BLANK | 19 | 0 | 19 | 0 |
| a BinaryExpression with a value and a boolean | OWNER BLANK | 18 | 0 | 18 | 0 |
| a value of type ResolvedConfigFilePath | OWNER BLANK | 18 | 0 | 18 | 0 |
| a value of type __String | OWNER BLANK | 18 | 0 | 18 | 0 |
| &#124;&#124;= | OWNER BLANK | 0 | 15 | 0 | 18 |
| a function returning U &#124; undefined | OWNER BLANK | 16 | 0 | 16 | 0 |
| a string as a condition | OWNER BLANK | 0 | 14 | 0 | 14 |
| a generic function as a value | OWNER BLANK | 13 | 0 | 13 | 0 |
| reading Comparison | OWNER BLANK | 12 | 0 | 12 | 0 |
| a boolean &#124; undefined as a condition | OWNER BLANK | 0 | 11 | 0 | 12 |
| a namespace | OWNER BLANK | 0 | 11 | 0 | 11 |
| a value of type T &#124; undefined | OWNER BLANK | 11 | 0 | 11 | 0 |
| a value of type unknown | OWNER BLANK | 9 | 0 | 13 | 0 |
| reading ModuleKind | OWNER BLANK | 11 | 0 | 11 | 0 |
| a function returning __String | OWNER BLANK | 10 | 0 | 10 | 0 |
| a call through ?. (an optional call) | OWNER BLANK | 9 | 0 | 10 | 0 |
| an index signature | OWNER BLANK | 0 | 8 | 0 | 11 |
| a ModuleDeclaration | OWNER BLANK | 9 | 0 | 9 | 0 |
| an array of T | OWNER BLANK | 9 | 0 | 9 | 0 |
| a PrefixUnaryExpression on a string | OWNER BLANK | 8 | 0 | 9 | 0 |
| a Map whose keys aren&#x27;t strings, numbers, booleans, objects, arrays, maps or functions | OWNER BLANK | 8 | 0 | 8 | 0 |
| a PrefixUnaryExpression on a number | OWNER BLANK | 8 | 0 | 8 | 0 |
| new an Identifier | OWNER BLANK | 8 | 0 | 8 | 0 |
| a BinaryExpression as a statement | OWNER BLANK | 7 | 0 | 7 | 0 |
| an array of never | OWNER BLANK | 7 | 0 | 7 | 0 |
| for...of over an object | OWNER BLANK | 7 | 0 | 7 | 0 |
| reading EmitFlags | OWNER BLANK | 7 | 0 | 7 | 0 |
| reading NodeResolutionFeatures | OWNER BLANK | 7 | 0 | 7 | 0 |
| this outside a method | OWNER BLANK | 7 | 0 | 7 | 0 |
| a BinaryExpression with a boolean and a value | OWNER BLANK | 6 | 0 | 6 | 0 |
| a BinaryExpression with a number and a number | OWNER BLANK | 6 | 0 | 6 | 0 |
| a BinaryExpression with a number &#124; undefined and a number | OWNER BLANK | 6 | 0 | 6 | 0 |
| a PrefixUnaryExpression on a number &#124; undefined | OWNER BLANK | 6 | 0 | 6 | 0 |
| a field of type true &#124; undefined | OWNER BLANK | 6 | 0 | 6 | 0 |
| a function returning any | OWNER BLANK | 6 | 0 | 6 | 0 |
| a value of type ResolvedConfigFileName | OWNER BLANK | 6 | 0 | 6 | 0 |
| reading AssignmentDeclarationKind | OWNER BLANK | 6 | 0 | 6 | 0 |
| reading ModuleResolutionKind | OWNER BLANK | 6 | 0 | 6 | 0 |
| a method read as a value (liftToBlock would lose its object, and this with it) | OWNER BLANK | 0 | 5 | 0 | 6 |
| a method read as a value (readFile would lose its object, and this with it) | OWNER BLANK | 0 | 5 | 0 | 6 |
| a BinaryExpression with a string and a string | OWNER BLANK | 5 | 0 | 5 | 0 |
| a BinaryExpression with a value and a number | OWNER BLANK | 5 | 0 | 5 | 0 |
| a PrefixUnaryExpression on a boolean &#124; undefined | OWNER BLANK | 5 | 0 | 5 | 0 |
| a definite assignment assertion ! | OWNER BLANK | 0 | 5 | 0 | 5 |
| a function returning Path | OWNER BLANK | 5 | 0 | 5 | 0 |
| a generator function | OWNER BLANK | 0 | 5 | 0 | 5 |
| a label | OWNER BLANK | 0 | 5 | 0 | 5 |
| a method read as a value (trace would lose its object, and this with it) | OWNER BLANK | 0 | 5 | 0 | 5 |
| a value of type NonNullable&lt;T&gt; | OWNER BLANK | 5 | 0 | 5 | 0 |
| a value of type object | OWNER BLANK | 5 | 0 | 5 | 0 |
| for...in without a proven fixed plain-object origin (arrays, prototypes and absent synthetic fields cannot be enumerated soundly) | OWNER BLANK | 5 | 0 | 5 | 0 |
| reading Extensions | OWNER BLANK | 5 | 0 | 5 | 0 |
| reading ScriptTarget | OWNER BLANK | 5 | 0 | 5 | 0 |
| reading SymbolFlags | OWNER BLANK | 5 | 0 | 5 | 0 |
| reading TransformFlags | OWNER BLANK | 5 | 0 | 5 | 0 |
| the comma operator | OWNER BLANK | 0 | 5 | 0 | 5 |
| yield (generators) | OWNER BLANK | 0 | 5 | 0 | 5 |
| a BinaryExpression with a string and a boolean | OWNER BLANK | 4 | 0 | 4 | 0 |
| a field of type string &#124; DiagnosticMessageChain | OWNER BLANK | 4 | 0 | 4 | 0 |
| a function returning PackageJson[K] &#124; undefined | OWNER BLANK | 4 | 0 | 4 | 0 |
| a function returning __String &#124; undefined | OWNER BLANK | 4 | 0 | 4 | 0 |
| a function with an optional or rest parameter, as a value | OWNER BLANK | 4 | 0 | 4 | 0 |
| a method in object destructuring | OWNER BLANK | 0 | 4 | 0 | 4 |
| a method read as a value (fileExists would lose its object, and this with it) | OWNER BLANK | 0 | 4 | 0 | 4 |
| a method read as a value (getSourceFile would lose its object, and this with it) | OWNER BLANK | 0 | 4 | 0 | 4 |
| a value of type T &#124; Program | OWNER BLANK | 4 | 0 | 4 | 0 |
| reading ScriptKind | OWNER BLANK | 4 | 0 | 4 | 0 |
| regex replacement other than a string | OWNER BLANK | 4 | 0 | 4 | 0 |
| Object.entries on a shape not proven by a plain literal or its const binding | OWNER BLANK | 3 | 0 | 3 | 0 |
| a computed field name | OWNER BLANK | 3 | 0 | 3 | 0 |
| a function returning CompilerOptionsValue | OWNER BLANK | 3 | 0 | 3 | 0 |
| a function returning ResolvedConfigFileName | OWNER BLANK | 3 | 0 | 3 | 0 |
| a function returning T &#124; T[] &#124; undefined | OWNER BLANK | 3 | 0 | 3 | 0 |
| a function returning T &#124; readonly T[] &#124; undefined | OWNER BLANK | 3 | 0 | 3 | 0 |
| a method read as a value (afterProgramCreate would lose its object, and this with it) | OWNER BLANK | 0 | 3 | 0 | 3 |
| a method read as a value (afterProgramEmitAndDiagnostics would lose its object, and this with it) | OWNER BLANK | 0 | 3 | 0 | 3 |
| a method read as a value (directoryExists would lose its object, and this with it) | OWNER BLANK | 0 | 3 | 0 | 3 |
| a method read as a value (getDirectories would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 5 |
| a method read as a value (now would lose its object, and this with it) | OWNER BLANK | 0 | 3 | 0 | 3 |
| a method read as a value (onWatchStatusChange would lose its object, and this with it) | OWNER BLANK | 0 | 3 | 0 | 3 |
| a union of differently held members as a condition | OWNER BLANK | 0 | 3 | 0 | 3 |
| a value of type K | OWNER BLANK | 3 | 0 | 3 | 0 |
| a value of type string &#124; (void &amp; { __escapedIdentifier: void; }) &#124; (string &amp; { __escapedIdentifier: void; }) | OWNER BLANK | 3 | 0 | 3 | 0 |
| a value of type string &#124; null &#124; undefined | OWNER BLANK | 3 | 0 | 3 | 0 |
| an ElementAccessExpression | OWNER BLANK | 3 | 0 | 3 | 0 |
| an array of U | OWNER BLANK | 3 | 0 | 3 | 0 |
| in | OWNER BLANK | 0 | 2 | 0 | 4 |
| reading AccessKind | OWNER BLANK | 3 | 0 | 3 | 0 |
| reading OuterExpressionKinds | OWNER BLANK | 3 | 0 | 3 | 0 |
| reading TypeFlags | OWNER BLANK | 3 | 0 | 3 | 0 |
| reading UsingKind | OWNER BLANK | 3 | 0 | 3 | 0 |
| the void operator | OWNER BLANK | 0 | 3 | 0 | 3 |
| a BinaryExpression with a boolean and a string | OWNER BLANK | 2 | 0 | 2 | 0 |
| a BinaryExpression with a value and a string | OWNER BLANK | 2 | 0 | 2 | 0 |
| a PrefixUnaryExpression on a union of differently held members | OWNER BLANK | 2 | 0 | 2 | 0 |
| a declaration directly in a case (wrap the case in a block) | OWNER BLANK | 2 | 0 | 2 | 0 |
| a field of type &quot;boolean&quot; &#124; &quot;list&quot; &#124; &quot;listOrElement&quot; &#124; &quot;number&quot; &#124; &quot;object&quot; &#124; &quot;string&quot; &#124; Map&lt;string, string &#124; number&gt; | OWNER BLANK | 2 | 0 | 2 | 0 |
| a field of type boolean &#124; (() =&gt; boolean) &#124; undefined | OWNER BLANK | 2 | 0 | 2 | 0 |
| a field of type true &#124; Node &#124; undefined | OWNER BLANK | 2 | 0 | 2 | 0 |
| a function inside a function (a closure) | OWNER BLANK | 2 | 0 | 2 | 0 |
| a function returning Extract&lt;ClassDeclaration, Pick&lt;...&gt;&gt; &#124; Extract&lt;...&gt; | OWNER BLANK | 2 | 0 | 2 | 0 |
| a function returning Path &#124; undefined | OWNER BLANK | 2 | 0 | 2 | 0 |
| a function returning U | OWNER BLANK | 2 | 0 | 2 | 0 |
| a function returning V | OWNER BLANK | 2 | 0 | 2 | 0 |
| a function returning object | OWNER BLANK | 2 | 0 | 2 | 0 |
| a function value returning boolean &#124; undefined | OWNER BLANK | 2 | 0 | 2 | 0 |
| a function value returning union of differently held members | OWNER BLANK | 2 | 0 | 2 | 0 |
| a method read as a value (clearTimeout would lose its object, and this with it) | OWNER BLANK | 0 | 2 | 0 | 2 |
| a method read as a value (createDirectory would lose its object, and this with it) | OWNER BLANK | 0 | 2 | 0 | 2 |
| a method read as a value (enableCPUProfiler would lose its object, and this with it) | OWNER BLANK | 0 | 2 | 0 | 2 |
| a method read as a value (getCanonicalFileName would lose its object, and this with it) | OWNER BLANK | 0 | 2 | 0 | 2 |
| a method read as a value (getEnvironmentVariable would lose its object, and this with it) | OWNER BLANK | 0 | 2 | 0 | 2 |
| a method read as a value (getGlobalTypingsCacheLocation would lose its object, and this with it) | OWNER BLANK | 0 | 2 | 0 | 2 |
| a method read as a value (getParsedCommandLine would lose its object, and this with it) | OWNER BLANK | 0 | 2 | 0 | 2 |
| a method read as a value (readDirectory would lose its object, and this with it) | OWNER BLANK | 0 | 0 | 0 | 4 |
| a method read as a value (setPrototypeOf would lose its object, and this with it) | OWNER BLANK | 0 | 2 | 0 | 2 |
| a method read as a value (setTimeout would lose its object, and this with it) | OWNER BLANK | 0 | 2 | 0 | 2 |
| a method read as a value (toKey would lose its object, and this with it) | OWNER BLANK | 0 | 2 | 0 | 2 |
| a method read as a value (trackSymbol would lose its object, and this with it) | OWNER BLANK | 0 | 2 | 0 | 2 |
| a method read as a value (useCaseSensitiveFileNames would lose its object, and this with it) | OWNER BLANK | 0 | 2 | 0 | 2 |
| a method read as a value (watchDirectory would lose its object, and this with it) | OWNER BLANK | 0 | 2 | 0 | 2 |
| a method read as a value (watchFile would lose its object, and this with it) | OWNER BLANK | 0 | 2 | 0 | 2 |
| a method read as a value (writeFile would lose its object, and this with it) | OWNER BLANK | 0 | 2 | 0 | 2 |
| a number &#124; undefined as a condition | OWNER BLANK | 0 | 2 | 0 | 2 |
| a tagged template other than the intrinsic String.raw | OWNER BLANK | 2 | 0 | 2 | 0 |
| a value of type (EmitNode &amp; { autoGenerate: AutoGenerateInfo; }) &#124; (EmitNode &amp; { autoGenerate: AutoGenerateInfo; }) | OWNER BLANK | 2 | 0 | 2 | 0 |
| a value of type CompilerOptionsValue | OWNER BLANK | 2 | 0 | 2 | 0 |
| a value of type T &#124; T[] | OWNER BLANK | 2 | 0 | 2 | 0 |
| a value of type TInArray | OWNER BLANK | 2 | 0 | 2 | 0 |
| a value of type WrappedExpression&lt;AnonymousFunctionDefinition&gt; | OWNER BLANK | 2 | 0 | 2 | 0 |
| new a ParenthesizedExpression | OWNER BLANK | 2 | 0 | 2 | 0 |
| reading BuilderFileEmit | OWNER BLANK | 2 | 0 | 2 | 0 |
| reading BuilderProgramKind | OWNER BLANK | 2 | 0 | 2 | 0 |
| reading DiagnosticCategory | OWNER BLANK | 2 | 0 | 2 | 0 |
| reading FileWatcherEventKind | OWNER BLANK | 2 | 0 | 2 | 0 |
| reading JsxEmit | OWNER BLANK | 2 | 0 | 2 | 0 |
| reading ModuleInstanceState | OWNER BLANK | 2 | 0 | 2 | 0 |
| reading NodeFactoryFlags | OWNER BLANK | 2 | 0 | 2 | 0 |
| reading SignatureFlags | OWNER BLANK | 2 | 0 | 2 | 0 |
| reading TypePredicateKind | OWNER BLANK | 2 | 0 | 2 | 0 |
| reading getOptionsNameMap | OWNER BLANK | 2 | 0 | 2 | 0 |
| var | OWNER BLANK | 0 | 2 | 0 | 2 |
| .length on a value | OWNER BLANK | 1 | 0 | 1 | 0 |
| ?.[] on a value | OWNER BLANK | 1 | 0 | 1 | 0 |
| Object.assign on a shape not proven by a plain literal or its const binding | OWNER BLANK | 1 | 0 | 1 | 0 |
| RegExp with a nonconstant pattern | OWNER BLANK | 1 | 0 | 1 | 0 |
| String as a value outside equality or typeof (overloaded calls and static properties need their own representation) | OWNER BLANK | 1 | 0 | 1 | 0 |
| a BinaryExpression with a boolean and a boolean &#124; undefined | OWNER BLANK | 1 | 0 | 1 | 0 |
| a BinaryExpression with a string and a number | OWNER BLANK | 1 | 0 | 1 | 0 |
| a ClassExpression | OWNER BLANK | 1 | 0 | 1 | 0 |
| a Map of T | OWNER BLANK | 1 | 0 | 1 | 0 |
| a PostfixUnaryExpression | OWNER BLANK | 1 | 0 | 1 | 0 |
| a SatisfiesExpression | OWNER BLANK | 1 | 0 | 1 | 0 |
| a YieldExpression as a statement | OWNER BLANK | 1 | 0 | 1 | 0 |
| a boolean &#124; undefined variable a function value captures | OWNER BLANK | 1 | 0 | 1 | 0 |
| a class instantiated with TOuterState | OWNER BLANK | 1 | 0 | 1 | 0 |
| a destructured parameter beside a parameter with a default | OWNER BLANK | 1 | 0 | 1 | 0 |
| a field from a boolean &#124; undefined variable | OWNER BLANK | 1 | 0 | 1 | 0 |
| a field of type AnyBuildOrder &#124; undefined | OWNER BLANK | 1 | 0 | 1 | 0 |
| a field of type NodeArray&lt;ParameterDeclaration&gt; &#124; readonly JSDocParameterTag[] | OWNER BLANK | 1 | 0 | 1 | 0 |
| a field of type false &#124; VersionPaths &#124; undefined | OWNER BLANK | 1 | 0 | 1 | 0 |
| a field of type string &#124; false &#124; undefined | OWNER BLANK | 1 | 0 | 1 | 0 |
| a field of type string &#124; number &#124; undefined | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning (AssignmentExpression&lt;EqualsToken&gt; &amp; { readonly left: GeneratedIdentifier; }) &#124; undefined | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning (ConstructorDeclaration &amp; { body: Block; }) &#124; undefined | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning AnyValidImportOrReExport | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning AnyValidImportOrReExport &#124; undefined | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning CanonicalKey | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning ClassNamedEvaluationHelperBlock | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning ClassThisAssignmentBlock | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning ExpressionWithTypeArguments &amp; { readonly expression: Identifier &#124; PropertyAccessEntityNameExpression; } | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning Extract&lt;AssignmentExpression&lt;EqualsToken&gt; &amp; { readonly left: Identifier; readonly right: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; }, Pick&lt;...&gt;&gt; &#124; ... 7 more ... &#124; Extract&lt;...&gt; | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning HasJSDoc &#124; undefined | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning MemberName &#124; (Expression &amp; (NumericLiteral &#124; StringLiteralLike)) | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning ModeAwareCacheKey | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning Node &#124; (TIn &amp; undefined) &#124; (TVisited &amp; undefined) | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning NodeArray&lt;Node&gt; &#124; (TInArray &amp; undefined) | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning NodeArray&lt;TOut&gt; &#124; (TInArray &amp; undefined) | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning PathPathComponents | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning ResolvedConfigFilePath | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning T &#124; T[] | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning T &#124; readonly T[] | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning T1 &amp; T2 | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning TEntry &#124; undefined | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning TOut | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning TOut &#124; (TIn &amp; undefined) &#124; (TVisited &amp; undefined) | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning TOut &#124; undefined | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning TPrivateEntry &#124; undefined | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning object &#124; undefined | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning readonly Node[] &#124; (TInArray &amp; undefined) | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning readonly TOut[] &#124; (TInArray &amp; undefined) | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning string &#124; object | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning undefined | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning unknown | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning void &#124; SolutionBuilder&lt;EmitAndSemanticDiagnosticsBuilderProgram&gt; | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning void &#124; SolutionBuilder&lt;EmitAndSemanticDiagnosticsBuilderProgram&gt; &#124; WatchOfConfigFile&lt;EmitAndSemanticDiagnosticsBuilderProgram&gt; | OWNER BLANK | 1 | 0 | 1 | 0 |
| a function returning void &#124; WatchOfConfigFile&lt;EmitAndSemanticDiagnosticsBuilderProgram&gt; | OWNER BLANK | 1 | 0 | 1 | 0 |
| a method read as a value (add would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (base64decode would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (base64encode would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (clearScreen would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (compare would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (convertToArrayAssignmentElement would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (convertToObjectAssignmentElement would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (createIntersectionTypeNode would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (createJSDocClassTag would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (createJSDocDeprecatedTag would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (createJSDocLink would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (createJSDocLinkCode would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (createJSDocLinkPlain would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (createJSDocOverrideTag would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (createJSDocPrivateTag would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (createJSDocProtectedTag would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (createJSDocPublicTag would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (createJSDocReadonlyTag would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (createUnionTypeNode would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (deleteFile would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (fill would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (getBuildInfo would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (getCurrentDirectory would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (getModifiedTime would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (getNearestAncestorDirectoryWithPackageJson would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (getOrCreateCacheForModuleName would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (globalCacheResolutionModuleName would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (hasOwnProperty would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (log would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (realpath would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (remove would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (repeat would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (replace would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (reportCyclicStructureError would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (reportInaccessibleThisError would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (reportInaccessibleUniqueSymbolError would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (reportInferenceFallback would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (reportLikelyUnsafeImportRequiredError would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (reportNonSerializableProperty would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (reportNonlocalAugmentation would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (reportPrivateInBaseOfClassExpression would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (reportTruncationError would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (setModifiedTime would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (toString would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a method read as a value (writeOutputIsTTY would lose its object, and this with it) | OWNER BLANK | 0 | 1 | 0 | 1 |
| a number as a condition | OWNER BLANK | 0 | 1 | 0 | 1 |
| a number &#124; undefined argument to substring | OWNER BLANK | 1 | 0 | 1 | 0 |
| a template interpolating an object, an array, a map, a function or undefined | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type &quot;&quot; &#124; ResolvedConfigFileName &#124; undefined | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type (AssignmentExpression&lt;EqualsToken&gt; &amp; { readonly left: Identifier; readonly right: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; } &amp; BinaryExpression) &#124; (... &amp; ... 1 more ... &amp; BinaryExpression) | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type (ConstructorDeclaration &amp; { body: Block; }) &#124; undefined | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type AccessExpression &#124; RequireOrImportCall | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type AccessorDeclaration &amp; { readonly name: BigIntLiteral &#124; ComputedPropertyName &#124; Identifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; StringLiteral; } | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type BindingElement &amp; { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; } | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type EmitNode &amp; { autoGenerate: AutoGenerateInfo; } | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type EntityNameExpression &#124; (LeftHandSideExpression &amp; BindableStaticNameExpression) | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type ExportAssignment &amp; { readonly expression: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; } | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type HasJSDoc | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type HasJSDoc &#124; undefined | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type IncludeTypeSpaceImports | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type IncrementalBuildInfoFilePendingEmit | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type IncrementalMultiFileEmitBuildInfoFileInfo | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type JSDocImportTag &#124; CanHaveModuleSpecifier | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type NamedEvaluation | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type ParameterDeclaration &amp; { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; } | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type PropertyAssignment &amp; { readonly name: Identifier; readonly initializer: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; } | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type PropertyDeclaration &amp; { readonly initializer: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; } | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type RequireOrImportCall | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type ResolvedConfigFileName &#124; undefined | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type ResolvedModuleWithFailedLookupLocations &amp; ResolvedTypeReferenceDirectiveWithFailedLookupLocations | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type ShorthandPropertyAssignment &amp; { readonly objectAssignmentInitializer: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; } | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type SourceFile | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type T &#124; readonly T[] | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type T1 | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type TData | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type TEntry | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type T[&quot;kind&quot;] | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type TypeParameterDeclaration &amp; { parent: JSDocTemplateTag; } | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type U | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type U &#124; readonly U[] &#124; undefined | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type V | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type VariableDeclaration &amp; { readonly name: Identifier; readonly initializer: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; } | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type WatchFactoryHost &amp; { trace?(s: string): void; } | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type WrappedExpression&lt;T&gt; | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type __String &amp; string | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type false &#124; RegExpExecArray &#124; null | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type object &#124; undefined | OWNER BLANK | 1 | 0 | 1 | 0 |
| a value of type undefined | OWNER BLANK | 1 | 0 | 1 | 0 |
| a void call used as a value | OWNER BLANK | 1 | 0 | 1 | 0 |
| an array of T &#124; U | OWNER BLANK | 1 | 0 | 1 | 0 |
| an array of V | OWNER BLANK | 1 | 0 | 1 | 0 |
| an array of unknown | OWNER BLANK | 1 | 0 | 1 | 0 |
| an import cycle | OWNER BLANK | 0 | 1 | 0 | 1 |
| arguments | OWNER BLANK | 0 | 1 | 0 | 1 |
| assigning a field of a value | OWNER BLANK | 1 | 0 | 1 | 0 |
| debugger | OWNER BLANK | 0 | 1 | 0 | 1 |
| delete | OWNER BLANK | 0 | 1 | 0 | 1 |
| inherited library member hasOwnProperty read as an own field | OWNER BLANK | 0 | 1 | 0 | 1 |
| inherited library member prototype read as an own field | OWNER BLANK | 0 | 1 | 0 | 1 |
| inherited library member replace read as an own field | OWNER BLANK | 0 | 1 | 0 | 1 |
| instantiating a generic function makes a value of type T &#124; undefined written where T is read | OWNER BLANK | 0 | 1 | 0 | 1 |
| lastIndexOf with these arguments | OWNER BLANK | 1 | 0 | 1 | 0 |
| new Map from something that isn&#x27;t [key, value] pairs | OWNER BLANK | 1 | 0 | 1 | 0 |
| optional chaining to .size on a value | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading Error | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading FileIncludeKind | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading FlattenLevel | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading ImportsNotUsedAsValues | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading Instruction | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading InternalNodeBuilderFlags | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading IntrinsicTypeKind | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading IterationTypeKind | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading JSDocParsingMode | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading ListFormat | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading ModuleDetectionKind | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading ModuleSpecifierEnding | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading NewLineKind | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading NodeBuilderFlags | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading PragmaKindFlags | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading RegularExpressionFlags | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading SnippetKind | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading StatisticType | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading TypeFacts | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading UpToDateStatusType | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading WatchFileKind | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading addAggregateStatistic | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading addOutput | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading captureMapping | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading convertToFunctionBlock | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading createBaseSourceFileNode | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading createIntlCollatorStringComparer | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading createPollingIntervalQueue | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading enter | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading getAccessorNameVisibilityError | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading getPackageJsonInfo | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading getSymbolWalker | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading optionDependsOnRecursive | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading parseStrings | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading transformSourceFile | OWNER BLANK | 1 | 0 | 1 | 0 |
| reading transformSourceFileOrBundle | OWNER BLANK | 1 | 0 | 1 | 0 |
| spreading an array of other elements | OWNER BLANK | 1 | 0 | 1 | 0 |
| storing any in a field | OWNER BLANK | 1 | 0 | 1 | 0 |
| storing string &#124; number in a field | OWNER BLANK | 1 | 0 | 1 | 0 |
| storing true &#124; Node &#124; undefined in a field | OWNER BLANK | 1 | 0 | 1 | 0 |
| a method read as a value (getMemoryUsage would lose its object, and this with it) | OWNER BLANK | 0 | 0 | 0 | 1 |
| reading WatchLogLevel | OWNER BLANK | 0 | 0 | 1 | 0 |
| reading getUnusedExpectations | OWNER BLANK | 0 | 0 | 1 | 0 |
| reading getValueCandidate | OWNER BLANK | 0 | 0 | 1 | 0 |

</details>

Latent lowering, main: measured on a checker-rejected program.
NotYet: 2059; Refused: 2140.

| Reason | Owner | NotYet | Refused | Total |
| --- | --- | ---: | ---: | ---: |
| a type predicate | 01a1143b-d691 | 0 | 588 | 588 |
| reading SyntaxKind | compiler/stage3-front | 484 | 0 | 484 |
| the non-null assertion ! | 01a1130a | 0 | 484 | 484 |
| a function without a body | 01a1143b-f5d4 | 191 | 0 | 191 |
| enum | compiler/stage3-front | 0 | 162 | 162 |
| an EnumDeclaration | compiler/stage3-front | 154 | 0 | 154 |
| a PrefixUnaryExpression on a value | 01a1143b-f5d4 | 85 | 0 | 85 |
| a method call through a structural signature in a program with statics; use typeof the declaring class | 01a1143c | 80 | 0 | 80 |
| an ExportDeclaration | 01a113e3-a058 | 0 | 77 | 77 |
| a function returning T &#124; undefined | 01a1143c | 66 | 0 | 66 |

Counts are unique finding sites, not attempt events. Skipped dependencies, errors and panics
are retained separately in JSON. This measurement does not establish successful lowering or native output.

Latent lowering, area: measured on a checker-rejected program.
NotYet: 2075; Refused: 2104.

| Reason | Owner | NotYet | Refused | Total |
| --- | --- | ---: | ---: | ---: |
| a type predicate | 01a1143b-d691 | 0 | 588 | 588 |
| the non-null assertion ! | 01a1130a | 0 | 520 | 520 |
| reading SyntaxKind | compiler/stage3-front | 484 | 0 | 484 |
| a function without a body | 01a1143b-f5d4 | 191 | 0 | 191 |
| enum | compiler/stage3-front | 0 | 162 | 162 |
| an EnumDeclaration | compiler/stage3-front | 154 | 0 | 154 |
| a PrefixUnaryExpression on a value | 01a1143b-f5d4 | 86 | 0 | 86 |
| a method call through a structural signature in a program with statics; use typeof the declaring class | 01a1143c | 82 | 0 | 82 |
| an ExportDeclaration | 01a113e3-a058 | 0 | 77 | 77 |
| a function returning T &#124; undefined | 01a1143c | 66 | 0 | 66 |

Counts are unique finding sites, not attempt events. Skipped dependencies, errors and panics
are retained separately in JSON. This measurement does not establish successful lowering or native output.

Whole program includes imported diagnostics. Own file uses the primary diagnostic location.
Global and external diagnostics are counted separately in JSON; they have no compiler-file location.
Non-source inputs fail the extension gate and are excluded from source denominators.
