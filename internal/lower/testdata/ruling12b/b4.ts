enum SyntaxKind { PropertyAccess = 212, ElementAccess = 213, Other = 214 }
interface Node { readonly kind: SyntaxKind; readonly pos: number }
interface PropertyAccess extends Node { readonly kind: SyntaxKind.PropertyAccess; readonly name: string }
interface ElementAccess extends Node { readonly kind: SyntaxKind.ElementAccess; readonly argument: string | number }
interface Declared extends Node { readonly isDeclaration: boolean }
type StringLiteral = "literal";
type LiteralAccess = ElementAccess & Declared & { readonly argument: StringLiteral };
function isPropertyAccess(node: Node): node is PropertyAccess { return node.kind === SyntaxKind.Other; }
function getNameOrArgument(expr: PropertyAccess | LiteralAccess): string {
    if (isPropertyAccess(expr)) return expr.name;
    return expr.argument.toUpperCase();
}
const property: PropertyAccess = { kind: SyntaxKind.PropertyAccess, pos: 0, name: "name" };
const element: LiteralAccess = { kind: SyntaxKind.ElementAccess, pos: 1, isDeclaration: true, argument: "literal" };
console.log(getNameOrArgument(property));
