interface Base { readonly kind: 'jsdoc-function' | 'method' | 'function' | 'type-alias' | 'interface' | 'module' | 'command' | 'source' | 'json-source' | 'json-statement' | 'object-literal' | 'template-literal-type' | 'template-literal-type-span' | 'template-middle' | 'template-tail' | 'type-literal' | 'property-signature' | 'method-signature' | 'case' | 'variable' | 'return' | 'identifier' | 'string' | 'keyword' | 'type-parameter' | 'parameter' | 'export' | 'declare' | 'const' | 'in' | 'out' | 'decorator' | 'type-reference' | 'jsdoc' | 'jsdoc-text' | 'jsdoc-link'; }
interface Range { readonly pos: number; readonly end: number; }
interface NodeArray<T> extends ReadonlyArray<T>, Range { readonly hasTrailingComma: boolean; }
interface Identifier extends Base { readonly kind: 'identifier'; readonly text: string; }
interface StringLiteral extends Base { readonly kind: 'string'; readonly text: string; }
interface KeywordTypeNode extends Base { readonly kind: 'keyword'; readonly text: string; }
interface ModifierToken extends Base { readonly kind: 'export' | 'declare' | 'const' | 'in' | 'out'; }
interface Decorator extends Base { readonly kind: 'decorator'; readonly expression: Identifier; }
type Modifier = ModifierToken;
type ModifierLike = Modifier | Decorator;
interface TypeParameterDeclaration extends Base { readonly kind: 'type-parameter'; readonly modifiers?: NodeArray<Modifier>; readonly name: Identifier; readonly constraint?: KeywordTypeNode; }
interface ParameterDeclaration extends Base { readonly kind: 'parameter'; readonly name: Identifier; }
interface JSDocFunctionType extends Base { readonly kind: 'jsdoc-function'; readonly parameters: NodeArray<ParameterDeclaration>; }
interface MethodDeclaration extends Base { readonly kind: 'method'; readonly name: Identifier; readonly typeParameters?: NodeArray<TypeParameterDeclaration>; readonly parameters: NodeArray<ParameterDeclaration>; }
interface FunctionDeclaration extends Base { readonly kind: 'function'; readonly name?: Identifier; readonly typeParameters?: NodeArray<TypeParameterDeclaration>; readonly parameters: NodeArray<ParameterDeclaration>; }
interface TypeAliasDeclaration extends Base { readonly kind: 'type-alias'; readonly name: Identifier; readonly typeParameters?: NodeArray<TypeParameterDeclaration>; readonly type: KeywordTypeNode; }
type HasType = JSDocFunctionType | MethodDeclaration | FunctionDeclaration | TypeAliasDeclaration;
interface PropertySignature extends Base { readonly kind: 'property-signature'; readonly name: Identifier; }
interface MethodSignature extends Base { readonly kind: 'method-signature'; readonly name: Identifier; }
type TypeElement = PropertySignature | MethodSignature;
interface InterfaceDeclaration extends Base { readonly kind: 'interface'; readonly name: Identifier; readonly typeParameters?: NodeArray<TypeParameterDeclaration>; readonly members: NodeArray<TypeElement>; }
interface TypeLiteralNode extends Base { readonly kind: 'type-literal'; readonly members: NodeArray<TypeElement>; }
type ObjectTypeDeclaration = InterfaceDeclaration | TypeLiteralNode;
interface ModuleDeclaration extends Base { readonly kind: 'module'; readonly modifiers?: NodeArray<ModifierLike>; readonly name: Identifier | StringLiteral; }
interface Diagnostic { category: number; code: number; start: number | undefined; length: number | undefined; messageText: string; }
interface ParsedCommandLine extends Base { readonly kind: 'command'; fileNames: string[]; errors: Diagnostic[]; }
interface SourceFile extends Base { readonly kind: 'source'; fileName: string; moduleAugmentations: readonly (StringLiteral | Identifier)[]; packageJsonLocations?: readonly string[]; }
interface ObjectLiteralExpression extends Base { readonly kind: 'object-literal'; readonly pos: number; }
interface JsonObjectExpressionStatement extends Base { readonly kind: 'json-statement'; readonly expression: ObjectLiteralExpression; }
interface JsonSourceFile extends Base { readonly kind: 'json-source'; readonly statements: NodeArray<JsonObjectExpressionStatement>; }
interface TemplateMiddle extends Base { readonly kind: 'template-middle'; readonly text: string; }
interface TemplateTail extends Base { readonly kind: 'template-tail'; readonly text: string; }
interface TemplateLiteralTypeSpan extends Base { readonly kind: 'template-literal-type-span'; readonly type: KeywordTypeNode; readonly literal: TemplateMiddle | TemplateTail; }
interface TemplateLiteralTypeNode extends Base { readonly kind: 'template-literal-type'; readonly head: string; readonly templateSpans: NodeArray<TemplateLiteralTypeSpan>; }
interface VariableStatement extends Base { readonly kind: 'variable'; readonly name: Identifier; }
interface ReturnStatement extends Base { readonly kind: 'return'; }
type Statement = VariableStatement | ReturnStatement;
interface CaseClause extends Base { readonly kind: 'case'; readonly expression: Identifier; readonly statements: NodeArray<Statement>; }
interface NodeWithTypeArguments extends Base { readonly kind: 'type-reference'; readonly typeName: Identifier; readonly typeArguments?: NodeArray<KeywordTypeNode>; }
interface JSDocText extends Base { readonly kind: 'jsdoc-text'; text: string; }
interface JSDocLink extends Base { readonly kind: 'jsdoc-link'; readonly name?: Identifier; text: string; }
type JSDocComment = JSDocText | JSDocLink;
interface JSDoc extends Base { readonly kind: 'jsdoc'; readonly comment?: string | NodeArray<JSDocComment>; }
function jsdocFunction(node: Base): JSDocFunctionType { return node as JSDocFunctionType; }
function method(node: Base): MethodDeclaration { return node as MethodDeclaration; }
function functionDeclaration(node: Base): FunctionDeclaration { return node as FunctionDeclaration; }
function typeAlias(node: Base): TypeAliasDeclaration { return node as TypeAliasDeclaration; }
function interfaceDeclaration(node: Base): InterfaceDeclaration { return node as InterfaceDeclaration; }
function typeLiteral(node: Base): TypeLiteralNode { return node as TypeLiteralNode; }
function moduleDeclaration(node: Base): ModuleDeclaration { return node as ModuleDeclaration; }
function commandLine(value: Base): ParsedCommandLine { return value as ParsedCommandLine; }
function sourceFile(node: Base): SourceFile { return node as SourceFile; }
function jsonSourceFile(node: Base): JsonSourceFile { return node as JsonSourceFile; }
function templateLiteralType(node: Base): TemplateLiteralTypeNode { return node as TemplateLiteralTypeNode; }
function caseClause(node: Base): CaseClause { return node as CaseClause; }
function nodeWithTypeArguments(node: Base): NodeWithTypeArguments { return node as NodeWithTypeArguments; }
function typeParameter(node: Base): TypeParameterDeclaration { return node as TypeParameterDeclaration; }
function jsDoc(node: Base): JSDoc { return node as JSDoc; }
function typeParameterNames(node: HasType | InterfaceDeclaration): string { return node.kind === 'jsdoc-function' || node.typeParameters === undefined ? 'none' : node.typeParameters.map(item => item.name.text).join(','); }
function memberNames(node: ObjectTypeDeclaration): string { return node.members.map(item => `${item.kind === 'method-signature' ? 'm' : 'p'}${item.name.text}`).join(','); }
const raw = {kind: 'type-reference' as const, typeName: {kind: 'identifier' as const, text: 'Map'}, typeArguments: Object.assign([{kind: 'identifier' as const, text: 'K'}], {pos: 0, end: 1, hasTrailingComma: false})};
const typeArguments = nodeWithTypeArguments(raw).typeArguments!;
console.log(typeArguments[0]!.text);
console.log(typeArguments[0]!.kind);
