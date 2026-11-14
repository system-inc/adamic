# Paired latent census

| Measure | Raw findings | Distinct reads |
| --- | ---: | ---: |
| Lowered (no remaining finding) | 0 | 0 |
| Moved by reason | 0 | 0 |
| New findings | 0 | 0 |
| New sites | 0 | 0 |
| Units whose first stop moved | 0 | 0 |
| Regressions | 0 | 0 |
| before NotYet | 83 | 83 |
| before Refused | 815 | 784 |
| after NotYet | 83 | 83 |
| after Refused | 815 | 784 |

| Reason | Raw before | Raw after | Reads before | Reads after | Reads lowered |
| --- | ---: | ---: | ---: | ---: | ---: |
| JSON.parse: its result's type can't be proven from the text | 1 | 1 | 1 | 1 | 0 |
| a () => never seen as a (node: ElementAccessExpression \| EnumMember \| PropertyAccessExpression) => string \| number \| undefined (one keeps something weakly that the other keeps strongly) | 1 | 1 | 1 | 1 | 0 |
| a (_?: unknown) => void seen as a (flags: LexicalEnvironmentFlags, value: boolean) => void (one keeps something weakly that the other keeps strongly) | 1 | 1 | 1 | 1 | 0 |
| a CompilerOptions seen as SourceMapOptions (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a Map or Set interface implemented by a value that isn't a Map or Set | 26 | 26 | 26 | 26 | 0 |
| a PrinterOptions seen as Pick<CompilerOptions, "module" \| "target"> (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a boolean \| undefined variable a function value captures | 1 | 1 | 1 | 1 | 0 |
| a call through ?. (an optional call) | 3 | 3 | 3 | 3 | 0 |
| a case whose type differs from the switch's | 3 | 3 | 3 | 3 | 0 |
| a cast the runtime can't check | 134 | 134 | 134 | 134 | 0 |
| a checked predicate target NamedDeclaration that narrows ArrowFunction, which isn't its supertype | 1 | 1 | 1 | 1 | 0 |
| a conditional whose branches have different types | 1 | 1 | 1 | 1 | 0 |
| a for...of destructuring an object | 1 | 1 | 1 | 1 | 0 |
| a function returning CapturedThis | 1 | 1 | 1 | 1 | 0 |
| a function returning T \| undefined | 1 | 1 | 1 | 1 | 0 |
| a function taking CancellationToken seen as one taking CancellationToken \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking NodeArray<T> seen as one taking NodeArray<T> \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking TypeNode seen as one taking Node (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function taking readonly string[] seen as one taking readonly string[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 0 |
| a function value that captures the variable its own initializer declares | 1 | 1 | 1 | 1 | 0 |
| a literal method through a view that erases its receiver | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (afterProgramCreate would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (createDirectory would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (directoryExists would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (emitNodeWithNotification would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (fileExists would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (getCanonicalFileName would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getCommonSourceDirectory would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getCompilerOptions would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getFileIncludeReasons would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getSourceFile would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (getSourceFileByPath would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getSourceFiles would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (hasGlobalName would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (hasInvalidatedLibResolutions would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (hasInvalidatedResolutions would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (isEmitNotificationEnabled would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (liftToBlock would lose its object, and this with it) | 11 | 11 | 11 | 11 | 0 |
| a method read as a value (onReleaseOldSourceFile would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (onReleaseParsedCommandLine would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (parenthesizeBranchOfConditionalExpression would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method read as a value (parenthesizeCheckTypeOfConditionalType would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (parenthesizeConciseBodyOfArrowFunction would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (parenthesizeConditionOfConditionalExpression would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (parenthesizeConstituentTypeOfIntersectionType would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (parenthesizeConstituentTypeOfUnionType would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
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
| a method read as a value (readFile would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (substituteNode would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports | 1 | 1 | 1 | 1 | 0 |
| a read narrowed to undefined of a property a call may write | 1 | 1 | 1 | 1 | 0 |
| a read of a void value, as a returned value | 1 | 1 | 1 | 1 | 0 |
| a rest parameter that isn't an array | 1 | 1 | 1 | 1 | 0 |
| a structural relation deeper than 64 type pairs while checking void function results | 1 | 1 | 1 | 1 | 0 |
| a tuple literal with more values than its tuple has elements | 1 | 1 | 1 | 1 | 0 |
| a value of type (node: CommentRange) => boolean seen as (value: SynthesizedComment) => boolean, which can write number where -1 is read | 1 | 1 | 1 | 1 | 0 |
| a value of type (resolution: ResolvedModuleWithFailedLookupLocations) => ResolvedModuleFull \| undefined seen as (oldResolution: ResolvedModuleWithFailedLookupLocations) => ResolutionWithResolvedFileName \| undefined, which can write string \| undefined where string is read | 1 | 1 | 1 | 1 | 0 |
| a value of type ArrayLiteralExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type AssignmentExpression<EqualsToken> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 11 | 11 | 11 | 11 | 0 |
| a value of type BinaryExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type BindableObjectDefinePropertyCall | 1 | 1 | 1 | 1 | 0 |
| a value of type BindingName seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type BindingName seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Block seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type Block seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 10 | 10 | 10 | 10 | 0 |
| a value of type BreakStatement seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type CallExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 7 | 7 | 7 | 7 | 0 |
| a value of type CapturedThis seen as Expression \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type CapturedThis seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type CapturedThis seen as PrimaryExpression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type CapturedThis seen as string \| BindingName, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Children seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ClassDeclaration seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ClassDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type ClassExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type CompilerHost seen as ModuleResolutionHost, whose method useCaseSensitiveFileNames becomes a writable field: a write of boolean \| (() => boolean) \| undefined would replace a method of type () => boolean | 12 | 12 | 12 | 12 | 0 |
| a value of type ContinueStatement seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Diagnostic seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 | 1 | 0 |
| a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 17 | 17 | 17 | 17 | 0 |
| a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where SourceFile is read | 4 | 4 | 4 | 4 | 0 |
| a value of type DiagnosticWithLocation[] seen as Diagnostic[], which can write Diagnostic where DiagnosticWithLocation is read | 2 | 2 | 2 | 2 | 0 |
| a value of type DiagnosticWithLocation[] seen as readonly Diagnostic[] \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 | 1 | 0 |
| a value of type DiagnosticWithLocation[] seen as readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 | 1 | 0 |
| a value of type DiagnosticWithLocation[] \| undefined seen as DiagnosticRelatedInformation[] \| undefined, which can write DiagnosticRelatedInformation where DiagnosticWithLocation is read | 2 | 2 | 2 | 2 | 0 |
| a value of type DiagnosticWithLocation[] \| undefined seen as readonly Diagnostic[] \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 | 1 | 0 |
| a value of type ElementAccessExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type EmitNode & { autoGenerate: AutoGenerateInfo; } | 1 | 1 | 1 | 1 | 0 |
| a value of type EmptyStatement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type Expression seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type Expression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 9 | 9 | 9 | 9 | 0 |
| a value of type Expression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 12 | 12 | 12 | 12 | 0 |
| a value of type Expression \| GeneratedIdentifier seen as Expression \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ExpressionStatement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 17 | 17 | 17 | 17 | 0 |
| a value of type FlowArrayMutation \| FlowAssignment seen as FlowNode, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 5 | 5 | 5 | 5 | 0 |
| a value of type FlowCall seen as FlowNode, which can write BinaryExpression \| CallExpression where CallExpression is read | 2 | 2 | 2 | 2 | 0 |
| a value of type FlowNode seen as FlowNode \| undefined, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 14 | 14 | 14 | 14 | 0 |
| a value of type FlowNode seen as FlowNode[] \| FlowNode \| undefined, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 5 | 5 | 5 | 5 | 0 |
| a value of type ForStatement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 4 | 4 | 4 | 4 | 0 |
| a value of type FunctionDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type FunctionDeclaration[] \| VariableDeclaration[] \| undefined seen as Statement[] \| VariableDeclaration[] \| undefined, which can write Statement where FunctionDeclaration is read | 1 | 1 | 1 | 1 | 0 |
| a value of type FunctionExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 4 | 4 | 4 | 4 | 0 |
| a value of type GeneratedIdentifier seen as Expression \| GeneratedIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GeneratedIdentifier seen as GeneratedIdentifier \| Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GeneratedIdentifier seen as Identifier \| PrivateIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type GeneratedIdentifier seen as Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as Identifier \| PrivateIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as Node \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier \| Node seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GeneratedIdentifier \| Identifier seen as GeneratedIdentifier \| Identifier \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GeneratedIdentifier \| Identifier seen as Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 5 | 5 | 5 | 5 | 0 |
| a value of type GeneratedIdentifier \| Identifier seen as string \| BindingName, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GeneratedIdentifier \| Identifier seen as string \| GeneratedIdentifier \| Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GeneratedIdentifier \| Identifier \| undefined seen as string \| ModuleExportName \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Identifier & GeneratedIdentifier & { readonly escapedText: { __escapedIdentifier: void; } & "__this"; } seen as BindingName, which can write EmitNode \| undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Identifier seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Identifier seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Identifier seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 10 | 10 | 10 | 10 | 0 |
| a value of type Identifier[] seen as ModuleExportName[] \| undefined, which can write ModuleExportName where Identifier is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Identifier[] \| undefined seen as ModuleExportName[] \| undefined, which can write ModuleExportName where Identifier is read | 1 | 1 | 1 | 1 | 0 |
| a value of type IfStatement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type ImportDeclaration seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type InitializedVariableDeclaration | 2 | 2 | 2 | 2 | 0 |
| a value of type InitializedVariableDeclaration seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type LeftHandSideExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ModuleDeclaration seen as Mutable<ModuleDeclaration \| SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ModuleName seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type NewExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Node seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type Node seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Node \| undefined seen as EmitNode \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of EmitFlags would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type NodeArray<Expression> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type NodeArray<ExpressionStatement> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type NodeArray<ParameterDeclaration> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type NodeArray<Statement> seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type NodeArray<Statement> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 16 | 16 | 16 | 16 | 0 |
| a value of type NumericLiteral seen as Mutable<LiteralExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type NumericLiteral seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ObjectLiteralExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ParameterDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type ParenthesizedExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type PostfixUnaryExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type PropertyAccessExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 8 | 8 | 8 | 8 | 0 |
| a value of type PropertyAccessExpression \| SyntheticSuper seen as LeftHandSideExpression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type PropertyAccessExpression \| SyntheticSuper seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type PropertyAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 5 | 5 | 5 | 5 | 0 |
| a value of type PropertyName seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type ReferencedFile & { kind: FileIncludeKind.LibReferenceDirective; } seen as ReferencedFile, which can write ReferencedFileKind where FileIncludeKind.LibReferenceDirective is read | 1 | 1 | 1 | 1 | 0 |
| a value of type ReportFileInError[] seen as (ReportFileInError \| undefined)[], which can write ReportFileInError \| undefined where ReportFileInError is read | 1 | 1 | 1 | 1 | 0 |
| a value of type ReturnStatement seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ReturnStatement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 12 | 12 | 12 | 12 | 0 |
| a value of type ShorthandPropertyAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type SourceFile seen as Mutable<ModuleDeclaration \| SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type SourceFile \| undefined seen as FileReasonToChainCache \| undefined, which can write DiagnosticMessageChain[] \| undefined where RedirectInfo \| undefined is read | 1 | 1 | 1 | 1 | 0 |
| a value of type SpreadAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type SpreadElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Statement seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type Statement[] \| undefined seen as CaseClause[] \| undefined, which can write CaseClause where Statement is read | 1 | 1 | 1 | 1 | 0 |
| a value of type StringLiteral seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type StringLiteral seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type SuperExpression seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type SuperExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Symbol \| undefined seen as Declaration \| undefined, which can write number \| undefined where number is read | 2 | 2 | 2 | 2 | 0 |
| a value of type SyntheticSuper seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 4 | 4 | 4 | 4 | 0 |
| a value of type SyntheticSuper seen as string \| BindingName, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type T seen as Mutable<T>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of T["pos"] would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ThisCapturingVariableDeclaration seen as VariableDeclaration, which can write EmitNode \| undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 1 | 1 | 1 | 1 | 0 |
| a value of type ThrowStatement seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ThrowStatement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type VariableDeclaration seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 8 | 8 | 8 | 8 | 0 |
| a value of type VariableDeclarationList seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 11 | 11 | 11 | 11 | 0 |
| a value of type VariableStatement seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type VariableStatement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 9 | 9 | 9 | 9 | 0 |
| a value of type YieldExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type YieldExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type any seen as T, a type parameter whose constraint Node can be written, so it can write what any can't hold | 1 | 1 | 1 | 1 | 0 |
| a value of type any: any has no proven type | 2 | 2 | 2 | 2 | 0 |
| a value of type never | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as DiagnosticArguments, which can write string \| number \| boolean \| readonly string[] \| SourceFile \| undefined where never is read | 3 | 3 | 3 | 3 | 0 |
| a value of type never[] seen as DiagnosticWithLocation[], which can write DiagnosticWithLocation where never is read | 2 | 2 | 2 | 2 | 0 |
| a value of type never[] seen as Diagnostic[], which can write Diagnostic where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as Identifier[], which can write Identifier where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as ResolvedProjectReference[], which can write ResolvedProjectReference where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as TransformerFactory<Bundle \| SourceFile>[], which can write TransformerFactory<Bundle \| SourceFile> where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[] \| undefined, which can write SourceFile \| undefined where SourceFile is read | 2 | 2 | 2 | 2 | 0 |
| a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 | 1 | 0 |
| a value of type string \| GeneratedIdentifier \| Identifier seen as string \| ModuleExportName, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type string[] seen as DiagnosticArguments, which can write string \| number \| boolean \| readonly string[] \| SourceFile \| undefined where string is read | 1 | 1 | 1 | 1 | 0 |
| a value of type string[] seen as string[] \| undefined, which can write string where never is read | 1 | 1 | 1 | 1 | 0 |
| a void value at the argument site through () seen as number | 1 | 1 | 1 | 1 | 0 |
| a void-returning function at the assignment site seen as (callback: (label?: unknown) => void, ms: number, label?: unknown) => unknown | 1 | 1 | 1 | 1 | 0 |
| a write through a union of ArrowFunction \| FunctionExpression \| GetAccessorDeclaration \| MethodDeclaration \| SetAccessorDeclaration into a slot that holds BinaryExpression \| CallExpression in FlowArrayMutation | 1 | 1 | 1 | 1 | 0 |
| a write through a union to a slot SyntaxKindFeatureFlags[] doesn't name | 2 | 2 | 2 | 2 | 0 |
| an assignment used as a value whose target write type IsBlockScopedContainer cannot be read as EntityNameExpression \| IsContainer | 1 | 1 | 1 | 1 | 0 |
| an assignment used as a value whose target write type Resolution cannot be read as void | 1 | 1 | 1 | 1 | 0 |
| an assignment used as a value whose target write type Resolution \| undefined cannot be read as Resolution | 1 | 1 | 1 | 1 | 0 |
| an assignment value to a member | 1 | 1 | 1 | 1 | 0 |
| an object refinement using an open numeric enum as a literal tag | 170 | 170 | 139 | 139 | 0 |
| an optional Map get through a structural receiver | 2 | 2 | 2 | 2 | 0 |
| an optional chain longer than one step | 2 | 2 | 2 | 2 | 0 |
| an unproven predicate argument for parameter predicate (argument "modifier => !isExportModifier(modifier) && !isDefaultModifier(modifier)") | 2 | 2 | 2 | 2 | 0 |
| an unproven predicate argument for parameter predicate (argument "predicate") | 2 | 2 | 2 | 2 | 0 |
| an unproven relation from Identifier to Identifier: optional field id has no proven compatible presence/type | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from NodeArray<ModifierLike> & readonly Decorator[] to NodeArray<Decorator>: optional field concat.parameter.element.slice.element.parent.name has no proven compatible presence/type | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from string to T: the source is not assignable to the target | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ConditionalType | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.DotDotDotToken | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Identifier | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.JSDoc | 2 | 2 | 2 | 2 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.JSDocTypeExpression | 4 | 4 | 4 | 4 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.VariableDeclaration | 1 | 1 | 1 | 1 | 0 |
| assigning a non-tuple array's length: array writes must preserve its existing indices | 1 | 1 | 1 | 1 | 0 |
| checked predicate membership over an open tag domain | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target ElementAccessExpression & Declaration & { readonly argumentExpression: NumericLiteral \| StringLiteralLike; } & ... | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target ExportDeclaration & { readonly isTypeOnly: true; readonly moduleSpecifier: Expression; } | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target ImportClause & { readonly isTypeOnly: true; readonly name: Identifier; } | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target JSDocTag | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target NamedDeclaration & { name: DeclarationName; } | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target ReferencedFile | 1 | 1 | 1 | 1 | 0 |
| destructuring a value made in the assignment itself | 1 | 1 | 1 | 1 | 0 |
| new an Identifier | 1 | 1 | 1 | 1 | 0 |
| optional property _propertyAccessExpressionLikeQualifiedNameBrand in PropertyAccessEntityNameExpression absent from structural source never, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property createProgram in IncrementalProgramOptions<EmitAndSemanticDiagnosticsBuilderProgram> absent from structural source IncrementalCompilationOptions, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property equalsToken in ShorthandPropertyAssignment absent from structural source never, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property failedLookupLocations in ResolvedModuleWithFailedLookupLocations absent from structural source { resolvedModule: ResolvedModuleFull; }, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property failedLookupLocations in ResolvedTypeReferenceDirectiveWithFailedLookupLocations absent from structural source { resolvedTypeReferenceDirective: ResolvedTypeReferenceDirective \| undefined; }, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property id in ArrayBindingPattern absent from structural source never, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property id in Expression absent from structural source never, which can hide fields | 3 | 3 | 3 | 3 | 0 |
| optional property id in ExternalModuleReference absent from structural source never, which can hide fields | 3 | 3 | 3 | 3 | 0 |
| optional property id in Node absent from structural source never, which can hide fields | 3 | 3 | 3 | 3 | 0 |
| optional property onUnRecoverableConfigFileDiagnostic in CompilerHost & { onUnRecoverableConfigFileDiagnostic?: DiagnosticReporter; } absent from structural source CompilerHost, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property skipTrivia in SourceMapSource absent from structural source SourceFile, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| optional property source in SourceMapRange absent from structural source TextRange, which can hide fields | 3 | 3 | 3 | 3 | 0 |
| overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined | 1 | 1 | 1 | 1 | 0 |
| overload 1 of visitNode parameter node cannot be served by implementation parameter node | 1 | 1 | 1 | 1 | 0 |
| overload 2 of getParseTreeNode result T \| undefined cannot be served by implementation result Node \| undefined | 1 | 1 | 1 | 1 | 0 |
| overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor | 1 | 1 | 1 | 1 | 0 |
| toLocaleTimeString: Date methods other than getTime/valueOf in the fs file host | 1 | 1 | 1 | 1 | 0 |
| writing a checked object field that may hold undefined or null, whose checked write can't change the representation its slot holds | 8 | 8 | 8 | 8 | 0 |
| writing a checked object field without its source-slot type certificate | 1 | 1 | 1 | 1 | 0 |
| writing a possibly absent optional own field | 1 | 1 | 1 | 1 | 0 |

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
