| Reason mentioning view, branded supertype, or predicate | A | B |
|---|---:|---:|
| Array.isArray on a tuple or an erased object/any/unknown view | 1 | 1 |
| a branded supertype inside a checked predicate's array target | 2 | 2 |
| a checked predicate target DynamicNamedDeclaration that narrows GetAccessorDeclaration, which isn't its supertype | 1 | 1 |
| a checked predicate target LeftHandSideExpression over a generic argument TIn & ({} \| null), read on the true side as Expression & LeftHandSideExpression | 1 | 1 |
| a checked predicate target NamedDeclaration that narrows ArrowFunction, which isn't its supertype | 1 | 1 |
| a checked predicate target TypeNode that narrows undefined, which isn't its supertype | 1 | 1 |
| a checked predicate target whose base has no required field to check (CompilerOptions & { types: string[]; }) | 1 | 1 |
| a checked predicate target whose check would read past the depth bound (BuilderProgramStateWithDefinedProgram) | 1 | 1 |
| a checked predicate target whose check would read past the depth bound (CircularBuildOrder) | 1 | 1 |
| a checked predicate target whose check would read past the depth bound (CommandLineOptionOfCustomType) | 1 | 1 |
| a checked predicate target whose check would read past the depth bound (IncrementalBundleEmitBuildInfo) | 1 | 1 |
| a checked view with a callable contract | 1 | 1 |
| a literal method through a view that erases its receiver | 1 | 1 |
| a type predicate whose return is not proven (there is no body proving this parameter) | 14 | 14 |
| an object with a nullable field viewed as unknown or object (null and undefined slot tags) | 1 | 1 |
| an unproven predicate argument for parameter callback (argument "decl => isVariableDeclarationInitializedToRequire(decl)") | 1 | 1 |
| an unproven predicate argument for parameter callback (argument "n => isExportAssignment(n) \|\| isStatement(n) \|\| isVariableDeclaration(n) \|\| isPropertyDeclaration(n) \|\| isParameter(n)") | 1 | 1 |
| an unproven predicate argument for parameter f (argument "helper => !helper.scoped") | 1 | 1 |
| an unproven predicate argument for parameter f (argument "p => isParameterPropertyDeclaration(p, constructor)") | 1 | 1 |
| an unproven predicate argument for parameter nodeTest (argument "isModuleDeclaration") | 1 | 1 |
| an unproven predicate argument for parameter predicate (argument "modifier => !isExportModifier(modifier) && !isDefaultModifier(modifier)") | 2 | 2 |
| an unproven predicate argument for parameter predicate (argument "predicate") | 2 | 2 |
| an unproven predicate argument for parameter test (argument "createAssignmentCallback ? isIdentifier : isExpression") | 1 | 1 |
| checked predicate membership over an open tag domain | 47 | 50 |
| checked predicate overload target AssignmentExpression<AmpersandAmpersandEqualsToken \| BarBarEqualsToken \| EqualsToken \| QuestionQuestionEqualsToken> & ... | 0 | 1 |
| checked predicate overload target BinaryExpression & { operatorToken: Token<SyntaxKind.CommaToken>; } | 1 | 1 |
| checked predicate overload target BindableStaticElementAccessExpression | 0 | 2 |
| checked predicate overload target ClassNamedEvaluationHelperBlock | 1 | 1 |
| checked predicate overload target ClassThisAssignmentBlock | 1 | 1 |
| checked predicate overload target ElementAccessExpression & Declaration & { readonly argumentExpression: NumericLiteral \| StringLiteralLike; } & ... | 1 | 1 |
| checked predicate overload target ImportClause & { readonly isTypeOnly: true; readonly name: Identifier; } | 1 | 1 |
| checked predicate overload target ImportEqualsDeclaration & { moduleReference: ExternalModuleReference; } | 1 | 1 |
| checked predicate overload target JSDocTag | 0 | 1 |
| checked predicate overload target LiteralImportTypeNode | 1 | 1 |
| checked predicate overload target LiteralLikeElementAccessExpression | 1 | 1 |
| checked predicate overload target NamedDeclaration & { name: DeclarationName; } | 1 | 1 |
| checked predicate overload target NonIncrementalBuildInfo | 1 | 1 |
| checked predicate overload target ParameterPropertyDeclaration | 1 | 1 |
| checked predicate overload target PrefixUnaryExpression & { operand: NumericLiteral; } | 0 | 1 |
| checked predicate overload target PrefixUnaryExpression & { operator: SyntaxKind.PlusToken; operand: NumericLiteral; } | 0 | 1 |
| checked predicate overload target ReferencedFile | 1 | 1 |
| checked view field modifiers of type NodeArray<Modifier> \| NodeArray<ModifierLike> \| undefined | 1 | 1 |
| checked view union arm awaits views-v3: array element kind | 0 | 1 |
| detached hasOwnProperty through an object view (a value with a different native representation may be hidden by the view) | 1 | 1 |
