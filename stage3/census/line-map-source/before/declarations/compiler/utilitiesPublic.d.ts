import { type __String, type AccessorDeclaration, type ArrayBindingElement, type AssertionExpression, type AutoAccessorPropertyDeclaration, type BindingElement, type BindingName, type BindingPattern, type BooleanLiteral, type BreakOrContinueStatement, type CallChain, type CallExpression, type CallLikeExpression, type CaseOrDefaultClause, type ClassElement, type ClassLikeDeclaration, type CompilerOptions, type ConciseBody, type ConstructorDeclaration, type ConstructorTypeNode, type Declaration, type DeclarationName, type DeclarationStatement, type DeclarationWithTypeParameters, type Decorator, type Diagnostic, type ElementAccessChain, type EntityName, type ExportSpecifier, type Expression, type FileReference, type ForInitializer, type FunctionLikeDeclaration, type FunctionTypeNode, type GetAccessorDeclaration, type HasDecorators, type HasExpressionInitializer, type HasModifiers, type Identifier, type ImportAttributeName, type ImportSpecifier, type IterationStatement, type JSDocAugmentsTag, type JSDocClassTag, type JSDocComment, type JSDocDeprecatedTag, type JSDocEnumTag, type JSDocImplementsTag, type JSDocLink, type JSDocLinkCode, type JSDocLinkPlain, type JSDocOverrideTag, type JSDocParameterTag, type JSDocPrivateTag, type JSDocPropertyLikeTag, type JSDocProtectedTag, type JSDocPublicTag, type JSDocReadonlyTag, type JSDocReturnTag, type JSDocSatisfiesTag, type JSDocSignature, type JSDocTag, type JSDocTemplateTag, type JSDocThisTag, type JSDocTypedefTag, type JSDocTypeTag, type JsxAttributeLike, type JsxCallLike, type JsxChild, type JsxExpression, type JsxOpeningLikeElement, type JsxTagNameExpression, type KeywordSyntaxKind, type LabeledStatement, type LeftHandSideExpression, type LiteralExpression, type MemberName, type Modifier, ModifierFlags, type ModifierLike, type ModuleBody, type ModuleReference, type NamedExportBindings, type NamedImportBindings, type NewExpression, type Node, type NodeArray, NodeFlags, type NonNullChain, type NullLiteral, type ObjectLiteralElement, type ObjectLiteralElementLike, type ParameterDeclaration, type PrefixUnaryExpression, type PrivateIdentifier, type PropertyAccessChain, type PropertyAccessExpression, type PropertyName, type QualifiedName, type SetAccessorDeclaration, type SignatureDeclaration, type SortedReadonlyArray, type SourceFile, type Statement, type StringLiteral, type StringLiteralLike, type Symbol, SyntaxKind, type TemplateLiteral, type TemplateLiteralToken, type TemplateMiddle, type TemplateTail, type TextChangeRange, type TextSpan, type TypeElement, type TypeNode, type TypeOnlyAliasDeclaration, type TypeOnlyExportDeclaration, type TypeOnlyImportDeclaration, type TypeParameterDeclaration, type VariableDeclaration } from "./_namespaces/ts.js";
export declare function isExternalModuleNameRelative(moduleName: string): boolean;
export declare function sortAndDeduplicateDiagnostics<T extends Diagnostic>(diagnostics: readonly T[]): SortedReadonlyArray<T>;
export declare function getDefaultLibFileName(options: CompilerOptions): string;
export declare function textSpanEnd(span: TextSpan): number;
export declare function textSpanIsEmpty(span: TextSpan): boolean;
export declare function textSpanContainsPosition(span: TextSpan, position: number): boolean;
export declare function textSpanContainsTextSpan(span: TextSpan, other: TextSpan): boolean;
export declare function textSpanOverlapsWith(span: TextSpan, other: TextSpan): boolean;
export declare function textSpanOverlap(span1: TextSpan, span2: TextSpan): TextSpan | undefined;
export declare function textSpanIntersectsWithTextSpan(span: TextSpan, other: TextSpan): boolean;
export declare function textSpanIntersectsWith(span: TextSpan, start: number, length: number): boolean;
export declare function decodedTextSpanIntersectsWith(start1: number, length1: number, start2: number, length2: number): boolean;
export declare function textSpanIntersectsWithPosition(span: TextSpan, position: number): boolean;
export declare function textSpanIntersection(span1: TextSpan, span2: TextSpan): TextSpan | undefined;
export declare function createTextSpan(start: number, length: number): TextSpan;
export declare function createTextSpanFromBounds(start: number, end: number): TextSpan;
export declare function textChangeRangeNewSpan(range: TextChangeRange): TextSpan;
export declare function textChangeRangeIsUnchanged(range: TextChangeRange): boolean;
export declare function createTextChangeRange(span: TextSpan, newLength: number): TextChangeRange;
export declare const unchangedTextChangeRange: TextChangeRange;
/**
 * Called to merge all the changes that occurred across several versions of a script snapshot
 * into a single change.  i.e. if a user keeps making successive edits to a script we will
 * have a text change from V1 to V2, V2 to V3, ..., Vn.
 *
 * This function will then merge those changes into a single change range valid between V1 and
 * Vn.
 */
