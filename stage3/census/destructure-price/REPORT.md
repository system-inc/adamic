| Site (AST binding start) | Type | Public or internal | Members (`*` = method signature) | Calls per lane run |
| --- | --- | --- | --- | --- |
| src/compiler/sys.ts:590 | RecursiveDirectoryWatcherHost | Internal / anonymous | watchDirectory, useCaseSensitiveFileNames, getCurrentDirectory, getAccessibleSortedChildDirectories*, fileSystemEntryExists, realpath*, setTimeout, clearTimeout | Not certified |
| src/compiler/sys.ts:993 | CreateSystemWatchFunctions | Internal / anonymous | pollingWatchFileWorker, getModifiedTime, setTimeout, clearTimeout, fsWatchWorker, fileSystemEntryExists, useCaseSensitiveFileNames, getCurrentDirectory, fsSupportsRecursiveFsWatch, getAccessibleSortedChildDirectories*, realpath*, tscWatchFile, useNonPollingWatchers, tscWatchDirectory, inodeWatching, fsWatchWithTimestamp, sysLog | Not certified |
| src/compiler/utilities.ts:8213 | MutateMapSkippingNewValuesOptions<K, T, U> | Internal / anonymous | onDeleteValue*, onExistingValue* | Not certified |
| src/compiler/utilities.ts:8253 | MutateMapOptions<K, T, U> | Internal / anonymous | createNewValue* | Not certified |
| src/compiler/utilities.ts:11317 | EvaluationResolver | Internal / anonymous | evaluateElementAccessExpression*, evaluateEntityNameExpression* | Not certified |
| src/compiler/parser.ts:1472 | NodeFactory | Public | createNodeArray*, createNumericLiteral*, createStringLiteral*, createLiteralLikeNode*, createIdentifier*, createPrivateIdentifier*, createToken*, createArrayLiteralExpression*, createObjectLiteralExpression*, createPropertyAccessExpression*, createPropertyAccessChain*, createElementAccessExpression*, createElementAccessChain*, createCallExpression*, createCallChain*, createNewExpression*, createParenthesizedExpression*, createBlock*, createVariableStatement*, createExpressionStatement*, createIfStatement*, createWhileStatement*, createForStatement*, createForOfStatement*, createVariableDeclaration*, createVariableDeclarationList* | Not certified |
| src/compiler/moduleSpecifiers.ts:583 | ModuleSpecifierPreferences | Internal / anonymous | getAllowedEndingsInPreferredOrder*, relativePreference, excludeRegexes | Not certified |
| src/compiler/sourcemap.ts:40 | Timer | Internal / anonymous | enter*, exit* | Not certified |
| src/compiler/transformers/ts.ts:236 | TransformationContext | Public | factory, getEmitHelperFactory*, startLexicalEnvironment*, resumeLexicalEnvironment*, endLexicalEnvironment*, hoistVariableDeclaration* | Not certified |
| src/compiler/transformers/classFields.ts:355 | TransformationContext | Public | factory, getEmitHelperFactory*, hoistVariableDeclaration*, endLexicalEnvironment*, startLexicalEnvironment*, resumeLexicalEnvironment*, addBlockScopedVariable* | Not certified |
| src/compiler/transformers/typeSerializer.ts:133 | TransformationContext | Public | factory, hoistVariableDeclaration* | Not certified |
| src/compiler/transformers/legacyDecorators.ts:89 | TransformationContext | Public | factory, getEmitHelperFactory*, hoistVariableDeclaration* | Not certified |
| src/compiler/transformers/esDecorators.ts:296 | TransformationContext | Public | factory, getEmitHelperFactory*, startLexicalEnvironment*, endLexicalEnvironment*, hoistVariableDeclaration* | Not certified |
| src/compiler/transformers/es2017.ts:120 | TransformationContext | Public | factory, getEmitHelperFactory*, resumeLexicalEnvironment*, endLexicalEnvironment*, hoistVariableDeclaration* | Not certified |
| src/compiler/transformers/es2018.ts:155 | TransformationContext | Public | factory, getEmitHelperFactory*, resumeLexicalEnvironment*, endLexicalEnvironment*, hoistVariableDeclaration* | Not certified |
| src/compiler/transformers/es2020.ts:43 | TransformationContext | Public | factory, hoistVariableDeclaration* | Not certified |
| src/compiler/transformers/es2021.ts:26 | TransformationContext | Public | hoistVariableDeclaration*, factory | Not certified |
| src/compiler/transformers/esnext.ts:77 | TransformationContext | Public | factory, getEmitHelperFactory*, hoistVariableDeclaration*, startLexicalEnvironment*, endLexicalEnvironment* | Not certified |
| src/compiler/transformers/jsx.ts:99 | TransformationContext | Public | factory, getEmitHelperFactory* | Not certified |
| src/compiler/transformers/es2016.ts:22 | TransformationContext | Public | factory, hoistVariableDeclaration* | Not certified |
| src/compiler/transformers/es2015.ts:489 | TransformationContext | Public | factory, getEmitHelperFactory*, startLexicalEnvironment*, resumeLexicalEnvironment*, endLexicalEnvironment*, hoistVariableDeclaration* | Not certified |
| src/compiler/transformers/generators.ts:337 | TransformationContext | Public | factory, getEmitHelperFactory*, resumeLexicalEnvironment*, endLexicalEnvironment*, hoistFunctionDeclaration*, hoistVariableDeclaration* | Not certified |
| src/compiler/transformers/module/module.ts:192 | TransformationContext | Public | factory, getEmitHelperFactory*, startLexicalEnvironment*, endLexicalEnvironment*, hoistVariableDeclaration* | Not certified |
| src/compiler/transformers/module/system.ts:143 | TransformationContext | Public | factory, startLexicalEnvironment*, endLexicalEnvironment*, hoistVariableDeclaration* | Not certified |
| src/compiler/transformers/module/esnextAnd2015.ts:68 | TransformationContext | Public | factory, getEmitHelperFactory* | Not certified |
| src/compiler/emitter.ts:771 | Timer | Internal / anonymous | enter*, exit* | Not certified |
| src/compiler/emitter.ts:1215 | PrintHandlers | Public | hasGlobalName*, onEmitNode*, isEmitNotificationEnabled*, substituteNode*, onBeforeEmitNode, onAfterEmitNode, onBeforeEmitNodeArray, onAfterEmitNodeArray, onBeforeEmitToken, onAfterEmitToken | Not certified |
| src/compiler/emitter.ts:1275 | Timer | Internal / anonymous | enter*, exit* | Not certified |
| src/compiler/expressionToTypeNode.ts:199 | Recovery boundary (anonymous) | Internal / anonymous | finalizeBoundary*, startRecoveryScope*, hadError*, markError* | Not certified |

## Stopped: no certified hybrid price

The runtime counter failed a default-callback control. A supplied function is proxied and counted, but a destructuring default runs after that proxy has returned undefined. The control executed the default callback once and observed **zero** counted calls, where **one** was required. This affects the actual PrintHandlers bindings: onEmitNode defaults to noEmitNotification, and substituteNode defaults to noEmitSubstitution. The current counter cannot certify per-member heat. No extra-call price is extrapolated from its partial data.

The independent parser/printer positive control passed: it parsed and printed `const answer = 42;`, recorded factory calls on the parser path, and recorded printer timer calls. That proves the counter fires on these paths; it does not repair the missing default path. The retained instrument.cjs and hooks.cjs are failed measurement prototypes, not validated reusable counters.

I stopped the private counting oracle after demonstrating that gap. Install and build passed (11.124 seconds and 40.014 seconds). Its unfiltered eight-worker test phase was intentionally terminated by SIGTERM after 458.505 seconds. No final test totals exist. No whole-lane heat is claimed.

I also stopped the full census, with shell exit **143**. It emitted only a header and no file records or final relation trailer. Thus the census count is **unknown**, not zero and not 29. Census/AST agreement and individual census misses cannot be established. The header confirms full mode and CommonJSBundle; the driver ran with LATENT_FULL=1 and LATENT_ASSERT_NO_OUTPUT=1, selecting all 54 destructuring files. Its diagnostic arrays are empty and checker_rejected is false; the fixed historical label “measured on a checker-rejected program” does not establish rejection here. Partial timing rows reached binder, builder, builderPublic and builderState, not a completed-file coverage ledger.

Two private type experiments were interrupted earlier to release measured memory pressure: approximately 15 GB was in use against a 16 GB worker limit while oracle and census overlapped. They produced neither completed stock strictFunctionTypes diagnostic ledgers nor the property-road bundled d.ts. No diagnostic totals, exact property-road d.ts diff, or hybrid declaration-identity result are claimed. This is a measurement stop; the method rule remains unchanged.

## Completed static and declaration measurements

The final adapted tree has **79 compiler roots**. The typed AST scan found **54 files** containing object binding patterns and **29 sites** taking a method-signature member. Four are destructured parameters without initializers: sys.ts:590, sys.ts:993, utilities.ts:11317 and moduleSpecifiers.ts:583. A literal initializer-only scan finds **25**; including parameter types finds **29**. These are AST counts, not census counts. Table locations denote binding-pattern starts, which can differ from refusal locations at individual members.

The public declaration build passed after npm ci, using node_modules/.bin/hereby dts (27.6 seconds). Public types are **PrintHandlers, NodeFactory and TransformationContext**. The AST inventory contains **19 public sites** taking **82 method-signature members** under the proposal's public-type rule. These are static binding counts, not a measured valid wrapper implementation or call price. They include internal members taken from a public type and exclude members already declared as function properties.

The other named types are absent from the built public typescript.d.ts. The anonymous recovery boundary is recorded separately as internal source evidence. tsserverlibrary.d.ts re-exports ./typescript.js rather than repeating interface declarations; public presence through that re-export and literal declaration presence are distinguished in the inventory.

The parser binding is src/compiler/parser.ts:1472 in namespace initialization, not a function invocation. The 17 TransformationContext transformer paths are named individually in the leading table and AST evidence.

## Per-site type evidence

