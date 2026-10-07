(a) Fresh literal initialization: **2 sites**.
(b) Adamic source `never`: **47 sites**; stock-checker discrepancy retained below.
(c) Tag-checked class downcast: **17 sites**.
Rest: **34 sites** stay refused; every site has source, target and write witnesses below.
All **100** write-reaching sites occur once; the **291** read-only sites and existing adaptation edits are unchanged.

# October 7 05:30 ruling

This replaces the 04:35 write handoff with exactly one bucket per original write-reaching location.
The full original 100-site inventory and all 58 witnesses remain below for traceability.

## Proof boundaries

Case (a) uses the expressly sanctioned `sys.ts:155` initial-construction pattern. The literal is the allocation arm of `(v || (v = {}))[k] = x`: no second alias to that allocation exists before the immediate initializing write. The primary wider-view binding is part of that expression. This does not preserve freshness after storing, passing, capturing or returning the value. Later writes on the old-value arm have a source binding whose static type already declares the indexed optional members, and are not additional `{}` widening sites.

**Case (b) is a classification of the Adamic census relation, not an independent dead-code proof for upstream tsc.** All 47 records say `Source: never`, including the ruling examples at utilities.ts:4416 (two relations), factory/utilities.ts:363 and factory/nodeFactory.ts:7105. The stock checker does not set `TypeFlags.Never` at any of these 47 located expressions. At checker.ts:52739 its printed type is `never` but its flags are the error/any representation; 46 print inhabited types. The original census and bound stock evidence are both preserved. Under the ruling these 47 static Adamic relations go in (b); the compiler must reconcile the discrepancy before using them as a native unreachable-code proof. No declaration or reachability adaptation is proposed here.

Case (c) requires a positive, dominating tag condition on the same checker-bound value as the downcast. Parentheses and assertions are unwrapped; identifier identity is a checker symbol, not spelling. The analyzer recognizes true branches of `if`, conditional expressions and the right-hand evaluation of `&&`. `TypeFlags.TypeParameter` proves TypeParameter; `TypeFlags.Union` proves UnionType; `TypeFlags.Object` together with `ObjectFlags.Anonymous` proves AnonymousType. Enum members resolve to their declarations. Composite TypeVariable tags, calls returning allegedly appropriate types, and unproved mutable flag aliases do not qualify.

The proof is deliberately local and conservative. The original write graph joins callers and paths. A rest witness is a possible alias-reaching write, not a claim that every path executes it. Rest therefore includes relations for which constructor/result provenance or a flag alias needs more proof, even when a future compiler analysis may establish that they are allowed. Guards license the class view, not unrelated writes to a different shape.

## (a) Fresh literal initialization