export declare function collapseTextChangeRangesAcrossMultipleVersions(changes: readonly TextChangeRange[]): TextChangeRange;
export declare function getTypeParameterOwner(d: Declaration): Declaration | undefined;
export type ParameterPropertyDeclaration = ParameterDeclaration & {
    parent: ConstructorDeclaration;
    name: Identifier;
};
export declare function isParameterPropertyDeclaration(node: Node, parent: Node): node is ParameterPropertyDeclaration;
export declare function isEmptyBindingPattern(node: BindingName): node is BindingPattern;
export declare function isEmptyBindingElement(node: BindingElement | ArrayBindingElement): boolean;
export declare function walkUpBindingElementsAndPatterns(binding: BindingElement): VariableDeclaration | ParameterDeclaration;
export declare function getCombinedModifierFlags(node: Declaration): ModifierFlags;
export declare function getCombinedNodeFlags(node: Node): NodeFlags;
/**
 * Checks to see if the locale is in the appropriate format,
 * and if it is, attempts to set the appropriate language.
 */
export declare function validateLocaleAndSetLanguage(locale: string, sys: {
    getExecutingFilePath(): string;
    resolvePath(path: string): string;
    fileExists(fileName: string): boolean;
    readFile(fileName: string): string | undefined;
}, errors?: Diagnostic[]): void;
export declare function getOriginalNode(node: Node): Node;
export declare function getOriginalNode<T extends Node>(node: Node, nodeTest: (node: Node) => node is T): T;
export declare function getOriginalNode(node: Node | undefined): Node | undefined;
export declare function getOriginalNode<T extends Node>(node: Node | undefined, nodeTest: (node: Node) => node is T): T | undefined;
/**
 * Iterates through the parent chain of a node and performs the callback on each parent until the callback
 * returns a truthy value, then returns that value.
 * If no such value is found, it applies the callback until the parent pointer is undefined or the callback returns "quit"
 * At that point findAncestor returns undefined.
 */
export declare function findAncestor<T extends Node>(node: Node | undefined, callback: (element: Node) => element is T): T | undefined;
export declare function findAncestor(node: Node | undefined, callback: (element: Node) => boolean | "quit"): Node | undefined;
/**
 * Gets a value indicating whether a node originated in the parse tree.
 *
 * @param node The node to test.
 */
export declare function isParseTreeNode(node: Node): boolean;
/**
 * Gets the original parse tree node for a node.
 *
 * @param node The original node.
 * @returns The original parse tree node if found; otherwise, undefined.
 */
export declare function getParseTreeNode(node: Node | undefined): Node | undefined;
/**
 * Gets the original parse tree node for a node.
 *
 * @param node The original node.
 * @param nodeTest A callback used to ensure the correct type of parse tree node is returned.
 * @returns The original parse tree node if found; otherwise, undefined.
 */
export declare function getParseTreeNode<T extends Node>(node: T | undefined, nodeTest?: (node: Node) => node is T): T | undefined;
/** Add an extra underscore to identifiers that start with two underscores to avoid issues with magic names like '__proto__' */
export declare function escapeLeadingUnderscores(identifier: string): __String;
/**
 * Remove extra underscore from escaped identifier text content.
 *
 * @param identifier The escaped identifier text.
 * @returns The unescaped identifier text.
 */
export declare function unescapeLeadingUnderscores(identifier: __String): string;
export declare function idText(identifierOrPrivateName: Identifier | PrivateIdentifier): string;
/**
 * If the text of an Identifier matches a keyword (including contextual and TypeScript-specific keywords), returns the
 * SyntaxKind for the matching keyword.
 */