The JSON [AST inventory](evidence/ast-sites.json) records enclosing functions, each destructured member, method-signature status, source declaration text and location, and initializer text. `enclosing: <top-level>` at the parser binding denotes namespace initialization, not a function invocation. Exhaustive implementation and this-use attribution remains outstanding. A supplementary private contextual-object-literal scan found providers and unresolved edges, but did not establish every flow, including external public API or test inputs. It is not treated as a completed implementation census.

## Public declaration quotations

### PrintHandlers

```typescript
interface PrintHandlers {
        /**
         * A hook used by the Printer when generating unique names to avoid collisions with
         * globally defined names that exist outside of the current source file.
         */
        hasGlobalName?(name: string): boolean;
        /**
         * A hook used by the Printer to provide notifications prior to emitting a node. A
         * compatible implementation **must** invoke `emitCallback` with the provided `hint` and
         * `node` values.
         * @param hint A hint indicating the intended purpose of the node.
         * @param node The node to emit.
         * @param emitCallback A callback that, when invoked, will emit the node.
         * @example
         * ```ts
         * var printer = createPrinter(printerOptions, {
         *   onEmitNode(hint, node, emitCallback) {
         *     // set up or track state prior to emitting the node...
         *     emitCallback(hint, node);
         *     // restore state after emitting the node...
         *   }
         * });
         * ```
         */
        onEmitNode?(hint: EmitHint, node: Node, emitCallback: (hint: EmitHint, node: Node) => void): void;
        /**
         * A hook used to check if an emit notification is required for a node.
         * @param node The node to emit.
         */
        isEmitNotificationEnabled?(node: Node): boolean;
        /**
         * A hook used by the Printer to perform just-in-time substitution of a node. This is
         * primarily used by node transformations that need to substitute one node for another,
         * such as replacing `myExportedVar` with `exports.myExportedVar`.
         * @param hint A hint indicating the intended purpose of the node.
         * @param node The node to emit.
         * @example
         * ```ts
         * var printer = createPrinter(printerOptions, {
         *   substituteNode(hint, node) {
         *     // perform substitution if necessary...
         *     return node;
         *   }
         * });
         * ```
         */
        substituteNode?(hint: EmitHint, node: Node): Node;
    }
```

### NodeFactory

