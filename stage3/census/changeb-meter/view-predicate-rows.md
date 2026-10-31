# Named view and predicate rows

All these rows are unchanged. No reason contains “branded supertype”; no changed row mentions a view read or write. Counts use the meter’s distinct (kind, where, reason, text) key.

| Kind and exact reason | A | B | Example locations (up to five) |
|---|---:|---:|---|
| NotYet: Array.isArray on a tuple or an erased object/any/unknown view | 1 | 1 | src/compiler/utilities.ts:8765:54 |
| NotYet: a checked predicate target whose base has no required field to check (CompilerOptions & { types: string[]; }) | 1 | 1 | src/compiler/utilities.ts:9028:1 |
| NotYet: a checked predicate target whose check would read past the depth bound (BuilderProgramStateWithDefinedProgram) | 1 | 1 | src/compiler/builder.ts:268:1 |
| NotYet: a checked predicate target whose check would read past the depth bound (CircularBuildOrder) | 1 | 1 | src/compiler/tsbuildPublic.ts:259:1 |
| NotYet: a checked predicate target whose check would read past the depth bound (CommandLineOptionOfCustomType) | 1 | 1 | src/compiler/commandLineParser.ts:1728:1 |
| NotYet: a checked predicate target whose check would read past the depth bound (IncrementalBundleEmitBuildInfo) | 1 | 1 | src/compiler/builder.ts:1181:1 |
| NotYet: a literal method through a view that erases its receiver | 1 | 1 | src/compiler/program.ts:528:37 |
| NotYet: an object with a nullable field viewed as unknown or object (null and undefined slot tags) | 1 | 1 | src/compiler/builder.ts:295:59 |
| NotYet: checked predicate membership over an open tag domain | 45 | 45 | src/compiler/factory/nodeConverters.ts:55:13<br>src/compiler/factory/nodeFactory.ts:7287:20<br>src/compiler/factory/utilities.ts:1024:21<br>src/compiler/factory/utilities.ts:1096:17<br>src/compiler/factory/utilities.ts:1746:27 |
| NotYet: checked predicate overload target BinaryExpression & { operatorToken: Token<SyntaxKind.CommaToken>; } | 1 | 1 | src/compiler/factory/utilities.ts:607:1 |
| NotYet: checked predicate overload target ClassElement | 1 | 1 | src/compiler/utilitiesPublic.ts:1732:1 |
| NotYet: checked predicate overload target ElementAccessExpression & Declaration & { readonly argumentExpression: NumericLiteral &#124; StringLiteralLike; } & ... | 1 | 1 | src/compiler/utilities.ts:4104:1 |
| NotYet: checked predicate overload target ImportClause & { readonly isTypeOnly: true; readonly name: Identifier; } | 1 | 1 | src/compiler/utilitiesPublic.ts:1536:1 |
| NotYet: checked predicate overload target ImportEqualsDeclaration & { moduleReference: ExternalModuleReference; } | 1 | 1 | src/compiler/utilities.ts:3773:1 |
| NotYet: checked predicate overload target JSDocPropertyLikeTag | 1 | 1 | src/compiler/utilitiesPublic.ts:1449:1 |
| NotYet: checked predicate overload target LateBoundDeclaration | 1 | 1 | src/compiler/emitter.ts:1158:18 |
| NotYet: checked predicate overload target LiteralExpression | 1 | 1 | src/compiler/utilitiesPublic.ts:1498:1 |
| NotYet: checked predicate overload target LiteralImportTypeNode | 1 | 1 | src/compiler/utilities.ts:2724:1 |
| NotYet: checked predicate overload target LiteralLikeElementAccessExpression | 1 | 1 | src/compiler/utilities.ts:4139:1 |
| NotYet: checked predicate overload target NamedDeclaration & { name: DeclarationName; } | 1 | 1 | src/compiler/utilitiesPublic.ts:946:1 |
| NotYet: checked predicate overload target NonIncrementalBuildInfo | 1 | 1 | src/compiler/builder.ts:1192:1 |
| NotYet: checked predicate overload target ParameterPropertyDeclaration | 1 | 1 | src/compiler/utilitiesPublic.ts:615:1 |
| NotYet: checked predicate overload target ReferencedFile | 1 | 1 | src/compiler/program.ts:1162:1 |
| NotYet: checked predicate overload target TypeElement | 1 | 1 | src/compiler/utilitiesPublic.ts:1782:1 |
| NotYet: checked view field modifiers of type NodeArray<Modifier> &#124; NodeArray<ModifierLike> &#124; undefined | 1 | 1 | src/compiler/utilities.ts:5340:11 |
| NotYet: detached hasOwnProperty through an object view (a value with a different native representation may be hidden by the view) | 1 | 1 | src/compiler/core.ts:1267:32 |
| Refused: a checked view with a callable contract | 1 | 1 | src/compiler/sys.ts:1659:13 |
| Refused: a type predicate whose return is not proven (there is no body proving this parameter) | 33 | 33 | src/compiler/core.ts:1778:100<br>src/compiler/core.ts:1783:97<br>src/compiler/core.ts:2459:112<br>src/compiler/core.ts:2459:66<br>src/compiler/core.ts:2459:91 |
| Refused: an unproven predicate argument for parameter callback (argument "decl => isVariableDeclarationInitializedToRequire(decl)") | 1 | 1 | src/compiler/utilities.ts:3908:53 |
| Refused: an unproven predicate argument for parameter callback (argument "n => isExportAssignment(n) &#124;&#124; isStatement(n) &#124;&#124; isVariableDeclaration(n) &#124;&#124; isPropertyDeclaration(n) &#124;&#124; isParameter(n)") | 1 | 1 | src/compiler/transformers/declarations/diagnostics.ts:708:43 |
| Refused: an unproven predicate argument for parameter f (argument "helper => !helper.scoped") | 1 | 1 | src/compiler/factory/utilities.ts:761:47 |
| Refused: an unproven predicate argument for parameter f (argument "p => isParameterPropertyDeclaration(p, constructor)") | 1 | 1 | src/compiler/transformers/ts.ts:1384:44 |
| Refused: an unproven predicate argument for parameter nodeTest (argument "isModuleDeclaration") | 1 | 1 | src/compiler/transformers/ts.ts:1975:47 |
| Refused: an unproven predicate argument for parameter predicate (argument "modifier => !isExportModifier(modifier) && !isDefaultModifier(modifier)") | 2 | 2 | src/compiler/factory/nodeConverters.ts:66:40<br>src/compiler/factory/nodeConverters.ts:84:36 |
| Refused: an unproven predicate argument for parameter predicate (argument "predicate") | 2 | 2 | src/compiler/utilitiesPublic.ts:1280:52<br>src/compiler/utilitiesPublic.ts:1285:38 |
| Refused: an unproven predicate argument for parameter test (argument "createAssignmentCallback ? isIdentifier : isExpression") | 1 | 1 | src/compiler/transformers/destructuring.ts:185:34 |
| Refused: an unproven predicate argument for parameter test (argument "isNodeArray") | 5 | 5 | src/compiler/factory/parenthesizerRules.ts:688:71<br>src/compiler/factory/parenthesizerRules.ts:694:67<br>src/compiler/factory/parenthesizerRules.ts:696:74<br>src/compiler/factory/parenthesizerRules.ts:701:63<br>src/compiler/factory/parenthesizerRules.ts:704:62 |