export declare function identifierToKeywordKind(node: Identifier): KeywordSyntaxKind | undefined;
export declare function symbolName(symbol: Symbol): string;
export declare function getNameOfJSDocTypedef(declaration: JSDocTypedefTag): Identifier | PrivateIdentifier | undefined;
export declare function getNameOfDeclaration(declaration: Declaration | Expression | undefined): DeclarationName | undefined;
export declare function getDecorators(node: HasDecorators): readonly Decorator[] | undefined;
export declare function getModifiers(node: HasModifiers): readonly Modifier[] | undefined;
/**
 * Gets the JSDoc parameter tags for the node if present.
 *
 * @remarks Returns any JSDoc param tag whose name matches the provided
 * parameter, whether a param tag on a containing function
 * expression, or a param tag on a variable declaration whose
 * initializer is the containing function. The tags closest to the
 * node are returned first, so in the previous example, the param
 * tag on the containing function expression would be first.
 *
 * For binding patterns, parameter tags are matched by position.
 */
export declare function getJSDocParameterTags(param: ParameterDeclaration): readonly JSDocParameterTag[];
/**
 * Gets the JSDoc type parameter tags for the node if present.
 *
 * @remarks Returns any JSDoc template tag whose names match the provided
 * parameter, whether a template tag on a containing function
 * expression, or a template tag on a variable declaration whose
 * initializer is the containing function. The tags closest to the
 * node are returned first, so in the previous example, the template
 * tag on the containing function expression would be first.
 */
export declare function getJSDocTypeParameterTags(param: TypeParameterDeclaration): readonly JSDocTemplateTag[];
/**
 * Return true if the node has JSDoc parameter tags.
 *
 * @remarks Includes parameter tags that are not directly on the node,
 * for example on a variable declaration whose initializer is a function expression.
 */
export declare function hasJSDocParameterTags(node: FunctionLikeDeclaration | SignatureDeclaration): boolean;
/** Gets the JSDoc augments tag for the node if present */
export declare function getJSDocAugmentsTag(node: Node): JSDocAugmentsTag | undefined;
/** Gets the JSDoc implements tags for the node if present */
export declare function getJSDocImplementsTags(node: Node): readonly JSDocImplementsTag[];
/** Gets the JSDoc class tag for the node if present */
export declare function getJSDocClassTag(node: Node): JSDocClassTag | undefined;
/** Gets the JSDoc public tag for the node if present */
export declare function getJSDocPublicTag(node: Node): JSDocPublicTag | undefined;
/** Gets the JSDoc private tag for the node if present */
export declare function getJSDocPrivateTag(node: Node): JSDocPrivateTag | undefined;
/** Gets the JSDoc protected tag for the node if present */
export declare function getJSDocProtectedTag(node: Node): JSDocProtectedTag | undefined;
/** Gets the JSDoc protected tag for the node if present */
export declare function getJSDocReadonlyTag(node: Node): JSDocReadonlyTag | undefined;
export declare function getJSDocOverrideTagNoCache(node: Node): JSDocOverrideTag | undefined;
/** Gets the JSDoc deprecated tag for the node if present */
export declare function getJSDocDeprecatedTag(node: Node): JSDocDeprecatedTag | undefined;
/** Gets the JSDoc enum tag for the node if present */
export declare function getJSDocEnumTag(node: Node): JSDocEnumTag | undefined;
/** Gets the JSDoc this tag for the node if present */
export declare function getJSDocThisTag(node: Node): JSDocThisTag | undefined;
/** Gets the JSDoc return tag for the node if present */
export declare function getJSDocReturnTag(node: Node): JSDocReturnTag | undefined;
/** Gets the JSDoc template tag for the node if present */
export declare function getJSDocTemplateTag(node: Node): JSDocTemplateTag | undefined;
export declare function getJSDocSatisfiesTag(node: Node): JSDocSatisfiesTag | undefined;
/** Gets the JSDoc type tag for the node if present and valid */
export declare function getJSDocTypeTag(node: Node): JSDocTypeTag | undefined;
/**
 * Gets the type node for the node if provided via JSDoc.
 *
 * @remarks The search includes any JSDoc param tag that relates
 * to the provided parameter, for example a type tag on the
 * parameter itself, or a param tag on a containing function
 * expression, or a param tag on a variable declaration whose
 * initializer is the containing function. The tags closest to the
 * node are examined first, so in the previous example, the type
 * tag directly on the node would be returned.
 */
