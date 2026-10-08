interface Base { readonly kind: 'class' | 'class-expression' | 'interface' | 'export-declaration' | 'function-expression' | 'arrow' | 'constructor' | 'variable-statement' | 'import-equals' | 'import-attributes' | 'import-attribute' | 'identifier' | 'string' | 'keyword' | 'type-parameter' | 'parameter' | 'export' | 'async' | 'static' | 'declare' | 'decorator' | 'heritage-clause' | 'heritage' | 'property-signature' | 'method-signature' | 'property' | 'computed' | 'source' | 'context' | 'resolved'; }
interface Range { readonly pos: number; readonly end: number; }
interface NodeArray<T> extends ReadonlyArray<T>, Range { readonly hasTrailingComma: boolean; }
interface Identifier extends Base { readonly kind: 'identifier'; readonly text: string; }
interface StringLiteral extends Base { readonly kind: 'string'; readonly text: string; }
interface TypeNode extends Base { readonly kind: 'keyword'; readonly text: string; }
interface ModifierToken extends Base { readonly kind: 'export' | 'async' | 'static' | 'declare'; }
interface Decorator extends Base { readonly kind: 'decorator'; readonly expression: Identifier; }
type Modifier = ModifierToken;
type ModifierLike = Modifier | Decorator;
interface TypeParameterDeclaration extends Base { readonly kind: 'type-parameter'; readonly name: Identifier; readonly constraint?: TypeNode; }
interface ParameterDeclaration extends Base { readonly kind: 'parameter'; readonly name: Identifier; }
interface ExpressionWithTypeArguments extends Base { readonly kind: 'heritage'; readonly expression: Identifier; }
interface HeritageClause extends Base { readonly kind: 'heritage-clause'; readonly token: 96 | 119; readonly types: NodeArray<ExpressionWithTypeArguments>; }
interface PropertySignature extends Base { readonly kind: 'property-signature'; readonly name: Identifier; readonly type?: TypeNode; }
interface MethodSignature extends Base { readonly kind: 'method-signature'; readonly name: Identifier; readonly parameters: NodeArray<ParameterDeclaration>; }
type TypeElement = PropertySignature | MethodSignature;
interface ClassDeclaration extends Base { readonly kind: 'class'; readonly modifiers?: NodeArray<ModifierLike>; readonly name?: Identifier; readonly typeParameters?: NodeArray<TypeParameterDeclaration>; }
interface ClassExpression extends Base { readonly kind: 'class-expression'; readonly modifiers?: NodeArray<ModifierLike>; readonly name?: Identifier; }
interface InterfaceDeclaration extends Base { readonly kind: 'interface'; readonly modifiers?: NodeArray<ModifierLike>; readonly name: Identifier; readonly heritageClauses?: NodeArray<HeritageClause>; readonly members: NodeArray<TypeElement>; }
interface ExportDeclaration extends Base { readonly kind: 'export-declaration'; readonly modifiers?: NodeArray<ModifierLike>; readonly isTypeOnly: boolean; readonly moduleSpecifier?: StringLiteral; }
interface FunctionExpression extends Base { readonly kind: 'function-expression'; readonly modifiers?: NodeArray<Modifier>; readonly name?: Identifier; readonly parameters: NodeArray<ParameterDeclaration>; }
interface ArrowFunction extends Base { readonly kind: 'arrow'; readonly modifiers?: NodeArray<Modifier>; readonly parameters: NodeArray<ParameterDeclaration>; }
interface ConstructorDeclaration extends Base { readonly kind: 'constructor'; readonly modifiers?: NodeArray<ModifierLike>; readonly parameters: NodeArray<ParameterDeclaration>; }
interface VariableStatement extends Base { readonly kind: 'variable-statement'; readonly modifiers?: NodeArray<ModifierLike>; readonly name: Identifier; }
interface ImportEqualsDeclaration extends Base { readonly kind: 'import-equals'; readonly modifiers?: NodeArray<ModifierLike>; readonly name: Identifier; readonly isTypeOnly: boolean; }
type HasType = FunctionExpression | ArrowFunction | ConstructorDeclaration;
type ObjectTypeDeclaration = ClassDeclaration | ClassExpression | InterfaceDeclaration;
interface ImportAttribute extends Base { readonly kind: 'import-attribute'; readonly name: Identifier | StringLiteral; readonly value: StringLiteral; }
interface ImportAttributes extends Base { readonly kind: 'import-attributes'; readonly token: 118 | 132; readonly elements: NodeArray<ImportAttribute>; readonly multiLine?: boolean; }
interface ComputedPropertyName extends Base { readonly kind: 'computed'; readonly expression: Identifier; }
interface ClassElementBase extends Base { readonly kind: 'property'; readonly pos: number; }
type ElementWithComputedPropertyName = ClassElementBase & { name: ComputedPropertyName; };
interface Type { flags: number; }
interface IndexInfo { keyType: Type; type: Type; isReadonly: boolean; components?: ElementWithComputedPropertyName[]; }
interface ResolvedType extends Base { readonly kind: 'resolved'; indexInfos: readonly IndexInfo[]; }
interface FileReference { pos: number; end: number; fileName: string; preserve?: boolean; }
interface SourceFile extends Base { readonly kind: 'source'; fileName: string; libReferenceDirectives: readonly FileReference[]; }
interface ReverseMappedSymbol { escapedName: string; flags: number; links: { propertyType?: Type; }; }
interface NodeBuilderContext extends Base { readonly kind: 'context'; flags: number; reverseMappedStack: ReverseMappedSymbol[] | undefined; }
function classDeclaration(node: Base): ClassDeclaration { return node as ClassDeclaration; }
function classExpression(node: Base): ClassExpression { return node as ClassExpression; }
function interfaceDeclaration(node: Base): InterfaceDeclaration { return node as InterfaceDeclaration; }
function exportDeclaration(node: Base): ExportDeclaration { return node as ExportDeclaration; }
function functionExpression(node: Base): FunctionExpression { return node as FunctionExpression; }
function arrow(node: Base): ArrowFunction { return node as ArrowFunction; }
function constructorDeclaration(node: Base): ConstructorDeclaration { return node as ConstructorDeclaration; }
function variableStatement(node: Base): VariableStatement { return node as VariableStatement; }
function importEquals(node: Base): ImportEqualsDeclaration { return node as ImportEqualsDeclaration; }
function importAttributes(node: Base): ImportAttributes { return node as ImportAttributes; }
function sourceFile(node: Base): SourceFile { return node as SourceFile; }
function resolved(type: Base): ResolvedType { return type as ResolvedType; }
function context(value: Base): NodeBuilderContext { return value as NodeBuilderContext; }
function modifierList(modifiers: NodeArray<ModifierLike> | undefined): string { return modifiers === undefined ? 'none' : modifiers.map(modifier => modifier.kind === 'decorator' ? `@${modifier.expression.text}` : modifier.kind).join(' '); }
function parameterNames(node: HasType): string { return node.parameters.map(parameter => parameter.name.text).join(','); }
function heritageNames(node: ObjectTypeDeclaration): string { return node.kind === 'interface' && node.heritageClauses !== undefined ? node.heritageClauses.map(clause => clause.types.map(type => type.expression.text).join('&')).join('|') : 'none'; }
const raw = {kind: 'source' as const, fileName: 'a.ts', libReferenceDirectives: [{pos: 0, end: '3', fileName: 'es2020'}]};
const directives = sourceFile(raw).libReferenceDirectives;
console.log(directives[0]!.fileName);
console.log(`${directives[0]!.end + 1}`);
