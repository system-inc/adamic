const enum CommentDirectiveType { ExpectError, Ignore }
interface Base { readonly kind: 'class-expression' | 'conditional' | 'function-expression' | 'function-type' | 'import-type' | 'jsdoc-import' | 'jsdoc-typedef' | 'mapped' | 'method-signature' | 'source' | 'tagged-template' | 'type-query' | 'anonymous' | 'constructor' | 'identifier' | 'keyword' | 'type-parameter' | 'parameter' | 'property-signature' | 'async' | 'export' | 'jsdoc-text' | 'jsdoc-link' | 'tag-identifier'; }
interface Range { readonly pos: number; readonly end: number; }
interface NodeArray<T> extends ReadonlyArray<T>, Range { readonly hasTrailingComma: boolean; }
interface Identifier extends Base { readonly kind: 'identifier'; readonly text: string; }
interface KeywordTypeNode extends Base { readonly kind: 'keyword'; readonly text: string; }
interface ModifierToken extends Base { readonly kind: 'async' | 'export'; }
type Modifier = ModifierToken;
interface TypeParameterDeclaration extends Base { readonly kind: 'type-parameter'; readonly name: Identifier; }
interface ParameterDeclaration extends Base { readonly kind: 'parameter'; readonly name: Identifier; }
interface PropertySignature extends Base { readonly kind: 'property-signature'; readonly name: Identifier; }
interface MethodSignature extends Base { readonly kind: 'method-signature'; readonly modifiers?: NodeArray<Modifier>; readonly name: Identifier; readonly typeParameters?: NodeArray<TypeParameterDeclaration>; readonly parameters: NodeArray<ParameterDeclaration>; }
type TypeElement = PropertySignature | MethodSignature;
interface ClassExpression extends Base { readonly kind: 'class-expression'; readonly typeParameters?: NodeArray<TypeParameterDeclaration>; }
interface FunctionExpression extends Base { readonly kind: 'function-expression'; readonly modifiers?: NodeArray<Modifier>; readonly parameters: NodeArray<ParameterDeclaration>; }
interface FunctionTypeNode extends Base { readonly kind: 'function-type'; readonly typeParameters?: NodeArray<TypeParameterDeclaration>; readonly parameters: NodeArray<ParameterDeclaration>; readonly type: KeywordTypeNode; }
interface ConstructorDeclaration extends Base { readonly kind: 'constructor'; readonly typeParameters?: NodeArray<TypeParameterDeclaration>; readonly parameters: NodeArray<ParameterDeclaration>; }
interface MappedTypeNode extends Base { readonly kind: 'mapped'; readonly typeParameter: TypeParameterDeclaration; readonly type?: KeywordTypeNode; readonly members?: NodeArray<TypeElement>; }
type HasType = MethodSignature | FunctionExpression | FunctionTypeNode | ConstructorDeclaration | MappedTypeNode;
interface ImportTypeNode extends Base { readonly kind: 'import-type'; readonly isTypeOf: boolean; readonly argument: KeywordTypeNode; readonly typeArguments?: NodeArray<KeywordTypeNode>; }
interface TypeQueryNode extends Base { readonly kind: 'type-query'; readonly exprName: Identifier; readonly typeArguments?: NodeArray<KeywordTypeNode>; }
interface TaggedTemplateExpression extends Base { readonly kind: 'tagged-template'; readonly tag: Identifier; readonly typeArguments?: NodeArray<KeywordTypeNode>; readonly template: string; }
interface JSDocText extends Base { readonly kind: 'jsdoc-text'; text: string; }
interface JSDocLink extends Base { readonly kind: 'jsdoc-link'; text: string; }
type JSDocComment = JSDocText | JSDocLink;
interface JSDocImportTag extends Base { readonly kind: 'jsdoc-import'; readonly tagName: Identifier; readonly comment?: string | NodeArray<JSDocComment>; readonly moduleSpecifier: Identifier; }
interface JSDocTypedefTag extends Base { readonly kind: 'jsdoc-typedef'; readonly tagName: Identifier; readonly comment?: string | NodeArray<JSDocComment>; readonly name?: Identifier; }
interface TypeParameter { flags: number; }
interface Type { flags: number; }
interface ConditionalRoot { checkType: Type; extendsType: Type; isDistributive: boolean; inferTypeParameters?: TypeParameter[]; outerTypeParameters?: TypeParameter[]; instantiations?: Map<string, Type>; }
interface ConditionalType extends Base { readonly kind: 'conditional'; flags: number; root: ConditionalRoot; }
interface AnonymousType extends Base { readonly kind: 'anonymous'; flags: number; aliasTypeArguments?: readonly Type[]; }
interface CommentDirective { range: Range; type: CommentDirectiveType; }
interface DiagnosticWithLocation { category: number; code: number; start: number; length: number; messageText: string; }
interface SourceFile extends Base { readonly kind: 'source'; fileName: string; commentDirectives?: CommentDirective[]; parseDiagnostics: DiagnosticWithLocation[]; }
function classExpression(node: Base): ClassExpression { return node as ClassExpression; }
function functionExpression(node: Base): FunctionExpression { return node as FunctionExpression; }
function functionType(node: Base): FunctionTypeNode { return node as FunctionTypeNode; }
function constructorDeclaration(node: Base): ConstructorDeclaration { return node as ConstructorDeclaration; }
function mapped(node: Base): MappedTypeNode { return node as MappedTypeNode; }
function methodSignature(node: Base): MethodSignature { return node as MethodSignature; }
function importType(node: Base): ImportTypeNode { return node as ImportTypeNode; }
function typeQuery(node: Base): TypeQueryNode { return node as TypeQueryNode; }
function taggedTemplate(node: Base): TaggedTemplateExpression { return node as TaggedTemplateExpression; }
function jsdocImport(node: Base): JSDocImportTag { return node as JSDocImportTag; }
function jsdocTypedef(node: Base): JSDocTypedefTag { return node as JSDocTypedefTag; }
function conditional(type: Base): ConditionalType { return type as ConditionalType; }
function anonymous(type: Base): AnonymousType { return type as AnonymousType; }
function sourceFile(node: Base): SourceFile { return node as SourceFile; }
function typeParameterNames(node: HasType | ClassExpression): string { return node.kind === 'function-expression' || node.kind === 'mapped' || node.typeParameters === undefined ? 'none' : node.typeParameters.map(item => item.name.text).join(','); }
function parameterNames(node: HasType): string { return node.kind === 'mapped' ? '-' : node.parameters.map(item => item.name.text).join(','); }
function commentText(comment: string | NodeArray<JSDocComment> | undefined): string { return comment === undefined ? 'none' : typeof comment === 'string' ? comment : comment.map(item => item.text).join(''); }
function typeArgumentText(node: ImportTypeNode | TypeQueryNode | TaggedTemplateExpression): string { return node.typeArguments === undefined ? '' : `<${node.typeArguments.map(item => item.text).join(', ')}>`; }
const raw = {kind: 'conditional' as const, flags: 16777216, root: {checkType: {flags: 1}, extendsType: {flags: 2}, isDistributive: true, outerTypeParameters: [{flags: 'T'}]}};
const outer = conditional(raw).root.outerTypeParameters!;
console.log(`${outer.length}`);
console.log(`${outer[0]!.flags}`);