export declare function getJSDocType(node: Node): TypeNode | undefined;
/**
 * Gets the return type node for the node if provided via JSDoc return tag or type tag.
 *
 * @remarks `getJSDocReturnTag` just gets the whole JSDoc tag. This function
 * gets the type from inside the braces, after the fat arrow, etc.
 */
export declare function getJSDocReturnType(node: Node): TypeNode | undefined;
/** Get all JSDoc tags related to a node, including those on parent nodes. */
export declare function getJSDocTags(node: Node): readonly JSDocTag[];
/** Gets all JSDoc tags that match a specified predicate */
export declare function getAllJSDocTags<T extends JSDocTag>(node: Node, predicate: (tag: JSDocTag) => tag is T): readonly T[];
/** Gets all JSDoc tags of a specified kind */
export declare function getAllJSDocTagsOfKind(node: Node, kind: SyntaxKind): readonly JSDocTag[];
/** Gets the text of a jsdoc comment, flattening links to their text. */
export declare function getTextOfJSDocComment(comment?: string | NodeArray<JSDocComment>): string | undefined;
/**
 * Gets the effective type parameters. If the node was parsed in a
 * JavaScript file, gets the type parameters from the `@template` tag from JSDoc.
 *
 * This does *not* return type parameters from a jsdoc reference to a generic type, eg
 *
 * type Id = <T>(x: T) => T
 * /** @type {Id} /
 * function id(x) { return x }
 */
export declare function getEffectiveTypeParameterDeclarations(node: DeclarationWithTypeParameters): readonly TypeParameterDeclaration[];
export declare function getEffectiveConstraintOfTypeParameter(node: TypeParameterDeclaration): TypeNode | undefined;
export declare function isMemberName(node: Node): node is MemberName;
export declare function isPropertyAccessChain(node: Node): node is PropertyAccessChain;
export declare function isElementAccessChain(node: Node): node is ElementAccessChain;
export declare function isCallChain(node: Node): node is CallChain;
export declare function isOptionalChain(node: Node): node is PropertyAccessChain | ElementAccessChain | CallChain | NonNullChain;
export declare function isNullishCoalesce(node: Node): boolean;
export declare function isConstTypeReference(node: Node): boolean;
export declare function skipPartiallyEmittedExpressions(node: Expression): Expression;
export declare function skipPartiallyEmittedExpressions(node: Node): Node;
export declare function isNonNullChain(node: Node): node is NonNullChain;
export declare function isBreakOrContinueStatement(node: Node): node is BreakOrContinueStatement;
export declare function isNamedExportBindings(node: Node): node is NamedExportBindings;
export declare function isJSDocPropertyLikeTag(node: Node): node is JSDocPropertyLikeTag;
/**
 * True if kind is of some token syntax kind.
 * For example, this is true for an IfKeyword but not for an IfStatement.
 * Literals are considered tokens, except TemplateLiteral, but does include TemplateHead/Middle/Tail.
 */
export declare function isTokenKind(kind: SyntaxKind): boolean;
/**
 * True if node is of some token syntax kind.
 * For example, this is true for an IfKeyword but not for an IfStatement.
 * Literals are considered tokens, except TemplateLiteral, but does include TemplateHead/Middle/Tail.
 */
export declare function isToken(n: Node): boolean;
export declare function isLiteralExpression(node: Node): node is LiteralExpression;
export declare function isTemplateLiteralToken(node: Node): node is TemplateLiteralToken;
export declare function isTemplateMiddleOrTemplateTail(node: Node): node is TemplateMiddle | TemplateTail;
export declare function isImportOrExportSpecifier(node: Node): node is ImportSpecifier | ExportSpecifier;
export declare function isTypeOnlyImportDeclaration(node: Node): node is TypeOnlyImportDeclaration;
export declare function isTypeOnlyExportDeclaration(node: Node): node is TypeOnlyExportDeclaration;
export declare function isTypeOnlyImportOrExportDeclaration(node: Node): node is TypeOnlyAliasDeclaration;
export declare function isPartOfTypeOnlyImportOrExportDeclaration(node: Node): boolean;
export declare function isStringTextContainingNode(node: Node): node is StringLiteral | TemplateLiteralToken;
export declare function isImportAttributeName(node: Node): node is ImportAttributeName;
export declare function isModifier(node: Node): node is Modifier;
export declare function isEntityName(node: Node): node is EntityName;
export declare function isPropertyName(node: Node): node is PropertyName;
export declare function isBindingName(node: Node): node is BindingName;
export declare function isFunctionLike(node: Node | undefined): node is SignatureDeclaration;
export declare function isClassElement(node: Node): node is ClassElement;
export declare function isClassLike(node: Node): node is ClassLikeDeclaration;
export declare function isAccessor(node: Node): node is AccessorDeclaration;
export declare function isAutoAccessorPropertyDeclaration(node: Node): node is AutoAccessorPropertyDeclaration;
export declare function isModifierLike(node: Node): node is ModifierLike;
export declare function isTypeElement(node: Node): node is TypeElement;
export declare function isClassOrTypeElement(node: Node): node is ClassElement | TypeElement;
export declare function isObjectLiteralElementLike(node: Node): node is ObjectLiteralElementLike;
/**
 * Node test that determines whether a node is a valid type node.
 * This differs from the `isPartOfTypeNode` function which determines whether a node is *part*
 * of a TypeNode.
 */