| Site in src/compiler | Source | Target | Allocation and immediate write | Escape proof |
|---|---|---|---|---|
| sys.ts:155:35 | {} | Partial<Levels> | src/compiler/sys.ts:155:50; src/compiler/sys.ts:155:17: (customLevels &#124;&#124; (customLevels = {}))[level] = Number(customLevel) | Empty literal has no initializer callbacks, spreads or references. Its only initial binding is the wider-view receiver binding; the same evaluation immediately writes through that receiver. No other alias to this allocation can be stored, passed or captured before this initial write. Freshness ends at that binding/initialization; no later write through an escaped value is licensed. This narrowly follows the expressly sanctioned sys.ts:155 construction pattern, not a general exemption for stored literals. Primary binding: src/compiler/sys.ts:146:13. |
| commandLineParser.ts:3785:33 | {} | CompilerOptions | src/compiler/commandLineParser.ts:3785:50; src/compiler/commandLineParser.ts:3785:13: (defaultOptions &#124;&#124; (defaultOptions = {}))[opt.name] = convertJsonOption(opt, jsonOptions[id], basePath, errors) | Empty literal has no initializer callbacks, spreads or references. Its only initial binding is the wider-view receiver binding; the same evaluation immediately writes through that receiver. No other alias to this allocation can be stored, passed or captured before this initial write. Freshness ends at that binding/initialization; no later write through an escaped value is licensed. This narrowly follows the expressly sanctioned sys.ts:155 construction pattern, not a general exemption for stored literals. Primary binding: src/compiler/commandLineParser.ts:3777:117. |

## (b) Adamic static source never

| Site in src/compiler | Adamic source | Target | Stock bound expression type | Never evidence |
|---|---|---|---|---|
| utilities.ts:4416:20 | never | Expression | (LiteralExpression & StringLiteral) &#124; undefined | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| utilities.ts:4416:52 | never | LiteralExpression & StringLiteral | LiteralExpression & StringLiteral | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| factory/nodeFactory.ts:7105:63 | never | ParameterDeclaration | ParameterDeclaration | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| factory/utilities.ts:363:96 | never | BigIntLiteral | AccessorDeclaration & { readonly name: StringLiteral &#124; Identifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; ComputedPropertyName &#124; BigIntLiteral; } | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| checker.ts:13970:83 | never | LateBoundName | (TypeElement &#124; ClassElement &#124; ObjectLiteralElement) & (LateBoundDeclaration &#124; LateBoundBinaryExpressionDeclaration) | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| checker.ts:20007:71 | never | Expression | LiteralExpression & StringLiteral | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| checker.ts:32496:51 | never | Declaration | (MethodDeclaration & DynamicNamedDeclaration) &#124; (GetAccessorDeclaration & DynamicNamedDeclaration) &#124; (SetAccessorDeclaration & DynamicNamedDeclaration) &#124; (PropertyAssignment & DynamicNamedDeclaration) &#124; (ShorthandPropertyAssignment & DynamicNamedDeclaration) &#124; (SpreadAssignment & DynamicNamedDeclaration) | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| checker.ts:33480:121 | never | ArrayBindingPattern | StringLiteral &#124; Identifier &#124; ObjectBindingPattern &#124; ArrayBindingPattern &#124; NoSubstitutionTemplateLiteral &#124; ... 6 more ... &#124; JsxNamespacedName | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| checker.ts:33486:62 | never | ComputedPropertyName | ComputedPropertyName | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| checker.ts:44963:183 | never | ArrayBindingPattern | Identifier | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| checker.ts:47283:40 | never | Node | StringLiteral &#124; Identifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; ComputedPropertyName &#124; PrivateIdentifier &#124; BigIntLiteral | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| checker.ts:47284:36 | never | Node | ClassElement &#124; ParameterPropertyDeclaration | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| checker.ts:47301:13 | never | Node | ClassElement &#124; ParameterPropertyDeclaration &#124; undefined | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| checker.ts:48433:60 | never | Node | ImportEqualsDeclaration &#124; ImportClause &#124; NamespaceImport &#124; ImportSpecifier &#124; BindingElementOfBareOrAccessedRequire | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| checker.ts:48454:23 | never | Node | AliasDeclarationNode | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| checker.ts:48459:100 | never | Node | ImportEqualsDeclaration &#124; VariableDeclarationInitializedTo<RequireOrImportCall &#124; AccessExpression> &#124; ... 4 more ... &#124; BindingElementOfBareOrAccessedRequire | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| checker.ts:48462:25 | never | Node | ImportEqualsDeclaration &#124; VariableDeclarationInitializedTo<RequireOrImportCall &#124; AccessExpression> &#124; ... 4 more ... &#124; BindingElementOfBareOrAccessedRequire | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| checker.ts:48521:72 | never | Node | ExportSpecifier &#124; VariableDeclarationInitializedTo<RequireOrImportCall &#124; AccessExpression> &#124; ... 4 more ... &#124; BindingElementOfBareOrAccessedRequire | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| checker.ts:48523:27 | never | Node | ExportSpecifier &#124; VariableDeclarationInitializedTo<RequireOrImportCall &#124; AccessExpression> &#124; ... 4 more ... &#124; BindingElementOfBareOrAccessedRequire | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| checker.ts:48523:69 | never | Node | ExportSpecifier &#124; VariableDeclarationInitializedTo<RequireOrImportCall &#124; AccessExpression> &#124; ... 4 more ... &#124; BindingElementOfBareOrAccessedRequire | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| checker.ts:48529:72 | never | Node | ExportSpecifier &#124; ImportClause &#124; NamespaceImport &#124; ImportSpecifier &#124; NamespaceExport &#124; BindingElementOfBareOrAccessedRequire | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| checker.ts:48535:27 | never | Node | ExportSpecifier &#124; ImportClause &#124; NamespaceImport &#124; ImportSpecifier &#124; NamespaceExport &#124; BindingElementOfBareOrAccessedRequire | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| checker.ts:48547:31 | never | Node | AliasDeclarationNode | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| checker.ts:50318:76 | never | Node | LiteralExpression & StringLiteral | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| checker.ts:52739:39 | never | Node | never | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| transformers/ts.ts:1069:52 | never | Node | ParameterPropertyDeclaration | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| transformers/ts.ts:1453:41 | never | Identifier | Identifier | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| transformers/ts.ts:1460:25 | never | Node | ParameterPropertyDeclaration | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| transformers/classFields.ts:694:20 | never | BigIntLiteral | PropertyAssignment & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| transformers/classFields.ts:729:20 | never | ArrayBindingPattern | VariableDeclaration & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| transformers/classFields.ts:753:20 | never | ArrayBindingPattern | ParameterDeclaration & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| transformers/classFields.ts:777:20 | never | ArrayBindingPattern | BindingElement & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| transformers/esDecorators.ts:1412:99 | never | PrivateIdentifier | (node: PrivateIdentifierMethodDeclaration, modifiers: ModifiersArray &#124; undefined) => ObjectLiteralExpression | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| transformers/esDecorators.ts:1427:99 | never | PrivateIdentifier | (node: PrivateIdentifierGetAccessorDeclaration, modifiers: ModifiersArray &#124; undefined) => ObjectLiteralExpression | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| transformers/esDecorators.ts:1442:99 | never | PrivateIdentifier | (node: PrivateIdentifierSetAccessorDeclaration, modifiers: ModifiersArray &#124; undefined) => ObjectLiteralExpression | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| transformers/esDecorators.ts:1974:20 | never | BigIntLiteral | PropertyAssignment & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| transformers/esDecorators.ts:1996:20 | never | ArrayBindingPattern | VariableDeclaration & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| transformers/esDecorators.ts:2020:20 | never | ArrayBindingPattern | BindingElement & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| transformers/module/module.ts:1566:55 | never | ExternalModuleReference | ImportEqualsDeclaration & { moduleReference: ExternalModuleReference; } | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| transformers/module/module.ts:1588:63 | never | ExternalModuleReference | ImportEqualsDeclaration & { moduleReference: ExternalModuleReference; } | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| transformers/module/system.ts:775:73 | never | ExternalModuleReference | ImportEqualsDeclaration & { moduleReference: ExternalModuleReference; } | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| transformers/module/esnextAnd2015.ts:277:55 | never | ExternalModuleReference | ImportEqualsDeclaration & { moduleReference: ExternalModuleReference; } | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| binder.ts:738:67 | never | ArrayBindingPattern | StringLiteral &#124; Identifier &#124; ObjectBindingPattern &#124; ArrayBindingPattern &#124; NoSubstitutionTemplateLiteral &#124; ... 6 more ... &#124; JsxNamespacedName | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| binder.ts:3293:68 | never | Expression | (PropertyAccessExpression & DynamicNamedDeclaration) &#124; (BindablePropertyAssignmentExpression & DynamicNamedDeclaration) &#124; (BindablePropertyAssignmentExpression & DynamicNamedBinaryExpression) &#124; (ElementAccessExpression & ... 2 more ... & DynamicNamedDeclaration) | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| binder.ts:3313:64 | never | Expression | (PropertyAccessExpression & DynamicNamedDeclaration) &#124; (BindablePropertyAssignmentExpression & DynamicNamedDeclaration) &#124; (BindablePropertyAssignmentExpression & DynamicNamedBinaryExpression) &#124; (ElementAccessExpression & ... 2 more ... & DynamicNamedDeclaration) | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| binder.ts:3427:55 | never | Expression | BindablePropertyAssignmentExpression & (DynamicNamedDeclaration &#124; DynamicNamedBinaryExpression) | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |
| binder.ts:3705:86 | never | Node | ParameterPropertyDeclaration | Original refusal census Source=never; stock Never flag=False. Compiler reconciliation required. |

## (c) Guarded class view

| Site in src/compiler | Source | Target | Dominating tag check on the same symbol | Declared member evidence |
|---|---|---|---|---|
| symbolWalker.ts:104:36 | Type | TypeParameter | src/compiler/symbolWalker.ts:103:17: type.flags & TypeFlags.TypeParameter; bound receiver=type | src/compiler/types.ts:6881:5; writes: [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:1881:121 | Type | TypeParameter | src/compiler/checker.ts:1881:46: type && type.flags & TypeFlags.TypeParameter; bound receiver=type | src/compiler/types.ts:6881:5; writes: [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType |
| checker.ts:6942:69 | Type | TypeParameter | src/compiler/checker.ts:6939:21: type.flags & TypeFlags.TypeParameter && contains(context.inferTypeParameters, type); bound receiver=type | src/compiler/types.ts:6881:5; writes: [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:6954:97 | Type | TypeParameter | src/compiler/checker.ts:6939:21: type.flags & TypeFlags.TypeParameter && contains(context.inferTypeParameters, type); bound receiver=type | src/compiler/types.ts:6881:5; writes: [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType |
| checker.ts:13234:71 | Type | TypeParameter | src/compiler/checker.ts:13233:21: baseConstructorType.flags & TypeFlags.TypeParameter; bound receiver=baseConstructorType | src/compiler/types.ts:6881:5; writes: [W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type |
| checker.ts:14959:132 | Type | TypeParameter | src/compiler/checker.ts:14959:44: constraint && constraint.flags & TypeFlags.TypeParameter; bound receiver=constraint | src/compiler/types.ts:6881:5; writes: [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:15035:49 | IntersectionType | AnonymousType | src/compiler/checker.ts:15024:17: type.flags & TypeFlags.Object; bound receiver=type<br>src/compiler/checker.ts:15034:26: (type as ObjectType).objectFlags & ObjectFlags.Anonymous; bound receiver=type | src/compiler/types.ts:6779:5; writes: [W034](#w034): instantiations = new Map<string, TypeReference>()<br>[W045](#w045): members = members<br>[W046](#w046): properties = emptyArray<br>[W047](#w047): callSignatures = callSignatures<br>[W048](#w048): constructSignatures = constructSignatures<br>[W049](#w049): indexInfos = indexInfos<br>[W050](#w050): properties = getNamedMembers(members, type.symbol)<br>[W051](#w051): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W052](#w052): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W012](#w012): members = undefined<br>[W016](#w016): callSignatures = getSignaturesOfSymbol(symbol)<br>[W017](#w017): constructSignatures = constructSignatures<br>[W030](#w030): instantiations = new Map<string, Type>() |
| checker.ts:15149:84 | InstantiableType | TypeParameter | src/compiler/checker.ts:15149:16: type.flags & TypeFlags.TypeParameter; bound receiver=type | src/compiler/types.ts:6881:5; writes: [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:15371:67 | Type | TypeParameter | src/compiler/checker.ts:15370:17: t.flags & TypeFlags.TypeParameter; bound receiver=t | src/compiler/types.ts:6881:5; writes: [W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type |
| checker.ts:20867:24 | Type | TypeParameter | src/compiler/checker.ts:20866:17: typeVariable.flags & TypeFlags.TypeParameter; bound receiver=typeVariable | src/compiler/types.ts:6881:5; writes: [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:22931:116 | Type | TypeParameter | src/compiler/checker.ts:22931:17: source.flags & TypeFlags.TypeParameter && source.symbol?.declarations?.[0]; bound receiver=source | src/compiler/types.ts:6881:5; writes: [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:22932:59 | Type | TypeParameter | src/compiler/checker.ts:22931:17: source.flags & TypeFlags.TypeParameter && source.symbol?.declarations?.[0] && !getConstraintOfType(source as TypeVariable); bound receiver=source | src/compiler/types.ts:6881:5; writes: [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:23160:66 | UnionOrIntersectionType | UnionType | src/compiler/checker.ts:23137:17: target.flags & TypeFlags.Union; bound receiver=target | src/compiler/types.ts:6754:5; writes: [W038](#w038): keyPropertyName = mapByKeyProperty ? keyPropertyName : "" as __String<br>[W039](#w039): constituentMap = mapByKeyProperty |
| checker.ts:25113:86 | Type | TypeParameter | src/compiler/checker.ts:25113:16: type.flags & TypeFlags.TypeParameter; bound receiver=type | src/compiler/types.ts:6881:5; writes: [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:34595:106 | Type | TypeParameter | src/compiler/checker.ts:34593:13: containingType.flags & TypeFlags.TypeParameter; bound receiver=containingType | src/compiler/types.ts:6881:5; writes: [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:34595:166 | Type | TypeParameter | src/compiler/checker.ts:34593:13: containingType.flags & TypeFlags.TypeParameter; bound receiver=containingType | src/compiler/types.ts:6881:5; writes: [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:34614:57 | Type | TypeParameter | src/compiler/checker.ts:34613:17: thisType.flags & TypeFlags.TypeParameter; bound receiver=thisType | src/compiler/types.ts:6881:5; writes: [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |

## Rest: full refused write list

No further source adaptation is made. These require a source declaration or further compiler proof within the three permitted cases. A fresh result from a helper is not a fresh object literal; a tag supplied to a constructor is not itself a dominating tag check; a composite TypeVariable tag does not prove TypeParameter. The generic and sentinel relations retain the original conservative aliases and their exact write witnesses.

| Site in src/compiler | Source type | Target type | First refused p | What is written |
|---|---|---|---|---|
| parser.ts:1854:40 | JSDoc[] | JSDocArray | jsDocCache | [W058](#w058): jsDocCache = tags<br>[W055](#w055): jsDocCache = undefined |
| commandLineParser.ts:3767:87 | TypeAcquisition | CompilerOptions | all | [W053](#w053): all = convertJsonOption(opt, jsonOptions[id], basePath, errors) |
| checker.ts:5568:16 | Type | TypeParameter | constraint | [W035](#w035): constraint = markerSuperType<br>[W036](#w036): constraint = markerSuperTypeForCheck<br>[W013](#w013): isThisType = true<br>[W014](#w014): constraint = type<br>[W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W026](#w026): isThisType = true<br>[W027](#w027): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W029](#w029): target = typeParameter<br>[W031](#w031): mapper = mapper<br>[W037](#w037): constraint = instantiateType(target, makeUnaryTypeMapper(source, syntheticParam))<br>[W043](#w043): target = tp<br>[W044](#w044): mapper = mapper |
| checker.ts:7122:128 | Type | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:7164:93 | Type | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:13286:47 | ObjectType | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:13602:16 | Type | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:14505:34 | UnionType | ObjectType | members | [W045](#w045): members = members<br>[W046](#w046): properties = emptyArray<br>[W047](#w047): callSignatures = callSignatures<br>[W048](#w048): constructSignatures = constructSignatures<br>[W049](#w049): indexInfos = indexInfos<br>[W050](#w050): properties = getNamedMembers(members, type.symbol)<br>[W051](#w051): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W052](#w052): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W012](#w012): members = undefined<br>[W016](#w016): callSignatures = getSignaturesOfSymbol(symbol)<br>[W017](#w017): constructSignatures = constructSignatures |
| checker.ts:14567:34 | IntersectionType | ObjectType | members | [W045](#w045): members = members<br>[W046](#w046): properties = emptyArray<br>[W047](#w047): callSignatures = callSignatures<br>[W048](#w048): constructSignatures = constructSignatures<br>[W049](#w049): indexInfos = indexInfos<br>[W050](#w050): properties = getNamedMembers(members, type.symbol)<br>[W051](#w051): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W052](#w052): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W012](#w012): members = undefined<br>[W016](#w016): callSignatures = getSignaturesOfSymbol(symbol)<br>[W017](#w017): constructSignatures = constructSignatures |
| checker.ts:16568:39 | ObjectType | AnonymousType | target | [W034](#w034): instantiations = new Map<string, TypeReference>()<br>[W022](#w022): mapper = instantiatedSignature.mapper<br>[W030](#w030): instantiations = new Map<string, Type>() |
| checker.ts:16961:69 | IntrinsicType | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:16972:93 | IntrinsicType | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:19938:30 | Type | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:20668:76 | Type | TypeParameter | constraint | [W028](#w028): constraint = noConstraintType |
| checker.ts:20893:98 | IntrinsicType | AnonymousType | target | [W034](#w034): instantiations = new Map<string, TypeReference>()<br>[W045](#w045): members = members<br>[W046](#w046): properties = emptyArray<br>[W047](#w047): callSignatures = callSignatures<br>[W048](#w048): constructSignatures = constructSignatures<br>[W049](#w049): indexInfos = indexInfos<br>[W050](#w050): properties = getNamedMembers(members, type.symbol)<br>[W051](#w051): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W052](#w052): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W012](#w012): members = undefined<br>[W016](#w016): callSignatures = getSignaturesOfSymbol(symbol)<br>[W017](#w017): constructSignatures = constructSignatures<br>[W030](#w030): instantiations = new Map<string, Type>() |
| checker.ts:20952:16 | IntrinsicType | ObjectType | members | [W045](#w045): members = members<br>[W046](#w046): properties = emptyArray<br>[W047](#w047): callSignatures = callSignatures<br>[W048](#w048): constructSignatures = constructSignatures<br>[W049](#w049): indexInfos = indexInfos<br>[W050](#w050): properties = getNamedMembers(members, type.symbol)<br>[W051](#w051): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W052](#w052): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W012](#w012): members = undefined<br>[W016](#w016): callSignatures = getSignaturesOfSymbol(symbol)<br>[W017](#w017): constructSignatures = constructSignatures |
| checker.ts:20952:43 | IntrinsicType | ObjectType | members | [W045](#w045): members = members<br>[W046](#w046): properties = emptyArray<br>[W047](#w047): callSignatures = callSignatures<br>[W048](#w048): constructSignatures = constructSignatures<br>[W049](#w049): indexInfos = indexInfos<br>[W050](#w050): properties = getNamedMembers(members, type.symbol)<br>[W051](#w051): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W052](#w052): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W012](#w012): members = undefined<br>[W016](#w016): callSignatures = getSignaturesOfSymbol(symbol)<br>[W017](#w017): constructSignatures = constructSignatures |
| checker.ts:20967:24 | ObjectType | AnonymousType | target | [W034](#w034): instantiations = new Map<string, TypeReference>()<br>[W030](#w030): instantiations = new Map<string, Type>()<br>[W032](#w032): target = type<br>[W033](#w033): mapper = mapper |
| checker.ts:23674:67 | Type | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:23680:71 | Type | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:23898:60 | Type | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:25200:31 | InterfaceType | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:25200:44 | IndexedAccessType | TypeParameter | default | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:31797:31 | InterfaceType | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:32168:37 | IntrinsicType | ObjectType | members | [W045](#w045): members = members<br>[W046](#w046): properties = emptyArray<br>[W047](#w047): callSignatures = callSignatures<br>[W048](#w048): constructSignatures = constructSignatures<br>[W049](#w049): indexInfos = indexInfos<br>[W050](#w050): properties = getNamedMembers(members, type.symbol)<br>[W051](#w051): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W052](#w052): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W012](#w012): members = undefined<br>[W016](#w016): callSignatures = getSignaturesOfSymbol(symbol)<br>[W017](#w017): constructSignatures = constructSignatures |
| checker.ts:32169:17 | IntrinsicType | ObjectType | members | [W045](#w045): members = members<br>[W046](#w046): properties = emptyArray<br>[W047](#w047): callSignatures = callSignatures<br>[W048](#w048): constructSignatures = constructSignatures<br>[W049](#w049): indexInfos = indexInfos<br>[W050](#w050): properties = getNamedMembers(members, type.symbol)<br>[W051](#w051): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W052](#w052): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W012](#w012): members = undefined<br>[W016](#w016): callSignatures = getSignaturesOfSymbol(symbol)<br>[W017](#w017): constructSignatures = constructSignatures |
| checker.ts:32170:17 | IntrinsicType | ObjectType | members | [W045](#w045): members = members<br>[W046](#w046): properties = emptyArray<br>[W047](#w047): callSignatures = callSignatures<br>[W048](#w048): constructSignatures = constructSignatures<br>[W049](#w049): indexInfos = indexInfos<br>[W050](#w050): properties = getNamedMembers(members, type.symbol)<br>[W051](#w051): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W052](#w052): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W012](#w012): members = undefined<br>[W016](#w016): callSignatures = getSignaturesOfSymbol(symbol)<br>[W017](#w017): constructSignatures = constructSignatures |
| checker.ts:37997:31 | Type | SyntheticDefaultModuleType | syntheticType | [W040](#w040): defaultOnlyType = type |
| checker.ts:38009:31 | Type | SyntheticDefaultModuleType | syntheticType | [W041](#w041): syntheticType = isValidSpreadType(type) ? getSpreadType(type, defaultContainingObject, anonymousSymbol, /*objectFlags*/ 0, /*readonly*/ false) : defaultContainingObject<br>[W042](#w042): syntheticType = type |
| checker.ts:38334:16 | IntrinsicType | ObjectType | members | [W045](#w045): members = members<br>[W046](#w046): properties = emptyArray<br>[W047](#w047): callSignatures = callSignatures<br>[W048](#w048): constructSignatures = constructSignatures<br>[W049](#w049): indexInfos = indexInfos<br>[W050](#w050): properties = getNamedMembers(members, type.symbol)<br>[W051](#w051): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W052](#w052): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W012](#w012): members = undefined<br>[W016](#w016): callSignatures = getSignaturesOfSymbol(symbol)<br>[W017](#w017): constructSignatures = constructSignatures |
| checker.ts:38334:79 | IntrinsicType | ObjectType | members | [W045](#w045): members = members<br>[W046](#w046): properties = emptyArray<br>[W047](#w047): callSignatures = callSignatures<br>[W048](#w048): constructSignatures = constructSignatures<br>[W049](#w049): indexInfos = indexInfos<br>[W050](#w050): properties = getNamedMembers(members, type.symbol)<br>[W051](#w051): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W052](#w052): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W012](#w012): members = undefined<br>[W016](#w016): callSignatures = getSignaturesOfSymbol(symbol)<br>[W017](#w017): constructSignatures = constructSignatures |
| builder.ts:575:36 | DiagnosticRelatedInformation | Diagnostic | reportsUnnecessary | [W007](#w007): reportsUnnecessary = diagnostic.reportsUnnecessary<br>[W008](#w008): reportsDeprecated = diagnostic.reportDeprecated<br>[W009](#w009): source = diagnostic.source<br>[W010](#w010): skippedOn = diagnostic.skippedOn<br>[W011](#w011): relatedInformation = relatedInformation ?<br>            relatedInformation.length ?<br>                relatedInformation.map(r => convertToDiagnosticRelatedInformation(r, diagnosticFilePath, newProgram, toPathInBuildInfoDirectory)) :<br>                [] :<br>            undefined |
| builder.ts:1505:48 | ReusableDiagnosticRelatedInformation | ReusableDiagnostic | reportsUnnecessary | [W002](#w002): reportsUnnecessary = diagnostic.reportsUnnecessary<br>[W003](#w003): reportDeprecated = diagnostic.reportsDeprecated<br>[W004](#w004): source = diagnostic.source<br>[W005](#w005): skippedOn = diagnostic.skippedOn<br>[W006](#w006): relatedInformation = relatedInformation ?<br>                relatedInformation.length ?<br>                    relatedInformation.map(r => toReusableDiagnosticRelatedInformation(r, diagnosticFilePath)) :<br>                    [] :<br>                undefined |
| tsbuildPublic.ts:313:18 | SolutionBuilderHostBase<T> | SolutionBuilderHost<T> | reportErrorSummary | [W057](#w057): reportErrorSummary = reportErrorSummary |

## Reproduction and failing proofs

The saved wave-2 tree supplies the exact line coordinates of the 100-site handoff. Source pin and original alias analysis are in [provenance](evidence/read-write/provenance.json) and [analysis](evidence/read-write/analysis.json.gz). The rebucketing uses the same strict checker options and adds no source edits. The exact per-site proof is in [ruling evidence](evidence/ruling/analysis.json.gz).

```sh
NODE_PATH=STOCK_6_0_3_NODE_MODULES node stage3/adapt/75-optional-widening/rebucket.cjs TREE stage3/adapt/75-optional-widening/evidence/read-write/analysis.json.gz ruling.json > ruling.log 2>&1
NODE_PATH=STOCK_6_0_3_NODE_MODULES node stage3/adapt/75-optional-widening/probe-ruling.cjs MUTANT_TREE mutants.json > mutants.log 2>&1
python3 stage3/adapt/75-optional-widening/render-ruling.py ruling.json stage3/adapt/75-optional-widening/READ_WRITE.md > render.log 2>&1
```

Mutant results and commands are recorded in [mutants](evidence/ruling/mutants.json). Full-apply revalidation is independent of this saved 391-site classification. Full apply passes; the composed inventory is 415 to 391 and the latent meter is 82 to 76. The default full oracle has 106,366 passes and one pre-existing API snapshot failure under the sanctioned lines. See [README.md](README.md#full-apply-revalidation-october-7) and [full evidence](evidence/full-apply/provenance.json).

## Original 100-site write inventory and witness anchors

**100 original sites; 58 distinct assignment witnesses.** The new rest bucket is above. Each W identifier
links to its exact write location, receiver, assigned expression and checker type below.
Sites retain the census line numbers on the wave-2 tree. Columns distinguish repeated
relations on one line. A repeated witness is still listed against every site it reaches.

| Site in src/compiler | Census source type | Census target type | First refused p | What is written |
|---|---|---|---|---|
| sys.ts:155:35 | {} | Partial<Levels> | Low | [W056](#w056): Low = Number(customLevel) |
| utilities.ts:4416:20 | never | Expression | id | [W015](#w015): id = nextNodeId |
| utilities.ts:4416:52 | never | LiteralExpression & StringLiteral | id | [W015](#w015): id = nextNodeId |
| factory/nodeFactory.ts:7105:63 | never | ParameterDeclaration | modifiers | [W054](#w054): modifiers = undefined |
| factory/utilities.ts:363:96 | never | BigIntLiteral | id | [W015](#w015): id = nextNodeId |
| parser.ts:1854:40 | JSDoc[] | JSDocArray | jsDocCache | [W058](#w058): jsDocCache = tags<br>[W055](#w055): jsDocCache = undefined |
| commandLineParser.ts:3767:87 | TypeAcquisition | CompilerOptions | all | [W053](#w053): all = convertJsonOption(opt, jsonOptions[id], basePath, errors) |
| commandLineParser.ts:3785:33 | {} | CompilerOptions | all | [W053](#w053): all = convertJsonOption(opt, jsonOptions[id], basePath, errors) |
| symbolWalker.ts:104:36 | Type | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:1881:121 | Type | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType |
| checker.ts:5568:16 | Type | TypeParameter | constraint | [W035](#w035): constraint = markerSuperType<br>[W036](#w036): constraint = markerSuperTypeForCheck<br>[W013](#w013): isThisType = true<br>[W014](#w014): constraint = type<br>[W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W026](#w026): isThisType = true<br>[W027](#w027): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W029](#w029): target = typeParameter<br>[W031](#w031): mapper = mapper<br>[W037](#w037): constraint = instantiateType(target, makeUnaryTypeMapper(source, syntheticParam))<br>[W043](#w043): target = tp<br>[W044](#w044): mapper = mapper |
| checker.ts:6942:69 | Type | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:6954:97 | Type | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType |
| checker.ts:7122:128 | Type | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:7164:93 | Type | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:13234:71 | Type | TypeParameter | constraint | [W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type |
| checker.ts:13286:47 | ObjectType | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:13602:16 | Type | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:13970:83 | never | LateBoundName | id | [W015](#w015): id = nextNodeId |
| checker.ts:14505:34 | UnionType | ObjectType | members | [W045](#w045): members = members<br>[W046](#w046): properties = emptyArray<br>[W047](#w047): callSignatures = callSignatures<br>[W048](#w048): constructSignatures = constructSignatures<br>[W049](#w049): indexInfos = indexInfos<br>[W050](#w050): properties = getNamedMembers(members, type.symbol)<br>[W051](#w051): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W052](#w052): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W012](#w012): members = undefined<br>[W016](#w016): callSignatures = getSignaturesOfSymbol(symbol)<br>[W017](#w017): constructSignatures = constructSignatures |
| checker.ts:14567:34 | IntersectionType | ObjectType | members | [W045](#w045): members = members<br>[W046](#w046): properties = emptyArray<br>[W047](#w047): callSignatures = callSignatures<br>[W048](#w048): constructSignatures = constructSignatures<br>[W049](#w049): indexInfos = indexInfos<br>[W050](#w050): properties = getNamedMembers(members, type.symbol)<br>[W051](#w051): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W052](#w052): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W012](#w012): members = undefined<br>[W016](#w016): callSignatures = getSignaturesOfSymbol(symbol)<br>[W017](#w017): constructSignatures = constructSignatures |
| checker.ts:14959:132 | Type | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:15035:49 | IntersectionType | AnonymousType | target | [W034](#w034): instantiations = new Map<string, TypeReference>()<br>[W045](#w045): members = members<br>[W046](#w046): properties = emptyArray<br>[W047](#w047): callSignatures = callSignatures<br>[W048](#w048): constructSignatures = constructSignatures<br>[W049](#w049): indexInfos = indexInfos<br>[W050](#w050): properties = getNamedMembers(members, type.symbol)<br>[W051](#w051): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W052](#w052): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W012](#w012): members = undefined<br>[W016](#w016): callSignatures = getSignaturesOfSymbol(symbol)<br>[W017](#w017): constructSignatures = constructSignatures<br>[W030](#w030): instantiations = new Map<string, Type>() |
| checker.ts:15149:84 | InstantiableType | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:15371:67 | Type | TypeParameter | constraint | [W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type |
| checker.ts:16568:39 | ObjectType | AnonymousType | target | [W034](#w034): instantiations = new Map<string, TypeReference>()<br>[W022](#w022): mapper = instantiatedSignature.mapper<br>[W030](#w030): instantiations = new Map<string, Type>() |
| checker.ts:16961:69 | IntrinsicType | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:16972:93 | IntrinsicType | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:19938:30 | Type | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:20007:71 | never | Expression | id | [W015](#w015): id = nextNodeId |
| checker.ts:20668:76 | Type | TypeParameter | constraint | [W028](#w028): constraint = noConstraintType |
| checker.ts:20867:24 | Type | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:20893:98 | IntrinsicType | AnonymousType | target | [W034](#w034): instantiations = new Map<string, TypeReference>()<br>[W045](#w045): members = members<br>[W046](#w046): properties = emptyArray<br>[W047](#w047): callSignatures = callSignatures<br>[W048](#w048): constructSignatures = constructSignatures<br>[W049](#w049): indexInfos = indexInfos<br>[W050](#w050): properties = getNamedMembers(members, type.symbol)<br>[W051](#w051): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W052](#w052): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W012](#w012): members = undefined<br>[W016](#w016): callSignatures = getSignaturesOfSymbol(symbol)<br>[W017](#w017): constructSignatures = constructSignatures<br>[W030](#w030): instantiations = new Map<string, Type>() |
| checker.ts:20952:16 | IntrinsicType | ObjectType | members | [W045](#w045): members = members<br>[W046](#w046): properties = emptyArray<br>[W047](#w047): callSignatures = callSignatures<br>[W048](#w048): constructSignatures = constructSignatures<br>[W049](#w049): indexInfos = indexInfos<br>[W050](#w050): properties = getNamedMembers(members, type.symbol)<br>[W051](#w051): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W052](#w052): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W012](#w012): members = undefined<br>[W016](#w016): callSignatures = getSignaturesOfSymbol(symbol)<br>[W017](#w017): constructSignatures = constructSignatures |
| checker.ts:20952:43 | IntrinsicType | ObjectType | members | [W045](#w045): members = members<br>[W046](#w046): properties = emptyArray<br>[W047](#w047): callSignatures = callSignatures<br>[W048](#w048): constructSignatures = constructSignatures<br>[W049](#w049): indexInfos = indexInfos<br>[W050](#w050): properties = getNamedMembers(members, type.symbol)<br>[W051](#w051): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W052](#w052): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W012](#w012): members = undefined<br>[W016](#w016): callSignatures = getSignaturesOfSymbol(symbol)<br>[W017](#w017): constructSignatures = constructSignatures |
| checker.ts:20967:24 | ObjectType | AnonymousType | target | [W034](#w034): instantiations = new Map<string, TypeReference>()<br>[W030](#w030): instantiations = new Map<string, Type>()<br>[W032](#w032): target = type<br>[W033](#w033): mapper = mapper |
| checker.ts:22931:116 | Type | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:22932:59 | Type | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:23160:66 | UnionOrIntersectionType | UnionType | resolvedReducedType | [W038](#w038): keyPropertyName = mapByKeyProperty ? keyPropertyName : "" as __String<br>[W039](#w039): constituentMap = mapByKeyProperty |
| checker.ts:23674:67 | Type | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:23680:71 | Type | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:23898:60 | Type | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:25113:86 | Type | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:25200:31 | InterfaceType | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:25200:44 | IndexedAccessType | TypeParameter | default | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:31797:31 | InterfaceType | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:32168:37 | IntrinsicType | ObjectType | members | [W045](#w045): members = members<br>[W046](#w046): properties = emptyArray<br>[W047](#w047): callSignatures = callSignatures<br>[W048](#w048): constructSignatures = constructSignatures<br>[W049](#w049): indexInfos = indexInfos<br>[W050](#w050): properties = getNamedMembers(members, type.symbol)<br>[W051](#w051): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W052](#w052): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W012](#w012): members = undefined<br>[W016](#w016): callSignatures = getSignaturesOfSymbol(symbol)<br>[W017](#w017): constructSignatures = constructSignatures |
| checker.ts:32169:17 | IntrinsicType | ObjectType | members | [W045](#w045): members = members<br>[W046](#w046): properties = emptyArray<br>[W047](#w047): callSignatures = callSignatures<br>[W048](#w048): constructSignatures = constructSignatures<br>[W049](#w049): indexInfos = indexInfos<br>[W050](#w050): properties = getNamedMembers(members, type.symbol)<br>[W051](#w051): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W052](#w052): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W012](#w012): members = undefined<br>[W016](#w016): callSignatures = getSignaturesOfSymbol(symbol)<br>[W017](#w017): constructSignatures = constructSignatures |
| checker.ts:32170:17 | IntrinsicType | ObjectType | members | [W045](#w045): members = members<br>[W046](#w046): properties = emptyArray<br>[W047](#w047): callSignatures = callSignatures<br>[W048](#w048): constructSignatures = constructSignatures<br>[W049](#w049): indexInfos = indexInfos<br>[W050](#w050): properties = getNamedMembers(members, type.symbol)<br>[W051](#w051): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W052](#w052): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W012](#w012): members = undefined<br>[W016](#w016): callSignatures = getSignaturesOfSymbol(symbol)<br>[W017](#w017): constructSignatures = constructSignatures |
| checker.ts:32496:51 | never | Declaration | localSymbol | [W001](#w001): localSymbol = local |
| checker.ts:33480:121 | never | ArrayBindingPattern | id | [W015](#w015): id = nextNodeId |
| checker.ts:33486:62 | never | ComputedPropertyName | id | [W015](#w015): id = nextNodeId |
| checker.ts:34595:106 | Type | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:34595:166 | Type | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:34614:57 | Type | TypeParameter | constraint | [W018](#w018): default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType<br>[W019](#w019): default = resolvingDefaultType<br>[W020](#w020): default = defaultType<br>[W021](#w021): default = circularConstraintType<br>[W023](#w023): constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType<br>[W024](#w024): constraint = getInferredTypeParameterConstraint(typeParameter) &#124;&#124; noConstraintType<br>[W025](#w025): constraint = type<br>[W028](#w028): constraint = noConstraintType<br>[W044](#w044): mapper = mapper |
| checker.ts:37997:31 | Type | SyntheticDefaultModuleType | syntheticType | [W040](#w040): defaultOnlyType = type |
| checker.ts:38009:31 | Type | SyntheticDefaultModuleType | syntheticType | [W041](#w041): syntheticType = isValidSpreadType(type) ? getSpreadType(type, defaultContainingObject, anonymousSymbol, /*objectFlags*/ 0, /*readonly*/ false) : defaultContainingObject<br>[W042](#w042): syntheticType = type |
| checker.ts:38334:16 | IntrinsicType | ObjectType | members | [W045](#w045): members = members<br>[W046](#w046): properties = emptyArray<br>[W047](#w047): callSignatures = callSignatures<br>[W048](#w048): constructSignatures = constructSignatures<br>[W049](#w049): indexInfos = indexInfos<br>[W050](#w050): properties = getNamedMembers(members, type.symbol)<br>[W051](#w051): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W052](#w052): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W012](#w012): members = undefined<br>[W016](#w016): callSignatures = getSignaturesOfSymbol(symbol)<br>[W017](#w017): constructSignatures = constructSignatures |
| checker.ts:38334:79 | IntrinsicType | ObjectType | members | [W045](#w045): members = members<br>[W046](#w046): properties = emptyArray<br>[W047](#w047): callSignatures = callSignatures<br>[W048](#w048): constructSignatures = constructSignatures<br>[W049](#w049): indexInfos = indexInfos<br>[W050](#w050): properties = getNamedMembers(members, type.symbol)<br>[W051](#w051): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W052](#w052): objectTypeWithoutAbstractConstructSignatures = typeCopy<br>[W012](#w012): members = undefined<br>[W016](#w016): callSignatures = getSignaturesOfSymbol(symbol)<br>[W017](#w017): constructSignatures = constructSignatures |
| checker.ts:44963:183 | never | ArrayBindingPattern | id | [W015](#w015): id = nextNodeId |
| checker.ts:47283:40 | never | Node | id | [W015](#w015): id = nextNodeId |
| checker.ts:47284:36 | never | Node | id | [W015](#w015): id = nextNodeId |
| checker.ts:47301:13 | never | Node | id | [W015](#w015): id = nextNodeId |
| checker.ts:48433:60 | never | Node | id | [W015](#w015): id = nextNodeId |
| checker.ts:48454:23 | never | Node | id | [W015](#w015): id = nextNodeId |
| checker.ts:48459:100 | never | Node | id | [W015](#w015): id = nextNodeId |
| checker.ts:48462:25 | never | Node | id | [W015](#w015): id = nextNodeId |
| checker.ts:48521:72 | never | Node | id | [W015](#w015): id = nextNodeId |
| checker.ts:48523:27 | never | Node | id | [W015](#w015): id = nextNodeId |
| checker.ts:48523:69 | never | Node | id | [W015](#w015): id = nextNodeId |
| checker.ts:48529:72 | never | Node | id | [W015](#w015): id = nextNodeId |
| checker.ts:48535:27 | never | Node | id | [W015](#w015): id = nextNodeId |
| checker.ts:48547:31 | never | Node | id | [W015](#w015): id = nextNodeId |
| checker.ts:50318:76 | never | Node | id | [W015](#w015): id = nextNodeId |
| checker.ts:52739:39 | never | Node | id | [W015](#w015): id = nextNodeId |
| transformers/ts.ts:1069:52 | never | Node | id | [W015](#w015): id = nextNodeId |
| transformers/ts.ts:1453:41 | never | Identifier | id | [W015](#w015): id = nextNodeId |
| transformers/ts.ts:1460:25 | never | Node | id | [W015](#w015): id = nextNodeId |
| transformers/classFields.ts:694:20 | never | BigIntLiteral | id | [W015](#w015): id = nextNodeId |
| transformers/classFields.ts:729:20 | never | ArrayBindingPattern | id | [W015](#w015): id = nextNodeId |
| transformers/classFields.ts:753:20 | never | ArrayBindingPattern | id | [W015](#w015): id = nextNodeId |
| transformers/classFields.ts:777:20 | never | ArrayBindingPattern | id | [W015](#w015): id = nextNodeId |
| transformers/esDecorators.ts:1412:99 | never | PrivateIdentifier | id | [W015](#w015): id = nextNodeId |
| transformers/esDecorators.ts:1427:99 | never | PrivateIdentifier | id | [W015](#w015): id = nextNodeId |
| transformers/esDecorators.ts:1442:99 | never | PrivateIdentifier | id | [W015](#w015): id = nextNodeId |
| transformers/esDecorators.ts:1974:20 | never | BigIntLiteral | id | [W015](#w015): id = nextNodeId |
| transformers/esDecorators.ts:1996:20 | never | ArrayBindingPattern | id | [W015](#w015): id = nextNodeId |
| transformers/esDecorators.ts:2020:20 | never | ArrayBindingPattern | id | [W015](#w015): id = nextNodeId |
| transformers/module/module.ts:1566:55 | never | ExternalModuleReference | id | [W015](#w015): id = nextNodeId |
| transformers/module/module.ts:1588:63 | never | ExternalModuleReference | id | [W015](#w015): id = nextNodeId |
| transformers/module/system.ts:775:73 | never | ExternalModuleReference | id | [W015](#w015): id = nextNodeId |
| transformers/module/esnextAnd2015.ts:277:55 | never | ExternalModuleReference | id | [W015](#w015): id = nextNodeId |
| builder.ts:575:36 | DiagnosticRelatedInformation | Diagnostic | reportsUnnecessary | [W007](#w007): reportsUnnecessary = diagnostic.reportsUnnecessary<br>[W008](#w008): reportsDeprecated = diagnostic.reportDeprecated<br>[W009](#w009): source = diagnostic.source<br>[W010](#w010): skippedOn = diagnostic.skippedOn<br>[W011](#w011): relatedInformation = relatedInformation ?<br>            relatedInformation.length ?<br>                relatedInformation.map(r => convertToDiagnosticRelatedInformation(r, diagnosticFilePath, newProgram, toPathInBuildInfoDirectory)) :<br>                [] :<br>            undefined |
| builder.ts:1505:48 | ReusableDiagnosticRelatedInformation | ReusableDiagnostic | reportsUnnecessary | [W002](#w002): reportsUnnecessary = diagnostic.reportsUnnecessary<br>[W003](#w003): reportDeprecated = diagnostic.reportsDeprecated<br>[W004](#w004): source = diagnostic.source<br>[W005](#w005): skippedOn = diagnostic.skippedOn<br>[W006](#w006): relatedInformation = relatedInformation ?<br>                relatedInformation.length ?<br>                    relatedInformation.map(r => toReusableDiagnosticRelatedInformation(r, diagnosticFilePath)) :<br>                    [] :<br>                undefined |
| tsbuildPublic.ts:313:18 | SolutionBuilderHostBase<T> | SolutionBuilderHost<T> | reportErrorSummary | [W057](#w057): reportErrorSummary = reportErrorSummary |
| binder.ts:738:67 | never | ArrayBindingPattern | id | [W015](#w015): id = nextNodeId |
| binder.ts:3293:68 | never | Expression | id | [W015](#w015): id = nextNodeId |
| binder.ts:3313:64 | never | Expression | id | [W015](#w015): id = nextNodeId |
| binder.ts:3427:55 | never | Expression | id | [W015](#w015): id = nextNodeId |
| binder.ts:3705:86 | never | Node | id | [W015](#w015): id = nextNodeId |

## Exact write witnesses

### W001

Location: `src/compiler/binder.ts:922:17`. Missing member: `localSymbol`.
Receiver type: `Declaration`. Operation: `=`.

```typescript
node.localSymbol = local
```

Assigned value type: `Symbol`.
Bound property declarations: `src/compiler/types.ts:1759:22`.

### W002

Location: `src/compiler/builder.ts:1506:13`. Missing member: `reportsUnnecessary`.
Receiver type: `ReusableDiagnostic`. Operation: `=`.

```typescript
result.reportsUnnecessary = diagnostic.reportsUnnecessary
```

Assigned value type: `{} | undefined`.
Bound property declarations: `src/compiler/builder.ts:97:5`.

### W003

Location: `src/compiler/builder.ts:1507:13`. Missing member: `reportDeprecated`.
Receiver type: `ReusableDiagnostic`. Operation: `=`.

```typescript
result.reportDeprecated = diagnostic.reportsDeprecated
```

Assigned value type: `{} | undefined`.
Bound property declarations: `src/compiler/builder.ts:98:5`.

### W004

Location: `src/compiler/builder.ts:1508:13`. Missing member: `source`.
Receiver type: `ReusableDiagnostic`. Operation: `=`.

```typescript
result.source = diagnostic.source
```

Assigned value type: `string | undefined`.
Bound property declarations: `src/compiler/builder.ts:99:5`.

### W005

Location: `src/compiler/builder.ts:1509:13`. Missing member: `skippedOn`.
Receiver type: `ReusableDiagnostic`. Operation: `=`.

```typescript
result.skippedOn = diagnostic.skippedOn
```

Assigned value type: `keyof CompilerOptions | undefined`.
Bound property declarations: `src/compiler/builder.ts:101:5`.

### W006

Location: `src/compiler/builder.ts:1511:13`. Missing member: `relatedInformation`.
Receiver type: `ReusableDiagnostic`. Operation: `=`.

```typescript
result.relatedInformation = relatedInformation ?
                relatedInformation.length ?
                    relatedInformation.map(r => toReusableDiagnosticRelatedInformation(r, diagnosticFilePath)) :
                    [] :
                undefined
```

Assigned value type: `ReusableDiagnosticRelatedInformation[] | undefined`.
Bound property declarations: `src/compiler/builder.ts:100:5`.

### W007

Location: `src/compiler/builder.ts:576:9`. Missing member: `reportsUnnecessary`.
Receiver type: `Diagnostic`. Operation: `=`.

```typescript
result.reportsUnnecessary = diagnostic.reportsUnnecessary
```

Assigned value type: `{} | undefined`.
Bound property declarations: `src/compiler/types.ts:7272:5`.

### W008

Location: `src/compiler/builder.ts:577:9`. Missing member: `reportsDeprecated`.
Receiver type: `Diagnostic`. Operation: `=`.

```typescript
result.reportsDeprecated = diagnostic.reportDeprecated
```

Assigned value type: `{} | undefined`.
Bound property declarations: `src/compiler/types.ts:7274:5`.

### W009

Location: `src/compiler/builder.ts:578:9`. Missing member: `source`.
Receiver type: `Diagnostic`. Operation: `=`.

```typescript
result.source = diagnostic.source
```

Assigned value type: `string | undefined`.
Bound property declarations: `src/compiler/types.ts:7275:5`.

### W010

Location: `src/compiler/builder.ts:579:9`. Missing member: `skippedOn`.
Receiver type: `Diagnostic`. Operation: `=`.

```typescript
result.skippedOn = diagnostic.skippedOn
```

Assigned value type: `keyof CompilerOptions | undefined`.
Bound property declarations: `src/compiler/types.ts:7277:22`.

### W011

Location: `src/compiler/builder.ts:581:9`. Missing member: `relatedInformation`.
Receiver type: `Diagnostic`. Operation: `=`.

```typescript
result.relatedInformation = relatedInformation ?
            relatedInformation.length ?
                relatedInformation.map(r => convertToDiagnosticRelatedInformation(r, diagnosticFilePath, newProgram, toPathInBuildInfoDirectory)) :
                [] :
            undefined
```

Assigned value type: `DiagnosticRelatedInformation[] | undefined`.
Bound property declarations: `src/compiler/types.ts:7276:5`.

### W012

Location: `src/compiler/checker.ts:13369:13`. Missing member: `members`.
Receiver type: `InterfaceType`. Operation: `=`.

```typescript
type.members = undefined
```

Assigned value type: `undefined`.
Bound property declarations: `src/compiler/types.ts:6613:22`.

### W013

Location: `src/compiler/checker.ts:13493:17`. Missing member: `isThisType`.
Receiver type: `TypeParameter`. Operation: `=`.

```typescript
type.thisType.isThisType = true
```

Assigned value type: `true`.
Bound property declarations: `src/compiler/types.ts:6889:5`.

### W014

Location: `src/compiler/checker.ts:13494:17`. Missing member: `constraint`.
Receiver type: `TypeParameter`. Operation: `=`.

```typescript
type.thisType.constraint = type
```

Assigned value type: `InterfaceType`.
Bound property declarations: `src/compiler/types.ts:6881:5`.

### W015

Location: `src/compiler/checker.ts:1462:9`. Missing member: `id`.
Receiver type: `Node`. Operation: `=`.

```typescript
node.id = nextNodeId
```

Assigned value type: `number`.
Bound property declarations: `src/compiler/types.ts:947:22`.

### W016

Location: `src/compiler/checker.ts:14662:13`. Missing member: `callSignatures`.
Receiver type: `AnonymousType`. Operation: `=`.

```typescript
type.callSignatures = getSignaturesOfSymbol(symbol)
```

Assigned value type: `Signature[]`.
Bound property declarations: `src/compiler/types.ts:6615:22`.

### W017

Location: `src/compiler/checker.ts:14683:13`. Missing member: `constructSignatures`.
Receiver type: `AnonymousType`. Operation: `=`.

```typescript
type.constructSignatures = constructSignatures
```

Assigned value type: `Signature[]`.
Bound property declarations: `src/compiler/types.ts:6616:22`.

### W018

Location: `src/compiler/checker.ts:15460:17`. Missing member: `default`.
Receiver type: `TypeParameter`. Operation: `=`.

```typescript
typeParameter.default = targetDefault ? instantiateType(targetDefault, typeParameter.mapper) : noConstraintType
```

Assigned value type: `Type`.
Bound property declarations: `src/compiler/types.ts:6883:5`.

### W019

Location: `src/compiler/checker.ts:15464:17`. Missing member: `default`.
Receiver type: `TypeParameter`. Operation: `=`.

```typescript
typeParameter.default = resolvingDefaultType
```

Assigned value type: `ResolvedType`.
Bound property declarations: `src/compiler/types.ts:6883:5`.

### W020

Location: `src/compiler/checker.ts:15469:21`. Missing member: `default`.
Receiver type: `TypeParameter`. Operation: `=`.

```typescript
typeParameter.default = defaultType
```

Assigned value type: `Type`.
Bound property declarations: `src/compiler/types.ts:6883:5`.

### W021

Location: `src/compiler/checker.ts:15475:13`. Missing member: `default`.
Receiver type: `TypeParameter`. Operation: `=`.

```typescript
typeParameter.default = circularConstraintType
```

Assigned value type: `ResolvedType`.
Bound property declarations: `src/compiler/types.ts:6883:5`.

### W022

Location: `src/compiler/checker.ts:16569:17`. Missing member: `mapper`.
Receiver type: `AnonymousType`. Operation: `=`.

```typescript
newReturnType.mapper = instantiatedSignature.mapper
```

Assigned value type: `TypeMapper | undefined`.
Bound property declarations: `src/compiler/types.ts:6780:5`.

### W023

Location: `src/compiler/checker.ts:16845:17`. Missing member: `constraint`.
Receiver type: `TypeParameter`. Operation: `=`.

```typescript
typeParameter.constraint = targetConstraint ? instantiateType(targetConstraint, typeParameter.mapper) : noConstraintType
```

Assigned value type: `Type`.
Bound property declarations: `src/compiler/types.ts:6881:5`.

### W024

Location: `src/compiler/checker.ts:16850:21`. Missing member: `constraint`.
Receiver type: `TypeParameter`. Operation: `=`.

```typescript
typeParameter.constraint = getInferredTypeParameterConstraint(typeParameter) || noConstraintType
```

Assigned value type: `Type`.
Bound property declarations: `src/compiler/types.ts:6881:5`.

### W025

Location: `src/compiler/checker.ts:16859:21`. Missing member: `constraint`.
Receiver type: `TypeParameter`. Operation: `=`.

```typescript
typeParameter.constraint = type
```

Assigned value type: `Type`.
Bound property declarations: `src/compiler/types.ts:6881:5`.

### W026

Location: `src/compiler/checker.ts:17916:9`. Missing member: `isThisType`.
Receiver type: `TypeParameter`. Operation: `=`.

```typescript
type.thisType.isThisType = true
```

Assigned value type: `true`.
Bound property declarations: `src/compiler/types.ts:6889:5`.

### W027

Location: `src/compiler/checker.ts:17917:9`. Missing member: `constraint`.
Receiver type: `TypeParameter`. Operation: `=`.

```typescript
type.thisType.constraint = type
```

Assigned value type: `TupleType & InterfaceTypeWithDeclaredMembers`.
Bound property declarations: `src/compiler/types.ts:6881:5`.

### W028

Location: `src/compiler/checker.ts:20668:75`. Missing member: `constraint`.
Receiver type: `TypeParameter`. Operation: `=`.

```typescript
(tp.restrictiveInstantiation as TypeParameter).constraint = noConstraintType
```

Assigned value type: `ResolvedType`.
Bound property declarations: `src/compiler/types.ts:6881:5`.

### W029

Location: `src/compiler/checker.ts:20674:9`. Missing member: `target`.
Receiver type: `TypeParameter`. Operation: `=`.

```typescript
result.target = typeParameter
```

Assigned value type: `TypeParameter`.
Bound property declarations: `src/compiler/types.ts:6885:5`.

### W030

Location: `src/compiler/checker.ts:20774:17`. Missing member: `instantiations`.
Receiver type: `AnonymousType`. Operation: `=`.

```typescript
target.instantiations = new Map<string, Type>()
```

Assigned value type: `Map<string, Type>`.
Bound property declarations: `src/compiler/types.ts:6781:5`.

### W031

Location: `src/compiler/checker.ts:20975:13`. Missing member: `mapper`.
Receiver type: `TypeParameter`. Operation: `=`.

```typescript
freshTypeParameter.mapper = mapper
```

Assigned value type: `TypeMapper`.
Bound property declarations: `src/compiler/types.ts:6887:5`.

### W032

Location: `src/compiler/checker.ts:20980:9`. Missing member: `target`.
Receiver type: `AnonymousType`. Operation: `=`.

```typescript
result.target = type
```

Assigned value type: `AnonymousType`.
Bound property declarations: `src/compiler/types.ts:6779:5`.

### W033

Location: `src/compiler/checker.ts:20981:9`. Missing member: `mapper`.
Receiver type: `AnonymousType`. Operation: `=`.

```typescript
result.mapper = mapper
```

Assigned value type: `TypeMapper`.
Bound property declarations: `src/compiler/types.ts:6780:5`.

### W034

Location: `src/compiler/checker.ts:2155:5`. Missing member: `instantiations`.
Receiver type: `GenericType`. Operation: `=`.

```typescript
emptyGenericType.instantiations = new Map<string, TypeReference>()
```

Assigned value type: `Map<string, TypeReference>`.
Bound property declarations: `src/compiler/types.ts:6696:5`.

### W035

Location: `src/compiler/checker.ts:2168:5`. Missing member: `constraint`.
Receiver type: `TypeParameter`. Operation: `=`.

```typescript
markerSubType.constraint = markerSuperType
```

Assigned value type: `TypeParameter`.
Bound property declarations: `src/compiler/types.ts:6881:5`.

### W036

Location: `src/compiler/checker.ts:2173:5`. Missing member: `constraint`.
Receiver type: `TypeParameter`. Operation: `=`.

```typescript
markerSubTypeForCheck.constraint = markerSuperTypeForCheck
```

Assigned value type: `TypeParameter`.
Bound property declarations: `src/compiler/types.ts:6881:5`.

### W037

Location: `src/compiler/checker.ts:22933:17`. Missing member: `constraint`.
Receiver type: `TypeParameter`. Operation: `=`.

```typescript
syntheticParam.constraint = instantiateType(target, makeUnaryTypeMapper(source, syntheticParam))
```

Assigned value type: `Type`.
Bound property declarations: `src/compiler/types.ts:6881:5`.

### W038

Location: `src/compiler/checker.ts:28046:13`. Missing member: `keyPropertyName`.
Receiver type: `UnionType`. Operation: `=`.

```typescript
unionType.keyPropertyName = mapByKeyProperty ? keyPropertyName : "" as __String
```

Assigned value type: `__String`.
Bound property declarations: `src/compiler/types.ts:6760:5`.

### W039

Location: `src/compiler/checker.ts:28047:13`. Missing member: `constituentMap`.
Receiver type: `UnionType`. Operation: `=`.

```typescript
unionType.constituentMap = mapByKeyProperty
```

Assigned value type: `Map<number, Type> | undefined`.
Bound property declarations: `src/compiler/types.ts:6762:5`.

### W040

Location: `src/compiler/checker.ts:38000:17`. Missing member: `defaultOnlyType`.
Receiver type: `SyntheticDefaultModuleType`. Operation: `=`.

```typescript
synthType.defaultOnlyType = type
```

Assigned value type: `ResolvedType`.
Bound property declarations: `src/compiler/types.ts:6861:5`.

### W041

Location: `src/compiler/checker.ts:38017:21`. Missing member: `syntheticType`.
Receiver type: `SyntheticDefaultModuleType`. Operation: `=`.

```typescript
synthType.syntheticType = isValidSpreadType(type) ? getSpreadType(type, defaultContainingObject, anonymousSymbol, /*objectFlags*/ 0, /*readonly*/ false) : defaultContainingObject
```

Assigned value type: `Type`.
Bound property declarations: `src/compiler/types.ts:6860:5`.

### W042

Location: `src/compiler/checker.ts:38020:21`. Missing member: `syntheticType`.
Receiver type: `SyntheticDefaultModuleType`. Operation: `=`.

```typescript
synthType.syntheticType = type
```

Assigned value type: `Type`.
Bound property declarations: `src/compiler/types.ts:6860:5`.

### W043

Location: `src/compiler/checker.ts:41631:17`. Missing member: `target`.
Receiver type: `TypeParameter`. Operation: `=`.

```typescript
newTypeParameter.target = tp
```

Assigned value type: `TypeParameter`.
Bound property declarations: `src/compiler/types.ts:6885:5`.

### W044

Location: `src/compiler/checker.ts:41643:17`. Missing member: `mapper`.
Receiver type: `TypeParameter`. Operation: `=`.

```typescript
tp.mapper = mapper
```

Assigned value type: `TypeMapper`.
Bound property declarations: `src/compiler/types.ts:6887:5`.

### W045

Location: `src/compiler/checker.ts:5650:9`. Missing member: `members`.
Receiver type: `ResolvedType`. Operation: `=`.

```typescript
resolved.members = members
```

Assigned value type: `SymbolTable`.
Bound property declarations: `src/compiler/types.ts:6817:5`.

### W046

Location: `src/compiler/checker.ts:5651:9`. Missing member: `properties`.
Receiver type: `ResolvedType`. Operation: `=`.

```typescript
resolved.properties = emptyArray
```

Assigned value type: `never[]`.
Bound property declarations: `src/compiler/types.ts:6818:5`.

### W047

Location: `src/compiler/checker.ts:5652:9`. Missing member: `callSignatures`.
Receiver type: `ResolvedType`. Operation: `=`.

```typescript
resolved.callSignatures = callSignatures
```

Assigned value type: `readonly Signature[]`.
Bound property declarations: `src/compiler/types.ts:6819:5`.

### W048

Location: `src/compiler/checker.ts:5653:9`. Missing member: `constructSignatures`.
Receiver type: `ResolvedType`. Operation: `=`.

```typescript
resolved.constructSignatures = constructSignatures
```

Assigned value type: `readonly Signature[]`.
Bound property declarations: `src/compiler/types.ts:6820:5`.

### W049

Location: `src/compiler/checker.ts:5654:9`. Missing member: `indexInfos`.
Receiver type: `ResolvedType`. Operation: `=`.

```typescript
resolved.indexInfos = indexInfos
```

Assigned value type: `readonly IndexInfo[]`.
Bound property declarations: `src/compiler/types.ts:6821:5`.

### W050

Location: `src/compiler/checker.ts:5656:39`. Missing member: `properties`.
Receiver type: `ResolvedType`. Operation: `=`.

```typescript
resolved.properties = getNamedMembers(members, type.symbol)
```

Assigned value type: `Symbol[]`.
Bound property declarations: `src/compiler/types.ts:6818:5`.

### W051

Location: `src/compiler/checker.ts:5676:9`. Missing member: `objectTypeWithoutAbstractConstructSignatures`.
Receiver type: `ResolvedType`. Operation: `=`.

```typescript
type.objectTypeWithoutAbstractConstructSignatures = typeCopy
```

Assigned value type: `ResolvedType`.
Bound property declarations: `src/compiler/types.ts:6618:22`.

### W052

Location: `src/compiler/checker.ts:5677:9`. Missing member: `objectTypeWithoutAbstractConstructSignatures`.
Receiver type: `ResolvedType`. Operation: `=`.

```typescript
typeCopy.objectTypeWithoutAbstractConstructSignatures = typeCopy
```

Assigned value type: `ResolvedType`.
Bound property declarations: `src/compiler/types.ts:6618:22`.

### W053

Location: `src/compiler/commandLineParser.ts:3785:13`. Missing member: `all`.
Receiver type: `CompilerOptions | WatchOptions | TypeAcquisition`. Operation: `=`.

```typescript
(defaultOptions || (defaultOptions = {}))[opt.name] = convertJsonOption(opt, jsonOptions[id], basePath, errors)
```

Assigned value type: `CompilerOptionsValue`.
Bound property declarations: `src/compiler/types.ts:7415:22`.

### W054

Location: `src/compiler/expressionToTypeNode.ts:469:21`. Missing member: `modifiers`.
Receiver type: `Mutable<ParameterDeclaration>`. Operation: `=`.

```typescript
(visited as Mutable<ParameterDeclaration>).modifiers = undefined
```

Assigned value type: `undefined`.
Bound property declarations: `src/compiler/types.ts:1897:5`.

### W055

Location: `src/compiler/parser.ts:3173:13`. Missing member: `jsDocCache`.
Receiver type: `JSDocArray`. Operation: `=`.

```typescript
node.jsDoc.jsDocCache = undefined
```

Assigned value type: `undefined`.
Bound property declarations: `src/compiler/types.ts:964:5`.

### W056

Location: `src/compiler/sys.ts:155:17`. Missing member: `Low`.
Receiver type: `Partial<Levels>`. Operation: `=`.

```typescript
(customLevels || (customLevels = {}))[level] = Number(customLevel)
```

Assigned value type: `number`.
Bound property declarations: `src/compiler/sys.ts:115:5`.

### W057

Location: `src/compiler/tsbuildPublic.ts:314:5`. Missing member: `reportErrorSummary`.
Receiver type: `SolutionBuilderHost<T>`. Operation: `=`.

```typescript
host.reportErrorSummary = reportErrorSummary
```

Assigned value type: `ReportEmitErrorSummary | undefined`.
Bound property declarations: `src/compiler/tsbuildPublic.ts:242:5`.

### W058

Location: `src/compiler/utilitiesPublic.ts:1267:13`. Missing member: `jsDocCache`.
Receiver type: `JSDocArray`. Operation: `=`.

```typescript
node.jsDoc.jsDocCache = tags
```

Assigned value type: `readonly JSDocTag[]`.
Bound property declarations: `src/compiler/types.ts:964:5`.

## Complete read-only list

**291 sites.** Wait for the compiler checked view; no new owner edit proposed.

| Site in src/compiler | Census source type | Census target type | First refused p |
|---|---|---|---|
| tracing.ts:205:67 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| tracing.ts:206:65 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| sys.ts:1618:77 | System | ModuleResolutionHost | trace |
| sys.ts:1632:43 | { readonly throwIfNoEntry: false; } | StatSyncOptions & { bigint?: false &#124; undefined; throwIfNoEntry: false; } | bigint |
| utilitiesPublic.ts:932:55 | never | Node | id |
| utilitiesPublic.ts:932:81 | never | Identifier | id |
| utilities.ts:1024:47 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| utilities.ts:1212:46 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| utilities.ts:1247:42 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| utilities.ts:1267:24 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| utilities.ts:1282:38 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| utilities.ts:2010:30 | LiteralLikeNode | TemplateLiteralLikeNode | rawText |
| utilities.ts:2528:67 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| utilities.ts:2529:65 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| utilities.ts:2533:70 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| utilities.ts:3481:80 | never | Node | id |
| utilities.ts:3780:14 | never | ExternalModuleReference | id |
| utilities.ts:4369:20 | never | NoSubstitutionTemplateLiteral | templateFlags |
| utilities.ts:4369:52 | never | LiteralExpression & StringLiteral | id |
| utilities.ts:6750:38 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| utilities.ts:6839:25 | never | GetAccessorDeclaration | modifiers |
| utilities.ts:6841:27 | never | GetAccessorDeclaration | modifiers |
| utilities.ts:6844:27 | never | SetAccessorDeclaration | modifiers |
| utilities.ts:7946:37 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| utilities.ts:7951:37 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| utilities.ts:7961:37 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| utilities.ts:7973:37 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| utilities.ts:7979:37 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| utilities.ts:8807:39 | DiagnosticRelatedInformation | Diagnostic | reportsUnnecessary |
| utilities.ts:8807:44 | DiagnosticRelatedInformation | Diagnostic | reportsUnnecessary |
| utilities.ts:9231:41 | Pick<CompilerOptions, "noImplicitAny" &#124; "strict"> | CompilerOptions | all |
| utilities.ts:9237:41 | Pick<CompilerOptions, "noImplicitThis" &#124; "strict"> | CompilerOptions | all |
| utilities.ts:9243:41 | Pick<CompilerOptions, "strict" &#124; "strictNullChecks"> | CompilerOptions | all |
| utilities.ts:9249:41 | Pick<CompilerOptions, "strict" &#124; "strictFunctionTypes"> | CompilerOptions | all |
| utilities.ts:9255:41 | Pick<CompilerOptions, "strict" &#124; "strictBindCallApply"> | CompilerOptions | all |
| utilities.ts:9261:41 | Pick<CompilerOptions, "strict" &#124; "strictPropertyInitialization"> | CompilerOptions | all |
| utilities.ts:9267:41 | Pick<CompilerOptions, "strict" &#124; "strictBuiltinIteratorReturn"> | CompilerOptions | all |
| utilities.ts:9280:41 | Pick<CompilerOptions, "strict" &#124; "useUnknownInCatchVariables"> | CompilerOptions | all |
| utilities.ts:10955:56 | Type | TypeParameter | constraint |
| utilities.ts:12189:121 | never | NoSubstitutionTemplateLiteral | templateFlags |
| factory/nodeFactory.ts:7105:28 | never | Node | id |
| factory/nodeFactory.ts:7302:56 | never | Node | id |
| factory/nodeFactory.ts:7302:104 | never | BigIntLiteral | id |
| factory/nodeFactory.ts:7488:45 | TextRange | SourceMapRange | source |
| factory/emitNode.ts:133:12 | Node | SourceMapRange | source |
| factory/emitNode.ts:133:45 | Node | SourceMapRange | source |
| factory/utilities.ts:368:58 | never | BigIntLiteral | id |
| parser.ts:2629:13 | NumericLiteral | TemplateLiteralLikeNode | rawText |
| parser.ts:2630:13 | NumericLiteral | TemplateLiteralLikeNode | rawText |
| parser.ts:3768:13 | NumericLiteral | TemplateLiteralLikeNode | rawText |
| parser.ts:10032:36 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| parser.ts:10032:48 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| parser.ts:10104:37 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| parser.ts:10261:42 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| parser.ts:10641:111 | { resolutionMode: ModuleKind.CommonJS &#124; ModuleKind.ESNext; } | FileReference | preserve |
| parser.ts:10641:158 | { preserve: true; } | FileReference | resolutionMode |
| parser.ts:10644:104 | { preserve: true; } | FileReference | resolutionMode |
| parser.ts:10647:100 | { preserve: true; } | FileReference | resolutionMode |
| parser.ts:10745:53 | { [index: string]: string &#124; { value: string; pos: number; end: number; }; } | { types?: { value: string; pos: number; end: number; } &#124; undefined; } & { lib?: { value: string; pos: number; end: number; } &#124; undefined; } & { path?: { value: string; pos: number; end: number; } &#124; undefined; } & ... & ... & ... | types |
| parser.ts:10777:45 | { [index: string]: string &#124; undefined; } | { types?: { value: string; pos: number; end: number; } &#124; undefined; } & { lib?: { value: string; pos: number; end: number; } &#124; undefined; } & { path?: { value: string; pos: number; end: number; } &#124; undefined; } & ... & ... & ... | types |
| commandLineParser.ts:1993:117 | {} | WatchOptions | watchFile |
| commandLineParser.ts:2128:12 | OptionsBase | CompilerOptions | all |
| commandLineParser.ts:2679:9 | { include: readonly string[] &#124; undefined; exclude: readonly string[] &#124; undefined; } | TSConfig & { watchOptions?: object &#124; undefined; } | watchOptions |
| commandLineParser.ts:2697:79 | Record<string, CompilerOptionsValue> | CompilerOptions | all |
| moduleNameResolver.ts:130:36 | PackageJsonPathFields | PackageJson | version |
| moduleNameResolver.ts:140:12 | { path: string; extension: string; packageId: PackageId &#124; undefined; resolvedUsingTsExtension: boolean &#124; undefined; } | Resolved | originalPath |
| moduleNameResolver.ts:2409:82 | PackageJsonPathFields | PackageJson | version |
| moduleNameResolver.ts:2422:51 | PackageJsonPathFields | PackageJson | version |
| moduleNameResolver.ts:2432:34 | PackageJsonPathFields | PackageJson | version |
| moduleNameResolver.ts:2494:56 | PackageJsonPathFields | PackageJson | version |
| moduleNameResolver.ts:2497:93 | PackageJsonPathFields | PackageJson | version |
| moduleNameResolver.ts:2498:116 | PackageJsonPathFields | PackageJson | version |
| moduleSpecifiers.ts:200:71 | Pick<SourceFile, "fileName" &#124; "impliedNodeFormat"> | Pick<SourceFile, "fileName" &#124; "impliedNodeFormat" &#124; "packageJsonScope"> | packageJsonScope |
| moduleSpecifiers.ts:243:63 | Pick<SourceFile, "fileName" &#124; "impliedNodeFormat"> | Pick<SourceFile, "fileName" &#124; "impliedNodeFormat" &#124; "packageJsonScope"> | packageJsonScope |
| moduleSpecifiers.ts:881:141 | never | BigIntLiteral | id |
| moduleSpecifiers.ts:1277:30 | { moduleFileToTry: string; } | { moduleFileToTry: string; packageRootPath?: string &#124; undefined; blockedByExports?: true &#124; undefined; verbatimFromExports?: true &#124; undefined; } | packageRootPath |
| checker.ts:2123:133 | Type | TypeParameter | constraint |
| checker.ts:2815:109 | DiagnosticRelatedInformation | Diagnostic | reportsUnnecessary |
| checker.ts:2815:174 | DiagnosticRelatedInformation | Diagnostic | reportsUnnecessary |
| checker.ts:3053:43 | never | Node | id |
| checker.ts:3054:66 | never | Node | id |
| checker.ts:3185:56 | never | Node | id |
| checker.ts:3186:71 | never | Node | id |
| checker.ts:3186:84 | never | Node | id |
| checker.ts:3809:34 | ResolvedRefAndOutputDts | ResolvedRefAndSource | source |
| checker.ts:3809:79 | ResolvedRefAndSource | ResolvedRefAndOutputDts | outputDts |
| checker.ts:4745:17 | never | Identifier | id |
| checker.ts:6948:87 | Type | TypeParameter | constraint |
| checker.ts:6960:54 | Type | TypeParameter | constraint |
| checker.ts:7133:90 | Type | TypeParameter | constraint |
| checker.ts:7620:68 | never | Node | id |
| checker.ts:7635:37 | never | BigIntLiteral | id |
| checker.ts:7688:90 | Declaration | NamedDeclaration | name |
| checker.ts:9267:95 | Type | TypeParameter | constraint |
| checker.ts:9535:29 | never | NamedExports | id |
| checker.ts:9539:33 | never | NamedExports | id |
| checker.ts:11157:106 | never | Node | id |
| checker.ts:11161:101 | never | Node | id |
| checker.ts:11162:24 | never | PrivateIdentifier | id |
| checker.ts:13266:58 | Type | TypeParameter | constraint |
| checker.ts:15025:22 | IntersectionType | ObjectType | members |
| checker.ts:15028:27 | IntersectionType | ObjectType | members |
| checker.ts:15034:27 | IntersectionType | ObjectType | members |
| checker.ts:15166:59 | Type | TypeParameter | constraint |
| checker.ts:15349:68 | Type | TypeParameter | constraint |
| checker.ts:15372:25 | Type | TypeParameter | constraint |
| checker.ts:16205:31 | never | JSDocThisTag | id |
| checker.ts:16597:33 | Type | TypeParameter | constraint |
| checker.ts:19947:47 | Type | TypeParameter | constraint |
| checker.ts:19948:47 | Type | TypeParameter | constraint |
| checker.ts:20740:72 | ArrayTypeNode | NodeWithTypeArguments | typeArguments |
| checker.ts:21339:119 | { errors?: Diagnostic[] &#124; undefined; } | ErrorOutputContainer | skipLogging |
| checker.ts:21482:139 | { errors?: Diagnostic[] &#124; undefined; } | ErrorOutputContainer | skipLogging |
| checker.ts:21575:145 | { errors?: Diagnostic[] &#124; undefined; } | ErrorOutputContainer | skipLogging |
| checker.ts:21578:134 | { errors?: Diagnostic[] &#124; undefined; } | ErrorOutputContainer | skipLogging |
| checker.ts:21665:145 | { errors?: Diagnostic[] &#124; undefined; } | ErrorOutputContainer | skipLogging |
| checker.ts:21668:134 | { errors?: Diagnostic[] &#124; undefined; } | ErrorOutputContainer | skipLogging |
| checker.ts:24384:40 | never | Node | id |
| checker.ts:26159:40 | Declaration | NamedDeclaration | name |
| checker.ts:26515:31 | Type | TypeParameter | constraint |
| checker.ts:26915:182 | Type | TypeParameter | constraint |
| checker.ts:33485:111 | never | Node | id |
| checker.ts:33491:85 | never | Node | id |
| checker.ts:34595:31 | Type | TypeParameter | constraint |
| checker.ts:34827:77 | never | Node | id |
| checker.ts:35380:128 | never | Node | id |
| checker.ts:35790:48 | Type | TypeParameter | constraint |
| checker.ts:36738:62 | DiagnosticRelatedInformation | Diagnostic | reportsUnnecessary |
| checker.ts:38488:20 | never | ArrayBindingPattern | id |
| checker.ts:38497:16 | never | ArrayBindingPattern | id |
| checker.ts:42245:92 | never | Node | id |
| checker.ts:44510:134 | never | Node | id |
| checker.ts:44924:67 | never | Node | id |
| checker.ts:44962:67 | never | Node | id |
| checker.ts:47073:135 | never | Node | id |
| checker.ts:47296:33 | never | Node | id |
| checker.ts:47297:33 | never | Node | id |
| checker.ts:47298:22 | never | Node | id |
| checker.ts:48395:45 | never | Node | id |
| checker.ts:48408:28 | never | Node | id |
| checker.ts:48408:113 | never | Node | id |
| checker.ts:48409:61 | never | Node | id |
| checker.ts:48410:40 | never | Node | id |
| checker.ts:48472:57 | never | Node | id |
| checker.ts:48520:33 | never | Node | id |
| checker.ts:48540:58 | never | Node | id |
| checker.ts:48552:35 | never | Node | id |
| checker.ts:49331:52 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| checker.ts:51475:76 | never | Node | id |
| checker.ts:51487:41 | never | BigIntLiteral | id |
| checker.ts:52410:57 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| checker.ts:52411:55 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| transformers/destructuring.ts:607:63 | never | ArrayLiteralExpression | multiLine |
| transformers/destructuring.ts:617:64 | never | BindingElement | propertyName |
| transformers/classThis.ts:122:59 | Identifier | SourceMapRange | source |
| transformers/namedEvaluation.ts:205:68 | Identifier | SourceMapRange | source |
| transformers/namedEvaluation.ts:265:75 | never | BigIntLiteral | id |
| transformers/namedEvaluation.ts:268:9 | never | BigIntLiteral | id |
| transformers/namedEvaluation.ts:288:9 | never | ShorthandPropertyAssignment | equalsToken |
| transformers/namedEvaluation.ts:311:46 | never | Identifier | id |
| transformers/namedEvaluation.ts:314:9 | never | ArrayBindingPattern | id |
| transformers/namedEvaluation.ts:315:9 | never | ArrayBindingPattern | id |
| transformers/namedEvaluation.ts:341:46 | never | Identifier | id |
| transformers/namedEvaluation.ts:344:9 | never | ArrayBindingPattern | id |
| transformers/namedEvaluation.ts:347:9 | never | ArrayBindingPattern | id |
| transformers/namedEvaluation.ts:373:46 | never | Identifier | id |
| transformers/namedEvaluation.ts:376:9 | never | ArrayBindingPattern | id |
| transformers/namedEvaluation.ts:379:9 | never | ArrayBindingPattern | id |
| transformers/namedEvaluation.ts:396:9 | never | PropertyDeclaration | modifiers |
| transformers/namedEvaluation.ts:437:9 | never | BinaryExpression | id |
| transformers/namedEvaluation.ts:459:9 | never | ExportAssignment | modifiers |
| transformers/ts.ts:995:45 | TextRange | SourceMapRange | source |
| transformers/ts.ts:1059:35 | never | Node | id |
| transformers/ts.ts:1064:21 | never | BigIntLiteral | id |
| transformers/ts.ts:1432:27 | never | Node | id |
| transformers/ts.ts:1457:33 | never | Expression | id |
| transformers/ts.ts:1626:40 | TextRange | SourceMapRange | source |
| transformers/ts.ts:2039:62 | EnumDeclaration | SourceMapRange | source |
| transformers/ts.ts:2042:46 | ModuleDeclaration | SourceMapRange | source |
| transformers/ts.ts:2529:39 | TextRange | SourceMapRange | source |
| transformers/ts.ts:2532:38 | TextRange | SourceMapRange | source |
| transformers/ts.ts:2565:33 | Identifier | SourceMapRange | source |
| transformers/classFields.ts:932:41 | Expression | SourceMapRange | source |
| transformers/classFields.ts:935:47 | Expression | SourceMapRange | source |
| transformers/classFields.ts:2465:42 | ParameterDeclaration | SourceMapRange | source |
| transformers/classFields.ts:2469:42 | TextRange | SourceMapRange | source |
| transformers/classFields.ts:2505:43 | TextRange | SourceMapRange | source |
| transformers/classFields.ts:2611:42 | Identifier | SourceMapRange | source |
| transformers/classFields.ts:3306:50 | Identifier | SourceMapRange | source |
| transformers/legacyDecorators.ts:435:40 | TextRange | SourceMapRange | source |
| transformers/legacyDecorators.ts:517:40 | TextRange | SourceMapRange | source |
| transformers/legacyDecorators.ts:660:35 | TextRange | SourceMapRange | source |
| transformers/legacyDecorators.ts:698:39 | TextRange | SourceMapRange | source |
| transformers/legacyDecorators.ts:836:50 | Identifier | SourceMapRange | source |
| transformers/esDecorators.ts:610:56 | Identifier | SourceMapRange | source |
| transformers/esDecorators.ts:619:56 | Identifier | SourceMapRange | source |
| transformers/esDecorators.ts:893:52 | TextRange | SourceMapRange | source |
| transformers/esDecorators.ts:923:62 | Identifier | SourceMapRange | source |
| transformers/esDecorators.ts:1079:56 | TextRange | SourceMapRange | source |
| transformers/esDecorators.ts:1087:56 | TextRange | SourceMapRange | source |
| transformers/esDecorators.ts:1231:40 | TextRange | SourceMapRange | source |
| transformers/esDecorators.ts:1358:56 | TextRange | SourceMapRange | source |
| transformers/esDecorators.ts:1390:56 | TextRange | SourceMapRange | source |
| transformers/esDecorators.ts:1514:149 | never | PrivateIdentifier | id |
| transformers/esDecorators.ts:1597:45 | Expression | SourceMapRange | source |
| transformers/esDecorators.ts:1600:51 | Expression | SourceMapRange | source |
| transformers/esDecorators.ts:1612:50 | BigIntLiteral | SourceMapRange | source |
| transformers/esDecorators.ts:1702:20 | never | ArrayBindingPattern | id |
| transformers/esDecorators.ts:1720:40 | TextRange | SourceMapRange | source |
| transformers/esDecorators.ts:2287:33 | TextRange | SourceMapRange | source |
| transformers/esDecorators.ts:2295:35 | TextRange | SourceMapRange | source |
| transformers/es2017.ts:630:13 | VariableDeclaration | SourceMapRange | source |
| transformers/es2018.ts:790:51 | Expression | SourceMapRange | source |
| transformers/es2018.ts:794:53 | Expression | SourceMapRange | source |
| transformers/esnext.ts:357:39 | never | ArrayBindingPattern | id |
| transformers/esnext.ts:540:43 | ClassDeclaration | SourceMapRange | source |
| transformers/esnext.ts:597:42 | VariableStatement | SourceMapRange | source |
| transformers/esnext.ts:618:39 | VariableDeclaration | SourceMapRange | source |
| transformers/jsx.ts:111:20 | never | Identifier | id |
| transformers/jsx.ts:115:16 | never | Identifier | id |
| transformers/jsx.ts:171:177 | never | ArrayBindingPattern | id |
| transformers/jsx.ts:359:63 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| transformers/es2015.ts:1504:17 | never | ArrayBindingPattern | id |
| transformers/es2015.ts:1505:17 | never | ArrayBindingPattern | id |
| transformers/es2015.ts:2173:49 | Node | SourceMapRange | source |
| transformers/es2015.ts:2360:35 | BigIntLiteral | SourceMapRange | source |
| transformers/es2015.ts:2369:41 | BigIntLiteral | SourceMapRange | source |
| transformers/es2015.ts:2636:71 | TextRange | SourceMapRange | source |
| transformers/es2015.ts:2805:52 | TextRange | SourceMapRange | source |
| transformers/es2015.ts:3046:52 | TextRange | SourceMapRange | source |
| transformers/es2015.ts:4836:39 | SuperExpression | SourceMapRange | source |
| transformers/generators.ts:752:17 | VariableStatement | SourceMapRange | source |
| transformers/generators.ts:1390:79 | ArrayBindingPattern | SourceMapRange | source |
| transformers/generators.ts:1393:13 | InitializedVariableDeclaration | SourceMapRange | source |
| transformers/generators.ts:2077:50 | Identifier | SourceMapRange | source |
| transformers/module/module.ts:1618:73 | never | ExternalModuleReference | id |
| transformers/module/system.ts:776:80 | never | ExternalModuleReference | id |
| transformers/module/esnextAnd2015.ts:289:73 | never | ExternalModuleReference | id |
| transformers/declarations/diagnostics.ts:249:87 | never | Node | id |
| transformers/declarations.ts:1252:55 | never | FalseLiteral | id |
| transformers/declarations.ts:1252:71 | never | FalseLiteral | id |
| transformers/declarations.ts:1252:101 | never | LiteralExpression & StringLiteral | id |
| transformers/declarations.ts:1264:70 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| transformers/declarations.ts:1264:139 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| transformers/declarations.ts:1328:56 | { diagnosticMessage: DiagnosticMessage; errorNode: ExportAssignment; } | SymbolAccessibilityDiagnostic | typeName |
| transformer.ts:671:25 | {} | CompilerOptions | all |
| emitter.ts:769:39 | CompilerOptions | PrinterOptions | omitTrailingSemicolon |
| emitter.ts:1408:32 | SourceFile | SourceMapSource | skipTrivia |
| emitter.ts:1444:66 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| emitter.ts:3961:112 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| emitter.ts:3961:180 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| emitter.ts:6233:96 | SourceMapSource | SourceFileLike | getPositionOfLineAndCharacter |
| program.ts:470:41 | CompilerOptions | PrinterOptions | omitTrailingSemicolon |
| program.ts:670:67 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| program.ts:712:89 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| program.ts:713:87 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| program.ts:714:58 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| program.ts:732:57 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| program.ts:733:76 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| program.ts:767:89 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| program.ts:1578:67 | CompilerHost | CompilerHost & { onUnRecoverableConfigFileDiagnostic?: DiagnosticReporter &#124; undefined; } | onUnRecoverableConfigFileDiagnostic |
| program.ts:1608:42 | { resolvedModule: ResolvedModuleFull; } | ResolvedModuleWithFailedLookupLocations | failedLookupLocations |
| program.ts:1609:13 | { resolvedModule: ResolvedModuleFull; } | ResolvedModuleWithFailedLookupLocations | failedLookupLocations |
| program.ts:1654:13 | { resolvedTypeReferenceDirective: ResolvedTypeReferenceDirective &#124; undefined; } | ResolvedTypeReferenceDirectiveWithFailedLookupLocations | failedLookupLocations |
| program.ts:2972:42 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| program.ts:3926:57 | ResolvedModuleWithFailedLookupLocations | ResolutionWithFailedLookupLocations | isInvalidated |
| resolutionCache.ts:997:27 | ResolvedTypeReferenceDirectiveWithFailedLookupLocations | CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations | isInvalidated |
| resolutionCache.ts:1026:27 | ResolvedModuleWithFailedLookupLocations | CachedResolvedModuleWithFailedLookupLocations | isInvalidated |
| resolutionCache.ts:1051:26 | ResolvedModuleWithFailedLookupLocations | CachedResolvedModuleWithFailedLookupLocations | isInvalidated |
| watch.ts:250:60 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| watch.ts:771:47 | CompilerOptions | PrinterOptions | omitTrailingSemicolon |
| watch.ts:877:45 | CompilerOptions | PrinterOptions | omitTrailingSemicolon |
| watch.ts:990:53 | IncrementalCompilationOptions | IncrementalProgramOptions<EmitAndSemanticDiagnosticsBuilderProgram> | createProgram |
| watchPublic.ts:475:39 | CompilerOptions | PrinterOptions | omitTrailingSemicolon |
| watchPublic.ts:732:36 | CompilerOptions | PrinterOptions | omitTrailingSemicolon |
| tsbuildPublic.ts:430:18 | SolutionBuilderWithWatchHost<T> | SolutionBuilderHost<T> | reportErrorSummary |
| tsbuildPublic.ts:454:17 | SolutionBuilderHost<T> | ModuleResolutionHost | getGlobalTypingsCacheLocation |
| tsbuildPublic.ts:475:17 | SolutionBuilderHost<T> | ModuleResolutionHost | getGlobalTypingsCacheLocation |
| tsbuildPublic.ts:488:17 | SolutionBuilderHost<T> | ModuleResolutionHost | getGlobalTypingsCacheLocation |
| tsbuildPublic.ts:499:66 | SolutionBuilderHost<T> | ProgramHost<T> & { onUnRecoverableConfigFileDiagnostic?: DiagnosticReporter &#124; undefined; } | onUnRecoverableConfigFileDiagnostic |
| executeCommandLine.ts:109:41 | SourceFile | SourceFileLike | getPositionOfLineAndCharacter |
| executeCommandLine.ts:824:9 | BuildOptions | CompilerOptions | all |
| executeCommandLine.ts:859:66 | BuildOptions | CompilerOptions | all |
| executeCommandLine.ts:860:44 | BuildOptions | CompilerOptions | all |
| executeCommandLine.ts:889:62 | BuildOptions | CompilerOptions | all |
| executeCommandLine.ts:890:39 | BuildOptions | CompilerOptions | all |
| expressionToTypeNode.ts:438:64 | never | StringLiteral | textSourceNode |
| expressionToTypeNode.ts:439:71 | never | StringLiteral | textSourceNode |
| expressionToTypeNode.ts:442:57 | never | FalseLiteral | id |
| binder.ts:3251:52 | never | ShorthandPropertyAssignment | equalsToken |
| binder.ts:3426:124 | never | PropertyAccessEntityNameExpression | _propertyAccessExpressionLikeQualifiedNameBrand |
