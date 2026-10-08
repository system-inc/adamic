interface Base { readonly kind: 'class' | 'heritage-clause' | 'heritage' | 'identifier' | 'set' | 'constructor' | 'property' | 'parameter' | 'source' | 'command' | 'indexed' | 'template'; }
interface Range { readonly pos: number; readonly end: number; }
interface NodeArray<T> extends ReadonlyArray<T>, Range { readonly hasTrailingComma: boolean; }
interface Identifier extends Base { readonly kind: 'identifier'; readonly text: string; }
interface Type { flags: number; aliasTypeArguments?: readonly Type[]; }
interface IndexedAccessType extends Base { readonly kind: 'indexed'; flags: number; objectType: Type; indexType: Type; }
interface TemplateLiteralType extends Base { readonly kind: 'template'; texts: readonly string[]; types: readonly Type[]; }
interface ExpressionWithTypeArguments extends Base { readonly kind: 'heritage'; readonly expression: Identifier; }
interface HeritageClause extends Base { readonly kind: 'heritage-clause'; readonly token: 96 | 119; readonly types: NodeArray<ExpressionWithTypeArguments>; }
interface ClassDeclaration extends Base { readonly kind: 'class'; readonly name?: Identifier; readonly heritageClauses?: NodeArray<HeritageClause>; }
interface ParameterDeclaration extends Base { readonly kind: 'parameter'; readonly name: Identifier; }
interface SetAccessorDeclaration extends Base { readonly kind: 'set'; readonly name: Identifier; readonly parameters: NodeArray<ParameterDeclaration>; }
interface ConstructorDeclaration extends Base { readonly kind: 'constructor'; readonly parameters: NodeArray<ParameterDeclaration>; }
interface PropertyDeclaration extends Base { readonly kind: 'property'; readonly name: Identifier; }
type HasType = SetAccessorDeclaration | ConstructorDeclaration | PropertyDeclaration;
interface DiagnosticMessageChain { messageText: string; category: number; code: number; next?: DiagnosticMessageChain[]; }
interface SourceFile extends Base { readonly kind: 'source'; fileName: string; bindDiagnostics: DiagnosticWithLocation[]; }
interface DiagnosticRelatedInformation { category: number; code: number; file: SourceFile | undefined; start: number | undefined; length: number | undefined; messageText: string | DiagnosticMessageChain; }
interface Diagnostic extends DiagnosticRelatedInformation { source?: string; relatedInformation?: DiagnosticRelatedInformation[]; }
interface DiagnosticWithLocation extends Diagnostic { file: SourceFile; start: number; length: number; }
interface ProjectReference { path: string; originalPath?: string; prepend?: boolean; circular?: boolean; }
interface ParsedCommandLine extends Base { readonly kind: 'command'; fileNames: string[]; projectReferences?: readonly ProjectReference[]; errors: Diagnostic[]; }
function classDeclaration(node: Base): ClassDeclaration { return node as ClassDeclaration; }
function setAccessor(node: Base): SetAccessorDeclaration { return node as SetAccessorDeclaration; }
function constructorDeclaration(node: Base): ConstructorDeclaration { return node as ConstructorDeclaration; }
function indexed(type: Base): IndexedAccessType { return type as IndexedAccessType; }
function template(type: Base): TemplateLiteralType { return type as TemplateLiteralType; }
function sourceFile(node: Base): SourceFile { return node as SourceFile; }
function commandLine(value: Base): ParsedCommandLine { return value as ParsedCommandLine; }
function parameterNames(node: HasType): string { return node.kind === 'property' ? 'none' : node.parameters.map(parameter => parameter.name.text).join(','); }
function relatedCodes(diagnostic: Diagnostic): string { return diagnostic.relatedInformation === undefined ? 'absent' : diagnostic.relatedInformation.map(related => related.code).join(','); }
const raw = {kind: 'command' as const, fileNames: [], errors: [{category: 1, code: 5001, file: undefined, start: undefined, length: undefined, messageText: 'top', relatedInformation: 'none'}]};
const related = commandLine(raw).errors[0]!.relatedInformation;
console.log(related === undefined ? 'absent' : `${related.length}`);
