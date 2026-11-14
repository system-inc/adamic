# Paired latent census

| Measure | Raw findings | Distinct reads |
| --- | ---: | ---: |
| Lowered (no remaining finding) | 0 | 0 |
| Moved by reason | 0 | 0 |
| New findings | 0 | 0 |
| New sites | 0 | 0 |
| Units whose first stop moved | 0 | 0 |
| Regressions | 0 | 0 |
| before NotYet | 116 | 116 |
| before Refused | 649 | 645 |
| after NotYet | 116 | 116 |
| after Refused | 649 | 645 |

| Reason | Raw before | Raw after | Reads before | Reads after | Reads lowered |
| --- | ---: | ---: | ---: | ---: | ---: |
| Object enumeration with getter or accessor descriptors | 2 | 2 | 2 | 2 | 0 |
| Reflect.get by a run-time key in a program with a class getter that does more than return its fields and literals joined by operators (the getter runs inside the read, unseen by the passes that follow calls) | 1 | 1 | 1 | 1 | 0 |
| RegExp with a nonconstant pattern | 1 | 1 | 1 | 1 | 0 |
| a BinaryExpression with a string and a union of differently held members | 1 | 1 | 1 | 1 | 0 |
| a Map or Set going where a plain object type is held | 1 | 1 | 1 | 1 | 0 |
| a Map or Set interface implemented by a value that isn't a Map or Set | 3 | 3 | 3 | 3 | 0 |
| a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions | 1 | 1 | 1 | 1 | 0 |
| a call spreading elements that cannot be packed | 1 | 1 | 1 | 1 | 0 |
| a call through ?. (an optional call) | 1 | 1 | 1 | 1 | 0 |
| a cast that makes a viewed readonly array writable | 2 | 2 | 2 | 2 | 0 |
| a cast the runtime can't check | 115 | 115 | 115 | 115 | 0 |
| a checked predicate target whose base has no required field to check (CompilerOptions & { types: string[]; }) | 1 | 1 | 1 | 1 | 0 |
| a checked predicate target whose check would read past the depth bound (PackageJsonInfo) | 1 | 1 | 1 | 1 | 0 |
| a class used from a function that may run before the class's declaration (its temporal dead zone) | 1 | 1 | 1 | 1 | 0 |
| a comparator that doesn't take two elements and return a number | 1 | 1 | 1 | 1 | 0 |
| a computed field name | 2 | 2 | 2 | 2 | 0 |
| a destructured name held otherwise than its field | 1 | 1 | 1 | 1 | 0 |
| a dictionary member narrowed to an array without an element-checked view | 6 | 6 | 6 | 6 | 0 |
| a function returning (AssignmentExpression<EqualsToken> & { readonly left: GeneratedIdentifier; }) \| undefined | 1 | 1 | 1 | 1 | 0 |
| a function returning ClassNamedEvaluationHelperBlock | 1 | 1 | 1 | 1 | 0 |
| a function returning T \| undefined | 2 | 2 | 2 | 2 | 0 |
| a function taking ((node: Node) => VisitResult<Node>) \| undefined seen as one taking ((node: Node) => VisitResult<Node \| undefined>) \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 0 |
| a function taking BinaryOperatorToken seen as one taking BinaryOperatorToken \| BinaryOperator (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking Expression seen as one taking Expression \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking Expression[] seen as one taking readonly Expression[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 0 |
| a function taking JSDocTypeExpression \| undefined seen as one taking JSDocTypeExpression \| JSDocTypeLiteral \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 0 |
| a function taking NodeArray<JSDocComment> \| undefined seen as one taking string \| NodeArray<JSDocComment> \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 26 | 26 | 26 | 26 | 0 |
| a function taking PollingInterval seen as one taking number \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking boolean seen as one taking boolean \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 0 |
| a function taking readonly ModifierLike[] seen as one taking readonly ModifierLike[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking readonly Modifier[] \| ModifierFlags seen as one taking readonly Modifier[] \| ModifierFlags \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking readonly Modifier[] \| undefined seen as one taking readonly ModifierLike[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 4 | 4 | 4 | 4 | 0 |
| a function taking readonly ParameterDeclaration[] seen as one taking readonly ParameterDeclaration[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking string seen as one taking string \| MemberName (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a host object of type Performance seen as PerformanceTime \| undefined, which names , a host method its runtime shape doesn't hold (call it on the host's own type, or wrap it in a closure) | 1 | 1 | 1 | 1 | 0 |
| a literal method through a view that erases its receiver | 1 | 1 | 1 | 1 | 0 |
| a member typed any (167(0)) related to QualifiedName at the assertion site | 1 | 1 | 1 | 1 | 0 |
| a method in object destructuring | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (createComma would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
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
| a method read as a value (createUnionTypeNode would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getCanonicalFileName would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getEnvironmentVariable would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (globalCacheResolutionModuleName would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (liftToBlock would lose its object, and this with it) | 4 | 4 | 4 | 4 | 0 |
| a method read as a value (parenthesizeConstituentTypesOfIntersectionType would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (parenthesizeConstituentTypesOfUnionType would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (writeFile would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method sharing its name with a getter; rename the getter or use a callable arrow field in the interface and implementation | 2 | 2 | 2 | 2 | 0 |
| a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports | 1 | 1 | 1 | 1 | 0 |
| a rest parameter that isn't an array | 1 | 1 | 1 | 1 | 0 |
| a tagged template other than the intrinsic String.raw | 2 | 2 | 2 | 2 | 0 |
| a union or optional field read in a program with record storage | 3 | 3 | 3 | 3 | 0 |
| a union or optional tuple length; declare arrays where the values are constructed, or carry an explicit length alongside each tuple | 1 | 1 | 1 | 1 | 0 |
| a value of type () => { value: Mapping; done: false; } \| { value: never; done: true; } seen as (...[value]: [] \| [any]) => IteratorResult<Mapping, any>, which can write any where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; } & BinaryExpression) \| (... & ... 1 more ... & BinaryExpression) | 1 | 1 | 1 | 1 | 0 |
| a value of type (EmitNode & { autoGenerate: AutoGenerateInfo; }) \| (EmitNode & { autoGenerate: AutoGenerateInfo; }) | 1 | 1 | 1 | 1 | 0 |
| a value of type (baseDir: string, moduleName: string) => { module: any; modulePath: string; error: undefined; } \| { module: undefined; modulePath: undefined; error: unknown; } seen as (baseDir: string, moduleName: string) => ModuleImportResult, which can write string \| undefined where string is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (identifierOrPrivateName: Identifier \| PrivateIdentifier) => string seen as (name: GeneratedIdentifier \| GeneratedPrivateIdentifier) => string, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type (left: MappedPosition, right: MappedPosition) => boolean seen as EqualityComparer<SourceMappedPosition> \| undefined, which can write string \| undefined where string is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (recordTempVariable: ((node: Identifier) => void) \| undefined, reservedInNestedScopes?: boolean \| undefined, prefix?: string \| GeneratedNamePart \| undefined, suffix?: string \| undefined) => GeneratedIdentifier seen as { (recordTempVariable: ((node: Identifier) => void) \| undefined, reservedInNestedScopes?: boolean \| undefined): Identifier; (recordTempVariable: ((node: Identifier) => void) \| undefined, reservedInNestedScopes?: boolean \| undefined, prefix?: string \| ... 1 more ... \| undefined, suffix?: string \| undefined): Identifi..., whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type AccessorDeclaration & { readonly name: BigIntLiteral \| ComputedPropertyName \| Identifier \| NoSubstitutionTemplateLiteral \| NumericLiteral \| StringLiteral; } | 1 | 1 | 1 | 1 | 0 |
| a value of type ArrayLiteralExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type AssertClause seen as Mutable<AssertClause>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type AssertEntry seen as Mutable<AssertEntry>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type AssignmentExpression<EqualsToken> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 8 | 8 | 8 | 8 | 0 |
| a value of type BigIntLiteral \| NoSubstitutionTemplateLiteral \| NumericLiteral \| StringLiteral seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type BindingElement & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | 1 | 1 | 0 |
| a value of type Block seen as Mutable<Block>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type BreakStatement seen as Mutable<BreakStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type CallExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 4 | 4 | 4 | 4 | 0 |
| a value of type CallSignatureDeclaration seen as Mutable<CallSignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type CaseBlock seen as Mutable<CaseBlock>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ClassExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ClassStaticBlockDeclaration seen as Mutable<ClassStaticBlockDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type CompilerHost \| ResolutionCacheHost seen as ModuleResolutionHost, whose method useCaseSensitiveFileNames becomes a writable field: a write of boolean \| (() => boolean) \| undefined would replace a method of type () => boolean | 9 | 9 | 9 | 9 | 0 |
| a value of type CompilerOptions & { types: string[]; } seen as CompilerOptions, which can write string[] \| undefined where string[] is read | 2 | 2 | 2 | 2 | 0 |
| a value of type ConstructSignatureDeclaration seen as Mutable<ConstructSignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ConstructorDeclaration seen as Mutable<ConstructorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type ConstructorTypeNode seen as Mutable<ConstructorTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ContinueStatement seen as Mutable<ContinueStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type DiagnosticWithDetachedLocation seen as Diagnostic, which can write SourceFile \| undefined where undefined is read | 4 | 4 | 4 | 4 | 0 |
| a value of type DiagnosticWithDetachedLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where undefined is read | 4 | 4 | 4 | 4 | 0 |
| a value of type DiagnosticWithLocation[] seen as Diagnostic[], which can write Diagnostic where DiagnosticWithLocation is read | 2 | 2 | 2 | 2 | 0 |
| a value of type ElementAccessExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type ExportAssignment & { readonly expression: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | 1 | 1 | 0 |
| a value of type Expression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 10 | 10 | 10 | 10 | 0 |
| a value of type Expression \| Identifier seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ExpressionStatement seen as Mutable<ExpressionStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ExpressionStatement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type FunctionExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type FunctionTypeNode seen as Mutable<FunctionTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GeneratedIdentifier seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type GeneratedIdentifier seen as Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GeneratedIdentifier seen as Node \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as GeneratedIdentifier \| GeneratedPrivateIdentifier \| Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GeneratedIdentifier \| Identifier seen as Identifier \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GeneratedIdentifier \| Identifier seen as Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type GeneratedPrivateIdentifier seen as Node \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GetAccessorDeclaration \| SetAccessorDeclaration seen as Mutable<AccessorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Identifier seen as Mutable<Identifier>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type Identifier seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Identifier seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type Identifier[] \| undefined seen as ModuleExportName[] \| undefined, which can write ModuleExportName where Identifier is read | 1 | 1 | 1 | 1 | 0 |
| a value of type ImportAttribute seen as Mutable<ImportAttribute>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ImportAttributes seen as Mutable<ImportAttributes>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ImportClause seen as Mutable<ImportClause>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ImportDeclaration seen as Mutable<ImportDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ImportTypeAssertionContainer seen as Mutable<ImportTypeAssertionContainer>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type ImportTypeNode seen as Mutable<ImportTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type IndexSignatureDeclaration seen as Mutable<IndexSignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocAugmentsTag seen as Mutable<JSDocAugmentsTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocCallbackTag seen as Mutable<JSDocCallbackTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocFunctionType seen as Mutable<JSDocFunctionType>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocImplementsTag seen as Mutable<JSDocImplementsTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocImportTag seen as Mutable<JSDocImportTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocLink seen as Mutable<JSDocLink>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocLinkCode seen as Mutable<JSDocLinkCode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocLinkPlain seen as Mutable<JSDocLinkPlain>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocNameReference seen as Mutable<JSDocNameReference>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocOverloadTag seen as Mutable<JSDocOverloadTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocParameterTag seen as Mutable<JSDocParameterTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocPropertyTag seen as Mutable<JSDocPropertyTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocReturnTag seen as JSDocTag & { typeExpression?: JSDocTypeExpression; }, whose readonly field typeExpression becomes writable: a readonly field may hold something narrower than JSDocTypeExpression \| undefined, which a write of JSDocTypeExpression \| undefined would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type JSDocSatisfiesTag seen as JSDocTag & { typeExpression?: JSDocTypeExpression; }, whose readonly field typeExpression becomes writable: a readonly field may hold something narrower than JSDocTypeExpression, which a write of JSDocTypeExpression \| undefined would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type JSDocSeeTag seen as Mutable<JSDocSeeTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocSignature seen as Mutable<JSDocSignature>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocTemplateTag seen as Mutable<JSDocTemplateTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocText seen as Mutable<JSDocText>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocThisTag seen as JSDocTag & { typeExpression?: JSDocTypeExpression; }, whose readonly field typeExpression becomes writable: a readonly field may hold something narrower than JSDocTypeExpression, which a write of JSDocTypeExpression \| undefined would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type JSDocThrowsTag seen as JSDocTag & { typeExpression?: JSDocTypeExpression; }, whose readonly field typeExpression becomes writable: a readonly field may hold something narrower than JSDocTypeExpression \| undefined, which a write of JSDocTypeExpression \| undefined would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type JSDocTypeExpression seen as Mutable<JSDocTypeExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocTypeLiteral seen as Mutable<JSDocTypeLiteral>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocTypeTag seen as JSDocTag & { typeExpression?: JSDocTypeExpression; }, whose readonly field typeExpression becomes writable: a readonly field may hold something narrower than JSDocTypeExpression, which a write of JSDocTypeExpression \| undefined would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type JSDocTypedefTag seen as Mutable<JSDocTypedefTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocUnknownTag seen as Mutable<JSDocUnknownTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type LeftHandSideExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type Map<Path, DirectoryWatchesOfFailedLookup> seen as Map<string, DirectoryWatchesOfFailedLookup>, which can write string where Path is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<Path, ModeAwareCache<CachedResolvedModuleWithFailedLookupLocations>> seen as Map<string, ModeAwareCache<CachedResolvedModuleWithFailedLookupLocations>>, which can write string where Path is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<Path, ModeAwareCache<CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations>> seen as Map<string, ModeAwareCache<CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations>>, which can write string where Path is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<Path, string[]> seen as InvokeMap, which can write true \| string[] where string[] is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Map<string, CachedResolvedModuleWithFailedLookupLocations> seen as Map<string, ResolutionWithFailedLookupLocations> \| Set<ResolutionWithFailedLookupLocations> \| undefined, which can write ResolutionWithFailedLookupLocations where CachedResolvedModuleWithFailedLookupLocations is read | 1 | 1 | 1 | 1 | 0 |
| a value of type MappedTypeNode seen as Mutable<MappedTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type MethodDeclaration seen as Mutable<MethodDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type MissingDeclaration seen as Mutable<MissingDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type ModifierLike seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type ModuleResolutionState & { host: GetPackageJsonEntrypointsHost; } | 1 | 1 | 1 | 1 | 0 |
| a value of type ModuleResolutionState & { host: GetPackageJsonEntrypointsHost; } seen as ModuleResolutionState, which can write ModuleResolutionHost where ModuleResolutionHost & GetPackageJsonEntrypointsHost is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<GeneratedIdentifier> seen as Identifier \| PrivateIdentifier, which can write EmitNode \| undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 2 | 2 | 2 | 2 | 0 |
| a value of type Mutable<GeneratedIdentifier> seen as Identifier, which can write EmitNode \| undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 4 | 4 | 4 | 4 | 0 |
| a value of type Mutable<GeneratedIdentifier> seen as Node, which can write EmitNode \| undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<GeneratedPrivateIdentifier> seen as Identifier \| PrivateIdentifier, which can write EmitNode \| undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<GeneratedPrivateIdentifier> seen as Node, which can write EmitNode \| undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<NoSubstitutionTemplateLiteral> seen as Mutable<TemplateLiteralLikeNode>, which can write SyntaxKind where SyntaxKind.NoSubstitutionTemplateLiteral is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<T> seen as T, a type parameter whose constraint JSDocType & { readonly type: TypeNode \| undefined; readonly postfix: boolean; } can be written, so it can write what Mutable<T> can't hold | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<T> seen as T, a type parameter whose constraint JSDocType & { readonly type: TypeNode \| undefined; } can be written, so it can write what Mutable<T> can't hold | 1 | 1 | 1 | 1 | 0 |
| a value of type Mutable<T> seen as T, a type parameter whose constraint Node can be written, so it can write what Mutable<T> can't hold | 1 | 1 | 1 | 1 | 0 |
| a value of type NamedImports seen as Mutable<NamedImports>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type NamespaceExport seen as Mutable<NamespaceExport>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type NamespaceExportDeclaration seen as Mutable<NamespaceExportDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type NamespaceImport seen as Mutable<NamespaceImport>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Node seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type NodeArray<ClassElement> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type NodeArray<ExpressionStatement> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type NodeArray<Statement> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type NodeArray<never> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type NonNullExpression seen as Mutable<NonNullExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type OptionalTypeNode seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type OptionalTypeNode seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ParameterDeclaration & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | 1 | 1 | 0 |
| a value of type PropertyAccessExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type PropertyAssignment & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | 1 | 1 | 0 |
| a value of type PropertyAssignment seen as Mutable<PropertyAssignment \| ShorthandPropertyAssignment>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type PropertyAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type PropertyDeclaration & { readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | 1 | 1 | 0 |
| a value of type PropertySignature seen as Mutable<PropertySignature>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type Readonly<BuilderState> \| undefined seen as BuilderState \| undefined, whose readonly field fileInfos becomes writable: a readonly field may hold something narrower than Map<Path, FileInfo>, which a write of Map<Path, FileInfo> would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Readonly<PathPathComponents<Path>> | 7 | 7 | 7 | 7 | 0 |
| a value of type Resolved \| undefined seen as { resolvedFileName: string; originalPath: string \| undefined; extension: string; isExternalLibraryImport: boolean; packageId: PackageId \| undefined; resolvedUsingTsExtension: boolean; } \| undefined, whose readonly field originalPath becomes writable: a readonly field may hold something narrower than string \| true \| undefined, which a write of string \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ResolvedModuleFull \| undefined seen as { path: string; originalPath: string \| true; extension: string; packageId: PackageId \| undefined; resolvedUsingTsExtension: boolean \| undefined; } \| undefined, whose readonly field originalPath becomes writable: a readonly field may hold something narrower than string \| undefined, which a write of string \| true would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ReturnStatement seen as Mutable<ReturnStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type SearchResult<Resolved> seen as { value: { resolved: Resolved; isExternalLibraryImport: true; } \| undefined; } \| undefined, whose readonly field value becomes writable: a readonly field may hold something narrower than Resolved \| undefined, which a write of { resolved: Resolved; isExternalLibraryImport: true; } \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Set<Path> \| undefined seen as Set<string> \| undefined, which can write string where Path is read | 1 | 1 | 1 | 1 | 0 |
| a value of type SetAccessorDeclaration seen as Mutable<SetAccessorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ShorthandPropertyAssignment & { readonly objectAssignmentInitializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | 1 | 1 | 0 |
| a value of type ShorthandPropertyAssignment seen as Mutable<PropertyAssignment \| ShorthandPropertyAssignment>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type SourceFile seen as EmitNode \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of EmitFlags would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type SourceFile seen as Mutable<SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type SwitchStatement seen as Mutable<SwitchStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type System seen as ModuleResolutionHost, whose method directoryExists becomes a writable field: a write of ((directoryName: string) => boolean) \| undefined would replace a method of type (path: string) => boolean | 1 | 1 | 1 | 1 | 0 |
| a value of type T seen as Mutable<T>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of T["pos"] would replace | 9 | 9 | 9 | 9 | 0 |
| a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type T \| undefined | 1 | 1 | 1 | 1 | 0 |
| a value of type TaggedTemplateExpression seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type TemplateStringsArray | 1 | 1 | 1 | 1 | 0 |
| a value of type TypeOperatorNode seen as Mutable<TypeOperatorNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type TypeParameterDeclaration seen as Mutable<TypeParameterDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type VariableDeclaration & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | 1 | 1 | 0 |
| a value of type VariableStatement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type WrappedExpression<AnonymousFunctionDefinition> | 2 | 2 | 2 | 2 | 0 |
| a value of type YieldExpression seen as Mutable<YieldExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type any: any has no proven type | 5 | 5 | 5 | 5 | 0 |
| a value of type never | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as ChildDirectoryWatcher[], which can write ChildDirectoryWatcher where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as CommentRange[], which can write CommentRange where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as SourceFile[], which can write SourceFile where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as SourceMappedPosition[], which can write SourceMappedPosition where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as string[], which can write string where never is read | 2 | 2 | 2 | 2 | 0 |
| a value of type { arguments: { name?: string; } & { path: string; }; range: CommentRange; } \| { arguments: { name: string; }; range: CommentRange; } \| { arguments: { factory: string; }; range: CommentRange; } \| ... 5 more ... \| ... seen as ({ arguments: { name?: string; } & { path: string; }; range: CommentRange; } \| { arguments: { name: string; }; range: CommentRange; } \| { arguments: { factory: string; }; range: CommentRange; } \| ... 5 more ... \| ...)[] \| ... 8 more ... \| ..., which can write { name?: string; } & { path: string; } where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type { compilerOptions: CompilerOptions; traceEnabled: boolean; affectingLocations: string[] \| undefined; resultFromCache?: ResolvedModuleWithFailedLookupLocations; ... 9 more ...; host: GetPackageJsonEntrypointsHost; } seen as ModuleResolutionState & { host: GetPackageJsonEntrypointsHost; }, which can write string[] \| undefined where never[] is read | 1 | 1 | 1 | 1 | 0 |
| a value of type { name: string \| undefined; path: string; }[] seen as AmdDependency[], which can write AmdDependency where { name: string \| undefined; path: string; } is read | 1 | 1 | 1 | 1 | 0 |
| a value of type { resolvedFileName: string; originalPath: string \| undefined; extension: string; isExternalLibraryImport: boolean; packageId: PackageId \| undefined; resolvedUsingTsExtension: boolean; } \| undefined seen as ResolvedModuleFull \| undefined, which can write boolean \| undefined where boolean is read | 1 | 1 | 1 | 1 | 0 |
| a value of type { value: never; done: true; } seen as IteratorResult<Mapping, any>, which can write any where never is read | 14 | 14 | 14 | 14 | 0 |
| a value without nominal ancestry seen as BinaryExpressionStateMachine<TOuterState, TState, TResult> | 18 | 18 | 18 | 18 | 0 |
| a { all?: boolean; allowImportingTsExtensions?: boolean; allowJs?: boolean; allowNonTsExtensions?: boolean; allowArbitraryExtensions?: boolean; allowSyntheticDefaultImports?: boolean; allowUmdGlobalAccess?: boolean; ... 129 more ...; moduleResolution: ModuleResolutionKind; } seen as CompilerOptions (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a { host: ModuleResolutionHost; traceEnabled: boolean; failedLookupLocations: string[] \| undefined; affectingLocations: string[] \| undefined; resultFromCache?: ResolvedModuleWithFailedLookupLocations; ... 8 more ...; reportDiagnostic: ...; } seen as ModuleResolutionState (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| an array of any | 1 | 1 | 1 | 1 | 0 |
| an assignment used as a value without a known target write type | 1 | 1 | 1 | 1 | 0 |
| an assignment value to a member | 2 | 2 | 2 | 2 | 0 |
| an object refinement using an open numeric enum as a literal tag | 124 | 124 | 120 | 120 | 0 |
| an optional Map get through a structural receiver | 1 | 1 | 1 | 1 | 0 |
| an optional chain longer than one step | 1 | 1 | 1 | 1 | 0 |
| an overloaded function as a value | 1 | 1 | 1 | 1 | 0 |
| an uninitialized object field without a supported declared slot type | 1 | 1 | 1 | 1 | 0 |
| an unproven predicate argument for parameter f (argument "helper => !helper.scoped") | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from PackageJsonPathFields to PackageJson: optional field version has no proven compatible presence/type | 2 | 2 | 2 | 2 | 0 |
| an unproven value assigned to a numeric literal or enum member slot PunctuationOrKeywordSyntaxKind | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind | 5 | 5 | 5 | 5 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ExtendsKeyword \| SyntaxKind.ImplementsKeyword | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.FunctionExpression | 2 | 2 | 2 | 2 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ImportAttributes | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.JSDocTypeExpression | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.KeyOfKeyword \| SyntaxKind.ReadonlyKeyword \| SyntaxKind.UniqueKeyword | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NoSubstitutionTemplateLiteral \| SyntaxKind.TemplateHead \| SyntaxKind.TemplateMiddle \| SyntaxKind.TemplateTail | 2 | 2 | 2 | 2 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NumericLiteral \| SyntaxKind.BigIntLiteral \| SyntaxKind.StringLiteral \| SyntaxKind.JsxText \| SyntaxKind.JsxTextAllWhiteSpaces \| SyntaxKind.RegularExpressionLiteral \| SyntaxKind.NoSubstitutionTemplateLiteral | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.WithKeyword \| SyntaxKind.AssertKeyword | 2 | 2 | 2 | 2 | 0 |
| assigning an element of a value | 1 | 1 | 1 | 1 | 0 |
| assigning to an AsExpression | 1 | 1 | 1 | 1 | 0 |
| checked predicate membership over an open tag domain | 2 | 2 | 2 | 2 | 0 |
| checked predicate overload target BinaryExpression & { operatorToken: Token<SyntaxKind.CommaToken>; } | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target ClassNamedEvaluationHelperBlock | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target NamedDeclaration & { name: DeclarationName; } | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target SourceMapping | 1 | 1 | 1 | 1 | 0 |
| delete | 1 | 1 | 1 | 1 | 0 |
| destructuring a value | 2 | 2 | 2 | 2 | 0 |
| detached hasOwnProperty through an object view (an Error may be hidden by the view) | 1 | 1 | 1 | 1 | 0 |
| for...in over an array (holes and own enumerable properties are not represented; use for...of for elements) | 1 | 1 | 1 | 1 | 0 |
| for...of over an object | 1 | 1 | 1 | 1 | 0 |
| function identity comparison in a program with specialized generic function values | 1 | 1 | 1 | 1 | 0 |
| hasOwnProperty through an object view (an Error may be hidden by the view) | 1 | 1 | 1 | 1 | 0 |
| in when a possibly absent spread has no property presence descriptors | 2 | 2 | 2 | 2 | 0 |
| inherited library member add read as an own field | 1 | 1 | 1 | 1 | 0 |
| inherited library member call read as an own field | 1 | 1 | 1 | 1 | 0 |
| inherited library member get read as an own field | 4 | 4 | 4 | 4 | 0 |
| inherited library member has read as an own field | 2 | 2 | 2 | 2 | 0 |
| new a ParenthesizedExpression | 3 | 3 | 3 | 3 | 0 |
| new an Identifier | 4 | 4 | 4 | 4 | 0 |
| node:timers.clearTimeout value needs exactly (timeoutId: unknown) => void, not (timeout: unknown) => void with no slot | 2 | 2 | 2 | 2 | 0 |
| optional property equalsToken in ShorthandPropertyAssignment absent from structural source never, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields | 4 | 4 | 4 | 4 | 0 |
| optional property id in ArrayBindingPattern absent from structural source never, which can hide fields | 6 | 6 | 6 | 6 | 0 |
| optional property id in BigIntLiteral absent from structural source never, which can hide fields | 4 | 4 | 4 | 4 | 0 |
| optional property id in BinaryExpression absent from structural source never, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property id in ExternalModuleReference absent from structural source never, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property id in Identifier absent from structural source never, which can hide fields | 3 | 3 | 3 | 3 | 0 |
| optional property id in Node absent from structural source never, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property isInvalidated in CachedResolvedModuleWithFailedLookupLocations absent from structural source ResolvedModuleWithFailedLookupLocations, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property isInvalidated in CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations absent from structural source ResolvedTypeReferenceDirectiveWithFailedLookupLocations, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property isInvalidated in ResolutionWithFailedLookupLocations absent from structural source ResolvedModuleWithFailedLookupLocations, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property isInvalidated in ResolutionWithFailedLookupLocations absent from structural source ResolvedTypeReferenceDirectiveWithFailedLookupLocations, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property jsDocCache in JSDocArray absent from structural source JSDoc[], which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property modifiers in ExportAssignment absent from structural source never, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property modifiers in PropertyDeclaration absent from structural source never, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property name in PackageJson absent from structural source object, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property originalPath in Resolved absent from structural source { path: string; extension: string; packageId: PackageId \| undefined; resolvedUsingTsExtension: boolean \| undefined; }, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property preserve in FileReference absent from structural source { resolutionMode: ModuleKind.CommonJS \| ModuleKind.ESNext; }, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property resolutionMode in FileReference absent from structural source { preserve: true; }, which can hide fields | 3 | 3 | 3 | 3 | 0 |
| optional property source in SourceMapRange absent from structural source TextRange, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property types in { types?: { value: string; pos: number; end: number; }; } & { lib?: { value: string; pos: number; end: number; }; } & { path?: { value: string; pos: number; end: number; }; } & { "no-default-lib"?: string; } & { "resolution-mode"?: string; } & ... absent from structural source { [index: string]: string \| undefined; }, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property types in { types?: { value: string; pos: number; end: number; }; } & { lib?: { value: string; pos: number; end: number; }; } & { path?: { value: string; pos: number; end: number; }; } & { "no-default-lib"?: string; } & { "resolution-mode"?: string; } & ... absent from structural source { [index: string]: string \| { value: string; pos: number; end: number; }; }, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property version in PackageJson absent from structural source PackageJsonPathFields, which can hide fields | 5 | 5 | 5 | 5 | 0 |
| optional record property access | 1 | 1 | 1 | 1 | 0 |
| overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined | 1 | 1 | 1 | 1 | 0 |
| overload 1 of injectClassNamedEvaluationHelperBlockIfMissing result Extract<ClassDeclaration, Pick<...>> \| Extract<...> cannot be served by implementation result ClassLikeDeclaration | 1 | 1 | 1 | 1 | 0 |
| overload 1 of transformNamedEvaluation result Extract<AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; }, Pick<...>> \| ... 7 more ... \| Extract<...> cannot be served by implementation result BinaryExpression \| BindingElement \| ExportAssignment \| ParameterDeclaration \| PropertyAssignment \| PropertyDeclaration \| ShorthandPropertyAssignment \| VariableDeclaration | 1 | 1 | 1 | 1 | 0 |
| sort without a comparator | 1 | 1 | 1 | 1 | 0 |
| writing a checked object field that may hold undefined or null, whose checked write can't change the representation its slot holds | 5 | 5 | 5 | 5 | 0 |
| writing a possibly absent optional own field | 9 | 9 | 9 | 9 | 0 |

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
