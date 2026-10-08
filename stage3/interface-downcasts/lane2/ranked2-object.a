interface Base { readonly kind: 'variables' | 'call' | 'object' | 'tuple' | 'array'; }
interface VariableDeclaration { readonly pos: number; }
interface Expression { readonly pos: number; }
interface ObjectLiteralElementLike { readonly pos: number; }
interface Range { readonly pos: number; readonly end: number; }
interface NodeArray<T> extends ReadonlyArray<T>, Range { readonly hasTrailingComma: boolean; }
enum ElementFlags { Required = 1, Optional = 2 }
interface VariableDeclarationList extends Base { readonly kind: 'variables'; readonly declarations: NodeArray<VariableDeclaration>; }
interface CallExpression extends Base { readonly kind: 'call'; readonly arguments: NodeArray<Expression>; }
interface ObjectLiteralExpression extends Base { readonly kind: 'object'; readonly properties: NodeArray<ObjectLiteralElementLike>; }
interface TupleType extends Base { readonly kind: 'tuple'; readonly elementFlags: readonly ElementFlags[]; }
interface ArrayLiteralExpression extends Base { readonly kind: 'array'; readonly elements: NodeArray<Expression>; }
function variables(node: Base): VariableDeclarationList { return node as VariableDeclarationList; }
function call(node: Base): CallExpression { return node as CallExpression; }
function object(node: Base): ObjectLiteralExpression { return node as ObjectLiteralExpression; }
function tuple(node: Base): TupleType { return node as TupleType; }
function array(node: Base): ArrayLiteralExpression { return node as ArrayLiteralExpression; }
const raw = {kind: 'object' as const, properties: Object.assign([{pos: 1}, {pos: 2}], {pos: 0, end: 2, hasTrailingComma: false})};
const values = object(raw).properties;
console.log(`${values.length}:${values[0]!.pos}`);
