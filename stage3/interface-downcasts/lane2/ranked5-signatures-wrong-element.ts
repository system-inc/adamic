interface Base { readonly kind: 'resolved' | 'class' | 'tuple' | 'heritage' | 'jsdoc' | 'function' | 'method' | 'arrow' | 'identifier' | 'keyword' | 'named' | 'export' | 'abstract' | 'decorator' | 'tag' | 'parameter'; }
interface Range { readonly pos: number; readonly end: number; }
interface NodeArray<T> extends ReadonlyArray<T>, Range { readonly hasTrailingComma: boolean; }
interface Identifier extends Base { readonly kind: 'identifier'; readonly text: string; }
interface TypeNode extends Base { readonly pos: number; }
interface NamedTupleMember extends TypeNode { readonly kind: 'named'; readonly name: Identifier; readonly type: TypeNode; }
interface TupleTypeNode extends TypeNode { readonly kind: 'tuple'; readonly elements: NodeArray<TypeNode | NamedTupleMember>; }
interface ModifierToken extends Base { readonly kind: 'export' | 'abstract'; readonly pos: number; }
interface Decorator extends Base { readonly kind: 'decorator'; readonly pos: number; readonly expression: Identifier; }
type ModifierLike = ModifierToken | Decorator;
interface ClassDeclaration extends Base { readonly kind: 'class'; readonly modifiers?: NodeArray<ModifierLike>; readonly name?: Identifier; }
interface ExpressionWithTypeArguments extends TypeNode { readonly kind: 'heritage'; readonly expression: Identifier; readonly typeArguments?: NodeArray<TypeNode>; }
interface JSDocTag extends Base { readonly kind: 'tag'; readonly tagName: Identifier; readonly comment?: string; }
interface JSDoc extends Base { readonly kind: 'jsdoc'; readonly tags?: NodeArray<JSDocTag>; readonly comment?: string; }
interface ParameterDeclaration extends Base { readonly kind: 'parameter'; readonly name: Identifier; }
interface FunctionDeclaration extends Base { readonly kind: 'function'; readonly parameters: NodeArray<ParameterDeclaration>; }
interface MethodDeclaration extends Base { readonly kind: 'method'; readonly parameters: NodeArray<ParameterDeclaration>; }
interface ArrowFunction extends Base { readonly kind: 'arrow'; readonly parameters: NodeArray<ParameterDeclaration>; }
type FunctionLikeDeclaration = FunctionDeclaration | MethodDeclaration | ArrowFunction;
interface Symbol { readonly escapedName: string; }
interface Signature { declaration?: FunctionLikeDeclaration; parameters: readonly Symbol[]; minArgumentCount: number; }
interface ResolvedType extends Base { readonly kind: 'resolved'; callSignatures: readonly Signature[]; constructSignatures: readonly Signature[]; }
function resolved(type: Base): ResolvedType { return type as ResolvedType; }
function classDeclaration(node: Base): ClassDeclaration { return node as ClassDeclaration; }
function tuple(node: Base): TupleTypeNode { return node as TupleTypeNode; }
function heritage(node: Base): ExpressionWithTypeArguments { return node as ExpressionWithTypeArguments; }
function jsDoc(node: Base): JSDoc { return node as JSDoc; }
function functionDeclaration(node: Base): FunctionDeclaration { return node as FunctionDeclaration; }
function method(node: Base): MethodDeclaration { return node as MethodDeclaration; }
function arrow(node: Base): ArrowFunction { return node as ArrowFunction; }
function isNamedTupleMember(node: TypeNode): node is NamedTupleMember { return node.kind === 'named'; }
function parametersOf(declaration: FunctionLikeDeclaration): NodeArray<ParameterDeclaration> { return declaration.parameters; }
const raw = {kind: 'resolved' as const, callSignatures: [], constructSignatures: [{parameters: [], minArgumentCount: 'one'}]};
const signatures = resolved(raw).constructSignatures;
console.log(`${signatures[0]!.minArgumentCount}`);
