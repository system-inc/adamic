enum SyntaxKind { PropertyAccess = 212, ElementAccess = 213, Other = 214 }
interface Node { readonly kind: SyntaxKind; readonly pos: number }
interface PropertyAccess extends Node { readonly kind: SyntaxKind.PropertyAccess; readonly name: string }
interface ElementAccess extends Node { readonly kind: SyntaxKind.ElementAccess; readonly argument: string | number }
interface Declared extends Node { readonly isDeclaration: boolean }
type StringLiteral = "literal";
type LiteralAccess = ElementAccess & Declared & { readonly argument: StringLiteral };
function isLiteralAccess(node: Node): node is ElementAccess {
    return node.kind === SyntaxKind.ElementAccess;
}
function assertNever(node: unknown): never { throw new Error("assertNever"); }
function read(expr: LiteralAccess): string {
    if (isLiteralAccess(expr)) return expr.argument;
    return assertNever(expr);
}
const element: LiteralAccess = { kind: SyntaxKind.ElementAccess, pos: 1, isDeclaration: true, argument: "literal" };
console.log(read(element));