```typescript
interface NodeFactory {
        createNodeArray<T extends Node>(elements?: readonly T[], hasTrailingComma?: boolean): NodeArray<T>;
        createNumericLiteral(value: string | number, numericLiteralFlags?: TokenFlags): NumericLiteral;
        createBigIntLiteral(value: string | PseudoBigInt): BigIntLiteral;
        createStringLiteral(text: string, isSingleQuote?: boolean): StringLiteral;
        createStringLiteralFromNode(sourceNode: PropertyNameLiteral | PrivateIdentifier, isSingleQuote?: boolean): StringLiteral;
        createRegularExpressionLiteral(text: string): RegularExpressionLiteral;
        createIdentifier(text: string): Identifier;
        /**
         * Create a unique temporary variable.
         * @param recordTempVariable An optional callback used to record the temporary variable name. This
         * should usually be a reference to `hoistVariableDeclaration` from a `TransformationContext`, but
         * can be `undefined` if you plan to record the temporary variable manually.
         * @param reservedInNestedScopes When `true`, reserves the temporary variable name in all nested scopes
         * during emit so that the variable can be referenced in a nested function body. This is an alternative to
         * setting `EmitFlags.ReuseTempVariableScope` on the nested function itself.
         */
        createTempVariable(recordTempVariable: ((node: Identifier) => void) | undefined, reservedInNestedScopes?: boolean): Identifier;
        /**
         * Create a unique temporary variable for use in a loop.
         * @param reservedInNestedScopes When `true`, reserves the temporary variable name in all nested scopes
         * during emit so that the variable can be referenced in a nested function body. This is an alternative to
         * setting `EmitFlags.ReuseTempVariableScope` on the nested function itself.
         */
        createLoopVariable(reservedInNestedScopes?: boolean): Identifier;
        /** Create a unique name based on the supplied text. */
        createUniqueName(text: string, flags?: GeneratedIdentifierFlags): Identifier;
        /** Create a unique name generated for a node. */
        getGeneratedNameForNode(node: Node | undefined, flags?: GeneratedIdentifierFlags): Identifier;
        createPrivateIdentifier(text: string): PrivateIdentifier;
        createUniquePrivateName(text?: string): PrivateIdentifier;
        getGeneratedPrivateNameForNode(node: Node): PrivateIdentifier;
        createToken(token: SyntaxKind.SuperKeyword): SuperExpression;
        createToken(token: SyntaxKind.ThisKeyword): ThisExpression;
        createToken(token: SyntaxKind.NullKeyword): NullLiteral;
        createToken(token: SyntaxKind.TrueKeyword): TrueLiteral;
        createToken(token: SyntaxKind.FalseKeyword): FalseLiteral;
        createToken(token: SyntaxKind.EndOfFileToken): EndOfFileToken;
        createToken(token: SyntaxKind.Unknown): Token<SyntaxKind.Unknown>;
        createToken<TKind extends PunctuationSyntaxKind>(token: TKind): PunctuationToken<TKind>;
        createToken<TKind extends KeywordTypeSyntaxKind>(token: TKind): KeywordTypeNode<TKind>;
        createToken<TKind extends ModifierSyntaxKind>(token: TKind): ModifierToken<TKind>;
        createToken<TKind extends KeywordSyntaxKind>(token: TKind): KeywordToken<TKind>;
        createSuper(): SuperExpression;
        createThis(): ThisExpression;
        createNull(): NullLiteral;
        createTrue(): TrueLiteral;
        createFalse(): FalseLiteral;
        createModifier<T extends ModifierSyntaxKind>(kind: T): ModifierToken<T>;
        createModifiersFromModifierFlags(flags: ModifierFlags): Modifier[] | undefined;
        createQualifiedName(left: EntityName, right: string | Identifier): QualifiedName;
        updateQualifiedName(node: QualifiedName, left: EntityName, right: Identifier): QualifiedName;
        createComputedPropertyName(expression: Expression): ComputedPropertyName;
        updateComputedPropertyName(node: ComputedPropertyName, expression: Expression): ComputedPropertyName;
        createTypeParameterDeclaration(modifiers: readonly Modifier[] | undefined, name: string | Identifier, constraint?: TypeNode, defaultType?: TypeNode): TypeParameterDeclaration;
        updateTypeParameterDeclaration(node: TypeParameterDeclaration, modifiers: readonly Modifier[] | undefined, name: Identifier, constraint: TypeNode | undefined, defaultType: TypeNode | undefined): TypeParameterDeclaration;
        createParameterDeclaration(modifiers: readonly ModifierLike[] | undefined, dotDotDotToken: DotDotDotToken | undefined, name: string | BindingName, questionToken?: QuestionToken, type?: TypeNode, initializer?: Expression): ParameterDeclaration;
        updateParameterDeclaration(node: ParameterDeclaration, modifiers: readonly ModifierLike[] | undefined, dotDotDotToken: DotDotDotToken | undefined, name: string | BindingName, questionToken: QuestionToken | undefined, type: TypeNode | undefined, initializer: Expression | undefined): ParameterDeclaration;
        createDecorator(expression: Expression): Decorator;
        updateDecorator(node: Decorator, expression: Expression): Decorator;
        createPropertySignature(modifiers: readonly Modifier[] | undefined, name: PropertyName | string, questionToken: QuestionToken | undefined, type: TypeNode | undefined): PropertySignature;
        updatePropertySignature(node: PropertySignature, modifiers: readonly Modifier[] | undefined, name: PropertyName, questionToken: QuestionToken | undefined, type: TypeNode | undefined): PropertySignature;
        createPropertyDeclaration(modifiers: readonly ModifierLike[] | undefined, name: string | PropertyName, questionOrExclamationToken: QuestionToken | ExclamationToken | undefined, type: TypeNode | undefined, initializer: Expression | undefined): PropertyDeclaration;
        updatePropertyDeclaration(node: PropertyDeclaration, modifiers: readonly ModifierLike[] | undefined, name: string | PropertyName, questionOrExclamationToken: QuestionToken | ExclamationToken | undefined, type: TypeNode | undefined, initializer: Expression | undefined): PropertyDeclaration;
        createMethodSignature(modifiers: readonly Modifier[] | undefined, name: string | PropertyName, questionToken: QuestionToken | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, parameters: readonly ParameterDeclaration[], type: TypeNode | undefined): MethodSignature;
        updateMethodSignature(node: MethodSignature, modifiers: readonly Modifier[] | undefined, name: PropertyName, questionToken: QuestionToken | undefined, typeParameters: NodeArray<TypeParameterDeclaration> | undefined, parameters: NodeArray<ParameterDeclaration>, type: TypeNode | undefined): MethodSignature;
        createMethodDeclaration(modifiers: readonly ModifierLike[] | undefined, asteriskToken: AsteriskToken | undefined, name: string | PropertyName, questionToken: QuestionToken | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, parameters: readonly ParameterDeclaration[], type: TypeNode | undefined, body: Block | undefined): MethodDeclaration;
        updateMethodDeclaration(node: MethodDeclaration, modifiers: readonly ModifierLike[] | undefined, asteriskToken: AsteriskToken | undefined, name: PropertyName, questionToken: QuestionToken | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, parameters: readonly ParameterDeclaration[], type: TypeNode | undefined, body: Block | undefined): MethodDeclaration;
        createConstructorDeclaration(modifiers: readonly ModifierLike[] | undefined, parameters: readonly ParameterDeclaration[], body: Block | undefined): ConstructorDeclaration;
        updateConstructorDeclaration(node: ConstructorDeclaration, modifiers: readonly ModifierLike[] | undefined, parameters: readonly ParameterDeclaration[], body: Block | undefined): ConstructorDeclaration;
        createGetAccessorDeclaration(modifiers: readonly ModifierLike[] | undefined, name: string | PropertyName, parameters: readonly ParameterDeclaration[], type: TypeNode | undefined, body: Block | undefined): GetAccessorDeclaration;
        updateGetAccessorDeclaration(node: GetAccessorDeclaration, modifiers: readonly ModifierLike[] | undefined, name: PropertyName, parameters: readonly ParameterDeclaration[], type: TypeNode | undefined, body: Block | undefined): GetAccessorDeclaration;
        createSetAccessorDeclaration(modifiers: readonly ModifierLike[] | undefined, name: string | PropertyName, parameters: readonly ParameterDeclaration[], body: Block | undefined): SetAccessorDeclaration;
        updateSetAccessorDeclaration(node: SetAccessorDeclaration, modifiers: readonly ModifierLike[] | undefined, name: PropertyName, parameters: readonly ParameterDeclaration[], body: Block | undefined): SetAccessorDeclaration;
        createCallSignature(typeParameters: readonly TypeParameterDeclaration[] | undefined, parameters: readonly ParameterDeclaration[], type: TypeNode | undefined): CallSignatureDeclaration;
        updateCallSignature(node: CallSignatureDeclaration, typeParameters: NodeArray<TypeParameterDeclaration> | undefined, parameters: NodeArray<ParameterDeclaration>, type: TypeNode | undefined): CallSignatureDeclaration;
        createConstructSignature(typeParameters: readonly TypeParameterDeclaration[] | undefined, parameters: readonly ParameterDeclaration[], type: TypeNode | undefined): ConstructSignatureDeclaration;
        updateConstructSignature(node: ConstructSignatureDeclaration, typeParameters: NodeArray<TypeParameterDeclaration> | undefined, parameters: NodeArray<ParameterDeclaration>, type: TypeNode | undefined): ConstructSignatureDeclaration;
        createIndexSignature(modifiers: readonly ModifierLike[] | undefined, parameters: readonly ParameterDeclaration[], type: TypeNode): IndexSignatureDeclaration;
        updateIndexSignature(node: IndexSignatureDeclaration, modifiers: readonly ModifierLike[] | undefined, parameters: readonly ParameterDeclaration[], type: TypeNode): IndexSignatureDeclaration;
        createTemplateLiteralTypeSpan(type: TypeNode, literal: TemplateMiddle | TemplateTail): TemplateLiteralTypeSpan;
        updateTemplateLiteralTypeSpan(node: TemplateLiteralTypeSpan, type: TypeNode, literal: TemplateMiddle | TemplateTail): TemplateLiteralTypeSpan;
        createClassStaticBlockDeclaration(body: Block): ClassStaticBlockDeclaration;
        updateClassStaticBlockDeclaration(node: ClassStaticBlockDeclaration, body: Block): ClassStaticBlockDeclaration;
        createKeywordTypeNode<TKind extends KeywordTypeSyntaxKind>(kind: TKind): KeywordTypeNode<TKind>;
        createTypePredicateNode(assertsModifier: AssertsKeyword | undefined, parameterName: Identifier | ThisTypeNode | string, type: TypeNode | undefined): TypePredicateNode;
        updateTypePredicateNode(node: TypePredicateNode, assertsModifier: AssertsKeyword | undefined, parameterName: Identifier | ThisTypeNode, type: TypeNode | undefined): TypePredicateNode;
        createTypeReferenceNode(typeName: string | EntityName, typeArguments?: readonly TypeNode[]): TypeReferenceNode;
        updateTypeReferenceNode(node: TypeReferenceNode, typeName: EntityName, typeArguments: NodeArray<TypeNode> | undefined): TypeReferenceNode;
        createFunctionTypeNode(typeParameters: readonly TypeParameterDeclaration[] | undefined, parameters: readonly ParameterDeclaration[], type: TypeNode): FunctionTypeNode;
        updateFunctionTypeNode(node: FunctionTypeNode, typeParameters: NodeArray<TypeParameterDeclaration> | undefined, parameters: NodeArray<ParameterDeclaration>, type: TypeNode): FunctionTypeNode;
        createConstructorTypeNode(modifiers: readonly Modifier[] | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, parameters: readonly ParameterDeclaration[], type: TypeNode): ConstructorTypeNode;
        updateConstructorTypeNode(node: ConstructorTypeNode, modifiers: readonly Modifier[] | undefined, typeParameters: NodeArray<TypeParameterDeclaration> | undefined, parameters: NodeArray<ParameterDeclaration>, type: TypeNode): ConstructorTypeNode;
        createTypeQueryNode(exprName: EntityName, typeArguments?: readonly TypeNode[]): TypeQueryNode;
        updateTypeQueryNode(node: TypeQueryNode, exprName: EntityName, typeArguments?: readonly TypeNode[]): TypeQueryNode;
        createTypeLiteralNode(members: readonly TypeElement[] | undefined): TypeLiteralNode;
        updateTypeLiteralNode(node: TypeLiteralNode, members: NodeArray<TypeElement>): TypeLiteralNode;
        createArrayTypeNode(elementType: TypeNode): ArrayTypeNode;
        updateArrayTypeNode(node: ArrayTypeNode, elementType: TypeNode): ArrayTypeNode;
        createTupleTypeNode(elements: readonly (TypeNode | NamedTupleMember)[]): TupleTypeNode;
        updateTupleTypeNode(node: TupleTypeNode, elements: readonly (TypeNode | NamedTupleMember)[]): TupleTypeNode;
        createNamedTupleMember(dotDotDotToken: DotDotDotToken | undefined, name: Identifier, questionToken: QuestionToken | undefined, type: TypeNode): NamedTupleMember;
        updateNamedTupleMember(node: NamedTupleMember, dotDotDotToken: DotDotDotToken | undefined, name: Identifier, questionToken: QuestionToken | undefined, type: TypeNode): NamedTupleMember;
        createOptionalTypeNode(type: TypeNode): OptionalTypeNode;
        updateOptionalTypeNode(node: OptionalTypeNode, type: TypeNode): OptionalTypeNode;
        createRestTypeNode(type: TypeNode): RestTypeNode;
        updateRestTypeNode(node: RestTypeNode, type: TypeNode): RestTypeNode;
        createUnionTypeNode(types: readonly TypeNode[]): UnionTypeNode;
        updateUnionTypeNode(node: UnionTypeNode, types: NodeArray<TypeNode>): UnionTypeNode;
        createIntersectionTypeNode(types: readonly TypeNode[]): IntersectionTypeNode;
        updateIntersectionTypeNode(node: IntersectionTypeNode, types: NodeArray<TypeNode>): IntersectionTypeNode;
        createConditionalTypeNode(checkType: TypeNode, extendsType: TypeNode, trueType: TypeNode, falseType: TypeNode): ConditionalTypeNode;
        updateConditionalTypeNode(node: ConditionalTypeNode, checkType: TypeNode, extendsType: TypeNode, trueType: TypeNode, falseType: TypeNode): ConditionalTypeNode;
        createInferTypeNode(typeParameter: TypeParameterDeclaration): InferTypeNode;
        updateInferTypeNode(node: InferTypeNode, typeParameter: TypeParameterDeclaration): InferTypeNode;
        createImportTypeNode(argument: TypeNode, attributes?: ImportAttributes, qualifier?: EntityName, typeArguments?: readonly TypeNode[], isTypeOf?: boolean): ImportTypeNode;
        updateImportTypeNode(node: ImportTypeNode, argument: TypeNode, attributes: ImportAttributes | undefined, qualifier: EntityName | undefined, typeArguments: readonly TypeNode[] | undefined, isTypeOf?: boolean): ImportTypeNode;
        createParenthesizedType(type: TypeNode): ParenthesizedTypeNode;
        updateParenthesizedType(node: ParenthesizedTypeNode, type: TypeNode): ParenthesizedTypeNode;
        createThisTypeNode(): ThisTypeNode;
        createTypeOperatorNode(operator: SyntaxKind.KeyOfKeyword | SyntaxKind.UniqueKeyword | SyntaxKind.ReadonlyKeyword, type: TypeNode): TypeOperatorNode;
        updateTypeOperatorNode(node: TypeOperatorNode, type: TypeNode): TypeOperatorNode;
        createIndexedAccessTypeNode(objectType: TypeNode, indexType: TypeNode): IndexedAccessTypeNode;
        updateIndexedAccessTypeNode(node: IndexedAccessTypeNode, objectType: TypeNode, indexType: TypeNode): IndexedAccessTypeNode;
        createMappedTypeNode(readonlyToken: ReadonlyKeyword | PlusToken | MinusToken | undefined, typeParameter: TypeParameterDeclaration, nameType: TypeNode | undefined, questionToken: QuestionToken | PlusToken | MinusToken | undefined, type: TypeNode | undefined, members: NodeArray<TypeElement> | undefined): MappedTypeNode;
        updateMappedTypeNode(node: MappedTypeNode, readonlyToken: ReadonlyKeyword | PlusToken | MinusToken | undefined, typeParameter: TypeParameterDeclaration, nameType: TypeNode | undefined, questionToken: QuestionToken | PlusToken | MinusToken | undefined, type: TypeNode | undefined, members: NodeArray<TypeElement> | undefined): MappedTypeNode;
        createLiteralTypeNode(literal: LiteralTypeNode["literal"]): LiteralTypeNode;
        updateLiteralTypeNode(node: LiteralTypeNode, literal: LiteralTypeNode["literal"]): LiteralTypeNode;
        createTemplateLiteralType(head: TemplateHead, templateSpans: readonly TemplateLiteralTypeSpan[]): TemplateLiteralTypeNode;
        updateTemplateLiteralType(node: TemplateLiteralTypeNode, head: TemplateHead, templateSpans: readonly TemplateLiteralTypeSpan[]): TemplateLiteralTypeNode;
        createObjectBindingPattern(elements: readonly BindingElement[]): ObjectBindingPattern;
        updateObjectBindingPattern(node: ObjectBindingPattern, elements: readonly BindingElement[]): ObjectBindingPattern;
        createArrayBindingPattern(elements: readonly ArrayBindingElement[]): ArrayBindingPattern;
        updateArrayBindingPattern(node: ArrayBindingPattern, elements: readonly ArrayBindingElement[]): ArrayBindingPattern;
        createBindingElement(dotDotDotToken: DotDotDotToken | undefined, propertyName: string | PropertyName | undefined, name: string | BindingName, initializer?: Expression): BindingElement;
        updateBindingElement(node: BindingElement, dotDotDotToken: DotDotDotToken | undefined, propertyName: PropertyName | undefined, name: BindingName, initializer: Expression | undefined): BindingElement;
        createArrayLiteralExpression(elements?: readonly Expression[], multiLine?: boolean): ArrayLiteralExpression;
        updateArrayLiteralExpression(node: ArrayLiteralExpression, elements: readonly Expression[]): ArrayLiteralExpression;
        createObjectLiteralExpression(properties?: readonly ObjectLiteralElementLike[], multiLine?: boolean): ObjectLiteralExpression;
        updateObjectLiteralExpression(node: ObjectLiteralExpression, properties: readonly ObjectLiteralElementLike[]): ObjectLiteralExpression;
        createPropertyAccessExpression(expression: Expression, name: string | MemberName): PropertyAccessExpression;
        updatePropertyAccessExpression(node: PropertyAccessExpression, expression: Expression, name: MemberName): PropertyAccessExpression;
        createPropertyAccessChain(expression: Expression, questionDotToken: QuestionDotToken | undefined, name: string | MemberName): PropertyAccessChain;
        updatePropertyAccessChain(node: PropertyAccessChain, expression: Expression, questionDotToken: QuestionDotToken | undefined, name: MemberName): PropertyAccessChain;
        createElementAccessExpression(expression: Expression, index: number | Expression): ElementAccessExpression;
        updateElementAccessExpression(node: ElementAccessExpression, expression: Expression, argumentExpression: Expression): ElementAccessExpression;
        createElementAccessChain(expression: Expression, questionDotToken: QuestionDotToken | undefined, index: number | Expression): ElementAccessChain;
        updateElementAccessChain(node: ElementAccessChain, expression: Expression, questionDotToken: QuestionDotToken | undefined, argumentExpression: Expression): ElementAccessChain;
        createCallExpression(expression: Expression, typeArguments: readonly TypeNode[] | undefined, argumentsArray: readonly Expression[] | undefined): CallExpression;
        updateCallExpression(node: CallExpression, expression: Expression, typeArguments: readonly TypeNode[] | undefined, argumentsArray: readonly Expression[]): CallExpression;
        createCallChain(expression: Expression, questionDotToken: QuestionDotToken | undefined, typeArguments: readonly TypeNode[] | undefined, argumentsArray: readonly Expression[] | undefined): CallChain;
        updateCallChain(node: CallChain, expression: Expression, questionDotToken: QuestionDotToken | undefined, typeArguments: readonly TypeNode[] | undefined, argumentsArray: readonly Expression[]): CallChain;
        createNewExpression(expression: Expression, typeArguments: readonly TypeNode[] | undefined, argumentsArray: readonly Expression[] | undefined): NewExpression;
        updateNewExpression(node: NewExpression, expression: Expression, typeArguments: readonly TypeNode[] | undefined, argumentsArray: readonly Expression[] | undefined): NewExpression;
        createTaggedTemplateExpression(tag: Expression, typeArguments: readonly TypeNode[] | undefined, template: TemplateLiteral): TaggedTemplateExpression;
        updateTaggedTemplateExpression(node: TaggedTemplateExpression, tag: Expression, typeArguments: readonly TypeNode[] | undefined, template: TemplateLiteral): TaggedTemplateExpression;
        createTypeAssertion(type: TypeNode, expression: Expression): TypeAssertion;
        updateTypeAssertion(node: TypeAssertion, type: TypeNode, expression: Expression): TypeAssertion;
        createParenthesizedExpression(expression: Expression): ParenthesizedExpression;
        updateParenthesizedExpression(node: ParenthesizedExpression, expression: Expression): ParenthesizedExpression;
        createFunctionExpression(modifiers: readonly Modifier[] | undefined, asteriskToken: AsteriskToken | undefined, name: string | Identifier | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, parameters: readonly ParameterDeclaration[] | undefined, type: TypeNode | undefined, body: Block): FunctionExpression;
        updateFunctionExpression(node: FunctionExpression, modifiers: readonly Modifier[] | undefined, asteriskToken: AsteriskToken | undefined, name: Identifier | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, parameters: readonly ParameterDeclaration[], type: TypeNode | undefined, body: Block): FunctionExpression;
        createArrowFunction(modifiers: readonly Modifier[] | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, parameters: readonly ParameterDeclaration[], type: TypeNode | undefined, equalsGreaterThanToken: EqualsGreaterThanToken | undefined, body: ConciseBody): ArrowFunction;
        updateArrowFunction(node: ArrowFunction, modifiers: readonly Modifier[] | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, parameters: readonly ParameterDeclaration[], type: TypeNode | undefined, equalsGreaterThanToken: EqualsGreaterThanToken, body: ConciseBody): ArrowFunction;
        createDeleteExpression(expression: Expression): DeleteExpression;
        updateDeleteExpression(node: DeleteExpression, expression: Expression): DeleteExpression;
        createTypeOfExpression(expression: Expression): TypeOfExpression;
        updateTypeOfExpression(node: TypeOfExpression, expression: Expression): TypeOfExpression;
        createVoidExpression(expression: Expression): VoidExpression;
        updateVoidExpression(node: VoidExpression, expression: Expression): VoidExpression;
        createAwaitExpression(expression: Expression): AwaitExpression;
        updateAwaitExpression(node: AwaitExpression, expression: Expression): AwaitExpression;
        createPrefixUnaryExpression(operator: PrefixUnaryOperator, operand: Expression): PrefixUnaryExpression;
        updatePrefixUnaryExpression(node: PrefixUnaryExpression, operand: Expression): PrefixUnaryExpression;
        createPostfixUnaryExpression(operand: Expression, operator: PostfixUnaryOperator): PostfixUnaryExpression;
        updatePostfixUnaryExpression(node: PostfixUnaryExpression, operand: Expression): PostfixUnaryExpression;
        createBinaryExpression(left: Expression, operator: BinaryOperator | BinaryOperatorToken, right: Expression): BinaryExpression;
        updateBinaryExpression(node: BinaryExpression, left: Expression, operator: BinaryOperator | BinaryOperatorToken, right: Expression): BinaryExpression;
        createConditionalExpression(condition: Expression, questionToken: QuestionToken | undefined, whenTrue: Expression, colonToken: ColonToken | undefined, whenFalse: Expression): ConditionalExpression;
        updateConditionalExpression(node: ConditionalExpression, condition: Expression, questionToken: QuestionToken, whenTrue: Expression, colonToken: ColonToken, whenFalse: Expression): ConditionalExpression;
        createTemplateExpression(head: TemplateHead, templateSpans: readonly TemplateSpan[]): TemplateExpression;
        updateTemplateExpression(node: TemplateExpression, head: TemplateHead, templateSpans: readonly TemplateSpan[]): TemplateExpression;
        createTemplateHead(text: string, rawText?: string, templateFlags?: TokenFlags): TemplateHead;
        createTemplateHead(text: string | undefined, rawText: string, templateFlags?: TokenFlags): TemplateHead;
        createTemplateMiddle(text: string, rawText?: string, templateFlags?: TokenFlags): TemplateMiddle;
        createTemplateMiddle(text: string | undefined, rawText: string, templateFlags?: TokenFlags): TemplateMiddle;
        createTemplateTail(text: string, rawText?: string, templateFlags?: TokenFlags): TemplateTail;
        createTemplateTail(text: string | undefined, rawText: string, templateFlags?: TokenFlags): TemplateTail;
        createNoSubstitutionTemplateLiteral(text: string, rawText?: string): NoSubstitutionTemplateLiteral;
        createNoSubstitutionTemplateLiteral(text: string | undefined, rawText: string): NoSubstitutionTemplateLiteral;
        createYieldExpression(asteriskToken: AsteriskToken, expression: Expression): YieldExpression;
        createYieldExpression(asteriskToken: undefined, expression: Expression | undefined): YieldExpression;
        updateYieldExpression(node: YieldExpression, asteriskToken: AsteriskToken | undefined, expression: Expression | undefined): YieldExpression;
        createSpreadElement(expression: Expression): SpreadElement;
        updateSpreadElement(node: SpreadElement, expression: Expression): SpreadElement;
        createClassExpression(modifiers: readonly ModifierLike[] | undefined, name: string | Identifier | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, heritageClauses: readonly HeritageClause[] | undefined, members: readonly ClassElement[]): ClassExpression;
        updateClassExpression(node: ClassExpression, modifiers: readonly ModifierLike[] | undefined, name: Identifier | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, heritageClauses: readonly HeritageClause[] | undefined, members: readonly ClassElement[]): ClassExpression;
        createOmittedExpression(): OmittedExpression;
        createExpressionWithTypeArguments(expression: Expression, typeArguments: readonly TypeNode[] | undefined): ExpressionWithTypeArguments;
        updateExpressionWithTypeArguments(node: ExpressionWithTypeArguments, expression: Expression, typeArguments: readonly TypeNode[] | undefined): ExpressionWithTypeArguments;
        createAsExpression(expression: Expression, type: TypeNode): AsExpression;
        updateAsExpression(node: AsExpression, expression: Expression, type: TypeNode): AsExpression;
        createNonNullExpression(expression: Expression): NonNullExpression;
        updateNonNullExpression(node: NonNullExpression, expression: Expression): NonNullExpression;
        createNonNullChain(expression: Expression): NonNullChain;
        updateNonNullChain(node: NonNullChain, expression: Expression): NonNullChain;
        createMetaProperty(keywordToken: MetaProperty["keywordToken"], name: Identifier): MetaProperty;
        updateMetaProperty(node: MetaProperty, name: Identifier): MetaProperty;
        createSatisfiesExpression(expression: Expression, type: TypeNode): SatisfiesExpression;
        updateSatisfiesExpression(node: SatisfiesExpression, expression: Expression, type: TypeNode): SatisfiesExpression;
        createTemplateSpan(expression: Expression, literal: TemplateMiddle | TemplateTail): TemplateSpan;
        updateTemplateSpan(node: TemplateSpan, expression: Expression, literal: TemplateMiddle | TemplateTail): TemplateSpan;
        createSemicolonClassElement(): SemicolonClassElement;
        createBlock(statements: readonly Statement[], multiLine?: boolean): Block;
        updateBlock(node: Block, statements: readonly Statement[]): Block;
        createVariableStatement(modifiers: readonly ModifierLike[] | undefined, declarationList: VariableDeclarationList | readonly VariableDeclaration[]): VariableStatement;
        updateVariableStatement(node: VariableStatement, modifiers: readonly ModifierLike[] | undefined, declarationList: VariableDeclarationList): VariableStatement;
        createEmptyStatement(): EmptyStatement;
        createExpressionStatement(expression: Expression): ExpressionStatement;
        updateExpressionStatement(node: ExpressionStatement, expression: Expression): ExpressionStatement;
        createIfStatement(expression: Expression, thenStatement: Statement, elseStatement?: Statement): IfStatement;
        updateIfStatement(node: IfStatement, expression: Expression, thenStatement: Statement, elseStatement: Statement | undefined): IfStatement;
        createDoStatement(statement: Statement, expression: Expression): DoStatement;
        updateDoStatement(node: DoStatement, statement: Statement, expression: Expression): DoStatement;
        createWhileStatement(expression: Expression, statement: Statement): WhileStatement;
        updateWhileStatement(node: WhileStatement, expression: Expression, statement: Statement): WhileStatement;
        createForStatement(initializer: ForInitializer | undefined, condition: Expression | undefined, incrementor: Expression | undefined, statement: Statement): ForStatement;
        updateForStatement(node: ForStatement, initializer: ForInitializer | undefined, condition: Expression | undefined, incrementor: Expression | undefined, statement: Statement): ForStatement;
        createForInStatement(initializer: ForInitializer, expression: Expression, statement: Statement): ForInStatement;
        updateForInStatement(node: ForInStatement, initializer: ForInitializer, expression: Expression, statement: Statement): ForInStatement;
        createForOfStatement(awaitModifier: AwaitKeyword | undefined, initializer: ForInitializer, expression: Expression, statement: Statement): ForOfStatement;
        updateForOfStatement(node: ForOfStatement, awaitModifier: AwaitKeyword | undefined, initializer: ForInitializer, expression: Expression, statement: Statement): ForOfStatement;
        createContinueStatement(label?: string | Identifier): ContinueStatement;
        updateContinueStatement(node: ContinueStatement, label: Identifier | undefined): ContinueStatement;
        createBreakStatement(label?: string | Identifier): BreakStatement;
        updateBreakStatement(node: BreakStatement, label: Identifier | undefined): BreakStatement;
        createReturnStatement(expression?: Expression): ReturnStatement;
        updateReturnStatement(node: ReturnStatement, expression: Expression | undefined): ReturnStatement;
        createWithStatement(expression: Expression, statement: Statement): WithStatement;
        updateWithStatement(node: WithStatement, expression: Expression, statement: Statement): WithStatement;
        createSwitchStatement(expression: Expression, caseBlock: CaseBlock): SwitchStatement;
        updateSwitchStatement(node: SwitchStatement, expression: Expression, caseBlock: CaseBlock): SwitchStatement;
        createLabeledStatement(label: string | Identifier, statement: Statement): LabeledStatement;
        updateLabeledStatement(node: LabeledStatement, label: Identifier, statement: Statement): LabeledStatement;
        createThrowStatement(expression: Expression): ThrowStatement;
        updateThrowStatement(node: ThrowStatement, expression: Expression): ThrowStatement;
        createTryStatement(tryBlock: Block, catchClause: CatchClause | undefined, finallyBlock: Block | undefined): TryStatement;
        updateTryStatement(node: TryStatement, tryBlock: Block, catchClause: CatchClause | undefined, finallyBlock: Block | undefined): TryStatement;
        createDebuggerStatement(): DebuggerStatement;
        createVariableDeclaration(name: string | BindingName, exclamationToken?: ExclamationToken, type?: TypeNode, initializer?: Expression): VariableDeclaration;
        updateVariableDeclaration(node: VariableDeclaration, name: BindingName, exclamationToken: ExclamationToken | undefined, type: TypeNode | undefined, initializer: Expression | undefined): VariableDeclaration;
        createVariableDeclarationList(declarations: readonly VariableDeclaration[], flags?: NodeFlags): VariableDeclarationList;
        updateVariableDeclarationList(node: VariableDeclarationList, declarations: readonly VariableDeclaration[]): VariableDeclarationList;
        createFunctionDeclaration(modifiers: readonly ModifierLike[] | undefined, asteriskToken: AsteriskToken | undefined, name: string | Identifier | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, parameters: readonly ParameterDeclaration[], type: TypeNode | undefined, body: Block | undefined): FunctionDeclaration;
        updateFunctionDeclaration(node: FunctionDeclaration, modifiers: readonly ModifierLike[] | undefined, asteriskToken: AsteriskToken | undefined, name: Identifier | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, parameters: readonly ParameterDeclaration[], type: TypeNode | undefined, body: Block | undefined): FunctionDeclaration;
        createClassDeclaration(modifiers: readonly ModifierLike[] | undefined, name: string | Identifier | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, heritageClauses: readonly HeritageClause[] | undefined, members: readonly ClassElement[]): ClassDeclaration;
        updateClassDeclaration(node: ClassDeclaration, modifiers: readonly ModifierLike[] | undefined, name: Identifier | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, heritageClauses: readonly HeritageClause[] | undefined, members: readonly ClassElement[]): ClassDeclaration;
        createInterfaceDeclaration(modifiers: readonly ModifierLike[] | undefined, name: string | Identifier, typeParameters: readonly TypeParameterDeclaration[] | undefined, heritageClauses: readonly HeritageClause[] | undefined, members: readonly TypeElement[]): InterfaceDeclaration;
        updateInterfaceDeclaration(node: InterfaceDeclaration, modifiers: readonly ModifierLike[] | undefined, name: Identifier, typeParameters: readonly TypeParameterDeclaration[] | undefined, heritageClauses: readonly HeritageClause[] | undefined, members: readonly TypeElement[]): InterfaceDeclaration;
        createTypeAliasDeclaration(modifiers: readonly ModifierLike[] | undefined, name: string | Identifier, typeParameters: readonly TypeParameterDeclaration[] | undefined, type: TypeNode): TypeAliasDeclaration;
        updateTypeAliasDeclaration(node: TypeAliasDeclaration, modifiers: readonly ModifierLike[] | undefined, name: Identifier, typeParameters: readonly TypeParameterDeclaration[] | undefined, type: TypeNode): TypeAliasDeclaration;
        createEnumDeclaration(modifiers: readonly ModifierLike[] | undefined, name: string | Identifier, members: readonly EnumMember[]): EnumDeclaration;
        updateEnumDeclaration(node: EnumDeclaration, modifiers: readonly ModifierLike[] | undefined, name: Identifier, members: readonly EnumMember[]): EnumDeclaration;
        createModuleDeclaration(modifiers: readonly ModifierLike[] | undefined, name: ModuleName, body: ModuleBody | undefined, flags?: NodeFlags): ModuleDeclaration;
        updateModuleDeclaration(node: ModuleDeclaration, modifiers: readonly ModifierLike[] | undefined, name: ModuleName, body: ModuleBody | undefined): ModuleDeclaration;
        createModuleBlock(statements: readonly Statement[]): ModuleBlock;
        updateModuleBlock(node: ModuleBlock, statements: readonly Statement[]): ModuleBlock;
        createCaseBlock(clauses: readonly CaseOrDefaultClause[]): CaseBlock;
        updateCaseBlock(node: CaseBlock, clauses: readonly CaseOrDefaultClause[]): CaseBlock;
        createNamespaceExportDeclaration(name: string | Identifier): NamespaceExportDeclaration;
        updateNamespaceExportDeclaration(node: NamespaceExportDeclaration, name: Identifier): NamespaceExportDeclaration;
        createImportEqualsDeclaration(modifiers: readonly ModifierLike[] | undefined, isTypeOnly: boolean, name: string | Identifier, moduleReference: ModuleReference): ImportEqualsDeclaration;
        updateImportEqualsDeclaration(node: ImportEqualsDeclaration, modifiers: readonly ModifierLike[] | undefined, isTypeOnly: boolean, name: Identifier, moduleReference: ModuleReference): ImportEqualsDeclaration;
        createImportDeclaration(modifiers: readonly ModifierLike[] | undefined, importClause: ImportClause | undefined, moduleSpecifier: Expression, attributes?: ImportAttributes): ImportDeclaration;
        updateImportDeclaration(node: ImportDeclaration, modifiers: readonly ModifierLike[] | undefined, importClause: ImportClause | undefined, moduleSpecifier: Expression, attributes: ImportAttributes | undefined): ImportDeclaration;
        createImportClause(phaseModifier: ImportPhaseModifierSyntaxKind | undefined, name: Identifier | undefined, namedBindings: NamedImportBindings | undefined): ImportClause;
        /** @deprecated */ createImportClause(isTypeOnly: boolean, name: Identifier | undefined, namedBindings: NamedImportBindings | undefined): ImportClause;
        updateImportClause(node: ImportClause, phaseModifier: ImportPhaseModifierSyntaxKind | undefined, name: Identifier | undefined, namedBindings: NamedImportBindings | undefined): ImportClause;
        /** @deprecated */ updateImportClause(node: ImportClause, isTypeOnly: boolean, name: Identifier | undefined, namedBindings: NamedImportBindings | undefined): ImportClause;
        /** @deprecated */ createAssertClause(elements: NodeArray<AssertEntry>, multiLine?: boolean): AssertClause;
        /** @deprecated */ updateAssertClause(node: AssertClause, elements: NodeArray<AssertEntry>, multiLine?: boolean): AssertClause;
        /** @deprecated */ createAssertEntry(name: AssertionKey, value: Expression): AssertEntry;
        /** @deprecated */ updateAssertEntry(node: AssertEntry, name: AssertionKey, value: Expression): AssertEntry;
        /** @deprecated */ createImportTypeAssertionContainer(clause: AssertClause, multiLine?: boolean): ImportTypeAssertionContainer;
        /** @deprecated */ updateImportTypeAssertionContainer(node: ImportTypeAssertionContainer, clause: AssertClause, multiLine?: boolean): ImportTypeAssertionContainer;
        createImportAttributes(elements: NodeArray<ImportAttribute>, multiLine?: boolean): ImportAttributes;
        updateImportAttributes(node: ImportAttributes, elements: NodeArray<ImportAttribute>, multiLine?: boolean): ImportAttributes;
        createImportAttribute(name: ImportAttributeName, value: Expression): ImportAttribute;
        updateImportAttribute(node: ImportAttribute, name: ImportAttributeName, value: Expression): ImportAttribute;
        createNamespaceImport(name: Identifier): NamespaceImport;
        updateNamespaceImport(node: NamespaceImport, name: Identifier): NamespaceImport;
        createNamespaceExport(name: ModuleExportName): NamespaceExport;
        updateNamespaceExport(node: NamespaceExport, name: ModuleExportName): NamespaceExport;
        createNamedImports(elements: readonly ImportSpecifier[]): NamedImports;
        updateNamedImports(node: NamedImports, elements: readonly ImportSpecifier[]): NamedImports;
        createImportSpecifier(isTypeOnly: boolean, propertyName: ModuleExportName | undefined, name: Identifier): ImportSpecifier;
        updateImportSpecifier(node: ImportSpecifier, isTypeOnly: boolean, propertyName: ModuleExportName | undefined, name: Identifier): ImportSpecifier;
        createExportAssignment(modifiers: readonly ModifierLike[] | undefined, isExportEquals: boolean | undefined, expression: Expression): ExportAssignment;
        updateExportAssignment(node: ExportAssignment, modifiers: readonly ModifierLike[] | undefined, expression: Expression): ExportAssignment;
        createExportDeclaration(modifiers: readonly ModifierLike[] | undefined, isTypeOnly: boolean, exportClause: NamedExportBindings | undefined, moduleSpecifier?: Expression, attributes?: ImportAttributes): ExportDeclaration;
        updateExportDeclaration(node: ExportDeclaration, modifiers: readonly ModifierLike[] | undefined, isTypeOnly: boolean, exportClause: NamedExportBindings | undefined, moduleSpecifier: Expression | undefined, attributes: ImportAttributes | undefined): ExportDeclaration;
        createNamedExports(elements: readonly ExportSpecifier[]): NamedExports;
        updateNamedExports(node: NamedExports, elements: readonly ExportSpecifier[]): NamedExports;
        createExportSpecifier(isTypeOnly: boolean, propertyName: string | ModuleExportName | undefined, name: string | ModuleExportName): ExportSpecifier;
        updateExportSpecifier(node: ExportSpecifier, isTypeOnly: boolean, propertyName: ModuleExportName | undefined, name: ModuleExportName): ExportSpecifier;
        createExternalModuleReference(expression: Expression): ExternalModuleReference;
        updateExternalModuleReference(node: ExternalModuleReference, expression: Expression): ExternalModuleReference;
        createJSDocAllType(): JSDocAllType;
        createJSDocUnknownType(): JSDocUnknownType;
        createJSDocNonNullableType(type: TypeNode, postfix?: boolean): JSDocNonNullableType;
        updateJSDocNonNullableType(node: JSDocNonNullableType, type: TypeNode): JSDocNonNullableType;
        createJSDocNullableType(type: TypeNode, postfix?: boolean): JSDocNullableType;
        updateJSDocNullableType(node: JSDocNullableType, type: TypeNode): JSDocNullableType;
        createJSDocOptionalType(type: TypeNode): JSDocOptionalType;
        updateJSDocOptionalType(node: JSDocOptionalType, type: TypeNode): JSDocOptionalType;
        createJSDocFunctionType(parameters: readonly ParameterDeclaration[], type: TypeNode | undefined): JSDocFunctionType;
        updateJSDocFunctionType(node: JSDocFunctionType, parameters: readonly ParameterDeclaration[], type: TypeNode | undefined): JSDocFunctionType;
        createJSDocVariadicType(type: TypeNode): JSDocVariadicType;
        updateJSDocVariadicType(node: JSDocVariadicType, type: TypeNode): JSDocVariadicType;
        createJSDocNamepathType(type: TypeNode): JSDocNamepathType;
        updateJSDocNamepathType(node: JSDocNamepathType, type: TypeNode): JSDocNamepathType;
        createJSDocTypeExpression(type: TypeNode): JSDocTypeExpression;
        updateJSDocTypeExpression(node: JSDocTypeExpression, type: TypeNode): JSDocTypeExpression;
        createJSDocNameReference(name: EntityName | JSDocMemberName): JSDocNameReference;
        updateJSDocNameReference(node: JSDocNameReference, name: EntityName | JSDocMemberName): JSDocNameReference;
        createJSDocMemberName(left: EntityName | JSDocMemberName, right: Identifier): JSDocMemberName;
        updateJSDocMemberName(node: JSDocMemberName, left: EntityName | JSDocMemberName, right: Identifier): JSDocMemberName;
        createJSDocLink(name: EntityName | JSDocMemberName | undefined, text: string): JSDocLink;
        updateJSDocLink(node: JSDocLink, name: EntityName | JSDocMemberName | undefined, text: string): JSDocLink;
        createJSDocLinkCode(name: EntityName | JSDocMemberName | undefined, text: string): JSDocLinkCode;
        updateJSDocLinkCode(node: JSDocLinkCode, name: EntityName | JSDocMemberName | undefined, text: string): JSDocLinkCode;
        createJSDocLinkPlain(name: EntityName | JSDocMemberName | undefined, text: string): JSDocLinkPlain;
        updateJSDocLinkPlain(node: JSDocLinkPlain, name: EntityName | JSDocMemberName | undefined, text: string): JSDocLinkPlain;
        createJSDocTypeLiteral(jsDocPropertyTags?: readonly JSDocPropertyLikeTag[], isArrayType?: boolean): JSDocTypeLiteral;
        updateJSDocTypeLiteral(node: JSDocTypeLiteral, jsDocPropertyTags: readonly JSDocPropertyLikeTag[] | undefined, isArrayType: boolean | undefined): JSDocTypeLiteral;
        createJSDocSignature(typeParameters: readonly JSDocTemplateTag[] | undefined, parameters: readonly JSDocParameterTag[], type?: JSDocReturnTag): JSDocSignature;
        updateJSDocSignature(node: JSDocSignature, typeParameters: readonly JSDocTemplateTag[] | undefined, parameters: readonly JSDocParameterTag[], type: JSDocReturnTag | undefined): JSDocSignature;
        createJSDocTemplateTag(tagName: Identifier | undefined, constraint: JSDocTypeExpression | undefined, typeParameters: readonly TypeParameterDeclaration[], comment?: string | NodeArray<JSDocComment>): JSDocTemplateTag;
        updateJSDocTemplateTag(node: JSDocTemplateTag, tagName: Identifier | undefined, constraint: JSDocTypeExpression | undefined, typeParameters: readonly TypeParameterDeclaration[], comment: string | NodeArray<JSDocComment> | undefined): JSDocTemplateTag;
        createJSDocTypedefTag(tagName: Identifier | undefined, typeExpression?: JSDocTypeExpression | JSDocTypeLiteral, fullName?: Identifier | JSDocNamespaceDeclaration, comment?: string | NodeArray<JSDocComment>): JSDocTypedefTag;
        updateJSDocTypedefTag(node: JSDocTypedefTag, tagName: Identifier | undefined, typeExpression: JSDocTypeExpression | JSDocTypeLiteral | undefined, fullName: Identifier | JSDocNamespaceDeclaration | undefined, comment: string | NodeArray<JSDocComment> | undefined): JSDocTypedefTag;
        createJSDocParameterTag(tagName: Identifier | undefined, name: EntityName, isBracketed: boolean, typeExpression?: JSDocTypeExpression, isNameFirst?: boolean, comment?: string | NodeArray<JSDocComment>): JSDocParameterTag;
        updateJSDocParameterTag(node: JSDocParameterTag, tagName: Identifier | undefined, name: EntityName, isBracketed: boolean, typeExpression: JSDocTypeExpression | undefined, isNameFirst: boolean, comment: string | NodeArray<JSDocComment> | undefined): JSDocParameterTag;
        createJSDocPropertyTag(tagName: Identifier | undefined, name: EntityName, isBracketed: boolean, typeExpression?: JSDocTypeExpression, isNameFirst?: boolean, comment?: string | NodeArray<JSDocComment>): JSDocPropertyTag;
        updateJSDocPropertyTag(node: JSDocPropertyTag, tagName: Identifier | undefined, name: EntityName, isBracketed: boolean, typeExpression: JSDocTypeExpression | undefined, isNameFirst: boolean, comment: string | NodeArray<JSDocComment> | undefined): JSDocPropertyTag;
        createJSDocTypeTag(tagName: Identifier | undefined, typeExpression: JSDocTypeExpression, comment?: string | NodeArray<JSDocComment>): JSDocTypeTag;
        updateJSDocTypeTag(node: JSDocTypeTag, tagName: Identifier | undefined, typeExpression: JSDocTypeExpression, comment: string | NodeArray<JSDocComment> | undefined): JSDocTypeTag;
        createJSDocSeeTag(tagName: Identifier | undefined, nameExpression: JSDocNameReference | undefined, comment?: string | NodeArray<JSDocComment>): JSDocSeeTag;
        updateJSDocSeeTag(node: JSDocSeeTag, tagName: Identifier | undefined, nameExpression: JSDocNameReference | undefined, comment?: string | NodeArray<JSDocComment>): JSDocSeeTag;
        createJSDocReturnTag(tagName: Identifier | undefined, typeExpression?: JSDocTypeExpression, comment?: string | NodeArray<JSDocComment>): JSDocReturnTag;
        updateJSDocReturnTag(node: JSDocReturnTag, tagName: Identifier | undefined, typeExpression: JSDocTypeExpression | undefined, comment: string | NodeArray<JSDocComment> | undefined): JSDocReturnTag;
        createJSDocThisTag(tagName: Identifier | undefined, typeExpression: JSDocTypeExpression, comment?: string | NodeArray<JSDocComment>): JSDocThisTag;
        updateJSDocThisTag(node: JSDocThisTag, tagName: Identifier | undefined, typeExpression: JSDocTypeExpression | undefined, comment: string | NodeArray<JSDocComment> | undefined): JSDocThisTag;
        createJSDocEnumTag(tagName: Identifier | undefined, typeExpression: JSDocTypeExpression, comment?: string | NodeArray<JSDocComment>): JSDocEnumTag;
        updateJSDocEnumTag(node: JSDocEnumTag, tagName: Identifier | undefined, typeExpression: JSDocTypeExpression, comment: string | NodeArray<JSDocComment> | undefined): JSDocEnumTag;
        createJSDocCallbackTag(tagName: Identifier | undefined, typeExpression: JSDocSignature, fullName?: Identifier | JSDocNamespaceDeclaration, comment?: string | NodeArray<JSDocComment>): JSDocCallbackTag;
        updateJSDocCallbackTag(node: JSDocCallbackTag, tagName: Identifier | undefined, typeExpression: JSDocSignature, fullName: Identifier | JSDocNamespaceDeclaration | undefined, comment: string | NodeArray<JSDocComment> | undefined): JSDocCallbackTag;
        createJSDocOverloadTag(tagName: Identifier | undefined, typeExpression: JSDocSignature, comment?: string | NodeArray<JSDocComment>): JSDocOverloadTag;
        updateJSDocOverloadTag(node: JSDocOverloadTag, tagName: Identifier | undefined, typeExpression: JSDocSignature, comment: string | NodeArray<JSDocComment> | undefined): JSDocOverloadTag;
        createJSDocAugmentsTag(tagName: Identifier | undefined, className: JSDocAugmentsTag["class"], comment?: string | NodeArray<JSDocComment>): JSDocAugmentsTag;
        updateJSDocAugmentsTag(node: JSDocAugmentsTag, tagName: Identifier | undefined, className: JSDocAugmentsTag["class"], comment: string | NodeArray<JSDocComment> | undefined): JSDocAugmentsTag;
        createJSDocImplementsTag(tagName: Identifier | undefined, className: JSDocImplementsTag["class"], comment?: string | NodeArray<JSDocComment>): JSDocImplementsTag;
        updateJSDocImplementsTag(node: JSDocImplementsTag, tagName: Identifier | undefined, className: JSDocImplementsTag["class"], comment: string | NodeArray<JSDocComment> | undefined): JSDocImplementsTag;
        createJSDocAuthorTag(tagName: Identifier | undefined, comment?: string | NodeArray<JSDocComment>): JSDocAuthorTag;
        updateJSDocAuthorTag(node: JSDocAuthorTag, tagName: Identifier | undefined, comment: string | NodeArray<JSDocComment> | undefined): JSDocAuthorTag;
        createJSDocClassTag(tagName: Identifier | undefined, comment?: string | NodeArray<JSDocComment>): JSDocClassTag;
        updateJSDocClassTag(node: JSDocClassTag, tagName: Identifier | undefined, comment: string | NodeArray<JSDocComment> | undefined): JSDocClassTag;
        createJSDocPublicTag(tagName: Identifier | undefined, comment?: string | NodeArray<JSDocComment>): JSDocPublicTag;
        updateJSDocPublicTag(node: JSDocPublicTag, tagName: Identifier | undefined, comment: string | NodeArray<JSDocComment> | undefined): JSDocPublicTag;
        createJSDocPrivateTag(tagName: Identifier | undefined, comment?: string | NodeArray<JSDocComment>): JSDocPrivateTag;
        updateJSDocPrivateTag(node: JSDocPrivateTag, tagName: Identifier | undefined, comment: string | NodeArray<JSDocComment> | undefined): JSDocPrivateTag;
        createJSDocProtectedTag(tagName: Identifier | undefined, comment?: string | NodeArray<JSDocComment>): JSDocProtectedTag;
        updateJSDocProtectedTag(node: JSDocProtectedTag, tagName: Identifier | undefined, comment: string | NodeArray<JSDocComment> | undefined): JSDocProtectedTag;
        createJSDocReadonlyTag(tagName: Identifier | undefined, comment?: string | NodeArray<JSDocComment>): JSDocReadonlyTag;
        updateJSDocReadonlyTag(node: JSDocReadonlyTag, tagName: Identifier | undefined, comment: string | NodeArray<JSDocComment> | undefined): JSDocReadonlyTag;
        createJSDocUnknownTag(tagName: Identifier, comment?: string | NodeArray<JSDocComment>): JSDocUnknownTag;
        updateJSDocUnknownTag(node: JSDocUnknownTag, tagName: Identifier, comment: string | NodeArray<JSDocComment> | undefined): JSDocUnknownTag;
        createJSDocDeprecatedTag(tagName: Identifier | undefined, comment?: string | NodeArray<JSDocComment>): JSDocDeprecatedTag;
        updateJSDocDeprecatedTag(node: JSDocDeprecatedTag, tagName: Identifier | undefined, comment?: string | NodeArray<JSDocComment>): JSDocDeprecatedTag;
        createJSDocOverrideTag(tagName: Identifier | undefined, comment?: string | NodeArray<JSDocComment>): JSDocOverrideTag;
        updateJSDocOverrideTag(node: JSDocOverrideTag, tagName: Identifier | undefined, comment?: string | NodeArray<JSDocComment>): JSDocOverrideTag;
        createJSDocThrowsTag(tagName: Identifier, typeExpression: JSDocTypeExpression | undefined, comment?: string | NodeArray<JSDocComment>): JSDocThrowsTag;
        updateJSDocThrowsTag(node: JSDocThrowsTag, tagName: Identifier | undefined, typeExpression: JSDocTypeExpression | undefined, comment?: string | NodeArray<JSDocComment> | undefined): JSDocThrowsTag;
        createJSDocSatisfiesTag(tagName: Identifier | undefined, typeExpression: JSDocTypeExpression, comment?: string | NodeArray<JSDocComment>): JSDocSatisfiesTag;
        updateJSDocSatisfiesTag(node: JSDocSatisfiesTag, tagName: Identifier | undefined, typeExpression: JSDocTypeExpression, comment: string | NodeArray<JSDocComment> | undefined): JSDocSatisfiesTag;
        createJSDocImportTag(tagName: Identifier | undefined, importClause: ImportClause | undefined, moduleSpecifier: Expression, attributes?: ImportAttributes, comment?: string | NodeArray<JSDocComment>): JSDocImportTag;
        updateJSDocImportTag(node: JSDocImportTag, tagName: Identifier | undefined, importClause: ImportClause | undefined, moduleSpecifier: Expression, attributes: ImportAttributes | undefined, comment: string | NodeArray<JSDocComment> | undefined): JSDocImportTag;
        createJSDocText(text: string): JSDocText;
        updateJSDocText(node: JSDocText, text: string): JSDocText;
        createJSDocComment(comment?: string | NodeArray<JSDocComment> | undefined, tags?: readonly JSDocTag[] | undefined): JSDoc;
        updateJSDocComment(node: JSDoc, comment: string | NodeArray<JSDocComment> | undefined, tags: readonly JSDocTag[] | undefined): JSDoc;
        createJsxElement(openingElement: JsxOpeningElement, children: readonly JsxChild[], closingElement: JsxClosingElement): JsxElement;
        updateJsxElement(node: JsxElement, openingElement: JsxOpeningElement, children: readonly JsxChild[], closingElement: JsxClosingElement): JsxElement;
        createJsxSelfClosingElement(tagName: JsxTagNameExpression, typeArguments: readonly TypeNode[] | undefined, attributes: JsxAttributes): JsxSelfClosingElement;
        updateJsxSelfClosingElement(node: JsxSelfClosingElement, tagName: JsxTagNameExpression, typeArguments: readonly TypeNode[] | undefined, attributes: JsxAttributes): JsxSelfClosingElement;
        createJsxOpeningElement(tagName: JsxTagNameExpression, typeArguments: readonly TypeNode[] | undefined, attributes: JsxAttributes): JsxOpeningElement;
        updateJsxOpeningElement(node: JsxOpeningElement, tagName: JsxTagNameExpression, typeArguments: readonly TypeNode[] | undefined, attributes: JsxAttributes): JsxOpeningElement;
        createJsxClosingElement(tagName: JsxTagNameExpression): JsxClosingElement;
        updateJsxClosingElement(node: JsxClosingElement, tagName: JsxTagNameExpression): JsxClosingElement;
        createJsxFragment(openingFragment: JsxOpeningFragment, children: readonly JsxChild[], closingFragment: JsxClosingFragment): JsxFragment;
        createJsxText(text: string, containsOnlyTriviaWhiteSpaces?: boolean): JsxText;
        updateJsxText(node: JsxText, text: string, containsOnlyTriviaWhiteSpaces?: boolean): JsxText;
        createJsxOpeningFragment(): JsxOpeningFragment;
        createJsxJsxClosingFragment(): JsxClosingFragment;
        updateJsxFragment(node: JsxFragment, openingFragment: JsxOpeningFragment, children: readonly JsxChild[], closingFragment: JsxClosingFragment): JsxFragment;
        createJsxAttribute(name: JsxAttributeName, initializer: JsxAttributeValue | undefined): JsxAttribute;
        updateJsxAttribute(node: JsxAttribute, name: JsxAttributeName, initializer: JsxAttributeValue | undefined): JsxAttribute;
        createJsxAttributes(properties: readonly JsxAttributeLike[]): JsxAttributes;
        updateJsxAttributes(node: JsxAttributes, properties: readonly JsxAttributeLike[]): JsxAttributes;
        createJsxSpreadAttribute(expression: Expression): JsxSpreadAttribute;
        updateJsxSpreadAttribute(node: JsxSpreadAttribute, expression: Expression): JsxSpreadAttribute;
        createJsxExpression(dotDotDotToken: DotDotDotToken | undefined, expression: Expression | undefined): JsxExpression;
        updateJsxExpression(node: JsxExpression, expression: Expression | undefined): JsxExpression;
        createJsxNamespacedName(namespace: Identifier, name: Identifier): JsxNamespacedName;
        updateJsxNamespacedName(node: JsxNamespacedName, namespace: Identifier, name: Identifier): JsxNamespacedName;
        createCaseClause(expression: Expression, statements: readonly Statement[]): CaseClause;
        updateCaseClause(node: CaseClause, expression: Expression, statements: readonly Statement[]): CaseClause;
        createDefaultClause(statements: readonly Statement[]): DefaultClause;
        updateDefaultClause(node: DefaultClause, statements: readonly Statement[]): DefaultClause;
        createHeritageClause(token: HeritageClause["token"], types: readonly ExpressionWithTypeArguments[]): HeritageClause;
        updateHeritageClause(node: HeritageClause, types: readonly ExpressionWithTypeArguments[]): HeritageClause;
        createCatchClause(variableDeclaration: string | BindingName | VariableDeclaration | undefined, block: Block): CatchClause;
        updateCatchClause(node: CatchClause, variableDeclaration: VariableDeclaration | undefined, block: Block): CatchClause;
        createPropertyAssignment(name: string | PropertyName, initializer: Expression): PropertyAssignment;
        updatePropertyAssignment(node: PropertyAssignment, name: PropertyName, initializer: Expression): PropertyAssignment;
        createShorthandPropertyAssignment(name: string | Identifier, objectAssignmentInitializer?: Expression): ShorthandPropertyAssignment;
        updateShorthandPropertyAssignment(node: ShorthandPropertyAssignment, name: Identifier, objectAssignmentInitializer: Expression | undefined): ShorthandPropertyAssignment;
        createSpreadAssignment(expression: Expression): SpreadAssignment;
        updateSpreadAssignment(node: SpreadAssignment, expression: Expression): SpreadAssignment;
        createEnumMember(name: string | PropertyName, initializer?: Expression): EnumMember;
        updateEnumMember(node: EnumMember, name: PropertyName, initializer: Expression | undefined): EnumMember;
        createSourceFile(statements: readonly Statement[], endOfFileToken: EndOfFileToken, flags: NodeFlags): SourceFile;
        updateSourceFile(node: SourceFile, statements: readonly Statement[], isDeclarationFile?: boolean, referencedFiles?: readonly FileReference[], typeReferences?: readonly FileReference[], hasNoDefaultLib?: boolean, libReferences?: readonly FileReference[]): SourceFile;
        createNotEmittedStatement(original: Node): NotEmittedStatement;
        createNotEmittedTypeElement(): NotEmittedTypeElement;
        createPartiallyEmittedExpression(expression: Expression, original?: Node): PartiallyEmittedExpression;
        updatePartiallyEmittedExpression(node: PartiallyEmittedExpression, expression: Expression): PartiallyEmittedExpression;
        createCommaListExpression(elements: readonly Expression[]): CommaListExpression;
        updateCommaListExpression(node: CommaListExpression, elements: readonly Expression[]): CommaListExpression;
        createBundle(sourceFiles: readonly SourceFile[]): Bundle;
        updateBundle(node: Bundle, sourceFiles: readonly SourceFile[]): Bundle;
        createComma(left: Expression, right: Expression): BinaryExpression;
        createAssignment(left: ObjectLiteralExpression | ArrayLiteralExpression, right: Expression): DestructuringAssignment;
        createAssignment(left: Expression, right: Expression): AssignmentExpression<EqualsToken>;
        createLogicalOr(left: Expression, right: Expression): BinaryExpression;
        createLogicalAnd(left: Expression, right: Expression): BinaryExpression;
        createBitwiseOr(left: Expression, right: Expression): BinaryExpression;
        createBitwiseXor(left: Expression, right: Expression): BinaryExpression;
        createBitwiseAnd(left: Expression, right: Expression): BinaryExpression;
        createStrictEquality(left: Expression, right: Expression): BinaryExpression;
        createStrictInequality(left: Expression, right: Expression): BinaryExpression;
        createEquality(left: Expression, right: Expression): BinaryExpression;
        createInequality(left: Expression, right: Expression): BinaryExpression;
        createLessThan(left: Expression, right: Expression): BinaryExpression;
        createLessThanEquals(left: Expression, right: Expression): BinaryExpression;
        createGreaterThan(left: Expression, right: Expression): BinaryExpression;
        createGreaterThanEquals(left: Expression, right: Expression): BinaryExpression;
        createLeftShift(left: Expression, right: Expression): BinaryExpression;
        createRightShift(left: Expression, right: Expression): BinaryExpression;
        createUnsignedRightShift(left: Expression, right: Expression): BinaryExpression;
        createAdd(left: Expression, right: Expression): BinaryExpression;
        createSubtract(left: Expression, right: Expression): BinaryExpression;
        createMultiply(left: Expression, right: Expression): BinaryExpression;
        createDivide(left: Expression, right: Expression): BinaryExpression;
        createModulo(left: Expression, right: Expression): BinaryExpression;
        createExponent(left: Expression, right: Expression): BinaryExpression;
        createPrefixPlus(operand: Expression): PrefixUnaryExpression;
        createPrefixMinus(operand: Expression): PrefixUnaryExpression;
        createPrefixIncrement(operand: Expression): PrefixUnaryExpression;
        createPrefixDecrement(operand: Expression): PrefixUnaryExpression;
        createBitwiseNot(operand: Expression): PrefixUnaryExpression;
        createLogicalNot(operand: Expression): PrefixUnaryExpression;
        createPostfixIncrement(operand: Expression): PostfixUnaryExpression;
        createPostfixDecrement(operand: Expression): PostfixUnaryExpression;
        createImmediatelyInvokedFunctionExpression(statements: readonly Statement[]): CallExpression;
        createImmediatelyInvokedFunctionExpression(statements: readonly Statement[], param: ParameterDeclaration, paramValue: Expression): CallExpression;
        createImmediatelyInvokedArrowFunction(statements: readonly Statement[]): ImmediatelyInvokedArrowFunction;
        createImmediatelyInvokedArrowFunction(statements: readonly Statement[], param: ParameterDeclaration, paramValue: Expression): ImmediatelyInvokedArrowFunction;
        createVoidZero(): VoidExpression;
        createExportDefault(expression: Expression): ExportAssignment;
        createExternalModuleExport(exportName: Identifier): ExportDeclaration;
        restoreOuterExpressions(outerExpression: Expression | undefined, innerExpression: Expression, kinds?: OuterExpressionKinds): Expression;
        /**
         * Updates a node that may contain modifiers, replacing only the modifiers of the node.
         */
        replaceModifiers<T extends HasModifiers>(node: T, modifiers: readonly Modifier[] | ModifierFlags | undefined): T;
        /**
         * Updates a node that may contain decorators or modifiers, replacing only the decorators and modifiers of the node.
         */
        replaceDecoratorsAndModifiers<T extends HasModifiers & HasDecorators>(node: T, modifiers: readonly ModifierLike[] | undefined): T;
        /**
         * Updates a node that contains a property name, replacing only the name of the node.
         */
        replacePropertyName<T extends AccessorDeclaration | MethodDeclaration | MethodSignature | PropertyDeclaration | PropertySignature | PropertyAssignment>(node: T, name: T["name"]): T;
    }
```

