# Paired latent census

| Measure | Raw findings | Distinct reads |
| --- | ---: | ---: |
| Lowered (no remaining finding) | 0 | 0 |
| Moved by reason | 0 | 0 |
| New findings | 1 | 1 |
| New sites | 1 | 1 |
| Units whose first stop moved | 0 | 0 |
| Regressions | 1 | 1 |
| before NotYet | 620 | 603 |
| before Refused | 4708 | 4512 |
| after NotYet | 620 | 603 |
| after Refused | 4709 | 4513 |

| Reason | Raw before | Raw after | Reads before | Reads after | Reads lowered |
| --- | ---: | ---: | ---: | ---: | ---: |
| ?. to a number, which would be number \| undefined | 1 | 1 | 1 | 1 | 0 |
| Array.isArray on a tuple or an erased object/any/unknown view | 1 | 1 | 1 | 1 | 0 |
| JSON.parse: its result's type can't be proven from the text | 1 | 1 | 1 | 1 | 0 |
| JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) | 2 | 2 | 2 | 2 | 0 |
| Object enumeration with a getter or accessor | 3 | 3 | 3 | 3 | 0 |
| Object enumeration with getter or accessor descriptors | 13 | 13 | 13 | 13 | 0 |
| Object.assign on a shape not proven by a plain literal or its const binding | 1 | 1 | 1 | 1 | 0 |
| Object.fromEntries | 1 | 1 | 1 | 1 | 0 |
| Object.getPrototypeOf | 1 | 1 | 1 | 1 | 0 |
| Reflect.get by a run-time key in a program with a class getter that does more than return its fields and literals joined by operators (the getter runs inside the read, unseen by the passes that follow calls) | 1 | 1 | 1 | 1 | 0 |
| RegExp with a nonconstant pattern | 4 | 4 | 4 | 4 | 0 |
| a () => never seen as a (node: ElementAccessExpression \| EnumMember \| PropertyAccessExpression) => string \| number \| undefined (one keeps something weakly that the other keeps strongly) | 1 | 1 | 1 | 1 | 0 |
| a () => true seen as a () => RepopulateDiagnosticChainInfo (one keeps something weakly that the other keeps strongly) | 1 | 1 | 1 | 1 | 0 |
| a (_?: unknown) => void seen as a (flags: LexicalEnvironmentFlags, value: boolean) => void (one keeps something weakly that the other keeps strongly) | 1 | 1 | 1 | 1 | 0 |
| a (current: Node) => boolean seen as a (element: Node) => "quit" \| boolean (one keeps something weakly that the other keeps strongly) | 1 | 1 | 1 | 1 | 0 |
| a (n: Node) => boolean seen as a (element: Node) => "quit" \| boolean (one keeps something weakly that the other keeps strongly) | 1 | 1 | 1 | 1 | 0 |
| a BaseParsedCommandLine seen as ParsedCommandLine (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a BinaryExpression with a string and a union of differently held members | 1 | 1 | 1 | 1 | 0 |
| a BinaryExpression with a union of differently held members and a union of differently held members | 1 | 1 | 1 | 1 | 0 |
| a BuildOptions seen as { extendedDiagnostics?: boolean; diagnostics?: boolean; } (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a ClassExpression | 1 | 1 | 1 | 1 | 0 |
| a CompilerOptions seen as CompilerOptions (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a CompilerOptions seen as MapLike<TsConfigSourceFile \| CompilerOptionsValue> (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a CompilerOptions seen as OptionsBase (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a CompilerOptions seen as OptionsBase \| undefined (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a CompilerOptions seen as SourceMapOptions (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a CompilerOptions seen as { extendedDiagnostics?: boolean; diagnostics?: boolean; } (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a Map of CompilerOptionsValue | 1 | 1 | 1 | 1 | 0 |
| a Map or Set going where a plain object type is held | 1 | 1 | 1 | 1 | 0 |
| a Map or Set interface implemented by a value that isn't a Map or Set | 50 | 50 | 50 | 50 | 0 |
| a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions | 2 | 2 | 2 | 2 | 0 |
| a PostfixUnaryExpression | 2 | 2 | 2 | 2 | 0 |
| a PrinterOptions seen as Pick<CompilerOptions, "module" \| "target"> (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a Record<string, CompilerOptionsValue> seen as CompilerOptions (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a SpreadElement | 1 | 1 | 1 | 1 | 0 |
| a WatchOptions seen as OptionsBase (fixed objects and records have different storage; copy explicitly) | 2 | 2 | 2 | 2 | 0 |
| a WatchOptions \| undefined seen as OptionsBase \| undefined (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a boolean \| undefined variable a function value captures | 3 | 3 | 3 | 3 | 0 |
| a branded supertype inside a checked predicate's array target | 2 | 2 | 2 | 2 | 0 |
| a call spreading elements that cannot be packed | 5 | 5 | 5 | 5 | 0 |
| a call spreading something other than an array | 1 | 1 | 1 | 1 | 0 |
| a call through ?. (an optional call) | 29 | 29 | 29 | 29 | 0 |
| a case whose type differs from the switch's | 6 | 6 | 6 | 6 | 0 |
| a cast the runtime can't check | 972 | 972 | 972 | 972 | 0 |
| a checked cast from never[] to [] whose value isn't a plain object | 1 | 1 | 1 | 1 | 0 |
| a checked predicate target DynamicNamedDeclaration that narrows GetAccessorDeclaration, which isn't its supertype | 1 | 1 | 1 | 1 | 0 |
| a checked predicate target NamedDeclaration that narrows ArrowFunction, which isn't its supertype | 1 | 1 | 1 | 1 | 0 |
| a checked predicate target TypeNode that narrows undefined, which isn't its supertype | 1 | 1 | 1 | 1 | 0 |
| a checked predicate target whose base has no required field to check (CompilerOptions & { types: string[]; }) | 1 | 1 | 1 | 1 | 0 |
| a checked predicate target whose base has no required field to check ({ [key: string]: unknown; }) | 1 | 1 | 1 | 1 | 0 |
| a checked predicate target whose check would read past the depth bound (BuilderProgramStateWithDefinedProgram) | 1 | 1 | 1 | 1 | 0 |
| a checked predicate target whose check would read past the depth bound (CircularBuildOrder) | 1 | 1 | 1 | 1 | 0 |
| a checked predicate target whose check would read past the depth bound (CommandLineOptionOfCustomType) | 1 | 1 | 1 | 1 | 0 |
| a checked predicate target whose check would read past the depth bound (IncrementalBundleEmitBuildInfo) | 1 | 1 | 1 | 1 | 0 |
| a checked view with a callable contract | 1 | 1 | 1 | 1 | 0 |
| a class used from a function that may run before the class's declaration (its temporal dead zone) | 4 | 4 | 4 | 4 | 0 |
| a comparator that doesn't take two elements and return a number | 1 | 1 | 1 | 1 | 0 |
| a computed field name | 3 | 3 | 3 | 3 | 0 |
| a conditional whose branches have different types | 1 | 1 | 1 | 1 | 0 |
| a destructured name held otherwise than its field | 2 | 2 | 2 | 2 | 0 |
| a destructured name that isn't plain | 1 | 1 | 1 | 1 | 0 |
| a dictionary member narrowed to an array without an element-checked view | 6 | 6 | 6 | 6 | 0 |
| a dynamic getter or method read, or a nullable field (runtime dispatch, slot tags and narrowed rereads) | 1 | 1 | 1 | 1 | 0 |
| a filter callback that doesn't return a boolean | 2 | 2 | 2 | 2 | 0 |
| a for...of destructuring an object | 1 | 1 | 1 | 1 | 0 |
| a function or method with a this parameter followed by another parameter (dynamic receivers are not implemented) | 7 | 7 | 7 | 7 | 0 |
| a function returning (AssignmentExpression<EqualsToken> & { readonly left: GeneratedIdentifier; }) \| undefined | 1 | 1 | 1 | 1 | 0 |
| a function returning (ConstructorDeclaration & { body: Block; }) \| undefined | 1 | 1 | 1 | 1 | 0 |
| a function returning AnyValidImportOrReExport | 1 | 1 | 1 | 1 | 0 |
| a function returning AnyValidImportOrReExport \| undefined | 1 | 1 | 1 | 1 | 0 |
| a function returning CanonicalKey | 1 | 1 | 1 | 1 | 0 |
| a function returning CapturedThis | 1 | 1 | 1 | 1 | 0 |
| a function returning ClassNamedEvaluationHelperBlock | 1 | 1 | 1 | 1 | 0 |
| a function returning ClassThisAssignmentBlock | 1 | 1 | 1 | 1 | 0 |
| a function returning CommandLineOption[] \| SortedReadonlyArray<CommandLineOption> | 1 | 1 | 1 | 1 | 0 |
| a function returning ExpressionWithTypeArguments & { readonly expression: Identifier \| PropertyAccessEntityNameExpression; } | 1 | 1 | 1 | 1 | 0 |
| a function returning HasJSDoc \| undefined | 1 | 1 | 1 | 1 | 0 |
| a function returning ImmediatelyInvokedArrowFunction | 1 | 1 | 1 | 1 | 0 |
| a function returning MemberName \| (Expression & (NumericLiteral \| StringLiteralLike)) | 1 | 1 | 1 | 1 | 0 |
| a function returning ResolvedConfigFileName | 2 | 2 | 2 | 2 | 0 |
| a function returning SortedReadonlyArray<Diagnostic> | 1 | 1 | 1 | 1 | 0 |
| a function returning SortedReadonlyArray<EmitHelper> \| undefined | 1 | 1 | 1 | 1 | 0 |
| a function returning SortedReadonlyArray<T> | 1 | 1 | 1 | 1 | 0 |
| a function returning SortedReadonlyArray<string> | 1 | 1 | 1 | 1 | 0 |
| a function returning T | 1 | 1 | 1 | 1 | 0 |
| a function returning T \| undefined | 3 | 3 | 3 | 3 | 0 |
| a function returning U \| undefined | 1 | 1 | 1 | 1 | 0 |
| a function returning any | 1 | 1 | 1 | 1 | 0 |
| a function taking ((node: Node) => VisitResult<Node>) \| undefined seen as one taking ((node: Node) => VisitResult<Node \| undefined>) \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 0 |
| a function taking () => T seen as one taking () => T (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 | 3 | 3 | 3 | 0 |
| a function taking (resolvedProjectReference: ResolvedProjectReference) => T \| undefined seen as one taking (resolvedProjectReference: ResolvedProjectReference) => T \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 0 |
| a function taking BinaryExpressionStateMachine<TOuterState, TState, TResult> seen as one taking BinaryExpressionStateMachine<TOuterState, TState, TResult> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 6 | 6 | 6 | 6 | 0 |
| a function taking BinaryOperator seen as one taking SyntaxKind (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 0 |
| a function taking BinaryOperatorToken seen as one taking BinaryOperatorToken \| BinaryOperator (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking CancellationToken seen as one taking CancellationToken \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking Expression seen as one taking Expression \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking Expression[] seen as one taking readonly Expression[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 0 |
| a function taking Identifier seen as one taking Node (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking IndexSignatureDeclaration \| undefined seen as one taking SignatureDeclaration \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking JSDocTypeExpression \| undefined seen as one taking JSDocTypeExpression \| JSDocTypeLiteral \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 0 |
| a function taking Node seen as one taking Node \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 0 |
| a function taking Node seen as one taking [node: Node] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 8 | 8 | 8 | 8 | 0 |
| a function taking NodeArray<Expression> seen as one taking readonly Expression[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking NodeArray<T> seen as one taking NodeArray<T> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking NodeArray<T> seen as one taking NodeArray<T> \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking NodeArray<TypeNode> \| undefined seen as one taking readonly TypeNode[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking Path seen as one taking [filePath: Path] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking Path seen as one taking any[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking Path seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 0 |
| a function taking PollingInterval seen as one taking number \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking SourceFile seen as one taking SourceFile \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 | 3 | 3 | 3 | 0 |
| a function taking SourceFile seen as one taking [file: SourceFile, options: CompilerOptions] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking SymbolFlags seen as one taking SymbolFlags \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking SyntaxKind.MethodSignature \| SyntaxKind.MethodDeclaration \| SyntaxKind.Constructor \| SyntaxKind.GetAccessor \| SyntaxKind.SetAccessor \| SyntaxKind.CallSignature \| SyntaxKind.ConstructSignature \| ... 6 more ... \| SyntaxKind.JSDocFunctionType seen as one taking SyntaxKind (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking TypeFacts.None seen as one taking number (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking TypeNode seen as one taking Node (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking Visitor seen as one taking Visitor<TIn, Node \| undefined> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 0 |
| a function taking [callback: (...args: any[]) => void, ms: number, ...args: any[]] seen as one taking (...args: any[]) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking [data: string] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 5 | 5 | 5 | 5 | 0 |
| a function taking [fileName: string, languageVersionOrOptions: CreateSourceFileOptions \| ScriptTarget, onError?: ((message: string) => void) \| undefined, shouldCreateNewSourceFile?: boolean \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 0 |
| a function taking [fileName: string, text: string, writeByteOrderMark: boolean, onError?: ((message: string) => void) \| undefined, sourceFiles?: readonly SourceFile[] \| undefined, data?: WriteFileCallbackData \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 | 3 | 3 | 3 | 0 |
| a function taking [fileName: string] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking [filePath: Path] seen as one taking Path (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking [languageVersionOrOptions: CreateSourceFileOptions \| ScriptTarget, onError?: ((message: string) => void) \| undefined, shouldCreateNewSourceFile?: boolean \| undefined] seen as one taking CreateSourceFileOptions \| ScriptTarget (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking [libFileName: string] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking [libraryName: string, resolveFrom: string, options: CompilerOptions, libFileName: string] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking [moduleLiterals: readonly StringLiteralLike[], containingFile: string, redirectedReference: ResolvedProjectReference \| undefined, options: CompilerOptions, containingSourceFile: SourceFile, reusedNames: ... \| undefined] seen as one taking readonly StringLiteralLike[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 0 |
| a function taking [moduleNames: string[], containingFile: string, reusedNames: string[] \| undefined, redirectedReference: ResolvedProjectReference \| undefined, options: CompilerOptions, containingSourceFile?: SourceFile \| undefined] seen as one taking string[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 0 |
| a function taking [name: string] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 0 |
| a function taking [node: ConstructorTypeNode, typeParameters: NodeArray<TypeParameterDeclaration> \| undefined, parameters: NodeArray<ParameterDeclaration>, type: TypeNode] \| ... seen as one taking ConstructorTypeNode (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking [node: Node] seen as one taking BindingElement \| OmittedExpression (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking [node: Node] seen as one taking Declaration (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 0 |
| a function taking [path: string, callback: DirectoryWatcherCallback, recursive?: boolean \| undefined, options?: WatchOptions \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking [path: string, callback: FileWatcherCallback, pollingInterval?: number \| undefined, options?: WatchOptions \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking [path: string, extensions?: readonly string[] \| undefined, exclude?: readonly string[] \| undefined, include?: readonly string[] \| undefined, depth?: number \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking [path: string] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 8 | 8 | 8 | 8 | 0 |
| a function taking [s: string] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 | 3 | 3 | 3 | 0 |
| a function taking [typeDirectiveReferences: readonly (string \| FileReference)[], containingFile: string, redirectedReference: ResolvedProjectReference \| undefined, options: CompilerOptions, containingSourceFile: SourceFile \| undefined, reusedNames: ... \| undefined] seen as one taking readonly T[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 0 |
| a function taking [typeParameters: readonly TypeParameterDeclaration[] \| undefined, parameters: readonly ParameterDeclaration[], type: TypeNode] \| [modifiers: readonly Modifier[] \| undefined, typeParameters: ... \| undefined, parameters: ..., type: TypeNode] seen as one taking readonly Modifier[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking [typeReferenceDirectiveNames: string[] \| readonly FileReference[], containingFile: string, redirectedReference: ResolvedProjectReference \| undefined, options: CompilerOptions, containingFileMode?: ResolutionMode] seen as one taking string[] \| readonly FileReference[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 0 |
| a function taking [writeByteOrderMark: boolean, onError?: ((message: string) => void) \| undefined, sourceFiles?: readonly SourceFile[] \| undefined, data?: WriteFileCallbackData \| undefined] seen as one taking boolean (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking boolean seen as one taking boolean \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 | 3 | 3 | 3 | 0 |
| a function taking readonly (string \| FileReference)[] seen as one taking [typeDirectiveReferences: readonly (string \| FileReference)[], containingFile: string, redirectedReference: ResolvedProjectReference \| undefined, options: CompilerOptions, containingSourceFile: SourceFile \| undefined, reusedNames: ... \| undefined] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking readonly Modifier[] \| undefined seen as one taking readonly ModifierLike[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 4 | 4 | 4 | 4 | 0 |
| a function taking readonly ParameterDeclaration[] seen as one taking readonly ParameterDeclaration[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking readonly StringLiteralLike[] seen as one taking [moduleLiterals: readonly StringLiteralLike[], containingFile: string, redirectedReference: ResolvedProjectReference \| undefined, options: CompilerOptions, containingSourceFile: SourceFile, reusedNames: ... \| undefined] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking readonly T[] seen as one taking readonly T[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking readonly T[] \| undefined seen as one taking readonly T[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking readonly string[] seen as one taking readonly string[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking string seen as one taking [data: string] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 0 |
| a function taking string seen as one taking [fileName: string, text: string, writeByteOrderMark: boolean, onError?: ((message: string) => void) \| undefined, sourceFiles?: readonly SourceFile[] \| undefined, data?: WriteFileCallbackData \| undefined] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 | 3 | 3 | 3 | 0 |
| a function taking string seen as one taking [fileName: string] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 0 |
| a function taking string seen as one taking [libFileName: string] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking string seen as one taking [name: string] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking string seen as one taking [path: string] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 7 | 7 | 7 | 7 | 0 |
| a function taking string seen as one taking [s: string] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking string seen as one taking any[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 0 |
| a function taking string seen as one taking string \| MemberName (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking string seen as one taking string \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 0 |
| a function taking string[] seen as one taking [moduleNames: string[], containingFile: string, reusedNames: string[] \| undefined, redirectedReference: ResolvedProjectReference \| undefined, options: CompilerOptions, containingSourceFile?: SourceFile \| undefined] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking string[] \| readonly FileReference[] seen as one taking [typeReferenceDirectiveNames: string[] \| readonly FileReference[], containingFile: string, redirectedReference: ResolvedProjectReference \| undefined, options: CompilerOptions, containingFileMode?: ResolutionMode] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function value that captures the variable its own initializer declares | 1 | 1 | 1 | 1 | 0 |
| a function without a body | 2 | 2 | 2 | 2 | 0 |
| a generator function declared inside a function or a block (#qk8rztp) | 1 | 1 | 1 | 1 | 0 |
| a generic function value without a concrete contextual signature | 1 | 1 | 1 | 1 | 0 |
| a generic instantiated at void whose body reads the value: getPropFromRaw's T = void (read from the signature), read at src/compiler/commandLineParser.ts:3253:24 as a returned value | 1 | 1 | 1 | 1 | 0 |
| a generic instantiation whose type argument for T, constrained to Node, can't be read | 7 | 7 | 7 | 7 | 0 |
| a generic instantiation whose type argument for TIn, constrained to Node \| undefined, can't be read | 17 | 17 | 17 | 17 | 0 |
| a generic instantiation whose type argument for TIn, constrained to Node, can't be read | 9 | 9 | 9 | 9 | 0 |
| a library method value outside a const alias, typed call/apply, or supported map callback (its receiver and callable ABI are not proven); wrap the call in an arrow | 1 | 1 | 1 | 1 | 0 |
| a literal method through a view that erases its receiver | 5 | 5 | 5 | 5 | 0 |
| a member typed any (()) related to T \| undefined at the initializer site | 1 | 1 | 1 | 1 | 0 |
| a member typed any ((0)) related to Expression at the assignment site | 1 | 1 | 1 | 1 | 0 |
| a member typed any ((0)) related to string at the argument site | 1 | 1 | 1 | 1 | 0 |
| a member typed any ((0)) related to string at the relation site | 1 | 1 | 1 | 1 | 0 |
| a member typed any (167(0)) related to QualifiedName at the assertion site | 2 | 2 | 2 | 2 | 0 |
| a member typed any ([]) related to number at the initializer site | 1 | 1 | 1 | 1 | 0 |
| a member typed any ([]) related to string at the return site | 1 | 1 | 1 | 1 | 0 |
| a member typed any (config, declared): any has no proven type, so every read of it through a typed view is unchecked | 0 | 1 | 0 | 1 | 0 |
| a member typed any (getTokenConstructor()) related to new <TKind extends SyntaxKind>(kind: TKind, pos: number, end: number) => Token<TKind> at the initializer site | 1 | 1 | 1 | 1 | 0 |
| a member typed any (keys()next()<1>) related to undefined at the assignment site | 1 | 1 | 1 | 1 | 0 |
| a member typed any (set(0)) related to number \| undefined at the object field site | 1 | 1 | 1 | 1 | 0 |
| a method in object destructuring | 5 | 5 | 5 | 5 | 0 |
| a method read as a value (add would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (afterProgramCreate would lose its object, and this with it) | 3 | 3 | 3 | 3 | 0 |
| a method read as a value (afterProgramEmitAndDiagnostics would lose its object, and this with it) | 3 | 3 | 3 | 3 | 0 |
| a method read as a value (base64decode would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (base64encode would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (clearMarks would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (clearMeasures would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (clearScreen would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (clearTimeout would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (cloneNode would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (compare would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (convertToArrayAssignmentElement would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (convertToObjectAssignmentElement would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (createComma would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (createDirectory would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (createExpressionStatement would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (createHash would lose its object, and this with it) | 3 | 3 | 3 | 3 | 0 |
| a method read as a value (createIntersectionTypeNode would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (createJSDocClassTag would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (createJSDocDeprecatedTag would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (createJSDocLink would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (createJSDocLinkCode would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (createJSDocLinkPlain would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (createJSDocOverrideTag would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (createJSDocPrivateTag would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (createJSDocProtectedTag would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (createJSDocPublicTag would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (createJSDocReadonlyTag would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (createPropertyDeclaration would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (createTemplateMiddle would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (createTemplateTail would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (createUnionTypeNode would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (deleteFile would lose its object, and this with it) | 4 | 4 | 4 | 4 | 0 |
| a method read as a value (directoryExists would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (emit would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (emitBuildInfo would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (emitNextAffectedFile would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (emitNodeWithNotification would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (enableCPUProfiler would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (fileExists would lose its object, and this with it) | 10 | 10 | 10 | 10 | 0 |
| a method read as a value (fileIsOpen would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (fill would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (fromCodePoint would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getAllDependencies would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getBuildInfo would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (getCachedDirectoryStructureHost would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getCanonicalFileName would lose its object, and this with it) | 9 | 9 | 9 | 9 | 0 |
| a method read as a value (getCommonSourceDirectory would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (getCompilationSettings would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getCompilerOptions would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getCurrentDirectory would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (getCurrentProgram would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getDeclarationDiagnostics would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getDefaultLibLocation would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getEnvironmentVariable would lose its object, and this with it) | 4 | 4 | 4 | 4 | 0 |
| a method read as a value (getFileIncludeReasons would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getGlobalTypingsCacheLocation would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (getMemoryUsage would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getModifiedTime would lose its object, and this with it) | 4 | 4 | 4 | 4 | 0 |
| a method read as a value (getModuleResolutionCache would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getNewLine would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getOrCreateCacheForModuleName would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getParsedCommandLine would lose its object, and this with it) | 6 | 6 | 6 | 6 | 0 |
| a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getSemanticDiagnostics would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getSemanticDiagnosticsOfNextAffectedFile would lose its object, and this with it) | 3 | 3 | 3 | 3 | 0 |
| a method read as a value (getSourceFile would lose its object, and this with it) | 8 | 8 | 8 | 8 | 0 |
| a method read as a value (getSourceFileByPath would lose its object, and this with it) | 3 | 3 | 3 | 3 | 0 |
| a method read as a value (getSourceFiles would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getSymlinkCache would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (globalCacheResolutionModuleName would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (hasChangedEmitSignature would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (hasGlobalName would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (hasInvalidatedLibResolutions would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (hasInvalidatedResolutions would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (isEmitNotificationEnabled would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (liftToBlock would lose its object, and this with it) | 20 | 20 | 20 | 20 | 0 |
| a method read as a value (log would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (mark would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (measure would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (nonEscapingWrite would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (now would lose its object, and this with it) | 5 | 5 | 5 | 5 | 0 |
| a method read as a value (onChangedAutomaticTypeDirectiveNames would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (onDiscoveredSymlink would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (onInvalidatedResolution would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (onReleaseOldSourceFile would lose its object, and this with it) | 3 | 3 | 3 | 3 | 0 |
| a method read as a value (onReleaseParsedCommandLine would lose its object, and this with it) | 3 | 3 | 3 | 3 | 0 |
| a method read as a value (onWatchStatusChange would lose its object, and this with it) | 3 | 3 | 3 | 3 | 0 |
| a method read as a value (parenthesizeBranchOfConditionalExpression would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (parenthesizeCheckTypeOfConditionalType would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (parenthesizeConciseBodyOfArrowFunction would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (parenthesizeConditionOfConditionalExpression would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (parenthesizeConstituentTypeOfIntersectionType would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (parenthesizeConstituentTypeOfUnionType would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (parenthesizeConstituentTypesOfIntersectionType would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (parenthesizeConstituentTypesOfUnionType would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (parenthesizeElementTypeOfTupleType would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (parenthesizeExpressionForDisallowedComma would lose its object, and this with it) | 12 | 12 | 12 | 12 | 0 |
| a method read as a value (parenthesizeExpressionOfComputedPropertyName would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (parenthesizeExpressionOfExportDefault would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (parenthesizeExpressionOfExpressionStatement would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (parenthesizeExpressionOfNew would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (parenthesizeExtendsTypeOfConditionalType would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (parenthesizeLeadingTypeArgument would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (parenthesizeLeftSideOfAccess would lose its object, and this with it) | 7 | 7 | 7 | 7 | 0 |
| a method read as a value (parenthesizeNonArrayTypeOfPostfixType would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (parenthesizeOperandOfPostfixUnary would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (parenthesizeOperandOfPrefixUnary would lose its object, and this with it) | 6 | 6 | 6 | 6 | 0 |
| a method read as a value (parenthesizeOperandOfReadonlyTypeOperator would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (parenthesizeOperandOfTypeOperator would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (parenthesizeTypeOfOptionalType would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (readDirectory would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (readFile would lose its object, and this with it) | 5 | 5 | 5 | 5 | 0 |
| a method read as a value (realpath would lose its object, and this with it) | 3 | 3 | 3 | 3 | 0 |
| a method read as a value (releaseProgram would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (remove would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (repeat would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (reportCyclicStructureError would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (reportInaccessibleThisError would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (reportInaccessibleUniqueSymbolError would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (reportInferenceFallback would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (reportLikelyUnsafeImportRequiredError would lose its object, and this with it) | 3 | 3 | 3 | 3 | 0 |
| a method read as a value (reportNonSerializableProperty would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (reportNonlocalAugmentation would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (reportPrivateInBaseOfClassExpression would lose its object, and this with it) | 3 | 3 | 3 | 3 | 0 |
| a method read as a value (reportTruncationError would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (resolveLibrary would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (resolveModuleNameLiterals would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (resolveModuleNames would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (resolveTypeReferenceDirectiveReferences would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (resolveTypeReferenceDirectives would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (scheduleInvalidateResolutionsOfFailedLookupLocations would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (setModifiedTime would lose its object, and this with it) | 4 | 4 | 4 | 4 | 0 |
| a method read as a value (setPrototypeOf would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (setTimeout would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (substituteNode would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (throwIfCancellationRequested would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (toKey would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (toPath would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (toString would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (trace would lose its object, and this with it) | 4 | 4 | 4 | 4 | 0 |
| a method read as a value (trackSymbol would lose its object, and this with it) | 4 | 4 | 4 | 4 | 0 |
| a method read as a value (useCaseSensitiveFileNames would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (watchAffectingFileLocation would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (watchDirectory would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (watchDirectoryOfFailedLookupLocation would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (watchFile would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (watchTypeRootsDirectory would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (writeFile would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (writeLog would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (writeOutputIsTTY would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method sharing its name with a getter; rename the getter or use a callable arrow field in the interface and implementation | 5 | 5 | 5 | 5 | 0 |
| a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports | 2 | 2 | 2 | 2 | 0 |
| a non-object representation passed through an opaque object parameter | 32 | 32 | 32 | 32 | 0 |
| a parameter that isn't a plain name | 1 | 1 | 1 | 1 | 0 |
| a primitive brand member __canonicalKey whose type is not void | 1 | 1 | 1 | 1 | 0 |
| a primitive brand member _isResolvedConfigFileName whose type is not void | 2 | 2 | 2 | 2 | 0 |
| a read of a void value, as a returned value | 3 | 3 | 3 | 3 | 0 |
| a rest parameter that isn't an array | 1 | 1 | 1 | 1 | 0 |
| a spread argument T's type argument can be inferred from, judged against its constraint unknown[] | 1 | 1 | 1 | 1 | 0 |
| a structural relation deeper than 64 type pairs while checking void function results | 25 | 25 | 25 | 25 | 0 |
| a tagged template other than the intrinsic String.raw | 2 | 2 | 2 | 2 | 0 |
| a template interpolating a union with an object, an array, a map or a function in it | 1 | 1 | 1 | 1 | 0 |
| a template interpolating an object, an array, a map, a function or undefined | 2 | 2 | 2 | 2 | 0 |
| a tuple literal leaving out an element of type ModuleSpecifierEnding | 1 | 1 | 1 | 1 | 0 |
| a tuple literal with more values than its tuple has elements | 1 | 1 | 1 | 1 | 0 |
| a type predicate whose return is not proven (there is no body proving this parameter) | 5 | 5 | 5 | 5 | 0 |
| a union of differently held members variable a function value captures | 4 | 4 | 4 | 4 | 0 |
| a union or optional field read in a program with record storage | 2 | 2 | 2 | 2 | 0 |
| a union or optional tuple length; declare arrays where the values are constructed, or carry an explicit length alongside each tuple | 1 | 1 | 1 | 1 | 0 |
| a value of type ((node: PrivateIdentifierPropertyDeclaration, modifiers: ModifiersArray \| undefined) => ObjectLiteralExpression) \| undefined seen as ((node: PropertyDeclaration & { readonly name: PrivateIdentifier; }, modifiers: ModifiersArray \| undefined) => Expression) \| undefined, whose readonly field name becomes writable: a readonly field may hold something narrower than PrivateIdentifier, which a write of PrivateIdentifier would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type () => { diagnosticMessage: DiagnosticMessage; errorNode: ExportAssignment; } seen as GetSymbolAccessibilityDiagnostic, which can write Node where ExportAssignment is read | 1 | 1 | 1 | 1 | 0 |
| a value of type () => { diagnosticMessage: DiagnosticMessage; errorNode: ExpressionWithTypeArguments; typeName: Identifier \| undefined; } seen as GetSymbolAccessibilityDiagnostic, which can write Node where ExpressionWithTypeArguments is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; } & BinaryExpression) \| (... & ... 1 more ... & BinaryExpression) | 1 | 1 | 1 | 1 | 0 |
| a value of type (ClassDeclaration \| EnumDeclaration \| ExportAssignment \| ExportDeclaration \| FunctionDeclaration \| ... 5 more ... \| VariableStatement)[] seen as Statement[], which can write Statement where ClassDeclaration \| EnumDeclaration \| ExportAssignment \| ExportDeclaration \| FunctionDeclaration \| ... 5 more ... \| VariableStatement is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (ConstructorDeclaration & { body: Block; }) \| undefined | 1 | 1 | 1 | 1 | 0 |
| a value of type (EmitNode & { autoGenerate: AutoGenerateInfo; }) \| (EmitNode & { autoGenerate: AutoGenerateInfo; }) | 2 | 2 | 2 | 2 | 0 |
| a value of type (IndexSignatureDeclaration \| PropertyDeclaration)[] \| undefined seen as ClassElement[] \| undefined, which can write ClassElement where IndexSignatureDeclaration \| PropertyDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (JsxElement \| JsxExpression \| JsxFragment \| JsxSelfClosingElement \| JsxText)[] seen as Node[], which can write Node where JsxElement \| JsxExpression \| JsxFragment \| JsxSelfClosingElement \| JsxText is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (ModuleDeclaration & { name: StringLiteral; }) \| undefined | 1 | 1 | 1 | 1 | 0 |
| a value of type (TextRange & { name: string; })[] \| undefined seen as {}[] \| undefined, which can write {} where TextRange & { name: string; } is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (TextRange & { value: number; })[] \| undefined seen as {}[] \| undefined, which can write {} where TextRange & { value: number; } is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (VariableDeclaration & { name: Identifier; }) \| undefined | 1 | 1 | 1 | 1 | 0 |
| a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (baseDir: string, moduleName: string) => { module: any; modulePath: string; error: undefined; } \| { module: undefined; modulePath: undefined; error: unknown; } seen as (baseDir: string, moduleName: string) => ModuleImportResult, which can write string \| undefined where string is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (chain: DiagnosticMessageChain \| ReusableDiagnosticMessageChain) => RepopulateDiagnosticChainInfo \| undefined seen as (chain: DiagnosticMessageChain) => RepopulateDiagnosticChainInfo \| undefined, which can write ReusableDiagnosticMessageChain[] \| undefined where DiagnosticMessageChain[] \| undefined is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (chain: ReusableDiagnosticMessageChain) => RepopulateDiagnosticChainInfo seen as (chain: DiagnosticMessageChain \| ReusableDiagnosticMessageChain) => RepopulateDiagnosticChainInfo \| undefined, which can write ReusableDiagnosticMessageChain[] \| undefined where DiagnosticMessageChain[] \| undefined is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (identifierOrPrivateName: Identifier \| PrivateIdentifier) => string seen as (name: GeneratedIdentifier \| GeneratedPrivateIdentifier) => string, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type (left: MappedPosition, right: MappedPosition) => boolean seen as EqualityComparer<SourceMappedPosition> \| undefined, which can write string \| undefined where string is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (node: CommentRange) => boolean seen as (value: SynthesizedComment) => boolean, which can write number where -1 is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (node: Node) => boolean seen as AnyFunction, which can write void where boolean is read | 4 | 4 | 4 | 4 | 0 |
| a value of type (node: PrivateIdentifierGetAccessorDeclaration, modifiers: ModifiersArray \| undefined) => ObjectLiteralExpression seen as ((node: GetAccessorDeclaration & { readonly name: PrivateIdentifier; }, modifiers: ModifiersArray \| undefined) => Expression) \| undefined, whose readonly field name becomes writable: a readonly field may hold something narrower than PrivateIdentifier, which a write of PrivateIdentifier would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type (node: PrivateIdentifierMethodDeclaration, modifiers: ModifiersArray \| undefined) => ObjectLiteralExpression seen as ((node: MethodDeclaration & { readonly name: PrivateIdentifier; }, modifiers: ModifiersArray \| undefined) => Expression) \| undefined, whose readonly field name becomes writable: a readonly field may hold something narrower than PrivateIdentifier, which a write of PrivateIdentifier would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type (node: PrivateIdentifierSetAccessorDeclaration, modifiers: ModifiersArray \| undefined) => ObjectLiteralExpression seen as ((node: SetAccessorDeclaration & { readonly name: PrivateIdentifier; }, modifiers: ModifiersArray \| undefined) => Expression) \| undefined, whose readonly field name becomes writable: a readonly field may hold something narrower than PrivateIdentifier, which a write of PrivateIdentifier would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type (readonly IncrementalBuildInfoFileId[])[] \| undefined seen as {}[] \| undefined, which can write {} where readonly IncrementalBuildInfoFileId[] is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (readonly [() => IntrinsicType, __String])[] seen as (readonly [() => Type, __String])[], which can write readonly [() => Type, __String] where readonly [() => IntrinsicType, __String] is read | 2 | 2 | 2 | 2 | 0 |
| a value of type (recordTempVariable: ((node: Identifier) => void) \| undefined, reservedInNestedScopes?: boolean \| undefined, prefix?: string \| GeneratedNamePart \| undefined, suffix?: string \| undefined) => GeneratedIdentifier seen as { (recordTempVariable: ((node: Identifier) => void) \| undefined, reservedInNestedScopes?: boolean \| undefined): Identifier; (recordTempVariable: ((node: Identifier) => void) \| undefined, reservedInNestedScopes?: boolean \| undefined, prefix?: string \| ... 1 more ... \| undefined, suffix?: string \| undefined): Identifi..., whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type (resolution: ResolvedModuleWithFailedLookupLocations) => ResolvedModuleFull \| undefined seen as (oldResolution: ResolvedModuleWithFailedLookupLocations) => ResolutionWithResolvedFileName \| undefined, which can write string \| undefined where string is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (resolution: ResolvedModuleWithFailedLookupLocations) => ResolvedModuleFull \| undefined seen as GetResolutionWithResolvedFileName<CachedResolvedModuleWithFailedLookupLocations, ResolutionWithResolvedFileName>, which can write string \| undefined where string is read | 3 | 3 | 3 | 3 | 0 |
| a value of type (sourceFile: SourceFile \| undefined, cancellationToken: CancellationToken \| undefined) => readonly DiagnosticWithLocation[] seen as (sourceFile?: SourceFile \| undefined, cancellationToken?: CancellationToken \| undefined) => readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (sourceFile: SourceFile) => readonly DiagnosticWithLocation[] seen as (sourceFile: SourceFile, cancellationToken: CancellationToken \| undefined) => readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (sourceFile: SourceFile, cancellationToken: CancellationToken \| undefined) => readonly DiagnosticWithLocation[] seen as (sourceFile: SourceFile, cancellationToken: CancellationToken \| undefined) => readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (symbol: Symbol) => { visitedTypes: Type[]; visitedSymbols: Symbol[]; } seen as (root: Symbol) => { visitedTypes: readonly Type[]; visitedSymbols: readonly Symbol[]; }, which can write readonly Type[] where Type[] is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (symbolAccessibilityResult: SymbolAccessibilityResult) => { diagnosticMessage: DiagnosticMessage; errorNode: DeclarationDiagnosticProducing; typeName: DeclarationName \| undefined; } \| undefined seen as (symbolAccessibilityResult: SymbolAccessibilityResult) => SymbolAccessibilityDiagnostic \| undefined, which can write Node where DeclarationDiagnosticProducing is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (type: Type) => { visitedTypes: Type[]; visitedSymbols: Symbol[]; } seen as (root: Type) => { visitedTypes: readonly Type[]; visitedSymbols: readonly Symbol[]; }, which can write readonly Type[] where Type[] is read | 1 | 1 | 1 | 1 | 0 |
| a value of type AccessorDeclaration & { readonly name: BigIntLiteral \| ComputedPropertyName \| Identifier \| NoSubstitutionTemplateLiteral \| NumericLiteral \| StringLiteral; } | 1 | 1 | 1 | 1 | 0 |
| a value of type AccessorDeclaration[] seen as T \| (T \| AccessorDeclaration)[] \| AccessorDeclaration, which can write T \| AccessorDeclaration where AccessorDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type AnonymousType seen as AnonymousType, which can write AnonymousType \| undefined where GenericType is read | 1 | 1 | 1 | 1 | 0 |
| a value of type AnonymousType \| DeferredTypeReference seen as AnonymousType, which can write AnonymousType \| undefined where GenericType is read | 1 | 1 | 1 | 1 | 0 |
| a value of type AnonymousType \| GenericType seen as AnonymousType, which can write AnonymousType \| undefined where GenericType is read | 1 | 1 | 1 | 1 | 0 |
| a value of type ArrayLiteralExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type ArrayLiteralExpression \| AssignmentExpression<EqualsToken> \| BindingElement \| ElementAccessExpression \| ... 8 more ... \| VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type AssertClause seen as Mutable<AssertClause>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type AssertClause seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type AssertEntry seen as Mutable<AssertEntry>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type AssertEntry seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type AssignmentExpression<EqualsToken> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 25 | 25 | 25 | 25 | 0 |
| a value of type AssignmentPattern \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type AutoAccessorPropertyDeclaration \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type AwaitedTypeInstantiation seen as Type \| undefined, which can write Symbol \| undefined where Symbol is read | 1 | 1 | 1 | 1 | 0 |
| a value of type AwaitedTypeInstantiation \| Type seen as Type, which can write Symbol \| undefined where Symbol is read | 1 | 1 | 1 | 1 | 0 |
| a value of type BigIntLiteral \| ComputedPropertyName \| Identifier \| JsxNamespacedName \| NoSubstitutionTemplateLiteral \| NumericLiteral \| PrivateIdentifier \| StringLiteral seen as Type, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type BigIntLiteral \| ComputedPropertyName \| Identifier \| JsxNamespacedName \| NoSubstitutionTemplateLiteral \| NumericLiteral \| PrivateIdentifier \| StringLiteral \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type BigIntLiteral \| ComputedPropertyName \| Identifier \| NoSubstitutionTemplateLiteral \| NumericLiteral \| PrivateIdentifier \| StringLiteral \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type BigIntLiteral \| NoSubstitutionTemplateLiteral \| NumericLiteral \| StringLiteral seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type BinaryExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type BindableObjectDefinePropertyCall | 1 | 1 | 1 | 1 | 0 |
| a value of type BindingElement & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | 1 | 1 | 0 |
| a value of type BindingElement[] seen as unknown[], which can write unknown where BindingElement is read | 2 | 2 | 2 | 2 | 0 |
| a value of type BindingElement[] \| undefined seen as {}[] \| undefined, which can write {} where BindingElement is read | 1 | 1 | 1 | 1 | 0 |
| a value of type BindingName seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type BindingName seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type BindingOrAssignmentElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type BindingOrAssignmentElement[] \| undefined seen as {}[] \| undefined, which can write {} where BindingOrAssignmentElement is read | 3 | 3 | 3 | 3 | 0 |
| a value of type Block seen as Mutable<Block>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Block seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Block seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type Block seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 16 | 16 | 16 | 16 | 0 |
| a value of type Block \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type BreakStatement seen as Mutable<BreakStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type BreakStatement seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type BreakStatement seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type BuildOptions seen as BuildOptions \| CompilerOptions, which can write TsConfigSourceFile \| CompilerOptionsValue where CompilerOptionsValue is read | 5 | 5 | 5 | 5 | 0 |
| a value of type BuilderProgramState seen as ReusableBuilderProgramState, which can write Map<Path, readonly Diagnostic[] \| readonly ReusableDiagnostic[]> where Map<Path, readonly Diagnostic[]> is read | 9 | 9 | 9 | 9 | 0 |
| a value of type BuilderProgramStateWithDefinedProgram seen as BuilderProgramState, which can write Program \| undefined where Program is read | 12 | 12 | 12 | 12 | 0 |
| a value of type CachedDirectoryStructureHost \| WatchCompilerHostOfFilesAndCompilerOptionsOrConfigFile<T> seen as DirectoryStructureHost, which can write ((path: string, extensions?: readonly string[] \| undefined, exclude?: readonly string[] \| undefined, include?: readonly string[] \| undefined, depth?: number \| undefined) => string[]) \| undefined where ((path: string, extensions?: readonly string[] \| undefined, exclude?: readonly string[] \| undefined, include?: readonly string[] \| undefined, depth?: number \| undefined) => string[]) & ((path: string, extensions?: ... \| undefined, exclude?: ... \| undefined, include?: ... \| undefined, depth?: number \| undefined) => .... is read | 1 | 1 | 1 | 1 | 0 |
| a value of type CallExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 24 | 24 | 24 | 24 | 0 |
| a value of type CallExpression \| undefined seen as NodeArray<Expression> \| undefined, whose readonly field transformFlags becomes writable: a readonly field may hold something narrower than TransformFlags, which a write of TransformFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type CallExpression[] \| undefined seen as {}[] \| undefined, which can write {} where CallExpression is read | 2 | 2 | 2 | 2 | 0 |
| a value of type CallLikeExpression \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type CallSignatureDeclaration seen as Mutable<CallSignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type CallSignatureDeclaration seen as Mutable<SignatureDeclarationBase>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type CapturedThis seen as Expression \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type CapturedThis seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type CapturedThis seen as PrimaryExpression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type CapturedThis seen as string \| BindingName, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type CaseBlock seen as Mutable<CaseBlock>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type CaseBlock seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Children seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ClassDeclaration seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ClassDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 4 | 4 | 4 | 4 | 0 |
| a value of type ClassDeclaration \| ClassExpression \| InferTypeNode \| InterfaceDeclaration \| JSDocCallbackTag \| ... 4 more ... \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ClassElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type ClassElement[] \| undefined seen as {}[] \| undefined, which can write {} where ClassElement is read | 5 | 5 | 5 | 5 | 0 |
| a value of type ClassExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type ClassStaticBlockDeclaration seen as Mutable<ClassStaticBlockDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type ClassStaticBlockDeclaration \| PropertyDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type CommentDirective[] \| undefined seen as {}[] \| undefined, which can write {} where CommentDirective is read | 3 | 3 | 3 | 3 | 0 |
| a value of type CompilerHost & ResolutionCacheHost seen as CompilerHostLikeForCache, which can write WriteFileCallback \| undefined where WriteFileCallback is read | 1 | 1 | 1 | 1 | 0 |
| a value of type CompilerHost & ResolutionCacheHost seen as ResolutionCacheHost, which can write boolean \| (() => boolean) \| undefined where (() => boolean) & (boolean \| (() => boolean) \| undefined) is read | 1 | 1 | 1 | 1 | 0 |
| a value of type CompilerHost seen as CompilerHostLikeForCache, which can write WriteFileCallback \| undefined where WriteFileCallback is read | 2 | 2 | 2 | 2 | 0 |
| a value of type CompilerOptions & { types: string[]; } seen as CompilerOptions, which can write string[] \| undefined where string[] is read | 2 | 2 | 2 | 2 | 0 |
| a value of type CompilerOptions \| TypeAcquisition \| WatchOptions \| undefined seen as CompilerOptions \| TypeAcquisition \| WatchOptions, which can write TsConfigSourceFile \| CompilerOptionsValue where CompilerOptionsValue is read | 1 | 1 | 1 | 1 | 0 |
| a value of type CompilerOptionsValue seen as TsConfigSourceFile \| CompilerOptionsValue, which can write string \| number where string is read | 5 | 5 | 5 | 5 | 0 |
| a value of type ComputedPropertyName seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ConditionalExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type ConstructSignatureDeclaration seen as Mutable<ConstructSignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ConstructSignatureDeclaration seen as Mutable<SignatureDeclarationBase>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ConstructorDeclaration seen as Mutable<ConstructorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type ConstructorDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ConstructorTypeNode seen as Mutable<ConstructorTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ConstructorTypeNode seen as Mutable<SignatureDeclarationBase>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ContinueStatement seen as Mutable<ContinueStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ContinueStatement seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ContinueStatement seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Declaration seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Declaration \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 6 | 6 | 6 | 6 | 0 |
| a value of type Declaration \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Declaration[] seen as (ArrayTypeNode \| Declaration \| NodeWithTypeArguments \| TupleTypeNode)[], which can write ArrayTypeNode \| Declaration \| NodeWithTypeArguments \| TupleTypeNode where Declaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Declaration[] seen as Node[], which can write Node where Declaration is read | 7 | 7 | 7 | 7 | 0 |
| a value of type Decorator[] \| undefined seen as {}[] \| undefined, which can write {} where Decorator is read | 3 | 3 | 3 | 3 | 0 |
| a value of type DestructuringAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Diagnostic seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 3 | 3 | 3 | 3 | 0 |
| a value of type DiagnosticMessageChain seen as DiagnosticMessageChain \| ReusableDiagnosticMessageChain, which can write ReusableDiagnosticMessageChain[] \| undefined where DiagnosticMessageChain[] \| undefined is read | 2 | 2 | 2 | 2 | 0 |
| a value of type DiagnosticMessageChain \| { messageText: string; category: DiagnosticCategory; code: number; repopulateInfo?: () => RepopulateDiagnosticChainInfo; canonicalHead?: CanonicalDiagnostic; next: ... \| undefined; } seen as ReusableDiagnosticMessageChain, which can write ReusableDiagnosticMessageChain[] \| undefined where DiagnosticMessageChain[] \| undefined is read | 1 | 1 | 1 | 1 | 0 |
| a value of type DiagnosticMessageChain[] seen as ReusableDiagnosticMessageChain[] \| undefined, which can write ReusableDiagnosticMessageChain where DiagnosticMessageChain is read | 1 | 1 | 1 | 1 | 0 |
| a value of type DiagnosticMessageChain[] seen as ReusableDiagnosticMessageChain[], which can write ReusableDiagnosticMessageChain[] \| undefined where DiagnosticMessageChain[] \| undefined is read | 1 | 1 | 1 | 1 | 0 |
| a value of type DiagnosticRelatedInformation[] \| undefined seen as {}[] \| undefined, which can write {} where DiagnosticRelatedInformation is read | 1 | 1 | 1 | 1 | 0 |
| a value of type DiagnosticWithDetachedLocation seen as Diagnostic, which can write SourceFile \| undefined where undefined is read | 4 | 4 | 4 | 4 | 0 |
| a value of type DiagnosticWithDetachedLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where undefined is read | 5 | 5 | 5 | 5 | 0 |
| a value of type DiagnosticWithLocation seen as Diagnostic \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 | 1 | 0 |
| a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 109 | 109 | 109 | 109 | 0 |
| a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 | 1 | 0 |
| a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where SourceFile is read | 69 | 69 | 69 | 69 | 0 |
| a value of type DiagnosticWithLocation \| undefined seen as Diagnostic \| undefined, which can write SourceFile \| undefined where SourceFile is read | 2 | 2 | 2 | 2 | 0 |
| a value of type DiagnosticWithLocation[] seen as Diagnostic[], which can write Diagnostic where DiagnosticWithLocation is read | 7 | 7 | 7 | 7 | 0 |
| a value of type DiagnosticWithLocation[] seen as readonly Diagnostic[] \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 | 1 | 0 |
| a value of type DiagnosticWithLocation[] seen as readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 | 1 | 0 |
| a value of type DiagnosticWithLocation[] \| undefined seen as DiagnosticRelatedInformation[] \| undefined, which can write DiagnosticRelatedInformation where DiagnosticWithLocation is read | 2 | 2 | 2 | 2 | 0 |
| a value of type DiagnosticWithLocation[] \| undefined seen as readonly Diagnostic[] \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 | 1 | 0 |
| a value of type DiagnosticWithLocation[] \| undefined seen as {}[] \| undefined, which can write {} where DiagnosticWithLocation is read | 1 | 1 | 1 | 1 | 0 |
| a value of type ElementAccessExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 6 | 6 | 6 | 6 | 0 |
| a value of type ElementWithComputedPropertyName[] \| undefined seen as {}[] \| undefined, which can write {} where ElementWithComputedPropertyName is read | 1 | 1 | 1 | 1 | 0 |
| a value of type EmitHelper[] \| undefined seen as {}[] \| undefined, which can write {} where EmitHelper is read | 2 | 2 | 2 | 2 | 0 |
| a value of type EmitNode & { autoGenerate: AutoGenerateInfo; } | 1 | 1 | 1 | 1 | 0 |
| a value of type EmptyStatement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type EntityNameExpression \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type EnumDeclaration seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type EnumDeclaration \| ModuleDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ExportAssignment & { readonly expression: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | 1 | 1 | 0 |
| a value of type ExportDeclaration \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Expression seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 6 | 6 | 6 | 6 | 0 |
| a value of type Expression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 9 | 9 | 9 | 9 | 0 |
| a value of type Expression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 42 | 42 | 42 | 42 | 0 |
| a value of type Expression \| GeneratedIdentifier seen as Expression \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Expression \| GeneratedIdentifier seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Expression \| Identifier seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Expression \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 5 | 5 | 5 | 5 | 0 |
| a value of type Expression \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ExpressionStatement seen as Mutable<ExpressionStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ExpressionStatement seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ExpressionStatement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 26 | 26 | 26 | 26 | 0 |
| a value of type ExpressionWithTypeArguments & { readonly expression: Identifier \| PropertyAccessEntityNameExpression; } | 1 | 1 | 1 | 1 | 0 |
| a value of type Expression[] seen as {}[], which can write {} where Expression is read | 2 | 2 | 2 | 2 | 0 |
| a value of type Expression[] \| undefined seen as {}[] \| undefined, which can write {} where Expression is read | 10 | 10 | 10 | 10 | 0 |
| a value of type Extension[] seen as string[], which can write string where Extension is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Extension[][] \| readonly (readonly Extension[] \| Extension \| undefined)[] where an array goes | 1 | 1 | 1 | 1 | 0 |
| a value of type FileReference seen as { pos: number \| undefined; end: number \| undefined; }, which can write number \| undefined where number is read | 3 | 3 | 3 | 3 | 0 |
| a value of type FlowArrayMutation \| FlowAssignment seen as FlowNode, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 5 | 5 | 5 | 5 | 0 |
| a value of type FlowCall seen as FlowNode, which can write BinaryExpression \| CallExpression where CallExpression is read | 2 | 2 | 2 | 2 | 0 |
| a value of type FlowNode seen as FlowNode \| undefined, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 19 | 19 | 19 | 19 | 0 |
| a value of type FlowNode seen as FlowNode[] \| FlowNode \| undefined, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 7 | 7 | 7 | 7 | 0 |
| a value of type FlowNode \| undefined seen as false \| FlowNode \| undefined, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 6 | 6 | 6 | 6 | 0 |
| a value of type ForStatement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 4 | 4 | 4 | 4 | 0 |
| a value of type FunctionDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type FunctionDeclaration[] \| VariableDeclaration[] \| undefined seen as Statement[] \| VariableDeclaration[] \| undefined, which can write Statement where FunctionDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type FunctionExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 7 | 7 | 7 | 7 | 0 |
| a value of type FunctionTypeNode seen as Mutable<FunctionTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type FutureSourceFile \| SourceFile seen as Pick<SourceFile, "fileName" \| "impliedNodeFormat" \| "packageJsonScope">, whose readonly field fileName becomes writable: a readonly field may hold something narrower than string, which a write of string would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type FutureSourceFile \| SourceFile seen as Pick<SourceFile, "fileName" \| "impliedNodeFormat">, whose readonly field fileName becomes writable: a readonly field may hold something narrower than string, which a write of string would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type GeneratedIdentifier seen as Expression \| GeneratedIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GeneratedIdentifier seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type GeneratedIdentifier seen as GeneratedIdentifier \| Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GeneratedIdentifier seen as Identifier \| PrivateIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type GeneratedIdentifier seen as Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 5 | 5 | 5 | 5 | 0 |
| a value of type GeneratedIdentifier seen as Node \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as GeneratedIdentifier \| GeneratedPrivateIdentifier \| Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as Identifier \| PrivateIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as Node \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 4 | 4 | 4 | 4 | 0 |
| a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier \| Node seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 5 | 5 | 5 | 5 | 0 |
| a value of type GeneratedIdentifier \| Identifier seen as GeneratedIdentifier \| Identifier \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GeneratedIdentifier \| Identifier seen as Identifier \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GeneratedIdentifier \| Identifier seen as Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 7 | 7 | 7 | 7 | 0 |
| a value of type GeneratedIdentifier \| Identifier seen as string \| BindingName, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GeneratedIdentifier \| Identifier seen as string \| GeneratedIdentifier \| Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GeneratedIdentifier \| Identifier \| undefined seen as string \| ModuleExportName \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GeneratedPrivateIdentifier seen as Node \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GenericType \| ResolvedType seen as ObjectType, which can write SymbolTable \| undefined where SymbolTable is read | 3 | 3 | 3 | 3 | 0 |
| a value of type GetAccessorDeclaration \| MethodDeclaration \| PropertyAssignment \| SetAccessorDeclaration \| ShorthandPropertyAssignment \| SpreadAssignment \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GetAccessorDeclaration \| SetAccessorDeclaration seen as Mutable<AccessorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type HasJSDoc \| undefined | 1 | 1 | 1 | 1 | 0 |
| a value of type Identifier & GeneratedIdentifier & { readonly escapedText: { __escapedIdentifier: void; } & "__this"; } seen as BindingName, which can write EmitNode \| undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Identifier seen as Mutable<Identifier>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type Identifier seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Identifier seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 6 | 6 | 6 | 6 | 0 |
| a value of type Identifier seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 21 | 21 | 21 | 21 | 0 |
| a value of type Identifier \| TextRange seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type Identifier \| undefined seen as Identifier \| TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type Identifier \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Identifier[] seen as ModuleExportName[] \| undefined, which can write ModuleExportName where Identifier is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Identifier[] seen as {}[], which can write {} where Identifier is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Identifier[] \| undefined seen as ModuleExportName[] \| undefined, which can write ModuleExportName where Identifier is read | 2 | 2 | 2 | 2 | 0 |
| a value of type Identifier[] \| undefined seen as {}[] \| undefined, which can write {} where Identifier is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Identifier[][] seen as ModuleExportName[][], which can write ModuleExportName[] where Identifier[] is read | 4 | 4 | 4 | 4 | 0 |
| a value of type IfStatement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type ImmediatelyInvokedArrowFunction seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ImportAttribute seen as Mutable<ImportAttribute>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ImportAttribute seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ImportAttributes seen as Mutable<ImportAttributes>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ImportAttributes seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ImportClause seen as Mutable<ImportClause>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ImportClause seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ImportDeclaration seen as Mutable<ImportDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ImportDeclaration seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type ImportTypeAssertionContainer seen as Mutable<ImportTypeAssertionContainer>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type ImportTypeAssertionContainer seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ImportTypeNode seen as Mutable<ImportTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ImportTypeNode seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type IncrementalBuildInfoDiagnosticOfFile[] \| undefined seen as {}[] \| undefined, which can write {} where IncrementalBuildInfoDiagnosticOfFile is read | 1 | 1 | 1 | 1 | 0 |
| a value of type IncrementalBuildInfoDiagnostic[] \| undefined seen as {}[] \| undefined, which can write {} where IncrementalBuildInfoDiagnostic is read | 2 | 2 | 2 | 2 | 0 |
| a value of type IncrementalBuildInfoEmitSignature[] \| undefined seen as {}[] \| undefined, which can write {} where IncrementalBuildInfoEmitSignature is read | 1 | 1 | 1 | 1 | 0 |
| a value of type IncrementalBuildInfoFileId[] \| undefined seen as {}[] \| undefined, which can write {} where IncrementalBuildInfoFileId is read | 1 | 1 | 1 | 1 | 0 |
| a value of type IncrementalBuildInfoFilePendingEmit[] \| undefined seen as {}[] \| undefined, which can write {} where IncrementalBuildInfoFilePendingEmit is read | 1 | 1 | 1 | 1 | 0 |
| a value of type IncrementalBuildInfoResolvedRoot[] \| undefined seen as {}[] \| undefined, which can write {} where IncrementalBuildInfoResolvedRoot is read | 1 | 1 | 1 | 1 | 0 |
| a value of type IncrementalMultiFileEmitBuildInfoFileInfo | 1 | 1 | 1 | 1 | 0 |
| a value of type IndexInfo \| undefined seen as IndexSignatureDeclaration \| undefined, which can write number \| undefined where number is read | 1 | 1 | 1 | 1 | 0 |
| a value of type IndexInfo[] seen as IndexInfo[], which can write IndexInfo where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type IndexInfo[] \| undefined seen as {}[] \| undefined, which can write {} where IndexInfo is read | 3 | 3 | 3 | 3 | 0 |
| a value of type IndexSignatureDeclaration seen as Mutable<IndexSignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type IndexSignatureDeclaration seen as Mutable<SignatureDeclarationBase>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type IndexSignatureDeclaration[] seen as Declaration[] \| undefined, which can write Declaration where IndexSignatureDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type InitializedVariableDeclaration | 2 | 2 | 2 | 2 | 0 |
| a value of type InitializedVariableDeclaration seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type IntrinsicType[] seen as TypeParameter[], which can write TypeParameter where IntrinsicType is read | 1 | 1 | 1 | 1 | 0 |
| a value of type IntrinsicType[] \| undefined seen as TypeParameter[] \| undefined, which can write TypeParameter where IntrinsicType is read | 1 | 1 | 1 | 1 | 0 |
| a value of type IntrinsicType[] \| undefined seen as TypeParameter[], which can write TypeParameter where IntrinsicType is read | 1 | 1 | 1 | 1 | 0 |
| a value of type IterationTypes[] seen as (IterationTypes \| undefined)[], which can write IterationTypes \| undefined where IterationTypes is read | 1 | 1 | 1 | 1 | 0 |
| a value of type IterationTypes[] \| undefined seen as {}[] \| undefined, which can write {} where IterationTypes is read | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocArray | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocAugmentsTag seen as Mutable<JSDocAugmentsTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocAugmentsTag seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocCallbackTag seen as Mutable<JSDocCallbackTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocCallbackTag seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocComment[] seen as Node[], which can write Node where JSDocComment is read | 3 | 3 | 3 | 3 | 0 |
| a value of type JSDocFunctionType seen as Mutable<JSDocFunctionType>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocFunctionType seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocImplementsTag seen as Mutable<JSDocImplementsTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocImplementsTag seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocImportTag seen as Mutable<JSDocImportTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocImportTag seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocImportTag \| CanHaveModuleSpecifier | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocLink seen as Mutable<JSDocLink>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocLink seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocLinkCode seen as Mutable<JSDocLinkCode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocLinkCode seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocLinkPlain seen as Mutable<JSDocLinkPlain>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocLinkPlain seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocNameReference seen as Mutable<JSDocNameReference>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocNameReference seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocOverloadTag seen as Mutable<JSDocOverloadTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocOverloadTag seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocParameterTag seen as Mutable<JSDocParameterTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocParameterTag seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocParameterTag[] seen as Node[], which can write Node where JSDocParameterTag is read | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocParameterTag[] \| undefined seen as {}[] \| undefined, which can write {} where JSDocParameterTag is read | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocPropertyLikeTag[] \| undefined seen as {}[] \| undefined, which can write {} where JSDocPropertyLikeTag is read | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocPropertyTag seen as Mutable<JSDocPropertyTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocPropertyTag seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocPropertyTag[] \| undefined seen as {}[] \| undefined, which can write {} where JSDocPropertyTag is read | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocReturnTag seen as JSDocTag & { typeExpression?: JSDocTypeExpression; }, whose readonly field typeExpression becomes writable: a readonly field may hold something narrower than JSDocTypeExpression \| undefined, which a write of JSDocTypeExpression \| undefined would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type JSDocSatisfiesTag seen as JSDocTag & { typeExpression?: JSDocTypeExpression; }, whose readonly field typeExpression becomes writable: a readonly field may hold something narrower than JSDocTypeExpression, which a write of JSDocTypeExpression \| undefined would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type JSDocSeeTag seen as Mutable<JSDocSeeTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocSeeTag seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocSignature seen as Mutable<JSDocSignature>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocSignature seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocSignature \| SignatureDeclaration \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocTag[] seen as Node[], which can write Node where JSDocTag is read | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocTemplateTag seen as Mutable<JSDocTemplateTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocTemplateTag seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocText seen as Mutable<JSDocText>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocText seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocThisTag seen as JSDocTag & { typeExpression?: JSDocTypeExpression; }, whose readonly field typeExpression becomes writable: a readonly field may hold something narrower than JSDocTypeExpression, which a write of JSDocTypeExpression \| undefined would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type JSDocThrowsTag seen as JSDocTag & { typeExpression?: JSDocTypeExpression; }, whose readonly field typeExpression becomes writable: a readonly field may hold something narrower than JSDocTypeExpression \| undefined, which a write of JSDocTypeExpression \| undefined would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type JSDocTypeExpression seen as Mutable<JSDocTypeExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocTypeExpression seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocTypeExpression \| undefined seen as Signature \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SignatureFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocTypeLiteral seen as Mutable<JSDocTypeLiteral>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocTypeLiteral seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocTypeTag seen as JSDocTag & { typeExpression?: JSDocTypeExpression; }, whose readonly field typeExpression becomes writable: a readonly field may hold something narrower than JSDocTypeExpression, which a write of JSDocTypeExpression \| undefined would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type JSDocTypedefTag seen as Mutable<JSDocTypedefTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocTypedefTag seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocUnknownTag seen as Mutable<JSDocUnknownTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocUnknownTag seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JsxAttributeLike \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JsxChild[] seen as Node[], which can write Node where JsxChild is read | 1 | 1 | 1 | 1 | 0 |
| a value of type LeftHandSideExpression & GeneratedIdentifier seen as Expression, which can write EmitNode \| undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 2 | 2 | 2 | 2 | 0 |
| a value of type LeftHandSideExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type Map<Path, Diagnostic[]> seen as Map<Path, readonly Diagnostic[]> \| undefined, which can write readonly Diagnostic[] where Diagnostic[] is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<Path, DirectoryWatchesOfFailedLookup> seen as Map<string, DirectoryWatchesOfFailedLookup>, which can write string where Path is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<Path, FileWatcher> seen as Map<string, FileWatcher>, which can write string where Path is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<Path, ModeAwareCache<CachedResolvedModuleWithFailedLookupLocations>> seen as Map<string, ModeAwareCache<CachedResolvedModuleWithFailedLookupLocations>>, which can write string where Path is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<Path, ModeAwareCache<CachedResolvedModuleWithFailedLookupLocations>> seen as Map<string, ModeAwareCache<ResolutionWithFailedLookupLocations>>, which can write string where Path is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<Path, ModeAwareCache<CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations>> seen as Map<string, ModeAwareCache<CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations>>, which can write string where Path is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<Path, ModeAwareCache<CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations>> seen as Map<string, ModeAwareCache<ResolutionWithFailedLookupLocations>>, which can write string where Path is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<Path, readonly ReusableDiagnostic[]> seen as Map<Path, readonly Diagnostic[] \| readonly ReusableDiagnostic[]>, which can write readonly Diagnostic[] \| readonly ReusableDiagnostic[] where readonly ReusableDiagnostic[] is read | 2 | 2 | 2 | 2 | 0 |
| a value of type Map<Path, readonly ReusableDiagnostic[]> \| undefined seen as Map<Path, readonly Diagnostic[] \| readonly ReusableDiagnostic[]> \| undefined, which can write readonly Diagnostic[] \| readonly ReusableDiagnostic[] where readonly ReusableDiagnostic[] is read | 2 | 2 | 2 | 2 | 0 |
| a value of type Map<Path, string[]> seen as InvokeMap, which can write true \| string[] where string[] is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<ResolvedConfigFilePath, BuildInfoCacheEntry> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where BuildInfoCacheEntry is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<ResolvedConfigFilePath, ConfigFileCacheEntry> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where ConfigFileCacheEntry is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<ResolvedConfigFilePath, Map<Path, Date>> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where Map<Path, Date> is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<ResolvedConfigFilePath, ProgramUpdateLevel> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where ProgramUpdateLevel is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<ResolvedConfigFilePath, Set<string> \| undefined> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where Set<string> \| undefined is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<ResolvedConfigFilePath, T> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where T is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<ResolvedConfigFilePath, UpToDateStatus> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where UpToDateStatus is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<ResolvedConfigFilePath, readonly Diagnostic[]> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where readonly Diagnostic[] is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<ResolvedConfigFilePath, true> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where true is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<__String, TransientSymbol> seen as SymbolTable \| undefined, which can write Symbol where TransientSymbol is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<string, CachedResolvedModuleWithFailedLookupLocations> seen as Map<string, ResolutionWithFailedLookupLocations> \| Set<ResolutionWithFailedLookupLocations> \| undefined, which can write ResolutionWithFailedLookupLocations where CachedResolvedModuleWithFailedLookupLocations is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<string, ImportsNotUsedAsValues> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ImportsNotUsedAsValues is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<string, JsxEmit> seen as Map<string, string \| number>, which can write string \| number where JsxEmit is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<string, Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>> seen as ScriptTargetFeatures, which can write string where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<string, ModuleDetectionKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ModuleDetectionKind is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<string, ModuleResolutionKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ModuleResolutionKind is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<string, NewLineKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where NewLineKind is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<string, PollingWatchKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where PollingWatchKind is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<string, SingleFileWatcher<FileWatcherCallback>> seen as Map<string, SingleFileWatcher<FileWatcherCallback \| FsWatchCallback>>, which can write SingleFileWatcher<FileWatcherCallback \| FsWatchCallback> where SingleFileWatcher<FileWatcherCallback> is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<string, SingleFileWatcher<FsWatchCallback>> seen as Map<string, SingleFileWatcher<FileWatcherCallback \| FsWatchCallback>>, which can write SingleFileWatcher<FileWatcherCallback \| FsWatchCallback> where SingleFileWatcher<FsWatchCallback> is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<string, WatchDirectoryKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where WatchDirectoryKind is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<string, WatchFileKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where WatchFileKind is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<string, [VariableDeclarationList, VariableDeclaration[]]> seen as Map<string, [CatchClause \| VariableDeclarationList, VariableDeclaration[]]>, which can write [CatchClause \| VariableDeclarationList, VariableDeclaration[]] where [VariableDeclarationList, VariableDeclaration[]] is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<string, never[]> seen as Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>, which can write string[] where never[] is read | 10 | 10 | 10 | 10 | 0 |
| a value of type Map<string, string> seen as Map<string, string \| number>, which can write string \| number where string is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<string, string[] \| never[]> seen as Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>, which can write string where never is read | 3 | 3 | 3 | 3 | 0 |
| a value of type MapLike<string[]> seen as MapLike<readonly string[]>, which can write readonly string[] where string[] is read | 2 | 2 | 2 | 2 | 0 |
| a value of type MappedTypeNode seen as Mutable<MappedTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type MappedTypeNode seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type MemberName \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type MethodDeclaration seen as Mutable<MethodDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type MethodDeclaration \| PropertyAssignment \| AccessorDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type MissingDeclaration seen as Mutable<MissingDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type ModifierLike seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type ModifierLike[] seen as Node[], which can write Node where ModifierLike is read | 1 | 1 | 1 | 1 | 0 |
| a value of type ModifierLike[] \| undefined seen as {}[] \| undefined, which can write {} where ModifierLike is read | 4 | 4 | 4 | 4 | 0 |
| a value of type Modifier[] \| undefined seen as {}[] \| undefined, which can write {} where Modifier is read | 1 | 1 | 1 | 1 | 0 |
| a value of type ModuleBlock seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ModuleDeclaration seen as Mutable<ModuleDeclaration \| SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ModuleDeclaration seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ModuleDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ModuleExportName[] \| undefined seen as {}[] \| undefined, which can write {} where ModuleExportName is read | 4 | 4 | 4 | 4 | 0 |
| a value of type ModuleName seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ModuleName seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ModuleResolutionState & { host: GetPackageJsonEntrypointsHost; } | 1 | 1 | 1 | 1 | 0 |
| a value of type ModuleResolutionState & { host: GetPackageJsonEntrypointsHost; } seen as ModuleResolutionState, which can write ModuleResolutionHost where ModuleResolutionHost & GetPackageJsonEntrypointsHost is read | 1 | 1 | 1 | 1 | 0 |
| a value of type ModuleSpecifierResolutionHost & ModuleResolutionHost seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where (() => boolean) & (boolean \| (() => boolean) \| undefined) is read | 2 | 2 | 2 | 2 | 0 |
| a value of type Mutable<ArrayBindingPattern> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ArrayBindingPattern is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ArrayLiteralExpression> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ArrayLiteralExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ArrayTypeNode> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ArrayType is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ArrowFunction> seen as Mutable<SignatureDeclarationBase>, which can write SyntaxKind.MethodSignature \| SyntaxKind.MethodDeclaration \| SyntaxKind.Constructor \| SyntaxKind.GetAccessor \| SyntaxKind.SetAccessor \| SyntaxKind.CallSignature \| SyntaxKind.ConstructSignature \| ... 6 more ... \| SyntaxKind.JSDocFunctionType where SyntaxKind.ArrowFunction is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<AsExpression> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.AsExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<AwaitExpression> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.AwaitExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<BinaryExpression> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.BinaryExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<BindingElement> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.BindingElement is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<Bundle> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.Bundle is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<CallChain> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.CallExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<CallExpression> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.CallExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<CaseClause> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.CaseClause is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<CatchClause> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.CatchClause is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ClassDeclaration> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ClassDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ClassExpression> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ClassExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ClassStaticBlockDeclaration> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ClassStaticBlockDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<CommaListExpression> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.CommaListExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ComputedPropertyName> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ComputedPropertyName is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ConditionalExpression> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ConditionalExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ConditionalTypeNode> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ConditionalType is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ConstructorDeclaration> seen as Mutable<SignatureDeclarationBase>, which can write Node where ClassLikeDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<Decorator> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.Decorator is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<DefaultClause> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.DefaultClause is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<DeleteExpression> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.DeleteExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<DoStatement> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.DoStatement is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ElementAccessChain> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ElementAccessExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ElementAccessExpression> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ElementAccessExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<EnumDeclaration> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.EnumDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<EnumMember> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.EnumMember is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ExportAssignment> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ExportAssignment is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ExportDeclaration> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ExportDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ExportSpecifier> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ExportSpecifier is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ExpressionWithTypeArguments> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ExpressionWithTypeArguments is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ExternalModuleReference> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ExternalModuleReference is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ForInStatement> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ForInStatement is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ForOfStatement> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ForOfStatement is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ForStatement> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ForStatement is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<FunctionDeclaration> seen as Mutable<SignatureDeclarationBase>, which can write SyntaxKind.MethodSignature \| SyntaxKind.MethodDeclaration \| SyntaxKind.Constructor \| SyntaxKind.GetAccessor \| SyntaxKind.SetAccessor \| SyntaxKind.CallSignature \| SyntaxKind.ConstructSignature \| ... 6 more ... \| SyntaxKind.JSDocFunctionType where SyntaxKind.FunctionDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<FunctionExpression> seen as Mutable<SignatureDeclarationBase>, which can write SyntaxKind.MethodSignature \| SyntaxKind.MethodDeclaration \| SyntaxKind.Constructor \| SyntaxKind.GetAccessor \| SyntaxKind.SetAccessor \| SyntaxKind.CallSignature \| SyntaxKind.ConstructSignature \| ... 6 more ... \| SyntaxKind.JSDocFunctionType where SyntaxKind.FunctionExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<FunctionTypeNode> seen as Mutable<SignatureDeclarationBase>, which can write SyntaxKind.MethodSignature \| SyntaxKind.MethodDeclaration \| SyntaxKind.Constructor \| SyntaxKind.GetAccessor \| SyntaxKind.SetAccessor \| SyntaxKind.CallSignature \| SyntaxKind.ConstructSignature \| ... 6 more ... \| SyntaxKind.JSDocFunctionType where SyntaxKind.FunctionType is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<GeneratedIdentifier> seen as Identifier \| PrivateIdentifier, which can write EmitNode \| undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 2 | 2 | 2 | 2 | 0 |
| a value of type Mutable<GeneratedIdentifier> seen as Identifier, which can write EmitNode \| undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 4 | 4 | 4 | 4 | 0 |
| a value of type Mutable<GeneratedIdentifier> seen as Node, which can write EmitNode \| undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<GeneratedPrivateIdentifier> seen as Identifier \| PrivateIdentifier, which can write EmitNode \| undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<GeneratedPrivateIdentifier> seen as Node, which can write EmitNode \| undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<GetAccessorDeclaration> seen as Mutable<SignatureDeclarationBase>, which can write Node where InterfaceDeclaration \| ObjectLiteralExpression \| TypeLiteralNode \| ClassLikeDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<HeritageClause> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.HeritageClause is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<IfStatement> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.IfStatement is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ImportEqualsDeclaration> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ImportEqualsDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ImportSpecifier> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ImportSpecifier is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<IndexedAccessTypeNode> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.IndexedAccessType is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<InferTypeNode> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.InferType is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<InterfaceDeclaration> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.InterfaceDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<JSDoc> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.JSDoc is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<JSDocEnumTag> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.JSDocEnumTag is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<JSDocMemberName> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.JSDocMemberName is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<JsxAttribute> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.JsxAttribute is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<JsxAttributes> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.JsxAttributes is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<JsxClosingElement> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.JsxClosingElement is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<JsxElement> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.JsxElement is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<JsxExpression> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.JsxExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<JsxFragment> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.JsxFragment is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<JsxNamespacedName> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.JsxNamespacedName is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<JsxOpeningElement> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.JsxOpeningElement is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<JsxSelfClosingElement> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.JsxSelfClosingElement is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<JsxSpreadAttribute> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.JsxSpreadAttribute is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<JsxText> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.JsxText is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<LabeledStatement> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.LabeledStatement is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<LiteralTypeNode> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.LiteralType is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<MetaProperty> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.MetaProperty is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<MethodDeclaration> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.MethodDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<MethodSignature> seen as Mutable<SignatureDeclarationBase>, which can write Node where InterfaceDeclaration \| TypeLiteralNode is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ModuleBlock> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ModuleBlock is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ModuleDeclaration> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ModuleDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<NamedExports> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.NamedExports is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<NamedTupleMember> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.NamedTupleMember is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<NamespaceExportDeclaration> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.NamespaceExportDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<NewExpression> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.NewExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<NoSubstitutionTemplateLiteral> seen as Mutable<TemplateLiteralLikeNode>, which can write SyntaxKind where SyntaxKind.NoSubstitutionTemplateLiteral is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<NonNullChain> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.NonNullExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<NonNullExpression> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.NonNullExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ObjectBindingPattern> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ObjectBindingPattern is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ObjectLiteralExpression> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ObjectLiteralExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<OptionalTypeNode> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.OptionalType is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ParameterDeclaration> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.Parameter is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ParenthesizedExpression> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ParenthesizedExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ParenthesizedTypeNode> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ParenthesizedType is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<PartiallyEmittedExpression> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.PartiallyEmittedExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<PostfixUnaryExpression> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.PostfixUnaryExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<PrefixUnaryExpression> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.PrefixUnaryExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<PropertyAccessChain> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.PropertyAccessExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<PropertyAccessExpression> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.PropertyAccessExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<PropertyAssignment> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.PropertyAssignment is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<PropertyDeclaration> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.PropertyDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<PropertySignature> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.PropertySignature is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<QualifiedName> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.QualifiedName is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<RestTypeNode> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.RestType is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<SatisfiesExpression> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.SatisfiesExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<SetAccessorDeclaration> seen as Mutable<SignatureDeclarationBase>, which can write Node where InterfaceDeclaration \| ObjectLiteralExpression \| TypeLiteralNode \| ClassLikeDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ShorthandPropertyAssignment> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ShorthandPropertyAssignment is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<SourceFile> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.SourceFile is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<SpreadAssignment> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.SpreadAssignment is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<SpreadElement> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.SpreadElement is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<SyntheticReferenceExpression> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.SyntheticReferenceExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<T> seen as Mutable<Node>, which can write SyntaxKind where T["kind"] is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<T> seen as T, a type parameter whose constraint JSDocType & { readonly type: TypeNode \| undefined; readonly postfix: boolean; } can be written, so it can write what Mutable<T> can't hold | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<T> seen as T, a type parameter whose constraint JSDocType & { readonly type: TypeNode \| undefined; } can be written, so it can write what Mutable<T> can't hold | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<T> seen as T, a type parameter whose constraint Node can be written, so it can write what Mutable<T> can't hold | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<TaggedTemplateExpression> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.TaggedTemplateExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<TemplateExpression> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.TemplateExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<TemplateLiteralTypeNode> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.TemplateLiteralType is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<TemplateLiteralTypeSpan> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.TemplateLiteralTypeSpan is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<TemplateSpan> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.TemplateSpan is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<ThrowStatement> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.ThrowStatement is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<TryStatement> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.TryStatement is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<TupleTypeNode> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.TupleType is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<TypeAliasDeclaration> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.TypeAliasDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<TypeAssertion> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.TypeAssertionExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<TypeLiteralNode> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.TypeLiteral is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<TypeOfExpression> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.TypeOfExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<TypePredicateNode> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.TypePredicate is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<TypeQueryNode> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.TypeQuery is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<TypeReferenceNode> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.TypeReference is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<VariableDeclaration> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.VariableDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<VariableDeclarationList> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.VariableDeclarationList is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<VariableStatement> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.VariableStatement is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<VoidExpression> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.VoidExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<WhileStatement> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.WhileStatement is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<WithStatement> seen as Mutable<Node>, which can write SyntaxKind where SyntaxKind.WithStatement is read | 1 | 1 | 1 | 1 | 0 |
| a value of type NamedImports seen as Mutable<NamedImports>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type NamedImports seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type NamespaceExport seen as Mutable<NamespaceExport>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type NamespaceExport seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type NamespaceExportDeclaration seen as Mutable<NamespaceExportDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type NamespaceImport seen as Mutable<NamespaceImport>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type NamespaceImport seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type NewExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Node seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 5 | 5 | 5 | 5 | 0 |
| a value of type Node seen as Node \| SourceMapRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Node seen as Node \| TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Node seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Node seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Node \| SourceMapRange seen as SourceMapRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Node \| TextRange seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type Node \| undefined seen as EmitNode \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of EmitFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Node \| undefined seen as NodeLinks \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of NodeCheckFlags would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type Node \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Node \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type NodeArray<ClassElement> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 7 | 7 | 7 | 7 | 0 |
| a value of type NodeArray<Expression> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type NodeArray<ExpressionStatement> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type NodeArray<ModifierLike> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type NodeArray<Node> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type NodeArray<ParameterDeclaration> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type NodeArray<Statement> seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 4 | 4 | 4 | 4 | 0 |
| a value of type NodeArray<Statement> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 32 | 32 | 32 | 32 | 0 |
| a value of type NodeArray<T> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type NodeArray<never> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type NodeBuilderContext seen as SyntacticTypeNodeBuilderContext, which can write Required<Pick<SymbolTracker, "reportInferenceFallback">> where SymbolTrackerImpl is read | 11 | 11 | 11 | 11 | 0 |
| a value of type Node[] seen as unknown[], which can write unknown where Node is read | 12 | 12 | 12 | 12 | 0 |
| a value of type NonNullExpression seen as Mutable<NonNullExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type NonNullable<T>[] seen as Node[], which can write Node where NonNullable<T> is read | 1 | 1 | 1 | 1 | 0 |
| a value of type NumberLiteralType seen as LiteralType, which can write string \| number \| PseudoBigInt where number is read | 1 | 1 | 1 | 1 | 0 |
| a value of type NumericLiteral seen as Mutable<LiteralExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type NumericLiteral seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ObjectBindingOrAssignmentPattern seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ObjectLiteralElementLike[] \| undefined seen as {}[] \| undefined, which can write {} where ObjectLiteralElementLike is read | 4 | 4 | 4 | 4 | 0 |
| a value of type ObjectLiteralExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type OptionalTypeNode seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type OptionalTypeNode seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ParameterDeclaration & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | 1 | 1 | 0 |
| a value of type ParameterDeclaration seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ParameterDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 8 | 8 | 8 | 8 | 0 |
| a value of type ParameterDeclaration \| PropertyDeclaration \| PropertySignature \| SignatureDeclaration seen as Mutable<ParameterDeclaration \| PropertyDeclaration \| PropertySignature \| SignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ParameterDeclaration \| VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ParameterDeclaration \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type ParameterDeclaration[] \| undefined seen as {}[] \| undefined, which can write {} where ParameterDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type ParameterPropertyDeclaration | 1 | 1 | 1 | 1 | 0 |
| a value of type ParenthesizedExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 8 | 8 | 8 | 8 | 0 |
| a value of type ParseConfigHost seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where boolean is read | 1 | 1 | 1 | 1 | 0 |
| a value of type PathPathComponents | 1 | 1 | 1 | 1 | 0 |
| a value of type Path[] seen as string[], which can write string where Path is read | 1 | 1 | 1 | 1 | 0 |
| a value of type PollingIntervalQueue | 1 | 1 | 1 | 1 | 0 |
| a value of type PostfixUnaryExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type PragmaPseudoMap[TKey][] \| PragmaPseudoMap[TKey] | 1 | 1 | 1 | 1 | 0 |
| a value of type PrimitiveLiteral | 1 | 1 | 1 | 1 | 0 |
| a value of type PrivateIdentifierInExpression | 1 | 1 | 1 | 1 | 0 |
| a value of type PropertyAccessExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 16 | 16 | 16 | 16 | 0 |
| a value of type PropertyAccessExpression \| SyntheticSuper seen as LeftHandSideExpression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type PropertyAccessExpression \| SyntheticSuper seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type PropertyAccessExpression \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type PropertyAssignment & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | 1 | 1 | 0 |
| a value of type PropertyAssignment seen as Mutable<PropertyAssignment \| ShorthandPropertyAssignment>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type PropertyAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 9 | 9 | 9 | 9 | 0 |
| a value of type PropertyDeclaration & { readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | 1 | 1 | 0 |
| a value of type PropertyDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type PropertyDeclaration[] \| undefined seen as ClassElement[] \| undefined, which can write ClassElement where PropertyDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type PropertyName seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type PropertyName \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type PropertyName[] \| undefined seen as {}[] \| undefined, which can write {} where PropertyName is read | 1 | 1 | 1 | 1 | 0 |
| a value of type PropertySignature seen as Mutable<PropertySignature>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type Readonly<BuilderState> \| undefined seen as BuilderState \| undefined, whose readonly field fileInfos becomes writable: a readonly field may hold something narrower than Map<Path, FileInfo>, which a write of Map<Path, FileInfo> would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Readonly<PathPathComponents> | 7 | 7 | 7 | 7 | 0 |
| a value of type Readonly<ReusableBuilderProgramState> \| undefined seen as BuilderState \| undefined, whose readonly field fileInfos becomes writable: a readonly field may hold something narrower than Map<Path, FileInfo>, which a write of Map<Path, FileInfo> would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ReferencedFile & { kind: FileIncludeKind.LibReferenceDirective; } seen as ReferencedFile, which can write ReferencedFileKind where FileIncludeKind.LibReferenceDirective is read | 1 | 1 | 1 | 1 | 0 |
| a value of type ReportFileInError[] seen as (ReportFileInError \| undefined)[], which can write ReportFileInError \| undefined where ReportFileInError is read | 1 | 1 | 1 | 1 | 0 |
| a value of type RequireOrImportCall | 1 | 1 | 1 | 1 | 0 |
| a value of type RequireOrImportCall[] \| undefined seen as {}[] \| undefined, which can write {} where RequireOrImportCall is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Resolved \| undefined seen as { resolvedFileName: string; originalPath: string \| undefined; extension: string; isExternalLibraryImport: boolean; packageId: PackageId \| undefined; resolvedUsingTsExtension: boolean; } \| undefined, whose readonly field originalPath becomes writable: a readonly field may hold something narrower than string \| true \| undefined, which a write of string \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ResolvedConfigFileName | 1 | 1 | 1 | 1 | 0 |
| a value of type ResolvedModuleFull \| undefined seen as { path: string; originalPath: string \| true; extension: string; packageId: PackageId \| undefined; resolvedUsingTsExtension: boolean \| undefined; } \| undefined, whose readonly field originalPath becomes writable: a readonly field may hold something narrower than string \| undefined, which a write of string \| true would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ResolvedType seen as ObjectType, which can write SymbolTable \| undefined where SymbolTable is read | 17 | 17 | 17 | 17 | 0 |
| a value of type ResolvedType \| TypeReference seen as ObjectType, which can write SymbolTable \| undefined where SymbolTable is read | 1 | 1 | 1 | 1 | 0 |
| a value of type ReturnStatement seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ReturnStatement seen as Mutable<ReturnStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ReturnStatement seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ReturnStatement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 12 | 12 | 12 | 12 | 0 |
| a value of type SearchResult<Resolved> seen as { value: { resolved: Resolved; isExternalLibraryImport: true; } \| undefined; } \| undefined, whose readonly field value becomes writable: a readonly field may hold something narrower than Resolved \| undefined, which a write of { resolved: Resolved; isExternalLibraryImport: true; } \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Set<Path> \| undefined seen as Set<string> \| undefined, which can write string where Path is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Set<__String> seen as Set<__String \| undefined>, which can write __String \| undefined where __String is read | 1 | 1 | 1 | 1 | 0 |
| a value of type SetAccessorDeclaration seen as Mutable<SetAccessorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ShorthandPropertyAssignment & { readonly objectAssignmentInitializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | 1 | 1 | 0 |
| a value of type ShorthandPropertyAssignment seen as Mutable<PropertyAssignment \| ShorthandPropertyAssignment>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ShorthandPropertyAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Signature \| undefined seen as Type \| undefined, which can write TypeFlags where SignatureFlags is read | 3 | 3 | 3 | 3 | 0 |
| a value of type Signature[] \| undefined seen as {}[] \| undefined, which can write {} where Signature is read | 2 | 2 | 2 | 2 | 0 |
| a value of type SortedArray<Diagnostic> | 1 | 1 | 1 | 1 | 0 |
| a value of type SortedReadonlyArray<Canonicalized> | 1 | 1 | 1 | 1 | 0 |
| a value of type SortedReadonlyArray<MappedPosition> \| undefined | 1 | 1 | 1 | 1 | 0 |
| a value of type SortedReadonlyArray<string> | 1 | 1 | 1 | 1 | 0 |
| a value of type SourceFile seen as EmitNode \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of EmitFlags would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type SourceFile seen as Mutable<ModuleDeclaration \| SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type SourceFile seen as Mutable<SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type SourceFile \| undefined seen as FileReasonToChainCache \| undefined, which can write DiagnosticMessageChain[] \| undefined where RedirectInfo \| undefined is read | 1 | 1 | 1 | 1 | 0 |
| a value of type SourceFile \| undefined seen as NodeLinks \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of NodeCheckFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type SourceFile \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type SpreadAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type SpreadElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Statement seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type Statement[] seen as Node[], which can write Node where Statement is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Statement[] seen as {}[], which can write {} where Statement is read | 11 | 11 | 11 | 11 | 0 |
| a value of type Statement[] \| undefined seen as CaseClause[] \| undefined, which can write CaseClause where Statement is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Statement[] \| undefined seen as {}[] \| undefined, which can write {} where Statement is read | 28 | 28 | 28 | 28 | 0 |
| a value of type StringLiteral seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type StringLiteral seen as Mutable<StringLiteral>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type StringLiteral seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 6 | 6 | 6 | 6 | 0 |
| a value of type StringLiteralLike[] \| undefined seen as {}[] \| undefined, which can write {} where StringLiteralLike is read | 2 | 2 | 2 | 2 | 0 |
| a value of type StringLiteralType seen as LiteralType, which can write string \| number \| PseudoBigInt where string is read | 1 | 1 | 1 | 1 | 0 |
| a value of type StringLiteralType[] seen as Type[], which can write Type where StringLiteralType is read | 1 | 1 | 1 | 1 | 0 |
| a value of type StringLiteralType[] seen as {}[], which can write {} where StringLiteralType is read | 1 | 1 | 1 | 1 | 0 |
| a value of type SuperExpression seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type SuperExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type SwitchStatement seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type SwitchStatement seen as Mutable<SwitchStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Symbol seen as Declaration \| undefined, which can write number \| undefined where number is read | 2 | 2 | 2 | 2 | 0 |
| a value of type Symbol seen as Type \| undefined, which can write TypeFlags where SymbolFlags is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Symbol seen as Type, which can write TypeFlags where SymbolFlags is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Symbol \| undefined seen as Declaration \| undefined, which can write number \| undefined where number is read | 5 | 5 | 5 | 5 | 0 |
| a value of type Symbol \| undefined seen as GenericType \| undefined, which can write TypeFlags where SymbolFlags is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Symbol \| undefined seen as Node \| undefined, which can write number \| undefined where number is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Symbol \| undefined seen as SourceFile \| undefined, which can write number \| undefined where number is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Symbol \| undefined seen as Type \| undefined, which can write TypeFlags where SymbolFlags is read | 10 | 10 | 10 | 10 | 0 |
| a value of type Symbol[] seen as Symbol[], which can write Symbol where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Symbol[] seen as unknown[], which can write unknown where Symbol is read | 2 | 2 | 2 | 2 | 0 |
| a value of type Symbol[] seen as {}[], which can write {} where Symbol is read | 3 | 3 | 3 | 3 | 0 |
| a value of type Symbol[] \| undefined seen as {}[] \| undefined, which can write {} where Symbol is read | 4 | 4 | 4 | 4 | 0 |
| a value of type SyntheticExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type SyntheticSuper seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 4 | 4 | 4 | 4 | 0 |
| a value of type SyntheticSuper seen as string \| BindingName, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type System seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where boolean is read | 1 | 1 | 1 | 1 | 0 |
| a value of type T | 2 | 2 | 2 | 2 | 0 |
| a value of type T seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type T seen as Mutable<T>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of T["pos"] would replace | 11 | 11 | 11 | 11 | 0 |
| a value of type T seen as T, a type parameter whose constraint ClassDeclaration \| ClassExpression \| GetAccessorDeclaration \| MethodDeclaration \| ParameterDeclaration \| PropertyDeclaration \| SetAccessorDeclaration can be written, so it can write what T can't hold | 1 | 1 | 1 | 1 | 0 |
| a value of type T seen as T, a type parameter whose constraint EntityNameOrEntityNameExpression can be written, so it can write what T can't hold | 1 | 1 | 1 | 1 | 0 |
| a value of type T seen as T, a type parameter whose constraint HasModifiers can be written, so it can write what T can't hold | 1 | 1 | 1 | 1 | 0 |
| a value of type T seen as T, a type parameter whose constraint MethodDeclaration \| MethodSignature \| PropertyAssignment \| PropertyDeclaration \| PropertySignature \| AccessorDeclaration can be written, so it can write what T can't hold | 1 | 1 | 1 | 1 | 0 |
| a value of type T seen as T, a type parameter whose constraint ModifierSyntaxKind can be written, so it can write what T can't hold | 1 | 1 | 1 | 1 | 0 |
| a value of type T seen as T, a type parameter whose constraint Node can be written, so it can write what T can't hold | 2 | 2 | 2 | 2 | 0 |
| a value of type T seen as T, a type parameter whose constraint Node \| undefined can be written, so it can write what T can't hold | 1 | 1 | 1 | 1 | 0 |
| a value of type T seen as T, a type parameter whose constraint ResolutionWithFailedLookupLocations can be written, so it can write what T can't hold | 1 | 1 | 1 | 1 | 0 |
| a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 4 | 4 | 4 | 4 | 0 |
| a value of type T \| undefined | 1 | 1 | 1 | 1 | 0 |
| a value of type TKind seen as TKind, a type parameter whose constraint KeywordTypeSyntaxKind can be written, so it can write what TKind can't hold | 1 | 1 | 1 | 1 | 0 |
| a value of type T[] seen as (DiagnosticMessageChain \| ReusableDiagnosticMessageChain)[] \| undefined, which can write DiagnosticMessageChain \| ReusableDiagnosticMessageChain where T is read | 3 | 3 | 3 | 3 | 0 |
| a value of type T[] seen as (T \| undefined)[], which can write T \| undefined where T is read | 1 | 1 | 1 | 1 | 0 |
| a value of type T[] seen as Node[], which can write Node where T is read | 1 | 1 | 1 | 1 | 0 |
| a value of type T[] seen as Statement[], which can write Statement where T is read | 4 | 4 | 4 | 4 | 0 |
| a value of type T[] seen as {}[], which can write {} where T is read | 2 | 2 | 2 | 2 | 0 |
| a value of type T[] \| undefined seen as (T \| undefined)[], which can write T \| undefined where T is read | 1 | 1 | 1 | 1 | 0 |
| a value of type TaggedTemplateExpression seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type TemplateLiteralTypeSpan[] seen as Node[], which can write Node where TemplateLiteralTypeSpan is read | 1 | 1 | 1 | 1 | 0 |
| a value of type TemplateSpan[] seen as Node[], which can write Node where TemplateSpan is read | 1 | 1 | 1 | 1 | 0 |
| a value of type TemplateStringsArray | 1 | 1 | 1 | 1 | 0 |
| a value of type ThisCapturingVariableDeclaration seen as VariableDeclaration, which can write EmitNode \| undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 1 | 1 | 1 | 1 | 0 |
| a value of type ThrowStatement seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ThrowStatement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type TransientSymbol[] seen as Symbol[], which can write Symbol where TransientSymbol is read | 1 | 1 | 1 | 1 | 0 |
| a value of type TsConfigSourceFile \| CompilerOptionsValue seen as string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined, which can write string \| number where string is read | 3 | 3 | 3 | 3 | 0 |
| a value of type Type seen as TypeNode, which can write number \| undefined where number is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Type \| undefined seen as Expression \| undefined, which can write number \| undefined where number is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Type \| undefined seen as JSDocTypeExpression \| undefined, which can write number \| undefined where number is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Type \| undefined seen as Signature \| undefined, which can write SignatureFlags where TypeFlags is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Type \| undefined seen as Symbol \| undefined, which can write SymbolFlags where TypeFlags is read | 4 | 4 | 4 | 4 | 0 |
| a value of type Type \| undefined seen as TypeNode \| undefined, which can write number \| undefined where number is read | 3 | 3 | 3 | 3 | 0 |
| a value of type TypeAcquisition seen as CompilerOptions \| TypeAcquisition, which can write TsConfigSourceFile \| CompilerOptionsValue where CompilerOptionsValue is read | 1 | 1 | 1 | 1 | 0 |
| a value of type TypeCheckerHost seen as Program, whose readonly field redirectTargetsMap becomes writable: a readonly field may hold something narrower than RedirectTargetsMap, which a write of MultiMap<Path, string> would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type TypeNode \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 7 | 7 | 7 | 7 | 0 |
| a value of type TypeNode[] seen as Node[], which can write Node where TypeNode is read | 1 | 1 | 1 | 1 | 0 |
| a value of type TypeOperatorNode seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type TypeOperatorNode seen as Mutable<TypeOperatorNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type TypeParameterDeclaration & { parent: JSDocTemplateTag; } | 1 | 1 | 1 | 1 | 0 |
| a value of type TypeParameterDeclaration seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type TypeParameterDeclaration seen as Mutable<TypeParameterDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type TypeParameterDeclaration \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type TypeParameterDeclaration[] seen as Node[], which can write Node where TypeParameterDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type TypeParameter[] \| undefined seen as {}[] \| undefined, which can write {} where TypeParameter is read | 4 | 4 | 4 | 4 | 0 |
| a value of type TypeVariable[] seen as (Type \| undefined)[], which can write Type \| undefined where TypeVariable is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Type[] seen as (Type \| undefined)[], which can write Type \| undefined where Type is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Type[] seen as unknown[], which can write unknown where Type is read | 2 | 2 | 2 | 2 | 0 |
| a value of type Type[] seen as {}[], which can write {} where Type is read | 5 | 5 | 5 | 5 | 0 |
| a value of type Type[] \| undefined seen as {}[] \| undefined, which can write {} where Type is read | 24 | 24 | 24 | 24 | 0 |
| a value of type VariableDeclaration & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | 1 | 1 | 0 |
| a value of type VariableDeclaration seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 9 | 9 | 9 | 9 | 0 |
| a value of type VariableDeclaration \| DestructuringAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type VariableDeclarationList seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 13 | 13 | 13 | 13 | 0 |
| a value of type VariableDeclaration[] seen as {}[], which can write {} where VariableDeclaration is read | 2 | 2 | 2 | 2 | 0 |
| a value of type VariableDeclaration[] \| undefined seen as {}[] \| undefined, which can write {} where VariableDeclaration is read | 3 | 3 | 3 | 3 | 0 |
| a value of type VariableStatement seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type VariableStatement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 13 | 13 | 13 | 13 | 0 |
| a value of type VariableStatement[] seen as (ExportDeclaration \| VariableStatement)[], which can write ExportDeclaration \| VariableStatement where VariableStatement is read | 1 | 1 | 1 | 1 | 0 |
| a value of type WatchCompilerHostOfFilesAndCompilerOptionsOrConfigFile<T> seen as (CompilerHost \| ProgramHost<T>) & { onUnRecoverableConfigFileDiagnostic?: DiagnosticReporter; }, which can write DiagnosticReporter \| undefined where DiagnosticReporter is read | 1 | 1 | 1 | 1 | 0 |
| a value of type WatchCompilerHostOfFilesAndCompilerOptionsOrConfigFile<T> seen as DirectoryStructureHost, which can write ((path: string, extensions?: readonly string[] \| undefined, exclude?: readonly string[] \| undefined, include?: readonly string[] \| undefined, depth?: number \| undefined) => string[]) \| undefined where ((path: string, extensions?: readonly string[] \| undefined, exclude?: readonly string[] \| undefined, include?: readonly string[] \| undefined, depth?: number \| undefined) => string[]) & ((path: string, extensions?: ... \| undefined, exclude?: ... \| undefined, include?: ... \| undefined, depth?: number \| undefined) => .... is read | 1 | 1 | 1 | 1 | 0 |
| a value of type WatchCompilerHostOfFilesAndCompilerOptionsOrConfigFile<T> seen as WatchFactoryHost & { trace?(s: string): void; }, which can write boolean \| (() => boolean) where () => boolean is read | 1 | 1 | 1 | 1 | 0 |
| a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read | 1 | 1 | 1 | 1 | 0 |
| a value of type WildcardDirectoryWatcher<T> seen as { watcher: FileWatcher; }, which can write FileWatcher where T is read | 2 | 2 | 2 | 2 | 0 |
| a value of type WrappedExpression<AnonymousFunctionDefinition> | 2 | 2 | 2 | 2 | 0 |
| a value of type YieldExpression seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type YieldExpression seen as Mutable<YieldExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type YieldExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type YieldExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 4 | 4 | 4 | 4 | 0 |
| a value of type [() => Type, __String][] seen as (readonly [() => Type, __String])[], which can write readonly [() => Type, __String] where [() => Type, __String] is read | 1 | 1 | 1 | 1 | 0 |
| a value of type [Identifier, BindingOrAssignmentElement][] \| undefined seen as {}[] \| undefined, which can write {} where [Identifier, BindingOrAssignmentElement] is read | 1 | 1 | 1 | 1 | 0 |
| a value of type __String[] \| undefined seen as {}[] \| undefined, which can write {} where __String is read | 1 | 1 | 1 | 1 | 0 |
| a value of type any | 21 | 21 | 21 | 21 | 0 |
| a value of type any seen as T, a type parameter whose constraint Node can be written, so it can write what any can't hold | 2 | 2 | 2 | 2 | 0 |
| a value of type false \| FlowNode \| undefined seen as false \| FlowAssignment \| FlowLabel \| FlowReduceLabel \| FlowStart \| FlowSwitchClause \| undefined, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 2 | 2 | 2 | 2 | 0 |
| a value of type never | 1 | 1 | 1 | 1 | 0 |
| a value of type never seen as T, a type parameter whose constraint any can be written, so it can write what never can't hold | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as (SourceMapRange \| undefined)[], which can write SourceMapRange \| undefined where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as BaseType[], which can write BaseType where never is read | 14 | 14 | 14 | 14 | 0 |
| a value of type never[] seen as ChildDirectoryWatcher[], which can write ChildDirectoryWatcher where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as ClassElement[], which can write ClassElement where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as CommentRange[], which can write CommentRange where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as Declaration[], which can write Declaration where never is read | 3 | 3 | 3 | 3 | 0 |
| a value of type never[] seen as DiagnosticArguments, which can write string \| number \| boolean \| readonly string[] \| SourceFile \| undefined where never is read | 3 | 3 | 3 | 3 | 0 |
| a value of type never[] seen as DiagnosticWithLocation[], which can write DiagnosticWithLocation where never is read | 2 | 2 | 2 | 2 | 0 |
| a value of type never[] seen as Diagnostic[], which can write Diagnostic where never is read | 3 | 3 | 3 | 3 | 0 |
| a value of type never[] seen as Identifier[], which can write Identifier where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as IndexInfo[], which can write IndexInfo where never is read | 7 | 7 | 7 | 7 | 0 |
| a value of type never[] seen as Path[], which can write Path where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as ProjectReference[], which can write ProjectReference where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as ResolvedConfigFileName[], which can write ResolvedConfigFileName where never is read | 2 | 2 | 2 | 2 | 0 |
| a value of type never[] seen as ResolvedProjectReference[], which can write ResolvedProjectReference where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as Signature[], which can write Signature where never is read | 8 | 8 | 8 | 8 | 0 |
| a value of type never[] seen as SourceFile[], which can write SourceFile where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as SourceMappedPosition[], which can write SourceMappedPosition where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as StringLiteralLike[], which can write StringLiteralLike where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as Symbol[] \| undefined, which can write Symbol where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as Symbol[], which can write Symbol where never is read | 12 | 12 | 12 | 12 | 0 |
| a value of type never[] seen as TransformerFactory<Bundle \| SourceFile>[], which can write TransformerFactory<Bundle \| SourceFile> where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as TypeParameter[] \| undefined, which can write TypeParameter where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as TypeParameter[], which can write TypeParameter where never is read | 2 | 2 | 2 | 2 | 0 |
| a value of type never[] seen as Type[], which can write Type where never is read | 2 | 2 | 2 | 2 | 0 |
| a value of type never[] seen as VarianceFlags[], which can write VarianceFlags where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as string[] \| never[], which can write string where never is read | 3 | 3 | 3 | 3 | 0 |
| a value of type never[] seen as string[], which can write string where never is read | 5 | 5 | 5 | 5 | 0 |
| a value of type never[] \| SortedArray<DiagnosticWithLocation> seen as Diagnostic[], which can write Diagnostic where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[][] seen as string[][], which can write string where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type number[] seen as string \| (string \| number)[] \| undefined, which can write string \| number where number is read | 1 | 1 | 1 | 1 | 0 |
| a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[] \| undefined, which can write SourceFile \| undefined where SourceFile is read | 2 | 2 | 2 | 2 | 0 |
| a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 3 | 3 | 3 | 3 | 0 |
| a value of type readonly Extension[][] seen as readonly string[][], which can write string where Extension is read | 3 | 3 | 3 | 3 | 0 |
| a value of type readonly IncrementalBundleEmitBuildInfoFileInfo[] \| readonly IncrementalMultiFileEmitBuildInfoFileInfo[] where an array goes | 1 | 1 | 1 | 1 | 0 |
| a value of type readonly TypeParameter[] \| undefined seen as TypeParameterDeclaration[] \| undefined, which can write TypeParameterDeclaration where TypeParameter is read | 1 | 1 | 1 | 1 | 0 |
| a value of type readonly string[] \| undefined seen as RegExp[] \| undefined, which can write RegExp where string is read | 1 | 1 | 1 | 1 | 0 |
| a value of type string \| FileInfo \| { version: string; signature: false; affectsGlobalScope: true \| undefined; impliedFormat: ResolutionMode; } seen as IncrementalMultiFileEmitBuildInfoFileInfo, which can write string \| false \| undefined where string \| undefined is read | 1 | 1 | 1 | 1 | 0 |
| a value of type string \| GeneratedIdentifier \| Identifier seen as string \| ModuleExportName, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type string \| number \| boolean \| string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> seen as CompilerOptionsValue, which can write string \| number where string is read | 1 | 1 | 1 | 1 | 0 |
| a value of type string[] seen as CompilerOptionsValue, which can write string \| number where string is read | 5 | 5 | 5 | 5 | 0 |
| a value of type string[] seen as DiagnosticArguments, which can write string \| number \| boolean \| readonly string[] \| SourceFile \| undefined where string is read | 1 | 1 | 1 | 1 | 0 |
| a value of type string[] seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what string[] can't hold | 1 | 1 | 1 | 1 | 0 |
| a value of type string[] seen as TsConfigSourceFile \| CompilerOptionsValue, which can write string \| number where string is read | 1 | 1 | 1 | 1 | 0 |
| a value of type string[] seen as string \| (string \| number)[] \| undefined, which can write string \| number where string is read | 1 | 1 | 1 | 1 | 0 |
| a value of type string[] seen as {}[], which can write {} where string is read | 1 | 1 | 1 | 1 | 0 |
| a value of type string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] seen as unknown[], which can write unknown where string is read | 2 | 2 | 2 | 2 | 0 |
| a value of type string[] \| undefined seen as {}[] \| undefined, which can write {} where string is read | 7 | 7 | 7 | 7 | 0 |
| a value of type string[][] \| readonly (string \| readonly string[] \| undefined)[] where an array goes | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"CheckMode", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof CheckMode is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"EmitFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof EmitFlags is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"FlowFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof FlowFlags is read | 2 | 2 | 2 | 2 | 0 |
| a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"GeneratedIdentifierFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof GeneratedIdentifierFlags is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"ModifierFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ModifierFlags is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"NodeCheckFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof NodeCheckFlags is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"NodeFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof NodeFlags is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"ObjectFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ObjectFlags is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"RelationComparisonResult", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof RelationComparisonResult is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"ScriptKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ScriptKind is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"SignatureCheckMode", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SignatureCheckMode is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"SignatureFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SignatureFlags is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"SnippetKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SnippetKind is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"SymbolFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SymbolFlags is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"SyntaxKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SyntaxKind is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"TransformFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TransformFlags is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"TypeFacts", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TypeFacts is read | 1 | 1 | 1 | 1 | 0 |
| a value of type typeof import("src/compiler/_namespaces/ts") seen as Record<"TypeFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TypeFlags is read | 1 | 1 | 1 | 1 | 0 |
| a value of type undefined seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what undefined can't hold | 1 | 1 | 1 | 1 | 0 |
| a value of type undefined seen as T, a type parameter whose constraint Declaration can be written, so it can write what undefined can't hold | 1 | 1 | 1 | 1 | 0 |
| a value of type undefined seen as T, a type parameter whose constraint Node can be written, so it can write what undefined can't hold | 5 | 5 | 5 | 5 | 0 |
| a value of type undefined seen as T, a type parameter whose constraint Type can be written, so it can write what undefined can't hold | 2 | 2 | 2 | 2 | 0 |
| a value of type undefined seen as T, a type parameter whose constraint WatchedFileWithIsClosed can be written, so it can write what undefined can't hold | 3 | 3 | 3 | 3 | 0 |
| a value of type { (fileName: string): DiagnosticWithLocation[]; (): Diagnostic[]; } seen as { (): Diagnostic[]; (fileName: string): DiagnosticWithLocation[]; }, which can write Diagnostic where DiagnosticWithLocation is read | 1 | 1 | 1 | 1 | 0 |
| a value of type { affectedFile: SourceFile; emitKind: BuilderFileEmit.Js \| BuilderFileEmit.JsMap \| BuilderFileEmit.JsInlineMap \| BuilderFileEmit.DtsErrors \| ... 6 more ... \| BuilderFileEmit.All; } seen as { affectedFile: Program \| SourceFile \| undefined; emitKind: BuilderFileEmit; }, which can write Program \| SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 | 1 | 0 |
| a value of type { arguments: { name?: string; } & { path: string; }; range: CommentRange; } \| { arguments: { name: string; }; range: CommentRange; } \| { arguments: { factory: string; }; range: CommentRange; } \| ... 5 more ... \| ... seen as ({ arguments: { name?: string; } & { path: string; }; range: CommentRange; } \| { arguments: { name: string; }; range: CommentRange; } \| { arguments: { factory: string; }; range: CommentRange; } \| ... 5 more ... \| ...)[] \| ... 8 more ... \| ..., which can write { name?: string; } & { path: string; } where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type { compilerOptions: CompilerOptions; traceEnabled: boolean; affectingLocations: string[] \| undefined; resultFromCache?: ResolvedModuleWithFailedLookupLocations; ... 9 more ...; host: GetPackageJsonEntrypointsHost; } seen as ModuleResolutionState & { host: GetPackageJsonEntrypointsHost; }, which can write string[] \| undefined where never[] is read | 1 | 1 | 1 | 1 | 0 |
| a value of type { ending: ModuleSpecifierEnding; value: string; }[] seen as { ending: ModuleSpecifierEnding \| undefined; value: string; }[], which can write ModuleSpecifierEnding \| undefined where ModuleSpecifierEnding is read | 1 | 1 | 1 | 1 | 0 |
| a value of type { id: number; flowNode: FlowNode; edges: never[]; text: string; lane: number; endLane: number; level: number; circular: false; } seen as FlowGraphNode, which can write FlowGraphEdge[] where never[] is read | 1 | 1 | 1 | 1 | 0 |
| a value of type { introducesError: boolean; node: Identifier \| PropertyAccessEntityNameExpression; sym?: never; } \| { introducesError: boolean; node: Identifier \| PropertyAccessEntityNameExpression; sym: Symbol \| undefined; } seen as { introducesError: boolean; node: LeftHandSideExpression; }, which can write LeftHandSideExpression where Identifier \| PropertyAccessEntityNameExpression is read | 1 | 1 | 1 | 1 | 0 |
| a value of type { major: number; minor: number; patch: number; prerelease: string; build: string; } seen as { major: string \| number; minor: number; patch: number; prerelease: string \| readonly string[]; build: string \| readonly string[]; }, which can write string \| number where number is read | 1 | 1 | 1 | 1 | 0 |
| a value of type { name: __String; oldSymbol: Symbol; }[] \| undefined seen as {}[] \| undefined, which can write {} where { name: __String; oldSymbol: Symbol; } is read | 1 | 1 | 1 | 1 | 0 |
| a value of type { name: string \| undefined; path: string; }[] seen as AmdDependency[], which can write AmdDependency where { name: string \| undefined; path: string; } is read | 1 | 1 | 1 | 1 | 0 |
| a value of type { noInferenceFallback?: boolean \| undefined; enclosingDeclaration: ModuleDeclaration; enclosingFile: SourceFile \| undefined; flags: NodeBuilderFlags; ... 28 more ...; out: WriterContextOut; } seen as NodeBuilderContext, which can write Node \| undefined where ModuleDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type { prepend?: boolean \| undefined; circular?: boolean \| undefined; path: string; originalPath: undefined; }[] \| undefined seen as readonly ProjectReference[] \| undefined, which can write string \| undefined where undefined is read | 1 | 1 | 1 | 1 | 0 |
| a value of type { referencedName: StringLiteral; name: PropertyName; } \| { referencedName: Identifier; name: ComputedPropertyName; } seen as { referencedName: Expression \| undefined; name: PropertyName \| undefined; }, which can write Expression \| undefined where StringLiteral is read | 1 | 1 | 1 | 1 | 0 |
| a value of type { resolvedFileName: string; originalPath: string \| undefined; extension: string; isExternalLibraryImport: boolean; packageId: PackageId \| undefined; resolvedUsingTsExtension: boolean; } \| undefined seen as ResolvedModuleFull \| undefined, which can write boolean \| undefined where boolean is read | 1 | 1 | 1 | 1 | 0 |
| a value of type { value: never; done: true; } seen as IteratorResult<Mapping, any>, which can write any where never is read | 14 | 14 | 14 | 14 | 0 |
| a value without nominal ancestry seen as BinaryExpressionStateMachine<TOuterState, TState, TResult> | 12 | 12 | 12 | 12 | 0 |
| a void value at the argument site through () seen as number | 1 | 1 | 1 | 1 | 0 |
| a void value at the argument site through [Symbol.iterator]().next()<TReturn> seen as any | 1 | 1 | 1 | 1 | 0 |
| a void-returning function at the assignment site seen as (callback: (...args: any[]) => void, ms: number, ...args: any[]) => any | 1 | 1 | 1 | 1 | 0 |
| a write through a union of ArrowFunction \| FunctionExpression \| GetAccessorDeclaration \| MethodDeclaration \| SetAccessorDeclaration into a slot that holds BinaryExpression \| CallExpression in FlowArrayMutation | 1 | 1 | 1 | 1 | 0 |
| a write through a union of string into a slot that holds false in FilePresenceUnknownOnHost | 1 | 1 | 1 | 1 | 0 |
| a write through a union to a slot (SourceMapRange \| undefined)[] doesn't name | 1 | 1 | 1 | 1 | 0 |
| a write through a union to a slot SyntaxKindFeatureFlags[] doesn't name | 2 | 2 | 2 | 2 | 0 |
| a { all?: boolean; allowImportingTsExtensions?: boolean; allowJs?: boolean; allowNonTsExtensions?: boolean; allowArbitraryExtensions?: boolean; allowSyntheticDefaultImports?: boolean; allowUmdGlobalAccess?: boolean; ... 129 more ...; moduleResolution: ModuleResolutionKind; } seen as CompilerOptions (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a { allowImportingTsExtensions: { dependencies: "rewriteRelativeImportExtensions"[]; computeValue: (compilerOptions: Pick<CompilerOptions, "allowImportingTsExtensions" \| "rewriteRelativeImportExtensions">) => boolean; }; ... 24 more ...; useUnknownInCatchVariables: ...; } seen as Record<string, { dependencies: readonly string[]; computeValue: (options: CompilerOptions) => CompilerOptionsValue; }> (fixed objects and records have different storage; copy explicitly) | 2 | 2 | 2 | 2 | 0 |
| a { config?: any; error?: Diagnostic \| undefined; } seen as { config?: AdamicJsonRecoveryObject; error?: Diagnostic \| undefined; } (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a { host: ModuleResolutionHost; traceEnabled: boolean; failedLookupLocations: string[] \| undefined; affectingLocations: string[] \| undefined; resultFromCache?: ResolvedModuleWithFailedLookupLocations; ... 8 more ...; reportDiagnostic: ...; } seen as ModuleResolutionState (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a {} \| undefined seen as AdamicJsonRecoveryValue \| undefined (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a {} \| undefined seen as TsConfigSourceFile \| CompilerOptionsValue (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| an ESM namespace object used as a value; use qualified live bindings | 19 | 19 | 19 | 19 | 0 |
| an ElementAccessExpression | 1 | 1 | 1 | 1 | 0 |
| an ambient function declaration | 2 | 2 | 2 | 2 | 0 |
| an ambient value declaration | 7 | 7 | 7 | 7 | 0 |
| an arbitrary number or a value from another enum assigned to Extension.Mjs; its members are a closed union | 1 | 1 | 1 | 1 | 0 |
| an array going where a plain object type is held | 2 | 2 | 2 | 2 | 0 |
| an array of unknown with erased element storage (retain its declared element type before reading elements) | 2 | 2 | 2 | 2 | 0 |
| an assignment result with different storage | 1 | 1 | 1 | 1 | 0 |
| an assignment value to a member | 3 | 3 | 3 | 3 | 0 |
| an enum inside a function or block; declare it at module scope | 4 | 4 | 4 | 4 | 0 |
| an indexed mutation without a closed const literal origin | 2 | 2 | 2 | 2 | 0 |
| an object refinement using an open numeric enum as a literal tag | 973 | 973 | 777 | 777 | 0 |
| an object with a nullable field viewed as unknown or object (null and undefined slot tags) | 1 | 1 | 1 | 1 | 0 |
| an optional Map get through a structural receiver | 5 | 5 | 5 | 5 | 0 |
| an optional chain longer than one step | 5 | 5 | 5 | 5 | 0 |
| an unproven predicate argument for parameter callback (argument "decl => isVariableDeclarationInitializedToRequire(decl)") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter callback (argument "isFunctionOrSourceFile") | 3 | 3 | 3 | 3 | 0 |
| an unproven predicate argument for parameter callback (argument "l => isImportTypeNode(l) \|\| isExportDeclaration(l) \|\| isImportDeclaration(l) \|\| isJSDocImportTag(l)") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter callback (argument "n => isExportAssignment(n) \|\| isStatement(n) \|\| isVariableDeclaration(n) \|\| isPropertyDeclaration(n) \|\| isParameter(n)") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter callback (argument "n => isInterfaceDeclaration(n) \|\| isTypeAliasDeclaration(n) \|\| isTypeLiteralNode(n)") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter callback (argument "or(isClassLike, isInterfaceDeclaration)") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter callback (argument "or(isImportDeclaration, isImportEqualsDeclaration)") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter callback (argument "or(isImportDeclaration, isJSDocImportTag, isExportDeclaration)") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter callback (argument "or(isImportEqualsDeclaration, isExportDeclaration)") | 2 | 2 | 2 | 2 | 0 |
| an unproven predicate argument for parameter callback (argument "or(isJSDocLinkLike, isJSDocNameReference, isJSDocMemberName)") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter f (argument "d => isClassDeclaration(d) \|\| isClassExpression(d)") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter f (argument "helper => !helper.scoped") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter f (argument "p => isParameterPropertyDeclaration(p, constructor)") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter nodeTest (argument "(n): n is FunctionDeclaration \| VariableDeclaration => isFunctionDeclaration(n) \|\| isVariableDeclaration(n)") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter nodeTest (argument "isDeclaration") | 2 | 2 | 2 | 2 | 0 |
| an unproven predicate argument for parameter nodeTest (argument "isExportSpecifier") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter nodeTest (argument "isFunctionDeclaration") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter nodeTest (argument "isPropertyAccessExpression") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter predicate (argument "d => isVariableDeclaration(d)") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter predicate (argument "m => m.kind === SyntaxKind.ExportKeyword") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter predicate (argument "modifier => !isExportModifier(modifier) && !isDefaultModifier(modifier)") | 2 | 2 | 2 | 2 | 0 |
| an unproven predicate argument for parameter predicate (argument "or(isPropertyDeclaration, isVariableDeclaration)") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter predicate (argument "predicate") | 2 | 2 | 2 | 2 | 0 |
| an unproven predicate argument for parameter test (argument "createAssignmentCallback ? isIdentifier : isExpression") | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from "" to T: the source is not assignable to the target | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from Declaration to NamedDeclaration: optional field name has no proven compatible presence/type | 2 | 2 | 2 | 2 | 0 |
| an unproven relation from DiagnosticRelatedInformation to Diagnostic: optional field reportsUnnecessary has no proven compatible presence/type | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from Identifier to Identifier: optional field id has no proven compatible presence/type | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from ImportEqualsDeclaration & { moduleReference: ExternalModuleReference; } to ImportEqualsDeclaration: optional field moduleReference.id has no proven compatible presence/type | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from InstantiableType \| UnionOrIntersectionType to TypeParameter: optional field constraint has no proven compatible presence/type | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from LiteralLikeNode to TemplateLiteralLikeNode: optional field rawText has no proven compatible presence/type | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from Node \| undefined to T \| undefined: the source is not assignable to the target | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from NodeArray<BindingElement> \| NodeArray<Expression> \| NodeArray<ArrayBindingElement> to NodeArray<Node>: optional field concat.parameter.element.slice.element.propertyName has no proven compatible presence/type | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from NodeArray<ModifierLike> & readonly Decorator[] to NodeArray<Decorator>: optional field concat.parameter.element.slice.element.parent.name has no proven compatible presence/type | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from ObjectType to AnonymousType: optional field target has no proven compatible presence/type | 2 | 2 | 2 | 2 | 0 |
| an unproven relation from PackageJsonPathFields to PackageJson: optional field version has no proven compatible presence/type | 2 | 2 | 2 | 2 | 0 |
| an unproven relation from SolutionBuilderHost<T> \| SolutionBuilderWithWatchHost<T> to SolutionBuilderHost<T>: optional field reportErrorSummary has no proven compatible presence/type | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from SolutionBuilderHostBase<T> to SolutionBuilderHost<T>: optional field reportErrorSummary has no proven compatible presence/type | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from StructuredType to AnonymousType: optional field target has no proven compatible presence/type | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from StructuredType to ObjectType: optional field members has no proven compatible presence/type | 3 | 3 | 3 | 3 | 0 |
| an unproven relation from Type to SyntheticDefaultModuleType: optional field syntheticType has no proven compatible presence/type | 2 | 2 | 2 | 2 | 0 |
| an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | 26 | 26 | 26 | 26 | 0 |
| an unproven relation from Type to TypeVariable: optional field constraint has no proven compatible presence/type | 2 | 2 | 2 | 2 | 0 |
| an unproven relation from UnionOrIntersectionType to UnionType: optional field resolvedReducedType has no proven compatible presence/type | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from string to P: the source is not assignable to the target | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from string to T: the source is not assignable to the target | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot BuildStep.Done | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot BuilderFileEmit.Js \| BuilderFileEmit.JsMap \| BuilderFileEmit.JsInlineMap \| BuilderFileEmit.DtsErrors \| BuilderFileEmit.DtsEmit \| ... 5 more ... \| BuilderFileEmit.All | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot BuilderFileEmit.None | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot CompilerOptions \| BuilderFileEmit | 5 | 5 | 5 | 5 | 0 |
| an unproven value assigned to a numeric literal or enum member slot Connection | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot ElementFlags.Optional | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot ModifierFlags.None | 5 | 5 | 5 | 5 | 0 |
| an unproven value assigned to a numeric literal or enum member slot ModuleKind.ESNext | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot NodeFlags.Const | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot PunctuationOrKeywordSyntaxKind | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind | 15 | 15 | 15 | 15 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.BindingElement | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Block | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ClassStaticBlockDeclaration | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ComputedPropertyName | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ConditionalType | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.DotDotDotToken | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ExportAssignment | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ExpressionWithTypeArguments | 3 | 3 | 3 | 3 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ExtendsKeyword \| SyntaxKind.ImplementsKeyword | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.FunctionExpression | 2 | 2 | 2 | 2 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Identifier | 5 | 5 | 5 | 5 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ImportAttributes | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ImportDeclaration | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.JSDoc | 2 | 2 | 2 | 2 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.JSDocTypeExpression | 14 | 14 | 14 | 14 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.KeyOfKeyword \| SyntaxKind.ReadonlyKeyword \| SyntaxKind.UniqueKeyword | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.MethodSignature \| SyntaxKind.MethodDeclaration \| SyntaxKind.Constructor \| SyntaxKind.GetAccessor \| SyntaxKind.SetAccessor \| SyntaxKind.CallSignature \| SyntaxKind.ConstructSignature \| ... 6 more ... \| SyntaxKind.JSDocFunctionType | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NamespaceExport | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NamespaceImport | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NoSubstitutionTemplateLiteral \| SyntaxKind.TemplateHead \| SyntaxKind.TemplateMiddle \| SyntaxKind.TemplateTail | 2 | 2 | 2 | 2 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NumericLiteral \| SyntaxKind.BigIntLiteral \| SyntaxKind.StringLiteral \| SyntaxKind.JsxText \| SyntaxKind.JsxTextAllWhiteSpaces \| SyntaxKind.RegularExpressionLiteral \| SyntaxKind.NoSubstitutionTemplateLiteral | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ObjectLiteralExpression | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Parameter | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.PropertyAssignment | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.QuestionToken | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.SourceFile | 8 | 8 | 8 | 8 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.StringLiteral | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.TypeKeyword | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Unknown \| SyntaxKind.NumericLiteral \| SyntaxKind.BigIntLiteral \| SyntaxKind.StringLiteral \| SyntaxKind.JsxText \| SyntaxKind.RegularExpressionLiteral \| SyntaxKind.NoSubstitutionTemplateLiteral | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.VariableDeclaration | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.WithKeyword \| SyntaxKind.AssertKeyword | 2 | 2 | 2 | 2 | 0 |
| an unproven value assigned to a numeric literal or enum member slot Ternary | 2 | 2 | 2 | 2 | 0 |
| an unproven value assigned to a numeric literal or enum member slot Ternary.False | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot VarianceFlags.Contravariant | 1 | 1 | 1 | 1 | 0 |
| an unrestricted dictionary write that can violate named property all; use a literal key or a value accepted by every named property | 1 | 1 | 1 | 1 | 0 |
| an unrestricted dictionary write that can violate named property configFile; use a literal key or a value accepted by every named property | 1 | 1 | 1 | 1 | 0 |
| assigning a field of a value | 2 | 2 | 2 | 2 | 0 |
| assigning an element of a value | 1 | 1 | 1 | 1 | 0 |
| assigning to a NonNullExpression | 1 | 1 | 1 | 1 | 0 |
| assigning to an ObjectLiteralExpression | 1 | 1 | 1 | 1 | 0 |
| checked predicate membership over an open tag domain | 34 | 34 | 34 | 34 | 0 |
| checked predicate overload target AssignmentExpression<AmpersandAmpersandEqualsToken \| BarBarEqualsToken \| EqualsToken \| QuestionQuestionEqualsToken> & ... | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target BinaryExpression & { operatorToken: Token<SyntaxKind.CommaToken>; } | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target BindableStaticElementAccessExpression | 2 | 2 | 2 | 2 | 0 |
| checked predicate overload target ClassNamedEvaluationHelperBlock | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target ClassThisAssignmentBlock | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target ElementAccessExpression & Declaration & { readonly argumentExpression: NumericLiteral \| StringLiteralLike; } & ... | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target ImportClause & { readonly isTypeOnly: true; readonly name: Identifier; } | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target ImportEqualsDeclaration & { moduleReference: ExternalModuleReference; } | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target JSDocTag | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target LiteralImportTypeNode | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target LiteralLikeElementAccessExpression | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target NamedDeclaration & { name: DeclarationName; } | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target NonIncrementalBuildInfo | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target ParameterPropertyDeclaration | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target PrefixUnaryExpression & { operand: NumericLiteral; } | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target PrefixUnaryExpression & { operator: SyntaxKind.PlusToken; operand: NumericLiteral; } | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target ReferencedFile | 1 | 1 | 1 | 1 | 0 |
| checked view field modifiers of type NodeArray<Modifier> \| NodeArray<ModifierLike> \| undefined | 1 | 1 | 1 | 1 | 0 |
| checked view union arm awaits views-v3: array element kind | 1 | 1 | 1 | 1 | 0 |
| delete | 1 | 1 | 1 | 1 | 0 |
| destructuring a value | 2 | 2 | 2 | 2 | 0 |
| destructuring anything but a tuple into [names] | 2 | 2 | 2 | 2 | 0 |
| detached hasOwnProperty through an object view (an Error may be hidden by the view) | 1 | 1 | 1 | 1 | 0 |
| for...in over an array (holes and own enumerable properties are not represented; use for...of for elements) | 1 | 1 | 1 | 1 | 0 |
| for...of over an object | 1 | 1 | 1 | 1 | 0 |
| function identity comparison in a program with specialized generic function values | 1 | 1 | 1 | 1 | 0 |
| in when a possibly absent spread has no property presence descriptors | 3 | 3 | 3 | 3 | 0 |
| indexOf on a tuple (a tuple is held as an object, not an array, so far; write it as an array where it's made) | 1 | 1 | 1 | 1 | 0 |
| inherited library member add read as an own field | 1 | 1 | 1 | 1 | 0 |
| inherited library member compare read as an own field | 1 | 1 | 1 | 1 | 0 |
| inherited library member get read as an own field | 4 | 4 | 4 | 4 | 0 |
| inherited library member has read as an own field | 2 | 2 | 2 | 2 | 0 |
| inherited library member now read as an own field | 1 | 1 | 1 | 1 | 0 |
| inherited library member push read as an own field | 2 | 2 | 2 | 2 | 0 |
| instanceof against a value that isn't a declared class | 1 | 1 | 1 | 1 | 0 |
| new a ParenthesizedExpression | 3 | 3 | 3 | 3 | 0 |
| new an Identifier | 8 | 8 | 8 | 8 | 0 |
| node:fs.watch | 2 | 2 | 1 | 1 | 0 |
| node:inspector.CallFrame.url | 8 | 8 | 4 | 4 | 0 |
| node:inspector.Profile.nodes | 2 | 2 | 1 | 1 | 0 |
| node:inspector.ProfileNode.callFrame | 8 | 8 | 4 | 4 | 0 |
| node:inspector.Session | 5 | 5 | 3 | 3 | 0 |
| node:inspector.Session.connect | 2 | 2 | 1 | 1 | 0 |
| node:inspector.Session.disconnect | 2 | 2 | 1 | 1 | 0 |
| node:inspector.Session.post | 6 | 6 | 3 | 3 | 0 |
| node:module.__filename and __dirname: CommonJS module values with no ES-module or reproducible native answer yet (fix: pass the module's path in from where the program is started) | 4 | 4 | 4 | 4 | 0 |
| node:module.require | 1 | 1 | 1 | 1 | 0 |
| node:process.Process.stdout used as a value | 1 | 1 | 1 | 1 | 0 |
| optional chaining to .size on a value | 4 | 4 | 4 | 4 | 0 |
| optional property Low in Partial<Levels> absent from structural source {}, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property _propertyAccessExpressionLikeQualifiedNameBrand in PropertyAccessEntityNameExpression absent from structural source never, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property all in CompilerOptions absent from structural source {}, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property constraint in TypeParameter absent from structural source InterfaceType, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property constraint in TypeParameter absent from structural source Type, which can hide fields | 14 | 14 | 14 | 14 | 0 |
| optional property createProgram in IncrementalProgramOptions<BuilderProgram> absent from structural source IncrementalCompilationOptions, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property createProgram in IncrementalProgramOptions<EmitAndSemanticDiagnosticsBuilderProgram> absent from structural source IncrementalCompilationOptions, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property details in FileReasonToChainCache absent from structural source { fileIncludeReasonDetails: DiagnosticMessageChain \| undefined; redirectInfo: DiagnosticMessageChain[] \| undefined; }, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property equalsToken in ShorthandPropertyAssignment absent from structural source never, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property failedLookupLocations in ResolvedModuleWithFailedLookupLocations absent from structural source { resolvedModule: ResolvedModuleFull; }, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property failedLookupLocations in ResolvedTypeReferenceDirectiveWithFailedLookupLocations absent from structural source { resolvedTypeReferenceDirective: ResolvedTypeReferenceDirective \| undefined; }, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property getGlobalTypingsCacheLocation in ModuleResolutionHost absent from structural source SolutionBuilderHost<T>, which can hide fields | 3 | 3 | 3 | 3 | 0 |
| optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields | 36 | 36 | 36 | 36 | 0 |
| optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceMapSource, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property id in ArrayBindingPattern absent from structural source never, which can hide fields | 18 | 18 | 18 | 18 | 0 |
| optional property id in AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; } absent from structural source never, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property id in BigIntLiteral absent from structural source never, which can hide fields | 10 | 10 | 10 | 10 | 0 |
| optional property id in BinaryExpression absent from structural source never, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property id in Expression absent from structural source never, which can hide fields | 6 | 6 | 6 | 6 | 0 |
| optional property id in ExternalModuleReference absent from structural source never, which can hide fields | 7 | 7 | 7 | 7 | 0 |
| optional property id in FalseLiteral absent from structural source never, which can hide fields | 3 | 3 | 3 | 3 | 0 |
| optional property id in Identifier absent from structural source never, which can hide fields | 6 | 6 | 6 | 6 | 0 |
| optional property id in JSDocThisTag absent from structural source never, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property id in LateBoundName absent from structural source never, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property id in LiteralExpression & StringLiteral absent from structural source never, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property id in NamedExports absent from structural source never, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property id in Node absent from structural source never, which can hide fields | 63 | 63 | 63 | 63 | 0 |
| optional property id in PrivateIdentifier absent from structural source never, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property isInvalidated in CachedResolvedModuleWithFailedLookupLocations absent from structural source ResolvedModuleWithFailedLookupLocations, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property isInvalidated in CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations absent from structural source ResolvedTypeReferenceDirectiveWithFailedLookupLocations, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property isInvalidated in ResolutionWithFailedLookupLocations absent from structural source ResolvedModuleWithFailedLookupLocations, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property isInvalidated in ResolutionWithFailedLookupLocations absent from structural source ResolvedTypeReferenceDirectiveWithFailedLookupLocations, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property jsDocCache in JSDocArray absent from structural source JSDoc[], which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property localSymbol in Declaration absent from structural source never, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property modifiers in ExportAssignment absent from structural source never, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property modifiers in GetAccessorDeclaration absent from structural source never, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property modifiers in PropertyDeclaration absent from structural source never, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property modifiers in SetAccessorDeclaration absent from structural source never, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property onUnRecoverableConfigFileDiagnostic in CompilerHost & { onUnRecoverableConfigFileDiagnostic?: DiagnosticReporter; } absent from structural source CompilerHost, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property onUnRecoverableConfigFileDiagnostic in ProgramHost<T> & { onUnRecoverableConfigFileDiagnostic?: DiagnosticReporter; } absent from structural source SolutionBuilderHost<T>, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property originalPath in Resolved absent from structural source { path: string; extension: string; packageId: PackageId \| undefined; resolvedUsingTsExtension: boolean \| undefined; }, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property packageJsonScope in Pick<SourceFile, "fileName" \| "impliedNodeFormat" \| "packageJsonScope"> absent from structural source Pick<SourceFile, "fileName" \| "impliedNodeFormat">, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property packageRootPath in { moduleFileToTry: string; packageRootPath?: string; blockedByExports?: true; verbatimFromExports?: true; } absent from structural source { moduleFileToTry: string; }, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property preserve in FileReference absent from structural source { resolutionMode: ModuleKind.CommonJS \| ModuleKind.ESNext; }, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property repopulateInfo in DiagnosticMessageChain absent from structural source SerializedDiagnosticMessageChain, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property reportsUnnecessary in Diagnostic absent from structural source DiagnosticRelatedInformation, which can hide fields | 5 | 5 | 5 | 5 | 0 |
| optional property reportsUnnecessary in ReusableDiagnostic absent from structural source ReusableDiagnosticRelatedInformation, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property resolutionMode in FileReference absent from structural source { preserve: true; }, which can hide fields | 3 | 3 | 3 | 3 | 0 |
| optional property skipLogging in ErrorOutputContainer absent from structural source { errors?: Diagnostic[] \| undefined; }, which can hide fields | 6 | 6 | 6 | 6 | 0 |
| optional property skipTrivia in SourceMapSource absent from structural source SourceFile, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property source in SourceMapRange absent from structural source TextRange, which can hide fields | 23 | 23 | 23 | 23 | 0 |
| optional property templateFlags in NoSubstitutionTemplateLiteral absent from structural source never, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property textSourceNode in StringLiteral absent from structural source never, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property types in { types?: { value: string; pos: number; end: number; }; } & { lib?: { value: string; pos: number; end: number; }; } & { path?: { value: string; pos: number; end: number; }; } & { "no-default-lib"?: string; } & { "resolution-mode"?: string; } & ... absent from structural source { [index: string]: string \| undefined; }, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property types in { types?: { value: string; pos: number; end: number; }; } & { lib?: { value: string; pos: number; end: number; }; } & { path?: { value: string; pos: number; end: number; }; } & { "no-default-lib"?: string; } & { "resolution-mode"?: string; } & ... absent from structural source { [index: string]: string \| { value: string; pos: number; end: number; }; }, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property version in PackageJson absent from structural source PackageJsonPathFields, which can hide fields | 5 | 5 | 5 | 5 | 0 |
| optional property watchFile in WatchOptions absent from structural source {}, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property watchOptions in TSConfig & { watchOptions?: object \| undefined; } absent from structural source { include: readonly string[] \| undefined; exclude: readonly string[] \| undefined; }, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property watcher in ParsedConfig absent from structural source { parsedCommandLine: ParsedCommandLine \| undefined; }, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional record property access | 3 | 3 | 3 | 3 | 0 |
| overload 1 of arrayFrom result U[] cannot be served by implementation result (T \| U)[] | 1 | 1 | 1 | 1 | 0 |
| overload 1 of assertEachNode parameter test cannot be served by implementation parameter test | 1 | 1 | 1 | 1 | 0 |
| overload 1 of assertNode parameter test cannot be served by implementation parameter test | 1 | 1 | 1 | 1 | 0 |
| overload 1 of captureMapping result Required<Mapping> cannot be served by implementation result Mapping | 1 | 1 | 1 | 1 | 0 |
| overload 1 of concatenate result T[] cannot be served by implementation result readonly T[] \| undefined | 1 | 1 | 1 | 1 | 0 |
| overload 1 of convertOptionsFromJson result WatchOptions \| undefined cannot be served by implementation result CompilerOptions \| TypeAcquisition \| WatchOptions \| undefined | 1 | 1 | 1 | 1 | 0 |
| overload 1 of convertOrRepopulateDiagnosticMessageChainArray parameter array cannot be served by implementation parameter array | 1 | 1 | 1 | 1 | 0 |
| overload 1 of createBuilderProgram result SemanticDiagnosticsBuilderProgram cannot be served by implementation result BuilderProgram \| undefined | 1 | 1 | 1 | 1 | 0 |
| overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined | 1 | 1 | 1 | 1 | 0 |
| overload 1 of find result U \| undefined cannot be served by implementation result T \| undefined | 1 | 1 | 1 | 1 | 0 |
| overload 1 of findAncestor result T \| undefined cannot be served by implementation result Node \| undefined | 1 | 1 | 1 | 1 | 0 |
| overload 1 of getPathComponents result PathPathComponents cannot be served by implementation result string[] | 1 | 1 | 1 | 1 | 0 |
| overload 1 of getProperties result readonly InitializedPropertyDeclaration[] cannot be served by implementation result readonly PropertyDeclaration[] | 1 | 1 | 1 | 1 | 0 |
| overload 1 of getSuperContainer result SuperContainer \| undefined cannot be served by implementation result SuperContainerOrFunctions \| undefined | 1 | 1 | 1 | 1 | 0 |
| overload 1 of getSupportedExtensions result readonly Extension[][] cannot be served by implementation result readonly string[][] | 1 | 1 | 1 | 1 | 0 |
| overload 1 of getSupportedExtensionsWithJsonIfResolveJsonModule parameter supportedExtensions cannot be served by implementation parameter supportedExtensions | 1 | 1 | 1 | 1 | 0 |
| overload 1 of getThisContainer result ThisContainer cannot be served by implementation result ArrowFunction \| ComputedPropertyName \| ThisContainer | 1 | 1 | 1 | 1 | 0 |
| overload 1 of injectClassNamedEvaluationHelperBlockIfMissing result Extract<ClassDeclaration, Pick<...>> \| Extract<...> cannot be served by implementation result ClassLikeDeclaration | 1 | 1 | 1 | 1 | 0 |
| overload 1 of nodeCanBeDecorated result true cannot be served by implementation result boolean | 1 | 1 | 1 | 1 | 0 |
| overload 1 of or type parameter P cannot satisfy implementation parameter T | 1 | 1 | 1 | 1 | 0 |
| overload 1 of parenthesizeConciseBodyOfArrowFunction result Expression cannot be served by implementation result ConciseBody | 1 | 1 | 1 | 1 | 0 |
| overload 1 of sameMap result U[] cannot be served by implementation result readonly U[] \| undefined | 1 | 1 | 1 | 1 | 0 |
| overload 1 of skipOuterExpressions result T cannot be served by implementation result Node | 1 | 1 | 1 | 1 | 0 |
| overload 1 of skipParentheses result Expression cannot be served by implementation result Node | 1 | 1 | 1 | 1 | 0 |
| overload 1 of skipPartiallyEmittedExpressions result Expression cannot be served by implementation result Node | 1 | 1 | 1 | 1 | 0 |
| overload 1 of transformAsyncFunctionBody result Block cannot be served by implementation result ConciseBody | 1 | 1 | 1 | 1 | 0 |
| overload 1 of transformFunctionBody result Block cannot be served by implementation result ConciseBody | 1 | 1 | 1 | 1 | 0 |
| overload 1 of transformNamedEvaluation result Extract<AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; }, Pick<...>> \| ... 7 more ... \| Extract<...> cannot be served by implementation result BinaryExpression \| BindingElement \| ExportAssignment \| ParameterDeclaration \| PropertyAssignment \| PropertyDeclaration \| ShorthandPropertyAssignment \| VariableDeclaration | 1 | 1 | 1 | 1 | 0 |
| overload 1 of visitArray parameter visitor cannot be served by implementation parameter visitor | 1 | 1 | 1 | 1 | 0 |
| overload 1 of visitArrayWorker parameter visitor cannot be served by implementation parameter visitor | 1 | 1 | 1 | 1 | 0 |
| overload 1 of visitFunctionBody result Block cannot be served by implementation result ConciseBody \| undefined | 1 | 1 | 1 | 1 | 0 |
| overload 1 of visitNode parameter node cannot be served by implementation parameter node | 1 | 1 | 1 | 1 | 0 |
| overload 1 of visitNodes parameter visitor cannot be served by implementation parameter visitor | 1 | 1 | 1 | 1 | 0 |
| overload 1 of visitParameterList parameter nodesVisitor cannot be served by implementation parameter nodesVisitor | 1 | 1 | 1 | 1 | 0 |
| overload 2 of arrayToMap result Map<K, V2> cannot be served by implementation result Map<K, V1 \| V2> | 1 | 1 | 1 | 1 | 0 |
| overload 2 of getParseTreeNode result T \| undefined cannot be served by implementation result Node \| undefined | 1 | 1 | 1 | 1 | 0 |
| overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor | 1 | 1 | 1 | 1 | 0 |
| reading require | 1 | 1 | 1 | 1 | 0 |
| replacing a represented method at runtime | 2 | 2 | 2 | 2 | 0 |
| slice with an index that isn't a number | 1 | 1 | 1 | 1 | 0 |
| spreading an array of other elements | 1 | 1 | 1 | 1 | 0 |
| this outside a method | 1 | 1 | 1 | 1 | 0 |
| toLocaleTimeString: Date methods other than getTime/valueOf in the fs file host | 1 | 1 | 1 | 1 | 0 |
| var | 2 | 2 | 2 | 2 | 0 |
| writing a checked object field without its source-slot type certificate | 17 | 17 | 17 | 17 | 0 |
| writing a possibly absent optional own field | 4 | 4 | 4 | 4 | 0 |

## Raw Findings

Moved by reason:


New findings:

- `src/compiler/commandLineParser.ts:2278:82`: Refused: a member typed any (config, declared): any has no proven type, so every read of it through a typed view is unchecked

Units whose first stop moved:

None.

Panics: 0 before, 0 after.

## Distinct Reads

Moved by reason:


New findings:

- `src/compiler/commandLineParser.ts:2278:82`: Refused: a member typed any (config, declared): any has no proven type, so every read of it through a typed view is unchecked

Units whose first stop moved:

None.

Panics: 0 before, 0 after.

The JSON retains every lowered site, move, new finding, unit stop and panic.
