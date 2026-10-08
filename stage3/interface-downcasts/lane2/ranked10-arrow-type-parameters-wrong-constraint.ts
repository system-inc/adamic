interface Base { readonly kind: 'arrow' | 'get' | 'set' | 'type-parameter' | 'identifier' | 'string' | 'keyword' | 'export' | 'static' | 'decorator' | 'named-exports' | 'export-specifier' | 'source' | 'array-binding' | 'object-binding' | 'binding-element' | 'omitted' | 'call' | 'class-expression' | 'heritage-clause' | 'heritage' | 'property' | 'method' | 'template-tag' | 'interface-type' | 'reference' | 'variable'; }
interface Range { readonly pos: number; readonly end: number; }
interface NodeArray<T> extends ReadonlyArray<T>, Range { readonly hasTrailingComma: boolean; }
interface Identifier extends Base { readonly kind: 'identifier'; readonly text: string; }
interface StringLiteral extends Base { readonly kind: 'string'; readonly text: string; }
interface TypeNode extends Base { readonly kind: 'keyword'; readonly text: string; }
interface ModifierToken extends Base { readonly kind: 'export' | 'static'; }
interface Decorator extends Base { readonly kind: 'decorator'; readonly expression: Identifier; }
type ModifierLike = ModifierToken | Decorator;
interface TypeParameterDeclaration extends Base { readonly kind: 'type-parameter'; readonly name: Identifier; readonly constraint?: TypeNode; readonly default?: TypeNode; }
interface ArrowFunction extends Base { readonly kind: 'arrow'; readonly typeParameters?: NodeArray<TypeParameterDeclaration>; }
interface GetAccessorDeclaration extends Base { readonly kind: 'get'; readonly modifiers?: NodeArray<ModifierLike>; readonly name: Identifier; }
interface SetAccessorDeclaration extends Base { readonly kind: 'set'; readonly modifiers?: NodeArray<ModifierLike>; readonly name: Identifier; }
type HasType = ArrowFunction | GetAccessorDeclaration | SetAccessorDeclaration;
interface ExportSpecifier extends Base { readonly kind: 'export-specifier'; readonly isTypeOnly: boolean; readonly propertyName?: Identifier | StringLiteral; readonly name: Identifier | StringLiteral; }
interface NamedExports extends Base { readonly kind: 'named-exports'; readonly elements: NodeArray<ExportSpecifier>; }
interface FileReference { pos: number; end: number; fileName: string; resolutionMode?: 1 | 99; preserve?: boolean; }
interface SourceFile extends Base { readonly kind: 'source'; fileName: string; referencedFiles: readonly FileReference[]; typeReferenceDirectives: readonly FileReference[]; }
interface OmittedExpression extends Base { readonly kind: 'omitted'; }
interface BindingElement extends Base { readonly kind: 'binding-element'; readonly propertyName?: Identifier; readonly name: BindingName; readonly initializer?: Identifier; }
type ArrayBindingElement = BindingElement | OmittedExpression;
interface ObjectBindingPattern extends Base { readonly kind: 'object-binding'; readonly elements: NodeArray<BindingElement>; }
interface ArrayBindingPattern extends Base { readonly kind: 'array-binding'; readonly elements: NodeArray<ArrayBindingElement>; }
type BindingName = Identifier | ObjectBindingPattern | ArrayBindingPattern;
interface CallExpression extends Base { readonly kind: 'call'; readonly expression: Identifier; readonly typeArguments?: NodeArray<TypeNode>; }
interface ExpressionWithTypeArguments extends Base { readonly kind: 'heritage'; readonly expression: Identifier; }
interface HeritageClause extends Base { readonly kind: 'heritage-clause'; readonly token: 96 | 119; readonly types: NodeArray<ExpressionWithTypeArguments>; }
interface ClassElement extends Base { readonly kind: 'property' | 'method'; readonly name: Identifier; }
interface ClassExpression extends Base { readonly kind: 'class-expression'; readonly name?: Identifier; readonly heritageClauses?: NodeArray<HeritageClause>; readonly members: NodeArray<ClassElement>; }
interface JSDocTemplateTag extends Base { readonly kind: 'template-tag'; readonly tagName: Identifier; readonly constraint: TypeNode | undefined; readonly typeParameters: NodeArray<TypeParameterDeclaration>; }
interface TypeParameter { flags: number; }
interface InterfaceType extends Base { readonly kind: 'interface-type'; flags: number; typeParameters: TypeParameter[] | undefined; }
interface Type { flags: number; }
interface TypeReference extends Base { readonly kind: 'reference'; flags: number; resolvedTypeArguments?: readonly Type[]; }
interface Declaration extends Base { readonly kind: 'variable' | 'class-expression'; readonly pos: number; }
interface Symbol { escapedName: string; flags: number; declarations?: Declaration[]; }
function arrow(node: Base): ArrowFunction { return node as ArrowFunction; }
function getAccessor(node: Base): GetAccessorDeclaration { return node as GetAccessorDeclaration; }
function setAccessor(node: Base): SetAccessorDeclaration { return node as SetAccessorDeclaration; }
function namedExports(node: Base): NamedExports { return node as NamedExports; }
function sourceFile(node: Base): SourceFile { return node as SourceFile; }
function arrayBinding(node: Base): ArrayBindingPattern { return node as ArrayBindingPattern; }
function objectBinding(node: Base): ObjectBindingPattern { return node as ObjectBindingPattern; }
function call(node: Base): CallExpression { return node as CallExpression; }
function classExpression(node: Base): ClassExpression { return node as ClassExpression; }
function templateTag(node: Base): JSDocTemplateTag { return node as JSDocTemplateTag; }
function interfaceType(type: Base): InterfaceType { return type as InterfaceType; }
function typeReference(type: Base): TypeReference { return type as TypeReference; }
function modifierText(node: HasType): string { return node.kind === 'arrow' || node.modifiers === undefined ? 'none' : node.modifiers.map(modifier => modifier.kind === 'decorator' ? `@${modifier.expression.text}` : modifier.kind).join(' '); }
function bindingText(name: BindingName): string { return name.kind === 'identifier' ? name.text : name.kind === 'object-binding' ? `{${name.elements.map(element => bindingText(element.name)).join(',')}}` : `[${name.elements.map(element => element.kind === 'omitted' ? '' : bindingText(element.name)).join(',')}]`; }
function firstDeclaration(symbol: Symbol | undefined): string { const declarations = symbol?.declarations; return declarations === undefined ? 'none' : `${declarations.length}:${declarations[0]!.kind}`; }
const raw = {kind: 'arrow' as const, typeParameters: Object.assign([{kind: 'type-parameter' as const, name: {kind: 'identifier' as const, text: 'T'}, constraint: 'string'}], {pos: 0, end: 1, hasTrailingComma: false})};
const first = arrow(raw).typeParameters![0]!;
console.log(first.name.text);
console.log(first.constraint === undefined ? '-' : first.constraint.text);
