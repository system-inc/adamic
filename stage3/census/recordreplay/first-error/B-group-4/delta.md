# Paired latent census

| Measure | Raw findings | Distinct reads |
| --- | ---: | ---: |
| Lowered (no remaining finding) | 0 | 0 |
| Moved by reason | 0 | 0 |
| New findings | 0 | 0 |
| New sites | 0 | 0 |
| Units whose first stop moved | 0 | 0 |
| Regressions | 0 | 0 |
| before NotYet | 96 | 96 |
| before Refused | 460 | 438 |
| after NotYet | 96 | 96 |
| after Refused | 460 | 438 |

| Reason | Raw before | Raw after | Reads before | Reads after | Reads lowered |
| --- | ---: | ---: | ---: | ---: | ---: |
| Object enumeration with getter or accessor descriptors | 4 | 4 | 4 | 4 | 0 |
| Object.fromEntries | 1 | 1 | 1 | 1 | 0 |
| RegExp with a nonconstant pattern | 1 | 1 | 1 | 1 | 0 |
| a BaseParsedCommandLine seen as ParsedCommandLine (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a BuildOptions seen as { extendedDiagnostics?: boolean; diagnostics?: boolean; } (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a CompilerOptions seen as MapLike<TsConfigSourceFile \| CompilerOptionsValue> (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a CompilerOptions seen as MapLike<string \| number \| boolean \| string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined> (fixed objects and records have different storage; copy explicitly) | 2 | 2 | 2 | 2 | 0 |
| a CompilerOptions seen as OptionsBase (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a Map of CompilerOptionsValue | 1 | 1 | 1 | 1 | 0 |
| a Map or Set going where a plain object type is held | 1 | 1 | 1 | 1 | 0 |
| a Map<string, JsxEmit> where Map<string, string \| number> goes: its JsxEmit elements are stored as number and read there as string \| number, held as a boxed union (build it as the slot's type where it's made, or copy its elements into one) | 1 | 1 | 1 | 1 | 0 |
| a MapLike<string \| number \| boolean \| string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| null \| undefined> \| undefined seen as WatchOptions \| undefined (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a Record<string, CompilerOptionsValue> seen as CompilerOptions (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a SpreadElement | 1 | 1 | 1 | 1 | 0 |
| a WatchOptions seen as MapLike<string \| number \| boolean \| string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| null \| undefined> (fixed objects and records have different storage; copy explicitly) | 2 | 2 | 2 | 2 | 0 |
| a WatchOptions seen as OptionsBase (fixed objects and records have different storage; copy explicitly) | 2 | 2 | 2 | 2 | 0 |
| a WatchOptions \| undefined seen as OptionsBase \| undefined (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a boolean \| undefined variable a function value captures | 2 | 2 | 2 | 2 | 0 |
| a call spreading elements that cannot be packed | 1 | 1 | 1 | 1 | 0 |
| a call through ?. (an optional call) | 1 | 1 | 1 | 1 | 0 |
| a case whose type differs from the switch's | 2 | 2 | 2 | 2 | 0 |
| a cast that makes a viewed readonly array writable | 1 | 1 | 1 | 1 | 0 |
| a cast the runtime can't check | 92 | 92 | 92 | 92 | 0 |
| a checked predicate target whose base has no required field to check ({ [key: string]: unknown; }) | 1 | 1 | 1 | 1 | 0 |
| a checked predicate target whose check would read past the depth bound (BuilderProgramStateWithDefinedProgram) | 1 | 1 | 1 | 1 | 0 |
| a checked predicate target whose check would read past the depth bound (CircularBuildOrder) | 1 | 1 | 1 | 1 | 0 |
| a checked predicate target whose check would read past the depth bound (CommandLineOptionOfCustomType) | 1 | 1 | 1 | 1 | 0 |
| a checked predicate target whose check would read past the depth bound (IncrementalBundleEmitBuildInfo) | 1 | 1 | 1 | 1 | 0 |
| a dynamic getter or method read, or a nullable field (runtime dispatch, slot tags and narrowed rereads) | 2 | 2 | 2 | 2 | 0 |
| a filter callback that doesn't return a boolean | 2 | 2 | 2 | 2 | 0 |
| a function returning CanonicalKey | 1 | 1 | 1 | 1 | 0 |
| a function returning ImmediatelyInvokedArrowFunction | 1 | 1 | 1 | 1 | 0 |
| a function returning U \| undefined | 1 | 1 | 1 | 1 | 0 |
| a function taking Path seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 0 |
| a function taking Visitor seen as one taking Visitor<TIn, Node \| undefined> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking boolean seen as one taking boolean \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a generic instantiated at void whose body reads the value: getPropFromRaw's T = void (read from the signature), read at /workspace/scratch/recordreplay/candidate2/src/compiler/commandLineParser.ts:3253:24 as a returned value | 1 | 1 | 1 | 1 | 0 |
| a member typed any (config, declared): any has no proven type, so every read of it through a typed view is unchecked | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (createExpressionStatement would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getBuildInfo would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getCanonicalFileName would lose its object, and this with it) | 5 | 5 | 5 | 5 | 0 |
| a method read as a value (getParsedCommandLine would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getSourceFile would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (hasChangedEmitSignature would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method sharing its name with a getter; rename the getter or use a callable arrow field in the interface and implementation | 1 | 1 | 1 | 1 | 0 |
| a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables | 1 | 1 | 1 | 1 | 0 |
| a primitive brand member __canonicalKey whose type is not void | 1 | 1 | 1 | 1 | 0 |
| a read narrowed to undefined of a binding a call may write | 1 | 1 | 1 | 1 | 0 |
| a read the checker narrowed to never of a binding a call may write | 1 | 1 | 1 | 1 | 0 |
| a structural relation deeper than 64 type pairs while checking promise-returning functions seen as void-returning | 3 | 3 | 3 | 3 | 0 |
| a template interpolating a union with an object, an array, a map or a function in it | 1 | 1 | 1 | 1 | 0 |
| a template interpolating an object, an array, a map, a function or undefined | 1 | 1 | 1 | 1 | 0 |
| a value of type ((node: PrivateIdentifierPropertyDeclaration, modifiers: ModifiersArray \| undefined) => ObjectLiteralExpression) \| undefined seen as ((node: PropertyDeclaration & { readonly name: PrivateIdentifier; }, modifiers: ModifiersArray \| undefined) => Expression) \| undefined, whose readonly field name becomes writable: a readonly field may hold something narrower than PrivateIdentifier, which a write of PrivateIdentifier would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type (chain: DiagnosticMessageChain \| ReusableDiagnosticMessageChain) => RepopulateDiagnosticChainInfo \| undefined seen as (chain: DiagnosticMessageChain) => RepopulateDiagnosticChainInfo \| undefined, which can write ReusableDiagnosticMessageChain[] \| undefined where DiagnosticMessageChain[] \| undefined is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (node: PrivateIdentifierGetAccessorDeclaration, modifiers: ModifiersArray \| undefined) => ObjectLiteralExpression seen as ((node: GetAccessorDeclaration & { readonly name: PrivateIdentifier; }, modifiers: ModifiersArray \| undefined) => Expression) \| undefined, whose readonly field name becomes writable: a readonly field may hold something narrower than PrivateIdentifier, which a write of PrivateIdentifier would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type (node: PrivateIdentifierMethodDeclaration, modifiers: ModifiersArray \| undefined) => ObjectLiteralExpression seen as ((node: MethodDeclaration & { readonly name: PrivateIdentifier; }, modifiers: ModifiersArray \| undefined) => Expression) \| undefined, whose readonly field name becomes writable: a readonly field may hold something narrower than PrivateIdentifier, which a write of PrivateIdentifier would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type (node: PrivateIdentifierSetAccessorDeclaration, modifiers: ModifiersArray \| undefined) => ObjectLiteralExpression seen as ((node: SetAccessorDeclaration & { readonly name: PrivateIdentifier; }, modifiers: ModifiersArray \| undefined) => Expression) \| undefined, whose readonly field name becomes writable: a readonly field may hold something narrower than PrivateIdentifier, which a write of PrivateIdentifier would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type (sourceFile: SourceFile \| undefined, cancellationToken: CancellationToken \| undefined) => readonly DiagnosticWithLocation[] seen as (sourceFile?: SourceFile \| undefined, cancellationToken?: CancellationToken \| undefined) => readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Block seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type BuilderProgramState seen as ReusableBuilderProgramState, which can write Map<Path, readonly Diagnostic[] \| readonly ReusableDiagnostic[]> where Map<Path, readonly Diagnostic[]> is read | 9 | 9 | 9 | 9 | 0 |
| a value of type BuilderProgramStateWithDefinedProgram seen as BuilderProgramState, which can write Program \| undefined where Program is read | 12 | 12 | 12 | 12 | 0 |
| a value of type CallExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 10 | 10 | 10 | 10 | 0 |
| a value of type ClassElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ClassStaticBlockDeclaration \| PropertyDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type CompilerOptions \| TypeAcquisition \| WatchOptions \| undefined seen as CompilerOptions \| TypeAcquisition \| WatchOptions, which can write TsConfigSourceFile \| CompilerOptionsValue where CompilerOptionsValue is read | 1 | 1 | 1 | 1 | 0 |
| a value of type CompilerOptionsValue seen as TsConfigSourceFile \| CompilerOptionsValue, which can write string \| number where string is read | 5 | 5 | 5 | 5 | 0 |
| a value of type ConstructorDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Diagnostic seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 | 1 | 0 |
| a value of type DiagnosticMessageChain seen as DiagnosticMessageChain \| ReusableDiagnosticMessageChain, which can write ReusableDiagnosticMessageChain[] \| undefined where DiagnosticMessageChain[] \| undefined is read | 2 | 2 | 2 | 2 | 0 |
| a value of type DiagnosticMessageChain \| { messageText: string; category: DiagnosticCategory; code: number; repopulateInfo?: () => RepopulateDiagnosticChainInfo; canonicalHead?: CanonicalDiagnostic; next: ... \| undefined; } seen as ReusableDiagnosticMessageChain, which can write ReusableDiagnosticMessageChain[] \| undefined where DiagnosticMessageChain[] \| undefined is read | 1 | 1 | 1 | 1 | 0 |
| a value of type DiagnosticMessageChain[] seen as ReusableDiagnosticMessageChain[] \| undefined, which can write ReusableDiagnosticMessageChain where DiagnosticMessageChain is read | 1 | 1 | 1 | 1 | 0 |
| a value of type DiagnosticMessageChain[] seen as ReusableDiagnosticMessageChain[], which can write ReusableDiagnosticMessageChain[] \| undefined where DiagnosticMessageChain[] \| undefined is read | 1 | 1 | 1 | 1 | 0 |
| a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 5 | 5 | 5 | 5 | 0 |
| a value of type DiagnosticWithLocation \| undefined seen as Diagnostic \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Expression seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 4 | 4 | 4 | 4 | 0 |
| a value of type Expression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 16 | 16 | 16 | 16 | 0 |
| a value of type GeneratedIdentifier seen as Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Identifier seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type Identifier seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 4 | 4 | 4 | 4 | 0 |
| a value of type Identifier \| TextRange seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type Identifier \| undefined seen as Identifier \| TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type ImmediatelyInvokedArrowFunction seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type IncrementalMultiFileEmitBuildInfoFileInfo | 1 | 1 | 1 | 1 | 0 |
| a value of type LeftHandSideExpression & GeneratedIdentifier seen as Expression, which can write EmitNode \| undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 2 | 2 | 2 | 2 | 0 |
| a value of type Map<Path, Diagnostic[]> seen as Map<Path, readonly Diagnostic[]> \| undefined, which can write readonly Diagnostic[] where Diagnostic[] is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<Path, FileWatcher> seen as Map<string, FileWatcher>, which can write string where Path is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<Path, readonly ReusableDiagnostic[]> seen as Map<Path, readonly Diagnostic[] \| readonly ReusableDiagnostic[]>, which can write readonly Diagnostic[] \| readonly ReusableDiagnostic[] where readonly ReusableDiagnostic[] is read | 2 | 2 | 2 | 2 | 0 |
| a value of type Map<Path, readonly ReusableDiagnostic[]> \| undefined seen as Map<Path, readonly Diagnostic[] \| readonly ReusableDiagnostic[]> \| undefined, which can write readonly Diagnostic[] \| readonly ReusableDiagnostic[] where readonly ReusableDiagnostic[] is read | 2 | 2 | 2 | 2 | 0 |
| a value of type Map<ResolvedConfigFilePath, BuildInfoCacheEntry> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where BuildInfoCacheEntry is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<ResolvedConfigFilePath, ConfigFileCacheEntry> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where ConfigFileCacheEntry is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<ResolvedConfigFilePath, Map<Path, Date>> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where Map<Path, Date> is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<ResolvedConfigFilePath, ProgramUpdateLevel> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where ProgramUpdateLevel is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<ResolvedConfigFilePath, Set<string> \| undefined> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where Set<string> \| undefined is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<ResolvedConfigFilePath, T> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where T is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<ResolvedConfigFilePath, UpToDateStatus> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where UpToDateStatus is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<ResolvedConfigFilePath, readonly Diagnostic[]> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where readonly Diagnostic[] is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<ResolvedConfigFilePath, true> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where true is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<string, ImportsNotUsedAsValues> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ImportsNotUsedAsValues is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<string, JsxEmit> seen as Map<string, string \| number>, which can write string \| number where JsxEmit is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<string, ModuleDetectionKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ModuleDetectionKind is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<string, ModuleResolutionKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ModuleResolutionKind is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<string, NewLineKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where NewLineKind is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<string, PollingWatchKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where PollingWatchKind is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<string, WatchDirectoryKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where WatchDirectoryKind is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<string, WatchFileKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where WatchFileKind is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<string, string> seen as Map<string, string \| number>, which can write string \| number where string is read | 1 | 1 | 1 | 1 | 0 |
| a value of type MapLike<string \| number \| boolean \| string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined> seen as OptionsBase \| undefined, which can write string \| number where string is read | 1 | 1 | 1 | 1 | 0 |
| a value of type MethodDeclaration \| PropertyAssignment \| AccessorDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Node seen as Node \| SourceMapRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Node seen as Node \| TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Node \| SourceMapRange seen as SourceMapRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Node \| TextRange seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type NodeArray<ClassElement> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type NodeArray<Statement> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type NonNullable<T> | 1 | 1 | 1 | 1 | 0 |
| a value of type ParameterDeclaration seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ParameterDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type ParameterDeclaration \| PropertyDeclaration \| PropertySignature \| SignatureDeclaration seen as Mutable<ParameterDeclaration \| PropertyDeclaration \| PropertySignature \| SignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ParseConfigHost seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where boolean is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Path[] seen as string[], which can write string where Path is read | 1 | 1 | 1 | 1 | 0 |
| a value of type PrimitiveLiteral | 1 | 1 | 1 | 1 | 0 |
| a value of type PrivateEnvironment<{ className: undefined; weakSetName: undefined; }, PrivateIdentifierInfo> seen as PrivateEnvironment<PrivateEnvironmentData, PrivateIdentifierInfo>, which can write Identifier \| undefined where undefined is read | 1 | 1 | 1 | 1 | 0 |
| a value of type PrivateIdentifierInExpression | 1 | 1 | 1 | 1 | 0 |
| a value of type PropertyAccessExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type PropertyDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type PropertyName seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Readonly<ReusableBuilderProgramState> \| undefined seen as BuilderState \| undefined, whose readonly field fileInfos becomes writable: a readonly field may hold something narrower than Map<Path, FileInfo>, which a write of Map<Path, FileInfo> would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type SolutionBuilderHost<T> seen as ModuleResolutionHost, whose method useCaseSensitiveFileNames becomes a writable field: a write of boolean \| (() => boolean) \| undefined would replace a method of type () => boolean | 3 | 3 | 3 | 3 | 0 |
| a value of type SolutionBuilderWithWatchHost<T> seen as WatchFactoryHost & { trace?(s: string): void; }, whose method useCaseSensitiveFileNames becomes a writable field: a write of boolean \| (() => boolean) would replace a method of type () => boolean | 1 | 1 | 1 | 1 | 0 |
| a value of type StringLiteral seen as Mutable<StringLiteral>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type StringLiteral seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type TsConfigSourceFile \| CompilerOptionsValue seen as string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined, which can write string \| number where string is read | 2 | 2 | 2 | 2 | 0 |
| a value of type TypeAcquisition seen as CompilerOptions \| TypeAcquisition, which can write TsConfigSourceFile \| CompilerOptionsValue where CompilerOptionsValue is read | 1 | 1 | 1 | 1 | 0 |
| a value of type any: any has no proven type | 8 | 8 | 8 | 8 | 0 |
| a value of type never | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as (SourceMapRange \| undefined)[], which can write SourceMapRange \| undefined where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as Diagnostic[], which can write Diagnostic where never is read | 2 | 2 | 2 | 2 | 0 |
| a value of type never[] seen as ProjectReference[], which can write ProjectReference where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as ResolvedConfigFileName[], which can write ResolvedConfigFileName where never is read | 2 | 2 | 2 | 2 | 0 |
| a value of type number[] seen as string \| (string \| number)[] \| undefined, which can write string \| number where number is read | 1 | 1 | 1 | 1 | 0 |
| a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 2 | 2 | 2 | 2 | 0 |
| a value of type readonly IncrementalBundleEmitBuildInfoFileInfo[] \| readonly IncrementalMultiFileEmitBuildInfoFileInfo[] where an array goes | 1 | 1 | 1 | 1 | 0 |
| a value of type string \| FileInfo \| { version: string; signature: false; affectsGlobalScope: true \| undefined; impliedFormat: ResolutionMode; } seen as IncrementalMultiFileEmitBuildInfoFileInfo, which can write string \| false \| undefined where string \| undefined is read | 1 | 1 | 1 | 1 | 0 |
| a value of type string \| number \| boolean \| string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> seen as CompilerOptionsValue, which can write string \| number where string is read | 1 | 1 | 1 | 1 | 0 |
| a value of type string[] seen as CompilerOptionsValue, which can write string \| number where string is read | 5 | 5 | 5 | 5 | 0 |
| a value of type string[] seen as string \| (string \| number)[] \| undefined, which can write string \| number where string is read | 1 | 1 | 1 | 1 | 0 |
| a value of type string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] seen as unknown[], which can write unknown where string is read | 2 | 2 | 2 | 2 | 0 |
| a value of type { prepend?: boolean \| undefined; circular?: boolean \| undefined; path: string; originalPath: undefined; }[] \| undefined seen as readonly ProjectReference[] \| undefined, which can write string \| undefined where undefined is read | 1 | 1 | 1 | 1 | 0 |
| a value of type { referencedName: StringLiteral; name: PropertyName; } \| { referencedName: Identifier; name: ComputedPropertyName; } seen as { referencedName: Expression \| undefined; name: PropertyName \| undefined; }, which can write Expression \| undefined where StringLiteral is read | 1 | 1 | 1 | 1 | 0 |
| a write through a union to a slot (SourceMapRange \| undefined)[] doesn't name | 1 | 1 | 1 | 1 | 0 |
| a { config?: any; error?: Diagnostic \| undefined; } seen as { config?: AdamicJsonRecoveryObject; error?: Diagnostic \| undefined; } (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a {} \| undefined seen as AdamicJsonRecoveryValue \| undefined (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a {} \| undefined seen as TsConfigSourceFile \| CompilerOptionsValue (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| an array going where a plain object type is held | 2 | 2 | 2 | 2 | 0 |
| an assignment result with different storage | 1 | 1 | 1 | 1 | 0 |
| an assignment used as a value whose target write type Diagnostic cannot be read as void | 1 | 1 | 1 | 1 | 0 |
| an assignment used as a value whose target write type Diagnostic[] \| undefined cannot be read as void | 1 | 1 | 1 | 1 | 0 |
| an assignment used as a value whose target write type Program \| undefined cannot be read as void | 1 | 1 | 1 | 1 | 0 |
| an assignment value to a member | 1 | 1 | 1 | 1 | 0 |
| an element of an array literal a type argument is inferred from, whose type in its constraint {} can't be read | 5 | 5 | 5 | 5 | 0 |
| an indexed mutation without a closed const literal origin | 2 | 2 | 2 | 2 | 0 |
| an object refinement using an open numeric enum as a literal tag | 106 | 106 | 84 | 84 | 0 |
| an object with a nullable field viewed as unknown or object (null and undefined slot tags) | 1 | 1 | 1 | 1 | 0 |
| an optional Map get through a structural receiver | 2 | 2 | 2 | 2 | 0 |
| an unproven relation from SolutionBuilderHost<T> \| SolutionBuilderWithWatchHost<T> to SolutionBuilderHost<T>: optional field reportErrorSummary has no proven compatible presence/type | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from SolutionBuilderHostBase<T> to SolutionBuilderHost<T>: optional field reportErrorSummary has no proven compatible presence/type | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from string to T: the source is not assignable to the target | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot BuildStep.Done | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot BuilderFileEmit.Js \| BuilderFileEmit.JsMap \| BuilderFileEmit.JsInlineMap \| BuilderFileEmit.DtsErrors \| BuilderFileEmit.DtsEmit \| ... 5 more ... \| BuilderFileEmit.All | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot BuilderFileEmit.None | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot CompilerOptions \| BuilderFileEmit | 5 | 5 | 5 | 5 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ClassStaticBlockDeclaration | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ExpressionWithTypeArguments | 1 | 1 | 1 | 1 | 0 |
| an unrestricted dictionary write that can violate named property all; use a literal key or a value accepted by every named property | 1 | 1 | 1 | 1 | 0 |
| an unrestricted dictionary write that can violate named property configFile; use a literal key or a value accepted by every named property | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target NonIncrementalBuildInfo | 1 | 1 | 1 | 1 | 0 |
| forEach with a callback returning boolean \| undefined | 1 | 1 | 1 | 1 | 0 |
| in when a possibly absent spread has no property presence descriptors | 1 | 1 | 1 | 1 | 0 |
| inherited library member push read as an own field | 2 | 2 | 2 | 2 | 0 |
| instanceof against a value that isn't a declared class | 1 | 1 | 1 | 1 | 0 |
| method bind with an assigned receiver identifier | 3 | 3 | 3 | 3 | 0 |
| optional chaining to .size on a value | 3 | 3 | 3 | 3 | 0 |
| optional property all in CompilerOptions absent from structural source object, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property id in ArrayBindingPattern absent from structural source never, which can hide fields | 6 | 6 | 6 | 6 | 0 |
| optional property id in AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; } absent from structural source never, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property id in BigIntLiteral absent from structural source never, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property id in FalseLiteral absent from structural source never, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property id in PrivateIdentifier absent from structural source never, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property onUnRecoverableConfigFileDiagnostic in ProgramHost<T> & { onUnRecoverableConfigFileDiagnostic?: DiagnosticReporter; } absent from structural source SolutionBuilderHost<T>, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property repopulateInfo in DiagnosticMessageChain absent from structural source SerializedDiagnosticMessageChain, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property reportsUnnecessary in Diagnostic absent from structural source DiagnosticRelatedInformation, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property reportsUnnecessary in ReusableDiagnostic absent from structural source ReusableDiagnosticRelatedInformation, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property source in SourceMapRange absent from structural source TextRange, which can hide fields | 11 | 11 | 11 | 11 | 0 |
| optional property textSourceNode in StringLiteral absent from structural source never, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property watchOptions in TSConfig & { watchOptions?: object \| undefined; } absent from structural source { include: readonly string[] \| undefined; exclude: readonly string[] \| undefined; }, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional record property access | 1 | 1 | 1 | 1 | 0 |
| overload 1 of arrayFrom result U[] cannot be served by implementation result (T \| U)[] | 1 | 1 | 1 | 1 | 0 |
| overload 1 of convertOptionsFromJson result WatchOptions \| undefined cannot be served by implementation result CompilerOptions \| TypeAcquisition \| WatchOptions \| undefined | 1 | 1 | 1 | 1 | 0 |
| overload 1 of convertOrRepopulateDiagnosticMessageChainArray parameter array cannot be served by implementation parameter array | 1 | 1 | 1 | 1 | 0 |
| overload 1 of createBuilderProgram result SemanticDiagnosticsBuilderProgram cannot be served by implementation result BuilderProgram \| undefined | 1 | 1 | 1 | 1 | 0 |
| overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined | 1 | 1 | 1 | 1 | 0 |
| overload 1 of getSupportedExtensionsWithJsonIfResolveJsonModule parameter supportedExtensions cannot be served by implementation parameter supportedExtensions | 1 | 1 | 1 | 1 | 0 |
| overload 1 of sameMap result U[] cannot be served by implementation result readonly U[] \| undefined | 1 | 1 | 1 | 1 | 0 |
| overload 1 of visitNode parameter node cannot be served by implementation parameter node | 1 | 1 | 1 | 1 | 0 |
| overload 2 of arrayToMap result Map<K, V2> cannot be served by implementation result Map<K, V1 \| V2> | 1 | 1 | 1 | 1 | 0 |
| overload 2 of getParseTreeNode result T \| undefined cannot be served by implementation result Node \| undefined | 1 | 1 | 1 | 1 | 0 |
| writing a checked object field that may hold undefined or null, whose checked write can't change the representation its slot holds | 6 | 6 | 6 | 6 | 0 |
| writing a checked object field without its source-slot type certificate | 1 | 1 | 1 | 1 | 0 |
| writing a possibly absent optional own field | 6 | 6 | 6 | 6 | 0 |

## Raw Findings

Moved by reason:


New findings:

None.

Units whose first stop moved:

None.

Panics: 0 before, 0 after.

## Distinct Reads

Moved by reason:


New findings:

None.

Units whose first stop moved:

None.

Panics: 0 before, 0 after.

The JSON retains every lowered site, move, new finding, unit stop and panic.
