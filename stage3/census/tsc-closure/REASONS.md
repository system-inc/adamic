# Exact reason counts

Unique `(kind, where, reason, text)` sites; dependencies, errors and panics remain separately labelled.

## latent / all

| Kind and exact reason | Sites |
| --- | ---: |
| Refused: a cast the runtime can't check | 1396 |
| Refused: an object refinement using an open numeric enum as a literal tag | 1182 |
| Refused: the non-null assertion ! | 767 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on node) | 388 |
| NotYet: a function without a body | 191 |
| NotYet: a PrefixUnaryExpression on a value | 110 |
| NotYet: a method call through a structural signature in a program with statics; use typeof the declaring class | 96 |
| NotYet: a NonNullExpression | 71 |
| NotYet: a function returning T \| undefined | 71 |
| Refused: a type predicate whose return is not proven (there is no body proving this parameter) | 71 |
| NotYet: reading Debug | 55 |
| NotYet: a function returning T | 53 |
| NotYet: a BinaryExpression with a value and a boolean | 40 |
| NotYet: a BinaryExpression with a value and a value | 40 |
| Refused: the comma operator | 38 |
| Refused: a type predicate whose return is not proven (return paths through KindSwitchStatement are not verified) | 36 |
| NotYet: a field of type boolean \| undefined | 33 |
| NotYet: a PrefixUnaryExpression on a number | 31 |
| NotYet: a value of type T | 30 |
| NotYet: a value of type any | 30 |
| Refused: a value as a condition | 29 |
| Refused: a value of type SourceFile seen as SourceFileLike, which can write readonly number[] \| undefined where readonly number[] is read | 28 |
| Refused: \|\|= | 25 |
| NotYet: a parameter that isn't a plain name | 22 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on kind) | 22 |
| NotYet: a value of type Path | 21 |
| NotYet: a value of type __String | 19 |
| NotYet: a value of type ResolvedConfigFilePath | 18 |
| NotYet: a function returning U \| undefined | 17 |
| Refused: a string as a condition | 17 |
| Refused: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 16 |
| Refused: a boolean \| undefined as a condition | 15 |
| NotYet: a generic function as a value | 14 |
| NotYet: a value of type unknown | 14 |
| Refused: a method read as a value (liftToBlock would lose its object, and this with it) | 14 |
| Refused: a type predicate whose return is not proven (branch is not a trusted parameter check) | 14 |
| Refused: a value of type FlowNode seen as FlowNode \| undefined, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 14 |
| Refused: a value of type { value: never; done: true; } seen as IteratorResult<Mapping, any>, which can write any where never is read | 14 |
| Refused: a value of type Node seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 13 |
| Refused: a method in object destructuring | 12 |
| Refused: a method read as a value (parenthesizeExpressionForDisallowedComma would lose its object, and this with it) | 12 |
| Refused: a number as a condition | 12 |
| Refused: a value without nominal ancestry seen as BinaryExpressionStateMachine<TOuterState, TState, TResult> | 12 |
| NotYet: a PrefixUnaryExpression on a string | 11 |
| NotYet: a call through ?. (an optional call) | 11 |
| NotYet: a value of type T \| undefined | 11 |
| Refused: a definite assignment assertion ! | 11 |
| Refused: a namespace | 11 |
| Refused: a value of type T seen as Mutable<T>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of T["pos"] would replace | 11 |
| Refused: a value of type never[] seen as string[], which can write string where never is read | 11 |
| Refused: an index signature | 11 |
| NotYet: a BinaryExpression with a boolean and a value | 10 |
| NotYet: a function returning __String | 10 |
| Refused: a value of type BuilderProgramStateWithDefinedProgram seen as BuilderProgramState, which can write Program \| undefined where Program is read | 10 |
| Refused: a value of type Map<string, never[]> seen as Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>, which can write string[] where never[] is read | 10 |
| Refused: optional property id in Node absent from structural source never, which can hide fields | 10 |
| NotYet: a ModuleDeclaration | 9 |
| NotYet: an array of T | 9 |
| NotYet: an enum inside a function or block; declare it at module scope | 9 |
| NotYet: for...of over an object | 9 |
| Refused: a function taking Identifier seen as one taking Identifier \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 9 |
| Refused: a function taking boolean seen as one taking boolean \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 9 |
| Refused: a value of type BuilderProgramState seen as ReusableBuilderProgramState, which can write Map<Path, readonly Diagnostic[] \| readonly ReusableDiagnostic[]> where Map<Path, readonly Diagnostic[]> is read | 9 |
| Refused: optional property source in SourceMapRange absent from structural source TextRange, which can hide fields | 9 |
| NotYet: new an Identifier | 8 |
| NotYet: a BinaryExpression as a statement | 7 |
| NotYet: a BinaryExpression with a number and a number | 7 |
| NotYet: a BinaryExpression with a number \| undefined and a number | 7 |
| NotYet: a BinaryExpression with a string and a string | 7 |
| NotYet: a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions | 7 |
| NotYet: a value of type ResolvedConfigFileName | 7 |
| NotYet: this outside a method | 7 |
| Refused: a method read as a value (parenthesizeLeftSideOfAccess would lose its object, and this with it) | 7 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.JSDocTypeExpression | 7 |
| Refused: optional property id in ArrayBindingPattern absent from structural source never, which can hide fields | 7 |
| Refused: optional property id in BigIntLiteral absent from structural source never, which can hide fields | 7 |
| NotYet: a PrefixUnaryExpression on a number \| undefined | 6 |
| NotYet: a field of type true \| undefined | 6 |
| NotYet: a function returning any | 6 |
| NotYet: a value of type object | 6 |
| NotYet: an array of never | 6 |
| NotYet: regex replacement other than a string | 6 |
| Refused: a function taking BinaryExpressionStateMachine<TOuterState, TState, TResult> seen as one taking BinaryExpressionStateMachine<TOuterState, TState, TResult> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 6 |
| Refused: a method read as a value (parenthesizeOperandOfPrefixUnary would lose its object, and this with it) | 6 |
| Refused: a method read as a value (readFile would lose its object, and this with it) | 6 |
| Refused: a parameter property | 6 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on token) | 6 |
| Refused: a value of type NonNullable<T> seen as T, a type parameter whose constraint any can be written, so it can write what NonNullable<T> can't hold | 6 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint Node can be written, so it can write what undefined can't hold | 6 |
| Refused: var | 6 |
| NotYet: a BinaryExpression with a number and a boolean | 5 |
| NotYet: a BinaryExpression with a value and a number | 5 |
| NotYet: a PrefixUnaryExpression on a boolean \| undefined | 5 |
| NotYet: a function returning Path | 5 |
| NotYet: for...in without a proven fixed plain-object origin (arrays, prototypes and absent synthetic fields cannot be enumerated soundly) | 5 |
| NotYet: reading performance | 5 |
| NotYet: reading ts | 5 |
| Refused: a generator function | 5 |
| Refused: a method read as a value (trace would lose its object, and this with it) | 5 |
| Refused: a value of type FlowArrayMutation \| FlowAssignment seen as FlowNode, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 5 |
| Refused: a value of type FlowNode seen as FlowNode[] \| FlowNode \| undefined, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 5 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier \| Node seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 5 |
| Refused: a value of type Node seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 5 |
| Refused: a value of type never[] seen as Diagnostic[], which can write Diagnostic where never is read | 5 |
| Refused: an unproven value assigned to a numeric literal or enum member slot CompilerOptions \| BuilderFileEmit | 5 |
| Refused: optional property all in CompilerOptions absent from structural source BuildOptions, which can hide fields | 5 |
| Refused: optional property id in Expression absent from structural source never, which can hide fields | 5 |
| Refused: optional property id in ExternalModuleReference absent from structural source never, which can hide fields | 5 |
| Refused: optional property version in PackageJson absent from structural source PackageJsonPathFields, which can hide fields | 5 |
| Refused: yield (generators) | 5 |
| NotYet: a BinaryExpression with a string and a boolean | 4 |
| NotYet: a computed field name | 4 |
| NotYet: a field of type string \| DiagnosticMessageChain | 4 |
| NotYet: a function returning PackageJson[K] \| undefined | 4 |
| NotYet: a function returning __String \| undefined | 4 |
| NotYet: a function with an optional or rest parameter, as a value | 4 |
| NotYet: a value of type NonNullable<T> | 4 |
| NotYet: a value of type T \| Program | 4 |
| NotYet: an ElementAccessExpression | 4 |
| NotYet: reading Parser | 4 |
| Refused: a function taking readonly Modifier[] \| undefined seen as one taking readonly ModifierLike[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 4 |
| Refused: a method read as a value (fileExists would lose its object, and this with it) | 4 |
| Refused: a method read as a value (getSourceFile would lose its object, and this with it) | 4 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on expr) | 4 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on f) | 4 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on n) | 4 |
| Refused: a value of type CompilerOptionsValue seen as TsConfigSourceFile \| CompilerOptionsValue, which can write string \| number where string is read | 4 |
| Refused: a value of type DiagnosticWithDetachedLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where undefined is read | 4 |
| Refused: a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where SourceFile is read | 4 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 4 |
| Refused: a value of type GeneratedIdentifier \| Identifier seen as Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 4 |
| Refused: a value of type Identifier[][] seen as ModuleExportName[][], which can write ModuleExportName[] where Identifier[] is read | 4 |
| Refused: a value of type Mutable<GeneratedIdentifier> seen as Identifier, which can write EmitNode \| undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 4 |
| Refused: a value of type NodeArray<Statement> seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 4 |
| Refused: a value of type string[] seen as CompilerOptionsValue, which can write string \| number where string is read | 4 |
| Refused: an arbitrary number or a value from another enum assigned to InternalSymbolName.Call; its members are a closed union | 4 |
| Refused: in | 4 |
| Refused: optional property id in Identifier absent from structural source never, which can hide fields | 4 |
| Refused: the void operator | 4 |
| NotYet: Object.entries on a shape not proven by a plain literal or its const binding | 3 |
| NotYet: a case whose type differs from the switch's | 3 |
| NotYet: a declaration directly in a case (wrap the case in a block) | 3 |
| NotYet: a function inside a function (a closure) | 3 |
| NotYet: a function returning CompilerOptionsValue | 3 |
| NotYet: a function returning ResolvedConfigFileName | 3 |
| NotYet: a function returning T \| T[] \| undefined | 3 |
| NotYet: a function returning T \| readonly T[] \| undefined | 3 |
| NotYet: a function value returning union of differently held members | 3 |
| NotYet: a value of type CompilerOptionsValue | 3 |
| NotYet: a value of type K | 3 |
| NotYet: a value of type string \| (void & { __escapedIdentifier: void; }) \| (string & { __escapedIdentifier: void; }) | 3 |
| NotYet: a value of type string \| null \| undefined | 3 |
| NotYet: an array of U | 3 |
| NotYet: optional chaining to .size on a value | 3 |
| Refused: a function taking () => T seen as one taking () => T (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 |
| Refused: a function taking [fileName: string, text: string, writeByteOrderMark: boolean, onError?: ((message: string) => void) \| undefined, sourceFiles?: readonly SourceFile[] \| undefined, data?: WriteFileCallbackData \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 |
| Refused: a function taking number seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 |
| Refused: a function taking string seen as one taking [fileName: string, text: string, writeByteOrderMark: boolean, onError?: ((message: string) => void) \| undefined, sourceFiles?: readonly SourceFile[] \| undefined, data?: WriteFileCallbackData \| undefined] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 |
| Refused: a method read as a value (afterProgramCreate would lose its object, and this with it) | 3 |
| Refused: a method read as a value (afterProgramEmitAndDiagnostics would lose its object, and this with it) | 3 |
| Refused: a method read as a value (getCanonicalFileName would lose its object, and this with it) | 3 |
| Refused: a method read as a value (now would lose its object, and this with it) | 3 |
| Refused: a method read as a value (onWatchStatusChange would lose its object, and this with it) | 3 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on element) | 3 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on entry) | 3 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on info) | 3 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on member) | 3 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on tag) | 3 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on value) | 3 |
| Refused: a union of differently held members as a condition | 3 |
| Refused: a value of type BindingOrAssignmentElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type DiagnosticWithLocation[] seen as Diagnostic[], which can write Diagnostic where DiagnosticWithLocation is read | 3 |
| Refused: a value of type Expression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type FutureSourceFile \| SourceFile seen as Pick<SourceFile, "fileName" \| "impliedNodeFormat">, whose readonly field fileName becomes writable: a readonly field may hold something narrower than string, which a write of string would replace | 3 |
| Refused: a value of type Identifier seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type Map<string, string[] \| never[]> seen as Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>, which can write string where never is read | 3 |
| Refused: a value of type SourceFile seen as SourceFile \| SourceFileLike, which can write readonly number[] \| undefined where readonly number[] is read | 3 |
| Refused: a value of type never[] seen as DiagnosticArguments, which can write string \| number \| boolean \| readonly string[] \| SourceFile \| undefined where never is read | 3 |
| Refused: a value of type never[] seen as ResolvedConfigFileName[], which can write ResolvedConfigFileName where never is read | 3 |
| Refused: a value of type never[] seen as string[] \| never[], which can write string where never is read | 3 |
| Refused: a value of type readonly Extension[][] seen as readonly string[][], which can write string where Extension is read | 3 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint WatchedFileWithIsClosed can be written, so it can write what undefined can't hold | 3 |
| Refused: an arbitrary number or a value from another enum assigned to Extension.Ts; its members are a closed union | 3 |
| Refused: optional property rawText in TemplateLiteralLikeNode absent from structural source NumericLiteral, which can hide fields | 3 |
| Refused: optional property reportsUnnecessary in Diagnostic absent from structural source DiagnosticRelatedInformation, which can hide fields | 3 |
| Refused: optional property resolutionMode in FileReference absent from structural source { preserve: true; }, which can hide fields | 3 |
| SkippedDependency: a dependency function whose body has checker diagnostics (measurement skipped) | 3 |
| NotYet: RegExp with a nonconstant pattern | 2 |
| NotYet: a BinaryExpression with a boolean and a string | 2 |
| NotYet: a BinaryExpression with a string and a number | 2 |
| NotYet: a BinaryExpression with a value and a string | 2 |
| NotYet: a PrefixUnaryExpression on a union of differently held members | 2 |
| NotYet: a call returning void \| undefined | 2 |
| NotYet: a field of type "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number> | 2 |
| NotYet: a field of type boolean \| (() => boolean) \| undefined | 2 |
| NotYet: a field of type true \| Node \| undefined | 2 |
| NotYet: a function returning Extract<ClassDeclaration, Pick<...>> \| Extract<...> | 2 |
| NotYet: a function returning Path \| undefined | 2 |
| NotYet: a function returning U | 2 |
| NotYet: a function returning V | 2 |
| NotYet: a function returning object | 2 |
| NotYet: a function returning undefined | 2 |
| NotYet: a function value returning boolean \| undefined | 2 |
| NotYet: a tagged template other than the intrinsic String.raw | 2 |
| NotYet: a value of type (EmitNode & { autoGenerate: AutoGenerateInfo; }) \| (EmitNode & { autoGenerate: AutoGenerateInfo; }) | 2 |
| NotYet: a value of type BindableStaticNameExpression | 2 |
| NotYet: a value of type T \| T[] | 2 |
| NotYet: a value of type TInArray | 2 |
| NotYet: a value of type WrappedExpression<AnonymousFunctionDefinition> | 2 |
| NotYet: new a ParenthesizedExpression | 2 |
| NotYet: reading getOptionsNameMap | 2 |
| NotYet: reading tracingEnabled | 2 |
| NotYet: replacing a represented method at runtime | 2 |
| Refused: a function taking ((node: Node) => VisitResult<Node>) \| undefined seen as one taking ((node: Node) => VisitResult<Node \| undefined>) \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking BinaryOperator seen as one taking SyntaxKind (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking Expression seen as one taking Expression \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking Expression[] seen as one taking readonly Expression[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking GeneratedIdentifierFlags seen as one taking GeneratedIdentifierFlags \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking NodeFlags seen as one taking NodeFlags \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking T seen as one taking Expression (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking [fileName: string, languageVersionOrOptions: CreateSourceFileOptions \| ScriptTarget, onError?: ((message: string) => void) \| undefined, shouldCreateNewSourceFile?: boolean \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking string seen as one taking [fileName: string] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking string seen as one taking any[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a method read as a value (clearTimeout would lose its object, and this with it) | 2 |
| Refused: a method read as a value (createDirectory would lose its object, and this with it) | 2 |
| Refused: a method read as a value (emitNodeWithNotification would lose its object, and this with it) | 2 |
| Refused: a method read as a value (enableCPUProfiler would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getBuildInfo would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getEnvironmentVariable would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getGlobalTypingsCacheLocation would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getParsedCommandLine would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getSemanticDiagnosticsOfNextAffectedFile would lose its object, and this with it) | 2 |
| Refused: a method read as a value (hasGlobalName would lose its object, and this with it) | 2 |
| Refused: a method read as a value (isEmitNotificationEnabled would lose its object, and this with it) | 2 |
| Refused: a method read as a value (parenthesizeBranchOfConditionalExpression would lose its object, and this with it) | 2 |
| Refused: a method read as a value (parenthesizeConstituentTypesOfIntersectionType would lose its object, and this with it) | 2 |
| Refused: a method read as a value (parenthesizeConstituentTypesOfUnionType would lose its object, and this with it) | 2 |
| Refused: a method read as a value (parenthesizeNonArrayTypeOfPostfixType would lose its object, and this with it) | 2 |
| Refused: a method read as a value (setPrototypeOf would lose its object, and this with it) | 2 |
| Refused: a method read as a value (setTimeout would lose its object, and this with it) | 2 |
| Refused: a method read as a value (substituteNode would lose its object, and this with it) | 2 |
| Refused: a method read as a value (toKey would lose its object, and this with it) | 2 |
| Refused: a method read as a value (trackSymbol would lose its object, and this with it) | 2 |
| Refused: a method read as a value (useCaseSensitiveFileNames would lose its object, and this with it) | 2 |
| Refused: a method read as a value (watchDirectory would lose its object, and this with it) | 2 |
| Refused: a method read as a value (watchFile would lose its object, and this with it) | 2 |
| Refused: a method read as a value (writeFile would lose its object, and this with it) | 2 |
| Refused: a number \| undefined as a condition | 2 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on declaration) | 2 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on predicate) | 2 |
| Refused: a value of type (identifierOrPrivateName: Identifier \| PrivateIdentifier) => string seen as (name: GeneratedIdentifier \| GeneratedPrivateIdentifier) => string, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 2 |
| Refused: a value of type ArrayLiteralExpression \| AssignmentExpression<EqualsToken> \| BindingElement \| ElementAccessExpression \| ... 8 more ... \| VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Block seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type BlockLike seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ClassDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ClassStaticBlockDeclaration seen as Mutable<ClassStaticBlockDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ConstructorDeclaration seen as Mutable<ConstructorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Diagnostic seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 2 |
| Refused: a value of type DiagnosticWithLocation[] \| undefined seen as DiagnosticRelatedInformation[] \| undefined, which can write DiagnosticRelatedInformation where DiagnosticWithLocation is read | 2 |
| Refused: a value of type Expression seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type FlowCall seen as FlowNode, which can write BinaryExpression \| CallExpression where CallExpression is read | 2 |
| Refused: a value of type FutureSourceFile \| SourceFile seen as Pick<SourceFile, "fileName" \| "impliedNodeFormat" \| "packageJsonScope">, whose readonly field fileName becomes writable: a readonly field may hold something narrower than string, which a write of string would replace | 2 |
| Refused: a value of type GeneratedIdentifier seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 2 |
| Refused: a value of type GeneratedIdentifier seen as Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 2 |
| Refused: a value of type Identifier seen as Mutable<Identifier>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Identifier[] \| undefined seen as ModuleExportName[] \| undefined, which can write ModuleExportName where Identifier is read | 2 |
| Refused: a value of type ImportTypeAssertionContainer seen as Mutable<ImportTypeAssertionContainer>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type MissingDeclaration seen as Mutable<MissingDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ModifierLike seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ModuleSpecifierResolutionHost & ModuleResolutionHost seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where (() => boolean) & (boolean \| (() => boolean) \| undefined) is read | 2 |
| Refused: a value of type Node seen as Node \| TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Node \| TextRange seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type NodeArray<Statement> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ParameterDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type PropertySignature seen as Mutable<PropertySignature>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type SourceFile seen as EmitNode \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of EmitFlags would replace | 2 |
| Refused: a value of type SourceFile seen as Mutable<SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Statement seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Symbol \| undefined seen as Declaration \| undefined, which can write number \| undefined where number is read | 2 |
| Refused: a value of type Symbol[] seen as unknown[], which can write unknown where Symbol is read | 2 |
| Refused: a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Type[] seen as unknown[], which can write unknown where Type is read | 2 |
| Refused: a value of type any seen as T, a type parameter whose constraint Node can be written, so it can write what any can't hold | 2 |
| Refused: a value of type never[] seen as ChildDirectoryWatcher[], which can write ChildDirectoryWatcher where never is read | 2 |
| Refused: a value of type never[] seen as SourceFile[], which can write SourceFile where never is read | 2 |
| Refused: a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[] \| undefined, which can write SourceFile \| undefined where SourceFile is read | 2 |
| Refused: a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 2 |
| Refused: a value of type string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] seen as unknown[], which can write unknown where string is read | 2 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"FlowFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof FlowFlags is read | 2 |
| Refused: an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | 2 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.FunctionExpression | 2 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NoSubstitutionTemplateLiteral \| SyntaxKind.TemplateHead \| SyntaxKind.TemplateMiddle \| SyntaxKind.TemplateTail | 2 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.TypeKeyword | 2 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.WithKeyword \| SyntaxKind.AssertKeyword | 2 |
| Refused: delete | 2 |
| Refused: optional property all in CompilerOptions absent from structural source {}, which can hide fields | 2 |
| Refused: optional property equalsToken in ShorthandPropertyAssignment absent from structural source never, which can hide fields | 2 |
| Refused: optional property id in LiteralExpression & StringLiteral absent from structural source never, which can hide fields | 2 |
| Refused: optional property modifiers in GetAccessorDeclaration absent from structural source never, which can hide fields | 2 |
| Refused: optional property omitTrailingSemicolon in PrinterOptions absent from structural source CompilerOptions, which can hide fields | 2 |
| Refused: optional property packageJsonScope in Pick<SourceFile, "fileName" \| "impliedNodeFormat" \| "packageJsonScope"> absent from structural source Pick<SourceFile, "fileName" \| "impliedNodeFormat">, which can hide fields | 2 |
| Refused: optional property templateFlags in NoSubstitutionTemplateLiteral absent from structural source never, which can hide fields | 2 |
| Refused: optional property textSourceNode in StringLiteral absent from structural source never, which can hide fields | 2 |
| panic: Unhandled case in Node.Text: *ast.ComputedPropertyName | 2 |
| NotYet: .length on a value | 1 |
| NotYet: ?. to a number, which would be number \| undefined | 1 |
| NotYet: ?.[] on a value | 1 |
| NotYet: Object.assign on a shape not proven by a plain literal or its const binding | 1 |
| NotYet: a BinaryExpression with a boolean and a boolean \| undefined | 1 |
| NotYet: a BinaryExpression with a boolean and a number | 1 |
| NotYet: a BinaryExpression with a boolean \| undefined and a boolean | 1 |
| NotYet: a BinaryExpression with a number and a string | 1 |
| NotYet: a BinaryExpression with a union of differently held members and a union of differently held members | 1 |
| NotYet: a ClassExpression | 1 |
| NotYet: a Map of CompilerOptionsValue | 1 |
| NotYet: a Map of T | 1 |
| NotYet: a PostfixUnaryExpression | 1 |
| NotYet: a YieldExpression as a statement | 1 |
| NotYet: a boolean \| undefined variable a function value captures | 1 |
| NotYet: a class instantiated with TOuterState | 1 |
| NotYet: a destructured name held otherwise than its field | 1 |
| NotYet: a destructured name that isn't plain | 1 |
| NotYet: a destructured parameter beside a parameter with a default | 1 |
| NotYet: a field from a boolean \| undefined variable | 1 |
| NotYet: a field of type AnyBuildOrder \| undefined | 1 |
| NotYet: a field of type NodeArray<ParameterDeclaration> \| readonly JSDocParameterTag[] | 1 |
| NotYet: a field of type false \| VersionPaths \| undefined | 1 |
| NotYet: a field of type string \| false \| undefined | 1 |
| NotYet: a field of type string \| number \| undefined | 1 |
| NotYet: a function returning (AssignmentExpression<EqualsToken> & { readonly left: GeneratedIdentifier; }) \| undefined | 1 |
| NotYet: a function returning (ConstructorDeclaration & { body: Block; }) \| undefined | 1 |
| NotYet: a function returning AnyValidImportOrReExport | 1 |
| NotYet: a function returning AnyValidImportOrReExport \| undefined | 1 |
| NotYet: a function returning CanonicalKey | 1 |
| NotYet: a function returning ClassNamedEvaluationHelperBlock | 1 |
| NotYet: a function returning ClassThisAssignmentBlock | 1 |
| NotYet: a function returning ExpressionWithTypeArguments & { readonly expression: Identifier \| PropertyAccessEntityNameExpression; } | 1 |
| NotYet: a function returning Extract<AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; }, Pick<...>> \| ... 7 more ... \| Extract<...> | 1 |
| NotYet: a function returning HasJSDoc \| undefined | 1 |
| NotYet: a function returning MemberName \| (Expression & (NumericLiteral \| StringLiteralLike)) | 1 |
| NotYet: a function returning ModeAwareCacheKey | 1 |
| NotYet: a function returning Node \| (TIn & undefined) \| (TVisited & undefined) | 1 |
| NotYet: a function returning NodeArray<Node> \| (TInArray & undefined) | 1 |
| NotYet: a function returning NodeArray<TOut> \| (TInArray & undefined) | 1 |
| NotYet: a function returning PathPathComponents | 1 |
| NotYet: a function returning ResolvedConfigFilePath | 1 |
| NotYet: a function returning T \| T[] | 1 |
| NotYet: a function returning T \| readonly T[] | 1 |
| NotYet: a function returning T1 & T2 | 1 |
| NotYet: a function returning TEntry \| undefined | 1 |
| NotYet: a function returning TOut | 1 |
| NotYet: a function returning TOut \| (TIn & undefined) \| (TVisited & undefined) | 1 |
| NotYet: a function returning TOut \| undefined | 1 |
| NotYet: a function returning TPrivateEntry \| undefined | 1 |
| NotYet: a function returning object \| undefined | 1 |
| NotYet: a function returning readonly Node[] \| (TInArray & undefined) | 1 |
| NotYet: a function returning readonly TOut[] \| (TInArray & undefined) | 1 |
| NotYet: a function returning string \| object | 1 |
| NotYet: a function returning unknown | 1 |
| NotYet: a function returning void \| SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> | 1 |
| NotYet: a function returning void \| SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> \| WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> | 1 |
| NotYet: a function returning void \| WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> | 1 |
| NotYet: a literal method through a view that erases its receiver | 1 |
| NotYet: a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables | 1 |
| NotYet: a number \| undefined argument to substring | 1 |
| NotYet: a template interpolating an object, an array, a map, a function or undefined | 1 |
| NotYet: a value of type "" \| ResolvedConfigFileName \| undefined | 1 |
| NotYet: a value of type (AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; } & BinaryExpression) \| (... & ... 1 more ... & BinaryExpression) | 1 |
| NotYet: a value of type (ConstructorDeclaration & { body: Block; }) \| undefined | 1 |
| NotYet: a value of type (ModuleDeclaration & { name: StringLiteral; }) \| undefined | 1 |
| NotYet: a value of type AccessorDeclaration & { readonly name: BigIntLiteral \| ComputedPropertyName \| Identifier \| NoSubstitutionTemplateLiteral \| NumericLiteral \| StringLiteral; } | 1 |
| NotYet: a value of type BindingElement & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type ClassNamedEvaluationHelperBlock | 1 |
| NotYet: a value of type EmitNode & { autoGenerate: AutoGenerateInfo; } | 1 |
| NotYet: a value of type EntityNameExpression \| (LeftHandSideExpression & BindableStaticNameExpression) | 1 |
| NotYet: a value of type ExportAssignment & { readonly expression: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type HasJSDoc | 1 |
| NotYet: a value of type HasJSDoc \| undefined | 1 |
| NotYet: a value of type IncludeTypeSpaceImports | 1 |
| NotYet: a value of type IncrementalBuildInfoFilePendingEmit | 1 |
| NotYet: a value of type IncrementalMultiFileEmitBuildInfoFileInfo | 1 |
| NotYet: a value of type JSDocImportTag \| CanHaveModuleSpecifier | 1 |
| NotYet: a value of type NamedEvaluation | 1 |
| NotYet: a value of type ParameterDeclaration & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type PropertyAssignment & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type PropertyDeclaration & { readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type RequireOrImportCall | 1 |
| NotYet: a value of type ResolvedConfigFileName \| undefined | 1 |
| NotYet: a value of type ResolvedModuleWithFailedLookupLocations & ResolvedTypeReferenceDirectiveWithFailedLookupLocations | 1 |
| NotYet: a value of type ShorthandPropertyAssignment & { readonly objectAssignmentInitializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type SourceFile | 1 |
| NotYet: a value of type T \| readonly T[] | 1 |
| NotYet: a value of type T1 | 1 |
| NotYet: a value of type TData | 1 |
| NotYet: a value of type TEntry | 1 |
| NotYet: a value of type T["kind"] | 1 |
| NotYet: a value of type TypeNode & LiteralTypeNode & { readonly literal: StringLiteral; } | 1 |
| NotYet: a value of type TypeParameterDeclaration & { parent: JSDocTemplateTag; } | 1 |
| NotYet: a value of type U | 1 |
| NotYet: a value of type U \| readonly U[] \| undefined | 1 |
| NotYet: a value of type V | 1 |
| NotYet: a value of type VariableDeclaration & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type WatchFactoryHost & { trace?(s: string): void; } | 1 |
| NotYet: a value of type WrappedExpression<T> | 1 |
| NotYet: a value of type __String & string | 1 |
| NotYet: a value of type false \| RegExpExecArray \| null | 1 |
| NotYet: a value of type object \| undefined | 1 |
| NotYet: a value of type undefined | 1 |
| NotYet: a void call used as a value | 1 |
| NotYet: an array of T \| U | 1 |
| NotYet: an array of V | 1 |
| NotYet: an array of unknown | 1 |
| NotYet: an optional chain longer than one step | 1 |
| NotYet: assigning a field of a value | 1 |
| NotYet: indexOf on a tuple (a tuple is held as an object, not an array, so far; write it as an array where it's made) | 1 |
| NotYet: lastIndexOf with these arguments | 1 |
| NotYet: new Map from something that isn't [key, value] pairs | 1 |
| NotYet: reading BuilderState | 1 |
| NotYet: reading IncrementalParser | 1 |
| NotYet: reading addAggregateStatistic | 1 |
| NotYet: reading addOutput | 1 |
| NotYet: reading captureMapping | 1 |
| NotYet: reading convertToFunctionBlock | 1 |
| NotYet: reading createBaseSourceFileNode | 1 |
| NotYet: reading createIntlCollatorStringComparer | 1 |
| NotYet: reading createPollingIntervalQueue | 1 |
| NotYet: reading enter | 1 |
| NotYet: reading getAccessorNameVisibilityError | 1 |
| NotYet: reading getPackageJsonInfo | 1 |
| NotYet: reading getParenthesizeLeftSideOfBinaryForOperator | 1 |
| NotYet: reading getPreferredEnding | 1 |
| NotYet: reading getSymbolWalker | 1 |
| NotYet: reading getUnusedExpectations | 1 |
| NotYet: reading getValueCandidate | 1 |
| NotYet: reading getVariableDeclarationTypeVisibilityError | 1 |
| NotYet: reading inferPreference | 1 |
| NotYet: reading lookupFromPackageJson | 1 |
| NotYet: reading needJsx | 1 |
| NotYet: reading optionDependsOnRecursive | 1 |
| NotYet: reading parseStrings | 1 |
| NotYet: reading serializeTypeOfDeclaration | 1 |
| NotYet: reading transformSourceFile | 1 |
| NotYet: reading transformSourceFileOrBundle | 1 |
| NotYet: spreading an array of other elements | 1 |
| NotYet: storing any in a field | 1 |
| NotYet: storing string \| number in a field | 1 |
| NotYet: storing true \| Node \| undefined in a field | 1 |
| Refused: &&= | 1 |
| Refused: Object.defineProperty | 1 |
| Refused: a function taking (symbol: Symbol) => boolean seen as one taking ((symbol: Symbol) => boolean) \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking BinaryOperatorToken seen as one taking BinaryOperatorToken \| BinaryOperator (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking JSDocTypeExpression \| undefined seen as one taking JSDocTypeExpression \| JSDocTypeLiteral \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking LogLevel seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking Node seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking Node \| undefined seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking NodeArray<Expression> seen as one taking readonly Expression[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking NodeArray<T> seen as one taking NodeArray<T> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking NodeArray<T> seen as one taking NodeArray<T> \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking NodeArray<TypeNode> \| undefined seen as one taking readonly TypeNode[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking OuterExpressionKinds seen as one taking OuterExpressionKinds \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking PollingInterval seen as one taking number \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking SourceFile seen as one taking [file: SourceFile, options: CompilerOptions] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking T seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking TIn seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking TokenFlags seen as one taking TokenFlags \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking TypeNode seen as one taking Node (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking Visitor seen as one taking Visitor<TIn, Node \| undefined> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [callback: (...args: any[]) => void, ms: number, ...args: any[]] seen as one taking (...args: any[]) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [fileName: string] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [node: ConstructorTypeNode, typeParameters: NodeArray<TypeParameterDeclaration> \| undefined, parameters: NodeArray<ParameterDeclaration>, type: TypeNode] \| ... seen as one taking ConstructorTypeNode (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [path: string, callback: DirectoryWatcherCallback, recursive?: boolean \| undefined, options?: WatchOptions \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [path: string, callback: FileWatcherCallback, pollingInterval?: number \| undefined, options?: WatchOptions \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [timeoutId: any] seen as one taking unknown (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [typeParameters: readonly TypeParameterDeclaration[] \| undefined, parameters: readonly ParameterDeclaration[], type: TypeNode] \| [modifiers: readonly Modifier[] \| undefined, typeParameters: ... \| undefined, parameters: ..., type: TypeNode] seen as one taking readonly Modifier[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [writeByteOrderMark: boolean, onError?: ((message: string) => void) \| undefined, sourceFiles?: readonly SourceFile[] \| undefined, data?: WriteFileCallbackData \| undefined] seen as one taking boolean (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking never seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking number seen as one taking any[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking number seen as one taking number \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking readonly ParameterDeclaration[] seen as one taking readonly ParameterDeclaration[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking readonly T[] \| undefined seen as one taking readonly T[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking string seen as one taking string \| MemberName (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking string \| undefined seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a method read as a value (add would lose its object, and this with it) | 1 |
| Refused: a method read as a value (base64decode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (base64encode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (clearScreen would lose its object, and this with it) | 1 |
| Refused: a method read as a value (compare would lose its object, and this with it) | 1 |
| Refused: a method read as a value (convertToArrayAssignmentElement would lose its object, and this with it) | 1 |
| Refused: a method read as a value (convertToObjectAssignmentElement would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createComma would lose its object, and this with it) | 1 |
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
| Refused: a method read as a value (directoryExists would lose its object, and this with it) | 1 |
| Refused: a method read as a value (emit would lose its object, and this with it) | 1 |
| Refused: a method read as a value (emitBuildInfo would lose its object, and this with it) | 1 |
| Refused: a method read as a value (emitNextAffectedFile would lose its object, and this with it) | 1 |
| Refused: a method read as a value (fill would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getAllDependencies would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getCurrentDirectory would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getDeclarationDiagnostics would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getDirectories would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getMemoryUsage would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getModifiedTime would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getNearestAncestorDirectoryWithPackageJson would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getOrCreateCacheForModuleName would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getSemanticDiagnostics would lose its object, and this with it) | 1 |
| Refused: a method read as a value (globalCacheResolutionModuleName would lose its object, and this with it) | 1 |
| Refused: a method read as a value (hasChangedEmitSignature would lose its object, and this with it) | 1 |
| Refused: a method read as a value (hasOwnProperty would lose its object, and this with it) | 1 |
| Refused: a method read as a value (log would lose its object, and this with it) | 1 |
| Refused: a method read as a value (nonEscapingWrite would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeCheckTypeOfConditionalType would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeConciseBodyOfArrowFunction would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeConditionOfConditionalExpression would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeConstituentTypeOfIntersectionType would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeConstituentTypeOfUnionType would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeElementTypeOfTupleType would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeExpressionOfComputedPropertyName would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeExpressionOfExportDefault would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeExpressionOfExpressionStatement would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeExpressionOfNew would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeExtendsTypeOfConditionalType would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeLeadingTypeArgument would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeOperandOfPostfixUnary would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeOperandOfReadonlyTypeOperator would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeOperandOfTypeOperator would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeTypeOfOptionalType would lose its object, and this with it) | 1 |
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
| Refused: a method read as a value (setBlocking would lose its object, and this with it) | 1 |
| Refused: a method read as a value (setModifiedTime would lose its object, and this with it) | 1 |
| Refused: a method read as a value (toString would lose its object, and this with it) | 1 |
| Refused: a method read as a value (tryEnableSourceMapsForHost would lose its object, and this with it) | 1 |
| Refused: a method read as a value (writeOutputIsTTY would lose its object, and this with it) | 1 |
| Refused: a type predicate whose return is not proven (asserts cond needs a boolean parameter) | 1 |
| Refused: a type predicate whose return is not proven (false return can still contain LateBoundDeclaration) | 1 |
| Refused: a type predicate whose return is not proven (normal return has not narrowed value to NonNullable<T>) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on array) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on buildOrder) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on diagnostic) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on file) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on location) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on option) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on options) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on p) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on program) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on sourceFile) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on symbol) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on type) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on x) | 1 |
| Refused: a type predicate whose return is not proven (the body's true narrowing does not match null \| undefined) | 1 |
| Refused: a type predicate whose return is not proven (the predicate parameter is assigned) | 1 |
| Refused: a type predicate whose return is not proven (true return narrows to MappedPosition, not SourceMappedPosition) | 1 |
| Refused: a type predicate whose return is not proven (true return narrows to Mapping, not SourceMapping) | 1 |
| Refused: a type predicate whose return is not proven (true return narrows to ReusableBuilderProgramState, not BuilderProgramStateWithDefinedProgram) | 1 |
| Refused: a value of type (AbstractKeyword \| AccessorKeyword \| AsyncKeyword \| ConstKeyword \| DeclareKeyword \| Decorator \| ... 7 more ... \| StaticKeyword)[] \| undefined seen as Node[] \| undefined, which can write Node where AbstractKeyword \| AccessorKeyword \| AsyncKeyword \| ConstKeyword \| DeclareKeyword \| Decorator \| ... 7 more ... \| StaticKeyword is read | 1 |
| Refused: a value of type (baseDir: string, moduleName: string) => { module: any; modulePath: string; error: undefined; } \| { module: undefined; modulePath: undefined; error: unknown; } seen as (baseDir: string, moduleName: string) => ModuleImportResult, which can write string \| undefined where string is read | 1 |
| Refused: a value of type (left: MappedPosition, right: MappedPosition) => boolean seen as EqualityComparer<SourceMappedPosition> \| undefined, which can write string \| undefined where string is read | 1 |
| Refused: a value of type (node: CommentRange) => boolean seen as (value: SynthesizedComment) => boolean, which can write number where -1 is read | 1 |
| Refused: a value of type (recordTempVariable: ((node: Identifier) => void) \| undefined, reservedInNestedScopes?: boolean \| undefined, prefix?: string \| GeneratedNamePart \| undefined, suffix?: string \| undefined) => GeneratedIdentifier seen as { (recordTempVariable: ((node: Identifier) => void) \| undefined, reservedInNestedScopes?: boolean \| undefined): Identifier; (recordTempVariable: ((node: Identifier) => void) \| undefined, reservedInNestedScopes?: boolean \| undefined, prefix?: string \| ... 1 more ... \| undefined, suffix?: string \| undefined): Identifi..., whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type (symbol: Symbol) => { visitedTypes: Type[]; visitedSymbols: Symbol[]; } seen as (root: Symbol) => { visitedTypes: readonly Type[]; visitedSymbols: readonly Symbol[]; }, which can write readonly Type[] where Type[] is read | 1 |
| Refused: a value of type (symbolAccessibilityResult: SymbolAccessibilityResult) => { diagnosticMessage: DiagnosticMessage; errorNode: DeclarationDiagnosticProducing; typeName: DeclarationName \| undefined; } \| undefined seen as (symbolAccessibilityResult: SymbolAccessibilityResult) => SymbolAccessibilityDiagnostic \| undefined, which can write Node where DeclarationDiagnosticProducing is read | 1 |
| Refused: a value of type (type: Type) => { visitedTypes: Type[]; visitedSymbols: Symbol[]; } seen as (root: Type) => { visitedTypes: readonly Type[]; visitedSymbols: readonly Symbol[]; }, which can write readonly Type[] where Type[] is read | 1 |
| Refused: a value of type AssertClause seen as Mutable<AssertClause>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type AssertEntry seen as Mutable<AssertEntry>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type BinaryExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Block seen as Mutable<Block>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type BreakStatement seen as Mutable<BreakStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type CallSignatureDeclaration seen as Mutable<CallSignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type CaseBlock seen as Mutable<CaseBlock>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Children seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ClassDeclaration \| FunctionDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ClassElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type CompilerHost seen as CompilerHostLikeForCache, which can write WriteFileCallback \| undefined where WriteFileCallback is read | 1 |
| Refused: a value of type CompilerOptions & { types: string[]; } seen as CompilerOptions, which can write string[] \| undefined where string[] is read | 1 |
| Refused: a value of type ConstructSignatureDeclaration seen as Mutable<ConstructSignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ConstructorTypeNode seen as Mutable<ConstructorTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ContinueStatement seen as Mutable<ContinueStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type DestructuringAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type DiagnosticWithLocation \| undefined seen as Diagnostic \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 |
| Refused: a value of type ElementAccessExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type EnumDeclaration \| ModuleDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ExportAssignment seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Expression \| GeneratedIdentifier seen as Expression \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type Expression \| GeneratedIdentifier seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type ExpressionStatement seen as Mutable<ExpressionStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Extension[] seen as string[], which can write string where Extension is read | 1 |
| Refused: a value of type FlowArrayMutation \| FlowAssignment \| FlowCall \| FlowCondition \| FlowLabel \| FlowReduceLabel \| FlowStart \| FlowUnreachable seen as FlowNode, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 1 |
| Refused: a value of type FunctionTypeNode seen as Mutable<FunctionTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type GeneratedIdentifier seen as Expression \| GeneratedIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier seen as Node \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as GeneratedIdentifier \| GeneratedPrivateIdentifier \| Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as Identifier \| PrivateIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as Node \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier \| Identifier \| PrivateIdentifier seen as Identifier \| PrivateIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| Identifier seen as Identifier \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedPrivateIdentifier seen as Node \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GetAccessorDeclaration \| SetAccessorDeclaration seen as Mutable<AccessorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Identifier seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Identifier seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Identifier[] seen as ModuleExportName[] \| undefined, which can write ModuleExportName where Identifier is read | 1 |
| Refused: a value of type ImportAttribute seen as Mutable<ImportAttribute>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ImportAttributes seen as Mutable<ImportAttributes>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ImportClause seen as Mutable<ImportClause>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ImportDeclaration seen as Mutable<ImportDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ImportEqualsDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ImportTypeNode seen as Mutable<ImportTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type IndexSignatureDeclaration seen as Mutable<IndexSignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocAugmentsTag seen as Mutable<JSDocAugmentsTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocCallbackTag seen as Mutable<JSDocCallbackTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocFunctionType seen as Mutable<JSDocFunctionType>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocImplementsTag seen as Mutable<JSDocImplementsTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocImportTag seen as Mutable<JSDocImportTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocLink seen as Mutable<JSDocLink>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocLinkCode seen as Mutable<JSDocLinkCode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocLinkPlain seen as Mutable<JSDocLinkPlain>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocNameReference seen as Mutable<JSDocNameReference>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocOverloadTag seen as Mutable<JSDocOverloadTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocParameterTag seen as Mutable<JSDocParameterTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocPropertyTag seen as Mutable<JSDocPropertyTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocSeeTag seen as Mutable<JSDocSeeTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocSignature seen as Mutable<JSDocSignature>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocTemplateTag seen as Mutable<JSDocTemplateTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocText seen as Mutable<JSDocText>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocTypeExpression seen as Mutable<JSDocTypeExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocTypeLiteral seen as Mutable<JSDocTypeLiteral>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocTypedefTag seen as Mutable<JSDocTypedefTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocUnknownTag seen as Mutable<JSDocUnknownTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
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
| Refused: a value of type Map<string, WatchDirectoryKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where WatchDirectoryKind is read | 1 |
| Refused: a value of type Map<string, WatchFileKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where WatchFileKind is read | 1 |
| Refused: a value of type Map<string, string> seen as Map<string, string \| number>, which can write string \| number where string is read | 1 |
| Refused: a value of type MappedTypeNode seen as Mutable<MappedTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type MethodDeclaration seen as Mutable<MethodDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type MethodDeclaration \| PropertyAssignment \| AccessorDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ModuleDeclaration seen as Mutable<ModuleDeclaration \| SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ModuleExportName seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ModuleName seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ModuleResolutionState & { host: GetPackageJsonEntrypointsHost; } seen as ModuleResolutionState, which can write ModuleResolutionHost where ModuleResolutionHost & GetPackageJsonEntrypointsHost is read | 1 |
| Refused: a value of type Mutable<NoSubstitutionTemplateLiteral> seen as Mutable<TemplateLiteralLikeNode>, which can write SyntaxKind where SyntaxKind.NoSubstitutionTemplateLiteral is read | 1 |
| Refused: a value of type Mutable<T> seen as T, a type parameter whose constraint JSDocType & { readonly type: TypeNode \| undefined; readonly postfix: boolean; } can be written, so it can write what Mutable<T> can't hold | 1 |
| Refused: a value of type Mutable<T> seen as T, a type parameter whose constraint JSDocType & { readonly type: TypeNode \| undefined; } can be written, so it can write what Mutable<T> can't hold | 1 |
| Refused: a value of type Mutable<T> seen as T, a type parameter whose constraint Node can be written, so it can write what Mutable<T> can't hold | 1 |
| Refused: a value of type NamedImports seen as Mutable<NamedImports>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type NamespaceExport seen as Mutable<NamespaceExport>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type NamespaceExportDeclaration seen as Mutable<NamespaceExportDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type NamespaceImport seen as Mutable<NamespaceImport>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Node seen as Node \| SourceMapRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Node \| SourceMapRange seen as SourceMapRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Node \| undefined seen as EmitNode \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of EmitFlags would replace | 1 |
| Refused: a value of type NonNullExpression seen as Mutable<NonNullExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ObjectBindingOrAssignmentPattern seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type OptionalTypeNode seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ParameterDeclaration \| PropertyDeclaration \| PropertySignature \| SignatureDeclaration seen as Mutable<ParameterDeclaration \| PropertyDeclaration \| PropertySignature \| SignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ParameterDeclaration \| VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ParameterDeclaration \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type ParseConfigHost seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where boolean is read | 1 |
| Refused: a value of type Path[] seen as string[], which can write string where Path is read | 1 |
| Refused: a value of type PropertyAssignment seen as Mutable<PropertyAssignment \| ShorthandPropertyAssignment>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Readonly<BuilderState> \| undefined seen as BuilderState \| undefined, whose readonly field fileInfos becomes writable: a readonly field may hold something narrower than Map<Path, FileInfo>, which a write of Map<Path, FileInfo> would replace | 1 |
| Refused: a value of type ReferencedFile & { kind: FileIncludeKind.LibReferenceDirective; } seen as ReferencedFile, which can write ReferencedFileKind where FileIncludeKind.LibReferenceDirective is read | 1 |
| Refused: a value of type ReportFileInError[] seen as (ReportFileInError \| undefined)[], which can write ReportFileInError \| undefined where ReportFileInError is read | 1 |
| Refused: a value of type ResolvedModuleFull \| undefined seen as { path: string; originalPath: string \| true; extension: string; packageId: PackageId \| undefined; resolvedUsingTsExtension: boolean \| undefined; } \| undefined, whose readonly field originalPath becomes writable: a readonly field may hold something narrower than string \| undefined, which a write of string \| true would replace | 1 |
| Refused: a value of type ReturnStatement seen as Mutable<ReturnStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type SearchResult<Resolved> seen as { value: { resolved: Resolved; isExternalLibraryImport: true; } \| undefined; } \| undefined, whose readonly field value becomes writable: a readonly field may hold something narrower than Resolved \| undefined, which a write of { resolved: Resolved; isExternalLibraryImport: true; } \| undefined would replace | 1 |
| Refused: a value of type SetAccessorDeclaration seen as Mutable<SetAccessorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ShorthandPropertyAssignment seen as Mutable<PropertyAssignment \| ShorthandPropertyAssignment>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type SourceFile seen as Mutable<ModuleDeclaration \| SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type SourceFile \| undefined seen as FileReasonToChainCache \| undefined, which can write DiagnosticMessageChain[] \| undefined where RedirectInfo \| undefined is read | 1 |
| Refused: a value of type SourceMapSource seen as SourceFileLike, which can write readonly number[] \| undefined where readonly number[] is read | 1 |
| Refused: a value of type Statement[] seen as Node[], which can write Node where Statement is read | 1 |
| Refused: a value of type StringLiteral seen as Mutable<StringLiteral>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type SwitchStatement seen as Mutable<SwitchStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type System seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where boolean is read | 1 |
| Refused: a value of type T seen as T, a type parameter whose constraint ClassDeclaration \| ClassExpression \| GetAccessorDeclaration \| MethodDeclaration \| ParameterDeclaration \| PropertyDeclaration \| SetAccessorDeclaration can be written, so it can write what T can't hold | 1 |
| Refused: a value of type T seen as T, a type parameter whose constraint HasModifiers can be written, so it can write what T can't hold | 1 |
| Refused: a value of type T seen as T, a type parameter whose constraint MethodDeclaration \| MethodSignature \| PropertyAssignment \| PropertyDeclaration \| PropertySignature \| AccessorDeclaration can be written, so it can write what T can't hold | 1 |
| Refused: a value of type T seen as T, a type parameter whose constraint ModifierSyntaxKind can be written, so it can write what T can't hold | 1 |
| Refused: a value of type T seen as T, a type parameter whose constraint Node can be written, so it can write what T can't hold | 1 |
| Refused: a value of type T seen as T, a type parameter whose constraint Node \| undefined can be written, so it can write what T can't hold | 1 |
| Refused: a value of type TKind seen as TKind, a type parameter whose constraint KeywordTypeSyntaxKind can be written, so it can write what TKind can't hold | 1 |
| Refused: a value of type T[] seen as (T \| undefined)[], which can write T \| undefined where T is read | 1 |
| Refused: a value of type T[][] seen as (readonly T[])[], which can write readonly T[] where T[] is read | 1 |
| Refused: a value of type TaggedTemplateExpression seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type TsConfigSourceFile \| CompilerOptionsValue seen as string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined, which can write string \| number where string is read | 1 |
| Refused: a value of type TypeOperatorNode seen as Mutable<TypeOperatorNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type TypeParameterDeclaration seen as Mutable<TypeParameterDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type VariableDeclaration \| DestructuringAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type VariableDeclaration[] seen as VariableStatement[], which can write VariableStatement where VariableDeclaration is read | 1 |
| Refused: a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read | 1 |
| Refused: a value of type YieldExpression seen as Mutable<YieldExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type never seen as T, a type parameter whose constraint Node can be written, so it can write what never can't hold | 1 |
| Refused: a value of type never seen as T, a type parameter whose constraint any can be written, so it can write what never can't hold | 1 |
| Refused: a value of type never seen as U, a type parameter whose constraint {} can be written, so it can write what never can't hold | 1 |
| Refused: a value of type never[] seen as (JSDoc \| JSDocTag)[], which can write JSDoc \| JSDocTag where never is read | 1 |
| Refused: a value of type never[] seen as (JSDocCallbackTag \| JSDocEnumTag \| JSDocTypedefTag)[], which can write JSDocCallbackTag \| JSDocEnumTag \| JSDocTypedefTag where never is read | 1 |
| Refused: a value of type never[] seen as (SourceMapRange \| undefined)[], which can write SourceMapRange \| undefined where never is read | 1 |
| Refused: a value of type never[] seen as CommentRange[], which can write CommentRange where never is read | 1 |
| Refused: a value of type never[] seen as Comparator[][], which can write Comparator[] where never is read | 1 |
| Refused: a value of type never[] seen as Declaration[], which can write Declaration where never is read | 1 |
| Refused: a value of type never[] seen as Expression[], which can write Expression where never is read | 1 |
| Refused: a value of type never[] seen as FlowNode[], which can write FlowNode where never is read | 1 |
| Refused: a value of type never[] seen as JSDocImportTag[], which can write JSDocImportTag where never is read | 1 |
| Refused: a value of type never[] seen as ProjectReference[], which can write ProjectReference where never is read | 1 |
| Refused: a value of type never[] seen as RequireOrImportCall[], which can write RequireOrImportCall where never is read | 1 |
| Refused: a value of type never[] seen as ResolvedProjectReference[], which can write ResolvedProjectReference where never is read | 1 |
| Refused: a value of type never[] seen as SourceMappedPosition[], which can write SourceMappedPosition where never is read | 1 |
| Refused: a value of type never[] seen as StringLiteralLike[], which can write StringLiteralLike where never is read | 1 |
| Refused: a value of type never[] seen as TransformerFactory<Bundle \| SourceFile>[], which can write TransformerFactory<Bundle \| SourceFile> where never is read | 1 |
| Refused: a value of type never[] \| SortedArray<DiagnosticWithLocation> seen as Diagnostic[], which can write Diagnostic where never is read | 1 |
| Refused: a value of type never[][] seen as string[][], which can write string where never is read | 1 |
| Refused: a value of type number[] seen as string \| (string \| number)[] \| undefined, which can write string \| number where number is read | 1 |
| Refused: a value of type readonly string[] \| undefined seen as RegExp[] \| undefined, which can write RegExp where string is read | 1 |
| Refused: a value of type string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined seen as string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined, which can write string \| number where string is read | 1 |
| Refused: a value of type string \| number \| boolean \| string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> seen as CompilerOptionsValue, which can write string \| number where string is read | 1 |
| Refused: a value of type string[] seen as string \| (string \| number)[] \| undefined, which can write string \| number where string is read | 1 |
| Refused: a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"CheckMode", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof CheckMode is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"EmitFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof EmitFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"GeneratedIdentifierFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof GeneratedIdentifierFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"ModifierFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ModifierFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"NodeCheckFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof NodeCheckFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"NodeFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof NodeFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"ObjectFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ObjectFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"RelationComparisonResult", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof RelationComparisonResult is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"ScriptKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ScriptKind is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"SignatureCheckMode", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SignatureCheckMode is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"SignatureFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SignatureFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"SnippetKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SnippetKind is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"SymbolFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SymbolFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"SyntaxKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SyntaxKind is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"TransformFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TransformFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"TypeFacts", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TypeFacts is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"TypeFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TypeFlags is read | 1 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what undefined can't hold | 1 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint Declaration can be written, so it can write what undefined can't hold | 1 |
| Refused: a value of type { (fileName: string): DiagnosticWithLocation[]; (): Diagnostic[]; } seen as { (): Diagnostic[]; (fileName: string): DiagnosticWithLocation[]; }, which can write Diagnostic where DiagnosticWithLocation is read | 1 |
| Refused: a value of type { affectedFile: SourceFile; emitKind: BuilderFileEmit.Js \| BuilderFileEmit.JsMap \| BuilderFileEmit.JsInlineMap \| BuilderFileEmit.DtsErrors \| ... 6 more ... \| BuilderFileEmit.All; } seen as { affectedFile: Program \| SourceFile \| undefined; emitKind: BuilderFileEmit; }, which can write Program \| SourceFile \| undefined where SourceFile is read | 1 |
| Refused: a value of type { all?: boolean; allowImportingTsExtensions?: boolean; allowJs?: boolean; allowNonTsExtensions?: boolean; allowArbitraryExtensions?: boolean; allowSyntheticDefaultImports?: boolean; allowUmdGlobalAccess?: boolean; ... 129 more ...; moduleResolution: ModuleResolutionKind; } seen as CompilerOptions, which can write ModuleResolutionKind \| undefined where ModuleResolutionKind is read | 1 |
| Refused: a value of type { arguments: { name?: string; } & { path: string; }; range: CommentRange; } \| { arguments: { name: string; }; range: CommentRange; } \| { arguments: { factory: string; }; range: CommentRange; } \| ... 5 more ... \| ... seen as ({ arguments: { name?: string; } & { path: string; }; range: CommentRange; } \| { arguments: { name: string; }; range: CommentRange; } \| { arguments: { factory: string; }; range: CommentRange; } \| ... 5 more ... \| ...)[] \| ... 8 more ... \| ..., which can write { name?: string; } & { path: string; } where never is read | 1 |
| Refused: a value of type { compilerOptions: CompilerOptions; traceEnabled: boolean; affectingLocations: string[] \| undefined; resultFromCache?: ResolvedModuleWithFailedLookupLocations; ... 9 more ...; host: GetPackageJsonEntrypointsHost; } seen as ModuleResolutionState & { host: GetPackageJsonEntrypointsHost; }, which can write string[] \| undefined where never[] is read | 1 |
| Refused: a value of type { ending: ModuleSpecifierEnding; value: string; }[] seen as { ending: ModuleSpecifierEnding \| undefined; value: string; }[], which can write ModuleSpecifierEnding \| undefined where ModuleSpecifierEnding is read | 1 |
| Refused: a value of type { host: ModuleResolutionHost; traceEnabled: boolean; failedLookupLocations: string[] \| undefined; affectingLocations: string[] \| undefined; resultFromCache?: ResolvedModuleWithFailedLookupLocations; ... 8 more ...; reportDiagnostic: ...; } seen as ModuleResolutionState, which can write CompilerOptions where { all?: boolean; allowImportingTsExtensions?: boolean; allowJs?: boolean; allowNonTsExtensions?: boolean; allowArbitraryExtensions?: boolean; allowSyntheticDefaultImports?: boolean; allowUmdGlobalAccess?: boolean; ... 129 more ...; moduleResolution: ModuleResolutionKind; } is read | 1 |
| Refused: a value of type { id: number; flowNode: FlowNode; edges: never[]; text: string; lane: number; endLane: number; level: number; circular: false; } seen as FlowGraphNode, which can write FlowGraphEdge[] where never[] is read | 1 |
| Refused: a value of type { major: number; minor: number; patch: number; prerelease: string; build: string; } seen as { major: string \| number; minor: number; patch: number; prerelease: string \| readonly string[]; build: string \| readonly string[]; }, which can write string \| number where number is read | 1 |
| Refused: a value of type { name: string \| undefined; path: string; }[] seen as AmdDependency[], which can write AmdDependency where { name: string \| undefined; path: string; } is read | 1 |
| Refused: an arbitrary number or a value from another enum assigned to Extension.Mjs; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to ForegroundColorEscapeSequences.Grey; its members are a closed union | 1 |
| Refused: an unproven relation from Identifier to Identifier: optional field id has no proven compatible presence/type | 1 |
| Refused: an unproven relation from ImportEqualsDeclaration & { moduleReference: ExternalModuleReference; } to ImportEqualsDeclaration: optional field moduleReference.id has no proven compatible presence/type | 1 |
| Refused: an unproven relation from LiteralLikeNode to TemplateLiteralLikeNode: optional field rawText has no proven compatible presence/type | 1 |
| Refused: an unproven relation from NodeArray<ModifierLike> & readonly Decorator[] to NodeArray<Decorator>: optional field concat.parameter.element.slice.element.parent.name has no proven compatible presence/type | 1 |
| Refused: an unproven relation from PackageJsonPathFields to PackageJson: optional field version has no proven compatible presence/type | 1 |
| Refused: an unproven relation from SolutionBuilderHostBase<T> to SolutionBuilderHost<T>: optional field reportErrorSummary has no proven compatible presence/type | 1 |
| Refused: an unproven relation from string to T: the source is not assignable to the target | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot BuilderFileEmit.Js \| BuilderFileEmit.JsMap \| BuilderFileEmit.JsInlineMap \| BuilderFileEmit.DtsErrors \| BuilderFileEmit.DtsEmit \| ... 5 more ... \| BuilderFileEmit.All | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot BuilderFileEmit.None | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot CharacterCodes.plus | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot CharacterCodes.slash | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot Connection | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot EmitOnly.Dts | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot ModuleKind.ESNext | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot NodeFlags.NestedNamespace | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot PunctuationOrKeywordSyntaxKind | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.BindingElement | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ConditionalType | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ExtendsKeyword \| SyntaxKind.ImplementsKeyword | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Identifier | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ImportAttributes | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.KeyOfKeyword \| SyntaxKind.ReadonlyKeyword \| SyntaxKind.UniqueKeyword | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NamespaceExport | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NamespaceImport | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NumericLiteral \| SyntaxKind.BigIntLiteral \| SyntaxKind.StringLiteral \| SyntaxKind.JsxText \| SyntaxKind.JsxTextAllWhiteSpaces \| SyntaxKind.RegularExpressionLiteral \| SyntaxKind.NoSubstitutionTemplateLiteral | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.SourceFile | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.StringLiteral | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Unknown \| SyntaxKind.NumericLiteral \| SyntaxKind.BigIntLiteral \| SyntaxKind.StringLiteral \| SyntaxKind.JsxText \| SyntaxKind.RegularExpressionLiteral \| SyntaxKind.NoSubstitutionTemplateLiteral | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.VariableDeclaration | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot TempFlags | 1 |
| Refused: arguments | 1 |
| Refused: debugger | 1 |
| Refused: inherited library member hasOwnProperty read as an own field | 1 |
| Refused: inherited library member replace read as an own field | 1 |
| Refused: optional property Low in Partial<Levels> absent from structural source {}, which can hide fields | 1 |
| Refused: optional property _propertyAccessExpressionLikeQualifiedNameBrand in PropertyAccessEntityNameExpression absent from structural source never, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source OptionsBase, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "noImplicitAny" \| "strict">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "noImplicitThis" \| "strict">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" \| "strictBindCallApply">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" \| "strictBuiltinIteratorReturn">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" \| "strictFunctionTypes">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" \| "strictNullChecks">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" \| "strictPropertyInitialization">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" \| "useUnknownInCatchVariables">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source TypeAcquisition, which can hide fields | 1 |
| Refused: optional property createProgram in IncrementalProgramOptions<EmitAndSemanticDiagnosticsBuilderProgram> absent from structural source IncrementalCompilationOptions, which can hide fields | 1 |
| Refused: optional property id in BinaryExpression absent from structural source never, which can hide fields | 1 |
| Refused: optional property id in FalseLiteral absent from structural source never, which can hide fields | 1 |
| Refused: optional property jsDocCache in JSDocArray absent from structural source JSDoc[], which can hide fields | 1 |
| Refused: optional property modifiers in ExportAssignment absent from structural source never, which can hide fields | 1 |
| Refused: optional property modifiers in ParameterDeclaration absent from structural source never, which can hide fields | 1 |
| Refused: optional property modifiers in PropertyDeclaration absent from structural source never, which can hide fields | 1 |
| Refused: optional property modifiers in SetAccessorDeclaration absent from structural source never, which can hide fields | 1 |
| Refused: optional property packageRootPath in { moduleFileToTry: string; packageRootPath?: string; blockedByExports?: true; verbatimFromExports?: true; } absent from structural source { moduleFileToTry: string; }, which can hide fields | 1 |
| Refused: optional property preserve in FileReference absent from structural source { resolutionMode: ModuleKind.CommonJS \| ModuleKind.ESNext; }, which can hide fields | 1 |
| Refused: optional property skipTrivia in SourceMapSource absent from structural source SourceFile, which can hide fields | 1 |
| Refused: optional property types in { types?: { value: string; pos: number; end: number; }; } & { lib?: { value: string; pos: number; end: number; }; } & { path?: { value: string; pos: number; end: number; }; } & { "no-default-lib"?: string; } & { "resolution-mode"?: string; } & ... absent from structural source { [index: string]: string \| undefined; }, which can hide fields | 1 |
| Refused: optional property types in { types?: { value: string; pos: number; end: number; }; } & { lib?: { value: string; pos: number; end: number; }; } & { path?: { value: string; pos: number; end: number; }; } & { "no-default-lib"?: string; } & { "resolution-mode"?: string; } & ... absent from structural source { [index: string]: string \| { value: string; pos: number; end: number; }; }, which can hide fields | 1 |
| Refused: optional property watchFile in WatchOptions absent from structural source {}, which can hide fields | 1 |
| panic: Unhandled case in Node.Text: *ast.QualifiedName | 1 |

## latent / compiler

| Kind and exact reason | Sites |
| --- | ---: |
| Refused: a cast the runtime can't check | 1396 |
| Refused: an object refinement using an open numeric enum as a literal tag | 1182 |
| Refused: the non-null assertion ! | 767 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on node) | 388 |
| NotYet: a function without a body | 191 |
| NotYet: a PrefixUnaryExpression on a value | 110 |
| NotYet: a method call through a structural signature in a program with statics; use typeof the declaring class | 96 |
| NotYet: a NonNullExpression | 71 |
| NotYet: a function returning T \| undefined | 71 |
| Refused: a type predicate whose return is not proven (there is no body proving this parameter) | 71 |
| NotYet: reading Debug | 55 |
| NotYet: a function returning T | 53 |
| NotYet: a BinaryExpression with a value and a boolean | 40 |
| NotYet: a BinaryExpression with a value and a value | 40 |
| Refused: the comma operator | 38 |
| Refused: a type predicate whose return is not proven (return paths through KindSwitchStatement are not verified) | 36 |
| NotYet: a field of type boolean \| undefined | 33 |
| NotYet: a PrefixUnaryExpression on a number | 31 |
| NotYet: a value of type T | 30 |
| NotYet: a value of type any | 30 |
| Refused: a value as a condition | 29 |
| Refused: a value of type SourceFile seen as SourceFileLike, which can write readonly number[] \| undefined where readonly number[] is read | 28 |
| Refused: \|\|= | 25 |
| NotYet: a parameter that isn't a plain name | 22 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on kind) | 22 |
| NotYet: a value of type Path | 21 |
| NotYet: a value of type __String | 19 |
| NotYet: a value of type ResolvedConfigFilePath | 18 |
| NotYet: a function returning U \| undefined | 17 |
| Refused: a string as a condition | 17 |
| Refused: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 16 |
| Refused: a boolean \| undefined as a condition | 15 |
| NotYet: a generic function as a value | 14 |
| NotYet: a value of type unknown | 14 |
| Refused: a method read as a value (liftToBlock would lose its object, and this with it) | 14 |
| Refused: a type predicate whose return is not proven (branch is not a trusted parameter check) | 14 |
| Refused: a value of type FlowNode seen as FlowNode \| undefined, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 14 |
| Refused: a value of type { value: never; done: true; } seen as IteratorResult<Mapping, any>, which can write any where never is read | 14 |
| Refused: a value of type Node seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 13 |
| Refused: a method in object destructuring | 12 |
| Refused: a method read as a value (parenthesizeExpressionForDisallowedComma would lose its object, and this with it) | 12 |
| Refused: a number as a condition | 12 |
| Refused: a value without nominal ancestry seen as BinaryExpressionStateMachine<TOuterState, TState, TResult> | 12 |
| NotYet: a PrefixUnaryExpression on a string | 11 |
| NotYet: a call through ?. (an optional call) | 11 |
| NotYet: a value of type T \| undefined | 11 |
| Refused: a definite assignment assertion ! | 11 |
| Refused: a namespace | 11 |
| Refused: a value of type T seen as Mutable<T>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of T["pos"] would replace | 11 |
| Refused: a value of type never[] seen as string[], which can write string where never is read | 11 |
| Refused: an index signature | 11 |
| NotYet: a BinaryExpression with a boolean and a value | 10 |
| NotYet: a function returning __String | 10 |
| Refused: a value of type BuilderProgramStateWithDefinedProgram seen as BuilderProgramState, which can write Program \| undefined where Program is read | 10 |
| Refused: a value of type Map<string, never[]> seen as Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>, which can write string[] where never[] is read | 10 |
| Refused: optional property id in Node absent from structural source never, which can hide fields | 10 |
| NotYet: a ModuleDeclaration | 9 |
| NotYet: an array of T | 9 |
| NotYet: an enum inside a function or block; declare it at module scope | 9 |
| NotYet: for...of over an object | 9 |
| Refused: a function taking Identifier seen as one taking Identifier \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 9 |
| Refused: a function taking boolean seen as one taking boolean \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 9 |
| Refused: a value of type BuilderProgramState seen as ReusableBuilderProgramState, which can write Map<Path, readonly Diagnostic[] \| readonly ReusableDiagnostic[]> where Map<Path, readonly Diagnostic[]> is read | 9 |
| Refused: optional property source in SourceMapRange absent from structural source TextRange, which can hide fields | 9 |
| NotYet: new an Identifier | 8 |
| NotYet: a BinaryExpression as a statement | 7 |
| NotYet: a BinaryExpression with a number and a number | 7 |
| NotYet: a BinaryExpression with a number \| undefined and a number | 7 |
| NotYet: a BinaryExpression with a string and a string | 7 |
| NotYet: a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions | 7 |
| NotYet: a value of type ResolvedConfigFileName | 7 |
| NotYet: this outside a method | 7 |
| Refused: a method read as a value (parenthesizeLeftSideOfAccess would lose its object, and this with it) | 7 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.JSDocTypeExpression | 7 |
| Refused: optional property id in ArrayBindingPattern absent from structural source never, which can hide fields | 7 |
| Refused: optional property id in BigIntLiteral absent from structural source never, which can hide fields | 7 |
| NotYet: a PrefixUnaryExpression on a number \| undefined | 6 |
| NotYet: a field of type true \| undefined | 6 |
| NotYet: a function returning any | 6 |
| NotYet: a value of type object | 6 |
| NotYet: an array of never | 6 |
| NotYet: regex replacement other than a string | 6 |
| Refused: a function taking BinaryExpressionStateMachine<TOuterState, TState, TResult> seen as one taking BinaryExpressionStateMachine<TOuterState, TState, TResult> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 6 |
| Refused: a method read as a value (parenthesizeOperandOfPrefixUnary would lose its object, and this with it) | 6 |
| Refused: a method read as a value (readFile would lose its object, and this with it) | 6 |
| Refused: a parameter property | 6 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on token) | 6 |
| Refused: a value of type NonNullable<T> seen as T, a type parameter whose constraint any can be written, so it can write what NonNullable<T> can't hold | 6 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint Node can be written, so it can write what undefined can't hold | 6 |
| Refused: var | 6 |
| NotYet: a BinaryExpression with a number and a boolean | 5 |
| NotYet: a BinaryExpression with a value and a number | 5 |
| NotYet: a PrefixUnaryExpression on a boolean \| undefined | 5 |
| NotYet: a function returning Path | 5 |
| NotYet: for...in without a proven fixed plain-object origin (arrays, prototypes and absent synthetic fields cannot be enumerated soundly) | 5 |
| NotYet: reading performance | 5 |
| Refused: a generator function | 5 |
| Refused: a method read as a value (trace would lose its object, and this with it) | 5 |
| Refused: a value of type FlowArrayMutation \| FlowAssignment seen as FlowNode, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 5 |
| Refused: a value of type FlowNode seen as FlowNode[] \| FlowNode \| undefined, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 5 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier \| Node seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 5 |
| Refused: a value of type Node seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 5 |
| Refused: a value of type never[] seen as Diagnostic[], which can write Diagnostic where never is read | 5 |
| Refused: an unproven value assigned to a numeric literal or enum member slot CompilerOptions \| BuilderFileEmit | 5 |
| Refused: optional property all in CompilerOptions absent from structural source BuildOptions, which can hide fields | 5 |
| Refused: optional property id in Expression absent from structural source never, which can hide fields | 5 |
| Refused: optional property id in ExternalModuleReference absent from structural source never, which can hide fields | 5 |
| Refused: optional property version in PackageJson absent from structural source PackageJsonPathFields, which can hide fields | 5 |
| Refused: yield (generators) | 5 |
| NotYet: a BinaryExpression with a string and a boolean | 4 |
| NotYet: a computed field name | 4 |
| NotYet: a field of type string \| DiagnosticMessageChain | 4 |
| NotYet: a function returning PackageJson[K] \| undefined | 4 |
| NotYet: a function returning __String \| undefined | 4 |
| NotYet: a function with an optional or rest parameter, as a value | 4 |
| NotYet: a value of type NonNullable<T> | 4 |
| NotYet: a value of type T \| Program | 4 |
| NotYet: an ElementAccessExpression | 4 |
| NotYet: reading Parser | 4 |
| Refused: a function taking readonly Modifier[] \| undefined seen as one taking readonly ModifierLike[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 4 |
| Refused: a method read as a value (fileExists would lose its object, and this with it) | 4 |
| Refused: a method read as a value (getSourceFile would lose its object, and this with it) | 4 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on expr) | 4 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on f) | 4 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on n) | 4 |
| Refused: a value of type CompilerOptionsValue seen as TsConfigSourceFile \| CompilerOptionsValue, which can write string \| number where string is read | 4 |
| Refused: a value of type DiagnosticWithDetachedLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where undefined is read | 4 |
| Refused: a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where SourceFile is read | 4 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 4 |
| Refused: a value of type GeneratedIdentifier \| Identifier seen as Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 4 |
| Refused: a value of type Identifier[][] seen as ModuleExportName[][], which can write ModuleExportName[] where Identifier[] is read | 4 |
| Refused: a value of type Mutable<GeneratedIdentifier> seen as Identifier, which can write EmitNode \| undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 4 |
| Refused: a value of type NodeArray<Statement> seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 4 |
| Refused: a value of type string[] seen as CompilerOptionsValue, which can write string \| number where string is read | 4 |
| Refused: an arbitrary number or a value from another enum assigned to InternalSymbolName.Call; its members are a closed union | 4 |
| Refused: in | 4 |
| Refused: optional property id in Identifier absent from structural source never, which can hide fields | 4 |
| Refused: the void operator | 4 |
| NotYet: Object.entries on a shape not proven by a plain literal or its const binding | 3 |
| NotYet: a case whose type differs from the switch's | 3 |
| NotYet: a declaration directly in a case (wrap the case in a block) | 3 |
| NotYet: a function inside a function (a closure) | 3 |
| NotYet: a function returning CompilerOptionsValue | 3 |
| NotYet: a function returning ResolvedConfigFileName | 3 |
| NotYet: a function returning T \| T[] \| undefined | 3 |
| NotYet: a function returning T \| readonly T[] \| undefined | 3 |
| NotYet: a function value returning union of differently held members | 3 |
| NotYet: a value of type CompilerOptionsValue | 3 |
| NotYet: a value of type K | 3 |
| NotYet: a value of type string \| (void & { __escapedIdentifier: void; }) \| (string & { __escapedIdentifier: void; }) | 3 |
| NotYet: a value of type string \| null \| undefined | 3 |
| NotYet: an array of U | 3 |
| NotYet: optional chaining to .size on a value | 3 |
| Refused: a function taking () => T seen as one taking () => T (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 |
| Refused: a function taking [fileName: string, text: string, writeByteOrderMark: boolean, onError?: ((message: string) => void) \| undefined, sourceFiles?: readonly SourceFile[] \| undefined, data?: WriteFileCallbackData \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 |
| Refused: a function taking number seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 |
| Refused: a function taking string seen as one taking [fileName: string, text: string, writeByteOrderMark: boolean, onError?: ((message: string) => void) \| undefined, sourceFiles?: readonly SourceFile[] \| undefined, data?: WriteFileCallbackData \| undefined] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 |
| Refused: a method read as a value (afterProgramCreate would lose its object, and this with it) | 3 |
| Refused: a method read as a value (afterProgramEmitAndDiagnostics would lose its object, and this with it) | 3 |
| Refused: a method read as a value (getCanonicalFileName would lose its object, and this with it) | 3 |
| Refused: a method read as a value (now would lose its object, and this with it) | 3 |
| Refused: a method read as a value (onWatchStatusChange would lose its object, and this with it) | 3 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on element) | 3 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on entry) | 3 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on info) | 3 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on member) | 3 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on tag) | 3 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on value) | 3 |
| Refused: a union of differently held members as a condition | 3 |
| Refused: a value of type BindingOrAssignmentElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type DiagnosticWithLocation[] seen as Diagnostic[], which can write Diagnostic where DiagnosticWithLocation is read | 3 |
| Refused: a value of type Expression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type FutureSourceFile \| SourceFile seen as Pick<SourceFile, "fileName" \| "impliedNodeFormat">, whose readonly field fileName becomes writable: a readonly field may hold something narrower than string, which a write of string would replace | 3 |
| Refused: a value of type Identifier seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type Map<string, string[] \| never[]> seen as Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>, which can write string where never is read | 3 |
| Refused: a value of type SourceFile seen as SourceFile \| SourceFileLike, which can write readonly number[] \| undefined where readonly number[] is read | 3 |
| Refused: a value of type never[] seen as DiagnosticArguments, which can write string \| number \| boolean \| readonly string[] \| SourceFile \| undefined where never is read | 3 |
| Refused: a value of type never[] seen as ResolvedConfigFileName[], which can write ResolvedConfigFileName where never is read | 3 |
| Refused: a value of type never[] seen as string[] \| never[], which can write string where never is read | 3 |
| Refused: a value of type readonly Extension[][] seen as readonly string[][], which can write string where Extension is read | 3 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint WatchedFileWithIsClosed can be written, so it can write what undefined can't hold | 3 |
| Refused: an arbitrary number or a value from another enum assigned to Extension.Ts; its members are a closed union | 3 |
| Refused: optional property rawText in TemplateLiteralLikeNode absent from structural source NumericLiteral, which can hide fields | 3 |
| Refused: optional property reportsUnnecessary in Diagnostic absent from structural source DiagnosticRelatedInformation, which can hide fields | 3 |
| Refused: optional property resolutionMode in FileReference absent from structural source { preserve: true; }, which can hide fields | 3 |
| SkippedDependency: a dependency function whose body has checker diagnostics (measurement skipped) | 3 |
| NotYet: RegExp with a nonconstant pattern | 2 |
| NotYet: a BinaryExpression with a boolean and a string | 2 |
| NotYet: a BinaryExpression with a string and a number | 2 |
| NotYet: a BinaryExpression with a value and a string | 2 |
| NotYet: a PrefixUnaryExpression on a union of differently held members | 2 |
| NotYet: a call returning void \| undefined | 2 |
| NotYet: a field of type "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number> | 2 |
| NotYet: a field of type boolean \| (() => boolean) \| undefined | 2 |
| NotYet: a field of type true \| Node \| undefined | 2 |
| NotYet: a function returning Extract<ClassDeclaration, Pick<...>> \| Extract<...> | 2 |
| NotYet: a function returning Path \| undefined | 2 |
| NotYet: a function returning U | 2 |
| NotYet: a function returning V | 2 |
| NotYet: a function returning object | 2 |
| NotYet: a function returning undefined | 2 |
| NotYet: a function value returning boolean \| undefined | 2 |
| NotYet: a tagged template other than the intrinsic String.raw | 2 |
| NotYet: a value of type (EmitNode & { autoGenerate: AutoGenerateInfo; }) \| (EmitNode & { autoGenerate: AutoGenerateInfo; }) | 2 |
| NotYet: a value of type BindableStaticNameExpression | 2 |
| NotYet: a value of type T \| T[] | 2 |
| NotYet: a value of type TInArray | 2 |
| NotYet: a value of type WrappedExpression<AnonymousFunctionDefinition> | 2 |
| NotYet: new a ParenthesizedExpression | 2 |
| NotYet: reading getOptionsNameMap | 2 |
| NotYet: reading tracingEnabled | 2 |
| NotYet: replacing a represented method at runtime | 2 |
| Refused: a function taking ((node: Node) => VisitResult<Node>) \| undefined seen as one taking ((node: Node) => VisitResult<Node \| undefined>) \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking BinaryOperator seen as one taking SyntaxKind (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking Expression seen as one taking Expression \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking Expression[] seen as one taking readonly Expression[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking GeneratedIdentifierFlags seen as one taking GeneratedIdentifierFlags \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking NodeFlags seen as one taking NodeFlags \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking T seen as one taking Expression (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking [fileName: string, languageVersionOrOptions: CreateSourceFileOptions \| ScriptTarget, onError?: ((message: string) => void) \| undefined, shouldCreateNewSourceFile?: boolean \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking string seen as one taking [fileName: string] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking string seen as one taking any[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a method read as a value (clearTimeout would lose its object, and this with it) | 2 |
| Refused: a method read as a value (createDirectory would lose its object, and this with it) | 2 |
| Refused: a method read as a value (emitNodeWithNotification would lose its object, and this with it) | 2 |
| Refused: a method read as a value (enableCPUProfiler would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getBuildInfo would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getEnvironmentVariable would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getGlobalTypingsCacheLocation would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getParsedCommandLine would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getSemanticDiagnosticsOfNextAffectedFile would lose its object, and this with it) | 2 |
| Refused: a method read as a value (hasGlobalName would lose its object, and this with it) | 2 |
| Refused: a method read as a value (isEmitNotificationEnabled would lose its object, and this with it) | 2 |
| Refused: a method read as a value (parenthesizeBranchOfConditionalExpression would lose its object, and this with it) | 2 |
| Refused: a method read as a value (parenthesizeConstituentTypesOfIntersectionType would lose its object, and this with it) | 2 |
| Refused: a method read as a value (parenthesizeConstituentTypesOfUnionType would lose its object, and this with it) | 2 |
| Refused: a method read as a value (parenthesizeNonArrayTypeOfPostfixType would lose its object, and this with it) | 2 |
| Refused: a method read as a value (setPrototypeOf would lose its object, and this with it) | 2 |
| Refused: a method read as a value (setTimeout would lose its object, and this with it) | 2 |
| Refused: a method read as a value (substituteNode would lose its object, and this with it) | 2 |
| Refused: a method read as a value (toKey would lose its object, and this with it) | 2 |
| Refused: a method read as a value (trackSymbol would lose its object, and this with it) | 2 |
| Refused: a method read as a value (useCaseSensitiveFileNames would lose its object, and this with it) | 2 |
| Refused: a method read as a value (watchDirectory would lose its object, and this with it) | 2 |
| Refused: a method read as a value (watchFile would lose its object, and this with it) | 2 |
| Refused: a method read as a value (writeFile would lose its object, and this with it) | 2 |
| Refused: a number \| undefined as a condition | 2 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on declaration) | 2 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on predicate) | 2 |
| Refused: a value of type (identifierOrPrivateName: Identifier \| PrivateIdentifier) => string seen as (name: GeneratedIdentifier \| GeneratedPrivateIdentifier) => string, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 2 |
| Refused: a value of type ArrayLiteralExpression \| AssignmentExpression<EqualsToken> \| BindingElement \| ElementAccessExpression \| ... 8 more ... \| VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Block seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type BlockLike seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ClassDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ClassStaticBlockDeclaration seen as Mutable<ClassStaticBlockDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ConstructorDeclaration seen as Mutable<ConstructorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Diagnostic seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 2 |
| Refused: a value of type DiagnosticWithLocation[] \| undefined seen as DiagnosticRelatedInformation[] \| undefined, which can write DiagnosticRelatedInformation where DiagnosticWithLocation is read | 2 |
| Refused: a value of type Expression seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type FlowCall seen as FlowNode, which can write BinaryExpression \| CallExpression where CallExpression is read | 2 |
| Refused: a value of type FutureSourceFile \| SourceFile seen as Pick<SourceFile, "fileName" \| "impliedNodeFormat" \| "packageJsonScope">, whose readonly field fileName becomes writable: a readonly field may hold something narrower than string, which a write of string would replace | 2 |
| Refused: a value of type GeneratedIdentifier seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 2 |
| Refused: a value of type GeneratedIdentifier seen as Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 2 |
| Refused: a value of type Identifier seen as Mutable<Identifier>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Identifier[] \| undefined seen as ModuleExportName[] \| undefined, which can write ModuleExportName where Identifier is read | 2 |
| Refused: a value of type ImportTypeAssertionContainer seen as Mutable<ImportTypeAssertionContainer>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type MissingDeclaration seen as Mutable<MissingDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ModifierLike seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ModuleSpecifierResolutionHost & ModuleResolutionHost seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where (() => boolean) & (boolean \| (() => boolean) \| undefined) is read | 2 |
| Refused: a value of type Node seen as Node \| TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Node \| TextRange seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type NodeArray<Statement> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ParameterDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type PropertySignature seen as Mutable<PropertySignature>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type SourceFile seen as EmitNode \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of EmitFlags would replace | 2 |
| Refused: a value of type SourceFile seen as Mutable<SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Statement seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Symbol \| undefined seen as Declaration \| undefined, which can write number \| undefined where number is read | 2 |
| Refused: a value of type Symbol[] seen as unknown[], which can write unknown where Symbol is read | 2 |
| Refused: a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Type[] seen as unknown[], which can write unknown where Type is read | 2 |
| Refused: a value of type any seen as T, a type parameter whose constraint Node can be written, so it can write what any can't hold | 2 |
| Refused: a value of type never[] seen as ChildDirectoryWatcher[], which can write ChildDirectoryWatcher where never is read | 2 |
| Refused: a value of type never[] seen as SourceFile[], which can write SourceFile where never is read | 2 |
| Refused: a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[] \| undefined, which can write SourceFile \| undefined where SourceFile is read | 2 |
| Refused: a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 2 |
| Refused: a value of type string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] seen as unknown[], which can write unknown where string is read | 2 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"FlowFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof FlowFlags is read | 2 |
| Refused: an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | 2 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.FunctionExpression | 2 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NoSubstitutionTemplateLiteral \| SyntaxKind.TemplateHead \| SyntaxKind.TemplateMiddle \| SyntaxKind.TemplateTail | 2 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.TypeKeyword | 2 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.WithKeyword \| SyntaxKind.AssertKeyword | 2 |
| Refused: delete | 2 |
| Refused: optional property all in CompilerOptions absent from structural source {}, which can hide fields | 2 |
| Refused: optional property equalsToken in ShorthandPropertyAssignment absent from structural source never, which can hide fields | 2 |
| Refused: optional property id in LiteralExpression & StringLiteral absent from structural source never, which can hide fields | 2 |
| Refused: optional property modifiers in GetAccessorDeclaration absent from structural source never, which can hide fields | 2 |
| Refused: optional property omitTrailingSemicolon in PrinterOptions absent from structural source CompilerOptions, which can hide fields | 2 |
| Refused: optional property packageJsonScope in Pick<SourceFile, "fileName" \| "impliedNodeFormat" \| "packageJsonScope"> absent from structural source Pick<SourceFile, "fileName" \| "impliedNodeFormat">, which can hide fields | 2 |
| Refused: optional property templateFlags in NoSubstitutionTemplateLiteral absent from structural source never, which can hide fields | 2 |
| Refused: optional property textSourceNode in StringLiteral absent from structural source never, which can hide fields | 2 |
| panic: Unhandled case in Node.Text: *ast.ComputedPropertyName | 2 |
| NotYet: .length on a value | 1 |
| NotYet: ?. to a number, which would be number \| undefined | 1 |
| NotYet: ?.[] on a value | 1 |
| NotYet: Object.assign on a shape not proven by a plain literal or its const binding | 1 |
| NotYet: a BinaryExpression with a boolean and a boolean \| undefined | 1 |
| NotYet: a BinaryExpression with a boolean and a number | 1 |
| NotYet: a BinaryExpression with a boolean \| undefined and a boolean | 1 |
| NotYet: a BinaryExpression with a number and a string | 1 |
| NotYet: a BinaryExpression with a union of differently held members and a union of differently held members | 1 |
| NotYet: a ClassExpression | 1 |
| NotYet: a Map of CompilerOptionsValue | 1 |
| NotYet: a Map of T | 1 |
| NotYet: a PostfixUnaryExpression | 1 |
| NotYet: a YieldExpression as a statement | 1 |
| NotYet: a boolean \| undefined variable a function value captures | 1 |
| NotYet: a class instantiated with TOuterState | 1 |
| NotYet: a destructured name held otherwise than its field | 1 |
| NotYet: a destructured name that isn't plain | 1 |
| NotYet: a destructured parameter beside a parameter with a default | 1 |
| NotYet: a field from a boolean \| undefined variable | 1 |
| NotYet: a field of type AnyBuildOrder \| undefined | 1 |
| NotYet: a field of type NodeArray<ParameterDeclaration> \| readonly JSDocParameterTag[] | 1 |
| NotYet: a field of type false \| VersionPaths \| undefined | 1 |
| NotYet: a field of type string \| false \| undefined | 1 |
| NotYet: a field of type string \| number \| undefined | 1 |
| NotYet: a function returning (AssignmentExpression<EqualsToken> & { readonly left: GeneratedIdentifier; }) \| undefined | 1 |
| NotYet: a function returning (ConstructorDeclaration & { body: Block; }) \| undefined | 1 |
| NotYet: a function returning AnyValidImportOrReExport | 1 |
| NotYet: a function returning AnyValidImportOrReExport \| undefined | 1 |
| NotYet: a function returning CanonicalKey | 1 |
| NotYet: a function returning ClassNamedEvaluationHelperBlock | 1 |
| NotYet: a function returning ClassThisAssignmentBlock | 1 |
| NotYet: a function returning ExpressionWithTypeArguments & { readonly expression: Identifier \| PropertyAccessEntityNameExpression; } | 1 |
| NotYet: a function returning Extract<AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; }, Pick<...>> \| ... 7 more ... \| Extract<...> | 1 |
| NotYet: a function returning HasJSDoc \| undefined | 1 |
| NotYet: a function returning MemberName \| (Expression & (NumericLiteral \| StringLiteralLike)) | 1 |
| NotYet: a function returning ModeAwareCacheKey | 1 |
| NotYet: a function returning Node \| (TIn & undefined) \| (TVisited & undefined) | 1 |
| NotYet: a function returning NodeArray<Node> \| (TInArray & undefined) | 1 |
| NotYet: a function returning NodeArray<TOut> \| (TInArray & undefined) | 1 |
| NotYet: a function returning PathPathComponents | 1 |
| NotYet: a function returning ResolvedConfigFilePath | 1 |
| NotYet: a function returning T \| T[] | 1 |
| NotYet: a function returning T \| readonly T[] | 1 |
| NotYet: a function returning T1 & T2 | 1 |
| NotYet: a function returning TEntry \| undefined | 1 |
| NotYet: a function returning TOut | 1 |
| NotYet: a function returning TOut \| (TIn & undefined) \| (TVisited & undefined) | 1 |
| NotYet: a function returning TOut \| undefined | 1 |
| NotYet: a function returning TPrivateEntry \| undefined | 1 |
| NotYet: a function returning object \| undefined | 1 |
| NotYet: a function returning readonly Node[] \| (TInArray & undefined) | 1 |
| NotYet: a function returning readonly TOut[] \| (TInArray & undefined) | 1 |
| NotYet: a function returning string \| object | 1 |
| NotYet: a function returning unknown | 1 |
| NotYet: a function returning void \| SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> | 1 |
| NotYet: a function returning void \| SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> \| WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> | 1 |
| NotYet: a function returning void \| WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> | 1 |
| NotYet: a literal method through a view that erases its receiver | 1 |
| NotYet: a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables | 1 |
| NotYet: a number \| undefined argument to substring | 1 |
| NotYet: a template interpolating an object, an array, a map, a function or undefined | 1 |
| NotYet: a value of type "" \| ResolvedConfigFileName \| undefined | 1 |
| NotYet: a value of type (AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; } & BinaryExpression) \| (... & ... 1 more ... & BinaryExpression) | 1 |
| NotYet: a value of type (ConstructorDeclaration & { body: Block; }) \| undefined | 1 |
| NotYet: a value of type (ModuleDeclaration & { name: StringLiteral; }) \| undefined | 1 |
| NotYet: a value of type AccessorDeclaration & { readonly name: BigIntLiteral \| ComputedPropertyName \| Identifier \| NoSubstitutionTemplateLiteral \| NumericLiteral \| StringLiteral; } | 1 |
| NotYet: a value of type BindingElement & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type ClassNamedEvaluationHelperBlock | 1 |
| NotYet: a value of type EmitNode & { autoGenerate: AutoGenerateInfo; } | 1 |
| NotYet: a value of type EntityNameExpression \| (LeftHandSideExpression & BindableStaticNameExpression) | 1 |
| NotYet: a value of type ExportAssignment & { readonly expression: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type HasJSDoc | 1 |
| NotYet: a value of type HasJSDoc \| undefined | 1 |
| NotYet: a value of type IncludeTypeSpaceImports | 1 |
| NotYet: a value of type IncrementalBuildInfoFilePendingEmit | 1 |
| NotYet: a value of type IncrementalMultiFileEmitBuildInfoFileInfo | 1 |
| NotYet: a value of type JSDocImportTag \| CanHaveModuleSpecifier | 1 |
| NotYet: a value of type NamedEvaluation | 1 |
| NotYet: a value of type ParameterDeclaration & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type PropertyAssignment & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type PropertyDeclaration & { readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type RequireOrImportCall | 1 |
| NotYet: a value of type ResolvedConfigFileName \| undefined | 1 |
| NotYet: a value of type ResolvedModuleWithFailedLookupLocations & ResolvedTypeReferenceDirectiveWithFailedLookupLocations | 1 |
| NotYet: a value of type ShorthandPropertyAssignment & { readonly objectAssignmentInitializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type SourceFile | 1 |
| NotYet: a value of type T \| readonly T[] | 1 |
| NotYet: a value of type T1 | 1 |
| NotYet: a value of type TData | 1 |
| NotYet: a value of type TEntry | 1 |
| NotYet: a value of type T["kind"] | 1 |
| NotYet: a value of type TypeNode & LiteralTypeNode & { readonly literal: StringLiteral; } | 1 |
| NotYet: a value of type TypeParameterDeclaration & { parent: JSDocTemplateTag; } | 1 |
| NotYet: a value of type U | 1 |
| NotYet: a value of type U \| readonly U[] \| undefined | 1 |
| NotYet: a value of type V | 1 |
| NotYet: a value of type VariableDeclaration & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type WatchFactoryHost & { trace?(s: string): void; } | 1 |
| NotYet: a value of type WrappedExpression<T> | 1 |
| NotYet: a value of type __String & string | 1 |
| NotYet: a value of type false \| RegExpExecArray \| null | 1 |
| NotYet: a value of type object \| undefined | 1 |
| NotYet: a value of type undefined | 1 |
| NotYet: a void call used as a value | 1 |
| NotYet: an array of T \| U | 1 |
| NotYet: an array of V | 1 |
| NotYet: an array of unknown | 1 |
| NotYet: an optional chain longer than one step | 1 |
| NotYet: assigning a field of a value | 1 |
| NotYet: indexOf on a tuple (a tuple is held as an object, not an array, so far; write it as an array where it's made) | 1 |
| NotYet: lastIndexOf with these arguments | 1 |
| NotYet: new Map from something that isn't [key, value] pairs | 1 |
| NotYet: reading BuilderState | 1 |
| NotYet: reading IncrementalParser | 1 |
| NotYet: reading addAggregateStatistic | 1 |
| NotYet: reading addOutput | 1 |
| NotYet: reading captureMapping | 1 |
| NotYet: reading convertToFunctionBlock | 1 |
| NotYet: reading createBaseSourceFileNode | 1 |
| NotYet: reading createIntlCollatorStringComparer | 1 |
| NotYet: reading createPollingIntervalQueue | 1 |
| NotYet: reading enter | 1 |
| NotYet: reading getAccessorNameVisibilityError | 1 |
| NotYet: reading getPackageJsonInfo | 1 |
| NotYet: reading getParenthesizeLeftSideOfBinaryForOperator | 1 |
| NotYet: reading getPreferredEnding | 1 |
| NotYet: reading getSymbolWalker | 1 |
| NotYet: reading getUnusedExpectations | 1 |
| NotYet: reading getValueCandidate | 1 |
| NotYet: reading getVariableDeclarationTypeVisibilityError | 1 |
| NotYet: reading inferPreference | 1 |
| NotYet: reading lookupFromPackageJson | 1 |
| NotYet: reading needJsx | 1 |
| NotYet: reading optionDependsOnRecursive | 1 |
| NotYet: reading parseStrings | 1 |
| NotYet: reading serializeTypeOfDeclaration | 1 |
| NotYet: reading transformSourceFile | 1 |
| NotYet: reading transformSourceFileOrBundle | 1 |
| NotYet: spreading an array of other elements | 1 |
| NotYet: storing any in a field | 1 |
| NotYet: storing string \| number in a field | 1 |
| NotYet: storing true \| Node \| undefined in a field | 1 |
| Refused: &&= | 1 |
| Refused: Object.defineProperty | 1 |
| Refused: a function taking (symbol: Symbol) => boolean seen as one taking ((symbol: Symbol) => boolean) \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking BinaryOperatorToken seen as one taking BinaryOperatorToken \| BinaryOperator (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking JSDocTypeExpression \| undefined seen as one taking JSDocTypeExpression \| JSDocTypeLiteral \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking LogLevel seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking Node seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking Node \| undefined seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking NodeArray<Expression> seen as one taking readonly Expression[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking NodeArray<T> seen as one taking NodeArray<T> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking NodeArray<T> seen as one taking NodeArray<T> \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking NodeArray<TypeNode> \| undefined seen as one taking readonly TypeNode[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking OuterExpressionKinds seen as one taking OuterExpressionKinds \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking PollingInterval seen as one taking number \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking SourceFile seen as one taking [file: SourceFile, options: CompilerOptions] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking T seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking TIn seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking TokenFlags seen as one taking TokenFlags \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking TypeNode seen as one taking Node (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking Visitor seen as one taking Visitor<TIn, Node \| undefined> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [callback: (...args: any[]) => void, ms: number, ...args: any[]] seen as one taking (...args: any[]) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [fileName: string] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [node: ConstructorTypeNode, typeParameters: NodeArray<TypeParameterDeclaration> \| undefined, parameters: NodeArray<ParameterDeclaration>, type: TypeNode] \| ... seen as one taking ConstructorTypeNode (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [path: string, callback: DirectoryWatcherCallback, recursive?: boolean \| undefined, options?: WatchOptions \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [path: string, callback: FileWatcherCallback, pollingInterval?: number \| undefined, options?: WatchOptions \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [timeoutId: any] seen as one taking unknown (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [typeParameters: readonly TypeParameterDeclaration[] \| undefined, parameters: readonly ParameterDeclaration[], type: TypeNode] \| [modifiers: readonly Modifier[] \| undefined, typeParameters: ... \| undefined, parameters: ..., type: TypeNode] seen as one taking readonly Modifier[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [writeByteOrderMark: boolean, onError?: ((message: string) => void) \| undefined, sourceFiles?: readonly SourceFile[] \| undefined, data?: WriteFileCallbackData \| undefined] seen as one taking boolean (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking never seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking number seen as one taking any[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking number seen as one taking number \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking readonly ParameterDeclaration[] seen as one taking readonly ParameterDeclaration[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking readonly T[] \| undefined seen as one taking readonly T[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking string seen as one taking string \| MemberName (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking string \| undefined seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a method read as a value (add would lose its object, and this with it) | 1 |
| Refused: a method read as a value (base64decode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (base64encode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (clearScreen would lose its object, and this with it) | 1 |
| Refused: a method read as a value (compare would lose its object, and this with it) | 1 |
| Refused: a method read as a value (convertToArrayAssignmentElement would lose its object, and this with it) | 1 |
| Refused: a method read as a value (convertToObjectAssignmentElement would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createComma would lose its object, and this with it) | 1 |
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
| Refused: a method read as a value (directoryExists would lose its object, and this with it) | 1 |
| Refused: a method read as a value (emit would lose its object, and this with it) | 1 |
| Refused: a method read as a value (emitBuildInfo would lose its object, and this with it) | 1 |
| Refused: a method read as a value (emitNextAffectedFile would lose its object, and this with it) | 1 |
| Refused: a method read as a value (fill would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getAllDependencies would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getCurrentDirectory would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getDeclarationDiagnostics would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getDirectories would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getMemoryUsage would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getModifiedTime would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getNearestAncestorDirectoryWithPackageJson would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getOrCreateCacheForModuleName would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getSemanticDiagnostics would lose its object, and this with it) | 1 |
| Refused: a method read as a value (globalCacheResolutionModuleName would lose its object, and this with it) | 1 |
| Refused: a method read as a value (hasChangedEmitSignature would lose its object, and this with it) | 1 |
| Refused: a method read as a value (hasOwnProperty would lose its object, and this with it) | 1 |
| Refused: a method read as a value (log would lose its object, and this with it) | 1 |
| Refused: a method read as a value (nonEscapingWrite would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeCheckTypeOfConditionalType would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeConciseBodyOfArrowFunction would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeConditionOfConditionalExpression would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeConstituentTypeOfIntersectionType would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeConstituentTypeOfUnionType would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeElementTypeOfTupleType would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeExpressionOfComputedPropertyName would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeExpressionOfExportDefault would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeExpressionOfExpressionStatement would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeExpressionOfNew would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeExtendsTypeOfConditionalType would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeLeadingTypeArgument would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeOperandOfPostfixUnary would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeOperandOfReadonlyTypeOperator would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeOperandOfTypeOperator would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeTypeOfOptionalType would lose its object, and this with it) | 1 |
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
| Refused: a method read as a value (toString would lose its object, and this with it) | 1 |
| Refused: a method read as a value (writeOutputIsTTY would lose its object, and this with it) | 1 |
| Refused: a type predicate whose return is not proven (asserts cond needs a boolean parameter) | 1 |
| Refused: a type predicate whose return is not proven (false return can still contain LateBoundDeclaration) | 1 |
| Refused: a type predicate whose return is not proven (normal return has not narrowed value to NonNullable<T>) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on array) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on buildOrder) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on diagnostic) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on file) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on location) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on option) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on options) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on p) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on program) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on sourceFile) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on symbol) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on type) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on x) | 1 |
| Refused: a type predicate whose return is not proven (the body's true narrowing does not match null \| undefined) | 1 |
| Refused: a type predicate whose return is not proven (the predicate parameter is assigned) | 1 |
| Refused: a type predicate whose return is not proven (true return narrows to MappedPosition, not SourceMappedPosition) | 1 |
| Refused: a type predicate whose return is not proven (true return narrows to Mapping, not SourceMapping) | 1 |
| Refused: a type predicate whose return is not proven (true return narrows to ReusableBuilderProgramState, not BuilderProgramStateWithDefinedProgram) | 1 |
| Refused: a value of type (AbstractKeyword \| AccessorKeyword \| AsyncKeyword \| ConstKeyword \| DeclareKeyword \| Decorator \| ... 7 more ... \| StaticKeyword)[] \| undefined seen as Node[] \| undefined, which can write Node where AbstractKeyword \| AccessorKeyword \| AsyncKeyword \| ConstKeyword \| DeclareKeyword \| Decorator \| ... 7 more ... \| StaticKeyword is read | 1 |
| Refused: a value of type (baseDir: string, moduleName: string) => { module: any; modulePath: string; error: undefined; } \| { module: undefined; modulePath: undefined; error: unknown; } seen as (baseDir: string, moduleName: string) => ModuleImportResult, which can write string \| undefined where string is read | 1 |
| Refused: a value of type (left: MappedPosition, right: MappedPosition) => boolean seen as EqualityComparer<SourceMappedPosition> \| undefined, which can write string \| undefined where string is read | 1 |
| Refused: a value of type (node: CommentRange) => boolean seen as (value: SynthesizedComment) => boolean, which can write number where -1 is read | 1 |
| Refused: a value of type (recordTempVariable: ((node: Identifier) => void) \| undefined, reservedInNestedScopes?: boolean \| undefined, prefix?: string \| GeneratedNamePart \| undefined, suffix?: string \| undefined) => GeneratedIdentifier seen as { (recordTempVariable: ((node: Identifier) => void) \| undefined, reservedInNestedScopes?: boolean \| undefined): Identifier; (recordTempVariable: ((node: Identifier) => void) \| undefined, reservedInNestedScopes?: boolean \| undefined, prefix?: string \| ... 1 more ... \| undefined, suffix?: string \| undefined): Identifi..., whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type (symbol: Symbol) => { visitedTypes: Type[]; visitedSymbols: Symbol[]; } seen as (root: Symbol) => { visitedTypes: readonly Type[]; visitedSymbols: readonly Symbol[]; }, which can write readonly Type[] where Type[] is read | 1 |
| Refused: a value of type (symbolAccessibilityResult: SymbolAccessibilityResult) => { diagnosticMessage: DiagnosticMessage; errorNode: DeclarationDiagnosticProducing; typeName: DeclarationName \| undefined; } \| undefined seen as (symbolAccessibilityResult: SymbolAccessibilityResult) => SymbolAccessibilityDiagnostic \| undefined, which can write Node where DeclarationDiagnosticProducing is read | 1 |
| Refused: a value of type (type: Type) => { visitedTypes: Type[]; visitedSymbols: Symbol[]; } seen as (root: Type) => { visitedTypes: readonly Type[]; visitedSymbols: readonly Symbol[]; }, which can write readonly Type[] where Type[] is read | 1 |
| Refused: a value of type AssertClause seen as Mutable<AssertClause>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type AssertEntry seen as Mutable<AssertEntry>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type BinaryExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Block seen as Mutable<Block>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type BreakStatement seen as Mutable<BreakStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type CallSignatureDeclaration seen as Mutable<CallSignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type CaseBlock seen as Mutable<CaseBlock>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Children seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ClassDeclaration \| FunctionDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ClassElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type CompilerHost seen as CompilerHostLikeForCache, which can write WriteFileCallback \| undefined where WriteFileCallback is read | 1 |
| Refused: a value of type CompilerOptions & { types: string[]; } seen as CompilerOptions, which can write string[] \| undefined where string[] is read | 1 |
| Refused: a value of type ConstructSignatureDeclaration seen as Mutable<ConstructSignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ConstructorTypeNode seen as Mutable<ConstructorTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ContinueStatement seen as Mutable<ContinueStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type DestructuringAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type DiagnosticWithLocation \| undefined seen as Diagnostic \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 |
| Refused: a value of type ElementAccessExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type EnumDeclaration \| ModuleDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ExportAssignment seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Expression \| GeneratedIdentifier seen as Expression \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type Expression \| GeneratedIdentifier seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type ExpressionStatement seen as Mutable<ExpressionStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Extension[] seen as string[], which can write string where Extension is read | 1 |
| Refused: a value of type FlowArrayMutation \| FlowAssignment \| FlowCall \| FlowCondition \| FlowLabel \| FlowReduceLabel \| FlowStart \| FlowUnreachable seen as FlowNode, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 1 |
| Refused: a value of type FunctionTypeNode seen as Mutable<FunctionTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type GeneratedIdentifier seen as Expression \| GeneratedIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier seen as Node \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as GeneratedIdentifier \| GeneratedPrivateIdentifier \| Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as Identifier \| PrivateIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as Node \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier \| Identifier \| PrivateIdentifier seen as Identifier \| PrivateIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| Identifier seen as Identifier \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedPrivateIdentifier seen as Node \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GetAccessorDeclaration \| SetAccessorDeclaration seen as Mutable<AccessorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Identifier seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Identifier seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Identifier[] seen as ModuleExportName[] \| undefined, which can write ModuleExportName where Identifier is read | 1 |
| Refused: a value of type ImportAttribute seen as Mutable<ImportAttribute>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ImportAttributes seen as Mutable<ImportAttributes>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ImportClause seen as Mutable<ImportClause>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ImportDeclaration seen as Mutable<ImportDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ImportEqualsDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ImportTypeNode seen as Mutable<ImportTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type IndexSignatureDeclaration seen as Mutable<IndexSignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocAugmentsTag seen as Mutable<JSDocAugmentsTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocCallbackTag seen as Mutable<JSDocCallbackTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocFunctionType seen as Mutable<JSDocFunctionType>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocImplementsTag seen as Mutable<JSDocImplementsTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocImportTag seen as Mutable<JSDocImportTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocLink seen as Mutable<JSDocLink>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocLinkCode seen as Mutable<JSDocLinkCode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocLinkPlain seen as Mutable<JSDocLinkPlain>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocNameReference seen as Mutable<JSDocNameReference>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocOverloadTag seen as Mutable<JSDocOverloadTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocParameterTag seen as Mutable<JSDocParameterTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocPropertyTag seen as Mutable<JSDocPropertyTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocSeeTag seen as Mutable<JSDocSeeTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocSignature seen as Mutable<JSDocSignature>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocTemplateTag seen as Mutable<JSDocTemplateTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocText seen as Mutable<JSDocText>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocTypeExpression seen as Mutable<JSDocTypeExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocTypeLiteral seen as Mutable<JSDocTypeLiteral>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocTypedefTag seen as Mutable<JSDocTypedefTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocUnknownTag seen as Mutable<JSDocUnknownTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
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
| Refused: a value of type Map<string, WatchDirectoryKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where WatchDirectoryKind is read | 1 |
| Refused: a value of type Map<string, WatchFileKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where WatchFileKind is read | 1 |
| Refused: a value of type Map<string, string> seen as Map<string, string \| number>, which can write string \| number where string is read | 1 |
| Refused: a value of type MappedTypeNode seen as Mutable<MappedTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type MethodDeclaration seen as Mutable<MethodDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type MethodDeclaration \| PropertyAssignment \| AccessorDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ModuleDeclaration seen as Mutable<ModuleDeclaration \| SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ModuleExportName seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ModuleName seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ModuleResolutionState & { host: GetPackageJsonEntrypointsHost; } seen as ModuleResolutionState, which can write ModuleResolutionHost where ModuleResolutionHost & GetPackageJsonEntrypointsHost is read | 1 |
| Refused: a value of type Mutable<NoSubstitutionTemplateLiteral> seen as Mutable<TemplateLiteralLikeNode>, which can write SyntaxKind where SyntaxKind.NoSubstitutionTemplateLiteral is read | 1 |
| Refused: a value of type Mutable<T> seen as T, a type parameter whose constraint JSDocType & { readonly type: TypeNode \| undefined; readonly postfix: boolean; } can be written, so it can write what Mutable<T> can't hold | 1 |
| Refused: a value of type Mutable<T> seen as T, a type parameter whose constraint JSDocType & { readonly type: TypeNode \| undefined; } can be written, so it can write what Mutable<T> can't hold | 1 |
| Refused: a value of type Mutable<T> seen as T, a type parameter whose constraint Node can be written, so it can write what Mutable<T> can't hold | 1 |
| Refused: a value of type NamedImports seen as Mutable<NamedImports>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type NamespaceExport seen as Mutable<NamespaceExport>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type NamespaceExportDeclaration seen as Mutable<NamespaceExportDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type NamespaceImport seen as Mutable<NamespaceImport>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Node seen as Node \| SourceMapRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Node \| SourceMapRange seen as SourceMapRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Node \| undefined seen as EmitNode \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of EmitFlags would replace | 1 |
| Refused: a value of type NonNullExpression seen as Mutable<NonNullExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ObjectBindingOrAssignmentPattern seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type OptionalTypeNode seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ParameterDeclaration \| PropertyDeclaration \| PropertySignature \| SignatureDeclaration seen as Mutable<ParameterDeclaration \| PropertyDeclaration \| PropertySignature \| SignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ParameterDeclaration \| VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ParameterDeclaration \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type ParseConfigHost seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where boolean is read | 1 |
| Refused: a value of type Path[] seen as string[], which can write string where Path is read | 1 |
| Refused: a value of type PropertyAssignment seen as Mutable<PropertyAssignment \| ShorthandPropertyAssignment>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Readonly<BuilderState> \| undefined seen as BuilderState \| undefined, whose readonly field fileInfos becomes writable: a readonly field may hold something narrower than Map<Path, FileInfo>, which a write of Map<Path, FileInfo> would replace | 1 |
| Refused: a value of type ReferencedFile & { kind: FileIncludeKind.LibReferenceDirective; } seen as ReferencedFile, which can write ReferencedFileKind where FileIncludeKind.LibReferenceDirective is read | 1 |
| Refused: a value of type ReportFileInError[] seen as (ReportFileInError \| undefined)[], which can write ReportFileInError \| undefined where ReportFileInError is read | 1 |
| Refused: a value of type ResolvedModuleFull \| undefined seen as { path: string; originalPath: string \| true; extension: string; packageId: PackageId \| undefined; resolvedUsingTsExtension: boolean \| undefined; } \| undefined, whose readonly field originalPath becomes writable: a readonly field may hold something narrower than string \| undefined, which a write of string \| true would replace | 1 |
| Refused: a value of type ReturnStatement seen as Mutable<ReturnStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type SearchResult<Resolved> seen as { value: { resolved: Resolved; isExternalLibraryImport: true; } \| undefined; } \| undefined, whose readonly field value becomes writable: a readonly field may hold something narrower than Resolved \| undefined, which a write of { resolved: Resolved; isExternalLibraryImport: true; } \| undefined would replace | 1 |
| Refused: a value of type SetAccessorDeclaration seen as Mutable<SetAccessorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ShorthandPropertyAssignment seen as Mutable<PropertyAssignment \| ShorthandPropertyAssignment>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type SourceFile seen as Mutable<ModuleDeclaration \| SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type SourceFile \| undefined seen as FileReasonToChainCache \| undefined, which can write DiagnosticMessageChain[] \| undefined where RedirectInfo \| undefined is read | 1 |
| Refused: a value of type SourceMapSource seen as SourceFileLike, which can write readonly number[] \| undefined where readonly number[] is read | 1 |
| Refused: a value of type Statement[] seen as Node[], which can write Node where Statement is read | 1 |
| Refused: a value of type StringLiteral seen as Mutable<StringLiteral>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type SwitchStatement seen as Mutable<SwitchStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type System seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where boolean is read | 1 |
| Refused: a value of type T seen as T, a type parameter whose constraint ClassDeclaration \| ClassExpression \| GetAccessorDeclaration \| MethodDeclaration \| ParameterDeclaration \| PropertyDeclaration \| SetAccessorDeclaration can be written, so it can write what T can't hold | 1 |
| Refused: a value of type T seen as T, a type parameter whose constraint HasModifiers can be written, so it can write what T can't hold | 1 |
| Refused: a value of type T seen as T, a type parameter whose constraint MethodDeclaration \| MethodSignature \| PropertyAssignment \| PropertyDeclaration \| PropertySignature \| AccessorDeclaration can be written, so it can write what T can't hold | 1 |
| Refused: a value of type T seen as T, a type parameter whose constraint ModifierSyntaxKind can be written, so it can write what T can't hold | 1 |
| Refused: a value of type T seen as T, a type parameter whose constraint Node can be written, so it can write what T can't hold | 1 |
| Refused: a value of type T seen as T, a type parameter whose constraint Node \| undefined can be written, so it can write what T can't hold | 1 |
| Refused: a value of type TKind seen as TKind, a type parameter whose constraint KeywordTypeSyntaxKind can be written, so it can write what TKind can't hold | 1 |
| Refused: a value of type T[] seen as (T \| undefined)[], which can write T \| undefined where T is read | 1 |
| Refused: a value of type T[][] seen as (readonly T[])[], which can write readonly T[] where T[] is read | 1 |
| Refused: a value of type TaggedTemplateExpression seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type TsConfigSourceFile \| CompilerOptionsValue seen as string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined, which can write string \| number where string is read | 1 |
| Refused: a value of type TypeOperatorNode seen as Mutable<TypeOperatorNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type TypeParameterDeclaration seen as Mutable<TypeParameterDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type VariableDeclaration \| DestructuringAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type VariableDeclaration[] seen as VariableStatement[], which can write VariableStatement where VariableDeclaration is read | 1 |
| Refused: a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read | 1 |
| Refused: a value of type YieldExpression seen as Mutable<YieldExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type never seen as T, a type parameter whose constraint Node can be written, so it can write what never can't hold | 1 |
| Refused: a value of type never seen as T, a type parameter whose constraint any can be written, so it can write what never can't hold | 1 |
| Refused: a value of type never seen as U, a type parameter whose constraint {} can be written, so it can write what never can't hold | 1 |
| Refused: a value of type never[] seen as (JSDoc \| JSDocTag)[], which can write JSDoc \| JSDocTag where never is read | 1 |
| Refused: a value of type never[] seen as (JSDocCallbackTag \| JSDocEnumTag \| JSDocTypedefTag)[], which can write JSDocCallbackTag \| JSDocEnumTag \| JSDocTypedefTag where never is read | 1 |
| Refused: a value of type never[] seen as (SourceMapRange \| undefined)[], which can write SourceMapRange \| undefined where never is read | 1 |
| Refused: a value of type never[] seen as CommentRange[], which can write CommentRange where never is read | 1 |
| Refused: a value of type never[] seen as Comparator[][], which can write Comparator[] where never is read | 1 |
| Refused: a value of type never[] seen as Declaration[], which can write Declaration where never is read | 1 |
| Refused: a value of type never[] seen as Expression[], which can write Expression where never is read | 1 |
| Refused: a value of type never[] seen as FlowNode[], which can write FlowNode where never is read | 1 |
| Refused: a value of type never[] seen as JSDocImportTag[], which can write JSDocImportTag where never is read | 1 |
| Refused: a value of type never[] seen as ProjectReference[], which can write ProjectReference where never is read | 1 |
| Refused: a value of type never[] seen as RequireOrImportCall[], which can write RequireOrImportCall where never is read | 1 |
| Refused: a value of type never[] seen as ResolvedProjectReference[], which can write ResolvedProjectReference where never is read | 1 |
| Refused: a value of type never[] seen as SourceMappedPosition[], which can write SourceMappedPosition where never is read | 1 |
| Refused: a value of type never[] seen as StringLiteralLike[], which can write StringLiteralLike where never is read | 1 |
| Refused: a value of type never[] seen as TransformerFactory<Bundle \| SourceFile>[], which can write TransformerFactory<Bundle \| SourceFile> where never is read | 1 |
| Refused: a value of type never[] \| SortedArray<DiagnosticWithLocation> seen as Diagnostic[], which can write Diagnostic where never is read | 1 |
| Refused: a value of type never[][] seen as string[][], which can write string where never is read | 1 |
| Refused: a value of type number[] seen as string \| (string \| number)[] \| undefined, which can write string \| number where number is read | 1 |
| Refused: a value of type readonly string[] \| undefined seen as RegExp[] \| undefined, which can write RegExp where string is read | 1 |
| Refused: a value of type string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined seen as string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined, which can write string \| number where string is read | 1 |
| Refused: a value of type string \| number \| boolean \| string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> seen as CompilerOptionsValue, which can write string \| number where string is read | 1 |
| Refused: a value of type string[] seen as string \| (string \| number)[] \| undefined, which can write string \| number where string is read | 1 |
| Refused: a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"CheckMode", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof CheckMode is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"EmitFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof EmitFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"GeneratedIdentifierFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof GeneratedIdentifierFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"ModifierFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ModifierFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"NodeCheckFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof NodeCheckFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"NodeFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof NodeFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"ObjectFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ObjectFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"RelationComparisonResult", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof RelationComparisonResult is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"ScriptKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ScriptKind is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"SignatureCheckMode", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SignatureCheckMode is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"SignatureFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SignatureFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"SnippetKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SnippetKind is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"SymbolFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SymbolFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"SyntaxKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SyntaxKind is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"TransformFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TransformFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"TypeFacts", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TypeFacts is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"TypeFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TypeFlags is read | 1 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what undefined can't hold | 1 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint Declaration can be written, so it can write what undefined can't hold | 1 |
| Refused: a value of type { (fileName: string): DiagnosticWithLocation[]; (): Diagnostic[]; } seen as { (): Diagnostic[]; (fileName: string): DiagnosticWithLocation[]; }, which can write Diagnostic where DiagnosticWithLocation is read | 1 |
| Refused: a value of type { affectedFile: SourceFile; emitKind: BuilderFileEmit.Js \| BuilderFileEmit.JsMap \| BuilderFileEmit.JsInlineMap \| BuilderFileEmit.DtsErrors \| ... 6 more ... \| BuilderFileEmit.All; } seen as { affectedFile: Program \| SourceFile \| undefined; emitKind: BuilderFileEmit; }, which can write Program \| SourceFile \| undefined where SourceFile is read | 1 |
| Refused: a value of type { all?: boolean; allowImportingTsExtensions?: boolean; allowJs?: boolean; allowNonTsExtensions?: boolean; allowArbitraryExtensions?: boolean; allowSyntheticDefaultImports?: boolean; allowUmdGlobalAccess?: boolean; ... 129 more ...; moduleResolution: ModuleResolutionKind; } seen as CompilerOptions, which can write ModuleResolutionKind \| undefined where ModuleResolutionKind is read | 1 |
| Refused: a value of type { arguments: { name?: string; } & { path: string; }; range: CommentRange; } \| { arguments: { name: string; }; range: CommentRange; } \| { arguments: { factory: string; }; range: CommentRange; } \| ... 5 more ... \| ... seen as ({ arguments: { name?: string; } & { path: string; }; range: CommentRange; } \| { arguments: { name: string; }; range: CommentRange; } \| { arguments: { factory: string; }; range: CommentRange; } \| ... 5 more ... \| ...)[] \| ... 8 more ... \| ..., which can write { name?: string; } & { path: string; } where never is read | 1 |
| Refused: a value of type { compilerOptions: CompilerOptions; traceEnabled: boolean; affectingLocations: string[] \| undefined; resultFromCache?: ResolvedModuleWithFailedLookupLocations; ... 9 more ...; host: GetPackageJsonEntrypointsHost; } seen as ModuleResolutionState & { host: GetPackageJsonEntrypointsHost; }, which can write string[] \| undefined where never[] is read | 1 |
| Refused: a value of type { ending: ModuleSpecifierEnding; value: string; }[] seen as { ending: ModuleSpecifierEnding \| undefined; value: string; }[], which can write ModuleSpecifierEnding \| undefined where ModuleSpecifierEnding is read | 1 |
| Refused: a value of type { host: ModuleResolutionHost; traceEnabled: boolean; failedLookupLocations: string[] \| undefined; affectingLocations: string[] \| undefined; resultFromCache?: ResolvedModuleWithFailedLookupLocations; ... 8 more ...; reportDiagnostic: ...; } seen as ModuleResolutionState, which can write CompilerOptions where { all?: boolean; allowImportingTsExtensions?: boolean; allowJs?: boolean; allowNonTsExtensions?: boolean; allowArbitraryExtensions?: boolean; allowSyntheticDefaultImports?: boolean; allowUmdGlobalAccess?: boolean; ... 129 more ...; moduleResolution: ModuleResolutionKind; } is read | 1 |
| Refused: a value of type { id: number; flowNode: FlowNode; edges: never[]; text: string; lane: number; endLane: number; level: number; circular: false; } seen as FlowGraphNode, which can write FlowGraphEdge[] where never[] is read | 1 |
| Refused: a value of type { major: number; minor: number; patch: number; prerelease: string; build: string; } seen as { major: string \| number; minor: number; patch: number; prerelease: string \| readonly string[]; build: string \| readonly string[]; }, which can write string \| number where number is read | 1 |
| Refused: a value of type { name: string \| undefined; path: string; }[] seen as AmdDependency[], which can write AmdDependency where { name: string \| undefined; path: string; } is read | 1 |
| Refused: an arbitrary number or a value from another enum assigned to Extension.Mjs; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to ForegroundColorEscapeSequences.Grey; its members are a closed union | 1 |
| Refused: an unproven relation from Identifier to Identifier: optional field id has no proven compatible presence/type | 1 |
| Refused: an unproven relation from ImportEqualsDeclaration & { moduleReference: ExternalModuleReference; } to ImportEqualsDeclaration: optional field moduleReference.id has no proven compatible presence/type | 1 |
| Refused: an unproven relation from LiteralLikeNode to TemplateLiteralLikeNode: optional field rawText has no proven compatible presence/type | 1 |
| Refused: an unproven relation from NodeArray<ModifierLike> & readonly Decorator[] to NodeArray<Decorator>: optional field concat.parameter.element.slice.element.parent.name has no proven compatible presence/type | 1 |
| Refused: an unproven relation from PackageJsonPathFields to PackageJson: optional field version has no proven compatible presence/type | 1 |
| Refused: an unproven relation from SolutionBuilderHostBase<T> to SolutionBuilderHost<T>: optional field reportErrorSummary has no proven compatible presence/type | 1 |
| Refused: an unproven relation from string to T: the source is not assignable to the target | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot BuilderFileEmit.Js \| BuilderFileEmit.JsMap \| BuilderFileEmit.JsInlineMap \| BuilderFileEmit.DtsErrors \| BuilderFileEmit.DtsEmit \| ... 5 more ... \| BuilderFileEmit.All | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot BuilderFileEmit.None | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot CharacterCodes.plus | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot CharacterCodes.slash | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot Connection | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot EmitOnly.Dts | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot ModuleKind.ESNext | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot NodeFlags.NestedNamespace | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot PunctuationOrKeywordSyntaxKind | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.BindingElement | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ConditionalType | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ExtendsKeyword \| SyntaxKind.ImplementsKeyword | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Identifier | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ImportAttributes | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.KeyOfKeyword \| SyntaxKind.ReadonlyKeyword \| SyntaxKind.UniqueKeyword | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NamespaceExport | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NamespaceImport | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NumericLiteral \| SyntaxKind.BigIntLiteral \| SyntaxKind.StringLiteral \| SyntaxKind.JsxText \| SyntaxKind.JsxTextAllWhiteSpaces \| SyntaxKind.RegularExpressionLiteral \| SyntaxKind.NoSubstitutionTemplateLiteral | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.SourceFile | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.StringLiteral | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Unknown \| SyntaxKind.NumericLiteral \| SyntaxKind.BigIntLiteral \| SyntaxKind.StringLiteral \| SyntaxKind.JsxText \| SyntaxKind.RegularExpressionLiteral \| SyntaxKind.NoSubstitutionTemplateLiteral | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.VariableDeclaration | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot TempFlags | 1 |
| Refused: arguments | 1 |
| Refused: debugger | 1 |
| Refused: inherited library member hasOwnProperty read as an own field | 1 |
| Refused: inherited library member replace read as an own field | 1 |
| Refused: optional property Low in Partial<Levels> absent from structural source {}, which can hide fields | 1 |
| Refused: optional property _propertyAccessExpressionLikeQualifiedNameBrand in PropertyAccessEntityNameExpression absent from structural source never, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source OptionsBase, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "noImplicitAny" \| "strict">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "noImplicitThis" \| "strict">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" \| "strictBindCallApply">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" \| "strictBuiltinIteratorReturn">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" \| "strictFunctionTypes">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" \| "strictNullChecks">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" \| "strictPropertyInitialization">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" \| "useUnknownInCatchVariables">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source TypeAcquisition, which can hide fields | 1 |
| Refused: optional property createProgram in IncrementalProgramOptions<EmitAndSemanticDiagnosticsBuilderProgram> absent from structural source IncrementalCompilationOptions, which can hide fields | 1 |
| Refused: optional property id in BinaryExpression absent from structural source never, which can hide fields | 1 |
| Refused: optional property id in FalseLiteral absent from structural source never, which can hide fields | 1 |
| Refused: optional property jsDocCache in JSDocArray absent from structural source JSDoc[], which can hide fields | 1 |
| Refused: optional property modifiers in ExportAssignment absent from structural source never, which can hide fields | 1 |
| Refused: optional property modifiers in ParameterDeclaration absent from structural source never, which can hide fields | 1 |
| Refused: optional property modifiers in PropertyDeclaration absent from structural source never, which can hide fields | 1 |
| Refused: optional property modifiers in SetAccessorDeclaration absent from structural source never, which can hide fields | 1 |
| Refused: optional property packageRootPath in { moduleFileToTry: string; packageRootPath?: string; blockedByExports?: true; verbatimFromExports?: true; } absent from structural source { moduleFileToTry: string; }, which can hide fields | 1 |
| Refused: optional property preserve in FileReference absent from structural source { resolutionMode: ModuleKind.CommonJS \| ModuleKind.ESNext; }, which can hide fields | 1 |
| Refused: optional property skipTrivia in SourceMapSource absent from structural source SourceFile, which can hide fields | 1 |
| Refused: optional property types in { types?: { value: string; pos: number; end: number; }; } & { lib?: { value: string; pos: number; end: number; }; } & { path?: { value: string; pos: number; end: number; }; } & { "no-default-lib"?: string; } & { "resolution-mode"?: string; } & ... absent from structural source { [index: string]: string \| undefined; }, which can hide fields | 1 |
| Refused: optional property types in { types?: { value: string; pos: number; end: number; }; } & { lib?: { value: string; pos: number; end: number; }; } & { path?: { value: string; pos: number; end: number; }; } & { "no-default-lib"?: string; } & { "resolution-mode"?: string; } & ... absent from structural source { [index: string]: string \| { value: string; pos: number; end: number; }; }, which can hide fields | 1 |
| Refused: optional property watchFile in WatchOptions absent from structural source {}, which can hide fields | 1 |
| panic: Unhandled case in Node.Text: *ast.QualifiedName | 1 |

## latent / outside_compiler

| Kind and exact reason | Sites |
| --- | ---: |
| NotYet: reading ts | 5 |
| Refused: a method read as a value (setBlocking would lose its object, and this with it) | 1 |
| Refused: a method read as a value (tryEnableSourceMapsForHost would lose its object, and this with it) | 1 |

## latent / unattributed

| Kind and exact reason | Sites |
| --- | ---: |

## full / all

| Kind and exact reason | Sites |
| --- | ---: |
| Refused: a cast the runtime can't check | 3645 |
| NotYet: a function inside a function (a closure) | 3110 |
| Refused: an object refinement using an open numeric enum as a literal tag | 2776 |
| Refused: the non-null assertion ! | 1754 |
| NotYet: a method call through a structural signature in a program with statics; use typeof the declaring class | 1589 |
| NotYet: reading node | 1043 |
| Refused: a number as a condition | 647 |
| NotYet: a PrefixUnaryExpression on a value | 630 |
| NotYet: a NonNullExpression | 553 |
| NotYet: reading Debug | 509 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on node) | 414 |
| NotYet: a BinaryExpression with a value and a value | 410 |
| NotYet: a PrefixUnaryExpression on a number | 316 |
| NotYet: a function without a body | 303 |
| Refused: a value as a condition | 271 |
| NotYet: a BinaryExpression with a value and a boolean | 258 |
| NotYet: reading result | 227 |
| NotYet: a BinaryExpression with a number and a number | 200 |
| NotYet: a field of type boolean \| undefined | 164 |
| NotYet: a value of type __String | 152 |
| Refused: var | 140 |
| NotYet: a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions | 138 |
| NotYet: a function returning T | 118 |
| NotYet: a value of type Path | 113 |
| NotYet: assigning to an Identifier | 113 |
| NotYet: a BinaryExpression with a boolean and a value | 111 |
| NotYet: a call through ?. (an optional call) | 105 |
| Refused: \|\|= | 105 |
| Refused: a string as a condition | 101 |
| NotYet: for...of over an object | 98 |
| NotYet: reading type | 94 |
| NotYet: a function returning T \| undefined | 93 |
| NotYet: a value of type any | 91 |
| Refused: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 90 |
| Refused: a boolean \| undefined as a condition | 80 |
| NotYet: a BinaryExpression with a number and a boolean | 72 |
| NotYet: a parameter that isn't a plain name | 72 |
| Refused: a type predicate whose return is not proven (there is no body proving this parameter) | 71 |
| NotYet: a BinaryExpression with a boolean and a number | 62 |
| Refused: a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where SourceFile is read | 59 |
| NotYet: an array of never | 57 |
| NotYet: a declaration directly in a case (wrap the case in a block) | 56 |
| Refused: optional property id in Node absent from structural source never, which can hide fields | 55 |
| NotYet: a value of type T | 54 |
| NotYet: reading performance | 52 |
| NotYet: a PrefixUnaryExpression on a boolean \| undefined | 49 |
| Refused: the comma operator | 47 |
| NotYet: reading symbol | 46 |
| NotYet: reading name | 44 |
| NotYet: a BinaryExpression as a statement | 43 |
| NotYet: reading updated | 39 |
| NotYet: a BinaryExpression with a value and a number | 37 |
| Refused: a type predicate whose return is not proven (return paths through KindSwitchStatement are not verified) | 37 |
| NotYet: reading expression | 36 |
| NotYet: a PrefixUnaryExpression on a string | 35 |
| Refused: a value of type SourceFile seen as SourceFileLike, which can write readonly number[] \| undefined where readonly number[] is read | 34 |
| NotYet: reading host | 33 |
| NotYet: this outside a method | 33 |
| NotYet: a function value returning union of differently held members | 32 |
| NotYet: a call returning void \| undefined | 31 |
| NotYet: a void call used as a value | 31 |
| NotYet: reading statement | 31 |
| NotYet: an ElementAccessExpression | 27 |
| NotYet: reading declaration | 27 |
| NotYet: an array of T | 25 |
| NotYet: replacing a represented method at runtime | 25 |
| Refused: a value of type TupleTypeReference seen as TypeReference, which can write GenericType where TupleType is read | 25 |
| Refused: an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | 24 |
| NotYet: a field of type string \| NodeArray<JSDocComment> \| undefined | 23 |
| NotYet: reading clone | 23 |
| NotYet: reading compilerOptions | 23 |
| Refused: a method in object destructuring | 23 |
| Refused: optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<TypeParameterDeclaration>, which can hide fields | 23 |
| NotYet: a value of type __String \| undefined | 22 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on kind) | 22 |
| NotYet: a BinaryExpression with a number \| undefined and a number | 21 |
| NotYet: a BinaryExpression with a string and a string | 21 |
| NotYet: a value of type T \| undefined | 21 |
| NotYet: a value of type unknown | 21 |
| NotYet: new an Identifier | 21 |
| Refused: optional property source in SourceMapRange absent from structural source TextRange, which can hide fields | 21 |
| panic: statement panic: runtime error: invalid memory address or nil pointer dereference | 21 |
| NotYet: a PrefixUnaryExpression on a number \| undefined | 20 |
| NotYet: a value of type SolutionBuilderState<T> | 20 |
| Refused: a type predicate whose return is not proven (branch is not a trusted parameter check) | 20 |
| Refused: optional property return in ArrayIterator<Statement> absent from structural source ArrayIterator<JsonObjectExpressionStatement>, which can hide fields | 20 |
| NotYet: a call to a PropertyAccessExpression | 19 |
| NotYet: reading resolved | 19 |
| Refused: a value of type FlowNode seen as FlowNode \| undefined, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 19 |
| NotYet: a BinaryExpression with a boolean \| undefined and a boolean | 18 |
| NotYet: a function returning U \| undefined | 18 |
| NotYet: a function returning __String | 18 |
| NotYet: a generic function as a value | 18 |
| NotYet: a value of type ResolvedConfigFilePath | 18 |
| NotYet: reading sourceFile | 18 |
| Refused: a method read as a value (liftToBlock would lose its object, and this with it) | 18 |
| NotYet: a function returning __String \| undefined | 17 |
| NotYet: reading options | 17 |
| NotYet: reading visited | 17 |
| Refused: optional property id in ArrayBindingPattern absent from structural source never, which can hide fields | 17 |
| NotYet: a BinaryExpression with a number and a value | 16 |
| NotYet: reading candidate | 16 |
| Refused: a value of type ResolvedType seen as ObjectType, which can write SymbolTable \| undefined where SymbolTable is read | 16 |
| NotYet: a function returning Path | 15 |
| NotYet: a value of type object | 15 |
| NotYet: reading file | 15 |
| NotYet: reading sig | 15 |
| Refused: a number \| undefined as a condition | 15 |
| Refused: a value of type never[] seen as Diagnostic[], which can write Diagnostic where never is read | 15 |
| Refused: yield (generators) | 15 |
| NotYet: a Set of __String (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far) | 14 |
| NotYet: a function returning any | 14 |
| NotYet: an array of any | 14 |
| NotYet: reading block | 14 |
| NotYet: reading child | 14 |
| NotYet: reading parameter | 14 |
| Refused: a value of type Node seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 14 |
| Refused: a value of type never[] seen as BaseType[], which can write BaseType where never is read | 14 |
| Refused: a value of type { value: never; done: true; } seen as IteratorResult<Mapping, any>, which can write any where never is read | 14 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.JSDocTypeExpression | 14 |
| NotYet: a BinaryExpression with a boolean and a boolean \| undefined | 13 |
| NotYet: reading diag | 13 |
| NotYet: reading index | 13 |
| NotYet: reading related | 13 |
| Refused: a value of type never[] seen as string[], which can write string where never is read | 13 |
| Refused: the void operator | 13 |
| NotYet: a function with an optional or rest parameter, as a value | 12 |
| NotYet: a value of type ResolvedConfigFileName | 12 |
| NotYet: optional chaining to .size on a value | 12 |
| NotYet: reading t | 12 |
| Refused: a definite assignment assertion ! | 12 |
| Refused: a generator function | 12 |
| Refused: a method read as a value (parenthesizeExpressionForDisallowedComma would lose its object, and this with it) | 12 |
| Refused: a union of differently held members as a condition | 12 |
| Refused: a value of type Node[] seen as unknown[], which can write unknown where Node is read | 12 |
| Refused: a value of type never[] seen as Signature[], which can write Signature where never is read | 12 |
| Refused: a value of type never[] seen as Symbol[], which can write Symbol where never is read | 12 |
| Refused: a value without nominal ancestry seen as BinaryExpressionStateMachine<TOuterState, TState, TResult> | 12 |
| NotYet: a function returning undefined | 11 |
| NotYet: a value of type NonNullable<T> | 11 |
| NotYet: a value of type TKind | 11 |
| NotYet: reading cache | 11 |
| Refused: a namespace | 11 |
| Refused: a value of type NodeBuilderContext seen as SyntacticTypeNodeBuilderContext, which can write Required<Pick<SymbolTracker, "reportInferenceFallback">> where SymbolTrackerImpl is read | 11 |
| Refused: a value of type T seen as Mutable<T>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of any would replace | 11 |
| Refused: an index signature | 11 |
| Refused: in | 11 |
| NotYet: a field of type true \| Node \| undefined | 10 |
| NotYet: a function returning T[] | 10 |
| NotYet: an enum inside a function or block; declare it at module scope | 10 |
| NotYet: an optional chain longer than one step | 10 |
| NotYet: for...in without a proven fixed plain-object origin (arrays, prototypes and absent synthetic fields cannot be enumerated soundly) | 10 |
| NotYet: reading body | 10 |
| NotYet: reading decl | 10 |
| NotYet: reading exception | 10 |
| NotYet: reading flags | 10 |
| NotYet: reading reference | 10 |
| Refused: a value of type BuilderProgramStateWithDefinedProgram seen as BuilderProgramState, which can write Program \| undefined where Program is read | 10 |
| Refused: a value of type Map<string, never[]> seen as Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>, which can write string[] where never[] is read | 10 |
| Refused: a value of type Symbol \| undefined seen as Type \| undefined, which can write TypeFlags where SymbolFlags is read | 10 |
| Refused: optional property constraint in TypeParameter absent from structural source Type, which can hide fields | 10 |
| Refused: optional property id in BigIntLiteral absent from structural source never, which can hide fields | 10 |
| NotYet: a BinaryExpression with a number and a string | 9 |
| NotYet: a ModuleDeclaration | 9 |
| NotYet: a PrefixUnaryExpression on a union of differently held members | 9 |
| NotYet: a value of type T["kind"] | 9 |
| NotYet: reading Parser | 9 |
| NotYet: reading elem | 9 |
| NotYet: reading existing | 9 |
| NotYet: reading expr | 9 |
| NotYet: reading parsed | 9 |
| NotYet: reading restType | 9 |
| NotYet: storing boolean \| undefined in a field | 9 |
| Refused: a function taking Identifier seen as one taking Identifier \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 9 |
| Refused: a value of type BuilderProgramState seen as ReusableBuilderProgramState, which can write Map<Path, readonly Diagnostic[] \| readonly ReusableDiagnostic[]> where Map<Path, readonly Diagnostic[]> is read | 9 |
| Refused: a value of type DiagnosticWithLocation[] seen as Diagnostic[], which can write Diagnostic where DiagnosticWithLocation is read | 9 |
| Refused: a value of type Expression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 9 |
| Refused: a value of type never[] seen as IndexInfo[], which can write IndexInfo where never is read | 9 |
| NotYet: a BinaryExpression with a string and a number | 8 |
| NotYet: a field of type true \| undefined | 8 |
| NotYet: a value of type K | 8 |
| NotYet: a value of type readonly T[] | 8 |
| NotYet: a value of type readonly T[] \| undefined | 8 |
| NotYet: new a ParenthesizedExpression | 8 |
| NotYet: reading bundle | 8 |
| NotYet: reading diagnostics | 8 |
| NotYet: reading left | 8 |
| NotYet: reading links | 8 |
| NotYet: reading parent | 8 |
| NotYet: reading reduced | 8 |
| NotYet: reading specifier | 8 |
| NotYet: reading statements | 8 |
| NotYet: reading target | 8 |
| NotYet: reading temp | 8 |
| NotYet: reading text | 8 |
| Refused: a function taking boolean seen as one taking boolean \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 8 |
| Refused: a method read as a value (realpath would lose its object, and this with it) | 8 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on type) | 8 |
| Refused: a value of type never[] seen as Type[], which can write Type where never is read | 8 |
| NotYet: a BinaryExpression with a string and a boolean | 7 |
| NotYet: a Set of Path (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far) | 7 |
| NotYet: a field of type string \| DiagnosticMessageChain | 7 |
| NotYet: a value of type BindableStaticNameExpression | 7 |
| NotYet: assigning a field of a value | 7 |
| NotYet: reading array | 7 |
| NotYet: reading buildOrder | 7 |
| NotYet: reading context | 7 |
| NotYet: reading current | 7 |
| NotYet: reading element | 7 |
| NotYet: reading firstDecorator | 7 |
| NotYet: reading flowType | 7 |
| NotYet: reading initializer | 7 |
| NotYet: reading res | 7 |
| NotYet: reading scope | 7 |
| NotYet: reading source | 7 |
| NotYet: reading token | 7 |
| NotYet: reading ts | 7 |
| NotYet: reading typeNode | 7 |
| NotYet: regex replacement other than a string | 7 |
| Refused: a method read as a value (parenthesizeLeftSideOfAccess would lose its object, and this with it) | 7 |
| Refused: a value of type FlowNode seen as FlowNode[] \| FlowNode \| undefined, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 7 |
| Refused: a value of type GeneratedIdentifier \| Identifier seen as Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 7 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.SourceFile | 7 |
| Refused: optional property id in ExternalModuleReference absent from structural source never, which can hide fields | 7 |
| Refused: optional property id in Identifier absent from structural source never, which can hide fields | 7 |
| Refused: optional property members in ObjectType absent from structural source IntrinsicType, which can hide fields | 7 |
| NotYet: ?.[] on a value | 6 |
| NotYet: a BinaryExpression with a boolean and a string | 6 |
| NotYet: a BinaryExpression with a value and a string | 6 |
| NotYet: a function value returning boolean \| undefined | 6 |
| NotYet: a value of type HasJSDoc \| undefined | 6 |
| NotYet: a value of type TOuterState | 6 |
| NotYet: assigning to a NonNullExpression | 6 |
| NotYet: reading autoGenerate | 6 |
| NotYet: reading c | 6 |
| NotYet: reading cacheKey | 6 |
| NotYet: reading directoryWatcher | 6 |
| NotYet: reading errorNode | 6 |
| NotYet: reading includes | 6 |
| NotYet: reading isGenerator | 6 |
| NotYet: reading iterationTypes | 6 |
| NotYet: reading localOrExportSymbol | 6 |
| NotYet: reading map | 6 |
| NotYet: reading packageInfo | 6 |
| NotYet: reading params | 6 |
| NotYet: reading pos | 6 |
| NotYet: reading propertyName | 6 |
| NotYet: reading queue | 6 |
| NotYet: reading start | 6 |
| NotYet: reading typeParameters | 6 |
| NotYet: reading v | 6 |
| NotYet: reading value | 6 |
| NotYet: storing boolean in a field | 6 |
| Refused: a function taking BinaryExpressionStateMachine<TOuterState, TState, TResult> seen as one taking BinaryExpressionStateMachine<TOuterState, TState, TResult> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 6 |
| Refused: a method read as a value (fileExists would lose its object, and this with it) | 6 |
| Refused: a method read as a value (parenthesizeOperandOfPrefixUnary would lose its object, and this with it) | 6 |
| Refused: a method read as a value (readFile would lose its object, and this with it) | 6 |
| Refused: a parameter property | 6 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on n) | 6 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on token) | 6 |
| Refused: a value of type Declaration[] seen as Node[], which can write Node where Declaration is read | 6 |
| Refused: a value of type Expression seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 6 |
| Refused: a value of type FlowNode \| undefined seen as false \| FlowNode \| undefined, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 6 |
| Refused: a value of type Identifier seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 6 |
| Refused: a value of type NonNullable<T> seen as T, a type parameter whose constraint any can be written, so it can write what NonNullable<T> can't hold | 6 |
| Refused: a value of type TypeNode \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 6 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint Node can be written, so it can write what undefined can't hold | 6 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.TypeKeyword | 6 |
| Refused: optional property id in Expression absent from structural source never, which can hide fields | 6 |
| NotYet: a BinaryExpression with a boolean \| undefined and a number | 5 |
| NotYet: a BinaryExpression with a number and a boolean \| undefined | 5 |
| NotYet: a case whose type differs from the switch's | 5 |
| NotYet: a function returning Token<TKind> | 5 |
| NotYet: a function returning U | 5 |
| NotYet: a value of type string \| null \| undefined | 5 |
| NotYet: an array of U | 5 |
| NotYet: reading arg | 5 |
| NotYet: reading assignedName | 5 |
| NotYet: reading buildOptions | 5 |
| NotYet: reading classType | 5 |
| NotYet: reading configFile | 5 |
| NotYet: reading entityName | 5 |
| NotYet: reading exportSpecifiers | 5 |
| NotYet: reading extensions | 5 |
| NotYet: reading innerExpression | 5 |
| NotYet: reading jsFilePath | 5 |
| NotYet: reading merged | 5 |
| NotYet: reading objectFlags | 5 |
| NotYet: reading path | 5 |
| NotYet: reading propType | 5 |
| NotYet: reading thisType | 5 |
| NotYet: reading typeName | 5 |
| NotYet: reading types | 5 |
| NotYet: reading valueDeclaration | 5 |
| NotYet: reading watcher | 5 |
| NotYet: reading yieldType | 5 |
| Refused: a method read as a value (directoryExists would lose its object, and this with it) | 5 |
| Refused: a method read as a value (getCanonicalFileName would lose its object, and this with it) | 5 |
| Refused: a method read as a value (trace would lose its object, and this with it) | 5 |
| Refused: a value of type CompilerOptionsValue seen as TsConfigSourceFile \| CompilerOptionsValue, which can write string \| number where string is read | 5 |
| Refused: a value of type DiagnosticWithDetachedLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where undefined is read | 5 |
| Refused: a value of type FlowArrayMutation \| FlowAssignment seen as FlowNode, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 5 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier \| Node seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 5 |
| Refused: a value of type Node seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 5 |
| Refused: a value of type Symbol \| undefined seen as Declaration \| undefined, which can write number \| undefined where number is read | 5 |
| Refused: a value of type string[] seen as CompilerOptionsValue, which can write string \| number where string is read | 5 |
| Refused: an arbitrary number or a value from another enum assigned to InternalSymbolName.Call; its members are a closed union | 5 |
| Refused: an unproven value assigned to a numeric literal or enum member slot CompilerOptions \| BuilderFileEmit | 5 |
| Refused: an unproven value assigned to a numeric literal or enum member slot ModifierFlags.None | 5 |
| Refused: optional property all in CompilerOptions absent from structural source BuildOptions, which can hide fields | 5 |
| Refused: optional property reportsUnnecessary in Diagnostic absent from structural source DiagnosticRelatedInformation, which can hide fields | 5 |
| Refused: optional property return in ArrayIterator<Declaration> absent from structural source ArrayIterator<ClassElement>, which can hide fields | 5 |
| Refused: optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<HeritageClause>, which can hide fields | 5 |
| Refused: optional property skipLogging in ErrorOutputContainer absent from structural source { errors?: Diagnostic[] \| undefined; }, which can hide fields | 5 |
| Refused: optional property version in PackageJson absent from structural source PackageJsonPathFields, which can hide fields | 5 |
| NotYet: a BinaryExpression with a string and a value | 4 |
| NotYet: a VoidExpression as a statement | 4 |
| NotYet: a computed field name | 4 |
| NotYet: a field of type "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number> | 4 |
| NotYet: a function returning CompilerOptionsValue | 4 |
| NotYet: a function returning NodeArray<T> | 4 |
| NotYet: a function returning PackageJson[K] \| undefined | 4 |
| NotYet: a function returning R | 4 |
| NotYet: a function returning SortedReadonlyArray<T> | 4 |
| NotYet: a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables | 4 |
| NotYet: a value of type (ConstructorDeclaration & { body: Block; }) \| undefined | 4 |
| NotYet: a value of type (EmitNode & { autoGenerate: AutoGenerateInfo; }) \| (EmitNode & { autoGenerate: AutoGenerateInfo; }) | 4 |
| NotYet: a value of type CompilerOptionsValue | 4 |
| NotYet: a value of type ExpressionWithTypeArguments & { readonly expression: Identifier \| PropertyAccessEntityNameExpression; } | 4 |
| NotYet: a value of type HasJSDoc | 4 |
| NotYet: a value of type T \| Program | 4 |
| NotYet: a value of type string \| (void & { __escapedIdentifier: void; }) \| (string & { __escapedIdentifier: void; }) | 4 |
| NotYet: an array of ResolvedConfigFileName | 4 |
| NotYet: an array of boolean \| undefined | 4 |
| NotYet: assigning to an ObjectLiteralExpression | 4 |
| NotYet: reading BuilderState | 4 |
| NotYet: reading args | 4 |
| NotYet: reading argument | 4 |
| NotYet: reading asteriskToken | 4 |
| NotYet: reading baseTypeNode | 4 |
| NotYet: reading bindings | 4 |
| NotYet: reading cached | 4 |
| NotYet: reading cachedChain | 4 |
| NotYet: reading canReportSummary | 4 |
| NotYet: reading chain | 4 |
| NotYet: reading childrenTargetType | 4 |
| NotYet: reading commonSourceDirectory | 4 |
| NotYet: reading config | 4 |
| NotYet: reading constraint | 4 |
| NotYet: reading container | 4 |
| NotYet: reading cooked | 4 |
| NotYet: reading descriptorName | 4 |
| NotYet: reading directoryExists | 4 |
| NotYet: reading e | 4 |
| NotYet: reading emitNode | 4 |
| NotYet: reading emitResult | 4 |
| NotYet: reading excludeRegex | 4 |
| NotYet: reading exportStatement | 4 |
| NotYet: reading extensionGroup | 4 |
| NotYet: reading func | 4 |
| NotYet: reading getCurrentDirectory | 4 |
| NotYet: reading hasDefaultClause | 4 |
| NotYet: reading id | 4 |
| NotYet: reading identifier | 4 |
| NotYet: reading importDeclaration | 4 |
| NotYet: reading inference | 4 |
| NotYet: reading instantiation | 4 |
| NotYet: reading jsxFactorySymbol | 4 |
| NotYet: reading lexicallyScopedSymbol | 4 |
| NotYet: reading mappedType | 4 |
| NotYet: reading members | 4 |
| NotYet: reading normalized | 4 |
| NotYet: reading openParenPosition | 4 |
| NotYet: reading original | 4 |
| NotYet: reading output | 4 |
| NotYet: reading parameters | 4 |
| NotYet: reading parseNode | 4 |
| NotYet: reading prop | 4 |
| NotYet: reading propertyOriginalNode | 4 |
| NotYet: reading props | 4 |
| NotYet: reading rawText | 4 |
| NotYet: reading refPath | 4 |
| NotYet: reading referencedName | 4 |
| NotYet: reading regularType | 4 |
| NotYet: reading resolvedModule | 4 |
| NotYet: reading restParamSymbol | 4 |
| NotYet: reading restParameter | 4 |
| NotYet: reading reversed | 4 |
| NotYet: reading seen | 4 |
| NotYet: reading sourceIndex | 4 |
| NotYet: reading sourceSymbol | 4 |
| NotYet: reading state | 4 |
| NotYet: reading targetSymbol | 4 |
| NotYet: reading typeSet | 4 |
| NotYet: reading watchCompilerHost | 4 |
| Refused: Object.defineProperties | 4 |
| Refused: a function taking readonly Modifier[] \| undefined seen as one taking readonly ModifierLike[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 4 |
| Refused: a method read as a value (getSourceFile would lose its object, and this with it) | 4 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on expr) | 4 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on f) | 4 |
| Refused: a value of type Declaration \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 4 |
| Refused: a value of type Expression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 4 |
| Refused: a value of type Expression \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 4 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 4 |
| Refused: a value of type Identifier[][] seen as ModuleExportName[][], which can write ModuleExportName[] where Identifier[] is read | 4 |
| Refused: a value of type Mutable<GeneratedIdentifier> seen as Identifier, which can write EmitNode \| undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 4 |
| Refused: a value of type NodeArray<Statement> seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 4 |
| Refused: a value of type SyntheticSuper seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 4 |
| Refused: a value of type never[] seen as Node[], which can write Node where never is read | 4 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Identifier | 4 |
| Refused: optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<ClassElement>, which can hide fields | 4 |
| Refused: optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<T>, which can hide fields | 4 |
| Refused: optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<TypeElement>, which can hide fields | 4 |
| Refused: optional property return in ArrayIterator<any> absent from structural source ArrayIterator<Child>, which can hide fields | 4 |
| NotYet: Object.entries on a shape not proven by a plain literal or its const binding | 3 |
| NotYet: RegExp with a nonconstant pattern | 3 |
| NotYet: a BinaryExpression with a number \| undefined and a boolean | 3 |
| NotYet: a BinaryExpression with a union of differently held members and a value | 3 |
| NotYet: a ConditionalExpression as a statement | 3 |
| NotYet: a Map of false \| MutableFileSystemEntries | 3 |
| NotYet: a call to an Identifier | 3 |
| NotYet: a function returning ImmediatelyInvokedArrowFunction | 3 |
| NotYet: a function returning ResolvedConfigFileName | 3 |
| NotYet: a function returning T \| T[] \| undefined | 3 |
| NotYet: a function returning T \| readonly T[] \| undefined | 3 |
| NotYet: a function returning object | 3 |
| NotYet: a narrowed boolean \| undefined field; copy the field into a local and narrow that local instead | 3 |
| NotYet: a value of type BindableAccessExpression | 3 |
| NotYet: a value of type BindableObjectDefinePropertyCall | 3 |
| NotYet: a value of type Children \| undefined | 3 |
| NotYet: a value of type EndOfFileToken | 3 |
| NotYet: a value of type Identifier \| __String | 3 |
| NotYet: a value of type ImmediatelyInvokedArrowFunction | 3 |
| NotYet: a value of type InitializedVariableDeclaration | 3 |
| NotYet: a value of type T[] | 3 |
| NotYet: a value of type U | 3 |
| NotYet: a value of type WrappedExpression<AnonymousFunctionDefinition> | 3 |
| NotYet: a value of type X | 3 |
| NotYet: a value of type false \| TypeOnlyAliasDeclaration \| undefined | 3 |
| NotYet: a value of type object \| undefined | 3 |
| NotYet: a value of type undefined | 3 |
| NotYet: an array of NonNullable<T> | 3 |
| NotYet: an array of undefined | 3 |
| NotYet: assigning an element of a value | 3 |
| NotYet: destructuring anything but a tuple into [names] | 3 |
| NotYet: reading addUndefined | 3 |
| NotYet: reading ancestorFacts | 3 |
| NotYet: reading arrayLiteral | 3 |
| NotYet: reading awaitedType | 3 |
| NotYet: reading baseTypes | 3 |
| NotYet: reading buildInfo | 3 |
| NotYet: reading builderProgram | 3 |
| NotYet: reading callee | 3 |
| NotYet: reading canUseSourceFile | 3 |
| NotYet: reading checkType | 3 |
| NotYet: reading childrenNameType | 3 |
| NotYet: reading clause | 3 |
| NotYet: reading constructor | 3 |
| NotYet: reading currentNode | 3 |
| NotYet: reading decorationStatements | 3 |
| NotYet: reading decorator | 3 |
| NotYet: reading diagnosticStart | 3 |
| NotYet: reading diff | 3 |
| NotYet: reading elementType | 3 |
| NotYet: reading emitKind | 3 |
| NotYet: reading emitSuperHelpers | 3 |
| NotYet: reading escapedValue | 3 |
| NotYet: reading evaluated | 3 |
| NotYet: reading exportEquals | 3 |
| NotYet: reading exports | 3 |
| NotYet: reading firstAccessorWithDecorators | 3 |
| NotYet: reading firstDecl | 3 |
| NotYet: reading firstDeclaration | 3 |
| NotYet: reading functionName | 3 |
| NotYet: reading getter | 3 |
| NotYet: reading graphNode | 3 |
| NotYet: reading hostNode | 3 |
| NotYet: reading indexInfos | 3 |
| NotYet: reading indexType | 3 |
| NotYet: reading inferredProp | 3 |
| NotYet: reading info | 3 |
| NotYet: reading instantiated | 3 |
| NotYet: reading introducesError | 3 |
| NotYet: reading isAmbient | 3 |
| NotYet: reading isAsync | 3 |
| NotYet: reading isCallExpression | 3 |
| NotYet: reading isSimilarNode | 3 |
| NotYet: reading item | 3 |
| NotYet: reading jsdocAliasDecl | 3 |
| NotYet: reading jsdocParameters | 3 |
| NotYet: reading key | 3 |
| NotYet: reading kind | 3 |
| NotYet: reading languageVersion | 3 |
| NotYet: reading lastDecorator | 3 |
| NotYet: reading lastError | 3 |
| NotYet: reading lastParam | 3 |
| NotYet: reading line | 3 |
| NotYet: reading major | 3 |
| NotYet: reading mapping | 3 |
| NotYet: reading method | 3 |
| NotYet: reading methodType | 3 |
| NotYet: reading modifier | 3 |
| NotYet: reading moduleSymbol | 3 |
| NotYet: reading moveModifiers | 3 |
| NotYet: reading nameType | 3 |
| NotYet: reading named | 3 |
| NotYet: reading newSymbol | 3 |
| NotYet: reading objectType | 3 |
| NotYet: reading oldEmitKind | 3 |
| NotYet: reading oldOptions | 3 |
| NotYet: reading oldStartN | 3 |
| NotYet: reading opcode | 3 |
| NotYet: reading parentType | 3 |
| NotYet: reading parts | 3 |
| NotYet: reading promoteToIIFE | 3 |
| NotYet: reading property | 3 |
| NotYet: reading questionToken | 3 |
| NotYet: reading range | 3 |
| NotYet: reading regularNew | 3 |
| NotYet: reading resolution | 3 |
| NotYet: reading resolvedRequire | 3 |
| NotYet: reading resolvedSymbol | 3 |
| NotYet: reading restElement | 3 |
| NotYet: reading s | 3 |
| NotYet: reading sourceEmitHelpers | 3 |
| NotYet: reading sourceFileOrBundle | 3 |
| NotYet: reading suggestion | 3 |
| NotYet: reading symbolFromSymbolTable | 3 |
| NotYet: reading tag | 3 |
| NotYet: reading targetFlags | 3 |
| NotYet: reading targetHasRestElement | 3 |
| NotYet: reading targetPropertySymbol | 3 |
| NotYet: reading typeArguments | 3 |
| NotYet: reading typeCopy | 3 |
| NotYet: reading typeSymbol | 3 |
| NotYet: reading varStatement | 3 |
| Refused: a function taking () => T seen as one taking () => T (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 |
| Refused: a function taking [fileName: string, text: string, writeByteOrderMark: boolean, onError?: ((message: string) => void) \| undefined, sourceFiles?: readonly SourceFile[] \| undefined, data?: WriteFileCallbackData \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 |
| Refused: a function taking number seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 |
| Refused: a function taking string seen as one taking [fileName: string, text: string, writeByteOrderMark: boolean, onError?: ((message: string) => void) \| undefined, sourceFiles?: readonly SourceFile[] \| undefined, data?: WriteFileCallbackData \| undefined] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 |
| Refused: a method read as a value (afterProgramCreate would lose its object, and this with it) | 3 |
| Refused: a method read as a value (afterProgramEmitAndDiagnostics would lose its object, and this with it) | 3 |
| Refused: a method read as a value (getDirectories would lose its object, and this with it) | 3 |
| Refused: a method read as a value (getParsedCommandLine would lose its object, and this with it) | 3 |
| Refused: a method read as a value (getSemanticDiagnosticsOfNextAffectedFile would lose its object, and this with it) | 3 |
| Refused: a method read as a value (now would lose its object, and this with it) | 3 |
| Refused: a method read as a value (onWatchStatusChange would lose its object, and this with it) | 3 |
| Refused: a method read as a value (reportLikelyUnsafeImportRequiredError would lose its object, and this with it) | 3 |
| Refused: a method read as a value (reportPrivateInBaseOfClassExpression would lose its object, and this with it) | 3 |
| Refused: a method read as a value (trackSymbol would lose its object, and this with it) | 3 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on block) | 3 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on d) | 3 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on element) | 3 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on entry) | 3 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on info) | 3 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on member) | 3 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on tag) | 3 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on value) | 3 |
| Refused: a value of type BindingOrAssignmentElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type ClassDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type Diagnostic seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 3 |
| Refused: a value of type FutureSourceFile \| SourceFile seen as Pick<SourceFile, "fileName" \| "impliedNodeFormat">, whose readonly field fileName becomes writable: a readonly field may hold something narrower than string, which a write of string would replace | 3 |
| Refused: a value of type GeneratedIdentifier seen as Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 3 |
| Refused: a value of type GenericType \| ResolvedType seen as ObjectType, which can write SymbolTable \| undefined where SymbolTable is read | 3 |
| Refused: a value of type Identifier seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type Identifier \| TextRange seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type Identifier \| undefined seen as Identifier \| TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type Map<string, string[] \| never[]> seen as Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>, which can write string where never is read | 3 |
| Refused: a value of type NodeArray<Statement> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type ParameterDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type PropertyName seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type Signature \| undefined seen as Type \| undefined, which can write TypeFlags where SignatureFlags is read | 3 |
| Refused: a value of type SourceFile seen as SourceFile \| SourceFileLike, which can write readonly number[] \| undefined where readonly number[] is read | 3 |
| Refused: a value of type TsConfigSourceFile \| CompilerOptionsValue seen as string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined, which can write string \| number where string is read | 3 |
| Refused: a value of type Type \| undefined seen as Symbol \| undefined, which can write SymbolFlags where TypeFlags is read | 3 |
| Refused: a value of type Type \| undefined seen as TypeNode \| undefined, which can write number \| undefined where number is read | 3 |
| Refused: a value of type YieldExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type never[] seen as Declaration[], which can write Declaration where never is read | 3 |
| Refused: a value of type never[] seen as DiagnosticArguments, which can write string \| number \| boolean \| readonly string[] \| SourceFile \| undefined where never is read | 3 |
| Refused: a value of type never[] seen as ResolvedConfigFileName[], which can write ResolvedConfigFileName where never is read | 3 |
| Refused: a value of type never[] seen as TypeParameter[], which can write TypeParameter where never is read | 3 |
| Refused: a value of type never[] seen as string[] \| never[], which can write string where never is read | 3 |
| Refused: a value of type readonly Extension[][] seen as readonly string[][], which can write string where Extension is read | 3 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint WatchedFileWithIsClosed can be written, so it can write what undefined can't hold | 3 |
| Refused: an arbitrary number or a value from another enum assigned to Extension.Ts; its members are a closed union | 3 |
| Refused: an unproven relation from StructuredType to ObjectType: optional field members has no proven compatible presence/type | 3 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ExpressionWithTypeArguments | 3 |
| Refused: optional property id in FalseLiteral absent from structural source never, which can hide fields | 3 |
| Refused: optional property id in LiteralExpression & StringLiteral absent from structural source never, which can hide fields | 3 |
| Refused: optional property omitTrailingSemicolon in PrinterOptions absent from structural source CompilerOptions, which can hide fields | 3 |
| Refused: optional property rawText in TemplateLiteralLikeNode absent from structural source NumericLiteral, which can hide fields | 3 |
| Refused: optional property resolutionMode in FileReference absent from structural source { preserve: true; }, which can hide fields | 3 |
| Refused: optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<EnumMember>, which can hide fields | 3 |
| Refused: optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<TemplateSpan>, which can hide fields | 3 |
| Refused: optional property return in ArrayIterator<Statement> absent from structural source ArrayIterator<ExpressionStatement>, which can hide fields | 3 |
| SkippedDependency: a dependency function whose body has checker diagnostics (measurement skipped) | 3 |
| NotYet: a BinaryExpression with a number \| undefined and a number \| undefined | 2 |
| NotYet: a BinaryExpression with a union of differently held members and a boolean | 2 |
| NotYet: a BinaryExpression with a union of differently held members and a string | 2 |
| NotYet: a BinaryExpression with a value and a number \| undefined | 2 |
| NotYet: a Map of HostFileInfo | 2 |
| NotYet: a Map of VisitResult<ExportAssignment \| LateVisibilityPaintedStatement \| undefined> | 2 |
| NotYet: a YieldExpression as a statement | 2 |
| NotYet: a boolean \| undefined variable a function value captures | 2 |
| NotYet: a destructured name that isn't plain | 2 |
| NotYet: a field of type AnyBuildOrder \| undefined | 2 |
| NotYet: a field of type boolean \| (() => boolean) \| undefined | 2 |
| NotYet: a field of type string \| number \| boolean \| DiagnosticMessage \| undefined | 2 |
| NotYet: a field of type string \| number \| undefined | 2 |
| NotYet: a function returning Extract<ClassDeclaration, Pick<...>> \| Extract<...> | 2 |
| NotYet: a function returning ImmediatelyInvokedFunctionExpression | 2 |
| NotYet: a function returning ModeAwareCacheKey | 2 |
| NotYet: a function returning Path \| undefined | 2 |
| NotYet: a function returning T \| EmptyStatement \| undefined | 2 |
| NotYet: a function returning U[] | 2 |
| NotYet: a function returning U[] \| undefined | 2 |
| NotYet: a function returning V | 2 |
| NotYet: a function returning WatchFactory<X, Y>[T] | 2 |
| NotYet: a function returning readonly T[] | 2 |
| NotYet: a function returning readonly T[] \| undefined | 2 |
| NotYet: a function returning void \| "skip" | 2 |
| NotYet: a tagged template other than the intrinsic String.raw | 2 |
| NotYet: a tuple element of type string \| number \| boolean \| readonly string[] \| SourceFile \| undefined | 2 |
| NotYet: a union of differently held members variable a function value captures | 2 |
| NotYet: a value of type (AssignmentExpression<EqualsToken> & { readonly left: GeneratedIdentifier; }) \| undefined | 2 |
| NotYet: a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) \| undefined | 2 |
| NotYet: a value of type AnonymousFunctionDefinition | 2 |
| NotYet: a value of type Canonicalized | 2 |
| NotYet: a value of type CompilerHost & ReadBuildProgramHost | 2 |
| NotYet: a value of type ExpressionWithTypeArguments & { expression: Identifier \| PropertyAccessEntityNameExpression; } | 2 |
| NotYet: a value of type IncrementalBuildInfoFileId | 2 |
| NotYet: a value of type K \| undefined | 2 |
| NotYet: a value of type Map<K, T> | 2 |
| NotYet: a value of type ParameterPropertyDeclaration | 2 |
| NotYet: a value of type Path \| undefined | 2 |
| NotYet: a value of type SourceFile | 2 |
| NotYet: a value of type T \| T[] | 2 |
| NotYet: a value of type T \| readonly T[] | 2 |
| NotYet: a value of type TInArray | 2 |
| NotYet: a value of type ThisCapturingVariableDeclaration | 2 |
| NotYet: a value of type TypeNode & LiteralTypeNode & { readonly literal: StringLiteral; } | 2 |
| NotYet: a value of type U \| readonly U[] \| undefined | 2 |
| NotYet: a value of type U \| undefined | 2 |
| NotYet: a value of type V \| undefined | 2 |
| NotYet: a value of type false \| RegExpExecArray \| null | 2 |
| NotYet: a value of type never | 2 |
| NotYet: an array of Child | 2 |
| NotYet: an array of V | 2 |
| NotYet: destructuring a value | 2 |
| NotYet: indexOf on a tuple (a tuple is held as an object, not an array, so far; write it as an array where it's made) | 2 |
| NotYet: lastIndexOf with these arguments | 2 |
| NotYet: passing union of differently held members to a function value | 2 |
| NotYet: reading IncrementalParser | 2 |
| NotYet: reading accessibleSymbolChain | 2 |
| NotYet: reading accessorDeclarations | 2 |
| NotYet: reading add | 2 |
| NotYet: reading allAccessors | 2 |
| NotYet: reading allowStructuralFallback | 2 |
| NotYet: reading assignClassAliasInStaticBlock | 2 |
| NotYet: reading assignment | 2 |
| NotYet: reading attributesType | 2 |
| NotYet: reading baseClassType | 2 |
| NotYet: reading baseSymbol | 2 |
| NotYet: reading cacheAssignment | 2 |
| NotYet: reading cachedPackageJson | 2 |
| NotYet: reading cachedTypes | 2 |
| NotYet: reading call | 2 |
| NotYet: reading callbackToAdd | 2 |
| NotYet: reading capturedLeft | 2 |
| NotYet: reading change0 | 2 |
| NotYet: reading classLikeDeclaration | 2 |
| NotYet: reading columns | 2 |
| NotYet: reading commentRange | 2 |
| NotYet: reading comments | 2 |
| NotYet: reading compilerHost | 2 |
| NotYet: reading constructSignatures | 2 |
| NotYet: reading containingCall | 2 |
| NotYet: reading containingFileName | 2 |
| NotYet: reading contextualType | 2 |
| NotYet: reading curCategory | 2 |
| NotYet: reading declBlocked | 2 |
| NotYet: reading declarationContainer | 2 |
| NotYet: reading declarationFilePath | 2 |
| NotYet: reading declarationTransform | 2 |
| NotYet: reading declarations | 2 |
| NotYet: reading declaredType | 2 |
| NotYet: reading decorators | 2 |
| NotYet: reading deduplicated | 2 |
| NotYet: reading defaultIndex | 2 |
| NotYet: reading derived | 2 |
| NotYet: reading diagnostic | 2 |
| NotYet: reading directoryPath | 2 |
| NotYet: reading dupFile | 2 |
| NotYet: reading effectiveTarget | 2 |
| NotYet: reading elementTypes | 2 |
| NotYet: reading emitFlags | 2 |
| NotYet: reading emitSkipped | 2 |
| NotYet: reading enclosingDeclaration | 2 |
| NotYet: reading endLabel | 2 |
| NotYet: reading endLength | 2 |
| NotYet: reading enter | 2 |
| NotYet: reading equalsToken | 2 |
| NotYet: reading errorInfo | 2 |
| NotYet: reading errorRecord | 2 |
| NotYet: reading escapedText | 2 |
| NotYet: reading expandedParams | 2 |
| NotYet: reading exportName | 2 |
| NotYet: reading exportStarFunction | 2 |
| NotYet: reading exportedName | 2 |
| NotYet: reading extraInitializersName | 2 |
| NotYet: reading fileInfos | 2 |
| NotYet: reading filePath | 2 |
| NotYet: reading first | 2 |
| NotYet: reading firstArgument | 2 |
| NotYet: reading firstBase | 2 |
| NotYet: reading firstParameterIsThis | 2 |
| NotYet: reading firstType | 2 |
| NotYet: reading flattenContext | 2 |
| NotYet: reading freshType | 2 |
| NotYet: reading fromCache | 2 |
| NotYet: reading fullName | 2 |
| NotYet: reading generatorYieldType | 2 |
| NotYet: reading getCanonicalFileName | 2 |
| NotYet: reading getCommonSourceDirectory | 2 |
| NotYet: reading getModifiedTime | 2 |
| NotYet: reading getOptionsNameMap | 2 |
| NotYet: reading gutterWidth | 2 |
| NotYet: reading hasArguments | 2 |
| NotYet: reading hasDefault | 2 |
| NotYet: reading heritageClause | 2 |
| NotYet: reading hooks | 2 |
| NotYet: reading indexSymbol | 2 |
| NotYet: reading init | 2 |
| NotYet: reading initializerStatement | 2 |
| NotYet: reading instanceType | 2 |
| NotYet: reading invalidatedProject | 2 |
| NotYet: reading isAnyLike | 2 |
| NotYet: reading isAutomaticTypeInNonNull | 2 |
| NotYet: reading isFinite | 2 |
| NotYet: reading isOptionalChain | 2 |
| NotYet: reading isSetonlyAccessor | 2 |
| NotYet: reading isSingleNonGenericCandidate | 2 |
| NotYet: reading isUsed | 2 |
| NotYet: reading isVoidPromiseError | 2 |
| NotYet: reading iteratedType | 2 |
| NotYet: reading json | 2 |
| NotYet: reading jsxFragmentFactoryName | 2 |
| NotYet: reading keyPropertyName | 2 |
| NotYet: reading label | 2 |
| NotYet: reading lanes | 2 |
| NotYet: reading last | 2 |
| NotYet: reading lastId | 2 |
| NotYet: reading lastModifier | 2 |
| NotYet: reading lastParamVariadicType | 2 |
| NotYet: reading leadingNewlines | 2 |
| NotYet: reading leftIdentifier | 2 |
| NotYet: reading length | 2 |
| NotYet: reading lineNumber | 2 |
| NotYet: reading lines | 2 |
| NotYet: reading linesBeforeDot | 2 |
| NotYet: reading linkType | 2 |
| NotYet: reading list | 2 |
| NotYet: reading literalType | 2 |
| NotYet: reading loadPackageJsonMainState | 2 |
| NotYet: reading local | 2 |
| NotYet: reading longestParamType | 2 |
| NotYet: reading mapped | 2 |
| NotYet: reading mapperCache | 2 |
| NotYet: reading match | 2 |
| NotYet: reading meaning | 2 |
| NotYet: reading methodParameterType | 2 |
| NotYet: reading methodSignatures | 2 |
| NotYet: reading min | 2 |
| NotYet: reading missing | 2 |
| NotYet: reading modifiers | 2 |
| NotYet: reading mustBeRemoved | 2 |
| NotYet: reading names | 2 |
| NotYet: reading newParametersArray | 2 |
| NotYet: reading newParsedCommandLine | 2 |
| NotYet: reading next | 2 |
| NotYet: reading nextChange | 2 |
| NotYet: reading nextKey | 2 |
| NotYet: reading nextLineStart | 2 |
| NotYet: reading noTruncation | 2 |
| NotYet: reading nodeId | 2 |
| NotYet: reading oldNoInferenceFallback | 2 |
| NotYet: reading oldState | 2 |
| NotYet: reading openBracePosition | 2 |
| NotYet: reading operandConstraint | 2 |
| NotYet: reading operatorKind | 2 |
| NotYet: reading operatorToken | 2 |
| NotYet: reading origin | 2 |
| NotYet: reading originalClassDecl | 2 |
| NotYet: reading otherAccessor | 2 |
| NotYet: reading override | 2 |
| NotYet: reading param | 2 |
| NotYet: reading paramCount | 2 |
| NotYet: reading paramSymbol | 2 |
| NotYet: reading parametersWithPropertyAssignments | 2 |
| NotYet: reading parentSymbol | 2 |
| NotYet: reading parenthesizerRule | 2 |
| NotYet: reading parseTreeNode | 2 |
| NotYet: reading parsedCommandLine | 2 |
| NotYet: reading patterns | 2 |
| NotYet: reading potentiallyUnusedIdentifiers | 2 |
| NotYet: reading pragma | 2 |
| NotYet: reading predicate | 2 |
| NotYet: reading previous | 2 |
| NotYet: reading primaryTypes | 2 |
| NotYet: reading propName | 2 |
| NotYet: reading prototypeProperty | 2 |
| NotYet: reading questionDotToken | 2 |
| NotYet: reading r | 2 |
| NotYet: reading real | 2 |
| NotYet: reading reducedTypes | 2 |
| NotYet: reading ref | 2 |
| NotYet: reading relatedInfo | 2 |
| NotYet: reading relativeFileName | 2 |
| NotYet: reading remove | 2 |
| NotYet: reading resolvedProject | 2 |
| NotYet: reading resolvedTypeSymbol | 2 |
| NotYet: reading restIdent | 2 |
| NotYet: reading returnMethod | 2 |
| NotYet: reading returnTypeNode | 2 |
| NotYet: reading root | 2 |
| NotYet: reading setter | 2 |
| NotYet: reading shortest | 2 |
| NotYet: reading signatureDeclaration | 2 |
| NotYet: reading sorted | 2 |
| NotYet: reading sourceFilePath | 2 |
| NotYet: reading sourceFileWithAddedExtension | 2 |
| NotYet: reading sourceFlags | 2 |
| NotYet: reading sourceMappings | 2 |
| NotYet: reading sourceRoot | 2 |
| NotYet: reading sourceStart | 2 |
| NotYet: reading sourceSymbolFile | 2 |
| NotYet: reading sourceType | 2 |
| NotYet: reading span | 2 |
| NotYet: reading specifierSourceImports | 2 |
| NotYet: reading spread | 2 |
| NotYet: reading spreadType | 2 |
| NotYet: reading staticType | 2 |
| NotYet: reading str | 2 |
| NotYet: reading substitute | 2 |
| NotYet: reading sym | 2 |
| NotYet: reading symbolName | 2 |
| NotYet: reading targetIndex | 2 |
| NotYet: reading targetProp | 2 |
| NotYet: reading templateArguments | 2 |
| NotYet: reading terminalWidth | 2 |
| NotYet: reading testedSymbol | 2 |
| NotYet: reading thisAccess | 2 |
| NotYet: reading timerToUpdateChildWatches | 2 |
| NotYet: reading toWatch | 2 |
| NotYet: reading tracingEnabled | 2 |
| NotYet: reading transformed | 2 |
| NotYet: reading trueType | 2 |
| NotYet: reading typeLiteralNode | 2 |
| NotYet: reading typeLiteralSymbol | 2 |
| NotYet: reading typeOnlyDeclaration | 2 |
| NotYet: reading typeParameter | 2 |
| NotYet: reading typeParams | 2 |
| NotYet: reading typesVersions | 2 |
| NotYet: reading useDefineForClassFields | 2 |
| NotYet: reading valueSymbol | 2 |
| NotYet: reading valueType | 2 |
| NotYet: reading varDecl | 2 |
| NotYet: reading variable | 2 |
| NotYet: reading versionPaths | 2 |
| NotYet: reading visitedAccessorName | 2 |
| NotYet: reading visitedSym | 2 |
| NotYet: reading widened | 2 |
| NotYet: reading yieldedType | 2 |
| NotYet: storing true \| Node \| undefined in a field | 2 |
| Refused: Object.create | 2 |
| Refused: a function taking ((node: Node) => VisitResult<Node>) \| undefined seen as one taking ((node: Node) => VisitResult<Node \| undefined>) \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking BinaryOperator seen as one taking SyntaxKind (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking Expression seen as one taking Expression \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking Expression[] seen as one taking readonly Expression[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking GeneratedIdentifierFlags seen as one taking GeneratedIdentifierFlags \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking Node seen as one taking [node: Node] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking NodeFlags seen as one taking NodeFlags \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking T seen as one taking Expression (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking Visitor seen as one taking Visitor<TIn, Node \| undefined> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking [fileName: string, languageVersionOrOptions: CreateSourceFileOptions \| ScriptTarget, onError?: ((message: string) => void) \| undefined, shouldCreateNewSourceFile?: boolean \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking string seen as one taking [fileName: string] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking string seen as one taking any[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a method read as a value (clearTimeout would lose its object, and this with it) | 2 |
| Refused: a method read as a value (createDirectory would lose its object, and this with it) | 2 |
| Refused: a method read as a value (emitNodeWithNotification would lose its object, and this with it) | 2 |
| Refused: a method read as a value (enableCPUProfiler would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getBuildInfo would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getCurrentDirectory would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getEnvironmentVariable would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getGlobalTypingsCacheLocation would lose its object, and this with it) | 2 |
| Refused: a method read as a value (hasGlobalName would lose its object, and this with it) | 2 |
| Refused: a method read as a value (isEmitNotificationEnabled would lose its object, and this with it) | 2 |
| Refused: a method read as a value (parenthesizeBranchOfConditionalExpression would lose its object, and this with it) | 2 |
| Refused: a method read as a value (parenthesizeConstituentTypesOfIntersectionType would lose its object, and this with it) | 2 |
| Refused: a method read as a value (parenthesizeConstituentTypesOfUnionType would lose its object, and this with it) | 2 |
| Refused: a method read as a value (parenthesizeNonArrayTypeOfPostfixType would lose its object, and this with it) | 2 |
| Refused: a method read as a value (reportInaccessibleUniqueSymbolError would lose its object, and this with it) | 2 |
| Refused: a method read as a value (setPrototypeOf would lose its object, and this with it) | 2 |
| Refused: a method read as a value (setTimeout would lose its object, and this with it) | 2 |
| Refused: a method read as a value (substituteNode would lose its object, and this with it) | 2 |
| Refused: a method read as a value (toKey would lose its object, and this with it) | 2 |
| Refused: a method read as a value (useCaseSensitiveFileNames would lose its object, and this with it) | 2 |
| Refused: a method read as a value (watchDirectory would lose its object, and this with it) | 2 |
| Refused: a method read as a value (watchFile would lose its object, and this with it) | 2 |
| Refused: a method read as a value (writeFile would lose its object, and this with it) | 2 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on declaration) | 2 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on predicate) | 2 |
| Refused: a value of type (identifierOrPrivateName: Identifier \| PrivateIdentifier) => string seen as (name: GeneratedIdentifier \| GeneratedPrivateIdentifier) => string, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 2 |
| Refused: a value of type (readonly [() => IntrinsicType, __String])[] seen as (readonly [() => Type, __String])[], which can write readonly [() => Type, __String] where readonly [() => IntrinsicType, __String] is read | 2 |
| Refused: a value of type ArrayLiteralExpression \| AssignmentExpression<EqualsToken> \| BindingElement \| ElementAccessExpression \| ... 8 more ... \| VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type BindingElement[] seen as unknown[], which can write unknown where BindingElement is read | 2 |
| Refused: a value of type Block seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type BlockLike seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type BreakStatement seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type CapturedThis seen as Expression \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 2 |
| Refused: a value of type CapturedThis seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 2 |
| Refused: a value of type ClassElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ClassStaticBlockDeclaration seen as Mutable<ClassStaticBlockDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ClassStaticBlockDeclaration \| PropertyDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type CompilerOptions & { types: string[]; } seen as CompilerOptions, which can write string[] \| undefined where string[] is read | 2 |
| Refused: a value of type ConstructorDeclaration seen as Mutable<ConstructorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ContinueStatement seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type DiagnosticWithLocation \| undefined seen as Diagnostic \| undefined, which can write SourceFile \| undefined where SourceFile is read | 2 |
| Refused: a value of type DiagnosticWithLocation[] \| undefined seen as DiagnosticRelatedInformation[] \| undefined, which can write DiagnosticRelatedInformation where DiagnosticWithLocation is read | 2 |
| Refused: a value of type ElementAccessExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type FlowCall seen as FlowNode, which can write BinaryExpression \| CallExpression where CallExpression is read | 2 |
| Refused: a value of type FutureSourceFile \| SourceFile seen as Pick<SourceFile, "fileName" \| "impliedNodeFormat" \| "packageJsonScope">, whose readonly field fileName becomes writable: a readonly field may hold something narrower than string, which a write of string would replace | 2 |
| Refused: a value of type GeneratedIdentifier seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 2 |
| Refused: a value of type GeneratedIdentifier seen as Identifier \| PrivateIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 2 |
| Refused: a value of type Identifier seen as Mutable<Identifier>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Identifier[] \| undefined seen as ModuleExportName[] \| undefined, which can write ModuleExportName where Identifier is read | 2 |
| Refused: a value of type ImportTypeAssertionContainer seen as Mutable<ImportTypeAssertionContainer>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type LeftHandSideExpression & GeneratedIdentifier seen as Expression, which can write EmitNode \| undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 2 |
| Refused: a value of type MethodDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type MissingDeclaration seen as Mutable<MissingDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ModifierLike seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ModuleSpecifierResolutionHost & ModuleResolutionHost seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where (() => boolean) & (boolean \| (() => boolean) \| undefined) is read | 2 |
| Refused: a value of type Node seen as Node \| TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Node \| TextRange seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Node \| undefined seen as NodeLinks \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of NodeCheckFlags would replace | 2 |
| Refused: a value of type NumericLiteral seen as Mutable<LiteralExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ParameterDeclaration \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 2 |
| Refused: a value of type PropertySignature seen as Mutable<PropertySignature>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ReturnStatement seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type SourceFile seen as EmitNode \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of EmitFlags would replace | 2 |
| Refused: a value of type SourceFile seen as Mutable<SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Statement seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Symbol seen as Declaration \| undefined, which can write number \| undefined where number is read | 2 |
| Refused: a value of type Symbol[] seen as unknown[], which can write unknown where Symbol is read | 2 |
| Refused: a value of type T seen as T, a type parameter whose constraint Node can be written, so it can write what T can't hold | 2 |
| Refused: a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Type[] seen as unknown[], which can write unknown where Type is read | 2 |
| Refused: a value of type VariableDeclaration seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type VariableStatement seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type any seen as T, a type parameter whose constraint Node can be written, so it can write what any can't hold | 2 |
| Refused: a value of type false \| FlowNode \| undefined seen as false \| FlowAssignment \| FlowLabel \| FlowReduceLabel \| FlowStart \| FlowSwitchClause \| undefined, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 2 |
| Refused: a value of type never[] seen as (Identifier \| StringLiteral)[], which can write Identifier \| StringLiteral where never is read | 2 |
| Refused: a value of type never[] seen as ChildDirectoryWatcher[], which can write ChildDirectoryWatcher where never is read | 2 |
| Refused: a value of type never[] seen as DiagnosticWithLocation[], which can write DiagnosticWithLocation where never is read | 2 |
| Refused: a value of type never[] seen as SourceFile[], which can write SourceFile where never is read | 2 |
| Refused: a value of type never[] seen as StringLiteralLike[], which can write StringLiteralLike where never is read | 2 |
| Refused: a value of type never[] seen as VariableDeclaration[], which can write VariableDeclaration where never is read | 2 |
| Refused: a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[] \| undefined, which can write SourceFile \| undefined where SourceFile is read | 2 |
| Refused: a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 2 |
| Refused: a value of type string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] seen as unknown[], which can write unknown where string is read | 2 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"FlowFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof FlowFlags is read | 2 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint Type can be written, so it can write what undefined can't hold | 2 |
| Refused: an unproven relation from Declaration to NamedDeclaration: optional field name has no proven compatible presence/type | 2 |
| Refused: an unproven relation from ObjectType to AnonymousType: optional field target has no proven compatible presence/type | 2 |
| Refused: an unproven relation from Type to SyntheticDefaultModuleType: optional field syntheticType has no proven compatible presence/type | 2 |
| Refused: an unproven relation from Type to TypeVariable: optional field constraint has no proven compatible presence/type | 2 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.FunctionExpression | 2 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NoSubstitutionTemplateLiteral \| SyntaxKind.TemplateHead \| SyntaxKind.TemplateMiddle \| SyntaxKind.TemplateTail | 2 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.WithKeyword \| SyntaxKind.AssertKeyword | 2 |
| Refused: an unproven value assigned to a numeric literal or enum member slot Ternary | 2 |
| Refused: delete | 2 |
| Refused: inherited library member setPrototypeOf read as an own field | 2 |
| Refused: optional property all in CompilerOptions absent from structural source {}, which can hide fields | 2 |
| Refused: optional property constraint in TypeParameter absent from structural source InterfaceType, which can hide fields | 2 |
| Refused: optional property equalsToken in ShorthandPropertyAssignment absent from structural source never, which can hide fields | 2 |
| Refused: optional property id in NamedExports absent from structural source never, which can hide fields | 2 |
| Refused: optional property modifiers in GetAccessorDeclaration absent from structural source never, which can hide fields | 2 |
| Refused: optional property packageJsonScope in Pick<SourceFile, "fileName" \| "impliedNodeFormat" \| "packageJsonScope"> absent from structural source Pick<SourceFile, "fileName" \| "impliedNodeFormat">, which can hide fields | 2 |
| Refused: optional property return in ArrayIterator<Diagnostic> absent from structural source ArrayIterator<DiagnosticWithLocation>, which can hide fields | 2 |
| Refused: optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<JSDocTag>, which can hide fields | 2 |
| Refused: optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<TemplateLiteralTypeSpan>, which can hide fields | 2 |
| Refused: optional property templateFlags in NoSubstitutionTemplateLiteral absent from structural source never, which can hide fields | 2 |
| Refused: optional property textSourceNode in StringLiteral absent from structural source never, which can hide fields | 2 |
| panic: statement panic: Unhandled case in Node.Text: *ast.ComputedPropertyName | 2 |
| NotYet: .length on a value | 1 |
| NotYet: ?. to a number, which would be number \| undefined | 1 |
| NotYet: Array as a value outside equality or typeof (overloaded calls and static properties need their own representation) | 1 |
| NotYet: JSON.stringify a union containing containers without runtime element metadata | 1 |
| NotYet: JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) | 1 |
| NotYet: Object.assign on a shape not proven by a plain literal or its const binding | 1 |
| NotYet: a BinaryExpression with a boolean and a boolean | 1 |
| NotYet: a BinaryExpression with a boolean \| undefined and a boolean \| undefined | 1 |
| NotYet: a BinaryExpression with a boolean \| undefined and a value | 1 |
| NotYet: a BinaryExpression with a number and a union of differently held members | 1 |
| NotYet: a BinaryExpression with a number \| undefined and a boolean \| undefined | 1 |
| NotYet: a BinaryExpression with a number \| undefined and a string | 1 |
| NotYet: a BinaryExpression with a union of differently held members and a union of differently held members | 1 |
| NotYet: a BinaryExpression with a value and a boolean \| undefined | 1 |
| NotYet: a ClassExpression | 1 |
| NotYet: a Map of CompilerOptionsValue | 1 |
| NotYet: a Map of RedirectsCacheKey | 1 |
| NotYet: a Map of ResolvedConfigFilePath | 1 |
| NotYet: a Map of T | 1 |
| NotYet: a Map of string \| number | 1 |
| NotYet: a Map whose key and value types aren't known | 1 |
| NotYet: a PostfixUnaryExpression | 1 |
| NotYet: a Set of ResolvedConfigFilePath (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far) | 1 |
| NotYet: a SpreadElement | 1 |
| NotYet: a base that isn't a declared class | 1 |
| NotYet: a call returning T | 1 |
| NotYet: a call returning any | 1 |
| NotYet: a case that isn't a constant | 1 |
| NotYet: a class instantiated with TOuterState | 1 |
| NotYet: a class method through a view that erases its prototype origin | 1 |
| NotYet: a comparator that doesn't take two elements and return a number | 1 |
| NotYet: a destructured name held otherwise than its field | 1 |
| NotYet: a destructured parameter beside a parameter with a default | 1 |
| NotYet: a field from a boolean \| undefined variable | 1 |
| NotYet: a field of type "boolean" \| "list" \| "number" \| "object" \| "string" \| Map<string, string \| number> | 1 |
| NotYet: a field of type "boolean" \| "number" \| "object" \| "string" \| Map<string, string \| number> | 1 |
| NotYet: a field of type "circularity" \| boolean | 1 |
| NotYet: a field of type 0 \| boolean \| undefined | 1 |
| NotYet: a field of type NodeArray<ParameterDeclaration> \| readonly JSDocParameterTag[] | 1 |
| NotYet: a field of type boolean \| (() => boolean) | 1 |
| NotYet: a field of type false \| Type \| undefined | 1 |
| NotYet: a field of type false \| VersionPaths | 1 |
| NotYet: a field of type false \| VersionPaths \| undefined | 1 |
| NotYet: a field of type false \| string[] \| undefined | 1 |
| NotYet: a field of type string \| false | 1 |
| NotYet: a field of type string \| false \| undefined | 1 |
| NotYet: a for...of destructuring an object | 1 |
| NotYet: a function returning () => T | 1 |
| NotYet: a function returning (...args: T) => boolean | 1 |
| NotYet: a function returning (AssignmentExpression<EqualsToken> & { readonly left: GeneratedIdentifier; }) \| undefined | 1 |
| NotYet: a function returning (ConstructorDeclaration & { body: Block; }) \| undefined | 1 |
| NotYet: a function returning (T \| U)[] | 1 |
| NotYet: a function returning (arg: A) => T | 1 |
| NotYet: a function returning (node: BinaryExpression, outerState: TOuterState) => TResult | 1 |
| NotYet: a function returning A | 1 |
| NotYet: a function returning AnyValidImportOrReExport | 1 |
| NotYet: a function returning AnyValidImportOrReExport \| undefined | 1 |
| NotYet: a function returning BuildInvalidedProject<T> | 1 |
| NotYet: a function returning BuildInvalidedProject<T> \| UpdateOutputFileStampsProject | 1 |
| NotYet: a function returning CacheWithRedirects<K, V> | 1 |
| NotYet: a function returning CanonicalKey | 1 |
| NotYet: a function returning CapturedThis | 1 |
| NotYet: a function returning ClassExpression \| ImmediatelyInvokedArrowFunction | 1 |
| NotYet: a function returning ClassNamedEvaluationHelperBlock | 1 |
| NotYet: a function returning ClassStaticBlockDeclaration \| Decorator \| PrivateIdentifierGetAccessorDeclaration \| ... 5 more ... \| undefined | 1 |
| NotYet: a function returning ClassThisAssignmentBlock | 1 |
| NotYet: a function returning Declaration & HasModifiers | 1 |
| NotYet: a function returning EndOfFileToken | 1 |
| NotYet: a function returning EvaluatorResult<T> | 1 |
| NotYet: a function returning ExpressionWithTypeArguments & { expression: Identifier \| PropertyAccessEntityNameExpression; } | 1 |
| NotYet: a function returning ExpressionWithTypeArguments & { readonly expression: Identifier \| PropertyAccessEntityNameExpression; } | 1 |
| NotYet: a function returning Extract<AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; }, Pick<...>> \| ... 7 more ... \| Extract<...> | 1 |
| NotYet: a function returning HasJSDoc \| undefined | 1 |
| NotYet: a function returning IncrementalBuildInfoFileId | 1 |
| NotYet: a function returning IncrementalBuildInfoFileIdListId | 1 |
| NotYet: a function returning InferenceContext \| (T & undefined) | 1 |
| NotYet: a function returning InvalidatedProject<T> \| undefined | 1 |
| NotYet: a function returning Map<K, V1 \| V2> | 1 |
| NotYet: a function returning MemberName \| (Expression & (NumericLiteral \| StringLiteralLike)) | 1 |
| NotYet: a function returning MissingList<T> | 1 |
| NotYet: a function returning ModeAwareCache<T> | 1 |
| NotYet: a function returning ModuleOrTypeReferenceResolutionCache<T> | 1 |
| NotYet: a function returning MultiMap<K, V> | 1 |
| NotYet: a function returning Mutable<Token<TKind>> | 1 |
| NotYet: a function returning Node \| (TIn & undefined) \| (TVisited & undefined) | 1 |
| NotYet: a function returning NodeArray<Node> \| (TInArray & undefined) | 1 |
| NotYet: a function returning NodeArray<NonNullable<T>> \| undefined | 1 |
| NotYet: a function returning NodeArray<TOut> \| (TInArray & undefined) | 1 |
| NotYet: a function returning NonNullable<T> | 1 |
| NotYet: a function returning NonRelativeNameResolutionCache<T> | 1 |
| NotYet: a function returning PathPathComponents | 1 |
| NotYet: a function returning PerDirectoryResolutionCache<T> | 1 |
| NotYet: a function returning RedirectsCacheKey | 1 |
| NotYet: a function returning ResolvedConfigFilePath | 1 |
| NotYet: a function returning ReusableDiagnosticMessageChain | 1 |
| NotYet: a function returning SolutionBuilder<T> | 1 |
| NotYet: a function returning SyntheticSuper | 1 |
| NotYet: a function returning T \| EmptyStatement | 1 |
| NotYet: a function returning T \| Identifier | 1 |
| NotYet: a function returning T \| NumericLiteral \| StringLiteral \| BooleanLiteral | 1 |
| NotYet: a function returning T \| StringLiteral | 1 |
| NotYet: a function returning T \| T[] | 1 |
| NotYet: a function returning T \| readonly T[] | 1 |
| NotYet: a function returning T1 & T2 | 1 |
| NotYet: a function returning TEntry \| undefined | 1 |
| NotYet: a function returning TOut | 1 |
| NotYet: a function returning TOut \| (TIn & undefined) \| (TVisited & undefined) | 1 |
| NotYet: a function returning TOut \| undefined | 1 |
| NotYet: a function returning TPrivateEntry \| undefined | 1 |
| NotYet: a function returning TResult | 1 |
| NotYet: a function returning T[][] | 1 |
| NotYet: a function returning TransformationResult<T> | 1 |
| NotYet: a function returning TypeMapper \| (T & undefined) | 1 |
| NotYet: a function returning TypeOnlyAliasDeclaration \| undefined | 1 |
| NotYet: a function returning V \| undefined | 1 |
| NotYet: a function returning V[] | 1 |
| NotYet: a function returning VisitResult<T> | 1 |
| NotYet: a function returning WatchCompilerHostOfConfigFile<T> | 1 |
| NotYet: a function returning WatchCompilerHostOfFilesAndCompilerOptions<T> | 1 |
| NotYet: a function returning never | 1 |
| NotYet: a function returning object \| undefined | 1 |
| NotYet: a function returning readonly (readonly T[])[] | 1 |
| NotYet: a function returning readonly Node[] \| (TInArray & undefined) | 1 |
| NotYet: a function returning readonly Resolution[] | 1 |
| NotYet: a function returning readonly TOut[] \| (TInArray & undefined) | 1 |
| NotYet: a function returning readonly U[] | 1 |
| NotYet: a function returning readonly U[] \| undefined | 1 |
| NotYet: a function returning string \| object | 1 |
| NotYet: a function returning unknown | 1 |
| NotYet: a function returning void \| SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> | 1 |
| NotYet: a function returning void \| SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> \| WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> | 1 |
| NotYet: a function returning void \| WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> | 1 |
| NotYet: a function returning void \| number \| Symbol | 1 |
| NotYet: a function returning { [P in K as `${P}`]?: T[]; } | 1 |
| NotYet: a function returning { modifiers: NodeArray<Modifier> \| undefined; referencedName: Expression \| undefined; name: PropertyName; initializersName: Identifier \| undefined; descriptorName: Identifier \| undefined; thisArg: Identifier \| undefined; extraInitializersName?: never; } \| ... | 1 |
| NotYet: a function returning { readonly min: number; readonly max: number; } | 1 |
| NotYet: a function value taking boolean \| undefined | 1 |
| NotYet: a literal method through a view that erases its receiver | 1 |
| NotYet: a non-accessor method in an accessor literal | 1 |
| NotYet: a number \| undefined argument to slice | 1 |
| NotYet: a number \| undefined argument to substring | 1 |
| NotYet: a template interpolating an object, an array, a map, a function or undefined | 1 |
| NotYet: a tuple literal leaving out an element of type ModuleSpecifierEnding | 1 |
| NotYet: a value of type "" \| ResolvedConfigFileName \| undefined | 1 |
| NotYet: a value of type ((t: T) => U) \| undefined | 1 |
| NotYet: a value of type (AmbientModuleDeclaration & { name: StringLiteral; }) \| undefined | 1 |
| NotYet: a value of type (AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; } & BinaryExpression) \| (... & ... 1 more ... & BinaryExpression) | 1 |
| NotYet: a value of type (ExportDeclaration & { readonly isTypeOnly: true; readonly moduleSpecifier: Expression; }) \| undefined | 1 |
| NotYet: a value of type (ModuleDeclaration & { name: StringLiteral; }) \| undefined | 1 |
| NotYet: a value of type (VariableDeclaration & { name: Identifier; }) \| undefined | 1 |
| NotYet: a value of type (s: string) => void | 1 |
| NotYet: a value of type AccessExpression \| RequireOrImportCall | 1 |
| NotYet: a value of type AccessorDeclaration & { readonly name: BigIntLiteral \| ComputedPropertyName \| Identifier \| NoSubstitutionTemplateLiteral \| NumericLiteral \| StringLiteral; } | 1 |
| NotYet: a value of type AliasDeclarationNode | 1 |
| NotYet: a value of type ArrowFunction \| BinaryExpression \| BindingElement \| Block \| BreakStatement \| CallSignatureDeclaration \| ... 64 more ... \| EndOfFileToken | 1 |
| NotYet: a value of type BindablePropertyAssignmentExpression \| PropertyAccessExpression \| LiteralLikeElementAccessExpression | 1 |
| NotYet: a value of type BindableStaticAccessExpression | 1 |
| NotYet: a value of type BindingElement & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type CallExpression \| BindableStaticAccessExpression | 1 |
| NotYet: a value of type Child | 1 |
| NotYet: a value of type ClassElement \| ParameterPropertyDeclaration | 1 |
| NotYet: a value of type ClassNamedEvaluationHelperBlock | 1 |
| NotYet: a value of type CustomTransformerFactory \| TransformerFactory<T> | 1 |
| NotYet: a value of type EmitNode & { autoGenerate: AutoGenerateInfo; } | 1 |
| NotYet: a value of type EntityNameExpression \| (LeftHandSideExpression & BindableStaticNameExpression) | 1 |
| NotYet: a value of type ExportAssignment & { readonly expression: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type Identifier \| PrivateIdentifier \| __String | 1 |
| NotYet: a value of type IncludeTypeSpaceImports | 1 |
| NotYet: a value of type IncrementalBuildInfoFileIdListId | 1 |
| NotYet: a value of type IncrementalBuildInfoFilePendingEmit | 1 |
| NotYet: a value of type IncrementalMultiFileEmitBuildInfoFileInfo | 1 |
| NotYet: a value of type JSDocImportTag \| CanHaveModuleSpecifier | 1 |
| NotYet: a value of type LeftHandSideExpression & Identifier | 1 |
| NotYet: a value of type Map<Path, ModeAwareCache<T>> \| undefined | 1 |
| NotYet: a value of type Map<string, SingleFileWatcher<T>> | 1 |
| NotYet: a value of type Map<string, WildcardDirectoryWatcher<T>> | 1 |
| NotYet: a value of type Map<string, [K, V[]]> | 1 |
| NotYet: a value of type MapLike<T> | 1 |
| NotYet: a value of type ModeAwareCacheKey | 1 |
| NotYet: a value of type NamedEvaluation | 1 |
| NotYet: a value of type NodeArray<Expression> & readonly [BindableStaticNameExpression, NumericLiteral \| StringLiteralLike, ObjectLiteralExpression] & Readonly<...> | 1 |
| NotYet: a value of type NodeArray<T> | 1 |
| NotYet: a value of type NonNullable<K> | 1 |
| NotYet: a value of type NonNullable<U> | 1 |
| NotYet: a value of type ParameterDeclaration & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type PathPathComponents | 1 |
| NotYet: a value of type PrimitiveLiteral | 1 |
| NotYet: a value of type PrivateEnvironment<TData, TEntry> | 1 |
| NotYet: a value of type PrivateIdentifierInExpression | 1 |
| NotYet: a value of type PropertyAccessExpression \| LiteralLikeElementAccessExpression | 1 |
| NotYet: a value of type PropertyAccessExpression \| SyntheticSuper | 1 |
| NotYet: a value of type PropertyAssignment & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type PropertyDeclaration & { readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type PropertyDeclaration \| ParameterPropertyDeclaration | 1 |
| NotYet: a value of type RedirectsCacheKey | 1 |
| NotYet: a value of type RedirectsCacheKey \| undefined | 1 |
| NotYet: a value of type ReferencedFile & { kind: FileIncludeKind.LibReferenceDirective; } | 1 |
| NotYet: a value of type ReplaceableIndexedAccessType | 1 |
| NotYet: a value of type RequireOrImportCall | 1 |
| NotYet: a value of type ResolvedConfigFileName \| undefined | 1 |
| NotYet: a value of type ResolvedModuleWithFailedLookupLocations & ResolvedTypeReferenceDirectiveWithFailedLookupLocations | 1 |
| NotYet: a value of type Set<K> | 1 |
| NotYet: a value of type ShorthandPropertyAssignment & { readonly objectAssignmentInitializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type SortedArray<T> | 1 |
| NotYet: a value of type Source | 1 |
| NotYet: a value of type SourceFileOrString | 1 |
| NotYet: a value of type T \| T[] \| readonly T[] \| undefined | 1 |
| NotYet: a value of type T1 | 1 |
| NotYet: a value of type TData | 1 |
| NotYet: a value of type TEntry | 1 |
| NotYet: a value of type TKind \| Token<TKind> | 1 |
| NotYet: a value of type TNode | 1 |
| NotYet: a value of type T[] \| undefined | 1 |
| NotYet: a value of type TransformedSuperCall | 1 |
| NotYet: a value of type TypeParameterDeclaration & { parent: JSDocTemplateTag; } | 1 |
| NotYet: a value of type UnaryExpression & (BigIntLiteral \| NumericLiteral) | 1 |
| NotYet: a value of type UnaryExpression & NumericLiteral | 1 |
| NotYet: a value of type V | 1 |
| NotYet: a value of type VariableDeclaration & { name: Identifier; } | 1 |
| NotYet: a value of type VariableDeclaration & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type VariableDeclarationList & { _usingBrand: void; } | 1 |
| NotYet: a value of type WatchFactoryHost & { trace?(s: string): void; } | 1 |
| NotYet: a value of type __String & string | 1 |
| NotYet: a value of type boolean \| V \| undefined | 1 |
| NotYet: a value of type readonly (readonly T[])[] | 1 |
| NotYet: a value of type readonly IncrementalBundleEmitBuildInfoFileInfo[] \| readonly IncrementalMultiFileEmitBuildInfoFileInfo[] where an array goes | 1 |
| NotYet: a value of type readonly K[] | 1 |
| NotYet: a value of type string \| null | 1 |
| NotYet: a value of type string \| object \| undefined | 1 |
| NotYet: a value of type { forEach: (callbackfn: (value: T, key: K, map: Map<K, T>) => void, thisArg?: any) => void; clear: () => void; } | 1 |
| NotYet: an array of CanonicalKey | 1 |
| NotYet: an array of Path | 1 |
| NotYet: an array of T \| U | 1 |
| NotYet: an array of TState | 1 |
| NotYet: an array of object | 1 |
| NotYet: an array of unknown | 1 |
| NotYet: destructuring a string | 1 |
| NotYet: for...in over an array (holes and own enumerable properties are not represented; use for...of for elements) | 1 |
| NotYet: new Map from something that isn't [key, value] pairs | 1 |
| NotYet: passing boolean \| undefined to a function value | 1 |
| NotYet: reading absoluteSourceFilePath | 1 |
| NotYet: reading abstractSignatures | 1 |
| NotYet: reading accessibleSymbolsFromExports | 1 |
| NotYet: reading accessorType | 1 |
| NotYet: reading activeLabel | 1 |
| NotYet: reading actualFileName | 1 |
| NotYet: reading addAggregateStatistic | 1 |
| NotYet: reading addOutput | 1 |
| NotYet: reading afterImportPos | 1 |
| NotYet: reading afterImportTagPos | 1 |
| NotYet: reading aliasDecl | 1 |
| NotYet: reading allSetOptions | 1 |
| NotYet: reading allowEmpty | 1 |
| NotYet: reading allowsStrings | 1 |
| NotYet: reading alreadyTransformed | 1 |
| NotYet: reading alternateResult | 1 |
| NotYet: reading alternateResultMessage | 1 |
| NotYet: reading ambientModuleDeclare | 1 |
| NotYet: reading ancestor | 1 |
| NotYet: reading annotatedNodes | 1 |
| NotYet: reading annotationSymbol | 1 |
| NotYet: reading applicableByArity | 1 |
| NotYet: reading arrayType | 1 |
| NotYet: reading arrowExpression | 1 |
| NotYet: reading attrs | 1 |
| NotYet: reading autoGenerateFlags | 1 |
| NotYet: reading awaitToken | 1 |
| NotYet: reading baseConstraint | 1 |
| NotYet: reading baseConstructorType | 1 |
| NotYet: reading baseName | 1 |
| NotYet: reading baseType | 1 |
| NotYet: reading baseTypeNodes | 1 |
| NotYet: reading bases | 1 |
| NotYet: reading bestMatchingType | 1 |
| NotYet: reading binaryExpression | 1 |
| NotYet: reading bindingElement | 1 |
| NotYet: reading bindingList | 1 |
| NotYet: reading bindingTarget | 1 |
| NotYet: reading bold | 1 |
| NotYet: reading build | 1 |
| NotYet: reading buildArray | 1 |
| NotYet: reading buildOrderFromState | 1 |
| NotYet: reading builtins | 1 |
| NotYet: reading cachedDiagnostics | 1 |
| NotYet: reading cachedResolvedSignatures | 1 |
| NotYet: reading callArgument | 1 |
| NotYet: reading callSignatures | 1 |
| NotYet: reading callTarget | 1 |
| NotYet: reading canExcludeDiscriminants | 1 |
| NotYet: reading canUseBreakOrContinue | 1 |
| NotYet: reading candidateDirectories | 1 |
| NotYet: reading candidateExists | 1 |
| NotYet: reading caseType | 1 |
| NotYet: reading ch1 | 1 |
| NotYet: reading ch2 | 1 |
| NotYet: reading ch3 | 1 |
| NotYet: reading chain1 | 1 |
| NotYet: reading changed | 1 |
| NotYet: reading charCode | 1 |
| NotYet: reading checkAttributesType | 1 |
| NotYet: reading checkBody | 1 |
| NotYet: reading checkTypeDeferred | 1 |
| NotYet: reading childFieldType | 1 |
| NotYet: reading childPropName | 1 |
| NotYet: reading childrenPropName | 1 |
| NotYet: reading classInstanceType | 1 |
| NotYet: reading className | 1 |
| NotYet: reading classSymbol | 1 |
| NotYet: reading classThis | 1 |
| NotYet: reading cloned | 1 |
| NotYet: reading closingLineTerminatorCount | 1 |
| NotYet: reading code2 | 1 |
| NotYet: reading collidingSymbol | 1 |
| NotYet: reading combined | 1 |
| NotYet: reading commentEnd | 1 |
| NotYet: reading commentText | 1 |
| NotYet: reading commonJSPropertyAccess | 1 |
| NotYet: reading commonResolved | 1 |
| NotYet: reading compareResult | 1 |
| NotYet: reading comparer | 1 |
| NotYet: reading compilerFilePath | 1 |
| NotYet: reading compilerOptionsProperty | 1 |
| NotYet: reading componentEqualityComparer | 1 |
| NotYet: reading computedPropertyName | 1 |
| NotYet: reading conditional | 1 |
| NotYet: reading configFileText | 1 |
| NotYet: reading constantValue | 1 |
| NotYet: reading constraintDeclaration | 1 |
| NotYet: reading constraintNode | 1 |
| NotYet: reading constructorDeclaration | 1 |
| NotYet: reading constructorFunction | 1 |
| NotYet: reading constructorLikeName | 1 |
| NotYet: reading constructorSymbol | 1 |
| NotYet: reading constructorTypeCount | 1 |
| NotYet: reading containerObjectType | 1 |
| NotYet: reading containingClassDecl | 1 |
| NotYet: reading containingMethod | 1 |
| NotYet: reading containingSourceFile | 1 |
| NotYet: reading contextSpecifier | 1 |
| NotYet: reading contextType | 1 |
| NotYet: reading contextualAwaitedType | 1 |
| NotYet: reading convertToFunctionBlock | 1 |
| NotYet: reading converters | 1 |
| NotYet: reading create | 1 |
| NotYet: reading createBaseSourceFileNode | 1 |
| NotYet: reading createNodeArray | 1 |
| NotYet: reading currentDetachedCommentInfo | 1 |
| NotYet: reading currentOptions | 1 |
| NotYet: reading currentWriterIndentSpacing | 1 |
| NotYet: reading customTransformer | 1 |
| NotYet: reading d | 1 |
| NotYet: reading data | 1 |
| NotYet: reading declarationName | 1 |
| NotYet: reading declaringClassDeclaration | 1 |
| NotYet: reading defaultType | 1 |
| NotYet: reading delim | 1 |
| NotYet: reading deprecatedTag | 1 |
| NotYet: reading diagnosticMessage | 1 |
| NotYet: reading diagnosticWithLocation | 1 |
| NotYet: reading dir | 1 |
| NotYet: reading dirPath | 1 |
| NotYet: reading discriminantCombinations | 1 |
| NotYet: reading discriminantType | 1 |
| NotYet: reading disposeScope | 1 |
| NotYet: reading dist | 1 |
| NotYet: reading doneType | 1 |
| NotYet: reading downlevelIteration | 1 |
| NotYet: reading downleveledImport | 1 |
| NotYet: reading effectiveExpr | 1 |
| NotYet: reading elementAccess | 1 |
| NotYet: reading elementFlags | 1 |
| NotYet: reading elements | 1 |
| NotYet: reading emitAsSingleStatement | 1 |
| NotYet: reading emitComments | 1 |
| NotYet: reading emitExplicitInitializer | 1 |
| NotYet: reading emitSourceMaps | 1 |
| NotYet: reading emitTrailingComma | 1 |
| NotYet: reading emittedAsTopLevel | 1 |
| NotYet: reading enclosingBlockScopeContainer | 1 |
| NotYet: reading enclosingClass | 1 |
| NotYet: reading enclosingContainer | 1 |
| NotYet: reading encodeURI | 1 |
| NotYet: reading end | 1 |
| NotYet: reading enqueue | 1 |
| NotYet: reading entry | 1 |
| NotYet: reading enumResult | 1 |
| NotYet: reading enumStatement | 1 |
| NotYet: reading envVarStatement | 1 |
| NotYet: reading equalityComparer | 1 |
| NotYet: reading errNode | 1 |
| NotYet: reading errorMessage | 1 |
| NotYet: reading errorSpan | 1 |
| NotYet: reading evaluateEntityNameExpression | 1 |
| NotYet: reading everyClauseChecks | 1 |
| NotYet: reading exclamationToken | 1 |
| NotYet: reading excludePattern | 1 |
| NotYet: reading excludeRe | 1 |
| NotYet: reading excludedProperties | 1 |
| NotYet: reading existingPending | 1 |
| NotYet: reading existingProp | 1 |
| NotYet: reading existingSpecifier | 1 |
| NotYet: reading existingTarget | 1 |
| NotYet: reading exit | 1 |
| NotYet: reading exitStatus | 1 |
| NotYet: reading exportClause | 1 |
| NotYet: reading exportContainer | 1 |
| NotYet: reading exportStars | 1 |
| NotYet: reading exported | 1 |
| NotYet: reading exportedNamesStorageRef | 1 |
| NotYet: reading expressions | 1 |
| NotYet: reading ext | 1 |
| NotYet: reading extendsType | 1 |
| NotYet: reading externalHelpersImportDeclaration | 1 |
| NotYet: reading externalHelpersModuleName | 1 |
| NotYet: reading externalHelpersModuleReference | 1 |
| NotYet: reading f1 | 1 |
| NotYet: reading factory | 1 |
| NotYet: reading facts | 1 |
| NotYet: reading failed | 1 |
| NotYet: reading failedSignatureDeclarations | 1 |
| NotYet: reading falseSubtype | 1 |
| NotYet: reading fileDiags | 1 |
| NotYet: reading fileExists | 1 |
| NotYet: reading fileName | 1 |
| NotYet: reading fileToErrorCount | 1 |
| NotYet: reading filesForEmit | 1 |
| NotYet: reading filesInError | 1 |
| NotYet: reading filesToDelete | 1 |
| NotYet: reading filtered | 1 |
| NotYet: reading finalizeBoundary | 1 |
| NotYet: reading finished | 1 |
| NotYet: reading firstAccessor | 1 |
| NotYet: reading firstChar | 1 |
| NotYet: reading firstComponent | 1 |
| NotYet: reading firstInterfaceDecl | 1 |
| NotYet: reading firstNonzeroSegment | 1 |
| NotYet: reading firstRelevantLocation | 1 |
| NotYet: reading firstStatement | 1 |
| NotYet: reading firstVariableMatch | 1 |
| NotYet: reading fixed | 1 |
| NotYet: reading following | 1 |
| NotYet: reading forInitializer | 1 |
| NotYet: reading forStatement | 1 |
| NotYet: reading forcedLookupLocation | 1 |
| NotYet: reading format | 1 |
| NotYet: reading fromComponent | 1 |
| NotYet: reading fromNameType | 1 |
| NotYet: reading functionLocation | 1 |
| NotYet: reading functionType | 1 |
| NotYet: reading generatedLine | 1 |
| NotYet: reading generatedName | 1 |
| NotYet: reading generatorFunc | 1 |
| NotYet: reading generatorInstantiation | 1 |
| NotYet: reading genericDiag | 1 |
| NotYet: reading get | 1 |
| NotYet: reading getAccessorType | 1 |
| NotYet: reading getFromDirectoryCache | 1 |
| NotYet: reading getFromNonRelativeNameCache | 1 |
| NotYet: reading getFunc | 1 |
| NotYet: reading getMapOfCacheRedirects | 1 |
| NotYet: reading getPackageJsonInfo | 1 |
| NotYet: reading getParenthesizeLeftSideOfBinaryForOperator | 1 |
| NotYet: reading getSourcePosition | 1 |
| NotYet: reading getUnscopedHelperName | 1 |
| NotYet: reading getUnusedExpectations | 1 |
| NotYet: reading globalCache | 1 |
| NotYet: reading hasEmptyObject | 1 |
| NotYet: reading hasExistingReasonToReportErrorOn | 1 |
| NotYet: reading hasInstanceProperty | 1 |
| NotYet: reading hasJSDocFunctionType | 1 |
| NotYet: reading hasLeadingModifier | 1 |
| NotYet: reading hasPrivateModifier | 1 |
| NotYet: reading hasSignatures | 1 |
| NotYet: reading hasTrailingDecorator | 1 |
| NotYet: reading hasTransformableStatics | 1 |
| NotYet: reading headerPadding | 1 |
| NotYet: reading helper | 1 |
| NotYet: reading helpers | 1 |
| NotYet: reading hostSourceFileInfo | 1 |
| NotYet: reading i | 1 |
| NotYet: reading ids | 1 |
| NotYet: reading ifStatement | 1 |
| NotYet: reading iife | 1 |
| NotYet: reading illegalContextMessage | 1 |
| NotYet: reading immediate | 1 |
| NotYet: reading immediateDeclaration | 1 |
| NotYet: reading implDecl | 1 |
| NotYet: reading implementationSharesContainerWithFirstOverload | 1 |
| NotYet: reading impliedNodeFormat | 1 |
| NotYet: reading importClause | 1 |
| NotYet: reading importDecl | 1 |
| NotYet: reading importSource | 1 |
| NotYet: reading imports | 1 |
| NotYet: reading includeFileRegexes | 1 |
| NotYet: reading includeRe | 1 |
| NotYet: reading indent | 1 |
| NotYet: reading indexedAccessType | 1 |
| NotYet: reading inferences | 1 |
| NotYet: reading initialLocationForSecondaryLookup | 1 |
| NotYet: reading initialType | 1 |
| NotYet: reading initializerWithoutParens | 1 |
| NotYet: reading initializersName | 1 |
| NotYet: reading inlinable | 1 |
| NotYet: reading innerIndexType | 1 |
| NotYet: reading innerMappedType | 1 |
| NotYet: reading instantiatedSignature | 1 |
| NotYet: reading instantiatedTemplateType | 1 |
| NotYet: reading instantiations | 1 |
| NotYet: reading internalFlags | 1 |
| NotYet: reading intrinsicAttribs | 1 |
| NotYet: reading intrinsicElementsType | 1 |
| NotYet: reading intrinsics | 1 |
| NotYet: reading invalidElement | 1 |
| NotYet: reading isAbstract | 1 |
| NotYet: reading isAnonymous | 1 |
| NotYet: reading isArrowFunctionInJsx | 1 |
| NotYet: reading isCallToReadHelper | 1 |
| NotYet: reading isCallbackTag | 1 |
| NotYet: reading isCapturedInFunction | 1 |
| NotYet: reading isClassWithConstructorReference | 1 |
| NotYet: reading isComparingJsxAttributes | 1 |
| NotYet: reading isConfigIdentical | 1 |
| NotYet: reading isDerivedClass | 1 |
| NotYet: reading isDosStyle | 1 |
| NotYet: reading isEitherEnum | 1 |
| NotYet: reading isExportEquals | 1 |
| NotYet: reading isIllegalExportDefaultInCJS | 1 |
| NotYet: reading isInExternalModule | 1 |
| NotYet: reading isJSDoc | 1 |
| NotYet: reading isJavaScript | 1 |
| NotYet: reading isKnownProperty | 1 |
| NotYet: reading isLengthPushOrUnshift | 1 |
| NotYet: reading isLongestMatchingPrefix | 1 |
| NotYet: reading isMarkdownOrJSDocLink | 1 |
| NotYet: reading isNameFirst | 1 |
| NotYet: reading isOptional | 1 |
| NotYet: reading isOverload | 1 |
| NotYet: reading isParameter | 1 |
| NotYet: reading isPlainJs | 1 |
| NotYet: reading isPromise | 1 |
| NotYet: reading isPropertyName | 1 |
| NotYet: reading isReservedWord | 1 |
| NotYet: reading isRest | 1 |
| NotYet: reading isSimpleLoop | 1 |
| NotYet: reading isStaticMethodSymbol | 1 |
| NotYet: reading isTypeOnly | 1 |
| NotYet: reading isValid | 1 |
| NotYet: reading isValue | 1 |
| NotYet: reading issuedDiagnostic | 1 |
| NotYet: reading iterator | 1 |
| NotYet: reading iteratorValueStatement | 1 |
| NotYet: reading jsDoc | 1 |
| NotYet: reading jsDocType | 1 |
| NotYet: reading jsxChildrenPropertyName | 1 |
| NotYet: reading jsxSpecific | 1 |
| NotYet: reading keyAttr | 1 |
| NotYet: reading laneCount | 1 |
| NotYet: reading lastChild | 1 |
| NotYet: reading lastElement | 1 |
| NotYet: reading lastJSDocParam | 1 |
| NotYet: reading lastPart | 1 |
| NotYet: reading lastSpan | 1 |
| NotYet: reading lastStatement | 1 |
| NotYet: reading leadingComments | 1 |
| NotYet: reading leadingLineTerminatorCount | 1 |
| NotYet: reading leftColumnHeadingLength | 1 |
| NotYet: reading leftIsNumeric | 1 |
| NotYet: reading leftTarget | 1 |
| NotYet: reading leftType | 1 |
| NotYet: reading leftmost | 1 |
| NotYet: reading lex | 1 |
| NotYet: reading limitedConstraint | 1 |
| NotYet: reading linesAfterDot | 1 |
| NotYet: reading literal | 1 |
| NotYet: reading literalValue | 1 |
| NotYet: reading literals | 1 |
| NotYet: reading localCheckDeclaration | 1 |
| NotYet: reading localIndexDeclaration | 1 |
| NotYet: reading localName | 1 |
| NotYet: reading localSymbol | 1 |
| NotYet: reading location | 1 |
| NotYet: reading mainExport | 1 |
| NotYet: reading mappedTypeNode | 1 |
| NotYet: reading mappedTypeVariable | 1 |
| NotYet: reading mapper | 1 |
| NotYet: reading mappings | 1 |
| NotYet: reading matchType | 1 |
| NotYet: reading maxErrors | 1 |
| NotYet: reading maxLength | 1 |
| NotYet: reading maxNonRestParam | 1 |
| NotYet: reading mayHaveNameCollisions | 1 |
| NotYet: reading memberProps | 1 |
| NotYet: reading metadataReference | 1 |
| NotYet: reading methodReturnType | 1 |
| NotYet: reading missingNode | 1 |
| NotYet: reading missingPaths | 1 |
| NotYet: reading mixinFlags | 1 |
| NotYet: reading mod | 1 |
| NotYet: reading modifierArray | 1 |
| NotYet: reading moduleBlock | 1 |
| NotYet: reading moduleStatement | 1 |
| NotYet: reading nameParts | 1 |
| NotYet: reading nameStr | 1 |
| NotYet: reading nameText | 1 |
| NotYet: reading namedBindings | 1 |
| NotYet: reading namespaceDeclaration | 1 |
| NotYet: reading narrowedType | 1 |
| NotYet: reading needCheckInitializer | 1 |
| NotYet: reading needJsExtensions | 1 |
| NotYet: reading needSyncEval | 1 |
| NotYet: reading needsModifierPreservingWrapper | 1 |
| NotYet: reading needsName | 1 |
| NotYet: reading needsOutParam | 1 |
| NotYet: reading needsUpdateInTypeRootWatch | 1 |
| NotYet: reading negative | 1 |
| NotYet: reading newElements | 1 |
| NotYet: reading newEndN | 1 |
| NotYet: reading newItem | 1 |
| NotYet: reading newName | 1 |
| NotYet: reading newParam | 1 |
| NotYet: reading newParams | 1 |
| NotYet: reading newSourceFile | 1 |
| NotYet: reading nodeConstructors | 1 |
| NotYet: reading nodeContextFlags | 1 |
| NotYet: reading nodeInAmbientContext | 1 |
| NotYet: reading nodeModulesDirectoryName | 1 |
| NotYet: reading nodeModulesFolder | 1 |
| NotYet: reading nodeModulesFolderExists | 1 |
| NotYet: reading nodeName | 1 |
| NotYet: reading nodes | 1 |
| NotYet: reading normalizedElements | 1 |
| NotYet: reading ns | 1 |
| NotYet: reading nullishSemantics | 1 |
| NotYet: reading numNodes | 1 |
| NotYet: reading numParameters | 1 |
| NotYet: reading numericValue | 1 |
| NotYet: reading objectLiterals | 1 |
| NotYet: reading objectProperties | 1 |
| NotYet: reading offset | 1 |
| NotYet: reading ok | 1 |
| NotYet: reading oldEnclosing | 1 |
| NotYet: reading oldEndN | 1 |
| NotYet: reading oldSignature | 1 |
| NotYet: reading oldSourceFiles | 1 |
| NotYet: reading oldStart2 | 1 |
| NotYet: reading oldTime | 1 |
| NotYet: reading onProgramCreateComplete | 1 |
| NotYet: reading onWatchStatusChange | 1 |
| NotYet: reading openBracketPosition | 1 |
| NotYet: reading operand | 1 |
| NotYet: reading optional | 1 |
| NotYet: reading optionalDeclaration | 1 |
| NotYet: reading optionsOfCurCategory | 1 |
| NotYet: reading optionsType | 1 |
| NotYet: reading originalClass | 1 |
| NotYet: reading originalCreateDirectory | 1 |
| NotYet: reading originalFile | 1 |
| NotYet: reading originalModuleSpecifier | 1 |
| NotYet: reading originalReadFile | 1 |
| NotYet: reading other | 1 |
| NotYet: reading otherFiles | 1 |
| NotYet: reading otherOption | 1 |
| NotYet: reading outPath | 1 |
| NotYet: reading outerTypeParameters | 1 |
| NotYet: reading outputDir | 1 |
| NotYet: reading overloadSignatures | 1 |
| NotYet: reading ownKey | 1 |
| NotYet: reading ownKeys | 1 |
| NotYet: reading ownMap | 1 |
| NotYet: reading ownOutputFilePath | 1 |
| NotYet: reading packageFileResult | 1 |
| NotYet: reading packageJsonMap | 1 |
| NotYet: reading packageName | 1 |
| NotYet: reading packageResult | 1 |
| NotYet: reading paramIdent | 1 |
| NotYet: reading parameterDeclaration | 1 |
| NotYet: reading parameterIndex | 1 |
| NotYet: reading parameterNode | 1 |
| NotYet: reading parameterRange | 1 |
| NotYet: reading parameterSymbol | 1 |
| NotYet: reading parentDeclaration | 1 |
| NotYet: reading parenthesizerRules | 1 |
| NotYet: reading pathAndExtension | 1 |
| NotYet: reading pathList | 1 |
| NotYet: reading pathToTopLevelNodeModules | 1 |
| NotYet: reading pattern | 1 |
| NotYet: reading peerDependencies | 1 |
| NotYet: reading pendingExpressions | 1 |
| NotYet: reading pendingKind | 1 |
| NotYet: reading possibleOption | 1 |
| NotYet: reading possibleOutOfBounds | 1 |
| NotYet: reading possiblyOutOfBoundsType | 1 |
| NotYet: reading precedingLineBreak | 1 |
| NotYet: reading prerelease | 1 |
| NotYet: reading prereleaseArray | 1 |
| NotYet: reading prevNodeIndex | 1 |
| NotYet: reading previousDuration | 1 |
| NotYet: reading primaryDeclaration | 1 |
| NotYet: reading primitive | 1 |
| NotYet: reading printNode | 1 |
| NotYet: reading printerOptions | 1 |
| NotYet: reading priority | 1 |
| NotYet: reading programDiagnosticsInFile | 1 |
| NotYet: reading prologueStatementCount | 1 |
| NotYet: reading promise | 1 |
| NotYet: reading propContext | 1 |
| NotYet: reading propNameType | 1 |
| NotYet: reading propNode | 1 |
| NotYet: reading propertyAccess | 1 |
| NotYet: reading propertyAssignmentType | 1 |
| NotYet: reading propertyTypes | 1 |
| NotYet: reading proto | 1 |
| NotYet: reading prototype | 1 |
| NotYet: reading prototypePropertyType | 1 |
| NotYet: reading prototypeSymbol | 1 |
| NotYet: reading prototypeType | 1 |
| NotYet: reading qualifiedName | 1 |
| NotYet: reading quick | 1 |
| NotYet: reading quickResult | 1 |
| NotYet: reading rawName | 1 |
| NotYet: reading rawSources | 1 |
| NotYet: reading react | 1 |
| NotYet: reading reactExports | 1 |
| NotYet: reading readExpression | 1 |
| NotYet: reading readFileWithCache | 1 |
| NotYet: reading readonlyMask | 1 |
| NotYet: reading realDeclarationPath | 1 |
| NotYet: reading recursionIdentity | 1 |
| NotYet: reading reducedType | 1 |
| NotYet: reading reexports | 1 |
| NotYet: reading referencedFileName | 1 |
| NotYet: reading referencedMap | 1 |
| NotYet: reading relativePath | 1 |
| NotYet: reading relevantTypeParameterConstraints | 1 |
| NotYet: reading removeNullable | 1 |
| NotYet: reading removeUndefined | 1 |
| NotYet: reading rename | 1 |
| NotYet: reading repopulatedChain | 1 |
| NotYet: reading reportErrors | 1 |
| NotYet: reading requiresAddingUndefined | 1 |
| NotYet: reading resolutionDiagnostic | 1 |
| NotYet: reading resolutionMode | 1 |
| NotYet: reading resolutionsChanged | 1 |
| NotYet: reading resolvedFromFile | 1 |
| NotYet: reading resolvedMethodReturnType | 1 |
| NotYet: reading resolvedModuleNames | 1 |
| NotYet: reading resolvedModuleSymbol | 1 |
| NotYet: reading resolvedValueSymbol | 1 |
| NotYet: reading resolver | 1 |
| NotYet: reading rest | 1 |
| NotYet: reading restParameterSymbols | 1 |
| NotYet: reading restSymbol | 1 |
| NotYet: reading resultFromDts | 1 |
| NotYet: reading resultType | 1 |
| NotYet: reading results | 1 |
| NotYet: reading returnOrPromisedType | 1 |
| NotYet: reading returnStatement | 1 |
| NotYet: reading returnType | 1 |
| NotYet: reading right | 1 |
| NotYet: reading rightCustomPrologueEnd | 1 |
| NotYet: reading rightHoistedFunctionsEnd | 1 |
| NotYet: reading rightHoistedVariablesEnd | 1 |
| NotYet: reading rightIdentifier | 1 |
| NotYet: reading rightStandardPrologueEnd | 1 |
| NotYet: reading rootExpr | 1 |
| NotYet: reading rootPathComponents | 1 |
| NotYet: reading rootResult | 1 |
| NotYet: reading rootSymbol | 1 |
| NotYet: reading runtimeImportSpecifier | 1 |
| NotYet: reading savedInStrictMode | 1 |
| NotYet: reading savedPreserveSourceNewlines | 1 |
| NotYet: reading scanner | 1 |
| NotYet: reading secondType | 1 |
| NotYet: reading seenNames | 1 |
| NotYet: reading segment | 1 |
| NotYet: reading segments | 1 |
| NotYet: reading separateBeginAndEnd | 1 |
| NotYet: reading serializeTypeOfDeclaration | 1 |
| NotYet: reading serializedName | 1 |
| NotYet: reading setFunc | 1 |
| NotYet: reading setProp | 1 |
| NotYet: reading setReadFileCache | 1 |
| NotYet: reading setterModifiers | 1 |
| NotYet: reading setterType | 1 |
| NotYet: reading shebang | 1 |
| NotYet: reading shouldConvertCondition | 1 |
| NotYet: reading shouldEmitComments | 1 |
| NotYet: reading shouldEmitDetachedComment | 1 |
| NotYet: reading shouldEmitDotDot | 1 |
| NotYet: reading shouldEmitSourceMaps | 1 |
| NotYet: reading shouldResolveAlias | 1 |
| NotYet: reading shouldResolveFactoryReference | 1 |
| NotYet: reading shouldTransformInitializers | 1 |
| NotYet: reading shouldTransformInitializersUsingSet | 1 |
| NotYet: reading shouldTransformThisInStaticInitializers | 1 |
| NotYet: reading shouldWriteNativeEvents | 1 |
| NotYet: reading sigRestType | 1 |
| NotYet: reading signatureNextType | 1 |
| NotYet: reading signatures | 1 |
| NotYet: reading singleLine | 1 |
| NotYet: reading singleQuote | 1 |
| NotYet: reading skipBindingPatterns | 1 |
| NotYet: reading snippetElement | 1 |
| NotYet: reading sortedIndex | 1 |
| NotYet: reading sourceConstraint | 1 |
| NotYet: reading sourceDiscriminantTypes | 1 |
| NotYet: reading sourceEnd | 1 |
| NotYet: reading sourceEndText | 1 |
| NotYet: reading sourceFileAbsolutePaths | 1 |
| NotYet: reading sourceFileNoExtension | 1 |
| NotYet: reading sourceFiles | 1 |
| NotYet: reading sourceHasBase | 1 |
| NotYet: reading sourceHasMoreParameters | 1 |
| NotYet: reading sourceInfos | 1 |
| NotYet: reading sourceIsJSConstructor | 1 |
| NotYet: reading sourceMapRange | 1 |
| NotYet: reading sourceMapUrlPos | 1 |
| NotYet: reading sourceMappingURL | 1 |
| NotYet: reading sourceProp | 1 |
| NotYet: reading sourceProperty | 1 |
| NotYet: reading sourceSignature | 1 |
| NotYet: reading sourceSignatures | 1 |
| NotYet: reading sourceStartText | 1 |
| NotYet: reading space | 1 |
| NotYet: reading spacesToEmit | 1 |
| NotYet: reading spec | 1 |
| NotYet: reading specialPropertyAssignmentKind | 1 |
| NotYet: reading specifierType | 1 |
| NotYet: reading specifiers | 1 |
| NotYet: reading spreadElement | 1 |
| NotYet: reading spreadIndex | 1 |
| NotYet: reading startPos | 1 |
| NotYet: reading startsOnNewLine | 1 |
| NotYet: reading stat | 1 |
| NotYet: reading statementExpression | 1 |
| NotYet: reading statementOffset | 1 |
| NotYet: reading staticBlock | 1 |
| NotYet: reading strName | 1 |
| NotYet: reading subsequentNode | 1 |
| NotYet: reading substituteConstraints | 1 |
| NotYet: reading superCall | 1 |
| NotYet: reading superPath | 1 |
| NotYet: reading superStatement | 1 |
| NotYet: reading symbolExport | 1 |
| NotYet: reading symbols | 1 |
| NotYet: reading syntacticBuilderResolver | 1 |
| NotYet: reading syntheticArgsSymbol | 1 |
| NotYet: reading system | 1 |
| NotYet: reading tagExpression | 1 |
| NotYet: reading tags | 1 |
| NotYet: reading targetDeclarationKind | 1 |
| NotYet: reading targetDepth | 1 |
| NotYet: reading targetFile | 1 |
| NotYet: reading targetHasBase | 1 |
| NotYet: reading targetIsJSConstructor | 1 |
| NotYet: reading targetIsOptional | 1 |
| NotYet: reading targetParam | 1 |
| NotYet: reading targetReturn | 1 |
| NotYet: reading targetStartText | 1 |
| NotYet: reading targetSymbolFile | 1 |
| NotYet: reading targetType | 1 |
| NotYet: reading tempVar | 1 |
| NotYet: reading templates | 1 |
| NotYet: reading thenFunction | 1 |
| NotYet: reading thisParam | 1 |
| NotYet: reading thisParameter | 1 |
| NotYet: reading thisParameters | 1 |
| NotYet: reading throwDiagnostic | 1 |
| NotYet: reading timerToInvalidateFailedLookupResolutions | 1 |
| NotYet: reading timerToUpdateProgram | 1 |
| NotYet: reading tok | 1 |
| NotYet: reading tokenSourceMapRanges | 1 |
| NotYet: reading tokenString | 1 |
| NotYet: reading tokenText | 1 |
| NotYet: reading tp | 1 |
| NotYet: reading trackSymbol | 1 |
| NotYet: reading trailingComments | 1 |
| NotYet: reading trailingNewlines | 1 |
| NotYet: reading trailingParts | 1 |
| NotYet: reading trampoline | 1 |
| NotYet: reading transform | 1 |
| NotYet: reading tripleSlash | 1 |
| NotYet: reading trueCondition | 1 |
| NotYet: reading tryStatement | 1 |
| NotYet: reading tsPriority | 1 |
| NotYet: reading tsconfigTime | 1 |
| NotYet: reading tupleTarget | 1 |
| NotYet: reading tupleType | 1 |
| NotYet: reading typeAlias | 1 |
| NotYet: reading typeArgumentTypes | 1 |
| NotYet: reading typeIsAutomatic | 1 |
| NotYet: reading typeKey | 1 |
| NotYet: reading typeKind | 1 |
| NotYet: reading typeOfArrayLiteral | 1 |
| NotYet: reading typeOfObjectLiteral | 1 |
| NotYet: reading typeOnlyDeclarationIsExportStar | 1 |
| NotYet: reading typeOrConstraint | 1 |
| NotYet: reading typePredicateVariable | 1 |
| NotYet: reading typeReferenceResolutionsChanged | 1 |
| NotYet: reading typeRoots | 1 |
| NotYet: reading typeVariable | 1 |
| NotYet: reading typescriptVersion | 1 |
| NotYet: reading undefinedStrippedTarget | 1 |
| NotYet: reading unionType | 1 |
| NotYet: reading uniqueFilled | 1 |
| NotYet: reading unwidenedType | 1 |
| NotYet: reading unwrappedExprType | 1 |
| NotYet: reading updatedText | 1 |
| NotYet: reading usageMode | 1 |
| NotYet: reading useCaseSensitiveFileNames | 1 |
| NotYet: reading useStrictDirective | 1 |
| NotYet: reading val | 1 |
| NotYet: reading valid | 1 |
| NotYet: reading validatedFilesSpec | 1 |
| NotYet: reading validatedFilesSpecBeforeSubstitution | 1 |
| NotYet: reading valueParam | 1 |
| NotYet: reading values | 1 |
| NotYet: reading variableDeclarator | 1 |
| NotYet: reading variableList | 1 |
| NotYet: reading variance | 1 |
| NotYet: reading varianceFlags | 1 |
| NotYet: reading version | 1 |
| NotYet: reading visibilityResult | 1 |
| NotYet: reading visibleDefaultBinding | 1 |
| NotYet: reading visitedNode | 1 |
| NotYet: reading voidIsNonOptional | 1 |
| NotYet: reading watchDirectoryKind | 1 |
| NotYet: reading watchFileKind | 1 |
| NotYet: reading write | 1 |
| NotYet: reading writeText | 1 |
| NotYet: spreading an array of other elements | 1 |
| NotYet: storing any in a field | 1 |
| NotYet: storing false \| Type in a field | 1 |
| NotYet: storing string \| number in a field | 1 |
| Refused: &&= | 1 |
| Refused: JSON.parse: its result's type can't be proven from the text | 1 |
| Refused: Object.defineProperty | 1 |
| Refused: Object.setPrototypeOf | 1 |
| Refused: a constructor object escaping before static fields are initialized | 1 |
| Refused: a function taking (symbol: Symbol) => boolean seen as one taking ((symbol: Symbol) => boolean) \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking BinaryOperatorToken seen as one taking BinaryOperatorToken \| BinaryOperator (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking JSDocTypeExpression \| undefined seen as one taking JSDocTypeExpression \| JSDocTypeLiteral \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking LogLevel seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking Node seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking Node \| undefined seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking NodeArray<Expression> seen as one taking readonly Expression[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking NodeArray<T> seen as one taking NodeArray<T> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking NodeArray<T> seen as one taking NodeArray<T> \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking NodeArray<TypeNode> \| undefined seen as one taking readonly TypeNode[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking OuterExpressionKinds seen as one taking OuterExpressionKinds \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking SourceFile seen as one taking [file: SourceFile, options: CompilerOptions] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking T seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking TIn seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking TokenFlags seen as one taking TokenFlags \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking TypeFacts.None seen as one taking number (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking TypeNode seen as one taking Node (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [callback: (...args: any[]) => void, ms: number, ...args: any[]] seen as one taking (...args: any[]) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [fileName: string] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [node: ConstructorTypeNode, typeParameters: NodeArray<TypeParameterDeclaration> \| undefined, parameters: NodeArray<ParameterDeclaration>, type: TypeNode] \| ... seen as one taking ConstructorTypeNode (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [node: Node] seen as one taking BindingElement \| OmittedExpression (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [node: Node] seen as one taking Declaration (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [path: string, callback: DirectoryWatcherCallback, recursive?: boolean \| undefined, options?: WatchOptions \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [path: string, callback: FileWatcherCallback, pollingInterval?: number \| undefined, options?: WatchOptions \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [timeoutId: any] seen as one taking unknown (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [typeParameters: readonly TypeParameterDeclaration[] \| undefined, parameters: readonly ParameterDeclaration[], type: TypeNode] \| [modifiers: readonly Modifier[] \| undefined, typeParameters: ... \| undefined, parameters: ..., type: TypeNode] seen as one taking readonly Modifier[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [writeByteOrderMark: boolean, onError?: ((message: string) => void) \| undefined, sourceFiles?: readonly SourceFile[] \| undefined, data?: WriteFileCallbackData \| undefined] seen as one taking boolean (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking never seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking number seen as one taking any[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking number seen as one taking number \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking readonly ParameterDeclaration[] seen as one taking readonly ParameterDeclaration[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking readonly T[] seen as one taking readonly T[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking readonly T[] \| undefined seen as one taking readonly T[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking string seen as one taking string \| MemberName (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a method read as a value (add would lose its object, and this with it) | 1 |
| Refused: a method read as a value (base64decode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (base64encode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (clearScreen would lose its object, and this with it) | 1 |
| Refused: a method read as a value (cloneNode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (compare would lose its object, and this with it) | 1 |
| Refused: a method read as a value (convertToArrayAssignmentElement would lose its object, and this with it) | 1 |
| Refused: a method read as a value (convertToObjectAssignmentElement would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createComma would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createExpressionStatement would lose its object, and this with it) | 1 |
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
| Refused: a method read as a value (createTemplateMiddle would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createTemplateTail would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createUnionTypeNode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (deleteFile would lose its object, and this with it) | 1 |
| Refused: a method read as a value (emit would lose its object, and this with it) | 1 |
| Refused: a method read as a value (emitBuildInfo would lose its object, and this with it) | 1 |
| Refused: a method read as a value (emitNextAffectedFile would lose its object, and this with it) | 1 |
| Refused: a method read as a value (fill would lose its object, and this with it) | 1 |
| Refused: a method read as a value (fromCharCode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getAllDependencies would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getDeclarationDiagnostics would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getMemoryUsage would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getModifiedTime would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getNearestAncestorDirectoryWithPackageJson would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getOrCreateCacheForModuleName would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getSemanticDiagnostics would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getSourceFileByPath would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getSymlinkCache would lose its object, and this with it) | 1 |
| Refused: a method read as a value (globalCacheResolutionModuleName would lose its object, and this with it) | 1 |
| Refused: a method read as a value (hasChangedEmitSignature would lose its object, and this with it) | 1 |
| Refused: a method read as a value (hasInvalidatedLibResolutions would lose its object, and this with it) | 1 |
| Refused: a method read as a value (hasInvalidatedResolutions would lose its object, and this with it) | 1 |
| Refused: a method read as a value (hasOwnProperty would lose its object, and this with it) | 1 |
| Refused: a method read as a value (log would lose its object, and this with it) | 1 |
| Refused: a method read as a value (nonEscapingWrite would lose its object, and this with it) | 1 |
| Refused: a method read as a value (onDiscoveredSymlink would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeCheckTypeOfConditionalType would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeConciseBodyOfArrowFunction would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeConditionOfConditionalExpression would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeConstituentTypeOfIntersectionType would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeConstituentTypeOfUnionType would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeElementTypeOfTupleType would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeExpressionOfComputedPropertyName would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeExpressionOfExportDefault would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeExpressionOfExpressionStatement would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeExpressionOfNew would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeExtendsTypeOfConditionalType would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeLeadingTypeArgument would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeOperandOfPostfixUnary would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeOperandOfReadonlyTypeOperator would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeOperandOfTypeOperator would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeTypeOfOptionalType would lose its object, and this with it) | 1 |
| Refused: a method read as a value (releaseProgram would lose its object, and this with it) | 1 |
| Refused: a method read as a value (remove would lose its object, and this with it) | 1 |
| Refused: a method read as a value (repeat would lose its object, and this with it) | 1 |
| Refused: a method read as a value (replace would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportCyclicStructureError would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportInaccessibleThisError would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportInferenceFallback would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportNonSerializableProperty would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportNonlocalAugmentation would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportTruncationError would lose its object, and this with it) | 1 |
| Refused: a method read as a value (resolveModuleNameLiterals would lose its object, and this with it) | 1 |
| Refused: a method read as a value (resolveModuleNames would lose its object, and this with it) | 1 |
| Refused: a method read as a value (setBlocking would lose its object, and this with it) | 1 |
| Refused: a method read as a value (setModifiedTime would lose its object, and this with it) | 1 |
| Refused: a method read as a value (throwIfCancellationRequested would lose its object, and this with it) | 1 |
| Refused: a method read as a value (toString would lose its object, and this with it) | 1 |
| Refused: a method read as a value (tryEnableSourceMapsForHost would lose its object, and this with it) | 1 |
| Refused: a method read as a value (writeOutputIsTTY would lose its object, and this with it) | 1 |
| Refused: a spread after the first field | 1 |
| Refused: a structural Object.keys view that can hide an iterable literal's symbol-key storage (adamic/symbol-key-view) | 1 |
| Refused: a type argument makes a value of type (left: MappedPosition, right: MappedPosition) => boolean seen as EqualityComparer<SourceMappedPosition> \| undefined, which can write string \| undefined where string is read | 1 |
| Refused: a type predicate whose return is not proven (asserts cond needs a boolean parameter) | 1 |
| Refused: a type predicate whose return is not proven (false return can still contain LateBoundDeclaration) | 1 |
| Refused: a type predicate whose return is not proven (normal return has not narrowed value to NonNullable<T>) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on arg) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on array) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on buildOrder) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on diagnostic) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on file) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on func) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on hostSourceFile) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on location) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on m) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on option) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on options) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on p) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on program) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on sourceFile) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on symbol) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on t) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on tagName) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on x) | 1 |
| Refused: a type predicate whose return is not proven (the body's true narrowing does not match null \| undefined) | 1 |
| Refused: a type predicate whose return is not proven (the predicate parameter is assigned) | 1 |
| Refused: a type predicate whose return is not proven (true return narrows to MappedPosition, not SourceMappedPosition) | 1 |
| Refused: a type predicate whose return is not proven (true return narrows to Mapping, not SourceMapping) | 1 |
| Refused: a type predicate whose return is not proven (true return narrows to ReusableBuilderProgramState, not BuilderProgramStateWithDefinedProgram) | 1 |
| Refused: a value of type ((node: PrivateIdentifierPropertyDeclaration, modifiers: ModifiersArray \| undefined) => ObjectLiteralExpression) \| undefined seen as ((node: PropertyDeclaration & { readonly name: PrivateIdentifier; }, modifiers: ModifiersArray \| undefined) => Expression) \| undefined, whose readonly field name becomes writable: a readonly field may hold something narrower than PrivateIdentifier, which a write of PrivateIdentifier would replace | 1 |
| Refused: a value of type () => { diagnosticMessage: DiagnosticMessage; errorNode: ExportAssignment; } seen as GetSymbolAccessibilityDiagnostic, which can write Node where ExportAssignment is read | 1 |
| Refused: a value of type (AbstractKeyword \| AccessorKeyword \| AsyncKeyword \| ConstKeyword \| DeclareKeyword \| Decorator \| ... 7 more ... \| StaticKeyword)[] \| undefined seen as Node[] \| undefined, which can write Node where AbstractKeyword \| AccessorKeyword \| AsyncKeyword \| ConstKeyword \| DeclareKeyword \| Decorator \| ... 7 more ... \| StaticKeyword is read | 1 |
| Refused: a value of type (ClassDeclaration \| EnumDeclaration \| ExportAssignment \| ExportDeclaration \| FunctionDeclaration \| ... 5 more ... \| VariableStatement)[] seen as Statement[], which can write Statement where ClassDeclaration \| EnumDeclaration \| ExportAssignment \| ExportDeclaration \| FunctionDeclaration \| ... 5 more ... \| VariableStatement is read | 1 |
| Refused: a value of type (left: MappedPosition, right: MappedPosition) => boolean seen as EqualityComparer<SourceMappedPosition> \| undefined, which can write string \| undefined where string is read | 1 |
| Refused: a value of type (node: CommentRange) => boolean seen as (value: SynthesizedComment) => boolean, which can write number where -1 is read | 1 |
| Refused: a value of type (node: PrivateIdentifierGetAccessorDeclaration, modifiers: ModifiersArray \| undefined) => ObjectLiteralExpression seen as ((node: GetAccessorDeclaration & { readonly name: PrivateIdentifier; }, modifiers: ModifiersArray \| undefined) => Expression) \| undefined, whose readonly field name becomes writable: a readonly field may hold something narrower than PrivateIdentifier, which a write of PrivateIdentifier would replace | 1 |
| Refused: a value of type (node: PrivateIdentifierMethodDeclaration, modifiers: ModifiersArray \| undefined) => ObjectLiteralExpression seen as ((node: MethodDeclaration & { readonly name: PrivateIdentifier; }, modifiers: ModifiersArray \| undefined) => Expression) \| undefined, whose readonly field name becomes writable: a readonly field may hold something narrower than PrivateIdentifier, which a write of PrivateIdentifier would replace | 1 |
| Refused: a value of type (node: PrivateIdentifierSetAccessorDeclaration, modifiers: ModifiersArray \| undefined) => ObjectLiteralExpression seen as ((node: SetAccessorDeclaration & { readonly name: PrivateIdentifier; }, modifiers: ModifiersArray \| undefined) => Expression) \| undefined, whose readonly field name becomes writable: a readonly field may hold something narrower than PrivateIdentifier, which a write of PrivateIdentifier would replace | 1 |
| Refused: a value of type (recordTempVariable: ((node: Identifier) => void) \| undefined, reservedInNestedScopes?: boolean \| undefined, prefix?: string \| GeneratedNamePart \| undefined, suffix?: string \| undefined) => GeneratedIdentifier seen as { (recordTempVariable: ((node: Identifier) => void) \| undefined, reservedInNestedScopes?: boolean \| undefined): Identifier; (recordTempVariable: ((node: Identifier) => void) \| undefined, reservedInNestedScopes?: boolean \| undefined, prefix?: string \| ... 1 more ... \| undefined, suffix?: string \| undefined): Identifi..., whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type (resolution: ResolvedModuleWithFailedLookupLocations) => ResolvedModuleFull \| undefined seen as (oldResolution: ResolvedModuleWithFailedLookupLocations) => ResolutionWithResolvedFileName \| undefined, which can write string \| undefined where string is read | 1 |
| Refused: a value of type (symbol: Symbol) => { visitedTypes: Type[]; visitedSymbols: Symbol[]; } seen as (root: Symbol) => { visitedTypes: readonly Type[]; visitedSymbols: readonly Symbol[]; }, which can write readonly Type[] where Type[] is read | 1 |
| Refused: a value of type (symbolAccessibilityResult: SymbolAccessibilityResult) => { diagnosticMessage: DiagnosticMessage; errorNode: DeclarationDiagnosticProducing; typeName: DeclarationName \| undefined; } \| undefined seen as (symbolAccessibilityResult: SymbolAccessibilityResult) => SymbolAccessibilityDiagnostic \| undefined, which can write Node where DeclarationDiagnosticProducing is read | 1 |
| Refused: a value of type (type: Type) => { visitedTypes: Type[]; visitedSymbols: Symbol[]; } seen as (root: Type) => { visitedTypes: readonly Type[]; visitedSymbols: readonly Symbol[]; }, which can write readonly Type[] where Type[] is read | 1 |
| Refused: a value of type AnonymousType seen as AnonymousType, which can write AnonymousType \| undefined where GenericType is read | 1 |
| Refused: a value of type AnonymousType \| DeferredTypeReference seen as AnonymousType, which can write AnonymousType \| undefined where GenericType is read | 1 |
| Refused: a value of type AssertClause seen as Mutable<AssertClause>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type AssertEntry seen as Mutable<AssertEntry>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type AutoAccessorPropertyDeclaration \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 |
| Refused: a value of type AwaitedTypeInstantiation \| Type seen as Type, which can write Symbol \| undefined where Symbol is read | 1 |
| Refused: a value of type BigIntLiteral \| ComputedPropertyName \| GeneratedIdentifier \| NoSubstitutionTemplateLiteral \| NumericLiteral \| PrivateIdentifier \| StringLiteral seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type BigIntLiteral \| ComputedPropertyName \| GeneratedIdentifier \| NoSubstitutionTemplateLiteral \| NumericLiteral \| StringLiteral seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type BigIntLiteral \| ComputedPropertyName \| Identifier \| JsxNamespacedName \| NoSubstitutionTemplateLiteral \| NumericLiteral \| PrivateIdentifier \| StringLiteral seen as Type, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 |
| Refused: a value of type BigIntLiteral \| ComputedPropertyName \| Identifier \| JsxNamespacedName \| NoSubstitutionTemplateLiteral \| NumericLiteral \| PrivateIdentifier \| StringLiteral \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 |
| Refused: a value of type BigIntLiteral \| ComputedPropertyName \| Identifier \| NoSubstitutionTemplateLiteral \| NumericLiteral \| PrivateIdentifier \| StringLiteral \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type BinaryExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type BindingName seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Block seen as Mutable<Block>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Block \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 |
| Refused: a value of type BreakStatement seen as Mutable<BreakStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type CallExpression \| undefined seen as NodeArray<Expression> \| undefined, whose readonly field transformFlags becomes writable: a readonly field may hold something narrower than TransformFlags, which a write of TransformFlags would replace | 1 |
| Refused: a value of type CallSignatureDeclaration seen as Mutable<CallSignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type CapturedThis seen as PrimaryExpression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type CapturedThis seen as string \| BindingName, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type CaseBlock seen as Mutable<CaseBlock>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Children seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ClassDeclaration seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ClassDeclaration \| ClassExpression \| InferTypeNode \| InterfaceDeclaration \| JSDocCallbackTag \| ... 4 more ... \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type ClassDeclaration \| FunctionDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type CompilerHost seen as CompilerHostLikeForCache, which can write WriteFileCallback \| undefined where WriteFileCallback is read | 1 |
| Refused: a value of type ConstructSignatureDeclaration seen as Mutable<ConstructSignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ConstructorTypeNode seen as Mutable<ConstructorTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ContinueStatement seen as Mutable<ContinueStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Declaration seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type Declaration \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 |
| Refused: a value of type DestructuringAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type DiagnosticMessageChain \| { messageText: string; category: DiagnosticCategory; code: number; repopulateInfo?: () => RepopulateDiagnosticChainInfo; canonicalHead?: CanonicalDiagnostic; next: ... \| undefined; } seen as ReusableDiagnosticMessageChain, which can write ReusableDiagnosticMessageChain[] \| undefined where DiagnosticMessageChain[] \| undefined is read | 1 |
| Refused: a value of type DiagnosticMessageChain[] seen as ReusableDiagnosticMessageChain[], which can write ReusableDiagnosticMessageChain where DiagnosticMessageChain is read | 1 |
| Refused: a value of type DiagnosticMessageChain[] seen as ReusableDiagnosticMessageChain[], which can write ReusableDiagnosticMessageChain[] \| undefined where DiagnosticMessageChain[] \| undefined is read | 1 |
| Refused: a value of type DiagnosticWithLocation seen as Diagnostic \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 |
| Refused: a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 |
| Refused: a value of type DiagnosticWithLocation[] seen as readonly Diagnostic[] \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 |
| Refused: a value of type DiagnosticWithLocation[] seen as readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 1 |
| Refused: a value of type DiagnosticWithLocation[] \| undefined seen as readonly Diagnostic[] \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 |
| Refused: a value of type EntityNameExpression \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type EnumDeclaration \| ModuleDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type EqualsGreaterThanToken seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ExportAssignment seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ExportDeclaration \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type Expression \| GeneratedIdentifier seen as Expression \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type Expression \| GeneratedIdentifier seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type Expression \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 |
| Refused: a value of type ExpressionStatement seen as Mutable<ExpressionStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Extension[] seen as string[], which can write string where Extension is read | 1 |
| Refused: a value of type FlowArrayMutation \| FlowAssignment \| FlowCall \| FlowCondition \| FlowLabel \| FlowReduceLabel \| FlowStart \| FlowUnreachable seen as FlowNode, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 1 |
| Refused: a value of type FunctionTypeNode seen as Mutable<FunctionTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type GeneratedIdentifier seen as Expression \| GeneratedIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier seen as GeneratedIdentifier \| Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier seen as Node \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as GeneratedIdentifier \| GeneratedPrivateIdentifier \| Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as Identifier \| PrivateIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as Node \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier \| Identifier \| PrivateIdentifier seen as Identifier \| PrivateIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| Identifier seen as GeneratedIdentifier \| Identifier \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| Identifier seen as Identifier \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| Identifier seen as string \| BindingName, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| Identifier seen as string \| GeneratedIdentifier \| Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| Identifier \| undefined seen as string \| ModuleExportName \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedPrivateIdentifier seen as Node \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GetAccessorDeclaration \| MethodDeclaration \| PropertyAssignment \| SetAccessorDeclaration \| ShorthandPropertyAssignment \| SpreadAssignment \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 |
| Refused: a value of type GetAccessorDeclaration \| SetAccessorDeclaration seen as Mutable<AccessorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Identifier & GeneratedIdentifier & { readonly escapedText: { __escapedIdentifier: void; } & "__this"; } seen as BindingName, which can write EmitNode \| undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 1 |
| Refused: a value of type Identifier seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Identifier \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type Identifier[] seen as ModuleExportName[] \| undefined, which can write ModuleExportName where Identifier is read | 1 |
| Refused: a value of type ImportAttribute seen as Mutable<ImportAttribute>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ImportAttributes seen as Mutable<ImportAttributes>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ImportClause seen as Mutable<ImportClause>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ImportDeclaration seen as Mutable<ImportDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ImportDeclaration seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ImportEqualsDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ImportTypeNode seen as Mutable<ImportTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type IndexInfo \| undefined seen as IndexSignatureDeclaration \| undefined, which can write number \| undefined where number is read | 1 |
| Refused: a value of type IndexSignatureDeclaration seen as Mutable<IndexSignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type InitializedVariableDeclaration seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type IntrinsicType[] seen as TypeParameter[], which can write TypeParameter where IntrinsicType is read | 1 |
| Refused: a value of type IntrinsicType[] \| undefined seen as TypeParameter[] \| undefined, which can write TypeParameter where IntrinsicType is read | 1 |
| Refused: a value of type IterationTypes[] seen as (IterationTypes \| undefined)[], which can write IterationTypes \| undefined where IterationTypes is read | 1 |
| Refused: a value of type JSDocAugmentsTag seen as Mutable<JSDocAugmentsTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocCallbackTag seen as Mutable<JSDocCallbackTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocFunctionType seen as Mutable<JSDocFunctionType>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocImplementsTag seen as Mutable<JSDocImplementsTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocImportTag seen as Mutable<JSDocImportTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocLink seen as Mutable<JSDocLink>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocLinkCode seen as Mutable<JSDocLinkCode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocLinkPlain seen as Mutable<JSDocLinkPlain>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocNameReference seen as Mutable<JSDocNameReference>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocOverloadTag seen as Mutable<JSDocOverloadTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocParameterTag seen as Mutable<JSDocParameterTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocPropertyTag seen as Mutable<JSDocPropertyTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocSeeTag seen as Mutable<JSDocSeeTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocSignature seen as Mutable<JSDocSignature>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocSignature \| SignatureDeclaration \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type JSDocTemplateTag seen as Mutable<JSDocTemplateTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocText seen as Mutable<JSDocText>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocTypeExpression seen as Mutable<JSDocTypeExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocTypeExpression \| undefined seen as Signature \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SignatureFlags would replace | 1 |
| Refused: a value of type JSDocTypeLiteral seen as Mutable<JSDocTypeLiteral>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocTypedefTag seen as Mutable<JSDocTypedefTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocUnknownTag seen as Mutable<JSDocUnknownTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Map<Path, Diagnostic[]> seen as Map<Path, readonly Diagnostic[]> \| undefined, which can write readonly Diagnostic[] where Diagnostic[] is read | 1 |
| Refused: a value of type Map<Path, DirectoryWatchesOfFailedLookup> seen as Map<string, DirectoryWatchesOfFailedLookup>, which can write string where Path is read | 1 |
| Refused: a value of type Map<Path, FileWatcher> seen as Map<string, FileWatcher>, which can write string where Path is read | 1 |
| Refused: a value of type Map<Path, ModeAwareCache<CachedResolvedModuleWithFailedLookupLocations>> seen as Map<string, ModeAwareCache<CachedResolvedModuleWithFailedLookupLocations>>, which can write string where Path is read | 1 |
| Refused: a value of type Map<Path, ModeAwareCache<CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations>> seen as Map<string, ModeAwareCache<CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations>>, which can write string where Path is read | 1 |
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
| Refused: a value of type Map<string, CachedResolvedModuleWithFailedLookupLocations> seen as Map<string, ResolutionWithFailedLookupLocations> \| Set<ResolutionWithFailedLookupLocations> \| undefined, which can write ResolutionWithFailedLookupLocations where CachedResolvedModuleWithFailedLookupLocations is read | 1 |
| Refused: a value of type Map<string, ImportsNotUsedAsValues> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ImportsNotUsedAsValues is read | 1 |
| Refused: a value of type Map<string, JsxEmit> seen as Map<string, string \| number>, which can write string \| number where JsxEmit is read | 1 |
| Refused: a value of type Map<string, Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>> seen as ScriptTargetFeatures, which can write string where never is read | 1 |
| Refused: a value of type Map<string, ModuleDetectionKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ModuleDetectionKind is read | 1 |
| Refused: a value of type Map<string, ModuleResolutionKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ModuleResolutionKind is read | 1 |
| Refused: a value of type Map<string, NewLineKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where NewLineKind is read | 1 |
| Refused: a value of type Map<string, PollingWatchKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where PollingWatchKind is read | 1 |
| Refused: a value of type Map<string, WatchDirectoryKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where WatchDirectoryKind is read | 1 |
| Refused: a value of type Map<string, WatchFileKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where WatchFileKind is read | 1 |
| Refused: a value of type Map<string, [VariableDeclarationList, VariableDeclaration[]]> seen as Map<string, [CatchClause \| VariableDeclarationList, VariableDeclaration[]]>, which can write [CatchClause \| VariableDeclarationList, VariableDeclaration[]] where [VariableDeclarationList, VariableDeclaration[]] is read | 1 |
| Refused: a value of type Map<string, string> seen as Map<string, string \| number>, which can write string \| number where string is read | 1 |
| Refused: a value of type MappedTypeNode seen as Mutable<MappedTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type MemberName \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type MethodDeclaration seen as Mutable<MethodDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type MethodDeclaration \| PropertyAssignment \| AccessorDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ModuleDeclaration seen as Mutable<ModuleDeclaration \| SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ModuleExportName seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ModuleName seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ModuleName seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ModuleResolutionState & { host: GetPackageJsonEntrypointsHost; } seen as ModuleResolutionState, which can write ModuleResolutionHost where ModuleResolutionHost & GetPackageJsonEntrypointsHost is read | 1 |
| Refused: a value of type Mutable<NoSubstitutionTemplateLiteral> seen as Mutable<TemplateLiteralLikeNode>, which can write SyntaxKind where SyntaxKind.NoSubstitutionTemplateLiteral is read | 1 |
| Refused: a value of type Mutable<T> seen as T, a type parameter whose constraint JSDocType & { readonly type: TypeNode \| undefined; readonly postfix: boolean; } can be written, so it can write what Mutable<T> can't hold | 1 |
| Refused: a value of type Mutable<T> seen as T, a type parameter whose constraint JSDocType & { readonly type: TypeNode \| undefined; } can be written, so it can write what Mutable<T> can't hold | 1 |
| Refused: a value of type Mutable<T> seen as T, a type parameter whose constraint Node can be written, so it can write what Mutable<T> can't hold | 1 |
| Refused: a value of type NamedImports seen as Mutable<NamedImports>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type NamespaceExport seen as Mutable<NamespaceExport>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type NamespaceExportDeclaration seen as Mutable<NamespaceExportDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type NamespaceImport seen as Mutable<NamespaceImport>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Node seen as Node \| SourceMapRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Node seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Node \| SourceMapRange seen as SourceMapRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Node \| undefined seen as EmitNode \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of EmitFlags would replace | 1 |
| Refused: a value of type Node \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type NonNullExpression seen as Mutable<NonNullExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type NumberLiteralType seen as LiteralType, which can write string \| number \| PseudoBigInt where number is read | 1 |
| Refused: a value of type ObjectBindingOrAssignmentPattern seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type OptionalTypeNode seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ParameterDeclaration seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ParameterDeclaration \| PropertyDeclaration \| PropertySignature \| SignatureDeclaration seen as Mutable<ParameterDeclaration \| PropertyDeclaration \| PropertySignature \| SignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ParameterDeclaration \| VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ParseConfigHost seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where boolean is read | 1 |
| Refused: a value of type Path[] seen as string[], which can write string where Path is read | 1 |
| Refused: a value of type PropertyAccessExpression \| SyntheticSuper seen as LeftHandSideExpression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type PropertyAccessExpression \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type PropertyAssignment seen as Mutable<PropertyAssignment \| ShorthandPropertyAssignment>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type PropertyDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type PropertyName \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type Readonly<BuilderState> \| undefined seen as BuilderState \| undefined, whose readonly field fileInfos becomes writable: a readonly field may hold something narrower than Map<Path, FileInfo>, which a write of Map<Path, FileInfo> would replace | 1 |
| Refused: a value of type ReferencedFile & { kind: FileIncludeKind.LibReferenceDirective; } seen as ReferencedFile, which can write ReferencedFileKind where FileIncludeKind.LibReferenceDirective is read | 1 |
| Refused: a value of type ReportFileInError[] seen as (ReportFileInError \| undefined)[], which can write ReportFileInError \| undefined where ReportFileInError is read | 1 |
| Refused: a value of type ResolvedModuleFull \| undefined seen as { path: string; originalPath: string \| true; extension: string; packageId: PackageId \| undefined; resolvedUsingTsExtension: boolean \| undefined; } \| undefined, whose readonly field originalPath becomes writable: a readonly field may hold something narrower than string \| undefined, which a write of string \| true would replace | 1 |
| Refused: a value of type ResolvedType \| TypeReference seen as ObjectType, which can write SymbolTable \| undefined where SymbolTable is read | 1 |
| Refused: a value of type ReturnStatement seen as Mutable<ReturnStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type SearchResult<Resolved> seen as { value: { resolved: Resolved; isExternalLibraryImport: true; } \| undefined; } \| undefined, whose readonly field value becomes writable: a readonly field may hold something narrower than Resolved \| undefined, which a write of { resolved: Resolved; isExternalLibraryImport: true; } \| undefined would replace | 1 |
| Refused: a value of type Set<Path> \| undefined seen as Set<string> \| undefined, which can write string where Path is read | 1 |
| Refused: a value of type Set<__String> seen as Set<__String \| undefined>, which can write __String \| undefined where __String is read | 1 |
| Refused: a value of type SetAccessorDeclaration seen as Mutable<SetAccessorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ShorthandPropertyAssignment seen as Mutable<PropertyAssignment \| ShorthandPropertyAssignment>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type SourceFile seen as Mutable<ModuleDeclaration \| SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type SourceFile seen as SourceFileLike \| undefined, which can write readonly number[] \| undefined where readonly number[] is read | 1 |
| Refused: a value of type SourceFile \| undefined seen as FileReasonToChainCache \| undefined, which can write DiagnosticMessageChain[] \| undefined where RedirectInfo \| undefined is read | 1 |
| Refused: a value of type SourceFile \| undefined seen as NodeLinks \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of NodeCheckFlags would replace | 1 |
| Refused: a value of type SourceFile \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type SourceMapSource seen as SourceFileLike, which can write readonly number[] \| undefined where readonly number[] is read | 1 |
| Refused: a value of type Statement[] seen as Node[], which can write Node where Statement is read | 1 |
| Refused: a value of type Statement[] \| undefined seen as CaseClause[] \| undefined, which can write CaseClause where Statement is read | 1 |
| Refused: a value of type StringLiteral seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type StringLiteral seen as Mutable<StringLiteral>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type StringLiteralType seen as LiteralType, which can write string \| number \| PseudoBigInt where string is read | 1 |
| Refused: a value of type StringLiteralType[] seen as Type[], which can write Type where StringLiteralType is read | 1 |
| Refused: a value of type SuperExpression seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type SuperExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type SwitchStatement seen as Mutable<SwitchStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Symbol seen as Type \| undefined, which can write TypeFlags where SymbolFlags is read | 1 |
| Refused: a value of type Symbol seen as Type, which can write TypeFlags where SymbolFlags is read | 1 |
| Refused: a value of type Symbol \| undefined seen as GenericType \| undefined, which can write TypeFlags where SymbolFlags is read | 1 |
| Refused: a value of type Symbol \| undefined seen as Node \| undefined, which can write number \| undefined where number is read | 1 |
| Refused: a value of type Symbol \| undefined seen as SourceFile \| undefined, which can write number \| undefined where number is read | 1 |
| Refused: a value of type Symbol[] seen as (Symbol \| undefined)[], which can write Symbol \| undefined where Symbol is read | 1 |
| Refused: a value of type Symbol[] seen as Symbol[], which can write Symbol where never is read | 1 |
| Refused: a value of type SyntheticSuper seen as string \| BindingName, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type T seen as T, a type parameter whose constraint ClassDeclaration \| ClassExpression \| GetAccessorDeclaration \| MethodDeclaration \| ParameterDeclaration \| PropertyDeclaration \| SetAccessorDeclaration can be written, so it can write what T can't hold | 1 |
| Refused: a value of type T seen as T, a type parameter whose constraint EntityNameOrEntityNameExpression can be written, so it can write what T can't hold | 1 |
| Refused: a value of type T seen as T, a type parameter whose constraint HasModifiers can be written, so it can write what T can't hold | 1 |
| Refused: a value of type T seen as T, a type parameter whose constraint MethodDeclaration \| MethodSignature \| PropertyAssignment \| PropertyDeclaration \| PropertySignature \| AccessorDeclaration can be written, so it can write what T can't hold | 1 |
| Refused: a value of type T seen as T, a type parameter whose constraint ModifierSyntaxKind can be written, so it can write what T can't hold | 1 |
| Refused: a value of type T seen as T, a type parameter whose constraint Node \| undefined can be written, so it can write what T can't hold | 1 |
| Refused: a value of type T seen as T, a type parameter whose constraint ResolutionWithFailedLookupLocations can be written, so it can write what T can't hold | 1 |
| Refused: a value of type TKind seen as TKind, a type parameter whose constraint KeywordTypeSyntaxKind can be written, so it can write what TKind can't hold | 1 |
| Refused: a value of type T[] seen as (T \| undefined)[], which can write T \| undefined where T is read | 1 |
| Refused: a value of type T[][] seen as (readonly T[])[], which can write readonly T[] where T[] is read | 1 |
| Refused: a value of type TaggedTemplateExpression seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ThisCapturingVariableDeclaration seen as VariableDeclaration, which can write EmitNode \| undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 1 |
| Refused: a value of type ThrowStatement seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type TransientSymbol[] seen as Symbol[], which can write Symbol where TransientSymbol is read | 1 |
| Refused: a value of type Type seen as TypeNode, which can write number \| undefined where number is read | 1 |
| Refused: a value of type Type \| undefined seen as Expression \| undefined, which can write number \| undefined where number is read | 1 |
| Refused: a value of type Type \| undefined seen as JSDocTypeExpression \| undefined, which can write number \| undefined where number is read | 1 |
| Refused: a value of type Type \| undefined seen as Signature \| undefined, which can write SignatureFlags where TypeFlags is read | 1 |
| Refused: a value of type TypeNode seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type TypeOperatorNode seen as Mutable<TypeOperatorNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type TypeParameterDeclaration seen as Mutable<TypeParameterDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type TypeParameterDeclaration \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type TypeVariable[] seen as (Type \| undefined)[], which can write Type \| undefined where TypeVariable is read | 1 |
| Refused: a value of type Type[] seen as (Type \| undefined)[], which can write Type \| undefined where Type is read | 1 |
| Refused: a value of type VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type VariableDeclaration \| DestructuringAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type VariableDeclarationList seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type VariableDeclaration[] seen as VariableStatement[], which can write VariableStatement where VariableDeclaration is read | 1 |
| Refused: a value of type VariableStatement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read | 1 |
| Refused: a value of type YieldExpression seen as Mutable<YieldExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type [() => Type, __String][] seen as (readonly [() => Type, __String])[], which can write readonly [() => Type, __String] where [() => Type, __String] is read | 1 |
| Refused: a value of type never seen as T, a type parameter whose constraint Node can be written, so it can write what never can't hold | 1 |
| Refused: a value of type never seen as T, a type parameter whose constraint any can be written, so it can write what never can't hold | 1 |
| Refused: a value of type never seen as U, a type parameter whose constraint {} can be written, so it can write what never can't hold | 1 |
| Refused: a value of type never[] seen as (JSDoc \| JSDocTag)[], which can write JSDoc \| JSDocTag where never is read | 1 |
| Refused: a value of type never[] seen as (JSDocCallbackTag \| JSDocEnumTag \| JSDocTypedefTag)[], which can write JSDocCallbackTag \| JSDocEnumTag \| JSDocTypedefTag where never is read | 1 |
| Refused: a value of type never[] seen as (SourceMapRange \| undefined)[], which can write SourceMapRange \| undefined where never is read | 1 |
| Refused: a value of type never[] seen as (readonly (Identifier \| StringLiteral)[])[], which can write readonly (Identifier \| StringLiteral)[] where never is read | 1 |
| Refused: a value of type never[] seen as BindingElement[], which can write BindingElement where never is read | 1 |
| Refused: a value of type never[] seen as CommentRange[], which can write CommentRange where never is read | 1 |
| Refused: a value of type never[] seen as Comparator[][], which can write Comparator[] where never is read | 1 |
| Refused: a value of type never[] seen as Expression[], which can write Expression where never is read | 1 |
| Refused: a value of type never[] seen as FlowNode[], which can write FlowNode where never is read | 1 |
| Refused: a value of type never[] seen as Identifier[], which can write Identifier where never is read | 1 |
| Refused: a value of type never[] seen as IntrinsicType[], which can write IntrinsicType where never is read | 1 |
| Refused: a value of type never[] seen as JSDocImportTag[], which can write JSDocImportTag where never is read | 1 |
| Refused: a value of type never[] seen as JsxAttributes[], which can write JsxAttributes where never is read | 1 |
| Refused: a value of type never[] seen as ParameterDeclaration[], which can write ParameterDeclaration where never is read | 1 |
| Refused: a value of type never[] seen as Path[], which can write Path where never is read | 1 |
| Refused: a value of type never[] seen as PotentiallyUnusedIdentifier[], which can write PotentiallyUnusedIdentifier where never is read | 1 |
| Refused: a value of type never[] seen as ProjectReference[], which can write ProjectReference where never is read | 1 |
| Refused: a value of type never[] seen as PropertyAssignment[], which can write PropertyAssignment where never is read | 1 |
| Refused: a value of type never[] seen as RequireOrImportCall[], which can write RequireOrImportCall where never is read | 1 |
| Refused: a value of type never[] seen as ResolvedProjectReference[], which can write ResolvedProjectReference where never is read | 1 |
| Refused: a value of type never[] seen as SourceMappedPosition[], which can write SourceMappedPosition where never is read | 1 |
| Refused: a value of type never[] seen as SymbolTable[], which can write SymbolTable where never is read | 1 |
| Refused: a value of type never[] seen as Symbol[] \| undefined, which can write Symbol where never is read | 1 |
| Refused: a value of type never[] seen as TransformerFactory<Bundle \| SourceFile>[], which can write TransformerFactory<Bundle \| SourceFile> where never is read | 1 |
| Refused: a value of type never[] seen as VarianceFlags[], which can write VarianceFlags where never is read | 1 |
| Refused: a value of type never[] \| SortedArray<DiagnosticWithLocation> seen as Diagnostic[], which can write Diagnostic where never is read | 1 |
| Refused: a value of type never[][] seen as string[][], which can write string where never is read | 1 |
| Refused: a value of type number[] seen as string \| (string \| number)[] \| undefined, which can write string \| number where number is read | 1 |
| Refused: a value of type readonly TypeParameter[] \| undefined seen as TypeParameterDeclaration[] \| undefined, which can write TypeParameterDeclaration where TypeParameter is read | 1 |
| Refused: a value of type readonly string[] \| undefined seen as RegExp[] \| undefined, which can write RegExp where string is read | 1 |
| Refused: a value of type string \| GeneratedIdentifier \| Identifier seen as string \| ModuleExportName, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined seen as string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined, which can write string \| number where string is read | 1 |
| Refused: a value of type string \| number \| boolean \| string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> seen as CompilerOptionsValue, which can write string \| number where string is read | 1 |
| Refused: a value of type string[] seen as DiagnosticArguments, which can write string \| number \| boolean \| readonly string[] \| SourceFile \| undefined where string is read | 1 |
| Refused: a value of type string[] seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what string[] can't hold | 1 |
| Refused: a value of type string[] seen as string \| (string \| number)[] \| undefined, which can write string \| number where string is read | 1 |
| Refused: a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"CheckMode", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof CheckMode is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"EmitFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof EmitFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"GeneratedIdentifierFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof GeneratedIdentifierFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"ModifierFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ModifierFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"NodeCheckFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof NodeCheckFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"NodeFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof NodeFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"ObjectFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ObjectFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"RelationComparisonResult", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof RelationComparisonResult is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"ScriptKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ScriptKind is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"SignatureCheckMode", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SignatureCheckMode is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"SignatureFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SignatureFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"SnippetKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SnippetKind is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"SymbolFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SymbolFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"SyntaxKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SyntaxKind is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"TransformFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TransformFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"TypeFacts", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TypeFacts is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"TypeFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TypeFlags is read | 1 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what undefined can't hold | 1 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint Declaration can be written, so it can write what undefined can't hold | 1 |
| Refused: a value of type { (fileName: string): DiagnosticWithLocation[]; (): Diagnostic[]; } seen as { (): Diagnostic[]; (fileName: string): DiagnosticWithLocation[]; }, which can write Diagnostic where DiagnosticWithLocation is read | 1 |
| Refused: a value of type { affectedFile: SourceFile; emitKind: BuilderFileEmit.Js \| BuilderFileEmit.JsMap \| BuilderFileEmit.JsInlineMap \| BuilderFileEmit.DtsErrors \| ... 6 more ... \| BuilderFileEmit.All; } seen as { affectedFile: Program \| SourceFile \| undefined; emitKind: BuilderFileEmit; }, which can write Program \| SourceFile \| undefined where SourceFile is read | 1 |
| Refused: a value of type { all?: boolean; allowImportingTsExtensions?: boolean; allowJs?: boolean; allowNonTsExtensions?: boolean; allowArbitraryExtensions?: boolean; allowSyntheticDefaultImports?: boolean; allowUmdGlobalAccess?: boolean; ... 129 more ...; moduleResolution: ModuleResolutionKind; } seen as CompilerOptions, which can write ModuleResolutionKind \| undefined where ModuleResolutionKind is read | 1 |
| Refused: a value of type { arguments: { name?: string; } & { path: string; }; range: CommentRange; } \| { arguments: { name: string; }; range: CommentRange; } \| { arguments: { factory: string; }; range: CommentRange; } \| ... 5 more ... \| ... seen as ({ arguments: { name?: string; } & { path: string; }; range: CommentRange; } \| { arguments: { name: string; }; range: CommentRange; } \| { arguments: { factory: string; }; range: CommentRange; } \| ... 5 more ... \| ...)[] \| ... 8 more ... \| ..., which can write { name?: string; } & { path: string; } where never is read | 1 |
| Refused: a value of type { compilerOptions: CompilerOptions; traceEnabled: boolean; affectingLocations: string[] \| undefined; resultFromCache?: ResolvedModuleWithFailedLookupLocations; ... 9 more ...; host: GetPackageJsonEntrypointsHost; } seen as ModuleResolutionState & { host: GetPackageJsonEntrypointsHost; }, which can write string[] \| undefined where never[] is read | 1 |
| Refused: a value of type { ending: ModuleSpecifierEnding; value: string; }[] seen as { ending: ModuleSpecifierEnding \| undefined; value: string; }[], which can write ModuleSpecifierEnding \| undefined where ModuleSpecifierEnding is read | 1 |
| Refused: a value of type { host: ModuleResolutionHost; traceEnabled: boolean; failedLookupLocations: string[] \| undefined; affectingLocations: string[] \| undefined; resultFromCache?: ResolvedModuleWithFailedLookupLocations; ... 8 more ...; reportDiagnostic: ...; } seen as ModuleResolutionState, which can write CompilerOptions where { all?: boolean; allowImportingTsExtensions?: boolean; allowJs?: boolean; allowNonTsExtensions?: boolean; allowArbitraryExtensions?: boolean; allowSyntheticDefaultImports?: boolean; allowUmdGlobalAccess?: boolean; ... 129 more ...; moduleResolution: ModuleResolutionKind; } is read | 1 |
| Refused: a value of type { id: number; flowNode: FlowNode; edges: never[]; text: string; lane: number; endLane: number; level: number; circular: false; } seen as FlowGraphNode, which can write FlowGraphEdge[] where never[] is read | 1 |
| Refused: a value of type { introducesError: boolean; node: Identifier \| PropertyAccessEntityNameExpression; sym?: never; } \| { introducesError: boolean; node: Identifier \| PropertyAccessEntityNameExpression; sym: Symbol \| undefined; } seen as { introducesError: boolean; node: LeftHandSideExpression; }, which can write LeftHandSideExpression where Identifier \| PropertyAccessEntityNameExpression is read | 1 |
| Refused: a value of type { major: number; minor: number; patch: number; prerelease: string; build: string; } seen as { major: string \| number; minor: number; patch: number; prerelease: string \| readonly string[]; build: string \| readonly string[]; }, which can write string \| number where number is read | 1 |
| Refused: a value of type { name: string \| undefined; path: string; }[] seen as AmdDependency[], which can write AmdDependency where { name: string \| undefined; path: string; } is read | 1 |
| Refused: a value of type { noInferenceFallback?: boolean \| undefined; enclosingDeclaration: ModuleDeclaration; enclosingFile: SourceFile \| undefined; flags: NodeBuilderFlags; ... 28 more ...; out: WriterContextOut; } seen as NodeBuilderContext, which can write Node \| undefined where ModuleDeclaration is read | 1 |
| Refused: a value of type { referencedName: StringLiteral; name: PropertyName; } \| { referencedName: Identifier; name: ComputedPropertyName; } seen as { referencedName: Expression \| undefined; name: PropertyName \| undefined; }, which can write Expression \| undefined where StringLiteral is read | 1 |
| Refused: an arbitrary number or a value from another enum assigned to Extension.Mjs; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to ForegroundColorEscapeSequences.Grey; its members are a closed union | 1 |
| Refused: an unproven relation from DiagnosticRelatedInformation to Diagnostic: optional field reportsUnnecessary has no proven compatible presence/type | 1 |
| Refused: an unproven relation from Identifier to Identifier: optional field id has no proven compatible presence/type | 1 |
| Refused: an unproven relation from ImportEqualsDeclaration & { moduleReference: ExternalModuleReference; } to ImportEqualsDeclaration: optional field moduleReference.id has no proven compatible presence/type | 1 |
| Refused: an unproven relation from InstantiableType \| UnionOrIntersectionType to TypeParameter: optional field constraint has no proven compatible presence/type | 1 |
| Refused: an unproven relation from LiteralLikeNode to TemplateLiteralLikeNode: optional field rawText has no proven compatible presence/type | 1 |
| Refused: an unproven relation from NodeArray<BindingElement> \| NodeArray<Expression> \| NodeArray<ArrayBindingElement> to NodeArray<Node>: optional field concat.parameter.element.slice.element.propertyName has no proven compatible presence/type | 1 |
| Refused: an unproven relation from NodeArray<ModifierLike> & readonly Decorator[] to NodeArray<Decorator>: optional field concat.element.parent.name has no proven compatible presence/type | 1 |
| Refused: an unproven relation from PackageJsonPathFields to PackageJson: optional field version has no proven compatible presence/type | 1 |
| Refused: an unproven relation from SolutionBuilderHostBase<T> to SolutionBuilderHost<T>: optional field reportErrorSummary has no proven compatible presence/type | 1 |
| Refused: an unproven relation from StructuredType to AnonymousType: optional field target has no proven compatible presence/type | 1 |
| Refused: an unproven relation from UnionOrIntersectionType to UnionType: optional field resolvedReducedType has no proven compatible presence/type | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot BuildStep.Done | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot BuilderFileEmit.Js \| BuilderFileEmit.JsMap \| BuilderFileEmit.JsInlineMap \| BuilderFileEmit.DtsErrors \| BuilderFileEmit.DtsEmit \| ... 5 more ... \| BuilderFileEmit.All | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot BuilderFileEmit.None | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot CharacterCodes.plus | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot CharacterCodes.slash | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot Connection | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot EmitOnly.Dts | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot ModuleKind.ESNext | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot NodeFlags.Const | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot NodeFlags.NestedNamespace | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot PunctuationOrKeywordSyntaxKind | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.BindingElement | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Block | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ClassStaticBlockDeclaration | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ConditionalType | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.DotDotDotToken | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ExportAssignment | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ExtendsKeyword \| SyntaxKind.ImplementsKeyword | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ImportAttributes | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ImportDeclaration | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.KeyOfKeyword \| SyntaxKind.ReadonlyKeyword \| SyntaxKind.UniqueKeyword | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.MethodSignature \| SyntaxKind.MethodDeclaration \| SyntaxKind.Constructor \| SyntaxKind.GetAccessor \| SyntaxKind.SetAccessor \| SyntaxKind.CallSignature \| SyntaxKind.ConstructSignature \| ... 6 more ... \| SyntaxKind.JSDocFunctionType | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NamespaceExport | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NamespaceImport | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NumericLiteral \| SyntaxKind.BigIntLiteral \| SyntaxKind.StringLiteral \| SyntaxKind.JsxText \| SyntaxKind.JsxTextAllWhiteSpaces \| SyntaxKind.RegularExpressionLiteral \| SyntaxKind.NoSubstitutionTemplateLiteral | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ObjectLiteralExpression | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Parameter | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.QuestionToken | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.StringLiteral | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Unknown \| SyntaxKind.NumericLiteral \| SyntaxKind.BigIntLiteral \| SyntaxKind.StringLiteral \| SyntaxKind.JsxText \| SyntaxKind.RegularExpressionLiteral \| SyntaxKind.NoSubstitutionTemplateLiteral | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.VariableDeclaration | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot TempFlags | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot Ternary.False | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot VarianceFlags.Contravariant | 1 |
| Refused: arguments | 1 |
| Refused: inherited library member compare read as an own field | 1 |
| Refused: inherited library member hasOwnProperty read as an own field | 1 |
| Refused: inherited library member repeat read as an own field | 1 |
| Refused: inherited library member replace read as an own field | 1 |
| Refused: optional property Low in Partial<Levels> absent from structural source {}, which can hide fields | 1 |
| Refused: optional property _propertyAccessExpressionLikeQualifiedNameBrand in PropertyAccessEntityNameExpression absent from structural source never, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source OptionsBase, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "noImplicitAny" \| "strict">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "noImplicitThis" \| "strict">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" \| "strictBindCallApply">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" \| "strictBuiltinIteratorReturn">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" \| "strictFunctionTypes">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" \| "strictNullChecks">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" \| "strictPropertyInitialization">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" \| "useUnknownInCatchVariables">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source TypeAcquisition, which can hide fields | 1 |
| Refused: optional property constraint in TypeParameter absent from structural source ObjectType, which can hide fields | 1 |
| Refused: optional property createProgram in IncrementalProgramOptions<EmitAndSemanticDiagnosticsBuilderProgram> absent from structural source IncrementalCompilationOptions, which can hide fields | 1 |
| Refused: optional property default in TypeParameter absent from structural source IndexedAccessType, which can hide fields | 1 |
| Refused: optional property id in BinaryExpression absent from structural source never, which can hide fields | 1 |
| Refused: optional property id in ComputedPropertyName absent from structural source never, which can hide fields | 1 |
| Refused: optional property id in JSDocThisTag absent from structural source never, which can hide fields | 1 |
| Refused: optional property id in PrivateIdentifier absent from structural source never, which can hide fields | 1 |
| Refused: optional property isInvalidated in CachedResolvedModuleWithFailedLookupLocations absent from structural source ResolvedModuleWithFailedLookupLocations, which can hide fields | 1 |
| Refused: optional property jsDocCache in JSDocArray absent from structural source JSDoc[], which can hide fields | 1 |
| Refused: optional property localSymbol in Declaration absent from structural source never, which can hide fields | 1 |
| Refused: optional property members in ObjectType absent from structural source IntersectionType, which can hide fields | 1 |
| Refused: optional property members in ObjectType absent from structural source UnionType, which can hide fields | 1 |
| Refused: optional property modifiers in ExportAssignment absent from structural source never, which can hide fields | 1 |
| Refused: optional property modifiers in ParameterDeclaration absent from structural source never, which can hide fields | 1 |
| Refused: optional property modifiers in PropertyDeclaration absent from structural source never, which can hide fields | 1 |
| Refused: optional property modifiers in SetAccessorDeclaration absent from structural source never, which can hide fields | 1 |
| Refused: optional property outputDts in ResolvedRefAndOutputDts absent from structural source ResolvedRefAndSource, which can hide fields | 1 |
| Refused: optional property packageRootPath in { moduleFileToTry: string; packageRootPath?: string; blockedByExports?: true; verbatimFromExports?: true; } absent from structural source { moduleFileToTry: string; }, which can hide fields | 1 |
| Refused: optional property preserve in FileReference absent from structural source { resolutionMode: ModuleKind.CommonJS \| ModuleKind.ESNext; }, which can hide fields | 1 |
| Refused: optional property reportsUnnecessary in ReusableDiagnostic absent from structural source ReusableDiagnosticRelatedInformation, which can hide fields | 1 |
| Refused: optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<ImportAttribute>, which can hide fields | 1 |
| Refused: optional property return in ArrayIterator<WatchedFileWithUnchangedPolls \| undefined> absent from structural source ArrayIterator<WatchedFileWithUnchangedPolls>, which can hide fields | 1 |
| Refused: optional property return in MapIterator<[string, WatchDirectoryFlags]> absent from structural source MapIterator<[string, WatchDirectoryFlags]>, which can hide fields | 1 |
| Refused: optional property skipLogging in ErrorOutputContainer absent from structural source { errors?: Diagnostic[]; }, which can hide fields | 1 |
| Refused: optional property skipTrivia in SourceMapSource absent from structural source SourceFile, which can hide fields | 1 |
| Refused: optional property source in ResolvedRefAndSource absent from structural source ResolvedRefAndOutputDts, which can hide fields | 1 |
| Refused: optional property target in AnonymousType absent from structural source IntrinsicType, which can hide fields | 1 |
| Refused: optional property types in { types?: { value: string; pos: number; end: number; }; } & { lib?: { value: string; pos: number; end: number; }; } & { path?: { value: string; pos: number; end: number; }; } & { "no-default-lib"?: string; } & { "resolution-mode"?: string; } & ... absent from structural source { [index: string]: string \| undefined; }, which can hide fields | 1 |
| Refused: optional property types in { types?: { value: string; pos: number; end: number; }; } & { lib?: { value: string; pos: number; end: number; }; } & { path?: { value: string; pos: number; end: number; }; } & { "no-default-lib"?: string; } & { "resolution-mode"?: string; } & ... absent from structural source { [index: string]: string \| { value: string; pos: number; end: number; }; }, which can hide fields | 1 |
| Refused: optional property watchFile in WatchOptions absent from structural source {}, which can hide fields | 1 |
| error: statement panic: Unhandled case in Node.Text: *ast.ComputedPropertyName | 1 |
| error: statement panic: runtime error: invalid memory address or nil pointer dereference | 1 |
| panic: Unhandled case in Node.Text: *ast.ComputedPropertyName | 1 |
| panic: Unhandled case in Node.Text: *ast.QualifiedName | 1 |

## full / compiler

| Kind and exact reason | Sites |
| --- | ---: |
| Refused: a cast the runtime can't check | 3645 |
| NotYet: a function inside a function (a closure) | 3110 |
| Refused: an object refinement using an open numeric enum as a literal tag | 2776 |
| Refused: the non-null assertion ! | 1754 |
| NotYet: a method call through a structural signature in a program with statics; use typeof the declaring class | 1589 |
| NotYet: reading node | 1043 |
| Refused: a number as a condition | 647 |
| NotYet: a PrefixUnaryExpression on a value | 630 |
| NotYet: a NonNullExpression | 553 |
| NotYet: reading Debug | 509 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on node) | 414 |
| NotYet: a BinaryExpression with a value and a value | 410 |
| NotYet: a PrefixUnaryExpression on a number | 316 |
| NotYet: a function without a body | 303 |
| Refused: a value as a condition | 271 |
| NotYet: a BinaryExpression with a value and a boolean | 258 |
| NotYet: reading result | 227 |
| NotYet: a BinaryExpression with a number and a number | 200 |
| NotYet: a field of type boolean \| undefined | 164 |
| NotYet: a value of type __String | 152 |
| Refused: var | 140 |
| NotYet: a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions | 138 |
| NotYet: a function returning T | 118 |
| NotYet: a value of type Path | 113 |
| NotYet: assigning to an Identifier | 113 |
| NotYet: a BinaryExpression with a boolean and a value | 111 |
| NotYet: a call through ?. (an optional call) | 105 |
| Refused: \|\|= | 105 |
| Refused: a string as a condition | 101 |
| NotYet: for...of over an object | 98 |
| NotYet: reading type | 94 |
| NotYet: a function returning T \| undefined | 93 |
| NotYet: a value of type any | 91 |
| Refused: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 90 |
| Refused: a boolean \| undefined as a condition | 80 |
| NotYet: a BinaryExpression with a number and a boolean | 72 |
| NotYet: a parameter that isn't a plain name | 72 |
| Refused: a type predicate whose return is not proven (there is no body proving this parameter) | 71 |
| NotYet: a BinaryExpression with a boolean and a number | 62 |
| Refused: a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where SourceFile is read | 59 |
| NotYet: an array of never | 57 |
| NotYet: a declaration directly in a case (wrap the case in a block) | 56 |
| Refused: optional property id in Node absent from structural source never, which can hide fields | 55 |
| NotYet: a value of type T | 54 |
| NotYet: reading performance | 52 |
| NotYet: a PrefixUnaryExpression on a boolean \| undefined | 49 |
| Refused: the comma operator | 47 |
| NotYet: reading symbol | 46 |
| NotYet: reading name | 44 |
| NotYet: a BinaryExpression as a statement | 43 |
| NotYet: reading updated | 39 |
| NotYet: a BinaryExpression with a value and a number | 37 |
| Refused: a type predicate whose return is not proven (return paths through KindSwitchStatement are not verified) | 37 |
| NotYet: reading expression | 36 |
| NotYet: a PrefixUnaryExpression on a string | 35 |
| Refused: a value of type SourceFile seen as SourceFileLike, which can write readonly number[] \| undefined where readonly number[] is read | 34 |
| NotYet: reading host | 33 |
| NotYet: this outside a method | 33 |
| NotYet: a function value returning union of differently held members | 32 |
| NotYet: a call returning void \| undefined | 31 |
| NotYet: a void call used as a value | 31 |
| NotYet: reading statement | 31 |
| NotYet: an ElementAccessExpression | 27 |
| NotYet: reading declaration | 27 |
| NotYet: an array of T | 25 |
| NotYet: replacing a represented method at runtime | 25 |
| Refused: a value of type TupleTypeReference seen as TypeReference, which can write GenericType where TupleType is read | 25 |
| Refused: an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | 24 |
| NotYet: a field of type string \| NodeArray<JSDocComment> \| undefined | 23 |
| NotYet: reading clone | 23 |
| NotYet: reading compilerOptions | 23 |
| Refused: a method in object destructuring | 23 |
| Refused: optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<TypeParameterDeclaration>, which can hide fields | 23 |
| NotYet: a value of type __String \| undefined | 22 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on kind) | 22 |
| NotYet: a BinaryExpression with a number \| undefined and a number | 21 |
| NotYet: a BinaryExpression with a string and a string | 21 |
| NotYet: a value of type T \| undefined | 21 |
| NotYet: a value of type unknown | 21 |
| NotYet: new an Identifier | 21 |
| Refused: optional property source in SourceMapRange absent from structural source TextRange, which can hide fields | 21 |
| panic: statement panic: runtime error: invalid memory address or nil pointer dereference | 21 |
| NotYet: a PrefixUnaryExpression on a number \| undefined | 20 |
| NotYet: a value of type SolutionBuilderState<T> | 20 |
| Refused: a type predicate whose return is not proven (branch is not a trusted parameter check) | 20 |
| Refused: optional property return in ArrayIterator<Statement> absent from structural source ArrayIterator<JsonObjectExpressionStatement>, which can hide fields | 20 |
| NotYet: a call to a PropertyAccessExpression | 19 |
| NotYet: reading resolved | 19 |
| Refused: a value of type FlowNode seen as FlowNode \| undefined, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 19 |
| NotYet: a BinaryExpression with a boolean \| undefined and a boolean | 18 |
| NotYet: a function returning U \| undefined | 18 |
| NotYet: a function returning __String | 18 |
| NotYet: a generic function as a value | 18 |
| NotYet: a value of type ResolvedConfigFilePath | 18 |
| NotYet: reading sourceFile | 18 |
| Refused: a method read as a value (liftToBlock would lose its object, and this with it) | 18 |
| NotYet: a function returning __String \| undefined | 17 |
| NotYet: reading options | 17 |
| NotYet: reading visited | 17 |
| Refused: optional property id in ArrayBindingPattern absent from structural source never, which can hide fields | 17 |
| NotYet: a BinaryExpression with a number and a value | 16 |
| NotYet: reading candidate | 16 |
| Refused: a value of type ResolvedType seen as ObjectType, which can write SymbolTable \| undefined where SymbolTable is read | 16 |
| NotYet: a function returning Path | 15 |
| NotYet: a value of type object | 15 |
| NotYet: reading file | 15 |
| NotYet: reading sig | 15 |
| Refused: a number \| undefined as a condition | 15 |
| Refused: a value of type never[] seen as Diagnostic[], which can write Diagnostic where never is read | 15 |
| Refused: yield (generators) | 15 |
| NotYet: a Set of __String (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far) | 14 |
| NotYet: a function returning any | 14 |
| NotYet: an array of any | 14 |
| NotYet: reading block | 14 |
| NotYet: reading child | 14 |
| NotYet: reading parameter | 14 |
| Refused: a value of type Node seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 14 |
| Refused: a value of type never[] seen as BaseType[], which can write BaseType where never is read | 14 |
| Refused: a value of type { value: never; done: true; } seen as IteratorResult<Mapping, any>, which can write any where never is read | 14 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.JSDocTypeExpression | 14 |
| NotYet: a BinaryExpression with a boolean and a boolean \| undefined | 13 |
| NotYet: reading diag | 13 |
| NotYet: reading index | 13 |
| NotYet: reading related | 13 |
| Refused: a value of type never[] seen as string[], which can write string where never is read | 13 |
| Refused: the void operator | 13 |
| NotYet: a function with an optional or rest parameter, as a value | 12 |
| NotYet: a value of type ResolvedConfigFileName | 12 |
| NotYet: optional chaining to .size on a value | 12 |
| NotYet: reading t | 12 |
| Refused: a definite assignment assertion ! | 12 |
| Refused: a generator function | 12 |
| Refused: a method read as a value (parenthesizeExpressionForDisallowedComma would lose its object, and this with it) | 12 |
| Refused: a union of differently held members as a condition | 12 |
| Refused: a value of type Node[] seen as unknown[], which can write unknown where Node is read | 12 |
| Refused: a value of type never[] seen as Signature[], which can write Signature where never is read | 12 |
| Refused: a value of type never[] seen as Symbol[], which can write Symbol where never is read | 12 |
| Refused: a value without nominal ancestry seen as BinaryExpressionStateMachine<TOuterState, TState, TResult> | 12 |
| NotYet: a function returning undefined | 11 |
| NotYet: a value of type NonNullable<T> | 11 |
| NotYet: a value of type TKind | 11 |
| NotYet: reading cache | 11 |
| Refused: a namespace | 11 |
| Refused: a value of type NodeBuilderContext seen as SyntacticTypeNodeBuilderContext, which can write Required<Pick<SymbolTracker, "reportInferenceFallback">> where SymbolTrackerImpl is read | 11 |
| Refused: a value of type T seen as Mutable<T>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of any would replace | 11 |
| Refused: an index signature | 11 |
| Refused: in | 11 |
| NotYet: a field of type true \| Node \| undefined | 10 |
| NotYet: a function returning T[] | 10 |
| NotYet: an enum inside a function or block; declare it at module scope | 10 |
| NotYet: an optional chain longer than one step | 10 |
| NotYet: for...in without a proven fixed plain-object origin (arrays, prototypes and absent synthetic fields cannot be enumerated soundly) | 10 |
| NotYet: reading body | 10 |
| NotYet: reading decl | 10 |
| NotYet: reading exception | 10 |
| NotYet: reading flags | 10 |
| NotYet: reading reference | 10 |
| Refused: a value of type BuilderProgramStateWithDefinedProgram seen as BuilderProgramState, which can write Program \| undefined where Program is read | 10 |
| Refused: a value of type Map<string, never[]> seen as Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>, which can write string[] where never[] is read | 10 |
| Refused: a value of type Symbol \| undefined seen as Type \| undefined, which can write TypeFlags where SymbolFlags is read | 10 |
| Refused: optional property constraint in TypeParameter absent from structural source Type, which can hide fields | 10 |
| Refused: optional property id in BigIntLiteral absent from structural source never, which can hide fields | 10 |
| NotYet: a BinaryExpression with a number and a string | 9 |
| NotYet: a ModuleDeclaration | 9 |
| NotYet: a PrefixUnaryExpression on a union of differently held members | 9 |
| NotYet: a value of type T["kind"] | 9 |
| NotYet: reading Parser | 9 |
| NotYet: reading elem | 9 |
| NotYet: reading existing | 9 |
| NotYet: reading expr | 9 |
| NotYet: reading parsed | 9 |
| NotYet: reading restType | 9 |
| NotYet: storing boolean \| undefined in a field | 9 |
| Refused: a function taking Identifier seen as one taking Identifier \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 9 |
| Refused: a value of type BuilderProgramState seen as ReusableBuilderProgramState, which can write Map<Path, readonly Diagnostic[] \| readonly ReusableDiagnostic[]> where Map<Path, readonly Diagnostic[]> is read | 9 |
| Refused: a value of type DiagnosticWithLocation[] seen as Diagnostic[], which can write Diagnostic where DiagnosticWithLocation is read | 9 |
| Refused: a value of type Expression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 9 |
| Refused: a value of type never[] seen as IndexInfo[], which can write IndexInfo where never is read | 9 |
| NotYet: a BinaryExpression with a string and a number | 8 |
| NotYet: a field of type true \| undefined | 8 |
| NotYet: a value of type K | 8 |
| NotYet: a value of type readonly T[] | 8 |
| NotYet: a value of type readonly T[] \| undefined | 8 |
| NotYet: new a ParenthesizedExpression | 8 |
| NotYet: reading bundle | 8 |
| NotYet: reading diagnostics | 8 |
| NotYet: reading left | 8 |
| NotYet: reading links | 8 |
| NotYet: reading parent | 8 |
| NotYet: reading reduced | 8 |
| NotYet: reading specifier | 8 |
| NotYet: reading statements | 8 |
| NotYet: reading target | 8 |
| NotYet: reading temp | 8 |
| NotYet: reading text | 8 |
| Refused: a function taking boolean seen as one taking boolean \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 8 |
| Refused: a method read as a value (realpath would lose its object, and this with it) | 8 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on type) | 8 |
| Refused: a value of type never[] seen as Type[], which can write Type where never is read | 8 |
| NotYet: a BinaryExpression with a string and a boolean | 7 |
| NotYet: a Set of Path (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far) | 7 |
| NotYet: a field of type string \| DiagnosticMessageChain | 7 |
| NotYet: a value of type BindableStaticNameExpression | 7 |
| NotYet: assigning a field of a value | 7 |
| NotYet: reading array | 7 |
| NotYet: reading buildOrder | 7 |
| NotYet: reading context | 7 |
| NotYet: reading current | 7 |
| NotYet: reading element | 7 |
| NotYet: reading firstDecorator | 7 |
| NotYet: reading flowType | 7 |
| NotYet: reading initializer | 7 |
| NotYet: reading res | 7 |
| NotYet: reading scope | 7 |
| NotYet: reading source | 7 |
| NotYet: reading token | 7 |
| NotYet: reading typeNode | 7 |
| NotYet: regex replacement other than a string | 7 |
| Refused: a method read as a value (parenthesizeLeftSideOfAccess would lose its object, and this with it) | 7 |
| Refused: a value of type FlowNode seen as FlowNode[] \| FlowNode \| undefined, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 7 |
| Refused: a value of type GeneratedIdentifier \| Identifier seen as Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 7 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.SourceFile | 7 |
| Refused: optional property id in ExternalModuleReference absent from structural source never, which can hide fields | 7 |
| Refused: optional property id in Identifier absent from structural source never, which can hide fields | 7 |
| Refused: optional property members in ObjectType absent from structural source IntrinsicType, which can hide fields | 7 |
| NotYet: ?.[] on a value | 6 |
| NotYet: a BinaryExpression with a boolean and a string | 6 |
| NotYet: a BinaryExpression with a value and a string | 6 |
| NotYet: a function value returning boolean \| undefined | 6 |
| NotYet: a value of type HasJSDoc \| undefined | 6 |
| NotYet: a value of type TOuterState | 6 |
| NotYet: assigning to a NonNullExpression | 6 |
| NotYet: reading autoGenerate | 6 |
| NotYet: reading c | 6 |
| NotYet: reading cacheKey | 6 |
| NotYet: reading directoryWatcher | 6 |
| NotYet: reading errorNode | 6 |
| NotYet: reading includes | 6 |
| NotYet: reading isGenerator | 6 |
| NotYet: reading iterationTypes | 6 |
| NotYet: reading localOrExportSymbol | 6 |
| NotYet: reading map | 6 |
| NotYet: reading packageInfo | 6 |
| NotYet: reading params | 6 |
| NotYet: reading pos | 6 |
| NotYet: reading propertyName | 6 |
| NotYet: reading queue | 6 |
| NotYet: reading start | 6 |
| NotYet: reading typeParameters | 6 |
| NotYet: reading v | 6 |
| NotYet: reading value | 6 |
| NotYet: storing boolean in a field | 6 |
| Refused: a function taking BinaryExpressionStateMachine<TOuterState, TState, TResult> seen as one taking BinaryExpressionStateMachine<TOuterState, TState, TResult> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 6 |
| Refused: a method read as a value (fileExists would lose its object, and this with it) | 6 |
| Refused: a method read as a value (parenthesizeOperandOfPrefixUnary would lose its object, and this with it) | 6 |
| Refused: a method read as a value (readFile would lose its object, and this with it) | 6 |
| Refused: a parameter property | 6 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on n) | 6 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on token) | 6 |
| Refused: a value of type Declaration[] seen as Node[], which can write Node where Declaration is read | 6 |
| Refused: a value of type Expression seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 6 |
| Refused: a value of type FlowNode \| undefined seen as false \| FlowNode \| undefined, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 6 |
| Refused: a value of type Identifier seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 6 |
| Refused: a value of type NonNullable<T> seen as T, a type parameter whose constraint any can be written, so it can write what NonNullable<T> can't hold | 6 |
| Refused: a value of type TypeNode \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 6 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint Node can be written, so it can write what undefined can't hold | 6 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.TypeKeyword | 6 |
| Refused: optional property id in Expression absent from structural source never, which can hide fields | 6 |
| NotYet: a BinaryExpression with a boolean \| undefined and a number | 5 |
| NotYet: a BinaryExpression with a number and a boolean \| undefined | 5 |
| NotYet: a case whose type differs from the switch's | 5 |
| NotYet: a function returning Token<TKind> | 5 |
| NotYet: a function returning U | 5 |
| NotYet: a value of type string \| null \| undefined | 5 |
| NotYet: an array of U | 5 |
| NotYet: reading arg | 5 |
| NotYet: reading assignedName | 5 |
| NotYet: reading buildOptions | 5 |
| NotYet: reading classType | 5 |
| NotYet: reading configFile | 5 |
| NotYet: reading entityName | 5 |
| NotYet: reading exportSpecifiers | 5 |
| NotYet: reading extensions | 5 |
| NotYet: reading innerExpression | 5 |
| NotYet: reading jsFilePath | 5 |
| NotYet: reading merged | 5 |
| NotYet: reading objectFlags | 5 |
| NotYet: reading path | 5 |
| NotYet: reading propType | 5 |
| NotYet: reading thisType | 5 |
| NotYet: reading typeName | 5 |
| NotYet: reading types | 5 |
| NotYet: reading valueDeclaration | 5 |
| NotYet: reading watcher | 5 |
| NotYet: reading yieldType | 5 |
| Refused: a method read as a value (directoryExists would lose its object, and this with it) | 5 |
| Refused: a method read as a value (getCanonicalFileName would lose its object, and this with it) | 5 |
| Refused: a method read as a value (trace would lose its object, and this with it) | 5 |
| Refused: a value of type CompilerOptionsValue seen as TsConfigSourceFile \| CompilerOptionsValue, which can write string \| number where string is read | 5 |
| Refused: a value of type DiagnosticWithDetachedLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where undefined is read | 5 |
| Refused: a value of type FlowArrayMutation \| FlowAssignment seen as FlowNode, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 5 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier \| Node seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 5 |
| Refused: a value of type Node seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 5 |
| Refused: a value of type Symbol \| undefined seen as Declaration \| undefined, which can write number \| undefined where number is read | 5 |
| Refused: a value of type string[] seen as CompilerOptionsValue, which can write string \| number where string is read | 5 |
| Refused: an arbitrary number or a value from another enum assigned to InternalSymbolName.Call; its members are a closed union | 5 |
| Refused: an unproven value assigned to a numeric literal or enum member slot CompilerOptions \| BuilderFileEmit | 5 |
| Refused: an unproven value assigned to a numeric literal or enum member slot ModifierFlags.None | 5 |
| Refused: optional property all in CompilerOptions absent from structural source BuildOptions, which can hide fields | 5 |
| Refused: optional property reportsUnnecessary in Diagnostic absent from structural source DiagnosticRelatedInformation, which can hide fields | 5 |
| Refused: optional property return in ArrayIterator<Declaration> absent from structural source ArrayIterator<ClassElement>, which can hide fields | 5 |
| Refused: optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<HeritageClause>, which can hide fields | 5 |
| Refused: optional property skipLogging in ErrorOutputContainer absent from structural source { errors?: Diagnostic[] \| undefined; }, which can hide fields | 5 |
| Refused: optional property version in PackageJson absent from structural source PackageJsonPathFields, which can hide fields | 5 |
| NotYet: a BinaryExpression with a string and a value | 4 |
| NotYet: a VoidExpression as a statement | 4 |
| NotYet: a computed field name | 4 |
| NotYet: a field of type "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number> | 4 |
| NotYet: a function returning CompilerOptionsValue | 4 |
| NotYet: a function returning NodeArray<T> | 4 |
| NotYet: a function returning PackageJson[K] \| undefined | 4 |
| NotYet: a function returning R | 4 |
| NotYet: a function returning SortedReadonlyArray<T> | 4 |
| NotYet: a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables | 4 |
| NotYet: a value of type (ConstructorDeclaration & { body: Block; }) \| undefined | 4 |
| NotYet: a value of type (EmitNode & { autoGenerate: AutoGenerateInfo; }) \| (EmitNode & { autoGenerate: AutoGenerateInfo; }) | 4 |
| NotYet: a value of type CompilerOptionsValue | 4 |
| NotYet: a value of type ExpressionWithTypeArguments & { readonly expression: Identifier \| PropertyAccessEntityNameExpression; } | 4 |
| NotYet: a value of type HasJSDoc | 4 |
| NotYet: a value of type T \| Program | 4 |
| NotYet: a value of type string \| (void & { __escapedIdentifier: void; }) \| (string & { __escapedIdentifier: void; }) | 4 |
| NotYet: an array of ResolvedConfigFileName | 4 |
| NotYet: an array of boolean \| undefined | 4 |
| NotYet: assigning to an ObjectLiteralExpression | 4 |
| NotYet: reading BuilderState | 4 |
| NotYet: reading args | 4 |
| NotYet: reading argument | 4 |
| NotYet: reading asteriskToken | 4 |
| NotYet: reading baseTypeNode | 4 |
| NotYet: reading bindings | 4 |
| NotYet: reading cached | 4 |
| NotYet: reading cachedChain | 4 |
| NotYet: reading canReportSummary | 4 |
| NotYet: reading chain | 4 |
| NotYet: reading childrenTargetType | 4 |
| NotYet: reading commonSourceDirectory | 4 |
| NotYet: reading config | 4 |
| NotYet: reading constraint | 4 |
| NotYet: reading container | 4 |
| NotYet: reading cooked | 4 |
| NotYet: reading descriptorName | 4 |
| NotYet: reading directoryExists | 4 |
| NotYet: reading e | 4 |
| NotYet: reading emitNode | 4 |
| NotYet: reading emitResult | 4 |
| NotYet: reading excludeRegex | 4 |
| NotYet: reading exportStatement | 4 |
| NotYet: reading extensionGroup | 4 |
| NotYet: reading func | 4 |
| NotYet: reading getCurrentDirectory | 4 |
| NotYet: reading hasDefaultClause | 4 |
| NotYet: reading id | 4 |
| NotYet: reading identifier | 4 |
| NotYet: reading importDeclaration | 4 |
| NotYet: reading inference | 4 |
| NotYet: reading instantiation | 4 |
| NotYet: reading jsxFactorySymbol | 4 |
| NotYet: reading lexicallyScopedSymbol | 4 |
| NotYet: reading mappedType | 4 |
| NotYet: reading members | 4 |
| NotYet: reading normalized | 4 |
| NotYet: reading openParenPosition | 4 |
| NotYet: reading original | 4 |
| NotYet: reading output | 4 |
| NotYet: reading parameters | 4 |
| NotYet: reading parseNode | 4 |
| NotYet: reading prop | 4 |
| NotYet: reading propertyOriginalNode | 4 |
| NotYet: reading props | 4 |
| NotYet: reading rawText | 4 |
| NotYet: reading refPath | 4 |
| NotYet: reading referencedName | 4 |
| NotYet: reading regularType | 4 |
| NotYet: reading resolvedModule | 4 |
| NotYet: reading restParamSymbol | 4 |
| NotYet: reading restParameter | 4 |
| NotYet: reading reversed | 4 |
| NotYet: reading seen | 4 |
| NotYet: reading sourceIndex | 4 |
| NotYet: reading sourceSymbol | 4 |
| NotYet: reading state | 4 |
| NotYet: reading targetSymbol | 4 |
| NotYet: reading typeSet | 4 |
| NotYet: reading watchCompilerHost | 4 |
| Refused: Object.defineProperties | 4 |
| Refused: a function taking readonly Modifier[] \| undefined seen as one taking readonly ModifierLike[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 4 |
| Refused: a method read as a value (getSourceFile would lose its object, and this with it) | 4 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on expr) | 4 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on f) | 4 |
| Refused: a value of type Declaration \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 4 |
| Refused: a value of type Expression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 4 |
| Refused: a value of type Expression \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 4 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 4 |
| Refused: a value of type Identifier[][] seen as ModuleExportName[][], which can write ModuleExportName[] where Identifier[] is read | 4 |
| Refused: a value of type Mutable<GeneratedIdentifier> seen as Identifier, which can write EmitNode \| undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 4 |
| Refused: a value of type NodeArray<Statement> seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 4 |
| Refused: a value of type SyntheticSuper seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 4 |
| Refused: a value of type never[] seen as Node[], which can write Node where never is read | 4 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Identifier | 4 |
| Refused: optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<ClassElement>, which can hide fields | 4 |
| Refused: optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<T>, which can hide fields | 4 |
| Refused: optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<TypeElement>, which can hide fields | 4 |
| Refused: optional property return in ArrayIterator<any> absent from structural source ArrayIterator<Child>, which can hide fields | 4 |
| NotYet: Object.entries on a shape not proven by a plain literal or its const binding | 3 |
| NotYet: RegExp with a nonconstant pattern | 3 |
| NotYet: a BinaryExpression with a number \| undefined and a boolean | 3 |
| NotYet: a BinaryExpression with a union of differently held members and a value | 3 |
| NotYet: a ConditionalExpression as a statement | 3 |
| NotYet: a Map of false \| MutableFileSystemEntries | 3 |
| NotYet: a call to an Identifier | 3 |
| NotYet: a function returning ImmediatelyInvokedArrowFunction | 3 |
| NotYet: a function returning ResolvedConfigFileName | 3 |
| NotYet: a function returning T \| T[] \| undefined | 3 |
| NotYet: a function returning T \| readonly T[] \| undefined | 3 |
| NotYet: a function returning object | 3 |
| NotYet: a narrowed boolean \| undefined field; copy the field into a local and narrow that local instead | 3 |
| NotYet: a value of type BindableAccessExpression | 3 |
| NotYet: a value of type BindableObjectDefinePropertyCall | 3 |
| NotYet: a value of type Children \| undefined | 3 |
| NotYet: a value of type EndOfFileToken | 3 |
| NotYet: a value of type Identifier \| __String | 3 |
| NotYet: a value of type ImmediatelyInvokedArrowFunction | 3 |
| NotYet: a value of type InitializedVariableDeclaration | 3 |
| NotYet: a value of type T[] | 3 |
| NotYet: a value of type U | 3 |
| NotYet: a value of type WrappedExpression<AnonymousFunctionDefinition> | 3 |
| NotYet: a value of type X | 3 |
| NotYet: a value of type false \| TypeOnlyAliasDeclaration \| undefined | 3 |
| NotYet: a value of type object \| undefined | 3 |
| NotYet: a value of type undefined | 3 |
| NotYet: an array of NonNullable<T> | 3 |
| NotYet: an array of undefined | 3 |
| NotYet: assigning an element of a value | 3 |
| NotYet: destructuring anything but a tuple into [names] | 3 |
| NotYet: reading addUndefined | 3 |
| NotYet: reading ancestorFacts | 3 |
| NotYet: reading arrayLiteral | 3 |
| NotYet: reading awaitedType | 3 |
| NotYet: reading baseTypes | 3 |
| NotYet: reading buildInfo | 3 |
| NotYet: reading builderProgram | 3 |
| NotYet: reading callee | 3 |
| NotYet: reading canUseSourceFile | 3 |
| NotYet: reading checkType | 3 |
| NotYet: reading childrenNameType | 3 |
| NotYet: reading clause | 3 |
| NotYet: reading constructor | 3 |
| NotYet: reading currentNode | 3 |
| NotYet: reading decorationStatements | 3 |
| NotYet: reading decorator | 3 |
| NotYet: reading diagnosticStart | 3 |
| NotYet: reading diff | 3 |
| NotYet: reading elementType | 3 |
| NotYet: reading emitKind | 3 |
| NotYet: reading emitSuperHelpers | 3 |
| NotYet: reading escapedValue | 3 |
| NotYet: reading evaluated | 3 |
| NotYet: reading exportEquals | 3 |
| NotYet: reading exports | 3 |
| NotYet: reading firstAccessorWithDecorators | 3 |
| NotYet: reading firstDecl | 3 |
| NotYet: reading firstDeclaration | 3 |
| NotYet: reading functionName | 3 |
| NotYet: reading getter | 3 |
| NotYet: reading graphNode | 3 |
| NotYet: reading hostNode | 3 |
| NotYet: reading indexInfos | 3 |
| NotYet: reading indexType | 3 |
| NotYet: reading inferredProp | 3 |
| NotYet: reading info | 3 |
| NotYet: reading instantiated | 3 |
| NotYet: reading introducesError | 3 |
| NotYet: reading isAmbient | 3 |
| NotYet: reading isAsync | 3 |
| NotYet: reading isCallExpression | 3 |
| NotYet: reading isSimilarNode | 3 |
| NotYet: reading item | 3 |
| NotYet: reading jsdocAliasDecl | 3 |
| NotYet: reading jsdocParameters | 3 |
| NotYet: reading key | 3 |
| NotYet: reading kind | 3 |
| NotYet: reading languageVersion | 3 |
| NotYet: reading lastDecorator | 3 |
| NotYet: reading lastError | 3 |
| NotYet: reading lastParam | 3 |
| NotYet: reading line | 3 |
| NotYet: reading major | 3 |
| NotYet: reading mapping | 3 |
| NotYet: reading method | 3 |
| NotYet: reading methodType | 3 |
| NotYet: reading modifier | 3 |
| NotYet: reading moduleSymbol | 3 |
| NotYet: reading moveModifiers | 3 |
| NotYet: reading nameType | 3 |
| NotYet: reading named | 3 |
| NotYet: reading newSymbol | 3 |
| NotYet: reading objectType | 3 |
| NotYet: reading oldEmitKind | 3 |
| NotYet: reading oldOptions | 3 |
| NotYet: reading oldStartN | 3 |
| NotYet: reading opcode | 3 |
| NotYet: reading parentType | 3 |
| NotYet: reading parts | 3 |
| NotYet: reading promoteToIIFE | 3 |
| NotYet: reading property | 3 |
| NotYet: reading questionToken | 3 |
| NotYet: reading range | 3 |
| NotYet: reading regularNew | 3 |
| NotYet: reading resolution | 3 |
| NotYet: reading resolvedRequire | 3 |
| NotYet: reading resolvedSymbol | 3 |
| NotYet: reading restElement | 3 |
| NotYet: reading s | 3 |
| NotYet: reading sourceEmitHelpers | 3 |
| NotYet: reading sourceFileOrBundle | 3 |
| NotYet: reading suggestion | 3 |
| NotYet: reading symbolFromSymbolTable | 3 |
| NotYet: reading tag | 3 |
| NotYet: reading targetFlags | 3 |
| NotYet: reading targetHasRestElement | 3 |
| NotYet: reading targetPropertySymbol | 3 |
| NotYet: reading typeArguments | 3 |
| NotYet: reading typeCopy | 3 |
| NotYet: reading typeSymbol | 3 |
| NotYet: reading varStatement | 3 |
| Refused: a function taking () => T seen as one taking () => T (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 |
| Refused: a function taking [fileName: string, text: string, writeByteOrderMark: boolean, onError?: ((message: string) => void) \| undefined, sourceFiles?: readonly SourceFile[] \| undefined, data?: WriteFileCallbackData \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 |
| Refused: a function taking number seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 |
| Refused: a function taking string seen as one taking [fileName: string, text: string, writeByteOrderMark: boolean, onError?: ((message: string) => void) \| undefined, sourceFiles?: readonly SourceFile[] \| undefined, data?: WriteFileCallbackData \| undefined] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 |
| Refused: a method read as a value (afterProgramCreate would lose its object, and this with it) | 3 |
| Refused: a method read as a value (afterProgramEmitAndDiagnostics would lose its object, and this with it) | 3 |
| Refused: a method read as a value (getDirectories would lose its object, and this with it) | 3 |
| Refused: a method read as a value (getParsedCommandLine would lose its object, and this with it) | 3 |
| Refused: a method read as a value (getSemanticDiagnosticsOfNextAffectedFile would lose its object, and this with it) | 3 |
| Refused: a method read as a value (now would lose its object, and this with it) | 3 |
| Refused: a method read as a value (onWatchStatusChange would lose its object, and this with it) | 3 |
| Refused: a method read as a value (reportLikelyUnsafeImportRequiredError would lose its object, and this with it) | 3 |
| Refused: a method read as a value (reportPrivateInBaseOfClassExpression would lose its object, and this with it) | 3 |
| Refused: a method read as a value (trackSymbol would lose its object, and this with it) | 3 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on block) | 3 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on d) | 3 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on element) | 3 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on entry) | 3 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on info) | 3 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on member) | 3 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on tag) | 3 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on value) | 3 |
| Refused: a value of type BindingOrAssignmentElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type ClassDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type Diagnostic seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 3 |
| Refused: a value of type FutureSourceFile \| SourceFile seen as Pick<SourceFile, "fileName" \| "impliedNodeFormat">, whose readonly field fileName becomes writable: a readonly field may hold something narrower than string, which a write of string would replace | 3 |
| Refused: a value of type GeneratedIdentifier seen as Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 3 |
| Refused: a value of type GenericType \| ResolvedType seen as ObjectType, which can write SymbolTable \| undefined where SymbolTable is read | 3 |
| Refused: a value of type Identifier seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type Identifier \| TextRange seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type Identifier \| undefined seen as Identifier \| TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type Map<string, string[] \| never[]> seen as Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>, which can write string where never is read | 3 |
| Refused: a value of type NodeArray<Statement> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type ParameterDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type PropertyName seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type Signature \| undefined seen as Type \| undefined, which can write TypeFlags where SignatureFlags is read | 3 |
| Refused: a value of type SourceFile seen as SourceFile \| SourceFileLike, which can write readonly number[] \| undefined where readonly number[] is read | 3 |
| Refused: a value of type TsConfigSourceFile \| CompilerOptionsValue seen as string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined, which can write string \| number where string is read | 3 |
| Refused: a value of type Type \| undefined seen as Symbol \| undefined, which can write SymbolFlags where TypeFlags is read | 3 |
| Refused: a value of type Type \| undefined seen as TypeNode \| undefined, which can write number \| undefined where number is read | 3 |
| Refused: a value of type YieldExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 |
| Refused: a value of type never[] seen as Declaration[], which can write Declaration where never is read | 3 |
| Refused: a value of type never[] seen as DiagnosticArguments, which can write string \| number \| boolean \| readonly string[] \| SourceFile \| undefined where never is read | 3 |
| Refused: a value of type never[] seen as ResolvedConfigFileName[], which can write ResolvedConfigFileName where never is read | 3 |
| Refused: a value of type never[] seen as TypeParameter[], which can write TypeParameter where never is read | 3 |
| Refused: a value of type never[] seen as string[] \| never[], which can write string where never is read | 3 |
| Refused: a value of type readonly Extension[][] seen as readonly string[][], which can write string where Extension is read | 3 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint WatchedFileWithIsClosed can be written, so it can write what undefined can't hold | 3 |
| Refused: an arbitrary number or a value from another enum assigned to Extension.Ts; its members are a closed union | 3 |
| Refused: an unproven relation from StructuredType to ObjectType: optional field members has no proven compatible presence/type | 3 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ExpressionWithTypeArguments | 3 |
| Refused: optional property id in FalseLiteral absent from structural source never, which can hide fields | 3 |
| Refused: optional property id in LiteralExpression & StringLiteral absent from structural source never, which can hide fields | 3 |
| Refused: optional property omitTrailingSemicolon in PrinterOptions absent from structural source CompilerOptions, which can hide fields | 3 |
| Refused: optional property rawText in TemplateLiteralLikeNode absent from structural source NumericLiteral, which can hide fields | 3 |
| Refused: optional property resolutionMode in FileReference absent from structural source { preserve: true; }, which can hide fields | 3 |
| Refused: optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<EnumMember>, which can hide fields | 3 |
| Refused: optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<TemplateSpan>, which can hide fields | 3 |
| Refused: optional property return in ArrayIterator<Statement> absent from structural source ArrayIterator<ExpressionStatement>, which can hide fields | 3 |
| SkippedDependency: a dependency function whose body has checker diagnostics (measurement skipped) | 3 |
| NotYet: a BinaryExpression with a number \| undefined and a number \| undefined | 2 |
| NotYet: a BinaryExpression with a union of differently held members and a boolean | 2 |
| NotYet: a BinaryExpression with a union of differently held members and a string | 2 |
| NotYet: a BinaryExpression with a value and a number \| undefined | 2 |
| NotYet: a Map of HostFileInfo | 2 |
| NotYet: a Map of VisitResult<ExportAssignment \| LateVisibilityPaintedStatement \| undefined> | 2 |
| NotYet: a YieldExpression as a statement | 2 |
| NotYet: a boolean \| undefined variable a function value captures | 2 |
| NotYet: a destructured name that isn't plain | 2 |
| NotYet: a field of type AnyBuildOrder \| undefined | 2 |
| NotYet: a field of type boolean \| (() => boolean) \| undefined | 2 |
| NotYet: a field of type string \| number \| boolean \| DiagnosticMessage \| undefined | 2 |
| NotYet: a field of type string \| number \| undefined | 2 |
| NotYet: a function returning Extract<ClassDeclaration, Pick<...>> \| Extract<...> | 2 |
| NotYet: a function returning ImmediatelyInvokedFunctionExpression | 2 |
| NotYet: a function returning ModeAwareCacheKey | 2 |
| NotYet: a function returning Path \| undefined | 2 |
| NotYet: a function returning T \| EmptyStatement \| undefined | 2 |
| NotYet: a function returning U[] | 2 |
| NotYet: a function returning U[] \| undefined | 2 |
| NotYet: a function returning V | 2 |
| NotYet: a function returning WatchFactory<X, Y>[T] | 2 |
| NotYet: a function returning readonly T[] | 2 |
| NotYet: a function returning readonly T[] \| undefined | 2 |
| NotYet: a function returning void \| "skip" | 2 |
| NotYet: a tagged template other than the intrinsic String.raw | 2 |
| NotYet: a tuple element of type string \| number \| boolean \| readonly string[] \| SourceFile \| undefined | 2 |
| NotYet: a union of differently held members variable a function value captures | 2 |
| NotYet: a value of type (AssignmentExpression<EqualsToken> & { readonly left: GeneratedIdentifier; }) \| undefined | 2 |
| NotYet: a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) \| undefined | 2 |
| NotYet: a value of type AnonymousFunctionDefinition | 2 |
| NotYet: a value of type Canonicalized | 2 |
| NotYet: a value of type CompilerHost & ReadBuildProgramHost | 2 |
| NotYet: a value of type ExpressionWithTypeArguments & { expression: Identifier \| PropertyAccessEntityNameExpression; } | 2 |
| NotYet: a value of type IncrementalBuildInfoFileId | 2 |
| NotYet: a value of type K \| undefined | 2 |
| NotYet: a value of type Map<K, T> | 2 |
| NotYet: a value of type ParameterPropertyDeclaration | 2 |
| NotYet: a value of type Path \| undefined | 2 |
| NotYet: a value of type SourceFile | 2 |
| NotYet: a value of type T \| T[] | 2 |
| NotYet: a value of type T \| readonly T[] | 2 |
| NotYet: a value of type TInArray | 2 |
| NotYet: a value of type ThisCapturingVariableDeclaration | 2 |
| NotYet: a value of type TypeNode & LiteralTypeNode & { readonly literal: StringLiteral; } | 2 |
| NotYet: a value of type U \| readonly U[] \| undefined | 2 |
| NotYet: a value of type U \| undefined | 2 |
| NotYet: a value of type V \| undefined | 2 |
| NotYet: a value of type false \| RegExpExecArray \| null | 2 |
| NotYet: a value of type never | 2 |
| NotYet: an array of Child | 2 |
| NotYet: an array of V | 2 |
| NotYet: destructuring a value | 2 |
| NotYet: indexOf on a tuple (a tuple is held as an object, not an array, so far; write it as an array where it's made) | 2 |
| NotYet: lastIndexOf with these arguments | 2 |
| NotYet: passing union of differently held members to a function value | 2 |
| NotYet: reading IncrementalParser | 2 |
| NotYet: reading accessibleSymbolChain | 2 |
| NotYet: reading accessorDeclarations | 2 |
| NotYet: reading add | 2 |
| NotYet: reading allAccessors | 2 |
| NotYet: reading allowStructuralFallback | 2 |
| NotYet: reading assignClassAliasInStaticBlock | 2 |
| NotYet: reading assignment | 2 |
| NotYet: reading attributesType | 2 |
| NotYet: reading baseClassType | 2 |
| NotYet: reading baseSymbol | 2 |
| NotYet: reading cacheAssignment | 2 |
| NotYet: reading cachedPackageJson | 2 |
| NotYet: reading cachedTypes | 2 |
| NotYet: reading call | 2 |
| NotYet: reading callbackToAdd | 2 |
| NotYet: reading capturedLeft | 2 |
| NotYet: reading change0 | 2 |
| NotYet: reading classLikeDeclaration | 2 |
| NotYet: reading columns | 2 |
| NotYet: reading commentRange | 2 |
| NotYet: reading comments | 2 |
| NotYet: reading compilerHost | 2 |
| NotYet: reading constructSignatures | 2 |
| NotYet: reading containingCall | 2 |
| NotYet: reading containingFileName | 2 |
| NotYet: reading contextualType | 2 |
| NotYet: reading curCategory | 2 |
| NotYet: reading declBlocked | 2 |
| NotYet: reading declarationContainer | 2 |
| NotYet: reading declarationFilePath | 2 |
| NotYet: reading declarationTransform | 2 |
| NotYet: reading declarations | 2 |
| NotYet: reading declaredType | 2 |
| NotYet: reading decorators | 2 |
| NotYet: reading deduplicated | 2 |
| NotYet: reading defaultIndex | 2 |
| NotYet: reading derived | 2 |
| NotYet: reading diagnostic | 2 |
| NotYet: reading directoryPath | 2 |
| NotYet: reading dupFile | 2 |
| NotYet: reading effectiveTarget | 2 |
| NotYet: reading elementTypes | 2 |
| NotYet: reading emitFlags | 2 |
| NotYet: reading emitSkipped | 2 |
| NotYet: reading enclosingDeclaration | 2 |
| NotYet: reading endLabel | 2 |
| NotYet: reading endLength | 2 |
| NotYet: reading enter | 2 |
| NotYet: reading equalsToken | 2 |
| NotYet: reading errorInfo | 2 |
| NotYet: reading errorRecord | 2 |
| NotYet: reading escapedText | 2 |
| NotYet: reading expandedParams | 2 |
| NotYet: reading exportName | 2 |
| NotYet: reading exportStarFunction | 2 |
| NotYet: reading exportedName | 2 |
| NotYet: reading extraInitializersName | 2 |
| NotYet: reading fileInfos | 2 |
| NotYet: reading filePath | 2 |
| NotYet: reading first | 2 |
| NotYet: reading firstArgument | 2 |
| NotYet: reading firstBase | 2 |
| NotYet: reading firstParameterIsThis | 2 |
| NotYet: reading firstType | 2 |
| NotYet: reading flattenContext | 2 |
| NotYet: reading freshType | 2 |
| NotYet: reading fromCache | 2 |
| NotYet: reading fullName | 2 |
| NotYet: reading generatorYieldType | 2 |
| NotYet: reading getCanonicalFileName | 2 |
| NotYet: reading getCommonSourceDirectory | 2 |
| NotYet: reading getModifiedTime | 2 |
| NotYet: reading getOptionsNameMap | 2 |
| NotYet: reading gutterWidth | 2 |
| NotYet: reading hasArguments | 2 |
| NotYet: reading hasDefault | 2 |
| NotYet: reading heritageClause | 2 |
| NotYet: reading hooks | 2 |
| NotYet: reading indexSymbol | 2 |
| NotYet: reading init | 2 |
| NotYet: reading initializerStatement | 2 |
| NotYet: reading instanceType | 2 |
| NotYet: reading invalidatedProject | 2 |
| NotYet: reading isAnyLike | 2 |
| NotYet: reading isAutomaticTypeInNonNull | 2 |
| NotYet: reading isFinite | 2 |
| NotYet: reading isOptionalChain | 2 |
| NotYet: reading isSetonlyAccessor | 2 |
| NotYet: reading isSingleNonGenericCandidate | 2 |
| NotYet: reading isUsed | 2 |
| NotYet: reading isVoidPromiseError | 2 |
| NotYet: reading iteratedType | 2 |
| NotYet: reading json | 2 |
| NotYet: reading jsxFragmentFactoryName | 2 |
| NotYet: reading keyPropertyName | 2 |
| NotYet: reading label | 2 |
| NotYet: reading lanes | 2 |
| NotYet: reading last | 2 |
| NotYet: reading lastId | 2 |
| NotYet: reading lastModifier | 2 |
| NotYet: reading lastParamVariadicType | 2 |
| NotYet: reading leadingNewlines | 2 |
| NotYet: reading leftIdentifier | 2 |
| NotYet: reading length | 2 |
| NotYet: reading lineNumber | 2 |
| NotYet: reading lines | 2 |
| NotYet: reading linesBeforeDot | 2 |
| NotYet: reading linkType | 2 |
| NotYet: reading list | 2 |
| NotYet: reading literalType | 2 |
| NotYet: reading loadPackageJsonMainState | 2 |
| NotYet: reading local | 2 |
| NotYet: reading longestParamType | 2 |
| NotYet: reading mapped | 2 |
| NotYet: reading mapperCache | 2 |
| NotYet: reading match | 2 |
| NotYet: reading meaning | 2 |
| NotYet: reading methodParameterType | 2 |
| NotYet: reading methodSignatures | 2 |
| NotYet: reading min | 2 |
| NotYet: reading missing | 2 |
| NotYet: reading modifiers | 2 |
| NotYet: reading mustBeRemoved | 2 |
| NotYet: reading names | 2 |
| NotYet: reading newParametersArray | 2 |
| NotYet: reading newParsedCommandLine | 2 |
| NotYet: reading next | 2 |
| NotYet: reading nextChange | 2 |
| NotYet: reading nextKey | 2 |
| NotYet: reading nextLineStart | 2 |
| NotYet: reading noTruncation | 2 |
| NotYet: reading nodeId | 2 |
| NotYet: reading oldNoInferenceFallback | 2 |
| NotYet: reading oldState | 2 |
| NotYet: reading openBracePosition | 2 |
| NotYet: reading operandConstraint | 2 |
| NotYet: reading operatorKind | 2 |
| NotYet: reading operatorToken | 2 |
| NotYet: reading origin | 2 |
| NotYet: reading originalClassDecl | 2 |
| NotYet: reading otherAccessor | 2 |
| NotYet: reading override | 2 |
| NotYet: reading param | 2 |
| NotYet: reading paramCount | 2 |
| NotYet: reading paramSymbol | 2 |
| NotYet: reading parametersWithPropertyAssignments | 2 |
| NotYet: reading parentSymbol | 2 |
| NotYet: reading parenthesizerRule | 2 |
| NotYet: reading parseTreeNode | 2 |
| NotYet: reading parsedCommandLine | 2 |
| NotYet: reading patterns | 2 |
| NotYet: reading potentiallyUnusedIdentifiers | 2 |
| NotYet: reading pragma | 2 |
| NotYet: reading predicate | 2 |
| NotYet: reading previous | 2 |
| NotYet: reading primaryTypes | 2 |
| NotYet: reading propName | 2 |
| NotYet: reading prototypeProperty | 2 |
| NotYet: reading questionDotToken | 2 |
| NotYet: reading r | 2 |
| NotYet: reading real | 2 |
| NotYet: reading reducedTypes | 2 |
| NotYet: reading ref | 2 |
| NotYet: reading relatedInfo | 2 |
| NotYet: reading relativeFileName | 2 |
| NotYet: reading remove | 2 |
| NotYet: reading resolvedProject | 2 |
| NotYet: reading resolvedTypeSymbol | 2 |
| NotYet: reading restIdent | 2 |
| NotYet: reading returnMethod | 2 |
| NotYet: reading returnTypeNode | 2 |
| NotYet: reading root | 2 |
| NotYet: reading setter | 2 |
| NotYet: reading shortest | 2 |
| NotYet: reading signatureDeclaration | 2 |
| NotYet: reading sorted | 2 |
| NotYet: reading sourceFilePath | 2 |
| NotYet: reading sourceFileWithAddedExtension | 2 |
| NotYet: reading sourceFlags | 2 |
| NotYet: reading sourceMappings | 2 |
| NotYet: reading sourceRoot | 2 |
| NotYet: reading sourceStart | 2 |
| NotYet: reading sourceSymbolFile | 2 |
| NotYet: reading sourceType | 2 |
| NotYet: reading span | 2 |
| NotYet: reading specifierSourceImports | 2 |
| NotYet: reading spread | 2 |
| NotYet: reading spreadType | 2 |
| NotYet: reading staticType | 2 |
| NotYet: reading str | 2 |
| NotYet: reading substitute | 2 |
| NotYet: reading sym | 2 |
| NotYet: reading symbolName | 2 |
| NotYet: reading targetIndex | 2 |
| NotYet: reading targetProp | 2 |
| NotYet: reading templateArguments | 2 |
| NotYet: reading terminalWidth | 2 |
| NotYet: reading testedSymbol | 2 |
| NotYet: reading thisAccess | 2 |
| NotYet: reading timerToUpdateChildWatches | 2 |
| NotYet: reading toWatch | 2 |
| NotYet: reading tracingEnabled | 2 |
| NotYet: reading transformed | 2 |
| NotYet: reading trueType | 2 |
| NotYet: reading ts | 2 |
| NotYet: reading typeLiteralNode | 2 |
| NotYet: reading typeLiteralSymbol | 2 |
| NotYet: reading typeOnlyDeclaration | 2 |
| NotYet: reading typeParameter | 2 |
| NotYet: reading typeParams | 2 |
| NotYet: reading typesVersions | 2 |
| NotYet: reading useDefineForClassFields | 2 |
| NotYet: reading valueSymbol | 2 |
| NotYet: reading valueType | 2 |
| NotYet: reading varDecl | 2 |
| NotYet: reading variable | 2 |
| NotYet: reading versionPaths | 2 |
| NotYet: reading visitedAccessorName | 2 |
| NotYet: reading visitedSym | 2 |
| NotYet: reading widened | 2 |
| NotYet: reading yieldedType | 2 |
| NotYet: storing true \| Node \| undefined in a field | 2 |
| Refused: Object.create | 2 |
| Refused: a function taking ((node: Node) => VisitResult<Node>) \| undefined seen as one taking ((node: Node) => VisitResult<Node \| undefined>) \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking BinaryOperator seen as one taking SyntaxKind (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking Expression seen as one taking Expression \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking Expression[] seen as one taking readonly Expression[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking GeneratedIdentifierFlags seen as one taking GeneratedIdentifierFlags \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking Node seen as one taking [node: Node] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking NodeFlags seen as one taking NodeFlags \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking T seen as one taking Expression (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking Visitor seen as one taking Visitor<TIn, Node \| undefined> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking [fileName: string, languageVersionOrOptions: CreateSourceFileOptions \| ScriptTarget, onError?: ((message: string) => void) \| undefined, shouldCreateNewSourceFile?: boolean \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking string seen as one taking [fileName: string] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a function taking string seen as one taking any[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 |
| Refused: a method read as a value (clearTimeout would lose its object, and this with it) | 2 |
| Refused: a method read as a value (createDirectory would lose its object, and this with it) | 2 |
| Refused: a method read as a value (emitNodeWithNotification would lose its object, and this with it) | 2 |
| Refused: a method read as a value (enableCPUProfiler would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getBuildInfo would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getCurrentDirectory would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getEnvironmentVariable would lose its object, and this with it) | 2 |
| Refused: a method read as a value (getGlobalTypingsCacheLocation would lose its object, and this with it) | 2 |
| Refused: a method read as a value (hasGlobalName would lose its object, and this with it) | 2 |
| Refused: a method read as a value (isEmitNotificationEnabled would lose its object, and this with it) | 2 |
| Refused: a method read as a value (parenthesizeBranchOfConditionalExpression would lose its object, and this with it) | 2 |
| Refused: a method read as a value (parenthesizeConstituentTypesOfIntersectionType would lose its object, and this with it) | 2 |
| Refused: a method read as a value (parenthesizeConstituentTypesOfUnionType would lose its object, and this with it) | 2 |
| Refused: a method read as a value (parenthesizeNonArrayTypeOfPostfixType would lose its object, and this with it) | 2 |
| Refused: a method read as a value (reportInaccessibleUniqueSymbolError would lose its object, and this with it) | 2 |
| Refused: a method read as a value (setPrototypeOf would lose its object, and this with it) | 2 |
| Refused: a method read as a value (setTimeout would lose its object, and this with it) | 2 |
| Refused: a method read as a value (substituteNode would lose its object, and this with it) | 2 |
| Refused: a method read as a value (toKey would lose its object, and this with it) | 2 |
| Refused: a method read as a value (useCaseSensitiveFileNames would lose its object, and this with it) | 2 |
| Refused: a method read as a value (watchDirectory would lose its object, and this with it) | 2 |
| Refused: a method read as a value (watchFile would lose its object, and this with it) | 2 |
| Refused: a method read as a value (writeFile would lose its object, and this with it) | 2 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on declaration) | 2 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on predicate) | 2 |
| Refused: a value of type (identifierOrPrivateName: Identifier \| PrivateIdentifier) => string seen as (name: GeneratedIdentifier \| GeneratedPrivateIdentifier) => string, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 2 |
| Refused: a value of type (readonly [() => IntrinsicType, __String])[] seen as (readonly [() => Type, __String])[], which can write readonly [() => Type, __String] where readonly [() => IntrinsicType, __String] is read | 2 |
| Refused: a value of type ArrayLiteralExpression \| AssignmentExpression<EqualsToken> \| BindingElement \| ElementAccessExpression \| ... 8 more ... \| VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type BindingElement[] seen as unknown[], which can write unknown where BindingElement is read | 2 |
| Refused: a value of type Block seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type BlockLike seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type BreakStatement seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type CapturedThis seen as Expression \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 2 |
| Refused: a value of type CapturedThis seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 2 |
| Refused: a value of type ClassElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ClassStaticBlockDeclaration seen as Mutable<ClassStaticBlockDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ClassStaticBlockDeclaration \| PropertyDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type CompilerOptions & { types: string[]; } seen as CompilerOptions, which can write string[] \| undefined where string[] is read | 2 |
| Refused: a value of type ConstructorDeclaration seen as Mutable<ConstructorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ContinueStatement seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type DiagnosticWithLocation \| undefined seen as Diagnostic \| undefined, which can write SourceFile \| undefined where SourceFile is read | 2 |
| Refused: a value of type DiagnosticWithLocation[] \| undefined seen as DiagnosticRelatedInformation[] \| undefined, which can write DiagnosticRelatedInformation where DiagnosticWithLocation is read | 2 |
| Refused: a value of type ElementAccessExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type FlowCall seen as FlowNode, which can write BinaryExpression \| CallExpression where CallExpression is read | 2 |
| Refused: a value of type FutureSourceFile \| SourceFile seen as Pick<SourceFile, "fileName" \| "impliedNodeFormat" \| "packageJsonScope">, whose readonly field fileName becomes writable: a readonly field may hold something narrower than string, which a write of string would replace | 2 |
| Refused: a value of type GeneratedIdentifier seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 2 |
| Refused: a value of type GeneratedIdentifier seen as Identifier \| PrivateIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 2 |
| Refused: a value of type Identifier seen as Mutable<Identifier>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Identifier[] \| undefined seen as ModuleExportName[] \| undefined, which can write ModuleExportName where Identifier is read | 2 |
| Refused: a value of type ImportTypeAssertionContainer seen as Mutable<ImportTypeAssertionContainer>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type LeftHandSideExpression & GeneratedIdentifier seen as Expression, which can write EmitNode \| undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 2 |
| Refused: a value of type MethodDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type MissingDeclaration seen as Mutable<MissingDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ModifierLike seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ModuleSpecifierResolutionHost & ModuleResolutionHost seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where (() => boolean) & (boolean \| (() => boolean) \| undefined) is read | 2 |
| Refused: a value of type Node seen as Node \| TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Node \| TextRange seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Node \| undefined seen as NodeLinks \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of NodeCheckFlags would replace | 2 |
| Refused: a value of type NumericLiteral seen as Mutable<LiteralExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ParameterDeclaration \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 2 |
| Refused: a value of type PropertySignature seen as Mutable<PropertySignature>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type ReturnStatement seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type SourceFile seen as EmitNode \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of EmitFlags would replace | 2 |
| Refused: a value of type SourceFile seen as Mutable<SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Statement seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Symbol seen as Declaration \| undefined, which can write number \| undefined where number is read | 2 |
| Refused: a value of type Symbol[] seen as unknown[], which can write unknown where Symbol is read | 2 |
| Refused: a value of type T seen as T, a type parameter whose constraint Node can be written, so it can write what T can't hold | 2 |
| Refused: a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type Type[] seen as unknown[], which can write unknown where Type is read | 2 |
| Refused: a value of type VariableDeclaration seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type VariableStatement seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 |
| Refused: a value of type any seen as T, a type parameter whose constraint Node can be written, so it can write what any can't hold | 2 |
| Refused: a value of type false \| FlowNode \| undefined seen as false \| FlowAssignment \| FlowLabel \| FlowReduceLabel \| FlowStart \| FlowSwitchClause \| undefined, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 2 |
| Refused: a value of type never[] seen as (Identifier \| StringLiteral)[], which can write Identifier \| StringLiteral where never is read | 2 |
| Refused: a value of type never[] seen as ChildDirectoryWatcher[], which can write ChildDirectoryWatcher where never is read | 2 |
| Refused: a value of type never[] seen as DiagnosticWithLocation[], which can write DiagnosticWithLocation where never is read | 2 |
| Refused: a value of type never[] seen as SourceFile[], which can write SourceFile where never is read | 2 |
| Refused: a value of type never[] seen as StringLiteralLike[], which can write StringLiteralLike where never is read | 2 |
| Refused: a value of type never[] seen as VariableDeclaration[], which can write VariableDeclaration where never is read | 2 |
| Refused: a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[] \| undefined, which can write SourceFile \| undefined where SourceFile is read | 2 |
| Refused: a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 2 |
| Refused: a value of type string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] seen as unknown[], which can write unknown where string is read | 2 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"FlowFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof FlowFlags is read | 2 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint Type can be written, so it can write what undefined can't hold | 2 |
| Refused: an unproven relation from Declaration to NamedDeclaration: optional field name has no proven compatible presence/type | 2 |
| Refused: an unproven relation from ObjectType to AnonymousType: optional field target has no proven compatible presence/type | 2 |
| Refused: an unproven relation from Type to SyntheticDefaultModuleType: optional field syntheticType has no proven compatible presence/type | 2 |
| Refused: an unproven relation from Type to TypeVariable: optional field constraint has no proven compatible presence/type | 2 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.FunctionExpression | 2 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NoSubstitutionTemplateLiteral \| SyntaxKind.TemplateHead \| SyntaxKind.TemplateMiddle \| SyntaxKind.TemplateTail | 2 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.WithKeyword \| SyntaxKind.AssertKeyword | 2 |
| Refused: an unproven value assigned to a numeric literal or enum member slot Ternary | 2 |
| Refused: delete | 2 |
| Refused: inherited library member setPrototypeOf read as an own field | 2 |
| Refused: optional property all in CompilerOptions absent from structural source {}, which can hide fields | 2 |
| Refused: optional property constraint in TypeParameter absent from structural source InterfaceType, which can hide fields | 2 |
| Refused: optional property equalsToken in ShorthandPropertyAssignment absent from structural source never, which can hide fields | 2 |
| Refused: optional property id in NamedExports absent from structural source never, which can hide fields | 2 |
| Refused: optional property modifiers in GetAccessorDeclaration absent from structural source never, which can hide fields | 2 |
| Refused: optional property packageJsonScope in Pick<SourceFile, "fileName" \| "impliedNodeFormat" \| "packageJsonScope"> absent from structural source Pick<SourceFile, "fileName" \| "impliedNodeFormat">, which can hide fields | 2 |
| Refused: optional property return in ArrayIterator<Diagnostic> absent from structural source ArrayIterator<DiagnosticWithLocation>, which can hide fields | 2 |
| Refused: optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<JSDocTag>, which can hide fields | 2 |
| Refused: optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<TemplateLiteralTypeSpan>, which can hide fields | 2 |
| Refused: optional property templateFlags in NoSubstitutionTemplateLiteral absent from structural source never, which can hide fields | 2 |
| Refused: optional property textSourceNode in StringLiteral absent from structural source never, which can hide fields | 2 |
| panic: statement panic: Unhandled case in Node.Text: *ast.ComputedPropertyName | 2 |
| NotYet: .length on a value | 1 |
| NotYet: ?. to a number, which would be number \| undefined | 1 |
| NotYet: Array as a value outside equality or typeof (overloaded calls and static properties need their own representation) | 1 |
| NotYet: JSON.stringify a union containing containers without runtime element metadata | 1 |
| NotYet: JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) | 1 |
| NotYet: Object.assign on a shape not proven by a plain literal or its const binding | 1 |
| NotYet: a BinaryExpression with a boolean and a boolean | 1 |
| NotYet: a BinaryExpression with a boolean \| undefined and a boolean \| undefined | 1 |
| NotYet: a BinaryExpression with a boolean \| undefined and a value | 1 |
| NotYet: a BinaryExpression with a number and a union of differently held members | 1 |
| NotYet: a BinaryExpression with a number \| undefined and a boolean \| undefined | 1 |
| NotYet: a BinaryExpression with a number \| undefined and a string | 1 |
| NotYet: a BinaryExpression with a union of differently held members and a union of differently held members | 1 |
| NotYet: a BinaryExpression with a value and a boolean \| undefined | 1 |
| NotYet: a ClassExpression | 1 |
| NotYet: a Map of CompilerOptionsValue | 1 |
| NotYet: a Map of RedirectsCacheKey | 1 |
| NotYet: a Map of ResolvedConfigFilePath | 1 |
| NotYet: a Map of T | 1 |
| NotYet: a Map of string \| number | 1 |
| NotYet: a Map whose key and value types aren't known | 1 |
| NotYet: a PostfixUnaryExpression | 1 |
| NotYet: a Set of ResolvedConfigFilePath (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far) | 1 |
| NotYet: a SpreadElement | 1 |
| NotYet: a base that isn't a declared class | 1 |
| NotYet: a call returning T | 1 |
| NotYet: a call returning any | 1 |
| NotYet: a case that isn't a constant | 1 |
| NotYet: a class instantiated with TOuterState | 1 |
| NotYet: a class method through a view that erases its prototype origin | 1 |
| NotYet: a comparator that doesn't take two elements and return a number | 1 |
| NotYet: a destructured name held otherwise than its field | 1 |
| NotYet: a destructured parameter beside a parameter with a default | 1 |
| NotYet: a field from a boolean \| undefined variable | 1 |
| NotYet: a field of type "boolean" \| "list" \| "number" \| "object" \| "string" \| Map<string, string \| number> | 1 |
| NotYet: a field of type "boolean" \| "number" \| "object" \| "string" \| Map<string, string \| number> | 1 |
| NotYet: a field of type "circularity" \| boolean | 1 |
| NotYet: a field of type 0 \| boolean \| undefined | 1 |
| NotYet: a field of type NodeArray<ParameterDeclaration> \| readonly JSDocParameterTag[] | 1 |
| NotYet: a field of type boolean \| (() => boolean) | 1 |
| NotYet: a field of type false \| Type \| undefined | 1 |
| NotYet: a field of type false \| VersionPaths | 1 |
| NotYet: a field of type false \| VersionPaths \| undefined | 1 |
| NotYet: a field of type false \| string[] \| undefined | 1 |
| NotYet: a field of type string \| false | 1 |
| NotYet: a field of type string \| false \| undefined | 1 |
| NotYet: a for...of destructuring an object | 1 |
| NotYet: a function returning () => T | 1 |
| NotYet: a function returning (...args: T) => boolean | 1 |
| NotYet: a function returning (AssignmentExpression<EqualsToken> & { readonly left: GeneratedIdentifier; }) \| undefined | 1 |
| NotYet: a function returning (ConstructorDeclaration & { body: Block; }) \| undefined | 1 |
| NotYet: a function returning (T \| U)[] | 1 |
| NotYet: a function returning (arg: A) => T | 1 |
| NotYet: a function returning (node: BinaryExpression, outerState: TOuterState) => TResult | 1 |
| NotYet: a function returning A | 1 |
| NotYet: a function returning AnyValidImportOrReExport | 1 |
| NotYet: a function returning AnyValidImportOrReExport \| undefined | 1 |
| NotYet: a function returning BuildInvalidedProject<T> | 1 |
| NotYet: a function returning BuildInvalidedProject<T> \| UpdateOutputFileStampsProject | 1 |
| NotYet: a function returning CacheWithRedirects<K, V> | 1 |
| NotYet: a function returning CanonicalKey | 1 |
| NotYet: a function returning CapturedThis | 1 |
| NotYet: a function returning ClassExpression \| ImmediatelyInvokedArrowFunction | 1 |
| NotYet: a function returning ClassNamedEvaluationHelperBlock | 1 |
| NotYet: a function returning ClassStaticBlockDeclaration \| Decorator \| PrivateIdentifierGetAccessorDeclaration \| ... 5 more ... \| undefined | 1 |
| NotYet: a function returning ClassThisAssignmentBlock | 1 |
| NotYet: a function returning Declaration & HasModifiers | 1 |
| NotYet: a function returning EndOfFileToken | 1 |
| NotYet: a function returning EvaluatorResult<T> | 1 |
| NotYet: a function returning ExpressionWithTypeArguments & { expression: Identifier \| PropertyAccessEntityNameExpression; } | 1 |
| NotYet: a function returning ExpressionWithTypeArguments & { readonly expression: Identifier \| PropertyAccessEntityNameExpression; } | 1 |
| NotYet: a function returning Extract<AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; }, Pick<...>> \| ... 7 more ... \| Extract<...> | 1 |
| NotYet: a function returning HasJSDoc \| undefined | 1 |
| NotYet: a function returning IncrementalBuildInfoFileId | 1 |
| NotYet: a function returning IncrementalBuildInfoFileIdListId | 1 |
| NotYet: a function returning InferenceContext \| (T & undefined) | 1 |
| NotYet: a function returning InvalidatedProject<T> \| undefined | 1 |
| NotYet: a function returning Map<K, V1 \| V2> | 1 |
| NotYet: a function returning MemberName \| (Expression & (NumericLiteral \| StringLiteralLike)) | 1 |
| NotYet: a function returning MissingList<T> | 1 |
| NotYet: a function returning ModeAwareCache<T> | 1 |
| NotYet: a function returning ModuleOrTypeReferenceResolutionCache<T> | 1 |
| NotYet: a function returning MultiMap<K, V> | 1 |
| NotYet: a function returning Mutable<Token<TKind>> | 1 |
| NotYet: a function returning Node \| (TIn & undefined) \| (TVisited & undefined) | 1 |
| NotYet: a function returning NodeArray<Node> \| (TInArray & undefined) | 1 |
| NotYet: a function returning NodeArray<NonNullable<T>> \| undefined | 1 |
| NotYet: a function returning NodeArray<TOut> \| (TInArray & undefined) | 1 |
| NotYet: a function returning NonNullable<T> | 1 |
| NotYet: a function returning NonRelativeNameResolutionCache<T> | 1 |
| NotYet: a function returning PathPathComponents | 1 |
| NotYet: a function returning PerDirectoryResolutionCache<T> | 1 |
| NotYet: a function returning RedirectsCacheKey | 1 |
| NotYet: a function returning ResolvedConfigFilePath | 1 |
| NotYet: a function returning ReusableDiagnosticMessageChain | 1 |
| NotYet: a function returning SolutionBuilder<T> | 1 |
| NotYet: a function returning SyntheticSuper | 1 |
| NotYet: a function returning T \| EmptyStatement | 1 |
| NotYet: a function returning T \| Identifier | 1 |
| NotYet: a function returning T \| NumericLiteral \| StringLiteral \| BooleanLiteral | 1 |
| NotYet: a function returning T \| StringLiteral | 1 |
| NotYet: a function returning T \| T[] | 1 |
| NotYet: a function returning T \| readonly T[] | 1 |
| NotYet: a function returning T1 & T2 | 1 |
| NotYet: a function returning TEntry \| undefined | 1 |
| NotYet: a function returning TOut | 1 |
| NotYet: a function returning TOut \| (TIn & undefined) \| (TVisited & undefined) | 1 |
| NotYet: a function returning TOut \| undefined | 1 |
| NotYet: a function returning TPrivateEntry \| undefined | 1 |
| NotYet: a function returning TResult | 1 |
| NotYet: a function returning T[][] | 1 |
| NotYet: a function returning TransformationResult<T> | 1 |
| NotYet: a function returning TypeMapper \| (T & undefined) | 1 |
| NotYet: a function returning TypeOnlyAliasDeclaration \| undefined | 1 |
| NotYet: a function returning V \| undefined | 1 |
| NotYet: a function returning V[] | 1 |
| NotYet: a function returning VisitResult<T> | 1 |
| NotYet: a function returning WatchCompilerHostOfConfigFile<T> | 1 |
| NotYet: a function returning WatchCompilerHostOfFilesAndCompilerOptions<T> | 1 |
| NotYet: a function returning never | 1 |
| NotYet: a function returning object \| undefined | 1 |
| NotYet: a function returning readonly (readonly T[])[] | 1 |
| NotYet: a function returning readonly Node[] \| (TInArray & undefined) | 1 |
| NotYet: a function returning readonly Resolution[] | 1 |
| NotYet: a function returning readonly TOut[] \| (TInArray & undefined) | 1 |
| NotYet: a function returning readonly U[] | 1 |
| NotYet: a function returning readonly U[] \| undefined | 1 |
| NotYet: a function returning string \| object | 1 |
| NotYet: a function returning unknown | 1 |
| NotYet: a function returning void \| SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> | 1 |
| NotYet: a function returning void \| SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> \| WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> | 1 |
| NotYet: a function returning void \| WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> | 1 |
| NotYet: a function returning void \| number \| Symbol | 1 |
| NotYet: a function returning { [P in K as `${P}`]?: T[]; } | 1 |
| NotYet: a function returning { modifiers: NodeArray<Modifier> \| undefined; referencedName: Expression \| undefined; name: PropertyName; initializersName: Identifier \| undefined; descriptorName: Identifier \| undefined; thisArg: Identifier \| undefined; extraInitializersName?: never; } \| ... | 1 |
| NotYet: a function returning { readonly min: number; readonly max: number; } | 1 |
| NotYet: a function value taking boolean \| undefined | 1 |
| NotYet: a literal method through a view that erases its receiver | 1 |
| NotYet: a non-accessor method in an accessor literal | 1 |
| NotYet: a number \| undefined argument to slice | 1 |
| NotYet: a number \| undefined argument to substring | 1 |
| NotYet: a template interpolating an object, an array, a map, a function or undefined | 1 |
| NotYet: a tuple literal leaving out an element of type ModuleSpecifierEnding | 1 |
| NotYet: a value of type "" \| ResolvedConfigFileName \| undefined | 1 |
| NotYet: a value of type ((t: T) => U) \| undefined | 1 |
| NotYet: a value of type (AmbientModuleDeclaration & { name: StringLiteral; }) \| undefined | 1 |
| NotYet: a value of type (AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; } & BinaryExpression) \| (... & ... 1 more ... & BinaryExpression) | 1 |
| NotYet: a value of type (ExportDeclaration & { readonly isTypeOnly: true; readonly moduleSpecifier: Expression; }) \| undefined | 1 |
| NotYet: a value of type (ModuleDeclaration & { name: StringLiteral; }) \| undefined | 1 |
| NotYet: a value of type (VariableDeclaration & { name: Identifier; }) \| undefined | 1 |
| NotYet: a value of type (s: string) => void | 1 |
| NotYet: a value of type AccessExpression \| RequireOrImportCall | 1 |
| NotYet: a value of type AccessorDeclaration & { readonly name: BigIntLiteral \| ComputedPropertyName \| Identifier \| NoSubstitutionTemplateLiteral \| NumericLiteral \| StringLiteral; } | 1 |
| NotYet: a value of type AliasDeclarationNode | 1 |
| NotYet: a value of type ArrowFunction \| BinaryExpression \| BindingElement \| Block \| BreakStatement \| CallSignatureDeclaration \| ... 64 more ... \| EndOfFileToken | 1 |
| NotYet: a value of type BindablePropertyAssignmentExpression \| PropertyAccessExpression \| LiteralLikeElementAccessExpression | 1 |
| NotYet: a value of type BindableStaticAccessExpression | 1 |
| NotYet: a value of type BindingElement & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type CallExpression \| BindableStaticAccessExpression | 1 |
| NotYet: a value of type Child | 1 |
| NotYet: a value of type ClassElement \| ParameterPropertyDeclaration | 1 |
| NotYet: a value of type ClassNamedEvaluationHelperBlock | 1 |
| NotYet: a value of type CustomTransformerFactory \| TransformerFactory<T> | 1 |
| NotYet: a value of type EmitNode & { autoGenerate: AutoGenerateInfo; } | 1 |
| NotYet: a value of type EntityNameExpression \| (LeftHandSideExpression & BindableStaticNameExpression) | 1 |
| NotYet: a value of type ExportAssignment & { readonly expression: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type Identifier \| PrivateIdentifier \| __String | 1 |
| NotYet: a value of type IncludeTypeSpaceImports | 1 |
| NotYet: a value of type IncrementalBuildInfoFileIdListId | 1 |
| NotYet: a value of type IncrementalBuildInfoFilePendingEmit | 1 |
| NotYet: a value of type IncrementalMultiFileEmitBuildInfoFileInfo | 1 |
| NotYet: a value of type JSDocImportTag \| CanHaveModuleSpecifier | 1 |
| NotYet: a value of type LeftHandSideExpression & Identifier | 1 |
| NotYet: a value of type Map<Path, ModeAwareCache<T>> \| undefined | 1 |
| NotYet: a value of type Map<string, SingleFileWatcher<T>> | 1 |
| NotYet: a value of type Map<string, WildcardDirectoryWatcher<T>> | 1 |
| NotYet: a value of type Map<string, [K, V[]]> | 1 |
| NotYet: a value of type MapLike<T> | 1 |
| NotYet: a value of type ModeAwareCacheKey | 1 |
| NotYet: a value of type NamedEvaluation | 1 |
| NotYet: a value of type NodeArray<Expression> & readonly [BindableStaticNameExpression, NumericLiteral \| StringLiteralLike, ObjectLiteralExpression] & Readonly<...> | 1 |
| NotYet: a value of type NodeArray<T> | 1 |
| NotYet: a value of type NonNullable<K> | 1 |
| NotYet: a value of type NonNullable<U> | 1 |
| NotYet: a value of type ParameterDeclaration & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type PathPathComponents | 1 |
| NotYet: a value of type PrimitiveLiteral | 1 |
| NotYet: a value of type PrivateEnvironment<TData, TEntry> | 1 |
| NotYet: a value of type PrivateIdentifierInExpression | 1 |
| NotYet: a value of type PropertyAccessExpression \| LiteralLikeElementAccessExpression | 1 |
| NotYet: a value of type PropertyAccessExpression \| SyntheticSuper | 1 |
| NotYet: a value of type PropertyAssignment & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type PropertyDeclaration & { readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type PropertyDeclaration \| ParameterPropertyDeclaration | 1 |
| NotYet: a value of type RedirectsCacheKey | 1 |
| NotYet: a value of type RedirectsCacheKey \| undefined | 1 |
| NotYet: a value of type ReferencedFile & { kind: FileIncludeKind.LibReferenceDirective; } | 1 |
| NotYet: a value of type ReplaceableIndexedAccessType | 1 |
| NotYet: a value of type RequireOrImportCall | 1 |
| NotYet: a value of type ResolvedConfigFileName \| undefined | 1 |
| NotYet: a value of type ResolvedModuleWithFailedLookupLocations & ResolvedTypeReferenceDirectiveWithFailedLookupLocations | 1 |
| NotYet: a value of type Set<K> | 1 |
| NotYet: a value of type ShorthandPropertyAssignment & { readonly objectAssignmentInitializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type SortedArray<T> | 1 |
| NotYet: a value of type Source | 1 |
| NotYet: a value of type SourceFileOrString | 1 |
| NotYet: a value of type T \| T[] \| readonly T[] \| undefined | 1 |
| NotYet: a value of type T1 | 1 |
| NotYet: a value of type TData | 1 |
| NotYet: a value of type TEntry | 1 |
| NotYet: a value of type TKind \| Token<TKind> | 1 |
| NotYet: a value of type TNode | 1 |
| NotYet: a value of type T[] \| undefined | 1 |
| NotYet: a value of type TransformedSuperCall | 1 |
| NotYet: a value of type TypeParameterDeclaration & { parent: JSDocTemplateTag; } | 1 |
| NotYet: a value of type UnaryExpression & (BigIntLiteral \| NumericLiteral) | 1 |
| NotYet: a value of type UnaryExpression & NumericLiteral | 1 |
| NotYet: a value of type V | 1 |
| NotYet: a value of type VariableDeclaration & { name: Identifier; } | 1 |
| NotYet: a value of type VariableDeclaration & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 |
| NotYet: a value of type VariableDeclarationList & { _usingBrand: void; } | 1 |
| NotYet: a value of type WatchFactoryHost & { trace?(s: string): void; } | 1 |
| NotYet: a value of type __String & string | 1 |
| NotYet: a value of type boolean \| V \| undefined | 1 |
| NotYet: a value of type readonly (readonly T[])[] | 1 |
| NotYet: a value of type readonly IncrementalBundleEmitBuildInfoFileInfo[] \| readonly IncrementalMultiFileEmitBuildInfoFileInfo[] where an array goes | 1 |
| NotYet: a value of type readonly K[] | 1 |
| NotYet: a value of type string \| null | 1 |
| NotYet: a value of type string \| object \| undefined | 1 |
| NotYet: a value of type { forEach: (callbackfn: (value: T, key: K, map: Map<K, T>) => void, thisArg?: any) => void; clear: () => void; } | 1 |
| NotYet: an array of CanonicalKey | 1 |
| NotYet: an array of Path | 1 |
| NotYet: an array of T \| U | 1 |
| NotYet: an array of TState | 1 |
| NotYet: an array of object | 1 |
| NotYet: an array of unknown | 1 |
| NotYet: destructuring a string | 1 |
| NotYet: for...in over an array (holes and own enumerable properties are not represented; use for...of for elements) | 1 |
| NotYet: new Map from something that isn't [key, value] pairs | 1 |
| NotYet: passing boolean \| undefined to a function value | 1 |
| NotYet: reading absoluteSourceFilePath | 1 |
| NotYet: reading abstractSignatures | 1 |
| NotYet: reading accessibleSymbolsFromExports | 1 |
| NotYet: reading accessorType | 1 |
| NotYet: reading activeLabel | 1 |
| NotYet: reading actualFileName | 1 |
| NotYet: reading addAggregateStatistic | 1 |
| NotYet: reading addOutput | 1 |
| NotYet: reading afterImportPos | 1 |
| NotYet: reading afterImportTagPos | 1 |
| NotYet: reading aliasDecl | 1 |
| NotYet: reading allSetOptions | 1 |
| NotYet: reading allowEmpty | 1 |
| NotYet: reading allowsStrings | 1 |
| NotYet: reading alreadyTransformed | 1 |
| NotYet: reading alternateResult | 1 |
| NotYet: reading alternateResultMessage | 1 |
| NotYet: reading ambientModuleDeclare | 1 |
| NotYet: reading ancestor | 1 |
| NotYet: reading annotatedNodes | 1 |
| NotYet: reading annotationSymbol | 1 |
| NotYet: reading applicableByArity | 1 |
| NotYet: reading arrayType | 1 |
| NotYet: reading arrowExpression | 1 |
| NotYet: reading attrs | 1 |
| NotYet: reading autoGenerateFlags | 1 |
| NotYet: reading awaitToken | 1 |
| NotYet: reading baseConstraint | 1 |
| NotYet: reading baseConstructorType | 1 |
| NotYet: reading baseName | 1 |
| NotYet: reading baseType | 1 |
| NotYet: reading baseTypeNodes | 1 |
| NotYet: reading bases | 1 |
| NotYet: reading bestMatchingType | 1 |
| NotYet: reading binaryExpression | 1 |
| NotYet: reading bindingElement | 1 |
| NotYet: reading bindingList | 1 |
| NotYet: reading bindingTarget | 1 |
| NotYet: reading bold | 1 |
| NotYet: reading build | 1 |
| NotYet: reading buildArray | 1 |
| NotYet: reading buildOrderFromState | 1 |
| NotYet: reading builtins | 1 |
| NotYet: reading cachedDiagnostics | 1 |
| NotYet: reading cachedResolvedSignatures | 1 |
| NotYet: reading callArgument | 1 |
| NotYet: reading callSignatures | 1 |
| NotYet: reading callTarget | 1 |
| NotYet: reading canExcludeDiscriminants | 1 |
| NotYet: reading canUseBreakOrContinue | 1 |
| NotYet: reading candidateDirectories | 1 |
| NotYet: reading candidateExists | 1 |
| NotYet: reading caseType | 1 |
| NotYet: reading ch1 | 1 |
| NotYet: reading ch2 | 1 |
| NotYet: reading ch3 | 1 |
| NotYet: reading chain1 | 1 |
| NotYet: reading changed | 1 |
| NotYet: reading charCode | 1 |
| NotYet: reading checkAttributesType | 1 |
| NotYet: reading checkBody | 1 |
| NotYet: reading checkTypeDeferred | 1 |
| NotYet: reading childFieldType | 1 |
| NotYet: reading childPropName | 1 |
| NotYet: reading childrenPropName | 1 |
| NotYet: reading classInstanceType | 1 |
| NotYet: reading className | 1 |
| NotYet: reading classSymbol | 1 |
| NotYet: reading classThis | 1 |
| NotYet: reading cloned | 1 |
| NotYet: reading closingLineTerminatorCount | 1 |
| NotYet: reading code2 | 1 |
| NotYet: reading collidingSymbol | 1 |
| NotYet: reading combined | 1 |
| NotYet: reading commentEnd | 1 |
| NotYet: reading commentText | 1 |
| NotYet: reading commonJSPropertyAccess | 1 |
| NotYet: reading commonResolved | 1 |
| NotYet: reading compareResult | 1 |
| NotYet: reading comparer | 1 |
| NotYet: reading compilerFilePath | 1 |
| NotYet: reading compilerOptionsProperty | 1 |
| NotYet: reading componentEqualityComparer | 1 |
| NotYet: reading computedPropertyName | 1 |
| NotYet: reading conditional | 1 |
| NotYet: reading configFileText | 1 |
| NotYet: reading constantValue | 1 |
| NotYet: reading constraintDeclaration | 1 |
| NotYet: reading constraintNode | 1 |
| NotYet: reading constructorDeclaration | 1 |
| NotYet: reading constructorFunction | 1 |
| NotYet: reading constructorLikeName | 1 |
| NotYet: reading constructorSymbol | 1 |
| NotYet: reading constructorTypeCount | 1 |
| NotYet: reading containerObjectType | 1 |
| NotYet: reading containingClassDecl | 1 |
| NotYet: reading containingMethod | 1 |
| NotYet: reading containingSourceFile | 1 |
| NotYet: reading contextSpecifier | 1 |
| NotYet: reading contextType | 1 |
| NotYet: reading contextualAwaitedType | 1 |
| NotYet: reading convertToFunctionBlock | 1 |
| NotYet: reading converters | 1 |
| NotYet: reading create | 1 |
| NotYet: reading createBaseSourceFileNode | 1 |
| NotYet: reading createNodeArray | 1 |
| NotYet: reading currentDetachedCommentInfo | 1 |
| NotYet: reading currentOptions | 1 |
| NotYet: reading currentWriterIndentSpacing | 1 |
| NotYet: reading customTransformer | 1 |
| NotYet: reading d | 1 |
| NotYet: reading data | 1 |
| NotYet: reading declarationName | 1 |
| NotYet: reading declaringClassDeclaration | 1 |
| NotYet: reading defaultType | 1 |
| NotYet: reading delim | 1 |
| NotYet: reading deprecatedTag | 1 |
| NotYet: reading diagnosticMessage | 1 |
| NotYet: reading diagnosticWithLocation | 1 |
| NotYet: reading dir | 1 |
| NotYet: reading dirPath | 1 |
| NotYet: reading discriminantCombinations | 1 |
| NotYet: reading discriminantType | 1 |
| NotYet: reading disposeScope | 1 |
| NotYet: reading dist | 1 |
| NotYet: reading doneType | 1 |
| NotYet: reading downlevelIteration | 1 |
| NotYet: reading downleveledImport | 1 |
| NotYet: reading effectiveExpr | 1 |
| NotYet: reading elementAccess | 1 |
| NotYet: reading elementFlags | 1 |
| NotYet: reading elements | 1 |
| NotYet: reading emitAsSingleStatement | 1 |
| NotYet: reading emitComments | 1 |
| NotYet: reading emitExplicitInitializer | 1 |
| NotYet: reading emitSourceMaps | 1 |
| NotYet: reading emitTrailingComma | 1 |
| NotYet: reading emittedAsTopLevel | 1 |
| NotYet: reading enclosingBlockScopeContainer | 1 |
| NotYet: reading enclosingClass | 1 |
| NotYet: reading enclosingContainer | 1 |
| NotYet: reading encodeURI | 1 |
| NotYet: reading end | 1 |
| NotYet: reading enqueue | 1 |
| NotYet: reading entry | 1 |
| NotYet: reading enumResult | 1 |
| NotYet: reading enumStatement | 1 |
| NotYet: reading envVarStatement | 1 |
| NotYet: reading equalityComparer | 1 |
| NotYet: reading errNode | 1 |
| NotYet: reading errorMessage | 1 |
| NotYet: reading errorSpan | 1 |
| NotYet: reading evaluateEntityNameExpression | 1 |
| NotYet: reading everyClauseChecks | 1 |
| NotYet: reading exclamationToken | 1 |
| NotYet: reading excludePattern | 1 |
| NotYet: reading excludeRe | 1 |
| NotYet: reading excludedProperties | 1 |
| NotYet: reading existingPending | 1 |
| NotYet: reading existingProp | 1 |
| NotYet: reading existingSpecifier | 1 |
| NotYet: reading existingTarget | 1 |
| NotYet: reading exit | 1 |
| NotYet: reading exitStatus | 1 |
| NotYet: reading exportClause | 1 |
| NotYet: reading exportContainer | 1 |
| NotYet: reading exportStars | 1 |
| NotYet: reading exported | 1 |
| NotYet: reading exportedNamesStorageRef | 1 |
| NotYet: reading expressions | 1 |
| NotYet: reading ext | 1 |
| NotYet: reading extendsType | 1 |
| NotYet: reading externalHelpersImportDeclaration | 1 |
| NotYet: reading externalHelpersModuleName | 1 |
| NotYet: reading externalHelpersModuleReference | 1 |
| NotYet: reading f1 | 1 |
| NotYet: reading factory | 1 |
| NotYet: reading facts | 1 |
| NotYet: reading failed | 1 |
| NotYet: reading failedSignatureDeclarations | 1 |
| NotYet: reading falseSubtype | 1 |
| NotYet: reading fileDiags | 1 |
| NotYet: reading fileExists | 1 |
| NotYet: reading fileName | 1 |
| NotYet: reading fileToErrorCount | 1 |
| NotYet: reading filesForEmit | 1 |
| NotYet: reading filesInError | 1 |
| NotYet: reading filesToDelete | 1 |
| NotYet: reading filtered | 1 |
| NotYet: reading finalizeBoundary | 1 |
| NotYet: reading finished | 1 |
| NotYet: reading firstAccessor | 1 |
| NotYet: reading firstChar | 1 |
| NotYet: reading firstComponent | 1 |
| NotYet: reading firstInterfaceDecl | 1 |
| NotYet: reading firstNonzeroSegment | 1 |
| NotYet: reading firstRelevantLocation | 1 |
| NotYet: reading firstStatement | 1 |
| NotYet: reading firstVariableMatch | 1 |
| NotYet: reading fixed | 1 |
| NotYet: reading following | 1 |
| NotYet: reading forInitializer | 1 |
| NotYet: reading forStatement | 1 |
| NotYet: reading forcedLookupLocation | 1 |
| NotYet: reading format | 1 |
| NotYet: reading fromComponent | 1 |
| NotYet: reading fromNameType | 1 |
| NotYet: reading functionLocation | 1 |
| NotYet: reading functionType | 1 |
| NotYet: reading generatedLine | 1 |
| NotYet: reading generatedName | 1 |
| NotYet: reading generatorFunc | 1 |
| NotYet: reading generatorInstantiation | 1 |
| NotYet: reading genericDiag | 1 |
| NotYet: reading get | 1 |
| NotYet: reading getAccessorType | 1 |
| NotYet: reading getFromDirectoryCache | 1 |
| NotYet: reading getFromNonRelativeNameCache | 1 |
| NotYet: reading getFunc | 1 |
| NotYet: reading getMapOfCacheRedirects | 1 |
| NotYet: reading getPackageJsonInfo | 1 |
| NotYet: reading getParenthesizeLeftSideOfBinaryForOperator | 1 |
| NotYet: reading getSourcePosition | 1 |
| NotYet: reading getUnscopedHelperName | 1 |
| NotYet: reading getUnusedExpectations | 1 |
| NotYet: reading globalCache | 1 |
| NotYet: reading hasEmptyObject | 1 |
| NotYet: reading hasExistingReasonToReportErrorOn | 1 |
| NotYet: reading hasInstanceProperty | 1 |
| NotYet: reading hasJSDocFunctionType | 1 |
| NotYet: reading hasLeadingModifier | 1 |
| NotYet: reading hasPrivateModifier | 1 |
| NotYet: reading hasSignatures | 1 |
| NotYet: reading hasTrailingDecorator | 1 |
| NotYet: reading hasTransformableStatics | 1 |
| NotYet: reading headerPadding | 1 |
| NotYet: reading helper | 1 |
| NotYet: reading helpers | 1 |
| NotYet: reading hostSourceFileInfo | 1 |
| NotYet: reading i | 1 |
| NotYet: reading ids | 1 |
| NotYet: reading ifStatement | 1 |
| NotYet: reading iife | 1 |
| NotYet: reading illegalContextMessage | 1 |
| NotYet: reading immediate | 1 |
| NotYet: reading immediateDeclaration | 1 |
| NotYet: reading implDecl | 1 |
| NotYet: reading implementationSharesContainerWithFirstOverload | 1 |
| NotYet: reading impliedNodeFormat | 1 |
| NotYet: reading importClause | 1 |
| NotYet: reading importDecl | 1 |
| NotYet: reading importSource | 1 |
| NotYet: reading imports | 1 |
| NotYet: reading includeFileRegexes | 1 |
| NotYet: reading includeRe | 1 |
| NotYet: reading indent | 1 |
| NotYet: reading indexedAccessType | 1 |
| NotYet: reading inferences | 1 |
| NotYet: reading initialLocationForSecondaryLookup | 1 |
| NotYet: reading initialType | 1 |
| NotYet: reading initializerWithoutParens | 1 |
| NotYet: reading initializersName | 1 |
| NotYet: reading inlinable | 1 |
| NotYet: reading innerIndexType | 1 |
| NotYet: reading innerMappedType | 1 |
| NotYet: reading instantiatedSignature | 1 |
| NotYet: reading instantiatedTemplateType | 1 |
| NotYet: reading instantiations | 1 |
| NotYet: reading internalFlags | 1 |
| NotYet: reading intrinsicAttribs | 1 |
| NotYet: reading intrinsicElementsType | 1 |
| NotYet: reading intrinsics | 1 |
| NotYet: reading invalidElement | 1 |
| NotYet: reading isAbstract | 1 |
| NotYet: reading isAnonymous | 1 |
| NotYet: reading isArrowFunctionInJsx | 1 |
| NotYet: reading isCallToReadHelper | 1 |
| NotYet: reading isCallbackTag | 1 |
| NotYet: reading isCapturedInFunction | 1 |
| NotYet: reading isClassWithConstructorReference | 1 |
| NotYet: reading isComparingJsxAttributes | 1 |
| NotYet: reading isConfigIdentical | 1 |
| NotYet: reading isDerivedClass | 1 |
| NotYet: reading isDosStyle | 1 |
| NotYet: reading isEitherEnum | 1 |
| NotYet: reading isExportEquals | 1 |
| NotYet: reading isIllegalExportDefaultInCJS | 1 |
| NotYet: reading isInExternalModule | 1 |
| NotYet: reading isJSDoc | 1 |
| NotYet: reading isJavaScript | 1 |
| NotYet: reading isKnownProperty | 1 |
| NotYet: reading isLengthPushOrUnshift | 1 |
| NotYet: reading isLongestMatchingPrefix | 1 |
| NotYet: reading isMarkdownOrJSDocLink | 1 |
| NotYet: reading isNameFirst | 1 |
| NotYet: reading isOptional | 1 |
| NotYet: reading isOverload | 1 |
| NotYet: reading isParameter | 1 |
| NotYet: reading isPlainJs | 1 |
| NotYet: reading isPromise | 1 |
| NotYet: reading isPropertyName | 1 |
| NotYet: reading isReservedWord | 1 |
| NotYet: reading isRest | 1 |
| NotYet: reading isSimpleLoop | 1 |
| NotYet: reading isStaticMethodSymbol | 1 |
| NotYet: reading isTypeOnly | 1 |
| NotYet: reading isValid | 1 |
| NotYet: reading isValue | 1 |
| NotYet: reading issuedDiagnostic | 1 |
| NotYet: reading iterator | 1 |
| NotYet: reading iteratorValueStatement | 1 |
| NotYet: reading jsDoc | 1 |
| NotYet: reading jsDocType | 1 |
| NotYet: reading jsxChildrenPropertyName | 1 |
| NotYet: reading jsxSpecific | 1 |
| NotYet: reading keyAttr | 1 |
| NotYet: reading laneCount | 1 |
| NotYet: reading lastChild | 1 |
| NotYet: reading lastElement | 1 |
| NotYet: reading lastJSDocParam | 1 |
| NotYet: reading lastPart | 1 |
| NotYet: reading lastSpan | 1 |
| NotYet: reading lastStatement | 1 |
| NotYet: reading leadingComments | 1 |
| NotYet: reading leadingLineTerminatorCount | 1 |
| NotYet: reading leftColumnHeadingLength | 1 |
| NotYet: reading leftIsNumeric | 1 |
| NotYet: reading leftTarget | 1 |
| NotYet: reading leftType | 1 |
| NotYet: reading leftmost | 1 |
| NotYet: reading lex | 1 |
| NotYet: reading limitedConstraint | 1 |
| NotYet: reading linesAfterDot | 1 |
| NotYet: reading literal | 1 |
| NotYet: reading literalValue | 1 |
| NotYet: reading literals | 1 |
| NotYet: reading localCheckDeclaration | 1 |
| NotYet: reading localIndexDeclaration | 1 |
| NotYet: reading localName | 1 |
| NotYet: reading localSymbol | 1 |
| NotYet: reading location | 1 |
| NotYet: reading mainExport | 1 |
| NotYet: reading mappedTypeNode | 1 |
| NotYet: reading mappedTypeVariable | 1 |
| NotYet: reading mapper | 1 |
| NotYet: reading mappings | 1 |
| NotYet: reading matchType | 1 |
| NotYet: reading maxErrors | 1 |
| NotYet: reading maxLength | 1 |
| NotYet: reading maxNonRestParam | 1 |
| NotYet: reading mayHaveNameCollisions | 1 |
| NotYet: reading memberProps | 1 |
| NotYet: reading metadataReference | 1 |
| NotYet: reading methodReturnType | 1 |
| NotYet: reading missingNode | 1 |
| NotYet: reading missingPaths | 1 |
| NotYet: reading mixinFlags | 1 |
| NotYet: reading mod | 1 |
| NotYet: reading modifierArray | 1 |
| NotYet: reading moduleBlock | 1 |
| NotYet: reading moduleStatement | 1 |
| NotYet: reading nameParts | 1 |
| NotYet: reading nameStr | 1 |
| NotYet: reading nameText | 1 |
| NotYet: reading namedBindings | 1 |
| NotYet: reading namespaceDeclaration | 1 |
| NotYet: reading narrowedType | 1 |
| NotYet: reading needCheckInitializer | 1 |
| NotYet: reading needJsExtensions | 1 |
| NotYet: reading needSyncEval | 1 |
| NotYet: reading needsModifierPreservingWrapper | 1 |
| NotYet: reading needsName | 1 |
| NotYet: reading needsOutParam | 1 |
| NotYet: reading needsUpdateInTypeRootWatch | 1 |
| NotYet: reading negative | 1 |
| NotYet: reading newElements | 1 |
| NotYet: reading newEndN | 1 |
| NotYet: reading newItem | 1 |
| NotYet: reading newName | 1 |
| NotYet: reading newParam | 1 |
| NotYet: reading newParams | 1 |
| NotYet: reading newSourceFile | 1 |
| NotYet: reading nodeConstructors | 1 |
| NotYet: reading nodeContextFlags | 1 |
| NotYet: reading nodeInAmbientContext | 1 |
| NotYet: reading nodeModulesDirectoryName | 1 |
| NotYet: reading nodeModulesFolder | 1 |
| NotYet: reading nodeModulesFolderExists | 1 |
| NotYet: reading nodeName | 1 |
| NotYet: reading nodes | 1 |
| NotYet: reading normalizedElements | 1 |
| NotYet: reading ns | 1 |
| NotYet: reading nullishSemantics | 1 |
| NotYet: reading numNodes | 1 |
| NotYet: reading numParameters | 1 |
| NotYet: reading numericValue | 1 |
| NotYet: reading objectLiterals | 1 |
| NotYet: reading objectProperties | 1 |
| NotYet: reading offset | 1 |
| NotYet: reading ok | 1 |
| NotYet: reading oldEnclosing | 1 |
| NotYet: reading oldEndN | 1 |
| NotYet: reading oldSignature | 1 |
| NotYet: reading oldSourceFiles | 1 |
| NotYet: reading oldStart2 | 1 |
| NotYet: reading oldTime | 1 |
| NotYet: reading onProgramCreateComplete | 1 |
| NotYet: reading onWatchStatusChange | 1 |
| NotYet: reading openBracketPosition | 1 |
| NotYet: reading operand | 1 |
| NotYet: reading optional | 1 |
| NotYet: reading optionalDeclaration | 1 |
| NotYet: reading optionsOfCurCategory | 1 |
| NotYet: reading optionsType | 1 |
| NotYet: reading originalClass | 1 |
| NotYet: reading originalCreateDirectory | 1 |
| NotYet: reading originalFile | 1 |
| NotYet: reading originalModuleSpecifier | 1 |
| NotYet: reading originalReadFile | 1 |
| NotYet: reading other | 1 |
| NotYet: reading otherFiles | 1 |
| NotYet: reading otherOption | 1 |
| NotYet: reading outPath | 1 |
| NotYet: reading outerTypeParameters | 1 |
| NotYet: reading outputDir | 1 |
| NotYet: reading overloadSignatures | 1 |
| NotYet: reading ownKey | 1 |
| NotYet: reading ownKeys | 1 |
| NotYet: reading ownMap | 1 |
| NotYet: reading ownOutputFilePath | 1 |
| NotYet: reading packageFileResult | 1 |
| NotYet: reading packageJsonMap | 1 |
| NotYet: reading packageName | 1 |
| NotYet: reading packageResult | 1 |
| NotYet: reading paramIdent | 1 |
| NotYet: reading parameterDeclaration | 1 |
| NotYet: reading parameterIndex | 1 |
| NotYet: reading parameterNode | 1 |
| NotYet: reading parameterRange | 1 |
| NotYet: reading parameterSymbol | 1 |
| NotYet: reading parentDeclaration | 1 |
| NotYet: reading parenthesizerRules | 1 |
| NotYet: reading pathAndExtension | 1 |
| NotYet: reading pathList | 1 |
| NotYet: reading pathToTopLevelNodeModules | 1 |
| NotYet: reading pattern | 1 |
| NotYet: reading peerDependencies | 1 |
| NotYet: reading pendingExpressions | 1 |
| NotYet: reading pendingKind | 1 |
| NotYet: reading possibleOption | 1 |
| NotYet: reading possibleOutOfBounds | 1 |
| NotYet: reading possiblyOutOfBoundsType | 1 |
| NotYet: reading precedingLineBreak | 1 |
| NotYet: reading prerelease | 1 |
| NotYet: reading prereleaseArray | 1 |
| NotYet: reading prevNodeIndex | 1 |
| NotYet: reading previousDuration | 1 |
| NotYet: reading primaryDeclaration | 1 |
| NotYet: reading primitive | 1 |
| NotYet: reading printNode | 1 |
| NotYet: reading printerOptions | 1 |
| NotYet: reading priority | 1 |
| NotYet: reading programDiagnosticsInFile | 1 |
| NotYet: reading prologueStatementCount | 1 |
| NotYet: reading promise | 1 |
| NotYet: reading propContext | 1 |
| NotYet: reading propNameType | 1 |
| NotYet: reading propNode | 1 |
| NotYet: reading propertyAccess | 1 |
| NotYet: reading propertyAssignmentType | 1 |
| NotYet: reading propertyTypes | 1 |
| NotYet: reading proto | 1 |
| NotYet: reading prototype | 1 |
| NotYet: reading prototypePropertyType | 1 |
| NotYet: reading prototypeSymbol | 1 |
| NotYet: reading prototypeType | 1 |
| NotYet: reading qualifiedName | 1 |
| NotYet: reading quick | 1 |
| NotYet: reading quickResult | 1 |
| NotYet: reading rawName | 1 |
| NotYet: reading rawSources | 1 |
| NotYet: reading react | 1 |
| NotYet: reading reactExports | 1 |
| NotYet: reading readExpression | 1 |
| NotYet: reading readFileWithCache | 1 |
| NotYet: reading readonlyMask | 1 |
| NotYet: reading realDeclarationPath | 1 |
| NotYet: reading recursionIdentity | 1 |
| NotYet: reading reducedType | 1 |
| NotYet: reading reexports | 1 |
| NotYet: reading referencedFileName | 1 |
| NotYet: reading referencedMap | 1 |
| NotYet: reading relativePath | 1 |
| NotYet: reading relevantTypeParameterConstraints | 1 |
| NotYet: reading removeNullable | 1 |
| NotYet: reading removeUndefined | 1 |
| NotYet: reading rename | 1 |
| NotYet: reading repopulatedChain | 1 |
| NotYet: reading reportErrors | 1 |
| NotYet: reading requiresAddingUndefined | 1 |
| NotYet: reading resolutionDiagnostic | 1 |
| NotYet: reading resolutionMode | 1 |
| NotYet: reading resolutionsChanged | 1 |
| NotYet: reading resolvedFromFile | 1 |
| NotYet: reading resolvedMethodReturnType | 1 |
| NotYet: reading resolvedModuleNames | 1 |
| NotYet: reading resolvedModuleSymbol | 1 |
| NotYet: reading resolvedValueSymbol | 1 |
| NotYet: reading resolver | 1 |
| NotYet: reading rest | 1 |
| NotYet: reading restParameterSymbols | 1 |
| NotYet: reading restSymbol | 1 |
| NotYet: reading resultFromDts | 1 |
| NotYet: reading resultType | 1 |
| NotYet: reading results | 1 |
| NotYet: reading returnOrPromisedType | 1 |
| NotYet: reading returnStatement | 1 |
| NotYet: reading returnType | 1 |
| NotYet: reading right | 1 |
| NotYet: reading rightCustomPrologueEnd | 1 |
| NotYet: reading rightHoistedFunctionsEnd | 1 |
| NotYet: reading rightHoistedVariablesEnd | 1 |
| NotYet: reading rightIdentifier | 1 |
| NotYet: reading rightStandardPrologueEnd | 1 |
| NotYet: reading rootExpr | 1 |
| NotYet: reading rootPathComponents | 1 |
| NotYet: reading rootResult | 1 |
| NotYet: reading rootSymbol | 1 |
| NotYet: reading runtimeImportSpecifier | 1 |
| NotYet: reading savedInStrictMode | 1 |
| NotYet: reading savedPreserveSourceNewlines | 1 |
| NotYet: reading scanner | 1 |
| NotYet: reading secondType | 1 |
| NotYet: reading seenNames | 1 |
| NotYet: reading segment | 1 |
| NotYet: reading segments | 1 |
| NotYet: reading separateBeginAndEnd | 1 |
| NotYet: reading serializeTypeOfDeclaration | 1 |
| NotYet: reading serializedName | 1 |
| NotYet: reading setFunc | 1 |
| NotYet: reading setProp | 1 |
| NotYet: reading setReadFileCache | 1 |
| NotYet: reading setterModifiers | 1 |
| NotYet: reading setterType | 1 |
| NotYet: reading shebang | 1 |
| NotYet: reading shouldConvertCondition | 1 |
| NotYet: reading shouldEmitComments | 1 |
| NotYet: reading shouldEmitDetachedComment | 1 |
| NotYet: reading shouldEmitDotDot | 1 |
| NotYet: reading shouldEmitSourceMaps | 1 |
| NotYet: reading shouldResolveAlias | 1 |
| NotYet: reading shouldResolveFactoryReference | 1 |
| NotYet: reading shouldTransformInitializers | 1 |
| NotYet: reading shouldTransformInitializersUsingSet | 1 |
| NotYet: reading shouldTransformThisInStaticInitializers | 1 |
| NotYet: reading shouldWriteNativeEvents | 1 |
| NotYet: reading sigRestType | 1 |
| NotYet: reading signatureNextType | 1 |
| NotYet: reading signatures | 1 |
| NotYet: reading singleLine | 1 |
| NotYet: reading singleQuote | 1 |
| NotYet: reading skipBindingPatterns | 1 |
| NotYet: reading snippetElement | 1 |
| NotYet: reading sortedIndex | 1 |
| NotYet: reading sourceConstraint | 1 |
| NotYet: reading sourceDiscriminantTypes | 1 |
| NotYet: reading sourceEnd | 1 |
| NotYet: reading sourceEndText | 1 |
| NotYet: reading sourceFileAbsolutePaths | 1 |
| NotYet: reading sourceFileNoExtension | 1 |
| NotYet: reading sourceFiles | 1 |
| NotYet: reading sourceHasBase | 1 |
| NotYet: reading sourceHasMoreParameters | 1 |
| NotYet: reading sourceInfos | 1 |
| NotYet: reading sourceIsJSConstructor | 1 |
| NotYet: reading sourceMapRange | 1 |
| NotYet: reading sourceMapUrlPos | 1 |
| NotYet: reading sourceMappingURL | 1 |
| NotYet: reading sourceProp | 1 |
| NotYet: reading sourceProperty | 1 |
| NotYet: reading sourceSignature | 1 |
| NotYet: reading sourceSignatures | 1 |
| NotYet: reading sourceStartText | 1 |
| NotYet: reading space | 1 |
| NotYet: reading spacesToEmit | 1 |
| NotYet: reading spec | 1 |
| NotYet: reading specialPropertyAssignmentKind | 1 |
| NotYet: reading specifierType | 1 |
| NotYet: reading specifiers | 1 |
| NotYet: reading spreadElement | 1 |
| NotYet: reading spreadIndex | 1 |
| NotYet: reading startPos | 1 |
| NotYet: reading startsOnNewLine | 1 |
| NotYet: reading stat | 1 |
| NotYet: reading statementExpression | 1 |
| NotYet: reading statementOffset | 1 |
| NotYet: reading staticBlock | 1 |
| NotYet: reading strName | 1 |
| NotYet: reading subsequentNode | 1 |
| NotYet: reading substituteConstraints | 1 |
| NotYet: reading superCall | 1 |
| NotYet: reading superPath | 1 |
| NotYet: reading superStatement | 1 |
| NotYet: reading symbolExport | 1 |
| NotYet: reading symbols | 1 |
| NotYet: reading syntacticBuilderResolver | 1 |
| NotYet: reading syntheticArgsSymbol | 1 |
| NotYet: reading system | 1 |
| NotYet: reading tagExpression | 1 |
| NotYet: reading tags | 1 |
| NotYet: reading targetDeclarationKind | 1 |
| NotYet: reading targetDepth | 1 |
| NotYet: reading targetFile | 1 |
| NotYet: reading targetHasBase | 1 |
| NotYet: reading targetIsJSConstructor | 1 |
| NotYet: reading targetIsOptional | 1 |
| NotYet: reading targetParam | 1 |
| NotYet: reading targetReturn | 1 |
| NotYet: reading targetStartText | 1 |
| NotYet: reading targetSymbolFile | 1 |
| NotYet: reading targetType | 1 |
| NotYet: reading tempVar | 1 |
| NotYet: reading templates | 1 |
| NotYet: reading thenFunction | 1 |
| NotYet: reading thisParam | 1 |
| NotYet: reading thisParameter | 1 |
| NotYet: reading thisParameters | 1 |
| NotYet: reading throwDiagnostic | 1 |
| NotYet: reading timerToInvalidateFailedLookupResolutions | 1 |
| NotYet: reading timerToUpdateProgram | 1 |
| NotYet: reading tok | 1 |
| NotYet: reading tokenSourceMapRanges | 1 |
| NotYet: reading tokenString | 1 |
| NotYet: reading tokenText | 1 |
| NotYet: reading tp | 1 |
| NotYet: reading trackSymbol | 1 |
| NotYet: reading trailingComments | 1 |
| NotYet: reading trailingNewlines | 1 |
| NotYet: reading trailingParts | 1 |
| NotYet: reading trampoline | 1 |
| NotYet: reading transform | 1 |
| NotYet: reading tripleSlash | 1 |
| NotYet: reading trueCondition | 1 |
| NotYet: reading tryStatement | 1 |
| NotYet: reading tsPriority | 1 |
| NotYet: reading tsconfigTime | 1 |
| NotYet: reading tupleTarget | 1 |
| NotYet: reading tupleType | 1 |
| NotYet: reading typeAlias | 1 |
| NotYet: reading typeArgumentTypes | 1 |
| NotYet: reading typeIsAutomatic | 1 |
| NotYet: reading typeKey | 1 |
| NotYet: reading typeKind | 1 |
| NotYet: reading typeOfArrayLiteral | 1 |
| NotYet: reading typeOfObjectLiteral | 1 |
| NotYet: reading typeOnlyDeclarationIsExportStar | 1 |
| NotYet: reading typeOrConstraint | 1 |
| NotYet: reading typePredicateVariable | 1 |
| NotYet: reading typeReferenceResolutionsChanged | 1 |
| NotYet: reading typeRoots | 1 |
| NotYet: reading typeVariable | 1 |
| NotYet: reading typescriptVersion | 1 |
| NotYet: reading undefinedStrippedTarget | 1 |
| NotYet: reading unionType | 1 |
| NotYet: reading uniqueFilled | 1 |
| NotYet: reading unwidenedType | 1 |
| NotYet: reading unwrappedExprType | 1 |
| NotYet: reading updatedText | 1 |
| NotYet: reading usageMode | 1 |
| NotYet: reading useCaseSensitiveFileNames | 1 |
| NotYet: reading useStrictDirective | 1 |
| NotYet: reading val | 1 |
| NotYet: reading valid | 1 |
| NotYet: reading validatedFilesSpec | 1 |
| NotYet: reading validatedFilesSpecBeforeSubstitution | 1 |
| NotYet: reading valueParam | 1 |
| NotYet: reading values | 1 |
| NotYet: reading variableDeclarator | 1 |
| NotYet: reading variableList | 1 |
| NotYet: reading variance | 1 |
| NotYet: reading varianceFlags | 1 |
| NotYet: reading version | 1 |
| NotYet: reading visibilityResult | 1 |
| NotYet: reading visibleDefaultBinding | 1 |
| NotYet: reading visitedNode | 1 |
| NotYet: reading voidIsNonOptional | 1 |
| NotYet: reading watchDirectoryKind | 1 |
| NotYet: reading watchFileKind | 1 |
| NotYet: reading write | 1 |
| NotYet: reading writeText | 1 |
| NotYet: spreading an array of other elements | 1 |
| NotYet: storing any in a field | 1 |
| NotYet: storing false \| Type in a field | 1 |
| NotYet: storing string \| number in a field | 1 |
| Refused: &&= | 1 |
| Refused: JSON.parse: its result's type can't be proven from the text | 1 |
| Refused: Object.defineProperty | 1 |
| Refused: Object.setPrototypeOf | 1 |
| Refused: a constructor object escaping before static fields are initialized | 1 |
| Refused: a function taking (symbol: Symbol) => boolean seen as one taking ((symbol: Symbol) => boolean) \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking BinaryOperatorToken seen as one taking BinaryOperatorToken \| BinaryOperator (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking JSDocTypeExpression \| undefined seen as one taking JSDocTypeExpression \| JSDocTypeLiteral \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking LogLevel seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking Node seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking Node \| undefined seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking NodeArray<Expression> seen as one taking readonly Expression[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking NodeArray<T> seen as one taking NodeArray<T> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking NodeArray<T> seen as one taking NodeArray<T> \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking NodeArray<TypeNode> \| undefined seen as one taking readonly TypeNode[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking OuterExpressionKinds seen as one taking OuterExpressionKinds \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking SourceFile seen as one taking [file: SourceFile, options: CompilerOptions] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking T seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking TIn seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking TokenFlags seen as one taking TokenFlags \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking TypeFacts.None seen as one taking number (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking TypeNode seen as one taking Node (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [callback: (...args: any[]) => void, ms: number, ...args: any[]] seen as one taking (...args: any[]) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [fileName: string] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [node: ConstructorTypeNode, typeParameters: NodeArray<TypeParameterDeclaration> \| undefined, parameters: NodeArray<ParameterDeclaration>, type: TypeNode] \| ... seen as one taking ConstructorTypeNode (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [node: Node] seen as one taking BindingElement \| OmittedExpression (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [node: Node] seen as one taking Declaration (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [path: string, callback: DirectoryWatcherCallback, recursive?: boolean \| undefined, options?: WatchOptions \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [path: string, callback: FileWatcherCallback, pollingInterval?: number \| undefined, options?: WatchOptions \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [timeoutId: any] seen as one taking unknown (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [typeParameters: readonly TypeParameterDeclaration[] \| undefined, parameters: readonly ParameterDeclaration[], type: TypeNode] \| [modifiers: readonly Modifier[] \| undefined, typeParameters: ... \| undefined, parameters: ..., type: TypeNode] seen as one taking readonly Modifier[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking [writeByteOrderMark: boolean, onError?: ((message: string) => void) \| undefined, sourceFiles?: readonly SourceFile[] \| undefined, data?: WriteFileCallbackData \| undefined] seen as one taking boolean (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking never seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking number seen as one taking any[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking number seen as one taking number \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking readonly ParameterDeclaration[] seen as one taking readonly ParameterDeclaration[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking readonly T[] seen as one taking readonly T[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking readonly T[] \| undefined seen as one taking readonly T[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a function taking string seen as one taking string \| MemberName (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 |
| Refused: a method read as a value (add would lose its object, and this with it) | 1 |
| Refused: a method read as a value (base64decode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (base64encode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (clearScreen would lose its object, and this with it) | 1 |
| Refused: a method read as a value (cloneNode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (compare would lose its object, and this with it) | 1 |
| Refused: a method read as a value (convertToArrayAssignmentElement would lose its object, and this with it) | 1 |
| Refused: a method read as a value (convertToObjectAssignmentElement would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createComma would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createExpressionStatement would lose its object, and this with it) | 1 |
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
| Refused: a method read as a value (createTemplateMiddle would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createTemplateTail would lose its object, and this with it) | 1 |
| Refused: a method read as a value (createUnionTypeNode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (deleteFile would lose its object, and this with it) | 1 |
| Refused: a method read as a value (emit would lose its object, and this with it) | 1 |
| Refused: a method read as a value (emitBuildInfo would lose its object, and this with it) | 1 |
| Refused: a method read as a value (emitNextAffectedFile would lose its object, and this with it) | 1 |
| Refused: a method read as a value (fill would lose its object, and this with it) | 1 |
| Refused: a method read as a value (fromCharCode would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getAllDependencies would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getDeclarationDiagnostics would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getMemoryUsage would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getModifiedTime would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getNearestAncestorDirectoryWithPackageJson would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getOrCreateCacheForModuleName would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getSemanticDiagnostics would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getSourceFileByPath would lose its object, and this with it) | 1 |
| Refused: a method read as a value (getSymlinkCache would lose its object, and this with it) | 1 |
| Refused: a method read as a value (globalCacheResolutionModuleName would lose its object, and this with it) | 1 |
| Refused: a method read as a value (hasChangedEmitSignature would lose its object, and this with it) | 1 |
| Refused: a method read as a value (hasInvalidatedLibResolutions would lose its object, and this with it) | 1 |
| Refused: a method read as a value (hasInvalidatedResolutions would lose its object, and this with it) | 1 |
| Refused: a method read as a value (hasOwnProperty would lose its object, and this with it) | 1 |
| Refused: a method read as a value (log would lose its object, and this with it) | 1 |
| Refused: a method read as a value (nonEscapingWrite would lose its object, and this with it) | 1 |
| Refused: a method read as a value (onDiscoveredSymlink would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeCheckTypeOfConditionalType would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeConciseBodyOfArrowFunction would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeConditionOfConditionalExpression would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeConstituentTypeOfIntersectionType would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeConstituentTypeOfUnionType would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeElementTypeOfTupleType would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeExpressionOfComputedPropertyName would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeExpressionOfExportDefault would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeExpressionOfExpressionStatement would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeExpressionOfNew would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeExtendsTypeOfConditionalType would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeLeadingTypeArgument would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeOperandOfPostfixUnary would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeOperandOfReadonlyTypeOperator would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeOperandOfTypeOperator would lose its object, and this with it) | 1 |
| Refused: a method read as a value (parenthesizeTypeOfOptionalType would lose its object, and this with it) | 1 |
| Refused: a method read as a value (releaseProgram would lose its object, and this with it) | 1 |
| Refused: a method read as a value (remove would lose its object, and this with it) | 1 |
| Refused: a method read as a value (repeat would lose its object, and this with it) | 1 |
| Refused: a method read as a value (replace would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportCyclicStructureError would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportInaccessibleThisError would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportInferenceFallback would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportNonSerializableProperty would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportNonlocalAugmentation would lose its object, and this with it) | 1 |
| Refused: a method read as a value (reportTruncationError would lose its object, and this with it) | 1 |
| Refused: a method read as a value (resolveModuleNameLiterals would lose its object, and this with it) | 1 |
| Refused: a method read as a value (resolveModuleNames would lose its object, and this with it) | 1 |
| Refused: a method read as a value (setModifiedTime would lose its object, and this with it) | 1 |
| Refused: a method read as a value (throwIfCancellationRequested would lose its object, and this with it) | 1 |
| Refused: a method read as a value (toString would lose its object, and this with it) | 1 |
| Refused: a method read as a value (writeOutputIsTTY would lose its object, and this with it) | 1 |
| Refused: a spread after the first field | 1 |
| Refused: a structural Object.keys view that can hide an iterable literal's symbol-key storage (adamic/symbol-key-view) | 1 |
| Refused: a type argument makes a value of type (left: MappedPosition, right: MappedPosition) => boolean seen as EqualityComparer<SourceMappedPosition> \| undefined, which can write string \| undefined where string is read | 1 |
| Refused: a type predicate whose return is not proven (asserts cond needs a boolean parameter) | 1 |
| Refused: a type predicate whose return is not proven (false return can still contain LateBoundDeclaration) | 1 |
| Refused: a type predicate whose return is not proven (normal return has not narrowed value to NonNullable<T>) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on arg) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on array) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on buildOrder) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on diagnostic) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on file) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on func) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on hostSourceFile) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on location) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on m) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on option) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on options) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on p) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on program) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on sourceFile) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on symbol) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on t) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on tagName) | 1 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on x) | 1 |
| Refused: a type predicate whose return is not proven (the body's true narrowing does not match null \| undefined) | 1 |
| Refused: a type predicate whose return is not proven (the predicate parameter is assigned) | 1 |
| Refused: a type predicate whose return is not proven (true return narrows to MappedPosition, not SourceMappedPosition) | 1 |
| Refused: a type predicate whose return is not proven (true return narrows to Mapping, not SourceMapping) | 1 |
| Refused: a type predicate whose return is not proven (true return narrows to ReusableBuilderProgramState, not BuilderProgramStateWithDefinedProgram) | 1 |
| Refused: a value of type ((node: PrivateIdentifierPropertyDeclaration, modifiers: ModifiersArray \| undefined) => ObjectLiteralExpression) \| undefined seen as ((node: PropertyDeclaration & { readonly name: PrivateIdentifier; }, modifiers: ModifiersArray \| undefined) => Expression) \| undefined, whose readonly field name becomes writable: a readonly field may hold something narrower than PrivateIdentifier, which a write of PrivateIdentifier would replace | 1 |
| Refused: a value of type () => { diagnosticMessage: DiagnosticMessage; errorNode: ExportAssignment; } seen as GetSymbolAccessibilityDiagnostic, which can write Node where ExportAssignment is read | 1 |
| Refused: a value of type (AbstractKeyword \| AccessorKeyword \| AsyncKeyword \| ConstKeyword \| DeclareKeyword \| Decorator \| ... 7 more ... \| StaticKeyword)[] \| undefined seen as Node[] \| undefined, which can write Node where AbstractKeyword \| AccessorKeyword \| AsyncKeyword \| ConstKeyword \| DeclareKeyword \| Decorator \| ... 7 more ... \| StaticKeyword is read | 1 |
| Refused: a value of type (ClassDeclaration \| EnumDeclaration \| ExportAssignment \| ExportDeclaration \| FunctionDeclaration \| ... 5 more ... \| VariableStatement)[] seen as Statement[], which can write Statement where ClassDeclaration \| EnumDeclaration \| ExportAssignment \| ExportDeclaration \| FunctionDeclaration \| ... 5 more ... \| VariableStatement is read | 1 |
| Refused: a value of type (left: MappedPosition, right: MappedPosition) => boolean seen as EqualityComparer<SourceMappedPosition> \| undefined, which can write string \| undefined where string is read | 1 |
| Refused: a value of type (node: CommentRange) => boolean seen as (value: SynthesizedComment) => boolean, which can write number where -1 is read | 1 |
| Refused: a value of type (node: PrivateIdentifierGetAccessorDeclaration, modifiers: ModifiersArray \| undefined) => ObjectLiteralExpression seen as ((node: GetAccessorDeclaration & { readonly name: PrivateIdentifier; }, modifiers: ModifiersArray \| undefined) => Expression) \| undefined, whose readonly field name becomes writable: a readonly field may hold something narrower than PrivateIdentifier, which a write of PrivateIdentifier would replace | 1 |
| Refused: a value of type (node: PrivateIdentifierMethodDeclaration, modifiers: ModifiersArray \| undefined) => ObjectLiteralExpression seen as ((node: MethodDeclaration & { readonly name: PrivateIdentifier; }, modifiers: ModifiersArray \| undefined) => Expression) \| undefined, whose readonly field name becomes writable: a readonly field may hold something narrower than PrivateIdentifier, which a write of PrivateIdentifier would replace | 1 |
| Refused: a value of type (node: PrivateIdentifierSetAccessorDeclaration, modifiers: ModifiersArray \| undefined) => ObjectLiteralExpression seen as ((node: SetAccessorDeclaration & { readonly name: PrivateIdentifier; }, modifiers: ModifiersArray \| undefined) => Expression) \| undefined, whose readonly field name becomes writable: a readonly field may hold something narrower than PrivateIdentifier, which a write of PrivateIdentifier would replace | 1 |
| Refused: a value of type (recordTempVariable: ((node: Identifier) => void) \| undefined, reservedInNestedScopes?: boolean \| undefined, prefix?: string \| GeneratedNamePart \| undefined, suffix?: string \| undefined) => GeneratedIdentifier seen as { (recordTempVariable: ((node: Identifier) => void) \| undefined, reservedInNestedScopes?: boolean \| undefined): Identifier; (recordTempVariable: ((node: Identifier) => void) \| undefined, reservedInNestedScopes?: boolean \| undefined, prefix?: string \| ... 1 more ... \| undefined, suffix?: string \| undefined): Identifi..., whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type (resolution: ResolvedModuleWithFailedLookupLocations) => ResolvedModuleFull \| undefined seen as (oldResolution: ResolvedModuleWithFailedLookupLocations) => ResolutionWithResolvedFileName \| undefined, which can write string \| undefined where string is read | 1 |
| Refused: a value of type (symbol: Symbol) => { visitedTypes: Type[]; visitedSymbols: Symbol[]; } seen as (root: Symbol) => { visitedTypes: readonly Type[]; visitedSymbols: readonly Symbol[]; }, which can write readonly Type[] where Type[] is read | 1 |
| Refused: a value of type (symbolAccessibilityResult: SymbolAccessibilityResult) => { diagnosticMessage: DiagnosticMessage; errorNode: DeclarationDiagnosticProducing; typeName: DeclarationName \| undefined; } \| undefined seen as (symbolAccessibilityResult: SymbolAccessibilityResult) => SymbolAccessibilityDiagnostic \| undefined, which can write Node where DeclarationDiagnosticProducing is read | 1 |
| Refused: a value of type (type: Type) => { visitedTypes: Type[]; visitedSymbols: Symbol[]; } seen as (root: Type) => { visitedTypes: readonly Type[]; visitedSymbols: readonly Symbol[]; }, which can write readonly Type[] where Type[] is read | 1 |
| Refused: a value of type AnonymousType seen as AnonymousType, which can write AnonymousType \| undefined where GenericType is read | 1 |
| Refused: a value of type AnonymousType \| DeferredTypeReference seen as AnonymousType, which can write AnonymousType \| undefined where GenericType is read | 1 |
| Refused: a value of type AssertClause seen as Mutable<AssertClause>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type AssertEntry seen as Mutable<AssertEntry>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type AutoAccessorPropertyDeclaration \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 |
| Refused: a value of type AwaitedTypeInstantiation \| Type seen as Type, which can write Symbol \| undefined where Symbol is read | 1 |
| Refused: a value of type BigIntLiteral \| ComputedPropertyName \| GeneratedIdentifier \| NoSubstitutionTemplateLiteral \| NumericLiteral \| PrivateIdentifier \| StringLiteral seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type BigIntLiteral \| ComputedPropertyName \| GeneratedIdentifier \| NoSubstitutionTemplateLiteral \| NumericLiteral \| StringLiteral seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type BigIntLiteral \| ComputedPropertyName \| Identifier \| JsxNamespacedName \| NoSubstitutionTemplateLiteral \| NumericLiteral \| PrivateIdentifier \| StringLiteral seen as Type, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 |
| Refused: a value of type BigIntLiteral \| ComputedPropertyName \| Identifier \| JsxNamespacedName \| NoSubstitutionTemplateLiteral \| NumericLiteral \| PrivateIdentifier \| StringLiteral \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 |
| Refused: a value of type BigIntLiteral \| ComputedPropertyName \| Identifier \| NoSubstitutionTemplateLiteral \| NumericLiteral \| PrivateIdentifier \| StringLiteral \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type BinaryExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type BindingName seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Block seen as Mutable<Block>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Block \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 |
| Refused: a value of type BreakStatement seen as Mutable<BreakStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type CallExpression \| undefined seen as NodeArray<Expression> \| undefined, whose readonly field transformFlags becomes writable: a readonly field may hold something narrower than TransformFlags, which a write of TransformFlags would replace | 1 |
| Refused: a value of type CallSignatureDeclaration seen as Mutable<CallSignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type CapturedThis seen as PrimaryExpression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type CapturedThis seen as string \| BindingName, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type CaseBlock seen as Mutable<CaseBlock>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Children seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ClassDeclaration seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ClassDeclaration \| ClassExpression \| InferTypeNode \| InterfaceDeclaration \| JSDocCallbackTag \| ... 4 more ... \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type ClassDeclaration \| FunctionDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type CompilerHost seen as CompilerHostLikeForCache, which can write WriteFileCallback \| undefined where WriteFileCallback is read | 1 |
| Refused: a value of type ConstructSignatureDeclaration seen as Mutable<ConstructSignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ConstructorTypeNode seen as Mutable<ConstructorTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ContinueStatement seen as Mutable<ContinueStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Declaration seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type Declaration \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 |
| Refused: a value of type DestructuringAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type DiagnosticMessageChain \| { messageText: string; category: DiagnosticCategory; code: number; repopulateInfo?: () => RepopulateDiagnosticChainInfo; canonicalHead?: CanonicalDiagnostic; next: ... \| undefined; } seen as ReusableDiagnosticMessageChain, which can write ReusableDiagnosticMessageChain[] \| undefined where DiagnosticMessageChain[] \| undefined is read | 1 |
| Refused: a value of type DiagnosticMessageChain[] seen as ReusableDiagnosticMessageChain[], which can write ReusableDiagnosticMessageChain where DiagnosticMessageChain is read | 1 |
| Refused: a value of type DiagnosticMessageChain[] seen as ReusableDiagnosticMessageChain[], which can write ReusableDiagnosticMessageChain[] \| undefined where DiagnosticMessageChain[] \| undefined is read | 1 |
| Refused: a value of type DiagnosticWithLocation seen as Diagnostic \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 |
| Refused: a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 |
| Refused: a value of type DiagnosticWithLocation[] seen as readonly Diagnostic[] \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 |
| Refused: a value of type DiagnosticWithLocation[] seen as readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 1 |
| Refused: a value of type DiagnosticWithLocation[] \| undefined seen as readonly Diagnostic[] \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 |
| Refused: a value of type EntityNameExpression \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type EnumDeclaration \| ModuleDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type EqualsGreaterThanToken seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ExportAssignment seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ExportDeclaration \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type Expression \| GeneratedIdentifier seen as Expression \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type Expression \| GeneratedIdentifier seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type Expression \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 |
| Refused: a value of type ExpressionStatement seen as Mutable<ExpressionStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Extension[] seen as string[], which can write string where Extension is read | 1 |
| Refused: a value of type FlowArrayMutation \| FlowAssignment \| FlowCall \| FlowCondition \| FlowLabel \| FlowReduceLabel \| FlowStart \| FlowUnreachable seen as FlowNode, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 1 |
| Refused: a value of type FunctionTypeNode seen as Mutable<FunctionTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type GeneratedIdentifier seen as Expression \| GeneratedIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier seen as GeneratedIdentifier \| Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier seen as Node \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as GeneratedIdentifier \| GeneratedPrivateIdentifier \| Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as Identifier \| PrivateIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as Node \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier \| Identifier \| PrivateIdentifier seen as Identifier \| PrivateIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| Identifier seen as GeneratedIdentifier \| Identifier \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| Identifier seen as Identifier \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| Identifier seen as string \| BindingName, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| Identifier seen as string \| GeneratedIdentifier \| Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedIdentifier \| Identifier \| undefined seen as string \| ModuleExportName \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GeneratedPrivateIdentifier seen as Node \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type GetAccessorDeclaration \| MethodDeclaration \| PropertyAssignment \| SetAccessorDeclaration \| ShorthandPropertyAssignment \| SpreadAssignment \| undefined seen as Type \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 |
| Refused: a value of type GetAccessorDeclaration \| SetAccessorDeclaration seen as Mutable<AccessorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Identifier & GeneratedIdentifier & { readonly escapedText: { __escapedIdentifier: void; } & "__this"; } seen as BindingName, which can write EmitNode \| undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 1 |
| Refused: a value of type Identifier seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Identifier \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type Identifier[] seen as ModuleExportName[] \| undefined, which can write ModuleExportName where Identifier is read | 1 |
| Refused: a value of type ImportAttribute seen as Mutable<ImportAttribute>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ImportAttributes seen as Mutable<ImportAttributes>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ImportClause seen as Mutable<ImportClause>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ImportDeclaration seen as Mutable<ImportDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ImportDeclaration seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ImportEqualsDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ImportTypeNode seen as Mutable<ImportTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type IndexInfo \| undefined seen as IndexSignatureDeclaration \| undefined, which can write number \| undefined where number is read | 1 |
| Refused: a value of type IndexSignatureDeclaration seen as Mutable<IndexSignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type InitializedVariableDeclaration seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type IntrinsicType[] seen as TypeParameter[], which can write TypeParameter where IntrinsicType is read | 1 |
| Refused: a value of type IntrinsicType[] \| undefined seen as TypeParameter[] \| undefined, which can write TypeParameter where IntrinsicType is read | 1 |
| Refused: a value of type IterationTypes[] seen as (IterationTypes \| undefined)[], which can write IterationTypes \| undefined where IterationTypes is read | 1 |
| Refused: a value of type JSDocAugmentsTag seen as Mutable<JSDocAugmentsTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocCallbackTag seen as Mutable<JSDocCallbackTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocFunctionType seen as Mutable<JSDocFunctionType>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocImplementsTag seen as Mutable<JSDocImplementsTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocImportTag seen as Mutable<JSDocImportTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocLink seen as Mutable<JSDocLink>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocLinkCode seen as Mutable<JSDocLinkCode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocLinkPlain seen as Mutable<JSDocLinkPlain>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocNameReference seen as Mutable<JSDocNameReference>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocOverloadTag seen as Mutable<JSDocOverloadTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocParameterTag seen as Mutable<JSDocParameterTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocPropertyTag seen as Mutable<JSDocPropertyTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocSeeTag seen as Mutable<JSDocSeeTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocSignature seen as Mutable<JSDocSignature>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocSignature \| SignatureDeclaration \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type JSDocTemplateTag seen as Mutable<JSDocTemplateTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocText seen as Mutable<JSDocText>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocTypeExpression seen as Mutable<JSDocTypeExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocTypeExpression \| undefined seen as Signature \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SignatureFlags would replace | 1 |
| Refused: a value of type JSDocTypeLiteral seen as Mutable<JSDocTypeLiteral>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocTypedefTag seen as Mutable<JSDocTypedefTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type JSDocUnknownTag seen as Mutable<JSDocUnknownTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Map<Path, Diagnostic[]> seen as Map<Path, readonly Diagnostic[]> \| undefined, which can write readonly Diagnostic[] where Diagnostic[] is read | 1 |
| Refused: a value of type Map<Path, DirectoryWatchesOfFailedLookup> seen as Map<string, DirectoryWatchesOfFailedLookup>, which can write string where Path is read | 1 |
| Refused: a value of type Map<Path, FileWatcher> seen as Map<string, FileWatcher>, which can write string where Path is read | 1 |
| Refused: a value of type Map<Path, ModeAwareCache<CachedResolvedModuleWithFailedLookupLocations>> seen as Map<string, ModeAwareCache<CachedResolvedModuleWithFailedLookupLocations>>, which can write string where Path is read | 1 |
| Refused: a value of type Map<Path, ModeAwareCache<CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations>> seen as Map<string, ModeAwareCache<CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations>>, which can write string where Path is read | 1 |
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
| Refused: a value of type Map<string, CachedResolvedModuleWithFailedLookupLocations> seen as Map<string, ResolutionWithFailedLookupLocations> \| Set<ResolutionWithFailedLookupLocations> \| undefined, which can write ResolutionWithFailedLookupLocations where CachedResolvedModuleWithFailedLookupLocations is read | 1 |
| Refused: a value of type Map<string, ImportsNotUsedAsValues> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ImportsNotUsedAsValues is read | 1 |
| Refused: a value of type Map<string, JsxEmit> seen as Map<string, string \| number>, which can write string \| number where JsxEmit is read | 1 |
| Refused: a value of type Map<string, Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>> seen as ScriptTargetFeatures, which can write string where never is read | 1 |
| Refused: a value of type Map<string, ModuleDetectionKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ModuleDetectionKind is read | 1 |
| Refused: a value of type Map<string, ModuleResolutionKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ModuleResolutionKind is read | 1 |
| Refused: a value of type Map<string, NewLineKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where NewLineKind is read | 1 |
| Refused: a value of type Map<string, PollingWatchKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where PollingWatchKind is read | 1 |
| Refused: a value of type Map<string, WatchDirectoryKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where WatchDirectoryKind is read | 1 |
| Refused: a value of type Map<string, WatchFileKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where WatchFileKind is read | 1 |
| Refused: a value of type Map<string, [VariableDeclarationList, VariableDeclaration[]]> seen as Map<string, [CatchClause \| VariableDeclarationList, VariableDeclaration[]]>, which can write [CatchClause \| VariableDeclarationList, VariableDeclaration[]] where [VariableDeclarationList, VariableDeclaration[]] is read | 1 |
| Refused: a value of type Map<string, string> seen as Map<string, string \| number>, which can write string \| number where string is read | 1 |
| Refused: a value of type MappedTypeNode seen as Mutable<MappedTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type MemberName \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type MethodDeclaration seen as Mutable<MethodDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type MethodDeclaration \| PropertyAssignment \| AccessorDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ModuleDeclaration seen as Mutable<ModuleDeclaration \| SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ModuleExportName seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ModuleName seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ModuleName seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ModuleResolutionState & { host: GetPackageJsonEntrypointsHost; } seen as ModuleResolutionState, which can write ModuleResolutionHost where ModuleResolutionHost & GetPackageJsonEntrypointsHost is read | 1 |
| Refused: a value of type Mutable<NoSubstitutionTemplateLiteral> seen as Mutable<TemplateLiteralLikeNode>, which can write SyntaxKind where SyntaxKind.NoSubstitutionTemplateLiteral is read | 1 |
| Refused: a value of type Mutable<T> seen as T, a type parameter whose constraint JSDocType & { readonly type: TypeNode \| undefined; readonly postfix: boolean; } can be written, so it can write what Mutable<T> can't hold | 1 |
| Refused: a value of type Mutable<T> seen as T, a type parameter whose constraint JSDocType & { readonly type: TypeNode \| undefined; } can be written, so it can write what Mutable<T> can't hold | 1 |
| Refused: a value of type Mutable<T> seen as T, a type parameter whose constraint Node can be written, so it can write what Mutable<T> can't hold | 1 |
| Refused: a value of type NamedImports seen as Mutable<NamedImports>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type NamespaceExport seen as Mutable<NamespaceExport>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type NamespaceExportDeclaration seen as Mutable<NamespaceExportDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type NamespaceImport seen as Mutable<NamespaceImport>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Node seen as Node \| SourceMapRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Node seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Node \| SourceMapRange seen as SourceMapRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Node \| undefined seen as EmitNode \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of EmitFlags would replace | 1 |
| Refused: a value of type Node \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type NonNullExpression seen as Mutable<NonNullExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type NumberLiteralType seen as LiteralType, which can write string \| number \| PseudoBigInt where number is read | 1 |
| Refused: a value of type ObjectBindingOrAssignmentPattern seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type OptionalTypeNode seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ParameterDeclaration seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ParameterDeclaration \| PropertyDeclaration \| PropertySignature \| SignatureDeclaration seen as Mutable<ParameterDeclaration \| PropertyDeclaration \| PropertySignature \| SignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ParameterDeclaration \| VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ParseConfigHost seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where boolean is read | 1 |
| Refused: a value of type Path[] seen as string[], which can write string where Path is read | 1 |
| Refused: a value of type PropertyAccessExpression \| SyntheticSuper seen as LeftHandSideExpression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type PropertyAccessExpression \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type PropertyAssignment seen as Mutable<PropertyAssignment \| ShorthandPropertyAssignment>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type PropertyDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type PropertyName \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type Readonly<BuilderState> \| undefined seen as BuilderState \| undefined, whose readonly field fileInfos becomes writable: a readonly field may hold something narrower than Map<Path, FileInfo>, which a write of Map<Path, FileInfo> would replace | 1 |
| Refused: a value of type ReferencedFile & { kind: FileIncludeKind.LibReferenceDirective; } seen as ReferencedFile, which can write ReferencedFileKind where FileIncludeKind.LibReferenceDirective is read | 1 |
| Refused: a value of type ReportFileInError[] seen as (ReportFileInError \| undefined)[], which can write ReportFileInError \| undefined where ReportFileInError is read | 1 |
| Refused: a value of type ResolvedModuleFull \| undefined seen as { path: string; originalPath: string \| true; extension: string; packageId: PackageId \| undefined; resolvedUsingTsExtension: boolean \| undefined; } \| undefined, whose readonly field originalPath becomes writable: a readonly field may hold something narrower than string \| undefined, which a write of string \| true would replace | 1 |
| Refused: a value of type ResolvedType \| TypeReference seen as ObjectType, which can write SymbolTable \| undefined where SymbolTable is read | 1 |
| Refused: a value of type ReturnStatement seen as Mutable<ReturnStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type SearchResult<Resolved> seen as { value: { resolved: Resolved; isExternalLibraryImport: true; } \| undefined; } \| undefined, whose readonly field value becomes writable: a readonly field may hold something narrower than Resolved \| undefined, which a write of { resolved: Resolved; isExternalLibraryImport: true; } \| undefined would replace | 1 |
| Refused: a value of type Set<Path> \| undefined seen as Set<string> \| undefined, which can write string where Path is read | 1 |
| Refused: a value of type Set<__String> seen as Set<__String \| undefined>, which can write __String \| undefined where __String is read | 1 |
| Refused: a value of type SetAccessorDeclaration seen as Mutable<SetAccessorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ShorthandPropertyAssignment seen as Mutable<PropertyAssignment \| ShorthandPropertyAssignment>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type SourceFile seen as Mutable<ModuleDeclaration \| SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type SourceFile seen as SourceFileLike \| undefined, which can write readonly number[] \| undefined where readonly number[] is read | 1 |
| Refused: a value of type SourceFile \| undefined seen as FileReasonToChainCache \| undefined, which can write DiagnosticMessageChain[] \| undefined where RedirectInfo \| undefined is read | 1 |
| Refused: a value of type SourceFile \| undefined seen as NodeLinks \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of NodeCheckFlags would replace | 1 |
| Refused: a value of type SourceFile \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type SourceMapSource seen as SourceFileLike, which can write readonly number[] \| undefined where readonly number[] is read | 1 |
| Refused: a value of type Statement[] seen as Node[], which can write Node where Statement is read | 1 |
| Refused: a value of type Statement[] \| undefined seen as CaseClause[] \| undefined, which can write CaseClause where Statement is read | 1 |
| Refused: a value of type StringLiteral seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type StringLiteral seen as Mutable<StringLiteral>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type StringLiteralType seen as LiteralType, which can write string \| number \| PseudoBigInt where string is read | 1 |
| Refused: a value of type StringLiteralType[] seen as Type[], which can write Type where StringLiteralType is read | 1 |
| Refused: a value of type SuperExpression seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type SuperExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type SwitchStatement seen as Mutable<SwitchStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type Symbol seen as Type \| undefined, which can write TypeFlags where SymbolFlags is read | 1 |
| Refused: a value of type Symbol seen as Type, which can write TypeFlags where SymbolFlags is read | 1 |
| Refused: a value of type Symbol \| undefined seen as GenericType \| undefined, which can write TypeFlags where SymbolFlags is read | 1 |
| Refused: a value of type Symbol \| undefined seen as Node \| undefined, which can write number \| undefined where number is read | 1 |
| Refused: a value of type Symbol \| undefined seen as SourceFile \| undefined, which can write number \| undefined where number is read | 1 |
| Refused: a value of type Symbol[] seen as (Symbol \| undefined)[], which can write Symbol \| undefined where Symbol is read | 1 |
| Refused: a value of type Symbol[] seen as Symbol[], which can write Symbol where never is read | 1 |
| Refused: a value of type SyntheticSuper seen as string \| BindingName, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type T seen as T, a type parameter whose constraint ClassDeclaration \| ClassExpression \| GetAccessorDeclaration \| MethodDeclaration \| ParameterDeclaration \| PropertyDeclaration \| SetAccessorDeclaration can be written, so it can write what T can't hold | 1 |
| Refused: a value of type T seen as T, a type parameter whose constraint EntityNameOrEntityNameExpression can be written, so it can write what T can't hold | 1 |
| Refused: a value of type T seen as T, a type parameter whose constraint HasModifiers can be written, so it can write what T can't hold | 1 |
| Refused: a value of type T seen as T, a type parameter whose constraint MethodDeclaration \| MethodSignature \| PropertyAssignment \| PropertyDeclaration \| PropertySignature \| AccessorDeclaration can be written, so it can write what T can't hold | 1 |
| Refused: a value of type T seen as T, a type parameter whose constraint ModifierSyntaxKind can be written, so it can write what T can't hold | 1 |
| Refused: a value of type T seen as T, a type parameter whose constraint Node \| undefined can be written, so it can write what T can't hold | 1 |
| Refused: a value of type T seen as T, a type parameter whose constraint ResolutionWithFailedLookupLocations can be written, so it can write what T can't hold | 1 |
| Refused: a value of type TKind seen as TKind, a type parameter whose constraint KeywordTypeSyntaxKind can be written, so it can write what TKind can't hold | 1 |
| Refused: a value of type T[] seen as (T \| undefined)[], which can write T \| undefined where T is read | 1 |
| Refused: a value of type T[][] seen as (readonly T[])[], which can write readonly T[] where T[] is read | 1 |
| Refused: a value of type TaggedTemplateExpression seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type ThisCapturingVariableDeclaration seen as VariableDeclaration, which can write EmitNode \| undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 1 |
| Refused: a value of type ThrowStatement seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type TransientSymbol[] seen as Symbol[], which can write Symbol where TransientSymbol is read | 1 |
| Refused: a value of type Type seen as TypeNode, which can write number \| undefined where number is read | 1 |
| Refused: a value of type Type \| undefined seen as Expression \| undefined, which can write number \| undefined where number is read | 1 |
| Refused: a value of type Type \| undefined seen as JSDocTypeExpression \| undefined, which can write number \| undefined where number is read | 1 |
| Refused: a value of type Type \| undefined seen as Signature \| undefined, which can write SignatureFlags where TypeFlags is read | 1 |
| Refused: a value of type TypeNode seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type TypeOperatorNode seen as Mutable<TypeOperatorNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type TypeParameterDeclaration seen as Mutable<TypeParameterDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type TypeParameterDeclaration \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 |
| Refused: a value of type TypeVariable[] seen as (Type \| undefined)[], which can write Type \| undefined where TypeVariable is read | 1 |
| Refused: a value of type Type[] seen as (Type \| undefined)[], which can write Type \| undefined where Type is read | 1 |
| Refused: a value of type VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type VariableDeclaration \| DestructuringAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type VariableDeclarationList seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type VariableDeclaration[] seen as VariableStatement[], which can write VariableStatement where VariableDeclaration is read | 1 |
| Refused: a value of type VariableStatement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read | 1 |
| Refused: a value of type YieldExpression seen as Mutable<YieldExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type [() => Type, __String][] seen as (readonly [() => Type, __String])[], which can write readonly [() => Type, __String] where [() => Type, __String] is read | 1 |
| Refused: a value of type never seen as T, a type parameter whose constraint Node can be written, so it can write what never can't hold | 1 |
| Refused: a value of type never seen as T, a type parameter whose constraint any can be written, so it can write what never can't hold | 1 |
| Refused: a value of type never seen as U, a type parameter whose constraint {} can be written, so it can write what never can't hold | 1 |
| Refused: a value of type never[] seen as (JSDoc \| JSDocTag)[], which can write JSDoc \| JSDocTag where never is read | 1 |
| Refused: a value of type never[] seen as (JSDocCallbackTag \| JSDocEnumTag \| JSDocTypedefTag)[], which can write JSDocCallbackTag \| JSDocEnumTag \| JSDocTypedefTag where never is read | 1 |
| Refused: a value of type never[] seen as (SourceMapRange \| undefined)[], which can write SourceMapRange \| undefined where never is read | 1 |
| Refused: a value of type never[] seen as (readonly (Identifier \| StringLiteral)[])[], which can write readonly (Identifier \| StringLiteral)[] where never is read | 1 |
| Refused: a value of type never[] seen as BindingElement[], which can write BindingElement where never is read | 1 |
| Refused: a value of type never[] seen as CommentRange[], which can write CommentRange where never is read | 1 |
| Refused: a value of type never[] seen as Comparator[][], which can write Comparator[] where never is read | 1 |
| Refused: a value of type never[] seen as Expression[], which can write Expression where never is read | 1 |
| Refused: a value of type never[] seen as FlowNode[], which can write FlowNode where never is read | 1 |
| Refused: a value of type never[] seen as Identifier[], which can write Identifier where never is read | 1 |
| Refused: a value of type never[] seen as IntrinsicType[], which can write IntrinsicType where never is read | 1 |
| Refused: a value of type never[] seen as JSDocImportTag[], which can write JSDocImportTag where never is read | 1 |
| Refused: a value of type never[] seen as JsxAttributes[], which can write JsxAttributes where never is read | 1 |
| Refused: a value of type never[] seen as ParameterDeclaration[], which can write ParameterDeclaration where never is read | 1 |
| Refused: a value of type never[] seen as Path[], which can write Path where never is read | 1 |
| Refused: a value of type never[] seen as PotentiallyUnusedIdentifier[], which can write PotentiallyUnusedIdentifier where never is read | 1 |
| Refused: a value of type never[] seen as ProjectReference[], which can write ProjectReference where never is read | 1 |
| Refused: a value of type never[] seen as PropertyAssignment[], which can write PropertyAssignment where never is read | 1 |
| Refused: a value of type never[] seen as RequireOrImportCall[], which can write RequireOrImportCall where never is read | 1 |
| Refused: a value of type never[] seen as ResolvedProjectReference[], which can write ResolvedProjectReference where never is read | 1 |
| Refused: a value of type never[] seen as SourceMappedPosition[], which can write SourceMappedPosition where never is read | 1 |
| Refused: a value of type never[] seen as SymbolTable[], which can write SymbolTable where never is read | 1 |
| Refused: a value of type never[] seen as Symbol[] \| undefined, which can write Symbol where never is read | 1 |
| Refused: a value of type never[] seen as TransformerFactory<Bundle \| SourceFile>[], which can write TransformerFactory<Bundle \| SourceFile> where never is read | 1 |
| Refused: a value of type never[] seen as VarianceFlags[], which can write VarianceFlags where never is read | 1 |
| Refused: a value of type never[] \| SortedArray<DiagnosticWithLocation> seen as Diagnostic[], which can write Diagnostic where never is read | 1 |
| Refused: a value of type never[][] seen as string[][], which can write string where never is read | 1 |
| Refused: a value of type number[] seen as string \| (string \| number)[] \| undefined, which can write string \| number where number is read | 1 |
| Refused: a value of type readonly TypeParameter[] \| undefined seen as TypeParameterDeclaration[] \| undefined, which can write TypeParameterDeclaration where TypeParameter is read | 1 |
| Refused: a value of type readonly string[] \| undefined seen as RegExp[] \| undefined, which can write RegExp where string is read | 1 |
| Refused: a value of type string \| GeneratedIdentifier \| Identifier seen as string \| ModuleExportName, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 |
| Refused: a value of type string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined seen as string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined, which can write string \| number where string is read | 1 |
| Refused: a value of type string \| number \| boolean \| string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> seen as CompilerOptionsValue, which can write string \| number where string is read | 1 |
| Refused: a value of type string[] seen as DiagnosticArguments, which can write string \| number \| boolean \| readonly string[] \| SourceFile \| undefined where string is read | 1 |
| Refused: a value of type string[] seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what string[] can't hold | 1 |
| Refused: a value of type string[] seen as string \| (string \| number)[] \| undefined, which can write string \| number where string is read | 1 |
| Refused: a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"CheckMode", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof CheckMode is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"EmitFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof EmitFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"GeneratedIdentifierFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof GeneratedIdentifierFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"ModifierFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ModifierFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"NodeCheckFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof NodeCheckFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"NodeFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof NodeFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"ObjectFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ObjectFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"RelationComparisonResult", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof RelationComparisonResult is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"ScriptKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof ScriptKind is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"SignatureCheckMode", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SignatureCheckMode is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"SignatureFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SignatureFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"SnippetKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SnippetKind is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"SymbolFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SymbolFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"SyntaxKind", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof SyntaxKind is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"TransformFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TransformFlags is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"TypeFacts", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TypeFacts is read | 1 |
| Refused: a value of type typeof import("/tmp/tsc-closure-adapted/src/compiler/_namespaces/ts") seen as Record<"TypeFlags", Record<string, string \| number>>, which can write Record<string, string \| number> where typeof TypeFlags is read | 1 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what undefined can't hold | 1 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint Declaration can be written, so it can write what undefined can't hold | 1 |
| Refused: a value of type { (fileName: string): DiagnosticWithLocation[]; (): Diagnostic[]; } seen as { (): Diagnostic[]; (fileName: string): DiagnosticWithLocation[]; }, which can write Diagnostic where DiagnosticWithLocation is read | 1 |
| Refused: a value of type { affectedFile: SourceFile; emitKind: BuilderFileEmit.Js \| BuilderFileEmit.JsMap \| BuilderFileEmit.JsInlineMap \| BuilderFileEmit.DtsErrors \| ... 6 more ... \| BuilderFileEmit.All; } seen as { affectedFile: Program \| SourceFile \| undefined; emitKind: BuilderFileEmit; }, which can write Program \| SourceFile \| undefined where SourceFile is read | 1 |
| Refused: a value of type { all?: boolean; allowImportingTsExtensions?: boolean; allowJs?: boolean; allowNonTsExtensions?: boolean; allowArbitraryExtensions?: boolean; allowSyntheticDefaultImports?: boolean; allowUmdGlobalAccess?: boolean; ... 129 more ...; moduleResolution: ModuleResolutionKind; } seen as CompilerOptions, which can write ModuleResolutionKind \| undefined where ModuleResolutionKind is read | 1 |
| Refused: a value of type { arguments: { name?: string; } & { path: string; }; range: CommentRange; } \| { arguments: { name: string; }; range: CommentRange; } \| { arguments: { factory: string; }; range: CommentRange; } \| ... 5 more ... \| ... seen as ({ arguments: { name?: string; } & { path: string; }; range: CommentRange; } \| { arguments: { name: string; }; range: CommentRange; } \| { arguments: { factory: string; }; range: CommentRange; } \| ... 5 more ... \| ...)[] \| ... 8 more ... \| ..., which can write { name?: string; } & { path: string; } where never is read | 1 |
| Refused: a value of type { compilerOptions: CompilerOptions; traceEnabled: boolean; affectingLocations: string[] \| undefined; resultFromCache?: ResolvedModuleWithFailedLookupLocations; ... 9 more ...; host: GetPackageJsonEntrypointsHost; } seen as ModuleResolutionState & { host: GetPackageJsonEntrypointsHost; }, which can write string[] \| undefined where never[] is read | 1 |
| Refused: a value of type { ending: ModuleSpecifierEnding; value: string; }[] seen as { ending: ModuleSpecifierEnding \| undefined; value: string; }[], which can write ModuleSpecifierEnding \| undefined where ModuleSpecifierEnding is read | 1 |
| Refused: a value of type { host: ModuleResolutionHost; traceEnabled: boolean; failedLookupLocations: string[] \| undefined; affectingLocations: string[] \| undefined; resultFromCache?: ResolvedModuleWithFailedLookupLocations; ... 8 more ...; reportDiagnostic: ...; } seen as ModuleResolutionState, which can write CompilerOptions where { all?: boolean; allowImportingTsExtensions?: boolean; allowJs?: boolean; allowNonTsExtensions?: boolean; allowArbitraryExtensions?: boolean; allowSyntheticDefaultImports?: boolean; allowUmdGlobalAccess?: boolean; ... 129 more ...; moduleResolution: ModuleResolutionKind; } is read | 1 |
| Refused: a value of type { id: number; flowNode: FlowNode; edges: never[]; text: string; lane: number; endLane: number; level: number; circular: false; } seen as FlowGraphNode, which can write FlowGraphEdge[] where never[] is read | 1 |
| Refused: a value of type { introducesError: boolean; node: Identifier \| PropertyAccessEntityNameExpression; sym?: never; } \| { introducesError: boolean; node: Identifier \| PropertyAccessEntityNameExpression; sym: Symbol \| undefined; } seen as { introducesError: boolean; node: LeftHandSideExpression; }, which can write LeftHandSideExpression where Identifier \| PropertyAccessEntityNameExpression is read | 1 |
| Refused: a value of type { major: number; minor: number; patch: number; prerelease: string; build: string; } seen as { major: string \| number; minor: number; patch: number; prerelease: string \| readonly string[]; build: string \| readonly string[]; }, which can write string \| number where number is read | 1 |
| Refused: a value of type { name: string \| undefined; path: string; }[] seen as AmdDependency[], which can write AmdDependency where { name: string \| undefined; path: string; } is read | 1 |
| Refused: a value of type { noInferenceFallback?: boolean \| undefined; enclosingDeclaration: ModuleDeclaration; enclosingFile: SourceFile \| undefined; flags: NodeBuilderFlags; ... 28 more ...; out: WriterContextOut; } seen as NodeBuilderContext, which can write Node \| undefined where ModuleDeclaration is read | 1 |
| Refused: a value of type { referencedName: StringLiteral; name: PropertyName; } \| { referencedName: Identifier; name: ComputedPropertyName; } seen as { referencedName: Expression \| undefined; name: PropertyName \| undefined; }, which can write Expression \| undefined where StringLiteral is read | 1 |
| Refused: an arbitrary number or a value from another enum assigned to Extension.Mjs; its members are a closed union | 1 |
| Refused: an arbitrary number or a value from another enum assigned to ForegroundColorEscapeSequences.Grey; its members are a closed union | 1 |
| Refused: an unproven relation from DiagnosticRelatedInformation to Diagnostic: optional field reportsUnnecessary has no proven compatible presence/type | 1 |
| Refused: an unproven relation from Identifier to Identifier: optional field id has no proven compatible presence/type | 1 |
| Refused: an unproven relation from ImportEqualsDeclaration & { moduleReference: ExternalModuleReference; } to ImportEqualsDeclaration: optional field moduleReference.id has no proven compatible presence/type | 1 |
| Refused: an unproven relation from InstantiableType \| UnionOrIntersectionType to TypeParameter: optional field constraint has no proven compatible presence/type | 1 |
| Refused: an unproven relation from LiteralLikeNode to TemplateLiteralLikeNode: optional field rawText has no proven compatible presence/type | 1 |
| Refused: an unproven relation from NodeArray<BindingElement> \| NodeArray<Expression> \| NodeArray<ArrayBindingElement> to NodeArray<Node>: optional field concat.parameter.element.slice.element.propertyName has no proven compatible presence/type | 1 |
| Refused: an unproven relation from NodeArray<ModifierLike> & readonly Decorator[] to NodeArray<Decorator>: optional field concat.element.parent.name has no proven compatible presence/type | 1 |
| Refused: an unproven relation from PackageJsonPathFields to PackageJson: optional field version has no proven compatible presence/type | 1 |
| Refused: an unproven relation from SolutionBuilderHostBase<T> to SolutionBuilderHost<T>: optional field reportErrorSummary has no proven compatible presence/type | 1 |
| Refused: an unproven relation from StructuredType to AnonymousType: optional field target has no proven compatible presence/type | 1 |
| Refused: an unproven relation from UnionOrIntersectionType to UnionType: optional field resolvedReducedType has no proven compatible presence/type | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot BuildStep.Done | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot BuilderFileEmit.Js \| BuilderFileEmit.JsMap \| BuilderFileEmit.JsInlineMap \| BuilderFileEmit.DtsErrors \| BuilderFileEmit.DtsEmit \| ... 5 more ... \| BuilderFileEmit.All | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot BuilderFileEmit.None | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot CharacterCodes.plus | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot CharacterCodes.slash | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot Connection | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot EmitOnly.Dts | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot ModuleKind.ESNext | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot NodeFlags.Const | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot NodeFlags.NestedNamespace | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot PunctuationOrKeywordSyntaxKind | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.BindingElement | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Block | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ClassStaticBlockDeclaration | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ConditionalType | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.DotDotDotToken | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ExportAssignment | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ExtendsKeyword \| SyntaxKind.ImplementsKeyword | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ImportAttributes | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ImportDeclaration | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.KeyOfKeyword \| SyntaxKind.ReadonlyKeyword \| SyntaxKind.UniqueKeyword | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.MethodSignature \| SyntaxKind.MethodDeclaration \| SyntaxKind.Constructor \| SyntaxKind.GetAccessor \| SyntaxKind.SetAccessor \| SyntaxKind.CallSignature \| SyntaxKind.ConstructSignature \| ... 6 more ... \| SyntaxKind.JSDocFunctionType | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NamespaceExport | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NamespaceImport | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NumericLiteral \| SyntaxKind.BigIntLiteral \| SyntaxKind.StringLiteral \| SyntaxKind.JsxText \| SyntaxKind.JsxTextAllWhiteSpaces \| SyntaxKind.RegularExpressionLiteral \| SyntaxKind.NoSubstitutionTemplateLiteral | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ObjectLiteralExpression | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Parameter | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.QuestionToken | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.StringLiteral | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Unknown \| SyntaxKind.NumericLiteral \| SyntaxKind.BigIntLiteral \| SyntaxKind.StringLiteral \| SyntaxKind.JsxText \| SyntaxKind.RegularExpressionLiteral \| SyntaxKind.NoSubstitutionTemplateLiteral | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.VariableDeclaration | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot TempFlags | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot Ternary.False | 1 |
| Refused: an unproven value assigned to a numeric literal or enum member slot VarianceFlags.Contravariant | 1 |
| Refused: arguments | 1 |
| Refused: inherited library member compare read as an own field | 1 |
| Refused: inherited library member hasOwnProperty read as an own field | 1 |
| Refused: inherited library member repeat read as an own field | 1 |
| Refused: inherited library member replace read as an own field | 1 |
| Refused: optional property Low in Partial<Levels> absent from structural source {}, which can hide fields | 1 |
| Refused: optional property _propertyAccessExpressionLikeQualifiedNameBrand in PropertyAccessEntityNameExpression absent from structural source never, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source OptionsBase, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "noImplicitAny" \| "strict">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "noImplicitThis" \| "strict">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" \| "strictBindCallApply">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" \| "strictBuiltinIteratorReturn">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" \| "strictFunctionTypes">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" \| "strictNullChecks">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" \| "strictPropertyInitialization">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" \| "useUnknownInCatchVariables">, which can hide fields | 1 |
| Refused: optional property all in CompilerOptions absent from structural source TypeAcquisition, which can hide fields | 1 |
| Refused: optional property constraint in TypeParameter absent from structural source ObjectType, which can hide fields | 1 |
| Refused: optional property createProgram in IncrementalProgramOptions<EmitAndSemanticDiagnosticsBuilderProgram> absent from structural source IncrementalCompilationOptions, which can hide fields | 1 |
| Refused: optional property default in TypeParameter absent from structural source IndexedAccessType, which can hide fields | 1 |
| Refused: optional property id in BinaryExpression absent from structural source never, which can hide fields | 1 |
| Refused: optional property id in ComputedPropertyName absent from structural source never, which can hide fields | 1 |
| Refused: optional property id in JSDocThisTag absent from structural source never, which can hide fields | 1 |
| Refused: optional property id in PrivateIdentifier absent from structural source never, which can hide fields | 1 |
| Refused: optional property isInvalidated in CachedResolvedModuleWithFailedLookupLocations absent from structural source ResolvedModuleWithFailedLookupLocations, which can hide fields | 1 |
| Refused: optional property jsDocCache in JSDocArray absent from structural source JSDoc[], which can hide fields | 1 |
| Refused: optional property localSymbol in Declaration absent from structural source never, which can hide fields | 1 |
| Refused: optional property members in ObjectType absent from structural source IntersectionType, which can hide fields | 1 |
| Refused: optional property members in ObjectType absent from structural source UnionType, which can hide fields | 1 |
| Refused: optional property modifiers in ExportAssignment absent from structural source never, which can hide fields | 1 |
| Refused: optional property modifiers in ParameterDeclaration absent from structural source never, which can hide fields | 1 |
| Refused: optional property modifiers in PropertyDeclaration absent from structural source never, which can hide fields | 1 |
| Refused: optional property modifiers in SetAccessorDeclaration absent from structural source never, which can hide fields | 1 |
| Refused: optional property outputDts in ResolvedRefAndOutputDts absent from structural source ResolvedRefAndSource, which can hide fields | 1 |
| Refused: optional property packageRootPath in { moduleFileToTry: string; packageRootPath?: string; blockedByExports?: true; verbatimFromExports?: true; } absent from structural source { moduleFileToTry: string; }, which can hide fields | 1 |
| Refused: optional property preserve in FileReference absent from structural source { resolutionMode: ModuleKind.CommonJS \| ModuleKind.ESNext; }, which can hide fields | 1 |
| Refused: optional property reportsUnnecessary in ReusableDiagnostic absent from structural source ReusableDiagnosticRelatedInformation, which can hide fields | 1 |
| Refused: optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<ImportAttribute>, which can hide fields | 1 |
| Refused: optional property return in ArrayIterator<WatchedFileWithUnchangedPolls \| undefined> absent from structural source ArrayIterator<WatchedFileWithUnchangedPolls>, which can hide fields | 1 |
| Refused: optional property return in MapIterator<[string, WatchDirectoryFlags]> absent from structural source MapIterator<[string, WatchDirectoryFlags]>, which can hide fields | 1 |
| Refused: optional property skipLogging in ErrorOutputContainer absent from structural source { errors?: Diagnostic[]; }, which can hide fields | 1 |
| Refused: optional property skipTrivia in SourceMapSource absent from structural source SourceFile, which can hide fields | 1 |
| Refused: optional property source in ResolvedRefAndSource absent from structural source ResolvedRefAndOutputDts, which can hide fields | 1 |
| Refused: optional property target in AnonymousType absent from structural source IntrinsicType, which can hide fields | 1 |
| Refused: optional property types in { types?: { value: string; pos: number; end: number; }; } & { lib?: { value: string; pos: number; end: number; }; } & { path?: { value: string; pos: number; end: number; }; } & { "no-default-lib"?: string; } & { "resolution-mode"?: string; } & ... absent from structural source { [index: string]: string \| undefined; }, which can hide fields | 1 |
| Refused: optional property types in { types?: { value: string; pos: number; end: number; }; } & { lib?: { value: string; pos: number; end: number; }; } & { path?: { value: string; pos: number; end: number; }; } & { "no-default-lib"?: string; } & { "resolution-mode"?: string; } & ... absent from structural source { [index: string]: string \| { value: string; pos: number; end: number; }; }, which can hide fields | 1 |
| Refused: optional property watchFile in WatchOptions absent from structural source {}, which can hide fields | 1 |
| panic: Unhandled case in Node.Text: *ast.ComputedPropertyName | 1 |
| panic: Unhandled case in Node.Text: *ast.QualifiedName | 1 |

## full / outside_compiler

| Kind and exact reason | Sites |
| --- | ---: |
| NotYet: reading ts | 5 |
| Refused: a method read as a value (setBlocking would lose its object, and this with it) | 1 |
| Refused: a method read as a value (tryEnableSourceMapsForHost would lose its object, and this with it) | 1 |

## full / unattributed

| Kind and exact reason | Sites |
| --- | ---: |
| error: statement panic: Unhandled case in Node.Text: *ast.ComputedPropertyName | 1 |
| error: statement panic: runtime error: invalid memory address or nil pointer dereference | 1 |
