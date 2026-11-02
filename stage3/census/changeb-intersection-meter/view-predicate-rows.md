# All view, branded-supertype and predicate reason rows

Includes unchanged rows. Counts use the meter distinct-site key.

| Kind and reason | A | B | A example locations | B example locations |
|---|---:|---:|---|---|
| NotYet: Array.isArray on a tuple or an erased object/any/unknown view | 1 | 1 | src/compiler/utilities.ts:8765:54 | src/compiler/utilities.ts:8765:54 |
| NotYet: a branded supertype inside a checked predicate's array target | 2 | 2 | src/compiler/transformers/destructuring.ts:171:14; src/compiler/visitorPublic.ts:548:9 | src/compiler/transformers/destructuring.ts:171:14; src/compiler/visitorPublic.ts:548:9 |
| NotYet: a checked predicate target DynamicNamedDeclaration that narrows GetAccessorDeclaration, which isn't its supertype | 1 | 1 | src/compiler/utilities.ts:6838:9 | src/compiler/utilities.ts:6838:9 |
| NotYet: a checked predicate target LeftHandSideExpression over a generic argument TIn & ({} &#124; null), read on the true side as Expression & LeftHandSideExpression | 1 | 0 | src/compiler/core.ts:1784:32 |  |
| NotYet: a checked predicate target NamedDeclaration that narrows ArrowFunction, which isn't its supertype | 1 | 1 | src/compiler/utilitiesPublic.ts:888:9 | src/compiler/utilitiesPublic.ts:888:9 |
| NotYet: a checked predicate target TypeNode that narrows undefined, which isn't its supertype | 1 | 1 | src/compiler/factory/utilities.ts:626:5 | src/compiler/factory/utilities.ts:626:5 |
| NotYet: a checked predicate target whose base has no required field to check (CompilerOptions & { types: string[]; }) | 1 | 1 | src/compiler/utilities.ts:9028:1 | src/compiler/utilities.ts:9028:1 |
| NotYet: a checked predicate target whose check would read past the depth bound (BuilderProgramStateWithDefinedProgram) | 1 | 1 | src/compiler/builder.ts:268:1 | src/compiler/builder.ts:268:1 |
| NotYet: a checked predicate target whose check would read past the depth bound (CircularBuildOrder) | 1 | 1 | src/compiler/tsbuildPublic.ts:259:1 | src/compiler/tsbuildPublic.ts:259:1 |
| NotYet: a checked predicate target whose check would read past the depth bound (CommandLineOptionOfCustomType) | 1 | 1 | src/compiler/commandLineParser.ts:1728:1 | src/compiler/commandLineParser.ts:1728:1 |
| NotYet: a checked predicate target whose check would read past the depth bound (IncrementalBundleEmitBuildInfo) | 1 | 1 | src/compiler/builder.ts:1181:1 | src/compiler/builder.ts:1181:1 |
| Refused: a checked view with a callable contract | 1 | 1 | src/compiler/sys.ts:1659:13 | src/compiler/sys.ts:1659:13 |
| NotYet: a literal method through a view that erases its receiver | 1 | 1 | src/compiler/program.ts:528:37 | src/compiler/program.ts:528:37 |
| Refused: a type predicate whose return is not proven (there is no body proving this parameter) | 14 | 14 | src/compiler/core.ts:2459:66; src/compiler/core.ts:2459:91; src/compiler/core.ts:2461:105; src/compiler/core.ts:2461:130; src/compiler/core.ts:2461:80 | src/compiler/core.ts:2459:66; src/compiler/core.ts:2459:91; src/compiler/core.ts:2461:105; src/compiler/core.ts:2461:130; src/compiler/core.ts:2461:80 |
| NotYet: an object with a nullable field viewed as unknown or object (null and undefined slot tags) | 1 | 1 | src/compiler/builder.ts:295:59 | src/compiler/builder.ts:295:59 |
| Refused: an unproven predicate argument for parameter callback (argument "decl => isVariableDeclarationInitializedToRequire(decl)") | 1 | 1 | src/compiler/utilities.ts:3908:53 | src/compiler/utilities.ts:3908:53 |
| Refused: an unproven predicate argument for parameter callback (argument "n => isExportAssignment(n) &#124;&#124; isStatement(n) &#124;&#124; isVariableDeclaration(n) &#124;&#124; isPropertyDeclaration(n) &#124;&#124; isParameter(n)") | 1 | 1 | src/compiler/transformers/declarations/diagnostics.ts:708:43 | src/compiler/transformers/declarations/diagnostics.ts:708:43 |
| Refused: an unproven predicate argument for parameter f (argument "helper => !helper.scoped") | 1 | 1 | src/compiler/factory/utilities.ts:761:47 | src/compiler/factory/utilities.ts:761:47 |
| Refused: an unproven predicate argument for parameter f (argument "p => isParameterPropertyDeclaration(p, constructor)") | 1 | 1 | src/compiler/transformers/ts.ts:1384:44 | src/compiler/transformers/ts.ts:1384:44 |
| Refused: an unproven predicate argument for parameter nodeTest (argument "isModuleDeclaration") | 1 | 1 | src/compiler/transformers/ts.ts:1975:47 | src/compiler/transformers/ts.ts:1975:47 |
| Refused: an unproven predicate argument for parameter predicate (argument "modifier => !isExportModifier(modifier) && !isDefaultModifier(modifier)") | 2 | 2 | src/compiler/factory/nodeConverters.ts:66:40; src/compiler/factory/nodeConverters.ts:84:36 | src/compiler/factory/nodeConverters.ts:66:40; src/compiler/factory/nodeConverters.ts:84:36 |
| Refused: an unproven predicate argument for parameter predicate (argument "predicate") | 2 | 2 | src/compiler/utilitiesPublic.ts:1280:52; src/compiler/utilitiesPublic.ts:1285:38 | src/compiler/utilitiesPublic.ts:1280:52; src/compiler/utilitiesPublic.ts:1285:38 |
| Refused: an unproven predicate argument for parameter test (argument "createAssignmentCallback ? isIdentifier : isExpression") | 1 | 1 | src/compiler/transformers/destructuring.ts:185:34 | src/compiler/transformers/destructuring.ts:185:34 |
| NotYet: checked predicate membership over an open tag domain | 50 | 50 | src/compiler/factory/nodeConverters.ts:55:13; src/compiler/factory/nodeFactory.ts:7287:20; src/compiler/factory/utilities.ts:1024:21; src/compiler/factory/utilities.ts:1096:17; src/compiler/factory/utilities.ts:1573:9 | src/compiler/factory/nodeConverters.ts:55:13; src/compiler/factory/nodeFactory.ts:7287:20; src/compiler/factory/utilities.ts:1024:21; src/compiler/factory/utilities.ts:1096:17; src/compiler/factory/utilities.ts:1573:9 |
| NotYet: checked predicate overload target AssignmentExpression<AmpersandAmpersandEqualsToken &#124; BarBarEqualsToken &#124; EqualsToken &#124; QuestionQuestionEqualsToken> & ... | 1 | 1 | src/compiler/utilities.ts:5522:1 | src/compiler/utilities.ts:5522:1 |
| NotYet: checked predicate overload target BinaryExpression & { operatorToken: Token<SyntaxKind.CommaToken>; } | 1 | 1 | src/compiler/factory/utilities.ts:607:1 | src/compiler/factory/utilities.ts:607:1 |
| NotYet: checked predicate overload target BindableStaticElementAccessExpression | 2 | 2 | src/compiler/utilities.ts:4148:1; src/compiler/utilities.ts:7563:1 | src/compiler/utilities.ts:4148:1; src/compiler/utilities.ts:7563:1 |
| NotYet: checked predicate overload target ClassNamedEvaluationHelperBlock | 1 | 1 | src/compiler/transformers/namedEvaluation.ts:139:1 | src/compiler/transformers/namedEvaluation.ts:139:1 |
| NotYet: checked predicate overload target ClassThisAssignmentBlock | 1 | 1 | src/compiler/transformers/classThis.ts:72:1 | src/compiler/transformers/classThis.ts:72:1 |
| NotYet: checked predicate overload target ElementAccessExpression & Declaration & { readonly argumentExpression: NumericLiteral &#124; StringLiteralLike; } & ... | 1 | 1 | src/compiler/utilities.ts:4104:1 | src/compiler/utilities.ts:4104:1 |
| NotYet: checked predicate overload target ImportClause & { readonly isTypeOnly: true; readonly name: Identifier; } | 1 | 1 | src/compiler/utilitiesPublic.ts:1536:1 | src/compiler/utilitiesPublic.ts:1536:1 |
| NotYet: checked predicate overload target ImportEqualsDeclaration & { moduleReference: ExternalModuleReference; } | 1 | 1 | src/compiler/utilities.ts:3773:1 | src/compiler/utilities.ts:3773:1 |
| NotYet: checked predicate overload target JSDocTag | 1 | 1 | src/compiler/utilitiesPublic.ts:2534:1 | src/compiler/utilitiesPublic.ts:2534:1 |
| NotYet: checked predicate overload target LiteralImportTypeNode | 1 | 1 | src/compiler/utilities.ts:2724:1 | src/compiler/utilities.ts:2724:1 |
| NotYet: checked predicate overload target LiteralLikeElementAccessExpression | 1 | 1 | src/compiler/utilities.ts:4139:1 | src/compiler/utilities.ts:4139:1 |
| NotYet: checked predicate overload target NamedDeclaration & { name: DeclarationName; } | 1 | 1 | src/compiler/utilitiesPublic.ts:946:1 | src/compiler/utilitiesPublic.ts:946:1 |
| NotYet: checked predicate overload target NonIncrementalBuildInfo | 1 | 1 | src/compiler/builder.ts:1192:1 | src/compiler/builder.ts:1192:1 |
| NotYet: checked predicate overload target ParameterPropertyDeclaration | 1 | 1 | src/compiler/utilitiesPublic.ts:615:1 | src/compiler/utilitiesPublic.ts:615:1 |
| NotYet: checked predicate overload target PrefixUnaryExpression & { operand: NumericLiteral; } | 1 | 1 | src/compiler/utilities.ts:5367:1 | src/compiler/utilities.ts:5367:1 |
| NotYet: checked predicate overload target PrefixUnaryExpression & { operator: SyntaxKind.PlusToken; operand: NumericLiteral; } | 1 | 1 | src/compiler/debug.ts:365:5 | src/compiler/debug.ts:365:5 |
| NotYet: checked predicate overload target ReferencedFile | 1 | 1 | src/compiler/program.ts:1162:1 | src/compiler/program.ts:1162:1 |
| NotYet: checked view field modifiers of type NodeArray<Modifier> &#124; NodeArray<ModifierLike> &#124; undefined | 1 | 1 | src/compiler/utilities.ts:5340:11 | src/compiler/utilities.ts:5340:11 |
| NotYet: checked view union arm awaits views-v3: array element kind | 1 | 1 | src/compiler/factory/nodeChildren.ts:20:16 | src/compiler/factory/nodeChildren.ts:20:16 |
| NotYet: detached hasOwnProperty through an object view (a value with a different native representation may be hidden by the view) | 1 | 1 | src/compiler/core.ts:1267:32 | src/compiler/core.ts:1267:32 |
