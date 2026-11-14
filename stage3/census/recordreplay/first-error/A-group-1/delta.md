# Paired latent census

| Measure | Raw findings | Distinct reads |
| --- | ---: | ---: |
| Lowered (no remaining finding) | 0 | 0 |
| Moved by reason | 0 | 0 |
| New findings | 0 | 0 |
| New sites | 0 | 0 |
| Units whose first stop moved | 0 | 0 |
| Regressions | 0 | 0 |
| before NotYet | 123 | 123 |
| before Refused | 544 | 515 |
| after NotYet | 123 | 123 |
| after Refused | 544 | 515 |

| Reason | Raw before | Raw after | Reads before | Reads after | Reads lowered |
| --- | ---: | ---: | ---: | ---: | ---: |
| ?. to a number, which would be number \| undefined | 1 | 1 | 1 | 1 | 0 |
| Array.isArray on a tuple or an erased object/any/unknown view | 1 | 1 | 1 | 1 | 0 |
| Object enumeration with getter or accessor descriptors | 4 | 4 | 4 | 4 | 0 |
| Object.assign on a shape not proven by a plain literal or its const binding | 1 | 1 | 1 | 1 | 0 |
| RegExp with a nonconstant pattern | 2 | 2 | 2 | 2 | 0 |
| a (current: Node) => boolean seen as a (element: Node) => "quit" \| boolean (one keeps something weakly that the other keeps strongly) | 1 | 1 | 1 | 1 | 0 |
| a (n: Node) => boolean seen as a (element: Node) => "quit" \| boolean (one keeps something weakly that the other keeps strongly) | 1 | 1 | 1 | 1 | 0 |
| a BinaryExpression with a string and a union of differently held members | 1 | 1 | 1 | 1 | 0 |
| a Map or Set interface implemented by a value that isn't a Map or Set | 3 | 3 | 3 | 3 | 0 |
| a PostfixUnaryExpression | 1 | 1 | 1 | 1 | 0 |
| a boolean \| undefined variable a function value captures | 1 | 1 | 1 | 1 | 0 |
| a branded supertype inside a checked predicate's array target | 1 | 1 | 1 | 1 | 0 |
| a call spreading elements that cannot be packed | 3 | 3 | 3 | 3 | 0 |
| a call spreading something other than an array | 1 | 1 | 1 | 1 | 0 |
| a call through ?. (an optional call) | 2 | 2 | 2 | 2 | 0 |
| a case whose type differs from the switch's | 2 | 2 | 2 | 2 | 0 |
| a cast that makes a viewed readonly array writable | 2 | 2 | 2 | 2 | 0 |
| a cast the runtime can't check | 163 | 163 | 163 | 163 | 0 |
| a checked cast from never[] to [] whose value isn't a plain object | 1 | 1 | 1 | 1 | 0 |
| a checked predicate target DynamicNamedDeclaration that narrows GetAccessorDeclaration, which isn't its supertype | 1 | 1 | 1 | 1 | 0 |
| a checked predicate target whose check would read past the depth bound (PackageJsonInfo) | 1 | 1 | 1 | 1 | 0 |
| a comparator that isn't a function | 1 | 1 | 1 | 1 | 0 |
| a dictionary member narrowed to an array without an element-checked view | 1 | 1 | 1 | 1 | 0 |
| a function or method with a this parameter followed by another parameter (dynamic receivers are not implemented) | 7 | 7 | 7 | 7 | 0 |
| a function returning (ConstructorDeclaration & { body: Block; }) \| undefined | 1 | 1 | 1 | 1 | 0 |
| a function returning AnyValidImportOrReExport | 1 | 1 | 1 | 1 | 0 |
| a function returning AnyValidImportOrReExport \| undefined | 1 | 1 | 1 | 1 | 0 |
| a function returning ClassThisAssignmentBlock | 1 | 1 | 1 | 1 | 0 |
| a function returning ExpressionWithTypeArguments & { readonly expression: Identifier \| PropertyAccessEntityNameExpression; } | 1 | 1 | 1 | 1 | 0 |
| a function returning HasJSDoc \| undefined | 1 | 1 | 1 | 1 | 0 |
| a function returning MemberName \| (Expression & (NumericLiteral \| StringLiteralLike)) | 1 | 1 | 1 | 1 | 0 |
| a function returning T | 1 | 1 | 1 | 1 | 0 |
| a function returning T \| undefined | 1 | 1 | 1 | 1 | 0 |
| a function returning U \| undefined | 1 | 1 | 1 | 1 | 0 |
| a function returning any | 1 | 1 | 1 | 1 | 0 |
| a function taking BinaryOperator seen as one taking SyntaxKind (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 0 |
| a function taking Identifier seen as one taking Node (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking NodeArray<Expression> seen as one taking readonly Expression[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking NodeArray<TypeNode> \| undefined seen as one taking readonly TypeNode[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a generic instantiation whose type argument for T, constrained to Node, can't be read | 1 | 1 | 1 | 1 | 0 |
| a member typed any (167(0)) related to QualifiedName at the assertion site | 1 | 1 | 1 | 1 | 0 |
| a method in object destructuring | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (afterProgramCreate would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (afterProgramEmitAndDiagnostics would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (base64decode would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (base64encode would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (convertToArrayAssignmentElement would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (convertToObjectAssignmentElement would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (enableCPUProfiler would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (getCanonicalFileName would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (onWatchStatusChange would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables | 1 | 1 | 1 | 1 | 0 |
| a parameter that isn't a plain name | 1 | 1 | 1 | 1 | 0 |
| a presence test of a method whose presence the receiver's construction doesn't prove; a view or host value has to read it | 2 | 2 | 2 | 2 | 0 |
| a read narrowed to undefined of a binding a call may write | 1 | 1 | 1 | 1 | 0 |
| a read narrowed to undefined of a property a call may write | 1 | 1 | 1 | 1 | 0 |
| a read of a void value, as a returned value | 1 | 1 | 1 | 1 | 0 |
| a spread argument T's type argument can be inferred from, judged against its constraint unknown[] | 1 | 1 | 1 | 1 | 0 |
| a structural relation deeper than 64 type pairs while checking promise-returning functions seen as void-returning | 3 | 3 | 3 | 3 | 0 |
| a template interpolating an object, an array, a map, a function or undefined | 1 | 1 | 1 | 1 | 0 |
| a tuple literal leaving out an element of type ModuleSpecifierEnding | 1 | 1 | 1 | 1 | 0 |
| a value of type () => { diagnosticMessage: DiagnosticMessage; errorNode: ExportAssignment; } seen as GetSymbolAccessibilityDiagnostic, which can write Node where ExportAssignment is read | 1 | 1 | 1 | 1 | 0 |
| a value of type () => { diagnosticMessage: DiagnosticMessage; errorNode: ExpressionWithTypeArguments; typeName: Identifier \| undefined; } seen as GetSymbolAccessibilityDiagnostic, which can write Node where ExpressionWithTypeArguments is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (ConstructorDeclaration & { body: Block; }) \| undefined | 1 | 1 | 1 | 1 | 0 |
| a value of type (IndexSignatureDeclaration \| PropertyDeclaration)[] \| undefined seen as ClassElement[] \| undefined, which can write ClassElement where IndexSignatureDeclaration \| PropertyDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (ModuleDeclaration & { name: StringLiteral; }) \| undefined | 1 | 1 | 1 | 1 | 0 |
| a value of type (VariableDeclaration & { name: Identifier; }) \| undefined | 1 | 1 | 1 | 1 | 0 |
| a value of type (symbol: Symbol) => { visitedTypes: Type[]; visitedSymbols: Symbol[]; } seen as (root: Symbol) => { visitedTypes: readonly Type[]; visitedSymbols: readonly Symbol[]; }, which can write readonly Type[] where Type[] is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (symbolAccessibilityResult: SymbolAccessibilityResult) => { diagnosticMessage: DiagnosticMessage; errorNode: DeclarationDiagnosticProducing; typeName: DeclarationName \| undefined; } \| undefined seen as (symbolAccessibilityResult: SymbolAccessibilityResult) => SymbolAccessibilityDiagnostic \| undefined, which can write Node where DeclarationDiagnosticProducing is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (type: Type) => { visitedTypes: Type[]; visitedSymbols: Symbol[]; } seen as (root: Type) => { visitedTypes: readonly Type[]; visitedSymbols: readonly Symbol[]; }, which can write readonly Type[] where Type[] is read | 1 | 1 | 1 | 1 | 0 |
| a value of type ArrayLiteralExpression \| AssignmentExpression<EqualsToken> \| BindingElement \| ElementAccessExpression \| ... 8 more ... \| VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type AssignmentExpression<EqualsToken> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 5 | 5 | 5 | 5 | 0 |
| a value of type BindingOrAssignmentElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type Block seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type BuildOptions seen as BuildOptions \| CompilerOptions, which can write TsConfigSourceFile \| CompilerOptionsValue where CompilerOptionsValue is read | 5 | 5 | 5 | 5 | 0 |
| a value of type CallExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type ClassDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type CompilerHost seen as CompilerHostLikeForCache, which can write WriteFileCallback \| undefined where WriteFileCallback is read | 1 | 1 | 1 | 1 | 0 |
| a value of type ComputedPropertyName seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ConditionalExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type DestructuringAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 11 | 11 | 11 | 11 | 0 |
| a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where SourceFile is read | 12 | 12 | 12 | 12 | 0 |
| a value of type DiagnosticWithLocation[] seen as Diagnostic[], which can write Diagnostic where DiagnosticWithLocation is read | 1 | 1 | 1 | 1 | 0 |
| a value of type EnumDeclaration seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type EnumDeclaration \| ModuleDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Expression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type Expression \| GeneratedIdentifier seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ExpressionStatement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 6 | 6 | 6 | 6 | 0 |
| a value of type ExpressionWithTypeArguments & { readonly expression: Identifier \| PropertyAccessEntityNameExpression; } | 1 | 1 | 1 | 1 | 0 |
| a value of type Extension[] seen as string[], which can write string where Extension is read | 1 | 1 | 1 | 1 | 0 |
| a value of type FutureSourceFile \| SourceFile seen as Pick<SourceFile, "fileName" \| "impliedNodeFormat" \| "packageJsonScope">, whose readonly field fileName becomes writable: a readonly field may hold something narrower than string, which a write of string would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type FutureSourceFile \| SourceFile seen as Pick<SourceFile, "fileName" \| "impliedNodeFormat">, whose readonly field fileName becomes writable: a readonly field may hold something narrower than string, which a write of string would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type HasJSDoc | 1 | 1 | 1 | 1 | 0 |
| a value of type HasJSDoc \| undefined | 1 | 1 | 1 | 1 | 0 |
| a value of type Identifier seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Identifier seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type JSDocArray | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocImportTag \| CanHaveModuleSpecifier | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<string, Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>> seen as ScriptTargetFeatures, which can write string where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<string, never[]> seen as Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>, which can write string[] where never[] is read | 10 | 10 | 10 | 10 | 0 |
| a value of type Map<string, string[] \| never[]> seen as Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>, which can write string where never is read | 3 | 3 | 3 | 3 | 0 |
| a value of type MapLike<string[]> seen as MapLike<readonly string[]>, which can write readonly string[] where string[] is read | 2 | 2 | 2 | 2 | 0 |
| a value of type ModuleBlock seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ModuleDeclaration seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ModuleDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ModuleName seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ModuleSpecifierResolutionHost & ModuleResolutionHost seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where (() => boolean) & (boolean \| (() => boolean) \| undefined) is read | 2 | 2 | 2 | 2 | 0 |
| a value of type Node seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type Node \| TextRange seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type NodeArray<ClassElement> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type NodeArray<Expression> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type NodeArray<ModifierLike> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type NodeArray<Statement> seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type NodeArray<Statement> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 7 | 7 | 7 | 7 | 0 |
| a value of type NodeArray<T> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ObjectBindingOrAssignmentPattern seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ParameterDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type ParameterDeclaration \| VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ParameterDeclaration \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ParameterPropertyDeclaration | 1 | 1 | 1 | 1 | 0 |
| a value of type ParenthesizedExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 7 | 7 | 7 | 7 | 0 |
| a value of type PropertyAccessExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type PropertyAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type PropertyDeclaration[] \| undefined seen as ClassElement[] \| undefined, which can write ClassElement where PropertyDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type RequireOrImportCall | 1 | 1 | 1 | 1 | 0 |
| a value of type SortedArray<DiagnosticWithLocation> \| undefined seen as Diagnostic[], which can write Diagnostic where DiagnosticWithLocation is read | 1 | 1 | 1 | 1 | 0 |
| a value of type SortedArray<DiagnosticWithLocation> \| undefined seen as SortedArray<Diagnostic> \| undefined, which can write Diagnostic where DiagnosticWithLocation is read | 2 | 2 | 2 | 2 | 0 |
| a value of type Statement[] seen as Node[], which can write Node where Statement is read | 1 | 1 | 1 | 1 | 0 |
| a value of type StringLiteral seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type T seen as Mutable<T>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of T["pos"] would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type TsConfigSourceFile \| CompilerOptionsValue seen as string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined, which can write string \| number where string is read | 1 | 1 | 1 | 1 | 0 |
| a value of type TypeParameterDeclaration & { parent: JSDocTemplateTag; } | 1 | 1 | 1 | 1 | 0 |
| a value of type VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type VariableDeclaration \| DestructuringAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type VariableDeclarationList seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type VariableStatement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type VariableStatement[] seen as (ExportDeclaration \| VariableStatement)[], which can write ExportDeclaration \| VariableStatement where VariableStatement is read | 1 | 1 | 1 | 1 | 0 |
| a value of type any: any has no proven type | 4 | 4 | 4 | 4 | 0 |
| a value of type never | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as Declaration[], which can write Declaration where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as StringLiteralLike[], which can write StringLiteralLike where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as string[] \| never[], which can write string where never is read | 3 | 3 | 3 | 3 | 0 |
| a value of type never[] seen as string[], which can write string where never is read | 2 | 2 | 2 | 2 | 0 |
| a value of type never[] \| SortedArray<DiagnosticWithLocation> seen as Diagnostic[], which can write Diagnostic where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[][] seen as string[][], which can write string where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type readonly Extension[][] seen as readonly string[][], which can write string where Extension is read | 3 | 3 | 3 | 3 | 0 |
| a value of type readonly string[] \| undefined seen as RegExp[] \| undefined, which can write RegExp where string is read | 1 | 1 | 1 | 1 | 0 |
| a value of type { (fileName: string): DiagnosticWithLocation[]; (): Diagnostic[]; } seen as { (): Diagnostic[]; (fileName: string): DiagnosticWithLocation[]; }, which can write Diagnostic where DiagnosticWithLocation is read | 1 | 1 | 1 | 1 | 0 |
| a value of type { ending: ModuleSpecifierEnding; value: string; }[] seen as { ending: ModuleSpecifierEnding \| undefined; value: string; }[], which can write ModuleSpecifierEnding \| undefined where ModuleSpecifierEnding is read | 1 | 1 | 1 | 1 | 0 |
| a { allowImportingTsExtensions: "rewriteRelativeImportExtensions"[]; target: never[]; module: "target"[]; moduleResolution: ("module" \| "target")[]; moduleDetection: ("module" \| "target")[]; isolatedModules: ...; ... 19 more ...; useUnknownInCatchVariables: ...; } bound to T, whose constraint Record<string, ("all" \| "allowArbitraryExtensions" \| "allowImportingTsExtensions" \| "allowJs" \| "allowNonTsExtensions" \| "allowSyntheticDefaultImports" \| "allowUmdGlobalAccess" \| ... 129 more ... \| "watch")[]> is stored otherwise (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a { allowImportingTsExtensions: { dependencies: "rewriteRelativeImportExtensions"[]; computeValue: (compilerOptions: Pick<CompilerOptions, "allowImportingTsExtensions" \| "rewriteRelativeImportExtensions">) => boolean; }; ... 24 more ...; useUnknownInCatchVariables: ...; } seen as Record<string, { dependencies: readonly string[]; computeValue: (options: CompilerOptions) => CompilerOptionsValue; }> (fixed objects and records have different storage; copy explicitly) | 2 | 2 | 2 | 2 | 0 |
| an argument to a rest parameter that isn't an array, which T's type argument can be inferred from, judged against its constraint unknown[] | 1 | 1 | 1 | 1 | 0 |
| an array of RequireOrImportCall | 1 | 1 | 1 | 1 | 0 |
| an array of unknown with erased element storage (retain its declared element type before reading elements) | 1 | 1 | 1 | 1 | 0 |
| an object refinement using an open numeric enum as a literal tag | 107 | 107 | 78 | 78 | 0 |
| an optional Map get through a structural receiver | 1 | 1 | 1 | 1 | 0 |
| an optional RegExp call | 1 | 1 | 1 | 1 | 0 |
| an optional chain longer than one step | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter callback (argument "decl => isVariableDeclarationInitializedToRequire(decl)") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter callback (argument "n => isExportAssignment(n) \|\| isStatement(n) \|\| isVariableDeclaration(n) \|\| isPropertyDeclaration(n) \|\| isParameter(n)") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter f (argument "p => isParameterPropertyDeclaration(p, constructor)") | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter test (argument "createAssignmentCallback ? isIdentifier : isExpression") | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from ImportEqualsDeclaration & { moduleReference: ExternalModuleReference; } to ImportEqualsDeclaration: optional field moduleReference.id has no proven compatible presence/type | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from LiteralLikeNode to TemplateLiteralLikeNode: optional field rawText has no proven compatible presence/type | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | 2 | 2 | 2 | 2 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.BindingElement | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ComputedPropertyName | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Identifier | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ImportDeclaration | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.JSDocTypeExpression | 2 | 2 | 2 | 2 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NamespaceExport | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NamespaceImport | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ObjectLiteralExpression | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.PropertyAssignment | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.StringLiteral | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Unknown \| SyntaxKind.EndOfFileToken \| SyntaxKind.SingleLineCommentTrivia \| SyntaxKind.MultiLineCommentTrivia \| SyntaxKind.NewLineTrivia \| SyntaxKind.WhitespaceTrivia \| SyntaxKind.ShebangTrivia \| ... 333 more ... \| SyntaxKind.Count | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Unknown \| SyntaxKind.EndOfFileToken \| SyntaxKind.SingleLineCommentTrivia \| SyntaxKind.MultiLineCommentTrivia \| SyntaxKind.NewLineTrivia \| SyntaxKind.WhitespaceTrivia \| SyntaxKind.ShebangTrivia \| ... 334 more ... \| SyntaxKind.Count | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Unknown \| SyntaxKind.EndOfFileToken \| SyntaxKind.SingleLineCommentTrivia \| SyntaxKind.MultiLineCommentTrivia \| SyntaxKind.NewLineTrivia \| SyntaxKind.WhitespaceTrivia \| SyntaxKind.ShebangTrivia \| ... 339 more ... \| SyntaxKind.Count | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Unknown \| SyntaxKind.EndOfFileToken \| SyntaxKind.SingleLineCommentTrivia \| SyntaxKind.MultiLineCommentTrivia \| SyntaxKind.NewLineTrivia \| SyntaxKind.WhitespaceTrivia \| SyntaxKind.ShebangTrivia \| ... 345 more ... \| SyntaxKind.Count | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Unknown \| SyntaxKind.NumericLiteral \| SyntaxKind.BigIntLiteral \| SyntaxKind.StringLiteral \| SyntaxKind.JsxText \| SyntaxKind.RegularExpressionLiteral \| SyntaxKind.NoSubstitutionTemplateLiteral | 1 | 1 | 1 | 1 | 0 |
| assigning to a NonNullExpression | 1 | 1 | 1 | 1 | 0 |
| checked predicate membership over an open tag domain | 2 | 2 | 2 | 2 | 0 |
| checked predicate overload target AssignmentExpression<AmpersandAmpersandEqualsToken \| BarBarEqualsToken \| EqualsToken \| QuestionQuestionEqualsToken> & ... | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target BindableObjectDefinePropertyCall | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target BindableStaticElementAccessExpression | 2 | 2 | 2 | 2 | 0 |
| checked predicate overload target ClassThisAssignmentBlock | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target ElementAccessExpression & Declaration & { readonly argumentExpression: NumericLiteral \| StringLiteralLike; } & ... | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target ImportEqualsDeclaration & { moduleReference: ExternalModuleReference; } | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target LiteralImportTypeNode | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target LiteralLikeElementAccessExpression | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target MissingPackageJsonInfo | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target ModuleResolutionHost | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target NamedDeclaration & { name: DeclarationName; } | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target ParameterPropertyDeclaration | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target PrefixUnaryExpression & { operand: NumericLiteral; } | 1 | 1 | 1 | 1 | 0 |
| checked view field modifiers of type NodeArray<Modifier> \| NodeArray<ModifierLike> \| undefined | 1 | 1 | 1 | 1 | 0 |
| indexOf on a tuple (a tuple is held as an object, not an array, so far; write it as an array where it's made) | 1 | 1 | 1 | 1 | 0 |
| inherited library member get read as an own field | 1 | 1 | 1 | 1 | 0 |
| new an Identifier | 1 | 1 | 1 | 1 | 0 |
| optional chaining to .size on a value | 1 | 1 | 1 | 1 | 0 |
| optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields | 7 | 7 | 7 | 7 | 0 |
| optional property id in ArrayBindingPattern absent from structural source never, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property id in BigIntLiteral absent from structural source never, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property id in Expression absent from structural source never, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property id in FalseLiteral absent from structural source never, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property id in Identifier absent from structural source never, which can hide fields | 3 | 3 | 3 | 3 | 0 |
| optional property id in LiteralExpression & StringLiteral absent from structural source never, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property id in Node absent from structural source never, which can hide fields | 12 | 12 | 12 | 12 | 0 |
| optional property modifiers in ArrowFunction absent from structural source never, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property modifiers in GetAccessorDeclaration absent from structural source never, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property modifiers in SetAccessorDeclaration absent from structural source never, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property packageJsonScope in Pick<SourceFile, "fileName" \| "impliedNodeFormat" \| "packageJsonScope"> absent from structural source Pick<SourceFile, "fileName" \| "impliedNodeFormat">, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property packageRootPath in { moduleFileToTry: string; packageRootPath?: string; blockedByExports?: true; verbatimFromExports?: true; } absent from structural source { moduleFileToTry: string; }, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property reportsUnnecessary in Diagnostic absent from structural source DiagnosticRelatedInformation, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property source in SourceMapRange absent from structural source TextRange, which can hide fields | 4 | 4 | 4 | 4 | 0 |
| optional property templateFlags in NoSubstitutionTemplateLiteral absent from structural source never, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| overload 1 of arrayFrom result U[] cannot be served by implementation result (T \| U)[] | 1 | 1 | 1 | 1 | 0 |
| overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined | 1 | 1 | 1 | 1 | 0 |
| overload 1 of getDiagnostics result DiagnosticWithLocation[] cannot be served by implementation result Diagnostic[] | 1 | 1 | 1 | 1 | 0 |
| overload 1 of getSupportedExtensionsWithJsonIfResolveJsonModule parameter supportedExtensions cannot be served by implementation parameter supportedExtensions | 1 | 1 | 1 | 1 | 0 |
| overload 1 of or with a non-array rest parameter | 1 | 1 | 1 | 1 | 0 |
| overload 1 of reduceLeft parameter f cannot be served by implementation parameter f | 1 | 1 | 1 | 1 | 0 |
| overload 1 of visitNode parameter node cannot be served by implementation parameter node | 1 | 1 | 1 | 1 | 0 |
| overload 2 of getParseTreeNode result T \| undefined cannot be served by implementation result Node \| undefined | 1 | 1 | 1 | 1 | 0 |
| overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor | 1 | 1 | 1 | 1 | 0 |
| spreading an array of other elements | 1 | 1 | 1 | 1 | 0 |
| var | 2 | 2 | 2 | 2 | 0 |
| writing a checked object field that may hold undefined or null, whose checked write can't change the representation its slot holds | 11 | 11 | 11 | 11 | 0 |
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