export declare function isTypeNode(node: Node): node is TypeNode;
export declare function isFunctionOrConstructorTypeNode(node: Node): node is FunctionTypeNode | ConstructorTypeNode;
export declare function isArrayBindingElement(node: Node): node is ArrayBindingElement;
export declare function isPropertyAccessOrQualifiedName(node: Node): node is PropertyAccessExpression | QualifiedName;
export declare function isCallLikeExpression(node: Node): node is CallLikeExpression;
export declare function isCallOrNewExpression(node: Node): node is CallExpression | NewExpression;
export declare function isTemplateLiteral(node: Node): node is TemplateLiteral;
export declare function isLeftHandSideExpression(node: Node): node is LeftHandSideExpression;
export declare function isLiteralTypeLiteral(node: Node): node is NullLiteral | BooleanLiteral | LiteralExpression | PrefixUnaryExpression;
/**
 * Determines whether a node is an expression based only on its kind.
 */
export declare function isExpression(node: Node): node is Expression;
export declare function isAssertionExpression(node: Node): node is AssertionExpression;
export declare function isIterationStatement(node: Node, lookInLabeledStatements: false): node is IterationStatement;
export declare function isIterationStatement(node: Node, lookInLabeledStatements: boolean): node is IterationStatement | LabeledStatement;
export declare function isConciseBody(node: Node): node is ConciseBody;
export declare function isForInitializer(node: Node): node is ForInitializer;
export declare function isModuleBody(node: Node): node is ModuleBody;
export declare function isNamedImportBindings(node: Node): node is NamedImportBindings;
export declare function isDeclarationStatement(node: Node): node is DeclarationStatement;
export declare function isStatement(node: Node): node is Statement;
export declare function isModuleReference(node: Node): node is ModuleReference;
export declare function isJsxTagNameExpression(node: Node): node is JsxTagNameExpression;
export declare function isJsxChild(node: Node): node is JsxChild;
export declare function isJsxAttributeLike(node: Node): node is JsxAttributeLike;
export declare function isStringLiteralOrJsxExpression(node: Node): node is StringLiteral | JsxExpression;
export declare function isJsxOpeningLikeElement(node: Node): node is JsxOpeningLikeElement;
export declare function isJsxCallLike(node: Node): node is JsxCallLike;
export declare function isCaseOrDefaultClause(node: Node): node is CaseOrDefaultClause;
/** True if node is of a kind that may contain comment text. */
export declare function isJSDocCommentContainingNode(node: Node): boolean;
export declare function isSetAccessor(node: Node): node is SetAccessorDeclaration;
export declare function isGetAccessor(node: Node): node is GetAccessorDeclaration;
/** True if has initializer node attached to it. */
export declare function hasOnlyExpressionInitializer(node: Node): node is HasExpressionInitializer;
export declare function isObjectLiteralElement(node: Node): node is ObjectLiteralElement;
export declare function isStringLiteralLike(node: Node | FileReference): node is StringLiteralLike;
export declare function isJSDocLinkLike(node: Node): node is JSDocLink | JSDocLinkCode | JSDocLinkPlain;
export declare function hasRestParameter(s: SignatureDeclaration | JSDocSignature): boolean;
export declare function isRestParameter(node: ParameterDeclaration | JSDocParameterTag): boolean;
export declare function isInternalDeclaration(node: Node, sourceFile?: SourceFile): boolean;