### TransformationContext

```typescript
interface TransformationContext extends CoreTransformationContext {
        /** Records a request for a non-scoped emit helper in the current context. */
        requestEmitHelper(helper: EmitHelper): void;
        /** Gets and resets the requested non-scoped emit helpers. */
        readEmitHelpers(): EmitHelper[] | undefined;
        /** Enables expression substitutions in the pretty printer for the provided SyntaxKind. */
        enableSubstitution(kind: SyntaxKind): void;
        /** Determines whether expression substitutions are enabled for the provided node. */
        isSubstitutionEnabled(node: Node): boolean;
        /**
         * Hook used by transformers to substitute expressions just before they
         * are emitted by the pretty printer.
         *
         * NOTE: Transformation hooks should only be modified during `Transformer` initialization,
         * before returning the `NodeTransformer` callback.
         */
        onSubstituteNode: (hint: EmitHint, node: Node) => Node;
        /**
         * Enables before/after emit notifications in the pretty printer for the provided
         * SyntaxKind.
         */
        enableEmitNotification(kind: SyntaxKind): void;
        /**
         * Determines whether before/after emit notifications should be raised in the pretty
         * printer when it emits a node.
         */
        isEmitNotificationEnabled(node: Node): boolean;
        /**
         * Hook used to allow transformers to capture state before or after
         * the printer emits a node.
         *
         * NOTE: Transformation hooks should only be modified during `Transformer` initialization,
         * before returning the `NodeTransformer` callback.
         */
        onEmitNode: (hint: EmitHint, node: Node, emitCallback: (hint: EmitHint, node: Node) => void) => void;
    }
```

## Outstanding

Census completion/reconciliation; exhaustive implementation and this-use attribution; corrected whole-lane counters including destructuring defaults; exact public property declaration diff; stock strictFunctionTypes diagnostics for both roads across 79 roots. The production-path positive control passed, but the default-path control failed. No adaptation or compiler change is committed.

## Provenance

Adamic `8e3a784aa5a601f57342c07b68f145fe56252c21`, fresh detached scratch worktree. Upstream `v6.0.3` at `050880ce59e30b356b686bd3144efe24f875ebc8`. Recursive shallow submodules initialized. Node v24.19.0 and Go go1.27.1. All writes, npm caches, Go caches and temporary files are in task scratch. Baseline adapted build: `stage3/apply.sh --build-product /tmp/destructure-price-scratch/adapted`. Shared checkout untouched. The measurement scripts currently retain absolute task paths for exact reproducibility in this worker.

## Evidence publication boundary

Automatic approval review rejected a push containing raw logs, process counters and source-derived provider evidence, citing sensitive payloads and an unverified remote. This final update publishes only the report. Raw measurements remain in private scratch; no claim depends on treating incomplete logs as complete evidence. The earlier pushed AST inventory and public declaration quotations remain available on this branch.
